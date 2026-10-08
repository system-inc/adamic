package oracle

import (
	"crypto/sha256"
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
				input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-original/" + group + "-" + variant + ".a"))
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
					found := map[string]string{"wrong": "boolean", "null": "null", "undefined": "undefined", "missing": "missing"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: " + label + " matches no member of string | number | undefined; expected string | number | undefined, found " + found + "\n")}
				}
				if variant == "missing" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: " + label + " is not initialized; expected string | number | undefined, found missing\n")}
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
				if variant == "undefined" {
					assertPrimitiveFirstMemberMutant(t, program, "value", false, "undefined\n")
				}
				if variant == "wrong" {
					if count := skipPrimitiveMemberChecks(program, "value"); count != 1 {
						t.Fatalf("want one primitive check mutation, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || string(got.stdout) != text {
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
		{"literal-union", "StringLiteralType | NumberLiteralType", "value", "string | number", []string{"string", "number", "wrong", "null", "undefined", "missing"}},
		{"diagnostic", "Diagnostic", "skippedOn", "keyof CompilerOptions | undefined", []string{"string", "number", "undefined", "wrong", "null", "missing"}},
		{"bundle-pending", "IncrementalBundleEmitBuildInfo", "pendingEmit", "IncrementalBuildInfoBundlePendingEmit | undefined", []string{"number", "false", "undefined", "wrong", "wrong-string", "null", "missing"}},
		{"resolved", "Resolved", "originalPath", "string | true | undefined", []string{"string", "true", "undefined", "wrong", "wrong-number", "null", "missing"}},
		{"graph-circular", "FlowGraphNode", "circular", "\"circularity\" | boolean", []string{"true", "false", "circularity", "wrong", "wrong-string", "undefined", "null", "missing"}},
		{"watch-version", "FilePresentOnHost | FilePresenceUnknownOnHost", "version", "string | false", []string{"string", "false", "wrong", "wrong-number", "undefined", "null", "missing"}},
		{"package-peer", "PackageJsonInfoContents", "peerDependencies", "string | false | undefined", []string{"string", "false", "undefined", "wrong", "wrong-number", "null", "missing"}},
		{"reusable-file", "ReusableDiagnosticRelatedInformation", "file", "string | false | undefined", []string{"string", "false", "undefined", "wrong", "wrong-number", "null", "missing"}},
		{"reusable-skipped", "ReusableDiagnostic", "skippedOn", "keyof CompilerOptions | undefined", []string{"string", "number", "undefined", "wrong", "null", "missing"}},
		{"node-links", "NodeLinks", "isExhaustive", "0 | boolean | undefined", []string{"true", "false", "zero", "undefined", "wrong-number", "wrong-string", "null", "missing"}},
		{"emit-node", "EmitNode", "constantValue", "string | number | undefined", []string{"string", "number", "undefined", "wrong", "null", "missing"}},
		{"emit-node-optional", "EmitNode", "constantValue", "string | number | undefined", []string{"string", "number", "undefined", "wrong", "null", "missing", "absent"}},
	} {
		for _, variant := range group.variants {
			t.Run(group.name+"/"+variant, func(t *testing.T) {
				input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-original/" + group.name + "-" + variant + ".a"))
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/types.d.ts"))), 1)
				for _, module := range []string{"types", "builder", "moduleNameResolver"} {
					bound = strings.ReplaceAll(bound, "'original-tsc-"+module+"'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler", module+".d.ts"))))
				}
				bound = strings.ReplaceAll(bound, "'original-tsc-private'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "primitive-private.d.ts"))))
				file := filepath.Join(t.TempDir(), "primitive.a")
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				text := map[string]string{"true": "true", "false": "false", "zero": "0", "undefined": "undefined", "wrong-number": "1", "wrong-string": "wrong", "null": "null", "missing": "undefined", "string": "word-built", "number": "42", "wrong": "true", "absent": "undefined"}[variant] + "\n"
				if variant == "wrong-number" && group.name != "node-links" {
					text = "42\n"
				}
				if group.name == "resolved" && variant == "wrong" {
					text = "false\n"
				}
				if group.name == "bundle-pending" && variant == "number" {
					text = "1\n"
				}
				if group.name == "graph-circular" {
					if variant == "circularity" {
						text = "circularity\n"
					}
					if variant == "wrong" {
						text = "42\n"
					}
				}
				if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
					t.Fatal("Node: " + diff)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				rootName := group.root
				if group.name == "literal-union" {
					rootName = "StringLiteralType"
					if variant == "number" {
						rootName = "NumberLiteralType"
					}
				}
				if group.name == "watch-version" {
					rootName = "FilePresentOnHost"
					if variant == "false" {
						rootName = "FilePresenceUnknownOnHost"
					}
				}
				root := false
				for _, contract := range program.ViewContracts {
					if contract.Name == rootName {
						for _, field := range contract.Fields {
							root = root || field.Name == group.field
						}
					}
				}
				assertOriginalPrimitiveFields(t, directory, program, rootName, rootName)
				if !root {
					t.Fatal("original complete receiver contract absent")
				}
				label := "value." + group.field
				if group.name == "emit-node" {
					label = "constantValue (field constantValue)"
				}
				if group.name == "emit-node-optional" {
					label = "value?." + group.field
				}
				want := run{stdout: []byte(text)}
				if strings.HasPrefix(variant, "wrong") || variant == "null" || (group.name == "literal-union" || group.name == "graph-circular" || group.name == "watch-version") && (variant == "undefined" || variant == "missing") {
					found := map[string]string{"wrong-number": "number", "wrong-string": "string", "wrong": "boolean", "null": "null", "undefined": "undefined", "missing": "missing"}[variant]
					if group.name == "graph-circular" && variant == "wrong" {
						found = "number"
					}
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: " + label + " matches no member of " + group.declared + "; expected " + group.declared + ", found " + found + "\n")}
				}
				if (group.name == "literal-union" || group.name == "bundle-pending" || group.name == "package-peer" || group.name == "reusable-file" || group.name == "builder-signature" || group.name == "graph-circular" || group.name == "watch-version") && variant == "missing" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: " + label + " is not initialized; expected " + group.declared + ", found missing\n")}
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
				if variant == "undefined" && group.name != "literal-union" && group.name != "graph-circular" && group.name != "watch-version" {
					assertPrimitiveFirstMemberMutant(t, program, group.field, group.name == "node-links" || group.name == "bundle-pending", "undefined\n")
				}
				if group.name == "literal-union" && variant == "number" {
					assertPrimitiveFirstMemberMutant(t, program, group.field, false, "42\n")
				}
				if group.name == "watch-version" && variant == "false" {
					assertPrimitiveFirstMemberMutant(t, program, group.field, false, "false\n")
				}
				if group.name == "graph-circular" && variant == "circularity" {
					assertPrimitiveBooleanFirstMember(t, program, group.field, "circularity\n")
				}
				if variant == "wrong-number" || variant == "wrong" || variant == "wrong-string" {
					if count := skipPrimitiveMemberChecks(program, group.field); count != 1 {
						t.Fatalf("want one primitive check mutation, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || string(got.stdout) != text {
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
	data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(directory, "primitive-manifest.json")))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Commit     string              `json:"upstream_commit"`
		Fields     map[string][]string `json:"fields"`
		PrivateSHA string              `json:"private_sha256"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.PrivateSHA != "" {
		private, err := os.ReadFile(checkedViewFixturePath(filepath.Join(directory, "primitive-private.d.ts")))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(private)) != manifest.PrivateSHA {
			t.Fatal("original private declaration drift")
		}
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

func TestPrimitiveOrdinaryBinding(t *testing.T) {
	file, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-original/ordinary-binding.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("word-built\n42\nundefined\n")}
	if diff := disagreement(want, onNode(t, file)); diff != "" {
		t.Fatal("Node: " + diff)
	}
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("%s: %s", diff, got.stderr)
		}
	}
	got, binary := nativelyUncached(t, program)
	if diff := disagreement(want, got); diff != "" {
		t.Fatalf("%s: %s", diff, got.stderr)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func assertPrimitiveFirstMemberMutant(t *testing.T, program *ir.Program, field string, numeric bool, expected string) {
	t.Helper()
	var replacement ir.Expression
	text := "firstfirst\n"
	if numeric {
		replacement = ir.Box{Value: ir.NumberConstant{Value: 0}}
		text = "0\n"
	} else {
		index := len(program.Strings)
		program.Strings = append(program.Strings, "first")
		replacement = ir.Box{Value: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}}
	}
	if count := dropLane4NamedHelperView(program, field, replacement); count != 1 {
		t.Fatalf("want one first-member substitution, got %d", count)
	}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != text {
			t.Fatalf("first-member mutant must execute valid code: stdout %q stderr %q", got.stdout, got.stderr)
		}
		if disagreement(run{stdout: []byte(expected)}, got) == "" {
			t.Fatal("untested first member escaped Node control")
		}
		t.Log("first-member substitution caught by Node control")
	}
}

func TestPrimitiveOrdinaryProperty(t *testing.T) {
	file, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-original/ordinary-property.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("word-built\n42\n")}
	if diff := disagreement(want, onNode(t, file)); diff != "" {
		t.Fatal("Node: " + diff)
	}
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("%s: %s", diff, got.stderr)
		}
	}
	got, binary := nativelyUncached(t, program)
	if diff := disagreement(want, got); diff != "" {
		t.Fatalf("%s: %s", diff, got.stderr)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

// Keep the actual field read and physical storage validation. Drop only declared
// primitive kind/member restrictions; the original wrong value must run on.
func skipPrimitiveMemberChecks(program *ir.Program, field string) int {
	count := 0
	rewrite := func(value any) any {
		property, ok := value.(ir.Property)
		if !ok || property.Name != field || !property.Nullish || property.Of != ir.Union || property.View == "" {
			return value
		}
		property.ViewContract = 0
		property.ViewAllowed = nil
		property.NullishKinds = (1 << ir.String) | (1 << ir.Number) | (1 << ir.Boolean)
		count++
		return property
	}
	for i := range program.Functions {
		program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
	}
	return count
}

func assertPrimitiveBooleanFirstMember(t *testing.T, program *ir.Program, field, expected string) {
	t.Helper()
	if count := dropLane4NamedHelperView(program, field, ir.Box{Value: ir.BooleanConstant{Value: true}}); count != 1 {
		t.Fatalf("want one boolean-first mutation, got %d", count)
	}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != "true\n" {
			t.Fatalf("boolean-first mutant must execute valid code: stdout %q stderr %q", got.stdout, got.stderr)
		}
		if disagreement(run{stdout: []byte(expected)}, got) == "" {
			t.Fatal("boolean-first mutant escaped Node control")
		}
		t.Log("boolean-first member substitution caught by Node control")
	}
}

func TestCheckedViewOriginalBuilderSignatureReceiverGap(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, variant := range []string{"string", "false", "undefined", "wrong", "wrong-number", "null", "missing"} {
		t.Run(variant, func(t *testing.T) {
			input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-original/builder-signature-" + variant + ".a"))
			if err != nil {
				t.Fatal(err)
			}
			bound := strings.Replace(string(input), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/builder.d.ts"))), 1)
			file := filepath.Join(t.TempDir(), "primitive.a")
			if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			text := map[string]string{"string": "word-built", "false": "false", "undefined": "undefined", "wrong": "true", "wrong-number": "42", "null": "null", "missing": "undefined"}[variant] + "\n"
			if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			_, err = lowered(t, file)
			if err == nil || !strings.Contains(err.Error(), "field child with unsupported compound intersection payload contract") {
				t.Fatalf("original intersection receiver guard: %v", err)
			}
			t.Log("original primitive selector component exists; intersection receiver remains lane 7")
		})
	}
}
