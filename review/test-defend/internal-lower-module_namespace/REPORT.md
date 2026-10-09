Five rows defended with unique production catches; five remain cannot-judge within the seven-mutant cap.
Current main 76c59c81e8617cea1892a01841494895927a712c; 275 top-level tests, 36 added since audit, all ten target names present.
Tests unchanged; full-package matrices and bounded linear-work recovery, diffs, coverage and logs retained.

[
  {
    "test": "TestModuleNamespaceInitializedReadProof",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [],
    "prior_subsumed_by": [
      "TestUndecidedCycleReadsUseReadyChecks"
    ],
    "defense": "defended",
    "unique_mutant": "D01 internal/lower/load_time_reads.go:102",
    "attempts": [
      {
        "mutant": "D01",
        "file_line": "internal/lower/load_time_reads.go:102",
        "change": "retain readiness for qualified property reads but not plain imported identifiers",
        "rows_failed": [
          "TestModuleNamespaceInitializedReadProof"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/D01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; module_namespace_test.go:65: provider initialized before consumer, but the qualified read retained a readiness check"
  },
  {
    "test": "TestCallableNamespaceReceiverStaysLoud",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [],
    "prior_subsumed_by": [
      "TestNamespaceReceiverArrowInheritsScope"
    ],
    "defense": "defended",
    "unique_mutant": "D02 internal/lower/namespaces.go:152",
    "attempts": [
      {
        "mutant": "D02",
        "file_line": "internal/lower/namespaces.go:152",
        "change": "disable the receiver refusal for function and namespace declaration merging, keeping member-function refusal",
        "rows_failed": [
          "TestCallableNamespaceReceiverStaysLoud"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/D02 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; namespace_callable_test.go:35: got /tmp/adamic-gate/TestCallableNamespaceReceiverStaysLoud924089560/001/main.a:1:12: stage 0 can't lower an explicit callable-namespace this parameter; pass state explicitly yet"
  },
  {
    "test": "TestNamespaceClassEarlyConstructionStaysLoud",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestCallableNamespaceLimitsStayLoud"
    ],
    "prior_subsumed_by": [
      "TestCallableNamespaceLimitsStayLoud"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "go test -count=1 -timeout 90s ./internal/lower/ -run ^TestNamespaceClassEarlyConstructionStaysLoud$ -coverpkg=./internal/lower -coverprofile=coverage/TestNamespaceClassEarlyConstructionStaysLoud.out; clean coverage passes; 24 exclusive covered lines identify construction and new edges. No construction-specific mutant was run within the seven-mutant cap; cannot judge. The broad D07 union replacement does not make this row fail.",
    "reason": "24 exclusive covered lines identify construction and new edges. No construction-specific mutant was run within the seven-mutant cap; cannot judge. The broad D07 union replacement does not make this row fail.",
    "owner_finding": "Name matches the NotYet refusal assertion. It does not test native construction behavior, but does not promise to."
  },
  {
    "test": "TestNamespaceAmbientHostInitialization",
    "package": "internal/lower",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "prior_subsumed_by": [],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "go test -count=1 -timeout 90s ./internal/lower/ -run ^TestNamespaceAmbientHostInitialization$ -coverpkg=./internal/lower -coverprofile=coverage/TestNamespaceAmbientHostInitialization.out; clean coverage passes; 16 exclusive covered lines include host and ownership paths. The row has acquired an executable refusal assertion since audit. No host-specific mutant was run within the cap; cannot judge.",
    "reason": "16 exclusive covered lines include host and ownership paths. The row has acquired an executable refusal assertion since audit. No host-specific mutant was run within the cap; cannot judge.",
    "owner_finding": "Current row checks both ambient preflight and executable early-read refusal. Host lowering checks only error text, and cwd admits a separate ownership refusal. A different node: error can satisfy the host assertion."
  },
  {
    "test": "TestNamespaceAmbientContextsDoNotExecute",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [],
    "prior_subsumed_by": [
      "TestNamespaceCallGraphCycleUnion"
    ],
    "defense": "defended",
    "unique_mutant": "D03 internal/lower/namespaces.go:22",
    "attempts": [
      {
        "mutant": "D03",
        "file_line": "internal/lower/namespaces.go:22",
        "change": "drop declaration-file ambient authority while keeping syntax and inherited flags",
        "rows_failed": [
          "TestNamespaceAmbientContextsDoNotExecute"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/D03 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; namespaces_ambient_test.go:68: declaration-file namespace executes"
  },
  {
    "test": "TestNamespaceCallGraphLinearWork",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [],
    "prior_subsumed_by": [
      "TestNamespaceCallGraphCycleUnion"
    ],
    "defense": "defended",
    "unique_mutant": "D06 internal/lower/namespaces_call_graph.go:29",
    "attempts": [
      {
        "mutant": "D06",
        "file_line": "internal/lower/namespaces_call_graph.go:29",
        "change": "do not reuse completed empty reach sets, retaining active-cycle and nonempty memoization",
        "rows_failed": [
          "TestNamespaceCallGraphLinearWork"
        ],
        "combined_run_timeouts": [
          "TestEnumInitializationReach"
        ],
        "isolated_enum_exit": 0
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/D06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestNamespaceCallGraphLinearWork$/^12$; namespaces_call_graph_test.go:44: nonlinear function walks: got 8191, want 13; D06-rest.log times out in TestEnumInitializationReach/call-graph-24; D06-enum-alone.log passes its expensive subcase; D06-enum-whole.log passes the complete enum row. Other rows finish passing in D06-rest.log; the combined timeout does not prove an individual catch",
    "bounded": true,
    "bound": "Target depth 12 catches; target depth 24 exceeds 90s. Combined exclusion matrix cooks in enum row, but that complete row separately passes within 90s; every other top-level row passes in exclusion matrix."
  },
  {
    "test": "TestNamespaceCallGraphCycleUnion",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestEnumNamespaceSharedCycle"
    ],
    "prior_subsumed_by": [
      "TestEnumNamespaceSharedCycle"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D07",
        "file_line": "internal/lower/namespaces_call_graph.go:140",
        "change": "replace completed component union with each member direct reads; also exercises early-class reach propagation",
        "rows_failed": [
          "TestEnumInitializationGraphMemo",
          "TestEnumInitializationReach",
          "TestEnumLimitsStayLoud",
          "TestEnumNamespaceSharedCycle",
          "TestNamespaceCallGraphCycleUnion",
          "TestNamespaceInitializationReachability"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/D07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; namespaces_call_graph_test.go:67: partial cycle reach for one: 1 namespaces",
    "reason": "No exclusive covered lines. Three namespace-only cycle members differ semantically from the two-member enum/namespace subsumer. D07 reproduces a shared union failure, not a unique three-member boundary fault. Fewer than three aimed attempts; cannot judge.",
    "owner_finding": "Name matches exact namespace identities, full cycle membership and expansion count. No identified name/assertion mismatch."
  },
  {
    "test": "TestNamespaceEnumInitializationIndependentOfModuleAnalysis",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestEnumInitializationReach",
      "TestNamespaceLimitsStayLoud"
    ],
    "prior_subsumed_by": [
      "TestEnumInitializationReach"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D05",
        "file_line": "internal/lower/namespaces.go:252",
        "change": "change namespace preflight reachable-enum diagnostic while independent enum preflight stays intact",
        "rows_failed": [
          "TestEnumInitializationReach",
          "TestNamespaceEnumInitializationIndependentOfModuleAnalysis",
          "TestNamespaceLimitsStayLoud"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/D05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; namespaces_call_graph_test.go:86: namespace enum lost its independent refusal: /tmp/adamic-gate/TestNamespaceEnumInitializationIndependentOfModuleAnalysis4239742949/001/graph.a:1:44: stage 0 can't lower reading a value before its runtime initialization; move the call after the enum declaration yet",
    "reason": "D05 makes this row fail, but TestEnumInitializationReach and TestNamespaceLimitsStayLoud also fail. Fewer than three aimed attempts within cap; cannot judge uniqueness.",
    "owner_finding": "Name matches the direct namespace preflight call without module scheduling. Its oracle only demands an enum diagnostic substring; it does not separately assert module-analysis state."
  },
  {
    "test": "TestParserFactoryBindingHoisting",
    "package": "internal/lower",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "prior_subsumed_by": [],
    "defense": "defended",
    "unique_mutant": "D04 internal/lower/modules.go:100",
    "attempts": [
      {
        "mutant": "D04",
        "file_line": "internal/lower/modules.go:100",
        "change": "clear destructured namespace var metadata",
        "rows_failed": [
          "TestParserFactoryBindingHoisting"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/D04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; namespaces_factory_test.go:42: NamespaceVar bindings = map[], want factoryCreateNodeArray and factoryCreateNumericLiteral once each"
  },
  {
    "test": "TestTscNamespaceDeclarationShapes",
    "package": "internal/lower",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "prior_subsumed_by": [],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "go test -count=1 -timeout 90s ./internal/lower/ -run ^TestTscNamespaceDeclarationShapes$ -coverpkg=./internal/lower -coverprofile=coverage/TestTscNamespaceDeclarationShapes.out; clean coverage passes; 14 exclusive covered lines include optional namespace local initialization and empty erased declaration bodies. No shape-specific mutant was run within the cap; cannot judge.",
    "reason": "14 exclusive covered lines include optional namespace local initialization and empty erased declaration bodies. No shape-specific mutant was run within the cap; cannot judge.",
    "owner_finding": "Tsc names fixture provenance, not an executed tsc oracle. Positive cases assert only acceptance, not the lowered contents; overload policy is self-derived."
  }
]

Brief ambiguities, limits and time costs

1. The report is from older main f91994f019703ba25d2918cf529c0e0b0c05d93c. Current main adds 36 tests and strengthens at least three selected rows: initialized reads require executable output and a console write; ambient-host preflight now asserts an executable refusal; parser-factory now checks specific binding metadata and a nonempty body. Old vacuity claims do not transfer unchanged. All selected rows remain in their original files.
2. Ten rows times three attempts would require thirty mutants. The explicit near-minute full-matrix cap allows seven. I prioritized unique behavior and do not claim three honest aimed attempts for the remaining five rows. They remain cannot-judge, not deletion candidates. D07 is a shared-union fallback replay, not a new proof of a three-member-specific fault.
3. Exclusive coverage is a lead. The parser metadata row has zero exclusive lines but asserts metadata others do not inspect. The linear-work row has zero exclusive lines but feeds empty reach sets; enum memoization feeds nonempty sets. TestEnumInitializationReach includes the same empty diamond graph, revealed by a combined-run timeout; its full isolated row passes. It checks acceptance and does not assert body expansion counts. D06 changes actual reuse of empty cached graphs without altering the expansion counter.
4. D06's whole-package binary times out at 90.345 seconds. The timeout panic aborts the binary; do not use that original run for uniqueness. Depth 12 alone fails with 8191 expansions versus 13. The all-other-rows exclusion matrix also times out, identifying TestEnumInitializationReach/call-graph-24 as its only unfinished row. All its other top-level rows finish, except baseline skips. That enum subcase then passes alone in 57.551s, and the complete enum row passes alone in 63.091s too. The combined timeout therefore does not prove an individual row failure. The target depth-12 row fails while every other top-level row has an observed pass across the exclusion matrix and complete enum rerun. This is a bounded unique defense; target depth 24 remains over budget.
5. Defended means unique among the executed current package rows. TestOriginalCycleLedger and TestOptionalWideningCensus require external project inputs and skip throughout. One mixed-union subcase is deferred compiler implementation. These skipped inputs remain unknown; no toolchain installation enables them. Scope and raw logs preserve their names. No repo-wide uniqueness is claimed.
6. Seven standalone diffs use starting-main line positions, apply independently, and pass go vet ./internal/lower/. No test, harness, fixture, external oracle or checker was modified. D05 only changes diagnostic wording; it proves the text oracle fires but does not establish an independent soundness check.
7. The final schema has only defended, not defended and cannot-judge, while the prose introduces subsumed for newly caught formerly untrue rows. No conflict is needed here: the formerly untrue parser row is uniquely defended; the other two remain unattempted within the cap.
8. Automatic approval review initially failed with an internal Habitat service error before dependency installation. The same authorized operation succeeded on retry. No unsafe-action rejection or user approval was needed.
9. Go coverage instrumentation initially compiled for 10.38 seconds; later solo commands took about two seconds including Go overhead. Rest-of-package coverage required three additional full runs. Their exclusion regexes and profiles are saved. Independent source reads and test listing establish the current scope.

Owner findings for unresolved rows and performance scope

- TestNamespaceClassEarlyConstructionStaysLoud: Name matches the NotYet refusal assertion. It does not test native construction behavior, but does not promise to.
- TestNamespaceAmbientHostInitialization: Current row checks both ambient preflight and executable early-read refusal. Host lowering checks only error text, and cwd admits a separate ownership refusal. A different node: error can satisfy the host assertion.
- TestNamespaceCallGraphCycleUnion: Name matches exact namespace identities, full cycle membership and expansion count. No identified name/assertion mismatch.
- TestNamespaceEnumInitializationIndependentOfModuleAnalysis: Name matches the direct namespace preflight call without module scheduling. Its oracle only demands an enum diagnostic substring; it does not separately assert module-analysis state.
- TestTscNamespaceDeclarationShapes: Tsc names fixture provenance, not an executed tsc oracle. Positive cases assert only acceptance, not the lowered contents; overload policy is self-derived.

Timing and exclusions

Warm toolchain worked; setup skipped. npm ci reported 630ms; nproc 5; Go1.27.1. Clean baseline binary 41.705s. Coverage commands totaled 170.008s wall. Seven standalone vet checks totaled 4.104s. Seven full mutation commands totaled 392.353s wall, including compilation and native builds; bounded reruns totaled 220.017s. Each mutation uses ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/Dxx. Separate native rebuild durations are not emitted by these tests; command wall and binary elapsed are recorded, not inferred.

No more than seven production mutants, no full depth-24 result for D06, no mutants aimed specifically at the unresolved host, shape or construction leads, no three-attempt rejection verdicts or claimed uniqueness based only on timeout-aborted runs, no skipped external-input coverage, no other package or repo-wide replay. No tests were deleted, rewritten or weakened. Raw JSON logs may be losslessly gzip-compressed after report generation; commands name the original log destinations.

Evidence files: rows.json, code-and-oracle.md, scope.json, coverage-exclusive.json, coverage-runs.json, coverage/*.out, mutant-plan.json, matrix.json, D06-bounded-matrix.json, unique-pass-lists.json, diffs/*.diff, logs/*, and execution scripts. Production source is restored byte for byte before evidence commit.
