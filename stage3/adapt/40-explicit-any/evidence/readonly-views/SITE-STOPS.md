# The 19 enum-view sites

All 19 real cast locations are Refused before and NotYet after in the existing refusal scan. No final cast has a Refused finding. Before diagnostics originate in invariance.go:673; final namespace diagnostics originate in module_namespace.go:27. These are measurement observations, not successful native lowering.

| Rule | Site / owner | Before | Final |
| --- | --- | --- | --- |
| 72 | `debug.ts:424:34`, formatSyntaxKind, SyntaxKind | Refused: writable record | NotYet: namespace value |
| 73 | `debug.ts:428:34`, formatSnippetKind, SnippetKind | Refused: writable record | NotYet: namespace value |
| 74 | `debug.ts:432:34`, formatScriptKind, ScriptKind | Refused: writable record | NotYet: namespace value |
| 75 | `debug.ts:436:35`, formatNodeFlags, NodeFlags | Refused: writable record | NotYet: namespace value |
| 76 | `debug.ts:440:35`, formatNodeCheckFlags, NodeCheckFlags | Refused: writable record | NotYet: namespace value |
| 77 | `debug.ts:444:35`, formatModifierFlags, ModifierFlags | Refused: writable record | NotYet: namespace value |
| 78 | `debug.ts:448:35`, formatTransformFlags, TransformFlags | Refused: writable record | NotYet: namespace value |
| 79 | `debug.ts:452:35`, formatEmitFlags, EmitFlags | Refused: writable record | NotYet: namespace value |
| 80 | `debug.ts:456:35`, formatSymbolFlags, SymbolFlags | Refused: writable record | NotYet: namespace value |
| 81 | `debug.ts:460:35`, formatTypeFlags, TypeFlags | Refused: writable record | NotYet: namespace value |
| 82 | `debug.ts:464:35`, formatSignatureFlags, SignatureFlags | Refused: writable record | NotYet: namespace value |
| 83 | `debug.ts:468:35`, formatObjectFlags, ObjectFlags | Refused: writable record | NotYet: namespace value |
| 84 | `debug.ts:472:35`, formatFlowFlags, FlowFlags | Refused: writable record | NotYet: namespace value |
| 85 | `debug.ts:476:36`, formatRelationComparisonResult, RelationComparisonResult | Refused: writable record | NotYet: namespace value |
| 86 | `debug.ts:480:34`, formatCheckMode, CheckMode | Refused: writable record | NotYet: namespace value |
| 87 | `debug.ts:484:34`, formatSignatureCheckMode, SignatureCheckMode | Refused: writable record | NotYet: namespace value |
| 88 | `debug.ts:488:35`, formatTypeFacts, TypeFacts | Refused: writable record | NotYet: namespace value |
| 89 | `debug.ts:519:56`, attachFlowNodeDebugInfoWorker, FlowFlags | Refused: writable record | NotYet: namespace value |
| 91 | `emitter.ts:5790:141`, makeName, GeneratedIdentifierFlags | Refused: writable record | NotYet: namespace value |

The 18 debug.ts sites use the full latent census restricted to debug.ts. The nineteenth, emitter.ts:5790:141 in makeName, uses the identical refusal scan without the unrelated emitter unit walk. The scan-only variant matches all 57 before-debug refusal-scan findings exactly. Its no-output guards remain enabled.

The independent emitter single-unit lowering probe stops earlier, at emitter.ts:5790:84 reading autoGenerate, on both trees. The cast-specific scan result does not claim that this earlier boundary has been removed.

Full raw records, exact reasons, phase-specific observations and actual namespace locations are retained in site-stops.json and the bounded piece directories.
