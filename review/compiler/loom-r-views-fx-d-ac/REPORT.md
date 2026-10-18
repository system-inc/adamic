Audited all 40 changed Adamic programs with the branch compiler and the exact pinned a-check predicate; no headers are needed.
Base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; combined source 45d2ef9ac165fd78b791df03d61d9c058de0b16b; delivery is one commit on compiler/loom-r-views-fx-d-ac.
All 40 checked, 0 NotYet, 0 refused, 0 type errors; Gate.aCheck exits 0. Reader guard, vet, 52 focused reader-package selections and flow checks pass.
Removed-header and wrong-reason mutants on a temporary refused control both fail the exact predicate; the correctly headed control passes.
Not covered: full-tree a-check, whole package suites, full gate, ahra or fresh reruns of the prior compiler mutants; original evidence is carried unchanged.

Read CLAUDE.md, docs/front-door.md and stage3/a-check-headers/README.md whole. The managed environment is ready after the previous attempt's container_creation transport timeout. No changes from that failed attempt existed. Fetch moved origin/main to 99c369e0, but the user's explicit dbdf977f5b base takes precedence. Created the requested branch there, then applied only 45d2ef9a with git cherry-pick --no-commit. Application was clean, with no conflicts.

The ordinary audit imports cloud/fast-gate/run.py from the exact README pin fbac28c62493f27a788edc02a18bc8edb68de5da, saved as pinned-gate.py without alteration. Gate.aCheck is invoked unchanged. Its build command builds this branch's ./cmd/adamic and each input is compiled in place with adamic c. The subprocess wrappers add hard 180-second build and 55-second per-file limits; compiler output, exit, time and source hashes are retained in initial-statuses.json. Generated C is measured and hashed, not committed. The predicate itself groups NotYet with checked; this audit separately records NotYet. All 40 returned exit 0 with empty diagnostics, so none was a NotYet stop. Total audit, including compiler build: 7.2 seconds. No full-tree scan was run, since this unit is the changed-file audit.

Headers added: zero. There were no refused or checker-error programs to annotate, and no header was invented for a clean program. All 40 complete program files, not just their bodies, are byte-identical to 45d2ef9a. unchanged-programs.json records SHA-256 and equality checks. Their original first lines are in before-hashes.json. No runtime C, compiler code, fixtures, Node records, hashes or position expectations changed. path-references.log records literal path searches in *_test.go and testdata. With no header insertion, no source line moved and no expectation needed updating. internal/oracle/counts.md is byte-identical to 45d2ef9a; no new fixture or counted operation was added, so counts were not regenerated again.

Because this change needs no headers, the negative proof uses a temporary copy of the existing stage3/census/repro/r37/main.a rather than changing a repository fixture or introducing a new .a program. It is refused for delete. The valid existing header passes Gate.aCheck with exit 0; removing the first line fails with outcome refused, expected checked; replacing delete in the header with deliberately wrong refusal reason fails with outcome refused and that incorrect expectation. Each proof invokes the actual branch compiler again, not a diagnostic stub. The source body is preserved in all variants, the original repository source never changes, and the temporary directory is removed automatically. The predicate outcomes, exact diagnostics and seconds are in header-mutants.json and .log; mutants are stored as .patch. The initial attempted historical control, r33's definite-assignment example, now compiles cleanly on this base and was rejected as a positive control; that observation is retained in initial-control-selection.log and is not credited as a mutant kill. No header on r33 was edited by this unit.

Setup command: export GOPROXY='https://proxy.golang.org|direct'; timeout 300 bash cloud/setup.sh > setup.log 2>&1; source /workspace/adamic-tools/env.sh. Exit 0. nproc=5, cpu.max=400000 100000 (four CPU share), Linux amd64, Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup timing lines: Go 0.067 s, Node 0.069 s, submodules 0.154 s, markdown ready 0.189 s (step duration 0.022 s), clang 0.467 s, shared cache 2.403 s, go build 38.798 s, tests deferred 39.012 s, cache warm 39.015 s, done 39.069 s. The failed read-only redirect while selecting a control did not execute that compiler command; the later permitted proof supplies all actual observations.

Exact verification commands (all test output went directly to files):

