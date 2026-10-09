package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnusedPrivateClassMembers flags a `private` class member nothing inside the class ever reads.
//
//	valid:   class C { private a = 1; m() { return this.a; } }
//	valid:   class C { private a = 1; m(other: C) { return other.a; } }
//	valid:   class C { private a = 1; m() { const self = this; return self.a; } }
//	valid:   class C { private set a(v) {} m() { this.a = 1; } }      an accessor call is a read
//	invalid: class C { private a = 1; }
//	invalid: class C { private a = 1; m() { this.a = 42; } }          written, never read
//	invalid: class C { constructor(private a: number) {} }
//
// A `private` member is visible only inside the class body that declares it, so one nothing there
// reads is dead weight the compiler will not mention. The `#private` half of the same question is
// the core rule of this name, which is already ported; this rule is the TypeScript keyword.
//
// # This is not a filter over the core rule, and the audit's line count says otherwise
//
// Upstream's file is sixty lines and delegates to `analyzeClassMemberUsage`, an 893-line scope
// analyzer in its shared utilities. So the real size is the analyzer, not the caller, which is the
// same trap the extension rules carry arriving through a different door: a shared helper rather
// than `getESLintCoreRule`.
//
// Measured against the core rule it sits beside, on upstream's own 99-case corpus driven through
// both: they agree on 72 cases and differ on 27, the wrapper producing 49 findings against the
// core's 20. Every one of the 27 mentions the `private` keyword, which the core cannot see at all,
// so there is no overlap to reuse and nothing here narrows the core.
//
// # Resolution replaces upstream's scope analysis, and answers better
//
// Upstream declares `requiresTypeChecking: false` and reaches a use through scope tracking: it
// follows `this` across nested scopes and chases aliases like `const self = this`. We have the
// checker, and `GetSymbolAtLocation` on a property NAME resolves straight to the declaring member.
// Probed on the three shapes upstream needs its analyzer for:
//
//	other.prop     where other: C          resolves to C's PropertyDeclaration
//	self.prop      after const self = this resolves to C's PropertyDeclaration
//	self2.prop     after a second alias    resolves to C's PropertyDeclaration
//
// This is the brief's rule about fidelity being to what a rule DECIDES rather than how it obtains
// what it needs. Reproducing an alias-chasing scope pass to answer a question one resolution call
// answers would be a workaround for a constraint we do not have.
//
// # And it produces a DIFFERENT ANSWER, in one direction, deliberately
//
// Upstream's reach has a hard edge that resolution does not. It follows `this`, ONE hop of
// `const self = this`, and a parameter annotated as the class. It does not follow two hops, and it
// does not follow a local holding an instance. Measured on the installed 8.67.0 build:
//
//	const a = this; a.prop                    clean in both
//	const a = this; const b = a; b.prop       upstream REPORTS, we are clean
//	method(thing: Foo) { return thing.prop }  clean in both, and in upstream's own corpus
//	const other = this.clone(); other.helper() upstream REPORTS, we are clean
//
// The member is genuinely read in every one of those, so upstream's reports are false positives.
// This rule declines to reproduce them, which is the brief's "a port is not obliged to carry a
// defect it can see", and the decline is stated here and pinned by a test rather than left silent.
//
// It is visible on the real tree rather than only in constructed cases: this rule finds three unused
// private members where upstream finds four, and the fourth is a method called eight lines above its
// own declaration through a local receiver. For an unused-thing rule the direction is the one that
// matters, since a false positive asks a reader to delete working code.
//
// # A `declare` field is never unused
//
//	valid:   class C { declare private readonly __brand: 'C'; }
//
// A `declare` field has no runtime existence. Its job is to make a type nominal, so nothing reading
// it is the design rather than dead weight, and deleting it would change what the type checker
// accepts while removing no code. Upstream reports it; this rule does not, deliberately.
//
// # A use outside the class body is not a use
//
// `const c = new C(); c.privateThing;` resolves to the member and upstream still REPORTS it, which
// reads backwards until you notice the compiler already rejects that line. An illegal reference
// cannot keep a member alive, so the search is bounded to the declaring class's own body. Upstream
// gets that from its scope structure; here it is a containment test, and two of its cases pin it.
//
// # Read against write, which is most of the judgment
//
// A member only ever assigned is unused: `this.a = 42` keeps nothing alive. The shapes that count
// as write-only are enumerated rather than delegated to `ast.IsWriteAccess`, which answers "does
// this store" rather than "does this ONLY store" and misses a rest element entirely. That
// enumeration is the same one the core rule of this name arrived at, and it is repeated here rather
// than shared because the two rules live in different packages and the core's helpers are
// unexported. If a third caller ever wants it, that is the moment to lift it onto the shelf.
//
// An accessor is the exception: reading or writing one calls a function body either way, so any
// reference to a `private get` or `private set` is a use. That is decided from what the member IS
// rather than from the reference, and it is why a write-only setter is correct code.
//
// # Cost
//
// One pass per class, and the checker is asked only at a property access whose name matches a
// declared private member by text. A class with no `private` member registers nothing.
var NoUnusedPrivateClassMembers = rule.Rule{
	Name: "@typescript-eslint/no-unused-private-class-members",

	// A use is attributed by resolving the property name to its declaration, which is the checker's
	// question and is what replaces upstream's scope pass.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return rule.Listeners{}
		}

		check := func(classNode *ast.Node, members *ast.NodeList) {
			if members == nil {
				return
			}
			declared := collectPrivateKeywordMembers(members)
			if len(declared) == 0 {
				return
			}
			markPrivateKeywordUses(ctx, classNode, declared)

			// Reported in declaration order, which is source order, so a class with several dead
			// members reads top to bottom.
			for _, member := range declared {
				if member.used {
					continue
				}
				message := rule.Message{
					Id: "unusedPrivateClassMember",
					Description: "Private class member '" + member.name +
						"' is defined but never used.",
				}
				if member.spanEnd == 0 {
					ctx.ReportNode(member.nameNode, message)
					continue
				}
				ctx.ReportRange(core.NewTextRange(
					rule.TokenRange(ctx.SourceFile, member.nameNode).Pos(), member.spanEnd), message)
			}
		}

		return rule.Listeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				check(node, node.AsClassDeclaration().Members)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				check(node, node.AsClassExpression().Members)
			},
		}
	},
}

