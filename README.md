# #et209ny record-stop proof

Change commit: c73b0e271290f6506efa889538e7afede0d44a5b (parent cf468df0747075e50d326a97ff03e8530d7a1a88).
Memory revision: a3f1d6b4558b88f8c2640a8d120be8e43acd76a6.
TypeScript v6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.

The fixture keeps @ts-check and writes both a node-array slice and a Record<string, number> into a union field of a cyclic Node. The scalar record type is not selected for graph membership. Both production versions stop with `a record written into a graph slot`; both scan probes record `first declaration/body/tail stop` with Rule `lowering.findCycles`. The stop is conservative over the union slot, not a claim that the scalar record itself closes a runtime cycle.

Only record-stop.ts and one README line changed in the source commit. No internal source, control targets or other fixture changed. The unchanged record-clean.ts completes in production and scan on both revisions.

## Four control outputs

### control-main-record-stop.log

```text
plain production record-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmpai0c7as4/control.ts:5:1: stage 0 can't lower a record written into a graph slot, where the write at /workspace/et209ny/tmp/tmpai0c7as4/control.ts:5:1 may close a cycle through it, in a graph program: a record can't join a graph region yet"}
record-stop production and scan-continuation control passed
```

### control-main-record-clean.log

```text
plain production record-clean: {"Completed": true, "Error": ""}
record-clean production and scan-continuation control passed
```

### control-memory-record-stop-final.log

```text
plain production record-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmpl_dzbhr5/control.ts:5:1: stage 0 can't lower a record written into a graph slot, where the write at /workspace/et209ny/tmp/tmpl_dzbhr5/control.ts:5:1 may close a cycle through it, in a graph program: a record of a type the program's graph doesn't select can't join a graph region yet"}
record-stop production and scan-continuation control passed
```

### control-memory-record-clean.log

```text
plain production record-clean: {"Completed": true, "Error": ""}
record-clean production and scan-continuation control passed
```

## Package and mutant

`go test -v -count=1 -timeout=10m ./stage3/census/scan-all` on cf468df0 plus the change:
`ok github.com/system-inc/adamic/stage3/census/scan-all 22.207s` (all tests, including eight controls and anchors, PASS). Full output: package-main.log.

The exact named-test mutant command from verify.sh was run with `SCAN_ALL_MUTANT_FIRST_ERROR=1`, `-v -count=1 -timeout=60s -run '^TestScanAllRecordStopAfterEarlierScanStop$'`. It failed with `continuation-target`, after observing only the checking pragma. The verified line is:

```text
first-error mutant red: TestScanAllRecordStopAfterEarlierScanStop
```

See mutant-TestScanAllRecordStopAfterEarlierScanStop.log and mutant-result.txt. This task ran that requested named mutant, not all of verify.sh's mutants or isolated anchor mutations.

## Build and environment

Fresh `git worktree add --detach` worktrees under /workspace/et209ny were made for each exact SHA. Each ran `git submodule update --init --recursive --depth 1`; logs retained. TypeScript was cloned with `--depth 1 --branch v6.0.3`.

The environment was:

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/et209ny/tmp GOTMPDIR=/workspace/et209ny/tmp
export GOCACHE=/workspace/et209ny/cache
export ADAMIC_BUILD_CACHE_DIR=/workspace/et209ny/products
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/et209ny/TypeScript
```

Main's two products were built by `go test -v -count=1 -timeout=30m -run '^TestProduct_ScanAll' ./stage3/census/scan-all`, using products_test.go. The initial cold run's ten-minute timeout is retained in products-main.log; the successful retry is products-main-retry.log. All four eventual binaries were built successfully; binaries.sha256 records their hashes.

Memory's revision predates products_test.go, so its binaries were built directly with the same recipe and flags, using its own unchanged make_overlay.py and templates. No test declarations or harness sources were copied into it:

```sh
python3 stage3/census/scan-all/make_overlay.py /workspace/et209ny/memory-overlay
go build -trimpath -ldflags=-buildid= -buildvcs=false \
  -overlay /workspace/et209ny/memory-overlay/overlay.json \
  -o /workspace/et209ny/memory-scan ./stage3/census/scan-all/tool/main.go
```

For Memory's production probe, memory-production-overlay.json maps only the absolute `stage3/census/scan-all/production/main.go` path to its production.go.txt. Build command:

```sh
go build -trimpath -ldflags=-buildid= -buildvcs=false \
  -overlay /workspace/et209ny/memory-production-overlay.json \
  -o /workspace/et209ny/memory-production ./stage3/census/scan-all/production/main.go
```

Each worktree's unchanged controls.py ran twice with SCAN_ALL_BINARY and SCAN_ALL_PRODUCTION pointing to its own two binaries. The four final logs are above; scan-*.json additionally records the complete scan streams directly against the final fixtures, including exact rule and words. Main's product paths appear in products-main-retry.log. Build-memory*.log capture the direct builds' output (successful Go builds are silent).

Exploratory candidates are retained as candidate1 through candidate6 logs. Candidates1-4 did not meet the target; candidate5 passed, then candidate6 reduced it to the final direct writes. Only the final eight-line fixture was committed and used for the final four control logs, package test, and mutant.
