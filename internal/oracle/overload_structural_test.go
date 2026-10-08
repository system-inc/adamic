package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"block", "evaluator", "fields"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/overload_structural_" + name + ".a", true, false})
	}
}

// The shared lowering boundary refuses before either backend receives IR.
// The general negative-fixture runner accepts only NotYet, so pin Refused here.
func TestOverloadStructuralFieldRefusal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/overload-structural/fields-mixed.a"))
	if err != nil {
		t.Fatal(err)
	}
	witness := onNode(t, path)
	if witness.exitCode != 0 || string(witness.stdout) != "7\n" || len(witness.stderr) != 0 {
		t.Fatalf("Node witness: %+v", witness)
	}
	program, err := lowered(t, path)
	var refusal *lower.Refused
	if program != nil || !errors.As(err, &refusal) || !strings.Contains(refusal.What, "result.value without a single storage representation") {
		t.Fatalf("wanted result.value storage refusal before emission, got %v", err)
	}
}
