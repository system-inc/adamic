package typescript

import (
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedThis = rule.Message{
	Id: "unexpectedThis",
	Description: "This function is called rather than constructed and nothing binds `this` to it, so " +
		"`this` is `undefined` here and every property read off it throws at runtime. The usual " +
		"cause is a method that was moved out of a class or an object literal and left its " +
		"`this.` prefixes behind. Take what it needs as a parameter, give the function a `this` " +
		"parameter so TypeScript can check the call sites, or make it a method again.",
}

// NoInvalidThis flags a `this` in a function where `this` is not bound to anything.
//
//	valid:   function foo(this: SomeType) { this.prop; }
//	valid:   class A { foo() { this.x; } }
//	valid:   const obj = { foo() { this.x; } };
//	valid:   function Foo() { this.x = 1; }              (capitalized, assumed a constructor)
//	valid:   array.forEach(function () { this.x; }, thisArg);
//	valid:   function foo() { this.x; }.bind(obj);
//	invalid: function foo() { this.x; }
//	invalid: const foo = function () { this.x; };
//	invalid: array.forEach(function () { this.x; });
//
// Ported from typescript-eslint's `no-invalid-this`, which extends ESLint core's rule of the same
// name. Both were read: the TypeScript layer adds the `this` parameter and the class-field arms, and
// core supplies the judgment through `astUtils.isDefaultThisBinding`, which is the real algorithm and
// is ported below as thisIsDefaultBoundIn.
//
// # The source type collapses most of core's machinery, and this was measured rather than assumed
//
// Core's rule is built on ESLint's code path analysis and consults `scope.isStrict` on every
// function, because in sloppy mode `this` falls back to the global object and is therefore always
// valid. It also special-cases `sourceType: module` and the `globalReturn` parser feature at the
// program level.
//
// None of that survives contact with what cohere lints. Every file here is a TypeScript module, so
// every function body is strict and top-level `this` is always `undefined`. That was confirmed against
// the installed rule at 8.67.0 rather than reasoned about: the whole 91-case corpus was driven through
// the ESLint Linter interface twice, once with the source types the corpus declares and once with
// every case forced to `module`, and the two runs agree on all 91. Two corpus cases carry
// `ecmaFeatures: { globalReturn: true }`, which reads as though the feature is what makes them report;
// it is not. Measured separately, they report under `sourceType: module` with no `globalReturn` at all
// and stay silent under `script` WITH it, so the module-ness is doing the work and the parser feature
// is doing none.
//
// So the strictness question has one answer here and the code path layer reduces to a stack of
// function nodes, which is what the walk below maintains.
//
// # The stack is upstream's, and its arms are not interchangeable
//
// typescript-eslint pushes onto a validity stack for four node types and core pushes for every code
// path, and the difference between "push true" and "push the answer to a question" is the whole
// TypeScript contribution:
//
//	FunctionDeclaration / FunctionExpression   push whether a `this` parameter is declared
//	PropertyDefinition / AccessorProperty      push true, a field initializer's `this` is the instance
//	ArrowFunctionExpression                    push NOTHING, an arrow inherits the enclosing binding
//
// The arrow arm is an absence rather than a case, and it is what makes
// `function foo() { this.x; z(() => this.x); }` report twice rather than once: both `this` reads
// resolve against the same enclosing function. A port that pushed for arrows would report the outer
// one and silently exempt the inner.
//
// A class static block is not on that list and is handled by the default-binding test instead, where
// upstream returns false for it outright.
//
// # capIsConstructor is on by default, so the zero-value struct is the wrong configuration
//
// The option surface is one boolean and its default is TRUE, which means a decoder that fills a
// zero-value struct inverts it and silently reports every `function Foo() { this.x = 1; }` in the
// tree. The corpus tests the option on eight cases and the decoder below is hand written for exactly
// this reason.
var NoInvalidThis = rule.Rule{
	Name: "@typescript-eslint/no-invalid-this",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := noInvalidThisSettingsFrom(options)

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// The walk is done here rather than through per-kind listeners because the rule
				// needs to know what it is INSIDE when it meets a `this`, and the listener walk is
				// pre-order with no exit hook to pop a stack on.
				walker := &noInvalidThisWalker{ctx: ctx, settings: settings}
				walker.walk(node)
			},
		}
	},
}

