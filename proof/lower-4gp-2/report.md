# Proof for #4gpbn7w — accessor dispatch absent fallback

HEAD: `8c35eec869fcd0d90f03a14361137b40a57b50d4` (compiler/accessor-dispatch-absent-2).
Base/main: `d57d3041901299743362fc8c52ac06dbb2db1d0e`.

## Decisive results

Overall: the requested proof is not fully green. The full lower suite fails one stored-C agreement test; IR guards pass and the complete admission delta passes. HEAD oracle pins pass. Base and the specified mutant fail all four accessor fixture subtests with AddressSanitizer heap-buffer-overflow, READ of size 8 immediately after a 48-byte allocation. No JSON pin fails. The exact native selector matches zero tests on HEAD, base, and mutant; exit 0 does not establish native pin coverage. The 14 JSON wasm backend `/3` subtests skip because ADAMIC_ORACLE_WASI is unset.

## Setup and step 1

Loaded `/workspace/adamic-tools/env.sh` before repository work. Fetched origin compiler/accessor-dispatch-absent-2, checked out HEAD detached, and completed `git submodule update --init --recursive` (exit 0). Submodules: cohere 5360753e19e4417675a90e92a1823bea66b13918; cohere/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f. No AGENTS.md files found in the repository or its ancestor paths.

`gofmt -l internal/lower internal/oracle`: exit 0, empty output.
`go vet ./internal/lower ./internal/oracle`: exit 0, 80 seconds, empty output. An initial timing wrapper failed because /usr/bin/time is absent before vet ran; Bash timing was used for the actual run.

Created `/tmp/adamic-gate` outside the repository (env.sh selects it as TMPDIR). Provisioned upstream https://github.com/microsoft/TypeScript.git at 050880ce59e30b356b686bd3144efe24f875ebc8 in `/tmp/adamic-proof-typescript`; exported ADAMIC_TYPESCRIPT_SOURCE to that directory. Copied stage3/api package.json and package-lock.json to `/tmp/adamic-proof-api`, ran npm ci there, and placed its node_modules/.bin on PATH. tsc 6.0.3; Node v24.19.0; Go go1.27.1 linux/amd64; clang 20.1.8.

Exported NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt, npm_config_cafile=/etc/ssl/certs/ca-certificates.crt, npm_config_strict_ssl=true for commands. External npm wrapper re-exports them because toolchain.Env filters child environments. Heavy commands run serially. No compiler fixes or unrelated commits are made.

## Step 2 — pins and base control

Oracle command:
```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestAccessorDispatchAbsentAgreesWithNode|TestJSONObjects' -count=1 -v -timeout 40m
```
Native command:
```sh
go test ./internal/native -run 'TestNullableSlot|TestJSONObjectsRuntime' -count=1 -v
```

HEAD oracle: exit 0, 115 seconds wall (Go test reports 105.711s). HEAD native: exit 0, 14 seconds wall, warning: no tests to run.

Base external worktree: `/workspace/scratch/adamic-proof-base`, checked out detached at the exact base. Copied HEAD's accessor_dispatch_absent_test.go and the four accessor_dispatch_absent fixtures into that worktree without staging or committing. Provisioned pinned submodules from local object stores. An initial symbolic-link attempt was rejected by Git and replaced with actual directories before any test ran. Base oracle: exit 1, Go test reports 18.130s. Base native: exit 0, no tests to run.

### Every oracle subtest result

Parent test outcomes and all leaf subtests are listed; runtime variants inside each accessor subtest are release/native, sanitized, and JavaScript. HEAD passes all three; base and mutant fail only sanitized, while native and JavaScript agree.

#### head

