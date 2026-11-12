# Compiler proof: #mez92e7

## Final verdict

- Baseline pins: PASS, all 10 subtests; command exit 0, wall 349 seconds.
- All 3 mutants caught. Shorthand: exactly f02 red. Binding: exactly f07 and f07b red. Logical assignment: f05c red with a native emitter panic; the process aborted, so the other expected logical reds were not observed. Full verbatim evidence follows below.
- Full lower suite retry: PASS, exit 0, wall 1193 seconds (19m53s); no failing tests or environment failures. First attempt was interrupted by the session/environment after about 24 minutes and yielded no verdict.
- IR suite: PASS, exit 0, wall 129 seconds; no failures.
- Formatting and vet: PASS.

The compiler change passes the requested baseline and full lower/IR tests, and each mutant makes the pins fail. The logical mutant is caught by a compiler panic rather than the table-promised sanitized native witnesses; this proof does not claim those unobserved witnesses.

Report checkpoints pushed to `origin/proof/lower-mez-a-report`: `a786b5cc9` after step 2; `2a406c288` after step 3; `5e8bf6e22` for the interrupted step 4 attempt; `b1973d89e` after the passing step 4 retry. This final report adds step 4b and the final verdict. Every commit on this branch changes only this report. No other branch was pushed, no force push was used, and all compiler mutations were restored.

Base: `11db2faf88c95c059a45ec1dde98937d2ad93ebb`

Parent: `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`

Commands run serially. No compiler changes committed.

## Step 1: PASS

Fetched origin compiler/record-storage-roads, checked out the exact HEAD detached, initialized recursive submodules. gofmt -l internal/lower internal/oracle printed nothing. go vet ./internal/lower ./internal/oracle exited 0.

## Provisioning: PASS

Loaded /workspace/adamic-tools/env.sh first. Created /tmp/adamic-gate outside the repository (mode 1777). Upstream https://github.com/microsoft/TypeScript.git checked out at 050880ce59e30b356b686bd3144efe24f875ebc8 under /tmp/adamic-typescript-upstream and exported ADAMIC_TYPESCRIPT_SOURCE. Installed stage3/api package.json and lockfile with npm ci under /tmp/adamic-api. External node_modules/.bin on PATH; tsc reports Version 6.0.3. Exported NODE_EXTRA_CA_CERTS and npm_config_cafile as /etc/ssl/certs/ca-certificates.crt and npm_config_strict_ssl=true for npm and test descendants.

## Step 2: PASS, all 10 subtests

Command: ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRecordStorageRoadsArePinned' -count=1 -v -timeout 40m

Complete verbatim output follows. No failing subtests.

