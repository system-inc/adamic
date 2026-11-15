# Completed watch-file 1532 re-cut proof

Candidate: `63f46490764862fef5af3ce6793557d4000994bc`. Parent exactly `c2f3c21f72b04e4c3acf7731913f0fc8ab89c2ca`. No implementation edits were needed during this resumed proof. Pushed and remote-verified as `typescript/watch-file-check-1532-2`; no force push.

Proof-only merge: `269c168bb3fc801034347c44f5af27eafa6c3d78`, parents candidate and main `fe5a1a54f045e9bce4e5aa32ca74885a317fbebd`. Baseline: chain tip merged with the same main, `0945d4355cad5afad46d35b14a18950db3775ade`. Neither merge is pushed.

## Results

- Full `stage3/apply.sh`: **0 on both trees**. Both product keys were freshly built from TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Full single-file sys.ts census: **0 on both runs**, `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_ONLY_FILE` set, CommonJSBundle. Both checker diagnostic lists are empty. The comparison deduplicates exact repeated observations then pairs by (kind, phase, reason, message with source positions stripped). **105 before, 104 after; 104 paired; 1 gone; 0 new; 0 changed**.
- The gone refusal is the original sys.ts:1532:13 nodeSystem.watchFile site, at **1548:13** in this composed baseline. Its original line reference is independently checked against the archived original cead sys-main.ts, not the pristine upstream line numbers.
- All five touched manifests regenerated from actual full-apply checkpoints using each directory's own digest/hash routine; crypto also cross-checked with its scoped-hashes.cjs. Every regenerated manifest is byte-identical to the candidate. **118 owner states checked**. No hash was entered manually.
- Five compiler builds: **0** (93 before/after, crypto before, final baseline/candidate). Final strict stock 6.0.3 checks: **79 roots, 0 diagnostics on each side**. Public typescript.d.ts and tsserverlibrary.d.ts are byte-identical.
- Lane: all 29 proof input files retain their expected hashes; none moved. sanctioned-api.json, shard-plan.json and test_check.py were not regenerated. No full 398-shard lane rerun was requested or claimed here.

## Directory proofs and mutants

All final positive proof invocations exit **0**. Explicit red controls exit **1**; generators that only construct a mutant exit 0 before an unmodified validator rejects it.

| Directory | Positive proofs, exit 0 | Negative checks |
| --- | --- | --- |
| 93-builtin-imports | bindings, owner-contract, built-witness | Binding mutants and owner/import negatives are asserted by their scripts. Scope mutant exits 1 but shares the waived main-red oracle; it is not independent evidence. |
| 99-correctness-core-multimap | contract, instrument, witness, body-proof, 95h contract | Restoring mutant generator 0; late adapter 1. Four 95h negative child exits are 1; owner drift/half-state negatives asserted. |
| 99-correctness-single-watcher-forwarder | contract, stock emits, original runtime witness (polling and fs paths) | Drift, half-owner, duplicate, moved and parse negatives asserted by contract. Runtime traces identical. |
| 99-correctness-timer-closures | contract, witness | Restored-slot, label and request mutants each 1. |
| 99-wasm-crypto-import | check, bindings, require-result contract, watch-directory contract/witness, resolution-host contract/runtime fixture, built-witness | Scope and ternary mutants 1. Require-result, resolve-host and three watch-directory mutant generators 0, each validator 1. Binding mutants and receiver-loss runtime negative asserted by scripts. |

The original interval fixture was rerun against freshly built baseline/candidate bundles: three enum members match; invalid 10 and undefined fail at the new boundary; the guard-restoring mutant exits **1** under the independent boundary validator. The non-dynamic interval-10 fixture passes for polling file, fsEvents file and polling directory. The original script's histogram input is explicitly labelled unmeasured/zero to require its non-vacuous fixture; no lane count is claimed from it.

## Authorized exception and resolved attempts

93 proofs/check.cjs is the sole user-authorized exception: it exits **1 on main, chain-plus-main and candidate-plus-main** on proper single-step pairs. The preceding recheck93 report retains that evidence. Its timer forwarding amendments occur in 93 itself and are missing from its undo oracle. No oracle source was changed.

Four initial proof attempts encountered sandbox `spawnSync ... EPERM` before assertions. Each was rerun unchanged outside the sandbox and passed: require-result contract, watch-directory contract, builtin built witness and crypto built witness. Two comparison-script setup assertions initially used the pristine upstream line reference and a pre-chain type annotation; the final comparator uses the archived original site and the actual nodeSystem declaration. All attempts remain in commands.jsonl; verification.json names their successful replacements. Preliminary suite/runtime summaries are preserved as attempt records and superseded by verification.json.

The immediate single-watcher-after stock emit records one temporary TS2322 because HostWatchFile is widened before the later timer adapter widens and checks the dynamic poller. The final completed-pipeline strict stock checks have zero diagnostics on both trees; runtime witness emission does not treat intermediate diagnostics as a final type result.

## Paths differing from 8499a514

All paths are under stage3/:

- adapt/99-correctness-core-multimap/README.md and owners.json retain the chain's diagnostic binding import amendment and utilities import pins.
- adapt/99-correctness-timer-closures/README.md and owners.json retain the InvokeMap and child-watcher amendments and recursive-watcher after pin.
- adapt/99-wasm-crypto-import/README.md and adapt.cjs retain require-result, fresh-resolution-host, watchDirectory and normalized whole-owner checking.
- adapt/99-wasm-crypto-import/proofs/targets.json pins the combined widened watch-file worker and chain transformations.
- upstream/LEDGER.md retains every parent byte and appends the original exact two-line addition; zero parent lines removed.

The six other paths from 8499a514's diff are byte-identical. final-source-audit.json lists every path. The commit message names every differing path.

## All command receipts

| Invocation | Actual exit | Expected exit | Seconds |
| --- | ---: | ---: | ---: |
| builtin-bindings | 0 | 0 | 14.151 |
| resolution-host-contract | 0 | 0 | 0.585 |
| typescript-dependencies | 0 | 0 | 15.157 |
| multimap-contract | 0 | 0 | 40.363 |
| single-watcher-contract | 0 | 0 | 1.054 |
| timer-contract | 0 | 0 | 7.29 |
| final-pin-regeneration | 0 | 0 | 9.267 |
| timer-restored-slot-mutant | 1 | 1 | 0.79 |
| timer-witness | 0 | 0 | 0.626 |
| timer-label-mutant | 1 | 1 | 2.105 |
| timer-request-mutant | 1 | 1 | 2.428 |
| build-builtin-before | 0 | 0 | 43.058 |
| builtin-owner-contract | 0 | 0 | 57.065 |
| crypto-check | 0 | 0 | 2.01 |
| crypto-bindings | 0 | 0 | 13.531 |
| crypto-scope-mutant | 1 | 1 | 0.367 |
| crypto-ternary-mutant | 1 | 1 | 2.815 |
| crypto-require-result-contract | 1 | 0 | 0.577 |
| crypto-watch-directory-contract | 1 | 0 | 0.636 |
| build-builtin-after | 0 | 0 | 40.463 |
| crypto-watch-directory-witness | 0 | 0 | 0.881 |
| crypto-require-result-mutate | 0 | 0 | 3.158 |
| crypto-require-result-mutant-red | 1 | 1 | 3.38 |
| crypto-resolve-host-mutate | 0 | 0 | 0.123 |
| crypto-resolve-host-mutant-red | 1 | 1 | 2.838 |
| crypto-watch-directory-restore-mutate | 0 | 0 | 0.065 |
| census-before | 0 | 0 | 152.048 |
| crypto-watch-directory-restore-mutant-red | 1 | 1 | 2.945 |
| crypto-watch-directory-unnormalized-mutate | 0 | 0 | 0.064 |
| crypto-watch-directory-unnormalized-mutant-red | 1 | 1 | 2.302 |
| crypto-watch-directory-optional-contract-mutate | 0 | 0 | 0.119 |
| crypto-watch-directory-optional-contract-mutant-red | 1 | 1 | 2.399 |
| multimap-mutant-generate | 0 | 0 | 0.809 |
| multimap-mutant-red | 1 | 1 | 0.519 |
| multimap-instrument | 0 | 0 | 4.995 |
| multimap-witness | 0 | 0 | 0.064 |
| multimap-body-proof | 0 | 0 | 0.568 |
| build-crypto-before | 0 | 0 | 31.48 |
| multimap-95h-contract | 0 | 0 | 36.48 |
| crypto-require-result-contract-retry | 0 | 0 | 40.142 |
| build-baseline | 0 | 0 | 45.499 |
| crypto-watch-directory-contract-retry | 0 | 0 | 32.544 |
| build-candidate | 0 | 0 | 37.666 |
| builtin-built-witness | 1 | 0 | 0.418 |
| crypto-built-witness | 1 | 0 | 0.417 |
| resolution-host-runtime-fixture | 0 | 0 | 8.499 |
| dynamic-interval-original-witness | 0 | 0 | 1.229 |
| non-dynamic-interval-10-fixture | 0 | 0 | 0.47 |
| dynamic-interval-guard-check | 0 | 0 | 0.729 |
| dynamic-interval-restored-guard-mutant | 1 | 1 | 0.62 |
| census-after | 0 | 0 | 150.136 |
| builtin-built-witness-retry | 0 | 0 | 2.547 |
| census-pair-comparison | 1 | 0 | 0.115 |
| crypto-built-witness-retry | 0 | 0 | 3.79 |
| census-pair-comparison-corrected-reference | 1 | 0 | 0.165 |
| census-pair-comparison-final | 0 | 0 | 0.165 |
| single-watcher-stock-before | 0 | 0 | 17.703 |
| single-watcher-stock-after | 0 | 0 | 14.57 |
| single-watcher-runtime-witness | 0 | 0 | 0.466 |
| census-pair-comparison-published | 0 | 0 | 0.115 |
| stock-final-baseline | 0 | 0 | 15.798 |
| builtin-scope-mutant-known-oracle | 1 | 1 | 0.518 |
| stock-final-candidate | 0 | 0 | 15.385 |

Full apply exits are recorded separately above; the same stdout/stderr logs are included. All runs use scratch TMPDIR, GOTMPDIR and Go cache. Source snapshots, generated full-owner manifests, emitted compiler source and candidate.patch stay local. This publication contains the requested raw census pair, logs (large assertion dumps reduced to excerpts with exact original hashes), proof scripts and structured receipts.
