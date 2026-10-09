u127: 35 named tests, grouped into 12 rows.
Base: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2
Whole-package and combined slice baselines timed out; isolated clean rows passed without skips.
Four production mutants plus two probes; bounded uniqueness only.
Evidence branch: test-audit/stage1-cohere-markdownblocks-leaf_composition_products.

```json
[
  {
    "test": "TestProduct_MarkdownLeafGo family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/leaf_composition_products_test.go:28",
    "members": [
      "TestProduct_MarkdownLeafGoLists",
      "TestProduct_MarkdownLeafGoDocLayout"
    ],
    "seconds": 1.222,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S1 changed construction passed; this verdict rests on one construction edit and is not a claim that every failure is impossible.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownLeafGo family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestProduct_MarkdownLeafGo(Lists|DocLayout)$'; S1.log: --- PASS: TestProduct_MarkdownLeafGoDocLayout (1.33s)",
    "construction_mutants": [
      "S1"
    ],
    "construction_kills": []
  },
  {
    "test": "TestProduct_MarkdownLeafLowered",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/leaf_composition_products_test.go:38",
    "members": [
      "TestProduct_MarkdownLeafLowered"
    ],
    "seconds": 1.007,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S2 changed construction passed; this verdict rests on one construction edit and is not a claim that every failure is impossible.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownLeafLowered"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestProduct_MarkdownLeafLowered$'; S2.log: --- PASS: TestProduct_MarkdownLeafLowered (0.00s)",
    "construction_mutants": [
      "S2"
    ],
    "construction_kills": []
  },
  {
    "test": "TestProduct_MarkdownLeafNative family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/leaf_composition_products_test.go:43",
    "members": [
      "TestProduct_MarkdownLeafNativeSanitized",
      "TestProduct_MarkdownLeafNativeRelease"
    ],
    "seconds": 1.108,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S3 changed construction passed; this verdict rests on one construction edit and is not a claim that every failure is impossible.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_MarkdownLeafNative family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestProduct_MarkdownLeafNative(Sanitized|Release)$'; S3.log: --- PASS: TestProduct_MarkdownLeafNativeSanitized (1.24s)",
    "construction_mutants": [
      "S3"
    ],
    "construction_kills": []
  },
  {
    "test": "TestMarkdownListLayout_Setup",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/list_layout_shards_test.go:127",
    "members": [
      "TestMarkdownListLayout_Setup"
    ],
    "seconds": 0.228,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S4 changed construction passed; this verdict rests on one construction edit and is not a claim that every failure is impossible.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownListLayout_Setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^(TestMarkdownListLayout_Setup|TestMarkdownListLayout)$'; S4.log: --- PASS: TestMarkdownListLayout_Setup (0.22s)",
    "construction_mutants": [
      "S4"
    ],
    "construction_kills": []
  },
  {
    "test": "TestMarkdownListLayout family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/list_layout_shards_test.go:155",
    "members": [
      "TestMarkdownListLayoutUnion",
      "TestMarkdownListLayout_000",
      "TestMarkdownListLayout_001",
      "TestMarkdownListLayout_002",
      "TestMarkdownListLayout_003",
      "TestMarkdownListLayout_004",
      "TestMarkdownListLayout_005",
      "TestMarkdownListLayout_006",
      "TestMarkdownListLayout_007",
      "TestMarkdownListLayout_008",
      "TestMarkdownListLayout_009",
      "TestMarkdownListLayout_010",
      "TestMarkdownListLayout_011",
      "TestMarkdownListLayout_012",
      "TestMarkdownListLayout_013",
      "TestMarkdownListLayout_014",
      "TestMarkdownListLayout_015"
    ],
    "seconds": 14.603,
    "oracle": "Independent Go cohere bytes and retained fork Node results; self requirements for build success, exit status, framing, and built-in mutant difference. Full semantic bytes are checked; leak checks use status only.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3 list_layout_shards_test.go:698: source Node lists output byte 80 in generated/list/10./nnnn",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestMarkdownLayout family"
    ],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 66.328,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownListLayout family",
      "TestMarkdownLayout family",
      "TestMdastMalformedEvents family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMarkdownListLayout(Union|_00[0-9]|_01[0-5]))$; M3-list-family.log: list_layout_shards_test.go:698: source Node lists output byte 80 in generated/list/10./nnnn",
    "probe_passing_members": [
      "TestMarkdownListLayoutUnion"
    ],
    "vacuous_subcases": [],
    "witness_experiment": "W2",
    "subsumption_basis": 3
  },
  {
    "test": "TestMarkdownListLayout",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/lists_test.go:24",
    "members": [
      "TestMarkdownListLayout"
    ],
    "seconds": 0.214,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S4 changed construction passed; this verdict rests on one construction edit and is not a claim that every failure is impossible.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownListLayout"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^(TestMarkdownListLayout_Setup|TestMarkdownListLayout)$'; S4.log: --- PASS: TestMarkdownListLayout (0.00s)",
    "construction_mutants": [
      "S4"
    ],
    "construction_kills": []
  },
  {
    "test": "TestMarkdownQuoteLayout",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/lists_test.go:32",
    "members": [
      "TestMarkdownQuoteLayout"
    ],
    "seconds": 3.117,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S5 broken construction was rejected.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S5 quote_layout_shards_test.go:55: quote layout setup failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownQuoteLayout"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownQuoteLayout$'; S5.log: quote_layout_shards_test.go:55: quote layout setup failed",
    "construction_mutants": [
      "S5"
    ],
    "construction_kills": [
      "S5"
    ]
  },
  {
    "test": "TestMarkdownTableLayout",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/lists_test.go:38",
    "members": [
      "TestMarkdownTableLayout"
    ],
    "seconds": 3.847,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S6 changed construction passed; this verdict rests on one construction edit and is not a claim that every failure is impossible.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownTableLayout"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownTableLayout$'; S6.log: --- PASS: TestMarkdownTableLayout (3.69s)",
    "construction_mutants": [
      "S6"
    ],
    "construction_kills": []
  },
  {
    "test": "TestMarkdownLayout family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/lists_test.go:46",
    "members": [
      "TestMarkdownCodeBlockLayout",
      "TestMarkdownHTMLBlockLayout",
      "TestMarkdownRootLayout"
    ],
    "seconds": 66.328,
    "oracle": "Independent Go cohere bytes and retained fork Node results; self requirements for build success, exit status, framing, and built-in mutant difference. Full semantic bytes are checked; leak checks use status only.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3 lists_test.go:342: source Node lists output byte 64841 in stage1/cohere/markdownblocks/GAPS.md",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestMarkdownListLayout family"
    ],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 14.603,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownListLayout family",
      "TestMarkdownLayout family",
      "TestMdastMalformedEvents family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMarkdownCodeBlockLayout|TestMarkdownHTMLBlockLayout|TestMarkdownRootLayout)$; M3-block-layouts.log: lists_test.go:342: source Node lists output byte 64841 in stage1/cohere/markdownblocks/GAPS.md",
    "probe_passing_members": [],
    "vacuous_subcases": [
      "generated/root/edge/",
      "generated/root/edge/ \n"
    ],
    "witness_experiment": "W1",
    "subsumption_basis": 3
  },
  {
    "test": "TestMarkdownLeafComposition",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/lists_test.go:89",
    "members": [
      "TestMarkdownLeafComposition"
    ],
    "seconds": 5.732,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S7 broken construction was rejected.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S7 leaf_composition_independent_test.go:50: leaf composition preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownLeafComposition"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownLeafComposition$'; S7.log: leaf_composition_independent_test.go:50: leaf composition preparation failed",
    "construction_mutants": [
      "S7"
    ],
    "construction_kills": [
      "S7"
    ]
  },
  {
    "test": "TestMarkdownStructureLayout_Setup",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/lists_test.go:106",
    "members": [
      "TestMarkdownStructureLayout_Setup"
    ],
    "seconds": 5.15,
    "oracle": "Successful preparation is a self expectation; no independent semantic answer is compared. S8 broken construction was rejected.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S8 structure_layout_shards_test.go:121 (origin; raw scratch line 120): structure layout setup failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownStructureLayout_Setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownStructureLayout_Setup$'; S8.log: structure_layout_shards_test.go:121 (origin; raw scratch line 120): structure layout setup failed",
    "construction_mutants": [
      "S8"
    ],
    "construction_kills": [
      "S8"
    ]
  },
  {
    "test": "TestMdastMalformedEvents family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/malformed_events_independent_test.go:126",
    "members": [
      "TestMdastMalformedEvents_000",
      "TestMdastMalformedEvents_001",
      "TestMdastMalformedEvents_002",
      "TestMdastMalformedEventsUnion"
    ],
    "seconds": 6.596,
    "oracle": "Go cohere error messages plus retained fork Node diagnostics; self exit70, panic prefix and empty stdout. Full diagnostic comparison rejects a different error with the same exit status.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4 malformed_events_independent_test.go:193: mismatch expected error exit70 got0",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownListLayout family",
      "TestMarkdownLayout family",
      "TestMdastMalformedEvents family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMdastMalformedEvents(_00[0-2]|Union))$; M4-malformed-family.log: malformed_events_independent_test.go:193: mismatch expected error exit70 got0",
    "probe_passing_members": [
      "TestMdastMalformedEventsUnion"
    ],
    "vacuous_subcases": [],
    "witness_experiment": "W3"
  }
]
```

