# Compiler proof: lower-inw-1 (#yqbg9vz)

HEAD: `4b868413ab4f2a7be21b41b33522279606336a7b`; base/main: `2ead00019e558b202f985ccdc5661753c1196dfd`. Measured 2026-11-11 UTC. No compiler changes retained. Only this report is committed.

## Decisive results

The full lower suite passes. The full oracle suite exits 1 with 13 environment failures, all reproduced individually at main; twelve pass after exposing the external pinned packages through a temporary link. The remaining WASI test requires an unavailable WASI_SYSROOT. The full admission delta is **void**, not a pass: four native builds still exceeded the gate’s 10-second limit on its retry. All five programs that ever timed out subsequently build alone within 10 seconds at both SHAs. Classification measured 0 admitted, 8 newly refused, and 14 changed stderr texts.


All 20 HEAD oracle pins pass. The exact stop-dropped mutant fails 14 write subtests while all six controls pass. Both i3 string-write variants exit 1 under sanitizers, print no stdout, and report an AddressSanitizer SEGV in adamic_release; Node exits 0 with `text1\n`. All six sibling programs type-check and print `text1\n` on Node; both extensions stop on every Adamic backend build.

## Environment and setup

Loaded `/workspace/adamic-tools/env.sh` first. Created its missing `TMPDIR=/tmp/adamic-gate` outside the repository. Fetched `origin compiler/in-narrow-write-1`, checked out the exact HEAD detached, and completed `git submodule update --init --recursive`: cohere `5360753e19e4417675a90e92a1823bea66b13918`, nested TypeScript fork `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Provisioned upstream `https://github.com/microsoft/TypeScript.git` in `/tmp/inw-typescript`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`; exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/inw-typescript`. Copied stage3/api package.json and lockfile to `/tmp/inw-api` and ran npm ci there (exit 0; three packages). PATH includes `/tmp/inw-api/node_modules/.bin`; tsc reports 6.0.3. Go 1.27.1 linux/amd64; Node 24.19.0; clang 20.1.8. All heavy top-level commands ran sequentially.

Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`. An external npm launcher re-exports these before exec because the admission tool filters npm’s environment through toolchain.Env. No TLS verification was disabled.

## 1. Formatting and vet

`gofmt -l internal/lower internal/oracle`: exit 0, empty output. `go vet ./internal/lower ./internal/oracle`: exit 0, wall 232 seconds, empty output.

## 2. HEAD pins

`go test ./internal/lower -run 'TestRefusalRulesRunInOrder|TestRefusalRuleOrdinalTakenOnce|TestRefusalRulesReadTheirSlotThroughSlotType|TestRefuseNodeWalksAsRefuseDoes' -count=1 -v -timeout 30m`: exit 0, wall 32 seconds.

```text
=== RUN   TestRefusalRulesRunInOrder
=== PAUSE TestRefusalRulesRunInOrder
=== RUN   TestRefusalRuleOrdinalTakenOnce
=== PAUSE TestRefusalRuleOrdinalTakenOnce
=== RUN   TestRefusalRulesReadTheirSlotThroughSlotType
=== PAUSE TestRefusalRulesReadTheirSlotThroughSlotType
=== RUN   TestRefuseNodeWalksAsRefuseDoes
=== PAUSE TestRefuseNodeWalksAsRefuseDoes
=== CONT  TestRefusalRulesRunInOrder
=== CONT  TestRefusalRulesReadTheirSlotThroughSlotType
--- PASS: TestRefusalRulesRunInOrder (0.00s)
=== CONT  TestRefusalRuleOrdinalTakenOnce
=== CONT  TestRefuseNodeWalksAsRefuseDoes
--- PASS: TestRefusalRuleOrdinalTakenOnce (0.00s)
--- PASS: TestRefuseNodeWalksAsRefuseDoes (0.00s)
--- PASS: TestRefusalRulesReadTheirSlotThroughSlotType (0.18s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.244s
```

`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestInNarrowedWritesArePinned' -count=1 -v -timeout 40m`: exit 0, wall 292 seconds.

| Subtest | Result |
|---|---|
| `w3_nullish_assignment.ts` | PASS |
| `w1_compound_assignment.a` | PASS |
| `w5_literal_key.ts` | PASS |
| `w5_literal_key.a` | PASS |
| `w4_destructuring_target.a` | PASS |
| `w4_destructuring_target.ts` | PASS |
| `i4_object_written.a` | PASS |
| `i4_object_written.ts` | PASS |
| `i3_string_written.ts` | PASS |
| `w2_increment.ts` | PASS |
| `w3_nullish_assignment.a` | PASS |
| `w2_increment.a` | PASS |
| `i2_read_typeof.ts` | PASS |
| `i0_presence_only.ts` | PASS |
| `i2_read_typeof.a` | PASS |
| `i0_presence_only.a` | PASS |
| `c1_declared_property_written.a` | PASS |
| `i3_string_written.a` | PASS |
| `w1_compound_assignment.ts` | PASS |
| `c1_declared_property_written.ts` | PASS |

No HEAD pin failures, so no failing got/want pairs.

## 3. Stop-dropped mutant

Read `internal/lower/in_narrowed_write.mutants.md`. Asserted `if !ast.IsAssignmentTarget(node) {` occurs exactly once in `internal/lower/in_narrowed_write.go`, replaced it with `if true {`, and reran exactly step 2b’s command: exit 1, wall 96 seconds. Restored the original bytes afterward; `git diff --exit-code` exited 0 with empty output.

Every red subtest and its entire failure text is included verbatim in the complete mutant log below (Go’s quoted stdout/stderr strings preserve the failure’s printed escapes). For each i3 extension: sanitized native exit 1, stdout `""`, ASan error **yes**, SEGV in adamic_release. Its complete stderr is the quoted stderr on its failure line below. Node exit 0, stdout `"text1\n"`.

```text
=== RUN   TestInNarrowedWritesArePinned
=== PAUSE TestInNarrowedWritesArePinned
=== CONT  TestInNarrowedWritesArePinned
=== RUN   TestInNarrowedWritesArePinned/c1_declared_property_written.a
=== PAUSE TestInNarrowedWritesArePinned/c1_declared_property_written.a
=== RUN   TestInNarrowedWritesArePinned/c1_declared_property_written.ts
=== PAUSE TestInNarrowedWritesArePinned/c1_declared_property_written.ts
=== RUN   TestInNarrowedWritesArePinned/i0_presence_only.a
=== PAUSE TestInNarrowedWritesArePinned/i0_presence_only.a
=== RUN   TestInNarrowedWritesArePinned/i0_presence_only.ts
=== PAUSE TestInNarrowedWritesArePinned/i0_presence_only.ts
=== RUN   TestInNarrowedWritesArePinned/i2_read_typeof.a
=== PAUSE TestInNarrowedWritesArePinned/i2_read_typeof.a
=== RUN   TestInNarrowedWritesArePinned/i2_read_typeof.ts
=== PAUSE TestInNarrowedWritesArePinned/i2_read_typeof.ts
=== RUN   TestInNarrowedWritesArePinned/i3_string_written.a
=== PAUSE TestInNarrowedWritesArePinned/i3_string_written.a
=== RUN   TestInNarrowedWritesArePinned/i3_string_written.ts
=== PAUSE TestInNarrowedWritesArePinned/i3_string_written.ts
=== RUN   TestInNarrowedWritesArePinned/i4_object_written.a
=== PAUSE TestInNarrowedWritesArePinned/i4_object_written.a
=== RUN   TestInNarrowedWritesArePinned/i4_object_written.ts
=== PAUSE TestInNarrowedWritesArePinned/i4_object_written.ts
=== RUN   TestInNarrowedWritesArePinned/w1_compound_assignment.a
=== PAUSE TestInNarrowedWritesArePinned/w1_compound_assignment.a
=== RUN   TestInNarrowedWritesArePinned/w1_compound_assignment.ts
=== PAUSE TestInNarrowedWritesArePinned/w1_compound_assignment.ts
=== RUN   TestInNarrowedWritesArePinned/w2_increment.a
=== PAUSE TestInNarrowedWritesArePinned/w2_increment.a
=== RUN   TestInNarrowedWritesArePinned/w2_increment.ts
=== PAUSE TestInNarrowedWritesArePinned/w2_increment.ts
=== RUN   TestInNarrowedWritesArePinned/w3_nullish_assignment.a
=== PAUSE TestInNarrowedWritesArePinned/w3_nullish_assignment.a
=== RUN   TestInNarrowedWritesArePinned/w3_nullish_assignment.ts
=== PAUSE TestInNarrowedWritesArePinned/w3_nullish_assignment.ts
=== RUN   TestInNarrowedWritesArePinned/w4_destructuring_target.a
=== PAUSE TestInNarrowedWritesArePinned/w4_destructuring_target.a
=== RUN   TestInNarrowedWritesArePinned/w4_destructuring_target.ts
=== PAUSE TestInNarrowedWritesArePinned/w4_destructuring_target.ts
=== RUN   TestInNarrowedWritesArePinned/w5_literal_key.a
=== PAUSE TestInNarrowedWritesArePinned/w5_literal_key.a
=== RUN   TestInNarrowedWritesArePinned/w5_literal_key.ts
=== PAUSE TestInNarrowedWritesArePinned/w5_literal_key.ts
=== CONT  TestInNarrowedWritesArePinned/c1_declared_property_written.a
=== CONT  TestInNarrowedWritesArePinned/w1_compound_assignment.a
=== CONT  TestInNarrowedWritesArePinned/w3_nullish_assignment.ts
=== CONT  TestInNarrowedWritesArePinned/i2_read_typeof.ts
=== CONT  TestInNarrowedWritesArePinned/w3_nullish_assignment.a
=== NAME  TestInNarrowedWritesArePinned/w1_compound_assignment.a
    in_narrowed_write_test.go:54: stop: got "internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.a:7:2: stage 0 can't lower a BinaryExpression with a union of differently held members and a string yet", want "internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet"
=== CONT  TestInNarrowedWritesArePinned/w5_literal_key.ts
    in_narrowed_write_test.go:54: stop: got "internal/oracle/testdata/in_narrowed_write/w5_literal_key.ts:7:2: stage 0 can't lower assigning an element of a value yet", want "internal/oracle/testdata/in_narrowed_write/w5_literal_key.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet"
=== CONT  TestInNarrowedWritesArePinned/w5_literal_key.a
=== NAME  TestInNarrowedWritesArePinned/w3_nullish_assignment.ts
    in_narrowed_write_test.go:51: lowered, want "internal/oracle/testdata/in_narrowed_write/w3_nullish_assignment.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet": Node exit 0 "6\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==9496==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x56205a64dfb0 bp 0x0000b4200101 sp 0x7fffdfa0a1e0 T0)\n==9496==The signal is caused by a READ memory access.\n==9496==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x56205a64dfb0 in adamic_retain /adamic-build/runtime/adamic.h:185:13\n    #1 0x56205a64dd5c in adamic_initialize_w3_nullish_assignment_ts_f7f34bcfa593a6604728444fc539398d /adamic-build/main.c:44:222\n    #2 0x56205a64da9d in adamic_initialize_w3_nullish_assignment_ts_dfc016de8076546d17560c46a6946bd7 /adamic-build/main.c:77:2\n    #3 0x56205a64da52 in main /adamic-build/main.c:85:2\n    #4 0x7f0fb629bca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x7f0fb629bd64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x56205a567590 in _start (/tmp/adamic-gate/TestInNarrowedWritesArePinnedw3_nullish_assignment.ts3479815194/001/program+0x80590)\n\n==9496==Register values:\nrax = 0x4014000000000000  rbx = 0x00007b8fb55e0020  rcx = 0x0802800000000001  rdx = 0x00007b8fb55e0058  \nrdi = 0x401400000000000c  rsi = 0x00007b8fb55e0050  rbp = 0x00000000b4200101  rsp = 0x00007fffdfa0a1e0  \n r8 = 0xf9f9f902f9f9f901   r9 = 0x00000ac48b4da410  r10 = 0x00000ac48b4da410  r11 = 0x00000ac48b4da410  \nr12 = 0x0000000000000000  r13 = 0x00007fffdfa0a368  r14 = 0x0000000000000000  r15 = 0x000056205a789140  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:185:13 in adamic_retain\n==9496==ABORTING\n"
=== CONT  TestInNarrowedWritesArePinned/w4_destructuring_target.ts
=== NAME  TestInNarrowedWritesArePinned/w5_literal_key.a
    in_narrowed_write_test.go:54: stop: got "internal/oracle/testdata/in_narrowed_write/w5_literal_key.a:7:2: stage 0 can't lower assigning an element of a value yet", want "internal/oracle/testdata/in_narrowed_write/w5_literal_key.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet"
=== CONT  TestInNarrowedWritesArePinned/w4_destructuring_target.a
=== NAME  TestInNarrowedWritesArePinned/w4_destructuring_target.ts
    in_narrowed_write_test.go:54: stop: got "internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.ts:7:27: stage 0 can't lower destructuring a value made in the assignment itself yet", want "internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.ts:7:12: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet"
=== CONT  TestInNarrowedWritesArePinned/i0_presence_only.ts
=== NAME  TestInNarrowedWritesArePinned/w3_nullish_assignment.a
    in_narrowed_write_test.go:51: lowered, want "internal/oracle/testdata/in_narrowed_write/w3_nullish_assignment.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet": Node exit 0 "6\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==9515==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x55e4081a2fb0 bp 0x00008ec00101 sp 0x7ffc0aea6e00 T0)\n==9515==The signal is caused by a READ memory access.\n==9515==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x55e4081a2fb0 in adamic_retain /adamic-build/runtime/adamic.h:185:13\n    #1 0x55e4081a2d5c in adamic_initialize_w3_nullish_assignment_a_fb0eeaac942f741a594d61e130501935 /adamic-build/main.c:44:220\n    #2 0x55e4081a2a9d in adamic_initialize_w3_nullish_assignment_a_5033b7f2656a0cdba702fdcbbe25f8d2 /adamic-build/main.c:77:2\n    #3 0x55e4081a2a52 in main /adamic-build/main.c:85:2\n    #4 0x7f5790da0ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x7f5790da0d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x55e4080bc590 in _start (/tmp/adamic-gate/TestInNarrowedWritesArePinnedw3_nullish_assignment.a1368248267/001/program+0x80590)\n\n==9515==Register values:\nrax = 0x4014000000000000  rbx = 0x00007bd78ffe0020  rcx = 0x0802800000000001  rdx = 0x00007bd78ffe0058  \nrdi = 0x401400000000000c  rsi = 0x00007bd78ffe0050  rbp = 0x000000008ec00101  rsp = 0x00007ffc0aea6e00  \n r8 = 0xf9f9f902f9f9f901   r9 = 0x00000abd01044e10  r10 = 0x00000abd01044e10  r11 = 0x00000abd01044e10  \nr12 = 0x0000000000000000  r13 = 0x00007ffc0aea6f88  r14 = 0x0000000000000000  r15 = 0x000055e4082de140  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:185:13 in adamic_retain\n==9515==ABORTING\n"
=== CONT  TestInNarrowedWritesArePinned/i2_read_typeof.a
=== NAME  TestInNarrowedWritesArePinned/w4_destructuring_target.a
    in_narrowed_write_test.go:54: stop: got "internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.a:7:27: stage 0 can't lower destructuring a value made in the assignment itself yet", want "internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.a:7:12: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet"
=== CONT  TestInNarrowedWritesArePinned/i0_presence_only.a
=== NAME  TestInNarrowedWritesArePinned/i0_presence_only.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types e7ba3ba7c026 hit 27.82
=== NAME  TestInNarrowedWritesArePinned/i0_presence_only.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types e7ba3ba7c026 hit 27.65
=== NAME  TestInNarrowedWritesArePinned/i2_read_typeof.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types e7ba3ba7c026 hit 27.71
=== NAME  TestInNarrowedWritesArePinned/i2_read_typeof.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types e7ba3ba7c026 hit 28.00
=== NAME  TestInNarrowedWritesArePinned/c1_declared_property_written.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types e7ba3ba7c026 miss 27.99
=== CONT  TestInNarrowedWritesArePinned/c1_declared_property_written.ts
=== CONT  TestInNarrowedWritesArePinned/i4_object_written.a
=== CONT  TestInNarrowedWritesArePinned/i4_object_written.ts
=== CONT  TestInNarrowedWritesArePinned/w2_increment.a
=== CONT  TestInNarrowedWritesArePinned/w1_compound_assignment.ts
    in_narrowed_write_test.go:54: stop: got "internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.ts:7:2: stage 0 can't lower a BinaryExpression with a union of differently held members and a string yet", want "internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet"
=== CONT  TestInNarrowedWritesArePinned/w2_increment.ts
=== NAME  TestInNarrowedWritesArePinned/w2_increment.a
    in_narrowed_write_test.go:51: lowered, want "internal/oracle/testdata/in_narrowed_write/w2_increment.a:8:3: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet": Node exit 0 "7\n", sanitized native exit 0 "7\n", stderr ""
=== CONT  TestInNarrowedWritesArePinned/i3_string_written.ts
=== NAME  TestInNarrowedWritesArePinned/i4_object_written.a
    in_narrowed_write_test.go:51: lowered, want "internal/oracle/testdata/in_narrowed_write/i4_object_written.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet": Node exit 0 "[object Object]1\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==10028==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x56304549bac2 bp 0x00007a600101 sp 0x7fffe775c540 T0)\n==10028==The signal is caused by a READ memory access.\n==10028==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x56304549bac2 in adamic_release /adamic-build/runtime/adamic.h:197:13\n    #1 0x56304549be0d in adamic_initialize_i4_object_written_a_4695595ac644999b8ee04beb42897cf8 /adamic-build/main.c:57:97\n    #2 0x56304549ba9d in adamic_initialize_i4_object_written_a_0382726d59b35e160929ddfdce1ea67c /adamic-build/main.c:73:2\n    #3 0x56304549ba52 in main /adamic-build/main.c:82:2\n    #4 0x7fe07c708ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x7fe07c708d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x5630453b5590 in _start (/tmp/adamic-gate/TestInNarrowedWritesArePinnedi4_object_written.a946053889/001/program+0x80590)\n\n==10028==Register values:\nrax = 0x0802800000000001  rbx = 0x00007c607b9e0020  rcx = 0x0000000000000000  rdx = 0x00000f8a0f73c004  \nrdi = 0x401400000000000c  rsi = 0x4014000000000000  rbp = 0x000000007a600101  rsp = 0x00007fffe775c540  \n r8 = 0x00000f8a0f73c00b   r9 = 0x0000000000000000  r10 = 0x00000f8a0f73c00b  r11 = 0x0000000000000000  \nr12 = 0x4014000000000000  r13 = 0x00000f8c0f73c00b  r14 = 0x00007c507b9e0020  r15 = 0x00007c607b9e0058  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:197:13 in adamic_release\n==10028==ABORTING\n"
=== CONT  TestInNarrowedWritesArePinned/i3_string_written.a
=== NAME  TestInNarrowedWritesArePinned/i4_object_written.ts
    in_narrowed_write_test.go:51: lowered, want "internal/oracle/testdata/in_narrowed_write/i4_object_written.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet": Node exit 0 "[object Object]1\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==10027==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x55699ae9eac2 bp 0x0000e2400101 sp 0x7fffba66caf0 T0)\n==10027==The signal is caused by a READ memory access.\n==10027==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x55699ae9eac2 in adamic_release /adamic-build/runtime/adamic.h:197:13\n    #1 0x55699ae9ee0d in adamic_initialize_i4_object_written_ts_a05935c791534eb7a35f25617861bb6d /adamic-build/main.c:57:98\n    #2 0x55699ae9ea9d in adamic_initialize_i4_object_written_ts_2acd6e858f43740d8471c27cfe1d31a2 /adamic-build/main.c:73:2\n    #3 0x55699ae9ea52 in main /adamic-build/main.c:82:2\n    #4 0x7f43e44faca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x7f43e44fad64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x55699adb8590 in _start (/tmp/adamic-gate/TestInNarrowedWritesArePinnedi4_object_written.ts1336797500/001/program+0x80590)\n\n==10027==Register values:\nrax = 0x0802800000000001  rbx = 0x00007bc3e37e0020  rcx = 0x0000000000000000  rdx = 0x00000f767c6fc004  \nrdi = 0x401400000000000c  rsi = 0x4014000000000000  rbp = 0x00000000e2400101  rsp = 0x00007fffba66caf0  \n r8 = 0x00000f767c6fc00b   r9 = 0x0000000000000000  r10 = 0x00000f767c6fc00b  r11 = 0x0000000000000000  \nr12 = 0x4014000000000000  r13 = 0x00000f787c6fc00b  r14 = 0x00007bb3e37e0020  r15 = 0x00007bc3e37e0058  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:197:13 in adamic_release\n==10027==ABORTING\n"
=== NAME  TestInNarrowedWritesArePinned/c1_declared_property_written.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types e7ba3ba7c026 hit 0.03
=== NAME  TestInNarrowedWritesArePinned/w2_increment.ts
    in_narrowed_write_test.go:51: lowered, want "internal/oracle/testdata/in_narrowed_write/w2_increment.ts:8:3: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet": Node exit 0 "7\n", sanitized native exit 0 "7\n", stderr ""
=== NAME  TestInNarrowedWritesArePinned/i3_string_written.ts
    in_narrowed_write_test.go:51: lowered, want "internal/oracle/testdata/in_narrowed_write/i3_string_written.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet": Node exit 0 "text1\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==10074==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x55bc90b6fac2 bp 0x000027600101 sp 0x7ffd7f90a410 T0)\n==10074==The signal is caused by a READ memory access.\n==10074==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x55bc90b6fac2 in adamic_release /adamic-build/runtime/adamic.h:197:13\n    #1 0x55bc90b6fdc2 in adamic_initialize_i3_string_written_ts_f670fd506740d383007990797a7be9f6 /adamic-build/main.c:51:98\n    #2 0x55bc90b6fa9d in adamic_initialize_i3_string_written_ts_3b89e1d1bfc25a24ffd791b5b4b68a7d /adamic-build/main.c:66:2\n    #3 0x55bc90b6fa52 in main /adamic-build/main.c:74:2\n    #4 0x7fe8295ceca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x7fe8295ced64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x55bc90a89590 in _start (/tmp/adamic-gate/TestInNarrowedWritesArePinnedi3_string_written.ts546808723/001/program+0x80590)\n\n==10074==Register values:\nrax = 0x0802800000000001  rbx = 0x00007c68289e0020  rcx = 0x0000000000000000  rdx = 0x00000ab7921963c4  \nrdi = 0x401400000000000c  rsi = 0x4014000000000000  rbp = 0x0000000027600101  rsp = 0x00007ffd7f90a410  \n r8 = 0xf9f9f90600000001   r9 = 0x00000ab81217e814  r10 = 0x00000ab81217e814  r11 = 0x00000ab81217e810  \nr12 = 0x00000f8d0513c00b  r13 = 0x00007ffd7f90a588  r14 = 0x00007c68289e0058  r15 = 0x4014000000000000  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:197:13 in adamic_release\n==10074==ABORTING\n"
=== NAME  TestInNarrowedWritesArePinned/i3_string_written.a
    in_narrowed_write_test.go:51: lowered, want "internal/oracle/testdata/in_narrowed_write/i3_string_written.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet": Node exit 0 "text1\n", sanitized native exit 1 "", stderr "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==10077==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x563897553ac2 bp 0x0000b9400101 sp 0x7ffdb919dd70 T0)\n==10077==The signal is caused by a READ memory access.\n==10077==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.\n    #0 0x563897553ac2 in adamic_release /adamic-build/runtime/adamic.h:197:13\n    #1 0x563897553dc2 in adamic_initialize_i3_string_written_a_d410d3bee68bac975edec536675e5448 /adamic-build/main.c:51:97\n    #2 0x563897553a9d in adamic_initialize_i3_string_written_a_99c64fecd9311c34f8eda7985e8ff44b /adamic-build/main.c:66:2\n    #3 0x563897553a52 in main /adamic-build/main.c:74:2\n    #4 0x7fd5bb517ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #5 0x7fd5bb517d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)\n    #6 0x56389746d590 in _start (/tmp/adamic-gate/TestInNarrowedWritesArePinnedi3_string_written.a653267657/001/program+0x80590)\n\n==10077==Register values:\nrax = 0x0802800000000001  rbx = 0x00007c55ba7e0020  rcx = 0x0000000000000000  rdx = 0x00000ac712ed2bc4  \nrdi = 0x401400000000000c  rsi = 0x4014000000000000  rbp = 0x00000000b9400101  rsp = 0x00007ffdb919dd70  \n r8 = 0xf9f9f90600000001   r9 = 0x00000ac792ebb014  r10 = 0x00000ac792ebb014  r11 = 0x00000ac792ebb010  \nr12 = 0x00000f8ab74fc00b  r13 = 0x00007ffdb919dee8  r14 = 0x00007c55ba7e0058  r15 = 0x4014000000000000  \nAddressSanitizer can not provide additional info.\nSUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:197:13 in adamic_release\n==10077==ABORTING\n"
--- FAIL: TestInNarrowedWritesArePinned (0.00s)
    --- FAIL: TestInNarrowedWritesArePinned/w1_compound_assignment.a (7.60s)
    --- FAIL: TestInNarrowedWritesArePinned/w5_literal_key.ts (0.27s)
    --- FAIL: TestInNarrowedWritesArePinned/w3_nullish_assignment.ts (8.15s)
    --- FAIL: TestInNarrowedWritesArePinned/w5_literal_key.a (0.33s)
    --- FAIL: TestInNarrowedWritesArePinned/w4_destructuring_target.ts (0.17s)
    --- FAIL: TestInNarrowedWritesArePinned/w3_nullish_assignment.a (8.37s)
    --- FAIL: TestInNarrowedWritesArePinned/w4_destructuring_target.a (0.27s)
    --- PASS: TestInNarrowedWritesArePinned/c1_declared_property_written.a (37.19s)
    --- PASS: TestInNarrowedWritesArePinned/i0_presence_only.a (28.74s)
    --- PASS: TestInNarrowedWritesArePinned/i2_read_typeof.ts (37.21s)
    --- PASS: TestInNarrowedWritesArePinned/i2_read_typeof.a (28.91s)
    --- PASS: TestInNarrowedWritesArePinned/i0_presence_only.ts (28.98s)
    --- FAIL: TestInNarrowedWritesArePinned/w1_compound_assignment.ts (0.35s)
    --- FAIL: TestInNarrowedWritesArePinned/w2_increment.a (0.88s)
    --- FAIL: TestInNarrowedWritesArePinned/i4_object_written.a (1.08s)
    --- FAIL: TestInNarrowedWritesArePinned/i4_object_written.ts (1.09s)
    --- PASS: TestInNarrowedWritesArePinned/c1_declared_property_written.ts (1.48s)
    --- FAIL: TestInNarrowedWritesArePinned/w2_increment.ts (1.13s)
    --- FAIL: TestInNarrowedWritesArePinned/i3_string_written.ts (0.75s)
    --- FAIL: TestInNarrowedWritesArePinned/i3_string_written.a (0.64s)
FAIL
gate cache: native hits=0 misses=20
gate cache: node hits=0 misses=26
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	38.964s
FAIL
```

## 4. Measured sibling probes (not pinned)

Built HEAD with `go build -o /tmp/adamic-cli ./cmd/adamic`: exit 0, wall 61 seconds. The supplied text was written identically to `.ts` and `.a` under `/tmp/inw-probes`. tsc runs against the `.ts` for both extension rows. Node uses `node --experimental-strip-types <name>.ts`; strip-types is available. Each native build used `-o /tmp/inw-probes/<name>.release` or `.sanitized --sanitize`. Each JavaScript command used `adamic-cli js <file>` and captured stdout as the generated-file content when successful.

The initial tsc invocation from the repository returned TS5112 because tsconfig.json was present; reran the same `tsc --strict --noEmit <name>.ts` from `/tmp/inw-probes`. All six then passed with empty stdout/stderr. This environmental invocation issue is not a program type error.

| Program | Extension | Stop text or (release, sanitized, JavaScript) against Node |
|---|---|---|
| s1_predicate | .a | /tmp/inw-probes/s1_predicate.a:5:2: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on value); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate) (Node: exit 0, `text1\n`) |
| s1_predicate | .ts | /tmp/inw-probes/s1_predicate.ts:4:1: stage 0 can't lower checked predicate overload target Plain & { extra: unknown; } yet (Node: exit 0, `text1\n`) |
| s3_annotated_record | .a | /tmp/inw-probes/s3_annotated_record.a:5:48: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as Plain & Record<"extra", unknown>, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) (Node: exit 0, `text1\n`) |
| s3_annotated_record | .ts | /tmp/inw-probes/s3_annotated_record.ts:5:48: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as Plain & Record<"extra", unknown>, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) (Node: exit 0, `text1\n`) |
| s4_unknown_field_view | .a | /tmp/inw-probes/s4_unknown_field_view.a:2:34: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as { extra: unknown; }, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) (Node: exit 0, `text1\n`) |
| s4_unknown_field_view | .ts | /tmp/inw-probes/s4_unknown_field_view.ts:2:34: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as { extra: unknown; }, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) (Node: exit 0, `text1\n`) |
| s5_asserts | .a | /tmp/inw-probes/s5_asserts.a:5:2: Adamic 0.1 refuses a type predicate whose return is not proven (branch is not a trusted parameter check); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate) (Node: exit 0, `text1\n`) |
| s5_asserts | .ts | /tmp/inw-probes/s5_asserts.ts:4:1: stage 0 can't lower checked predicate overload target Plain & { extra: unknown; } yet (Node: exit 0, `text1\n`) |
| s6_in_on_unknown | .a | /tmp/inw-probes/s6_in_on_unknown.a:4:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet (Node: exit 0, `text1\n`) |
| s6_in_on_unknown | .ts | /tmp/inw-probes/s6_in_on_unknown.ts:4:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet (Node: exit 0, `text1\n`) |
| s7_in_on_this | .a | /tmp/inw-probes/s7_in_on_this.a:5:4: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet (Node: exit 0, `text1\n`) |
| s7_in_on_this | .ts | /tmp/inw-probes/s7_in_on_this.ts:5:4: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet (Node: exit 0, `text1\n`) |

