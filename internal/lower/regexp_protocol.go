package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	regex "github.com/system-inc/adamic/internal/regexp"
	"strings"
)

// Node's u split path depends on its unobservable Smi/HeapNumber limit tag.
// Only immediate small integer literals prove that tag. Other flags and
// patterns without these assertions have no observable candidate-scan difference.
func (l *lowering) regexSplitLimitProven(receiver *ast.Node, limit ir.Expression) bool {
	if _, ok := limit.(ir.Undefined); ok {
		return true
	}
	if _, ok := limit.(ir.NumberConstant); ok {
		return true
	}
	return l.regexSplitPatternsSafe(receiver, 0)
}

// A loop over an immutable list of constant regexes can prove the absence of
// the V8-sensitive shapes without refusing its ordinary dynamic split limits.
func (l *lowering) regexSplitPatternsSafe(node *ast.Node, depth int) bool {
	if depth > 32 {
		return false
	}
	pattern, flags, known := l.regexSplitConstant(node)
	if known {
		if !strings.Contains(flags, "u") {
			return true
		}
		program, err := regex.Compile(pattern, flags)
		return err == nil && !program.MayMatchInsideUnicodePair()
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindArrayLiteralExpression {
		for _, item := range node.AsArrayLiteralExpression().Elements.Nodes {
			if !l.regexSplitPatternsSafe(item, depth+1) {
				return false
			}
		}
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	if declaration.AsVariableDeclaration().Initializer != nil {
		if !l.regexSplitArrayReadOnly(symbol, declaration) {
			return false
		}
		return l.regexSplitPatternsSafe(declaration.AsVariableDeclaration().Initializer, depth+1)
	}
	parent := declaration.Parent.Parent
	if parent != nil && parent.Kind == ast.KindForOfStatement {
		return l.regexSplitPatternsSafe(parent.AsForInOrOfStatement().Expression, depth+1)
	}
	return false
}

func (l *lowering) regexSplitArrayReadOnly(symbol *ast.Symbol, declaration *ast.Node) bool {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	safe := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !safe {
			return true
		}
		if ast.IsIdentifier(node) && l.symbol(node) == symbol && node != declaration.Name() {
			parent := node.Parent
			for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
				node, parent = parent, parent.Parent
			}
			if parent == nil || parent.Kind != ast.KindForOfStatement || parent.AsForInOrOfStatement().Expression != node {
				safe = false
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		visit(module.AsNode())
	}
	return safe
}

// This branch has no mutable compile protocol. Const bindings are followed by
// regexSplitPatternsSafe; new expressions must prove both string arguments.
func (l *lowering) regexSplitConstant(node *ast.Node) (string, string, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindRegularExpressionLiteral {
		text := node.Text()
		end := strings.LastIndex(text, "/")
		if end > 0 {
			return text[1:end], text[end+1:], true
		}
	}
	if node.Kind != ast.KindNewExpression || !l.isLibraryGlobal(node.AsNewExpression().Expression, "RegExp") {
		return "", "", false
	}
	args := node.AsNewExpression().Arguments
	if args == nil {
		return "", "", true
	}
	if len(args.Nodes) > 2 {
		return "", "", false
	}
	pattern, flags := "", ""
	var ok bool
	if len(args.Nodes) > 0 {
		pattern, ok = l.constantPattern(args.Nodes[0], 0)
		if !ok {
			return "", "", false
		}
	}
	if len(args.Nodes) > 1 {
		flags, ok = l.constantPattern(args.Nodes[1], 0)
		if !ok {
			return "", "", false
		}
	}
	return pattern, flags, true
}
