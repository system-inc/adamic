package typescript

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/consistentreturn"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ConsistentReturnSettings is the decoded option surface, which is the core rule's unchanged.
//
// Upstream's `schema: baseRule.meta.schema`, so the extension adds no option of its own. The
// wrapper reads `options?.treatUndefinedAsUnspecified === true` and passes everything else through.
type ConsistentReturnSettings struct {
	// TreatUndefinedAsUnspecified makes a return of `undefined` count as returning nothing. In the
	// extension this is a question about the argument's TYPE as well as its spelling.
	TreatUndefinedAsUnspecified bool
}

type consistentReturnWire struct {
	TreatUndefinedAsUnspecified *bool `json:"treatUndefinedAsUnspecified"`
}

// DecodeConsistentReturnOptions reads the option object off the config.
//
// # The config layer hands this a different shape than a fixture does
//
// `internal/lint/configuration/configuration.go` keeps `setting.Options = tuple[1]`, a single JSON
// value after the severity, while ESLint's `context.options` is every element after it. This rule's
// `meta.schema` declares ONE element, so the two shapes coincide here and the object arrives whole.
// Written out rather than assumed, because a decoder that is wrong for the config shape passes every
// fixture: a fixture hands the decoder bytes the test built, never bytes the config layer sliced.
// Pinned by `TestConsistentReturnDecoderAcceptsTheShapesTheConfigLayerDelivers`, which crosses that
// boundary explicitly.
//
// A malformed configuration is refused rather than decoded to an empty one. A decoder answering "no
// options" to a shape it did not understand produces a rule that registers, reports, and enforces
// something other than what the config says, which is the failure this repository exists to catch.
func DecodeConsistentReturnOptions(raw []byte) (any, error) {
	settings := ConsistentReturnSettings{}
	if len(raw) == 0 {
		return settings, nil
	}
	var wire consistentReturnWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return settings, err
	}
	if wire.TreatUndefinedAsUnspecified != nil {
		settings.TreatUndefinedAsUnspecified = *wire.TreatUndefinedAsUnspecified
	}
	return settings, nil
}

var messageConsistentReturnMissingReturn = rule.Message{
	Id: "missingReturn",
	Description: "Some paths through this function return a value and some run off the end, which " +
		"returns undefined. A caller reading one branch cannot tell which kind of function this " +
		"is, so the undefined arrives somewhere far from here.",
}

var messageConsistentReturnMissingReturnValue = rule.Message{
	Id: "missingReturnValue",
	Description: "This returns nothing while another return in the same function returns a value. " +
		"The caller receives undefined from one path and a value from another, and nothing in the " +
		"signature says which.",
}

var messageConsistentReturnUnexpectedReturnValue = rule.Message{
	Id: "unexpectedReturnValue",
	Description: "This returns a value while another return in the same function returns nothing. " +
		"A reader who saw the bare return will not expect a value to come back from here.",
}