// privateKeywordMember is one `private` member and whether anything read it.
type privateKeywordMember struct {
	name string

	// nameNode is where the finding points, and where its span starts.
	nameNode *ast.Node

	// spanEnd extends the finding past the name for a parameter property, whose span upstream runs
	// through the type annotation. Zero means the name's own end.
	spanEnd int

	// declaration is the node the checker resolves a use back to, compared by identity so two
	// classes declaring the same private name are never confused.
	declaration *ast.Node

	// isAccessor short-circuits the read test: any reference to a getter or setter calls a body.
	isAccessor bool

	// isHashPrivate marks a `#name` member, which is found by name identity rather than through the
	// checker: a `#name` is lexically unique to its class, so any occurrence of that spelling inside
	// the class IS this member and no resolution is needed. That is also why the two halves cannot
	// share one search: a `private` name is an ordinary property name and needs resolving.
	isHashPrivate bool

	used bool
}

// collectPrivateKeywordMembers finds every member the rule judges, in source order.
//
// Five declaration shapes carry the keyword and a sixth looks like it does. Probed rather than
// listed from the grammar: a property, a method, a getter, a setter, and an `accessor` property all
// take `private` as a modifier on the member, while a constructor parameter takes it as a modifier
// on the PARAMETER and declares a member as a side effect. A `#private` name is a different rule's
// subject and is skipped even though it is also private.
func collectPrivateKeywordMembers(members *ast.NodeList) []*privateKeywordMember {
	declared := []*privateKeywordMember{}
	for _, member := range members.Nodes {
		switch member.Kind {
		case ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
			ast.KindGetAccessor, ast.KindSetAccessor:
			name := member.Name()
			if name == nil || name.Kind == ast.KindComputedPropertyName {
				// A computed key names nothing knowable.
				continue
			}
			// Both kinds of private, which is upstream's `isPrivate() || isHashPrivate()`. A
			// `#name` is private by its spelling and needs no modifier.
			isHashPrivate := name.Kind == ast.KindPrivateIdentifier
			if !isHashPrivate && !hasPrivateKeyword(member) {
				continue
			}
			// A `declare` field emits nothing: no slot, no initializer, no runtime existence at all.
			// It exists to make the class nominal (`declare private readonly __brand: X`), so never
			// being read is its purpose rather than a defect, and there is no code to delete. Upstream
			// reports it (measured on 8.67.0), which is a divergence in our favour recorded in the
			// rule's .md. Only a property can carry the modifier, and never a `#name`, which the
			// compiler rejects, so this one test covers every shape.
			if ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient) {
				continue
			}
			declared = append(declared, &privateKeywordMember{
				name:          name.Text(),
				nameNode:      name,
				declaration:   member,
				isAccessor:    member.Kind == ast.KindGetAccessor || member.Kind == ast.KindSetAccessor,
				isHashPrivate: isHashPrivate,
			})

		case ast.KindConstructor:
			// A parameter property declares a member from the constructor signature. Upstream
			// reports on the parameter itself, so the span covers `name: Type` rather than the name
			// alone, which its corpus pins in three cases.
			for _, parameter := range member.Parameters() {
				if !hasPrivateKeyword(parameter) {
					continue
				}
				name := parameter.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					continue
				}
				declared = append(declared, &privateKeywordMember{
					name: name.Text(),
					// Upstream reports the parameter but its span starts at the NAME, not at the
					// modifiers: measured, `constructor(private readonly p: number = 1)` reports
					// `p: number` and covers neither `private readonly` nor the default. So the
					// finding is a computed range from the name's start to the parameter's end,
					// minus any initializer.
					nameNode:    name,
					spanEnd:     parameterPropertySpanEnd(parameter),
					declaration: parameter,
				})
			}
		}
	}
	return declared
}

