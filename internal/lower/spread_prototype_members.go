package lower

import (
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A spread copies only own enumerable data, and a class's getter, setter or method lives on its
// prototype, so { ...instance } never holds one. The checker still types the copy with such a member
// when the source is seen through a type that declares it in any form, a field, a method signature, a
// getter or setter signature or an object literal's accessor, or through a type parameter, which types
// the copy as what it stands for, every member included. So every prototype member the copy's type still
// claims is a lie about the copy, whatever form the claim takes, and by Adamic's amendment of ruling 91
// (#vjvcnnj, her ruling (1)) an Adamic class stays admitted unless its prototype holds a member the
// spread's copy is typed with, and such a spread is refused in .a, naming the class, the member and the
// fix. A method or accessor a class itself declares is one the checker leaves out of a spread's type
// (isSpreadableProperty), so a view that is a class type claims none of those. In .ts the spread stays
// and a read of the missing member stops instead (spread_copy_reads.go).
//
// What is judged is the type the program holds the copy as (Adamic's ruling (1b)), never the spread's
// source type: a copy's reads go through that type or through a hatch, and hatches are checked views that
// stop at the read, so a prototype member the held type doesn't claim can't be read off the copy without a
// check. The held type is the literal's contextual type where the expression's type is that context (a
// declared slot, a parameter, a return), and otherwise the literal's own type, which keeps the claim of
// every member the spread copied: under `satisfies` the expression keeps the literal's own type, though
// the checker gives the literal the target as its context (Language's must on (1b)). A contextual type
// that is or contains a type parameter, or one the checker can't resolve, falls back to the literal's own
// type, so seeing the copy as a narrower type only relaxes the rule where that type is concrete. A union
// held type is judged arm by arm, refused if any arm claims a prototype member the reaching class holds.
//
// It runs before finishAccessors, whose stop at such a read (spreadCopyRead) would otherwise answer
// first for a program the language refuses at the spread.

// spreadClaimedMember is a member a class reaching a spread holds on its prototype, under a name the
// spread's copy is typed with.
type spreadClaimedMember struct {
	// class reaches the spread; holder declares the member, the class itself or one of its bases.
	class, holder, name, kind string
	// claim is how the copy's held type comes to claim name, said after "a spread of <source>, ".
	claim string
}

// viewPart is one constituent of the type a spread's copy is held as, with the names it claims.
type viewPart struct {
	names []string
	// claim is spreadClaimedMember's, for a name of this part.
	claim func(name string) string
}

// checkSpreadPrototypeMembers refuses, in .a, each object spread whose copy is typed with a member that a
// class reaching it holds on its prototype (spreadViewParts). The closed world is every program class
// instance an expression in the modules has, as librarySpreadCandidates finds library types. A spread
// whose producers the allocation graph proves are all classless literals, accessor literals or
// instances of classes holding none of the copy's names on their prototypes reaches no such class.
func (l *lowering) checkSpreadPrototypeMembers(modules []*ast.SourceFile) error {
	var classes map[*checker.Type]string
	var graph *allocationFlowGraph
	for _, spread := range l.librarySpreads {
		if l.firstLinkSkips(spread.node) || !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(spread.node)), ".a") {
			continue
		}
		if classes == nil {
			classes = l.spreadClassCandidates(modules)
		}
		parts := l.spreadViewParts(spread.held, spread.heldInContext, spread.written, spread.mapper)
		if spread.heldOwnToo != nil {
			parts = append(parts, l.spreadViewParts(spread.heldOwnToo, false, spread.written, spread.mapper)...)
		}
		// A name the literal itself writes after its spread is judged too, though the copy holds it as its
		// own: `{ ...built, weight: 5 }` over a class holding weight as a setter on its prototype still ends
		// natively on the compiler bug, past #k7g6f9h's spread_with lowering (override_field_setter_a), so
		// the refusal stands for it until that lowers (Diagnostics' should on #vjvcnnj, left out for now).
		member, found := l.spreadPrototypeMember(spread.view, parts, classes)
		if !found {
			continue
		}
		// Built only once a type might refuse, so a program with no such spread builds no graph this
		// early. An Adamic class instance is proof here only when its chain holds none of the copy's
		// names on its prototype, unlike the library rule's, where every one is.
		if graph == nil {
			graph = l.flowGraph(l.result)
		}
		fields := map[string]bool{}
		for _, part := range parts {
			for _, name := range part.names {
				fields[name] = true
			}
		}
		if graph.plainSpreadSources(spread.source, map[int]bool{}, graph.prototypeFree(fields)) {
			continue
		}
		inherited := ""
		if member.holder != member.class {
			inherited = ", inherited from " + member.holder
		}
		return &Refused{
			Where: l.stopWhere(stopSite{node: spread.node, function: -1, statement: -1}),
			What:  "a spread of " + sourceExpression(spread.node.AsSpreadAssignment().Expression) + ", " + member.claim + ", while class " + member.class + ", which may reach it, holds " + member.name + " as a " + member.kind + " on its prototype" + inherited + ": a spread copies own properties only, so the copy has no " + member.name + " though its type says it does",
			Fix:   "copy the fields you mean explicitly, or declare " + member.name + " a field of " + member.holder,
		}
	}
	return nil
}

