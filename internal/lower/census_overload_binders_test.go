package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"strings"
	"testing"
)

func TestCensusOverloadBinderGuards(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stop string }{
		{"constraint", `function constrain(value: string): string;
function constrain<T extends 'yes'>(value: T): string { return value; }
console.log(constrain('no'));`, "cannot satisfy implementation parameter"},
		{"parameter", `function pair<T>(left: T, right: string): void;
function pair<T, Extra>(left: T, right: T): void {}
pair(42, 'no');`, "cannot be served by implementation parameter"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.stop) {
				t.Fatalf("want %q, got %v", probe.stop, err)
			}
		})
	}
}

// The resolved numeric overload now has a checked entry, rather than a refusal.
// The oracle pins exit 70 before the incompatible string can reach the caller.
func TestCensusOverloadBinderResultHasCheckedEntry(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/census_overload_binder_result_checked.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if !strings.Contains(function.Name, "wrong_overload_1_specialized_") {
			continue
		}
		for _, statement := range function.Body {
			branch, ok := statement.(ir.If)
			if !ok {
				continue
			}
			for _, guarded := range branch.Then {
				if _, ok := guarded.(ir.Panic); ok {
					return
				}
			}
		}
	}
	t.Fatal("resolved result has no guarded refusal in its checked entry")
}
