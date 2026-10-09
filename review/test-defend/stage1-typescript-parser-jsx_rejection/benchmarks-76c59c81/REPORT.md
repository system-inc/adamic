Two benchmark rows defended by distinct semantic count mutants in the complete count-caller matrix.
Setup not uniquely defended; real production build faults refute its prior untrue classification, but are shared catches.
Origin/main 76c59c81e8617cea1892a01841494895927a712c; 100 current tests, no names added or removed since the audit.

[
  {
    "test": "TestCompilerExpressionsAgree_Setup",
    "package": "stage1/typescript/parser",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D3",
        "file_line": "internal/native/native.go:110",
        "change": "if options.Sanitize { return fmt.Errorf(\"native: sanitized build unavailable\") }",
        "rows_failed": [
          "TestCompilerExpressionsAgree_Setup",
          "TestProduct_CompilerExpressionsNative",
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
          "TestCompilerExpressionsAgree_015"
        ]
      },
      {
        "mutant": "D4",
        "file_line": "internal/native/native.go:127",
        "change": "if err := os.WriteFile(filepath.Join(directory, \"main.c\"), []byte(source), 0o644); err != nil { -> if err := os.WriteFile(filepath.Join(directory, \"main.c\"), []byte(source), 0o644); err == nil {",
        "rows_failed": [
          "TestCompilerExpressionsAgree_Setup",
          "TestProduct_CompilerExpressionsNative",
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
          "TestCompilerExpressionsAgree_015"
        ]
      },
      {
        "mutant": "D5",
        "file_line": "internal/native/native.go:110",
        "change": "if output != \"\" { return nil }",
        "rows_failed": [
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
          "TestCompilerExpressionsAgree_015"
        ]
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestCompilerExpressionsAgree_Setup|TestProduct_CompilerExpressionsNative|TestCompilerExpressionsAgree_000|TestCompilerExpressionsAgree_001|TestCompilerExpressionsAgree_002|TestCompilerExpressionsAgree_003|TestCompilerExpressionsAgree_004|TestCompilerExpressionsAgree_005|TestCompilerExpressionsAgree_006|TestCompilerExpressionsAgree_007|TestCompilerExpressionsAgree_008|TestCompilerExpressionsAgree_009|TestCompilerExpressionsAgree_010|TestCompilerExpressionsAgree_011|TestCompilerExpressionsAgree_012|TestCompilerExpressionsAgree_013|TestCompilerExpressionsAgree_014|TestCompilerExpressionsAgree_015)$' > D3.log 2>&1; compiler_expressions_shards_test.go:106: build compiler-expressions-native: native: sanitized build unavailable",
    "bounded": true,
    "matrix_rows": [
      "TestCompilerExpressionsAgree_Setup",
      "TestProduct_CompilerExpressionsNative",
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
      "TestCompilerExpressionsAgree_015"
    ],
    "rows_passed": [],
    "production_verdict": "subsumed",
    "current_subsumed_by": [
      "TestProduct_CompilerExpressionsNative"
    ]
  },
  {
    "test": "TestPerformance",
    "package": "stage1/typescript/parser",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestWholePerformance"
    ],
    "defense": "defended",
    "unique_mutant": "D1 stage1/typescript/parser/main.ts:29",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "stage1/typescript/parser/main.ts:29",
        "change": "let count = 0;\n    for(const root -> let count = 0;\n    if(countOnly && !whole && parser.roots.length > 3) { return 0; }\n    for(const root",
        "rows_failed": [
          "TestPerformance"
        ]
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestPerformance|TestWholePerformance|TestNodeCountCheckCatchesMutant|TestWholeCountCheckCatchesMutant)$' > D1.log 2>&1; performance_test.go:95: Node warm-up count \"17\\n\" differs from \"557010\\n\"",
    "bounded": true,
    "matrix_rows": [
      "TestPerformance",
      "TestWholePerformance",
      "TestNodeCountCheckCatchesMutant",
      "TestWholeCountCheckCatchesMutant"
    ],
    "rows_passed": [
      "TestWholePerformance",
      "TestNodeCountCheckCatchesMutant",
      "TestWholeCountCheckCatchesMutant"
    ]
  },
  {
    "test": "TestWholePerformance",
    "package": "stage1/typescript/parser",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestPerformance"
    ],
    "defense": "defended",
    "unique_mutant": "D2 stage1/typescript/parser/nodes.ts:58",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "stage1/typescript/parser/nodes.ts:58",
        "change": "let count = 1;\n    for(const child of node.children) {\n        count += countTree -> let count = 1;\n    if(node.kind === 'InterfaceDeclaration') { return 0; }\n    for(const child of node.children) {\n        count += countTree",
        "rows_failed": [
          "TestWholePerformance"
        ]
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestPerformance|TestWholePerformance|TestNodeCountCheckCatchesMutant|TestWholeCountCheckCatchesMutant)$' > D2.log 2>&1; performance_test.go:95: Node warm-up count \"847479\\n\" differs from \"887803\\n\"",
    "bounded": true,
    "matrix_rows": [
      "TestPerformance",
      "TestWholePerformance",
      "TestNodeCountCheckCatchesMutant",
      "TestWholeCountCheckCatchesMutant"
    ],
    "rows_passed": [
      "TestPerformance",
      "TestNodeCountCheckCatchesMutant",
      "TestWholeCountCheckCatchesMutant"
    ]
  }
]

