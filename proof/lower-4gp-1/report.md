# Proof for #4gpbn7w

Final report: steps 1–3, 4, 4b and 6 complete. Initial steps 1–3 report was committed as `0a9a63e10` and pushed before starting step 4. Only this report is committed on `proof/lower-4gp-1-report`, based on the requested HEAD.

HEAD: `cc531f31875e526cade6f32a1d7e2ed49898cde2`. Base: `2f32716054f245622c2ac3c8f692d00fa51e66e9`.

## Decisive results

HEAD passes all four pins. Base and the documented mutant fail all four sanitized pins with AddressSanitizer heap-buffer-overflow, READ of size 8 immediately after a 48-byte allocation. Native release and JavaScript have no disagreements. The full lower suite fails one product/source consistency test; the full IR suite and complete admission delta pass.

## Setup and step 1

Loaded `/workspace/adamic-tools/env.sh` first. Created its missing TMPDIR `/tmp/adamic-gate` outside the repository. Fetched `origin compiler/accessor-dispatch-absent-1`, checked out HEAD detached, and completed `git submodule update --init --recursive`. `gofmt -l internal/lower internal/oracle` printed nothing; `go vet ./internal/lower ./internal/oracle` exited 0.

Upstream TypeScript is separately checked out at `/tmp/adamic-proof-4gp/typescript`, exact commit `050880ce59e30b356b686bd3144efe24f875ebc8`, exported as `ADAMIC_TYPESCRIPT_SOURCE`. Copied stage3/api package.json and lockfile outside the repository and ran `npm ci --ignore-scripts`; external tsc reports 6.0.3 and is on PATH. All test/gate invocations inherit `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, `npm_config_strict_ssl=true`. Heavy commands ran sequentially. `/usr/bin/time` was absent; wall seconds use the shell clock.

## Step 2: pins

Command at each revision:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestAccessorDispatchAbsentAgreesWithNode' -count=1 -v -timeout 40m
```

At base, HEAD’s test file and testdata were copied in as untracked files, never committed, then removed before returning to HEAD.

| Subtest | HEAD | Base sanitized | Mutant sanitized |
|---|---|---|---|
| presence.a | PASS | FAIL | FAIL |
| presence.ts | PASS | FAIL | FAIL |
| undefined_compare.a | PASS | FAIL | FAIL |
| undefined_compare.ts | PASS | FAIL | FAIL |

head: `exit=0 wall_seconds=172`.

base: `exit=1 wall_seconds=42`.

mutant: `exit=1 wall_seconds=39`.

## Step 3: mutant

The sole row in `internal/lower/accessor-dispatch-absent.mutants.md` replaces `fallback.Object = object` with `fallback.Object, fallback.Absent = object, false` in `internal/lower/class_accessors.go`. Asserted exactly one occurrence before applying. All four sanitized subtests red. Restored source and diagnostic helper; `git diff --exit-code` exited 0.

## Complete failure observations

The exact requested runs log stderr but omit exit/stdout values. Supplemental runs used the identical command with temporary logging of complete got/want records in `loomViewsDAgreesBackend`; assertions and execution were unchanged. The logging was restored afterward. Below are all supplemental failure observations, including full ASan stderr (quoted Go strings preserve every byte).

### base