// noInvalidThisSettings is what the rule reads after decoding.
type noInvalidThisSettings struct {
	// CapIsConstructor treats a function whose name starts with an uppercase letter as a
	// constructor, so `this` inside it is bound. Defaults to true, which is why this type exists
	// separately from its wire shape.
	CapIsConstructor bool
}

// DefaultNoInvalidThisSettings is upstream's `defaultOptions`, spelled out because the zero value of
// the struct is not it.
func DefaultNoInvalidThisSettings() noInvalidThisSettings {
	return noInvalidThisSettings{CapIsConstructor: true}
}

// noInvalidThisRawOptions is the wire shape. The pointer distinguishes an absent key from
// `capIsConstructor: false`, which mean opposite things for a default-true option.
type noInvalidThisRawOptions struct {
	CapIsConstructor *bool `json:"capIsConstructor"`
}

// DecodeNoInvalidThisOptions maps the wire key onto the setting the rule reads.
//
// Hand written rather than `rule.DecodeOptionsInto` because the only default is true: a rule
// configured as a bare "error" is handed nil options, the generic decoder errors on empty input, and
// the resulting zero-value struct would say capIsConstructor is false. That is not a weaker
// configuration, it is a different rule, and it reports every capitalized constructor function in
// the tree.
func DecodeNoInvalidThisOptions(raw []byte) (any, error) {
	settings := DefaultNoInvalidThisSettings()

	decoded, err := rule.DecodeOptionsInto[noInvalidThisRawOptions]()(raw)
	if err != nil {
		return settings, err
	}
	wire, _ := decoded.(noInvalidThisRawOptions)

	if wire.CapIsConstructor != nil {
		settings.CapIsConstructor = *wire.CapIsConstructor
	}
	return settings, nil
}

// noInvalidThisSettingsFrom recovers the settings from whatever the config layer handed over,
// falling back to the documented defaults for nil.
func noInvalidThisSettingsFrom(options any) noInvalidThisSettings {
	if settings, ok := rule.OptionsAs[noInvalidThisSettings](options); ok {
		return settings
	}
	return DefaultNoInvalidThisSettings()
}

// noInvalidThisWalker carries the validity stack across the recursive walk.
type noInvalidThisWalker struct {
	ctx      rule.Context
	settings noInvalidThisSettings

	// thisIsValid mirrors typescript-eslint's `thisIsValidStack`. Its top is the binding in force
	// for the `this` being visited, and an empty stack means top level, where a module's `this` is
	// `undefined` and therefore invalid.
	thisIsValid []bool
}

func (w *noInvalidThisWalker) walk(node *ast.Node) {
	if node == nil {
		return
	}

	pushed := false
	switch node.Kind {
	case ast.KindThisKeyword:
		if !w.currentlyValid() {
			w.ctx.ReportNode(node, messageUnexpectedThis)
		}

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		// The TypeScript contribution: a declared `this` parameter binds `this` by contract, and
		// the checker enforces it at every call site.
		w.thisIsValid = append(w.thisIsValid,
			declaresThisParameter(node) || !thisIsDefaultBoundIn(w.ctx, node, w.settings))
		pushed = true

	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
		// A method's `this` is the receiver and a static block's is the class. Upstream reaches
		// these through the same default-binding test, which answers false for all of them.
		w.thisIsValid = append(w.thisIsValid, true)
		pushed = true

	case ast.KindPropertyDeclaration:
		// A field initializer is an implicit function whose `this` is the instance. Upstream pushes
		// true for PropertyDefinition and AccessorProperty alike; `accessor` is a modifier here
		// rather than a separate node, so both arrive as this kind.
		w.thisIsValid = append(w.thisIsValid, true)
		pushed = true

	case ast.KindArrowFunction:
		// Deliberately no push. An arrow has no `this` of its own, so the enclosing binding stays
		// in force and a `this` inside it is judged exactly as one outside it would be.
	}

	node.ForEachChild(func(child *ast.Node) bool {
		w.walk(child)
		return false
	})

	if pushed {
		w.thisIsValid = w.thisIsValid[:len(w.thisIsValid)-1]
	}
}

