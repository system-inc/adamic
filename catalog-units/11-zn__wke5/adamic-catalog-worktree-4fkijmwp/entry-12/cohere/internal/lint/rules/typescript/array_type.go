package typescript

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimcore "github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The six message ids, each carrying upstream's three interpolations.
//
// `rule.Message` has no interpolation layer, so upstream's `{{className}}`, `{{type}}` and
// `{{readonlyPrefix}}` become concatenation at the report site. All three are worth carrying rather
// than dropping: one file can produce many findings under one id, and the quoted before-and-after is
// the only thing that separates them or tells a reader what the repair will write.
//
// The `type` slot is NOT the element type's source text in general. Upstream renders the text only
// when the element is a simple type and renders the single letter `T` otherwise, so
// `Array<string | number>` reads as `Array<T>`. That collapse is reproduced; see `arrayTypeMessageType`.

// messageArrayTypeArray is upstream's `errorStringArray`, for a generic that should be a suffix.
func messageArrayTypeArray(className string, elementType string, readonlyPrefix string) rule.Message {
	return rule.Message{
		Id: "errorStringArray",
		Description: "This array is written as `" + className + "<" + elementType + ">` while the " +
			"rest of the codebase writes `" + readonlyPrefix + elementType + "[]`, so a reader " +
			"scanning types has to recognize two spellings of one thing. Write `" +
			readonlyPrefix + elementType + "[]` instead.",
	}
}

// messageArrayTypeArrayReadonly is upstream's `errorStringArrayReadonly`.
//
// Distinct from the arm above because the suggested replacement already ends in `[]`: this arm is
// reached only for `Readonly<T[]>`, whose repair is `readonly T[]` rather than `readonly T[][]`.
func messageArrayTypeArrayReadonly(className string, elementType string, readonlyPrefix string) rule.Message {
	return rule.Message{
		Id: "errorStringArrayReadonly",
		Description: "This array is written as `" + className + "<" + elementType + ">` while the " +
			"rest of the codebase writes `" + readonlyPrefix + elementType + "`, so a reader " +
			"scanning types has to recognize two spellings of one thing. Write `" +
			readonlyPrefix + elementType + "` instead.",
	}
}

// messageArrayTypeArraySimple is upstream's `errorStringArraySimple`.
func messageArrayTypeArraySimple(className string, elementType string, readonlyPrefix string) rule.Message {
	return rule.Message{
		Id: "errorStringArraySimple",
		Description: "This array is written as `" + className + "<" + elementType + ">` over a " +
			"simple element type, where this project writes the suffix form. The generic spelling " +
			"is reserved for element types complicated enough that the brackets would be hard to " +
			"read. Write `" + readonlyPrefix + elementType + "[]` instead.",
	}
}

// messageArrayTypeArraySimpleReadonly is upstream's `errorStringArraySimpleReadonly`.
func messageArrayTypeArraySimpleReadonly(className string, elementType string, readonlyPrefix string) rule.Message {
	return rule.Message{
		Id: "errorStringArraySimpleReadonly",
		Description: "This array is written as `" + className + "<" + elementType + ">` over a " +
			"simple element type, where this project writes the suffix form. The generic spelling " +
			"is reserved for element types complicated enough that the brackets would be hard to " +
			"read. Write `" + readonlyPrefix + elementType + "` instead.",
	}
}

// messageArrayTypeGeneric is upstream's `errorStringGeneric`, for a suffix that should be generic.
func messageArrayTypeGeneric(className string, elementType string, readonlyPrefix string) rule.Message {
	return rule.Message{
		Id: "errorStringGeneric",
		Description: "This array is written as `" + readonlyPrefix + elementType + "[]` while the " +
			"rest of the codebase writes `" + className + "<" + elementType + ">`, so a reader " +
			"scanning types has to recognize two spellings of one thing. Write `" + className +
			"<" + elementType + ">` instead.",
	}
}

// messageArrayTypeGenericSimple is upstream's `errorStringGenericSimple`.
func messageArrayTypeGenericSimple(className string, elementType string, readonlyPrefix string) rule.Message {
	return rule.Message{
		Id: "errorStringGenericSimple",
		Description: "This array is written as `" + readonlyPrefix + elementType + "[]` over an " +
			"element type complicated enough that the trailing brackets are easy to miss. The " +
			"suffix form is reserved for simple element types here. Write `" + className + "<" +
			elementType + ">` instead.",
	}
}

