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
