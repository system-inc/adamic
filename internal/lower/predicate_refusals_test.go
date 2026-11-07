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
			path, err := filepath.Abs("../oracle/testdata/predicate_refusals/oct8_predicates_" + name + ".a")
			if err != nil {
				test.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				test.Fatal(err)
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
			_, err = proof.provePredicate(predicate)
			if err == nil ||
				!strings.Contains(err.Error(), "helper argument 0 does not occupy its predicate parameter") ||
				!strings.Contains(err.Error(), "pass the tested value at the helper's predicate parameter index") ||
				!strings.Contains(err.Error(), filepath.Base(path)+":5:10:") {
				test.Fatalf("want helper call path and parameter-index fix, got %v", err)
			}
		})
	}
}
