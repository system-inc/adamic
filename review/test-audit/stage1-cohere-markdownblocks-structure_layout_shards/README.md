Audited 124 named tests as 12 rows at d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb.
Three isolated clean runs passed for every row; nproc=5; no skips.
Verdicts: 2 bounded sacred, 3 subsumed, 4 setup-check, 2 witness, 1 helper.
Four production mutants were caught; all five functional rows rejected their own empty entry.
Sources restored; evidence only on the audit branch.

[
  {
    "test": "TestMarkdownStructureLayout family",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/structure_layout_shards_test.go"
    ],
    "seconds": 18.285,
    "oracle": "Go cohere and pinned Prettier 3.9.6 fork executed through Node; full output bytes, with self-written shard union and built-in mutant expectations.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: source Node lists output byte 2434 in generated/list-layout/edge/- a",
    "verdict": "subsumed",
    "subsumed_by": "TestMarkdownTableLayout family",
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "probe_kills": [
      "P1",
      "P2",
      "P3",
      "P4"
    ],
    "subsumer_seconds": 13.42,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownStructureLayout family",
      "TestMarkdownTableLayout family",
      "TestMarkdownTextSplitting family",
      "TestMarkdownWhitespaceLayout family",
      "TestMarkdownUnicodeWidths"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '(?:^TestMarkdownStructureLayout(Union|_[0-9]+)$)|(?:^TestMarkdownTableLayout(Union|_[0-9]+)$)|(?:^TestMarkdownWhitespaceLayout(Union|_[0-9]+)$)' > M2.log 2>&1; source Node lists output byte 2434 in generated/list-layout/edge/- a",
    "members": [
      "TestMarkdownStructureLayoutUnion",
      "TestMarkdownStructureLayout_000",
      "TestMarkdownStructureLayout_001",
      "TestMarkdownStructureLayout_002",
      "TestMarkdownStructureLayout_003",
      "TestMarkdownStructureLayout_004",
      "TestMarkdownStructureLayout_005",
      "TestMarkdownStructureLayout_006",
      "TestMarkdownStructureLayout_007",
      "TestMarkdownStructureLayout_008",
      "TestMarkdownStructureLayout_009",
      "TestMarkdownStructureLayout_010",
      "TestMarkdownStructureLayout_011",
      "TestMarkdownStructureLayout_012",
      "TestMarkdownStructureLayout_013",
      "TestMarkdownStructureLayout_014",
      "TestMarkdownStructureLayout_015"
    ],
    "entry_probes": [
      "P3"
    ],
    "subsumption_mutants": 2
  },
  {
    "test": "TestMarkdownTableLayout family",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/table_layout_shards_test.go"
    ],
    "seconds": 13.42,
    "oracle": "Go cohere and pinned Prettier 3.9.6 fork executed through Node; full output bytes, with self-written shard union and built-in mutant expectations.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: source Node lists output byte 4394 in generated/quote-layout/> a",
    "verdict": "subsumed",
    "subsumed_by": "TestMarkdownStructureLayout family",
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "probe_kills": [
      "P1",
      "P2",
      "P3",
      "P4"
    ],
    "subsumer_seconds": 18.285,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownStructureLayout family",
      "TestMarkdownTableLayout family",
      "TestMarkdownTextSplitting family",
      "TestMarkdownWhitespaceLayout family",
      "TestMarkdownUnicodeWidths"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '(?:^TestMarkdownStructureLayout(Union|_[0-9]+)$)|(?:^TestMarkdownTableLayout(Union|_[0-9]+)$)|(?:^TestMarkdownWhitespaceLayout(Union|_[0-9]+)$)' > M2.log 2>&1; source Node lists output byte 4394 in generated/quote-layout/> a",
    "members": [
      "TestMarkdownTableLayoutUnion",
      "TestMarkdownTableLayout_000",
      "TestMarkdownTableLayout_001",
      "TestMarkdownTableLayout_002",
      "TestMarkdownTableLayout_003",
      "TestMarkdownTableLayout_004",
      "TestMarkdownTableLayout_005",
      "TestMarkdownTableLayout_006",
      "TestMarkdownTableLayout_007"
    ],
    "entry_probes": [
      "P3"
    ],
    "subsumption_mutants": 2
  },
  {
    "test": "TestMarkdownTextSplitting family",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/text_independent_shards_test.go",
      "stage1/cohere/markdownblocks/text_independent_test.go"
    ],
    "seconds": 40.631,
    "oracle": "Go cohere and pinned Prettier 3.9.6 fork executed through Node; full output bytes, with self-written shard union and built-in mutant expectations.",
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
    "last_proven_fail": "M4: native: mismatch in \"generated/list/1./nnnnn\"",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": [
      "M4"
    ],
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownStructureLayout family",
      "TestMarkdownTableLayout family",
      "TestMarkdownTextSplitting family",
      "TestMarkdownWhitespaceLayout family",
      "TestMarkdownUnicodeWidths"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '(?:^TestMarkdownStructureLayout(Union|_[0-9]+)$)|(?:^TestMarkdownTableLayout(Union|_[0-9]+)$)|(?:^TestMarkdownWhitespaceLayout(Union|_[0-9]+)$)|(?:^TestMarkdownTextSplitting(Union|_[0-9]+)$)' > M4.log 2>&1; native: mismatch in \"generated/list/1./nnnnn\"",
    "members": [
      "TestMarkdownTextSplitting_000",
      "TestMarkdownTextSplitting_001",
      "TestMarkdownTextSplitting_002",
      "TestMarkdownTextSplitting_003",
      "TestMarkdownTextSplitting_004",
      "TestMarkdownTextSplitting_005",
      "TestMarkdownTextSplitting_006",
      "TestMarkdownTextSplitting_007",
      "TestMarkdownTextSplitting_008",
      "TestMarkdownTextSplitting_009",
      "TestMarkdownTextSplitting_010",
      "TestMarkdownTextSplitting_011",
      "TestMarkdownTextSplitting_012",
      "TestMarkdownTextSplitting_013",
      "TestMarkdownTextSplitting_014",
      "TestMarkdownTextSplitting_015",
      "TestMarkdownTextSplitting_016",
      "TestMarkdownTextSplitting_017",
      "TestMarkdownTextSplitting_018",
      "TestMarkdownTextSplitting_019",
      "TestMarkdownTextSplitting_020",
      "TestMarkdownTextSplitting_021",
      "TestMarkdownTextSplitting_022",
      "TestMarkdownTextSplitting_023",
      "TestMarkdownTextSplitting_024",
      "TestMarkdownTextSplitting_025",
      "TestMarkdownTextSplitting_026",
      "TestMarkdownTextSplitting_027",
      "TestMarkdownTextSplitting_028",
      "TestMarkdownTextSplitting_029",
      "TestMarkdownTextSplitting_030",
      "TestMarkdownTextSplitting_031",
      "TestMarkdownTextSplitting_032",
      "TestMarkdownTextSplitting_033",
      "TestMarkdownTextSplitting_034",
      "TestMarkdownTextSplitting_035",
      "TestMarkdownTextSplitting_036",
      "TestMarkdownTextSplitting_037",
      "TestMarkdownTextSplitting_038",
      "TestMarkdownTextSplitting_039",
      "TestMarkdownTextSplitting_040",
      "TestMarkdownTextSplitting_041",
      "TestMarkdownTextSplitting_042",
      "TestMarkdownTextSplitting_043",
      "TestMarkdownTextSplitting_044",
      "TestMarkdownTextSplitting_045",
      "TestMarkdownTextSplitting_046",
      "TestMarkdownTextSplitting_047",
      "TestMarkdownTextSplitting_048",
      "TestMarkdownTextSplitting_049",
      "TestMarkdownTextSplitting_050",
      "TestMarkdownTextSplitting_051",
      "TestMarkdownTextSplitting_052",
      "TestMarkdownTextSplitting_053",
      "TestMarkdownTextSplitting_054",
      "TestMarkdownTextSplitting_055",
      "TestMarkdownTextSplitting_056",
      "TestMarkdownTextSplitting_057",
      "TestMarkdownTextSplitting_058",
      "TestMarkdownTextSplitting_059",
      "TestMarkdownTextSplitting_060",
      "TestMarkdownTextSplitting_061",
      "TestMarkdownTextSplitting_062",
      "TestMarkdownTextSplitting_063",
      "TestMarkdownTextSplittingUnion"
    ],
    "entry_probes": [
      "P2"
    ]
  },
  {
    "test": "TestMarkdownWhitespaceLayout family",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/whitespace_layout_split_test.go"
    ],
    "seconds": 14.752,
    "oracle": "Go cohere and pinned Prettier 3.9.6 fork executed through Node; full output bytes, with self-written shard union and built-in mutant expectations.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M3: whitespace native first byte difference at 6 (lengths 895/907)",
    "verdict": "sacred",
    "subsumed_by": null,
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "probe_kills": [
      "P1",
      "P2",
      "P3",
      "P4"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownStructureLayout family",
      "TestMarkdownTableLayout family",
      "TestMarkdownTextSplitting family",
      "TestMarkdownWhitespaceLayout family",
      "TestMarkdownUnicodeWidths"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '(?:^TestMarkdownStructureLayout(Union|_[0-9]+)$)|(?:^TestMarkdownTableLayout(Union|_[0-9]+)$)|(?:^TestMarkdownWhitespaceLayout(Union|_[0-9]+)$)' > M3.log 2>&1; whitespace native first byte difference at 6 (lengths 895/907)",
    "members": [
      "TestMarkdownWhitespaceLayoutUnion",
      "TestMarkdownWhitespaceLayout_000",
      "TestMarkdownWhitespaceLayout_001",
      "TestMarkdownWhitespaceLayout_002",
      "TestMarkdownWhitespaceLayout_003",
      "TestMarkdownWhitespaceLayout_004",
      "TestMarkdownWhitespaceLayout_005",
      "TestMarkdownWhitespaceLayout_006",
      "TestMarkdownWhitespaceLayout_007",
      "TestMarkdownWhitespaceLayout_008",
      "TestMarkdownWhitespaceLayout_009",
      "TestMarkdownWhitespaceLayout_010",
      "TestMarkdownWhitespaceLayout_011",
      "TestMarkdownWhitespaceLayout_012",
      "TestMarkdownWhitespaceLayout_013",
      "TestMarkdownWhitespaceLayout_014",
      "TestMarkdownWhitespaceLayout_015",
      "TestMarkdownWhitespaceLayout_016",
      "TestMarkdownWhitespaceLayout_017",
      "TestMarkdownWhitespaceLayout_018",
      "TestMarkdownWhitespaceLayout_019",
      "TestMarkdownWhitespaceLayout_020",
      "TestMarkdownWhitespaceLayout_021",
      "TestMarkdownWhitespaceLayout_022",
      "TestMarkdownWhitespaceLayout_023"
    ],
    "entry_probes": [
      "P3",
      "P4"
    ]
  },
  {
    "test": "TestMarkdownUnicodeWidths",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/width_test.go"
    ],
    "seconds": 88.225,
    "oracle": "Live Go cohere widths compared byte for byte with native/source/backend and pinned Prettier plus emoji-regex 10.6.0, get-east-asian-width 1.6.0, narrow-emojis 0.0.3 through Node. Repository width-table regeneration is also checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: native mismatch \"stage1/cohere/markdownblocks/GAPS.md:1\" got \"44\" want \"43\"",
    "verdict": "subsumed",
    "subsumed_by": "TestMarkdownTableLayout family",
    "mutants_in_matrix": [
      "M1"
    ],
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 13.42,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownStructureLayout family",
      "TestMarkdownTableLayout family",
      "TestMarkdownTextSplitting family",
      "TestMarkdownWhitespaceLayout family",
      "TestMarkdownUnicodeWidths"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '(?:^TestMarkdownStructureLayout(Union|_[0-9]+)$)|(?:^TestMarkdownTableLayout(Union|_[0-9]+)$)|(?:^TestMarkdownWhitespaceLayout(Union|_[0-9]+)$)|(?:^TestMarkdownUnicodeWidths$)' > M1.log 2>&1; native mismatch \"stage1/cohere/markdownblocks/GAPS.md:1\" got \"44\" want \"43\"",
    "members": [
      "TestMarkdownUnicodeWidths"
    ],
    "entry_probes": [
      "P1"
    ],
    "subsumption_mutants": 1
  },
  {
    "test": "TestNativeBuildModesAreDistinct",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/support_test.go"
    ],
    "seconds": 0.229,
    "oracle": "Self-written sanitized/release output literals and artifact-key separation; clang produces the two executables.",
    "oracle_kind": "self",
    "kills": [
      "S1"
    ],
    "unique_kills": [],
    "last_proven_fail": "S1: clang build mode first byte difference at 0 (lengths 10/8)",
    "verdict": "setup-check",
    "subsumed_by": null,
    "mutants_in_matrix": [
      "S1"
    ],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestNativeBuildModesAreDistinct"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestNativeBuildModesAreDistinct$' > S1.log 2>&1; clang build mode first byte difference at 0 (lengths 10/8)",
    "members": [
      "TestNativeBuildModesAreDistinct"
    ],
    "entry_probes": []
  },
  {
    "test": "TestMarkdownTextShardDisagreement",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/text_independent_shards_test.go"
    ],
    "seconds": 0.009,
    "oracle": "Self-written planted disagreement must be detected exactly in its owning shard.",
    "oracle_kind": "self",
    "kills": [
      "W1"
    ],
    "unique_kills": [],
    "last_proven_fail": "W1: planted disagreement caught by [], want only shard-025",
    "verdict": "witness",
    "subsumed_by": null,
    "mutants_in_matrix": [
      "W1"
    ],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownTextShardDisagreement"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownTextShardDisagreement$' > W1.log 2>&1; planted disagreement caught by [], want only shard-025",
    "members": [
      "TestMarkdownTextShardDisagreement"
    ],
    "entry_probes": []
  },
  {
    "test": "TestMarkdownWhitespaceLayoutMutants",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/whitespace_layout_split_test.go"
    ],
    "seconds": 6.057,
    "oracle": "Live Go cohere fixture bytes; self-written expectation that altered whitespace source must differ. Weakened comparator is the deciding witness run.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "W2"
    ],
    "unique_kills": [],
    "last_proven_fail": "W2: whitespace mutant preserved newline survived shard",
    "verdict": "witness",
    "subsumed_by": null,
    "mutants_in_matrix": [
      "W2"
    ],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownWhitespaceLayoutMutants"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownWhitespaceLayoutMutants$' > W2.log 2>&1; whitespace mutant preserved newline survived shard",
    "members": [
      "TestMarkdownWhitespaceLayoutMutants"
    ],
    "entry_probes": []
  },
  {
    "test": "TestMarkdownTableLayout_Setup",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/table_layout_shards_test.go"
    ],
    "seconds": 3.86,
    "oracle": "Self: tableLayoutReady must populate inputs; its nil guard remains intact.",
    "oracle_kind": "self",
    "kills": [
      "S2"
    ],
    "unique_kills": [],
    "last_proven_fail": "S2: table layout setup failed",
    "verdict": "setup-check",
    "subsumed_by": null,
    "mutants_in_matrix": [
      "S2"
    ],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownTableLayout_Setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownTableLayout_Setup$' > S2.log 2>&1; table layout setup failed",
    "members": [
      "TestMarkdownTableLayout_Setup"
    ],
    "entry_probes": []
  },
  {
    "test": "TestMarkdownTextSplitting_Setup",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/text_independent_test.go"
    ],
    "seconds": 5.382,
    "oracle": "Self: textReadySetup must set ready; its completion guard remains intact.",
    "oracle_kind": "self",
    "kills": [
      "S3"
    ],
    "unique_kills": [],
    "last_proven_fail": "S3: shared text setup did not complete",
    "verdict": "setup-check",
    "subsumed_by": null,
    "mutants_in_matrix": [
      "S3"
    ],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownTextSplitting_Setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownTextSplitting_Setup$' > S3.log 2>&1; shared text setup did not complete",
    "members": [
      "TestMarkdownTextSplitting_Setup"
    ],
    "entry_probes": []
  },
  {
    "test": "TestMarkdownWhitespaceLayout_Setup",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/whitespace_layout_split_test.go"
    ],
    "seconds": 1.204,
    "oracle": "Self: whitespaceLayoutReadyProducts must populate goBinary; its guard remains intact.",
    "oracle_kind": "self",
    "kills": [
      "S4"
    ],
    "unique_kills": [],
    "last_proven_fail": "S4: whitespace layout setup failed",
    "verdict": "setup-check",
    "subsumed_by": null,
    "mutants_in_matrix": [
      "S4"
    ],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownWhitespaceLayout_Setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownWhitespaceLayout_Setup$' > S4.log 2>&1; whitespace layout setup failed",
    "members": [
      "TestMarkdownWhitespaceLayout_Setup"
    ],
    "entry_probes": []
  },
  {
    "test": "TestMarkdownWhitespaceLayoutBuildWorker",
    "package": "stage1/cohere/markdownblocks",
    "file": [
      "stage1/cohere/markdownblocks/whitespace_layout_split_test.go"
    ],
    "seconds": 0.007,
    "oracle": "No standalone expected answer: returns without ADAMIC_WHITESPACE_BUILD; used by whitespace family, its setup and mutant witness.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "subsumed_by": null,
    "mutants_in_matrix": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownWhitespaceLayoutBuildWorker"
    ],
    "evidence": "go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^TestMarkdownWhitespaceLayoutBuildWorker$' > worker-clean-1.log 2>&1; PASS, helper inactive",
    "members": [
      "TestMarkdownWhitespaceLayoutBuildWorker"
    ],
    "entry_probes": [],
    "parents": [
      "TestMarkdownWhitespaceLayout family",
      "TestMarkdownWhitespaceLayout_Setup",
      "TestMarkdownWhitespaceLayoutMutants"
    ]
  }
]

