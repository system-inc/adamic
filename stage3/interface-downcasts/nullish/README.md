# Nullish checked reads, October 8

Presence probe is pushed separately at a0cc0afd. See PRESENCE.md: post-write
observations match Node, while deleting through an alias exposes a native
Object.keys optimization bug. The original optional-number write is TS2412;
the explicit optional number-or-undefined control proves storage presence.
Integration d1937df8 was merged before this work, in fac326f8.

This batch implements T | null, T | undefined, and T | null | undefined for
number, boolean, string, structural object and array fields through checked
views. Null-containing unions use a boxed representation with an immortal null
identity; undefined remains NULL. Undefined-only unions retain existing Maybe
or reference representations. Checked reads validate presence, initialization,
producer storage, logical type and finite scalar literals before conversion.
Nested payload reads remain checked. Coalescing, truthiness, typeof and null and
undefined comparisons preserve the two identities. A narrowed union checks its
actual kind before unboxing. No asserted target creates a producer certificate.

The shared allocation flow still decides whether an unsupported read can see a
view. Nullable descriptors defer their unsupported present-family contract to
the read. Present unions requiring member selection stay named refusals until
the nullable adapter validates that selection. Nullable callable coverage was added in the next batch below; Map payloads
still require a collection certificate; nullable writes do not acquire new slot certificates.
Destructuring representation conversion and unsupported element representations
keep existing explicit refusals. Computed literal member accesses reuse property
access lowering. Legacy RegExp reference-null producers are preserved at boxed
Object.is observations, including direct calls and const bindings; their identity
regression prints true, false, false, true on Node and both backends.

## Evidence

`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewNullish|^TestCheckedViewLazy|^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(regexp|regexp_match|regexp_null_narrowed|library_object_is|unions|undefined_strings)\.a$' -v -count=1 -timeout 10m`

Twenty valid matrix fixtures match Node in sanitized native, release native and
JavaScript; leak checks pass. Forty source mutants change a present type, remove
a required property or substitute the excluded nullish alternative. Every one
exits 70 at node.value, naming expected and found. Two additional mutants change
an object's label or an array element: their later transitive string reads trap.
The helper receives the view through an interface parameter, rather than reading
only at the cast site. Node's output is recorded even for incorrect assertions;
Adamic intentionally stops when a checked contract fails.

Three temporary implementation mutants, restored after running, are documented
in implementation-mutants.json: collapsing null into undefined fails the Node
output comparison; admitting forbidden undefined fails the exit-70 requirement;
skipping the logical kind check fails the named field-read requirement. The
RegExp identity regression initially failed and now agrees with Node.

Full lower and JavaScript package run: JavaScript passes. Lower retains seven
known baseline failing test groups (overloaded shorthand, census predicate,
census overload, two phantom-array IR comparisons, nested-function gap wording,
phantom-array cycle). This is not a full green gate. Focused nullish/lazy and
selected legacy oracle checks pass (42.740s). Focused lower passes (4.505s),
native view/union/typeof passes (3.865s), JavaScript passes (0.667s), and vet
passes. The broader native filter also selected the known baseline
TestGraphLazyRegions failure (live-region accounting output); it is recorded,
not reported green.
Test command output is saved to log files, never piped.

## Exact remaining read table

This is an inventory, not the checker-option split delegated to TypeScript's
ledger worker. Reproduced with the adapted output of stage3/apply.sh from
234ab1aa5f728a5221fb6075c35b94f88a2c6437 on TypeScript 6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8:

`NODE_PATH=/tmp/lazy-stage3-cache/api/node_modules node stage3/interface-downcasts/lane4/read-demand.cjs /tmp/lazy-stage3-adapted /tmp/nullish-demand > /tmp/nullish-demand/run.log 2>&1`

read-sites.json.gz and read-pairs.json.gz are the exact nullish-member subset of
that script's output, including source locations, receiver ids, field names,
full declared types, read forms and overlapping families. read-summary.json
records 2,018 pairs and 9,101 reads: 8,712 property, 128 element and 261
destructuring reads. Family counts overlap and must not be added together.

All 2,018 production pairs and 9,101 production reads remain unverified by
successful whole-program lowering/reaching-view queries: the compiler's strict
checker policy has not yet been changed using the ledger. The inventory retains
Unknown fallback at every site. The fixture-proven five representation adapters
are not subtracted from this exact table or reported as compiled tsc sites.
Remaining child-family overlaps include 224 callable pairs/392 reads, 219
array-or-tuple pairs/1,195 reads, 221 dictionary pairs/868 reads, 131 tagged-object
union pairs/883 reads, 27 mixed-primitive pairs/124 reads and 60 object-plus-
primitive pairs/181 reads; the complete list is in read-summary.json.

The ledger worker owns the row-for-row own-tsconfig/Adamic-options/stage-0 split.
No diagnostic conversion or cast compilation counts are inferred here. Inserted
checks will be built from its only-under-Adamic-options column when supplied.

## Nullable callable and collection boundary follow-up

Nullable callable fields now pass all three alternatives and coalescing. The
present function is compared with GetNonNullableType of the target signature,
using the existing closed-world implementation/body proof. Null and undefined
never authorize a calling convention. The same helper first reads an ordinary
Target literal, then cast values. Four positive callable fixtures match Node,
sanitized and release native, JavaScript and leak checks. Eight source mutants
(wrong present kind, missing field, excluded nullish alternative) trap at
node.value in all three compiled runs. A ninth changes the actual implementation
return from string to number; it is refused at the helper read with the callable
family. Removing the implementation proof temporarily makes that refusal test
fail; the guard was restored and all tests rerun.

Nullable Map fields are admitted when unread (three Node/backend/leak matches).
Eight demanded Map cases, including five wrong-kind/nullish mutants, refuse
precisely at node.value with the collection family. Three valid Map cases are
also refused: this is an existing collection-certificate gap, not evidence of
nullable Map support. The native map has reference_values, not a complete value
kind/contract certificate; number and boolean storage cannot be distinguished
from that bit. Reading it using an asserted Map generic would be unsound. The
nullish wrapper preserves this refusal. The integrator can combine this batch
with a future producer-backed collection contract without replacing the flow
solver or trusting the target cast.

Follow-up command:
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewNullish$/^callable-|^TestCheckedViewNullishMutants$/^callable-|^TestCheckedViewNullishCallableSignatureMutant$|^TestCheckedViewNullishMap' -v -count=1`
passes in 7.210s after restoring the implementation mutant. Focused lower
view/callable tests pass in 2.544s. Compressed logs are in evidence/.

After this push, the exact production table still has 2,018 pairs/9,101 reads
awaiting successful lowering and reaching-view proof. None are subtracted based
on fixture coverage. Callable overlap remains 224 pairs/392 reads as an inventory,
not a remaining language-gap count; the wrapper now has fixture evidence.
Inserted option checks still await the ledger's only-under-Adamic-options rows.

## Finite literal follow-up

A nullable two-string-literal field matches Node for present, null and undefined.
Its mutant constructs not-allowed at runtime, preserving the string storage kind:
all three compiled runs exit 70 at node.value and name the allowed literal type.
This proves nullable logical-kind admission cannot skip finite value membership.
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewNullishLiterals$' -v -count=1`
passes in 0.876s. Its first attempted source mutation changed no bytes; that test
failed its expected-trap assertion. The corrected real-input mutant is committed.
Production pairs/reads remain 2,018/9,101 unverified; the exact inventory is unchanged.

## Nullable union selection, October 8

Integration e555d67e is merged in c5f19418, preserving both union and nullable
admission. Nullable reads select scalar members by kind and their individual
literal sets, tagged object members by the existing finite tag check, and a
single structural object member beside primitive alternatives by kind. Member
payload reads remain checked through the shared allocation flow. Ambiguous
array alternatives and untagged structural unions retain named read refusals.
No target type creates a producer certificate.

Nine valid fixtures (three shapes by three nullish forms) match Node, sanitized
and release native, JavaScript and leak checks. Nine wrong-member mutants and
three tagged-member bad-payload mutants stop with exit 70 at the read. The
number-both control catches nullable dispatch after the integration merge.
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewNullableSelection|^TestCheckedViewNullish$/^number-both$' -v -count=1`
passes in 12.188s. Production inventory remains 2,018 pairs/9,101 reads awaiting
lowering and reaching-view proof; fixture support is not subtracted. Ledger
inserted checks belong to stricter-checks worker 01a118d0 and are outside this unit.

## Map certificates, October 8

This supersedes the collection-certificate gap above. Map constructors now carry
complete scalar key/value contract identities, including finite literal sets.
Native headers and the JavaScript private WeakMap preserve producer evidence;
casts never manufacture it. Viewed reads compare that evidence with the target:
readonly schemas permit covariance, mutable schemas require invariance and the
same physical storage. Existing mutable-alias checks preserve certificates across
writes. Cloned maps receive the declared constructor schema. Unknown producers
stop at the read with exit 70 and name the field, expected type and source schema.
Nullable Map/primitive unions select the Map member and verify its certificate.

