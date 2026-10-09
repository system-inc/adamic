package react_conformance

// This file answers a question the goldens do not answer for themselves: which upstream rule owns
// each expected diagnostic.
//
// # Why an attribution layer has to exist at all
//
// The `## Error` block prints a heading and a message. It does not print the `ErrorCategory` that
// produced them, and the category is the only thing that names a rule. So the obvious join — match
// a fixture to a rule by its filename — is not available, and the two plausible substitutes are
// both wrong in ways that read as clean results.
//
// Matching on the `error.` filename is wrong because the prefix names the *fixture author's*
// intent, not the validator that ran. Measured at the pinned sha: all three fixtures whose names
// say error-boundaries (`error.todo-invalid-jsx-in-try-with-finally` and its two neighbours) expect
// a `Todo` from `BuildHIR::lowerStatement` instead, because React's lowering aborts on a `try` with
// no `catch` before the ErrorBoundaries validator is ever reached. A filename join scores
// error-boundaries against three fixtures that cannot exercise it.
//
// Guessing from the message text is wrong because messages are shared across categories and are
// assembled by interpolation. `Found missing/extra memoization dependencies` is built at
// `ValidateExhaustiveDependencies.ts:1078` by joining two words into a template, so it appears in
// no source file as a literal string and a grep for it returns nothing.
//
// # Where this map comes from
//
// Upstream's own emission sites, read at the pinned sha. Every diagnostic is constructed with an
// explicit `category: ErrorCategory.X`, or through one of four `CompilerError` helpers that hardcode
// one (`invariant` → Invariant, `throwTodo` → Todo, `throwInvalidJS` → Syntax, `throwInvalidConfig`
// → Config). The category is then mapped to a rule name by `getRuleForCategory`, whose switch is
// closed by `assertExhaustive`, so the rule list cannot drift without upstream failing to compile.
//
// # How this map was checked, rather than trusted
//
// Two independent cross-checks, both of which had to hold before the numbers below were written.
//
// First, coverage: the map attributes all 429 diagnostics with none left over. An earlier revision
// left 110 unattributed and then 79 and then 53, each time because a message was built by
// interpolation or reached through a `const` indirection. A map that covered 376 of 429 would have
// produced a per-rule table that looked entirely reasonable, which is why the residual is asserted
// at zero rather than eyeballed.
//
// Second, and this is the load-bearing one: the heading. Upstream's printer assigns exactly one of
// four headings per category (`CompilerError.ts:565`). Running every attributed category through
// that switch reproduces the heading the golden actually printed, for all 429, with zero
// mismatches. That check was not used to build the map, so it is genuine evidence: a
// misattribution between two categories with different headings could not survive it.
//
// It cannot catch a swap between two categories that share a heading (Globals and Refs both print
// `Error`), so it is a necessary rather than sufficient check, and that limit is stated here rather
// than left for a reader to discover.

// CategoryForMessage returns the upstream ErrorCategory that produces a diagnostic message.
//
// Keyed on the message's first line, which is what the golden prints after the heading and what
// upstream's own backend comparison asserts on.
func CategoryForMessage(message string) (string, bool) {
	if category, found := categoryByMessage[message]; found {
		return category, true
	}
	for _, pattern := range categoryByMessagePrefix {
		if len(message) >= len(pattern.Prefix) && message[:len(pattern.Prefix)] == pattern.Prefix {
			return pattern.Category, true
		}
	}
	return "", false
}

// messagePrefix matches the interpolated messages, which have no literal form to key on.
type messagePrefix struct {
	Prefix   string
	Category string
}

// categoryByMessagePrefix covers the diagnostics upstream assembles from a template.
//
// These cannot be exact-matched because the variable part is a node type, a binding name, or a
// joined word. The prefix is the invariant half, taken from the template in upstream's source
// rather than from the corpus, so a fixture using a spelling the corpus does not currently contain
// still attributes.
var categoryByMessagePrefix = []messagePrefix{
	{"(BuildHIR::node.lowerReorderableExpression) Expression type ", "Todo"},
	{"(BuildHIR::lowerExpression) Handle ", "Todo"},
	{"(BuildHIR::lowerStatement) Handle ", "Todo"},
	{"(BuildHIR::lowerStatement) Support ", "Todo"},
	{"[FindContextIdentifiers] Cannot handle Object destructuring assignment target ", "Todo"},
}

// RuleForCategory maps an upstream ErrorCategory to the lint rule name upstream gives it.
//
// Transcribed from `getRuleForCategoryImpl` in `CompilerError.ts` at the pinned sha, which is the
// function ESLint itself uses to decide which rule name a diagnostic is reported under. Kept as
// data rather than folded into the message map because the two change for different reasons: a
// message can be reworded without moving rules, and a category can be renamed without touching a
// message.
func RuleForCategory(category string) (string, bool) {
	rule, found := ruleByCategory[category]
	return rule, found
}

