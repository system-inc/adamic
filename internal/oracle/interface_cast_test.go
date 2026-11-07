package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func interfaceFixture(test *testing.T, name string) (*ir.Program, string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, interfaceFixturePath(name)))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	return program, path
}

func interfaceFixturePath(name string) string {
	aPath := "stage3/interface-downcasts/" + name + ".a"
	tsPath := "stage3/interface-downcasts/" + name + ".ts"
	_, aError := os.Stat(filepath.Join(repository, aPath))
	_, tsError := os.Stat(filepath.Join(repository, tsPath))
	if (aError == nil) == (tsError == nil) {
		panic("expected exactly one interface fixture: " + aPath + " or " + tsPath)
	}
	if aError == nil {
		return aPath
	}
	return tsPath
}

// Default tagged admission is compared to Node on successful programs.
func TestInterfaceCastOracle(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"visitor", "wrong-kind"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			program, path := interfaceFixture(test, name)
			truth := onNode(test, path)
			if truth.exitCode != 0 {
				test.Fatalf("Node source: %d %s", truth.exitCode, truth.stderr)
			}
			want := truth
			if name == "visitor" && string(truth.stdout) != "idid\n42\ntexttext\n5\ntrue\nonce2\ncalls 2\n" {
				test.Fatalf("unexpected source observation: %q", truth.stdout)
			}
			js := onJavaScriptBackend(test, program)
			if name == "wrong-kind" {
				if string(truth.stdout) != "casting\notherother\n" {
					test.Fatalf("source: %q", truth.stdout)
				}
				want = run{
					exitCode: 70,
					stdout:   []byte("casting\n"),
					stderr:   []byte("adamic: panic: cast failed: this Node is not a Identifier\n"),
				}
			}
			sanitized, binary := natively(test, program)
			for backend, result := range map[string]run{
				"JavaScript":       js,
				"native sanitized": sanitized,
				"native release":   released(test, program),
			} {
				if difference := disagreement(want, result); difference != "" {
					test.Fatalf(
						"%s: %s; exit %d stdout %q stderr %q",
						backend,
						difference,
						result.exitCode,
						result.stdout,
						result.stderr,
					)
				}
			}
			if name == "visitor" {
				if report := leaks(test, program, binary); report != "" {
					test.Fatal(report)
				}
			}
		})
	}
}

// This legal Node construction fails only when a required field is read through the cast.
func TestInterfaceCastChecksMalformedRead(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/missing-name.ts"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "missing\n" {
		test.Fatalf("Node: %d %q %s", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := run{exitCode: 70, stderr: []byte(
		"adamic: panic: field read failed: identifier(malformed).name is not initialized; expected string, found missing\n",
	)}
	actual, _ := nativelyUncached(test, program)
	for _, result := range []run{actual, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, result); difference != "" {
			test.Fatalf("malformed view read: %s; got %#v", difference, result)
		}
	}
}

// Mutants change only native IR after lowering.
func TestInterfaceCastRuntimeMutants(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"skip tag", "wrong tag", "twice operand"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			fixture := "visitor"
			if name == "skip tag" {
				fixture = "wrong-kind"
			}
			program, path := interfaceFixture(test, fixture)
			want := onNode(test, path)
			if name == "skip tag" {
				want = onJavaScriptBackend(test, program)
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
					return ir.Conditional{
						Condition: ir.Binary{Operator: ir.Equal, Left: cast.Value, Right: ir.Undefined{Of: ir.Object}},
						WhenTrue:  cast,
						WhenNot:   cast,
						Of:        ir.Object,
					}
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
				test.Fatal("mutant changed no cast")
			}
			result, binary := natively(test, program)
			if strings.Contains(string(result.stderr), "Sanitizer") {
				test.Fatalf("semantic comparison must catch mutant, got %s", result.stderr)
			}
			if difference := disagreement(want, result); difference == "" {
				test.Fatal("mutant survived")
			} else {
				test.Logf(
					"caught %s: %s; expected exit %d stdout %q, mutant exit %d stdout %q stderr %q",
					name,
					difference,
					want.exitCode,
					want.stdout,
					result.exitCode,
					result.stdout,
					result.stderr,
				)
			}
			if result.exitCode == 0 {
				if report := leaks(test, program, binary); report != "" {
					test.Fatal(report)
				}
			}
		})
	}
}

