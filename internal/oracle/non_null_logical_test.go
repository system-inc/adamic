package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"repro", "local", "field", "element", "short_circuit", "missing_or", "missing_and", "missing_nullish"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/non_null_logical_" + name + ".a", true,
			name == "missing_or" || name == "missing_and" || name == "missing_nullish",
		})
	}
}

func TestNonNullLogicalMissingReadStopsBeforeRight(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"missing_or", "missing_and", "missing_nullish"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_logical_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: non-null assertion failed: value! is null or undefined\n"), exitCode: 70}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			}
		})
	}
}

func TestNonNullLogicalAlwaysEvaluateRightMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_logical_short_circuit.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("0 0\n")}, want); difference != "" {
		t.Fatal("Node: " + difference)
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	changes := 0
	program.Main = mutateReadiness(program.Main, func(node any) any {
		block, ok := node.(ir.Block)
		if !ok {
			return node
		}
		body := []ir.Statement{}
		for _, statement := range block.Body {
			branch, ok := statement.(ir.If)
			if ok && len(branch.Then) == 1 {
				store, ok := branch.Then[0].(ir.Assign)
				if ok {
					if _, call := store.Value.(ir.Call); call {
						// Keep the conditional store, but incorrectly evaluate its RHS first.
						local := len(program.Locals)
						program.Locals = append(program.Locals, ir.Local{Name: "mutant_rhs", Type: store.Value.Type(), Function: -1})
						body = append(body, ir.Declare{Local: local, Value: store.Value})
						store.Value = ir.Read{Local: local, Of: store.Value.Type()}
						branch.Then = []ir.Statement{store}
						statement = branch
						changes++
					}
				}
			}
			body = append(body, statement)
		}
		block.Body = body
		return block
	})
	if changes != 3 {
		t.Fatalf("always-evaluate mutant changed %d RHS calls, want three", changes)
	}
	mutant, _ := nativelyUncached(t, program)
	if mutant.exitCode != 0 || string(mutant.stdout) != "3 0\n" || disagreement(want, mutant) != "stdout differs" {
		t.Fatalf("always-evaluate mutant escaped Node output: %#v", mutant)
	}
	t.Logf("always-evaluate RHS caught by Node: stdout %q versus %q, exit 0", mutant.stdout, want.stdout)
}
