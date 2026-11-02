# Adaptation 93 recut on main wave 3, #he8tysb

Main: `80c2d3282e44b527555d59690bb413cc883d2bf2`. Source: `d3efff7b594eca31e02bc83f6195fb736a2e64f0`.
The adaptation directory, its historical proofs/logs, and `stage3/lane/evidence/adapt-93-main/` are byte-identical to the source. None of its three lane manifests or root patch-set.md was copied. One new commit is based directly on this main tip.

## Official table generation

`bash stage3/apply.sh /workspace/scratch/adapt-93-tip-report/after --write-table`
regenerated the table. A fresh main control used `stage3/apply.py --adapt-checkout` on the same pinned TypeScript commit. No generated file was merged textually.
Counts are files / added / removed. Every changed row relative to main's committed table follows:

| Adaptation | Main's recorded row | Regenerated with 93 |
|---|---|---|
| 40-explicit-any | 17 / 106 / 106 | 18 / 107 / 107 |
| 93-builtin-imports | absent | 1 / 3 / 3 |
| 95-checker-axolotl-a | absent | 1 / 13 / 13 |
| 95-checker-axolotl-b | absent | 1 / 7 / 7 |
| 96-emit-axolotl-a | absent | 1 / 14 / 14 |
| 96-emit-axolotl-b | absent | 1 / 2 / 2 |
| 96-emit-axolotl-c | absent | 1 / 19 / 19 |
| 97-cli-axolotl-a | absent | 1 / 1 / 1 |
| 98-axolotl-a | absent | 2 / 8 / 8 |
| 98-axolotl-b | absent | 1 / 5 / 5 |
| 99-correctness-axolotl-a | absent | 4 / 12 / 12 |
| **Total** | 85 / 5525 / 5483 | 85 / 5610 / 5568 |

The fresh main control total is 85 / 5607 / 5565; with 93 it is 85 / 5610 / 5568.
Row 40 still accounts for 141ad1d2's tsbuildPublic.ts:424 timer slot, absent from main's stale table.

## Manifest input proof

Official preparation: `python3 stage3/lane/prepare_shards.py /workspace/scratch/adapt-93-tip-report/artifacts`.
Main's artifact key was `510ad1bf10456d87e5f0f1fd6d21db21a65f6a9de239c5773c287d4dc95d20e3`; the with-93 key is `82d58286bab3977b9a4a7cf512edfefe65ee86423469655aba89bbbf5f5ca741`.
The old full manifest was rejected:

```json
{"name": "stage3-lane-file-000", "status": "fail", "counts": {}, "seconds": 2.352198727996438, "errors": ["shared artifact inputs hash does not match this checkout and toolchain"]}
```

The same generators used for main's 398-unit manifests were run:

```sh
python3 stage3/lane/shard_plan.py /workspace/scratch/adapt-93-tip-report/artifacts/82d58286bab3977b9a4a7cf512edfefe65ee86423469655aba89bbbf5f5ca741
python3 stage3/lane/generate_shards.py
```

Only the artifact key and all 398 per-unit input hashes change. Definitions, selected identities, expected counts, API digest, harness digest and partitioning are unchanged. The evidence gzip's decompressed bytes were identical, so its timestamp-only rewrite was restored.
Main's old unit hash remains rejected against the regenerated plan:

```json
{"name": "stage3-lane-file-000", "status": "fail", "counts": {}, "seconds": 0.017765907003195025, "errors": ["unit inputs hash differs from the manifest"]}
```

The exact manifest hash was then passed to the unit:

```sh
python3 stage3/lane/shard.py stage3-lane-file-000 /workspace/scratch/adapt-93-tip-report/unit-regenerated --inputs-hash a64a51f30fb50e2e625073a558fd03c6be1f375faad9c8e8ee7014f29546050a
```

Exact output:

```json
{"name": "stage3-lane-file-000", "status": "pass", "counts": {"pass": 6, "fail": 0, "skip": 0}, "seconds": 16.18010723200132, "errors": []}
```

The immutable artifact is separate from both oracle working trees. Official preparation verifies a payload-valid cache hit after the unit and again after all tests and oracles. Its current key still equals the manifest key after recording this evidence, which lives under artifact.py's excluded lane/ directory.

## Whole lane and adapter

