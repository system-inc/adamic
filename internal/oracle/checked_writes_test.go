package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

var checkedWriteFixtures = []struct{ name, stdout, message string }{
	{"container-slice-fit", "2\n", ""},
	{"container-slice-misfit", "", "write failed: values[0] expects 1 | 2, got 3"},
	{"container-map-values-fit", "2\n", ""},
	{"container-map-values-misfit", "", "write failed: values[0] expects 1 | 2, got 3"},
	{"container-boolean-fit", "true\n", ""},
	{"container-boolean-misfit", "", "write failed: values[] expects true, got false"},
	{"container-fill-fit", "2\n", ""},
	{"container-fill-misfit", "", "write failed: values[] expects 1 | 2, got 3"},
	{"container-map-object-fit", "1\n", ""},
	{"container-map-object-misfit", "", "write failed: values[value] expects DiagnosticWithLocation, got object"},
	{"container-alias-fit", "2\n", ""},
	{"container-alias-misfit", "", "write failed: item.count expects number, got undefined"},
	{"container-number-fit", "2\n", ""},
	{"container-number-misfit", "", "write failed: values[0] expects 1 | 2, got 3"},
	{"container-string-fit", "right\n", ""},
	{"container-string-misfit", "", "write failed: values[] expects \"left\" | \"right\", got outside"},
	{"container-map-fit", "2\n", ""},
	{"container-map-misfit", "", "write failed: values[value] expects 1 | 2, got 3"},
	{"container-splice-fit", "2\n", ""},
	{"container-splice-misfit", "", "write failed: values[] expects 1 | 2, got 3"},
	{"container-object-fit", "1\n", ""},
	{"container-object-misfit", "", "write failed: values[] expects DiagnosticWithLocation, got object"},
	{"container-nested-fit", "1\n", ""},
	{"container-nested-misfit", "", "write failed: values[] expects (1 | 2)[], got array"},
	{"flow-node-fit", "call\n", ""},
	{"flow-node-misfit", "", "write failed: view.node expects BinaryExpression | CallExpression, got object"},
	{"flow-undefined-fit", "true\n", ""},
	{"flow-undefined-misfit", "", "write failed: view.node expects undefined, got object"},
	{"flow-array-fit", "true\n", ""},
	{"flow-array-misfit", "", "write failed: view.antecedents expects undefined, got array"},
	{"never-nullable-fit", "0\n", ""},
	{"never-nullable-misfit", "", "write failed: view[] expects never, got object"},
	{"never-index-misfit", "", "write failed: values[0] expects never, got 0"},
	{"never-splice-misfit", "", "write failed: values[] expects never, got 0"},
	{"never-wide-fit", "0\n1\n", ""},
	{"never-number-fit", "0\n0\n", ""},
	{"never-number-misfit", "", "write failed: values[] expects never, got 0"},
	{"never-string-fit", "0\n0\n", ""},
	{"never-string-misfit", "", "write failed: values[] expects never, got badbad"},
	{"never-object-fit", "0\n0\n", ""},
	{"never-object-misfit", "", "write failed: values[] expects never, got object"},
	{"diagnostic-alias-fit", "16\n", ""},
	{"diagnostic-alias-misfit", "", "write failed: wide.flags expects 16, got 0"},
	{"diagnostic-proof-fit", "newnew 16\n", ""},
	{"diagnostic-proof-misfit", "", "write failed: (view as Mutable<T>).file expects SynthesizedSourceFile, got object"},
	{"diagnostic-rich-fit", "newnew 3 2 3 4\nNaN -Infinity\n", ""},
	{"diagnostic-rich-misfit", "", "write failed: view.file expects SourceFile, got undefined"},
	{"diagnostic-related-fit", "7\n", ""},
	{"diagnostic-related-misfit", "", "write failed: view.start expects number, got undefined"},
	{"diagnostic-nested-fit", "newnew 16\n", ""},
	{"diagnostic-nested-misfit", "", "write failed: view.file expects SynthesizedSourceFile, got object"},
	{"diagnostic-inner-fit", "32\n", ""},
	{"diagnostic-inner-misfit", "", "write failed: view.file.metadata.flags expects 16 | 32, got 0"},
	{"reference-structure", "", "write failed: (node as Mutable<T>).parent expects { readonly name: \"left\"; }, got object"},
	{"flags-fit", "16\n268435472\n", ""},
	{"flags-misfit", "before 16\n", "write failed: (node as Mutable<T>).flags expects NodeFlags.Synthesized, got 0"},
	{"string-fit", "right\n", ""},
	{"string-misfit", "", "write failed: view.text expects \"left\" | \"right\", got outsideoutside"},
	{"boolean-fit", "true\n", ""},
	{"boolean-misfit", "", "write failed: view.enabled expects true, got false"},
	{"diagnostic-fit", "inputinput\n", ""},
	{"diagnostic-misfit", "", "write failed: view.file expects SourceFile, got undefined"},
	{"number-fit", "2 7\n", ""},
	{"number-misfit", "", "write failed: view.count expects 1 | 2, got 3"},
	{"compound-fit", "1\n", ""},
	{"compound-misfit", "", "write failed: view.count expects 0 | 1, got 2"},
	{"nullable-number-fit", "true\n2\n", ""},
	{"nullable-number-misfit", "", "write failed: view.count expects 1 | 2 | undefined, got 3"},
	{"range-fit", "0 1\n", ""},
	{"range-misfit", "", "write failed: range.end expects 1, got 2"},
	{"receiver-fit", "2 7\n", ""},
	{"parent-fit", "parentparent\n", ""},
	{"parent-misfit", "", "write failed: (copy as Mutable<T>).parent expects Parent, got undefined"},
}