// spreadClassCandidates is every instance type of a class the program declares that some expression in
// the modules has, with its class's name. A library class is the library rule's (librarySpreadRefusal).
func (l *lowering) spreadClassCandidates(modules []*ast.SourceFile) map[*checker.Type]string {
	candidates := map[*checker.Type]string{}
	var note func(*checker.Type)
	note = func(proven *checker.Type) {
		if proven == nil {
			return
		}
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				note(member)
			}
			return
		}
		symbol := proven.Symbol()
		if !isClassInstance(proven) || symbol == nil || libraryDeclared(symbol) {
			return
		}
		candidates[proven] = symbol.Name
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsExpression(node) {
			note(l.checker.GetTypeAtLocation(node))
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return candidates
}

// spreadHeld is the type a spread's literal is held as (ruling (1b)), and whether it is the literal's
// contextual type: that, when the literal's place gives the expression its context's type
// (spreadContextHeld) and the context is concrete, and the literal's own type otherwise. A contextual type
// is read as slotType reads it, at the literal. ownToo is the literal's own type where the program holds
// the copy as that as well as its context's, and nil otherwise.
func (l *lowering) spreadHeld(literal *ast.Node) (held *checker.Type, inContext bool, ownToo *checker.Type) {
	own := l.checker.GetTypeAtLocation(literal)
	contextHeld, alsoOwn := spreadContextHeld(literal)
	if !contextHeld {
		return own, false, nil
	}
	context := l.slotType(literal)
	if context == nil || context.Flags()&checker.TypeFlagsAny != 0 || l.mentionsTypeParameter(context, map[*checker.Type]bool{}) {
		return own, false, nil
	}
	if alsoOwn {
		return context, true, own
	}
	return context, true, nil
}

// spreadContextHeld reports whether a literal's place makes the expression's type its contextual type:
// through parentheses, a conditional's branch, a logical operator's operand, a comma's last operand, an
// array literal's element or an object literal's property value, it is written into a declared slot (an
// annotated variable, property or parameter, or an assignment's target), passed as an argument, or
// returned from a function whose return is held (spreadReturnHeld). Anywhere else the expression keeps
// the literal's own type: `satisfies` checks it against its target and leaves it the literal's, and an
// `as` is a cast the rule doesn't take as a slot. alsoOwn says the program holds the copy as the
// literal's own type too: an assignment is an expression whose type is its right side's own, so
// `const value = (held = { ...built })` holds the copy as held's type in held and as its own in value.
// Only where nothing reads the assignment's value (assignmentDiscarded) is the target's type the one
// holder; the own type is judged beside it otherwise, since the target still holds the copy and may claim
// an optional member the own type leaves out.
func spreadContextHeld(literal *ast.Node) (held, alsoOwn bool) {
	child := literal
	for parent := literal.Parent; parent != nil; child, parent = parent, parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().Condition == child {
				return false, false
			}
		case ast.KindBinaryExpression:
			switch parent.AsBinaryExpression().OperatorToken.Kind {
			case ast.KindEqualsToken:
				if parent.AsBinaryExpression().Right != child {
					return false, false
				}
				return true, !assignmentDiscarded(parent)
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			case ast.KindCommaToken:
				if parent.AsBinaryExpression().Right != child {
					return false, false
				}
			default:
				return false, false
			}
		case ast.KindPropertyAssignment:
			if parent.Initializer() != child {
				return false, false
			}
		case ast.KindVariableDeclaration, ast.KindPropertyDeclaration, ast.KindParameter:
			return parent.Type() != nil && parent.Initializer() == child, false
		case ast.KindCallExpression:
			return parent.AsCallExpression().Expression != child, false
		case ast.KindNewExpression:
			return parent.AsNewExpression().Expression != child, false
		case ast.KindReturnStatement:
			return spreadReturnHeld(ast.GetContainingFunction(parent))
		case ast.KindArrowFunction:
			if parent.AsArrowFunction().Body != child {
				return false, false
			}
			return spreadReturnHeld(parent)
		default:
			return false, false
		}
	}
	return false, false
}

