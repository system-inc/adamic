package react

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/static_single_assignment"
)

// messagePurityImpureCallId is the finding's id, and messagePurityImpureCallReason is everything
// after the sentence naming the builtin.
//
// Split rather than held as one `rule.Message` for the reason `globals.go` gives: this rule
// interpolates, the builtin's name is the part a reader needs first, and the invariant half lives
// in one place so a single edit moves every finding. `TestPurityMessageNamesTheBuiltin` asserts the
// rendered result rather than either half, which is the only assertion a format string cannot
// satisfy by accident.
const messagePurityImpureCallId = "impureFunctionCall"

const messagePurityImpureCallReason = "Calling an impure function during render makes the " +
	"component's output depend on when it happened to render rather than only on its props and " +
	"state. React may render a component more than once for a single update, and under Strict " +
	"Mode it deliberately does, so the value changes between renders that should have been " +
	"identical and the interface flickers or resets. Move the call into an event handler or an " +
	"effect, or lift the value into state so it is computed once."

// purityImpureBuiltins is the table, and it is the whole table.
//
// Three entries, which is not a simplification. `grep -c "impure: true"` over
// `oxc_react_compiler/src/react_compiler_hir/globals.rs` returns 3, and the same three appear in
// React's own `DEFAULT_SHAPES` at `eslint-plugin-react-hooks.development.js:30186,30203,30281`,
// each carrying the `canonicalName` this rule prints. Nothing else in either implementation sets
// the flag, so a fourth entry here would be an invention rather than a port.
//
// Keyed by the object name and then the property, because both halves are checked: `Math.floor` is
// pure and `new Date()` is pure, so a rule keyed on the object alone or on the property alone would
// report both. Measured on React, both are silent.
var purityImpureBuiltins = map[string]map[string]string{
	"Math":        {"random": "Math.random"},
	"Date":        {"now": "Date.now"},
	"performance": {"now": "performance.now"},
}

// purityGlobalContainers are the global objects whose properties are themselves the modelled
// globals, so `globalThis.Date.now()` reaches the same signature as `Date.now()`.
//
// Measured rather than assumed, and the asymmetry is real: `globalThis.Date.now()` REPORTS while
// `window.Date.now()` is SILENT, on React's own rule. The cause is in the table rather than in the
// analysis. `globals.rs:1280` registers `globalThis` with `add_object(shapes, ..., typed_globals)`,
// handing it every typed global as a property, and `global` beside it; `window` is not registered
// at all, so `window.Date` loads an unknown property of an unknown object and the chain dies there.
//
// Reproduced rather than improved on. Treating `window` as an alias of the global object would be
// the sensible reading and it is a divergence, because upstream does not do it.
var purityGlobalContainers = map[string]bool{"globalThis": true, "global": true}