```text
exit=0 wall_seconds=349

```

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
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== NAME  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 14c3de9db7b4 hit 268.75
=== NAME  TestRecordStorageRoadsArePinned/c2_logical_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 14c3de9db7b4 hit 268.75
=== NAME  TestRecordStorageRoadsArePinned/c3_generic_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 14c3de9db7b4 miss 268.75
--- PASS: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (49.60s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (49.68s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (49.68s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (49.69s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (0.33s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (0.35s)
    --- PASS: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.48s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (319.91s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (270.23s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (269.93s)
PASS
gate cache: native hits=0 misses=6
gate cache: node hits=0 misses=13
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	319.972s

```

## Step 3: all three mutants caught

Each table edit was applied separately, using bare pipes after decoding markdown escaping. The old text count was asserted exactly one. Original bytes were restored after each command, and git diff --exit-code returned 0 each time. Each run used the exact step 2 command with ADAMIC_GATE_UNCACHED=1.

Mutant 1: f02_shorthand.ts FAIL, all other 9 subtests PASS.

Mutant 2: f05c_and_assignment.ts FAIL with panic: native: no conversion from 14 to 4. The process aborted before any other subtest verdict. This is a killed mutant, but not the table-promised three completed native fault witnesses. No got/want comparison was printed for the panic; the complete panic and stack are preserved below.

Mutant 3: f07_constraint_binding.ts and f07b_written_type_argument.ts FAIL, all other 8 subtests PASS.

Complete verbatim outputs, including every observed red and its got/want or panic, follow.

```text
Mutant 1: The shorthand dropped
File: records.go
Old text count asserted: 1
Old: } else if node.Kind == ast.KindShorthandPropertyAssignment {
New: } else if false && node.Kind == ast.KindShorthandPropertyAssignment {
Expected red: f02_shorthand.ts
exit=1 wall_seconds=78.145
Restored original bytes; git diff --exit-code: 0

```

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
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== NAME  TestRecordStorageRoadsArePinned/f02_shorthand.ts
    record_storage_roads_test.go:51: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f02_shorthand.ts:6:32: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==9321==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x562b1d78ae1e bp 0x7ffd88517360 sp 0x7ffd88517180 T0)\n==9321==The signal is caused by a READ memory access.\n==9321==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x562b1d78ae1e in find /adamic-runtime/map.c:107:11\n    #1 0x562b1d78b385 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x562b1d7b65ba in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x562b1d7b65ba in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x562b1d740ddb in adamic_function_f02_shorthand_ts_fill_8ba76ae5db5816cc91ce0e622cbcb2e2 /adamic-build/main.c:34:2\n    #5 0x562b1d740ddb in adamic_initialize_f02_shorthand_ts_534c7785d2f286e4771872e3b209d0b2 /adamic-build/main.c:86:97\n    #6 0x562b1d740aad in adamic_initialize_f02_shorthand_ts_59fad384c6e4960bf1f1e644ebd4a028 /adamic-build/main.c:101:2\n    #7 0x562b1d740a52 in main /adamic-build/main.c:110:2\n    #8 0x7f1d9df11ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #9 0x7f1d9df11d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x562b1d65a590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf02_shorthand.ts1350477492/001/program+0x81590)\n\n==9321==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x0000562b1d886da0  rbp = 0x00007ffd88517360  rsp = 0x00007ffd88517180  \n r8 = 0x00000f71b3a3c00b   r9 = 0x0000000000000000  r10 = 0x00000f71b3a3c00b  r11 = 0x0000000000000000  \nr12 = 0x00007b1d9be00080  r13 = 0x00007b1d9be00080  r14 = 0x3ff0000000000000  r15 = 0x00000f63b37c0010  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==9321==ABORTING\n"
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== NAME  TestRecordStorageRoadsArePinned/c3_generic_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 3137a313abaa hit 28.60
=== NAME  TestRecordStorageRoadsArePinned/c2_logical_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 3137a313abaa miss 28.53
=== NAME  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 3137a313abaa hit 28.60
--- FAIL: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (6.28s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (6.31s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (0.25s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (0.69s)
    --- FAIL: TestRecordStorageRoadsArePinned/f02_shorthand.ts (7.21s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (0.28s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (0.29s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (29.87s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (36.41s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (36.48s)
FAIL
gate cache: native hits=0 misses=7
gate cache: node hits=0 misses=13
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	36.508s
FAIL

```

```text
Mutant 2: The logical assignments dropped
File: records.go
Old text count asserted: 1
Old: (viewSite(node) || logicalAssignmentRight(node))
New: (viewSite(node) || false && logicalAssignmentRight(node))
Expected red: f05_logical_assignment.ts, f05b_or_assignment.ts, f05c_and_assignment.ts
exit=1 wall_seconds=38.469
Restored original bytes; git diff --exit-code: 0

```

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
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
--- FAIL: TestRecordStorageRoadsArePinned (6.34s)
    --- FAIL: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (6.34s)
panic: native: no conversion from 14 to 4 [recovered, repanicked]

goroutine 28 [running]:
testing.tRunner.func1.2({0x2c2f680, 0x15fdbbb80e90})
	/workspace/adamic-tools/go/src/testing/testing.go:2123 +0x232
testing.tRunner.func1()
	/workspace/adamic-tools/go/src/testing/testing.go:2126 +0x329
panic({0x2c2f680?, 0x15fdbbb80e90?})
	/workspace/adamic-tools/go/src/runtime/panic.go:859 +0x125
github.com/system-inc/adamic/internal/native.converted(0x12d58b2?, 0x4, {0x15fdbc080780, 0x52})
	/workspace/adamic/internal/native/union.go:24 +0x3d9
github.com/system-inc/adamic/internal/native.logicalConverted(0x15fdbc084008?, 0x2c2f680?, {0x15fdbc080780?, 0x2c2f680?})
	/workspace/adamic/internal/native/taste.go:104 +0x37a
github.com/system-inc/adamic/internal/native.(*emitter).assignLogicalResult(0x15fdbc084008, {0x15fdbc0806c0, 0x51}, 0x52?, 0x4, {0x15fdbc080780?, 0x15fdbc0b6b78?})
	/workspace/adamic/internal/native/logical_assignment.go:150 +0x169
github.com/system-inc/adamic/internal/native.(*emitter).logicalAssignment(0x15fdbc084008, {{0x2d2b8b0, 0x15fdbec60150}, {0x2d2b870, 0x15fdbec495e0}, {0x0, 0x0}, {0x0, 0x0, 0x0}, ...})
	/workspace/adamic/internal/native/logical_assignment.go:130 +0x1eca
github.com/system-inc/adamic/internal/native.(*emitter).evaluate(0x15fdbc084008, {0x2d2c6d0?, 0x15fdbec5e0c0?})
	/workspace/adamic/internal/native/emit_expressions.go:115 +0xb165
github.com/system-inc/adamic/internal/native.(*emitter).value(0x15fdbc084008, {0x2d2c6d0, 0x15fdbec5e0c0})
	/workspace/adamic/internal/native/borrow.go:41 +0x249
github.com/system-inc/adamic/internal/native.(*emitter).statement(0x15fdbc084008, {0x2d2bcf0, 0x15fdbec2f9f0})
	/workspace/adamic/internal/native/emit_statements.go:196 +0x22f1
github.com/system-inc/adamic/internal/native.(*emitter).statementAt(0x15fdbc084008, 0x15fdbec13ee0)
	/workspace/adamic/internal/native/emit_statements.go:59 +0x65
github.com/system-inc/adamic/internal/native.(*emitter).moduleMain(0x15fdbc084008)
	/workspace/adamic/internal/native/outline.go:62 +0x592
github.com/system-inc/adamic/internal/native.cProgram(0x15fdbb77d208, 0xffffffffffffffff, 0xfa0, {0x0, 0x0, 0x0?})
	/workspace/adamic/internal/native/emit.go:203 +0x1e05
github.com/system-inc/adamic/internal/native.C(...)
	/workspace/adamic/internal/native/emit.go:25
github.com/system-inc/adamic/internal/oracle.nativelyUncached(0x15fdbb745448, 0x15fdbb77d208)
	/workspace/adamic/internal/oracle/oracle_test.go:725 +0xcb
github.com/system-inc/adamic/internal/oracle.natively(0x15fdbb745448, 0x15fdbb77d208)
	/workspace/adamic/internal/oracle/oracle_test.go:701 +0xa9
github.com/system-inc/adamic/internal/oracle.TestRecordStorageRoadsArePinned.func1(0x15fdbb745448)
	/workspace/adamic/internal/oracle/record_storage_roads_test.go:50 +0x699
testing.tRunner(0x15fdbb745448, 0x15fdbb72e3f0)
	/workspace/adamic-tools/go/src/testing/testing.go:2193 +0xea
created by testing.(*T).Run in goroutine 20
	/workspace/adamic-tools/go/src/testing/testing.go:2258 +0x4d4
FAIL	github.com/system-inc/adamic/internal/oracle	6.375s
FAIL

```

```text
Mutant 3: The binding check dropped
File: constraint_relation.go
Old text count asserted: 1
Old: if l.sameRecordStorage(binding.argument, binding.constraint, map[[2]*checker.Type]bool{}) {
New: if true || l.sameRecordStorage(binding.argument, binding.constraint, map[[2]*checker.Type]bool{}) {
Expected red: f07_constraint_binding.ts, f07b_written_type_argument.ts
exit=1 wall_seconds=64.800
Restored original bytes; git diff --exit-code: 0

```

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
=== CONT  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
=== CONT  TestRecordStorageRoadsArePinned/f05_logical_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
=== CONT  TestRecordStorageRoadsArePinned/f05c_and_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f05b_or_assignment.ts
=== CONT  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
=== CONT  TestRecordStorageRoadsArePinned/f01_assignment_control.ts
=== CONT  TestRecordStorageRoadsArePinned/f02_shorthand.ts
=== CONT  TestRecordStorageRoadsArePinned/c2_logical_record.ts
=== CONT  TestRecordStorageRoadsArePinned/c3_generic_record.ts
=== NAME  TestRecordStorageRoadsArePinned/f07_constraint_binding.ts
    record_storage_roads_test.go:51: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f07_constraint_binding.ts:6:13: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==10578==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x5574e52fcc9e bp 0x7fff30afb3c0 sp 0x7fff30afb1e0 T0)\n==10578==The signal is caused by a READ memory access.\n==10578==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x5574e52fcc9e in find /adamic-runtime/map.c:107:11\n    #1 0x5574e52fd205 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x5574e532843a in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x5574e532843a in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x5574e52b2cef in adamic_function_f07_constraint_binding_ts_fill_ccbfc29388c787cffa3f43bd046cd5cb /adamic-build/main.c:27:2\n    #5 0x5574e52b2cef in adamic_function_f07_constraint_binding_ts_grow_58b7c18cdc83cf6fc8b67c0e36c6be26 /adamic-build/main.c:61:78\n    #6 0x5574e52b2cef in adamic_initialize_f07_constraint_binding_ts_41d827a6c17faf170ab962569b27f57f /adamic-build/main.c:81:105\n    #7 0x5574e52b2a8d in adamic_initialize_f07_constraint_binding_ts_91f2fbad08336d0710142651ef402d21 /adamic-build/main.c:94:2\n    #8 0x5574e52b2a52 in main /adamic-build/main.c:102:2\n    #9 0x7f6e91344ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x7f6e91344d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #11 0x5574e51cc590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf07_constraint_binding.ts2236630060/001/program+0x81590)\n\n==10578==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x00005574e53f8d40  rbp = 0x00007fff30afb3c0  rsp = 0x00007fff30afb1e0  \n r8 = 0x00000f7dd207c00c   r9 = 0x0000000000000000  r10 = 0x00000f7dd207c00c  r11 = 0x0000000000000000  \nr12 = 0x00007b6e8f000080  r13 = 0x00000f6dd1e00010  r14 = 0x3ff0000000000000  r15 = 0x00007b6e8f000080  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==10578==ABORTING\n"
=== NAME  TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts
    record_storage_roads_test.go:51: lowered, want a stop at internal/oracle/testdata/record_storage_roads/f07b_written_type_argument.ts:6:18: Node exit 0 "1 3\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==10599==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x564281b31c9e bp 0x7fff6a46c850 sp 0x7fff6a46c680 T0)\n==10599==The signal is caused by a READ memory access.\n==10599==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x564281b31c9e in find /adamic-runtime/map.c:107:11\n    #1 0x564281b32205 in adamic_map_set /adamic-runtime/map.c:174:17\n    #2 0x564281b5d43a in adamic_record_define /adamic-runtime/record.c:130:2\n    #3 0x564281b5d43a in adamic_record_set /adamic-runtime/record.c:215:2\n    #4 0x564281ae7cef in adamic_function_f07b_written_type_argument_ts_fill_99b94b33c2bd819b190f4a4178e591df /adamic-build/main.c:27:2\n    #5 0x564281ae7cef in adamic_function_f07b_written_type_argument_ts_grow_d1e97e441dad0e5aef71fa57906f0477 /adamic-build/main.c:61:78\n    #6 0x564281ae7cef in adamic_initialize_f07b_written_type_argument_ts_fdd6cbc0b2fc2178db06ef87f5cfdf11 /adamic-build/main.c:81:109\n    #7 0x564281ae7a8d in adamic_initialize_f07b_written_type_argument_ts_e083c186a0773e5b88d84c7554db636a /adamic-build/main.c:94:2\n    #8 0x564281ae7a52 in main /adamic-build/main.c:102:2\n    #9 0x7fe504407ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #10 0x7fe504407d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #11 0x564281a01590 in _start (/tmp/adamic-gate/TestRecordStorageRoadsArePinnedf07b_written_type_argument.ts1194897040/001/program+0x81590)\n\n==10599==Register values:\nrax = 0x07fe000000000006  rbx = 0x3ff0000000000000  rcx = 0xfffffffffffffff8  rdx = 0x4008000000000000  \nrdi = 0x3ff0000000000030  rsi = 0x0000564281c2dd40  rbp = 0x00007fff6a46c850  rsp = 0x00007fff6a46c680  \n r8 = 0x00000f8ca06bc00c   r9 = 0x0000000000000000  r10 = 0x00000f8ca06bc00c  r11 = 0x0000000000000000  \nr12 = 0x00007be502200080  r13 = 0x00000f7ca0440010  r14 = 0x3ff0000000000000  r15 = 0x00007be502200080  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-runtime/map.c:107:11 in find\n==10599==ABORTING\n"
=== NAME  TestRecordStorageRoadsArePinned/c1_shorthand_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types c25be56de4f2 hit 26.36
=== NAME  TestRecordStorageRoadsArePinned/c2_logical_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types c25be56de4f2 miss 26.36
=== NAME  TestRecordStorageRoadsArePinned/c3_generic_record.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types c25be56de4f2 hit 26.17
--- FAIL: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (6.13s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (6.14s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (6.15s)
    --- PASS: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.39s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (0.65s)
    --- FAIL: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (6.89s)
    --- FAIL: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (1.00s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (27.42s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (33.98s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (27.19s)
FAIL
gate cache: native hits=0 misses=8
gate cache: node hits=0 misses=13
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	34.014s
FAIL

```

## Step 4: first attempt interrupted by environment/session interruption

Command: go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1

The session wait was aborted and the environment was subsequently marked starting. On readiness, lower.test was defunct after about 24 minutes, the parent go test and shell were absent, /tmp/lower.log was empty, and the post-command status file had not been written. No exit code or completed test verdict was captured. This attempt is INTERRUPTED (environment/session), not PASS or FAIL. A fresh serial attempt follows.

## Step 4 retry: PASS

Command: go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1

The retry completed successfully. Failing tests: none. Environment failures in the completed run: none. The first attempt interruption above is separate from this successful verdict.

```text
start_utc=2026-11-12T21:39:05Z

```

```text
exit=0 wall_seconds=1193

```

```text
ok  	github.com/system-inc/adamic/internal/lower	1184.867s

```

## Step 4b: PASS

Command: go test ./internal/ir -count=1 -timeout 30m

Failing tests: none. Complete output follows.

```text
exit=0 wall_seconds=129

```

```text
ok  	github.com/system-inc/adamic/internal/ir	96.130s

```

