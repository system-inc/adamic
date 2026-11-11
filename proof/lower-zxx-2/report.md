# Compiler proof #zxxp2h4 — for-of writes

Focused pins PASS; all four specified mutants are caught; full lower PASS; complete admission delta PASS. The full oracle command FAILS with 14 environment/resource failures: twelve package-resolution failures and the isolated counts-probe failure passed on targeted retries; the WASI-sysroot failure remains. The full oracle command was not rerun, so this is not an all-green full-suite result.

## Decisive results

| Check | Exit | Wall seconds | Result |
| --- | ---: | ---: | --- |
| Uncached forof_writes native oracle | 0 | 335.583 | PASS |
| Focused lower | 0 | 30.410 | PASS |
| Full lower | 0 | 1845.816 | PASS |
| Full oracle | 1 | 4099.915 | FAIL (inventory below) |
| Complete admission delta, workers 2 | 0 | 1381.264 | pass |

Admission: `void=false`; void reasons: none; `diff_count=4`; `admitted=0`; newly refused: **0**; classification stderr differences: **0**. No classification or runtime timeout occurred, so no admission timeout rerun at either SHA was necessary.

## Revisions and setup

- HEAD: `73211726b0674668dc2335d81af630a244127e69` (detached during all proof runs).
- Base: `16429a30452d027809f0ea85bc395479423e493e`, verified equal to `HEAD^`.
- Source branch fetched: `origin compiler/forof-writes-2`.
- Report branch: `proof/lower-zxx-2-report`, created from the tested HEAD; only `proof/lower-zxx-2/report.md` is to be committed.

Executed, sequentially:

```sh
git fetch origin compiler/forof-writes-2
git checkout --detach 73211726b0674668dc2335d81af630a244127e69
git submodule update --init --recursive
```

Submodule pins: `cohere` = `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` = `d92d9bfee114c80be2c375d72edae966176e3a4f`. No AGENTS.md was found in the workspace.

Loaded `/workspace/adamic-tools/env.sh` before the work. Created `/tmp/adamic-gate` outside the repository before Go ran. Environment for every proof command:

```sh
source /workspace/adamic-tools/env.sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-proof-zxx/typescript
export PATH=/tmp/adamic-proof-zxx/api/node_modules/.bin:$PATH
```

Upstream TypeScript checkout: `/tmp/adamic-proof-zxx/typescript`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, fetched from `https://github.com/microsoft/TypeScript.git`. The initial shared clone lacked this upstream commit; an exact shallow fetch supplied it.

Copied stage3/api’s package.json and package-lock.json to `/tmp/adamic-proof-zxx/api` and ran `npm ci --ignore-scripts --prefix /tmp/adamic-proof-zxx/api`. Packages: TypeScript `6.0.3`, @types/node `25.3.3` (and the lockfile’s dependency). `tsc --version` printed `Version 6.0.3`. All npm and admission children inherited the exported CA settings with strict SSL enabled. Go: `go1.27.1 linux/amd64`; Node: `v24.19.0`.

No two heavy commands were started concurrently. The gate ran after the test commands and their active children exited. A defunct clang process was present, with no running work. All evidence/logs stayed outside the repository until this final report.

## 1. Formatting and vet

`gofmt -l internal/lower internal/oracle`: exit 0, **empty output**.

`go vet ./internal/lower ./internal/oracle`: exit 0, wall 254.686 seconds, empty diagnostics.

`/usr/bin/time` was unavailable, so the first timed vet invocation did not start vet (shell exit 127). The successful rerun used Python `time.monotonic()` around the subprocess. All recorded wall times below use that method and include Go build/startup overhead.

## 2. Focused pins


```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/forof_writes/ -count=1 -v -timeout 40m
```

Exit 0; wall 335.583 seconds. Every selected test/subtest result:

```text
--- PASS: TestNativeAgreesWithNode_023/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a (312.51s)
--- PASS: TestNativeAgreesWithNode_065/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts (312.53s)
--- PASS: TestNativeAgreesWithNode_071/internal/oracle/testdata/forof_writes/r1_reassigned.a (312.59s)
--- PASS: TestNativeAgreesWithNode_081/internal/oracle/testdata/forof_writes/r1_reassigned.ts (312.65s)
```


```sh
go test ./internal/lower -run 'ForOf|LocalWritesAreLocalFieldsWrites|LocalFieldsClassifyEveryIntField|Spread' -count=1 -v
```

Exit 0; wall 30.410 seconds. Every selected test/subtest result:

