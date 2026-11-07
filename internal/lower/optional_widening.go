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

func (lowering *lowering) optionalWidened(
	source, target *checker.Type,
	skip map[string]bool,
	visited map[[2]*checker.Type]bool,
) *optionalWidening {
	if source == nil || target == nil {
		return nil
	}
	if weak := lowering.weakTarget(source); weak != nil {
		source = weak
	}
	if weak := lowering.weakTarget(target); weak != nil {
		target = weak
	}
	// A Pick of a record holds the whole record (records_partial.go), so none of its fields is hidden.
	if base := lowering.pickedRecordBase(source); base != nil {
		source = base
	}
	// Only an accepted assignment is a widening. Function rest parameters and checked
	// downcasts can pair unrelated types; those belong to their own relation rules.
	if skip == nil && !lowering.checker.IsTypeAssignableTo(source, target) {
		return nil
	}
	if source == target || visited[[2]*checker.Type{source, target}] {
		return nil
	}
	visited[[2]*checker.Type{source, target}] = true
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if found := lowering.optionalWidened(member, target, skip, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		// A value enters a union that names its own type as that member, so no other member's field is
		// hidden from it. Seeing the union as another member later is a view of its own, checked there.
		for _, member := range target.Types() {
			if member == source {
				return nil
			}
		}
		for _, member := range target.Types() {
			if lowering.checker.IsTypeAssignableTo(source, member) {
				if found := lowering.optionalWidened(source, member, skip, visited); found != nil {
					return found
				}
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return lowering.optionalWidened(lowering.checker.GetBaseConstraintOfType(source), target, skip, visited)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return lowering.optionalWidened(source, lowering.checker.GetBaseConstraintOfType(target), skip, visited)
	}
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 ||
		target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	// Container elements are views too. Avoid walking library methods, whose generic signatures
	// describe operations rather than value fields.
	if (lowering.isArrayType(source) || checker.IsTupleType(source)) &&
		(lowering.isArrayType(target) || checker.IsTupleType(target)) {
		sources, targets := lowering.checker.GetTypeArguments(source), lowering.checker.GetTypeArguments(target)
		for index, inside := range sources {
			at := index
			if !checker.IsTupleType(target) {
				at = 0
			}
			if at < len(targets) {
				if found := lowering.optionalWidened(inside, targets[at], nil, visited); found != nil {
					return found
				}
			}
		}
		return nil
	}
	for _, property := range lowering.checker.GetPropertiesOfType(target) {
		if skip[property.Name] {
			continue
		}
		declared := lowering.checker.GetPropertyOfType(source, property.Name)
		if declared == nil {
			if property.Flags&ast.SymbolFlagsOptional != 0 && !isClassInstance(source) {
				return &optionalWidening{property: property.Name, source: source, target: target}
			}
			continue
		}
		if found := lowering.optionalWidened(
			lowering.checker.GetTypeOfSymbol(declared),
			lowering.checker.GetTypeOfSymbol(property),
			nil,
			visited,
		); found != nil {
			return found
		}
	}
	sources := lowering.checker.GetSignaturesOfType(source, checker.SignatureKindCall)
	targets := lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(sources) == 1 &&
		len(targets) == 1 &&
		len(sources[0].TypeParameters()) == 0 &&
		len(targets[0].TypeParameters()) == 0 {
		if found := lowering.optionalWidened(
			lowering.checker.GetReturnTypeOfSignature(sources[0]),
			lowering.checker.GetReturnTypeOfSignature(targets[0]),
			nil,
			visited,
		); found != nil {
			return found
		}
		takes, gives := sources[0].Parameters(), targets[0].Parameters()
		for index := 0; index < len(takes) && index < len(gives); index++ {
			if found := lowering.optionalWidened(
				lowering.checker.GetTypeOfSymbol(gives[index]),
				lowering.checker.GetTypeOfSymbol(takes[index]),
				nil,
				visited,
			); found != nil {
				return found
			}
		}
	}
	return nil
}

func (lowering *lowering) optionalValue(node *ast.Node, target *checker.Type) *optionalWidening {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// The literal's fields, spreads and elements are checked separately at their own sites.
		return nil
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
			if found := lowering.optionalValue(branch, target); found != nil {
				return found
			}
		}
		return nil
	case ast.KindArrowFunction:
		// An arrow's expression body is the value it returns, judged as the return's value, so a fresh
		// literal there is a fresh literal. Its parameters are judged as any function's are.
		body := node.Body()
		sources, targets := lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeAtLocation(node),
			checker.SignatureKindCall,
		), lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
		if body == nil ||
			body.Kind == ast.KindBlock ||
			len(sources) != 1 ||
			len(targets) != 1 ||
			len(sources[0].TypeParameters()) != 0 ||
			len(targets[0].TypeParameters()) != 0 {
			break
		}
		if found := lowering.optionalValue(body, lowering.checker.GetReturnTypeOfSignature(targets[0])); found != nil {
			return found
		}
		takes, gives := sources[0].Parameters(), targets[0].Parameters()
		for index := 0; index < len(takes) && index < len(gives); index++ {
			if found := lowering.optionalWidened(
				lowering.checker.GetTypeOfSymbol(gives[index]),
				lowering.checker.GetTypeOfSymbol(takes[index]),
				nil,
				map[[2]*checker.Type]bool{},
			); found != nil {
				return found
			}
		}
		return nil
	}
	return lowering.optionalWidened(lowering.checker.GetTypeAtLocation(node), target, nil, map[[2]*checker.Type]bool{})
}

