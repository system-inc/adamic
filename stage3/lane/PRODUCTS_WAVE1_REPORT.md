Built four apply products, three offline lane products and exactly one Go shard declaration.
Branches: `typescript/stage3-apply-products` at `ab06c9886ccba3822d9e95effde1a7323e6a13b0`; B code at `26ba1d5d17501f9a7cd14072dad20503e8714540` on `typescript/stage3-lane-products` (final report-commit SHA accompanies delivery).
Cold lane units: 10.93 / 31.43 / 22.81 s; shard 001: 273/0/0 in 15.54 s; producer comparison: 143.365293 s, exit 0.
All 18 explicit mutants were caught; shard 002 remained green under shard 001's owned-file break.
Not covered: the other 150 Go shard declarations, the full unsharded oracle/union, remote store publication/fetch or other platforms.

## Scope and environment

Part A was cut from `origin/main` at d072c133 and re-cut from 8dd02d807a,
2888e54540 and ad37d7381f. Main's five newer apply changes win: supplied pinned
source, depth-one local clone, installed byte-pinned API packages, os.walk,
build-cache placement and no bytecode writes. Details and complete A evidence
are in [the apply report](../APPLY_PRODUCTS_WAVE1_REPORT.md).

The final B branch contains final A through an ordinary merge, with no history
rewrite or force push. Main advanced to
`87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4` during work, only outside stage3.
Both changes were also tested in temporary merge candidates on that tip.
Only stage3 paths are changed. No cloud/fast-gate edits, PR or Loom submission.

Linux x64, `nproc=5`, cgroup `cpu.max=400000 100000` (4 CPUs), Go 1.27.1,
Node 24.19.0, npm 11.9.0, Python 3.12.14, git 2.52.0, clang 20.1.8.

| Setup / supplied input command | Exit | Output |
|---|---:|---|
| `export GOPROXY='https://proxy.golang.org\|direct'; timeout 600 bash cloud/setup.sh` | 0 | Go ready 0.069 s; Node 0.072; clang 0.467; markdown install step 0.715, ready 0.909; submodules 167.505; shared cache 168.199; Go build 429.989; tests deferred 430.073; cache warm 430.074; done 430.110 s |
| `timeout 120 git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/hefhych-typescript` | 0 | Pin 050880ce59e30b356b686bd3144efe24f875ebc8 |
| `timeout 90 npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund --cache /tmp/hefhych-npm-cache` | 0 | 3 packages |
| `timeout 120 npm ci --prefix /tmp/hefhych-typescript --no-audit --no-fund --cache /tmp/hefhych-upstream-npm` | 0 | 338 packages, 10 s; outside every product |

Every build/test shell sourced `/workspace/adamic-tools/env.sh`. All product
measurements used `ADAMIC_TYPESCRIPT_SOURCE=/tmp/hefhych-typescript`,
`ADAMIC_BUILD_CACHE_DIR=/tmp/hefhych-cache-A` and remote-store reads disabled.
The supplied install is a declared offline gate input, verified against
`lane/node_modules.sha256`:
`87a89f1917d4d7d084df0c34e698dc23d314928551203f9c17b7858f9ff090a9`.
This Linux x64 pin includes the installed hidden lockfile, package bytes,
executable bits, directory names and symlink targets. It is not a network
install timed inside a product. Provisioning that input in Loom is outside
this source-only change; no runner needs that install once the artifact exists.

## Part A: apply

Each NN command was a fresh Go invocation, with that product missing and its
predecessors prepared. Go/toolchain and OS caches were not flushed. The table
reports the entire top-level test wall, including its fetch/key/setup work;
Go's external compile/link startup is not part of the test-function budget.

`timeout 90 go test ./stage3/applyproducts -run '^TestProduct_Stage3AdaptedNN$' -count=1 -timeout 90s -v`

| NN | Test wall s | Builder log s | Exit |
|---|---:|---:|---:|
| 10 | 18.60 | 18.52 | 0 |
| 40 | 39.70 | 39.52 | 0 |
| 70 | 47.66 | 47.46 | 0 |
| 99 | 23.41 | 23.20 | 0 |

