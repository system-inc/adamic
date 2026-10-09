# Automatic worker merges

`fast-gate-watch.sh` enqueues a `codex/*` branch only when its completed verdict is
`green: <the exact gated sha> ...` and `$ADAMIC_FAST_GATE_WATCH_STATE/auto-area-merge`
is a file (state defaults to `~/.adamic-fast-gate-watch`). The switch is absent by default.
Integration can create it after review and remove it to stop new attempts. An already
running integration transaction completes under its existing locks; removing the switch
also stops records publishing until it is enabled again. Disabled verdicts are not backfilled.

The watcher runs one dispatcher in the background, without consuming a gate slot. Queue
entries survive restarts; completed entries deduplicate the branch/SHA pair. A dispatcher
file lock works on macOS through Python's `fcntl`, and serializes dispatchers after a
watcher restart. All merges are serial, including merges into different areas. The unchanged
integration script owns the actual local and origin area locks. A lock refusal records
`locked` and retains the job for a later poll; it never drops it. Removing the switch leaves
the queue intact. No conflict is resolved by the dispatcher.

## Remote execution and the Cloud area slot

Merges and merged-package tests run over `ssh -l ahra cloud`; the dispatcher never invokes
`area-merge.sh` locally. Cloud's `ahra` user has the push credential. Its independent checkouts
are `~/area-merge/integration` (trusted `cloud/merge-tree` tools, pinned to the commit used for
routing on the Mac) and `~/area-merge/tree` (the repository), with area worktrees under
`~/area-merge/worktrees`. The remote wrapper sources `~/adamic-tools/env.sh`, puts
`~/adamic-tools/bin` first for the pinned node and adds `${ADAMIC_TYPESCRIPT_SOURCE}/bin` to PATH.
It runs integration's unchanged merge script under `taskset -c 0-23`.

The dispatcher waits until the watcher's slot table has a free `threadripper B` slot, counting
running records by their assigned slot (third column) and box (fourth; legacy entries default
to threadripper), exactly as the watcher does. `threadripper` is Cloud's name in this table;
`cloud` is the SSH alias for access as ahra. Workshop/Server/Home cannot substitute for it.

The watcher and dispatcher share a short directory guard, `<state>/slot-table.lock`, around
slot calculation and recording a process. SSH receives no script on stdin until its running
record exists: `<state>/running/<ssh-pid>` contains `area-merge/<area> <sha> B threadripper B`.
The entry remains while SSH runs and is removed in cleanup, before copying artifacts. The
reaper explicitly skips dead `area-merge/*` entries, so a dispatch failure never becomes a
void gate or gets inserted into the gate queue. Ordinary watcher shutdown releases its own
guard. An uncatchable death during allocation can leave the directory guard; integration
must verify the allocators stopped before removing it. Removing the switch cancels a pending
slot wait and leaves the queued job intact.

The remote wrapper retains stdout and copies transaction log directories into
`~/area-merge/out/<run-id>`, with a path manifest. The Mac copies this bundle into
`<state>/area-logs/<run-id>` (or `<state>/dry-run/<run-id>`), rewrites diagnostic paths to the
local copies for worker notices, and retains SSH output even if SCP fails. Only mktemp log
directories are bundled; arbitrary paths mentioned in test diagnostics are not copied.

Every dispatcher `ahra os send`, including named Circle recipients, passes
`--from system_adamic_developer_tools`. These sends still run from
`/Users/kirkouimet/Projects/ahra` and retain once-per-branch/recipient behavior.

Remote wiring review: based on `origin/devtools/fast-gate`
`a9749bc378266bb358928f315756012b77c5ac4f`. Twenty-nine dispatcher tests passed (the optional
integration-origin test was skipped); all nineteen `cloud/fast-gate/run_test.py` tests passed.
Shell syntax and whitespace checks passed. Transport tests use executable local SSH/SCP/Ahra
stubs and a local fake remote environment, including slot claim/cleanup, occupied-slot waits,
shared allocation guard, actual watcher allocation/reaping, remote affinity/environment/log
bundles, and sender identity. No real boxes were contacted and the switch was not enabled.
The first live remote merge is reserved for integration's review.

