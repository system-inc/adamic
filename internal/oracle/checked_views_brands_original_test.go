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
		PrivateHash  string              `json:"private_sha256"`
		NamesHash    string              `json:"names_sha256"`
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
	if manifest.NamesHash != "" {
		data, err := os.ReadFile(filepath.Join(declarations, "brand-names.d.ts"))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != manifest.NamesHash {
			t.Fatal("namespace declaration drift")
		}
	}
	if manifest.PrivateHash != "" {
		data, err := os.ReadFile(filepath.Join(declarations, "brand-private.d.ts"))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != manifest.PrivateHash {
			t.Fatal("private declaration drift")
		}
	}
	for _, pair := range []struct {
		typ, field string
		optional   bool
	}{{"Identifier", "escapedText", false}, {"Symbol", "escapedName", false}, {"PrivateIdentifier", "escapedText", false}, {"Identifier | PrivateIdentifier", "escapedText", false}, {"TransientSymbol", "escapedName", false}, {"MemberName", "escapedText", false}, {"UnionType", "keyPropertyName", true}, {"SourceFile", "localJsxFragmentNamespace", true}, {"SourceFile", "localJsxNamespace", true}, {"SymbolLinks", "typeOnlyExportStarName", true}, {"WideningContext", "propertyName", true}, {"typeof JsxNames", "IntrinsicElements", false}, {"typeof JsxNames", "IntrinsicAttributes", false}, {"typeof JsxNames", "Element", false}, {"typeof JsxNames", "IntrinsicClassAttributes", false}, {"typeof JsxNames", "JSX", false}, {"typeof JsxNames", "ElementAttributesPropertyNameContainer", false}, {"typeof JsxNames", "ElementChildrenAttributeNameContainer", false}, {"typeof JsxNames", "ElementClass", false}, {"typeof JsxNames", "ElementType", false}, {"typeof JsxNames", "LibraryManagedAttributes", false}, {"typeof ReactNames", "Fragment", false}, {"UniqueESSymbolType", "escapedName", false}, {"GeneratedIdentifier", "escapedText", false}, {"GeneratedPrivateIdentifier", "escapedText", false}, {"Identifier | undefined", "escapedText", false}, {"Symbol | undefined", "escapedName", false}, {"Identifier@34691", "escapedText", false}, {"PrivateIdentifier@55713", "escapedText", false}, {"Identifier@46232", "escapedText", false}, {"Mutable<Identifier>", "escapedText", false}, {"ActiveLabel", "name", false}, {"RenamedBinding", "name", false}, {"LeftHandSideExpression & Identifier", "escapedText", false}} {
		variants := []string{"good", "internal", "undefined", "wrong", "null", "missing"}
		union := pair.typ == "Identifier | PrivateIdentifier" || pair.typ == "MemberName"
		if union {
			variants = append(variants, "good-other", "internal-other", "undefined-other", "wrong-other", "null-other", "missing-other")
		}
		for _, fixtureVariant := range variants {
			variant := strings.TrimSuffix(fixtureVariant, "-other")
			t.Run(pair.typ+"/"+pair.field+"/"+fixtureVariant, func(t *testing.T) {
				name := strings.ToLower(pair.typ) + "-" + fixtureVariant
				if pair.optional || strings.HasPrefix(pair.typ, "typeof ") {
					name = strings.ToLower(pair.typ) + "-" + pair.field + "-" + fixtureVariant
				}
				fixture := "../../stage3/interface-downcasts/lane4/original/" + name + ".a"
				if pair.typ == "LeftHandSideExpression & Identifier" {
					fixture = "../../stage3/interface-downcasts/lane4/primitive-original/intersection-identifier-" + fixtureVariant + ".a"
				}
				input, err := os.ReadFile(fixture)
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
				bound = strings.Replace(bound, "'original-tsc-names'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "brand-names.d.ts"))), 1)
				bound = strings.Replace(bound, "'original-tsc-private'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "brand-private.d.ts"))), 1)
				bound = strings.Replace(bound, "'original-tsc-utilities'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/utilities.d.ts"))), 1)
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
				if pair.typ == "ActiveLabel" || pair.typ == "RenamedBinding" {
					found := false
					for _, parent := range program.ViewContracts {
						if parent.Name != "ArrowFunction" {
							continue
						}
						for _, field := range parent.Fields {
							if field.Name == "name" && program.ViewContracts[field.Contract-1].Unsupported == "never" {
								found = true
							}
						}
					}
					if !found {
						t.Fatal("original ArrowFunction.name: never collision witness missing")
					}
					t.Log("original unrelated fallback source: ArrowFunction.name: never; supported helper read retains __String check")
				}
				roots := []string{strings.Split(strings.TrimSuffix(pair.typ, " | undefined"), "@")[0]}
				if pair.typ == "Identifier | PrivateIdentifier" || pair.typ == "MemberName" {
					root := "Identifier"
					if pair.typ == "MemberName" {
						root = "PrivateIdentifier"
					}
					if strings.HasSuffix(fixtureVariant, "-other") {
						if root == "Identifier" {
							root = "PrivateIdentifier"
						} else {
							root = "Identifier"
						}
					}
					roots = []string{root}
				}
				if pair.typ == "LeftHandSideExpression & Identifier" {
					roots = []string{"Identifier"}
				}
				for _, root := range roots {
					complete := false
					for _, contract := range program.ViewContracts {
						if contract.Name != root {
							continue
						}
						fields := []string{}
						for _, field := range contract.Fields {
							fields = append(fields, field.Name)
						}
						slices.Sort(fields)
						complete = complete || slices.Equal(fields, manifest.Fields[root]) && len(fields) > 0
					}
					if !complete {
						t.Fatal("original field set omitted: " + root)
					}
				}
				want := run{stdout: []byte(text)}
				expression := "value." + pair.field
				if strings.HasSuffix(pair.typ, " | undefined") {
					expression = "value?." + pair.field
				}
				expected := "__String"
				if pair.optional {
					expected += " | undefined"
				}
				if variant == "wrong" || variant == "null" {
					found := map[string]string{"wrong": "number", "null": "null"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + expression + " is not a " + expected + "; expected " + expected + ", found " + found + "\n")}
				}
				if variant == "missing" && !pair.optional {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + expression + " is not initialized; expected __String, found missing\n")}
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
