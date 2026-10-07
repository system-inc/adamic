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