## Routing and holds

Integration owns routing. The dispatcher fetches `origin/cloud/merge-tree` into its stable
Git checkout at `<state>/area-integration` and loads `cloud/integration/area-route.py` from
that checkout. It calls `route(branch)` unchanged, accepting `(area, how)` or `("hold", why)`.
There is no watcher-owned fleet table or prefix policy. Importing the trusted source creates
no bytecode or other modifications inside integration's checkout. A missing, broken, or
malformed router holds the job for integration.

A routing hold, refusal, or unrecognized outcome remains in `<state>/area-queue/<key>.held`,
with branch, SHA, reason, routing decision and outcome. Only merged, red and conflict outcomes
become `.done`. The dispatcher reevaluates held routes each poll: a changed `(area, how)` puts
the held job back into the queue. Unchanged holds are not attempted again on every poll.
Integration can manually retry a cleared refusal by renaming that job's `.held` file to `.json`.

Refusal lines (`refused: ...`) are checked before exit status: they never become worker reds.
The two existing busy-lock messages are recorded as `locked`, remain `.held`, and retry next
poll so normal area serialization cannot drop a worker. All other refusals and unknown exits
stay held pending a changed route or manual requeue. Lock refusals also notify integration,
so a stale lock remains visible. Named failing-test/gofmt/vet verdicts with exit 1 are red;
exit 3 without a refusal is conflict; exit 0 without a refusal is merged.

`<state>/area-held.tsv` lists active held branches and reasons. Per-branch JSON entries under
`<state>/area-held` retain successful notification recipients across polls, restarts and new
SHAs. A hold sends branch, SHA and reason once to `system_adamic_integration`, and once to each
Circle named as `@name` or `system_*` in the reason, using `ahra os send` from
`/Users/kirkouimet/Projects/ahra`. Failed sends retry without repeating successful recipients.
Routing decisions are never inferred from changed files. No conflict is resolved here.

## Records and worker feedback

Each attempt journals a CSV row before acknowledging the queued job. A detached records
worktree appends it to `documentation/velocity/auto-area-merges.csv` and pushes only
`records/auto-area-merges`, with ordinary fast-forward pushes. Failed pushes retain the
journal for replay; replay checks the current remote CSV for the exact row before appending.
CSV columns: `utc,branch,sha,area,outcome,seconds,area_sha_after`. Routing holds use `no-area`;
non-lock refusals and unknown exits use `held`. An unchanged held decision generates no new
attempt/row each poll. The final area SHA is read from origin, including after a failed worker
merge whose main catch-up already succeeded.

Merged, red and conflict results look up the newest `ai.db` reply naming the complete branch.
A longer branch name cannot identify a shorter branch's worker. The dispatcher sends a one-line
success note on merged. Reds and conflicts send the script output, including conflicting paths,
named failing tests and referenced failure logs, using `ahra ai send <session> --message-file`
from `/Users/kirkouimet/Projects/ahra`. No session or a send failure is logged locally.
Refusals send no instruction to the worker to fix its branch. `ADAMIC_AI_DATABASE` has the gate's
same default. Dry runs send no worker or Circle messages.

## Previous routing follow-up validation

Merged gate tools `4af2c103d4e5e0f5de3a969fed5d4a3a7ea4276a` as a true merge. Enqueueing is
inside the reaping loop after `voidCause()` accepts the verdict, with exact SHA and switch
checks. Three slots, `--class`, shared classification and `cloud/land-*` watching are preserved.

Shell syntax checks passed. Nineteen offline tests passed, including actual watcher reaping,
route delegation, persistent holds/requeue, unknown outcomes, refusal priority, once-per-recipient
notifications, merged worker feedback, no-push behavior, and isolated Git records replay.

