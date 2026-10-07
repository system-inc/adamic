Handler benchmark: no HTTP. Native, Wasm via the Workers host, Adamic JS, and Node stripped source.
Instruments: C wait4 direct-child rusage user+system CPU and ru_maxrss; CLOCK_MONOTONIC fork-to-reap wall; ru_maxrss KiB converted to bytes.
Flags (effective): {"repo":"/tmp/adamic-handler-measure","entry":null,"corpus":null,"name":"handler","iterations":1,"rounds":3,"output":"/tmp/handler-release-proofs-alone/small-K.json","artifacts":"/tmp/handler-release-proofs-alone/build","adamic":null,"wasmOpt":"wasm-opt","node":"/workspace/adamic-tools/bin/node","timeoutMs":300000,"buildTimeoutMs":600000}
Build/runtime: {"clangOptimization":{"native":"-O2","wasm":"-Oz (entry and runtime)"},"wasmOpt":{"flags":["-Oz"],"version":"wasm-opt version 126 (version_126)"},"strip":{"tool":"/workspace/adamic-tools/wasi-sdk/bin/llvm-strip","version":"llvm-strip, compatible with GNU strip"},"releaseConfiguration":{"clang":"/workspace/adamic-tools/wasi-sdk/bin/clang","strip":"/workspace/adamic-tools/wasi-sdk/bin/llvm-strip","wasmOpt":"wasm-opt","commands":"/tmp/handler-release-proofs-alone/build/release-commands.jsonl"},"releaseReference":"aada8365085b8fea9f6cfd79230badbaa0dfce20","sanitize":false,"count":false,"nodeFlags":["--disable-warning=ExperimentalWarning"],"nodeVersion":"v24.19.0","nativeClangVersion":"clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)","wasmClangVersion":"clang version 20.1.8-wasi-sdk (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)","wasiSysroot":"/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot","helperFlags":["-O2","-std=c11","-Wall","-Wextra","-Werror","-pedantic"]}
Revisions: compiler de6e1de0b4eb3aea066df66606a46de0bd49a5ce; service f4ec96cf0a43c0b03b8f99dfcfd3bacb28e80569; host c7196b920db29c40e30210744a20a6c33a29d27f.
Machine: {"platform":"linux","release":"6.18.44","arch":"x64","cpus":5,"cpuModel":"AMD EPYC 9V74 80-Core Processor"}
Load before: 04:48:39 up 50 min,  0 users,  load average: 0.93, 1.10, 2.46; after: 04:48:45 up 51 min,  0 users,  load average: 0.93, 1.10, 2.45.
Startup: K=0 reads corpus and loads driver/handler; Wasm compiles module and creates lazy Workers host. First-call instantiation remains in K>0. none; fresh process per batch; JIT compilation and checksum cost included
Order: variants rotate each round; even rounds startup then measured, odd rounds measured then startup. Negative differences are retained as measurement noise.

Startup phases: fresh Node process: Wasm WebAssembly.compile then new Instance + _initialize using production WASI shim; JS oracle bootstrap + parse/type-strip + module load + K=0 corpus read. Native phases unavailable separately; K=0 measures exec and init.
Compression: {"instrument":"Node zlib Brotli, quality 11, generic mode; modules compressed individually","version":"1.2.0"}
Artifact definitions: native executable; Wasm reactor; emitted JS driver; Node source graph (driver, checksum, handler and local imports). Host/oracle runtime dependencies excluded.

| workload | variant | statistic | raw bytes | Brotli bytes | compile ms | instantiate + init ms | parse + module load ms | process startup K=0 ms |
|---|---|---|---:|---:|---:|---:|---:|---:|
| service | native | best of 3 | 398944 | 131297 | n/a | n/a | n/a | 0.950203 |
| service | native | median of 3 | 398944 | 131297 | n/a | n/a | n/a | 1.475383 |
| service | wasm | best of 3 | 86846 | 32293 | 0.737045 | 0.231369 | n/a | 105.891847 |
| service | wasm | median of 3 | 86846 | 32293 | 0.851054 | 0.262425 | n/a | 110.942747 |
| service | adamic-js | best of 3 | 31734 | 5649 | n/a | n/a | 4.542592 | 38.833278 |
| service | adamic-js | median of 3 | 31734 | 5649 | n/a | n/a | 5.602120 | 45.875386 |
| service | node-source | best of 3 | 10461 | 3025 | n/a | n/a | 64.991404 | 96.595031 |
| service | node-source | median of 3 | 10461 | 3025 | n/a | n/a | 67.394690 | 104.937493 |
| sieve | native | best of 3 | 352768 | 124406 | n/a | n/a | n/a | 0.926528 |
| sieve | native | median of 3 | 352768 | 124406 | n/a | n/a | n/a | 0.999926 |
| sieve | wasm | best of 3 | 45258 | 18059 | 0.847397 | 0.385030 | n/a | 82.482037 |
| sieve | wasm | median of 3 | 45258 | 18059 | 0.917973 | 0.518972 | n/a | 93.204889 |
| sieve | adamic-js | best of 3 | 10979 | 2710 | n/a | n/a | 4.297689 | 37.871652 |
| sieve | adamic-js | median of 3 | 10979 | 2710 | n/a | n/a | 5.542760 | 39.414108 |
| sieve | node-source | best of 3 | 2441 | 1010 | n/a | n/a | 54.660557 | 91.788576 |
| sieve | node-source | median of 3 | 2441 | 1010 | n/a | n/a | 69.515180 | 115.257383 |

