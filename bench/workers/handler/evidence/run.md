Handler benchmark: no HTTP. Native, Wasm via the Workers host, Adamic JS, and Node stripped source.
Instruments: C wait4 direct-child rusage user+system CPU and ru_maxrss; CLOCK_MONOTONIC fork-to-reap wall; ru_maxrss KiB converted to bytes.
Flags (effective): {"repo":"/tmp/adamic-handler-measure","entry":null,"corpus":null,"name":"handler","iterations":1000,"rounds":3,"output":"/tmp/handler-release-final.json","artifacts":"/tmp/handler-release-build","adamic":null,"wasmOpt":"wasm-opt","node":"/workspace/adamic-tools/bin/node","timeoutMs":300000,"buildTimeoutMs":600000}
Build/runtime: {"clangOptimization":{"native":"-O2","wasm":"-Oz (entry and runtime)"},"wasmOpt":{"flags":["-Oz"],"version":"wasm-opt version 126 (version_126)"},"strip":{"tool":"/workspace/adamic-tools/wasi-sdk/bin/llvm-strip","version":"llvm-strip, compatible with GNU strip"},"releaseConfiguration":{"clang":"/workspace/adamic-tools/wasi-sdk/bin/clang","strip":"/workspace/adamic-tools/wasi-sdk/bin/llvm-strip","wasmOpt":"wasm-opt","commands":"/tmp/handler-release-build/release-commands.jsonl"},"releaseReference":"aada8365085b8fea9f6cfd79230badbaa0dfce20","sanitize":false,"count":false,"nodeFlags":["--disable-warning=ExperimentalWarning"],"nodeVersion":"v24.19.0","nativeClangVersion":"clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)","wasmClangVersion":"clang version 20.1.8-wasi-sdk (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)","wasiSysroot":"/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot","helperFlags":["-O2","-std=c11","-Wall","-Wextra","-Werror","-pedantic"]}
Revisions: compiler de6e1de0b4eb3aea066df66606a46de0bd49a5ce; service f4ec96cf0a43c0b03b8f99dfcfd3bacb28e80569; host c7196b920db29c40e30210744a20a6c33a29d27f.
Machine: {"platform":"linux","release":"6.18.44","arch":"x64","cpus":5,"cpuModel":"AMD EPYC 9V74 80-Core Processor"}
Load before: 04:47:12 up 49 min,  0 users,  load average: 0.79, 1.14, 2.61; after: 04:47:51 up 50 min,  0 users,  load average: 1.06, 1.17, 2.56.
Startup: K=0 reads corpus and loads driver/handler; Wasm compiles module and creates lazy Workers host. First-call instantiation remains in K>0. none; fresh process per batch; JIT compilation and checksum cost included
Order: variants rotate each round; even rounds startup then measured, odd rounds measured then startup. Negative differences are retained as measurement noise.

Startup phases: fresh Node process: Wasm WebAssembly.compile then new Instance + _initialize using production WASI shim; JS oracle bootstrap + parse/type-strip + module load + K=0 corpus read. Native phases unavailable separately; K=0 measures exec and init.
Compression: {"instrument":"Node zlib Brotli, quality 11, generic mode; modules compressed individually","version":"1.2.0"}
Artifact definitions: native executable; Wasm reactor; emitted JS driver; Node source graph (driver, checksum, handler and local imports). Host/oracle runtime dependencies excluded.

| workload | variant | statistic | raw bytes | Brotli bytes | compile ms | instantiate + init ms | parse + module load ms | process startup K=0 ms |
|---|---|---|---:|---:|---:|---:|---:|---:|
| service | native | best of 3 | 398944 | 131297 | n/a | n/a | n/a | 0.900869 |
| service | native | median of 3 | 398944 | 131297 | n/a | n/a | n/a | 1.278249 |
| service | wasm | best of 3 | 86846 | 32293 | 0.900248 | 0.286142 | n/a | 69.038901 |
| service | wasm | median of 3 | 86846 | 32293 | 0.944966 | 0.323488 | n/a | 74.417964 |
| service | adamic-js | best of 3 | 31734 | 5649 | n/a | n/a | 3.749456 | 37.217160 |
| service | adamic-js | median of 3 | 31734 | 5649 | n/a | n/a | 3.762056 | 43.259088 |
| service | node-source | best of 3 | 10455 | 3026 | n/a | n/a | 45.762083 | 87.410468 |
| service | node-source | median of 3 | 10455 | 3026 | n/a | n/a | 55.366256 | 91.099988 |
| sieve | native | best of 3 | 352768 | 124406 | n/a | n/a | n/a | 1.244512 |
| sieve | native | median of 3 | 352768 | 124406 | n/a | n/a | n/a | 1.305970 |
| sieve | wasm | best of 3 | 45258 | 18059 | 0.665784 | 0.249427 | n/a | 112.674488 |
| sieve | wasm | median of 3 | 45258 | 18059 | 0.717833 | 0.331089 | n/a | 135.519207 |
| sieve | adamic-js | best of 3 | 10979 | 2710 | n/a | n/a | 3.341461 | 33.999082 |
| sieve | adamic-js | median of 3 | 10979 | 2710 | n/a | n/a | 3.528173 | 35.525441 |
| sieve | node-source | best of 3 | 2435 | 1010 | n/a | n/a | 38.412867 | 86.118051 |
| sieve | node-source | median of 3 | 2435 | 1010 | n/a | n/a | 38.966081 | 87.696869 |

