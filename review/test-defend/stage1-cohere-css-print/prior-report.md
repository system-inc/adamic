u079: 15 named functions exist at cf79ecec; none moved or vanished from the named files.
10 grouped rows: 6 subsumed, 3 untrue construction checks, 1 witness.
All four production mutants caught; no bounded unique kills or production survivors.
P2 passes the media-query entry; P4 fails the RegExp entry. Darwin product member skipped on Linux.
Evidence: test-audit/stage1-cohere-css-print, review/test-audit/stage1-cohere-css-print/.

```json
[
  {
    "test": "TestClosedPrinterRegexGap",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/print_test.go:56",
    "seconds": 1.099,
    "oracle": "Node and both backends must equal the self-written literal 2\\nOk\\n. parse(screen).kind is checked, but its tree contents are ignored; P2 passes.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3 print_test.go:67: native: 0 \"2\\nError\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCSSProfileSnapshotsAgree"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P4"
    ],
    "subsumer_seconds": 37.794,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestClosedPrinterRegexGap",
      "TestCSSPrinterThroughput",
      "TestCSSPrinterBoundaryProofs",
      "TestOptionalBooleanPrinterMatchesGo",
      "TestCSSProfileSnapshotsAgree",
      "TestSharedSliceAppendAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M3; selector /tmp/u079/selector=M3; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^Test(ClosedPrinterRegexGap|CSSPrinterBoundaryProofs|OptionalBooleanPrinterMatchesGo|SharedSliceAppendAgreesWithNode)$; print_test.go:67: native: 0 \"2\\nError\\n\"; log M3-fast.log",
    "subsumption_mutants": 1,
    "vacuous_entries": {
      "mediaquery.parse": true,
      "adamic_regex_test": false
    },
    "vacuous_subcases": [
      "P2: media-query assertion accepts an empty Ok tree; all three backends still print Ok."
    ]
  },
  {
    "test": "TestCSSPrinterThroughput",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/print_test.go:145",
    "seconds": 77.027,
    "oracle": "Go cohere and the pinned Prettier fork choose shared successful cases. Native, source Node, npm Prettier and fork summaries must match formatted counts and aggregate UTF-16 lengths. Equal-length wrong text could pass this checksum.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4 print_test.go:211: native: 70 \"\" adamic: panic: RangeError: Invalid code point 4249537",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCSSProfileSnapshotsAgree"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 37.794,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestClosedPrinterRegexGap",
      "TestCSSPrinterThroughput",
      "TestCSSPrinterBoundaryProofs",
      "TestOptionalBooleanPrinterMatchesGo",
      "TestCSSProfileSnapshotsAgree",
      "TestSharedSliceAppendAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M4; selector /tmp/u079/selector=M4; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^TestCSSPrinterThroughput$; print_test.go:211: native: 70 \"\" adamic: panic: RangeError: Invalid code point 4249537; log M4-throughput.log",
    "subsumption_mutants": 4
  },
  {
    "test": "TestCSSPrinterBoundaryProofs",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/print_test.go:241",
    "seconds": 1.244,
    "oracle": "Go cohere directly formats four boundary cases; port source on Node must match complete answers. npm Prettier is run for separate logged boundary proofs.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1 print_test.go:258: boundary proof: 0 \"\\\"a {\\\\n   b: c;\\\\n}\\\\n\\\"\\n\\\"// x\\\\ra {\\\\n}\\\\n\\\"\\nerror {\\\"column\\\":1,\\\"endColumn\\\":2,\\\"endLine\\\":1,\\\"endOffset\\\":2,\\\"line\\\":1,\\\"name\\\":\\\"CssSyntaxError\\\",\\\"offset\\\":0,\\\"reason\\\":\\\"Unknown word \\u00a0\\\"}\\nerror \\\"css: yaml front matter is formatted by the yaml printer, which css.Format cannot reach\\\"\\n\" ; Go \"\\\"a {\\\\n  b: c;\\\\n}\\\\n\\\"\\n\\\"// x\\\\ra {\\\\n}\\\\n\\\"\\nerror {\\\"column\\\":1,\\\"endColumn\\\":2,\\\"endLine\\\":1,\\\"endOffset\\\":2,\\\"line\\\":1,\\\"name\\\":\\\"CssSyntaxError\\\",\\\"offset\\\":0,\\\"reason\\\":\\\"Unknown word \\u00a0\\\"}\\nerror \\\"css: yaml front matter is formatted by the yaml printer, which css.Format cannot reach\\\"\\",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestOptionalBooleanPrinterMatchesGo"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 26.151,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestClosedPrinterRegexGap",
      "TestCSSPrinterThroughput",
      "TestCSSPrinterBoundaryProofs",
      "TestOptionalBooleanPrinterMatchesGo",
      "TestCSSProfileSnapshotsAgree",
      "TestSharedSliceAppendAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M1; selector /tmp/u079/selector=M1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^TestCSSPrinterBoundaryProofs$; print_test.go:258: boundary proof: 0 \"\\\"a {\\\\n   b: c;\\\\n}\\\\n\\\"\\n\\\"// x\\\\ra {\\\\n}\\\\n\\\"\\nerror {\\\"column\\\":1,\\\"endColumn\\\":2,\\\"endLine\\\":1,\\\"endOffset\\\":2,\\\"line\\\":1,\\\"name\\\":\\\"CssSyntaxError\\\",\\\"offset\\\":0,\\\"reason\\\":\\\"Unknown word \\u00a0\\\"}\\nerror \\\"css: yaml front matter is formatted by the yaml printer, which css.Format cannot reach\\\"\\n\" ; Go \"\\\"a {\\\\n  b: c;\\\\n}\\\\n\\\"\\n\\\"// x\\\\ra {\\\\n}\\\\n\\\"\\nerror {\\\"column\\\":1,\\\"endColumn\\\":2,\\\"endLine\\\":1,\\\"endOffset\\\":2,\\\"line\\\":1,\\\"name\\\":\\\"CssSyntaxError\\\",\\\"offset\\\":0,\\\"reason\\\":\\\"Unknown word \\u00a0\\\"}\\nerror \\\"css: yaml front matter is formatted by the yaml printer, which css.Format cannot reach\\\"\\; log M1-rerun-TestCSSPrinterBoundaryProofs.log",
    "subsumption_mutants": 1
  },
  {
    "test": "TestOptionalBooleanPrinterMatchesGo",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/print_test.go:296",
    "seconds": 26.151,
    "oracle": "Go cohere supplies complete six-case printer answers; Node source, native sanitizers and JavaScript backend must match exactly.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4 print_test.go:317: native ASan/UBSan: 0  line 2, byte 1: \"error \\\"css: Unexpected PostCSS node type: \\\\\\\"\\\\\\\".\\\"\", Go cohere \"\\\"a {\\\\n  b: c;\\\\n}\\\\n\\\"\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCSSProfileSnapshotsAgree"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 37.794,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestClosedPrinterRegexGap",
      "TestCSSPrinterThroughput",
      "TestCSSPrinterBoundaryProofs",
      "TestOptionalBooleanPrinterMatchesGo",
      "TestCSSProfileSnapshotsAgree",
      "TestSharedSliceAppendAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M4; selector /tmp/u079/selector=M4; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^Test(ClosedPrinterRegexGap|CSSPrinterBoundaryProofs|OptionalBooleanPrinterMatchesGo|SharedSliceAppendAgreesWithNode)$; print_test.go:317: native ASan/UBSan: 0  line 2, byte 1: \"error \\\"css: Unexpected PostCSS node type: \\\\\\\"\\\\\\\".\\\"\", Go cohere \"\\\"a {\\\\n  b: c;\\\\n}\\\\n\\\"\"; log M4-fast.log",
    "subsumption_mutants": 2
  },
  {
    "test": "TestProduct_CSSPrinterOracle family",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/printer_shards_test.go:134",
    "seconds": 1.391,
    "oracle": "Successful Go oracle compilation/path preparation, with no executable-existence or semantic-answer assertion. S2 publishes paths without building any Go oracle binary.",
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
      "TestProduct_CSSPrinterOracle family"
    ],
    "evidence": "ADAMIC_MUTANT=S2; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^TestProduct_CSSPrinter(ParserOracle|Oracle)$; TestProduct_CSSPrinterParserOracle: PASS; TestProduct_CSSPrinterOracle: PASS; construction-artifact-proofs.json",
    "construction_checks": [
      "S2"
    ],
    "members": [
      "TestProduct_CSSPrinterParserOracle",
      "TestProduct_CSSPrinterOracle"
    ]
  },
  {
    "test": "TestProduct_CSSPrinterNative family",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/printer_shards_test.go:136",
    "seconds": 1.684,
    "oracle": "Successful native product preparation, with no execution or artifact-existence assertion. S1 publishes paths while omitting native.Build; no native files are produced.",
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
      "TestProduct_CSSPrinterNative family"
    ],
    "evidence": "ADAMIC_MUTANT=S1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^TestProduct_CSSPrinter(SanitizedAndLowered|SemicolonMutant|IndentMutant|WidthMutant)$; TestProduct_CSSPrinterIndentMutant: PASS; TestProduct_CSSPrinterSanitizedAndLowered: PASS; TestProduct_CSSPrinterWidthMutant: PASS; TestProduct_CSSPrinterSemicolonMutant: PASS; construction-artifact-proofs.json",
    "construction_checks": [
      "S1"
    ],
    "members": [
      "TestProduct_CSSPrinterSanitizedAndLowered",
      "TestProduct_CSSPrinterSemicolonMutant",
      "TestProduct_CSSPrinterIndentMutant",
      "TestProduct_CSSPrinterWidthMutant",
      "TestProduct_CSSPrinterDarwinLeaks"
    ],
    "skipped_members": [
      "TestProduct_CSSPrinterDarwinLeaks"
    ]
  },
  {
    "test": "TestCSSPrinterShardingCatchesDisagreement",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/printer_shards_test.go:356",
    "seconds": 4.035,
    "oracle": "Self-written disagreement/hash-owner expectation and duplicate/missing-case rejection. W1 disables stdout comparison; S4 disables union validation, and each independently makes the row fail.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1 printer_shards_test.go:380: planted mode/case 0 caught by [], want exactly its hash owner",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCSSPrinterShardingCatchesDisagreement"
    ],
    "evidence": "ADAMIC_MUTANT=W1; selector /tmp/u079/selector=W1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^TestCSSPrinterShardingCatchesDisagreement$; printer_shards_test.go:380: planted mode/case 0 caught by [], want exactly its hash owner; log W1-construction.log",
    "witness_kills": [
      "W1"
    ],
    "construction_kills": [
      "S4"
    ]
  },
  {
    "test": "TestCSSProfileArtifacts",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/profile_test.go:18",
    "seconds": 47.061,
    "oracle": "Go cohere and the fork select shared sample inputs, but artifact assertions cover only calls/build success. S3 writes expected.txt under the wrong name and the test passes.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
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
      "TestCSSProfileArtifacts"
    ],
    "evidence": "ADAMIC_MUTANT=S3; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^TestCSSProfileArtifacts$; TestCSSProfileArtifacts: PASS; construction-artifact-proofs.json",
    "construction_checks": [
      "S3"
    ]
  },
  {
    "test": "TestCSSProfileSnapshotsAgree",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/profile_test.go:134",
    "seconds": 37.794,
    "oracle": "Go cohere supplies complete answers in default and narrow modes. Fresh source snapshot on Node and printer/profiled/counted native products must match stdout exactly.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4 profile_test.go:154: /tmp/u079/switched-profile printer 70 adamic: panic: missing CSS byte offset",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCSSPrinterThroughput"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 77.027,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestClosedPrinterRegexGap",
      "TestCSSPrinterThroughput",
      "TestCSSPrinterBoundaryProofs",
      "TestOptionalBooleanPrinterMatchesGo",
      "TestCSSProfileSnapshotsAgree",
      "TestSharedSliceAppendAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M4; selector /tmp/u079/selector=M4; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^TestCSSProfileSnapshotsAgree$; profile_test.go:154: /tmp/u079/switched-profile printer 70 adamic: panic: missing CSS byte offset; log M4-snapshots.log",
    "subsumption_mutants": 4
  },
  {
    "test": "TestSharedSliceAppendAgreesWithNode",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/shared_slice_test.go:9",
    "seconds": 0.355,
    "oracle": "Node, native and JavaScript backend each must print the self-written length literal 1152. It checks a length, not the appended bytes; M4 fails with ASan heap-buffer-overflow.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4 shared_slice_test.go:24: native: exit 1, stdout \"\", stderr \"=================================================================\\n==95911==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7c32e81e0061 at pc 0x5623454dc1e0 bp 0x7ffcb40ca530 sp 0x7ffcb40ca528\\nREAD of size 1 at 0x7c32e81e0061 thread T0\\n    #0 0x5623454dc1df in adamic_string_locate /home/agent/.cache/adamic/runtime/.build-1579466526/string_index.c:198:38\\n    #1 0x5623454d4d7c in adamic_string_slice /home/agent/.cache/adamic/runtime/.build-1579466526/string_slice_impl.h:34:40\\n    #2 0x562345483b3b in adamic_function_0_demonstrate /tmp/adamic-gate/adamic-build-3247023205",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestOptionalBooleanPrinterMatchesGo"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": 26.151,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestClosedPrinterRegexGap",
      "TestCSSPrinterThroughput",
      "TestCSSPrinterBoundaryProofs",
      "TestOptionalBooleanPrinterMatchesGo",
      "TestCSSProfileSnapshotsAgree",
      "TestSharedSliceAppendAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M4; selector /tmp/u079/selector=M4; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^Test(ClosedPrinterRegexGap|CSSPrinterBoundaryProofs|OptionalBooleanPrinterMatchesGo|SharedSliceAppendAgreesWithNode)$; shared_slice_test.go:24: native: exit 1, stdout \"\", stderr \"=================================================================\\n==95911==ERROR: AddressSanitizer: heap-buffer-overflow on address 0x7c32e81e0061 at pc 0x5623454dc1e0 bp 0x7ffcb40ca530 sp 0x7ffcb40ca528\\nREAD of size 1 at 0x7c32e81e0061 thread T0\\n    #0 0x5623454dc1df in adamic_string_locate /home/agent/.cache/adamic/runtime/.build-1579466526/string_index.c:198:38\\n    #1 0x5623454d4d7c in adamic_string_slice /home/agent/.cache/adamic/runtime/.build-1579466526/string_slice_impl.h:34:40\\n    #2 0x562345483b3b in adamic_function_0_demonstrate /tmp/adamic-gate/adamic-build-3247023205; log M4-fast.log",
    "subsumption_mutants": 1
  }
]
```