| Result | Test |
|---|---|
| PASS | `TestJSONObjectsRefused` |
| PASS | `TestJSONObjectsRefused/graph_type.a` |
| PASS | `TestJSONObjectsRefused/to_json_method.a` |
| PASS | `TestJSONObjectsRefused/weak_field.a` |
| PASS | `TestJSONObjectsRefused/typeof_class.a` |
| PASS | `TestJSONObjectsRefused/to_json_view.a` |
| PASS | `TestJSONObjectsRefused/to_json_computed.a` |
| PASS | `TestJSONObjectsRefused/to_json_inherited.a` |
| PASS | `TestJSONObjectsRefused/library_map.a` |
| PASS | `TestJSONObjectsRefused/library_date.a` |
| PASS | `TestAccessorDispatchAbsentAgreesWithNode` |
| PASS | `TestAccessorDispatchAbsentAgreesWithNode/presence.ts` |
| PASS | `TestAccessorDispatchAbsentAgreesWithNode/undefined_compare.ts` |
| PASS | `TestAccessorDispatchAbsentAgreesWithNode/presence.a` |
| PASS | `TestAccessorDispatchAbsentAgreesWithNode/undefined_compare.a` |
| PASS | `TestJSONObjectsAgreeWithNode` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/0` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/null_variable/3` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/2` |
| SKIP | `TestJSONObjectsAgreeWithNode/decoded/3` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/1` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/2` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/4` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/0` |
| SKIP | `TestJSONObjectsAgreeWithNode/nonfinite/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/4` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/1` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/2` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/functions/3` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/0` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/4` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/2` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/0` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/nullish/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/2` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/0` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/hidden/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/1` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/2` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/0` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/view_order/3` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/1` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/4` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/2` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/0` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/optional/3` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/4` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/2` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/0` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/nested/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/2` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/4` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/0` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/class/3` |
| PASS | `TestJSONObjectsAgreeWithNode/class/2` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/0` |
| PASS | `TestJSONObjectsAgreeWithNode/class/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/literal/3` |
| PASS | `TestJSONObjectsAgreeWithNode/class/1` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/4` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/0` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/2` |
| PASS | `TestJSONObjectsAgreeWithNode/class/0` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/logical_and_null/3` |
| SKIP | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/3` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/2` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/1` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/4` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/4` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/2` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/0` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/4` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/undefined_variable/3` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/2` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/0` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/1` |

#### base

| Result | Test |
|---|---|
| PASS | `TestJSONObjectsRefused` |
| PASS | `TestJSONObjectsRefused/typeof_class.a` |
| PASS | `TestJSONObjectsRefused/graph_type.a` |
| PASS | `TestJSONObjectsRefused/to_json_method.a` |
| PASS | `TestJSONObjectsRefused/to_json_computed.a` |
| PASS | `TestJSONObjectsRefused/weak_field.a` |
| PASS | `TestJSONObjectsRefused/to_json_view.a` |
| PASS | `TestJSONObjectsRefused/to_json_inherited.a` |
| PASS | `TestJSONObjectsRefused/library_map.a` |
| PASS | `TestJSONObjectsRefused/library_date.a` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode/presence.ts` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode/presence.a` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode/undefined_compare.a` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode/undefined_compare.ts` |
| PASS | `TestJSONObjectsAgreeWithNode` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/0` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/1` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/4` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/0` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/nullish/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/2` |
| SKIP | `TestJSONObjectsAgreeWithNode/hidden/3` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/2` |
| SKIP | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/3` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/4` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/1` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/0` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/2` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/1` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/view_order/3` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/0` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/4` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/2` |
| SKIP | `TestJSONObjectsAgreeWithNode/logical_and_null/3` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/1` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/2` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/4` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/0` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/1` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/0` |
| SKIP | `TestJSONObjectsAgreeWithNode/optional/3` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/4` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/4` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/0` |
| SKIP | `TestJSONObjectsAgreeWithNode/undefined_variable/3` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/2` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/2` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/1` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/1` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/0` |
| SKIP | `TestJSONObjectsAgreeWithNode/nested/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/4` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/2` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/4` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/class/3` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/0` |
| PASS | `TestJSONObjectsAgreeWithNode/class/4` |
| PASS | `TestJSONObjectsAgreeWithNode/class/2` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/0` |
| SKIP | `TestJSONObjectsAgreeWithNode/null_variable/3` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/2` |
| PASS | `TestJSONObjectsAgreeWithNode/class/1` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/1` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/literal/3` |
| PASS | `TestJSONObjectsAgreeWithNode/class/0` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/2` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/4` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/1` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/0` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/2` |
| SKIP | `TestJSONObjectsAgreeWithNode/decoded/3` |
| SKIP | `TestJSONObjectsAgreeWithNode/functions/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/1` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/2` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/4` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/2` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/0` |
| SKIP | `TestJSONObjectsAgreeWithNode/nonfinite/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/4` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/1` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/0` |

#### mutant

