Unit u157 at 23a19b8b48662dfb6a19baa00b8efbb52ab5275e.
All 15 requested names exist, grouped into nine rows.
Whole baseline exceeded 90 s; narrowed baseline passed in 53.334 binary seconds.
Four production mutants, six separate construction faults, eight empty-entry probes.
Source restored; evidence branch test-audit/stage1-typescript-parser-setup_deadline_products.

```json
[
  {
    "test": "TestProduct_ParserOracle",
    "members": [
      "TestProduct_ParserOracle"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/setup_deadline_products_test.go",
    "seconds": 0.426,
    "oracle": "self: construction must succeed; wrapper discards returned artifact path and checks no parser answers",
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
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [],
    "evidence": "S1 completed without failure in this row; P3 empty builder also passed. See S1.log and P3.log."
  },
  {
    "test": "TestProduct_CompilerExpressionsLower",
    "members": [
      "TestProduct_CompilerExpressionsLower"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/setup_deadline_products_test.go",
    "seconds": 1.881,
    "oracle": "self: construction must succeed; wrapper discards returned artifact path and checks no parser answers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S2: setup_deadline_products_test.go:33: build compiler-expressions-lowered: open /tmp/u157/cache/S2/.building-91819c136215-2941223516/absent/main.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [
      "S2"
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u157-typescript ADAMIC_BUILD_CACHE_DIR=/tmp/u157/cache/S2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestProduct_ParserOracle|TestProduct_CompilerExpressionsLower|TestProduct_CompilerExpressionsNative|TestProduct_WholeMutantsLower_Control|TestProduct_WholeMutantsLower_000|TestProduct_WholeMutantsLower_001|TestProduct_WholeMutantsLower_002|TestProduct_WholeMutantsNative_Control|TestProduct_WholeMutantsNative_000|TestProduct_WholeMutantsNative_001|TestProduct_WholeMutantsNative_002|TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees|TestWholeMutants_Setup)$' > review/test-audit/stage1-typescript-parser-setup_deadline_products/S2.log 2>&1; setup_deadline_products_test.go:33: build compiler-expressions-lowered: open /tmp/u157/cache/S2/.building-91819c136215-2941223516/absent/main.c: no such file or directory"
  },
  {
    "test": "TestProduct_CompilerExpressionsNative",
    "members": [
      "TestProduct_CompilerExpressionsNative"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/setup_deadline_products_test.go",
    "seconds": 3.904,
    "oracle": "self: construction must succeed; wrapper discards returned artifact path and checks no parser answers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S3: setup_deadline_products_test.go:37: build compiler-expressions-native: native: clang failed: exit status 1",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [
      "S2",
      "S3"
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u157-typescript ADAMIC_BUILD_CACHE_DIR=/tmp/u157/cache/S3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestProduct_ParserOracle|TestProduct_CompilerExpressionsLower|TestProduct_CompilerExpressionsNative|TestProduct_WholeMutantsLower_Control|TestProduct_WholeMutantsLower_000|TestProduct_WholeMutantsLower_001|TestProduct_WholeMutantsLower_002|TestProduct_WholeMutantsNative_Control|TestProduct_WholeMutantsNative_000|TestProduct_WholeMutantsNative_001|TestProduct_WholeMutantsNative_002|TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees|TestWholeMutants_Setup)$' > review/test-audit/stage1-typescript-parser-setup_deadline_products/S3.log 2>&1; setup_deadline_products_test.go:37: build compiler-expressions-native: native: clang failed: exit status 1"
  },
  {
    "test": "TestProduct_WholeMutantsLower family",
    "members": [
      "TestProduct_WholeMutantsLower_Control",
      "TestProduct_WholeMutantsLower_000",
      "TestProduct_WholeMutantsLower_001",
      "TestProduct_WholeMutantsLower_002"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/setup_deadline_products_test.go",
    "seconds": 0.908,
    "oracle": "self: construction must succeed; wrapper discards returned artifact path and checks no parser answers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S4: setup_deadline_products_test.go:62: build typescript-parser-lowered: open /tmp/u157/cache/S4/.building-f9752a3aaf18-785119383/absent/main.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [
      "S4"
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u157-typescript ADAMIC_BUILD_CACHE_DIR=/tmp/u157/cache/S4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestProduct_ParserOracle|TestProduct_CompilerExpressionsLower|TestProduct_CompilerExpressionsNative|TestProduct_WholeMutantsLower_Control|TestProduct_WholeMutantsLower_000|TestProduct_WholeMutantsLower_001|TestProduct_WholeMutantsLower_002|TestProduct_WholeMutantsNative_Control|TestProduct_WholeMutantsNative_000|TestProduct_WholeMutantsNative_001|TestProduct_WholeMutantsNative_002|TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees|TestWholeMutants_Setup)$' > review/test-audit/stage1-typescript-parser-setup_deadline_products/S4.log 2>&1; setup_deadline_products_test.go:62: build typescript-parser-lowered: open /tmp/u157/cache/S4/.building-f9752a3aaf18-785119383/absent/main.c: no such file or directory"
  },
  {
    "test": "TestProduct_WholeMutantsNative family",
    "members": [
      "TestProduct_WholeMutantsNative_Control",
      "TestProduct_WholeMutantsNative_000",
      "TestProduct_WholeMutantsNative_001",
      "TestProduct_WholeMutantsNative_002"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/setup_deadline_products_test.go",
    "seconds": 1.905,
    "oracle": "self: construction must succeed; wrapper discards returned artifact path and checks no parser answers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S5: setup_deadline_products_test.go:47: build typescript-parser-native: native: clang failed: exit status 1",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [
      "S4",
      "S5"
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u157-typescript ADAMIC_BUILD_CACHE_DIR=/tmp/u157/cache/S5 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestProduct_ParserOracle|TestProduct_CompilerExpressionsLower|TestProduct_CompilerExpressionsNative|TestProduct_WholeMutantsLower_Control|TestProduct_WholeMutantsLower_000|TestProduct_WholeMutantsLower_001|TestProduct_WholeMutantsLower_002|TestProduct_WholeMutantsNative_Control|TestProduct_WholeMutantsNative_000|TestProduct_WholeMutantsNative_001|TestProduct_WholeMutantsNative_002|TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees|TestWholeMutants_Setup)$' > review/test-audit/stage1-typescript-parser-setup_deadline_products/S5.log 2>&1; setup_deadline_products_test.go:47: build typescript-parser-native: native: clang failed: exit status 1"
  },
  {
    "test": "TestEveryTypeNodeKindAgrees",
    "members": [
      "TestEveryTypeNodeKindAgrees"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/type_kinds_test.go",
    "seconds": 20.211,
    "oracle": "Runs unmodified typescript-go via a Go overlay; compares full tree bytes from the port on Node and native. Type-kind inventory is also from typescript-go.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M2: type_kinds_test.go:27: Node doc types: case 0 line 3: port \"0 JSDocNullableType 11 15 0 0 -1 0 0\\t\\t\\t\\t\", Go \"0 JSDocVariadicType 11 15 0 0 -1 0 0\\t\\t\\t\\t\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u157-typescript ADAMIC_BUILD_CACHE_DIR=/tmp/u157/cache/M2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestProduct_ParserOracle|TestProduct_CompilerExpressionsLower|TestProduct_CompilerExpressionsNative|TestProduct_WholeMutantsLower_Control|TestProduct_WholeMutantsLower_000|TestProduct_WholeMutantsLower_001|TestProduct_WholeMutantsLower_002|TestProduct_WholeMutantsNative_Control|TestProduct_WholeMutantsNative_000|TestProduct_WholeMutantsNative_001|TestProduct_WholeMutantsNative_002|TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees|TestWholeMutants_Setup)$' > review/test-audit/stage1-typescript-parser-setup_deadline_products/M2.log 2>&1; type_kinds_test.go:27: Node doc types: case 0 line 3: port \"0 JSDocNullableType 11 15 0 0 -1 0 0\\t\\t\\t\\t\", Go \"0 JSDocVariadicType 11 15 0 0 -1 0 0\\t\\t\\t\\t\""
  },
  {
    "test": "TestWholeGeneratedAgrees",
    "members": [
      "TestWholeGeneratedAgrees"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/whole_generated_test.go",
    "seconds": 19.392,
    "oracle": "Runs unmodified typescript-go and compares full tree bytes with the port on Node and native.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: whole_generated_test.go:66: Node: case 0 line 3: port \"0 SourceFile 0 138 32 0 -1 0 0\\t\\t\\t\\t\", Go \"0 SourceFile 0 138 0 0 -1 0 0\\t\\t\\t\\t\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestEveryTypeNodeKindAgrees"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 20.211,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u157-typescript ADAMIC_BUILD_CACHE_DIR=/tmp/u157/cache/M1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestProduct_ParserOracle|TestProduct_CompilerExpressionsLower|TestProduct_CompilerExpressionsNative|TestProduct_WholeMutantsLower_Control|TestProduct_WholeMutantsLower_000|TestProduct_WholeMutantsLower_001|TestProduct_WholeMutantsLower_002|TestProduct_WholeMutantsNative_Control|TestProduct_WholeMutantsNative_000|TestProduct_WholeMutantsNative_001|TestProduct_WholeMutantsNative_002|TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees|TestWholeMutants_Setup)$' > review/test-audit/stage1-typescript-parser-setup_deadline_products/M1.log 2>&1; whole_generated_test.go:66: Node: case 0 line 3: port \"0 SourceFile 0 138 32 0 -1 0 0\\t\\t\\t\\t\", Go \"0 SourceFile 0 138 0 0 -1 0 0\\t\\t\\t\\t\""
  },
  {
    "test": "TestObsoleteImportAttributesAgrees",
    "members": [
      "TestObsoleteImportAttributesAgrees"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/whole_generated_test.go",
    "seconds": 20.991,
    "oracle": "Runs unmodified typescript-go and compares full tree bytes with the port on Node and native; obsolete case additionally requires Go diagnostics to be nonempty and all code 2880.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M3"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M3: whole_generated_test.go:89: Node: case 0 line 8: port \"2 ImportAttributes 17 39 0 0 1 0 0\\tWithKeyword\\t\\t\\t\", Go \"2 ImportAttributes 17 39 0 0 1 0 0\\tAssertKeyword\\t\\t\\t\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u157-typescript ADAMIC_BUILD_CACHE_DIR=/tmp/u157/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestProduct_ParserOracle|TestProduct_CompilerExpressionsLower|TestProduct_CompilerExpressionsNative|TestProduct_WholeMutantsLower_Control|TestProduct_WholeMutantsLower_000|TestProduct_WholeMutantsLower_001|TestProduct_WholeMutantsLower_002|TestProduct_WholeMutantsNative_Control|TestProduct_WholeMutantsNative_000|TestProduct_WholeMutantsNative_001|TestProduct_WholeMutantsNative_002|TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees|TestWholeMutants_Setup)$' > review/test-audit/stage1-typescript-parser-setup_deadline_products/M3.log 2>&1; whole_generated_test.go:89: Node: case 0 line 8: port \"2 ImportAttributes 17 39 0 0 1 0 0\\tWithKeyword\\t\\t\\t\", Go \"2 ImportAttributes 17 39 0 0 1 0 0\\tAssertKeyword\\t\\t\\t\""
  },
  {
    "test": "TestWholeMutants_Setup",
    "members": [
      "TestWholeMutants_Setup"
    ],
    "package": "stage1/typescript/parser",
    "file": "stage1/typescript/parser/whole_mutants_setup_test.go",
    "seconds": 2.185,
    "oracle": "self: prepared map must contain four binaries and every returned path must exist; no parser-answer comparison",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S6: whole_mutants_setup_test.go:75: shared setup must finish before a whole-mutant leaf runs",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_ParserOracle",
      "TestProduct_CompilerExpressionsLower",
      "TestProduct_CompilerExpressionsNative",
      "TestProduct_WholeMutantsLower family",
      "TestProduct_WholeMutantsNative family",
      "TestEveryTypeNodeKindAgrees",
      "TestWholeGeneratedAgrees",
      "TestObsoleteImportAttributesAgrees",
      "TestWholeMutants_Setup"
    ],
    "setup_kills": [
      "S1",
      "S4",
      "S5",
      "S6"
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u157-typescript ADAMIC_BUILD_CACHE_DIR=/tmp/u157/cache/S6 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestProduct_ParserOracle|TestProduct_CompilerExpressionsLower|TestProduct_CompilerExpressionsNative|TestProduct_WholeMutantsLower_Control|TestProduct_WholeMutantsLower_000|TestProduct_WholeMutantsLower_001|TestProduct_WholeMutantsLower_002|TestProduct_WholeMutantsNative_Control|TestProduct_WholeMutantsNative_000|TestProduct_WholeMutantsNative_001|TestProduct_WholeMutantsNative_002|TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees|TestWholeMutants_Setup)$' > review/test-audit/stage1-typescript-parser-setup_deadline_products/S6.log 2>&1; whole_mutants_setup_test.go:75: shared setup must finish before a whole-mutant leaf runs"
  }
]
```