The real TestProduct_Stage3AdaptedTree ran first. Then `go test -json -count=1 -timeout 60m ./stage3/lane/...` exited 0, with 421 top-level pass, 0 fail, 0 skip, in 734.624 seconds.
Both runs have outer bounds and a 3 GiB free-space floor.

The adapter directory is unchanged. Full sys.ts pins remain:

- Before: `5c45814e8fb6155701df1ef9be2755096ccac593ac6d53b13a4d26d90c7284d8`.
- After: `be2c80c35536b9d0a05a66df61b5f229155325353caf0718da64f459754ea121`.

The rerun reports zero writes and preserves bytes and nanosecond mtime. An added newline is refused with `target source moved: sys.ts` before any write, retaining the moved bytes and mtime. Source inventory, exact undo, both comments modes of the transpile comparison, and binding/collision checks pass. No pins were updated: none of the new adapters edits sys.ts.

The built-compiler Node witness also passes: fs, path and os load at module top in with-93 rather than inside getNodeSystem, and both tsc.js commands produce byte-identical stdout, stderr, exit status, JavaScript and declarations for the small project.

## What main changed, and fresh serial oracles

The requested repository diff since 470e16f8 contains 90 paths. Actual changed compiler files were read from each official adapter's before/after snapshot refs:

| Adaptation | Changed compiled source |
|---|---|
| 95-checker-axolotl-a | `src/compiler/checker.ts` |
| 95-checker-axolotl-b | `src/compiler/checker.ts` |
| 96-emit-axolotl-b | `src/compiler/transformer.ts` |
| 97-cli-axolotl-a | `src/compiler/program.ts` |
| 98-axolotl-a | `src/compiler/transformers/classFields.ts`, `src/compiler/transformers/es2017.ts` |
| 98-axolotl-b | `src/compiler/transformers/es2015.ts` |
| 99-correctness-axolotl-a | `src/compiler/transformers/esDecorators.ts`, `src/compiler/transformers/esnext.ts`, `src/compiler/transformers/jsx.ts`, `src/compiler/transformers/module/esnextAnd2015.ts` |

These are ten distinct compiler files. 96b changes sparse feature-flag writes and 97a changes a paths-substitution condition in JavaScript text; the remaining adapters add assertions erased by Node. The lane's API-sanction generator, its Python tests/Go wrappers and README, and its three generated manifests also moved. apply.py and sys.ts did not move.
Since compiled source changed, no prior oracle was reused. Both fresh trees used the unmodified serial recipe: `bash stage3/oracle/run.sh TREE RESULTS --workers 8 --limit-seconds 1800`, followed by the unmodified lane checker. No competing test jobs ran during either oracle. Each command has a 2400-second outer bound and a 3 GiB disk floor. Clean cached pages of completed-job files were evicted before each run without changing file contents, preserving the original proof's strict memory-pressure guard.

| Tree | Raw oracle exit | Lane | Counts | Wall seconds |
|---|---:|---|---|---:|
| control | 1 | PASS | {'passing': 106366, 'failing': 1, 'pending': 0} | 357.944 |
| after | 1 | PASS | {'passing': 106366, 'failing': 1, 'pending': 0} | 357.839 |

An earlier control attempt produced the expected lane PASS but was rejected by the existing resource guard: memory.events max rose from 84015 to 87937, with no OOM kills. Its complete logs, reports and driver are retained under logs/attempt-1-*. The five clean completed worktrees and old d3efff7b artifact listed in cleanup metadata were removed, while reports and pinned source were kept. The control tree was reset to the official final adaptation snapshot, ignored build/test outputs removed, and exact source equality checked before the retry. All clean regular-file cache pages are evicted before each accepted run; the memory-pressure acceptance guard and oracle/checker commands are unchanged.

Both raw exits 1 are the single sanctioned Public APIs failure; each lane checker exits 0. Memory-pressure counters do not increase during either run. Both upstream builds exit 0.
The baseline.diff files match byte-for-byte, SHA-256 `6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`.
The public built typescript.d.ts files match byte-for-byte, SHA-256 `b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`.

Commands, exit codes, disk minima, verdicts and hashes are in results.json; raw logs are under logs/ and driver source is under drivers/ as text. Historical adaptation 93 evidence remains unchanged beside this current proof.