```text
--- PASS: TestLocalWritesAreLocalFieldsWrites (0.00s)
--- PASS: TestForOfReassignedBindingIsNoPlainSpreadSource (0.00s)
--- PASS: TestLocalFieldsClassifyEveryIntField (0.04s)
--- PASS: TestMixedTupleSpreadIsNotYet (0.17s)
--- PASS: TestLibrarySpreadDateView (0.17s)
--- PASS: TestForOfReassignedSpreadSourceIsNotPlain (0.18s)
--- PASS: TestForOfConstBindingStillDispatches (0.20s)
--- PASS: TestNullableObjectSpreadKeepsNarrowerLiteralRefused (0.26s)
--- PASS: TestOptionalWideningSpreadOverwrite (0.13s)
--- PASS: TestLibrarySpreadPlainProducers (0.16s)
--- PASS: TestLibrarySpreadTypedArrayView (0.19s)
--- PASS: TestLibrarySpreadMapView (0.11s)
--- PASS: TestLibrarySpreadSetView (0.15s)
--- PASS: TestForOfReassignedPatternBindingIsUnknown (0.11s)
--- PASS: TestForOfReassignedBindingIsUnknown (0.13s)
--- PASS: TestCallableContractSpreadRefused (0.13s)
--- PASS: TestLibrarySpreadErrorView (0.13s)
--- PASS: TestLibrarySpreadRegExpUnion (0.07s)
--- PASS: TestClassFeaturesThrowingAccessorSpread (0.40s)
--- PASS: TestCallableContractPlainSpread (0.46s)
--- PASS: TestGeneratorVoidReturnIsNeverReadBySpreadOrArrayFrom (1.05s)
--- PASS: TestInputSpreadCoverage (0.01s)
--- PASS: TestInputSpreadCoverage/utf8At_alias.a (0.10s)
--- PASS: TestInputSpreadCoverage/fileStatus_alias.a (0.12s)
--- PASS: TestInputSpreadCoverage/writeTextFile_tuple.a (0.08s)
--- PASS: TestInputSpreadCoverage/writeTextFile_parentheses.a (0.07s)
--- PASS: TestInputSpreadCoverage/writeTextFile_nested.a (0.12s)
--- PASS: TestInputSpreadCoverage/writeTextFile_multiple.a (0.09s)
--- PASS: TestInputSpreadCoverage/writeTextFile_local.a (0.09s)
--- PASS: TestInputSpreadCoverage/writeTextFile_later.a (0.10s)
--- PASS: TestInputSpreadCoverage/writeTextFile_array.a (0.11s)
--- PASS: TestInputSpreadCoverage/utf8Length_tuple.a (0.09s)
--- PASS: TestInputSpreadCoverage/writeTextFile_alias.a (0.13s)
--- PASS: TestInputSpreadCoverage/utf8Length_nested.a (0.10s)
--- PASS: TestInputSpreadCoverage/utf8Length_parentheses.a (0.12s)
--- PASS: TestInputSpreadCoverage/utf8Length_array.a (0.07s)
--- PASS: TestInputSpreadCoverage/utf8Length_empty.a (0.14s)
--- PASS: TestInputSpreadCoverage/utf8Length_local.a (0.16s)
--- PASS: TestInputSpreadCoverage/utf8Length_alias.a (0.12s)
--- PASS: TestInputSpreadCoverage/utf8At_parentheses.a (0.08s)
--- PASS: TestInputSpreadCoverage/utf8At_nested.a (0.14s)
--- PASS: TestInputSpreadCoverage/utf8At_tuple.a (0.17s)
--- PASS: TestInputSpreadCoverage/utf8At_multiple.a (0.13s)
--- PASS: TestInputSpreadCoverage/utf8At_local.a (0.13s)
--- PASS: TestInputSpreadCoverage/utf8At_later.a (0.15s)
--- PASS: TestInputSpreadCoverage/readDirectory_alias.a (0.06s)
--- PASS: TestInputSpreadCoverage/readTextFile_tuple.a (0.08s)
--- PASS: TestInputSpreadCoverage/utf8At_array.a (0.12s)
--- PASS: TestInputSpreadCoverage/readTextFile_local.a (0.09s)
--- PASS: TestInputSpreadCoverage/readTextFile_parentheses.a (0.14s)
--- PASS: TestInputSpreadCoverage/readTextFile_array.a (0.07s)
--- PASS: TestInputSpreadCoverage/readTextFile_nested.a (0.15s)
--- PASS: TestInputSpreadCoverage/readDirectory_parentheses.a (0.09s)
--- PASS: TestInputSpreadCoverage/readTextFile_alias.a (0.14s)
--- PASS: TestInputSpreadCoverage/readDirectory_tuple.a (0.18s)
--- PASS: TestInputSpreadCoverage/readDirectory_nested.a (0.18s)
--- PASS: TestInputSpreadCoverage/readDirectory_local.a (0.15s)
--- PASS: TestInputSpreadCoverage/programArguments_alias.a (0.39s)
--- PASS: TestInputSpreadCoverage/readDirectory_array.a (0.43s)
--- PASS: TestInputSpreadCoverage/programArguments_parentheses.a (0.35s)
--- PASS: TestInputSpreadCoverage/programArguments_nested.a (0.38s)
--- PASS: TestInputSpreadCoverage/programArguments_tuple.a (0.41s)
--- PASS: TestInputSpreadCoverage/programArguments_local.a (0.11s)
--- PASS: TestInputSpreadCoverage/fileStatus_nested.a (0.12s)
--- PASS: TestInputSpreadCoverage/fileStatus_parentheses.a (0.06s)
--- PASS: TestInputSpreadCoverage/programArguments_array.a (0.13s)
--- PASS: TestInputSpreadCoverage/fileStatus_tuple.a (0.10s)
--- PASS: TestInputSpreadCoverage/fileStatus_local.a (0.15s)
--- PASS: TestInputSpreadCoverage/fileStatus_array.a (0.16s)
--- PASS: TestSpreadRecordsTheGettersItMayRun (0.00s)
--- PASS: TestSpreadRecordsTheGettersItMayRun/a_getter_of_a_class_no_Tree_can_be (0.05s)
--- PASS: TestSpreadRecordsTheGettersItMayRun/no_getter_anywhere (0.13s)
--- PASS: TestSpreadRecordsTheGettersItMayRun/an_object_literal's_getter (0.14s)
--- PASS: TestSpreadRecordsTheGettersItMayRun/a_getter_inherited_from_a_base (0.11s)
--- PASS: TestSpreadRecordsTheGettersItMayRun/a_class's_getter (0.10s)
```

The slash filter selects exactly the four forof_writes fixture subtests across the generated TestNativeAgreesWithNode shard family. Empty parent wrappers also pass; they are not additional fixture coverage. There were no baseline FAILs and therefore no missing baseline got/want.

## 3. Mutants, each applied alone and restored

Each old fenced Go fragment from `internal/lower/forof_writes.mutants.md` was asserted to occur exactly once before replacement. Each mutant reran **both** commands from step 2, with `ADAMIC_GATE_UNCACHED=1` for the oracle. Before each mutation the tree passed `git diff --exit-code`; the exact original file contents were restored in a finally block, and the same diff check passed afterward.

| Mutant | Old-text matches | Oracle exit / wall s | Lower exit / wall s | Restore diff exit |
| --- | ---: | --- | --- | ---: |
| 1a | 1 | 1 / 84.650 | 1 / 27.003 | 0 |
| 1b | 1 | 1 / 65.409 | 1 / 36.277 | 0 |
| 2 | 1 | 0 / 89.568 | 1 / 26.606 | 0 |
| 3 | 1 | 0 / 71.492 | 1 / 32.371 | 0 |

All four mutants are caught by the requested lower command. Mutants 1a and 1b also make their respective oracle fixtures fail on both backends. Mutant 2’s dedicated catch pin was run additionally because the requested lower regex omits Catch.

### Mutant 1a

File: `internal/lower/view_flow.go`. Old:

```go
	"ForOf.Local":    "a for-of binding takes each element",
```

New:
Delete the old line.

**oracle: exit 1; wall 84.650 seconds.**

Every red test/subtest:

```text
TestNativeAgreesWithNode_081
TestNativeAgreesWithNode_081/internal/oracle/testdata/forof_writes/r1_reassigned.ts
TestNativeAgreesWithNode_071
TestNativeAgreesWithNode_071/internal/oracle/testdata/forof_writes/r1_reassigned.a
```

Complete diagnostic text, including all emitted got/want and stderr (only scheduling messages and passing result lines omitted):

```text
=== NAME  TestNativeAgreesWithNode_065/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 97dba30c4c4f miss 34.40
=== NAME  TestNativeAgreesWithNode_071/internal/oracle/testdata/forof_writes/r1_reassigned.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 97dba30c4c4f hit 34.40
=== NAME  TestNativeAgreesWithNode_023/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 97dba30c4c4f hit 34.40
=== NAME  TestNativeAgreesWithNode_081/internal/oracle/testdata/forof_writes/r1_reassigned.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 97dba30c4c4f hit 34.40
=== NAME  TestNativeAgreesWithNode_081/internal/oracle/testdata/forof_writes/r1_reassigned.ts
    oracle_test.go:838: the release build: exit codes differ
        sanitized: exit 1, stdout "", stderr "=================================================================\n==9663==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b9cb25e0040 at pc 0x55f853970f09 bp 0x7ffd5868a4f0 sp 0x7ffd5868a4e8\nREAD of size 8 at 0x7b9cb25e0040 thread T0\n    #0 0x55f853970f08 in adamic_initialize_r1_reassigned_ts_ba7890163e0b7fb5aff28559c549aa26 /adamic-build/main.c:92:197\n    #1 0x55f853970a8d in adamic_initialize_r1_reassigned_ts_a29e3c1a2a9b87ae033c80e6da4992ab /adamic-build/main.c:117:2\n    #2 0x55f853970a52 in main /adamic-build/main.c:127:2\n    #3 0x7f5cb3585ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #4 0x7f5cb3585d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x55f85388a590 in _start (/tmp/adamic-gate/TestNativeAgreesWithNode_081internaloracletestdataforof_writ28398556/002/program+0x80590)\n\n0x7b9cb25e0040 is located 0 bytes after 48-byte region [0x7b9cb25e0010,0x7b9cb25e0040)\nallocated by thread T0 here:\n    #0 0x55f85392d3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55f8539985ef in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55f8539985ef in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55f853970a8d in adamic_initialize_r1_reassigned_ts_a29e3c1a2a9b87ae033c80e6da4992ab /adamic-build/main.c:117:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:92:197 in adamic_initialize_r1_reassigned_ts_ba7890163e0b7fb5aff28559c549aa26\nShadow bytes around the buggy address:\n  0x7b9cb25dfd80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9cb25dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9cb25dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9cb25dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9cb25dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n=>0x7b9cb25e0000: fa fa 00 00 00 00 00 00[fa]fa fa fa fa fa fa fa\n  0x7b9cb25e0080: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9cb25e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9cb25e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9cb25e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9cb25e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==9663==ABORTING\n"
        release:   exit 0, stdout "0\n1\n", stderr ""
    oracle_test.go:856: exit codes differ
        node:   exit 0, stdout "42\n1\n", stderr ""
        native: exit 1, stdout "", stderr "=================================================================\n==9663==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b9cb25e0040 at pc 0x55f853970f09 bp 0x7ffd5868a4f0 sp 0x7ffd5868a4e8\nREAD of size 8 at 0x7b9cb25e0040 thread T0\n    #0 0x55f853970f08 in adamic_initialize_r1_reassigned_ts_ba7890163e0b7fb5aff28559c549aa26 /adamic-build/main.c:92:197\n    #1 0x55f853970a8d in adamic_initialize_r1_reassigned_ts_a29e3c1a2a9b87ae033c80e6da4992ab /adamic-build/main.c:117:2\n    #2 0x55f853970a52 in main /adamic-build/main.c:127:2\n    #3 0x7f5cb3585ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #4 0x7f5cb3585d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x55f85388a590 in _start (/tmp/adamic-gate/TestNativeAgreesWithNode_081internaloracletestdataforof_writ28398556/002/program+0x80590)\n\n0x7b9cb25e0040 is located 0 bytes after 48-byte region [0x7b9cb25e0010,0x7b9cb25e0040)\nallocated by thread T0 here:\n    #0 0x55f85392d3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55f8539985ef in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55f8539985ef in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55f853970a8d in adamic_initialize_r1_reassigned_ts_a29e3c1a2a9b87ae033c80e6da4992ab /adamic-build/main.c:117:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:92:197 in adamic_initialize_r1_reassigned_ts_ba7890163e0b7fb5aff28559c549aa26\nShadow bytes around the buggy address:\n  0x7b9cb25dfd80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9cb25dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9cb25dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9cb25dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9cb25dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n=>0x7b9cb25e0000: fa fa 00 00 00 00 00 00[fa]fa fa fa fa fa fa fa\n  0x7b9cb25e0080: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9cb25e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9cb25e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9cb25e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9cb25e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==9663==ABORTING\n"
    oracle_test.go:862: JavaScript backend: stdout differs
        node:    exit 0, stdout "42\n1\n", stderr ""
        backend: exit 0, stdout "undefined\n1\n", stderr ""
--- FAIL: TestNativeAgreesWithNode_081 (0.00s)
    --- FAIL: TestNativeAgreesWithNode_081/internal/oracle/testdata/forof_writes/r1_reassigned.ts (44.93s)
=== NAME  TestNativeAgreesWithNode_071/internal/oracle/testdata/forof_writes/r1_reassigned.a
    oracle_test.go:838: the release build: exit codes differ
        sanitized: exit 1, stdout "", stderr "=================================================================\n==9662==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7be7d2be0040 at pc 0x55eec8b10f09 bp 0x7ffca98cb340 sp 0x7ffca98cb338\nREAD of size 8 at 0x7be7d2be0040 thread T0\n    #0 0x55eec8b10f08 in adamic_initialize_r1_reassigned_a_dcf141a7ab71bf906cd99ed3e9007ab7 /adamic-build/main.c:92:195\n    #1 0x55eec8b10a8d in adamic_initialize_r1_reassigned_a_b0cabbd5f785d390af5fc51d0fdeada5 /adamic-build/main.c:117:2\n    #2 0x55eec8b10a52 in main /adamic-build/main.c:127:2\n    #3 0x7fa7d3b43ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #4 0x7fa7d3b43d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x55eec8a2a590 in _start (/tmp/adamic-gate/TestNativeAgreesWithNode_071internaloracletestdataforof_writ964333684/002/program+0x80590)\n\n0x7be7d2be0040 is located 0 bytes after 48-byte region [0x7be7d2be0010,0x7be7d2be0040)\nallocated by thread T0 here:\n    #0 0x55eec8acd3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55eec8b385ef in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55eec8b385ef in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55eec8b10a8d in adamic_initialize_r1_reassigned_a_b0cabbd5f785d390af5fc51d0fdeada5 /adamic-build/main.c:117:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:92:195 in adamic_initialize_r1_reassigned_a_dcf141a7ab71bf906cd99ed3e9007ab7\nShadow bytes around the buggy address:\n  0x7be7d2bdfd80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be7d2bdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be7d2bdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be7d2bdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be7d2bdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n=>0x7be7d2be0000: fa fa 00 00 00 00 00 00[fa]fa fa fa fa fa fa fa\n  0x7be7d2be0080: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be7d2be0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be7d2be0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be7d2be0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be7d2be0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==9662==ABORTING\n"
        release:   exit 0, stdout "0\n1\n", stderr ""
    oracle_test.go:856: exit codes differ
        node:   exit 0, stdout "42\n1\n", stderr ""
        native: exit 1, stdout "", stderr "=================================================================\n==9662==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7be7d2be0040 at pc 0x55eec8b10f09 bp 0x7ffca98cb340 sp 0x7ffca98cb338\nREAD of size 8 at 0x7be7d2be0040 thread T0\n    #0 0x55eec8b10f08 in adamic_initialize_r1_reassigned_a_dcf141a7ab71bf906cd99ed3e9007ab7 /adamic-build/main.c:92:195\n    #1 0x55eec8b10a8d in adamic_initialize_r1_reassigned_a_b0cabbd5f785d390af5fc51d0fdeada5 /adamic-build/main.c:117:2\n    #2 0x55eec8b10a52 in main /adamic-build/main.c:127:2\n    #3 0x7fa7d3b43ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #4 0x7fa7d3b43d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x55eec8a2a590 in _start (/tmp/adamic-gate/TestNativeAgreesWithNode_071internaloracletestdataforof_writ964333684/002/program+0x80590)\n\n0x7be7d2be0040 is located 0 bytes after 48-byte region [0x7be7d2be0010,0x7be7d2be0040)\nallocated by thread T0 here:\n    #0 0x55eec8acd3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55eec8b385ef in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55eec8b385ef in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55eec8b10a8d in adamic_initialize_r1_reassigned_a_b0cabbd5f785d390af5fc51d0fdeada5 /adamic-build/main.c:117:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:92:195 in adamic_initialize_r1_reassigned_a_dcf141a7ab71bf906cd99ed3e9007ab7\nShadow bytes around the buggy address:\n  0x7be7d2bdfd80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be7d2bdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be7d2bdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be7d2bdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be7d2bdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n=>0x7be7d2be0000: fa fa 00 00 00 00 00 00[fa]fa fa fa fa fa fa fa\n  0x7be7d2be0080: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be7d2be0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be7d2be0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be7d2be0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be7d2be0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==9662==ABORTING\n"
    oracle_test.go:862: JavaScript backend: stdout differs
        node:    exit 0, stdout "42\n1\n", stderr ""
        backend: exit 0, stdout "undefined\n1\n", stderr ""
--- FAIL: TestNativeAgreesWithNode_071 (0.00s)
    --- FAIL: TestNativeAgreesWithNode_071/internal/oracle/testdata/forof_writes/r1_reassigned.a (44.96s)
FAIL
gate cache: native hits=0 misses=12
gate cache: node hits=0 misses=8
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	45.081s
FAIL
```