Production mutants, origin/main cf79ecec3723604428ab91ebcb283400d05a1548

| ID | File:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/css/print.ts:12 | default tabWidth: 2 → 3 | TestCSSPrinterThroughput, TestCSSPrinterBoundaryProofs, TestOptionalBooleanPrinterMatchesGo, TestCSSProfileSnapshotsAgree |
| M2 | stage1/cohere/css/print_doc.ts:103 | fits loop: remaining >= 0 → remaining > 0 | TestCSSPrinterThroughput, TestCSSProfileSnapshotsAgree |
| M3 | stage1/cohere/mediaquery/index.ts:33 | parse returns Error with audit early refusal before parsing | TestClosedPrinterRegexGap, TestCSSPrinterThroughput, TestCSSProfileSnapshotsAgree |
| M4 | internal/native/runtime/string_append.c:86 | drop result->length = written; | TestCSSPrinterThroughput, TestOptionalBooleanPrinterMatchesGo, TestCSSProfileSnapshotsAgree, TestSharedSliceAppendAgreesWithNode |

Probes and construction checks are excluded from production kills and uniqueness.

| ID | File:line | Operation | Observed result |
|---|---|---|---|
| P1 | stage1/cohere/css/print.ts:266 | probe | TestOptionalBooleanPrinterMatchesGo fail, TestCSSPrinterBoundaryProofs fail, TestCSSPrinterThroughput fail, TestCSSProfileSnapshotsAgree fail |
| P2 | stage1/cohere/mediaquery/index.ts:33 | probe | TestClosedPrinterRegexGap pass |
| P3 | internal/native/runtime/string_append.c:34 | probe | TestSharedSliceAppendAgreesWithNode fail |
| P4 | internal/native/runtime/regexp.c:621 | probe | TestClosedPrinterRegexGap fail |
| W1 | stage1/cohere/css/printer_shards_test.go:280 | construction | TestCSSPrinterShardingCatchesDisagreement fail |
| S1 | stage1/cohere/css/printer_shards_test.go:124 | construction | TestProduct_CSSPrinterIndentMutant pass, TestProduct_CSSPrinterSanitizedAndLowered pass, TestProduct_CSSPrinterWidthMutant pass, TestProduct_CSSPrinterSemicolonMutant pass |
| S4 | stage1/cohere/css/printer_shards_test.go:309 | construction | TestCSSPrinterShardingCatchesDisagreement fail |
| S2 | stage1/cohere/css/printer_shards_test.go:70 | construction | TestProduct_CSSPrinterParserOracle pass, TestProduct_CSSPrinterOracle pass |
| S3 | stage1/cohere/css/profile_test.go:62 | construction | TestCSSProfileArtifacts pass |

