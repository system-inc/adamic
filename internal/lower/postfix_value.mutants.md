# Mutant for the postfix update value (#f3n0gyp)

## M1: the new value

In `internal/lower/postfix_value.go`, `updateValue` returns the local after the write for a postfix update too (the held old value never replaces it), which is prefix's value.

Expected: internal/oracle's `TestPostfixUpdateValueOnANumericLocal` goes red on every row (Node prints the old value), and `TestPostfixUpdateValueOfAFieldStops` stays green.
