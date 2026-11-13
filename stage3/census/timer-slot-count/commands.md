# Commands and reproducibility

Executed from `/workspace/adamic` on Node v24.19.0. Only this census directory is committed. Scratch trees, dependencies, logs and runtime instrumentation are outside the repository. Initial adamic main was fetched as e3f59bf5509b1feef85b405807a732e11dc383b6.

The original preparation used `timeout` for finite setup/census commands: 600 seconds for source clone, 300 for API npm ci, 600 for upstream npm ci, 1200 for apply, and 300 for each scanner invocation. Each completed with its child's exit 0; no timeout occurred. `bounded.py` below is the repeatable wrapper that explicitly reports the child's actual return code, including termination signals. It exits with that code (128 + signal for negative Python return codes).

```sh
git fetch origin main
git switch -c typescript/timer-slot-count-995spev origin/main
D="$PWD/stage3/census/timer-slot-count"
python3 "$D/bounded.py" 600 git clone --depth 1 --branch v6.0.3 \
  https://github.com/microsoft/TypeScript.git /tmp/timer-slot-upstream
# Check HEAD is 050880ce59e30b356b686bd3144efe24f875ebc8.
python3 "$D/bounded.py" 300 npm ci --prefix "$PWD/stage3/api" --ignore-scripts --no-audit --no-fund
python3 "$D/bounded.py" 600 npm ci --prefix /tmp/timer-slot-upstream --ignore-scripts --no-audit --no-fund
ADAMIC_TYPESCRIPT_SOURCE=/tmp/timer-slot-upstream \
STAGE3_PRODUCT_STORE=/tmp/timer-slot-products \
  python3 "$D/bounded.py" 1200 stage3/apply.sh --build-product /tmp/timer-slot-adapted \
  > /tmp/timer-slot-apply.log 2>&1
# Setup generates diagnostics only, no adaptation, in the upstream scratch tree.
node /tmp/timer-slot-upstream/scripts/processDiagnosticMessages.mjs \
  /tmp/timer-slot-upstream/src/compiler/diagnosticMessages.json
# Provide all upstream type dependencies to the adapted census.
ln -s /tmp/timer-slot-upstream/node_modules /tmp/timer-slot-adapted/node_modules
NODE_PATH="$PWD/stage3/api/node_modules" \
  python3 "$D/bounded.py" 300 node "$D/count.cjs" /tmp/timer-slot-upstream \
  > /tmp/timer-slot-upstream.json 2> /tmp/timer-slot-upstream.status.log
NODE_PATH="$PWD/stage3/api/node_modules" \
  python3 "$D/bounded.py" 300 node "$D/count.cjs" /tmp/timer-slot-adapted \
  > /tmp/timer-slot-adapted.json 2> /tmp/timer-slot-adapted.status.log
```

The bounded wrapper prints its status to stderr, leaving child stdout parseable as JSON. The original oracle run used the earlier wrapper with the same status on stdout; its combined log retains that status. The measurement's final direct scans were:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" timeout 300 \
  node "$D/count.cjs" /tmp/timer-slot-upstream > /tmp/timer-slot-upstream.json
NODE_PATH="$PWD/stage3/api/node_modules" timeout 300 \
  node "$D/count.cjs" /tmp/timer-slot-adapted > /tmp/timer-slot-adapted.json
```

Both child exit codes were 0. Raw JSON from these scans is embedded in count.json under apiScans. Upstream source generation adds diagnosticInformationMap.generated.ts; adapted generation/adaptations add the additional source roots listed in the raw scans. The actual run linked the adapted type dependencies during apply, before its final packaging; no source was changed by that link. The source census ran after apply exited 0.

The optional cross-check uses the lane's oracle runner directly, rather than the lane's API guard, on a disposable instrumented adapted tree:

```sh
node "$D/instrument.cjs" /tmp/timer-slot-adapted /tmp/timer-slot-adapted.json
# Expected: {"files":8,"sites":19}
mkdir -p /tmp/timer-slot-runtime
TIMER_SLOT_OUTPUT=/tmp/timer-slot-runtime \
NODE_OPTIONS="--require=$D/runtime-preload.cjs" \
  python3 "$D/bounded.py" 1200 stage3/oracle/run.sh \
  /tmp/timer-slot-adapted /tmp/timer-slot-oracle --limit-seconds 1100 \
  > /tmp/timer-slot-oracle.log 2>&1
python3 "$D/summarize.py" /tmp/timer-slot-upstream.json \
  /tmp/timer-slot-adapted.json /tmp/timer-slot-runtime
```

The oracle installs/builds its own locked dependencies, so the static census dependency link is replaced before runtime. npm ci also emptied the linked upstream dependency directory; it was reinstalled before the final validation scans. `runtime-preload.cjs` writes per-process aggregates on normal process exit. Runtime captures count invocations of census call sites, including clearTimeout, not invocations of global timers or callbacks. Tail identity distinguishes object/function references within each process; primitive-only tails deduplicate across processes. Every distinct tail and count is in count.json. Zero-tail calls retain only their count, not callback or timer handle values. Instrumentation does not stringify/traverse supplied objects, execute callbacks early or alter tails. It does allocate wrappers and log counters, so timings are measurement timings.

The reviewed classifications live in classifications.json; summarize.py asserts every tail expression and pairs independently scanned source locations. count.cjs retains all timer-name candidates so exclusions can be audited. New candidates or changed expressions must be reviewed; the assertions deliberately reject stale classifications.

Measured oracle result: install child 0 (3.496 s), build child 0 (23.739 s), tests child 1 (296.127 s), oracle child 1, timeout false, total 323.443 s. Exactly 106366 passing, one failing, zero pending; the failure title and sole api/typescript.d.ts diff match stage3/lane/expected.json. Syntax checks passed for all three .cjs tools and AST parsing of both .py tools. The bounded runner was also checked with a child exiting 7: it reported childReturncode 7 and returned 7.

Final validation expanded the scanner to generic bulk copies. Its saved adapted raw scan uses `/tmp/timer-slot-adapted-static`, a source-only extraction of src/, scripts/ and package.json from the verified, original apply product (before runtime instrumentation). A whole-product restore was refused by Python data_filter because our temporary node_modules link was absolute; the source-only extraction excludes that link. The product manifest key was `31551ed146d278985daa0bb66085a803baba07e0ee5c922b496618203e68daec`. No adapted source was regenerated or edited for this validation. Type dependencies were reinstalled in upstream and linked into the restored source tree. The final scan commands were:

```sh
timeout 300 npm ci --prefix /tmp/timer-slot-upstream --ignore-scripts --no-audit --no-fund
NODE_PATH="$PWD/stage3/api/node_modules" timeout 300 node "$D/count.cjs" \
  /tmp/timer-slot-upstream > /tmp/timer-slot-upstream.json
NODE_PATH="$PWD/stage3/api/node_modules" timeout 300 node "$D/count.cjs" \
  /tmp/timer-slot-adapted-static > /tmp/timer-slot-adapted.json
python3 "$D/summarize.py" /tmp/timer-slot-upstream.json \
  /tmp/timer-slot-adapted.json /tmp/timer-slot-runtime
```

For a clean reproduction, use the primary commands above with the final count.cjs before instrumenting the disposable adapted tree; the extra restore/rescan is unnecessary. The summarizer requires missing-module/type-definition diagnostics to be absent and requires matching slot declarations and receiver types in both trees.