// ConsistentReturn is typescript-eslint's extension of the core rule of the same name.
//
//	valid:   declare function bar(): void;
//	         function foo(flag: boolean): void { if (flag) { return bar(); } return; }
//	valid:   function foo(flag?: boolean): number | void { if (flag) { return 42; } return; }
//	valid:   async function foo(f?: boolean): Promise<void> { if (f) { return bar(); } return; }
//	invalid: function foo(flag: boolean): any { if (flag) return true; else return; }
//	invalid: async function foo(flag: boolean): Promise<string> { if (flag) return; else return 'v'; }
//
// # This extension is a real narrowing, not a second name for the core rule
//
// Two of the three typescript-eslint extension rules ported alongside this one turned out to be
// aliases whose filters our core already carries. This one is not, and it was measured rather than
// assumed. Driving the installed 8.67.0 extension and the installed ESLint 10.8.1 core over
// upstream's own 30-case corpus for the extension:
//
//	identical verdicts   17 of 30
//	divergent            13 of 30
//
// Every one of the thirteen is the extension going SILENT where the core reports, and every one
// needs the type checker to reach. Twelve are the `void` suppression below and one is the typed
// `undefined` argument. A bare port of the core rule under this name would report on thirteen of
// upstream's nineteen passing cases.
//
// # The first filter: a bare `return` in a void-returning function is not a decision
//
// Upstream's `isReturnVoidOrThenableVoid`. It reads the FUNCTION's type, takes its call signatures,
// and asks whether any signature returns `void` -- or, for an `async` function, whether the
// signature's return type is a thenable whose awaited type is `void`, recursing through nested
// promises. If so, a `return;` inside it is dropped before the core rule ever sees it.
//
// `.some()` over the call signatures rather than the first is load bearing: an overloaded function
// carries one signature per overload, and upstream's own corpus asserts that
// `function foo(): boolean; function foo(flag: boolean): void;` is clean because ONE of the two is
// void. Measured through the probe: the checker hands that implementation two signatures, `boolean`
// and `void`, and only the second suppresses.
//
// # Why this is a hook into the judgment rather than a filter over its findings
//
// Upstream intercepts the `ReturnStatement` listener and returns early, so the suppressed return
// never reaches the core at all. That is not the same as dropping the finding it would produce,
// because the FIRST return in a function sets the expectation every later return is measured
// against. Suppress it as an input and a later `return 1;` is the first return and reports nothing;
// suppress it as an output and the bare return still sets a no-value expectation that the `return 1`
// then contradicts. The shared judgment therefore takes the question as a hook, asked while the
// returns are being collected.
//
// The same distinction was found the hard way on `no-dupe-class-members`, where filtering the output
// reported on a member whose own key was not computed.
//
// # The second filter: `treatUndefinedAsUnspecified` becomes a question about types
//
// The core rule reads the SPELLING: a bare identifier `undefined`, or any `void` expression. The
// extension additionally asks the checker whether the argument's type is exactly `undefined`, which
// catches `return undef;` for a `declare const undef: undefined`. Upstream expresses this by handing
// the core a synthesised return whose argument is null; the shared judgment reaches the same answer
// by treating the argument as absent, because nothing downstream reads it again.
//
// `flags === ts.TypeFlags.Undefined` is an EQUALITY on the whole flag word rather than a mask, so a
// union containing undefined does not qualify. Upstream's own corpus pins the difference: the
// `declare const undefOrNum: undefined | number` case REPORTS while the `declare const undef:
// undefined` case is clean, and only an equality separates them. Reproduced as an equality, with a
// fixture on each side.
//
// # The shelf's IsTypeFlagSet is the WRONG helper here, and using it costs upstream's own cases
//
// `type_checking.IsTypeFlagSet` reads the type's own flags. typescript-eslint's `isTypeFlagSet` first
// decomposes a union and ORs its constituents' flags, which is a different question. Measured on the
// corpus through a probe, where `voidFlag` is the shelf's answer and `voidUnionFlag` is upstream's:
//
//	void                     voidFlag=T   voidUnionFlag=T    agree
//	void | number            voidFlag=F   voidUnionFlag=T    DISAGREE
//	Promise<void | number>   arg0 void=F  arg0 union void=T  DISAGREE
//
// Both disagreements are upstream PASSING cases -- `function foo(flag?: boolean): number | void` and
// the `PromiseVoidNumber` alias -- so reaching for the shelf helper by name would have reported two
// of the nineteen clean cases while every other fixture stayed green. That is the failure the
// standard describes as probing a helper against its doc comment rather than against upstream: the
// shelf helper is accurate about what it does and is not the question this rule asks.
//
// `unionFlagSet` below is upstream's `getTypeFlags`, written here rather than added to the shelf
// because changing `type_checking.IsTypeFlagSet` would silently move every one of its other callers.
//
// # The name is not checkable from this package, and a sweep here says nothing about it
//
// A mutation renaming this rule survives a mutation sweep scoped to its own tests AND a sweep over
// the whole `typescript` package, because every instrument in both is a fixture that names the rule
// variable rather than its string. `internal/lint/registry` is what catches it, and it catches it
// twice: `TestEveryRuleShipsAFixturePair` loses the rule's test file, and
// `TestEveryRegisteredRuleIsReachableFromTheLiveConfig` reports the renamed name as unreachable.
//
// Measured, because "unscoped" reads as "sees everything" and does not: the sweep tool builds and
// tests one package, so widening its scope from one test pattern to the whole package widens nothing
// that matters here. Anything resting on a cross-package guard has to be scored by running that
// package's suite directly.
//
// # NeedsTypeChecker, and no ProgramReads
//
// `GetTypeAtLocation` on a function whose return type is an imported alias resolves across a module
// boundary, and `IsThenableType` walks to the `Promise` declaration in the default library. Both are
// the checker's answers, not reads of ctx.Program, and the findings cache keys them on the type
// fingerprint, which covers the import closure, and on the binary, which holds the default library.
var ConsistentReturn = rule.Rule{
	Name:             "@typescript-eslint/consistent-return",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[ConsistentReturnSettings](options)
		if !ok {
			settings = ConsistentReturnSettings{}
		}

		return rule.Listeners{
			// One listener over the whole file rather than one per function kind, because the
			// judgment is about a function as a whole and the walk here is pre-order: a listener on
			// the function cannot see the returns that follow it.
			ast.KindSourceFile: func(node *ast.Node) {
				consistentreturn.Judge(node,
					consistentreturn.Settings{
						TreatUndefinedAsUnspecified: settings.TreatUndefinedAsUnspecified,
					},
					consistentreturn.Hooks{
						SuppressBareReturn: func(scope *ast.Node, _ *ast.Node) bool {
							return consistentReturnIsVoidOrThenableVoid(ctx, scope)
						},
						TreatsArgumentAsUnspecified: func(argument *ast.Node) bool {
							return consistentReturnArgumentIsExactlyUndefined(ctx, argument)
						},
					},
					consistentReturnReporter(ctx),
					consistentReturnDescribe)
			},
		}
	},
}

