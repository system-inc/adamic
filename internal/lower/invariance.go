package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Cohere's rules own ordinary type relations. This pair proof remains for
// constrained sources, inferred joins, explicit satisfies proofs, instantiated
// generic writes and override ABI checks. The oracle probes show that removing
// the constrained-source proof reads an undefined value as a declared string.
// RunRule exposes file rules, not a pair-relation API, so these callers cannot
// delegate their concrete or representation-specific pairs to it yet.

// widening is a view refused: the parts of the source and the target at the slot where the target can
// write what the source can't read.
type widening struct {
	source, target *checker.Type

	// readonlyField names a readonly field seen as a writable one, when that's the slot.
	readonlyField string

	// parameter is a function's parameter of type source seen as taking target, which it can't.
	parameter bool
}

// widened finds the first mutable slot at which a value of type from, seen as type to, could be
// written something it can't hold, or returns nil.
func (l *lowering) widened(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	from, to = l.withoutUndefined(from), l.withoutUndefined(to)
	if from == to || visited[[2]*checker.Type{from, to}] {
		return nil
	}
	visited[[2]*checker.Type{from, to}] = true
	if members := l.definedMembers(from); len(members) > 1 {
		// A union is any one of its members, each seen as the target.
		for _, member := range members {
			if found := l.widened(member, to, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if members := l.definedMembers(to); len(members) > 1 {
		// Seen as a union, a value is seen as each member it can be, and narrowing picks any of them
		// to write through.
		for _, member := range members {
			if !l.checker.IsTypeAssignableTo(from, member) {
				continue
			}
			if found := l.widened(from, member, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if l.narrowedFrom(from, to) {
			return nil
		}
		if constraint := l.checker.GetBaseConstraintOfType(to); constraint != nil && l.canWrite(constraint, map[*checker.Type]bool{}) {
			return &widening{source: from, target: to}
		}
		return nil
	}
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		// What it can be is what its constraint allows: seen as anything wider, it's that seen so.
		constraint := l.checker.GetBaseConstraintOfType(from)
		if constraint == nil {
			return nil
		}
		return l.widened(constraint, to, visited)
	}
	if !l.structured(from) || !l.structured(to) {
		// A class instance target is walked as any object is: its fields are slots as a literal's
		// are, so Box<Dog> seen as Box<Animal>, or a plain object seen as a class, is judged by them.
		return nil
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		// A function seen as another is handed the other's arguments, and its results are seen as
		// the other's: each a view of its own.
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			takes, given := l.checker.GetTypeOfSymbol(fromParameters[index]), l.checker.GetTypeOfSymbol(toParameters[index])
			if !l.checker.IsTypeAssignableTo(given, takes) {
				// tsc relates a method's parameters both ways (method bivariance), so a method taking
				// a Dog can be seen as one taking any Animal, and handed a Cat.
				return &widening{source: takes, target: given, parameter: true}
			}
			if found := l.widened(given, takes, visited); found != nil {
				return found
			}
		}
		return l.widened(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]), visited)
	}
	fromContainers := l.containers(from)
	for _, viewed := range l.containers(to) {
		for _, inside := range fromContainers {
			if !l.sameContainer(inside, viewed) {
				continue
			}
			if found := l.widenedElements(inside, viewed, visited); found != nil {
				return found
			}
			break
		}
	}
	if to.Flags()&checker.TypeFlagsObject != 0 && len(l.containers(to)) > 0 {
		// An array, tuple, map or set: its elements are its slots, and its own properties (length,
		// size) are numbers on both sides.
		return nil
	}
	return l.widenedProperties(from, to, nil, false, visited)
}

// widenedProperties walks to's properties against from's, but those skip names: each field a slot
// to can write unless it's readonly, and each method a function view of its own. A copy's fields
// (fresh) are slots of their own, which nothing else writes, so only what they hold is walked.
func (l *lowering) widenedProperties(from *checker.Type, to *checker.Type, skip map[string]bool, fresh bool, visited map[[2]*checker.Type]bool) *widening {
	for _, viewed := range l.checker.GetPropertiesOfType(to) {
		inside := l.checker.GetPropertyOfType(from, viewed.Name)
		if inside == nil || skip[viewed.Name] {
			continue
		}
		source, target := l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)
		if viewed.Flags&ast.SymbolFlagsMethod != 0 || inside.Flags&ast.SymbolFlagsMethod != 0 {
			// A method isn't a slot anything writes, but what it returns and takes are views: the
			// program's own methods. The library's (an Iterable's, an array's own) are left to the
			// rules for the library's types, and a generic one has no one type to compare.
			if len(viewed.Declarations) == 0 || load.IsLibrary(ast.GetSourceFileOfNode(viewed.Declarations[0])) || l.generic(target) || l.generic(source) {
				continue
			}
			if found := l.widened(source, target, visited); found != nil {
				return found
			}
			continue
		}
		if !fresh && !l.checker.IsReadonlySymbol(viewed) && l.checker.IsReadonlySymbol(inside) {
			// tsc ignores readonly when it relates properties, so a readonly field, which may hold
			// something narrower than its type says (it's covariant), can be seen as a writable one,
			// and the wider type written into it. Only the identical type, which this walk never
			// enters, is the same field.
			return &widening{source: source, target: target, readonlyField: viewed.Name}
		}
		if !fresh && !l.checker.IsReadonlySymbol(viewed) && !l.checker.IsTypeAssignableTo(target, source) {
			return &widening{source: source, target: target}
		}
		if found := l.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// generic reports whether a function type has a type parameter of its own.
func (l *lowering) generic(proven *checker.Type) bool {
	for _, signature := range l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}

// widenedElements walks an array's, a tuple's, a map's or a set's type arguments in from against
// to's, each a slot to can write unless it's read only.
func (l *lowering) widenedElements(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	mutable := !l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(to) && to.TargetTupleType().IsReadonly())
	return l.widenedArguments(from, to, mutable, visited)
}

// widenedArguments is widenedElements with whether to's own slots can be written given: a fresh
// container's can't be written through anything else (freshValue).
func (l *lowering) widenedArguments(from *checker.Type, to *checker.Type, mutable bool, visited map[[2]*checker.Type]bool) *widening {
	fromArguments, toArguments := l.checker.GetTypeArguments(from), l.checker.GetTypeArguments(to)
	if checker.IsTupleType(from) && !checker.IsTupleType(to) && len(toArguments) == 1 {
		// A tuple seen as an array: every element of it is seen as the array's element.
		widenedTo := make([]*checker.Type, len(fromArguments))
		for index := range widenedTo {
			widenedTo[index] = toArguments[0]
		}
		toArguments = widenedTo
	}
	for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
		source, target := fromArguments[index], toArguments[index]
		if mutable && !l.checker.IsTypeAssignableTo(target, source) {
			return &widening{source: source, target: target}
		}
		if found := l.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// containers is a type's arrays, tuples, maps and sets: itself, or its members that are, when it's an
// intersection.
func (l *lowering) containers(proven *checker.Type) []*checker.Type {
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsIntersection != 0 {
		members = proven.Types()
	}
	var found []*checker.Type
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsObject != 0 && member.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
			(l.checker.IsArrayType(member) || checker.IsTupleType(member) || l.isLibraryType(member, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
			found = append(found, member)
		}
	}
	return found
}

// sameContainer reports whether two containers are the same kind, whose type arguments pair up: two
// arrays or tuples (readonly or not), two maps, or two sets.
func (l *lowering) sameContainer(from *checker.Type, to *checker.Type) bool {
	switch {
	case l.checker.IsArrayType(to) || checker.IsTupleType(to):
		return l.checker.IsArrayType(from) || checker.IsTupleType(from)
	case l.isLibraryType(to, "Map", "ReadonlyMap"):
		return l.isLibraryType(from, "Map", "ReadonlyMap")
	default:
		return l.isLibraryType(from, "Set", "ReadonlySet")
	}
}

// canWrite reports whether something can be written through a value of a type, at any depth: an
// array, tuple, map or set that isn't read only, or a property that isn't readonly, or one inside a
// readonly one.
func (l *lowering) canWrite(proven *checker.Type, visited map[*checker.Type]bool) bool {
	proven = l.withoutUndefined(proven)
	if proven == nil || visited[proven] || !l.structured(proven) || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 {
		return false
	}
	visited[proven] = true
	containers := l.containers(proven)
	for _, container := range containers {
		if !l.isLibraryType(container, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(container) && container.TargetTupleType().IsReadonly()) {
			return true
		}
		for _, element := range l.checker.GetTypeArguments(container) {
			if l.canWrite(element, visited) {
				return true
			}
		}
	}
	if proven.Flags()&checker.TypeFlagsObject != 0 && len(containers) > 0 {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		if !l.checker.IsReadonlySymbol(property) || l.canWrite(l.checker.GetTypeOfSymbol(property), visited) {
			return true
		}
	}
	return false
}

// narrowedFrom reports whether from is the type parameter to itself, or it narrowed: an intersection
// with it among its members (T & {}, what a check against undefined leaves).
func (l *lowering) narrowedFrom(from *checker.Type, to *checker.Type) bool {
	if from == to {
		return true
	}
	if from.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, member := range from.Types() {
			if member == to {
				return true
			}
		}
	}
	return false
}

// structured reports whether a type has parts to walk: an object or an intersection.
func (l *lowering) structured(proven *checker.Type) bool {
	return proven.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0
}

// withoutUndefined is a type with undefined taken out of it: the one type left, or the union itself
// when several are, whose members widened walks one by one (definedMembers).
func (l *lowering) withoutUndefined(proven *checker.Type) *checker.Type {
	if members := l.definedMembers(proven); len(members) == 1 {
		return members[0]
	}
	return proven
}

// definedMembers is a union's members but undefined, or the type itself when it isn't a union.
func (l *lowering) definedMembers(proven *checker.Type) []*checker.Type {
	if proven == nil {
		return nil
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return []*checker.Type{proven}
	}
	var members []*checker.Type
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined == 0 {
			members = append(members, member)
		}
	}
	return members
}