Survivors: none among M1 to M4 in the six-row bounded matrix. Kills outside that matrix are unknown. No equivalence claim is needed. S1, S2 and S3 are construction survivors, with witnesses in construction-artifact-proofs.json: zero native files, zero oracle binaries, and missing expected.txt with missing-expected.txt present.

The brief, limits and time costs

The reference commit 8de93800f4 was stale. Fresh origin/main was cf79ecec3723604428ab91ebcb283400d05a1548. Every requested name still existed in its named file. The test list, scope locations and standalone apply checks are saved.

The full package reached its 90-second budget before completing the unit. A batch containing the opt-in throughput row also cooked, and a subsequent batch containing cold products cooked. All assigned runnable rows were then proved green alone. The sixteen original printer shards also cooked at 90 seconds. Their partial outcomes are preserved, but they do not establish family success or package uniqueness. The production matrix was narrowed to the six assigned semantic rows. Three construction rows and the witness have their own construction/check experiments. No other Adamic packages were run to seek uniqueness.

The throughput opt-in is unusually expensive for a measurement row: three internal rounds each run native, Adamic on Node, fork Prettier and npm Prettier, then Go throughput. ADAMIC_CSS_PRINTER_BENCH_ONCE=1 was enabled; median binary cost is 77.027 seconds, without the default ten repetitions. It is subsumed on this four-mutant set, rather than slow-worthy, because none of its kills is unique. The aggregate checksum catches M2's 605005 output units versus the expected 604909, but equal-length wrong text could pass.

