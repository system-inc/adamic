TestWholeCompilerAgrees was not uniquely defended by three aimed production mutants.
All three fail Compiler and pass Yield, but WholeGeneratedAgrees and enabled WholePerformance also fail.
Evidence is bounded; all tests are unchanged and production source is restored.

[
  {
    "test": "TestWholeCompilerAgrees",
    "package": "stage1/typescript/parser",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestYieldLookaheadAgrees"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "stage1/typescript/parser/statements.ts:80",
        "change": "const declaration flags become let flags",
        "rows_failed": [
          "TestJsxNative",
          "TestJsxNode",
          "TestWholeCompilerAgrees",
          "TestWholeGeneratedAgrees",
          "TestWholePerformance"
        ]
      },
      {
        "mutant": "D2",
        "file_line": "stage1/typescript/parser/statements.ts:324",
        "change": "drop import-clause phase assignment",
        "rows_failed": [
          "TestObsoleteImportAttributesAgrees",
          "TestWholeCompilerAgrees",
          "TestWholeGeneratedAgrees",
          "TestWholePerformance"
        ]
      },
      {
        "mutant": "D3",
        "file_line": "stage1/typescript/parser/statements.ts:385",
        "change": "flip export type-only semantic flag",
        "rows_failed": [
          "TestObsoleteImportAttributesAgrees",
          "TestWholeCompilerAgrees",
          "TestWholeGeneratedAgrees",
          "TestWholePerformance"
        ]
      }
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/u158/corpus ADAMIC_PARSER_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend158/cache/D3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestWholeCompilerAgrees|TestYieldLookaheadAgrees|TestWholeGeneratedAgrees|TestWholePerformance)$' > D3-defense.log 2>&1; whole_test.go:34: Node: case 0 line 4: port \"1 ExportDeclaration 0 109 0 0 -1 0 0\\t\\t\\t\\t1\", Go \"1 ExportDeclaration 0 109 0 0 -1 0 0\\t\\t\\t\\t0\"",
    "observed_subsumers": [
      "TestWholeGeneratedAgrees",
      "TestWholePerformance"
    ],
    "bounded": true,
    "matrix_rows": [
      "TestCompilerExpressionsAgreeUnion",
      "TestCompilerExpressionsAgree_000",
      "TestCompilerExpressionsAgree_001",
      "TestCompilerExpressionsAgree_002",
      "TestCompilerExpressionsAgree_003",
      "TestCompilerExpressionsAgree_004",
      "TestCompilerExpressionsAgree_005",
      "TestCompilerExpressionsAgree_006",
      "TestCompilerExpressionsAgree_007",
      "TestCompilerExpressionsAgree_008",
      "TestCompilerExpressionsAgree_009",
      "TestCompilerExpressionsAgree_010",
      "TestCompilerExpressionsAgree_011",
      "TestCompilerExpressionsAgree_012",
      "TestCompilerExpressionsAgree_013",
      "TestCompilerExpressionsAgree_014",
      "TestCompilerExpressionsAgree_015",
      "TestCompilerExpressionsAgree_Setup",
      "TestEveryTypeNodeKindAgrees",
      "TestExpressionsAgree",
      "TestGeneratedExpressionsAgree",
      "TestJsxNative",
      "TestJsxNode",
      "TestObsoleteImportAttributesAgrees",
      "TestTypeOnlyImportCycleCompiles",
      "TestWholeCompilerAgrees",
      "TestWholeGeneratedAgrees",
      "TestWholePerformance",
      "TestYieldLookaheadAgrees"
    ],
    "code_under_test": "TypeScript parser/scanner port and whole-tree serialized declaration metadata",
    "oracle": "Unmodified typescript-go parser via testdata/oracle.go. Node and sanitized native run the port.",
    "name_assertion_gap": "None found: it compares complete serialized trees for all 77 files of the pinned compiler corpus. It promises no timing threshold or separate countTree comparison."
  }
]

CODE UNDER TEST: TypeScript parser/scanner port and whole-tree serialized declaration metadata. ORACLE: Unmodified typescript-go parser via testdata/oracle.go. Node and sanitized native run the port.

