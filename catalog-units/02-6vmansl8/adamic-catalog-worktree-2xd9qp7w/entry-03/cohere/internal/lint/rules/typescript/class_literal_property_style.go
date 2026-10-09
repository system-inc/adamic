package typescript

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageClassLiteralPropertyStylePreferField is upstream's `preferFieldStyle`.
func messageClassLiteralPropertyStylePreferField() rule.Message {
	return rule.Message{
		Id: "preferFieldStyle",
		Description: "This getter does nothing but return a literal, so it pays a function call " +
			"on every read to express a constant. A `readonly` field says the same thing, cannot " +
			"be reassigned, and lets the compiler and a reader see the value at the declaration.",
	}
}

// messageClassLiteralPropertyStylePreferFieldSuggestion is upstream's `preferFieldStyleSuggestion`.
func messageClassLiteralPropertyStylePreferFieldSuggestion() rule.Message {
	return rule.Message{
		Id:          "preferFieldStyleSuggestion",
		Description: "Replace the getter with a readonly field holding the same literal.",
	}
}

// messageClassLiteralPropertyStylePreferGetter is upstream's `preferGetterStyle`.
func messageClassLiteralPropertyStylePreferGetter() rule.Message {
	return rule.Message{
		Id: "preferGetterStyle",
		Description: "This project exposes class literals through getters, and this one is a " +
			"`readonly` field. A getter keeps the value off the instance and out of anything that " +
			"enumerates or serializes properties, which is the difference the convention exists for.",
	}
}

// messageClassLiteralPropertyStylePreferGetterSuggestion is upstream's `preferGetterStyleSuggestion`.
func messageClassLiteralPropertyStylePreferGetterSuggestion() rule.Message {
	return rule.Message{
		Id:          "preferGetterStyleSuggestion",
		Description: "Replace the readonly field with a getter returning the same literal.",
	}
}