### s1_predicate.a

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s1_predicate.a:5:2: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on value); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s1_predicate.a:5:2: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on value); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s1_predicate.a:5:2: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on value); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

No run: build stopped. 

### s1_predicate.ts

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s1_predicate.ts:4:1: stage 0 can't lower checked predicate overload target Plain & { extra: unknown; } yet
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s1_predicate.ts:4:1: stage 0 can't lower checked predicate overload target Plain & { extra: unknown; } yet
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s1_predicate.ts:4:1: stage 0 can't lower checked predicate overload target Plain & { extra: unknown; } yet
```

No run: build stopped. 

### s3_annotated_record.a

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s3_annotated_record.a:5:48: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as Plain & Record<"extra", unknown>, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s3_annotated_record.a:5:48: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as Plain & Record<"extra", unknown>, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s3_annotated_record.a:5:48: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as Plain & Record<"extra", unknown>, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. 

### s3_annotated_record.ts

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s3_annotated_record.ts:5:48: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as Plain & Record<"extra", unknown>, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s3_annotated_record.ts:5:48: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as Plain & Record<"extra", unknown>, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s3_annotated_record.ts:5:48: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as Plain & Record<"extra", unknown>, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. 

### s4_unknown_field_view.a

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s4_unknown_field_view.a:2:34: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as { extra: unknown; }, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s4_unknown_field_view.a:2:34: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as { extra: unknown; }, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s4_unknown_field_view.a:2:34: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as { extra: unknown; }, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. 

### s4_unknown_field_view.ts

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s4_unknown_field_view.ts:2:34: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as { extra: unknown; }, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s4_unknown_field_view.ts:2:34: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as { extra: unknown; }, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s4_unknown_field_view.ts:2:34: Adamic 0.1 refuses a value of type { name: string; extra: number; } seen as { extra: unknown; }, which can write unknown where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

No run: build stopped. 

### s5_asserts.a

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s5_asserts.a:5:2: Adamic 0.1 refuses a type predicate whose return is not proven (branch is not a trusted parameter check); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s5_asserts.a:5:2: Adamic 0.1 refuses a type predicate whose return is not proven (branch is not a trusted parameter check); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s5_asserts.a:5:2: Adamic 0.1 refuses a type predicate whose return is not proven (branch is not a trusted parameter check); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

No run: build stopped. 

### s5_asserts.ts

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s5_asserts.ts:4:1: stage 0 can't lower checked predicate overload target Plain & { extra: unknown; } yet
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s5_asserts.ts:4:1: stage 0 can't lower checked predicate overload target Plain & { extra: unknown; } yet
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s5_asserts.ts:4:1: stage 0 can't lower checked predicate overload target Plain & { extra: unknown; } yet
```

