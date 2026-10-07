Built: rebased the existing lint unit onto origin/main e011f8f6; no new helper reserved.
Commits: d6b4f388 is the tested code tip; this report and logs follow it.
Checks: owned fixture/core/refusal oracles pass, uncached input oracle passes, registry passes; shared TestRulesAgree fails.
Mutants: all seven compiling semantic mutants caught solely by output comparison on Node source, emitted JavaScript and sanitized native.
Not covered: full landing readiness, new corpus replay, full repository gate, live Tailwind program resolution, multiple automatic edits and arbitrary configured variable regexes.

The only pushed branch owned by this unit is codex/lint-wave1-01. A local backup preserves f7b92cb3. Current main is an ancestor of the rebased tip. Rebase conflicts were resolved with exact previously committed shared harness versions. The automatic merge duplicated the Settings import in main.ts; restoring the exact previously committed entry point removes that duplication without authoring a new shared API. Main supplies the legacy volume/comments/messages modules missing before the rebase. Rule-directory source files are unchanged by the rebase.

Setup: bash cloud/setup.sh, Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, cache warm 111s, total 111s, nproc 5.

Validation commands (all output redirected to the retained logs):

    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/lint/rules/typescript-consistent-type-assertions ./stage1/cohere/lint/rules/structure-tailwind-no-physical-direction ./stage1/cohere/lint/rules/better-tailwindcss-no-deprecated-classes -run 'TestAssertions$|TestRuleContracts$|TestDecisionCores$|TestListeners$|TestRefusals$' -count=1 -timeout 15m -v
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 10m -v
    go test ./stage1/cohere/lint/registry -count=1 -timeout 10m -v
    go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -timeout 15m -v

Assertions PASS 72.810s: 196 projected fixtures, 193 independent non-JSX sources, 44,254 output bytes/backend; equal-precedence mutant killed on all three sides. Prior three-rule suite PASS 104.116s: 179 projected fixtures, 161 independent sources, 58,087 projected and 52,052 independent output bytes; RTL/LTR exemption, description-reason exemption and display=block mutants all killed on three sides. Tailwind suite PASS 89.105s: 1,001 real Go core observations, 184,433 bytes/backend, 51 supported listener fixtures (three multiple-edit fixtures excluded), 17,695 finding/fix/first-plan bytes, five independent sources; version-gate, duplicate-guard and nil-system mutants all killed on three sides. Refusal tests pass for all three missing providers. These are scoped contracts, not a full rule certification for the last batch.

Uncached input oracle PASS 3.990s, six probe misses. Registry PASS 0.285s. Shared oracle FAIL 84.726s after capturing 1,699 unique cases: it attempts to decode a no-labels allowLoop configuration as @eslint-community/eslint-comments/require-description, then panics with json: unknown field "allowLoop". The stack points to shared collect calling the owned strict upstream decoder. Do not weaken Go's decoder or suppress the failure to obtain a green run. Shared routing is outside this unit's authorized rule directories. It has not been repaired.

Mutants, in order: equal precedence changed from > to >=; RTL/LTR exclusion removed; description-reason exclusion removed; display=block allowed; deprecation major-version boundary tightened; duplicate first-occurrence guard inverted; nil-system guard reversed. Each compiles and executes successfully before comparison catches the output change. The unknown-class mutant certifies only the nil-system/exemption core, not the missing live resolver.

Fresh helper ownership audit read all claims on 430 origin refs (17 distinct blobs), plus the existing HELPERS.md comment bundle reservation. Unclaimed six-consumer helpers remain, including loadDesignSystemThrough. No helper claim was pushed because the landing-first prerequisite is not fulfilled by a red shared oracle. No helper implementation or rule-ready count is claimed.

## Second landing refresh

Rebased cleanly onto newly fetched origin/main e8ba3d5d. Tested code tip: 3675d5de0371e3c3f51e8c6c6524f283507703df. No rule source changed and no shared file was edited. Only codex/lint-wave1-01 is pushed by this unit; the backup and work branches are local. Nothing is pushed to main or area branches.

Repeated the same exact commands above, with output retained as landing2 logs. Assertions PASS 90.591s; prior three-rule contracts PASS 114.941s; Tailwind cores/listeners/refusals PASS 100.493s. All seven mutants still compile/run and are caught only by comparison on source Node, emitted JavaScript and ASan/UBSan native. Input oracle PASS 3.928s with six probe misses; registry PASS 0.293s. Setup: Go/clang/Node/submodules ready 0s, warm cache and total 134s, nproc=5.

Shared TestRulesAgree FAIL 88.471s, same allowLoop strict-decoder failure after capturing 1,699 cases. This does not establish landing readiness. Previous multiple-edit/live-program/regex/independent-JSX/convergence limits remain. No complete corpus or full repository gate was repeated. No new helper is claimed while the landing-first gate remains red. Helper README reread and the complete readiness JSON parsed (198 frozen cohort, 46 initial helper-ready); selection remains deferred.

## Refresh onto f8013f0ba

Rebased cleanly onto origin/main f8013f0ba; tested code tip 540f11744f62264ea85589750aca364935e7c24d. No lint source differs from the previously pushed 64e1fdf8. No helper claim or implementation is added. Existing limitations remain; the shared harness is still f4d98cab and batch 8 Diagnostic SHA has not been supplied.