// ClassLiteralPropertyStyle requires one spelling for a class literal, a readonly field or a getter.
//
//	valid (fields):   class C { readonly x = 1; }
//	valid (fields):   class C { get x() { return 1; } set x(v) {} }   a paired setter exempts
//	valid (getters):  class C { get x() { return 1; } }
//	valid (getters):  class C { readonly x = 1; constructor() { this.x = 2; } }   assigned, so exempt
//	invalid (fields): class C { get x() { return 1; } }
//	invalid (getters):class C { readonly x = 1; }
//
// Ported from `@typescript-eslint/class-literal-property-style`, reading the clone at
// `packages/eslint-plugin/src/rules/class-literal-property-style.ts` and measuring every verdict and
// every repair against the installed 8.67.0 build driven through the ESLint 10.8.1 Linter API.
//
// # It offers SUGGESTIONS, not fixes, and the difference is the whole safety story
//
// `meta` carries `hasSuggestions: true` and NO `fixable`, so the edit engine never applies this
// repair unattended. Measured rather than read: `cohereAndFix` on a reporting input returns
// `fixed: false` and leaves the source untouched.
//
// That matters because the getter direction genuinely changes types, and a fix would apply that
// change without asking. Measured through the TypeScript compiler on the rewritten source rather
// than reasoned about:
//
//	readonly x: 1 | 2 = 1;   type `1 | 2`   becomes get x() { return 1; }   type `number`
//	readonly x = 1;          type `1`       becomes get x() { return 1; }   type `number`
//	readonly x: string = "a" type `string`  becomes get x() { return "a"; } type `string`
//	readonly x: number = 1;  type `number`  becomes get x() { return 1; }   type `number`
//
// The first two widen. The second is upstream's own core case and its most common input, so the
// widening is upstream's deliberate semantics rather than an oversight, and it is reproduced. It is
// only defensible because a human chooses it: shipping this as a fix would silently widen every
// literal-typed readonly field in the tree, which is the failure two rules shipped tonight.
//
// # The field direction preserves the return type, and the mechanism is worth copying
//
// A naive rebuild of `get x(): number { return 1; }` into a field would write `readonly x = 1;` and
// drop the annotation. Upstream does not: it splices the RAW SOURCE between the parameter list's
// closing parenthesis and the body's opening brace, so whatever lives there survives byte for byte.
// Measured on the installed build:
//
//	get x(): number { return 1; }         readonly x: number = 1;
//	public static get x(): 1 { return 1 } public static readonly x: 1 = 1;
//	get x()   :   number   { ... }        readonly x   :   number   = 1;      spacing preserved
//	get x() /* c */ { return 1; }         readonly x /* c */ = 1;             comment preserved
//
// This is the pattern the brief asks for and it is better than declining: rather than enumerating
// what might live in the span, it copies the span. Nothing inside it can be lost because nothing
// inside it is re-rendered.
//
// # A decorated getter reports with NO suggestion
//
// `getFixOrSuggest({ fixOrSuggest: node.decorators.length === 0 ? 'suggest' : 'none' })`. A
// decorator applies to the member as declared, so rewriting a getter into a field would change what
// the decorator receives. Measured: `class C { @dec get x() { return 1; } }` reports one diagnostic
// carrying zero suggestions. That is a decision to reproduce, not an omission, and it is the only
// input in this rule where the count of suggestions differs from the count of findings.
//
// # Where cohere is deliberately quieter: an override the keyword does not mark
//
// Upstream's `override` guard reads the keyword, and without `noImplicitOverride` an override carries
// none, so upstream reports conversions the compiler rejects. cohere asks the checker whether the
// member overrides a concrete base member of the other kind and declines exactly there; see
// classLiteralConversionBreaksAnOverride. This is the only use of the checker in the rule, which is
// why it is declared needed and why, with no checker, the rule falls back to upstream's verdicts.
var ClassLiteralPropertyStyle = rule.Rule{
	Name: "@typescript-eslint/class-literal-property-style",

	// For classLiteralConversionBreaksAnOverride alone. The base class is usually in another file
	// (StringSchema.ts against BaseSchema.ts), so the verdict reads the program as well.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The zero value is not the default and this is the normal path, not a defensive branch. A
		// rule configured as a bare `"error"` is handed nil options, because `DecodeOptionsInto`
		// errors on empty input and the config layer turns that into nil. A bare type assertion
		// would then yield the empty string, which matches neither arm below, and the rule would
		// register on every file and report nothing while every decoder-routed fixture stayed green.
		parsed, decoded := rule.OptionsAs[ClassLiteralPropertyStyleOptions](options)
		if !decoded {
			parsed = DefaultClassLiteralPropertyStyleOptions()
		}

		if parsed.Style == ClassLiteralPropertyStyleFields {
			return rule.Listeners{
				ast.KindGetAccessor: func(node *ast.Node) {
					reportClassLiteralGetterShouldBeField(ctx, node)
				},
			}
		}

		// The getters direction needs the whole class before it can judge any member, because a
		// property assigned in the constructor is exempt and the constructor may be written after
		// the property. Upstream gets that from a ClassBody enter/exit pair; there is no
		// `rule.OnExit` here, so the class is walked in one pass from its own listener.
		return rule.Listeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				reportClassLiteralFieldsShouldBeGetters(ctx, node.Members())
			},
			ast.KindClassExpression: func(node *ast.Node) {
				reportClassLiteralFieldsShouldBeGetters(ctx, node.Members())
			},
		}
	},
}

// reportClassLiteralGetterShouldBeField is upstream's `MethodDefinition` visitor under `fields`.
func reportClassLiteralGetterShouldBeField(ctx rule.Context, node *ast.Node) {
	accessor := node.AsGetAccessorDeclaration()

	// Upstream's `node.override` guard. An overriding getter cannot become a field without changing
	// what it overrides.
	if classLiteralHasModifier(node, ast.KindOverrideKeyword) {
		return
	}
	// The same guard for an override the keyword does not mark, and narrower: only where the field
	// would not compile. Deliberately quieter than upstream.
	if classLiteralConversionBreaksAnOverride(ctx, node, true) {
		return
	}

	body := accessor.Body
	if body == nil || len(body.AsBlock().Statements.Nodes) == 0 {
		return
	}

	// Upstream reads `body.body[0]` and requires it to be a return, so a getter doing anything
	// before returning is left alone even when the return itself is a literal.
	statement := body.AsBlock().Statements.Nodes[0]
	if statement.Kind != ast.KindReturnStatement {
		return
	}

	argument := statement.AsReturnStatement().Expression
	if !classLiteralIsSupportedLiteral(argument) {
		return
	}

	name := accessor.Name()
	if name == nil {
		return
	}

	// A getter paired with a setter of the same name is exempt: turning it into a field would leave
	// the setter with nothing to pair against. Upstream compares the STATIC MEMBER ACCESS VALUE
	// rather than the source text, so `get x` pairs with `set 'x'`.
	if classLiteralHasMatchingSetter(ctx, node, name) {
		return
	}

	// A decorator applies to the member as declared, so upstream reports without offering a repair.
	// The finding still fires; only the suggestion is withheld.
	if classLiteralHasDecorator(node) {
		ctx.ReportNode(classLiteralReportedKey(name), messageClassLiteralPropertyStylePreferField())
		return
	}

	replacement, buildable := classLiteralFieldReplacementFor(ctx, node, accessor, name, argument)
	if !buildable {
		ctx.ReportNode(classLiteralReportedKey(name), messageClassLiteralPropertyStylePreferField())
		return
	}

	ctx.ReportNodeWithSuggestions(classLiteralReportedKey(name), messageClassLiteralPropertyStylePreferField(),
		rule.Suggestion{
			Message: messageClassLiteralPropertyStylePreferFieldSuggestion(),
			Fixes: []rule.Fix{
				rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), replacement),
			},
		})
}

