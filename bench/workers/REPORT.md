Built the direct workerd harness, Node adapter, three fixtures and reproducible source mutants.
Base: origin/main e8ba3d5; branch: codex/workers-bench. Final commit is recorded in the delivery report.
Validation: final workerd and Node proofs, three-round CLI run, and filtered uncached oracle all exited 0.
Mutants: wrong CPU PID, changed response and removed warmup were each caught by their intended check.
Not covered: macOS execution, complete repository gate, real compute variants, compiler changes or deployed Cloudflare cold starts.

## Environment and setup

Linux 6.18.44, AMD EPYC 9V74, Node 24.19.0, Go 1.27.1, clang 20.1.8.
`nproc` printed 5; cgroup CPU quota is 4 CPUs.
`bash cloud/setup.sh > /tmp/workers-setup.log 2>&1` exited 0. Timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (139s)
setup: done in 139s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Every test shell sourced `/workspace/adamic-tools/env.sh`. npm initially lacked a
writable default cache; `/tmp/workers-npm-cache` resolved that. Registry lookup
returned newest version `1.20261007.1`. This command exited 0, adding two packages
in 7 seconds:

```sh
npm --cache /tmp/workers-npm-cache install --prefix /tmp/workers-runtime workerd@1.20261007.1 > /tmp/workers-install.log 2>&1
```

`workerd --version` printed `workerd 2026-10-07`; the harness also verified the
adjacent npm package version is exactly `1.20261007.1`.

## Final failure proofs

Both commands exited 0 and wrote `passed: true` in their `proofs.json`:

```sh
node bench/workers/prove.mjs workerd /tmp/workers-runtime/node_modules/.bin/workerd /tmp/workers-proofs-landed > /tmp/workers-proofs-landed.log 2>&1
node bench/workers/prove.mjs node workerd /tmp/workers-proofs-node-final > /tmp/workers-proofs-node-final.log 2>&1
```

| runner | concurrency | fast CPU ms/req | CPU fixture ms/req | idle drift ms/req | required gap ms/req |
|---|---:|---:|---:|---:|---:|
| workerd | 1 | 0.225 | 6.100 | 0.000 | 0.500 |
| workerd | 16 | 0.125 | 5.350 | 0.000 | 0.500 |
| Node adapter, NOT workerd | 1 | 0.500 | 5.575 | 0.050 | 0.500 |
| Node adapter, NOT workerd | 16 | 0.450 | 5.825 | 0.000 | 0.500 |

The wrong-PID source mutant read an unrelated idle child. The CPU separation
assertion failed, although responses and timing completed. The changed-response
variant returned `wrong` on `/echo`; correctness refused it before timing, with
zero measured and cold samples. The removed-warmup source mutant changed only
the actual warmup phase, leaving declared flags intact.

| runner | warmed first-eight median ms | mutant first-eight median ms | mutant later median ms | actual warmup normal/mutant |
|---|---:|---:|---:|---|
| workerd | 0.212 | 15.241 | 0.213 | 16 / 0 |
| Node adapter, NOT workerd | 0.450 | 17.359 | 0.449 | 16 / 0 |

The first-round request samples expose the first-use penalty directly. Each
mutant is saved as actual altered harness source beneath its proof directory;
every corresponding JSON contains all measured request latencies and counters.

Additional witnesses passed: measured counts and rotating variant order;
Worker-visible URL independent of ephemeral port; byte-correct UTF-8 POST body;
observed 3:1 weighted status sequence at both concurrencies; and workerd loading
an actual `.wasm` module whose return value matched Node WebAssembly API (42).

## CLI table and raw output

This command exited 0. It produced 36 measured cells, 2,304 raw measured requests,
and nine independent cold samples. These short fixture runs validate the harness;
they do not establish performance of a production Worker.

```sh
node bench/workers/run.mjs --workerd /tmp/workers-runtime/node_modules/.bin/workerd --requests 64 --warmup 16 --rounds 3 --cold-spawns 3 --output /tmp/workers-table.json > /tmp/workers-table.md 2>&1
```

Runner: workerd directly, no wrangler or proxy in request path. Version: workerd 2026-10-07; npm package: 1.20261007.1.
Instruments: Linux /proc/<pid>/stat utime+stime, CLK_TCK=100; /proc/<pid>/status VmRSS and VmHWM.
Flags (effective): {"runner":"workerd","workerd":"/tmp/workers-runtime/node_modules/.bin/workerd","suite":"/workspace/adamic/bench/workers/fixtures/suite.json","output":"/tmp/workers-table.json","requests":64,"warmup":16,"rounds":3,"coldSpawns":3,"concurrency":[1,16],"compatibilityDate":"2026-10-01","timeoutMs":10000,"idleSettleMs":100,"memoryPollMs":10}
Machine: {"platform":"linux","arch":"x64","node":"v24.19.0","release":"6.18.44","cpus":5,"cpuModel":"AMD EPYC 9V74 80-Core Processor"}
Load before: 04:08:03 up 10 min,  0 users,  load average: 0.54, 1.22, 1.06; after: 04:08:21 up 10 min,  0 users,  load average: 0.53, 1.17, 1.04.
Cold path: /health. Default HTTP Host: bench.invalid. Weighted selection: deterministic index modulo total weight.

