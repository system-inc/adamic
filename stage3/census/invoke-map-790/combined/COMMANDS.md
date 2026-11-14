# Combined-head commands and provenance

Implementation e53da7b28b7191e0dee15d87b67f80d859b83eef has the sole parent
8e3a784aa5a601f57342c07b68f145fe56252c21. The same scratch worktree, initialized
submodules, TypeScript source 050880ce59e30b356b686bd3144efe24f875ebc8,
stock API install, Go binary, measurement overlay and scratch caches from the
first amendment are reused. Their complete setup and hashes remain in ../.
No production compiler or census implementation changes were made.

Exactly one form was tried for #frhzgrc:

```ts
(newChildWatches || (newChildWatches = [] as ChildDirectoryWatcher[])).push(childWatcher);
```

The existing map annotation remains `const invokeMap: InvokeMap = new Map();`.
Stock TypeScript accepts the proposed assertion with zero diagnostics; the full
census accepts it with no new or changed finding. No typed-local or declaration
fallback was tried. Its entire census is retained as fresh-literal-cast-census
and identically copied to combined-census, since this trial source is the exact
source adopted by the final combined owner pin. No refused cast is included.

All four complete census invocations use:

```sh
source /tmp/invoke-map-4pc8wd8/env.sh
CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE=/tmp/invoke-map-4pc8wd8/main-tree/src/compiler/sys.ts \
/tmp/invoke-map-4pc8wd8/latent-census /tmp/invoke-map-4pc8wd8/main-tree/src/compiler \
/tmp/invoke-map-4pc8wd8/evidence-2/NAME-census.jsonl
```

NAME is fresh-literal-cast (combined), main, restore-map, or restore-child-array.
Each exits 0. The last three run sequentially on one source path, restoring
combined source in a finally block. Every actual sys.ts input is retained.
Main restores both original declarations. restore-map undoes only the map
annotation; restore-child-array undoes only the fresh-literal assertion.

`compare.py` takes BEFORE_JSONL AFTER_JSONL BEFORE_SYS AFTER_SYS OUTPUT_JSON.
It retains every finding's kind, phase, attempting unit, where, reason, text,
and boundary span. Exact byte-offset maps for changed type syntax are built
from unchanged lines plus the two small modified-line diffs; offsets in other
files are never changed. All complete deltas are retained as JSON and text.
Main/combined has 0 new, 0 changed, 2 gone. Each combined/mutant comparison has
one new finding at its own site, no gone or changed finding. Fresh main equals
the previous main raw stream byte-for-byte; restore-child-array equals the
previous map-only head raw stream byte-for-byte. No finding is filtered out.

Combined checks (each exit 0):

```sh
node /tmp/invoke-map-4pc8wd8/stock.cjs /tmp/invoke-map-4pc8wd8/main-tree /tmp/invoke-map-4pc8wd8/evidence-2/fresh-literal-cast-stock
/tmp/invoke-map-4pc8wd8/adamic types $(find /tmp/invoke-map-4pc8wd8/main-tree/src/compiler -name '*.ts' | sort)
node stage3/adapt/99-correctness-timer-closures/witness.cjs /tmp/invoke-map-4pc8wd8/main-tree
node stage3/adapt/99-correctness-timer-closures/contract.cjs /tmp/invoke-map-4pc8wd8/contract-input
(cd /tmp/invoke-map-4pc8wd8/main-tree && ./node_modules/.bin/hereby dts)
```

Stock uses the established strict compiler-project helper and all 79 recursive
compiler roots. The stock script is retained in ../stock.cjs. The 79 actual
JavaScript files are retained in javascript.tar.gz and each is byte-compared
with ../main-stock-javascript.tar.gz. Both declarations are outputs of upstream's
own hereby dts and byte-compared with ../main-dts. The baseline's stock,
Adamic, contract, witness and dts results are from exactly the same main/source
and toolchain as before, and are reused; all baseline artifacts remain in ../.
The combined contract uses the same exact before-owner fixture as the baseline.
All existing witnesses and contract code remain unchanged.

Fresh full lane:

```sh
stage3/lane/run.sh /tmp/invoke-map-4pc8wd8/lane-2 > /tmp/invoke-map-4pc8wd8/evidence-2/lane.log 2>&1
```

It runs from the actual combined implementation HEAD, with no artifact shortcut,
filters, shard substitution or edits to lane/oracle. Its apply tree is untouched
until apply exits. The previous lane's source tree was removed only after all
its receipts had been preserved and pushed; no shared checkout was modified.

The combined evidence is appended using a separate temporary Git index based on
the current evidence tip and commit-tree with that tip as parent. The running
implementation checkout stays at e53da7b28. All pushes are ordinary fast-forward
pushes; no existing ref is forced. No Loom command or submission is invoked.

## Resource retry

The first lane stops after a worker exits without output or a test failure.
Cgroup memory.max is 17179869184; memory.events records oom_kill=2. This
suggests memory pressure, but no per-process kernel diagnosis is claimed.
The failed run has no complete counts and is retained as a failed attempt.
All sys.ts inputs, output artifacts and full source hashes were already saved
before removing the task's unused Go cache, completed main-tree and failed
lane adapted-tree. No proof output or shared checkout file was removed.

```sh
stage3/lane/run.sh /tmp/invoke-map-4pc8wd8/lane-2-retry > /tmp/invoke-map-4pc8wd8/evidence-2/lane-retry.log 2>&1
```

The retry uses the unchanged lane, the same implementation HEAD, eight workers,
all runners and no test filter. Apply restores its correctly keyed product into
a new tree; the initial lane already proved a fresh local adaptation pass.
All 79 retry roots are independently checked against the original proof input
hashes, and sys.ts is compared directly with the frozen combined bytes.