// Purity flags a call to a known-impure builtin during render.
//
//	valid:   function Component() { const f = () => Math.random(); return <div onClick={f} />; }
//	valid:   function Component() { useEffect(() => { console.log(Math.random()); }, []); return <div />; }
//	valid:   function Component() { const [x] = useState(() => Math.random()); return <div>{x}</div>; }
//	valid:   function Component() { const Math = {random: () => 1}; return <div>{Math.random()}</div>; }
//	valid:   function Component() { return <div>{Math.floor(1.5)}</div>; }
//	valid:   function helper() { return Math.random(); }
//	invalid: function Component() { const d = Date.now(); return <div>{d}</div>; }
//	invalid: function Component() { const r = Math.random; return <div>{r()}</div>; }
//	invalid: function Component() { const f = () => Math.random(); return <div>{f()}</div>; }
//	invalid: function Component() { const v = useMemo(() => Math.random(), []); return <div>{v}</div>; }
//
// Ported from React's `purity` rule, `ErrorCategory.Purity` in the React Compiler. Both authorities
// were read and React was RUN rather than reasoned about, which is what produced most of what is
// recorded below:
//
//	react 7.1.1   node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js
//	              driven through the ESLint Linter API on 50-plus probe inputs
//	oxc           crates/oxc_react_compiler, emitting at
//	              react_compiler_inference/infer_mutation_aliasing_effects.rs:2425
//
// # The corpus is two files and it cannot see this rule at all
//
// React ships exactly two fixtures whose golden expects this diagnostic, and they are the same
// source twice, differing only in a second pragma:
//
//	function Component() {
//	  const date = Date.now();
//	  const now = performance.now();
//	  const rand = Math.random();
//	  return <Foo date={date} now={now} rand={rand} />;
//	}
//
// Three direct calls, at the top level of a component, with no nesting, no aliasing and no
// conditional. There is not one VALID fixture anywhere in the 650-file corpus containing any of the
// three builtins. So a rule that reported on any lexical occurrence of `Math.random` anywhere in
// any file would score 2 of 2, and a rule that is correct scores 2 of 2, and the corpus cannot tell
// them apart.
//
// That is why almost every sentence below cites a probe rather than a fixture. The corpus is a
// floor here in the weakest possible sense, and treating its pass rate as evidence about this rule
// would be the port brief's "a passing case is not evidence about the thing you think it is",
// arriving through the absence of cases rather than through a misleading one.
//
// # The rule is not a lexical match on three names, and that is the whole design
//
// The obvious port is "find a call whose callee spells one of three builtins, inside a component".
// It is wrong in both directions and each was measured on React's own rule:
//
//	const f = () => Math.random(); const x = f();    REPORTS, at the inner call
//	const f = () => Math.random(); <div onClick={f} />   SILENT, same lexical text
//
// One character of difference in whether the wrapper is CALLED decides it, and the lexical rule
// gets one of the two wrong whichever way it answers. The same pair separates `useMemo`, whose
// callback runs during render and REPORTS, from `useCallback`, whose callback does not and is
// SILENT, with byte-identical callbacks.
//
// And the span moves:
//
//	const r = Math.random; const x = r();            REPORTS, span is `r()`
//	let r = Math.random; r = Date.now; const x = r(); REPORTS `Date.now`, span is `r()`
//
// The second is the one that settles the design. The name printed is `Date.now`, the span is on
// `r`, and no reading of the syntax at the call site can produce either. What upstream is doing is
// following the VALUE: the impure signature is attached to the value loaded out of the global, and
// it travels through assignment, reassignment and control flow to whatever eventually gets called.
// That is single-assignment form with phi nodes, which is exactly what `internal/utilities/hir` is.
//
// So this rule uses the intermediate representation, and the decision was made by measuring the
// alternative rather than by preferring the machinery. Three rules shipped the same day
// deliberately used none of it; this one cannot, and the reassignment probe above is the one-line
// reason.
//
// # What is tracked, stated as the small abstract domain it is
//
// Every value in a lowered function falls in one of four states, and the pass is a single forward
// walk over blocks in reverse postorder assigning them:
//
//	container   the global object `Math`, `Date`, `performance`, or `globalThis`/`global`
//	impure      a value that IS one of the three impure functions, carrying its canonical name
//	callImpure  a nested function that, when called, performs an impure call
//	nothing     everything else
//
// A `MethodCall` on a container whose property names an impure builtin reports. A `CallExpression`
// on an `impure` value reports. A `CallExpression` on a `callImpure` value reports at the site
// recorded inside that function, which is why the nested-arrow finding points at the inner call
// rather than at the call that triggered it, exactly as React does.
//
// Reverse postorder makes one pass correct for everything but a back edge, which is the same
// limitation `static_components.go` records and reproduces.
//
// # Why the type checker is declared when nothing here reads a type
//
// This rule asks no type question. It declares `NeedsTypeChecker` because LOWERING does: without a
// checker, `internal/utilities/hir` resolves every free identifier to `LoadGlobal` carrying a name, and
// a shadowed binding becomes indistinguishable from the real global. Measured directly, with a nil
// checker `function Component() { const Math = {random: () => 1}; return Math.random(); }` lowers
// to `LoadGlobal Math` and this rule would report it; React is SILENT on that input, and so is this
// rule once the checker resolves the local binding. The declaration buys shadow correctness and
// nothing else, and `TestPurityRequiresTheTypedHarness` pins it.
//
// # The gate, taken whole and not re-derived
//
// `IsComponentOrHookLike` comes from the react shelf, written for `unsupported-syntax`. Two porters
// paid seventeen probe rounds and seven false positives for it, and a second copy here would be a
// second thing to keep in agreement with React. It is the right gate rather than a convenient one,
// verified on this rule's
// own diagnostics: `function helper() { return Math.random(); }` is silent, `function Component() {
// return Math.random(); }` is ALSO silent because it creates no JSX and calls no hook, and a class
// `render()` is silent. All three measured on React.
//
// A function the gate declines is still descended into, because a component nested inside a plain
// wrapper is a unit even when the wrapper is not. Measured: a `function Inner()` declared inside
// `function Outer()` reports on its own body.
//
// # Four silences that are upstream's, three reproduced and one a real gap
//
// Each was measured on React's own rule, on the input named. The first is this port falling short
// of upstream; the other three are upstream's own behaviour, reproduced deliberately.
//
//   - **A module-scope helper is opaque.** `const g = () => Math.random(); function Component() {
//     const x = g(); ... }` is SILENT, while the same helper declared INSIDE the component reports.
//     Upstream only knows a function's effects when it lowered that function as part of the unit,
//     and a module-scope one is not part of it. This is the largest hole in the rule and it is
//     upstream's.
//
//   - **`useEffect` and its neighbours freeze their arguments.** `useEffect(() => {
//     Math.random(); }, [])` is SILENT because `globals.rs:1409` attaches
//     `Freeze { value: "@rest" }` to the signature, and a frozen function's effects are not
//     re-raised on the caller. `useLayoutEffect` behaves the same. But `setTimeout(() => {
//     Math.random(); })` REPORTS, because `setTimeout` carries no such signature. Two inputs of
//     identical shape falling opposite ways, decided entirely by a table.
//
//   - **Any hook-named callee freezes a callback argument too.** `useCustom(() => {
//     Math.random(); })` is SILENT while `doThing(() => { Math.random(); })` REPORTS. The
//     discrimination is the hook NAME pattern on the callee, nothing about the function passed.
//
//   - **A property never carries the signature.** `const o = {r: Math.random}; o.r();` is SILENT,
//     though `const r = Math.random; r();` reports. Storing into an object loses the value, and
//     upstream does not chase it back out.
//
// # The one place this port is narrower than upstream, stated because the sweep found it
//
// Upstream reports when an impure value ESCAPES to somewhere it might be called, not only where it
// is called. Measured: `const f = () => Math.random(); doThing(f);` and `f.z = 1;` both REPORT at
// the inner `Math.random`, and so does `const x = f.Date.now();`, while `const x = f.Date;` alone
// is silent. This port re-raises a nested function's effects when the function is CALLED or PASSED
// as a callback argument, which covers the shapes that matter, and says nothing when the value
// escapes by being mutated or read through a property chain.
//
// Closing it means the escape half of the mutation aliasing model, which is the machinery
// `immutability` is blocked on. Guessing at a subset would report inputs upstream passes, so the
// gap is recorded and pinned by a fixture rather than approximated. See the fixture named
// "a property chain off a nested function value".
//
// # Where the two implementations differ, and it is one line
//
// oxc defaults `validate_no_impure_functions_in_render` to FALSE
// (`environment_config.rs:146`), so oxc's compiler is silent on every input unless the pragma turns
// it on, which is why both corpus fixtures carry `// @validateNoImpureFunctionsInRender`. React's
// ESLint plugin defaults it to TRUE (`eslint-plugin-react-hooks.development.js:51788`, inside
// `COMPILER_OPTIONS`), so the rule is live on ordinary source with no pragma at all.
//
// React is the authority and React is what a reader of this codebase runs, so this rule is
// unconditional and reads no pragma. Recorded because it is a real divergence from oxc and the
// differential will show it: any file calling `Date.now()` in a component reports here and does not
// under oxlint. The corpus is unaffected, since its two files set the pragma either way.
var Purity = rule.Rule{
	Name:             "react-hooks/purity",
	NeedsTypeChecker: true,
	// Reads other files only through shape readers (rule.ExportNameIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// The whole file at once, for the reason `static_components.go` gives: lowering already
			// descends into nested functions through the function arena, so listening per function
			// kind would lower every inner function twice, the second time without the bindings its
			// parent supplies.
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				// A file that cannot hold a component or a hook is not lowered at all; see
				// high_level_intermediate_representation.MayHoldComponentOrHook for why that is exact.
				if !high_level_intermediate_representation.MayHoldComponentOrHook(ctx) {
					return
				}
				forEachCompiledFunction(node, func(functionNode *ast.Node) {
					// Shared with the other rules that lower this same function; see high_level_intermediate_representation.ForFunction.
					lowered := high_level_intermediate_representation.ForFunction(ctx, functionNode)
					if lowered == nil {
						return
					}
					purityAnalyzeSubject(ctx, lowered)
				})
			},
		}
	},
}