Profile artifacts must be generated before snapshots are read. Enabling both in one concurrent batch risks reading an incomplete directory, so the clean artifact row ran first. Switched artifacts were freshly rebuilt once and reused only with the runtime selector, including native counter/profiler products. This avoids stale port products. Direct-build rows still call native.Build on each test invocation; their large main C translation unit is rebuilt despite the source selector. Those costs remain in the logs.

Input-only product declarations are families: the two Go oracle builders share cssOracleProduct, and five native builders share cssPrinterProduct. The Darwin wrapper's availability guard does not add a distinct assertion, so it belongs to the native family. The complete native family was retimed three times with the Darwin member included and skipped. Linux cannot install the macOS leaks tool. Its product recipe was not exercised. The three named mutant-product declarations merely build mutated products; they do not themselves witness a comparison failure.

The brief assumes one code entry per row. The closed-regex row reaches both mediaquery.parse and adamic_regex_test. P2's empty Ok tree is accepted by all three backends, while P4's false RegExp answer fails. vacuous is true for the accepted parse entry, with separate vacuous_entries preserving the nonvacuous RegExp entry. This aggregation is an explicit interpretation of an unspecified multi-entry case.

A timeout interrupted the first M1 fast batch. It was not counted as a kill. Every assigned fast row was rerun alone for M1, and those completed outcomes replace unknown entries. No other production or probe batch panicked. The broad caller baseline panic has no mutation verdict.