Mutants: all locations below refer to the starting origin/main commit.

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/markdownblocks/width.ts:23 | if(ascii) return text.length; -> if(ascii) return text.length + 1; | TestMarkdownStructureLayout family, TestMarkdownTableLayout family, TestMarkdownWhitespaceLayout family, TestMarkdownUnicodeWidths |
| M2 | stage1/cohere/markdownblocks/structure.ts:51 | if(!setext) return arena.concat -> if(setext) return arena.concat | TestMarkdownStructureLayout family, TestMarkdownTableLayout family, TestMarkdownWhitespaceLayout family |
| M3 | stage1/cohere/markdownblocks/whitespace.ts:77 | if(proseWrap !== 'always') return false; -> if(proseWrap !== 'never') return false; | TestMarkdownWhitespaceLayout family |
| M4 | stage1/cohere/markdownblocks/splitText.ts:62 | if(unit === 10) newline = true; -> if(unit === 9) newline = true; | TestMarkdownTextSplitting family |
| S1 | stage1/cohere/markdownblocks/support_test.go:189 | sanitize: options.Sanitize -> sanitize: false | TestNativeBuildModesAreDistinct |
| S2 | stage1/cohere/markdownblocks/table_layout_shards_test.go:43 | drop complete tableLayoutOnce.Do(...) statement; retain nil guard | TestMarkdownTableLayout_Setup |
| S3 | stage1/cohere/markdownblocks/text_independent_shards_test.go:276 | drop complete textSetupOnce.Do(...) statement; retain ready guard | TestMarkdownTextSplitting_Setup |
| S4 | stage1/cohere/markdownblocks/whitespace_layout_split_test.go:766 | drop complete whitespaceLayoutReadyOnce.Do(...) statement; retain goBinary guard | TestMarkdownWhitespaceLayout_Setup |
| W1 | stage1/cohere/markdownblocks/text_independent_shards_test.go:409 | comparison function body -> return nil | TestMarkdownTextShardDisagreement |
| W2 | stage1/cohere/markdownblocks/whitespace_layout_split_test.go:204 | return bytes.Equal(actual, expected) -> return true | TestMarkdownWhitespaceLayoutMutants |

