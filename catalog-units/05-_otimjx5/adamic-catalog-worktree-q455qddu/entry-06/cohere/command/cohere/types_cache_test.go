package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// A warm run's type diagnostics are a cold run's, through every kind of edit: one inside a body, which
// re-checks only the edited file; one to an export, which must re-check the files importing it; and a type
// error in the edited file itself, each introduced and then removed (#zqsdzbq).
//
// b.ts imports a.ts and five other files import nothing, so a body edit to a.ts leaves the other six
// replayable, and an export edit leaves only the five. Enough of them that re-checking a.ts and b.ts is
// under half the project, where the full check would take over.
func TestTypeDiagnosticsReplayOnlyWhatAnEditCannotReach(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write("CohereSettings.json", `{"rules":{"no-debugger":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	const original = "export function value(): number {\n    return 1;\n}\n"
	write("source/a.ts", original)
	write("source/b.ts", "import { value } from './a';\nexport const b: number = value();\n")
	for index := range 5 {
		write(fmt.Sprintf("source/c%d.ts", index), fmt.Sprintf("export const c%d: number = %d;\n", index, index))
	}

	typeError := regexp.MustCompile(`(?m)^\S+:\d+:\d+ - error TS\d+: .*$`)
	replayed := regexp.MustCompile(`; (\d+) of 7 files' semantic diagnostics replayed from cache`)
	run := func(arguments ...string) string {
		t.Helper()
		output, _ := runCohere(t, binary, root, append([]string{"--no-fix"}, arguments...)...)
		return output
	}
	run()

	for _, step := range []struct {
		name         string
		aTS          string
		wantErrors   int
		mostReplayed int
	}{
		{"a body edit", "export function value(): number {\n    return 2;\n}\n", 0, 6},
		{"an export edit its importer cannot take", "export function value(): string {\n    return 'two';\n}\n", 1, 5},
		{"the export edit undone", original, 0, 5},
		{"a type error in the edited file", "export function value(): number {\n    return 'one';\n}\n", 1, 6},
		{"the type error removed", original, 0, 6},
	} {
		write("source/a.ts", step.aTS)
		warm, cold := run(), run("--no-cache")
		warmErrors, coldErrors := typeError.FindAllString(warm, -1), typeError.FindAllString(cold, -1)
		if strings.Join(warmErrors, "\n") != strings.Join(coldErrors, "\n") || len(coldErrors) != step.wantErrors {
			t.Fatalf("%s: warm reports %d type errors and cold %d, want %d\n--- warm\n%s\n--- cold\n%s",
				step.name, len(warmErrors), len(coldErrors), step.wantErrors, warm, cold)
		}
		if strings.HasPrefix(warm, "cached: ") {
			t.Fatalf("%s: the run after the edit replayed the run before it:\n%s", step.name, warm)
		}
		match := replayed.FindStringSubmatch(warm)
		if match == nil {
			t.Fatalf("%s: the warm run replayed no file's semantic diagnostics, so it proves nothing about replaying:\n%s", step.name, warm)
		}
		if count, _ := strconv.Atoi(match[1]); count > step.mostReplayed {
			t.Fatalf("%s: %d files replayed, at most %d may: an edit that reaches an importer must re-check it\n%s",
				step.name, count, step.mostReplayed, warm)
		}
	}
}

// A run the run cache does not record, scoped and flagged the way the editor's save is, replays the types
// section too, and after an edit it still reports what a cold run of the same scope does (#hfv0ae3). Before,
// such a run read and rewrote the incremental build info whatever its scope.
func TestTypeDiagnosticsReplayOnAScopedRun(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write("CohereSettings.json", `{"rules":{"no-debugger":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	const original = "export function value(): number {\n    return 1;\n}\n"
	write("source/a.ts", original)
	write("source/b.ts", "import { value } from './a';\nexport const b: number = value();\n")
	for index := range 5 {
		write(fmt.Sprintf("source/c%d.ts", index), fmt.Sprintf("export const c%d: number = %d;\n", index, index))
	}

	typeError := regexp.MustCompile(`(?m)^\S+:\d+:\d+ - error TS\d+: .*$`)
	// The clause counts the files the line reports on, never the whole program the check covered: a scoped run
	// once read "3955 of 1 files" (#rxqptqp).
	replayed := regexp.MustCompile(`; (\d+) of (\d+) files' semantic diagnostics replayed from cache`)
	scoped := []string{"--types", "source/b.ts"}
	runCohere(t, binary, root, scoped...)

	for _, step := range []struct {
		name       string
		aTS        string
		wantErrors int
	}{
		{"a body edit", "export function value(): number {\n    return 2;\n}\n", 0},
		{"an export edit the scoped file cannot take", "export function value(): string {\n    return 'two';\n}\n", 1},
		{"the export edit undone", original, 0},
	} {
		write("source/a.ts", step.aTS)
		warm, _ := runCohere(t, binary, root, scoped...)
		cold, _ := runCohere(t, binary, root, append([]string{"--no-cache"}, scoped...)...)
		warmErrors, coldErrors := typeError.FindAllString(warm, -1), typeError.FindAllString(cold, -1)
		if strings.Join(warmErrors, "\n") != strings.Join(coldErrors, "\n") || len(coldErrors) != step.wantErrors {
			t.Fatalf("%s: warm reports %d type errors and cold %d, want %d\n--- warm\n%s\n--- cold\n%s",
				step.name, len(warmErrors), len(coldErrors), step.wantErrors, warm, cold)
		}
		clause := replayed.FindStringSubmatch(warm)
		if clause != nil {
			replayedCount, _ := strconv.Atoi(clause[1])
			total, _ := strconv.Atoi(clause[2])
			if replayedCount > total || total != 1 {
				t.Errorf("%s: the replay clause counts %d of %d files on a run scoped to one:\n%s", step.name, replayedCount, total, warm)
			}
		}
		// This fixture has no build info, so no file has a recorded declaration signature and every shape is
		// keyed on content (#5txm9gg): a body edit moves a.ts's shape and so b.ts's fingerprint, and b.ts is
		// checked again rather than replayed. That it is checked, and agrees with cold above, is the claim here.
		// Shape replay across a body edit is proven where a real signature is recorded (shape_reuse_test.go).
		if step.name == "a body edit" && clause != nil && clause[1] != "0" {
			t.Fatalf("%s: the scoped run replayed its one file across a body edit to a content-keyed import:\n%s", step.name, warm)
		}
		if step.name == "a body edit" && !strings.Contains(warm, "keyed on content") {
			t.Fatalf("%s: the run does not say its shapes are keyed on content:\n%s", step.name, warm)
		}
	}
}

// A file's semantic diagnostics depend on the compiler options as much as on its text, and no shape
// fingerprint sees an option. So turning on strict, with no source file touched, must not replay the
// diagnostics a lax run recorded: the implicit any it now reports would be replayed as clean (#hfv0ae3).
func TestTypeDiagnosticsDoNotReplayAcrossCompilerOptions(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tsconfig := func(strict bool) string {
		return fmt.Sprintf(`{"compilerOptions":{"strict":%t,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`, strict)
	}
	write("tsconfig.json", tsconfig(false))
	write("CohereSettings.json", `{"rules":{"no-debugger":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	write("source/a.ts", "export function identity(value) {\n    return value;\n}\n")
	for index := range 5 {
		write(fmt.Sprintf("source/c%d.ts", index), fmt.Sprintf("export const c%d: number = %d;\n", index, index))
	}

	typeError := regexp.MustCompile(`(?m)^\S+:\d+:\d+ - error TS\d+: .*$`)
	if lax, _ := runCohere(t, binary, root, "--no-fix"); len(typeError.FindAllString(lax, -1)) != 0 {
		t.Fatalf("the lax run reports a type error, so the fixture proves nothing:\n%s", lax)
	}

	write("tsconfig.json", tsconfig(true))
	warm, _ := runCohere(t, binary, root, "--no-fix")
	cold, _ := runCohere(t, binary, root, "--no-fix", "--no-cache")
	warmErrors, coldErrors := typeError.FindAllString(warm, -1), typeError.FindAllString(cold, -1)
	if len(coldErrors) == 0 {
		t.Fatalf("the strict cold run reports no type error, so the fixture proves nothing:\n%s", cold)
	}
	if strings.Join(warmErrors, "\n") != strings.Join(coldErrors, "\n") {
		t.Fatalf("after turning on strict, warm reports %d type errors and cold %d\n--- warm\n%s\n--- cold\n%s",
			len(warmErrors), len(coldErrors), warm, cold)
	}
}

// A run whose fixer rewrites a file still replays what the rewrite left alone (#891h54d). The rewrite
// rebuilds the graph, and the rebuilt graph used to carry neither shapes nor the findings cache, so the
// types phase full-checked the program and lint walked every rule over every file, though one file had
// changed. Here one file of seven gains a debugger statement the fixer removes: the run after the rewrite
// replays the other files' types and findings, and reports exactly what a run with no cache reports.
func TestARunWhoseFixRewritesAFileStillReplays(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write("CohereSettings.json", `{"rules":{"no-debugger":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	write("source/a.ts", "export function value(): number {\n    return 1;\n}\n")
	for index := range 6 {
		write(fmt.Sprintf("source/c%d.ts", index), fmt.Sprintf("import { value } from './a';\nexport const c%d: number = value() + %d;\n", index, index))
	}
	runCohere(t, binary, root)
	runCohere(t, binary, root)

	write("source/a.ts", "export function value(): number {\n    debugger;\n    return 1;\n}\n")
	warm, _ := runCohere(t, binary, root)
	if !strings.Contains(warm, "graph rebuilt") {
		t.Fatalf("the fixer rewrote nothing, so the rebuild this test is about never ran:\n%s", warm)
	}
	for _, clause := range []string{"files' semantic diagnostics replayed from cache", "files replayed from cache"} {
		if !strings.Contains(warm, clause) {
			t.Errorf("after the rewrite the run replayed nothing (%q missing):\n%s", clause, warm)
		}
	}
	cold, _ := runCohere(t, binary, root, "--no-cache")
	findings := func(output string) string {
		kept := []string{}
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, " - ") {
				kept = append(kept, line)
			}
		}
		return strings.Join(kept, "\n")
	}
	if findings(warm) != findings(cold) {
		t.Errorf("the run after the rewrite found something other than a run with no cache:\n--- warm\n%s\n--- cold\n%s", warm, cold)
	}
}
