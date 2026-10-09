package react_conformance_score

// This file is the rule engine the conformance harness was missing.
//
// `internal/react_conformance` could decide which fixtures a shipped rule is RESPONSIBLE for, and
// then reported every one of them as `not scored: no rule engine run`. That category was honest and
// it was the whole gap: a taxonomy that had never once asked a rule a question. This file supplies
// the missing input — fixture in, real rule over a real type graph, `{heading, message, line}` out
// — so that `Classify` has something to classify.
//
// # The join, and why it is by message id rather than by text
//
// cohere's diagnostic text was written for cohere's users and does not match React's prose, nor
// should it. So a finding is joined to the golden through the rule's own message id, which is the
// same join upstream's `RustBackendComparison-test.ts` makes between two backends. `Rules` below
// is that table, and every entry was checked against a message some golden in this corpus actually
// contains — `TestUpstreamMessageJoinIsNotByResemblance` fails an entry that was not.
//
// An id absent from the table reports NOTHING rather than being mapped by resemblance. That is
// deliberate and it costs recall: a rule can find the right thing under an id nobody joined and
// score as silent. The alternative is worse, because a wrong join produces a specific, confident,
// wrong failure, and this harness exists to stop exactly that.
//
// # What this file deliberately does not do: gate on the fixture's pragma
//
// The obvious design is to run a rule only when the fixture carries the pragma that turns its
// validator on, and it is wrong here. Measured two ways.
//
// From React's own executable (`eslint-plugin-react-hooks/cjs/...development.js`, the unminified
// bundle, at the schema around line 31614) the validators split by default:
//
//	validateNoSetStateInRender      z.boolean().default(true)    pragma is REDUNDANT
//	validateRefAccessDuringRender   z.boolean().default(true)    pragma is REDUNDANT
//	validateNoImpureFunctionsInRender      default(false)        pragma ENABLES
//	validateNoJSXInTryStatements           default(false)        pragma ENABLES
//	validateNoFreezingKnownMutableFunctions default(false)       pragma ENABLES
//
// And from the corpus, cross-tabulating the pragma against the attributed rule: 2 of the 11
// `set-state-in-render` fixtures carry no `@validateNoSetStateInRender`, and 20 of the 38 `refs`
// fixtures carry no `@validateRefAccessDuringRender`. Gating on the pragma would have silenced
// those 22 and scored the silence as correct, because their validator is on by default and the
// pragma is a restatement.
//
// The sharpest case is `incompatible-library`: all three of its fixtures LACK
// `@validateBlocklistedImports`, and the single fixture that carries it attributes to a different
// rule entirely. A pragma gate there scores zero fixtures and reads as a clean run.
//
// So: no gating, for the six rules wired here, none of which is an opt-in validator. If an opt-in
// validator is ever wired (`purity` via `@validateNoImpureFunctionsInRender`, `error-boundaries`
// via `@validateNoJSXInTryStatements`, `immutability` via
// `@validateNoFreezingKnownMutableFunctions`) the gate becomes load-bearing, because upstream
// recorded those goldens with the validator OFF unless the pragma turned it on. Those three are
// the reason this paragraph is here rather than a bare "pragmas do not matter".

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/react"
	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
	"github.com/system-inc/cohere/internal/types/program"
)