| Result | Test |
|---|---|
| PASS | `TestJSONObjectsRefused` |
| PASS | `TestJSONObjectsRefused/graph_type.a` |
| PASS | `TestJSONObjectsRefused/to_json_method.a` |
| PASS | `TestJSONObjectsRefused/weak_field.a` |
| PASS | `TestJSONObjectsRefused/typeof_class.a` |
| PASS | `TestJSONObjectsRefused/to_json_view.a` |
| PASS | `TestJSONObjectsRefused/to_json_computed.a` |
| PASS | `TestJSONObjectsRefused/to_json_inherited.a` |
| PASS | `TestJSONObjectsRefused/library_map.a` |
| PASS | `TestJSONObjectsRefused/library_date.a` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode/presence.a` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode/presence.ts` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode/undefined_compare.ts` |
| FAIL | `TestAccessorDispatchAbsentAgreesWithNode/undefined_compare.a` |
| PASS | `TestJSONObjectsAgreeWithNode` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/0` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/0` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/3` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/2` |
| SKIP | `TestJSONObjectsAgreeWithNode/logical_and_null/3` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/4` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/2` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/1` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_coalesce_undefined/0` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/undefined_variable/3` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/4` |
| PASS | `TestJSONObjectsAgreeWithNode/logical_and_null/0` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/2` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/nonfinite/3` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/4` |
| PASS | `TestJSONObjectsAgreeWithNode/undefined_variable/0` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/4` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/2` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/functions/3` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/4` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/2` |
| PASS | `TestJSONObjectsAgreeWithNode/functions/1` |
| PASS | `TestJSONObjectsAgreeWithNode/nonfinite/0` |
| SKIP | `TestJSONObjectsAgreeWithNode/null_variable/3` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/2` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/1` |
| PASS | `TestJSONObjectsAgreeWithNode/null_variable/0` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/4` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/decoded/3` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/2` |
| PASS | `TestJSONObjectsAgreeWithNode/decoded/0` |
| SKIP | `TestJSONObjectsAgreeWithNode/optional/3` |
| SKIP | `TestJSONObjectsAgreeWithNode/nullish/3` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/2` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/1` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/4` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/2` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/0` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/4` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/1` |
| PASS | `TestJSONObjectsAgreeWithNode/hidden/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/hidden/3` |
| SKIP | `TestJSONObjectsAgreeWithNode/view_order/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nullish/0` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/1` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/2` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/2` |
| PASS | `TestJSONObjectsAgreeWithNode/class/4` |
| PASS | `TestJSONObjectsAgreeWithNode/view_order/0` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/1` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/2` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/0` |
| PASS | `TestJSONObjectsAgreeWithNode/optional/4` |
| SKIP | `TestJSONObjectsAgreeWithNode/nested/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/4` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/1` |
| PASS | `TestJSONObjectsAgreeWithNode/class/2` |
| SKIP | `TestJSONObjectsAgreeWithNode/class/3` |
| PASS | `TestJSONObjectsAgreeWithNode/nested/0` |
| PASS | `TestJSONObjectsAgreeWithNode/class/0` |
| PASS | `TestJSONObjectsAgreeWithNode/class/1` |
| SKIP | `TestJSONObjectsAgreeWithNode/literal/3` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/2` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/4` |
| PASS | `TestJSONObjectsAgreeWithNode/literal/1` |

### Complete got and want for every base FAIL

The canonical regression reports full sanitizer stderr but omits stdout/exit values. A supplemental external Go overlay runs the same onNode/lowered/nativelyUncached helpers and records all three observable fields, serially, without editing tracked files. The JSON below contains the complete unabridged got and want from that supplemental run; ASan addresses/PIDs may differ from the canonical run. Empty stdout is due to abort before buffered stdout flush. These capture tests log observations and therefore have exit 0; the canonical tests above are the failure verdict.

#### presence.a — sanitized FAIL

```json
{
  "fixture": "presence.a",
  "got": {
    "exit": 1,
    "stderr": "=================================================================\n==13052==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7ba83a1e0080 at pc 0x56520de92daf bp 0x7ffc6ecfefb0 sp 0x7ffc6ecfefa8\nREAD of size 8 at 0x7ba83a1e0080 thread T0\n    #0 0x56520de92dae in adamic_function_property_get_m_44616eaa941cde68afaeb5a452c9e3f0 /adamic-build/main.c:125:214\n    #1 0x56520de92dae in adamic_function_presence_a_probe_683c300499b7b9215cf076fee0f8e6ad /adamic-build/main.c:45:79\n    #2 0x56520de92b4e in adamic_initialize_presence_a_1a79fe28c2235bf6795af6d7910f7e70 /adamic-build/main.c:139:90\n    #3 0x56520de92a6d in adamic_initialize_presence_a_c326d3af3aafdd9427865fd67712f5f7 /adamic-build/main.c:155:2\n    #4 0x56520de92a52 in main /adamic-build/main.c:165:2\n    #5 0x7f683aebdca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f683aebdd64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x56520ddac590 in _start (/tmp/adamic-gate/TestProofAccessorCapturepresence.a2998304128/001/program+0x81590)\n\n0x7ba83a1e0080 is located 0 bytes after 48-byte region [0x7ba83a1e0050,0x7ba83a1e0080)\nallocated by thread T0 here:\n    #0 0x56520de4f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x56520debb11f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x56520debb11f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x56520de92a6d in adamic_initialize_presence_a_c326d3af3aafdd9427865fd67712f5f7 /adamic-build/main.c:155:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:125:214 in adamic_function_property_get_m_44616eaa941cde68afaeb5a452c9e3f0\nShadow bytes around the buggy address:\n  0x7ba83a1dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7ba83a1dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7ba83a1dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7ba83a1dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7ba83a1e0000: fa fa fd fd fd fd fd fd fa fa 00 00 00 00 00 00\n=>0x7ba83a1e0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7ba83a1e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7ba83a1e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7ba83a1e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7ba83a1e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7ba83a1e0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==13052==ABORTING\n",
    "stdout": ""
  },
  "want": {
    "exit": 0,
    "stderr": "",
    "stdout": "has\nnone\nhas\n"
  }
}
```

#### presence.ts — sanitized FAIL

