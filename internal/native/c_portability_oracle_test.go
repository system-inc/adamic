package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestLongUnicodePortabilityClang(test *testing.T) {
	test.Parallel()
	longUnicodePortability(test, "clang")
}

func TestLongUnicodePortabilityGCC(test *testing.T) {
	test.Parallel()
	longUnicodePortability(test, "gcc")
}

func TestLongUnicodeLiteralByteCoverage(test *testing.T) {
	test.Parallel()
	program, _ := portabilityFixture(test)
	seen := map[byte]bool{}
	for _, value := range program.Strings {
		if len(value) <= longestLiteral {
			continue
		}
		if !utf8.ValidString(value) {
			test.Fatal("fixture contains invalid UTF-8")
		}
		for index := range len(value) {
			seen[value[index]] = true
		}
	}
	for value := 0x80; value <= 0xf4; value++ {
		if value == 0xc0 || value == 0xc1 {
			continue
		}
		if !seen[byte(value)] {
			test.Fatalf("missing valid high UTF-8 byte %02x", value)
		}
	}
	test.Log("long literals cover all 115 valid high UTF-8 bytes")
}

func portabilityFixture(test *testing.T) (*ir.Program, string) {
	test.Helper()
	file, err := filepath.Abs("../oracle/testdata/long_unicode_literals.a")
	if err != nil {
		test.Fatal(err)
	}
	loaded, err := load.Load([]string{file})
	if err != nil {
		test.Fatal(err)
	}
	program, err := lower.Lower(test.Context(), loaded)
	if err != nil {
		test.Fatal(err)
	}
	return program, file
}

func longUnicodePortability(test *testing.T, compiler string) {
	test.Helper()
	program, file := portabilityFixture(test)
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		test.Fatal(err)
	}
	want := portabilityRun(test, "node", "--disable-warning=ExperimentalWarning", runner, file)
	directory := test.TempDir()
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		test.Fatal(err)
	}
	if got := portabilityRun(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		runner,
		generated,
	); !bytes.Equal(got, want) {
		test.Fatalf("JavaScript output differs from source Node: got %q, want %q", got, want)
	}

	// Use the same compiler for the complete runtime and generated program. O0
	// keeps cold sanitizer setup within this leaf's test-grain budget.
	flags := []string{
		"-std=c11",
		"-Wpedantic",
		"-Werror",
		"-fsigned-char",
		"-O0",
		"-g",
		"-fsanitize=address,undefined",
		"-fno-sanitize-recover=all",
		"-fno-omit-frame-pointer",
		"-fno-pie",
		"-ffp-contract=off",
		"-fno-optimize-sibling-calls",
	}
	source := C(program)
	flags = append(flags, featureFlags(source)...)
	compilerPath, err := exec.LookPath(compiler)
	if err != nil {
		test.Fatal(err)
	}
	version, err := exec.Command(compilerPath, "--version").CombinedOutput()
	if err != nil {
		test.Fatal(err)
	}
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		test.Fatal(err)
	}
	kept := files[:0]
	for _, file := range files {
		if file.name == "node_host.c" && !strings.Contains(source, "#define ADAMIC_NODE_HOST 1\n") {
			continue
		}
		if file.name == "regexp_replace.c" && !strings.Contains(source, "#define ADAMIC_REGEXP_REPLACE_CALLBACK 1\n") {
			continue
		}
		kept = append(kept, file)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		test.Fatal(err)
	}
	library, err := cachedRuntime(
		kept,
		flags,
		compilerPath,
		string(version),
		filepath.Join(cache, "adamic", "c-portability"),
	)
	if err != nil {
		test.Fatal(err)
	}
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(source), 0o644); err != nil {
		test.Fatal(err)
	}
	binary := filepath.Join(directory, "native")
	arguments := append(append([]string{}, flags...), "-no-pie", "-I", filepath.Dir(library), main, "-o", binary)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	output, err := exec.Command(compilerPath, arguments...).CombinedOutput()
	if err != nil || len(output) != 0 {
		test.Fatalf("%s build: %v\n%s", compiler, err, output)
	}
	if got := portabilityRun(test, binary); !bytes.Equal(got, want) {
		test.Fatalf("%s native output differs from source Node: got %q, want %q", compiler, got, want)
	}
	test.Logf("%s: both backends match Node; ASan, UBSan and leak detection clean", compiler)
}

func portabilityRun(test *testing.T, name string, arguments ...string) []byte {
	test.Helper()
	command := exec.CommandContext(test.Context(), name, arguments...)
	command.Env = append(
		os.Environ(),
		"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1",
		"LSAN_OPTIONS=exitcode=23",
		"UBSAN_OPTIONS=halt_on_error=1",
	)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		test.Fatalf("%s: %v\n%s", name, err, stderr.String())
	}
	return stdout.Bytes()
}
