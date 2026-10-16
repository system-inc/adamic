# Main scanner measurement, October 14

Observed status: **scanner-blocked**, exit 1. Base and integrated compiler SHA:
`87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4`. Requested refs: `[]`; merges: `[]`; nproc: 5.
Total wall time: 128.141s. Summary validation: valid, no errors.

Node oracle: 509,014 skipped-trivia tokens, 860,418 retained-trivia tokens,
1,369,432 total, 466 errors, 81 files. No corpus count change.
The full coverage dump contains 108,018,963 bytes. Its SHA-256 is
`10e253011b72b3689350cc72123758e335e1c40d8256b02493016cb8c0d24581`.
Pass headers contain absolute input paths, so whole-dump byte counts and hashes
also reflect the scratch directory name.

| Split | Runner exit | Build exit | Node exit | Full-tree Node diff | Copied control diff | Token-end mutant diff |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | 1 | 1 | 0 | 0 | 0 | 1 |
| 1 | 1 | 1 | 0 | 0 | 0 | 1 |

Both builds refuse the same unchanged MapLike index signature. Native diff exits,
planted native one-byte mutant exits and measure.py timings are unavailable:
neither build produced a native scanner. The token-end mutant increments the
first token end from 76 to 77 in scratch Node output and is caught by diff.
The reference run also records comparison control exit 0 and end-mutant diff exit 1.

## Ordered stops

1. `corePublic.ts:9:5`; October 8 #1 (MapLike index signature).
   Diagnostic: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added
   Scope: first unchanged stop.
   Fresh witness: Node exit 0, build exit 1, diagnostic matches.
   Continuation placeholder: index-signature removal with throwing marker; later record observations are dependent.
2. `debug.ts:14:14`; October 8 #3 (Debug unchecked Error cast).
   Diagnostic: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
   Scope: behind throwing discovery placeholders.
   Fresh witness: Node exit 0, build exit 1, diagnostic matches.
   Continuation placeholder: smallest enclosing function body throws; signature retained/inferred.
3. `diagnosticInformationMap.generated.ts:13:34`; October 8 #4 (diagnostic factory unchecked cast).
   Diagnostic: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
   Scope: behind throwing discovery placeholders.
   Fresh witness: Node exit 0, build exit 1, diagnostic matches.
   Continuation placeholder: typed throwing diagnostic factory; later casts are placeholder-dependent.
4. `utilities.ts:65:22`; October 8 #5 (Uint16Array element type).
   Diagnostic: stage 0 can't lower typed array element type Uint16Array yet
   Scope: behind throwing discovery placeholders.
   Fresh witness: Node exit 0, build exit 1, diagnostic matches.
   Continuation placeholder: smallest enclosing function body throws; signature retained/inferred.
5. `scanner.ts:3510:67`; October 8 #7 (UTF16 worker unchecked String cast).
   Diagnostic: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
   Scope: behind throwing discovery placeholders.
   Fresh witness: Node exit 0, build exit 1, diagnostic matches.
   Continuation placeholder: typed throwing UTF16 worker.
6. `core.ts:107:20`; October 8 #8 (new Array(length) with holes).
   Diagnostic: stage 0 can't lower new an Identifier yet
   Scope: behind throwing discovery placeholders.
   Fresh witness: Node exit 0, build exit 1, diagnostic matches.
   Continuation placeholder: smallest enclosing function body throws; signature retained/inferred.
7. `scanner.ts:778:15`; new.
   Diagnostic: stage 0 can't lower destructuring a string yet
   Scope: behind throwing discovery placeholders.
   stopped: no freshly verified Node witness for this diagnostic.

Stop 5 matches the UTF16 worker construct despite the larger line shift from
3497 to 3510. Stop 7 is string length destructuring inside
checkForIdentifierStartAfterNumericLiteral. The unchanged runner stops there;
no additional witness or source bypass was added. Later historical stops remain
unmeasured, not retired.

## Commands and prerequisite failures

All commands that could run long had hard limits, wrote logs, and were monitored.
The worker first fetched origin, created the requested branch from d072c133,
exported `GOPROXY=https://proxy.golang.org|direct`, and ran
`timeout 900 bash cloud/setup.sh > /tmp/scanner-setup.log 2>&1` (exit 0).
The printed `/workspace/adamic-tools/env.sh` was sourced for subsequent runs.

```sh
# Initial literal command, exit 1; summary retained as initial-summary.json.
timeout 5400 bash stage3/drivers/scanner/scratch-run.sh /tmp/scanner-main > /tmp/scanner-main.log 2>&1; echo exit=$?
# Helper import retry, exit 1; prerequisite-retry-summary.json.
export PYTHONPATH=/workspace/adamic/stage3
timeout 5400 bash stage3/drivers/scanner/scratch-run.sh /tmp/scanner-main-retry > /tmp/scanner-main-retry.log 2>&1; echo exit=$?
# Prepare the pinned gate inputs, all exit 0.
timeout 600 git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/scanner-typescript-source > /tmp/scanner-source-setup.log 2>&1
timeout 180 npm ci --prefix /tmp/scanner-main-retry/pipeline/api --ignore-scripts --no-audit --no-fund > /tmp/scanner-api-setup.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/scanner-typescript-source
timeout 900 bash /tmp/scanner-main-retry/pipeline/apply.sh /tmp/scanner-prepared-product --build-product > /tmp/scanner-product-setup.log 2>&1
# Copied API lock/pin files to the standard scratch cache, then installed.
timeout 180 npm ci --prefix /home/agent/.cache/adamic-stage3/api --ignore-scripts --no-audit --no-fund > /tmp/scanner-cache-api-setup.log 2>&1
# Measured prepared run, exit 1 (scanner-blocked).
timeout 5400 bash stage3/drivers/scanner/scratch-run.sh /tmp/scanner-main-prepared > /tmp/scanner-main-prepared.log 2>&1; echo exit=$?
```

