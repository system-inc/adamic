# 99n proof

All required checks pass for the single kept site, H10 / CW157. The other 25 requested sites are dropped and remain unchanged. See [sites.json](sites.json) for every ID, [trial-deltas.json](trial-deltas.json) for independent proposal effects, and [CENSUS.md](CENSUS.md) for every declaration's next stop.

The final compiler base is main `ab8a85bebf2a14a06fa8c3c4ca8cc93201a2309f`. Work started on `529b6fdb427a7fa7ccd9bdf58d4347e72e49f16c`; main advanced with three Lower commits. The branch was fast-forwarded, both proof binaries were rebuilt, and all six census measurements plus the individual checker trials were rerun. No stage3, setup, source pin or Go workspace input changed between those main commits. `compiler-provenance.json` records the exact commits and binary, adapter and pin hashes.

| Check | Result |
| --- | --- |
| Requested timeout-wrapped setup with GOPROXY=https://proxy.golang.org\|direct | Exit 0; 1142.992 seconds |
| Full stage3/apply.sh --build-product | Exit 0; real sorted adaptation sequence and complete before/after refs |
| 99n patch table | One file, one line added, one line removed |
| Ordinary keyed replay to a new tree | Exit 0; all 81,504 non-Git payload entries byte-identical |
| Direct adapter replay on full output | Zero files, edits and writes; bytes and mtime preserved |
| Owner contract | Foreign edit accepted; changed-before, partial-signature and changed-after owners refused before any write |
| Adamic types | Zero diagnostics before and after, all 79 roots; actual shipped full-apply output also exits 0 |
| Stock TypeScript 6.0.3 | Zero diagnostics for baseline, H10, and every independently tested proposal |
| Normal TypeScript local builds | Both exit 0 |
| All 9 emitted JavaScript files | Byte-identical |
| Both public declarations | typescript.d.ts and tsserverlibrary.d.ts byte-identical |
| Entire debug.ts transpile | Byte-identical with comments kept and removed |
| Source and syntax comparison | Only debug.ts changes among 746 sources; 47 casts on each side, zero explicit any types on each side |
| Selected-file census | 105 debug.ts units measured against the complete 79-root program, full mode, no-output assertions |
| Real census pair against main coordinates | One removed finding/read; zero new findings, sites, first-stop changes, regressions or panics |
| generate_shards.py --check | Exit 0 on the final directory |
| Stage3 lane Go fixtures | Exit 0; 181.605 seconds with GOMAXPROCS=1 and -parallel=1 |
| Real adaptation sort | 99n after 99j and before 99-wasm on main; fixture with 99m puts 99n immediately after 99m |

The public declarations are 593,502 and 866 bytes. Their SHA-256 values, all nine JavaScript hashes, both debug.ts transpile hashes, source counts and syntax counts are in `comparison.json`. `full-source-comparison.json` links the real full-apply output to that measured H10 source. No public API sanction is added. A runtime oracle was unnecessary because every emitted JavaScript byte is unchanged.

The full apply and keyed replay ran with the frozen adapter input set before reporting helpers and archives were added. The executable adapter and its active pins have not changed since that apply. The ordinary replay restores a new keyed product; the separate direct replay proves the adapter itself makes no write.

The exact 79-root checker and census command lists, exits and elapsed times are in `measurements.json`. `types-before.log.gz`, `types-after.log.gz`, and `types-shipped.log.gz` retain checker output. Six independent raw census streams are retained, with separate stock-checker JSON results. The census uses the repository's unmodified generated scratch overlay, `LATENT_FULL=1`, `LATENT_ONLY_FILE` selecting debug.ts, and `LATENT_ASSERT_NO_OUTPUT=1`. It does not enable speculative mode or claim native compilation. The coordinate projection used by the real pair tool is explicit and audited in CENSUS.md and `projection.json`.

The lane command was:

```sh
GOMAXPROCS=1 go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
```

ADAMIC_TYPESCRIPT_SOURCE pointed to the exact pinned upstream checkout. Two earlier runs overlapped compiler or census work and hit existing 55-second Python wrapper deadlines. Their logs are preserved in `lane-concurrent.log.gz` and `lane-serial-earlier.log.gz`. The passing final run ran after all compiler and census work completed. No test selection, timeout, baseline, assertion or expected result changed. `lane.log.gz` retains that pass.

Reproduce the adapter contract with NODE_PATH set to stage3/api/node_modules:

```sh
node stage3/adapt/99-correctness-axolotl-n/contract.cjs BEFORE_TREE
python3 -B stage3/adapt/99-correctness-axolotl-n/pair.py BEFORE_JSONL AFTER_JSONL BEFORE_TREE AFTER_TREE OUTPUT_DIRECTORY
```

Logs and disposable trees remain under `/workspace/scratch/99n-*`; setup output is `/workspace/scratch/setup.log`. Only this new adaptation directory is committed.
