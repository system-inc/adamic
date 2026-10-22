package native

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/toolchain"
)

func TestTSGoCacheRelease(t *testing.T) { t.Parallel(); tsgoCacheWitness(t, Options{}) }
func TestTSGoCacheShippedRelease(t *testing.T) {
	t.Parallel()
	tsgoCacheWitness(t, Options{Release: true})
}
func TestTSGoCacheSanitized(t *testing.T) { t.Parallel(); tsgoCacheWitness(t, Options{Sanitize: true}) }

// The fixture exercises the include-order features without calling the checker ABI.
// Stub symbols permit a real checker link while keeping each test's setup bounded.
func tsgoCacheWitness(t *testing.T, options Options) {
	t.Helper()
	fixture, err := filepath.Abs("../oracle/testdata/arguments_length_extended.a")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{fixture})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	source := "#include \"tsgo_runtime.h\"\n" + C(program)
	node := toolchain.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", fixture)
	expected, err := node.CombinedOutput()
	if err != nil {
		t.Fatalf("source Node: %v\n%s", err, expected)
	}
	directory := t.TempDir()
	var stubs strings.Builder
	seen := map[string]bool{}
	for _, match := range regexp.MustCompile(`\b(tsgo_[a-z_]+)\(`).FindAllStringSubmatch(string(bridge.Header), -1) {
		if !seen[match[1]] {
			seen[match[1]] = true
			stubs.WriteString("void " + match[1] + "(void) {}\n")
		}
	}
	if len(seen) == 0 {
		t.Fatal("checker ABI declares no symbols")
	}
	if err := os.WriteFile(filepath.Join(directory, "stubs.c"), []byte(stubs.String()), 0644); err != nil {
		t.Fatal(err)
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	compile := toolchain.CommandIn(directory, compiler, "-c", "stubs.c", "-o", "stubs.o")
	if output, err := compile.CombinedOutput(); err != nil {
		t.Fatalf("stubs: %v\n%s", err, output)
	}
	archive := toolchain.CommandIn(directory, archiverName(compiler), "rcs", "checker.a", "stubs.o")
	if output, err := archive.CombinedOutput(); err != nil {
		t.Fatalf("archive: %v\n%s", err, output)
	}
	library, err := tsgoRuntimeLibrary(source, options)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(library)
	if err != nil {
		t.Fatal(err)
	}
	var first []byte
	for _, round := range []string{"first", "warm"} {
		binary := filepath.Join(directory, round, "program")
		if err := os.Mkdir(filepath.Dir(binary), 0755); err != nil {
			t.Fatal(err)
		}
		if err := BuildTSGo(source, binary, filepath.Join(directory, "checker.a"), options); err != nil {
			t.Fatal(err)
		}
		run := toolchain.Command(binary)
		run.Env = append(run.Env, "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		output, err := run.CombinedOutput()
		if err != nil || !bytes.Equal(output, expected) {
			t.Fatalf("%s: %v output %q, Node %q", round, err, output, expected)
		}
		content, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		for _, place := range []string{directory, filepath.Dir(library), "adamic-tsgo-", ".build-"} {
			if bytes.Contains(content, []byte(place)) {
				t.Errorf("%s binary names build place %q", round, place)
			}
		}
		if first == nil {
			first = content
		} else if !bytes.Equal(first, content) {
			t.Error("cold and warm binaries differ")
		}
	}
	warm, err := tsgoRuntimeLibrary(source, options)
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(warm)
	if err != nil {
		t.Fatal(err)
	}
	if library != warm || !os.SameFile(info, after) || !info.ModTime().Equal(after.ModTime()) {
		t.Error("warm build replaced the cached runtime")
	}
	if !t.Failed() {
		t.Logf("first and warm builds agree with source Node: %q; binary bytes and runtime archive unchanged", expected)
	}
}
