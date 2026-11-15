# Union keeping pairs

Remove the member-pair block before `present` in `sameKeepingWalkMode`.

Run:

```
go test ./internal/lower -run '^TestUnionKeepingPairs$/^p[5-8]' -v -count=1 -timeout 10m
```

The command exits 1. These subtests all fail because the assignment lowers again:

| Subtests | Types | Message |
| --- | --- | --- |
| `TestUnionKeepingPairs/p5.a`, `p5.ts` | `Strong` to `Weakly` | `got <nil>, want retention NotYet at assignment` |
| `TestUnionKeepingPairs/p6.a`, `p6.ts` | `{ readonly item: Strong }` to `{ readonly item: Weakly }` | `got <nil>, want retention NotYet at assignment` |
| `TestUnionKeepingPairs/p7.a`, `p7.ts` | `(x: Weakly) => string` to `(x: Strong) => string` | `got <nil>, want retention NotYet at assignment` |
| `TestUnionKeepingPairs/p8.a`, `p8.ts` | `readonly Strong[]` to `readonly Weakly[]` | `got <nil>, want retention NotYet at assignment` |

`Strong` and `Weakly` are the discriminated unions in `union_keeping_test.go`.
The file was restored after the run. The full output is in
`/workspace/scratch/pmfdg6q/mutant.log`.
