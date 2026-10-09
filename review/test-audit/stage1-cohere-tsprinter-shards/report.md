Unit u141 at a7448d73cd17f16362b6cbc5c5c111080da64e43: 37 requested functions, 13 live grouped rows.
Corrected whole baseline timed out; bounded baseline passed in 38.952 test seconds without skips.
Bounded verdicts: two sacred families, four setup-checks, six witnesses, one untrue setup row.
Four port mutants, five construction mutations, five weakened checks, and two separate probes are recorded.
Source restored; evidence pushed on the requested audit branch without a PR or main push.

[
  {
    "test": "TestCorpusShardAssignment",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/shards_test.go",
    "seconds": 0.009,
    "oracle": "Handwritten largest-first assignment.",
    "oracle_kind": "self",
    "kills": [
      "S1"
    ],
    "unique_kills": [
      "S1"
    ],
    "last_proven_fail": "S1: largest-first assignment [[0 4 1] [2 5 3]], want [[1 5 0] [3 4 2]]",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsAgainstGoAndPrettier_Setup",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestStatementsAgainstGoAndPrettier family",
      "TestTSCCorpusAgreement family",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [S1]; command in S1.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; largest-first assignment [[0 4 1] [2 5 3]], want [[1 5 0] [3 4 2]]",
    "samples": [
      0.008,
      0.009,
      0.01
    ],
    "members": [
      "TestCorpusShardAssignment"
    ],
    "witness_kills": []
  },
  {
    "test": "TestCorpusShardCoverage",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/shards_test.go",
    "seconds": 0.009,
    "oracle": "Handwritten missing, extra, and unterminated-answer rejection messages.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: error <nil>, want first.ts:0: unterminated answer",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [W1]; command in W1.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; error <nil>, want first.ts:0: unterminated answer",
    "samples": [
      0.009,
      0.009,
      0.008
    ],
    "members": [
      "TestCorpusShardCoverage"
    ],
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestCorpusShardTransport",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/shards_test.go",
    "seconds": 0.056,
    "oracle": "Node runs a custom echo child; expected bytes and ordering are handwritten.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "S2"
    ],
    "unique_kills": [],
    "last_proven_fail": "S2: transport: exit 0 stderr  diff fixture-00.ts:0: line 1: \"ok\\txxxxxxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxx\\rok\\txxxxxxxxxxxx\\rok\\txxxxxxxxxxx\\rok\\txxxxxxxxxx\\rok\\txxxxxxxxx\\rok\\txxxxxxxx\\rok\\txxxxxxx\\rok\\txxxxxx\\rok\\txxxxx\\rok\\txxxx\\rok\\txxx\\rok\\txx\\r\", Go cohere \"ok\\txxxxxxxxxxxxxxxxxxxxxxxx\"",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsAgainstGoAndPrettier_Setup",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestStatementsAgainstGoAndPrettier family",
      "TestTSCCorpusAgreement family",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [S2]; command in S2.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; transport: exit 0 stderr  diff fixture-00.ts:0: line 1: \"ok\\txxxxxxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxxx\\rok\\txxxxxxxxxxxxx\\rok\\txxxxxxxxxxxx\\rok\\txxxxxxxxxxx\\rok\\txxxxxxxxxx\\rok\\txxxxxxxxx\\rok\\txxxxxxxx\\rok\\txxxxxxx\\rok\\txxxxxx\\rok\\txxxxx\\rok\\txxxx\\rok\\txxx\\rok\\txx\\r\", Go cohere \"ok\\txxxxxxxxxxxxxxxxxxxxxxxx\"",
    "samples": [
      0.056,
      0.057,
      0.056
    ],
    "members": [
      "TestCorpusShardTransport"
    ],
    "witness_kills": []
  },
  {
    "test": "TestStatementsAgainstGoAndPrettier_Setup",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/statements_printer_units_test.go",
    "seconds": 2.99,
    "oracle": "Successful preparation only; S5 returned main.ts as the statement port and this row still passed.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsAgainstGoAndPrettier_Setup",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestStatementsAgainstGoAndPrettier family",
      "TestTSCCorpusAgreement family",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [S5]; command in S5.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; S5: setup passed with the wrong returned port path; the statement family caught it.",
    "samples": [
      3.025,
      2.903,
      2.99
    ],
    "members": [
      "TestStatementsAgainstGoAndPrettier_Setup"
    ],
    "witness_kills": []
  },
  {
    "test": "TestStatementsShardAssignmentStable",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/statements_shards_test.go",
    "seconds": 0.009,
    "oracle": "Handwritten stability under corpus growth and duplicate-key rejection.",
    "oracle_kind": "self",
    "kills": [
      "S3"
    ],
    "unique_kills": [
      "S3"
    ],
    "last_proven_fail": "S3: repeated stable key accepted",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsAgainstGoAndPrettier_Setup",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestStatementsAgainstGoAndPrettier family",
      "TestTSCCorpusAgreement family",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [S3]; command in S3.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; repeated stable key accepted",
    "samples": [
      0.009,
      0.009,
      0.009
    ],
    "members": [
      "TestStatementsShardAssignmentStable"
    ],
    "witness_kills": []
  },
  {
    "test": "TestStatementsShardSelection",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/statements_shards_test.go",
    "seconds": 0.009,
    "oracle": "Handwritten exact-once selection and invalid-selector rejection.",
    "oracle_kind": "self",
    "kills": [
      "S4"
    ],
    "unique_kills": [
      "S4"
    ],
    "last_proven_fail": "S4: 1 boxes: shard-000 selected 0 times",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsAgainstGoAndPrettier_Setup",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestStatementsAgainstGoAndPrettier family",
      "TestTSCCorpusAgreement family",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [S4]; command in S4.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; 1 boxes: shard-000 selected 0 times",
    "samples": [
      0.009,
      0.009,
      0.008
    ],
    "members": [
      "TestStatementsShardSelection"
    ],
    "witness_kills": []
  },
  {
    "test": "TestStatementsShardDisagreement",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/statements_shards_test.go",
    "seconds": 1.656,
    "oracle": "Handwritten planted owner, failure count, exit status, and case identity; a positive Node port run is a prerequisite.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: See complete failure output in W2.log",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [W2]; command in W2.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; See complete failure output in W2.log",
    "samples": [
      1.592,
      1.656,
      1.658
    ],
    "members": [
      "TestStatementsShardDisagreement"
    ],
    "witness_kills": [
      "W2"
    ]
  },
  {
    "test": "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/statements_shards_test.go",
    "seconds": 0.008,
    "oracle": "Handwritten rejection of missing, repeated, extra IDs and wrong shard count.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W3: invalid union accepted",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [W3]; command in W3.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; invalid union accepted",
    "samples": [
      0.01,
      0.008,
      0.007
    ],
    "members": [
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs"
    ],
    "witness_kills": [
      "W3"
    ]
  },
  {
    "test": "TestStatementsAgainstGoAndPrettier family",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/statements_printer_units_test.go",
    "seconds": 26.43,
    "oracle": "Go cohere supplies expected bytes; npm/embedded Prettier checks exact outcomes with pinned differences. Node port, native, and backend compare against Go. Gap refusal labels are handwritten. Leak checks also run.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2",
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4: Node case 1039 /workspace/adamic/stage1/cohere/css/parse_value.ts:4888:KindVariableStatement: line 1: \"ok\\tconst variable =\\\\n    scss && tree.maybe(first).type() === 'word' &&\\\\n    tree.maybe(first).string('value').startsWith('$');\\\\n\", Go cohere \"ok\\tconst variable =\\\\n    scss &&\\\\n    tree.maybe(first).type() === 'word' &&\\\\n    tree.maybe(first).string('value').startsWith('$');\\\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestStatementsShardDisagreement",
      "TestStatementsAgainstGoAndPrettier family",
      "TestTSCCorpusAgreement family",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement"
    ],
    "evidence": "run-matrix.py [M4]; command in M4.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; Node case 1039 /workspace/adamic/stage1/cohere/css/parse_value.ts:4888:KindVariableStatement: line 1: \"ok\\tconst variable =\\\\n    scss && tree.maybe(first).type() === 'word' &&\\\\n    tree.maybe(first).string('value').startsWith('$');\\\\n\", Go cohere \"ok\\tconst variable =\\\\n    scss &&\\\\n    tree.maybe(first).type() === 'word' &&\\\\n    tree.maybe(first).string('value').startsWith('$');\\\\n\"",
    "samples": [
      27.296,
      26.43,
      24.199
    ],
    "members": [
      "TestStatementsAgainstGoAndPrettier_000",
      "TestStatementsAgainstGoAndPrettier_001",
      "TestStatementsAgainstGoAndPrettier_002",
      "TestStatementsAgainstGoAndPrettier_003",
      "TestStatementsAgainstGoAndPrettier_004",
      "TestStatementsAgainstGoAndPrettier_005",
      "TestStatementsAgainstGoAndPrettier_006",
      "TestStatementsAgainstGoAndPrettier_007",
      "TestStatementsAgainstGoAndPrettier_008",
      "TestStatementsAgainstGoAndPrettier_009",
      "TestStatementsAgainstGoAndPrettier_010",
      "TestStatementsAgainstGoAndPrettier_011",
      "TestStatementsAgainstGoAndPrettier_012",
      "TestStatementsAgainstGoAndPrettier_013",
      "TestStatementsAgainstGoAndPrettier_014",
      "TestStatementsAgainstGoAndPrettier_015",
      "TestStatementsAgainstGoAndPrettierUnion"
    ],
    "witness_kills": []
  },
  {
    "test": "TestTSCCorpusAgreement family",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/tsc_printer_units_test.go",
    "seconds": 6.137,
    "oracle": "Go cohere supplies expected bytes. Node port, native, and backend check exact stdout, stderr and exit status; leak checks also run.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M1"
    ],
    "last_proven_fail": "M3: See complete failure output in M3.log",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsAgainstGoAndPrettier_Setup",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestStatementsAgainstGoAndPrettier family",
      "TestTSCCorpusAgreement family",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [M3]; command in M3.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; See complete failure output in M3.log",
    "samples": [
      6.165,
      6.137,
      5.835
    ],
    "members": [
      "TestTSCCorpusAgreement_Union",
      "TestTSCCorpusAgreement_000",
      "TestTSCCorpusAgreement_001",
      "TestTSCCorpusAgreement_002",
      "TestTSCCorpusAgreement_003",
      "TestTSCCorpusAgreement_004",
      "TestTSCCorpusAgreement_005",
      "TestTSCCorpusAgreement_006",
      "TestTSCCorpusAgreement_007"
    ],
    "witness_kills": []
  },
  {
    "test": "TestTSCCorpusAgreement_PlantedDisagreement",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/tsc_corpus_split_test.go",
    "seconds": 4.951,
    "oracle": "Go-derived positive answers plus the handwritten exactly-one-owner planted failure invariant.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W4: caught by 0 shards, want exactly one",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [W4]; command in W4.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; caught by 0 shards, want exactly one",
    "samples": [
      4.672,
      4.951,
      5.034
    ],
    "members": [
      "TestTSCCorpusAgreement_PlantedDisagreement"
    ],
    "witness_kills": [
      "W4"
    ]
  },
  {
    "test": "TestTSCShardPlantedDisagreement",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/tsc_units_test.go",
    "seconds": 0.355,
    "oracle": "Handwritten positive answers and exactly-one-owner planted failure invariant.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W4: planted disagreement caught by 0 shards, want exactly one",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [W4]; command in W4.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; planted disagreement caught by 0 shards, want exactly one",
    "samples": [
      0.332,
      0.355,
      0.367
    ],
    "members": [
      "TestTSCShardPlantedDisagreement"
    ],
    "witness_kills": [
      "W4"
    ]
  },
  {
    "test": "TestTSCShardUnionRejectsMissingAndRepeated",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/tsc_units_test.go",
    "seconds": 0.009,
    "oracle": "Handwritten rejection of missing and repeated IDs.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W5: missing id accepted",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCorpusShardAssignment",
      "TestCorpusShardCoverage",
      "TestCorpusShardTransport",
      "TestStatementsShardAssignmentStable",
      "TestStatementsShardSelection",
      "TestStatementsShardDisagreement",
      "TestStatementsShardUnionRejectsMissingAndRepeatedIDs",
      "TestTSCCorpusAgreement_PlantedDisagreement",
      "TestTSCShardPlantedDisagreement",
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "evidence": "run-matrix.py [W5]; command in W5.json: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run <recorded selector>; missing id accepted",
    "samples": [
      0.009,
      0.009,
      0.01
    ],
    "members": [
      "TestTSCShardUnionRejectsMissingAndRepeated"
    ],
    "witness_kills": [
      "W5"
    ]
  }
]

