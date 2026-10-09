package typescript

import (
	"encoding/json"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// OnlyThrowErrorOptions is upstream's option struct, field-for-field, plus the one split our wire
// format forces.
//
// `allow` arrives as ONE heterogeneous array whose entries are either a bare string (matched
// against the type's own name) or a `TypeOrValueSpecifier` object. `Allow` and `AllowInline` are
// the two typed halves of that array, exactly as `no-floating-promises` splits its two allowlists.
//
// The three booleans stay POINTERS all the way through, because every one of them defaults to
// TRUE. Binding them to plain `bool` would make an unconfigured rule behave as though all three had
// been explicitly turned off, which is the loudest possible failure: `throw someAny` would start
// reporting on every file in the tree. Absent has to stay distinguishable from false.
type OnlyThrowErrorOptions struct {
	Allow                []type_checking.TypeOrValueSpecifier
	AllowInline          []string
	AllowRethrowing      *bool
	AllowThrowingAny     *bool
	AllowThrowingUnknown *bool
}

func buildOnlyThrowErrorObjectMessage() rule.Message {
	return rule.Message{
		Id: "object",
		Description: "This throws a value that is not an Error, so whatever catches it gets no stack " +
			"trace and no message. The failure then surfaces as a bare string or object with no " +
			"record of where it came from, which is the difference between a five-minute fix and " +
			"an afternoon. Throw `new Error(...)` instead.",
	}
}

func buildOnlyThrowErrorUndefMessage() rule.Message {
	return rule.Message{
		Id: "undef",
		Description: "This throws `undefined`, which reaches the handler as a value carrying nothing " +
			"at all: no message, no stack, no type. It is almost always an expression that was " +
			"meant to produce an error and did not. Throw `new Error(...)` instead.",
	}
}

// OnlyThrowError flags a `throw` whose argument is not an Error.
//
//	valid:   throw new Error();
//	valid:   class CustomError extends Error {} throw new CustomError();
//	valid:   try {} catch (e) { throw e; }
//	valid:   function fun(value: any) { throw value; }
//	invalid: throw 'error';
//	invalid: throw undefined;
//	invalid: class CustomError {} throw new CustomError();
//	invalid: throw foo ? new Error() : 'literal';
//
// Only an Error carries a stack trace. Throwing a string, an object literal, or `undefined` gives
// the handler a value with no record of where the failure happened, and every logger and reporter
// downstream then prints something that describes nothing. The rule is about what the catcher
// receives, not about style.
//
// # What decides a finding
//
// One listener on `KindThrowStatement`, reading the argument's TYPE. The order below is upstream's
// and it is load-bearing, because two arms can both match the same input and the first one wins:
//
//	rethrow          an identifier bound by a catch clause, or by the rejection handler of a
//	                 `.catch`/`.then` on something thenable. Silent, and checked FIRST, so a
//	                 rethrown string is allowed while a bare string is not.
//	allow            the configured specifiers, matched against the type
//	undefined        reported as `undef` rather than `object`, and reported BEFORE the any/unknown
//	                 escapes, so `throw undefined` reports even with everything else permitted
//	any              silent unless `allowThrowingAny` is false
//	unknown          silent unless `allowThrowingUnknown` is false
//	Error-like       silent
//	everything else  reported as `object`
//
// The `undefined` arm sitting above the `any` and `unknown` arms is the one piece of the ordering
// that is not obvious, and upstream's own corpus pins it: `throw undefined` is invalid case 0 with
// no options at all, so it reports while every default is still permissive.
//
// The type test is `TypeFlagsUndefined`, which is a FLAG test rather than an equality, so a union
// containing `undefined` takes this arm. Measured against the installed rule:
// `declare const x: Error | undefined; throw x;` reports `object`, not `undef`, because the union
// as a whole does not carry the undefined flag at the top level -- `Error | undefined` is a union
// type whose own flags are `Union`. A bare `undefined` does.
//
// # Parentheses, and the one place our parser differs from ESTree
//
// ESTree does not represent parentheses as nodes at all, so upstream's rule is handed a bare
// `Identifier` for `throw (e)` and its `node.type !== Identifier` guard never sees the wrapper.
// Our parser DOES produce `KindParenthesizedExpression`, so a literal port of that guard would
// decline every parenthesized rethrow and report it as a violation.
//
// This is a parser difference rather than a rule difference, so the port skips parentheses in the
// rethrow check to reproduce the DECISION. Measured on the installed rule, holding source fixed and
// varying only `allowRethrowing`:
//
//	try {} catch (e) { throw (e); }    allowRethrowing default   SILENT
//	try {} catch (e) { throw (e); }    allowRethrowing: false    REPORTS object
//
// The second row is what proves the rethrow arm is actually firing on the parenthesized form rather
// than the case being silent for some other reason. Both are fixtured.
//
// The SPAN excludes the parentheses too, and that is measured rather than inherited: upstream
// reports `throw ('error')` at columns 8..15, which is `'error'` without its wrapper, because
// ESTree hands it the inner node. `ctx.ReportNode` on the skipped node reproduces it.
//
// # The rethrow arm, and what replaces the scope manager
//
// Upstream reaches for `context.sourceCode.getScope(node)` and `findVariable`, then reads the
// resolved variable's `defs`. We have no scope manager, and we do not need one: the checker answers
// the same question. `GetSymbolAtLocation` on the identifier returns the symbol, and its
// declarations carry the shape upstream reads off the definition. Probed on this tree:
//
//	try {} catch (e) { throw e; }            decl KindVariableDeclaration, parent KindCatchClause
//	p.catch(e => { throw e; })               decl KindParameter,           parent KindArrowFunction
//	let x = 1; p.catch(e => { throw x; })    decl KindVariableDeclaration, parent KindVariableDeclarationList
//
// So the catch arm is "one declaration, and its parent is a catch clause", and the promise arm is
// "one declaration, it is the FIRST parameter of an arrow function, and that arrow is the rejection
// handler of a `.catch`/`.then` on a thenable".
//
// Upstream filters `defs` to those where `isVariableDefinition` is true and requires exactly one.
// In its scope manager that flag is false for exactly one definition kind, `TypeDefinition`, so the
// filter means "exactly one non-type definition". A symbol's `Declarations` here is the merged
// list, and the equivalent narrowing is to require a single declaration: a name that is both a type
// and a value has two, and upstream declines it.
//
// # What the promise arm requires, all four conditions measured
//
// Every one of these was measured against the installed rule rather than read off the source,
// because the source states them as a conjunction and reading a conjunction tells you nothing about
// which conjunct is doing the work:
//
//	the handler must be an ARROW function      `.catch(function (e) { throw e; })`  REPORTS
//	the binding must be the FIRST parameter    `.catch((a, e) => { throw e; })`     REPORTS
//	it must be a plain identifier parameter    `.catch(({ e }) => { throw e; })`    REPORTS
//	the receiver must be THENABLE              a `{ catch(cb) {} }` object          REPORTS
//
// The rest parameter case is the same first-parameter rule seen from the other side:
// `.catch((...e) => { throw e; })` binds `e` to the whole argument list rather than to the
// rejection, so upstream reports it, and its corpus pins that as invalid case 40.
//
// The method NAME is read statically, and it resolves through a computed access and even through a
// const-folded variable key. Measured, with `allowRethrowing: false` as the control on each:
//
//	Promise.reject('x')['catch'](e => { throw e; })          SILENT (reports with rethrow off)
//	const k = 'catch'; Promise.reject('x')[k](e => ...)      SILENT (reports with rethrow off)
//	p?.catch(e => { throw e; })                              SILENT
//
// The const-folded key is the one this port does NOT reproduce, and it is stated rather than
// silent. Upstream's `getStaticMemberAccessValue` falls back to `getStaticValue`, an evaluator that
// resolves an identifier through the scope manager to a constant initializer. Reproducing it means
// a constant-folding pass, which is a substantially larger piece of machinery than the rest of this
// rule, and it costs exactly one thing: `obj[k](e => { throw e; })` where `k` is a `const` bound to
// `'catch'` or `'then'` reports here and is silent upstream. No case in the corpus writes that
// shape. The literal computed form `['catch']` IS reproduced, since that needs no evaluator.
//
// # Two arms of the `then` shape, and why the second argument matters
//
// `.then(onFulfilled, onRejected)` puts the rejection handler SECOND, so the arrow must be
// `arguments[1]`; `.then(onRejected)` has no rejection handler at all, and upstream reports it.
// That is invalid case 43 in the corpus. A spread anywhere before the handler makes the position
// unknowable, so upstream returns the object without an `onRejected` and the arm declines: invalid
// cases 41, 42 and 43 are exactly those three spread shapes.
//
// # `allow`, and the two specifier forms
//
// `allow` is checked against the type before any of the built-in escapes, so a configured specifier
// wins over the `undef` arm: upstream's valid case 32 allows `{from: 'lib', name: 'undefined'}` and
// `throw undefined` goes clean. Fixtured.
//
// # Cost
//
// `KindThrowStatement` is a rare anchor. The checker is consulted once per throw, and the rethrow
// arm short-circuits on anything that is not a bare identifier, so the thenable test runs only for
// a `throw e` inside a callback.
var OnlyThrowError = rule.Rule{
	Name: "@typescript-eslint/only-throw-error",

	// Every arm reads the argument's type, so the checker is required.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[OnlyThrowErrorOptions](options)

		// A rule configured as bare `"error"` is handed nil options, and `options.(T)` on nil yields
		// the zero value, which is three nil pointers. Every default here is TRUE, so defaulting
		// them is what keeps an unconfigured rule permissive rather than reporting every `any`.
		allowRethrowing := settings.AllowRethrowing == nil || *settings.AllowRethrowing
		allowThrowingAny := settings.AllowThrowingAny == nil || *settings.AllowThrowingAny
		allowThrowingUnknown := settings.AllowThrowingUnknown == nil || *settings.AllowThrowingUnknown

		return rule.Listeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				// Three nil checks in this listener are DEFENSIVE rather than behavioral, and the
				// mutation sweep says so: neutralizing any of them changes no verdict, and a
				// reachability probe over ten malformed shapes -- `throw;`, `throw ();`,
				// `throw (());`, a bare `throw` at the end of a function body, in a switch case,
				// after an `if`, in a `do` body, plus three well-formed controls -- never produced
				// a nil for any of them. Our parser recovers by SYNTHESIZING a zero-width
				// `KindIdentifier` rather than by leaving the field empty, which is why. Kept
				// because a panic in a linter takes down the whole run, and because the verdict
				// names the shapes it was taken over: a parser change that starts leaving these
				// empty voids it.
				//
				// The FOURTH nil check, on the resolved symbol below, is a different matter. It is
				// reachable, it was the one survivor of these four that a real input reached, and
				// upstream CRASHES on that input. See the test.
				argument := node.AsThrowStatement().Expression
				if argument == nil {
					return
				}

				// ESTree has no parenthesis node, so upstream's rule never sees one. Skipping here
				// is what makes the DECISION and the SPAN match; see the doc comment above.
				subject := ast.SkipParentheses(argument)
				if subject == nil {
					return
				}

				// A SYNTHESIZED node has no text to point at, so reporting on it emits a
				// zero-width finding at a position the author never wrote.
				//
				// Upstream cannot reach this at all: `throw;` and `throw ();` are both
				// "Expression expected" PARSE ERRORS for its parser, so the rule never runs and
				// there is no upstream verdict to be faithful to. Our parser RECOVERS, handing
				// back a zero-width `KindIdentifier` whose type is `any`, which is silent only
				// while `allowThrowingAny` is true. With the option off it reports on nothing.
				//
				// The test is AFTER the parenthesis skip rather than before it, and that is the
				// whole point: `throw;` puts the synthesized node directly under the statement, so
				// a check on the raw argument catches it, while `throw ();` wraps it in a real
				// `KindParenthesizedExpression` that is not itself missing. Measured with
				// `allowThrowingAny: false`, checking the raw argument left `throw ();` reporting
				// a zero-width `object` at offset 7 while `throw;` went clean, which is the
				// asymmetry a reader would never predict from the grammar.
				if ast.NodeIsMissing(subject) {
					return
				}

				if allowRethrowing && onlyThrowErrorIsRethrown(ctx, subject) {
					return
				}

				argumentType := ctx.TypeChecker.GetTypeAtLocation(subject)
				if argumentType == nil {
					return
				}

				if type_checking.TypeMatchesSomeSpecifier(argumentType, settings.Allow, settings.AllowInline, ctx.Program) {
					return
				}

				// A FLAG test rather than an equality, matching upstream's `isTypeFlagSet`, and it
				// sits above the any/unknown escapes on purpose: `throw undefined` reports even
				// with every default still permissive.
				if type_checking.IsTypeFlagSet(argumentType, checker.TypeFlagsUndefined) {
					ctx.ReportNode(subject, buildOnlyThrowErrorUndefMessage())
					return
				}

				if allowThrowingAny && type_checking.IsTypeFlagSet(argumentType, checker.TypeFlagsAny) {
					return
				}

				if allowThrowingUnknown && type_checking.IsTypeFlagSet(argumentType, checker.TypeFlagsUnknown) {
					return
				}

				if type_checking.IsErrorLike(ctx.Program, ctx.TypeChecker, argumentType) {
					return
				}

				ctx.ReportNode(subject, buildOnlyThrowErrorObjectMessage())
			},
		}
	},
}

