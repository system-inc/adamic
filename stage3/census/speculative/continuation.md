This continuation retries the original 50 incomplete whole-project files.
Tasks: #pqjagjf and #j4wwk2f, on top of 76487f84 on the same branch.
Four single-CPU processes use 600-second file budgets; checker has a separate 1,800-second probe.
Completed records retain their original checksums and resource metrics; each file still loads all 79 roots.
The unit reserves time for publication and stops measurements before its two-hour deadline.

`retry_stream.py BINARY ROOT BASELINE OUTPUT SECONDS RSS_MIB WORKERS DEADLINE_EPOCH`
requires identical source hashes, project root and binary hash before carrying any
completed record forward. It preserves INPUT.baseline.json and changes only the
retry budget in INPUT.json. Four threads supervise four independent processes;
each compiler process uses GOMAXPROCS=1. The child is the same resumable stream.py,
filtered by LATENT_RUN_ONLY. Its GNU timeout is capped by the shared deadline,
including when a job starts late. ATTEMPTS.json journals every completed attempt,
and atomic per-file records remain the coverage source of truth. Optional
LATENT_RETRY_FILES selects additional retries after a resource failure.

The 600-second batch retains the original census executable. Only the separate
checker executable adds optional progress sidecars. Its control observations are
byte-identical to the original executable with progress enabled. The hooks count
unique nodes belonging to the selected file, using its AST inventory. Contract
and refusal scans share the scanner count; walker and lowerer counts are separate.
The examined count is their union. Scanner traversal is not value lowering, and
none of these counters claims completed-file coverage. Phase, current source
location and completed top-level statements describe the last checkpoint. The
snapshot is written synchronously at safe points, at most once per second plus
phase/unit boundaries; there is no concurrent map access from a timer goroutine.

The first attempts exposed real memory pressure: /tmp is tmpfs, initially holding
7.4GiB of scratch artifacts within a 16GiB cgroup memory limit. Two children were
SIGKILLed below their own RSS limits, with cgroup OOM kills recorded. A parser
attempt also hit its 3GiB RSS watchdog. Inactive prior-run and setup artifacts were
preserved on disk with symlinks at their original paths, reducing tmpfs use to
3.3GiB. Subsequent retries use GOMEMLIMIT=1GiB, a 3GiB RSS allowance and a hard
6GiB address-space cap. GOMEMLIMIT controls collection pressure; it is not a hard
cap. The initial attempts keep their own original resource observations.

The original checker probe stopped after 166.995 seconds with SIGKILL, peak RSS
2,027,788KiB. It was superseded by a fresh 1,800-second probe with scanner progress
and the lower heap target. The latter runs as a separate file-selected process,
concurrently with the four retry workers; wall times include this contention.
The process does not publish incomplete checker findings as census coverage.

`watch_stream.py` logs live RSS, checker checkpoints, completed-file counts and
cgroup events every 30 seconds. Interactive progress inspections occur at least
every five minutes. The unit started at 2026-10-09 22:05:15 UTC. The batch deadline
is 23:58:15 UTC, leaving seven minutes to publish by the two-hour cutoff. A full
50-file budget has a 130-minute ceiling at four workers; actual completion depends
on observed runtimes. Files capped by the unit deadline are distinguished from
files receiving their full 600 seconds.

Focused controls: audit_retry.py proves four overlapping children, zero compiler
relaunches on resume, and no launches after expiration. A one-worker mutant and a
future-deadline mutant fail independent live-process/launch witnesses.
audit_progress.py proves byte-identical census output and completed progress
counts matching the separate coverage recount; shifted-node and shifted-byte
mutants fail. audit_stream.py retains buffered parity, checksum rejection,
dropped-record and shifted-depth mutants. The unchanged production C/JS identity
witnesses and their appended-output mutants pass again.

Final publication combines verified completed records before finalize_stream.py
and verify.py. Every source keeps the original adapted bytes and whole-project
checker context. Older 30-second results remain historical evidence, not extra
observations added to the new totals.

The first queue was paused after 13 actual measurements to prioritize fresh
parser/utilities attempts and the 37 not-yet-started files. The second queue
retains all verified completed records and prior failed metrics. Eight files
already timed out under their full 600 seconds and are not rerun again. Deferred
first-queue journal entries are dispatch history, not measurements. The retry
control also rejects checksum corruption and missing successful records, and
proves that a paused queue launches no new measurement.

The final checker probe timed out after 1800.151s, peak RSS 1170596KiB. Its last
checkpoint at 1793.992s was checker.ts:13554:47, 39/51 statements completed,
9302 walker nodes and 44739 lowerer nodes. Scanner/walker/lowerer union was
298509/298510 selected-file AST nodes; scanner coverage does not establish a
completed value walk. The partial checker JSONL has only its project header and
is archived as incomplete evidence, never included among completed records.

The longer budget exposed a distinct failure in factory/utilities.ts: exit 2
after 82.782s, peak RSS 1551612KiB, `fatal error: stack overflow` after a goroutine
stack exceeded 1000000000 bytes. The captured stack repeats lower.inferTypes at
generic.go:298 and :307. It is not a timeout or an RSS-watchdog kill. Its record
is incomplete and excluded; no production fix is made by this census unit.

An additional checker probe is scheduled after the four-worker batch ends, if
a full 1800-second budget remains before the shared measurement deadline. It
runs exclusively with GOMAXPROCS=4, GOMEMLIMIT=4GiB, 6GiB RSS and 12GiB address
space. This avoids CPU contention and gives checker the box's four-CPU budget.
Its progress uses the same instrumentation already proved observation-identical
on the dependency control. A successful complete record must match the original
project header and source coverage before joining the cumulative stream. Its
canonical metrics record the distinct producer binary hash, and PRODUCERS.json
retains configuration and the earlier 600-second batch metrics. The batch's
default binary and source identity remain unchanged. An incomplete probe remains
outside the table. The first 1800-second probe stays archived independently.

The four-worker batch finished at 2026-10-09 23:16:22 UTC. All 50 originally
incomplete files received actual measurements: 36 completed, 13 timed out after
a full 600 seconds, and factory/utilities.ts hit the stack overflow described
above. Parser and utilities had additional fresh attempts after their initial
resource failures. No file was truncated by the unit deadline. Retaining the
original 29 complete records yields 65 files and 3760099 TypeScript source bytes.

The exclusive checker probe received its full 1800-second budget and timed out
after 1800.233s, peak RSS 2985980KiB, below its 6GiB RSS allowance. Its final
persisted checkpoint at 1792.399s was checker.ts:25576:17 (KindIdentifier, byte
span 1525684..1525688), with 39/51 statements completed, 14807 walker nodes and
83862 lowerer nodes. The scanner union remained 298509/298510 nodes. Source
locations are not monotonic because lowering follows calls, so that byte span
is a checkpoint location, not a covered prefix. Partial checker findings are
excluded; the complete-file census remains at 65/79 files and 3760099/10009820
TypeScript bytes. Both 1800-second probes and the earlier pressure failure keep
separate logs and metrics. The larger standalone CPU/heap budget reached more
nodes; no individual causal attribution is established by these two runs.
