package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"p21", "p60", "p61", "p23", "p22_adapted", "p59_adapted", "p62_adapted", "p135_tagged_unread"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/lower/testdata/callable_producers/" + name + ".a", true, false})
	}
	for _, name := range []string{"p22", "p59", "p62", "p135_object_result", "p135_unique_object_result"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/lower/testdata/callable_producers/" + name + ".a", false, false})
	}

}

// Stale direct-function certificates must never emit a closure-code comparison.
// Inspect before clang so incompatible-pointer warnings are not a mutant kill.
func TestCallableProducerRegistryFiltersDirectFunctions(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/lower/testdata/callable_producers/p21.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	injected := 0
	for i := range program.ViewContracts {
		if !program.ViewContracts[i].ProducerCertified {
			continue
		}
		for index, function := range program.Functions {
			if !function.Closure || function.Receiver {
				program.ViewContracts[i].Functions = append(program.ViewContracts[i].Functions, index)
				injected++
			}
		}
	}
	if injected == 0 {
		t.Fatal("no direct function control")
	}
	code := native.C(program)
	for index, function := range program.Functions {
		if function.Closure && !function.Receiver {
			continue
		}
		for _, field := range []string{"code", "counted_code"} {
			if strings.Contains(code, fmt.Sprintf("->%s == adamic_function_%d_", field, index)) {
				t.Fatalf("direct function %d compared as closure code", index)
			}
		}
	}
}
