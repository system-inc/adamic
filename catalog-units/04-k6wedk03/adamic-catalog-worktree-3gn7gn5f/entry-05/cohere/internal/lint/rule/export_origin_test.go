package rule_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// exportOriginFiles is a program whose main.ts reaches React's hooks every way the lowering follows: a
// renamed re-export behind an `export *`, an `export *` of React itself, a const alias in another file, a cycle of `export *`s, an ambient
// module's re-export, a namespace import, and a default import. Every file but main.ts also holds a
// function whose body names React, so a test can rewrite those bodies and see whether any answer moves.
var exportOriginFiles = map[string]string{
	"tsconfig.json":                   `{"compilerOptions": {"strict": true, "target": "ES2022", "lib": ["ES2022"], "types": [], "module": "esnext", "moduleResolution": "bundler"}, "include": ["*.ts"]}`,
	"node_modules/react/package.json": `{"name": "react", "types": "index.d.ts"}`,
	"node_modules/react/index.d.ts": "export declare function useState<T>(initial: T): [T, (next: T) => void];\n" +
		"export declare function useMemo<T>(factory: () => T, dependencies: unknown[]): T;\n" +
		"declare const React: { useState: typeof useState; useMemo: typeof useMemo };\n" +
		"export default React;\n",
	"barrel.ts": "import * as React from 'react';\nexport { useState as useLocalState } from 'react';\n" +
		"export function barrelHelper() { return 1; }\nexport const barrelValue = React;\n",
	"star.ts":  "export * from './barrel';\nexport function starHelper() { return 2; }\n",
	"alias.ts": "import * as React from 'react';\nexport const useAliased = React.useMemo;\nexport function aliasHelper() { return 3; }\n",
	"cycle_a.ts": "export * from './cycle_b';\nexport { useState as useCycled } from 'react';\n" +
		"export function cycleHelper() { return 4; }\n",
	"cycle_b.ts":    "export * from './cycle_a';\n",
	"star_react.ts": "export * from 'react';\nexport function starReactHelper() { return 6; }\n",
	"hooks.d.ts":    "declare module 'hooks' {\n  export { useMemo as useHook } from 'react';\n}\n",
	"body.ts": "import * as React from 'react';\n" +
		"export function make() { return 5; }\nexport const viaBody = make();\n",
	"main.ts": "import * as React from 'react';\nimport ReactDefault from 'react';\nimport { useMemo as memo } from 'react';\n" +
		"import { useLocalState } from './star';\nimport { useAliased } from './alias';\n" +
		"import { useMemo as starMemo } from './star_react';\nimport { useCycled } from './cycle_b';\nimport { useHook } from 'hooks';\nimport { viaBody } from './body';\n" +
		"const localThing = (value: number) => value;\n" +
		"starMemo(() => 1, []);\nuseLocalState(1);\nuseAliased(() => 1, []);\nuseCycled(1);\nuseHook(() => 1, []);\n" +
		"React.useState(1);\nReact['useMemo'](() => 1, []);\nReactDefault.useMemo(() => 1, []);\n" +
		"memo(() => 1, []);\nReact;\nReactDefault;\nviaBody;\nlocalThing(1);\nparseInt('1');\n",
}

// exportOriginAnswers builds files and returns what ExportNameIn answers for the callee of each
// expression statement in main.ts, keyed by the callee's text, and what ImportBindingOf answers for each
// identifier callee.
func exportOriginAnswers(t *testing.T, files map[string]string) (exports map[string]string, bindings map[string]rule.ImportBinding) {
	t.Helper()
	directory := t.TempDir()
	for name, contents := range files {
		path := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(directory, "tsconfig.json"), CurrentDirectory: directory, SingleThreaded: true})
	if err != nil {
		t.Fatal(err)
	}
	var main *ast.SourceFile
	for _, sourceFile := range graph.ProjectFiles() {
		if filepath.Base(sourceFile.FileName().AsString()) == "main.ts" {
			main = sourceFile
		}
	}
	if main == nil {
		t.Fatal("the probe program does not hold main.ts")
	}
	checker, release := graph.CheckerForFile(context.Background(), main)
	defer release()
	exports = map[string]string{}
	bindings = map[string]rule.ImportBinding{}
	for _, statement := range main.Statements.Nodes {
		if statement.Kind != ast.KindExpressionStatement {
			continue
		}
		callee := statement.AsExpressionStatement().Expression
		if callee.Kind == ast.KindCallExpression {
			callee = callee.AsCallExpression().Expression
		}
		text := strings.TrimSpace(main.Text()[callee.Pos():callee.End()])
		exports[text] = rule.ExportNameIn(checker, callee, "react")
		if callee.Kind == ast.KindIdentifier {
			bindings[text] = rule.ImportBindingOf(callee, checker.GetSymbolAtLocation(callee))
		}
	}
	return exports, bindings
}