Cold detached validation worktrees unexpectedly rebuilt Go dependencies and hit the 120-second compilation backstop before a test binary produced output. Those failed validation attempts are retained. I stopped their children and validated the same standalone TypeScript patches sequentially in the warm starting workspace. All five native validations passed; every file was restored immediately. C patches pass the exact native runtime clang flags, and the five standalone witness/construction patches pass Go vet overlays. All thirteen standalone diffs apply to the starting origin/main. The switch and drivers are evidence only, and production sources are restored.

Setup used the warm Go 1.27.1 toolchain, so setup.sh was not run. nproc is 5. npm ci in stage3/api reported 471 ms; installing the required Prettier 3.9.6 oracle reported 822 ms. No other package-loaded node_modules directories were found. Assigned opt-ins were enabled: npm printer library, benchmark with once mode, profile artifact directory, and profile snapshots. The only assigned skipped function was TestProduct_CSSPrinterDarwinLeaks. Outside the assigned unit, the broad baseline also recorded TestCSSThroughput skipping its separate benchmark opt-in.

The 384-function port inventory comes from clean V8 traces over the entire printer corpus and both regression fixtures, before mutation. It includes module initializers and generated member initializers. Native runtime entry inventory is static, with untraced downstream C helpers marked conservative; native function coverage was not measured. This is a reachability limit, rather than a claim about native path coverage.