| variant | workload | c | statistic across rounds | p50 ms | p90 ms | p99 ms | mean ms | CPU ms/req | idle drift ms/req | idle RSS MiB | peak RSS MiB |
|---|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|
| fast | echo | 1 | best of 3 | 0.278 | 0.432 | 0.887 | 0.349 | 0.156 | 0.000 | 43.773 | 45.156 |
| fast | echo | 1 | median of 3 | 0.289 | 0.461 | 1.927 | 0.369 | 0.313 | 0.000 | 43.875 | 45.203 |
| fast | echo | 16 | best of 3 | 2.023 | 3.372 | 4.427 | 2.244 | 0.156 | 0.000 | 43.496 | 45.453 |
| fast | echo | 16 | median of 3 | 2.961 | 5.234 | 6.187 | 2.944 | 0.156 | 0.000 | 43.793 | 45.656 |
| fast | mixed | 1 | best of 3 | 0.251 | 0.292 | 0.647 | 0.263 | 0.156 | 0.000 | 44.152 | 45.359 |
| fast | mixed | 1 | median of 3 | 0.251 | 0.373 | 0.775 | 0.271 | 0.313 | 0.000 | 44.266 | 45.473 |
| fast | mixed | 16 | best of 3 | 1.354 | 2.683 | 3.150 | 1.749 | 0.000 | 0.000 | 43.758 | 45.543 |
| fast | mixed | 16 | median of 3 | 2.041 | 3.107 | 3.591 | 1.821 | 0.156 | 0.000 | 44.074 | 45.906 |
| cpu | echo | 1 | best of 3 | 5.788 | 6.496 | 8.001 | 5.886 | 5.625 | 0.000 | 43.730 | 54.301 |
| cpu | echo | 1 | median of 3 | 6.012 | 8.070 | 25.385 | 6.899 | 6.250 | 0.000 | 43.793 | 54.410 |
| cpu | echo | 16 | best of 3 | 88.997 | 94.823 | 98.951 | 88.330 | 5.625 | 0.000 | 43.176 | 54.586 |
| cpu | echo | 16 | median of 3 | 89.355 | 98.683 | 101.189 | 88.496 | 5.625 | 0.000 | 43.738 | 55.008 |
| cpu | mixed | 1 | best of 3 | 5.855 | 6.543 | 6.951 | 5.940 | 5.781 | 0.000 | 43.609 | 54.219 |
| cpu | mixed | 1 | median of 3 | 5.932 | 6.710 | 8.156 | 6.029 | 5.781 | 0.000 | 43.613 | 54.273 |
| cpu | mixed | 16 | best of 3 | 79.693 | 100.505 | 163.610 | 87.637 | 5.781 | 0.000 | 43.422 | 54.676 |
| cpu | mixed | 16 | median of 3 | 88.219 | 102.558 | 180.460 | 88.108 | 5.938 | 0.000 | 43.516 | 54.828 |
| allocate | echo | 1 | best of 3 | 0.634 | 1.173 | 4.322 | 0.945 | 0.781 | 0.000 | 43.457 | 109.078 |
| allocate | echo | 1 | median of 3 | 0.786 | 1.380 | 5.455 | 0.972 | 0.781 | 0.000 | 43.754 | 109.461 |
| allocate | echo | 16 | best of 3 | 9.108 | 17.093 | 18.150 | 10.126 | 0.625 | 0.000 | 43.617 | 109.805 |
| allocate | echo | 16 | median of 3 | 10.738 | 18.682 | 18.801 | 10.871 | 0.781 | 0.000 | 44.035 | 110.086 |
| allocate | mixed | 1 | best of 3 | 0.724 | 1.210 | 1.603 | 0.827 | 0.625 | 0.000 | 43.129 | 108.695 |
| allocate | mixed | 1 | median of 3 | 0.760 | 1.445 | 2.196 | 0.828 | 0.625 | 0.000 | 43.613 | 109.180 |
| allocate | mixed | 16 | best of 3 | 4.210 | 21.348 | 25.716 | 8.074 | 0.625 | 0.000 | 43.512 | 109.578 |
| allocate | mixed | 16 | median of 3 | 9.635 | 23.025 | 31.948 | 12.486 | 0.781 | 0.000 | 44.086 | 110.086 |

| variant | cold spawns | best ms | median ms |
|---|---:|---:|---:|
| fast | 3 | 23.120 | 26.444 |
| cpu | 3 | 26.062 | 26.690 |
| allocate | 3 | 22.121 | 23.468 |

Raw report: `/tmp/workers-table.json`. Proof raw reports and mutant sources:
`/tmp/workers-proofs-landed/` and `/tmp/workers-proofs-node-final/`.

## Repository checks and limits

```sh
ADAMIC_GATE_UNCACHED=1 go test ./bench ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/modules/main.a$' -count=1 -timeout 30m > /tmp/workers-filtered-oracle.log 2>&1
```

Exited 0: bench has no Go test files; internal/oracle passed in 12.351s.
`node --check` passed for harness, CLI, adapter and proof runner.
`git diff --check` passed. No compiler files changed. The full repository gate
was not run; only the named filtered uncached oracle was run.

The README distinguishes Linux VmHWM from macOS sampled RSS and explains CPU
resolution, idle drift, independent best/median statistics, closed-loop latency
and local spawn-to-response cold start. macOS is implemented but untested here.
The fixed cold path must be cheap. Correctness covers the explicit checklist.
No assertion of a compiler speedup follows from these synthetic measurements.
