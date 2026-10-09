package typescript

import (
	"encoding/json"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferPromiseRejectErrorsOptions is upstream's option struct, plus the one split our wire format
// forces on any rule carrying a TypeOrValueSpecifier list.
//
// `allow` arrives as ONE heterogeneous array whose entries are either a bare string matched against
// the type's own name or a specifier object. `Allow` and `AllowInline` are the two typed halves,
// exactly as only-throw-error and no-floating-promises split theirs.
//
// The three booleans are plain `bool` rather than pointers, and that is the OPPOSITE of
// only-throw-error's shape for the same-looking fields. It is correct for this rule and it is worth
// stating why, because copying the neighbour would be wrong: every default here is FALSE, so the
// zero value a nil-options rule receives already IS the default. only-throw-error needs pointers
// because its defaults are true and absent has to stay distinguishable from false.
type PreferPromiseRejectErrorsOptions struct {
	Allow                []type_checking.TypeOrValueSpecifier
	AllowInline          []string
	AllowEmptyReject     bool
	AllowThrowingAny     bool
	AllowThrowingUnknown bool
}

func buildPreferPromiseRejectErrorsMessage() rule.Message {
	return rule.Message{
		Id: "rejectAnError",
		Description: "This rejects a promise with something that is not an Error, so whatever " +
			"catches it gets no stack trace and no message. A rejection is the failure path, and " +
			"the handler that receives a bare string or object has no record of where the failure " +
			"came from. Reject with `new Error(...)` instead.",
	}
}

// PreferPromiseRejectErrors flags a promise rejection whose reason is not an Error.
//
//	valid:   Promise.reject(new Error());
//	valid:   new Promise((resolve, reject) => reject(new Error()));
//	valid:   const o = { reject(x: unknown) {} }; o.reject(5);
//	invalid: Promise.reject(5);
//	invalid: Promise.reject();
//	invalid: new Promise((resolve, reject) => reject(5));
//	invalid: new Promise((resolve, reject) => { const f = () => reject(5); f(); });
//
// Only an Error carries a stack trace. Rejecting with a string or an object literal gives every
// downstream handler a value with no record of where the failure happened.
//
// # Two anchors, and they answer different questions
//
// `KindCallExpression` catches the STATIC form, `Promise.reject(x)`. `KindNewExpression` catches the
// EXECUTOR form, `new Promise((resolve, reject) => reject(x))`, where the thing being called is a
// local binding rather than a member of Promise.
//
// # What decides a finding on the static form
//
// The callee must be a property access, its property must be named `reject`, and the RECEIVER's
// type must be promise-like or a promise constructor. The receiver test is what keeps a plain
// object with a `reject` method silent, and it is load-bearing rather than an optimization:
// measured, `const o = { reject(x: unknown) {} }; o.reject(5)` is clean upstream while a
// `Promise` subclass reports.
//
// Measured against the installed rule, holding the argument fixed at `5`:
//
//	Promise.reject(5)                              REPORTS
//	Promise['reject'](5)                           REPORTS   (computed literal key)
//	Promise?.reject(5)                             REPORTS   (optional chain)
//	class MyPromise extends Promise<void> {}       REPORTS   on MyPromise.reject(5)
//	const o = { reject(x: unknown) {} }            CLEAN     on o.reject(5)
//
// # What decides a finding on the executor form, all four conditions measured
//
// The source states these as a conjunction, which tells you nothing about which conjunct works:
//
//	the callee must be promise-CONSTRUCTOR-like     `new Foo((a, b) => b(5))` is clean
//	the executor must be a function                 a non-function first argument is clean
//	the reject binding must be the SECOND parameter `new Promise(resolve => ...)` is clean
//	it must be a plain identifier parameter         a destructured or rest second parameter is clean
//
// Then every REFERENCE to that binding which appears as a call's callee is checked. Two properties
// of that scan were measured rather than assumed, and both rule out a name-based shortcut:
//
//	new Promise((resolve, reject) => { const f = () => reject(5); f(); })   REPORTS
//	new Promise((resolve, reject) => { take(reject); })                     CLEAN
//
// The first says the scan must reach into NESTED functions, so it is not "calls directly in the
// executor body". The second says the reference must be the CALLEE, so it is not "any mention".
//
// And shadowing must be respected, which is the reason this resolves symbols instead of comparing
// text. Measured CLEAN upstream:
//
//	new Promise((resolve, reject) => {
//	  const inner = (reject: (r: unknown) => void) => reject(5);
//	  inner(() => {});
//	});
//
// The inner `reject` is a different binding, and a port matching on the NAME would report it.
// Upstream gets this from its scope manager's resolved references; we get the same answer from
// `GetSymbolAtLocation`, which resolves the identifier to the declaration it actually binds to.
//
// # What is checked once the call is found, and the order
//
// The order is upstream's and two arms can match the same input:
//
//	no argument      reported unless `allowEmptyReject`. `Promise.reject()` REPORTS by default.
//	allow            the configured specifiers, matched against the argument's type
//	any              silent only if `allowThrowingAny`, which defaults to FALSE here
//	unknown          silent only if `allowThrowingUnknown`, which defaults to FALSE here
//	Error-like       silent, and a READONLY Error-like counts too
//	everything else  reported
//
// The any/unknown defaults are the trap for anyone reading only-throw-error first: there they
// default to TRUE and a bare `throw someAny` is silent, while here `Promise.reject(someAny)`
// REPORTS with no configuration. Measured both ways rather than inherited.
//
// # A spread argument is reported rather than skipped
//
// `Promise.reject(...args)` REPORTS. Upstream reads `arguments.at(0)` without testing its kind, so
// a spread element is handed to the type check like any other argument and its type is not
// Error-like. Reproduced rather than improved on; measured.
//
// # Where the finding points
//
// The whole call expression, not the argument. `Promise.reject(5)` reports over
// `Promise.reject(5)`, which is what makes the no-argument case reportable at all.
//
// # One unresolved real-tree difference, recorded rather than left for the next reader to find
//
// On `modules/samsung/frame-tv/FrameTvApi.ts` the full `cohere` run produces FIVE findings and
// upstream, driven over the same file against the repository's root tsconfig, produces SIX. The
// missing one is line 644, a `reject(error)` inside an immediately-invoked async function nested
// under a `function (resolve, reject)` executor.
//
// What has been established, so the next person does not repeat it:
//
//   - It is not the rule's shape. Handed that file's own text through `rule_testing.RunTyped`, this
//     rule finds all SIX, at exactly upstream's lines. Reduced versions of the nesting -- arrow
//     executor and function executor, each wrapping a `socket.connect` callback around an async
//     IIFE -- both report here.
//   - It is not the rule's reach into nested functions generally, since lines 591, 597 and 601 in
//     the same file are found and sit inside the same kind of nesting.
//   - It is not a suppression comment; the file has none.
//   - The whole TAIL of that 765-line file is silent in the full run: 601 is the highest line any
//     rule at all reports there. That points at the program or file boundary the full run builds
//     rather than at this rule, and it is why this is recorded as unresolved rather than as a
//     divergence in the rule's judgment.
//
// It is stated here rather than in a fixture because a fixture would have to assert the wrong
// number to stay green: through `rule_testing` the rule already agrees with upstream on all six.
//
// # Cost
//
// Both anchors are common kinds, so both listeners exit as early as possible: the call arm tests the
// property name before asking the checker anything, and the new arm tests the argument shape before
// resolving. The reference scan runs only for a `new Promise` whose executor has a named second
// parameter.
var PreferPromiseRejectErrors = rule.Rule{
	Name: "@typescript-eslint/prefer-promise-reject-errors",

	// Every arm reads a type: the receiver's, the callee's, or the rejection reason's.
	NeedsTypeChecker: true,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[PreferPromiseRejectErrorsOptions](options)

		// A rule configured as bare `"error"` is handed nil options, and `options.(T)` on nil yields
		// the zero value. Every default here is FALSE, so the zero value is already correct and no
		// defaulting is needed. This is asserted by a fixture that bypasses the decoder, because the
		// day a default flips to true this line becomes wrong silently.

		checkRejectCall := func(call *ast.Node) {
			arguments := call.Arguments()

			if len(arguments) == 0 {
				if settings.AllowEmptyReject {
					return
				}
				ctx.ReportNode(call, buildPreferPromiseRejectErrorsMessage())
				return
			}

			// Upstream reads `arguments.at(0)` with no kind test, so a SPREAD element is type-checked
			// like any other argument. Measured: `Promise.reject(...args)` reports.
			argument := arguments[0]
			argumentType := ctx.TypeChecker.GetTypeAtLocation(argument)
			if argumentType == nil {
				return
			}

			if type_checking.TypeMatchesSomeSpecifier(argumentType, settings.Allow, settings.AllowInline, ctx.Program) {
				return
			}
			if settings.AllowThrowingAny && type_checking.IsTypeAnyType(argumentType) {
				return
			}
			if settings.AllowThrowingUnknown && type_checking.IsTypeUnknownType(argumentType) {
				return
			}
			if type_checking.IsErrorLike(ctx.Program, ctx.TypeChecker, argumentType) {
				return
			}
			// A readonly Error-like is a separate predicate upstream and not subsumed by the one
			// above: `Readonly<Error>` is not assignable through the same path.
			if type_checking.IsReadonlyErrorLike(ctx.Program, ctx.TypeChecker, argumentType) {
				return
			}

			ctx.ReportNode(call, buildPreferPromiseRejectErrorsMessage())
		}

		return rule.Listeners{
			// The static form: Promise.reject(x)
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				callee := ast.SkipParentheses(node.Expression())
				if callee == nil {
					return
				}

				receiver, propertyName := preferPromiseRejectErrorsStaticMemberAccess(callee)
				if receiver == nil || propertyName != "reject" {
					return
				}

				// The NAME alone is not enough: any object can have a `reject` method. Measured, a
				// plain object literal with one is clean upstream.
				receiverType := ctx.TypeChecker.GetTypeAtLocation(receiver)
				if receiverType == nil {
					return
				}
				if !type_checking.IsPromiseConstructorLike(ctx.Program, ctx.TypeChecker, receiverType) &&
					!type_checking.IsPromiseLike(ctx.Program, ctx.TypeChecker, receiverType) {
					return
				}

				checkRejectCall(node)
			},

			// The executor form: new Promise((resolve, reject) => reject(x))
			ast.KindNewExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				callee := ast.SkipParentheses(node.Expression())
				if callee == nil {
					return
				}
				calleeType := ctx.TypeChecker.GetTypeAtLocation(callee)
				if calleeType == nil || !type_checking.IsPromiseConstructorLike(ctx.Program, ctx.TypeChecker, calleeType) {
					return
				}

				arguments := node.Arguments()
				if len(arguments) == 0 {
					return
				}
				executor := ast.SkipParentheses(arguments[0])
				if executor == nil || !preferPromiseRejectErrorsIsFunction(executor) {
					return
				}

				// The reject binding is the SECOND parameter, and it has to be a plain identifier.
				// A destructured or rest second parameter binds something other than the rejection
				// callback, and upstream declines both; measured.
				parameters := executor.Parameters()
				if len(parameters) < 2 {
					return
				}
				rejectParameter := parameters[1]
				if rejectParameter.AsParameterDeclaration().DotDotDotToken != nil {
					return
				}
				// The identifier half of this test is SUBSUMED and kept deliberately. Mutating it
				// away survives the whole suite, and that verdict is correct: a destructured second
				// parameter has a `KindObjectBindingPattern` name, and `GetSymbolAtLocation` on a
				// binding pattern returns NIL, so the guard immediately below already declines every
				// input this one could. Probed directly on
				// `new Promise((resolve, { reject }: any) => reject(5))`, printing the name's kind
				// and the resolved symbol: `KindObjectBindingPattern`, symbol nil.
				//
				// Kept because it states the requirement where a reader looks for it rather than
				// leaving it to a nil check whose purpose reads as defensive, and because the
				// subsumption depends on a checker behaviour that is not obvious from here.
				rejectName := rejectParameter.Name()
				if rejectName == nil || !ast.IsIdentifier(rejectName) {
					return
				}

				// Resolving the binding to a SYMBOL rather than matching its text is what makes
				// shadowing work. A nested parameter with the same name resolves to a different
				// symbol, and upstream is silent on it.
				//
				// A DUPLICATE parameter name is a stated divergence, and it is a difference between
				// the two SUBSTRATES rather than between the two rules.
				//
				// Upstream's invalid case 49 is `new Promise(function (reject, reject) { reject(5); })`,
				// which it reports. ESLint's scope manager merges same-named parameters into ONE
				// variable whose `identifiers` list holds both, so asking for the variable
				// containing the second parameter finds the same variable the call site resolves
				// through, and the reference is found.
				//
				// Our checker does not merge them. Probed directly on that exact source, printing
				// the symbol pointer at each site:
				//
				//	param[0] name="reject" symbol=0x...b500
				//	param[1] name="reject" symbol=0x...b560   <- a DIFFERENT symbol
				//	call callee "reject"   symbol=0x...b500   <- resolves to the FIRST
				//
				// So the call binds to parameter zero while the rule anchors on parameter one, and
				// no symbol comparison can connect them. Reproducing upstream's answer would mean
				// falling back to name matching for this shape, which would cost the shadowing case
				// that the whole symbol approach exists to get right, and shadowing has real corpus
				// behind it while this has one case whose source is a TypeScript error anyway
				// (duplicate parameter names are illegal in strict mode).
				//
				// So: this input is silent here and reports upstream, deliberately. The test file
				// pins it as silent with the reasoning at the line, so it fails loudly if the
				// checker ever starts merging.
				rejectSymbol := ctx.TypeChecker.GetSymbolAtLocation(rejectName)
				if rejectSymbol == nil {
					return
				}

				// The scan reaches into nested functions, because a reject called from a closure
				// inside the executor still rejects this promise. Measured reporting upstream.
				preferPromiseRejectErrorsForEachCallOfSymbol(ctx, executor, rejectSymbol, checkRejectCall)
			},
		}
	},
}

