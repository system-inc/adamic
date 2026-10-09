u099 starts at a467d1a1571c43e0f01fc4efcb43890280e4a1ac; 20 of 21 requested names exist.  
TestPrinterAsGoCohere_Setup vanished; current Union member joins the AsGo family, yielding ten current rows.  
Full package cooked at 90s; bounded slice green in 24.677s; all four production mutants caught.  
Verdicts: two sacred, two witness, four setup-check, one subsumed, one cannot-judge; upstream Setup vacuous=true.  
nproc=5; all three timings per current row completed; production source restored; evidence is push-ready.

```json
[
  {
    "test": "TestPrinterPreflightPlantedDisagreement",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/preflight_units_test.go",
    "seconds": 4.332,
    "oracle": "npm Prettier and Go cohere, with self-written owner/count expectations. This is a witness of printerPreflightDisagreement. Shared error-prefixed refusals are accepted without checking diagnostic equality.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: preflight_units_test.go:156: planted disagreement caught by 0 shards, want exactly one",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier ADAMIC_AUDIT_PROBE=W01 timeout 120 go test -overlay=/tmp/u099-W01-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^TestPrinterPreflightPlantedDisagreement$'; preflight_units_test.go:156: planted disagreement caught by 0 shards, want exactly one",
    "guard_kills": [
      "W01"
    ]
  },
  {
    "test": "TestPrinterFileDriver",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/printer_test.go",
    "seconds": 1.777,
    "oracle": "Hand-written query output with four-space indentation. Node and native execution must match it and exit zero; stderr is not independently required to be empty.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M04"
    ],
    "unique_kills": [
      "M01"
    ],
    "last_proven_fail": "M01: printer_test.go:131: driver: exit 0, stdout \"query {\\n   hello(a: 1, b: 2)\\n}\\n\", stderr \"\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^(TestPrinterPreflightPlantedDisagreement|TestPrinterFileDriver|TestPrinterShardUnionRejectsMissingAndRepeated|TestPrinterShardSelection|TestPrinterShardPlantedDisagreement|TestPrinterAsGoCohere_Union|TestPrinterAsGoCohere_000|TestPrinterAsGoCohere_001|TestPrinterAsGoCohere_002|TestPrinterAsGoCohere_003|TestPrinterUpstreamPreflight_Setup|TestPrinterUpstreamPreflight_000|TestPrinterUpstreamPreflight_001|TestPrinterUpstreamPreflight_002|TestPrinterUpstreamPreflight_003|TestPrinterWhitespaceGap_Setup|TestPrinterWhitespaceGap_000|TestPrinterWhitespaceGap_001|TestPrinterWhitespaceGap_002|TestPrinterWhitespaceGap_003|TestPrinterWhitespaceGap_004)$'; printer_test.go:131: driver: exit 0, stdout \"query {\\n   hello(a: 1, b: 2)\\n}\\n\", stderr \"\"",
    "unique_kills_scope": "Only this bounded matrix; package and repository uniqueness remain unknown."
  },
  {
    "test": "TestPrinterShardUnionRejectsMissingAndRepeated",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/shards_test.go",
    "seconds": 0.007,
    "oracle": "Self-written valid/missing/repeated/changed-input unions. Witnesses suite construction validation, including payload identity.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W03: shards_test.go:286: invalid union accepted",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier ADAMIC_AUDIT_PROBE=W03 timeout 120 go test -overlay=/tmp/u099-W03-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^TestPrinterShardUnionRejectsMissingAndRepeated$'; shards_test.go:286: invalid union accepted",
    "guard_kills": [
      "W03"
    ]
  },
  {
    "test": "TestPrinterShardSelection",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/shards_test.go",
    "seconds": 0.007,
    "oracle": "Self-written modulo assignment, complete/disjoint selection, invalid selector and unset selector expectations.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S01: shards_test.go:308: shard-000 assigned 0 times",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P03"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier ADAMIC_AUDIT_PROBE=S01 timeout 120 go test -overlay=/tmp/u099-S01-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^TestPrinterShardSelection$'; shards_test.go:308: shard-000 assigned 0 times",
    "guard_kills": [
      "S01"
    ]
  },
  {
    "test": "TestPrinterShardPlantedDisagreement",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/grain_asgo_test.go",
    "seconds": 1.32,
    "oracle": "Go cohere and Node controls; self-written planted mismatch must be rejected by exactly its owning shard. Production control failures do not prove this witness.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W02: grain_asgo_test.go:340: planted disagreement caught by 0 shards, want exactly one",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier ADAMIC_AUDIT_PROBE=W02 timeout 120 go test -overlay=/tmp/u099-W02-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^TestPrinterShardPlantedDisagreement$'; grain_asgo_test.go:340: planted disagreement caught by 0 shards, want exactly one",
    "guard_kills": [
      "W02"
    ]
  },
  {
    "test": "TestPrinterAsGoCohere family",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/grain_asgo_test.go",
    "seconds": 9.99,
    "oracle": "Exact Node source, native, JavaScript-backend and Linux leak-run stdout/stderr/exit compared with Go cohere. Four direct live oracle regenerations match cached answers byte-for-byte. Union member checks construction only.",
    "oracle_kind": "external-run",
    "kills": [
      "M02",
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M03"
    ],
    "last_proven_fail": "M03: grain_asgo_test.go:299: native: shard-003 tabs/case-000546: line 547: \"ok\\t## leading the file\\\\n## ends in JavaScript-only whitespace\\\\n\\\\n## before a definition\\\\ntype T { ## after the brace\\\\n\\\\t## own line before a field\\\\n\\\\ta: Int ## end of line\\\\n\\\\tb(\\\\n\\\\t\\\\t## before an argument\\\\n\\\\t\\\\tx: Int ## after an argument\\\\n\\\\t): Int\\\\n}\\\\nquery {\\\\n\\\\ta ## after a selection\\\\n\\\\t## between selections\\\\n\\\\tb\\\\n}\\\\n## end of file\\\\n\", Go cohere \"ok\\t# leading the file\\\\n# ends in JavaScript-only whitespace\\\\n\\\\n# before a definition\\\\ntype T { # after the brace\\\\n\\\\t# own line before a field\\\\n\\\\ta: Int # end of line\\\\n\\\\tb(\\\\n\\\\t\\\\t# before an argument\\\\n\\\\t\\\\tx: Int # after an argument\\\\n\\\\t): Int\\\\n}\\\\nquery {\\\\n\\\\ta # after a selection\\\\n\\\\t# between selections\\\\n\\\\tb\\\\n}\\\\n# end of file\\\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^(TestPrinterPreflightPlantedDisagreement|TestPrinterFileDriver|TestPrinterShardUnionRejectsMissingAndRepeated|TestPrinterShardSelection|TestPrinterShardPlantedDisagreement|TestPrinterAsGoCohere_Union|TestPrinterAsGoCohere_000|TestPrinterAsGoCohere_001|TestPrinterAsGoCohere_002|TestPrinterAsGoCohere_003|TestPrinterUpstreamPreflight_Setup|TestPrinterUpstreamPreflight_000|TestPrinterUpstreamPreflight_001|TestPrinterUpstreamPreflight_002|TestPrinterUpstreamPreflight_003|TestPrinterWhitespaceGap_Setup|TestPrinterWhitespaceGap_000|TestPrinterWhitespaceGap_001|TestPrinterWhitespaceGap_002|TestPrinterWhitespaceGap_003|TestPrinterWhitespaceGap_004)$'; grain_asgo_test.go:299: native: shard-003 tabs/case-000546: line 547: \"ok\\t## leading the file\\\\n## ends in JavaScript-only whitespace\\\\n\\\\n## before a definition\\\\ntype T { ## after the brace\\\\n\\\\t## own line before a field\\\\n\\\\ta: Int ## end of line\\\\n\\\\tb(\\\\n\\\\t\\\\t## before an argument\\\\n\\\\t\\\\tx: Int ## after an argument\\\\n\\\\t): Int\\\\n}\\\\nquery {\\\\n\\\\ta ## after a selection\\\\n\\\\t## between selections\\\\n\\\\tb\\\\n}\\\\n## end of file\\\\n\", Go cohere \"ok\\t# leading the file\\\\n# ends in JavaScript-only whitespace\\\\n\\\\n# before a definition\\\\ntype T { # after the brace\\\\n\\\\t# own line before a field\\\\n\\\\ta: Int # end of line\\\\n\\\\tb(\\\\n\\\\t\\\\t# before an argument\\\\n\\\\t\\\\tx: Int # after an argument\\\\n\\\\t): Int\\\\n}\\\\nquery {\\\\n\\\\ta # after a selection\\\\n\\\\t# between selections\\\\n\\\\tb\\\\n}\\\\n# end of file\\\\n\"",
    "members": [
      "TestPrinterAsGoCohere_Union",
      "TestPrinterAsGoCohere_000",
      "TestPrinterAsGoCohere_001",
      "TestPrinterAsGoCohere_002",
      "TestPrinterAsGoCohere_003"
    ],
    "probe_untouched_members": [
      "TestPrinterAsGoCohere_Union: construction-only member, not a port entry"
    ],
    "unique_kills_scope": "Only this bounded matrix; package and repository uniqueness remain unknown."
  },
  {
    "test": "TestPrinterUpstreamPreflight_Setup",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/top_level_units_test.go",
    "seconds": 1.437,
    "oracle": "Union of constructed mode/case IDs and input/answer pairings. Both sides derive from enumeratePrinter; zero enumerated IDs are accepted, as P04 proves.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S02: top_level_units_test.go:9: shard-000: missing, changed or repeated id defaults/case-000000",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier ADAMIC_AUDIT_PROBE=S02 timeout 120 go test -overlay=/tmp/u099-S02-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^TestPrinterUpstreamPreflight_Setup$'; top_level_units_test.go:9: shard-000: missing, changed or repeated id defaults/case-000000",
    "guard_kills": [
      "S02"
    ],
    "probe_entry": "enumeratePrinter, the value-producing enumeration used by this construction check",
    "probe_evidence": "P04.log: union: 0 unique mode/case ids across 4 shards; PASS"
  },
  {
    "test": "TestPrinterUpstreamPreflight family",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/top_level_units_test.go",
    "seconds": 5.97,
    "oracle": "Go cohere compared with npm Prettier 3.9.6 and embedded fork across four modes. Shared error-prefixed refusals need not have matching diagnostics; five self-recorded whitespace gaps per mode are required. The Adamic port is never executed.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "Source proof: printerUpstreamUnit in preflight_units_test.go executes only Go cohere and npm/embedded Prettier, no TS/native port. All four port mutants pass this row, and no permitted mutation of its protected oracles is available.",
    "members": [
      "TestPrinterUpstreamPreflight_000",
      "TestPrinterUpstreamPreflight_001",
      "TestPrinterUpstreamPreflight_002",
      "TestPrinterUpstreamPreflight_003"
    ],
    "reason": "Not caused by bounded scope. This is an oracle-preflight row with no permitted Adamic code-under-test mutation; changing Go cohere or Prettier would mutate an oracle."
  },
  {
    "test": "TestPrinterWhitespaceGap_Setup",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/top_level_units_test.go",
    "seconds": 1.183,
    "oracle": "Live Go cohere corpus answers must contain the expected Syntax Error refusal class; self-written union covers the five repository whitespace case IDs.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S03: gaps_test.go:93: shard union 4 ids, unsplit 5",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier ADAMIC_AUDIT_PROBE=S03 timeout 120 go test -overlay=/tmp/u099-S03-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^TestPrinterWhitespaceGap_Setup$'; gaps_test.go:93: shard union 4 ids, unsplit 5",
    "guard_kills": [
      "S03"
    ]
  },
  {
    "test": "TestPrinterWhitespaceGap family",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/top_level_units_test.go",
    "seconds": 5.21,
    "oracle": "Exact native and Node port refusals against Go cohere; npm and embedded Prettier must emit the self-recorded empty formatted answers. Error text is checked exactly on the port sides.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M04"
    ],
    "unique_kills": [],
    "last_proven_fail": "M04: gaps_test.go:129: native: shard-004 stage1/cohere/graphql/printer/gaps/whitespace-cases.json/case-000000: line 1: \"error\\tSyntax Error: Expected <EOF>, found <SOF>. (1:1)\", Go cohere \"error\\tSyntax Error: Unexpected <EOF>. (1:1)\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPrinterFileDriver"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 1.777,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterPreflightPlantedDisagreement",
      "TestPrinterFileDriver",
      "TestPrinterShardUnionRejectsMissingAndRepeated",
      "TestPrinterShardSelection",
      "TestPrinterShardPlantedDisagreement",
      "TestPrinterAsGoCohere family",
      "TestPrinterUpstreamPreflight_Setup",
      "TestPrinterUpstreamPreflight family",
      "TestPrinterWhitespaceGap_Setup",
      "TestPrinterWhitespaceGap family"
    ],
    "evidence": "ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^(TestPrinterPreflightPlantedDisagreement|TestPrinterFileDriver|TestPrinterShardUnionRejectsMissingAndRepeated|TestPrinterShardSelection|TestPrinterShardPlantedDisagreement|TestPrinterAsGoCohere_Union|TestPrinterAsGoCohere_000|TestPrinterAsGoCohere_001|TestPrinterAsGoCohere_002|TestPrinterAsGoCohere_003|TestPrinterUpstreamPreflight_Setup|TestPrinterUpstreamPreflight_000|TestPrinterUpstreamPreflight_001|TestPrinterUpstreamPreflight_002|TestPrinterUpstreamPreflight_003|TestPrinterWhitespaceGap_Setup|TestPrinterWhitespaceGap_000|TestPrinterWhitespaceGap_001|TestPrinterWhitespaceGap_002|TestPrinterWhitespaceGap_003|TestPrinterWhitespaceGap_004)$'; gaps_test.go:129: native: shard-004 stage1/cohere/graphql/printer/gaps/whitespace-cases.json/case-000000: line 1: \"error\\tSyntax Error: Expected <EOF>, found <SOF>. (1:1)\", Go cohere \"error\\tSyntax Error: Unexpected <EOF>. (1:1)\"",
    "members": [
      "TestPrinterWhitespaceGap_000",
      "TestPrinterWhitespaceGap_001",
      "TestPrinterWhitespaceGap_002",
      "TestPrinterWhitespaceGap_003",
      "TestPrinterWhitespaceGap_004"
    ],
    "subsumption_mutants": 1,
    "subsumption_limit": "Hint based on M04 only, not a deletion recommendation."
  }
]
```

