Replayed all nine mutants with the deletion set skipped. The green baseline took 1,093.4 seconds; replays took 5,559.1 seconds. No diffs were stale. D6 required a timeout rerun; 12 optional tests skipped. Three mutants lost their last catcher, so at least one interface gap test must remain. Production code is restored, nothing was deleted, and [evidence is pushed](https://github.com/system-inc/adamic/tree/213cc82768e9446f8dcb579d6529d75e29722697/review/test-defend/deletion-set/stage1-cohere-estree).

```json
{
  "package": "stage1/cohere/estree",
  "main": "7b9d4272c28f59530ab13daa5c49067e47933b06",
  "skipped": [
    "TestInterfaceDefaultGap",
    "TestInterfaceTypeMethodGap",
    "TestUnattachedDecoratorControl"
  ],
  "mutants": [
    {
      "mutant": "M06",
      "file_line": "internal/lower/diagnostics.go:33",
      "branch": "test-audit/stage1-cohere-estree-gaps",
      "candidates_failed": ["TestInterfaceDefaultGap", "TestInterfaceTypeMethodGap"],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D1",
      "file_line": "stage1/cohere/estree/sourceParser.ts:1189",
      "branch": "test-defend/stage1-cohere-estree-deep",
      "candidates_failed": ["TestUnattachedDecoratorControl"],
      "still_caught_by": ["TestDecoratedExports"],
      "stale": false
    },
    {
      "mutant": "D3",
      "file_line": "stage1/cohere/estree/protocol.ts:36",
      "branch": "test-defend/stage1-cohere-estree-deep",
      "candidates_failed": ["TestUnattachedDecoratorControl"],
      "still_caught_by": ["TestAcceptanceGrammar"],
      "stale": false
    },
    {
      "mutant": "D2",
      "file_line": "internal/lower/class_inheritance.go:761",
      "branch": "test-defend/stage1-cohere-estree-gaps",
      "candidates_failed": ["TestInterfaceDefaultGap"],
      "still_caught_by": ["TestAcceptanceDiagnostics", "TestAcceptanceGrammar"],
      "witness_failures": ["TestAcceptanceDiagnosticControl"],
      "stale": false
    },
    {
      "mutant": "D3",
      "file_line": "internal/lower/iteration_origin.go:95",
      "branch": "test-defend/stage1-cohere-estree-gaps",
      "candidates_failed": ["TestInterfaceDefaultGap", "TestInterfaceTypeMethodGap"],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D4",
      "file_line": "internal/lower/iteration_origin.go:88",
      "branch": "test-defend/stage1-cohere-estree-gaps",
      "candidates_failed": ["TestInterfaceDefaultGap", "TestInterfaceTypeMethodGap"],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D5",
      "file_line": "internal/lower/expression.go:1024",
      "branch": "test-defend/stage1-cohere-estree-gaps",
      "candidates_failed": ["TestInterfaceTypeMethodGap"],
      "still_caught_by": ["TestAcceptanceGrammar"],
      "stale": false
    },
    {
      "mutant": "D6",
      "file_line": "internal/lower/expression.go:508",
      "branch": "test-defend/stage1-cohere-estree-gaps",
      "candidates_failed": ["TestInterfaceTypeMethodGap"],
      "still_caught_by": ["TestAcceptanceDiagnostics"],
      "panicking_tests": ["TestAcceptanceGrammar"],
      "note": "Thirty-minute timeout excluded from catcher credit; fresh-cache rerun failed cleanly on the diagnostics test's child CPU deadline.",
      "stale": false
    },
    {
      "mutant": "D7",
      "file_line": "internal/lower/expression.go:783",
      "branch": "test-defend/stage1-cohere-estree-gaps",
      "candidates_failed": ["TestInterfaceTypeMethodGap"],
      "still_caught_by": ["TestAcceptanceGrammar"],
      "stale": false
    }
  ],
  "keep": [
    {
      "test": "TestInterfaceDefaultGap",
      "because": "Shares the last catches of M06 and gaps D3/D4 with TestInterfaceTypeMethodGap; deleting both loses them."
    },
    {
      "test": "TestInterfaceTypeMethodGap",
      "because": "Shares the last catches of M06 and gaps D3/D4 with TestInterfaceDefaultGap; deleting both loses them."
    }
  ],
  "keep_requirement": "Retain at least one interface gap test; individual necessity of both was not established.",
  "deletable": ["TestUnattachedDecoratorControl"]
}
```