// preferPromiseRejectErrorsStaticMemberAccess reads a property access, returning the receiver and
// the property's name, for both the dotted and the computed-literal spellings.
//
// Upstream's `isStaticMemberAccessOfValue` additionally folds a `const` bound to a string literal
// through its static evaluator. That is not reproduced, and the cost is stated rather than silent:
// `const k = 'reject'; Promise[k](5)` reports here and is silent upstream. Reproducing it means a
// constant-folding pass, which is substantially larger than the rest of this rule, and no case in
// the corpus writes that shape. The same divergence is recorded on only-throw-error, which shares
// the upstream helper.
func preferPromiseRejectErrorsStaticMemberAccess(callee *ast.Node) (receiver *ast.Node, propertyName string) {
	switch {
	case ast.IsPropertyAccessExpression(callee):
		name := callee.AsPropertyAccessExpression().Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return nil, ""
		}
		return callee.AsPropertyAccessExpression().Expression, name.Text()

	case ast.IsElementAccessExpression(callee):
		argumentExpression := ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		if argumentExpression == nil || !ast.IsStringLiteralLike(argumentExpression) {
			return nil, ""
		}
		return callee.AsElementAccessExpression().Expression, argumentExpression.Text()
	}

	return nil, ""
}

// preferPromiseRejectErrorsIsFunction answers upstream's `isFunction`, which accepts an arrow, a
// function expression, and a function declaration.
//
// Measured rather than restricted to arrows: `new Promise(function (resolve, reject) { reject(5); })`
// reports upstream, so an arrow-only test would go silent on it.
func preferPromiseRejectErrorsIsFunction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration:
		return true
	}
	return false
}

