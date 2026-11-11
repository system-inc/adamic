# Adamic lower #7yt238q proof

Executed on 2026-11-11 (UTC), repository `/workspace/adamic`.

## Result

Formatting and vet passed. All 20 targeted baseline subtests passed. The documented mutant was detected by exactly eight subtests, as predicted; controls and refusal cases stayed green. The mutant was restored, and `git diff --exit-code` passed. Full IR and lower suites passed. WASI coverage was unavailable because `WASI_SYSROOT` was unset.

## Checkout and environment

- Ran `git fetch origin compiler/void-undefined-3`, checked out `53a2412a58f3aea09ca24c0eec931a1f690eb59c` detached, then completed `git submodule update --init --recursive`.
- HEAD's parent is `d43437079a7e7ca723168e67c5efb5bf7ebee216`, matching the supplied main commit.
- Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` first for setup and test commands. Created its configured work directory `/tmp/adamic-gate` outside the repository before invoking Go.
- Provisioned the upstream Microsoft TypeScript checkout outside the repository at `/tmp/adamic-proof-typescript`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that path.
- Copied `stage3/api/package.json` and `package-lock.json` to `/tmp/adamic-proof-api` and ran `npm ci --prefix /tmp/adamic-proof-api`. Pinned dependencies: TypeScript `6.0.3`, `@types/node` `25.3.3` (plus their locked dependency).
- Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt` and `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`; prepended `/tmp/adamic-proof-api/node_modules/.bin` to PATH. Verified `tsc --version`: `Version 6.0.3`.
- Tools: Go `go1.27.1 linux/amd64`, Node `v24.19.0`, Clang `20.1.8`.
- No repository dependency files were modified. Logs, backups, and setup files were kept outside the repository.

## Step 1: formatting and vet

```sh
gofmt -l internal/lower internal/oracle
go vet ./internal/oracle
```

Both exited 0 and printed nothing. An initial vet invocation ran before recursive submodule setup finished and exited 1 with `go: cannot load module cohere/TypeScript/tsc listed in go.work file: open cohere/TypeScript/tsc/go.mod: no such file or directory` (environment/setup timing). After submodule setup completed, the same vet command passed with empty output. The successful rerun is the step 1 result.

