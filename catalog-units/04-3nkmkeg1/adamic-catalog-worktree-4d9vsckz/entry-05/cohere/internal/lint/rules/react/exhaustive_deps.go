package react

import (
	"regexp"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var (
	messageExhaustiveDepsMissing = rule.Message{
		Id: "exhaustiveDepsMissing",
		Description: "This Hook reads a value that is not in its dependency array, so the callback " +
			"React keeps is the one built on the render where the array last changed. The value " +
			"inside it is frozen at that render and every later read sees the stale copy. Add the " +
			"value to the array, or stop reading it.",
	}
	messageExhaustiveDepsUnnecessary = rule.Message{
		Id: "exhaustiveDepsUnnecessary",
		Description: "This dependency array names a value the Hook never reads, so the Hook re-runs " +
			"whenever that value changes for no reason. Remove it from the array.",
	}
	messageExhaustiveDepsDuplicate = rule.Message{
		Id: "exhaustiveDepsDuplicate",
		Description: "This dependency array names the same value twice. The second copy changes " +
			"nothing about when the Hook re-runs, so it is noise that will drift out of step with " +
			"the first. Remove it.",
	}
	messageExhaustiveDepsConstruction = rule.Message{
		Id: "exhaustiveDepsConstruction",
		Description: "This dependency is a value built fresh on every render, so it is a different " +
			"object each time even when nothing about it changed. Comparing it against the previous " +
			"render never matches and the Hook re-runs every time. Move the value inside the " +
			"callback, or wrap the thing that builds it in useMemo or useCallback.",
	}
	messageExhaustiveDepsComplexExpression = rule.Message{
		Id: "exhaustiveDepsComplexExpression",
		Description: "A dependency array entry has to be a plain value or a property path so it can " +
			"be compared against the previous render. This entry is an expression, which cannot be " +
			"checked statically. Assign it to a variable above the Hook and name that instead.",
	}
	messageExhaustiveDepsLiteral = rule.Message{
		Id: "exhaustiveDepsLiteral",
		Description: "A literal in a dependency array never changes between renders, so it can never " +
			"cause the Hook to re-run and it is doing nothing. Remove it. If the intent was to " +
			"depend on a variable of that name, write the variable without quotes.",
	}
	messageExhaustiveDepsSpread = rule.Message{
		Id: "exhaustiveDepsSpread",
		Description: "A spread in a dependency array hides which values the Hook depends on, so " +
			"neither React nor this rule can tell whether the list is right. Write the dependencies " +
			"out one by one.",
	}
	messageExhaustiveDepsNotArrayLiteral = rule.Message{
		Id: "exhaustiveDepsNotArrayLiteral",
		Description: "The dependency argument is not an array written at the call site, so its " +
			"contents cannot be read here and nothing can cohere them. Pass an array literal.",
	}
	messageExhaustiveDepsAsyncEffect = rule.Message{
		Id: "exhaustiveDepsAsyncEffect",
		Description: "An async effect callback returns a promise, and React expects the return value " +
			"to be a cleanup function, so the cleanup never runs and overlapping runs race. Keep the " +
			"callback synchronous and declare the async function inside it.",
	}
	messageExhaustiveDepsMissingCallback = rule.Message{
		Id: "exhaustiveDepsMissingCallback",
		Description: "This Hook was called with no callback, so there is nothing for it to run. Pass " +
			"the function the Hook is supposed to invoke.",
	}
	messageExhaustiveDepsUselessWithoutDependencies = rule.Message{
		Id: "exhaustiveDepsUselessWithoutDependencies",
		Description: "useMemo and useCallback exist to reuse a value across renders, and they decide " +
			"whether to reuse by comparing a dependency array. Called with only a callback there is " +
			"nothing to compare, so the value is rebuilt every render and the Hook does nothing. " +
			"Pass a dependency array.",
	}
	messageExhaustiveDepsUnknownDependencies = rule.Message{
		Id: "exhaustiveDepsUnknownDependencies",
		Description: "The callback is not written at the call site, so what it reads cannot be seen " +
			"here and the dependency array cannot be checked against it. Pass an inline function.",
	}
	messageExhaustiveDepsEffectEvent = rule.Message{
		Id: "exhaustiveDepsEffectEvent",
		Description: "A function from useEffectEvent always sees the latest render's values, which " +
			"is the whole reason it exists, so listing it as a dependency re-runs the effect for a " +
			"change that could never have gone stale. Remove it from the array.",
	}
	messageExhaustiveDepsSetStateNoDependencies = rule.Message{
		Id: "exhaustiveDepsSetStateNoDependencies",
		Description: "This effect sets state and has no dependency array, so it runs after every " +
			"render, and the state it sets causes the next render. That is a loop with no exit. Pass " +
			"a dependency array saying what the effect actually depends on.",
	}
	messageExhaustiveDepsRefCleanup = rule.Message{
		Id: "exhaustiveDepsRefCleanup",
		Description: "A cleanup function runs after the component has moved on, so a ref read there " +
			"gives whatever the ref points at then, not what it pointed at when the effect ran. Copy " +
			"the ref's value into a local inside the effect and read that local in the cleanup.",
	}
	messageExhaustiveDepsRequireExplicitEffectDeps = rule.Message{
		Id: "exhaustiveDepsRequireExplicitEffectDeps",
		Description: "This effect was called with no second argument, and the configuration asks every " +
			"effect to say what it depends on. An effect with no array runs after every render, which " +
			"is sometimes the intent and often an oversight nobody can tell apart from it. Pass a " +
			"dependency array, or an explicit `undefined` to say every render is meant.",
	}
	messageExhaustiveDepsStaleAssignment = rule.Message{
		Id: "exhaustiveDepsStaleAssignment",
		Description: "Assigning to a variable declared in the component body from inside a Hook " +
			"writes to that render's copy, which is discarded when the next render makes a new one. " +
			"Hold the value in a useRef so it survives, or move the variable inside the callback.",
	}
)

// ExhaustiveDeps checks that a Hook's dependency array names exactly the values its callback reads.
//
//	valid:   function C(props) { useEffect(() => { log(props.foo); }, [props.foo]); }
//	valid:   function C(props) { useEffect(() => { log(props.foo.bar); }, [props.foo]); }
//	valid:   function C() { const [s, setS] = useState(0); useEffect(() => { setS(1); }, []); }
//	valid:   function C() { useEffect(() => { const local = {}; log(local); }, []); }
//	invalid: function C(props) { useEffect(() => { log(props.foo); }, []); }
//	invalid: function C() { const local = {}; useEffect(() => {}, [local]); }
//	invalid: function C() { const o = {}; useEffect(() => { log(o); }, [o]); }
//
// # This rule gives ADVICE, and that changes what a defect costs
//
// Every other rule in this catalog decides whether a line is wrong. This one also says what the
// dependency array should be, and both directions of error land on a person rather than in a
// report. A missing dependency is a stale closure: the callback React kept is the one built on the
// render where the array last changed, so the value inside it is frozen and every later read sees
// the old copy. A spurious dependency is worse in a different way: an array naming a value rebuilt
// on every render never compares equal, so the effect re-runs every render, and an effect that sets
// state then loops forever.
//
// That asymmetry is why the suggested array is reported as a `Suggestion` and never as a `Fix`; see
// "The repair" below.
//
// # The satisfaction rule, which is the whole algorithm in one sentence
//
// A declared dependency satisfies every path it is a prefix of and no path it is a suffix of:
// declaring `props.a.b` satisfies a read of `props.a.b.c` because the outer object changing is the
// only way the inner value can change, while declaring `props.a.b.c` does NOT satisfy a read of
// `props.a.b`, because `props.a.b` can be replaced wholesale without `.c` differing.
//
// React implements that by building a tree rather than comparing lists. `collectRecommendations`
// (bundle line 1056) puts every read path and every declared path into one trie keyed by the dots,
// marks a declared node `isSatisfiedRecursively`, marks a read node `isUsed`, marks every ancestor
// of a read `isSubtreeUsed`, and then walks down: a satisfied node stops the descent and takes its
// whole subtree with it, an unsatisfied used node is missing, and anything else recurses. That is
// reproduced here node for node, because the prefix asymmetry falls out of the descent order and
// any list-comparison shortcut gets one of the two directions wrong.
//
// The corpus pins both directions directly. `[props.foo]` against a read of `props.foo.bar.baz` is
// an upstream PASS, and `[props.foo.bar.baz]` against a read of `props.foo` is an upstream FAIL.
//
// # Which implementation this reproduces, and the seven places they disagree
//
// The authority for this rule is React: they wrote it, they ship it, and `exhaustive-deps` is their
// definition rather than a convention two linters converged on. oxc's `exhaustive_deps.rs` is a
// reimplementation, actively maintained (last changed 2026-08-17, a week before this was written),
// and its disagreements with React are defects rather than a dialect except where noted.
//
// Both were read whole and both were RUN, which is not the same thing and mattered here. React's
// rule is executable from `~/Projects/ahra` through the ESLint Linter API, version 7.1.1, and its
// own test corpus is public: `facebook/react`, `packages/eslint-plugin-react-hooks/__tests__/
// ESLintRuleExhaustiveDeps-test.js`, 314 cases. oxc's release binary is at
// `~/Projects/system/oxc/target/release/oxlint`. Each implementation was run against the OTHER's
// full corpus, which is the measurement that found every divergence below and is the reason none of
// them is a guess:
//
//	React's rule over oxc's 325 cases     4 disagreements
//	oxc's binary over React's 298 cases   3 disagreements
//
// Running React's rule over oxc's corpus with the default JavaScript parser first reported ELEVEN
// disagreements. Seven of those were my probe failing to parse TypeScript, not a rule difference at
// all, and they read as perfectly plausible findings. Re-running through `@typescript-eslint/parser`
// removed them. A probe that cannot parse its input reports silence, and silence is what a passing
// case looks like.
//
// The seven real disagreements, and how each was resolved:
//
//	function F() { useEffect(() => { foo() }, []); const foo = () => { bar() }; function bar() { foo() } }
//	    react  reports a missing dependency on `foo`
//	    oxc    silent
//	    SHIPPED: react.
//
// Mutual recursion between two functions in the component body. oxc decides stability by RECURSING
// into a function's body and asking whether everything it reads is itself stable, with a `visited`
// set to stop the cycle — and the `visited` set returns TRUE on re-entry, so a cycle is treated as
// stable. React's `isFunctionWithoutCapturedValues` looks one level deep at `fnScope.through` and
// does not recurse, so `foo` captures `bar`, `bar` is not stable-known, and `foo` is not stable.
// First principles agree with React and it is not close: `foo` and `bar` are both rebuilt on every
// render, so the effect really does capture a stale `foo`. oxc's `visited` short circuit answers
// "we are already asking about this symbol" with "yes, stable", which is the wrong default for a
// question whose whole point is detecting instability.
//
//	function C(props) { useEffect(() => { log(props.foo!.bar) }, [props.foo!.bar]) }
//	    react  reports missing `props.foo` AND a complex expression in the array
//	    oxc    silent
//	    SHIPPED: oxc, deliberately, and this is the one place the authority does not win.
//
// A TypeScript non-null assertion inside the dependency array. React's `analyzePropertyChain`
// throws on any node type it does not name, `TSNonNullExpression` is not named, and the throw is
// caught and turned into "complex expression". So React reports twice on `[props.foo!.bar]`: once
// saying the entry is unreadable, and once saying the dependency it could not read is missing.
//
// First principles say `props.foo!.bar` and `props.foo.bar` denote the same value — `!` is erased
// entirely at compile time and cannot change what is read at runtime. React's own corpus does not
// contain the case, and its silence is a consequence of a `throw` in a chain walk rather than a
// judgment anyone made. oxc unwraps it and is right. Step 1 of the resolution order is "who is
// actually right", and it outranks step 2, so this ships as oxc. The differential will show cohere
// agreeing with oxlint here; ESLint would disagree, and that is stated rather than hidden.
//
// The identical reasoning covers `as` and parentheses, which React also does not unwrap in the
// array while unwrapping them everywhere else, and which this rule unwraps consistently.
//
//	function E() { const foo = useCallback(() => { foo(); }, [foo]); }
//	    react  reports an unnecessary dependency on `foo`
//	    oxc    silent
//	    SHIPPED: react.
//
// A `useCallback` naming itself. React treats the callback's own binding as a dependency like any
// other, finds nothing in the callback reads it from an outer scope, and calls it unnecessary. It
// is right: `foo` inside its own initializer is the previous render's `foo`, so listing it makes
// the callback rebuild every time it rebuilds, which is a self-sustaining loop.
//
// The remaining two disagreements are Flow-only syntax — the `hook` keyword and a `({}: any)` cast
// — which no parser in this tree accepts, so they are out of scope rather than resolved.
//
// # Why the type checker, and what it changed
//
// React uses `eslint-scope` to answer "which binding does this identifier refer to, and is that
// binding declared between the callback and the component function". `isStableKnownHookValue`
// (bundle line 170) then decides stability by NAME: a `useRef` result is stable, a `useState`
// tuple's second element is stable, and the rule knows that because the initializer is spelled
// `useRef(` or `useState(`, not because anything was typed.
//
// This port resolves the binding through the checker and keeps the name test for stability, and the
// split is deliberate rather than a compromise.
//
// Resolution is not a workaround for a constraint we lack; it is exactly what `eslint-scope` does,
// so doing it with the checker is the same decision reached by a better route. The classification
// was measured against React before anything was built on it, on the shapes where "which scope" is
// the entire question:
//
//	const outerLocal = ...; function C() { useEffect(() => { log(outerLocal) }, []) }   both silent
//	function O() { const a = ...; function C() { const b = ...; useEffect(() => { log(a,b) }, []) } }
//	                                                                    both report b alone
//	function C() { const a = ...; { const b = ...; useEffect(() => { log(a,b) }, []) } } both report both
//
// A block scope between the component and the callback counts; a function scope above the component
// does not. Containment in the enclosing function node reproduces `pureScopes` exactly, and the
// nested-component row is the one that proves it, because a rule that walked all the way up would
// report `a` there and React does not.
//
// Stability is deliberately NOT resolved through the type checker, and that is the harder call. A
// `useRef` container really is stable and the checker could say so from the type, which would let
// this rule recognize a ref that arrived through an alias or a re-export where React's name test
// gives up. It would also report on shapes React is silent for, and this rule gives ADVICE: telling
// someone to remove a dependency React would keep is telling them to introduce a stale closure. The
// name test is what upstream's corpus pins, so the name test is what ships, and the extra reach the
// checker would give is declined on purpose. Recorded here because the next reader will see the
// checker sitting right there and wonder why stability does not use it.
//
// # The repair, and why it is a suggestion rather than a fix
//
// React computes the corrected array and offers it. It offers it as a SUGGESTION, promoted to an
// automatic fix only under an option spelled `enableDangerousAutofixThisMayCauseInfiniteLoops` —
// upstream naming the hazard in the identifier. oxc classifies its own version as
// `DangerousSuggestion` in five of its six dependency-array fix vectors.
//
// So both implementations already say this rewrite must not be applied unattended, and the reason
// is the asymmetry at the top of this comment. This rule therefore carries the corrected array in
// the diagnostic's `Suggestion`, never in `Fixes`, and `internal/fix.ProposalsFrom` collects only
// `Fixes`, so nothing applies it automatically. The advice still reaches the reader, because the
// suggested array is interpolated into the message text.
//
// One measurement rather than a principle, since the brief asks for it: the rewrite is a single
// replacement over the dependency array node, so it is one edit per finding and could not hit the
// overlap refusal in `internal/fix.Resolve` even if it were a fix. The decision is about what the
// edit MEANS, not about whether the engine would accept it.
//
// oxc's rewrite also formats differently from React's — three or more dependencies come out as
// `[\n\tx,\n\ty,\n\tz\n]` from its code generator against React's `[x, y, z]` — so there was no
// single upstream text to reproduce even if the repair were shipped. React's spelling is what the
// message carries.
//
// # What the corpus said that reading the sources did not
//
// Four things, all of them from running the cases rather than from either implementation's prose.
//
// The suggested array's ORDER is the read order, not a sort, and getting that wrong is silent.
// Upstream keeps its dependencies in a JavaScript Map and its missing paths in a Set, both of which
// iterate in insertion order, so the order is carried by the container and upstream never mentions
// it. A Go map has no order at all, so it has to be carried explicitly; sorting instead is the
// obvious substitute and it is wrong, because the suggestion is sorted LATER and only conditionally.
// This rule shipped that bug and no imported fixture could see it, since every corpus case with more
// than one missing dependency happens to be alphabetical. A mutation sweep found it: a mutant
// disabling the non-effect recompute survived, and comparing the two paths showed the MUTANT
// agreeing with React while the original did not.
//
// A dependency array is only sorted when the author already sorted it. `areDeclaredDepsAlphabetized`
// compares the declared keys against their own sort and only then sorts the suggestion, so
// `[b, a]` missing `c` suggests `[b, a, c]` while `[a, b]` missing `c` suggests `[a, b, c]`. Both
// implementations do this and neither says why; the effect is that the suggestion preserves an
// intentional order and tidies an accidental one.
//
// A missing dependency in a NON-effect Hook recomputes the whole suggestion from scratch rather
// than adding to what was declared. `useCallback` and `useMemo` drop unnecessary entries while
// fixing a missing one; `useEffect` keeps them. The tell is the second `collectRecommendations`
// call guarded by `!isEffect`, and no case in either corpus explains the asymmetry in words.
//
// `isEffect` is a REGEX over the hook name, `/Effect($|[^a-z])/`, not a list. So `useMyEffect2`
// is an effect and `useEffective` is not, and a hook named by the `additionalHooks` option inherits
// whichever answer its spelling gives. oxc spells the same test as `hook_name.contains("Effect")`,
// which differs on a name like `useEffectively` — unreachable in both corpora and noted rather than
// resolved.
//
// A read of `ref.current` is attributed to `ref`, but only outside a cleanup function. Inside one it
// becomes a different finding entirely, about the ref having moved on by the time cleanup runs.
//
// # Scope
//
// Ported: missing, unnecessary, and duplicate dependencies, the every-render construction warning,
// complex expressions and literals in the array, spreads, a non-array dependency argument, an async
// effect callback, a missing callback, useMemo and useCallback with no array, an unknown callback,
// a useEffectEvent function listed as a dependency, and set-state in an effect with no array. That
// is 15 of React's 17 diagnostic kinds and 215 of the 228 diagnostics its corpus asserts.
//
// Not ported, stated rather than silently missing:
//
// The `.current`-in-cleanup finding (6 diagnostics) needs every reference to the ref binding across
// the whole component, to ask whether any of them assigns `.current`, and that is a symbol-reference
// index rather than a walk of the callback. `internal/unused/references.go` holds one, but it is
// program-wide phase machinery keyed to exported declarations and it answers a different question
// than "every occurrence of this local binding". Building the local index this needs is a real piece
// of work for one diagnostic kind, and it is the honest reason this is out rather than in.
//
// The stale-assignment finding (7 diagnostics) needs write references to a component-scope binding
// from inside the callback. That one is reachable from the callback walk alone and is a candidate
// for a follow-up; it is out here because it also gates the rest of the rule — upstream returns
// early when it fires — and getting the gate wrong silences everything else.
//
// # One consequence outside this file, stated because nobody else will notice it
//
// `cmd/cohere-differential` plants a control file it calls `gate-only-exhaustive-deps`, whose whole
// job is to prove the harness can see a finding only the gate produces. It rests on cohere not
// implementing this rule, and its own comment says so at length, ending with: a control resting on a
// rule being impractical expires when somebody finds it practical. That is what just happened.
//
// The control is now broken: cohere reports the planted effect too, so the asymmetry it measures is
// gone and `ControlsProven` will stop discriminating. The comment beside it already names the
// replacement it would want, the eight enabled `better-tailwindcss` rules that reach the gate
// through a JavaScript plugin bridge cohere does not have. Not changed here, because that file is
// the differential's own and a rule port should not quietly rewrite the harness that judges it.
//
// # The options
//
// All four of React's options are decoded and honored (#d21war2), each checked against React's own
// rows for it. `additionalHooks` adds Hooks by pattern. `requireExplicitEffectDeps` reports an effect
// called with no second argument; an explicit `undefined` is an answer and satisfies it.
// `experimental_autoDependenciesHooks` names Hooks whose array the compiler fills in, so for those a
// literal `null` in the array position means "no array to check" and the absent array is not a
// setState loop.
//
// `enableDangerousAutofixThisMayCauseInfiniteLoops` promotes each repair React offers as a suggestion
// to a fix the engine applies unattended, at the five places React offers one: the corrected array,
// `[callback]` for a callback that is not a function written here, a dependency array added to an
// effect that sets state, removing a useEffectEvent function from the array, and wrapping a function
// used outside the Hook in useCallback. That is the hazard "The repair" above describes, and it is the
// author's explicit choice by an option named for it, so it is the one place this rule writes code.
var ExhaustiveDeps = rule.Rule{
	Name:             "react-hooks/exhaustive-deps",
	NeedsTypeChecker: true,
	Run:              runExhaustiveDeps,
}

// ExhaustiveDepsOptions is the rule's option surface.
//
// React's `meta.schema` declares four options, and all four are here. See "The options" on the rule.
type ExhaustiveDepsOptions struct {
	// AdditionalHooks is a regular expression naming further Hooks whose first argument is a
	// callback and whose second is a dependency array.
	//
	// A string rather than a compiled pattern because that is what the config carries. An
	// unparseable pattern disables the extension rather than failing the run, which is the one place
	// this rule departs from oxc's option handling: oxc's `from_configuration` returns an error and
	// refuses to start. Refusing to lint an entire tree because one optional regex is malformed is a
	// worse outcome than linting it with the built-in Hook list, and cohere's decoder has no channel
	// to report a config error per rule anyway. An empty string is not an extension, which both
	// implementations already agree on: oxc filters it out explicitly.
	AdditionalHooks string `json:"additionalHooks"`

	// EnableDangerousAutofixThisMayCauseInfiniteLoops applies each repair as a fix rather than offering
	// it as a suggestion. Off by default, and named by upstream for what it risks.
	EnableDangerousAutofixThisMayCauseInfiniteLoops bool `json:"enableDangerousAutofixThisMayCauseInfiniteLoops"`

	// ExperimentalAutoDependenciesHooks names Hooks whose dependency array may be the literal `null`,
	// meaning the compiler supplies it, so there is nothing to check.
	ExperimentalAutoDependenciesHooks []string `json:"experimental_autoDependenciesHooks"`

	// RequireExplicitEffectDeps reports an effect called with no dependency argument at all.
	RequireExplicitEffectDeps bool `json:"requireExplicitEffectDeps"`
}

// exhaustiveDepsRun is the options, read once per file into the shape the visitors use.
type exhaustiveDepsRun struct {
	additionalHooks           *regexp.Regexp
	autoDependenciesHooks     map[string]bool
	dangerousAutofix          bool
	requireExplicitEffectDeps bool
}

// report reports a finding, carrying its repair as a fix only under the dangerous option.
//
// Without the option a repair is not offered here at all, as before, because the message already
// says what to write and this rule never applies one unattended.
func (run exhaustiveDepsRun) report(ctx rule.Context, node *ast.Node, message rule.Message, fixes ...rule.Fix) {
	if run.dangerousAutofix && len(fixes) > 0 {
		ctx.ReportNodeWithFixes(node, message, fixes...)
		return
	}
	ctx.ReportNode(node, message)
}

// hookNames are the Hooks whose dependency array this rule checks, and the argument index the
// callback sits at.
//
// `useImperativeHandle` is the one that takes its callback second, because its first argument is
// the ref being attached to.
var hookCallbackIndexes = map[string]int{
	"useEffect":           0,
	"useLayoutEffect":     0,
	"useCallback":         0,
	"useMemo":             0,
	"useImperativeHandle": 1,
}

// hooksUselessWithoutDependencies are the two Hooks whose entire purpose is the comparison, so
// calling one with no array to compare means the call does nothing.
var hooksUselessWithoutDependencies = map[string]bool{"useMemo": true, "useCallback": true}

// effectHookNamePattern is React's own test for whether a Hook is an effect.
//
// A regular expression rather than a list, and the boundary matters: `Effect` must end the name or
// be followed by something that is not a lowercase letter, so `useMyEffect2` is an effect and
// `useEffective` is not. oxc spells this as a plain substring test, which disagrees on the second.
var effectHookNamePattern = regexp.MustCompile(`Effect($|[^a-z])`)

func runExhaustiveDeps(ctx rule.Context, options any) rule.Listeners {
	settings, _ := rule.OptionsAs[ExhaustiveDepsOptions](options)
	run := exhaustiveDepsRun{
		autoDependenciesHooks:     map[string]bool{},
		dangerousAutofix:          settings.EnableDangerousAutofixThisMayCauseInfiniteLoops,
		requireExplicitEffectDeps: settings.RequireExplicitEffectDeps,
	}
	if settings.AdditionalHooks != "" {
		// A malformed pattern disables the extension rather than failing the run. See the option's
		// own comment for why.
		run.additionalHooks, _ = regexp.Compile(settings.AdditionalHooks)
	}
	for _, name := range settings.ExperimentalAutoDependenciesHooks {
		run.autoDependenciesHooks[name] = true
	}

	return rule.Listeners{
		ast.KindCallExpression: func(node *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}
			visitHookCall(ctx, node, run)
		},
	}
}

