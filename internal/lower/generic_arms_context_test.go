package lower

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Hold the checker-derived signature at every generic operand, including its
// outer nullable context. No syntax-based propagation is used in this check.
func TestGenericFunctionValueOperandContexts(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../oracle/testdata/generic_function_value_arms.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	typeChecker, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	lowering := &lowering{program: program, checker: typeChecker}
	strings, numbers := 0, 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node.Text() == "identity" && typeChecker.GetContextualType(node, checker.ContextFlagsNone) != nil {
			signatures := typeChecker.GetSignaturesOfType(typeChecker.GetTypeAtLocation(node), checker.SignatureKindCall)
			if len(signatures) != 1 {
				t.Fatalf("%s: want one signature", program.Where(node))
			}
			resolved := lowering.contextualGenericSignature(node, signatures[0])
			if resolved == nil {
				t.Fatalf("%s: missing concrete operand signature", program.Where(node))
			}
			argument := typeChecker.GetTypeOfSymbol(resolved.Parameters()[0])
			result := typeChecker.GetReturnTypeOfSignature(resolved)
			if argument.Flags()&checker.TypeFlagsStringLike != 0 && result.Flags()&checker.TypeFlagsStringLike != 0 {
				strings++
			} else if argument.Flags()&checker.TypeFlagsNumberLike != 0 && result.Flags()&checker.TypeFlagsNumberLike != 0 {
				numbers++
			} else {
				t.Fatalf("%s: unexpected signature (%s) => %s", program.Where(node), typeChecker.TypeToString(argument), typeChecker.TypeToString(result))
			}
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].Node.ForEachChild(visit)
	if strings != 7 || numbers != 1 {
		t.Fatalf("got %d string and %d number operand contexts, want 7 and 1", strings, numbers)
	}
}
