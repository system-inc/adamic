package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func TestStableImportEqualsRefusal(test *testing.T) {
	test.Parallel()
	stableImportFormRefusal(
		test,
		"import_equals.a",
		"an ESM namespace object used as a value; use qualified live bindings",
		":2:25",
	)
}

func TestStableDynamicImportRefusal(test *testing.T) {
	test.Parallel()
	stableImportFormRefusal(test, "dynamic_import.a", "a dynamic import()", ":2:17")
}

func stableImportFormRefusal(test *testing.T, name, what, position string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join("testdata/source_identity_imports", name))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, region := range []bool{false, true} {
		program, loadError := load.Load([]string{path})
		if loadError != nil {
			test.Fatal(loadError)
		}
		if name == "import_equals.a" {
			// The existing namespace-value gap can refuse this fixture before
			// generic lowering asks for its imported type's identity. Hold the
			// fixture's route too, so dropping the edge cannot hide behind that gap.
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			reference := file.Statements.Nodes[0].AsImportEqualsDeclaration().ModuleReference.AsExternalModuleReference().Expression
			symbol := checked.GetSymbolAtLocation(reference)
			if symbol == nil || len(symbol.Declarations) == 0 || symbol.Declarations[0].Kind != ast.KindSourceFile {
				release()
				test.Fatal("missing fixture module")
			}
			lowering := &lowering{program: program, checker: checked}
			route := lowering.modulePath(symbol.Declarations[0].AsSourceFile())
			release()
			if !strings.Contains(route, "import-equals:true:x") {
				test.Fatalf("fixture import-equals route: %s", route)
			}
		}
		_, lowerWithOptionsError := LowerWithOptions(context.Background(), program, Options{ProgramRegion: region})
		var gap *NotYet
		if !errors.As(lowerWithOptionsError, &gap) || gap.What != what || gap.Where != path+position {
			test.Fatalf(
				"ProgramRegion=%t: got %v, want %s: stage 0 can't lower %s yet",
				region,
				lowerWithOptionsError,
				path+position,
				what,
			)
		}
	}
}

// Runtime import assignment is already rejected by the pinned ESM checker. The
// type-only fixture above gets past that boundary and exercises route indexing.
func TestStableRuntimeImportEqualsDiagnostic(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("testdata/source_identity_imports/runtime_import_equals.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	_, loadError := load.Load([]string{path})
	want := path +
		":2:1: error TS1202: Import assignment cannot be used when targeting ECMAScript modules. Consider using 'import * as ns from \"mod\"', 'import {a} from \"mod\"', 'import d from \"mod\"', or another module format instead."
	if loadError == nil || loadError.Error() != want {
		test.Fatalf("got %v; want %s", loadError, want)
	}
}

// Refusals cannot observe the route's stability. Check that moving the cluster
// and rewriting only its specifier preserves it, while changing the binding does not.
func TestStableImportEqualsRouteBindings(test *testing.T) {
	test.Parallel()
	routes := []string{}
	for index, binding := range []string{"x", "x", "other"} {
		root := test.TempDir()
		target := filepath.Join(root, "model.a")
		model, readError := os.ReadFile("testdata/source_identity_imports/model.a")
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(target, model, 0600); writeError != nil {
			test.Fatal(writeError)
		}
		specifier := "./model.a"
		if index > 0 {
			specifier = filepath.ToSlash(target)
		}
		entry := filepath.Join(root, "main.a")
		if writeError := os.WriteFile(
			entry,
			[]byte("import type "+binding+" = require("+strconv.Quote(specifier)+");"),
			0600,
		); writeError != nil {
			test.Fatal(writeError)
		}
		program, loadError := load.Load([]string{entry})
		if loadError != nil {
			test.Fatal(loadError)
		}
		file := program.Files()[0]
		checked, release := program.Checker(context.Background(), file)
		reference := file.Statements.Nodes[0].AsImportEqualsDeclaration().ModuleReference.AsExternalModuleReference().Expression
		symbol := checked.GetSymbolAtLocation(reference)
		if symbol == nil || len(symbol.Declarations) == 0 || symbol.Declarations[0].Kind != ast.KindSourceFile {
			release()
			test.Fatal("missing imported source module")
		}
		lowering := &lowering{program: program, checker: checked}
		route := lowering.modulePath(symbol.Declarations[0].AsSourceFile())
		release()
		if !strings.Contains(route, "import-equals:true:"+binding) || strings.Contains(route, root) {
			test.Fatalf("route must use import kind and binding, not specifier: %s", route)
		}
		routes = append(routes, route)
	}
	if routes[0] != routes[1] || routes[0] == routes[2] {
		test.Fatalf("specifier/binding identity mismatch: %q", routes)
	}
}
