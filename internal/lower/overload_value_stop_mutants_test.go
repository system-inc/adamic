package lower

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

// overloadValueStopRows are TestOverloadNarrowAsAValueStops' rows whose set has an overload served only per
// call, by narrowOverload's result check, read as a value into a narrower slot (#ytagqs5, Adamic's pin).
var overloadValueStopRows = []string{"module_const_into_a_narrower_slot", "nested_into_a_narrower_slot"}

// overloadValueStopMutantBinary builds the lower test binary with overloadSetValueStop's "an overloaded
// function as a value" return deleted, as lever 1 read loosely would delete it, a product.
func overloadValueStopMutantBinary(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "overload-value-stop.test", "./internal/lower", nil, []buildcache.Mutation{{File: "internal/lower/expression.go", Before: "\t\t\treturn l.notYet(node, \"an overloaded function as a value\")\n", After: ""}})
}
func TestProduct_OverloadValueStopMutant(t *testing.T) {
	t.Parallel()
	overloadValueStopMutantBinary(t)
}

// Under the mutant each row goes red with its program lowering outright ("got <nil>"), so the red is the
// value stop's alone and not a stop some other road says.
func TestOverloadValueStopMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := overloadValueStopMutantBinary(t)
	for _, row := range overloadValueStopRows {
		t.Run(row, func(t *testing.T) {
			t.Parallel()
			pin := "TestOverloadNarrowAsAValueStops"
			command := exec.CommandContext(t.Context(), binary, "-test.run", "^"+pin+"$/^"+row+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(t.Context(), command)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("the mutant missed %s/%s: %v\n%s\n%s", pin, row, err, stdout, stderr)
			}
			if !strings.Contains(string(stdout), "--- FAIL: "+pin+"/"+row+" ") {
				t.Fatalf("the mutant left %s/%s green:\n%s", pin, row, stdout)
			}
			if !strings.Contains(string(stdout), "got <nil>") {
				t.Fatalf("the mutant turned %s/%s red without lowering its program, so another road holds it:\n%s", pin, row, stdout)
			}
			t.Logf("the mutant turns %s/%s red, its program lowering", pin, row)
		})
	}
}