Starting origin/main: 76c59c81e8617cea1892a01841494895927a712c. Audit evidence: branch test-audit/stage1-typescript-parser-whole_mutants_split, commit 18587185, starting at 23a19b8b. The full report, rows, code/oracle contract, inventories, menus, matrix and original production diffs were read and preserved. The current go test -list census has the same Test names as the audit: no additions or disappearances.

The requested Go coverpkg profiles report no statements because the code under test is TypeScript. Both profiles are preserved. NODE_V8_COVERAGE collected actual Node port execution for each isolated row. Same-transform inline source maps map coverage points to original TypeScript lines. Native execution was not instrumented. Details and limitations are in coverage-method.txt; raw profiles, transformed source maps, both line sets and exclusive lines are saved. Compiler versus Yield call counts are variableList 15673 versus 0, importDeclaration 88 versus 0, and exportDeclaration 77 versus 0. D1: statements.ts:80 and D2:324 have exclusive mapped points. D3:385 is in the exclusive exportDeclaration function, with adjacent exclusive points at 384 and 386. This gave three semantic leads, not a uniqueness verdict. Compiler prints complete file trees including semantic declaration fields; Yield prints expression roots without those whole-tree fields.

| Mutant | Starting file:line | Allowed change | Observed failing rows |
|---|---|---|---|
| D1 | stage1/typescript/parser/statements.ts:80 | const declaration flags become let flags | TestJsxNative, TestJsxNode, TestWholeCompilerAgrees, TestWholeGeneratedAgrees, TestWholePerformance |
| D2 | stage1/typescript/parser/statements.ts:324 | drop import-clause phase assignment | TestObsoleteImportAttributesAgrees, TestWholeCompilerAgrees, TestWholeGeneratedAgrees, TestWholePerformance |
| D3 | stage1/typescript/parser/statements.ts:385 | flip export type-only semantic flag | TestObsoleteImportAttributesAgrees, TestWholeCompilerAgrees, TestWholeGeneratedAgrees, TestWholePerformance |

Each standalone diff has no switch and applies to the clean starting index. Each compiles through TestProduct_WholeMutantsNative_Control using the real sanitized port builder, with its own ADAMIC_BUILD_CACHE_DIR. D2 drops one assignment; phase remains used in control flow, so no unused declaration remains. Native catches are observed for D1 in TestJsxNative. Compiler, WholeGenerated and WholePerformance catches happen on Node output before later native comparisons; native build success is not presented as proof that a later assertion fired. All mutation runs completed within their test-binary budgets and have complete observed outcomes. No production mutant survived the selected matrix.

Every mutant matrix covers these current agreement/build rows, in separate bounded groups: TestCompilerExpressionsAgreeUnion, TestCompilerExpressionsAgree_000, TestCompilerExpressionsAgree_001, TestCompilerExpressionsAgree_002, TestCompilerExpressionsAgree_003, TestCompilerExpressionsAgree_004, TestCompilerExpressionsAgree_005, TestCompilerExpressionsAgree_006, TestCompilerExpressionsAgree_007, TestCompilerExpressionsAgree_008, TestCompilerExpressionsAgree_009, TestCompilerExpressionsAgree_010, TestCompilerExpressionsAgree_011, TestCompilerExpressionsAgree_012, TestCompilerExpressionsAgree_013, TestCompilerExpressionsAgree_014, TestCompilerExpressionsAgree_015, TestCompilerExpressionsAgree_Setup, TestEveryTypeNodeKindAgrees, TestExpressionsAgree, TestGeneratedExpressionsAgree, TestJsxNative, TestJsxNode, TestObsoleteImportAttributesAgrees, TestTypeOnlyImportCycleCompiles, TestWholeCompilerAgrees, TestWholeGeneratedAgrees, TestWholePerformance, TestYieldLookaheadAgrees. The native compilation validation row is recorded separately. All had clean observed passes before mutation. The shared build caches are isolated per mutant, including cached products copied from the mutated port. The matrices do not claim a full-package census of kills. Other construction, compiler-gap, count-witness, built-in-mutant-witness and rejection rows were not replayed as production agreement checks. In particular, a witness control/precondition failure would not prove its guarded check. Their results are unknown. No newly added row was omitted: the census is unchanged, and TypeOnlyImportCycleCompiles is included even though its fixture compiles an unrelated program rather than the parser port.

