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
