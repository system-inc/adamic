# Compiler proof: #mez92e7

## Decisive results

- Step 2: PASS, all 11 pin subtests; exit 0; wall time 155 seconds.
- Step 1: PASS, gofmt emitted nothing and go vet exited 0.
- Step 3: all four mutants detected (exit 1 each), restored, and clean-diff checked. Mutant 2 panicked; mutant 3 had an additional red.
- Step 4: PASS, full lower suite exit 0; wall time 844 seconds (14m 04s); no failures.
- Step 4b: pending.

## Identity and environment

Requested and tested source HEAD: `30379eb97798ece65ececd0693c2e0783fbbf6b7`, branch `compiler/record-storage-roads-2`, based on main `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`.

Loaded `/workspace/adamic-tools/env.sh` first. Created its missing `TMPDIR`, `/tmp/adamic-gate`, outside the repository. Commands run sequentially, one heavy command at a time. No compiler changes committed. Only this report is committed and only `proof/lower-mez2-a-report` is pushed.

Provisioned upstream `https://github.com/microsoft/TypeScript.git` outside the repository at `/tmp/mez92-proof/typescript`; verified HEAD `050880ce59e30b356b686bd3144efe24f875ebc8`; exported `ADAMIC_TYPESCRIPT_SOURCE` to that directory. Copied stage3/api's package.json and package-lock.json to `/tmp/mez92-proof/api`, ran `npm ci` there (3 packages), and prepended its node_modules/.bin to PATH. Verified tsc 6.0.3, Node v24.19.0, Go go1.27.1 linux/amd64.

For npm and every test process/child, exported:

```sh
NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
npm_config_strict_ssl=true
```

## Step 1 — checkout, submodules, formatting, vet

`git fetch origin compiler/record-storage-roads-2`: exit 0.
`git checkout --detach 30379eb97798ece65ececd0693c2e0783fbbf6b7`: exit 0.
`git submodule update --init --recursive`: exit 0. Cohere `5360753e19e4417675a90e92a1823bea66b13918`; nested TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.
`gofmt -l internal/lower internal/oracle`: exit 0, no output (asserted empty).
`go vet ./internal/lower ./internal/oracle`: exit 0, no output.

