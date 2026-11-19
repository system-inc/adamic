# Expansive generics cache (#bmqm0k1)

One edit, run on a house box or a sealed one, then put back.

## The cache ignored

In `expansiveGenerics` (expansive_generics_cache.go), walk whether or not the lowering holds a cache, so
every lowering of a latent census run walks the program again, as before #bmqm0k1:

| Old text | New text |
|---|---|
| `	if l.expansiveGenericsHeld == nil {` | `	if true {` |

Run:

```
go test ./internal/lower -run 'TestExpansiveGenericsIsWalkedOncePerCheckerAndProgram' -count=1 -v -timeout 60s
```

The red is the walk count, in both rows: `no stop: two lowerings sharing a cache walked 2 times, not once`,
then `refused: two lowerings sharing a cache walked 2 times, not once`. Each lowering still gets its own
walk's stop, so the rows' stop checks stay green; only the count says the cache was skipped.

`TestExpansiveGenericsCompareCatchesAWrongAnswer` goes red too, by another line: no lowering stores a stop, so
it stops at `the first lowering through the cache stored no stop for its checker and program`.
