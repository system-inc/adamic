package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/toolchain"
)

// splitProbe is the smallest program splitC takes: every declaration static but main's.
const splitProbe = "#include \"adamic.h\"\nstatic int adamic_answer(void);\nstatic int adamic_answer(void) { return 0; }\nint main(void) { return adamic_answer(); }\n"

// The native caches (the runtime's, the split units', the checker runtime's) live under the build cache, never under
// HOME on their own account: a sandboxed build names its cache and writes nothing in the home directory (#m27xj48). A
// child builds through each with HOME empty and read-only, no XDG_ variable, and the build cache this process uses, so
// its runtimes are found as they are here; the child must pass and HOME must stay empty. The units cache is on in the
// child (no ADAMIC_GATE_UNCACHED), so a unit is written to or found in it. Mutants: os.UserCacheDir back at any one of
// runtimeLibrary's, buildUnitsWithLibrary's or splitTSGoRuntime's cache, and that build fails writing under HOME.
func TestNativeCachesWriteNothingUnderHome(t *testing.T) {
	t.Parallel()
	const variable = "ADAMIC_TEST_NATIVE_HOME"
	if os.Getenv(variable) == "1" {
		if _, err := RuntimeLibrary("", Options{}); err != nil {
			t.Fatal(err)
		}
		if err := Build(splitProbe, filepath.Join(t.TempDir(), "split"), Options{Split: true, Jobs: 1}); err != nil {
			t.Fatal(err)
		}
		if _, err := splitTSGoRuntime(splitProbe, Options{}); err != nil {
			t.Fatal(err)
		}
		return
	}
	cache, err := location.BuildCache()
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	if err := os.Chmod(home, 0o500); err != nil {
		t.Fatal(err)
	}
	// Cleanups run last first: HOME is writable again before t.TempDir removes it.
	t.Cleanup(func() { os.Chmod(home, 0o700) })
	command := exec.Command(os.Args[0], "-test.run=^TestNativeCachesWriteNothingUnderHome$")
	command.Env = []string{variable + "=1", "HOME=" + home, "ADAMIC_BUILD_CACHE_DIR=" + cache, "PATH=" + os.Getenv("PATH")}
	// The child's builds resolve their tools as this process's do: the same tools directory, sandbox and Xcode.
	for _, name := range []string{"TMPDIR", "ADAMIC_TOOLS", "ADAMIC_SANDBOX", "DEVELOPER_DIR"} {
		if value := os.Getenv(name); value != "" {
			command.Env = append(command.Env, name+"="+value)
		}
	}
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("native builds under an empty read-only HOME: %v\n%s", err, output)
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		t.Errorf("a native build wrote %s under HOME", entry.Name())
	}
}

