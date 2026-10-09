With both candidates skipped, all nine mutants remain caught by ordinary tests outside the set. The baseline passed in 250.690 reported seconds; replay wall times ranged from 16.027 to 149.610 seconds, stopping at the first ordinary catch. No diffs were stale, no Go panics occurred, and production sources were restored. The baseline had 20 default opt-in skips. [Evidence](https://github.com/system-inc/adamic/tree/test-defend/deletion-set/stage1-typescript-parser/review/test-defend/deletion-set/stage1-typescript-parser) was pushed as `b0025984`.

```json
{
  "package": "stage1/typescript/parser",
  "main": "7b9d4272c28f59530ab13daa5c49067e47933b06",
  "skipped": [
    "TestJsxMemberNameRejection",
    "TestWholeCompilerAgrees"
  ],
  "mutants": [
    {
      "mutant": "audit-jsx_rejection-M1",
      "file_line": "stage1/typescript/parser/jsx.ts:33",
      "branch": "test-audit/stage1-typescript-parser-jsx_rejection",
      "candidates_failed": ["TestJsxMemberNameRejection"],
      "still_caught_by": ["TestJsxNameBoundaryRejections"],
      "stale": false
    },
    {
      "mutant": "audit-whole_mutants_split-M1",
      "file_line": "stage1/typescript/parser/nodes.ts:47",
      "branch": "test-audit/stage1-typescript-parser-whole_mutants_split",
      "candidates_failed": ["TestWholeCompilerAgrees"],
      "still_caught_by": ["TestJsxNative"],
      "stale": false
    },
    {
      "mutant": "audit-whole_mutants_split-M4",
      "file_line": "stage1/typescript/parser/parser.ts:37",
      "branch": "test-audit/stage1-typescript-parser-whole_mutants_split",
      "candidates_failed": ["TestWholeCompilerAgrees"],
      "still_caught_by": ["TestJsxNative"],
      "stale": false
    },
    {
      "mutant": "defend-jsx_rejection-D1",
      "file_line": "stage1/typescript/parser/jsx.ts:34",
      "branch": "test-defend/stage1-typescript-parser-jsx_rejection",
      "candidates_failed": ["TestJsxMemberNameRejection"],
      "still_caught_by": ["TestJsxNameBoundaryRejections"],
      "stale": false
    },
    {
      "mutant": "defend-jsx_rejection-D2",
      "file_line": "stage1/typescript/scanner/scanner.ts:941",
      "branch": "test-defend/stage1-typescript-parser-jsx_rejection",
      "candidates_failed": ["TestJsxMemberNameRejection"],
      "still_caught_by": ["TestJsxNameBoundaryRejections"],
      "stale": false
    },
    {
      "mutant": "defend-jsx_rejection-D3",
      "file_line": "stage1/typescript/parser/jsx.ts:55",
      "branch": "test-defend/stage1-typescript-parser-jsx_rejection",
      "candidates_failed": ["TestJsxMemberNameRejection"],
      "still_caught_by": ["TestJsxNameBoundaryRejections"],
      "stale": false
    },
    {
      "mutant": "defend-whole_mutants_split-D1",
      "file_line": "stage1/typescript/parser/statements.ts:80",
      "branch": "test-defend/stage1-typescript-parser-whole_mutants_split",
      "candidates_failed": ["TestWholeCompilerAgrees"],
      "still_caught_by": ["TestJsxNative"],
      "stale": false
    },
    {
      "mutant": "defend-whole_mutants_split-D2",
      "file_line": "stage1/typescript/parser/statements.ts:324",
      "branch": "test-defend/stage1-typescript-parser-whole_mutants_split",
      "candidates_failed": ["TestWholeCompilerAgrees"],
      "still_caught_by": ["TestWholeGeneratedAgrees"],
      "stale": false
    },
    {
      "mutant": "defend-whole_mutants_split-D3",
      "file_line": "stage1/typescript/parser/statements.ts:385",
      "branch": "test-defend/stage1-typescript-parser-whole_mutants_split",
      "candidates_failed": ["TestWholeCompilerAgrees"],
      "still_caught_by": ["TestWholeGeneratedAgrees"],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": [
    "TestJsxMemberNameRejection",
    "TestWholeCompilerAgrees"
  ]
}
```