No run: build stopped. 

### s6_in_on_unknown.a

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s6_in_on_unknown.a:4:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s6_in_on_unknown.a:4:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s6_in_on_unknown.a:4:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. 

### s6_in_on_unknown.ts

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s6_in_on_unknown.ts:4:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s6_in_on_unknown.ts:4:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s6_in_on_unknown.ts:4:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. 

### s7_in_on_this.a

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s7_in_on_this.a:5:4: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s7_in_on_this.a:5:4: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s7_in_on_this.a:5:4: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. 

### s7_in_on_this.ts

tsc: PASS (exit 0, stdout/stderr empty). Node: exit 0.

Node stdout:

```text
text1
```

Node stderr:

```text

```

release build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s7_in_on_this.ts:5:4: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. 

sanitized build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s7_in_on_this.ts:5:4: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. ASan: not run.

javascript build: exit 1. Full stderr:

```text
adamic: /tmp/inw-probes/s7_in_on_this.ts:5:4: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

No run: build stopped. 

### i3 sanitizer streams decoded from the verbatim mutant failures

#### i3_string_written.ts: sanitized native exit 1; ASan error: yes

stdout (empty):

```text

```

stderr:

```text
AddressSanitizer:DEADLYSIGNAL
=================================================================
==10074==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x55bc90b6fac2 bp 0x000027600101 sp 0x7ffd7f90a410 T0)
==10074==The signal is caused by a READ memory access.
==10074==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.
    #0 0x55bc90b6fac2 in adamic_release /adamic-build/runtime/adamic.h:197:13
    #1 0x55bc90b6fdc2 in adamic_initialize_i3_string_written_ts_f670fd506740d383007990797a7be9f6 /adamic-build/main.c:51:98
    #2 0x55bc90b6fa9d in adamic_initialize_i3_string_written_ts_3b89e1d1bfc25a24ffd791b5b4b68a7d /adamic-build/main.c:66:2
    #3 0x55bc90b6fa52 in main /adamic-build/main.c:74:2
    #4 0x7fe8295ceca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #5 0x7fe8295ced64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #6 0x55bc90a89590 in _start (/tmp/adamic-gate/TestInNarrowedWritesArePinnedi3_string_written.ts546808723/001/program+0x80590)

