# Agreement check: #fj0vqtf

Run date: 2026-11-12 (UTC).

Checked revision: `78e926c404ffe8864683ef0cf87cd1bccfc914d5` (detached), from `compiler/host-member-value-1`; parent/main `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`.

## Decisive results

- Formatting and vet: PASS; no formatting output or compiler errors.
- Value pins: exit 1, 2 PASS and 20 FAIL. All 20 failures compare measured got text against `"unmeasured"`; these are expected measurement failures, not 20 demonstrated semantic disagreements.
- Wave 1 write pins: exit 0, all 28 PASS.
- Lower subset: exit 0. Complete output below.
- Mutant (b): exit 1; 6 changed subtest results/diagnostics; 0 PASS/FAIL transitions. Source restored; `git diff --exit-code` exit 0.
- Mutant (c): exit 1; 0 changed subtest results/diagnostics; 0 PASS/FAIL transitions. Source restored; `git diff --exit-code` exit 0.

The value pins distinguish mutant (b) through changed got text, while mutant (c) is undetected by this suite at this revision. The `unmeasured` pins record exact stops; replacing them or claiming agreement for refused programs is outside this check.

## Setup and provenance

Loaded `/workspace/adamic-tools/env.sh` before repository work. The initial worktree was clean. Ran `git fetch origin compiler/host-member-value-1`, detached at the requested HEAD, and completed `git submodule update --init --recursive`. No preexisting changes were overwritten. Only one heavy top-level command ran at a time; test-internal parallelism was left as specified by the requested commands. No compiler fixes or pin edits were made.

Created the external Go work directory `/tmp/adamic-gate` before running Go. Provisioned upstream TypeScript from `https://github.com/microsoft/TypeScript.git` at exactly `050880ce59e30b356b686bd3144efe24f875ebc8` in `/tmp/fj0vqtf/typescript`, exported as `ADAMIC_TYPESCRIPT_SOURCE`. Copied `stage3/api/package.json` and `package-lock.json` to `/tmp/fj0vqtf/api` and ran `npm ci --prefix /tmp/fj0vqtf/api` (3 packages). Prepended its `node_modules/.bin` to PATH and exported its node_modules as NODE_PATH. `tsc --version`: `Version 6.0.3`; Go: `go1.27.1 linux/amd64`; Node: `v24.19.0`. Packages and upstream provisioning were outside the repository.

For npm and every test process, including admission-gate children, exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`.

## Step 1: formatting and vet

`gofmt -l internal/lower internal/oracle`: exit 0; stdout empty. No files required `gofmt -d`. `go vet ./internal/lower ./internal/oracle`: exit 0; stdout/stderr empty. No compiler errors.

## Step 2: value pins

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHostMemberValueAgreesOrStops$' -count=1 -v -timeout 40m`. Exit 1.

| Subtest | Outcome |
| --- | --- |
| `TestHostMemberValueAgreesOrStops/h2_plain_call.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/h2_plain_call.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/h3_reordered_layout.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a` | PASS |
| `TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts` | PASS |
| `TestHostMemberValueAgreesOrStops/r1_holder_run.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/r1_holder_run.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/r2_alias_field.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/r2_alias_field.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/r3_apply.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/r3_apply.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/s1_class_instance.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/s1_class_instance.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/s2_unknown_reach.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/v1_tsc_shape.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/v2_same_layout_this.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts` | FAIL |
| `TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a` | FAIL |
| `TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts` | FAIL |

Complete unmodified output, including every failing subtest’s whole got and want verbatim:

