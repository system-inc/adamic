# Undefined-only destructuring default

TestDestructureUnionAgrees/null/native, /sanitized and /javascript hold a present
null apart from undefined. In internal/lower/collections.go, replace
`ir.IsUndefined{Value: held}` in destructuredDefault's Conditional with
`ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: held}, Right: ir.IsNull{Value: held}}`.
The null rows print `d` instead of Node's `null`. Run the named tests with
ADAMIC_GATE_UNCACHED=1; a compile failure does not count as killing the mutant.