// hasPrivateKeyword reports whether a node carries the `private` modifier.
func hasPrivateKeyword(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindPrivateKeyword {
			return true
		}
	}
	return false
}

// markPrivateKeywordUses walks the class body and marks every member something reads.
//
// The walk is bounded to the class node, which is what makes a reference from outside not count:
// such a reference is illegal TypeScript and the compiler rejects it, so it cannot keep a member
// alive. Upstream reaches the same verdict through its scope structure and carries two cases for it.
func markPrivateKeywordUses(ctx rule.Context, classNode *ast.Node, declared []*privateKeywordMember) {
	byName := map[string][]*privateKeywordMember{}
	for _, member := range declared {
		byName[member.name] = append(byName[member.name], member)
	}

	var visit func(*ast.Node, bool)
	visit = func(node *ast.Node, insideNestedClass bool) {
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			name := node.AsPropertyAccessExpression().Name()
			switch {
			case name == nil:
			case name.Kind == ast.KindIdentifier:
				// A nested class is still inside the declaring class's body, so a receiver of the
				// outer type reads the member from there: `class Outer { private secret; make() {
				// return class { read(o: Outer) { return o.secret; } } } }` compiles, and
				// typescript-eslint counts it as a use. This arm was bounded at a nested class once,
				// on the reasoning that `this.secret` there does not compile, and it reported that
				// read (#yfkhkvy). The checker's resolution is the bound that matters: `this.secret`
				// inside the nested class names the nested class's own member or nothing, never the
				// outer declaration, so no bound on nesting is needed to keep it from crediting one.
				markPrivateKeywordAccess(ctx, node, name, byName)
			case name.Kind == ast.KindPrivateIdentifier:
				// A `#name` reaches its declaring class from any depth, so this arm is not bounded
				// by nesting the way the keyword arm is: `other.#foo` inside the class reaches the
				// same member `this.#foo` does. What DOES stop it is a nested class redeclaring the
				// same spelling, from which point inward the name means the inner member. Upstream
				// carries three cases for exactly that shadow.
				if !shadowedByNestedClass(node, name.Text(), classNode) {
					markHashPrivateUse(node, name, byName)
				}
			}
		case ast.KindShorthandPropertyAssignment:
			// `({ privateMember } = this)` is a destructuring ASSIGNMENT rather than a declaration,
			// so it parses as a shorthand property assignment inside an object literal rather than
			// as a binding element. Two of upstream's passing cases are this shape and both report
			// without this arm.
			if !insideNestedClass {
				markShorthandAssignmentUse(ctx, node, byName)
			}

		case ast.KindBindingElement:
			// `const { privateMember } = this` reads the member without any property access, so the
			// property-access arm above never sees it. Seven of upstream's passing cases are this
			// shape and every one of them reports without this arm.
			if !insideNestedClass {
				markDestructuredUse(ctx, node, byName)
			}

		case ast.KindPrivateIdentifier:
			// `#brand in obj` is the one position a private name takes that is not a property
			// access name, and it is a genuine read: the class is asking whether an object carries
			// the member.
			if node.Parent != nil && node.Parent.Kind != ast.KindPropertyAccessExpression &&
				!shadowedByNestedClass(node, node.Text(), classNode) {
				markHashPrivateUse(nil, node, byName)
			}
		}

		nested := insideNestedClass
		if node != classNode &&
			(node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression) {
			nested = true
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, nested)
			return false
		})
	}
	visit(classNode, false)
}

