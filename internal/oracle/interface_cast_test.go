package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func interfaceFixture(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, interfaceFixturePath(name)))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program, path
}

func interfaceFixturePath(name string) string {
	extension := ".a"
	if name == "default-staged" || name == "default-boxed-write" || name == "default-read-before-set" || strings.HasPrefix(name, "readiness-") {
		extension = ".ts"
	}
	return "stage3/interface-downcasts/" + name + extension
}

// Default tagged admission is compared to Node on successful programs.
// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
func TestInterfaceCastOracle(t *testing.T) {
	for _, name := range []string{"visitor", "wrong-kind"} {
		// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, name)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node source: %d %s", truth.exitCode, truth.stderr)
			}
			want := truth
			if name == "visitor" && string(truth.stdout) != "idid\n42\ntexttext\n5\ntrue\nonce2\ncalls 2\n" {
				t.Fatalf("unexpected source observation: %q", truth.stdout)
			}
			js := onJavaScriptBackend(t, program)
			if name == "wrong-kind" {
				if string(truth.stdout) != "casting\notherother\n" {
					t.Fatalf("source: %q", truth.stdout)
				}
				want = run{exitCode: 70, stdout: []byte("casting\n"), stderr: []byte("adamic: panic: cast failed: this Node is not a Identifier\n")}
			}
			sanitized, binary := natively(t, program)
			for backend, result := range map[string]run{"JavaScript": js, "native sanitized": sanitized, "native release": released(t, program)} {
				if difference := disagreement(want, result); difference != "" {
					t.Fatalf("%s: %s; exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
				}
			}
			if name == "visitor" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// This legal Node construction fails only when a required field is read through the cast.
// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
func TestInterfaceCastChecksMalformedRead(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/missing-name.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "missing\n" {
		t.Fatalf("Node: %d %q %s", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: identifier(malformed).name is not initialized; expected string, found missing\n")}
	actual, _ := nativelyUncached(t, program)
	for _, result := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Fatalf("malformed view read: %s; got %#v", difference, result)
		}
	}
}

// Mutants change only native IR after lowering.
// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
func TestInterfaceCastRuntimeMutants(t *testing.T) {
	for _, name := range []string{"skip tag", "wrong tag", "twice operand"} {
		// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
		t.Run(name, func(t *testing.T) {
			fixture := "visitor"
			if name == "skip tag" {
				fixture = "wrong-kind"
			}
			program, path := interfaceFixture(t, fixture)
			want := onNode(t, path)
			if name == "skip tag" {
				want = onJavaScriptBackend(t, program)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				cast, ok := value.(ir.CheckedCast)
				if !ok || changed {
					return value
				}
				if name == "twice operand" {
					if _, call := cast.Value.(ir.Call); !call {
						return value
					}
					changed = true
					return ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: cast.Value, Right: ir.Undefined{Of: ir.Object}}, WhenTrue: cast, WhenNot: cast, Of: ir.Object}
				}
				changed = true
				if name == "skip tag" {
					return cast.Value
				}
				cast.Allowed = []ir.Expression{ir.StringConstant{Index: len(program.Strings)}}
				program.Strings = append(program.Strings, "wrong")
				return cast
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				t.Fatal("mutant changed no cast")
			}
			result, binary := natively(t, program)
			if strings.Contains(string(result.stderr), "Sanitizer") {
				t.Fatalf("semantic comparison must catch mutant, got %s", result.stderr)
			}
			if difference := disagreement(want, result); difference == "" {
				t.Fatal("mutant survived")
			} else {
				t.Logf("caught %s: %s; expected exit %d stdout %q, mutant exit %d stdout %q stderr %q", name, difference, want.exitCode, want.stdout, result.exitCode, result.stdout, result.stderr)
			}
			if result.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// An unused malformed factory does not prevent a valid read in another module.
// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
func TestInterfaceCastImportedConstruction(t *testing.T) {
	directory := t.TempDir()
	sources := map[string]string{
		"nodes.a": `export interface Node { readonly kind: 'identifier' | 'number'; } export interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; } export function bad(): Node { return { kind: 'identifier' }; }`,
		"main.a":  `import { type Node, type Identifier } from './nodes.a'; function cast(node: Node): Identifier { return node as Identifier; } console.log(cast({ kind: 'identifier', name: 'ok' } as Identifier).name);`,
	}
	for name, source := range sources {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := lowered(t, filepath.Join(directory, "main.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, filepath.Join(directory, "main.a"))
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
}

// Numeric and boolean tags use the same default cast check.
// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
func TestInterfaceCastScalarTags(t *testing.T) {
	for _, test := range []struct {
		name, domain, target, argument string
		fails                          bool
	}{
		{"number passes", "4 | 9", "4", "4", false},
		{"number fails", "4 | 9", "4", "9", true},
		{"boolean passes", "boolean", "true", "true", false},
		{"boolean fails", "boolean", "true", "false", true},
	} {
		// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "scalar.a")
			source := "interface Base { readonly kind: " + test.domain + "; }\n" +
				"interface Member extends Base { readonly kind: " + test.target + "; readonly name: string; }\n" +
				"function make(kind: " + test.domain + "): Base { const raw = { kind, name: 'payload'.repeat(2) }; return raw; }\n" +
				"function member(node: Base): Member { return node as Member; }\n" +
				"console.log('casting'); console.log(member(make(" + test.argument + ")).name);\n"
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "casting\npayloadpayload\n" {
				t.Fatalf("Node source: %d %q %s", truth.exitCode, truth.stdout, truth.stderr)
			}
			want := truth
			if test.fails {
				want = run{exitCode: 70, stdout: []byte("casting\n"), stderr: []byte("adamic: panic: cast failed: this Base is not a Member\n")}
			}
			sanitized, binary := natively(t, program)
			for backend, result := range map[string]run{"native": sanitized, "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, result); difference != "" {
					t.Fatalf("%s: %s; exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
				}
			}
			if !test.fails {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			if test.fails {
				changed := false
				omit := func(value ir.Expression) ir.Expression {
					if cast, ok := value.(ir.CheckedCast); ok {
						changed = true
						return cast.Value
					}
					return value
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), omit)
				if !changed {
					t.Fatal("scalar tag mutant changed no cast")
				}
				mutant, mutantBinary := natively(t, program)
				if difference := disagreement(truth, mutant); difference != "" {
					t.Fatalf("omitting scalar tag must emit valid source behavior: %s", difference)
				}
				if difference := disagreement(want, mutant); difference == "" {
					t.Fatal("scalar tag omission mutant survived")
				} else {
					t.Logf("caught skip %s tag: %s; checked exit %d, mutant exit %d stdout %q", test.domain, difference, want.exitCode, mutant.exitCode, mutant.stdout)
				}
				if report := leaks(t, program, mutantBinary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// Readiness controls using ! run as TypeScript; their .a twins pin refusal.
func interfaceSource(name string) string {
	extension := ".a"
	switch name {
	case "default-staged", "default-boxed-write", "default-read-before-set", "readiness-identifier", "readiness-identifier-uninitialized", "readiness-number", "readiness-number-uninitialized":
		extension = ".ts"
	}
	return "stage3/interface-downcasts/" + name + extension
}
