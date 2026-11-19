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
	// handle is an assignment handle's failed relation, which can be a conflicting
	// declared field rather than an absent optional field.
	handle error
	// unstored names the library interface when source declares property only as a lib-declared member
	// the runtime shape lacks (libraryUnstoredMember): a view naming it reads it absent natively where Node
	// reads its value. That's a gap in what the runtime builds, so it's NotYet, not the hiding rule's
	// refusal. It's "" otherwise.
	unstored string
	// constrained says the walk reached property through a type parameter's or another instantiable
	// type's constraint, where the value may hold it under a key the checker never names, so no
	// relation view admits it (recordRelationView).
	constrained bool
}

// constrainedWidening marks found, when there is one, as reached through a constraint.
func constrainedWidening(found *optionalWidening) *optionalWidening {
	if found != nil {
		found.constrained = true
	}
	return found
}

// libraryStoredMembers names, for each library interface whose values the runtime makes, the members
// its runtime shape holds a slot for. Error's shape (exceptions.c) holds name and message; lib also
// declares stack, an own property of every Error on Node, and cause, which it holds when given one.
// Error is the live case (#sn788gx: a view { readonly stack?: string } over a real Error read none
// natively where Node printed the stack). #sn788gx's census of every runtime shape fills in the rest.
var libraryStoredMembers = map[string]map[string]bool{
	"Error": {"name": true, "message": true},
}

// libraryServedMethods names, for each library interface, the methods a program view reads correctly on
// the runtime's value though they live on its prototype: each one measured, never assumed. A RegExp's
// test, called or tested for presence through a program view, agrees with Node in the review
// lane (internal/oracle/testdata/review/agree/fxspptb_8e67677_structural_method.a and
// _structural_optional.a). Any other lib method a program view names is NotYet.
var libraryServedMethods = map[string]map[string]bool{
	"RegExp": {"test": true},
}

// libraryUnstoredMember names the library interface when member is a lib-declared member the runtime
// shape lacks: one a library interface in libraryStoredMembers declares, whose runtime shape holds no
// slot for it, or a method any library interface declares, which lives on the prototype and never in a
// runtime shape (#sn788gx's census: a Date seen as { readonly getTime?: () => number } read getTime
// absent natively where Node read it present, and one called through { getTime(): number } stopped at
// run time as a compiler bug). Read through that interface's own type, such a member is the inherited
// library member refusal's (a lib declaration is never an own field); read through any other view,
// nothing else stops it. It's "" for any other member.
func libraryUnstoredMember(member *ast.Symbol) string {
	owner, _ := libraryUnstoredMemberOf(member)
	return owner
}

// libraryUnstoredMemberOf is libraryUnstoredMember, and says whether the member is unstored because it's
// a method (which only a view the program declares reads as an own field: the library's own types, such
// as String.prototype.replace's searchValue or a Set's ReadonlySetLike, are lowering's library dispatch).
func libraryUnstoredMemberOf(member *ast.Symbol) (string, bool) {
	if member == nil {
		return "", false
	}
	for _, declaration := range member.Declarations {
		owner := declaration.Parent
		if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			continue
		}
		if stored, known := libraryStoredMembers[owner.Name().Text()]; known && !stored[member.Name] {
			return owner.Name().Text(), false
		}
		if declaration.Kind == ast.KindMethodSignature && !libraryServedMethods[owner.Name().Text()][member.Name] {
			return owner.Name().Text(), true
		}
	}
	return "", false
}

// programDeclared reports whether a view's member is declared by the program: some declaration of it sits
// outside TypeScript's library, Adamic's prelude (which redeclares lib's ReadonlySetLike, its has a
// property) and Node's.
func programDeclared(member *ast.Symbol) bool {
	for _, declaration := range member.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if !load.IsLibrary(file) && !load.IsPrelude(file) && !load.IsNodeLibrary(file) {
			return true
		}
	}
	return false
}

