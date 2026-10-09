// Package guard holds the checks that read cohere's own source as a whole: the rule packages' contracts
// (creating getters, TypeReach, program and option reads, leaf imports) and gofmt over the module. It has
// no code of its own. They read every rule file, so an edit to one reruns them, and they live apart from
// the code they used to share a package with so that the edit reruns them alone: in
// internal/release/dispatch, every lint edit reran dispatch's Swift and build tests too (#nxgt2ca).
package guard

import (
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// creatingGetters are the compiler's getters that write: each creates the symbol table it is asked for
// when it is missing, and stores it on the node or symbol it was handed. A rule runs on one worker
// while a checker on another resolves names across files through those same tables, so a rule that
// calls one writes shared state that another goroutine reads, a data race (found with -race by
// @system_cohere_cache, 2026-10-03: nine races in one cold api run, all from GetLocals in no-shadow
// and no-redeclare). A rule reads the table without creating it: node.Locals(), symbol.Members,
// symbol.Exports, and a nil table ranges as empty.
var creatingGetters = map[string]bool{"GetLocals": true, "GetMembers": true, "GetExports": true, "GetSymbolTable": true}

// No rule code calls a compiler getter that creates what it returns.
func TestRulesCallNoCompilerGetterThatWrites(t *testing.T) {
	t.Parallel()

	for _, call := range creatingGetterCalls(t, guardOverlay(t)) {
		t.Errorf("%s, which creates the table when it is missing and so writes to a node or symbol other "+
			"workers read; read it without creating it (node.Locals(), symbol.Members, symbol.Exports)", call)
	}
}

// creatingGetterCalls lists every call to a creating getter in the rule packages, read with overlay in
// place of the files it names. go/packages reads files itself, so `go test -overlay` never reaches
// this scan; the guard hands in the overlay GOFLAGS names (guardOverlay), and a probe hands in its own.
func creatingGetterCalls(t *testing.T, overlay map[string][]byte) []string {
	t.Helper()
	loaded := lintPackages(t, overlay)

	found := map[string]bool{}
	calls := []string{}
	for _, pkg := range loaded {
		if len(pkg.Errors) > 0 {
			t.Fatalf("loading %s: %v", pkg.PkgPath, pkg.Errors[0])
		}
		for _, file := range pkg.Syntax {
			path := pkg.Fset.Position(file.Pos()).Filename
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			ast.Inspect(file, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				function, ok := pkg.TypesInfo.Uses[selector.Sel].(*types.Func)
				if !ok || function.Pkg() == nil || !strings.Contains(function.Pkg().Path(), "TypeScript/tsc") {
					return true
				}
				if creatingGetters[function.Name()] {
					position := pkg.Fset.Position(selector.Pos())
					calls = append(calls, filepath.Base(filepath.Dir(position.Filename))+"/"+filepath.Base(position.Filename)+
						":"+strconv.Itoa(position.Line)+" calls "+function.Name())
				}
				found[function.Pkg().Path()] = true
				return true
			})
		}
	}
	if len(found) == 0 {
		t.Fatal("the scan saw no compiler function at all, so it proved nothing")
	}
	sort.Strings(calls)
	return calls
}

// The scan sees a creating getter where one is called: planted into no-shadow, it is named.
func TestTheCreatingGetterScanSeesAPlantedCall(t *testing.T) {
	t.Parallel()

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "internal/lint/rules/typescript/no_shadow.go")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	planted := strings.Replace(string(original), "range node.Locals()", "range ast.GetLocals(node)", 1)
	if planted == string(original) {
		t.Fatal("no-shadow no longer ranges over node.Locals(), so the plant missed; update the probe")
	}
	calls := creatingGetterCalls(t, map[string][]byte{path: []byte(planted)})
	if len(calls) != 1 || !strings.Contains(calls[0], "no_shadow.go") || !strings.HasSuffix(calls[0], "GetLocals") {
		t.Fatalf("a GetLocals call planted in no-shadow was not seen: %v", calls)
	}
}
