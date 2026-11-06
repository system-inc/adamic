# 99z elementAt: counted road (b), Step A complete

No initializer edit has been made. adapt.cjs remains the Step 1 no-op.
The zero count gate is met. Step B is not yet started at this checkpoint.
Its requested equal witnesses for absent fractional and NaN inputs contradict
the unchanged fallback's measured behavior; that clarification remains pending. No Math.trunc or lib edit is
introduced to conceal that difference.

## Rulings

Adamic, task axochk, Nov 1 16:06Z: "tsc is checked against the lib its own tsconfig declares."
Adamic, task b5w3ycg, Nov 6 21:49Z: road (b), keep only elementAt's fallback as a
JavaScript-changing adaptation, admitted only on a count. Lib es2022 is refused.
The earlier MEASURE-ELEMENTAT.md is historical; its missing lib ruling is now
supplied here. No adaptation pins this initializer; 99z owns the proposed work.

Base: origin/main d0279c2bb, merged without rewriting the existing feature branch,
with 99r amendment 980870faa63a164ad82200d7688218dbc31d163a retained. The fresh
full apply exited 0 and all 746 source files match the prior full-applied input
used for the measurement copy. See fresh-apply.json in the closeout archive.

## Complete call count

The single union of all 398 unchanged lane manifest units passed their ordinary
identity, baseline and API guards: 106,366 passing tests, one sanctioned public
API failure, zero skipped tests. The counter recorded 1,936 calls. Every observed
offset is -1, 0 or 1. Independent reclassification of every raw line agrees.
Non-integer offsets are zero even when undefined-array calls are considered.

| Class | Calls | Required caller sites and call counts |
| --- | ---: | --- |
| in-range non-negative integer | 414 | not requested |
| in-range negative integer | 1,138 | `src/compiler/transformers/es2015.ts:4452` (569); `src/compiler/transformers/es2015.ts:4471` (569) |
| out-of-range integer | 217 | `src/services/codefixes/convertToAsyncFunction.ts:389` (211); `src/services/codefixes/convertToAsyncFunction.ts:392` (5); `src/services/codefixes/convertToAsyncFunction.ts:395` (1) |
| non-integer | 0 | none |
| array undefined | 167 | `src/services/codefixes/convertToAsyncFunction.ts:269` (134); `src/services/codefixes/convertToAsyncFunction.ts:273` (3); `src/services/codefixes/convertToAsyncFunction.ts:277` (30) |

ELEMENTAT-COUNT.json contains the piece records, full callers and whole-run
comparison. Each successful piece was committed and plainly pushed before any
adaptation edit. The maximum measured successful oracle-piece duration was 110.45 seconds.
There are 20 successful pieces, each with at most 20 units.

## Measurement method and retries

The complete adapted source was copied into scratch. Only that copy's elementAt
initializer is wrapped, calling a file recorder before invoking the unchanged
original implementation. Its original body and feature test remain intact.
The wrapper preserves source line counts. instrument.json and the frozen src
and built-file manifest bind the measured code. Every piece rechecks that
manifest. The wrapper is never installed by adapt.cjs.

One synchronous append is made per call, with offset and array length. NaN,
infinities and negative zero are encoded explicitly. Undefined arrays have null
length and their own class. Callers are required and recorded for every class
except the first. Raw calls are inside the piece archives; caller source maps
and recorder hashes are in the closeout archive.

The first two successful pieces use two concurrent units and the normal
30-second deadline. An attempted third piece hit 17 deadlines; its complete
raw evidence is retained in elementat-count-03.tar.gz and excluded from the
canonical census. No result or diagnostic was relabeled. Subsequent pieces use
four units and the runner-supported STAGE3_SHARD_SECONDS=90 measurement bound;
tests and all identity, baseline and API guards remain unchanged. These pieces
remain below the requested 25-minute limit. The recorder also resolves source
maps only for exceptional calls, without globally enabling Node source maps.
A 263-call comparison against recorder v1 agrees on every offset, length, class
and required caller. Seed calls and diagnostic retries are excluded explicitly.

After the count, the unchanged stage3/lane/run.sh ran as a separate control,
with no calls added to the census. Its lane verdict passes. The complete test
identity multiset, counts and raw baseline diff equal the 398-shard union.
This verifies the entire lane rather than a selected test subset. The control
was bounded to 1,400 seconds and finished within that bound. Whole-control
phases, logs and raw task records are packed in the closeout archive.

## Witness conflict and stop

The zero non-integer count meets Step B's count gate. However, Step B also asks
for equal original/fallback witnesses for every class, including absent ones.
The existing fallback cannot meet that requirement on ordinary nonempty arrays:

| Input array | Offset | Original elementAt | Unchanged fallback | Equal |
| --- | --- | --- | --- | --- |
| [10, 20] | 0.5 | 10 | undefined | no |
| [10, 20] | -0.5 | 10 | undefined | no |
| [10, 20] | NaN | 10 | undefined | no |

The Node program extracts the original owner and its actual existing fallback,
plus toOffset, from the full-applied tree. It also records equal witnesses for
all four observed classes and for both infinities. elementat-witness.cjs and
witness-draft.json retain the exact inputs, bodies and results.

No claim of general equivalence is made, and no degenerate empty-array witness
is used to hide these counterexamples. The pending question is whether the
absent non-integer witnesses should document these differences under the
count-based ruling, or whether equality remains mandatory. At this Step A
checkpoint, no fallback adaptation, owner amendment, Step B mutant, paired
replay, declaration/JavaScript comparison or new per-file lowering census is
claimed. The exact source rewrite is authorized by the zero count; the unresolved
question concerns the all-class equality requirement.

## Evidence and reproduction

The elementat-count-NN archives retain each piece's raw calls, observations,
reports and logs, with internal SHA256SUMS and adjacent archive SHA-256 files.
The retry archive is separate from the failed third-piece attempt. The closeout
archive adds producer provenance, the independent count reconciliation, exact
source maps, Node witnesses, and unchanged whole-run control. All Go evidence
retains .go.txt names. No compiler or runner source is modified.

The reproduction scripts are elementat-instrument.cjs,
elementat-count-preload.cjs, elementat-prepare-count.py,
elementat-count-pieces.py, elementat-reconcile-count.py,
elementat-whole-control.py and elementat-witness.cjs. They use the task's
/workspace paths and fresh scratch directories. Failed attempts and planted
probes are never included in the canonical union.
