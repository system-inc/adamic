Proof for task #e2jp7my, compiler/tsc-version-zero-agrees-1

The requested agreement proof FAILS: earlier_static.a passes, but tsc_version_zero.ts panics at 128:7 on native, sanitized, and JavaScript while Node exits 0 with the expected four lines. The full lower suite passes. The full oracle suite fails with this compiler/runtime failure plus 13 environment failures, all 13 reproduced individually on main. The tsc_version_zero.ts oracle subtest does not exist on main, so the comparison cannot establish that its runtime behavior regressed in this commit.

The parameter-call mutant is red at the documented indirect call, 67:111. Because the unmodified witness already fails at runtime, this run is not a green-to-red mutation proof.

Preparation completed on 2026-11-11 in /workspace/adamic. Fetched origin compiler/tsc-version-zero-agrees-1, checked out 2478fed2bd8633edc501bae037e2bca9accc766d detached, and ran git submodule update --init --recursive successfully. Main comparison revision: 1bb9663d4a9cbf1466a758c41c2293c9c869e6ab. Submodules: cohere 5360753e19e4417675a90e92a1823bea66b13918; cohere/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f.

Loaded /workspace/adamic-tools/env.sh before setup and all proof commands. Created its missing work directory /tmp/adamic-gate outside the repository. Provisioned upstream https://github.com/microsoft/TypeScript.git at exactly 050880ce59e30b356b686bd3144efe24f875ebc8 in /tmp/adamic-proof/typescript; exported ADAMIC_TYPESCRIPT_SOURCE to that directory. Copied stage3/api/package.json and package-lock.json to /tmp/adamic-proof/api and ran npm ci there (3 packages installed); prepended its node_modules/.bin to PATH. Verified tsc Version 6.0.3, Node v24.19.0, and go1.27.1 linux/amd64.

For npm and all subsequent tests and children, exported NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt, npm_config_cafile=/etc/ssl/certs/ca-certificates.crt, and npm_config_strict_ssl=true. Heavy commands ran sequentially. No product source changes were retained. Only this report is committed, on proof/lower-e2j-1-report created from the requested HEAD; the only intended push destination is origin refs/heads/proof/lower-e2j-1-report.

1. Formatting and vet: PASS. gofmt -l internal/lower internal/oracle printed nothing. go vet ./internal/lower ./internal/oracle exited 0 with no diagnostics.

2. Focused oracle: FAIL, exit 1. Ran exactly:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStaticSelfConstructionAgreesWithNode' -count=1 -v -timeout 40m
```

Exactly two subtests ran: earlier_static.a PASS; tsc_version_zero.ts FAIL. Complete output, including every backend error:

```text
=== RUN   TestStaticSelfConstructionAgreesWithNode
=== PAUSE TestStaticSelfConstructionAgreesWithNode
=== CONT  TestStaticSelfConstructionAgreesWithNode
=== RUN   TestStaticSelfConstructionAgreesWithNode/earlier_static.a
=== PAUSE TestStaticSelfConstructionAgreesWithNode/earlier_static.a
=== RUN   TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts
=== PAUSE TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts
=== CONT  TestStaticSelfConstructionAgreesWithNode/earlier_static.a
=== CONT  TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts
    static_self_construction_test.go:33: native: exit codes differ, stderr "adamic: panic: predicate some at /workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:128:7 false branch: source readonly string[], target readonly string[]\n"
    static_self_construction_test.go:33: sanitized: exit codes differ, stderr "adamic: panic: predicate some at /workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:128:7 false branch: source readonly string[], target readonly string[]\n"
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 0b3bba574258 hit 197.76
=== NAME  TestStaticSelfConstructionAgreesWithNode/earlier_static.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 0b3bba574258 miss 199.28
=== NAME  TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts
    static_self_construction_test.go:33: javascript: exit codes differ, stderr "adamic: panic: predicate some at /workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:128:7 false branch: source readonly string[], target readonly string[]\n"
