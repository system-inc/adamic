# JSON and config family

Eleven explicit any tokens and nine direct any blockers are replaced. The AST
converter produces recursive JSON scalars/containers plus undefined for recovery
and check-only mode. Object-root recovery is completed only after the existing
object selection or empty-root branch. Non-null completion follows the original
returnValue guard. Equality/string helpers and the option serializer retain caller T.

Source-map JSON.parse has the builtin no-reviver recursive domain. JSON arrays
have none of the named map object fields. The original validator narrows the union;
no new runtime guard or assertion of a specific parsed shape is introduced.

Stock tsc: zero diagnostics. Both entire files emit byte-identical JavaScript.
Actual converter and validator bodies run on Node with malformed roots, bad values,
check-only mode and invalid maps; observations match. Changing the actual map
version check fails the valid/invalid shape assertions.

Census: direct any 35 -> 26; restoring the actual convertToJson return any yields
27, with exactly that extra any-return site. Refused 5151 -> 5159 and NotYet
1468 -> 1470: truthful unions/generics expose additional compiler lessons.
The mutant substitutes an any-return lesson for a JsonConfigValue-return lesson,
so its total is unchanged. Full reason deltas and raw observations are retained.
This is not a claim of native compilation.

Public convertToObject and raw configuration input contracts remain decision
residue: they accept arbitrary JavaScript, and narrowing public signatures adds
API differences outside the lane's existing 222 sanctions. tryParseJson also
feeds readers falsely declared object even for primitive JSON. The final residue
ledger has exact locations, runtime counterexamples and minimal programs.

Commands: stock tsc --noEmit; family-proof.cjs; json-proof.cjs; latent census
and census-report.py on control and real restored-return mutant. Outputs go to
logs. The final composed lane/oracle verifies behavior before the single push.