// methodViewOwnedElsewhere reports whether another rule already stops a value of type source going where
// the program view target is held, in its own words, so the method arm would be a second stop on the same
// road: a registered host object's view is the host view walk's (nodeHostViewReach, which stops a member
// no registration names: an Intl.Collator seen as { compare: F }, #zjndrx5's pins), and an array, Map or
// Set going where a plain object type is held is shape_fit.go's (#nw6d7ws: a Set seen as { has(x: string):
// boolean }). It defers only when that rule judges this very pair a stop, so it can't fail open.
func (l *lowering) methodViewOwnedElsewhere(source, target *checker.Type) bool {
	if _, host := l.nodeHostObject(source); host {
		return true
	}
	stop, judged := l.shapeFit(source, target, false)
	return judged && stop != ""
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
	// An intersection tsc reduces to never, whose discriminants conflict, holds no value, so it has no
	// field to hide, as never itself doesn't (the exemption below): transformers/ts.ts:1054's
	// ParameterPropertyDeclaration field. Only a reduction tsc has already computed is read. The
	// assignability check above computes it for every source it relates, and a source it didn't is judged
	// as an inhabited one, which fails closed.
	if source.Flags()&checker.TypeFlagsIntersection != 0 && source.ObjectFlags()&checker.ObjectFlagsIsNeverIntersectionComputed != 0 && source.ObjectFlags()&checker.ObjectFlagsIsNeverIntersection != 0 {
		return nil
	}
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
		return constrainedWidening(l.optionalWidened(l.checker.GetBaseConstraintOfType(source), target, skip, visited))
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return constrainedWidening(l.optionalWidened(source, l.checker.GetBaseConstraintOfType(target), skip, visited))
	}
	if target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	// Sources without Object or Intersection flags that need the walk are object,
	// unknown and instantiable types without a different base constraint.
	// Keep never and nullish values exempt: they have no fields to hide. Keep any
	// exempt as a host value or a hatch other rules judge, not a structural source.
	// Keep every primitive exempt. Seeing a primitive as an object type is a
	// representation question owned by #thrpakr (NotYet there, brands exempt).
	// A phantom brand's optional member never exists on the primitive it holds.
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		if source.Flags()&checker.TypeFlagsAny != 0 {
			return nil
		}
		if source.Flags()&(checker.TypeFlagsNever|checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			return nil
		}
		if source.Flags()&checker.TypeFlagsPrimitive != 0 {
			return nil
		}
		if source.Flags()&checker.TypeFlagsInstantiable != 0 {
			// The recursion asks again whether the constraint is assignable to target, and reads nil when it
			// isn't. That's safe: tsc relates an instantiable source to an object target through its
			// constraint, so an accepted source whose constraint isn't assignable was related another way
			// (to itself, or through a type parameter), which the arms above judge.
			if constraint := l.checker.GetBaseConstraintOfType(source); constraint != nil && constraint != source {
				return constrainedWidening(l.optionalWidened(constraint, target, skip, visited))
			}
		}
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
			if property.Flags&ast.SymbolFlagsOptional != 0 && !l.classOptionalAbsent(source, property.Name) {
				return &optionalWidening{property: property.Name, source: source, target: target}
			}
			continue
		}
		if libraryServedRegExpTest(declared) {
			l.regexpServedView = true
		}
		// A view the program declares, naming a member the source has only as a lib-declared member the
		// runtime shape lacks, would read it absent natively. A view that names it as the same library
		// member is read through that member, which the inherited library member refusal stops.
		if owner, method := libraryUnstoredMemberOf(declared); owner != "" && libraryUnstoredMember(property) == "" && (!method || programDeclared(property) && !l.methodViewOwnedElsewhere(source, target)) {
			name := property.Name
			if strings.HasPrefix(name, ast.InternalSymbolNamePrefix) {
				// A well-known symbol's key is internal and per program; name it as the source spells it.
				name = l.checker.SymbolToString(property)
			}
			return &optionalWidening{property: name, source: source, target: target, unstored: owner}
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
	case ast.KindBinaryExpression:
		if node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			// Assignment publishes its value through every target handle, so it is no longer fresh.
			return l.optionalAssignmentValue(node, target)
		}
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

// relationSite is the top-level relation optionalAtSite judged, when it's one value entering one slot:
// the node whose lowered value enters it, and the slot's type there. A relation view keys its entry by
// that node and views the value as that type (relation_views.go).
type relationSite struct {
	value  *ast.Node
	target *checker.Type
}

