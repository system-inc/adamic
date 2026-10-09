package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// UnboundMethodOptions is the rule's option surface.
type UnboundMethodOptions struct {
	// IgnoreStatic skips a `static` method, which cannot lose an instance it never had.
	//
	// It defaults to FALSE, so the generic decoder would happen to produce the right zero value. The
	// decoder below is hand-rolled anyway: relying on that coincidence leaves the next person to add
	// a default-true key here inheriting a decoder that silently inverts the rule.
	IgnoreStatic bool
}

// DefaultUnboundMethodSettings is upstream's `defaultOptions`.
func DefaultUnboundMethodSettings() UnboundMethodOptions {
	return UnboundMethodOptions{IgnoreStatic: false}
}

// unboundMethodRawOptions is the wire shape, with a pointer so an absent key stays distinguishable
// from an explicit false.
type unboundMethodRawOptions struct {
	IgnoreStatic *bool `json:"ignoreStatic"`
}

// DecodeUnboundMethodOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple before dispatch, so what arrives
// is the bare object rather than upstream's one-element array.
func DecodeUnboundMethodOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[unboundMethodRawOptions]()(raw)
	if err != nil {
		return DefaultUnboundMethodSettings(), err
	}

	wire, _ := decoded.(unboundMethodRawOptions)
	options := DefaultUnboundMethodSettings()
	if wire.IgnoreStatic != nil {
		options.IgnoreStatic = *wire.IgnoreStatic
	}
	return options, nil
}

// UnboundMethod flags a method referenced without the object it needs.
//
//	valid:   instance.method();                     called, so `this` is supplied
//	valid:   const f = instance.arrowProperty;      an arrow property captures `this`
//	valid:   const f = o.method;  where o.method declares `this: void`
//	valid:   const x = Math.floor;                  a natively bound global
//	invalid: const f = instance.method;
//	invalid: promise.then(instance.method);
//	invalid: const { method } = instance;
//
// Detaching a method from its object is the oldest bug in the language: the reference survives, the
// binding does not, and the failure surfaces at call time somewhere else entirely. TypeScript's type
// system does not model the loss, so nothing but a rule catches it.
//
// # Two messages, and which one fires is a judgment rather than a detail
//
//	unbound                        the method already declares a `this` parameter of some type
//	unboundWithoutThisAnnotation   it declares none, so `this: void` would settle the question
//
// The second is longer because it can offer that advice. Fifty of upstream's sixty-nine recorded
// findings take it and nineteen the short form, so collapsing the two satisfies nothing.
//
// # What makes a reference safe, and why the list is syntactic
//
// A reference is safe when the surrounding syntax consumes it immediately rather than storing it:
// being called, tested in a condition, compared for identity, deleted, `typeof`-ed. The list is
// upstream's `isSafeUse` and it is reproduced position by position, because each entry is a place
// where the value never outlives the expression. Two entries recurse: an assertion wrapper and a
// logical expression pass the question up, since `a && o.m` still yields the method reference.
//
// # A natively bound global is exempt, and the NAME list is load bearing for five cases
//
// Upstream keeps a set of `Namespace.member` strings built by REFLECTING over the running Node
// global object at module load, then falls back to a type-level check. That reflection cannot be
// reproduced in a compiled binary, so the question was whether the type-level check alone suffices.
//
// Measured rather than assumed: the installed rule was patched to answer false from the name lookup
// and the whole corpus re-run. Exactly FIVE of two hundred and eleven verdicts moved, and three of
// them are `console`. Upstream's own comment says why: `console` is declared in `@types/node`
// rather than in the default library, so `isSymbolFromDefaultLibrary` answers false for it and only
// the name list saves it.
//
// So the list is reproduced as a frozen table generated from the same globals upstream reflects
// over, rather than dropped. It is a snapshot of one runtime and upstream's own comment says the
// membership varies by build, which is a real difference: upstream reads the interpreter it happens
// to be running on, and this reads a list captured once. Recorded rather than hidden.
//
// In this tree the divergence is mostly moot in the other direction: measured, `console` resolves to
// `any` here because our fixture programs load no node types, and an `any` object type declines
// before any of this is consulted.
//
// # This rule reads the PROGRAM, not just the checker
//
// `IsBuiltinSymbolLike` and `IsSymbolFromDefaultLibrary` both walk the program's own default-library
// files, and the not-imported test compares against the current source file. Those are reads outside
// the file being linted, so the findings cache must not key on that file alone.
//
// # Cost
//
// Two anchors, both common, and the syntactic safe-use test declines the overwhelming majority of
// property accesses in real code before a single type question is asked.
var UnboundMethod = rule.Rule{
	Name: "@typescript-eslint/unbound-method",

	// Every finding is decided by resolving a property to its declaration, so there is no syntactic
	// subset of this rule that could run without the checker.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[UnboundMethodOptions](options)
		if !isSettings {
			settings = DefaultUnboundMethodSettings()
		}

		return rule.Listeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				access := node.AsPropertyAccessExpression()
				checkUnboundMemberAccess(ctx, settings, node, access.Expression, access.Name())
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				access := node.AsElementAccessExpression()
				checkUnboundMemberAccess(ctx, settings, node, access.Expression, access.ArgumentExpression)
			},
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				checkUnboundDestructuring(ctx, settings, node)
			},
			// The SECOND half of upstream's one ObjectPattern visitor.
			//
			// Upstream's parser calls both `const { m } = o` and `({ m } = o)` an ObjectPattern, so
			// one visitor covers both. Ours only calls the first a binding pattern; a destructuring
			// ASSIGNMENT target parses as an object LITERAL and reaches a different kind entirely.
			// Confirmed by running upstream's own parser on the two spellings and watching both come
			// back ObjectPattern.
			//
			// Without this listener the rule is silent on every assignment-form destructuring, which
			// is four of upstream's sixty-nine findings and, unlike most gaps, costs findings rather
			// than adding them.
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				checkUnboundDestructuringAssignment(ctx, settings, node)
			},
		}
	},
}

