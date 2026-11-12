# Oracle proof: lower both B

Tasks: #mez92e7 and #6jq01hc. Date: 2026-11-12 (UTC).

## Decisive result

The requested full oracle command completed with **exit 1**, with **three top-level failures, all environment failures caused by missing `WASI_SYSROOT`**. No non-environment failure was reported. This is not a fully green proof: WASI-dependent checks remain unverified in this environment.

- Tested HEAD: `d9c11411d498db77f81bca13454ca621c398bb12` (`compiler/record-storage-and-logical-views`).
- Its sole parent / requested main baseline: `ab6be657a4d947544674e883fc1c8bd718e68a1b`.
- Command: `go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1`.
- Exit: **1**.
- Wall time measured around the Go subprocess with Python `time.monotonic()`: **1381.832 seconds (23m 1.832s)**, including Go startup/build work.
- Go package summary: `FAIL github.com/system-inc/adamic/internal/oracle 1376.165s` (the original tab-separated line is retained below).
- Start: `2026-11-12T22:57:00.323967+00:00`.
- Finish: `2026-11-12T23:20:02.155800+00:00`.

## Every failing test and verbatim first errors

All three top-level failures below are **ENVIRONMENT**. `WASI_SYSROOT` was unset after loading `/workspace/adamic-tools/env.sh`. The stdout-columns test cannot reach its expected typed refusal because target setup first requires the sysroot. The two inspector mutant tests explicitly enable WASI in their child processes and fail while trying to build those WASI variants. Their nested failing test names and first error lines are also included verbatim. These are failures in mutant child processes, not additional failures of the unmutated top-level tests with those names.

```text
--- FAIL: TestStdoutColumnsRefusesWASI (0.42s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestInspectorDoomedAxiomMutantIsCaught (10.68s)
    inspector_doomed_test.go:171: build go_test_-c__./internal/oracle_inspector-doomed-Axiom.test fba221a084d7 hit 0.05
    inspector_doomed_test.go:178: TestInspectorDoomedRequiresThrowBeforeTrap goes red: inspector require must throw before a doomed use
    inspector_doomed_test.go:181: broken axiom must reach exact trap: exit status 1 === RUN   TestInspectorDoomedBrokenAxiomStopsLoudly
        === PAUSE TestInspectorDoomedBrokenAxiomStopsLoudly
        === CONT  TestInspectorDoomedBrokenAxiomStopsLoudly
            inspector_doomed_test.go:76: native: wasm32-wasi requires WASI_SYSROOT
        --- FAIL: TestInspectorDoomedBrokenAxiomStopsLoudly (4.54s)
        FAIL
--- FAIL: TestInspectorDoomedReadDropOracleMutantIsCaught (11.59s)
    inspector_doomed_test.go:187: build go_test_-c__./internal/oracle_inspector-doomed-ReadDropOracle.test 4973e3bb6f2c hit 0.05
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
        --- FAIL: TestInspectorDoomedReadBehavior (10.32s)
            --- FAIL: TestInspectorDoomedReadBehavior/d1 (4.89s)
            --- PASS: TestInspectorDoomedReadBehavior/d1-control (0.17s)
            --- FAIL: TestInspectorDoomedReadBehavior/hoisted (4.43s)
            --- FAIL: TestInspectorDoomedReadBehavior/hoisted-uncaught (0.82s)
        FAIL
```

The nested failures are `TestInspectorDoomedBrokenAxiomStopsLoudly`, `TestInspectorDoomedReadBehavior`, and its `d1`, `hoisted`, and `hoisted-uncaught` subtests. `d1-control` passed in the mutant child, as shown above.

No non-environment failures were reported, so there were **no qualifying tests to rerun alone on main**. No baseline rerun was performed, and no claim is made about main's test results. No separate WASI provisioning or follow-up suite run was performed.

## Setup completed