| ID | origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | stage1/cohere/graphql/printer/doc.ts:26 | `    tabWidth: 4,` to `    tabWidth: 3,` | TestPrinterFileDriver |
| M02 | stage1/cohere/graphql/printer/doc.ts:201 | `doc: node.parts[command.flat ? 1 : 0] ?? panic('branch'),` to `doc: node.parts[command.flat ? 0 : 0] ?? panic('branch'),` | TestPrinterFileDriver, TestPrinterAsGoCohere family |
| M03 | stage1/cohere/graphql/printer/printer.ts:204 | `this.docs.text(`#${decoded(comment.value).trimEnd()}`)` to `this.docs.text(`##${decoded(comment.value).trimEnd()}`)` | TestPrinterAsGoCohere family |
| M04 | stage1/cohere/graphql/parser.ts:1228 | `this.expectToken('<SOF>');` to `this.expectToken('<EOF>');` | TestPrinterFileDriver, TestPrinterAsGoCohere family, TestPrinterWhitespaceGap family |

M02 and M04 also fail the planted-disagreement witness during its healthy control comparison. Those precondition failures are recorded in matrix.json and do not count as production kills for that witness.

| Guard/probe | origin/main file:line | Scope | Observed result |
|---|---|---|---|
| W01 | stage1/cohere/graphql/printer/preflight_units_test.go:51 | witness weakening | TestPrinterPreflightPlantedDisagreement: fail; Go vet passed |
| W02 | stage1/cohere/graphql/printer/shards_test.go:265 | witness weakening | TestPrinterShardPlantedDisagreement: fail; Go vet passed |
| W03 | stage1/cohere/graphql/printer/shards_test.go:219 | construction comparison weakening | TestPrinterShardUnionRejectsMissingAndRepeated: fail; Go vet passed |
| S01 | stage1/cohere/graphql/printer/shards_test.go:251 | construction condition flip | TestPrinterShardSelection: fail; Go vet passed |
| S02 | stage1/cohere/graphql/printer/preflight_units_test.go:87 | construction dropped statement | TestPrinterUpstreamPreflight_Setup: fail; Go vet passed |
| S03 | stage1/cohere/graphql/printer/gaps_test.go:158 | construction off-by-one | TestPrinterWhitespaceGap_Setup: fail; Go vet passed |
| P02 | stage1/cohere/graphql/printer/shards_test.go:206 | empty construction entry probe | TestPrinterShardUnionRejectsMissingAndRepeated: fail; Go vet passed |
| P03 | stage1/cohere/graphql/printer/shards_test.go:232 | empty construction entry probe | TestPrinterShardSelection: fail; Go vet passed |
| P04 | stage1/cohere/graphql/printer/shards_test.go:181 | empty construction entry probe | TestPrinterUpstreamPreflight_Setup: pass; Go vet passed |
| P05 | stage1/cohere/graphql/printer/gaps_test.go:153 | empty construction entry probe | TestPrinterWhitespaceGap_Setup: fail; Go vet passed |
| P01 | stage1/cohere/graphql/printer/main.ts:41 | Return before command output | File-driver, four AsGo semantic members and five whitespace members fail; construction-only Union passes |

