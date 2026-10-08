Built: wasm32-wasi driver target, SDK setup flag, ABI compilation audit and Node WASI oracle.
Commits: implementation c0ffdf7a5ac57097f53df2556d84253738604f1d, based on main ef3d907ecdc4c771b016f7d9c52372def057a340.
Results: 299/299 emitted C units compile for wasm32; baseline execution oracle 0 pass, 299 fail; native packages and 13 selected native oracle fixtures pass.
Mutants: six native option/flag mutants, four CLI mutants, standalone Wasm stdout and exit mutants, and the existing native byte mutant were caught.
Not covered: merged-runtime execution, Wasm leaks or sanitizers, the complete repository gate, Workers hosting and the future adamic_request export.

October 7 follow-up: [request-report.md](request-report.md) records the available runtime, handler integration and new merged measurements. The blocked-runtime result and 8 MiB stack below are historical.

The agreed command is `adamic build --target wasm32-wasi <file> -o <out.wasm>`. Native stays the default. The target also works after `-o <out>`, alongside the existing build flags. Unknown or repeated targets are usage errors. WASI sanitizers, native CPU selection and tsgo archives are refused. A missing or nonexistent WASI_SYSROOT produces an explicit diagnostic.

Compilation uses `--target=wasm32-wasi --sysroot=$WASI_SYSROOT -DADAMIC_TARGET_WASI=1 -mno-atomics -O2`. The existing floating-point contraction and sibling-call restrictions remain. Linking selects `-mexec-model=command` and `-Wl,-z,stack-size=8388608`, with whole-archive runtime loading and libm. There is no pthread link flag. `main` remains the entry in generated C; the SDK supplies `_start`. The compiler prefers the SDK's bin/clang beside share/wasi-sysroot, otherwise clang on PATH. Archiving prefers llvm-ar beside the selected compiler. The runtime cache key includes target flags, sysroot path, compiler path/version and runtime source bytes.

Observed ABI audit: production emitter code has no pointer-to-integer packing, pointer-sized long declarations or printf of size_t. References use typed pointers and the reference member of adamic_value. Optional-number packing holds double/NaN bits only, not pointers. Unions use heap boxes. Indexes use size_t; proven numeric counters use int64_t to preserve the JavaScript exact-integer range on wasm32. Virtual methods cast function pointers back to their typed signature before calling. All 299 emitted C units compiled under the wasm32 ABI with the existing strict warning flags. No emission difference was needed. This compilation result does not establish runtime correctness. No runtime C or header was edited.

SDK installation and setup were run as:

```
bash cloud/setup.sh > /tmp/wasm-setup.log 2>&1
bash cloud/setup.sh --wasi-sdk > /tmp/wasm-sdk-setup.log 2>&1
source /workspace/adamic-tools/env.sh
```

The first warm-cache check overlapped my edits and failed with `vet: internal/native/library.go:29:12: undefined: ValidateOptions`. This was an edit race, not an installation failure. The second setup completed successfully. SDK 27 was downloaded from `https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-27/wasi-sdk-27.0-x86_64-linux.tar.gz` and extracted with `tar --no-same-owner -xz --strip-components 1` into `/workspace/adamic-tools/wasi-sdk`. The optional setup flag installs it and exports WASI_SYSROOT through env.sh without replacing native clang in PATH. The arm64 download path is implemented but was not tested.

Setup timings: Go ready 0s, clang ready 1s, Node ready 1s, WASI SDK ready 8s, submodules ready 8s, build cache warm 83s, done 83s. Go 1.27.1, clang 20.1.8, SDK clang 20.1.8-wasi-sdk, Node 24.19.0. `nproc` printed 5; cpu.max was `400000 100000`; reported memory was 17.6 GB. Exact lines are in setup.log.

Commands and observations, with all test output written directly to logs:

```
go test -count=1 -timeout 30m ./cmd/adamic ./internal/native > /tmp/wasm-native-tests.log 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -run 'TestWASIAgreesWithNode|TestWASIOracleCatchesMutants' -json ./internal/oracle > /tmp/wasm-oracle.json 2>&1
ADAMIC_ORACLE_WASI=1 go test -json -count=1 -timeout 30m ./internal/oracle -run '^TestWASIEmission$' > /tmp/wasm-emission.json 2>&1
ADAMIC_ORACLE_WASI=1 go test -count=1 ./internal/oracle -run '^TestWASIRunnerCatchesMutants$' > /tmp/wasm-runner-tests-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(unions|maybe_number_slots|classes|class_oct6_deep|strings|numbers)[.]a$' > /tmp/wasm-native-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run '^TestTheOracleCatchesOneByte$' > /tmp/wasm-native-mutant.log 2>&1
go test -count=1 ./cmd/adamic ./internal/native -run 'TestBuildTargetParsing|TestWASIRejectsTSGoArchive|TestWASITargetFlags|TestWASIRefusesUnsupportedOptions' > /tmp/wasm-target-tests-final.log 2>&1
go vet ./... > /tmp/wasm-vet-final.log 2>&1
gofmt -l cmd internal > /tmp/wasm-gofmt-final.log
```

Native package tests passed in 114.942s. The native fixture selector matched 13 fixtures, all passed uncached in 1.642s. The byte mutant test passed in 0.301s. Emission passed 299 fixtures in 31.939s. Standalone runner controls and mutants passed in 0.354s. Final targeted driver/native checks passed. Vet and formatting logs were empty. `bash -n cloud/setup.sh`, `node --check oracle/wasi.mjs` and `git diff --check` passed. The complete gate was not run; the touched packages and filtered oracle were used.

The baseline WASI execution run failed all 299 fixtures in 86.471s. Every failure was runtime compilation of adamic.c before execution. First diagnostics: signal.h rejects unsupported signals, signal/raise are undeclared, and struct sigaction is incomplete with SIG_DFL/SIG_IGN/SIGTERM/SIGINT/SIGHUP unavailable. These are manifestations of one runtime portability class. Later runtime translation units and program behavior were not reached, so no further failure classes are claimed. The emitted-program byte mutant was also blocked at runtime compilation and is not counted as a killed mutant.

Each native mutant was run separately with `go test -count=1 ./internal/native -run '^TestWASITargetFlags$'` or `'^TestWASIRefusesUnsupportedOptions$'`: omit ADAMIC_TARGET_WASI; accept an unknown target; accept sanitizers; accept native CPU selection; omit the missing-sysroot diagnostic; accept a nonexistent sysroot. All failed their intended tests with exit 1. The missing-sysroot mutant initially survived because the directory guard also rejected it; the check was tightened to require `requires WASI_SYSROOT`, then the same mutant failed. No compiler warning killed these mutants.

Four CLI mutants were run with `go test -count=1 ./cmd/adamic -run '^TestBuildTargetParsing$'` or `'^TestWASIRejectsTSGoArchive$'`: disable prefix target parsing, allow an unknown target name, allow a repeated target, and remove the tsgo archive refusal. Each failed its intended test with exit 1. All source mutations were restored. Full diagnostics are in mutants.log.

The standalone runner test compiled actual WASI C probes with SDK clang. Its control printed `probe\n` and exited 0, matching Node. A stdout mutant printed `probe!\n`, caught as `stdout differs`; an exit mutant printed the original bytes and returned 23, caught as `exit codes differ`. These verify the real Node WASI runner without linking Adamic's unfinished runtime. The native oracle's existing lowered dedication byte mutation was caught as `stdout differs`.

The requested integration scratch checkout is `/workspace/scratch/wasm32-driver`, branch `codex/wasm32-driver-test`, at c0ffdf7. Repeated remote checks showed no `codex/wasm32-runtime`. The explicit scratch command `git fetch origin codex/wasm32-runtime` returned exit 128: `fatal: couldn't find remote ref codex/wasm32-runtime`. Thus the requested merged-runtime pass/fail count is unavailable. The 0/299 baseline count above is explicitly without that branch. No runtime commits were merged or committed onto the driver branch.

When that branch exists, fetch it and merge origin/codex/wasm32-runtime only into the scratch branch, initialize its submodules, source env.sh and rerun the WASI execution and emitter-mutant commands above. The oracle compares stdout, stderr and exit code exactly, using the JavaScript backend for fixtures whose inserted checks fire. Wasm executions are always fresh. Host Node observations are uncached when ADAMIC_GATE_UNCACHED=1 is set. The runner provides the host filesystem as preopens for the existing filesystem fixtures.

Archived logs beside this report include compressed JSON for all WASI fixture outcomes and emitted-C outcomes. This unit adds no sanitizer or leak claim for Wasm; the native oracle retains those checks. Workers integration and adamic_request await the agreed runtime signature.
