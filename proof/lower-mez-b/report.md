# Compiler proof for #mez92e7

## Result

**FAIL at HEAD.** The requested oracle command exited **1** in **1,378.903 seconds (22m 58.903s)**. Nine non-environment subtests fail under `TestRecordDeclaredKeyStopsArePinned`; all nine pass when run individually on the specified main commit. The parent test also passes alone on main. These are regressions relative to main: HEAD changes the refusal or rejects programs whose pinned expectation is that they lower.

The other 15 top-level failures are environmental: twelve cannot resolve the installed TypeScript module, and three require an unconfigured WASI sysroot. These failures prevent a clean overall proof; they are not evidence of compiler regressions. Every failing top-level test and its first error lines appear below, including the nested failures from child test binaries.

## Completed oracle run

- Repository: `/workspace/adamic`, origin `https://github.com/system-inc/adamic.git`.
- Tested HEAD: `11db2faf88c95c059a45ec1dde98937d2ad93ebb` (`compiler/record-storage-roads`).
- Baseline main: `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`, HEAD's immediate parent.
- Command: `go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1` (the supervisor used the equivalent argv and combined stdout/stderr file redirection).
- Started: `2026-11-12T21:46:07.354350+00:00`. Finished: `2026-11-12T22:09:06.257743+00:00`.
- Exit: `1`. Wall time: `1378.903380` seconds, measured with Python's monotonic clock around the subprocess.
- Full local log: `/tmp/oracle.log`; SHA-256 `f5830dd26ac2ce1e28f1a3de21c0d946a39cfbaa0e65c1d7fd37716c1940c898`.
- Final Go output (package duration excludes some command overhead):

```text
FAIL
gate cache: native hits=2808 misses=5271
gate cache: node hits=1739 misses=6751
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	1373.294s
FAIL
```

## Non-environment failures: all pass on main

The following block is verbatim from HEAD's log. It lists the parent and every failing subtest with its first error line. The new storage-constraint diagnostic differs from the pinned diagnostic in four cases and replaces successful lowering in five cases.

```text
--- FAIL: TestRecordDeclaredKeyStopsArePinned (0.00s)
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g04_generic_class_value.ts (0.04s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g04_generic_class_value.ts:10:53: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g04_generic_class_value.ts:10:53: stage 0 can't lower reading Store yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n28_generic_class_field.ts (0.03s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n28_generic_class_field.ts:11:15: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n27_explicit_type_argument.ts (0.03s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n27_explicit_type_argument.ts:7:5: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n26_generic_return.ts (0.04s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n26_generic_return.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n05_generic_parameter.ts (0.05s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n05_generic_parameter.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g05_generic_call_result.ts (0.02s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g05_generic_call_result.ts:8:55: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g05_generic_call_result.ts:6:29: stage 0 can't lower a value of type T yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g02_heritage.ts (0.08s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/g02_heritage.ts:10:34: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g01_generic_arrow.ts (0.11s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g01_generic_arrow.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g01_generic_arrow.ts:5:30: stage 0 can't lower a value of type T yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g03_dictionary_argument.ts (0.03s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g03_dictionary_argument.ts:7:13: stage 0 can't lower a Dict bound to T, whose constraint Options is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g03_dictionary_argument.ts:7:18: Adamic 0.1 refuses optional property outFile in Options absent from structural source Dict, which can hide fields; declare outFile on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
```

Each row below was a separate sequential `go test` invocation on detached main, using the same toolchain and exported environment. No non-environment failure reproduces there. Subtest names are qualified by `TestRecordDeclaredKeyStopsArePinned/`.

| Test on main | Exit | Wall seconds | Fails there too? |
| --- | ---: | ---: | --- |
| `TestRecordDeclaredKeyStopsArePinned` | 0 | 45.814 | No; PASS |
| `g04_generic_class_value.ts` | 0 | 1.980 | No; PASS |
| `n28_generic_class_field.ts` | 0 | 6.495 | No; PASS |
| `n27_explicit_type_argument.ts` | 0 | 6.272 | No; PASS |
| `n26_generic_return.ts` | 0 | 6.129 | No; PASS |
| `n05_generic_parameter.ts` | 0 | 6.441 | No; PASS |
| `g05_generic_call_result.ts` | 0 | 2.268 | No; PASS |
| `g02_heritage.ts` | 0 | 6.198 | No; PASS |
| `g01_generic_arrow.ts` | 0 | 2.370 | No; PASS |
| `g03_dictionary_argument.ts` | 0 | 1.951 | No; PASS |