var unboundMethodBaseMessage = "A method that is not declared with `this: void` may cause " +
	"unintentional scoping of `this` when separated from its object.\n" +
	"Consider using an arrow function or explicitly `.bind()`ing the method to avoid calling the " +
	"method with an unintended `this` value. "

var unboundMethodUnboundMessage = rule.Message{
	Id:          "unbound",
	Description: unboundMethodBaseMessage,
}

var unboundMethodWithoutThisAnnotationMessage = rule.Message{
	Id: "unboundWithoutThisAnnotation",
	Description: unboundMethodBaseMessage +
		"\nIf a function does not access `this`, it can be annotated with `this: void`.",
}

// checkUnboundMemberAccess is upstream's `MemberExpression` visitor.
//
// Our parser splits a member expression into two kinds by whether the subscript is computed, so the
// one upstream visitor becomes two listeners over a shared body. That split is convenient rather than
// awkward: the computed branch is the only one that has to ask the checker what the subscript
// evaluates to, and the dotted branch reads its name straight off the node.
func checkUnboundMemberAccess(
	ctx rule.Context,
	settings UnboundMethodOptions,
	node *ast.Node,
	object *ast.Node,
	property *ast.Node,
) {
	if object == nil || property == nil {
		return
	}
	if isSafeUnboundUse(node) || isNativelyBoundMember(ctx, object, property) {
		return
	}

	propertyNames := accessedPropertyNames(ctx, node, property)
	if len(propertyNames) == 0 {
		return
	}

	objectType := ctx.TypeChecker.GetTypeAtLocation(object)
	if objectType == nil {
		return
	}
	for _, propertyName := range propertyNames {
		if reportUnboundConstituent(ctx, settings, node, propertyName, objectType) {
			break
		}
	}
}