// ExportNameIn follows each way a callee can come from React to the export it names, and answers nothing
// for what does not come from React.
func TestExportNameInFollowsReExportsToTheExport(t *testing.T) {
	t.Parallel()
	exports, _ := exportOriginAnswers(t, exportOriginFiles)
	want := map[string]string{
		"useLocalState":        "useState",
		"useAliased":           "useMemo",
		"useCycled":            "useState",
		"starMemo":             "useMemo",
		"useHook":              "useMemo",
		"React.useState":       "useState",
		"React['useMemo']":     "useMemo",
		"ReactDefault.useMemo": "useMemo",
		"React":                "*",
		"memo":                 "useMemo",
		"ReactDefault":         "default",
		"viaBody":              "",
		"localThing":           "",
		"parseInt":             "",
	}
	for callee, export := range want {
		got, found := exports[callee]
		if !found {
			t.Errorf("main.ts has no callee %q", callee)
			continue
		}
		if got != export {
			t.Errorf("ExportNameIn(%s) = %q, want %q", callee, got, export)
		}
	}
}

// Every answer comes from the files' shapes: rewriting the body of every function outside main.ts, to one
// that names a different React export, moves none of them. Editing export syntax does move one, so the
// comparison is not blind.
func TestExportNameInReadsOnlyTheShapesOfOtherFiles(t *testing.T) {
	t.Parallel()
	before, _ := exportOriginAnswers(t, exportOriginFiles)

	rewritten := map[string]string{}
	bodies := 0
	for name, contents := range exportOriginFiles {
		if name != "main.ts" {
			for _, body := range []string{"{ return 1; }", "{ return 2; }", "{ return 3; }", "{ return 4; }", "{ return 5; }", "{ return 6; }"} {
				if strings.Contains(contents, body) {
					contents = strings.Replace(contents, body, "{ const inner = React.useState; return inner; }", 1)
					bodies++
				}
			}
			if strings.Contains(contents, "React.useState; return inner") && !strings.Contains(contents, "import * as React") {
				contents = "import * as React from 'react';\n" + contents
			}
		}
		rewritten[name] = contents
	}
	if bodies != 6 {
		t.Fatalf("rewrote %d function bodies, want all 6", bodies)
	}
	after, _ := exportOriginAnswers(t, rewritten)
	for callee, export := range before {
		if after[callee] != export {
			t.Errorf("rewriting function bodies in other files moved ExportNameIn(%s) from %q to %q", callee, export, after[callee])
		}
	}

	reshaped := map[string]string{}
	for name, contents := range exportOriginFiles {
		reshaped[name] = contents
	}
	reshaped["barrel.ts"] = strings.Replace(reshaped["barrel.ts"], "useState as useLocalState", "useMemo as useLocalState", 1)
	moved, _ := exportOriginAnswers(t, reshaped)
	if moved["useLocalState"] != "useMemo" {
		t.Errorf("re-exporting useMemo as useLocalState left ExportNameIn(useLocalState) at %q, want useMemo", moved["useLocalState"])
	}
}

// ImportBindingOf says how each name is bound in main.ts: imported and from where, declared there, or a
// global.
func TestImportBindingOfReadsTheFilesOwnBinding(t *testing.T) {
	t.Parallel()
	_, bindings := exportOriginAnswers(t, exportOriginFiles)
	want := map[string]rule.ImportBinding{
		"useLocalState": {Kind: rule.ImportBindingSpecifier, Source: "./star", Imported: "useLocalState"},
		"useCycled":     {Kind: rule.ImportBindingSpecifier, Source: "./cycle_b", Imported: "useCycled"},
		"React":         {Kind: rule.ImportBindingNamespace, Source: "react"},
		"ReactDefault":  {Kind: rule.ImportBindingDefault, Source: "react"},
		"memo":          {Kind: rule.ImportBindingSpecifier, Source: "react", Imported: "useMemo"},
		"localThing":    {Kind: rule.ImportBindingModuleLocal},
		"parseInt":      {Kind: rule.ImportBindingGlobal},
	}
	for name, binding := range want {
		if bindings[name] != binding {
			t.Errorf("ImportBindingOf(%s) = %+v, want %+v", name, bindings[name], binding)
		}
	}
}
