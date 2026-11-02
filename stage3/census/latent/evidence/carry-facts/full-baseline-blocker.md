# Carry facts: blocked at baseline

Stopped at step 2 under the requested stopping rule. No compiler or census implementation was changed. No carry is justified by this evidence.

The full-mode, unchanged baseline was killed by SIGKILL (subprocess return code -9) after **2409.29095837 s** wall time. Peak child RSS was **16616968 KiB** (15.85 GiB). The cgroup memory limit is **17179869184 bytes** (16 GiB), memory.peak was 17179987968 bytes, and memory.events reported `oom 2`, `oom_kill 2`. These observations indicate memory exhaustion; the census produced no error text. Setup and overlay builds had finished before the baseline started; the scratch profiling build also finished well before the kill.

The incomplete ledger contains **4666 attempted units**, **2114.170802176 s** summed unit wall time and **3305.753860000 s** summed unit CPU time. The final row is `/workspace/scratch/tree/src/compiler/factory/parenthesizerRules.ts:675:1`. `timing-table.csv` retains every completed baseline timing; new timings, drops, and ratios are blank because there is no completed baseline or changed run. There is no valid top-20 comparison.

`latent.jsonl` contains **one line only**, the measurement/checker header: the driver buffers file observations until completion. Its partial-file SHA-256 is `526b9beb73d3e860c70eba756d7219a39bdbd51636d15ae679621395efcdd508`. This is **not** a baseline findings hash or a successful identity proof. No new `latent.jsonl` exists.

Step 3 was not reached. A scratch-only profiling binary and whole-program walker timers/answer probes were prepared, but never run. There are no measured program-wide costs/counts, no per-unit answer equality proof, no carry change, and no guard/mutant result. Tests and vet were not run because work stopped at the baseline, before implementation. The incomplete output and scratch instrumentation remain under `/workspace/scratch` and are not committed.

## Machine and pins

- Host: `98b2b7d59531`; `nproc`: 5; setup reports 17.6 GB memory; cgroup CPU quota `400000 100000` (4 CPUs), memory limit 16 GiB; no swap.
- Main fetched from `https://github.com/system-inc/adamic.git`: `9c607b11c6d7a6afad29853db5d86b488baaa08e`.
- Cohere: `06af2e399c6c1fb93fd570b774027942747400b4`.
- Cohere TypeScript: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Upstream and adapted-tree HEAD: TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Go: `go1.27.1 linux/amd64`; Node: `v24.19.0`; clang: `20.1.8`.
- Setup succeeded in 908.993 s. API npm install and adapted-tree application succeeded. All command output went to files, without piping logs into head or tail.

## Commands

From `/workspace/adamic`:

```sh
git fetch origin main
git switch -c typescript/census-carry-facts origin/main
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api > /workspace/scratch/api-npm.log 2>&1
git clone --branch v6.0.3 --depth 1 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream-clone.log 2>&1
git -C /workspace/scratch/upstream rev-parse HEAD > /workspace/scratch/upstream.sha
test "$(cat /workspace/scratch/upstream.sha)" = 050880ce59e30b356b686bd3144efe24f875ebc8
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/tree > /workspace/scratch/apply.log 2>&1
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/baseline-overlay > /workspace/scratch/baseline-overlay.log 2>&1
gofmt -w /workspace/scratch/baseline-overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/baseline-overlay/overlay.json -o /workspace/scratch/baseline-census ./stage3/census/latent/tool > /workspace/scratch/baseline-build.log 2>&1
```

The first overlay-generation attempt, made while setup was still fetching, failed because `cohere/TypeScript/tsc/go.mod` did not yet exist. The same command succeeded after submodule checkout completed.

GNU `/usr/bin/time` was absent (exit 127 before starting the census), so the actual baseline used this Python wrapper, preserving the meter's census arguments and environment:

```python
import json, os, pathlib, resource, subprocess, time
out = pathlib.Path('/workspace/scratch/baseline')
env = dict(os.environ, LATENT_TIMES=str(out/'latent-times.jsonl'),
           LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1')
start = time.monotonic()
with (out/'run.log').open('w') as log:
    result = subprocess.run([
        '/workspace/scratch/baseline-census',
        '/workspace/scratch/tree/src/compiler', str(out/'latent.jsonl')
    ], env=env, stdout=log, stderr=subprocess.STDOUT)
(out/'resources.json').write_text(json.dumps(dict(
    wall_seconds=time.monotonic()-start,
    peak_rss_kib=resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss,
    exit_code=result.returncode), indent=2)+'\n')
raise SystemExit(result.returncode)
```

No retry with changed GC or memory settings was made: the instruction was to stop when a step could not complete as written. Completing the measurements requires enough memory for the unchanged baseline, or an explicitly revised measurement procedure.
