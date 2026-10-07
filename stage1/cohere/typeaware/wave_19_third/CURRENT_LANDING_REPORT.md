Built: all nine algorithms rebased onto origin/area/stage1-lint b46914832d70e00847d82d5d221ab7bb24040c53, containing main c7991b900362796aefd111474e65eb5398e91953; no new claims.
Commits: prior pushed 411639b9e04f5c5549ca14d5986a6bdfd51d4560; rebased implementation 9a72d1f4cef9fad859729e2edb1ffdb7377664cc; this report commit is pushed only to codex/typeaware-wave-19.
Commands and outputs: 50 fresh Go comparisons, 801774 canonical bytes, normal and ASAN/UBSAN/LSAN PASS; checker 0.365s, registry 0.186s, named metadata PASS. Identical compiler/bridge/parser/rule binaries reused after a source-diff assertion.
Mutants: prior six latest byte mutants and four released-handle panic checks remain applicable to identical compiler and rule sources; not rerun for this registry-only rebase. Earlier six rule receipts remain recorded.
Uncovered: shared type-aware registry integration, lint emitted-JavaScript comparison, full repository gate and its 17 external correctness checks. No selected check skipped; no shared file, main or area branch changed by this worker.

The latest 27-commit rebase completed without conflicts. Area changed the shared lint registry and syntax-rule modules, but git diff 411639b9e HEAD was empty for cmd, internal, bridge, cohere, stage1/typescript, stage1/cohere/typeaware, tsconfig.json and CohereSettings.json. The existing compiled artifacts therefore have identical inputs. All comparison cases were rerun with live independent Go and native processes, including current repository texts; no comparison was skipped. The local rebuild script was executed with only its compiler/frontend build calls replaced by artifact existence assertions after the source equality assertion. No tracked script or shared harness was changed. Logs: /tmp/wave19-registry-base-parity.log, /tmp/wave19-registry-base-packages.log and /tmp/wave19-registry-base-metadata.log.

The updated RuleContext was inspected in full; it still has no checker program/query handle, and lint_test.go native.Build still omits bridge linking. The shared integration blocker persists. All 197 ranked rules are ported or named in claims across 639 fetched origin refs; no available rule remains. Prior timings and compiler rebuild/mutation evidence below describe the preceding b84a9d931 checkpoint, not a new measurement on b46914832. Both current remote bases were checked before pushing and remained unchanged. The full gate and its 17 external correctness checks were not run.

Previous checkpoint evidence follows.

The 26-commit rebase completed without conflicts. The new base changes proven-relation lowering and introduces a record runtime, so all six owned native frontends, covering nine rules, were rebuilt with the current compiler normally and under sanitizers. Shared bridge, owned algorithms, pinned Go cohere and its rule inputs were unchanged; existing checker archives and independent Go oracle binaries were reused. Comparisons cover controls, the frozen 77-root TypeScript compiler and 287-root repository populations, isolated/index options, strict typeof suggestions, three require-await contracts and four ASI sources.

The six latest mutations are wrong Symbol callee, invalid typeof array accepted, awaits ignored, wrong union flag, strict typeof option ignored and incorrect field initializer metadata. Each compiles and exits zero with empty native stderr before Go bytes catch it. The retained-handle mutant also exits zero with empty stderr and fails the required panic checks. Four released checker questions pass normally and sanitized. The unmodified production archive still refuses all four unregistered questions with panic 70.

The shared lint RuleContext still lacks a checker program/query handle, and its native build does not link the bridge archive. The seven previously prepared registration lines remain unapplied under the shared-file restriction. The latest three owned descriptors use ast.Kind names; private handed-node compatibility listeners retain internal numeric IDs. The earlier six retain their documented legacy driver limitation. Algorithms are green; production shared integration remains pending.

Fresh single-round whole-process measurements of the original three-rule runner, with output bytes compared before accepting timing:

| Corpus | Go seconds | Native seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 0.469423 | 2.217268 | 4.72x |
| Repository | 0.199640 | 0.441989 | 2.21x |

Native remains slower; these include loading and parsing and are single observations, not medians. The prior three-round latest-rule timings remain historical. Same-unit setup previously completed in 195s, with nproc 5; setup was not rerun for this rebase.

Commands sourced /workspace/adamic-tools/env.sh and set TMPDIR=/workspace/wave19-f801-scratch. Test output was redirected to logs:

```sh
python3 stage1/cohere/typeaware/wave_19_third/rebuild_landing.py "$TMPDIR" /workspace/wave19-typescript
python3 stage1/cohere/typeaware/wave_19_third/prove_native_mutants.py "$TMPDIR"
go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1
go test ./stage1/cohere/typeaware -run '^TestWave19ThirdNumericMetadata$' -count=1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/inherited_static_field_read\.a$' -count=1 -timeout=10m
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript ADAMIC_WAVE19_THIRD_RELEASE_ARTIFACTS="$TMPDIR/third-release" go test ./stage1/cohere/typeaware -run '^TestWave19ThirdReleasedHandles$' -count=1 -v
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript ADAMIC_WAVE19_THIRD_RELEASE_ARTIFACTS="$TMPDIR/third-release" go test ./stage1/cohere/typeaware -run '^TestWave19ThirdProductionPending$' -count=1 -v
```

Logs: /tmp/wave19-current-parity.log, /tmp/wave19-current-mutants.log, /tmp/wave19-current-packages.log, /tmp/wave19-current-metadata.log, /tmp/wave19-current-vet.log, /tmp/wave19-current-node.log, /tmp/wave19-current-release.log, /tmp/wave19-current-pending.log and /tmp/wave19-current-bench.log. The rebuild writes fresh raw streams and results to the historical local directory $TMPDIR/landing-b8; its directory name does not identify the compiler base. Benchmark streams are in $TMPDIR/current-base-bench. Logs and archives stay local.

Explicit all-head fetch refreshed 631 origin refs. The 197-rule linked VOLUME_REPORT all-family ranking, Markdown claims on every origin ref and the 26 baseline ports leave no available rule. No reservation was added. The remote main and lint base were rechecked before pushing and remained c7991b900 and b84a9d931 respectively. The full gate was not run, so no claim is made about the 17 newly required external correctness checks; none was skipped, relaxed or deleted to obtain this worker's selected checks.
