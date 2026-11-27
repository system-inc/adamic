package lower

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

// arrayStoredConstraintMutantBinary builds the lower test binary with arrayStoredPropertyNames' old clause restored:
// every intersection holding a type parameter is counted as stored as an array at once, its parameters' constraints
// never read, as before rulings.md:260's amendment (Nov 27, @system_adamic on #67fpbxv), a product.
func arrayStoredConstraintMutantBinary(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "array-stored-constraint.test", "./internal/lower", nil, []buildcache.Mutation{{File: "internal/lower/library_array_predicate.go", Before: "\t\t\t\tbounded = append(bounded, proven)\n", After: "\t\t\t\tadd(proven)\n"}})
}
func TestProduct_ArrayStoredConstraintMutant(t *testing.T) {
	t.Parallel()
	arrayStoredConstraintMutantBinary(t)
}

// Under the mutant each of TestArrayPredicateConstraintAdmitsNode's rows goes red with its program refused at the stop
// it met before the amendment, by its position and its words, so the red is Cloned's names keeping Node out of
// Array.isArray's domain again, and not a stop some other road says.
func TestArrayStoredConstraintMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := arrayStoredConstraintMutantBinary(t)
	for _, row := range arrayPredicateConstraintRows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			pin := "TestArrayPredicateConstraintAdmitsNode"
			command := exec.CommandContext(t.Context(), binary, "-test.run", "^"+pin+"$/^"+row.name+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(t.Context(), command)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("the mutant missed %s/%s: %v\n%s\n%s", pin, row.name, err, stdout, stderr)
			}
			if !strings.Contains(string(stdout), "--- FAIL: "+pin+"/"+row.name+" ") {
				t.Fatalf("the mutant left %s/%s green:\n%s", pin, row.name, stdout)
			}
			if stop := "main.ts" + row.where + ": stage 0 can't lower " + row.what + " yet"; !strings.Contains(string(stdout), stop) {
				t.Fatalf("the mutant turned %s/%s red without its old stop %q, so another road holds it:\n%s", pin, row.name, stop, stdout)
			}
			t.Logf("the mutant turns %s/%s red at its old stop, %s", pin, row.name, row.where)
		})
	}
}
