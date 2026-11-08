package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// optionalWidening names the relation whose optional field could be hidden by a structural view.
// This is independent of mutable invariance: even a readonly optional field can lie on a read.
type optionalWidening struct {
	property       string
	source, target *checker.Type
	// unstored names the library interface when source declares property only as a lib-declared member
	// the runtime shape lacks (libraryUnstoredMember): a view naming it reads it absent natively where Node
	// reads its value. That's a gap in what the runtime builds, so it's NotYet, not the hiding rule's
	// refusal. It's "" otherwise.
	unstored string
}

// libraryStoredMembers names, for each library interface whose values the runtime makes, the members
// its runtime shape holds a slot for. Error's shape (exceptions.c) holds name and message; lib also
// declares stack, an own property of every Error on Node, and cause, which it holds when given one.
// Error is the live case (#sn788gx: a view { readonly stack?: string } over a real Error read none
// natively where Node printed the stack). #sn788gx's census of every runtime shape fills in the rest.
var libraryStoredMembers = map[string]map[string]bool{
	"Error": {"name": true, "message": true},
}

// libraryUnstoredMember names the library interface when member is a lib-declared member the runtime
// shape lacks: one a library interface in libraryStoredMembers declares, whose runtime shape holds no
// slot for it. Read through that interface's own type, such a member is the inherited library member
// refusal's (a lib declaration is never an own field); read through any other view, nothing else stops
// it. It's "" for any other member.
func libraryUnstoredMember(member *ast.Symbol) string {
	if member == nil {
		return ""
	}
	for _, declaration := range member.Declarations {
		owner := declaration.Parent
		if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			continue
		}
		if stored, known := libraryStoredMembers[owner.Name().Text()]; known && !stored[member.Name] {
			return owner.Name().Text()
		}
	}
	return ""
}

func (l *lowering) optionalWidened(source, target *checker.Type, skip map[string]bool, visited map[[2]*checker.Type]bool) *optionalWidening {
	if source == nil || target == nil {
		return nil
	}
	if weak := l.weakTarget(source); weak != nil {
		source = weak
	}
	if weak := l.weakTarget(target); weak != nil {
		target = weak
	}
	// A Pick of a record holds the whole record (records_partial.go), so none of its fields is hidden.
	if base := l.pickedRecordBase(source); base != nil {
		source = base
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
		// A value enters a union that names its own type as that member, so no other member's field is
		// hidden from it. Seeing the union as another member later is a view of its own, checked there.
		for _, member := range target.Types() {
			if member == source {
				return nil
			}
		}
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
	if (l.isArrayType(source) || checker.IsTupleType(source)) && (l.isArrayType(target) || checker.IsTupleType(target)) {
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
		// A view the program declares, naming a member the source has only as a lib-declared member the
		// runtime shape lacks, would read it absent natively. A view that names it as the same library
		// member is read through that member, which the inherited library member refusal stops.
		if owner := libraryUnstoredMember(declared); owner != "" && libraryUnstoredMember(property) == "" {
			return &optionalWidening{property: property.Name, source: source, target: target, unstored: owner}
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
	case ast.KindArrowFunction:
		// An arrow's expression body is the value it returns, judged as the return's value, so a fresh
		// literal there is a fresh literal. Its parameters are judged as any function's are.
		body := node.Body()
		sources, targets := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall), l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
		if body == nil || body.Kind == ast.KindBlock || len(sources) != 1 || len(targets) != 1 || len(sources[0].TypeParameters()) != 0 || len(targets[0].TypeParameters()) != 0 {
			break
		}
		if found := l.optionalValue(body, l.checker.GetReturnTypeOfSignature(targets[0])); found != nil {
			return found
		}
		takes, gives := sources[0].Parameters(), targets[0].Parameters()
		for index := 0; index < len(takes) && index < len(gives); index++ {
			if found := l.optionalWidened(l.checker.GetTypeOfSymbol(gives[index]), l.checker.GetTypeOfSymbol(takes[index]), nil, map[[2]*checker.Type]bool{}); found != nil {
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
	if l.nodeFSFileReadOnlyArgument(node) {
		return nil
	}
	// Only argument positions of a directly imported node:* call can use the host exemption.
	argumentSite := node
	for argumentSite.Parent != nil && argumentSite.Parent.Kind == ast.KindParenthesizedExpression {
		argumentSite = argumentSite.Parent
	}
	if parent := argumentSite.Parent; parent != nil && parent.Kind == ast.KindCallExpression && l.optionalNodeHostCall(parent) {
		for index, argument := range parent.AsCallExpression().Arguments.Nodes {
			if ast.SkipParentheses(argument) == node && l.nodeHostConsumesArgument(parent, index) {
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
		found = l.optionalValue(as.Expression, l.checker.GetTypeAtLocation(node))
		if found == nil && !(as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) && as.Type.AsTypeReferenceNode().TypeName.Text() == "const") {
			found = l.optionalInSlot(node)
		}
	case ast.KindShorthandPropertyAssignment:
		target := l.slotType(node.Parent)
		if target != nil {
			found = l.optionalValue(node.Name(), l.checker.GetTypeOfPropertyOfType(target, node.Name().Text()))
		}
	case ast.KindSpreadAssignment:
		target := l.slotType(node.Parent)
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
		if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
			return nil
		}
		found = l.optionalInSlot(node)
	}
	return found
}

// optionalInSlot judges a value against the slot it goes into: the one written for it, or the one tsc
// implied (impliedTarget). A generic function is instantiated to its slot rather than seen through it.
func (l *lowering) optionalInSlot(node *ast.Node) *optionalWidening {
	if l.genericFunction(node) {
		return nil
	}
	var target *checker.Type
	if viewSite(node) {
		target = l.slotType(node)
	}
	if target == nil {
		target = l.impliedTarget(node)
	}
	if target == nil {
		return nil
	}
	return l.optionalValue(node, target)
}

func (l *lowering) refuseOptionalWidening(node *ast.Node) error {
	found := l.optionalAtSite(node)
	if found == nil {
		return nil
	}
	return l.optionalWideningRefusal(node, found)
}

// optionalWideningRefusal is the hiding rule's refusal of found at node.
func (l *lowering) optionalWideningRefusal(node *ast.Node, found *optionalWidening) error {
	if found.unstored != "" {
		// Diagnostics' words on #sn788gx, one template for every runtime shape the census adds.
		return l.notYet(node, "a view naming "+found.property+", which lib declares on "+found.unstored+" but the runtime's "+found.unstored+" doesn't hold (leave "+found.property+" out of the view's type)")
	}
	return &Refused{Where: l.program.Where(node),
		What: "optional property " + found.property + " in " + l.checker.TypeToString(found.target) + " absent from structural source " + l.checker.TypeToString(found.source) + ", which can hide fields",
		Fix:  "declare " + found.property + " on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"}
}

func init() {
	refusalRules.register(3300, "refuseOptionalWidening", (*lowering).refuseOptionalWidening)
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