| ID | Origin file:line | Change | Failed grouped rows |
|---|---|---|---|
| M1 | stage1/cohere/markdownblocks/codeblocks.ts:20 | Math.max(4, longest + 1) | TestMarkdownLayout family, TestMarkdownListLayout family |
| M2 | stage1/cohere/markdownblocks/htmlblocks.ts:15 | code === 33 | TestMarkdownLayout family, TestMarkdownListLayout family |
| M3 | stage1/cohere/markdownblocks/root.ts:99 | drop final parts.push(arena.hardline()) | TestMarkdownLayout family, TestMarkdownListLayout family |
| M4 | stage1/cohere/markdownblocks/mdastCompile.ts:74 | previous.type === token.type | TestMdastMalformedEvents family |

No production survivors in the bounded matrix. Outside rows and repo-wide uniqueness remain unknown.

The brief and costs:

- The supplied historical commit is not the fetched origin/main. This audit uses ce1c5a2f, not 8de93800f4. All 35 names remain. malformed_events_shards_test.go moved to malformed_events_independent_test.go.
- The brief says 15 rows while listing 35 functions. The current shared-checker bodies group into 12 rows. Product Go and native recipes share helpers; code/HTML/root share testBlockLayout; list and malformed-event shards include their unions.
- I initially misgrouped code/HTML/root. Nine unnecessary member timing runs were preserved, then three proper family timings were collected. This avoidable error cost about ten minutes.
- The whole baseline timed out at 90.043s in list preparation; the 35-name bounded baseline timed out at 90.038s in structure preparation. Neither reported an assertion failure first. Isolated clean rows all passed.
- Some legacy names now only prepare products. They do not execute their advertised semantic comparisons. Their construction experiments are separated from production kills.
- Construction verdicts rest on one selected edit each. Passing a changed path or dropped preparation does not establish that every possible failure is impossible. Treat these as bounded findings.
- The first switch duplicated a built-in fence mutation anchor. Its neutral control failed at mutation site count. Those runs are archived under invalid-switch and excluded from final verdicts; the corrected neutral controls and full matrices were replayed. This avoidable instrumentation error cost a second matrix run.
- Shared block fixture failure prevents sibling wrappers reaching their own body; grouping removes a false impression of independent kills. Raw member failures are preserved.
- A static import closure is an upper bound on reachable functions, including type-only imports. Exact dynamic function coverage was not measured.
- Empty probes use production printDocument and MdastCompiler.compile entries, preserving protocol adapters. Component producers were not separately probed. Direct native P1 replay accepts two naturally empty positive root cases; their names are listed as vacuous_subcases, while the family is not vacuous. Union/setup members that pass a probe are explicitly listed and are not semantic-entry vacuity proofs.
- The runtime source switch avoids content changes per selection, but the uncached full block fixture still lowers and builds on each invocation. Each standalone diff also received its own native compile validation.
- The source-executed Node port is code under test, not an oracle. The independent Go cohere answers and retained fork Node results were unchanged. Direct standalone native output witnesses are saved separately.
- The four-mutant set is smaller than three per row, because most rows now construct products and native preparation is expensive. Subsumption rests on three mutants only, and is not a deletion recommendation.
- No scope row skipped. Optional whole-package width SDK/corpus coverage beyond the named slice was not audited. No repo-wide replay was run.

