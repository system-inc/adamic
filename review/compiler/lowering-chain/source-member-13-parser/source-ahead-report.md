Historical report for 3c6f863d. Placeholder pending marks below are superseded by
[parser-next Unit A](../parser-next/UNIT-A.md), which records the delivered
c8f858b9 recheck on rehearsal 678d94f9. Program-region lifetime remains pending.

Built: construction presence, V8 first-write key order, inline NodeArray extras, honest undefined observations and checked scalar uses toward step 24.
Commits: rebased contracts onto rehearsal 3d0620c6; preserved published b95d352b through ancestry commit 8ee2057f; delivery SHA is reported after push.
Commands and outputs: focused source oracles, native ASan/UBSan/leaks, JavaScript, counts recording, a-check and go vet pass; logs are in ../evidence.
Mutants: three actual runtime order/JSON mutants disagree with Node; removed declared-T and pending-null guards fail their pins; four completion and four slot mutants fail; payload mutant differs in both backends.
Uncovered: rehearsal results pending codex/placeholder-nonnull; observable null! remains NotYet; Program-region lifetime pending step 06 #7g4qv2b; no whole-parser or check-erasure certificate.

No push to main or any area branch. No worker branch was merged. The authorized
rehearsal is the rebase base. An ancestry-only ours merge preserves the old published
feature tip so the final push can be nonforce; its tree is the rebased tree.

Toolchain: cloud/setup.sh completed in 44.597s, build 44.292s; nproc 5, quota 4.
GOPROXY uses https://proxy.golang.org|direct. Go 1.27.1, Node 24.19.0, clang 20.
Setup evidence remains evidence/setup.log.txt. Global counts initially failed on
missing @types/node 25.3.3. npm ci --ignore-scripts --prefix stage3/api installed
its pinned dependencies; global counts then passed. No dependency files changed.

Commands, with output redirected to /tmp logs and copied into evidence:

```
go test ./internal/lower -run '^TestParser(Factory|NodeArray)' -count=1 -v
go test ./internal/oracle -run '^TestParser(Construction|AheadFactory|AheadNodeArray)' -count=1 -v
go test ./internal/oracle -run '^TestParserConstructionSource$/node-array-unset$' -count=1 -v
go test ./internal/oracle -run '^TestParserConstructionNullPlaceholderPending$' -count=1 -v
go test ./internal/native -run '^TestParserConstruction(Backends|RuntimeMutants)$' -count=1 -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(casts|library_for_in)\.a$' -count=1 -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/class_features_private\.a$' -count=1 -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts
go vet ./internal/ir ./internal/lower ./internal/javascript ./internal/native ./internal/oracle
python3 stage3/parser-ahead/check-local.py
python3 stage3/parser-ahead/mutants.py
python3 stage3/parser-ahead/node-array/mutants.py
node --disable-warning=ExperimentalWarning stage3/parser-ahead/rulings/observe.cjs
```

No whole package test or full gate was run. The source-oracle command contains
six successful rulings reductions, two original factory reductions, the complete
control and its payload mutant, all four original metadata fields and one retained
Node-only boundary. The separate null-placeholder test records pending, not native
acceptance. The shared filtered oracle also selected proven_upcasts.a because Go's
subtest regular expressions are unanchored; it passed. An initial selector selected
no fixtures and is not credited. gofmt and git diff --check pass.

Mutants and catchers:

