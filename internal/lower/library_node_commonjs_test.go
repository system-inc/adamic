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

// commonJsReasonPrograms is each program under testdata/commonjs_reasons that's refused, the first
// NotYet it lowers to and where, in the program's own file: one per CommonJS reason the meter names
// (#redxes2), and requires that keep node:module.require.
var commonJsReasonPrograms = []struct{ name, where, what string }{
	{
		"filename",
		"main.ts:4:35",
		"node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started)",
	},
	{
		"dirname",
		"main.ts:4:36",
		"node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started)",
	},
	{
		"crypto",
		"main.ts:2:1",
		`node:module.require("crypto"): CommonJS loading with no ES-module answer yet (fix: import it with an import declaration)`,
	},
	{
		"spread",
		"main.ts:6:1",
		"node:module.require of a computed module path called with a spread or a second argument: CommonJS loading with no ES-module answer yet (fix: pass the module path as its one argument)",
	},
	{"fs", "main.ts:2:1", "node:module.require"},
	{"fs-asserted", "main.ts:4:1", "node:module.require"},
	{
		"crypto-asserted",
		"main.ts:3:1",
		`node:module.require("crypto"): CommonJS loading with no ES-module answer yet (fix: import it with an import declaration)`,
	},
}

// commonJsThrowPrograms is each program under testdata/commonjs_reasons whose require lowers as the
// throw Adamic ruled for it (Nov 3), and what its IR holds once for it: the host's MODULE_NOT_FOUND
// for source-map-support, alone and as install's receiver, and for a computed path, the function
// that throws the Error naming it.
var commonJsThrowPrograms = []struct{ name, holds string }{
	{"source-map-support", `Operation:"host_require_missing"`},
	{"source-map-support-install", `Operation:"host_require_missing"`},
	{"source-map-support-asserted", `Operation:"host_require_missing"`},
	{"computed", `Name:"require_javascript"`},
	{"inspector", `Operation:"host_inspector_unavailable"`},
	{"inspector-node", `Operation:"host_inspector_unavailable"`},
	{"inspector-promises", `Operation:"host_inspector_unavailable"`},
	{"inspector-wrapped", `Operation:"host_inspector_unavailable"`},
}

func TestNodeCommonJsNamesItsReasons(test *testing.T) {
	test.Parallel()
	for _, one := range commonJsReasonPrograms {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			source, readError := os.ReadFile(filepath.Join("testdata/commonjs_reasons", one.name+".ts"))
			if readError != nil {
				test.Fatal(readError)
			}
			_, lowerSourceAsError := lowerSourceAs(test, string(source), ".ts")
			var notYet *NotYet
			if !errors.As(lowerSourceAsError, &notYet) ||
				filepath.Base(notYet.Where) != one.where ||
				notYet.What != one.what {
				test.Fatalf("want NotYet at %s: %s, got %v", one.where, one.what, lowerSourceAsError)
			}
		})
	}
}

func TestNodeCommonJsRequiresLowerAsThrows(test *testing.T) {
	test.Parallel()
	for _, one := range commonJsThrowPrograms {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			source, readError := os.ReadFile(filepath.Join("testdata/commonjs_reasons", one.name+".ts"))
			if readError != nil {
				test.Fatal(readError)
			}
			program, lowerSourceAsError := lowerSourceAs(test, string(source), ".ts")
			if lowerSourceAsError != nil {
				test.Fatal(lowerSourceAsError)
			}
			if sites := commonJsSites(test, one.name); len(sites) != 0 {
				test.Fatalf("admitted require still counted by the meter: %v", sites)
			}
			if held := strings.Count(fmt.Sprintf("%#v", program), one.holds); held != 1 {
				test.Fatalf("want %s once, got %d", one.holds, held)
			}
		})
	}
}

// The meter's walk judges every node and counts each distinct (where, what) once, so a require
// call and its callee, both refused, are one site only while they carry one reason at one place.
// Each refused program is one site, as it was under the generic reason, and a program whose require
// lowers as a throw is none.
func TestNodeCommonJsSiteCountsOnce(test *testing.T) {
	test.Parallel()
	for _, one := range commonJsReasonPrograms {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			if want, sites := []string{
				one.where + ": " + one.what,
			}, commonJsSites(test, one.name); !slices.Equal(sites, want) {
				test.Fatalf("want sites %q, got %q", want, sites)
			}
		})
	}
	for _, one := range commonJsThrowPrograms {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			if sites := commonJsSites(test, one.name); len(sites) != 0 {
				test.Fatalf("want no sites, got %q", sites)
			}
		})
	}
}

// commonJsSites is each distinct site the meter's walk names in a program under
// testdata/commonjs_reasons, as where and what.
func commonJsSites(test *testing.T, name string) []string {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join("testdata/commonjs_reasons", name+".ts"))
	if readError != nil {
		test.Fatal(readError)
	}
	path := writeAgreementSource(test, "main.ts", string(source))
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	var file *ast.SourceFile
	for _, candidate := range program.Files() {
		if program.FileName(candidate) == path {
			file = candidate
		}
	}
	if file == nil {
		test.Fatal("missing main file")
	}
	checked, release := program.Checker(context.Background(), file)
	defer release()
	lowering := &lowering{program: program, checker: checked, result: &ir.Program{}}
	sites := []string{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var notYet *NotYet
		if nodeLibraryRefusalError := lowering.nodeLibraryRefusal(node); errors.As(nodeLibraryRefusalError, &notYet) {
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
