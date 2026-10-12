package native

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/toolchain"
)

type runtimeFile struct {
	name     string
	contents []byte
}

// runtimeBuilds serializes builders of the same disk entry within this process. Different flag
// sets can build concurrently. Across processes, only a complete directory is published by rename.
var runtimeBuilds sync.Map

// RuntimeLibrary returns a cached static library compiled with Flags(options). An empty directory
// uses the embedded runtime; the fuzzer supplies another checkout's runtime directory instead.
// Sources and headers are snapshotted together, so the key and the compiled bytes cannot disagree.
func RuntimeLibrary(directory string, options Options) (string, error) {
	return runtimeLibrary(directory, options, nil)
}

func RuntimeLibraryForSource(directory string, source string, options Options) (string, error) {
	return runtimeLibrary(directory, options, featureFlags(source))
}

// A RuntimeRequest is one archive of the embedded runtime a build links, as a supplier (SupplyRuntimes) is asked for
// it. Parts are everything its key reads, as runtimeKey hashes them (runtimeKeyParts), so a key made of them moves
// whenever the archive's does. Features and Options are what asked for it.
type RuntimeRequest struct {
	Parts    []string
	Features []string
	Options  Options
	files    []runtimeFile
	flags    []string
	compiler string
}

// Build compiles the archive into directory as the runtime cache holds one, runtime.a beside the runtime's files,
// compiling up to workers of its translation units at once: one where a test waits on it, as the cache compiles, and
// more where a product build is given the CPUs for it.
func (request RuntimeRequest) Build(directory string, workers int) error {
	return compileRuntime(request.files, request.flags, request.compiler, directory, workers)
}

// runtimeSupplier, when set, supplies every archive of the embedded runtime this process links, in place of the
// runtime cache (SupplyRuntimes).
var runtimeSupplier atomic.Pointer[func(RuntimeRequest) (string, error)]

// suppliedRuntimes holds what the supplier gave for each key, asked once per process, as the cache is.
var suppliedRuntimes sync.Map

type suppliedRuntime struct {
	once    sync.Once
	library string
	err     error
}

// SupplyRuntimes has supplier provide every archive of the embedded runtime this process links, instead of the runtime
// cache: it returns the directory holding the request's archive, which it may fill with the request's Build, or the
// error a build linking it fails with. It is asked once per key in a process. Another checkout's runtime
// (RuntimeLibrary's directory) and split tsgo's still come from the cache. internal/oracle's tests supply each from a
// buildcache product, which a runner reads and never builds (#sqmn7jk: lap 3's red was these archives compiled inside
// the first test to need one, 27 s alone where the program takes 5).
func SupplyRuntimes(supplier func(RuntimeRequest) (string, error)) {
	runtimeSupplier.Store(&supplier)
}

// RuntimeRequestFor is the request a build of a program with these features (as Features names them) and options
// makes of the supplier, for a product to build ahead.
func RuntimeRequestFor(options Options, features ...string) (RuntimeRequest, error) {
	var source strings.Builder
	for _, feature := range features {
		if !slicesContain(runtimeFeatures, feature) {
			return RuntimeRequest{}, fmt.Errorf("native: no runtime feature %s", feature)
		}
		fmt.Fprintf(&source, "#define %s 1\n", feature)
	}
	files, flags, compiler, version, err := runtimeInputs("", options, featureFlags(source.String()))
	if err != nil {
		return RuntimeRequest{}, err
	}
	return runtimeRequest(files, flags, compiler, version, options)
}

// runtimeFeatures are the runtime's features a program's C turns on with a leading #define (cProgram), in the order
// their -D flags are given.
var runtimeFeatures = []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"}

// Features are the runtime features a program's C turns on, in runtimeFeatures' order.
func Features(source string) []string {
	var features []string
	for _, feature := range runtimeFeatures {
		if strings.Contains(source, "#define "+feature+" 1\n") {
			features = append(features, feature)
		}
	}
	return features
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
	for _, feature := range Features(source) {
		flags = append(flags, "-D"+feature+"=1")
	}
	return flags
}

