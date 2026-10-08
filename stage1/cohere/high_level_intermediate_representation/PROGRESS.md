# Unit 2 local checkpoint: arena-based closures

Unit 2 and static-components are **not complete**. This landing extends the first slice and builds the construction denominator from executed Go tests instead of a source-text sample. Static-components has not been registered or claimed green.

## Current native certificate

The Oct 8 arena ruling resolves the nested-function ownership refusal. Both backends
now match **280/1,465 corpus functions**, with **37/37 probes** (317/1,502 including
probes). Native coverage rises by **58**, meeting Node coverage. Active non-Flow
coverage is 280/1,442. All 34 semantic lowering mutants compile, execute and disagree
on both backends, and an additional off-by-one FunctionIndex mutant stops loudly at
an arena read on both. The checker rejects FunctionIndex as BlockIndex.

The first full local run found one retained mutant invalidated by nested-root census
paths: dropping all post-abrupt statements removed the function a path selected. The
replacement moves post-abrupt instructions into the entry block while retaining the
nested table; the targeted rerun repeats the entire Go/Node/native census and passes
that corrected mutant and the index mutant. The retained regression and symbol tests
passed in the full run. See EVIDENCE.md for commands and the original diagnostics.

The function graph owner retains the HIRArena and root handle together. Nested
function targets are FunctionIndex values, with Go-local ordinals retained only for
dump compatibility. Each function owns a typed block table and BlockIndex order;
control-flow successors and predecessor edges are numeric handles. Nominal numeric
enums brand function, block, instruction, identifier and declaration indices. Every
function/block read validates its index; existing instruction/identifier reads stop
on missing entries. No node contains a reference back to its arena or another HIR
function. Context analysis is a named recursive method, without a self-held function
value. The ruling is recorded beside Arenas in docs/memory.md.

Unit 2, remaining flow/expression variants, JSX, ForFunction integration and
static-components remain unfinished. The plan branch is not pushed until the unit
passes. The earlier gap evidence and three-line reproducer are visible at
54ad652e6795fef214582ee3d0348157243ad545 on stage1-hir/wip.

## Construction corpus and coverage

At cohere 7945d102a6c18dd36adf9114a758ce646e8b2359, the executed construction census contains **1,465 corpus function graphs; 280 match byte for byte on Node and native Adamic**. Thirty-seven additional path probe graphs also match: 317/1,502 with probes included. Function graphs are distinct by source bytes, byte span, checker mode and constructed dump. Equal inputs/results reached by several tests are deduplicated and retain all caller provenance. Nested functions are recorded individually, as well as retained in their parent's dump. A parent is not certified by matching only its nested functions.

Every original Go HIR test runs through a test-file overlay that redirects its `Lower`, `ForFunction` and `ForFunctionWithoutManualMemoization` calls to observers. The original production implementations are called unchanged, their return values are returned unchanged, and the observer constructs/dumps a deep clone. For the memo-erased entry, a separate original Lower is observed because the erased result is not the construction oracle. Production files, test files on disk, and the cohere gitlink are untouched. This captures generated test inputs and their actual checker contexts, including multi-file programs, rather than approximating them by extracting raw strings.

The additional tagged fixture test visits **all 395 vendored upstream fixtures**. Forty Flow fixtures are explicitly excluded using cohere's `RequiresFlow` predicate; the 355 other inputs retain their original script extensions and use real programs with allowJs/JSX enabled. All function-like roots are lowered with the real checker, and nested IR functions are observed recursively. The upstream filenames remain in provenance even where no function can be constructed.

Forty-five original Go tests skip under this environment. Their names and required units are recorded in [SKIPPED.md](SKIPPED.md) and `testdata/construction-summary.json`. All 45 require the absent private Structure corpus; none is outside the eight rules’ reach. They are not silently counted as successes. A test building an IR by hand with no source node is a graph/pass test, not an AST construction input; it still runs, but contributes no source function to the construction denominator. The observer does not claim to export bodies of missing private corpora.

