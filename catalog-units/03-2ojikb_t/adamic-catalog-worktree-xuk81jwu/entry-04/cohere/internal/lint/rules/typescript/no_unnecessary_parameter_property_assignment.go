package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoUnnecessaryParameterPropertyAssignment = rule.Message{
	Id: "unnecessaryAssign",
	Description: "A constructor parameter carrying `public`, `private`, `protected`, `readonly` " +
		"or `override` is already assigned to the matching property before the constructor body " +
		"runs. Writing it again does nothing, and it invites the next reader to believe the two " +
		"names could diverge. Delete the assignment.",
}

var suggestionNoUnnecessaryParameterPropertyAssignment = rule.Message{
	Id:          "removeAssignment",
	Description: "Remove the unnecessary assignment.",
}

// NoUnnecessaryParameterPropertyAssignment flags a constructor that re-assigns a parameter property.
//
//	valid:   class Foo { constructor(public name: unknown) {} }
//	valid:   class Foo { constructor(name: unknown) { this.name = name; } }
//	valid:   class Foo { constructor(public name: unknown) { this.name = name + 'x'; } }
//	valid:   class Foo { constructor(public foo: number) { { const foo = 1; this.foo = foo; } } }
//	invalid: class Foo { constructor(public name: unknown) { this.name = name; } }
//	invalid: class Foo { constructor(private foo: string) { this['foo'] = foo; } }
//	invalid: class Foo { constructor(public foo: string) { this.foo ??= foo; } }
//
// Ported from `typescript/no-unnecessary-parameter-property-assignment`, which oxc in turn ports
// from `@typescript-eslint/no-unnecessary-parameter-property-assignment`.
//
// # What counts as a parameter property, measured rather than assumed
//
// Upstream's `FormalParameter::has_modifier` is `accessibility.is_some() || readonly || override`.
// Our tree spells the same set as `ModifierFlagsParameterPropertyModifier`, and that equivalence is
// a measurement rather than a reading: probed across `public`, `private`, `protected`, `readonly`,
// `override`, a bare parameter and a decorated one, the flag answers true for exactly the first
// five and false for the last two, agreeing with `has_modifier` on all seven. `override` is the
// one that surprises, since it creates no property of its own, and `ModifierFlagsAccessibilityModifier`
// would have silently dropped both it and `readonly`. Pinned upstream too: `override foo: string`
// and `readonly foo: string` each report on the release binary.
//
// # The walk stops at an assignment, and that is what the arrow cases are really about
//
// Upstream's visitor overrides `visit_assignment_expression` and never calls the default walker
// from it, so an assignment expression is a dead end: neither side is descended into. It separately
// overrides `visit_function`, which stops function declarations, function expressions and methods
// because each rebinds `this`. An arrow function is NOT stopped, because it closes over `this`.
//
// Reading the corpus alone suggests the opposite, because the one clean arrow case
// (`this.bar = () => { this.foo = foo; }`) looks like evidence that arrows are skipped. It is not.
// That case is clean because the enclosing assignment is the dead end, and the arrow is never
// reached at all. Measured on the release binary, with the arrow moved out from under an
// assignment: `call(() => { this.foo = foo; })` reports, and a bare `(() => { this.foo = foo; });`
// expression statement reports. Both would be silent under an arrow skip. A port that skipped
// arrows would pass every imported fixture and go silent on both of those.
//
// # A prior write to the parameter does NOT rescue the assignment
//
// `constructor(public foo: string) { foo = foo.trim(); this.foo = foo; }` **reports** on the
// release binary, even though the property and the parameter now genuinely differ. Upstream's
// `assigned_before_unnecessary` set records only targets of the form `this.X`, so a plain write to
// the binding is invisible to it. That is unsound and it is reproduced deliberately, with a fixture,
// because the gate runs oxlint and a port that quietly fixed it would read as a difference. The
// intuitive reading is recorded here so the next reader does not helpfully correct it back.
//
// # Parentheses fall two different ways inside one assignment
//
// oxc's parser has already dropped the parentheses around an assignment target by the time the rule
// sees it, while ours keeps a real ParenthesizedExpression node. So reproducing upstream means
// adding a skip our tree needs and upstream's source does not show. Measured on the release binary:
//
//	(this.foo) = foo    reports, and the span is the whole `(this.foo) = foo`
//	this.foo = (foo)    reports, via `get_inner_expression()` on the right
//	(this).foo = foo    SILENT, because the `this` object is matched by kind with no unwrapping
//
// Three inputs, three different answers, in one rule. The corpus writes none of them.
//
// # The class-body scan is position-independent and reads only property initializers
//
// A property definition anywhere in the class whose initializer assigns to `this.X`, directly or
// through an immediately invoked function, suppresses the finding for `X`. Measured both ways:
// `init = (this.foo = 1)` silences the constructor whether it is written above the constructor or
// below it. A sibling *method* assigning `this.foo` does not silence it, because a method is not a
// property definition. All three pinned on the release binary.
//
// # The repair is a suggestion rather than a fix, and upstream's range eats one blind byte
//
// Upstream calls `diagnostic_with_suggestion` and declares `suggestion` in its metadata, so `--fix`
// alone applies nothing and only `--fix-suggestions` does. The range is the assignment's span plus
// one trailing byte, which is the semicolon in every case the corpus writes. It is blind: measured
// on the release binary, `this.foo = foo` with no semicolon has its newline eaten and the next line
// joined, and `this.foo = foo, bar();` has its comma eaten, leaving ` bar();`. Reproduced exactly,
// as a suggestion, because a suggestion is the vehicle upstream chose and a human sees it before it
// lands. Shipping this as a fix would let the engine apply those two rewrites unattended.
var NoUnnecessaryParameterPropertyAssignment = rule.Rule{
	Name: "@typescript-eslint/no-unnecessary-parameter-property-assignment",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindConstructor: func(node *ast.Node) {
				constructor := node.AsConstructorDeclaration()
				if constructor == nil || constructor.Body == nil {
					return
				}

				// Upstream requires a ClassBody parent because its `MethodDefinition` kind covers
				// every method shape and the parent is what narrows it to a class. Our kind already
				// does that narrowing, and a guard here would be unreachable rather than merely
				// redundant.
				//
				// Measured on the parse shape rather than assumed, because a survivor said so: an
				// `ast.IsClassLike(node.Parent)` test was written here first, a mutation
				// neutralizing it survived the whole fixture set, and probing seven shapes showed
				// why. A construct signature in an interface or a type literal is
				// KindConstructSignature, an object literal member spelled `constructor` is
				// KindMethodDeclaration, and a bare `constructor(...)` at top level under error
				// recovery produces no KindConstructor at all. Only KindClassDeclaration and
				// KindClassExpression parent one, and both are class-like, so no input could reach
				// the decline.
				class := node.Parent
				if class == nil {
					return
				}

				parameterProperties := map[string]bool{}
				if constructor.Parameters != nil {
					for _, parameter := range constructor.Parameters.Nodes {
						if !ast.HasSyntacticModifier(parameter, ast.ModifierFlagsParameterPropertyModifier) {
							continue
						}
						// Upstream's `get_binding_identifier` answers nothing for a destructuring
						// pattern. TypeScript rejects one on a parameter property with error 1187,
						// so this never appears in code that compiles, but the PARSER still builds
						// the node and sets the modifier flag on it, and a linter has to survive
						// invalid source.
						//
						// This is crash protection rather than a behavioral filter, which is why a
						// mutation removing it survived the whole fixture set: no `ExpectFindings`
						// case can observe a panic. Measured directly on
						// `constructor(public { a }: { a: string })` — the parameter carries
						// ModifierFlagsParameterPropertyModifier, its name is a
						// KindObjectBindingPattern, and `Text()` panics with "Unhandled case in
						// Node.Text: *ast.BindingPattern". That is a fourth panicking shape beside
						// the three already written down for this accessor.
						name := parameter.Name()
						if name == nil || name.Kind != ast.KindIdentifier {
							continue
						}
						parameterProperties[name.Text()] = true
					}
				}
				if len(parameterProperties) == 0 {
					return
				}

				assignedBeforeConstructor := propertiesAssignedInClassBody(class)

				visitor := &parameterPropertyAssignmentVisitor{
					context:                   ctx,
					parameterProperties:       parameterProperties,
					assignedBeforeUnnecessary: map[string]bool{},
					assignedBeforeConstructor: assignedBeforeConstructor,
				}
				visitor.visit(constructor.Body)
			},
		}
	},
}