// visitHookCall is the rule's whole control flow, and its arm order is upstream's.
//
// The order is observable rather than incidental: a Hook with no callback reports about the missing
// callback and stops, so a `useCallback()` with no arguments never also reports that it has no
// dependency array.
func visitHookCall(ctx rule.Context, node *ast.Node, run exhaustiveDepsRun) {
	call := node.AsCallExpression()
	if call == nil {
		return
	}
	hookName := hookNameWithoutReactNamespace(call.Expression)
	if hookName == "" {
		return
	}
	callbackIndex, known := hookCallbackIndexes[hookName]
	if !known {
		if run.additionalHooks == nil || !run.additionalHooks.MatchString(hookName) {
			return
		}
		callbackIndex = 0
	}

	// A Hook at module scope has no component scope to resolve dependencies against, and
	// `rules-of-hooks` already owns that finding. Reporting it here would say the wrong thing about
	// why the line is wrong.
	componentScope := enclosingFunctionOfHookCall(node)
	if componentScope == nil {
		return
	}

	arguments := call.Arguments.Nodes
	if callbackIndex >= len(arguments) {
		ctx.ReportNode(node, messageExhaustiveDepsMissingCallback)
		return
	}
	callbackArgument := arguments[callbackIndex]

	var dependencyArrayNode *ast.Node
	if callbackIndex+1 < len(arguments) {
		candidate := arguments[callbackIndex+1]
		// An explicit `undefined` is spelled the same as passing nothing, which upstream treats as
		// the absent case rather than as an unreadable one.
		if !(candidate.Kind == ast.KindIdentifier && candidate.Text() == "undefined") {
			dependencyArrayNode = candidate
		}
	}

	isEffect := effectHookNamePattern.MatchString(hookName)

	// No second argument at all, which an explicit `undefined` is not: it is the author answering.
	// Upstream reports on the Hook's name and carries on.
	if run.requireExplicitEffectDeps && isEffect && callbackIndex+1 >= len(arguments) {
		ctx.ReportNode(call.Expression, messageExhaustiveDepsRequireExplicitEffectDeps)
	}

	// For a Hook whose array the compiler fills in, a literal `null` there is the absent array.
	autoDependencies := run.autoDependenciesHooks[hookName] && dependencyArrayNode != nil &&
		ast.SkipParentheses(dependencyArrayNode).Kind == ast.KindNullKeyword

	if (dependencyArrayNode == nil || autoDependencies) && !isEffect {
		if hooksUselessWithoutDependencies[hookName] {
			ctx.ReportNode(node, messageExhaustiveDepsUselessWithoutDependencies)
		}
		return
	}

	callback := unwrapExpression(callbackArgument)
	if callback == nil {
		return
	}
	switch callback.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		visitCallbackWithDependencies(ctx, run, callback, dependencyArrayNode, node, hookName,
			componentScope, isEffect)

	case ast.KindIdentifier:
		// The callback is a name rather than a literal function. Upstream follows the name to its
		// declaration and analyzes that function in place, so a helper defined beside the component
		// is checked as though it had been written inline.
		if dependencyArrayNode == nil || autoDependencies {
			return
		}
		if dependencyArrayNamesIdentifier(dependencyArrayNode, callback.Text()) {
			return
		}
		resolved, resolution := resolvedFunctionBody(ctx, callback)
		switch resolution {
		case callbackResolutionNotAFunction:
			// The name binds to something in the component that is not a function written here —
			// `const myEffect = debounce(() => {...}, delay)` is upstream's own case. The callback
			// itself is then the dependency, because it is rebuilt every render, and upstream says
			// so rather than giving up: the array has to name the callback.
			run.report(ctx, dependencyArrayNode, messageExhaustiveDepsMissing, rule.Fix{
				Range: rule.TokenRange(ctx.SourceFile, dependencyArrayNode),
				Text:  "[" + callback.Text() + "]",
			})
			return
		case callbackResolutionUnknown:
			ctx.ReportNode(node, messageExhaustiveDepsUnknownDependencies)
			return
		}
		visitCallbackWithDependencies(ctx, run, resolved, dependencyArrayNode, node, hookName,
			componentScope, isEffect)

	default:
		ctx.ReportNode(node, messageExhaustiveDepsUnknownDependencies)
	}
}