// classLiteralConversionBreaksAnOverride reports whether converting `member` to the other kind would
// be a compile error because it overrides a concrete base member of the kind it is now.
//
// `becomesProperty` is true for a getter about to become a field and false for a field about to
// become a getter. The two errors are TS2610 (a property overriding a base accessor) and TS2611 (an
// accessor overriding a base property), and the conditions are the vendored checker's own, in
// checkKindsOfPropertyMemberOverrides:
//
//	instance members only          the checker compares only the instance side, so a static
//	                               conversion compiles and keeps reporting
//	neither side private           a private member is not an override
//	base not from a mapped type    the checker skips those
//	base not abstract or interface a field may implement an abstract accessor, which is why the
//	                               seven `typeName` getters over `abstract get typeName()` in
//	                               ahra's schemas stay true positives
//	base of the member's own kind  a getter over a base property is ALREADY TS2611, and the
//	                               conversion repairs it, so that keeps reporting
//
// The real site is `nexus/source/validation/schema/StringSchema.ts:37` in ahra, a `typeDefault` getter
// overriding `BaseSchema`'s concrete getter, where ESLint's suggested field is TS2610.
func classLiteralConversionBreaksAnOverride(ctx rule.Context, member *ast.Node, becomesProperty bool) bool {
	if ctx.TypeChecker == nil || classLiteralHasModifier(member, ast.KindStaticKeyword) {
		return false
	}
	class := member.Parent
	if class == nil || (class.Kind != ast.KindClassDeclaration && class.Kind != ast.KindClassExpression) {
		return false
	}
	classSymbol := class.Symbol()
	memberSymbol := member.Symbol()
	if classSymbol == nil || memberSymbol == nil {
		return false
	}
	if checker.GetDeclarationModifierFlagsFromSymbol(memberSymbol)&ast.ModifierFlagsPrivate != 0 {
		return false
	}

	classType := ctx.TypeChecker.GetDeclaredTypeOfSymbol(classSymbol)
	if classType == nil {
		return false
	}
	for _, baseType := range ctx.TypeChecker.GetBaseTypes(classType) {
		base := ctx.TypeChecker.GetPropertyOfType(baseType, memberSymbol.Name)
		if base == nil {
			continue
		}
		baseKind := base.Flags & ast.SymbolFlagsPropertyOrAccessor
		if baseKind == 0 {
			continue
		}
		baseModifiers := checker.GetDeclarationModifierFlagsFromSymbol(base)
		if baseModifiers&ast.ModifierFlagsPrivate != 0 || base.CheckFlags&ast.CheckFlagsMapped != 0 {
			continue
		}
		if classLiteralBaseIsAbstractOrInterface(base, baseModifiers) {
			continue
		}
		if becomesProperty {
			return baseKind != ast.SymbolFlagsProperty
		}
		return baseKind == ast.SymbolFlagsProperty
	}
	return false
}

// classLiteralBaseIsAbstractOrInterface is the checker's arePropertiesAbstractOrInterface: when it
// holds, base and derived kinds need not match. A synthetic (intersection) property needs ANY
// declaration to qualify, every other property needs ALL of them.
//
// The checker also requires an abstract PROPERTY to have no initializer. That clause is omitted: an
// abstract property with an initializer is itself TS1267, so on compiling code it never decides, and
// a mutation deleting it survived every fixture for that reason.
func classLiteralBaseIsAbstractOrInterface(base *ast.Symbol, baseModifiers ast.ModifierFlags) bool {
	qualifies := func(declaration *ast.Node) bool {
		if declaration.Parent != nil && declaration.Parent.Kind == ast.KindInterfaceDeclaration {
			return true
		}
		return baseModifiers&ast.ModifierFlagsAbstract != 0
	}
	// With no declarations the ALL arm answers true, which is the checker's `core.Every` on an empty
	// list: no declaration to compare means no kind mismatch to report, so the rule keeps reporting.
	if base.CheckFlags&ast.CheckFlagsSynthetic != 0 {
		for _, declaration := range base.Declarations {
			if qualifies(declaration) {
				return true
			}
		}
		return false
	}
	for _, declaration := range base.Declarations {
		if !qualifies(declaration) {
			return false
		}
	}
	return true
}