```json
{
  "fixture": "presence.ts",
  "got": {
    "exit": 1,
    "stderr": "=================================================================\n==13066==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7c1f5e5e0080 at pc 0x557ab1a9edaf bp 0x7fffdd3fa2a0 sp 0x7fffdd3fa298\nREAD of size 8 at 0x7c1f5e5e0080 thread T0\n    #0 0x557ab1a9edae in adamic_function_property_get_m_1c12238205d371ba902a20fd90a9f40e /adamic-build/main.c:125:214\n    #1 0x557ab1a9edae in adamic_function_presence_ts_probe_2e12dd534000be433c747c336a6b7270 /adamic-build/main.c:45:79\n    #2 0x557ab1a9eb4e in adamic_initialize_presence_ts_5f15e72351266b926ca00eb5927a9e42 /adamic-build/main.c:139:91\n    #3 0x557ab1a9ea6d in adamic_initialize_presence_ts_f54b48d2cf47b3ace7df7799d1627f58 /adamic-build/main.c:155:2\n    #4 0x557ab1a9ea52 in main /adamic-build/main.c:165:2\n    #5 0x7fdf5f34cca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7fdf5f34cd64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x557ab19b8590 in _start (/tmp/adamic-gate/TestProofAccessorCapturepresence.ts873205646/001/program+0x81590)\n\n0x7c1f5e5e0080 is located 0 bytes after 48-byte region [0x7c1f5e5e0050,0x7c1f5e5e0080)\nallocated by thread T0 here:\n    #0 0x557ab1a5b3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x557ab1ac711f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x557ab1ac711f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x557ab1a9ea6d in adamic_initialize_presence_ts_f54b48d2cf47b3ace7df7799d1627f58 /adamic-build/main.c:155:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:125:214 in adamic_function_property_get_m_1c12238205d371ba902a20fd90a9f40e\nShadow bytes around the buggy address:\n  0x7c1f5e5dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c1f5e5dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c1f5e5dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c1f5e5dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7c1f5e5e0000: fa fa fd fd fd fd fd fd fa fa 00 00 00 00 00 00\n=>0x7c1f5e5e0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c1f5e5e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c1f5e5e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c1f5e5e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c1f5e5e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7c1f5e5e0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==13066==ABORTING\n",
    "stdout": ""
  },
  "want": {
    "exit": 0,
    "stderr": "",
    "stdout": "has\nnone\nhas\n"
  }
}
```

#### undefined_compare.a — sanitized FAIL

```json
{
  "fixture": "undefined_compare.a",
  "got": {
    "exit": 1,
    "stderr": "=================================================================\n==13080==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7bb555fe0080 at pc 0x561ceea3d0ec bp 0x7fff2384cad0 sp 0x7fff2384cac8\nREAD of size 8 at 0x7bb555fe0080 thread T0\n    #0 0x561ceea3d0eb in adamic_function_property_get_m_577812b7b7c3dbd368c0d9fd58c3f1c3 /adamic-build/main.c:137:214\n    #1 0x561ceea3d0eb in adamic_function_undefined_compare_a_value_b095eeb2d4c7c6c6065421f758b9daed /adamic-build/main.c:45:79\n    #2 0x561ceea3cbc8 in adamic_initialize_undefined_compare_a_27fee3aa71e72e30062bafa0b963388a /adamic-build/main.c:160:94\n    #3 0x561ceea3ca6d in adamic_initialize_undefined_compare_a_0c29b8b267bd9670af5dc93b13b75d74 /adamic-build/main.c:194:2\n    #4 0x561ceea3ca52 in main /adamic-build/main.c:204:2\n    #5 0x7f7556e83ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f7556e83d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x561cee956590 in _start (/tmp/adamic-gate/TestProofAccessorCaptureundefined_compare.a665044457/001/program+0x82590)\n\n0x7bb555fe0080 is located 0 bytes after 48-byte region [0x7bb555fe0050,0x7bb555fe0080)\nallocated by thread T0 here:\n    #0 0x561cee9f93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x561ceea654cf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x561ceea654cf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x561ceea3ca6d in adamic_initialize_undefined_compare_a_0c29b8b267bd9670af5dc93b13b75d74 /adamic-build/main.c:194:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:137:214 in adamic_function_property_get_m_577812b7b7c3dbd368c0d9fd58c3f1c3\nShadow bytes around the buggy address:\n  0x7bb555fdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bb555fdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bb555fdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bb555fdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7bb555fe0000: fa fa 00 00 00 00 00 00 fa fa 00 00 00 00 00 00\n=>0x7bb555fe0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bb555fe0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bb555fe0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bb555fe0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bb555fe0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7bb555fe0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==13080==ABORTING\n",
    "stdout": ""
  },
  "want": {
    "exit": 0,
    "stderr": "",
    "stdout": "5550\n"
  }
}
```

#### undefined_compare.ts — sanitized FAIL

