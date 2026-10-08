package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewRanked23OriginalArrays(t *testing.T) {
	originalRankedArrayOracle(t, "23", 1, 4)
}

// Not parallel: refreshing writes this group's measured rows.
func TestCheckedViewRanked23ArrayCounts(t *testing.T) {
	originalRankedArrayCountTest(t, "23")
}

func TestCheckedViewRanked23ArrayMutants(t *testing.T) {
	declarations, directory, probes := originalArrayInputs(t, "23")
	if declarations == "" {
		t.Skip("original declaration inputs required")
	}
	for _, probe := range probes {
		if probe.Diagnostic == "" {
			continue
		}
		t.Run(probe.Name, func(t *testing.T) {
			file := originalArrayFile(t, declarations, directory, probe.Name)
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.Diagnostic + "\n")}
			node := run{stdout: []byte(probe.Source)}
			if diff := disagreement(node, onNode(t, file)); diff != "" {
				t.Fatal(diff)
			}
			changed := 0
			switch probe.Name {
			case "callbacks-wrong-invoke":
				rewrite := func(n any) any {
					if read, ok := n.(ir.ArrayIndex); ok && read.View == "viewed.callbacks[0]" {
						read.ViewContract = 0
						changed++
						return read
					}
					return n
				}
				for i := range program.Functions {
					program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
				}
			case "callbacks-wrong-array":
				id := ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
				changed = changeTupleOriginalRead(program, func(p ir.Property) bool { return p.View == "viewed.callbacks" }, func(p ir.Property) ir.Property {
					p.Of = ir.Number
					p.ViewContract = id
					p.ViewType = "number"
					return p
				})
			case "callbacks-wrong-element":
				id := ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
				rewrite := func(n any) any {
					if read, ok := n.(ir.ArrayIndex); ok && read.View == "viewed.callbacks[0]" {
						read.Element = ir.Number
						read.ViewType = "number"
						read.ViewContract = id
						changed++
						return read
					}
					return n
				}
				for i := range program.Functions {
					program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
				}
			case "callbacks-wrong-parameter":
				for _, c := range program.ViewContracts {
					if c.Name == "FileWatcherCallback" && c.Result != 0 {
						param := &program.ViewContracts[c.Parameters[0]-1]
						param.Of = ir.Number
						param.RepresentationMask = 1 << ir.Number
						changed++
					}
				}
			case "callbacks-replace-source-arity":
				var extra ir.ViewContractID
				for _, c := range program.ViewContracts {
					if c.Name == "FileWatcherCallback" && len(c.Parameters) == 3 {
						extra = c.Parameters[2]
					}
				}
				rewrite := func(n any) any {
					if array, ok := n.(ir.ArrayLiteral); ok && array.ElementContract > 0 {
						c := &program.ViewContracts[array.ElementContract-1]
						if c.Kind == ir.ViewCallable && len(c.Parameters) == 2 {
							c.Parameters = append(c.Parameters, extra)
							changed++
						}
					}
					return n
				}
				for i := range program.Functions {
					program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
				}
				program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
			}
			if changed == 0 {
				t.Fatal("mutant changed no demanded descriptor")
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if disagreement(want, got) == "" {
					t.Fatal("mutant escaped exact rejection")
				}
				if probe.Name == "callbacks-wrong-invoke" {
					if got.exitCode != 0 || len(got.stderr) != 0 {
						t.Errorf("signature omission must execute: stdout %q stderr %q exit %d", got.stdout, got.stderr, got.exitCode)
					}
					continue
				}
				if diff := disagreement(node, got); diff != "" {
					t.Errorf("mutant must finish with Node output: %s; stdout %q stderr %q", diff, got.stdout, got.stderr)
				}
			}
			if probe.Name == "callbacks-wrong-invoke" && disagreement(node, actual) == "" {
				t.Fatal("native signature omission unexpectedly preserved Node output")
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("mutant caught by pinned stopping oracle in all three modes")
		})
	}
}

