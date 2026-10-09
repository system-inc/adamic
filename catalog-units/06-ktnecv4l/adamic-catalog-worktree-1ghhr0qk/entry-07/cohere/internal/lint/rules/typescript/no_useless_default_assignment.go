package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUselessDefaultAssignmentOptions is the rule's option surface.
type NoUselessDefaultAssignmentOptions struct {
	// AllowRuleToRunWithoutStrictNullChecks suppresses the file-level complaint that the rule needs
	// `strictNullChecks` to reason correctly.
	//
	// Upstream spells this `allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing`, and the shouted
	// tail is deliberate on their part: without the option the rule's whole judgment rests on a
	// type system that is not tracking undefined, so every verdict is unreliable rather than merely
	// less precise. The wire name below keeps upstream's spelling exactly, because a config written
	// for one tool has to mean the same thing in the other.
	AllowRuleToRunWithoutStrictNullChecks bool
}

// DefaultNoUselessDefaultAssignmentSettings is upstream's `defaultOptions`.
func DefaultNoUselessDefaultAssignmentSettings() NoUselessDefaultAssignmentOptions {
	return NoUselessDefaultAssignmentOptions{AllowRuleToRunWithoutStrictNullChecks: false}
}

// noUselessDefaultAssignmentRawOptions is the wire shape.
//
// The pointer is not needed for correctness here, since the single key defaults to FALSE and the
// generic decoder's zero value would happen to be right. It is written this way anyway, for the
// same reason its sibling in this package is: relying on that coincidence leaves the next person to
// add a default-true key inheriting a decoder that silently inverts their rule.
type noUselessDefaultAssignmentRawOptions struct {
	AllowRuleToRunWithoutStrictNullChecks *bool `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"`
}

// DecodeNoUselessDefaultAssignmentOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple before dispatch, so what
// arrives is the bare object rather than upstream's one-element array.
func DecodeNoUselessDefaultAssignmentOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noUselessDefaultAssignmentRawOptions]()(raw)
	if err != nil {
		return DefaultNoUselessDefaultAssignmentSettings(), err
	}

	wire, _ := decoded.(noUselessDefaultAssignmentRawOptions)
	options := DefaultNoUselessDefaultAssignmentSettings()
	if wire.AllowRuleToRunWithoutStrictNullChecks != nil {
		options.AllowRuleToRunWithoutStrictNullChecks = *wire.AllowRuleToRunWithoutStrictNullChecks
	}
	return options, nil
}