// purityAnalyzeSubject runs the validator over the functions React would have compiled.
//
// A declined function is descended into rather than skipped, because a component nested inside a
// plain wrapper is still a unit. That is the same shape `set_state_in_render.go` uses and it is
// upstream's traversal rather than a convenience.
func purityAnalyzeSubject(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}
	if function.Node == nil || !utilsreact.IsComponentOrHookLike(function.Node) {
		for _, nested := range function.Functions {
			purityAnalyzeSubject(ctx, nested)
		}
		return
	}
	// A unit found nested in a function that is not one is analyzed as React Compiler compiles it,
	// on its own; see AsCompilationUnit.
	function = high_level_intermediate_representation.AsCompilationUnit(ctx, function,
		high_level_intermediate_representation.ForFunction)
	purityReportImpureCalls(ctx, function, true)
}

// purityValueKind is the small abstract domain described on the rule.
type purityValueKind uint8

const (
	// purityValueNothing is every value the pass has no opinion about.
	purityValueNothing purityValueKind = iota
	// purityValueContainer is a global object whose properties are the modelled globals.
	purityValueContainer
	// purityValueImpure is a value that IS one of the three impure functions.
	purityValueImpure
	// purityValueCallsImpure is a nested function that performs an impure call when invoked.
	purityValueCallsImpure
)

// purityValue is one entry in the abstract state.
//
// `CanonicalName` is what the finding prints and `Range` is where it points. For an impure function
// the range is unused, because the finding points at the CALL rather than at the load; for a nested
// function that calls one, the range is the inner call site, which is where React points.
type purityValue struct {
	Kind          purityValueKind
	CanonicalName string
	Node          *ast.Node
	// Function is the nested function this value holds, set only for purityValueCallsImpure.
	Function *high_level_intermediate_representation.Function
}

// purityFindings is what one lowered function does when called, for the caller to re-raise.
type purityFinding struct {
	CanonicalName string
	Node          *ast.Node
}

