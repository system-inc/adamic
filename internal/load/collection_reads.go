package load

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// CollectionLookup follows only immutable local snapshots of an ordinary library Map lookup.
// A method merely named get, a mutable alias, or an assertion is not a lookup contract.
func CollectionLookup(check *checker.Checker, node *ast.Node) *ast.Node {
	return collectionLookup(check, node, map[*ast.Symbol]bool{})
}
func collectionLookup(check *checker.Checker, node *ast.Node, seen map[*ast.Symbol]bool) *ast.Node {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := check.GetSymbolAtLocation(node)
		if symbol == nil || seen[symbol] || len(symbol.Declarations) != 1 {
			return nil
		}
		seen[symbol] = true
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || declaration.AsVariableDeclaration().Initializer == nil {
			return nil
		}
		return collectionLookup(check, declaration.AsVariableDeclaration().Initializer, seen)
	}
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "get" || callee.AsPropertyAccessExpression().QuestionDotToken != nil || len(call.Arguments.Nodes) != 1 {
		return nil
	}
	proven := check.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression)
	symbol := proven.Symbol()
	if symbol == nil || (symbol.Name != "Map" && symbol.Name != "ReadonlyMap") || len(symbol.Declarations) == 0 || !IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return nil
	}
	return node
}

// The step 19 rule permits an otherwise checked read when only lookup absence prevents the
// receiving relation. Keep every other checker error, and remember the exact source occurrence
// for lowering. This does not suppress a wrong payload type or certify a stored undefined.
func (p *Program) collectionReadDiagnostic(ctx context.Context, diagnostic *ast.Diagnostic) bool {
	if diagnostic.Code() != 2322 && diagnostic.Code() != 2345 {
		return false
	}
	file := diagnostic.File()
	if file == nil {
		return false
	}
	check, release := p.compiler.GetTypeCheckerForFile(ctx, file)
	defer release()
	var candidate *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		start := scanner.GetTokenPosOfNode(node, file, false)
		if start > diagnostic.Pos() || node.End() < diagnostic.Pos()+diagnostic.Len() {
			return false
		}
		if node.Kind == ast.KindReturnStatement && start == diagnostic.Pos() {
			initial := node.AsReturnStatement().Expression
			if initial != nil && CollectionLookup(check, initial) != nil {
				candidate = initial
			}
		} else if node.Kind == ast.KindVariableDeclaration && node.Name() != nil && scanner.GetTokenPosOfNode(node.Name(), file, false) == diagnostic.Pos() {
			initial := node.AsVariableDeclaration().Initializer
			if initial != nil && CollectionLookup(check, initial) != nil {
				candidate = initial
			}
		} else if (ast.IsIdentifier(node) || node.Kind == ast.KindCallExpression) && start == diagnostic.Pos() && CollectionLookup(check, node) != nil {
			if target := check.GetContextualType(node, checker.ContextFlagsNone); target != nil {
				candidate = node
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if candidate == nil {
		return false
	}
	source := check.GetTypeAtLocation(candidate)
	target := check.GetContextualType(candidate, checker.ContextFlagsNone)
	if target == nil || !collectionIncludesUndefined(source) || collectionIncludesNull(source) || collectionIncludesUndefined(target) || !check.IsTypeAssignableTo(check.GetNonNullableType(source), target) {
		return false
	}
	if p.collectionReads == nil {
		p.collectionReads = map[*ast.Node]bool{}
	}
	p.collectionReads[ast.SkipParentheses(candidate)] = true
	return true
}
func collectionIncludesUndefined(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUndefined != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if collectionIncludesUndefined(member) {
				return true
			}
		}
	}
	return false
}
func (p *Program) CheckedCollectionRead(node *ast.Node) bool {
	return p.collectionReads[ast.SkipParentheses(node)]
}

func collectionIncludesNull(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsNull != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if collectionIncludesNull(part) {
				return true
			}
		}
	}
	return false
}