| Mutation | Catcher | Evidence |
| --- | --- | --- |
| Publish absent keys early | Native own-key-order stdout disagrees with Node | runtime-mutants.log.txt |
| Use declaration order for strings | Native own-key-order and JSON stdout disagree with Node | runtime-mutants.log.txt |
| Include array extra in JSON | Native [1,2,-1] differs from Node [1,2] | runtime-mutants.log.txt |
| Skip declared-T check | TestParserConstructionUnsetUse diagnostic differs | unset-use-mutant.log.txt |
| Drop pending-null guard | TestParserConstructionNullPlaceholderPending gets no boundary | null-guard-mutant.log.txt |
| Branch union instead of intersection | TestParserFactoryCompletion one-branch proof assertion | completion-mutants.json |
| Omit escape missing fields | TestParserFactoryCompletion before-escape assertion | completion-mutants.json |
| Erase earlier read obligation | TestParserFactoryReadBeforeCompletion assertion | completion-mutants.json |
| Trust asserted scalar payload | TestParserFactoryCompletion wrong-payload assertion | completion-mutants.json |
| Wrong pos slot | TestParserNodeArrayLayout slot assertion | node-array-mutants.json |
| Wrong end slot | TestParserNodeArrayLayout slot assertion | node-array-mutants.json |
| Wrong hasTrailingComma slot | TestParserNodeArrayLayout slot assertion | node-array-mutants.json |
| Wrong transformFlags slot | TestParserNodeArrayLayout slot assertion | node-array-mutants.json |
| Replace ready with wrong | Original Node stdout catches native and JavaScript payloads | source-oracles.log.txt |

The runtime mutants compile and execute normally, exit 0 with empty stderr, and
pass ASan/UBSan/leak checks; mismatch is their catcher. The proof and layout mutants
are compiler assertions, not native emission credit. The three source-prediction
mutants are recorded separately in observations.json and are not runtime credit.
The first declared-T mutant attempt lacked clang in PATH and is not credited; the
recorded rerun sourced the toolchain, compiled and executed both backends.

Counts: columns are allocation/free/retain/release/high-water/region counts.
Only owned parser rows change; all pre-existing non-parser values are unchanged.
The three old parser rows move into global adapter order; already-complete's values
stay unchanged. complete and escaped previously stopped before assignment; they
now execute, free their allocation and retain/release initialized or observed values.
New rows measure the corresponding executed source reductions: factory-use stops
at its pinned check (one allocation remains at process exit), key rows allocate key
strings and result arrays, JSON allocates serialization output, and length measures
only element operations. null-placeholder-pending has no counts row because it
cannot yet produce a backend program.

Changed from | stage3/parser-ahead/factories/complete.a | 1 | 0 | 2 | 1 | 1 | 0 | to | stage3/parser-ahead/factories/complete.a | 1 | 1 | 5 | 4 | 1 | 0 |
Changed from | stage3/parser-ahead/factories/escaped.a | 1 | 0 | 2 | 1 | 1 | 0 | to | stage3/parser-ahead/factories/escaped.a | 1 | 1 | 3 | 4 | 1 | 0 |
New: | stage3/parser-ahead/rulings/factory-unset.a | 3 | 3 | 13 | 20 | 3 | 0 |
New: | stage3/parser-ahead/rulings/factory-use.a | 1 | 0 | 2 | 3 | 1 | 0 |
New: | stage3/parser-ahead/rulings/own-key-order.a | 28 | 28 | 17 | 37 | 10 | 0 |
New: | stage3/parser-ahead/rulings/node-array-keys.a | 15 | 15 | 12 | 16 | 9 | 0 |
New: | stage3/parser-ahead/rulings/node-array-json.a | 2 | 2 | 5 | 9 | 2 | 0 |
New: | stage3/parser-ahead/rulings/node-array-length.a | 4 | 4 | 4 | 9 | 2 | 0 |
New: | stage3/parser-ahead/rulings/node-array-unset.a | 4 | 4 | 18 | 29 | 4 | 0 |
New: | stage3/parser-ahead/node-array/fields.a | 12 | 12 | 10 | 23 | 12 | 0 |

Design questions and remaining boundaries:

- The scout generic factory `return createBaseNode(kind) as T` is still checked and
  does not gain an unchecked generic type certificate. Only fresh scalar literals
  and reviewed direct single-literal-return functions reserve the fixed layout.
  Unknown helpers and alias effects retain checks or existing NotYet boundaries.
- A factory field is undefined before its first write; null! must remain null.
  Node prints `null` then `undefined` in null-placeholder-pending.a. Proposal: consume
  observable-null transport from codex/placeholder-nonnull when it is published on
  the authorized integration ref. Current field-name collision handling is
  conservative NotYet, not a claim that null became undefined.
- `Object.prototype.hasOwnProperty.call(node, 'beta')` reaches the train's existing
  delayed-library-result NotYet. Direct `node.hasOwnProperty('beta')` is held to Node
  here. Proposal: the callable-result lane should handle the original .call shape.
