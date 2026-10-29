package native

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/toolchain"
)

func TestRuntimeKeyIncludesEveryInput(t *testing.T) {
	t.Parallel()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	flags := Flags(Options{Sanitize: true, Count: true, slabs: true, cpu: "haswell"})
	key := runtimeKey(files, flags, "clang", "version 1", "")
	for index, flag := range flags {
		changed := append([]string{}, flags[:index]...)
		changed = append(changed, flags[index+1:]...)
		if runtimeKey(files, changed, "clang", "version 1", "") == key {
			t.Errorf("omitting %s did not change the key", flag)
		}
	}
	for index, file := range files {
		changed := append([]runtimeFile{}, files...)
		changed[index].contents = append([]byte{}, file.contents...)
		changed[index].contents[0] ^= 1
		if runtimeKey(changed, flags, "clang", "version 1", "") == key {
			t.Errorf("changing %s did not change the key", file.name)
		}
	}
	if runtimeKey(files, flags, "clang", "version 2", "") == key {
		t.Error("compiler version missing from key")
	}
	if runtimeKey(files, flags, "other-clang", "version 1", "") == key {
		t.Error("compiler path missing from key")
	}
	if runtimeKey(files, flags, "clang", "version 1", "/usr/include/stdio.h 0a") == key {
		t.Error("the system headers the compile reads missing from key")
	}
	changed := append([]string{}, flags...)
	changed[0], changed[1] = changed[1], changed[0]
	if runtimeKey(files, changed, "clang", "version 1", "") == key {
		t.Error("flag order missing from key")
	}
}

// Inputs that join to the same bytes must still key apart, or one runtime is served another's
// library. Each pair below keeps the count of parts the same, so only the boundaries differ.
func TestRuntimeKeyKeepsBoundaries(t *testing.T) {
	t.Parallel()
	files := []runtimeFile{{"a.c", []byte("int a;\n")}}
	if runtimeKey(files, []string{"a", "bc"}, "clang", "version 1", "") == runtimeKey(files, []string{"ab", "c"}, "clang", "version 1", "") {
		t.Error(`flags ["a" "bc"] and ["ab" "c"] share a key`)
	}
	if runtimeKey(files, nil, "clang", "version 1", "") == runtimeKey(files, nil, "clangv", "ersion 1", "") {
		t.Error(`compiler "clang" with version "version 1" and "clangv" with "ersion 1" share a key`)
	}
	split := []runtimeFile{{"a.c", []byte("bc")}}
	moved := []runtimeFile{{"a.cb", []byte("c")}}
	if runtimeKey(split, nil, "clang", "version 1", "") == runtimeKey(moved, nil, "clang", "version 1", "") {
		t.Error(`file "a.c" holding "bc" and "a.cb" holding "c" share a key`)
	}
}

// Count is checked in the linked runtime itself, not merely in a comparison of keys. Build an
// uncounted library first: a key missing ADAMIC_COUNT would reuse it and silently lose the report.
func TestRuntimeCacheKeepsCountFlags(t *testing.T) {
	t.Parallel()
	const source = "int main(void) { return 0; }\n"
	for _, count := range []bool{false, true, false} {
		binary := filepath.Join(t.TempDir(), "main")
		if err := Build(source, binary, Options{Count: count}); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(binary).CombinedOutput()
		if err != nil {
			t.Fatalf("run: %v\n%s", err, output)
		}
		want := ""
		if count {
			want = "adamic: counts: allocations 0 frees 0 retains 0 releases 0 peak 0 regions 0\n"
		}
		if string(output) != want {
			t.Errorf("count %v: got %q, want %q", count, output, want)
		}
	}
}

func smallRuntime(t *testing.T, sources fstest.MapFS, cache string, compiler string) string {
	t.Helper()
	files, err := readRuntime(sources, ".")
	if err != nil {
		t.Fatal(err)
	}
	library, err := cachedRuntime(files, Flags(Options{}), compiler, "test compiler", cache)
	if err != nil {
		t.Fatal(err)
	}
	return library
}

