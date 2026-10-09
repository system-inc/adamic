package core

import (
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noEmptyFunctionAllowKinds is upstream's `ALLOW_OPTIONS`, in its order.
//
// Fourteen values, and four of them describe TypeScript rather than JavaScript. That is worth
// stating because the typescript-eslint EXTENSION of this rule exists and reads as though it adds
// them: measured against eslint 10.8.1, core already handles a private constructor, a protected
// constructor, a constructor taking parameter properties, a decorated method and an override
// method. The extension's only real difference is SPELLING, `private-constructors` against
// `privateConstructors`, and core refuses the kebab form at config load.
//
// # That claim is now measured across the whole extension corpus rather than probed
//
// `@typescript-eslint/no-empty-function` is not ported here, and this is why. Driving upstream's
// core and upstream's extension over identical inputs on the installed 8.67.0 build, across the
// extension's own corpus of nine passing and seven failing cases: ZERO behavioural differences on
// every case where both configurations load. The only two divergent rows are the ones spelling an
// allow value in kebab, which eslint core rejects at config load rather than answering differently.
//
// `decoratedFunctions` and `overrideMethods` are spelled identically in both, and a parameter
// property is exempt in both UNCONDITIONALLY rather than behind an allow value, which is the part
// most likely to be misread from the extension's source.
//
// This rule was then run against that same extension corpus with the two kebab spellings translated
// and reproduced it exactly, nine of nine passing and seven of seven failing.
//
// So porting the extension would register a second NAME rather than a second check. A rule package
// may not import another rule package, so a wrapper would need this file's body lifted into
// `internal/utilities`, and the captured rule inventory carried no entry for this rule under either
// spelling, so nothing in the differential asks for the name. The audit lists the extension as an
// unported line item, which reads as work outstanding; it is not.
var noEmptyFunctionAllowKinds = []string{
	"functions",
	"arrowFunctions",
	"generatorFunctions",
	"methods",
	"generatorMethods",
	"getters",
	"setters",
	"constructors",
	"asyncFunctions",
	"asyncMethods",
	"privateConstructors",
	"protectedConstructors",
	"decoratedFunctions",
	"overrideMethods",
}

// NoEmptyFunctionOptions carries upstream's one option.
//
// `Allow` defaults to an EMPTY LIST, so the zero value is the safe direction: an absent key and an
// explicit empty list mean the same thing, and nothing is silently exempted.
type NoEmptyFunctionOptions struct {
	Allow []string `json:"allow"`
}

// DecodeNoEmptyFunctionOptions turns the configured object into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so empty input returns the defaults instead of an
// error the config layer turns into a nil. That path is harmless for this rule, whose one option
// defaults to empty, and it is written out anyway because the harmlessness is a property of today's
// default rather than of the code.
//
// A kind outside upstream's enum is refused rather than ignored. Upstream gets that refusal from its
// schema before the rule runs; there is no schema layer here, so it lives here, and a rule silently
// dropping an unknown kind would leave a project believing it had exempted something it had not.
func DecodeNoEmptyFunctionOptions(raw []byte) (any, error) {
	var options NoEmptyFunctionOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	for _, kind := range options.Allow {
		if !slices.Contains(noEmptyFunctionAllowKinds, kind) {
			return options, fmt.Errorf("no-empty-function does not know the kind %q; the kinds are %v",
				kind, noEmptyFunctionAllowKinds)
		}
	}
	return options, nil
}

var messageNoEmptyFunction = rule.Message{
	Id: "unexpected",
	Description: "This function body is empty, so nothing says whether that is deliberate. An empty " +
		"body reads identically whether the author meant a no-op, forgot to write it, or deleted " +
		"the contents and left the shell. A comment inside costs one line and settles it.",
}

var messageNoEmptyFunctionSuggestComment = rule.Message{
	Id:          "suggestComment",
	Description: "Add a comment inside the empty body, saying the emptiness is deliberate.",
}

// NoEmptyFunction flags a function whose body is an empty block with no comment in it.
//
//	valid:   function foo() { bar(); }
//	valid:   function foo() { /* empty */ }
//	valid:   class A { constructor(private x: number) {} }
//	invalid: function foo() {}
//	invalid: var foo = () => {};
//	invalid: class A { get foo() {} }
//
// A comment is what separates a deliberate no-op from a body somebody forgot to write, which is why
// the rule accepts one rather than demanding a statement.
//
// # Anchoring, which differs from upstream by more than it looks
//
// Upstream listens on three function types and reads `node.parent` to decide what it is: a `Property`
// with `kind: "get"` is a getter, a `MethodDefinition` with `kind: "constructor"` is a constructor,
// and so on. Our parser produces those as their OWN node kinds, so there is no wrapper to read: a
// class getter arrives as `KindGetAccessor` directly. This listens on seven kinds and reads the node
// rather than its parent.
//
// One place the parent still matters, and it is upstream's `parent.method` test:
// `var obj = { foo: function() {} }` is a plain function expression assigned to a property, which
// upstream classifies as `functions` rather than `methods`, while `var obj = { foo() {} }` is a
// method. Our parser draws the same line by giving the first a `KindFunctionExpression` under a
// `KindPropertyAssignment` and the second a `KindMethodDeclaration`.
//
// # The comment test is the whole clean set
//
// Three of the four clean patterns upstream generates per seed are a body holding a comment, so a
// port that only counted statements would report roughly a third of the corpus. `comments.ForFile`
// is the shelf's cached per-file scan and answers which comments fall inside a range.
//
// # The span is the body, not the function
//
// Upstream reports the function node with an explicit `loc` of `node.body.loc`, so the finding
// points at `{}`. Measured: `function a() {}` reports columns 14 to 16. A port reporting the
// function would pass all 1740 message-id fixtures while pointing at the wrong place on every one.
//
// # A suggestion, not a fix
//
// Upstream sets `hasSuggestions` and no `fixable`. Writing `/* empty */` into a body asserts that
// the emptiness is deliberate, which is a claim about intent rather than a change of spelling, so
// nothing may apply it unattended.
var NoEmptyFunction = rule.Rule{
	Name: "no-empty-function",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as bare "error" is handed nil, so the type assertion yields the zero
		// value. Correct here only because `allow` defaults to empty.
		settings, _ := rule.OptionsAs[NoEmptyFunctionOptions](options)

		report := func(node *ast.Node) {
			body := node.Body()
			// An overload signature, an abstract method and a declared function all have no body,
			// and none of them is an empty function. An arrow with a concise body has one that is
			// not a block, which upstream declines the same way.
			if body == nil || body.Kind != ast.KindBlock {
				return
			}
			if len(body.AsBlock().Statements.Nodes) != 0 {
				return
			}
			if noEmptyFunctionBodyHasComment(ctx, body) {
				return
			}
			if noEmptyFunctionIsAllowed(node, settings.Allow) {
				return
			}

			bodyRange := rule.TokenRange(ctx.SourceFile, body)
			ctx.ReportRangeWithSuggestions(bodyRange, rule.Message{
				Id: messageNoEmptyFunction.Id,
				Description: fmt.Sprintf("Unexpected empty %s.",
					noEmptyFunctionNameWithKind(node)),
			}, rule.Suggestion{
				Message: messageNoEmptyFunctionSuggestComment,
				// Upstream replaces the range BETWEEN the braces, which is why the output keeps the
				// braces the reported span includes: `{}` becomes `{ /* empty */ }`.
				Fixes: []rule.Fix{
					rule.ReplaceRange(
						core.NewTextRange(bodyRange.Pos()+1, bodyRange.End()-1), " /* empty */ "),
				},
			})
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: report,
			ast.KindFunctionExpression:  report,
			ast.KindArrowFunction:       report,
			ast.KindMethodDeclaration:   report,
			ast.KindGetAccessor:         report,
			ast.KindSetAccessor:         report,
			ast.KindConstructor:         report,
		}
	},
}

