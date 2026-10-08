package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewPrimitiveArrayPairs(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, group := range []string{"array-zero", "array-one", "array-two", "diagnostic-arguments"} {
		for _, variant := range []string{"number", "string", "wrong", "missing"} {
			t.Run(group+"/"+variant, func(t *testing.T) {
				input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-original/" + group + "-" + variant + ".a"))
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.ReplaceAll(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.Join(directory, "compiler/types.d.ts")))
				file := filepath.Join(t.TempDir(), "primitive.a")
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				text := map[string]string{"number": "42\n", "string": "word-built\n", "wrong": "true\n", "missing": "undefined\n"}[variant]
				if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
					t.Fatal("Node: " + diff)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				label := map[string]string{"array-zero": "value[0]", "array-one": "value[1]", "array-two": "value[2]", "diagnostic-arguments": "value[index]"}[group]
				want := run{stdout: []byte(text)}
				if variant == "wrong" || variant == "missing" {
					found := "boolean"
					if variant == "missing" {
						found = "undefined"
					}
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: " + label + " matches no member of string | number; expected string | number, found " + found + "\n")}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
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
					count := mutatePrimitiveArrayRead(program, func(read ir.ArrayIndex) ir.Expression {
						member := ir.ViewContractID(len(program.ViewContracts) + 1)
						program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Boolean})
						original := program.ViewContracts[read.ViewContract-1]
						original.Members = append(append([]ir.ViewContractID{}, original.Members...), member)
						program.ViewContracts = append(program.ViewContracts, original)
						read.ViewContract = ir.ViewContractID(len(program.ViewContracts))
						return read
					})
					if count != 1 {
						t.Fatalf("want one actual read mutation, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || string(got.stdout) != text {
							t.Fatalf("member mutant must execute original wrong value: %#v", got)
						}
						t.Log("actual-value member-check mutant caught by named refusal")
					}
				}
				if variant == "number" {
					if count := mutatePrimitiveArrayRead(program, func(read ir.ArrayIndex) ir.Expression { return ir.Box{Value: ir.NumberConstant{Value: 0}} }); count != 1 {
						t.Fatalf("want one first-member mutation, got %d", count)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 0 || string(got.stdout) != "0\n" {
							t.Fatalf("first-member mutant must execute valid release code: %#v", got)
						}
						t.Log("untested first-member mutant caught by Node control")
					}
				}
			})
		}
	}
}

func mutatePrimitiveArrayRead(program *ir.Program, mutate func(ir.ArrayIndex) ir.Expression) int {
	count := 0
	rewrite := func(value any) any {
		read, ok := value.(ir.ArrayIndex)
		if !ok || read.Element != ir.Union || read.View == "" {
			return value
		}
		count++
		return mutate(read)
	}
	for i := range program.Functions {
		program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
	}
	return count
}

func TestCheckedViewPrimitiveArraySafety(t *testing.T) {
	for _, sample := range []struct{ name, output, field, declared, found string }{
		{"ordinary", "42\nword-built\n", "", "", ""},
		{"alias", "42\n7\n", "", "", ""},
		{"index", "word-built\nnew-built\n1\n", "", "", ""},
		{"typeof", "string\n", "", "", ""},
		{"sparse", "undefined\n", "viewed.child[0]", "string | number", "undefined"},
		{"negative", "undefined\n", "viewed.child[-1]", "string | number", "undefined"},
		{"fraction", "undefined\n", "viewed.child[0.5]", "string | number", "undefined"},
		{"optional", "undefined\n", "", "", ""},
		{"finite-good", "0\n", "", "", ""},
		{"finite-wrong", "42\n", "viewed.child[0]", "string | 0", "number"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			program, file := interfaceFixture(t, "lane4/primitive-original/array-safety-"+sample.name)
			truth := onNode(t, file)
			if truth.exitCode != 0 || string(truth.stdout) != sample.output {
				t.Fatalf("Node: %#v", truth)
			}
			want := truth
			if sample.found != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: " + sample.field + " matches no member of " + sample.declared + "; expected " + sample.declared + ", found " + sample.found + "\n")}
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("%s: %#v", diff, got)
				}
			}
			if sample.found == "" {
				got, binary := nativelyUncached(t, program)
				if diff := disagreement(want, got); diff != "" {
					t.Fatal(diff)
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
