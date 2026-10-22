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
	"time"

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/ir"
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
// it. Parts are everything its key reads, as runtimeKey hashes them (runtimeKeyParts), spelled as this machine reaches
// them. Portable are the same parts as every machine spells them, for a key that travels: Compiler and Archiver are
// left for the caller to key by what they report, and every path inside the compiler's installation is named from
// it, so a runner reaching clang through $ADAMIC_TOOLS/bin keys as Workshop does through ~/.adamic-tools/bin (#94aa17s's
// class: typeaware's native keyed apart on the two). Features and Options are what asked for it.
type RuntimeRequest struct {
	Parts    []string
	Portable []string
	Compiler string
	Archiver string
	Features []string
	Options  Options
	files    []runtimeFile
	flags    []string
}

// Build compiles the archive into directory as the runtime cache holds one, runtime.a beside the runtime's files,
// compiling up to workers of its translation units at once: one where a test waits on it, as the cache compiles, and
// more where a product build is given the CPUs for it. Each compile holds a token of jobserver, the one it's handed and
// never the process's, and a nil one bounds none (#yzwm4rj): so the archive's build reads no state that reaches its
// bytes from outside its own initializers, and its key, which the jobserver never enters, is all its bytes depend on.
// The one outside value it reads is $ADAMIC_BUILD_LOG, which only names where its timing lines go (buildLog).
func (request RuntimeRequest) Build(directory string, workers int, jobserver *Jobserver) error {
	return compileRuntime(request.files, request.flags, request.Compiler, directory, workers, jobserver, partsKey(request.Parts))
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
var runtimeFeatures = []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST", "ADAMIC_CLOSURE_DISCARDS"}

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

// SourceFlags returns compiler flags matching the runtime features in emitted C.
func SourceFlags(source string, options Options) []string {
	return sourceFlags(source, options)
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
	return runtimeLibraryFromInputs(files, flags, compiler, version, options, directory == "", "native-runtime")
}

func runtimeLibraryFromInputs(files []runtimeFile, flags []string, compiler, version string, options Options, supplied bool, namespace string) (string, error) {
	if supplier := runtimeSupplier.Load(); supplier != nil && supplied {
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
	cache, err := cacheDirectory(namespace)
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
	// Compiler harnesses compile these files directly. Ordinary archives need
	// neither their empty translation units nor the large compiler-only tables.
	if !slicesContain(extraFlags, "-DADAMIC_REGEXP_RUNTIME_COMPILER=1") {
		kept := files[:0]
		for _, file := range files {
			if !strings.HasPrefix(file.name, "regexp_compile") {
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

// cacheDirectory is where a native cache named name is written: under the build cache, or the output directory of a
// process that only reads the cache (location.Output), never under this machine's home on native's own account, so a
// sandboxed build that names its cache writes nothing under HOME. An entry is looked for there first, then at the same
// place under the build cache (location.ReadPlaces).
func cacheDirectory(name string) (string, error) {
	cache, err := location.Output()
	if err != nil {
		return "", fmt.Errorf("native: cache directory: %w", err)
	}
	return filepath.Join(cache, name), nil
}

// cachedFile is the first place file, under a native cache's entry where this process writes it, is a regular file at:
// there, then at the same place under the build cache (location.ReadPlaces).
func cachedFile(file string) (string, bool) {
	for _, place := range location.ReadPlaces(file) {
		if info, err := os.Stat(place); err == nil && info.Mode().IsRegular() {
			return place, true
		}
	}
	return "", false
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

// RuntimeSourceDigest is the sha256 of every file of the embedded runtime, named and by what it holds, every feature's
// units included: the runtime as its sources say it, with nothing of the machine or the compiler that builds it, so it is
// the same on every machine. internal/oracle's counts table keys each row by it (countDigest).
func RuntimeSourceDigest() (string, error) {
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		return "", fmt.Errorf("native: runtime: %w", err)
	}
	parts := []string{"adamic-runtime-sources-v1"}
	for _, file := range files {
		parts = append(parts, fmt.Sprintf("file %s %x", file.name, sha256.Sum256(file.contents)))
	}
	return partsKey(parts), nil
}

// runtimeKey addresses a compiled runtime by everything its compile reads (runtimeKeyParts).
func runtimeKey(files []runtimeFile, flags []string, compiler string, version string, headers string) string {
	return partsKey(runtimeKeyParts(files, flags, compiler, version, headers))
}

// runtimeKeyParts is everything a compiled runtime's compile reads: the platform, the compiler and its version, the
// archiver beside it, and its sources (runtimeSourceParts): the system headers the compile reads (headersRead), so a
// runtime compiled against other C library or compiler headers is never taken for this one, the flags, and the
// runtime's files by what they hold. A supplied runtime's product is keyed by the same parts as every machine spells
// them (portableKeyParts), so the two keys can't move apart.
func runtimeKeyParts(files []runtimeFile, flags []string, compiler string, version string, headers string) []string {
	parts := append(runtimePlatformParts(), "compiler "+compiler, "version "+version, "archiver "+archiverName(compiler))
	return append(parts, runtimeSourceParts(files, flags, headers)...)
}

// runtimePlatformParts open every runtime key: its version and the platform it is compiled for.
func runtimePlatformParts() []string {
	// v2: objects no longer carry the build directory (-ffile-prefix-map), so a v1 archive built without it is never reused.
	// v3: the map covers the directory's resolved spelling too, so a v2 archive a Mac built under /private/var isn't reused.
	// v4: the archive's members carry no time (ZERO_AR_DATE), so a v3 archive Apple's ar stamped isn't reused.
	// v5: clang and ar run under toolchain.Env, not the caller's environment, so a v4 archive built under whatever the
	// caller had set (CPATH, SDKROOT, CCC_OVERRIDE_OPTIONS) isn't reused.
	// v6: the key is a list of named parts, each file by the sha256 of what it holds, so a product can be keyed by them.
	return []string{"adamic-runtime-v6", "ir-tags " + ir.TypeTagFingerprint(), "os " + goruntime.GOOS, "arch " + goruntime.GOARCH}
}

// runtimeSourceParts are what a runtime key reads of what is compiled: the system headers the compile reads, every
// flag, and every file of the runtime by what it holds.
func runtimeSourceParts(files []runtimeFile, flags []string, headers string) []string {
	parts := []string{"headers " + headers}
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
	return RuntimeRequest{Parts: runtimeKeyParts(files, flags, compiler, version, headers), Portable: portableKeyParts(files, flags, compiler, headers),
		Compiler: compiler, Archiver: archiverName(compiler), Features: features, Options: options, files: files, flags: flags}, nil
}

// portableKeyParts are runtimeKeyParts as every machine spells them: without the compiler, its version and the
// archiver, which a caller keys by their reports, and with each header and flag that names a path inside the
// compiler's installation (clang's resource headers, a WASI sysroot) naming it from there as <compiler>/..., however
// this machine reached the compiler. A path outside it (the C library's headers) is the same path on every machine.
func portableKeyParts(files []runtimeFile, flags []string, compiler string, headers string) []string {
	installation := compilerInstallation(compiler)
	relocate := func(path string) string {
		if real, err := filepath.EvalSymlinks(path); err == nil {
			path = real
		}
		if installation != "" {
			if relative, err := filepath.Rel(installation, path); err == nil && relative != ".." && !strings.HasPrefix(relative, "../") {
				return "<compiler>/" + filepath.ToSlash(relative)
			}
		}
		return path
	}
	var lines []string
	for _, line := range strings.Split(headers, "\n") {
		if path, digest, found := strings.Cut(line, " "); found {
			line = relocate(path) + " " + digest
		}
		lines = append(lines, line)
	}
	sort.Strings(lines)
	var relocated []string
	for _, flag := range flags {
		if sysroot, found := strings.CutPrefix(flag, "--sysroot="); found && sysroot != "" {
			flag = "--sysroot=" + relocate(sysroot)
		}
		relocated = append(relocated, flag)
	}
	return append(runtimePlatformParts(), runtimeSourceParts(files, relocated, strings.Join(lines, "\n"))...)
}

// compilerInstallation is the directory the compiler is installed in, as the system resolves it: the parent of the
// bin directory its program sits in (<ADAMIC_TOOLS>/llvm for <ADAMIC_TOOLS>/bin/clang, a link to llvm/bin/clang-20).
// "" when it can't be told, or is the root itself, which would name every path.
func compilerInstallation(compiler string) string {
	real, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		return ""
	}
	installation := filepath.Dir(filepath.Dir(real))
	if installation == filepath.Dir(installation) {
		return ""
	}
	return installation
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
	if found, ok := cachedFile(library); ok {
		buildLog("cArchive", "runtime.a", filepath.Base(directory), "hit", 0)
		return found, nil
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
	jobserver, err := defaultJobserver()
	if err != nil {
		return "", err
	}
	if err := compileRuntime(files, flags, compiler, temporary, 1, jobserver, filepath.Base(directory)); err != nil {
		return "", err
	}
	// Publish everything together. A competing process may already have published this key; its
	// identical complete entry wins, while this process removes only its own temporary directory.
	if err := os.Rename(temporary, directory); err != nil {
		found, ok := cachedFile(library)
		if !ok {
			return "", fmt.Errorf("native: publishing runtime: %w", err)
		}
		return found, nil
	}
	return library, nil
}

// compileRuntime compiles the runtime's files with flags into directory, up to workers translation units at once, each
// holding a token of jobserver (nil bounds none, and is never the process's), and archives them as runtime.a there, in
// the files' order whatever order they finished in. key is the archive's, which names each of its compiles and its
// archiving in the build log (buildLog).
func compileRuntime(files []runtimeFile, flags []string, compiler string, directory string, workers int, jobserver *Jobserver, key string) error {
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
			var output []byte
			var clang time.Duration
			if err := jobserver.Hold(func() error {
				return timed(&clang, func() (err error) { output, err = command.CombinedOutput(); return err })
			}); err != nil {
				failures[index] = fmt.Errorf("native: compiling runtime %s: %w\n%s", source, err, output)
				return
			}
			buildLog("cCompile", "runtime/"+source, key, "miss", clang)
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
	var archiving time.Duration
	var output []byte
	if err := timed(&archiving, func() (err error) { output, err = archive.CombinedOutput(); return err }); err != nil {
		return fmt.Errorf("native: archiving runtime: %w\n%s", err, output)
	}
	buildLog("cArchive", "runtime.a", key, "miss", archiving)
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
	jobserver, err := defaultJobserver()
	if err != nil {
		return "", err
	}
	command := toolchain.CommandIn(directory, compiler, append(append(append([]string{}, flags...), "-M"), sources...)...)
	var rules []byte
	err = jobserver.Hold(func() (err error) { rules, err = command.Output(); return err })
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

// buildLog appends one line about a clang step to $ADAMIC_BUILD_LOG when it is set, beside buildcache's census lines
// (internal/buildcache's record): '<kind> <name> <key12> hit|miss|off <seconds>'. kind is cCompile for a translation
// unit, cArchive for the runtime's archive, cLink for a link, and cBuild for one clang call that compiles a program's one
// unit and links it (a one-unit build off macOS), so neither half is taken for the other. name is the unit, archive or
// output it made; key12 is the first 12 of the cache key the step was looked up by, or - when nothing keys it. hit is an
// object or archive the cache held, miss one made into the cache, off a step that read no cache. seconds is how long the
// tool ran for it, a wait for a compile token left out: a split unit's hit is its preprocessing, an archive's hit runs
// nothing and is 0, and an archive's miss is ar alone, its units' compiles each on their own cCompile line. Loom reads
// these lines to put C compile and link on a run's critical path (#s0jdsaa); every reader of the census matches its own
// kind of line, so these pass them by.
//
// It lives in this file because compileRuntime reaches it, and an archive's product declares the files its build
// reaches (runtimeproducts' runtimeFiles, held by TestTheRuntimeBuildReachesOnlyItsDeclaredFiles).
//
// The line never reaches what is built: nothing it writes is an input to clang, and a log that can't be opened writes
// nothing and fails nothing, as record's doesn't.
func buildLog(kind, name, key, outcome string, seconds time.Duration) {
	path := os.Getenv("ADAMIC_BUILD_LOG")
	if path == "" {
		return
	}
	if key == "" {
		key = "-"
	}
	// Every whitespace character in name becomes _, so an output path holding a tab or a newline can neither split the
	// line nor write a second one.
	line := fmt.Sprintf("%s %s %s %s %.3f\n", kind, strings.Join(strings.Fields(name), "_"), key[:min(12, len(key))], outcome, seconds.Seconds())
	if file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
		file.WriteString(line)
		file.Close()
	}
}

// timed runs step and adds how long it took to total.
func timed(total *time.Duration, step func() error) error {
	started := time.Now()
	err := step()
	*total += time.Since(started)
	return err
}
