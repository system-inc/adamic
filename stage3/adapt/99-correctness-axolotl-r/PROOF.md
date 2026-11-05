# 99r proofs

The accepted change has four line replacements in three exact pinned core.ts
owners. No production compiler, other adaptation, or checker caller changes.
The complete applied src tree and the ordinary census snapshot match across all
746 files. Only core.ts differs between before and after, and precisely by the
three owner replacements in owners.json. Both checker.ts and the six-caller
README section remain byte-identical.

| Check | Result |
| --- | --- |
| Prescribed setup and GOPROXY | Exit 0; setup.log.gz retained from the preceding turn |
| Baseline full apply | Exit 0, retained apply-before.log.gz; all 746 source bytes verified unchanged by baseline-source-identity.json |
| Accepted full apply, including product packaging | Exit 0; incremental row is 1 file, 4 added lines, 4 removed lines |
| Actual full-apply replay | Zero files, edits and writes; bytes and mtime preserved |
| Owner contract | 24 refused fixtures before writes; foreign edit accepted; LF/CRLF preserved |
| Own operand types | never[], T[], never[], T[]; none any or unknown; exact unknown-hop shape |
| Stock extracted four-mint witness | Zero diagnostics under TypeScript 6.0.3 |
| Full-stock diagnostic delta | Zero new diagnostics; all 59 ambient-context diagnostics retained in audit-result.json |
| Adamic types over 79 recursive roots, before | Exit 0, zero diagnostics |
| Adamic types over 79 recursive roots, after | Exit 0, zero diagnostics |
| Fresh ordinary compiler builds, before and after | Both exit 0 |
| Whole core.ts JavaScript, comments retained and removed | Byte-identical in both modes |
| All 9 normally emitted JavaScript files | Byte-identical |
| typescript.d.ts | Byte-identical; SHA-256 c4edc2ad37cf1afeb4704ccb4bacf7c76b9f538a8e5685e4b276e2e7dd52796b |
| tsserverlibrary.d.ts | Byte-identical; SHA-256 cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd |
| Syntax counts in core.ts | Casts 27 to 27; any 18 to 15; unknown 19 to 23 |
| generate_shards --check | Exit 0 |
| Real sort | Directly after 99-correctness-axolotl-q |
| Full core site census | Same 82 distinct kind/location/reason findings before and after; no added or removed findings |
| Stage3 lane Go fixtures | Exit 1 even isolated: both API-generator wrappers exceed their unchanged 55-second bounds |
| Underlying API-generator Python suite | All 22 pass, 111.715 seconds; supplemental, not a Go-suite pass |
| Ordinary 79-root next-stop census | All 79 roots: 73 Refused, 4 NotYet, 2 accepted; CENSUS.md |

The preceding direct-cast stop is preserved in initial-direct-cast/. It is not the
accepted adapter. Its stock TS2352 does not apply to the newly authorized unknown
hop, whose four-expression clean witness is checker-clean. The actual stock
program preserves all ambient diagnostics rather than inventing or suppressing
host declarations. The independent Adamic types commands use the real embedded
host overlay and are zero before and after.

The fresh before/after builds run npm ci from the same upstream lock and
npm run build:compiler. Both public declarations are freshly emitted and bundled;
no stale or copied declaration output is compared. comparison.json records all
hashes, byte counts, source scope, AST counts and emitted-file comparisons.

The baseline full apply was completed in the preceding turn. Its whole-src
aggregate hash is still aa9e6257ac06527958898df1a0759488b83fc5b37ce79da5c6a07f025f48fb61
after the fresh baseline build. The after full apply runs every adaptation from
the pinned upstream commit. Its source is compared byte-for-byte to the census
snapshot, rather than using a partial-root fallback or a selectively adapted tree.

The ordinary census binary uses unchanged production Load and Lower. The
compiler source paths equal the underlying b729c7e0 stack. compiler-provenance.json
records the binaries, adapter pins and upstream. No Diagnostics lowering change
is pulled into this source-only task. The complete core site measurements use
the repository's existing no-output scratch overlay only, with LATENT_FULL=1,
LATENT_ONLY_FILE selecting the entire core.ts, and LATENT_ASSERT_NO_OUTPUT=1.
Ordinary Load, Lower and output are disabled in that measurement binary. It is
never used for the next-stop census. Its historical measurement-label text is
retained; the actual before/after metadata reports checker_rejected=false and
zero checker diagnostics. site-results.json preserves all mint reasons and the
unchanged relation at core.ts:1039:34. Raw streams preserve duplicates and units.

The initial lane run failed three unchanged 55-second wrappers while full apply,
site scans, builds and the ordinary census shared the worker: TestLaneShardsPython,
TestLaneGenerateApiRejectionsPython and TestLaneGenerateApiReconstructionPython.
lane-concurrent.log.gz retains the failure. A second run with only the census
competing failed API reconstruction at 55 seconds, retained in
lane-with-census.log.gz. An isolated third run, with all other proof processes
finished, still failed both API wrappers at their unchanged 55-second bounds;
lane-isolated.log.gz retains its exit-1 result (235.835 seconds). Concurrency alone
does not account for the final failure. The lane Go check remains failed.

The underlying test_generate_api suite was then run directly to cover the cases
the wrappers had not finished: all 22 pass in 111.715 seconds, with no assertions
removed. api-direct.log.gz retains the supplemental result. This is not reported
as a successful Go suite. No deadlines, test selection, wrapper code, or other
adaptation was changed. The branch is handed back with this explicit proof gap.

Commands, from the repository with /workspace/adamic-tools/env.sh sourced:

```sh
export GOPROXY='https://proxy.golang.org|direct'
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/r2-after
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-r/replay.cjs /workspace/r2-after
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-r/contract.cjs /workspace/r-before
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-r/audit.cjs /workspace/r-before
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-r/compare.cjs /workspace/r-before /workspace/r2-after /workspace/r2-probe
python3 stage3/lane/generate_shards.py --check
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
go build -o /workspace/scratch/r2-census ./stage3/census/tool
/workspace/scratch/r2-census /workspace/r2-probe/src/compiler /workspace/scratch/r2-census.jsonl
```

The supplemental direct Python suite runs separately from stage3/lane:

```sh
cd stage3/lane
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source python3 -E -S -B -m unittest test_generate_api
```

All logs are retained in files.

Both types commands pass all 79 sorted recursive .ts paths to the unchanged
adamic binary together; result JSON retains every path and status. The ordinary
census inspects each entry independently and retains its extra non-source and
whole-program rows too. The successful type-only proofs do not claim successful
lowering. The native oracle is skipped under the same type-only exception as 99k
through 99q.
