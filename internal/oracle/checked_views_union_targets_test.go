package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Original target declarations retain their full graph. Cast admission checks
// every union arm; the array and element reads keep their own lazy contracts.
func TestCheckedViewOriginalUnionTargets(t *testing.T) {
	declarations := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_BRAND_ORIGINAL_DECLS to complete pinned declarations")
	}
	verifyUnionTargetDeclarations(t, declarations)
	for _, group := range []struct {
		name, target, field string
		tags                []string
	}{
		{"function", "FunctionLikeDeclaration", "parameters", []string{"263", "175", "178", "179", "177", "219", "220"}},
		{"class", "ClassDeclaration | ClassExpression", "members", []string{"264", "232"}},
	} {
		for _, variant := range append(group.tags, "wrong-tag", "wrong-array") {
			t.Run(group.name+"/"+variant, func(t *testing.T) {
				input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/union-targets/" + group.name + "-" + variant + ".a")
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
				file := filepath.Join(t.TempDir(), "union.a")
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				node := onNode(t, file)
				text := "1\n"
				if variant == "wrong-array" {
					text = "undefined\n"
				}
				if variant == "wrong-element" {
					text = "1\nundefined\n"
				}
				if diff := disagreement(run{stdout: []byte(text)}, node); diff != "" {
					t.Fatal("Node: " + diff)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				want := run{stdout: []byte(text)}
				if strings.HasPrefix(variant, "wrong-") {
					want = run{exitCode: 70}
					if variant == "wrong-tag" {
						want.stderr = []byte("adamic: panic: cast failed: this Base is not a " + group.target + "\n")
					}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if want.stderr == nil && want.exitCode == 70 {
						if got.exitCode != 70 || !strings.Contains(string(got.stderr), "read failed:") {
							t.Fatalf("expected named read refusal, got %#v", got)
						}
						element := "ParameterDeclaration"
						if group.name == "class" {
							element = "ClassElement"
						}
						pin := "adamic: panic: field read failed: viewed." + group.field + " is not a NodeArray<" + element + ">; expected NodeArray<" + element + ">, found number\n"
						if string(got.stderr) != pin {
							t.Fatalf("refusal drift: %q", got.stderr)
						}
					} else if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
				}
				if want.exitCode == 0 {
					got, binary := nativelyUncached(t, program)
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

func TestCheckedViewOriginalUnionTargetTags(t *testing.T) {
	declarations := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, declarations)
	for _, group := range []struct {
		name, target string
		tags         []string
	}{
		{"function", "FunctionLikeDeclaration", []string{"263", "175", "178", "179", "177", "219", "220", "80"}},
		{"class", "ClassDeclaration | ClassExpression", []string{"264", "232", "80"}},
	} {
		for _, tag := range group.tags {
			t.Run(group.name+"/"+tag, func(t *testing.T) {
				input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/union-targets/" + group.name + "-cast-" + tag + ".a")
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
				file := filepath.Join(t.TempDir(), "cast.a")
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, file)
				if diff := disagreement(run{stdout: []byte("true\n")}, truth); diff != "" {
					t.Fatal(diff)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				want := truth
				if tag == "80" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: this Base is not a " + group.target + "\n")}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: stdout %q stderr %q", diff, got.stdout, got.stderr)
					}
				}
				if tag != "80" {
					got, binary := nativelyUncached(t, program)
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: %s", diff, got.stderr)
					}
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				} else {
					changed := 0
					rewrite := func(node any) any {
						if cast, ok := node.(ir.CheckedCast); ok {
							changed++
							return cast.Value
						}
						return node
					}
					// Reuse the test's generic IR visitor, preserving interface-held containers.
					for i := range program.Functions {
						program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
					}
					program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
					if changed != 1 {
						t.Fatalf("want one tag bypass, got %d", changed)
					}
					for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if diff := disagreement(truth, got); diff != "" {
							t.Fatalf("tag bypass must run valid code: %s stderr %s", diff, got.stderr)
						}
						t.Log("tag bypass mutant caught by exit-70 pin")
					}
				}
			})
		}
	}
}

func rewriteUnionTargetStatements(body []ir.Statement, rewrite func(any) any) []ir.Statement {
	var visit func(reflect.Value) reflect.Value
	visit = func(v reflect.Value) reflect.Value {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			out := reflect.New(v.Type()).Elem()
			mapped := visit(v.Elem())
			out.Set(reflect.ValueOf(rewrite(mapped.Interface())))
			return out
		case reflect.Struct:
			out := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				out.Field(i).Set(visit(v.Field(i)))
			}
			return out
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				out.Index(i).Set(visit(v.Index(i)))
			}
			return out
		default:
			return v
		}
	}
	return visit(reflect.ValueOf(body)).Interface().([]ir.Statement)
}

func verifyUnionTargetDeclarations(t *testing.T, directory string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(directory, "brand-manifest.json"))
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
		t.Fatal("original union declaration provenance changed")
	}
	for file, digest := range manifest.Declarations {
		data, err := os.ReadFile(filepath.Join(directory, file))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			t.Fatal("declaration drift: " + file)
		}
	}
}

// A supported wider string contract must not hide the viewed ArrowFunction's
// impossible name: never obligation. The forged string passes a primitive tag
// check, so only the retained target-family refusal can catch this mutant.
func TestCheckedViewScopedNeverWiderHelper(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/union-targets/never-wider-helper.a")
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/types.d.ts"))), 1)
	file := filepath.Join(t.TempDir(), "never-helper.a")
	if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, file)
	if diff := disagreement(run{stdout: []byte("ordinary\nforged\n")}, truth); diff != "" {
		t.Fatal(diff)
	}
	program, err := lowered(t, file)
	if os.Getenv("ADAMIC_NEVER_SCOPE_MUTANT") == "1" {
		if err != nil {
			t.Fatalf("fallback bypass must compile valid code: %v", err)
		}
		for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
			if diff := disagreement(truth, got); diff != "" {
				t.Fatalf("mutant must run on: %s stderr %s", diff, got.stderr)
			}
			t.Log("scope-guard bypass caught: unsupported viewed name ran on through wider helper")
		}
		return
	}
	suffix := "Adamic 0.1 refuses checked view read of field name with unsupported never contract; prove or implement the never contract before reading this field"
	if err == nil || !strings.HasSuffix(err.Error(), suffix) || !strings.Contains(err.Error(), "never-helper.a:4:") {
		t.Fatalf("wider helper lost original never guard: %v", err)
	}
	t.Log(err)
}