// spreadReturnHeld reports whether what function returns is held as its contextual return type: when its
// return type is annotated, or when the function itself sits in one of spreadContextHeld's places, so the
// program holds it as its context's function type. Otherwise the function keeps its own type, whose
// return is the returned expression's own: `(() => { return { ...built }; }) satisfies () => Light` gives
// the literal Light as its context, while make() is still typed as the copy's own type (Diagnostics' must
// on #vjvcnnj, satisfies_return_setter_a). alsoOwn is spreadContextHeld's, for the function: one held
// both ways returns the copy as its own type too.
func spreadReturnHeld(function *ast.Node) (held, alsoOwn bool) {
	if function == nil {
		return false, false
	}
	if function.Type() != nil {
		return true, false
	}
	return spreadContextHeld(function)
}

// assignmentDiscarded reports whether nothing reads an assignment's value: through parentheses it is the
// whole of an expression statement, or a for loop's initializer or incrementor.
func assignmentDiscarded(assignment *ast.Node) bool {
	child := assignment
	parent := assignment.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child, parent = parent, parent.Parent
	}
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindExpressionStatement:
		return true
	case ast.KindForStatement:
		return parent.AsForStatement().Initializer == child || parent.AsForStatement().Incrementor == child
	}
	return false
}

// spreadViewParts is what a spread's copy is claimed as, constituent by constituent of the type it is held
// as (held, as the source writes it), each part as the instantiation being lowered makes it; inContext says
// held is the literal's contextual type rather than its own. A union types the copy with any member's
// names, which a narrowing reaches, so each member is a part of its own, an arm: the union's properties
// would be only the names every member has. undefined and null add nothing to a copy. Held as its own
// type, a spread of a union of one type with undefined or null (source) is that type made partial (the
// checker's tryMergeUnionOfObjectTypeAndEmptyObject), so its copy claims no name present and has no part.
// A type parameter types the copy as itself, so every member of what it stands for is claimed, a method
// and an accessor included; any other part claims the names its type declares in any form the copy keeps
// (copyClaim). mapper is the spread's instantiation's, as librarySpread keeps it.
func (l *lowering) spreadViewParts(held *checker.Type, inContext bool, source *checker.Type, mapper *typeMapper) []viewPart {
	if held == nil {
		return nil
	}
	concrete := func(proven *checker.Type) *checker.Type {
		if mapper == nil {
			return proven
		}
		return instantiateType(l.checker, proven, mapper)
	}
	if !inContext && source != nil && source.Flags()&checker.TypeFlagsUnion != 0 {
		nullish, others := false, 0
		for _, member := range source.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				nullish = true
			} else {
				others++
			}
		}
		if nullish && others == 1 {
			return nil
		}
	}
	heldShown := l.checker.TypeToString(concrete(held))
	arms := 0
	if held.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range held.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) == 0 {
				arms++
			}
		}
	}
	var parts []viewPart
	var walk func(*checker.Type)
	walk = func(written *checker.Type) {
		if written.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
			return
		}
		if written.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range written.Types() {
				walk(member)
			}
			return
		}
		if written.Flags()&checker.TypeFlagsTypeParameter != 0 {
			parameter := l.checker.TypeToString(written)
			stands := []*checker.Type{concrete(written)}
			if stands[0].Flags()&checker.TypeFlagsUnion != 0 {
				stands = stands[0].Types()
			}
			for _, standing := range stands {
				if standing.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
					continue
				}
				var names []string
				for _, property := range l.checker.GetPropertiesOfType(standing) {
					names = append(names, property.Name)
				}
				shown := l.checker.TypeToString(standing)
				parts = append(parts, viewPart{names: names, claim: func(name string) string {
					return "whose copy's own type is " + parameter + ", which stands for " + shown + " here, so the copy is typed with " + name
				}})
			}
			return
		}
		instantiated := concrete(written)
		var names []string
		forms := map[string]string{}
		for _, property := range l.checker.GetPropertiesOfType(instantiated) {
			if form, claimed := copyClaim(property); claimed {
				names = append(names, property.Name)
				forms[property.Name] = form
			}
		}
		armShown := l.checker.TypeToString(instantiated)
		parts = append(parts, viewPart{names: names, claim: func(name string) string {
			switch {
			case !inContext:
				return "whose copy's own type declares " + name + " " + forms[name]
			case arms > 1:
				return "whose copy is held as " + heldShown + ", whose arm " + armShown + " declares " + name + " " + forms[name]
			}
			return "whose copy is held as " + heldShown + ", which declares " + name + " " + forms[name]
		}})
	}
	walk(held)
	return parts
}