// purityReportImpureCalls is the forward walk, and it returns what this function does when called.
//
// `emit` separates reporting from detecting. A nested function is analysed to learn whether calling
// it is impure, and that analysis must NOT report on its own: a `Math.random` inside an uncalled
// arrow is silent upstream, and the arrow is analysed either way. Only the enclosing unit reports,
// and only for a nested function it actually calls.
//
// The return value is the findings this function raises when invoked, which the caller records
// against the value holding it.
//
// State is per UNIT rather than per function, held in `purityUnit`. A nested function reads a
// binding from its enclosing scope through `LoadGlobal` when the checker resolved it to a
// declaration lowering did not give a local slot, so the wrapper in `const g = () => Math.random();
// const f = () => g(); f();` sees `g` only if the enclosing values are visible to it. Measured on
// React, that input REPORTS, and with per-function state it did not.
func purityReportImpureCalls(ctx rule.Context, function *high_level_intermediate_representation.Function, emit bool) []purityFinding {
	unit := &purityUnit{
		ctx:       ctx,
		values:    map[static_single_assignment.IdentifierId]purityValue{},
		byName:    map[string]purityValue{},
		nested:    map[*high_level_intermediate_representation.Function][]purityFinding{},
		reported:  map[*ast.Node]bool{},
		analysing: map[*high_level_intermediate_representation.Function]bool{},
	}
	return unit.walk(function, emit)
}

// purityUnit is the analysis state for one compiled unit, shared by every function inside it.
type purityUnit struct {
	ctx rule.Context
	// values is the abstract state keyed by value. Single-assignment form makes one entry per
	// value rather than per binding, which is what lets a forward walk be correct without a
	// fixpoint.
	values map[static_single_assignment.IdentifierId]purityValue
	// byName is the same state keyed by SOURCE NAME, which is how a nested function reaches a
	// binding declared in its parent. Lowering gives such a reference a `LoadGlobal` carrying the
	// name rather than a local place, so the identifier in the child is not the identifier in the
	// parent and only the name connects them.
	byName map[string]purityValue
	// nested is what each nested function does when called, computed on first call so an uncalled
	// one costs nothing.
	nested map[*high_level_intermediate_representation.Function][]purityFinding
	// analysing guards recursion. Reachable through callbackArguments: a function passed as a
	// callback to something it calls re-enters its own frame. See findingsOf.
	analysing map[*high_level_intermediate_representation.Function]bool
	// reported is the set of nodes already reported in this unit. A wrapper called twice raises
	// the same inner finding twice and React reports it ONCE, because the finding belongs to the
	// site that raised it rather than to each caller.
	reported map[*ast.Node]bool
}

// walk runs the forward pass over one function and returns what calling it does.
func (unit *purityUnit) walk(function *high_level_intermediate_representation.Function, emit bool) []purityFinding {
	if function == nil {
		return nil
	}

	values := unit.values
	var raised []purityFinding

	// report records a finding, and is the only place the diagnostic is built. When `emit` is
	// false the finding is still collected, because the caller needs it to decide what calling
	// this function does.
	report := func(finding purityFinding) {
		raised = append(raised, finding)
		if !emit || finding.Node == nil || unit.reported[finding.Node] {
			return
		}
		unit.reported[finding.Node] = true
		unit.ctx.ReportNode(finding.Node, rule.Message{
			Id: messagePurityImpureCallId,
			Description: fmt.Sprintf("`%s` is an impure function. %s",
				finding.CanonicalName, messagePurityImpureCallReason),
		})
	}

	for _, block := range function.Blocks {
		// A phi carries a value in from more than one path. Upstream's state merges at the join,
		// so a value that is impure on ANY incoming path is impure at the phi. Measured on
		// React: `let r = Math.random; if (c) { r = Date.now; } r();` reports, and the name it
		// prints is whichever path the merge kept.
		for _, phi := range block.Phis {
			// Which operand wins is invisible in a rule whose phi merge only decides a boolean, as
			// in `static_components.go`, and it is NOT invisible here: the operand chosen is the
			// builtin the message NAMES. Measured on
			// `let r = Math.random; if (c) { r = Date.now; } r();`, an unordered walk produced
			// `Math.random` on some runs and `Date.now` on others, and React always says
			// `Date.now`.
			//
			// So the predecessors are visited in block order and the LAST impure one wins, which is
			// the operand from the latest-numbered predecessor. Blocks are in reverse postorder, so
			// that is the write that appears last in the source, which is what React prints.
			var merged purityValue
			var mergedFrom static_single_assignment.BlockId
			for _, entry := range phi.Operands {
				incoming, ok := values[entry.Place.Identifier]
				if !ok || incoming.Kind == purityValueNothing {
					continue
				}
				if merged.Kind == purityValueNothing || entry.Predecessor > mergedFrom {
					merged, mergedFrom = incoming, entry.Predecessor
				}
			}
			if merged.Kind != purityValueNothing {
				values[phi.Place.Identifier] = merged
			}
		}

		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			unit.step(function, instruction, report)
		}
	}

	return raised
}

