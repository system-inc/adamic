package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestParserFactoryBindingHoisting(t *testing.T) {
	path, err := filepath.Abs("../oracle/testdata/namespaces_parser_factory.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checker, release := program.Checker(context.Background(), file)
	defer release()
	lowering := &lowering{program: program, checker: checker, result: &ir.Program{}, this: -1, functionIndex: -1}
	// Exercise hoisting independently of the old front-end destructuring refusal,
	// reproducing the latent census path that reached the binding pattern itself.
	if err := lowering.declareModule(file.Statements.Nodes); err != nil {
		t.Fatal(err)
	}
	for _, statement := range file.Statements.Nodes {
		if statement.Kind == ast.KindModuleDeclaration {
			if _, err := lowering.namespaceBody(statement); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestMissingBindingSymbolHasStructuredLocation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte("const { field: local } = { field: 1 };"), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checker, release := program.Checker(context.Background(), file)
	defer release()
	pattern := file.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name()
	lowering := &lowering{program: program, checker: checker, result: &ir.Program{}}
	_, err = lowering.declareLocal(pattern)
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.Where != program.Where(pattern) || !strings.Contains(notYet.What, "declaration no symbol") {
		t.Fatalf("got %v, want a structured missing-symbol diagnostic at %s", err, program.Where(pattern))
	}
}