// spreadPrototypeMember is the first, by class then member name, of the prototype members a candidate
// class assignable to view, the spread's source type, holds under a name a part of the held type claims.
// The candidate need not be assignable to the part: the part is how the copy is held, not what reaches
// the spread, and a union is refused if any arm claims such a member.
func (l *lowering) spreadPrototypeMember(view *checker.Type, parts []viewPart, classes map[*checker.Type]string) (spreadClaimedMember, bool) {
	var found []spreadClaimedMember
	for candidate, class := range classes {
		if !l.checker.IsTypeAssignableTo(candidate, view) {
			continue
		}
		for _, part := range parts {
			if len(part.names) == 0 {
				continue
			}
			for _, name := range part.names {
				holder, kind := prototypeHeld(l.checker.GetPropertyOfType(candidate, name))
				if holder == nil {
					continue
				}
				member := spreadClaimedMember{class: class, holder: class, name: name, kind: kind, claim: part.claim(name)}
				if !slices.Contains(candidate.Symbol().Declarations, holder) {
					member.holder = "an anonymous class"
					if holderName := holder.Name(); holderName != nil {
						member.holder = holderName.Text()
					}
				}
				found = append(found, member)
			}
		}
	}
	if len(found) == 0 {
		return spreadClaimedMember{}, false
	}
	// Stable, so of one class and name met through several parts the first part's claim is said.
	sort.SliceStable(found, func(i, j int) bool {
		if found[i].class != found[j].class {
			return found[i].class < found[j].class
		}
		return found[i].name < found[j].name
	})
	return found[0], true
}

// copyClaim is the form in which a view's property types a spread's copy with it, as the refusal says
// it: a field (a property signature or declaration, an object literal's property, a parameter property,
// or a mapped type's property with no declaration at all), a method (a method signature, or an object
// literal type's method), a getter, a setter, or a getter and setter. Every form counts, by Adamic's
// ruling (1). It claims nothing only for a method or an accessor a class declares, which the checker
// leaves out of a spread's type. An intersection's property takes any member's form; a union's members
// are each judged on their own (spreadViewParts).
func copyClaim(property *ast.Symbol) (string, bool) {
	if len(property.Declarations) == 0 {
		return "a field", true
	}
	field, method, getter, setter := false, false, false, false
	for _, declaration := range property.Declarations {
		switch declaration.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if declaration.Parent != nil && ast.IsClassLike(declaration.Parent) {
				continue
			}
		}
		switch declaration.Kind {
		case ast.KindMethodDeclaration, ast.KindMethodSignature:
			method = true
		case ast.KindGetAccessor:
			getter = true
		case ast.KindSetAccessor:
			setter = true
		default:
			field = true
		}
	}
	if field {
		return "a field", true
	}
	switch {
	case getter && setter:
		return "a getter and setter", true
	case getter:
		return "a getter", true
	case setter:
		return "a setter", true
	case method:
		return "a method", true
	}
	return "", false
}

// prototypeHeld reads a class instance's property as a member on a class's prototype: a getter, a
// setter or a method, with the class declaration holding it and the member's kind. An own data field,
// and a member no class declares, hold none.
func prototypeHeld(property *ast.Symbol) (*ast.Node, string) {
	if property == nil {
		return nil, ""
	}
	getter, setter, method := false, false, false
	var holder *ast.Node
	for _, declaration := range property.Declarations {
		if declaration.Parent == nil || !ast.IsClassLike(declaration.Parent) {
			continue
		}
		switch declaration.Kind {
		case ast.KindGetAccessor:
			getter = true
		case ast.KindSetAccessor:
			setter = true
		case ast.KindMethodDeclaration:
			method = true
		default:
			continue
		}
		if holder == nil {
			holder = declaration.Parent
		}
	}
	switch {
	case getter && setter:
		return holder, "getter and setter"
	case getter:
		return holder, "getter"
	case setter:
		return holder, "setter"
	case method:
		return holder, "method"
	}
	return nil, ""
}