Repeated commands, all redirected to retained refresh logs:

    bash cloud/setup.sh
    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/lint/rules/better-tailwindcss-no-deprecated-classes -run '^TestNumericListeners$|^TestDecisionCores$|^TestListeners$|^TestRefusals$' -count=1 -timeout 10m -v
    go test ./stage1/cohere/lint/rules/typescript-consistent-type-assertions ./stage1/cohere/lint/rules/structure-tailwind-no-physical-direction -run '^TestAssertions$|^TestRuleContracts$' -count=1 -timeout 10m -v
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 10m
    go test ./stage1/cohere/lint/registry -count=1 -timeout 10m

PASS: numeric/Tailwind package 112.066s, assertions 73.704s, original three-rule suite 126.531s, uncached compiler oracle 18.193s, registry 0.239s. Seven subscription and seven rule/core compiling mutants are caught only by comparison on source Node, emitted JavaScript and sanitized native. Assertions compare 44,254 bytes, prior three rules 58,087 projected/52,052 independent bytes, Tailwind core 184,433 bytes and supported listeners 17,695 bytes per side; numeric declarations 329 bytes. Full unknown-class rule remains uncertified.

Setup: Go/clang/Node/submodules ready 0s; cache warm and total 121s; nproc=5. Full shared oracle, corpus and repository gate not repeated. The previously observed shared allowLoop strict-decoder failure remains unresolved in unchanged harness code; do not call the branch fully landing-ready. Handed-node callbacks and numeric runtime kinds are still absent from the shared API. Push only codex/lint-wave1-01 with a lease against its verified old tip. No main or area push.

## Parked under Ahra instruction

The explicit parking exception authorizes this unit to move to helpers. Own branch is pushed and based on current origin/main f8013f0ba; owned rule/core/numeric oracles and all fourteen mutants are green (refresh logs). Shared allowLoop option routing, multiple automatic edits, live-program/regex adapters and handed-node/numeric runtime API gaps are named above. This is scoped green plus a named shared blocker, not a full rule certification. No batch 8 Diagnostic SHA has been supplied. Rebase and revalidate this parked branch when the named harness SHA lands, before taking further work.

## Parked refresh onto c01907a70

Rebased cleanly onto current origin/main c01907a70. Main changes, including developer-tool leak-check changes where present, are retained; no shared files reverted. Owned assertion contracts PASS 48.334s, original three-rule contracts PASS 71.036s, numeric/Tailwind contracts/refusals PASS 60.575s. All fourteen semantic/subscription mutants remain comparison-only catches on all three backends. Setup ready timings 0s, cache warm and total 55s, nproc 5. Shared harness blockers remain named above; no supplied batch 8 Diagnostic SHA. Parked under the explicit exception, full shared/corpus gate not repeated. New CFG helpers contain no regex; future regex ports follow the new shared RegExp translation rule.

## Named harness landing attempt

Fetched every origin branch. Current main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, already an ancestor of both owned branches. Attempted to merge the supplied ab70f38d47de1d4974082b38f84a56af2368b7af into codex/lint-wave1-01. Git reports ten unowned conflicts: docs/lint-registration.md, context.ts, harness_test.go, lint.ts, lint_test.go, main.ts, profile_test.go, registration_test.go, registry/registry.go and testdata/oracle.go. The lint paths are under stage1/cohere/lint. The merge was aborted without editing any shared file or reverting developer-tool changes. The named harness is not incorporated. Integration must resolve those shared conflicts. No new helper claim is made while this refresh is blocked.

Revalidated the existing parked tree with the scoped rule command above, additionally selecting TestNumericListeners. Assertions PASS 23.129s, original three-rule suite and numeric/Tailwind suite PASS; all fourteen compiling semantic/subscription mutants are caught only by byte comparison on source Node, emitted JavaScript and sanitized native. Uncached TestInputAgreesWithNode PASS 0.971s, six probe misses. Logs are named-harness-owned.log and named-harness-input.log. Setup ready lines are all 0s, cache warm and total 33s, nproc 5. Full shared oracle, complete source corpus and repository gate were not repeated. Existing live Tailwind, multi-edit and configured-regex refusals remain explicit; this run certifies the supported contracts only.

## Parked refresh onto b8fb957aa

Rebased cleanly onto current origin/main b8fb957aa. Main inherited-static-field and developer-tool changes retained; no shared conflict edit or revert. Repeated the scoped owned rule command with TestNumericListeners selected: PASS typescript-consistent-type-assertions 24.617s; structure-tailwind-no-physical-direction 46.634s; better-tailwindcss-no-deprecated-classes 35.683s. Fourteen compiling semantic/subscription mutants remain byte-comparison-only catches on source Node, emitted JavaScript and sanitized native. All seven rule.json subscriptions already use the registry's ast.Kind names. Numeric sidecar tests are historical helper-level Go enum checks, not a new registry API; no numeric stage1-parser dispatch claimed. Named ab70f38d4 integration still has the ten unowned shared-file conflicts recorded above; parked under Ahra's explicit exception. Full shared/corpus/repository gate not repeated; prior live-provider/multi-edit/configured-regex limitations remain explicit. Helper branch refreshed and pushed at 51ea33f1c before any new reservation. Setup ready 0s, warm and total 79s, nproc 5.