- NodeArray has four required fields at the TypeScript pin, including enum flags.
  Optional string extensions use existing optional references. Optional numeric
  extensions whose layout needs MaybeNumber remain NotYet; no key is fabricated.
- Program-region lifetime, recursive tree parent cycles and allocations retained
  across speculation rewind remain pending step 06 #7g4qv2b. Heap reductions and
  sanitizer passes do not certify that lifetime contract.
- AnalyzeFactory remains a bounded proof and production keeps scalar checks;
  no global field-name check erasure is certified. Speculation result specialization
  and whole-parser recursive-stack measurements are contractual, not built here.

The amended contract follows in full for compiler posting:

# Step 24 parser compiler contract

This contract serves #ktz9fek under #6z35tzs. The authorized build-ahead base is
compiler/rehearsal-placeholders 3d0620c6, comprising train 9f16421c and
codex/placeholder-nonnull 19193932. Scout evidence is codex/step24-scout 7e6ed35b.
TypeScript source is 050880ce59e30b356b686bd3144efe24f875ebc8.
Rehearsal-dependent results remain pending codex/placeholder-nonnull integration.
Program-region lifetime remains pending step 06 #7g4qv2b.

## 1. Factory completion

Lowering asks before trusting a factory result or erasing a checked field read:

```go
AnalyzeFactory(factory *ast.Node, target *checker.Type, checker *checker.Checker) FactoryCompletion
FactoryReadPlan(summary FactoryCompletion, field string) FieldReadPlan
```

Integration points: functions.go asks at a concrete factory body/return; cast.go
asks before admitting a base-result cast; object.go asks while forming a Property
read. class.go writes publish readiness after their RHS. The bounded analyzer supplies source proofs in its focused tests. Production
currently keeps declared-T checks rather than erasing them from those summaries;
no production check-erasure optimization has acceptance credit.

The summary identifies allocation sites and aliases by bound symbols, each
required field's declared type, typed writes definitely executed at every escape,
and reads before those writes. Each escape retains its source location and the
missing field set. Return, global/container store, a captured alias exposed to a
closure, and an opaque call receiving the object are escapes. Unknown control
flow, computed writes, accessor effects or alias provenance cannot certify a field.
A cast is never evidence of a write's runtime type. Required means non-optional,
including inherited required fields. Optional absence is not incompleteness.

Transfer is forward must analysis: literal initializer fields start written only
when their actual initializer types satisfy the declared contract; plain field
assignments add a field only after their RHS and all its effects finish. Joins
intersect facts over reachable predecessors. Loops add no unconditional writes
without a fixed point; early returns and throws retain separate exits. A helper
summary may transfer writes only when every implementation and alias effect is
known. Recursive or generic summaries are checked at the concrete specialization;
an unsolved summary yields checks, never a guessed certificate.

A completed return can be trusted only if every required field was written before
every escape on that returned allocation's paths. Later writes do not repair that
certificate. They do update runtime initialization state. A post-escape read after
a real write can therefore pass its runtime check. A field not written yet holds
an honest unset whose observable value is undefined. Observation is allowed in
presence tests, optional slots and flow-bounded save locals. Every consumption as
the declared T is checked first. Reuse codex/placeholder-nonnull machinery through
compiler integration; do not implement a second unset representation. The generic
base-to-Identifier result stays checked until text is written. Existing ir.Property
readiness/view metadata must distinguish observation from typed use; the future per-allocation certificate
must not erase checks globally by field name. Allocation reserves its fixed target
layout, with absent/uninitialized slots distinguished from initialized undefined.
Every slot carries presence during construction. Reserve storage at allocation
while keeping the property absent. The actual write publishes its value, presence
and initialization; it does not publish an invented initial property. No source
builder rewrite or invented zero/empty string is permitted. An unwritten factory field observes as undefined without claiming T is initialized.
A literal null! placeholder must observe null, and undefined! must observe undefined.
The rehearsal has not implemented observable null placeholders; field-name collisions
with null! return NotYet rather than treating null as undefined. This part is pending
codex/placeholder-nonnull, pinned by null-placeholder-pending.a.

