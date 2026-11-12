Proof for Adamic task #mez92e7, 2026-11-12 (UTC).

The full oracle proof **failed: exit 1, wall time 1,376.615 seconds (22m 56.615s)**. There are **nine non-environment failing leaf tests**, all in `TestRecordDeclaredKeyStopsArePinned`; **each passes when run alone at main**, and the parent test passes there too. Five cases now refuse lowering where their existing contract expects lowering; four produce a different refusal or diagnostic. These are changes relative to main, not environment failures. The change’s own `TestRecordStorageRoadsArePinned` passes all 11 cases at HEAD. The full run also has **15 environmental failing leaf tests**: 12 cannot resolve the TypeScript Node module, and three lack `WASI_SYSROOT`. This is not a passing compiler proof.

The completed command, run from `/workspace/adamic`:

```bash
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```

It started at `2026-11-12T21:40:59Z` and finished at `2026-11-12T22:03:56Z`. Bash’s built-in timer measured the full command, including Go’s build/link overhead; Go reported package elapsed time `1370.428s`. Timing and output remain outside the repository in `/tmp/oracle.time` and `/tmp/oracle.log`. The log contains 16 failing top-level tests: the 12 assignment-proof tests, the three WASI-dependent tests, and the record-declared-key parent with nine failing subtests. Mutant-child failures quoted below are nested evidence for their enclosing tests, not additional leaf failures in the top-level suite. There was no network-fetch failure or timeout reported in the completed run.

All non-environment failures and their first error lines follow **verbatim** from `/tmp/oracle.log`, including the failing parent and every failing subtest:

```text
--- FAIL: TestRecordDeclaredKeyStopsArePinned (0.00s)
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n28_generic_class_field.ts (0.03s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n28_generic_class_field.ts:11:15: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n27_explicit_type_argument.ts (0.08s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n27_explicit_type_argument.ts:7:5: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n26_generic_return.ts (0.05s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n26_generic_return.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/n05_generic_parameter.ts (0.03s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/n05_generic_parameter.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g05_generic_call_result.ts (0.02s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g05_generic_call_result.ts:8:55: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g05_generic_call_result.ts:6:29: stage 0 can't lower a value of type T yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g04_generic_class_value.ts (0.06s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g04_generic_class_value.ts:10:53: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g04_generic_class_value.ts:10:53: stage 0 can't lower reading Store yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g03_dictionary_argument.ts (0.04s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g03_dictionary_argument.ts:7:13: stage 0 can't lower a Dict bound to T, whose constraint Options is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g03_dictionary_argument.ts:7:18: Adamic 0.1 refuses optional property outFile in Options absent from structural source Dict, which can hide fields; declare outFile on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g01_generic_arrow.ts (0.05s)
        record_declared_keys_test.go:152: stop: got "internal/oracle/testdata/record_declared_keys_stops/g01_generic_arrow.ts:7:1: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet", want "internal/oracle/testdata/record_declared_keys_stops/g01_generic_arrow.ts:5:30: stage 0 can't lower a value of type T yet"
    --- FAIL: TestRecordDeclaredKeyStopsArePinned/g02_heritage.ts (0.09s)
        record_declared_keys_test.go:157: Lower: /workspace/adamic/internal/oracle/testdata/record_declared_keys_stops/g02_heritage.ts:10:34: stage 0 can't lower a Options bound to T, whose constraint Dict is stored otherwise (fixed objects and records have different storage; copy explicitly) yet
```

Each failing subtest was run in a separate Go invocation at main `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`, sequentially, using the same toolchain and exported environment as the full proof. None failed on main. The anchored selection pattern was:

```bash
go test ./internal/oracle -run '^TestRecordDeclaredKeyStopsArePinned$/^<escaped-filename>$' -count=1 -timeout 120m -v
```