// NoUselessDefaultAssignment flags a default value that can never be used.
//
//	valid:   function bar({ foo = '' }: { foo?: string }) { return foo; }
//	valid:   const { foo } = { foo: 'bar' };
//	valid:   function f(a: string | undefined = 'x') {}
//	invalid: function bar({ foo = '' }: { foo: string }) { return foo; }
//	invalid: function foo(a = undefined) {}
//	invalid: function f(p?: number | undefined = undefined) {}
//
// A default fires only when the incoming value is `undefined`. If the type says it never can be,
// the default is dead code that reads as a guarantee, and the next person to touch the function
// believes a fallback exists where none runs.
//
// # Three judgments, and they are genuinely different
//
//	uselessDefaultAssignment  the target is not optional, so the default cannot fire
//	uselessUndefined          the default IS `undefined`, which is what the value already is
//	preferOptionalSyntax      `= undefined` on a parameter whose annotation admits undefined;
//	                          the `?` marker says the same thing with no runtime branch
//
// Only the third rewrites anything beyond a deletion: it removes the default AND inserts `?` after
// the parameter name, so the declaration keeps its meaning. The other two only delete.
//
// # Where estree and our parser disagree, and why this rule has two anchors instead of one
//
// Upstream listens to a single `AssignmentPattern`. There is no such node here. TypeScript puts a
// default on the declaration that carries it, so the same source shape arrives as an `Initializer`
// on either a `Parameter` or a `BindingElement`. Probed rather than assumed:
//
//	function bar({ foo = '' }) {}   the default is on the BINDING ELEMENT, not the parameter
//	function baz(a = undefined) {}  the default is on the PARAMETER
//	const [x = 1] = ...             BINDING ELEMENT, parent is an array binding pattern
//	const { p = 'd' } = o           BINDING ELEMENT, parent is an object binding pattern
//
// So a port with one listener misses roughly half the corpus, and which half depends on which node
// it picked. The two listeners below are one rule, not two, and they share `checkDefault`.
//
// # The strictNullChecks gate, reproduced as a decision rather than as a finding
//
// Upstream reports `noStrictNullCheck` once per file, at position zero, when the option is off and
// the escape hatch is not set. Without `strictNullChecks` the checker does not track `undefined` in
// a type at all, so `canBeUndefined` answers false for everything and the rule would report every
// default in the file as useless.
//
// This port reproduces the gate by DECLINING to judge rather than by reporting at position zero,
// and that is a stated divergence rather than an oversight. A file-level finding anchored at 0:0
// has no node to point at, and the surrounding harness reports per node; more importantly the
// finding says nothing about the code, it says something about the configuration. The decline is
// made once per file in Run and recorded through Skip, so --coverage names the rule and the reason
// rather than letting a declined file read as a clean one (#pa7k7zv). Every project here resolves
// `strictNullChecks` on, ahra and www by TypeScript 6's default and api by writing `strict`, so the
// skip fires only on a project that turns it off (#6ar414z). Measured on the installed 8.67.0 build: under an
// unstrict project, upstream emits the `noStrictNullCheck` finding AND the real one; under the
// option, only the real one. The three corpus cases written against an unstrict project keep their
// real findings here, which is the half that is about the code.
//
// # Cost
//
// Both anchors are common, and the body exits immediately when there is no initializer, which is
// the overwhelming majority of parameters and binding elements. The checker is consulted only for a
// declaration that actually carries a default.
var NoUselessDefaultAssignment = rule.Rule{
	Name: "@typescript-eslint/no-useless-default-assignment",

	// Every judgment is a question about whether a type admits undefined.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// The compiler options.
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[NoUselessDefaultAssignmentOptions](options)
		if !isSettings {
			settings = DefaultNoUselessDefaultAssignmentSettings()
		}

		// canBeUndefined is upstream's predicate: `any` and `unknown` both admit undefined, and
		// otherwise some union constituent must carry the flag. A non-union is its own constituent.
		canBeUndefined := func(subject *checker.Type) bool {
			if subject == nil {
				return false
			}
			if type_checking.IsTypeAnyType(subject) || type_checking.IsTypeUnknownType(subject) {
				return true
			}
			for part := range type_checking.UnionTypePartsSeq(subject) {
				if type_checking.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					return true
				}
			}
			return false
		}

		// Without strictNullChecks the checker does not track undefined at all, so canBeUndefined
		// answers false for every type and the rule would report every default in the file. Upstream
		// reports a file-level finding and keeps going; this port declines the file instead, and says
		// so through Skip, since a declined file and a clean one otherwise read the same (#pa7k7zv).
		// See the rule's doc comment for the measurement and the reasoning.
		if ctx.Program != nil &&
			!type_checking.IsStrictCompilerOptionEnabled(ctx.Program.Options(), ctx.Program.Options().StrictNullChecks) &&
			!settings.AllowRuleToRunWithoutStrictNullChecks {
			ctx.Skip("strictNullChecks is off")
			return nil
		}

		return rule.Listeners{
			ast.KindParameter: func(node *ast.Node) {
				if ctx.TypeChecker == nil || ctx.Program == nil {
					return
				}
				parameter := node.AsParameterDeclaration()
				if parameter.Initializer == nil {
					return
				}
				checkUselessDefault(ctx, canBeUndefined, node, parameter.Initializer,
					parameter.Name(), parameter.Type, "parameter")
			},
			ast.KindBindingElement: func(node *ast.Node) {
				if ctx.TypeChecker == nil || ctx.Program == nil {
					return
				}
				element := node.AsBindingElement()
				if element.Initializer == nil {
					return
				}
				// A binding element never carries its own type annotation; the annotation lives on
				// whatever the pattern destructures. Passing nil here is what routes it to the
				// property and tuple paths rather than the parameter one.
				checkUselessDefault(ctx, canBeUndefined, node, element.Initializer,
					element.Name(), nil, "property")
			},
		}
	},
}