// Runtime witnesses are real TypeScript files. Adamic refusal controls are separate.
func checkedWriteFixture(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/checked-writes", name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program, path
}

func TestCheckedWiderWrites(t *testing.T) {
	for _, fixture := range checkedWriteFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			program, path := checkedWriteFixture(t, fixture.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %#v", truth)
			}
			want := truth
			if fixture.message != "" {
				want = run{exitCode: 70, stdout: []byte(fixture.stdout), stderr: []byte("adamic: panic: " + fixture.message + "\n")}
			} else if string(truth.stdout) != fixture.stdout {
				t.Fatalf("Node stdout %q", truth.stdout)
			}
			sanitized, binary := nativelyUncached(t, program)
			for name, got := range map[string]run{"native sanitized": sanitized, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s; got %#v", name, difference, got)
				}
			}
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			t.Logf("Node exit %d stdout %q; checked exit %d stderr %q; %d checks", truth.exitCode, truth.stdout, want.exitCode, want.stderr, len(program.WriteChecks))
		})
	}
}

func TestCheckedWiderWriteMutants(t *testing.T) {
	for _, mutant := range []string{"drop check", "drop literal set"} {
		t.Run(mutant, func(t *testing.T) {
			program, path := checkedWriteFixture(t, "flags-misfit")
			expected := run{exitCode: 70, stdout: []byte("before 16\n"), stderr: []byte("adamic: panic: " + "write failed: (node as Mutable<T>).flags expects NodeFlags.Synthesized, got 0" + "\n")}
			if mutant == "drop check" {
				// Mutate real stores, not a test predicate or a compile-time rejection.
				var visit func(reflect.Value)
				visit = func(v reflect.Value) {
					if !v.IsValid() {
						return
					}
					if v.Kind() == reflect.Interface {
						if write, ok := v.Interface().(ir.SetProperty); ok {
							write.WriteCheck = ""
							v.Set(reflect.ValueOf(write))
							return
						}
						visit(v.Elem())
						return
					}
					if v.Kind() == reflect.Struct {
						for i := 0; i < v.NumField(); i++ {
							visit(v.Field(i))
						}
					}
					if v.Kind() == reflect.Slice {
						for i := 0; i < v.Len(); i++ {
							visit(v.Index(i))
						}
					}
				}
				visit(reflect.ValueOf(&program.Main).Elem())
				visit(reflect.ValueOf(&program.Functions).Elem())
			} else {
				change := func(value ir.Expression) ir.Expression {
					if literal, ok := value.(ir.ObjectLiteral); ok {
						for i := range literal.Fields {
							if c := literal.Fields[i].Contract; c != nil && c.Declared == "NodeFlags.Synthesized" {
								copy := *c
								copy.Allowed = nil
								literal.Fields[i].Contract = &copy
							}
						}
						return literal
					}
					return value
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), change)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), change)
			}
			truth := onNode(t, path)
			got, binary := nativelyUncached(t, program)
			if disagreement(expected, got) == "" {
				t.Fatal("mutant survived")
			}
			if d := disagreement(truth, got); d != "" {
				t.Fatalf("mutant must run valid Node behavior, got %s %#v", d, got)
			}
			if d := disagreement(truth, onJavaScriptBackend(t, program)); d != "" {
				t.Fatal(d)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("caught %s: expected exit 70, mutant exit %d stdout %q", mutant, got.exitCode, got.stdout)
		})
	}
}