// ArrayType requires one spelling of an array type, `T[]` or `Array<T>`, chosen by configuration.
//
//	valid (default array):        let x: string[];
//	valid (default array):        let x: readonly string[];
//	valid (default generic):      let x: Array<string>;
//	valid (array-simple):         let x: string[];  and  let x: Array<string | number>;
//	valid:                        type G<Array> = { a: Array };      a shadowed name is not the global
//	invalid (default array):      let x: Array<string>;              becomes  let x: string[];
//	invalid (default generic):    let x: string[];                   becomes  let x: Array<string>;
//	invalid (array-simple):       let x: (string | number)[];        becomes  let x: Array<string | number>;
//
// Ported from `@typescript-eslint/array-type`, reading the clone at
// `packages/eslint-plugin/src/rules/array-type.ts` and measuring every verdict, every message and
// every repair against the installed 8.67.0 build driven through the ESLint 10.8.1 Linter API. The
// corpus carries ninety-nine before-and-after pairs, which is the specification for the fixer, and
// all ninety-nine were confirmed to equal what the running rule writes.
//
// # Two listeners answering opposite questions
//
// `KindArrayType` judges the suffix form and reports when the configuration wants generics.
// `KindTypeReference` judges `Array<T>`, `ReadonlyArray<T>` and `Readonly<T[]>` and reports when the
// configuration wants suffixes. A file can hold both and produce both findings.
//
// # `readonly` is a separate configuration axis, defaulting to whatever `default` says
//
// `readonly T[]` and `ReadonlyArray<T>` are governed by the `readonly` option, and an absent
// `readonly` falls back to `default` rather than to a fixed value. That fallback is why the decoder
// is hand-rolled; see `DecodeArrayTypeOptions`.
//
// # The element type in the message is COLLAPSED to `T` unless it is simple
//
// `getMessageType` renders the element's source text for a simple type and the bare letter `T`
// otherwise, so `Array<string | number>` reports as `Array<T>`. Measured on the installed build over
// the six ids. Nothing in a message-id fixture can see this, which is why the rendered text is
// asserted on every reporting row.
//
// # `isSimpleType` is a recursive judgment and `Array` is special inside it
//
// A keyword, an identifier, a qualified name and a suffix array are simple. A type reference is
// simple when its name is `Array` and it either takes no arguments or takes one simple argument, and
// otherwise only when it takes NO arguments at all. So `Array<string>` is simple, `Foo<string>` is
// not, and `Foo` alone is. Reproduced exactly, including that `Array` gets a recursion no other name
// gets.
//
// # Our parser keeps parentheses and TSESTree does not, which changes the ELEMENT and the REPAIR
//
// TSESTree has no parenthesized-type node: `(string)[]`'s element type is the string keyword itself,
// with the parentheses surviving only as tokens between the ranges. typescript-go produces a real
// `KindParenthesizedType`. Reading the element without skipping it would make every parenthesized
// element non-simple, and would anchor the repair inside the parentheses. Measured on the installed
// build, `let x: (string)[]` under `generic` becomes `let x: Array<string>` with the parentheses
// gone, which is what skipping reproduces.
//
// The skip is therefore fidelity to a parser difference rather than a convenience, and it runs in
// only one direction: the element is skipped for the JUDGMENT and for the START of the fix range,
// while `isParenthesized` on the type-reference side asks the opposite question and is answered by
// looking for a real parenthesized node.
//
// # A shadowed `Array` is not the global one, and this needs lexical scope
//
// Upstream walks `context.sourceCode.getScope(node)` upward and declines when any enclosing scope
// binds the name. There is no scope shelf in this tree, so the walk is written here over the syntax
// tree; `arrayTypeNameIsShadowed` holds the measurement. It is load-bearing rather than defensive:
// all four shadowing forms below are clean upstream and all four controls report.
//
//	type Array<Y> = Y; const y: Array<2>;                        clean, file scope
//	declare module '2' { type Array<Y> = Y; const y: Array<2>; } clean, module scope
//	function f() { type Array<Y> = Y; const y: Array<2>; }       clean, function scope
//	type G<Array> = { a: Array<string> };                        clean, type parameter
//	{ type Array<Y> = Y; } const y: Array<2>;                    REPORTS, a sibling block does not shadow
var ArrayType = rule.Rule{
	Name: "@typescript-eslint/array-type",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The zero value is not the default and this is the normal path, not a defensive branch. A
		// rule configured as a bare `"error"` is handed nil options, because `DecodeOptionsInto`
		// errors on empty input and the config layer turns that into nil. A bare type assertion
		// would then yield an empty `default`, which matches no arm below, and the rule would
		// register on every file and report nothing while every fixture stayed green.
		parsed, decoded := rule.OptionsAs[ArrayTypeOptions](options)
		if !decoded {
			parsed = DefaultArrayTypeOptions()
		}

		return rule.Listeners{
			ast.KindArrayType: func(node *ast.Node) {
				reportArrayTypeSuffixForm(ctx, node, parsed)
			},
			ast.KindTypeReference: func(node *ast.Node) {
				reportArrayTypeGenericForm(ctx, node, parsed)
			},
		}
	},
}