// markPrivateKeywordAccess attributes one property access to a declared member, if it is one.
//
// The name text is tested first and the checker second, which is the cheap-test-first ordering this
// rule's cost depends on: most property accesses in a class body name something else entirely, and
// those cost a map lookup rather than a resolution.
//
// Attribution is by DECLARATION IDENTITY rather than by name. Two classes in one file can each
// declare `private prop`, and a name comparison would let a use in one keep the other alive.
func markPrivateKeywordAccess(
	ctx rule.Context,
	access *ast.Node,
	name *ast.Node,
	byName map[string][]*privateKeywordMember,
) {
	candidates, named := byName[name.Text()]
	if !named {
		return
	}
	allUsed := true
	for _, candidate := range candidates {
		if !candidate.used {
			allUsed = false
		}
	}
	if allUsed {
		return
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil {
		return
	}
	for _, declaration := range symbol.Declarations {
		for _, candidate := range candidates {
			if candidate.isHashPrivate || candidate.declaration != declaration || candidate.used {
				continue
			}
			if privateKeywordAccessIsRead(access, candidate.isAccessor) {
				candidate.used = true
			}
		}
	}
}

// privateKeywordAccessIsRead reports whether one access counts as using the member.
//
// Any reference to an accessor is a use, since both directions call a body. Everything else is a use
// unless it only ever writes, and the conservative default matters in one direction: an unrecognized
// reference shape reads as a use and the member goes unreported, rather than reading as dead and
// accusing working code.
func privateKeywordAccessIsRead(access *ast.Node, isAccessor bool) bool {
	if isAccessor {
		return true
	}
	return !privateKeywordAccessIsWriteOnly(access)
}

// privateKeywordAccessIsWriteOnly enumerates the shapes that store without ever loading.
//
// `ast.IsWriteAccess` is deliberately not used, for the two reasons the core rule of this name
// records at its own copy of this logic: it answers "does this store" rather than "does this ONLY
// store", so it reports true for `x += 1` and `x++`, both of which load; and it under-reports a rest
// element in a destructuring target. Under-reporting a write here becomes a phantom read and the
// member goes unreported, so the forms are enumerated instead.
func privateKeywordAccessIsWriteOnly(access *ast.Node) bool {
	parent := access.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.Left != access || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			// The right side of an assignment is a read, and so is either side of any other
			// operator.
			return false
		}
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			return true
		}
		// A compound assignment loads before it stores, so it is a read unless the result is thrown
		// away: `this.a += 1;` as a statement observes nothing.
		return parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement

	case ast.KindPrefixUnaryExpression:
		operator := parent.AsPrefixUnaryExpression().Operator
		if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
			return false
		}
		return parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement

	case ast.KindPostfixUnaryExpression:
		return parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement

	case ast.KindForInStatement, ast.KindForOfStatement:
		// `for (this.a in obj)` stores each key. The other position, `for (const k in this.a)`, is
		// the object being enumerated and reads.
		return parent.AsForInOrOfStatement().Initializer == access

	case ast.KindArrayLiteralExpression, ast.KindSpreadElement:
		// A destructuring target parses as an array literal, so only the climb to an assignment
		// separates `[this.a] = bar` from `foo([this.a])`.
		return privateKeywordDestructuringTarget(parent)

	case ast.KindPropertyAssignment:
		// `({ x: this.a } = bar)` stores. The key position `({ [this.a]: v } = bar)` reaches here
		// through a computed name rather than directly, so the identity test separates them.
		if parent.AsPropertyAssignment().Initializer != access {
			return false
		}
		return privateKeywordDestructuringTarget(parent)

	case ast.KindBindingElement:
		// `[this.a = 1] = bar` puts a default on the target, which still only stores.
		return privateKeywordDestructuringTarget(parent)
	}

	return false
}