// noEmptyFunctionBodyHasComment reports whether any comment falls inside a body's braces.
//
// Upstream asks the source code for comment tokens within the body. `comments.ForFile` is the
// shelf's cached scan and answers the same question; reimplementing `GetLeadingCommentRanges` beside
// it is a thing two porters have nearly done.
//
// The range tested is the body's own, which already spans brace to brace, so a comment written
// before the function or after it falls outside and does not exempt.
func noEmptyFunctionBodyHasComment(ctx rule.Context, body *ast.Node) bool {
	bodyRange := rule.TokenRange(ctx.SourceFile, body)
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= bodyRange.Pos() && comment.Range.End() <= bodyRange.End() {
			return true
		}
	}
	return false
}

// noEmptyFunctionKind classifies a function the way upstream's `getKind` does.
//
// The prefix rule is upstream's exactly: a generator or an async marker turns `methods` into
// `generatorMethods` or `asyncMethods`, and the two never combine, because upstream returns on the
// first that matches. So an async generator method classifies as `generatorMethods`, which reads
// like an oversight and is reproduced rather than corrected.
func noEmptyFunctionKind(node *ast.Node) string {
	if node.Kind == ast.KindArrowFunction {
		return "arrowFunctions"
	}

	kind := "functions"
	switch node.Kind {
	case ast.KindGetAccessor:
		return "getters"
	case ast.KindSetAccessor:
		return "setters"
	case ast.KindConstructor:
		return "constructors"
	case ast.KindMethodDeclaration:
		kind = "methods"
	}

	// Generator first, then async, and never both. Upstream's `if/else if` over `node.generator`
	// and `node.async`.
	if consistentReturnIsGenerator(node) {
		return "generator" + noEmptyFunctionCapitalize(kind)
	}
	if noEmptyFunctionHasModifier(node, ast.KindAsyncKeyword) {
		return "async" + noEmptyFunctionCapitalize(kind)
	}
	return kind
}