| workload | metric | statistic | native | wasm | adamic-js | node-source |
|---|---|---|---:|---:|---:|---:|
| service | wall ms/req, startup subtracted | best of 3 | 0.020042 | 0.024900 | 0.011370 | 0.004481 |
| service | wall ms/req, startup subtracted | median of 3 | 0.020284 | 0.030743 | 0.014072 | 0.005249 |
| service | CPU ms/req, startup subtracted | best of 3 | 0.020029 | 0.036748 | 0.017542 | 0.010839 |
| service | CPU ms/req, startup subtracted | median of 3 | 0.020280 | 0.043316 | 0.019712 | 0.011352 |
| service | wall ms/req, raw | best of 3 | 0.020255 | 0.037303 | 0.017573 | 0.020288 |
| service | wall ms/req, raw | median of 3 | 0.020435 | 0.043741 | 0.021282 | 0.020432 |
| service | CPU ms/req, raw | best of 3 | 0.020233 | 0.051793 | 0.024043 | 0.030451 |
| service | CPU ms/req, raw | median of 3 | 0.020404 | 0.058899 | 0.027340 | 0.031010 |
| service | startup wall ms (K=0) | best of 3 | 0.900869 | 69.038901 | 37.217160 | 87.410468 |
| service | startup wall ms (K=0) | median of 3 | 1.278249 | 74.417964 | 43.259088 | 91.099988 |
| service | startup CPU ms (K=0) | best of 3 | 0.744000 | 85.438000 | 39.006000 | 114.596000 |
| service | startup CPU ms (K=0) | median of 3 | 1.222000 | 90.266000 | 45.763000 | 116.236000 |
| service | max RSS MiB | best of 3 | 0.906250 | 60.109375 | 38.675781 | 60.371094 |
| service | max RSS MiB | median of 3 | 0.914063 | 60.480469 | 38.773438 | 61.371094 |
| service | startup max RSS MiB | best of 3 | 0.898438 | 45.675781 | 28.625000 | 50.746094 |
| service | startup max RSS MiB | median of 3 | 1.015625 | 45.855469 | 28.750000 | 51.371094 |
| sieve | wall ms/req, startup subtracted | best of 3 | 0.774589 | 1.236263 | 0.555966 | 0.543063 |
| sieve | wall ms/req, startup subtracted | median of 3 | 0.829349 | 1.379223 | 0.601258 | 0.593166 |
| sieve | CPU ms/req, startup subtracted | best of 3 | 0.770403 | 1.256858 | 0.565602 | 0.553890 |
| sieve | CPU ms/req, startup subtracted | median of 3 | 0.826971 | 1.395161 | 0.612685 | 0.601920 |
| sieve | wall ms/req, raw | best of 3 | 0.775024 | 1.288274 | 0.567299 | 0.571769 |
| sieve | wall ms/req, raw | median of 3 | 0.829764 | 1.424397 | 0.613100 | 0.622398 |
| sieve | CPU ms/req, raw | best of 3 | 0.770811 | 1.313357 | 0.577619 | 0.586282 |
| sieve | CPU ms/req, raw | median of 3 | 0.827358 | 1.445463 | 0.625346 | 0.636262 |
| sieve | startup wall ms (K=0) | best of 3 | 1.244512 | 112.674488 | 33.999082 | 86.118051 |
| sieve | startup wall ms (K=0) | median of 3 | 1.305970 | 135.519207 | 35.525441 | 87.696869 |
| sieve | startup CPU ms (K=0) | best of 3 | 1.160000 | 124.908000 | 36.050000 | 97.178000 |
| sieve | startup CPU ms (K=0) | median of 3 | 1.202000 | 150.905000 | 37.982000 | 103.026000 |
| sieve | max RSS MiB | best of 3 | 1.757813 | 56.109375 | 109.636719 | 140.437500 |
| sieve | max RSS MiB | median of 3 | 1.769531 | 56.355469 | 109.804688 | 140.800781 |
| sieve | startup max RSS MiB | best of 3 | 0.875000 | 45.859375 | 28.375000 | 45.121094 |
| sieve | startup max RSS MiB | median of 3 | 0.875000 | 45.980469 | 28.500000 | 45.199219 |
