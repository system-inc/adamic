package lower

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// commonJSReasonPrograms is each program under testdata/commonjs_reasons that's refused, the first
// NotYet it lowers to and where, in the program's own file: one per CommonJS reason the meter names
// (#redxes2), and requires that keep node:module.require.
var commonJSReasonPrograms = []struct{ name, where, what string }{
	{"filename", "main.ts:4:35", "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started)"},
	{"dirname", "main.ts:4:36", "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started)"},
	{"crypto", "main.ts:2:1", `node:module.require("crypto"): CommonJS loading with no ES-module answer yet (fix: import it with an import declaration)`},
	{"spread", "main.ts:6:1", "node:module.require of a computed module path called with a spread or a second argument: CommonJS loading with no ES-module answer yet (fix: pass the module path as its one argument)"},
	{"fs", "main.ts:2:1", "node:module.require"},
	{"fs-asserted", "main.ts:4:1", "node:module.require"},
	{"crypto-asserted", "main.ts:3:1", `node:module.require("crypto"): CommonJS loading with no ES-module answer yet (fix: import it with an import declaration)`},
	{"inspector", "main.ts:2:1", "node:module.require"},
}

// commonJSThrowPrograms is each program under testdata/commonjs_reasons whose require lowers as the
// throw Adamic ruled for it (Nov 3), and what its IR holds once for it: the host's MODULE_NOT_FOUND
// for source-map-support, alone and as install's receiver, and for a computed path, the function
// that throws the Error naming it.
var commonJSThrowPrograms = []struct{ name, holds string }{
	{"source-map-support", `Operation:"host_require_missing"`},
	{"source-map-support-install", `Operation:"host_require_missing"`},
	{"source-map-support-asserted", `Operation:"host_require_missing"`},
	{"computed", `Name:"require_javascript"`},
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

func TestNodeCommonJSRequiresLowerAsThrows(t *testing.T) {
	t.Parallel()
	for _, one := range commonJSThrowPrograms {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata/commonjs_reasons", one.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSourceAs(t, string(source), ".ts")
			if err != nil {
				t.Fatal(err)
			}
			if held := strings.Count(fmt.Sprintf("%#v", program), one.holds); held != 1 {
				t.Fatalf("want %s once, got %d", one.holds, held)
			}
		})
	}
}

// The meter's walk judges every node and counts each distinct (where, what) once, so a require
// call and its callee, both refused, are one site only while they carry one reason at one place.
// Each refused program is one site, as it was under the generic reason, and a program whose require
// lowers as a throw is none.
func TestNodeCommonJSSiteCountsOnce(t *testing.T) {
	t.Parallel()
	for _, one := range commonJSReasonPrograms {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			if want, sites := []string{one.where + ": " + one.what}, commonJSSites(t, one.name); !slices.Equal(sites, want) {
				t.Fatalf("want sites %q, got %q", want, sites)
			}
		})
	}
	for _, one := range commonJSThrowPrograms {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			if sites := commonJSSites(t, one.name); len(sites) != 0 {
				t.Fatalf("want no sites, got %q", sites)
			}
		})
	}
}

// commonJSSites is each distinct site the meter's walk names in a program under
// testdata/commonjs_reasons, as where and what.
func commonJSSites(t *testing.T, name string) []string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata/commonjs_reasons", name+".ts"))
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
	return sites
}
