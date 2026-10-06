package lint

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var batch6Rules = []string{
	"@next/next/no-assign-module-variable",
	"@typescript-eslint/default-param-last",
	"@typescript-eslint/no-confusing-non-null-assertion",
	"@typescript-eslint/no-duplicate-enum-values",
	"@typescript-eslint/no-dynamic-delete",
	"@typescript-eslint/no-extra-non-null-assertion",
	"@typescript-eslint/no-misused-new",
	"@typescript-eslint/no-unnecessary-parameter-property-assignment",
	"@typescript-eslint/prefer-as-const",
	"default-case-last",
}

func batch6Generated(t *testing.T) []string {
	t.Helper()
	sources := []string{
		"export let a, module = {}; let {module: renamed} = x; for(let module of xs){}; function f(module: unknown){ let moduleName; }",
		"function f(a = 1, b: number) {} function g(a: number, b = 1) {} declare function h(a = 1, b: number): void; class C { constructor(public a = 1, b: number) {} }",
		"a! = b; a! == b; a! === b; a! in b; a! instanceof b; a + b! == c; a + b! in c; (a + b!) == c; a! != b; a! !== b;",
		"enum E { A = 0x10, B = 16, C = 16, D = 's', E = 's', F = 's', G = -1, H = -1, I = `s`, J = 1n, K = 1n }",
		"delete x[k]; delete x['k']; delete x[-1]; delete x[+1]; delete x?.[k]; delete x[((k))];",
		"x!!!; (x!)?.y; x!?.[k]; x!?.(); x!.y; x?.y!;",
		"interface I { new(): I; constructor(); } class C { new(): C; } class D { new(): D {} } type T = { constructor(); new(): T; };",
		"class C { constructor(public x: string) { this.x = x; (() => { this.x = x; })(); } } class D { constructor(x: string) { this.x = x; } }",
		"let a: 0x10 = 16; let b = 's' as 's'; class C { x: 's' = 's'; } let c: true = true; let d: -1 = -1;",
		"switch(x) { case 1: default: break; case 2: break; } switch(x) { case 1: default: } switch(x) { default: default: case 1: }",
	}
	var rows []string
	for i, source := range sources {
		path := filepath.Join(t.TempDir(), fmt.Sprintf("family-%d.ts", i+1))
		source = "// é 😀\r\n" + source + "\r\n"
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\t"+batch6Rules[i])
	}
	return rows
}

// Not parallel: compiles each executable mutant and runs its complete observations under sanitizers.
func TestBatch6Mutants(t *testing.T) {
	oracle := goOracle(t)
	path := manifest(t, batch6Generated(t))
	want := execute(t, "", oracle, "--manifest", path).output
	for _, rule := range batch6Rules {
		if !bytes.Contains(want, []byte("  "+rule+"  ")) {
			t.Fatalf("positive control absent for %s", rule)
		}
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	compare(t, oracle, buildPort(t, directory, true), directory, path)
	changes := []struct{ name, file, from, to string }{
		{"module identifier suppressed", "no_assign_module_variable.ts", "ctx.node(name).text === 'module'", "ctx.node(name).text === 'other'"},
		{"defaulted parameter suppressed", "default_param_last.ts", "cursor < last", "cursor < 0"},
		{"confusing operator suppressed", "no_confusing_non_null_assertion.ts", "!ctx.text(left).endsWith('!')", "ctx.text(left).endsWith('!')"},
		{"enum anchors on later duplicate", "no_duplicate_enum_values.ts", "                earlier,", "                value,"},
		{"dynamic identifier accepted", "no_dynamic_delete.ts", "node.kind === 'StringLiteral' ||", "node.kind === 'Identifier' || node.kind === 'StringLiteral' ||"},
		{"redundant bang fix empty", "no_extra_non_null_assertion.ts", "ctx.node(index).end - 1", "ctx.node(index).end"},
		{"construct span widened", "no_misused_new.ts", "ctx.start(member) + 3", "ctx.start(member) + 4"},
		{"assignment suggestion drops semicolon", "no_unnecessary_parameter_property_assignment.ts", "finding.editEnd++;", "finding.editEnd += 0;"},
		{"const insertion changes type", "prefer_as_const.ts", "' as const'", "' as never'"},
		{"default position reversed", "default_case_last.ts", "position !== clauses.length - 1", "position === clauses.length - 1"},
		{"second suggestion omitted", "no_confusing_non_null_assertion.ts", "finding.suggestions.push(wrap)", "finding.suggestions.push(first)"},
		{"wrap closing edit wrong", "no_confusing_non_null_assertion.ts", "new ExtraEdit(end, end, ')')", "new ExtraEdit(end, end, ']')"},
	}
	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			directory := mutantFile(t, change.file, change.from, change.to)
			for _, side := range []struct {
				name string
				run  execution
			}{
				{"Node", node(t, directory, path, false)},
				{"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)},
			} {
				diff := difference(side.run.output, want)
				if diff == "" {
					t.Fatalf("mutant survived on %s", side.name)
				}
				t.Logf("caught on %s: %s", side.name, diff)
			}
		})
	}
}

func batch6Selected(name string) bool {
	for _, rule := range batch6Rules {
		if rule == name {
			return true
		}
	}
	return strings.Contains("|no-debugger|no-empty|eqeqeq|no-var|no-duplicate-case|", "|"+name+"|")
}

// The one known parser refusal is pinned byte for byte; no broad parse-error exclusion.
func checkBatch6Refusal(t *testing.T, oracle, binary, directory, row string) {
	path := manifest(t, []string{row})
	answer := execute(t, "", oracle, "--manifest", path)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		args []string
	}{
		{"node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path}},
		{binary, []string{"--manifest", path}},
	} {
		output, err := os.CreateTemp(t.TempDir(), "refusal-")
		if err != nil {
			t.Fatal(err)
		}
		command := exec.Command(side.name, side.args...)
		command.Stdout, command.Stderr = output, output
		runErr := command.Run()
		output.Close()
		data, err := os.ReadFile(output.Name())
		if err != nil {
			t.Fatal(err)
		}
		exit, ok := runErr.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || !bytes.Contains(data, []byte("parser slice expected CloseParenToken, got OpenBraceToken at 33")) {
			t.Fatalf("unexpected refusal: %v %s", runErr, data)
		}
		t.Logf("EXPLICIT LIMIT on %s: %s; Go emits %d bytes", side.name, strings.TrimSpace(string(data)), len(answer.output))
	}
}

// Not parallel: the count mutant must pass ordinary byte parity before the count check kills it.
func TestBatch6CountCheck(t *testing.T) {
	oracle := goOracle(t)
	path := manifest(t, batch6Generated(t))
	ordinary := execute(t, "", oracle, "--manifest", path).output
	count := execute(t, "", oracle, "--manifest", path, "--count").output
	directory := mutantFile(t, "main.ts", "        return linter.findings.length;", "        return linter.findings.length + 1;")
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name            string
		ordinary, count execution
	}{
		{"Node", node(t, directory, path, false), node(t, directory, path, true)},
		{"native", execute(t, "", binary, "--manifest", path), execute(t, "", binary, "--manifest", path, "--count")},
	} {
		if !bytes.Equal(side.ordinary.output, ordinary) {
			t.Fatalf("%s was caught outside count check", side.name)
		}
		if bytes.Equal(side.count.output, count) {
			t.Fatalf("count mutant survived on %s", side.name)
		}
		t.Logf("count check alone caught %s: mutant=%q Go=%q; ordinary output identical", side.name, side.count.output, count)
	}
}
