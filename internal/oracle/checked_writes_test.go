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

// Authored witnesses are .a; only the harness materializes their TypeScript inputs.
// Loading their original .a paths separately holds the stronger source-file promise.
func checkedWriteFixture(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "stage3/checked-writes", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name+".ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
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
			expected := run{exitCode: 70, stdout: []byte("before 16\n"), stderr: []byte("adamic: panic: " + checkedWriteFixtures[1].message + "\n")}
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
		rows = append(rows, fmt.Sprintf("| stage3/checked-writes/%s.a (TypeScript input) | %s | %s | %s | %s | %s | %s |", fixture.name, match[1], match[2], match[3], match[4], match[5], match[6]))
	}
	return rows
}

func TestCheckedWiderWritesAdamicRefuses(t *testing.T) {
	for _, name := range []string{"flags-fit", "string-fit", "boolean-fit", "diagnostic-fit", "parent-misfit"} {
		_, err := lowered(t, filepath.Join(repository, "stage3/checked-writes", name+".a"))
		if err == nil || !strings.Contains(err.Error(), "refuses") {
			t.Fatalf("%s .a: %v", name, err)
		}
	}
}

// A never-element container has no fitting element. Whole-container views still need
// element contracts, so admitting a scalar-field contract must not relax this refusal.
func TestCheckedWiderWritesKeepContainerRefusals(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(repository, "stage3/checked-writes/shared-never.a"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "shared-never.ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "which can write number where never is read") {
		t.Fatalf("never[] refusal: %v", err)
	}
}

func TestCheckedWiderWritesKeepSpreadOverrideRefusal(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(repository, "stage3/checked-writes/spread-override.a"))
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

func TestCheckedWiderWritesKeepReferenceStructureRefusal(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(repository, "stage3/checked-writes/reference-structure.a"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "reference-structure.ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "structural field contract") {
		t.Fatalf("reference contract refusal: %v", err)
	}
}
