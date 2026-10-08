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
