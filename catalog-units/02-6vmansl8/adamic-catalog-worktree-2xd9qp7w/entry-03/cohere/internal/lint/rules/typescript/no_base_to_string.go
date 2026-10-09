package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noBaseToStringUsefulness is upstream's `Usefulness` enum.
//
// The three values are a lattice rather than a flag, and the string spellings are load bearing: they
// are interpolated straight into the message, so `Sometimes` renders as "may" and `Never` as "will".
type noBaseToStringUsefulness string

const (
	// noBaseToStringAlways means the value stringifies usefully, so nothing is reported.
	noBaseToStringAlways noBaseToStringUsefulness = "always"
	// noBaseToStringNever means it definitely does not.
	noBaseToStringNever noBaseToStringUsefulness = "will"
	// noBaseToStringSometimes means some constituent does and some does not.
	noBaseToStringSometimes noBaseToStringUsefulness = "may"
)

// NoBaseToStringOptions is the rule's option surface.
type NoBaseToStringOptions struct {
	// CheckUnknown extends the rule to `unknown` and to an unconstrained type parameter.
	CheckUnknown bool
	// IgnoredTypeNames are stringified type names to leave alone.
	//
	// This is the one option in this family whose default is NOT a zero value, which makes the
	// decoder load bearing rather than defensive: a decoder returning an empty slice for an absent
	// key would silently start reporting on Error, RegExp, URL and URLSearchParams.
	IgnoredTypeNames []string
}

// DefaultNoBaseToStringSettings is upstream's `defaultOptions`.
func DefaultNoBaseToStringSettings() NoBaseToStringOptions {
	return NoBaseToStringOptions{
		CheckUnknown:     false,
		IgnoredTypeNames: []string{"Error", "RegExp", "URL", "URLSearchParams"},
	}
}

// noBaseToStringRawOptions is the wire shape, with pointers so an absent key stays distinguishable
// from an explicit empty one.
type noBaseToStringRawOptions struct {
	CheckUnknown     *bool     `json:"checkUnknown"`
	IgnoredTypeNames *[]string `json:"ignoredTypeNames"`
}

// DecodeNoBaseToStringOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple before dispatch, so what arrives
// is the bare object rather than upstream's one-element array.
//
// The pointer on IgnoredTypeNames is what separates an absent key from an explicit empty list, and
// the two mean opposite things here: absent keeps the four builtin names, explicit-empty clears them
// and asks the rule to report on an `Error`. Upstream expresses the same distinction through
// `option.ignoredTypeNames ?? []` over a defaults merge.
func DecodeNoBaseToStringOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noBaseToStringRawOptions]()(raw)
	if err != nil {
		return DefaultNoBaseToStringSettings(), err
	}

	wire, _ := decoded.(noBaseToStringRawOptions)
	options := DefaultNoBaseToStringSettings()
	if wire.CheckUnknown != nil {
		options.CheckUnknown = *wire.CheckUnknown
	}
	if wire.IgnoredTypeNames != nil {
		options.IgnoredTypeNames = *wire.IgnoredTypeNames
	}
	return options, nil
}

