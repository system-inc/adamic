# #et209ny: combined record-stop and cycle-stop proof

Change commit: 0fac40b6923b1cca00849cf2319f8681a799db0b on branch typescript/scan-all-record-stop-2.
Parent: cf468df0747075e50d326a97ff03e8530d7a1a88.
Second tested lowering revision: a3f1d6b4558b88f8c2640a8d120be8e43acd76a6.

The single change commit includes only record-stop.ts, cycle-stop.ts and two README line replacements. Both fixtures retain the leading `// @ts-check`. All other fixtures, controls.py and internal/ are unchanged.

record-stop uses a Node[] | Record<string, number> field written with a node-array slice and then a scalar record. cycle-stop is six lines: the field rs: Node[] | Record<string, number>[] is written with [a], making a self-cycle, then with [record], whose explicitly annotated record holds a scalar. This retains the record-inside-array path through holdsRecordSeen on main / heldRecordsSeen after Memory. The explicit record annotation avoids an earlier fixed-object/record storage conversion stop from inline [{ k: 1 }].

Both stops rest on judgeRecordSlots refusing a scalar record type, which is stricter than soundness needs (a record of scalars can't close a cycle); if that rule later narrows to records that can hold a graph value, both fixtures move again. This describes Memory's judgeRecordSlots boundary; main's older noteRecordSlot stop refuses all records in such slots. These controls pin the current conservative stop, rather than demonstrating a runtime cycle through scalar record values.

## Results

All eight controls passed on each revision (sixteen runs), using each side's own two binaries. Every output is reproduced below and retained in a separate control-*.log. All stop controls fail plain production lowering with their target words and retain the checking pragma scan stop followed by the required production boundary. Every clean control completes in production and scan. Complete scan streams are scan-*.json. Both record and cycle stops on both revisions have Kind `first declaration/body/tail stop`, Rule `lowering.findCycles`, and words containing `a record written into a graph slot`.

The entire scan-all package on cf468df0 plus the combined change:

```text
go test -v -count=1 -timeout=10m ./stage3/census/scan-all
PASS
ok github.com/system-inc/adamic/stage3/census/scan-all 2.437s
```

package-main.log includes all eight controls, the anchors and both product declarations.

Both requested named first-error mutants failed for continuation-target, as verify.sh requires:

```text
first-error mutant red: TestScanAllRecordStopAfterEarlierScanStop
first-error mutant red: TestScanAllCycleStopAfterEarlierScanStop
```

Each used `SCAN_ALL_MUTANT_FIRST_ERROR=1 go test -v -count=1 -timeout=60s -run '^TEST_NAME$' ./stage3/census/scan-all`. Complete failures are mutant-TestScanAll{Record,Cycle}StopAfterEarlierScanStop.log; mutant-result.txt records the verified lines. Only these two requested mutants were run for this follow-up.

## Setup and binary provenance

This follow-up reused the task's two fresh detached worktrees, recursive shallow submodules, TypeScript v6.0.3 clone and successfully built binaries from the first round. Main was returned to detached cf468df0 before applying both fixtures and both README lines, then committed once on the new branch. Memory remains at a3f1d6b4 with only the same two fixtures edited. Binary SHA-256 hashes are in binaries.sha256; they match the original round. The main package test fetched both unchanged products from cache, with hit logs in package-main.log. No lowering implementation was edited or overlaid beyond the instrumented scan product's standard recipe.

The environment was:

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/et209ny/tmp GOTMPDIR=/workspace/et209ny/tmp
export GOCACHE=/workspace/et209ny/cache
export ADAMIC_BUILD_CACHE_DIR=/workspace/et209ny/products
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/et209ny/TypeScript
```

TypeScript was cloned with `--depth 1 --branch v6.0.3` and is 050880ce59e30b356b686bd3144efe24f875ebc8. Each original worktree ran `git submodule update --init --recursive --depth 1`. Original setup and build logs are included here for standalone provenance.

Main's products were built using its products_test.go declarations:

```sh
go test -v -count=1 -timeout=30m -run '^TestProduct_ScanAll' ./stage3/census/scan-all
```

products-main-retry.log records the successful cold builds. products-main.log is the initial ten-minute cold-build timeout from the first round, retained for completeness.

Memory predates products_test.go, so its probes were built directly using the same recipes and flags, with its own unchanged overlay generator and templates:

```sh
python3 stage3/census/scan-all/make_overlay.py /workspace/et209ny/memory-overlay
go build -trimpath -ldflags=-buildid= -buildvcs=false \
  -overlay /workspace/et209ny/memory-overlay/overlay.json \
  -o /workspace/et209ny/memory-scan ./stage3/census/scan-all/tool/main.go
go build -trimpath -ldflags=-buildid= -buildvcs=false \
  -overlay /workspace/et209ny/memory-production-overlay.json \
  -o /workspace/et209ny/memory-production ./stage3/census/scan-all/production/main.go
```

The production overlay replaces only the absolute production/main.go path with production.go.txt. Successful Go builds are silent; build-memory*.log retain their output. No harness sources were copied into Memory.

Main's SCAN_ALL_BINARY is /workspace/et209ny/products/local/f41b91c97d420fc9eb22166e8bd071817a2199431b4480a919fd6657089e0f5e/probe; SCAN_ALL_PRODUCTION is /workspace/et209ny/products/local/fc4cd7210fc6457c5f06879df6180b575aff9160c0a68cac51b0924c8e12d68f/probe. Memory uses /workspace/et209ny/memory-scan and /workspace/et209ny/memory-production respectively.

Each side ran its unchanged `python3 stage3/census/scan-all/controls.py NAME` with those environment variables for count-stop, count-clean, rest-stop, rest-clean, record-stop, record-clean, cycle-stop and cycle-clean. Every process returned exit 0. Scan JSON was additionally captured by calling that side's scan binary directly on each pragma-bearing fixture. The final six-line cycle fixture was used in all logs here; earlier exploration outputs are excluded.

## Every control output

### control-main-count-stop.log

```text
plain production count-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmpl3nunurz/control.ts:3:15: stage 0 can't lower bind of a function that observes its argument count or collects rest arguments yet"}
count-stop production and scan-continuation control passed
```

### control-main-count-clean.log

```text
plain production count-clean: {"Completed": true, "Error": ""}
count-clean production and scan-continuation control passed
```

### control-main-rest-stop.log

```text
plain production rest-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmphxjlcqco/control.ts:3:15: stage 0 can't lower bind of a rest-parameter function yet"}
rest-stop production and scan-continuation control passed
```

### control-main-rest-clean.log

```text
plain production rest-clean: {"Completed": true, "Error": ""}
rest-clean production and scan-continuation control passed
```

### control-main-record-stop.log

```text
plain production record-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmpphsctn4g/control.ts:5:1: stage 0 can't lower a record written into a graph slot, where the write at /workspace/et209ny/tmp/tmpphsctn4g/control.ts:5:1 may close a cycle through it, in a graph program: a record can't join a graph region yet"}
record-stop production and scan-continuation control passed
```

### control-main-record-clean.log

```text
plain production record-clean: {"Completed": true, "Error": ""}
record-clean production and scan-continuation control passed
```

### control-main-cycle-stop.log

```text
plain production cycle-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmp6mk0x9lj/control.ts:3:1: stage 0 can't lower a record written into a graph slot, where the write at /workspace/et209ny/tmp/tmp6mk0x9lj/control.ts:3:1 may close a cycle through it, in a graph program: a record can't join a graph region yet"}
cycle-stop production and scan-continuation control passed
```

### control-main-cycle-clean.log

```text
plain production cycle-clean: {"Completed": true, "Error": ""}
cycle-clean production and scan-continuation control passed
```

### control-memory-count-stop.log

```text
plain production count-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmpndcp27yh/control.ts:3:15: stage 0 can't lower bind of a function that observes its argument count or collects rest arguments yet"}
count-stop production and scan-continuation control passed
```

### control-memory-count-clean.log

```text
plain production count-clean: {"Completed": true, "Error": ""}
count-clean production and scan-continuation control passed
```

### control-memory-rest-stop.log

```text
plain production rest-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmppi6j60bb/control.ts:3:15: stage 0 can't lower bind of a rest-parameter function yet"}
rest-stop production and scan-continuation control passed
```

### control-memory-rest-clean.log

```text
plain production rest-clean: {"Completed": true, "Error": ""}
rest-clean production and scan-continuation control passed
```

### control-memory-record-stop.log

```text
plain production record-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmpsy0oux12/control.ts:5:1: stage 0 can't lower a record written into a graph slot, where the write at /workspace/et209ny/tmp/tmpsy0oux12/control.ts:5:1 may close a cycle through it, in a graph program: a record of a type the program's graph doesn't select can't join a graph region yet"}
record-stop production and scan-continuation control passed
```

### control-memory-record-clean.log

```text
plain production record-clean: {"Completed": true, "Error": ""}
record-clean production and scan-continuation control passed
```

### control-memory-cycle-stop.log

```text
plain production cycle-stop: {"Completed": false, "Error": "/workspace/et209ny/tmp/tmpv3ww1619/control.ts:3:1: stage 0 can't lower a record written into a graph slot, where the write at /workspace/et209ny/tmp/tmpv3ww1619/control.ts:3:1 may close a cycle through it, in a graph program: a record of a type the program's graph doesn't select can't join a graph region yet"}
cycle-stop production and scan-continuation control passed
```

### control-memory-cycle-clean.log

```text
plain production cycle-clean: {"Completed": true, "Error": ""}
cycle-clean production and scan-continuation control passed
```

