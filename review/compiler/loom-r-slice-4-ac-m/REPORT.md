Re-cut slice 4 onto current main, retaining main infrastructure and all branch lowering.
Base 5a94465370ea0172e25aae8c13478e59e40c2bb4; source f353ac21ba3bba9c45f6f85b5bedb283a98d43b3; delivery SHA is in the handoff.
All current touched-package top-level tests are green or built-in skips; 84/84 a-check, guards, vet and gofmt pass.
Union-layout and generator-owner overlay mutants both fail their exact refusal tests with got <nil>.
Not covered: full repository/Loom gate, optional WASI and opt-in benchmark probes; no ahra command or submission.

Assumptions: the latest explicit full touched-package instruction defines this re-cut validation scope. The two conflict intents are compatible: normalize readonly views before main layout checks and retain typed generator parameters. The supplied main had advanced; fetched newest origin/main 5a94465370ea0172e25aae8c13478e59e40c2bb4.

Conflicts and resolutions:

- internal/lower/functions.go, signature incoming destructured parameter: main a43b4e9b preserves ir.String storage for string destructuring. Retain that heldAs computation and add the branch generator noteLocal ownership/cycle tracking after allocating the local. No main line reverted.
- internal/lower/object.go, elementType contextual array hunk: main 49414912 adds the refusal for a union-typed array built from a narrower literal, preserving contextual weak storage. Keep that refusal and storage logic, while resolving the branch readonlyArrayView before the array check. The else-if header becomes separate contextual handling so normalization can happen first; no refusal is removed.
- internal/oracle/counts.md: restore the full main table, then regenerate exactly once on Linux with clang 20.1.8. No hand merge. All 59 changed keys are source-own changed keys (counts-rows.json); 55 additions, four modified rows.

Additional main reader repairs required by the new lowering: move the following byte-identical files from internal/oracle/testdata/review/refused/ to ../agree/. The review corpus README registers these through its globs, so no reader is deleted or disabled. The original refusal suite correctly failed because they are now accepted; the agreement suite now requires source Node, JavaScript, sanitized native, release native and leak agreement.

- fxspptb_iterators_hidden_return.a: agreement 1.02 s.
- fxspptb_iterators_override_this.a: agreement 2.01 s.

review-reclassification.json proves unchanged hashes. No header is needed on either checked program. The exact pinned Gate.aCheck rebuilt this compiler and accepts all 84 paths: 63 checked, 5 NotYet, 16 refused, 0 type errors (final.json).