Survivors: none among the four production mutants in the bounded matrix. Protected-oracle preflight passes are not treated as mutation survivors of the port. No equivalent candidates.

The brief and workflow friction:

- The brief is tied to 8de93800f4, while fetched origin/main is a467d1a1571c43e0f01fc4efcb43890280e4a1ac. TestPrinterAsGoCohere_Setup is absent from go test -list. Its current union-only counterpart is TestPrinterAsGoCohere_Union. The four AsGo wrappers moved to grain_asgo_test.go; the planted shard-disagreement witness moved there too. Scope was validated by names, not by the old file list.

- The header promises eleven rows and lists twenty-one functions. At the current commit, twenty listed names survive. Adding the current union member to its numbered family, as the family rule requires, yields ten actual rows. No object was fabricated for a nonexistent Setup function. Upstream and whitespace Setup functions remain separate construction-only rows; their numbered members run semantic comparisons.

- The full package baseline reached 90 seconds without any recorded assertion failure. Its timeout was a cooked run, not a green baseline. I narrowed to all current members of this unit plus the current AsGo union. That slice passed in 24.677s before mutation. All subsequent matrix runs completed within budget. Kills outside this slice are unknown.

- The four mode preflight units never execute the TypeScript/native port. They compare protected Go cohere and Prettier oracles. A production port mutant cannot prove their worth, and mutating those external implementations would violate the brief. Their cannot-judge verdict is about missing permitted code-under-test mutations, not bounded uniqueness. The separate planted preflight disagreement witness can be judged by allowed comparison weakening.