**lower: exit 1; wall 27.003 seconds.**

Every red test/subtest:

```text
TestLocalWritesAreLocalFieldsWrites
TestForOfReassignedBindingIsNoPlainSpreadSource
TestForOfReassignedSpreadSourceIsNotPlain
TestForOfReassignedBindingIsUnknown
```

Complete diagnostic text, including all emitted got/want and stderr (only scheduling messages and passing result lines omitted):

```text
    forof_writes_test.go:199: localFields says ForOf.Local writes a local, and flowLocalWrites doesn't say how collect takes it
--- FAIL: TestLocalWritesAreLocalFieldsWrites (0.00s)
    forof_writes_test.go:187: a for-of binding reassigned with a literal is taken as a plain spread source
--- FAIL: TestForOfReassignedBindingIsNoPlainSpreadSource (0.00s)
=== NAME  TestForOfReassignedSpreadSourceIsNotPlain
    forof_writes_test.go:168: got <nil>, want the library spread refusal
--- FAIL: TestForOfReassignedSpreadSourceIsNotPlain (0.17s)
=== NAME  TestForOfReassignedBindingIsUnknown
    forof_writes_test.go:96: reach is {Sites:[-2] Unknown:false Reasons:[]}, want Unknown for a for-of binding
--- FAIL: TestForOfReassignedBindingIsUnknown (0.19s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	2.216s
FAIL
```

### Mutant 1b

File: `internal/lower/view_flow.go`. Old:

```go
	"Binding.Local":  "a for-of binding takes each element",
```

New:
Delete the old line.

**oracle: exit 1; wall 65.409 seconds.**

Every red test/subtest:

```text
TestNativeAgreesWithNode_065
TestNativeAgreesWithNode_065/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts
TestNativeAgreesWithNode_023
TestNativeAgreesWithNode_023/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a
```

Complete diagnostic text, including all emitted got/want and stderr (only scheduling messages and passing result lines omitted):

