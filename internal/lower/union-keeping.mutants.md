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

## Repeating lap through a callable slot (#ttq1599)

Remove the open-pair check at the top of `sameKeepingStorage` (the `nesting.open[pair]`
early return and the set's bookkeeping), so a storage pair already open on the walk's
path is compared again.

```
go test -count=1 -v -timeout 5m -run '^TestUnionKeepingRepeatingLapFinishes$' ./internal/lower/
```

Expected: `TestUnionKeepingRepeatingLapFinishes/member.a`, `member.ts`, `field.a` and
`field.ts` each fail with "the keeping walk didn't finish in 60 s", or the process stops
on Go's stack limit first. Either one is red. The types repeat with the same ids every
lap, so the nesting bound never says no, and the ordinary union block's fresh visited
never sees the pair again.

## Source method facing a view property (#ttq1599)

Restore `inside.Flags&ast.SymbolFlagsMethod == 0` in
`sameKeepingStorage`'s property loop, leaving the view-method `continue` alone.

```
go test -count=1 -v -run '^TestUnionKeepingMethodPropertyStop$' ./internal/lower/
```

The command exits 1. `TestUnionKeepingMethodPropertyStop/.a` and
`TestUnionKeepingMethodPropertyStop/.ts` fail with
`got <nil>, want main.a:8:117: stage 0 can't lower ... yet` and
`got <nil>, want main.ts:8:117: stage 0 can't lower ... yet`, respectively.
The object-literal method's `(x: number | string) => string` is seen through
`(x: number) => string`; the assignment lowers again when source methods are
skipped. The mutant was planted once and the file restored after the run.
Output: `/workspace/scratch/ttq1599/source-method-mutant.log`.

## Retained storage arms (#q6wf1bp)

Each mutant below was planted once using an isolated Go overlay, then excluded
from subsequent runs; the working compiler source remained unmutated.
Rows 11 and 12 retain their original unreadable-side skips. Row 6 retains its
combined-signature skip, and unequal tuple lengths retain the shorter-list comparison.

### Callable own properties (row 5)

Restore the top-level either-callable skip, and make the one-callable arm in `sameCallablePair` return true instead of comparing properties.

```
go test -count=1 -v -timeout 3m ./internal/lower/ -run '^TestUnionKeepingPositiveProofArms$/^(constructor_static|nested_constructor_static)'
```

With the mutant, the command exits 1. The pins `constructor_static.a/.ts and nested_constructor_static.a/.ts` under
`TestUnionKeepingPositiveProofArms` go red: `got same without a storage proof`.
Output: `/workspace/scratch/q6wf1bp/final-mutants/mutants/c-callable-properties/run.log`.

### Property source against an index view (row 9)

Remove the property-to-index comparisons when the source has no index infos and the view does.

```
go test -count=1 -v -timeout 3m ./internal/lower/ -run '^TestUnionKeepingPositiveProofArms$/^property_to_index'
```

With the mutant, the command exits 1. The pins `property_to_index.a/.ts` under
`TestUnionKeepingPositiveProofArms` go red: `got same without a storage proof`.
Output: `/workspace/scratch/q6wf1bp/final-mutants/mutants/e-property-index/run.log`.

### Every tuple element against an array view (row 10)

Restore comparison only up to the shorter type-argument list. This skips the tuple’s later elements against the array’s one argument.

```
go test -count=1 -v -timeout 3m ./internal/lower/ -run '^TestUnionKeepingPositiveProofArms$/^tuple_later_element'
```

With the mutant, the command exits 1. The pins `tuple_later_element.a/.ts` under
`TestUnionKeepingPositiveProofArms` go red: `got same without a storage proof`.
Output: `/workspace/scratch/q6wf1bp/final-mutants/mutants/f-tuple-elements/run.log`.

### Typed-array storage kind (row 14)

Remove the typed-array kind check from `sameKeepingStorage`.

```
go test -count=1 -v -timeout 3m ./internal/lower/ -run '^TestUnionKeepingPositiveProofArms$/^typed_array_kind'
```

With the mutant, the command exits 1. The pins `typed_array_kind.a/.ts` under
`TestUnionKeepingPositiveProofArms` go red: `got same without a storage proof`.
Output: `/workspace/scratch/q6wf1bp/final-mutants/mutants/g-typed-array/run.log`.

### Buffer storage view (row 14)

Remove the `nodeBufferView` check from `sameKeepingStorage`.

```
go test -count=1 -v -timeout 3m ./internal/lower/ -run '^TestUnionKeepingPositiveProofArms$/^node_buffer_kind'
```

With the mutant, the command exits 1. The pins `node_buffer_kind.a/.ts` under
`TestUnionKeepingPositiveProofArms` go red: `got same without a storage proof`.
Output: `/workspace/scratch/q6wf1bp/final-mutants/mutants/g-buffer/run.log`.

### Explicit receiver slots (row 15)

Remove both the explicit-this presence check and the explicit-this slot comparison from `sameCallablePair`.

```
go test -count=1 -v -timeout 3m ./internal/lower/ -run '^TestUnionKeepingPositiveProofArms$/^(this_slot|this_missing)'
```

With the mutant, the command exits 1. The pins `this_slot.a/.ts and this_missing.a/.ts` under
`TestUnionKeepingPositiveProofArms` go red: `got same without a storage proof`.
Output: `/workspace/scratch/q6wf1bp/final-mutants/mutants/h-this-slots/run.log`.

### Accessor against a field view (row 16)

Skip source get/set accessor symbols in the property loop.

```
go test -count=1 -v -timeout 3m ./internal/lower/ -run '^TestUnionKeepingPositiveProofViewsStop$/^r16'
```

With the mutant, the command exits 1. The pins `r16.a/.ts` under
`TestUnionKeepingPositiveProofViewsStop` go red: the view lowers instead of producing NotYet.
Output: `/workspace/scratch/q6wf1bp/final-mutants/mutants/r16-accessor/run.log`.

## View method signatures (row 7, #q6wf1bp)

Restore the `viewed.Flags&ast.SymbolFlagsMethod != 0` `continue` in
`sameKeepingStorage`'s property loop. Plant it once using a Go overlay, then
exclude the overlay from all subsequent runs.

```
go test -count=1 -v -timeout 5m ./internal/lower/ -run '^TestUnionKeepingViewMethodsStop$/^union_weak_function_strong_view'
```

With the mutant, the command exits 1. Both calling pins
`TestUnionKeepingViewMethodsStop/union_weak_function_strong_view.a` and `.ts`
fail with `got <nil>, want NotYet at main.a:10:24 for a Held seen as a Viewed`
and the corresponding `main.ts:10:24` text. The view lowers again.
The no-call ArrayLike pair is pinned at its earlier unsupported parameter stop,
`main.a:8:15` and `main.ts:8:15`, rather than at the view.
Output: `/workspace/scratch/q6wf1bp/row7/mutant.log`.
