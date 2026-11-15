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
The file was restored. Output: `/workspace/scratch/pmfdg6q/cross-product/first-target-mutant.log`.

## Parameter direction

Undo the argument direction in both parameter calls to `sameCallable`, passing the
function parameter before the view parameter again.

```
go test ./internal/lower -run '^TestUnionKeepingParameterDirections$' -v -count=1 -timeout 10m
```

The command exits 1. These four subtests fail with
`got <nil>, want the callable view to stop`:

- `TestUnionKeepingParameterDirections/weak_array_like_function_strong_array_view.a`
- `TestUnionKeepingParameterDirections/weak_array_like_function_strong_array_view.ts`
- `TestUnionKeepingParameterDirections/union_weak_array_like_function_strong_array_view.a`
- `TestUnionKeepingParameterDirections/union_weak_array_like_function_strong_array_view.ts`

The types are `(x: WeakElements) => string` and `(x: StrongElements) => string`,
directly or as the `read` field of the `Held` and `Viewed` unions.
`WeakElements` holds `ArrayLike<Weak<Box>>`; `StrongElements` holds `readonly Box[]`.
These check the direction in ordinary and weak-only traversal. The original
Weak-slot function pins in both directions also stay; the reverse casts stop.
The file was restored. Output: `/workspace/scratch/pmfdg6q/cross-product/direction-mutant.log`.
