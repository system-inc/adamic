Built: rebased all 36 wave-15 commits onto requested origin/area/stage1-lint 7481e0324, which includes current main 39638d9e2; no new claims.
SHAs: prior remote 4cba317fb; tested rebased code a1b6a5f63; evidence commit follows in git log.
Checks: nine suites PASS 972.944s; bridge 64.350s/checker 0.414s; Node 0.915s; registry 0.036s; JSX parser 110.163s; adapter and Unicode scripts PASS; vet clean.
Mutants: 22 legacy byte mutants, eight JSX/Unicode byte mutants, three JSX refusal mutants, released-handle/registry guards and landed parser/registry mutants caught.
Uncovered: full repository gate and registry certification; live checker/factory wiring, memo/capture analysis and Unicode quoted names remain incomplete.

Explicitly requested area rebase completed without conflicts. The fetched tip had advanced from the user's 50a5f105 merge to 7481e0324, which includes harness 41eb6eab2, shared helpers, the pre-push checker and current main. Both origin/main and origin/area/stage1-lint remain ancestors after the final fetch. No protected compiler, shared parser, harness, generator or other worker's rule was edited. All their changes were inherited. Updated CLAUDE.md and docs/lint-registration.md were read. The latter requires future integrated rules under stage1/cohere/lint/rules/<slug>; legacy typeaware snapshots were not registered or presented as complete implementations.

The new pre-push checker was tried on the committed legacy claim:

    ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript python3 cloud/lint-wave-check.py --claim stage1/cohere/typeaware/claims/wave-15.md

It printed FAIL claim: --claim must name a reservation under stage1/cohere/lint/claims/. Exact output is retained in validation/area-rebase/prepush.log. No successful receipt or registry certification is claimed. This push publishes the user-requested rebase of existing partial work to its own branch, not a completed new-rule reservation under that checker. The claim was not moved, reclassified or hidden to make the check pass. No shared validator was edited. Integrated factory/visitor, oracle/provenance/witness descriptors and live checker-fact wiring remain actual work, not a claim-format-only fix.

Verification, with /workspace/adamic-tools/env.sh sourced and TMPDIR=/tmp/wave15-area-tmp:

    go run ./cmd/lint-registry
    go test -v -count=1 -timeout=10m ./stage1/cohere/lint/registry ./stage1/typescript/parser -run 'TestDeterministicRegeneration|TestDescriptorRejections|TestAdamicRuleModule|TestJsx'
    go test -v -count=1 -timeout=30m ./stage1/cohere/typeaware -run '^TestWave15|^TestCoverageAgreementAndMutants$'
    ADAMIC_TSGO_CORPUS=/workspace/wave-15-typescript go test -v -count=1 -timeout=15m ./bridge/tsgo/...
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(class_layouts|inherited_static_field_read|override_same_representation)\.a$'
    go vet ./bridge/tsgo/... ./stage1/cohere/typeaware ./stage1/cohere/lint/registry
    go build -o /tmp/wave15-area-adamic ./cmd/adamic
    python3 stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/check.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_indirect.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_unicode.py

All verification commands above passed. The pre-push claim-layout check is separate and did not pass. Every test wrote output to a log. Both wave and inherited coverage corpus settings explicitly used /workspace/wave-15-compiler.manifest (77 files) and /workspace/wave-15-repository.manifest (287 files), with ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript. Artifact variables point to /tmp/wave15-area/{original,continuation,leaked,core,regex,throw,arrow,backreference,coverage}. Adapter scripts used the newly rebuilt compiler through ADAMIC_WAVE15_SIXTH_COMPILER and separate area artifact directories. Logs, deterministic canonical stream archives, stdout hashes and the 36-commit provenance map are retained in validation/area-rebase.

Inherited corpus bytes remain identical: compiler 16,589 findings/7,120,921 bytes and repository 180 findings/85,151 bytes, ordinary and sanitized. Original three compiler two findings/5,629 bytes and repository one finding/18,704 bytes. All 22 successfully executing suite mutants are caught only by independent Go bytes: before, cast, methods, coercion, caller, parameter, invariant, optional, alias, unused, arrow, backreference, listener, mock, label, throw, leaked, flags, suggestion, default, find, sort. Released checker handles panic 70; the compiling released-registry mutant exits zero and is caught by the required panic.

Original adapter controls still match 17 findings and two option variants under sanitizers, with three byte mutants and three refusal mutants. Expanded context controls match 22 findings/7,770 bytes with branch-order, provider-factory, class-owner and function-kind byte mutants. Unicode IsUpper and RegExp Lu agree across all 1,114,112 code points on Go, Node, native and sanitized native, 1,886 positive points/10,634 bytes. The compiling Lu-to-Ll mutant is caught by independent Go/Node bytes. These controls do not own live checker handles. Production satisfies panic exit 2 and native explicit refusal exit 70 are retained outside successful byte agreement.

Landed parser checks cover 348 JSX inputs/277,842 whole-tree bytes under ordinary and sanitized native, plus their Go/Node comparisons, JSX/scanner/name-boundary mutants. Registry rejection/module/regeneration tests pass. Focused Node oracle has 10 native and seven Node misses, zero hits.

Native versus Go, original three: compiler 2.061862s versus 0.333565s. Arrow compiler 2.143194s versus 0.339110s. No speedup is claimed. Unicode sweep native 0.250s/sanitized 2.929s; its Go go-run timing includes compilation, so no end-to-end speed ratio is inferred.

Setup passed in 48s: Go/clang/Node/submodules readiness 0s, build cache warm 48s; nproc 5, CPU quota 4, memory 17.6GB. The full go test ./... gate was not run. Known JSX snapshot wiring, memo/capture and quoted-name gaps, parked HIR claims and partial regex helper/matcher limits remain. The prior hand-written regex scanners are not promoted to complete ports under the new RegExp requirement. Nothing was pushed to main or any area branch. The user-authorized own-branch rebase is published using an exact lease against 4cba317fb.