func checkedWriteCounts(t *testing.T) []string {
	rows := []string{}
	for _, fixture := range checkedWriteFixtures {
		program, _ := checkedWriteFixture(t, fixture.name)
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		name, args := pinnedStack(binary)
		got := execute(t, name, args...)
		match := countsLine.FindSubmatch(got.stderr)
		if match == nil {
			t.Fatalf("missing counts: %#v", got)
		}
		rows = append(rows, fmt.Sprintf("| stage3/checked-writes/%s.ts | %s | %s | %s | %s | %s | %s |", fixture.name, match[1], match[2], match[3], match[4], match[5], match[6]))
	}
	rows = append(rows, counted(t, "stage3/checked-writes/proven-number.a", false, nil, false, false), counted(t, "stage3/checked-writes/proven-containers.a", false, nil, false, false))
	return rows
}

func TestCheckedWiderWritesAdamicRefuses(t *testing.T) {
	for _, name := range []string{"flags-refused", "string-refused", "shared-never", "flow-refused", "array-refused", "map-refused"} {
		path := filepath.Join(repository, "stage3/checked-writes", name+".a")
		expected, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".refused")
		if err != nil {
			t.Fatal(err)
		}
		_, err = lowered(t, path)
		if err == nil || !strings.Contains(err.Error(), strings.TrimSpace(string(expected))) {
			t.Fatalf("%s .a: %v; want %s", name, err, expected)
		}
	}
}

func TestCheckedWiderWritesProvenAdamic(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/checked-writes/proven-number.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.WriteChecks) != 0 {
		t.Fatal("proven write inserted a runtime check")
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "16\n" {
		t.Fatalf("Node: %#v", truth)
	}
	got, binary := nativelyUncached(t, program)
	for name, result := range map[string]run{"native sanitized": got, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if d := disagreement(truth, result); d != "" {
			t.Fatalf("%s: %s", name, d)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

// Clearing the bottom allocation contract must admit the original Node store.
func TestCheckedNeverContractMutant(t *testing.T) {
	program, path := checkedWriteFixture(t, "never-number-misfit")
	changes := 0
	change := func(value ir.Expression) ir.Expression {
		if literal, ok := value.(ir.ArrayLiteral); ok && literal.Never {
			literal.Never = false
			changes++
			return literal
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), change)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), change)
	if changes == 0 {
		t.Fatal("mutant changed no allocation contract")
	}
	truth := onNode(t, path)
	got, binary := nativelyUncached(t, program)
	if got.exitCode == 70 {
		t.Fatal("mutant survived")
	}
	if d := disagreement(truth, got); d != "" {
		t.Fatal(d)
	}
	if d := disagreement(truth, onJavaScriptBackend(t, program)); d != "" {
		t.Fatal(d)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("caught drop never contract: expected exit 70, mutant exit %d stdout %q", got.exitCode, got.stdout)
}

func TestCheckedWiderWritesKeepSpreadOverrideRefusal(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(repository, "stage3/checked-writes/spread-override.ts"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "spread-override.ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "overriding a checked field contract") {
		t.Fatalf("spread contract refusal: %v", err)
	}
}

func TestCheckedDiagnosticReferenceMutants(t *testing.T) {
	for _, mutant := range []struct{ name, fixture string }{
		{"drop alias check", "diagnostic-alias-misfit"},
		{"drop allocation proof", "diagnostic-proof-misfit"},
		{"drop nested contract", "diagnostic-nested-misfit"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			program, path := checkedWriteFixture(t, mutant.fixture)
			var pinned string
			for _, fixture := range checkedWriteFixtures {
				if fixture.name == mutant.fixture {
					pinned = fixture.message
				}
			}
			expected := run{exitCode: 70, stderr: []byte("adamic: panic: " + pinned + "\n")}
			changes := 0
			change := func(value ir.Expression) ir.Expression {
				literal, ok := value.(ir.ObjectLiteral)
				if !ok {
					return value
				}
				for index, field := range literal.Fields {
					contract := field.Contract
					if contract == nil || contract.Declared != "SynthesizedSourceFile" {
						continue
					}
					copy := *contract
					if mutant.name == "drop allocation proof" {
						copy.Reference = false
					} else {
						copy.Fields = nil
						copy.Structural = true
					}
					literal.Fields[index].Contract = &copy
					changes++
				}
				return literal
			}
			if mutant.name == "drop alias check" {
				var visit func(reflect.Value)
				visit = func(value reflect.Value) {
					switch value.Kind() {
					case reflect.Interface:
						if value.IsNil() {
							return
						}
						if write, ok := value.Interface().(ir.SetProperty); ok && write.Name == "flags" && write.WriteCheck != "" {
							write.WriteCheck = ""
							value.Set(reflect.ValueOf(write))
							changes++
							return
						}
						visit(value.Elem())
					case reflect.Struct:
						for i := 0; i < value.NumField(); i++ {
							visit(value.Field(i))
						}
					case reflect.Slice:
						for i := 0; i < value.Len(); i++ {
							visit(value.Index(i))
						}
					}
				}
				visit(reflect.ValueOf(&program.Main).Elem())
				visit(reflect.ValueOf(&program.Functions).Elem())
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), change)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), change)
			if changes == 0 {
				t.Fatal("mutant did not change a checked store or allocation contract")
			}
			truth := onNode(t, path)
			got, binary := nativelyUncached(t, program)
			if disagreement(expected, got) == "" {
				t.Fatal("mutant survived the pinned failure")
			}
			if d := disagreement(truth, got); d != "" {
				t.Fatalf("mutant must run valid Node behavior: %s", d)
			}
			if d := disagreement(truth, onJavaScriptBackend(t, program)); d != "" {
				t.Fatal(d)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("caught %s: expected exit 70, mutant exit %d stdout %q", mutant.name, got.exitCode, got.stdout)
		})
	}
}

