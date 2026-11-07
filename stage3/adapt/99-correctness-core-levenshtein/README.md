# L2 levenshtein: report, checker blocks the exact proposal

Task b5w3ycg. Stopped before adapter construction. This directory's adapt.cjs
is an explicit no-op so normal adaptation discovery remains usable. No compiler
source, existing adapter, library setting or LEDGER.md entry changes.

The requested JavaScript-changing established-shape candidate replaces only:

```ts
let previous = new Array(s2.length + 1);
let current = new Array(s2.length + 1);
```

with:

```ts
let previous = new Array<number>(s2.length + 1).fill(0);
let current = new Array<number>(s2.length + 1).fill(0);
```

The complete applied baseline passes ordinary adamic types on all 79 sorted
compiler roots. A scratch copy differing only at these two initializers fails
with seven diagnostics. Adamic fixes noUncheckedIndexedAccess to true in
internal/load/load.go:62. A filled number[] still has number | undefined reads;
the checker does not infer the loop's coverage. Four arithmetic reads report
TS2532, current[j] reports TS2322, the dist argument reports TS2345, and res
reports TS18048. full-types.json retains every exact location and message.
The baseline has zero diagnostics. provenance.json records source and binary
identity. All 746 source files are compared; only compiler/core.ts differs,
with precisely the two initializer substitutions in proposal.diff.

The pinned TypeScript 6.0.3 isolated owner check independently agrees: zero
baseline diagnostics, seven candidate diagnostics. A measurement-only proposal
adds six non-null assertions to previous[j - 1] (three occurrences), previous[j],
current[j - 1], and previous[s2.length]. That isolated stock check has zero
diagnostics. Those additional body changes are outside the explicit instruction
that nothing else in the body changes, and were not implemented or admitted.
Proceeding with that variant would require changing the unchanged-body constraint.

## Adapter 30 is present and untouched

30-indexed-reads/sites.json already records s1[i - 1] and s2[j - 1] in
levenshteinWithMax (entries at lines 966 and 978). Its assertions s1[i - 1]!
and s2[j - 1]! are present in the actual fully applied baseline and proposal.
The user clarified that these expression/occurrence pins do not reserve the
whole owner. Both assertions remain unchanged by the two initializer edits.

The focused paired replay runs actual adapter 30 against independent complete
source copies of the fully applied baseline and proposed candidate. Both exit
1 with the identical preexisting refusal: declined site changed: core.ts:992:21.
This is 30's original site label for addRange's from[i] decline, outside this
owner, rather than the current adapted source coordinate. Both leave every
source byte unchanged. Thus this full-tree refusal is not caused by our edit.

A second pair starts from the full apply's refs/stage3/30-indexed-reads-after
checkpoint and changes only our two initializers on one side. Both replays
exit 0 with identical stdout/stderr and zero source changes. replay30.json
includes all four outcomes and the exact checkpoint tree identity. These
results establish that the initializer edit preserves 30's expression pins.
A paired replay of every adapter was not run after the checker blocker.

## Initialization argument, line by line

This argument explains value preservation for the two initializers. It is not
an assertion that the rejected program passes the checker or the lane count.

1. The two allocations retain their original s2.length + 1 lengths. The
   number type argument is erased; fill(0) changes initial holes to zeros.
2. The first for loop writes previous[i] = i for every index 0..s2.length
   before any row or final read. Empty s1 also reads a written previous slot.
3. Every row computes minJ and maxJ. For finite max, minJ >= 1 and
   maxJ <= s2.length. An empty middle interval performs no middle-loop reads.
4. current[0] = i writes the boundary before any current read.
5. The j < minJ loop writes current[1..minJ - 1] to big.
6. The minJ..maxJ loop advances in order. Each current[j - 1] read sees
   current[0], a slot from the first fill loop, or the preceding middle-loop
   write. It writes current[j] before advancing to the next j.
7. All previous[j - 1] and previous[j] reads use the fully initialized
   preceding row. The existing asserted string reads are untouched.
8. The maxJ + 1 loop writes the remaining slots through s2.length to big.
   With current[0] and the earlier loops, every slot 0..s2.length is assigned
   before a swap. When max is negative, fill ranges can overlap or extend;
   the assigned values still match the original execution.
9. The give-up branch returns undefined based on the same colMin comparison,
   without reading an initial zero.
10. The swap makes this fully written current row previous and reuses the
    old row. Steps 4..8 overwrite every slot the next row can read. Induction
    covers the final previous[s2.length] read as well.

NaN and positive Infinity select minJ = 1 and maxJ = s2.length and have the
same coverage. Negative Infinity with nonempty s1 does not terminate in either
version. No real-call frequency is claimed for any of these classes.

## Node witnesses and unmeasured count

All ten Node witnesses compare equal using Object.is. They cover equal strings,
empty s1, empty s2, case-only difference, give-up undefined, max below one,
zero max, negative max, empty both, and an ordinary finite edit distance.
witness.json retains inputs and both results. They are independent edge
witnesses, not calls counted from the lane oracle.

The real-call count was not started. No zero-counterexample count, class count
of zero, successful after census or disappearance of F22..F36 is claimed.
The failure is the mandatory zero-diagnostic checker proof, before construction.
Adamic's Nov 7 03:20Z established-shape ruling is not an admission of this
checker-rejected candidate. No JavaScript change is submitted, so no ledger
acceptance line is added. PROOF.md records the complete stopping boundary.
