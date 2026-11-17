# An overload's Extract over intersection members, not yet in .ts: mutants (#0kkpw2e set 2)

Each mutant is one exact edit on this head, applied alone, run, then restored from `HEAD`. The pins are
internal/lower/overload_kind_extract_intersection_test.go: `TestOverloadKindExtractIntersectionIsNotYet` (x2,
tsc's transformNamedEvaluation shape, in .ts), `TestOverloadKindExtractIntersectionStaysRefusedInA` (the same
program as .a) and `TestOverloadKindExtractIntersectionControlStaysRefused` (a narrower union with an
intersection member that isn't an Extract). Set 1's pins in overload_kind_extract_test.go run beside them.

## 1. The .ts test dropped: the gap is named in .a too

File: internal/lower/census_small.go

Old:

```go
		if l.checkedAssertionSource(implementation) && l.checkedAssertionSource(overload) && l.overloadKindExtractIntersection(promised, produced) {
```

New:

```go
		if l.overloadKindExtractIntersection(promised, produced) {
```

Expected: `TestOverloadKindExtractIntersectionStaysRefusedInA` reds: the .a program stops as the NotYet, not as
the covariance refusal, so a .a file would hold a gap. The .ts row stays green (it was taken there already),
the control stays green (its promise isn't an Extract, so the branch never reads it), and so do set 1's pins.
