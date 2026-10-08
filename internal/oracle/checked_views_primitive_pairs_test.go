package oracle

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"testing"
)

func TestCheckedViewBrandCandidatePairs(t *testing.T) {
	data, err := os.ReadFile("../../stage3/interface-downcasts/lane4/brand-pair-fixtures.json")
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