```text
=== NAME  TestNativeAgreesWithNode_071/internal/oracle/testdata/forof_writes/r1_reassigned.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types a19c16807715 hit 26.61
=== NAME  TestNativeAgreesWithNode_023/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types a19c16807715 miss 26.61
=== NAME  TestNativeAgreesWithNode_065/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types a19c16807715 hit 26.61
=== NAME  TestNativeAgreesWithNode_081/internal/oracle/testdata/forof_writes/r1_reassigned.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types a19c16807715 hit 26.61
=== NAME  TestNativeAgreesWithNode_065/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts
    oracle_test.go:838: the release build: exit codes differ
        sanitized: exit 1, stdout "", stderr "=================================================================\n==10672==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7c231d5e0040 at pc 0x55fb3525011f bp 0x7ffe613652f0 sp 0x7ffe613652e8\nREAD of size 8 at 0x7c231d5e0040 thread T0\n    #0 0x55fb3525011e in adamic_initialize_r2_pattern_reassigned_ts_382e5c95379ccb2eed4aca9cbcf1a80b /adamic-build/main.c:103:205\n    #1 0x55fb3524fa8d in adamic_initialize_r2_pattern_reassigned_ts_435d65dc2bde7643cfb0a7cda1364039 /adamic-build/main.c:128:2\n    #2 0x55fb3524fa52 in main /adamic-build/main.c:139:2\n    #3 0x7fe31e5a6ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #4 0x7fe31e5a6d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x55fb35169590 in _start (/tmp/adamic-gate/TestNativeAgreesWithNode_065internaloracletestdataforof_writ68821666/002/program+0x80590)\n\n0x7c231d5e0040 is located 0 bytes after 48-byte region [0x7c231d5e0010,0x7c231d5e0040)\nallocated by thread T0 here:\n    #0 0x55fb3520c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55fb3527796f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55fb3527796f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55fb3524fa8d in adamic_initialize_r2_pattern_reassigned_ts_435d65dc2bde7643cfb0a7cda1364039 /adamic-build/main.c:128:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:103:205 in adamic_initialize_r2_pattern_reassigned_ts_382e5c95379ccb2eed4aca9cbcf1a80b\nShadow bytes around the buggy address:\n  0x7c231d5dfd80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c231d5dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c231d5dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c231d5dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c231d5dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n=>0x7c231d5e0000: fa fa 00 00 00 00 00 00[fa]fa fa fa fa fa fa fa\n  0x7c231d5e0080: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c231d5e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c231d5e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c231d5e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c231d5e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==10672==ABORTING\n"
        release:   exit 0, stdout "0\n1\n", stderr ""
    oracle_test.go:856: exit codes differ
        node:   exit 0, stdout "42\n1\n", stderr ""
        native: exit 1, stdout "", stderr "=================================================================\n==10672==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7c231d5e0040 at pc 0x55fb3525011f bp 0x7ffe613652f0 sp 0x7ffe613652e8\nREAD of size 8 at 0x7c231d5e0040 thread T0\n    #0 0x55fb3525011e in adamic_initialize_r2_pattern_reassigned_ts_382e5c95379ccb2eed4aca9cbcf1a80b /adamic-build/main.c:103:205\n    #1 0x55fb3524fa8d in adamic_initialize_r2_pattern_reassigned_ts_435d65dc2bde7643cfb0a7cda1364039 /adamic-build/main.c:128:2\n    #2 0x55fb3524fa52 in main /adamic-build/main.c:139:2\n    #3 0x7fe31e5a6ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #4 0x7fe31e5a6d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x55fb35169590 in _start (/tmp/adamic-gate/TestNativeAgreesWithNode_065internaloracletestdataforof_writ68821666/002/program+0x80590)\n\n0x7c231d5e0040 is located 0 bytes after 48-byte region [0x7c231d5e0010,0x7c231d5e0040)\nallocated by thread T0 here:\n    #0 0x55fb3520c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55fb3527796f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55fb3527796f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55fb3524fa8d in adamic_initialize_r2_pattern_reassigned_ts_435d65dc2bde7643cfb0a7cda1364039 /adamic-build/main.c:128:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:103:205 in adamic_initialize_r2_pattern_reassigned_ts_382e5c95379ccb2eed4aca9cbcf1a80b\nShadow bytes around the buggy address:\n  0x7c231d5dfd80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c231d5dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c231d5dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c231d5dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c231d5dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n=>0x7c231d5e0000: fa fa 00 00 00 00 00 00[fa]fa fa fa fa fa fa fa\n  0x7c231d5e0080: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c231d5e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c231d5e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c231d5e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c231d5e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==10672==ABORTING\n"
    oracle_test.go:862: JavaScript backend: stdout differs
        node:    exit 0, stdout "42\n1\n", stderr ""
        backend: exit 0, stdout "NaN\n1\n", stderr ""
--- FAIL: TestNativeAgreesWithNode_065 (0.00s)
    --- FAIL: TestNativeAgreesWithNode_065/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts (34.76s)
=== NAME  TestNativeAgreesWithNode_023/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a
    oracle_test.go:838: the release build: exit codes differ
        sanitized: exit 1, stdout "", stderr "=================================================================\n==10667==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b46b19e0040 at pc 0x559f7b76011f bp 0x7fff07a58b80 sp 0x7fff07a58b78\nREAD of size 8 at 0x7b46b19e0040 thread T0\n    #0 0x559f7b76011e in adamic_initialize_r2_pattern_reassigned_a_9caefb0ece1b19db94c79a631d3a4d65 /adamic-build/main.c:103:203\n    #1 0x559f7b75fa8d in adamic_initialize_r2_pattern_reassigned_a_aae2745952bad6112795923bcbf0935e /adamic-build/main.c:128:2\n    #2 0x559f7b75fa52 in main /adamic-build/main.c:139:2\n    #3 0x7f06b2911ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #4 0x7f06b2911d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x559f7b679590 in _start (/tmp/adamic-gate/TestNativeAgreesWithNode_023internaloracletestdataforof_writ124491925/002/program+0x80590)\n\n0x7b46b19e0040 is located 0 bytes after 48-byte region [0x7b46b19e0010,0x7b46b19e0040)\nallocated by thread T0 here:\n    #0 0x559f7b71c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x559f7b78796f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x559f7b78796f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x559f7b75fa8d in adamic_initialize_r2_pattern_reassigned_a_aae2745952bad6112795923bcbf0935e /adamic-build/main.c:128:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:103:203 in adamic_initialize_r2_pattern_reassigned_a_9caefb0ece1b19db94c79a631d3a4d65\nShadow bytes around the buggy address:\n  0x7b46b19dfd80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b46b19dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b46b19dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b46b19dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b46b19dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n=>0x7b46b19e0000: fa fa 00 00 00 00 00 00[fa]fa fa fa fa fa fa fa\n  0x7b46b19e0080: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b46b19e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b46b19e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b46b19e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b46b19e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==10667==ABORTING\n"
        release:   exit 0, stdout "0\n1\n", stderr ""
    oracle_test.go:856: exit codes differ
        node:   exit 0, stdout "42\n1\n", stderr ""
        native: exit 1, stdout "", stderr "=================================================================\n==10667==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b46b19e0040 at pc 0x559f7b76011f bp 0x7fff07a58b80 sp 0x7fff07a58b78\nREAD of size 8 at 0x7b46b19e0040 thread T0\n    #0 0x559f7b76011e in adamic_initialize_r2_pattern_reassigned_a_9caefb0ece1b19db94c79a631d3a4d65 /adamic-build/main.c:103:203\n    #1 0x559f7b75fa8d in adamic_initialize_r2_pattern_reassigned_a_aae2745952bad6112795923bcbf0935e /adamic-build/main.c:128:2\n    #2 0x559f7b75fa52 in main /adamic-build/main.c:139:2\n    #3 0x7f06b2911ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #4 0x7f06b2911d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x559f7b679590 in _start (/tmp/adamic-gate/TestNativeAgreesWithNode_023internaloracletestdataforof_writ124491925/002/program+0x80590)\n\n0x7b46b19e0040 is located 0 bytes after 48-byte region [0x7b46b19e0010,0x7b46b19e0040)\nallocated by thread T0 here:\n    #0 0x559f7b71c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x559f7b78796f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x559f7b78796f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x559f7b75fa8d in adamic_initialize_r2_pattern_reassigned_a_aae2745952bad6112795923bcbf0935e /adamic-build/main.c:128:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:103:203 in adamic_initialize_r2_pattern_reassigned_a_9caefb0ece1b19db94c79a631d3a4d65\nShadow bytes around the buggy address:\n  0x7b46b19dfd80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b46b19dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b46b19dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b46b19dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b46b19dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n=>0x7b46b19e0000: fa fa 00 00 00 00 00 00[fa]fa fa fa fa fa fa fa\n  0x7b46b19e0080: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b46b19e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b46b19e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b46b19e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b46b19e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==10667==ABORTING\n"
    oracle_test.go:862: JavaScript backend: stdout differs
        node:    exit 0, stdout "42\n1\n", stderr ""
        backend: exit 0, stdout "NaN\n1\n", stderr ""
--- FAIL: TestNativeAgreesWithNode_023 (0.00s)
    --- FAIL: TestNativeAgreesWithNode_023/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a (34.77s)
FAIL
gate cache: native hits=0 misses=12
gate cache: node hits=0 misses=8
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	34.907s
FAIL
```