// noEmptyFunctionCapitalize upper-cases the first letter, so `methods` becomes `Methods`.
func noEmptyFunctionCapitalize(kind string) string {
	if kind == "" {
		return kind
	}
	return string(kind[0]-('a'-'A')) + kind[1:]
}

// noEmptyFunctionHasModifier reports whether a node carries a modifier keyword.
func noEmptyFunctionHasModifier(node *ast.Node, want ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == want {
			return true
		}
	}
	return false
}

// noEmptyFunctionIsAllowed reports whether the options exempt this function.
//
// Three tests, in upstream's order. The plain kind match is the common one. The constructor arm adds
// accessibility and, unconditionally, a constructor taking parameter properties, because
// `constructor(private x: number)` is not an empty constructor at all: the parameter list IS the
// body, assigning the field. The accessor and method arm adds decorators and `override`.
func noEmptyFunctionIsAllowed(node *ast.Node, allow []string) bool {
	kind := noEmptyFunctionKind(node)
	if slices.Contains(allow, kind) {
		return true
	}

	if kind == "constructors" {
		if noEmptyFunctionHasModifier(node, ast.KindPrivateKeyword) &&
			slices.Contains(allow, "privateConstructors") {
			return true
		}
		if noEmptyFunctionHasModifier(node, ast.KindProtectedKeyword) &&
			slices.Contains(allow, "protectedConstructors") {
			return true
		}
		// Unconditional, with no option behind it: upstream exempts a parameter-property
		// constructor whatever `allow` says, because the parameters are doing the work a body
		// otherwise would.
		if noEmptyFunctionHasParameterProperties(node) {
			return true
		}
	}

	// Upstream's `/(?:g|s)etters|methods$/iu` over the kind string, which matches `getters`,
	// `setters`, `methods`, `generatorMethods` and `asyncMethods` and nothing else.
	if kind == "getters" || kind == "setters" ||
		kind == "methods" || kind == "generatorMethods" || kind == "asyncMethods" {
		if len(node.Decorators()) > 0 && slices.Contains(allow, "decoratedFunctions") {
			return true
		}
		if noEmptyFunctionHasModifier(node, ast.KindOverrideKeyword) &&
			slices.Contains(allow, "overrideMethods") {
			return true
		}
	}
	return false
}