// checkUnboundDestructuring is upstream's `ObjectPattern` visitor.
//
// Pulling a method out of an object by destructuring detaches it exactly as a member access does, and
// upstream reports on the KEY rather than on the whole pattern, so `const { a, method } = o` points at
// `method` alone.
func checkUnboundDestructuring(ctx rule.Context, settings UnboundMethodOptions, node *ast.Node) {
	if isInsideTypeDeclaration(node) {
		return
	}

	// Upstream reads the initializer from three parent shapes. Our parser reaches the first through
	// a variable declaration and the other two through a binary expression, since it files a default
	// value and an assignment under kinds rather than under distinct node types.
	var initializer *ast.Node
	isAssignmentPattern := false
	if parent := node.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindVariableDeclaration:
			initializer = parent.AsVariableDeclaration().Initializer
		case ast.KindParameter:
			// `function ({ m }: Foo = NativeObject) {}`. Upstream calls this an AssignmentPattern
			// and treats it specially below, because the default is only one of two possible
			// sources for the value.
			initializer = parent.AsParameterDeclaration().Initializer
			isAssignmentPattern = initializer != nil
		case ast.KindBindingElement:
			initializer = parent.AsBindingElement().Initializer
			isAssignmentPattern = initializer != nil
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken != nil &&
				binary.OperatorToken.Kind == ast.KindEqualsToken &&
				binary.Left == node {
				initializer = binary.Right
			}
		}
	}

	patternType := ctx.TypeChecker.GetTypeAtLocation(node)

	for _, element := range node.AsBindingPattern().Elements.Nodes {
		binding := element.AsBindingElement()

		// Upstream requires the property KEY be an identifier. A shorthand binding has no separate
		// property name here, so the key is the binding's own name; a renamed one carries both.
		key := binding.PropertyName
		if key == nil {
			key = binding.Name()
		}
		if key == nil || key.Kind != ast.KindIdentifier {
			continue
		}
		keyName := key.AsIdentifier().Text

		if initializer != nil {
			if !isNativelyBoundMember(ctx, initializer, key) {
				initializerType := ctx.TypeChecker.GetTypeAtLocation(initializer)
				if initializerType != nil {
					property := checker.Checker_getPropertyOfType(ctx.TypeChecker, initializerType, keyName)
					if reportIfUnboundMethod(ctx, settings, key, property) {
						continue
					}
				}
			} else if !isAssignmentPattern {
				// A natively bound member reached through a plain initializer is done. Under an
				// assignment pattern the default is only one of the two values this binding can
				// take, so the declared type is still worth checking.
				continue
			}
		}

		if patternType != nil {
			reportUnboundConstituent(ctx, settings, key, keyName, patternType)
		}
	}
}

// reportUnboundConstituent is upstream's `checkUnionConstituentsAndReport`.
//
// A property can come from any constituent of a union of intersections, and finding it in one is
// enough. Upstream stops at the first constituent that reports, which is what keeps a union of two
// unbound shapes from producing two findings on one reference.
func reportUnboundConstituent(
	ctx rule.Context,
	settings UnboundMethodOptions,
	reportNode *ast.Node,
	propertyName string,
	subjectType *checker.Type,
) bool {
	for unionPart := range type_checking.UnionTypePartsSeq(subjectType) {
		for _, intersectionPart := range type_checking.IntersectionTypeParts(unionPart) {
			property := checker.Checker_getPropertyOfType(ctx.TypeChecker, intersectionPart, propertyName)
			if reportIfUnboundMethod(ctx, settings, reportNode, property) {
				return true
			}
		}
	}
	return false
}

// unboundThisParameterState is upstream's `firstParamIsThis`, which is a TRI-state.
//
// Upstream types it `boolean | undefined` and its message test is `firstParamIsThis === false`, a
// strict comparison. So the two boolean values are not the whole story: a declaration that was never
// examined for a `this` parameter leaves the field undefined, and undefined is not false, so it takes
// the SHORT message rather than the long one.
//
// Modelling this as a Go bool is the natural port and it is wrong. It cost twenty-one of upstream's
// sixty-nine findings, every one of them reporting the right node with the wrong message.
type unboundThisParameterState int

