# Compiler proof: #mez92e7

## Decisive results

- Step 2: PASS, all 11 pin subtests; exit 0; wall time 155 seconds.
- Step 1: PASS, gofmt emitted nothing and go vet exited 0.
- Steps 3, 4, 4b: pending.

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