func runtimeLibrary(directory string, options Options, extraFlags []string) (string, error) {
	files, flags, compiler, version, err := runtimeInputs(directory, options, extraFlags)
	if err != nil {
		return "", err
	}
	if supplier := runtimeSupplier.Load(); supplier != nil && directory == "" {
		request, err := runtimeRequest(files, flags, compiler, version, options)
		if err != nil {
			return "", err
		}
		value, _ := suppliedRuntimes.LoadOrStore(partsKey(request.Parts), &suppliedRuntime{})
		supplied := value.(*suppliedRuntime)
		supplied.once.Do(func() {
			held, err := (*supplier)(request)
			if err != nil {
				supplied.err = err
				return
			}
			supplied.library = filepath.Join(held, "runtime.a")
			if info, err := os.Stat(supplied.library); err != nil || !info.Mode().IsRegular() {
				supplied.err = fmt.Errorf("native: the supplied runtime %s holds no runtime.a", held)
			}
		})
		return supplied.library, supplied.err
	}
	cache, err := cacheDirectory("native-runtime")
	if err != nil {
		return "", err
	}
	return cachedRuntime(files, flags, compiler, version, cache)
}

// runtimeInputs are what a runtime archive is compiled from: the runtime's files (directory's, or the embedded
// runtime's when it is empty, without the units of features the flags leave off), the flags, and the compiler with its
// version.
func runtimeInputs(directory string, options Options, extraFlags []string) ([]runtimeFile, []string, string, string, error) {
	if err := ValidateOptions(options); err != nil {
		return nil, nil, "", "", err
	}
	var sources fs.FS = runtime
	root := "runtime"
	if directory != "" {
		sources, root = os.DirFS(directory), "."
	}
	files, err := readRuntime(sources, root)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("native: runtime: %w", err)
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
	// The program toolchain resolves is the one the key names and every step of the build runs.
	compiler, err := toolchain.Path(compilerName(options))
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("native: %w", err)
	}
	version, err := toolchain.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("native: clang --version: %w\n%s", err, version)
	}
	return files, append(Flags(options), extraFlags...), compiler, string(version), nil
}