const (
	// unboundThisParameterUnexamined is upstream's `undefined`: the declaration was judged dangerous
	// without its parameter list ever being read, which happens for a class property holding a plain
	// function expression. Upstream reports the SHORT message here.
	unboundThisParameterUnexamined unboundThisParameterState = iota
	// unboundThisParameterPresent is upstream's `true`: a first parameter named `this` exists and is
	// typed as something other than void, so the author has already thought about `this`.
	unboundThisParameterPresent
	// unboundThisParameterAbsent is upstream's `false`: the parameter list was read and carries no
	// `this`, which is the only state that earns the longer message offering the annotation.
	unboundThisParameterAbsent
)

// reportIfUnboundMethod is upstream's `checkIfMethodAndReport`.
func reportIfUnboundMethod(
	ctx rule.Context,
	settings UnboundMethodOptions,
	reportNode *ast.Node,
	property *ast.Symbol,
) bool {
	if property == nil {
		return false
	}

	dangerous, thisParameter := isDangerousUnboundMethod(property, settings.IgnoreStatic)
	if !dangerous {
		return false
	}

	// Upstream's `firstParamIsThis === false ? long : short`, with the strict comparison preserved:
	// only the examined-and-absent state takes the long form.
	message := unboundMethodUnboundMessage
	if thisParameter == unboundThisParameterAbsent {
		message = unboundMethodWithoutThisAnnotationMessage
	}
	ctx.ReportNode(reportNode, message)
	return true
}

// isDangerousUnboundMethod is upstream's `checkIfMethod`.
//
// The switch is over the property's VALUE DECLARATION, which is what separates a method from an
// arrow-function property that captured `this` at construction. Measured against our checker: a class
// method arrives as KindMethodDeclaration, an interface or type-literal method as KindMethodSignature,
// an arrow or function property as KindPropertyDeclaration, and an object-literal member as
// KindPropertyAssignment, matching upstream's four cases one for one.
func isDangerousUnboundMethod(
	property *ast.Symbol,
	ignoreStatic bool,
) (bool, unboundThisParameterState) {
	valueDeclaration := property.ValueDeclaration
	if valueDeclaration == nil {
		// Upstream works around a compiler issue here and declines rather than guessing.
		return false, unboundThisParameterUnexamined
	}

	switch valueDeclaration.Kind {
	case ast.KindPropertyDeclaration:
		// A class property is dangerous only when it holds a plain `function` expression, which has
		// its own `this`. An arrow function captured the instance and is safe.
		//
		// The parameter list is deliberately NOT read here, matching upstream, which returns only
		// `{ dangerous }` from this arm and leaves `firstParamIsThis` undefined. That is why
		// `class Foo { unbound = function () {} }` takes the short message while
		// `class Foo { unbound() {} }` takes the long one, on the same reference.
		initializer := valueDeclaration.AsPropertyDeclaration().Initializer
		dangerous := initializer != nil && initializer.Kind == ast.KindFunctionExpression
		return dangerous, unboundThisParameterUnexamined

	case ast.KindPropertyAssignment:
		initializer := valueDeclaration.AsPropertyAssignment().Initializer
		if initializer == nil || initializer.Kind != ast.KindFunctionExpression {
			return false, unboundThisParameterUnexamined
		}
		return methodIsDangerous(initializer, initializer.AsFunctionExpression().Parameters, ignoreStatic)

	case ast.KindMethodDeclaration:
		return methodIsDangerous(valueDeclaration,
			valueDeclaration.AsMethodDeclaration().Parameters, ignoreStatic)

	case ast.KindMethodSignature:
		return methodIsDangerous(valueDeclaration,
			valueDeclaration.AsMethodSignatureDeclaration().Parameters, ignoreStatic)
	}

	return false, unboundThisParameterUnexamined
}

