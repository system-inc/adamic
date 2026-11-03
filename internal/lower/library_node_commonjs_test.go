package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// commonJSReasonPrograms is each program under testdata/commonjs_reasons, the first NotYet it
// lowers to and where, in the program's own file: one per CommonJS reason the meter names
// (#redxes2), and requires that keep node:module.require.
var commonJSReasonPrograms = []struct{ name, where, what string }{
	{"filename", "main.ts:4:35", "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the path in, such as one built from process.cwd())"},
	{"dirname", "main.ts:4:36", "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the path in, such as one built from process.cwd())"},
	{"crypto", "main.ts:2:1", `node:module.require("crypto"): CommonJS loading with no ES-module answer yet (fix: import it with an import declaration)`},
	{"source-map-support", "main.ts:4:1", `node:module.require("source-map-support"): CommonJS loading with no ES-module answer yet (fix: import it with an import declaration)`},
	{"computed", "main.ts:5:1", "node:module.require of a computed module path: CommonJS loading with no ES-module answer yet (fix: import each module it can load with an import declaration)"},
	{"fs", "main.ts:2:1", "node:module.require"},
	{"inspector", "main.ts:2:1", "node:module.require"},
}

func TestNodeCommonJSNamesItsReasons(t *testing.T) {
	t.Parallel()
	for _, one := range commonJSReasonPrograms {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata/commonjs_reasons", one.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSourceAs(t, string(source), ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || filepath.Base(notYet.Where) != one.where || notYet.What != one.what {
				t.Fatalf("want NotYet at %s: %s, got %v", one.where, one.what, err)
			}
		})
	}
}

// The meter's walk judges every node and counts each distinct (where, what) once, so a require
// call and its callee, both refused, are one site only while they carry one reason at one place.
// Each program is one site, as it was under the generic reason.
func TestNodeCommonJSSiteCountsOnce(t *testing.T) {
	t.Parallel()
	for _, one := range commonJSReasonPrograms {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata/commonjs_reasons", one.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			path := writeAgreementSource(t, "main.ts", string(source))
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			var file *ast.SourceFile
			for _, candidate := range program.Files() {
				if program.FileName(candidate) == path {
					file = candidate
				}
			}
			if file == nil {
				t.Fatal("missing main file")
			}
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			sites := []string{}
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				var notYet *NotYet
				if err := l.nodeLibraryRefusal(node); errors.As(err, &notYet) {
					site := filepath.Base(notYet.Where) + ": " + notYet.What
					if !slices.Contains(sites, site) {
						sites = append(sites, site)
					}
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			if want := []string{one.where + ": " + one.what}; !slices.Equal(sites, want) {
				t.Fatalf("want sites %q, got %q", want, sites)
			}
		})
	}
}
