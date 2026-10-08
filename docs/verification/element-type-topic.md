# Main-based elementType topic

Base: origin/main 6998ebc24ae353193cb1495d3d51308131a4b5c7. Branch: codex/notyet-element-type-topic.

Only this worker's five non-merge commits were carried: f544f4a2 -> 1f9a826e, 024f7856 -> 039eea81, abd8fd74 -> fb8dccf2, 17a770d9 -> bf1b9069, 8c8611b0 -> c2c98950. There are no merge commits after the main base. Census replay and checked non-null worker commits were excluded; checked views were neither carried nor merged.

The morning census is e8c283b5ed32477805357b652a170b85a04b2469, stage3/notyet-table/rerun-0730/after/roots.csv. Its elementType diagnostics have 188 roots in 16 kinds. All 81 adapted source byte counts and hashes match its source manifest. One example per kind was replayed in a detached checkout with the measurement-only census tool. No measurement-tool commit entered this branch. Compact observations are in stage3/notyet-element-type/morning-replays.json; full output is in /tmp/element-type-topic-replay-*.log.

## Ranked kinds

These are root counts from the census, not 188 individually certified compiler sites. Concrete storage fixtures prove the listed runtime rules; an earlier replay stop never proves that an original site lowered.

| Kind | Roots | Topic result and replay observation |
| --- | ---: | --- |
| never | 90 | Skipped pending concrete contextual storage or checked views for widening. builderState:567:38 encounters the earlier structural method/statics stop at 566:18. The alias counterexample is pinned to Node; .a correctly refuses its incompatible mutable aliases. No arbitrary storage was assigned to never[]. |
| T | 43 | Cancelled isolated generic census echoes, independently audited. checker:46986:35 still reproduces without a specialization. Concrete generic fixtures pass. |
| any | 21 | Refused for a ruling under 0.1's no-any doctrine. checker:11031:21 reproduces on the checker-rejected .ts census input. No unchecked universal storage was introduced. |
| boolean or undefined | 6 | Concrete array rule passes after the main port supplies packing. binder:1933:17 signature gone; named next statement stop at 1956:21, with other earlier dependency findings. Optional-boolean callback values remain a separate dependency. |
| U | 6 | Cancelled generic echoes. The first morning example, core:2524:17, now stops at NonNullExpression on this main-only base. |
| ResolvedConfigFileName | 5 | Scalar-brand element storage certified. tsbuildPublic:2226:28 encounters the earlier branded value at 2226:16; global scalar-brand representation remains owned elsewhere. |
| undefined | 3 | Concrete array rule passes. binder:1940:40 signature gone; next statement stop at 1956:21. |
| NonNullable<T> | 3 | Cancelled generic echoes. core:739:31 reproduces; concrete specializations pass. |
| Path | 2 | New exact required-brand empty-array holder and slice fixture passes in both backends. Populated scalar brands use the existing cast-free optional-never reduction. builderState:479:23 encounters earlier Path value and Set representation stops. This does not certify creation or traversal of required-brand scalar values. |
| V | 2 | Cancelled generic echoes. core:107:25 reproduces; concrete specializations pass. |
| Child | 2 | Cancelled generic echoes. emitter:4734:86 reproduces; constrained concrete specialization passes. |
| Build-info array union | 1 | Concrete compatible-reference array alternatives pass. builder:2407:5 signature gone; named stop at 2407:32 value of IncrementalMultiFileEmitBuildInfoFileInfo, plus earlier statics/Map stops. |
| object | 1 | Concrete array storage and cycle refusals pass. checker:15326:33 signature gone; next 15327:16 BinaryExpression with a value and a value. |
| CanonicalKey | 1 | Scalar-brand array storage certified. commandLineParser:4131:47 signature gone, with earlier dynamic RegExp and branded Map-key stops. |
| unknown | 1 | Refused for a ruling; core:684:12 reproduces. The unsafe acceptance mutant collapses null to undefined. Needs distinct null storage and the checked-view representation work; skipped on this topic. |
| TState | 1 | Cancelled generic echoes. factory/utilities:1396:9 reproduces; concrete number/boolean state specializations pass. |

Accounting: 19 roots belong to concrete storage kinds covered by fixtures, 57 generic roots are cancelled as isolated census echoes, 90 never roots are skipped, and 22 any/unknown roots retain their refusal/ruling requirements. The former IncrementalBuildInfoRoot and Extension/string array-alternative signatures are absent from this morning root list; their carried fixtures remain tested.

## Main port and runtime-owner review

Main lacked the optional-boolean slot packing supplied by an area dependency. This topic supplies only the array storage prerequisites: a uint8_t member and pack/unpack declarations in runtime/adamic.h, member selection in emit_values.go, packing in slots.go, and new separate runtime/element_type_boolean_slots.c. Tags are undefined 0, false 1, true 2. Review this new helper with the runtime owner when combining topic branches. Existing sort variants remain together in sort_undefined.c from the carried, reviewed consolidation.

The old optional-boolean find callbacks depended on another worker's function-value lowering. Those three expressions were removed from this worker's fixture and are explicitly outside this main-based certification. No worker-owned lowering function was edited to recreate that dependency. Named comparator sorting remains covered, including push/shorten ownership behavior in the carried fixtures. Exact Path holders add no production lowering rule.

## Checks and mutants

All test output was redirected to /tmp/element-type-topic-*.log. No package-wide tests or full gate were run.

- Setup: GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Node .049s, Go .050s, markdown .138s, submodules .143s, clang .314s, Go build 83.319s, deferred tests 83.517s, cache 83.519s, total 83.570s. nproc 5, cgroup four cores.
- go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/element_type_|TestElementType' -count=1 -v: passed, 6.611s. Six positive fixtures agree with Node in JavaScript and native under ASan/UBSan; unknown keeps its NotYet and the never alias keeps its existing invariant-mutable refusal, with separate Node observations pinned.
- go test ./internal/lower -run 'TestElementTypeObjectArraysCannotCloseCycles|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat' -count=1: passed, 1.819s.
- go test ./internal/native -run 'TestCEndsInNewline|TestMaybeNumbersPackIntoOneDouble|TestLoopArrayHoldC' -count=1: passed, .452s.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts: passed, 53.224s. Only this unit's changed boolean and brand rows changed; all positive fixture counts end with zero live allocations.
- NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/notyet-element-type/generic_echoes.cjs /tmp/element-type-morning-roots.json /tmp/element-type-census-input stage3/notyet-element-type/generic-echoes.json: passed, 57 sites. GENERIC_ECHO_MUTANT=drop-parameters fails with missing generic owner.

Nine registered native C mutants build and exit normally with empty sanitizer stderr, and are caught only by Node stdout: wrong generic index; false unpacked as true; optional-boolean sort calling undefined; boxed union identity instead of numeric equality; union sort calling undefined; lost appended sort tail; wrong shortened sort length; discarded flattened contents; branded array length incremented. The new Path length mutant passed its check in 1.788s together with the corresponding oracle fixture.

Five Go overlays remove scalar-brand, undefined, object, array-union or object-intersection support; each positive fixture fails lowering. The sixth accepts unknown as Union; sanitized native exits cleanly but stdout differs from Node. Removing the minimal NonPrimitive cycle hook makes the two existing object-array cycle refusal tests fail. No claim is made that an earlier stop, C compilation error or sanitizer error killed these stdout mutants.

The cycle-hook mutant exited 1 in .128s: both object-array cycle tests reported wrongly accepted nil errors. A second never example, checker:11003:40, was replayed as well; its observation is recorded beside the ranked kind evidence.
