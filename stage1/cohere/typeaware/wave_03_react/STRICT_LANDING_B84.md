Built: rebased existing wave-03 onto lint area b84a9d931, including main c7991b900; strict React failure retained.
Commits: previous pushed 4a2adcf2b; rebased implementation ce8adcaaf3b05d2ac77805c974cfe5dd8d3ea926; this report accompanies the own-branch push.
Command/output: all five owned suites run, package FAIL with Go exit 1 after 641.942s; four suites covering twelve rules PASS.
Mutants: twelve rule byte-comparison mutants plus guard and retained-registry mutants caught; interpolation whitespace mutant caught solely by comparison.
Uncovered: blocked full React oracle, whole repository gate and the 17 external-input correctness checks; no new claims.

The requested lint area advanced to b84a9d9314b65d3d0261ee017e233287b4f071da and includes fetched main c7991b900362796aefd111474e65eb5398e91953. Rebase completed cleanly, retaining integrated compiler, runtime, harness and developer-tool changes. No shared implementation or protected compiler file was edited by this unit. Only codex/typeaware-wave-03 is pushed.

After sourcing /workspace/adamic-tools/env.sh, the exact owned command was redirected to landing-recheck-oracles.log:

```sh
ADAMIC_WAVE03_FINAL_ARTIFACTS=/workspace/wave-03/landing-recheck-final ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants|ReactAgreementAndMutants|FinalAgreementAndMutants)$' -timeout 40m -count=1 -v
```

The package fails with exit 1. Final passes in 173.20s, more/core 152.11s, Nexus 150.89s and original 162.33s. React FAILS in 3.40s at its constructor preflight. This is not a skip or a green expected-error test. The twelve active rules repeat exact findings, fixes and suggestions comparisons, compiler/repository corpus runs, native sanitizers, compiled mutants and released-handle checks. Final controls retain 136 parseable inputs, 100 findings and 71 suggestions. More/core retains 566 controls and 318 findings. Nexus retains 94 controls and 96 findings, ordered 12 and imported one. Original retains 46 control findings, loose-this four, repository 91 and compiler 178. Existing corpus root manifests are unchanged.

The exact failure is `BLOCKED: required new RegExp(pattern, 'u') is not supported by native lowering`, followed by `stage 0 can't lower RegExp with a nonconstant pattern yet` at stage1/cohere/typeaware/wave_03_react/gaps/general_regex.a:3:24. Minimal source, standalone build command and strict-failure mutation proof are in REQUIRED_FAILURE.md. The rule option helper uses the required constructor without a handwritten fallback. Closing this requires native runtime/lowering support outside this rule unit. The full branch is not oracle-green, so no additional claims were made.

`go test ./bridge/tsgo/checker -count=1` passed in 1.096s; `go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware` passed; `go run ./cmd/lint-registry` passed. The interpolation literal was rebuilt with this base's compiler in native, sanitizer and emitted-JavaScript profiles. All fifteen controls agree with independent Go and Node, exiting zero with empty stderr. Its compiled whitespace-star-to-plus mutant remains caught solely by output comparison. These bounded helper checks do not certify the failed React suite.

Newest three-rule whole-process timing observations: repository native 1.486480523s versus Go 0.236127011s; compiler native 6.680865328s versus Go 0.523026750s. These are single runs, not a statistical performance claim. Native remains slower. All timing and phase lines are retained in the log.

No whole repository gate or pass of the 17 announced external-input checks is claimed. The shared emitted-JavaScript checker profile and the previously documented shared inspect-refusal mutation-anchor gap were not exercised by this filter. No check was skipped, relaxed or removed in this unit. Exact logs, streams, interpolation inputs and outputs are compressed in strict_landing_validation/, with uncompressed hashes in files.json. Earlier passing package reports predate the replacement of the React skip with a failure and remain historical evidence only.