// dependencyRead is one property path the callback reads, and where it read it.
type dependencyRead struct {
	// path is the dotted chain, `props.foo.bar`, which is the key the whole algorithm works in.
	path string

	// node is the outermost member expression the path was read through, which is what a finding
	// points at.
	node *ast.Node

	// stable is set when the value cannot change between renders, so it never needs declaring.
	stable bool

	// order is the position of the FIRST read of this path in the callback's walk.
	//
	// Upstream never records this because it does not have to: its `dependencies` is a JavaScript
	// Map, which iterates in insertion order, so the read order is carried by the container. A Go
	// map has no order at all, so it has to be carried explicitly or the suggested array would
	// differ between runs of the same input.
	//
	// Sorting instead of recording is the obvious substitute and it is WRONG, which cost a round
	// trip here. The suggested array is sorted only when the author's own array was already sorted,
	// so imposing a sort at gather time silently sorts every suggestion: an author who wrote
	// `[props.c, props.b]` gets told to write `[props.a, props.b, props.c]` where React says
	// `[props.c, props.b, props.a]`. No imported fixture can see it, because every corpus case with
	// more than one missing dependency is alphabetical already.
	order int

	// optional records, per path prefix, whether that step was reached through `?.`, so the
	// suggested array can be written back the way it was read.
	optional map[string]bool
}

