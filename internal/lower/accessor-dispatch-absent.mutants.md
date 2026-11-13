# Mutants for the accessor dispatcher's fallback (#4gpbn7w)

One edit to class_accessors.go, run on a house box, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The fallback's Absent dropped | `fallback.Object = object` | `fallback.Object, fallback.Absent = object, false` | internal/oracle's TestAccessorDispatchAbsentAgreesWithNode, the sanitized run of presence.a, presence.ts, undefined_compare.a and undefined_compare.ts: a read past the `{}` literal's allocation, as on main 2f327160 |
| A checked read with no receiver type let through | `if expression.View != "" {` | `if expression.View != "" && l.result.CheckedField(expression.ViewReceiverTypeID, expression.Name) {` | internal/lower's TestAccessorDispatchRefusesAPredicateCheckedKindRead: the predicate's argument.kind read lowers through the dispatcher's plain load |
