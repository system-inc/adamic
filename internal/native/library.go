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
	return runtimeLibrary(directory, options, nil)
}

func RuntimeLibraryForSource(directory string, source string, options Options) (string, error) {
	return runtimeLibrary(directory, options, featureFlags(source))
}

// sourceFlags compiles anything built for one program's emitted C: its runtime library, a split
// unit, the checker-archive build. It is the one home for those flags, so a new build path that
// starts from Flags(options) alone meets adamic.h's layouts without the program's features.
func sourceFlags(source string, options Options) []string {
	return append(Flags(options), featureFlags(source)...)
}

// featureFlags are the runtime features emitted C turns on with its leading #defines, as -D flags,
// so every unit compiled with that C (the runtime's own .c files, and a header included before
// those #defines) sees the same layouts.
func featureFlags(source string) []string {
	var flags []string
	for _, feature := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"} {
		if strings.Contains(source, "#define "+feature+" 1\n") {
			flags = append(flags, "-D"+feature+"=1")
		}
	}
	return flags
}

func runtimeLibrary(directory string, options Options, extraFlags []string) (string, error) {
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
	if !slicesContain(extraFlags, "-DADAMIC_REGEXP_REPLACE_CALLBACK=1") {
		kept := files[:0]
		for _, file := range files {
			if file.name != "regexp_replace.c" {
				kept = append(kept, file)
			}
		}
		files = kept
	}
	if !slicesContain(extraFlags, "-DADAMIC_NODE_HOST=1") {
		kept := files[:0]
		for _, file := range files {
			if file.name != "node_host.c" {
				kept = append(kept, file)
			}
		}
		files = kept
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
	return cachedRuntime(files, append(Flags(options), extraFlags...), compiler, string(version), filepath.Join(cache, "adamic", "runtime"))
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

func cachedRuntime(files []runtimeFile, flags []string, compiler string, version string, cache string) (result string, failure error) {
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
	for _, feature := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"} {
		if slicesContain(flags, "-D"+feature+"=1") {
			files = append([]runtimeFile(nil), files...)
			for i := range files {
				if files[i].name == "adamic.h" {
					files[i].contents = append([]byte("#define "+feature+" 1\n"), files[i].contents...)
				}
			}
		}
	}

	// Headers live beside the archive, from the same snapshot that produced its objects.
	if err := os.Chmod(temporary, 0o755); err != nil {
		return "", err
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(temporary, file.name), file.contents, 0o644); err != nil {
			return "", err
		}
	}
	var objects []string
	for _, file := range files {
		if !strings.HasSuffix(file.name, ".c") {
			continue
		}
		object := filepath.Join(temporary, strings.TrimSuffix(file.name, ".c")+".o")
		arguments := append(append([]string{}, flags...), "-c", filepath.Join(temporary, file.name), "-o", object)
		if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
			return "", fmt.Errorf("native: compiling runtime %s: %w\n%s", file.name, err, output)
		}
		objects = append(objects, object)
	}
	if len(objects) == 0 {
		return "", fmt.Errorf("native: runtime has no C sources")
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

func slicesContain(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
