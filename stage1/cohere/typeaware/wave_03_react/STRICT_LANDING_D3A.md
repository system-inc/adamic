Built: rebased existing wave-03 onto lint area d3a37422c; no new claims or implementation changes.
Commits: previous pushed 4c2b3629a; rebased implementation ea3d30d8f76dee76e481aeca7ace5d7c0ef538c8; this report accompanies the own-branch push.
Command/output: initial five-suite run FAIL in 628.038s; final-suite disk recovery retry PASS in 113.578s; twelve active rules pass, React fails.
Mutants: active rule comparison mutants, guard and retained-registry mutants caught; native sanitizer and released-handle checks pass; interpolation mutant caught.
Uncovered: React runtime RegExp construction, whole repository gate and 17 external-input checks; branch is not oracle-green.

Rebased cleanly onto origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, including fetched main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and its typeof/runtime lookup fixes. Integration changes were retained. All 77 compiler and 287 repository roots remain present. No compiler, shared harness or generator files were edited. Only codex/typeaware-wave-03 is pushed.

After sourcing /workspace/adamic-tools/env.sh, commands ran with output redirected to the named logs:

```sh
ADAMIC_WAVE03_FINAL_ARTIFACTS=/workspace/wave-03/d3a-final ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants|ReactAgreementAndMutants|FinalAgreementAndMutants)$' -timeout 40m -count=1 -v
ADAMIC_WAVE03_FINAL_ARTIFACTS=/workspace/wave-03/d3a-final-retry ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03FinalAgreementAndMutants$' -timeout 30m -count=1 -v
go test ./bridge/tsgo/checker -count=1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
go run ./cmd/lint-registry
```

The initial package exits 1 after 628.038s. More/core PASS in 161.07s, Nexus PASS in 157.42s, original PASS in 168.06s. React FAILS in 4.64s on the known constructor refusal. The initial final suite fails from ENOSPC while writing 032-final-released-run.stdout and is not counted as green. After cleanup the entire final suite repeats and PASSes in 113.578s with exit 0. Checker PASSes in 1.855s; vet and registry pass.

The 32GB root overlay filled with historical scratch executables and archives, not source or corpus growth. Cleanup removed only 353 obsolete ELF executables/archives from /workspace/wave-03, totaling 13,698,032,415 bytes. Source, corpus, logs, evidence and current d3a artifacts were preserved. The cleanup manifest is included. Historical binary paths may require rebuilding; their logs and checked-in evidence remain. No repository or toolchain files were deleted. The initial interpolation build also failed from ENOSPC; its subsequent native, sanitizer, emitted-JavaScript and mutant builds passed.

Twelve active rules repeat exact findings/fixes/suggestions comparisons, both corpora, per-rule compiled mutants, sanitizers and released-handle checks. Final retry: 136 parseable controls, 100 findings, 71 suggestions. More/core: 566 controls, 318 findings. Nexus: 94 controls, 96 findings, ordered 12 and imported one. Original: 46 control findings, loose-this four, repository 91 and compiler 178. Each comparison-only mutant is rejected; final retained-registry mutant exits 0 and is rejected against expected panic 70. Fresh interpolation repeats 15 Go/native/emitted-JavaScript/sanitizer controls; the whitespace-star-to-plus mutant exits normally and is caught solely by byte comparison.

Successful final-retry whole-process observations: repository native 1.102920796s versus Go 0.209327953s; compiler native 5.733539614s versus Go 0.442182258s. These are single-run observations, not statistical benchmarks; native remains slower. Raw logs retain phase timings.

The required React oracle fails with `stage 0 can't lower RegExp with a nonconstant pattern yet` at gaps/general_regex.a:3:24. Full React compilation reaches the same refusal at boolean_pattern.a:4:23. The option helper uses new RegExp(pattern, 'u'); no handwritten fallback or skip is present. See REQUIRED_FAILURE.md for the minimal reproducer. Runtime regex lowering is outside this unit's rule territory and is not the allowed React IR/SSA/capture parking category. No additional rules were claimed.

Complete fresh logs, both initial and retry final-suite streams, interpolation results and cleanup manifest are compressed in typeof_landing_validation/; files.json hashes the uncompressed contents. Failed initial streams are historical failures, not successful evidence. The full repository gate, 17 external-input checks, shared emitted-JavaScript checker profile and previously documented inspect-refusal mutation-anchor gap were not exercised. No check was skipped, relaxed or deleted.