Number, boolean, string and finite scalar schemas have Node/backend evidence.
Structural, array, callable, branded and mixed-union entry schemas retain named
Map key/value read refusals. Multiple indistinguishable reference alternatives
remain refused. These are explicit boundaries, not fabricated certificates.

The combined nullable, selection, lazy-flow and Map oracle plus selected legacy
Map/Set regressions passed in 98.340s (sanitized/release native, JavaScript and
positive leak checks). Additional mixed-union controls passed in 2.554s; boolean
and phantom-brand controls passed in 0.600s. The brand fixture initially used an
unsupported unique-symbol brand declaration; it was corrected to a valid void
brand so the test demonstrates refusal at the Map field read. Scoped lower,
native and JavaScript packages passed in 4.109s/30.203s/0.932s; full IR and
JavaScript packages and go vet also passed. The full repository gate was not run.

Wrong key/value schemas, wrong present kinds, excluded nullish alternatives,
mutable widening and uncertified producers are caught by read checks or the
existing invariant-mutable diagnostic. Temporarily disabling the native Map
certificate guard made the oracle fail: a boolean-valued producer ran on through
a number-valued view. Restoring it passed the Map and selection suite in 17.953s.
Compressed command/output logs are in evidence/; the implementation mutation is
recorded in implementation-mutants.json.

After this push the exact production inventory remains **2,018 pairs/9,101
reads awaiting lowering and reaching-view proof**. No production site counts
are subtracted from fixture coverage. Ledger-driven option checks remain owned
by stricter-checks worker 01a118d0.

## Structural Map entries, October 8 follow-up

Complete structural key/value schemas now reuse the existing recursive slot
certificate and structural subtype relation. Readonly covariance and mutable
invariance retain their original rules; incomplete member descriptors cannot
stamp a Map constructor. Object keys keep identity semantics. Four new positives
match Node (structural values, finite covariance, object keys, and packed optional
number values). Five new source mutants stop at the demanded Map read or at a
transitive payload read: wrong structural value schema, mutable widening, wrong
object-key schema, wrong schema with an unread payload, and a viewed bad payload
passed through Map.get to a helper that also accepts ordinary entries.

The last mutant prints 1 then wrong on Node; both native modes and JavaScript
print 1 then stop at entry.count. Wrong entry schemas stop at node.value even
when the program never reads an entry. Temporarily replacing structural schema
comparison with physical object-kind equality made the unread-payload mutant run
on and print object (exit 0); the oracle failed. Restoring the relation passed
all Map controls in 15.309s. The optional-number entry control prints true, 9, 7
and passes in 0.479s. Scoped IR/lower/native/JavaScript checks passed in
0.009s/5.876s/47.990s/1.110s. Evidence is compressed under evidence/.

Arrays, callables, branded types, null-containing and mixed entry unions remain
outside the Map certificate adapter. Some nullable Map entry representations
also remain blocked by mapTypes/slotless, independently of view admission.
Production inventory is unchanged: 2,018 pairs/9,101 reads remain unverified by
whole-program lowering and reaching-view proof. Ledger checks are outside scope.

## Array Map entries and boundaries, October 8 follow-up

Map key/value certificates now include complete array contracts, retaining the
shared adapter's readonly bit. Readonly arrays permit element covariance;
writable arrays require element invariance and reject readonly sources. Every
comparison preserves the physical element representation. Recursive array schema
comparison terminates through active contract pairs. Producer descriptors with
unsupported descendants still cannot stamp a constructor.

Array values, readonly finite-element covariance, mutable same-schema writes
and a recursive empty array match Node in both native modes and JavaScript,
with leak checks. Wrong array schemas stop at node.value. A bad viewed array
payload passed through Map.get to a helper prints 1 then stops at values[0];
Node prints 1 then wrong. Mutable element widening is rejected by the Map read
certificate. An additional wrong-schema fixture leaves entries unread: only
reading node.value is enough to catch the incompatible schema. Temporarily
accepting all array pointer schemas makes it print object and exit 0; the oracle
fails. Restoring comparison passes the Map and nullable-selection suite in
30.336s, and the recursive-array control in 0.479s. Scoped IR/lower/native/JS
checks pass in 0.018s/4.106s/34.052s/0.867s.

Four demanded family controls (null-containing entry union, mixed scalar entry
union, callable entry, tuple entry) each refuse at node.value with the Map
key/value certificate family. Their four unread twins all compile and match
Node/backend/leak checks. This boundary suite passes in 9.691s. The remaining
entry gaps require storage and producer adapters: mapTypes explicitly refuses
Union and MaybeBoolean slot storage; callable contracts need implementation/body
proof from producers rather than asserted signatures; tuples need position/rest
contracts; nominal/phantom entries still lack authorized identity certificates.
These controls measure refusals, not implementation of those families.

After this push the exact production table remains 2,018 pairs/9,101 reads
awaiting whole-program lowering and reaching-view proof. No site is removed from
that table on fixture evidence. The complete repository gate remains unrun.

## NodeArray and undefined entry completion, October 8

Integration 267fd75b is merged in fc55836a. Array certificates now retain the
readonly bit from the inherited array base, not just the interface's own name,
and compare every target own-field schema with producer evidence. Missing own
fields are refused even when optional: an erased producer schema cannot prove
an unknown extra property's type. Own writable fields retain invariance.
NodeArray values and finite element covariance match Node. Wrong own schemas
and a readonly source viewed as a writable array stop at node.value. Their
unread-payload controls prove this is the Map certificate check, not a later
property or element guard.

Undefined-containing array and structural entries now use their pointer ABI and
complete optional descriptor. Both controls print true, undefined, 7 on Node and
all compiled runs. A Map containing undefined-valued arrays cannot certify a
nonoptional array-entry target. This was proved independently by disabling the
undefined source-schema guard: the wrong-schema control ran on (exit 0).

Phantom intersections are refused recursively inside Map entry arrays and object
fields, as well as directly. Finite primitive kind evidence cannot manufacture
a brand. Two demanded nested-brand controls refuse at node.value; their unread
twins compile and match Node. Disabling the recursive brand exclusion makes both
named-refusal assertions fail because lowering admits the read. This mutation
is an admission failure, not an observed runtime crash.

Four new implementation mutations were independently caught: skip array own-field
comparison, forget readonly producer restrictions, allow undefined schemas for
nonoptional entries, and omit nested phantom exclusions. The first three run on
and print object with exit 0; the fourth admits two forbidden reads. All guards
were restored before the final checks.