// consistentReturnReporter wires the shared judgment's findings into this rule's context.
func consistentReturnReporter(ctx rule.Context) consistentreturn.Reporter {
	return consistentreturn.Reporter{
		ReportNode: func(node *ast.Node, messageId string, description string) {
			ctx.ReportNode(node, rule.Message{Id: messageId, Description: description})
		},
		ReportRange: func(textRange core.TextRange, messageId string, description string) {
			ctx.ReportRange(textRange, rule.Message{Id: messageId, Description: description})
		},
		TokenRange: func(node *ast.Node) core.TextRange {
			return rule.TokenRange(ctx.SourceFile, node)
		},
	}
}

// consistentReturnDescribe renders the sentence for one finding.
//
// The casing difference between the two shapes is upstream's and both corpora assert it: the
// per-return messages take a capitalised name, `missingReturn` takes a lower-case one.
func consistentReturnDescribe(messageId string, name string) string {
	if messageId == consistentreturn.MessageIdMissingReturn {
		return fmt.Sprintf("Expected to return a value at the end of %s. %s",
			name, messageConsistentReturnMissingReturn.Description)
	}
	description := messageConsistentReturnMissingReturnValue.Description
	if messageId == consistentreturn.MessageIdUnexpectedReturnValue {
		description = messageConsistentReturnUnexpectedReturnValue.Description
	}
	return fmt.Sprintf("%s %s %s", name, consistentreturn.Verb(messageId), description)
}

// consistentReturnIsVoidOrThenableVoid is upstream's `isReturnVoidOrThenableVoid`.
//
// The function's own type, its call signatures, and `.some()` over them: one void-returning overload
// is enough, which upstream's overload case asserts.
//
// The `async` branch is chosen by the function's own `async` modifier rather than by whether the
// return type happens to be thenable, because that is what upstream reads (`node.async`). A
// non-async function declared to return `Promise<void>` therefore takes the plain branch and is NOT
// suppressed, which upstream's corpus pins directly: `function foo(flag?: boolean): Promise<void>`
// with a bare return REPORTS, while the same signature with `async` is clean.
func consistentReturnIsVoidOrThenableVoid(ctx rule.Context, scope *ast.Node) bool {
	// The Program is a scope for the judgment but has no type, and upstream's wrapper only ever
	// receives the three function kinds.
	if !consistentreturn.IsScope(scope) {
		return false
	}
	if ctx.TypeChecker == nil {
		return false
	}

	functionType := ctx.TypeChecker.GetTypeAtLocation(scope)
	if functionType == nil {
		return false
	}
	isAsync := scope.ModifierFlags()&ast.ModifierFlagsAsync != 0

	for _, signature := range type_checking.GetCallSignatures(ctx.TypeChecker, functionType) {
		returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
		if returnType == nil {
			continue
		}
		if isAsync {
			if consistentReturnIsPromiseVoid(ctx, scope, returnType) {
				return true
			}
			continue
		}
		if consistentReturnUnionFlagSet(returnType, checker.TypeFlagsVoid) {
			return true
		}
	}
	return false
}