// methodIsDangerous is upstream's `checkMethod`.
//
// A method is safe only when it explicitly declares `this: void`, or when it is static and the option
// says to skip those. Declaring `this` as anything ELSE is still dangerous, and that is what selects
// the shorter of the two messages: the author has already thought about `this` here, so telling them
// to annotate it would be wrong.
func methodIsDangerous(
	declaration *ast.Node,
	parameters *ast.NodeList,
	ignoreStatic bool,
) (bool, unboundThisParameterState) {
	// This arm DID read the parameter list, so its answer is never the unexamined state. That is the
	// whole distinction between here and the property-declaration arm above.
	thisParameter := unboundThisParameterAbsent
	thisParameterIsVoid := false

	if parameters != nil && len(parameters.Nodes) > 0 {
		first := parameters.Nodes[0].AsParameterDeclaration()
		name := first.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "this" {
			thisParameter = unboundThisParameterPresent
			thisParameterIsVoid = first.Type != nil && first.Type.Kind == ast.KindVoidKeyword
		}
	}

	if thisParameterIsVoid {
		return false, thisParameter
	}
	if ignoreStatic && ast.IsStatic(declaration) {
		return false, thisParameter
	}
	return true, thisParameter
}

// accessedPropertyNames is upstream's `getAccessedPropertyNames`.
//
// A dotted access names one property. A computed one names as many as its subscript type has literal
// constituents, so `o[cond ? 'a' : 'b']` asks about both, and a subscript that is not a literal names
// none and the rule declines.
func accessedPropertyNames(ctx rule.Context, node *ast.Node, property *ast.Node) []string {
	if node.Kind == ast.KindPropertyAccessExpression {
		if property.Kind == ast.KindIdentifier {
			return []string{property.AsIdentifier().Text}
		}
		// A private name is not a property upstream can look up by string.
		return nil
	}

	subscriptType := ctx.TypeChecker.GetTypeAtLocation(property)
	if subscriptType == nil {
		return nil
	}
	var names []string
	for part := range type_checking.UnionTypePartsSeq(subscriptType) {
		if !type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral) {
			continue
		}
		// The checker renders a string literal type quoted and a numeric one bare, so the quotes
		// have to come off to compare against a property name. Upstream reads `part.value` directly,
		// which is already the cooked value.
		rendered := ctx.TypeChecker.TypeToString(part)
		if len(rendered) >= 2 && (rendered[0] == '"' || rendered[0] == '\'') &&
			rendered[len(rendered)-1] == rendered[0] {
			rendered = rendered[1 : len(rendered)-1]
		}
		names = append(names, rendered)
	}
	return names
}

// isNativelyBoundMember is upstream's `isNativelyBound`.
//
// Two routes, and the first exists because the second is not enough. Upstream's comment says a
// declaration can come from `@types/node` rather than from the default library, in which case the
// type-level test answers false for something that is in fact natively bound.
func isNativelyBoundMember(ctx rule.Context, object *ast.Node, property *ast.Node) bool {
	if object.Kind == ast.KindIdentifier && property != nil && property.Kind == ast.KindIdentifier {
		objectSymbol := ctx.TypeChecker.GetSymbolAtLocation(object)
		if objectSymbol != nil && isSymbolNotFromThisFile(ctx, objectSymbol) {
			if nativelyBoundMembers[object.AsIdentifier().Text+"."+property.AsIdentifier().Text] {
				return true
			}
		}
	}

	objectType := ctx.TypeChecker.GetTypeAtLocation(object)
	if objectType == nil {
		return false
	}
	if !type_checking.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, objectType,
		supportedGlobalTypeNames...) {
		return false
	}
	if property == nil {
		return false
	}
	propertyType := ctx.TypeChecker.GetTypeAtLocation(property)
	if propertyType == nil {
		return false
	}
	propertySymbol := checker.Type_symbol(propertyType)
	return propertySymbol != nil && type_checking.IsSymbolFromDefaultLibrary(ctx.Program, propertySymbol)
}