| workload | metric | statistic | native | wasm | adamic-js | node-source |
|---|---|---|---:|---:|---:|---:|
| service | wall ms/req, startup subtracted | best of 3 | 0.054123 | -0.397483 | -0.699373 | 0.071744 |
| service | wall ms/req, startup subtracted | median of 3 | 0.077443 | -0.122957 | 0.509960 | 1.592774 |
| service | CPU ms/req, startup subtracted | best of 3 | 0.051500 | 0.227333 | -0.473833 | 0.791000 |
| service | CPU ms/req, startup subtracted | median of 3 | 0.095667 | 0.793167 | 0.309333 | 1.708000 |
| service | wall ms/req, raw | best of 3 | 0.212490 | 18.092975 | 7.251424 | 17.561326 |
| service | wall ms/req, raw | median of 3 | 0.329140 | 19.076659 | 8.155858 | 19.875166 |
| service | CPU ms/req, raw | best of 3 | 0.200667 | 21.105167 | 7.652000 | 23.483167 |
| service | CPU ms/req, raw | median of 3 | 0.317000 | 22.249000 | 8.449667 | 25.196500 |
| service | startup wall ms (K=0) | best of 3 | 0.950203 | 105.891847 | 38.833278 | 96.595031 |
| service | startup wall ms (K=0) | median of 3 | 1.475383 | 110.942747 | 45.875386 | 104.937493 |
| service | startup CPU ms (K=0) | best of 3 | 0.895000 | 124.105000 | 41.018000 | 123.455000 |
| service | startup CPU ms (K=0) | median of 3 | 1.328000 | 125.267000 | 48.842000 | 136.153000 |
| service | max RSS MiB | best of 3 | 0.890625 | 46.480469 | 28.875000 | 50.996094 |
| service | max RSS MiB | median of 3 | 1.003906 | 46.605469 | 29.000000 | 51.621094 |
| service | startup max RSS MiB | best of 3 | 0.890625 | 45.652344 | 28.500000 | 50.871094 |
| service | startup max RSS MiB | median of 3 | 1.007813 | 45.777344 | 28.625000 | 51.125000 |
| sieve | wall ms/req, startup subtracted | best of 3 | 0.930423 | 5.996576 | 5.123714 | -31.583556 |
| sieve | wall ms/req, startup subtracted | median of 3 | 1.487791 | 6.662694 | 8.125912 | 3.919579 |
| sieve | CPU ms/req, startup subtracted | best of 3 | 0.920000 | 8.529333 | 6.846333 | -31.991333 |
| sieve | CPU ms/req, startup subtracted | median of 3 | 1.469000 | 10.403333 | 10.456000 | 3.293333 |
| sieve | wall ms/req, raw | best of 3 | 1.239266 | 37.064872 | 20.749796 | 37.379002 |
| sieve | wall ms/req, raw | median of 3 | 1.985348 | 37.806146 | 21.223231 | 42.338707 |
| sieve | CPU ms/req, raw | best of 3 | 1.211667 | 44.162000 | 23.690667 | 42.931333 |
| sieve | CPU ms/req, raw | median of 3 | 1.945000 | 44.765667 | 23.901000 | 47.538333 |
| sieve | startup wall ms (K=0) | best of 3 | 0.926528 | 82.482037 | 37.871652 | 91.788576 |
| sieve | startup wall ms (K=0) | median of 3 | 0.999926 | 93.204889 | 39.414108 | 115.257383 |
| sieve | startup CPU ms (K=0) | best of 3 | 0.875000 | 96.292000 | 40.335000 | 115.917000 |
| sieve | startup CPU ms (K=0) | median of 3 | 0.928000 | 108.709000 | 41.813000 | 132.735000 |
| sieve | max RSS MiB | best of 3 | 1.757813 | 51.105469 | 36.199219 | 48.980469 |
| sieve | max RSS MiB | median of 3 | 1.757813 | 51.355469 | 36.257813 | 48.992188 |
| sieve | startup max RSS MiB | best of 3 | 0.890625 | 45.859375 | 28.500000 | 45.125000 |
| sieve | startup max RSS MiB | median of 3 | 0.917969 | 45.984375 | 28.625000 | 45.250000 |