Runtime clearance: no internal/native/runtime/*.c or *.h line changes at all against main. The source commit also has no runtime C/header changes, so the extra line list is empty.

Tests and commands:

All used sourced /workspace/adamic-tools/env.sh, GOMAXPROCS=4, GOFLAGS including -p=4, -count=1 and -parallel=4. Every test command wrote directly to its JSONL log and had a hard outer timeout. Each full go test used ./<package>/...; no test was removed.

- internal/flow: original 90 s package deadline after 563 passing top-level tests; all 312 unfinished tests passed in 13 bounded batches (-timeout 90s). Full inventory 875/875 proved in flow-completeness.json.
- internal/native/...: original 10 m package deadline with only TestAsyncHostPromisesThread (18 s in progress), TestOrdinaryToPrimitive (6 s) and TestQuarantineIsPerThreadUnderTSan (3 s) unfinished. Only those three were recovered in a bounded batch. All 477 current top-level tests across native and its subpackages have pass/skip results, proved by native-completeness.json. No completed leaf was over 60 s or repeated for a deadline.
- internal/lower/... (-timeout 10m), internal/fresh/..., internal/ir/..., internal/javascript/..., internal/load/... and stage3/project-loader/probe/... (-timeout 90s): full suites exit 0.
- internal/oracle/... (-timeout 14m): full run 752.538 s, all 797 top-level tests reached a terminal result; only TestReviewProgramsRefuse and its two now-accepted iterator children failed. After the directory repair, go test ./internal/oracle -run ^TestReviewPrograms -count=1 -parallel=4 -timeout 150s -json exited 0 in 33.727 s, including both newly registered agreement leaves. The current 797 top-level results are all pass/skip; validation.json retains original failures and corrected results.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -timeout 10m -json -args -update-counts: exit 0, 239.802 s; only one regeneration.
- go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -json -timeout 90s: pass.
- go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -json -timeout 90s: final post-repair pass (full ir suite also passed).
- go vet ./internal/flow/... ./internal/fresh/... ./internal/ir/... ./internal/javascript/... ./internal/load/... ./internal/lower/... ./internal/native/... ./internal/oracle/... ./stage3/project-loader/probe/...: exit 0; repeated after the path repair.
- gofmt -l on all changed Go files: no output. Production git diff --check passes. Historical review logs/patches and the verbatim Test262 license retain source whitespace.

Every completed test event is under 60 s; exact results and times are in test-seconds.json. All five requested guards are recorded in validation.json. The new compiler-consuming stage3/project-loader/probe remains declared in cloud/fast-gate/compiler-dependencies.json. Default skip reasons are retained in raw logs; they include WASI opt-ins, benchmark opt-ins and the stage3-only hook.

Conflict mutants (real package builds through overlays, sources never modified):

- union-layout: disable the retained main refusal; TestFX5NumberLiteralRefused fails with got <nil>, exit 1, 0.03 s.
- generator-owner: omit noteLocal for the whole destructured generator parameter; TestGeneratorDestructuredParameterCycle fails with got <nil>, exit 1, 0.09 s.

Both ordinary controls passed in the full lowering suite. Overlay inputs are .go.txt, with commands and logs retained. Existing source in-suite mutants also ran as part of the complete suites. No claim is made that external historical standalone mutant scripts were all rerun.

Every changed counts row, before to after (A/F/R/L/P/G):

- `internal/fresh/testdata/regexp_tree.ts`: 91/91/81/85/35/0 -> 91/91/85/89/35/0
- `internal/oracle/testdata/class_wrong_output_refused/iterators_hidden_return.a`: absent -> 28/28/16/20/8/0
- `internal/oracle/testdata/class_wrong_output_refused/iterators_override_this.a`: absent -> 57/57/31/49/9/0
- `internal/oracle/testdata/generators/basic.a`: absent -> 30/30/85/101/14/0
- `internal/oracle/testdata/generators/cancellation.a`: absent -> 23/23/65/74/14/0
- `internal/oracle/testdata/generators/captures.a`: absent -> 31/31/95/95/24/0
- `internal/oracle/testdata/generators/checker.a`: absent -> 20/20/104/93/16/0
- `internal/oracle/testdata/generators/close.a`: absent -> 27/27/57/59/13/0
- `internal/oracle/testdata/generators/completions.a`: absent -> 59/59/166/175/38/0
- `internal/oracle/testdata/generators/default_error.a`: absent -> 13/13/31/30/12/0
- `internal/oracle/testdata/generators/defaults.a`: absent -> 29/29/104/107/21/0
- `internal/oracle/testdata/generators/delegate.a`: absent -> 89/89/316/332/51/0
- `internal/oracle/testdata/generators/delegate_throw.a`: absent -> 28/28/158/151/22/0
- `internal/oracle/testdata/generators/first_reference.a`: absent -> 17/17/73/72/12/0
- `internal/oracle/testdata/generators/map_iterator.a`: absent -> 24/24/95/103/15/0
- `internal/oracle/testdata/generators/owned.a`: absent -> 32/32/132/130/21/0
- `internal/oracle/testdata/generators/reentrant.a`: absent -> 14/14/53/54/11/0
- `internal/oracle/testdata/generators/throw.a`: absent -> 16/16/75/74/12/0
- `internal/oracle/testdata/generic_body_indexed.a`: absent -> 2/2/0/2/2/0
- `internal/oracle/testdata/generic_body_read.a`: absent -> 10/10/0/10/3/0
- `internal/oracle/testdata/notyet_element_access/nodearray.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/notyet_element_access/sorted.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/notyet_element_access/template.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/regexp.a`: 437/437/382/417/58/0 -> 437/437/383/418/58/0
- `internal/oracle/testdata/user_iterators.a`: 663/663/455/904/64/0 -> 1050/1050/677/994/97/0
- `internal/oracle/testdata/user_iterators_rest_tdz.a`: 5/0/5/3/5/0 -> 11/3/7/3/8/0
- `stage3/fixtures/generics/01_identity.a`: absent -> 3/3/3/7/3/0
- `stage3/fixtures/generics/02_optional_return.a`: absent -> 5/5/3/9/3/0
- `stage3/fixtures/generics/03_callback_return.a`: absent -> 8/8/3/11/3/0
- `stage3/fixtures/generics/05_nested_class.a`: absent -> 8/8/5/16/7/0
- `stage3/fixtures/generics/09_recursive_optional.a`: absent -> 3/3/1/4/3/0
- `stage3/fixtures/generics/10_identifier_multimap.a`: absent -> 11/11/27/36/9/0
- `stage3/fixtures/generics/14_optional_literal_union.a`: absent -> 3/3/0/3/3/0
- `stage3/fixtures/generics/15_array_callback.a`: absent -> 18/18/7/27/4/0
- `stage3/fixtures/iteration-dispatch/body_throw.a`: absent -> 22/22/17/24/12/0
- `stage3/fixtures/iteration-dispatch/break.a`: absent -> 20/20/11/18/10/0
- `stage3/fixtures/iteration-dispatch/cached_next.a`: absent -> 29/29/17/34/12/0
- `stage3/fixtures/iteration-dispatch/class.a`: absent -> 22/22/11/22/7/0
- `stage3/fixtures/iteration-dispatch/close_throw_body.a`: absent -> 22/22/20/26/12/0
- `stage3/fixtures/iteration-dispatch/close_throw_break.a`: absent -> 20/20/16/21/10/0
- `stage3/fixtures/iteration-dispatch/completion.a`: absent -> 20/20/10/20/11/0
- `stage3/fixtures/iteration-dispatch/done_throw.a`: absent -> 13/13/14/16/10/0
- `stage3/fixtures/iteration-dispatch/labeled_continue.a`: absent -> 44/44/22/38/12/0
- `stage3/fixtures/iteration-dispatch/method_getters.a`: absent -> 26/26/16/24/15/0
- `stage3/fixtures/iteration-dispatch/next_throw.a`: absent -> 12/12/13/16/7/0
- `stage3/fixtures/iteration-dispatch/object.a`: absent -> 26/26/14/28/11/0
- `stage3/fixtures/iteration-dispatch/return.a`: absent -> 21/21/11/18/13/0
- `stage3/fixtures/iteration-dispatch/value_throw.a`: absent -> 13/13/14/16/10/0
- `stage3/fixtures/iteration/array_view_stress.a`: absent -> 34/34/50/72/16/0
- `stage3/fixtures/iteration/array_view_weak.a`: absent -> 5/5/8/13/4/0
- `stage3/fixtures/iteration/collections.a`: absent -> 35/35/24/47/11/0
- `stage3/fixtures/iteration/delegated_generator.a`: absent -> 20/20/70/74/15/0
- `stage3/fixtures/iteration/generator.a`: absent -> 14/14/33/35/13/0
- `stage3/fixtures/iteration/object_iteration.a`: absent -> 5/5/1/6/3/0
- `stage3/fixtures/iteration/strings.a`: absent -> 20/20/14/27/9/0
- `stage3/fixtures/iteration/test262_array_views.a`: absent -> 11/11/25/38/11/0
- `stage3/fixtures/iteration/user_forwarding.a`: absent -> 22/22/7/17/14/0
- `stage3/project-references-source/fixture/chain/app/main-print.a`: absent -> 2/2/0/2/2/0
- `stage3/project-references-source/fixture/two/app/main.a`: absent -> 2/2/0/2/2/0

Setup: initial run failed because compilation overlapped temporary conflict markers in functions.go/object.go. Resolved those markers and reran setup successfully, 45.121 s. House clang 20.1.8, Go 1.27.1, Node 24.19.0; nproc=5, cgroup cpu.max=400000 100000. Full setup timing/error lines are in command-output.json. Unused external Go cache was reclaimed for disk headroom; no source or evidence was removed.
