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

## Enum display reflection

All 23 sites applied next. formatEnum, getEnumMembers and the cache key use
Record<string, string | number>, matching forward numeric members and reverse
string names. Namespace casts project only the actual enum field being read;
they do not falsely claim every namespace export is an enum or directly read
const enums as a value in typespace. FileIncludeKind uses its numeric reverse
map. All actual compiler call sites supply these numeric enums; test-runner
call sites were inspected and likewise supply ScriptTarget/JSDocParsingMode.
The existing typeof value === number filter is preserved. No unknown, inserted
narrowing, runtime statement, callback implementation or host-input edit occurs.

Census 171 -> 148, zero new consumer errors, all 10 emitted JavaScript files
identical. The only changed declarations are debug.d.ts and
typescript.internal.d.ts, with precisely formatEnum's enumObject parameter
changed. Public API reference unchanged from the brand class. The first Map
AST-parent assumption was rejected before writes; parsing showed NewExpression,
which is now required by the plan. Type-only positions are checked by the AST
and the independent source reconstruction.

Full combined-tree oracle: 106,367 passing, zero failing/pending, empty
baseline.diff, 225.780 seconds. Artifact/census/source/API/idempotence/owner
mutants are caught; an enumObject Boolean-map mutant produces stock TS2345.
The real API reference mutant causes one failing test and one baseline diff,
then exact restoration and zero remaining differences. The earlier brand
audit and consumer limitations still apply. See evidence/enum-display.

## Diagnostic substitution arguments

All six sites applied: scanner/parser arg0 is optional string | number;
formatter and resolution trace arguments use DiagnosticArguments[number][].
Actual callers establish the shared domain as string | number | boolean |
readonly string[] | SourceFile | undefined. The existing formatting code
string-coerces these values and checks defined placeholders; no check was
inserted. Two owning generic declarations in program.ts now constrain their
SourceFileOrString parameter to SourceFile | string, matching both callers.
The alias and type import are also declaration-only edits.

The initial string | number rule exposed 16 TS2345 errors. Every diagnostic,
including exact text, is in evidence/diagnostic/exploratory-consumer-errors.json.
The final owner declarations produce zero new consumer errors. No consumer
statement was changed. All ten emitted JavaScript files are byte-identical;
714 declarations are compared, six differ exactly by the reviewed projection.
The API reference differs only by ErrorCallback.arg0's any -> string | number,
proved mechanically. Census 148 -> 142, adapter second pass changes zero bytes.

Full oracle with 10+30+40 and without 20: 106,367 passing, zero failing/pending,
baseline.diff zero bytes, 210.184 seconds. Seven artifact/source/census/API/
idempotence/owner mutants fail their checks. A real diagnostic argument-domain
mutation is caught by stock TS2345. An unrelated actual API-reference edit
fails one test and produces one baseline difference; its reference is restored
exactly, with zero remaining differences. See evidence/diagnostic.

Stopping after these three classes: brands (36), enum display (23), diagnostic
arguments (6), each independently built, proved, committed and pushed. Together
with the previous three owners, 68 of 210 tokens are removed; 142 remain.
JSON/config (38) and process/stream extensions (3) retain their checked-boundary
rules above and are held for the separate runtime-behavior decision. Other
classes retain proposals and exact locations in CLASSIFICATION.md; their
contracts have not yet received the owner, consumer and oracle proof. In
particular, timers need handle/argument generics propagated across owners;
constructor/copy/callback classes need correlation and variance work. This is
unfinished proof work, not a claim that those type-only edits are impossible.
