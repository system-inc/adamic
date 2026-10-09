Unit u111 at origin/main 16f436a16a8e3b9cd2343f449eb1b0b4577c135c: all 15 listed functions exist.
Grouped into 13 rows: 1 bounded sacred, 4 witnesses, 4 setup-checks, 1 untrue, 3 cannot-judge.
All three production mutants caught; no production survivors in the bounded matrix.
ProfileArtifacts passes its empty builder probe with every native artifact absent.
Evidence on test-audit/stage1-cohere-lint-profile_compilation_main; production source restored.

[
  {
    "test": "TestProfileCompilation family",
    "package": "stage1/cohere/lint",
    "file": [
      "stage1/cohere/lint/profile_compilation_main_test.go",
      "stage1/cohere/lint/profile_test.go"
    ],
    "seconds": 2.021,
    "oracle": "Live unchanged Go cohere wire/count output, compared with Node, emitted JavaScript, scanner, profiled and counted output; self-written corpus union invariants.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02",
      "M03"
    ],
    "unique_kills": [
      "M01",
      "M02",
      "M03"
    ],
    "last_proven_fail": "M03: profile_test.go:128: Node: case 0 line 2: port \"fixed\\tvar a;\\\\u000a\", Go \"/workspace/u111-tmp/TestProfileCompilation_0002686935148/003/no-var-0.ts:1:1\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "python3 /workspace/u111-tmp/switch_matrix.py (writes M03 to /workspace/u111-selector); timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestProfileCompilationUnion|TestProfileCompilationPlantedFailure|TestProfileCompilationBuildLower|TestProfileCompilationBuildC|TestProfileCompilationBuildJavaScript|TestProfileCompilationBuildNative|TestProfileCompilation_Setup|TestProfileCompilation_000|TestCommentFoldMutant|TestPositionIndexMutant|TestRegistrationMutant|TestFactoryHooks_Setup|TestNestedOutsideModuleCopy)$' > M03.log 2>&1; profile_test.go:128: Node: case 0 line 2: port \"fixed\\tvar a;\\\\u000a\", Go \"/workspace/u111-tmp/TestProfileCompilation_0002686935148/003/no-var-0.ts:1:1\"",
    "members": [
      "TestProfileCompilationUnion",
      "TestProfileCompilation_000"
    ]
  },
  {
    "test": "TestProfileCompilationBuildEmission family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_setup_clock_regression_test.go",
    "seconds": 0.289,
    "oracle": "Own emitted artifacts and successful serialization/write operations. No emitted answer is executed by these build wrappers.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S03: profile_compilation_main_test.go:325: open /workspace/u111-cache/S03/6f597cf5f80523855d7d0175780a5a69c2430ab48b800f6cf32f917b8b7d8fda/program.gob: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u111-tmp/checks/S03/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestProfileCompilationBuildLower|TestProfileCompilationBuildC|TestProfileCompilationBuildJavaScript|TestProfileCompilationBuildNative|TestProfileCompilation_Setup)$' > S03.log 2>&1; profile_compilation_main_test.go:325: open /workspace/u111-cache/S03/6f597cf5f80523855d7d0175780a5a69c2430ab48b800f6cf32f917b8b7d8fda/program.gob: no such file or directory",
    "members": [
      "TestProfileCompilationBuildC",
      "TestProfileCompilationBuildJavaScript"
    ],
    "setup_kills": [
      "S03"
    ]
  },
  {
    "test": "TestProfileCompilationPlantedFailure",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_setup_clock_regression_test.go",
    "seconds": 3.854,
    "oracle": "Go cohere comparison in child; parent expects exactly one failed compilation case and planted disagreement text.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: lint_setup_clock_regression_test.go:113: wrong planted-failure attribution: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u111-tmp/checks/W01/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestProfileCompilationPlantedFailure|TestCommentFoldMutant|TestPositionIndexMutant)$' > W01.log 2>&1; lint_setup_clock_regression_test.go:113: wrong planted-failure attribution: <nil>",
    "witness_kills": [
      "W01"
    ]
  },
  {
    "test": "TestProfileCompilationBuildLower",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_setup_clock_regression_test.go",
    "seconds": 0.218,
    "oracle": "Only completion of compilationLowered is checked. S03 still passes while writing lost.gob instead of program.gob; the returned path and IR contents are not checked.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "S03.log: BuildLower PASS despite missing program.gob",
    "construction_escapes": [
      "S03"
    ],
    "resting_on": "three production mutants and one construction filename fault; no deletion recommendation"
  },
  {
    "test": "TestProfileCompilationBuildNative",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_setup_clock_regression_test.go",
    "seconds": 0.453,
    "oracle": "Successful builds and reading/copying scanner, counted and profiled files. No runtime answer comparison here.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S04: profile_compilation_main_test.go:243: open /workspace/u111-cache/S04/258835ad696fbc9086060bf8335dc2e4cc3e719a2d563fef1f21fac3af0d160c/scanner: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u111-tmp/checks/S04/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestProfileCompilationBuildNative|TestProfileCompilation_Setup)$' > S04.log 2>&1; profile_compilation_main_test.go:243: open /workspace/u111-cache/S04/258835ad696fbc9086060bf8335dc2e4cc3e719a2d563fef1f21fac3af0d160c/scanner: no such file or directory",
    "setup_kills": [
      "S03",
      "S04"
    ]
  },
  {
    "test": "TestProfileCompilation_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_setup_clock_regression_test.go",
    "seconds": 1.752,
    "oracle": "Preparation must finish without errors. It prepares Go oracle outputs but does not compare candidate answers.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S03: profile_compilation_main_test.go:325: open /workspace/u111-cache/S03/6f597cf5f80523855d7d0175780a5a69c2430ab48b800f6cf32f917b8b7d8fda/program.gob: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u111-tmp/checks/S03/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestProfileCompilationBuildLower|TestProfileCompilationBuildC|TestProfileCompilationBuildJavaScript|TestProfileCompilationBuildNative|TestProfileCompilation_Setup)$' > S03.log 2>&1; profile_compilation_main_test.go:325: open /workspace/u111-cache/S03/6f597cf5f80523855d7d0175780a5a69c2430ab48b800f6cf32f917b8b7d8fda/program.gob: no such file or directory",
    "setup_kills": [
      "S03",
      "S04"
    ]
  },
  {
    "test": "TestProfileArtifacts",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/profile_test.go",
    "seconds": null,
    "oracle": "Git checkout pin plus own artifact construction succeeds. Empty buildProfile probe passes with all native artifacts absent in a fresh directory.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "P02.log: PASS with absent main.c/scanner/counted/profiled; clean-TestProfileArtifacts-0.log timed out at 90s",
    "probe_ids": [
      "P02"
    ],
    "seconds_lower_bound": 90
  },
  {
    "test": "TestProfileSnapshotsAgree",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/profile_test.go",
    "seconds": null,
    "oracle": "Intended live Go cohere agreement over upstream/compiler/stage1 corpus and snapshot backends; enabled clean run timed out before completing agreement.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "clean-TestProfileSnapshotsAgree-0.log: panic: test timed out after 1m30s",
    "seconds_lower_bound": 90
  },
  {
    "test": "TestCommentFoldMutant",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/profile_test.go",
    "seconds": 7.841,
    "oracle": "Live Go cohere output; built-in fold mutation must disagree on Node and emitted JavaScript.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: profile_test.go:244: fold mutant survived on Node",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u111-tmp/checks/W01/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestProfileCompilationPlantedFailure|TestCommentFoldMutant|TestPositionIndexMutant)$' > W01.log 2>&1; profile_test.go:244: fold mutant survived on Node",
    "witness_kills": [
      "W01"
    ]
  },
  {
    "test": "TestPositionIndexMutant",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/profile_test.go",
    "seconds": 7.586,
    "oracle": "Live Go cohere output; built-in anchor mutation must disagree on Node and emitted JavaScript.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: profile_test.go:263: position mutant survived on Node",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u111-tmp/checks/W01/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestProfileCompilationPlantedFailure|TestCommentFoldMutant|TestPositionIndexMutant)$' > W01.log 2>&1; profile_test.go:263: position mutant survived on Node",
    "witness_kills": [
      "W01"
    ]
  },
  {
    "test": "TestRegistrationMutant",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/registration_test.go",
    "seconds": 7.659,
    "oracle": "Live Go cohere output for no-debugger; built-in descriptor mutation must disagree on Node and emitted JavaScript.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W02: registration_test.go:57: listener omission survived on Node",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u111-tmp/checks/W02/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestRegistrationMutant)$' > W02.log 2>&1; registration_test.go:57: listener omission survived on Node",
    "witness_kills": [
      "W02"
    ]
  },
  {
    "test": "TestFactoryHooks_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/registration_test.go",
    "seconds": 4.149,
    "oracle": "Own fixture pair and prepared-product construction, including exactly four directory/manifest rows and nonnil prepared state.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S02: factory_hooks_shards_test.go:203: incomplete factory hook fixture setup",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u111-tmp/checks/S02/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestFactoryHooks_Setup)$' > S02.log 2>&1; factory_hooks_shards_test.go:203: incomplete factory hook fixture setup",
    "setup_kills": [
      "S02"
    ]
  },
  {
    "test": "TestNestedOutsideModuleCopy",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/registration_test.go",
    "seconds": 0.609,
    "oracle": "Own copied newline expectation on Node/native; built-in bad rewrite must produce Node ERR_MODULE_NOT_FOUND and a load.Load error. No independent comparison helper guards the negative checks.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFactoryHooks_Setup",
      "TestProfileCompilation family",
      "TestProfileCompilationBuildEmission family",
      "TestNestedOutsideModuleCopy",
      "TestProfileCompilationBuildLower",
      "TestProfileCompilation_Setup",
      "TestProfileCompilationBuildNative",
      "TestProfileCompilationPlantedFailure",
      "TestPositionIndexMutant",
      "TestCommentFoldMutant",
      "TestRegistrationMutant"
    ],
    "evidence": "clean-TestNestedOutsideModuleCopy-{0,1,2}.log: PASS; W03-not-run.txt explains unavailable permitted weakening"
  }
]

