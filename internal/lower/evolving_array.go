package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An automatic array's write receiver can still be any[] while the checker
// gives its reads an evolved element type. Choose storage only when every
// typed observation agrees and each intrinsic push fits that contract. Any
// escaping read, rebinding, or incompatible write loses this proof.
func (l *lowering) evolvingArrayElement(node *ast.Node) (ir.Type, bool) {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return 0, false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return 0, false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Type() != nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return 0, false
	}
	initial := declaration.AsVariableDeclaration().Initializer
	if initial == nil {
		return 0, false
	}
	initial = ast.SkipParentheses(initial)
	if initial.Kind != ast.KindArrayLiteralExpression || len(initial.AsArrayLiteralExpression().Elements.Nodes) != 0 {
		return 0, false
	}
	declared := l.checker.GetTypeAtLocation(declaration.Name())
	if !l.checker.IsArrayType(declared) || l.checker.GetElementTypeOfArrayType(declared).Flags()&checker.TypeFlagsAny == 0 {
		return 0, false
	}
	root := declaration
	for root.Parent != nil && !ast.IsFunctionLike(root) {
		root = root.Parent
	}
	var observed *checker.Type
	var writes []*ast.Node
	valid := true
	var visit ast.Visitor
	visit = func(use *ast.Node) bool {
		if ast.IsIdentifier(use) && !ast.IsPartOfTypeNode(use) && use != declaration.Name() && l.symbol(use) == symbol {
			read := use
			for read.Parent != nil && read.Parent.Kind == ast.KindParenthesizedExpression {
				read = read.Parent
			}
			if read.Parent != nil && read.Parent.Kind == ast.KindPropertyAccessExpression && read.Parent.AsPropertyAccessExpression().Expression == read {
				access := read.Parent
				if access.Name().Text() == "push" && access.Parent != nil && access.Parent.Kind == ast.KindCallExpression && access.Parent.AsCallExpression().Expression == access {
					call := access.Parent
					if hasSpread(call) || call.AsCallExpression().QuestionDotToken != nil {
						valid = false
					} else {
						writes = append(writes, call.AsCallExpression().Arguments.Nodes...)
					}
					return false
				}
				if access.Name().Text() == "length" {
					return false
				}
			}
			proven := l.concrete(l.checker.GetTypeAtLocation(use))
			if !l.checker.IsArrayType(proven) {
				valid = false
			} else {
				element := l.concrete(l.checker.GetElementTypeOfArrayType(proven))
				if element.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNever) != 0 {
					valid = false
				} else if observed == nil {
					observed = element
				} else if !identicalTypes(l.checker, observed, element) {
					valid = false
				}
			}
		}
		return use.ForEachChild(visit)
	}
	root.ForEachChild(visit)
	if !valid || observed == nil {
		return 0, false
	}
	held, known := l.kept(observed)
	if !known || slotless(held) && held != ir.Union {
		return 0, false
	}
	for _, write := range writes {
		actual := l.concrete(l.checker.GetTypeAtLocation(write))
		if actual.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 || !l.checker.IsTypeAssignableTo(actual, observed) || !l.sameKeeping(actual, observed, map[[2]*checker.Type]bool{}) || l.widened(actual, observed, map[[2]*checker.Type]bool{}) != nil {
			return 0, false
		}
	}
	return held, true
}