// NoBaseToString flags a value being stringified that has no useful string form.
//
//	valid:   String(new Date());               Date declares its own toString
//	valid:   `${'literal'}`;
//	valid:   [1, 2].join(', ');
//	invalid: String({});                       renders as [object Object]
//	invalid: `${{}}`;
//	invalid: '' + {};
//	invalid: [{}, {}].join(', ');
//
// `[object Object]` in a log line or a user-facing string is one of the quietest defects available:
// nothing throws, the type system is satisfied, and the information is simply gone. The rule asks
// whether a value being coerced to a string has anything to say.
//
// # The judgment is a three-valued recursion over the TYPE, not over the syntax
//
// A value stringifies usefully Always, Never, or Sometimes, and the third is what a union produces
// when its constituents disagree. That value is interpolated into the message, so the certainty is
// something the reader sees: "will use Object's default" against "may use Object's default".
//
// The recursion walks unions, intersections, tuple elements, array element types, and a type
// parameter's constraint, with a visited set so a self-referential array terminates. Two combining
// rules, both upstream's and both asymmetric on purpose:
//
//	union          all Never is Never, all Always is Always, anything else is Sometimes
//	intersection   ANY Always is Always, otherwise Never
//
// The intersection rule is the one that reads oddly and is right: intersecting a useless shape with
// a useful one gives a value that has the useful `toString`, so the whole is fine.
//
// # The base question is where a coercion method is DECLARED
//
// Upstream asks for `toLocaleString`, `toString` and `valueOf` in that order, and for each looks at
// every declaration. A declaration whose parent is anything other than the `Object` interface means
// somebody wrote their own, so the type is useful. Only when all the ones that exist come from
// `Object` itself is the value useless. Measured against our checker: `{}` finds all three on
// `Object`, `string` finds `toString` on `String`, and a class with its own method finds it on a
// class declaration, which is exactly the discrimination upstream needs.
//
// An explicit `[Symbol.toPrimitive]` short-circuits that entirely, because such a declaration is by
// construction user-defined.
//
// # `unknown` is a third state and the option is what reads it
//
// The base question answers yes, no, or "no coercion method at all", and that last case covers both
// `any` and `unknown`. Upstream distinguishes them by flags, and only reports on `unknown` when
// `checkUnknown` is set. An unconstrained type parameter is treated as `unknown` for the same reason.
//
// # This rule reads the PROGRAM
//
// `IsSymbolFromDefaultLibrary` walks the program's own default-library files to decide whether the
// `Symbol` in `[Symbol.toPrimitive]` is the builtin one, which is a read outside the file being
// linted.
//
// # Cost
//
// Five anchors, and the recursion is bounded by the type's own depth with a visited set. Every anchor
// is a coercion site rather than an arbitrary expression, so the rule declines almost everything in a
// real file before asking a type question.
var NoBaseToString = rule.Rule{
	Name: "@typescript-eslint/no-base-to-string",

	// Every finding is decided by walking a type, so there is no syntactic subset of this rule.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[NoBaseToStringOptions](options)
		if !isSettings {
			settings = DefaultNoBaseToStringSettings()
		}
		state := &noBaseToStringState{ctx: ctx, settings: settings}

		return rule.Listeners{
			// `a + b` and `a += b`, which our parser files under one kind rather than upstream's two.
			ast.KindBinaryExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil {
					return
				}
				if binary.OperatorToken.Kind != ast.KindPlusToken &&
					binary.OperatorToken.Kind != ast.KindPlusEqualsToken {
					return
				}

				leftType := ctx.TypeChecker.GetTypeAtLocation(binary.Left)
				rightType := ctx.TypeChecker.GetTypeAtLocation(binary.Right)
				if leftType == nil || rightType == nil {
					return
				}

				// Only a concatenation WITH a string coerces the other side, and upstream checks the
				// left first, so `'' + {}` reports the right and `{} + ''` reports the left.
				if type_checking.GetTypeName(ctx.TypeChecker, leftType) == "string" {
					state.checkExpression(binary.Right, rightType)
					return
				}
				if binary.Left.Kind == ast.KindPrivateIdentifier {
					return
				}
				if type_checking.GetTypeName(ctx.TypeChecker, rightType) == "string" {
					state.checkExpression(binary.Left, leftType)
				}
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				call := node.AsCallExpression()

				// `String(x)`, but only when `String` is the global rather than a local shadow.
				if state.isBuiltinStringCall(call) {
					argument := call.Arguments.Nodes[0]
					if argument.Kind != ast.KindSpreadElement {
						state.checkExpression(argument, nil)
					}
					return
				}

				// `x.join()`, `x.toString()`, `x.toLocaleString()`. Upstream reaches these through
				// three selectors on the member expression; here the call is the anchor and the
				// callee is read off it.
				callee := call.Expression
				if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
					return
				}
				access := callee.AsPropertyAccessExpression()
				name := access.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}
				receiver := access.Expression
				if receiver == nil {
					return
				}

				switch name.AsIdentifier().Text {
				case "join":
					// The constrained type, because a type parameter's constraint is what says
					// whether the elements stringify.
					receiverType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, receiver)
					if receiverType == nil {
						return
					}
					state.checkExpressionForArrayJoin(receiver, receiverType)
				case "toString", "toLocaleString":
					state.checkExpression(receiver, nil)
				}
			},

			// A template literal coerces every interpolation, unless the whole thing is a tag's
			// argument, where the tag decides what to do with the raw parts.
			ast.KindTemplateExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
					return
				}
				for _, span := range node.AsTemplateExpression().TemplateSpans.Nodes {
					expression := span.AsTemplateSpan().Expression
					if expression != nil {
						state.checkExpression(expression, nil)
					}
				}
			},
		}
	},
}