* timeout 360 python3 review/compiler/loom-r-views-fx-d-ac/audit.py initial > initial-audit.log 2>&1: exit 0, 40 checked, predicate PASS.
* timeout 300 python3 review/compiler/loom-r-views-fx-d-ac/header-mutants.py > header-mutants.log 2>&1: exit 0 from the proof harness, valid control predicate exit 0 and both mutant predicate exits 1.
* timeout 480 go test $(cat review/compiler/loom-r-views-fx-d/reader-packages-test-imports.txt) -run '^(TestFX7|TestLoomViewsD|TestCallableContract|TestViewCheckedCastMetadata$|TestRuntimeSetsAreEachUsed$)' -count=1 -v -timeout 55s > readers-focused.log 2>&1: exit 0, all 52 packages passed or compiled with no matching tests. The selection exercises all 40 carried compiler-focused leaves, including P09/P67/P72 across native release, sanitized native and JavaScript. It does not select the prior over-budget unchanged TestViewMixedUnionUnknownAndUnavailable.
* timeout 150 go test ./internal/flow -run '^TestFlowCorpus(UnitsCoverEveryProgram|SetupIsShared|Remainder)$' -count=1 -v -timeout 55s > flow-corpus.log 2>&1: exit 0, 0.02 / 0.20 / 13.80 s. No fixture exclusion or package test was edited.
* timeout 150 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s > reader-guard.log 2>&1: exit 0, 3.43 s.
* timeout 150 go vet ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle > vet.log 2>&1: exit 0, no diagnostics.

No new test leaf was added or edited by this pass. Every observed top-level leaf is below 60 s; leaf-seconds.json records each one. The production sources, runtime fix, earlier counts regeneration and original 18 mutant catches remain exactly those of 45d2ef9a, whose complete evidence is included in the combined change. They were not claimed as freshly rerun header mutants. No full gate, ahra command, PR or Loom submission was run.

Per-file status table:

| File | Status | Seconds |
|---|---|---:|
| internal/lower/testdata/fx7_tuple_recognition/p68.a | checked | 0.0640 |
| internal/lower/testdata/fx7_tuple_recognition/p69.a | checked | 0.0616 |
| internal/lower/testdata/fx7_wrong_aborts/bad_callable.a | checked | 0.0543 |
| internal/lower/testdata/fx7_wrong_aborts/p18.a | checked | 0.0696 |
| internal/lower/testdata/fx7_wrong_aborts/p20.a | checked | 0.0666 |
| internal/lower/testdata/fx7_wrong_aborts/wide.a | checked | 0.0533 |
| internal/oracle/testdata/loom_views_d/checked-cast.a | checked | 0.0592 |
| review/compiler/loom-views-fx-d/baseline-sources/158a78dd8b4f/fxspptb_oct9_native_p72_tuple_count_mismatch.a | checked | 0.0665 |
| review/compiler/loom-views-fx-d/baseline-sources/166413c08ae4/p69.a | checked | 0.0592 |
| review/compiler/loom-views-fx-d/baseline-sources/19560c042dd4/depth-2000.a | checked | 0.0660 |
| review/compiler/loom-views-fx-d/baseline-sources/1b417795343f/fxspptb_oct9_native_p35_class_in_union.a | checked | 0.0548 |
| review/compiler/loom-views-fx-d/baseline-sources/1e717dc21ea2/fxspptb_oct9_native_p67_static_class_union_member.a | checked | 0.0642 |
| review/compiler/loom-views-fx-d/baseline-sources/2212de141cf1/fxspptb_oct9_views_p68_object_view_over_tuple.a | checked | 0.0767 |
| review/compiler/loom-views-fx-d/baseline-sources/27fe1106b197/wide.a | checked | 0.0639 |
| review/compiler/loom-views-fx-d/baseline-sources/2db1d7d06600/depth-200.a | checked | 0.0663 |
| review/compiler/loom-views-fx-d/baseline-sources/6a43453375e8/fxspptb_oct9_native_p69_deep_recursive_union.a | checked | 0.0650 |
| review/compiler/loom-views-fx-d/baseline-sources/702776fe1a62/scalar-tuple.a | checked | 0.0577 |
| review/compiler/loom-views-fx-d/baseline-sources/78a920f10d1e/depth-65.a | checked | 0.0617 |
| review/compiler/loom-views-fx-d/baseline-sources/8d3e50ec301c/fxspptb_oct9_views_p32_boxed_slot_string_good.a | checked | 0.0629 |
| review/compiler/loom-views-fx-d/baseline-sources/939d79256bbb/fxspptb_oct9_views_p33_boxed_slot_union_member.a | checked | 0.0691 |
| review/compiler/loom-views-fx-d/baseline-sources/948f2e39c0b4/fxspptb_oct9_native_p20_callable_member_call_good.a | checked | 0.0564 |
| review/compiler/loom-views-fx-d/baseline-sources/b1a2a62ba1f4/bad_callable.a | checked | 0.0549 |
| review/compiler/loom-views-fx-d/baseline-sources/ce11e85355d8/fxspptb_oct9_native_p33_map_in_union.a | checked | 0.0584 |
| review/compiler/loom-views-fx-d/baseline-sources/d7286c5c4c87/fxspptb_oct9_views_p69_tuple_view_over_tuple_union.a | checked | 0.0592 |
| review/compiler/loom-views-fx-d/baseline-sources/d9df1651d811/fxspptb_oct9_native_p17_callable_call_first_half.a | checked | 0.0644 |
| review/compiler/loom-views-fx-d/baseline-sources/dc4b7e9bcfea/fxspptb_oct9_native_p70_recursive_depth_70.a | checked | 0.0683 |
| review/compiler/loom-views-fx-d/baseline-sources/dfe95dc594a1/fxspptb_oct9_native_p18_cast_kind_no_field_types.a | checked | 0.0652 |
| review/compiler/loom-views-fx-d/baseline-sources/e3f464103722/fxspptb_oct9_native_p34_typed_array_in_union.a | checked | 0.0855 |
| review/compiler/loom-views-fx-d/baseline-sources/f8c1252799ca/fxspptb_oct9_native_p26_callable_or_undefined.a | checked | 0.0631 |
| review/compiler/loom-views-fx-d/baseline-sources/f93df345ed22/fxspptb_oct9_views_p09_tuple_destructure_wrong.a | checked | 0.0674 |
| stage3/interface-downcasts/fx7/depth-200.a | checked | 0.0642 |
| stage3/interface-downcasts/fx7/depth-2000.a | checked | 0.0672 |
| stage3/interface-downcasts/fx7/depth-65.a | checked | 0.0691 |
| stage3/interface-downcasts/fx7/p09.a | checked | 0.0906 |
| stage3/interface-downcasts/fx7/p32.a | checked | 0.0850 |
| stage3/interface-downcasts/fx7/p33.a | checked | 0.0754 |
| stage3/interface-downcasts/fx7/p67.a | checked | 0.0525 |
| stage3/interface-downcasts/fx7/p72.a | checked | 0.0672 |
| stage3/interface-downcasts/fx7/scalar-static.a | checked | 0.0569 |
| stage3/interface-downcasts/fx7/scalar-tuple.a | checked | 0.0750 |