// Rules are the cohere rules wired to upstream rule names, with the message join for each.
//
// Keyed by UPSTREAM's rule name, which is what `Fixture.Rules()` reports and what `ShippedRules` is
// keyed by, so the three tables join without a translation step.
//
// Seven rules rather than the fifteen cohere ships. `globals` is here rather than in `score_test.go`
// so there is ONE mechanism rather than two that can drift; its own test still pins its numbers. The
// other eight are absent for stated reasons rather than by omission:
//
//	set-state-in-effect   zero error-named fixtures in this corpus, so nothing to score
//	static-components     zero error-named fixtures in this corpus
//	void-use-memo         zero error-named fixtures in this corpus
//	error-boundaries      its three fixtures all expect a `Todo` from BuildHIR, so the validator is
//	                      never reached upstream either; attribution keeps them out
//	purity                opt-in validator, needs the pragma gate this file does not implement
//	immutability          being ported by another agent as this was written; also opt-in
//	incompatible-library  no message id join checked yet
//	refs                  no message id join checked yet
var Rules = map[string]RuleUnderTest{
	"globals": {
		Upstream: "globals",
		Rule:     react.Globals,
		Messages: map[string]UpstreamMessage{
			"globalReassignment": {Heading: "Error", Text: "Cannot reassign variables declared outside of the component/hook"},
		},
	},
	"hooks": {
		Upstream: "hooks",
		Rule:     react.RulesOfHooks,
		Messages: map[string]UpstreamMessage{
			"rulesOfHooksConditional": {Heading: "Error", Text: hooksConditional},
			"rulesOfHooksLoop":        {Heading: "Error", Text: hooksConditional},
			"rulesOfHooksCallback":    {Heading: "Error", Text: hooksNestedFunction},
			"rulesOfHooksNotComponent": {
				Heading: "Error", Text: hooksNestedFunction,
			},
			"rulesOfHooksTopLevel":       {Heading: "Error", Text: hooksNestedFunction},
			"rulesOfHooksClassComponent": {Heading: "Error", Text: hooksNestedFunction},
			"rulesOfHooksAsync":          {Heading: "Error", Text: hooksNestedFunction},
		},
	},
	"set-state-in-render": {
		Upstream: "set-state-in-render",
		Rule:     react.SetStateInRender,
		Messages: map[string]UpstreamMessage{
			"setStateInRender":  {Heading: "Error", Text: "Cannot call setState during render"},
			"setStateInUseMemo": {Heading: "Error", Text: "Calling setState from useMemo may trigger an infinite loop"},
		},
	},
	"use-memo": {
		Upstream: "use-memo",
		Rule:     react.UseMemo,
		Messages: map[string]UpstreamMessage{
			"useMemoCallbackNotInline":              {Heading: "Error", Text: "Expected the first argument to be an inline function expression"},
			"useMemoDependencyListNotArrayLiteral":  {Heading: "Error", Text: "Expected the dependency list for useMemo to be an array literal"},
			"useMemoCallbackHasParameters":          {Heading: "Error", Text: "useMemo() callbacks may not accept parameters"},
			"useMemoCallbackAsyncOrGenerator":       {Heading: "Error", Text: "useMemo() callbacks may not be async or generator functions"},
			"useMemoCallbackReassignsOuterVariable": {Heading: "Error", Text: "useMemo() callbacks may not reassign variables declared outside of the callback"},
		},
	},
	"config": {
		Upstream: "config",
		Rule:     react.Config,
		Messages: map[string]UpstreamMessage{
			"invalidTypeConfiguration": {Heading: "Error", Text: "Invalid type configuration for module"},
		},
	},
	"gating": {
		Upstream: "gating",
		Rule:     react.Gating,
		Messages: map[string]UpstreamMessage{
			"invalidGatingDirective": {Heading: "Error", Text: "Dynamic gating directive is not a valid JavaScript identifier"},
		},
	},
	"unsupported-syntax": {
		Upstream: "unsupported-syntax",
		Rule:     react.UnsupportedSyntax,
		Messages: map[string]UpstreamMessage{
			"unsupportedEval": {Heading: "Compilation Skipped", Text: "The 'eval' function is not supported"},
		},
	},
}