Final commands (outputs retained under evidence/):

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewMap|^TestCheckedViewNullableSelection|^TestCheckedViewNodeArray|^TestCheckedViewOptionalArray' -v -count=1 -timeout 10m
PASS, 34.060s
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'View|Map|Contract' -count=1 -timeout 10m
PASS, 0.010s / 4.044s / 31.999s / 0.809s
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript
PASS, no output
```

Remaining Map entry blockers: null-containing and mixed unions need the Map
storage ABI beyond slotless Union/MaybeBoolean; callable entries need producer
implementation/body evidence; tuples need per-position/rest contracts; nominal
and phantom entries need authorized identity witnesses. Repeated reference-kind
nullable union alternatives retain their named selector refusal. Production
inventory stays 2,018 pairs/9,101 reads awaiting whole-program lowering and shared
reaching-view proof. Ledger checks remain with worker 01a118d0. The whole
repository gate was not run.

## Mixed and null-containing Map storage, October 8

Integration 9b768f65 is merged in 3c3665e1, preserving ArrayReadonly and the
callable lane's DiscardResult metadata. Map entries now use existing owned boxed
union references: numbers own boxes, booleans use immortal boxes, references
retain their producer objects, null is adamic_null and undefined is NULL. Map
constructors, set/get, cloning, values iteration, replacement, deletion, clear
and forEach retain the shared runtime ownership paths. Tagged boolean|undefined
entries now decode through maybeSlot, preserving false/true/undefined separately.
Function-value slots accept boxed union references so Map callbacks can consume
these entries without inventing a second callback ABI.

Producer-to-view comparison checks every source union alternative, the allowed
null/undefined alternatives, finite values, and nested array variance. Writable
Map schemas retain invariance. Physical storage must still match: a numeric Map
cannot be reinterpreted as a boxed-union Map. Cross-representation read adapters
are not supplied by this batch; such reads stop at the Map certificate boundary.

Twelve positives cover nullable number, boolean, string, object, array and mixed
scalar entries with null and null|undefined forms. Three schema mutants stop at
node.value: wrong mixed member, forbidden null and forbidden undefined. The
boolean|undefined control prints false, true, undefined, true, undefined. Clone
and lifecycle iteration matches Node; clearing a Map during its first callback
prints number, 7, 0 and remains sanitizer/leak-clean. A first lifecycle probe
used unguarded String on a union containing null; the existing ToPrimitive
boundary refused it, so the final control narrows its numeric value first.

Five implementation mutations were caught independently: ignore union members,
allow forbidden null, allow forbidden undefined, collapse null's sentinel, and
skip tagged-boolean unpacking. All fail behavioral assertions without a compiler
warning kill. Restored Map/nullish/selection/callable-source oracles pass in
78.426s; the live mutation control passes in 0.755s. Scoped IR/lower/native/JS
View|Map|Contract|Callable checks pass in 0.016s/4.024s/31.072s/0.799s; vet passes.
The broader Census selection still fails five groups. Reverting the callable-slot
change for a baseline comparison reproduces four failing groups, including a
predicate-marker boundary that previously failed with NotYet instead of Refused.
One old predicate-marker test instead expected the now-removed union-slot NotYet;
stored marker calls retain integration's producer arity check at runtime. Both
current and baseline logs are retained; no full gate pass is claimed.

After this push the production table remains 2,018 pairs/9,101 reads awaiting
whole-program lowering and reaching-view proof. Callable entries, tuples,
nominal/phantom witnesses and cross-representation read conversion remain;
ledger-driven checks are untouched.

## Callable Map producer proofs (October 8)

Callable Map entries now have private signature descriptors, preserving the
shared refusal for callable array elements. Constructors, ordinary .set writes,
and copied entry arrays check the actual producer's immutable code identity
before it enters storage. Nullable entries bypass this check only for the
permitted null/undefined representations. Readonly Map comparison is
contravariant in callable parameters and covariant in callable results; mutable
Map comparison retains both directions. Physical calling conventions must match.

Node controls cover plain, null, undefined, combined nullable callable entries,
and a nullable Map clone. Parameter and result schema mutants exit 70 at
node.value. Two IR mutants replace a constructor/.set producer with a closure
whose actual result is string while storage promises number; both backends stop
at the storage site. Disabling each backend's producer guard independently makes
these mutants run on (exit 0), killing the implementation mutation by behavior,
without compiler warnings. Logs are in evidence/map-callable-*.

This batch handles fixed-arity scalar signatures supported by the integration
callable producer table. Void, rest, generic, object/array/union calling
conventions remain outside this certificate. Tuples, nominal/phantom witnesses,
and conversion between physical entry representations remain unsupported.
The exact production inventory remains 2,018 pairs / 9,101 reads pending
whole-program lowering and shared reaching-view proof; fixture controls do not
subtract production rows. Ledger-driven inserted checks remain with their worker.

Restored full Map oracle selection passes in 51.036s (native sanitizer/release,
JS, Node and positive-fixture leak checks). Scoped View|Map|Contract|Callable
package checks pass: lower 4.879s, native 34.342s, JS 0.976s; IR has no matching
tests. Vet passes. The previously recorded broader census failures are unchanged
in scope; this batch does not claim a full repository gate.

## Fixed tuple Map entries (October 8)

Merged integration 8afdc710 before this batch. Fixed, nonempty tuple entries have
per-position contracts in their existing object storage. Readonly tuples permit
position covariance; writable tuples require both directions. Map certificates
also preserve tuple arity and readonly provenance. Optional/rest positions and
callable positions remain refused. Ordinary tuple writes remain a stage-0 gap.
Nullable tuple entries retain the null sentinel and undefined distinctly.
JavaScript narrowing explicitly preserves a tuple's array identity while native
narrowing preserves its object storage; neither changes source object identity.

Node controls print 7, seven, 2 through a helper, also for a nullable tuple entry.
Wrong element schema, wrong length, readonly-to-writable and writable literal
widening mutants stop at node.value. An IR payload mutant puts a string in the
numeric position while preserving the declared producer schema: both backends
stop at helper pair[0]. Removing that read's guard runs on with exit 0 and kills
the implementation mutation by behavior. Evidence is in map-tuple-*.log.gz.
Map/nullish/nullable-selection oracle checks pass in 127.248s; final tuple
selection passes in 10.585s. The exact production inventory remains 2,018 pairs /
9,101 reads pending whole-program lowering and shared reaching-view proof.

## Approved phantom Map brands (October 8)

Map schemas now reuse the existing primitive and array phantom-brand proofs.
Only absent members accepted by phantomBase/phantomArrayBase are erased; a real
intersection or nominal class still lacks this certificate. Primitive brands
preserve finite literal domains. Array brands preserve their element contract
and readonly provenance. Existing nested array/object brand controls now compile.
Node and both backends print the same primitive kind followed by undefined for
the phantom member. A wider string producer cannot certify a finite branded
string: the schema mutant exits 70 at node.value.

The full Map oracle passes in 68.595s. Scoped View|Map|Contract|Callable|Tuple
checks pass (lower 4.466s, native 30.949s, JS 0.763s); vet passes. Adding Phantom
to the selection exposes three inherited groups: both array-brand IR-equivalence
tests and the old cycle-refusal expectation. Reverting view_maps.go to the tuple
checkpoint reproduces all three; logs retain this comparison. No full gate pass
is claimed.

The branded string|branded void constructor probe prints string, undefined, true
on Node but remains a named Map-entry representation refusal. This is a storage
conversion gap, not a completed branded-nullish control. Its executable gap test
pins the refusal. Production demand remains 2,018 pairs / 9,101 reads awaiting
whole-program lowering and shared reaching-view proof; broader callable work
stays with lane 5.

## Readonly Map storage conversions (October 8)

This checkpoint closes the branded string|branded void gap above. Constructor
and .set conversions evaluate branded void once and check that its payload is
undefined before selecting reference storage. The effects control prints
2, string, undefined, true, matching Node.

Readonly Map certificates now admit top-level number/boolean to boxed unions,
number|undefined and boolean|undefined to boxed unions, supported reference
families to boxed unions, number to packed number|undefined, and boolean to
packed boolean|undefined. Mutable Maps retain invariant storage. The Map keeps
its original storage and identity; get, values/entries snapshots, cloning,
forEach, direct loops and stored iterators acquire owned converted values.
Supported reference controls include string, object, array, fixed tuple and
already certified callable entries. This does not broaden callable conventions.

Controls match Node for false, missing versus present undefined, NaN, negative
zero, iterator exhaustion and values held across Map.clear(). A boolean iterator
exposed a packed-byte decoder being interpreted as a pointer; the nullish reader
now unpacks it. IR mutants corrupt the producer storage header and the branded
void payload: unsupported storage and a present payload stop with exit 70.
Ten independent implementation mutations are caught by behavior: numeric boxing,
boolean boxing, numeric presence, boolean presence, numeric undefined, boolean
undefined, reference ownership (ASan use-after-free), packed boolean field
presence, native undefined guard and JavaScript undefined guard. Evidence and
individual logs are in map-storage-mutants.json and map-storage-*.log.gz;
run-map-storage-mutants.py restores every mutated source.

Exact remaining fixture gaps are pinned by TestCheckedViewMapStorageGaps:
optional object/array to nullable unions, nested array element conversions and
key storage conversions stop at node.value with exit 70. Optional/rest tuples
refuse at the named value read during compilation. True nominal/intersection
witnesses, callable tuple positions, ordinary tuple writes and broader callable
conventions remain outside this batch. Array snapshots with boxed-union or
packed-boolean elements still encounter the existing array representation gap;
direct Map value loops are covered. Production inventory remains 2,018 pairs /
9,101 reads pending whole-program lowering and shared reaching-view proof.
Fixture controls are not subtracted from that inventory.

Uncached oracle selection Map|Nullish|NullableSelection passes in 128.420s
(native sanitizer/release, JavaScript, Node and positive-fixture leak checks).
Scoped View|Map|Contract|Callable|Tuple checks pass: IR 0.014s, lower 5.460s,
native 41.677s, JavaScript 1.033s. Vet passes. Added gap controls pass in 1.903s
and boolean mixed-union control in 0.516s. Previously documented broader gate
failures remain outside this selection; no full repository gate pass is claimed.

## Optional reference conversion admission (October 8)

Merged integration 322bd65d before this group. Optional object and array source
contracts now compare their present payload through private descriptor copies;
the target must separately admit undefined. Original producer descriptors and
IDs remain unchanged. String, object, array and already certified callable
controls cover get, present undefined, identity, values/entries loops, cloning
and callbacks across Map.clear(), matching Node and passing leak checks. Targets
excluding undefined stop at node.value for all four source families. Removing
the separate undefined guard makes object and array mutants run on with exit 0;
the behavioral mutant is recorded and reproducible with the storage runner.

Uncached Map|Nullish|NullableSelection oracle selection passes in 175.085s.
Optional controls pass in 6.151s. Scoped View|Map|Contract|Callable|Tuple packages
pass: IR 0.017s, lower 6.324s, native 55.113s, JavaScript 1.431s; vet passes.
Evidence is in map-optional-reference-*.log.gz. Optional/rest tuples now belong
to lane 4c and are not developed here. Nested element and key storage conversion,
and nominal runtime witnesses remain next. The conservative nullish inventory
still has 2,018 pairs / 9,101 reads awaiting production reachability and lowering;
these are unmeasured production rows, not a claim that every row still refuses.
The overlapping intersection inventory is 15 pairs / 42 reads; fixture success
cannot establish how many production rows in either inventory are discharged.

## Nested numeric and numeric key conversions (October 8)

Readonly nested arrays now admit number to number|undefined through the shared
array extraction adapter. Helper indexing, iteration, join, map/filter/forEach,
slice and includes match Node, also two levels deep. Array identity stays true
and source writes remain observable. Incompatible element schemas and mutable
array widening stop at node.value. An IR payload mutant keeps the numeric
producer certificate but stores strings: both backends stop at the helper's
values[element]. Disabling the native extraction guard runs on and kills that
implementation mutant by behavior.

Readonly Map keys admit number to number|undefined. Constructor-owned key storage
stays numeric; get/has convert incoming queries and reject undefined as a miss.
Key reads, values/entries iteration, snapshots, clone, callbacks and stored
iterators preserve SameValueZero: undefined is separate from NaN and zero,
-0 looks up zero, and an undefined key added to a cloned Map stays distinct from
NaN. Snapshot arrays now record their actual converted storage. A control enables
the shared array checking policy alongside numeric keys. A Boolean-storage IR
mutant with a Number certificate stops at native lookup with exit 70, reporting
expected boolean and found number|undefined. Removing query presence or key-read
presence independently fails the Node comparison.

Uncached Map|Nullish|NullableSelection regression passes in 155.765s. After final
snapshot metadata, the Map oracle selection passes in 83.633s. The diagnostic-only
follow-up and payload mutants pass in 16.080s. Scoped packages pass: IR 0.010s,
lower 5.190s, native 41.234s, JavaScript 1.015s; vet passes. Gap probes pass in
1.493s. Evidence and three behavioral mutants are in map-nested-key-*. Named
remaining fixture gaps are nested boxed unions, nested packed booleans and boxed
key domains, all stopping at node.value. Optional/rest tuples stay with lane 4c.
Nominal witnesses are next. Production reachability remains unmeasured for the
conservative 2,018 pairs / 9,101 reads; this group does not subtract fixture
controls from those production totals or claim a full repository gate.

## Nominal data-class Map witnesses (October 8)

Private entry descriptors now reuse existing erased class identities. Guards at
constructor, .set and copied-entry stores verify nominal key and value producers
before a Map receives its declared certificate. Structural lookalikes cannot
acquire that certificate. Private descriptors do not enable unrelated nominal
view paths. Data-class fields still carry their logical constraints and checked
reads. Parent/child identity uses existing class ancestry; generic data-class
controls retain their concrete field contracts. No callable convention changes.

Node and both backends agree for nominal keys/values, derived-to-base covariance,
generic data classes, null/undefined widening, actual nullish entries, clones and
single evaluation (two constructor calls stay two). Six IR mutants substitute a
structural lookalike at key/value constructor, .set and copy sites: both backends
stop with exit 70, naming the store and expected class identity. Disabling each
backend's guard independently makes all six mutants run on; both implementation
mutations are caught by behavior without compiler warnings. Reproduce them with
run-map-storage-mutants.py nominal-native nominal-javascript.

Callable-bearing classes and nominal classes beneath entry aggregates remain
named refusals at node.value. Their unread fixtures print box and match Node.
Required primitive brands are a separate existing language boundary: both read
and unread controls refuse the non-void brand at its type declaration, before a
view read. This is not claimed as lazy contract support. Optional/rest tuples
belong to lane 4c; broader callable conventions remain with lane 5.

Uncached Map|Nullish|NullableSelection selection passes in 160.824s; all nominal
controls, six producer mutants and gap controls pass in 8.562s. Scoped
View|Map|Contract|Callable|Tuple packages pass: IR 0.018s, lower 5.162s,
native 42.069s, JavaScript 0.908s; vet passes. Evidence is in map-nominal-*.
No full repository gate pass is claimed.

The lane's conservative demand inventory remains 2,018 pairs / 9,101 reads.
Exact production support, reaching-view proof and discharged rows are unmeasured;
these totals are not counts of remaining compile refusals. The overlapping
intersection family has 15 pairs / 42 reads. Array/tuple rows are combined in the
inventory, so transferring tuples does not justify subtracting every array row.
The current executable gap table and these measurement limits are in
lane-remaining.json. Remaining storage gaps are nested boxed unions, nested
packed booleans and boxed key domains; each stops at node.value in its fixture.

## October 8 continuation blocked by integration rollback

Owner tip 0ea22091 remains green under its recorded scoped gates. Fetching the
designated integration branch returned bd05075f5b8cbd592e8ece1a098d582eb43cf510.
Its report says the owner merge was rejected: the ordinary
censusCallableSlotless Union exception loses the mixed-union callback ABI
refusals, while restoring that guard refuses entry-live-mutation.a:5:82.
The integrator reverted the entire owner item, including Map storage adapters.
An attempted merge produced content and modify/delete conflicts; it was aborted
without selecting either side or changing compiler code. The designated
integrator owns reconciliation, and broader callable conventions belong to
lane 5. This blocks building the next groups against the newest shared tree.

Pending groups remain nested boxed union storage, nested packed boolean storage,
boxed key domains, and nominal stores beneath entry aggregates. The conservative
demand inventory remains 2,018 pairs / 9,101 reads; exact remaining reaching-view
pairs and reads are unmeasured. The array/tuple overlap is 1,195 reads, not an
exact packed-boolean or nested-union count. No new admission, test pass, or mutant
kill is claimed for this documentation checkpoint.

Required primitive brands hit the existing language refusal in
internal/lower/phantom_brands.go:133, function (*lowering).phantomRefusal:
"Adamic 0.1 refuses a primitive brand member brand whose type is not void; make
brand void (or optional and typed undefined) so the brand is phantom". Both read
and unread controls refuse at the type declaration. This is not a lazy-view
read refusal.

## October 8 nested primitive boxed arrays

Continued on the owner branch per the lead, without adopting the reverted
callable boundary or changing primitive-brand refusals. Readonly Map certificates
now admit nested array conversions from scalar/packed-number storage to primitive
boxed union elements. A private temporary array owns newly boxed numeric reads
for each consuming operation; the original array keeps its physical storage and
identity. This owner can retain O(consumed elements) boxes until operation exit.
Escaping results retain their own references. No second flow graph was added.

Each extraction validates full primitive union membership, including finite
literals and distinct null/undefined. Helpers, indexes, iteration, join, map,
filter, forEach, reduce and searches use the shared read adapter; copied slices
retain source storage metadata. Searches compare boxed primitive payloads and
implement includes(NaN). Unwired boxed-array consumers and spreads refuse at
the consumer. Nullable primitive String conversion uses the existing union
runtime. Array templates use the checked join loop; object/callable union arrays
remain outside this admission. Packed boolean sources remain the next group.

Node controls cover number/boolean/string widening, NaN, negative zero, identity,
original-alias pushes, mixed entries containing real null and undefined, finite
literals and helper consumers. Wrong schema stops at node.value. A real IR
producer mutation stores strings under the unchanged number certificate and
stops at values[element] in both backends. A second stores actual 2 under the
unchanged literal-1 certificate and stops at values[0] in both backends.

Four independent implementation counterfactuals were run and restored: omit
native membership, omit JavaScript membership, box number 0 instead of the read
number, and omit the temporary owner's take. The first two let the finite mutant
run on and are caught; boxing 0 disagrees with Node; dropping the owner leaks
1,080 bytes in 45 allocations under LeakSanitizer. All are semantic failures,
not compiler warnings. The reproducible runner and raw compressed logs are
under evidence/map-nested-boxed-* and evidence/map-boxed-*.

Validation: uncached Map|Nullish|NullableSelection oracle 169.625s passes; touched
IR/lower/native/JavaScript packages with View|Map|Contract|Callable|Tuple and the
recorded Phantom baseline exclusion pass (.012s, 4.269s, 63.318s, 1.091s); whole
vet passes. Expanded controls and finite payload mutant pass in 4.803s. The full
repository gate and production reachability census were not rerun. Remaining
conservative inventory is still 2,018 pairs / 9,101 reads, with exact remaining
production pairs/reads unmeasured. Reduced pending groups are packed booleans,
boxed key domains and nominal entries beneath aggregates.

## October 8 packed boolean arrays

Readonly Map certificates now admit boolean array elements widened to packed
boolean|undefined, and packed boolean elements widened to boxed nullish unions.
Array metadata records storage 9 separately from pointers. The shared decoder
unpacks the byte before reading false/true/presence; absent snapshots use byte 2,
and ordinary boolean reads widened to storage 9 are explicitly packed present.
Scalar writes into the same packed representation are checked and permitted.
Find unpacks the returned byte instead of interpreting its value as presence.

Node controls cover true, false, stored undefined, alias pushes, identity, join,
index, for-of, map, filter, forEach, at, find, some/every, reduce, includes, indexOf
and lastIndexOf. The original unread-elements probe is admitted. Wrong element
schema stops at node.value. A real string producer under the unchanged boolean
certificate stops at helper values[element]; a false producer under declared
true|undefined stops at values[0], both backends.

Consumer audit: JSON.stringify bypasses checked array extraction and its runtime
schema decoder cannot use a converted alias safely. Boxed/packed boolean array
arguments now refuse by name at that consumer. Unwired array operations and
spreads remain named consumer refusals. No primitive-brand rule or callable
convention was changed.

Five runtime implementation mutations were independently run and restored:
drop the packed presence test, pack every boolean absent, use false for an absent
callback/loop snapshot, and drop the native/JavaScript finite literal checks.
Each loses Node behavior or lets the false payload mutant run on. A sixth,
compile-admission counterfactual drops the JSON boundary and loses both required
consumer refusals; it is not claimed as a native runtime kill.

Validation: expanded Node controls/IR mutants/consumer refusals pass in 5.510s;
uncached Map|Nullish|NullableSelection|TestCheckedViewArrayJSONStorageRefusal
oracle passes in 173.183s; scoped IR/lower/native/JavaScript packages pass in
.012s, 5.348s, 43.445s and 1.121s, with the previously recorded Phantom baseline
exclusion; whole vet passes. Raw logs and all six counterfactual results are
preserved in evidence/map-nested-packed-* and evidence/map-packed-*. No full
repository gate or new production census is claimed.

The conservative inventory remains 2,018 pairs / 9,101 reads; exact remaining
production pairs and reads are unmeasured. Reduced pending conversion groups are
boxed key domains and nominal stores beneath entry aggregates. JSON remains a
named consumer gap, and broader callable-bearing nominal stores stay with lane 5.

### October 8: primitive boxed Map key domains

Readonly Map certificates now admit numeric, boolean, string, packed optional numeric
and packed optional boolean keys into complete primitive boxed unions. Boolean keys
also widen to packed optional boolean. Lookup converts the query back to constructor
storage; snapshots, callbacks, direct loops and stored iterators own converted key
references. Mutable physical key storage remains invariant. Boxed primitive Maps use
SameValueZero, including NaN, dynamic string equality and canonical positive zero;
normalizing a key does not mutate the caller's original negative-zero box.

Eight restored controls match Node in sanitized native, release native and JavaScript,
with no leaks (8.289 seconds). The uncached Map/nullish/nullable-selection/JSON-boundary
gate passed in 194.217 seconds. Touched-package checks passed (IR 0.013s, lower 5.448s,
native 43.813s, JavaScript 0.933s), and go vet ./... passed. The full repository gate
was not run; the scoped package command retained the documented Phantom exclusion.

Seven independent implementation mutants were caught and restored: identity equality,
NaN inequality, wrong numeric payload, missing reference retention (AddressSanitizer
heap-use-after-free with dynamic strings), address-based key hashing, omitted negative
zero normalization, and lost undefined-query decoding. An initial address hash mutant
with only aligned low bits survived because all keys collided; the recorded mutant
shifts the address so equal fresh boxes choose different buckets. Evidence is in
map-boxed-keys-*.log.gz and map-boxed-key-mutants.json.

The exact production reaching-view discharge remains unmeasured: the conservative
inventory is still 2,018 pairs / 9,101 reads. Closing fixtures is not a production count.
Nested nominal entry aggregates and boxed/packed JSON extraction remain named gaps;
callable-bearing classes belong to lane 5. Required primitive brands remain refused
by internal/lower/phantom_brands.go, (*lowering).phantomRefusal: a required brand member
must be void (or optional and typed undefined). No primitive-brand rule changed.

### October 8: nested immutable nominal witnesses

Map certificates now carry data-class identities through readonly structural entry
fields and the existing fixed tuple adapter. Class fields also admit null, undefined
and both. Native and JavaScript recursively validate the nominal leaf at constructors,
set calls and copied entries. Every field result and class receiver registered by these
private witnesses checks class identity at its read, including ordinary and generic
helpers, callbacks, direct loops and aliases held in fields. The canonical general
class-view descriptor remains unsupported; a private witness discharges only property
reads whose backend emits the identity check. Required primitive brands are unchanged.

Six real lookalike payload mutants are caught at constructor, set, copied entry,
field result, ordinary receiver and generic receiver boundaries. Four independent
native/JavaScript producer/read guard counterfactuals were caught and restored. The
producer-only fixture performs no descendant reads, so a later read guard cannot mask
a disabled producer guard. A distinct class with identical public fields also stops
at node.value when its Map schema differs. Existing root nominal guard counterfactuals
were rerun and caught as well.

The uncached Map/nullish/nullable-selection/JSON-boundary gate passed in 193.944 seconds.
Touched-package checks passed (IR 0.012s, lower 5.727s, native 38.086s, JavaScript 1.153s);
go vet ./... passed. After restricting private witness discharge to property reads and
isolating the producer-only controls, the focused oracle passed in 12.756 seconds.
The final restored oracle passed in 9.499s; lower/IR checks passed in 3.629s/0.015s.
Logs are recorded under evidence/map-nested-nominal-*.
The full repository gate was not run; the scoped package command retains the documented
Phantom exclusion.

Readonly arrays of classes, mutable structural paths and recursive aggregate paths
retain named refusals at node.value; each has read/unread controls, and every unread
control matches Node. Recursive private descriptors require a runtime visited witness
before their producer schema can be certified. Nullable structural aggregate roots and
optional structural fields are also outside this bounded adapter. Callable-bearing
classes remain with lane 5; optional/rest tuples remain with lane 4c. Exact production
reaching-view counts remain unmeasured (conservative inventory: 2,018 pairs / 9,101 reads).

### Nominal arrays, October 8

Private Map witnesses now traverse finite reference arrays containing data
classes, including nested arrays and class-or-undefined elements. Shared array
extraction checks class identity at an indexed read or callback/loop extraction.
Array literals and length constructors retain their own logical element contract;
a mutable source alias checks incoming class identity without rewriting that
contract to a wider viewed type. Sparse producer scans skip absent indices and
check present undefined using the declared element contract.

Controls cover array methods, helper reads, nested arrays, private data fields,
source alias pushes, generic optional elements, and sparse arrays. Three isolated
payload mutants substitute a structural lookalike at a producer, element read,
and mutable source push. Six implementation counterfactuals disable each backend's
corresponding check; all fail behaviorally, without a compilation failure.
Evidence is under `evidence/map-nominal-array-*`.

Class-or-null boxed array elements, optional structural fields, mutable structural
entry witnesses, recursive nominal aggregates, and uncertified nonliteral source
write contracts remain outside this group. Callable conventions remain with lane
5. Production reaching-view discharge is unmeasured; the 2,018-pair/9,101-read
inventory is not a compilation count. Required primitive brands retain the
existing language refusal in `internal/lower/phantom_brands.go:133`,
`(*lowering).phantomRefusal`: a primitive brand member whose type is not void must
be made void (or optional and typed undefined) so the brand is phantom.

Nominal array validation: the scoped uncached oracle (`Map|Nullish|
NullableSelection|TestCheckedViewArrayJSONStorageRefusal|TestCheckedViewNominalArray|
TestArrayHoles(Milestone|Refusals)$`) passed in 177.796 seconds. Touched-package
checks (`View|Map|Contract|Callable|Tuple`, with the inherited Phantom exclusion)
passed: IR 0.011s, lower 5.954s, native 35.910s, JavaScript 0.981s. Scoped vet passed.
The complete repository gate was not run for this group.

### Mutable nominal fields, October 8

Finite structural entry descriptors retain each field's actual mutability rather
than refusing all writable paths. Class-valued source slots use a private complete
nominal certificate. The existing checked field-write path validates the incoming
class and its recursively certified data fields; mutable assignment compatibility
still requires both directions. Every escaped class field result also checks
identity, independently of the producer or write certificate.

Controls compare Node with native sanitized/release and JavaScript for a mutable
Map entry and source-alias replacement. An isolated write payload mutant supplies
a structural lookalike and stops at the write. A second mutant removes the write
certificate to model unchecked corruption and stops at the next field read.
The corresponding four backend guard bypasses all fail behaviorally. Evidence is
under `evidence/map-nominal-mutable-*` and `evidence/map-mutable-nominal-*`.

Optional aggregate fields, nullable aggregate adapters, nullable class source-slot
widening and recursive nominal aggregates remain outside this bounded group.
Broader callable conventions stay with lane 5. Exact production discharge remains
unmeasured; no fixture closure is subtracted from the conservative demand inventory.

Mutable nominal validation: the scoped uncached oracle (`Map|Nullish|
NullableSelection|TestCheckedViewArrayJSONStorageRefusal|TestCheckedViewNominalArray|
TestCheckedViewMutableNominal`) passed in 188.573 seconds. Touched-package checks
(`View|Map|Contract|Callable|Tuple`, inherited Phantom exclusion) passed: IR 0.014s,
lower 5.371s, native 36.508s, JavaScript 0.999s. Scoped vet passed. The complete
repository gate was not run for this group.

### Nullable nominal aggregates, October 8

Finite private nominal aggregate descriptors now have nullish wrappers for
`Entry | null`, `Entry | undefined`, and `Entry | null | undefined`, preserving
the physical representation. Controls cover immutable structural entries and
nominal array entries, present members and stored missing values. Producers
traverse present aggregates and skip only the declared sentinel alternatives.
Existing class identity checks remain independent at escaped field/element reads.

Five producer-only payload mutants test nested structural lookalikes in all
three variants and the excluded nullish sentinel in each single-sentinel variant.
The two backend producer guard bypasses fail behaviorally across these fixtures.
The opposite-null mutant uses the ordinary boxed null IR; JSON's special null IR
would instead have supplied a native undefined pointer and is not evidence for
this check. Evidence is under `evidence/map-nominal-aggregate-*` and
`evidence/map-aggregate-nominal-*`.

Optional aggregate fields, nullable class source-slot widening, boxed class/null
array elements and recursive nominal aggregates remain outside this group.
Boxed/packed array JSON extraction retains its named consumer refusal. Callable
conventions stay with lane 5 and required primitive brand refusals stay unchanged.
Exact production reaching-view discharge remains unmeasured.

Nullable aggregate validation: the scoped uncached oracle (`Map|Nullish|
NullableSelection|TestCheckedViewArrayJSONStorageRefusal|TestCheckedViewNominalArray|
TestCheckedViewMutableNominal|TestCheckedViewNullableNominalAggregate`) passed in
183.157 seconds. Touched-package checks (`View|Map|Contract|Callable|Tuple`, inherited
Phantom exclusion) passed: IR 0.017s, lower 5.474s, native 36.916s, JavaScript 1.033s.
Scoped vet passed. The complete repository gate was not run for this group.

### Optional nominal fields, October 8

Finite structural producer descriptors now admit readonly optional nominal
fields. Native producer extraction passes the optional field's absence allowance
to the shared nullish reader, retaining readiness and actual present-value checks.
Class identity remains checked independently at subsequent reads. JavaScript's
ordinary-object producer predicate now rejects arrays, Maps and Sets, so an
all-optional schema cannot admit an incompatible container merely because its
fields are absent.

Node/native/JavaScript controls distinguish absent `child` from present undefined:
stdout for the presence control is `7\nmissing\nmissing\n7\ntrue\n\nchild\n`.
The true line is `Object.hasOwn(present, 'child')`; empty and present key lists are
empty and `child`. `Object.hasOwn` on the absent binding remains excluded by its
existing declared-public-field/complete-literal-shape rules; this group does not
change reflection admission.

Three payload mutants supply a structural lookalike at a producer, corrupt a
field after production, or supply an array to an all-optional object schema. Four
identity-check bypasses and a separate JavaScript object-kind bypass all fail
behaviorally. Evidence is under `evidence/map-nominal-optional-*` and
`evidence/map-optional-nominal-*`. Mutable optional nominal writes, nullable class
source-slot widening, boxed class/null array elements and recursive nominal
aggregates remain outside this group. Exact production discharge is unmeasured.

Optional nominal validation: the first scoped oracle exposed a fixed-tuple
JavaScript predicate regression, which was corrected without changing tuple
conventions. The restored optional/tuple controls passed in 3.295 seconds, and
the final scoped uncached oracle (`Map|Nullish|NullableSelection|
TestCheckedViewArrayJSONStorageRefusal|TestCheckedViewNominalArray|
TestCheckedViewMutableNominal|TestCheckedViewNullableNominalAggregate|
TestCheckedViewOptionalNominal`) passed in 165.989 seconds. Touched-package checks
(`View|Map|Contract|Callable|Tuple`, inherited Phantom exclusion) passed: IR 0.011s,
lower 4.961s, native 34.598s, JavaScript 0.964s; the final JavaScript check passed
in 0.682s. Scoped vet passed. The complete repository gate was not run.

### Checked JSON array extraction, October 8

JSON array descriptors now carry the existing shared array read metadata. Native
serialization calls the physical adapter at each present element and validates
primitive membership, including finite literals and distinct null/undefined
alternatives. Scalar snapshots are boxed into an operation-owned scratch array
and released after serialization. Nested array and literal descriptors preserve
these readers; the replacer key list uses the same extraction. Source arrays and
their constructor storage stay unchanged. JavaScript applies equivalent element
checks during normalization after all arguments have been evaluated. Generic
helper elements are concretized before recording their nullish alternatives.
Map and function elements get physical kind checks before opaque serialization;
functions are not invoked and callable conventions are unchanged.

Fourteen Node/native sanitized/release/JavaScript controls cover numeric and
boolean widening, packed false/undefined/true, mixed primitives and dynamically
built strings, finite number/string domains, nested arrays/literals, replacer key
lists, sparse boolean holes, argument mutation/replacement, generic nullable
helpers, Maps and functions. Six payload mutants supply excluded finite values,
key values, nested values, or incompatible Map/function representations. Five
behavioral counterfactuals remove backend domain checks, change the native opaque
representation expectation, omit generic concretization, or omit scratch-owner
release. The latter is caught by leak detection; none relies on a build failure.
Evidence is under `evidence/json-array-*`.

A full filesystem initially prevented a runtime build. Only stale oversized Go
build-cache artifacts were removed, recovering roughly 28 GB; tests were then
rerun. A stock-JSON filter initially selected no tests and was corrected to the
full path-based subtest names. All nine stock JSON fixtures passed in 2.154s.
Final scoped uncached oracle (`Map|Nullish|NullableSelection|JSONArray|
ArrayJSONStorage|NominalArray|MutableNominal|OptionalNominal|NullableNominal|
JSONStringify|TestArrayHoles(Milestone|Refusals)$`) passed in 213.595s. Final touched
packages (`View|Map|Contract|Callable|Tuple|JSON|Generic`, inherited Phantom
exclusion) passed: IR 0.016s, lower 7.030s, native 37.920s, JavaScript 1.117s.
Scoped vet passed. The complete repository gate was not run.

Nullable length constructors retain their existing representation refusal; sparse
JSON is covered through a supported boolean constructor and a nullable view.
Nonprimitive boxed JSON members, hidden object/toJSON behavior and recursive
array types retain named refusals. Boxed class/null array reads, mutable optional
nominal writes, nullable nominal slot widening and recursive nominal witnesses
remain outside this group. Production reaching-view discharge is unmeasured;
no fixture closure is subtracted from the 2,018-pair/9,101-read inventory.

### Nullable class array elements, October 8

Finite data-class array elements now retain private nominal witnesses across
class/null, class/undefined and class/null/undefined representations. The native
physical adapter selects reference snapshots for nullable classes; the independent
read guard checks class identity and declared sentinels. Producer scans use the
same physical adapter followed by their own recursive certificate, preserving
independent guard evidence. JavaScript checks each present element through the
same nominal descriptor. Holes remain distinct from present undefined.

Constructor-owned class element contracts govern push and indexed source-alias
writes. Readonly Map arrays may widen class references into nullable class
references without changing the source representation or its original write
contract. Fifteen Node controls cover extraction, helpers, generic scalar helpers,
callbacks, loops, slices, nullable writes and readonly widening. Twelve isolated
payload mutants cover producer/read/write lookalikes and rejected null/undefined writes to
an original nonnullable source after readonly widening. Seven implementation
counterfactuals remove each backend producer/read/write guard or the readonly
storage adapter; each was killed by behavior, not a build failure.

Initial probes exposed the existing boxed class array lowering refusal and C
pointer comparison issues. The first broad gate additionally found the missing
readonly reference storage certificate; the corrected controls pass. Class join
stringification keeps its existing named refusal. Evidence is under
`evidence/nullable-elements-*`. Remaining owned scope includes mutable optional
nominal slots, nullable nominal source-slot conversions and recursive nominal
aggregates. Required primitive brands remain refused by
`internal/lower/phantom_brands.go:133`, `(*lowering).phantomRefusal`:
`Adamic 0.1 refuses a primitive brand member brand whose type is not void; make brand void (or optional and typed undefined) so the brand is phantom`.
The conservative inventory remains 2,018 pairs / 9,101 reads; exact production
reaching-view discharge is unmeasured and fixture closures are not subtracted.

Nullable class element validation: the final restored scoped uncached oracle
(`Map|Nullish|NullableSelection|JSONArray|ArrayJSONStorage|NominalArray|
MutableNominal|OptionalNominal|NullableNominal|JSONStringify|
TestArrayHoles(Milestone|Refusals)$`) passed in 253.639 seconds. Final touched
packages (`View|Map|Contract|Callable|Tuple|JSON|Generic`, inherited Phantom
exclusion) passed: IR 0.017s, lower 10.007s, native 51.276s, JavaScript 1.813s.
Scoped vet passed. The undefined-only source-widening mutant was tightened to
write undefined, which the target permits, isolating the source guard; all three
restored source mutants passed in 1.625s and both backend source-write bypasses
were killed again. The complete repository gate was not run.

### Optional mutable nominal references, October 8

Finite structural entry descriptors now admit optional mutable class/undefined
members. Their complete private class descriptors compare through the shared
reference representation while preserving undefined admission. Existing writes
continue to consult the actual object's constructor slot contract. Missing fields
remain absent; a write to an existing slot changes its payload and readiness.
Four Node controls cover producer, read and write use, absent entries, helper and
generic writes, and a present slot changed to undefined: keys still print `child`
and hasOwn prints `true`. Three payload mutants isolate fake class identity at
producer, read and write. A fourth source program passes a tagged viewed object
with a narrower actual class slot into a nullable helper; Node runs on, while
both backends must stop at the incompatible slot write.

The narrowed helper probe initially ran on. Its source certificate was computed
and discarded while lowering the helper before a later view activated the field
name. Class source certificates now remain on those writes, so backend activation
uses the complete declared source type rather than a physical object tag. This
uses existing write metadata and shared flow, with no second flow graph.

Eight implementation counterfactuals remove either backend producer/read/write
check, erase undefined admission from slot comparison, or discard the helper's
early source certificate. Each is killed by behavior. Positive controls alone
initially failed to kill the undefined-admission counterfactual; those survivor
logs are retained. The initial ordinary narrowing fixture met the existing static
mutable-invariance refusal, so the final mutant uses an admitted tagged view.
Initial reflection controls hit existing Object.keys/in refusals; the supported
plain-literal reflection control is held to Node. Evidence is under
`evidence/optional-mutable-*`.

Escaped aliases still cannot add an absent property without a declared physical
slot; nullable class/null source-slot conversions and recursive nominal witnesses
remain outside this group. Required primitive brands and broader callable
conventions are unchanged. The conservative inventory remains 2,018 pairs /
9,101 reads; exact production reaching-view discharge is unmeasured.

Optional mutable validation: the final restored scoped uncached oracle
(`Map|Nullish|NullableSelection|JSONArray|ArrayJSONStorage|NominalArray|
MutableNominal|OptionalNominal|NullableNominal|JSONStringify|
OptionalMutableNominal|TestArrayHoles(Milestone|Refusals)$`) passed in 306.628s.
Touched packages (`View|Map|Contract|Callable|Tuple|JSON|Generic|Class`, inherited
Phantom exclusion) passed: IR 0.048s, lower 12.862s, native 49.381s, JavaScript
1.276s. Scoped vet passed. The complete repository gate was not run.

### Nullable nominal source-slot conversions, October 8

Finite class/null and class/null/undefined reference members now use the checked
boxed slot adapter. Complete private class descriptors compare their logical
null and undefined admissions separately from their common reference storage.
Direct null literals receive a source certificate only through this adapter;
other unsupported slot conversions keep their existing refusals. Native writes
retain the incoming reference, release the old reference, and stamp the actual
boxed storage type. Optional existing nullable class slots use the same adapter.

Nine Node controls cover class/null/both writes, initially null payloads, optional
existing slots, and independently read-only producer/read probes. Three class
lookalike payload mutants isolate producer, read, and write identity guards. A
fourth probe writes a nullable value through a tagged view into an actual required
class slot: Node runs on, both native modes and JavaScript stop at that write.
Eight counterfactuals remove each backend identity guard, erase null admission
from source containment, or bypass the unboxed boundary refusal. All were killed
without a build failure. Evidence is under `evidence/nullable-slots-*`.

The helper controls exposed a separate convention gap: a nullable class argument
can arrive as an unboxed Object reference, so null and undefined cannot both be
recovered at the field store. Such writes now have a named refusal at the source
value: `writing a nullable nominal value from an unboxed reference; preserve null
and undefined through the callable boundary`. Three refusal controls are held to
Node. The broader callable convention stays with lane 5; no ABI change was made.
An initial payload mutant targeted a literal null's certificate and survived;
the isolated write probe now corrupts the class-valued write instead. Initial
failure and survivor observations are preserved in the expanded probe log.

Recursive nominal producer descriptors still need a runtime visited witness;
escaped aliases adding absent slots, nullable array length constructors, boxed
nonprimitive JSON and recursive JSON remain outside this group. Required
primitive brands retain `(*lowering).phantomRefusal` in
`internal/lower/phantom_brands.go:133`: `Adamic 0.1 refuses a primitive brand member
brand whose type is not void; make brand void (or optional and typed undefined)
so the brand is phantom`. Conservative inventory remains 2,018 pairs / 9,101
reads; exact production reaching-view discharge is unmeasured.

Nullable slot validation: the restored scoped uncached oracle (`Map|Nullish|
NullableSelection|JSONArray|ArrayJSONStorage|NominalArray|MutableNominal|
OptionalNominal|NullableNominal|JSONStringify|OptionalMutableNominal|
TestArrayHoles(Milestone|Refusals)$`) passed in 231.588s. Touched package scope
(`View|Map|Contract|Callable|Tuple|JSON|Generic|Class`, inherited Phantom exclusion)
passed: IR 0.020s, lower 8.677s, native 37.309s, JavaScript 1.032s; scoped vet
passed. After restricting literal-null certificates to nominal union stores, the
final focused uncached oracle passed in 6.908s and final IR/lower checks passed
in 0.011s/4.014s; scoped vet passed again. The full repository gate was not run.

### Readonly nominal field storage widening, October 8

Readonly class fields beneath Map entries reuse the existing counted-reference
adapter when a source Object slot is read as a nullable Union slot. Logical
covariance and nullish containment remain separate checks, and mutable fields
still require both logical directions. The actual producer storage and slot
contracts remain unchanged. Direct and nested required class fields widen to
class/null, class/undefined and class/null/undefined. Class/undefined source fields
also widen to boxed both and can change to a class and back to undefined through
an original mutable alias.

Eight Node controls pass on sanitized native, release and JavaScript. Initial
controls measured the missing storage certificate: null and both failed while
undefined already passed. Later alias writes exposed missing source-write metadata
because the producer schema called the field readonly. A readonly schema can
describe a mutable allocation held by another alias, so all private nominal field
names retain source-write certificates. No callable convention changed.

Two payload mutants independently corrupt an original Object slot with a class
lookalike after Map producer validation: each stops at its nullable field read.
Four check-removal counterfactuals omit readonly storage compatibility, readonly
schema alias-write registration, or either backend read identity guard. All were
killed without a build failure. Evidence is in `evidence/nominal-slot-widen-*`.
The conservative inventory remains 2,018 pairs / 9,101 reads, with production
reaching-view discharge unmeasured. Recursive nominal producer schemas need a
runtime visited witness; unboxed nullable callable values retain a named refusal
and broader callable conventions remain with lane 5. Required primitive brands
retain the existing language refusal.

Readonly field widening validation: the restored scoped uncached oracle
(`Map|Nominal|TestArrayHoles(Milestone|Refusals)$`) passed in 191.211s.
Touched packages (`View|Map|Contract|Generic|Class`, inherited Phantom exclusion)
passed: IR 0.016s, lower 9.696s, native 40.612s, JavaScript 1.104s. Scoped vet
passed. The focused restored controls and payload mutants passed in 4.637s.
The full repository gate was not run.

## October 8 owner merge of integration ffe428ab

Merge resolution restores the owner Map descriptor, producer, tuple and storage
adapter dependencies while retaining integration's boxed callable dispatch and
producer guards. The callable boxing implementation is unchanged from ffe428ab.
Map reads return owned reference snapshots; a Map callback releases its key and
value snapshots both normally and on throw. Set callbacks keep integration's
borrowed-key convention. The counted map-parameter control records one additional
owned key snapshot, with balanced allocation/free totals.

Conservative assumption: the old slotless Map value and unsupported key refusals stay at every
Map.forEach call, even where another callback convention could eventually prove
it safe. Constructors, get and iterators retain the owner's storage support.
entry-live-mutation.a is unchanged and refuses at 5:69 with
"a Map forEach callback with a slotless value representation"; Node prints
"number\n7\n0\n". Forty-seven callback sites have explicit refusal assertions.
Forty-six iterator companions separately retain Node-held storage controls and
payload mutants, rather than counting compile failures as storage proof.

The conservative demand inventory remains 2,018 pairs / 9,101 reads. Exact
remaining production pairs and reads remain unmeasured. The fixture gap table
adds the slotless Map callback boundary without subtracting fixture cases from
production totals. Recursive nominal witness work is unfinished and excluded
from this merge. Required primitive brands retain the existing language refusal.

The required primitive brand refusal remains in
internal/lower/phantom_brands.go:133, (*lowering).phantomRefusal:
"a primitive brand member brand whose type is not void" (fix:
"make brand void (or optional and typed undefined) so the brand is phantom").

The 39 existing Map/Set callback runtime controls pass against Node (8.956s),
including object, array and closure keys, named callbacks, delete/clear/growth,
reassignment, retained values and throws. Removing key snapshot retention in an
isolated copy causes real AddressSanitizer heap-use-after-free in
map_foreach_keys.a and map_foreach_named_keys.a. Restoring the retain makes the
same controls pass (0.662s). No build failure is counted as mutant evidence.
Original mixed callback ABI and wider-helper never-refusal controls pass using
complete declarations emitted from pinned, unmodified TypeScript (3.292s).

Scoped compiler packages pass: IR 0.014s, lower 13.148s, native 55.921s and
JavaScript 1.357s (View|Map|Contract|Callable|Tuple|Nullish|Nullable, excluding
Phantom); lowering passes again after the final key refusal (4.773s). Vet passes.
Extra Phantom tests have four failures; an isolated unmodified ffe428ab run
reproduces all four: three erased-IR metadata comparisons and the cycle refusal.
These are inherited and are not hidden by a full package/gate pass claim.
Evidence and reproduction details are in evidence/merge-ffe428ab*.

The filtered ordinary Map callback count measurement remeasures 39 selected
fixtures and updates only those rows; its whole-table comparison fails because
the selection omits other fixtures. That run is measurement evidence, not a
counts-gate pass. Allocations equal frees in every selected callback row. The
extra retained key snapshots and integration's boxed dispatch remain visible in
the reference-count changes. No whole counts or repository gate is claimed.

The full fresh oracle also located two nonconflicting rollback removals. Restore
MaybeBoolean unpacking in native/maybe.go and the tuple marker in
lower/expression.go; packed-boolean and nullable-tuple controls then pass in
1.682s. The old primitive-array producer-gap expectation is replaced by actual
storage controls: four valid scalar-array reads match Node, and wrong payload
and missing-index mutants stop with exit 70 at the read (2.658s). These changes
restore storage support without changing callable dispatch or its refusals.

After those restorations, the scoped compiler gate passes with -failfast:
IR 0.016s, lower 7.156s, native 38.225s, JavaScript 1.065s; vet passes.

Final requested oracle selection is green (1066.495s):
`ADAMIC_GATE_UNCACHED=0 go test ./internal/oracle -run
'Map|Nullish|NullableSelection|CheckedView' -v -count=1 -timeout 45m`.
Native executions were all fresh: 0 result-cache hits / 1,375 misses. Node had
88 hits / 4,449 misses; no probe result was cached. The fresh preceding full
selection exposed only the two restored storage dependencies and the outdated
primitive-array gap expectation; their targeted controls and mutants pass.
The definitive complete green log is evidence/merge-ffe428ab-final-green.log.gz.
This finished merge unit is committed and pushed once under the standing rule;
no additional worker branch was merged.


October 8: recursive readonly nominal producer witnesses

Readonly structural object cycles now admit finite nominal data-class leaves,
with null, undefined, and combined-nullish links. Eight source controls match
Node in native and JavaScript, with zero leaks and recorded allocation rows.
Producer traversal uses owned snapshots and an iterative worklist, keyed by
both object address and schema. A valid self-cycle terminates and releases its
references after the test restores the acyclic payload.

Four real payload mutants stop with exit 70: a nested class lookalike before
Map construction, the same corruption after construction at an escaped class
read, an undeclared null root, and one shared object reached under two different
class-bearing schemas. Removing only producer/read nominal checks in a scratch
worktree lets all four payloads run on in native, released native, and JavaScript.
Changing visitation to address-only independently lets the shared-schema mutant
run on in all three modes. Counterfactual logs record successful execution of
corrupted payloads; the address-only red log is the original test catching that
missing check. None of these counterfactual changes are on this branch.

The final scoped oracle passes in 10.643s. Compiler packages pass: IR 0.012s,
lower 5.074s, native 103.972s, JavaScript 0.876s; scoped vet passes. The inherited
phantom failures remain excluded with -skip Phantom; no full repository gate
is claimed. entry-live-mutation remains refused at 5:69 for its slotless Map
callback value. Broader callable boundaries remain lane 5 dependencies.

Recursive arrays, mutable aggregate cycles and recursive class declarations
retain their existing refusals and need separate producer proofs. Required
primitive brands remain refused in internal/lower/phantom_brands.go,
(*lowering).phantomRefusal: "a primitive brand member brand whose type is not
void". The conservative inventory remains 2,018 pairs / 9,101 reads; exact
production reaching-view discharge is unmeasured. Evidence is in
recursive-nominal.json and the corresponding compressed logs.


October 8: recursive readonly reference-array paths

The visited witness now handles object/array back edges with nonnullable
structural entry elements. Array descriptors are interned before their element
is lowered, and incomplete producer schemas remain refused. Native checks
physical reference storage before reading a slot, retains each work snapshot,
and skips sparse holes. JavaScript traverses present indices under the same
schema obligations. Three additional controls match Node and record allocation
rows with allocations equal to frees. A real object-array back edge terminates
and passes the leak check after cleanup.

Nested class producer and escaped-read lookalikes, plus a scalar array injected
into reference-array storage, stop with exit 70 in native, released native and
JavaScript. All three corrupted payloads run on when nominal checks alone are
omitted in the scratch counterfactual; no bypass is committed. The complete
recursive matrix and counts pass in 11.543s; the additional physical-storage
mutant passes in 0.536s. Existing nominal tests pass in 27.941s and retain the
entry-live-mutation refusal. Compiler packages pass: IR 0.013s, lower 6.443s,
native 39.553s, JavaScript 1.212s; scoped vet passes. No full gate is claimed.

Mutable recursive aggregates and recursive class declarations still need their
own producer/storage proofs. Callable families remain lane 5 dependencies,
optional/rest tuples remain lane 4c, and required primitive brands keep the
existing language refusal. The 2,018-pair / 9,101-read inventory and unmeasured
production discharge remain unchanged. Evidence: recursive-array.json and logs.


October 8: mutable recursive object schemas

Nonoptional mutable structural object edges now use the visited producer
witness. Four controls match Node, including direct and helper class writes
before a later view activates. Four actual lookalike payload mutants stop at
the producer, escaped read, direct write or helper write in all three execution
modes. Removing nominal checks alone lets all four payloads run on; no bypass
is committed. Allocations equal frees in the four new counted rows.

The unit oracle passes in 10.162s, nominal regression in 37.089s, and the exact
remaining source-write/callback refusals in 0.704s. Compiler packages pass:
IR 0.018s, lower 11.686s, native 54.237s, JavaScript 1.779s; scoped vet passes.
The inherited phantom failures remain excluded; no full gate is claimed.

Nullable aggregate link writes stay refused by internal/lower/class.go,
(*lowering).setProperty: "storing Entry | null in a field", pinned at
entry-nominal-recursive-mutable-link-write.a:9:1. Mutable recursive arrays and
recursive data-class schemas remain on the table. Callable boundaries and
optional/rest tuples retain their lane ownership; required primitive brands
retain their language refusal. The inventory remains 2,018 pairs / 9,101 reads,
with exact production reaching-view discharge unmeasured. Evidence is in
recursive-mutable.json and its compressed logs.


October 8: recursive class storage prerequisites

Four recursive readonly-array class probes match stock Node but remain refused
at node.value: 6:33 for control/read/write, 6:52 for the producer probe. The
refusal gate passes in 0.632s. A candidate interned class descriptors before
lowering back edges, but the readonly array field initializer stopped at
<write>.children because its original array element certificate was unavailable.
A candidate nullable class-field layout also exposed ASan heap-use-after-free
in adamic_map_set after graph adoption and producer traversal. Those candidate
changes were restored, with patches kept under /tmp for diagnosis; neither
class admission nor a nullable class layout is enabled by this unit.

The prerequisites are certified class array initialization and nullable class
layout/graph-region ownership. Recursive class reads remain named Map key/value
certificate refusals until those source-storage proofs are available. No other
worker branch was merged. This unit records the blockers and pins four valid
Node controls to the existing read refusal; it adds no new runtime check.
Evidence is recursive-class-blocker.json, raw candidate red logs, decoded ASan
stderr, and the green refusal log. Inventory remains 2,018 pairs / 9,101 reads,
with exact production discharge unmeasured.


October 8: nullable nominal aggregate source-link writes

Complete scalar/object graphs with nominal class witnesses now preserve null,
undefined and combined-nullish links across direct and helper source writes.
Six Node controls (including fresh links) pass all three execution modes and
leak checks. Three forged-link payloads exit 70 at the direct write, helper
write or independent escaped class read; with nominal checks removed the same
payloads run on in all three modes. The 21 recursive count rows were refreshed.

Graph adoption now rebinds an inline Map entry pointer after moving the Map.
Removing this repair reproduces ASan use-after-free in adamic_map_set. Checked
reference writes use the existing graph-aware hold/drop helpers; removing this
change reproduces use-after-free in adamic_release after an undefined link
write. This fixes the Map relocation prerequisite observed by the earlier
recursive class candidate; recursive class admission remains refused.

The unit gate passes in 9.128s and the nominal/optional/refusal regression in
41.879s. The callback refusal at entry-live-mutation.a:5:69 is preserved. Scoped
compiler checks and vet are recorded separately. Five standalone graph harness
failures reproduce on the baseline: four have stale layout byte expectations,
and one uses the old closure signature. Phantom failures are also inherited.
Evidence: evidence/nullable-aggregate-link.json and its logs. Inventory remains
2,018 pairs / 9,101 reads, with production reaching-view discharge unmeasured.
Required primitive brands retain internal/lower/phantom_brands.go:133,
(*lowering).phantomRefusal: a primitive brand member brand whose type is not void.
