package lower

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func TestPredicateSummaryParameterIndex(test *testing.T) {
	test.Parallel()
	for _, name := range []string{
		"p05_helper_second_parameter",
		"p06_helper_second_parameter_undefined",
		"p13_filter_wrong_parameter_helper",
	} {
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				"../oracle/testdata/predicate_refusals/oct8_predicates_" + name + ".a",
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			proof := &lowering{program: program, checker: checked}
			var predicate *ast.Node
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindFunctionDeclaration && node.Name() != nil && node.Name().Text() == "g" {
					predicate = node.Type()
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			if predicate == nil {
				test.Fatal("missing g predicate")
			}
			_, provePredicateError := proof.provePredicate(predicate)
			if provePredicateError == nil ||
				!strings.Contains(
					provePredicateError.Error(),
					"helper argument 0 does not occupy its predicate parameter",
				) ||
				!strings.Contains(
					provePredicateError.Error(),
					"pass the tested value at the helper's predicate parameter index",
				) ||
				!strings.Contains(provePredicateError.Error(), filepath.Base(path)+":5:10:") {
				test.Fatalf("want helper call path and parameter-index fix, got %v", provePredicateError)
			}
		})
	}
}
