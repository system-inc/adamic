package oracle

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCheckedViewOriginalEvaluatorPrimitivePairs(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, group := range []string{"evaluator", "evaluator-optional"} {
		variants := []string{"string", "number", "undefined", "wrong", "null", "missing"}
		if group == "evaluator-optional" {
			variants = append(variants, "absent")
		}
		for _, variant := range variants {
			t.Run(group+"/"+variant, func(t *testing.T) {
				input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/primitive-original/" + group + "-" + variant + ".a")
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/types.d.ts"))), 1)
				file := filepath.Join(t.TempDir(), "primitive.a")
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				text := map[string]string{"string": "word-built", "number": "42", "undefined": "undefined", "wrong": "true", "null": "null", "missing": "undefined", "absent": "undefined"}[variant] + "\n"
				if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
					t.Fatal("Node: " + diff)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				full := false
				for _, contract := range program.ViewContracts {
					if contract.Name != "EvaluatorResult<string | number | undefined>" {
						continue
					}
					var fields []string
					for _, field := range contract.Fields {
						fields = append(fields, field.Name)
					}
					slices.Sort(fields)
					full = slices.Equal(fields, []string{"hasExternalReferences", "isSyntacticallyString", "resolvedOtherFiles", "value"})
				}
				assertOriginalPrimitiveFields(t, directory, program, "EvaluatorResult<string | number | undefined>", "EvaluatorResult")
				if !full {
					t.Fatal("original evaluator fields reduced or absent")
				}
				label := "value.value"
				if group == "evaluator-optional" {
					label = "value?.value"
				}
				want := run{stdout: []byte(text)}
				if variant == "wrong" || variant == "null" {
					found := map[string]string{"wrong": "boolean", "null": "null"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + label + " matches no member of string | number | undefined; expected string | number | undefined, found " + found + "\n")}
				}
				if variant == "missing" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + label + " is not initialized; expected string | number | undefined, found missing\n")}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: stdout %q stderr %q", diff, got.stdout, got.stderr)
					}
				}
				if want.exitCode == 0 {
					got, binary := nativelyUncached(t, program)
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: %s", diff, got.stderr)
					}
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
				if variant == "wrong" {
					index := len(program.Strings)
					program.Strings = append(program.Strings, "unchecked")
					replacement := ir.Box{Value: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}}
					if count := dropLane4NamedHelperView(program, "value", replacement); count != 1 {
						t.Fatalf("expected one primitive member-check bypass, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || string(got.stdout) != "uncheckedunchecked\n" {
							t.Fatalf("mutant must run valid release code: stdout %q stderr %q", got.stdout, got.stderr)
						}
						t.Log("primitive member-check bypass caught by named refusal pin")
					}
				}
			})
		}
	}
}

func TestCheckedViewOriginalFinitePrimitiveFields(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, group := range []struct {
		name, root, field, declared string
		variants                    []string
	}{
		{"node-links", "NodeLinks", "isExhaustive", "0 | boolean | undefined", []string{"true", "false", "zero", "undefined", "wrong-number", "wrong-string", "null", "missing"}},
		{"emit-node", "EmitNode", "constantValue", "string | number | undefined", []string{"string", "number", "undefined", "wrong", "null", "missing"}},
		{"emit-node-optional", "EmitNode", "constantValue", "string | number | undefined", []string{"string", "number", "undefined", "wrong", "null", "missing", "absent"}},
	} {
		for _, variant := range group.variants {
			t.Run(group.name+"/"+variant, func(t *testing.T) {
				input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/primitive-original/" + group.name + "-" + variant + ".a")
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/types.d.ts"))), 1)
				file := filepath.Join(t.TempDir(), "primitive.a")
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				text := map[string]string{"true": "true", "false": "false", "zero": "0", "undefined": "undefined", "wrong-number": "1", "wrong-string": "wrong", "null": "null", "missing": "undefined", "string": "word-built", "number": "42", "wrong": "true", "absent": "undefined"}[variant] + "\n"
				if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
					t.Fatal("Node: " + diff)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				root := false
				for _, contract := range program.ViewContracts {
					if contract.Name == group.root {
						for _, field := range contract.Fields {
							root = root || field.Name == group.field
						}
					}
				}
				assertOriginalPrimitiveFields(t, directory, program, group.root, group.root)
				if !root {
					t.Fatal("original complete receiver contract absent")
				}
				label := "value." + group.field
				if group.name == "emit-node-optional" {
					label = "value?." + group.field
				}
				want := run{stdout: []byte(text)}
				if strings.HasPrefix(variant, "wrong") || variant == "null" {
					found := map[string]string{"wrong-number": "number", "wrong-string": "string", "wrong": "boolean", "null": "null"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + label + " matches no member of " + group.declared + "; expected " + group.declared + ", found " + found + "\n")}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: stdout %q stderr %q", diff, got.stdout, got.stderr)
					}
				}
				if want.exitCode == 0 {
					got, binary := nativelyUncached(t, program)
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: %s", diff, got.stderr)
					}
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
				if variant == "wrong-number" || variant == "wrong" {
					index := len(program.Strings)
					program.Strings = append(program.Strings, "unchecked")
					replacement := ir.Box{Value: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}}
					if count := dropLane4NamedHelperView(program, group.field, replacement); count != 1 {
						t.Fatalf("want one member-check bypass, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || string(got.stdout) != "uncheckedunchecked\n" {
							t.Fatalf("mutant must run valid release code: stdout %q stderr %q", got.stdout, got.stderr)
						}
						t.Log("member/literal-check bypass caught by named refusal pin")
					}
				}
			})
		}
	}
}

func assertOriginalPrimitiveFields(t *testing.T, directory string, program *ir.Program, name, sourceName string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(directory, "primitive-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Commit string              `json:"upstream_commit"`
		Fields map[string][]string `json:"fields"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatal("original primitive provenance changed")
	}
	for _, contract := range program.ViewContracts {
		if contract.Name != name {
			continue
		}
		fields := []string{}
		for _, field := range contract.Fields {
			fields = append(fields, field.Name)
		}
		slices.Sort(fields)
		if !slices.Equal(fields, manifest.Fields[sourceName]) {
			t.Fatalf("original %s receiver fields differ: got %v want %v", name, fields, manifest.Fields[sourceName])
		}
		return
	}
	t.Fatal("original primitive receiver absent: " + name)
}