// # The `hooks` result, and why 25 of 57 fail
//
// Measured 2026-08-24: `rules-of-hooks` scores 32 passed and 25 failed against React's own Hooks
// goldens. Those 25 are the deliverable of this wiring rather than a problem with it, and they
// partition cleanly into four causes, every one hand-checked against the golden.
//
// The root cause of the first three is one fact stated in the rule's own doc comment: cohere's
// `rules-of-hooks` is a port of **ESLint's** `rules-of-hooks`, whose subject is where a hook CALL
// sits in the control-flow graph. React's `Hooks` ERROR CATEGORY is emitted by the COMPILER, which
// runs a wider analysis over its own intermediate representation. The two share a name and a
// message vocabulary and are not the same validator, so a fixture can be a true Hooks violation
// upstream and outside the ported rule's subject entirely.
//
//	9  hook referenced as a VALUE rather than called
//	   `const x = useState;`, `<Child foo={useFoo} />`, a hook in a ternary or a conditional test.
//	   ESLint's rule never looks at a bare reference; the compiler reports it because lowering has
//	   to place the value somewhere. Wholly outside the ported rule's subject.
//	   error.invalid-assign-hook-to-local.js, error.invalid-pass-hook-as-prop.js,
//	   error.invalid-pass-hook-as-call-arg.js, error.propertyload-hook.js,
//	   error.hook-property-load-local-hook.js, error.invalid-ternary-with-hook-values.js,
//	   rules-of-hooks/error.invalid-call-phi-possibly-hook.js,
//	   rules-of-hooks/error.invalid-hook-as-conditional-test.js,
//	   rules-of-hooks/error.invalid-hook-reassigned-in-conditional.js
//
//	8  a conditional call the ported rule cannot SEE as a hook call
//	   Two shapes. An optional call — `useConditionalHook?.()` — and a hook reached through an
//	   ALIAS, either an aliased import (`import {useState as state}`) or a property of a local. The
//	   ported rule decides hook-ness from the callee's spelling, deliberately and for fidelity
//	   (`rules_of_hooks.go` records that it answers nothing by resolution because oxc reads the
//	   syntactic parent and never resolves). `state()` is not spelled like a hook, so nothing fires.
//	   The compiler resolves through the module graph and knows it is `useState`.
//
//	4  hook identity not stable across renders
//	   `const useMedia = useVideoPlayer(); useMedia();` — a hook obtained from another hook's return
//	   value or from a prop. Upstream's message is "must be the same function on every render", an
//	   inter-procedural value-tracking judgment. ESLint's rule has no counterpart.
//
//	3  the violation IS found, under a different one of React's four Hooks messages
//	   The only category here that is about this join rather than about scope. React chooses between
//	   "called conditionally" and "called within function expressions" by the nature of the
//	   ENCLOSING function; cohere reports the innermost violation it sees. Measured, upstream is not
//	   self-consistent by shape either: `normalFunctionWithConditionalHook` (a conditional hook in a
//	   plain function) gets the CONDITIONAL message, while the same conditional hook inside a
//	   returned function expression gets the NESTED-FUNCTION one, twice. Both implementations report
//	   at the same line and both are right that the code is wrong; they disagree on which sentence
//	   describes it.
//	   rules-of-hooks/error.invalid-rules-of-hooks-d740d54e9c21.js,
//	   rules-of-hooks/error.invalid.invalid-rules-of-hooks-0a1dbff27ba0.js,
//	   rules-of-hooks/error.invalid.invalid-rules-of-hooks-d842d36db450.js
//
//	1  a COUNT difference, same family as the three above
//	   rules-of-hooks/error.invalid.invalid-rules-of-hooks-0de1224ce64b.js: a hook inside a
//	   `useEffect` callback inside a returned function expression. The golden expects the
//	   nested-function message TWICE and cohere reports it once, because React reports the inner
//	   hook AND the enclosing function expression while cohere reports the innermost violation only.
//	   Read as a possible defect first and reclassified after reading the golden: the disagreement
//	   is how many findings one nesting produces, not whether the code is wrong.
//
// These are recorded here, in prose, rather than moved into `statedDivergences`. The distinction is
// deliberate and it is the same one `verdict.go` draws: a stated divergence is a boundary somebody
// DECIDED and a test holds in place, whereas these 25 are an unclosed gap between two validators
// that share a name. Parking them in the excluded column would turn a real 32-of-57 into a
// decorative 32-of-32, which is exactly the pressure `verdict.go` warns that a scoring harness
// creates. They should stay failures until somebody decides to widen the rule or to state the
// boundary formally.