| ID | Origin/main location | Change | Observed failing rows |
|---|---|---|---|
| M1 | stage1/cohere/tsprinter/expressions.ts:1811 | docs.concat([printed, docs.text(','), docs.hardline()]) | TestTSCCorpusAgreement family, TestTSCCorpusAgreement_PlantedDisagreement, TestTSCShardPlantedDisagreement |
| M2 | stage1/cohere/tsprinter/expressions.ts:1821 | text: printed === '' ? '' : `${printed} ` | TestStatementsShardDisagreement, TestStatementsAgainstGoAndPrettier family, TestTSCCorpusAgreement family, TestTSCCorpusAgreement_PlantedDisagreement |
| M3 | stage1/cohere/tsprinter/literals.ts:3 | if(raw.length !== 1) | TestStatementsAgainstGoAndPrettier family, TestTSCCorpusAgreement family, TestTSCCorpusAgreement_PlantedDisagreement |
| M4 | stage1/cohere/tsprinter/syntax.ts:95 | rank(parent) !== rank(child) || parent.slice(0) !== '**' | TestStatementsAgainstGoAndPrettier family |
| S1 | stage1/cohere/tsprinter/shards_test.go:40 | return weights[order[i]] < weights[order[j]] | TestCorpusShardAssignment |
| S2 | stage1/cohere/tsprinter/shards_test.go:168 | result.WriteByte('\r') | TestCorpusShardTransport, TestTSCCorpusAgreement family, TestTSCCorpusAgreement_PlantedDisagreement, TestTSCShardPlantedDisagreement, TestTSCShardUnionRejectsMissingAndRepeated |
| S3 | stage1/cohere/tsprinter/statements_shards_test.go:168 |  | TestStatementsShardAssignmentStable |
| S4 | stage1/cohere/tsprinter/statements_shards_test.go:387 | return shard%n != i | TestStatementsShardSelection |
| S5 | stage1/cohere/tsprinter/statements_test.go:117 | filepath.Abs("main.ts") | TestStatementsAgainstGoAndPrettier family |
| W1 | stage1/cohere/tsprinter/shards_test.go:134 | func mergeCorpus(plan *corpusPlan, outputs [][]byte) ([]byte, error) { 	return nil, nil } | TestCorpusShardCoverage, TestCorpusShardTransport, TestTSCCorpusAgreement_PlantedDisagreement, TestTSCShardPlantedDisagreement, TestTSCShardUnionRejectsMissingAndRepeated |
| W2 | stage1/cohere/tsprinter/statements_shards_test.go:427 | func statementDisagreement(name string, result run, want string, labels []string) error { 	return nil } | TestStatementsShardDisagreement |
| W3 | stage1/cohere/tsprinter/statements_shards_test.go:245 | func statementUnion(count int, shards [][]int) error { 	return nil } | TestStatementsShardUnionRejectsMissingAndRepeatedIDs |
| W4 | stage1/cohere/tsprinter/expression_units_test.go:217 | func expressionDisagreement(plan *corpusPlan, number int, want []byte, result run) error { 	return nil } | TestTSCCorpusAgreement_PlantedDisagreement, TestTSCShardPlantedDisagreement |
| W5 | stage1/cohere/tsprinter/expression_units_test.go:186 | func expressionUnion(plan *corpusPlan, outputs [][]byte, want []byte) error { 	return nil } | TestTSCShardUnionRejectsMissingAndRepeated |

