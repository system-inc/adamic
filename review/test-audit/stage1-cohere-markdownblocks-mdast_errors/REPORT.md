u128: all 29 supplied names exist at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2; grouped into 13 rows.
Clean whole package cooked at 90s; bounded combined run also cooked, without assertion failures; all selected rows passed alone.
Four port mutants caught; two bounded sacred rows, one slow-worthy, two mutually subsumed on one mutant.
Five setup-check rows, two witness rows, one cannot-judge compiler-initialization row; three readiness rows are vacuous.
Warm setup skipped; nproc 5; source restored; standalone diffs and all logs saved here.

CODE UNDER TEST AND ORACLES were named in commentary before mutation. Port entries compile(), observePath(), preprocess(), printQuote(); optional row calls Lower. Setup rows check suite construction. Preflight has no Adamic product, so only its witnessed check was weakened. Go cohere and Node fork source were never mutated by this audit. Existing tests build their own built-in mutants; this audit changed only its recorded production sites or permitted comparison/construction edits.

[
  {
    "test": "TestMdastMalformedEvents_Setup",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/mdast_errors_test.go:5",
    "seconds": 3.782,
    "oracle": "Build/readiness success of suite-owned malformed-event products; no runtime comparison in this row.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "G1 malformed_events_independent_test.go:252: build markdownblocks-malformed-events-go-errors: Go errors: exit status 2",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestMdastIdentifierWitnesses",
      "TestMdastMalformedEvents_Setup",
      "TestNativeMdastConstruction"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u128/G1-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMdastMalformedEvents_Setup)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/G1.log 2>&1; malformed_events_independent_test.go:252: build markdownblocks-malformed-events-go-errors: Go errors: exit status 2",
    "members": [
      "TestMdastMalformedEvents_Setup"
    ],
    "timing_samples": [
      3.782,
      4.16,
      3.422
    ],
    "entry_probe_results": {
      "P10": {
        "TestMdastMalformedEvents_Setup": "pass"
      }
    },
    "construction_kills": [
      "G1"
    ]
  },
  {
    "test": "TestMdastIdentifierWitnesses",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/mdast_gap_test.go:12",
    "seconds": 6.309,
    "oracle": "Actual Go cohere mdast and pinned Node fork, recorded Go/fork witness snapshots, and native/source/backend comparison with Go. Known casing disagreement is retained.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1 mdast_gap_test.go:84: native Go-held identifier witness first byte difference at 48 (lengths 1757/1754)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNativeMdastConstruction"
    ],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 34.608,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMdastIdentifierWitnesses",
      "TestMdastMalformedEvents_Setup",
      "TestNativeMdastConstruction"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMdastMalformedEvents_Setup|TestMdastIdentifierWitnesses|TestNativeMdastConstruction)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/M1.log 2>&1; mdast_gap_test.go:84: native Go-held identifier witness first byte difference at 48 (lengths 1757/1754)",
    "members": [
      "TestMdastIdentifierWitnesses"
    ],
    "timing_samples": [
      6.309,
      6.35,
      6.145
    ],
    "entry_probe_results": {
      "P1": {
        "TestMdastMalformedEvents_Setup": "pass",
        "TestMdastIdentifierWitnesses": "fail",
        "TestNativeMdastConstruction": "fail"
      }
    },
    "subsumption_mutants": 1
  },
  {
    "test": "TestNativeMdastConstruction",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/mdast_test.go:16",
    "seconds": 34.608,
    "oracle": "Actual Go cohere mdast construction and pinned Node fork; full canonical tree bytes compared with native/source/backend, plus sanitizer/leak checks.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1 mdast_test.go:128: native stage1/cohere/markdownblocks/GAPS.md byte66851 got\"  }\\\\\\\\\\\\\\\\n  10 |   const getIndex = ((): ((id: string) => number) => {\\\\\\\\\\\\\\\\n```\\\\\\\\\\\\\\\\n\\\"\\\\\\\\n````\\\\\\\\n\\\\nroot\\\\t-1\\\\tfalse\\\\tfalse\\\\t\\\\t0\\\\tfalse\\\\t-\\\\tfalse\\\\t-\\\\t-\\\\t-\\\\t\\\\t-\\\\t-\\\\t-\\\\t\\\\t\\\\t\\\\t1,1,0,837,1,63219\\\\t-\\\\t-\\\\t\\\\t\\\\t\\\\t\\\\nh\" want\"  }\\\\\\\\\\\\\\\\n  10 |   const getIndex = ((): ((id: string) => number) => {\\\\\\\\\\\\\\\\n```\\\\\\\\\\\\\\\\n\\\"\\\\\\\\n````\\\\\\\\n\\\\nroot\\\\t162\\\\tfalse\\\\tfalse\\\\t\\\\t0\\\\tfalse\\\\t-\\\\tfalse\\\\t-\\\\t-\\\\t-\\\\t\\\\t-\\\\t-\\\\t-\\\\t\\\\t\\\\t\\\\t1,1,0,837,1,63219\\\\t-\\\\t-\\\\t\\\\t\\\\t\\\\t\\\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestMdastIdentifierWitnesses"
    ],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 6.309,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMdastIdentifierWitnesses",
      "TestMdastMalformedEvents_Setup",
      "TestNativeMdastConstruction"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMdastMalformedEvents_Setup|TestMdastIdentifierWitnesses|TestNativeMdastConstruction)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/M1.log 2>&1; mdast_test.go:128: native stage1/cohere/markdownblocks/GAPS.md byte66851 got\"  }\\\\\\\\\\\\\\\\n  10 |   const getIndex = ((): ((id: string) => number) => {\\\\\\\\\\\\\\\\n```\\\\\\\\\\\\\\\\n\\\"\\\\\\\\n````\\\\\\\\n\\\\nroot\\\\t-1\\\\tfalse\\\\tfalse\\\\t\\\\t0\\\\tfalse\\\\t-\\\\tfalse\\\\t-\\\\t-\\\\t-\\\\t\\\\t-\\\\t-\\\\t-\\\\t\\\\t\\\\t\\\\t1,1,0,837,1,63219\\\\t-\\\\t-\\\\t\\\\t\\\\t\\\\t\\\\nh\" want\"  }\\\\\\\\\\\\\\\\n  10 |   const getIndex = ((): ((id: string) => number) => {\\\\\\\\\\\\\\\\n```\\\\\\\\\\\\\\\\n\\\"\\\\\\\\n````\\\\\\\\n\\\\nroot\\\\t162\\\\tfalse\\\\tfalse\\\\t\\\\t0\\\\tfalse\\\\t-\\\\tfalse\\\\t-\\\\t-\\\\t-\\\\t\\\\t-\\\\t-\\\\t-\\\\t\\\\t\\\\t\\\\t1,1,0,837,1,63219\\\\t-\\\\t-\\\\t\\\\t\\\\t\\\\t\\\\n\"",
    "members": [
      "TestNativeMdastConstruction"
    ],
    "timing_samples": [
      34.608,
      33.75,
      35.246
    ],
    "entry_probe_results": {
      "P1": {
        "TestMdastMalformedEvents_Setup": "pass",
        "TestMdastIdentifierWitnesses": "fail",
        "TestNativeMdastConstruction": "fail"
      }
    },
    "subsumption_mutants": 1
  },
  {
    "test": "TestOptionalStringInitializationWitness",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/optional_gap_test.go:10",
    "seconds": 0.33,
    "oracle": "Original source on Node decides missing; handwritten missing label is also checked, then native/backend output compared with Node.",
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
    "probe_kills": [
      "P5"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOptionalStringInitializationWitness"
    ],
    "evidence": "Three clean timings and P5.log; no production mutation claim.",
    "members": [
      "TestOptionalStringInitializationWitness"
    ],
    "timing_samples": [
      0.334,
      0.33,
      0.317
    ],
    "entry_probe_results": {
      "P5": {
        "TestOptionalStringInitializationWitness": "fail"
      }
    },
    "cannot_judge_reason": "No compiler initialization mutant was included in the fixed four-native-mutant menu. Mutating the optional-string source would change the Node oracle fixture; Lower was only empty-probed. This row remains unaudited for production worthiness and truth."
  },
  {
    "test": "TestMarkdownAstPath",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/path_test.go:17",
    "seconds": 60.292,
    "oracle": "Actual Go AstPath and pinned Node fork observations; full path bytes, plus handwritten key-gap values.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M2 path_test.go:130: native stage1/cohere/markdownblocks/GAPS.md byte28 got \"N0|N0||-1|U|0|-1|-1|0|-1|-1|false|false|false|false||0,-1,-1,-1,-1|-1|-1|false|true|true|false|false|false\\\\nN0,Pchildren,A0|A0|c\" want \"N0|N0||-1|U|0|-1|-1|0|-1|-1|true|false|false|false||0,-1,-1,-1,-1|-1|-1|false|true|true|false|false|false\\\\nN0,Pchildren,A0|A0|ch\"",
    "verdict": "slow-worthy",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownAstPath"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMarkdownAstPath)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/M2.log 2>&1; path_test.go:130: native stage1/cohere/markdownblocks/GAPS.md byte28 got \"N0|N0||-1|U|0|-1|-1|0|-1|-1|false|false|false|false||0,-1,-1,-1,-1|-1|-1|false|true|true|false|false|false\\\\nN0,Pchildren,A0|A0|c\" want \"N0|N0||-1|U|0|-1|-1|0|-1|-1|true|false|false|false||0,-1,-1,-1,-1|-1|-1|false|true|true|false|false|false\\\\nN0,Pchildren,A0|A0|ch\"",
    "members": [
      "TestMarkdownAstPath"
    ],
    "timing_samples": [
      60.292,
      61.872,
      60.083
    ],
    "entry_probe_results": {
      "P2": {
        "TestMarkdownAstPath": "fail"
      }
    }
  },
  {
    "test": "TestMarkdownParserPrefixes",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/prefix_test.go:16",
    "seconds": 2.142,
    "oracle": "Actual Go micromark preprocessing/prefix primitives; native/source/backend exact bytes compared with Go.",
    "oracle_kind": "external-run",
    "kills": [
      "M3"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M3 prefix_test.go:104: native parser prefixes differs at byte 86172 in \"generated/prefix/aa\\x00\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownParserPrefixes"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMarkdownParserPrefixes)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/M3.log 2>&1; prefix_test.go:104: native parser prefixes differs at byte 86172 in \"generated/prefix/aa\\x00\"",
    "members": [
      "TestMarkdownParserPrefixes"
    ],
    "timing_samples": [
      2.16,
      2.101,
      2.142
    ],
    "entry_probe_results": {
      "P3": {
        "TestMarkdownParserPrefixes": "fail"
      }
    }
  },
  {
    "test": "TestWholeDocumentOraclePreflightMutants",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/preflight_shards_test.go:416",
    "seconds": 2.069,
    "oracle": "Pinned Node fork output; row expects prebuilt Go mutants to disagree. W2 disables the witnessed difference check.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2 preflight_shards_test.go:458: mutant survived byte comparison",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestWholeDocumentOraclePreflightMutants"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u128/W2-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestWholeDocumentOraclePreflightMutants)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/W2.log 2>&1; preflight_shards_test.go:458: mutant survived byte comparison",
    "members": [
      "TestWholeDocumentOraclePreflightMutants"
    ],
    "timing_samples": [
      1.994,
      2.069,
      2.147
    ],
    "entry_probe_results": {},
    "witness_kills": [
      "W2"
    ]
  },
  {
    "test": "TestMarkdownQuoteLayoutNative",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/quote_layout_shards_test.go:64",
    "seconds": 4.22,
    "oracle": "Suite-owned native-product readiness; returned products are ignored by the top-level row.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "G2 quote_layout_shards_test.go:93: quote native setup failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownQuoteLayout family",
      "TestMarkdownQuoteLayoutNative",
      "TestMarkdownQuoteLayout_Setup"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u128/G2-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMarkdownQuoteLayoutNative|TestMarkdownQuoteLayout_Setup)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/G2.log 2>&1; quote_layout_shards_test.go:93: quote native setup failed",
    "members": [
      "TestMarkdownQuoteLayoutNative"
    ],
    "timing_samples": [
      4.22,
      4.359,
      4.101
    ],
    "entry_probe_results": {
      "P11": {
        "TestMarkdownQuoteLayoutNative": "pass"
      }
    },
    "construction_kills": [
      "G2"
    ]
  },
  {
    "test": "TestMarkdownQuoteLayout_Setup",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/quote_layout_shards_test.go:102",
    "seconds": 4.144,
    "oracle": "Suite-owned quote readiness; returned products are ignored by the top-level row.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "G2 quote_layout_shards_test.go:93: quote native setup failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownQuoteLayout family",
      "TestMarkdownQuoteLayoutNative",
      "TestMarkdownQuoteLayout_Setup"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u128/G2-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMarkdownQuoteLayoutNative|TestMarkdownQuoteLayout_Setup)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/G2.log 2>&1; quote_layout_shards_test.go:93: quote native setup failed",
    "members": [
      "TestMarkdownQuoteLayout_Setup"
    ],
    "timing_samples": [
      4.068,
      4.144,
      4.197
    ],
    "entry_probe_results": {
      "P12": {
        "TestMarkdownQuoteLayout_Setup": "pass"
      }
    },
    "construction_kills": [
      "G2"
    ]
  },
  {
    "test": "TestSampleRetainsFixedMarkdownInputs",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/sample_test.go:119",
    "seconds": 0.008,
    "oracle": "Handwritten selected labels, counts and numeric-answer bytes.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "G3 sample_test.go:125: lost fixed inputs: [{false b.md }] (1 files)",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P6",
      "P7",
      "P8"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestGeneratedLayoutSelection",
      "TestSampleRetainsFixedMarkdownInputs"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u128/G3-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestSampleRetainsFixedMarkdownInputs|TestGeneratedLayoutSelection)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/G3.log 2>&1; sample_test.go:125: lost fixed inputs: [{false b.md }] (1 files)",
    "members": [
      "TestSampleRetainsFixedMarkdownInputs"
    ],
    "timing_samples": [
      0.008,
      0.008,
      0.007
    ],
    "entry_probe_results": {
      "P6": {
        "TestSampleRetainsFixedMarkdownInputs": "fail",
        "TestGeneratedLayoutSelection": "fail"
      },
      "P7": {
        "TestSampleRetainsFixedMarkdownInputs": "fail",
        "TestGeneratedLayoutSelection": "fail"
      },
      "P8": {
        "TestSampleRetainsFixedMarkdownInputs": "fail"
      }
    },
    "construction_kills": [
      "G3"
    ]
  },
  {
    "test": "TestGeneratedLayoutSelection",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/sample_test.go:153",
    "seconds": 0.008,
    "oracle": "Handwritten generated/control selection and numeric index expectations.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "G3 sample_test.go:160: indexed corpus/control: []",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P6",
      "P7",
      "P9"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestGeneratedLayoutSelection",
      "TestSampleRetainsFixedMarkdownInputs"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u128/G3-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestSampleRetainsFixedMarkdownInputs|TestGeneratedLayoutSelection)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/G3.log 2>&1; sample_test.go:160: indexed corpus/control: []",
    "members": [
      "TestGeneratedLayoutSelection"
    ],
    "timing_samples": [
      0.009,
      0.008,
      0.007
    ],
    "entry_probe_results": {
      "P6": {
        "TestSampleRetainsFixedMarkdownInputs": "fail",
        "TestGeneratedLayoutSelection": "fail"
      },
      "P7": {
        "TestSampleRetainsFixedMarkdownInputs": "fail",
        "TestGeneratedLayoutSelection": "fail"
      },
      "P9": {
        "TestGeneratedLayoutSelection": "fail"
      }
    },
    "construction_kills": [
      "G3"
    ]
  },
  {
    "test": "TestWholeDocumentOraclePreflight family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/preflight_shards_test.go:341",
    "seconds": 3.263,
    "oracle": "Go cohere versus pinned Node fork, self-recorded auto gaps, union coverage and a planted off-output mismatch; W1 weakens the guarded comparison.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1 preflight_shards_test.go:385: planted disagreement did not reach exactly one shard",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestWholeDocumentOraclePreflight family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u128/W1-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestWholeDocumentOraclePreflightUnion|TestWholeDocumentOraclePreflight_000|TestWholeDocumentOraclePreflight_001|TestWholeDocumentOraclePreflight_002|TestWholeDocumentOraclePreflight_003)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/W1.log 2>&1; preflight_shards_test.go:385: planted disagreement did not reach exactly one shard",
    "members": [
      "TestWholeDocumentOraclePreflightUnion",
      "TestWholeDocumentOraclePreflight_000",
      "TestWholeDocumentOraclePreflight_001",
      "TestWholeDocumentOraclePreflight_002",
      "TestWholeDocumentOraclePreflight_003"
    ],
    "timing_samples": [
      3.308,
      3.263,
      2.986
    ],
    "entry_probe_results": {},
    "witness_kills": [
      "W1"
    ],
    "weakened_check_passes": [
      "TestWholeDocumentOraclePreflight_001",
      "TestWholeDocumentOraclePreflight_002",
      "TestWholeDocumentOraclePreflight_003"
    ]
  },
  {
    "test": "TestMarkdownQuoteLayout family",
    "package": "stage1/cohere/markdownblocks",
    "file": "stage1/cohere/markdownblocks/quote_layout_shards_test.go:135",
    "seconds": 15.759,
    "oracle": "Go cohere layout, pinned Node Markdown/doc printers and full output bytes; union coverage and planted mismatch use self expectations. W3 tests the built-in disagreement guard.",
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
    "last_proven_fail": "M4 quote_layout_shards_test.go:491: source Node lists output byte 53 in generated/block/> ```",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P4"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMarkdownQuoteLayout family",
      "TestMarkdownQuoteLayoutNative",
      "TestMarkdownQuoteLayout_Setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run ^(TestMarkdownQuoteLayoutNative|TestMarkdownQuoteLayout_Setup|TestMarkdownQuoteLayoutUnion|TestMarkdownQuoteLayout_000|TestMarkdownQuoteLayout_001|TestMarkdownQuoteLayout_002|TestMarkdownQuoteLayout_003|TestMarkdownQuoteLayout_004|TestMarkdownQuoteLayout_005|TestMarkdownQuoteLayout_006|TestMarkdownQuoteLayout_007|TestMarkdownQuoteLayout_008|TestMarkdownQuoteLayout_009|TestMarkdownQuoteLayout_010|TestMarkdownQuoteLayout_011)$ > /workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-mdast_errors/M4.log 2>&1; quote_layout_shards_test.go:491: source Node lists output byte 53 in generated/block/> ```",
    "members": [
      "TestMarkdownQuoteLayoutUnion",
      "TestMarkdownQuoteLayout_000",
      "TestMarkdownQuoteLayout_001",
      "TestMarkdownQuoteLayout_002",
      "TestMarkdownQuoteLayout_003",
      "TestMarkdownQuoteLayout_004",
      "TestMarkdownQuoteLayout_005",
      "TestMarkdownQuoteLayout_006",
      "TestMarkdownQuoteLayout_007",
      "TestMarkdownQuoteLayout_008",
      "TestMarkdownQuoteLayout_009",
      "TestMarkdownQuoteLayout_010",
      "TestMarkdownQuoteLayout_011"
    ],
    "timing_samples": [
      14.443,
      16.747,
      15.759
    ],
    "entry_probe_results": {
      "P4": {
        "TestMarkdownQuoteLayoutNative": "pass",
        "TestMarkdownQuoteLayout_Setup": "pass",
        "TestMarkdownQuoteLayoutUnion": "pass",
        "TestMarkdownQuoteLayout_000": "fail",
        "TestMarkdownQuoteLayout_001": "fail",
        "TestMarkdownQuoteLayout_002": "fail",
        "TestMarkdownQuoteLayout_003": "fail",
        "TestMarkdownQuoteLayout_004": "fail",
        "TestMarkdownQuoteLayout_005": "fail",
        "TestMarkdownQuoteLayout_006": "fail",
        "TestMarkdownQuoteLayout_007": "fail",
        "TestMarkdownQuoteLayout_008": "fail",
        "TestMarkdownQuoteLayout_009": "fail",
        "TestMarkdownQuoteLayout_010": "fail",
        "TestMarkdownQuoteLayout_011": "fail"
      }
    },
    "witness_kills": [
      "W3"
    ],
    "probe_not_called_by": [
      "TestMarkdownQuoteLayoutUnion"
    ]
  }
]

