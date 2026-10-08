Tagged: 0 certified conforms, 0 certified conforms-if, 0 certified never, 1,758 undecidable.
Untagged: 0 certified conforms, 0 certified conforms-if, 0 certified never, 1,178 undecidable.
Evidence: all 2,936 ledger locations/text match; stock TypeScript 6.0.3 has zero diagnostics.
Checks/depth: unknown at every site, not zero; no complete typed allocation/initialization certificate is available.
Status: proposed design audit only; lane staffing held and no compiler behavior changed.

# Shape conformance audit, October 7, 2026

These counts mean **not certified by the available evidence**, not that shape
conformance is impossible or that all these casts must fail. The table does not
replace an interface type with a supposed actual runtime shape. No whole-program
allocation-set, typed-shape, construction/readiness or alias-effect analysis exists
for this corpus in the current branch. Implementing that analysis is substantial
compiler work and was not completed inside the 30-minute audit window.

The pinned inputs are TypeScript 6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8 and the unchanged assertions ledger
SHA256 15739c81ad33071014c88e0b4514b08c7562cb3191d0bd20e7df321fc2bc1cab.
The 1,758 tagged sites are the existing other-bucket refinement; the 1,178 untagged
sites are the assertions ledger's original structural-downcast bucket. No cast
category was remeasured. The audit resolves constant-binding initializers up to
12 steps, distinguishes parameters, mutable bindings, properties, factories and
opaque declarations, and accepts no prior assertion as allocation evidence.
There are no complete direct literal candidates in these selected sites.

[Per-site table](../stage3/interface-downcasts/shape-conformance-sites.json) includes
location, exact expression, source/target types, origin location, outcome and reason,
unknown shape/check/depth entries, required brand/any fields and declaration-only
missing/wider field lists. Those lists are comparison hints, not conformance verdicts:
a base interface can omit fields actually present on a subtype object. An unknown
check count or recursion depth is represented as null, never as zero.

| Unresolved evidence | Tagged | Untagged |
| --- | ---: | ---: |
| Parameter allocation domain | 1,297 | 830 |
| Property or element allocation domain | 189 | 67 |
| Factory/call construction and effects summary | 61 | 153 |
| Opaque/declared call boundary certificate | 26 | 32 |
| Mutable binding assignments | 41 | 9 |
| Destructuring provenance | 85 | 5 |
| Binding without allocation certificate | 54 | 18 |
| Prior assertion | 2 | 16 |
| Other expression allocation analysis | 2 | 24 |
| Conditional allocation join | 1 | 22 |
| Constructor readiness/typed-shape certificate | 0 | 2 |
| Total | 1,758 | 1,178 |

A parameter is not inherently shape-less. A complete global universe of typed
allocation shapes can support a table without precise per-cast points-to analysis;
flow analysis can then reduce its size and certify which outcomes a site reaches.
The counts above expose missing certificates, not a requirement to invent a new
runtime view for every parameter. Sites can have several runtime shape outcomes,
so an eventual audit must retain the shape-specific cells as well as any site summary.

## What the current representation proves

`internal/native/runtime/adamic.h:130` defines a physical layout containing names,
reference bits and method dispatch. `internal/native/emit_objects.go:179` interns
layouts by those names/reference bits and method function identities. It does not
intern full declared field types. Two literals with `ready: 0` and `ready: true`
share exactly the same adamic_shape_0; their per-object representation bytes are
1 and 2. The generated C observation is asserted in shape-layout-collision.log.
Both ordinary native and Node source executions print `0,true` and agree.
This is evidence that current shape identity cannot itself certify a field's type,
not a current compiler miscompile. String/object/array slots can likewise share
reference-layout bits. Typed shape identities or separate immutable schema identities
must be introduced before treating the proposed table as a type proof.

The non-null worker intentionally permits `undefined!` and `null!` assertion
initializers. Their declared type remains T while the shared readiness byte says uninitialized.
The existing default-read-before-set.a therefore has an Identifier-shaped object
and must still fail at the read; default-staged.a fills the field later and succeeds.
The independent native/release/JavaScript assertions were rerun and pass in 1.022s.
A separate scratch witness casts a held Node before filling escapedText, prints
`cast passed`, assigns `okok`, and then reads through Identifier. Native, generated
JavaScript with the normal oracle runtime, and original-source Node all print
`cast passed\nokok\n`. Its source and outputs are preserved in
shape-cast-before-store.log. This directly demonstrates a currently legal staged
cast that an eager payload initialization check would reject.