```text
internal/oracle/testdata/accessor_dispatch_absent/presence.a sanitized got exit=1 stdout="" stderr="=================================================================\n==9833==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7bc454be0080 at pc 0x55802f016daf bp 0x7ffccf1f16e0 sp 0x7ffccf1f16d8\nREAD of size 8 at 0x7bc454be0080 thread T0\n    #0 0x55802f016dae in adamic_function_property_get_m_44616eaa941cde68afaeb5a452c9e3f0 /adamic-build/main.c:125:214\n    #1 0x55802f016dae in adamic_function_presence_a_probe_683c300499b7b9215cf076fee0f8e6ad /adamic-build/main.c:45:79\n    #2 0x55802f016b4e in adamic_initialize_presence_a_1a79fe28c2235bf6795af6d7910f7e70 /adamic-build/main.c:139:90\n    #3 0x55802f016a6d in adamic_initialize_presence_a_c326d3af3aafdd9427865fd67712f5f7 /adamic-build/main.c:155:2\n    #4 0x55802f016a52 in main /adamic-build/main.c:165:2\n    #5 0x7f84557daca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f84557dad64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x55802ef30590 in _start (/tmp/adamic-gate/TestAccessorDispatchAbsentAgreesWithNodepresence.a2269893916/002/program+0x81590)\n\n0x7bc454be0080 is located 0 bytes after 48-byte region [0x7bc454be0050,0x7bc454be0080)\nallocated by thread T0 here:\n    #0 0x55802efd33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55802f03f11f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55802f03f11f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55802f016a6d in adamic_initialize_presence_a_c326d3af3aafdd9427865fd67712f5f7 /adamic-build/main.c:155:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:125:214 in adamic_function_property_get_m_44616eaa941cde68afaeb5a452c9e3f0\nShadow bytes around the buggy address:\n  0x7bc454bdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bc454bdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bc454bdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bc454bdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bc454be0000: fa fa fd fd fd fd fd fd fa fa 00 00 00 00 00 00\n=>0x7bc454be0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bc454be0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bc454be0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bc454be0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bc454be0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bc454be0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==9833==ABORTING\n" want exit=0 stdout="has\nnone\nhas\n" stderr=""
```

```text
internal/oracle/testdata/accessor_dispatch_absent/undefined_compare.a sanitized got exit=1 stdout="" stderr="=================================================================\n==9842==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b4dc6be0080 at pc 0x55ef439c60ec bp 0x7ffceeffcf00 sp 0x7ffceeffcef8\nREAD of size 8 at 0x7b4dc6be0080 thread T0\n    #0 0x55ef439c60eb in adamic_function_property_get_m_577812b7b7c3dbd368c0d9fd58c3f1c3 /adamic-build/main.c:137:214\n    #1 0x55ef439c60eb in adamic_function_undefined_compare_a_value_b095eeb2d4c7c6c6065421f758b9daed /adamic-build/main.c:45:79\n    #2 0x55ef439c5bc8 in adamic_initialize_undefined_compare_a_27fee3aa71e72e30062bafa0b963388a /adamic-build/main.c:160:94\n    #3 0x55ef439c5a6d in adamic_initialize_undefined_compare_a_0c29b8b267bd9670af5dc93b13b75d74 /adamic-build/main.c:194:2\n    #4 0x55ef439c5a52 in main /adamic-build/main.c:204:2\n    #5 0x7f0dc7b87ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f0dc7b87d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x55ef438df590 in _start (/tmp/adamic-gate/TestAccessorDispatchAbsentAgreesWithNodeundefined_compare.a113265893/002/program+0x82590)\n\n0x7b4dc6be0080 is located 0 bytes after 48-byte region [0x7b4dc6be0050,0x7b4dc6be0080)\nallocated by thread T0 here:\n    #0 0x55ef439823a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55ef439ee4cf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55ef439ee4cf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55ef439c5a6d in adamic_initialize_undefined_compare_a_0c29b8b267bd9670af5dc93b13b75d74 /adamic-build/main.c:194:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:137:214 in adamic_function_property_get_m_577812b7b7c3dbd368c0d9fd58c3f1c3\nShadow bytes around the buggy address:\n  0x7b4dc6bdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b4dc6bdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b4dc6bdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b4dc6bdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b4dc6be0000: fa fa 00 00 00 00 00 00 fa fa 00 00 00 00 00 00\n=>0x7b4dc6be0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4dc6be0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4dc6be0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4dc6be0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4dc6be0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4dc6be0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==9842==ABORTING\n" want exit=0 stdout="5550\n" stderr=""
```

