Built: rebased existing wave-03 onto lint area b46914832 after its registry migration; no new claims or implementation changes.
Commits: previous pushed e2dc8c6c3; rebased implementation be2a20e30e683e02a8b1a361a5c9e60882c5b4f7; this report accompanies the own-branch push.
Command/output: five owned suites ran; package FAIL, Go exit 1 after 615.691s; four suites covering twelve rules PASS.
Mutants: twelve rule byte-comparison mutants plus guard and retained-registry mutants caught; full native sanitizer and handle checks repeated for active rules.
Uncovered: blocked React oracle, whole repository gate and 17 external-input correctness checks; native dynamic RegExp remains the named blocker.

The lint area advanced to b46914832d70e00847d82d5d221ab7bb24040c53, merging lint-rules/legacy's registry migration. Fetched main remains c7991b900. Rebase completed cleanly and retained integration changes. All 287 required repository roots are present; no roots or checks were dropped. The migration changes shared lint code, while the compiler/runtime and owned source remain unchanged. No shared implementation file was edited by this unit. Only codex/typeaware-wave-03 is pushed.

After sourcing /workspace/adamic-tools/env.sh, the full command ran with output redirected to b469-oracles.log:

```sh
ADAMIC_WAVE03_FINAL_ARTIFACTS=/workspace/wave-03/b469-final ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants|ReactAgreementAndMutants|FinalAgreementAndMutants)$' -timeout 40m -count=1 -v
```

Final passes in 150.98s, more/core 147.71s, Nexus 153.28s and original 161.10s. React FAILS in 2.61s. The overall package fails with exit 1. Active rules repeat full exact findings/fixes/suggestions comparison, both retained corpus manifests, compiled mutants, sanitizer runs and released-handle checks. Final: 136 parseable controls, 100 findings and 71 suggestions. More/core: 566 controls, 318 findings. Nexus: 94 controls, 96 findings, ordered 12 and imported one. Original: 46 control findings, loose-this four, repository 91 and compiler 178. All corresponding Go/native/sanitizer streams agree.

`go test ./bridge/tsgo/checker -count=1` passes in 1.146s; `go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware` passes; `go run ./cmd/lint-registry` passes. The previous fifteen-control interpolation proof and its compiled whitespace mutation remain same-source/compiler evidence; that isolated helper was not rerun this turn. No fresh claim is made for it.

Newest three-rule whole-process observations: repository native 1.348543436s versus Go 0.226569347s; compiler native 7.711150604s versus Go 0.642671055s. These are single runs, not statistical speed claims. Native remains slower. Complete timing and phase lines for all active rules are retained.

The exact failed oracle prints `BLOCKED: required new RegExp(pattern, 'u') is not supported by native lowering`, followed by `stage 0 can't lower RegExp with a nonconstant pattern yet` at stage1/cohere/typeaware/wave_03_react/gaps/general_regex.a:3:24. It is a failure, not a skip. The standalone source and build reproducer remain in REQUIRED_FAILURE.md. The required option helper uses the constructor without a handwritten matcher fallback. Compiler/runtime support is outside this unit's rule territory; the full branch is not oracle-green and no additional claims were made.

Exact logs and final-suite streams are compressed under registry_landing_validation/, with uncompressed hashes in files.json. The full repository gate, 17 announced external-input checks, shared JavaScript external-checker profile and previously documented shared inspect-refusal mutation-anchor gap were not exercised by this filter. No skipped, relaxed or deleted check was used to obtain green. Earlier passing-package reports predate the strict constructor failure and remain historical evidence only.
