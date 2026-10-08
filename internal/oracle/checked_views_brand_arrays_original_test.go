package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestCheckedViewBrandArrayOriginalPair(t *testing.T) {
	declarations := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_BRAND_ORIGINAL_DECLS to complete pinned declarations")
	}
	data, err := os.ReadFile(filepath.Join(declarations, "brand-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Commit       string            `json:"upstream_commit"`
		Declarations map[string]string `json:"declarations"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Declarations) != 78 {
		t.Fatal("original declaration provenance changed")
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
	for _, variant := range []string{"good", "internal", "undefined", "wrong", "null", "bounds"} {
		t.Run(variant, func(t *testing.T) {
			input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/original/array-" + variant + ".a")
			if err != nil {
				t.Fatal(err)
			}
			bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
			file := filepath.Join(t.TempDir(), "array-"+variant+".a")
			if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			text := map[string]string{"good": "word-built", "internal": "__importAttributes", "undefined": "undefined", "wrong": "42", "null": "null", "bounds": "undefined"}[variant] + "\n"
			if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, file)
			if variant == "null" {
				if err == nil || !strings.HasSuffix(err.Error(), "stage 0 can't lower an array of null yet") {
					t.Fatalf("expected named null-only producer refusal, got %v", err)
				}
				t.Log("null-only producer remains a compile refusal; no runtime pin claimed")
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			complete := false
			for i, contract := range program.ViewContracts {
				if contract.Name != "__String" || contract.Of != ir.String || !contract.Undefined {
					continue
				}
				members, ok := ir.PrimitiveViewMembers(program, ir.ViewContractID(i+1))
				stringMember, undefinedMember := false, false
				for _, member := range members {
					stringMember = stringMember || member.Kind == ir.ViewScalar && member.Of == ir.String && len(member.Allowed) == 0
					undefinedMember = undefinedMember || member.Kind == ir.ViewUndefined
				}
				complete = complete || ok && stringMember && undefinedMember
			}
			if !complete {
				t.Fatal("complete original branded element contract missing")
			}
			want := run{stdout: []byte(text)}
			if variant == "wrong" || variant == "null" {
				found := map[string]string{"wrong": "number", "null": "nullish"}[variant]
				want = run{exitCode: 70, stderr: []byte("adamic: panic: element read failed: values[index] expected __String, found " + found + "\n")}
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
				if count := replaceOriginalBrandArrayRead(program, replacement); count != 1 {
					t.Fatalf("expected one checked element mutation, got %d", count)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || string(got.stdout) != "uncheckedunchecked\n" {
						t.Fatalf("mutant must run valid release code: %#v", got)
					}
					t.Logf("array member-check bypass caught: exit %d stdout %q", got.exitCode, got.stdout)
				}
			}
		})
	}
}

func replaceOriginalBrandArrayRead(program *ir.Program, replacement ir.Expression) int {
	changed := 0
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(v reflect.Value) reflect.Value {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			out := reflect.New(v.Type()).Elem()
			out.Set(rewrite(v.Elem()))
			return out
		case reflect.Struct:
			if read, ok := v.Interface().(ir.ArrayIndex); ok && read.View != "" {
				changed++
				return reflect.ValueOf(replacement)
			}
			out := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				out.Field(i).Set(rewrite(v.Field(i)))
			}
			return out
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				out.Index(i).Set(rewrite(v.Index(i)))
			}
			return out
		default:
			return v
		}
	}
	for i := range program.Functions {
		program.Functions[i].Body = rewrite(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}
	return changed
}