CODE UNDER TEST: TestPerformance and TestWholePerformance exercise the TypeScript port main.run/countTree in stage1/typescript/parser/main.ts and nodes.ts, compiled natively and executed on Node. Setup exercises successful construction through production native.Build, including options and file writes. No oracle, test or harness source was edited.

ORACLE: The two benchmarks compare external TypeScript-Go recursive counts; Whole also compares full AST bytes before timing. Setup relies on successful construction without errors, a self-written construction criterion. It never executes or checks the returned product path.

Reach and baseline: The full enabled baseline timed out at 90.021 seconds during native preparation, without an ordinary assertion failure. The current count-caller matrix then passed in 82.520 seconds, and the setup/shared-caller matrix in 22.215 seconds. The three individual coverage runs also passed. No mutation run cooked. The count matrix runs TestPerformance, TestWholePerformance, TestNodeCountCheckCatchesMutant and TestWholeCountCheckCatchesMutant. Static caller search finds --count only in those four top-level tests. All other current rows construct products or run non-count output; D1's new return is guarded by countOnly and D2 changes only countTree, which the non-count path never calls. Thus the bounded count matrix includes every affected execution route. Other 96 rows were not replayed per counter mutant and are not claimed to have passed.

For the broad native-builder mutations, the bounded matrix runs Setup, TestProduct_CompilerExpressionsNative and all sixteen TestCompilerExpressionsAgree_NNN shards. Other native-build consumers remain unmeasured. These shared catches already defeat exclusivity, so wider execution cannot improve Setup's defense. All eighteen selected clean rows passed before mutation.

Coverage: Each requested row ran with -coverpkg=./internal/native,./internal/lower,./internal/load and a separate -coverprofile. The Setup comparison ran TestProduct_CompilerExpressionsNative with the same coverage options. Each comparison shares 4428 covered Go blocks and has no exclusive Go production block. Those profiles cover compiler preparation, not TypeScript or generated C subprocess execution. Node V8 coverage supplements them with actual port execution. TestPerformance has no exclusive port lines against Whole; Whole alone covers driver byte-position/printing paths. Function counts show countTree called 3342060 times in Performance and 5326818 times in Whole, with printTree called zero versus 887803 times. Shared lines still receive different roots and modes.

An independent observation of the unchanged parser over the pinned corpus found 77 files, 72 with more than three expression roots, a maximum of 23955 roots, and 789 interface declarations. The expression count witness has three roots; the whole count witness uses a const declaration and a type alias, with no interface declaration. D1 returns early only for count-only expression mode when root count exceeds three. D2 returns early only at interface nodes in countTree. Both preserve full AST printing and the built-in mutant replacement sites.

D1: TestPerformance fails with Node count 17 versus Go 557010. WholePerformance and both count witnesses pass. D2: WholePerformance fails with Node count 847479 versus Go 887803. Performance and both count witnesses pass. These are semantic omissions in large expression counting and declaration subtree counting, not speed changes. Standalone diffs compile through native Build in the real matrix, including sanitized count-witness products.