// classLiteralFieldReplacementFor builds upstream's field text for a getter.
//
// The span between the parameter list's closing parenthesis and the body's opening brace is copied
// from the source rather than re-rendered, which is what preserves a return annotation, its spacing,
// and any comment written there. See the rule's doc comment for the four measured inputs.
func classLiteralFieldReplacementFor(
	ctx rule.Context,
	node *ast.Node,
	accessor *ast.GetAccessorDeclaration,
	name *ast.Node,
	argument *ast.Node,
) (string, bool) {
	body := accessor.Body
	if body == nil {
		return "", false
	}

	// The closing parenthesis is the last `)` before the body. Scanning back from the body's own
	// start rather than forward from the parameter list, because a return annotation sits between
	// them and its own text may contain parentheses.
	closingParen := classLiteralClosingParenBefore(ctx, node, body)
	if closingParen < 0 {
		return "", false
	}

	sourceText := ctx.SourceFile.Text()
	bodyStart := type_checking.TrimNodeTextRange(ctx.SourceFile, body).Pos()
	if closingParen >= bodyStart || bodyStart > len(sourceText) {
		return "", false
	}
	betweenParensAndBody := sourceText[closingParen:bodyStart]

	nameText, readable := classLiteralNameText(ctx, name)
	if !readable {
		return "", false
	}
	argumentText, argumentReadable := classLiteralNodeText(ctx, argument)
	if !argumentReadable {
		return "", false
	}

	var builder strings.Builder
	builder.WriteString(classLiteralPrintModifiers(node, "readonly"))
	builder.WriteString(nameText)
	builder.WriteString(betweenParensAndBody)
	builder.WriteString("= ")
	builder.WriteString(argumentText)
	builder.WriteString(";")
	return builder.String(), true
}

// classLiteralClosingParenBefore finds the offset just past the `)` closing a member's parameter
// list, by scanning back from the body.
//
// Upstream asks `sourceCode.getTokenBefore(returnType ?? body)`, which is a token-stream question
// this tree has no equivalent for. Scanning the raw text backwards from the body is the same answer
// for well-formed source and it degrades to a decline rather than a panic on anything else, which
// is what the boolean on the caller is for.
//
// Returning an offset PAST the paren rather than at it, matching upstream's `closingParen.range[1]`.
func classLiteralClosingParenBefore(ctx rule.Context, node *ast.Node, body *ast.Node) int {
	sourceText := ctx.SourceFile.Text()
	memberRange := rule.TokenRange(ctx.SourceFile, node)
	bodyStart := type_checking.TrimNodeTextRange(ctx.SourceFile, body).Pos()
	if bodyStart > len(sourceText) {
		return -1
	}

	for offset := bodyStart - 1; offset >= memberRange.Pos(); offset-- {
		if sourceText[offset] == ')' {
			return offset + 1
		}
	}
	return -1
}