- Preflight deliberately accepts any pair of error-prefixed results as a shared refusal, without matching diagnostics, and requires exactly five recorded whitespace gaps. This is weaker than exact error agreement. The whitespace port family, in contrast, checks exact Go refusal text. The file-driver golden answer is hand-written and does not independently reject stderr when exit and stdout match.

- The upstream Setup constructs both sides of its union from the same enumeration. Returning nil from enumeratePrinter leaves four empty mode shards, logs zero union IDs, and passes. I probed this value-producing construction entry because Setup produces no port answer and does not call the port command. This probe choice is explicit in rows.json; it is a construction vacuity finding, not a port-entry claim.

- Production mutants M02/M04 break the shard witness's healthy control. Those failures cannot establish witness quality. W02 instead weakens its comparison and proves the planted mismatch is no longer caught, causing the witness to fail. W01 does the same for preflight. W03 loosens union payload comparison and proves changed input is wrongly accepted.

- The unit contains several guard tests, construction tests and a protected-oracle preflight, rather than eleven independent port semantics. Four production mutants were chosen from different code paths before observing outcomes. Six separate harness guard/construction changes and five entry probes were kept out of production kills and uniqueness. Whitespace subsumption rests on one broad parser-token mutant and is only a defender hint.

- printerLoweredProduct hashes the entire repository directory, including review evidence, rather than only compiler/port inputs. Writing logs under review during execution would change cache keys after every run. All evidence stayed under /tmp/u099/evidence until tests finished. This avoids audit artifacts repeatedly forcing native rebuilds, but central replay with a different repository evidence tree may still pay a new cold build.

