# Host-member-write proof report

Date: 2026-11-12 (UTC)

All requested checks passed. All 28 host-member-write programs passed; there were no failures or skips. The IR package test passed.

## Repository and setup

- Repository: `https://github.com/system-inc/adamic.git`.
- Ran `git fetch origin compiler/host-member-write-5`, checked out detached HEAD `1d29e9fc46493d2cdbfc3371bbb8ca0fabdd9f34`, and ran `git submodule update --init --recursive`; all exited 0.
- Parent: `730e7c26e590af0c57c545ba0cb905b09e4bee54`, matching the supplied main commit.
- Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` before repository work and in every check environment. Created its configured `TMPDIR=/tmp/adamic-gate` outside the repository before Go checks.
- Provisioned upstream `https://github.com/microsoft/TypeScript.git` outside the repository at `/tmp/adamic-hmw5-proof/TypeScript`, checked out detached commit `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that path.
- Copied `stage3/api/package.json` and `stage3/api/package-lock.json` to `/tmp/adamic-hmw5-proof/api` and ran `npm ci --prefix /tmp/adamic-hmw5-proof/api` successfully (3 packages). Pins: TypeScript `6.0.3`, `@types/node` `25.3.3`.
- Set `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt` and `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`; prepended `/tmp/adamic-hmw5-proof/api/node_modules/.bin` to `PATH`. Verified `tsc --version`: `Version 6.0.3`.
- Tools: Go `go1.27.1 linux/amd64`, Node `v24.19.0`, npm `11.9.0`.
- Working tree was clean before checks and after all checks. No compiler code was changed.

## Step 1: formatting and vet

`gofmt -l internal/oracle internal/lower`: exit **0**, empty output (as required).

`go vet ./internal/oracle`: exit **0**, empty output; no diagnostics.

