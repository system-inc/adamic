Commands used (all logs redirected to files; no compiler changes committed):

```bash
cd /workspace/adamic
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
source /workspace/adamic-tools/env.sh
git fetch origin optimizer/branded-supertype-road2 > /workspace/scratch/fetch-b.log 2>&1
# HEAD was 141ad1d244d6a2d9c5403858086d505dfdef5fd5 during npm ci and apply.
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api > /workspace/scratch/npm-ci.log 2>&1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream-clone.log 2>&1
test "$(git -C /workspace/scratch/upstream rev-parse HEAD)" = 050880ce59e30b356b686bd3144efe24f875ebc8
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/tree > /workspace/scratch/apply.log 2>&1

git worktree add --detach /workspace/scratch/compiler-A 141ad1d244d6a2d9c5403858086d505dfdef5fd5 > /workspace/scratch/worktree-A.log 2>&1
git worktree add --detach /workspace/scratch/compiler-B 9e4f2184fd0d887bb582294b2ff91b1683c9e2f0 > /workspace/scratch/worktree-B.log 2>&1
for side in A B; do
  worktree=/workspace/scratch/compiler-$side
  overlay=/workspace/scratch/overlay-$side
  binary=/workspace/scratch/census-$side
  rmdir "$worktree/cohere"
  ln -s /workspace/adamic/cohere "$worktree/cohere"
  cd "$worktree"
  python3 stage3/census/latent/make_overlay.py "$worktree" "$overlay" > /workspace/scratch/overlay-$side.log 2>&1
  gofmt -w "$overlay"/*.go
  go build -buildvcs=false -overlay="$overlay/overlay.json" -o "$binary" ./stage3/census/latent/tool > /workspace/scratch/build-$side.log 2>&1
  sha256sum "$binary"
done
cd /workspace/adamic
git switch -c typescript/changeb-road2-meter
bash stage3/census/changeb-road2-meter/run-censuses.sh > /workspace/scratch/paired-runner.log 2>&1
# The saved Python runner starts these two commands concurrently, with the same
# environment except the timing output path. wait4 records peak RSS, user/system
# CPU and monotonic elapsed time. SIGTERM at 5400s; SIGKILL after 60s grace.
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/workspace/scratch/paired/A-times.jsonl /workspace/scratch/census-A /workspace/scratch/tree/src/compiler /workspace/scratch/paired/A.jsonl
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/workspace/scratch/paired/B-times.jsonl /workspace/scratch/census-B /workspace/scratch/tree/src/compiler /workspace/scratch/paired/B.jsonl

python3 stage3/adapt/40-explicit-any/census-delta.py /workspace/scratch/paired/A.jsonl /workspace/scratch/paired/B.jsonl stage3/census/changeb-road2-meter > /workspace/scratch/delta.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/census/changeb-road2-meter/analyze.py > /workspace/scratch/analysis.log 2>&1
```

The worktrees were prepared while setup was downloading submodules; binary builds began after setup passed. Binary builds and adaptation overlapped. No build or adaptation workload overlapped the paired censuses. Both SHAs already contain identical `stage3/census/latent/timing/` and meter code; no recorder backport was applied. Nothing under `internal/` was changed. Overlays are outside the worktrees.

An initial wrapper tried `/usr/bin/time -v`, which is absent on this box. Both shell launches exited 127 before either census executed. Those logs are retained in `launch-failed/`; the actual measurement used the saved Python runner's `wait4` accounting.

Evidence-only commit and foreground push:

```bash
git add stage3/census/changeb-road2-meter
git diff --cached --check
git commit -m "Record paired road-2 B versus A latent census on tsc"
git push origin typescript/changeb-road2-meter
```
