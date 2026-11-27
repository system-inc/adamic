# Mutants of the ?. read to a scalar (#0d56xpd, rule (d))

Each mutant replaces exactly one line, checked to match once, runs `go test -run
'TestOptionalScalar' ./internal/lower/ ./internal/oracle/`, and is put back
from a copy of the head file. Each is red beside the same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `fixedFieldRead` (object.go): the read made `Optional: false`, so the receiver isn't tested | TestOptionalScalarRead's u-rows, on the placeholder parent |
| M2 | `optionalScalarRead` (optional_scalar.go): the check dropped, the maybe returned where the number goes | TestOptionalScalarRead/s1_*, TestOptionalScalarReadLowers/s1_* |
| M3 | `placeholderAllowsUnset` (placeholder.go): the `?.` receiver not counted as a presence test | TestOptionalScalarRead's u-rows, in the placeholder's words |
| M4 | `placeholderChainObject` (placeholder.go): the `?.` test dropped, so a plain read through the placeholder counts too | TestOptionalScalarRead/s2_plain.ts |
| M5 | `optionalScalarRead`: the bitwise operand's conversion dropped, so it takes the check | TestOptionalScalarRead/u5_bitwise.ts |
| M6 | `nullishNumericSwitch` (switch_nullish_numeric.go): a number pair the checker calls a number not taken as nullish | TestOptionalScalarRead/u1_switch_plain.ts, TestOptionalScalarReadLowers/u1_switch_plain |

accessKind's own shape, `const { parent } = node; switch (parent?.kind)` (u1_switch.ts), waits on Lower: #jjf9ppe checks
a destructured binding of a field that may hold its placeholder where it's bound, so on the placeholder parent it
stops there, in #jjf9ppe's words, before the `?.` is reached, where Node prints "other". The plain form, `const parent
= node.parent` (u1_switch_plain.ts), is already a placeholder slot and agrees with Node. If Lower rules that an
unannotated destructured const can be a placeholder slot like the plain one, her head moves u1_switch to agree with
Node, and accessKind clears then.