// reportArrayTypeSuffixForm is upstream's `TSArrayType` visitor.
func reportArrayTypeSuffixForm(ctx rule.Context, node *ast.Node, options ArrayTypeOptions) {
	elementType := node.AsArrayTypeNode().ElementType
	if elementType == nil {
		return
	}

	// `readonly T[]` is a type operator wrapping the array, so the readonly-ness is read off the
	// parent and the finding is reported on the parent rather than on the array.
	readonlyOperator := arrayTypeReadonlyOperatorParentOf(node)
	isReadonly := readonlyOperator != nil

	current := options.settingFor(isReadonly)
	if current == ArrayTypeArray {
		return
	}
	if current == ArrayTypeArraySimple && arrayTypeIsSimple(elementType) {
		return
	}

	errorNode := node
	className := "Array"
	readonlyPrefix := ""
	if isReadonly {
		errorNode = readonlyOperator
		className = "ReadonlyArray"
		readonlyPrefix = "readonly "
	}

	messageType := arrayTypeMessageType(ctx, elementType)
	message := messageArrayTypeGeneric(className, messageType, readonlyPrefix)
	if current == ArrayTypeArraySimple {
		message = messageArrayTypeGenericSimple(className, messageType, readonlyPrefix)
	}

	// The element for the fix is the element AFTER skipping parentheses, matching TSESTree, whose
	// element type is already the inner node. Without the skip the repair would open its generic
	// inside the parentheses and leave the closing one stranded.
	inner := arrayTypeSkipParentheses(elementType)
	errorRange := type_checking.TrimNodeTextRange(ctx.SourceFile, errorNode)
	innerRange := type_checking.TrimNodeTextRange(ctx.SourceFile, inner)

	ctx.ReportRangeWithFixes(errorRange, message,
		rule.ReplaceRange(shimcore.NewTextRange(errorRange.Pos(), innerRange.Pos()), className+"<"),
		rule.ReplaceRange(shimcore.NewTextRange(innerRange.End(), errorRange.End()), ">"))
}

// reportArrayTypeGenericForm is upstream's `TSTypeReference` visitor.
func reportArrayTypeGenericForm(ctx rule.Context, node *ast.Node, options ArrayTypeOptions) {
	reference := node.AsTypeReferenceNode()
	typeName := reference.TypeName
	if typeName == nil || typeName.Kind != ast.KindIdentifier {
		return
	}

	name := typeName.Text()
	if name != "Array" && name != "ReadonlyArray" && name != "Readonly" {
		return
	}

	// A heritage clause is not a type position upstream can see, so it must not be judged here.
	//
	// This is a parser difference and it is the one place this port shipped a false positive that
	// the entire 193-case corpus could not see. TSESTree gives an interface's `extends` and a
	// class's `implements` their own node types, `TSInterfaceHeritage` and `TSClassImplements`, so
	// upstream's `TSTypeReference` visitor never fires on them. typescript-go reuses
	// `KindTypeReference` under a `KindHeritageClause` for both, so the listener fires and the rule
	// would report a rewrite that is not even legal: `interface I extends string[] {}` does not
	// parse.
	//
	// Measured. All four heritage forms are silent upstream against a control that reports:
	//
	//	interface I extends Array<string> {}          clean
	//	interface I extends ReadonlyArray<string> {}  clean
	//	class C implements Array<string> {}           clean
	//	class C extends Array<string> {}              clean
	//	let x: Array<string>;                         REPORTS, the control
	//
	// Found by the dry run rather than by any fixture: cohere reported 713 findings on the ahra tree
	// where ESLint reported 711, and both extra were `extends Array<...>` in one file. A class's
	// `extends` was already correct by accident, since typescript-go spells that one
	// `KindExpressionWithTypeArguments`, which this listener never receives. Only the other two
	// shapes needed the guard, and relying on the accident for the third would have been a coin
	// flip nobody documented.
	if node.Parent != nil && node.Parent.Kind == ast.KindHeritageClause {
		return
	}

	var typeArguments []*ast.Node
	if reference.TypeArguments != nil {
		typeArguments = reference.TypeArguments.Nodes
	}

	// `Readonly<T>` only concerns this rule when its argument is a suffix array, because only then
	// does it spell an array at all. Upstream tests `params[0].type` with optional chaining, so a
	// `Readonly` with no arguments falls out here rather than later.
	isReadonlyOfArray := name == "Readonly" &&
		len(typeArguments) > 0 && typeArguments[0].Kind == ast.KindArrayType
	if name == "Readonly" && !isReadonlyOfArray {
		return
	}

	if arrayTypeNameIsShadowed(node, name) {
		return
	}

	isReadonlyForm := name == "ReadonlyArray" || isReadonlyOfArray

	current := options.settingFor(isReadonlyForm)
	if current == ArrayTypeGeneric {
		return
	}

	readonlyPrefix := ""
	if isReadonlyForm {
		readonlyPrefix = "readonly "
	}

	// Upstream computes the message id before testing the argument count, and the order does not
	// matter because both paths return without reporting. Written in upstream's order anyway.
	if len(typeArguments) != 1 {
		return
	}
	if current == ArrayTypeArraySimple && !arrayTypeIsSimple(typeArguments[0]) {
		return
	}

	className := "Array"
	if isReadonlyForm {
		className = name
	}

	argument := typeArguments[0]
	messageType := arrayTypeMessageType(ctx, argument)

	var message rule.Message
	switch {
	case current == ArrayTypeArray && isReadonlyOfArray:
		message = messageArrayTypeArrayReadonly(className, messageType, readonlyPrefix)
	case current == ArrayTypeArray:
		message = messageArrayTypeArray(className, messageType, readonlyPrefix)
	case isReadonlyForm && name != "ReadonlyArray":
		message = messageArrayTypeArraySimpleReadonly(className, messageType, readonlyPrefix)
	default:
		message = messageArrayTypeArraySimple(className, messageType, readonlyPrefix)
	}

	// The repair wraps the argument in whatever punctuation the new spelling needs. Two independent
	// decisions, and upstream computes both before building either end of the rewrite.
	//
	// `typeParens` covers an argument whose suffix form would bind wrongly: a union, a function
	// type, an intersection, a type operator, an infer, a constructor type, a conditional, or a
	// reference literally named `ReadonlyArray`.
	//
	// `parentParens` covers the OUTER context: a readonly rewrite sitting directly inside another
	// suffix array needs its own parentheses, because `readonly string[][]` would read as an array
	// of readonly arrays only by accident. It is skipped when the element is already parenthesized.
	typeParens := arrayTypeNeedsParentheses(argument)
	parentParens := readonlyPrefix != "" && arrayTypeIsDirectElementOfAnArray(node)

	start := ""
	if parentParens {
		start += "("
	}
	start += readonlyPrefix
	if typeParens {
		start += "("
	}

	end := ""
	if typeParens {
		end += ")"
	}
	if !isReadonlyOfArray {
		end += "[]"
	}
	if parentParens {
		end += ")"
	}

	nodeRange := type_checking.TrimNodeTextRange(ctx.SourceFile, node)
	argumentRange := type_checking.TrimNodeTextRange(ctx.SourceFile, argument)

	ctx.ReportRangeWithFixes(nodeRange, message,
		rule.ReplaceRange(shimcore.NewTextRange(nodeRange.Pos(), argumentRange.Pos()), start),
		rule.ReplaceRange(shimcore.NewTextRange(argumentRange.End(), nodeRange.End()), end))
}