// privateKeywordDestructuringTarget reports whether a node sits on the left of a destructuring
// assignment.
//
// The climb exists because a target and an ordinary array or object value parse to the same kinds,
// so only the assignment above distinguishes them, and it can be several levels up.
func privateKeywordDestructuringTarget(node *ast.Node) bool {
	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression,
			ast.KindPropertyAssignment, ast.KindSpreadElement,
			ast.KindParenthesizedExpression, ast.KindBindingElement:
			current = parent
			continue
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary.Left == current && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
		default:
			return false
		}
	}
	return false
}

// markHashPrivateUse marks a `#name` member used, by name identity rather than through the checker.
//
// A private name is lexically scoped to its class, so the spelling alone identifies the member and
// there is nothing to resolve. `access` is the property access when there is one and nil for the
// `#brand in obj` form, which has no receiver and is always a read.
func markHashPrivateUse(access *ast.Node, name *ast.Node, byName map[string][]*privateKeywordMember) {
	for _, candidate := range byName[name.Text()] {
		if !candidate.isHashPrivate || candidate.used {
			continue
		}
		if candidate.nameNode == name {
			// The declaration itself is not a use. Without this every member is marked used the
			// moment it is declared and the rule reports nothing ever.
			continue
		}
		if access == nil || privateKeywordAccessIsRead(access, candidate.isAccessor) {
			candidate.used = true
		}
	}
}

// parameterPropertySpanEnd is where upstream's finding on a parameter property stops.
//
// Measured on the installed 8.67.0 build across three shapes, since the corpus writes all three:
//
//	constructor(private p: number)               reports `p: number`
//	constructor(private readonly p: number)      reports `p: number`
//	constructor(private readonly p: number = 1)  reports `p: number`, not the default
//
// So the span runs from the name through the type annotation and stops before any initializer,
// which is the parameter's end when there is none.
func parameterPropertySpanEnd(parameter *ast.Node) int {
	declaration := parameter.AsParameterDeclaration()
	if declaration.Type != nil {
		return declaration.Type.End()
	}
	if declaration.Initializer != nil {
		return declaration.Name().End()
	}
	return parameter.End()
}

// shadowedByNestedClass reports whether a `#name` occurrence is inside a nested class that
// redeclares that same spelling.
//
// From such a class inward the name means the inner member, so crediting the occurrence to the outer
// declaration goes silent on a genuinely dead outer member. The walk stops at `outermost`, which is
// the class whose members are being judged.
func shadowedByNestedClass(node *ast.Node, name string, outermost *ast.Node) bool {
	for current := node.Parent; current != nil && current != outermost; current = current.Parent {
		var members *ast.NodeList
		switch current.Kind {
		case ast.KindClassDeclaration:
			members = current.AsClassDeclaration().Members
		case ast.KindClassExpression:
			members = current.AsClassExpression().Members
		default:
			continue
		}
		if members == nil {
			continue
		}
		for _, member := range members.Nodes {
			memberName := member.Name()
			if memberName != nil && memberName.Kind == ast.KindPrivateIdentifier &&
				memberName.Text() == name {
				return true
			}
		}
	}
	return false
}

