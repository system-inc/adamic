package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Complete original types are supplied by the shared declaration-only adapter.
func TestCheckedViewBrandsOriginalPairs(t *testing.T) {
	declarations := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_BRAND_ORIGINAL_DECLS to pinned complete declarations")
	}
	data, err := os.ReadFile(filepath.Join(declarations, "brand-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Commit       string              `json:"upstream_commit"`
		Declarations map[string]string   `json:"declarations"`
		Fields       map[string][]string `json:"fields"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Declarations) != 78 {
		t.Fatal("original provenance changed")
	}
	for file, digest := range manifest.Declarations {
		data, err := os.ReadFile(filepath.Join(declarations, file))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			t.Fatal("declaration drift: " + file)
		}
	}
	for _, pair := range []struct{ typ, field string }{{"Identifier", "escapedText"}, {"Symbol", "escapedName"}} {
		for _, variant := range []string{"good", "internal", "undefined", "wrong", "null", "missing"} {
			t.Run(pair.typ+"/"+variant, func(t *testing.T) {
				name := strings.ToLower(pair.typ) + "-" + variant
				input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/original/" + name + ".a")
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
				file := filepath.Join(t.TempDir(), name+".a")
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				text := map[string]string{"good": "word-built", "internal": "__importAttributes", "undefined": "undefined", "wrong": "42", "null": "null", "missing": "undefined"}[variant] + "\n"
				if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
					t.Fatal("Node: " + diff)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				complete := false
				for _, contract := range program.ViewContracts {
					if contract.Name != pair.typ {
						continue
					}
					fields := []string{}
					for _, field := range contract.Fields {
						fields = append(fields, field.Name)
					}
					slices.Sort(fields)
					complete = complete || slices.Equal(fields, manifest.Fields[pair.typ]) && len(fields) > 0
				}
				if !complete {
					t.Fatal("original field set omitted: " + pair.typ)
				}
				want := run{stdout: []byte(text)}
				if variant == "wrong" || variant == "null" {
					found := map[string]string{"wrong": "number", "null": "null"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value." + pair.field + " is not a __String; expected __String, found " + found + "\n")}
				}
				if variant == "missing" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value." + pair.field + " is not initialized; expected __String, found missing\n")}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s; got %#v", diff, got)
					}
				}
				if want.exitCode == 0 {
					got, binary := nativelyUncached(t, program)
					if diff := disagreement(want, got); diff != "" {
						t.Fatal(diff)
					}
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
				if variant == "wrong" {
					index := len(program.Strings)
					program.Strings = append(program.Strings, "unchecked")
					replacement := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}
					if count := dropLane4NamedHelperView(program, pair.field, replacement); count != 1 {
						t.Fatalf("expected one helper mutation, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || string(got.stdout) != "uncheckedunchecked\n" {
							t.Fatalf("mutant must execute release code: %#v", got)
						}
						t.Logf("member-check bypass caught: exit %d stdout %q", got.exitCode, got.stdout)
					}
				}
			})
		}
	}
}