// visitCallbackWithDependencies is upstream's `visitFunctionWithDependencies`.
func visitCallbackWithDependencies(ctx rule.Context, run exhaustiveDepsRun, callback *ast.Node,
	dependencyArrayNode *ast.Node, hookCall *ast.Node, hookName string, componentScope *ast.Node,
	isEffect bool) {

	if isEffect && ast.HasSyntacticModifier(callback, ast.ModifierFlagsAsync) {
		ctx.ReportNode(callback, messageExhaustiveDepsAsyncEffect)
	}

	reads := gatherDependencies(ctx, callback, componentScope)

	isAutoDependenciesHook := run.autoDependenciesHooks[hookName]
	if dependencyArrayNode == nil {
		// A Hook the compiler supplies the array for is not the loop this finding describes.
		if isAutoDependenciesHook {
			return
		}
		reportSetStateWithoutDependencies(ctx, run, callback, hookCall, reads, componentScope)
		return
	}
	if isAutoDependenciesHook && ast.SkipParentheses(dependencyArrayNode).Kind == ast.KindNullKeyword {
		return
	}

	// A non-array dependency argument reports and then continues with an EMPTY declared list rather
	// than stopping. Upstream's own behaviour and it took a probe to see: `useEffect(fn, deps)` where
	// `deps` is a variable reports twice, once that the list is unreadable and once that everything
	// the callback reads is missing from it. Returning early here silences the second, which is the
	// finding that actually tells the reader what to write.
	//
	// The per-entry findings are COLLECTED here and reported below rather than as the array is read,
	// so that the finding about the array as a whole comes first. That is upstream's order and it is
	// observable: the summary's span starts at the `[` and an entry's starts inside it, so a reader
	// scanning a file top to bottom meets the summary first. Reporting as the entries are read
	// inverts it, which every id assertion in the fixtures is blind to and a span-ordered one is not.
	declared, entryFindings := readDeclaredDependencies(ctx, dependencyArrayNode, componentScope)

	recommendation := collectRecommendations(reads, declared, isEffect)

	if len(recommendation.missing) == 0 && len(recommendation.unnecessary) == 0 &&
		len(recommendation.duplicate) == 0 {
		reportEveryRenderConstructions(ctx, run, declared, componentScope, callback, dependencyArrayNode)
		reportEntryFindings(ctx, run, entryFindings)
		return
	}

	suggested := recommendation.suggested
	if !isEffect && len(recommendation.missing) > 0 {
		// A non-effect Hook recomputes the whole array rather than adding to what was declared, so
		// fixing a missing dependency also drops the unnecessary ones. Upstream's second
		// `collectRecommendations` call, guarded by `!isEffect`.
		suggested = collectRecommendations(reads, nil, isEffect).suggested
	}
	if declaredDependenciesAreSorted(declared) {
		sort.Strings(suggested)
	}

	message, count := firstNonEmptyWarning(recommendation)
	if count == 0 {
		reportEntryFindings(ctx, run, entryFindings)
		return
	}
	suggestedArray := formatDependencyArray(suggested, reads)
	if run.dangerousAutofix {
		ctx.ReportNodeWithFixes(dependencyArrayNode, message, rule.Fix{
			Range: rule.TokenRange(ctx.SourceFile, dependencyArrayNode),
			Text:  suggestedArray,
		})
	} else {
		ctx.ReportNodeWithSuggestions(dependencyArrayNode, message,
			rule.Suggestion{Message: rule.Message{
				Id:          "exhaustiveDepsUpdateArray",
				Description: "Update the dependency array to " + suggestedArray + ".",
			}})
	}
	reportEntryFindings(ctx, run, entryFindings)
}

// entryFinding is one thing wrong with a single dependency array entry, held until the finding
// about the whole array has been reported. See the call site for why the order matters.
type entryFinding struct {
	node    *ast.Node
	message rule.Message
	// fixes is the repair upstream offers for the entry, applied only under the dangerous option.
	fixes []rule.Fix
}

func reportEntryFindings(ctx rule.Context, run exhaustiveDepsRun, findings []entryFinding) {
	for _, finding := range findings {
		run.report(ctx, finding.node, finding.message, finding.fixes...)
	}
}

// firstNonEmptyWarning picks which of the three list findings to report, in upstream's order.
//
// One diagnostic per Hook call rather than one per bad dependency, and the order is missing, then
// unnecessary, then duplicate. A call with both a missing and an unnecessary dependency reports
// only about the missing one, and the suggested array fixes both.
func firstNonEmptyWarning(recommendation recommendations) (rule.Message, int) {
	if len(recommendation.missing) > 0 {
		return messageExhaustiveDepsMissing, len(recommendation.missing)
	}
	if len(recommendation.unnecessary) > 0 {
		return messageExhaustiveDepsUnnecessary, len(recommendation.unnecessary)
	}
	if len(recommendation.duplicate) > 0 {
		return messageExhaustiveDepsDuplicate, len(recommendation.duplicate)
	}
	return rule.Message{}, 0
}

// gatherDependencies walks the callback and records every property path it reads out of the
// component scope, keyed by path so a value read five times is one dependency.
//
// This is upstream's `gatherDependenciesRecursively` reached by a different route. Upstream iterates
// `scope.references` from `eslint-scope`; this walks the callback's own subtree and resolves each
// identifier through the checker, which answers the same question — see the rule's doc comment for
// the measurement that established the two classifications agree.
func gatherDependencies(ctx rule.Context, callback *ast.Node, componentScope *ast.Node) map[string]*dependencyRead {
	reads := map[string]*dependencyRead{}

	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier && !isNonDependencyIdentifier(node) {
			recordDependencyRead(ctx, node, callback, componentScope, reads)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	// The callback's own parameters are its own bindings, not dependencies, and walking the body
	// alone would also miss a default value that reads the component scope. Walking the whole node
	// and letting the scope classification decline the parameters is what upstream does, since a
	// parameter resolves to a declaration inside the callback.
	walk(callback)

	return reads
}

// recordDependencyRead resolves one identifier and, if it names a value from the component scope,
// files the property path it was read through.
func recordDependencyRead(ctx rule.Context, identifier *ast.Node, callback *ast.Node,
	componentScope *ast.Node, reads map[string]*dependencyRead) {

	symbol := resolveIdentifier(ctx, identifier)
	if symbol == nil || len(symbol.Declarations) == 0 {
		// An unresolved name is a global or a typo, and either way it does not change between
		// renders of this component. Upstream skips it for the same reason, and its corpus pins the
		// decision with a passing case whose callback reads `unresolved`.
		return
	}

	// A binding declared inside the callback is the callback's own and never a dependency. A binding
	// outside the component function belongs to a scope that does not re-render, so it cannot go
	// stale. Only the space between them counts, which is upstream's `pureScopes`.
	//
	// Every declaration is asked rather than only the first, because a symbol can carry more than
	// one and which one comes first is not a property this rule should depend on. The question here
	// is "is this binding anywhere in the callback", so any declaration inside it settles the case.
	inCallback, inComponent := false, false
	for _, declaration := range symbol.Declarations {
		if nodeContains(callback, declaration) {
			inCallback = true
			break
		}
		if nodeContains(componentScope, declaration) {
			inComponent = true
		}
	}
	if inCallback || !inComponent {
		return
	}

	// A reference to the binding the Hook call is itself being assigned to is not a dependency,
	// because at the moment the callback is built that binding holds nothing yet. Upstream's own
	// comment is "Ignore references to the function itself as it's not defined yet", and its clean
	// case is a recursive `useCallback`:
	//
	//	const recursive = useCallback((n) => (n <= 0 ? 0 : n + recursive(n - 1)), []);
	//
	// Without this the rule tells the reader to add `recursive` to its own dependency array, which
	// would rebuild the callback every time it rebuilds. The corpus is the only thing that says so.
	if declaresTheHookCallItself(symbol, callback) {
		return
	}

	chainRoot := outermostDependencyExpression(identifier)
	optional := map[string]bool{}
	path, readable := analyzePropertyChain(chainRoot, optional)
	if !readable {
		return
	}

	if existing, found := reads[path]; found {
		_ = existing
		return
	}
	reads[path] = &dependencyRead{
		path:     path,
		node:     chainRoot,
		stable:   isStableValue(ctx, symbol, callback, componentScope, map[*ast.Symbol]bool{}),
		optional: optional,
		order:    len(reads),
	}
}

// outermostDependencyExpression walks out from an identifier to the largest property path it is the
// root of, which is what actually gets declared as a dependency.
//
// Upstream's `getDependency`. Reading `props.foo.bar` produces the dependency `props.foo.bar` rather
// than `props`, because declaring the outer object is not required when only the inner value is
// read. Three things stop the walk, and each is a real decision:
//
//   - a computed access, `props[key]`, because the key is not statically known;
//   - a `.current`, because a ref's contents are the mutable part and depending on them is
//     meaningless;
//   - a CALL, `props.foo()`, and this one runs opposite to the intuitive reading. Calling a method
//     depends on the OBJECT rather than on the method: `props.toggleEditMode()` is a dependency on
//     `props`, not on `props.toggleEditMode`. The walk stops BEFORE taking the member expression,
//     so it keeps the object. Upstream's own reasoning is that a method reads the receiver, so the
//     receiver is what has to be current, and depending on the bound method alone would miss a
//     change to the object it closes over.
//
// Getting that direction backwards is silent rather than loud: the rule keeps reporting, it just
// names `props.toggleEditMode` where upstream names `props`, so the finding count matches and only
// the advice is wrong. Four upstream cases separate the two and all four are in the fixtures.
//
// A fourth shape has no walk at all: a member expression on the LEFT of an assignment,
// `props.foo = x`, yields the object. Writing to a property is not reading it.
func outermostDependencyExpression(identifier *ast.Node) *ast.Node {
	current := identifier
	for {
		parent := current.Parent
		if parent == nil {
			return current
		}
		// A parenthesis or a cast around the path is not part of the path.
		if isTransparentWrapper(parent) {
			current = parent
			continue
		}
		if parent.Kind != ast.KindPropertyAccessExpression {
			return assignmentTargetObject(current)
		}
		access := parent.AsPropertyAccessExpression()
		if access.Expression != current {
			return assignmentTargetObject(current)
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.Text() == "current" {
			return assignmentTargetObject(current)
		}
		// The call arm. The walk stops here and keeps what it already had, which for
		// `props.foo()` is `props`.
		if grandparent := parent.Parent; grandparent != nil &&
			grandparent.Kind == ast.KindCallExpression &&
			grandparent.AsCallExpression().Expression == parent {
			return current
		}
		current = parent
	}
}

// assignmentTargetObject unwraps a member expression that is being written to, because writing
// `props.foo = x` depends on `props` rather than on `props.foo`.
func assignmentTargetObject(node *ast.Node) *ast.Node {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return node
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return node
	}
	binary := parent.AsBinaryExpression()
	if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken ||
		binary.Left != node {
		return node
	}
	return node.AsPropertyAccessExpression().Expression
}

// analyzePropertyChain turns a member expression into the dotted path that names it, recording for
// each prefix whether that step was written with `?.`.
//
// Upstream's function of the same name, with one deliberate difference: upstream throws on any node
// kind it does not name, and the throw is what turns a non-null assertion or a cast in the
// dependency array into a "complex expression" finding. This unwraps those instead, for the reason
// recorded in the rule's doc comment. A computed access still fails, because `a[b]` genuinely
// cannot be compared as a path.
//
// The second result separates "this is not a path" from "this is the empty path", which the caller
// needs because an unreadable dependency array entry is its own finding.
func analyzePropertyChain(node *ast.Node, optional map[string]bool) (string, bool) {
	node = unwrapExpression(node)
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		name := node.Text()
		if optional != nil {
			optional[name] = false
		}
		return name, true

	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		objectPath, readable := analyzePropertyChain(access.Expression, optional)
		if !readable {
			return "", false
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "", false
		}
		path := objectPath + "." + name.Text()
		if optional != nil {
			// A prefix already marked optional stays optional: `a?.b.c` writes `?.` once and every
			// step after it is reached only when the first one was non-null, so re-writing the path
			// has to keep the `?.` where it was written.
			if access.QuestionDotToken != nil {
				if _, already := optional[path]; !already {
					optional[path] = true
				}
			} else {
				optional[path] = false
			}
		}
		return path, true
	}
	return "", false
}