// reportClassLiteralFieldsShouldBeGetters is upstream's ClassBody enter/exit pair under `getters`.
//
// Upstream collects properties on the way down and judges them on the way out, because a property
// assigned in the constructor is exempt and the constructor can be written after it. There is no
// exit hook here, so the class's members are read in one pass: the exclusion set is built from every
// constructor first, then the properties are judged against it. Same two phases, one traversal.
func reportClassLiteralFieldsShouldBeGetters(ctx rule.Context, members []*ast.Node) {
	excluded := map[string]bool{}
	for _, member := range members {
		if member.Kind != ast.KindConstructor {
			continue
		}
		classLiteralCollectConstructorAssignments(ctx, member, excluded)
	}

	for _, member := range members {
		if member.Kind != ast.KindPropertyDeclaration {
			continue
		}
		property := member.AsPropertyDeclaration()

		// Upstream's three guards, in upstream's order. A `declare` field has no initializer at
		// runtime, and an `override` field cannot become a getter without changing what it overrides.
		if !classLiteralHasModifier(member, ast.KindReadonlyKeyword) {
			continue
		}
		if classLiteralHasModifier(member, ast.KindDeclareKeyword) {
			continue
		}
		if classLiteralHasModifier(member, ast.KindOverrideKeyword) {
			continue
		}
		// An unmarked override whose getter would not compile. Deliberately quieter than upstream.
		if classLiteralConversionBreaksAnOverride(ctx, member, false) {
			continue
		}

		value := property.Initializer
		if !classLiteralIsSupportedLiteral(value) {
			continue
		}

		name := property.Name()
		if name == nil {
			continue
		}

		if key, keyed := classLiteralStaticMemberAccessValue(ctx, name); keyed && excluded[key] {
			continue
		}

		nameText, readable := classLiteralNameText(ctx, name)
		if !readable {
			continue
		}
		valueText, valueReadable := classLiteralNodeText(ctx, value)
		if !valueReadable {
			continue
		}

		// A decorated field reports with no suggestion, mirroring the getter direction above. This
		// one is this port's decision rather than upstream's: upstream offers the suggestion, and it
		// replaces the whole member from the decorator on while `classLiteralPrintModifiers` carries
		// only accessibility and `static`, so `@dec public static readonly foo = 'x'` became
		// `public static get foo() { return 'x'; }` with the decorator gone. Carrying it would not be
		// right either, because a field decorator and an accessor decorator receive different
		// arguments, which is the same reason upstream withholds the repair in the other direction.
		if classLiteralHasDecorator(member) {
			ctx.ReportNode(classLiteralReportedKey(name), messageClassLiteralPropertyStylePreferGetter())
			continue
		}

		replacement := classLiteralPrintModifiers(member, "get") + nameText +
			"() { return " + valueText + "; }"

		ctx.ReportNodeWithSuggestions(classLiteralReportedKey(name), messageClassLiteralPropertyStylePreferGetter(),
			rule.Suggestion{
				Message: messageClassLiteralPropertyStylePreferGetterSuggestion(),
				Fixes: []rule.Fix{
					rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, member), replacement),
				},
			})
	}
}

// classLiteralCollectConstructorAssignments records every `this.x = ...` written in a constructor.
//
// Upstream's selector is `MethodDefinition[kind="constructor"] ThisExpression` followed by a walk up
// to the nearest function, checking that function IS the constructor. That walk exists to decline an
// assignment inside a nested function, whose `this` need not be the instance. The same question is
// answered here by walking down and refusing to descend into a nested function-like, which reaches
// the same set without the upward walk.
func classLiteralCollectConstructorAssignments(ctx rule.Context, constructor *ast.Node, excluded map[string]bool) {
	body := constructor.AsConstructorDeclaration().Body
	if body == nil {
		return
	}

	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		// A nested function gets its own `this`, so an assignment inside one is not an assignment
		// to this instance. An arrow function does NOT rebind `this`, so it is walked into.
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor, ast.KindClassDeclaration, ast.KindClassExpression:
			return
		}

		if node.Kind == ast.KindBinaryExpression {
			binary := node.AsBinaryExpression()
			if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
				if key, keyed := classLiteralThisPropertyKey(ctx, binary.Left); keyed {
					excluded[key] = true
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	body.ForEachChild(func(child *ast.Node) bool {
		walk(child)
		return false
	})
}

// classLiteralThisPropertyKey reads the property name off a `this.x` or `this['x']` target.
func classLiteralThisPropertyKey(ctx rule.Context, target *ast.Node) (string, bool) {
	if target == nil {
		return "", false
	}

	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		access := target.AsPropertyAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindThisKeyword {
			return "", false
		}
		if access.Name() == nil {
			return "", false
		}
		return access.Name().Text(), true

	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindThisKeyword {
			return "", false
		}
		return classLiteralStaticValueOf(ctx, access.ArgumentExpression)
	}

	return "", false
}

// classLiteralHasMatchingSetter is upstream's `hasDuplicateKeySetter`.
//
// The comparison is on the static member access VALUE rather than on the source text, so a getter
// keyed `x` pairs with a setter keyed `'x'` or `["x"]`.
func classLiteralHasMatchingSetter(ctx rule.Context, getter *ast.Node, name *ast.Node) bool {
	key, keyed := classLiteralStaticMemberAccessValue(ctx, name)
	if !keyed {
		return false
	}

	parent := getter.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
	default:
		return false
	}

	for _, member := range parent.Members() {
		if member.Kind != ast.KindSetAccessor {
			continue
		}
		memberName := member.Name()
		if memberName == nil {
			continue
		}
		if otherKey, otherKeyed := classLiteralStaticMemberAccessValue(ctx, memberName); otherKeyed && otherKey == key {
			return true
		}
	}
	return false
}

