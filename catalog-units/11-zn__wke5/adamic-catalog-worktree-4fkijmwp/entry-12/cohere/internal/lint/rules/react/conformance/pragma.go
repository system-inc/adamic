package react_conformance

import (
	"fmt"
	"sort"
	"strings"

	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// UpstreamSha is the facebook/react commit the vendored corpus was taken from.
//
// The `## Error` expectations are prose. They carry react.dev URLs and multi-paragraph guidance,
// and they change without notice, so a corpus that tracks `main` would rewrite the scoreboard under
// the implementation and call the movement progress. Pinning makes an upstream change a diff a
// human reads rather than a number that drifts.
const UpstreamSha = "bd6ea412c6732b3b946a2827fcaac3a1c8f2e863"

// UpstreamFixtureDirectory is where the vendored files came from, relative to the repository root.
const UpstreamFixtureDirectory = "compiler/packages/babel-plugin-react-compiler/src/__tests__/fixtures/compiler"

// ExpectedErrorFixtureCount is how many `error.*` and `todo.error.*` fixtures the corpus holds.
//
// The number that moved last time and the reason the vendoring tool refuses a path-anchored filter:
// a research pass reported 221 by grepping the fixture root, which missed the nine subdirectories
// holding 103 of them. Wrong by a third, and it read exactly like a clean result.
const ExpectedErrorFixtureCount = 325

// ExpectedCleanFixtureCount is how many fixtures carrying the preserve-memoization pragma expect no
// error at all.
//
// These are the only population where silence is the correct answer and a finding is a defect, which
// makes them the sole false-positive oracle in the corpus. Their absence was not upstream's doing:
// `Load` rejected every fixture not named `error.*`, so the tree was error-only by construction and
// the resulting inability to measure over-reporting was attributed to upstream for hours.
//
// Asserted exactly rather than as a floor, because the failure that produced the wrong number could
// not be caught by one. Collecting these by copying them into a single directory loses any whose
// basenames collide, and four do -- `optional-member-expression-single.js` and three siblings exist
// in more than one subdirectory. That collection returned 66, wrote no error, and the four losses
// were invisible.
const ExpectedCleanFixtureCount = 70

// ExpectedFixtureCount is the whole corpus, both populations.
//
// Spelled as the sum rather than as a literal, so the two populations cannot silently trade against
// each other: a run that lost four clean fixtures and gained four error ones would hold a correct
// total and a wrong corpus.
const ExpectedFixtureCount = ExpectedErrorFixtureCount + ExpectedCleanFixtureCount

// ExpectedFlowFixtureCount is how many fixtures need a Flow parser, across both populations.
//
// 35 error-named and 5 clean. They are declined rather than scored, which is a stated exclusion
// rather than a skip -- see the `flow` entry in ignoredPragmas.
const ExpectedFlowFixtureCount = 40

// Pragma is one `@key` or `@key:value` directive from a fixture's configuration comment.
type Pragma struct {
	Key string

	// Value is the text after the first `:`, still in its source spelling: `"infer"` keeps its
	// quotes, `false` stays a string. Nothing here interprets it, because nothing here yet has a
	// configuration to apply it to, and inventing a decoder for a consumer that does not exist is
	// how a wrong interpretation gets baked in before anything can contradict it.
	Value string

	// HasValue distinguishes `@enableSomething` from `@enableSomething:false`. Without it the two
	// collapse into the same zero Value and the negation reads as the assertion.
	HasValue bool
}

// String renders a Pragma back into its source spelling, for error messages.
func (p Pragma) String() string {
	if p.HasValue {
		return "@" + p.Key + ":" + p.Value
	}
	return "@" + p.Key
}

// UnknownPragmaError reports a directive the runner does not model.
//
// This is an error rather than a skip on purpose, and the reason is the whole point of the type.
//
// A pragma reconfigures what a fixture expects. `@validateRefAccessDuringRender` turns a validation
// on; without it the same source is legal and the expectation is wrong. So an unmodelled directive
// does not make a fixture merely unrunnable, it makes the fixture's verdict *unrelated to the
// implementation being scored* — and a runner that skips it reports "not applicable" for a case it
// silently got wrong, while a runner that ignores it reports a pass or a fail it has not earned.
//
// Upstream demonstrates the failure rather than merely permitting it. `parseConfigPragmaForTests`
// in `Utils/TestUtils.ts` reads:
//
//	for (const {key, value: val} of splitPragma(pragma)) {
//	  if (!hasOwnProperty(defaultOptions, key)) {
//	    continue;
//	  }
//
// and `parseConfigPragmaEnvironmentForTest` has the identical `continue` against
// `EnvironmentConfigSchema.shape`. Both drop an unrecognised key without a word. Measured against
// the 325 error fixtures at the pinned sha, six distinct directives in active use fall through
// those two `continue`s: `@skip`, `@flow`, `@enableFlowSuppressions`, `@enableNewMutationAliasingModel`,
// `@enablePropagateDepsInHIR`, and `@compilationMode(infer)` — the last because `splitPragma` splits
// values on `:` only, so the parenthesised spelling yields the key `compilationMode(infer)`, which
// matches nothing. `@skip` is the sharpest of them: it is not read anywhere in `packages/snap`, so
// eight fixtures carry an instruction that has never done anything.
//
// Refusing is cheap here and silence is not. A refusal is one line naming the directive and the
// file; a silent skip is a green suite over a corpus a third of which was never scored.
type UnknownPragmaError struct {
	FixturePath string
	Pragma      Pragma
}

func (e *UnknownPragmaError) Error() string {
	return fmt.Sprintf(
		"%s: unrecognised pragma %s; the runner refuses rather than skipping, because an unmodelled directive changes what the fixture expects and a skipped fixture that reports as a pass is exactly the defect this suite exists to catch. Add it to knownPragmas with a note on what it means, or record it in ignoredPragmas with the evidence that it does not affect diagnostics",
		e.FixturePath, e.Pragma,
	)
}

// knownPragmas are the directives the runner models, mapped to why the runner is allowed to know
// about them.
//
// Membership here is a claim, not a formality: it says a human read the directive and decided the
// runner's handling of it is correct. Every entry names the upstream schema that recognises it, so
// the next reader can check the claim rather than trusting it.
//
// The set is deliberately the directives the 325 error fixtures actually use at the pinned sha,
// rather than all 55 upstream recognises. A runner that pre-blesses directives no fixture uses has
// pre-approved configurations nobody has looked at, which is the same silence with extra steps.
var knownPragmas = map[string]string{
	// PluginOptions keys (`Entrypoint/Options.ts`, `defaultOptions`).
	"compilationMode":        "PluginOptions: which functions the compiler treats as components",
	"outputMode":             "PluginOptions: what the pipeline emits; irrelevant to diagnostics but recognised upstream",
	"dynamicGating":          "PluginOptions: runtime feature gate",
	"eslintSuppressionRules": "PluginOptions: which eslint suppressions the compiler honours",
	"panicThreshold": "PluginOptions: whether a compilation error rethrows or is logged and the " +
		"function skipped (`Entrypoint/Program.ts:256`). All three fixtures carrying it pass " +
		"`\"none\"`, which is also upstream's default at `Entrypoint/Options.ts:314`, so it changes " +
		"nothing for them. Modelled rather than ignored because it is a real compiler option that " +
		"would change behaviour at another value, and an entry here says the runner knows that.",

	// EnvironmentConfigSchema keys (`HIR/Environment.ts`).
	"validatePreserveExistingMemoizationGuarantees": "Environment: preserve-manual-memoization validation",
	"validateRefAccessDuringRender":                 "Environment: refs validation",
	"validateExhaustiveMemoizationDependencies":     "Environment: memo dependency exhaustiveness",
	"validateExhaustiveEffectDependencies":          "Environment: effect dependency exhaustiveness",
	"validateNoSetStateInRender":                    "Environment: set-state-in-render validation",
	"validateNoFreezingKnownMutableFunctions":       "Environment: immutability validation",
	"validateNoCapitalizedCalls":                    "Environment: capitalized-call validation",
	"validateNoJSXInTryStatements":                  "Environment: error-boundaries validation",
	"validateNoImpureFunctionsInRender":             "Environment: purity validation",
	"validateNoDerivedComputationsInEffects":        "Environment: derived-computation-in-effect validation",
	"validateSourceLocations":                       "Environment: asserts diagnostics carry locations",
	"validateBlocklistedImports":                    "Environment: incompatible-library validation",
	"enablePreserveExistingMemoizationGuarantees":   "Environment: enables the memoization-preservation pass",
	"enableOptionalDependencies":                    "Environment: optional chaining in dependency paths",
	"enableTransitivelyFreezeFunctionExpressions":   "Environment: freezing propagation",
	"enableTreatSetIdentifiersAsStateSetters":       "Environment: setter inference by naming",
	"enableTreatRefLikeIdentifiersAsRefs":           "Environment: ref inference by naming",
	"enableUseKeyedState":                           "Environment: keyed state hook",
	"enableAssumeHooksFollowRulesOfReact":           "Environment: hook purity assumption",
	"enableCustomTypeDefinitionForReanimated":       "Environment: reanimated shape definitions",
	"throwUnknownException__testonly":               "Environment: test-only pipeline exception injection",
}

// ignoredPragmas are directives the runner recognises and deliberately does not act on, each with
// the evidence for why not acting on it is safe.
//
// This exists so that "the runner does nothing with this" is a written, reviewable claim instead of
// the absence of a branch. A directive silently absent from both maps is refused; a directive here
// has had someone argue the case in the line next to it.
var ignoredPragmas = map[string]string{
	"flow": "selects the parser, not the rules. `packages/snap/src/compiler.ts:36` is the whole of " +
		"upstream's handling: `source.indexOf('@flow') !== -1 ? 'flow' : 'typescript'`. It changes " +
		"which syntax parses, never which diagnostics a parsed program produces. A Go runner with no " +
		"Flow parser must decline these fixtures as unparseable rather than score them, which is what " +
		"RequiresFlow reports; that is a stated exclusion, not a skip.",

	"skip": "inert upstream. Grepped the whole of `packages/snap/src` at the pinned sha: no code " +
		"reads `@skip`. Eight of the 325 carry it and upstream runs all eight regardless. Honouring " +
		"it here would drop fixtures upstream scores, so the runner scores them too and records that " +
		"the directive is dead rather than silently obeying an instruction nothing else obeys.",

	"enableFlowSuppressions": "not in `EnvironmentConfigSchema.shape` at the pinned sha, so upstream's " +
		"own parser drops it at the `continue`. The recognised spelling is the PluginOptions key " +
		"`flowSuppressions`. Whatever the fixture author intended, upstream applies nothing, so the " +
		"expectation was recorded with the directive inactive and the runner matches that by also " +
		"applying nothing.",

	"enableNewMutationAliasingModel": "not in `EnvironmentConfigSchema.shape` at the pinned sha; " +
		"upstream drops it. Thirteen of the 325 carry it. Their expectations were therefore generated " +
		"with the flag inactive, so ignoring it reproduces the conditions the goldens were recorded " +
		"under. This is the most load-bearing entry in this map and the one most likely to go stale: " +
		"if upstream lands the key, these thirteen expectations change meaning and this note is the " +
		"only thing that will say so.",

	"enablePropagateDepsInHIR": "not in `EnvironmentConfigSchema.shape` at the pinned sha; upstream " +
		"drops it at the `continue`. Same reasoning as enableNewMutationAliasingModel.",

	"loggerTestOnly": "a snap output-formatting switch, not a compiler config. " +
		"`packages/snap/src/compiler.ts:76` reads it off the first line and `:349` uses it to decide " +
		"whether to serialise the compiler's log events into an extra section of the golden. It is " +
		"in neither `EnvironmentConfigSchema.shape` nor `Entrypoint/Options.ts`, so the compiler " +
		"never sees it and it cannot change which diagnostics a program produces -- it changes what " +
		"upstream prints about them. A runner comparing diagnostics rather than reproducing snap's " +
		"output format has nothing to do with it.",

	"expectNothingCompiled": "an assertion made by upstream's test harness about its own output, not " +
		"a compiler config. `packages/snap/src/compiler.ts:365` is the whole of its handling: " +
		"it filters the log for `CompileSuccess` and `CompileError` events and fails the fixture if " +
		"the presence of those events disagrees with the directive. It appears in neither " +
		"`EnvironmentConfigSchema.shape` nor `Entrypoint/Options.ts`, so nothing in the compiler " +
		"reads it and it cannot change which diagnostics a program produces. Two of the clean " +
		"fixtures carry it. What it asserts -- that the compiler bailed out entirely rather than " +
		"compiling anything -- is a claim about the pipeline's control flow that a rule-level runner " +
		"has no events to check, so declining to model it loses nothing a diagnostic comparison " +
		"would have caught.",

	"compilationMode(infer)": "a spelling upstream cannot parse. `splitPragma` splits a value on `:` " +
		"only, so `@compilationMode(infer)` yields the key `compilationMode(infer)`, which matches no " +
		"schema entry and is dropped. The fixture reads as configured and is not. Recorded rather than " +
		"normalised to `compilationMode`, because normalising it would score the fixture under a " +
		"configuration upstream never applied when it recorded the expectation.",
}

// ParsePragmas reads the directives from a fixture's first line.
//
// Scope is deliberate. Upstream passes only the first line to the pragma parser, so a directive on
// line two configures nothing, and a runner that scanned the whole file would apply configuration
// upstream ignored and disagree with the goldens for a reason that looks like a rule defect.
//
// The split reproduces upstream's `splitPragma` exactly, including its sharp edge: the value is
// everything after the FIRST `:`, and a key with no `:` is truncated at the first space. That is why
// `@compilationMode(infer)` yields the key `compilationMode(infer)` here rather than being helpfully
// normalised. Reproducing the flaw is the point; a parser that is kinder than upstream's scores
// fixtures under a configuration upstream did not use.
func ParsePragmas(firstLine string) []Pragma {
	if !strings.Contains(firstLine, "@") {
		return nil
	}

	var pragmas []Pragma
	// Entry 0 is whatever preceded the first `@` (the `//` and any prose), never a directive.
	for _, entry := range strings.Split(firstLine, "@")[1:] {
		entry = text.TrimWhitespace(entry)
		if entry == "" {
			continue
		}
		if index := strings.Index(entry, ":"); index >= 0 {
			pragmas = append(pragmas, Pragma{
				Key:      entry[:index],
				Value:    entry[index+1:],
				HasValue: true,
			})
			continue
		}
		key, _, _ := strings.Cut(entry, " ")
		if key != "" {
			pragmas = append(pragmas, Pragma{Key: key})
		}
	}
	return pragmas
}

// CheckPragmas returns an error for the first directive the runner does not model.
//
// One error rather than a list, because the intended response is to go read the directive and add
// it to one of the two maps, and a wall of them invites triaging the list instead of reading any
// entry in it.
func CheckPragmas(fixturePath string, pragmas []Pragma) error {
	for _, pragma := range pragmas {
		if _, ok := knownPragmas[pragma.Key]; ok {
			continue
		}
		if _, ok := ignoredPragmas[pragma.Key]; ok {
			continue
		}
		return &UnknownPragmaError{FixturePath: fixturePath, Pragma: pragma}
	}
	return nil
}

// KnownPragmaKeys lists the modelled directives, sorted. For tests and for reporting.
func KnownPragmaKeys() []string {
	keys := make([]string, 0, len(knownPragmas))
	for key := range knownPragmas {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// IgnoredPragmaKeys lists the deliberately-inert directives, sorted.
func IgnoredPragmaKeys() []string {
	keys := make([]string, 0, len(ignoredPragmas))
	for key := range ignoredPragmas {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
