package css

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// cssParserSanitizedAgreement is the agreement group's one sanitized native run: on Linux leak detection is on in it,
// so it checks the output, ASan/UBSan and LeakSanitizer, where a second run checked leaks before (#testaudit).
func cssParserSanitizedAgreement(test *testing.T, binary, cases string) run {
	test.Helper()
	return cssParserExecute(test, cssParserASAN(true), binary, cases)
}

// cssParserReleaseMutant is the unmutated parser's emitted C with every release dropped, built sanitized, as a
// product: its output is the parser's, and only LeakSanitizer can tell it apart.
func cssParserReleaseMutant(test *testing.T) string {
	test.Helper()
	program := cssParserPreparedProgram(test, -1)
	options := native.Options{Sanitize: true}
	dir := oracles.ParserProduct(test, oracles.ParserProductInputs{
		Name:      "parser release mutant sanitized",
		Files:     []string{program.c, repository + "/internal/native", repository + "/go.mod"},
		Flags:     append(native.Flags(options), "mutation=adamic_release dropped"),
		Toolchain: buildcache.Tool("clang", "--version"),
	}, func(dir string) error {
		source, readError := os.ReadFile(program.c)
		if readError != nil {
			return readError
		}
		code := strings.ReplaceAll(string(source), "adamic_release(", "(void)(")
		if code == string(source) {
			return errors.New("release mutant changed nothing")
		}
		return native.Build(code, filepath.Join(dir, "parser"), options)
	})
	return filepath.Join(dir, "parser")
}

func TestProduct_CSSParserReleaseMutant(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux runs leak detection inside the sanitized run")
	}
	cssParserReleaseMutant(test)
}

// The merged run still fails on what the removed leak run caught. A parser that drops its releases prints the parser's
// answers and passes ASan/UBSan with leak detection off, so only leak detection catches it, and the agreement's one
// sanitized run fails it as a leak with the output still right. The unmutated parser passes the same run.
func TestCSSParserSanitizedRunCatchesALeak(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux runs leak detection inside the sanitized run")
	}
	corpus := oracles.LoadParserCorpus(test)
	cases, want := oracles.WriteParserSlice(test, corpus, []int{0, 1, 2, 80})
	answers := func(name string, result run) {
		test.Helper()
		if difference := firstDifference(string(result.stdout), want); difference != "" {
			test.Fatalf("%s: output differs from Go: %s", name, difference)
		}
	}
	unmutated := cssParserSanitizedAgreement(test, cssParserPreparedNative(test, -1), cases)
	if unmutated.exitCode != 0 || len(unmutated.stderr) != 0 {
		test.Fatalf("the unmutated parser: exit %d, %s", unmutated.exitCode, unmutated.stderr)
	}
	answers("the unmutated parser", unmutated)
	mutant := cssParserReleaseMutant(test)
	unchecked := cssParserExecute(test, cssParserASAN(false), mutant, cases)
	if unchecked.exitCode != 0 || len(unchecked.stderr) != 0 {
		test.Fatalf(
			"with leak detection off the release mutant should pass: exit %d, %s",
			unchecked.exitCode,
			unchecked.stderr,
		)
	}
	answers("the release mutant with leak detection off", unchecked)
	caught := cssParserSanitizedAgreement(test, mutant, cases)
	if caught.exitCode == 0 || !strings.Contains(string(caught.stderr), "LeakSanitizer: detected memory leaks") {
		test.Fatalf(
			"the merged run let a parser that drops its releases pass: exit %d, %.300s",
			caught.exitCode,
			caught.stderr,
		)
	}
	answers("the release mutant", caught)
	for _, line := range strings.Split(string(caught.stderr), "\n") {
		if strings.Contains(line, "SUMMARY:") {
			test.Logf("planted leak caught by the merged run: %s", line)
		}
	}
}