// isSymbolNotFromThisFile is upstream's `isNotImported`, named for what it measures.
//
// Upstream asks whether the symbol's declaration lives in a file OTHER than the one being linted, and
// calls that "not imported", which reads backwards: a global like `Math` is declared in a library
// file and is therefore "not imported" in this sense, while a locally declared `const Math = ...`
// shadows it and is. The name here says what the comparison does rather than what it implies.
func isSymbolNotFromThisFile(ctx rule.Context, symbol *ast.Symbol) bool {
	valueDeclaration := symbol.ValueDeclaration
	if valueDeclaration == nil {
		// Upstream works around the same compiler issue here and answers false.
		return false
	}
	declarationFile := ast.GetSourceFileOfNode(valueDeclaration)
	return declarationFile != nil && ctx.SourceFile != nil && declarationFile != ctx.SourceFile
}

// isInsideTypeDeclaration is upstream's `isNodeInsideTypeDeclaration`.
//
// A destructuring pattern written inside a type has no runtime binding to lose, so the rule declines.
// Upstream reads a `declare` flag on two of these; ours carries the ambient flag on the node itself
// and propagates it, so the flag test covers both of those cases and the nested ones upstream's
// ancestor walk was written to reach.
func isInsideTypeDeclaration(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindFunctionType,
			ast.KindMethodSignature:
			return true
		case ast.KindClassDeclaration, ast.KindVariableStatement, ast.KindFunctionDeclaration:
			if parent.Flags&ast.NodeFlagsAmbient != 0 {
				return true
			}
		case ast.KindMethodDeclaration:
			if ast.HasSyntacticModifier(parent, ast.ModifierFlagsAbstract) {
				return true
			}
		}
	}
	return false
}

// isSafeUnboundUse is upstream's `isSafeUse`.
//
// Every entry is a syntactic position where the reference is consumed rather than stored, so the
// binding it lost never matters. The list is upstream's and it is reproduced position by position.
func isSafeUnboundUse(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	// The reference is immediately tested or stepped, never retained.
	case ast.KindIfStatement,
		ast.KindForStatement,
		ast.KindSwitchStatement,
		ast.KindWhileStatement,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindPostfixUnaryExpression:
		return true

	case ast.KindCallExpression:
		// Safe only as the CALLEE. Passed as an argument it is being detached, which is the whole
		// shape this rule exists to catch.
		return parent.AsCallExpression().Expression == node

	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node

	case ast.KindTaggedTemplateExpression:
		return parent.AsTaggedTemplateExpression().Tag == node

	case ast.KindPrefixUnaryExpression:
		// `!o.m` and `++o.m`. Upstream's list is on one UnaryExpression node type; ours splits the
		// update operators onto their own kinds, and `delete`, `typeof` and `void` onto three more.
		operator := parent.AsPrefixUnaryExpression().Operator
		return operator == ast.KindExclamationToken ||
			operator == ast.KindPlusPlusToken ||
			operator == ast.KindMinusMinusToken

	case ast.KindDeleteExpression, ast.KindTypeOfExpression, ast.KindVoidExpression:
		return true

	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		// An identity or instance comparison consumes the reference without calling it.
		case ast.KindExclamationEqualsToken,
			ast.KindExclamationEqualsEqualsToken,
			ast.KindEqualsEqualsToken,
			ast.KindEqualsEqualsEqualsToken,
			ast.KindInstanceOfKeyword:
			return true

		case ast.KindEqualsToken:
			// Assigning TO the member is defining it rather than detaching it. The second arm is
			// upstream's `super.m = this.m` carve-out, where the right side is a super member
			// access being installed on the instance.
			if binary.Left == node {
				return true
			}
			return isSuperMemberAccess(node) && isThisMemberAccess(binary.Left)

		case ast.KindAmpersandAmpersandToken:
			// `o.m && ...` is safe on the LEFT, because `&&` yields the left operand only when it
			// is falsy and a method reference never is. Anywhere else the expression can still
			// produce the reference, so the question passes up.
			if binary.Left == node {
				return true
			}
			return isSafeUnboundUse(parent)

		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return isSafeUnboundUse(parent)
		}
		return false

	// An assertion or a non-null wrapper does not consume the reference, it relabels it, so the
	// question passes through. A parenthesis is a real node here and is folded away by upstream's
	// parser, so upstream has no arm for it; without one, every parenthesized safe use reports.
	case ast.KindNonNullExpression,
		ast.KindAsExpression,
		ast.KindTypeAssertionExpression,
		ast.KindSatisfiesExpression,
		ast.KindParenthesizedExpression:
		return isSafeUnboundUse(parent)
	}

	return false
}