// noBaseToStringState carries what every step of the recursion needs.
//
// A struct rather than a closure chain because the recursion is deep and threading four values
// through eight functions obscures which of them actually decide anything.
type noBaseToStringState struct {
	ctx      rule.Context
	settings NoBaseToStringOptions
}

// checkExpression is upstream's `checkExpression`.
//
// A literal is skipped outright: its string form is whatever the author wrote.
func (state *noBaseToStringState) checkExpression(node *ast.Node, knownType *checker.Type) {
	// The parenthesis unwrap, which upstream has no counterpart for because its parser folds the
	// node away before the rule runs. Two things here depend on it and they fail differently:
	//
	//	the literal skip below     `('x')` would not be recognized as a literal
	//	the message text           it interpolates getText(node), so `({}).toString()` renders the
	//	                           offending expression as `({})` where upstream renders `{}`
	//
	// The second is what upstream's own corpus catches: two of its cases write exactly that shape
	// and record the name without the parentheses. Written as a loop because `((x))` nests, and
	// deliberately not ast.SkipParentheses, which dereferences its argument.
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		inner := node.AsParenthesizedExpression().Expression
		if inner == nil {
			break
		}
		node = inner
	}
	if node == nil {
		return
	}

	// Upstream's literal skip, SUBSUMED here and kept anyway.
	//
	// A mutation removing it survives every fixture, so the rule was run with and without the branch
	// over a regular expression, a string, a number, a boolean and a template, and both versions
	// answered zero findings on all of them. Every literal type declares its own coercion method, so
	// the base question below already answers Always and this never decides anything.
	//
	// Kept because it is upstream's, because it declines in one kind test what otherwise costs a type
	// query and a walk, and because the premise is a library declaration rather than this rule's: the
	// same shape as the boolean carve-out above.
	if isNoBaseToStringLiteral(node) {
		return
	}

	subjectType := knownType
	if subjectType == nil {
		subjectType = state.ctx.TypeChecker.GetTypeAtLocation(node)
	}
	if subjectType == nil {
		return
	}

	certainty := state.collectToStringCertainty(subjectType, nil)
	if certainty == noBaseToStringAlways {
		return
	}

	nodeRange := rule.TokenRange(state.ctx.SourceFile, node)
	name := state.ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
	state.ctx.ReportNode(node, rule.Message{
		Id: "baseToString",
		Description: "'" + name + "' " + string(certainty) +
			" use Object's default stringification format ('[object Object]') when stringified.",
	})
}

// checkExpressionForArrayJoin is upstream's `checkExpressionForArrayJoin`.
func (state *noBaseToStringState) checkExpressionForArrayJoin(node *ast.Node, subjectType *checker.Type) {
	certainty := state.collectJoinCertainty(subjectType, nil)
	if certainty == noBaseToStringAlways {
		return
	}

	nodeRange := rule.TokenRange(state.ctx.SourceFile, node)
	name := state.ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
	state.ctx.ReportNode(node, rule.Message{
		Id: "baseArrayJoin",
		Description: "Using `join()` for " + name + " " + string(certainty) +
			" use Object's default stringification format ('[object Object]') when stringified.",
	})
}