```text
=== RUN   TestHostMemberValueAgreesOrStops
=== PAUSE TestHostMemberValueAgreesOrStops
=== CONT  TestHostMemberValueAgreesOrStops
=== RUN   TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== PAUSE TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== RUN   TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== PAUSE TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== RUN   TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== PAUSE TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== RUN   TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== PAUSE TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== RUN   TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== PAUSE TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== RUN   TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== PAUSE TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== RUN   TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== PAUSE TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== RUN   TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== RUN   TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== PAUSE TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== RUN   TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== RUN   TestHostMemberValueAgreesOrStops/r3_apply.a
=== PAUSE TestHostMemberValueAgreesOrStops/r3_apply.a
=== RUN   TestHostMemberValueAgreesOrStops/r3_apply.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r3_apply.ts
=== RUN   TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== PAUSE TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== RUN   TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== PAUSE TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== RUN   TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== PAUSE TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== RUN   TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== PAUSE TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== RUN   TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== PAUSE TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== RUN   TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== RUN   TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== PAUSE TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== RUN   TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== RUN   TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== PAUSE TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== RUN   TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== CONT  TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== CONT  TestHostMemberValueAgreesOrStops/r3_apply.ts
=== CONT  TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== CONT  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== CONT  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== NAME  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2b_function_expression_this.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== NAME  TestHostMemberValueAgreesOrStops/r3_apply.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r3_apply.ts:17:20: stage 0 can't lower host.read, a method read as a value, passed as an argument yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== NAME  TestHostMemberValueAgreesOrStops/h2_plain_call.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h2_plain_call.a:15:14: stage 0 can't lower a call of original with no object, which holds host.read, a method whose function takes its object as this yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== NAME  TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v1_tsc_shape.ts:13:22: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== NAME  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2b_function_expression_this.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== NAME  TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s2_unknown_reach.a:8:19: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== NAME  TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v1_tsc_shape.a:13:22: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== NAME  TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s2_unknown_reach.ts:8:19: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== NAME  TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2_same_layout_this.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r3_apply.a
=== NAME  TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2_same_layout_this.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== NAME  TestHostMemberValueAgreesOrStops/r3_apply.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r3_apply.a:17:20: stage 0 can't lower host.read, a method read as a value, passed as an argument yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== NAME  TestHostMemberValueAgreesOrStops/r2_alias_field.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r2_alias_field.ts:15:20: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== NAME  TestHostMemberValueAgreesOrStops/r1_holder_run.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r1_holder_run.a:16:31: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== NAME  TestHostMemberValueAgreesOrStops/s1_class_instance.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s1_class_instance.a:11:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/s1_class_instance.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s1_class_instance.ts:11:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== CONT  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== NAME  TestHostMemberValueAgreesOrStops/r1_holder_run.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r1_holder_run.ts:16:31: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== NAME  TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h3_reordered_layout.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== NAME  TestHostMemberValueAgreesOrStops/r2_alias_field.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r2_alias_field.a:15:20: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h3_reordered_layout.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/h2_plain_call.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h2_plain_call.ts:15:14: stage 0 can't lower a call of original with no object, which holds host.read, a method whose function takes its object as this yet", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 216.20
=== NAME  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 miss 216.20
--- FAIL: TestHostMemberValueAgreesOrStops (0.00s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a (0.20s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r3_apply.ts (0.21s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h2_plain_call.a (0.22s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts (0.22s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts (0.25s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s2_unknown_reach.a (0.20s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v1_tsc_shape.a (0.21s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts (0.20s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts (0.27s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2_same_layout_this.a (0.26s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r3_apply.a (0.09s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r2_alias_field.ts (0.13s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r1_holder_run.a (0.16s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s1_class_instance.a (0.19s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s1_class_instance.ts (0.20s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r1_holder_run.ts (0.11s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts (0.12s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r2_alias_field.a (0.18s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h3_reordered_layout.a (0.08s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h2_plain_call.ts (0.07s)
    --- PASS: TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a (257.73s)
    --- PASS: TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts (257.72s)
FAIL
gate cache: native hits=0 misses=4
gate cache: node hits=0 misses=4
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	258.377s
FAIL
```

## Step 2b: wave 1 write pins

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHostMemberWriteAgreesOrStops$' -count=1 -v -timeout 40m`. Exit 0. No failures.

