# Mutants of splice with a spread among its items (#p1qw4kf)

The test is internal/oracle/splice_spread_test.go, `TestSpliceSpread`. Each mutant is one exact line of
`internal/lower/array_call_arguments.go` or `internal/lower/object.go` replaced, run alone with the unmutated head
beside it, then put back byte for byte.

## M1: the insertion index taken from the raw start

```go
at := ir.Binary{Operator: ir.Add, Left: b.read(index), Right: b.read(position)}
```

becomes

```go
at := ir.Binary{Operator: ir.Add, Left: start, Right: b.read(position)}
```

Red: `TestSpliceSpread/bounds.ts`. values.splice(-1, 0, ...[8, 9]) inserts 8 at -1, which the runtime resolves to
index 2, then 9 at -1 + 1, which is 0, where Node prints `[0,1,2,8,9,3]`. A NaN or a negative infinite start
inserts every item at 0, reversed.

## M2: the removal done after the insertion

```go
b.body = append(b.body, removal, insertion)
```

becomes

```go
b.body = append(b.body, insertion, removal)
```

Red: `TestSpliceSpread/bounds.ts`, `empty_and_count.ts`, `mixed.ts` and `self.ts`, each with a count over 0: the
removal at the resolved index takes the items just inserted, not what was there.

## M3: the spread expanded after the removal

```go
b := l.libraryArrayBuilder([]ir.Expression{array, arguments[0], arguments[1], packed})
```

becomes, for a splice whose only item is one spread, the spread's source handed to the helper as it is, so its
elements are read after the removal and while the insertion grows it:

```go
unpacked := ir.Expression(packed)
if len(packed.Elements) == 1 && packed.Spread[0] {
	unpacked = packed.Elements[0]
}
b := l.libraryArrayBuilder([]ir.Expression{array, arguments[0], arguments[1], unpacked})
```

Red: `TestSpliceSpread/self.ts`. values.splice(1, 0, ...values) on `[1,2,3]` reads values[1] after 1 went in at 1,
and prints `[1,1,1,1,2,3]` where Node prints `[1,1,2,3,2,3]`. The programs whose spread source the splice doesn't
touch stay green.

## M4: the spread-aware arm dropped

```go
if name == "splice" && arraySpliceSpreads(node.AsCallExpression().Arguments.Nodes) {
```

becomes

```go
if false && name == "splice" && arraySpliceSpreads(node.AsCallExpression().Arguments.Nodes) {
```

Red: `TestSpliceSpread/site_shape.ts`, and every other program with a spread, as a stop: the old per-item lowering
reaches the spread and stops with NotYet "a SpreadElement". `stop_representation.ts` goes red too, its stop no longer
in push's words.
