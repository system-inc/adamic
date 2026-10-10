package lower

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func TestOverloadInferenceWitnesses(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../oracle/testdata/overload_inference_contracts.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{checker: checked, program: program}
	union, optional := file.Statements.Nodes[0], file.Statements.Nodes[1]
	first := checked.GetTypeAtLocation(union.TypeParameters()[0].Name())
	second := checked.GetTypeAtLocation(union.TypeParameters()[1].Name())
	unionResult := checked.GetReturnTypeOfSignature(checked.GetSignatureFromDeclaration(union))
	text := checked.GetTypeAtLocation(file.Statements.Nodes[2].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name())
	number := checked.GetTypeAtLocation(file.Statements.Nodes[3].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name())
	t.Run("one compatible binder", func(t *testing.T) {
		inferred := map[*checker.Type]*checker.Type{first: text}
		l.inferTypes(unionResult, text, inferred)
		if inferred[second] != text {
			t.Fatalf("missing compatible union witness")
		}
	})
	t.Run("incompatible known member", func(t *testing.T) {
		inferred := map[*checker.Type]*checker.Type{first: number}
		l.inferTypes(unionResult, text, inferred)
		if inferred[second] != nil {
			t.Fatalf("inferred a union witness despite incompatible known number")
		}
	})
	t.Run("multiple unknown binders", func(t *testing.T) {
		inferred := map[*checker.Type]*checker.Type{}
		l.inferTypes(unionResult, text, inferred)
		if len(inferred) != 0 {
			t.Fatalf("invented witnesses for multiple unknown binders")
		}
	})
	t.Run("optional rigid binder", func(t *testing.T) {
		binder := checked.GetTypeAtLocation(optional.TypeParameters()[0].Name())
		result := checked.GetReturnTypeOfSignature(checked.GetSignatureFromDeclaration(optional))
		inferred := map[*checker.Type]*checker.Type{}
		l.inferTypes(result, first, inferred)
		if inferred[binder] != first {
			t.Fatalf("lost the rigid binder behind undefined")
		}
	})
}