The brief and costs need these qualifications:

- The Go coverage instruction cannot directly instrument a .ts port. Empty Go profiles were preserved rather than being called exclusive production coverage; V8 execution and source maps supplied usable leads. Generated JavaScript offsets differ from original TypeScript offsets, so original coordinates were recovered through source maps, not by counting lines in the .ts file at a V8 offset.

- The original audit subsumption rests on two broad mutations in a two-row bounded matrix. These three new mutations falsify that specific Compiler-by-Yield subsumption on declaration metadata. They do not uniquely defend Compiler because two other named rows catch all three. The not-defended verdict is about these attempts; it does not prove behavioral equivalence or authorize deletion.

- WholePerformance is disabled by default. ADAMIC_PARSER_BENCH=1 was enabled for its green baseline and every mutation run containing it. Its whole-tree preflight uses the same pinned compiler corpus, but release native rather than sanitized native. WholeGenerated also catches all three with independent smaller whole-file inputs. Compiler retains a default sanitized whole-corpus comparison, a distinction these mutations do not remove.

- Whole-package baseline cooked at 90.080 test-binary seconds during sequential JSX rejection work, without a failing Test event before timeout. A first combined reached-row baseline cooked at 90.109 seconds; another group cooked at 90.031 with ExpressionsAgree alone unobserved. Smaller clean groups passed, including a 70.709-second isolated ExpressionsAgree baseline. The runner initially labels a timeout as baseline not green and stops; it then resumes only after narrowing to a clean pass. No red assertion baseline was used.

- Cold native builds and the ordinary expression row dominate cost: that row prepares multiple port variants serially, even though its expression output does not expose these declaration semantic fields. Per-mutant caches force fresh compilation as requested. No matrix timeout, Go panic abort or unobserved selected row remains. The full package was not repeated per mutant after its baseline timeout; outside the listed bounded rows, kills remain unknown.

- The requested row name matches its assertions: it compares full serialized trees for 77 pinned compiler files on Node and sanitized native. It does not claim a performance threshold or a separately computed node count. No name/assertion gap was demonstrated.

- Warm env.sh worked, so setup bootstrap was skipped. npm ci in stage3/api ran before baseline but was not separately timed. The pinned TypeScript corpus at /tmp/u158/corpus was already present and verified by compilerManifest. nproc was measured as 5. No oracle, harness, test or source input was changed; only three port-source mutations ran. Production source was restored before diff checks and the evidence commit.

Build and run costs:

{
  "bootstrap_seconds": 0,
  "npm_ci": "Run before baseline; not separately timed.",
  "nproc": 5,
  "native_build_validation": [
    {
      "mutant": "D1",
      "wall_seconds": 31.702286407999054,
      "test_binary_seconds": 28.459,
      "log": "D1-build.log"
    },
    {
      "mutant": "D2",
      "wall_seconds": 30.1596254770011,
      "test_binary_seconds": 24.292,
      "log": "D2-build.log"
    },
    {
      "mutant": "D3",
      "wall_seconds": 26.633562900999095,
      "test_binary_seconds": 23.547,
      "log": "D3-build.log"
    }
  ],
  "bounded_run_wall_seconds": 1048.67356110099,
  "matrix_wall_seconds": 745.0299269279931,
  "session": "about 27 minutes; see raw log timestamps",
  "limitations": "Whole-package and first combined reached baselines, plus isolated coverage/benchmark baselines, are outside bounded_run_wall_seconds. The cooked baseline-other-agreements run is included in that total. npm was not separately timed. No mutation matrix run cooked."
}

All commands, passed rows and failing output lines are in runs.json, matrix.json and raw logs. No main push and no pull request.
