With `TestLexerMatchesGo` skipped, the baseline passed in 428.8 seconds and all six mutants retained non-witness catchers. Replays took 365.3 seconds; none were stale or broken. Slow runs stopped after their first clean catch. Production sources are restored, and evidence is pushed to `test-defend/deletion-set/stage1-cohere-yaml` under `review/test-defend/deletion-set/stage1-cohere-yaml/`.

```json
{
  "package": "stage1/cohere/yaml",
  "main": "b8bcadb2c493173855f19d7e5c508b34f5eeb5b6",
  "skipped": ["TestLexerMatchesGo"],
  "mutants": [
    {
      "mutant": "M1",
      "file_line": "internal/lower/object.go:1265",
      "branch": "test-audit/stage1-cohere-yaml-gaps",
      "candidates_failed": ["TestLexerMatchesGo"],
      "still_caught_by": ["TestFileDriver_Setup"],
      "stale": false
    },
    {
      "mutant": "D3",
      "file_line": "stage1/cohere/yaml/lexer.ts:635",
      "branch": "test-defend/stage1-cohere-yaml-gaps",
      "candidates_failed": ["TestLexerMatchesGo"],
      "still_caught_by": ["TestComposeMatchGo"],
      "stale": false
    },
    {
      "mutant": "D2",
      "file_line": "stage1/cohere/yaml/lexer.ts:483",
      "branch": "test-defend/stage1-cohere-yaml-gaps",
      "candidates_failed": ["TestLexerMatchesGo"],
      "still_caught_by": ["TestComposeMatchGo"],
      "stale": false
    },
    {
      "mutant": "D1",
      "file_line": "stage1/cohere/yaml/lexer.ts:359",
      "branch": "test-defend/stage1-cohere-yaml-gaps",
      "candidates_failed": ["TestLexerMatchesGo"],
      "still_caught_by": ["TestComposeMatchGo"],
      "stale": false
    },
    {
      "mutant": "D3",
      "file_line": "stage1/cohere/yaml/lexer.ts:286",
      "branch": "test-defend/stage1-cohere-yaml-gaps",
      "candidates_failed": ["TestLexerMatchesGo"],
      "still_caught_by": ["TestComposeMatchGo"],
      "stale": false
    },
    {
      "mutant": "D4",
      "file_line": "stage1/cohere/yaml/lexer.ts:365",
      "branch": "test-defend/stage1-cohere-yaml-gaps",
      "candidates_failed": ["TestLexerMatchesGo"],
      "still_caught_by": ["TestComposeMatchGo"],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": ["TestLexerMatchesGo"]
}
```