1. Loaded `/workspace/adamic-tools/env.sh` before repository preparation and test commands. It adds the provisioned Go/tool binaries and sets `GOTOOLCHAIN=auto` and `TMPDIR=/tmp/adamic-gate`. Created `/tmp/adamic-gate` outside the repository before running Go.
2. Successfully ran `git fetch origin compiler/record-storage-and-logical-views`, checked out the requested HEAD detached, and ran `git submodule update --init --recursive`. The tracked working tree was clean before testing and before writing this report.
3. Verified recursive submodule revisions:
   - `cohere`: `5360753e19e4417675a90e92a1823bea66b13918`.
   - `cohere/TypeScript`: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
4. Provisioned the distinct pinned **upstream** checkout outside the repository at `/workspace/adamic-tools/typescript-upstream`, fetching from `https://github.com/microsoft/TypeScript.git` and detaching at `050880ce59e30b356b686bd3144efe24f875ebc8`. Exported `ADAMIC_TYPESCRIPT_SOURCE` to this directory for the oracle command and its children.
5. Copied `stage3/api/package.json` and `package-lock.json` outside the repository to `/workspace/adamic-tools/stage3-api`, then successfully ran `npm ci` there. Prepended `/workspace/adamic-tools/stage3-api/node_modules/.bin` to `PATH`; its `tsc --version` returned `Version 6.0.3`.
6. Also successfully ran **`npm ci` inside `/workspace/adamic/stage3/api` itself**. This populated the exact repository-local `node_modules` path used by assignment-proof tests. Its `tsc --version` also returned `Version 6.0.3`. Both installs reported `added 3 packages`; the lock pins TypeScript `6.0.3` and `@types/node` `25.3.3`.
7. Exported these variables for both npm installs and the full oracle command, inherited by admission-gate and other children:

```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/adamic-tools/typescript-upstream
export PATH=/workspace/adamic-tools/stage3-api/node_modules/.bin:$PATH
```

Observed tool versions: `go version go1.27.1 linux/amd64`, Node `v24.19.0`, TypeScript `6.0.3`.

## Execution interruptions and limits

- The first timing wrapper could not launch Go because `/usr/bin/time` is absent (wrapper exit 127). This was a launch/setup error, not a failing oracle test. Timing was replaced with Python's monotonic clock.
- An initial actual suite run started at `2026-11-12T22:27:53.314403+00:00`, but environment/tool interruption occurred before Go returned any package output. Its Go parent and supervisor terminated without recording a suite result; that attempt is not counted as a completed proof.
- After readiness was restored, the suite was restarted under a detached supervisor with separate result metadata. A process check immediately afterward discovered that the interrupted attempt had left an orphaned `oracle.test` process. It was terminated and confirmed defunct. There was a brief unintended overlap during detection and cleanup; otherwise heavy commands were started sequentially. Only the supervised replacement remained active for the completed run reported above. This overlap means the one-heavy-command constraint was not perfectly maintained across the interruption.
- The full test log is `/tmp/oracle.log`, SHA-256 `8323003ca5839fffce07c7512bc8f43c2fbd6d6ad3df475f8d30cecf59c5f60d`. Completed-run metadata is `/tmp/adamic-proof/oracle-result.json`. These external artifacts are not committed.
- The non-verbose Go output does not enumerate passing or skipped tests; the supported conclusion is that no other failures were reported, not that every possible WASI case was exercised.

The final summary from the completed run is verbatim:

```text
FAIL
gate cache: native hits=2820 misses=5315
gate cache: node hits=1739 misses=6803
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	1376.165s
FAIL

```

## Report delivery

The report branch is `proof/lower-both-b-report`, created directly from tested HEAD `d9c11411d498db77f81bca13454ca621c398bb12`. The only file authorized for commit and publication is `proof/lower-both-b/report.md`; the only push destination is `origin` branch `proof/lower-both-b-report`. No compiler source, lockfile, installed dependency, or other proof artifact is included in the report commit.