## Step 2: uncached baseline oracle

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestVoidMemberCountsAsUndefined|TestVoidResultReadStillRefused' -count=1 -v -timeout 40m
```

Exit 0; package duration 287.525s. Every subtest passed; there are no baseline FAIL messages. Per-subtest results are in the table below. Release, sanitized native, and JavaScript comparisons ran. Each of the 12 union-member cases logged that its WASI leg did not run because `WASI_SYSROOT` was unset; this is an environment coverage limitation, not a failing subtest. Native cache hits/misses: 0/36; Node cache hits/misses: 0/32; probe: 0/0.

## Step 3: documented mutant

Read `internal/lower/void-undefined.mutants.md`. Asserted that the exact old text below occurred once in `internal/lower/expression.go` before replacing it:

```go
        if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 || l.phantomUndefined(member) {
```

Replacement:

```go
        if member.Flags()&checker.TypeFlagsUndefined != 0 || l.phantomUndefined(member) {
```

The assertion succeeded (count 1). Reran exactly step 2's command, including `ADAMIC_GATE_UNCACHED=1`, `-count=1`, `-v`, and `-timeout 40m`.

Exit 1; package duration 41.333s. Exactly these subtests went red:

- `TestVoidMemberCountsAsUndefined/alias.a`
- `TestVoidMemberCountsAsUndefined/alias.ts`
- `TestVoidMemberCountsAsUndefined/field.a`
- `TestVoidMemberCountsAsUndefined/field.ts`
- `TestVoidMemberCountsAsUndefined/field_alias.a`
- `TestVoidMemberCountsAsUndefined/field_alias.ts`
- `TestVoidMemberCountsAsUndefined/witness.a`
- `TestVoidMemberCountsAsUndefined/witness.ts`

All eight failed with stdout disagreements on release, sanitized, and JavaScript backends. Nullish and control cases and all eight refusal cases remained PASS. Every failing message is preserved in full in the mutant transcript below. The same WASI environment coverage limitation applied. Native cache hits/misses: 0/36; Node: 0/32; probe: 0/0.

Restored the file byte-for-byte from its external backup. `git diff --exit-code` exited 0 with no output, and `git status --short` was empty. No mutant changes remain.

### Per-subtest results

| Subtest | Baseline | Mutant |
| --- | --- | --- |
| `TestVoidMemberCountsAsUndefined/alias.a` | PASS | FAIL |
| `TestVoidMemberCountsAsUndefined/alias.ts` | PASS | FAIL |
| `TestVoidMemberCountsAsUndefined/control.a` | PASS | PASS |
| `TestVoidMemberCountsAsUndefined/control.ts` | PASS | PASS |
| `TestVoidMemberCountsAsUndefined/field.a` | PASS | FAIL |
| `TestVoidMemberCountsAsUndefined/field.ts` | PASS | FAIL |
| `TestVoidMemberCountsAsUndefined/field_alias.a` | PASS | FAIL |
| `TestVoidMemberCountsAsUndefined/field_alias.ts` | PASS | FAIL |
| `TestVoidMemberCountsAsUndefined/nullish.a` | PASS | PASS |
| `TestVoidMemberCountsAsUndefined/nullish.ts` | PASS | PASS |
| `TestVoidMemberCountsAsUndefined/witness.a` | PASS | FAIL |
| `TestVoidMemberCountsAsUndefined/witness.ts` | PASS | FAIL |
| `TestVoidResultReadStillRefused/void_argument.a` | PASS | PASS |
| `TestVoidResultReadStillRefused/void_argument.ts` | PASS | PASS |
| `TestVoidResultReadStillRefused/void_binding.a` | PASS | PASS |
| `TestVoidResultReadStillRefused/void_binding.ts` | PASS | PASS |
| `TestVoidResultReadStillRefused/void_field.a` | PASS | PASS |
| `TestVoidResultReadStillRefused/void_field.ts` | PASS | PASS |
| `TestVoidResultReadStillRefused/void_read.a` | PASS | PASS |
| `TestVoidResultReadStillRefused/void_read.ts` | PASS | PASS |

## Step 4: full suites

```sh
go test ./internal/ir -count=1 -timeout 30m
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

| Suite | Exit | Test duration | Failing tests / first error line |
| --- | --- | --- | --- |
| `internal/ir` | 0 | 125.199s | None |
| `internal/lower` | 0 | 1704.975s | None |

No environment failures were reported by either suite. These commands did not use `-v`, so passing-test skip details are not available in their output.

Complete IR output:

```text
ok  	github.com/system-inc/adamic/internal/ir	125.199s

```

Complete `/tmp/lower.log`:

```text
ok  	github.com/system-inc/adamic/internal/lower	1704.975s

```

## Report delivery

The report branch is `proof/lower-7yt-3-report`, created from the tested detached HEAD. The sole file to be committed and pushed is `proof/lower-7yt-3/report.md`. No compiler or dependency changes are included. The push destination is only `origin` branch `proof/lower-7yt-3-report`.

## Complete baseline transcript

```text
=== RUN   TestVoidMemberCountsAsUndefined
=== PAUSE TestVoidMemberCountsAsUndefined
=== RUN   TestVoidResultReadStillRefused
=== PAUSE TestVoidResultReadStillRefused
=== CONT  TestVoidMemberCountsAsUndefined
=== RUN   TestVoidMemberCountsAsUndefined/witness.a
=== CONT  TestVoidResultReadStillRefused
=== PAUSE TestVoidMemberCountsAsUndefined/witness.a
=== RUN   TestVoidMemberCountsAsUndefined/witness.ts
=== PAUSE TestVoidMemberCountsAsUndefined/witness.ts
=== RUN   TestVoidResultReadStillRefused/void_read.a
=== PAUSE TestVoidResultReadStillRefused/void_read.a
=== RUN   TestVoidMemberCountsAsUndefined/alias.a
=== PAUSE TestVoidMemberCountsAsUndefined/alias.a
=== RUN   TestVoidMemberCountsAsUndefined/alias.ts
=== PAUSE TestVoidMemberCountsAsUndefined/alias.ts
=== RUN   TestVoidMemberCountsAsUndefined/field.a
=== PAUSE TestVoidMemberCountsAsUndefined/field.a
=== RUN   TestVoidMemberCountsAsUndefined/field.ts
=== PAUSE TestVoidMemberCountsAsUndefined/field.ts
=== RUN   TestVoidMemberCountsAsUndefined/field_alias.a
=== PAUSE TestVoidMemberCountsAsUndefined/field_alias.a
=== RUN   TestVoidMemberCountsAsUndefined/field_alias.ts
=== PAUSE TestVoidMemberCountsAsUndefined/field_alias.ts
=== RUN   TestVoidMemberCountsAsUndefined/nullish.a
=== PAUSE TestVoidMemberCountsAsUndefined/nullish.a
=== RUN   TestVoidMemberCountsAsUndefined/nullish.ts
=== PAUSE TestVoidMemberCountsAsUndefined/nullish.ts
=== RUN   TestVoidMemberCountsAsUndefined/control.a
=== PAUSE TestVoidMemberCountsAsUndefined/control.a
=== RUN   TestVoidResultReadStillRefused/void_read.ts
=== PAUSE TestVoidResultReadStillRefused/void_read.ts
=== RUN   TestVoidMemberCountsAsUndefined/control.ts
=== PAUSE TestVoidMemberCountsAsUndefined/control.ts
=== CONT  TestVoidMemberCountsAsUndefined/witness.a
=== CONT  TestVoidMemberCountsAsUndefined/field_alias.ts
=== CONT  TestVoidMemberCountsAsUndefined/field.a
=== CONT  TestVoidMemberCountsAsUndefined/alias.a
=== RUN   TestVoidResultReadStillRefused/void_argument.a
=== PAUSE TestVoidResultReadStillRefused/void_argument.a
=== RUN   TestVoidResultReadStillRefused/void_argument.ts
=== PAUSE TestVoidResultReadStillRefused/void_argument.ts
=== RUN   TestVoidResultReadStillRefused/void_binding.a
=== PAUSE TestVoidResultReadStillRefused/void_binding.a
=== RUN   TestVoidResultReadStillRefused/void_binding.ts
=== PAUSE TestVoidResultReadStillRefused/void_binding.ts
=== RUN   TestVoidResultReadStillRefused/void_field.a
=== PAUSE TestVoidResultReadStillRefused/void_field.a
=== RUN   TestVoidResultReadStillRefused/void_field.ts
=== PAUSE TestVoidResultReadStillRefused/void_field.ts
=== CONT  TestVoidMemberCountsAsUndefined/control.a
=== NAME  TestVoidMemberCountsAsUndefined/witness.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 241.52
=== NAME  TestVoidMemberCountsAsUndefined/field_alias.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 miss 241.52
=== NAME  TestVoidMemberCountsAsUndefined/control.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 241.52
=== NAME  TestVoidMemberCountsAsUndefined/alias.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 241.53
=== NAME  TestVoidMemberCountsAsUndefined/field.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 241.53
=== NAME  TestVoidMemberCountsAsUndefined/field_alias.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/field_alias.ts on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/alias.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/alias.a on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/field.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/field.a on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/control.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/control.a on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/witness.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/witness.a on wasm32-wasi: WASI_SYSROOT is unset
=== CONT  TestVoidMemberCountsAsUndefined/nullish.ts
=== CONT  TestVoidMemberCountsAsUndefined/nullish.a
=== CONT  TestVoidMemberCountsAsUndefined/alias.ts
=== CONT  TestVoidMemberCountsAsUndefined/field_alias.a
=== CONT  TestVoidMemberCountsAsUndefined/field.ts
=== NAME  TestVoidMemberCountsAsUndefined/alias.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 0.04
=== NAME  TestVoidMemberCountsAsUndefined/field.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 0.02
=== NAME  TestVoidMemberCountsAsUndefined/field_alias.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 0.03
=== NAME  TestVoidMemberCountsAsUndefined/nullish.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 0.01
=== NAME  TestVoidMemberCountsAsUndefined/field.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/field.ts on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/nullish.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 0.01
=== CONT  TestVoidMemberCountsAsUndefined/witness.ts
=== NAME  TestVoidMemberCountsAsUndefined/alias.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/alias.ts on wasm32-wasi: WASI_SYSROOT is unset
=== CONT  TestVoidMemberCountsAsUndefined/control.ts
=== NAME  TestVoidMemberCountsAsUndefined/nullish.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/nullish.ts on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/field_alias.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/field_alias.a on wasm32-wasi: WASI_SYSROOT is unset
=== CONT  TestVoidResultReadStillRefused/void_read.a
=== NAME  TestVoidMemberCountsAsUndefined/nullish.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/nullish.a on wasm32-wasi: WASI_SYSROOT is unset
=== CONT  TestVoidResultReadStillRefused/void_field.ts
=== CONT  TestVoidResultReadStillRefused/void_field.a
=== CONT  TestVoidResultReadStillRefused/void_binding.ts
=== CONT  TestVoidResultReadStillRefused/void_binding.a
=== CONT  TestVoidResultReadStillRefused/void_argument.ts
=== CONT  TestVoidResultReadStillRefused/void_argument.a
=== CONT  TestVoidResultReadStillRefused/void_read.ts
=== NAME  TestVoidMemberCountsAsUndefined/control.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 0.01
=== NAME  TestVoidMemberCountsAsUndefined/witness.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 0.01
=== NAME  TestVoidMemberCountsAsUndefined/control.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/control.ts on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/witness.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/witness.ts on wasm32-wasi: WASI_SYSROOT is unset
--- PASS: TestVoidMemberCountsAsUndefined (0.00s)
    --- PASS: TestVoidMemberCountsAsUndefined/field.a (283.63s)
    --- PASS: TestVoidMemberCountsAsUndefined/field_alias.ts (283.64s)
    --- PASS: TestVoidMemberCountsAsUndefined/alias.a (283.64s)
    --- PASS: TestVoidMemberCountsAsUndefined/witness.a (283.65s)
    --- PASS: TestVoidMemberCountsAsUndefined/control.a (283.65s)
    --- PASS: TestVoidMemberCountsAsUndefined/field.ts (2.23s)
    --- PASS: TestVoidMemberCountsAsUndefined/alias.ts (2.26s)
    --- PASS: TestVoidMemberCountsAsUndefined/nullish.ts (2.31s)
    --- PASS: TestVoidMemberCountsAsUndefined/field_alias.a (2.33s)
    --- PASS: TestVoidMemberCountsAsUndefined/nullish.a (2.35s)
    --- PASS: TestVoidMemberCountsAsUndefined/control.ts (1.39s)
    --- PASS: TestVoidMemberCountsAsUndefined/witness.ts (1.43s)
--- PASS: TestVoidResultReadStillRefused (0.00s)
    --- PASS: TestVoidResultReadStillRefused/void_read.a (0.55s)
    --- PASS: TestVoidResultReadStillRefused/void_field.ts (0.57s)
    --- PASS: TestVoidResultReadStillRefused/void_field.a (0.68s)
    --- PASS: TestVoidResultReadStillRefused/void_binding.ts (0.39s)
    --- PASS: TestVoidResultReadStillRefused/void_binding.a (0.44s)
    --- PASS: TestVoidResultReadStillRefused/void_argument.ts (0.44s)
    --- PASS: TestVoidResultReadStillRefused/void_argument.a (0.31s)
    --- PASS: TestVoidResultReadStillRefused/void_read.ts (0.46s)
PASS
gate cache: native hits=0 misses=36
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	287.525s

```

## Complete mutant transcript (all FAIL messages)

```text
=== RUN   TestVoidMemberCountsAsUndefined
=== PAUSE TestVoidMemberCountsAsUndefined
=== RUN   TestVoidResultReadStillRefused
=== PAUSE TestVoidResultReadStillRefused
=== CONT  TestVoidMemberCountsAsUndefined
=== RUN   TestVoidMemberCountsAsUndefined/witness.a
=== PAUSE TestVoidMemberCountsAsUndefined/witness.a
=== RUN   TestVoidMemberCountsAsUndefined/witness.ts
=== PAUSE TestVoidMemberCountsAsUndefined/witness.ts
=== RUN   TestVoidMemberCountsAsUndefined/alias.a
=== PAUSE TestVoidMemberCountsAsUndefined/alias.a
=== RUN   TestVoidMemberCountsAsUndefined/alias.ts
=== PAUSE TestVoidMemberCountsAsUndefined/alias.ts
=== RUN   TestVoidMemberCountsAsUndefined/field.a
=== PAUSE TestVoidMemberCountsAsUndefined/field.a
=== RUN   TestVoidMemberCountsAsUndefined/field.ts
=== PAUSE TestVoidMemberCountsAsUndefined/field.ts
=== RUN   TestVoidMemberCountsAsUndefined/field_alias.a
=== PAUSE TestVoidMemberCountsAsUndefined/field_alias.a
=== RUN   TestVoidMemberCountsAsUndefined/field_alias.ts
=== PAUSE TestVoidMemberCountsAsUndefined/field_alias.ts
=== RUN   TestVoidMemberCountsAsUndefined/nullish.a
=== PAUSE TestVoidMemberCountsAsUndefined/nullish.a
=== RUN   TestVoidMemberCountsAsUndefined/nullish.ts
=== PAUSE TestVoidMemberCountsAsUndefined/nullish.ts
=== RUN   TestVoidMemberCountsAsUndefined/control.a
=== PAUSE TestVoidMemberCountsAsUndefined/control.a
=== RUN   TestVoidMemberCountsAsUndefined/control.ts
=== PAUSE TestVoidMemberCountsAsUndefined/control.ts
=== CONT  TestVoidMemberCountsAsUndefined/witness.a
=== CONT  TestVoidMemberCountsAsUndefined/field_alias.ts
=== CONT  TestVoidMemberCountsAsUndefined/field.a
=== CONT  TestVoidMemberCountsAsUndefined/field_alias.a
=== CONT  TestVoidMemberCountsAsUndefined/field.ts
=== NAME  TestVoidMemberCountsAsUndefined/field_alias.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 29.17
=== NAME  TestVoidMemberCountsAsUndefined/field.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 miss 29.17
=== NAME  TestVoidMemberCountsAsUndefined/field_alias.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 29.17
=== NAME  TestVoidMemberCountsAsUndefined/witness.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 29.17
=== NAME  TestVoidMemberCountsAsUndefined/field.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 28.98
=== NAME  TestVoidMemberCountsAsUndefined/field.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/field.ts on wasm32-wasi: WASI_SYSROOT is unset
    void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
    void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
    void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "false false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
=== NAME  TestVoidMemberCountsAsUndefined/witness.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/witness.a on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/field.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/field.a on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/witness.a
    void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false true object\ntrue false true object\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
=== NAME  TestVoidMemberCountsAsUndefined/field.a
    void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
=== NAME  TestVoidMemberCountsAsUndefined/witness.a
    void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false true object\ntrue false true object\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
    void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "true false true object\nfalse false true undefined\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
=== NAME  TestVoidMemberCountsAsUndefined/field.a
    void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
    void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "false false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
=== NAME  TestVoidMemberCountsAsUndefined/field_alias.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/field_alias.a on wasm32-wasi: WASI_SYSROOT is unset
    void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
    void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
    void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "false false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
=== NAME  TestVoidMemberCountsAsUndefined/field_alias.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/field_alias.ts on wasm32-wasi: WASI_SYSROOT is unset
    void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
    void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
    void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "false false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
=== CONT  TestVoidMemberCountsAsUndefined/alias.a
=== CONT  TestVoidMemberCountsAsUndefined/alias.ts
=== CONT  TestVoidMemberCountsAsUndefined/witness.ts
=== CONT  TestVoidMemberCountsAsUndefined/control.a
=== CONT  TestVoidResultReadStillRefused
=== RUN   TestVoidResultReadStillRefused/void_read.a
=== PAUSE TestVoidResultReadStillRefused/void_read.a
=== RUN   TestVoidResultReadStillRefused/void_read.ts
=== PAUSE TestVoidResultReadStillRefused/void_read.ts
=== RUN   TestVoidResultReadStillRefused/void_argument.a
=== PAUSE TestVoidResultReadStillRefused/void_argument.a
=== RUN   TestVoidResultReadStillRefused/void_argument.ts
=== PAUSE TestVoidResultReadStillRefused/void_argument.ts
=== RUN   TestVoidResultReadStillRefused/void_binding.a
=== PAUSE TestVoidResultReadStillRefused/void_binding.a
=== RUN   TestVoidResultReadStillRefused/void_binding.ts
=== PAUSE TestVoidResultReadStillRefused/void_binding.ts
=== RUN   TestVoidResultReadStillRefused/void_field.a
=== PAUSE TestVoidResultReadStillRefused/void_field.a
=== RUN   TestVoidResultReadStillRefused/void_field.ts
=== PAUSE TestVoidResultReadStillRefused/void_field.ts
=== CONT  TestVoidMemberCountsAsUndefined/control.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 0.04
=== NAME  TestVoidMemberCountsAsUndefined/witness.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 0.01
=== NAME  TestVoidMemberCountsAsUndefined/control.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 0.02
=== NAME  TestVoidMemberCountsAsUndefined/alias.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 0.04
=== NAME  TestVoidMemberCountsAsUndefined/control.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/control.ts on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/witness.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/witness.ts on wasm32-wasi: WASI_SYSROOT is unset
    void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false true object\ntrue false true object\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
    void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false true object\ntrue false true object\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
    void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "true false true object\nfalse false true undefined\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
=== CONT  TestVoidMemberCountsAsUndefined/nullish.ts
=== CONT  TestVoidMemberCountsAsUndefined/nullish.a
=== NAME  TestVoidMemberCountsAsUndefined/alias.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 0.02
=== NAME  TestVoidMemberCountsAsUndefined/alias.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/alias.a on wasm32-wasi: WASI_SYSROOT is unset
    void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object\ntrue false object\nfalse false object\ntrue false object\ntrue false object\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
    void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object\ntrue false object\nfalse false object\ntrue false object\ntrue false object\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
    void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "true false object\nfalse false undefined\nfalse false object\ntrue false object\nfalse false undefined\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
=== NAME  TestVoidMemberCountsAsUndefined/control.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/control.a on wasm32-wasi: WASI_SYSROOT is unset
=== CONT  TestVoidResultReadStillRefused/void_read.a
=== CONT  TestVoidResultReadStillRefused/void_binding.a
=== NAME  TestVoidMemberCountsAsUndefined/alias.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/alias.ts on wasm32-wasi: WASI_SYSROOT is unset
    void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object\ntrue false object\nfalse false object\ntrue false object\ntrue false object\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
    void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object\ntrue false object\nfalse false object\ntrue false object\ntrue false object\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
    void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "true false object\nfalse false undefined\nfalse false object\ntrue false object\nfalse false undefined\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
=== CONT  TestVoidResultReadStillRefused/void_field.ts
=== CONT  TestVoidResultReadStillRefused/void_field.a
=== CONT  TestVoidResultReadStillRefused/void_binding.ts
=== CONT  TestVoidResultReadStillRefused/void_argument.a
=== CONT  TestVoidResultReadStillRefused/void_argument.ts
=== CONT  TestVoidResultReadStillRefused/void_read.ts
--- PASS: TestVoidResultReadStillRefused (0.00s)
    --- PASS: TestVoidResultReadStillRefused/void_read.a (0.33s)
    --- PASS: TestVoidResultReadStillRefused/void_binding.a (0.37s)
    --- PASS: TestVoidResultReadStillRefused/void_field.ts (0.38s)
    --- PASS: TestVoidResultReadStillRefused/void_field.a (0.38s)
    --- PASS: TestVoidResultReadStillRefused/void_binding.ts (0.55s)
    --- PASS: TestVoidResultReadStillRefused/void_argument.a (0.56s)
    --- PASS: TestVoidResultReadStillRefused/void_argument.ts (0.52s)
    --- PASS: TestVoidResultReadStillRefused/void_read.ts (0.30s)
=== NAME  TestVoidMemberCountsAsUndefined/nullish.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 0.02
=== NAME  TestVoidMemberCountsAsUndefined/nullish.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 533d2ea9bac5 hit 0.02
=== NAME  TestVoidMemberCountsAsUndefined/nullish.ts
    void_undefined_test.go:53: wasi leg not run: void_undefined/nullish.ts on wasm32-wasi: WASI_SYSROOT is unset
=== NAME  TestVoidMemberCountsAsUndefined/nullish.a
    void_undefined_test.go:53: wasi leg not run: void_undefined/nullish.a on wasm32-wasi: WASI_SYSROOT is unset
--- FAIL: TestVoidMemberCountsAsUndefined (0.00s)
    --- FAIL: TestVoidMemberCountsAsUndefined/field.ts (38.26s)
    --- FAIL: TestVoidMemberCountsAsUndefined/witness.a (38.26s)
    --- FAIL: TestVoidMemberCountsAsUndefined/field.a (38.27s)
    --- FAIL: TestVoidMemberCountsAsUndefined/field_alias.a (38.27s)
    --- FAIL: TestVoidMemberCountsAsUndefined/field_alias.ts (38.27s)
    --- PASS: TestVoidMemberCountsAsUndefined/control.ts (1.49s)
    --- FAIL: TestVoidMemberCountsAsUndefined/witness.ts (1.51s)
    --- FAIL: TestVoidMemberCountsAsUndefined/alias.a (1.59s)
    --- PASS: TestVoidMemberCountsAsUndefined/control.a (1.59s)
    --- FAIL: TestVoidMemberCountsAsUndefined/alias.ts (1.74s)
    --- PASS: TestVoidMemberCountsAsUndefined/nullish.ts (1.42s)
    --- PASS: TestVoidMemberCountsAsUndefined/nullish.a (1.46s)
FAIL
gate cache: native hits=0 misses=36
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	41.333s
FAIL

```
