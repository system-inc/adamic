# Compiler and stage1 comparison shards

Base: `39856c192241f80e0e871a36052ba6ae1b8d2deb`. Code: `41c841049`.

Only `TestCompilerAndStage1Agree` and its executable adapter changed. The pin check, file walk and count log are byte-identical to the base. The Go oracle remains one unsharded `execute` call. Source Node, emitted JavaScript and sanitized native use `shards.Run` with `min(runtime.NumCPU(), 8)`. The adapters retain regular-file stdout capture, nonzero-exit checks, failure on stderr, and a 600-second backend timeout. The merger still rejects missing, repeated and extra cases.

All 868 compiler/stage1 files agree byte for byte: 30,325,158 bytes on all three port backends against Go.

| Backend | Shard 0 | Shard 1 | Shard 2 | Shard 3 | Shard 4 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Node | 14.408s | 16.324s | 13.146s | 27.591s | 13.619s |
| emitted JavaScript | 27.964s | 25.107s | 27.950s | 66.226s | 26.161s |
| sanitized native | 190.225s | 204.961s | 192.284s | 439.400s | 185.586s |

Test wall time: 557.42s. Slowest native shard: 439.400s, 160.600s below the 600-second timeout. The native batch, including launch/output merging, took 440.267s, leaving a conservative 159.733s margin.

This cloud box exposes five CPUs, with a four-CPU quota (`cpu.max=400000 100000`): all five native shards initially used about 80% of a CPU. Load samples are saved. Arbitrary integration load is not bounded by this measurement. The supplied seat baseline was 452.7s for the test and 335.6s for native. The local test is slower than that baseline; the machines and load differ, so these are not a same-machine speed comparison.

The first full-package attempt was terminated by an environment refresh during JSX validation, without a test failure or package result. Its log is explicitly labeled interrupted. The complete retry ran in a detached runner and supplied every timing above.

## Failure proof

A scratch Go overlay changes exactly one byte in the first payload line of native shard zero, after the case header. Case numbers and manifest cardinality remain valid, the sanitized binary exits successfully, and the ordinary Go comparison rejects the merged output with `sanitized native: case 0`. The Go test exits 1. The overlay and planted edit were never committed; the code file remained unchanged. See `proof.log` and `proof-wall.json`.

## Package gate

`go test ./stage1/cohere/lint -count=1 -v -json -timeout=60m`

PASS, exit 0, 2918.937s process wall time. TypeScript 6.0.3 is pinned at `050880ce59e30b356b686bd3144efe24f875ebc8` and `git status --porcelain --ignored` was empty before and after. WASI SDK 27 reports clang 20.1.8-wasi-sdk. Both profile variables used the same fresh directory: `/workspace/scratch/compiler-shards-profile.2huf_ti1`.

Only the opt-in throughput benchmarks skipped: `TestThroughput`, `TestJsxLintReleaseAndThroughput`. No correctness test skipped.

The shard merger package and command-diagnostics smoke check also passed. Setup timing lines, SDK version, raw gate, proof and load logs are saved beside this report. No full repository gate was run.