// classLiteralStaticMemberAccessValue is upstream's `getStaticMemberAccessValue`.
//
// A non-computed identifier or private name answers its own text. Anything else is evaluated as a
// static value, which is how a computed key written as a literal reaches the same namespace as the
// bare spelling.
//
// Upstream's `getStaticValue` evaluates arbitrary constant expressions against scope; this answers
// only literals, which is every shape the corpus and this tree write. The narrowing is stated rather
// than silent: a computed key like `[SOME_CONSTANT]` is keyed by upstream and declined here, which
// costs an exemption rather than adding a finding, so the failure direction is silence.
func classLiteralStaticMemberAccessValue(ctx rule.Context, key *ast.Node) (string, bool) {
	if key == nil {
		return "", false
	}

	switch key.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return key.Text(), true
	case ast.KindComputedPropertyName:
		return classLiteralStaticValueOf(ctx, key.AsComputedPropertyName().Expression)
	}

	return classLiteralStaticValueOf(ctx, key)
}

// classLiteralStaticValueOf evaluates the shapes a key can take, upstream's `getStaticValue`.
//
// Two arms. A literal answers its own cooked text. An IDENTIFIER is resolved through a `const` or
// `let` declaration in the same file whose initializer is itself a literal, which is the subset of
// upstream's scope-backed evaluation this rule actually needs.
//
// The identifier arm is not optional and an earlier draft without it failed two of upstream's own
// clean cases. Measured on the installed build, with the fifth row as the control that stops the
// resolution from being written too widely:
//
//	let p1 = 'p1';   set [p1](v) {} get [p1]() {...}                CLEAN, one variable, both keys
//	const p1 = 'k'; const p2 = 'k';  set [p2] / get [p1]            CLEAN, two names, one value
//	const p1 = 'a'; const p2 = 'b';  set [p2] / get [p1]            REPORTS, two values
//	const p1 = 'x';  set x(v) {} get [p1]() {...}                   CLEAN, across spellings
//	                 set [foo()](v) {} get [foo()]() {...}          REPORTS, unevaluable both sides
//
// The fourth is why a syntactic name comparison is wrong: upstream pairs a computed key with a bare
// one when the values agree, and no amount of matching source text reaches that. The fifth is why
// the resolution must not fall back to comparing expressions: two identical unevaluable keys do NOT
// pair upstream.
//
// # What this declines, stated rather than left silent
//
// Upstream evaluates arbitrary constant expressions against scope, so `[PREFIX + 'x']` and a member
// of a `const enum` are keyed there and are not here. Both cost an EXEMPTION rather than adding a
// finding, so the failure direction is a false positive on a getter that upstream would have paired
// with its setter. No case in the corpus and no shape measured in this tree reaches it. Closing the
// gap means a constant evaluator with scope, which is substrate rather than a line in this rule.
func classLiteralStaticValueOf(ctx rule.Context, expression *ast.Node) (string, bool) {
	if expression == nil {
		return "", false
	}
	switch expression.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return expression.Text(), true
	case ast.KindIdentifier:
		return classLiteralResolveVariableLiteral(ctx, expression.Text())
	}
	return "", false
}

// classLiteralResolveVariableLiteral finds a file-scope variable of this name whose initializer is a
// literal, and answers that literal's value.
//
// Deliberately narrow, and the narrowness is the point rather than a shortcut. The question upstream
// asks is "what value does this key have", and the only shape this rule has ever been shown to meet
// is a variable initialized to a literal. Anything else declines, which loses an exemption rather
// than inventing a finding.
//
// The LAST matching declaration wins, matching how a reader would read the file and how a second
// declaration of the same name shadows the first at the point the class is written. No corpus case
// writes two, so this is stated rather than measured, and it is stated because the alternative is a
// reader assuming the first wins.
func classLiteralResolveVariableLiteral(ctx rule.Context, name string) (string, bool) {
	if ctx.SourceFile == nil {
		return "", false
	}

	value := ""
	found := false
	for _, statement := range ctx.SourceFile.AsNode().Statements() {
		if statement.Kind != ast.KindVariableStatement {
			continue
		}
		list := statement.AsVariableStatement().DeclarationList
		if list == nil {
			continue
		}
		for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
			declarationName := declaration.Name()
			if declarationName == nil || declarationName.Kind != ast.KindIdentifier {
				continue
			}
			if declarationName.Text() != name {
				continue
			}
			initializer := declaration.AsVariableDeclaration().Initializer
			if initializer == nil {
				continue
			}
			switch initializer.Kind {
			case ast.KindStringLiteral, ast.KindNumericLiteral,
				ast.KindNoSubstitutionTemplateLiteral:
				value, found = initializer.Text(), true
			}
		}
	}
	return value, found
}