// optionalAtSite uses the same relation sites as the other view rules, before lowering.
// A leading spread is not proof of absence: it copies fields hidden by the source's type too.
// It gives the relation it judged as a relationSite where that's one value entering one slot, and nil
// where it isn't: a cast's operand related to the cast's own type (castRefusal's to judge), or a value
// optionalValue looked inside (a conditional's arms, an arrow's body, an assignment used as a value).
func (l *lowering) optionalAtSite(node *ast.Node) (*optionalWidening, *relationSite) {
	if l.nodeFSFileReadOnlyArgument(node) {
		return nil, nil
	}
	// Only argument positions of a directly imported node:* call can use the host exemption.
	argumentSite := node
	for argumentSite.Parent != nil && argumentSite.Parent.Kind == ast.KindParenthesizedExpression {
		argumentSite = argumentSite.Parent
	}
	if parent := argumentSite.Parent; parent != nil && parent.Kind == ast.KindCallExpression && l.optionalNodeHostCall(parent) {
		for index, argument := range parent.AsCallExpression().Arguments.Nodes {
			if ast.SkipParentheses(argument) == node && l.nodeHostConsumesArgument(parent, index) {
				return nil, nil
			}
		}
	}
	var found *optionalWidening
	var site *relationSite
	switch node.Kind {
	case ast.KindAsExpression:
		// The cast relates its operand to its own type, and its own type then goes into its slot as a
		// named value's does: f(wide as Node) into a Positioned hides name as const node = wide as Node;
		// f(node) does (#8v8112n). x as const is its operand's own literal type, and its operand is judged
		// as itself.
		as := node.AsAsExpression()
		found = l.optionalValue(as.Expression, l.checker.GetTypeAtLocation(node))
		if found == nil && !(as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) && as.Type.AsTypeReferenceNode().TypeName.Text() == "const") {
			var slot *checker.Type
			found, slot = l.optionalInSlot(node)
			site = &relationSite{value: node, target: slot}
		}
	case ast.KindShorthandPropertyAssignment:
		target := l.slotType(node.Parent)
		if target != nil {
			slot := l.checker.GetTypeOfPropertyOfType(target, node.Name().Text())
			found = l.optionalValue(node.Name(), slot)
			site = &relationSite{value: node.Name(), target: slot}
		}
	case ast.KindSpreadAssignment:
		target := l.slotType(node.Parent)
		if target == nil {
			return nil, nil
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
			site = &relationSite{value: expression, target: target}
		}
	default:
		if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
			return nil, nil
		}
		var slot *checker.Type
		found, slot = l.optionalInSlot(node)
		if optionalValueItself(node) {
			site = &relationSite{value: node, target: slot}
		}
	}
	if found == nil || site == nil || site.target == nil {
		return found, nil
	}
	return found, site
}

// optionalValueItself says whether optionalValue judges node as the value itself, by its own type,
// rather than looking inside it: a conditional's arms, an arrow's body and parameters, or the handles
// of an assignment used as a value. A literal is judged at its own fields' sites. A compound or logical
// assignment used as a value publishes through its target too, so it's no one value either.
func optionalValueItself(node *ast.Node) bool {
	switch skipped := ast.SkipParentheses(node); skipped.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindConditionalExpression, ast.KindArrowFunction:
		return false
	case ast.KindBinaryExpression:
		return !ast.IsAssignmentOperator(skipped.AsBinaryExpression().OperatorToken.Kind)
	}
	return true
}

// optionalInSlot judges a value against the slot it goes into: the one written for it, or the one tsc
// implied (impliedTarget), and gives that slot's type. A generic function is instantiated to its slot
// rather than seen through it.
func (l *lowering) optionalInSlot(node *ast.Node) (*optionalWidening, *checker.Type) {
	if l.genericFunction(node) {
		return nil, nil
	}
	var target *checker.Type
	if viewSite(node) {
		target = l.slotType(node)
	}
	if target == nil {
		target = l.impliedTarget(node)
	}
	if target == nil {
		return nil, nil
	}
	return l.optionalValue(node, target), target
}

// refuseOptionalWidening is rulings.md:71's hiding rule, and in .ts it's rulings.md:72's road past it
// (#0b9xqhh): a relation the checker accepts whose only failure here is an optional field of the target
// absent from a structural source isn't refused. recordRelationView records it, and the value crosses as
// an untagged checked view of the slot's type, registered as a structural cast's view is
// (relationViewValue), so every read of a field through that type is checked: absent, or holding the
// field's type, else a panic. That rests on #80cs7rz's premise: a value admitted by a view stays checked
// under every type it's later seen as, and any road where the flow loses the view's origin is checked or
// NotYet program-wide, never read raw (checkViewWidenedReads). A relation no lowering took a view for is
// refused here once lowering is done, in these words (checkRelationViewsTaken). A .a file always refuses.
func (l *lowering) refuseOptionalWidening(node *ast.Node) error {
	found, site := l.optionalAtSite(node)
	if found == nil {
		return nil
	}
	if l.recordRelationView(node, found, site) {
		return nil
	}
	return l.optionalWideningRefusal(node, found)
}

// optionalWideningRefusal is the hiding rule's refusal of found at node.
func (l *lowering) optionalWideningRefusal(node *ast.Node, found *optionalWidening) error {
	if found.handle != nil {
		return found.handle
	}
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
