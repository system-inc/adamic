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