var ruleByCategory = map[string]string{
	"CapitalizedCalls":             "capitalized-calls",
	"Config":                       "config",
	"EffectDependencies":           "memoized-effect-dependencies",
	"EffectDerivationsOfState":     "no-deriving-state-in-effects",
	"EffectExhaustiveDependencies": "exhaustive-effect-dependencies",
	"EffectSetState":               "set-state-in-effect",
	"ErrorBoundaries":              "error-boundaries",
	"FBT":                          "fbt",
	"Gating":                       "gating",
	"Globals":                      "globals",
	"Hooks":                        "hooks",
	"Immutability":                 "immutability",
	"IncompatibleLibrary":          "incompatible-library",
	"Invariant":                    "invariant",
	"MemoDependencies":             "memo-dependencies",
	"PreserveManualMemo":           "preserve-manual-memoization",
	"Purity":                       "purity",
	"Refs":                         "refs",
	"RenderSetState":               "set-state-in-render",
	"StaticComponents":             "static-components",
	"Suppression":                  "rule-suppression",
	"Syntax":                       "syntax",
	"Todo":                         "todo",
	"UnsupportedSyntax":            "unsupported-syntax",
	"UseMemo":                      "use-memo",
	"VoidUseMemo":                  "void-use-memo",
}

// HeadingForCategory returns the severity word upstream's printer prints for a category.
//
// Transcribed from the switch in `printErrorSummary` (`CompilerError.ts:565`), which closes with
// `assertExhaustive` over every `ErrorCategory`. This exists to be run against the corpus as a
// cross-check: the heading is recorded independently in each golden, so a category attributed
// wrongly across a heading boundary produces a mismatch that nothing else in this package would
// notice.
func HeadingForCategory(category string) (string, bool) {
	heading, found := headingByCategory[category]
	return heading, found
}

var headingByCategory = map[string]string{
	"CapitalizedCalls":             "Error",
	"Config":                       "Error",
	"EffectDerivationsOfState":     "Error",
	"EffectSetState":               "Error",
	"ErrorBoundaries":              "Error",
	"FBT":                          "Error",
	"Gating":                       "Error",
	"Globals":                      "Error",
	"Hooks":                        "Error",
	"Immutability":                 "Error",
	"Purity":                       "Error",
	"Refs":                         "Error",
	"RenderSetState":               "Error",
	"StaticComponents":             "Error",
	"Suppression":                  "Error",
	"Syntax":                       "Error",
	"UseMemo":                      "Error",
	"VoidUseMemo":                  "Error",
	"MemoDependencies":             "Error",
	"EffectExhaustiveDependencies": "Error",

	"EffectDependencies":  "Compilation Skipped",
	"IncompatibleLibrary": "Compilation Skipped",
	"PreserveManualMemo":  "Compilation Skipped",
	"UnsupportedSyntax":   "Compilation Skipped",

	"Invariant": "Invariant",
	"Todo":      "Todo",
}

// Categories returns the categories a fixture's diagnostics belong to, and whether every one was
// attributed.
//
// The bool is not decoration. A fixture with an unattributed diagnostic must not be scored against
// any rule, because the rule that owns it is exactly the unknown; scoring it would credit or blame
// whichever rule happened to be running.
func (f Fixture) Categories() (categories []string, complete bool) {
	seen := map[string]bool{}
	complete = true
	for _, expectedError := range f.Expected.Errors {
		category, found := CategoryForMessage(expectedError.Message)
		if !found {
			complete = false
			continue
		}
		if !seen[category] {
			seen[category] = true
			categories = append(categories, category)
		}
	}
	return categories, complete
}

// Rules returns the upstream rule names a fixture's diagnostics belong to.
func (f Fixture) Rules() (rules []string, complete bool) {
	categories, complete := f.Categories()
	seen := map[string]bool{}
	for _, category := range categories {
		rule, found := RuleForCategory(category)
		if !found {
			complete = false
			continue
		}
		if !seen[rule] {
			seen[rule] = true
			rules = append(rules, rule)
		}
	}
	return rules, complete
}