// parameterPropertyAssignmentVisitor walks a constructor body the way upstream's visitor does.
//
// Written as an explicit descent rather than as a listener plus an ancestor walk, because the two
// stopping conditions are properties of the descent itself: an assignment expression is never
// descended into at all, and a function-like that rebinds `this` ends the walk. An ancestor walk
// asking "how many functions am I inside" cannot express the first of those, and the census records
// a sibling rule shipping a live false positive from exactly that substitution.
type parameterPropertyAssignmentVisitor struct {
	context                   rule.Context
	parameterProperties       map[string]bool
	assignedBeforeUnnecessary map[string]bool
	assignedBeforeConstructor map[string]bool
}

func (v *parameterPropertyAssignmentVisitor) visit(node *ast.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	// Upstream's `visit_function` override. Each of these rebinds `this`, so an assignment inside
	// one writes to a different receiver and says nothing about this constructor's parameter
	// property. An arrow function is deliberately absent: it closes over `this` and upstream walks
	// into it.
	//
	// A nested CONSTRUCTOR belongs here and its absence was a real defect, caught by upstream case
	// 16 and by nothing else. Upstream's `MethodDefinition` holds a `Function` as its value, so a
	// nested class's constructor is stopped by the same override that stops a method. Without this
	// arm the descent reports the inner assignment, and then the rule's own KindConstructor
	// listener reports it a second time when the walk reaches `class Bar` on its own: three
	// findings where upstream produces two, with the extra one sitting on top of a correct one so
	// it reads as a duplicate rather than as a wrong span.
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return

	case ast.KindBinaryExpression:
		if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			v.visitAssignment(node)
			// Upstream's override returns on every path without calling the default walker, so an
			// assignment is a dead end and neither side is descended into. This is what makes
			// `this.bar = () => { this.foo = foo; }` clean, rather than any property of arrows.
			return
		}
	}

	node.ForEachChild(func(child *ast.Node) bool {
		v.visit(child)
		return false
	})
}