// currentlyValid reads the top of the stack.
//
// An empty stack is top level. Every file cohere lints is a TypeScript module, where top-level
// `this` is `undefined`, so the answer there is false. Core reaches the same answer through
// `node.sourceType === "module"`.
func (w *noInvalidThisWalker) currentlyValid() bool {
	if len(w.thisIsValid) == 0 {
		return false
	}
	return w.thisIsValid[len(w.thisIsValid)-1]
}

// declaresThisParameter answers whether a function declares an explicit `this` parameter.
//
// TypeScript spells the receiver as a first parameter literally named `this`, which the checker then
// enforces at every call site, so a `this` inside the body is bound by contract. Upstream tests
// `param.type === Identifier && param.name === 'this'` over every parameter rather than only the
// first, and that width is reproduced here: the grammar allows `this` only in first position, but the
// parser recovers from `function f(x, this: T)` and hands back the illegal shape, which the corpus
// exercises directly with `z(function (x, this: context) { console.log(x, this); });` as a PASSING
// case.
func declaresThisParameter(node *ast.Node) bool {
	parameters := node.Parameters()
	for _, parameter := range parameters {
		name := parameter.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "this" {
			return true
		}
	}
	return false
}

// thisIsDefaultBoundIn is ESLint's `astUtils.isDefaultThisBinding`, which is where the judgment lives.
//
// It answers "would calling this function leave `this` as the default binding", and the rule reports
// exactly when it says yes. The traversal walks OUTWARD from the function looking for something that
// binds a receiver, and every arm below is upstream's, in upstream's order.
//
// The direction is worth stating because it inverts twice on the way to the report site: this returns
// true when `this` is UNBOUND, the stack stores whether `this` is VALID, and the report fires when
// the stack says invalid.
func thisIsDefaultBoundIn(ctx rule.Context, node *ast.Node, settings noInvalidThisSettings) bool {
	// A capitalized name is the convention for a constructor, and upstream trusts it. This is the
	// single option surface, and it reads the function's own name.
	if settings.CapIsConstructor && startsWithUpperCase(functionOwnName(node)) {
		return false
	}
	if hasJSDocThisTag(ctx, node) {
		return false
	}

	isAnonymous := functionOwnName(node) == ""
	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			return true
		}

		switch parent.Kind {
		// typescript-go keeps parentheses as a node and the parser upstream ports from does not,
		// so every arm below would otherwise be unreachable through a parenthesized form. This is
		// a parser difference rather than a judgment, and it is load-bearing on five corpus cases:
		// `obj.foo = (function () { return function () { this.x; }; })();` and
		// `(function () { this.x; }).call(obj);` both put a parenthesis between the function and
		// the node that binds its receiver, and without this skip both reach the default arm and
		// report. Upstream sees no such node and needs no such arm.
		case ast.KindParenthesizedExpression:
			current = parent
			continue

		// A default parameter value is upstream's AssignmentPattern, whose arm is shared with
		// AssignmentExpression: a capitalized target binds by naming convention and anything else
		// does not. typescript-go models the same source as a Parameter carrying an Initializer, so
		// the shape is reached by a different kind and answered the same way. The corpus case is
		// `function foo(Ctor = function () { this.x; }) {}`, which passes because Ctor is
		// capitalized, and its lowercase twin four lines below it reports.
		case ast.KindParameter:
			if parent.Initializer() != current {
				return true
			}
			name := parent.Name()
			return !(settings.CapIsConstructor && isAnonymous &&
				name != nil && name.Kind == ast.KindIdentifier && startsWithUpperCase(name.Text()))

		// Look through to the destination: `obj.foo = nativeFoo || function foo() {};`
		case ast.KindBinaryExpression:
			if isLogicalOperator(parent) {
				current = parent
				continue
			}
			return assignmentBindsThis(parent, current, isAnonymous, settings)

		case ast.KindConditionalExpression:
			current = parent
			continue

		// If the enclosing function is immediately invoked, follow its return value out.
		//
		// The step out is to the CALL, not to the function's own parent, and the two differ here
		// where they do not upstream: `(function () { return function () {}; })()` gives the inner
		// function a parenthesized parent, so `enclosing.Parent` is the parenthesis rather than the
		// call. Upstream's `func.parent` IS the call because its parser drops the node. Stepping to
		// the parenthesis instead leaves the loop reading the call as an ordinary CallExpression
		// arm, which answers "nothing binds this" and reports, on three corpus cases that pass.
		case ast.KindReturnStatement:
			enclosing := enclosingFunctionOf(parent)
			if enclosing == nil {
				return true
			}
			call := callInvoking(enclosing)
			if call == nil {
				return true
			}
			current = call
			continue

		case ast.KindArrowFunction:
			if current != parent.Body() {
				return true
			}
			call := callInvoking(parent)
			if call == nil {
				return true
			}
			current = call
			continue

		// A method or a property's value has a receiver; anything else in these nodes does not.
		//
		// The Initializer test is upstream's `parent.value !== currentNode` and is INERT for both
		// kinds here, kept for fidelity. A computed key is a ComputedPropertyName node in
		// typescript-go rather than a direct child, for an object literal and a class field alike,
		// so a function whose parent is either of these kinds is always its value. Measured twice:
		// a 14-shape probe over object literals found no counterexample, and the class field shape
		// `class A { [function () { this.x; }]: number = 1; }` parses as
		// ComputedPropertyName < PropertyDeclaration, so it reaches the default arm and reports
		// without consulting this line. An earlier draft of this comment claimed the class field
		// case DID reach here and that claim was wrong; the mutant survived the fixture written on
		// the strength of it, which is how the error was caught.
		case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
			return parent.Initializer() != current
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			return false

		// `var Foo = function () {};` binds by naming convention only.
		case ast.KindVariableDeclaration:
			return !(settings.CapIsConstructor && isAnonymous &&
				parent.Initializer() == current &&
				parent.Name() != nil && parent.Name().Kind == ast.KindIdentifier &&
				startsWithUpperCase(parent.Name().Text()))

		// `.bind(obj)`, `.call(obj)`, `.apply(obj, [])` all supply a receiver.
		//
		// The `Expression() == current` half is upstream's and is INERT here, kept for fidelity
		// rather than because it decides anything. Upstream needs it because its MemberExpression
		// carries both the object and the property as expression children; typescript-go puts the
		// property in its own `name` slot, so a function reaching this arm is always the receiver.
		// Measured rather than argued: a probe walked 14 shapes including computed access,
		// string and numeric keys, spreads, accessors and `q['bind'](obj)`, asserting the function
		// is always the Expression, and found no counterexample. The mutant dropping this test
		// therefore survives the sweep as EQUIVALENT rather than as a fixture gap.
		case ast.KindPropertyAccessExpression:
			if parent.Expression() == current && isBindCallOrApply(parent) {
				return !(isCallee(parent) && callHasNonNullishArgument(parent.Parent, 0))
			}
			return true

		// `Reflect.apply(fn, obj, [])`, `Array.from([], fn, obj)`, `list.forEach(fn, obj)`.
		case ast.KindCallExpression:
			return callDoesNotBindThis(parent, current)

		default:
			return true
		}
	}
	return true
}

