# Batch 6: non-null split and adaptation 49

The branch rebased cleanly from 2ce525eb. Its first current-main base was
92d19601. Baseline c4c59914 and subsequent main bdb89962 have identical stage3
inputs; those main advances change Go tests outside stage3 only.

The fresh full lane (bash stage3/lane/run.sh /workspace/batch6-49) passes
in 564.292 seconds: apply/build succeed, 106366 upstream tests pass, one known
Public APIs failure remains, and zero tests are pending. The baseline lane has
exactly the same counts and baseline.diff bytes. artifact-proof.cjs verifies
all ten built JavaScript files, public API bytes, every local baseline and three
real byte mutants. The zero-instrumentation audit checks 756 files and catches
its injected-hook mutant.

Focused check.cjs, idempotence.cjs, mutants.cjs and preflight-tests.cjs pass:
eleven assertions repaired, eleven blockers retained, all eleven restored-owner
mutants caught, and a removed-owner mutant rejected before writing. The split
again gives 50 placeholders, seven property/call lies and fifteen other lies.
API/helper tests pass, catching five measurement and four slot-observer mutants.
Three incoming .a refusal headers and all three real header mutants pass on
the current compiler. The receipt identifies its actual build commit 6cd03842.

The exact slots-bundle.cjs and read-probe.cjs project measurements each match
all stdout/stderr/exit goldens for 301 projects with TSC_JOBS=1 and a 55-second
per-case deadline. Slots take 450.742 seconds; moved reads take 189.244 seconds.
The slot collector reproduces 25 observed write-before-read rows, 19 read-before-
write rows and six unobserved rows. The read collector sees all 116 probes,
38 reached and 78 unvisited, with 910898035 checks and no nullish read.
runtime-mutants.cjs against the pinned original observer catches all 22 real
undefined unwraps on their triggering inputs. nonnull-proof.json holds receipts.

Each focused process has a 90-second deadline; lanes and aggregate project
runs have 900-second deadlines, with four CPU affinity. Output goes to files.
No whole Go package/full Adamic gate or historical native feature-area replay.
Unvisited reads and the eleven retained blockers remain outside any safety claim.
No new fixture bodies or compiler files were authored. Setup is the shared
batch-6 setup: 74.750 seconds, nproc 5, cgroup quota four CPUs.
