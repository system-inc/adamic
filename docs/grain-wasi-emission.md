# WASI emission investigation, 82jnd8v

The pilot timeout was not reproduced on current main. This branch adds phase diagnostics and makes the existing top-level test parallel. It does not claim to fix the pilot, add caches, defer work, change flags, change the fixture set, or change compiler production code. Sharding without evidence of a test-side bottleneck would not establish a cause.

Base: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2, freshly fetched origin/main, rather than the pilot's a467d1a1. This difference and the absence of the pilot's surrounding workload limit comparison. The same fixture steps are rerun in every process: lowering and clang are uncached, ADAMIC_GATE_UNCACHED=1 is set, and each clang invocation writes a new object in t.TempDir. Go's package build cache was warmed by setup; this is cold fixture execution, not a clean Go toolchain installation.

Instrument: time.Now/time.Since immediately around lowered(), native.C(), and bounded(...).CombinedOutput(), and Go's verbose subtest elapsed time. The before runs contain only this observational instrumentation; the after runs also put t.Parallel() first in the parent. No speedup is claimed. The first fresh-only run additionally used -cpuprofile /tmp/grain-wasi-before-fresh.cpu.

Build flags: Linux amd64; nproc=5; cgroup cpu.max=400000 100000 (4 CPUs); GOMAXPROCS=4; Go 1.27.1; native clang 20.1.8; WASI SDK 27 clang 20.1.8-wasi-sdk; Node v24.19.0; WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot. Normal native.Flags(wasm32-wasi), including -O2, are unchanged. Setup completed in 47.795s. Load averages were not sampled at both boundaries of each run; a post-measurement sample was 1.37 1.60 0.77.

All measurements are seconds, single runs. Paths below are under internal/oracle/testdata/.

| Fixture | Loop | Leaf | Lowered | C emission | WASI clang |
| --- | --- | ---: | ---: | ---: | ---: |
| optional_widening_fresh.a | before alone | 0.06 | 0.025 | 0.000 | 0.035 |
| optional_widening_fresh.a | before family | 0.10 | 0.040 | 0.000 | 0.059 |
| optional_widening_fresh.a | after alone | 0.07 | 0.033 | 0.000 | 0.037 |
| optional_widening_fresh.a | after family | 0.09 | 0.044 | 0.000 | 0.050 |
| optional_widening_declared.a | before alone | 0.07 | 0.028 | 0.000 | 0.038 |
| optional_widening_declared.a | before family | 0.08 | 0.045 | 0.000 | 0.037 |
| optional_widening_declared.a | after alone | 0.07 | 0.030 | 0.000 | 0.035 |
| optional_widening_declared.a | after family | 0.09 | 0.045 | 0.000 | 0.042 |
| optional_indexing_chain.a | before alone | 0.12 | 0.040 | 0.002 | 0.078 |
| optional_indexing_chain.a | before family | 0.13 | 0.058 | 0.002 | 0.073 |
| optional_indexing_chain.a | after alone | 0.12 | 0.041 | 0.002 | 0.079 |
| optional_indexing_chain.a | after family | 0.15 | 0.061 | 0.002 | 0.085 |
| optional_widening_class.a | before alone | 0.08 | 0.031 | 0.000 | 0.045 |
| optional_widening_class.a | before family | 0.10 | 0.058 | 0.000 | 0.043 |
| optional_widening_class.a | after alone | 0.07 | 0.029 | 0.000 | 0.037 |
| optional_widening_class.a | after family | 0.07 | 0.025 | 0.000 | 0.045 |
| optional_indexing_map.a | before alone | 0.11 | 0.035 | 0.001 | 0.069 |
| optional_indexing_map.a | before family | 0.12 | 0.028 | 0.001 | 0.092 |
| optional_indexing_map.a | after alone | 0.11 | 0.033 | 0.002 | 0.071 |
| optional_indexing_map.a | after family | 0.15 | 0.068 | 0.002 | 0.080 |

Commands, with /workspace/adamic-tools/env.sh sourced and the environment above:

```sh
go test -timeout 300s ./internal/oracle -run 'TestWASIEmission/internal/oracle/testdata/<slug>.a$' -parallel 1 -count=1 -v
go test -timeout 300s ./internal/oracle -run '^TestWASIEmission$' -parallel 4 -count=1 -v
go test -timeout 300s ./cmd/adamic-gate -run '^TestEveryTestIsParallelOrSaysWhy$' -count=1 -v
```

The first family attempt failed eleven leaves because pinned @types/node 25.3.3 was absent. After timeout 600 npm ci --prefix stage3/api, both complete families passed: before 29.391s, after 29.655s. Both visited exactly the same 935 registered fixture identities: 919 passed and 16 retained their existing does-not-lower skip. Maximum passing leaf: before user_iterators.a 1.51s, after user_iterators.a 1.58s. Thus no executing leaf approached 60s on this box. No new skip was added and none was removed.

The lane's parallel policy test passed. A temporary emitted-C mutant appended _Static_assert(sizeof(void *) == 8, "planted wrong WASI ABI") to optional_widening_fresh.a's generated C. Clang rejected 4 == 8 and the leaf failed with emitted C: exit status 1, proving failure propagation survived the instrumentation. The source was restored; this was an assertion error, not a warning kill.

Logs: /tmp/grain-wasi-before-*.log, /tmp/grain-wasi-after-*.log, /tmp/grain-wasi-parallel-policy.log, /tmp/grain-wasi-abi-mutant.log, /tmp/grain-wasi-setup.log. The initial missing-input attempt is /tmp/grain-wasi-before-family.log; the complete baseline is /tmp/grain-wasi-before-family-complete.log.

Unresolved: the pilot's timeout cause, its exact a467d1a1 tree and load, fresh fleet scheduling at 90s, and execution on other instances/platforms. Current main passes the requested fixture-duration threshold here; these observations cannot prove the pilot would pass after this diagnostic change.