==10074==Register values:
rax = 0x0802800000000001  rbx = 0x00007c68289e0020  rcx = 0x0000000000000000  rdx = 0x00000ab7921963c4  
rdi = 0x401400000000000c  rsi = 0x4014000000000000  rbp = 0x0000000027600101  rsp = 0x00007ffd7f90a410  
 r8 = 0xf9f9f90600000001   r9 = 0x00000ab81217e814  r10 = 0x00000ab81217e814  r11 = 0x00000ab81217e810  
r12 = 0x00000f8d0513c00b  r13 = 0x00007ffd7f90a588  r14 = 0x00007c68289e0058  r15 = 0x4014000000000000  
AddressSanitizer can not provide additional info.
SUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:197:13 in adamic_release
==10074==ABORTING
```

#### i3_string_written.a: sanitized native exit 1; ASan error: yes

stdout (empty):

```text

```

stderr:

```text
AddressSanitizer:DEADLYSIGNAL
=================================================================
==10077==ERROR: AddressSanitizer: SEGV on unknown address (pc 0x563897553ac2 bp 0x0000b9400101 sp 0x7ffdb919dd70 T0)
==10077==The signal is caused by a READ memory access.
==10077==Hint: this fault was caused by a dereference of a high value address (see register values below).  Disassemble the provided pc to learn which register was used.
    #0 0x563897553ac2 in adamic_release /adamic-build/runtime/adamic.h:197:13
    #1 0x563897553dc2 in adamic_initialize_i3_string_written_a_d410d3bee68bac975edec536675e5448 /adamic-build/main.c:51:97
    #2 0x563897553a9d in adamic_initialize_i3_string_written_a_99c64fecd9311c34f8eda7985e8ff44b /adamic-build/main.c:66:2
    #3 0x563897553a52 in main /adamic-build/main.c:74:2
    #4 0x7fd5bb517ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #5 0x7fd5bb517d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #6 0x56389746d590 in _start (/tmp/adamic-gate/TestInNarrowedWritesArePinnedi3_string_written.a653267657/001/program+0x80590)

