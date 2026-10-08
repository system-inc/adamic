package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The integrated tuple selectors now certify these original helper reads.
// Each original member is retained, including both disjoint tuple arities.
func TestCheckedViewOriginalEmitTupleHelpers(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, variant := range []string{"", "-empty", "-tuple", "-wrong"} {
		t.Run(variant, func(t *testing.T) {
			input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/primitive-original/emit-tuple-frontier" + variant + ".a")
			if err != nil {
				t.Fatal(err)
			}
			bound := strings.Replace(string(input), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/builder.d.ts"))), 1)
			file := filepath.Join(t.TempDir(), "tuple.a")
			if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			text := map[string]string{"": "string", "-empty": "object", "-tuple": "object", "-wrong": "boolean"}[variant] + "\n"
			if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(text)}
			if variant == "-wrong" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value[1] matches no member of [] | EmitSignature; expected [] | EmitSignature, found boolean\n")}
			}
			for _, got := range []run{onJavaScriptBackend(t, program), releasedUncached(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("%s stdout %q stderr %q exit %d", diff, got.stdout, got.stderr, got.exitCode)
				}
			}
			got, binary := nativelyUncached(t, program)
			if diff := disagreement(want, got); diff != "" {
				t.Fatalf("sanitized: %s stderr %q", diff, got.stderr)
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			if variant == "-wrong" {
				var original ir.Property
				if changed := changeTupleOriginalRead(program, func(p ir.Property) bool { return p.View == "value[1]" }, func(p ir.Property) ir.Property { original = p; return p }); changed != 1 {
					t.Fatalf("wanted one union selector, got %d", changed)
				}
				contract := &program.ViewContracts[original.ViewContract-1]
				members := append([]ir.ViewContractID(nil), contract.Members...)
				program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Boolean, Name: "boolean"})
				replacement := ir.ViewContractID(len(program.ViewContracts))
				changed := false
				for index, id := range members {
					if program.ViewContracts[id-1].Kind == ir.ViewScalar && program.ViewContracts[id-1].Of == ir.String {
						members[index] = replacement
						changed = true
					}
				}
				if !changed {
					t.Fatal("no scalar member changed")
				}
				program.ViewContracts[original.ViewContract-1].Members = members
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || string(got.stdout) != text {
						t.Fatalf("membership mutant must run: stdout %q stderr %q exit %d", got.stdout, got.stderr, got.exitCode)
					}
					t.Log("wrong-member selection caught by named refusal pin")
				}
			}
			if variant == "-empty" {
				index := len(program.Strings)
				program.Strings = append(program.Strings, "first")
				replacement := ir.Box{Value: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}}
				if changed := dropLane4NamedHelperView(program, "1", replacement); changed != 1 {
					t.Fatalf("first-member mutant changed %d reads", changed)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || string(got.stdout) != "string\n" {
						t.Fatalf("first-member mutant must run: stdout %q stderr %q exit %d", got.stdout, got.stderr, got.exitCode)
					}
					t.Log("untested first member caught by Node empty-tuple control")
				}
			}
		})
	}
}

func TestCheckedViewOriginalEmitTupleValues(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, variant := range []string{"string", "empty", "single", "nested", "arity"} {
		t.Run(variant, func(t *testing.T) {
			input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/primitive-original/emit-tuple-values-" + variant + ".a")
			if err != nil {
				t.Fatal(err)
			}
			bound := strings.Replace(string(input), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/builder.d.ts"))), 1)
			file := filepath.Join(t.TempDir(), "tuple-values.a")
			if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			text := map[string]string{"string": "word-built", "empty": "empty", "single": "word-built", "nested": "false", "arity": "first"}[variant] + "\n"
			if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(text)}
			if variant == "nested" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: member[0] is not a string | undefined; expected string | undefined, found boolean\n")}
			}
			if variant == "arity" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value[1] matches no member of [] | EmitSignature; expected [] | EmitSignature, found object\n")}
			}
			for _, got := range []run{onJavaScriptBackend(t, program), releasedUncached(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("%s stdout %q stderr %q exit %d", diff, got.stdout, got.stderr, got.exitCode)
				}
			}
			got, binary := nativelyUncached(t, program)
			if diff := disagreement(want, got); diff != "" {
				t.Fatalf("sanitized: %s stderr %q", diff, got.stderr)
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			if variant == "empty" {
				changed := 0
				mutate := func(node any) any {
					if narrow, ok := node.(ir.Narrow); ok && narrow.Tuple && narrow.To == ir.Object {
						narrow.Tuple = false
						changed++
						return narrow
					}
					return node
				}
				for index := range program.Functions {
					program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, mutate)
				}
				program.Main = mutateReadiness(program.Main, mutate)
				if changed == 0 {
					t.Fatal("no tuple-union narrowing marker removed")
				}
				got := onJavaScriptBackend(t, program)
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "a union value does not match its narrowed type") {
					t.Fatalf("marker mutant must reproduce narrowing failure: stdout %q stderr %q exit %d", got.stdout, got.stderr, got.exitCode)
				}
				t.Log("tuple-union marker omission caught by Node empty-tuple control (JavaScript)")
			}
			if variant == "nested" {
				if changed := changeTupleOriginalRead(program, func(p ir.Property) bool { return p.View == "member[0]" }, func(p ir.Property) ir.Property { p.View = ""; return p }); changed != 1 {
					t.Fatalf("transitive mutant changed %d reads", changed)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || disagreement(want, got) == "" {
						t.Fatalf("transitive omission must execute: stdout %q stderr %q exit %d", got.stdout, got.stderr, got.exitCode)
					}
					t.Log("transitive tuple-member omission caught by refusal pin")
				}
			}
		})
	}
}
