package main

import (
	"strings"
	"testing"
)

// optionsProject writes a project whose tsconfig carries the compiler options given and whose one file
// carries a `debugger` statement, a lint finding, plus whatever source is added.
func optionsProject(t *testing.T, compilerOptions string, source string) string {
	t.Helper()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json": `{"compilerOptions":{"target":"ES2022","module":"esnext","moduleResolution":"bundler",` +
			`"strict":true,"noEmit":true` + compilerOptions + `},"include":["**/*.ts"]}`,
		"CohereSettings.json": `{"rules":{"no-debugger":"error"}}`,
		"index.ts":            "export function value(): number {\n  debugger;\n  return 1;\n}\n" + source,
	})
	return root
}

// A diagnostic about the compiler options is reported and fails the run, and lint still runs: an option
// TypeScript 7 removed (TS5102, excalidraw's baseUrl), a removed option's value (TS5108), and an option
// TypeScript 7 does not know at all (TS5023, which TypeScript 5's removed options all are), which used to
// refuse the build outright. A source type error beside the same option still stops lint, so the gate is
// the source and not the count. And the control: with neither, lint runs and the run is red on the
// finding alone, so the lint line below is the gate at work and not a rule that always reports.
func TestAnOptionsDiagnosticFailsTheRunWithoutStoppingLint(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)

	for _, shape := range []struct {
		name            string
		compilerOptions string
		code            string
	}{
		{"a removed option", `,"baseUrl":"."`, "TS5102"},
		{"a removed option value", `,"moduleResolution":"node10"`, "TS5108"},
		{"an option TypeScript 7 does not know", `,"importsNotUsedAsValues":"remove"`, "TS5023"},
		{"an option value of the wrong type", `,"strictNullChecks":"yes"`, "TS5024"},
	} {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			output, code := runCohere(t, binary, optionsProject(t, shape.compilerOptions, ""), "--no-fix", "--no-format")
			if code == 0 {
				t.Fatalf("an options diagnostic passed the run:\n%s", output)
			}
			if !strings.Contains(output, shape.code) {
				t.Fatalf("the run did not report %s:\n%s", shape.code, output)
			}
			if strings.Contains(output, "lint did not run") || !strings.Contains(output, "[no-debugger/") {
				t.Fatalf("%s stopped lint:\n%s", shape.code, output)
			}
		})
	}

	t.Run("a source type error still stops lint", func(t *testing.T) {
		t.Parallel()
		root := optionsProject(t, `,"baseUrl":"."`, "export const wrong: number = 'text';\n")
		output, code := runCohere(t, binary, root, "--no-fix", "--no-format")
		if code == 0 || !strings.Contains(output, "TS2322") || !strings.Contains(output, "TS5102") {
			t.Fatalf("the run did not report both diagnostics and fail (exit %d):\n%s", code, output)
		}
		if !strings.Contains(output, "lint did not run (types bailed: 1 type diagnostics") || strings.Contains(output, "[no-debugger/") {
			t.Fatalf("a source type error did not stop lint, or the bail counted the options diagnostic:\n%s", output)
		}
	})

	t.Run("control: no options diagnostic, and lint reports", func(t *testing.T) {
		t.Parallel()
		output, code := runCohere(t, binary, optionsProject(t, "", ""), "--no-fix", "--no-format")
		if code == 0 || !strings.Contains(output, "[no-debugger/") || strings.Contains(output, "error TS") {
			t.Fatalf("the clean-options project did not fail on its lint finding alone (exit %d):\n%s", code, output)
		}
	})
}

// A tsconfig that does not parse, or whose extends cannot be read, still refuses the build: the options
// cohere would check against are not the ones written, which is not something to report beside a run.
func TestAConfigThatDoesNotParseStillRefusesTheBuild(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	for _, shape := range []struct {
		name     string
		tsconfig string
	}{
		{"a JSON syntax error", `{"compilerOptions":{"strict":true,,},"include":["**/*.ts"]}`},
		{"an extends that is not there", `{"extends":"./missing.json","include":["**/*.ts"]}`},
	} {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeTree(t, root, map[string]string{
				"tsconfig.json":       shape.tsconfig,
				"CohereSettings.json": `{"rules":{"no-debugger":"error"}}`,
				"index.ts":            "export const value = 1;\n",
			})
			output, code := runCohere(t, binary, root, "--no-fix", "--no-format")
			if code == 0 || !strings.Contains(output, "building the type graph") {
				t.Fatalf("a tsconfig that does not parse built a graph (exit %d):\n%s", code, output)
			}
		})
	}
}