func (v *parameterPropertyAssignmentVisitor) visitAssignment(node *ast.Node) {
	assignment := node.AsBinaryExpression()

	propertyName, ok := thisPropertyNameOfAssignmentTarget(assignment.Left)
	if !ok {
		return
	}

	if !isUnnecessaryAssignmentOperator(assignment.OperatorToken.Kind) {
		// A compound arithmetic assignment reads the property, so the parameter property's value is
		// now load-bearing and every later plain assignment to the same name is necessary.
		v.assignedBeforeUnnecessary[propertyName] = true
		return
	}

	right := skipOuterExpressions(assignment.Right)
	if right == nil || right.Kind != ast.KindIdentifier {
		return
	}
	if right.Text() != propertyName {
		return
	}
	if !v.parameterProperties[propertyName] {
		return
	}

	// Upstream additionally asks the semantic layer whether the identifier on the right resolves to
	// the parameter's own binding, which is what makes the block-shadow case clean:
	//
	//	constructor(public foo: number) { { const foo = 1; this.foo = foo; } }
	//
	// We have no reference index, so the same question is answered structurally by walking out from
	// the identifier looking for an intervening declaration of that name. See the helper.
	if identifierIsShadowedBefore(right, propertyName) {
		return
	}

	if v.assignedBeforeUnnecessary[propertyName] {
		return
	}
	if v.assignedBeforeConstructor[propertyName] {
		return
	}

	// Anchored on the reported range rather than on `node.Pos()`. A node's `Pos()` is the position
	// BEFORE its leading trivia, so a deletion starting there swallows the indentation of the line
	// as well as the assignment, and every one of upstream's 21 fix vectors keeps that indentation.
	// `rule.TokenRange` is the same trivia-stripped span the diagnostic itself carries, which is
	// also what makes the finding and its repair point at the same bytes.
	reported := rule.TokenRange(v.context.SourceFile, node)

	v.context.ReportNodeWithSuggestions(node, messageNoUnnecessaryParameterPropertyAssignment,
		rule.Suggestion{
			Message: suggestionNoUnnecessaryParameterPropertyAssignment,
			Fixes: []rule.Fix{{
				// Upstream's `Span::new(span.start, span.end + 1)`, which reaches one byte past the
				// expression to take the semicolon with it. The blind byte is documented at the
				// rule and is why this is a suggestion rather than a fix.
				Range: core.NewTextRange(reported.Pos(), reported.End()+1),
				Text:  "",
			}},
		})
}

