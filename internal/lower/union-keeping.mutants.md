# Union keeping stop mutants

The stop rejects a multi-value union when either side mentions Weak, including
when the scan cannot prove the type free of Weak. Member-by-member splitting is
the follow-up, preserved separately from this narrow head.

## Remove the stop

Remove the pre-`present` guard in `sameKeepingWalk`, retaining `mentionsWeak`.
Run `go test ./internal/lower -run '^TestUnionKeepingAssignments$' -v`.

Observed exit 1: all fourteen p5–p8 rows lower again and fail with
`want assignment NotYet, got program=true error=<nil>`. The rows are p5 and p6
before/after, p7, and p8 before/after, each in `.a` and `.ts`. Both p9 rows keep
their original stop and pass. Named red test:
`TestUnionKeepingAssignments/p5_after.a`.

P7 uses a function declaration instead of an explicitly typed arrow initializer.
Otherwise the conservative stop rejects the initializer's identical Weak-union
function view before reaching the cross-view assignment. The declaration leaves
the measured assignment at line 7, column 117.

## Search only top-level members

Replace `mentionsWeak` with a scan that asks `weakTarget` only of the root and
its immediate union/intersection members, without descending into properties,
arguments or signatures. Run
`go test ./internal/lower -run '^TestUnionKeepingNestedWeak$' -v`.

Observed exit 1: both `.a` and `.ts` lower again and fail with
`want nested Weak assignment NotYet at 7:24, got program=true error=<nil>`.
Named red tests: `TestUnionKeepingNestedWeak/.a` and
`TestUnionKeepingNestedWeak/.ts`.

The field is two levels deep: `v: { readonly inner: Weak<Box> }`, viewed from
the same discriminated union whose `inner` is strong. The non-mutant has an
assignment `NotYet` at line 7, column 24 in both extensions.

Both mutants were built with Go overlays; the mutant replacements were never
written to production files. The fixed source was checked against its saved
copy after the stop-removal run.

## Leaves, limits and generic instances

`any`, `unknown`, `never`, primitives/literals (including booleans), and type
parameters are known leaves. They contain no structural Weak child to walk.
The scanner first applies the current generic mapper; the stop also counts
members of the concrete instance, rather than a generic union that collapses
under substitution. Unhandled kinds, recursive references cut short by the
active visited set, and scans reaching 64 active types fail closed.

`TestUnionKeepingDepthLimit` reaches that arm from a program with 70 finite
readonly object levels and no Weak. In both extensions it stops at the union
initializer, line 74, column 20, rather than declaring an incomplete scan safe.

`TestUnionKeepingGenericWeakInstance` uses the requested generic return form
with `T extends object`, the constraint required by the prelude's `Weak<T>`.
Called with `T = Box`, it stops at `return x`, line 4, column 9, in `.a` and `.ts`.
`TestUnionKeepingMappedWeakArgument` additionally hides Weak only inside `T`.
A direct T witness supplies the mapper (main does not infer it from the union
alias alone); the `T = Weak<Box>` instance stops at line 5, column 24.

The checked-downcast and `box?.name ?? 'none'` controls, and the no-Weak union
controls, remain held to source Node. The split follow-up patch is kept outside
this branch; this commit does not change generic inference or split unions.
