# Mutants for the accessor dispatcher's fallback (#4gpbn7w)

One edit to class_accessors.go, run on a house box, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The fallback's Absent dropped | `fallback.Object = object` | `fallback.Object, fallback.Absent = object, false` | internal/oracle's TestAccessorDispatchAbsentAgreesWithNode, the sanitized run of presence.a, presence.ts, undefined_compare.a and undefined_compare.ts: a read past the `{}` literal's allocation, as on main 2f327160 |

The read guard at the dispatcher (`expression.View != "" && l.result.CheckedField(...)`) is main's own and has no new mutant here. Narrowing it to View alone (-4) refused programs that lower and agree (class_features_accessors.a, getters_census_throwing_object.a, forof_writes' r1 and r2, among others, Loom job J-hr71dbd), and the read it was meant for, a predicate's argument.kind, stops earlier at refuseReachedContractGetters, which internal/lower's TestAPredicateCheckedKindReadWithAGetterStopsAtItsContract holds.