Survivors: none among the planted mutations in the observed bounded runs. S5 passed the setup row but failed the statement family, so it is a row-level blind spot rather than a globally surviving mutation.

Brief issues and costs

The intro says 15 rows. The 38 supplied functions group into 14 rows under the family rules with witnesses separate and unions grouped with their agreement families. TestStatementsSetupSelection vanished, leaving 13 live rows. Statement wrappers, union and setup moved to statements_printer_units_test.go; TSC wrappers moved to tsc_printer_units_test.go. Each member is listed in rows.json. All mutant locations refer to starting origin/main a7448d73, not the historical 8de93800f4.

The keeper's TypeScript requirement was satisfied before the baseline: a git checkout at 050880ce59e30b356b686bd3144efe24f875ebc8. The first baseline nevertheless hit a setup error because the oracle-input hash refuses npm's .bin/prettier symlink. npm ci --no-bin-links repaired it. No mutation audit ran on that setup-red baseline. The corrected whole run timed out at 90 seconds without a preceding row failure. The bounded baseline passed all 37 surviving requested functions without skips.

Production witness failures are not verdict-bearing catches. M1 broke two witness positive preconditions. M2/M3/probes also broke some prerequisites. The matrix table records those observations; only W1 through W5 determine the six witness verdicts. M1 is unique among applicable production agreement rows, despite two witness precondition failures. Its unique_kills entry is not a claim that exactly one top-level function failed. Both sacred verdicts are bounded, and package-wide uniqueness remains unknown. M4 failed only the statement family after members were grouped.

