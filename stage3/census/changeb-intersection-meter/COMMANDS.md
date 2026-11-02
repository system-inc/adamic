# Commands and pins

A: `1b516b1b53b2a49abd619f6c220516091222b350`.
B: `021e43f3e40d76fee81a62942d3b4280f846647e` (sole parent A).
Upstream TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Unit-times reference: `1bab347531e8c238d97940ea843ad3c166509e18`.
Both pins already contain the recorder and have identical `stage3` trees; no recorder patch was applied.

Initial checkout/setup source: `80c2d3282e44b527555d59690bb413cc883d2bf2`.
Commands below ran in `/workspace/adamic` unless a different directory is stated.

```bash
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
git fetch origin 1b516b1b53b2a49abd619f6c220516091222b350 021e43f3e40d76fee81a62942d3b4280f846647e 1bab347531e8c238d97940ea843ad3c166509e18 > /workspace/scratch/fetch-inputs.log 2>&1
git worktree add --detach /workspace/scratch/compiler-A 1b516b1b53b2a49abd619f6c220516091222b350 > /workspace/scratch/worktree-A.log 2>&1
git worktree add --detach /workspace/scratch/compiler-B 021e43f3e40d76fee81a62942d3b4280f846647e > /workspace/scratch/worktree-B.log 2>&1
git worktree add -b typescript/changeb-intersection-meter /workspace/scratch/evidence 1b516b1b53b2a49abd619f6c220516091222b350 > /workspace/scratch/evidence-worktree.log 2>&1
cd /workspace/scratch/compiler-A
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api > /workspace/scratch/npm-ci.log 2>&1
git clone --branch v6.0.3 --single-branch https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream-clone.log 2>&1
```

The adapter, exact meter overlay builds, and paired run commands are in [build.sh](build.sh) and [run-censuses.sh](run-censuses.sh). Execution status and binary hashes are recorded in REPORT.md and the accompanying artifacts. These scripts are not evidence of execution on their own.

The initial full-history upstream clone was canceled (exit 143) after its pack download exceeded 2 GB. Git removed the incomplete destination during cancellation. A subsequent attempted move of that already-removed directory failed without modifying anything. The source clone was then completed using the same tag with shallow history:

```bash
git clone --depth 1 --branch v6.0.3 --single-branch https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream-shallow-clone.log 2>&1
```

Setup exited 0 in 965.589 seconds.

Execution from `/workspace/scratch/evidence`:

```bash
bash stage3/census/changeb-intersection-meter/build.sh > /workspace/scratch/build-driver.log 2>&1
source /workspace/adamic-tools/env.sh
node stage3/census/changeb-intersection-meter/direct-calls.cjs /workspace/scratch/evidence/stage3/census/changeb-intersection-meter/direct-calls.json > /workspace/scratch/direct-calls.log 2>&1
bash stage3/census/changeb-intersection-meter/run-censuses.sh > /workspace/scratch/run-driver.log 2>&1
```

The first invocation of the run script exited 127 because `/usr/bin/time` was absent; neither compiler binary started. The saved `launch-attempt-*` logs and exits record that attempt. The runner was changed to `python3 "$out/run-resource.py" "$out/$side-resource.json" timeout --kill-after=60s 5400 ...`, then invoked with the same command. This successful invocation started A and B concurrently; both exited 0.

Comparison commands (before raw ledgers were gzipped):

```bash
export PYTHONDONTWRITEBYTECODE=1
out=/workspace/scratch/evidence/stage3/census/changeb-intersection-meter
python3 /workspace/scratch/compiler-A/stage3/adapt/40-explicit-any/census-delta.py "$out/A.jsonl" "$out/B.jsonl" "$out" > "$out/delta.log" 2>&1
python3 "$out/analyze.py" > "$out/analysis.log" 2>&1
python3 "$out/render-details.py" > "$out/render.log" 2>&1
sha256sum /workspace/scratch/census-A /workspace/scratch/census-B > "$out/binary-sha256.txt"
```

`analyze.py` invokes the pinned meter's `lowering_summary` directly with all 79 expected source paths, checks identical checker headers/unit sets, and checks one timing per attempted unit. It also accepts gzipped inputs. `view-rows.py` regenerates the unchanged and changed view/predicate rows from the compressed ledgers.