// classLiteralIsSupportedLiteral is upstream's `isSupportedLiteral`.
//
// A literal, a template with exactly one quasi (so no substitutions), or a tagged template whose
// quasi likewise has one. The single-quasi test is what rules out a template holding a
// substitution, which is not a constant.
//
// Upstream's `Literal` covers a string, number, boolean, null, regex and bigint. TSESTree folds all
// of those into one node type and typescript-go does not, so the arms are enumerated here. `true`,
// `false` and `null` are keyword nodes rather than literals in this tree, which is a parser
// difference rather than a narrowing; all three are measured as reporting upstream.
func classLiteralIsSupportedLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindNoSubstitutionTemplateLiteral:
		return true

	case ast.KindTemplateExpression:
		// A template carrying substitutions has more than one quasi upstream and is unsupported.
		// typescript-go only produces a TemplateExpression when there IS a substitution, so this
		// arm exists to decline rather than to accept, and it is written out because the intuitive
		// reading is that a template literal is a literal.
		//
		// Rerouting this LABEL to another kind survives the sweep, and that survival is inert
		// rather than a fixture gap: the arm returns false and so does the default below it, so an
		// unreachable label and the real behavior produce identical output. Mutating the arm's
		// EFFECT settles it instead, and returning true here fails seven lines. This is the label
		// trap the porting brief names, met on the one arm in this rule whose whole content is a
		// refusal.
		return false

	case ast.KindTaggedTemplateExpression:
		// Upstream tests `node.quasi.quasis.length === 1`, so a tagged template is supported exactly
		// when its template carries no substitution.
		return node.AsTaggedTemplateExpression().Template != nil &&
			node.AsTaggedTemplateExpression().Template.Kind == ast.KindNoSubstitutionTemplateLiteral
	}

	return false
}

// classLiteralPrintModifiers is upstream's `printNodeModifiers`.
//
// Upstream concatenates the accessibility keyword, then `static` when present, then the final
// keyword, and trims the leading space that a missing accessibility leaves behind. So a member with
// neither modifier produces `readonly ` and one with both produces `public static readonly `. Only accessibility and `static` are carried: `abstract`, `declare` and
// `override` are all excluded by a guard before this is reached, and `readonly` on a field is being
// replaced rather than preserved.
func classLiteralPrintModifiers(node *ast.Node, final string) string {
	var builder strings.Builder

	for _, modifier := range classLiteralModifierNodes(node) {
		switch modifier.Kind {
		case ast.KindPublicKeyword:
			builder.WriteString("public")
		case ast.KindPrivateKeyword:
			builder.WriteString("private")
		case ast.KindProtectedKeyword:
			builder.WriteString("protected")
		}
	}

	if classLiteralHasModifier(node, ast.KindStaticKeyword) {
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString("static")
	}

	if builder.Len() > 0 {
		builder.WriteString(" ")
	}
	builder.WriteString(final)
	builder.WriteString(" ")
	return builder.String()
}

// classLiteralModifierNodes returns a member's modifiers, or nothing.
func classLiteralModifierNodes(node *ast.Node) []*ast.Node {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return nil
	}
	return modifiers.Nodes
}

// classLiteralHasModifier answers whether a member carries one modifier keyword.
func classLiteralHasModifier(node *ast.Node, keyword ast.Kind) bool {
	for _, modifier := range classLiteralModifierNodes(node) {
		if modifier.Kind == keyword {
			return true
		}
	}
	return false
}

// classLiteralHasDecorator answers whether a member carries any decorator.
//
// typescript-go puts decorators in the same modifier list as `public` and `static`, so this reads
// the list rather than a separate field the way TSESTree's `node.decorators` does.
func classLiteralHasDecorator(node *ast.Node) bool {
	return classLiteralHasModifier(node, ast.KindDecorator)
}

// classLiteralNameText renders a member's key the way upstream's suggestion writes it.
//
// Upstream reads `sourceCode.getText(node.key)`, where TSESTree's `key` for a computed member is the
// INNER expression, and then re-adds the brackets with `node.computed ? `[${name}]` : name`. Our
// `Name()` for a computed member IS the bracketed node, so its own text already carries them and
// re-adding would double them. Same output, one fewer step.
func classLiteralNameText(ctx rule.Context, name *ast.Node) (string, bool) {
	return classLiteralNodeText(ctx, name)
}