// An unused malformed factory does not prevent a valid read in another module. The modules are .ts, where
// the cast stays a checked view: in .a a downcast whose checks no proof erases is refused (#m8dvvmf).
func TestInterfaceCastImportedConstruction(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	sources := map[string]string{
		"nodes.ts": `export interface Node { readonly kind: 'identifier' | 'number'; } export interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; } export function bad(): Node { return { kind: 'identifier' }; }`,
		"main.ts":  `import { type Node, type Identifier } from './nodes.ts'; function cast(node: Node): Identifier { return node as Identifier; } console.log(cast({ kind: 'identifier', name: 'ok' } as Identifier).name);`,
	}
	for name, source := range sources {
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	program, loweredError := lowered(test, filepath.Join(directory, "main.ts"))
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, filepath.Join(directory, "main.ts"))
	actual, _ := nativelyUncached(test, program)
	for _, got := range []run{actual, onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatal(difference)
		}
	}
}

// Numeric and boolean tags use the same default cast check, in .ts, where the cast stays a checked view: in
// .a a downcast whose checks no proof erases is refused (#m8dvvmf).
func TestInterfaceCastScalarTags(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		name, domain, target, argument string
		fails                          bool
	}{
		{"number passes", "4 | 9", "4", "4", false},
		{"number fails", "4 | 9", "4", "9", true},
		{"boolean passes", "boolean", "true", "true", false},
		{"boolean fails", "boolean", "true", "false", true},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), "scalar.ts")
			source := "interface Base { readonly kind: " +
				testCase.domain +
				"; }\n" +
				"interface Member extends Base { readonly kind: " +
				testCase.target +
				"; readonly name: string; }\n" +
				"function make(kind: " +
				testCase.domain +
				"): Base { const raw = { kind, name: 'payload'.repeat(2) }; return raw; }\n" +
				"function member(node: Base): Member { return node as Member; }\n" +
				"console.log('casting'); console.log(member(make(" +
				testCase.argument +
				")).name);\n"
			if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
				test.Fatal(writeError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			truth := onNode(test, path)
			if truth.exitCode != 0 || string(truth.stdout) != "casting\npayloadpayload\n" {
				test.Fatalf("Node source: %d %q %s", truth.exitCode, truth.stdout, truth.stderr)
			}
			want := truth
			if testCase.fails {
				want = run{
					exitCode: 70,
					stdout:   []byte("casting\n"),
					stderr:   []byte("adamic: panic: cast failed: this Base is not a Member\n"),
				}
			}
			sanitized, binary := natively(test, program)
			for backend, result := range map[string]run{
				"native":     sanitized,
				"JavaScript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(want, result); difference != "" {
					test.Fatalf(
						"%s: %s; exit %d stdout %q stderr %q",
						backend,
						difference,
						result.exitCode,
						result.stdout,
						result.stderr,
					)
				}
			}
			if !testCase.fails {
				if report := leaks(test, program, binary); report != "" {
					test.Fatal(report)
				}
			}
			if testCase.fails {
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
					test.Fatal("scalar tag mutant changed no cast")
				}
				mutant, mutantBinary := natively(test, program)
				if difference := disagreement(truth, mutant); difference != "" {
					test.Fatalf("omitting scalar tag must emit valid source behavior: %s", difference)
				}
				if difference := disagreement(want, mutant); difference == "" {
					test.Fatal("scalar tag omission mutant survived")
				} else {
					test.Logf(
						"caught skip %s tag: %s; checked exit %d, mutant exit %d stdout %q",
						testCase.domain,
						difference,
						want.exitCode,
						mutant.exitCode,
						mutant.stdout,
					)
				}
				if report := leaks(test, program, mutantBinary); report != "" {
					test.Fatal(report)
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