The first apply failed with `ModuleNotFoundError: No module named caches` because
scratch-run.py copies apply.py without its caches.py helper. PYTHONPATH resolves
only that import. The next apply failed because a cold worker lacked
ADAMIC_TYPESCRIPT_SOURCE; the pipeline also needed its pinned API install.
The prepared product uses the runner-selected 00/10/42/50/51 profile unchanged.
All source and adaptation work stayed in scratch. Both failed summaries and
apply stderr files are retained.

Main advanced to 87ad5a4f while prerequisites were prepared. The completed
measurement uses that fetched tip; the delivery branch was fast-forwarded to it
before adding this evidence. The commit message retains the requested d072c133
wording; this report and BLOCKERS identify the SHA actually measured.

No phase exceeded 30 minutes. The longest phase was the first fetch, 1174.462s.
Log tails, output-directory listings and process checks were made throughout;
nothing ran 15 minutes unwatched.

## Toolchain timings

Worker setup exit 0, total 834.004s. Timing lines are in worker-setup.log.
Node ready 0.064s, Go ready 0.065s, clang ready 0.466s, Markdown installed
3.531s (ready 3.713s), submodules ready 525.156s, shared cache ready 530.736s,
Go build ready 833.873s, cache warm 833.968s. nproc=5; CPU quota 4.
Go 1.27.1, clang 20.1.8, Node 24.19.0.

Measured runner setup timing lines:

```text
setup: node ready (0.017s)
setup: go ready (0.016s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.006s
setup: submodules ready (0.061s)
setup: markdown dependencies ready (0.062s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.155s)
setup: shared cache ready (0.211s)
setup: go build ready (59.354s)
setup: test binaries deferred (use --warm-tests) (59.455s)
setup: build cache warm (59.456s)
setup: build-flags commit=87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=1.14 1.29 0.94 1/177 9117 load-after=1.52 1.29 0.96 1/183 9655
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (59.484s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.ajcsFz
```

## Every measured phase

| Phase | Exit | Wall seconds |
| --- | --- | --- |
| fetch | 0 | 3.273 |
| main-ref | 0 | 0.004 |
| worktree | 0 | 1.969 |
| integrated-ref | 0 | 0.004 |
| cohere-pin | 0 | 0.003 |
| cohere-init-local | 0 | 0.003 |
| cohere-shared-ref | 0 | 0.001 |
| cohere-shared-git | 0 | 0.001 |
| cohere-checkout | 0 | 0.315 |
| cohere-typescript-pin | 0 | 0.004 |
| cohere-typescript-init-local | 0 | 0.003 |
| cohere-typescript-shared-ref | 0 | 0.001 |
| cohere-typescript-shared-git | 0 | 0.001 |
| cohere-typescript-checkout | 0 | 1.07 |
| setup | 0 | 59.532 |
| nproc | 0 | 0.004 |
| compiler-node-types | 0 | 0.466 |
| compiler-build | 0 | 4.431 |
| apply | 0 | 8.892 |
| corpus-scanner | 0 | 0.016 |
| corpus-type-imports | 0 | 1.834 |
| reference | 0 | 10.383 |
| slice | 0 | 3.937 |
| scanner-0 | 1 | 6.694 |
| full-tree-check-0 | 0 | 0.164 |
| scanner-1 | 1 | 6.702 |
| full-tree-check-1 | 0 | 0.165 |
| discovery-01 | 1 | 0.416 |
| witness-01-00-node | 0 | 0.214 |
| witness-01-00-build | 1 | 0.064 |
| discovery-02 | 1 | 0.416 |
| witness-02-00-node | 0 | 0.265 |
| witness-02-00-build | 1 | 0.064 |
| discovery-02-placeholder | 0 | 0.215 |
| discovery-03 | 1 | 0.466 |
| witness-03-00-node | 0 | 0.265 |
| witness-03-00-build | 1 | 0.068 |
| discovery-03-placeholder | 0 | 0.72 |
| discovery-04 | 1 | 0.518 |
| witness-04-00-node | 0 | 0.215 |
| witness-04-00-build | 1 | 0.064 |
| discovery-04-placeholder | 0 | 0.215 |
| discovery-05 | 1 | 4.539 |
| witness-05-00-node | 0 | 0.265 |
| witness-05-00-build | 1 | 0.064 |
| discovery-06 | 1 | 4.332 |
| witness-06-00-node | 0 | 0.265 |
| witness-06-00-build | 1 | 0.064 |
| discovery-06-placeholder | 0 | 0.215 |
| discovery-07 | 1 | 4.23 |

Initial and first-retry phase exits and wall seconds are preserved in their
complete summaries. No phases are inferred past a failed prerequisite.

## Validation and limits

`timeout 90 python3 -B stage3/drivers/scanner/test_scratch_runner.py` exited 0:
five tests passed. Existing mutants reject a conflict falsely labeled pass and
a pass whose native-byte mutant diff exits 0. No new mutant was added.
Both measured scanner token-end mutants were caught by diff exit 1.

Raw failing split builds and all failing discovery/witness-build stderr are
retained here. No retained file exceeds 200 KB; no compression was needed.
No binaries or generated source trees are included.

Not covered: native scanner execution or timings, native/Node equality, native
one-byte mutants, parser-driven rescans, stops after the seventh diagnostic, or
a full compiler/repository gate. This delivery changes documentation/evidence only.
