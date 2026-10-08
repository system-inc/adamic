package oracle

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewBrandCandidatePairs(t *testing.T) {
	data, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/brand-pair-fixtures.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pairs []struct {
		ID       string
		Fixture  string
		Field    string
		Optional bool
	}
	if err = json.Unmarshal(data, &pairs); err != nil {
		t.Fatal(err)
	}
	for _, pair := range pairs {
		for _, variant := range []string{"good", "undefined", "wrong", "null", "missing"} {
			t.Run(pair.ID+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane4/brand-pairs/"+pair.Fixture+"-"+variant)
				text := map[string]string{"good": "word", "undefined": "undefined", "wrong": "42", "null": "null", "missing": "undefined"}[variant]
				if difference := disagreement(run{stdout: []byte(text + "\n")}, onNode(t, path)); difference != "" {
					t.Fatal("source Node: " + difference)
				}
				want := run{stdout: []byte(text + "\n")}
				expected := "__String"
				if pair.Optional {
					expected = "__String | undefined"
				}
				if variant == "wrong" || variant == "null" {
					found := map[string]string{"wrong": "number", "null": "null"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value." + pair.Field + " is not a " + expected + "; expected " + expected + ", found " + found + "\n")}
				}
				if variant == "missing" && !pair.Optional {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value." + pair.Field + " is not initialized; expected " + expected + ", found missing\n")}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s; got %#v", difference, got)
					}
				}
				if want.exitCode == 0 {
					got, _ := nativelyUncached(t, program)
					if difference := disagreement(want, got); difference != "" {
						t.Fatal(difference)
					}
				}
				if variant == "wrong" {
					index := len(program.Strings)
					program.Strings = append(program.Strings, "unchecked")
					replacement := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}
					if count := dropLane4NamedHelperView(program, pair.Field, replacement); count != 1 {
						t.Fatalf("want one mutated helper field, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 {
							t.Fatalf("mutant must run valid release code: %#v", got)
						}
						if disagreement(want, got) == "" {
							t.Fatal("unchecked helper read escaped pin")
						}
						t.Logf("read-check bypass caught: exit %d stdout %q", got.exitCode, got.stdout)
					}
				}
			})
		}
	}
}

// This reduced finite-key contract is component coverage, not a tsc pair.
// Real CompilerOptions has a string index signature, so keyof also admits numbers.
func TestCheckedViewFiniteStringKeyComponent(t *testing.T) {
	for _, pair := range []string{"9761", "97180"} {
		for _, variant := range []string{"target", "strict", "undefined", "missing", "wrong", "number", "null"} {
			t.Run(pair+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane4/primitive-pairs/"+pair+"-skippedOn-"+variant)
				text := map[string]string{"target": "target", "strict": "strict", "undefined": "undefined", "missing": "undefined", "wrong": "notAnOption", "number": "42", "null": "null"}[variant]
				if difference := disagreement(run{stdout: []byte(text + "\n")}, onNode(t, path)); difference != "" {
					t.Fatal("source Node: " + difference)
				}
				want := run{stdout: []byte(text + "\n")}
				declared := "keyof CompilerOptions | undefined"
				if variant == "wrong" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.skippedOn matches no member of " + declared + "; expected " + declared + ", found string\n")}
				}
				if variant == "number" || variant == "null" {
					found := map[string]string{"number": "number", "null": "null"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.skippedOn matches no member of " + declared + "; expected " + declared + ", found " + found + "\n")}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s; got %#v", difference, got)
					}
				}
				if want.exitCode == 0 {
					got, _ := nativelyUncached(t, program)
					if difference := disagreement(want, got); difference != "" {
						t.Fatal(difference)
					}
				}
				if variant == "wrong" {
					index := len(program.Strings)
					program.Strings = append(program.Strings, "unchecked")
					replacement := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}
					if count := dropLane4NamedHelperView(program, "skippedOn", replacement); count != 1 {
						t.Fatalf("want one mutated helper field, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 {
							t.Fatalf("mutant must run valid release code: %#v", got)
						}
						if disagreement(want, got) == "" {
							t.Fatal("literal-member read bypass escaped pin")
						}
						t.Logf("literal-member bypass caught: exit %d stdout %q", got.exitCode, got.stdout)
					}
				}
			})
		}
	}
}

// Certified homogeneous producers use checked primitive reads. Heterogeneous
// array literals still require their own producer-storage adapter.
func TestCheckedViewPrimitiveArrayPairStorage(t *testing.T) {
	for _, variant := range []string{"string", "number", "plain-number", "plain-string", "wrong", "bounds"} {
		t.Run(variant, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-pairs/6849-element-" + variant + ".a"))
			if err != nil {
				t.Fatal(err)
			}
			text := map[string]string{"string": "word", "number": "42", "plain-number": "42", "plain-string": "word", "wrong": "true", "bounds": "undefined"}[variant]
			if difference := disagreement(run{stdout: []byte(text + "\n")}, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), loaded)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(text + "\n")}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if variant == "wrong" || variant == "bounds" {
					if got.exitCode != 70 || !strings.Contains(string(got.stderr), "values") {
						t.Fatalf("array payload/presence mutant ran on: %#v", got)
					}
					continue
				}
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if variant != "wrong" && variant != "bounds" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// Integration's nullable selector now holds this open-key shape at runtime.
// This reduced index-signature witness is not original CompilerOptions coverage.
func TestCheckedViewOpenCompilerOptionKeySelection(t *testing.T) {
	for _, pair := range []string{"9761", "97180"} {
		for _, variant := range []string{"string", "number", "undefined", "missing", "wrong", "null"} {
			t.Run(pair+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane4/primitive-pairs/"+pair+"-skippedOn-open-"+variant)
				text := map[string]string{"string": "notAnOption", "number": "42", "undefined": "undefined", "missing": "undefined", "wrong": "true", "null": "null"}[variant]
				if difference := disagreement(run{stdout: []byte(text + "\n")}, onNode(t, path)); difference != "" {
					t.Fatal("source Node: " + difference)
				}
				want := run{stdout: []byte(text + "\n")}
				if variant == "wrong" || variant == "null" {
					found := map[string]string{"wrong": "boolean", "null": "null"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.skippedOn matches no member of keyof CompilerOptions | undefined; expected keyof CompilerOptions | undefined, found " + found + "\n")}
				}
				actual, binary := nativelyUncached(t, program)
				for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s; got %#v", difference, got)
					}
				}
				if want.exitCode == 0 {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
				if want.exitCode != 0 {
					index := len(program.Strings)
					program.Strings = append(program.Strings, "unchecked")
					replacement := ir.Box{Value: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}}
					if count := dropLane4NamedHelperView(program, "skippedOn", replacement); count != 1 {
						t.Fatalf("want one read mutation, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || string(got.stdout) != "uncheckedunchecked\n" {
							t.Fatalf("mutant must run valid release code: %#v", got)
						}
						t.Logf("open-key read bypass caught: exit %d stdout %q", got.exitCode, got.stdout)
					}
				}
			})
		}
	}
}