Exact baseline commands and their output, in execution order:

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$'
ok  	github.com/system-inc/adamic/internal/oracle	24.608s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^g04_generic_class_value\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	0.056s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^n28_generic_class_field\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	4.146s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^n27_explicit_type_argument\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	3.976s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^n26_generic_return\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	3.929s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^n05_generic_parameter\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	4.191s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^g05_generic_call_result\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	0.039s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^g02_heritage\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	4.258s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^g01_generic_arrow\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	0.045s
```

```text
go test ./internal/oracle -count=1 -timeout 120m -run '^TestRecordDeclaredKeyStopsArePinned$/^g03_dictionary_argument\.ts$'
ok  	github.com/system-inc/adamic/internal/oracle	0.039s
```

## Environment failures

### TypeScript module resolution: twelve tests

All twelve are **ENVIRONMENT: missing module resolution**, despite the pinned packages being installed outside the repository and `tsc 6.0.3` being on PATH. `assignmentProofTypecheck` in `internal/oracle/assignment_proofs_test.go` explicitly sets child `NODE_PATH` to `/workspace/adamic/stage3/api/node_modules`. That location has no installation; the external install is `/tmp/adamic-proof-api/node_modules`. No repository symlink or package installation was added to hide this condition. These failures were not rerun on main because they are environmental.

First error lines for each test, verbatim:

```text
--- FAIL: TestAssignmentProofWiderTargetMutant (0.03s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofWiderTarget (0.03s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofCompoundMutant (0.03s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofCompound (0.04s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofChainedMutant (0.07s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofChained (0.10s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.05s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofReturnDefined (0.05s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.06s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofIfNarrowing (0.04s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.06s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

```text
--- FAIL: TestAssignmentProofScannerKeyword (0.04s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
```

### WASI sysroot: three top-level tests and their child failures

All three are **ENVIRONMENT: missing WASI_SYSROOT**. Native WASI compilation stops before the expected compiler/runtime behavior can be checked. The two inspector mutant tests expose additional named failures in their child test binaries, shown verbatim below. They are child failures, not extra top-level failures of the suite. These environment failures were not rerun on main.

```text
--- FAIL: TestStdoutColumnsRefusesWASI (0.41s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
```

```text
--- FAIL: TestInspectorDoomedAxiomMutantIsCaught (10.60s)
    inspector_doomed_test.go:171: build go_test_-c__./internal/oracle_inspector-doomed-Axiom.test 3f6d05d8a937 hit 0.03
    inspector_doomed_test.go:178: TestInspectorDoomedRequiresThrowBeforeTrap goes red: inspector require must throw before a doomed use
    inspector_doomed_test.go:181: broken axiom must reach exact trap: exit status 1 === RUN   TestInspectorDoomedBrokenAxiomStopsLoudly
        === PAUSE TestInspectorDoomedBrokenAxiomStopsLoudly
        === CONT  TestInspectorDoomedBrokenAxiomStopsLoudly
            inspector_doomed_test.go:76: native: wasm32-wasi requires WASI_SYSROOT
        --- FAIL: TestInspectorDoomedBrokenAxiomStopsLoudly (4.60s)
        FAIL
```

```text
--- FAIL: TestInspectorDoomedReadDropOracleMutantIsCaught (11.59s)
    inspector_doomed_test.go:187: build go_test_-c__./internal/oracle_inspector-doomed-ReadDropOracle.test cf20687d9146 hit 0.04
    inspector_doomed_test.go:190: dropping read must reach both exact traps: exit status 1 === RUN   TestInspectorDoomedReadBehavior
        === PAUSE TestInspectorDoomedReadBehavior
        === CONT  TestInspectorDoomedReadBehavior
        === RUN   TestInspectorDoomedReadBehavior/d1
            inspector_doomed_test.go:123: native: wasm32-wasi requires WASI_SYSROOT
        === RUN   TestInspectorDoomedReadBehavior/d1-control
        === RUN   TestInspectorDoomedReadBehavior/hoisted
            inspector_doomed_test.go:123: native: wasm32-wasi requires WASI_SYSROOT
        === RUN   TestInspectorDoomedReadBehavior/hoisted-uncaught
            inspector_doomed_test.go:123: native: wasm32-wasi requires WASI_SYSROOT
        --- FAIL: TestInspectorDoomedReadBehavior (10.19s)
            --- FAIL: TestInspectorDoomedReadBehavior/d1 (4.75s)
            --- PASS: TestInspectorDoomedReadBehavior/d1-control (0.16s)
            --- FAIL: TestInspectorDoomedReadBehavior/hoisted (4.27s)
            --- FAIL: TestInspectorDoomedReadBehavior/hoisted-uncaught (1.01s)
        FAIL
```

No completed-suite failure was a network fetch failure or timeout under load.

## Setup and execution record

1. Loaded `/workspace/adamic-tools/env.sh` before repository work. It sets PATH, `GOTOOLCHAIN=auto`, and `TMPDIR=/tmp/adamic-gate`; created that missing work directory outside the repository before Go ran. The initial checkout was clean.
2. Ran `git fetch origin compiler/record-storage-roads`, checked out the requested HEAD detached, and ran `git submodule update --init --recursive`, all successfully. Pins: `cohere` = `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` = `d92d9bfee114c80be2c375d72edae966176e3a4f`.
3. Provisioned upstream `https://github.com/microsoft/TypeScript.git` outside the repository at `/tmp/adamic-proof-typescript`, fetched and checked out detached `050880ce59e30b356b686bd3144efe24f875ebc8`, and verified that exact HEAD. Exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-proof-typescript` for the oracle and baseline commands.
4. Copied `stage3/api/package.json` and its lockfile to `/tmp/adamic-proof-api` and ran `npm ci --prefix /tmp/adamic-proof-api` successfully (three packages). Pinned direct packages: TypeScript `6.0.3`, `@types/node` `25.3.3`. Prepended `/tmp/adamic-proof-api/node_modules/.bin` to PATH and verified `tsc --version` = `Version 6.0.3`.
5. Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true` before npm and every oracle/baseline run, inherited by admission-gate children. TLS verification remained enabled.
6. Verified Go `go1.27.1 linux/amd64`, Clang `20.1.8`, and Node `v24.19.0`. The repository remained clean before testing. Ran one heavy command at a time, including all baseline invocations.
7. Setup/run interruptions: `/usr/bin/time` was absent, so the first timing launch exited 127 before Go started; this is an environment/tool issue, not a failing test. A subsequent oracle attempt was interrupted by the workspace/runtime transition after about 34 minutes, leaving an empty buffered log and no reliable completion record. A detached supervisor attempt was terminated when its launch command ended and also produced no completion record. Neither attempt is counted as a completed proof. The reported 1,378.903-second result comes solely from the later foreground-supervised run, whose exit and monotonic wall time were recorded in `/tmp/oracle-run.json`.
8. After the full run, checked out the specified main commit detached and ran the parent test and each failing non-environment subtest alone, sequentially. All ten commands exited 0. Local baseline metadata is `/tmp/oracle-main-results.json`; individual logs are `/tmp/oracle-main-0.log` through `/tmp/oracle-main-9.log`.
9. Restored the requested HEAD and created `proof/lower-mez-b-report` from it for this report. Publication is restricted to this report file and that branch; compiler source, fixtures, and dependency files were not changed or committed. This proof used the [cloud environment runtime skill](skill://plugin_connector_1p_c5b7d5df5d7081918f2c4be5a633ed5d/cloud-environment-runtime/SKILL.md) for environment and package guidance.

## Publication

Report path: `proof/lower-mez-b/report.md`. Report branch: `proof/lower-mez-b-report`, based directly on `11db2faf88c95c059a45ec1dde98937d2ad93ebb`. The only authorized remote update is creation of `refs/heads/proof/lower-mez-b-report` on origin. No commit or push to any other branch is part of this proof.