// purityStep interprets one instruction against the abstract state.
//
// Split out of the block walk so the type switch reads as the transfer function it is, one arm per
// instruction shape that moves a value.
func (unit *purityUnit) step(
	function *high_level_intermediate_representation.Function,
	instruction *high_level_intermediate_representation.Instruction,
	report func(purityFinding),
) {
	values := unit.values
	switch value := instruction.Value.(type) {
	case *high_level_intermediate_representation.LoadGlobal:
		// Only a TRUE global can be one of these. An import or a module-scope binding of the same
		// name is a different thing and upstream models neither, which is what makes
		// `import Math from './m'` silent.
		if !unit.namesATrueGlobal(instruction.Node) {
			return
		}
		if _, isImpureContainer := purityImpureBuiltins[value.Name]; isImpureContainer {
			values[instruction.LValue.Identifier] = purityValue{Kind: purityValueContainer, CanonicalName: value.Name}
			return
		}
		if purityGlobalContainers[value.Name] {
			// `globalThis` itself carries no name to print; it is transparent, and the property
			// loaded off it is what names the container.
			values[instruction.LValue.Identifier] = purityValue{Kind: purityValueContainer}
			return
		}
		// Not a modelled global, so this may be a binding from an enclosing function that lowering
		// could not give a local slot. `byName` is the only link between the two, because the
		// identifier in the child function is not the identifier in the parent.
		if known, ok := unit.byName[value.Name]; ok {
			values[instruction.LValue.Identifier] = known
		}

	case *high_level_intermediate_representation.PropertyLoad:
		purityLoadProperty(function, instruction, values, value.Object, value.Property)

	case *high_level_intermediate_representation.Destructure:
		// `const {random} = Math;` reports upstream, so destructuring has to carry the signature
		// the same way a property load does. The pattern names the properties taken.
		purityDestructure(function, instruction, values, value)

	case *high_level_intermediate_representation.MethodCall:
		purityMethodCall(function, instruction, values, value, report)
		unit.callbackArguments(function, value.Args, purityCalleeName(function, value.Property), report)

	case *high_level_intermediate_representation.CallExpression:
		unit.callExpression(function, instruction, value, report)
		unit.callbackArguments(function, value.Args, purityCalleeIdentifierName(function, value.Callee), report)

	case *high_level_intermediate_representation.FunctionExpression:
		// Recorded but NOT analysed here. Analysis happens on first CALL, which is what keeps an
		// uncalled arrow silent.
		if int(value.Function) < len(function.Functions) {
			values[instruction.LValue.Identifier] = purityValue{
				Kind:     purityValueCallsImpure,
				Function: function.Functions[value.Function],
			}
		}

	case *high_level_intermediate_representation.StoreLocal:
		purityCopy(values, value.Value.Identifier, value.LValue.Identifier, instruction.LValue.Identifier)
		unit.recordByName(function, value.LValue, value.Value)

	case *high_level_intermediate_representation.LoadLocal:
		purityCopy(values, value.Place.Identifier, instruction.LValue.Identifier)

	case *high_level_intermediate_representation.StoreContext:
		purityCopy(values, value.Value.Identifier, value.LValue.Identifier, instruction.LValue.Identifier)
		unit.recordByName(function, value.LValue, value.Value)

	case *high_level_intermediate_representation.LoadContext:
		// A read of a binding captured from an enclosing function. The identifier is the ENCLOSING
		// function's, because `Context` places are the parent's values threaded in, so the ordinary
		// copy already crosses the boundary when the parent recorded that value. The name-keyed
		// fallback covers the case where it did not, which is a binding lowering gave no local slot.
		if !purityCopy(values, value.Place.Identifier, instruction.LValue.Identifier) {
			unit.loadByName(function, value.Place, instruction)
		}

	case *high_level_intermediate_representation.StoreGlobal:
		// A reassignment through a global binding, which is how `let r = Math.random; r = Date.now`
		// lowers when the checker cannot give the binding a local identity. The value still has to
		// travel, and this is the instruction carrying it.
		purityCopy(values, value.Value.Identifier, instruction.LValue.Identifier)
	}
}

// purityCopy propagates one value's state onto one or more destinations, and reports whether the
// source had any state to propagate.
//
// # The `purityValueNothing` half of the test is EQUIVALENT, and it is kept deliberately
//
// A mutation dropping it SURVIVED. That is not a fixture gap: the map never holds a
// `purityValueNothing` entry, so presence in the map and having state are the same question.
// Established by enumerating every write into `values` and `byName` — six construction sites, all
// of which name `purityValueContainer`, `purityValueImpure` or `purityValueCallsImpure`, plus five
// propagation sites, all of which copy a value already guarded here or at the phi merge. No path
// constructs the zero value and stores it.
//
// Kept rather than deleted because it is the invariant's only statement in code, and the zero value
// of the enum IS `purityValueNothing`, so a future arm that stores a partly-built value would break
// silently without it. Recorded as measured-equivalent rather than as a live guard, so the next
// reader does not go looking for the input that separates them.
func purityCopy(values map[static_single_assignment.IdentifierId]purityValue, from static_single_assignment.IdentifierId, to ...static_single_assignment.IdentifierId) bool {
	source, ok := values[from]
	if !ok || source.Kind == purityValueNothing {
		return false
	}
	for _, destination := range to {
		values[destination] = source
	}
	return true
}