```text
internal/oracle/testdata/accessor_dispatch_absent/presence.ts sanitized got exit=1 stdout="" stderr="=================================================================\n==9847==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b9470be0080 at pc 0x563f120d4daf bp 0x7fffc92177c0 sp 0x7fffc92177b8\nREAD of size 8 at 0x7b9470be0080 thread T0\n    #0 0x563f120d4dae in adamic_function_property_get_m_1c12238205d371ba902a20fd90a9f40e /adamic-build/main.c:125:214\n    #1 0x563f120d4dae in adamic_function_presence_ts_probe_2e12dd534000be433c747c336a6b7270 /adamic-build/main.c:45:79\n    #2 0x563f120d4b4e in adamic_initialize_presence_ts_5f15e72351266b926ca00eb5927a9e42 /adamic-build/main.c:139:91\n    #3 0x563f120d4a6d in adamic_initialize_presence_ts_f54b48d2cf47b3ace7df7799d1627f58 /adamic-build/main.c:155:2\n    #4 0x563f120d4a52 in main /adamic-build/main.c:165:2\n    #5 0x7f5471935ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f5471935d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x563f11fee590 in _start (/tmp/adamic-gate/TestAccessorDispatchAbsentAgreesWithNodepresence.ts4118463642/002/program+0x81590)\n\n0x7b9470be0080 is located 0 bytes after 48-byte region [0x7b9470be0050,0x7b9470be0080)\nallocated by thread T0 here:\n    #0 0x563f120913a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x563f120fd11f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x563f120fd11f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x563f120d4a6d in adamic_initialize_presence_ts_f54b48d2cf47b3ace7df7799d1627f58 /adamic-build/main.c:155:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:125:214 in adamic_function_property_get_m_1c12238205d371ba902a20fd90a9f40e\nShadow bytes around the buggy address:\n  0x7b9470bdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9470bdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9470bdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9470bdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9470be0000: fa fa fd fd fd fd fd fd fa fa 00 00 00 00 00 00\n=>0x7b9470be0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9470be0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9470be0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9470be0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9470be0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9470be0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==9847==ABORTING\n" want exit=0 stdout="has\nnone\nhas\n" stderr=""
```

```text
internal/oracle/testdata/accessor_dispatch_absent/undefined_compare.ts sanitized got exit=1 stdout="" stderr="=================================================================\n==9850==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7bcd28fe0080 at pc 0x55e28f3e40ec bp 0x7ffd1fc64b30 sp 0x7ffd1fc64b28\nREAD of size 8 at 0x7bcd28fe0080 thread T0\n    #0 0x55e28f3e40eb in adamic_function_property_get_m_1628164efe41e8c5c544e1f16a0eb7d3 /adamic-build/main.c:137:214\n    #1 0x55e28f3e40eb in adamic_function_undefined_compare_ts_value_42834c9f3e37389f1ac8821c5f76bead /adamic-build/main.c:45:79\n    #2 0x55e28f3e3bc8 in adamic_initialize_undefined_compare_ts_7557cfe860544855e6a9a9e6debb0219 /adamic-build/main.c:160:95\n    #3 0x55e28f3e3a6d in adamic_initialize_undefined_compare_ts_f02d3d6ed7057c4d8d72f313b62471a2 /adamic-build/main.c:194:2\n    #4 0x55e28f3e3a52 in main /adamic-build/main.c:204:2\n    #5 0x7f8d29e9dca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f8d29e9dd64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x55e28f2fd590 in _start (/tmp/adamic-gate/TestAccessorDispatchAbsentAgreesWithNodeundefined_compare.ts2272945788/002/program+0x82590)\n\n0x7bcd28fe0080 is located 0 bytes after 48-byte region [0x7bcd28fe0050,0x7bcd28fe0080)\nallocated by thread T0 here:\n    #0 0x55e28f3a03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55e28f40c4cf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55e28f40c4cf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55e28f3e3a6d in adamic_initialize_undefined_compare_ts_f02d3d6ed7057c4d8d72f313b62471a2 /adamic-build/main.c:194:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:137:214 in adamic_function_property_get_m_1628164efe41e8c5c544e1f16a0eb7d3\nShadow bytes around the buggy address:\n  0x7bcd28fdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bcd28fdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bcd28fdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bcd28fdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bcd28fe0000: fa fa 00 00 00 00 00 00 fa fa 00 00 00 00 00 00\n=>0x7bcd28fe0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bcd28fe0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bcd28fe0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bcd28fe0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bcd28fe0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bcd28fe0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==9850==ABORTING\n" want exit=0 stdout="5550\n" stderr=""
```

