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

func TestCheckedViewOriginalCommandDefault(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	auditBytes, err := os.ReadFile("../../stage3/interface-downcasts/lane4/primitive-original/remaining-candidates.json")
	if err != nil {
		t.Fatal(err)
	}
	var audit struct {
		Fields map[string][]string `json:"fields"`
	}
	if err := json.Unmarshal(auditBytes, &audit); err != nil {
		t.Fatal(err)
	}

	for _, variant := range []string{"string", "number", "boolean", "undefined", "message", "wrong", "nested"} {
		t.Run(variant, func(t *testing.T) {
			input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/primitive-original/command-default-" + variant + ".a")
			if err != nil {
				t.Fatal(err)
			}
			bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/types.d.ts"))), 1)
			file := filepath.Join(t.TempDir(), "remaining.a")
			if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			text := map[string]string{"string": "word-built", "number": "42", "boolean": "true", "undefined": "undefined", "message": "42", "wrong": "object", "nested": "false"}[variant] + "\n"
			if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}

			root := map[string]string{"number": "CommandLineOptionOfNumberType", "boolean": "CommandLineOptionOfBooleanType"}[variant]
			if root == "" {
				root = "CommandLineOptionOfStringType"
			}
			for _, name := range []string{root, "DiagnosticMessage"} {
				full := false
				for _, contract := range program.ViewContracts {
					if contract.Name == name {
						var fields []string
						for _, field := range contract.Fields {
							fields = append(fields, field.Name)
						}
						slices.Sort(fields)
						full = full || slices.Equal(fields, audit.Fields[name]) && len(fields) > 0
					}
				}
				if !full {
					t.Fatal("original field set omitted:", name)
				}
			}
			complete := false
			for _, contract := range program.ViewContracts {
				if contract.Name == "string | number | boolean | DiagnosticMessage | undefined" && contract.Kind == ir.ViewUnion && len(contract.Members) == 6 && contract.Unsupported == "" {
					complete = true
				}
			}
			if !complete {
				t.Fatal("original merged field descriptor absent or incomplete")
			}
			want := run{stdout: []byte(text)}
			if variant == "wrong" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: option.defaultValueDescription matches no member of string | number | boolean | DiagnosticMessage | undefined; expected string | number | boolean | DiagnosticMessage | undefined, found null\n")}
			}
			if variant == "nested" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: member.code is not a number; expected number, found boolean\n")}
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("%s: %#v", diff, got)
				}
			}
			if want.exitCode == 0 {
				got, _ := nativelyUncached(t, program)
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("sanitized: %s %#v", diff, got)
				}
			}
			if variant == "number" {
				assertPrimitiveFirstMemberMutant(t, program, "defaultValueDescription", true, text)
			}
			if variant == "wrong" {
				if changed := changeObjectPrimitiveRead(program, func(read ir.Property) bool { return read.View == "option.defaultValueDescription" }, func(read ir.Property) ir.Property { read.View = ""; return read }); changed != 1 {
					t.Fatal("expected one skipped union selector")
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || disagreement(want, got) == "" {
						t.Fatalf("member mutant must execute: %#v", got)
					}
					t.Log("member-check omission caught by named refusal pin")
				}
			}
			if variant == "nested" {
				if changed := changeObjectPrimitiveRead(program, func(read ir.Property) bool { return read.View == "member.code" }, func(read ir.Property) ir.Property { read.View = ""; return read }); changed != 1 {
					t.Fatal("expected one transitive mutant")
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || disagreement(want, got) == "" {
						t.Fatalf("transitive mutant must execute: %#v", got)
					}
					t.Log("transitive omission caught by refusal pin")
				}
			}
		})
	}
}