// markDestructuredUse marks a member read through a destructuring pattern.
//
// Two shapes, and they resolve differently. A RENAMED binding, `{ privateMember: other }`, carries a
// property name that resolves straight to the declaration like any property access. A SHORTHAND
// binding, `{ privateMember }`, does not: probed, its name resolves to the binding element itself
// rather than to the property, so the member has to be found on the type of the object being
// destructured instead.
//
// Both are reads by construction. A destructuring pattern loads; there is no write-only form of it,
// which is why nothing here consults the write enumeration.
func markDestructuredUse(ctx rule.Context, element *ast.Node, byName map[string][]*privateKeywordMember) {
	binding := element.AsBindingElement()
	if binding.PropertyName != nil {
		if binding.PropertyName.Kind == ast.KindIdentifier {
			markResolvedUse(ctx, binding.PropertyName, byName)
		}
		return
	}

	name := binding.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}
	candidates, named := byName[name.Text()]
	if !named {
		return
	}
	source := destructuringSource(element)
	if source == nil {
		return
	}
	sourceType := ctx.TypeChecker.GetTypeAtLocation(source)
	if sourceType == nil {
		return
	}
	property := ctx.TypeChecker.GetPropertyOfType(sourceType, name.Text())
	if property == nil {
		return
	}
	for _, declaration := range property.Declarations {
		for _, candidate := range candidates {
			if !candidate.isHashPrivate && candidate.declaration == declaration {
				candidate.used = true
			}
		}
	}
}

// destructuringSource finds the expression a binding element is destructuring out of.
//
// The pattern can nest, so the climb walks out through object and array patterns to the declaration
// or assignment that supplies the value.
func destructuringSource(element *ast.Node) *ast.Node {
	for current := element.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern, ast.KindBindingElement:
			continue
		case ast.KindVariableDeclaration:
			return current.AsVariableDeclaration().Initializer
		case ast.KindParameter:
			// A destructuring pattern in a parameter position normally destructures the ARGUMENT,
			// which no caller here can see, so there is nothing to resolve against. The exception is
			// a default: `({ privateMember } = this) => {}` writes the object being destructured
			// right there as the initializer, and upstream carries it as a passing case. Traced
			// rather than guessed, since that shape reads as an assignment and parses as a binding
			// pattern with a default.
			return current.AsParameterDeclaration().Initializer
		default:
			return nil
		}
	}
	return nil
}

// markResolvedUse marks whichever declared member a name resolves to.
func markResolvedUse(ctx rule.Context, name *ast.Node, byName map[string][]*privateKeywordMember) {
	candidates, named := byName[name.Text()]
	if !named {
		return
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil {
		return
	}
	for _, declaration := range symbol.Declarations {
		for _, candidate := range candidates {
			if !candidate.isHashPrivate && candidate.declaration == declaration {
				candidate.used = true
			}
		}
	}
}

// markShorthandAssignmentUse marks a member read through a destructuring assignment.
//
// `({ privateMember } = this)` loads the member out of the object on the right, which makes it a
// read. The name resolves to the shorthand assignment itself rather than to the property, the same
// way a shorthand binding does, so the member is found on the source object's type instead.
func markShorthandAssignmentUse(
	ctx rule.Context,
	shorthand *ast.Node,
	byName map[string][]*privateKeywordMember,
) {
	name := shorthand.AsShorthandPropertyAssignment().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}
	candidates, named := byName[name.Text()]
	if !named {
		return
	}
	source := shorthandAssignmentSource(shorthand)
	if source == nil {
		return
	}
	sourceType := ctx.TypeChecker.GetTypeAtLocation(source)
	if sourceType == nil {
		return
	}
	property := ctx.TypeChecker.GetPropertyOfType(sourceType, name.Text())
	if property == nil {
		return
	}
	for _, declaration := range property.Declarations {
		for _, candidate := range candidates {
			if !candidate.isHashPrivate && candidate.declaration == declaration {
				candidate.used = true
			}
		}
	}
}

// shorthandAssignmentSource finds the right-hand side of the destructuring assignment a shorthand
// property sits in, or nil when the object literal is an ordinary value rather than a target.
func shorthandAssignmentSource(shorthand *ast.Node) *ast.Node {
	for current := shorthand.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
			ast.KindPropertyAssignment, ast.KindParenthesizedExpression:
			continue
		case ast.KindBinaryExpression:
			binary := current.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindEqualsToken {
				return binary.Right
			}
			return nil
		case ast.KindParameter:
			// `({ privateMember } = this) => {}` is a destructuring assignment used as a parameter
			// DEFAULT, so the object literal is the parameter's name and `this` is its initializer.
			// Upstream carries this shape as a passing case and it reaches here rather than through
			// a binary expression.
			return current.AsParameterDeclaration().Initializer
		default:
			return nil
		}
	}
	return nil
}