Run integration's own routing coverage test as part of this suite:

```sh
ADAMIC_FAST_GATE_WATCH_STATE=/path/to/review-state \
  PYTHONDONTWRITEBYTECODE=1 python3 cloud/auto-area-merge-test.py --integration
```

This fetches and runs the unchanged `cloud/integration/area-route-test.py` in the trusted
checkout. In this Linux review workspace, integration tools
`09d3f789aae1bfe33243dc3edbca46aeae9171bc` reported 390 unlanded branches: 275 routed, 23 held
by a named row, 71 undecided one-offs, and failed on these remaining recurring families:

- `async`: codex/async-ordinary, codex/async-typeof
- `census`: codex/census-generic-returns, codex/census-small-families, codex/census-small-families-3
- `error`: codex/error-classes-counts, codex/error-classes-counts-2
- `generic`: codex/generic-function-value, codex/generic-function-value-host-scratch
- `non`: codex/non-null-check, codex/non-null-checked, codex/non-null-checked-area, codex/non-null-narrowed-number
- `require`: codex/require-builtins, codex/require-builtins-2
- `shared`: codex/shared-ssa-conditional-copy, codex/shared-ssa-unit-a
- `string`: codex/string-views, codex/string-views-concurrency
- `tsgo`: codex/tsgo-c-library, codex/tsgo-errors-as-values

The Mac's live `ai.db`, fleet roster and Ahra CLI are unavailable here. The integration test's
fleet fallback therefore has no local evidence, which may account for these failures. Integration
owns any metadata correction; its files were not edited. The switch remains absent and no area
was pushed during review. Live worker/Circle delivery and Mac execution are not verified.

## Original local dry run (historical; current dry runs use SSH)

```sh
ADAMIC_FAST_GATE_WATCH_STATE=/path/to/review-state \
  bash cloud/auto-area-merge.sh --dry-run codex/stage3-a-check-headers \
  150c331b9429284004a0ac137e59a3207877a33a
```

Dry run bypasses the enable switch, passes `--no-push`, sends no worker messages, and writes
CSV locally to `<state>/dry-run/documentation/velocity/auto-area-merges.csv`. It does not
publish records or areas. Integration's `--no-push` still acquires/releases origin lock refs
if an actual merge is needed. Dry rows use `merged` for a successful no-push judgment; inspect
the log to distinguish a tested candidate from a branch the area already contains.

Demonstrated in the Linux review workspace, with the switch absent, using integration tools
`f8dfc8f3bafa2866fcc3becdb9924d64bef1efa0`. This verifies the already-contained path, not a new
merge's tests or the Mac execution environment. The real worker's published gate evidence is
`gate-logs/150c331b9429/20261008T050742Z/fast`: green in 14.9 seconds, 46 pass, 0 skip.

Selected area: `stage3`. Unchanged area-merge output:

```text
area/stage3 ef3141e9 already holds codex/stage3-a-check-headers 150c331b
```

CSV row:

```csv
2026-10-08T09:17:24Z,codex/stage3-a-check-headers,150c331b9429284004a0ac137e59a3207877a33a,stage3,merged,1.006,ef3141e9b1152ab51b51497f8ce3a2799449c8a3
```

Original implementation verification: `bash -n cloud/fast-gate-watch.sh cloud/auto-area-merge.sh` and
`PYTHONDONTWRITEBYTECODE=1 python3 cloud/auto-area-merge-test.py` (8 offline tests). Tests cover
mapping overrides/ambiguity, switch-off, lock retries, queue deduplication, no-push invocation,
failure feedback, exact session lookup, and real Git records retry against an isolated local origin.

During review, an initial records test incorrectly used the source repository's origin and
published a synthetic row. A normal forward correction removed it; no area was pushed. The
isolation bug was fixed and the entire test suite then passed against the local test origin.