// isTransparentWrapper reports whether a node adds no meaning to the expression inside it.
//
// A parenthesis, a `!`, an `as`, and a `satisfies` all denote the same value as their operand, so a
// path is the same path with or without them. Upstream unwraps the first three when reading the
// callback and does not when reading the dependency array; this rule unwraps consistently, which is
// the divergence recorded in the doc comment.
func isTransparentWrapper(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression,
		ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
		return true
	}
	return false
}

// unwrapExpression strips every transparent wrapper off an expression.
func unwrapExpression(node *ast.Node) *ast.Node {
	for node != nil && isTransparentWrapper(node) {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		}
	}
	return node
}

// dependencyTreeNode is one segment of a property path, and the three facts the descent reads.
//
// The tree is what makes the prefix asymmetry fall out instead of having to be stated: `props.a.b`
// and `props.a.b.c` share the `props.a.b` node, so marking that node satisfied stops the descent
// before it can reach `.c`, while marking `.c` satisfied leaves `props.a.b` unsatisfied above it.
type dependencyTreeNode struct {
	// isUsed is set when the callback reads exactly this path.
	isUsed bool

	// isSatisfiedRecursively is set when the dependency array declares this path, which covers
	// everything beneath it.
	isSatisfiedRecursively bool

	// isSubtreeUsed is set when the callback reads this path or anything under it. It is what
	// separates a declared dependency that is doing work from one that is unnecessary.
	isSubtreeUsed bool

	children map[string]*dependencyTreeNode
	order    []string
}

func newDependencyTreeNode() *dependencyTreeNode {
	return &dependencyTreeNode{children: map[string]*dependencyTreeNode{}}
}

// childForPath walks or creates the chain of nodes naming a dotted path.
func (node *dependencyTreeNode) childForPath(path string) *dependencyTreeNode {
	current := node
	for _, segment := range strings.Split(path, ".") {
		child, found := current.children[segment]
		if !found {
			child = newDependencyTreeNode()
			current.children[segment] = child
			current.order = append(current.order, segment)
		}
		current = child
	}
	return current
}

// markAncestorsSubtreeUsed marks every node along a path as having a read somewhere beneath it.
func (node *dependencyTreeNode) markAncestorsSubtreeUsed(path string) {
	current := node
	for _, segment := range strings.Split(path, ".") {
		child, found := current.children[segment]
		if !found {
			return
		}
		child.isSubtreeUsed = true
		current = child
	}
}

// declaredDependency is one entry written in the dependency array.
type declaredDependency struct {
	key  string
	node *ast.Node

	// external is set when the entry names a binding from outside the component, which changes what
	// the rule does with an unnecessary one: an outer-scope value in an EFFECT's array is kept in
	// the suggestion rather than dropped, because removing it would change when the effect runs for
	// a reason the reader did not ask for.
	external bool
}

// recommendations is what the tree descent concluded.
type recommendations struct {
	suggested   []string
	missing     []string
	unnecessary []string
	duplicate   []string
}

// collectRecommendations is upstream's function of the same name, reproduced node for node.
//
// The descent in `scanTree` is the satisfaction rule made mechanical: a satisfied node ends the
// descent and takes its subtree with it, an unsatisfied node that is itself read is missing, and
// anything else is a node nobody named directly so the answer is further down.
func collectRecommendations(reads map[string]*dependencyRead, declared []declaredDependency,
	isEffect bool) recommendations {

	tree := newDependencyTreeNode()

	// Reads are entered in the order the callback READ them, which is what upstream's insertion-
	// ordered map gives it for free. See `dependencyRead.order` for why sorting here is not a
	// harmless substitute.
	readPaths := make([]string, 0, len(reads))
	for path, read := range reads {
		if read.stable {
			continue
		}
		readPaths = append(readPaths, path)
	}
	sort.Slice(readPaths, func(first, second int) bool {
		return reads[readPaths[first]].order < reads[readPaths[second]].order
	})
	for _, path := range readPaths {
		tree.childForPath(path).isUsed = true
		tree.markAncestorsSubtreeUsed(path)
	}

	for _, dependency := range declared {
		tree.childForPath(dependency.key).isSatisfiedRecursively = true
	}
	// A stable value is treated as already satisfied rather than as unread, so declaring it is
	// allowed and omitting it is allowed. That is why `[]` is clean when the callback calls a
	// `useState` setter, and why naming the setter anyway is not an unnecessary dependency.
	stablePaths := make([]string, 0, len(reads))
	for path, read := range reads {
		if read.stable {
			stablePaths = append(stablePaths, path)
		}
	}
	sort.Slice(stablePaths, func(first, second int) bool {
		return reads[stablePaths[first]].order < reads[stablePaths[second]].order
	})
	for _, path := range stablePaths {
		tree.childForPath(path).isSatisfiedRecursively = true
		tree.markAncestorsSubtreeUsed(path)
	}

	// Missing paths are kept in DESCENT order rather than sorted, and the difference is observable.
	//
	// Upstream collects them into a JavaScript `Set`, which iterates in insertion order, and the
	// suggested array is only sorted later and only when the author's own array was already sorted.
	// Sorting here instead — which is what this did first, to make a Go map's iteration deterministic
	// — quietly sorts every suggestion, so an author who wrote `[props.c, props.b]` was told to write
	// `[props.a, props.b, props.c]` where React says `[props.c, props.b, props.a]`.
	//
	// No fixture in the imported corpus could see it, because every corpus case with more than one
	// missing dependency happens to be alphabetical already. It was found by a mutation that
	// SURVIVED: disabling the non-effect recompute changed the output on three shapes, and on all
	// three the mutant was the one that matched React. A survivor that turns out to be right about
	// the code is the most useful kind and this brief's categories do not name it.
	missing := []string{}
	missingSeen := map[string]bool{}
	satisfying := map[string]bool{}
	scanDependencyTree(tree, "", &missing, missingSeen, satisfying)

	result := recommendations{}
	seen := map[string]bool{}
	for _, dependency := range declared {
		if satisfying[dependency.key] {
			if !seen[dependency.key] {
				seen[dependency.key] = true
				result.suggested = append(result.suggested, dependency.key)
			} else {
				result.duplicate = appendOnce(result.duplicate, dependency.key)
			}
			continue
		}
		// An effect keeps a dependency it does not read, unless the entry is a ref's contents or an
		// outer-scope value. The reasoning upstream gives is that an effect's array is also how a
		// reader says "re-run when this changes", and honoring an unread entry is less surprising
		// than silently dropping it. `useMemo` and `useCallback` get no such benefit, because their
		// array is purely a cache key.
		if isEffect && !strings.HasSuffix(dependency.key, ".current") && !dependency.external {
			if !seen[dependency.key] {
				seen[dependency.key] = true
				result.suggested = append(result.suggested, dependency.key)
			}
			continue
		}
		result.unnecessary = appendOnce(result.unnecessary, dependency.key)
	}

	result.missing = missing
	result.suggested = append(result.suggested, missing...)

	return result
}

