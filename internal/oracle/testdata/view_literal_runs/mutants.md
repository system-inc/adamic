# Mutants of the literal run test (#zpzm9ta)

Each is one edit to `internal/native/view_literal_runs.go`, in `viewLiteralTests`, made alone, with the test it must
turn red. Kind is the runs 0..7 and 10..13, so every fixture here reads through two run tests.

## M1: the trunc test dropped

```go
tests = append(tests, fmt.Sprintf("(%s >= %s && %s <= %s && %s == trunc(%s))", value, cNumber(run[0].value), value, cNumber(run[len(run)-1].value), value, value))
```

becomes

```go
tests = append(tests, fmt.Sprintf("(%s >= %s && %s <= %s)", value, cNumber(run[0].value), value, cNumber(run[len(run)-1].value)))
```

Red: `go test ./internal/oracle -run 'TestViewLiteralRunsStopAtTheRead/stop_fraction.ts'`. 0.5 lies inside the first
run's bounds, so natively and sanitized the read lets it through and the program runs on to print `node:fs 0.5` with
exit 0, where the view's literal failure and exit 70 are wanted. `viewLiteralRunsEmitted` stays green, since it counts
the upper bounds, so the run on is what turns it red. The JavaScript backend stays green: its check is
`allowed.includes`, which this edit doesn't touch.

## M2: `>=` becomes `>`

```go
"(%s >= %s && %s <= %s && %s == trunc(%s))"
```

becomes

```go
"(%s > %s && %s <= %s && %s == trunc(%s))"
```

Red: `go test ./internal/oracle -run 'TestViewLiteralRunsHonestReadsAgreeWithNode'`. 0 is the first run's lowest
member, and honest.ts's first checked read holds 0, so natively and sanitized it stops there with
`field read failed: imported.argument.literal expected Kind, found number 0` and exit 70, where Node prints `node:fs 0`
and goes on. The JavaScript backend stays green, as for M1. The unit test `TestViewLiteralRunsAreRangeTests` in
internal/native goes red with it, its expected line spelling `>=`.
