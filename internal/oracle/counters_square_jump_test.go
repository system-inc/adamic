package oracle

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/counters_square_jump.a", true, false})
}

func TestCounterSquareJumpDoubleProductMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/counters_square_jump.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	counter := -1
	for id, local := range program.Locals {
		if local.Name == "i" {
			counter = id
			if local.CounterGuard == nil || !local.CounterGuard.Square {
				test.Fatal("square loop must have a guarded integer path")
			}
		}
	}
	if counter < 0 {
		test.Fatal("missing square counter")
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "iteration 0: 0\niterations 1\n" {
		test.Fatalf("unexpected Node result: %+v", truth)
	}
	sanitized, _ := natively(test, program)
	for name, result := range map[string]run{"release": released(test, program), "sanitized": sanitized} {
		if difference := disagreement(truth, result); difference != "" {
			test.Fatalf("%s: %s; stderr %s", name, difference, result.stderr)
		}
	}
	// The update to 2^52 fits the counter, but its square does not fit int64.
	// Mutate only the integer path's double product, leaving the entry guard
	// and ordinary double fallback intact.
	code := native.C(program)
	local := emittedLocalSymbol(test, code, "adamic_local_", "i")
	product := fmt.Sprintf("((double)%s) * ((double)%s)", local, local)
	if strings.Count(code, product) != 1 {
		test.Fatal("expected one square product with double counter reads")
	}
	code = strings.Replace(code, product, local+" * "+local, 1)
	binary := filepath.Join(test.TempDir(), "integer-square-mutant")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	result := execute(test, binary)
	diagnostic := string(result.stderr)
	if result.exitCode == 0 ||
		!strings.Contains(diagnostic, "signed integer overflow: 4503599627370496 * 4503599627370496") ||
		!strings.Contains(diagnostic, "UndefinedBehaviorSanitizer") {
		test.Fatalf("integer-square mutant survived UBSan: exit %d, stderr %s", result.exitCode, diagnostic)
	}
	test.Log("UBSan caught the mutant: signed integer overflow: 4503599627370496 * 4503599627370496")
}