FieldReadPlan distinguishes observation/save from declared-T consumption, with
field, factory, expected contract and use location. An unset observation returns
undefined. A typed use evaluates the value once and checks initialization/runtime
type before consuming it. Failure exits 70 naming the field, factory and use; the
placeholder mechanism owns the shared check and its final pinned wording. A loud
stop on every read is forbidden. Presence and bounded transport never manufacture
a certificate that the declared T is initialized.
An unreifiable field representation is NotYet: `stage 0 can't lower factory field
<field> of type <type> yet`. Explicit any and forged scalar payloads remain Refused:
`use unknown and narrow it, or declare the field's proven type`. Missing proof alone
is not a permanent refusal. The first bounded analyzer may return Unknown and all
fields Checked; this is a planning result, not permission to emit unsupported reads.

## 2. NodeArray

The source declares NodeArray<T> and MutableNodeArray<T>, inheriting TextRange:
required pos:number, end:number, hasTrailingComma:boolean, transformFlags:TransformFlags.
These are the four declared extra fields at the source pin. NodeArray itself has
no optional extra field there. A future declaration change must update the schema,
not silently lose a field. ReadonlyArray's library methods are not metadata slots.

```go
NodeArrayLayout(element ir.Type, extras []FieldContract) (ArrayLayout, error)
LowerNodeArrayRead(receiver ir.Expression, layout ArrayLayout, field string) (ir.Expression, error)
LowerNodeArrayWrite(receiver ir.Expression, layout ArrayLayout, field string, value ir.Expression) ([]ir.Statement, error)
```

Layout is one array allocation: ordinary count/header/length/capacity/element-buffer
ownership followed by fixed typed metadata slots and presence/initialization bits.
It preserves Array identity, numeric indexing, Array.isArray, iteration and length;
there is no containing object, wrapper, side map or arbitrary expando store. Field
order is pos, end, hasTrailingComma, transformFlags. Positions are number values
with UTF-16 semantics; flags retain their declared enum contract, not a boolean.
Required slots start absent and unset until actual writes. Physical slot order
does not determine own-key order. Integer-like keys enumerate ascending; string
keys enumerate in first-write order. A rewrite keeps its original order.
Object.keys, for...in, in, hasOwnProperty and object JSON.stringify share this
presence/order state, with their ordinary JavaScript operation semantics.
Optional slots use step 17
presence plus value representation, distinguishing absent from present undefined.

pos/end/transformFlags lower to numeric metadata loads; hasTrailingComma to a
boolean metadata load. Unset observations and declared-T consumption follow
the shared placeholder rule above, checking the use rather than every read. Metadata writes use declared slots and value contracts, preserving
JS write order. Unknown metadata keys are NotYet: `stage 0 can't lower NodeArray
metadata field <field> yet`; arbitrary properties remain Refused: `declare a fixed
array field instead of adding an expando`. Array algorithms returning new ordinary
arrays do not copy metadata unless source explicitly writes it. slice is not a
NodeArray factory. Alias writes must be visible through the same allocation.
Array Object.keys lists present indices ascending, followed by extra string fields
in first-write order. Array JSON.stringify ignores extra fields. length counts
only elements; metadata never changes array identity, element count or serialization.

object.go handles named reads and array construction; class.go handles metadata
writes. Built IR uses ArrayLiteral{Element, Elements, Metadata []Field} with absent
fixed extras and unchanged ir.Array identity. ObjectLiteral fields carry Absent and
Uninitialized. Native construction.c allocates the metadata inline in the array's
allocation. A metadata pointer names interior storage; it is not another allocation
or wrapper. The ordinary element buffer retains existing array ownership.

Lowering's nodeArrayProperty and factoryFieldRead use existing DynamicProperty and
ir.Union for observations and unannotated const saves. unsetSpecialization evaluates
that observation once, checks its primitive and finite literal/enum contract, then
narrows. Optional number/boolean destinations use existing MaybeOf; optional string
uses its existing reference representation. Required metadata writes and optional
string extras publish presence at their real write. Optional numeric extras that
need an unsupported layout remain NotYet, rather than inventing a value.