| ID | Origin file:line | Change | Failing grouped rows |
|---|---|---|---|
| M1 | stage1/cohere/markdownblocks/mdastCompile.ts:620 | `const root = this.arena.add('root', true, false); -> const root = this.arena.add('root', false, false);` | TestMdastIdentifierWitnesses, TestNativeMdastConstruction |
| M2 | stage1/cohere/markdownblocks/astPath.ts:82 | `return this.stack.length === 1; -> return this.stack.length === 0;` | TestMarkdownAstPath |
| M3 | stage1/cohere/markdownblocks/preprocess.ts:39 | `code: 65533 -> code: 65534` | TestMarkdownParserPrefixes |
| M4 | stage1/cohere/markdownblocks/quotes.ts:44 | `arena.text('> ') -> arena.text('# ')` | TestMarkdownQuoteLayout family |
| P1 | stage1/cohere/markdownblocks/mdastCompile.ts:619 | `empty entry` | TestMdastIdentifierWitnesses, TestNativeMdastConstruction |
| P2 | stage1/cohere/markdownblocks/pathObserve.ts:72 | `empty entry` | TestMarkdownAstPath |
| P3 | stage1/cohere/markdownblocks/preprocess.ts:7 | `empty entry` | TestMarkdownParserPrefixes |
| P4 | stage1/cohere/markdownblocks/quotes.ts:27 | `empty entry` | TestMarkdownQuoteLayout family |
| W1 | stage1/cohere/markdownblocks/preflight_shards_test.go:260 | `weaken witnessed comparison` | TestWholeDocumentOraclePreflight family |
| W2 | stage1/cohere/markdownblocks/preflight_shards_test.go:451 | `weaken witnessed comparison` | TestWholeDocumentOraclePreflightMutants |
| W3 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:172 | `weaken witnessed comparison` | TestMarkdownQuoteLayout family |
| G2 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:81 | `drop constructed product value` | TestMarkdownQuoteLayoutNative, TestMarkdownQuoteLayout_Setup |
| G3 | stage1/cohere/markdownblocks/sample_test.go:51 | `drop generated input condition` | TestGeneratedLayoutSelection, TestSampleRetainsFixedMarkdownInputs |
| G1 | stage1/cohere/markdownblocks/malformed_events_independent_test.go:374 | `change construction build option` | TestMdastMalformedEvents_Setup |
| P5 | internal/lower/lower.go:20 | `empty Lower entry` | TestOptionalStringInitializationWitness |
| P6 | stage1/cohere/markdownblocks/sample_test.go:35 | `empty construction entry` | TestGeneratedLayoutSelection, TestSampleRetainsFixedMarkdownInputs |
| P7 | stage1/cohere/markdownblocks/sample_test.go:58 | `empty construction entry` | TestGeneratedLayoutSelection, TestSampleRetainsFixedMarkdownInputs |
| P8 | stage1/cohere/markdownblocks/sample_test.go:104 | `empty construction entry` | TestSampleRetainsFixedMarkdownInputs |
| P9 | stage1/cohere/markdownblocks/sample_test.go:16 | `empty construction entry` | TestGeneratedLayoutSelection |
| P10 | stage1/cohere/markdownblocks/malformed_events_independent_test.go:238 | `empty construction entry` |  |
| P11 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:69 | `empty construction entry` |  |
| P12 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:107 | `empty construction entry` |  |

