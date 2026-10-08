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

## Mapping assumptions

The dispatcher fetches `origin/cloud/merge-tree` into its own stable Git checkout at
`<state>/area-integration`, and uses that checkout's complete `cloud/integration` file set.
It invokes that checkout's `area-merge.sh` unchanged. Candidate branches cannot supply the
merge tools or mapping tables.

Exact entries in `census-owners.tsv` override prefixes, including overrides with no area.
Otherwise, use the literal `prefixOwners` table from integration's `branch-census.py`, plus
worker prefixes that spell an area exactly: `codex/<area>-`. Resolve the resulting owners
through `areas.tsv`. The census's `system_cohere_lint` and `system_cohere_format` are aliases
for `stage1-lint` and `stage1-format`; their shared parent `system_cohere_adamic` is ambiguous.
These are the assumptions used to reconcile the census and area tables. All matching
prefixes are considered, so competing areas are rejected instead of choosing the first.
An unknown or ambiguous mapping logs `no area for codex/x` and records `no-area` with an
empty area column. There is no ownership inference from changed files or commit messages.

## Records and worker feedback

Each attempt journals a CSV row before acknowledging the queued job. A detached records
worktree appends it to `documentation/velocity/auto-area-merges.csv` and pushes only
`records/auto-area-merges`, with ordinary fast-forward pushes. Failed pushes retain the
journal for replay; replay checks the current remote CSV for the exact row before appending.
CSV columns: `utc,branch,sha,area,outcome,seconds,area_sha_after`. The final SHA is read from
origin, including after a failed worker merge whose main catch-up already succeeded.

Conflict (integration exit 3) and red results look up the newest `ai.db` reply containing
the literal branch name, as the fast gate does. The dispatcher sends the script's output,
including conflicting paths and named new failing tests, plus referenced failure logs, using
`ahra ai send <session> --message-file <file>` from `/Users/kirkouimet/Projects/ahra`.
No session or a send failure is logged locally. `ADAMIC_AI_DATABASE` has the gate's same default.

## Review dry run

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

Verification: `bash -n cloud/fast-gate-watch.sh cloud/auto-area-merge.sh` and
`PYTHONDONTWRITEBYTECODE=1 python3 cloud/auto-area-merge-test.py` (8 offline tests). Tests cover
mapping overrides/ambiguity, switch-off, lock retries, queue deduplication, no-push invocation,
failure feedback, exact session lookup, and real Git records retry against an isolated local origin.

During review, an initial records test incorrectly used the source repository's origin and
published a synthetic row. A normal forward correction removed it; no area was pushed. The
isolation bug was fixed and the entire test suite then passed against the local test origin.
