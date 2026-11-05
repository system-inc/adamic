# 99q proofs

All four kept source edits are type-only. The only changed source file is
core.ts, with exactly one changed line in each of the four pinned owners.
The normal compiler emits identical JavaScript and public declarations.
comparison.json retains hashes, byte counts and exact source-scope assertions.

| Check | Result |
| --- | --- |
| Prescribed setup, GOPROXY=https://proxy.golang.org\|direct | Exit 0, 522.384 seconds |
| Baseline full apply from pinned TypeScript, with product packaging | Exit 0 |
| Accepted stack full apply with product packaging | Exit 0 |
| Direct adapter replay after full apply | Exit 0; zero files and edits |
| Owner contract | Pass: 29 refused fixtures before any write, foreign selectIndex edit accepted, LF/CRLF preserved |
| Contract replay | Zero writes; source bytes and mtime preserved |
| Adamic types, all sorted 79 roots, before | Exit 0, zero diagnostics |
| Adamic types, all sorted 79 roots, after | Exit 0, zero diagnostics |
| Ordinary compiler build, before and after | Both exit 0 |
| Both public declarations | typescript.d.ts and tsserverlibrary.d.ts byte-identical |
| Whole core.ts transpile, comments retained and removed | Byte-identical in both modes |
| All nine normally emitted JavaScript files | Byte-identical |
| Complete applied src versus census src | All 746 source files byte-identical |
| Added casts / any | None; total cast nodes 28 to 27 and any nodes 19 to 18 |
| generate_shards.py --check, final | Exit 0 |
| Stage3 lane Go fixtures, initial | Pass, 116.954 seconds |
| Stage3 lane Go fixtures with 99q present | Pass, 103.311 seconds |
| Direct API-generator Python tests | All 22 pass, 71.245 seconds |
| Ordinary main per-file compiler census | All 79 roots; see CENSUS.md |

Commands, with the tool environment sourced from /workspace/adamic-tools/env.sh:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/q-before
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/q-after
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-q/adapt.cjs /workspace/q-after
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-q/contract.cjs /workspace/q-before
npm run build:compiler
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-q/compare.cjs /workspace/q-before /workspace/q-after /workspace/q-probe
python3 stage3/lane/generate_shards.py --check
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
```

The npm build command ran separately in q-before and q-after, after each full
apply's packaging had completed. The 79 sorted paths under src/compiler were
passed together to the ordinary binary's `types` command before and after.
The same binary's ordinary `c binder.ts` reproduces the baseline deduplicate wall.

The two Go API sanction wrappers both passed on this four-CPU-quota worker.
Their unchanged underlying suite was also run directly in stage3/lane:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source python3 -E -S -B -m unittest test_generate_api
```

The first direct invocation omitted ADAMIC_TYPESCRIPT_SOURCE and failed all
22 setup checks immediately. Its log is retained as api-direct.log.gz. The
configured invocation passes all 22 tests; api-direct-configured.log.gz retains
that result. No test, deadline or selection was weakened. There were no 55-second
wrapper timeouts on this worker.

The site census uses the repository's existing measurement-only overlay, with
LATENT_FULL=1, LATENT_ONLY_FILE naming the complete core.ts, and
LATENT_ASSERT_NO_OUTPUT=1. It scans every core.ts owner, not just the edited one.
The complete before and trial streams, including unit/dependency context, are
retained. alone-summary.json compares each proposal by itself with the baseline;
sites-summary.json compares each trial with the previous accepted stack.
H02's calls-only and larger return/map trials are retained too. The no-output
measurement binary is never used for the final ordinary per-file census.

Trial patches and complete-file source hashes are in trial-sources.json. These
are evidence hashes, not whole-file adaptation gates. Runtime gates pin only
the four accepted owners. The two dropped sites and the unedited or scout are
not pinned. The final adapter changes no skipWhile, clear, filterMutate or
createQueue owner. The entire source-tree comparison enforces this scope.

The rejected concatenate trial has six Adamic checker errors. Its ordinary
build fails too. Its freshly emitted internal declarations were independently
rebundled with upstream scripts/dtsBundler.mjs to check the public declaration
impact without mistaking stale copied output for a new result. Both public
files match the baseline; concatenate-public.json records the hashes. The
Node/stock-checker witness runs independently of the failing whole compiler
build. No public declaration diff is hidden.

Compiler provenance is frozen in compiler-provenance.json. The ordinary compiler
is unmodified d40c2470e. Its relevant source tree equals the deliverable base.
A later fetch observed 006b368745e6a37f18b1427b1a63a78285f92694; the retained
main-recheck.diff.gz shows generator-to-Map/Set work, without a fix to these
type-only relations. This later compiler was not used for the census.

The lane oracle is skipped under the requested type-only exception. No runtime
behavior change is claimed. The new declaration-only concatenate correction
is deliberately not included because its callers fail the required contract.
