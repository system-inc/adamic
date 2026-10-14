# Literal one-command rerun, October 14

The fixed runner reaches the same seven stops without the earlier manual
prerequisites or PYTHONPATH workaround. The exact command was:

```sh
timeout 5400 bash stage3/drivers/scanner/scratch-run.sh /tmp/scanner-main-2 > /tmp/scanner-main-2.log 2>&1
```

Exit 1, status `scanner-blocked`, total 166.323s.
Compiler base and integrated SHA: `87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4`.
No requested refs or merges. nproc=5. No phase exceeded 30 minutes;
log tails and output-directory listings were checked throughout.

## Prerequisite preparation

Earlier scratch trees and the manually populated stage3 cache were archived
under /workspace/scanner-first-run-retained before this run. No source checkout,
API install or adaptation product was supplied by hand. apply.stderr records
a product cache miss and a fresh local build. Toolchain setup is the runner's
normal cloud/setup.sh phase.

The runner copied caches.py beside apply.py, cloned TypeScript v6.0.3 into
its output directory, verified commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, and installed the pinned API
packages in the copied pipeline and the cache selected by stage3/caches.py.
Both installs passed apply.py's independent package-byte verification.
The summary records these gate-input paths and all preparation phases.
The source, compiler, API and adapted trees remain scratch artifacts only.

## Regression proof

The pipeline-copy statements were extracted unchanged to copy_pipeline, and
the new test ran that actual copy's apply.py --key from a different working
directory, with PYTHONPATH removed. Before adding caches.py, the test command
exited 1 with the original ModuleNotFoundError: No module named caches.
The other five tests passed. before-fix-tests.log preserves the failure.
After the fix, all nine tests pass, including declared build-cache placement
and explicit errors for a missing source path or incorrect source commit.

```sh
timeout 90 python3 -B stage3/drivers/scanner/test_scratch_runner.py > /tmp/scanner-followup-tests.log 2>&1
```

The old missing-copy behavior is caught specifically by importing caches;
no network, compiler build, or unrelated prerequisite kills that regression.
Existing false-pass summary mutants remain caught by the summary tests.

## Scanner observations

Both split modes build-exit 1 at unchanged corePublic.ts:9:5. Both Node
comparisons against the full-tree reference exit 0, copied controls exit 0,
and token-end mutants are caught with diff exit 1. Node still emits 509,014
skipped-trivia and 860,418 retained-trivia tokens, 466 errors across 81 files.
No native scanner executable, native-byte mutant or measure.py timing exists.

All seven locations and first-line diagnostics match the prior summary exactly.
The first stop is unchanged source; later stops are behind discovery placeholders.
Stops 1 through 6 retain freshly verified Node-success/build-failure witnesses.
The walk stops at string destructuring because no fresh catalogue witness is
available. It is not an exhaustive stop census.

1. `corePublic.ts:9:5`: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added
2. `debug.ts:14:14`: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
3. `diagnosticInformationMap.generated.ts:13:34`: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
4. `utilities.ts:65:22`: stage 0 can't lower typed array element type Uint16Array yet
5. `scanner.ts:3510:67`: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
6. `core.ts:107:20`: stage 0 can't lower new an Identifier yet
7. `scanner.ts:778:15`: stage 0 can't lower destructuring a string yet

## Setup timing lines

```text
setup: node ready (0.020s)
setup: go ready (0.023s)
setup: submodules ready (0.061s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.008s
setup: markdown dependencies ready (0.072s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.153s)
setup: shared cache ready (2.213s)
setup: go build ready (65.345s)
setup: test binaries deferred (use --warm-tests) (65.442s)
setup: build cache warm (65.444s)
setup: build-flags commit=87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.14 0.10 0.27 1/185 11177 load-after=1.84 0.50 0.39 1/190 11841
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (65.471s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.C94Ekd
```

## Phase exits and wall seconds

| Phase | Exit | Seconds |
| --- | --- | --- |
| fetch | 0 | 1.82 |
| main-ref | 0 | 0.004 |
| worktree | 0 | 2.019 |
| integrated-ref | 0 | 0.004 |
| cohere-pin | 0 | 0.004 |
| cohere-init-local | 0 | 0.004 |
| cohere-shared-ref | 0 | 0.004 |
| cohere-shared-git | 0 | 0.001 |
| cohere-checkout | 0 | 0.314 |
| cohere-typescript-pin | 0 | 0.004 |
| cohere-typescript-init-local | 0 | 0.004 |
| cohere-typescript-shared-ref | 0 | 0.004 |
| cohere-typescript-shared-git | 0 | 0.001 |
| cohere-typescript-checkout | 0 | 1.318 |
| setup | 0 | 65.493 |
| nproc | 0 | 0.004 |
| compiler-node-types | 0 | 0.465 |
| compiler-build | 0 | 4.987 |
| typescript-source | 0 | 16.168 |
| typescript-source-ref | 0 | 0.008 |
| pipeline-api | 0 | 0.415 |
| cache-api | 0 | 0.416 |
| pipeline-api-digest | 0 | 0.114 |
| apply | 0 | 25.577 |
| corpus-scanner | 0 | 0.008 |
| corpus-type-imports | 0 | 1.783 |
| reference | 0 | 10.365 |
| slice | 0 | 3.779 |
| scanner-0 | 1 | 6.538 |
| full-tree-check-0 | 0 | 0.164 |
| scanner-1 | 1 | 6.442 |
| full-tree-check-1 | 0 | 0.165 |
| discovery-01 | 1 | 0.466 |
| witness-01-00-node | 0 | 0.265 |
| witness-01-00-build | 1 | 0.066 |
| discovery-02 | 1 | 0.418 |
| witness-02-00-node | 0 | 0.215 |
| witness-02-00-build | 1 | 0.065 |
| discovery-02-placeholder | 0 | 0.215 |
| discovery-03 | 1 | 0.469 |
| witness-03-00-node | 0 | 0.265 |
| witness-03-00-build | 1 | 0.065 |
| discovery-03-placeholder | 0 | 0.718 |
| discovery-04 | 1 | 0.569 |
| witness-04-00-node | 0 | 0.215 |
| witness-04-00-build | 1 | 0.064 |
| discovery-04-placeholder | 0 | 0.214 |
| discovery-05 | 1 | 4.237 |
| witness-05-00-node | 0 | 0.265 |
| witness-05-00-build | 1 | 0.067 |
| discovery-06 | 1 | 4.128 |
| witness-06-00-node | 0 | 0.215 |
| witness-06-00-build | 1 | 0.064 |
| discovery-06-placeholder | 0 | 0.214 |
| discovery-07 | 1 | 4.229 |

Saved summary validation and the exact-stop, cold-cache and token-count audits
pass. Full compiler and repository gates were not run; only scanner driver code
and documentation changed. Raw failing build stderr and regression logs are
retained here, without binaries or generated trees.