`testdata/construction-summary.json` records the corpus/probe counts, matched graph keys and caller spans, Flow exclusions, skipped Go tests and observed instruction kinds. The complete sources, dumps, per-function manifest and original Go test log are generated at test time under a temporary directory. Set `HIR_CENSUS_EXPORT=/tmp/hir-unit2-census` to retain them for diagnosis; absolute temporary paths are not part of the dump.

## What now lowers

Function declarations, function expressions and block/concise arrows with simple identifier parameters; inferred names for anonymous functions assigned directly to variables; numeric, bigint, string, non-substitution template, bool and null literals; parentheses; prefix unary operators, typeof and void; non-short-circuit binary operators; comma expressions; expression and empty statements; terminal returns; resident-symbol parameter/local reads, let/const/var declarations, assignments and compound assignments, prefix/postfix updates, global/module/import loads, global stores and property/computed loads and stores. Other syntax is declined. Blocks, if/else, returns/throws (including subsequent unreachable code), ternaries, logical &&/||/??, while loops and unlabeled break/continue are also admitted. Call/new expressions, named and computed method calls, spread arguments and direct optional calls are admitted. Typed/optional/rest/destructured parameters and nested function creation remain outside this seam.

Literal strings use the parser's deterministic UTF-16 `written` escape alphabet, including newlines, backslashes and surrogate pairs. Source positions are converted from the port parser's UTF-16 indexes to Go's UTF-8 byte positions before becoming places or instruction spans. Concise arrows return the expression place directly; block returns copy into the shared returns place, matching Go. Identifiers/declarations and parameter versions come from the imported SSA implementation, not a local replacement.

The new coverage driver consumes the entire census and emits either a dump or an explicit decline for every record. The oracle's eligibility predicate describes this slice's grammar; every eligible row must independently parse and lower on the port and compare exactly. It is a scheduling boundary for partial coverage, not a whole-unit pass. Admission cannot turn a mismatch into a decline.

The Go hir-v1 adapter now serializes all instruction/terminal payload variants, phis, nested functions, context declarations and outlined function references, including orphan instructions retained in the complete table. General payloads use sorted JSON field names; AST references are stable kind/span handles. Existing Primitive, LoadLocal, UnaryExpression and BinaryExpression records use the explicit field formats implemented by the slice. Scopes remain `scopes -` at construction; scope analysis has not run. The adapter does not infer or normalize IDs to hide differences.

The original eight semantic mutants are retained on both Node and native Adamic: literal boolean flip; binary right operand replaced by left; unary operator replaced by plus; comma returning left; parameter moved to context; return store replaced by nil; concise arrow returning the unused returns slot; inferred function name omitted. They compile and execute successfully before their answers disagree.

## Remaining unit 2 work

1. Complete the remaining core variants and terminal/edge visitors. The multi-block adapter and complete instruction-table dump now cover the admitted control-flow paths.
2. Complete captures and contextual binding reads/writes through the resident checker. Local/parameter reads, versioned declarations, module/import classification and checker-less test behavior now match the admitted census. No lexical-only approximation has been introduced. `stage1/cohere/lint/context.ts` already provides `context.checker`; its `Checker.ask` bridge must support the symbol identity/binding queries needed by Go lower.go:1245 onward, rather than inventing local IDs independently of checker resolution.
3. Lower destructuring, object/array/template/regex/type-cast/await/yield forms, JSX, nested function tables, context identifiers, and optional-chain control flow. Give each major path a caught native/Node mutant and an exact corpus dump comparison.
4. Add switch, do/for/iterator loops, labeled break/continue and exception/finally terminals. If/else, ternaries, short-circuit expressions, while and unlabeled jumps now match. Finish the adapter using the imported SSA graph and construct modules. Keep structural fallthrough separate from real predecessors.
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

## Statement and expression lowering seam

**100/1,465 corpus functions match Go on Node and native Adamic, up from 51**.
Fourteen path probes match as well (114/1,479 raw census graphs). Sixteen semantic
lowering mutants compile, run successfully and disagree on both runtimes. The
prior twelve straight-line regression cases still match. The symbol-facts witness
now covers 22 selectors, including computed declaration names, and its independent
identity-collapse mutant is caught.

