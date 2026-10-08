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

func TestCounterSquareJumpDoubleProductMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/counters_square_jump.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	counter := -1
	for id, local := range program.Locals {
		if local.Name == "i" {
			counter = id
			if local.CounterGuard == nil || !local.CounterGuard.Square {
				t.Fatal("square loop must have a guarded integer path")
			}
		}
	}
	if counter < 0 {
		t.Fatal("missing square counter")
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "iteration 0: 0\niterations 1\n" {
		t.Fatalf("unexpected Node result: %+v", truth)
	}
	sanitized, _ := natively(t, program)
	for name, result := range map[string]run{"release": released(t, program), "sanitized": sanitized} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("%s: %s; stderr %s", name, difference, result.stderr)
		}
	}
	// The update to 2^52 fits the counter, but its square does not fit int64.
	// Mutate only the integer path's double product, leaving the entry guard
	// and ordinary double fallback intact.
	code := native.C(program)
	local := fmt.Sprintf("adamic_local_%d_i", counter)
	product := fmt.Sprintf("((double)%s) * ((double)%s)", local, local)
	if strings.Count(code, product) != 1 {
		t.Fatal("expected one square product with double counter reads")
	}
	code = strings.Replace(code, product, local+" * "+local, 1)
	binary := filepath.Join(t.TempDir(), "integer-square-mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	diagnostic := string(result.stderr)
	if result.exitCode == 0 || !strings.Contains(diagnostic, "signed integer overflow: 4503599627370496 * 4503599627370496") || !strings.Contains(diagnostic, "UndefinedBehaviorSanitizer") {
		t.Fatalf("integer-square mutant survived UBSan: exit %d, stderr %s", result.exitCode, diagnostic)
	}
	t.Log("UBSan caught the mutant: signed integer overflow: 4503599627370496 * 4503599627370496")
}
