u109 started at origin/main 16f436a16a8e3b9cd2343f449eb1b0b4577c135c; all 31 requested names exist, grouped into 15 rows.
The clean whole package cooked at 90.186 seconds; the restored bounded run passed in 27.645 seconds.
Verdicts: 9 setup-check, 2 witness, 2 limited untrue, 1 subsumed, 1 cannot-judge; no unique production kills proved.
Three production mutants, nine construction breaks, two weakened checks, and twelve separate probes were recorded; four mutants survived.
Eight rows passed their own empty-entry probe; nproc=5; evidence is on test-audit/stage1-cohere-lint-harness under this directory.

```json
[
  {
    "test": "TestEmittedJavaScriptMismatch",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/harness_test.go",
    "seconds": 0.024,
    "oracle": "Handwritten shard declaration census and exactly-once assignment; this row now checks construction.",
    "oracle_kind": "self",
    "kills": [
      "H1"
    ],
    "unique_kills": [],
    "last_proven_fail": "H1 shard declarations: got 1 want 2",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestEmittedJavaScriptMismatch"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestEmittedJavaScriptMismatch$; H1.log; stage1/cohere/lint/harness_test.go:12: shard declarations: got 1 want 2"
  },
  {
    "test": "TestEmittedJavaScriptMismatch_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/harness_test.go",
    "seconds": 0.385,
    "oracle": "Successful immutable product construction and nonempty oracle path; no runtime output comparison.",
    "oracle_kind": "self",
    "kills": [
      "H5"
    ],
    "unique_kills": [],
    "last_proven_fail": "H5 emitted mismatch setup did not complete",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestEmittedJavaScriptMismatch_Setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestEmittedJavaScriptMismatch_Setup$; H5.log; stage1/cohere/lint/harness_test.go:19: emitted mismatch setup did not complete"
  },
  {
    "test": "TestCompleteSuggestionSerialization",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/harness_test.go",
    "seconds": 4.51,
    "oracle": "Handwritten exactly-once census and synthetic planted disagreement rejected by completeSuggestionEqual; this row is a union/witness, not a runtime serialization comparison.",
    "oracle_kind": "self",
    "kills": [
      "W1"
    ],
    "unique_kills": [],
    "last_proven_fail": "W1 planted disagreement caught by []",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCompleteSuggestionSerialization"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestCompleteSuggestionSerialization$; W1.log; stage1/cohere/lint/harness_test.go:42: planted disagreement caught by []"
  },
  {
    "test": "TestSuggestionAlongsideAutomaticFix",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/harness_test.go",
    "seconds": 0.027,
    "oracle": "Handwritten AST census: one parallel call and one literal case index per shard; runtime comparisons moved into numbered leaves.",
    "oracle_kind": "self",
    "kills": [
      "H2"
    ],
    "unique_kills": [],
    "last_proven_fail": "H2 shard enumeration changed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestSuggestionAlongsideAutomaticFix$; H2.log; stage1/cohere/lint/harness_test.go:47: shard enumeration changed"
  },
  {
    "test": "TestJsxLintReleaseAndThroughput",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_integration_test.go",
    "seconds": null,
    "oracle": "Go cohere full finding comparison and Go/Node/native count equality; nonzero count and a handwritten JSX census. Both clean attempts timed out before completion.",
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
    "matrix_rows": [],
    "evidence": "baseline-benchmark-split.log: panic: test timed out after 1m30s"
  },
  {
    "test": "TestJsxLintTrees family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_integration_test.go",
    "seconds": 2.803,
    "oracle": "Go TypeScript whole-tree bytes compared with Node running the original parser port source and sanitized native output.",
    "oracle_kind": "external-run",
    "kills": [
      "M2",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3 Node: case 0 line 12: port \"3 JsxText 25 36 0 0 -1 0 0\\t\\t\\\\u000a          \\t\\t0\", Go \"3 JsxText 25 36 0 0 -1 0 0\\t\\t\\\\u000a          \\t\\t1\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestJsxLintTreesSetupIsolation"
    ],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 23.457,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTrees family",
      "TestJsxLintTreesSetupIsolation",
      "TestJsxLintTreesUnion",
      "TestJsxLintTrees_Setup",
      "TestProduct_jsx_oracle family",
      "TestProduct_jsx_tree_lowered",
      "TestProduct_jsx_tree_native",
      "TestProduct_jsx_tree_setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestJsxLintTrees_[0-9]{3}|TestJsxLintTreesUnion|TestJsxLintTrees_Setup|TestProduct_jsx_.*)$; M3.log; stage1/cohere/lint/jsx_integration_test.go:155: Node: case 0 line 12: port \"3 JsxText 25 36 0 0 -1 0 0\\t\\t\\\\u000a          \\t\\t0\", Go \"3 JsxText 25 36 0 0 -1 0 0\\t\\t\\\\u000a          \\t\\t1\"",
    "members": [
      "TestJsxLintTrees_000",
      "TestJsxLintTrees_001",
      "TestJsxLintTrees_002",
      "TestJsxLintTrees_003",
      "TestJsxLintTrees_004",
      "TestJsxLintTrees_005",
      "TestJsxLintTrees_006",
      "TestJsxLintTrees_007",
      "TestJsxLintTrees_008",
      "TestJsxLintTrees_009",
      "TestJsxLintTrees_010",
      "TestJsxLintTrees_011",
      "TestJsxLintTrees_012",
      "TestJsxLintTrees_013",
      "TestJsxLintTrees_014",
      "TestJsxLintTrees_015"
    ],
    "subsumption_mutants": 2
  },
  {
    "test": "TestProduct_jsx_oracle family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_products_test.go",
    "seconds": 0.502,
    "oracle": "Shared jsxGoOracle constructor with two input recipes. Checks command errors, but does not inspect the returned executable; H6 accepted absent oracle files.",
    "oracle_kind": "self",
    "kills": [
      "H7"
    ],
    "unique_kills": [],
    "last_proven_fail": "H7 build jsx-membership: build jsx-membership: exit status 1",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 5,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTrees family",
      "TestJsxLintTreesSetupIsolation",
      "TestJsxLintTreesUnion",
      "TestJsxLintTrees_Setup",
      "TestProduct_jsx_oracle family",
      "TestProduct_jsx_tree_lowered",
      "TestProduct_jsx_tree_native",
      "TestProduct_jsx_tree_setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestProduct_jsx_(membership|parser)$; H7.log; stage1/cohere/lint/jsx_products_test.go:8: build jsx-membership: build jsx-membership: exit status 1",
    "members": [
      "TestProduct_jsx_membership",
      "TestProduct_jsx_parser"
    ]
  },
  {
    "test": "TestProduct_jsx_tree_lowered",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_products_test.go",
    "seconds": 0.489,
    "oracle": "Successful lowering/write return only; H8 wrote empty.c instead of program.c and the row passed. No planted valid construction mutant caught by this row. One construction attempt, limited verdict.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTrees family",
      "TestJsxLintTreesSetupIsolation",
      "TestJsxLintTreesUnion",
      "TestJsxLintTrees_Setup",
      "TestProduct_jsx_oracle family",
      "TestProduct_jsx_tree_lowered",
      "TestProduct_jsx_tree_native",
      "TestProduct_jsx_tree_setup"
    ],
    "evidence": "M1.log: selected row passed; M2.log: selected row passed; M3.log: selected row passed; H8.log: selected row passed"
  },
  {
    "test": "TestProduct_jsx_tree_native",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_products_test.go",
    "seconds": 0.956,
    "oracle": "Successful native build return only; H9 disabled sanitization and passed with __asan_init absent. No valid planted construction mutant caught by this row. One construction attempt, limited verdict.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTrees family",
      "TestJsxLintTreesSetupIsolation",
      "TestJsxLintTreesUnion",
      "TestJsxLintTrees_Setup",
      "TestProduct_jsx_oracle family",
      "TestProduct_jsx_tree_lowered",
      "TestProduct_jsx_tree_native",
      "TestProduct_jsx_tree_setup"
    ],
    "evidence": "M1.log: selected row passed; M2.log: selected row passed; M3.log: selected row passed; H9.log: selected row passed"
  },
  {
    "test": "TestProduct_jsx_tree_setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_products_test.go",
    "seconds": 1.245,
    "oracle": "Shared setup constructor and its nonempty directory guard; no explicit bundle consumption.",
    "oracle_kind": "self",
    "kills": [
      "H4"
    ],
    "unique_kills": [],
    "last_proven_fail": "H4 shared JSX preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTrees family",
      "TestJsxLintTreesSetupIsolation",
      "TestJsxLintTreesUnion",
      "TestJsxLintTrees_Setup",
      "TestProduct_jsx_oracle family",
      "TestProduct_jsx_tree_lowered",
      "TestProduct_jsx_tree_native",
      "TestProduct_jsx_tree_setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestProduct_jsx_tree_setup|TestJsxLintTrees_Setup|TestJsxLintTreesUnion|TestJsxLintTreesSetupIsolation)$; H4.log; stage1/cohere/lint/jsx_products_test.go:28: shared JSX preparation failed"
  },
  {
    "test": "TestJsxLintTreesSetupIsolation",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_setup_isolation_test.go",
    "seconds": 23.457,
    "oracle": "Cold standalone shard must pass and log named build misses. Its child also compares Go/Node/native trees, so observed M2/M3 failures count against uniqueness; H4 proves the construction check.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2",
      "M3",
      "H4"
    ],
    "unique_kills": [],
    "last_proven_fail": "H4 shared JSX preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P5"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTrees family",
      "TestJsxLintTreesSetupIsolation",
      "TestJsxLintTreesUnion",
      "TestJsxLintTrees_Setup",
      "TestProduct_jsx_oracle family",
      "TestProduct_jsx_tree_lowered",
      "TestProduct_jsx_tree_native",
      "TestProduct_jsx_tree_setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestProduct_jsx_tree_setup|TestJsxLintTrees_Setup|TestJsxLintTreesUnion|TestJsxLintTreesSetupIsolation)$; H4.log; stage1/cohere/lint/jsx_integration_test.go:144: shared JSX preparation failed"
  },
  {
    "test": "TestJsxLintTreesShardCoverage",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_shards_test.go",
    "seconds": 0.027,
    "oracle": "Handwritten stable ownership, relocated identity, invalid union and shard-selection invariants.",
    "oracle_kind": "self",
    "kills": [
      "H3"
    ],
    "unique_kills": [],
    "last_proven_fail": "H3 empty corpus accepted",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P10",
      "P11",
      "P12"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTreesShardCoverage"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestJsxLintTreesShardCoverage$; H3.log; stage1/cohere/lint/jsx_shards_test.go:359: empty corpus accepted",
    "vacuous_subcases": [
      "P10/P11: positive ownership stability and relocated identity checks before empty-corpus rejection"
    ]
  },
  {
    "test": "TestJsxLintTrees_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_shards_test.go",
    "seconds": 1.248,
    "oracle": "Shared setup constructor and its nonempty directory guard.",
    "oracle_kind": "self",
    "kills": [
      "H4"
    ],
    "unique_kills": [],
    "last_proven_fail": "H4 shared JSX preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTrees family",
      "TestJsxLintTreesSetupIsolation",
      "TestJsxLintTreesUnion",
      "TestJsxLintTrees_Setup",
      "TestProduct_jsx_oracle family",
      "TestProduct_jsx_tree_lowered",
      "TestProduct_jsx_tree_native",
      "TestProduct_jsx_tree_setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestProduct_jsx_tree_setup|TestJsxLintTrees_Setup|TestJsxLintTreesUnion|TestJsxLintTreesSetupIsolation)$; H4.log; stage1/cohere/lint/jsx_shards_test.go:511: shared JSX preparation failed"
  },
  {
    "test": "TestJsxLintTreesUnion",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_shards_test.go",
    "seconds": 1.244,
    "oracle": "AST declaration census and exact live bundle union. Kept separate because it asserts declaration and coverage facts absent from runtime shards.",
    "oracle_kind": "self",
    "kills": [
      "H4"
    ],
    "unique_kills": [],
    "last_proven_fail": "H4 shared JSX preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P6"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTrees family",
      "TestJsxLintTreesSetupIsolation",
      "TestJsxLintTreesUnion",
      "TestJsxLintTrees_Setup",
      "TestProduct_jsx_oracle family",
      "TestProduct_jsx_tree_lowered",
      "TestProduct_jsx_tree_native",
      "TestProduct_jsx_tree_setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestProduct_jsx_tree_setup|TestJsxLintTrees_Setup|TestJsxLintTreesUnion|TestJsxLintTreesSetupIsolation)$; H4.log; stage1/cohere/lint/jsx_shards_test.go:668: shared JSX preparation failed"
  },
  {
    "test": "TestJsxLintTreesShardDisagreement",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/jsx_shards_test.go",
    "seconds": 0.049,
    "oracle": "Synthetic planted tree disagreement must fail only its owning child shard and include the planted message.",
    "oracle_kind": "self",
    "kills": [
      "W2"
    ],
    "unique_kills": [],
    "last_proven_fail": "W2 planted disagreement survived:",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestJsxLintTreesShardDisagreement"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestJsxLintTreesShardDisagreement$; W2.log; stage1/cohere/lint/jsx_shards_test.go:744: planted disagreement survived:"
  }
]
```