// checkUselessDefault is upstream's `checkAssignmentPattern`, shared by both anchors.
//
// `declarationType` is the annotation when the anchor is a parameter and nil for a binding element,
// which is the only thing that distinguishes the two paths through the `undefined` branch.
func checkUselessDefault(
	ctx rule.Context,
	canBeUndefined func(*checker.Type) bool,
	node *ast.Node,
	initializer *ast.Node,
	name *ast.Node,
	declarationType *ast.Node,
	kindText string,
) {
	// `= undefined` is decided without asking what the target type is, because assigning the value
	// a thing already has cannot help whatever the type says.
	if isUndefinedIdentifier(initializer) {
		// On a parameter with an annotation that admits undefined, upstream prefers the `?` marker
		// and offers a two-part repair. The annotation test is what separates this from the plain
		// uselessUndefined case: `function f(p?: number | undefined = undefined)` gets the marker,
		// `function foo(a = undefined)` has no annotation and simply loses the default.
		if declarationType != nil &&
			canBeUndefined(checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, declarationType)) {
			reportPreferOptionalSyntax(ctx, node, initializer, name, declarationType)
			return
		}
		ctx.ReportNodeWithFixes(initializer, uselessUndefinedMessage(kindText),
			removeDefaultFix(ctx, initializer, name, declarationType))
		return
	}

	targetType, resolved := defaultTargetType(ctx, node)
	if !resolved || targetType == nil {
		return
	}
	if canBeUndefined(targetType) {
		return
	}
	ctx.ReportNodeWithFixes(initializer, uselessDefaultAssignmentMessage(kindText),
		removeDefaultFix(ctx, initializer, name, declarationType))
}

// defaultTargetType answers what type the defaulted target actually has.
//
// The second return distinguishes "the type is this" from "this port could not work it out", so a
// nil type never reads as a resolved absence. Upstream returns null for both and relies on the
// caller checking; keeping the two apart makes the declines visible at the call site.
func defaultTargetType(ctx rule.Context, node *ast.Node) (*checker.Type, bool) {
	if ast.IsParameterDeclaration(node) {
		return contextualParameterType(ctx, node)
	}

	if !ast.IsBindingElement(node) {
		return nil, false
	}

	pattern := node.Parent
	if pattern == nil {
		return nil, false
	}

	sourceType, resolved := bindingPatternSourceType(ctx, pattern)
	if !resolved || sourceType == nil {
		return nil, false
	}

	if ast.IsObjectBindingPattern(pattern) {
		propertyName := bindingElementPropertyName(node)
		if propertyName == "" {
			return nil, false
		}
		symbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, sourceType, propertyName)
		if symbol == nil {
			return nil, false
		}
		// An OPTIONAL property destructured out of a conditionally-built object needs one more
		// question asked, and skipping it produced three false positives on upstream's own passing
		// cases. `const { a = 'default' } = cond ? { a: 'Hello' } : {}` has `a` optional on the
		// union, but the checker's type for the property is just `string`, because the branch that
		// omits it contributes nothing to read. So the default is live and the naive answer says
		// it is dead.
		//
		// Upstream's test is whether EVERY branch of the initializer supplies the property. If one
		// does not, the property can genuinely arrive undefined and the default is reachable.
		if type_checking.IsSymbolFlagSet(symbol, ast.SymbolFlagsOptional) {
			if initializer := conditionalInitializerAbove(pattern); initializer != nil &&
				!hasPropertyInAllBranches(initializer, propertyName) {
				return nil, false
			}
		}
		return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, symbol), true
	}

	if ast.IsArrayBindingPattern(pattern) {
		// TUPLE ONLY, and the narrowness is upstream's rather than a simplification.
		//
		// A tuple states each position's type and its length, so element 0 of `[string]` is a
		// string and element 1 provably does not exist. A plain `Array<string>` states neither:
		// every index has the same element type and the checker cannot say whether the index is in
		// range, so a default on `const [foo = ''] = g` fires whenever the array is short.
		//
		// This was written with a number-index fallback first, on the reasoning that upstream has
		// one. Upstream does have one, in `getArrayElementType`, and it is reached from exactly one
		// call site: resolving the SOURCE of a nested pattern, never the target being reported on.
		// The fallback here produced five false positives on upstream's own passing cases, all of
		// them a top-level array pattern over a plain array. Read rather than reasoned about after
		// the corpus disagreed: the fallback belongs in `arrayElementType` below.
		if !checker.IsTupleType(sourceType) {
			return nil, false
		}
		elements := pattern.AsBindingPattern().Elements
		if elements == nil {
			return nil, false
		}
		index := -1
		for position, element := range elements.Nodes {
			if element == node {
				index = position
				break
			}
		}
		typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, sourceType)
		if index < 0 || index >= len(typeArguments) {
			// Past the tuple's end, so this element provably does not exist and a default for it
			// is live. Corpus: `declare const tuple: [string]; const [a, b = 'default'] = tuple;`
			return nil, false
		}
		return typeArguments[index], true
	}

	return nil, false
}