==10077==Register values:
rax = 0x0802800000000001  rbx = 0x00007c55ba7e0020  rcx = 0x0000000000000000  rdx = 0x00000ac712ed2bc4  
rdi = 0x401400000000000c  rsi = 0x4014000000000000  rbp = 0x00000000b9400101  rsp = 0x00007ffdb919dd70  
 r8 = 0xf9f9f90600000001   r9 = 0x00000ac792ebb014  r10 = 0x00000ac792ebb014  r11 = 0x00000ac792ebb010  
r12 = 0x00000f8ab74fc00b  r13 = 0x00007ffdb919dee8  r14 = 0x00007c55ba7e0058  r15 = 0x4014000000000000  
AddressSanitizer can not provide additional info.
SUMMARY: AddressSanitizer: SEGV /adamic-build/runtime/adamic.h:197:13 in adamic_release
==10077==ABORTING
```

## 5. Full lower suite

```text
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

exit 0, wall 1514 seconds. No failing tests; no main reruns needed. Complete log:

```text
ok  	github.com/system-inc/adamic/internal/lower	1510.030s
```

## 6. Full oracle suite and main comparison

```text
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```

exit 1, wall 4153 seconds. Thirteen failing tests, all marked ENV below. Each was rerun individually, sequentially, in `/tmp/inw-main` detached at the specified main SHA with the same pinned submodules and same missing module-link/WASI conditions. Each main command was `go test ./internal/oracle -run '^<test>$' -count=1 -v -timeout 120m`.