Setup mutations are separate from port changes. S1, S2, S3 and S4 prove four setup-check verdicts. S5 changes statementPrepareCommon's returned Port from statementsMain.ts to main.ts. The setup row only calls preparation, so it still passes; the statement family rejects the wrong program. This supports untrue for that row under these construction attempts, not a claim that every preparation error is invisible or that the row should be deleted.

Probes target formatExpression and formatFile, the formatter entries used by the two drivers. They return Ok with empty text while the drivers still emit protocol rows. Both families reject their applicable probe answers. Probe IDs never support sacred, subsumed, or witness verdicts. Other rows were not probed on their own construction entries and have vacuous=null. Probe diffs are separate from mutant diffs.

M4's exact standalone flip first failed TS2367: the early return narrowed parent to '**', making a later '%' comparison disjoint. The switched version had avoided the narrowing. The final one-line diff compares parent.slice(0), an identity operation on its string argument, to prevent this typing issue without inserting a statement or changing the intended condition flip. The native build passed. Its first standalone matrix timed out while cold-building statement products and external oracle observations: fourteen statement shards failed and all eight TSC shards passed before the timeout. The final narrowed warm replay supplies completed family observations. Initial diagnostic, switched run, cooked run, and final replay logs are all retained. The repair was driven by the compiler diagnostic, not by which row caught M4.