| ID | Starting origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | stage1/cohere/lint/main.ts:67 | const column = start - (offsets[lines[line - 1] ?? 0] ?? 0) + 1; -> const column = start - (offsets[lines[line - 1] ?? 0] ?? 0) + 2; (change constant) | TestProfileCompilation family, TestProfileCompilationPlantedFailure |
| M02 | stage1/cohere/lint/main.ts:137 | drop fixed declaration and fixed-output log (drop statement) | TestProfileCompilation family, TestProfileCompilationPlantedFailure |
| M03 | stage1/cohere/lint/lint.ts:116 | rules.visit(index, parent); ->  (drop statement) | TestProfileCompilation family, TestProfileCompilationPlantedFailure |
| W01 | stage1/cohere/lint/lint_test.go:158 | func difference(got, want []byte) string { return "" } (weaken agreement comparison) | TestProfileCompilationPlantedFailure, TestPositionIndexMutant, TestCommentFoldMutant |
| W02 | stage1/cohere/lint/registration_test.go:56 | equality becomes reflexive true (weaken equality comparison) | TestRegistrationMutant |
| S01 | stage1/cohere/lint/profile_compilation_main_test.go:32 | const testProfileCompilationShards = 2 (change construction constant) | TestProfileCompilation family |
| S02 | stage1/cohere/lint/factory_hooks_shards_test.go:227 | for index := 0; index < 1; index++ { (off-by-one construction bound) | TestFactoryHooks_Setup |
| S03 | stage1/cohere/lint/profile_compilation_main_test.go:312 | os.Create(filepath.Join(dir, "lost.gob")) (change construction filename) | TestProfileCompilationBuildEmission family, TestProfileCompilationBuildNative, TestProfileCompilation_Setup |
| S04 | stage1/cohere/lint/profile_compilation_main_test.go:203 | native.Build(string(source), filepath.Join(dir, kind+"-lost"), options) (change construction filename) | TestProfileCompilationBuildNative, TestProfileCompilation_Setup |
| P02 | stage1/cohere/lint/profile_test.go:62 | buildProfile returns at entry (empty construction probe) |  |
| P01 | stage1/cohere/lint/main.ts:11 | run returns 0 at entry (empty port answer probe) | TestProfileCompilation family, TestProfileCompilationPlantedFailure |

Survivors: none among M01-M03. Production failures in TestProfileCompilationPlantedFailure are broken agreement preconditions and never count as witness kills or as a second production defender. S03 is caught by consumers but escapes its producer BuildLower. P02 is a probe, not a surviving mutant.

CODE UNDER TEST and ORACLE were named before applying mutations: port main.ts answer formatting and lint.ts Linter.walk; own construction/comparison helpers for setup/witness checks; unchanged Go cohere for semantic answers and self-written construction/attribution contracts. No Go cohere matcher, descriptor oracle adapter, Node implementation, fixture or compiler production implementation was mutated. Three menu mutations were fixed before catches. M02 drops both the fixed binding and its print to avoid an unused binding. All switches are instrumentation; standalone diffs have no switch.

Clarifications, process deviations and costs:

- The heading says 14 rows but the list contains 15 functions. All exist at the newly fetched starting commit, which differs from the brief's old 8de93800f4. Six compilation functions moved to lint_setup_clock_regression_test.go. Per the expanded family rule, Union plus _000 is one family, and BuildC plus BuildJavaScript are one per-kind emission family. Setup and Native/Lower builders have additional product contracts, so remain separate. There are 13 grouped rows, of which 11 entered the production matrix.
- A clean whole-package run hit the binary's 90-second timeout while the large mutant suite ran. No prior assertion failure was observed. The runnable assigned slice, individual samples and switched neutral control passed. Both full-package uniqueness and repo uniqueness are unknown. unique_kills names only bounded uniqueness, excluding witness precondition failures as instructed. excluded-potential-callers.txt saves the wider callers for central replay.
- The artifact row cold-timed out after producing scanner but before completing counted/profiled products. Missing products were completed from its already emitted C using native runtime clang flags, with independent deadlines. The enabled snapshot row still timed out. It has no completed baseline or mutant matrix, and cannot be judged within budget. Artifacts and snapshots have null medians, not invented 90-second medians. The cold planted-failure row also timed out; after independently warming its products it passed three retries, whose median is reported separately from the cold timeout.
- Build-only rows check production and consumption of artifacts rather than runtime semantics. All three valid semantic mutants pass BuildLower, emission builders, Native and Setup. Construction faults establish the setup-check verdicts separately. BuildLower also passes S03, which creates lost.gob with no program.gob. Its untrue label rests only on the sampled construction/semantic faults, and is not a claim that it cannot ever fail on a compiler error. Do not delete it from this bounded finding.
- The empty buildProfile probe uses a fresh /workspace/u111-empty-profile directory. TestProfileArtifacts passes, while main.c, scanner, counted and profiled are all absent (checks.json). It still copies port sources, writes the compiler manifest and copies the oracle before calling the empty builder; the measured absence concerns precisely the builder's artifacts. Vacuity is proven, but no ordinary semantic mutant could be judged against its over-budget baseline, so the verdict remains cannot-judge. Other unprobed entries have null vacuity. P01 judges only the family that executes run, not rows that merely compile it or witnesses with broken preconditions.
- CommentFold/PositionIndex/Registration witnesses can pass when a separate production mutation changes output for another reason. Production survival therefore says nothing about witness strength. W01/W02 show their designated checks fail when agreement is weakened. The planted-failure parent demands one child failure and specific attribution; W01 makes the child pass and correctly causes that parent to fail.
- NestedOutsideModuleCopy combines positive import-copy checks with inline negative checks directly against Node's module loader and load.Load. There is no separate agreement helper to weaken for that negative witness. Making its own assertions always fail would be tautological, and changing Node/load.Load would mutate the oracle for the copy harness. No such edit was attempted. Its median is measured; witness strength is cannot-judge, not inferred from its name or green run.
- The production transitive import inventory (181 files, 618 named declarations) was saved before mutant definition. It conservatively includes registered rules and imported parser code; it is not dynamic path coverage. Relevant setup/comparison bodies were read and listed before fault planning; the complete per-row static harness reach artifact was finalized after execution, which falls short of saving the whole harness inventory before planning. This ordering deviation is recorded rather than presented as prior evidence.
- Three port mutants, one port probe, two weakened comparisons, four construction faults and one construction probe were used. They are explicitly separated. The approximate three-per-row target applies here to the one semantic port family. No construction/witness fault contributes sacred or subsumed evidence. Subsumption is not inferred from zero-kill rows.
- Warm tools worked, so setup was skipped. npm ci ran in stage3/api before baseline; no additional node_modules directory is used by these tests. The exact v6.0.3 TypeScript checkout was installed and its Git HEAD checked against the test's pin. Artifact/snapshot opt-ins were enabled. Outside-slice throughput/JSX-throughput opt-ins skipped in the whole baseline. No assigned row was skipped; two timed out instead.
- Standalone verification used applied diffs in isolated main/lint source copies, changing only import locations to the identical absolute graph so relative imports resolve. Every diff applies to the starting index. Port diffs were verified with the real Adamic C command and runtime clang flags; Go comparison/construction diffs passed go vet overlays. The first combined native M01 verification cooked at 90 seconds while other checks ran. It was narrowed to separate C-emission and clang stages. Each narrowed stage stayed below 90 seconds, but combined rebuild times exceeded 90. No compiler mutant was planted; runtime selector changes reuse one switched compiled product graph. Construction writer mutations had isolated cache directories, with only unaffected IR/emission products seeded for S04, so caches cannot conceal changed product writers.
- Timing samples are the test binary's own package ok line, three separate count=1 runs per completed row/family. Cold setup costs are retained in logs. Artifact recovery and warm planted retries briefly overlapped other timing work; medians reflect the observed workspace load, not an isolated benchmark. nproc is 5, and cpu.max permits 4 CPU cores.

Timing evidence:

- Setup: 0 s; npm ci command 0.555 s; list command 7.519 s; compiler checkout 15.606 s.
- Whole baseline command 93.423 s; standalone baseline/timing commands 518.167 s; warm planted retries 18.663 s; family timings 19.620 s.
- Switched C build 66.298 binary s; switched Native build 18.571 binary s. Neutral control and production/probe command total 133.115 s. Witness/construction/probe commands 107.835 s, vet commands 3.409 s.
- M02 standalone native rebuild: emission 84.545 s, clang 13.637 s; statuses 0/0.
- M01 standalone native rebuild: emission 89.719 s, clang 13.828 s; statuses 0/0.
- P01 standalone native rebuild: emission 83.617 s, clang 5.401 s; statuses 0/0.
- M03 standalone native rebuild: emission 89.970 s, clang 13.739 s; statuses 0/0.

Not covered: whole-package mutant replay, repository uniqueness, complete snapshot agreement, exhaustive runtime paths, primary empty probes for Lower/C/JavaScript/Native.Build, and a permitted weakening of the nested module-loader witness. Production source restored; final restored runnable-slice check is logged. Total work about 38 minutes, exceeding the approximate 30-minute port budget to finish independent diff compilation and publish complete evidence.
