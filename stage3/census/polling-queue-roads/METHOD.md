# Measurement method and reproduction

Base: origin/main `1a1c0dc18e6546b5861ee56352c542194471ed70`, detached worktree created at `/tmp/polling-queue-roads`, then new report-only branch `typescript/polling-queue-roads`. Shared checkout source was not edited. Worktree bookkeeping and the fetch necessarily update Git's shared object/ref database.

Setup commands:

```sh
git fetch origin main
git worktree add --detach /tmp/polling-queue-roads origin/main
cd /tmp/polling-queue-roads
git submodule update --init --recursive
git clone --filter=blob:none --no-checkout https://github.com/microsoft/TypeScript.git /tmp/polling-typescript
git -C /tmp/polling-typescript checkout 050880ce59e30b356b686bd3144efe24f875ebc8
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/polling-typescript
(cd stage3/api && npm ci)
stage3/apply.sh --build-product /tmp/polling-adapted
```

Recursive submodules initialized successfully, without credential errors. Upstream `git describe --tags --exact-match HEAD` returned `v6.0.3`; the tag's remote object was independently confirmed. Node is v24.19.0; stock TypeScript is the API lockfile's 6.0.3. See `evidence/setup.json` for exact commits.

The README, owners.json and adapt.cjs of 99-correctness-timer-closures were read before constructing variants. `roads.cjs` takes that owner's exact after-text, including `pollScheduled: unknown` and the captured timer closure. Only that whole owner is replaced in scratch sys.ts; pins and adapters are never amended. All variants and full sys.ts inputs are retained under evidence. Full sys.ts copies normalize CRLF to LF uniformly for comparisons; positions by line/column are unchanged in the baseline.

`measure.cjs` uses the repository's existing strict compiler configuration helper, then explicitly resolves pinned Node types from stage3/api/node_modules/@types and uses lib.es2024.d.ts. It checks all 79 compiler roots with stock TypeScript, preserving strict/noUncheckedIndexedAccess/exactOptionalPropertyTypes and the other inherited options. Full effective options and root filenames are in typecheck.*.json. An initial baseline invocation lacked Node types (65 dependency diagnostics); after correcting this setup it was repeated, giving zero diagnostics. That incomplete setup result is superseded, not treated as a road's measurement.

JavaScript is stock `transpileModule` output of the entire sys.ts, ES2024/ESNext, comments retained, consistently for the baseline and roads. The complete emitted modules and unified diffs are retained. This is per-module emitted JavaScript, not a bundled executable or Adamic output.

Census build:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/polling-overlay
go build -buildvcs=false -overlay=/tmp/polling-overlay/overlay.json -o /tmp/polling-census ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE=/tmp/polling-adapted/src/compiler/sys.ts \
/tmp/polling-census /tmp/polling-adapted/src/compiler evidence/census.ROAD.jsonl
```

The unchanged census generator creates its usual measurement-only scratch overlay; no production compiler edit is made. Its first build attempt preceded completion of recursive submodule initialization and failed to find cohere/TypeScript/tsc/go.mod; repeating after initialization succeeded. The no-output guard is enabled on all actual measurements. CommonJSBundle is recorded in every header; the final relation trailer is required by compare.py, so an incomplete stream cannot count as a completed run. Binary and overlay hashes are in census-build-manifest.json. Raw streams retain all attempts and findings. The filter limits attempted files while loading the full 79-root program. The stream's inherited measurement label says “checker-rejected program”; its actual checker_rejected flag and diagnostic arrays are the authority for this run.

`compare.py` deduplicates sys.ts findings by (kind, where, reason, text), including Boundary events, and maps unchanged source lines to the baseline with SequenceMatcher. This keeps purely shifted findings from appearing as new behavior. New or removed edited sites remain explicit. Full raw locations, attempt provenance and source bytes remain available to audit the mapping. No finding is discarded because it is unrelated to the owner.

`witness.cjs` extracts the actual stat/polling helpers from the fully adapted tree and executes each owner through stock-transpiled JavaScript in Node vm contexts. It starts files at all three intervals, uses chunk size two, accelerates unchanged thresholds, modifies all files twice, closes one watcher and fires 65 deterministic scheduled callbacks. It records stats, changes, schedules, queue promotions and changed-list entries; traces are compared by deep equality. Reduced thresholds accelerate queue transitions; this is a deterministic behavioral witness rather than a timing/performance model or a proof for arbitrary schedules. All roads are witnessed, including any rejected by the census, whose Node trace is only supplementary.

`oracle-count.cjs` instruments the original emitted watcher at CommonJS module load, wrapping each newly created queue in a Proxy before its three field initializations. Every string-key property get/set is counted, including push method lookup and the indexed/length accesses made by push and by the poller. Comparisons, local bindings and parameter/return flow are not counted as operations. Factory executions are counted separately. Proxy operations forward with Reflect; no callback, timeout or return value is replaced. Per-process counters are saved on normal Node process exit. The positive control yields three factory calls and fifteen operations for constructing a watcher and adding one low-interval file, including nine initialization writes. The full lane oracle runs on a separate original-tree copy, never on an edited road. The operation count is coverage evidence, not a road's performance estimate.

The initial oracle copy incorrectly excluded directories named node_modules at every depth, omitting 17 test fixture directories. Project-baseline failures exposed this measurement error; that oracle was stopped and superseded, and its counters/verdict are not used. The final oracle uses a fresh complete copy (only the root .git metadata excluded, symlinks preserved), with every input file verified byte-for-byte against the original adapted tree before npm ci. See oracle-copy-correction.json and oracle-input-manifest.json.gz for the correction and full verified input inventory. Census/type-check/emission/witness measurements used the complete adapted tree throughout and were unaffected.