// collectJoinCertainty is upstream's `collectJoinCertainty`.
//
// Joining is only interesting for something array-shaped. Anything else is Always, because `join` on
// a non-array is either a user-defined method or a type error the compiler already reports.
func (state *noBaseToStringState) collectJoinCertainty(
	subjectType *checker.Type,
	visited []*checker.Type,
) noBaseToStringUsefulness {
	if type_checking.IsUnionType(subjectType) {
		return combineNoBaseToStringUnion(type_checking.UnionTypeParts(subjectType),
			func(part *checker.Type) noBaseToStringUsefulness {
				return state.collectJoinCertainty(part, visited)
			})
	}
	if type_checking.IsIntersectionType(subjectType) {
		return combineNoBaseToStringIntersection(type_checking.IntersectionTypeParts(subjectType),
			func(part *checker.Type) noBaseToStringUsefulness {
				return state.collectJoinCertainty(part, visited)
			})
	}
	if checker.IsTupleType(subjectType) {
		return state.collectTupleCertainty(subjectType, visited)
	}
	if checker.Checker_isArrayType(state.ctx.TypeChecker, subjectType) {
		return state.collectArrayCertainty(subjectType, visited)
	}
	return noBaseToStringAlways
}

// collectTupleCertainty is upstream's `collectTupleCertainty`.
//
// The combining rule differs from a union's on purpose: every element of a tuple is stringified, so
// ANY useless element makes the join useless, where a union only takes one branch.
func (state *noBaseToStringState) collectTupleCertainty(
	subjectType *checker.Type,
	visited []*checker.Type,
) noBaseToStringUsefulness {
	sawSometimes := false
	for _, argument := range checker.Checker_getTypeArguments(state.ctx.TypeChecker, subjectType) {
		switch state.collectToStringCertainty(argument, visited) {
		case noBaseToStringNever:
			return noBaseToStringNever
		case noBaseToStringSometimes:
			sawSometimes = true
		}
	}
	if sawSometimes {
		return noBaseToStringSometimes
	}
	return noBaseToStringAlways
}

// collectArrayCertainty is upstream's `collectArrayCertainty`.
//
// Upstream asserts the number index type is present and would throw if it were not. Ours declines
// instead: an assertion that cannot fail in upstream's tests can still fail on a shape our checker
// produces, and taking the whole run down over it would cost every rule this file.
func (state *noBaseToStringState) collectArrayCertainty(
	subjectType *checker.Type,
	visited []*checker.Type,
) noBaseToStringUsefulness {
	elementType := type_checking.GetNumberIndexType(state.ctx.TypeChecker, subjectType)
	if elementType == nil {
		return noBaseToStringAlways
	}
	return state.collectToStringCertainty(elementType, visited)
}