// onlyThrowErrorIsRethrown answers whether this identifier is a caught value being thrown again.
//
// Two shapes count, and they are the two upstream names: a catch clause binding, and the rejection
// handler parameter of a `.catch`/`.then` on something thenable. Anything that is not a bare
// identifier is not a rethrow, which is what keeps `throw 'literal'` reportable.
func onlyThrowErrorIsRethrown(ctx rule.Context, node *ast.Node) bool {
	if !ast.IsIdentifier(node) {
		return false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}

	// Upstream requires exactly one non-type definition. A merged symbol carries more than one
	// declaration, and upstream declines those, so the single-declaration test is the equivalent
	// narrowing rather than a simplification of it.
	if len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration == nil || declaration.Parent == nil {
		return false
	}

	// try { ... } catch (x) { throw x; }
	//
	// A destructured catch parameter binds a NEW name taken out of the exception rather than the
	// exception itself, and its declaration's parent is the binding pattern rather than the clause,
	// so it falls out here without a special case. Measured: upstream reports it too.
	if declaration.Parent.Kind == ast.KindCatchClause {
		return true
	}

	// promise.catch(x => { throw x; })
	// promise.then(onFulfilled, x => { throw x; })
	if declaration.Kind != ast.KindParameter || !ast.IsArrowFunction(declaration.Parent) {
		return false
	}

	arrow := declaration.Parent
	parameters := arrow.Parameters()
	if len(parameters) == 0 || parameters[0] != declaration {
		return false
	}

	// A rest parameter binds the whole argument list rather than the rejection value, so the name
	// is not the caught error even though it is first. Upstream's corpus pins this as reporting.
	if declaration.AsParameterDeclaration().DotDotDotToken != nil {
		return false
	}

	call := arrow.Parent
	if call == nil || !ast.IsCallExpression(call) {
		return false
	}

	receiver, onRejected := onlyThrowErrorParsePromiseHandlingCall(call)
	if receiver == nil || onRejected != arrow {
		return false
	}

	// The name alone is not enough: any object can have a method called `catch`. Upstream tests the
	// receiver's type, and a plain `{ catch(cb) {} }` object reports.
	return type_checking.IsThenableType(ctx.TypeChecker, receiver, nil)
}