```json
{
  "fixture": "undefined_compare.ts",
  "got": {
    "exit": 1,
    "stderr": "=================================================================\n==13094==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b56083e0080 at pc 0x55ab41d9c0ec bp 0x7ffe86b180e0 sp 0x7ffe86b180d8\nREAD of size 8 at 0x7b56083e0080 thread T0\n    #0 0x55ab41d9c0eb in adamic_function_property_get_m_1628164efe41e8c5c544e1f16a0eb7d3 /adamic-build/main.c:137:214\n    #1 0x55ab41d9c0eb in adamic_function_undefined_compare_ts_value_42834c9f3e37389f1ac8821c5f76bead /adamic-build/main.c:45:79\n    #2 0x55ab41d9bbc8 in adamic_initialize_undefined_compare_ts_7557cfe860544855e6a9a9e6debb0219 /adamic-build/main.c:160:95\n    #3 0x55ab41d9ba6d in adamic_initialize_undefined_compare_ts_f02d3d6ed7057c4d8d72f313b62471a2 /adamic-build/main.c:194:2\n    #4 0x55ab41d9ba52 in main /adamic-build/main.c:204:2\n    #5 0x7f16092ccca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f16092ccd64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x55ab41cb5590 in _start (/tmp/adamic-gate/TestProofAccessorCaptureundefined_compare.ts1897652188/001/program+0x82590)\n\n0x7b56083e0080 is located 0 bytes after 48-byte region [0x7b56083e0050,0x7b56083e0080)\nallocated by thread T0 here:\n    #0 0x55ab41d583a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55ab41dc44cf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55ab41dc44cf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55ab41d9ba6d in adamic_initialize_undefined_compare_ts_f02d3d6ed7057c4d8d72f313b62471a2 /adamic-build/main.c:194:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:137:214 in adamic_function_property_get_m_1628164efe41e8c5c544e1f16a0eb7d3\nShadow bytes around the buggy address:\n  0x7b56083dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b56083dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b56083dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b56083dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b56083e0000: fa fa 00 00 00 00 00 00 fa fa 00 00 00 00 00 00\n=>0x7b56083e0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56083e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56083e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56083e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56083e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56083e0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==13094==ABORTING\n",
    "stdout": ""
  },
  "want": {
    "exit": 0,
    "stderr": "",
    "stdout": "5550\n"
  }
}
```

## Step 3 — mutant

Only mutant in internal/lower/accessor-dispatch-absent.mutants.md: replace `fallback.Object = object` with `fallback.Object, fallback.Absent = object, false` in internal/lower/class_accessors.go. Asserted old text count exactly 1 before applying. Ran the exact oracle and native commands above, sequentially, then the supplemental output capture. Oracle: exit 1; wall timing:

```text
wall_seconds=35 exit=1
```
Native: exit 0, no tests to run. Every red leaf is one of the four accessor fixtures listed above; each sanitized run is an ASan heap-buffer-overflow read. Complete got/want for each:

#### presence.a — sanitized FAIL

```json
{
  "fixture": "presence.a",
  "got": {
    "exit": 1,
    "stderr": "=================================================================\n==15257==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b813cfe0080 at pc 0x56401e74adaf bp 0x7ffcde40f470 sp 0x7ffcde40f468\nREAD of size 8 at 0x7b813cfe0080 thread T0\n    #0 0x56401e74adae in adamic_function_property_get_m_d56e9fd6799340cdc74b4f86d15a0ac0 /adamic-build/main.c:125:213\n    #1 0x56401e74adae in adamic_function_presence_a_probe_683c300499b7b9215cf076fee0f8e6ad /adamic-build/main.c:45:79\n    #2 0x56401e74ab4e in adamic_initialize_presence_a_1a79fe28c2235bf6795af6d7910f7e70 /adamic-build/main.c:139:90\n    #3 0x56401e74aa6d in adamic_initialize_presence_a_c326d3af3aafdd9427865fd67712f5f7 /adamic-build/main.c:155:2\n    #4 0x56401e74aa52 in main /adamic-build/main.c:165:2\n    #5 0x7f413dc7dca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f413dc7dd64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x56401e664590 in _start (/tmp/adamic-gate/TestProofAccessorCapturepresence.a3271552287/001/program+0x81590)\n\n0x7b813cfe0080 is located 0 bytes after 48-byte region [0x7b813cfe0050,0x7b813cfe0080)\nallocated by thread T0 here:\n    #0 0x56401e7073a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x56401e77311f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x56401e77311f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x56401e74aa6d in adamic_initialize_presence_a_c326d3af3aafdd9427865fd67712f5f7 /adamic-build/main.c:155:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:125:213 in adamic_function_property_get_m_d56e9fd6799340cdc74b4f86d15a0ac0\nShadow bytes around the buggy address:\n  0x7b813cfdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b813cfdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b813cfdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b813cfdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b813cfe0000: fa fa fd fd fd fd fd fd fa fa 00 00 00 00 00 00\n=>0x7b813cfe0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b813cfe0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b813cfe0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b813cfe0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b813cfe0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b813cfe0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==15257==ABORTING\n",
    "stdout": ""
  },
  "want": {
    "exit": 0,
    "stderr": "",
    "stdout": "has\nnone\nhas\n"
  }
}
```

#### presence.ts — sanitized FAIL