### mutant

```text
internal/oracle/testdata/accessor_dispatch_absent/presence.ts sanitized got exit=1 stdout="" stderr="=================================================================\n==11406==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7bd013be0080 at pc 0x55abcb8acdaf bp 0x7fff386a3120 sp 0x7fff386a3118\nREAD of size 8 at 0x7bd013be0080 thread T0\n    #0 0x55abcb8acdae in adamic_function_property_get_m_b5c7fd8aaea146ee758414c94a1d6119 /adamic-build/main.c:125:213\n    #1 0x55abcb8acdae in adamic_function_presence_ts_probe_2e12dd534000be433c747c336a6b7270 /adamic-build/main.c:45:79\n    #2 0x55abcb8acb4e in adamic_initialize_presence_ts_5f15e72351266b926ca00eb5927a9e42 /adamic-build/main.c:139:91\n    #3 0x55abcb8aca6d in adamic_initialize_presence_ts_f54b48d2cf47b3ace7df7799d1627f58 /adamic-build/main.c:155:2\n    #4 0x55abcb8aca52 in main /adamic-build/main.c:165:2\n    #5 0x7f9014832ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f9014832d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x55abcb7c6590 in _start (/tmp/adamic-gate/TestAccessorDispatchAbsentAgreesWithNodepresence.ts4034800783/002/program+0x81590)\n\n0x7bd013be0080 is located 0 bytes after 48-byte region [0x7bd013be0050,0x7bd013be0080)\nallocated by thread T0 here:\n    #0 0x55abcb8693a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55abcb8d511f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55abcb8d511f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55abcb8aca6d in adamic_initialize_presence_ts_f54b48d2cf47b3ace7df7799d1627f58 /adamic-build/main.c:155:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:125:213 in adamic_function_property_get_m_b5c7fd8aaea146ee758414c94a1d6119\nShadow bytes around the buggy address:\n  0x7bd013bdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bd013bdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bd013bdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bd013bdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bd013be0000: fa fa fd fd fd fd fd fd fa fa 00 00 00 00 00 00\n=>0x7bd013be0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bd013be0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bd013be0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bd013be0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bd013be0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bd013be0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==11406==ABORTING\n" want exit=0 stdout="has\nnone\nhas\n" stderr=""
```