// A checker build's binary names no place the checker archive was on this machine: clang links it through a link in
// its build directory, and on macOS ld's debug map, which names an archive's members by the archive's path, is given
// that path relative (#m27xj48: the debug map named the archive's absolute path, so a binary differed with where it was
// built). The archive here is stubs of the checker ABI, compiled with debug info as Go's c-archive is, under a
// directory no build could name by chance; both BuildTSGo and BuildSplitTSGo link it. On Linux the linker copies the
// DWARF in and names no input, so this holds there with or without the link. Mutants (on macOS): the archive's
// absolute path back on BuildTSGo's link line, or on buildUnitsWithLibrary's; objectPrefixFlags off.
func TestCheckerBuildsNameNoPlaceOfTheArchive(t *testing.T) {
	t.Parallel()
	compiler, err := toolchain.Path(compilerName(Options{}))
	if err != nil {
		t.Fatal(err)
	}
	place := filepath.Join(t.TempDir(), "checker-archive-place")
	if err := os.Mkdir(place, 0o755); err != nil {
		t.Fatal(err)
	}
	var stubs strings.Builder
	var names []string
	for _, match := range regexp.MustCompile(`\b(tsgo_[a-z_]+)\(`).FindAllStringSubmatch(string(bridge.Header), -1) {
		if !slices.Contains(names, match[1]) {
			names = append(names, match[1])
			stubs.WriteString("void " + match[1] + "(void) {}\n")
		}
	}
	if len(names) == 0 {
		t.Fatal("the checker header declares no tsgo_ function")
	}
	if err := os.WriteFile(filepath.Join(place, "stubs.c"), []byte(stubs.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	// The stubs' own debug info names where they were compiled as /checker-stubs, as Go's names no place of this
	// build's, so only the link can put the archive's directory in a binary.
	compile := toolchain.CommandIn(place, compiler, append(prefixMaps(place, "/checker-stubs"), "-g", "-c", "stubs.c", "-o", "stubs.o")...)
	if output, err := compile.CombinedOutput(); err != nil {
		t.Fatalf("compiling the stubs: %v\n%s", err, output)
	}
	archiver := toolchain.CommandIn(place, archiverName(compiler), "rcs", "tsgo.a", "stubs.o")
	if output, err := archiver.CombinedOutput(); err != nil {
		t.Fatalf("archiving the stubs: %v\n%s", err, output)
	}
	archive := filepath.Join(place, "tsgo.a")
	spellings := [][]byte{[]byte(place)}
	if real, err := filepath.EvalSymlinks(place); err == nil && real != place {
		spellings = append(spellings, []byte(real))
	}
	builds := map[string]func(output string) error{
		"BuildTSGo":      func(output string) error { return BuildTSGo(splitProbe, output, archive, Options{}) },
		"BuildSplitTSGo": func(output string) error { return BuildSplitTSGo(splitProbe, output, archive, Options{Jobs: 1}) },
	}
	for _, name := range []string{"BuildTSGo", "BuildSplitTSGo"} {
		binary := filepath.Join(t.TempDir(), "checker")
		if err := builds[name](binary); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		content, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		// On macOS the debug map names the stubs' object, so the archive was linked with its debug info, and names it
		// by the link's name alone, a whole string of the string table: neither the archive's place nor the build
		// directory's, which -oso_prefix takes off.
		if goruntime.GOOS == "darwin" && !bytes.Contains(content, []byte("\x00"+checkerArchive+"(stubs.o)\x00")) {
			t.Errorf("%s: the binary's debug map doesn't name the stubs as %s(stubs.o)", name, checkerArchive)
		}
		for _, spelling := range spellings {
			if bytes.Contains(content, spelling) {
				t.Errorf("%s: the binary names the archive's directory %s", name, spelling)
			}
		}
	}
}

// A sanitized split build's binary names no place of this machine's: its units' objects and the runtime library come
// from the build cache and the link runs in a scratch directory, and on macOS ld's debug map names each object it links
// with debug info (a sanitized build's all have it) by its path (#m27xj48: the build cache's path, 57 times). Linked by
// a link's name in the scratch directory, which -oso_prefix takes off, the debug map names each by that name alone. On
// Linux the linker names no input, so this holds there either way. Mutants (on macOS): the objects' paths back on the
// link line, or the runtime library's; objectPrefixFlags off.
func TestSplitBuildsNameNoPlaceOfTheirInputs(t *testing.T) {
	t.Parallel()
	cache, err := location.BuildCache()
	if err != nil {
		t.Fatal(err)
	}
	spellings := [][]byte{[]byte(cache), []byte("adamic-units-")}
	if real, err := filepath.EvalSymlinks(cache); err == nil && real != cache {
		spellings = append(spellings, []byte(real))
	}
	binary := filepath.Join(t.TempDir(), "split")
	if err := Build(splitProbe, binary, Options{Split: true, Jobs: 2, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	// The debug map names a unit's object and a runtime member, so both were linked with their debug info, each by the
	// link's name alone (unit-0 holds only declarations, so it has none).
	if goruntime.GOOS == "darwin" {
		for _, linked := range []*regexp.Regexp{regexp.MustCompile(`\x00unit-[0-9]+\.o\x00`), regexp.MustCompile(`\x00runtime/[^\x00/]+\.a\([^\x00/]+\.o\)\x00`)} {
			if !linked.Match(content) {
				t.Errorf("the binary's debug map names nothing as %s", linked)
			}
		}
	}
	for _, spelling := range spellings {
		if count := bytes.Count(content, spelling); count > 0 {
			t.Errorf("the binary names %s %d times", spelling, count)
		}
	}
}
