package lower

import (
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (lowering *lowering) isArrayPredicateCall(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	return call.QuestionDotToken == nil &&
		len(call.Arguments.Nodes) == 1 &&
		callee.Kind == ast.KindPropertyAccessExpression &&
		callee.AsPropertyAccessExpression().QuestionDotToken == nil &&
		callee.Name().Text() == "isArray" &&
		lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array") &&
		lowering.arrayPredicateMember(callee)
}

// The prelude's readonly unknown[] overload merges into the library's isArray, so the builtin is
// the member whose every declaration is the library's or the prelude's, never the program's.
func (lowering *lowering) arrayPredicateMember(callee *ast.Node) bool {
	symbol := lowering.memberSymbol(callee)
	if symbol == nil {
		return false
	}
	roots := lowering.checker.GetRootSymbols(symbol)
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if len(root.Declarations) == 0 {
			return false
		}
		for _, declaration := range root.Declarations {
			if file := ast.GetSourceFileOfNode(declaration); !load.IsLibrary(file) && !load.IsPrelude(file) {
				return false
			}
		}
	}
	return true
}

func (lowering *lowering) arrayIsArray(node *ast.Node) (ir.Expression, bool, error) {
	if !lowering.isArrayPredicateCall(node) {
		return nil, false, nil
	}
	argument := node.AsCallExpression().Arguments.Nodes[0]
	if !lowering.arrayPredicateDomain(lowering.checker.GetTypeAtLocation(argument)) &&
		!lowering.exactObject(argument, 0) {
		return nil, true, lowering.notYet(argument, "Array.isArray on a tuple or an erased object/any/unknown view")
	}
	value, expressionError := lowering.expression(argument)
	if expressionError != nil {
		return nil, true, expressionError
	}
	return ir.ArrayIsArray{Value: fit(value, ir.Union)}, true, nil
}

// isArray's predicate, the prelude's readonly unknown[] overload in place of the bundled any[]
// one, describes the runtime test, not an element contract. Only array members already promised
// by the declared input survive.
func (lowering *lowering) arrayPredicateMembers(proven *checker.Type, array bool) *checker.Type {
	members := []*checker.Type{}
	var collect func(*checker.Type)
	collect = func(member *checker.Type) {
		member = lowering.concrete(member)
		if member.Flags()&checker.TypeFlagsUnion != 0 {
			for _, part := range member.Types() {
				collect(part)
			}
			return
		}
		if member.Flags()&checker.TypeFlagsUnknown != 0 {
			if array {
				members = append(members, predicateReadonlyArray(lowering.checker, member, true))
			} else {
				members = append(members, member)
			}
			return
		}
		if member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
			return
		}
		_, _, nodeArray := lowering.nodeArrayLayoutOf(member)
		if (lowering.isArrayType(member) || nodeArray) == array {
			members = append(members, member)
		}
	}
	collect(proven)
	return lowering.checker.GetUnionType(members)
}

// Tuples currently have an object representation. Erased object views can hide
// one too, so the native array brand cannot decide their JavaScript arrayness.
func (lowering *lowering) arrayPredicateDomain(proven *checker.Type) bool {
	proven = lowering.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !lowering.arrayPredicateDomain(member) {
				return false
			}
		}
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnknown != 0 {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter|checker.TypeFlagsIntersection) != 0 ||
		checker.IsTupleType(proven) {
		return false
	}
	if _, _, known := lowering.nodeArrayLayoutOf(proven); known {
		return true
	}
	if lowering.isArrayType(proven) {
		return true
	}
	return proven.Flags()&checker.TypeFlagsObject == 0 ||
		isClassInstance(proven) ||
		lowering.isLibraryType(proven, "Uint8Array", "Int32Array", "Float64Array")
}

// The pinned checker owns construction of the readonly array type, just as it
// owns predicate flow facts. This introduces no any element contract.
//
//go:linkname predicateReadonlyArray github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).createArrayTypeEx
func predicateReadonlyArray(receiver *checker.Checker, element *checker.Type, readonly bool) *checker.Type
