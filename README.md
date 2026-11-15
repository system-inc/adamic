# Completed: watch-file 1532 re-cut validated

See [completed/README.md](completed/README.md) and [completed/verification.json](completed/verification.json). Candidate `63f46490764862fef5af3ce6793557d4000994bc` passes the remaining proof with the user-authorized 93 scope-oracle exception. Earlier stop reports remain below as history.

# Latest result: chain-plus-main also fails 93 single-step check

See [recheck93/README.md](recheck93/README.md). Main, chain-plus-main and candidate-plus-main all return adapter exit 0 and check exit 1 on the immediate single-step pair. Stopped as explicitly requested.

# Stopped: watch-file 1532 re-cut

No adaptation branch pushed. Unverified local candidate: 63f46490764862fef5af3ce6793557d4000994bc. Parent exactly c2f3c21f72b04e4c3acf7731913f0fc8ab89c2ca. Scratch proof merge with main fe5a1a54: 512f2a8ef (candidate tree identical; subsequent amendment changed only the message). The proof merge is not pushed.

## Blocker

93-builtin-imports/proofs/check.cjs exited 1 on actual full-apply before/after snapshots. Its source oracle permits only three imports and removed declarations, but the adapter also changes timer forwarders and widens fsWatchFileWorker. The oracle does not undo these edits. Changing this proof script is outside 8499a514's diff and is not a forced pin change, so it conflicts with the byte-identical carry-over rule.

## Proof status

- Actual 93 source-scope proof: exit 1. Command and snapshot tree IDs: 93-pipeline-result.json.
- 99-wasm-crypto-import/proofs/resolve-host-contract.cjs: exit 0; 128 presence cases.
- Census overlay/binary build: exit 0; full CommonJSBundle census pair not run. No claim about sys.ts:1532, new or changed findings.
- Full apply and baseline apply: cancelled, exit 143, after the proof blocker. No full-green claim.
- Remaining proofs, witnesses and mutants: not run after stop. No all-mutants-red claim.
- Lane: none of the changed paths is among sanctioned-api.json's 29 proof inputs. No lane files regenerated; no freshness pass claimed.

## Paths differing from 8499a514

Paths below are relative to stage3/:

- adapt/99-correctness-core-multimap/README.md: retains diagnostic binding import amendment.
- adapt/99-correctness-core-multimap/owners.json: retains utilities namespace-import pins.
- adapt/99-correctness-timer-closures/README.md: retains InvokeMap and child-watcher amendments.
- adapt/99-correctness-timer-closures/owners.json: retains recursive-watcher after text and digest.
- adapt/99-wasm-crypto-import/README.md: retains require-result, fresh-resolution-host and watchDirectory amendments.
- adapt/99-wasm-crypto-import/adapt.cjs: retains those transformations and normalized whole-owner assertion.
- adapt/99-wasm-crypto-import/proofs/targets.json: combines widened watch-file worker with chain owner transformations. Hashes generated using the directory's scoped-hashes.cjs; final full-pipeline pin validation was not reached.
- upstream/LEDGER.md: exact parent bytes followed by original two-line addition; zero parent lines removed.

The other six original-diff paths are byte-identical. Other owner pins were mechanically merged; requested full-pipeline regeneration was not completed before the stop.

## Setup and retained evidence

Fresh detached scratch worktrees; shared checkout files unchanged. Recursive depth-1 submodules initialized; npm ci in stage3/api completed. TypeScript source pinned to v6.0.3 050880ce59e30b356b686bd3144efe24f875ebc8. TMPDIR, GOTMPDIR, Go cache and census module cache in scratch. No Go source created under stage3. Nothing submitted.

Full logs, actual before/after source snapshots and candidate.patch remain under /workspace/scratch/watch-file-proof. This published bundle contains commands, statuses, hashes and a short assertion excerpt only; no raw source snapshots or patch. Automatic approval review rejected the larger raw-source evidence upload; this narrower bundle excludes that material.
