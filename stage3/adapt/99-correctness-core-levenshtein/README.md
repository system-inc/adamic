# levenshteinWithMax: filled number rows and checked indexed reads

JavaScript-changing established-shape adaptation. TypeScript 6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8. Base main f0271c6a7; the existing
816190260 stopping report is preserved in initial-report.tar.gz. The user then
explicitly authorized the six non-null assertions and continuation on this branch.
Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.

Both row allocations become `new Array<number>(s2.length + 1).fill(0)`, and six
indexed reads gain `!`. No other owner text changes. The complete exact owner
texts and hashes are pinned in owners.json. Missing, duplicate, nested,
drifted and half-adapted owners refuse before any write. Replay writes nothing
and preserves bytes and mtime. LF/CRLF and all foreign source text survive.

30-indexed-reads already supplies s1[i - 1]! and s2[j - 1]! (sites.json entries
at lines 966 and 978). They are present in the before-pin and remain untouched.
The six new assertions use 30's rule class: checked indexed reads, erased in
emitted JavaScript and checked loudly by the native runtime. 30 could not add
these while the row elements were any. The user confirmed that its two
expression/occurrence pins do not reserve the complete owner.

## Each checked read is in range

Let n = s2.length. Both rows are filled number arrays of length n + 1 and
retain that representation across swaps. Whenever the middle loop executes,
minJ >= 1 and maxJ <= n; j increases from minJ through maxJ. The j - 1
predecessor is therefore nonnegative. `minJ` is 1 or the ceiling of the
positive difference `i - max`, so it is at least 1. `maxJ` is the floor of
a value no larger than n, so it is at most n. The final index n is always
in the row.

| Site | Read index | Why in range |
| --- | --- | --- |
| core.ts:2246:20 case-insensitive equal substitution (`previous[j - 1]!`) | `j - 1` | 0 <= j - 1 <= n - 1, because 1 <= minJ <= j <= maxJ <= n |
| core.ts:2247:20 case-insensitive unequal substitution (`previous[j - 1]!`) | `j - 1` | 0 <= j - 1 <= n - 1, because 1 <= minJ <= j <= maxJ <= n |
| core.ts:2249:19 equal code-unit branch (`previous[j - 1]!`) | `j - 1` | 0 <= j - 1 <= n - 1, because 1 <= minJ <= j <= maxJ <= n |
| core.ts:2250:39 delete candidate (`previous[j]!`) | `j` | 1 <= j <= n, because minJ <= j <= maxJ <= n and minJ >= 1 |
| core.ts:2250:68 insert candidate (`current[j - 1]!`) | `j - 1` | 0 <= j - 1 <= n - 1, because 1 <= minJ <= j <= maxJ <= n |
| core.ts:2267:17 final result (`previous[s2.length]!`) | `s2.length` | n = s2.length >= 0, and the row has length n + 1 |

The first five assertions are inside that middle loop. The last is after all
rows (including the empty-s1 case). Filling makes every in-range slot present;
the per-row assignment proof below additionally shows that a new zero cannot
influence the returned distance. These assertions resolve the seven checker
diagnostics in the original initializer-only proposal. They add no emitted
JavaScript beyond the two fill(0) calls.

## Why the initial zeros cannot affect a returned value

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


## Measurement before implementation

The complete count finished and was independently reconciled before adapt.cjs
was replaced with an implementation. Scratch instrumentation runs the exact
original and accepted owner bodies on every invocation, compares Object.is
return equality, throws on a difference, and returns the original result.
No compiler or lane-runner source is changed. In the accepted union, all 398
unchanged lane shards run once in 20 bounded pieces with their normal identity, baseline, API and
count guards. Raw arguments and both returns are retained. Input source and
built-file manifests remain frozen. Piece archives have SHA256SUMS and outer
SHA-256 files. The initial duplicate-node_modules setup failure recorded no
calls and is retained separately, excluded from this accepted union. A second
partial count was stopped after observed memory contention (2,611 calls, zero
differences), and the contended full control had three test timeouts. Both are
retained as excluded attempts. The accepted full control and count run in
isolation with all normal guards; excluded calls do not inflate the totals.

| Calls / class flag | Count |
| --- | ---: |
| Total calls | 798965 |
| Return differences | 0 |
| equalStrings | 0 |
| emptyS1 | 0 |
| emptyS2 | 0 |
| caseOnly | 157 |
| giveUp | 797742 |
| maxBelowOne | 46696 |

Class flags overlap and are not a partition. The separate unchanged full
stage3/lane/run.sh control runs from the pinned main worktree, reusing its
verified baseline product. Reconciliation checks identical test identities,
counts and raw baseline diff for that full control and the 398-shard union.
Node witnesses cover every requested edge class and ordinary finite distance;
they are excluded from the real-call count.

Full apply, full paired replay, 79-root types, JavaScript, public declarations,
complete core.ts census and initializer-only mutant results are in PROOF.md.
The paired replay requires adapter 30's outcomes to remain identical. Its
known addRange decline refusal on a fully adapted baseline is preexisting and
must match on both sides; successful replay is separately checked at its own
post-apply checkpoint. A refusal caused by these edits is a stop.