// onlyThrowErrorParsePromiseHandlingCall reads a syntactically possible `.catch`/`.then` call,
// returning the receiver and the rejection handler. It does not check the receiver's type; the
// caller does that.
//
// Upstream keeps `parseThenCall` and `parseCatchCall` separate because other rules use each one on
// its own. Here there is one caller wanting one answer, so they are one function, and the shape
// each returns is the pair this rule reads. The spread handling is upstream's: a spread anywhere at
// or before the handler's position makes the position unknowable, so the call is parsed as having
// no rejection handler rather than being guessed at.
func onlyThrowErrorParsePromiseHandlingCall(call *ast.Node) (receiver *ast.Node, onRejected *ast.Node) {
	callee := ast.SkipParentheses(call.Expression())
	if callee == nil {
		return nil, nil
	}

	var methodName string
	switch {
	case ast.IsPropertyAccessExpression(callee):
		// A PRIVATE name counts. Upstream's static member reader accepts `Identifier` and
		// `PrivateIdentifier` alike, so a thenable class with a `#catch` method is a catch call as
		// far as this rule is concerned.
		//
		// That is not obvious and it is not what this port first wrote. Restricting to
		// `KindIdentifier` reads as the careful choice and is a DIVERGENCE. Measured on the
		// installed rule, on a class that is itself thenable so the receiver test cannot be what
		// decides it:
		//
		//	class C {
		//	  then(cb: (v: number) => void, r?: (e: any) => void): C { return this; }
		//	  #catch(cb: (e: any) => void) {}
		//	  m() { this.#catch(e => { throw e; }); }
		//	}
		//
		// Upstream is SILENT on that and the identifier-only version reports. Found by a mutation
		// that survived: dropping the kind guard changed nothing on any fixture, which sent this
		// back to upstream rather than to a new fixture, and upstream disagreed with the guard.
		//
		// The `#` has to come OFF, and that is the second half of the same finding. Our
		// `Node.Text()` on a private name returns `"#catch"`, while ESTree's `PrivateIdentifier`
		// carries `name: "catch"` with the hash stripped -- verified by reading the property off
		// the parser directly. Comparing our text against `"catch"` would therefore never match and
		// the arm would be inert while LOOKING correct, which is the doubled-hash failure this
		// project has already shipped once, arriving as a silent non-match rather than as a visibly
		// wrong message.
		name := callee.AsPropertyAccessExpression().Name()
		switch {
		case name == nil:
			return nil, nil
		case name.Kind == ast.KindIdentifier:
			methodName = name.Text()
		case name.Kind == ast.KindPrivateIdentifier:
			methodName = strings.TrimPrefix(name.Text(), "#")
		default:
			return nil, nil
		}
		receiver = callee.AsPropertyAccessExpression().Expression
	case ast.IsElementAccessExpression(callee):
		// `p['catch'](...)`. Only a literal key is read. Upstream additionally folds a `const`
		// bound to a string literal through its static evaluator; that is the one stated
		// divergence, and the doc comment on the rule names its cost.
		argumentExpression := ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		if argumentExpression == nil || !ast.IsStringLiteralLike(argumentExpression) {
			return nil, nil
		}
		methodName = argumentExpression.Text()
		receiver = callee.AsElementAccessExpression().Expression
	default:
		return nil, nil
	}

	if receiver == nil {
		return nil, nil
	}

	arguments := call.Arguments()

	switch methodName {
	case "catch":
		if len(arguments) == 0 {
			return receiver, nil
		}
		if arguments[0].Kind == ast.KindSpreadElement {
			return receiver, nil
		}
		return receiver, arguments[0]

	case "then":
		// The rejection handler is the SECOND argument. A one-argument `.then` has none at all, and
		// upstream reports a throw inside it.
		if len(arguments) < 2 {
			return receiver, nil
		}
		if arguments[0].Kind == ast.KindSpreadElement || arguments[1].Kind == ast.KindSpreadElement {
			return receiver, nil
		}
		return receiver, arguments[1]
	}

	return nil, nil
}