Tsc's own createBaseIdentifier casts before assigning escapedText, then leaves symbol
uninitialized until a later phase; utilities.Identifier leaves parent uninitialized.
A type-only Conforms cell cannot erase readiness checks. An eager initialization
check at the cast would reject staged construction that the accepted ruling allows.
Retain readiness checks until a dominating-store/construction proof, even where
an immutable typed shape proves the field's eventual type. If Conforms-if tests a
not-yet-filled payload eagerly, that similarly changes staged-builder behavior.

Required fields ending in Brand occur in 1,661 tagged and 96 untagged targets.
Tsc treats many as phantom markers, while constructors omit them. The literal rule
that every required property physically exists needs an explicit treatment of these
markers; stock type declarations do not certify their presence. These are overlapping
contract observations, not proven Never counts for every flowing runtime shape.

## Conditional checks and aliases

Given an initialized, typed shape whose field is A | B, a cast to A needs a current
value test and conversion if representations differ. If an alias can later store B,
keep that field's read check. This applies to reachable nested objects and array
contents as well as direct writes to the root: readonly access through T does not
freeze another writable alias. Recursion needs shape/type-pair memoization for cycles,
and dynamic arrays/dictionaries need element/key membership checks unless their
validated declared contract already establishes the target contract. The table alone
does not prove these effects. No per-site guard count, finite depth, or check erasure
is claimed before the actual shape graph and effects are known.

## Host and dynamic-key boundaries

[Boundary inventory](../stage3/interface-downcasts/shape-conformance-boundaries.json)
observes three direct JSON.parse calls, nine require calls, four Object.create calls
and one Object.assign call in the compiler source. This is a syntactic inventory,
not a transitive count of every host-produced value. The JSON/require expressions
all have stock-checker type any. For example, sourcemap.ts:424 parses arbitrary text
and validates it in isRawSourceMap; sys.ts:1619 loads a dynamically named module.
Object.assign in utilities.ts:8578 replaces allocator callbacks. CompilerHost,
watching and filesystem callbacks also need boundary contracts even without a
syntactic require at the cast site.

All ordinary native plain objects currently have a physical shape pointer.
That does **not** imply every external value has a statically certified semantic
shape. Arrays and maps have their own heap kinds/element storage, not that plain
object shape table. JSON.parse is currently explicitly Refused in
internal/lower/library_json_stringify.go:14, so there is no admitted native JSON
producer whose validation can be assumed complete. Host JavaScript objects returned
by require or JSON.parse carry no Adamic schema certificate. Future admitted boundary
values require validation/reification or retain checked access; this audit does not
silently trust them.

Native dictionaries in runtime/record.c all use record_shape containing hidden
slot 0, independent of their observable keys. A declared value type can prove the
kind of a **present** entry, but cannot prove that a required target key exists.
Two dictionaries with different key sets share this storage shape. A shape-only
lookup therefore needs key-set evidence or a current membership check, with effect
tracking if deletion/replacement is allowed. Arbitrary JSON keys cannot be covered
by a finite compile-time table of literal key sets without boundary conversion.

## Callable contracts under this proposal

A class method whose implementation/signature has already been independently proven
can participate in typed-shape conformance. So can a record-held closure with a proven
function identity/body contract; it is not necessary to refuse every function-typed
field forever. Shape identity or a runtime function tag alone cannot certify an opaque
function's parameter/result contract. Tsc's nineteen tagged callable targets are
SourceFile; the previous seventy-one untagged callable findings include array built-ins.
Those are construction/intrinsic/body-proof obligations under this proposal, not
ninety automatic permanent refusals. Dynamic host callbacks still need certificates.

## Reproduction and verification

```sh
source /workspace/adamic-tools/env.sh
NODE_PATH=/tmp/interface-downcasts-api/node_modules node stage3/interface-downcasts/shape-conformance-audit.cjs /tmp/interface-downcasts-typescript stage3/fixtures/assertions/ledger.json stage3/interface-downcasts/other-locations.tsv stage3/interface-downcasts > /tmp/interface-shape-conformance-audit.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestDefaultTaggedSourceViews/(default-staged|default-read-before-set|default-wrong-boolean)' -count=1 -v -timeout 30m > /tmp/interface-shape-readiness-witnesses.log 2>&1
```

The audit asserts the pinned commit/API version, zero checker diagnostics, exact
2,936 source matches and zero unmatched ledger entries. Logs and summaries are
committed under stage3/interface-downcasts. No compiler implementation or counts row
changed, no lane plan was pushed, and no full compiler gate or benchmark is claimed
for this audit-only commit. Only codex/interface-downcasts is pushed.
