package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/types/program"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tsast "github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	tsparser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
	"golang.org/x/tools/go/packages"
)

func fixtureAnalyzer(t *testing.T) (*analyzer, *declaration) {
	t.Helper()
	fset := token.NewFileSet()
	texts := []string{
		"package core; type context struct { TypeChecker *int }; var Probe = func(c context) { helper(c) }; func private() {}",
		"package core; func helper(c context) { _ = c.TypeChecker; c.ReportNodeWithFixes(); final() }; func (context) ReportNodeWithFixes() {}; func final() {}",
	}
	files := []*ast.File{}
	for i, text := range texts {
		name := "/fixture/rule.go"
		if i == 1 {
			name = "/fixture/helper.go"
		}
		f, e := parser.ParseFile(fset, name, text, 0)
		if e != nil {
			t.Fatal(e)
		}
		files = append(files, f)
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}
	checked, e := (&types.Config{}).Check(rulesPrefix+"core", fset, files, info)
	if e != nil {
		t.Fatal(e)
	}
	p := &packages.Package{PkgPath: rulesPrefix + "core", Types: checked, TypesInfo: info}
	a := &analyzer{root: "/fixture", fset: fset, declarations: map[types.Object]*declaration{}, byName: map[string]*declaration{}}
	for i, f := range files {
		file := fset.Position(f.Pos()).Filename
		for _, decl := range f.Decls {
			switch n := decl.(type) {
			case *ast.FuncDecl:
				a.add(info.Defs[n.Name], n, p, file)
			case *ast.GenDecl:
				for _, s := range n.Specs {
					if v, ok := s.(*ast.ValueSpec); ok {
						for _, id := range v.Names {
							a.add(info.Defs[id], v, p, file)
						}
					}
				}
			}
		}
		_ = i
	}
	return a, a.byName[rulesPrefix+"core.Probe"]
}
func TestTransitiveSiblingAndMethodDependencies(t *testing.T) {
	t.Parallel()
	a, d := fixtureAnalyzer(t)
	deps, questions, fix, _ := a.trace(d)
	observed := map[string]dependency{}
	for _, dep := range deps {
		observed[dep.Symbol] = dep
	}
	helper := observed[rulesPrefix+"core.helper"]
	final := observed[rulesPrefix+"core.final"]
	if !helper.Direct || final.Direct || !strings.HasSuffix(final.Location, "helper.go:1") {
		t.Fatalf("lost resolved direct/transitive sibling: %#v", deps)
	}
	if len(questions) != 1 || !fix {
		t.Fatalf("hidden checker/fixer not seen: %v, fixer %t", questions, fix)
	}
	if _, ok := observed[rulesPrefix+"core.private"]; ok {
		t.Fatal("unreachable private function was charged")
	}
}
func TestHelperRankingCountsRulesOnce(t *testing.T) {
	t.Parallel()
	d := dependency{Symbol: "shared.h", Package: "shared", Kind: "sibling"}
	rows := rankings([]entry{{Name: "a", Dependencies: []dependency{d, d}}, {Name: "b", Dependencies: []dependency{d}}})
	if len(rows) != 1 || rows[0].Count != 2 || strings.Join(rows[0].Rules, ",") != "a,b" {
		t.Fatalf("bad denominator: %#v", rows)
	}
}
func TestUnknownFrequencyIsNotZero(t *testing.T) {
	t.Parallel()
	if countText(emptyFrequency()) != "unknown" {
		t.Fatal("unmeasured count printed as zero")
	}
	zero := 0
	if countText(frequency{Count: &zero}) != "0" {
		t.Fatal("measured zero was discarded")
	}
}
func TestRegisteredCorpusControl(t *testing.T) {
	t.Parallel()
	// Actual Go rule, independent of the static call extractor.
	source := `debugger;`
	file := parseControl(source)
	r := findRegistration("no-debugger")
	f := emptyFrequency()
	n := 0
	f.Count = &n
	measured(r, file, nil, &f)
	if n != 1 || f.Offered != 1 || f.Listeners != 1 || len(f.Failures) != 0 {
		t.Fatalf("positive control failed: %#v, count %d", f, n)
	}
	clean := parseControl(`const debuggerEnabled = true;`)
	measured(r, clean, nil, &f)
	if n != 1 {
		t.Fatalf("clean control fired: %d", n)
	}
}