func TestCheckedWiderWritesProvenContainers(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/checked-writes/proven-containers.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.WriteChecks) != 0 {
		t.Fatal("proven container stores inserted checks")
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "2 1\n" {
		t.Fatalf("Node: %#v", truth)
	}
	got, binary := nativelyUncached(t, program)
	for name, result := range map[string]run{"native sanitized": got, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if d := disagreement(truth, result); d != "" {
			t.Fatalf("%s: %s", name, d)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestCheckedFlowContainerContractMutants(t *testing.T) {
	for _, fixture := range []string{"flow-node-misfit", "flow-undefined-misfit", "flow-array-misfit", "container-number-misfit", "container-string-misfit", "container-map-misfit", "container-splice-misfit", "container-object-misfit", "container-nested-misfit", "container-boolean-misfit", "container-fill-misfit", "container-map-object-misfit", "container-alias-misfit"} {
		t.Run(fixture, func(t *testing.T) {
			program, path := checkedWriteFixture(t, fixture)

			changes := 0
			change := func(value ir.Expression) ir.Expression {
				if container, ok := value.(ir.ContractContainer); ok && container.Contract != nil && fixture != "container-alias-misfit" {
					container.Contract = nil
					changes++
					return container
				}
				if strings.HasPrefix(fixture, "flow-") || fixture == "container-alias-misfit" {
					if object, ok := value.(ir.ObjectLiteral); ok {
						for i, field := range object.Fields {
							if field.Contract != nil && (field.Name == "node" || field.Name == "antecedents" || field.Name == "count") {
								copy := *field.Contract
								copy.Reference = false
								copy.NullishOnly = false
								if fixture == "container-alias-misfit" {
									copy.Nullable = true
								}
								if fixture == "flow-array-misfit" {
									copy.Kind = ir.Array
								}
								object.Fields[i].Contract = &copy
								changes++
							}
						}
						return object
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), change)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), change)
			if changes == 0 {
				t.Fatal("mutant changed no allocation contract")
			}
			truth := onNode(t, path)
			got, binary := nativelyUncached(t, program)
			if d := disagreement(truth, got); d != "" {
				t.Fatalf("mutant must match Node: %s; got %#v", d, got)
			}
			if d := disagreement(truth, onJavaScriptBackend(t, program)); d != "" {
				t.Fatal(d)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("caught drop allocation contract: expected exit 70; mutant exit %d stdout %q", got.exitCode, got.stdout)
		})
	}
}