- Warm tools lacked the requested npm Prettier/GraphQL scratch installation. I installed prettier 3.9.6 and graphql 17.0.2, then ran npm ci in that directory and stage3/api before baseline. All current unit members ran without skips. Outside this scope, the whole baseline recorded TestPrinterThroughput as opt-in skipped and TestProduct_GraphQLPrinterRelease as Darwin-only skipped; timeout prevents claiming a complete package skip inventory.

- Existing cached external answers were not simply trusted. Four direct live Go cohere oracle runs regenerated inputs and answers, matching the cached bytes exactly. Node V8 coverage over all four clean mode corpora records 171 reached function records in ten source modules. This substantiates source reach, but is not exact native C coverage; the complete named-source inventory remains conservative.

- Go probe overlays insert a line. Raw P02/P03 failure logs consequently show line numbers one higher than origin/main. The production and guard table uses original mutation locations; original failure locations for those two probes are shards_test.go:286 and :308. Raw logs are preserved without alteration.

- The default filesystem sandbox is read-only. Authorized writes, runs and push required escalated execution, and every automatic review approved them. No user approval was requested or action blocked. No compiler/runtime code or oracle data was mutated; the only harness changes are labelled guard/construction evidence.

Timing and coverage limits:

Toolchain setup was skipped because env.sh worked; nproc=5. npm's own install timing lines are saved separately. Fresh switched build cache misses were [('GraphQL_printer_source', 0.0), ('lowered_GraphQL_printer', 5.31), ('lowered_GraphQL_printer', 4.26), ('GraphQL_printer_split_native_sanitize=true', 3.71), ('sanitized_GraphQL_printer', 9.26)], totaling 22.54s of product build/checksum work, overlapping because tests run concurrently. These are not a single sequential rebuild duration.

