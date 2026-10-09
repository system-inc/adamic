Defended none of the three rows with these six production attempts.
Each row has three attempted differences caught by another observed row.
Tests remain unchanged; source restored; all diffs apply to origin/main 76c59c81e8617cea1892a01841494895927a712c.

[
  {
    "test": "TestJsxMemberNameRejection",
    "package": "stage1/typescript/parser",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestJsxNode"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "stage1/typescript/parser/jsx.ts:34",
        "change": "JSX expected name at -> JSX invalid name at",
        "rows_failed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections"
        ],
        "rows_passed": [
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "aim": "member diagnostic class, contrasted with private-member boundary"
      },
      {
        "mutant": "D2",
        "file_line": "stage1/typescript/scanner/scanner.ts:941",
        "change": "while(this.code() === 45 || isIdentifierPart(this.code())) -> while(this.code() === 46 || isIdentifierPart(this.code()))",
        "rows_failed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections",
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "rows_passed": [
          "TestProduct_JsxMutantsLower_008",
          "TestProduct_JsxMutantsNative_001",
          "TestProduct_JsxMutantsNative_003",
          "TestProduct_JsxMutantsNative_004",
          "TestProduct_JsxMutantsNative_005"
        ],
        "aim": "dash extension used by the permissive member variant versus legal names"
      },
      {
        "mutant": "D3",
        "file_line": "stage1/typescript/parser/jsx.ts:55",
        "change": "if(tag) { -> if(!tag) {",
        "rows_failed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections",
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "rows_passed": [
          "TestProduct_JsxMutantsLower_005",
          "TestProduct_JsxMutantsLower_006",
          "TestProduct_JsxMutantsLower_007",
          "TestProduct_JsxMutantsLower_008",
          "TestProduct_JsxMutantsNative_001",
          "TestProduct_JsxMutantsNative_006",
          "TestProduct_JsxMutantsNative_007",
          "TestProduct_JsxMutantsNative_008"
        ],
        "aim": "member descent versus plain tag and attribute names"
      }
    ],
    "evidence": [
      {
        "mutant": "D1",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^(TestJsxMutants_000|TestJsxMutants_001|TestJsxMutants_002|TestJsxMutants_003|TestJsxMutants_004|TestJsxMutants_005|TestJsxMutants_006|TestJsxMutants_007|TestJsxMutants_008|TestJsxMutantsUnion|TestProduct_JsxMutantsOracle|TestProduct_JsxMutantsLower_Control|TestProduct_JsxMutantsNative_Control|TestProduct_JsxMutantsLower_000|TestProduct_JsxMutantsNative_000|TestProduct_JsxMutantsLower_001|TestProduct_JsxMutantsNative_001|TestProduct_JsxMutantsLower_002|TestProduct_JsxMutantsNative_002|TestProduct_JsxMutantsLower_003|TestProduct_JsxMutantsNative_003|TestProduct_JsxMutantsLower_004|TestProduct_JsxMutantsNative_004|TestProduct_JsxMutantsLower_005|TestProduct_JsxMutantsNative_005|TestProduct_JsxMutantsLower_006|TestProduct_JsxMutantsNative_006|TestProduct_JsxMutantsLower_007|TestProduct_JsxMutantsNative_007|TestProduct_JsxMutantsLower_008|TestProduct_JsxMutantsNative_008|TestJsxMemberNameRejection|TestJsxNameBoundaryRejections|TestJsxScannerMutants|TestJsxNode|TestJsxNative)$ > D1-matrix.log 2>&1",
        "line": "jsx_rejection_test.go:51: Node failed to reject: 70 adamic: panic: JSX invalid name at 21 in /workspace/scratch/defend-jsx/tmp/TestJsxMemberNameRejection762463858/001/dashed-member.tsx",
        "log": "D1-matrix.log"
      },
      {
        "mutant": "D2",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^(TestJsxMutants_000|TestJsxMutants_001|TestJsxMutants_002|TestJsxMutants_003|TestJsxMutants_004|TestJsxMutants_005|TestJsxMutants_006|TestJsxMutants_007|TestJsxMutants_008|TestJsxMutantsUnion|TestProduct_JsxMutantsOracle|TestProduct_JsxMutantsLower_Control|TestProduct_JsxMutantsNative_Control|TestProduct_JsxMutantsLower_000|TestProduct_JsxMutantsNative_000|TestProduct_JsxMutantsLower_001|TestProduct_JsxMutantsNative_001|TestProduct_JsxMutantsLower_002|TestProduct_JsxMutantsNative_002|TestProduct_JsxMutantsLower_003|TestProduct_JsxMutantsNative_003|TestProduct_JsxMutantsLower_004|TestProduct_JsxMutantsNative_004|TestProduct_JsxMutantsLower_005|TestProduct_JsxMutantsNative_005|TestProduct_JsxMutantsLower_006|TestProduct_JsxMutantsNative_006|TestProduct_JsxMutantsLower_007|TestProduct_JsxMutantsNative_007|TestProduct_JsxMutantsLower_008|TestProduct_JsxMutantsNative_008|TestJsxMemberNameRejection|TestJsxNameBoundaryRejections|TestJsxScannerMutants|TestJsxNode|TestJsxNative)$ > D2-matrix.log 2>&1",
        "line": "jsx_rejection_test.go:66: node [--disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/scratch/defend-jsx/tmp/TestJsxMemberNameRejection4204156728/008/main.ts /workspace/scratch/defend-jsx/tmp/TestJsxMemberNameRejection4204156728/001/dashed-member.tsx --whole]: exit status 70",
        "log": "D2-matrix.log"
      },
      {
        "mutant": "D3",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^(TestJsxMutants_000|TestJsxMutants_001|TestJsxMutants_002|TestJsxMutants_003|TestJsxMutants_004|TestJsxMutants_005|TestJsxMutants_006|TestJsxMutants_007|TestJsxMutants_008|TestJsxMutantsUnion|TestProduct_JsxMutantsOracle|TestProduct_JsxMutantsLower_Control|TestProduct_JsxMutantsNative_Control|TestProduct_JsxMutantsLower_000|TestProduct_JsxMutantsNative_000|TestProduct_JsxMutantsLower_001|TestProduct_JsxMutantsNative_001|TestProduct_JsxMutantsLower_002|TestProduct_JsxMutantsNative_002|TestProduct_JsxMutantsLower_003|TestProduct_JsxMutantsNative_003|TestProduct_JsxMutantsLower_004|TestProduct_JsxMutantsNative_004|TestProduct_JsxMutantsLower_005|TestProduct_JsxMutantsNative_005|TestProduct_JsxMutantsLower_006|TestProduct_JsxMutantsNative_006|TestProduct_JsxMutantsLower_007|TestProduct_JsxMutantsNative_007|TestProduct_JsxMutantsLower_008|TestProduct_JsxMutantsNative_008|TestJsxMemberNameRejection|TestJsxNameBoundaryRejections|TestJsxScannerMutants|TestJsxNode|TestJsxNative)$ > D3-matrix.log 2>&1",
        "line": "jsx_rejection_test.go:66: node [--disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/scratch/defend-jsx/tmp/TestJsxMemberNameRejection3519992172/008/main.ts /workspace/scratch/defend-jsx/tmp/TestJsxMemberNameRejection3519992172/001/dashed-member.tsx --whole]: exit status 70",
        "log": "D3-matrix.log"
      }
    ],
    "coverage": "coverage-diffs.json; V8 profiles are supplemental and transformed-source offsets are not exact original line coverage."
  },
  {
    "test": "TestJsxNode",
    "package": "stage1/typescript/parser",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestJsxNative"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D4",
        "file_line": "stage1/typescript/parser/jsx.ts:53",
        "change": "name = this.make('ThisKeyword', pos); -> name = this.make('ThisKeyword', pos + 1);",
        "rows_failed": [
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "rows_passed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections"
        ],
        "aim": "Node/native shared source, this-tag position boundary"
      },
      {
        "mutant": "D5",
        "file_line": "stage1/typescript/parser/jsx.ts:71",
        "change": "this.parser.node(id).end = this.parser.scanner.fullStart; -> this.parser.node(id).end = this.parser.scanner.fullStart + 1;",
        "rows_failed": [
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "rows_passed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections"
        ],
        "aim": "Node/native shared text-end byte boundary, including Unicode and CRLF"
      },
      {
        "mutant": "D6",
        "file_line": "stage1/typescript/parser/jsx.ts:49",
        "change": "const right = this.identifier();\n            return this.make('JsxNamespacedName', pos, [name, right]); -> const right = this.identifier();\n            return this.make('JsxNamespacedName', pos, [right, name]);",
        "rows_failed": [
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "rows_passed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections"
        ],
        "aim": "Node/native shared namespace input, child order"
      }
    ],
    "evidence": [
      {
        "mutant": "D4",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^TestJsxNode$ > D4-TestJsxNode.log 2>&1",
        "line": "jsx_test.go:92: case 12 line 223: port \"7 ThisKeyword 12 15 0 0 -1 0 0\\t\\t\\t\\t\", Go \"7 ThisKeyword 11 15 0 0 -1 0 0\\t\\t\\t\\t\"",
        "log": "D4-TestJsxNode.log"
      },
      {
        "mutant": "D5",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^TestJsxNode$ > D5-TestJsxNode.log 2>&1",
        "line": "jsx_test.go:92: case 6 line 89: port \"6 JsxText 19 24 0 0 -1 0 0\\t\\ttext\\t\\t0\", Go \"6 JsxText 19 23 0 0 -1 0 0\\t\\ttext\\t\\t0\"",
        "log": "D5-TestJsxNode.log"
      },
      {
        "mutant": "D6",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^TestJsxNode$ > D6-TestJsxNode.log 2>&1",
        "line": "jsx_test.go:92: case 8 line 134: port \"8 Identifier 45 49 0 0 -1 0 0\\t\\tlang\\t\\t\", Go \"8 Identifier 40 44 0 0 -1 0 0\\t\\txml\\t\\t\"",
        "log": "D6-TestJsxNode.log"
      }
    ],
    "coverage": "coverage-diffs.json; V8 profiles are supplemental and transformed-source offsets are not exact original line coverage."
  },
  {
    "test": "TestJsxNative",
    "package": "stage1/typescript/parser",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestJsxNode"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D4",
        "file_line": "stage1/typescript/parser/jsx.ts:53",
        "change": "name = this.make('ThisKeyword', pos); -> name = this.make('ThisKeyword', pos + 1);",
        "rows_failed": [
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "rows_passed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections"
        ],
        "aim": "Node/native shared source, this-tag position boundary"
      },
      {
        "mutant": "D5",
        "file_line": "stage1/typescript/parser/jsx.ts:71",
        "change": "this.parser.node(id).end = this.parser.scanner.fullStart; -> this.parser.node(id).end = this.parser.scanner.fullStart + 1;",
        "rows_failed": [
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "rows_passed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections"
        ],
        "aim": "Node/native shared text-end byte boundary, including Unicode and CRLF"
      },
      {
        "mutant": "D6",
        "file_line": "stage1/typescript/parser/jsx.ts:49",
        "change": "const right = this.identifier();\n            return this.make('JsxNamespacedName', pos, [name, right]); -> const right = this.identifier();\n            return this.make('JsxNamespacedName', pos, [right, name]);",
        "rows_failed": [
          "TestJsxNative",
          "TestJsxNode",
          "TestJsxScannerMutants"
        ],
        "rows_passed": [
          "TestJsxMemberNameRejection",
          "TestJsxNameBoundaryRejections"
        ],
        "aim": "Node/native shared namespace input, child order"
      }
    ],
    "evidence": [
      {
        "mutant": "D4",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^TestJsxNative$ > D4-TestJsxNative.log 2>&1",
        "line": "jsx_test.go:107: case 12 line 223: port \"7 ThisKeyword 12 15 0 0 -1 0 0\\t\\t\\t\\t\", Go \"7 ThisKeyword 11 15 0 0 -1 0 0\\t\\t\\t\\t\"",
        "log": "D4-TestJsxNative.log"
      },
      {
        "mutant": "D5",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^TestJsxNative$ > D5-TestJsxNative.log 2>&1",
        "line": "jsx_test.go:107: case 6 line 89: port \"6 JsxText 19 24 0 0 -1 0 0\\t\\ttext\\t\\t0\", Go \"6 JsxText 19 23 0 0 -1 0 0\\t\\ttext\\t\\t0\"",
        "log": "D5-TestJsxNative.log"
      },
      {
        "mutant": "D6",
        "command": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run ^TestJsxNative$ > D6-TestJsxNative.log 2>&1",
        "line": "jsx_test.go:107: case 8 line 134: port \"8 Identifier 45 49 0 0 -1 0 0\\t\\tlang\\t\\t\", Go \"8 Identifier 40 44 0 0 -1 0 0\\t\\txml\\t\\t\"",
        "log": "D6-TestJsxNative.log"
      }
    ],
    "coverage": "coverage-diffs.json; V8 profiles are supplemental and transformed-source offsets are not exact original line coverage."
  }
]