Mutants, with every site against the starting commit:

| ID | Kind | Origin file:line | Change | Failed grouped rows |
| --- | --- | --- | --- | --- |
| M1 | production | stage1/typescript/parser/nodes.ts:31 | off-by-one printable ASCII upper bound 126 -> 125 |  |
| M2 | production | stage1/typescript/parser/nodes.ts:51 | change recursive tree depth increment 1 -> 2 | TestJsxLintTrees family, TestJsxLintTreesSetupIsolation |
| M3 | production | stage1/typescript/parser/jsx.ts:69 | flip JSX text whitespace semantics | TestJsxLintTrees family, TestJsxLintTreesSetupIsolation |
| H1 | setup | stage1/cohere/lint/emitted_javascript_shards_test.go:30 | change declared shard count 1 -> 2 | TestEmittedJavaScriptMismatch |
| H2 | setup | stage1/cohere/lint/suggestion_alongside_shards_test.go:31 | change declared shard count 3 -> 4 | TestSuggestionAlongsideAutomaticFix |
| W1 | witness | stage1/cohere/lint/complete_suggestion_serialization_shards_test.go:66 | return early with unconditional agreement | TestCompleteSuggestionSerialization |
| W2 | witness | stage1/cohere/lint/jsx_shards_test.go:317 | drop the complete tree disagreement assertion | TestJsxLintTreesShardDisagreement |
| H3 | setup | stage1/cohere/lint/jsx_shards_test.go:253 | flip empty-corpus rejection condition to impossible bound | TestJsxLintTreesShardCoverage |
| H4 | setup | stage1/cohere/lint/jsx_shards_test.go:516 | drop the complete shared preparation statement | TestJsxLintTreesSetupIsolation, TestJsxLintTreesUnion, TestJsxLintTrees_Setup, TestProduct_jsx_tree_setup |
| H5 | setup | stage1/cohere/lint/emitted_javascript_shards_test.go:39 | drop the complete mismatch product preparation statement | TestEmittedJavaScriptMismatch_Setup |
| H6 | setup | stage1/cohere/lint/jsx_shards_test.go:98 | drop the oracle construction command, preserving oracle source |  |
| H7 | setup | stage1/cohere/lint/jsx_shards_test.go:95 | change the build working-directory option from cohere root to scratch output | TestProduct_jsx_oracle family |
| H8 | setup | stage1/cohere/lint/jsx_shards_test.go:445 | change the construction output filename program.c -> empty.c |  |
| H9 | setup | stage1/cohere/lint/jsx_shards_test.go:462 | flip the sanitizer construction option |  |

