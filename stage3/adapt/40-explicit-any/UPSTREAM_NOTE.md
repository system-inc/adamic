# Upstream note: proven public API types

On October 7, 2026, @system_adamic sanctioned the 28 public API annotation
changes in adaptation 40 for TypeScript 6.0.3: 27 nominal brand fields change
from `any` to `undefined`, and `ErrorCallback.arg0` changes from optional `any`
to optional `string | number`. `arg0` is the first diagnostic substitution
argument and the third callback parameter; `message` and `length` are unchanged.

Brand keys, required/optional status and declaration modifiers are preserved.
The markers are erased nominal declarations, with no runtime field reads,
writes or initializers in the reviewed source. Downstream code can no longer
use the marker type as an unchecked arbitrary value, and a manufactured marker
must satisfy `undefined` where the field is required. Existing assertions and
staged construction remain as upstream wrote them.

Scanner callback implementers now receive `arg0` as `string | number | undefined`.
A downstream invocation passing an unrelated object or boolean is rejected;
callbacks must accept the declared argument domain. This changes the static
contract, with no inserted validation and no change to JavaScript behavior.

The sanction requires a type proven true at every owner, byte-identical emitted
JavaScript, an empty oracle baseline diff, and a mechanically exact API snapshot
exception. The standing rule permits subsequent public API `any` replacements
under that same discipline; it does not sanction unproved replacements or new
runtime checks.

The brand and diagnostic classes each meet those conditions independently.
All 10 emitted JavaScript files were compared byte for byte; all 714 declaration
outputs were compared against their reviewed owner projections. Full stage 3
oracles with adaptations 10, 30 and 40, excluding 20, each passed 106,367 tests
with zero failures and an empty baseline diff. Checker, artifact and actual
API-reference mutants were caught, and mutated references restored exactly.

Evidence: `evidence/brands/public-api.patch` records exactly 27 brand fields;
`evidence/diagnostic/public-api.patch` records exactly `ErrorCallback.arg0`.
The adapter projects those named owners into the reference mechanically rather
than accepting generated output wholesale. Per-class census, diagnostics,
proofs, oracle reports and mutants are in those evidence directories, with
commands and limitations in `PROGRESS.md`. No ledger owned by another worker
is edited; this note is supplied for its upstream entry.
