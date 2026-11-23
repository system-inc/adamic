# Mutant for the postfix update value (#f3n0gyp)

## M1: the new value

In `internal/lower/postfix_value.go`, the helper returns the local after the write (`l.expression(operand)` read again) instead of the held old value, which is prefix's value.

Expected: internal/oracle's `TestPostfixUpdateValueOnANumericLocal` goes red on every row (Node prints the old value), and `TestPostfixUpdateValueOfAFieldStops` stays green.