// consistentReturnIsPromiseVoid is upstream's `isPromiseVoid`, including its recursion.
//
// A thenable type reference whose first type argument is `void`, or is itself a promise of void.
// `Promise<Promise<void | undefined>>` is one of upstream's passing cases and only the recursion
// reaches it.
//
// The type-reference guard is not defensive. `Checker_getTypeArguments` PANICS on a type that is not
// a reference, reproduced directly in `internal/consistent_return_ext_probe` on a plain `void`
// return type, and a panic costs every rule in the package its verdict on that file. Upstream gets
// the same guard from `tsutils.isTypeReference` and can afford to treat it as a type narrowing;
// here it is a crash guard as well.
func consistentReturnIsPromiseVoid(ctx rule.Context, node *ast.Node, subject *checker.Type) bool {
	// A depth bound rather than an unbounded recursion. Upstream has none, because its recursion is
	// over resolved type arguments which terminate; the bound is here because a self-referential
	// alias would otherwise be a hang rather than a finding, and 32 is far past any real promise
	// nesting.
	for depth := 0; depth < 32; depth++ {
		if subject == nil {
			return false
		}
		if !type_checking.IsThenableType(ctx.TypeChecker, node, subject) {
			return false
		}
		if !consistentReturnIsTypeReference(subject) {
			return false
		}
		arguments := checker.Checker_getTypeArguments(ctx.TypeChecker, subject)
		if len(arguments) == 0 {
			return false
		}
		awaited := arguments[0]
		if awaited == nil {
			return false
		}
		if consistentReturnUnionFlagSet(awaited, checker.TypeFlagsVoid) {
			return true
		}
		subject = awaited
	}
	return false
}

// consistentReturnIsTypeReference is `tsutils.isTypeReference`: an object type with the Reference
// object flag, which is what makes `Checker_getTypeArguments` safe to call.
func consistentReturnIsTypeReference(subject *checker.Type) bool {
	return type_checking.IsTypeFlagSet(subject, checker.TypeFlagsObject) &&
		checker.Type_objectFlags(subject)&checker.ObjectFlagsReference != 0
}

// consistentReturnUnionFlagSet is typescript-eslint's `isTypeFlagSet`, which is NOT the shelf's.
//
// Upstream's `getTypeFlags` ORs the flags of every union constituent before masking, so
// `void | number` answers true for Void where the shelf's `type_checking.IsTypeFlagSet` answers false.
// Both of the corpus cases that separate the two readings are upstream PASSING cases, so the shelf
// helper would have reported them. See the rule's doc comment for the measurement.
//
// Not added to `internal/lint/checking` under a new name, because the shelf already has a function
// spelled almost identically and a second one beside it is how two callers end up asking different
// questions believing they asked the same one.
func consistentReturnUnionFlagSet(subject *checker.Type, flags checker.TypeFlags) bool {
	if subject == nil {
		return false
	}
	var combined checker.TypeFlags
	for part := range type_checking.UnionTypePartsSeq(subject) {
		if part == nil {
			continue
		}
		combined |= checker.Type_flags(part)
	}
	return combined&flags != 0
}

// consistentReturnArgumentIsExactlyUndefined is upstream's typed `treatUndefinedAsUnspecified` arm.
//
// `returnValueType.flags === ts.TypeFlags.Undefined`, written as an equality because that is what
// upstream writes.
//
// # The equality and a mask are indistinguishable here, and that was measured rather than argued
//
// The obvious reading is that the equality is load bearing: a union containing undefined would carry
// extra bits and a mask would wrongly accept it. That reading is WRONG, and a first draft of this
// file shipped it as a documented decision with a test on each side.
//
// A union does not carry its constituents' bits. `undefined | number` has flags 134217728, which is
// `Union` alone, so `flags & Undefined` is already zero and the mask agrees with the equality. Probed
// across eight shapes in `internal/consistent_return_ext_probe`:
//
//	undefined (declared, literal, or `void 0`)   flags 4          equality T   mask T
//	undefined | number, and an optional param    flags 134217728  equality F   mask F
//	any                                          flags 1          equality F   mask F
//	never                                        flags 262144     equality F   mask F
//
// No input separates them, and a mutation turning the equality into a mask SURVIVES the whole
// fixture set for that reason. The survival is genuine equivalence rather than a blind spot, which
// is a different claim from "the fixtures cannot see it" and is why it is written down here instead
// of chased with another fixture.
//
// Kept as an equality anyway, because it is upstream's spelling and the two are equivalent only as
// long as the checker keeps representing unions this way. Nothing is asserted about the difference,
// since there is none to assert.
func consistentReturnArgumentIsExactlyUndefined(ctx rule.Context, argument *ast.Node) bool {
	if ctx.TypeChecker == nil || argument == nil {
		return false
	}
	argumentType := ctx.TypeChecker.GetTypeAtLocation(argument)
	if argumentType == nil {
		return false
	}
	return checker.Type_flags(argumentType) == checker.TypeFlagsUndefined
}