// The three Hooks messages the corpus uses, spelled once because they are long and because a typo
// in one copy of a 200-character literal is invisible in review.
//
// Upstream emits four distinct Hooks messages and cohere's rule carries nine message ids, so the
// join is many-to-one in both directions. That asymmetry is real rather than a modelling shortcut:
// React's `Hooks` category makes one distinction (called conditionally / referenced as a value /
// called somewhere that is not a component body / not stable across renders) where cohere's rule
// makes nine, because cohere's messages were written to tell a developer which shape they wrote.
const (
	hooksConditional    = "Hooks must always be called in a consistent order, and may not be called conditionally. See the Rules of Hooks (https://react.dev/warnings/invalid-hook-call-warning)"
	hooksNestedFunction = "Hooks must be called at the top level in the body of a function component or custom hook, and may not be called within function expressions. See the Rules of Hooks (https://react.dev/warnings/invalid-hook-call-warning)"
)

// UpstreamMessage is the golden-side text a cohere message id corresponds to.
type UpstreamMessage struct {
	Heading string
	Text    string
}

// RuleUnderTest pairs a cohere rule with the upstream rule name it implements and the message join.
type RuleUnderTest struct {
	Upstream string
	Rule     rule.Rule
	Messages map[string]UpstreamMessage
}

// tsConfigText is the compiler configuration each fixture is built under.
//
// `allowJs` because 300 of the 325 inputs are `.js`, and `strict: false` because the corpus is
// minimal repros that would not survive strict mode and whose expectations were never recorded
// under it.
const tsConfigText = `{
  "compilerOptions": {
    "target": "esnext",
    "module": "esnext",
    "moduleResolution": "bundler",
    "jsx": "preserve",
    "strict": false,
    "noEmit": true,
    "skipLibCheck": true,
    "allowJs": true
  },
  "include": ["**/*.ts", "**/*.tsx", "**/*.js", "**/*.jsx"]
}`