## Step 2 — unmodified pins

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRecordStorageRoadsArePinned' -count=1 -v -timeout 40m
```

Exit 0; wall time 155 seconds. All 11 subtests PASS; no failures and therefore no failing got/want pairs. Complete output follows verbatim:

```text
=== RUN   TestRecordStorageRoadsArePinned
=== PAUSE TestRecordStorageRoadsArePinned
=== CONT  TestRecordStorageRoadsArePinned
=== RUN   TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== RUN   TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== PAUSE TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== RUN   TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== PAUSE TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== RUN   TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== RUN   TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== RUN   TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== NAME  TestRecordStorageRoadsArePinned/c3_generic_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 9418edb2ee6d miss 118.60
=== NAME  TestRecordStorageRoadsArePinned/c2_logical_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 9418edb2ee6d hit 118.60
=== NAME  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 9418edb2ee6d hit 118.60
--- PASS: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (22.91s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (22.91s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (22.92s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (22.92s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (0.19s)
    --- PASS: TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts (0.20s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (0.26s)
    --- PASS: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.22s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (119.55s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (119.35s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (142.46s)
PASS
gate cache: native hits=0 misses=6
gate cache: node hits=0 misses=14
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	142.487s
```

## Step 3 — four independent mutants

Read Old/New directly from `internal/lower/record-storage-roads.mutants.md`; removed markdown escaping before pipe characters. For each edit, asserted the Old text occurred exactly once, replaced it once, ran the exact step-2 command with ADAMIC_GATE_UNCACHED=1, restored the original bytes in a finally block, then ran `git diff --exit-code`. Every restoration check exited 0 with no output.

All red subtests and complete got/want diagnostics are retained verbatim in the logs below. ASan errors and the native-emitter panic are effects of the mutants, not environment failures. The observed outcome differs from the table narrative: mutant 2 panics on &&=, and mutant 4 stops at a different position rather than lowering to a native crash.

### Mutant 1: The shorthand dropped

File: `internal/lower/records.go`. Old occurrence count: 1. Test exit: 1; wall time: 45.19 seconds; restore diff exit: 0.

Old:

```go
} else if node.Kind == ast.KindShorthandPropertyAssignment {
```

New:

```go
} else if false && node.Kind == ast.KindShorthandPropertyAssignment {
```

Red subtests: `f02_shorthand.ts`. Completed PASS subtests: 10.

Complete command output, verbatim:

```text
=== RUN   TestRecordStorageRoadsArePinned
=== PAUSE TestRecordStorageRoadsArePinned
=== CONT  TestRecordStorageRoadsArePinned
=== RUN   TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== RUN   TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== PAUSE TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== RUN   TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== PAUSE TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== RUN   TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== RUN   TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== RUN   TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== NAME  TestRecordStorageRoadsArePinned/f02_shorthand.ts
    record_storage_roads_test.go:52: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f02_shorthand.ts:6:32: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==9250==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x55a1268a2e1e bp 0x7ffd97909080 sp 0x7ffd97908ea0 T0)\n==9250==The signal is caused by a READ memory access.\n==9250==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x55a1268a2e1e in find /adamic-runtime/map.c:107:11\n    #1 0x55a1268a3385 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x55a1268ce5ba in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x55a1268ce5ba in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x55a126858ddb in adamic_function_f02_shorthand_ts_fill_8ba76ae5db5816cc91ce0e622cbcb2e2 /adamic-build/main.c:34:2\n    #5 0x55a126858ddb in adamic_initialize_f02_shorthand_ts_534c7785d2f286e4771872e3b209d0b2 /adamic-build/main.c:86:97\n    #6 0x55a126858aad in adamic_initialize_f02_shorthand_ts_59fad384c6e4960bf1f1e644ebd4a028 /adamic-build/main.c:101:2\n    #7 0x55a126858a52 in main /adamic-build/main.c:110:2\n    #8 0x7f013af07ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #9 0x7f013af07d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x55a126772590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf02_shorthand.ts183499002/001/program+0x81590)\n\n==9250==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x000055a12699eda0  rbp = 0x00007ffd97909080  rsp = 0x00007ffd97908ea0  \n r8 = 0x00000f6e273fc00b   r9 = 0x0000000000000000  r10 = 0x00000f6e273fc00b  r11 = 0x0000000000000000  \nr12 = 0x00007b0138c00080  r13 = 0x00007b0138c00080  r14 = 0x3ff0000000000000  r15 = 0x00000f6027180010  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==9250==ABORTING\n"
=== NAME  TestRecordStorageRoadsArePinned/c3_generic_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ef62f1e537b4 miss 16.96
=== NAME  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ef62f1e537b4 hit 16.96
=== NAME  TestRecordStorageRoadsArePinned/c2_logical_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ef62f1e537b4 hit 16.96
--- FAIL: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (3.63s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (3.66s)
    --- PASS: TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts (0.19s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (0.20s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (0.14s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (0.18s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (0.17s)
    --- FAIL: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.37s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (21.38s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (21.41s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (21.41s)
FAIL
gate cache: native hits=0 misses=7
gate cache: node hits=0 misses=14
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	21.441s
FAIL
```

### Mutant 2: The logical assignments dropped

File: `internal/lower/records.go`. Old occurrence count: 1. Test exit: 1; wall time: 26.496 seconds; restore diff exit: 0.

Old:

```go
(viewSite(node) || logicalAssignmentRight(node))
```

New:

```go
(viewSite(node) || false && logicalAssignmentRight(node))
```

Red subtests: `f05b_or_assignment.ts`, `f05_logical_assignment.ts`, `f05c_and_assignment.ts`. Completed PASS subtests: 5.

The process panicked while emitting f05c_and_assignment.ts: got `panic: native: no conversion from 14 to 4 [recovered, repanicked]`. No normal got/want line was emitted for that subtest. Its expected stop position (verbatim .stop contents) is:

```text
internal/oracle/testdata/record_storage_roads/f05c_and_assignment.ts:6:9
```

The expected diagnostic must start at that position and carry `(fixed objects and records have different storage; copy explicitly)`. Three controls were started but had no terminal PASS/FAIL before the panic; they are not counted as passes.

Complete command output, verbatim:

```text
=== RUN   TestRecordStorageRoadsArePinned
=== PAUSE TestRecordStorageRoadsArePinned
=== CONT  TestRecordStorageRoadsArePinned
=== RUN   TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== RUN   TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== PAUSE TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== RUN   TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== PAUSE TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== RUN   TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== RUN   TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== RUN   TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== NAME  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
    record_storage_roads_test.go:52: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f05b_or_assignment.ts:6:9: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==9991==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x5559a6504dde bp 0x7ffdc43e1fd0 sp 0x7ffdc43e1e00 T0)\n==9991==The signal is caused by a READ memory access.\n==9991==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x5559a6504dde in find /adamic-runtime/map.c:107:11\n    #1 0x5559a6505345 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x5559a653057a in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x5559a653057a in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x5559a64badab in adamic_function_f05b_or_assignment_ts_fill_a71bda3feff0a8c743ad774e3fae37a1 /adamic-build/main.c:28:2\n    #5 0x5559a64badab in adamic_initialize_f05b_or_assignment_ts_c84a5777cb056da4dc539c8b3d931623 /adamic-build/main.c:94:102\n    #6 0x5559a64baa9d in adamic_initialize_f05b_or_assignment_ts_a3d71d6bdeebdd74b8754fc9242e5c6b /adamic-build/main.c:109:2\n    #7 0x5559a64baa52 in main /adamic-build/main.c:117:2\n    #8 0x7fb35a556ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #9 0x7fb35a556d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x5559a63d4590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf05b_or_assignment.ts3574359670/001/program+0x81590)\n\n==9991==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x00005559a6600d60  rbp = 0x00007ffdc43e1fd0  rsp = 0x00007ffdc43e1e00  \n r8 = 0x00000f866b2fc00c   r9 = 0x0000000000000000  r10 = 0x00000f866b2fc00c  r11 = 0x0000000000000000  \nr12 = 0x00007bb358400080  r13 = 0x00000f766b080010  r14 = 0x3ff0000000000000  r15 = 0x00007c33597e0020  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==9991==ABORTING\n"
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== NAME  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
    record_storage_roads_test.go:52: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f05_logical_assignment.ts:6:9: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==9996==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x55ce64bf9dde bp 0x7fff77810a00 sp 0x7fff77810820 T0)\n==9996==The signal is caused by a READ memory access.\n==9996==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x55ce64bf9dde in find /adamic-runtime/map.c:107:11\n    #1 0x55ce64bfa345 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x55ce64c2557a in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x55ce64c2557a in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x55ce64bafdab in adamic_function_f05_logical_assignment_ts_fill_58ae384374dfb08c9d9dd94ef0c0b7f1 /adamic-build/main.c:28:2\n    #5 0x55ce64bafdab in adamic_initialize_f05_logical_assignment_ts_e08c170b5dfe4f39445d8f482a7d711b /adamic-build/main.c:93:106\n    #6 0x55ce64bafa9d in adamic_initialize_f05_logical_assignment_ts_51070b7710f2752c3d05318ba4695428 /adamic-build/main.c:108:2\n    #7 0x55ce64bafa52 in main /adamic-build/main.c:116:2\n    #8 0x7fb4e8e55ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #9 0x7fb4e8e55d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x55ce64ac9590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf05_logical_assignment.ts3124799557/001/program+0x81590)\n\n==9996==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x000055ce64cf5d60  rbp = 0x00007fff77810a00  rsp = 0x00007fff77810820  \n r8 = 0x00000f869cffc00c   r9 = 0x0000000000000000  r10 = 0x00000f869cffc00c  r11 = 0x0000000000000000  \nr12 = 0x00007bb4e6c00080  r13 = 0x00000f769cd80010  r14 = 0x3ff0000000000000  r15 = 0x00007c34e7fe0020  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==9996==ABORTING\n"
=== CONT  TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
--- FAIL: TestRecordStorageRoadsArePinned (4.91s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (4.19s)
    --- PASS: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.15s)
    --- FAIL: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (4.60s)
    --- FAIL: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (4.60s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (0.17s)
    --- PASS: TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts (0.17s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (0.12s)
    --- FAIL: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (0.14s)
panic: native: no conversion from 14 to 4 [recovered, repanicked]

goroutine 29 [running]:
testing.tRunner.func1.2({0x2c2f810, 0xb8f6b886e80})
	/workspace/adamic-tools/go/src/testing/testing.go:2123 +0x232
testing.tRunner.func1()
	/workspace/adamic-tools/go/src/testing/testing.go:2126 +0x329
panic({0x2c2f810?, 0xb8f6b886e80?})
	/workspace/adamic-tools/go/src/runtime/panic.go:859 +0x125
github.com/system-inc/adamic/internal/native.converted(0x12d58b2?, 0x4, {0xb8f6b884780, 0x52})
	/workspace/adamic/internal/native/union.go:24 +0x3d9
github.com/system-inc/adamic/internal/native.logicalConverted(0xb8f6b88a008?, 0x2c2f810?, {0xb8f6b884780?, 0x2c2f810?})
	/workspace/adamic/internal/native/taste.go:104 +0x37a
github.com/system-inc/adamic/internal/native.(*emitter).assignLogicalResult(0xb8f6b88a008, {0xb8f6b8846c0, 0x51}, 0x52?, 0x4, {0xb8f6b884780?, 0xb8f6b8beb08?})
	/workspace/adamic/internal/native/logical_assignment.go:150 +0x169
github.com/system-inc/adamic/internal/native.(*emitter).logicalAssignment(0xb8f6b88a008, {{0x2d2ba40, 0xb8f6c07bc80}, {0x2d2ba00, 0xb8f6b12b580}, {0x0, 0x0}, {0x0, 0x0, 0x0}, ...})
	/workspace/adamic/internal/native/logical_assignment.go:130 +0x1eca
github.com/system-inc/adamic/internal/native.(*emitter).evaluate(0xb8f6b88a008, {0x2d2c860?, 0xb8f6b07e480?})
	/workspace/adamic/internal/native/emit_expressions.go:115 +0xb165
github.com/system-inc/adamic/internal/native.(*emitter).value(0xb8f6b88a008, {0x2d2c860, 0xb8f6b07e480})
	/workspace/adamic/internal/native/borrow.go:41 +0x249
github.com/system-inc/adamic/internal/native.(*emitter).statement(0xb8f6b88a008, {0x2d2be80, 0xb8f6c57e120})
	/workspace/adamic/internal/native/emit_statements.go:196 +0x22f1
github.com/system-inc/adamic/internal/native.(*emitter).statementAt(0xb8f6b88a008, 0xb8f6b1301a0)
	/workspace/adamic/internal/native/emit_statements.go:59 +0x65
github.com/system-inc/adamic/internal/native.(*emitter).moduleMain(0xb8f6b88a008)
	/workspace/adamic/internal/native/outline.go:62 +0x592
github.com/system-inc/adamic/internal/native.cProgram(0xb8f6b58c908, 0xffffffffffffffff, 0xfa0, {0x0, 0x0, 0x0?})
	/workspace/adamic/internal/native/emit.go:203 +0x1e05
github.com/system-inc/adamic/internal/native.C(...)
	/workspace/adamic/internal/native/emit.go:25
github.com/system-inc/adamic/internal/oracle.nativelyUncached(0xb8f6a91f448, 0xb8f6b58c908)
	/workspace/adamic/internal/oracle/oracle_test.go:725 +0xcb
github.com/system-inc/adamic/internal/oracle.natively(0xb8f6a91f448, 0xb8f6b58c908)
	/workspace/adamic/internal/oracle/oracle_test.go:701 +0xa9
github.com/system-inc/adamic/internal/oracle.TestRecordStorageRoadsArePinned.func1(0xb8f6a91f448)
	/workspace/adamic/internal/oracle/record_storage_roads_test.go:51 +0x699
testing.tRunner(0xb8f6a91f448, 0xb8f6a89f170)
	/workspace/adamic-tools/go/src/testing/testing.go:2193 +0xea
created by testing.(*T).Run in goroutine 21
	/workspace/adamic-tools/go/src/testing/testing.go:2258 +0x4d4
FAIL	github.com/system-inc/adamic/internal/oracle	4.933s
FAIL
```

### Mutant 3: The binding check dropped

File: `internal/lower/constraint_relation.go`. Old occurrence count: 1. Test exit: 1; wall time: 41.465 seconds; restore diff exit: 0.

Old:

```go
if l.sameRecordStorage(binding.argument, binding.constraint, map[[2]*checker.Type]bool{}) {
```

New:

```go
if true || l.sameRecordStorage(binding.argument, binding.constraint, map[[2]*checker.Type]bool{}) {
```

Red subtests: `f07c_instantiation_expression.ts`, `f07b_written_type_argument.ts`, `f07_constraint_binding.ts`. Completed PASS subtests: 8.

Additional red beyond the table: `f07c_instantiation_expression.ts`; it stopped at 6:11 with the ExpressionWithTypeArguments diagnostic rather than the expected storage stop at 6:16.

Complete command output, verbatim:

```text
=== RUN   TestRecordStorageRoadsArePinned
=== PAUSE TestRecordStorageRoadsArePinned
=== CONT  TestRecordStorageRoadsArePinned
=== RUN   TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== RUN   TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== PAUSE TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== RUN   TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== PAUSE TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== RUN   TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== RUN   TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== RUN   TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== NAME  TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
    record_storage_roads_test.go:56: stop: got "internal/oracle/testdata/record_storage_roads/f07c_instantiation_expression.ts:6:11: stage 0 can't lower an ExpressionWithTypeArguments yet", want one at internal/oracle/testdata/record_storage_roads/f07c_instantiation_expression.ts:6:16 carrying "(fixed objects and records have different storage; copy explicitly)"
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== NAME  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
    record_storage_roads_test.go:52: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f07b_written_type_argument.ts:6:18: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==10589==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x565254b0fc9e bp 0x7ffd1f9bc9f0 sp 0x7ffd1f9bc820 T0)\n==10589==The signal is caused by a READ memory access.\n==10589==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x565254b0fc9e in find /adamic-runtime/map.c:107:11\n    #1 0x565254b10205 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x565254b3b43a in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x565254b3b43a in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x565254ac5cef in adamic_function_f07b_written_type_argument_ts_fill_99b94b33c2bd819b190f4a4178e591df /adamic-build/main.c:27:2\n    #5 0x565254ac5cef in adamic_function_f07b_written_type_argument_ts_grow_d1e97e441dad0e5aef71fa57906f0477 /adamic-build/main.c:61:78\n    #6 0x565254ac5cef in adamic_initialize_f07b_written_type_argument_ts_fdd6cbc0b2fc2178db06ef87f5cfdf11 /adamic-build/main.c:81:109\n    #7 0x565254ac5a8d in adamic_initialize_f07b_written_type_argument_ts_e083c186a0773e5b88d84c7554db636a /adamic-build/main.c:94:2\n    #8 0x565254ac5a52 in main /adamic-build/main.c:102:2\n    #9 0x7f8c02c23ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x7f8c02c23d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #11 0x5652549df590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf07b_written_type_argument.ts37959404/001/program+0x81590)\n\n==10589==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x0000565254c0bd40  rbp = 0x00007ffd1f9bc9f0  rsp = 0x00007ffd1f9bc820  \n r8 = 0x00000f81803bc00c   r9 = 0x0000000000000000  r10 = 0x00000f81803bc00c  r11 = 0x0000000000000000  \nr12 = 0x00007b8c00a00080  r13 = 0x00000f7180140010  r14 = 0x3ff0000000000000  r15 = 0x00007b8c00a00080  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==10589==ABORTING\n"
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== NAME  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
    record_storage_roads_test.go:52: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f07_constraint_binding.ts:6:13: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==10595==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x5591a8262c9e bp 0x7fffdc096960 sp 0x7fffdc096780 T0)\n==10595==The signal is caused by a READ memory access.\n==10595==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x5591a8262c9e in find /adamic-runtime/map.c:107:11\n    #1 0x5591a8263205 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x5591a828e43a in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x5591a828e43a in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x5591a8218cef in adamic_function_f07_constraint_binding_ts_fill_ccbfc29388c787cffa3f43bd046cd5cb /adamic-build/main.c:27:2\n    #5 0x5591a8218cef in adamic_function_f07_constraint_binding_ts_grow_58b7c18cdc83cf6fc8b67c0e36c6be26 /adamic-build/main.c:61:78\n    #6 0x5591a8218cef in adamic_initialize_f07_constraint_binding_ts_41d827a6c17faf170ab962569b27f57f /adamic-build/main.c:81:105\n    #7 0x5591a8218a8d in adamic_initialize_f07_constraint_binding_ts_91f2fbad08336d0710142651ef402d21 /adamic-build/main.c:94:2\n    #8 0x5591a8218a52 in main /adamic-build/main.c:102:2\n    #9 0x7f41bbe41ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x7f41bbe41d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #11 0x5591a8132590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf07_constraint_binding.ts2102296956/001/program+0x81590)\n\n==10595==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x00005591a835ed40  rbp = 0x00007fffdc096960  rsp = 0x00007fffdc096780  \n r8 = 0x00000f78375fc00c   r9 = 0x0000000000000000  r10 = 0x00000f78375fc00c  r11 = 0x0000000000000000  \nr12 = 0x00007b41b9c00080  r13 = 0x00000f6837380010  r14 = 0x3ff0000000000000  r15 = 0x00007b41b9c00080  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==10595==ABORTING\n"
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== NAME  TestRecordStorageRoadsArePinned/c3_generic_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 773866d1e5b1 hit 16.58
=== NAME  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 773866d1e5b1 miss 16.58
=== NAME  TestRecordStorageRoadsArePinned/c2_logical_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 773866d1e5b1 hit 16.45
--- FAIL: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (3.93s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (3.95s)
    --- FAIL: TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts (0.17s)
    --- FAIL: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (4.26s)
    --- FAIL: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (4.27s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (0.18s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (0.14s)
    --- PASS: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.16s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (17.15s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (17.49s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (21.44s)
FAIL
gate cache: native hits=0 misses=8
gate cache: node hits=0 misses=14
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	21.460s
FAIL
```

### Mutant 4: The instantiation expression unbound

File: `internal/lower/type_argument_bindings.go`. Old occurrence count: 1. Test exit: 1; wall time: 44.0 seconds; restore diff exit: 0.

Old:

```go
return l.instantiationTypeArgumentBindings(node, visit)
```

New:

```go
return nil
```

Red subtests: `f07c_instantiation_expression.ts`. Completed PASS subtests: 10.

Complete command output, verbatim:

```text
=== RUN   TestRecordStorageRoadsArePinned
=== PAUSE TestRecordStorageRoadsArePinned
=== CONT  TestRecordStorageRoadsArePinned
=== RUN   TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== RUN   TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== PAUSE TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== RUN   TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== PAUSE TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== RUN   TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== RUN   TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== RUN   TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== NAME  TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
    record_storage_roads_test.go:56: stop: got "internal/oracle/testdata/record_storage_roads/f07c_instantiation_expression.ts:7:13: stage 0 can't lower a Point bound to T, whose constraint Bag is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want one at internal/oracle/testdata/record_storage_roads/f07c_instantiation_expression.ts:6:16 carrying "(fixed objects and records have different storage; copy explicitly)"
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== NAME  TestRecordStorageRoadsArePinned/c3_generic_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types caa9e5920e46 hit 15.26
=== NAME  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types caa9e5920e46 hit 15.27
=== NAME  TestRecordStorageRoadsArePinned/c2_logical_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types caa9e5920e46 miss 15.25
--- FAIL: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (3.79s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (3.80s)
    --- FAIL: TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts (3.81s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (3.82s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (0.18s)
    --- PASS: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.16s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (0.17s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (0.16s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (15.72s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (19.69s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (15.73s)
FAIL
gate cache: native hits=0 misses=6
gate cache: node hits=0 misses=14
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	19.712s
FAIL
```

## Step 4 — full lower suite

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Exit 0. Wall time 844 seconds (14m 04s), measured around the entire command including compilation. Go reported 827.621 seconds for the test execution. Every failing test / first error line: none. Environment failures: none reported.

Complete `/tmp/lower.log`, verbatim:

```text
ok  	github.com/system-inc/adamic/internal/lower	827.621s
```

Checkpoint history: step-2 report `b1608dd3e`; step-3 report `79202b616`. Each was committed and pushed immediately after the respective step; all report commits descend from the requested source HEAD with no force pushes.