// preferPromiseRejectErrorsForEachCallOfSymbol visits every call in the executor whose callee is an
// identifier resolving to the given symbol.
//
// This replaces upstream's `variable.references` walk, which reads a resolved-reference index we do
// not have. The checker answers the same question one identifier at a time, so the scan is a walk
// of the executor's own subtree asking `GetSymbolAtLocation` at each identifier in callee position.
//
// Two properties are deliberate and both are pinned by fixtures:
//
//   - It recurses into NESTED functions. A reject called from a closure inside the executor still
//     rejects this promise, and upstream reports it.
//   - It only considers an identifier in CALLEE position. Passing reject somewhere as a value is not
//     a rejection, and upstream is silent on it.
func preferPromiseRejectErrorsForEachCallOfSymbol(
	ctx rule.Context,
	root *ast.Node,
	target *ast.Symbol,
	visit func(call *ast.Node),
) {
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		if ast.IsCallExpression(node) {
			callee := ast.SkipParentheses(node.Expression())

			// The `IsIdentifier` half is INERT and kept as documentation of the intent. Mutating
			// it away survives the suite, and that is correct rather than a fixture gap: the symbol
			// comparison below already does all the work. Probed over four non-identifier callee
			// shapes inside an executor, printing each callee's kind and whether it resolved to the
			// reject parameter's symbol:
			//
			//	(reject)(5)          KindIdentifier               matches   (parens already skipped)
			//	o.reject(5)          KindPropertyAccessExpression no match  (resolves to the property)
			//	[reject][0](5)       KindElementAccessExpression  no match
			//	(0, reject)(5)       KindBinaryExpression         no match
			//
			// Nothing that is not an identifier resolves to a parameter symbol, so the guard can
			// only ever decline inputs the comparison declines anyway.
			if callee != nil && ast.IsIdentifier(callee) {
				// Symbol identity rather than name equality: this is what declines a shadowing
				// inner parameter that happens to be spelled the same.
				if ctx.TypeChecker.GetSymbolAtLocation(callee) == target {
					visit(node)
				}
			}
		}

		node.ForEachChild(walk)
		return false
	}

	root.ForEachChild(walk)
}

