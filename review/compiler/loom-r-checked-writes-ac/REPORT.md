Built compiler/loom-r-checked-writes-ac from pinned main, carrying the complete 5aa205ec change plus first-line header metadata.
Source: 5aa205ec00329ce78789ddb62277f0ccdcc1f337. Base: dbdf977f5b8b0979d414ebe6f20e7aadd312e309. Delivery is one commit; the final response gives its SHA.
Real branch compiler and unchanged Gate.aCheck accept all 28 changed .a files: 5 checked, 1 NotYet, 22 refused, 0 checker errors.
Removed-header and wrong-reason mutants are both rejected by Gate.aCheck; 193 focused roots, guards, vet and 12 direct count checks pass.
Not covered: whole repository a-check, whole-package/full gate, ahra, other platforms, or a new run of the 19 historical standalone compiler mutants. No PR or Loom submission.

Read CLAUDE.md, docs/front-door.md and stage3/a-check-headers/README.md. Fetch origin, checkout -b compiler/loom-r-checked-writes-ac dbdf977f5b, then git cherry-pick --no-commit 5aa205ec00329ce78789ddb62277f0ccdcc1f337. The exact pinned base takes precedence over the generic current-main instruction. No compiler or runtime line differs from 5aa205ec. No cohere source was copied.

Headers and scope:

The complete scope is git diff --name-only --diff-filter=AM dbdf977f5b 5aa205ec -- '*.a', including review evidence: 28 programs. paths.json is the exact inventory. Each was compiled in place with the branch's own go-built adamic-acheck, using its c command, with a hard 30-second per-file bound. The final scan repeats actual compilation after the header edits. Every full diagnostic, exit and body SHA-256 is recorded in status.json; initial and final stderr/C outputs have individual logs.

All 22 refused files already had matching abbreviated refusal headers in 5aa205ec. No missing header was discovered. The conservative choice is to replace each with the FULL actual refusal reason, rather than depend on a short substring. headers.json lists every old and new first line. No checker-error header is needed. Remove the one legacy // a-check: compile header from checked stage3/checked-writes/proven-generic.a, following the instruction that clean programs take no header. Net: 22 refusal headers refreshed, one clean header removed, zero newly inserted error-header lines. There are no program body edits and no new .a files beyond the original change.

The five checked programs are four proven stage3 programs and the review push-loop witness. internal/native/testdata/checked_library_push.a is NotYet: stage 0 can't lower a SpreadElement yet (line 7, column 14). It remains without a header. Gate.aCheck groups this NotYet with checked, so its own report has 6 checked and 22 refused; this report separates the requested four statuses.

First-line body integrity and positions:

Every edited file's bytes below its original gate metadata are unchanged. All 22 refusal headers replace one line with one line, preserving diagnostic line positions. The only line shift is proven-generic.a, minus one line from removing its compile header. The path/basename search covers *_test.go, testdata and Markdown reports: path-references.json. Its two test readers are internal/oracle/checked_views_next_test.go TestCheckedViewsNextProvenAdamic and internal/oracle/checked_writes_test.go's proven fixture/count registration. The report mention in stage3/checked-writes/next-views-report.md asserts no inserted checks and pins no position. Neither those readers nor the other report/count references pin a source line. No expectation or report correction is needed. Exact-path refusal sidecars still match and their focused tests pass.

body-audit.log and emitted-c-identity.json prove all five checked .a outputs and all eleven TypeScript adapter C outputs are byte-identical before/after. The adapters are temporary copies of the existing source bytes for the existing TypeScript-mode boundary, not authored or committed TypeScript programs. The body hashes and full first-line changes are in headers.json/status.json.

A-check predicate and its mutants:

The historical README tooling hashes did not expose cloud/fast-gate/run.py in the local objects. Fetch the existing devtools/fast-gate branch read-only and use its exact unchanged Gate.aCheck at 348ac0ff8936d366936238165e53967af6a4b93e. gate.py retains that source; SHA-256 b1a2e35d19a98d2d8e8ebd36a0f681037ed04dcddd325bc7d49b55b5e6e093cc. No predicate edit or reimplementation is used. The harness reuses the separately built branch compiler, redirects only its executable path, and otherwise executes Gate.aCheck's own classification, first-line reading, expectation matching and failure path.

Final actual scan: exit 0, 28/28 accepted. On internal/oracle/testdata/checked-runtime-fields_refused/directory.a, capture the real branch compiler diagnostic, then independently:

- Remove its header: unchanged Gate.aCheck exits 1, refused but expected checked.
- Replace the reason with deliberately unrelated header mutant: unchanged Gate.aCheck exits 1, the refusal does not match that reason.

This is the README's predicate-only proof over captured real diagnostics. It does not claim a fresh compiler invocation for each comment mutant. Both mutants change the real first line and leave the body intact, restore the file in finally, and are followed by a passing restored predicate over all 28 files. header-mutants.json records both exact failure messages and catchers. check-headers.py is the complete harness. No mutant program is left in the tree.