// functionOwnName returns the name a function carries itself, or empty for an anonymous one.
//
// Upstream reads `node.id`, which exists only for a declaration or a named function expression. A
// name inherited from a variable or a property is deliberately NOT this: upstream handles that
// separately in the VariableDeclaration and AssignmentExpression arms, and only for anonymous
// functions, so folding it in here would report differently on
// `var Foo = function bar() { this.x; };`.
func functionOwnName(node *ast.Node) string {
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

// startsWithUpperCase is upstream's constructor-naming convention test.
//
// Upstream writes `name[0] !== name[0].toLocaleLowerCase()`, which is a "has an uppercase form that
// differs" test rather than an "is uppercase" test, and the two disagree on characters with no case
// at all. Reproduced with unicode.ToLower rather than unicode.IsUpper for that reason: `_foo` and
// `1foo` are lowercase-invariant and answer false either way, but a caseless letter answers false
// here and true under IsUpper. The shelf's own react.IsLikelyComponentName uses IsUpper and is
// therefore the wrong helper for this rule, which is why this is written out.
func startsWithUpperCase(name string) bool {
	if name == "" {
		return false
	}
	first := []rune(name)[0]
	return unicode.ToLower(first) != first
}

// hasJSDocThisTag answers whether a `@this` tag documents this function.
//
// Upstream asks two questions and reports if either says yes, and both are reproduced because both
// were measured to matter. This is the largest single piece of machinery behind a rule whose own
// file is a hundred lines, and it is a good example of the rule file being a thin caller.
//
// The first question is `getJSDocComment`, which walks OUTWARD from the function to whatever node
// the documentation would sit on, and then requires the token immediately before that node to be a
// real JSDoc block. The walk is why the boundary is not a token gap, measured at 8.67.0:
//
//	function foo() { /** @this*/ return function () { this.x; }; }        silent
//	function foo() { /** @this*/ var q = function () { this.x; }; }       silent
//	function foo() { /** @this*/ if (a) { return function () {...}; } }   silent
//	function foo() { /** @this*/ return (function () { this.x; }); }      silent
//	function foo() { /** @this*/ q(function () { this.x; }); }            REPORTS
//	function foo() { /** @this*/ const q = 1; return fn; }                REPORTS
//
// The fifth is the one that pins the shape: a function passed as a CALL ARGUMENT does not walk out
// at all, because upstream excludes a call or new parent before the loop starts. The sixth closes
// the walk for a different reason: the ancestor it reaches is the return statement, and the token
// before that is `;` rather than a comment.
//
// The second question is `getCommentsBefore`, upstream's stated fallback for callbacks, which reads
// every comment token in the gap before the function itself, block or line, and asks the same
// pattern of each:
//
//	z(/* @this Obj */ function () { this.x; });                  silent, the callback case
//	z(/* @this Obj */ 1, function () { this.x; });               REPORTS, the gap belongs to `1`
//	/* @this Obj */ /* second */ function bar() {...}            silent, every comment is read
//	/* second */ /* @this Obj */ function bar() {...}            silent, likewise reversed
//	// @this Obj                                                 silent, a line comment counts here
//
// The pattern itself is `/^[\s*]*@this/mu`: anchored per line, allowing asterisks in the prefix, and
// with NO closing word boundary. So `@thisx` satisfies it and `@nothis` does not, both measured.
// That reads as a defect and is reproduced rather than corrected, because fidelity is the authority
// where upstream gives no reasoning.
func hasJSDocThisTag(ctx rule.Context, node *ast.Node) bool {
	if ctx.SourceFile == nil {
		return false
	}
	if commentTextHasThisTag(jsDocCommentFor(ctx, node)) {
		return true
	}
	for _, comment := range commentsImmediatelyBefore(ctx, node) {
		if commentTextHasThisTag(comment.Text) {
			return true
		}
	}
	return false
}

// jsDocCommentFor is upstream's `getJSDocComment` for the function shapes this rule reaches, which
// are a declaration and an expression. It returns the documenting comment's text, or empty.
func jsDocCommentFor(ctx rule.Context, node *ast.Node) string {
	if node.Kind == ast.KindFunctionDeclaration {
		return jsDocBlockBefore(ctx, node)
	}

	parent := node.Parent
	if parent == nil {
		return jsDocBlockBefore(ctx, node)
	}
	// A function handed straight to a call does not walk out. This is what separates
	// `q(function () {...})` from `var q = function () {...}` and it is upstream's own guard.
	if parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindNewExpression {
		return jsDocBlockBefore(ctx, node)
	}

	for parent != nil {
		if len(commentsImmediatelyBefore(ctx, parent)) > 0 || isFunctionLike(parent) ||
			parent.Kind == ast.KindMethodDeclaration || parent.Kind == ast.KindPropertyAssignment {
			break
		}
		parent = parent.Parent
	}
	if parent != nil && parent.Kind != ast.KindFunctionDeclaration && parent.Kind != ast.KindSourceFile {
		return jsDocBlockBefore(ctx, parent)
	}
	return jsDocBlockBefore(ctx, node)
}

// jsDocBlockBefore is upstream's `findJSDocComment`: the token immediately before the node has to be
// a block comment whose text opens with an asterisk, and it has to end no more than one line above.
func jsDocBlockBefore(ctx rule.Context, node *ast.Node) string {
	preceding := commentsImmediatelyBefore(ctx, node)
	if len(preceding) == 0 {
		return ""
	}
	last := preceding[len(preceding)-1]
	if !last.IsBlock || !strings.HasPrefix(strings.TrimPrefix(last.Text, "/*"), "*") {
		return ""
	}
	nodeLine, _ := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile,
		rule.TokenRange(ctx.SourceFile, node).Pos())
	commentLine, _ := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile, last.Range.End())
	if nodeLine-commentLine > 1 {
		return ""
	}
	return last.Text
}

