# Adaptation 93 recut and lane input hashes, #he8tysb

Main: `470e16f8315f7a83c5c7b7f6f2a77d8328db0a25`. Source: `cb48b53696f9e027aac94193cb257542880b2591`.
The adaptation 93 directory is byte-identical to the source commit. This is one new commit on main.

## Official patch table

`bash stage3/apply.sh /workspace/scratch/adapt-93-main-report/table-tree --write-table`
regenerated the table, without a textual merge. Counts are files / added / removed:

| Row | Count |
|---|---|
| 40-explicit-any | 18 / 107 / 107 |
| 93-builtin-imports | 1 / 3 / 3 |
| Main control total | 85 / 5559 / 5517 |
| With-93 total | 85 / 5562 / 5520 |

Main's table was stale since `141ad1d2`: its `tsbuildPublic.ts:424` timer slot adds the eighteenth file and the 107th added and removed lines to row 40. Official generation also includes main's existing 96a and 96c rows.

## Input-hash proof

The shared artifact was built with `python3 stage3/lane/prepare_shards.py /workspace/scratch/adapt-93-main-report/artifacts`.
Its final key is `d821beb196821d5e90426235d234432ff850ba06f9936f9579ee1874cbcd94a3`.

The old manifest first failed in the shared-artifact guard:

```json
{"name": "stage3-lane-file-000", "status": "fail", "counts": {}, "seconds": 2.301464362999468, "errors": ["shared artifact inputs hash does not match this checkout and toolchain"]}
```

The exact regeneration commands were:

```sh
python3 stage3/lane/shard_plan.py /workspace/scratch/adapt-93-main-report/artifacts/d821beb196821d5e90426235d234432ff850ba06f9936f9579ee1874cbcd94a3
python3 stage3/lane/generate_shards.py
```

These are the repository generators for the same three files changed by `470e16f8`.
All 398 definitions, selected test identities, expected counts, partition settings, API digest and runtime digest remain identical; every unit hash changed. The generator's compressed identity listing decompressed byte-identically, so its gzip timestamp-only rewrite was restored.
The old hash against the new manifest is rejected (exit 1):

```json
{"name": "stage3-lane-file-000", "status": "fail", "counts": {}, "seconds": 0.016821453000375186, "errors": ["unit inputs hash differs from the manifest"]}
```

The manifest's new hash was passed verbatim to `shard.py`:

```sh
python3 stage3/lane/shard.py stage3-lane-file-000 /workspace/scratch/adapt-93-main-report/unit-final --inputs-hash 30e9f434ae662b4210bfac9571a123f514d3fe1cc055e638e9a53ac3a649833c
```

Exact output:

```json
{"name": "stage3-lane-file-000", "status": "pass", "counts": {"pass": 6, "fail": 0, "skip": 0}, "seconds": 16.518768435998936, "errors": []}
```

The full oracle's mutable with-93 tree is preserved separately in `/workspace/scratch/adapt-93-main-report/oracle-after-artifact/tree`. After the oracles, the shared artifact was rebuilt by the official producer at the same key. The success line above is the final unit run on that restored artifact, followed by an official cache-hit verification of its complete payload. The archived oracle view also passes the lane checker again.

The first cleanup check left cohere uninitialized and go list rejected its missing go.work module. It is now a proper checkout at 06af2e399c6c1fb93fd570b774027942747400b4, with TypeScript at d92d9bfee114c80be2c375d72edae966176e3a4f. A shared clone attempt hit the source repository's partial-history limit; existing immutable objects were hard-linked into private repositories and their complete pinned trees checked out instead. go list and a final real Go manifest-shard test both pass with that clean dependency layout. Logs retain the initial setup failures and the successful checks.

## Whole lane and adapter guards

`go test -json -count=1 -timeout 60m ./stage3/lane/...` exited 0:
418 top-level pass, 0 fail, 0 skip; wall 731.577 seconds.
The real `TestProduct_Stage3AdaptedTree` was built ahead, not substituted with a fake.

Both full-source hash pins remain unchanged:

- Before: `5c45814e8fb6155701df1ef9be2755096ccac593ac6d53b13a4d26d90c7284d8`.
- After: `be2c80c35536b9d0a05a66df61b5f229155325353caf0718da64f459754ea121`.

The rerun reports zero files and preserves sys.ts bytes and nanosecond mtime.
Appending one newline is rejected with `target source moved: sys.ts` before any write; bytes and mtime of the moved target remain intact. Source scope, exact undo, comments-kept and comments-removed transpile comparisons and binding/collision checks all pass.

## Fresh serial oracles

No previous control was reused. Between `7b9c0fe7` and main, adapters 96a and 96c add edits to transformer.ts and transformers/generators.ts. The three lane input-hash files also changed; apply.py and the lane harness did not. sys.ts did not change, so 93's pins need no update.
Before each oracle, clean cached pages of completed-job files were evicted with posix_fadvise; no file contents changed. Both fresh trees used the unmodified `stage3/oracle/run.sh TREE RESULTS --workers 8 --limit-seconds 1800`, followed by the unmodified lane checker. The runs were serial, with no other test jobs active; each had a 2400-second outer bound and a 3 GiB free-space floor.

| Tree | Upstream exit | Lane | Pass / sanctioned API fail / skip | Wall seconds |
|---|---:|---|---|---:|
| Main | 1 | PASS | 106366 / 1 / 0 | 353.721 |
| With 93 | 1 | PASS | 106366 / 1 / 0 | 337.954 |

Each raw upstream exit 1 is the one sanctioned Public APIs failure. The lane checker exits 0.
The baseline.diff files match byte-for-byte: SHA-256 `6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`.
The public built typescript.d.ts files match byte-for-byte: SHA-256 `b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`.
The detailed verdicts, commands, exit codes, disk minima and unchanged memory-pressure counters are in results.json; raw command and oracle logs are compressed under logs/.

This evidence lives under lane/, which artifact.py excludes, so recording the proof does not create an input-hash self-reference.
