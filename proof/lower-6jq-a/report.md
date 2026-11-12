# Task #6jq01hc compiler proof

Target: `adf8bee664447eb990aaab04fa787fbb5bddf7a5`.

Only this report is committed. Commands run serially; compiler source restored after mutation.

## Environment

Loaded `/workspace/adamic-tools/env.sh` first. Created its missing `TMPDIR=/tmp/adamic-gate` outside the repository. Provisioned upstream `https://github.com/microsoft/TypeScript.git` at `050880ce59e30b356b686bd3144efe24f875ebc8` in `/tmp/lower-6jq-proof/TypeScript`; exported `ADAMIC_TYPESCRIPT_SOURCE` to it.

Copied stage3/api's unchanged package.json and package-lock.json to `/tmp/lower-6jq-proof/api` and ran `npm ci --ignore-scripts --no-audit --no-fund` there (exit 0). Its node_modules/.bin is first on PATH; `tsc --version` reports `Version 6.0.3`.

Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true` for npm and all validation commands/children. Heavy commands run one at a time.

## Step 1 — checkout, format, vet

Fetched `origin compiler/logical-assignment-views`, checked out the exact target detached, and updated submodules recursively. Parent chain matches `30379eb97798ece65ececd0693c2e0783fbbf6b7` then `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`.

`gofmt -l internal/lower internal/oracle`: exit 0, no output.

`go vet ./internal/lower ./internal/oracle`: exit 0, no output.

## Step 2 — unmodified pins

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestLogicalAssignmentViewsArePinned|TestRecordStorageRoadsArePinned' -count=1 -v -timeout 40m`

Exit: 0; wall time: 175 seconds.


Verbatim complete output (includes every subtest result and all failure got/want diagnostics):

```text
=== RUN   TestLogicalAssignmentViewsArePinned
=== PAUSE TestLogicalAssignmentViewsArePinned
=== RUN   TestRecordStorageRoadsArePinned
=== PAUSE TestRecordStorageRoadsArePinned
=== CONT  TestLogicalAssignmentViewsArePinned
=== CONT  TestRecordStorageRoadsArePinned
=== RUN   TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c03_or_assign_honest.ts
=== PAUSE TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c03_or_assign_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c04_and_assign_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c04_and_assign_honest.ts
=== PAUSE TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c07_optional_nullish_honest.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c07_optional_nullish_honest.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v01_assign_control.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v01_assign_control.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts
=== PAUSE TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v03_or_assign.ts
=== RUN   TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v03_or_assign.ts
=== PAUSE TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v04_and_assign.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v04_and_assign.ts
=== RUN   TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v06_readonly_nullish.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v06_readonly_nullish.ts
=== RUN   TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v07_hiding_nullish.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v07_hiding_nullish.ts
=== RUN   TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v09_hiding_assign_control.ts
=== RUN   TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v09_hiding_assign_control.ts
=== CONT  TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== RUN   TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts
=== CONT  TestLogicalAssignmentViewsArePinned/v09_hiding_assign_control.ts
=== CONT  TestLogicalAssignmentViewsArePinned/v07_hiding_nullish.ts
=== RUN   TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestLogicalAssignmentViewsArePinned/v06_readonly_nullish.ts
=== CONT  TestLogicalAssignmentViewsArePinned/v04_and_assign.ts
=== CONT  TestLogicalAssignmentViewsArePinned/v03_or_assign.ts
=== CONT  TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts
=== CONT  TestLogicalAssignmentViewsArePinned/v01_assign_control.ts
=== CONT  TestLogicalAssignmentViewsArePinned/c07_optional_nullish_honest.ts
=== CONT  TestLogicalAssignmentViewsArePinned/c04_and_assign_honest.ts
=== CONT  TestLogicalAssignmentViewsArePinned/c03_or_assign_honest.ts
=== NAME  TestLogicalAssignmentViewsArePinned/c07_optional_nullish_honest.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ae748eda92a7 miss 132.07
=== NAME  TestLogicalAssignmentViewsArePinned/c04_and_assign_honest.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ae748eda92a7 hit 132.07
=== NAME  TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ae748eda92a7 hit 132.07
=== NAME  TestLogicalAssignmentViewsArePinned/c03_or_assign_honest.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ae748eda92a7 hit 132.00
=== NAME  TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ae748eda92a7 hit 132.08
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
--- PASS: TestLogicalAssignmentViewsArePinned (0.00s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v07_hiding_nullish.ts (27.30s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts (27.31s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v09_hiding_assign_control.ts (27.31s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v06_readonly_nullish.ts (27.31s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v04_and_assign.ts (0.23s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v01_assign_control.ts (0.22s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v03_or_assign.ts (0.23s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c04_and_assign_honest.ts (132.98s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c03_or_assign_honest.ts (132.98s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts (160.53s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts (133.24s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c07_optional_nullish_honest.ts (133.02s)
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== NAME  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ae748eda92a7 hit 0.01
=== NAME  TestRecordStorageRoadsArePinned/c3_generic_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ae748eda92a7 hit 0.01
=== NAME  TestRecordStorageRoadsArePinned/c2_logical_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types ae748eda92a7 hit 0.01
--- PASS: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (0.16s)
    --- PASS: TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts (0.18s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (0.17s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (0.18s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (0.17s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (0.15s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (0.17s)
    --- PASS: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.17s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (0.59s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (0.49s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (0.48s)
PASS
gate cache: native hits=0 misses=16
gate cache: node hits=0 misses=31
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	161.388s
```