func parseControl(source string) *tsast.SourceFile {
	return tsparser.ParseSourceFile(tsast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/control.ts"), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/control.ts"))}, source, core.ScriptKindTS)
}
func findRegistration(name string) rule.Registration {
	for _, r := range rule.Registered() {
		if r.Rule.Name == name {
			return r
		}
	}
	panic("missing control rule")
}

func TestBranchEvidenceRequiresExecutableSelector(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"this.enabled('no-debugger')", "this.selected === \"no-debugger\"", "this.enabled(\n'no-debugger')"} {
		if !hasSelector(text, "no-debugger") {
			t.Fatalf("missed selector: %s", text)
		}
	}
	for _, text := range []string{"const message = 'no-debugger';", "this.enabled('no-debugger-extra')", "case 'no-debugger/unexpectedDebugger':"} {
		if hasSelector(text, "no-debugger") {
			t.Fatalf("message table is not a port: %s", text)
		}
	}
}

func TestAdamicCorpusUsesCompleteProgram(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	check := func(extension string) error {
		source := filepath.Join(directory, "control"+extension)
		if err := os.WriteFile(source, []byte("debugger;"), 0644); err != nil {
			t.Fatal(err)
		}
		config := filepath.Join(directory, "tsconfig.json")
		raw, err := json.Marshal(map[string]any{"sourceExtensions": []string{".a"}, "compilerOptions": map[string]any{"target": "ESNext", "strict": true}, "files": []string{source}})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(config, raw, 0644); err != nil {
			t.Fatal(err)
		}
		graph, err := program.Build(program.Options{CurrentDirectory: directory, ConfigFileName: config, SingleThreaded: true})
		if err == nil && len(graph.ProjectFiles()) != 1 {
			t.Fatal("incomplete .ts control")
		}
		return err
	}
	if err := check(".ts"); err != nil {
		t.Fatalf("TypeScript control failed: %v", err)
	}
	if err := check(".a"); err != nil {
		t.Fatalf("Adamic program failed: %v", err)
	}
	source := filepath.Join(directory, "control.a")
	if err := os.WriteFile(source, []byte("const value: number = 1;"), 0644); err != nil {
		t.Fatal(err)
	}
	registration := findRegistration("@typescript-eslint/unbound-method")
	items := []entry{}
	for _, subject := range rule.Registered() {
		items = append(items, entry{Name: subject.Rule.Name, Repository: emptyFrequency()})
	}
	measureCorpus(directory, directory, "repository", []string{source}, items)
	for _, item := range items {
		if item.Name == registration.Rule.Name && (item.Repository.Count == nil || item.Repository.Offered != 1) {
			t.Fatalf("Adamic program was not measured: %#v", item.Repository)
		}
	}
}

func TestFailedFamilyCannotBeMarkedPassed(t *testing.T) {
	t.Parallel()
	for _, log := range []string{"FAIL\t" + rulesPrefix + "core\t1s", "ok  \t" + rulesPrefix + "tailwind\t1s", "    fake: ok  \t" + rulesPrefix + "core\t1s"} {
		if familyPassed([]byte(log), "core") {
			t.Fatalf("failed or unrun family marked passed: %q", log)
		}
	}
	if !familyPassed([]byte("ok  \t"+rulesPrefix+"core\t1s"), "core") {
		t.Fatal("passing family not recognized")
	}
}

func TestInventoryFileNameBytes(t *testing.T) {
	t.Parallel()
	file := parseControl("const value = 1;")
	subject := rule.Registration{Rule: rule.Rule{Run: func(rule.Context, any) rule.Listeners {
		panic("path probe")
	}}}
	report := emptyFrequency()
	measured(subject, file, nil, &report)
	if len(report.Failures) != 1 || report.Failures[0] != "/control.ts: path probe" {
		t.Fatalf("inventory path bytes changed: %#v", report.Failures)
	}
}
