Task #b2wbbha adds statement pieces to the census overlay. Production compiler
source remains unchanged. The union comparison is the admission check for this
mode, not an inference that arbitrary splitting preserves a compiler IR.

`plan_pieces.cjs FILE N PLAN.json` uses stock TypeScript 6.0.3. It inventories
one atom per top-level statement. For the pinned checker.ts, the statements
inside createTypeChecker replace that function as atoms; statements outside the
function remain atoms. An atom retains its exact UTF-8 byte span and syntax kind.
The planner assigns the largest atoms first to the currently smallest bucket,
then restores source order within each bucket. It produces at most N nonempty
pieces: an indivisible function or namespace can dominate a bucket, and a file
with fewer than N atoms gets fewer pieces. It does not rewrite source bytes.

`LATENT_PIECE_PLAN` and `LATENT_PIECE_ID` enable selection only inside the guarded
measurement overlay, together with speculative mode and single-file selection.
Every piece loads the same 79-root compiler project and repeats the existing
census refusal/contract scan, inheritance/accessor registration, global/enum/class
registration and nested declaration registration. These are the existing census
adaptations for a checker-rejected program, rather than a new production lowering
path. The assigned body roots use the same speculative recovery and typed state
snapshots. Generic calls cannot lower an unassigned body in the selected file;
foreign dependencies retain their existing whole-project observation behavior.
There is no exceptions fixpoint, cycle, readiness, borrow or backend pass.

A checker body atom receives its enclosing createTypeChecker signature and
function context without lowering the enclosing body. Actual ancestor reads use
the existing checked lexical-binding seeding. Body findings carry their scheduling
atom as provenance; declaration/scanner findings carry `prepass`. A finding's
actual site can lie in a foreign dependency, so provenance is checked against the
scheduled unit rather than incorrectly requiring every site span inside its atom.

`run_pieces.py BINARY ROOT PLAN OUTPUT SECONDS RSS_MIB WORKERS` persists each
completed piece with a checksum and verifies input identity before reuse. Worker
count is independent of identity, permitting a larger worker pool on resume.
`LATENT_PIECES=0,1,...` selects a sample; omit it to run the entire plan. Each child
uses one CPU, GOMEMLIMIT=1GiB and GOGC=200. GNU timeout and a two-second KILL grace
bound wall time; timed_run.py applies an RSS watchdog and a hard address-space cap
of twice the RSS allowance. Progress sidecars count this piece's assigned units.
The supervisor records child CPU usage and the first sampled wall time of each
phase; the first `walk` sample includes loading, scanning and registration, with
roughly 100ms sampling latency plus scheduling delay. Progress and timing do not
change observations. A failed or timed-out piece is never a completed record.

`check_piece_union.py ROOT PLAN RUN WHOLE_RECORD STOCK_JSON OUTPUT` first checks
that stock AST statements equal the planner inventory, that atoms partition the
plan exactly once, and that every piece and every scheduled finding has valid
provenance. All headers and source hashes must agree. It combines typed failed
boundaries and independently resolves depths through stock AST ancestor chains,
then compares the unique `(kind, where, reason, text, site_where, site_kind,
site_start, site_end)` findings and depths against the whole-file record. Full
added/removed sites, depth differences and boundary differences are in UNION.json.
Raw per-piece depths are provisional and must never be added as depth tables.

Dropped-piece and duplicated-piece mutants fail the exact piece inventory.
The moved-site mutant transfers a body finding between records while preserving
its site identity, reason and overall multiplicity; atom/unit provenance rejects
it even though an ordinary union of sites would be unchanged. The timing control
rejects a shifted prepass timestamp and erased CPU usage. Existing progress,
stream depth/coverage and output identity controls remain applicable.

A piece row has coverage for its assigned AST roots, not a completed file. Its
source_bytes is an input identity, not additional complete-file coverage. Do not
feed piece rows directly to finalize_stream.py or sum their source byte fields.
The published 65-file table stays unchanged. This unit intentionally does not
finish any of the 14 remaining files; the six checker samples do not establish
checker completeness. A later big-box publication must require every planned
piece, audit the structural envelope, combine boundaries and recount the whole
project before claiming new complete-file bytes.

The exact watch.ts and transformers/ts.ts references are the raw whole-file
records archived by 0442c4a9 under evidence/continuation/retry-priority/records.
The new lossless piece records, plans, checksums, metrics, comparison ledgers and
producer provenance are archived under evidence/pieces. Sources must retain the
original adapted corpus bytes and absolute root used by those records when
reproducing the byte comparisons.

Reproduction commands, with stdout/stderr redirected to separate named logs:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
timeout 180 python3 stage3/census/latent/make_overlay.py "$PWD" SCRATCH/overlay
timeout 180 go build -buildvcs=false -overlay=SCRATCH/overlay/overlay.json -o SCRATCH/census ./stage3/census/latent/tool
timeout 30 node stage3/census/speculative/plan_pieces.cjs ROOT/watch.ts 8 SCRATCH/watch-plan.json
timeout -k 5 2500 python3 stage3/census/speculative/run_pieces.py SCRATCH/census ROOT SCRATCH/watch-plan.json SCRATCH/watch-run 600 3072 3
timeout 30 node stage3/census/speculative/plan_pieces.cjs ROOT/transformers/ts.ts 5 SCRATCH/ts-plan.json
timeout -k 5 3100 python3 stage3/census/speculative/run_pieces.py SCRATCH/census ROOT SCRATCH/ts-plan.json SCRATCH/ts-run 600 3072 1
# Compare both unions before proceeding; use decompressed 0442c4a9 records and stock.json.
timeout 120 python3 stage3/census/speculative/check_piece_union.py ROOT PLAN RUN WHOLE_RECORD STOCK_JSON UNION_OUTPUT
timeout 180 python3 stage3/census/speculative/plan_remaining.py ROOT stage3/census/speculative/RESULT.json SCRATCH/plans 64
LATENT_PIECES=0,1,2,3,6,9 timeout -k 5 1900 python3 stage3/census/speculative/run_pieces.py SCRATCH/census ROOT SCRATCH/plans/checker.ts.plan.json SCRATCH/checker-sample 900 3072 4
timeout 30 python3 stage3/census/speculative/project_pieces.py SCRATCH/plans SCRATCH/checker-sample SCRATCH/ts-run SCRATCH/watch-union SCRATCH/ts-union PROJECTION_OUTPUT
```

The projection charges measured checker prepass overhead to every planned piece,
then extrapolates remaining body cost by byte size. Completed samples retain their
actual times. Median and sample-extrema scenarios use longest-first scheduling on
14, 32 and 64 single-CPU workers. These are conditional estimates, not statistical
intervals or completion guarantees. Other files can have different scan costs,
pathological bodies or the previously observed generic-inference stack overflow.
The attempt-budget ceiling is separate from a forecast of completed coverage.
Scaling assumes equivalent cores and enough RAM for the worker pool.

Final result: both unions match; five checker samples complete and piece 0 is
censored at 900 seconds. Publication tools now explicitly reject piece-tagged
records as complete files; audit_piece_publication.py proves all three guards.
See evidence/pieces/README.md for the complete report and conditional projection.