// Analyze runs one cohere rule over one fixture and returns what it found, in the shape the
// conformance comparison expects.
//
// It returns `ErrUnsupported` rather than an empty Result for anything that stopped the rule from
// being asked at all — a parse failure, a missing checker. An empty Result is the claim "this rule
// found nothing here", which on this corpus is a wrong answer rather than an absent one, and the
// two have to stay in different columns.
func Analyze(subject RuleUnderTest, fixture react_conformance.Fixture, directory string) (react_conformance.Result, error) {
	// The fixture is written under ITS OWN filename, never under the one the golden's location line
	// names, and this is the opposite of what this function did until 2026-08-24.
	//
	// The inherited comment read: "The golden's location line names a `.ts` or `.tsx` file even when
	// the input is `.js`, and the extension decides how the parser reads JSX. Following the golden
	// keeps the parse the same one upstream made." The first clause is true and the conclusion is
	// backwards. Upstream parsed the `.js` INPUT, with JSX enabled; the `.ts` in the golden is what
	// upstream's error printer rendered, not what it read. Following it hands a JSX file to a parser
	// that reads `<div />` as a type assertion, and the resulting parse damage is silent.
	//
	// Measured on `error.invalid-eval-unsupported.js`, whose source is four lines with a `return
	// <div />`: written as the golden's `.ts` the rule reports 0 findings, written as its own `.js`
	// it reports 1, which is the golden's answer. The rule already ships that exact source as a
	// passing fixture of its own (`unsupported_syntax_test.go:18`), so the harness was manufacturing
	// a failure for a rule that was right.
	//
	// This is the failure this whole package is built against, arriving inside the instrument: a
	// silent parse difference reads as a rule defect, and it reads as one CONFIDENTLY, with a
	// specific missing diagnostic named.
	name := filepath.Base(fixture.Name)

	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(fixture.Source), 0o644); err != nil {
		return react_conformance.Result{}, err
	}
	configPath := filepath.Join(directory, "tsconfig.json")
	if err := os.WriteFile(configPath, []byte(tsConfigText), 0o644); err != nil {
		return react_conformance.Result{}, err
	}

	graph, err := program.Build(program.Options{
		ConfigFileName:   configPath,
		CurrentDirectory: directory,
		SingleThreaded:   true,
	})
	if err != nil {
		return react_conformance.Result{}, &react_conformance.ErrUnsupported{Reason: "the fixture does not parse: " + err.Error()}
	}

	projectFiles := graph.ProjectFiles()
	if len(projectFiles) == 0 {
		return react_conformance.Result{}, &react_conformance.ErrUnsupported{Reason: "the fixture produced no project files"}
	}

	var sourceFile *ast.SourceFile
	wanted := filepath.ToSlash(path)
	for _, candidate := range projectFiles {
		if filepath.ToSlash(candidate.FileName().AsString()) == wanted {
			sourceFile = candidate
		}
	}
	if sourceFile == nil {
		return react_conformance.Result{}, &react_conformance.ErrUnsupported{Reason: "the fixture is not in the built program"}
	}

	fileChecker, release := graph.CheckerForFile(context.Background(), sourceFile)
	defer release()
	if fileChecker == nil {
		// A nil checker would make a type-aware rule take its decline path and report nothing,
		// which would score as a failure it did not earn.
		//
		// Crash protection rather than a behavioural filter, and unreachable through this harness:
		// probed `CheckerForFile` over a component, an empty file, `const = = =;` and an unclosed
		// function body, and it returned non-nil for all four. A mutation replacing this decline
		// with an empty Result therefore survives the suite, because nothing can reach it. Kept
		// rather than deleted: the cost is one comparison and the branch is what stands between a
		// future construction that CAN produce a nil checker and a silent wrong answer.
		return react_conformance.Result{}, &react_conformance.ErrUnsupported{Reason: "no type checker for the fixture"}
	}

	var diagnostics []rule.Diagnostic
	ruleContext := rule.Context{
		SourceFile:  sourceFile,
		Program:     rule.ViewProgram(graph.Program, sourceFile, subject.Rule),
		TypeChecker: fileChecker,
		FileCache:   rule.NewFileCache(),
		Report: func(diagnostic rule.Diagnostic) {
			diagnostic.RuleName = subject.Rule.Name
			if diagnostic.SourceFile == nil {
				diagnostic.SourceFile = sourceFile
			}
			diagnostics = append(diagnostics, diagnostic)
		},
	}

	listeners := subject.Rule.Run(ruleContext, nil)
	if listeners != nil {
		walk(sourceFile.AsNode(), listeners)
	}

	result := react_conformance.Result{}
	for _, diagnostic := range diagnostics {
		message, found := subject.Messages[diagnostic.Message.Id]
		if !found {
			// Dropped rather than mapped by resemblance. `TestUnjoinedMessageIdsAreEnumerated` is
			// what stops this being silent: it runs the same rules with the join emptied and pins
			// the set of ids that land here at zero, with a control asserting it observed findings
			// at all so the zero cannot come from a probe that never ran.
			//
			// Measured: 48 findings across the wired rules on this corpus, drawn from 9 distinct
			// ids, and none of them lands here. A mutation making this branch assign a default
			// message instead of skipping therefore SURVIVES the suite, and that is an equivalent
			// mutant rather than a fixture gap — no input in this corpus can reach the branch. The
			// nine exercised ids are globalReassignment, rulesOfHooksConditional (27),
			// rulesOfHooksLoop (7), rulesOfHooksNotComponent (5), rulesOfHooksCallback (4),
			// invalidGatingDirective, unsupportedEval, useMemoCallbackNotInline,
			// useMemoDependencyListNotArrayLiteral and useMemoCallbackReassignsOuterVariable.
			//
			// That verdict expires if a rule grows a message or a fixture is added, which is what
			// the enumeration test is for.
			continue
		}
		line, _ := lineAndCharacterOfPosition(sourceFile, diagnostic.Range.Pos())
		result.Errors = append(result.Errors, react_conformance.ReportedError{
			Heading: message.Heading,
			Message: message.Text,
			Line:    line + 1,
		})
	}
	return result, nil
}

