package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	// The generic strip transform preserves const enums; this witness needs actual TypeScript emission.
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		return []string{counted(t, "stage3/namespace-value/erased-block.a", false, nil, false, false)}
	})
	for _, name := range []string{"qualified-live", "tracing-qualified", "stored", "passed", "returned", "identity", "staged", "merged", "structural-view", "order", "effects", "prototype-presence", "imports/main", "wider-reference"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"stage3/namespace-value/" + name + ".a", true, false})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"stage3/fixtures/namespaces/10_tracing_escape.a", true, false})

}

func TestNamespaceValueBoundary(t *testing.T) {
	for _, test := range []struct{ name, expression, stdout string }{
		{"stored", "const held = State;", "2\n7\n"},
		{"passed", "inspect(State)", "2\n"},
		{"returned", "return State;", "2\n7\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/namespace-value", test.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != test.stdout {
				t.Fatalf("source Node: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			assertNamespaceAgreement(t, path, program)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			before, after, found := strings.Cut(string(source), test.expression)
			if !found {
				t.Fatal("missing escape site")
			}
			// Replacing only the escaped value by current fields models the tempting snapshot implementation.
			snapshot := "{ value: State.value, advance: State.advance }"
			replacement := map[string]string{"stored": "const held = " + snapshot + ";", "passed": "inspect(" + snapshot + ")", "returned": "return " + snapshot + ";"}[test.name]
			mutantPath := filepath.Join(t.TempDir(), "snapshot.a")
			if err := os.WriteFile(mutantPath, []byte(before+replacement+after), 0600); err != nil {
				t.Fatal(err)
			}
			mutant, err := lowered(t, mutantPath)
			if err != nil {
				t.Fatal(err)
			}
			nodeMutant := onNode(t, mutantPath)
			js := onJavaScriptBackend(t, mutant)
			observed, binary := nativelyUncached(t, mutant)
			for backend, result := range map[string]run{"JavaScript": js, "native": observed} {
				if d := disagreement(nodeMutant, result); d != "" {
					t.Fatalf("mutant must match its own Node source, %s %s", backend, d)
				}
				if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(truth, result) != "stdout differs" {
					t.Fatalf("snapshot must finish normally and lose the live binding, %s %+v", backend, result)
				}
			}
			if report := leaksUncached(t, mutant, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("Node %q; snapshot mutant %q caught in both backends", truth.stdout, observed.stdout)
		})
	}
}

func TestTracingNamespaceValueBoundary(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/namespaces/10_tracing_escape.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "1,3\nparse\n" {
		t.Fatalf("unchanged stock reduction Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	assertNamespaceAgreement(t, path, program)
	t.Logf("unchanged tracing escape matches TypeScript emit on Node: %q", truth.stdout)
}

func assertNamespaceAgreement(t *testing.T, path string, program *ir.Program) {
	t.Helper()
	truth := onTypeScriptNamespace(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("TypeScript emit on Node: %+v", truth)
	}
	native, binary := nativelyUncached(t, program)
	for backend, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if d := disagreement(truth, result); d != "" {
			t.Fatalf("%s %s: got %+v, Node %+v", backend, d, result, truth)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func onTypeScriptNamespace(t *testing.T, path string) run {
	t.Helper()
	return execute(t, "node", filepath.Join(repository, "stage3/namespace-value/typescript-node.mjs"), path)
}

func TestNamespaceObjectAgreement(t *testing.T) {
	for _, name := range []string{"identity", "staged", "merged", "structural-view", "order", "effects", "prototype-presence", "imports/main", "wider-reference", "erased-block"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(repository, "stage3/namespace-value", name+".a")
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			assertNamespaceAgreement(t, path, program)
		})
	}
}

func TestNamespaceReadonlyWiderViewStop(t *testing.T) {
	path := filepath.Join(repository, "stage3/namespace-value/wider-const.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onTypeScriptNamespace(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "2\n" || len(truth.stderr) != 0 {
		t.Fatalf("TypeScript Node: %+v", truth)
	}
	want := run{stderr: []byte("adamic: panic: namespace write failed: value is readonly\n"), exitCode: 70}
	native, _ := nativelyUncached(t, program)
	for backend, result := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if d := disagreement(want, result); d != "" {
			t.Fatalf("%s %s: %+v", backend, d, result)
		}
	}
}

func TestNamespaceDescriptorBoundary(t *testing.T) {
	path := filepath.Join(repository, "stage3/namespace-value/descriptor.a")
	_, err := lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "descriptor reflection on an escaped namespace object") {
		t.Fatalf("missing descriptor boundary: %v", err)
	}
	truth := onTypeScriptNamespace(t, path)
	if truth.exitCode != 0 || len(truth.stdout) != 0 || len(truth.stderr) != 0 {
		t.Fatalf("TypeScript descriptor source: %+v", truth)
	}
}

func TestNamespaceObjectMutants(t *testing.T) {
	for _, name := range []string{"second-object", "early-key", "wrong-order", "drop-readonly"} {
		t.Run(name, func(t *testing.T) {
			fixture := map[string]string{"second-object": "identity", "early-key": "staged", "wrong-order": "order", "drop-readonly": "wider-const"}[name]
			path := filepath.Join(repository, "stage3/namespace-value", fixture+".a")
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onTypeScriptNamespace(t, path)
			changed := false
			switch name {
			case "second-object":
				for f := range program.Functions {
					if program.Functions[f].Name != "retain" {
						continue
					}
					for i, statement := range program.Functions[f].Body {
						if ret, ok := statement.(ir.Return); ok {
							ret.Value = ir.ObjectLiteral{Record: true, Fields: []ir.Field{{Name: "value", Value: ir.Box{Value: ir.Property{Object: ret.Value, Name: "value", Of: ir.Number, Namespace: true, View: "State.value"}}}}}
							program.Functions[f].Body[i] = ret
							changed = true
						}
					}
				}
			case "early-key":
				var install ir.SetProperty
				position := -1
				for i, statement := range program.Main {
					if set, ok := statement.(ir.SetProperty); ok && set.NamespaceInstall && set.Name == "later" {
						install = set
						position = i
						break
					}
				}
				if position >= 0 {
					program.Main = append(program.Main[:position], program.Main[position+1:]...)
					for i, statement := range program.Main {
						if init, ok := statement.(ir.If); ok {
							if _, ok := init.Condition.(ir.IsUndefined); ok {
								init.Then = append(init.Then, install)
								program.Main[i] = init
								changed = true
								break
							}
						}
					}
				}
			case "wrong-order":
				positions := []int{}
				for i, statement := range program.Main {
					if set, ok := statement.(ir.SetProperty); ok && set.NamespaceInstall {
						positions = append(positions, i)
					}
				}
				if len(positions) == 2 {
					a, b := positions[0], positions[1]
					program.Main[a], program.Main[b] = program.Main[b], program.Main[a]
					changed = true
				}
			case "drop-readonly":
				for i, statement := range program.Main {
					if set, ok := statement.(ir.SetProperty); ok && set.NamespaceReadonly {
						set.NamespaceReadonly = false
						program.Main[i] = set
						changed = true
					}
				}
			}
			if !changed {
				t.Fatal("mutant did not change the named mechanism")
			}
			native, binary := nativelyUncached(t, program)
			js := onJavaScriptBackend(t, program)
			for backend, result := range map[string]run{"native": native, "JavaScript": js} {
				if result.exitCode != 0 || len(result.stderr) != 0 {
					t.Fatalf("mutant must finish cleanly: %s %+v", backend, result)
				}
				if name == "drop-readonly" {
					if d := disagreement(truth, result); d != "" {
						t.Fatalf("unchecked mutant must match Node: %s %s", backend, d)
					}
				} else if d := disagreement(truth, result); d != "stdout differs" {
					t.Fatalf("semantic mutant not caught by stdout: %s %s", backend, d)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("%s caught: original Node %q; mutant %q", name, truth.stdout, native.stdout)
		})
	}
}

func TestNamespaceDirectWritesStayRefused(t *testing.T) {
	path := filepath.Join(repository, "stage3/namespace-value/readonly-const.a")
	_, err := load.Load([]string{path})
	var check *load.CheckError
	if !errors.As(err, &check) || !strings.Contains(err.Error(), "TS2540") {
		t.Fatalf("const: want TS2540, got %v", err)
	}
	path = filepath.Join(repository, "stage3/namespace-value/readonly-function.a")
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "namespace function member") {
		t.Fatalf("function: want ruled refusal, got %v", err)
	}
	truth := onTypeScriptNamespace(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "2\n" || len(truth.stderr) != 0 {
		t.Fatalf("observed TypeScript function replacement: %+v", truth)
	}
}

func TestNamespaceContainerLimitsStayNamed(t *testing.T) {
	for _, test := range []struct {
		name, reason string
		refused      bool
	}{{"spread", "object spread with escaped namespace storage", false}, {"nullable", "nullable namespace export", false}, {"cycle", "cycle reference counting can't free", true}} {
		path := filepath.Join(repository, "stage3/namespace-value", test.name+".a")
		_, err := lowered(t, path)
		if test.refused {
			var refusal *lower.Refused
			if !errors.As(err, &refusal) || !strings.Contains(refusal.What, test.reason) {
				t.Fatalf("%s: %v", test.name, err)
			}
		} else {
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, test.reason) {
				t.Fatalf("%s: %v", test.name, err)
			}
		}
	}
}
