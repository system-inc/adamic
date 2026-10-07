Built: rebased existing wave-03 work onto origin/area/stage1-lint 7481e0324; no new rules or claims.
Commits: previous own tip 375301531; rebased implementation 31dcc090020a4c4c1eb40d178a15bdbb2a827285; this report accompanies the own-branch push.
Checks: twelve rules pass four full suites in a 570.413s run; three-rule React suite explicitly skipped for native runtime RegExp refusal.
Mutants: all twelve rule comparison mutants plus guard and retained-handle registry mutants caught; literal interpolation mutant caught solely by comparison.
Uncovered: full React oracle under dynamic option patterns, whole repository gate, shared JavaScript external-checker execution.

The requested integration base now points to 7481e0324e34a2537aafa9db7eeacda50405611b. It includes fetched origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad and the named harness 41eb6eab2. Rebase completed cleanly; integrated harness and developer-tool changes were retained. No shared implementation file was edited. Only codex/typeaware-wave-03 is pushed, never main or area.

Commands source /workspace/adamic-tools/env.sh and redirect outputs to area_validation logs. Registry validation `go run ./cmd/lint-registry` passed (15 generated descriptors; generated files remain ignored). `go test ./bridge/tsgo/checker -count=1` passed in 1.038s. `go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware` passed.

The full owned suite command was:

```sh
ADAMIC_WAVE03_FINAL_ARTIFACTS=/workspace/wave-03/area-final ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants|ReactAgreementAndMutants|FinalAgreementAndMutants)$' -timeout 40m -count=1 -v
```

Full suites passed: final 132.76s, more/core 143.05s, Nexus 149.79s, original 142.40s. React explicitly prints BLOCKED and SKIP in 2.40s; the package PASS does not certify those three skipped rules. All active suites repeat exact findings, fixes and suggestions comparisons on controls and both existing corpus manifests, native sanitizers, compiled comparison mutants and released-handle checks. The final suite has 136 parseable controls, 100 findings and 71 suggestions; its stream is 64608 identical bytes. More/core has 566 controls and 318 findings; Nexus has 94 controls and 96 findings plus ordered 12 and imported one; original has 46 controls findings plus loose-this four. Original repository findings on this integrated baseline are 91, with Go/native/sanitizer agreement, versus 86 on the earlier main baseline. Compiler findings remain 178. These observations use the retained root manifests against current checked-out repository contents; the compiler source pin remains frozen.

Newest three-rule whole-process timings observed in this run: repository native 0.924206127s versus Go 0.206603541s; compiler native 6.011127633s versus Go 0.427304383s. These are single observations, not a statistical speed claim; native remains slower. All phase and timing lines are preserved in area-oracles.log.gz.

The message interpolation helper was rebuilt with the area compiler in native, emitted-JavaScript and sanitizer profiles. Fifteen independent Go/native/Node/sanitizer controls agree, with zero exits and empty stderr. Its whitespace-star-to-plus mutant compiles, runs normally, and is killed only by comparison. This helper does not substitute for a full React rule oracle.

Current area retains the exact blocker: `stage 0 can't lower RegExp with a nonconstant pattern yet` at wave_03_react/gaps/general_regex.a:3:24. The required option implementation in boolean_pattern.a uses `new RegExp(pattern, 'u')`; there is no handwritten fallback. The older full-driver probe and migration details remain in REGEX_MIGRATION.md. This is a native regex lowering blocker, not the IR/SSA/capture-analysis parking category. The branch is not described as fully oracle-green; no new claims are taken until this existing blocker is resolved or integrated. No compiler workaround was made.

Exact logs, streams, interpolation inputs and outputs are compressed under area_validation/ with uncompressed SHA-256 hashes in files.json. Previous main-baseline reports are historical evidence. No whole-repository gate or generic registry integration of the private type-aware drivers is claimed.
