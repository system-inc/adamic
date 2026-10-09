# Commands actually run

All measurements sent stdout/stderr to logs. Paths below are the preserved
scratch paths on this worker. Read-only exploration and failed setup attempts
are described in REPORT.md rather than presented as successful checks.

```sh
git fetch origin refs/heads/main:refs/remotes/origin/main \
  refs/heads/codex/stage3-explicit-any-2:refs/remotes/origin/codex/stage3-explicit-any-2 \
  refs/heads/cloud/land-area-next:refs/remotes/origin/cloud/land-area-next > /tmp/scanner-any-fetch.log 2>&1
git merge --no-edit origin/main > /tmp/scanner-any-main-merge.log 2>&1
git worktree add -b scratch/scanner-any-main /workspace/scanner-any-main origin/main > /tmp/scanner-any-worktree.log 2>&1
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scanner-any-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
```

From /workspace/scanner-any-main:

```sh
STAGE3_CACHE=/workspace/scratch/native3-cache bash stage3/lane/run.sh /workspace/scratch/scanner-any-main-lane > /tmp/scanner-any-main-lane.log 2>&1
```

From /workspace/adamic, the first 42 run failed under overlapping memory load;
the second ran alone and passed. Each lane invoked its unchanged apply.sh and
oracle/run.sh. Both used all runners, eight workers and no filter.

```sh
STAGE3_CACHE=/workspace/scratch/native3-cache bash stage3/lane/run.sh /workspace/scratch/scanner-any-42-lane > /tmp/scanner-any-42-lane.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache bash stage3/lane/run.sh /workspace/scratch/scanner-any-42-lane-solo > /tmp/scanner-any-42-lane-solo.log 2>&1
NODE_PATH=/workspace/scratch/native3-cache/api/node_modules node stage3/adapt/42-scanner-any/prove.cjs /workspace/scratch/scanner-any-main-lane/adapted-tree /workspace/scratch/scanner-any-42-lane-solo/adapted-tree stage3/drivers/scanner/main.a /workspace/scratch/scanner-any-owner-proof-final > /tmp/scanner-any-owner-proof-final.log 2>&1
NODE_PATH=/workspace/scratch/native3-cache/api/node_modules node stage3/adapt/42-scanner-any/adapt.cjs /workspace/scratch/native3-slice > /tmp/scanner-any-slice-idempotent.log 2>&1
```

The compiler command below ran through Python subprocess.run, with monotonic
wall time and RUSAGE_CHILDREN recorded in compiler-time.json. Its cwd was the
existing, never-pushed scratch checkout /workspace/scanner-native3-next at
28285421. No compiler source change was made.

```sh
go build -buildvcs=false -o /workspace/scratch/scanner-any-next-adamic ./cmd/adamic > /tmp/scanner-any-compiler-build.log 2>&1
```

From /workspace/adamic:

```sh
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=0 bash stage3/drivers/scanner/run.sh /workspace/scratch/scanner-any-native-0 --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-any-next-adamic --compiler-cwd /workspace/scanner-native3-next > /tmp/scanner-any-native-0.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 bash stage3/drivers/scanner/run.sh /workspace/scratch/scanner-any-native-1 --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-any-next-adamic --compiler-cwd /workspace/scanner-native3-next > /tmp/scanner-any-native-1.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache bash stage3/drivers/scanner/run.sh /workspace/scratch/scanner-any-full-node --tree /workspace/scratch/scanner-any-42-lane/adapted-tree --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --node-only > /tmp/scanner-any-full-node.log 2>&1
python3 /tmp/scanner-any-checkpoints.py > /tmp/scanner-any-checkpoints.log 2>&1
python3 /tmp/scanner-any-witnesses.py > /tmp/scanner-any-witnesses-final.log 2>&1
bash -n stage3/drivers/scanner/adapt-slice.sh
node --check stage3/adapt/42-scanner-any/adapt.cjs
node --check stage3/adapt/42-scanner-any/prove.cjs
git diff --check
```

The recorded one-shot checkpoint and witness scripts are under
stage3/drivers/scanner/evidence/scanner-any. Their child build/diff/Node commands
and the initial/final placeholder patches are retained there. The checkpoint
preparation is detailed in REPLAY.md; it is separate from the real scanner run.
Final evidence audit compared lane counts, titles, sanctions and baseline diff,
all 725 built JavaScript/declaration hashes, every source file (only scanner.ts
changed), and counted both Node passes directly. No full Go-package test or
full gate was run as confirmation.