```json
{
  "fixture": "presence.ts",
  "got": {
    "exit": 1,
    "stderr": "=================================================================\n==15271==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b56b1fe0080 at pc 0x563519884daf bp 0x7ffca2454620 sp 0x7ffca2454618\nREAD of size 8 at 0x7b56b1fe0080 thread T0\n    #0 0x563519884dae in adamic_function_property_get_m_b5c7fd8aaea146ee758414c94a1d6119 /adamic-build/main.c:125:213\n    #1 0x563519884dae in adamic_function_presence_ts_probe_2e12dd534000be433c747c336a6b7270 /adamic-build/main.c:45:79\n    #2 0x563519884b4e in adamic_initialize_presence_ts_5f15e72351266b926ca00eb5927a9e42 /adamic-build/main.c:139:91\n    #3 0x563519884a6d in adamic_initialize_presence_ts_f54b48d2cf47b3ace7df7799d1627f58 /adamic-build/main.c:155:2\n    #4 0x563519884a52 in main /adamic-build/main.c:165:2\n    #5 0x7f16b2d19ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f16b2d19d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x56351979e590 in _start (/tmp/adamic-gate/TestProofAccessorCapturepresence.ts3971919032/001/program+0x81590)\n\n0x7b56b1fe0080 is located 0 bytes after 48-byte region [0x7b56b1fe0050,0x7b56b1fe0080)\nallocated by thread T0 here:\n    #0 0x5635198413a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x5635198ad11f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x5635198ad11f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x563519884a6d in adamic_initialize_presence_ts_f54b48d2cf47b3ace7df7799d1627f58 /adamic-build/main.c:155:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:125:213 in adamic_function_property_get_m_b5c7fd8aaea146ee758414c94a1d6119\nShadow bytes around the buggy address:\n  0x7b56b1fdfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b56b1fdfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b56b1fdff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b56b1fdff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b56b1fe0000: fa fa fd fd fd fd fd fd fa fa 00 00 00 00 00 00\n=>0x7b56b1fe0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56b1fe0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56b1fe0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56b1fe0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56b1fe0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b56b1fe0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==15271==ABORTING\n",
    "stdout": ""
  },
  "want": {
    "exit": 0,
    "stderr": "",
    "stdout": "has\nnone\nhas\n"
  }
}
```

#### undefined_compare.a — sanitized FAIL

```json
{
  "fixture": "undefined_compare.a",
  "got": {
    "exit": 1,
    "stderr": "=================================================================\n==15285==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7be9e85e0080 at pc 0x5591f5c3e0ec bp 0x7fff73596360 sp 0x7fff73596358\nREAD of size 8 at 0x7be9e85e0080 thread T0\n    #0 0x5591f5c3e0eb in adamic_function_property_get_m_3e3d69c65abdb5017ffb80ed85afce1d /adamic-build/main.c:137:214\n    #1 0x5591f5c3e0eb in adamic_function_undefined_compare_a_value_b095eeb2d4c7c6c6065421f758b9daed /adamic-build/main.c:45:79\n    #2 0x5591f5c3dbc8 in adamic_initialize_undefined_compare_a_27fee3aa71e72e30062bafa0b963388a /adamic-build/main.c:160:94\n    #3 0x5591f5c3da6d in adamic_initialize_undefined_compare_a_0c29b8b267bd9670af5dc93b13b75d74 /adamic-build/main.c:194:2\n    #4 0x5591f5c3da52 in main /adamic-build/main.c:204:2\n    #5 0x7fa9e942eca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7fa9e942ed64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x5591f5b57590 in _start (/tmp/adamic-gate/TestProofAccessorCaptureundefined_compare.a3731563845/001/program+0x82590)\n\n0x7be9e85e0080 is located 0 bytes after 48-byte region [0x7be9e85e0050,0x7be9e85e0080)\nallocated by thread T0 here:\n    #0 0x5591f5bfa3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x5591f5c664cf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x5591f5c664cf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x5591f5c3da6d in adamic_initialize_undefined_compare_a_0c29b8b267bd9670af5dc93b13b75d74 /adamic-build/main.c:194:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:137:214 in adamic_function_property_get_m_3e3d69c65abdb5017ffb80ed85afce1d\nShadow bytes around the buggy address:\n  0x7be9e85dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be9e85dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be9e85dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be9e85dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7be9e85e0000: fa fa 00 00 00 00 00 00 fa fa 00 00 00 00 00 00\n=>0x7be9e85e0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be9e85e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be9e85e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be9e85e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be9e85e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7be9e85e0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==15285==ABORTING\n",
    "stdout": ""
  },
  "want": {
    "exit": 0,
    "stderr": "",
    "stdout": "5550\n"
  }
}
```

#### undefined_compare.ts — sanitized FAIL