| Subtest | Outcome |
| --- | --- |
| `TestHostMemberWriteAgreesOrStops/a1_devirtualized.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/a2_function_expression.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/a2_function_expression.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/a3_cycle_local.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/f1_frozen.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/f1_frozen.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s1_narrow_wide.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s1_narrow_wide.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s3_class_instance.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s3_class_instance.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s4_exported_parameter.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s4_exported_parameter.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s6_accessor_pair.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s6_accessor_pair.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s6b_getter_only.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s6b_getter_only.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s7_spread.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s7_spread.ts` | PASS |
| `TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.a` | PASS |
| `TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.ts` | PASS |

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
=== CONT  TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts
=== CONT  TestHostMemberWriteAgreesOrStops/a3_cycle_local.a
=== CONT  TestHostMemberWriteAgreesOrStops/a2_function_expression.a
=== CONT  TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a
=== NAME  TestHostMemberWriteAgreesOrStops/a2_function_expression.a
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.75
=== NAME  TestHostMemberWriteAgreesOrStops/a1_devirtualized.a
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.75
=== NAME  TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.80
=== NAME  TestHostMemberWriteAgreesOrStops/a3_cycle_local.a
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.86
=== NAME  TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.89
=== CONT  TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s6_accessor_pair.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.a
=== CONT  TestHostMemberWriteAgreesOrStops/s7_spread.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s7_spread.a
=== CONT  TestHostMemberWriteAgreesOrStops/s6b_getter_only.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s6b_getter_only.a
=== CONT  TestHostMemberWriteAgreesOrStops/a2_function_expression.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.a
=== CONT  TestHostMemberWriteAgreesOrStops/s6_accessor_pair.a
=== CONT  TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s1_narrow_wide.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.a
=== CONT  TestHostMemberWriteAgreesOrStops/f1_frozen.ts
=== CONT  TestHostMemberWriteAgreesOrStops/s1_narrow_wide.a
=== CONT  TestHostMemberWriteAgreesOrStops/f1_frozen.a
=== CONT  TestHostMemberWriteAgreesOrStops/s4_exported_parameter.a
=== CONT  TestHostMemberWriteAgreesOrStops/s4_exported_parameter.ts
=== CONT  TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts
=== NAME  TestHostMemberWriteAgreesOrStops/a2_function_expression.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.02
=== NAME  TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.02
=== NAME  TestHostMemberWriteAgreesOrStops/f1_frozen.a
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.03
=== CONT  TestHostMemberWriteAgreesOrStops/s3_class_instance.ts
=== NAME  TestHostMemberWriteAgreesOrStops/f1_frozen.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.04
=== NAME  TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types 38e40c6c0428 hit 0.01
--- PASS: TestHostMemberWriteAgreesOrStops (0.00s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s3_class_instance.a (0.21s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a3_cycle_local.a (45.21s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a2_function_expression.a (45.25s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a1_devirtualized.a (45.31s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s6_accessor_pair.ts (0.10s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.ts (45.37s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a4_cycle_rewritten.a (45.18s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.ts (0.08s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s8_earlier_refusal.a (0.10s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s7_spread.ts (0.09s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s7_spread.a (0.10s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s6b_getter_only.ts (0.10s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.a (0.04s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s6b_getter_only.a (0.08s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s6_accessor_pair.a (0.06s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s5_arrow_parameter.ts (0.09s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s1_narrow_wide.ts (0.15s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.ts (0.13s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s2_narrow_wide_object.a (0.09s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s1_narrow_wide.a (0.10s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s4_exported_parameter.a (0.06s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s4_exported_parameter.ts (0.10s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a2_function_expression.ts (1.39s)
    --- PASS: TestHostMemberWriteAgreesOrStops/f1_frozen.a (1.23s)
    --- PASS: TestHostMemberWriteAgreesOrStops/s3_class_instance.ts (0.11s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a3_cycle_local.ts (1.75s)
    --- PASS: TestHostMemberWriteAgreesOrStops/f1_frozen.ts (1.42s)
    --- PASS: TestHostMemberWriteAgreesOrStops/a1_devirtualized.ts (1.38s)
PASS
gate cache: native hits=0 misses=20
gate cache: node hits=0 misses=20
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	47.354s
```

## Step 2c: lower subset

Command: `go test ./internal/lower -count=1 -timeout 60m -run 'Method|HostMember|Unbound|Refus|Presence|ScanningReference|CycleFinderEquality|Bind|Optional|Detached' > /tmp/lower-subset.log 2>&1`. Exit 0. The following is the whole output, preserving every failing test’s full failure text if present:

```text
ok  	github.com/system-inc/adamic/internal/lower	1171.620s
```

## Step 3: mutants

Applied exactly the edits documented in `internal/lower/host-member-value.mutants.md`, one at a time. Before each edit, asserted that the documented old text occurs exactly once in `internal/lower/host_member_value.go` and that the file matches the saved original. Repeated step 2’s exact command for each. After each run, restored the original bytes and ran `git diff --exit-code`. Comparison includes changed got diagnostics even when both runs FAIL against `unmeasured`; helper-build timing/cache chatter is excluded.

### Mutant (b)

```text
Mutant (b): exact old-text occurrences = 1
```

Exit: 1. Restoration diff exit: 0; diff output empty.

Every changed subtest, with complete original and mutant diagnostic text:

```text
TestHostMemberValueAgreesOrStops/r1_holder_run.a FAIL -> FAIL
BASELINE:
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r1_holder_run.a:16:31: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
MUTANT:
    host_member_value_test.go:58: lowered, want "unmeasured"
TestHostMemberValueAgreesOrStops/r1_holder_run.ts FAIL -> FAIL
BASELINE:
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r1_holder_run.ts:16:31: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
MUTANT:
    host_member_value_test.go:58: lowered, want "unmeasured"
TestHostMemberValueAgreesOrStops/r2_alias_field.a FAIL -> FAIL
BASELINE:
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r2_alias_field.a:15:20: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
MUTANT:
    host_member_value_test.go:58: lowered, want "unmeasured"
TestHostMemberValueAgreesOrStops/r2_alias_field.ts FAIL -> FAIL
BASELINE:
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r2_alias_field.ts:15:20: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
MUTANT:
    host_member_value_test.go:58: lowered, want "unmeasured"
TestHostMemberValueAgreesOrStops/r3_apply.a FAIL -> FAIL
BASELINE:
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r3_apply.a:17:20: stage 0 can't lower host.read, a method read as a value, passed as an argument yet", want "unmeasured"
MUTANT:
    host_member_value_test.go:58: lowered, want "unmeasured"
TestHostMemberValueAgreesOrStops/r3_apply.ts FAIL -> FAIL
BASELINE:
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r3_apply.ts:17:20: stage 0 can't lower host.read, a method read as a value, passed as an argument yet", want "unmeasured"
MUTANT:
    host_member_value_test.go:58: lowered, want "unmeasured"
CHANGED=6; STATUS_CHANGED=0
```

Complete mutant test output:

```text
=== RUN   TestHostMemberValueAgreesOrStops
=== PAUSE TestHostMemberValueAgreesOrStops
=== CONT  TestHostMemberValueAgreesOrStops
=== RUN   TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== PAUSE TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== RUN   TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== PAUSE TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== RUN   TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== PAUSE TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== RUN   TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== PAUSE TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== RUN   TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== PAUSE TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== RUN   TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== PAUSE TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== RUN   TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== PAUSE TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== RUN   TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== RUN   TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== PAUSE TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== RUN   TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== RUN   TestHostMemberValueAgreesOrStops/r3_apply.a
=== PAUSE TestHostMemberValueAgreesOrStops/r3_apply.a
=== RUN   TestHostMemberValueAgreesOrStops/r3_apply.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r3_apply.ts
=== RUN   TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== PAUSE TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== RUN   TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== PAUSE TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== RUN   TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== PAUSE TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== RUN   TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== PAUSE TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== RUN   TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== PAUSE TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== RUN   TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== RUN   TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== PAUSE TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== RUN   TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== RUN   TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== PAUSE TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== RUN   TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== CONT  TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== CONT  TestHostMemberValueAgreesOrStops/r3_apply.ts
=== CONT  TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== CONT  TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== CONT  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== NAME  TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v1_tsc_shape.a:13:22: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r3_apply.a
=== NAME  TestHostMemberValueAgreesOrStops/r3_apply.ts
    host_member_value_test.go:58: lowered, want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== NAME  TestHostMemberValueAgreesOrStops/h2_plain_call.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h2_plain_call.a:15:14: stage 0 can't lower a call of original with no object, which holds host.read, a method whose function takes its object as this yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== NAME  TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v1_tsc_shape.ts:13:22: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== NAME  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2b_function_expression_this.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== NAME  TestHostMemberValueAgreesOrStops/r1_holder_run.ts
    host_member_value_test.go:58: lowered, want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== NAME  TestHostMemberValueAgreesOrStops/r1_holder_run.a
    host_member_value_test.go:58: lowered, want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== NAME  TestHostMemberValueAgreesOrStops/r2_alias_field.ts
    host_member_value_test.go:58: lowered, want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== NAME  TestHostMemberValueAgreesOrStops/r3_apply.a
    host_member_value_test.go:58: lowered, want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== NAME  TestHostMemberValueAgreesOrStops/r2_alias_field.a
    host_member_value_test.go:58: lowered, want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== NAME  TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h3_reordered_layout.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== NAME  TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h3_reordered_layout.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== NAME  TestHostMemberValueAgreesOrStops/h2_plain_call.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h2_plain_call.ts:15:14: stage 0 can't lower a call of original with no object, which holds host.read, a method whose function takes its object as this yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== NAME  TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s2_unknown_reach.a:8:19: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== NAME  TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2_same_layout_this.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== NAME  TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s2_unknown_reach.ts:8:19: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== NAME  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2b_function_expression_this.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== NAME  TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2_same_layout_this.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/s1_class_instance.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s1_class_instance.ts:11:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/s1_class_instance.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s1_class_instance.a:11:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types c36f1817b970 miss 27.10
=== NAME  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types c36f1817b970 hit 27.10
--- FAIL: TestHostMemberValueAgreesOrStops (0.00s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v1_tsc_shape.a (0.29s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r3_apply.ts (0.30s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h2_plain_call.a (0.31s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts (0.33s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a (0.37s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r1_holder_run.ts (0.29s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r1_holder_run.a (0.26s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r2_alias_field.ts (0.33s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r3_apply.a (0.35s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r2_alias_field.a (0.34s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h3_reordered_layout.a (0.14s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts (0.15s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h2_plain_call.ts (0.11s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s2_unknown_reach.a (0.06s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts (0.07s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts (0.07s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts (0.06s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2_same_layout_this.a (0.07s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s1_class_instance.ts (0.07s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s1_class_instance.a (0.06s)
    --- PASS: TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts (34.26s)
    --- PASS: TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a (34.26s)
FAIL
gate cache: native hits=0 misses=4
gate cache: node hits=0 misses=4
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	34.955s
FAIL
```

### Mutant (c)

```text
Mutant (c): exact old-text occurrences = 1
```

Exit: 1. Restoration diff exit: 0; diff output empty.

Every changed subtest, with complete original and mutant diagnostic text:

```text
CHANGED=0; STATUS_CHANGED=0
```

Complete mutant test output:

```text
=== RUN   TestHostMemberValueAgreesOrStops
=== PAUSE TestHostMemberValueAgreesOrStops
=== CONT  TestHostMemberValueAgreesOrStops
=== RUN   TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== PAUSE TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== RUN   TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== PAUSE TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== RUN   TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== PAUSE TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== RUN   TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== PAUSE TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== RUN   TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== PAUSE TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== RUN   TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== PAUSE TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== RUN   TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== PAUSE TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== RUN   TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== RUN   TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== PAUSE TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== RUN   TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== RUN   TestHostMemberValueAgreesOrStops/r3_apply.a
=== PAUSE TestHostMemberValueAgreesOrStops/r3_apply.a
=== RUN   TestHostMemberValueAgreesOrStops/r3_apply.ts
=== PAUSE TestHostMemberValueAgreesOrStops/r3_apply.ts
=== RUN   TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== PAUSE TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== RUN   TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== PAUSE TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== RUN   TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== PAUSE TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== RUN   TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== PAUSE TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== RUN   TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== PAUSE TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== RUN   TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== RUN   TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== PAUSE TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== RUN   TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== RUN   TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== PAUSE TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== RUN   TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== PAUSE TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== CONT  TestHostMemberValueAgreesOrStops/h2_plain_call.a
=== CONT  TestHostMemberValueAgreesOrStops/r3_apply.ts
=== CONT  TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
=== CONT  TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
=== CONT  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
=== NAME  TestHostMemberValueAgreesOrStops/r3_apply.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r3_apply.ts:17:20: stage 0 can't lower host.read, a method read as a value, passed as an argument yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
=== NAME  TestHostMemberValueAgreesOrStops/s2_unknown_reach.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s2_unknown_reach.a:8:19: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
=== NAME  TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v1_tsc_shape.ts:13:22: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s1_class_instance.ts
=== NAME  TestHostMemberValueAgreesOrStops/h2_plain_call.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h2_plain_call.a:15:14: stage 0 can't lower a call of original with no object, which holds host.read, a method whose function takes its object as this yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s1_class_instance.a
=== NAME  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2b_function_expression_this.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r1_holder_run.a
=== NAME  TestHostMemberValueAgreesOrStops/s1_class_instance.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s1_class_instance.ts:11:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r3_apply.a
=== NAME  TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2_same_layout_this.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r2_alias_field.ts
=== NAME  TestHostMemberValueAgreesOrStops/v2_same_layout_this.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2_same_layout_this.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r2_alias_field.a
=== NAME  TestHostMemberValueAgreesOrStops/s1_class_instance.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s1_class_instance.a:11:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/r1_holder_run.ts
=== NAME  TestHostMemberValueAgreesOrStops/r1_holder_run.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r1_holder_run.a:16:31: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
=== NAME  TestHostMemberValueAgreesOrStops/r3_apply.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r3_apply.a:17:20: stage 0 can't lower host.read, a method read as a value, passed as an argument yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
=== NAME  TestHostMemberValueAgreesOrStops/r2_alias_field.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r2_alias_field.ts:15:20: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
=== NAME  TestHostMemberValueAgreesOrStops/r2_alias_field.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r2_alias_field.a:15:20: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
=== NAME  TestHostMemberValueAgreesOrStops/r1_holder_run.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r1_holder_run.ts:16:31: stage 0 can't lower host.read, a method read as a value, stored in an object's field yet", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
=== NAME  TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h3_reordered_layout.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
=== NAME  TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2b_function_expression_this.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
=== NAME  TestHostMemberValueAgreesOrStops/v1_tsc_shape.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v1_tsc_shape.a:13:22: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== CONT  TestHostMemberValueAgreesOrStops/h2_plain_call.ts
=== NAME  TestHostMemberValueAgreesOrStops/h3_reordered_layout.a
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h3_reordered_layout.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s2_unknown_reach.ts:8:19: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/h2_plain_call.ts
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h2_plain_call.ts:15:14: stage 0 can't lower a call of original with no object, which holds host.read, a method whose function takes its object as this yet", want "unmeasured"
=== NAME  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types cee9e367c0b9 hit 21.22
=== NAME  TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts
    oracle_test.go:653: build go_build_./oracle/json_types.go_json-types cee9e367c0b9 miss 21.22
--- FAIL: TestHostMemberValueAgreesOrStops (0.00s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r3_apply.ts (0.17s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s2_unknown_reach.a (0.18s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts (0.18s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h2_plain_call.a (0.19s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a (0.22s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s1_class_instance.ts (0.11s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts (0.14s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2_same_layout_this.a (0.14s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s1_class_instance.a (0.13s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r1_holder_run.a (0.11s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r3_apply.a (0.11s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r2_alias_field.ts (0.12s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r2_alias_field.a (0.12s)
    --- FAIL: TestHostMemberValueAgreesOrStops/r1_holder_run.ts (0.14s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts (0.08s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts (0.17s)
    --- FAIL: TestHostMemberValueAgreesOrStops/v1_tsc_shape.a (0.07s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h3_reordered_layout.a (0.09s)
    --- FAIL: TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts (0.12s)
    --- FAIL: TestHostMemberValueAgreesOrStops/h2_plain_call.ts (0.09s)
    --- PASS: TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a (28.18s)
    --- PASS: TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts (28.20s)
FAIL
gate cache: native hits=0 misses=4
gate cache: node hits=0 misses=4
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	28.643s
FAIL
```

## Delivery

The report is the only intended repository change. It is committed on `proof/lower-fj0w2-2-report`, created from the checked revision, and the only authorized push destination is `origin` branch `proof/lower-fj0w2-2-report`. No compiler change, dependency install, or mutant is included in the report commit.