func TestCheckedViewCallableArrayResultStorage(t *testing.T) {
	file := filepath.Join(t.TempDir(), "result-storage.a")
	source := "interface Base { readonly kind: string; }\ninterface Target extends Base { readonly values: (() => string | number)[]; }\nfunction read(node: Target): string { return typeof node.values[0]; }\nconst raw = {kind: 'probe', values: [(): number => 1]};\nconst base: Base = raw;\nconsole.log(read(base as Target));\n"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	node := run{stdout: []byte("function\n")}
	if diff := disagreement(node, onNode(t, file)); diff != "" {
		t.Fatal(diff)
	}
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.values[0] expected () => string | number, found function with incompatible result representation\n")}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("result storage: %s; stdout %q stderr %q exit %d", diff, got.stdout, got.stderr, got.exitCode)
		}
	}
	changed := 0
	for _, c := range program.ViewContracts {
		if c.Name == "() => string | number" && c.Result > 0 {
			program.ViewContracts[c.Result-1].Of = ir.Number
			changed++
		}
	}
	if changed == 0 {
		t.Fatal("result mutant changed no signature")
	}
	mutated, binary := nativelyUncached(t, program)
	for _, got := range []run{mutated, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if disagreement(want, got) == "" {
			t.Fatal("physical result mutant escaped stopping oracle")
		}
		if diff := disagreement(node, got); diff != "" {
			t.Errorf("physical result mutant: %s; stderr %q", diff, got.stderr)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestCheckedViewCallableArrayVoidAdapterMutants(t *testing.T) {
	declarations, directory, _ := originalArrayInputs(t, "23")
	if declarations == "" {
		t.Skip("original declaration inputs required")
	}
	for _, name := range []string{"callbacks-short-arity", "callbacks-valued-void-producer"} {
		t.Run(name, func(t *testing.T) {
			file := originalArrayFile(t, declarations, directory, name)
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			node := run{stdout: []byte("a\n")}
			if diff := disagreement(node, onNode(t, file)); diff != "" {
				t.Fatal(diff)
			}
			changed := 0
			for _, c := range program.ViewContracts {
				if c.Name == "FileWatcherCallback" && c.Result > 0 {
					if name == "callbacks-short-arity" {
						program.ViewContracts[c.Result-1].Of = ir.Type(255)
					} else {
						program.ViewContracts[c.Result-1].Of = ir.Boolean
					}
					changed++
				}
			}
			if changed == 0 {
				t.Fatal("void adapter mutant changed no signature")
			}
			actual, _ := nativelyUncached(t, program)
			reason := "function with arity 2"
			if name == "callbacks-valued-void-producer" {
				reason = "function with incompatible result representation"
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: viewed.callbacks[0] expected FileWatcherCallback, found " + reason + "\n")}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if disagreement(node, got) == "" {
					t.Fatal("void adapter mutant escaped Node oracle")
				}
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("void adapter mutant: %s; stderr %q", diff, got.stderr)
				}
			}
			t.Log("void adapter mutant caught by Node control in all three modes")
		})
	}
}

func TestCheckedViewCallableArrayVoidResultObservation(t *testing.T) {
	declarations, directory, _ := originalArrayInputs(t, "23")
	if declarations == "" {
		t.Skip("original declaration inputs required")
	}
	file := originalArrayFile(t, declarations, directory, "callbacks-valued-void-producer")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(data), "viewed.callbacks[0]!('a', 0);", "console.log(typeof viewed.callbacks[0]!('a', 0));", 1)
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	node := run{stdout: []byte("a\nnumber\n")}
	if diff := disagreement(node, onNode(t, file)); diff != "" {
		t.Fatal(diff)
	}
	_, err = lowered(t, file)
	if err == nil || !strings.Contains(err.Error(), "stage 0 can't lower a void call used as a value yet") {
		t.Fatalf("void result boundary changed: %v", err)
	}
}