// arrayTypeReadonlyOperatorParentOf returns the `readonly` operator wrapping this array, or nil.
//
// Upstream tests `node.parent.type === TSTypeOperator && node.parent.operator === 'readonly'`.
// typescript-go spells the operator as a token kind on the same node shape, so this is the same
// test written against a different encoding. `keyof` and `unique` are the other two operators and
// neither makes the array readonly, so the token test is load-bearing rather than a formality.
func arrayTypeReadonlyOperatorParentOf(node *ast.Node) *ast.Node {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindTypeOperator {
		return nil
	}
	if parent.AsTypeOperatorNode().Operator != ast.KindReadonlyKeyword {
		return nil
	}
	return parent
}

// arrayTypeSkipParentheses is `ast.SkipTypeParentheses` with a nil guard.
//
// The shim helper dereferences its argument on the first loop test, so a nil reaches it as a panic
// rather than as a nil result. The walk recovers per FILE rather than per rule, so one such panic
// would cost every rule that file. The callers here all read a field that the parser can leave nil
// on malformed input, which is exactly the shape error recovery produces.
func arrayTypeSkipParentheses(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipTypeParentheses(node)
}

// arrayTypeIsSimple is upstream's `isSimpleType`.
//
// The parenthesis skip at the top has no counterpart upstream and is required by the same parser
// difference the rule's doc comment describes: TSESTree never hands this function a parenthesized
// node because it has none, so without the skip every `(string)` would answer false where upstream
// answers true.
func arrayTypeIsSimple(node *ast.Node) bool {
	node = arrayTypeSkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindAnyKeyword,
		ast.KindBooleanKeyword,
		ast.KindNeverKeyword,
		ast.KindNumberKeyword,
		ast.KindBigIntKeyword,
		ast.KindObjectKeyword,
		ast.KindStringKeyword,
		ast.KindSymbolKeyword,
		ast.KindUnknownKeyword,
		ast.KindVoidKeyword,
		ast.KindNullKeyword,
		ast.KindArrayType,
		ast.KindUndefinedKeyword,
		ast.KindThisType,
		ast.KindQualifiedName:
		return true

	case ast.KindTypeReference:
		reference := node.AsTypeReferenceNode()
		typeName := reference.TypeName
		if typeName == nil {
			return false
		}

		var typeArguments []*ast.Node
		if reference.TypeArguments != nil {
			typeArguments = reference.TypeArguments.Nodes
		}

		// `Array` gets a recursion no other name gets: a bare `Array` is simple and `Array<X>` is
		// simple exactly when `X` is. Every other reference is simple only when it takes no
		// arguments at all, so `Foo` is simple and `Foo<string>` is not.
		if typeName.Kind == ast.KindIdentifier && typeName.Text() == "Array" {
			if reference.TypeArguments == nil {
				return true
			}
			if len(typeArguments) == 1 {
				return arrayTypeIsSimple(typeArguments[0])
			}
			return false
		}

		if reference.TypeArguments != nil {
			return false
		}
		return arrayTypeIsSimple(typeName)
	}

	return false
}

