package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

var callPresencePaths = []string{"direct", "closure", "method", "apply", "spread"}

func init() {
	for _, name := range callPresencePaths {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/call_presence_" + name + ".a", true, false})
	}
}

// ECMA-262 counts supplied arguments, including explicit undefined, while defaults
// run only for undefined inputs. All five boundaries must preserve both facts.
func TestCallPresenceECMA262(t *testing.T) {
	t.Parallel()
	for _, name := range callPresencePaths {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/call_presence_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "0:9\n1:9\n1:0\n" {
				t.Fatalf("source Node: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			actual, binary := natively(t, program)
			if difference := disagreement(truth, actual); difference != "" {
				t.Fatalf("native: %s: %+v", difference, actual)
			}
			if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
				t.Fatalf("JavaScript: %s", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(truth, onWASI(t, native.C(program))); difference != "" {
					t.Fatalf("WASI: %s", difference)
				}
			}
			t.Logf("%s: Node, native and JavaScript agree: %q; WASI checked when enabled", name, truth.stdout)
			changed := 0
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.ArgumentsCount == 0 {
					continue
				}
				count := function.ArgumentsCount - 1
				arity := len(function.Parameters)
				if function.Receiver {
					arity--
				}
				if function.RestElement != 0 {
					arity--
				}
				mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
					if read, ok := value.(ir.Read); ok && read.Local == count {
						changed++
						return ir.NumberConstant{Value: float64(arity)}
					}
					return value
				})
			}
			if changed == 0 {
				t.Fatal("arity mutant changed no count read")
			}
			mutated, mutantBinary := natively(t, program)
			if mutated.exitCode != 0 || len(mutated.stderr) != 0 {
				t.Fatalf("arity mutant must execute cleanly: %+v", mutated)
			}
			if report := leaks(t, program, mutantBinary); report != "" {
				t.Fatal(report)
			}
			if difference := disagreement(truth, mutated); difference != "stdout differs" {
				t.Fatalf("arity mutant: %q", difference)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(truth, onWASI(t, native.C(program))); difference != "stdout differs" {
					t.Fatalf("WASI arity mutant: %q", difference)
				}
			}
			t.Logf("%s: arity mutant caught only by Node stdout, clean exit 0: %q", name, mutated.stdout)
		})
	}
}