**lower: exit 1; wall 36.277 seconds.**

Every red test/subtest:

```text
TestLocalWritesAreLocalFieldsWrites
TestForOfReassignedPatternBindingIsUnknown
```

Complete diagnostic text, including all emitted got/want and stderr (only scheduling messages and passing result lines omitted):

```text
    forof_writes_test.go:199: localFields says Binding.Local writes a local, and flowLocalWrites doesn't say how collect takes it
--- FAIL: TestLocalWritesAreLocalFieldsWrites (0.00s)
=== NAME  TestForOfReassignedPatternBindingIsUnknown
    forof_writes_test.go:113: reach is {Sites:[-3] Unknown:false Reasons:[]}, want Unknown for a for-of pattern's binding
--- FAIL: TestForOfReassignedPatternBindingIsUnknown (0.17s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	2.558s
FAIL
```

### Mutant 2

File: `internal/lower/view_flow.go`. Old:

```go
	"Try.CatchLocal": "a catch binding takes what was thrown",
```

New:
Delete the old line.

**oracle: exit 0; wall 89.568 seconds.**

No red test/subtest.


```text
--- PASS: TestNativeAgreesWithNode_071/internal/oracle/testdata/forof_writes/r1_reassigned.a (40.58s)
--- PASS: TestNativeAgreesWithNode_023/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a (40.63s)
--- PASS: TestNativeAgreesWithNode_081/internal/oracle/testdata/forof_writes/r1_reassigned.ts (40.64s)
--- PASS: TestNativeAgreesWithNode_065/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts (40.66s)
```

**lower: exit 1; wall 26.606 seconds.**

Every red test/subtest:

```text
TestLocalWritesAreLocalFieldsWrites
```

Complete diagnostic text, including all emitted got/want and stderr (only scheduling messages and passing result lines omitted):

```text
    forof_writes_test.go:199: localFields says Try.CatchLocal writes a local, and flowLocalWrites doesn't say how collect takes it
--- FAIL: TestLocalWritesAreLocalFieldsWrites (0.00s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	2.456s
FAIL
```

**catch: exit 1; wall 2.746 seconds.**

Every red test/subtest:

```text
TestCatchBindingAssignedIsUnknown
```

Complete diagnostic text, including all emitted got/want and stderr (only scheduling messages and passing result lines omitted):

```text
    forof_writes_test.go:148: reach is {Sites:[-1] Unknown:false Reasons:[]}, want Unknown for a catch binding
--- FAIL: TestCatchBindingAssignedIsUnknown (0.00s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.020s
FAIL
```

### Mutant 3

File: `internal/lower/library_spread.go`. Old:

```go
		if _, unknownSource := graph.unknownAt(node); active[node] || unknownSource || len(graph.sources[node]) == 0 {
```

New:

```go
		if _, unknownSource := graph.unknownAt(node); active[node] || unknownSource && graph.unknownParameters[node] || len(graph.sources[node]) == 0 {
```

**oracle: exit 0; wall 71.492 seconds.**

No red test/subtest.


```text
--- PASS: TestNativeAgreesWithNode_081/internal/oracle/testdata/forof_writes/r1_reassigned.ts (38.28s)
--- PASS: TestNativeAgreesWithNode_065/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts (38.35s)
--- PASS: TestNativeAgreesWithNode_071/internal/oracle/testdata/forof_writes/r1_reassigned.a (38.39s)
--- PASS: TestNativeAgreesWithNode_023/internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a (38.40s)
```

**lower: exit 1; wall 32.371 seconds.**

Every red test/subtest:

```text
TestForOfReassignedBindingIsNoPlainSpreadSource
TestForOfReassignedSpreadSourceIsNotPlain
```

Complete diagnostic text, including all emitted got/want and stderr (only scheduling messages and passing result lines omitted):

