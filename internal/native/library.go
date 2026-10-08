package native

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
)

type runtimeFile struct {
	name     string
	contents []byte
}

// runtimeBuilds serializes builders of the same disk entry within this process. Different flag
// sets can build concurrently. Failures live only here, never on disk, so a fresh process retries.
// Across processes, only a complete directory is published by rename.
var runtimeBuilds sync.Map

type runtimeBuild struct {
	lock sync.Mutex
	err  error
}

// RuntimeLibrary returns a cached static library compiled with Flags(options). An empty directory
// uses the embedded runtime; the fuzzer supplies another checkout's runtime directory instead.
// Sources and headers are snapshotted together, so the key and the compiled bytes cannot disagree.
func RuntimeLibrary(directory string, options Options) (string, error) {
	if shippedRelease(options) && options.Profile != "" && !options.profileValidated {
		return "", fmt.Errorf("native: profile runtime requires Build to verify its emitted C first")
	}
	if err := ValidateOptions(options); err != nil {
		return "", err
	}
	var sources fs.FS = runtime
	root := "runtime"
	if directory != "" {
		sources, root = os.DirFS(directory), "."
	}
	files, err := readRuntime(sources, root)
	if err != nil {
		return "", fmt.Errorf("native: runtime: %w", err)
	}
	compiler, err := exec.LookPath(compilerName(options))
	if err != nil {
		return "", fmt.Errorf("native: %w", err)
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("native: clang --version: %w\n%s", err, version)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("native: cache directory: %w", err)
	}
	jobs := 1
	if useSplit(options) {
		jobs, err = splitJobs(options)
		if err != nil {
			return "", err
		}
	}
	return cachedRuntime(files, Flags(options), compiler, string(version), filepath.Join(cache, "adamic", "runtime"), jobs)
}

func readRuntime(sources fs.FS, root string) ([]runtimeFile, error) {
	entries, err := fs.ReadDir(sources, root)
	if err != nil {
		return nil, err
	}
	var files []runtimeFile
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h")) {
			continue
		}
		contents, err := fs.ReadFile(sources, path.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		files = append(files, runtimeFile{entry.Name(), contents})
	}
	return files, nil
}

func runtimeKey(files []runtimeFile, flags []string, compiler string, version string) string {
	hash := sha256.New()
	// Length prefixes preserve flag boundaries, order and arbitrary source bytes.
	part := func(value string) { fmt.Fprintf(hash, "%d:", len(value)); hash.Write([]byte(value)) }
	part("adamic-runtime-v1")
	for _, flag := range flags {
		if strings.HasPrefix(flag, "-fprofile-") {
			part("stable-profile-source-names-v1")
			break
		}
	}
	part(goruntime.GOOS)
	part(goruntime.GOARCH)
	part(compiler)
	part(version)
	fmt.Fprintf(hash, "%d:", len(flags))
	for _, flag := range flags {
		part(flag)
	}
	fmt.Fprintf(hash, "%d:", len(files))
	for _, file := range files {
		part(file.name)
		fmt.Fprintf(hash, "%d:", len(file.contents))
		hash.Write(file.contents)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func cachedRuntime(files []runtimeFile, flags []string, compiler string, version string, cache string, parallel ...int) (result string, failure error) {
	directory := filepath.Join(cache, runtimeKey(files, flags, compiler, version))
	library := filepath.Join(directory, "runtime.a")
	value, _ := runtimeBuilds.LoadOrStore(directory, &runtimeBuild{})
	build := value.(*runtimeBuild)
	build.lock.Lock()
	defer build.lock.Unlock()
	if build.err != nil {
		return "", build.err
	}
	// Remember every failure after the key is known, including temporary-directory failures.
	defer func() { build.err = failure }()
	if info, err := os.Stat(library); err == nil && info.Mode().IsRegular() {
		return library, nil
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", fmt.Errorf("native: runtime cache: %w", err)
	}
	temporary, err := os.MkdirTemp(cache, ".build-")
	if err != nil {
		return "", fmt.Errorf("native: %w", err)
	}
	defer os.RemoveAll(temporary)
	// Headers live beside the archive, from the same snapshot that produced its objects.
	if err := os.Chmod(temporary, 0o755); err != nil {
		return "", err
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(temporary, file.name), file.contents, 0o644); err != nil {
			return "", err
		}
	}
	var sources []runtimeFile
	for _, file := range files {
		if strings.HasSuffix(file.name, ".c") {
			sources = append(sources, file)
		}
	}
	if len(sources) == 0 {
		return "", fmt.Errorf("native: runtime has no C sources")
	}
	jobs := 1
	if len(parallel) != 0 {
		jobs = parallel[0]
	}
	if jobs < 1 {
		return "", fmt.Errorf("native: runtime jobs must be positive")
	}
	if jobs > len(sources) {
		jobs = len(sources)
	}
	objects := make([]string, len(sources))
	compile := func(index int) error {
		file := sources[index]
		object := filepath.Join(temporary, strings.TrimSuffix(file.name, ".c")+".o")
		arguments := append(append([]string{}, flags...), "-c", filepath.Join(temporary, file.name), "-o", object)
		command := exec.Command(compiler, arguments...)
		for _, flag := range flags {
			if strings.HasPrefix(flag, "-fprofile-") {
				// ThinLTO uses source identity in private names and GUIDs. Keep it
				// independent of random cache directory names for reproducible binaries.
				command.Args[len(flags)+2] = file.name
				command.Dir = temporary
				break
			}
		}
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("native: compiling runtime %s: %w\n%s", file.name, err, output)
		}
		objects[index] = object
		return nil
	}
	if jobs == 1 {
		for index := range sources {
			if err := compile(index); err != nil {
				return "", err
			}
		}
	} else {
		// Keep archive order and error selection independent of completion order.
		// Every worker finishes before the cache snapshot can be published or removed.
		errors := make([]error, len(sources))
		work := make(chan int)
		var workers sync.WaitGroup
		for worker := 0; worker < jobs; worker++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for index := range work {
					errors[index] = compile(index)
				}
			}()
		}
		for index := range sources {
			work <- index
		}
		close(work)
		workers.Wait()
		for _, err := range errors {
			if err != nil {
				return "", err
			}
		}
	}
	arguments := append([]string{"rcs", filepath.Join(temporary, "runtime.a")}, objects...)
	if output, err := exec.Command(archiverName(compiler), arguments...).CombinedOutput(); err != nil {
		return "", fmt.Errorf("native: archiving runtime: %w\n%s", err, output)
	}
	// Publish everything together. A competing process may already have published this key; its
	// identical complete entry wins, while this process removes only its own temporary directory.
	if err := os.Rename(temporary, directory); err != nil {
		if info, statErr := os.Stat(library); statErr != nil || !info.Mode().IsRegular() {
			return "", fmt.Errorf("native: publishing runtime: %w", err)
		}
	}
	return library, nil
}

// RuntimeLinkFlags retains every translation unit, as Build did when linking the object files
// directly. In particular a count-reporting destructor must run even for an empty main.
func RuntimeLinkFlags(library string) []string {
	if goruntime.GOOS == "darwin" {
		return []string{"-Xlinker", "-force_load", "-Xlinker", library}
	}
	return []string{"-Xlinker", "--whole-archive", library, "-Xlinker", "--no-whole-archive"}
}
