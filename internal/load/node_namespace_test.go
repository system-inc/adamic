package load

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestNodeNamespaceAnnotationUsesPinnedTimeout(t *testing.T) {
	file, err := filepath.Abs("../oracle/testdata/node_namespace_timeout.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := Load([]string{file})
	if err != nil {
		t.Fatalf("qualified Timeout annotation must load pinned Node declarations: %v", err)
	}
	c, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	var timeout, optional *checker.Type
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypeReference {
			timeout = c.GetTypeAtLocation(node)
		}
		if node.Kind == ast.KindUnionType {
			optional = c.GetTypeAtLocation(node)
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	if timeout == nil || timeout.Flags()&checker.TypeFlagsAny != 0 || timeout.Symbol() == nil || timeout.Symbol().Name != "Timeout" {
		t.Fatal("qualified annotation lost the real Timeout type")
	}
	fromTimers := false
	for _, declaration := range timeout.Symbol().Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if IsNodeLibrary(file) && strings.HasSuffix(file.FileName().AsString(), "/timers.d.ts") {
			fromTimers = true
		}
	}
	if !fromTimers {
		t.Fatal("Timeout must be the class from the pinned @types/node timers.d.ts")
	}
	if optional == nil || optional.Flags()&checker.TypeFlagsUnion == 0 || len(optional.Types()) != 2 {
		t.Fatal("Timeout or undefined must retain both union members")
	}
	foundTimeout, foundUndefined := false, false
	for _, member := range optional.Types() {
		foundTimeout = foundTimeout || member == timeout
		foundUndefined = foundUndefined || member.Flags()&checker.TypeFlagsUndefined != 0
		if member.Flags()&checker.TypeFlagsAny != 0 {
			t.Fatal("optional Timeout must never become any")
		}
	}
	if !foundTimeout || !foundUndefined {
		t.Fatal("optional annotation changed its checker type identities")
	}
}

func TestNodeNamespaceAliasesKeepTimeout(t *testing.T) {
	program, err := nodeSource(t, `type Handle = NodeJS.Timeout; type Copy = Handle;`)
	if err != nil {
		t.Fatal(err)
	}
	c, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	for _, statement := range program.Files()[0].Statements.Nodes {
		typ := c.GetTypeAtLocation(statement.AsTypeAliasDeclaration().Type)
		if typ.Flags()&checker.TypeFlagsAny != 0 || typ.Symbol() == nil || typ.Symbol().Name != "Timeout" {
			t.Fatal("a local alias must preserve Timeout")
		}
	}
}

func TestNodeNamespaceUnresolvedTypesStayRejected(t *testing.T) {
	for _, test := range []struct{ source, code, name string }{
		{`type Handle = NodeJS.DoesNotExist;`, "TS2694", "DoesNotExist"},
		{`type Handle = MissingNode.Timeout;`, "TS2503", "MissingNode"},
	} {
		_, err := nodeSource(t, test.source)
		if err == nil || !strings.Contains(err.Error(), test.code) || !strings.Contains(err.Error(), test.name) {
			t.Fatalf("unresolved type must retain its named checker refusal: %v", err)
		}
	}
}

func TestNodeNamespaceLocalDeclarationIsNotReplaced(t *testing.T) {
	program, err := nodeSource(t, `declare namespace NodeJS { interface Timeout { readonly tag: 'local'; } } type Handle = NodeJS.Timeout;`)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range program.compiler.GetSourceFiles() {
		if IsNodeLibrary(file) {
			t.Fatal("a resolved local namespace must not request ambient Node declarations")
		}
	}
}

func TestNodeNamespaceLocalMissingMemberIsNotReplaced(t *testing.T) {
	// Missing a member does not make the locally declared namespace a request
	// for Node globals. Buffer remains unavailable without such a request.
	_, err := nodeSource(t, `declare namespace NodeJS { interface Other {} } type Handle = NodeJS.DoesNotExist; type Bytes = Buffer;`)
	if err == nil || !strings.Contains(err.Error(), "TS2694") || !strings.Contains(err.Error(), "DoesNotExist") || !strings.Contains(err.Error(), "Cannot find name 'Buffer'") {
		t.Fatalf("unresolved local member must not load unrelated Node globals: %v", err)
	}
}