// cacheDirectory is where a native cache named name lives: under the build cache (location.BuildCache), never under
// this machine's home on native's own account, so a sandboxed build that names its cache writes nothing under HOME.
func cacheDirectory(name string) (string, error) {
	cache, err := location.BuildCache()
	if err != nil {
		return "", fmt.Errorf("native: cache directory: %w", err)
	}
	return filepath.Join(cache, name), nil
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

// runtimeKey addresses a compiled runtime by everything its compile reads (runtimeKeyParts).
func runtimeKey(files []runtimeFile, flags []string, compiler string, version string, headers string) string {
	return partsKey(runtimeKeyParts(files, flags, compiler, version, headers))
}

// runtimeKeyParts is everything a compiled runtime's compile reads: the sources and headers, each by its name and what
// it holds, the flags, the compiler and its version, the archiver beside it, the platform, and headers, the system headers the compile reads
// (headersRead), so a runtime compiled against other C library or compiler headers is never taken for this one. A
// supplied runtime's product is keyed by the same parts (RuntimeRequest), so the two keys can't move apart.
func runtimeKeyParts(files []runtimeFile, flags []string, compiler string, version string, headers string) []string {
	// v2: objects no longer carry the build directory (-ffile-prefix-map), so a v1 archive built without it is never reused.
	// v3: the map covers the directory's resolved spelling too, so a v2 archive a Mac built under /private/var isn't reused.
	// v4: the archive's members carry no time (ZERO_AR_DATE), so a v3 archive Apple's ar stamped isn't reused.
	// v5: clang and ar run under toolchain.Env, not the caller's environment, so a v4 archive built under whatever the
	// caller had set (CPATH, SDKROOT, CCC_OVERRIDE_OPTIONS) isn't reused.
	// v6: the key is a list of named parts, each file by the sha256 of what it holds, so a product can be keyed by them.
	parts := []string{"adamic-runtime-v6", "os " + goruntime.GOOS, "arch " + goruntime.GOARCH, "compiler " + compiler, "version " + version,
		"archiver " + archiverName(compiler), "headers " + headers}
	for _, flag := range flags {
		parts = append(parts, "flag "+flag)
	}
	for _, file := range files {
		parts = append(parts, fmt.Sprintf("file %s %x", file.name, sha256.Sum256(file.contents)))
	}
	return parts
}

// partsKey hashes parts, each behind its length, so parts that join to the same bytes still key apart.
func partsKey(parts []string) string {
	hash := sha256.New()
	fmt.Fprintf(hash, "%d:", len(parts))
	for _, part := range parts {
		fmt.Fprintf(hash, "%d:", len(part))
		hash.Write([]byte(part))
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// runtimeRequest is the request for the archive these inputs compile, its parts read as the cache's key reads them.
func runtimeRequest(files []runtimeFile, flags []string, compiler string, version string, options Options) (RuntimeRequest, error) {
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		return RuntimeRequest{}, err
	}
	var features []string
	for _, feature := range runtimeFeatures {
		if slicesContain(flags, "-D"+feature+"=1") {
			features = append(features, feature)
		}
	}
	return RuntimeRequest{Parts: runtimeKeyParts(files, flags, compiler, version, headers), Features: features, Options: options,
		files: files, flags: flags, compiler: compiler}, nil
}

func cachedRuntime(files []runtimeFile, flags []string, compiler string, version string, cache string) (string, error) {
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		return "", err
	}
	directory := filepath.Join(cache, runtimeKey(files, flags, compiler, version, headers))
	library := filepath.Join(directory, "runtime.a")
	value, _ := runtimeBuilds.LoadOrStore(directory, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
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
	if err := os.Chmod(temporary, 0o755); err != nil {
		return "", err
	}
	if err := compileRuntime(files, flags, compiler, temporary, 1); err != nil {
		return "", err
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

// compileRuntime compiles the runtime's files with flags into directory, up to workers translation units at once, and
// archives them as runtime.a there, in the files' order whatever order they finished in.
func compileRuntime(files []runtimeFile, flags []string, compiler string, directory string, workers int) error {
	for _, feature := range runtimeFeatures {
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
	var sources []string
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0o644); err != nil {
			return err
		}
		if strings.HasSuffix(file.name, ".c") {
			sources = append(sources, file.name)
		}
	}
	if len(sources) == 0 {
		return fmt.Errorf("native: runtime has no C sources")
	}
	objects := make([]string, len(sources))
	failures := make([]error, len(sources))
	slots := make(chan struct{}, max(1, workers))
	var compiles sync.WaitGroup
	for index, source := range sources {
		slots <- struct{}{}
		compiles.Go(func() {
			defer func() { <-slots }()
			objects[index] = filepath.Join(directory, strings.TrimSuffix(source, ".c")+".o")
			// The build directory is random and the cache sits where each machine keeps it, so neither may reach an
			// object: mapped, and compiled from inside it, the archive's bytes are a function of its key (Oct 9: a native
			// product carried ~/.cache/adamic/runtime/.build-<random> in its debug info and never matched a rebuild).
			arguments := append(append(append([]string{}, flags...), prefixMaps(directory, "/adamic-runtime")...), "-c", source, "-o", objects[index])
			command := toolchain.CommandIn(directory, compiler, arguments...)
			if output, err := command.CombinedOutput(); err != nil {
				failures[index] = fmt.Errorf("native: compiling runtime %s: %w\n%s", source, err, output)
			}
		})
	}
	compiles.Wait()
	if err := errors.Join(failures...); err != nil {
		return err
	}
	// Each member carries no time: Apple's ar stamps the time it ran into a member's header, which ld's debug map copies
	// beside the member's name, so two builds of one program linked against runtimes archived a second apart differed
	// (TestBuildIsTheSameFromAnyDirectory on a Mac, Oct 12). ZERO_AR_DATE has it write zero; llvm-ar writes zero anyway.
	arguments := append([]string{"rcs", filepath.Join(directory, "runtime.a")}, objects...)
	archive := toolchain.Command(archiverName(compiler), arguments...)
	archive.Env = append(archive.Env, "ZERO_AR_DATE=1")
	if output, err := archive.CombinedOutput(); err != nil {
		return fmt.Errorf("native: archiving runtime: %w\n%s", err, output)
	}
	return nil
}

// prefixMaps maps directory to name in everything a compile records of a path (debug info, __FILE__), under each
// spelling of it: as given, and as the system resolves it, which clang may record instead (macOS's temporary and cache
// directories sit under /var, a link to /private/var, and a build there kept its directory, landable-12 on a Mac).
func prefixMaps(directory, name string) []string {
	maps := []string{"-ffile-prefix-map=" + directory + "=" + name}
	if real, err := filepath.EvalSymlinks(directory); err == nil && real != directory {
		maps = append(maps, "-ffile-prefix-map="+real+"="+name)
	}
	return maps
}

// headerScans holds each process's scan of the headers a runtime's compile reads, once per sources, flags and compiler:
// the system's headers don't change while a process runs.
var headerScans sync.Map

type headerScan struct {
	once    sync.Once
	headers string
	err     error
}

// headersRead is every header outside the runtime's own sources that compiling them reads, as the compiler lists its
// dependencies (-M): the C library's headers and the compiler's own, each by its path and the sha256 of what it holds.
// A traced build reads the runtime cache as no entry, as it reads Go's build cache (#vt46geg), and only an address of
// everything the compile read makes that safe: these headers are the one part the sources, flags and compiler don't name.
func headersRead(files []runtimeFile, flags []string, compiler string) (string, error) {
	value, _ := headerScans.LoadOrStore(runtimeKey(files, flags, compiler, "", ""), &headerScan{})
	scan := value.(*headerScan)
	scan.once.Do(func() { scan.headers, scan.err = scanHeaders(files, flags, compiler) })
	return scan.headers, scan.err
}

func scanHeaders(files []runtimeFile, flags []string, compiler string) (string, error) {
	directory, err := os.MkdirTemp("", "adamic-runtime-headers-")
	if err != nil {
		return "", fmt.Errorf("native: runtime headers: %w", err)
	}
	defer os.RemoveAll(directory)
	var sources []string
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(directory, file.name), file.contents, 0o644); err != nil {
			return "", fmt.Errorf("native: runtime headers: %w", err)
		}
		if strings.HasSuffix(file.name, ".c") {
			sources = append(sources, file.name)
		}
	}
	if len(sources) == 0 {
		return "", nil
	}
	command := toolchain.CommandIn(directory, compiler, append(append(append([]string{}, flags...), "-M"), sources...)...)
	rules, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return "", fmt.Errorf("native: listing the runtime's headers: %w\n%s", err, exit.Stderr)
		}
		return "", fmt.Errorf("native: listing the runtime's headers: %w", err)
	}
	seen := map[string]bool{}
	var headers []string
	for _, name := range dependencies(rules) {
		if !filepath.IsAbs(name) {
			name = filepath.Join(directory, name)
		}
		name = filepath.Clean(name)
		if seen[name] || strings.HasPrefix(name, directory+string(filepath.Separator)) {
			continue
		}
		seen[name] = true
		contents, err := os.ReadFile(name)
		if err != nil {
			return "", fmt.Errorf("native: runtime header %s: %w", name, err)
		}
		headers = append(headers, fmt.Sprintf("%s %x", name, sha256.Sum256(contents)))
	}
	sort.Strings(headers)
	return strings.Join(headers, "\n"), nil
}

// dependencies are the prerequisites of make rules as clang -M writes them: "object: source header ...", a line
// continued by a backslash, a space inside a path escaped by one.
func dependencies(rules []byte) []string {
	var names []string
	for _, line := range strings.Split(strings.ReplaceAll(string(rules), "\\\n", " "), "\n") {
		_, prerequisites, found := strings.Cut(line, ": ")
		if !found {
			continue
		}
		var name strings.Builder
		for index := 0; index < len(prerequisites); index++ {
			switch character := prerequisites[index]; {
			case character == '\\' && index+1 < len(prerequisites) && prerequisites[index+1] == ' ':
				name.WriteByte(' ')
				index++
			case character == ' ' || character == '\t':
				if name.Len() > 0 {
					names = append(names, name.String())
					name.Reset()
				}
			default:
				name.WriteByte(character)
			}
		}
		if name.Len() > 0 {
			names = append(names, name.String())
		}
	}
	return names
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
