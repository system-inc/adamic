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
func (lowering *lowering) memberKey(name *ast.Node, class int) string {
	if name.Kind == ast.KindPrivateIdentifier {
		if declaring := lowering.declaringClass(name); declaring != nil {
			return lowering.privateName(name.Text(), declaring)
		}
		return name.Text() + "@" + strconv.Itoa(class)
	}
	return name.Text()
}

// declaringClass is the class declaration a private name belongs to, or nil.
func (lowering *lowering) declaringClass(name *ast.Node) *ast.Node {
	if symbol := lowering.checker.GetSymbolAtLocation(name); symbol != nil && len(symbol.Declarations) > 0 {
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
func (lowering *lowering) privateName(spelling string, class *ast.Node) string {
	if lowering.privateOwners == nil {
		lowering.privateOwners = map[*ast.Node]int{}
	}
	owner, isKnown := lowering.privateOwners[class]
	if !isKnown {
		owner = len(lowering.privateOwners) + 1
		lowering.privateOwners[class] = owner
	}
	return spelling + "@" + strconv.Itoa(owner)
}

func (lowering *lowering) hasPrivateStorage(proven *checker.Type) bool {
	proven = lowering.concrete(proven)
	for _, property := range lowering.checker.GetPropertiesOfType(proven) {
		for _, declaration := range property.Declarations {
			if declaration.Kind == ast.KindPropertyDeclaration && declaration.Name().Kind == ast.KindPrivateIdentifier {
				return true
			}
		}
	}
	return false
}

func (lowering *lowering) objectKeys(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression ||
		callee.Name().Text() != "keys" ||
		!lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
		return nil, false, nil
	}
	arguments := nodesOf(node.AsCallExpression().Arguments)
	if len(arguments) != 1 {
		return nil, true, lowering.notYet(node, "Object.keys without exactly one argument")
	}
	// Node enumerates none of an Error's own properties, and natively they're fields like any other's
	// (errorMayHide). A literal, or a const bound to one, is never an Error.
	if !lowering.exactObject(arguments[0], 0) && lowering.errorMayHide(arguments[0]) {
		return nil, true, lowering.notYet(
			node,
			"Object.keys of a value that may hold an Error, whose own properties Node doesn't enumerate",
		)
	}
	if value, handled, recordObjectCallError := lowering.recordObjectCall(node, "keys"); handled {
		return value, true, recordObjectCallError
	}
	// A spread of a possibly absent plain object has synthetic optional slots, not real keys.
	// Class and accessor descriptors provide the public shape for their stored fields.
	proven := lowering.checker.GetTypeAtLocation(arguments[0])
	if _, _, known := lowering.nodeArrayLayoutOf(proven); known {
		return lowering.objectCall(node, "keys")
	}
	if !isClassInstance(proven) && lowering.iteratorMember(proven) != nil {
		return nil, true, lowering.notYet(node, "Object.keys on a literal with symbol-key storage")
	}
	// A structural view does not remove the iterable literal's hidden symbol slot.
	// Fresh explicit string-keyed literals cannot carry storage erased by a view.
	argument := ast.SkipParentheses(arguments[0])
	fresh := argument.Kind == ast.KindObjectLiteralExpression
	if fresh {
		for _, property := range argument.AsObjectLiteralExpression().Properties.Nodes {
			fresh = fresh && property.Kind != ast.KindSpreadAssignment
		}
	}
	if !fresh && !isClassInstance(proven) && !lowering.isStaticType(proven) && lowering.iteratorMember(proven) == nil {
		modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
		if moduleOrderError != nil {
			return nil, true, moduleOrderError
		}
		var hazard error
		var visit ast.Visitor
		visit = func(candidate *ast.Node) bool {
			if candidate.Kind == ast.KindObjectLiteralExpression {
				shape := lowering.checker.GetTypeAtLocation(candidate)
				if lowering.iteratorMember(shape) != nil && lowering.iterationShapeFits(shape, proven) {
					hazard = &Refused{
						Where: lowering.program.Where(node),
						What:  "a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view)",
						Fix:   "pass a new plain object containing only the desired string-keyed fields to Object.keys",
					}
					return true
				}
			}
			return candidate.ForEachChild(visit)
		}
		for _, module := range modules {
			module.AsNode().ForEachChild(visit)
		}
		if hazard != nil {
			return nil, true, hazard
		}
	}
	if !isClassInstance(proven) && !lowering.isStaticType(proven) && !lowering.hasAccessorStorage(proven) {
		if lowering.isLibraryType(proven, "RegExp") || lowering.includesUndefined(proven) {
			return lowering.objectCall(node, "keys")
		}
		for _, property := range lowering.checker.GetPropertiesOfType(proven) {
			if property.Flags&ast.SymbolFlagsOptional != 0 {
				return lowering.objectCall(node, "keys")
			}
		}
	}
	value, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	if value.Type() != ir.Object {
		return nil, true, lowering.notYet(node, "Object.keys on a non-object")
	}
	return ir.ObjectKeys{Object: value}, true, nil
}
