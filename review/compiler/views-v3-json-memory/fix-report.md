Replaced repeated element-kind stores with the existing view runtime helper.
Base: current origin/compiler/views-v3 remains 8880e6bc; fetched and merge reported already up to date.
Build and focused fixture suites passed; deterministic JSON regression passed in 1.74 seconds on four CPUs.
Restoring the stores failed the direct-store bound with 2,885 statements.
No checks, runtime files, allocation contracts or language admissions changed; no full package or gate run.

The no-array-views allocation path now calls adamic_array_view_storage instead of storing element_kind directly. The existing helper is in internal/native/runtime/view_arrays.c. Numeric kinds remain numeric; reference arrays retain the same normalized heap-pointer kind 10. The runtime's references normalization agrees with the emitter's existing IsReference normalization. The array-view path and all checks are unchanged. No helper outside view_* needs clearance.

The entire JSON translation unit differs only by replacing 2,885 stores with calls using exactly the same array and kind arguments: 2,856 number certificates and 29 heap-pointer certificates. fix-results.json records byte counts and hashes. C grows from 1,992,867 to 2,027,487 bytes, while sanitizer instrumentation becomes smaller because the actual store lives in a separate runtime translation unit.

The new TestPortElementMetadataEmission lowers the existing JSON port and bounds direct element_kind stores at zero. It also requires exactly 2,885 helper calls, preserving the allocation certificates. RSS is evidence, not a test threshold. A temporary Go overlay restores the old per-site stores; the same test fails in 2.96 seconds with "element metadata direct stores 2885 exceed bound 0". This is an assertion failure, not a compiler warning or timeout. The overlay never changes the working source.

With the same clang version, V3 runtime headers, sanitizer flags and JSON input, clang peaks at 1,233,560 KiB before and 287,616 KiB after. Compile seconds are 22.881902 before and 9.511506 after. The decrease is 945,944 KiB, about 924 MiB. The prior stripped-store diagnostic remains historical evidence, not the implemented fix. The original pool's kill is still not reproduced as a V3-only failure; this fix removes the measured compiler-memory increase without weakening the program.

Commands, with output redirected to the retained logs:

```sh
source /workspace/adamic-tools/env.sh
GOMAXPROCS=4 go test ./stage1/cohere/json -run '^TestPortElementMetadataEmission$' -count=1 -v
GOMAXPROCS=4 go test -overlay=/tmp/v3-json-measure/metadata-mutant-overlay.json ./stage1/cohere/json -run '^TestPortElementMetadataEmission$' -count=1 -v
GOMAXPROCS=4 go build -o /tmp/v3-json-measure/fixed-adamic ./cmd/adamic
/tmp/v3-json-measure/fixed-adamic c /workspace/v3-json-artifacts/v3-killed-native/003/main.ts
GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArrays$' -count=1 -parallel=4 -v
GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestArrayHolesMilestone$' -count=1 -v
```

The 23 checked-array fixtures pass in the sanitized native backend, release native backend and JavaScript backend. Fitting outcomes match Node; intentional inserted-check differences retain their pinned exit 70 diagnostics. The three hole-constructor controls also match Node in both backends and pass leak checks. Package elapsed seconds were 27.761 and 1.706 respectively. These existing tests were not edited. The only added or touched top-level test in this fix is TestPortElementMetadataEmission, passing in 1.74 seconds with GOMAXPROCS=4. No .a fixture changed or was added; no counts row moves and a-check has no new files for this fix. The three admission test files remain untouched.

Toolchain setup and nproc evidence are in the preceding measurement report: setup 118.149 seconds, nproc 5, CPU quota four, memory limit 16 GiB. This fix reused that installed toolchain. The complete measured clang command is in evidence/fix-clang.rss. No runtime file or cohere source was copied or edited.

The required integration lane command runs after committing this fix. Its output is recorded separately under this review directory and included in the delivery reply.

The additional existing full-port control `GOMAXPROCS=4 go test ./stage1/cohere/json -run '^TestSingleFileStdoutDriver$' -count=1 -v` passed in 13.33 seconds. It compiles the complete fixed JSON translation unit under sanitizers and matches both small file-driver outputs to Node and Go cohere. Its log is retained.

Committed-tip lane output:

```text
lane checks 1.7 s: gofmt and tools on 48 Go files, t.Parallel on 7 test packages; no t.Parallel analyzer on this tree; vet 7 packages
```