// arrayTypeNeedsParentheses is upstream's `typeNeedsParentheses`.
//
// It answers whether the argument, written as `X[]`, would bind wrongly without parentheses. The
// `ReadonlyArray` arm is the odd one and it is upstream's: a reference to that name specifically
// needs wrapping because the rewrite puts `readonly ` in front of it.
//
// The parenthesis skip is again ours rather than upstream's, for the same parser reason: a
// parenthesized union arrives here as a real node and would answer false, where TSESTree hands
// upstream the union itself.
func arrayTypeNeedsParentheses(node *ast.Node) bool {
	node = arrayTypeSkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindTypeReference:
		typeName := node.AsTypeReferenceNode().TypeName
		if typeName == nil {
			return false
		}
		return arrayTypeNeedsParentheses(typeName)

	case ast.KindUnionType,
		ast.KindFunctionType,
		ast.KindIntersectionType,
		ast.KindTypeOperator,
		ast.KindInferType,
		ast.KindConstructorType,
		ast.KindConditionalType:
		return true

	case ast.KindIdentifier:
		return node.Text() == "ReadonlyArray"
	}

	return false
}

// arrayTypeIsDirectElementOfAnArray is upstream's `node.parent.type === TSArrayType &&
// !isParenthesized(node.parent.elementType, sourceCode)`.
//
// Upstream asks its `isParenthesized` helper, which looks for surrounding parenthesis TOKENS,
// because its tree has no node for them. Ours does, so the same question is "is the parent array's
// element a parenthesized node".
//
// # The parenthesis test is SUBSUMED here, and the subsumption is a parser difference
//
// A mutant forcing the last line to true survives, and the reason is not a fixture gap. TSESTree
// flattens parentheses, so `(ReadonlyArray<string>)[]` gives upstream a type reference whose PARENT
// IS the array, and its token-level `isParenthesized` is the only thing that can tell the two apart.
// Our parser keeps a `KindParenthesizedType` in between, so the same input gives this function a
// node whose parent is that parenthesized node, and the FIRST guard declines before the last line
// runs. Traced by running the mutant: the fixes it proposes for both parenthesized shapes are byte
// for byte the ones the unmutated rule proposes.
//
// Scored to separate the two: dropping the parent-kind guard fails four lines and neutralizing the
// whole helper fails twenty-three, so the function is load-bearing while its final line is not.
//
// The line is kept rather than deleted because it is upstream's question and because the
// subsumption rests on where the parser puts a node rather than on anything this rule decides. If a
// caller ever reached this with a node whose parent is an array despite a parenthesis, the guard
// would be the only thing answering correctly.
func arrayTypeIsDirectElementOfAnArray(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindArrayType {
		return false
	}
	element := parent.AsArrayTypeNode().ElementType
	if element == nil {
		return false
	}
	return element.Kind != ast.KindParenthesizedType
}

// arrayTypeMessageType is upstream's `getMessageType`.
//
// A simple element is rendered as its own source text and everything else collapses to the single
// letter `T`, so `Array<string | number>` reports as `Array<T>`. The collapse is invisible to a
// message-id assertion, which is why every reporting fixture asserts the rendered text.
//
// The text is taken from the node AS WRITTEN, parentheses included, because upstream slices the
// range of the node TSESTree gave it and that range covers whatever the source held. Only the
// simplicity JUDGMENT skips parentheses.
func arrayTypeMessageType(ctx rule.Context, node *ast.Node) string {
	if !arrayTypeIsSimple(node) {
		return "T"
	}
	trimmed := type_checking.TrimNodeTextRange(ctx.SourceFile, node)
	return ctx.SourceFile.Text()[trimmed.Pos():trimmed.End()]
}

// ArrayTypeSetting is one of upstream's three spellings for one axis.
type ArrayTypeSetting string

const (
	// ArrayTypeArray is `"array"`: always write the suffix form.
	ArrayTypeArray ArrayTypeSetting = "array"

	// ArrayTypeArraySimple is `"array-simple"`: write the suffix form for a simple element type and
	// the generic form for anything else, so the brackets never trail a complicated type.
	ArrayTypeArraySimple ArrayTypeSetting = "array-simple"

	// ArrayTypeGeneric is `"generic"`: always write `Array<T>`.
	ArrayTypeGeneric ArrayTypeSetting = "generic"
)

// ArrayTypeOptions is the rule's configuration.
//
// Upstream's schema is one object with `default` and `readonly`, both drawn from the three
// spellings above, `additionalProperties: false`, and `defaultOptions` of `{ default: 'array' }`.
type ArrayTypeOptions struct {
	// Default governs a mutable array and is upstream's `default`, whose own default is `array`.
	Default ArrayTypeSetting

	// Readonly governs `readonly T[]`, `ReadonlyArray<T>` and `Readonly<T[]>`. Upstream's `readonly`
	// is optional and falls back to `default` rather than to a fixed value, so this field holds the
	// resolved answer and the decoder does the falling back.
	Readonly ArrayTypeSetting
}