// isClassInstance is the instance side of a class, generic or not.
func isClassInstance(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if proven.ObjectFlags()&checker.ObjectFlagsClass != 0 {
		return true
	}
	return proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil && proven.Target().ObjectFlags()&checker.ObjectFlagsClass != 0
}

// viewSite reports whether an expression is a value going into a typed slot: an initializer (a
// variable's, a parameter's default, a class field's), the right of an assignment, an argument, a
// returned value, a literal's property (shorthand too) or element, or an arrow's expression body.
// Parentheses around it are the same site.
func viewSite(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		// A destructuring declaration without a type keeps nothing of the value: its names take the
		// types of its parts, and the pattern's own type (its fields any) is no view of it.
		declaration := parent.AsVariableDeclaration()
		if ast.IsBindingPattern(declaration.Name()) && declaration.Type == nil {
			return false
		}
		return declaration.Initializer == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression != node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression != node
	case ast.KindReturnStatement, ast.KindArrayLiteralExpression:
		return true
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == node
	case ast.KindArrowFunction:
		// A block has no value type; its return expressions are separate view sites.
		return node.Kind != ast.KindBlock && parent.AsArrowFunction().Body == node
	}
	return false
}

// Cohere owns ordinary views. These witnesses still require an Adamic proof:
// source constraints and inferred joins that the upstream walker does not relate.
func (l *lowering) refuseWidening(node *ast.Node) error {
	switch node.Kind {
	case ast.KindAsExpression:
		as := node.AsAsExpression()
		if as.Type.Kind == ast.KindTypeReference && as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
			return nil
		}
		source, target := l.checker.GetTypeAtLocation(as.Expression), l.checker.GetTypeAtLocation(node)
		if l.checker.IsTypeAssignableTo(source, target) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target) {
			return l.provenRelation(node, as.Expression, target)
		}
		if found := l.freshOrWidened(as.Expression, source, target); found != nil {
			return localRelationRefusal(l.wideningRefusal(node, source, target, found), "adamic/proven-relation")
		}
		return nil
	case ast.KindSatisfiesExpression:
		satisfies := node.AsSatisfiesExpression()
		return l.provenRelation(node, satisfies.Expression, l.checker.GetTypeAtLocation(satisfies.Type))
	}
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) || l.genericFunction(node) {
		return nil
	}
	own := l.checker.GetTypeAtLocation(node)
	var target *checker.Type
	if viewSite(node) {
		target = l.checker.GetContextualType(node, checker.ContextFlagsNone)
	}
	inferred := l.impliedTarget(node)
	constrained := own.Flags()&checker.TypeFlagsTypeParameter != 0
	if inferred == nil && !constrained {
		return nil
	}
	if target == nil {
		target = inferred
	}
	if target == nil {
		return nil
	}
	if found := l.freshOrWidened(node, own, target); found != nil {
		id := "adamic/invariant-inferred-view"
		if constrained {
			id = "adamic/invariant-constraint-view"
		}
		return localRelationRefusal(l.wideningRefusal(node, own, target, found), id)
	}
	return nil
}