```text
internal/oracle/testdata/accessor_dispatch_absent/presence.a sanitized got exit=1 stdout="" stderr="=================================================================\n==11408==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b4afa5e0080 at pc 0x55bb4324cdaf bp 0x7ffc9cd9cfd0 sp 0x7ffc9cd9cfc8\nREAD of size 8 at 0x7b4afa5e0080 thread T0\n    #0 0x55bb4324cdae in adamic_function_property_get_m_d56e9fd6799340cdc74b4f86d15a0ac0 /adamic-build/main.c:125:213\n    #1 0x55bb4324cdae in adamic_function_presence_a_probe_683c300499b7b9215cf076fee0f8e6ad /adamic-build/main.c:45:79\n    #2 0x55bb4324cb4e in adamic_initialize_presence_a_1a79fe28c2235bf6795af6d7910f7e70 /adamic-build/main.c:139:90\n    #3 0x55bb4324ca6d in adamic_initialize_presence_a_c326d3af3aafdd9427865fd67712f5f7 /adamic-build/main.c:155:2\n    #4 0x55bb4324ca52 in main /adamic-build/main.c:165:2\n    #5 0x7f0afb349ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f0afb349d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x55bb43166590 in _start (/tmp/adamic-gate/TestAccessorDispatchAbsentAgreesWithNodepresence.a2728167497/002/program+0x81590)\n\n0x7b4afa5e0080 is located 0 bytes after 48-byte region [0x7b4afa5e0050,0x7b4afa5e0080)\nallocated by thread T0 here:\n    #0 0x55bb432093a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55bb4327511f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55bb4327511f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55bb4324ca6d in adamic_initialize_presence_a_c326d3af3aafdd9427865fd67712f5f7 /adamic-build/main.c:155:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:125:213 in adamic_function_property_get_m_d56e9fd6799340cdc74b4f86d15a0ac0\nShadow bytes around the buggy address:\n  0x7b4afa5dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b4afa5dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b4afa5dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b4afa5dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b4afa5e0000: fa fa fd fd fd fd fd fd fa fa 00 00 00 00 00 00\n=>0x7b4afa5e0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4afa5e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4afa5e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4afa5e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4afa5e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b4afa5e0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==11408==ABORTING\n" want exit=0 stdout="has\nnone\nhas\n" stderr=""
```

```text
internal/oracle/testdata/accessor_dispatch_absent/undefined_compare.a sanitized got exit=1 stdout="" stderr="=================================================================\n==11407==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b8999fe0080 at pc 0x55993c54f0ec bp 0x7ffd60ef4390 sp 0x7ffd60ef4388\nREAD of size 8 at 0x7b8999fe0080 thread T0\n    #0 0x55993c54f0eb in adamic_function_property_get_m_3e3d69c65abdb5017ffb80ed85afce1d /adamic-build/main.c:137:214\n    #1 0x55993c54f0eb in adamic_function_undefined_compare_a_value_b095eeb2d4c7c6c6065421f758b9daed /adamic-build/main.c:45:79\n    #2 0x55993c54ebc8 in adamic_initialize_undefined_compare_a_27fee3aa71e72e30062bafa0b963388a /adamic-build/main.c:160:94\n    #3 0x55993c54ea6d in adamic_initialize_undefined_compare_a_0c29b8b267bd9670af5dc93b13b75d74 /adamic-build/main.c:194:2\n    #4 0x55993c54ea52 in main /adamic-build/main.c:204:2\n    #5 0x7f499adbfca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f499adbfd64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x55993c468590 in _start (/tmp/adamic-gate/TestAccessorDispatchAbsentAgreesWithNodeundefined_compare.a763022461/002/program+0x82590)\n\n0x7b8999fe0080 is located 0 bytes after 48-byte region [0x7b8999fe0050,0x7b8999fe0080)\nallocated by thread T0 here:\n    #0 0x55993c50b3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55993c5774cf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55993c5774cf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55993c54ea6d in adamic_initialize_undefined_compare_a_0c29b8b267bd9670af5dc93b13b75d74 /adamic-build/main.c:194:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:137:214 in adamic_function_property_get_m_3e3d69c65abdb5017ffb80ed85afce1d\nShadow bytes around the buggy address:\n  0x7b8999fdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b8999fdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b8999fdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b8999fdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b8999fe0000: fa fa 00 00 00 00 00 00 fa fa 00 00 00 00 00 00\n=>0x7b8999fe0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b8999fe0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b8999fe0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b8999fe0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b8999fe0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b8999fe0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==11407==ABORTING\n" want exit=0 stdout="5550\n" stderr=""
```

