Rebased the owned branch onto current main with merge history preserved; adopted harness 41eb6eab2.
Commits: tested tip f12149b6de9ac88e27736dffb67c7d21e705b2e1; final evidence commit is its descendant on codex/typeaware-wave-04.
Checks: six completed-rule corpus oracles, sanitizers, released handles, partial JSX/React kernels and named listeners passed.
Mutants: six complete-rule byte mutants, both retained-handle mutants, all helper/reporting/refusal/listener mutants caught.
Not covered: complete JSX source ports/corpus parity, full repository gate and emitted JS linking for checker-importing listener modules.

## Landing

`git rebase --rebase-merges origin/main` succeeded on main b8fb957aa839a9e8cb0b54279dd9864fa317bd30. Preserving the imported harness's merge topology avoids the earlier flattening conflict. No manual shared-file reconciliation was necessary; the previous claim that integration had to unblock that rebase is superseded. `git merge --no-edit origin/lint-rules/harness` then adopted 41eb6eab2b6de45ede0a40250765be295ee25fbd. Both are ancestors. Owned source under stage1/cohere/typeaware and bridge/tsgo/checker is byte-identical to the previously pushed 5f25f2d49 before this evidence addition. No shared registration/harness edits and no new claims.

## Checks and timing

All commands source /workspace/adamic-tools/env.sh. A fresh compiler was built with `go build -o /workspace/typeaware-wave-04-landing-b8/adamic ./cmd/adamic`. Every test wrote directly to a log under that scratch directory; logs, exact streams and compiler/source provenance are retained in evidence/landing-b8.

`ADAMIC_WAVE04_ARTIFACTS=/workspace/typeaware-wave-04-landing-b8/original ADAMIC_WAVE04_COMPILER_MANIFEST=/workspace/typeaware-wave-04/compiler.manifest ADAMIC_WAVE04_REPOSITORY_MANIFEST=/workspace/typeaware-wave-04/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/typeaware-wave-04/typescript go test ./stage1/cohere/typeaware -run '^TestWave04AgreementAndMutants$' -count=1 -v -timeout=30m`: PASS 217.701s. Normal and sanitizer comparisons: controls 39 findings/17080 bytes, repository 14/23598, compiler 200/77043. Casing, matching-return and strict-void mutants compile/run zero with empty stderr; only Go byte comparisons catch bytes 5736, 121 and 1584. Released handle panics 70; retained registry mutant exits zero and fails the required-panic assertion.

`python3 stage1/cohere/typeaware/wave_04_next/validate.py /workspace/typeaware-wave-04-landing-b8/next --adamic /workspace/typeaware-wave-04-landing-b8/adamic --compiler-manifest /workspace/typeaware-wave-04/compiler.manifest --compiler-config /workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json --repository-manifest /workspace/typeaware-wave-04/repository.manifest --blocking-controls /workspace/typeaware-wave-04-next/blocking-controls-final --process-controls /workspace/typeaware-wave-04-next/process-controls-ts`: PASS all three continuation rules, complete findings/fixes/suggestions bytes, controls, normal/sanitized corpora and handle checks. Race-timeout, process-exit and blocking-stream span mutants compile and run cleanly; only comparisons catch them. Native versus Go: repository 0.327833s versus 0.148245s; compiler 1.949514s versus 0.376753s. Runs overlapped other checks, so these are not isolated benchmarks.

Fresh-compiler Python validators for wave_04_jsx validate_partial.py and validate_references.py: PASS 56 records/4941 bytes and 64 snapshot nodes/302 records/3801 bytes across Go/native/sanitized/source Node/emitted JS. Three reporting byte mutants, three removed-refusal mutants, three in-memory named metadata mutations and eight reference byte mutants caught.

Fresh-compiler wave_04_react validate_partial.py, validate_refs_kernel.py, validate_refs_environment.py and validate_refs_checks.py: PASS reporting and all previously parked helper comparisons across backends. Counts: 11 reporting findings/5948 bytes; joins 1297 records/55984 bytes; environment 451/9943; checks 180/3136. All six reporting/refusal, nine join, six environment and four predicate mutants caught. Source analyses remain explicitly refused.

Fresh-compiler validate_listeners.py: PASS nine named declarations/658 bytes on Go/native/sanitized/source Node, with nine clean compiling comparison-only mutants. validate_rule_json.py: PASS nine manifest/module/production comparisons and nine in-memory named metadata mutants. Emitted JS checker linking remains uncovered by that listener validator.

`go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1`: PASS 0.179s and 0.054s. Filtered Node oracle covering TestTheOracleCatchesOneByte and native closures/method_closures/generic_functions/inherited_static_field_read: PASS 1.398s. `go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware ./stage1/cohere/lint/registry`: PASS, empty log. No full repository gate.

The three JSX claims remain unfinished own source implementation work (full factories/adapters, fragment alias/options/attributes, undef bindings/declaration-file facts, context provider/component/construction/stability decisions and final ranges). Named manifest compatibility is resolved; no shared numeric-kind blocker is claimed. HIR-dependent React claims remain parked. This unit satisfies landing-first validation without adding new rule scope.