// isUnnecessaryAssignmentOperator is upstream's operator filter.
//
// Plain `=` plus the three logical assignments, and nothing else. A logical assignment qualifies
// because a parameter property is assigned before the body runs, so `??=` and `&&=` and `||=` all
// resolve against a value the constructor did not compute. The arithmetic forms read the property
// first, which is why they land in the other set.
func isUnnecessaryAssignmentOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// thisPropertyNameOfAssignmentTarget is upstream's `get_property_name`.
//
// It answers only for a target written directly on `this`, and only when the key is statically
// known. A computed key holding an identifier or a template is declined even though our tree could
// often resolve it, because upstream declines it: `this[name] = foo` and “ this[`${foo}`] = foo “
// are both clean cases in the corpus.
//
// The parenthesis skip on the target has no counterpart in upstream's source and is required to
// match it. oxc's parser converts a parenthesized assignment target into the bare target before the
// rule runs, so `(this.foo) = foo` reports there; our parser keeps the node. The `this` object
// underneath is matched by kind with NO skip, because upstream matches `Expression::ThisExpression`
// directly and `(this).foo = foo` is silent on the release binary.
func thisPropertyNameOfAssignmentTarget(target *ast.Node) (string, bool) {
	target = skipParentheses(target)
	if target == nil {
		return "", false
	}

	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		access := target.AsPropertyAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindThisKeyword {
			return "", false
		}
		if access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return "", false
		}
		return access.Name().Text(), true

	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindThisKeyword {
			return "", false
		}
		if access.ArgumentExpression == nil ||
			access.ArgumentExpression.Kind != ast.KindStringLiteral {
			return "", false
		}
		return access.ArgumentExpression.Text(), true
	}

	return "", false
}

// skipParentheses unwraps parenthesized expressions and nothing else.
func skipParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.Expression()
	}
	return node
}

// skipOuterExpressions is upstream's `get_inner_expression` on the right of the assignment.
//
// It unwraps parentheses and the type-level wrappers that erase at runtime, which is what makes
// `this.foo = foo!` and `this.foo = foo as any` report while leaving a real expression alone. Both
// of those are failing cases in the corpus.
func skipOuterExpressions(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindNonNullExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindTypeAssertionExpression,
			ast.KindExpressionWithTypeArguments:
			node = node.Expression()
		default:
			return node
		}
	}
	return nil
}

// identifierIsShadowedBefore answers whether a name resolves to something other than the
// constructor parameter, by looking for an intervening declaration between the two.
//
// This replaces upstream's `symbol_references` cross-check, which asks the semantic layer whether
// the identifier on the right is the same symbol as the parameter binding. The only corpus case
// that turns on it is a block-scoped redeclaration:
//
//	constructor(public foo: number) { { const foo = 1; this.foo = foo; } }
//
// Walking out from the identifier to the constructor and asking each enclosing block whether it
// declares that name answers the same question for every shape a constructor body can hold, because
// a parameter is the outermost binding of its name inside its own body. Anything nearer shadows it,
// and there is nothing between the parameter and the body for a declaration to hide in.
func identifierIsShadowedBefore(identifier *ast.Node, name string) bool {
	for current := identifier.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindConstructor {
			return false
		}
		if blockDeclaresName(current, name) {
			return true
		}
		if ast.IsFunctionLike(current) {
			// A function-like nearer than the constructor rebinds names of its own. The descent
			// already refuses to enter the kinds that rebind `this`, so this is reached only for an
			// arrow, whose parameters can still shadow.
			if functionParametersDeclareName(current, name) {
				return true
			}
		}
	}
	return false
}

// blockDeclaresName answers whether a statement container declares the name directly.
//
// The kind test is not an optimization. `Node.Statements()` routes through `Node.StatementList()`,
// which PANICS with "Unhandled case in Node.StatementList" on anything that is not one of these
// three kinds, exactly the way `Node.ParameterList()` and `Node.Text()` panic off their own kinds.
// This walk climbs through arbitrary ancestors, so it meets a BinaryExpression on the second case
// it ever sees, and the unguarded version took the whole test binary down.
func blockDeclaresName(node *ast.Node, name string) bool {
	switch node.Kind {
	case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock:
	default:
		return false
	}

	for _, statement := range node.Statements() {
		if statement.Kind != ast.KindVariableStatement {
			continue
		}
		declarationList := statement.AsVariableStatement().DeclarationList
		if declarationList == nil {
			continue
		}
		list := declarationList.AsVariableDeclarationList()
		if list == nil || list.Declarations == nil {
			continue
		}
		for _, declaration := range list.Declarations.Nodes {
			declarationName := declaration.Name()
			if declarationName != nil &&
				declarationName.Kind == ast.KindIdentifier &&
				declarationName.Text() == name {
				return true
			}
		}
	}
	return false
}