W1 replaces the whole merger body and removes its now-unused bytes import. W2 through W5 replace checker bodies with return nil. These are the explicitly allowed witness weakenings. Every standalone Go diff applies to origin/main and passes go vet. Every final standalone port diff compiled through sanitized and release native product builds. No compile-warning kill supports a verdict. Source and helper instrumentation were restored afterward, verified against origin/main.

The function inventory is a conservative transitive declaration and helper-call inventory, not exact per-case dynamic coverage. It includes candidate methods whose precise case reachability was not resolved. Mutant sites were chosen from that source inventory before observing catches. The port switch reads a selector file at module initialization; both switched native products were built once and reused across selectors. No compiler or external oracle was mutated. No ADAMIC_NATIVE_SPLIT cache-key assumption was used. The neutral switched run passed all requested rows after refreshing external observations for changed self-corpus inputs.

The requested remote branch already contains keeper attempt aabd7a7d. Its history is preserved by a merge, not a force push. None of its timings or results support this audit. Raw logs are force-added because repository ignore rules exclude .log files. No PR or main push was made.

Timing and omissions

Warm setup skipped; nproc=5. npm wall seconds: npm-api 0.563, npm-prettier 0.802, npm-prettier-no-bin 1.497.
Whole baseline command wall seconds: 92.168, corrected 95.377. Bounded baseline 41.156s wall, 38.952s test time.
Thirty-nine timing command wall seconds: 210.179. Switched product build 57.512s wall, 50.539s test time.
Standalone native validation wall seconds: M1 24.951, M2 31.059, M3 26.463, M4 41.325. First rejected M4 build: 4.038s.
Neutral and final mutation/probe commands wall seconds: 428.785. Cooked M4 and initial switched M4 timings are separate in their JSON files.
Whole-package and repository-wide uniqueness, exact dynamic function coverage, and other construction-entry probes remain uncovered. No other package was used as an audit target. The whole package did not complete within 90 seconds. Outside-slice kills are unknown. M4 narrowing is recorded in its command selector; other construction/port matrices ran the 37 requested functions, and weakened checks ran the ten fast non-family/non-setup rows. No deletion is recommended.