The independent `prove_tree.py` ran main's original plain builder, 100.461376 s,
and hashed every non-.git entry, sorted: path, kind, full mode, symlink target,
file bytes and empty directories. Both whole-tree hashes were:

`b122537d07680ddad789bf9e543ac807efdeb6fe6748b49d325e83764306901b`

| A command | Exit | Observed result |
|---|---:|---|
| `timeout 90 go run ./stage3/applyproducts/fetch` | 0 | Returned product 99 |
| `timeout 240 python3 stage3/applyproducts/prove_tree.py /workspace/scratch/hefhych-main-apply PRODUCT99` | 0 | Equality, then rejection of a real appended byte sequence |
| `timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch -run 'TracksInputs\|Deadline' -count=1 -timeout 90s -v` | 0 | Four key checks and process-group deadline proof |
| `timeout 90 python3 stage3/applyproducts/prove_inputs.py` | 0 | Four intended Go failures, each `-timeout 90s` |
| `timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch -count=1 -timeout 90s -v` on committed A | 0 | Whole changed packages |
| `timeout 90 env -u PYTHONDONTWRITEBYTECODE python3 -B stage3/test_apply.py` on corrected committed A | 0 | 25 fixtures, 13.951 s |
| `timeout 90 go vet ./stage3/applyproducts ./stage3/applyproducts/fetch` | 0 | No diagnostics |
| Prescribed root lane checks on corrected A | 0 | 1.5 s; 3 Go files, 1 test package, vet 1 package |

## Part B: lane products

Install fetches only apply 99 and validates the supplied install's lockfile,
versions and committed byte pin before making independent copies. Compiler
fetches only install; harness fetches only compiler. A missing predecessor
fails instead of building it. Shard consumers fetch only harness and cannot
build any product. `npm run build` also builds upstream tests; the explicit
harness step completes that producer phase and packages the result.

Keys bind the same broad stage3 source set used by artifact.input_key, the
lane recipes and identity helper, the installed-package byte pin, execution
settings and tool versions. API installed bytes remain bound by the existing
api/node_modules.sha256. The three TracksInputs fixtures preserve modes and
symlinks when relocating their inputs and reject a changed real worker file.

Current buildcache publishes regular-file blobs. The only installed links are
.bin executable aliases; install replaces them with relative shell wrappers
which exec the original target, preserving that target's module resolution.
No compiler or test source is changed. Both upstream builds actually used
these wrappers. The finished product has no symlinks.

The final directory itself contains ready.json, tree/, api/ and
shards-ready.json, plus artifact.tgz, phase metadata and an empty npm config.
Only successful phase exits are embedded; measured timings stay in the logs.
The archive is sorted with normalized ownership/mtime and gzip -1n. It is an
ordinary stored buildcache blob beside the directly usable regular-file layout.

Every product invocation below was fresh, its own product absent and its
predecessor prepared, with no other build running:

`timeout 90 go test ./stage3/lane -run '^TestProduct_Stage3LanePHASE$' -count=1 -timeout 90s -v`

| PHASE | Test wall s | Builder log s | Exit |
|---|---:|---:|---:|
| Install | 10.93 | 10.70 | 0 |
| Compiler | 31.43 | 31.09 | 0 |
| Harness | 22.81 | 22.46 | 0 |

`unshare -n true` exited 1, Operation not permitted. The logs show no source or
package fetch. Install is a checked local copy. Compiler and harness set npm's
offline mode, use empty user/global configs, and put npm's transient cache
inside their scratch product, removing it afterward. No product writes HOME.
Source/package preparation above is explicitly outside the products.

### Identity adjustment

