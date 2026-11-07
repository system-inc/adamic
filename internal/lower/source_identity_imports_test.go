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
	path, err := filepath.Abs(filepath.Join("testdata/source_identity_imports", name))
	if err != nil {
		test.Fatal(err)
	}
	for _, region := range []bool{false, true} {
		program, err := load.Load([]string{path})
		if err != nil {
			test.Fatal(err)
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
			l := &lowering{program: program, checker: checked}
			route := l.modulePath(symbol.Declarations[0].AsSourceFile())
			release()
			if !strings.Contains(route, "import-equals:true:x") {
				test.Fatalf("fixture import-equals route: %s", route)
			}
		}
		_, err = LowerWithOptions(context.Background(), program, Options{ProgramRegion: region})
		var gap *NotYet
		if !errors.As(err, &gap) || gap.What != what || gap.Where != path+position {
			test.Fatalf(
				"ProgramRegion=%t: got %v, want %s: stage 0 can't lower %s yet",
				region,
				err,
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
	path, err := filepath.Abs("testdata/source_identity_imports/runtime_import_equals.a")
	if err != nil {
		test.Fatal(err)
	}
	_, err = load.Load([]string{path})
	want := path +
		":2:1: error TS1202: Import assignment cannot be used when targeting ECMAScript modules. Consider using 'import * as ns from \"mod\"', 'import {a} from \"mod\"', 'import d from \"mod\"', or another module format instead."
	if err == nil || err.Error() != want {
		test.Fatalf("got %v; want %s", err, want)
	}
}

// Refusals cannot observe the route's stability. Check that moving the cluster
// and rewriting only its specifier preserves it, while changing the binding does not.
func TestStableImportEqualsRouteBindings(test *testing.T) {
	test.Parallel()
	routes := []string{}
	for i, binding := range []string{"x", "x", "other"} {
		root := test.TempDir()
		target := filepath.Join(root, "model.a")
		model, err := os.ReadFile("testdata/source_identity_imports/model.a")
		if err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(target, model, 0600); err != nil {
			test.Fatal(err)
		}
		specifier := "./model.a"
		if i > 0 {
			specifier = filepath.ToSlash(target)
		}
		entry := filepath.Join(root, "main.a")
		if err := os.WriteFile(
			entry,
			[]byte("import type "+binding+" = require("+strconv.Quote(specifier)+");"),
			0600,
		); err != nil {
			test.Fatal(err)
		}
		program, err := load.Load([]string{entry})
		if err != nil {
			test.Fatal(err)
		}
		file := program.Files()[0]
		checked, release := program.Checker(context.Background(), file)
		reference := file.Statements.Nodes[0].AsImportEqualsDeclaration().ModuleReference.AsExternalModuleReference().Expression
		symbol := checked.GetSymbolAtLocation(reference)
		if symbol == nil || len(symbol.Declarations) == 0 || symbol.Declarations[0].Kind != ast.KindSourceFile {
			release()
			test.Fatal("missing imported source module")
		}
		l := &lowering{program: program, checker: checked}
		route := l.modulePath(symbol.Declarations[0].AsSourceFile())
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
