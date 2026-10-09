u155: origin/main ef819b8e03c26ee3c5ac7bc1d8f067b77d553640; nproc 5.
All 31 requested Tests exist; family grouping yields two rows.
Verdicts: JSX witness family and product setup-check family.
Both rows pass every member's own empty-entry probe and are vacuous.
Evidence: review/test-audit/stage1-typescript-parser-jsx_mutants_products/.

```json
[
  {
    "test": "TestJsxMutants family",
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/jsx_mutants_products_test.go",
    "seconds": 27.105,
    "oracle": "Executed typescript-go AST bytes compared with an unchanged native control; deliberate Node/native mutant output must differ. Self-written child proof requires exactly owner 004 to reject its planted Node survivor; self-written union checks coverage.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: jsx_mutants_products_test.go:217: owner failed to catch planted survivor: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestJsxMutants family",
      "TestProduct_JsxMutants family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u155/evidence/W1.overlay.json -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestJsxMutants_000|TestJsxMutants_001|TestJsxMutants_002|TestJsxMutants_003|TestJsxMutants_004|TestJsxMutants_005|TestJsxMutants_006|TestJsxMutants_007|TestJsxMutants_008|TestJsxMutantsUnion|TestProduct_JsxMutantsOracle|TestProduct_JsxMutantsLower_Control|TestProduct_JsxMutantsNative_Control|TestProduct_JsxMutantsLower_000|TestProduct_JsxMutantsNative_000|TestProduct_JsxMutantsLower_001|TestProduct_JsxMutantsNative_001|TestProduct_JsxMutantsLower_002|TestProduct_JsxMutantsNative_002|TestProduct_JsxMutantsLower_003|TestProduct_JsxMutantsNative_003|TestProduct_JsxMutantsLower_004|TestProduct_JsxMutantsNative_004|TestProduct_JsxMutantsLower_005|TestProduct_JsxMutantsNative_005|TestProduct_JsxMutantsLower_006|TestProduct_JsxMutantsNative_006|TestProduct_JsxMutantsLower_007|TestProduct_JsxMutantsNative_007|TestProduct_JsxMutantsLower_008|TestProduct_JsxMutantsNative_008)$' > W1.log 2>&1; jsx_mutants_products_test.go:217: owner failed to catch planted survivor: <nil>",
    "members": [
      "TestJsxMutants_000",
      "TestJsxMutants_001",
      "TestJsxMutants_002",
      "TestJsxMutants_003",
      "TestJsxMutants_004",
      "TestJsxMutants_005",
      "TestJsxMutants_006",
      "TestJsxMutants_007",
      "TestJsxMutants_008",
      "TestJsxMutantsUnion"
    ],
    "timing_samples": [
      27.221,
      27.105,
      27.1
    ],
    "own_probes": [
      "P1",
      "P2"
    ],
    "probe_entry_map": "probe-entry-map.json",
    "raw_failing_line": "jsx_mutants_products_test.go:217: owner failed to catch planted survivor: <nil>",
    "construction_kills": [
      "S2"
    ]
  },
  {
    "test": "TestProduct_JsxMutants family",
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/jsx_mutants_products_test.go",
    "seconds": 8.204,
    "oracle": "Self-written build recipes and successful product preparation. Go builds the oracle, and lowering/native compilation build control and mutant products; wrappers ignore returned paths and do not compare parser answers.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: jsx_mutants_products_test.go:405: open /tmp/u155/cache/S1/9d9743cc642e357826f14e756dc5184d4d607a8592e9f9b547049a744398f58b/program.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestJsxMutants family",
      "TestProduct_JsxMutants family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u155/evidence/S1.overlay.json -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestJsxMutants_000|TestJsxMutants_001|TestJsxMutants_002|TestJsxMutants_003|TestJsxMutants_004|TestJsxMutants_005|TestJsxMutants_006|TestJsxMutants_007|TestJsxMutants_008|TestJsxMutantsUnion|TestProduct_JsxMutantsOracle|TestProduct_JsxMutantsLower_Control|TestProduct_JsxMutantsNative_Control|TestProduct_JsxMutantsLower_000|TestProduct_JsxMutantsNative_000|TestProduct_JsxMutantsLower_001|TestProduct_JsxMutantsNative_001|TestProduct_JsxMutantsLower_002|TestProduct_JsxMutantsNative_002|TestProduct_JsxMutantsLower_003|TestProduct_JsxMutantsNative_003|TestProduct_JsxMutantsLower_004|TestProduct_JsxMutantsNative_004|TestProduct_JsxMutantsLower_005|TestProduct_JsxMutantsNative_005|TestProduct_JsxMutantsLower_006|TestProduct_JsxMutantsNative_006|TestProduct_JsxMutantsLower_007|TestProduct_JsxMutantsNative_007|TestProduct_JsxMutantsLower_008|TestProduct_JsxMutantsNative_008)$' > S1.log 2>&1; jsx_mutants_products_test.go:405: open /tmp/u155/cache/S1/9d9743cc642e357826f14e756dc5184d4d607a8592e9f9b547049a744398f58b/program.c: no such file or directory",
    "members": [
      "TestProduct_JsxMutantsOracle",
      "TestProduct_JsxMutantsLower_Control",
      "TestProduct_JsxMutantsNative_Control",
      "TestProduct_JsxMutantsLower_000",
      "TestProduct_JsxMutantsNative_000",
      "TestProduct_JsxMutantsLower_001",
      "TestProduct_JsxMutantsNative_001",
      "TestProduct_JsxMutantsLower_002",
      "TestProduct_JsxMutantsNative_002",
      "TestProduct_JsxMutantsLower_003",
      "TestProduct_JsxMutantsNative_003",
      "TestProduct_JsxMutantsLower_004",
      "TestProduct_JsxMutantsNative_004",
      "TestProduct_JsxMutantsLower_005",
      "TestProduct_JsxMutantsNative_005",
      "TestProduct_JsxMutantsLower_006",
      "TestProduct_JsxMutantsNative_006",
      "TestProduct_JsxMutantsLower_007",
      "TestProduct_JsxMutantsNative_007",
      "TestProduct_JsxMutantsLower_008",
      "TestProduct_JsxMutantsNative_008"
    ],
    "timing_samples": [
      8.204,
      8.127,
      8.583
    ],
    "own_probes": [
      "P3",
      "P4",
      "P5"
    ],
    "probe_entry_map": "probe-entry-map.json",
    "raw_failing_line": "jsx_mutants_products_test.go:405: open /tmp/u155/cache/S1/9d9743cc642e357826f14e756dc5184d4d607a8592e9f9b547049a744398f58b/program.c: no such file or directory",
    "construction_kills": [
      "S1"
    ]
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| W1 | stage1/typescript/parser/jsx_mutants_products_test.go:202 | Disable survivor comparison | TestJsxMutants family |
| S1 | stage1/typescript/parser/jsx_mutants_products_test.go:83 | Rename required generated C artifact | TestJsxMutants family, TestProduct_JsxMutants family |
| S2 | stage1/typescript/parser/jsx_mutants_products_test.go:132 | Drop whole seen[id] assignment | TestJsxMutants family |
| P1 | stage1/typescript/parser/jsx_mutants_products_test.go:164 | Return at runner entry | none |
| P2 | stage1/typescript/parser/jsx_mutants_products_test.go:114 | Return at union entry | none |
| P3 | stage1/typescript/parser/jsx_mutants_products_test.go:27 | Return empty oracle path | TestJsxMutants family |
| P4 | stage1/typescript/parser/jsx_mutants_products_test.go:61 | Return empty lower path | TestJsxMutants family, TestProduct_JsxMutants family |
| P5 | stage1/typescript/parser/jsx_mutants_products_test.go:88 | Return empty native path | TestJsxMutants family |

Survivors: no production mutants were planted. All three check/construction edits are caught. Empty-entry survivors are probe findings, not equivalent production candidates.

The brief calls this 22 rows but supplies 31 Test functions. All 31 exist in the named file at current origin/main ef819b8e03c26ee3c5ac7bc1d8f067b77d553640; none moved or vanished. Applying the family rule produces two rows. The nine execution wrappers differ only by shard input to jsxMutantsRun; their coverage union joins that family. All 21 product wrappers reach jsxMutantsFetch and buildcache.Product with different recipes, covering an oracle, a control and nine mutants at lower/native levels. They assert successful construction and no parser answer. Full members and individual matrix outcomes are retained, so central review can split the family differently without losing observations.

The whole clean package timed out at the binary's 90.020-second line. No assertion failure was recorded before the timeout; later rows are unknown. The entire requested 31-test slice then passed in 64.336 seconds. All matrix runs use that same slice with no shard-selection variable and no skipped members. The initial baseline lacked the optional pinned TypeScript compiler corpus. I fetched its exact 050880ce59e30b356b686bd3144efe24f875ebc8 commit, then enabled ADAMIC_TYPESCRIPT_SOURCE for the timing and matrix runs. This slice does not call compilerManifest, so that corpus does not affect its oracle. The initial skip list is in baseline-skips.json; out-of-slice optional tests were not rerun because the whole package was already over budget.

This is a witness/product-construction unit. Production printer/parser defects cannot decide whether the witnesses detect a disabled agreement check. I used the brief's allowed weakened comparison and construction edits, and kept their failures out of production kills. mutants_in_matrix is zero. The nine port mutants the suite builds are pre-existing test inputs, not audit mutants chosen here. Their successful native construction and execution are baseline evidence; they do not establish production mutation coverage or uniqueness for this audit.

W1 changes the survivor condition to false && bytes.Equal. Each normal deliberate mutant still executes successfully, but its survivor guard cannot fail. The child proof plants the oracle answer as owner 004's Node output. That child now passes, and its parent reports owner failed to catch planted survivor. Exactly TestJsxMutants_004 fails; the other eight leaves pass. This is direct weakened-check evidence for the grouped witness verdict. The control parser comparison and external typescript-go oracle were left intact. A broken native/control precondition would not have been counted as witness proof.

The normal mutant check accepts any aggregate AST-output difference from typescript-go. It does not demand the intended changed node, nor prove unaffected inputs agree. The unaltered control comparison is stronger: every AST byte must match typescript-go. Commands require successful execution and no stderr. The planted child overrides Node output only; no separately planted native survivor was tested. The child proof additionally requires a specific error string, failed owner leaf and successful nonowner leaves; it therefore checks ownership rather than merely counting failures. The union supplies self-written enumeration and coverage checks, not an outside semantic oracle.

S1 misnames generated C as missing-program.c. Lowering still constructs its product, but native preparation cannot read program.c. It catches the product family's native members; other members remain green. A fresh ADAMIC_BUILD_CACHE_DIR isolates this construction defect from successful cached artifacts. S2 drops the seen[id] assignment from the union; the coverage member reports zero seen cases versus its complete enumeration. Both are suite-construction edits permitted by the brief, not production changes. Their member-level failures are retained instead of pretending every member of a family failed.

Empty-entry probes have different owners. P1 returns from jsxMutantsRun and P2 returns from jsxMutantsUnion; all members of their family pass their own entry probe, because P1 also skips its child proof. P3 returns an empty oracle path, P4 an empty lower-product path, and P5 an empty native path. The corresponding product wrappers ignore these returns and pass. P4 does fail native wrappers that read the absent lower-product prerequisite, but those wrappers' own entry is the native builder, and their own P5 passes. Counting P4 against native-wrapper vacuity would violate the own-entry rule. probe-entry-map.json records the mapping. Both grouped rows are vacuous under their own entries despite their demonstrated witness/setup verdicts.

The product family has several entry points, so vacuity cannot be inferred from a single whole-row probe run. I evaluated each constituent against its own entry and aggregated those observations. All constituents passed their own entry probes. No probe is used to award witness or setup-check. P1/P4/P5 remove imports made unused by their dropped function bodies, so each standalone diff compiles; those removals are documented in catalog.json. Private Go overlays avoid modifying tracked source, and origin line numbers are mapped from unchanged diagnostic lines while raw logs are retained.

The fixed production-mutant menu and separate native-rebuild cap are not applicable to these allowed harness/check edits. No selector switch was added: private overlays compile each edit, and a dedicated cache handles the one construction edit that must force a fresh build. Build product miss durations can overlap and do not measure clang alone. Child proof output is captured inside a parent's log, so the matrix counts only actual top-level JSON fail actions; a failed proof child expected by a passing parent is not a matrix kill.

Evidence was copied after all tests. It includes the list, full family membership, exact commands and environments, raw outputs, probe owners, standalone diffs and replay instructions. No unrelated packages were run. Full transitive compiler/runtime function coverage was not measured; the local check/construction functions and their preparation helpers were read. No production parser mutation coverage, repository-wide uniqueness or out-of-slice verdict is claimed. Warm env.sh avoided setup; npm ci in stage3/api still ran before baseline.

Setup skipped because env.sh worked. API npm ci reported 348 ms. Whole baseline binary timed out at 90.020 s; slice baseline passed in 64.336 s. Both rows ran alone three times. Total elapsed approximately eleven minutes. Corpus fetch wall timing was not recorded. Subsequent command wall total: 281.503 s. Per-product rebuild times are in builds.json.

bounded-baseline: binary 64.336 s; sum of logged misses 174.53 s.

W1: binary 26.816 s; sum of logged misses 0 s.

S1: binary 22.28 s; sum of logged misses 71.73 s.

Not covered: tests outside the requested slice, completed optional whole-package runs, production parser mutation coverage, full transitive function coverage and repository-wide uniqueness. All eight standalone diffs apply to the starting source and all Go overlays passed go vet. Production source is unchanged.