shard.py is byte-for-byte unchanged. Its original input, payload, API, runtime,
inventory, baseline, count and elapsed-time checks all still run. The smallest
identity change is in artifact.py: exclude physical api/node_modules, which is
already verified against its byte pin, and explicitly bind the committed
upstream installed-byte pin. This avoids the runner's extra API lockfile marker
and lets consumers verify identity without a second upstream install. Install
rejects changed package bytes even when their versions and hidden lock match.

`timeout 90 python3 -B stage3/lane/bind_products.py ARTIFACT`, exit 0, only
rebound existing artifact/runtime/unit hashes. All 151 memberships and expected
counts are unchanged. No additional Go shards were generated.

Final artifact input key:
`5c7fd1a63052cc4fe90daf07589cb7b09864a3d90f9c49f1343a980c73eb0d7f`.
Runtime: `8a9f89bc3fcbd760f04ebbcbca1990ef150d844621816275c7554d01f833513c`.
Stock API: `1512f1871eebe448dac94546c5a6337daa1cb64ff20c80f3d9e93f85a0f48dce`.

## Part C: exact sizes and reuse

Logical regular-file bytes, including .git inside tree/, counted directly from
the measured directory. Directories are not counted as files.

| Measured object | Bytes | Files |
|---|---:|---:|
| Required artifact payload: tree/, api/, ready.json, shards-ready.json | 756,522,976 | 92,206 |
| Stored artifact.tgz | 197,009,214 | 1 |
| Complete buildcache product, including archive and 2 metadata/config files | 953,532,283 | 92,209 |

Current main stores per-file blobs, not an implicit archive. The archive size
above is the actual archive file produced and stored by this recipe, not an
estimate of a future transport. The layout's files are also stored; its two
additional files total 93 bytes. Symlink count: zero.

Archive SHA-256:
`5b5c2a537911e240f0e2e5acb3587dc8488813b9ac6bc3e415067daa21e1190a`.
Final product directory:
`/tmp/hefhych-cache-A/local/9987ccd59bf2f7038386b08ea2f4390b9b88d151e33d5af1b8eb6e0412a3b4aa`.

| Reuse / comparison command | Seconds | Exit / output |
|---|---:|---|
| Second fresh `go test ./stage3/lane -run '^TestProduct_Stage3LaneHarness$' -count=1 -timeout 90s -v` | 1.269482 complete process; 0.44 test; 0.10 final lookup | 0; compiler and harness both log hit, no build |
| Original main `prepare_shards.py`, fresh apply/store/artifact directories | 143.365293 complete process; 143.290 producer report | 0; cache_hit=false |