// onlyThrowErrorRawOptions is the wire shape, which is not the shape the rule reasons with.
//
// Three keys bind straight through and one does not. `allow` arrives as one heterogeneous array
// whose entries are either a bare string or a specifier object whose `from` is one of `file`,
// `lib`, `package`. In the struct the rule reads, the bare-string form lives in `AllowInline` and
// `from` is a `uint8` enum, so neither can be expressed as a struct tag.
type onlyThrowErrorRawOptions struct {
	Allow                []onlyThrowErrorRawSpecifier `json:"allow"`
	AllowRethrowing      *bool                        `json:"allowRethrowing"`
	AllowThrowingAny     *bool                        `json:"allowThrowingAny"`
	AllowThrowingUnknown *bool                        `json:"allowThrowingUnknown"`
}

// onlyThrowErrorRawSpecifier is one entry of the allowlist, in either of its two wire forms.
type onlyThrowErrorRawSpecifier struct {
	inline string

	From    string
	Name    []string
	Path    string
	Package string
}

// UnmarshalJSON accepts both wire forms: a bare string, or the object the schema describes.
//
// `name` is itself two shapes, a string or an array of strings. An unrecognized `from` is dropped
// rather than erroring, which matches how `no-floating-promises` treats the same field: the
// specifier then matches nothing instead of failing the whole run, and dropping it says nothing
// rather than silently meaning `file`, which is the zero value.
func (s *onlyThrowErrorRawSpecifier) UnmarshalJSON(raw []byte) error {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		*s = onlyThrowErrorRawSpecifier{inline: asString}
		return nil
	}

	var object struct {
		From    string          `json:"from"`
		Name    json.RawMessage `json:"name"`
		Path    string          `json:"path"`
		Package string          `json:"package"`
	}
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return err
	}

	*s = onlyThrowErrorRawSpecifier{From: object.From, Path: object.Path, Package: object.Package}

	var names []string
	if err := json.Unmarshal(object.Name, &names); err == nil {
		s.Name = names
		return nil
	}
	var name string
	if err := json.Unmarshal(object.Name, &name); err != nil {
		return err
	}
	s.Name = []string{name}
	return nil
}

