Built: all nine wave-19 algorithms rebased onto origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da, containing current origin/main c7991b900362796aefd111474e65eb5398e91953; no new rule claims.
Commits: prior pushed ef1e00ee613387c3e1acc6b55aefae7bf550b5d7; rebased implementation 6b40c29dbdebfe49d50c0c60e86262f4fe9619bf; the commit containing this report is pushed only to codex/typeaware-wave-19.
Commands and outputs: fresh compiler rebuild passed 50 comparisons and 801774 canonical Go bytes normally and with ASAN/UBSAN/LSAN; checker/registry/metadata/vet and filtered Node passed. Released-handle gate PASS 54.744s; production refusal PASS 14.033s.
Mutants: six freshly compiled latest-rule mutations caught only by independent Go bytes at 56, 2365, 8388, 66, 3334 and 628; retained-registry mutant caught by all four required released-handle panics. Earlier six rule mutation receipts remain in prior reports.
Uncovered: shared type-aware registry integration, lint emitted-JavaScript comparison, full repository gate and its 17 external correctness checks. No selected check skipped; no shared file, main or area branch changed by this worker.

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
