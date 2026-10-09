package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	hir "github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreserveManualMemoizationValueUnmemoized = rule.Message{
	Id: "preserveManualMemoizationValueUnmemoized",
	Description: "This value was memoized in source but not in the compiled output, so the " +
		"stability the `useMemo` or `useCallback` promised is gone. Code downstream that " +
		"compares this value by identity, a dependency array, a `memo` boundary, or an effect " +
		"that should not re-run, will now see a fresh value on renders where it used to see the " +
		"same one. React Compiler reports this rather than compiling the component, because " +
		"quietly weakening a guarantee the code already relies on is worse than not optimizing " +
		"it. Simplify what the callback captures so the memoization can be preserved.",
}

var messagePreserveManualMemoizationDependencyMutable = rule.Message{
	Id: "preserveManualMemoizationDependencyMutable",
	Description: "This dependency may be modified after the memo block that reads it, so the " +
		"memoized value cannot be trusted. A dependency array is a claim that the cached value " +
		"stays correct until one of these entries changes; a later mutation of the entry itself " +
		"changes what the value should be without changing what the array compares, so the " +
		"stale value survives. Finish mutating the value before the memo block reads it, or " +
		"depend on something that is not mutated afterwards.",
}

// PreserveManualMemoization flags a `useMemo` or `useCallback` whose memoization the compiler could
// not preserve.
//
//	valid:   function Component(props) { const x = useMemo(() => [props.a], [props.a]); return <div>{x}</div>; }
//	valid:   function Component(props) { const cb = useCallback(() => props.a, [props.a]); return <div onClick={cb} />; }
//	invalid: function Foo(props) { let x = []; x.push(props); const cb = useCallback(() => [x], [x]); x = makeArray(); return cb; }
//	invalid: function Component(props) { const data = useMemo(() => { const x = []; x.push(props?.items); return x; }, [props?.items, props.cond]); return <Validate output={data} />; }
//
// # This is a compiler validation pass, not an abstract syntax tree walk
//
// Upstream is `ValidatePreservedManualMemoization.ts`, 618 lines running over React Compiler's own
// high level intermediate representation after the whole memoization pipeline has run. There is no
// visitor to port and no `create()` to mirror: the judgment is "compare what the developer wrote
// against what survived compilation", and neither side of that comparison exists until the function
// has been lowered, had its mutable ranges inferred, had reactive scopes assigned, merged, pruned,
// and rebuilt into a reactive tree.
//
// That pipeline is already built and measured in
// `internal/lint/ecmascript/high_level_intermediate_representation`, where the pass itself lives as
// `ValidatePreservedManualMemoization` behind the `AnalyzePreservedManualMemoization` entry point.
// This file is the rule surface over it: walk the outermost function-like nodes, lower each one,
// run the analysis, and turn each finding into a diagnostic anchored where upstream anchors it.
// `ForEachFunctionLike` is exported from that package for exactly this purpose.
//
// # Where the findings point, measured rather than read
//
// Every span below was established by driving React's own rule through the ESLint Linter interface
// over upstream's vendored fixture corpus, then reading the reported range out of the source. The
// two conditions anchor differently and a single fixture proves it, because it reports both at once:
//
//	const cb = useCallback(() => [x], [x]);
//	                       ^^^^^^^^^  value memoized in source but not in output
//	                                   ^  this dependency may be modified later
//
// That is `new-mutability/error.invalid-useCallback-captures-reassigned-context.js`, reported by
// upstream at 11:26 and 11:38 with the two distinct messages. The intermediate representation
// already carries both anchors on `Identifier.Node`: the value condition names the callback, whose
// node is the arrow function, and the dependency condition names the written dependency, whose node
// is the identifier inside the array. So the rule reports on the finding's own identifier node and
// needs no anchor logic of its own, which was measured rather than assumed.
//
// Upstream's third firing condition, the inferred-against-written dependency comparison, reports at
// the memo callback instead (`validateInferredDep` takes `memoLocation`, which
// `makeManualMemoizationMarkers` sets to `fnExpr.loc`). The pass here folds that condition into the
// same `ValueUnmemoized` kind and reports it at the dependency it could not match, which is a
// recorded divergence in span rather than in verdict: the finding appears on the same line, inside
// the same call, naming the dependency that caused it. Measured on
// `error.hoist-optional-member-expression-with-conditional-optional.js`, where upstream points at
// `() => {` on line 4 and this rule points at `props` on line 5.
//
// # One anchor per declaration rather than per use, which is a bounded substrate gap
//
// When several memo blocks in one function name the same value, upstream reports one finding per
// block at the block's own column, and this rule reports the same NUMBER of findings collapsed onto
// FEWER distinct columns. Measured against the installed rule on two `useCallback` blocks sharing
// one mutated dependency: upstream gives four findings at 6:25, 6:37, 7:25 and 7:37, and this rule
// gives the same four at THREE anchors. The two callbacks keep their own anchors because they are
// distinct syntax; only the shared dependency collapses, because both blocks name one
// declaration.
//
// The count is faithful and the anchor is not, and the cause is one field. Upstream's `loc` comes
// from the operand's own `Place`, which is a per-USE record; a `PreserveManualMemoizationFinding`
// carries an `IdentifierId`, which names the DECLARATION, so the per-use span is discarded before
// the finding is built even though `Place.Range` carries it. Closing it means adding a range to the
// finding at both report sites in
// `internal/lint/ecmascript/high_level_intermediate_representation/preserve_manual_memoization.go`, which
// is a change to a shared pass rather than to this rule, and it belongs in its own commit.
//
// Its cost on the real tree, measured: 287 findings at 221 distinct locations across 3,516 files, so
// 66 findings land on a location that already carries one. A reader sees a repeated line rather than
// a wrong one, and every repeat names a real memo block that really did lose its memoization.
//
// # What upstream reports that this cannot
//
// Upstream renders one message per finding carrying the inferred dependency and the source
// dependency list, computed from the compiler's own pretty printer. The pass here reports the
// judgment and the condition without reconstructing that text, so the two messages below are fixed
// rather than interpolated. That is a deliberate subset: the finding, its condition and its span are
// what a reader acts on, and a partially reconstructed dependency rendering would be a second place
// for the two implementations to disagree.
//
// # What it costs, measured, because it is the most expensive rule in the tree
//
// The first dry run over ahra: 287 findings across 3,516 files in 7,603ms, 43.4% of all rule time,
// when the rule lowered every outermost function in every file. Since the spelling gate
// (`hir.MayNameManualMemoization`) it analyzes only a function that can name a memo hook, and a
// profile on 2026-10-03 put it near 620ms. Since #rwsffzm it starts from a copy of the shared
// lowering rather than lowering the function again, which on a cold ahra run took 665K objects and
// 0.06 GB off. What it pays per function it does analyze is inherent: the
// compiler's intermediate representation and eighteen pipeline passes over it -- mutable range
// inference, scope assignment, alignment, merging, four prunes and a reactive rebuild -- because the
// judgment being ported is "what survived compilation", and nothing cheaper can answer it.
//
// Recorded here rather than left to be found because a share that large is a config decision
// somebody may want to make, and it should be made knowing the cost is inherent to the question
// rather than to this implementation.
//
// # Why this needs the type checker
//
// Lowering is keyed on `*ast.Symbol` identity throughout, and `Lower` takes the checker directly.
// Beyond that, the pipeline consults it for reactivity inference and for scope pruning. The
// declaration is for binding and resolution rather than for any single type question.
var PreserveManualMemoization = rule.Rule{
	// The upstream spelling. This rule ships in `eslint-plugin-react-hooks`, which was measured
	// rather than inferred: the plugin's rule table contains it and `eslint-plugin-react`'s does
	// not, and configuring it as `react/preserve-manual-memoization` makes ESLint refuse to lint
	// anything at all with `Could not find "preserve-manual-memoization" in plugin "react"`.
	Name: "react-hooks/preserve-manual-memoization",

	// Lowering resolves every binding through the checker, so this is the resolution half of the
	// port brief's scope table rather than the flag half. It is also the binder declaration the
	// brief describes: the locals tables the pipeline reads are only populated under a program.
	NeedsTypeChecker: true,
	// Reads other files only through shape readers (rule.ExportNameIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Declined once per file rather than once per node. `NeedsTypeChecker` governs the
		// REGISTRATION path and says nothing about the harness path, where a context is built by
		// hand with no checker at all, so the declaration is not the guard.
		//
		// What this guard is worth was measured rather than assumed, because a mutation
		// neutralizing it SURVIVES the fixture set and that survival is honest. Handed a nil
		// checker the pipeline does not panic: `Lower` returns a function, `Construct` runs, and
		// the analysis completes reporting zero. So the verdict is silence either way, and no
		// fixture can separate the two. The guard is a cost decline, not a correctness one -- it
		// refuses the file once instead of lowering and running the whole memoization pipeline over
		// every function in it to reach a foregone zero.
		//
		// It is kept for that reason and because the equivalence is a property of today's pipeline
		// rather than a contract: every pass below consults the checker, and one of them growing a
		// dereference would turn this from wasted work into the panic that costs every rule on the
		// file its verdict.
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				text := ctx.SourceFile.Text()
				if !hir.SpellsManualMemoization(text) && strings.IndexByte(text, '\\') < 0 {
					return
				}
				// Outermost function-like nodes only. A nested function is reached through
				// `Function.Functions` during lowering rather than by walking into it, because an
				// identifier id names one value in this function and a different value in a nested
				// one, so visiting both would compare across two numbering spaces.
				hir.ForEachFunctionLike(node, func(functionNode *ast.Node) {
					// Every finding is raised inside a memo block: the value condition at
					// `FinishMemoize`, the mutable dependency condition at `StartMemoize`, and the
					// inferred-dependency comparison only while a block opened by `StartMemoize` is
					// open (`compareInferredDependencies` returns on an empty `sourceDeps`). Those
					// markers are built only by `DropManualMemoization`, for a call it recognises, so
					// a function the spelling check clears holds no finding and is not lowered. The
					// check runs once on the file text above and again on each outermost function's
					// own range here, because lowering reads nothing outside that range: a file that
					// memoizes in one component still skips every other function in it.
					if !hir.MayNameManualMemoization(functionNode, text[functionNode.Pos():functionNode.End()]) {
						return
					}
					// A copy of the shared lowering rather than a lowering of its own. The pipeline
					// rewrites the graph it is handed, so it cannot read the shared one, and a copy
					// skips the control-flow build and single-assignment construction the other react
					// rules have already paid for on this function (#rwsffzm).
					function := hir.CloneFunction(hir.ForFunction(ctx, functionNode))
					if function == nil {
						return
					}
					for _, finding := range hir.AnalyzePreservedManualMemoization(function, ctx.TypeChecker) {
						identifier := function.Identifier(finding.Identifier)
						if identifier == nil || identifier.Node == nil {
							// Crash protection rather than a filter, which is worth stating because
							// a mutation removing this guard SURVIVES the whole fixture set and
							// reads as a blind spot.
							//
							// Measured over all 395 vendored corpus fixtures: 48 findings, zero
							// with a nil identifier, zero with a nil node. So no fixture can reach
							// this branch and one written for it would assert nothing. What it
							// prevents is a panic, which no `ExpectFindings` fixture can see:
							// `rule.TokenRange` reads `node.Loc` inside its own nil branch
							// (`internal/rule/rule.go:452`), so a nil node handed to `ReportNode`
							// takes the walk down. The walk recovers per FILE rather than per rule,
							// so that costs every rule its verdict on the file, not just this one.
							//
							// It is reachable in principle rather than merely defensive: the memo
							// markers are built by `newMarkerInstruction`, which creates its
							// temporary with `function.NewIdentifier("", nil, 0)` -- a nil node by
							// construction. The corpus never routes one into a finding today.
							continue
						}
						ctx.ReportNode(identifier.Node, messageFor(finding.Kind))
					}
				})
			},
		}
	},
}

// messageFor picks the message for a finding's condition.
//
// The two conditions are separate judgments with separate remedies, and upstream gives them separate
// detail messages, so they are separate ids here rather than one message with interpolated text.
func messageFor(kind hir.PreserveManualMemoizationKind) rule.Message {
	if kind == hir.PreserveManualMemoizationDependencyMutable {
		return messagePreserveManualMemoizationDependencyMutable
	}
	return messagePreserveManualMemoizationValueUnmemoized
}