This seam wires symbol identity into parameter/local reads, declaration versions,
let/const/var declarations, assignments, compound assignments and prefix/postfix
updates. It lowers global/module/import reads, global stores, and named/computed
property reads/stores. Assignments lower their RHS before evaluating a property
receiver; instruction temporaries remain distinct from named binding definitions.
All renaming continues through imported SSA. No lexical binding resolver is used.
The census preserves the actual original multi-file checker context by exporting
compiler-fact snapshots independently of the IR; absent facts refuse exact spans.
The live checker adapter is separate from the pure snapshot reader, so replay does
not link unused checker-library calls.

Eight new lowering mutants cover local-symbol resolution, import provenance,
declaration kind, assignment result, compound operator, update operation, named
property load and computed key. Five probes supplement real corpus paths. General
payload strings now use Go encoding/json-compatible escaping, including HTML
characters, control bytes and U+2028/U+2029.

The original Go score/pruning tests reload Flow fixtures outside the fixture loader:
**23 Flow graph records** were present in the existing denominator, but none was
previously matched. They now carry explicit `excluded: Flow` metadata and full
provenance in construction-summary.json. The original 40 fixture-loader exclusions
stay unchanged. For continuity we retain the raw 1,465 denominator; the active
non-Flow denominator is **1,442**. This is the agreed Flow exclusion, not a parser
failure converted to successful coverage. `summarize_census.py` regenerates evidence
after the complete comparison passes.

Still ahead: structured control flow and its complete table dump; calls/new,
arrays/objects/templates/casts, destructuring, closures/context bindings, optional
chains, JSX, ForFunction integration and static-components. This is a landable
partial statement/expression seam, not a claim that unit 2 or any rule is complete.

## Initial control-flow seam

**112/1,465 corpus functions match Go on Node and native Adamic**, a further
increase of **12** over the statement/expression seam (and 61 over its predecessor).
Twenty probes also match: **132/1,485** with probes included. The active non-Flow
coverage is 112/1,442; the 23 catalogued Flow reloads and 40 fixture-loader exclusions
remain unchanged. The original 45 private-corpus skips retain their unit assignments.

This seam adds blocks, if/else, returns/throws, code after abrupt exits, conditional
expressions, &&/||/??, while loops and unlabeled break/continue. Function-local block
IDs and the lookup table survive RPO ordering; unreachable entries are removed or
replaced with the same structural placeholder as Go. Real successors and structural
fallthroughs are separate adapter edges. All RPO, predecessor/evaluation ordering,
phi construction and elimination come from the imported SSA module. There is no
local copy of any of those algorithms.

The hir-v1 dump now includes block kinds, constructed phis/operands, all admitted
terminal payload fields and **orphan instruction** records in original table order.
A structural fallthrough with no incoming real edges becomes an empty Unreachable
placeholder; its removed instructions remain in the table with their original
unassigned order and IDs. The adapter emits every instruction exactly once, either
under its block or as an orphan. Prior certified straight-line dumps remain unchanged.
Every observed Go dump is printed twice and checked for byte stability.

Eight new mutants are caught on Node and native: swap if successors, take the wrong
ternary arm, swap logical short-circuit arms, exit instead of taking the while back
edge, send continue to the break target, turn throw into return, replace return by
Unreachable, and drop post-abrupt instructions. Together with the prior sixteen,
**24 lowering mutants** compile and run successfully before their dumps disagree.
The independent symbol-identity mutant is also retained. Six new probes exercise
control flow and orphan retention; the probes themselves do not raise corpus coverage.

Remaining construction work includes calls/new/spreads, objects/arrays/templates/
casts, destructuring and typed parameters, closures/context bindings, do/for/iterator
loops, switch/labels, exception/finally and optional chains. JSX, live ForFunction
integration/cache, full visitors/cloning/postdominators and static-components remain
unfinished. No rule is claimed green at this seam, and no blocking Adamic language
gap was encountered.


## Calls and constructors: verified local checkpoint