// contextualParameterType answers a parameter's type from the signature the function is checked
// against, rather than from the parameter itself.
//
// This is the subtle half of upstream's rule and the reason it is not simply "read the annotation".
// A callback written inline gets its parameter types from the signature it is passed to, so
// `[1, 2, 3].map((a = 42) => a + 1)` has a default on a parameter the caller ALWAYS supplies, and
// nothing about the arrow itself says so. Upstream asks for the contextual type, takes its call
// signatures, and reports only when NO signature could ever leave the parameter undefined.
//
// A signature leaves it usable when the parameter is past the end of that signature's list, is a
// rest parameter, is marked optional, is a type parameter, or has a type admitting undefined. Any
// one of those on any overload is enough, which is why the loop below is an "any" rather than an
// "all".
func contextualParameterType(ctx rule.Context, node *ast.Node) (*checker.Type, bool) {
	function := node.Parent
	if function == nil {
		return nil, false
	}

	// Upstream restricts this path to a function EXPRESSION or an arrow, because only those are
	// checked against a contextual type. Anything else DECLINES, and that decline is load-bearing
	// rather than a narrowing this port could tidy up.
	//
	// The first version of this returned the parameter's declared type here instead, which reads
	// like an improvement and produced four false positives on upstream's own passing cases:
	// `function test(a: string = 'default')` reports, because `string` cannot be undefined. It is
	// wrong because a parameter carrying a default is IMPLICITLY OPTIONAL in TypeScript, so callers
	// may omit it and the default fires every time they do. The annotation describes what the
	// parameter is once bound, not what the caller must pass, and reading it as the latter inverts
	// the rule for the single most ordinary shape it will ever see.
	if !ast.IsArrowFunction(function) && !ast.IsFunctionExpression(function) {
		return nil, false
	}

	contextualType := checker.Checker_getContextualType(ctx.TypeChecker, function, checker.ContextFlagsNone)
	if contextualType == nil {
		return nil, false
	}

	signatures := checker.Checker_getSignaturesOfType(ctx.TypeChecker, contextualType, checker.SignatureKindCall)
	if len(signatures) == 0 {
		return nil, false
	}
	// When the contextual type's own signature IS this function, there is no caller contract to
	// compare against and the parameter's declared type is the whole story.
	if checker.Signature_declaration(signatures[0]) == function {
		return nil, false
	}

	parameterIndex := parameterIndexIn(function, node)
	if parameterIndex < 0 {
		return nil, false
	}

	for _, signature := range signatures {
		parameters := checker.Signature_parameters(signature)
		if parameterIndex >= len(parameters) {
			// This overload does not supply the parameter at all, so the default can fire.
			return nil, false
		}
		parameterSymbol := parameters[parameterIndex]
		if parameterSymbol == nil {
			return nil, false
		}
		if parameterSymbol.ValueDeclaration != nil &&
			type_checking.IsRestParameterDeclaration(parameterSymbol.ValueDeclaration) {
			return nil, false
		}
		if type_checking.IsSymbolFlagSet(parameterSymbol, ast.SymbolFlagsOptional) {
			return nil, false
		}
		parameterType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameterSymbol)
		if parameterType == nil || type_checking.IsTypeParameter(parameterType) {
			return nil, false
		}
		if type_checking.IsTypeAnyType(parameterType) || type_checking.IsTypeUnknownType(parameterType) {
			return nil, false
		}
		for part := range type_checking.UnionTypePartsSeq(parameterType) {
			if type_checking.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
				return nil, false
			}
		}
	}

	// No signature can leave the parameter undefined, so the default is dead. The type returned is
	// one the caller's canBeUndefined will reject; the first signature's parameter type is the
	// honest answer to "what does this parameter actually receive".
	parameters := checker.Signature_parameters(signatures[0])
	return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameters[parameterIndex]), true
}

