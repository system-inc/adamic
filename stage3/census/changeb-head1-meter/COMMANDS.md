Exact compiler pins:
- A (change B): `84fccdc821a3155ae66c8137c2834eaff5c6cb69`
- B (optimizer head 1): `4956436960b289f8791e2de0a3ba5cc22faa5eed`
- Initial main/evidence parent: `470e16f8315f7a83c5c7b7f6f2a77d8328db0a25`
- Upstream TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3)
- Shared cohere: `06af2e399c6c1fb93fd570b774027942747400b4`
- Shared cohere/TypeScript: `d92d9bfee114c80be2c375d72edae966176e3a4f`

Both compiler pins already contain `stage3/census/latent/timing/`; no recorder patch was applied. Their stage3 trees are identical, and A's stage3 equals initial main's.

```bash
cd /workspace/adamic
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
git fetch origin 84fccdc821a3155ae66c8137c2834eaff5c6cb69 4956436960b289f8791e2de0a3ba5cc22faa5eed 1bab347531e8c238d97940ea843ad3c166509e18 > /workspace/scratch/fetch.log 2>&1
git worktree add --detach /workspace/scratch/compiler-A 84fccdc821a3155ae66c8137c2834eaff5c6cb69 > /workspace/scratch/worktree-A.log 2>&1
git worktree add --detach /workspace/scratch/compiler-B 4956436960b289f8791e2de0a3ba5cc22faa5eed > /workspace/scratch/worktree-B.log 2>&1
source /workspace/adamic-tools/env.sh
cd /workspace/scratch/compiler-A
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api > /workspace/scratch/npm-ci.log 2>&1
git clone --no-checkout https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream-clone.log 2>&1
git -C /workspace/scratch/upstream checkout --detach 050880ce59e30b356b686bd3144efe24f875ebc8 >> /workspace/scratch/upstream-clone.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/tree > /workspace/scratch/apply.log 2>&1
for side in A B; do
  cd /workspace/scratch/compiler-$side
  rmdir cohere
  ln -s /workspace/adamic/cohere cohere
  python3 stage3/census/latent/make_overlay.py /workspace/scratch/compiler-$side /workspace/scratch/overlay-$side > /workspace/scratch/build-$side.log 2>&1
  gofmt -w /workspace/scratch/overlay-$side/*.go >> /workspace/scratch/build-$side.log 2>&1
  go build -buildvcs=false -overlay=/workspace/scratch/overlay-$side/overlay.json -o /workspace/scratch/census-$side ./stage3/census/latent/tool >> /workspace/scratch/build-$side.log 2>&1
done
sha256sum /workspace/scratch/census-A /workspace/scratch/census-B > /workspace/scratch/binary-sha256.txt
cd /workspace/adamic
bash stage3/census/changeb-head1-meter/run-censuses.sh
python3 stage3/adapt/40-explicit-any/census-delta.py /workspace/scratch/A.jsonl /workspace/scratch/B.jsonl stage3/census/changeb-head1-meter
PYTHONDONTWRITEBYTECODE=1 python3 stage3/census/changeb-head1-meter/analyze.py
```

The first A overlay attempt ran before setup had initialized the dependency checkout and failed because the detached worktree had no cohere module. The retry above shared the initialized pinned checkout, as existing census evidence does. No files under `internal/` were edited. All overlay files live in scratch. Tree preparation and compiler builds overlapped setup; census measurement starts only after setup, both builds, and tree preparation finish. Logs were redirected directly to files, never piped through head/tail.

`/usr/bin/time` is not installed on this image. The first paired launch wrapper exited 127 before either binary was invoked. `run-censuses.sh` was then changed to use the included `time-run.py`, which records monotonic elapsed time and Linux `wait4` resource usage (peak RSS in KiB) for the timeout/census child tree. The actual census commands, environment, paired launch and 90-minute timeout are unchanged. No package installation was needed.