// commentsImmediatelyBefore is upstream's `getCommentsBefore`: every comment sitting in the trivia
// between the previous real token and this node, in source order.
//
// Other comments in the gap do not close it, which is why the span is scanned for comments rather
// than merely tested for whitespace. Any non-comment text does close it, and that is what makes the
// gap belong to `1` rather than to the callback in `z(/* @this */ 1, function () {})`.
func commentsImmediatelyBefore(ctx rule.Context, node *ast.Node) []comments.Comment {
	tokenStart := rule.TokenRange(ctx.SourceFile, node).Pos()
	text := ctx.SourceFile.Text()
	if tokenStart > len(text) {
		return nil
	}

	found := []comments.Comment{}
	cursor := tokenStart
	all := comments.ForFile(ctx)
	for index := len(all) - 1; index >= 0; index-- {
		comment := all[index]
		if comment.Range.End() > cursor {
			continue
		}
		if strings.TrimSpace(text[comment.Range.End():cursor]) != "" {
			break
		}
		found = append([]comments.Comment{comment}, found...)
		cursor = comment.Range.Pos()
	}
	return found
}

// isFunctionLike answers whether a node introduces its own `this` or parameter scope, which is where
// upstream's outward walk stops.
func isFunctionLike(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
}

// commentTextHasThisTag tests upstream's `/^[\s*]*@this/mu` against a comment's text.
//
// Line by line rather than with a regular expression, because the pattern is anchored per line and
// asks one question of each: after any run of whitespace and asterisks, does `@this` begin. The
// delimiters are stripped first, since upstream matches against `comment.value` rather than the raw
// source of the comment.
func commentTextHasThisTag(text string) bool {
	if text == "" {
		return false
	}
	value := strings.TrimPrefix(text, "/*")
	value = strings.TrimSuffix(value, "*/")
	value = strings.TrimPrefix(value, "//")

	for _, line := range strings.Split(value, "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " \t\r*"), "@this") {
			return true
		}
	}
	return false
}