// declaredParameterType reads a parameter's type from the signature its own function declares.
func declaredParameterType(ctx rule.Context, node *ast.Node) *checker.Type {
	parameter := node.AsParameterDeclaration()
	if parameter.Type != nil {
		return checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, parameter.Type)
	}
	return nil
}

// bindingPatternSourceType answers the type of whatever a destructuring pattern is taking apart.
//
// Four sources, and they recurse: an initializer on a variable declaration, a parameter's own
// declared type, a nested object pattern's property, or a nested array pattern's element.
func bindingPatternSourceType(ctx rule.Context, pattern *ast.Node) (*checker.Type, bool) {
	parent := pattern.Parent
	if parent == nil {
		return nil, false
	}

	if ast.IsVariableDeclaration(parent) {
		declaration := parent.AsVariableDeclaration()
		if declaration.Initializer == nil {
			return nil, false
		}
		return ctx.TypeChecker.GetTypeAtLocation(declaration.Initializer), true
	}

	if ast.IsParameterDeclaration(parent) {
		signature := ctx.TypeChecker.GetSignatureFromDeclaration(parent.Parent)
		if signature == nil {
			return nil, false
		}
		parameters := checker.Signature_parameters(signature)
		index := parameterIndexIn(parent.Parent, parent)
		if index < 0 || index >= len(parameters) {
			return nil, false
		}
		return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameters[index]), true
	}

	// A pattern nested inside another binding element: its source is that element's own type,
	// which the same machinery already knows how to answer.
	//
	// The ARRAY case cannot go through `defaultTargetType`, because that path is tuple-only by
	// design and the element being resolved here is a source rather than a report target. Upstream
	// splits the same way, and it is the one place its number-index fallback is reached.
	if ast.IsBindingElement(parent) {
		grandparent := parent.Parent
		if grandparent != nil && ast.IsArrayBindingPattern(grandparent) {
			outerType, resolved := bindingPatternSourceType(ctx, grandparent)
			if !resolved || outerType == nil {
				return nil, false
			}
			index := -1
			if elements := grandparent.AsBindingPattern().Elements; elements != nil {
				for position, element := range elements.Nodes {
					if element == parent {
						index = position
						break
					}
				}
			}
			if index < 0 {
				return nil, false
			}
			return arrayElementType(ctx, outerType, index)
		}
		return defaultTargetType(ctx, parent)
	}

	return nil, false
}