Header predicate proof:

| Variant | Predicate exit | Seconds |
|---|---:|---:|
| control | 0 | 2.2468 |
| removed-header | 1 | 1.7412 |
| wrong-reason | 1 | 1.6594 |

Measured passing test leaves:

| Test | Seconds |
|---|---:|
| TestFX7TupleMarkerRejectsOrdinaryArray | 0.17 |
| TestFX7ViewCycle | 0.19 |
| TestCallableContractPlainSpread | 1.87 |
| TestCallableContractReceiverEvaluatedOnce | 1.91 |
| TestCallableContractP20 | 2.00 |
| TestCallableContractP18 | 2.07 |
| TestCallableContractSpreadRefused | 0.42 |
| TestCallableContractWideProducer | 2.41 |
| TestCallableContractReceiverBinding | 0.71 |
| TestViewCheckedCastMetadata | 0.00 |
| TestFX7TupleSelectionBeforeDiagnosticKindRelease | 4.19 |
| TestFX7TupleSelectionBeforeDiagnosticKindSanitized | 4.95 |
| TestFX7ViewCycleAndStackGuard | 5.03 |
| TestFX7TextScalarStatic | 10.29 |
| TestFX7Depth65 | 10.59 |
| TestFX7TextP33 | 11.23 |
| TestFX7TextP32 | 11.44 |
| TestFX7TextScalarStaticJavaScript | 0.15 |
| TestFX7TextScalarTuple | 1.79 |
| TestFX7TextP72 | 2.09 |
| TestLoomViewsDOriginal | 12.64 |
| TestFX7TextScalarStaticSanitized | 1.15 |
| TestFX7TextP67 | 1.96 |
| TestLoomViewsDTupleUnionJavaScript | 1.15 |
| TestLoomViewsDWideCallableJavaScript | 1.35 |
| TestLoomViewsDCallableUnion | 2.79 |
| TestLoomViewsDBadCallable | 3.03 |
| TestLoomViewsDWideCallableSanitized | 2.19 |
| TestLoomViewsDTupleObjectJavaScript | 0.34 |
| TestLoomViewsDTupleObjectSanitized | 1.81 |
| TestLoomViewsDTupleUnionSanitized | 1.08 |
| TestFX7Depth200 | 1.66 |
| TestLoomViewsDTupleUnion | 0.88 |
| TestFX7Depth65JavaScript | 0.43 |
| TestFX7TextP09 | 1.97 |
| TestLoomViewsDWideCallable | 0.80 |
| TestRuntimeSetsAreEachUsed | 5.99 |
| TestLoomViewsDTupleObject | 0.65 |
| TestFX7Depth2000 | 2.38 |
| TestFX7Depth65Sanitized | 1.00 |
| TestCallTargetReaders | 3.43 |
| TestFlowCorpusUnitsCoverEveryProgram | 0.02 |
| TestFlowCorpusSetupIsShared | 0.20 |
| TestFlowCorpusRemainder | 13.80 |
