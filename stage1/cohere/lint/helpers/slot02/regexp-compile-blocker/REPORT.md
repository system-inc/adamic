Reproduced the blocker for the next highest unclaimed concrete helper, regexp.Compile; no new helper claimed or credited.
Prior completed work remains pushed at 238e54a2, containing current main 71d7e491 and lint area b28757f3.
Setup PASS 46.933s, nproc 5; dynamic RegExp Node prints yes, Adamic exits 1 with explicit NotYet; constant control passes sanitized native.
The unchanged TestRegExpNativeRefusals passes; replacing the dynamic pattern with a literal removes the refusal. This is a refusal control, not a semantic helper mutant.
Stopped at the shared dynamic RegExp gap; no shared files changed and no full gate rerun.

Landing audit: git fetch explicitly refreshed main, area/stage1-lint and all twenty origin codex/lint-helpers* branches. Both integration tips are unchanged ancestors of the published 238e54a2 branch. The prior oracle selection remains current: batch13 helper PASS 209.265s, inherited rules/compiler/link-only selection PASS 298.899s, twelve compiling semantic mutants caught. All thirty-nine existing helper claims are complete and pushed. Only this branch has been pushed by this worker.

Selection audit: read readiness.json and nineteen claim files from all twenty helper branches in full. Exact symbol boundary matching avoids prefix confusion. The five comments helpers with twenty-four consumers are already owned and delivered by codex/lint-helpers in HELPERS.md; they are excluded from new work. regexp.Compile is the remaining highest-count concrete symbol, with four consumers: @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Other unclaimed concrete helpers have at most three consumers. Full captured claim texts and ranking are in evidence/claims-and-ranking.json.gz. No claim was pushed because the dependency check found a shared-toolchain blocker before implementation.

Reproducer: testdata/dynamic_pattern.a takes a runtime string and calls new RegExp(pattern, 'u'), as required for options patterns. Node 24.19.0 prints yes and exits zero. Adamic reports stage 0 can't lower RegExp with a nonconstant pattern yet and exits one. internal/lower/regexp.go explicitly requires constantPattern to resolve constructor arguments. Actual cohere Compile accepts runtime source/flags and returns compilation errors as values; a fixed pattern is not its contract. Delegating the entire compile operation to an external callback would leave the helper unbuilt. No handwritten matcher or shared compiler workaround was added.

Commands after source /workspace/adamic-tools/env.sh, each writing directly to its own log:

```sh
bash cloud/setup.sh > /tmp/slot02-regexp-blocker-setup.log 2>&1
nproc > /tmp/slot02-regexp-blocker-nproc.log
node --input-type=module-typescript < stage1/cohere/lint/helpers/slot02/regexp-compile-blocker/testdata/dynamic_pattern.a > /tmp/slot02-regexp-node.log 2>&1
go run ./cmd/adamic build stage1/cohere/lint/helpers/slot02/regexp-compile-blocker/testdata/dynamic_pattern.a -o /tmp/slot02-regexp-dynamic --sanitize > /tmp/slot02-regexp-dynamic.log 2>&1
go run ./cmd/adamic build stage1/cohere/lint/helpers/slot02/regexp-compile-blocker/testdata/constant_control.a -o /tmp/slot02-regexp-constant --sanitize > /tmp/slot02-regexp-constant-build.log 2>&1
/tmp/slot02-regexp-constant > /tmp/slot02-regexp-constant-run.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower -run '^TestRegExpNativeRefusals$' -count=1 -v -timeout=30m > /tmp/slot02-regexp-refusals.log 2>&1
```

The two fixture files initially lived directly in this owned directory; recorded compilation paths reflect that. They were moved under testdata to keep intentional refusal witnesses out of normal source copying. The final filenames were rerun before publication. An initial Node invocation used the unsupported module-types spelling, and an initial probe attempted console.log(boolean), which the embedded prelude rejects; neither is credited as evidence for the RegExp gap. The corrected probes use module-typescript and string output. Complete corrected logs are losslessly archived with checked gzip decompression.

The literal-pattern control differs only in the constructor argument and builds under ASan/UBSan; its yes output matches dynamic source Node, with no stderr. This establishes that the reproducer isolates runtime pattern lowering. It does not implement Compile or prove its acceptance/error/matching behavior, and no semantic mutant credit is claimed. The user explicitly instructed stopping at other shared blockers rather than editing shared files, so no next helper was claimed. Whole-repository checks and the other required correctness suites were not rerun for this evidence-only change.