// noEmptyFunctionHasParameterProperties reports whether any parameter carries an accessibility or
// `readonly` modifier, which makes it a field declaration rather than a plain parameter.
func noEmptyFunctionHasParameterProperties(node *ast.Node) bool {
	functionLike := node.FunctionLikeData()
	if functionLike == nil || functionLike.Parameters == nil {
		return false
	}
	for _, parameter := range functionLike.Parameters.Nodes {
		if modifiers := parameter.Modifiers(); modifiers != nil && len(modifiers.Nodes) > 0 {
			return true
		}
	}
	return false
}

// noEmptyFunctionNameWithKind is upstream's `getFunctionNameWithKind`, for every kind this rule can
// be handed.
//
// A neighbouring rule in this package has a version of the same helper and its doc comment says it
// is "narrowed to the two kinds this rule can be handed", so it cannot serve here: this corpus
// asserts TWENTY distinct renderings and that one produces three of them.
//
// The shape, read off those twenty rather than inferred:
//
//	arrow function                    an arrow, never named even when assigned to a variable
//	async arrow function              the async prefix reaches arrows too
//	function                          an anonymous function expression
//	function 'foo'                    a named one, single quotes
//	generator function 'foo'          generator and async are prefixes on the noun
//	constructor                       never named, even though the class has a name
//	getter 'foo' / setter 'foo'       the accessor's key
//	static method 'foo'               static is a prefix in the NAME, though it is not an allow kind
//	generator method 'foo'            for `{foo: function*() {}}`, where the name is the PROPERTY's
//
// That last row is the one worth pausing on. The function there is anonymous, and upstream still
// renders `foo`, because it reads the name from the enclosing property rather than from the
// function. So the name and the KIND come from different places for the same node.
func noEmptyFunctionNameWithKind(node *ast.Node) string {
	prefix := ""
	if noEmptyFunctionHasModifier(node, ast.KindStaticKeyword) {
		prefix = "static "
	}
	if consistentReturnIsGenerator(node) {
		prefix += "generator "
	} else if noEmptyFunctionHasModifier(node, ast.KindAsyncKeyword) {
		prefix += "async "
	}

	switch node.Kind {
	case ast.KindArrowFunction:
		// Never named, even when assigned to a variable, and upstream's corpus asserts that for
		// `const foo = async () => {}`.
		return prefix + "arrow function"
	case ast.KindConstructor:
		// Never named, and that is a property of the PARSER rather than a choice here. A mutation
		// passing `node.Name()` to the joining helper survives every fixture, because a
		// `KindConstructor` carries no name on any input: probed across five shapes including a
		// named class expression and a quoted `'constructor'` key, which parses as an ordinary
		// method rather than a constructor. There is no input on which the two versions differ.
		return prefix + "constructor"
	case ast.KindGetAccessor:
		return noEmptyFunctionJoinName(prefix+"getter", node.Name())
	case ast.KindSetAccessor:
		return noEmptyFunctionJoinName(prefix+"setter", node.Name())
	case ast.KindMethodDeclaration:
		return noEmptyFunctionJoinName(prefix+"method", node.Name())
	}

	// A function declaration or expression. Its own name wins; failing that, the property or
	// variable it is being assigned to supplies one, which is how `{foo: function*() {}}` renders
	// as `generator method 'foo'` rather than as an anonymous generator function.
	if name := node.Name(); name != nil {
		return noEmptyFunctionJoinName(prefix+"function", name)
	}
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAssignment {
		// Upstream calls this a METHOD here rather than a function, because the property makes it
		// one, while its `getKind` still classifies it as `functions` for the allow list. The two
		// answers genuinely differ for the same node.
		return noEmptyFunctionJoinName(prefix+"method", parent.Name())
	}
	return prefix + "function"
}

// noEmptyFunctionJoinName appends a quoted name when there is one to append.
func noEmptyFunctionJoinName(kind string, name *ast.Node) string {
	if name == nil {
		return kind
	}
	text, settled := property.Name(name, property.Static)
	if !settled || text == "" {
		return kind
	}
	return kind + " '" + text + "'"
}
