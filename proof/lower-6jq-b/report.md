# Compiler proof: #6jq01hc

## Result

**The oracle suite is not green at the requested HEAD.** The completed run exited **1** after **2413.655 seconds (40m 13.655s)**. Sixteen top-level tests failed: one non-environment parent with nine failing subtests, plus fifteen environment failures. All nine non-environment subtests pass individually on the supplied main revision; the parent also passes alone there. These are failures relative to the two-commit branch compared with main; this comparison does not isolate which of the two commits causes them.

The twelve missing-TypeScript-module failures pass in a supplemental HEAD run after a temporary symlink exposes the external install at the test helper's hardcoded location. Three WASI-related top-level failures remain blocked by missing `WASI_SYSROOT`. No network-fetch failure or timeout was reported by the completed suite. The full suite was not rerun after the temporary symlink check.

## Revisions and preparation

- Repository: `https://github.com/system-inc/adamic.git`.
- Tested HEAD: `adf8bee664447eb990aaab04fa787fbb5bddf7a5`, fetched from `origin compiler/logical-assignment-views`, checked out detached.
- Comparison main: `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`.
- Branch commits above main: `30379eb97` (record storage roads / constraint binding), then `adf8bee66` (logical assignment view sites, task #6jq01hc).
- `git submodule update --init --recursive` completed with exit 0. Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`. These pins are the same on main.

`/workspace/adamic-tools/env.sh` was sourced first. Its `TMPDIR=/tmp/adamic-gate` directory was created outside the repository before Go ran. No AGENTS.md files were found. Heavy commands were launched one at a time; test-owned children retained the suite's normal behavior.

The pinned upstream TypeScript checkout was provisioned at `/workspace/adamic-proof-tools/typescript`, outside the repository, with upstream `https://github.com/microsoft/TypeScript.git` and verified detached HEAD `050880ce59e30b356b686bd3144efe24f875ebc8`. A shared clone of the submodule supplied existing objects; its missing pin was fetched directly from Microsoft upstream with `git fetch --depth=1 origin 050880ce59e30b356b686bd3144efe24f875ebc8`.

`stage3/api/package.json` and its lockfile were copied to `/workspace/adamic-proof-tools`; `npm ci --prefix /workspace/adamic-proof-tools` installed three locked packages there, including TypeScript **6.0.3** and `@types/node` **25.3.3**. The installed `tsc --version` printed `Version 6.0.3`. No packages were installed inside the repository.

Environment exported for npm, Go, and their children:

```bash
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/adamic-proof-tools/typescript
export PATH=/workspace/adamic-proof-tools/node_modules/.bin:$PATH
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```

Tool versions: Go `go1.27.1 linux/amd64`; Node `v24.19.0`; Clang `20.1.8`. The environment reports four CPU cores of cgroup capacity and 16 GiB memory. `WASI_SYSROOT` was not configured.

## Completed full-suite run

```bash
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```

A Python wrapper invoked precisely these Go arguments with stdout and stderr both opened on `/tmp/oracle.log`; elapsed time was measured with `time.monotonic()`, including Go build time.

- Start UTC: `2026-11-12T22:56:59.896849+00:00`.
- End UTC: `2026-11-12T23:37:13.551631+00:00`.
- Exit: `1`.
- Wall time: `2413.655` seconds.
- Go package-reported time: `2404.146s`.
- Log: `/tmp/oracle.log`, 431 lines / 29132 bytes.
- SHA-256: `21a6c3b66ff0106213e5b391727b9819a8d7e6d7f0f6d41ccfecd147ade92683`.

```text
FAIL
gate cache: native hits=2808 misses=5281
gate cache: node hits=1739 misses=6769
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	2404.146s
FAIL
```

Earlier attempts: `/usr/bin/time` was missing, so its wrapper exited 127 before Go launched. A subsequent first Go attempt started at `2026-11-12T21:49:52.976317+00:00`, ran for more than 66 minutes, and was interrupted during an environment/connection interruption. Its wrapper did not produce a completion result and its buffered log was empty. That attempt has no usable completed exit/wall measurement and is not counted as a test result. The reported completed run was repeated with warmed caches and a detached wrapper. An old `/tmp/oracle-end.txt` belongs to the failed timing wrapper and was not used.

## Non-environment failures: HEAD versus main

The following is the complete failing parent block from the completed HEAD log, verbatim. These are diagnostic/lowering expectation failures, not missing-tool, network, or load-timeout failures.

```text
--- FAIL: TestRecordDeclaredKeyStopsArePinned (0.00s)
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g04_generic_class_value.ts (0.06s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g04_generic_class_value.ts:10:53: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g04_generic_class_value.ts:10:53: stage 0 can't lower reading Store yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n28_generic_class_field.ts (0.05s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n28_generic_class_field.ts:11:15: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n26_generic_return.ts (0.07s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n26_generic_return.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n27_explicit_type_argument.ts (0.13s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n27_explicit_type_argument.ts:7:5: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g05_generic_call_result.ts (0.07s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g05_generic_call_result.ts:8:55: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g05_generic_call_result.ts:6:29: stage 0 can't lower a value of type T yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n05_generic_parameter.ts (0.14s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n05_generic_parameter.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g03_dictionary_argument.ts (0.04s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g03_dictionary_argument.ts:7:13: stage 0 can't lower a Dict bound to T, whose constraint Options is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g03_dictionary_argument.ts:7:18: Adamic 0.1 refuses optional property outFile in Options absent from structural source Dict, which can hide fields; declare outFile on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g01_generic_arrow.ts (0.04s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g01_generic_arrow.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g01_generic_arrow.ts:5:30: stage 0 can't lower a value of type T yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g02_heritage.ts (0.07s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/g02_heritage.ts:10:34: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
```

Each failing subtest was run alone on main, sequentially, using the following command form (the `.ts` dot was escaped in the actual regular expression):

```bash
go test ./internal/oracle -count=1 -timeout 120m -v \
  -run '^TestRecordDeclaredKeyStopsArePinned$/^<subtest-name>\.ts$'
```

| Failing HEAD subtest | Main exit | Main wall seconds | Fails on main too? |
| --- | ---: | ---: | --- |
| `g04_generic_class_value.ts` | 0 | 32.909 | **No — PASS** |
| `n28_generic_class_field.ts` | 0 | 32.392 | **No — PASS** |
| `n26_generic_return.ts` | 0 | 9.352 | **No — PASS** |
| `n27_explicit_type_argument.ts` | 0 | 9.334 | **No — PASS** |
| `g05_generic_call_result.ts` | 0 | 2.788 | **No — PASS** |
| `n05_generic_parameter.ts` | 0 | 9.703 | **No — PASS** |
| `g03_dictionary_argument.ts` | 0 | 2.807 | **No — PASS** |
| `g01_generic_arrow.ts` | 0 | 3.171 | **No — PASS** |
| `g02_heritage.ts` | 0 | 9.629 | **No — PASS** |

The parent was also run alone with `-run '^TestRecordDeclaredKeyStopsArePinned$'`: exit **0**, wall **16.792s**, package time **13.569s**, with all 61 fixture subtests passing. Baseline commands and logs are retained under `/workspace/adamic-proof-tools/baseline/` (`results.json`, `01.log` through `10.log`); all selected subtests emitted matching PASS lines, so these were executed rather than unmatched filters.

Exact isolated baseline commands, in execution order:

```bash
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^g04_generic_class_value\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^n28_generic_class_field\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^n26_generic_return\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^n27_explicit_type_argument\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^g05_generic_call_result\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^n05_generic_parameter\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^g03_dictionary_argument\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^g01_generic_arrow\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$/^g02_heritage\.ts$'
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestRecordDeclaredKeyStopsArePinned$'
```

## Environment failures: missing WASI sysroot

Three top-level tests failed because the WASI compiler environment lacks `WASI_SYSROOT`. Their first error output and all nested failing test names are preserved below, verbatim. The nested `TestInspectorDoomedBrokenAxiomStopsLoudly` and `TestInspectorDoomedReadBehavior` failures are output from mutant child binaries, not additional top-level stock-suite failures. No main rerun was required for these environment failures.

### `TestStdoutColumnsRefusesWASI` — environment

```text
--- FAIL: TestStdoutColumnsRefusesWASI (1.06s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
```

### `TestInspectorDoomedAxiomMutantIsCaught` — environment

```text
--- FAIL: TestInspectorDoomedAxiomMutantIsCaught (19.36s)
    inspector_doomed_test.go:171: build go_test_-c__./internal/oracle_inspector-doomed-Axiom.test b37ee364f6de hit 0.06
    inspector_doomed_test.go:178: TestInspectorDoomedRequiresThrowBeforeTrap goes red: inspector require must throw before a doomed use
    inspector_doomed_test.go:181: broken axiom must reach exact trap: exit status 1 === RUN   TestInspectorDoomedBrokenAxiomStopsLoudly
        === PAUSE TestInspectorDoomedBrokenAxiomStopsLoudly
        === CONT  TestInspectorDoomedBrokenAxiomStopsLoudly
            inspector_doomed_test.go:76: native: wasm32-wasi requires WASI_SYSROOT
        --- FAIL: TestInspectorDoomedBrokenAxiomStopsLoudly (8.25s)
        FAIL
```

### `TestInspectorDoomedReadDropOracleMutantIsCaught` — environment

```text
--- FAIL: TestInspectorDoomedReadDropOracleMutantIsCaught (21.91s)
    inspector_doomed_test.go:187: build go_test_-c__./internal/oracle_inspector-doomed-ReadDropOracle.test 60ef178869ab hit 0.06
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
        --- FAIL: TestInspectorDoomedReadBehavior (19.21s)
            --- FAIL: TestInspectorDoomedReadBehavior/d1 (8.70s)
            --- PASS: TestInspectorDoomedReadBehavior/d1-control (0.28s)
            --- FAIL: TestInspectorDoomedReadBehavior/hoisted (8.76s)
            --- FAIL: TestInspectorDoomedReadBehavior/hoisted-uncaught (1.46s)
        FAIL
```

## Environment failures: external TypeScript module resolution

The following twelve tests failed in the full run because `check-types.cjs` could not resolve `typescript`. Each block preserves the first error lines verbatim, including the missing-module cause and require stack. `tsc` was already version 6.0.3 on PATH: the helper in `internal/oracle/assignment_proofs_test.go` explicitly sets its child's `NODE_PATH` to `<repository>/stage3/api/node_modules`, overriding a parent `NODE_PATH`. This is an environment/layout requirement, not a compiler diagnostic. No main rerun was required for these environment failures.

### `TestAssignmentProofWiderTargetMutant` — environment

```text
--- FAIL: TestAssignmentProofWiderTargetMutant (0.12s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofWiderTarget` — environment

```text
--- FAIL: TestAssignmentProofWiderTarget (0.08s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofCompoundMutant` — environment

```text
--- FAIL: TestAssignmentProofCompoundMutant (0.07s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofCompound` — environment

```text
--- FAIL: TestAssignmentProofCompound (0.26s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofChainedMutant` — environment

```text
--- FAIL: TestAssignmentProofChainedMutant (0.09s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofChained` — environment

```text
--- FAIL: TestAssignmentProofChained (0.06s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofReturnDefinedMutant` — environment

```text
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.07s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofReturnDefined` — environment

```text
--- FAIL: TestAssignmentProofReturnDefined (0.06s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofIfNarrowingMutant` — environment

```text
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.07s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofIfNarrowing` — environment

```text
--- FAIL: TestAssignmentProofIfNarrowing (0.07s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofScannerKeyword` — environment

```text
--- FAIL: TestAssignmentProofScannerKeyword (0.06s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### `TestAssignmentProofScannerKeywordMutant` — environment

```text
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.07s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
```

### Supplemental checks at HEAD

First, a parent `NODE_PATH=/workspace/adamic-proof-tools/node_modules` alone was tested: exit **1**, wall **4.879s**, with all twelve still failing because of the helper's override. Log: `/tmp/oracle-assignment-nodepath.log`.

Next, a temporary symlink `stage3/api/node_modules -> /workspace/adamic-proof-tools/node_modules` exposed the existing external install to that hardcoded location. No packages or source files were installed or changed in the repository. The same twelve tests then all passed:

```bash
go test ./internal/oracle -count=1 -timeout 120m -v -run '^TestAssignmentProof(ScannerKeyword|IfNarrowing|ReturnDefined|Chained|Compound|WiderTarget)(Mutant)?$'
```

Exit **0**, wall **19.693s**, package time **15.479s**. Log: `/tmp/oracle-assignment-external-link.log`. The temporary symlink was removed afterward. The twelve full-suite failures remain recorded above; this targeted rerun confirms their environment classification and does not replace the full-suite result.

```text
--- PASS: TestAssignmentProofWiderTarget (10.41s)
--- PASS: TestAssignmentProofWiderTargetMutant (10.46s)
--- PASS: TestAssignmentProofChained (10.46s)
--- PASS: TestAssignmentProofScannerKeyword (10.48s)
--- PASS: TestAssignmentProofCompoundMutant (10.53s)
--- PASS: TestAssignmentProofReturnDefined (2.79s)
--- PASS: TestAssignmentProofCompound (3.23s)
--- PASS: TestAssignmentProofIfNarrowingMutant (3.21s)
--- PASS: TestAssignmentProofReturnDefinedMutant (3.31s)
--- PASS: TestAssignmentProofIfNarrowing (3.30s)
--- PASS: TestAssignmentProofChainedMutant (1.85s)
--- PASS: TestAssignmentProofScannerKeywordMutant (1.81s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	15.479s
```

## Report publication scope

The repository was restored to detached `adf8bee664447eb990aaab04fa787fbb5bddf7a5` after the main comparisons, and was clean after removal of the temporary symlink. The report destination is `proof/lower-6jq-b/report.md`; the only authorized publication branch is `proof/lower-6jq-b-report`, created from that HEAD. Only this report is to be staged, committed, and pushed to origin on that branch. Compiler code, fixture expectations, dependencies, logs, and external provisioning artifacts are not part of the commit.