| ID | origin/main location | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/typescript/parser/nodes.ts:13 | `optional = false; -> optional = true;` | TestEveryTypeNodeKindAgrees, TestWholeGeneratedAgrees, TestObsoleteImportAttributesAgrees |
| M2 | stage1/typescript/parser/parser.ts:268 | `this.make('JSDocVariadicType', pos, [type]) -> this.make('JSDocNullableType', pos, [type])` | TestEveryTypeNodeKindAgrees |
| M3 | stage1/typescript/parser/statements.ts:238 | `node.operator = operator; -> node.operator = 'WithKeyword';` | TestObsoleteImportAttributesAgrees |
| M4 | stage1/typescript/parser/nodes.ts:58 | `export function countTree(nodes: readonly ParseNode[], root: number): number {     const node = nodes[root] ?? panic('missing parse node');     let count = 1; -> export function countTree(nodes: readonly ParseNode[], root: number): number {     const node = nodes[root] ?? panic('missing parse node');     let count = 0;` |  |
| S1 | stage1/typescript/parser/whole_mutants_split_test.go:91 | `filepath.Join(dir, "oracle"), virtual -> filepath.Join(dir, "absent/oracle"), virtual` | TestWholeMutants_Setup |
| S2 | stage1/typescript/parser/compiler_expressions_shards_test.go:67 | `os.WriteFile(filepath.Join(dir, "main.c"), -> os.WriteFile(filepath.Join(dir, "absent/main.c"),` | TestProduct_CompilerExpressionsLower, TestProduct_CompilerExpressionsNative |
| S3 | stage1/typescript/parser/compiler_expressions_shards_test.go:83 | `native.Build(string(source), filepath.Join(dir, "parser"), -> native.Build(string(source), filepath.Join(dir, "absent/parser"),` | TestProduct_CompilerExpressionsNative |
| S4 | stage1/typescript/parser/whole_mutants_split_test.go:144 | `os.WriteFile(filepath.Join(dir, "main.c"), -> os.WriteFile(filepath.Join(dir, "absent/main.c"),` | TestProduct_WholeMutantsLower family, TestProduct_WholeMutantsNative family, TestWholeMutants_Setup |
| S5 | stage1/typescript/parser/whole_mutants_split_test.go:160 | `native.Build(string(source), filepath.Join(dir, "scanner"), -> native.Build(string(source), filepath.Join(dir, "absent/scanner"),` | TestProduct_WholeMutantsNative family, TestWholeMutants_Setup |
| S6 | stage1/typescript/parser/whole_mutants_setup_test.go:105 | `	for i, source := range directories { 		if binaries[i] == "" { 			t.Fatal("setup product build did not complete") 		} 		inputs.Binaries[wholeMutantsSourceHash(t, source)] = binaries[i] 	}  -> ` | TestWholeMutants_Setup |