func localRelationRefusal(err error, id string) error {
	if refused, ok := err.(*Refused); ok {
		refused.Fix = strings.ReplaceAll(refused.Fix, "adamic/invariant-mutable", id)
		refused.Fix = strings.ReplaceAll(refused.Fix, "adamic/nominal-class", id)
	}
	return err
}

func (l *lowering) wideningRefusal(node *ast.Node, own, contextual *checker.Type, found *widening) error {
	what := "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read"
	fix := "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
	if found.parameter {
		what = "a function taking " + l.checker.TypeToString(found.source) + " seen as one taking " + l.checker.TypeToString(found.target) + " (tsc relates a method's parameters both ways), so it can be handed what it can't take"
		fix = "write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)"
	}
	if found.readonlyField != "" {
		what = "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", whose readonly field " + found.readonlyField + " becomes writable: a readonly field may hold something narrower than " + l.checker.TypeToString(found.source) + ", which a write of " + l.checker.TypeToString(found.target) + " would replace"
		fix = "keep " + found.readonlyField + " readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
	}
	if found.target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		what = "a value of type " + l.checker.TypeToString(found.source) + " seen as " + l.checker.TypeToString(found.target) + ", a type parameter whose constraint " + l.checker.TypeToString(l.checker.GetBaseConstraintOfType(found.target)) + " can be written, so it can write what " + l.checker.TypeToString(found.source) + " can't hold"
		fix = "take it as " + l.checker.TypeToString(found.source) + ", or constrain " + l.checker.TypeToString(found.target) + " to something readonly, which can't write (adamic/invariant-mutable)"
	}
	return &Refused{Where: l.program.Where(node), What: what, Fix: fix}
}