Survivors:

- M1: Node directly evaluating written("~") prints `~` before and `\u007e` after. See M1-before.log and M1-after.log. This is changed behavior unguarded by the bounded matrix; outside rows are unknown.
- H6: both product tests pass without invoking Go build. The reference construction has oracle executables; both H6 products contain only overlay.json and lack oracle. See artifact-witnesses.json.
- H8: the lowered product row passes with empty.c present and program.c absent. See artifact-witnesses.json. This is a construction gap, not a production survivor.
- H9: the native product row passes after flipping sanitize=true to false. `nm` exits zero for both reference and mutant binaries; __asan_init is present in M1-nm.log and absent in H9-nm.log. The reference uses the original sanitizer construction, although its independent ASCII-bound mutant was active. This proves the sanitizer option changed, without claiming byte-for-byte clean-binary identity.

There are no equivalent-candidate survivors. Probes do not count as mutants, kills for worthiness, uniqueness, or subsumption. H and W mutations apply only to construction or witnessed comparisons. Neither Go cohere nor Go TypeScript oracle source was mutated.

Scope and problems with the brief:

- The brief says 15 rows and lists 31 functions. Sixteen JSX tree wrappers share one checker. The two Go-oracle product wrappers also share jsxGoOracle with different input recipes. Those groupings produce 15 rows. All names remain present. The other product recipes differ, so they remain separate.
- Names have stale implications. TestEmittedJavaScriptMismatch is now a shard-census check; TestSuggestionAlongsideAutomaticFix is now an AST-census check. Their runtime checks moved to numbered leaves outside this named scope. TestCompleteSuggestionSerialization is now a union and planted-disagreement witness, after preparation. The delegate files are recorded in go-helper-inventory.txt.
- The union test asserts declaration and exact coverage facts absent from runtime shards. It remains its own construction row under the explicit distinct-assertions exception to family grouping.
- The whole package contains thousands of compiler agreement leaves and cannot fit the 90-second budget. Its first run stopped in the serial opted-in benchmark, before releasing paused rows. A narrowed run also cooked during build preparation. Those are timeouts, not assertion-red baselines. Split compilation made both remaining preparation consumers pass, and all bounded rows then passed alone and together.
- The benchmark still cooked alone with warm dependencies and ADAMIC_NATIVE_SPLIT=1. Its median and quality verdict are unknown. No second or third timing was attempted after it cooked. Its source compares full finding output before checking count equality, but no completed run supports an empirical oracle-strength verdict.
- Product wrappers can return successfully with no product at all. P2/P3/P4/P5/P7/P8/P9 show empty-entry acceptance. Construction-error catches still justify setup-check where demonstrated; empty acceptance is a separate vacuity finding. The lowered and native product untrue verdicts rest on one valid construction break each, not an exhaustive search.
- Isolation is mixed: its purpose is cold preparation, but its child performs an actual parser agreement check. Its observed production failures are included in the matrix and remove uniqueness from the JSX family. Its setup-check verdict separately rests on H4. The family is subsumed only on M2 and M3; this is a two-mutant hint, not a deletion recommendation. The family is faster.
- The production matrix includes the named parser consumers and isolation. It cannot establish package or repository uniqueness. Other lint-port consumers also import written and parser functions; their results are unknown and require central replay. Construction and witness runs are bounded to the check they exercise, so their one-row failures do not establish package uniqueness either.
- Exact dynamic function reach was not established. Conservative TypeScript declaration and Go delegate inventories are supplied, including functions that may not execute on these fixtures. This does not fulfill an exact all-reached-functions proof; the tested mutation sites and their actual failures are demonstrated.
- Port mutations used the allowed per-mutant rebuild alternative, with only three distinct production mutations and a separate entry probe. Go construction edits were also compiled independently instead of using one switch. This added vet/build overhead; no stale products were reused across mutation caches.
- The first H9 prototype did not compile because its changed input left data unused. It was rejected before any matrix run and replaced with the sanitizer-option flip. H9-rejected-vet.log and H9-rejected-prototype.txt document it; the prototype is excluded from replay diffs and verdicts. Removing probe bodies also required removing newly unused imports. P3-initial-vet.log records that cost.
- Drop-statement mutations H4 and H5 drop the whole preparation statement, including its closure. W2 drops the whole assertion block. Runtime failure line numbers shift under those edits; matrix.json maps them back to origin/main. Logs preserve the literal runtime output.
- P12 panicked on a nil shard-selection function. It was run alone and rerun alone in P12-recheck.log. No later code in that row is claimed to have executed. Coverage's positive ownership and relocation checks survive P10/P11 before the negative empty-corpus check fails, recorded as vacuous_subcases.
- The package's cold-isolation child has a 600-second timeout, while this audit gives its parent only 90 seconds. It passed alone three times under the audit budget. Its initial parallel baseline took 63.52 seconds, but its alone median was 23.457 seconds, so that first shared-run time is not used for the cost verdict.
- Warm tools did not remove npm preparation. npm ci was run in stage3/api. The scoped Node entry uses oracle/node.mjs and Node's built-in type stripping; no additional scoped node_modules dependency was found. All requested rows were attempted with their applicable opt-in; none skipped in the completed bounded runs. The benchmark was enabled and cooked.

