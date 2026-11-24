# Family 63 splice truncations

Adamic ruled this adaptation on Nov 23 2026 (#bbssdcy, family 63, row 324) and accepts the head after its review and census.

Rulings 323 and 497 keep direct array length assignment refused. Row 324
allows this JavaScript-changing adaptation for bounded truncations. The
two calls below are expression statements whose return values are unused.

- `program.ts`, `computeCommonSourceDirectoryOfFilenames`. At a non-root
  mismatch, `0 < i < n <= commonPathComponents.length`. Canonicalization
  receives strings and cannot reach the path arrays. The first-file assignment
  stores the component array itself, so either truncation mutates that same array
  in place. The later assignment using `sourcePathComponents.length` is outside
  the ruling.
- `sys.ts`, `pollWatchedFileQueue`'s `nextPollIndex`. The wrapping branch
  gives `0 <= definedValueCopyToIndex < pollIndex === queue.length` with no
  callback between the checks and the truncation. Both resets follow the call.

The ruled `factory/emitNode.ts` site stops at the condition. A caller of the
public `moveEmitHelpers` API can empty the source helper array through
`getEmitHelpers(source)` in its predicate and then return true. The old write
throws for the resulting negative length. The proposed negative splice start
returns normally. On Nov 23 at 09:06Z, Adamic ruled that the site stays stopped
and the head goes on with two sites.

Run the guard, idempotence, array and emitted-owner checks on a tree that
`stage3/apply.sh` wrote with this adapter. These cover both zero-cut cases,
the resumed-poll case and the bounded queue and directory sweeps. The queue
sweep covers every in-bounds start and chunk size from 0 through length plus
one for queues of up to four slots. The directory sweep covers every list of
one or two paths and every list of three paths with at most two directory
names per path. The domain uses roots `c:/` and `d:/` and up to three directory
names per path drawn from `a`, `A` and `b`:
`NODE_PATH=stage3/api/node_modules node stage3/adapt/99-seats-build22-63-splice-truncations/test.cjs <applied-tree>`.
The test reads this adapter's input from its `-before` ref and its recorded
output from its `-after` ref. Later adapters can edit `program.ts` or `sys.ts`
without affecting these checks.