// freshOrWidened is widened for a value at a site, but a fresh one: a container made right there and
// held by nothing else (freshValue) can't be written through another name, so its own slots are
// covariant, and only what's inside it, held elsewhere too, is walked as a view.
func (l *lowering) freshOrWidened(node *ast.Node, own *checker.Type, contextual *checker.Type) *widening {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// Made as the type it's written into, held by nothing else: its own parts are sites.
		return nil
	case ast.KindConditionalExpression:
		// Either branch is the value: each judged as it is.
		conditional := node.AsConditionalExpression()
		for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
			if found := l.freshOrWidened(branch, l.checker.GetTypeAtLocation(branch), contextual); found != nil {
				return found
			}
		}
		return nil
	}
	visited := map[[2]*checker.Type]bool{}
	if !l.freshValue(node) {
		return l.widened(own, contextual, visited)
	}
	own, contextual = l.withoutUndefined(own), l.withoutUndefined(contextual)
	for _, viewed := range l.containers(contextual) {
		for _, inside := range l.containers(own) {
			if l.sameContainer(inside, viewed) {
				return l.widenedArguments(inside, viewed, false, visited)
			}
		}
	}
	return l.widened(own, contextual, visited)
}

// freshValue reports whether an expression makes a container nothing else holds: a copy (slice,
// filter, map, concat, Array.from) or a new Map, Set or Array.
func (l *lowering) freshValue(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNewExpression:
		callee := node.AsNewExpression().Expression
		return l.isLibraryGlobal(callee, "Map") || l.isLibraryGlobal(callee, "Set") || l.isLibraryGlobal(callee, "Array")
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			return false
		}
		access := callee.AsPropertyAccessExpression()
		method := access.Name().Text()
		if method == "from" {
			return l.isLibraryGlobal(access.Expression, "Array")
		}
		switch method {
		case "slice", "filter", "map", "concat":
			return l.checker.IsArrayType(l.checker.GetTypeAtLocation(access.Expression)) || l.isLibraryType(l.checker.GetTypeAtLocation(access.Expression), "ReadonlyArray")
		}
	}
	return false
}

// isExpression reports whether a node is one a value can be: anything but a statement, a declaration,
// a name being declared or a type.
func (l *lowering) isExpression(node *ast.Node) bool {
	return ast.IsExpression(node) && !ast.IsPartOfTypeNode(node)
}

// impliedTarget is the type tsc gave what a value with no type written for it goes into, or nil: a
// conditional's (or ??'s, ||'s, &&'s) own type for its branches, an array literal's element type for
// its elements, and a function's inferred result for what it returns.
func (l *lowering) impliedTarget(node *ast.Node) *checker.Type {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindConditionalExpression:
		if conditional := parent.AsConditionalExpression(); conditional.WhenTrue == child || conditional.WhenFalse == child {
			return l.checker.GetTypeAtLocation(parent)
		}
	case ast.KindBinaryExpression:
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindQuestionQuestionToken, ast.KindBarBarToken, ast.KindAmpersandAmpersandToken:
			return l.checker.GetTypeAtLocation(parent)
		}
	case ast.KindArrayLiteralExpression:
		literal := l.checker.GetTypeAtLocation(parent)
		if l.checker.IsArrayType(literal) {
			if arguments := l.checker.GetTypeArguments(literal); len(arguments) == 1 {
				return arguments[0]
			}
		}
	case ast.KindReturnStatement, ast.KindArrowFunction:
		if parent.Kind == ast.KindArrowFunction && parent.AsArrowFunction().Body != child {
			return nil
		}
		function := parent
		for function != nil && !ast.IsFunctionLike(function) {
			function = function.Parent
		}
		if function == nil {
			return nil
		}
		return l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(function))
	}
	return nil
}

// genericFunction reports whether an expression is a function with type parameters of its own,
// which tsc instantiates to the type it's written into rather than viewing it through that type.
func (l *lowering) genericFunction(node *ast.Node) bool {
	for _, signature := range l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}
