package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// optionalWidening names the relation whose optional field could be hidden by a structural view.
// This is independent of mutable invariance: even a readonly optional field can lie on a read.
type optionalWidening struct {
	property       string
	source, target *checker.Type
}

func (l *lowering) optionalWidened(source, target *checker.Type, skip map[string]bool, visited map[[2]*checker.Type]bool) *optionalWidening {
	if source == nil || target == nil {
		return nil
	}
	source, target = l.phantomArrayView(source), l.phantomArrayView(target)
	if weak := l.weakTarget(source); weak != nil {
		source = weak
	}
	if weak := l.weakTarget(target); weak != nil {
		target = weak
	}
	// Only an accepted assignment is a widening. Function rest parameters and checked
	// downcasts can pair unrelated types; those belong to their own relation rules.
	if skip == nil && !l.checker.IsTypeAssignableTo(source, target) {
		return nil
	}
	if source == target || visited[[2]*checker.Type{source, target}] {
		return nil
	}
	visited[[2]*checker.Type{source, target}] = true
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if found := l.optionalWidened(member, target, skip, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if l.checker.IsTypeAssignableTo(source, member) {
				if found := l.optionalWidened(source, member, skip, visited); found != nil {
					return found
				}
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.optionalWidened(l.checker.GetBaseConstraintOfType(source), target, skip, visited)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.optionalWidened(source, l.checker.GetBaseConstraintOfType(target), skip, visited)
	}
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	// Container elements are views too. Avoid walking library methods, whose generic signatures
	// describe operations rather than value fields.
	if (l.checker.IsArrayType(source) || checker.IsTupleType(source)) && (l.checker.IsArrayType(target) || checker.IsTupleType(target)) {
		sources, targets := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
		for index, inside := range sources {
			at := index
			if !checker.IsTupleType(target) {
				at = 0
			}
			if at < len(targets) {
				if found := l.optionalWidened(inside, targets[at], nil, visited); found != nil {
					return found
				}
			}
		}
		return nil
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if skip[property.Name] {
			continue
		}
		declared := l.checker.GetPropertyOfType(source, property.Name)
		if declared == nil {
			if property.Flags&ast.SymbolFlagsOptional != 0 && !isClassInstance(source) {
				return &optionalWidening{property: property.Name, source: source, target: target}
			}
			continue
		}
		if found := l.optionalWidened(l.checker.GetTypeOfSymbol(declared), l.checker.GetTypeOfSymbol(property), nil, visited); found != nil {
			return found
		}
	}
	sources := l.checker.GetSignaturesOfType(source, checker.SignatureKindCall)
	targets := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(sources) == 1 && len(targets) == 1 && len(sources[0].TypeParameters()) == 0 && len(targets[0].TypeParameters()) == 0 {
		if found := l.optionalWidened(l.checker.GetReturnTypeOfSignature(sources[0]), l.checker.GetReturnTypeOfSignature(targets[0]), nil, visited); found != nil {
			return found
		}
		takes, gives := sources[0].Parameters(), targets[0].Parameters()
		for index := 0; index < len(takes) && index < len(gives); index++ {
			if found := l.optionalWidened(l.checker.GetTypeOfSymbol(gives[index]), l.checker.GetTypeOfSymbol(takes[index]), nil, visited); found != nil {
				return found
			}
		}
	}
	return nil
}

func (l *lowering) optionalValue(node *ast.Node, target *checker.Type) *optionalWidening {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// The literal's fields, spreads and elements are checked separately at their own sites.
		return nil
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
			if found := l.optionalValue(branch, target); found != nil {
				return found
			}
		}
		return nil
	}
	return l.optionalWidened(l.checker.GetTypeAtLocation(node), target, nil, map[[2]*checker.Type]bool{})
}

// optionalAtSite uses the same relation sites as the other view rules, before lowering.
// A leading spread is not proof of absence: it copies fields hidden by the source's type too.
func (l *lowering) optionalAtSite(node *ast.Node) *optionalWidening {
	// Only argument positions of a directly imported node:* call can use the host exemption.
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindCallExpression && l.optionalNodeHostCall(parent) {
		for index, argument := range parent.AsCallExpression().Arguments.Nodes {
			if argument == node && l.nodeHostConsumesArgument(parent, index) {
				return nil
			}
		}
	}
	var found *optionalWidening
	switch node.Kind {
	case ast.KindAsExpression:
		found = l.optionalValue(node.AsAsExpression().Expression, l.checker.GetTypeAtLocation(node))
	case ast.KindShorthandPropertyAssignment:
		target := l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
		if target != nil {
			found = l.optionalValue(node.Name(), l.checker.GetTypeOfPropertyOfType(target, node.Name().Text()))
		}
	case ast.KindSpreadAssignment:
		target := l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
		if target == nil {
			return nil
		}
		skip := map[string]bool{}
		after := false
		for _, property := range node.Parent.AsObjectLiteralExpression().Properties.Nodes {
			if property == node {
				after = true
			} else if after && property.Name() != nil {
				skip[property.Name().Text()] = true
			}
		}
		expression := ast.SkipParentheses(node.AsSpreadAssignment().Expression)
		if expression.Kind != ast.KindObjectLiteralExpression {
			found = l.optionalWidened(l.checker.GetTypeAtLocation(expression), target, skip, map[[2]*checker.Type]bool{})
		}
	default:
		if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) || l.genericFunction(node) {
			return nil
		}
		var target *checker.Type
		if viewSite(node) {
			target = l.checker.GetContextualType(node, checker.ContextFlagsNone)
		}
		if target == nil {
			target = l.impliedTarget(node)
		}
		if target == nil {
			return nil
		}
		found = l.optionalValue(node, target)
	}
	return found
}

func (l *lowering) refuseOptionalWidening(node *ast.Node) error {
	found := l.optionalAtSite(node)
	if found == nil {
		return nil
	}
	return &Refused{Where: l.program.Where(node),
		What: "optional property " + found.property + " in " + l.checker.TypeToString(found.target) + " absent from structural source " + l.checker.TypeToString(found.source) + ", which can hide fields",
		Fix:  "declare " + found.property + " on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"}
}

// Resolve the import binding rather than trusting the callee's spelling.
func (l *lowering) optionalNodeHostCall(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression {
		callee = ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	}
	symbol := l.checker.GetSymbolAtLocation(callee)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
			if ancestor.Kind == ast.KindImportDeclaration {
				specifier := ancestor.ModuleSpecifier()
				return specifier != nil && strings.HasPrefix(specifier.Text(), "node:")
			}
		}
	}
	return false
}
