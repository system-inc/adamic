Built: rebased existing wave-03 work onto lint area d65a8f931, preserving the integrated runtime optimizations; no new claims.
Commits: previous pushed tip fbe1c037f; rebased implementation 266200a651b9333e480949216feae98e6f1d8cad; this evidence accompanies the own-branch push.
Checks: twelve rules pass four full owned suites in a 631.669s run; the three-rule React suite explicitly skipped for its native regex blocker.
Mutants: twelve rule byte-comparison mutants plus guards and retained-registry mutants caught; compiled interpolation whitespace mutant caught solely by comparison.
Uncovered: full React oracle with dynamic option patterns, whole repository gate and shared emitted-JavaScript external-checker execution.

Origin/area/stage1-lint advanced from 7481e0324 to d65a8f931c98655936ae04c6899f38f14862b73e, merging codex/lint-runtime-profile's release-drain, destruction-path, string-header comparison and reverse byte-search improvements. Rebase completed cleanly and preserved those changes. Fetched main remains 39638d9e2 and is an ancestor. No compiler, shared harness, generator or runtime file was edited by this unit. Only codex/typeaware-wave-03 is pushed.

After sourcing /workspace/adamic-tools/env.sh, outputs were redirected to the retained logs. `go test ./bridge/tsgo/checker -count=1` passed in 1.207s; `go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware` passed; `go run ./cmd/lint-registry` passed. The owned suite command was:

```sh
ADAMIC_WAVE03_FINAL_ARTIFACTS=/workspace/wave-03/recheck-final ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants|ReactAgreementAndMutants|FinalAgreementAndMutants)$' -timeout 40m -count=1 -v
```

Final passed in 138.65s; more/core 154.15s; Nexus 155.65s; original 179.06s. React printed the exact known blocker and SKIP in 4.15s. Package PASS does not certify the skipped suite. The twelve active rules repeat full findings, fixes and suggestions comparison, sanitizer runs, compiled mutants and released-handle probes under the newly integrated runtime. Final controls retain 136 parseable inputs, 100 findings and 71 suggestions; both final-rule corpus streams remain zero findings and byte-identical. More/core retains 566 controls and 318 findings. Nexus retains 94 controls and 96 findings plus ordered 12 and imported one. Original controls retain 46 findings plus loose-this four; repository has 91 findings and compiler 178, matching Go and native sanitizers. Both existing corpus root manifests were retained.

Newest three-rule whole-process observations: repository native 1.055499561s versus Go 0.281304608s; compiler native 6.883128772s versus Go 0.446837477s. These are individual runs, not a statistical comparison of the runtime optimization. Native remains slower. Full timing lines for all owned suites are retained in recheck-oracles.log.gz.

The migrated literal interpolation helper was rebuilt with the refreshed area's compiler under native, emitted-JavaScript and sanitizer profiles. Fifteen controls match independent Go and Node with zero exits and empty stderr. Its compiled whitespace-star-to-plus mutant is still caught solely by output comparison. Exact inputs and outputs are preserved. This bounded helper check does not certify the complete blocked React driver.

The exact constructor refusal remains `stage 0 can't lower RegExp with a nonconstant pattern yet` at wave_03_react/gaps/general_regex.a:3:24. The owned option helper implements the mandatory `new RegExp(pattern, 'u')` without a handwritten matcher fallback. Current area still has the explicit refusal in internal/lower/regexp.go. This is native regex lowering, not the parked IR/SSA/capture-analysis category. No new claims were taken; the branch is not called fully oracle-green while this existing blocker persists.

Logs and streams are compressed in runtime_area_validation/; files.json contains uncompressed SHA-256 hashes. Previous reports remain historical baseline evidence. The whole repository gate and full blocked React timings were not run. No source implementation changed during this landing unit.