// settingFor picks the axis a node is judged on.
func (o ArrayTypeOptions) settingFor(isReadonly bool) ArrayTypeSetting {
	if isReadonly {
		return o.Readonly
	}
	return o.Default
}

// DefaultArrayTypeOptions is upstream's configured-nothing behavior.
//
// Exported because a fixture asserting the default has to be able to name it, and because a caller
// starting from the Go zero value would get two empty settings, which match no arm and silence the
// rule on every input while looking configured.
func DefaultArrayTypeOptions() ArrayTypeOptions {
	return ArrayTypeOptions{
		Default:  ArrayTypeArray,
		Readonly: ArrayTypeArray,
	}
}

// arrayTypeRawOptions is the wire shape.
//
// Both fields are pointers so an absent key stays distinguishable from an explicit one long enough
// for the fallbacks to be applied. That distinction is the whole content of this decoder: `readonly`
// absent means "whatever `default` resolved to", and no non-pointer field can say that.
type arrayTypeRawOptions struct {
	Default  *string `json:"default"`
	Readonly *string `json:"readonly"`
}

// DecodeArrayTypeOptions maps the wire keys onto the settings the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` for two reasons, and the second is the one that
// would have shipped silently. Neither default is a Go zero value, so a generic decode over a
// non-pointer struct yields two empty strings that match no arm and silence the rule. And `readonly`
// falls back to the RESOLVED `default` rather than to a constant, which is a dependency between two
// fields that no struct tag can express.
//
// Empty input and `null` are upstream's defaults. A spelling outside upstream's three is refused,
// naming the key and the value. This used to fall back to the default for an unknown spelling, which
// ran a setting nobody wrote and loaded clean doing it; upstream's schema is an enum and refuses it at
// load, and the config layer now refuses a decoder's error by rule name (#rfbha44).
func DecodeArrayTypeOptions(raw []byte) (any, error) {
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return DefaultArrayTypeOptions(), nil
	}
	decoded, err := rule.DecodeOptionsInto[arrayTypeRawOptions]()(raw)
	if err != nil {
		return DefaultArrayTypeOptions(), fmt.Errorf("expected an object of default and readonly, got %s: %w", trimmed, err)
	}

	wire, _ := decoded.(arrayTypeRawOptions)
	options := DefaultArrayTypeOptions()

	if wire.Default != nil {
		setting, recognized := arrayTypeSettingFromWire(*wire.Default)
		if !recognized {
			return DefaultArrayTypeOptions(), fmt.Errorf(`default %q is not "array", "array-simple" or "generic"`, *wire.Default)
		}
		options.Default = setting
	}

	// Resolved after `default`, and from it, which is upstream's `options.readonly ?? defaultOption`.
	options.Readonly = options.Default
	if wire.Readonly != nil {
		setting, recognized := arrayTypeSettingFromWire(*wire.Readonly)
		if !recognized {
			return DefaultArrayTypeOptions(), fmt.Errorf(`readonly %q is not "array", "array-simple" or "generic"`, *wire.Readonly)
		}
		options.Readonly = setting
	}

	return options, nil
}

// arrayTypeSettingFromWire reads one of upstream's three spellings.
func arrayTypeSettingFromWire(value string) (ArrayTypeSetting, bool) {
	switch ArrayTypeSetting(value) {
	case ArrayTypeArray:
		return ArrayTypeArray, true
	case ArrayTypeArraySimple:
		return ArrayTypeArraySimple, true
	case ArrayTypeGeneric:
		return ArrayTypeGeneric, true
	}
	return "", false
}