// isLogicalOperator answers whether a binary expression is `||`, `&&` or `??`.
//
// These are upstream's LogicalExpression, which it looks THROUGH rather than deciding on, because
// `obj.foo = cached || function () {}` still assigns the function to `obj.foo`. Every other binary
// operator, assignment included, is a decision point.
func isLogicalOperator(node *ast.Node) bool {
	switch node.AsBinaryExpression().OperatorToken.Kind {
	case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
		return true
	}
	return false
}

// assignmentBindsThis handles upstream's AssignmentExpression and AssignmentPattern arms.
//
// A member-expression target binds a receiver outright, and a capitalized identifier target binds one
// by naming convention for an anonymous function only. Anything else leaves `this` default.
func assignmentBindsThis(parent *ast.Node, current *ast.Node, isAnonymous bool, settings noInvalidThisSettings) bool {
	binary := parent.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Right != current {
		return true
	}
	left := binary.Left
	if left.Kind == ast.KindPropertyAccessExpression || left.Kind == ast.KindElementAccessExpression {
		return false
	}
	if settings.CapIsConstructor && isAnonymous && left.Kind == ast.KindIdentifier &&
		startsWithUpperCase(left.Text()) {
		return false
	}
	return true
}

// enclosingFunctionOf walks up to the function a node sits inside, which is upstream's
// `getUpperFunction`.
func enclosingFunctionOf(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
			return current
		}
	}
	return nil
}

// callInvoking returns the call expression that immediately invokes a node, or nil.
//
// The parenthesis walk is the same one isCallee does, but the CALL is what the caller needs rather
// than a yes or no, because stepping out of an immediately invoked function has to land on the call
// and typescript-go puts a parenthesis in between. Sharing the walk keeps the two answers from
// drifting: a shape isCallee accepts is a shape this returns a call for.
func callInvoking(node *ast.Node) *ast.Node {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return nil
	}
	if parent.AsCallExpression().Expression != current {
		return nil
	}
	return parent
}