// conditionalInitializerAbove finds the conditional expression a destructuring pattern reads from.
//
// Upstream walks UP from the pattern rather than looking only at its immediate parent, because a
// nested pattern is several levels below the declaration that carries the initializer. The walk
// stops at the first variable declaration it meets and answers its initializer only when that
// initializer branches; a plain object literal cannot hide a missing property from the checker.
func conditionalInitializerAbove(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			return nil
		}
		if !ast.IsVariableDeclaration(parent) {
			continue
		}
		initializer := parent.AsVariableDeclaration().Initializer
		if initializer == nil {
			return nil
		}
		if initializer.Kind == ast.KindConditionalExpression || ast.IsBinaryExpression(initializer) {
			return initializer
		}
		return nil
	}
	return nil
}

// hasPropertyInAllBranches answers whether every branch of an initializer supplies a property.
//
// An object literal answers directly. A conditional recurses into BOTH arms and requires both, so
// `cond ? { a: 1 } : {}` is false and the default it guards is reachable. Anything else answers
// false, which is the safe direction: an initializer this cannot read is one whose branches cannot
// be enumerated, and reporting would claim a guarantee nobody established.
func hasPropertyInAllBranches(expression *ast.Node, propertyName string) bool {
	if expression == nil {
		return false
	}

	if ast.IsObjectLiteralExpression(expression) {
		properties := expression.AsObjectLiteralExpression().Properties
		if properties == nil {
			return false
		}
		for _, property := range properties.Nodes {
			name := property.Name()
			if name == nil {
				continue
			}
			// A method and a shorthand both count: upstream reads the KEY, so `{ a() {} }` supplies
			// `a` exactly as `{ a: 1 }` does. The corpus pins that with a case whose branch holds a
			// method, and it reports.
			if staticPropertyKeyText(name) == propertyName {
				return true
			}
		}
		return false
	}

	if expression.Kind == ast.KindConditionalExpression {
		conditional := expression.AsConditionalExpression()
		return hasPropertyInAllBranches(conditional.WhenTrue, propertyName) &&
			hasPropertyInAllBranches(conditional.WhenFalse, propertyName)
	}

	return false
}

// arrayElementType answers the type at one index of an array or tuple.
//
// This is upstream's `getArrayElementType`, and it exists to resolve the SOURCE a nested pattern
// destructures rather than to judge a default. That distinction is the whole reason the number-index
// fallback lives here and not in the reporting path: asking "what type sits at this index" is
// answerable for a plain array, while asking "can this index be missing" is not.
//
// A tuple answers from its type arguments when the index is in range. Everything else, including an
// index past a tuple's end, falls back to the number index type.
func arrayElementType(ctx rule.Context, arrayType *checker.Type, index int) (*checker.Type, bool) {
	if checker.IsTupleType(arrayType) {
		typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, arrayType)
		if index < len(typeArguments) {
			return typeArguments[index], true
		}
	}
	elementType := checker.Checker_getIndexTypeOfType(ctx.TypeChecker, arrayType,
		checker.Checker_numberType(ctx.TypeChecker))
	if elementType == nil {
		return nil, false
	}
	return elementType, true
}

// bindingElementPropertyName answers which property of the source a binding element reads.
//
// `{ a }` reads `a` from its own name; `{ 'literal-key': x }` and `{ [`a`]: x }` carry a
// property name node instead. A computed key that is not a plain literal is declined rather than
// guessed at, which is upstream's behaviour too.
func bindingElementPropertyName(node *ast.Node) string {
	element := node.AsBindingElement()
	if element.PropertyName != nil {
		return staticPropertyKeyText(element.PropertyName)
	}
	name := element.Name()
	if name != nil && ast.IsIdentifier(name) {
		return name.Text()
	}
	return ""
}

// staticPropertyKeyText reads a property key that is knowable without evaluating anything.
//
// Upstream accepts an identifier, a literal (stringified), and a template with no substitutions.
// The template case is not decoration: the corpus writes one, and it reports.
func staticPropertyKeyText(key *ast.Node) string {
	if ast.IsComputedPropertyName(key) {
		inner := key.AsComputedPropertyName().Expression
		if inner == nil {
			return ""
		}
		key = inner
	}
	switch key.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return key.Text()
	}
	return ""
}

