# Commands and reproduction

Workspace `/workspace/scratch/k031ey0-5`; source repository `/workspace/adamic`.

```bash
git fetch origin main compiler/viewed-elements-nesting-5 typescript/next-stops-f44b5c27-2
git worktree add --detach /workspace/scratch/k031ey0-5/parent f44b5c272a39bcbd7cd2fb278939ac1c1832ce2e
git worktree add --detach /workspace/scratch/k031ey0-5/head e64bc123c51fc333133684f0abb3bee8e522acf3
git -C /workspace/scratch/k031ey0-5/parent submodule update --init --recursive --depth 1
# Head uses a scratch symlink to the same pinned cohere checkout.
source /workspace/adamic-tools/env.sh
mkdir -p /tmp/adamic-gate
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/k031ey0-5/upstream
npm ci --ignore-scripts --no-audit --no-fund --prefix /workspace/scratch/k031ey0-5/parent/stage3/api
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/k031ey0-5/upstream bash /workspace/scratch/k031ey0-5/parent/stage3/apply.sh /workspace/scratch/k031ey0-5/adapted
# Apply seam.diff only in head's scratch checkout, before making its overlay.
for side in parent head; do
 cd /workspace/scratch/k031ey0-5/$side
 go build -buildvcs=false -o ../$side-census ./stage3/census/tool
 python3 stage3/census/latent/make_overlay.py "$PWD" ../$side-overlay
 gofmt -w ../$side-overlay/*.go
 python3 stage3/meter/entry_overlay.py ../$side-overlay ../$side-entry
 gofmt -w ../$side-entry/*.go
 go build -buildvcs=false -overlay=/workspace/scratch/k031ey0-5/$side-entry/overlay.json -o ../$side-entry-census ./stage3/census/latent/tool
done
```

The preserved measure.py is the executable runner. Run groups separately, sequentially;
each root gets its own child, parent then head, and each child performs the tool's two
observations. Each child's command, exit code, elapsed time and box/load snapshots
are in its `*-status.json`. Timeout is 1780 seconds, followed by bounded termination.

```bash
python3 /workspace/scratch/k031ey0-5/evidence/measure.py group 1
python3 /workspace/scratch/k031ey0-5/evidence/measure.py group 2
python3 /workspace/scratch/k031ey0-5/evidence/measure.py group 3
python3 /workspace/scratch/k031ey0-5/evidence/measure.py group 4
python3 /workspace/scratch/k031ey0-5/evidence/measure.py latent parent
python3 /workspace/scratch/k031ey0-5/evidence/measure.py latent head
```

Latent children use LATENT_FULL=0, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TIMES and
LATENT_PREPASS_TIMES, as set explicitly by measure.py. The entry-overlay is the
daily meter's resolved-entry reach invocation. The seam prints the error and
continues the existing latent measurement; it does not manufacture unit stops.

Initial head ordinary build without -buildvcs=false exited 1: Git rejected the scratch submodule symlink. Both ordinary binaries were then built with -buildvcs=false; all final builds exited 0.

Continuation, per user instruction (groups of at most 10, each pushed before the next):
```bash
python3 /workspace/scratch/k031ey0-5/evidence/measure.py range 61 70
# Commit and push this group before starting range 71 79.
```

```bash
python3 /workspace/scratch/k031ey0-5/evidence/measure.py range 71 79
# Commit and push this group before either latent pass.
```