Verification commands, all output redirected directly to the named logs:

- export GOPROXY='https://proxy.golang.org|direct'; timeout 180 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh; export GOMAXPROCS=4.
- timeout 120 go build -o review/compiler/loom-r-checked-writes-ac/adamic-acheck ./cmd/adamic: exit 0. Compiler SHA-256 19ee1120d9a2d88105b93d222539403f904f36e8cc35c98c09eaac50bf669f3a.
- timeout 180 python3 review/compiler/loom-r-checked-writes-ac/check-headers.py: PASS actual Gate.aCheck 28 files; both mutants caught; restored predicate passes.
- timeout 120 python3 review/compiler/loom-r-checked-writes-ac/audit-bodies.py: PASS 23 bodies, 16 emitted C identities, unchanged counts.md and path-pin audit.
- timeout 120 go test ./cmd/adamic -run '^TestProduct_CommandRuntimeSanitized$' -count=1 -json -timeout=90s: PASS, leaf 1.14s.
- timeout 600 python3 review/compiler/loom-r-checked-writes-ac/run-focused.py: all 193 original focused roots pass in 14 bounded invocations, 203 events, maximum 4.01s. The exact inventory and every command/test/time are in focused-tests.json and focused-results.json. This reruns every path-reading oracle/native fixture test, its Node truth, sanitized/release native and JavaScript comparisons, exact refusals/messages, and the 45 embedded mutant witnesses. The inherited report retains the prior 19 standalone compiler-mutant catches; those standalone overlays were not rerun for comment-only changes.
- timeout 120 go test ./internal/ir -run '^(TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted)$' -count=1 -parallel=4 -json -timeout=90s: PASS; reader guard 6.37s.
- timeout 120 go test ./internal/lower -run '^TestNativeAgreementRowsNameRegisteredPrograms$' -count=1 -json -timeout=90s: PASS 0.01s.
- timeout 120 go test ./internal/oracle -run '^Test(RuntimeSetsAreEachUsed|CountsAreRecordedTable|FamiliesOwnEveryCaseOnce|FamilyWrappersAreCurrent)$' -count=1 -parallel=4 -json -timeout=90s: all four PASS, maximum 0.94s.
- timeout 120 go test ./cmd/adamic-gate -run '^TestEveryTestIsParallelOrSaysWhy$' -count=1 -json -timeout=90s: PASS 0.23s.
- timeout 180 go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle ./internal/regexp ./stage1/cohere/values ./stage3/fixtures: exit 0.

Counts were NOT regenerated. counts.md is byte-identical to 5aa205ec. Beyond the 16 C identities, fresh measurements of the 11 header-edited TypeScript adapters and proven-generic.a agree with every recorded counter:
  timeout 120 go test -overlay review/compiler/loom-r-checked-writes-ac/counts-overlay.json ./internal/oracle -run '^TestHeaderCounts' -count=1 -parallel=4 -json -timeout=90s
PASS all 12 probes, package 2.617s, maximum leaf 2.60s. These reuse existing prepare/measure helpers through an evidence-only overlay, never modify production tests or the count table, and each calls t.Parallel first. counts-probe.go.txt and counts-keys.json record the probe source and keys. No new compiler-consuming package or runtime set is introduced. All completed test leaves are below 60 seconds on the four-CPU quota.

Individual count probe times:

| Probe | Seconds |
| --- | ---: |
| TestHeaderCounts10 | 1.49 |
| TestHeaderCounts06 | 1.5 |
| TestHeaderCounts00 | 1.53 |
| TestHeaderCounts08 | 0.21 |
| TestHeaderCounts07 | 0.21 |
| TestHeaderCounts09 | 0.27 |
| TestHeaderCounts05 | 0.22 |
| TestHeaderCounts03 | 0.28 |
| TestHeaderCounts04 | 0.23 |
| TestHeaderCounts01 | 0.17 |
| TestHeaderCounts02 | 0.19 |
| TestHeaderCounts11 | 2.6 |

Setup observations:

Linux x86_64. nproc=5, cpu.max=400000 100000 (four CPUs). Go 1.27.1, Node 24.19.0, clang 20.1.8. setup.log retains all timing lines: go 0.085s; node 0.087s; submodules 0.088s; markdown dependencies 0.202s (step 0.013s); clang 0.434s; shared cache 0.934s; Go build 22.838s; test binaries deferred 22.926s; build cache warm 22.927s; done 22.953s. Setup succeeds without a workaround. No unexpected test failure, test timeout or over-budget leaf occurred in this header pass.

