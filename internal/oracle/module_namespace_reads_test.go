package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"main", "early", "initialized", "direct_early", "direct_initialized", "hoisted"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/module_namespace_reads/" + name + ".a", true, false})
	}
}

func TestModuleNamespaceReadinessMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"early", "direct_early"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/module_namespace_reads", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			mutateStringExpressions(reflect.ValueOf(program).Elem(), func(value ir.Expression) ir.Expression {
				if read, ok := value.(ir.Read); ok && read.Checked && program.Locals[read.Local].Name == "value" {
					read.Checked = false
					changed = true
					return read
				}
				return value
			})
			if !changed || truth.exitCode != 70 {
				t.Fatalf("no firing readiness check: changed=%v Node=%+v", changed, truth)
			}
			compiled, _ := natively(t, program)
			for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
				if got.exitCode != 0 || disagreement(truth, got) != "exit codes differ" {
					t.Fatalf("%s removed-readiness mutant survived: %+v", backend, got)
				}
				t.Logf("%s treat reaching read as initialized: caught by Node exit 70; mutant exit 0", backend)
			}
		})
	}
}

func TestModuleNamespaceLiveBindingMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/module_namespace_reads/main.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	// Snapshotting the original zero instead of reading the real mutable export
	// loses mark's writes without causing a build or sanitizer failure.
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
		if read, ok := value.(ir.Read); ok && program.Locals[read.Local].Name == "count" {
			changed = true
			return ir.NumberConstant{Value: 0}
		}
		return value
	})
	if !changed {
		t.Fatal("mutant replaced no live export reads")
	}
	compiled, _ := natively(t, program)
	for backend, got := range map[string]run{"native": compiled, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("%s frozen-export mutant survived: %+v", backend, got)
		}
		t.Logf("%s frozen export caught by Node stdout; mutant exits 0 with clean sanitizers", backend)
	}
}