// classLiteralReportedKey is the node a finding points at.
//
// Upstream reports `node.key`, and TSESTree's `key` for a computed member is the INNER expression
// with the brackets living on the member rather than on the key. Our `Name()` is the bracketed node,
// so reporting it directly spans one character more on each side than upstream does.
//
// Measured on the installed build: `static readonly [myValue] = 1;` reports columns covering
// `myValue`, not `[myValue]`. Three of upstream's own reporting cases assert that span and all three
// failed before this existed, which is what a span assertion is for: every message id was already
// correct.
func classLiteralReportedKey(name *ast.Node) *ast.Node {
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		if inner := name.AsComputedPropertyName().Expression; inner != nil {
			return inner
		}
	}
	return name
}

// classLiteralNodeText slices a node's own source text, trivia excluded.
func classLiteralNodeText(ctx rule.Context, node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	trimmed := type_checking.TrimNodeTextRange(ctx.SourceFile, node)
	sourceText := ctx.SourceFile.Text()
	if trimmed.Pos() < 0 || trimmed.End() > len(sourceText) || trimmed.Pos() > trimmed.End() {
		return "", false
	}
	return sourceText[trimmed.Pos():trimmed.End()], true
}

// ClassLiteralPropertyStyleSetting is upstream's one option, `'fields' | 'getters'`.
type ClassLiteralPropertyStyleSetting string

const (
	// ClassLiteralPropertyStyleFields is `"fields"`, upstream's default: a literal is a readonly
	// field and a getter returning a literal reports.
	ClassLiteralPropertyStyleFields ClassLiteralPropertyStyleSetting = "fields"

	// ClassLiteralPropertyStyleGetters is `"getters"`: a literal is a getter and a readonly field
	// holding one reports.
	ClassLiteralPropertyStyleGetters ClassLiteralPropertyStyleSetting = "getters"
)

// ClassLiteralPropertyStyleOptions is the rule's configuration.
//
// Upstream's schema is a bare STRING rather than an object, which is unusual enough to be worth
// naming: `schema: [{ type: 'string', enum: ['fields', 'getters'] }]` and
// `defaultOptions: ['fields']`.
type ClassLiteralPropertyStyleOptions struct {
	Style ClassLiteralPropertyStyleSetting
}

// DefaultClassLiteralPropertyStyleOptions is upstream's configured-nothing behavior.
//
// Exported because a fixture asserting the default has to be able to name it, and because a caller
// starting from the Go zero value would get an empty style, which matches neither arm and silences
// the rule on every input while looking configured.
func DefaultClassLiteralPropertyStyleOptions() ClassLiteralPropertyStyleOptions {
	return ClassLiteralPropertyStyleOptions{Style: ClassLiteralPropertyStyleFields}
}

// DecodeClassLiteralPropertyStyleOptions reads upstream's bare-string option.
//
// Hand-written rather than `rule.DecodeOptionsInto` for two reasons. The wire value is a JSON STRING
// rather than an object, so there is no struct for the generic helper to unmarshal into. And the
// default is `fields` rather than a Go zero value, so a generic decode would hand the rule an empty
// style that matches no arm and silences it while every fixture stayed green.
//
// Empty input and `null` are upstream's default. Anything else that is not one of the two spellings is
// refused, naming the value. This used to read an unknown spelling as the default, on the reasoning
// that a typo should not silence the rule; it did not silence it, it ran a style nobody wrote and
// loaded clean doing it. Upstream's schema is a string enum and refuses it at load, and the config
// layer now refuses a decoder's error by rule name, so refusing is what tells the author (#rfbha44).
func DecodeClassLiteralPropertyStyleOptions(raw []byte) (any, error) {
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return DefaultClassLiteralPropertyStyleOptions(), nil
	}
	var wire string
	if err := json.Unmarshal(raw, &wire); err != nil {
		return DefaultClassLiteralPropertyStyleOptions(),
			fmt.Errorf(`expected a style string, "fields" or "getters", got %s`, trimmed)
	}
	switch ClassLiteralPropertyStyleSetting(wire) {
	case ClassLiteralPropertyStyleFields:
		return ClassLiteralPropertyStyleOptions{Style: ClassLiteralPropertyStyleFields}, nil
	case ClassLiteralPropertyStyleGetters:
		return ClassLiteralPropertyStyleOptions{Style: ClassLiteralPropertyStyleGetters}, nil
	}
	return DefaultClassLiteralPropertyStyleOptions(), fmt.Errorf(`style %q is not "fields" or "getters"`, wire)
}