Setup attempts: D3 makes native.Build report unavailable sanitized compilation. Setup, its product wrapper and all sixteen compiler shards fail. D4 flips the source-write error condition, so a successful write is reported as an error; the same eighteen rows fail. These are production builder faults, not malformed generated C caught by -Werror. D5 returns success without constructing an executable. Setup and its product wrapper pass, but all sixteen downstream shards fail with missing executable errors. D5 is an empty-success construction probe and contributes no positive uniqueness claim. Three attempts find no Setup-only catch. D3/D4 nevertheless prove Setup can fail under real production mutations, so untrue is too strong: its observed production status is subsumed by TestProduct_CompilerExpressionsNative.

Findings about names: Neither performance row asserts a timing threshold, ratio bound or regression tolerance. They establish different counting contracts and report timing. The Whole name also promises a full-tree preflight, which its assertions do check. Setup does initialize compiler products and detects reported build failures, but successful initialization does not establish artifact existence. D5 directly demonstrates that limitation. No test was deleted, rewritten or weakened.

Friction and unclear instructions:
1. The full parser package exceeds the 90-second binary budget before reaching many parallel rows. I narrowed by current count-only callers and by Setup's shared build callers. All scopes and unknown results outside them are explicit.
2. Go -coverprofile cannot cover a TypeScript port or generated C executing in subprocesses. Go compiler/builder coverage plus Node V8 port coverage avoids presenting harness coverage as parser execution coverage. Native instruction-level coverage was not collected.
3. The setup row's requested comparison is the rest of the package. One other row covers every production block it covers; therefore its set of exclusive Go blocks against the whole rest is also empty. I measured that sufficient subset rather than replaying the entire rest with coverage, which already exceeded budget.
4. Exclusive lines were not enough: the benchmark counter visits shared code over different graph roots. Whole's exclusive print preflight is already covered elsewhere, so I aimed at declaration counting rather than treating those lines as uniqueness proof.
5. The verdict instructions say an untrue row with a shared production catch becomes subsumed, but the final defense enum excludes subsumed. rows.json uses defense=not defended for Setup and production_verdict=subsumed with the observed catcher. This preserves both meanings.
6. Setup's prior audit weakened construction in a harness helper. That showed a missing-artifact blind spot, but did not establish that no production mutation can fail Setup. The two production error mutants correct that inference.
7. Aiming at inputs is permitted, but named test/path selectors would not provide a semantic defense. D1 uses expression-mode root-list size, and D2 uses a declaration kind. Neither checks a test name, corpus filename or audit environment selector.
8. Every count-mode route was included, including the newer whole-count witness in the current package. Scope comparison actually found no added or removed test names; no update was inferred merely from the newer origin/main commit.
9. Early-return guards preserve TypeScript flow narrowing and built-in mutation-site strings. A whole-body return or replacement could have created preparation failures rather than semantic counter evidence.
10. The repository ignores .log files, so evidence logs must be explicitly force-added. Source restoration was verified before packaging.

Costs and limits: Warm setup skipped, npm ci reported 350 ms, nproc 5. Five mutations were run, below the seven-mutant ceiling; both count matrices took about 73 command-wall seconds and stayed below 90 binary seconds. Builder matrices took about 17 wall seconds each. Native rebuilds are included in real matrix runtime. Every Go builder diff passed go vet ./internal/native/. All standalone diffs apply independently to the starting origin/main commit. No other packages, broad builder consumers, exhaustive native coverage or repo-wide uniqueness were tested. Source edits are restored. Full report, profiles, coverage differences, caller searches, input-shape observation and raw JSON logs are retained.

The requested defense branch already held another worker's JSX defense. The initial push was rejected. This session was moved into benchmarks-76c59c81 and the remote history merged, preserving both evidence sets without force-pushing.

V8 source-offset correction: the runner uses stripTypeScriptTypes(mode=transform), so direct original-source line offsets were only preliminary estimates. Final coverage-diffs.json and port-v8-coverage.json map generated UTF-16 offsets through source maps. maps.mjs, map_coverage.py and source-maps preserve the mapping evidence. The semantic counter defenses and function-call counts are unchanged. This correction cost extra packaging time.