## Step 3 — logical assignments dropped mutant

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestLogicalAssignmentViewsArePinned|TestRecordStorageRoadsArePinned' -count=1 -v -timeout 40m`

Exit: 1; wall time: 25 seconds.


Asserted old text occurs exactly once, then replaced only the documented row (bare pipes). Restored original bytes afterward; `git diff --exit-code` passed.

Observed red diagnostics: `TestRecordStorageRoadsArePinned/f05b_or_assignment.ts` lowered instead of stopping and ASan crashed in map.c find; `TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts` lowered instead of stopping and native printed an extra `tom woof` while Node exited 70; `TestRecordStorageRoadsArePinned/f05c_and_assignment.ts` failed by panic `native: no conversion from 14 to 4` at native/union.go:24 during native emission. The latter panic aborted the package. The v02 fatal diagnostic appeared before the abort but its final FAIL line did not. Other unfinished subtests have no result; no PASS or FAIL is inferred for them. This is a mutant/compiler failure, not an environment failure.

Exact applied text:

```text
Old text occurs exactly once; applied only documented mutant.
OLD: (binary.OperatorToken.Kind == ast.KindEqualsToken || logicalAssignment(binary.OperatorToken.Kind))
NEW: (binary.OperatorToken.Kind == ast.KindEqualsToken || false && logicalAssignment(binary.OperatorToken.Kind))
```

Verbatim complete output (includes every subtest result and all failure got/want diagnostics):