Adds CallExpression, MethodCall, NewExpression and ordered argument records, including
spread flags and direct optional calls. A named method emits a string primitive at
the property-name span and retains the receiver; computed methods evaluate the key.
The callee/receiver/key precede left-to-right arguments. Parenthesized methods follow
Go's plain-call path. Generic call/new type arguments are erased. All operand visits
run through the existing imported SSA construction, with the original evaluation order.

React origins currently resolve direct named/default/namespace imports and same-file
const aliases using resident compiler symbol identities and exact declaration spans.
Import renames, namespace/default member access, string computed members and cyclic
const aliases are handled. General cross-file barrel re-exports and aliased-symbol
chains still need generic compiler export-shape facts and the complete export-origin
resolver; the current corpus matches do not certify that missing path.

The owner changed the push policy during this checkpoint: push only after a whole
unit's own fixtures, tests and mutants pass locally. This checkpoint stays local;
unit 2 is unfinished. Next: closures/context bindings, the remaining expression and
control-flow forms, JSX, ForFunction integration and static-components. The original
Flow exclusions and 45 skipped-test unit assignments remain unchanged.


Final local comparison: **222/1,465 corpus functions**, an increase of **110**
over 112; **24/24 probes**, hence 246/1,489 with probes. Active non-Flow coverage
is 222/1,442. Direct optional calls add a witness but do not independently move the
corpus count. All **29 lowering mutants** compile, run and disagree with the oracle
on both Node and native. The five new mutations remove a spread flag, substitute the
method receiver, substitute the constructor callee, clear React origins, and clear
optional-call status. The retained 12-case regression and its return-store mutant
pass. All 22 resident symbol selectors and the identity-collapse mutant also pass.
No unit gate or push was initiated. This is a local partial checkpoint of unit 2.


## Closure WIP details

Nested functions retain separate tables; first-encounter captures map child Context
places to parent FunctionExpression.Captures. Grandparent captures are re-captured
through intermediate builders. Shadowed names use checker identities. Context analysis
runs before lowering; outer reassignment shared with an inner reference registers
ContextDeclarations and emits StoreContext. Captured writes also emit StoreContext.
Function declarations use Go's HoistedFunction kind (6) and declaration-node ranges.
The SSA adapter visits context reads/writes/captures, uses the imported contextual
store predicates, and constructs every nested function recursively. No graph analysis
was copied from SSA or mutation_aliasing.

The census now records enclosing-root byte spans and a nested function-index path.
A context-bearing child is admitted only when its enclosing root is also supported;
the port constructs the whole root and selects that child's graph afterwards. A
context-free child can still be lowered independently, including when its parent is
outside the current grammar. This distinguishes standalone and captured lowerings
without taking any identifier/capture answers from the Go dump. Deduplicated graph
keys and the original 1,465-function denominator remain unchanged.

Six new probe sources exercise repeated capture reads, grandparent capture forwarding,
outer/context writes, inner captured writes, hoisted function declarations and shadowed
parameters. Their recursively observed graphs raise probes from 24 to 37. New mutants
change a capture read to a local read, substitute the capture pairing, change an outer
context store to a local store, omit context declaration registration, and change a
hoisted declaration to an ordinary let. All five execute and disagree on Node; none
is counted as a caught native mutant. Existing 29 native witnesses remain certified
at the previous checkpoint. Flow exclusions and private-corpus skip assignments stay
unchanged.

Next work remains closures once the native refusal is resolved, remaining flow and
expression variants, JSX, full export provenance, ForFunction/cache/visitors/cloning,
and static-components. This WIP is local and no unit gate has been triggered. The
oldest unpushed checkpoint is about ten minutes old, below the 90-minute backup limit.


## Arena representation checkpoint

The prior closure-WIP section records the rejected pointer representation and its
Node-only certificate historically; it is superseded by the native certificate above.
The original diagnostic/reproducer remains in EVIDENCE.md for @system_adamic as the
reason for the representation change. The result is an owned graph with numeric edges,
not Weak references or a readonly-array exception to the cycle rule. All SSA algorithms
remain imported by reference; the adapter alone changes storage and typed boundaries.
The next construction work is remaining control flow and JSX, followed by complete
export provenance, ForFunction/cache/full visitors/cloning and static-components.