## Step 2: uncached host-member-write proof

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHostMemberWriteAgreesOrStops$' -count=1 -v -timeout 60m
```

Exit **0**; package **PASS**, elapsed **189.095s**. All **28/28** program subtests passed. No FAIL messages or skips occurred.

| Program subtest | Result | Duration |
| --- | --- | --- |
| `a1_devirtualized.a` | PASS | 187.62s |
| `a1_devirtualized.ts` | PASS | 1.07s |
| `a2_function_expression.a` | PASS | 0.92s |
| `a2_function_expression.ts` | PASS | 0.90s |
| `a3_cycle_local.a` | PASS | 187.45s |
| `a3_cycle_local.ts` | PASS | 187.43s |
| `a4_cycle_rewritten.a` | PASS | 187.47s |
| `a4_cycle_rewritten.ts` | PASS | 187.62s |
| `f1_frozen.a` | PASS | 1.09s |
| `f1_frozen.ts` | PASS | 0.82s |
| `s1_narrow_wide.a` | PASS | 0.04s |
| `s1_narrow_wide.ts` | PASS | 0.08s |
| `s2_narrow_wide_object.a` | PASS | 0.08s |
| `s2_narrow_wide_object.ts` | PASS | 0.09s |
| `s3_class_instance.a` | PASS | 0.08s |
| `s3_class_instance.ts` | PASS | 0.05s |
| `s4_exported_parameter.a` | PASS | 0.10s |
| `s4_exported_parameter.ts` | PASS | 0.06s |
| `s5_arrow_parameter.a` | PASS | 0.03s |
| `s5_arrow_parameter.ts` | PASS | 0.05s |
| `s6_accessor_pair.a` | PASS | 0.04s |
| `s6_accessor_pair.ts` | PASS | 0.13s |
| `s6b_getter_only.a` | PASS | 0.03s |
| `s6b_getter_only.ts` | PASS | 0.09s |
| `s7_spread.a` | PASS | 0.06s |
| `s7_spread.ts` | PASS | 0.10s |
| `s8_earlier_refusal.a` | PASS | 0.05s |
| `s8_earlier_refusal.ts` | PASS | 0.03s |

Cache statistics: native hits=0 misses=20; Node hits=0 misses=20; probe hits=0 misses=0.

### Complete oracle output

```text
=== RUN   TestHostMemberWriteAgreesOrStops
=== PAUSE TestHostMemberWriteAgreesOrStops
=== CONT  TestHostMemberWriteAgreesOrStops
=== RUN   TestHostMemberWriteAgreesOrStops/a1_devirtualized.a
=== PAUSE TestHostMemberWriteAgreesOrStops/a1_devirtualized.a
=== RUN   TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts
=== RUN   TestHostMemberWriteAgreesOrStops/a2_function_expression.a
=== PAUSE TestHostMemberWriteAgreesOrStops/a2_function_expression.a
=== RUN   TestHostMemberWriteAgreesOrStops/a2_function_expression.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/a2_function_expression.ts
=== RUN   TestHostMemberWriteAgreesOrStops/a3_cycle_local.a
=== PAUSE TestHostMemberWriteAgreesOrStops/a3_cycle_local.a
=== RUN   TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts
=== RUN   TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a
=== PAUSE TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a
=== RUN   TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts
=== RUN   TestHostMemberWriteAgreesOrStops/f1_frozen.a
=== PAUSE TestHostMemberWriteAgreesOrStops/f1_frozen.a
=== RUN   TestHostMemberWriteAgreesOrStops/f1_frozen.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/f1_frozen.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s1_narrow_wide.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s1_narrow_wide.a
=== RUN   TestHostMemberWriteAgreesOrStops/s1_narrow_wide.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s1_narrow_wide.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.a
=== RUN   TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s3_class_instance.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s3_class_instance.a
=== RUN   TestHostMemberWriteAgreesOrStops/s3_class_instance.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s3_class_instance.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s4_exported_parameter.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s4_exported_parameter.a
=== RUN   TestHostMemberWriteAgreesOrStops/s4_exported_parameter.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s4_exported_parameter.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.a
=== RUN   TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s6_accessor_pair.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s6_accessor_pair.a
=== RUN   TestHostMemberWriteAgreesOrStops/s6_accessor_pair.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s6_accessor_pair.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s6b_getter_only.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s6b_getter_only.a
=== RUN   TestHostMemberWriteAgreesOrStops/s6b_getter_only.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s6b_getter_only.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s7_spread.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s7_spread.a
=== RUN   TestHostMemberWriteAgreesOrStops/s7_spread.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s7_spread.ts
=== RUN   TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.a
=== PAUSE TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.a
=== RUN   TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.ts
=== PAUSE TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.ts
=== CONT  TestHostMemberWriteAgreesOrStops/a1_devirtualized.a
=== CONT  TestHostMemberWriteAgreesOrStops/s3_class_instance.a
=== CONT  TestHostMemberWriteAgreesOrStops/s6_accessor_pair.ts
=== CONT  TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s1_narrow_wide.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.a
=== CONT  TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a
=== CONT  TestHostMemberWriteAgreesOrStops/s6_accessor_pair.a
=== CONT  TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts
=== CONT  TestHostMemberWriteAgreesOrStops/a3_cycle_local.a
=== NAME  TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 miss 133.73
=== NAME  TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 133.73
=== NAME  TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 133.73
=== NAME  TestHostMemberWriteAgreesOrStops/a3_cycle_local.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 133.73
=== NAME  TestHostMemberWriteAgreesOrStops/a1_devirtualized.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 133.73
=== CONT  TestHostMemberWriteAgreesOrStops/a2_function_expression.ts
=== CONT  TestHostMemberWriteAgreesOrStops/f1_frozen.ts
=== CONT  TestHostMemberWriteAgreesOrStops/a2_function_expression.a
=== CONT  TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.a
=== CONT  TestHostMemberWriteAgreesOrStops/s7_spread.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s7_spread.a
=== CONT  TestHostMemberWriteAgreesOrStops/s6b_getter_only.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s6b_getter_only.a
=== CONT  TestHostMemberWriteAgreesOrStops/s4_exported_parameter.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.a
=== CONT  TestHostMemberWriteAgreesOrStops/f1_frozen.a
=== CONT  TestHostMemberWriteAgreesOrStops/s4_exported_parameter.a
=== CONT  TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts
=== NAME  TestHostMemberWriteAgreesOrStops/f1_frozen.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 0.01
=== NAME  TestHostMemberWriteAgreesOrStops/a2_function_expression.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 0.01
=== NAME  TestHostMemberWriteAgreesOrStops/a2_function_expression.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 0.01
=== CONT  TestHostMemberWriteAgreesOrStops/s3_class_instance.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s1_narrow_wide.a
=== NAME  TestHostMemberWriteAgreesOrStops/f1_frozen.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 0.01
=== NAME  TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types cb9b0017f660 hit 0.01
--- PASS: TestHostMemberWriteAgreesOrStops (0.00s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s3_class_instance.a (0.08s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s1_narrow_wide.ts (0.08s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s6_accessor_pair.ts (0.13s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.a (0.08s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.ts (0.09s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s6_accessor_pair.a (0.04s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts (187.43s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a (187.47s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts (187.62s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a1_devirtualized.a (187.62s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a3_cycle_local.a (187.45s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.ts (0.05s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.ts (0.03s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.a (0.05s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s7_spread.ts (0.10s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s7_spread.a (0.06s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s6b_getter_only.a (0.03s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s6b_getter_only.ts (0.09s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s4_exported_parameter.ts (0.06s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.a (0.03s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s4_exported_parameter.a (0.10s)
    --- PASS: TestHostMemberWriteAgreesOrStops/f1_frozen.ts (0.82s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s3_class_instance.ts (0.05s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s1_narrow_wide.a (0.04s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a2_function_expression.ts (0.90s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a2_function_expression.a (0.92s)
    --- PASS: TestHostMemberWriteAgreesOrStops/f1_frozen.a (1.09s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts (1.07s)
PASS
gate cache: native hits=0 misses=20
gate cache: node hits=0 misses=20
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	189.095s
```

## Step 3: IR package

```sh
go test ./internal/ir -count=1 -timeout 30m
```

Exit **0**; no failures.

```text
ok  	github.com/system-inc/adamic/internal/ir	72.040s
```

## Report delivery

The report branch is `proof/lower-hmw-5-report`, created from the tested HEAD `1d29e9fc46493d2cdbfc3371bbb8ca0fabdd9f34`. Its sole intended change is this file, `proof/lower-hmw-5/report.md`. Only that branch is designated for push to origin.