```text
    forof_writes_test.go:187: a for-of binding reassigned with a literal is taken as a plain spread source
--- FAIL: TestForOfReassignedBindingIsNoPlainSpreadSource (0.00s)
=== NAME  TestForOfReassignedSpreadSourceIsNotPlain
    forof_writes_test.go:168: got <nil>, want the library spread refusal
--- FAIL: TestForOfReassignedSpreadSourceIsNotPlain (0.25s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	2.356s
FAIL
```

The table omissions expect the named local-writing row to remain present. The plain-spread assertion expects the binding not to be a plain source. These assertions print prose rather than explicit got/want labels; their whole diagnostics are retained above. Sanitizer addresses, PIDs and stack texts are preserved exactly as emitted, without normalization.

## 4. Full lower suite


```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Exit 0; wall 1845.816 seconds (30m 45.816s). Failing tests: **none**; environment failures: **none**.

```text
ok  	github.com/system-inc/adamic/internal/lower	1841.231s
```

## 5. Full oracle suite


```sh
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```

Exit 1; wall 4099.915 seconds (68m 19.915s). **14 failing top-level tests.** Every failing test and its first error line follows.

| Failing test | First error line | Classification |
| --- | --- | --- |
| `TestAssignmentProofWiderTargetMutant` | `assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofWiderTarget` | `assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofCompoundMutant` | `assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofCompound` | `assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofChainedMutant` | `assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofChained` | `assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofReturnDefinedMutant` | `assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofReturnDefined` | `assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofIfNarrowingMutant` | `assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofIfNarrowing` | `assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofScannerKeywordMutant` | `assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestAssignmentProofScannerKeyword` | `assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | ENV: hardcoded NODE_PATH cannot resolve external TypeScript; targeted retry passes |
| `TestStdoutColumnsRefusesWASI` | `library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT` | ENV: WASI_SYSROOT absent; remains unresolved |
| `TestFamiliesCatchAPlantedCase_CountsAreRecorded` | `families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s:` | RESOURCE/TIMEOUT: 90 s nested-probe limit; isolated retry passes |

The twelve assignment-proof errors continue with `Error: Cannot find module 'typescript'` / `MODULE_NOT_FOUND`. Their helper explicitly overwrites NODE_PATH with `<repository>/stage3/api/node_modules`. The initial external npm install put tsc on PATH but does not satisfy this hardcoded module-resolution path.

Supplemental checks, run serially after the full suite: temporarily linked `stage3/api/node_modules` to `/tmp/adamic-proof-zxx/api/node_modules`, ran the assignment tests and counts probe alone, then removed the symlink via EXIT trap. Packages remained installed outside the repository. No compiler changes were made.


```sh
go test ./internal/oracle -run '^TestAssignmentProof' -count=1 -v -timeout 20m
```