// collectToStringCertainty is upstream's `collectToStringCertainty`, the heart of the rule.
func (state *noBaseToStringState) collectToStringCertainty(
	subjectType *checker.Type,
	visited []*checker.Type,
) noBaseToStringUsefulness {
	// The visited set exists for a self-referential array or tuple, where the recursion would
	// otherwise not terminate. A slice rather than a map: these are single-digit lengths and a
	// linear scan avoids allocating a map per recursion.
	for _, seen := range visited {
		if seen == subjectType {
			return noBaseToStringAlways
		}
	}

	if type_checking.IsTypeParameter(subjectType) {
		constraint := checker.Checker_getBaseConstraintOfType(state.ctx.TypeChecker, subjectType)
		if constraint != nil {
			return state.collectToStringCertainty(constraint, visited)
		}
		// An unconstrained parameter is `unknown`, so it follows the same option.
		if state.settings.CheckUnknown {
			return noBaseToStringSometimes
		}
		return noBaseToStringAlways
	}

	// Upstream's comment says the Boolean type definition is missing toString(), so without this the
	// base question would answer Never for a boolean and report on `${true}`.
	//
	// It is SUBSUMED here and kept anyway. A mutation removing it survives every fixture, which sent
	// this to a direct comparison rather than to an equivalence argument: the rule was run with and
	// without the branch over a boolean, a boolean literal, a boolean under String(), and an array of
	// booleans joined, and both versions answered zero findings on all four. Our lib declares
	// `toString` on the `Boolean` interface, so the Object-interface test below already answers
	// Always and this branch never decides anything.
	//
	// Kept because it is upstream's and because the premise belongs to a library declaration rather
	// than to this rule: a lib revision that dropped that declaration would turn this from redundant
	// into the only thing standing between the rule and a false positive on every interpolated
	// boolean.
	if type_checking.IsTypeFlagSet(subjectType,
		checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral) {
		return noBaseToStringAlways
	}

	if state.isIgnoredTypeName(subjectType) {
		return noBaseToStringAlways
	}

	if type_checking.IsIntersectionType(subjectType) {
		return combineNoBaseToStringIntersection(type_checking.IntersectionTypeParts(subjectType),
			func(part *checker.Type) noBaseToStringUsefulness {
				return state.collectToStringCertainty(part, visited)
			})
	}
	if type_checking.IsUnionType(subjectType) {
		return combineNoBaseToStringUnion(type_checking.UnionTypeParts(subjectType),
			func(part *checker.Type) noBaseToStringUsefulness {
				return state.collectToStringCertainty(part, visited)
			})
	}
	if checker.IsTupleType(subjectType) {
		return state.collectTupleCertainty(subjectType, append(visited, subjectType))
	}
	if checker.Checker_isArrayType(state.ctx.TypeChecker, subjectType) {
		return state.collectArrayCertainty(subjectType, append(visited, subjectType))
	}

	coercion, known := state.isToStringLikeFromObject(subjectType)
	if !known {
		// No coercion method at all, which covers `any` and `unknown`. Only `unknown` is reportable,
		// and only under the option.
		if state.settings.CheckUnknown &&
			checker.Type_flags(subjectType) == checker.TypeFlagsUnknown {
			return noBaseToStringSometimes
		}
		return noBaseToStringAlways
	}
	if coercion {
		return noBaseToStringNever
	}
	return noBaseToStringAlways
}

// isIgnoredTypeName is upstream's two ignoredTypeNames checks.
//
// The first covers a GENERIC ignored type, where the type as written carries type arguments and the
// name has to be read off the alias or the symbol rather than off the rendered form. The second is
// the ordinary case and also walks base types, so a class extending Error is ignored too.
func (state *noBaseToStringState) isIgnoredTypeName(subjectType *checker.Type) bool {
	symbol := noBaseToStringAliasOrSymbol(subjectType)
	if symbol != nil && len(symbol.Declarations) > 0 {
		declaration := symbol.Declarations[0]
		if noBaseToStringDeclarationCanHaveTypeParameters(declaration) &&
			noBaseToStringDeclarationHasTypeParameters(declaration) &&
			noBaseToStringNameIsIgnored(state.settings.IgnoredTypeNames, symbol.Name) {
			return true
		}
	}

	return state.matchesTypeOrBaseType(subjectType, nil)
}

// matchesTypeOrBaseType is upstream's `matchesTypeOrBaseType` with the ignored-name predicate inlined.
//
// It walks base types, so a class extending an ignored one is ignored as well. The visited list
// guards against a cycle in a heritage chain, which error recovery can produce.
func (state *noBaseToStringState) matchesTypeOrBaseType(
	subjectType *checker.Type,
	visited []*checker.Type,
) bool {
	for _, seen := range visited {
		if seen == subjectType {
			return false
		}
	}
	if noBaseToStringNameIsIgnored(state.settings.IgnoredTypeNames,
		type_checking.GetTypeName(state.ctx.TypeChecker, subjectType)) {
		return true
	}
	visited = append(visited, subjectType)
	for _, baseType := range state.baseTypesOf(subjectType) {
		if state.matchesTypeOrBaseType(baseType, visited) {
			return true
		}
	}
	return false
}