// DecodeOnlyThrowErrorOptions maps upstream's JSON onto the struct the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` because `allow` arrives as one heterogeneous
// array and leaves as two typed fields, and because `from` is a string on the wire and an integer
// enum in the struct. The three booleans stay pointers all the way through so that "absent" stays
// distinguishable from "false"; every one of them defaults to TRUE, so collapsing that distinction
// would turn an unconfigured rule into a maximally strict one.
func DecodeOnlyThrowErrorOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[onlyThrowErrorRawOptions]()(raw)
	if err != nil {
		return OnlyThrowErrorOptions{}, err
	}

	wire, _ := decoded.(onlyThrowErrorRawOptions)

	options := OnlyThrowErrorOptions{
		AllowRethrowing:      wire.AllowRethrowing,
		AllowThrowingAny:     wire.AllowThrowingAny,
		AllowThrowingUnknown: wire.AllowThrowingUnknown,
	}
	for _, entry := range wire.Allow {
		if entry.inline != "" {
			options.AllowInline = append(options.AllowInline, entry.inline)
			continue
		}

		specifier := type_checking.TypeOrValueSpecifier{Name: entry.Name, Path: entry.Path, Package: entry.Package}
		switch entry.From {
		case "file":
			specifier.From = type_checking.TypeOrValueSpecifierFromFile
		case "lib":
			specifier.From = type_checking.TypeOrValueSpecifierFromLib
		case "package":
			specifier.From = type_checking.TypeOrValueSpecifierFromPackage
		default:
			continue
		}
		options.Allow = append(options.Allow, specifier)
	}

	return options, nil
}