// preferPromiseRejectErrorsRawOptions is the wire shape, which is not the shape the rule reasons
// with. Three booleans bind straight through; `allow` does not, for the reasons on the options type.
type preferPromiseRejectErrorsRawOptions struct {
	Allow                []preferPromiseRejectErrorsRawSpecifier `json:"allow"`
	AllowEmptyReject     bool                                    `json:"allowEmptyReject"`
	AllowThrowingAny     bool                                    `json:"allowThrowingAny"`
	AllowThrowingUnknown bool                                    `json:"allowThrowingUnknown"`
}

// preferPromiseRejectErrorsRawSpecifier is one entry of the allowlist, in either wire form.
type preferPromiseRejectErrorsRawSpecifier struct {
	inline string

	From    string
	Name    []string
	Path    string
	Package string
}

// UnmarshalJSON accepts both wire forms: a bare string, or the specifier object.
//
// `name` is itself two shapes, a string or an array of strings. An unrecognized `from` is dropped
// rather than erroring, matching how the sibling rules treat the same field: the specifier then
// matches nothing instead of failing the run, and dropping it says nothing rather than silently
// meaning `file`, which is the zero value.
func (s *preferPromiseRejectErrorsRawSpecifier) UnmarshalJSON(raw []byte) error {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		*s = preferPromiseRejectErrorsRawSpecifier{inline: asString}
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

	*s = preferPromiseRejectErrorsRawSpecifier{From: object.From, Path: object.Path, Package: object.Package}

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

// DecodePreferPromiseRejectErrorsOptions maps upstream's JSON onto the struct the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` because `allow` arrives as one heterogeneous
// array and leaves as two typed fields, and because `from` is a string on the wire and an integer
// enum in the struct.
func DecodePreferPromiseRejectErrorsOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[preferPromiseRejectErrorsRawOptions]()(raw)
	if err != nil {
		return PreferPromiseRejectErrorsOptions{}, err
	}

	wire, _ := decoded.(preferPromiseRejectErrorsRawOptions)

	options := PreferPromiseRejectErrorsOptions{
		AllowEmptyReject:     wire.AllowEmptyReject,
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
		// This arm has no fixture that can see it, and the reason is a HARNESS limit rather than a
		// gap. A `from: package` specifier resolves against a package declaration, which upstream's
		// own corpus writes as an ambient `declare module 'errors'`, and ruletest pins
		// `moduleDetection: "force"` (internal/rule_testing/program.go:27), under which that module is
		// unresolvable. Measured: the imported type comes back as `any`, confirmed by the same input
		// going clean under `allowThrowingAny`. An `any` carries no symbol, so no specifier can
		// match it and this arm cannot change a verdict here.
		//
		// A mutation replacing it with `continue` therefore survives the suite, correctly. The `file`
		// and `lib` arms beside it are both caught, which is the control saying the sweep can see
		// this switch at all.
		case "package":
			specifier.From = type_checking.TypeOrValueSpecifierFromPackage
		default:
			continue
		}
		options.Allow = append(options.Allow, specifier)
	}

	return options, nil
}
