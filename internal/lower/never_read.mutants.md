# The earliest narrowing's mutants (#p6vawef)

Each is one edit, applied alone, then the named test run with -v.

## 1. The between started at the last narrowing

In internal/lower/never_read.go, old:

```go
	case l.neverPlaceWrittenBetween(node, making.earliest.from, making.earliest.observed):
```

New:

```go
	case l.neverPlaceWrittenBetween(node, making.madeNever.from, making.madeNever.observed):
```

Expected: TestNeverReadAcrossAWriteIsNotYet goes red at nv1, nv3 and nv5, each lowering with no error: the
call between the checks, or after the first, runs before the check that made the read never, so nothing
sits between that check and the read, and the read traps where Node prints stray. The let rows stay not
yet, at narrowedNeverReadGap.

## 2. The test for what can write the place dropped

In internal/lower/never_read.go, old:

```go
	case l.neverPlaceWrittenBetween(node, making.earliest.from, making.earliest.observed):
```

New:

```go
	case false && l.neverPlaceWrittenBetween(node, making.earliest.from, making.earliest.observed):
```

Expected: TestNeverReadAcrossAWriteIsNotYet goes red at nv1, nv3 and nv5, each lowering to the trap.

## 3. narrowedNeverReadGap unregistered, the earliest narrowing alone

In internal/lower/narrowed_nullish.go, old:

```go
	refusalRules.register(3340, "narrowedNeverReadGap", (*lowering).narrowedNeverReadGap)
```

New:

```go
	_ = (*lowering).narrowedNeverReadGap
```

Expected: TestNeverReadAcrossAWriteIsNotYet stays green on every row, nv2 and nv6 to nv9 stopping at the
same read as never_read.go's not yet. TestNeverReadOfACallWrittenLetIsNotYet goes red on its words alone,
and TestRefusalRulesRunInOrder misses 3340.

Measured beside it (#p6vawef, not a pin): a numeric enum let a closure writes between two checks, nv2's road
with E { A, B } for N, lowers again, to enum_never.go's trap, since never_read.go leaves an enum's never read
to it. Run with B, the call puts A back, and both backends stop at exit 70 ("unreachable value 0") where
Node prints seen 0. So narrowedNeverReadGap still holds a road the earliest narrowing doesn't: the enum
trap isn't judged by what runs between. The same shape through a property (holder.e) stops that way on
main with the rule in.

## 4. enum_never's read judged by nothing (#afs4xvb)

In internal/lower/expression.go and internal/lower/object.go, each `if err := l.enumNeverReadGap(...); err
!= nil {` becomes `if err := error(nil); err != nil {`.

Expected: TestEnumNeverAcrossAWriteIsNotYet goes red at path and deep, .a and .ts, each lowering to
enum_never's trap, which exits 70 where Node prints seen 0 (J-xz537sq). The let row stays not yet, at
narrowedNeverReadGap, and both switch rows at their scrutinee.

## 5. An element read always counted as code

In internal/lower/narrowed_nullish.go, elementReadRunsNothing returns false.

Expected: TestEnumNeverWithNothingBetweenTraps goes red at index, .a and .ts: the later checks' values[0]
reads count as code between the earliest narrowing and the read, and a name a number spells stops then.

## 6. The switch's scrutinee judged by nothing

In internal/lower/object.go, `if err := l.enumSwitchScrutineeGap(statement.Expression); err != nil {`
becomes `if err := error(nil); err != nil {`.

Expected: TestEnumNeverAcrossAWriteIsNotYet goes red at switch_implicit and switch_default, .a and .ts.
switch_implicit lowers to the implicit default's trap, which exits 70 where Node prints after
(J-xz537sq); switch_default stops later, at its default's own read of holder.e, never at the scrutinee.