// namesATrueGlobal asks the checker whether an identifier binds to a real global rather than to
// something declared in this program.
//
// The `BindingKind` on a `LoadGlobal` cannot answer this. Lowering sets it to
// `GlobalBindingKindGlobal` unconditionally (`lower_expression.go:179`, whose own comment says
// distinguishing an import remains a later pass), so filtering on it is inert: measured, a resolved
// `import Math from './m'` arrives with kind Global and would report.
//
// So the question is asked of the checker directly. A true global is declared only in a declaration
// file, which is where the standard library lives; a module-scope binding, an import, or a
// declaration anywhere in this program has a declaration in a real source file and is not ours.
// Measured on React, `import Math from './m'` is SILENT, and so is a module-scope `const Math`.
//
// Every declaration is examined rather than the first. The question is "is this name declared
// anywhere in the program", so more declarations only mean more chances to disqualify it, and
// indexing `Declarations[0]` would answer about whichever one merging happened to put first.
// A name the checker cannot resolve at all is a true global: that is `Math` in a file with no
// standard library, which is what the fixture harness produces.
func (unit *purityUnit) namesATrueGlobal(node *ast.Node) bool {
	if node == nil {
		return false
	}
	identifier := node
	if !ast.IsIdentifier(identifier) {
		return false
	}
	return !rule.IsDeclaredInASourceFile(unit.ctx.TypeChecker.GetSymbolAtLocation(identifier))
}

// loadByName recovers a captured binding's state through its source name.
//
// Reached only when the identifier itself carried nothing. A nested function's `Context` place is
// minted in the child, so `const g = () => Math.random(); const f = () => g(); f();` gives the `g`
// inside `f` an identifier the parent never wrote state against, and only the name connects them.
// Measured on React, that input REPORTS.
func (unit *purityUnit) loadByName(function *high_level_intermediate_representation.Function, place high_level_intermediate_representation.Place, instruction *high_level_intermediate_representation.Instruction) {
	if int(place.Identifier) >= len(function.Identifiers) {
		return
	}
	name := function.Identifiers[place.Identifier].Name
	if name == "" {
		return
	}
	if known, ok := unit.byName[name]; ok {
		unit.values[instruction.LValue.Identifier] = known
	}
}

// purityLoadProperty resolves `container.property` to an impure signature where one exists.
//
// Two shapes reach here and both were measured. `Math.random` off the container is the impure
// function itself. `globalThis.Date` off a transparent container is another container, which is
// what makes `globalThis.Date.now()` report while `window.Date.now()` does not.
func purityLoadProperty(
	function *high_level_intermediate_representation.Function,
	instruction *high_level_intermediate_representation.Instruction,
	values map[static_single_assignment.IdentifierId]purityValue,
	object high_level_intermediate_representation.Place,
	property string,
) {
	container, ok := values[object.Identifier]
	if !ok || container.Kind != purityValueContainer {
		return
	}
	if container.CanonicalName == "" {
		// A transparent container: the property names the real one, if it is one of ours.
		if _, isImpureContainer := purityImpureBuiltins[property]; isImpureContainer {
			values[instruction.LValue.Identifier] = purityValue{Kind: purityValueContainer, CanonicalName: property}
		}
		return
	}
	if canonical, isImpure := purityImpureBuiltins[container.CanonicalName][property]; isImpure {
		values[instruction.LValue.Identifier] = purityValue{Kind: purityValueImpure, CanonicalName: canonical}
	}
}

// purityDestructure carries the signature through `const {random} = Math`.
func purityDestructure(
	function *high_level_intermediate_representation.Function,
	instruction *high_level_intermediate_representation.Instruction,
	values map[static_single_assignment.IdentifierId]purityValue,
	value *high_level_intermediate_representation.Destructure,
) {
	container, ok := values[value.Value.Identifier]
	if !ok || container.Kind != purityValueContainer || container.CanonicalName == "" {
		return
	}
	properties := purityImpureBuiltins[container.CanonicalName]
	if properties == nil {
		return
	}
	object, isObject := value.Pattern.(*high_level_intermediate_representation.ObjectPattern)
	if !isObject {
		return
	}
	for _, property := range object.Properties {
		// The table lookup is the only guard needed, and the two that used to sit around it were
		// both measured redundant rather than removed on taste.
		//
		// A computed key was declined explicitly first. `ObjectPatternProperty.Key` is documented
		// as empty when `ComputedKey` is set, and the empty string is not a key in any of these
		// maps, so the lookup already declines it. A mutation neutralizing that guard SURVIVED and
		// then produced identical output on `const {[props.k]: r} = Math; r();`, which is silent
		// both ways and silent upstream.
		//
		// A nested pattern was declined explicitly too, on the reasoning that
		// `const {random: {call}} = Math` binds something other than the function. Also subsumed:
		// a non-place pattern has no single identifier to write against, and the mutation
		// substituting a zero pattern wrote to identifier 0, which is the reserved invalid value
		// nothing loads. Measured identical on that exact input.
		canonical, isImpure := properties[property.Key]
		if !isImpure {
			continue
		}
		place, isPlace := property.Value.(*high_level_intermediate_representation.PlacePattern)
		if !isPlace {
			continue
		}
		values[place.Place.Identifier] = purityValue{Kind: purityValueImpure, CanonicalName: canonical}
	}
}