Survivors:
None among M1-M4. Each changed behavior and was caught after native build. W/G edits are witness/construction evidence, not production survivors. Passing P10/P11/P12 are empty-answer findings, not survivors or sacred kills.

Brief ambiguities, costs and limits:
- The brief says 13 rows and lists 29 functions. Grouping the numbered preflight and quote shards with their unions produces exactly 13. The union members assert coverage and a planted mismatch rather than invoking the same runtime checker. This follows the brief's explicit numbered-shards-plus-union family rule; readiness functions and the full-corpus preflight mutant witness remain separate. Members are listed exhaustively.
- All 29 names remain in the supplied files at this starting commit; none moved or vanished. Current inventory, not the historical 8de93800f4 file mapping, determined scope.
- Whole package 90.041s cooked during TestMarkdownListLayout_Setup, outside this unit. Bounded all-29 run 90.037s also cooked with NativeMdastConstruction and AstPath unfinished. They then passed alone in 34.953s and approximately 62s. Thirty-nine isolated timing runs all passed. No red assertion baseline was audited.
- Native port/compiler run limit of four overrides the approximate three-mutants-per-row target. Four constants were chosen from separate reached code functions before seeing kills; no test-derived string failure was used as an oracle mutation. Only four production mutations support these verdicts.
- Optional-string initialization has no separate markdown port implementation. Its source is the original Node fixture; mutating it would mutate the oracle input. A compiler initialization mutation was not included under the four-mutant menu. This is explicit incomplete coverage, marked cannot-judge, not a claim that a meaningful compiler mutant is impossible. P5 proves it rejects an empty Lower answer but cannot establish worthiness or production truth.
- Empty probes are additional to the production menu. P1-P4 were validated through native compilation and execution in their target tests; P5 returns nil IR and the already isolated optional row fails with a recovered Go panic. No later row was silently counted. Go probe early returns use an always-true branch to retain referenced imports and make standalone diffs vet-clean.
- Readiness functions are the setup rows' direct construction entries. P10/P11/P12 return an empty product set at entry, and each top-level row ignores that return. Thus they pass their own probes and are vacuous even though G1/G2 demonstrate construction failures. Port probes passing a compile-only setup row are not used to judge that row.
- Sampling checks call multiple construction entries. P6-P9 probe each selected input/numeric/answer/selection entry directly. Each relevant probe fails; no vacuous inference comes from an unprobed helper.
- Preflight compares Go cohere against Node, with no Adamic port. Production mutants would be wrong here. W1 weakens the off comparison: the union and shard 000 fail their planted guard, while shards 001-003 still pass. W2 makes the separate built-in-mutant witness fail. These are witness verdicts, not production kills or uniqueness claims.
- Quote runtime checks include both real external comparisons and built-in source mutants. M4 proves production failure. W3 independently disables the shared byte check and makes union/runtime planted-mismatch checks fail. The family keeps its production verdict; the extra witness evidence is recorded separately.
- Mdast identifier and optional rows contain Witness in their names but run actual native behavior checks, not solely planted disagreement checks. Their labels alone do not determine witness classification. Identifier also checks self-recorded snapshots and a known Go/fork gap.
- M1 is shared by Identifier and NativeMdastConstruction. Reciprocal subsumption rests on exactly one mutant, not a deletion recommendation. Identifier's measured median is much cheaper. Path's bounded unique kill plus median over 60s supports slow-worthy. Package/repository uniqueness outside each named caller set is unknown.
- The static inventory follows transitive relative imports from the actual probe entries and lists functions/methods conservatively. It is a reachability superset, not dynamic proof that every method executed. Mutation sites themselves are confirmed reached by failures. Compiler internals beyond the empty Lower entry were not exhaustively traced. This limitation prevents claiming a complete function-level coverage audit.
- Source mutations were applied sequentially as standalone diffs, then restored. Each source change invalidates content-keyed native products. Go harness edits use overlays, keeping source files and native products untouched. Construction G1/G2 use their own cache directories. P5 uses /tmp/u128/cache/P5, so the compiler probe cannot read an old native product.
- G1 deliberately changes the setup's child build output flag and catches construction failure; it does not mutate Go cohere or count a child compile failure as a production kill. G2 drops the constructed sanitized-product value; G3 drops generated-input inclusion. Their parent Go code passes vet.
- No full package mutation matrix was rerun after its clean timeout. Per-mutation row lists are explicit in matrix-commands.json and matrix-top-level.json. Results outside each list are unknown, not passes. All verdicts are bounded accordingly.
- Timing medians use three isolated -count=1 binary ok lines, including each family as a complete selected run. They measure warm product-cache behavior where the test uses caches; direct native build rows still rebuild. Outer Go compilation is reported separately.
- Node dependencies in stage3/api were installed before baseline. Selected rows use pinned fork bundles or the repository Node runner, not additional node_modules directories. No selected row skipped. The full package's width opt-in row did not reach execution before timeout; its external npm SDK was not installed or enabled in this bounded unit. Other package skips remain unknown.
- Function/source/test reads and two cooked baseline attempts consumed budget. No setup installation was needed. Go overlay failure line numbers in empty probes can shift by one; raw logs retain scratch positions. Origin locations in the mutant table and row file fields are authoritative; production/W/G edits do not shift lines.
- Every diff applies to the starting commit. All Go edits pass vet via their exact overlays; port edits compile with the tests' native build flags. rebuilds.json retains product lowering/native-build times. Direct natively builds do not emit separate build timers, so command wall is an upper bound rather than an invented native-only time.

Build/run timing:
Setup 0s; nproc 5. npm install not separately timed. Whole baseline shell 92.311s, binary 90.041s cooked; bounded shell 92.071s, binary 90.037s cooked. Isolated diagnostic baseline shells 37.111s and 64.118s. Thirty-nine timing commands total 493.125s shell wall. Matrix/probe commands total 421.418s shell wall. Exact-overlay Go vet checks total 4.589s.
M1-M4 command walls and explicit per-product rebuild measurements are in matrix-commands.json/rebuilds.json. M4 lowered 35.73s, release native 6.79s, sanitized native 24.93s. P4 has its own rebuilds, not a stale product. Initial whole-package compile/setup and native builds overlap, so these sums are work totals, not elapsed session time.
Audit started about 12:45 UTC and evidence completed within the 30-minute stage1 budget. Final restoration and smoke logs retained.

Not covered: package rows outside the supplied set; complete dynamic function reachability; compiler initialization production mutation; all port branches; census/full external corpora; SDK-gated width oracle; repository-wide uniqueness. No PR or main push.