func runLibrary(t *testing.T, library string, want string) {
	t.Helper()
	directory := t.TempDir()
	source := filepath.Join(directory, "main.c")
	if err := os.WriteFile(source, []byte("#include <stdio.h>\nint answer(void);\nint main(void) { printf(\"%d\\n\", answer()); return 0; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "main")
	arguments := append([]string{"-o", binary, source}, RuntimeLinkFlags(library)...)
	if output, err := toolchain.Command("clang", arguments...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, output)
	}
	if got := runWithInput(t, "", binary); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Change the actual input after caching, and link both libraries: this catches hashing filenames
// alone, omitting headers, and compiling something other than the bytes that were hashed.
func TestRuntimeCacheRebuildsChangedSources(t *testing.T) {
	t.Parallel()
	cache := t.TempDir()
	sources := fstest.MapFS{
		"answer.c": &fstest.MapFile{Data: []byte("#include \"answer.h\"\nint answer(void) { return ANSWER + 0; }\n")},
		"answer.h": &fstest.MapFile{Data: []byte("#define ANSWER 1\n")},
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	first := smallRuntime(t, sources, cache, compiler)
	sources["answer.c"].Data = []byte("#include \"answer.h\"\nint answer(void) { return ANSWER + 1; }\n")
	second := smallRuntime(t, sources, cache, compiler)
	if first == second {
		t.Fatal("a changed runtime source reused the cached library")
	}
	sources["answer.h"].Data = []byte("#define ANSWER 3\n")
	third := smallRuntime(t, sources, cache, compiler)
	if second == third {
		t.Fatal("a changed runtime header reused the cached library")
	}
	runLibrary(t, first, "1\n")
	runLibrary(t, second, "2\n")
	runLibrary(t, third, "4\n")
}

// With ADAMIC_BUILD_OUTPUT_DIR set, a runtime the build cache already holds is found there, and nothing is written for
// it anywhere; a runtime it doesn't hold is built and published in the output directory, at the place it would have in
// the cache. Mutants: cachedRuntime looking only where it writes (the second build runs, into the output directory);
// cacheDirectory still the build cache's.
// Not parallel: t.Setenv.
func TestARuntimeTheCacheHoldsIsReadThereAndANewOneWrittenToTheOutputDirectory(t *testing.T) {
	cache, output := t.TempDir(), filepath.Join(t.TempDir(), "output")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	t.Setenv("ADAMIC_SANDBOX", "")
	t.Setenv(location.OutputDirectoryVariable, "")
	sources := fstest.MapFS{
		"answer.c": &fstest.MapFile{Data: []byte("#include \"answer.h\"\nint answer(void) { return ANSWER + 0; }\n")},
		"answer.h": &fstest.MapFile{Data: []byte("#define ANSWER 5\n")},
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	written, err := cacheDirectory("native-runtime")
	if err != nil {
		t.Fatal(err)
	}
	held := smallRuntime(t, sources, written, compiler)
	t.Setenv(location.OutputDirectoryVariable, output)
	if written, err = cacheDirectory("native-runtime"); err != nil || written != filepath.Join(output, "native-runtime") {
		t.Fatalf("with an output directory the runtime cache is written at %q (%v)", written, err)
	}
	if found := smallRuntime(t, sources, written, compiler); found != held {
		t.Fatalf("the runtime the cache holds was taken from %s, want %s", found, held)
	}
	if _, err := os.Stat(output); err == nil {
		t.Error("finding a runtime the cache holds wrote into the output directory")
	}
	sources["answer.h"].Data = []byte("#define ANSWER 6\n")
	built := smallRuntime(t, sources, written, compiler)
	if !strings.HasPrefix(built, written+string(filepath.Separator)) {
		t.Fatalf("a new runtime was published at %s, not in the output directory's runtime cache", built)
	}
	runLibrary(t, held, "5\n")
	runLibrary(t, built, "6\n")
}

// Many callers begin on an empty cache together. A compiler wrapper records actual compilations,
// so the test requires one build as well as an intact archive every caller can link.
// Not parallel: publish the executable wrapper before other tests fork.
func TestRuntimeCacheConcurrentBuilders(t *testing.T) {
	directory := t.TempDir()
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(directory, "clang")
	log := filepath.Join(directory, "compiled")
	// Quote paths for /bin/sh without relying on the shell's environment or global test settings.
	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }
	// Neither its version nor the listing of the headers its compile reads (-M) is a compilation.
	script := fmt.Sprintf("#!/bin/sh\ncase \" $* \" in *' --version '*|*' -M '*) ;; *) echo compiled >> %s ;; esac\nexec %s \"$@\"\n", quote(log), quote(compiler))
	if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	files := []runtimeFile{{"answer.c", []byte("int answer(void) { return 42; }\n")}}
	cache := filepath.Join(directory, "cache")
	const callers = 32
	libraries := make([]string, callers)
	failures := make([]error, callers)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for index := range callers {
		workers.Go(func() {
			<-start
			libraries[index], failures[index] = cachedRuntime(files, Flags(Options{}), wrapper, "test compiler", cache)
		})
	}
	close(start)
	workers.Wait()
	for index, failure := range failures {
		if failure != nil {
			t.Fatalf("caller %d: %v", index, failure)
		}
		if libraries[index] != libraries[0] {
			t.Fatal("callers got different libraries")
		}
	}
	compiled, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(compiled) != "compiled\n" {
		t.Fatalf("want one compilation, got %q", compiled)
	}
	for _, library := range libraries {
		runLibrary(t, library, "42\n")
	}
	if entries := publishedEntries(t, cache); len(entries) != 1 {
		t.Fatalf("want one published entry with no temporary builds, got %q", entries)
	}
}

// Processes do not share runtimeBuilds. Every one must see a complete archive and headers even
// when several independent builds race to publish the same directory.
func TestRuntimeCacheConcurrentProcesses(t *testing.T) {
	t.Parallel()
	const variable = "ADAMIC_TEST_RUNTIME_CACHE"
	files := []runtimeFile{
		{"answer.c", []byte("#include \"answer.h\"\nint answer(void) { return ANSWER; }\n")},
		{"answer.h", []byte("#define ANSWER 42\n")},
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	if cache := os.Getenv(variable); cache != "" {
		library, err := cachedRuntime(files, Flags(Options{}), compiler, "process test", cache)
		if err != nil {
			t.Fatal(err)
		}
		header, err := os.ReadFile(filepath.Join(filepath.Dir(library), "answer.h"))
		if err != nil || string(header) != "#define ANSWER 42\n" {
			t.Fatalf("published header: %q, %v", header, err)
		}
		runLibrary(t, library, "42\n")
		return
	}
	cache := filepath.Join(t.TempDir(), "cache")
	const callers = 8
	outputs := make([][]byte, callers)
	failures := make([]error, callers)
	var workers sync.WaitGroup
	for index := range callers {
		workers.Go(func() {
			command := exec.Command(os.Args[0], "-test.run=^TestRuntimeCacheConcurrentProcesses$")
			command.Env = append(os.Environ(), variable+"="+cache)
			outputs[index], failures[index] = command.CombinedOutput()
		})
	}
	workers.Wait()
	for index, failure := range failures {
		if failure != nil {
			t.Errorf("process %d: %v\n%s", index, failure, outputs[index])
		}
	}
	if entries := publishedEntries(t, cache); len(entries) != 1 {
		t.Fatalf("want one published entry with no temporary builds, got %q", entries)
	}
}

// publishedEntries are the names in a runtime cache other than the locks beside its entries (lockRuntime).
func publishedEntries(t *testing.T, cache string) []string {
	t.Helper()
	entries, err := os.ReadDir(cache)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".lock") {
			names = append(names, entry.Name())
		}
	}
	return names
}

// Processes asking for one runtime archive at once build it once: the rest wait on the lock beside its entry
// (lockRuntime), find what the builder published, and take the same bytes. Each asks from a process of its own, as
// Loom's product tests do, having scanned the headers its key reads before a common start, so all of them look for the
// archive together and miss. The build log counts the builds: one cArchive miss with its units' compiles, and for every
// other process a hit of 0 seconds, as a hit always logs, the wait never counted as compiling. The lock reaches neither
// the archive nor its key: one process alone on another cache builds the same bytes under the same key.
// Mutant: the lock dropped, leaving runtimeBuilds' mutex, which is one per process (a miss for each process).
func TestProcessesAskingForOneRuntimeBuildItOnce(t *testing.T) {
	t.Parallel()
	const variable = "ADAMIC_TEST_ONE_RUNTIME_BUILD"
	// Enough units that a build takes long enough for every process to miss inside it without the lock.
	var files []runtimeFile
	for index := range 16 {
		files = append(files, runtimeFile{fmt.Sprintf("unit%d.c", index), []byte(fmt.Sprintf("int unit%d(void) { return %d; }\n", index, index))})
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	if asked := os.Getenv(variable); asked != "" {
		cache, start, _ := strings.Cut(asked, "|")
		if _, err := headersRead(files, Flags(Options{}), compiler); err != nil {
			t.Fatal(err)
		}
		at, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Until(time.Unix(0, at)))
		library, err := cachedRuntime(files, Flags(Options{}), compiler, "one build test", cache)
		if err != nil {
			t.Fatal(err)
		}
		archive, err := os.ReadFile(library)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("runtime.a %s %x\n", filepath.Base(filepath.Dir(library)), sha256.Sum256(archive))
		return
	}
	directory := t.TempDir()
	// ask has processes ask for the archive at once from cache, each logging to log, and returns the key and the digest
	// of the archive each read, one line each.
	ask := func(cache string, log string, processes int) []string {
		t.Helper()
		// Every process has started and scanned its headers by then, so they ask together.
		start := time.Now().Add(3 * time.Second).UnixNano()
		outputs := make([][]byte, processes)
		failures := make([]error, processes)
		var workers sync.WaitGroup
		for index := range processes {
			workers.Go(func() {
				command := exec.Command(os.Args[0], "-test.run=^TestProcessesAskingForOneRuntimeBuildItOnce$", "-test.count=1")
				command.Env = append(os.Environ(), fmt.Sprintf("%s=%s|%d", variable, cache, start), "ADAMIC_BUILD_LOG="+log)
				outputs[index], failures[index] = command.Output()
			})
		}
		workers.Wait()
		var read []string
		for index, failure := range failures {
			if failure != nil {
				t.Fatalf("process %d: %v\n%s", index, failure, outputs[index])
			}
			for _, line := range strings.Split(string(outputs[index]), "\n") {
				if archive, found := strings.CutPrefix(line, "runtime.a "); found {
					read = append(read, archive)
				}
			}
		}
		if len(read) != processes {
			t.Fatalf("%d of %d processes reported the archive they read: %q", len(read), processes, read)
		}
		if entries := publishedEntries(t, cache); len(entries) != 1 {
			t.Fatalf("want one published entry with no temporary builds, got %q", entries)
		}
		return read
	}
	const processes = 4
	log := filepath.Join(directory, "build.log")
	together := ask(filepath.Join(directory, "together"), log, processes)
	for _, read := range together[1:] {
		if read != together[0] {
			t.Fatalf("processes asking for one archive read different ones: %q", together)
		}
	}
	written, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	outcomes := map[string]int{}
	for _, line := range strings.Split(strings.TrimSuffix(string(written), "\n"), "\n") {
		match := buildLogLine.FindStringSubmatch(line)
		if match == nil {
			t.Fatalf("%q isn't a build log line", line)
		}
		outcomes[match[1]+" "+match[4]]++
		if match[1] == "cArchive" && match[4] == "hit" && match[5] != "0.000" {
			t.Fatalf("%q, want a hit of 0 seconds however long it waited", match[0])
		}
	}
	if outcomes["cArchive miss"] != 1 || outcomes["cCompile miss"] != len(files) || outcomes["cArchive hit"] != processes-1 {
		t.Fatalf("%d processes asking for one archive built it %d times (%d compiles) and found it %d times, want once (%d) and %d",
			processes, outcomes["cArchive miss"], outcomes["cCompile miss"], outcomes["cArchive hit"], len(files), processes-1)
	}
	if alone := ask(filepath.Join(directory, "alone"), filepath.Join(directory, "alone.log"), 1); alone[0] != together[0] {
		t.Fatalf("one process alone built %s, and %d together %s: want the same key and bytes", alone[0], processes, together[0])
	}
	t.Logf("%d processes asked for one archive at once: one built it, %d found it, all read the bytes one process alone builds", processes, processes-1)
}

// A runtime is addressed by the system headers its compile reads, by what they hold: a C library or compiler upgrade
// that changes one makes another runtime, never the cached one (#vt46geg: a traced build reads the runtime cache as no
// entry, which only an address of everything the compile read makes safe). Mutants: runtimeKey not hashing headers;
// scanHeaders skipping what is outside the sources; headersRead's memo shared across flags.
// Not parallel: it clears headerScans, the process's memo every runtime build reads.
func TestARuntimeIsKeyedByTheSystemHeadersItsCompileReads(t *testing.T) {
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	system := t.TempDir()
	header := filepath.Join(system, "answer_system.h")
	files := []runtimeFile{{"answer.c", []byte("#include <answer_system.h>\nint answer(void) { return ANSWER; }\n")}}
	flags := append(Flags(Options{}), "-isystem", system)
	cache := t.TempDir()
	build := func(value string) string {
		t.Helper()
		if err := os.WriteFile(header, []byte("#define ANSWER "+value+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		// A new process scans again; within one, the system's headers are taken as fixed.
		headerScans.Clear()
		headers, err := headersRead(files, flags, compiler)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(headers, header) {
			t.Fatalf("the headers the compile reads don't name %s:\n%s", header, headers)
		}
		library, err := cachedRuntime(files, flags, compiler, "test compiler", cache)
		if err != nil {
			t.Fatal(err)
		}
		return filepath.Dir(library)
	}
	first, second := build("42"), build("43")
	if first == second {
		t.Fatalf("a changed system header kept the runtime %s", first)
	}
	if again := build("42"); again != first {
		t.Fatalf("the same system header made another runtime: %s, then %s", first, again)
	}
	// The scan with -isystem is still held; the same sources without it are another scan, which can't find the header.
	plain, err := headersRead(files, Flags(Options{}), compiler)
	if err == nil || strings.Contains(plain, header) {
		t.Fatalf("without -isystem the compile can't find the header, so listing it must fail: %q, %v", plain, err)
	}
}

// clang -M writes make rules: a backslash continues a line and escapes a space inside a path.
func TestDependenciesReadClangsRules(t *testing.T) {
	t.Parallel()
	rules := "answer.o: answer.c /usr/include/stdio.h \\\n  /opt/with\\ space/x.h\nother.o: other.c\n"
	got := dependencies([]byte(rules))
	want := []string{"answer.c", "/usr/include/stdio.h", "/opt/with space/x.h", "other.c"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// An archive is the same bytes however many of its units compiled at once: a product compiles them four at a time and
// the cache one at a time, and either must be the other's archive. Mutant: archiving the objects as they finish.
func TestARuntimeIsTheSameOnAnyNumberOfWorkers(t *testing.T) {
	t.Parallel()
	var files []runtimeFile
	for index := range 8 {
		files = append(files, runtimeFile{fmt.Sprintf("unit%d.c", index), []byte(fmt.Sprintf("int unit%d(void) { return %d; }\n", index, index))})
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	var archives [][]byte
	for _, workers := range []int{1, 4, 8} {
		directory := t.TempDir()
		if err := compileRuntime(files, Flags(Options{}), compiler, directory, workers, nil, ""); err != nil {
			t.Fatal(err)
		}
		archive, err := os.ReadFile(filepath.Join(directory, "runtime.a"))
		if err != nil {
			t.Fatal(err)
		}
		archives = append(archives, archive)
	}
	for index, archive := range archives[1:] {
		if !bytes.Equal(archive, archives[0]) {
			t.Errorf("the archive compiled on %d workers differs from the one compiled on 1", []int{4, 8}[index])
		}
	}
}

// A supplier (SupplyRuntimes) gives every archive of the embedded runtime, asked once per key however many builds
// want it, and its error is the build's; another checkout's runtime never goes to it. In a child process, since the
// supplier is the process's. Mutants: the supplier asked for a directory's runtime; asked again for a key it gave.
func TestASuppliedRuntimeIsAskedOncePerKey(t *testing.T) {
	t.Parallel()
	const variable = "ADAMIC_TEST_SUPPLIED_RUNTIME"
	if directory := os.Getenv(variable); directory != "" {
		var asked atomic.Int64
		SupplyRuntimes(func(request RuntimeRequest) (string, error) {
			asked.Add(1)
			if request.Options.Count {
				return "", fmt.Errorf("no counted runtime here")
			}
			held := filepath.Join(directory, fmt.Sprint(asked.Load()))
			if err := os.MkdirAll(held, 0o755); err != nil {
				return "", err
			}
			return held, os.WriteFile(filepath.Join(held, "runtime.a"), nil, 0o644)
		})
		libraries := make([]string, 8)
		var callers sync.WaitGroup
		for index := range libraries {
			callers.Go(func() {
				library, err := RuntimeLibrary("", Options{})
				if err != nil {
					t.Error(err)
				}
				libraries[index] = library
			})
		}
		callers.Wait()
		for _, library := range libraries {
			if library != filepath.Join(directory, "1", "runtime.a") {
				t.Fatalf("got %s, want the supplier's first archive", library)
			}
		}
		if _, err := RuntimeLibrary("", Options{Count: true}); err == nil || !strings.Contains(err.Error(), "no counted runtime here") {
			t.Fatalf("the supplier's error isn't the build's: %v", err)
		}
		checkout := t.TempDir()
		if err := os.WriteFile(filepath.Join(checkout, "answer.c"), []byte("int answer(void) { return 42; }\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := RuntimeLibrary(checkout, Options{}); err != nil {
			t.Fatal(err)
		}
		if got := asked.Load(); got != 2 {
			t.Fatalf("the supplier was asked %d times, want 2 (the plain runtime once, the counted once)", got)
		}
		return
	}
	command := exec.Command(os.Args[0], "-test.run=^TestASuppliedRuntimeIsAskedOncePerKey$")
	command.Env = append(os.Environ(), variable+"="+t.TempDir(), "ADAMIC_BUILD_CACHE_DIR="+t.TempDir(), "ADAMIC_BUILD_OUTPUT_DIR=")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
}

// A runtime's portable parts are one list however the compiler was reached: a runner's $ADAMIC_TOOLS/bin/clang and
// Workshop's ~/.adamic-tools/bin/clang, a link to that directory, name its resource headers and a sysroot inside its
// installation as <compiler>/..., and leave out the compiler and its version, which a product keys by its report. A
// header outside the installation keeps its path. Mutant: portableKeyParts naming paths as reached.
func TestARuntimesPortablePartsAreTheSameThroughEitherLink(t *testing.T) {
	t.Parallel()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tools := filepath.Join(home, "adamic-tools")
	for _, directory := range []string{"llvm/bin", "llvm/lib/clang/20/include", "bin"} {
		if err := os.MkdirAll(filepath.Join(tools, directory), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for name, mode := range map[string]os.FileMode{"llvm/bin/clang-20": 0o755, "llvm/lib/clang/20/include/stddef.h": 0o644} {
		if err := os.WriteFile(filepath.Join(tools, name), []byte("#!/bin/sh\n"), mode); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(tools, "llvm/bin/clang-20"), filepath.Join(tools, "bin/clang")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("adamic-tools", filepath.Join(home, ".adamic-tools")); err != nil {
		t.Fatal(err)
	}
	files := []runtimeFile{{"answer.c", []byte("int answer(void) { return 42; }\n")}}
	parts := func(root string) []string {
		compiler := filepath.Join(root, "bin/clang")
		headers := filepath.Join(root, "llvm/lib/clang/20/include/stddef.h") + " 0a\n/usr/include/stdio.h 0b"
		return portableKeyParts(files, append(Flags(Options{}), "--sysroot="+filepath.Join(root, "llvm/lib")), compiler, headers)
	}
	runner, workshop := parts(tools), parts(filepath.Join(home, ".adamic-tools"))
	if strings.Join(runner, "\n") != strings.Join(workshop, "\n") {
		t.Fatalf("one installation through two links gave\n%s\nand\n%s", strings.Join(runner, "\n"), strings.Join(workshop, "\n"))
	}
	joined := strings.Join(runner, "\n")
	for _, want := range []string{"<compiler>/lib/clang/20/include/stddef.h 0a", "/usr/include/stdio.h 0b", "flag --sysroot=<compiler>/lib"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the portable parts don't hold %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, home) || strings.Contains(joined, "compiler ") || strings.Contains(joined, "version ") {
		t.Errorf("the portable parts name the compiler or where it was reached:\n%s", joined)
	}
}
