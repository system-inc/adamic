| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | stage1/cohere/estree/protocol.ts:20 | `unit >= 32 && unit <= 126` → `unit >= 33 && unit <= 126` | TestRecoveredGrammar, TestScalarEdges family |
| M02 | stage1/cohere/estree/values.ts:23 | `value.number = flag ? 1 : 0;` → `value.number = flag ? 0 : 1;` | TestRecoveredGrammar |
| M03 | stage1/cohere/estree/pipeline.ts:33 | `parser.awaitContext = true;` → `parser.awaitContext = false;` | TestRecoveredGrammar |
| M04 | stage1/cohere/estree/values.ts:81 | `let result = '0';` → `let result = '1';` | TestRecoveredGrammar, TestScalarEdges family |