Warm tool setup: skipped; npm ci reported 458 ms; nproc 5. Isolated clean command wall time including redundant member runs: 1185.871s. Mutation/probe command wall time: 329.523s. Neutral controls: 164.217s. Construction experiments including vet: 221.470s. Standalone native build time: 262.283s. Two 90-second baselines and the archived invalid switch runs are additional. Full per-build timings and commands are in the logs and standalone-builds.json.
 Archived invalid mutation/probe runs: 397.016s, plus their failed neutral control.

The requested approximate thirty-minute port budget was exceeded. The redundant timing runs, repeated uncached large-fixture builds, neutral controls and standalone validation explain the overrun. Setup-to-finish UTC timestamps are in the logs. Production source and tests were restored; only evidence is committed.

Construction and weakened-check changes, excluded from production kills:

| ID | Origin file:line | Change | Failed requested members |
|---|---|---|---|
| S1 | stage1/cohere/markdownblocks/leaf_composition_independent_test.go:485 | return filepath.Join(product, "oracle") -> return filepath.Join(product, "") |  |
| S2 | stage1/cohere/markdownblocks/leaf_composition_products_test.go:24 | prepareLeafCompositionLowered(t, &p) -> drop statement |  |
| S3 | stage1/cohere/markdownblocks/leaf_composition_independent_test.go:44 | return filepath.Join(directory, "port") -> return filepath.Join(directory, "") |  |
| S4 | stage1/cohere/markdownblocks/list_layout_shards_test.go:820 | sanitized: manifest.Paths[2] -> sanitized: "" |  |
| S5 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:286 | products.goBinary = binary -> drop statement | TestMarkdownQuoteLayout |
| S6 | stage1/cohere/markdownblocks/lists_test.go:724 | return inputs, files -> return inputs[:0], files |  |
| S7 | stage1/cohere/markdownblocks/leaf_composition_independent_test.go:93 | leafCompositionPrepared, leafCompositionBuilds = fixture, products -> drop statement | TestMarkdownLeafComposition |
| S8 | stage1/cohere/markdownblocks/structure_layout_shards_test.go:117 | structureLayoutComplete = state -> drop statement | TestMarkdownStructureLayout_Setup |
| W1 | stage1/cohere/markdownblocks/lists_test.go:508 | if bytes.Equal(result.stdout, want.stdout) { -> if true { | TestMarkdownRootLayout/root_double_line, TestMarkdownCodeBlockLayout/fence_minimum, TestMarkdownRootLayout/ignored_source_span, TestMarkdownCodeBlockLayout/fence_longest, TestMarkdownRootLayout/ignore_range_closing_comment, TestMarkdownRootLayout, TestMarkdownCodeBlockLayout/code_indentation, TestMarkdownCodeBlockLayout, TestMarkdownHTMLBlockLayout/HTML_root_trim, TestMarkdownHTMLBlockLayout/HTML_comment_line, TestMarkdownHTMLBlockLayout/HTML_literal_root, TestMarkdownHTMLBlockLayout |
| W2 | stage1/cohere/markdownblocks/list_layout_shards_test.go:215 | if bytes.Equal(result.stdout, fixture.mutantWant) { -> if true { | TestMarkdownListLayout_012, TestMarkdownListLayout_005, TestMarkdownListLayout_009 |
| W3 | stage1/cohere/markdownblocks/malformed_events_independent_test.go:227 | if bytes.Equal(mutant.stderr, []byte("adamic: panic: "+string(truth.stdout))) { -> if true { | TestMdastMalformedEvents_002, TestMdastMalformedEvents_001, TestMdastMalformedEvents_000 |

Empty-answer probes, excluded from kills:

| ID | Origin entry | Empty answer | Failed rows |
|---|---|---|---|
| P1 | document.ts:213 printDocument | empty string | list and block layout families |
| P2 | mdastCompile.ts:619 compile | empty MdastArena | malformed-event family |

Construction survivors with observed witnesses:

- S1: returned regular executable before, directory after; Go product family passed.
- S2: generated C bytes 1,949,282 before, zero after; lowered product row passed.
- S3: returned regular executable before, directory after; native product family passed.
- S4: sanitized executable path present before, empty after; both list preparation rows passed.
- S6: 4,070 table inputs before, zero after; table preparation row passed.

The supplemental observation test is stored as text and was removed from the package. It only logs outputs; no verdict rests on inserted logging statements. Commands and observations are in construction-survivor-witnesses.json and S*-observe-*.log.

Standalone native builds:

- M1: 66.364s; exit 0.
- M2: 62.554s; exit 0.
- M3: 61.916s; exit 0.
- M4: 6.468s; exit 0.
- P1: 58.416s; exit 0.
- P2: 6.566s; exit 0.

Final UTC: 2026-10-09T13:47:24Z. Work exceeded the approximate port budget substantially, including about ten minutes of redundant member timings and the discarded switch matrix.
