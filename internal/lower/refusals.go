package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// refusal is one construct Adamic 0.1 doesn't allow (docs/0.1.md, "What's refused in 0.1"), and the
// fix to offer for it.
type refusal struct {
	what string
	fix  string
}

// refusals by syntax kind. Each is checked before lowering, so a program learns it has written
// something 0.1 refuses for good, never that stage 0 hasn't got to it yet.
var refusals = map[ast.Kind]refusal{
	ast.KindAwaitExpression:   {"await", "0.1 has no async; it arrives with the concurrency model"},
	ast.KindYieldExpression:   {"yield (generators)", "build an array, or call a function per item"},
	ast.KindDecorator:         {"a decorator", "write the behavior where it applies; 0.1 doesn't rewrite classes at runtime"},
	ast.KindGetAccessor:       {"a getter", "write a method: in 0.1 reading a property is just a read"},
	ast.KindSetAccessor:       {"a setter", "write a method: in 0.1 writing a property is just a write"},
	ast.KindLabeledStatement:  {"a label", "move the loop into a function and return from it"},
	ast.KindWithStatement:     {"with", "name the object you mean"},
	ast.KindDeleteExpression:  {"delete", "an object's shape is fixed; use a Map for keys that come and go"},
	ast.KindDebuggerStatement: {"debugger", "remove it"},
	ast.KindEnumDeclaration:   {"enum", "use a union of string literals, like 'Circle' | 'Square'"},
	ast.KindModuleDeclaration: {"a namespace", "use a module: a file of its own, with named exports"},
	ast.KindVoidExpression:    {"the void operator", "evaluate the expression as a statement"},
	ast.KindIndexSignature:    {"an index signature", "use a Map, which keeps keys in the order they were added"},
	ast.KindExportAssignment:  {"export default", "export by name: one name for one thing"},
	ast.KindTypePredicate:     {"a type predicate", "narrow where you use it, with ===, typeof or instanceof (adamic/no-type-predicate)"},
	ast.KindNonNullExpression: {"the non-null assertion !", "write ?? panic('why it can't be missing'), or narrow and handle the missing case"},
}

// refusedOperators are binary operators 0.1 refuses.
var refusedOperators = map[ast.Kind]refusal{
	ast.KindEqualsEqualsToken:             {"==", "use ===, which doesn't coerce"},
	ast.KindExclamationEqualsToken:        {"!=", "use !==, which doesn't coerce"},
	ast.KindInKeyword:                     {"in", "an object's shape is known; use a discriminant, or a Map"},
	ast.KindCommaToken:                    {"the comma operator", "write each expression as its own statement"},
	ast.KindAmpersandAmpersandEqualsToken: {"&&=", "write the if"},
	ast.KindBarBarEqualsToken:             {"||=", "write the if"},
}

// refuse walks a module for what 0.1 refuses and returns the first, with where it is and the fix.
func (l *lowering) refuse(module *ast.SourceFile) error {
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if refused, isRefused := refusals[node.Kind]; isRefused {
			found = &Refused{Where: l.program.Where(node), What: refused.what, Fix: refused.fix}
			return true
		}
		if node.Kind == ast.KindBinaryExpression {
			if refused, isRefused := refusedOperators[node.AsBinaryExpression().OperatorToken.Kind]; isRefused {
				found = &Refused{Where: l.program.Where(node.AsBinaryExpression().OperatorToken), What: refused.what, Fix: refused.fix}
				return true
			}
		}
		if ast.IsFunctionLike(node) && ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
			found = &Refused{Where: l.program.Where(node), What: "an async function", Fix: "0.1 has no async; it arrives with the concurrency model"}
			return true
		}
		if node.Kind == ast.KindIdentifier && node.Text() == "arguments" {
			// JavaScript's arguments object, not a variable the program named arguments.
			if symbol := l.checker.GetSymbolAtLocation(node); symbol != nil && len(symbol.Declarations) == 0 {
				found = &Refused{Where: l.program.Where(node), What: "arguments", Fix: "name the parameters, or take a rest parameter"}
				return true
			}
		}
		if node.Kind == ast.KindPropertyAccessExpression && !called(node) {
			// A method read as a value loses its object: this is undefined when it's called.
			access := node.AsPropertyAccessExpression()
			// Math.random has its own refusal, called or not.
			isRandom := l.isLibraryGlobal(access.Expression, "Math") && access.Name().Text() == "random"
			if symbol := l.checker.GetSymbolAtLocation(node); symbol != nil && symbol.Flags&ast.SymbolFlagsMethod != 0 && !isRandom {
				object := "its object"
				if ast.IsIdentifier(access.Expression) || access.Expression.Kind == ast.KindThisKeyword {
					object = scannedText(access.Expression)
				}
				parameters := "value"
				if signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(symbol), checker.SignatureKindCall); len(signatures) > 0 {
					names := []string{}
					for _, parameter := range signatures[0].Parameters() {
						names = append(names, parameter.Name)
					}
					parameters = strings.Join(names, ", ")
				}
				found = &Refused{Where: l.program.Where(node), What: "a method read as a value (" + symbol.Name + " would lose its object, and this with it)", Fix: "call it in an arrow that keeps the object: (" + parameters + ") => " + object + "." + symbol.Name + "(" + parameters + ") (unbound-method)"}
				return true
			}
		}
		if err := l.refuseWidening(node); err != nil {
			found = err
			return true
		}
		node.ForEachChild(visit)
		return false
	}
	module.AsNode().ForEachChild(visit)
	return found
}

// called reports whether a property access is what a call calls, through any parentheses around it:
// a method called on its object, which keeps it.
func called(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node.Parent != nil && node.Parent.Kind == ast.KindCallExpression && node.Parent.AsCallExpression().Expression == node
}

// scannedText is an identifier's name, or this.
func scannedText(node *ast.Node) string {
	if node.Kind == ast.KindThisKeyword {
		return "this"
	}
	return node.Text()
}