// baseTypesOf reaches a class's bases, including through a generic INSTANTIATION.
//
// `Checker_getBaseTypes` answers nothing for an instantiated generic: measured, `Boom` reports one
// base and walks to `Error`, while `Boom<number>` and `Boom<T>` report zero and the walk stops
// immediately. Upstream never sees this because its `getBaseTypes` resolves through the reference
// itself, so the whole heritage chain is available to it for free.
//
// The recovery is the one a sibling rule already uses: an instantiated generic is a non-deferred type
// reference whose `Target()` is the declared type, and the bases live on the target. Without it a
// class extending `Error` through a generic subclass is not recognized as ignored, which is one of
// upstream's own passing cases and would be a false positive on every generic error class in a tree.
func (state *noBaseToStringState) baseTypesOf(subjectType *checker.Type) []*checker.Type {
	target := subjectType
	if checker.IsNonDeferredTypeReference(subjectType) {
		if referenced := subjectType.Target(); referenced != nil {
			target = referenced
		}
	}
	if checker.Type_objectFlags(target)&
		(checker.ObjectFlagsInterface|checker.ObjectFlagsClass) == 0 {
		return nil
	}
	return checker.Checker_getBaseTypes(state.ctx.TypeChecker, target)
}

// isToStringLikeFromObject is upstream's `isToStringLikeFromObject`.
//
// The three returns are yes, no, and "no coercion method found at all", which the caller reads as
// `any` or `unknown`. The second return says whether the first is meaningful.
func (state *noBaseToStringState) isToStringLikeFromObject(subjectType *checker.Type) (bool, bool) {
	// An explicit [Symbol.toPrimitive] is by construction user-defined, so nothing else matters.
	for _, property := range checker.Checker_getPropertiesOfType(state.ctx.TypeChecker, subjectType) {
		if property.ValueDeclaration != nil &&
			state.isSymbolToPrimitiveMethod(property.ValueDeclaration) {
			return false, true
		}
	}

	foundFallbackOnObject := false
	for _, propertyName := range []string{"toLocaleString", "toString", "valueOf"} {
		candidate := checker.Checker_getPropertyOfType(state.ctx.TypeChecker, subjectType, propertyName)
		if candidate == nil || len(candidate.Declarations) == 0 {
			continue
		}

		// A declaration from anywhere other than the `Object` interface means somebody wrote their
		// own coercion, so the type is useful.
		for _, declaration := range candidate.Declarations {
			if !noBaseToStringDeclaredOnObjectInterface(declaration) {
				return false, true
			}
		}
		foundFallbackOnObject = true
	}

	if foundFallbackOnObject {
		return true, true
	}
	return false, false
}

// isSymbolToPrimitiveMethod is upstream's `isSymbolToPrimitiveMethod`.
//
// It matches a method signature whose name is the computed `[Symbol.toPrimitive]`, and requires that
// `Symbol` resolve to the builtin rather than to a local of that name.
func (state *noBaseToStringState) isSymbolToPrimitiveMethod(declaration *ast.Node) bool {
	if declaration.Kind != ast.KindMethodSignature {
		return false
	}
	name := declaration.AsMethodSignatureDeclaration().Name()
	if name == nil || name.Kind != ast.KindComputedPropertyName {
		return false
	}
	computed := name.AsComputedPropertyName().Expression
	if computed == nil || computed.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := computed.AsPropertyAccessExpression()
	if access.Expression == nil || access.Expression.Kind != ast.KindIdentifier ||
		access.Expression.AsIdentifier().Text != "Symbol" {
		return false
	}
	if access.Name() == nil || access.Name().Kind != ast.KindIdentifier ||
		access.Name().AsIdentifier().Text != "toPrimitive" {
		return false
	}
	symbol := state.ctx.TypeChecker.GetSymbolAtLocation(access.Expression)
	return symbol != nil && type_checking.IsSymbolFromDefaultLibrary(state.ctx.Program, symbol)
}

