# Type-only class applications

Brands: 36 owner types changed from any to undefined, preserving marker keys
and optionality. Required markers remain required; the optional qualified-name
marker remains optional. Branded primitives and AST construction use nominal
assertions or staged constructors already present in upstream, rather than
materializing marker properties. Across all 710 source files, all 53 occurrences
of the reviewed marker names are erased declarations and none is a runtime read,
write, or field initializer. Service marker declarations remain unchanged.

The combined comparison tree includes adaptations 10 and 30, excludes 20, and
starts with the previously proved three adaptation-40 owner edits. Adaptation
30's four U-zero expressions are therefore already in the before JavaScript;
this unit's byte-identical claim is against that combined starting tree, not
against upstream without adaptation 30. See evidence/brands/composition.json.

Census: original 210, previous 207, after brands 171. Full builds and stock
6.0.3 compiler checks pass with zero new consumer errors. All 10 emitted
JavaScript files match byte for byte. All 714 declarations are compared; eight
change only through the reviewed marker owner projection. The public API
reference changes by exactly 27 marker annotations. The adapter applies that
limited exception mechanically, not by copying generated baseline output.
Every other baseline is held unchanged. Adapter second pass changes zero bytes.

The new proof is class-proof.cjs, with scratch state-before.json carrying the
complete before sources and declaration texts. Committed before/after JSONs
retain the complete census and JavaScript hashes; source.patch and
public-api.patch give the exact edits. The earlier verify.cjs and
classify-check.cjs describe historical standalone snapshots at commits 96ed162
and dca3240; they are not the verifier for the combined current tree.

Reproduce each class sequentially on an already built tree:

```sh
node class-proof.cjs before "$tree" "$proof" brands > before.log 2>&1
node adapt.cjs "$tree" > adapt.log 2>&1
npm run build --prefix "$tree" > build.log 2>&1
node class-proof.cjs api "$tree" "$proof" brands > api.log 2>&1
node class-proof.cjs after "$tree" "$proof" brands > after.log 2>&1
stage3/oracle/run.sh "$tree" "$oracle" > oracle.log 2>&1
```

Hold JSON/config (38): unknown at genuinely untyped input boundaries, checked
object/field narrowing there, and validated recursive value types internally.
Hold optional process/stream extensions (3): unknown plus checked extension
shape, including method callability. Their new checks would change bad-input
behavior. No checks were inserted, and no consumer was fixed. All remaining
class rules and exact original locations are in CLASSIFICATION.md and
evidence/classification.json; those counts describe the pre-brand snapshot.

The other classes remain proposals until independently proved. Native Adamic
acceptance, the full Go gate, and soundness of upstream's remaining assertions
are not claimed.

Full stage 3 oracle: 106,367 passing, zero failing/pending, empty baseline.diff
(0 bytes), 216.134 seconds. A focused unittest Public APIs run with an unrelated
type appended to the real reference failed one API test and produced one
baseline difference. The reference was restored byte for byte and every local
baseline again matched its reference. The marker-read audit mutant, seven
source/artifact/idempotence/owner mutants, stock checker brand mismatch mutant
(TS2322), and API baseline mutant were all caught. Replay of the full adapter
on 10+30 sources with the three previous helper edits undone reproduced every
compiler source and the exact API exception; its second pass removed zero.