Per-file actual status table follows. Full refusal strings and unchanged body hashes are also machine-readable in status.json. Compiler, runtime, all .refused/.stderr expectations, original Go tests and counts are unchanged from 5aa205ec; only the 23 first-line edits above and new review evidence are added to that change.

| File | Status | Actual reason/code |
| --- | --- | --- |
| internal/native/testdata/checked_library_push.a | NotYet | adamic: /workspace/adamic/internal/native/testdata/checked_library_push.a:7:14: stage 0 can't lower a SpreadElement yet  |
| internal/oracle/testdata/checked-runtime-fields_refused/directory.a | refused | a value of type Result[] seen as { names: readonly string[]; }[], whose readonly field names becomes writable: a readonly field may hold something narrower than readonly string[], which a write of readonly string[] would replace; keep names readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-runtime-fields_refused/fs.a | refused | a value of type Result[] seen as { text: string; }[], whose readonly field text becomes writable: a readonly field may hold something narrower than string, which a write of string would replace; keep text readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-runtime-fields_refused/map.a | refused | a value of type Result[] seen as { done?: boolean \| undefined; value: string \| undefined; }[], whose readonly field done becomes writable: a readonly field may hold something narrower than boolean \| undefined, which a write of boolean \| undefined would replace; keep done readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-runtime-fields_refused/regex.a | refused | a value of type Result[] seen as { done?: boolean \| undefined; value: readonly (string \| undefined)[] \| undefined; }[], whose readonly field done becomes writable: a readonly field may hold something narrower than boolean \| undefined, which a write of boolean \| undefined would replace; keep done readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-runtime-fields_refused/set.a | refused | a value of type Result[] seen as { done?: boolean \| undefined; value: string \| undefined; }[], whose readonly field done becomes writable: a readonly field may hold something narrower than boolean \| undefined, which a write of boolean \| undefined would replace; keep done readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-runtime-fields_refused/status.a | refused | a value of type Result[] seen as { type: string; size: number; symbolicLink: boolean; }[], whose readonly field type becomes writable: a readonly field may hold something narrower than string, which a write of string would replace; keep type readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-write-messages_refused/02_shared-empty_out.a | refused | a value of type never[] seen as number[], which can write number where never is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-write-messages_refused/03_flow-node_out.a | refused | a value of type FlowArrayMutation seen as FlowNodeBase, which can write BinaryExpression \| BindingElement where BinaryExpression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-write-messages_refused/13_declaration-array_out.a | refused | a value of type Declaration[] seen as Node[], which can write Node where Declaration is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-write-messages_refused/15_detached-diagnostic_out.a | refused | a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| internal/oracle/testdata/checked-write-messages_refused/16_flow-assignment-union_out.a | refused | a value of type FlowArrayMutation seen as FlowNodeBase, which can write BinaryExpression \| BindingElement where BinaryExpression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| review/compiler/checked-writes-follow/push-loop.a | checked |  |
| stage3/checked-writes/array-refused.a | refused | a value of type (1 \| 2)[] seen as number[], which can write number where 1 \| 2 is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| stage3/checked-writes/emit-refused.a | refused | a value of type GeneratedIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace; keep emitNode readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
| stage3/checked-writes/flags-refused.a | refused | a value of type T seen as Mutable<T>, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of T["flags"] would replace; keep flags readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| stage3/checked-writes/flow-refused.a | refused | a value of type FlowMutation seen as FlowNode, which can write Expression where BinaryExpression \| CallExpression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| stage3/checked-writes/map-refused.a | refused | a value of type Map<string, 1 \| 2> seen as Map<string, number>, which can write number where 1 \| 2 is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| stage3/checked-writes/proven-containers.a | checked |  |
| stage3/checked-writes/proven-emit.a | checked |  |
| stage3/checked-writes/proven-generic.a | checked |  |
| stage3/checked-writes/proven-number.a | checked |  |
| stage3/checked-writes/refused-branch-bottom.a | refused | a value of type number[] seen as number[], which can write number where never is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| stage3/checked-writes/refused-branch-diagnostic.a | refused | a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| stage3/checked-writes/refused-optional-boolean.a | refused | a value of type { readonly multiLine: true \| undefined; } seen as { multiLine: boolean \| undefined; }, whose readonly field multiLine becomes writable: a readonly field may hold something narrower than true \| undefined, which a write of boolean \| undefined would replace; keep multiLine readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
| stage3/checked-writes/refused-overload-result.a | refused | overload 1 of createTempVariable result GeneratedIdentifier cannot be served by implementation result Identifier; make the implementation result covariant with every overload result |
| stage3/checked-writes/shared-never.a | refused | a value of type never[] seen as number[], which can write number where never is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| stage3/checked-writes/string-refused.a | refused | a value of type { readonly text: "left" \| "right"; } seen as { text: string; }, whose readonly field text becomes writable: a readonly field may hold something narrower than "left" \| "right", which a write of string would replace; keep text readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable) |