// purityMethodCall handles `Math.random()`, where the receiver and the property arrive together.
//
// Lowering keeps a method call whole rather than splitting it into a property load plus a call,
// because the receiver is the `this` of the call. So this arm is the direct form and it is the one
// both corpus fixtures exercise. The property is a `Primitive` holding the name, which is how a
// computed `Math['random']()` reaches the same place: measured on React, that input REPORTS.
func purityMethodCall(
	function *high_level_intermediate_representation.Function,
	instruction *high_level_intermediate_representation.Instruction,
	values map[static_single_assignment.IdentifierId]purityValue,
	value *high_level_intermediate_representation.MethodCall,
	report func(purityFinding),
) {
	container, ok := values[value.Receiver.Identifier]
	if !ok || container.Kind != purityValueContainer || container.CanonicalName == "" {
		return
	}
	property, named := purityPropertyName(function, value.Property)
	if !named {
		return
	}
	if canonical, isImpure := purityImpureBuiltins[container.CanonicalName][property]; isImpure {
		report(purityFinding{CanonicalName: canonical, Node: instruction.Node})
	}
}

// callExpression handles a call through a VALUE rather than through a receiver.
//
// Two things can be called impurely. An `impure` value is one of the three builtins reached through
// an alias, and it reports at the call. A `callsImpure` value is a nested function, and calling it
// re-raises whatever that function does, pointing at the INNER call site rather than at this one,
// which is what React does and is why `const f = () => Math.random(); f();` reports on line 2.
//
// The nested function is held as a POINTER rather than as a `FunctionId`, because an id is an index
// into the declaring function's own arena and a wrapper can be reached from a sibling scope where
// that index means something else. The pointer is stable wherever the value travels.
func (unit *purityUnit) callExpression(
	function *high_level_intermediate_representation.Function,
	instruction *high_level_intermediate_representation.Instruction,
	value *high_level_intermediate_representation.CallExpression,
	report func(purityFinding),
) {
	callee, ok := unit.values[value.Callee.Identifier]
	if !ok {
		return
	}
	switch callee.Kind {
	case purityValueImpure:
		report(purityFinding{CanonicalName: callee.CanonicalName, Node: instruction.Node})
	case purityValueCallsImpure:
		for _, finding := range unit.findingsOf(callee.Function) {
			report(finding)
		}
	}
}

// callbackArguments re-raises the effects of a function passed as an argument.
//
// # Why an argument counts as a call at all
//
// A function handed to another function is usually invoked, and upstream models that: measured on
// React, every one of these REPORTS at the inner `Math.random`, none of them naming the outer call:
//
//	props.items.map(() => Math.random())
//	doThing(() => { Math.random(); })
//	window.addEventListener('x', () => { Math.random(); })
//	props.p.then(() => { Math.random(); })
//	<li key={Math.random()} />                     (the same shape inside a map callback)
//
// Without this, the rule misses the single most common real purity bug, which is a random key on a
// mapped list. That was found by running the rule on realistic shapes rather than on the corpus,
// which contains none of them.
//
// # What is exempt, and the exemption is a hook-name test with one carve-out
//
// A hook freezes its callback, so its effects are not re-raised on the caller. `useEffect` does it
// through an explicit `Freeze { value: "@rest" }` on its signature (`globals.rs:1409`), and any
// callee whose name matches the hook pattern gets the same treatment: measured, `useCustom(() =>
// Math.random())` is SILENT while `doThing` with a byte-identical callback REPORTS.
//
// `useMemo` is the one exception, and it is a semantic one rather than a special case: its callback
// is the only hook callback that runs DURING render, so its impurity is render impurity. Measured,
// `useMemo(() => Math.random(), [])` REPORTS while `useCallback(() => Math.random(), [])` with the
// same callback is SILENT, and so are `useState`, `useRef`, `useLayoutEffect` and
// `useInsertionEffect` with a lazy initializer.
//
// So the test is: a hook-named callee exempts its callback arguments, unless it is `useMemo`.
func (unit *purityUnit) callbackArguments(
	function *high_level_intermediate_representation.Function,
	args []high_level_intermediate_representation.Argument,
	calleeName string,
	report func(purityFinding),
) {
	if calleeName != "useMemo" && utilsreact.IsCompilerHookName(calleeName) {
		return
	}
	for _, argument := range args {
		// A spread argument is not skipped explicitly, and the guard that did so was measured
		// SUBSUMED by the kind test below: a spread's place is the ITERABLE, so it can never hold
		// `purityValueCallsImpure`, which is only ever written against a `FunctionExpression`
		// result. A mutation neutralizing the skip changed no output.
		//
		// Recorded because upstream does reach through it and this port does not.
		// `const fns = [() => Math.random()]; doThing(...fns);` REPORTS on React and is silent
		// here, for the same reason `const o = {r: Math.random}` is: a value stored into a
		// container is not followed back out. It is the object-property gap arriving through an
		// array rather than a new one.
		// The kind test is EQUIVALENT and kept for legibility rather than for behaviour. Only a
		// `FunctionExpression` result ever gets a non-nil `Function`, so a container or an impure
		// value reaching here hands `findingsOf` a nil and gets nil back. Measured on
		// `doThing(Math.random)`, `doThing(r)` after `const r = Math.random`, and `doThing(m)`
		// after `const m = Math`: all three silent with the test and without it.
		held, ok := unit.values[argument.Place.Identifier]
		if !ok || held.Kind != purityValueCallsImpure {
			continue
		}
		for _, finding := range unit.findingsOf(held.Function) {
			report(finding)
		}
	}
}