// arrayTypeNameIsShadowed answers upstream's scope walk: does any ENCLOSING scope bind this name?
//
// Upstream writes `for (let scope = getScope(node); scope.upper; scope = scope.upper) if
// (scope.set.has(name)) return;`, which is eslint-scope. There is no scope shelf in this tree, so
// the walk is over the syntax tree instead, and the two have to agree about three things: which
// nodes open a scope, which declarations bind a name into the scope they sit in, and that only
// ENCLOSING scopes count.
//
// All three were measured against the installed build rather than derived, because a scope model
// written from intuition is exactly the thing that looks right and is wrong in one arm.
//
// # Every declaration form that binds, measured
//
// Upstream's `scope.set` holds variables regardless of whether the name is a type or a value, so a
// value declaration shadows a type position. All of these are CLEAN upstream with `default: array`,
// against a control of `const y: Array<2>;` which reports:
//
//	type Array<Y> = Y;                 declare const Array: unknown;
//	interface Array<Y> { z: Y }        declare let Array: unknown;
//	class Array<Y> { z: Y }            declare function Array(): void;
//	enum Array { A }                   import { Array } from 'm';
//	namespace Array { ... }            function f(Array: unknown) { ... }
//	type G<Array> = ...                a type parameter
//
// # And what does NOT bind
//
// A member is not a variable, so none of these shadow and all three REPORT:
//
//	class C { Array: unknown; m(): Array<2> { ... } }
//	interface I { Array: unknown; m(): Array<2>; }
//	type T = { Array: unknown; m(): Array<2> };
//
// # Only enclosing scopes, and hoisting does not matter
//
//	{ type Array<Y> = Y; } const y: Array<2>;    REPORTS, a sibling block is not enclosing
//	const y: Array<2>; type Array<Y> = Y;        CLEAN, a later declaration in the SAME scope binds
//
// The second is why this collects a scope's declarations rather than only those written above the
// reference: the binding is lexical, not positional.
func arrayTypeNameIsShadowed(node *ast.Node, name string) bool {
	// Upstream's loop condition is `scope.upper`, so the OUTERMOST scope is never examined. That
	// outermost scope is eslint-scope's GLOBAL scope, where the real `Array` lives, and skipping it
	// is what keeps the rule from declining every input.
	//
	// The file is NOT that scope. A file gets its own module or script scope whose `upper` is the
	// global one, so the file's own declarations are examined and a file-scope `type Array<Y>` does
	// shadow. An earlier draft stopped at the source file on the reasoning that it was the
	// outermost scope, and eleven measured rows disagreed: every file-scope shadowing form reported
	// where upstream is silent. There is no node here for the global scope, so the walk simply runs
	// to the root and the built-in `Array` is never in it.
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if !arrayTypeOpensAScope(ancestor) {
			continue
		}
		if arrayTypeScopeBinds(ancestor, name) {
			return true
		}
	}
	return false
}

// arrayTypeOpensAScope answers whether a node introduces a lexical scope a declaration can bind in.
//
// A deliberately generous list: a node that does not open a scope contributes no declarations, so
// including one that should not be here can only cause a name to be looked for in a place where it
// will not be found. The failure direction that matters is the other one, and the enumeration above
// is what guards it.
func arrayTypeOpensAScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindBlock,
		ast.KindModuleBlock,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor,
		ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindForStatement,
		ast.KindForInStatement,
		ast.KindForOfStatement,
		ast.KindCatchClause,
		ast.KindCaseBlock,
		ast.KindMappedType,
		ast.KindConditionalType:
		return true
	}
	return false
}

// arrayTypeScopeBinds answers whether one scope declares the name.
//
// Four sources, matching what eslint-scope puts in a scope's variable set: a type parameter, a
// function or catch parameter, a variable or declaration statement in the scope's body, and a
// binding pattern inside any of those. A class or interface MEMBER is deliberately absent, measured
// on the installed build and pinned by three fixtures.
//
// # Every list accessor here is kind-gated, and that is a crash guard
//
// `Parameters()` reaches `FunctionLikeData()` and dereferences it, and `TypeParameterList()` panics
// on its default arm, so both take the run down on a node that has no such list. The walk recovers
// per FILE rather than per rule, so one panic costs every rule that file. This was not theoretical:
// an ungated `Parameters()` panicked on upstream's own sixty-third passing case the first time the
// corpus ran, which is a shape a hand-written fixture would not have reached.
func arrayTypeScopeBinds(scope *ast.Node, name string) bool {
	switch scope.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression,
		ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		if arrayTypeAnyDeclarationNameIs(scope.TypeParameters(), name) {
			return true
		}

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		// A function-like node owns both lists, and both bind into the scope it opens.
		if arrayTypeAnyDeclarationNameIs(scope.TypeParameters(), name) {
			return true
		}
		if arrayTypeAnyDeclarationNameIs(scope.Parameters(), name) {
			return true
		}

	case ast.KindCatchClause:
		if variable := scope.AsCatchClause().VariableDeclaration; variable != nil {
			if arrayTypeDeclarationNameIs(variable, name) {
				return true
			}
		}

	case ast.KindForStatement:
		// `for (let Array = 0; ; )` binds into the loop's own scope. The initializer is an
		// expression or a declaration list, and only the latter binds.
		if initializer := scope.AsForStatement().Initializer; initializer != nil {
			if arrayTypeDeclarationListBinds(initializer, name) {
				return true
			}
		}

	case ast.KindForInStatement, ast.KindForOfStatement:
		if initializer := scope.AsForInOrOfStatement().Initializer; initializer != nil {
			if arrayTypeDeclarationListBinds(initializer, name) {
				return true
			}
		}
	}

	for _, statement := range arrayTypeStatementsOf(scope) {
		if arrayTypeStatementBinds(statement, name) {
			return true
		}
	}

	return false
}

// arrayTypeAnyDeclarationNameIs answers whether any declaration in a list binds the name.
func arrayTypeAnyDeclarationNameIs(declarations []*ast.Node, name string) bool {
	for _, declaration := range declarations {
		if arrayTypeDeclarationNameIs(declaration, name) {
			return true
		}
	}
	return false
}

