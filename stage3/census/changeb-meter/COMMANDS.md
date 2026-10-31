# Commands and provenance

- Initial checkout/setup SHA: `21ba732affe8f460276d5ac9e11fb682210dd5cb`.
- Main tip fetched for adaptation: `4e45d80b59bdd03c054d7611c50e6b89cace1a0b`.
- A: `e0391d48aab3227d3aa5f4c93da39e3fbb2513be`.
- B: `8c7afca2230e38952f3bf51357338a6b0f21ea8f`.
- Identical meter-only recorder patch on A and B: `1bab347531e8c238d97940ea843ad3c166509e18`.
- TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Shared cohere: `06af2e399c6c1fb93fd570b774027942747400b4`; its TypeScript submodule: `d92d9bfee114c80be2c375d72edae966176e3a4f`.

```bash
# /workspace/adamic
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
git fetch origin main > /workspace/scratch/fetch-main.log 2>&1
git fetch origin e0391d48aab3227d3aa5f4c93da39e3fbb2513be 8c7afca2230e38952f3bf51357338a6b0f21ea8f 1bab347531e8c238d97940ea843ad3c166509e18 > /workspace/scratch/fetch-compilers.log 2>&1
git worktree add --detach /workspace/scratch/main-source 4e45d80b59bdd03c054d7611c50e6b89cace1a0b
# /workspace/scratch/main-source
export PATH=/workspace/adamic-tools/go/bin:/workspace/adamic-tools/bin:$PATH
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api > /workspace/scratch/npm-main.log 2>&1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream-clone.log 2>&1
test "$(git -C /workspace/scratch/upstream rev-parse HEAD)" = 050880ce59e30b356b686bd3144efe24f875ebc8
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/tree > /workspace/scratch/apply.log 2>&1
# /workspace/adamic
git worktree add --detach /workspace/scratch/change-a e0391d48aab3227d3aa5f4c93da39e3fbb2513be
git worktree add --detach /workspace/scratch/change-b 8c7afca2230e38952f3bf51357338a6b0f21ea8f
git diff --binary 1bab347531e8c238d97940ea843ad3c166509e18^ 1bab347531e8c238d97940ea843ad3c166509e18 -- stage3/census/latent stage3/meter > /workspace/scratch/unit-times.patch
git -C /workspace/scratch/change-a apply /workspace/scratch/unit-times.patch
git -C /workspace/scratch/change-b apply /workspace/scratch/unit-times.patch
# Each worktree's empty cohere directory was replaced with a symlink to the setup checkout (identical gitlink).
rmdir /workspace/scratch/change-a/cohere
ln -s /workspace/adamic/cohere /workspace/scratch/change-a/cohere
rmdir /workspace/scratch/change-b/cohere
ln -s /workspace/adamic/cohere /workspace/scratch/change-b/cohere
# In each respective compiler worktree, with setup's Go on PATH and GOPROXY above:
python3 stage3/census/latent/make_overlay.py /workspace/scratch/change-a /workspace/scratch/overlay-a > /workspace/scratch/overlay-a.log 2>&1
gofmt -w /workspace/scratch/overlay-a/*.go
go build -buildvcs=false -overlay=/workspace/scratch/overlay-a/overlay.json -o /workspace/scratch/census-a ./stage3/census/latent/tool > /workspace/scratch/build-a.log 2>&1
python3 stage3/census/latent/make_overlay.py /workspace/scratch/change-b /workspace/scratch/overlay-b > /workspace/scratch/overlay-b.log 2>&1
gofmt -w /workspace/scratch/overlay-b/*.go
go build -buildvcs=false -overlay=/workspace/scratch/overlay-b/overlay.json -o /workspace/scratch/census-b ./stage3/census/latent/tool > /workspace/scratch/build-b.log 2>&1
bash /workspace/scratch/run-paired.sh
python3 /workspace/scratch/main-source/stage3/adapt/40-explicit-any/census-delta.py A.jsonl B.jsonl .
python3 analyze.py
```

Setup, source adaptation, and binary preparation overlapped; measurements began only after all completed. An initial A overlay attempt used the unrelated system `go`; a second preceded submodule availability. Both failed before building a binary and were retried with setup's Go and the pinned submodule. The first timing patch export omitted `--binary` and failed atomically; the binary patch above applied successfully to both worktrees. An extra npm install ran in the initial checkout; adaptation used the separate main-tip checkout and its own npm install. No compiler sources under `internal/` were edited.

The shell launcher’s first attempt found `/usr/bin/time` absent (exit 127 on both sides, before either binary started). The actual measurement command was:

```bash
python3 stage3/census/changeb-meter/run-paired.py
```

`run-paired.py` starts both `timeout --kill-after=60s 5400 env ...` equivalents before waiting, then uses `os.wait4` for each child's wall time, user/system CPU and peak RSS (KiB). The environment variables are passed through `Popen(env=...)`. `run-paired.sh` is retained as the unsuccessful GNU-time launcher, not the measured launcher. The existing meter buffers per-file census records until all callers complete; `*.times.jsonl` provides live unit progress if the 90-minute bound is hit.