// purityCalleeName reads a method call's property name, for the hook test.
func purityCalleeName(function *high_level_intermediate_representation.Function, property high_level_intermediate_representation.Place) string {
	name, _ := purityPropertyName(function, property)
	return name
}

// purityCalleeIdentifierName reads a plain call's callee name from the value that defined it.
//
// A hook is normally called through a bare identifier, which lowering gives a `LoadGlobal` carrying
// the name whether it is imported or global, so this reads that instruction rather than asking the
// checker.
func purityCalleeIdentifierName(function *high_level_intermediate_representation.Function, callee high_level_intermediate_representation.Place) string {
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil || instruction.LValue.Identifier != callee.Identifier {
				continue
			}
			if global, isGlobal := instruction.Value.(*high_level_intermediate_representation.LoadGlobal); isGlobal {
				return global.Name
			}
			return ""
		}
	}
	return ""
}

// findingsOf analyses a nested function on its first call and remembers the answer.
//
// # The recursion guard is REACHABLE, and the way that was established is the point
//
// A guard was written here first. A mutation neutralizing it survived the whole fixture set, and
// probing three self-referential shapes — `const f = () => f();`, a mutually recursive pair, and a
// self-calling wrapper containing `Date.now()` — showed the in-progress branch never firing. It was
// deleted as unreachable, with that measurement recorded at the line.
//
// It was wrong, and the real tree found it: a dry run over Kirk's 3,407 files died with
// `fatal error: stack overflow`, and the cycle in the trace is
// `findingsOf -> walk -> step -> callbackArguments -> findingsOf` on one `*high_level_intermediate_representation.Function` pointer.
//
// Two things made the earlier verdict false. The probes predated `callbackArguments`, so the only
// path into `findingsOf` was a direct call, and a self-reference on that path resolves through
// `byName`, which the enclosing store has not written yet. Passing a function to something that
// receives it — which is what an argument is — has no such ordering, so a function reaching its own
// value as a callback argument re-enters the frame it is in.
//
// The general lesson, which is why this is written out rather than fixed quietly: an unreachability
// verdict is only as good as the paths that existed when it was measured, and adding a caller
// invalidates it silently. No fixture went red when the guard was removed, and none went red when
// the caller was added; only real code found it.
func (unit *purityUnit) findingsOf(function *high_level_intermediate_representation.Function) []purityFinding {
	if function == nil {
		return nil
	}
	if findings, analysed := unit.nested[function]; analysed {
		return findings
	}
	if unit.analysing[function] {
		return nil
	}
	unit.analysing[function] = true
	findings := unit.walk(function, false)
	delete(unit.analysing, function)
	unit.nested[function] = findings
	return findings
}

// recordByName remembers a named binding's state so a nested function can reach it.
//
// Only a place carrying a source name is recorded, because the link this restores is exactly the
// one lowering drops: a reference from a child function to a parent's binding arrives as a
// `LoadGlobal` carrying that name and nothing else.
func (unit *purityUnit) recordByName(function *high_level_intermediate_representation.Function, target high_level_intermediate_representation.Place, source high_level_intermediate_representation.Place) {
	if int(target.Identifier) >= len(function.Identifiers) {
		return
	}
	name := function.Identifiers[target.Identifier].Name
	if name == "" {
		return
	}
	if known, ok := unit.values[source.Identifier]; ok && known.Kind != purityValueNothing {
		unit.byName[name] = known
	}
}

// purityPropertyName reads the static name out of a method call's property place.
//
// The property of a `MethodCall` is a place holding a `Primitive` whose value is the name, so the
// instruction that DEFINED it is what has to be read. A dotted access lowers the name directly
// (`lower_expression.go:218`); an element access lowers the argument expression, so `Math[k]()`
// arrives here as whatever `k` lowered to. Declining a non-`Primitive` is what makes a genuinely
// dynamic key silent, which is measured: React passes `Math[props.k]()`.
//
// # There is no separate string test, and that is measured rather than an oversight
//
// The obvious form returns `primitive.Value.(string)`'s ok flag as well. A mutation ignoring that
// flag SURVIVED, and it is subsumed rather than a fixture gap: a failed assertion yields the empty
// string, the empty string is not a key in `purityImpureBuiltins` or in any of its inner maps, and
// the only consumers of this name are those two lookups. So a numeric or boolean key reaches
// silence either way. Confirmed on `(Math as any)[0]()` and `(Math as any)[true]()`, both silent
// with the flag and without it.
func purityPropertyName(function *high_level_intermediate_representation.Function, property high_level_intermediate_representation.Place) (string, bool) {
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil || instruction.LValue.Identifier != property.Identifier {
				continue
			}
			primitive, isPrimitive := instruction.Value.(*high_level_intermediate_representation.Primitive)
			if !isPrimitive {
				return "", false
			}
			name, _ := primitive.Value.(string)
			return name, true
		}
	}
	return "", false
}
