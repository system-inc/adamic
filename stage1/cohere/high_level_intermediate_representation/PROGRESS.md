# Unit 2 clean seam: complete oracle census, straight-line expression lowering

Unit 2 and static-components are **not complete**. This landing extends the first slice and builds the construction denominator from executed Go tests instead of a source-text sample. Static-components has not been registered or claimed green.

## Construction corpus and coverage

At cohere 7945d102a6c18dd36adf9114a758ce646e8b2359, the executed construction census contains **1,465 corpus function graphs; 51 match byte for byte on Node and native Adamic**. Nine additional path probes also match: 60/1,474 with probes included. Function graphs are distinct by source bytes, byte span, checker mode and constructed dump. Equal inputs/results reached by several tests are deduplicated and retain all caller provenance. Nested functions are recorded individually, as well as retained in their parent's dump. A parent is not certified by matching only its nested functions.

Every original Go HIR test runs through a test-file overlay that redirects its `Lower`, `ForFunction` and `ForFunctionWithoutManualMemoization` calls to observers. The original production implementations are called unchanged, their return values are returned unchanged, and the observer constructs/dumps a deep clone. For the memo-erased entry, a separate original Lower is observed because the erased result is not the construction oracle. Production files, test files on disk, and the cohere gitlink are untouched. This captures generated test inputs and their actual checker contexts, including multi-file programs, rather than approximating them by extracting raw strings.

The additional tagged fixture test visits **all 395 vendored upstream fixtures**. Forty Flow fixtures are explicitly excluded using cohere's `RequiresFlow` predicate; the 355 other inputs retain their original script extensions and use real programs with allowJs/JSX enabled. All function-like roots are lowered with the real checker, and nested IR functions are observed recursively. The upstream filenames remain in provenance even where no function can be constructed.

Forty-five original Go tests skip under this environment. Their names and required units are recorded in [SKIPPED.md](SKIPPED.md) and `testdata/construction-summary.json`. All 45 require the absent private Structure corpus; none is outside the eight rules’ reach. They are not silently counted as successes. A test building an IR by hand with no source node is a graph/pass test, not an AST construction input; it still runs, but contributes no source function to the construction denominator. The observer does not claim to export bodies of missing private corpora.

`testdata/construction-summary.json` records the corpus/probe counts, matched graph keys and caller spans, Flow exclusions, skipped Go tests and observed instruction kinds. The complete sources, dumps, per-function manifest and original Go test log are generated at test time under a temporary directory. Set `HIR_CENSUS_EXPORT=/tmp/hir-unit2-census` to retain them for diagnosis; absolute temporary paths are not part of the dump.

## What now lowers

Function declarations, function expressions and block/concise arrows with simple unused identifier parameters; inferred names for anonymous functions assigned directly to variables; numeric, bigint, string, non-substitution template, bool and null literals; parentheses; prefix unary operators, typeof and void; non-short-circuit binary operators; comma expressions; expression and empty statements; terminal returns. Other syntax is declined. Typed/optional/rest/destructured parameters, binding reads and nested function creation remain outside this seam.

Literal strings use the parser's deterministic UTF-16 `written` escape alphabet, including newlines, backslashes and surrogate pairs. Source positions are converted from the port parser's UTF-16 indexes to Go's UTF-8 byte positions before becoming places or instruction spans. Concise arrows return the expression place directly; block returns copy into the shared returns place, matching Go. Identifiers/declarations and parameter versions come from the imported SSA implementation, not a local replacement.

The new coverage driver consumes the entire census and emits either a dump or an explicit decline for every record. The oracle's eligibility predicate describes this slice's grammar; every eligible row must independently parse and lower on the port and compare exactly. It is a scheduling boundary for partial coverage, not a whole-unit pass. Admission cannot turn a mismatch into a decline.

The Go hir-v1 adapter now serializes all instruction/terminal payload variants, phis, nested functions, context declarations and outlined function references. General payloads use sorted JSON field names; AST references are stable kind/span handles. Existing Primitive, LoadLocal, UnaryExpression and BinaryExpression records use the explicit field formats implemented by the slice. Scopes remain `scopes -` at construction; scope analysis has not run. The adapter does not infer or normalize IDs to hide differences.

