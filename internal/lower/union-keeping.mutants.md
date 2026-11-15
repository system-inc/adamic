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

## First target member only

Change the target loop to `for _, target := range toMembers[:1]`.

```
go test ./internal/lower -run '^TestUnionKeepingPairs$/^later_weak_member' -v -count=1 -timeout 10m
```

The command exits 1. `TestUnionKeepingPairs/later_weak_member.a` and
`TestUnionKeepingPairs/later_weak_member.ts` both fail with
`got <nil>, want retention NotYet at assignment`. The types are `Strong` and
`{ readonly kind: 'w'; readonly w: number } | { readonly kind: 'v'; readonly v: Weak<Box> }`.
The first target has no `v`; the second target changes `v` from strong to Weak.
The file was restored. Output: `/workspace/scratch/pmfdg6q/storage/first-target-mutant.log`.

## Parameter direction

Undo the argument direction in both parameter calls to `sameCallable`, passing the
function parameter before the view parameter again.

```
go test ./internal/lower -run '^TestUnionKeepingParameterDirections$' -v -count=1 -timeout 2m
```

The command exits 0: this mutant survives the remaining pins. Both directions of
the direct Weak-slot function view still stop. The direct array-like function
view also stops when the flip is undone, since storage mode compares index values
in both directions. The production flip is retained. No existing fixture was
changed to manufacture a failure.

The file was restored. Output: `/workspace/scratch/pmfdg6q/storage/direction-mutant.log`.

## Descend into signatures again

Replace weak-only dispatch to `sameKeepingStorage` with the ordinary structural
walk, keeping `typeNesting.enter` and its deferred leaves on both sides. This
restores descent into method and callable-field signatures.

```
go test ./internal/lower -run '^TestUnionKeepingPairsAgree$/^promise_union_view' -v -count=1 -timeout 2m
```

The command exits 1. `TestUnionKeepingPairsAgree/promise_union_view.a` and
`TestUnionKeepingPairsAgree/promise_union_view.ts` fail with:

```
stage 0 can't lower a Promise<string> seen as a string | PromiseLike<string> (one keeps something weakly that the other keeps strongly) yet
```

The nesting bound fails closed instead of overflowing. The file was restored.
Output: `/workspace/scratch/pmfdg6q/storage/signatures-mutant.log`.