// collectMessageIds runs a rule over a fixture and returns every message id it reported.
//
// Deliberately a second pass over the same fixture rather than a refactor threading a callback
// through `Analyze`. `Analyze` is the function the whole score depends on, and it is worth more to
// keep it a straight read than to save one program build in a test that runs 78 fixtures.
func collectMessageIds(subject RuleUnderTest, fixture react_conformance.Fixture, directory string) ([]string, error) {
	name := filepath.Base(fixture.Name)
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(fixture.Source), 0o644); err != nil {
		return nil, err
	}
	configPath := filepath.Join(directory, "tsconfig.json")
	if err := os.WriteFile(configPath, []byte(tsConfigText), 0o644); err != nil {
		return nil, err
	}

	graph, err := program.Build(program.Options{
		ConfigFileName:   configPath,
		CurrentDirectory: directory,
		SingleThreaded:   true,
	})
	if err != nil {
		return nil, &react_conformance.ErrUnsupported{Reason: "the fixture does not parse: " + err.Error()}
	}

	var sourceFile *ast.SourceFile
	wanted := filepath.ToSlash(path)
	for _, candidate := range graph.ProjectFiles() {
		if filepath.ToSlash(candidate.FileName().AsString()) == wanted {
			sourceFile = candidate
		}
	}
	if sourceFile == nil {
		return nil, &react_conformance.ErrUnsupported{Reason: "the fixture is not in the built program"}
	}

	fileChecker, release := graph.CheckerForFile(context.Background(), sourceFile)
	defer release()

	var ids []string
	ruleContext := rule.Context{
		SourceFile:  sourceFile,
		Program:     rule.ViewProgram(graph.Program, sourceFile, subject.Rule),
		TypeChecker: fileChecker,
		FileCache:   rule.NewFileCache(),
		Report: func(diagnostic rule.Diagnostic) {
			ids = append(ids, diagnostic.Message.Id)
		},
	}
	listeners := subject.Rule.Run(ruleContext, nil)
	if listeners != nil {
		walk(sourceFile.AsNode(), listeners)
	}
	return ids, nil
}

// analyzeRaw runs a rule and returns the raw message ids it reported, before the message join.
//
// It exists so the join's DROPS can be enumerated rather than merely trusted. `Analyze` silently
// skips a finding whose id is not in the rule's table, which is the correct default and is also
// invisible; this is the instrument that makes it visible. Used by
// `TestUnjoinedMessageIdsAreEnumerated`.
func analyzeRaw(subject RuleUnderTest, fixture react_conformance.Fixture, directory string) ([]string, error) {
	probe := subject
	probe.Messages = nil
	ids, err := collectMessageIds(probe, fixture, directory)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// SelectFixtures returns the flow-free fixtures whose every diagnostic belongs to one upstream rule.
//
// Selecting on the ATTRIBUTED rule rather than on the filename is the whole reason attribution
// exists: the `error.` prefix names the fixture author's intent, not the validator that ran. For
// `globals` the two differ by 11 against 23.
func SelectFixtures(fixtures []react_conformance.Fixture, upstream string) []react_conformance.Fixture {
	var selected []react_conformance.Fixture
	for _, fixture := range fixtures {
		if fixture.RequiresFlow() {
			continue
		}
		rules, complete := fixture.Rules()
		if !complete || len(rules) != 1 || rules[0] != upstream {
			continue
		}
		selected = append(selected, fixture)
	}
	return selected
}

// UpstreamNames lists the wired rules, sorted, so a report's rows do not move between runs.
func UpstreamNames() []string {
	names := make([]string, 0, len(Rules))
	for name := range Rules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// walk visits every node, dispatching to the listeners registered for its kind.
func walk(node *ast.Node, listeners rule.Listeners) {
	if node == nil {
		return
	}
	if listener, found := listeners[node.Kind]; found {
		listener(node)
	}
	node.ForEachChild(func(child *ast.Node) bool {
		walk(child, listeners)
		return false
	})
}

// lineAndCharacterOfPosition converts an offset into a zero-based line and column.
//
// Written here rather than taken from the shim because the shim's spelling has moved before, and a
// scoring harness that stops compiling for a rename is worse than twelve lines of arithmetic.
func lineAndCharacterOfPosition(sourceFile *ast.SourceFile, position int) (line int, character int) {
	text := sourceFile.Text()
	if position > len(text) {
		position = len(text)
	}
	line = 0
	lastNewline := -1
	for index := 0; index < position; index++ {
		if text[index] == '\n' {
			line++
			lastNewline = index
		}
	}
	return line, position - lastNewline - 1
}