| Failing test | First error line at HEAD (verbatim) | Environment cause | Also fails at main? | Main exit / wall seconds |
|---|---|---|---|---|
| TestStdoutColumnsRefusesWASI | library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT | ENV: WASI_SYSROOT unavailable | Yes, same cause | 1 / 256.490 |
| TestAssignmentProofReturnDefinedMutant | assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 9.698 |
| TestAssignmentProofWiderTargetMutant | assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 4.413 |
| TestAssignmentProofWiderTarget | assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 3.270 |
| TestAssignmentProofCompoundMutant | assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 3.928 |
| TestAssignmentProofChainedMutant | assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 2.879 |
| TestAssignmentProofCompound | assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 3.372 |
| TestAssignmentProofChained | assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 2.967 |
| TestAssignmentProofReturnDefined | assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 3.091 |
| TestAssignmentProofIfNarrowingMutant | assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 2.525 |
| TestAssignmentProofIfNarrowing | assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 3.390 |
| TestAssignmentProofScannerKeywordMutant | assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 2.560 |
| TestAssignmentProofScannerKeyword | assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | ENV: hard-coded stage3/api/node_modules absent; TypeScript installed externally | Yes, same cause | 1 / 2.583 |

The assignment-proof helper explicitly sets NODE_PATH to `<repository>/stage3/api/node_modules`; putting tsc on PATH alone does not satisfy that module lookup. After preserving the original failure and main comparison, temporarily linked that path at HEAD to `/tmp/inw-api/node_modules` (packages still installed outside the repository). Targeted follow-up:

```text
go test ./internal/oracle -run '^TestAssignmentProof(ScannerKeyword|IfNarrowing|ReturnDefined|Chained|Compound|WiderTarget)(Mutant)?$' -count=1 -v -timeout 40m
```

exit 0, wall 28 seconds. All twelve pass. The link is removed before the report commit. The full oracle suite was not rerun, and is not claimed green. No WASI sysroot was fabricated or refusal check bypassed.

Original full oracle log, verbatim:

```text
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestStdoutColumnsRefusesWASI (0.89s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.19s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestAssignmentProofWiderTargetMutant (0.08s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.12s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.09s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.16s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.17s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.05s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.04s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.06s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.05s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.05s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.06s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
FAIL
gate cache: native hits=3052 misses=6520
gate cache: node hits=1745 misses=6541
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	4144.377s
FAIL
```

Targeted corrected-environment log, verbatim:

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
=== CONT  TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofIfNarrowingMutant
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.85
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.86
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.86
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.92
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.98
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:192: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "12 12 1\n14 14 2\n", mutant "2 2 1\n2 2 2\n"
--- PASS: TestAssignmentProofCompoundMutant (11.67s)
=== CONT  TestAssignmentProofCompound
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:162: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "4 1\n-1 2\n", mutant "4 1\n7 2\n"
--- PASS: TestAssignmentProofIfNarrowingMutant (11.70s)
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:695: gate cache native hit
=== CONT  TestAssignmentProofIfNarrowing
--- PASS: TestAssignmentProofScannerKeyword (11.70s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofChained
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofChained (11.71s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:202: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128 1\n", mutant "128 0 1\n"
--- PASS: TestAssignmentProofWiderTargetMutant (11.74s)
=== CONT  TestAssignmentProofChainedMutant
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.02
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.02
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.02
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.01
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.02
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofCompound (3.48s)
=== CONT  TestAssignmentProofReturnDefinedMutant
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:182: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 value1 1\n", mutant "value1 value1 undefined 1\n"
--- PASS: TestAssignmentProofChainedMutant (3.61s)
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:152: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128\n83 83\n80 80\n80 80\n80 80\n", mutant "80 80\n80 80\n80 80\n80 80\n80 80\n"
--- PASS: TestAssignmentProofScannerKeywordMutant (3.69s)
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofWiderTarget (3.84s)
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofIfNarrowing (3.86s)
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
    assignment_proofs_test.go:172: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.01
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
    assignment_proofs_test.go:167: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types cd783efd3c21 hit 0.01
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:172: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 1\nvalue2 value2 2\n", mutant "value1 undefined 1\nvalue2 undefined 2\n"
--- PASS: TestAssignmentProofReturnDefinedMutant (1.73s)
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:681: gate cache native miss
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofReturnDefined (1.90s)
PASS
gate cache: native hits=12 misses=24
gate cache: node hits=0 misses=24
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	17.353s
```

## 7. Admission delta

```text
python3 cloud/admission-corpus/manifest.py --sha 4b868413ab4f2a7be21b41b33522279606336a7b > /tmp/inw-proof/manifest.json
go run ./cmd/adamic-admission-delta --base 2ead00019e558b202f985ccdc5661753c1196dfd --head 4b868413ab4f2a7be21b41b33522279606336a7b --manifest /tmp/inw-proof/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json
```

Ran alone, with no other tests, builds, or inspection commands during the gate. HEAD’s parent was verified to be the requested main SHA. The gate used a separate HEAD worktree and copied the externally installed pinned packages into it. No budget, corpus filter, or shard was used.

`go run` exit 1, wall 1636 seconds. The tool itself returned exit **3**, printed by go run as `exit status 3`. Verdict: **void**. Complete corpus: true.

| Measure | Result |
|---|---|
| workers | 2 |
| diff_count | 20 |
| admitted | 0 |
| timed_out_runs | 5 |
| retried_runs | 5 |
| still_timed_out_runs | 4 |
| accepted_witnesses | 48 |
| ruled_runtime_stops | 52 |
| sampling_size | 115 |
| omitted | 0 |
| newly refused | 8 |
| classified programs | 5253 |
| classification stderr differences | 14 |

Manifest corpora: 2 witnesses, 4,305 fixtures, 105 gaps, 20 review, 0 fuzz, 821 stage3. Classification counts: {"accepted-by-both": 3288, "newly-refused": 8, "refused-by-both": 1957}. No classification timeout occurred. The timeouts were native builds during runtime checks.

Void reasons, verbatim:

```text
internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts: native compile timeout
```

Gate stderr, verbatim:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 5 runs timed out, 5 retried alone, 4 timed out again
changed programs: 20
admission delta is void: internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts: native compile timeout
exit status 3
```

### Every classification stderr difference

#### internal/oracle/testdata/in_narrowed_write/i3_string_written.a — newly-refused

main/base stderr (empty):

```text

```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/i3_string_written.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/i3_string_written.ts — newly-refused

main/base stderr (empty):

```text

```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/i3_string_written.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/i4_object_written.a — newly-refused

main/base stderr (empty):

```text

```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/i4_object_written.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/i4_object_written.ts — newly-refused

main/base stderr (empty):

```text

```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/i4_object_written.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.a — refused-by-both

main/base stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.a:7:2: stage 0 can't lower a BinaryExpression with a union of differently held members and a string yet
```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.ts — refused-by-both

main/base stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.ts:7:2: stage 0 can't lower a BinaryExpression with a union of differently held members and a string yet
```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w1_compound_assignment.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w2_increment.a — newly-refused

main/base stderr (empty):

```text

```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w2_increment.a:8:3: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w2_increment.ts — newly-refused

main/base stderr (empty):

```text

```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w2_increment.ts:8:3: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w3_nullish_assignment.a — newly-refused

main/base stderr (empty):

```text

```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w3_nullish_assignment.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w3_nullish_assignment.ts — newly-refused

main/base stderr (empty):

```text

```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w3_nullish_assignment.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.a — refused-by-both

main/base stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.a:7:27: stage 0 can't lower destructuring a value made in the assignment itself yet
```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.a:7:12: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.ts — refused-by-both

main/base stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.ts:7:27: stage 0 can't lower destructuring a value made in the assignment itself yet
```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w4_destructuring_target.ts:7:12: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w5_literal_key.a — refused-by-both

main/base stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w5_literal_key.a:7:2: stage 0 can't lower assigning an element of a value yet
```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w5_literal_key.a:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

#### internal/oracle/testdata/in_narrowed_write/w5_literal_key.ts — refused-by-both

main/base stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w5_literal_key.ts:7:2: stage 0 can't lower assigning an element of a value yet
```

HEAD stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-3351459020/head/internal/oracle/testdata/in_narrowed_write/w5_literal_key.ts:7:2: stage 0 can't lower a write to 'extra', which only an `in` test added to the type yet
```

### All timeout reruns alone at both SHAs

Reran all five programs marked timed_out, including nested_token.ts, which had already recovered in the gate’s retry. Built main CLI from the exact main worktree (`go build -o /tmp/inw-main-cli ./cmd/adamic`: exit 0, wall 23 seconds). Used the existing restored HEAD CLI for HEAD. Every rerun used the original 10-second native-build limit, one process group at a time, with no concurrent commands. A 120-second follow-up was prepared but never needed. Source Git blob hashes matched the gate’s input at both SHAs. These later successful runs do not retroactively change the original gate verdict from void.

| Program | SHA | Build exit | Build wall seconds | Native exit | Native stdout (JSON string) | Native stderr (JSON string) |
|---|---|---|---|---|---|---|
| nested_boolean.ts | main | 0 | 2.425 | 70 | "other\n" | "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n" |
| nested_boolean.ts | head | 0 | 2.651 | 70 | "other\n" | "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n" |
| nested_boolean_repaired.ts | main | 0 | 2.172 | 0 | "other\n4 true\n" | "" |
| nested_boolean_repaired.ts | head | 0 | 2.455 | 0 | "other\n4 true\n" | "" |
| nested_literal.ts | main | 0 | 2.313 | 70 | "other\n" | "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n" |
| nested_literal.ts | head | 0 | 2.340 | 70 | "other\n" | "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n" |
| nested_literal_repaired.ts | main | 0 | 2.507 | 0 | "other\nnode:fs 11\n" | "" |
| nested_literal_repaired.ts | head | 0 | 2.224 | 0 | "other\nnode:fs 11\n" | "" |
| nested_token.ts | main | 0 | 2.540 | 70 | "other\n" | "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n" |
| nested_token.ts | head | 0 | 2.235 | 70 | "other\n" | "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n" |

Every timeout-rerun build had empty stdout and stderr. The three negative witnesses exit 70 with the same named check and stdout prefix at both SHAs; repaired programs exit 0 with the same stdout as source Node. These observations establish that the recorded timeouts did not persist in the serial reruns; the original gate remains void rather than a demonstrated full admission pass.

### Evidence hashes

SHA-256 of the external full admission JSON: `99b03941a038dd9401e75cc91ea6a8e8ee5cb5fe6fefeccf4b7d9f6839538006`. SHA-256 of external timeout rerun JSON: `73c5716572014a1a677f67f8b9e3a88f5cf5cdb6417d9e55bb4c5927923f08dc`. All requested outcomes and changed stderr texts are embedded above; raw files remain under `/tmp/inw-proof`.

## Repository state and delivery

The mutant was restored and the pre-report diff check passed. Temporary stage3/api/node_modules link removed. Report branch: `proof/lower-inw-1-report`, created from the exact checked-out HEAD. The only committed path is `proof/lower-inw-1/report.md`; no compiler change or other file is committed. Push destination is only `origin refs/heads/proof/lower-inw-1-report`.