Survivors: none among M1-M4. Construction and witness edits all failed their designated rows. Probes are separate from production kills.

Brief ambiguities and costs:

* The file list is stale: text_test.go and text_shards_test.go moved to the independent test files. All 124 names remain.
* The heading says 12 rows but supplies 124 functions. Grouping each numbered family with its union and treating setup, witness and helper entries separately gives 12 rows. Setup functions differ in purpose from formatting leaves.
* Warm tools did not mean warm content-addressed products. The same list driver is lowered under separate structure, table and whitespace recipes, so preparation repeated work.
* Whole-package and whole-unit clean runs each timed out at 90 seconds during preparation. They were narrowed to isolated complete families. No assertion failure or skip was observed in the narrowed clean baselines.
* The Unicode-width row is close to the budget: 89.001, 88.225 and 87.500 seconds. Its small standalone native program also requires expensive native compilation.
* A standalone M2 build was stopped at 90.133 seconds, then narrowed to heading construction and successfully built in 38.420 seconds using supported split native builds. Other successful native build times are in standalone-builds.json.
* The executor transport disconnected during preparation. Retrying cost roughly 90 seconds; saved test logs and all 36 timing results survived.
* Four port mutants follow the stage1 rebuild limit. Subsumption rests on two catches for structure/table and one for widths, so it is a small-set hint, not a deletion recommendation.
* No runtime selector existed in the original port, so scratch sources read /tmp/adamic-u130-mutant. Each product was compiled once with all selectors, and a positive control passed before the matrix. Standalone diffs contain no selector.
* Allowed construction and witness edits were tested individually after the port matrix. Their Go recompilations are included in matrix wall costs. They are not production uniqueness evidence.
* The static inventory lists potentially reachable port functions, not dynamic execution coverage. This audit cannot certify that every listed branch ran.
* The suggested 30-minute stage1 budget was exceeded. Three width runs alone took 264.726 test-binary seconds; required product builds, entry probes and exact standalone native checks account for further cost.