// categoryByMessage maps a diagnostic's first line to the ErrorCategory that produced it.
//
// Every entry was read from an emission site in upstream's source at the pinned sha, never inferred
// from the corpus. The distinction matters: inferring from the corpus would produce a map that
// agrees with whatever the 325 happen to contain, and a message whose rule has no error fixture
// would simply be absent with nothing to say so.
//
// The set is exactly the 58 literal messages the corpus uses, plus the five templated forms in
// categoryByMessagePrefix. Restricting it to what the corpus uses is deliberate for the same reason
// knownPragmas is: pre-blessing messages nobody has looked at is silence with extra steps.
var categoryByMessage = map[string]string{

	// CapitalizedCalls
	"Capitalized functions are reserved for components, which must be invoked with JSX. If this is a component, render it with JSX. Otherwise, ensure that it has no hook calls and rename it to begin with a lowercase letter. Alternatively, if you know for a fact that this function is not a component, you can allowlist it via the compiler config": "CapitalizedCalls",

	// Config
	"Invalid type configuration for module": "Config",

	// EffectDerivationsOfState
	"Values derived from props and state should be calculated during render, not in an effect. (https://react.dev/learn/you-might-not-need-an-effect#updating-state-based-on-props-or-state)": "EffectDerivationsOfState",

	// EffectExhaustiveDependencies
	"Found extra effect dependencies":         "EffectExhaustiveDependencies",
	"Found missing effect dependencies":       "EffectExhaustiveDependencies",
	"Found missing/extra effect dependencies": "EffectExhaustiveDependencies",

	// Gating
	"Dynamic gating directive is not a valid JavaScript identifier": "Gating",

	// Globals
	"Cannot reassign variables declared outside of the component/hook": "Globals",

	// Hooks
	"Hooks may not be referenced as normal values, they must be called. See https://react.dev/reference/rules/react-calls-components-and-hooks#never-pass-around-hooks-as-regular-values":                                        "Hooks",
	"Hooks must always be called in a consistent order, and may not be called conditionally. See the Rules of Hooks (https://react.dev/warnings/invalid-hook-call-warning)":                                                      "Hooks",
	"Hooks must be called at the top level in the body of a function component or custom hook, and may not be called within function expressions. See the Rules of Hooks (https://react.dev/warnings/invalid-hook-call-warning)": "Hooks",
	"Hooks must be the same function on every render, but this value may change over time to a different function. See https://react.dev/reference/rules/react-calls-components-and-hooks#dont-dynamically-use-hooks":            "Hooks",

	// Immutability
	"Cannot access variable before it is declared":         "Immutability",
	"Cannot modify local variables after render completes": "Immutability",
	"Cannot reassign variable after render completes":      "Immutability",
	"Cannot reassign variable in async function":           "Immutability",
	"This value cannot be modified":                        "Immutability",

	// IncompatibleLibrary
	"Use of incompatible library": "IncompatibleLibrary",

	// Invariant
	"(BuildHIR::lowerAssignment) Could not find binding for declaration.":                                "Invariant",
	"Const declaration cannot be referenced as an expression":                                            "Invariant",
	"Expected a variable declaration":                                                                    "Invariant",
	"Expected all references to a variable to be consistently local or context references":               "Invariant",
	"Expected consistent kind for destructuring":                                                         "Invariant",
	"Expected temporaries to be promoted to named identifiers in an earlier pass":                        "Invariant",
	"Unexpected empty block with `goto` terminal":                                                        "Invariant",
	"[Codegen] Internal error: MethodCall::property must be an unpromoted + unmemoized MemberExpression": "Invariant",
	"[InferMutationAliasingEffects] Expected value kind to be initialized":                               "Invariant",

	// MemoDependencies
	"Found extra memoization dependencies":         "MemoDependencies",
	"Found missing memoization dependencies":       "MemoDependencies",
	"Found missing/extra memoization dependencies": "MemoDependencies",

	// PreserveManualMemo
	"Existing memoization could not be preserved": "PreserveManualMemo",

	// Purity
	"Cannot call impure function during render": "Purity",

	// Refs
	"Cannot access refs during render": "Refs",

	// RenderSetState
	"Calling setState from useMemo may trigger an infinite loop": "RenderSetState",
	"Cannot call setState during render":                         "RenderSetState",

	// Suppression
	"React Compiler has skipped optimizing this component because one or more React ESLint rules were disabled":            "Suppression",
	"React Compiler has skipped optimizing this component because one or more React rule violations were reported by Flow": "Suppression",

	// Syntax
	"Cannot reassign a `const` variable":      "Syntax",
	"Expected a non-reserved identifier name": "Syntax",

	// Todo
	"(BuildHIR::lowerExpression) Support UpdateExpression where argument is a global": "Todo",
	"Bailing out due to blocklisted import":                                           "Todo",
	"Important source location has wrong node type in generated code":                 "Todo",
	"Important source location missing in generated code":                             "Todo",
	"Support destructuring of context variables":                                      "Todo",
	"Support duplicate fbt tags":                                                      "Todo",
	"Support functions with unreachable code that may contain hoisted declarations":   "Todo",
	"Support local variables named `fbt`":                                             "Todo",
	"Support non-trivial for..in inits":                                               "Todo",
	"Support non-trivial for..of inits":                                               "Todo",
	"Support spread syntax for hook arguments":                                        "Todo",
	"[PruneHoistedContexts] Rewrite hoisted function references":                      "Todo",
	"[hoisting] EnterSSA: Expected identifier to be defined before being used":        "Todo",

	// UnsupportedSyntax
	"The 'eval' function is not supported": "UnsupportedSyntax",

	// UseMemo
	"Expected the dependency list for useMemo to be an array literal":                 "UseMemo",
	"Expected the first argument to be an inline function expression":                 "UseMemo",
	"useMemo() callbacks may not accept parameters":                                   "UseMemo",
	"useMemo() callbacks may not be async or generator functions":                     "UseMemo",
	"useMemo() callbacks may not reassign variables declared outside of the callback": "UseMemo"}