Fresh scalar object assertions and direct, single-literal-return base factories
reserve the reviewed result fields at allocation. Opaque factories, unknown alias
layouts and unrepresentable fields retain existing boundaries. The original scout's
complete and escaped reductions now match source Node in both backends.

The source fixtures read all four declared extras and exercise absent observations,
optional slots, saved locals, write order, JSON, length and array identity. Focused
source oracles execute real native and JavaScript code. Independent hand-built IR
fixtures additionally isolate the three runtime mutants. These are not stub passes;
rehearsal integration and Program-region lifetime still have no acceptance credit.

## 3. Speculation and checked specialization

```go
LowerSpeculativeResult(value ir.Expression, actual, wanted *checker.Type) (ir.Expression, error)
```

cast.go and concrete generic result lowering ask for this plan; control.go uses
the existing ToBoolean lowering for success/rewind branches. Planned IR is
CheckedSpecialization{Value, Contract, Message}, preserving the wanted result
representation with a runtime certificate. No unresolved checker type reaches C.

The proven direction is an ordinary value conversion. Otherwise emit a runtime
contract check at the consuming specialization, evaluate the result once and keep
its ownership. Failure is exit 70: `adamic: panic: speculative result failed:
<expression> expected <type>, found <runtime type>`. Unsupported reification is
NotYet, naming the result type. Never erase a cast to an unresolved type parameter.

Keep tsc's closures, saved token, diagnostic length, pending error and context bits.
lookAhead always rewinds; failed tryParse uses ECMAScript ToBoolean, including false,
0, -0, NaN, empty string, null and undefined. Objects and arrays are truthy.
Successful tryParse commits. Reparse preserves diagnostics as tsc writes it.
Do not rewind allocations or cached bits that tsc retains. A speculative result can
escape: do not end its storage region just because scanner state rewinds.

## 4. Program lifetime plug

The parser rulings #ktz9fek are blocked on step 06 #7g4qv2b for Program
lifetime. This part remains pending. Parser tree and NodeArray allocation sites
request the Program region from step 06
through the runtime owner's region allocation interface. Metadata-owned references
participate in that region's destruction; parent cycles use the approved region
ownership contract. The parser compiler does not implement a collector or invent
statement-region lifetime for returned trees. Until Program lifetime and escape
integration are present, tree-lifetime tests are pending even if heap reductions run.

## 5. Guarded recursion

Use ordinary recursive C, with the existing runtime stack guard at every emitted
function entry, including callbacks and scanner/factory calls. Failure is exit 70:
`adamic: panic: RangeError: Maximum call stack size exceeded`. JavaScript keeps its
own guarded backend behavior. Guard unwind and exception cleanup must preserve
region and reference ownership. Never set a frame-count limit of 859: that is the
scout's observed maximum active parser functions on its bounded corpus. Native
frame bytes, sanitizer overhead and whole-parser stack use require measurement.

## Validation contracts and design questions

Completed, conditional, alias-write, escape-before-write, wrong-type-write and
unknown-call witnesses are held to source Node. An incomplete read that Node returns
undefined for must also observe undefined in Adamic. Declared-T use must check
first and stop loudly if still unset, naming field, factory and use. Independent mutants
remove completion's branch intersection, escape recording and type validation;
remove a field-read check; change each NodeArray slot mapping; and remove the stack
guard. Compiler assertions or behavior differences count; build failures do not.

Node prints undefined from the scout's generic base-to-Identifier return. Proposal:
return a checked result until text is written, retaining the original source cast
and permitting honest unset observations before declared-T consumption.
Node exposes absent versus present undefined through own keys. Proposal: retain
presence bits even for required slots during construction; do not synthesize keys.
Node slice drops NodeArray extras. Proposal: preserve ordinary Array return semantics.
Node lookahead may return an allocated object after rewind. Proposal: its lifetime
belongs to Program, not a rewind arena. These proposals do not certify unexecuted
branches, key-reflection identity or the full parser.

The amended own-key/array contracts require three executed backend mutants:
publish a key before its first write; enumerate strings by declaration order;
serialize array extra fields. All three now compile and execute independently against actual native runtime
copies. Each exits normally with sanitizer/leak checks and disagrees with source Node
stdout. Source prediction mutants remain separate evidence and are not counted as
native mutant passes.