// isBuiltinStringCall is upstream's `isBuiltInStringCall`.
//
// Upstream asks scope analysis whether any variable named `String` is declared, and treats "none" as
// the global. We have no scope index, so the symbol is resolved and its declarations examined: a
// `String` declared in this file is a shadow, and one declared only in the default library is the
// builtin. That is the complement of the naive predicate, which the brief warns answers false for
// exactly the globals a rule like this must recognize.
func (state *noBaseToStringState) isBuiltinStringCall(call *ast.CallExpression) bool {
	callee := call.Expression
	if callee == nil || callee.Kind != ast.KindIdentifier ||
		callee.AsIdentifier().Text != "String" {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}

	symbol := state.ctx.TypeChecker.GetSymbolAtLocation(callee)
	if symbol == nil {
		// Unresolvable, so nothing says it is a shadow.
		return true
	}
	for _, declaration := range symbol.Declarations {
		if !type_checking.IsSourceFileDefaultLibrary(state.ctx.Program,
			ast.GetSourceFileOfNode(declaration)) {
			return false
		}
	}
	return true
}

// combineNoBaseToStringUnion is upstream's `collectUnionTypeCertainty`.
func combineNoBaseToStringUnion(
	parts []*checker.Type,
	certaintyOf func(*checker.Type) noBaseToStringUsefulness,
) noBaseToStringUsefulness {
	allNever := true
	allAlways := true
	for _, part := range parts {
		switch certaintyOf(part) {
		case noBaseToStringNever:
			allAlways = false
		case noBaseToStringAlways:
			allNever = false
		default:
			allNever = false
			allAlways = false
		}
	}
	if allNever {
		return noBaseToStringNever
	}
	if allAlways {
		return noBaseToStringAlways
	}
	return noBaseToStringSometimes
}

// combineNoBaseToStringIntersection is upstream's `collectIntersectionTypeCertainty`.
//
// ANY useful constituent makes the whole useful, because the intersection has that constituent's
// coercion method. Everything else is Never, including Sometimes, which is upstream's behavior and
// reads as a rounding-down rather than an oversight.
func combineNoBaseToStringIntersection(
	parts []*checker.Type,
	certaintyOf func(*checker.Type) noBaseToStringUsefulness,
) noBaseToStringUsefulness {
	for _, part := range parts {
		if certaintyOf(part) == noBaseToStringAlways {
			return noBaseToStringAlways
		}
	}
	return noBaseToStringNever
}

// noBaseToStringAliasOrSymbol is upstream's `type.aliasSymbol ?? type.getSymbol()`.
func noBaseToStringAliasOrSymbol(subjectType *checker.Type) *ast.Symbol {
	if alias := checker.Type_alias(subjectType); alias != nil {
		if symbol := alias.Symbol(); symbol != nil {
			return symbol
		}
	}
	return checker.Type_symbol(subjectType)
}

// noBaseToStringDeclarationCanHaveTypeParameters is upstream's `canHaveTypeParameters`.
func noBaseToStringDeclarationCanHaveTypeParameters(declaration *ast.Node) bool {
	switch declaration.Kind {
	case ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration, ast.KindClassDeclaration:
		return true
	}
	return false
}

// noBaseToStringDeclarationHasTypeParameters is upstream's `decl.typeParameters` truthiness.
func noBaseToStringDeclarationHasTypeParameters(declaration *ast.Node) bool {
	parameters := declaration.TypeParameterList()
	return parameters != nil && len(parameters.Nodes) > 0
}

// noBaseToStringDeclaredOnObjectInterface answers upstream's
// `isInterfaceDeclaration(parent) && parent.name.text === 'Object'`.
func noBaseToStringDeclaredOnObjectInterface(declaration *ast.Node) bool {
	parent := declaration.Parent
	if parent == nil || parent.Kind != ast.KindInterfaceDeclaration {
		return false
	}
	name := parent.AsInterfaceDeclaration().Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "Object"
}

// noBaseToStringNameIsIgnored is upstream's `ignoredTypeNames.includes(name)`.
func noBaseToStringNameIsIgnored(ignoredTypeNames []string, name string) bool {
	for _, ignored := range ignoredTypeNames {
		if ignored == name {
			return true
		}
	}
	return false
}

// isNoBaseToStringLiteral is upstream's `node.type === Literal` skip.
//
// Upstream's ESTree has one Literal node; ours splits every kind of literal onto its own, so the one
// test becomes this list. A template with no interpolation is a literal there and here.
func isNoBaseToStringLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNullKeyword,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword:
		return true
	}
	return false
}
