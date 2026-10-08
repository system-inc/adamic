package lower

import (
	"context"
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedSignatureInstantiation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.a")
	source := `interface Node { readonly flags: number }
function clone<T extends Node>(value:T):T { return value; }
const produce: <U extends Node>(value:U)=>U = clone;
const concrete: (value:Node)=>Node = clone;
type Mutable<T> = { -readonly [K in keyof T]: T[K] };
type PartialCopy<T> = { [K in keyof T]?: T[K] };
function identity<T extends Node>(value:Mutable<T>):T { return value; }
function unproven<T extends Node>(value:PartialCopy<T>):T { return value as T; }
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	comparisons := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			declaration := node.AsVariableDeclaration()
			own := checked.GetTypeAtLocation(declaration.Initializer)
			view := checked.GetTypeAtLocation(declaration.Name())
			producing, receiving := l.checkedSignaturePair(checked.GetSignaturesOfType(own, checker.SignatureKindCall)[0], checked.GetSignaturesOfType(view, checker.SignatureKindCall)[0])
			if checked.GetReturnTypeOfSignature(producing) != checked.GetReturnTypeOfSignature(receiving) {
				t.Fatalf("%s: producer binder was not instantiated", declaration.Name().Text())
			}
			if found := l.widened(own, view, map[[2]*checker.Type]bool{}); found != nil {
				t.Fatalf("%s: invariant instantiated relation refused", declaration.Name().Text())
			}
			comparisons++
		}
		if node.Kind == ast.KindFunctionDeclaration && (node.Name().Text() == "identity" || node.Name().Text() == "unproven") {
			signature := checked.GetSignatureFromDeclaration(node)
			source := checked.GetTypeOfSymbol(signature.Parameters()[0])
			target := checked.GetReturnTypeOfSignature(signature)
			got := l.checkedMutableIdentity(source, target)
			if got != (node.Name().Text() == "identity") {
				t.Fatalf("%s: mapped identity proof %t", node.Name().Text(), got)
			}
			comparisons++
		}
		node.ForEachChild(visit)
		return false
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	if comparisons != 4 {
		t.Fatalf("measured %d signature relations", comparisons)
	}
}

func TestCheckedSignatureEveryReceivingOverload(t *testing.T) {
	_, err := lowerSource(t, `interface Node { readonly flags: number }
interface Narrow extends Node { readonly name: string }
interface Receiver { handle(node: Narrow): Narrow; handle(node: Node): Node }
const producer = { handle(node: Narrow): Narrow { return node; } };
const receiver: Receiver = producer;
`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "function taking Narrow seen as one taking Node") {
		t.Fatalf("second receiving overload was not refused: %v", err)
	}
}
