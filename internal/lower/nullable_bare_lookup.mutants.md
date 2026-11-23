# Mutants for the nullable bare lookup (#92cnfpf)

Each is one edit to nullable_objects.go's nullableBareLookup, run on a rented box, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The guard back to object-held values alone | `if !value.Type().IsReference() \|\| value.Type() == ir.Union {` | `if true { return false }` ahead of it | internal/lower's TestNullableBareLookupStaysNotYet, all 8 rows: a Map get of a missing key, a read past the end, a map of number arrays or null and a map of maps or null, each in .a and .ts (Loom job J-r1qh76b) |
| The guard back to a list of kinds | `if !value.Type().IsReference() \|\| value.Type() == ir.Union {` | `switch value.Type() { case ir.Array, ir.Map, ir.Closure, ir.Promise, ir.Uint8Array, ir.Int32Array, ir.Float64Array, ir.Uint16Array, ir.Generator: default: return false }` | Survives, as no program today can build what it lets through (Loom job J-bsryb11) |

## Why the list mutant survives

The list leaves out three bare reference kinds: String, Record and Weak. None of them reaches the guard bare from a lookup today:

- representation's nullable-union branch boxes a member held as String or Record (expression.go:181, `boxed = … || of == ir.String || … || of == ir.Record`). So a `(string | null)[]` or a `(Record<string, number> | null)[]` read past its end, or its pop, arrives as a Union, which the guard never judges. J-bsryb11 lowered the Record read and the pop under both the rule and the list alike.
- Weak goes by its own road, as Weak<Target> held weakly, never as a nullable union member.

The inverted rule is a guard for the day a kind is added to IsReference, or a nullable String or Record lookup stops being boxed. Either one would make this mutant go red: a bare String or Record lookup of a type holding both null and undefined. A pin for it belongs beside TestNullableBareLookupStaysNotYet then.

A survivor here isn't a gap: the shape it would let through can't be built yet.