| Main test selection under `TestRecordDeclaredKeyStopsArePinned` | Exit | Wall seconds | Fails on main? |
| --- | ---: | ---: | --- |
| `n28_generic_class_field.ts` | 0 | 43.361 | No |
| `n27_explicit_type_argument.ts` | 0 | 6.815 | No |
| `n26_generic_return.ts` | 0 | 6.678 | No |
| `n05_generic_parameter.ts` | 0 | 6.586 | No |
| `g05_generic_call_result.ts` | 0 | 1.925 | No |
| `g04_generic_class_value.ts` | 0 | 2.221 | No |
| `g03_dictionary_argument.ts` | 0 | 1.977 | No |
| `g01_generic_arrow.ts` | 0 | 2.170 | No |
| `g02_heritage.ts` | 0 | 6.894 | No |
| `TestRecordDeclaredKeyStopsArePinned` | 0 | 11.473 | No |

The final table row is the parent alone, selected with `-run '^TestRecordDeclaredKeyStopsArePinned$'`. The first isolated invocation includes main’s initial test-binary build. Each invocation’s command, exit, wall time and log path are retained in `/tmp/mez2-main-results.json`; verbose logs are `/tmp/mez2-main-<filename>.log` and `/tmp/mez2-main-parent.log`. Main was checked out detached for these comparisons; both submodule revisions were unchanged. HEAD was restored afterward.

**Environment: missing TypeScript Node module (12 tests).** The pinned packages were installed outside the repository, and `tsc 6.0.3` was available on PATH. However, `assignmentProofTypecheck` in `internal/oracle/assignment_proofs_test.go:73` explicitly replaces the child’s `NODE_PATH` with `../../stage3/api/node_modules`, where no packages were installed. Therefore `require('typescript')` cannot find the external installation. Every affected test and its first error lines are quoted verbatim below:

```text
--- FAIL: TestAssignmentProofWiderTargetMutant (0.03s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofWiderTarget (0.03s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofCompoundMutant (0.05s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofCompound (0.05s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofChainedMutant (0.03s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofChained (0.04s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofReturnDefined (0.05s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.06s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.03s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofIfNarrowing (0.04s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.05s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

```text
--- FAIL: TestAssignmentProofScannerKeyword (0.03s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
```

An additional HEAD-only check selected exactly these 12 tests with an anchored alternation, `-count=1 -timeout 120m -v`, and exported `NODE_PATH=/workspace/scratch/mez2-api/node_modules`. It still failed with exit 1 in 3.039 seconds because the helper overwrites that value; the errors remain `Cannot find module 'typescript'`. This is not evidence that these tests pass. No repository symlink or dependency installation was added to work around their hard-coded lookup. Logs: `/tmp/mez2-assignment-nodepath.log`; command and timing: `/tmp/mez2-assignment-nodepath.json`. Environment failures did not require main comparisons.

**Environment: missing WASI sysroot (three tests).** These failures explicitly say `native: wasm32-wasi requires WASI_SYSROOT`. The mutant-child failure names and their error lines are included verbatim with their enclosing failing tests:

```text
--- FAIL: TestStdoutColumnsRefusesWASI (0.34s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestInspectorDoomedAxiomMutantIsCaught (10.88s)
    inspector_doomed_test.go:171: build go_test_-c__./internal/oracle_inspector-doomed-Axiom.test dfd90b785233 hit 0.05
    inspector_doomed_test.go:178: TestInspectorDoomedRequiresThrowBeforeTrap goes red: inspector require must throw before a doomed use
    inspector_doomed_test.go:181: broken axiom must reach exact trap: exit status 1 === RUN   TestInspectorDoomedBrokenAxiomStopsLoudly
        === PAUSE TestInspectorDoomedBrokenAxiomStopsLoudly
        === CONT  TestInspectorDoomedBrokenAxiomStopsLoudly
            inspector_doomed_test.go:76: native: wasm32-wasi requires WASI_SYSROOT
        --- FAIL: TestInspectorDoomedBrokenAxiomStopsLoudly (4.84s)
        FAIL
--- FAIL: TestInspectorDoomedReadDropOracleMutantIsCaught (11.73s)
    inspector_doomed_test.go:187: build go_test_-c__./internal/oracle_inspector-doomed-ReadDropOracle.test 5b4f4d711f27 hit 0.04
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
        --- FAIL: TestInspectorDoomedReadBehavior (10.50s)
            --- FAIL: TestInspectorDoomedReadBehavior/d1 (4.85s)
            --- PASS: TestInspectorDoomedReadBehavior/d1-control (0.14s)
            --- FAIL: TestInspectorDoomedReadBehavior/hoisted (4.45s)
            --- FAIL: TestInspectorDoomedReadBehavior/hoisted-uncaught (1.06s)
        FAIL
```

The nested failures are `TestInspectorDoomedBrokenAxiomStopsLoudly` under the axiom mutant, and `TestInspectorDoomedReadBehavior` with `d1`, `hoisted`, and `hoisted-uncaught` under the read-drop mutant. They all report the same missing sysroot. No sysroot was provisioned or WASI success claimed; these are environment failures, not compiler regressions established by this run.

The change’s own witness was also verified alone at HEAD after restoring the requested revision:

```bash
go test ./internal/oracle -run '^TestRecordStorageRoadsArePinned$' -count=1 -timeout 120m -v
```

Result: **exit 0, wall 8.193 seconds; all 11 subtests passed**, covering the fixed-object refusals and real-record controls. The log and timing are `/tmp/mez2-roads.log` and `/tmp/mez2-roads.time`. This passing witness does not erase the nine existing-contract failures above.

Preparation completed successfully before the full proof. `/workspace/adamic-tools/env.sh` was sourced first. `/tmp/adamic-gate` was created outside the repository because that environment file sets it as `TMPDIR`. `git fetch origin compiler/record-storage-roads-2`, detached checkout of `30379eb97798ece65ececd0693c2e0783fbbf6b7`, and `git submodule update --init --recursive` all exited 0. HEAD is exactly one commit after the supplied main revision. Submodule pins: `cohere` = `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` = `d92d9bfee114c80be2c375d72edae966176e3a4f`. The repository was clean before report creation.

Upstream `https://github.com/microsoft/TypeScript.git` was fetched into `/workspace/scratch/mez2-typescript` and checked out detached at exactly `050880ce59e30b356b686bd3144efe24f875ebc8`. `stage3/api/package.json` and its lockfile were copied to `/workspace/scratch/mez2-api`; `npm ci --prefix /workspace/scratch/mez2-api` exited 0 and installed the locked dependencies (`typescript` 6.0.3 and `@types/node` 25.3.3). All installations were outside the repository. Verified versions: TypeScript 6.0.3, Node v24.19.0, Go go1.27.1 linux/amd64, clang 20.1.8.

The proof and main comparisons exported the following after sourcing the environment file; npm setup used the same certificate settings. Child processes inherited these settings (the assignment helper’s separate `NODE_PATH` override is explained above):

```bash
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/mez2-typescript
export PATH=/workspace/scratch/mez2-api/node_modules/.bin:$PATH
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```

Two earlier attempts did not supply a usable test verdict. The initial timing wrapper failed before starting Go because `/usr/bin/time` was absent (launcher exit 127); it was replaced with Bash’s built-in timer. A subsequent foreground run was interrupted during the environment reconnect after roughly 37 minutes, leaving empty log and timing files and no captured proof exit status. It is not counted as a test result. The completed run was restarted under an external supervisor (`/tmp/mez2-oracle-run.sh`) so its log, exit and timing survived reconnects. Only one heavy top-level command ran at a time throughout setup, testing, and comparisons.

Report publication is restricted to `proof/lower-mez2-b/report.md` on new branch `proof/lower-mez2-b-report`, created directly from `30379eb97798ece65ececd0693c2e0783fbbf6b7`. No compiler or test files were edited. Only this report is to be committed and pushed, with the explicit destination `origin HEAD:refs/heads/proof/lower-mez2-b-report`; no other branch is a push target.
