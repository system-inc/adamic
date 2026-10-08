Built: separate tagged cast stops and checked-view read stops with source cast locations; complete observer erasure and cast-site verifier controls for step 09 #b5w3ycg.
Commits: library merge 88df481b (3c91ee02); compiler c8d40d60; this evidence commit.
Commands/results: strict verifier 20/20, all original 18 contracts pass; observer 40 controls; focused scanner/provenance tests and native/JS checked-view tests pass; Node WASI 10 tests pass.
Mutants: retained emitted check rejected; missing cast site rejected in both backends; omission of each new audit rejected; emitted-C guard omission caught; joined-helper origin omission caught in both backends.
Not covered: full 3,957 cast population or full gate; Error remains library #ddwcejg; global counts has 39 inherited failures and existing optional/tuple pins fail on baseline 080d978d too.

The runtime goldens, fixture classifications, exit codes and stdout remain unchanged. Tagged downcasts keep `cast failed:`. Checked-view fields use `field read failed:` plus `view cast at file:line:column`. Existing element-read pins remain `element read failed:`; generic acceptance permits field or element read stops and requires the cast location. Untagged acceptance requires the field-read prefix and the location computed independently from the fixture source using stock TypeScript.

Exact verifier assertion: verify.cjs:31. Its untagged predicate is runtime-contract.cjs:17. The missing-site receipt mutant is verify.cjs:45-49. It retains the field-read stop, exit and stdout, removes only the cast site, and must be rejected for native and JavaScript. Omitting the location predicate itself makes the verifier exit 1. The observer recognizes adamicCast, adamicCheckedViewCast and their panic read checks, removes the complete set, and rejects a control with one actual check left. Omitting that completeness audit makes the observer exit 1.

Provenance uses the existing allocation-flow graph and projection index. Joined and unknown origins report possible cast sites instead of inventing a unique origin. Diagnostic decoration leaves executable operands and original IR read markers intact. An exact joined-helper pin reports origins.a:4:12 and origins.a:7:12; its mutant retains the type check while dropping only provenance. Existing message pins retain their field/type/found/output requirements and add source-AST-validated context. Parent read context may precede a nested field path.

Library 3c91ee02 has parent f05aec3c, an ancestor of integration base 432d4913; the only incoming commit is 3c91ee02. Merged as 88df481b. Initial WASI checks lacked WASI_SYSROOT; cloud/setup.sh --wasi-sdk installed SDK 27, then ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestArrayHolesWASI' -count=1 -v passed under Node WASI.

Commands, each redirected to its evidence log:
- GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh. Setup 38.044s; SDK setup 48.150s; nproc 5, cpu.max 400000 100000.
- go build -o /tmp/step09-provenance-adamic ./cmd/adamic
- NODE_PATH=$PWD/stage3/api/node_modules node stage3/fixtures/checked-casts/observe.cjs /tmp/step09-provenance-adamic /tmp/step09-provenance-final-observe
- node stage3/fixtures/checked-casts/verify.cjs --require-runtime
- go test ./internal/oracle -run 'TestStep09|TestViewReadCastOrigins|TestScanner' -count=1 -v (pass 23.628s)
- go test ./internal/native -run 'TestCheckedView|TestView' -count=1 (pass 4.404s)
- go test ./internal/javascript -run 'TestCheckedView|TestView' -count=1 (pass 0.628s)
- go test ./internal/lower -run 'TestStep09|TestScanner|TestViewScalar|TestViewGeneric' -count=1 (pass 0.370s)
- go test ./internal/oracle -run TestScannerCastCounts -count=1 -v (pass; canonical counts unchanged)
- go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle (pass)
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts (fails 39 inherited rows; no successful global refresh claimed)
- Focused 90-function legacy pin selection: failures only optional read pins and tuple refusal, reproduced on detached baseline 080d978d; full log archived. Additional direct dictionary/callable/lazy pin selection passes. Baseline tuple expects a refusal that is already admitted; optional pins already expect 'is not' where runtime says 'matches no member'. These type pins were preserved.
- Native emitted-C mutant, receiver overlay mutants, observer audit omission and verifier site-predicate omission: commands and outputs in provenance logs and existing scripts.

Error stays pinned by TestScannerErrorLibraryCastNotYet to node:globals.ErrorConstructor.captureStackTrace at error-library-cast.a:3:1 (internal/oracle/scanner_constructor_views_test.go:48). No Error library facts were fabricated. String constructor checks remain green.

No protected emitter/lower/oracle files changed, no other view lane was merged, no cohere files were copied, no whole-package tests or full gate run. Historical DELIVERY.md and WORKER.md describe the earlier prefix migration; this report supersedes that diagnostic/verifier decision.
