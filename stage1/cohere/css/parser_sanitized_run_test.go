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
)

// cssParserSanitizedAgreement is the agreement group's one sanitized native run: on Linux leak detection is on in it,
// so it checks the output, ASan/UBSan and LeakSanitizer, where a second run checked leaks before (#testaudit).
func cssParserSanitizedAgreement(t *testing.T, binary, cases string) run {
	t.Helper()
	return cssParserExecute(t, cssParserASAN(true), binary, cases)
}

// cssParserReleaseMutant is the unmutated parser's emitted C with every release dropped, built sanitized, as a
// product: its output is the parser's, and only LeakSanitizer can tell it apart.
func cssParserReleaseMutant(t *testing.T) string {
	t.Helper()
	program := cssParserPreparedProgram(t, -1)
	options := native.Options{Sanitize: true}
	dir := cssParserProduct(t, cssParserInputs{Name: "parser release mutant sanitized", Files: []string{program.c, repository + "/internal/native", repository + "/go.mod"}, Flags: append(native.Flags(options), "mutation=adamic_release dropped"), Toolchain: buildcache.Tool("clang", "--version")}, func(dir string) error {
		source, err := os.ReadFile(program.c)
		if err != nil {
			return err
		}
		code := strings.ReplaceAll(string(source), "adamic_release(", "(void)(")
		if code == string(source) {
			return errors.New("release mutant changed nothing")
		}
		return native.Build(code, filepath.Join(dir, "parser"), options)
	})
	return filepath.Join(dir, "parser")
}

func TestProduct_CSSParserReleaseMutant(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux runs leak detection inside the sanitized run")
	}
	cssParserReleaseMutant(t)
}

// The merged run still fails on what the removed leak run caught. A parser that drops its releases prints the parser's
// answers and passes ASan/UBSan with leak detection off, so only leak detection catches it, and the agreement's one
// sanitized run fails it as a leak with the output still right. The unmutated parser passes the same run.
func TestCSSParserSanitizedRunCatchesALeak(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux runs leak detection inside the sanitized run")
	}
	corpus := cssParserTopCorpus(t)
	cases, want := writeCSSParserSlice(t, corpus, []int{0, 1, 2, 80})
	answers := func(name string, result run) {
		t.Helper()
		if difference := firstDifference(string(result.stdout), want); difference != "" {
			t.Fatalf("%s: output differs from Go: %s", name, difference)
		}
	}
	unmutated := cssParserSanitizedAgreement(t, cssParserPreparedNative(t, -1), cases)
	if unmutated.exitCode != 0 || len(unmutated.stderr) != 0 {
		t.Fatalf("the unmutated parser: exit %d, %s", unmutated.exitCode, unmutated.stderr)
	}
	answers("the unmutated parser", unmutated)
	mutant := cssParserReleaseMutant(t)
	unchecked := cssParserExecute(t, cssParserASAN(false), mutant, cases)
	if unchecked.exitCode != 0 || len(unchecked.stderr) != 0 {
		t.Fatalf("with leak detection off the release mutant should pass: exit %d, %s", unchecked.exitCode, unchecked.stderr)
	}
	answers("the release mutant with leak detection off", unchecked)
	caught := cssParserSanitizedAgreement(t, mutant, cases)
	if caught.exitCode == 0 || !strings.Contains(string(caught.stderr), "LeakSanitizer: detected memory leaks") {
		t.Fatalf("the merged run let a parser that drops its releases pass: exit %d, %.300s", caught.exitCode, caught.stderr)
	}
	answers("the release mutant", caught)
	for _, line := range strings.Split(string(caught.stderr), "\n") {
		if strings.Contains(line, "SUMMARY:") {
			t.Logf("planted leak caught by the merged run: %s", line)
		}
	}
}
