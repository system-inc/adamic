# SourceFileLike position candidate: built and validated

Adaptation: [`typescript/sourcefile-like-position-1`](https://github.com/system-inc/adamic/tree/typescript/sourcefile-like-position-1), commit `0ed68349e7c60764a5333f0d2d326965a64d53d1`. One commit on main tip `578ace8577a350d7f7623316c40532c8d40e9998`.
Upstream: `050880ce59e30b356b686bd3144efe24f875ebc8`, exact tag `v6.0.3`.

All required checks passed. Nothing was submitted to Loom.

| Check | Result |
| --- | --- |
| Ownership | New `99-correctness-sourcefile-like-position`; 71-writable-views owns lineMap only |
| JavaScript | 604 files / 1,208 emits byte-identical, comments kept and removed |
| Full scanner census | 138 -> 137; only scanner.ts:469:12 removed; no new or changed findings |
| Full types.ts census | 0 -> 0 findings |
| SourceFile-to-SourceFileLike replay | utilities.ts:7959:37: no refusal on either tree |
| Upstream declarations | Separate npm ci and unmodified Hereby dts on main and candidate both pass |
| typescript.d.ts and tsserverlibrary.d.ts | Both byte-identical; hashes and deterministic gzip copies retained |
| Lane API check | Pass; sanctioned-api.json unchanged; 273 sanctioned declarations |
| Adamic types | All 79 sorted compiler roots; exit 0, zero diagnostics |
| Adapter contract | Exact measured candidate; no-write idempotence; owner-drift, duplicate and parse-error no-write refusals |
| Restoring mutant | Reinstates the method and scanner.ts:469:12's unbound-method refusal (exit 1) |
| Shard regeneration | All three generated files byte-identical; generate_shards.py --check passes |
| Full lane | PASS: 106,366 passing, one sanctioned API baseline failure, zero pending |
| Applied source identity | All 747 source files equal the measured candidate |

## Declaration measurement replacement

The initial flattened declaration probe was invalid on main (64 diagnostics from missing Node ambient types). It measured no declaration equivalence and is retained only as historical evidence. At the user's direction it was replaced by TypeScript 6.0.3's own declaration build: `npm ci` and `node_modules/.bin/hereby dts`, preserving every project's options and invoking the unmodified `scripts/dtsBundler.mjs`. Main succeeded before the candidate run began. Each ran in its own adapted checkout and dependency installation.

`upstream-declarations/comparison.json` records exact byte equality. `api-check.json` calls the unchanged lane API guard on actual bundled declarations. `adamic-types.json` retains the exact 79-root command and exit; complete stdout is compressed as `adamic-types.stdout.gz` and stderr is empty. The replacement commands, build logs and all result codes are in `upstream-declarations/`.

## Final lane and main advance

The unfiltered, eight-worker lane passed without changing checker options, oracle expectations, sanctions or worker count. `upstream-declarations/full-lane/` retains its report, execution, verdict, patch table, oracle report, baseline diff and compressed phase logs. `candidate-apply-key.txt`, `adapter-inputs.json` and `final-input-check.json` identify the tested adapter bytes; `applied-source-comparison.json` proves all source files equal the measured candidate.

Measurements began on main `751879544ea6e627e67c91d6ab3c4ba595be6356`. Main advanced during the lane by two unrelated wasm errno/docs/test-coverage commits. Before committing the adaptation, the branch was fast-forwarded onto main's new tip. `main-advance.json` lists every changed path and proves stage3 inputs, the apply key and adapter bytes unchanged. No load/lower sources, dependencies, stage3 pipeline or TypeScript pin changed, so the passed lane measures the same relevant inputs on the final base.

The initial environment interruption killed the first apply without an exit code; only the successful fresh-output rerun was used. Raw censuses and timings, the complete finding diff, exact candidate patch, source hashes, owner search and original command ledger remain alongside this report. The evidence branch contains logs and censuses; the adaptation directory contains only adapt.cjs, owners.json, mutant.cjs, README.md and PROOF.md.
