# Step 24 parser compiler contract

This contract serves #ktz9fek under #6z35tzs. The authorized build-ahead base is
compiler/rehearsal-placeholders 678d94f9, including delivered
codex/placeholder-nonnull c8f858b9. Scout evidence is codex/step24-scout 7e6ed35b.
TypeScript source is 050880ce59e30b356b686bd3144efe24f875ebc8.
The construction source fixtures have been rechecked with the delivered placeholder machinery. Their placeholder pending marks are cleared.
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
The refreshed rehearsal transports both values faithfully. The historically named
null-placeholder-pending.a now checks null! as null and an unwritten factory field
as undefined, in Node and both real backends; its pending mark is cleared.

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
placeholder-dependent fixture marks are cleared on the delivered machinery. Program-region lifetime still has no acceptance credit.

## 3. Speculation and checked specialization

```go
LowerSpeculativeResult(value ir.Expression, actual, wanted *checker.Type) (ir.Expression, error)
```

cast.go asks checkedSpeculativeResult for a direct declared lookAhead, tryParse or
speculationHelper call whose checker result is unknown. Its callable parameter
is resolved before admitting this specialization. Proven generic results use the
ordinary lowering. The implemented IR reuses Union, checked Narrow and Property
views, with a generated helper taking the completed result once. Required fixed
scalar object fields and finite scalar literal contracts are checked in the helper.
No unresolved checker type reaches C. Arbitrary nested, callable, optional-field
and class result contracts remain NotYet. A contextual adaptation other than a
direct unknown-result assertion is not yet implemented.

The proven direction is an ordinary value conversion. Otherwise emit a runtime
contract check at the consuming specialization, evaluate the result once and keep
its ownership. Failure is exit 70: `adamic: panic: speculative result failed:
callback <expression> at <call site>; expected <type>`. Unsupported reification is
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