```text
=== RUN   TestLogicalAssignmentViewsArePinned
=== PAUSE TestLogicalAssignmentViewsArePinned
=== RUN   TestRecordStorageRoadsArePinned
=== PAUSE TestRecordStorageRoadsArePinned
=== CONT  TestRecordStorageRoadsArePinned
=== CONT  TestLogicalAssignmentViewsArePinned
=== RUN   TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== RUN   TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== PAUSE TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== PAUSE TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== RUN   TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== PAUSE TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== RUN   TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c03_or_assign_honest.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c03_or_assign_honest.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c04_and_assign_honest.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c04_and_assign_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== PAUSE TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== RUN   TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== RUN   TestLogicalAssignmentViewsArePinned/c07_optional_nullish_honest.ts
=== RUN   TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/c07_optional_nullish_honest.ts
=== PAUSE TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v01_assign_control.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v01_assign_control.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v03_or_assign.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v03_or_assign.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v04_and_assign.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v04_and_assign.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v06_readonly_nullish.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v06_readonly_nullish.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v07_hiding_nullish.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v07_hiding_nullish.ts
=== RUN   TestLogicalAssignmentViewsArePinned/v09_hiding_assign_control.ts
=== PAUSE TestLogicalAssignmentViewsArePinned/v09_hiding_assign_control.ts
=== CONT  TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== NAME  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
    record_storage_roads_test.go:52: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f05b_or_assignment.ts:6:9: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==9509==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x559295d47dde bp 0x7fffaa0cf450 sp 0x7fffaa0cf280 T0)\n==9509==The signal is caused by a READ memory access.\n==9509==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x559295d47dde in find /adamic-runtime/map.c:107:11\n    #1 0x559295d48345 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x559295d7357a in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x559295d7357a in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x559295cfddab in adamic_function_f05b_or_assignment_ts_fill_a71bda3feff0a8c743ad774e3fae37a1 /adamic-build/main.c:28:2\n    #5 0x559295cfddab in adamic_initialize_f05b_or_assignment_ts_c84a5777cb056da4dc539c8b3d931623 /adamic-build/main.c:94:102\n    #6 0x559295cfda9d in adamic_initialize_f05b_or_assignment_ts_a3d71d6bdeebdd74b8754fc9242e5c6b /adamic-build/main.c:109:2\n    #7 0x559295cfda52 in main /adamic-build/main.c:117:2\n    #8 0x7ff804b7aca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #9 0x7ff804b7ad64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x559295c17590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf05b_or_assignment.ts2323709354/001/program+0x81590)\n\n==9509==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x0000559295e43d60  rbp = 0x00007fffaa0cf450  rsp = 0x00007fffaa0cf280  \n r8 = 0x00000f8f007bc00c   r9 = 0x0000000000000000  r10 = 0x00000f8f007bc00c  r11 = 0x0000000000000000  \nr12 = 0x00007bf802a00080  r13 = 0x00000f7f00540010  r14 = 0x3ff0000000000000  r15 = 0x00007c7803de0020  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==9509==ABORTING\n"
=== CONT  TestLogicalAssignmentViewsArePinned/v01_assign_control.ts
=== NAME  TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts
    logical_assignment_views_test.go:51: lowered, want a stop at internal/oracle/testdata/logical_assignment_views/v02_nullish_assign.ts:6:13: Node exit 70 "rex woof\n", sanitized native exit 0 "rex woof\ntom woof\n", stderr ""
=== CONT  TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts
--- FAIL: TestRecordStorageRoadsArePinned (4.80s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (4.40s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (0.19s)
    --- FAIL: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (4.68s)
    --- FAIL: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (0.20s)
panic: native: no conversion from 14 to 4 [recovered, repanicked]

goroutine 51 [running]:
testing.tRunner.func1.2({0x2c30f10, 0x34b029e18f00})
	/workspace/adamic-tools/go/src/testing/testing.go:2123 +0x232
testing.tRunner.func1()
	/workspace/adamic-tools/go/src/testing/testing.go:2126 +0x329
panic({0x2c30f10?, 0x34b029e18f00?})
	/workspace/adamic-tools/go/src/runtime/panic.go:859 +0x125
github.com/system-inc/adamic/internal/native.converted(0x12d68b2?, 0x4, {0x34b029d10780, 0x52})
	/workspace/adamic/internal/native/union.go:24 +0x3d9
github.com/system-inc/adamic/internal/native.logicalConverted(0x34b02b954788?, 0x2c30f10?, {0x34b029d10780?, 0x2c30f10?})
	/workspace/adamic/internal/native/taste.go:104 +0x37a
github.com/system-inc/adamic/internal/native.(*emitter).assignLogicalResult(0x34b02b954788, {0x34b029d106c0, 0x51}, 0x52?, 0x4, {0x34b029d10780?, 0x34b02ad74b78?})
	/workspace/adamic/internal/native/logical_assignment.go:150 +0x169
github.com/system-inc/adamic/internal/native.(*emitter).logicalAssignment(0x34b02b954788, {{0x2d2d140, 0x34b02d5c3290}, {0x2d2d100, 0x34b02cb84c00}, {0x0, 0x0}, {0x0, 0x0, 0x0}, ...})
	/workspace/adamic/internal/native/logical_assignment.go:130 +0x1eca
github.com/system-inc/adamic/internal/native.(*emitter).evaluate(0x34b02b954788, {0x2d2df60?, 0x34b029e7f740?})
	/workspace/adamic/internal/native/emit_expressions.go:115 +0xb165
github.com/system-inc/adamic/internal/native.(*emitter).value(0x34b02b954788, {0x2d2df60, 0x34b029e7f740})
	/workspace/adamic/internal/native/borrow.go:41 +0x249
github.com/system-inc/adamic/internal/native.(*emitter).statement(0x34b02b954788, {0x2d2d580, 0x34b02d6b3b50})
	/workspace/adamic/internal/native/emit_statements.go:196 +0x22f1
github.com/system-inc/adamic/internal/native.(*emitter).statementAt(0x34b02b954788, 0x34b02bdaa660)
	/workspace/adamic/internal/native/emit_statements.go:59 +0x65
github.com/system-inc/adamic/internal/native.(*emitter).moduleMain(0x34b02b954788)
	/workspace/adamic/internal/native/outline.go:62 +0x592
github.com/system-inc/adamic/internal/native.cProgram(0x34b029a3fb08, 0xffffffffffffffff, 0xfa0, {0x0, 0x0, 0x0?})
	/workspace/adamic/internal/native/emit.go:203 +0x1e05
github.com/system-inc/adamic/internal/native.C(...)
	/workspace/adamic/internal/native/emit.go:25
github.com/system-inc/adamic/internal/oracle.nativelyUncached(0x34b029964248, 0x34b029a3fb08)
	/workspace/adamic/internal/oracle/oracle_test.go:725 +0xcb
github.com/system-inc/adamic/internal/oracle.natively(0x34b029964248, 0x34b029a3fb08)
	/workspace/adamic/internal/oracle/oracle_test.go:701 +0xa9
github.com/system-inc/adamic/internal/oracle.TestRecordStorageRoadsArePinned.func1(0x34b029964248)
	/workspace/adamic/internal/oracle/record_storage_roads_test.go:51 +0x699
testing.tRunner(0x34b029964248, 0x34b029e16270)
	/workspace/adamic-tools/go/src/testing/testing.go:2193 +0xea
created by testing.(*T).Run in goroutine 13
	/workspace/adamic-tools/go/src/testing/testing.go:2258 +0x4d4
FAIL	github.com/system-inc/adamic/internal/oracle	4.850s
FAIL
```
