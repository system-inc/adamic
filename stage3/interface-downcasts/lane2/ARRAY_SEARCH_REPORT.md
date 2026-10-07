Built: checked indexOf/includes/lastIndexOf consume only selected positions in both backends, using shared array-view read contracts.
Branch: codex/views-arrays-callables-parser, after bf544d5c; published SHA follows in the push report.
Validation: nine Node controls match sanitized/release native and emitted JavaScript; the uncached focused compiler/oracle gate and package vet pass.
Mutants: contract dispatch, literal, lazy stopping, SameValueZero, bounds, identity, JavaScript bounds, and native/JavaScript source-write guards all fail semantically.
Limits: array-like interface representation, own fields, optional receiver chains, tuples, remaining mutations and lazy admission are incomplete; no full tsc-site unlocking is claimed.

## Implementation

ArraySearch carries ArrayViewRead metadata, resolved by the existing readiness
finalizer. viewArraySearch is the named lower hook in the shared object/library
array dispatch. Native libraryArraySearch delegates to emitViewArraySearch when
that metadata is enabled; JavaScript uses its corresponding named emitter.
No callable files or certification hooks were changed.

Searches evaluate and hold receiver/value before the from argument; source alias
replacement in the final argument must not change the selected receiver. Native
retains reference operands across those calls. Each comparison uses the existing
selected-element decoder and kind/literal/object checks. No field read scans the
array, and a successful early search never touches later malformed elements.

indexOf and lastIndexOf skip holes and use strict scalar or reference equality.
includes visits missing positions as undefined and uses SameValueZero, including
NaN. Objects compare by identity. Negative, fractional, infinite and NaN starts
follow Node; omitted lastIndexOf starts at the end, while explicit undefined starts
at zero. Sparse search preserves source storage and ownership without materializing
the array. Nullish element contracts still require their shared admission lane.

## Original source contracts at writes

The pointer-storage certificate cannot prove source object fields, nominal class
identity, nested arrays, map parameters or closure signatures. Reference writes
(other than strings) now stop loudly with an uncertified source element contract
rather than overwrite an allocation which another alias sees through a narrower
contract. The new witness writes a value-only record into a view of records whose
source alias also requires secret. Both backends stop at the write with exit 70;
skipping the guard is caught separately in native and JavaScript. Allocator-to-
source-contract certification is still needed before these writes can be admitted.
Scalar numeric/boolean/string paths remain covered. This is an explicit limit on
mutation coverage, not a claim that pointer storage certifies an object shape.

## Census remainder

| Family | Original demand | Remaining mechanism or dependency |
| --- | ---: | --- |
| Array contracts | 334 pairs / 3,189 reads | Partial; complex contracts and per-read proof not counted as unlocked |
| Element / consumer reads | 251 / 1,602 | Partial; searches added, other missing consumers remain |
| Own array fields | 30 / 72 | 30 / 72 |
| Tuples | 6 / 9 | 6 / 9 |
| Refused intrinsic names | 35 / 128 | 15 / 27 under sort, splice, unshift, concat and shift |

The search family demand is 15 pairs / 52 reads: indexOf 11 / 34, includes 3 / 16,
lastIndexOf 1 / 2. Seven pairs / 27 reads have ordinary array receivers whose
scalar/reference search mechanism is implemented. Eight pairs / 25 reads have
NodeArray receivers (one also optional) and still need array-like interface
admission/representation. The join checkpoint retains two optional receiver-chain
reads. Counting unresolved receiver dependencies as well as the still unsupported
intrinsic names gives 25 pairs / 54 reads within the original 35 / 128 demand.
These are mechanism/dependency counts, not observed full-source lowering counts;
object element contracts can still need other lanes. No transitive graph census
is subtracted to invent a successful cast count.

The 27 reads under remaining unsupported intrinsic names are sort 4 pairs / 11,
splice 4 / 7, unshift 4 / 5, concat 2 / 3, shift 1 / 1. Own fields have the larger
72-read gap. Their NodeArray interfaces currently reach inherited callable
contracts eagerly and need the shared lazy admission contract before this lane
can integrate their native property representation safely.

## Evidence and coordination

array-search-node-controls.json records all nine original source Node outputs.
The positive controls cover the full start-index matrix, NaN/signed zero, boolean
values, object identity, sparse arrays, short-circuiting before a malformed later
element, and replacing the captured receiver alias during argument evaluation.
Negative controls pin number-versus-string and literal failures with exit 70 on
both backends. Each uses both sanitized native and release native.

Restored, uncached validation (array-search-logs/focused.log):

    ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/oracle -run 'Test.*View|TestPhantom|TestArrayHolesAliasRefusals|TestArrayHolesNullableSearchRefusal|Test.*ArrayJoin|Test.*ArraySearch|Test.*ArraySlice|Test.*ArrayPop|Test.*ArrayPush|Test.*Template' -count=1

Observed after the source-write guard and mutant restoration: lower 6.748s; native
0.665s; JavaScript 0.891s; IR 0.028s with no matching tests; oracle 44.457s. All pass. go vet on those four compiler packages passes.
The full repository gate was not run; the earlier unrelated predicate failure in
TestArrayHolesMilestone is unchanged.

run-array-search-mutants.py records each semantic failure in array-search-logs,
restores files in finally, and rejects clang/build/JavaScript syntax failures as
proof. The seven search mutants and two source-write guard mutants all fail their pinned oracles. The previous eleven
native-array mutants remain recorded under native-array-logs.

The lead now assigns all inter-lane merges/conflicts to codex/views-integration.
This lane stops merging Lane 1 directly. The integrator has now published lazy
admission at f1c91970, without the bf544d5c native-array checkpoint. Publish this
lane's search tip, then attempt that integration SHA; leave any conflicts to the
integrator and report the exact boundary. No cast conflict is resolved locally.
