Consumers now build missing predecessors through the producer recipe and key.
typescript/stage3-lane-products-4 is one lane-only commit on 5f38748e; its final tip is recorded in delivery.
Fresh-cache install builds all six apply products and succeeds; apply 99 then reuses the same key.
The independent producer-key comparison passes; changing the consumer to product70 fails it.
Only stage3/lane changes; final-shard cold proof and wave2 wrappers are in the child branch.

# Producer reuse

adaptedTree calls the existing exported applyproducts.Product(t,5), the current product99 slot. No producer API edits, input recipe, key, flags or worker implementation are copied. TestStage3LaneAdaptedTreeMatchesProducer compares the consumer directory against applyproducts.Get(), the producer's final-product resolver. The wrong-slot mutant selects70 and fails with both product key directories named. The producer boundary index is therefore guarded against drift.

All lane predecessor resolution happens inside the buildcache.Product callback. A hit returns immediately; a miss calls the same laneProduct recipe for its predecessor. A shard uses laneProduct for harness. Nothing relies on a preexisting recorded read set. There is no fetch-only error callback.

# Fresh-cache proof and all walls

| Command | Exit | Process wall seconds |
|---|---:|---:|
| `["go", "test", "./stage3/lane", "-run", "^TestProduct_Stage3LaneInstall$", "-count=1", "-timeout", "14m", "-v"]` | 0 | 231.668250 |
| `["go", "test", "./stage3/applyproducts", "-run", "^TestProduct_Stage3Adapted99$", "-count=1", "-timeout", "90s", "-v"]` | 0 | 1.222069 |
| `["go", "test", "./stage3/lane", "-run", "^TestStage3LaneAdaptedTreeMatchesProducer$", "-count=1", "-timeout", "90s", "-v"]` | 0 | 1.872135 |
| `["go", "test", "./stage3/lane", "-run", "^TestStage3LaneAdaptedTreeMatchesProducer$", "-count=1", "-timeout", "90s", "-v"]` | 1 | 1.920472 |

The cache directory did not exist before the coordinator created it. Install logs show misses for every apply boundary10,20,40,65,70,99 and the lane install. The following apply99 command shows the same99 key with hit and no miss. This cold chain is the user-authorized exception to the usual90s/60s unit limits; its Go deadline is14m with an860-second subprocess bound. Normal prepared checks use90s.

# Setup

```text
submodule cache: off
v24.19.0
go version go1.27.1 linux/amd64
setup: go ready (0.024s)
setup: node ready (0.024s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.012s
setup: markdown dependencies ready (0.080s)
setup: submodules ready (0.082s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.190s)
setup: shared cache ready (1.159s)
setup: go build ready (19.225s)
setup: test binaries deferred (use --warm-tests) (19.315s)
setup: build cache warm (19.317s)
setup: build-flags commit=5f38748e9453e7ca5cb4dc549d0b0b8d2c585662 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.25 0.06 0.02 1/911 139173 load-after=1.71 0.40 0.13 1/915 139794
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (19.344s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.BpsjKh

```

nproc5 with a four-CPU cgroup quota. Supplied TypeScript source is /tmp/hefhych-typescript; source and NodeProject installs were reused without network fetches inside products. Prior disposable task cache products were cleared to provide inodes for the fresh chain; logs and previously committed evidence were preserved.

Existing runtime, payload, phase, inventory, baseline, API, count, exit and deadline checks are unchanged. No stage3/applyproducts file is edited. This base still has the prior apply implementation; its parallel git-read fix will be supplied by the owner's re-cut. No Workshop trace, full compiler gate, or actual Loom first-tree build is claimed.