Reuse process time was measured with Python time.monotonic around subprocess.run,
with a 95 s outer limit. The original producer was run from a clean main
reference worktree on the same instance, using the same source and API bytes,
prepared offline npm cache and no concurrent build:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/hefhych-typescript \
ADAMIC_BUILD_CACHE_DIR=/workspace/scratch/hefhych-baseline-build-cache \
STAGE3_CACHE=/workspace/scratch/hefhych-baseline-apply-cache \
STAGE3_PRODUCT_STORE=/workspace/scratch/hefhych-baseline-store \
npm_config_cache=/tmp/hefhych-upstream-npm npm_config_offline=true \
timeout --kill-after=5 300 python3 -B stage3/lane/prepare_shards.py /workspace/scratch/hefhych-baseline-artifacts
```

Original producer phase measurements: apply 109.560 s; locked install 2.288 s;
build 21.839 s; harness 0.874 s. All exits 0. Its verification/packaging work
also counts in the complete producer time. This is a cold original plain-apply
comparison, not a warm invocation of the new chained apply CLI.

## Part D: one Go shard

`timeout 90 env -u ADAMIC_TYPESCRIPT_SOURCE go test ./stage3/lane -run
'^TestShard_Stage3LaneFile001$' -count=1 -timeout 90s -v`, exit 0:
**15.54 s** top-level wall, **15.185743 s** reported by shard.py, counts
**273 pass / 0 fail / 0 skip**. The final product was already built.

The test loads the manifest from disk, selects exactly one named unit, runs its
entire argv from the repository root, sets STAGE3_ARTIFACT and a fresh
t.TempDir() STAGE3_RESULTS, and requires exit 0, status pass and matching counts.
It disables bytecode writes through the environment without retyping argv.
Missing, empty or duplicate selections cannot silently pass.

The owned-file mutant modifies only a private copy of
src/testRunner/unittests/tsserver/projectReferencesSourcemap.ts. One runtime
throw is inserted into its verifyAction callback; the private harness is
rebuilt successfully, then passed through shard.py's existing --mutant-runner
proof option. The product itself is never modified.

| Final mutant command / observation | Exit | Result |
|---|---:|---|
| Private `npm exec --offline hereby -- tests` | 0 | Owned-source mutant compiles; not a build-error kill |
| Manifest argv for file-001 plus private --mutant-runner | 1 | 264 pass / 9 fail / 0 skip; count, failure-title and upstream-exit checks reject it |
| Manifest argv for file-002 plus the same runner | 0 | 6 pass / 0 fail / 0 skip; status pass |
| Actual Go shard with shards.json absent | 1 | Missing file error |
| Actual Go shard with version 1 and empty units | 1 | missing version 1 units |
| Manifest command exits 0 but records 272 passes | 1 | Go's recorded-count comparison rejects it |
| Manifest command records correct counts but exits 1 | 1 | Go's command-exit check rejects it |

Every command has a 90 s deadline. `timeout 180 python3 -B
stage3/lane/prove_shard.py ARTIFACT` exits 0 only when all intended outcomes
hold and the final product's payload hash remains unchanged. Source hashes,
full observations and failures are preserved in its private proof directory
and compressed evidence.

## All explicit mutants

| Mutant | Catcher / nested exit |
|---|---|
| A10 drops source.json | TestStage3Adapted10TracksInputs, changed source pin reused product key; 1 |
| A40 drops source.json | Corresponding TracksInputs assertion; 1 |
| A70 drops source.json | Corresponding TracksInputs assertion; 1 |
| A99 drops source.json | Corresponding TracksInputs assertion; 1 |
| Append real bytes to reference core.ts | Independent whole-tree inequality; proof script 0 for successful rejection |
| Kill only apply worker parent | TestProductDeadlineKillsDescendants, child still running; 1 |
| Force --write-table | ApplyCheckoutTests, ordinary apply dirtied the source checkout; 1 |
| Install drops products.py | TestStage3LaneInstallTracksInputs, changed lane worker reused product key; 1 |
| Compiler drops products.py | Corresponding TracksInputs assertion; 1 |
| Harness drops products.py | Corresponding TracksInputs assertion; 1 |
| Drop upstream package identity | ArtifactTests.test_upstream_installed_bytes_invalidate_key; 1 |
| Re-include physical API install / marker | ArtifactTests.test_installed_parser_is_keyed_by_its_byte_pin; 1 |
| Skip supplied package byte-pin validation | Real changed installed bytes fixture, ValueError not raised; 1 |
| Owned-file throw, missing manifest, empty units, wrong counts, nonzero exit | Five D mutants in the table above; all intended red, control 002 green |

`timeout 150 python3 -B stage3/lane/prove_products.py` exits 0 only after
all six B mutants fail for their intended assertions. Each Go invocation has
-count=1 -timeout 90s; each Python fixture is bounded at 90 s. Every recipe,
manifest and reference-byte mutation is restored with finally cleanup.

## Final validation, limitations and corrections

The following commands ran from the repository root with logs, each bounded:

| Command | Exit | Output |
|---|---:|---|
| `timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch ./stage3/lane -count=1 -timeout 90s -v` | 0 | Whole packages, including all seven products, keys and the one shard |
| `timeout 90 python3 -B -m unittest discover -s stage3/lane -p 'test_*.py' -v` | 0 | 71 fixtures, 10.931 s |
| `timeout 90 env -u PYTHONDONTWRITEBYTECODE python3 -B stage3/test_apply.py` | 0 | 25 fixtures, 13.657 s |
| `timeout 90 go vet ./stage3/applyproducts ./stage3/applyproducts/fetch ./stage3/lane` | 0 | No diagnostics |
| `git diff --check` | 0 | No diagnostics |

Committed-branch lane checks use the prescribed root command under timeout 90:
`git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show
origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`.
The initial B code check passed in 1.5 s: gofmt/tools on 5 Go files,
parallel analyzer on 2 packages, vet 2 packages. Final tip checks and the
current-main merge candidate are recorded in the delivery addendum below.

Failed or superseded work is not hidden:

- Initial A tests could not load the unfinished setup submodule; retry passed.
  Setup-overlapped A40 took 71.96 s and is a sizing failure. All final A
  measurements were repeated after setup finished.
- Obsolete A cache copies filled the separate 8.8 GB /tmp tmpfs, causing an
  equality and a worktree setup failure. Only task-owned obsolete builds were
  removed; fresh equality and fixtures passed. Workspace has a separate 32 GB
  volume and holds the reference and private proof trees.
- Initial B key relocation fixtures flattened evidence symlinks, so three
  checks failed. They now preserve links and modes; all pass.
- Early lane cold measurements were 13.98/34.85/27.35 s. Committing the supplied
  install pin removed an unnecessary consumer dependency on an external install;
  the next measurements were 11.10/33.96/24.29 s. A's fixture correction changed
  the broad source identity, so all three were built cold again. The final
  delivery values are the 10.93/31.43/22.81 s table, not those earlier runs.
- /usr/bin/time is absent (127); reuse was remeasured with Python's clock.
- The first committed reader rerun exposed an empty helper-seeding commit in
  test_apply.py (exit 1 before apply). Pre-commit tests had new helpers to seed.
  A now uses --allow-empty, and committed/current-main tests and the real
  write-table mutant pass their intended outcomes. The premature first A push
  required one corrective fast-forward push, 7c4c2110 to ab06c988. No force push.
- Existing Python reader fixtures produced oracle __pycache__ files in child
  processes. They were removed; the Go shard explicitly suppresses bytecode,
  and final validation propagates that setting except where apply's own
  protection is deliberately tested.

The 150 remaining Go shard declarations and all-unit union/full unsharded oracle
were not run or generated. Shard 002 ran only as the isolation control. There is
no claim of remote publication, traced settlement, remote download performance,
uncached whole-repository gate or non-Linux-x64 support. Local reuse is labeled
as a hit, not a network fetch. The cold measurements include copying, builds,
verification and archive creation, but retain warm installed tools and OS caches.
No new .a fixtures or counts rows were added. No HOME cache was used by a product.

### Delivery addendum

Current main: `87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4`.
Final B code merge candidate:
`525e59440b287d9bd80052e71f61272ce757b19a`.
`git merge --no-commit --no-ff typescript/stage3-lane-products` exited 0 in the
separate main reference worktree. The merged code passed all three Go packages
(exit 0; lane 17.578 s), all 71 lane fixtures (0, 11.239 s), all 25 apply fixtures
(0, 15.183 s), and vet (0). Go product/shard consumers passed with
ADAMIC_TYPESCRIPT_SOURCE unset. Apply's own bytecode protection was tested with
PYTHONDONTWRITEBYTECODE unset. Every test command used timeout 90 and Go's
-timeout 90s. The temporary merge was aborted, exit 0; no main commit or push.

The report/evidence commit changes no recipe or artifact identity. Exact final
branch tip SHAs and the final push result accompany delivery. A's corrected tip
is `ab06c9886ccba3822d9e95effde1a7323e6a13b0` and is an ancestor of B.
Products used the /tmp tmpfs; the original producer comparison used the workspace
volume to avoid filling /tmp. These are observed unit-budget and reuse numbers,
not an isolated filesystem-throughput comparison.

Compressed logs, raw size/bind metadata, canonical ready markers, baseline phase
logs and owned-source mutant provenance are in `evidence/products-wave1/`.