```text
internal/oracle/testdata/accessor_dispatch_absent/undefined_compare.ts sanitized got exit=1 stdout="" stderr="=================================================================\n==11422==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b6c5afe0080 at pc 0x5645c07400ec bp 0x7fffde53dd50 sp 0x7fffde53dd48\nREAD of size 8 at 0x7b6c5afe0080 thread T0\n    #0 0x5645c07400eb in adamic_function_property_get_m_a4af98e966cd9f61b1187919fd9a18fe /adamic-build/main.c:137:214\n    #1 0x5645c07400eb in adamic_function_undefined_compare_ts_value_42834c9f3e37389f1ac8821c5f76bead /adamic-build/main.c:45:79\n    #2 0x5645c073fbc8 in adamic_initialize_undefined_compare_ts_7557cfe860544855e6a9a9e6debb0219 /adamic-build/main.c:160:95\n    #3 0x5645c073fa6d in adamic_initialize_undefined_compare_ts_f02d3d6ed7057c4d8d72f313b62471a2 /adamic-build/main.c:194:2\n    #4 0x5645c073fa52 in main /adamic-build/main.c:204:2\n    #5 0x7f2c5bdc4ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f2c5bdc4d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x5645c0659590 in _start (/tmp/adamic-gate/TestAccessorDispatchAbsentAgreesWithNodeundefined_compare.ts4179702015/002/program+0x82590)\n\n0x7b6c5afe0080 is located 0 bytes after 48-byte region [0x7b6c5afe0050,0x7b6c5afe0080)\nallocated by thread T0 here:\n    #0 0x5645c06fc3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x5645c07684cf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x5645c07684cf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x5645c073fa6d in adamic_initialize_undefined_compare_ts_f02d3d6ed7057c4d8d72f313b62471a2 /adamic-build/main.c:194:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:137:214 in adamic_function_property_get_m_a4af98e966cd9f61b1187919fd9a18fe\nShadow bytes around the buggy address:\n  0x7b6c5afdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b6c5afdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b6c5afdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b6c5afdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b6c5afe0000: fa fa 00 00 00 00 00 00 fa fa 00 00 00 00 00 00\n=>0x7b6c5afe0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b6c5afe0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b6c5afe0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b6c5afe0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b6c5afe0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b6c5afe0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==11422==ABORTING\n" want exit=0 stdout="5550\n" stderr=""
```

## Step 4: full lower suite — FAIL

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

exit=1 wall_seconds=916

One failing test, `TestNativeAgreementLeanBuildIsBuildsBytes`. First error line:

```text
agree_test.go:143: ObjectDestructuringAssignmentGetter: the lower agreement C holds 26415 bytes of C, and this test lowers 26427 other bytes from its source
```

Environment classification: **not established as environmental**. This is a held-product/source consistency failure, not a missing tool, missing dependency, network error or timeout. The held lower agreement C product was a cache hit (`c0c5637d25f6`). The mismatch is reported as observed; no compiler or product changes were made. No other failing test appeared in `/tmp/lower.log`. Package test time 900.655 seconds; total shell wall time 916 seconds.

## Step 4b: full IR suite — PASS

```sh
go test ./internal/ir -count=1 -timeout 30m
```

exit=0 wall_seconds=78

No failures. This full suite includes `TestCallTargetReadersLower` and the IR field guards. Package time 61.114 seconds.

## Step 6: admission delta — PASS

Ran alone after lower and IR had exited, with no competing heavy job:

```sh
python3 cloud/admission-corpus/manifest.py --sha cc531f31875e526cade6f32a1d7e2ed49898cde2 > /tmp/adamic-proof-4gp/manifest.json
go run ./cmd/adamic-admission-delta \
  --base 2f32716054f245622c2ac3c8f692d00fa51e66e9 \
  --head cc531f31875e526cade6f32a1d7e2ed49898cde2 \
  --manifest /tmp/adamic-proof-4gp/manifest.json \
  --manifest-generator cloud/admission-corpus/manifest.py \
  --workers 2 --json
```

exit=0 wall_seconds=793