Subsumption rests on only one mutant for the closed-regex, boundary and shared-slice rows, two for optional booleans, and four for throughput/snapshots. Throughput and snapshots mutually subsume one another. These are bounded retention hints for the defender wave, not deletion recommendations. Three construction checks accept missing or misnamed artifacts; their construction survivors are separate from production survivors.

Timing and rebuild details are in timings.json, complete-native-family-timings.json, workspace-port-validations.json, builds.log and matrix-runs.json. Native product timing includes checking, lowering, C emission and clang, not clang alone. The switched artifact product built in 52.256 binary seconds; the original four-native-product cold family run took 32.312 seconds. Actual standalone native rebuild timings are reported below. Total elapsed exceeded the approximate 30-minute budget because of mandatory isolated timings, repeated cooked baselines, direct native rebuilds and the cold validation attempt. Package and repository uniqueness, macOS behavior, unsupported opt-in corpora outside this unit, and all omitted caller outcomes remain unmeasured.

| Standalone | Binary seconds | Command wall seconds |
|---|---:|---:|
| M1 | 27.093 | 29.149 |
| M2 | 25.351 | 27.396 |
| M3 | 25.872 | 28.022 |
| P1 | 19.641 | 21.827 |
| P2 | 26.516 | 28.883 |

Original isolated timing runs consumed 629.805 binary seconds in total. Matrix commands consumed 898.600 wall seconds. These totals are phase costs and must not be summed with overlapping validation attempts as elapsed time.

Full JSON, raw logs, standalone diffs, family membership, function inventory, selector source, compile validations and reproducible drivers are included in this directory. No PR was opened, and no production edits are committed.

Elapsed from fresh audit-branch creation to final review: 57.3 minutes. Warm setup: 0 s. Isolated timing binaries: 629.805 s. Matrix command walls: 898.600 s. Successful standalone port validation command walls: 135.277 s. costs.json records exact timestamps. This exceeded the approximate 30-minute budget.
