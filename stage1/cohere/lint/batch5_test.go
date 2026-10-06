package lint

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var batch5Rules = []string{
	"@typescript-eslint/no-dynamic-delete", "@typescript-eslint/no-import-type-side-effects", "@typescript-eslint/no-misused-new", "@typescript-eslint/no-this-alias", "@typescript-eslint/prefer-as-const", "@typescript-eslint/no-confusing-non-null-assertion", "@typescript-eslint/no-extra-non-null-assertion", "@typescript-eslint/no-duplicate-enum-values", "@typescript-eslint/no-explicit-any", "@typescript-eslint/no-useless-empty-export",
}
var batch5Controls = []struct{ file, rule, source, options string }{
	{"no_dynamic_delete.ts", batch5Rules[0], "delete obj[key]; delete obj['x']; delete obj[-1]; delete obj?.[key];", ""},
	{"no_import_type_side_effects.ts", batch5Rules[1], "import { type A as B, type C } from 'module'; import type { D } from 'm'; import { E, type F } from 'n';", ""},
	{"no_misused_new.ts", batch5Rules[2], "interface I { new(): I; constructor(): void; } declare class C { new(): C; }", ""},
	{"no_this_alias.ts", batch5Rules[3], "const self = this; const {x} = this; const safe = this; self += this; const parenthesized = (this);", `{"ReportDestructuring":true,"AllowedNames":["safe"]}`},
	{"prefer_as_const.ts", batch5Rules[4], "const a: 'x' = 'x'; let b = 2 as 2; const c: 1 = 2; class C { p: 3 = 3; }", ""},
	{"no_confusing_non_null_assertion.ts", batch5Rules[5], "a! == b; a! === b; a! = b; a! in b; a! instanceof B; (a!) == b;", ""},
	{"no_extra_non_null_assertion.ts", batch5Rules[6], "a!!!; (b!)!; c!?.d; e!?.(); f!;", ""},
	{"no_duplicate_enum_values.ts", batch5Rules[7], "enum E { A=1, B=1.0, C=1, D='x', F='x', G='x', H=-1, I=-1 }", ""},
	{"no_explicit_any.ts", batch5Rules[8], "let x: any; function f(...a: Array<any>): any { return x; } let y: unknown;", `{"FixToUnknown":true,"IgnoreRestArgs":true}`},
	{"no_useless_empty_export.ts", batch5Rules[9], "import 'm'; export {}; export type {}; export {} from 'n';", ""},
}

