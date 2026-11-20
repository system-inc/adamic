# Mutants for a fixed tuple's length written natively (#87dwc5c)

One edit to class.go, run on a house box, then put back. Measured at c8752be9, Loom job J-k3warys.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| setProperty's tuple length arm removed | `if target.Name().Text() == "length" && checker.IsTupleType(` | `if false && target.Name().Text() == "length" && checker.IsTupleType(` | internal/oracle's TestTupleLengthWrite, all 12 rows. same, parenthesized and the conformance program tuple_length_control_fix, .a and .ts, lower to a SetProperty of "length" and exit 70 on release and sanitized native with `adamic: panic: compiler bug: a field the checker proved is there is missing`, where Node and the JavaScript backend print 2. named_value, optional_element and through_field, .a and .ts, lower where they want tupleLengthWrite's NotYet |