// scanDependencyTree is the descent that reads the satisfaction rule off the tree.
//
// Missing paths are appended in the order the descent reaches them; see the caller for why that
// order is load-bearing rather than incidental.
func scanDependencyTree(node *dependencyTreeNode, prefix string, missing *[]string,
	missingSeen map[string]bool, satisfying map[string]bool) {
	for _, segment := range node.order {
		child := node.children[segment]
		path := segment
		if prefix != "" {
			path = prefix + "." + segment
		}
		if child.isSatisfiedRecursively {
			// Satisfied ends the descent and covers everything beneath, which is the half of the
			// rule that makes `[props.a.b]` cover a read of `props.a.b.c`. It only counts as
			// SATISFYING something if the callback actually reads inside it; otherwise the entry is
			// declared and unused, and the caller calls that unnecessary.
			if child.isSubtreeUsed {
				satisfying[path] = true
			}
			continue
		}
		if child.isUsed {
			// Read here and not declared here or above. This is the other half of the rule: a
			// declaration further DOWN, at `props.a.b.c`, never reached this node, so it does not
			// stop the descent and the read at `props.a.b` is still missing.
			if !missingSeen[path] {
				missingSeen[path] = true
				*missing = append(*missing, path)
			}
			continue
		}
		scanDependencyTree(child, path, missing, missingSeen, satisfying)
	}
}

func appendOnce(list []string, value string) []string {
	for _, existing := range list {
		if existing == value {
			return list
		}
	}
	return append(list, value)
}

// readDeclaredDependencies reads the dependency array's entries and collects whatever cannot be
// read, without reporting: the caller reports these after the finding about the whole array, which
// is the order upstream's spans put them in.
//
// An argument that is not an array literal reports immediately and yields an empty list, so the
// caller goes on to compare the callback's reads against nothing and says what is missing. That one
// IS reported here rather than deferred, because it is about the argument rather than about an entry
// and its span is the same span the summary would use.
func readDeclaredDependencies(ctx rule.Context, dependencyArrayNode *ast.Node,
	componentScope *ast.Node) ([]declaredDependency, []entryFinding) {

	array := unwrapExpression(dependencyArrayNode)
	if array == nil || array.Kind != ast.KindArrayLiteralExpression {
		// Deferred with the entry findings rather than reported here, even though its span is the
		// argument rather than an entry inside it. Both findings then carry the SAME span, so source
		// position cannot order them and upstream's own order is the only tie-break: it reports what
		// is missing first and the unreadable list second.
		return nil, []entryFinding{{node: dependencyArrayNode, message: messageExhaustiveDepsNotArrayLiteral}}
	}

	declared := []declaredDependency{}
	findings := []entryFinding{}
	for _, element := range array.AsArrayLiteralExpression().Elements.Nodes {
		if element == nil || element.Kind == ast.KindOmittedExpression {
			// A hole, `[,,x,,]`, is not an entry. Upstream skips it and its corpus has a passing
			// case built entirely of holes.
			continue
		}
		if element.Kind == ast.KindSpreadElement {
			findings = append(findings, entryFinding{node: element, message: messageExhaustiveDepsSpread})
			continue
		}
		if isEffectEventResult(ctx, element, componentScope) {
			// Upstream's repair removes the entry's own range and nothing around it, so a separating
			// comma stays behind: `[a, onEvent]` becomes `[a, ]`.
			findings = append(findings, entryFinding{node: element, message: messageExhaustiveDepsEffectEvent,
				fixes: []rule.Fix{{Range: rule.TokenRange(ctx.SourceFile, element)}}})
			continue
		}
		key, readable := analyzePropertyChain(element, nil)
		if !readable {
			if isLiteralExpression(unwrapExpression(element)) {
				findings = append(findings, entryFinding{node: element, message: messageExhaustiveDepsLiteral})
			} else {
				findings = append(findings,
					entryFinding{node: element, message: messageExhaustiveDepsComplexExpression})
			}
			continue
		}
		declared = append(declared, declaredDependency{
			key:      key,
			node:     element,
			external: !isDeclaredInComponent(ctx, element, componentScope),
		})
	}
	return declared, findings
}

// isDeclaredInComponent reports whether a dependency array entry's root identifier binds to
// something inside the component function.
//
// Upstream asks the same question through `componentScope.through`, which is the set of references
// the component scope does not resolve itself.
func isDeclaredInComponent(ctx rule.Context, element *ast.Node, componentScope *ast.Node) bool {
	root := unwrapExpression(element)
	for root != nil && root.Kind == ast.KindPropertyAccessExpression {
		root = unwrapExpression(root.AsPropertyAccessExpression().Expression)
	}
	if root == nil || root.Kind != ast.KindIdentifier {
		return false
	}
	symbol := resolveIdentifier(ctx, root)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if nodeContains(componentScope, declaration) {
			return true
		}
	}
	return false
}

// declaredDependenciesAreSorted reports whether the author already wrote the array in order.
//
// The suggestion is only sorted when the answer is yes, so an intentional grouping survives being
// corrected while an accidental one gets tidied. Upstream's `areDeclaredDepsAlphabetized`.
func declaredDependenciesAreSorted(declared []declaredDependency) bool {
	keys := make([]string, len(declared))
	for index, dependency := range declared {
		keys[index] = dependency.key
	}
	return sort.StringsAreSorted(keys)
}

// formatDependencyArray writes a suggested array back out, restoring the optional chaining the
// paths were read through.
//
// A path read as `props.foo?.bar` has to be written back with the `?.` in the same place, because
// `props.foo.bar` would throw where the original did not.
func formatDependencyArray(paths []string, reads map[string]*dependencyRead) string {
	formatted := make([]string, 0, len(paths))
	for _, path := range paths {
		formatted = append(formatted, formatDependencyPath(path, reads))
	}
	return "[" + strings.Join(formatted, ", ") + "]"
}

func formatDependencyPath(path string, reads map[string]*dependencyRead) string {
	segments := strings.Split(path, ".")
	built := ""
	for index, segment := range segments {
		if index == 0 {
			built = segment
			continue
		}
		prefix := strings.Join(segments[:index+1], ".")
		separator := "."
		if optionalAtPrefix(prefix, reads) {
			separator = "?."
		}
		built += separator + segment
	}
	return built
}

// optionalAtPrefix asks every recorded read whether this exact prefix was reached through `?.`.
//
// Keyed by the full prefix rather than by the read that produced it, because two reads of the same
// object can spell the chain differently and upstream's map is shared across all of them: the first
// answer for a prefix is the one that sticks.
func optionalAtPrefix(prefix string, reads map[string]*dependencyRead) bool {
	for _, read := range reads {
		if optional, recorded := read.optional[prefix]; recorded && optional {
			return true
		}
	}
	return false
}

// isStableValue reports whether a binding's value is the same object on every render, so leaving it
// out of a dependency array cannot cause a stale read.
//
// Recognized by the NAME of the Hook that produced it rather than by its type, which is upstream's
// `isStableKnownHookValue`. The rule's doc comment records why the type checker is deliberately not
// used here even though it could answer more.
//
// Four shapes are stable:
//
//	const r = useRef(...)                  the container never changes, only `r.current` does
//	const [, setS] = useState(...)         the setter is guaranteed stable by React
//	const [, dispatch] = useReducer(...)   same guarantee
//	const [, start] = useTransition()      same guarantee
//	const x = useEffectEvent(...)          stable by construction, that is its purpose
//	const N = 1                            a const bound to a primitive literal cannot change
//
// The array-destructuring shapes are stable only in their SECOND position. `const [s, setS] =
// useState()` makes `setS` stable and `s` emphatically not: `s` is the value that changes, and
// treating it as stable would silence the most common real finding this rule makes.
//
// A function declared in the component body is stable when everything it reads is itself stable,
// which is upstream's `isFunctionWithoutCapturedValues`. That recursion is deliberately ONE level
// deep, matching React: oxc recurses without limit and treats a cycle as stable, which makes two
// mutually recursive component-body functions look stable when neither is. The measurement is in
// the rule's doc comment.
func isStableValue(ctx rule.Context, symbol *ast.Symbol, callback *ast.Node, componentScope *ast.Node,
	visiting map[*ast.Symbol]bool) bool {

	if visiting[symbol] {
		// Re-entry means a cycle, and a cycle is NOT stable: two functions that call each other are
		// both rebuilt on every render. This is the opposite of oxc's answer and it is the
		// divergence the doc comment records.
		return false
	}

	declaration := stableDeclarationOf(symbol)
	if declaration == nil {
		return false
	}

	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		return isStableVariableDeclaration(ctx, declaration, symbol, callback, componentScope, visiting)

	case ast.KindBindingElement:
		return isStableDestructuredHookResult(ctx, declaration, componentScope)

	case ast.KindFunctionDeclaration:
		visiting[symbol] = true
		defer delete(visiting, symbol)
		return functionReadsOnlyStableValues(ctx, declaration, componentScope, visiting)
	}
	return false
}

// stableDeclarationOf picks the declaration stability is decided from.
//
// The EARLIEST by source position rather than index zero, because a symbol can carry more than one
// declaration and the order `Declarations` arrives in is not something this rule should depend on.
// Upstream reads `defs[0]`, which for the shapes this rule sees is the value declaration, and the
// earliest-by-position choice reproduces that without inheriting the indexing assumption. A merged
// type declaration cannot be a stable Hook result, so a shape with several declarations is answered
// the same way either way.
func stableDeclarationOf(symbol *ast.Symbol) *ast.Node {
	var earliest *ast.Node
	for _, declaration := range symbol.Declarations {
		if earliest == nil || declaration.Pos() < earliest.Pos() {
			earliest = declaration
		}
	}
	return earliest
}

func isStableVariableDeclaration(ctx rule.Context, declaration *ast.Node, symbol *ast.Symbol,
	callback *ast.Node, componentScope *ast.Node, visiting map[*ast.Symbol]bool) bool {

	variable := declaration.AsVariableDeclaration()
	initializer := unwrapExpression(variable.Initializer)
	if initializer == nil {
		return false
	}

	// A function bound to a name in the component body is stable when everything it reads is.
	switch initializer.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		visiting[symbol] = true
		defer delete(visiting, symbol)
		return functionReadsOnlyStableValues(ctx, initializer, componentScope, visiting)
	}

	// A `const` bound to a primitive literal cannot change, so it never needs declaring. `let` is
	// excluded because it can be reassigned, and a regular expression literal is excluded because it
	// is a fresh object each evaluation.
	if isConstDeclaration(declaration) && isPrimitiveLiteral(initializer) {
		return true
	}

	if initializer.Kind != ast.KindCallExpression {
		return false
	}
	hookName := hookNameWithoutReactNamespace(initializer.AsCallExpression().Expression)
	return hookName == "useRef" || hookName == "useEffectEvent"
}

