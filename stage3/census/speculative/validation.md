The main port's controls and their mutants pass; the whole-project census remains bounded and resumable.
Compiler base: 946a8f095a7fa419a92117406314b7b3d44630f0; original corpus bytes are preserved.
Every measurement and build has an outer timeout; focused Go tests use -timeout 90s.
All outputs are captured in named logs, losslessly archived under evidence/main-stream/.
Historical validation and directory evidence are retained at f6bb0b41 and evidence/directories/.

Source `/workspace/adamic-tools/env.sh` after setup, with GOPROXY exported as
`https://proxy.golang.org|direct`. The cold setup hit its 240s limit during build
warming. The warm retry completed in 74.816s: Go/node .079s, submodules .120s,
markdown .122s, clang .201s, shared cache .500s, build 74.698s, test binaries
74.788s, cache warm 74.789s. nproc=5, cgroup quota=4 CPUs. Go 1.27.1,
Node 24.19.0, clang 20.1.8. No whole-package confirmation or full gate was run.

Focused logged checks: audit.py (eight nested sites), audit_signature.py (five
checker-clean sites through depth 3), audit_dependency.py (foreign depths 1/2),
audit_topology.py (eight exact identities), prove_identity.py, audit_snapshot.py,
audit_binding.py, audit_output_guards.py and isolation.py. They reject depth-zero,
no-stubs, omitted-boundary, legacy-site-key, span-only, token-identity, hidden-binding,
shared-slice, scalar-value, cursor, private-field, mapper-identity, schema,
IR-output and loader-output mutants. The unchanged Gate.aCheck at 0b76ca8b
checks six controls with zero mismatches and catches eight header mutants.

The final streaming binary passes audit_stream.py: buffered findings parity,
resume skipping, checksum rejection, dropped successful records, depth shifted by
one and independent arithmetic mutants. Resource witnesses catch a sleeping
process at its deadline, resident-memory excess and address-space allocation
failure. RLIMIT_AS is the hard memory cap; RSS watchdog sampling is documented.

The standalone checker probe uses timeout 90s, 6GiB RSS allowance and 12GiB hard
address-space cap. It exits 124 after 90.099s with peak RSS 940780KiB and no complete
record. All-file speculative measurements use 30s per file under an outer 3600s
timeout. Corpus full/no-stubs baselines use 5s per file and count only completed
records. No 15-minute interval is left unmonitored.

The mapper fixture from a64f77ee has matching source Node, generated JS Node and
sanitized native output, with empty stderr. A scratch overlay reverting inferTypes
from 6b33ec61 to its parent fails that fixture. Main's equivalent fix is therefore
reported separately as already on main; no production mapper commit is carried.

Pinned main, normal overlay-off and speculative flag on the production binary
emit identical C (6907 bytes) and JS (9930 bytes). Appended-output mutants fail
both byte comparisons. Production source diff is empty.

Reproduction uses the commands in streaming.md. Run finalize_stream.py under
an outer timeout with NODE_PATH selecting the pinned 6.0.3 stage3 API cache, then
verify.py on the resulting raw stream, RESULT.json and compiler root. Finalizer
and verifier log the required shifted-depth and dropped-file failures. Lane
checks run after each commit before its push; their logs are archived.

Continuation tasks #pqjagjf and #j4wwk2f retain the same compiler and corpus.
The new setup completed in 15.926s: Go/node .021s, submodules .059s,
markdown .065s, clang .147s, shared cache .857s, build 15.808s,
cache warm 15.898s. nproc=5, four-CPU quota; tools retain the versions above.
Logs for this continuation are under evidence/continuation, with historical
30-second measurements preserved separately under evidence/main-stream.

The retry launcher runs four single-CPU children with 600-second time limits,
3GiB RSS watchdogs and 6GiB address-space caps. The separate checker probe uses
1800 seconds with the same hard memory bounds and a 1GiB Go heap target.
Its last persisted AST checkpoint is archived independently of census coverage.
The initial memory-failed attempts and the prioritized continuation queue keep
separate journals. A deferred queue entry is not counted as a measured file.

Focused commands, each with stdout/stderr in its named log:

```sh
# Controls have outer timeouts and internal deadlines; output directories are scratch.
python3 stage3/census/speculative/audit_retry.py SCRATCH
python3 stage3/census/speculative/audit_progress.py OLD_BINARY PROGRESS_BINARY SCRATCH
python3 stage3/census/speculative/audit_stream.py PROGRESS_BINARY SCRATCH
python3 stage3/census/speculative/isolation.py REPOSITORY SCRATCH 946a8f095a7fa419a92117406314b7b3d44630f0
# Background measurements have an additional shared unit deadline.
python3 stage3/census/speculative/retry_stream.py BINARY ROOT BASELINE OUTPUT 600 3072 4 DEADLINE_EPOCH
# Finalization has a 300s supervisor and individual 90s stock/report audits.
python3 stage3/census/speculative/timed_run.py 300 6144 METRICS LOG python3 stage3/census/speculative/finalize_stream.py ROOT OUTPUT FINAL
# Independent verifier has an outer 90s timeout.
python3 stage3/census/speculative/verify.py FINAL/speculative.jsonl FINAL/RESULT.json ROOT
```

The retry control measures four overlapping independent processes, resume with
zero launches, paused queues with zero launches and expired deadlines. Serial
worker and future-deadline mutants fail those independent observations; corrupt
checksums and dropped successful records fail before reuse. The progress control
compares raw census output byte for byte and rejects shifted node/byte counters.
The stream control rechecks buffered parity and depth/coverage mutants. The final
stream audit and independent verifier recheck the published cumulative tables.
Overlay-off C and JS retain their prior byte sizes and hashes; appended-output
mutants are caught again. No production compiler file or .a fixture changed.

The isolated four-CPU checker probe ran after all retry workers ended, with
GOMEMLIMIT=4GiB, GOGC=200, a 1800s timeout, 6GiB RSS and 12GiB address-space
bounds. It timed out after 1800.233s with peak RSS 2985980KiB. The earlier
one-CPU probe remains archived. All 50 incomplete files were actually measured
by the batch; none lost its budget to the shared deadline. Thirteen timed out
at 600s and factory/utilities.ts exited 2 with a captured stack overflow.

Task #b2wbbha adds exact 8-piece watch and 5-piece transformer union checks.
Both match 0442c4a9 sites/reasons/combined depths/boundaries with zero differences.
Dropped, duplicated and moved-site mutants fail independent partition/provenance
audits. The final binary retains piece-off observations and actual watch-piece
observations byte for byte. New resume, timing and publication guards pass their
mutants; stream depth/coverage and C/JS identity witnesses pass again. Five checker
samples complete and createNodeBuilder times out at 900.070s; no new full-file
coverage is claimed. See evidence/pieces/README.md and SHA256.json.