Standalone source-only native validation wall seconds: {'M01': 8.683, 'M02': 6.053, 'M03': 5.684, 'M04': 5.677, 'P01': 8.562}, totaling 34.66s. Every production diff and P01 compiled through TestProduct_GraphQLPrinterSanitized with the port's native flags. Every Go harness/probe diff passed Go vet with its overlay.

Thirty clean timing runs totaled 94.336 test-binary seconds and 143.724 outer wall seconds. Medians: {'TestPrinterPreflightPlantedDisagreement': 4.332, 'TestPrinterFileDriver': 1.777, 'TestPrinterShardUnionRejectsMissingAndRepeated': 0.007, 'TestPrinterShardSelection': 0.007, 'TestPrinterShardPlantedDisagreement': 1.32, 'TestPrinterAsGoCohere family': 9.99, 'TestPrinterUpstreamPreflight_Setup': 1.437, 'TestPrinterUpstreamPreflight family': 5.97, 'TestPrinterWhitespaceGap_Setup': 1.183, 'TestPrinterWhitespaceGap family': 5.21}. Production/probe matrix binary seconds: {'M01': 23.328, 'M02': 23.689, 'M03': 23.2, 'M04': 16.357, 'P01': 16.41}. Whole-package baseline: 90.045s timeout; clean requested slice: 24.677s.

Not covered: full-package completion or uniqueness, tests outside the unit, repository-wide replay, Darwin-only leak products, throughput opt-in behavior, exhaustive native dynamic reachability, or mutation of protected preflight oracles. Current unit families ran completely, not sampled; every current row had three successful standalone timings.

Replay any M01.diff through M04.diff individually against the starting commit, build with the native product test, then use the recorded bounded matrix command. They contain no selector switch. For the amortized recorded runs, apply switch.diff and write the desired ID (or empty baseline) to /tmp/u099-mutant before execution. W/S/P Go diffs are separate; their exact overlay commands and required probe variables are in their run JSON. scope.json lists family members; rows.json and matrix.json contain the full machine-readable verdicts and observed outcomes.

Elapsed session work: approximately 24 minutes, including reading and evidence assembly. npm reported stage3 ci 360ms, pinned package install 783ms, and scratch package ci 629ms. These are npm timing lines, not separate command wall measurements.
