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
	if temporary := os.Getenv("TMPDIR"); temporary != "" {
		command.Env = append(command.Env, "TMPDIR="+temporary)
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
	compiler, err := exec.LookPath(compilerName(Options{}))
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
	compile := exec.Command(compiler, append(prefixMaps(place, "/checker-stubs"), "-g", "-c", "stubs.c", "-o", "stubs.o")...)
	compile.Dir = place
	if output, err := compile.CombinedOutput(); err != nil {
		t.Fatalf("compiling the stubs: %v\n%s", err, output)
	}
	archiver := exec.Command(archiverName(compiler), "rcs", "tsgo.a", "stubs.o")
	archiver.Dir = place
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
