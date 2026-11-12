# Compiler proof: #mez92e7

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

