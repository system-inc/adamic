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

func TestParserFactoryBindingHoisting(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../oracle/testdata/namespaces_parser_factory.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checker, release := program.Checker(context.Background(), file)
	defer release()
	lowering := &lowering{program: program, checker: checker, result: &ir.Program{}, this: -1, functionIndex: -1}
	// Exercise hoisting independently of the old front-end destructuring refusal,
	// reproducing the latent census path that reached the binding pattern itself.
	if declareModuleError := lowering.declareModule(file.Statements.Nodes); declareModuleError != nil {
		test.Fatal(declareModuleError)
	}
	bindings := map[string]int{}
	for _, local := range lowering.result.Locals {
		if local.NamespaceVar {
			bindings[local.Name]++
		}
	}
	if len(bindings) != 2 || bindings["factoryCreateNodeArray"] != 1 || bindings["factoryCreateNumericLiteral"] != 1 {
		test.Fatalf(
			"NamespaceVar bindings = %v, want factoryCreateNodeArray and factoryCreateNumericLiteral once each",
			bindings,
		)
	}
	for _, statement := range file.Statements.Nodes {
		if statement.Kind == ast.KindModuleDeclaration {
			body, namespaceBodyError := lowering.namespaceBody(statement)
			if namespaceBodyError != nil {
				test.Fatal(namespaceBodyError)
			}
			if len(body) == 0 {
				test.Fatal("parser factory namespace body is empty")
			}
		}
	}
}

func TestMissingBindingSymbolHasStructuredLocation(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte("const { field: local } = { field: 1 };"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checker, release := program.Checker(context.Background(), file)
	defer release()
	pattern := file.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name()
	lowering := &lowering{program: program, checker: checker, result: &ir.Program{}}
	_, declareLocalError := lowering.declareLocal(pattern)
	var notYet *NotYet
	if !errors.As(declareLocalError, &notYet) ||
		notYet.Where != program.Where(pattern) ||
		!strings.Contains(notYet.What, "declaration no symbol") {
		test.Fatalf(
			"got %v, want a structured missing-symbol diagnostic at %s",
			declareLocalError,
			program.Where(pattern),
		)
	}
}