```json
{
  "fixture": "undefined_compare.ts",
  "got": {
    "exit": 1,
    "stderr": "=================================================================\n==15299==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7b9b181e0080 at pc 0x5573d42ae0ec bp 0x7ffeafc45380 sp 0x7ffeafc45378\nREAD of size 8 at 0x7b9b181e0080 thread T0\n    #0 0x5573d42ae0eb in adamic_function_property_get_m_a4af98e966cd9f61b1187919fd9a18fe /adamic-build/main.c:137:214\n    #1 0x5573d42ae0eb in adamic_function_undefined_compare_ts_value_42834c9f3e37389f1ac8821c5f76bead /adamic-build/main.c:45:79\n    #2 0x5573d42adbc8 in adamic_initialize_undefined_compare_ts_7557cfe860544855e6a9a9e6debb0219 /adamic-build/main.c:160:95\n    #3 0x5573d42ada6d in adamic_initialize_undefined_compare_ts_f02d3d6ed7057c4d8d72f313b62471a2 /adamic-build/main.c:194:2\n    #4 0x5573d42ada52 in main /adamic-build/main.c:204:2\n    #5 0x7f5b1902bca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x7f5b1902bd64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #7 0x5573d41c7590 in _start (/tmp/adamic-gate/TestProofAccessorCaptureundefined_compare.ts1118169499/001/program+0x82590)\n\n0x7b9b181e0080 is located 0 bytes after 48-byte region [0x7b9b181e0050,0x7b9b181e0080)\nallocated by thread T0 here:\n    #0 0x5573d426a3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x5573d42d64cf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x5573d42d64cf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x5573d42ada6d in adamic_initialize_undefined_compare_ts_f02d3d6ed7057c4d8d72f313b62471a2 /adamic-build/main.c:194:2\n\nSUMMARY: AddressSanitizer: heap-buffer-overflow /adamic-build/main.c:137:214 in adamic_function_property_get_m_a4af98e966cd9f61b1187919fd9a18fe\nShadow bytes around the buggy address:\n  0x7b9b181dfe00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9b181dfe80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9b181dff00: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9b181dff80: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n  0x7b9b181e0000: fa fa 00 00 00 00 00 00 fa fa 00 00 00 00 00 00\n=>0x7b9b181e0080:[fa]fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9b181e0100: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9b181e0180: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9b181e0200: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9b181e0280: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\n  0x7b9b181e0300: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa\nShadow byte legend (one shadow byte represents 8 application bytes):\n  Addressable:           00\n  Partially addressable: 01 02 03 04 05 06 07 \n  Heap left redzone:       fa\n  Freed heap region:       fd\n  Stack left redzone:      f1\n  Stack mid redzone:       f2\n  Stack right redzone:     f3\n  Stack after return:      f5\n  Stack use after scope:   f8\n  Global redzone:          f9\n  Global init order:       f6\n  Poisoned by user:        f7\n  Container overflow:      fc\n  Array cookie:            ac\n  Intra object redzone:    bb\n  ASan internal:           fe\n  Left alloca redzone:     ca\n  Right alloca redzone:    cb\n==15299==ABORTING\n",
    "stdout": ""
  },
  "want": {
    "exit": 0,
    "stderr": "",
    "stdout": "5550\n"
  }
}
```

Restored class_accessors.go from original HEAD; `git diff --exit-code` exit 0, empty output.

## Step 4 — full lower suite

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Exit **1**; wall time **617 seconds** (10m17s); Go test time 604.133s. One failing test. No environment failures.

| Failing test | First error line | Environment? |
|---|---|---|
| TestNativeAgreementLeanBuildIsBuildsBytes | agree_test.go:143: ObjectDestructuringAssignmentGetter: the lower agreement C holds 26427 bytes of C, and this test lowers 26427 other bytes from its source | No: stored C differs from current lowering, despite equal lengths |

The preceding agree_test.go:126/:163/:192 messages are build and successful case logs, not error lines. Full lower output:

```text
--- FAIL: TestNativeAgreementLeanBuildIsBuildsBytes (18.31s)
    agree_test.go:126: build lower_agreement_C_v1 ba9001407e78 hit 0.01
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.02
    agree_test.go:192: ArrayNarrowingBooleanCallback, features []: 643984 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.03
    agree_test.go:192: ArrayNarrowingFindLastIndex, features []: 644384 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.10
    agree_test.go:192: ArrayNarrowingOptionalNumber, features []: 643936 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.02
    agree_test.go:192: ArrayNarrowingOptionalString, features []: 644128 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.06
    agree_test.go:192: ArrayNarrowingPreservedUnion, features []: 644664 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.07
    agree_test.go:192: ArrayNarrowingSomeNegated, features []: 644664 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.02
    agree_test.go:192: FX5MixedLiteral, features []: 643896 bytes, the same six ways
    agree_test.go:143: ObjectDestructuringAssignmentGetter: the lower agreement C holds 26427 bytes of C, and this test lowers 26427 other bytes from its source
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.04
    agree_test.go:192: ObjectDestructuringAssignmentSwap, features []: 644776 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.10
    agree_test.go:192: ObjectDestructuringAssignmentWiden, features []: 650136 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.06
    agree_test.go:192: PlantedWrongExit, features []: 643664 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.05
    agree_test.go:192: PlantedWrongOutput, features []: 643664 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.07
    agree_test.go:192: RegExpEscapedSlashSource, features []: 645472 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.07
    agree_test.go:192: RegExpSlashClassSource, features []: 645560 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.02
    agree_test.go:192: RegExpUnicodeClassSource, features []: 645472 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.08
    agree_test.go:192: TypedArrayFromArray, features []: 643704 bytes, the same six ways
    agree_test.go:163: build native-runtime a9be2110e4cb hit 0.05
    agree_test.go:192: TypedArrayFromEmptyArray, features []: 643608 bytes, the same six ways
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	604.133s
FAIL
```