// isCallee answers whether a node is the thing being called in a call expression.
//
// The parenthesis skip is deliberate and measured. `(function () { this.x; })()` is an immediately
// invoked expression whose callee is parenthesized, and typescript-go keeps the parentheses as a
// node where the parser upstream ports from does not, so without the skip every parenthesized
// invocation reads as a non-call and the arm above returns early.
func isCallee(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	return parent.AsCallExpression().Expression == current
}

// isBindCallOrApply answers whether a property access names `bind`, `call` or `apply`.
func isBindCallOrApply(node *ast.Node) bool {
	name := node.AsPropertyAccessExpression().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	switch name.Text() {
	case "bind", "call", "apply":
		return true
	}
	return false
}

// callHasNonNullishArgument answers whether a call supplies a real value in the given position.
//
// `null` and `undefined` do not bind a receiver, which is why upstream tests the argument's value
// rather than merely its presence: `foo.bind(null)` leaves `this` default and reports.
func callHasNonNullishArgument(call *ast.Node, index int) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) <= index {
		return false
	}
	return !isNullOrUndefinedLiteral(arguments.Nodes[index])
}

// isNullOrUndefinedLiteral is upstream's `isNullOrUndefined`, which accepts the `null` literal, the
// `undefined` identifier, and `void <anything>`.
func isNullOrUndefinedLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return node.Text() == "undefined"
	case ast.KindVoidExpression:
		return true
	}
	return false
}

// callDoesNotBindThis is upstream's CallExpression arm, covering the three shapes that pass a
// receiver as an ordinary argument.
//
// Each is a different position, which is the reason this cannot be one test: Reflect.apply takes the
// receiver second of three, Array.from takes it third of three, and a thisArg method takes it second
// of two. Getting the arity wrong in either direction silently changes which calls bind.
func callDoesNotBindThis(call *ast.Node, current *ast.Node) bool {
	arguments := call.AsCallExpression().Arguments
	count := 0
	if arguments != nil {
		count = len(arguments.Nodes)
	}
	callee := call.AsCallExpression().Expression

	if isSpecificMemberAccess(callee, "Reflect", "apply") {
		return count != 3 || arguments.Nodes[0] != current ||
			isNullOrUndefinedLiteral(arguments.Nodes[1])
	}
	if isSpecificMemberAccess(callee, "Array", "from") ||
		isSpecificMemberAccess(callee, "Array", "fromAsync") {
		return count != 3 || arguments.Nodes[1] != current ||
			isNullOrUndefinedLiteral(arguments.Nodes[2])
	}
	if isThisArgMethodAccess(callee) {
		return count != 2 || arguments.Nodes[0] != current ||
			isNullOrUndefinedLiteral(arguments.Nodes[1])
	}
	return true
}

// isSpecificMemberAccess answers whether a callee is exactly `object.property`.
func isSpecificMemberAccess(callee *ast.Node, object string, property string) bool {
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.Expression == nil || access.Expression.Kind != ast.KindIdentifier ||
		access.Expression.Text() != object {
		return false
	}
	name := access.Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == property
}

// thisArgMethodNames is upstream's `arrayOrTypedArrayPattern` method list: the Array methods whose
// second argument is a thisArg.
//
// A plain lookup rather than upstream's regular expression over the receiver, because upstream's
// pattern also constrains the RECEIVER to an array-looking name and this does not need to: the
// method name alone is what its `isMethodWhichHasThisArg` tests, walking down a member chain to the
// last property and matching the name against this set.
var thisArgMethodNames = map[string]struct{}{
	"every": {}, "filter": {}, "find": {}, "findIndex": {}, "findLast": {},
	"findLastIndex": {}, "flatMap": {}, "forEach": {}, "map": {}, "some": {},
}

// isThisArgMethodAccess answers whether a callee names one of the Array methods that take a thisArg.
func isThisArgMethodAccess(callee *ast.Node) bool {
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	_, ok := thisArgMethodNames[name.Text()]
	return ok
}