Production survivor M4: countTree stops counting each node. count-before.log prints 7, count-after.log and count-after-native.log print 0. This is unguarded count-only behavior in the bounded matrix; kills outside it are unknown. TestWholeCountCheckCatchesMutant is a caller outside the slice, not run here.

Code under test: TypeScript parser port; for construction rows the Go preparation recipes themselves. Oracle: typescript-go for semantic answers, self-written setup expectations for construction. No oracle source, copied-file list or corpus was mutated. S1-S6 edit only construction under the explicit setup-check exception. They never support sacred, subsumed or production kills.

Functions: functions-static.txt records declarations; functions-reached.json records observed V8 calls; callers.log lists additional callers. Static declarations are not claimed as proof that every branch was reached.

Brief ambiguities, limitations and costs:

- Fifteen named functions group into nine rows. Four Lower wrappers and four Native wrappers each share a recipe and differ only by input. Family timings are grouped package runs, not sums of member durations.
- Product-only tests do not execute a parser or compare an answer. Their construction faults need the allowed setup-check exception; production semantic defects passing them do not imply semantic coverage.
- Product wrappers discard returned paths. S1 tests whether a wrongly placed oracle artifact is noticed; the logs distinguish successful construction from prepared-input validation. Empty-entry probes show whether each row notices no construction at all.
- Whole-package timeout is not a red assertion baseline. It timed out before parallel rows could finish; narrowed clean baseline passed. All verdicts are bounded, and neither package-wide nor repo-wide uniqueness is established. Other callers identified by grep were not added when that would exceed the budget.
- The pinned corpus was absent and was installed at 050880ce59e30b356b686bd3144efe24f875ebc8. Corpus-dependent setup was enabled. Performance benchmarks outside this slice remain unmeasured; no requested row skipped in the bounded baseline.
- Four-rebuild stage1 exception limits production mutants below three per grouped row. Construction edits are listed separately, not used to inflate production uniqueness.
- count-only mode is reached by other package tests but not these semantic agreement rows. M4 survival is explicitly bounded to this matrix, with an observed output witness.
- Audit runner error: numeric count output is valid JSON but is not a test event. Parsing was fixed to accept only objects; completed timings were retained, and no production mutant had been planted before that failure.
- Optional type-kind coverage includes a corpus construction check and doc-type agreement in one row. The port entry exercised for doc-type agreement is docTypes, so P2 judges its vacuity, not P1 file.
- Empty-entry probes P3-P8 target construction subjects only for their own rows. Their effects on preparation for semantic rows are not semantic probe kills.

Timing: {"nproc": 5, "warm_tool_verification_seconds": 0.019276350998552516, "npm_seconds": 0.6945820190012455, "whole_baseline_wall_seconds": 91.77505630900123, "whole_baseline_binary_seconds": 90.023, "bounded_baseline_binary_seconds": 53.334, "builds": {"M1-build": 20.38339996400464, "M2-build": 19.889101626999036, "M3-build": 15.755020793003496, "M4-build": 16.05189452000195, "P1-build": 15.925584058000823, "P2-build": 14.987471633998211}, "recorded_run_wall_seconds": 1139.0638242979694, "all_recorded_wall_seconds": 1247.1841623910004}. Matrix test-binary times include internal product builds; standalone build validations have separate wall measurements. Warm setup reused, no cloud/setup.sh.

Not covered: full-package replay after narrowing, other packages, repository-wide uniqueness, all parser branches, nonlisted setup recipes, performance opt-ins and alternate systems. Standalone diffs apply to the pinned starting commit; build/vet and apply-check logs are saved.