--- FAIL: TestStaticSelfConstructionAgreesWithNode (0.00s)
    --- FAIL: TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts (238.27s)
    --- PASS: TestStaticSelfConstructionAgreesWithNode/earlier_static.a (238.27s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=4
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	238.290s
FAIL
```

Ran the oracle's Node invocation exactly, from the repository root:

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/static_self_construction/tsc_version_zero.ts
```

Exit 0; complete output:

```text
0.0.0-0
1.2.3-beta.1+build.5
2.0.0
Debug Failure. False expression: Invalid argument: major
```

3. Parameter calls never known: RED, exit 1. Asserted in Python that the old text occurred exactly once in internal/lower/class_static.go, then replaced it once as specified by internal/lower/static-self-construction.mutants.md:

```text
old: if judged && target.Kind == ast.KindParameter
new: if false && target.Kind == ast.KindParameter
```

Ran exactly:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts' -count=1 -v
```

Complete output:

```text
=== RUN   TestStaticSelfConstructionAgreesWithNode
=== PAUSE TestStaticSelfConstructionAgreesWithNode
=== CONT  TestStaticSelfConstructionAgreesWithNode
=== RUN   TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts
=== PAUSE TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts
=== CONT  TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts
    static_self_construction_test.go:33: /workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:67:111: stage 0 can't lower an indirect call before static fields are initialized; move it after the field initializers yet
--- FAIL: TestStaticSelfConstructionAgreesWithNode (0.00s)
    --- FAIL: TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts (6.12s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=1
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	6.137s
FAIL
```

It stops lowering at 67:111, the indirect call, as the mutant table predicts. Restored internal/lower/class_static.go with git restore; git diff --exit-code then exited 0 with no output.

4. Counts census: PASS, exit 0. Ran go test ./internal/oracle -run 'TestCountsFilesCensus' -count=1 -v. The census does not report a missing counts file for the moved fixture.

```text
=== RUN   TestCountsFilesCensus
=== PAUSE TestCountsFilesCensus
=== CONT  TestCountsFilesCensus
--- PASS: TestCountsFilesCensus (0.06s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	0.071s
```

5. Full lower suite: PASS, exit 0. Wall time 1364.255 seconds (22m 44.255s). Ran go test ./internal/lower -count=1 -timeout 90m with stdout and stderr redirected to /tmp/lower.log. Every failing test: none, including no environment failures.

```text
ok  	github.com/system-inc/adamic/internal/lower	1342.585s
```

6. Full oracle suite: FAIL, exit 1. Wall time 3524.250 seconds (58m 44.250s). Ran go test ./internal/oracle -count=1 -timeout 180m with stdout and stderr redirected to /tmp/oracle.log. TestStaticSelfConstructionAgreesWithNode fails through its tsc_version_zero.ts child. There are 14 failing leaf tests: one compiler/runtime failure and 13 environment failures. All are listed below with the first actual error line (cache-status log lines are omitted), cause, and the individual main rerun result.

| Failing test | First error line | Cause | Individual rerun on main, wall time |
|---|---|---|---|
| `TestStdoutColumnsRefusesWASI` | `library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT` | Environment: WASI_SYSROOT is not configured; build stops before the expected typed refusal. | FAIL, exit 1; same missing WASI_SYSROOT; 38.078s |
| `TestAssignmentProofWiderTargetMutant` | `assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 3.008s |
| `TestAssignmentProofWiderTarget` | `assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.684s |
| `TestAssignmentProofCompoundMutant` | `assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.780s |
| `TestAssignmentProofCompound` | `assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.415s |
| `TestAssignmentProofChainedMutant` | `assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.365s |
| `TestAssignmentProofChained` | `assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.872s |
| `TestAssignmentProofReturnDefinedMutant` | `assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.728s |
| `TestAssignmentProofReturnDefined` | `assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.538s |
| `TestAssignmentProofIfNarrowingMutant` | `assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.448s |
| `TestAssignmentProofIfNarrowing` | `assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 3.158s |
| `TestAssignmentProofScannerKeywordMutant` | `assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 2.378s |
| `TestAssignmentProofScannerKeyword` | `assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520` | Environment: helper forces NODE_PATH to stage3/api/node_modules; external npm installation is outside that lookup path. Cannot find module 'typescript'. | FAIL, exit 1; same module-resolution error; 4.761s |
| `TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts` | `static_self_construction_test.go:33: native: exit codes differ, stderr "adamic: panic: predicate some at /workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:128:7 false branch: source readonly string[], target readonly string[]\n"` | Compiler/runtime: predicate some false-branch panic at 128:7, also on sanitized and JavaScript. Node exits 0. Not an environment failure. | Absent; exit 0 with [no tests to run], not a passing equivalent; 2.720s |

The twelve assignment-proof tests fail before they reach their intended oracle or mutant checks. internal/oracle/assignment_proofs_test.go:73 overwrites NODE_PATH with /workspace/adamic/stage3/api/node_modules. The requested external installation supplies tsc on PATH, but that hard-coded module lookup does not reach /tmp/adamic-proof/api/node_modules. This is the same environment limitation on both revisions. The WASI test likewise stops at missing WASI_SYSROOT on both revisions.

The compiler/runtime failure has a different cause: execution reaches the predicate check at tsc_version_zero.ts:128:7 and panics on its false branch even though both printed types are readonly string[]. All three backends disagree with Node's successful execution. This is an observed runtime predicate failure; no source-level root-cause fix was attempted.

All 14 failing leaf selections were rerun separately and sequentially while checked out detached at main 1bb9663d4a9cbf1466a758c41c2293c9c869e6ab, in the same environment. Each command used go test ./internal/oracle -run with every slash-separated name component anchored by ^ and $, followed by -count=1 -v -timeout 180m. Exact commands and results:

```text
go test ./internal/oracle -run '^TestStdoutColumnsRefusesWASI$' -count=1 -v -timeout 180m
exit=1; wall_seconds=38.078
go test ./internal/oracle -run '^TestAssignmentProofWiderTargetMutant$' -count=1 -v -timeout 180m
exit=1; wall_seconds=3.008
go test ./internal/oracle -run '^TestAssignmentProofWiderTarget$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.684
go test ./internal/oracle -run '^TestAssignmentProofCompoundMutant$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.780
go test ./internal/oracle -run '^TestAssignmentProofCompound$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.415
go test ./internal/oracle -run '^TestAssignmentProofChainedMutant$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.365
go test ./internal/oracle -run '^TestAssignmentProofChained$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.872
go test ./internal/oracle -run '^TestAssignmentProofReturnDefinedMutant$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.728
go test ./internal/oracle -run '^TestAssignmentProofReturnDefined$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.538
go test ./internal/oracle -run '^TestAssignmentProofIfNarrowingMutant$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.448
go test ./internal/oracle -run '^TestAssignmentProofIfNarrowing$' -count=1 -v -timeout 180m
exit=1; wall_seconds=3.158
go test ./internal/oracle -run '^TestAssignmentProofScannerKeywordMutant$' -count=1 -v -timeout 180m
exit=1; wall_seconds=2.378
go test ./internal/oracle -run '^TestAssignmentProofScannerKeyword$' -count=1 -v -timeout 180m
exit=1; wall_seconds=4.761
go test ./internal/oracle -run '^TestStaticSelfConstructionAgreesWithNode$/^tsc_version_zero\.ts$' -count=1 -v -timeout 180m
exit=0; wall_seconds=2.720
```

The exact main selection for the newly added subtest produces no test execution, not an equivalent pass:

```text
=== RUN   TestStaticSelfConstructionAgreesWithNode
=== PAUSE TestStaticSelfConstructionAgreesWithNode
=== CONT  TestStaticSelfConstructionAgreesWithNode
--- PASS: TestStaticSelfConstructionAgreesWithNode (0.00s)
testing: warning: no tests to run
PASS
ok  	github.com/system-inc/adamic/internal/oracle	0.013s [no tests to run]
```

Also ran the failing parent test itself on main, go test ./internal/oracle -run '^TestStaticSelfConstructionAgreesWithNode$' -count=1 -v -timeout 180m. It exits 0 and runs only earlier_static.a:

```text
=== RUN   TestStaticSelfConstructionAgreesWithNode
=== PAUSE TestStaticSelfConstructionAgreesWithNode
=== CONT  TestStaticSelfConstructionAgreesWithNode
=== RUN   TestStaticSelfConstructionAgreesWithNode/earlier_static.a
=== PAUSE TestStaticSelfConstructionAgreesWithNode/earlier_static.a
=== CONT  TestStaticSelfConstructionAgreesWithNode/earlier_static.a
    static_self_construction_test.go:33: gate cache node miss
    static_self_construction_test.go:33: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 0b3bba574258 hit 0.41
--- PASS: TestStaticSelfConstructionAgreesWithNode (0.00s)
    --- PASS: TestStaticSelfConstructionAgreesWithNode/earlier_static.a (8.95s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	8.979s
```

Returned to 2478fed2bd8633edc501bae037e2bca9accc766d after the comparisons. git diff --exit-code exited 0, and git status --short was empty before creating this report.

Full oracle failure output follows (the preceding 78 identical native cache-bypass notices are informational: "native: unit main.c compiles uncached: main.c holds a directive past its first line").

```text
--- FAIL: TestStdoutColumnsRefusesWASI (0.75s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestAssignmentProofWiderTargetMutant (0.06s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.07s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.06s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.06s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.08s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.06s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.05s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.06s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.05s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.08s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.06s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.07s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestStaticSelfConstructionAgreesWithNode (0.00s)
    --- FAIL: TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts (5.51s)
        static_self_construction_test.go:33: gate cache node miss
        static_self_construction_test.go:33: native: exit codes differ, stderr "adamic: panic: predicate some at /workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:128:7 false branch: source readonly string[], target readonly string[]\n"
        static_self_construction_test.go:33: sanitized: exit codes differ, stderr "adamic: panic: predicate some at /workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:128:7 false branch: source readonly string[], target readonly string[]\n"
        static_self_construction_test.go:33: gate cache node miss
        oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 0b3bba574258 hit 0.04
        static_self_construction_test.go:33: javascript: exit codes differ, stderr "adamic: panic: predicate some at /workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:128:7 false branch: source readonly string[], target readonly string[]\n"
FAIL
gate cache: native hits=3046 misses=6493
gate cache: node hits=1741 misses=6503
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	3515.509s
FAIL
```

Wall times include the go command's build/startup overhead and were measured with Python time.monotonic around subprocess.run. An initial timing wrapper attempt could not start the lower suite because /usr/bin/time is absent (exit 127); it was immediately replaced by the monotonic wrapper, and the complete requested lower command then ran once to completion. This wrapper issue is not a test failure. The full-suite logs remain at /tmp/lower.log and /tmp/oracle.log; focused, mutant, census, and main rerun logs remain under /tmp/adamic-proof. No full suite was rerun on main: only failing tests and the failing oracle parent were run there.