// functionParametersDeclareName answers whether a function-like binds the name as a parameter.
func functionParametersDeclareName(node *ast.Node, name string) bool {
	if !ast.IsFunctionLike(node) {
		return false
	}
	parameters := node.ParameterList()
	if parameters == nil {
		return false
	}
	for _, parameter := range parameters.Nodes {
		parameterName := parameter.Name()
		if parameterName != nil &&
			parameterName.Kind == ast.KindIdentifier &&
			parameterName.Text() == name {
			return true
		}
	}
	return false
}

// propertiesAssignedInClassBody is upstream's `assigned_before_constructor` scan.
//
// It reads only property definitions carrying an initializer, and the initializer shapes it
// understands are exactly upstream's: a bare assignment, and an immediately invoked arrow or
// function whose body assigns. Position in the class does not matter, which was measured rather
// than read: `init = (this.foo = 1)` silences the constructor from above it and from below it.
//
// The parenthesis skip here IS in upstream's source, as `without_parentheses()`, and it is
// load-bearing on our tree because every corpus case writes the initializer as `(this.foo += 1)`.
func propertiesAssignedInClassBody(class *ast.Node) map[string]bool {
	assigned := map[string]bool{}

	members := class.Members()
	if members == nil {
		return assigned
	}

	for _, member := range members {
		if member.Kind != ast.KindPropertyDeclaration {
			continue
		}
		initializer := member.Initializer()
		if initializer == nil {
			continue
		}
		for _, assignment := range assignmentsInsideInitializer(initializer) {
			if name, ok := thisPropertyNameOfAssignmentTarget(assignment.AsBinaryExpression().Left); ok {
				assigned[name] = true
			}
		}
	}

	return assigned
}

// assignmentsInsideInitializer is upstream's `get_assignments_inside_expression`.
//
// Two shapes only: the initializer is itself an assignment, or it is a call whose callee is an
// arrow or a function expression, in which case both the arrow's expression body and any
// expression-statement assignment in a block body count. Anything else contributes nothing, which
// is why a sibling method assigning `this.foo` does not suppress the finding.
func assignmentsInsideInitializer(expression *ast.Node) []*ast.Node {
	var assignments []*ast.Node

	expression = skipParentheses(expression)
	if expression == nil {
		return assignments
	}

	// Deliberately NOT unwrapping parentheses, and this is the single most counter-intuitive line
	// in the rule. Upstream calls `without_parentheses()` on the initializer and on the callee and
	// nowhere else, so a parenthesized assignment reached through either arm below is simply not
	// recognized and the suppression does not happen. Adding a skip here reads as a free
	// correctness improvement and is a divergence. All four shapes measured on the release binary:
	//
	//	init = (() => this.foo = 1)()      suppresses   arrow expression body, bare
	//	init = (() => (this.foo = 1))()    REPORTS      arrow expression body, parenthesized
	//	init = (() => { this.foo = 1; })() suppresses   block body, bare
	//	init = (() => { (this.foo = 1); })() REPORTS    block body, parenthesized
	//
	// The corpus writes only the two bare forms, so a port unwrapping here passes every imported
	// fixture and silently suppresses two inputs upstream reports on.
	appendIfAssignment := func(node *ast.Node) {
		if node != nil && node.Kind == ast.KindBinaryExpression &&
			ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			assignments = append(assignments, node)
		}
	}

	switch expression.Kind {
	case ast.KindCallExpression:
		callee := skipParentheses(expression.AsCallExpression().Expression)
		if callee == nil {
			return assignments
		}

		var body *ast.Node
		switch callee.Kind {
		case ast.KindArrowFunction:
			body = callee.AsArrowFunction().Body
		case ast.KindFunctionExpression:
			body = callee.AsFunctionExpression().Body
		default:
			return assignments
		}
		if body == nil {
			return assignments
		}

		// An arrow with an expression body: upstream's `get_expression()` arm.
		if body.Kind != ast.KindBlock {
			appendIfAssignment(body)
			return assignments
		}

		for _, statement := range body.Statements() {
			if statement.Kind != ast.KindExpressionStatement {
				continue
			}
			appendIfAssignment(statement.AsExpressionStatement().Expression)
		}

	case ast.KindBinaryExpression:
		appendIfAssignment(expression)
	}

	return assignments
}