Exit 0; wall 40.173 seconds.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofCompoundMutant
=== CONT  TestAssignmentProofReturnDefinedMutant
=== CONT  TestAssignmentProofWiderTargetMutant
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.90
=== NAME  TestAssignmentProofChained
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.94
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.95
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 1.00
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 1.02
=== NAME  TestAssignmentProofChained
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:172: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 1\nvalue2 value2 2\n", mutant "value1 undefined 1\nvalue2 undefined 2\n"
--- PASS: TestAssignmentProofReturnDefinedMutant (12.98s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofChained
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofChained (12.99s)
=== CONT  TestAssignmentProofIfNarrowingMutant
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:202: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128 1\n", mutant "128 0 1\n"
--- PASS: TestAssignmentProofWiderTargetMutant (13.01s)
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofScannerKeyword (13.09s)
=== CONT  TestAssignmentProofCompound
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:192: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "12 12 1\n14 14 2\n", mutant "2 2 1\n2 2 2\n"
--- PASS: TestAssignmentProofCompoundMutant (13.19s)
=== CONT  TestAssignmentProofIfNarrowing
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.02
=== NAME  TestAssignmentProofCompound
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.01
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.05
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.02
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.21
=== NAME  TestAssignmentProofCompound
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofCompound (10.40s)
=== CONT  TestAssignmentProofChainedMutant
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:162: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "4 1\n-1 2\n", mutant "4 1\n7 2\n"
--- PASS: TestAssignmentProofIfNarrowingMutant (11.02s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofWiderTarget (11.62s)
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofReturnDefined (12.94s)
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofIfNarrowing (14.06s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.01
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 8ad6a79e2a5b hit 0.01
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:152: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128\n83 83\n80 80\n80 80\n80 80\n", mutant "80 80\n80 80\n80 80\n80 80\n80 80\n"
--- PASS: TestAssignmentProofScannerKeywordMutant (5.35s)
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:182: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 value1 1\n", mutant "value1 value1 undefined 1\n"
--- PASS: TestAssignmentProofChainedMutant (5.93s)
PASS
gate cache: native hits=12 misses=24
gate cache: node hits=0 misses=24
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	29.448s
```


```sh
go test ./internal/oracle -run '^TestFamiliesCatchAPlantedCase_CountsAreRecorded$' -count=1 -v -timeout 20m
```

Exit 0; wall 28.505 seconds.

```text
=== RUN   TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== PAUSE TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== CONT  TestFamiliesCatchAPlantedCase_CountsAreRecorded
    families_generated_test.go:5190: the mutant planted in dedication/dedication.a failed exactly TestCountsAreRecorded_003, with "the counts changed"
--- PASS: TestFamiliesCatchAPlantedCase_CountsAreRecorded (24.51s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	24.536s
```

The counts failure is attributed to contention/resource timing because the same planted-case probe passed alone, with the intended counts disagreement. The original full-suite failure is retained. No WASI sysroot was provisioned; TestStdoutColumnsRefusesWASI remains an environment failure. The full oracle suite was not repeated.

## 6. Complete admission delta


```sh
/tmp/adamic-proof-zxx/admission-delta --base 16429a30452d027809f0ea85bc395479423e493e --head 73211726b0674668dc2335d81af630a244127e69 --manifest /tmp/adamic-proof-zxx/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json
```

Exit 0; measured wall 1381.264 seconds. Gate phase total: 1380.479 seconds.


```json
{
  "base": "16429a30452d027809f0ea85bc395479423e493e",
  "head": "73211726b0674668dc2335d81af630a244127e69",
  "verdict": "pass",
  "void": false,
  "void_reason": null,
  "workers": 2,
  "classification": "lowering-only",
  "complete_corpus": true,
  "diff_count": 4,
  "admitted": 0,
  "omitted": 0,
  "timed_out_runs": 0,
  "retried_runs": 0,
  "still_timed_out_runs": 0,
  "generator_revision": "73211726b0674668dc2335d81af630a244127e69",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "535524f8d969e1625ebf62c789278c5e9f8783a3",
  "phase_seconds": {
    "base_build": 57.385143841,
    "checkout": 16.211962227,
    "checkout_and_provision": 17.314153818,
    "classification": 781.647453506,
    "head_build": 28.219463515,
    "manifest_and_verification": 22.443306946,
    "provision": 1.102156547,
    "runtime": 473.460620185,
    "total": 1380.478540647
  },
  "program_count": 5218,
  "classes": {
    "accepted-by-both": 3274,
    "refused-by-both": 1944
  },
  "newly_refused": 0,
  "classification_stderr_differences": 0,
  "classification_timeouts": 0
}
```

Missing void_reason in gate JSON means no void reason; rendered above as null.

All programs whose classification stderr differs, with both texts: **none**. Comparison used the exact `base_compile.stderr` and `compile.stderr` strings for all 5,218 entries; no path, PID, whitespace or text normalization. No newly refused paths exist.

Diff corpus:

```json
[
  {
    "path": "internal/oracle/testdata/forof_writes/r1_reassigned.a",
    "blob": "c8c716ad8de7e52c936a4d5d01e29c7eab7818c3"
  },
  {
    "path": "internal/oracle/testdata/forof_writes/r1_reassigned.ts",
    "blob": "c8c716ad8de7e52c936a4d5d01e29c7eab7818c3"
  },
  {
    "path": "internal/oracle/testdata/forof_writes/r2_pattern_reassigned.a",
    "blob": "c4c82c401a709f98e0badcf2e44057594e154255"
  },
  {
    "path": "internal/oracle/testdata/forof_writes/r2_pattern_reassigned.ts",
    "blob": "c4c82c401a709f98e0badcf2e44057594e154255"
  }
]
```

Corpus entry counts: witnesses 2; fixtures 4,270; gaps 105; review 20; fuzz 0; stage3 821. Classification totals: accepted by both 3,274; refused by both 1,944. The gate omitted zero programs. All four new fixture programs were accepted by both revisions; the native pins and mutants above provide the semantic regression evidence.

Gate stderr:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 4
```

No classification observation has an error or void reason, and no runtime timed_out entry exists. With zero timeouts, there is no program to rerun alone at either SHA.

## Evidence integrity

Raw evidence remains in `/tmp/adamic-proof-zxx`, plus the requested `/tmp/lower.log` and `/tmp/oracle.log`. These are local paths, not committed attachments. The report contains all requested baseline subtest statuses, every mutant failure and full emitted got/want diagnostics, all full-suite failure first lines, and the complete list (empty) of differing classification stderr. SHA-256 hashes of the raw command outputs:

| Output | SHA-256 |
| --- | --- |
| `/tmp/adamic-proof-zxx/gofmt.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/adamic-proof-zxx/vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/adamic-proof-zxx/pins-oracle.log` | `3cd2d7e6eccc04eca5e599dae9d0060a5dbe3821d5c6b2f331555ffc63835b51` |
| `/tmp/adamic-proof-zxx/pins-lower.log` | `31b2011e4449687b339c3cdc86cb33cd14493801df417db2a9760f0776b0948c` |
| `/tmp/lower.log` | `88e456140dd8171e92bddf7ddc271e4185c17f00557ee7e36238f13f4df0816d` |
| `/tmp/oracle.log` | `020577fe989e99cd9d26c1e1cd7fc2a2e88c3ba5f2e47c9e95d73bc6fcce9b86` |
| `/tmp/adamic-proof-zxx/oracle-env-retry.log` | `3abb3a11bd22d2caaec3ace1f34d63ba035bec312691639a44230d04c4e19070` |
| `/tmp/adamic-proof-zxx/oracle-counts-retry.log` | `361af6127b4b11a57ffd0a471c396403f47f2a0c21811b370a959779c9247249` |
| `/tmp/adamic-proof-zxx/admission.json` | `36bef5b299ea6f0bddd7c43fb188b1e123af4d966c4a93b4613d62b02e77cb9a` |
| `/tmp/adamic-proof-zxx/admission.stderr` | `ffb1f830f236b8294b0fcdc234d58bf3aaca60ffaa0e832f6ed99248153300c0` |
| `/tmp/adamic-proof-zxx/manifest.json` | `01531df448352ca241d38ad41ec49159e80741bf7f295ff978bd6d3d85eeffcc` |
| `/tmp/adamic-proof-zxx/mutant-1a-lower.log` | `023008c54d92117e383199079b9ccd7f144b2bd3161fd291eed908e1d2022fcb` |
| `/tmp/adamic-proof-zxx/mutant-1a-oracle.log` | `a5e126d3947038f47ae1e83c01affbcbc54d2dab26579852c098e1557f3c3ce4` |
| `/tmp/adamic-proof-zxx/mutant-1b-lower.log` | `fb55f998601e4c65e760bc5bf0d0e5844b5daee7e162b0e5a21e4ab7940f2594` |
| `/tmp/adamic-proof-zxx/mutant-1b-oracle.log` | `bfe151c6dc06d55a5a4a81353176dfb8d131cacb6ba1f1ea92512b4bc7854f5f` |
| `/tmp/adamic-proof-zxx/mutant-2-catch.log` | `4f0f636c5e37fdc5a4c44b363341a41b3662a9d542f58ec4d192fac3ff045483` |
| `/tmp/adamic-proof-zxx/mutant-2-lower.log` | `4e3c39385754da3a29d22d31c11a17af84112445b0dd8aba3732c07662d67539` |
| `/tmp/adamic-proof-zxx/mutant-2-oracle.log` | `ecee948999b43fbf332dd5c11adab714bf96dfdb419a88ea43e806da80f0feb5` |
| `/tmp/adamic-proof-zxx/mutant-3-lower.log` | `8846da4aa43916f48b2a95f591b94abde2f50fa0bed68274c51e476031d284d7` |
| `/tmp/adamic-proof-zxx/mutant-3-oracle.log` | `b3cf3a7558e9159142f327e787130ffe1201d33423bb2002eaa3ead77acf5a59` |
