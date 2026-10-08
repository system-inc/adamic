# Step 24 parser compiler contract

This contract serves #ktz9fek under #6z35tzs. Base is train slice 1,
9f16421c7910607cd06c006b719ff126bd15a7f7. Two live branches have number 1;
the newer commit is selected. Scout evidence is codex/step24-scout 7e6ed35b.
TypeScript source is 050880ce59e30b356b686bd3144efe24f875ebc8.
Signatures below are contracts for integration, not claims that instructions
already exist. Pending means no native acceptance credit.

## 1. Factory completion

Lowering asks before trusting a factory result or erasing a checked field read:

```go
AnalyzeFactory(factory *ast.Node, target *checker.Type, checker *checker.Checker) FactoryCompletion
FactoryReadPlan(summary FactoryCompletion, field string) FieldReadPlan
```

Integration points: functions.go asks at a concrete factory body/return; cast.go
asks before admitting a base-result cast; object.go asks while forming a Property
read. class.go writes publish readiness after their RHS. The first analyzer is
not wired into these paths yet: no production check erasure has acceptance credit.

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
a real write can therefore pass its runtime check. Reads before completion use
existing ir.Property readiness/view metadata; the future per-allocation certificate
must not erase checks globally by field name. Allocation reserves its fixed target
layout, with absent/uninitialized slots distinguished from initialized undefined.
Writes publish the value before marking the slot initialized. No source builder
rewrite, invented zero, empty string or undefined initialization is permitted.

FieldReadPlan has Proven or Checked, field, expected contract and read location.
Checked reads evaluate the receiver once, check presence, initialization and runtime
type, then load. Both backends stop at exit 70 with:

```
adamic: panic: field read failed: <expression> expected <type>, found <missing|uninitialized|runtime type>
```

The actual checked-view lane's pinned wording takes precedence at integration.
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
Required slots start uninitialized until actual writes. Optional slots use step 17
presence plus value representation, distinguishing absent from present undefined.

pos/end/transformFlags lower to numeric metadata loads; hasTrailingComma to a
boolean metadata load. Unproven readiness/type uses the same field-read check and
exit 70 text above. Metadata writes use declared slots and value contracts, preserving
JS write order. Unknown metadata keys are NotYet: `stage 0 can't lower NodeArray
metadata field <field> yet`; arbitrary properties remain Refused: `declare a fixed
array field instead of adding an expando`. Array algorithms returning new ordinary
arrays do not copy metadata unless source explicitly writes it. slice is not a
NodeArray factory. Alias writes must be visible through the same allocation.

object.go handles named reads and array construction; class.go handles metadata
writes; native/emit_arrays.go and the JavaScript emitter consume the same layout.
Planned IR nodes are NodeArrayLiteral{Element, Elements, Layout, Metadata},
NodeArrayRead{Array, Layout, Slot, Contract, Checked} and
NodeArrayWrite{Array, Layout, Slot, Value, Contract}. Their Type remains ir.Array;
field reads have the slot representation. Ownership visitors must walk all operands
and metadata references before admission. These nodes are contractual signatures,
not speculative additions to ir.go. The built ArrayLayout is slot planning only.

Both backends require metadata support before these instructions can be emitted.
A Node-only schema fixture is pending, not a backend pass. No ir.Type or instruction
is added merely to describe a future runtime ABI.

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

Parser tree and NodeArray allocation sites request the Program region from step 06
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
undefined for must loudly stop in Adamic, not print a default. Independent mutants
remove completion's branch intersection, escape recording and type validation;
remove a field-read check; change each NodeArray slot mapping; and remove the stack
guard. Compiler assertions or behavior differences count; build failures do not.

Node prints undefined from the scout's generic base-to-Identifier return. Proposal:
return a checked result until text is written, retaining the original source cast.
Node exposes absent versus present undefined through own keys. Proposal: retain
presence bits even for required slots during construction; do not synthesize keys.
Node slice drops NodeArray extras. Proposal: preserve ordinary Array return semantics.
Node lookahead may return an allocated object after rewind. Proposal: its lifetime
belongs to Program, not a rewind arena. These proposals do not certify unexecuted
branches, key-reflection identity or the full parser.
