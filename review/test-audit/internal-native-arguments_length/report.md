Audit u045 at 2b1be38362046455e0e5454d8b0e674a8630e95d.
14 names exist; no moves or vanished names. No requested row skipped.
Bounded verdicts: 8 sacred, 2 subsumed, 3 witness, 1 untrue.
17 production mutants, 8 entry probes, 1 weakened-check run.
Whole package timed out at 90.047s; uniqueness outside names.json is unknown.

Full per-row deliverable: rows.json. Matrix: matrix.json. All individual logs include JSON test results. observed-results.json lists the rows observed to pass, fail or skip per run. Standalone diffs are in diffs/, against the starting main commit. switch.diff records the scratch switch; sources and existing tests are restored. Every Go mutant/probe/witness diff was independently vetted. Every C mutant was compiled by the actual sanitized runtime build with native Flags, while no Go mutation was selected.

Survivors, all observed with an additional scratch observation probe (never part of the kill matrix):
M06: captured-root borrows=0 before, 1 after. Removing the root capture refusal changes the borrow planner, unguarded within this matrix.
M08: region marked=4 ineligible-marked=0 before, marked=21 ineligible-marked=21 after. The row checks nonempty region/spread counts and some freshness/escape/consumption facts, but did not reject marking the wrong statements. Unguarded within this matrix.
M11: fixture stdout "77\n2\n3\n" before, "77\n3\n4\n" after. The convention/arity tests do not execute this fixture to verify actual argument counts. Unguarded within this matrix.
No equivalent candidates.

Vacuity: the parser's emitted-C negative check passes empty C. It checks three IR function names but no positive emitted-C witness, so that construction still passes when the emitter answers nothing. RuntimeFeaturesIgnoreLiterals passes P_C but fails P_LIBRARY: the two entry results are explicitly separate in probe_entries/probe_kills. Unicode metadata has no callable native answer entry, so vacuity is null. NoReader's plain fixture passes empty C, while counted fixtures fail it.

Witnesses: W_BUILD runs clang but discards its error in Build; all three rows fail their required rejection assertions. Their production-mutant failures are broken preconditions and excluded from kills and verdicts. The optional-method row also contains a built-in missing-thunk witness, but its normal Node-versus-native check fails M12, independently proving its production defense.

Brief friction and practical limits:
* The named historical commit differs from current origin/main. I used current main, with all locations anchored to 2b1be383.
* Entire native package has large corpus and opt-in suites and exceeded 90 seconds. The matrix was narrowed to the 14 unit rows, established as reaching the inventoried functions by coverage and caller leads. Other callers were saved in caller-leads.txt but not run. This cannot establish package-wide uniqueness; sacred means unique only within this bounded matrix.
* Three rows are witnesses although they are not named Mutants; they generate invalid C and assert clang rejects it. Production failures cannot decide their witness verdicts.
* A row can call two main entries. RuntimeFeaturesIgnoreLiterals has opposite empty-C and empty-library probe results. The JSON reports both and sets vacuous true for the empty-C facet.
* Unicode-version metadata is a constant check with no native-answer function to probe. No invented probe was applied to Node or embed.FS.
* Rebuild timings are the measured test section from CONT to the first case sweep, including Build and tiny fixture preparation. Isolated clang-only time was not instrumented. Three C mutations required source-specific rebuilds.
* Exhaustive runtime coverage was not collected: Go coverage inventories native package functions; C case reach was read from source. All output went to log files. No other package's tests ran.
* 17 mutants is a bounded sample, not proof that an untrue row is impossible to make fail. The negative parser check's missing positive C assertion is a concrete owner finding; no test was rewritten or deleted.

Timing: setup skipped (warm tools), npm reports 627ms, nproc=5. Whole baseline 90.047s, bounded baseline 15.855s; 42 solo runs sum to 63.299 binary seconds. Matrix/probe/check and standalone vet wall totals and three rebuild sections are in timings.json. Sources restored before evidence commit.

Mutant table

| id | origin/main file:line | change | failed rows |
|---|---|---|---|
| M01 | internal/native/borrow.go:70 | flip pass-through consumer decision | TestPassThroughsAreNotConsumers |
| M02 | internal/native/borrow.go:186 | change initial chain safety constant | TestBorrowChainDeclarations, TestBorrowChainTargets |
| M03 | internal/native/borrow.go:200 | flip field-name mutation guard | TestBorrowChainDeclarations, TestBorrowChainTargets |
| M04 | internal/native/borrow.go:214 | flip unknown callback guard | TestBorrowChainTargets |
| M05 | internal/native/element_borrow.go:40 | change borrow acceptance condition | TestBorrowChainDeclarations |
| M06 | internal/native/element_borrow.go:276 | drop captured-root refusal |  |
| M07 | internal/native/reuse.go:131 | drop virtual ownership propagation loop by empty iteration list | TestInheritanceMemoryPlans |
| M08 | internal/native/region.go:86 | flip region eligibility |  |
| M09 | internal/native/region.go:134 | change class freshness result | TestInheritanceMemoryPlans |
| M10 | internal/native/emit.go:52 | flip counted convention feature emission | TestClosureConventionDropCount, TestClosureConventionWrongOrder, TestNoReaderCallingConvention, TestOptionalMethodThunksMatchNode |
| M11 | internal/native/arguments_length.go:179 | off-by-one packed actual count |  |
| M12 | internal/native/emit_objects.go:326 | drop proven structural thunk requirement | TestOptionalMethodThunksMatchNode |
| M13 | internal/native/native.go:86 | change floating point contraction option | TestArithmeticIsNeverFused |
| M14 | internal/native/runtime/case_tables.h:8 | change Unicode version constant | TestCaseTablesMatchNodesUnicode |
| M15 | internal/native/runtime/case.c:42 | off-by-one simple table start bound | TestCaseMappingMatchesNode |
| M16 | internal/native/runtime/case.c:177 | off-by-one full mapping iteration bound | TestCaseMappingMatchesNode |
| M17 | internal/native/library.go:49 | flip feature marker selection | TestClosureConventionRuntimeDropCount, TestClosureConventionRuntimeFeaturesIgnoreLiterals, TestOptionalMethodThunksMatchNode |