| Result | Value |
|---|---|
| Exit | 0 |
| Verdict | `pass` |
| Void | `false` |
| Void reasons | None, at run or program level |
| Complete corpus | `true` |
| Workers | 2 |
| diff_count | 4 |
| Admitted | 0 |
| Newly refused | 0 |
| Accepted by both | 3373 |
| Refused by both | 2004 |
| Classification stderr differs | 0 programs |
| Runtime timeouts / retries / still timed out | 0 / 0 / 0 |

Compared `base_compile.stderr` and `compile.stderr` as complete strings for every one of 5377 unique programs: no differences, so there are no differing texts to list. No program-level void reason and no classification or runtime timeout was present; no timeout rerun at either SHA was needed.

Diff programs:

- `internal/oracle/testdata/accessor_dispatch_absent/presence.a` (blob `482aae6df715f3c4e6988aac492b4b26921d05ef`)
- `internal/oracle/testdata/accessor_dispatch_absent/presence.ts` (blob `482aae6df715f3c4e6988aac492b4b26921d05ef`)
- `internal/oracle/testdata/accessor_dispatch_absent/undefined_compare.a` (blob `8f509852a71da04ee675f1b2f07011dcb8d1df2e`)
- `internal/oracle/testdata/accessor_dispatch_absent/undefined_compare.ts` (blob `8f509852a71da04ee675f1b2f07011dcb8d1df2e`)

Gate stderr:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 4
```

Phase timings (seconds):

```json
{
  "base_build": 22.442851683,
  "checkout": 8.966929494,
  "checkout_and_provision": 9.613328494,
  "classification": 445.875861979,
  "head_build": 13.413989619,
  "manifest_and_verification": 11.396306841,
  "provision": 0.646373773,
  "runtime": 287.705776215,
  "total": 790.453665949
}
```

## Evidence and publication

All raw execution logs remain outside the repository. SHA-256 hashes:

| Evidence | SHA-256 |
|---|---|
| `/tmp/adamic-proof-4gp/head.log` | `e50c3243d5fcdf5f8eceaeeea9ba8eda1ec2904d09fa51468346c7feb0410c74` |
| `/tmp/adamic-proof-4gp/base.log` | `7e9f51fd6b659087d5f47cfd1bd8024d59879ff3decbe45a0a4f649715869fae` |
| `/tmp/adamic-proof-4gp/base-detail.log` | `594b739ad6e0fa5c17d4fee392b2a431f301b3fa9a36c65861297788bcdcddd1` |
| `/tmp/adamic-proof-4gp/mutant.log` | `1fd2309c135519d9b380eca0bc3fed02717e8ce1a165a3b35c7658470a0aa74a` |
| `/tmp/adamic-proof-4gp/mutant-detail.log` | `92b469c0fc95245801e52a39538d92082f0114601312f318164bc30005731c20` |
| `/tmp/adamic-proof-4gp/ir.log` | `40000be913ff90b2809783e5055875f13b77354847b1546954d25e9a398a13a5` |
| `/tmp/adamic-proof-4gp/manifest.json` | `ca59d19d64bfdeb2a8549f9ad8370be85221d00d8453eff60ffd3dd8efd1d295` |
| `/tmp/adamic-proof-4gp/admission.json` | `7fbb964fd587806986209415b057b83a92ba4210fb65dd650b5077db908c9dff` |
| `/tmp/adamic-proof-4gp/admission.stderr` | `ffb1f830f236b8294b0fcdc234d58bf3aaca60ffaa0e832f6ed99248153300c0` |
| `/tmp/lower.log` | `97e3a09b55fedc50de8bc919fcb97fff35f5203cdf6e90d9775a861d5da5b449` |

No implementation, fixture, test or dependency file was committed. All temporary test logging, base overlays and the mutant were restored; the report branch contains only this report relative to the requested HEAD. No branch other than `proof/lower-4gp-1-report` was pushed.