// isStableDestructuredHookResult answers the array-destructuring shapes, which are stable only in
// their second position.
//
// React's guarantee is about the value the Hook returns, not about the NAME it was bound to, so a
// setter that gets reassigned is no longer stable. Upstream tracks the binding's write count and
// gives up past one; this asks whether anything in the component assigns to the name, which is the
// same question for the shapes that can reach here — the destructuring itself is not an assignment
// expression, so any write found is a genuine reassignment. Upstream's own case for this is
// `let [count, setCount] = useState(0); setCount = unstableProp`, and without the check that reads
// as stable and the whole finding disappears.
func isStableDestructuredHookResult(ctx rule.Context, declaration *ast.Node,
	componentScope *ast.Node) bool {

	pattern := declaration.Parent
	if pattern == nil || pattern.Kind != ast.KindArrayBindingPattern {
		return false
	}
	elements := pattern.AsBindingPattern().Elements.Nodes
	if len(elements) != 2 || elements[1] != declaration {
		return false
	}

	variable := pattern.Parent
	if variable == nil || variable.Kind != ast.KindVariableDeclaration {
		return false
	}
	initializer := unwrapExpression(variable.AsVariableDeclaration().Initializer)
	if initializer == nil || initializer.Kind != ast.KindCallExpression {
		return false
	}
	switch hookNameWithoutReactNamespace(initializer.AsCallExpression().Expression) {
	case "useState", "useReducer", "useActionState", "useTransition":
		return !isReassignedInComponent(ctx, declaration, componentScope)
	}
	return false
}

// isReassignedInComponent reports whether anything in the component assigns to a binding.
//
// A walk of the component rather than a reference index, because the question is scoped to one
// function body and the binding is a local: every occurrence that could write to it is inside the
// node this rule already holds.
func isReassignedInComponent(ctx rule.Context, declaration *ast.Node, componentScope *ast.Node) bool {
	target := ctx.TypeChecker.GetSymbolAtLocation(declaration.Name())
	if target == nil {
		return false
	}
	reassigned := false
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || reassigned {
			return
		}
		if node.Kind == ast.KindBinaryExpression {
			binary := node.AsBinaryExpression()
			if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
				written := unwrapExpression(binary.Left)
				if written != nil && written.Kind == ast.KindIdentifier &&
					ctx.TypeChecker.GetSymbolAtLocation(written) == target {
					reassigned = true
					return
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(componentScope)
	return reassigned
}

// functionReadsOnlyStableValues reports whether every component-scope value a function reads is
// itself stable, so the function's identity changing does not matter.
func functionReadsOnlyStableValues(ctx rule.Context, function *ast.Node, componentScope *ast.Node,
	visiting map[*ast.Symbol]bool) bool {

	stable := true
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || !stable {
			return
		}
		if node.Kind == ast.KindIdentifier && !isNonDependencyIdentifier(node) {
			symbol := resolveIdentifier(ctx, node)
			if symbol != nil && len(symbol.Declarations) > 0 {
				inFunction, inComponent := false, false
				for _, declaration := range symbol.Declarations {
					if nodeContains(function, declaration) {
						inFunction = true
						break
					}
					if nodeContains(componentScope, declaration) {
						inComponent = true
					}
				}
				if !inFunction && inComponent &&
					!isStableValue(ctx, symbol, function, componentScope, visiting) {
					stable = false
					return
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(function)
	return stable
}

// reportEveryRenderConstructions reports a dependency that is correctly declared and still wrong,
// because the value it names is built fresh on every render.
//
// This runs only when the array is otherwise correct, which is upstream's ordering and matters: a
// call with a missing dependency reports about that instead, because fixing the array is the first
// step and this warning would be noise until then.
//
// The reason it is its own check rather than part of the comparison is that the array IS right by
// the comparison's standards. `const o = {}; useEffect(() => log(o), [o])` names exactly what it
// reads. The defect is one level down, in what `o` is: a new object each render, so the comparison
// against the previous render never matches and the effect runs every time.
func reportEveryRenderConstructions(ctx rule.Context, run exhaustiveDepsRun, declared []declaredDependency,
	componentScope *ast.Node, callback *ast.Node, dependencyArrayNode *ast.Node) {

	for _, dependency := range declared {
		// Only a bare name can be a construction. `props.foo` names a property of something the
		// component was handed, which this rule has no say over.
		if strings.Contains(dependency.key, ".") {
			continue
		}
		root := unwrapExpression(dependency.node)
		if root == nil || root.Kind != ast.KindIdentifier {
			continue
		}
		symbol := resolveIdentifier(ctx, root)
		if symbol == nil {
			continue
		}
		declaration := stableDeclarationOf(symbol)
		if declaration == nil || !nodeContains(componentScope, declaration) ||
			nodeContains(callback, declaration) {
			continue
		}
		if !isEveryRenderConstruction(declaration) {
			continue
		}
		run.report(ctx, declaration, messageExhaustiveDepsConstruction,
			wrapConstructionInUseCallback(ctx, declaration, symbol, componentScope, callback, dependencyArrayNode)...)
	}
}

// wrapConstructionInUseCallback is upstream's one construction repair: a variable whose initializer
// is a function, used somewhere other than the Hook, wrapped in its own useCallback. Upstream offers
// nothing for any other construction, so neither does this.
//
// One replacement of the initializer rather than an insertion on each side, because two fixes in one
// diagnostic are not applied together and half of the pair is broken code.
func wrapConstructionInUseCallback(ctx rule.Context, declaration *ast.Node, symbol *ast.Symbol,
	componentScope *ast.Node, callback *ast.Node, dependencyArrayNode *ast.Node) []rule.Fix {

	if declaration.Kind != ast.KindVariableDeclaration {
		return nil
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return nil
	}
	switch initializer.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
	default:
		return nil
	}
	if !isUsedOutsideOfHook(ctx, symbol, declaration, componentScope, callback, dependencyArrayNode) {
		return nil
	}
	initializerRange := rule.TokenRange(ctx.SourceFile, initializer)
	return []rule.Fix{{
		Range: initializerRange,
		Text:  "useCallback(" + ctx.SourceFile.Text()[initializerRange.Pos():initializerRange.End()] + ")",
	}}
}

// isUsedOutsideOfHook is upstream's question of the same name: is the binding read anywhere other
// than inside the Hook's callback or its dependency array, or written a second time after its
// declaration. Only then does wrapping it pay, since a value used only by the Hook belongs inside it.
func isUsedOutsideOfHook(ctx rule.Context, symbol *ast.Symbol, declaration *ast.Node, componentScope *ast.Node,
	callback *ast.Node, dependencyArrayNode *ast.Node) bool {

	used := false
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || used {
			return
		}
		if node.Kind == ast.KindIdentifier && node != declaration.Name() && !isNonDependencyIdentifier(node) &&
			resolveIdentifier(ctx, node) == symbol {
			parent := node.Parent
			written := parent != nil && parent.Kind == ast.KindBinaryExpression &&
				parent.AsBinaryExpression().Left == node && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind)
			if written || (!nodeContains(callback, node) && !nodeContains(dependencyArrayNode, node)) {
				used = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(componentScope)
	return used
}

// isEveryRenderConstruction reports whether a declaration binds a value that is rebuilt each render.
func isEveryRenderConstruction(declaration *ast.Node) bool {
	switch declaration.Kind {
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		return true
	case ast.KindVariableDeclaration:
		variable := declaration.AsVariableDeclaration()
		if variable.Name() == nil || variable.Name().Kind != ast.KindIdentifier {
			return false
		}
		return constructsANewValue(variable.Initializer)
	}
	return false
}

// constructsANewValue is upstream's `getConstructionExpressionType`, reduced to the yes-or-no this
// rule needs.
//
// Object, array, function, class, JSX, `new`, and a regular expression literal all evaluate to a
// fresh value every time the expression runs. The interesting part is the recursion: a ternary or a
// logical operator constructs if EITHER side does, because the branch that does is reachable, and an
// assignment constructs if its right side does. That recursion is what makes this semantic rather
// than a list of node kinds, and it is why `cond ? {} : x` is caught.
func constructsANewValue(node *ast.Node) bool {
	node = unwrapExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression,
		ast.KindNewExpression, ast.KindRegularExpressionLiteral,
		ast.KindJsxElement, ast.KindJsxFragment, ast.KindJsxSelfClosingElement:
		return true

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return constructsANewValue(conditional.WhenTrue) || constructsANewValue(conditional.WhenFalse)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return constructsANewValue(binary.Left) || constructsANewValue(binary.Right)
		case ast.KindEqualsToken:
			return constructsANewValue(binary.Right)
		}
	}
	return false
}

// reportSetStateWithoutDependencies reports an effect that calls a state setter and has no array, so
// it runs after every render and each run causes the next one.
func reportSetStateWithoutDependencies(ctx rule.Context, run exhaustiveDepsRun, callback *ast.Node,
	hookCall *ast.Node, reads map[string]*dependencyRead, componentScope *ast.Node) {

	found := false
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		// A setter mentioned inside a NESTED function does not count, because that function runs
		// when something calls it rather than on every render. Upstream asks whether the reference's
		// nearest function scope is the callback itself, and the corpus separates the two directly:
		// `useEffect(() => { setD(1); })` reports and `useEffect(() => { function g(){ setD(1); } })`
		// does not.
		if node != callback && ast.IsFunctionLike(node) {
			return
		}
		if node.Kind == ast.KindIdentifier && !isNonDependencyIdentifier(node) &&
			isStateSetterName(ctx, node, componentScope, callback) {
			// A MENTION rather than a call. `fetchData.then(setData)` hands the setter off to be
			// called later and upstream reports it, so testing for a call expression here would miss
			// two of its own cases.
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(callback)
	if !found {
		return
	}
	// Upstream's repair adds the array the effect needs right after the callback. It joins the
	// paths as they are, without the `?.` formatting the corrected array gets.
	suggested := collectRecommendations(reads, nil, true).suggested
	run.report(ctx, hookCall, messageExhaustiveDepsSetStateNoDependencies, rule.Fix{
		Range: core.NewTextRange(callback.End(), callback.End()),
		Text:  ", [" + strings.Join(suggested, ", ") + "]",
	})
}

// isStateSetterName reports whether an identifier names the setter half of a `useState` pair
// declared in the component body.
//
// The setter specifically, not the value: `const [s, setS] = useState()` makes `setS` a setter and
// `s` not one, and calling `s()` is not a state update.
func isStateSetterName(ctx rule.Context, identifier *ast.Node, componentScope *ast.Node,
	callback *ast.Node) bool {

	symbol := resolveIdentifier(ctx, identifier)
	if symbol == nil {
		return false
	}
	declaration := stableDeclarationOf(symbol)
	if declaration == nil || declaration.Kind != ast.KindBindingElement {
		return false
	}
	if !nodeContains(componentScope, declaration) || nodeContains(callback, declaration) {
		return false
	}
	pattern := declaration.Parent
	if pattern == nil || pattern.Kind != ast.KindArrayBindingPattern {
		return false
	}
	elements := pattern.AsBindingPattern().Elements.Nodes
	if len(elements) != 2 || elements[1] != declaration {
		return false
	}
	variable := pattern.Parent
	if variable == nil || variable.Kind != ast.KindVariableDeclaration {
		return false
	}
	initializer := unwrapExpression(variable.AsVariableDeclaration().Initializer)
	if initializer == nil || initializer.Kind != ast.KindCallExpression {
		return false
	}
	return hookNameWithoutReactNamespace(initializer.AsCallExpression().Expression) == "useState"
}

// hookNameWithoutReactNamespace reads a callee's Hook name, with or without the `React.` prefix.
//
// `React.useEffect(...)` and `useEffect(...)` are the same Hook and both implementations say so.
// A deeper namespace is not: `a.b.useEffect` is nobody's React import.
func hookNameWithoutReactNamespace(callee *ast.Node) string {
	callee = unwrapExpression(callee)
	if callee == nil {
		return ""
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text()

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		object := unwrapExpression(access.Expression)
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "React" {
			return ""
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return ""
		}
		return name.Text()
	}
	return ""
}

// enclosingFunctionOfHookCall returns the function the Hook is called from, which is the component
// scope every dependency is measured against.
func enclosingFunctionOfHookCall(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLike(current) {
			return current
		}
	}
	return nil
}

// declaresTheHookCallItself reports whether a symbol is the variable the Hook call is being
// assigned to, so a reference to it from inside the callback is a self-reference.
//
// The callback's parent is the Hook call; the Hook call's parent is the variable declaration.
func declaresTheHookCallItself(symbol *ast.Symbol, callback *ast.Node) bool {
	hookCall := callback.Parent
	if hookCall == nil || hookCall.Kind != ast.KindCallExpression {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		if unwrapExpression(declaration.AsVariableDeclaration().Initializer) == hookCall {
			return true
		}
	}
	return false
}

// nodeContains reports whether a node is the ancestor of another, or is it.
func nodeContains(ancestor *ast.Node, node *ast.Node) bool {
	if ancestor == nil || node == nil {
		return false
	}
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

// resolveIdentifier reports the binding an identifier occurrence refers to.
//
// The shorthand-property case needs its own accessor, for the reason
// `internal/rules/core/no_useless_assignment.go` measured on real code: in `f({ v })` the identifier
// resolves through `GetSymbolAtLocation` to the PROPERTY's symbol rather than to the variable it
// reads, so the read is filed under a symbol nothing else touches. Here that would lose a
// dependency rather than invent one, which is the direction that hurts: a value read only through a
// shorthand would look unread and the rule would call declaring it unnecessary.
func resolveIdentifier(ctx rule.Context, identifier *ast.Node) *ast.Symbol {
	if parent := identifier.Parent; parent != nil &&
		parent.Kind == ast.KindShorthandPropertyAssignment {
		if symbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent); symbol != nil {
			return symbol
		}
	}
	return ctx.TypeChecker.GetSymbolAtLocation(identifier)
}

// isNonDependencyIdentifier reports whether an identifier occurrence cannot be a value read.
//
// A property NAME shares its spelling with a variable and resolves to something else entirely, so
// `o.v` reads `o` and merely names `v`. A shorthand property is deliberately absent: it looks like a
// name and is a real read, which is what `resolveIdentifier` exists to handle.
//
// The type positions matter more here than they would in a JavaScript-only rule. `useEffect(() => {
// const x: Foo = ...; }, [])` must not depend on `Foo`, and typescript-go resolves a type reference
// to a real symbol, so without this the rule would demand types in dependency arrays.
func isNonDependencyIdentifier(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}
	// The type walk comes FIRST, and the ordering is the whole reason this function has a comment.
	//
	// It was written second, after the name-position switch, and three upstream clean cases went red:
	// `Wrap<target>[]`, `foo<typeof target>()`, and `'a' as keyof typeof target`. In all three the
	// identifier sits inside a type node whose kind is ALSO in the switch below — `KindTypeReference`
	// and `KindTypeQuery` — and the switch answers `parent.Name() == identifier`, which is false for
	// a type ARGUMENT or for the operand of a `typeof`. So the switch returned false and the walk
	// below was never reached, and the rule demanded a type be listed in a dependency array.
	//
	// This is the direction that hurts. A rule telling someone to add `VirtuosoItem` to their
	// dependency array is advice that cannot be followed, and the corpus caught it only because
	// upstream had written the cases.
	for current := identifier.Parent; current != nil; current = current.Parent {
		if ast.IsTypeNode(current) {
			return true
		}
		if ast.IsFunctionLike(current) || current.Kind == ast.KindSourceFile {
			break
		}
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Name() == identifier
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Name() == identifier
	case ast.KindQualifiedName, ast.KindMethodDeclaration, ast.KindPropertyDeclaration,
		ast.KindPropertySignature, ast.KindMethodSignature, ast.KindGetAccessor,
		ast.KindSetAccessor, ast.KindEnumMember, ast.KindImportSpecifier,
		ast.KindImportClause, ast.KindNamespaceImport, ast.KindTypeParameter,
		ast.KindTypeReference, ast.KindTypeQuery, ast.KindJsxAttribute:
		return parent.Name() == identifier
	}
	return false
}

// isLiteralExpression reports whether a dependency array entry is a literal, which is its own
// finding rather than the generic complex-expression one.
func isLiteralExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

func isPrimitiveLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNullKeyword,
		ast.KindNoSubstitutionTemplateLiteral:
		return true
	}
	return false
}