Timing and uncovered work:

Toolchain setup was skipped because env.sh worked. npm ci took 0.475 seconds; registry generation took 0.321 seconds; listing took 6.214 seconds. nproc was 5. Three-alone timing commands totaled 190.928 wall seconds. Family costs use one binary invocation for all members, not a sum of per-member rounded PASS times. seconds in rows.json are medians of the test binary's package line, with ADAMIC_NATIVE_SPLIT=1 and warm persistent products. Cold costs are visible in baseline logs.

Whole baseline: 91.847 command wall seconds and 90.186 binary seconds. First bounded baseline: 91.692 wall seconds, cooked. Split preparation retry: 81.322 wall seconds and 79.628 binary seconds, passed. Benchmark alone: 91.675 wall seconds, cooked. Final restored bounded baseline: 29.301 wall seconds and 27.645 package-line seconds, passed.

Initial valid mutation runs totaled 181.523 wall seconds; H9/P1 took 42.866; entry probes P2-P12 took 77.916; the additional isolation replays took 75.156. These overlap neither production timings nor each other. Rejected compile attempts and reading/reporting time are additional. Overall unit work was approximately 25 minutes, within the stage1-port allowance rather than the 20-minute ordinary-unit target.

Native rebuild timing for every production mutation and the port probe is recorded below. Native compiler execution is demonstrated by successful product builds before comparisons. Separate caches are /tmp/u109/cache/<id>; isolation children use their own fresh TempDir caches. The complete commands are in run-results-first.json, run-results.json, isolation-results.json, probe-results.json, timings.json, and matrix.json.

- M1: jsx_shards_test.go:524: build jsx-tree-lowered 79cdb12e1d16 miss 5.50

- M1: jsx_shards_test.go:524: build jsx-tree-native af23f502335f miss 1.39

- M2: jsx_products_test.go:18: build jsx-tree-lowered 9fb437427f9c miss 6.37

- M2: jsx_shards_test.go:524: build jsx-tree-native 60053f2d2b80 miss 1.35

- M3: jsx_shards_test.go:524: build jsx-tree-lowered 90d5856ea6a9 miss 5.59

- M3: jsx_shards_test.go:524: build jsx-tree-native eecce87d4168 miss 1.27

- P1: jsx_products_test.go:18: build jsx-tree-lowered f8db97af2ee4 miss 1.70

- P1: jsx_shards_test.go:524: build jsx-tree-native 8928bb07c2f5 miss 2.86

Uncovered: completed benchmark execution and three-run median; other package rows and repo-wide replay; an exact dynamic TypeScript reach proof; stronger semantic construction mutants for product-only rows. No tests or production fixes were committed. Only audit evidence is intended for push. No pull request or main push was opened.