// isSuperMemberAccess answers whether a node is `super.something`.
func isSuperMemberAccess(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	return node.AsPropertyAccessExpression().Expression.Kind == ast.KindSuperKeyword
}

// isThisMemberAccess answers whether a node is `this.something`.
func isThisMemberAccess(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	return node.AsPropertyAccessExpression().Expression.Kind == ast.KindThisKeyword
}

// supportedGlobalTypeNames is upstream's `SUPPORTED_GLOBAL_TYPES`.
//
// These are the TYPE names, which differ from the value names above: the value `Number` has the type
// `NumberConstructor`, while `Math` and `JSON` are their own type names.
var supportedGlobalTypeNames = []string{
	"NumberConstructor",
	"ObjectConstructor",
	"StringConstructor",
	"SymbolConstructor",
	"ArrayConstructor",
	"Array",
	"ProxyConstructor",
	"Console",
	"DateConstructor",
	"Atomics",
	"Math",
	"JSON",
}

// checkUnboundDestructuringAssignment is the assignment half of upstream's `ObjectPattern` visitor.
//
// `({ m } = o)` detaches `m` exactly as `const { m } = o` does, and upstream reaches both through one
// node type. Ours parses the assignment target as an object literal, so the shape has to be
// recognized by POSITION: an object literal directly on the left of an assignment is a destructuring
// target, and one anywhere else is an ordinary object being constructed.
//
// The initializer is the assignment's right side, which is upstream's `node.parent.right` arm.
func checkUnboundDestructuringAssignment(ctx rule.Context, settings UnboundMethodOptions, node *ast.Node) {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return
	}
	binary := parent.AsBinaryExpression()
	if binary.OperatorToken == nil ||
		binary.OperatorToken.Kind != ast.KindEqualsToken ||
		binary.Left != node {
		return
	}
	if isInsideTypeDeclaration(node) {
		return
	}

	initializer := binary.Right
	patternType := ctx.TypeChecker.GetTypeAtLocation(node)

	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		// Upstream requires the property be a Property whose key is an identifier. A shorthand
		// assignment target is its own kind here and carries the name directly; a renamed one is a
		// property assignment whose NAME is the key being read off the source object.
		var key *ast.Node
		switch property.Kind {
		case ast.KindShorthandPropertyAssignment:
			key = property.AsShorthandPropertyAssignment().Name()
		case ast.KindPropertyAssignment:
			key = property.AsPropertyAssignment().Name()
		default:
			continue
		}
		if key == nil || key.Kind != ast.KindIdentifier {
			continue
		}
		keyName := key.AsIdentifier().Text

		if initializer != nil {
			if !isNativelyBoundMember(ctx, initializer, key) {
				initializerType := ctx.TypeChecker.GetTypeAtLocation(initializer)
				if initializerType != nil {
					found := checker.Checker_getPropertyOfType(ctx.TypeChecker, initializerType, keyName)
					if reportIfUnboundMethod(ctx, settings, key, found) {
						continue
					}
				}
			} else {
				// An assignment has no default value, so unlike the binding-pattern half there is no
				// second source for this value and a natively bound member is done.
				continue
			}
		}

		if patternType != nil {
			reportUnboundConstituent(ctx, settings, key, keyName, patternType)
		}
	}
}
