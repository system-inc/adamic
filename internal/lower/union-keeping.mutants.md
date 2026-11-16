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

The parameter-direction mutant and its dedicated pin were removed. With no
assignability calls, the cross product of union member pairs is symmetric:
reversing direction changes no answer. The ordinary fixed-parameter and
rest-element comparisons therefore use main's order.

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

## Stored Weak member identity

Remove the `insideSlot` member-identity block from the union walk.

```
go test ./internal/lower -run '^TestUnionKeepingWeakMembers$/^field' -v -count=1 -timeout 10m
```

The command exits 1. `TestUnionKeepingWeakMembers/field.a` and
`TestUnionKeepingWeakMembers/field.ts` both fail with `got <nil>, want NotYet`.
The types are `{ readonly item: Weak<A> | Weak<B> | C }` and
`{ readonly item: Weak<A> | B | C }`. The view loses B's Weak keeping while both
whole field unions still contain a Weak member.

The rule runs only below the top value pair, through stored contents or a
callable parameter or result. Ordinary writes from `Named` into
`Weak<Counted | Named>` still lower. The top-level mixed-representation view
retains its existing NotYet, and the keeping walk's top pair stays exempt.

The file was restored. Output:
`/workspace/scratch/pmfdg6q/member-scope/member-rule-mutant.log`.
The runtime probe report is
`/workspace/scratch/pmfdg6q/weak-member-probes/report.md`.

## Callable field twins (#svbn0aj)

Make `sameCallablePair` answer true when both sides are callable, restoring the
old skip in the inner stored-value comparison. Function members of the union
remain skipped at the top of `sameKeepingStorage` (#q6wf1bp).

```
go test -count=1 -v -run '^TestUnionKeepingCallableFieldsStop$/^a1_parameter' ./internal/lower/
```

The command exits 1. `TestUnionKeepingCallableFieldsStop/a1_parameter.a` and
`TestUnionKeepingCallableFieldsStop/a1_parameter.ts` fail with
`got <nil>, want NotYet at main.a:6:24` and
`got <nil>, want NotYet at main.ts:6:24`, respectively. The field types are
`(x: number | string) => string` and `(x: number) => string`. The assignment
lowers again when callable twins are not compared. The file was restored.
Output: `/workspace/scratch/svbn0aj/callable-twins-mutant.log`.

The a2 result-width pin also stops: `() => number` seen as
`() => number | string`. The same top-level view stops at `1:61` in both `.a`
and `.ts`, so this agrees with the ordinary callable walk, not a new price.
Top-level probe outputs: `/workspace/scratch/svbn0aj/top-result-a.log` and
`/workspace/scratch/svbn0aj/top-result-ts.log`. Identical function types and
unchanged function fields keep lowering and agreeing with Node.

## Callable growth nesting reset

Restore the unconditional `nesting = &keepingNesting{}` at each ordinary union
block, leaving the fresh visited map unchanged.

```
go test -count=1 -v -timeout 5s -run '^TestUnionKeepingCallableGrow$/^\.a$' ./internal/lower/
go test -count=1 -v -timeout 5s -run '^TestUnionKeepingCallableGrow$/^\.ts$' ./internal/lower/
```

Both commands exit 1 with `panic: test timed out after 5s`. The mutant times out;
no stack overflow was observed before the timeout. The pin loads
`G<number> | { readonly kind: 'n' }` and its structural `Twin<number>` union
without a checker relation, then runs the keeping walk. Its callable field
parameter grows through `G<G<T>> | { readonly kind: 'n' }`. With shared nesting,
both file modes finish and answer not same at the nesting bound.

The file was restored. Full outputs:
`/workspace/scratch/svbn0aj/nesting/reset-mutant-a.log` and
`/workspace/scratch/svbn0aj/nesting/reset-mutant-ts.log`.

## Single callable union member (#xmnp0x8)

Restore the unconditional `return true` for callable union members at the top
of `sameKeepingStorage`, removing its `singleCallable` comparison.

```
go test -count=1 -v -run '^TestUnionKeepingFunctionMemberStop$' ./internal/lower/
```

The command exits 1. `TestUnionKeepingFunctionMemberStop/.a` and
`TestUnionKeepingFunctionMemberStop/.ts` fail with
`got <nil>, want NotYet at main.a:7:19` and
`got <nil>, want NotYet at main.ts:7:19`, respectively. The types are
`((x: number | string) => string) | Box` and
`((x: number) => string) | Box`. The parameter view lowers again when the only
callable member is skipped. The file was restored.
Output: `/workspace/scratch/xmnp0x8/member-skip-mutant.log`.

The result twin `(() => number) | Box` to `(() => number | string) | Box`
also stops at the view. Its top-level scalar-result view stops the same way
at `1:61`; the union member follows the ordinary callable rule. Two or more
callable members retain the combined-signature behavior left to #q6wf1bp.
