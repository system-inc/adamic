package lower

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func TestOverloadInferenceWitnesses(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../oracle/testdata/overload_inference_contracts.a")
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
	lowering := &lowering{checker: checked, program: program}
	union, optional := file.Statements.Nodes[0], file.Statements.Nodes[1]
	first := checked.GetTypeAtLocation(union.TypeParameters()[0].Name())
	second := checked.GetTypeAtLocation(union.TypeParameters()[1].Name())
	unionResult := checked.GetReturnTypeOfSignature(checked.GetSignatureFromDeclaration(union))
	text := checked.GetTypeAtLocation(
		file.Statements.Nodes[2].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name(),
	)
	number := checked.GetTypeAtLocation(
		file.Statements.Nodes[3].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name(),
	)
	test.Run("one compatible binder", func(test *testing.T) {
		inferred := map[*checker.Type]*checker.Type{first: text}
		lowering.inferTypes(unionResult, text, inferred)
		if inferred[second] != text {
			test.Fatalf("missing compatible union witness")
		}
	})
	test.Run("incompatible known member", func(test *testing.T) {
		inferred := map[*checker.Type]*checker.Type{first: number}
		lowering.inferTypes(unionResult, text, inferred)
		if inferred[second] != nil {
			test.Fatalf("inferred a union witness despite incompatible known number")
		}
	})
	test.Run("multiple unknown binders", func(test *testing.T) {
		inferred := map[*checker.Type]*checker.Type{}
		lowering.inferTypes(unionResult, text, inferred)
		if len(inferred) != 0 {
			test.Fatalf("invented witnesses for multiple unknown binders")
		}
	})
	test.Run("optional rigid binder", func(test *testing.T) {
		binder := checked.GetTypeAtLocation(optional.TypeParameters()[0].Name())
		result := checked.GetReturnTypeOfSignature(checked.GetSignatureFromDeclaration(optional))
		inferred := map[*checker.Type]*checker.Type{}
		lowering.inferTypes(result, first, inferred)
		if inferred[binder] != first {
			test.Fatalf("lost the rigid binder behind undefined")
		}
	})
}
