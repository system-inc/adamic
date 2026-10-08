package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// memberKey is a class member's slot or method name. A private name is qualified by the class that
// declares it (privateName), the same spelling fieldName gives every read and write of it, so a
// definition and its uses always agree, whichever instantiation or static object holds it.
func (l *lowering) memberKey(name *ast.Node, class int) string {
	if name.Kind == ast.KindComputedPropertyName {
		return iteratorSlot
	}
	if name.Kind == ast.KindPrivateIdentifier {
		if declaring := l.declaringClass(name); declaring != nil {
			return l.privateName(name.Text(), declaring)
		}
		return name.Text() + "@" + strconv.Itoa(class)
	}
	return name.Text()
}

// declaringClass is the class declaration a private name belongs to, or nil.
func (l *lowering) declaringClass(name *ast.Node) *ast.Node {
	if symbol := l.checker.GetSymbolAtLocation(name); symbol != nil && len(symbol.Declarations) > 0 {
		if class := symbol.Declarations[0].Parent; class != nil && class.Kind == ast.KindClassDeclaration {
			return class
		}
	}
	if name.Parent != nil && name.Parent.Parent != nil && name.Parent.Parent.Kind == ast.KindClassDeclaration {
		return name.Parent.Parent
	}
	return nil
}

// privateName is a private name's field name: its spelling qualified by the class declaring it, numbered
// in the order classes are first met. By the declaration, not an instantiation: inside Box<number>,
// other.#label with other a Box<string> is Box<string>'s slot, and every instantiation holds its private
// members under the same names; a class chain holds each declaration once, so names never collide.
func (l *lowering) privateName(spelling string, class *ast.Node) string {
	if l.privateOwners == nil {
		l.privateOwners = map[*ast.Node]int{}
	}
	owner, isKnown := l.privateOwners[class]
	if !isKnown {
		owner = len(l.privateOwners) + 1
		l.privateOwners[class] = owner
	}
	return spelling + "@" + strconv.Itoa(owner)
}

func (l *lowering) hasPrivateStorage(proven *checker.Type) bool {
	proven = l.concrete(proven)
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		for _, declaration := range property.Declarations {
			if declaration.Kind == ast.KindPropertyDeclaration && declaration.Name().Kind == ast.KindPrivateIdentifier {
				return true
			}
		}
	}
	return false
}

func (l *lowering) objectKeys(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "keys" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
		return nil, false, nil
	}
	arguments := nodesOf(node.AsCallExpression().Arguments)
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, "Object.keys without exactly one argument")
	}
	// A spread of a possibly absent plain object has synthetic optional slots, not real keys.
	// Class and accessor descriptors provide the public shape for their stored fields.
	proven := l.checker.GetTypeAtLocation(arguments[0])
	if !isClassInstance(proven) && l.iteratorMember(proven) != nil {
		return nil, true, l.notYet(node, "Object.keys on a literal with symbol-key storage")
	}
	// Structural views preserve the object's storage, including a literal's hidden symbol slot.
	// Only filter when such a literal can inhabit this view; ordinary Object.keys stays unchanged.
	filterSymbol, explicitSlot := false, false
	argument := ast.SkipParentheses(arguments[0])
	fresh := argument.Kind == ast.KindObjectLiteralExpression
	if fresh {
		for _, property := range argument.AsObjectLiteralExpression().Properties.Nodes {
			fresh = fresh && property.Kind != ast.KindSpreadAssignment
		}
	}
	if !fresh && !isClassInstance(proven) && !l.isStaticType(proven) && l.iteratorMember(proven) == nil {
		modules, err := l.moduleOrder(l.program.Files()[0])
		if err != nil {
			return nil, true, err
		}
		var visit ast.Visitor
		visit = func(candidate *ast.Node) bool {
			// An explicit public string key with this spelling must never be mistaken for
			// synthetic symbol storage. Class and accessor literals can expose such keys.
			var members []*ast.Node
			var shape *checker.Type
			if candidate.Kind == ast.KindClassDeclaration && candidate.Name() != nil {
				members = classMembersWithParameters(candidate)
				shape = l.checker.GetTypeAtLocation(candidate.Name())
			} else if candidate.Kind == ast.KindObjectLiteralExpression {
				members = candidate.AsObjectLiteralExpression().Properties.Nodes
				shape = l.checker.GetTypeAtLocation(candidate)
			}
			if shape != nil && l.iterationShapeFits(shape, proven) {
				for _, member := range members {
					name := member.Name()
					if name != nil && (ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral) && name.Text() == iteratorSlot {
						explicitSlot = true
					}
				}
			}
			if candidate.Kind == ast.KindObjectLiteralExpression {
				shape := l.checker.GetTypeAtLocation(candidate)
				if l.iteratorMember(shape) != nil && l.iterationShapeFits(shape, proven) {
					filterSymbol = true
				}
			}
			return candidate.ForEachChild(visit)
		}
		for _, module := range modules {
			module.AsNode().ForEachChild(visit)
		}
	}
	if filterSymbol && explicitSlot {
		return nil, true, &Refused{Where: l.program.Where(node),
			What: "a structural Object.keys view mixing symbol storage and an explicit __adamic_symbol_iterator string key (adamic/symbol-key-view)",
			Fix:  "rename the explicit __adamic_symbol_iterator member, or pass a fresh object containing only the desired string-keyed fields to Object.keys"}
	}
	if !isClassInstance(proven) && !l.isStaticType(proven) && !l.hasAccessorStorage(proven) {
		if l.isLibraryType(proven, "RegExp", "Error") || l.includesUndefined(proven) {
			return l.objectCall(node, "keys")
		}
		for _, property := range l.checker.GetPropertiesOfType(proven) {
			if property.Flags&ast.SymbolFlagsOptional != 0 {
				return l.objectCall(node, "keys")
			}
		}
	}
	value, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if value.Type() != ir.Object {
		return nil, true, l.notYet(node, "Object.keys on a non-object")
	}
	keys := ir.ObjectKeys{Object: value}
	if !filterSymbol {
		return keys, true, nil
	}
	function := len(l.result.Functions)
	key := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "key", Type: ir.String, Function: function})
	l.result.Functions = append(l.result.Functions, ir.Function{
		Name: "string_key", Closure: true, Parameters: []int{key}, Returns: ir.Boolean,
		Body: []ir.Statement{ir.Return{Value: ir.Binary{Operator: ir.NotEqual,
			Left: ir.Read{Local: key, Of: ir.String}, Right: ir.StringConstant{Index: l.constant(iteratorSlot)},
		}}},
	})
	return ir.ArrayVisit{Method: "filter", Array: keys, Callback: ir.MakeClosure{Function: function}, Element: ir.String, Returns: ir.Boolean}, true, nil
}
