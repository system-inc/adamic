Three assigned rows defended by three aimed production mutants in completed bounded matrices.
No test or oracle changed; no deletion is recommended.
Exact commands, passing rows, standalone diffs and original coverage are in adjacent artifacts.

[
  {
    "test": "TestTokenizerEvents family",
    "package": "stage1/cohere/markdownblocks",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestMicromarkInputChunks"
    ],
    "defense": "defended",
    "unique_mutant": "E1 stage1/cohere/markdownblocks/tokenizerEvents.ts:75",
    "attempts": [
      {
        "mutant": "E1",
        "file_line": "stage1/cohere/markdownblocks/tokenizerEvents.ts:75",
        "change": "this.consumed = true; -> this.consumed = false;",
        "rows_failed": [
          "TestTokenizerEvents family"
        ],
        "failed_members": [
          "TestTokenizerEvents_255",
          "TestTokenizerEvents_511",
          "TestTokenizerEvents_003"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/markdown-defense/cache/E1-bounded timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^(TestMicromarkInputChunks|TestFrontMatterStage|TestParserRepresentationProbes|TestArrayGrowthWitness|TestMarkdownASTPreprocessing|TestMarkdownLeafComposition_000|TestMarkdownLeafComposition_001|TestMarkdownLeafComposition_002|TestMarkdownLeafComposition_003|TestMdastIdentifierScalars|TestTokenizerEvents_003|TestTokenizerEvents_255|TestTokenizerEvents_511)$' > E1-bounded.log 2>&1; tokenizer_events_independent_test.go:117: TestTokenizerEvents_255: native disagreement in case id 631",
    "rows_passed": [
      "TestArrayGrowthWitness",
      "TestFrontMatterStage",
      "TestMicromarkInputChunks",
      "TestMarkdownASTPreprocessing",
      "TestMdastIdentifierScalars",
      "TestParserRepresentationProbes",
      "TestMarkdownLeafComposition_001",
      "TestMarkdownLeafComposition_002",
      "TestMarkdownLeafComposition_003",
      "TestMarkdownLeafComposition_000"
    ],
    "bounded": true,
    "matrix_rows": [
      "TestMicromarkInputChunks",
      "TestFrontMatterStage",
      "TestParserRepresentationProbes",
      "TestArrayGrowthWitness",
      "TestMarkdownASTPreprocessing",
      "TestMarkdownLeafComposition_000",
      "TestMarkdownLeafComposition_001",
      "TestMarkdownLeafComposition_002",
      "TestMarkdownLeafComposition_003",
      "TestMdastIdentifierScalars",
      "TestTokenizerEvents_003",
      "TestTokenizerEvents_255",
      "TestTokenizerEvents_511"
    ],
    "rows_outside_matrix": "unknown"
  },
  {
    "test": "TestFrontMatterStage",
    "package": "stage1/cohere/markdownblocks",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestParserRepresentationProbes"
    ],
    "defense": "defended",
    "unique_mutant": "F1 stage1/cohere/markdownblocks/frontmatter.ts:56",
    "attempts": [
      {
        "mutant": "F1",
        "file_line": "stage1/cohere/markdownblocks/frontmatter.ts:56",
        "change": "text.indexOf('\\n', 3) -> text.indexOf('\\n', 4)",
        "rows_failed": [
          "TestFrontMatterStage"
        ],
        "failed_members": [
          "TestFrontMatterStage"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/markdown-defense/cache/F1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^(TestMicromarkInputChunks|TestFrontMatterStage|TestParserRepresentationProbes|TestArrayGrowthWitness|TestMarkdownASTPreprocessing|TestMarkdownLeafComposition_000|TestMarkdownLeafComposition_001|TestMarkdownLeafComposition_002|TestMarkdownLeafComposition_003|TestMdastIdentifierScalars|TestTokenizerEvents_003)$' > F1.log 2>&1; frontmatter_test.go:104: native front matter first byte difference at 67550 (lengths 654514/655456)",
    "rows_passed": [
      "TestArrayGrowthWitness",
      "TestMicromarkInputChunks",
      "TestMdastIdentifierScalars",
      "TestTokenizerEvents_003",
      "TestMarkdownASTPreprocessing",
      "TestParserRepresentationProbes",
      "TestMarkdownLeafComposition_001",
      "TestMarkdownLeafComposition_003",
      "TestMarkdownLeafComposition_002",
      "TestMarkdownLeafComposition_000"
    ],
    "bounded": true,
    "matrix_rows": [
      "TestMicromarkInputChunks",
      "TestFrontMatterStage",
      "TestParserRepresentationProbes",
      "TestArrayGrowthWitness",
      "TestMarkdownASTPreprocessing",
      "TestMarkdownLeafComposition_000",
      "TestMarkdownLeafComposition_001",
      "TestMarkdownLeafComposition_002",
      "TestMarkdownLeafComposition_003",
      "TestMdastIdentifierScalars",
      "TestTokenizerEvents_003"
    ],
    "rows_outside_matrix": "unknown"
  },
  {
    "test": "TestParserRepresentationProbes",
    "package": "stage1/cohere/markdownblocks",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFrontMatterStage"
    ],
    "defense": "defended",
    "unique_mutant": "P1 internal/lower/object.go:1265",
    "attempts": [
      {
        "mutant": "P1",
        "file_line": "internal/lower/object.go:1265",
        "change": "if len(arguments) != 1 {\n\t\t\treturn nil, true, l.notYet(node, \"push with other than one value\") -> if len(arguments) < 1 {\n\t\t\treturn nil, true, l.notYet(node, \"push with other than one value\")",
        "rows_failed": [
          "TestParserRepresentationProbes"
        ],
        "failed_members": [
          "TestParserRepresentationProbes"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/markdown-defense/cache/P1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdownblocks/ -run '^(TestMicromarkInputChunks|TestFrontMatterStage|TestParserRepresentationProbes|TestArrayGrowthWitness|TestMarkdownASTPreprocessing|TestMarkdownLeafComposition_000|TestMarkdownLeafComposition_001|TestMarkdownLeafComposition_002|TestMarkdownLeafComposition_003|TestMdastIdentifierScalars|TestTokenizerEvents_003)$' > P1.log 2>&1; gaps_test.go:51: NotYet gap changed: <nil>; update GAPS.md",
    "rows_passed": [
      "TestArrayGrowthWitness",
      "TestFrontMatterStage",
      "TestMicromarkInputChunks",
      "TestMdastIdentifierScalars",
      "TestTokenizerEvents_003",
      "TestMarkdownASTPreprocessing",
      "TestMarkdownLeafComposition_001",
      "TestMarkdownLeafComposition_002",
      "TestMarkdownLeafComposition_003",
      "TestMarkdownLeafComposition_000"
    ],
    "bounded": true,
    "matrix_rows": [
      "TestMicromarkInputChunks",
      "TestFrontMatterStage",
      "TestParserRepresentationProbes",
      "TestArrayGrowthWitness",
      "TestMarkdownASTPreprocessing",
      "TestMarkdownLeafComposition_000",
      "TestMarkdownLeafComposition_001",
      "TestMarkdownLeafComposition_002",
      "TestMarkdownLeafComposition_003",
      "TestMdastIdentifierScalars",
      "TestTokenizerEvents_003"
    ],
    "rows_outside_matrix": "unknown"
  }
]

## Evidence and limits

Read method.md for code under test, oracle, semantic differences and source reachability. Coverage-diffs.json compares each subject with its prior subsumer. Go profiles cover compiler packages, not TypeScript execution. The representative tokenizer profile is member 003. The full family matrix cooked at 90.056 binary seconds; uncompleted rows remain unknown. Its ordinary event mismatches are supplemental. The completed E1-bounded run is the verdict evidence: only 003/255/511 fail, grouped as one family.

F1 completed with only FrontMatterStage failing. The package has two independent front-matter ports: frontmatter.ts is this row's code under test; parseFrontMatter.ts belongs to AST construction. Confusing those imports would produce an incorrect reachability claim.

P1 completed with only RepresentationProbes failing on gaps/10_multiple_push.ts. The original rejection boundary becomes acceptance, and a standalone native rebuild prints 1 while Node prints 1,2. The evidence changes no expected diagnostic text. Positive/compiler controls pass. Go vet validates the standalone compiler patch; the witness validates native building.

Passed row lists and exact matrix selectors are in results.json and matrix.json. Full scope has 763 tests, identical names to the audit. No added or vanished test was found. Wider unassigned fixtures and unrun family members remain unknown. These defenses establish unique observed catches in the permitted bounded matrices, not an unbounded package-wide execution claim. Static call tracing supports the targeted differences but does not replace runtime evidence.

## Places the brief was unclear or cost time

- Whole-package baseline timed out at 90.107 seconds during cold suite preparation, without ordinary assertion failures. The bounded clean control completed and passed at 74.389 seconds. Four isolated coverage runs also passed.
- go -coverpkg cannot instrument the TypeScript port. Compiler coverage was requested and collected; TypeScript source imports and semantic input/state differences supply port leads. Full-family coverage cannot complete within 90 seconds, so member 003 is explicitly representative.
- The full 512-member tokenizer mutant run cooked. It was replayed over three members and ten comparison tests; no timeout-attributed row failure is counted as a kill.
- Native preparation fingerprinting includes broad source sets. Even a port-local mutation rebuilds unrelated selected products; build phase timing lines are kept in each log. Isolated caches were used for every mutant and the narrowed replay.
- The prior audit's mutual subsumption between front matter and representation was based on one broad native conditional-selection mutant. The current semantic mutants exercise different contracts.
- There are two separately implemented front-matter ports with different filenames and result interfaces. Only one belongs to FrontMatterStage.
- A multiple-push source scan initially counted legal trailing commas as extra arguments. It was corrected before any reachability conclusion: only gaps/10_multiple_push.ts has multiple arguments. The final inventory is saved.
- The tokenizer family contains local mutation witnesses in 000-002 and a construction union. The defense uses 003/255/511 positive comparisons, so witness preconditions and family members are not treated as competing rows.
- Source Node, emitted JavaScript and native executors are checked inside each assigned row. There are no separate Node/native twin rows to classify here.
- Throughput is logged without a speed threshold, but none of the assigned names promises a cost gate. Every assigned behavior has assertions and a proven semantic catch. No name/assertion mismatch was found for an undefended row because all three were defended.
- One successful aimed mutant per row was sufficient. The requested maximum of three attempts is not a requirement to keep searching after a unique catch.

## Time and scope

Warm tools reused, setup 0 seconds; nproc 5; npm ci reported 368 ms. Four completed coverage commands took about 37.913 wall seconds. Clean and mutant binary times: clean 74.389, full tokenizer 90.056 (cooked), F1 77.494, P1 recorded in matrix.json, narrowed tokenizer 74.900. Whole-package baseline 90.107 seconds. Compiler/source restoration and standalone apply checks succeeded. Session approximately 14 minutes; cold native phase times are included in logs, not separate end-to-end rebuild estimates.

No other package matrix, repo-wide replay, full 512-member completed tokenizer run, or actual TypeScript line-coverage profile was performed. No test was deleted, rewritten or weakened; no PR or main push.
