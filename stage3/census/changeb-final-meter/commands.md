Commands executed from `/workspace/adamic` (the checkout starts at A, `470e16f8315f7a83c5c7b7f6f2a77d8328db0a25`):

```bash
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
git fetch origin 84fccdc821a3155ae66c8137c2834eaff5c6cb69 1bab347531e8c238d97940ea843ad3c166509e18 > /workspace/scratch/fetch.log 2>&1
git switch -c typescript/changeb-final-meter
source /workspace/adamic-tools/env.sh
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api > /workspace/scratch/npm-ci.log 2>&1
git clone --branch v6.0.3 --depth 1 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream-clone.log 2>&1
test "$(git -C /workspace/scratch/upstream rev-parse HEAD)" = 050880ce59e30b356b686bd3144efe24f875ebc8
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/tree > /workspace/scratch/apply.log 2>&1
git worktree add --detach /workspace/scratch/compiler-A 470e16f8315f7a83c5c7b7f6f2a77d8328db0a25 > /workspace/scratch/worktree-A.log 2>&1
git worktree add --detach /workspace/scratch/compiler-B 84fccdc821a3155ae66c8137c2834eaff5c6cb69 > /workspace/scratch/worktree-B.log 2>&1
for side in A B; do
  wt=/workspace/scratch/compiler-$side
  test "$(git -C "$wt" rev-parse HEAD:cohere)" = "$(git -C /workspace/adamic/cohere rev-parse HEAD)"
  rmdir "$wt/cohere"
  ln -s /workspace/adamic/cohere "$wt/cohere"
  cd "$wt"
  python3 stage3/census/latent/make_overlay.py "$wt" /workspace/scratch/overlay-$side > /workspace/scratch/overlay-$side.log 2>&1
  gofmt -w /workspace/scratch/overlay-$side/*.go
  go build -buildvcs=false -overlay=/workspace/scratch/overlay-$side/overlay.json -o /workspace/scratch/census-$side ./stage3/census/latent/tool > /workspace/scratch/build-$side.log 2>&1
done
sha256sum /workspace/scratch/census-A /workspace/scratch/census-B > /workspace/scratch/binary-sha256s.txt
cd /workspace/adamic
bash stage3/census/changeb-final-meter/run-censuses.sh
python3 stage3/adapt/40-explicit-any/census-delta.py /workspace/scratch/changeb-final-meter/A.jsonl /workspace/scratch/changeb-final-meter/B.jsonl stage3/census/changeb-final-meter
python3 stage3/census/changeb-final-meter/analyze.py
```

The two worktrees were prepared while setup was running. The binary build ran concurrently with adaptation; both finished before either census started. No timing patch was applied: both exact SHAs already contain timing tree `b12676c48d04c7930df05eb0984be1fbba0f7a13`. The dependency symlink follows the meter's existing runner and points to the verified shared cohere pin. Nothing under `internal/` was edited.

An initial runner attempted `/usr/bin/time -v`, which is unavailable on this box; it exited 127 before either binary ran. Its logs are retained under `launch-failed/`. The actual runner is `run-censuses.py`, started by `run-censuses.sh`. It launches both commands with `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1` and independent `LATENT_TIMES` paths, and uses `os.wait4` for wall time, CPU and peak RSS. Each child is bounded by `timeout --kill-after=5s 5400`. The resource JSONs record the literal argv. No census result or timing was collected by the failed launch.

Supplementary source inspection and assembly after both censuses finished:

```bash
source /workspace/adamic-tools/env.sh
node stage3/census/changeb-final-meter/inspect-claims.cjs
node stage3/census/changeb-final-meter/cast-callers.cjs
PYTHONDONTWRITEBYTECODE=1 python3 stage3/census/changeb-final-meter/assemble.py
git add stage3/census/changeb-final-meter
git commit -m "Record paired final change B latent census against A"
git push origin typescript/changeb-final-meter
```

The analysis was corrected to select `KindImportDeclaration` / `KindExportDeclaration` / `KindInterfaceDeclaration` / `KindTypeAliasDeclaration` (the literal meter kind names) after an initial renderer-only empty-controls error. Both completed census streams were reused unchanged. The raw SHA-256s in census-delta.json match the decompressed committed ledgers.
