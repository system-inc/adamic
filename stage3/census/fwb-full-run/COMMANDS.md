# Commands and pins

Repository: https://github.com/system-inc/adamic.git

- Main: `141ad1d244d6a2d9c5403858086d505dfdef5fd5`
- Fix: `caa19bcd7ff9ad977e6e72fde659acade7f0c4ab`
- TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`
- Cohere: `06af2e399c6c1fb93fd570b774027942747400b4`
- Cohere TypeScript: `d92d9bfee114c80be2c375d72edae966176e3a4f`

Read all of `stage3/meter/twice-daily.sh` first. The existing clone was clean and at the pinned main commit.

```bash
cd /workspace/adamic
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
git fetch origin typescript/latent-private-ir-fields
git worktree add --detach /workspace/scratch/main-source 141ad1d244
git worktree add --detach /workspace/scratch/fix-source caa19bcd7ff9ad977e6e72fde659acade7f0c4ab
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/tree > /workspace/scratch/apply.log 2>&1
bash /workspace/scratch/build-entry.sh > /workspace/scratch/build-entry.log 2>&1
```

`build-entry.sh` contains the exact per-worktree submodule and overlay build commands. `run-census.sh` records the two sequential census commands with their flags, bounds, log paths and Bash timing. These were launched separately, waiting for main to exit before starting fix.

An initial launch used `/usr/bin/time -p -o /workspace/scratch/results/main.time` before `timeout`. `/usr/bin/time` was absent, so this launch exited 127 before running the census. Bash built-in `time` was used for both actual runs.

The statecopy test command is:

```bash
cd /workspace/scratch/fix-source
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
go test -count=1 ./stage3/census/latent/statecopy/
```

Results are compressed with Python `gzip.compress(data, mtime=0)` and analyzed with `python3 stage3/census/fwb-full-run/analyze.py`.