Eight semantic mutants are caught on both Node and native Adamic: literal boolean flip; binary right operand replaced by left; unary operator replaced by plus; comma returning left; parameter moved to context; return store replaced by nil; concise arrow returning the unused returns slot; inferred function name omitted. They compile and execute successfully before their answers disagree.

## Remaining unit 2 work

1. Complete core variants and graph terminal/edge visitors beyond the one-block adapter. Emit all owned instruction-table entries, including unreachable/orphan instructions, in the oracle before admitting paths that create them. The current dump covers instructions reachable through blocks; its existing complete identifier table already exposes unreachable allocations.
2. Bind locals and captures through the resident checker, including module/import classification and checker-less test behavior. No lexical-only approximation has been introduced. `stage1/cohere/lint/context.ts` already provides `context.checker`; its `Checker.ask` bridge must support the symbol identity/binding queries needed by Go lower.go:1245 onward, rather than inventing local IDs independently of checker resolution.
3. Lower declarations, assignment/update/destructuring, globals/properties/computed accesses, calls/new/spread, object/array/template/regex/type-cast/await/yield forms, JSX, nested function tables, context identifiers, and optional-chain control flow. Give each major path a caught native/Node mutant and an exact corpus dump comparison.
4. Add branches/short-circuit/ternaries, switch, loops/iterators, break/continue/labels and exception/finally terminals. Finish the adapter using the imported SSA graph and construct modules. Keep structural fallthrough separate from real predecessors.
5. Implement the actual rule-context `ForFunction` cache and cheap spelling gate, clone/full visitors, and postdominator/control-dominator analyses. Construct each cached function exactly once and preserve source/checker handles for rule consumers.
6. Reach 1,465/1,465 (or the freshly exported count after a pin/input change), with all declines eliminated or specifically escalated as language/source gaps. Supply private corpora to eliminate relevant environment skips; resolve Flow through the agreed parser boundary rather than claiming it was tested.
7. Port `react-hooks/static-components` into its own `stage1/cohere/lint/rules/<slug>/` directory, following docs/lint-registration.md: concrete listener/factory, exact messages, descriptor, tagged original Go rule/options adapter, raw finding witnesses and a caught rule mutant. Then regenerate the registry and run the discovered comparison harness to green. No shared registry lines or premature rule directory were added at this seam.

The owner confirms mutation_aliasing arrives with tonight's lint-area pin bump; it is not needed for the remaining unit 2 or unit 3 work. No new blocking Adamic language gap was encountered in this seam.

## Binding interface seam

The resident checker now supports the generic `symbol` compiler-facts question:
opaque program-owned identity, name and declaration/import AST metadata. It uses
the shorthand value accessor before the ordinary symbol accessor and bound-node
fallback, matching Go lower.go:1259. Alias symbols keep their identities; resolving
an import to its export would lose the import declaration that lowering needs.
`symbol.ts` owns the validated Adamic reader; `symbol_main.ts` is its independent
compiler-fact witness driver. Facts are not HIR payloads or binding-kind answers.
No production cohere files or gitlink change.

This is a foundation inside the binding seam, **not completed binding lowering**.
Construction remains 51/1,465 plus 9/9 probes. The next step is to export these facts
from the original census checker contexts, wire parameter/local bindings and
module/import classification into lowering, then admit those grammar paths and
compare their hir-v1 dumps. Declarations, assignments/updates, captures and their
mutants still follow before control flow, JSX, ForFunction and static-components.
The symbol-identity-collapse mutant is an additional compiler-fact witness; it
is not counted as a completed lowering-path mutant.

Verification for this seam: `TestResidentSymbolFacts` passes **19 exact selectors**
against independent resident symbol pointer identities (including two shadowed
`value` bindings), across Go, live native Adamic and Node/native replay. Repeating
each Go query preserves bytes and identity. Default, namespace and renamed imports
are all exercised, as are shorthand resolution, absent globals and Unicode byte
spans. Collapsing all nonzero symbol identities to one compiles and runs but changes
the answer on Node and native. `go test ./bridge/tsgo/checker` passes. The complete
construction census and original straight-line regression pass again, with all
eight lowering mutants caught; no construction admission changes in this seam.
