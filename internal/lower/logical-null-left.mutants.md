# Mutant for a null left of && kept in a union result (#ayvxxpc)

The pins are `TestLogicalNullLefts` (internal/oracle/logical_null_left_test.go) over
internal/oracle/testdata/logical_null_left/, each lowering and agreeing with Node in its .a and .ts twin.

## The null arm off

internal/lower/logical_null_left.go, `logicalNullLeft`. Old:

```go
	held := left.Type()
	if !held.IsReference() || held == ir.String || held == ir.Union || held == ir.Weak {
```

New:

```go
	held := left.Type()
	if held != 0 || !held.IsReference() || held == ir.String || held == ir.Union || held == ir.Weak {
```

Red: a1, a5 and a9 panic natively ("union member where the checker narrowed it away"), and a2 and a3 print
the undefined reading (false, other) where Node prints true and is null, in .a and .ts.