CODE UNDER TEST: TypeScript parser/scanner port and native compilation. ORACLE: unchanged Go TypeScript parser; Node executes the port. Full details are in method.md.

Mutants, origin/main sites:

| ID | Site | Change |
|---|---|---|
| D1 | stage1/typescript/parser/jsx.ts:34 | JSX expected name at -> JSX invalid name at |
| D2 | stage1/typescript/scanner/scanner.ts:941 | while(this.code() === 45 \|\| isIdentifierPart(this.code())) -> while(this.code() === 46 \|\| isIdentifierPart(this.code())) |
| D3 | stage1/typescript/parser/jsx.ts:55 | if(tag) { -> if(!tag) { |
| D4 | stage1/typescript/parser/jsx.ts:53 | name = this.make('ThisKeyword', pos); -> name = this.make('ThisKeyword', pos + 1); |
| D5 | stage1/typescript/parser/jsx.ts:71 | this.parser.node(id).end = this.parser.scanner.fullStart; -> this.parser.node(id).end = this.parser.scanner.fullStart + 1; |
| D6 | stage1/typescript/parser/jsx.ts:49 | const right = this.identifier();             return this.make('JsxNamespacedName', pos, [name, right]); -> const right = this.identifier();             return this.make('JsxNamespacedName', pos, [right, name]); |

Coverage: coverage-diffs.json retains all exclusive Go blocks. Member versus Node: 4420 exclusive compiler blocks, 8 shared. Node versus Native: no exclusive compiler blocks, 8 shared. Native versus Node: 4421 exclusive compiler blocks, 8 shared. This reflects compilation, not port answer semantics. mapped-v8-coverage.json identifies member-only original-port mappings jsx.ts:33 and main.ts:64-66 versus Node. Boundary private-member input reaches the same invalid-name guard; Node/native use the same manifest corpus.

Unclear instructions and time costs:

- Whole-package baseline timed out at 90.017 seconds, with no assertion failure. The three defender coverage runs and all selected boundary/scanner/product/shard clean controls passed. Aggregate mutation matrices also cooked; all unfinished defender rows were rerun alone. Unknown leaves remain explicitly unknown.
- Go coverage cannot instrument the TypeScript port or child native executables. Requested Go profiles cover actual load/lower/native compilation; Node V8 profiles and clean-HEAD source maps supplement port coverage. Exact native-versus-Node TypeScript line exclusivity is unmeasured.
- The audit sampled only a few rows. Current listing has 100 tests, exactly the same names as the audit list. The defense matrix expands to all 36 JSX caller/product names, including previously omitted witnesses and preparation wrappers.
- The Node/native rows share every source input through jsxManifest. Their different backends and sanitized/release native modes remain distinct coverage obligations. Source-mutant overlap does not establish that either backend check can be removed.
- The member row combines original refusal with a permissive built-in source variant. D2 and D3 fail during that positive variant. These are observed test failures, not demonstrations that its original refusal check caught a regression.
- D3 preserves the original rejection marker but moves its diagnostic from token position 21 to 14; the row proceeds and fails later on its permissive variant. It compares Go diagnostic class and Node/native stderr equality, not Go diagnostic position.
- The member row logs that its check catches the built-in mutant exit zero, but never feeds that mutant through the refusal assertion. It only checks that permissive output contains custom-element. The refusal name itself is supported by its original assertions; the logged witness claim is stronger than demonstrated.
- Built-in mutant recipes are exact source-string matches. D6 changes an anchor used by a namespace mutant; any such construction failure is retained separately and excluded from behavioral catches.
- Six fixed-menu mutants were used, three per subject question with Node/native sharing three attempts. No oracle, harness or test was mutated. Compiler/runtime backend-specific semantic mutations beyond this port-source set were not explored.
- Results are bounded. No row was deleted or rewritten, and no absence-of-uniqueness result is a deletion recommendation. Full-package and repository-wide uniqueness outside the listed matrix remains unknown.

Timing: warm setup skipped (0 seconds), npm reports 606 milliseconds, nproc 5. Sum of completed matrix/control command walls: 769.328 seconds; coverage command walls: 59.408 seconds. The whole-package baseline binary took 90.017 seconds and timed out. Cold native rebuilds are included in command and row times, not separately phase-timed. No other package tests were run.