func batch5Generated(t *testing.T) []string {
	t.Helper()
	var rows []string
	for i, c := range batch5Controls {
		path := filepath.Join(t.TempDir(), fmt.Sprintf("batch5-%d.ts", i))
		if err := os.WriteFile(path, []byte(c.source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t\t\tfalse\t%s\t", path, c.rule, c.options))
	}
	// Independent spelling, trivia and ancestry combinations beyond imported fixtures.
	for _, prefix := range []string{"", "/* start */\n", "const é = '😀';\r\n"} {
		for _, key := range []string{"name", "((name))", "+1", "-(1)", "`x`", "1n", "true", "null", "(name, 'x')", "0x10"} {
			path := filepath.Join(t.TempDir(), "delete.ts")
			source := prefix + "delete /* operand */ ((obj[" + key + "]));"
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			rows = append(rows, path+"\t"+batch5Rules[0])
		}
		for _, operators := range []string{"=", "+=", "&&=", "??="} {
			for _, target := range []string{"self", "(self)", "(self as any)", "(<any>self)", "self!", "obj.self", "[self]"} {
				path := filepath.Join(t.TempDir(), "this.ts")
				source := prefix + target + operators + "this;"
				if err := os.WriteFile(path, []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
				rows = append(rows, path+"\t"+batch5Rules[3]+"\t\t\tfalse\t{\"ReportDestructuring\":true}")
			}
		}
	}
	for _, source := range []string{
		"import type { type A } from 'm'; import { type 'foo' as F, type as as as } from 'n';",
		"const a /* key */ : /* type */ 0x10 /* end */ = /* value */ 16; const b: '\\u0078' = 'x';",
		"let x: 'a' = 'a'; import {type A, type B} from 'm'; let y:any; z!!!; export {};",
		"a + b! == c; a! /* trivia */ instanceof B; x! != y; x!!==y;",
		"enum E { A=0x10, B=16, C=1e400, D=1e401, S='x', T='\\u0078', U='x' }",
		"function f(cb: (...a: any[]) => any, ...args: Map<any,Array<any>>) {}",
		"x!!!!!!!!!!!!;",
	} {
		path := filepath.Join(t.TempDir(), "composed.ts")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\tall\t\t\tfalse\t{\"IgnoreRestArgs\":true,\"FixToUnknown\":true}")
	}
	t.Logf("batch5 generated: %d cases", len(rows))
	return rows
}
func TestBatch5Controls(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	compare(t, goOracle(t), buildPort(t, directory, true), directory, manifest(t, batch5Generated(t)))
}

// Not parallel: each scratch mutant compiles and executes against its matching Go control.
func TestBatch5Mutants(t *testing.T) {
	oracle := goOracle(t)
	rows := batch5Generated(t)
	for i, c := range batch5Controls {
		t.Run(c.rule, func(t *testing.T) {
			path := manifest(t, []string{rows[i]})
			want := execute(t, "", oracle, "--manifest", path).output
			directory := mutant(t, "const rule = '"+c.rule+"';", "const rule = 'omitted-batch5-rule';", c.file)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("mutant survived on %s", side.name)
				}
				t.Logf("caught on %s: %s", side.name, difference(side.run.output, want))
			}
		})
	}
}

func checkBatch5RecoveryRefusal(t *testing.T, oracle, binary, directory, row string) {
	t.Helper()
	path := manifest(t, []string{strings.TrimSuffix(row, "unsupported-recovery") + "recovery"})
	answer := execute(t, "", oracle, "--manifest", path)
	if !bytes.Contains(answer.output, []byte("unexpectedAny")) {
		t.Fatal("Go recovery finding disappeared")
	}
	t.Logf("EXPLICIT LIMIT: Go recovered: %s", answer.output)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		args []string
	}{{binary, []string{"--manifest", path}}, {"node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path}}} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		command := exec.CommandContext(ctx, side.name, side.args...)
		out, err := os.CreateTemp(t.TempDir(), "refusal-")
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout = out
		var stderr bytes.Buffer
		command.Stderr = &stderr
		err = command.Run()
		timedOut := ctx.Err() == context.DeadlineExceeded
		cancel()
		out.Close()
		if err == nil || (!timedOut && !strings.Contains(stderr.String(), "adamic: panic: parser slice unsupported primary CloseBraceToken")) {
			t.Fatalf("expected parser refusal: %v: %s", err, &stderr)
		}
		t.Logf("parser recovery gap: %s timeout=%t: %v: %s", side.name, timedOut, err, &stderr)
	}
}

func TestBatch5RepairPayloadMutants(t *testing.T) {
	oracle := goOracle(t)
	rows := batch5Generated(t)
	for _, change := range []struct {
		name, file, from, to string
		control              int
	}{
		{"additional automatic edit lost", "batch5_context.ts", "for(const extra of edits.slice(1))", "for(const extra of edits.slice(2))", 1},
		{"second suggestion lost", "batch5_context.ts", "for(const proposed of suggestions)", "for(const proposed of suggestions.slice(0, 1))", 5},
		{"suggestion promoted to automatic edit", "batch5_context.ts", "edit === undefined ? (suggestion === undefined ? '' : 'suggestion') : 'fix'", "edit === undefined ? (suggestion === undefined ? '' : 'fix') : 'fix'", 5},
	} {
		t.Run(change.name, func(t *testing.T) {
			path := manifest(t, []string{rows[change.control]})
			want := execute(t, "", oracle, "--manifest", path).output
			directory := mutant(t, change.from, change.to, change.file)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("mutant survived on %s", side.name)
				}
				t.Logf("caught on %s: %s", side.name, difference(side.run.output, want))
			}
		})
	}
}

func TestBatch5RecoveryGaps(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	for _, file := range []string{"gaps/2_interface_method_body.ts", "gaps/3_type_method_body.ts"} {
		path, err := filepath.Abs(file)
		if err != nil {
			t.Fatal(err)
		}
		checkBatch5RecoveryRefusal(t, oracle, binary, directory, path+"\t"+batch5Rules[8]+"\t\t\tfalse\t\tunsupported-recovery")
	}
}