// isConstDeclaration reports whether a variable declaration was written with `const`.
func isConstDeclaration(declaration *ast.Node) bool {
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return false
	}
	return list.AsVariableDeclarationList().Flags&ast.NodeFlagsConst != 0
}

// isEffectEventResult reports whether a dependency array entry names a value that came out of
// `useEffectEvent`.
//
// React ships that Hook specifically so a function can be excluded from a dependency array: it
// always sees the latest render's values, so listing it re-runs the effect for a change that could
// never have gone stale. The finding is the inverse of every other one in this rule, which is why
// it is checked before the entry is treated as an ordinary dependency.
func isEffectEventResult(ctx rule.Context, element *ast.Node, componentScope *ast.Node) bool {
	root := unwrapExpression(element)
	if root == nil || root.Kind != ast.KindIdentifier {
		return false
	}
	symbol := resolveIdentifier(ctx, root)
	if symbol == nil {
		return false
	}
	declaration := stableDeclarationOf(symbol)
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration ||
		!nodeContains(componentScope, declaration) {
		return false
	}
	initializer := unwrapExpression(declaration.AsVariableDeclaration().Initializer)
	if initializer == nil || initializer.Kind != ast.KindCallExpression {
		return false
	}
	return hookNameWithoutReactNamespace(initializer.AsCallExpression().Expression) == "useEffectEvent"
}

// dependencyArrayNamesIdentifier reports whether the array already lists a bare name.
//
// The case is a Hook whose callback is passed by name and whose array names that same function, so
// the callback is already declared as its own dependency and upstream stops there rather than trying
// to analyze a function it can see is being tracked.
func dependencyArrayNamesIdentifier(dependencyArrayNode *ast.Node, name string) bool {
	array := unwrapExpression(dependencyArrayNode)
	if array == nil || array.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	for _, element := range array.AsArrayLiteralExpression().Elements.Nodes {
		candidate := unwrapExpression(element)
		if candidate != nil && candidate.Kind == ast.KindIdentifier && candidate.Text() == name {
			return true
		}
	}
	return false
}

// callbackResolution says how far following a named callback got.
//
// Three answers rather than two, because upstream reports differently for each. A name bound to a
// function written here is analyzed in place; a name bound to something else in the component is a
// missing dependency naming the callback; and a name that leads nowhere readable is the unknown
// case where nothing can be said about the array.
type callbackResolution int

const (
	callbackResolutionFound callbackResolution = iota
	callbackResolutionNotAFunction
	callbackResolutionUnknown
)

// resolvedFunctionBody follows a callback passed by name to the function it names.
func resolvedFunctionBody(ctx rule.Context, identifier *ast.Node) (*ast.Node, callbackResolution) {
	symbol := resolveIdentifier(ctx, identifier)
	if symbol == nil {
		return nil, callbackResolutionUnknown
	}
	declaration := stableDeclarationOf(symbol)
	if declaration == nil {
		return nil, callbackResolutionUnknown
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return declaration, callbackResolutionFound
	case ast.KindParameter:
		// A callback arriving as a parameter cannot be seen at all.
		return nil, callbackResolutionUnknown
	case ast.KindVariableDeclaration:
		initializer := unwrapExpression(declaration.AsVariableDeclaration().Initializer)
		if initializer == nil {
			return nil, callbackResolutionUnknown
		}
		switch initializer.Kind {
		case ast.KindArrowFunction, ast.KindFunctionExpression:
			return initializer, callbackResolutionFound
		}
		return nil, callbackResolutionNotAFunction
	}
	return nil, callbackResolutionUnknown
}