## Step 4b — IR source and field guards

```sh
go test ./internal/ir -count=1 -timeout 30m
```

Exit **0**; wall **56 seconds**; Go test time 41.675s. No failing tests. This full suite includes TestCallTargetReadersLower and the IR field guards. Output:

```text
ok  	github.com/system-inc/adamic/internal/ir	41.675s
```

## Supplemental native coverage for the empty selector

The requested TestNullableSlot|TestJSONObjectsRuntime selector matches no current test names. To exercise the existing tests in the relevant source files, additionally ran on HEAD:

```sh
go test ./internal/native -run 'TestAbsentNullableFieldMetadata|TestJSONObjectsNullPointer' -count=1 -v
```

Exit **0**, wall **3 seconds**, Go test 1.414s. PASS: TestJSONObjectsNullPointers, TestJSONObjectsNullPointerMutants, TestAbsentNullableFieldMetadata, and its null, undefined, and uninitialized_union subtests. No skips or failures. This supplements, and does not relabel, the requested selector's zero-test result.

## Step 6 — complete admission delta

Generated the manifest with the repository's own cloud/admission-corpus/manifest.py at the requested HEAD (generator blob 09d061a03b1e016387e975d25ee0cd5b75ac6c1f). The report-only branch does not enter this comparison: both --base and --head are explicit original SHAs.

```sh
python3 cloud/admission-corpus/manifest.py --sha 8c35eec869fcd0d90f03a14361137b40a57b50d4 --repository /workspace/adamic > /tmp/adamic-proof-4gp/manifest.json
go run ./cmd/adamic-admission-delta --base d57d3041901299743362fc8c52ac06dbb2db1d0e --head 8c35eec869fcd0d90f03a14361137b40a57b50d4 --manifest /tmp/adamic-proof-4gp/manifest.json --workers 2 --json
```

Ran alone, with no other proof/build/test command running. Exported both CA variables and strict_ssl=true, with the external npm wrapper for the gate's filtered npm environment. The gate provisions its temporary checkout's API install from its own pinned lockfile; the invoking repository's packages remain outside the repository.

| Measure | Result |
|---|---|
| Exit | 0 |
| Wall time | 509 seconds (8m29s) |
| Verdict | pass |
| Void | false |
| Void reasons | none (no void_reason field; no per-program void_reason) |
| diff_count | 4 |
| admitted | 0 |
| newly refused | 0 |
| accepted-by-both | 3380 |
| refused-by-both | 2006 |
| Programs with classification stderr differences | 0 |
| complete_corpus | true |
| workers | 2 |
| classification | lowering-only |
| omitted | 0 |
| budget | 0 (unbounded) |
| negative witnesses / accepted witnesses | 50 / 50 |
| Node agreements | 13 |
| ruled run-time stops | 52 |
| stale / pending run-time stops | 0 / 0 |
| runtime timeout runs / retried / still timed out | 0 / 0 / 0 |

Corpus sizes: witnesses 2, fixtures 4438, gaps 105, review 20, fuzz 0, stage3 821; total 5386. Compile timeout 45s; runtime timeout 10s. No compile observations or runtime observations contain error=timeout. No isolated timeout reruns are needed at either SHA.

The four changed programs:

- `internal/oracle/testdata/accessor_dispatch_absent/presence.a` (blob `482aae6df715f3c4e6988aac492b4b26921d05ef`)
- `internal/oracle/testdata/accessor_dispatch_absent/presence.ts` (blob `482aae6df715f3c4e6988aac492b4b26921d05ef`)
- `internal/oracle/testdata/accessor_dispatch_absent/undefined_compare.a` (blob `8f509852a71da04ee675f1b2f07011dcb8d1df2e`)
- `internal/oracle/testdata/accessor_dispatch_absent/undefined_compare.ts` (blob `8f509852a71da04ee675f1b2f07011dcb8d1df2e`)

Every program's base_compile.stderr was compared byte-for-byte with compile.stderr: **zero differences**, so there are no differing pairs of texts to list. All classification classes are accepted-by-both or refused-by-both. No program was newly admitted or refused.

Gate stderr:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 4
```

## Report publication and scope

First report after steps 1–3: d5b34e072, pushed to origin/proof/lower-4gp-2-report before starting step 4. This final update adds steps 4, 4b, 6 and supplemental native coverage. Only proof/lower-4gp-2/report.md is committed, on proof/lower-4gp-2-report created from the requested HEAD; no other branch is pushed. Compiler source and mutants were restored before the first report commit. The full lower and IR suites ran on that report branch with source identical to the requested HEAD. No implementation or generated-C fixture was changed to hide the lower failure.
