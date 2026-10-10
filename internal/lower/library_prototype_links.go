package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Empty complete shapes cannot contain a field back-edge. Prototype-only cycles
// are rejected before retain. Nonempty prototypes require inherited field
// dispatch and the compiler's ordinary graph-region ownership proof.
func (l *lowering) objectPrototypeLink(node *ast.Node, name string) (ir.Expression, bool, error) {
	fail := func(reason string) (ir.Expression, bool, error) {
		return nil, true, l.notYet(node, "Object."+name+": "+reason)
	}
	args := node.AsCallExpression().Arguments.Nodes
	if name == "setPrototypeOf" && strings.Contains(l.program.Where(node), ".a:") {
		return nil, true, &Refused{Where: l.program.Where(node), What: "changing prototypes after creation", Fix: "outside the .a sound subset; use a .ts program with proven heap lifetimes"}
	}
	count := 1
	if name == "setPrototypeOf" {
		count = 2
	}
	if len(args) != count {
		return fail("unproven argument count")
	}
	prototype := args[0]
	if name == "setPrototypeOf" {
		prototype = args[1]
	}
	if name == "getPrototypeOf" {
		if !l.exactObject(args[0], 0) && !l.emptyPrototypeAllocation(args[0]) {
			return fail("receiver identity and prototype provenance are unproven")
		}
	} else {
		intrinsic := l.objectPrototypeObject(prototype)
		if !intrinsic && (!l.exactObject(prototype, 0) || len(l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(prototype))) != 0) {
			for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(prototype)) {
				of, known := l.representation(l.checker.GetTypeOfSymbol(field))
				if !known || of == ir.Object || of == ir.Array || of == ir.Map || of == ir.Closure || of == ir.Union {
					return fail("field '" + field.Name + "' plus [[Prototype]] link can close a counted cycle; compiler graph-region proof is the way out")
				}
			}
			return fail("prototype link across region lifetimes; compiler graph-region proof is the way out (prototype identity, inherited data or hidden fields are unproven)")
		}
		if name == "setPrototypeOf" && !l.exactObject(args[0], 0) && !l.emptyPrototypeAllocation(args[0]) {
			return fail("prototype link across region lifetimes; compiler graph-region proof is the way out (receiver field '<hidden field>' plus [[Prototype]] link is unproven)")
		}
		if name == "create" {
			parent := node.Parent
			if parent == nil || parent.Kind != ast.KindVariableDeclaration || parent.AsVariableDeclaration().Type == nil || len(l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(parent.Name()))) != 0 {
				return fail("fresh result must have an explicit empty object view; inherited data fields are not implemented")
			}
		}
	}
	call := ir.ObjectCall{Method: name, Returns: ir.Object}
	for _, arg := range args {
		if l.objectPrototypeObject(arg) {
			call.Arguments = append(call.Arguments, ir.ObjectCall{Method: "intrinsicObjectPrototype", Returns: ir.Object})
			continue
		}
		value, err := l.expression(arg)
		if err != nil {
			return nil, true, err
		}
		// An immutable Object.create binding proves an object even in the compiler
		// representation of {}, which also admits boxed primitive values.
		if value.Type() == ir.Union && l.emptyPrototypeAllocation(arg) {
			value = ir.Narrow{Value: value, To: ir.Object}
		}
		if value.Type() != ir.Object {
			return fail("receiver or prototype is not a proven plain object")
		}
		call.Arguments = append(call.Arguments, value)
	}
	return call, true, nil
}
func (l *lowering) emptyPrototypeAllocation(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Type == nil || variable.Initializer == nil || l.objectBindingAssigned(declaration, symbol) {
		return false
	}
	if len(l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(node))) != 0 {
		return false
	}
	initializer := ast.SkipParentheses(variable.Initializer)
	if initializer.Kind != ast.KindCallExpression {
		return false
	}
	call := initializer.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "create" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object")
}

func (l *lowering) objectPrototypeObject(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "prototype" && node.AsPropertyAccessExpression().QuestionDotToken == nil && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Object")
}