// parameterIndexIn answers a parameter's position in its function's list.
func parameterIndexIn(function *ast.Node, parameter *ast.Node) int {
	parameters := function.Parameters()
	if parameters == nil {
		return -1
	}
	for position, candidate := range parameters {
		if candidate == parameter {
			return position
		}
	}
	return -1
}

// isUndefinedIdentifier answers whether an initializer is literally the identifier `undefined`.
//
// Upstream tests the name rather than the type, so a variable shadowing `undefined` would be
// treated as the keyword. That is upstream's behaviour and it is reproduced; the alternative would
// report differently from the tool this is diffed against.
func isUndefinedIdentifier(initializer *ast.Node) bool {
	return ast.IsIdentifier(initializer) && initializer.Text() == "undefined"
}

// removeDefaultFix deletes the default and the `=` introducing it, and NOTHING ELSE.
//
// The end is the initializer's end. The start is the end of whatever the default follows, which is
// the TYPE ANNOTATION when there is one and the bound name otherwise. Getting that wrong is how a
// fixer silently destroys type information, and this one did on its first version:
//
//	p2: number | undefined = undefined   from the NAME's end  -> `p2?`           annotation gone
//	                                     from the TYPE's end  -> `p2?: number | undefined`
//
// The first spelling compiles, so nothing fails; the declaration simply becomes implicitly `any`
// and the next reader has no way to see that a type was deleted. Upstream computes the start from
// `node.left.range[1]`, and in estree `left` is the whole annotated pattern INCLUDING its type
// annotation, so upstream's one expression is two different positions in our tree. Two of the
// corpus's fix vectors catch it and no message-id assertion can.
//
// Trailing trivia between the annotation and the `=` belongs to the annotation, so a comment
// sitting there survives. The corpus pins that with a parameter carrying five comments, whose
// repair keeps four of them and takes only the one attached to the deleted default.
func removeDefaultFix(ctx rule.Context, initializer *ast.Node, name *ast.Node, annotation *ast.Node) rule.Fix {
	start := name.End()
	if annotation != nil {
		start = annotation.End()
	}
	return rule.RemoveRange(core.NewTextRange(start, initializer.End()))
}

// reportPreferOptionalSyntax emits the one finding whose repair is more than a deletion.
//
// Two edits: the default goes, and a `?` lands immediately after the parameter's name. The insert
// is anchored on the name's end rather than before the annotation, because a comment can sit
// between them. The corpus pins that with `/* comment */ x?: SomeType`, whose repair keeps the
// comment where it was.
func reportPreferOptionalSyntax(ctx rule.Context, node *ast.Node, initializer *ast.Node, name *ast.Node,
	annotation *ast.Node) {
	fixes := []rule.Fix{removeDefaultFix(ctx, initializer, name, annotation)}
	if ast.IsIdentifier(name) {
		fixes = append(fixes, rule.ReplaceRange(core.NewTextRange(name.End(), name.End()), "?"))
	}
	ctx.ReportNodeWithFixes(initializer, preferOptionalSyntaxMessage(), fixes...)
}

func uselessDefaultAssignmentMessage(kindText string) rule.Message {
	return rule.Message{
		Id:          "uselessDefaultAssignment",
		Description: "Default value is useless because the " + kindText + " is not optional.",
	}
}

func uselessUndefinedMessage(kindText string) rule.Message {
	return rule.Message{
		Id: "uselessUndefined",
		Description: "Default value is useless because it is undefined. Optional " + kindText +
			"s are already undefined by default.",
	}
}

func preferOptionalSyntaxMessage() rule.Message {
	return rule.Message{
		Id: "preferOptionalSyntax",
		Description: "Using `= undefined` to make a parameter optional adds unnecessary runtime " +
			"logic. Use the `?` optional syntax instead.",
	}
}
