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

func TestTSGoCacheRelease(test *testing.T) { test.Parallel(); tsgoCacheWitness(test, Options{}) }
func TestTSGoCacheShippedRelease(test *testing.T) {
	test.Parallel()
	tsgoCacheWitness(test, Options{Release: true})
}
func TestTSGoCacheSanitized(test *testing.T) {
	test.Parallel()
	tsgoCacheWitness(test, Options{Sanitize: true})
}

// The fixture exercises the include-order features without calling the checker ABI.
// Stub symbols permit a real checker link while keeping each test's setup bounded.
func tsgoCacheWitness(test *testing.T, options Options) {
	test.Helper()
	fixture, absolutePathError := filepath.Abs("../oracle/testdata/arguments_length_extended.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	loaded, loadError := load.Load([]string{fixture})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := "#include \"tsgo_runtime.h\"\n" + C(program)
	node := toolchain.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", fixture)
	expected, combinedOutputError := node.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("source Node: %v\n%s", combinedOutputError, expected)
	}
	directory := test.TempDir()
	var stubs strings.Builder
	seen := map[string]bool{}
	for _, match := range regexp.MustCompile(`\b(tsgo_[a-z_]+)\(`).FindAllStringSubmatch(string(bridge.Header), -1) {
		if !seen[match[1]] {
			seen[match[1]] = true
			stubs.WriteString("void " + match[1] + "(void) {}\n")
		}
	}
	if len(seen) == 0 {
		test.Fatal("checker ABI declares no symbols")
	}
	if writeError := os.WriteFile(
		filepath.Join(directory, "stubs.c"),
		[]byte(stubs.String()),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	compile := toolchain.CommandIn(directory, compiler, "-c", "stubs.c", "-o", "stubs.o")
	if output, combinedOutputError := compile.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("stubs: %v\n%s", combinedOutputError, output)
	}
	archive := toolchain.CommandIn(directory, archiverName(compiler), "rcs", "checker.a", "stubs.o")
	if output, combinedOutputError := archive.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("archive: %v\n%s", combinedOutputError, output)
	}
	library, tsgoRuntimeLibraryError := tsgoRuntimeLibrary(source, options)
	if tsgoRuntimeLibraryError != nil {
		test.Fatal(tsgoRuntimeLibraryError)
	}
	info, statError := os.Stat(library)
	if statError != nil {
		test.Fatal(statError)
	}
	var first []byte
	for _, round := range []string{"first", "warm"} {
		binary := filepath.Join(directory, round, "program")
		if createDirectoryError := os.Mkdir(filepath.Dir(binary), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if buildTsGoError := BuildTSGo(
			source,
			binary,
			filepath.Join(directory, "checker.a"),
			options,
		); buildTsGoError != nil {
			test.Fatal(buildTsGoError)
		}
		run := toolchain.Command(binary)
		run.Env = append(run.Env, "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		output, combinedOutputError := run.CombinedOutput()
		if combinedOutputError != nil || !bytes.Equal(output, expected) {
			test.Fatalf("%s: %v output %q, Node %q", round, combinedOutputError, output, expected)
		}
		content, readError := os.ReadFile(binary)
		if readError != nil {
			test.Fatal(readError)
		}
		for _, place := range []string{directory, filepath.Dir(library), "adamic-tsgo-", ".build-"} {
			if bytes.Contains(content, []byte(place)) {
				test.Errorf("%s binary names build place %q", round, place)
			}
		}
		if first == nil {
			first = content
		} else if !bytes.Equal(first, content) {
			test.Error("cold and warm binaries differ")
		}
	}
	warm, tsgoRuntimeLibraryError := tsgoRuntimeLibrary(source, options)
	if tsgoRuntimeLibraryError != nil {
		test.Fatal(tsgoRuntimeLibraryError)
	}
	after, statError := os.Stat(warm)
	if statError != nil {
		test.Fatal(statError)
	}
	if library != warm || !os.SameFile(info, after) || !info.ModTime().Equal(after.ModTime()) {
		test.Error("warm build replaced the cached runtime")
	}
	if !test.Failed() {
		test.Logf(
			"first and warm builds agree with source Node: %q; binary bytes and runtime archive unchanged",
			expected,
		)
	}
}
