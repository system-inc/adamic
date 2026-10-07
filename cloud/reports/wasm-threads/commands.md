# Commands and observations

Every shell sources /workspace/adamic-tools/env.sh. Native PATH remains native
clang; TestWASI alone prepends /workspace/adamic-tools/wasi-sdk/bin.
WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot.
Node v24.19.0, native clang 20.1.8, WASI SDK 27 clang 20.1.8, Go 1.27.1.
nproc=5; cgroup cpu.max=400000 100000.

Setup: bash cloud/setup.sh --wasi-sdk > /tmp/wasm-threads-setup.log 2>&1.
Printed go ready 0s; clang ready 1s; node ready 1s; wasi sdk ready 5s;
submodules ready 5s; build cache warm 209s; done 209s, 5 processors, 17.6 GB.

Full no-hooks and hooked WASI oracle, independently compiled test binaries:

```sh
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle -run '^TestWASI' > /tmp/wasm-threads-before.json 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle -run '^TestWASI' > /tmp/wasm-threads-after.json 2>&1
```

The baseline test binary started before hooks were edited and carries the
embedded baseline runtime. The hooks run compiled a separate binary afterward.
Both include every concurrency fixture in the oracle's merged fixtures list.
TestWASIEmission compiles emitted C separately and is counted separately from
TestWASIAgreesWithNode. counts.json names every failing leaf.

Strict runtime test:

```sh
ADAMIC_TEST_WASI=1 PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH go test -json -count=1 -timeout 30m ./internal/native -run '^TestWASI$' > /tmp/wasm-threads-native-wasi.json 2>&1
ADAMIC_TEST_WASI=1 ADAMIC_WASI_RUNTIME=/tmp/wasm-threads/before PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH go test -json -count=1 -timeout 30m ./internal/native -run '^TestWASI$' > /tmp/wasm-threads-before-native-wasi.json 2>&1
```

Before: unavailable pause in adamic.c. After: share.c shift-count-overflow, two
shifts by 33 on uintptr_t. TestWASI compiles strict runtime objects without
unused-code warning exemptions; no fixtures run after that compile failure.
Driver oracle: clang rejects -pthread together with -mno-atomics before runtime
compilation. Neither compile failure is counted as a killed semantic mutant.

Native oracle and TLS mutant:

```sh
ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode|TestConcurrency' > /tmp/wasm-threads-native-oracle.json 2>&1
ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/native -run '^TestParallelChecksCatchMutants/field_cache$' > /tmp/wasm-threads-tls-mutant.json 2>&1
go vet ./internal/native ./internal/oracle > /tmp/wasm-threads-vet.log 2>&1
```

Every one of the 50 runtime C files was compiled independently before and after
using native clang, no ADAMIC_TARGET_WASI, with these release flags:

```text
-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign -O2 -ffp-contract=off -pthread
-fno-optimize-sibling-calls
```

The before objects use an immutable pre-hook snapshot. All 50 objects are
byte-identical (cmp for each; SHA-256 pairs in object-identity.json).
These compare the combined runtime-area/concurrency baseline to W7, rather than
attempting to remove independent upstream runtime-area changes.

probe.dis shows TLS increment and atomic fetch-add lowered to ordinary i32
loads/adds/stores with memory-address relocations and no atomic instruction.
count.dis, heap.dis and exceptions.dis show actual runtime counter/TLS lowering.
They were compiled with WASI SDK clang --target=wasm32-wasi -mno-atomics -O2;
runtime objects also use ADAMIC_TARGET_WASI=1 and strict C11 warnings.

Mutants and limits:
- Native TLS-to-global field cache: existing field_cache mutant compiled and
  TSan caught a data race in adamic_object_data_field, exit 66.
- Attempting a WASI thread: standalone pthread-stub.c executed the real libc
  pthread_create; returned 6 (ENXIO), worker was never entered. This is a stub
  behavior probe, not a completed runtime pool mutant proof.
- Reverse WASI visitation: not executed; both production build paths remain
  blocked by out-of-territory driver/share.c failures. No kill is claimed.

No speed measurement, wasm shared-memory race test, scaling/async/moves sibling
integration, complete repository gate, or green WASI landing is claimed.