// optionalAtSite uses the same relation sites as the other view rules, before lowering.
// A leading spread is not proof of absence: it copies fields hidden by the source's type too.
func (lowering *lowering) optionalAtSite(node *ast.Node) *optionalWidening {
	if lowering.nodeFSFileReadOnlyArgument(node) {
		return nil
	}
	// Only argument positions of a directly imported node:* call can use the host exemption.
	argumentSite := node
	for argumentSite.Parent != nil && argumentSite.Parent.Kind == ast.KindParenthesizedExpression {
		argumentSite = argumentSite.Parent
	}
	if parent := argumentSite.Parent; parent != nil &&
		parent.Kind == ast.KindCallExpression &&
		lowering.optionalNodeHostCall(parent) {
		for index, argument := range parent.AsCallExpression().Arguments.Nodes {
			if ast.SkipParentheses(argument) == node && lowering.nodeHostConsumesArgument(parent, index) {
				return nil
			}
		}
	}
	var found *optionalWidening
	switch node.Kind {
	case ast.KindAsExpression:
		// The cast relates its operand to its own type, and its own type then goes into its slot as a
		// named value's does: f(wide as Node) into a Positioned hides name as const node = wide as Node;
		// f(node) does (#8v8112n). x as const is its operand's own literal type, and its operand is judged
		// as itself.
		as := node.AsAsExpression()
		found = lowering.optionalValue(as.Expression, lowering.checker.GetTypeAtLocation(node))
		if found == nil && !(as.Type.Kind == ast.KindTypeReference &&
			ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) &&
			as.Type.AsTypeReferenceNode().TypeName.Text() == "const") {
			found = lowering.optionalInSlot(node)
		}
	case ast.KindShorthandPropertyAssignment:
		target := lowering.slotType(node.Parent)
		if target != nil {
			found = lowering.optionalValue(
				node.Name(),
				lowering.checker.GetTypeOfPropertyOfType(target, node.Name().Text()),
			)
		}
	case ast.KindSpreadAssignment:
		target := lowering.slotType(node.Parent)
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
			found = lowering.optionalWidened(
				lowering.checker.GetTypeAtLocation(expression),
				target,
				skip,
				map[[2]*checker.Type]bool{},
			)
		}
	default:
		if node.Kind == ast.KindParenthesizedExpression || !lowering.isExpression(node) {
			return nil
		}
		found = lowering.optionalInSlot(node)
	}
	return found
}

// optionalInSlot judges a value against the slot it goes into: the one written for it, or the one tsc
// implied (impliedTarget). A generic function is instantiated to its slot rather than seen through it.
func (lowering *lowering) optionalInSlot(node *ast.Node) *optionalWidening {
	if lowering.genericFunction(node) {
		return nil
	}
	var target *checker.Type
	if viewSite(node) {
		target = lowering.slotType(node)
	}
	if target == nil {
		target = lowering.impliedTarget(node)
	}
	if target == nil {
		return nil
	}
	return lowering.optionalValue(node, target)
}

func (lowering *lowering) refuseOptionalWidening(node *ast.Node) error {
	found := lowering.optionalAtSite(node)
	if found == nil {
		return nil
	}
	return lowering.optionalWideningRefusal(node, found)
}

// optionalWideningRefusal is the hiding rule's refusal of found at node.
func (lowering *lowering) optionalWideningRefusal(node *ast.Node, found *optionalWidening) error {
	return &Refused{Where: lowering.program.Where(node),
		What: "optional property " +
			found.property +
			" in " +
			lowering.checker.TypeToString(found.target) +
			" absent from structural source " +
			lowering.checker.TypeToString(found.source) +
			", which can hide fields",
		Fix: "declare " +
			found.property +
			" on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"}
}

func init() {
	refusalRules.register(3300, "refuseOptionalWidening", (*lowering).refuseOptionalWidening)
}

// Resolve the import binding rather than trusting the callee's spelling.
func (lowering *lowering) optionalNodeHostCall(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression {
		callee = ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	}
	symbol := lowering.checker.GetSymbolAtLocation(callee)
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