Uncovered: other package rows, other packages, repo-wide uniqueness, opt-in full repository census and optional performance repeats. All supplied rows ran and none skipped. Family unions are included. The helper is classified by its environment guard and its known parent calls, not by its inactive standalone PASS.

Costs (seconds):

{
  "setup_seconds": 0,
  "npm_seconds": 0.9963279419971514,
  "nproc": 5,
  "whole_package_baseline_wall": 92.5093177460003,
  "unit_baseline_wall": 92.3016283399993,
  "36_isolated_runs_wall": 810.9240082819997,
  "switched_product_build_wall": 213.97728041099617,
  "instrumented_control_wall": 11.787398624997877,
  "mutant_probe_and_check_matrix_wall": 505.4148313339938,
  "standalone_native_build_wall": 373.8933040650045,
  "standalone_compile_driver_wall": 6.675196895001136,
  "note": "Standalone compilation overlapped the matrix; sums are not elapsed session time."
}

Compilation validation: all ten mutant diffs apply to the starting source tree. S1-S4 and W1-W2 pass go vet. M1-M4 and P1-P4 each successfully built natively; M2 required the recorded narrowed retry. Native compile adapters and the split-build wrapper are saved alongside logs.

Failing assertion text in results.json is quoted without shifted scratch file-line prefixes. Mutation locations in mutants.json and the table are origin lines. Raw output is preserved in every log.

Elapsed before commit: 2799.3 seconds (46.7 minutes), from branch creation. This exceeded the suggested stage1 budget.