// arrayTypeDeclarationListBinds answers whether a variable declaration list binds the name.
//
// The node reaches here from a loop initializer, which is an expression when the loop reuses an
// existing binding and a declaration list when it introduces one. Only the second binds.
func arrayTypeDeclarationListBinds(node *ast.Node, name string) bool {
	if node == nil || node.Kind != ast.KindVariableDeclarationList {
		return false
	}
	for _, declaration := range node.AsVariableDeclarationList().Declarations.Nodes {
		if arrayTypeDeclarationNameIs(declaration, name) {
			return true
		}
	}
	return false
}

// arrayTypeStatementsOf returns the statements a scope holds, or nothing for a scope that has none.
//
// `Statements()` PANICS on a kind with no statement list, so the kinds are named rather than tried.
// A function's own body is a Block, which the walk reaches on its own as an ancestor, so nothing is
// read for a function-like node here: it opens a scope for its parameters and type parameters, which
// the switch above already read.
func arrayTypeStatementsOf(scope *ast.Node) []*ast.Node {
	switch scope.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock:
		return scope.Statements()
	}
	return nil
}

// arrayTypeStatementBinds answers whether one statement declares the name.
//
// A variable statement binds every name in its declaration list; every other declaration form binds
// its own name. `import { Array } from 'm'` binds through its named bindings, which is a shape with
// no `Name()` of its own.
func arrayTypeStatementBinds(statement *ast.Node, name string) bool {
	switch statement.Kind {
	case ast.KindVariableStatement:
		list := statement.AsVariableStatement().DeclarationList
		if list == nil {
			return false
		}
		for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
			if arrayTypeDeclarationNameIs(declaration, name) {
				return true
			}
		}
		return false

	case ast.KindTypeAliasDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindClassDeclaration,
		ast.KindEnumDeclaration,
		ast.KindFunctionDeclaration,
		ast.KindModuleDeclaration:
		return arrayTypeDeclarationNameIs(statement, name)

	case ast.KindImportDeclaration:
		return arrayTypeImportBinds(statement, name)
	}

	return false
}

// arrayTypeImportBinds answers whether an import statement brings the name into scope.
//
// Three shapes bind: a default import, a namespace import, and one of a named import's elements.
// `imports.BindingsOf` is the shelf function that decides which of the three an import carries, and
// it is called rather than respelled. Reading it first was worth it: it guards a nil element list
// that a hand-written walk of `AsNamedImports().Elements.Nodes` dereferences, which is a shape error
// recovery produces from `import {} from` and `import { from`.
func arrayTypeImportBinds(statement *ast.Node, name string) bool {
	bindings := imports.BindingsOf(statement)

	if arrayTypeBindingNameIs(bindings.Default, name) {
		return true
	}
	if arrayTypeDeclarationNameIs(bindings.Namespace, name) {
		return true
	}
	for _, element := range bindings.Named {
		if arrayTypeDeclarationNameIs(element, name) {
			return true
		}
	}
	return false
}

// arrayTypeDeclarationNameIs answers whether a declaration binds this name.
//
// A plain declaration answers on its own identifier. A DESTRUCTURING pattern binds names that are
// not the declaration's name, so the pattern is walked. That arm is here because it was measured
// rather than reasoned about: an earlier draft returned false for a pattern with the gap recorded at
// this line, and the installed build disagrees.
//
//	function f() { const { Array } = x; const y: Array<2>; }   CLEAN upstream
//	function f() { const [Array] = x;   const y: Array<2>; }   CLEAN upstream
//	function f() { const Array = x;     const y: Array<2>; }   CLEAN, the named control
//	function f() {                      const y: Array<2>; }   REPORTS, the absent control
//
// So a documented gap would have shipped a false positive on both destructuring forms, and the
// documentation would have told the next reader the narrowness was deliberate.
func arrayTypeDeclarationNameIs(declaration *ast.Node, name string) bool {
	if declaration == nil {
		return false
	}
	return arrayTypeBindingNameIs(declaration.Name(), name)
}

// arrayTypeBindingNameIs answers whether a binding name, possibly a pattern, binds this name.
//
// Three shapes: an identifier binds itself, an object pattern binds through its elements, and an
// array pattern likewise. A binding element's own `Name()` is what it introduces, so `{ a: Array }`
// binds `Array` rather than `a`, and the recursion handles a nested pattern.
func arrayTypeBindingNameIs(bindingName *ast.Node, name string) bool {
	if bindingName == nil {
		return false
	}

	switch bindingName.Kind {
	case ast.KindIdentifier:
		return bindingName.Text() == name

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		for _, element := range bindingName.AsBindingPattern().Elements.Nodes {
			// An array pattern's hole is an omitted expression, which has no name.
			if element.Kind != ast.KindBindingElement {
				continue
			}
			if arrayTypeBindingNameIs(element.Name(), name) {
				return true
			}
		}
	}

	return false
}
