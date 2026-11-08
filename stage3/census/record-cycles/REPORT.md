# Record-cycle census, task a9e21hc

This is the hypothetical with records as nodes and edges to their value types. It is not a count of records selected by today's Go lowering. At [internal/lower/graph_types.go:262](https://github.com/system-inc/adamic/blob/fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d/internal/lower/graph_types.go#L262), only arrays, tuples, maps and sets enter as acyclic caches of graph elements. Records are absent there, and the earlier check near line 87 refuses cycles through records.

The source-checker result below distinguishes ownership cycles proved by fields/elements/value edges from unsure types whose answer depends on lowering information. The full Go component selection cannot be reproduced exactly from the stock TypeScript checker. Treat the unsure column as unresolved exposure, not as zero.

## Findings

| Value-edge cycle type | Unambiguous literal sites | Unsure literal sites | Declaration |
| --- | --- | --- | --- |
| AdamicJsonRecoveryObject | 3 | 1 | AdamicJsonRecoveryObject (src/compiler/commandLineParser.ts:2439:1) |
| Args | 37 | 0 | Args (src/compiler/tracing.ts:54:5) |

| Top file | Evaluation index | Unambiguous sites | Unsure sites |
| --- | --- | --- | --- |
| src/compiler/checker.ts | 33 | 16 | 0 |
| src/compiler/program.ts | 65 | 12 | 0 |
| src/compiler/commandLineParser.ts | 27 | 4 | 1 |
| src/compiler/emitter.ts | 63 | 3 | 0 |
| src/compiler/transformer.ts | 62 | 2 | 0 |
| src/compiler/binder.ts | 29 | 1 | 0 |
| src/compiler/parser.ts | 26 | 1 | 0 |
| src/compiler/tracing.ts | 8 | 1 | 0 |

The full source-view SCC approximation contains 352 source-used indexed/mapped record candidates and 169 unique candidate literal sites. The two value-edge cycle types are a proved subset, not a complete lowered-component count. Structural compatibility through empty objects accounts for the additional source-view exposure, which remains unsure.

## Snapshot and scope

- Adamic origin/main snapshot: `fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d`.
- TypeScript source: `050880ce59e30b356b686bd3144efe24f875ebc8`, version `6.0.3`.
- Adapted with this snapshot's `stage3/apply.sh`; no compiler source changes in this commit.
- Compiler config: `src/compiler/tsconfig.json`, including its inherited options. Root: `src/tsc/tsc.ts`.
- Runtime import/export closure: 81 modules, 250 checker source files including declarations. Type declarations and type nodes in the runtime closure are inventoried too; generic templates stay unsure.
- Runtime source SHA-256: `1d08ef599e3936c7b3735638bbd5cf210c27f602125c42bc5378da3dbbf2b3f4`. Script SHA-256: `7284b1eeed69a01231bf4a23dcefed01a2a3dca04973ad3a77bfea62785acdc6`.
- Evaluation-order indices are zero based, depth-first imports/exports in source order, then the module, with back edges skipped, mirroring `ESMModuleOrder` in `internal/lower/modules.go`.

## Counts as JSON

```json
{
  "totals": {
    "cycle": {
      "recordTypes": 2,
      "unambiguousLiteralTypePairs": 40,
      "unsureLiteralTypePairs": 1,
      "literalTypePairs": 41,
      "uniqueLiteralSites": 41
    },
    "acyclic": {
      "recordTypes": 0,
      "unambiguousLiteralTypePairs": 0,
      "unsureLiteralTypePairs": 0,
      "literalTypePairs": 0,
      "uniqueLiteralSites": 0
    },
    "unsure": {
      "recordTypes": 437,
      "unambiguousLiteralTypePairs": 52,
      "unsureLiteralTypePairs": 150,
      "literalTypePairs": 202,
      "uniqueLiteralSites": 128
    }
  },
  "componentRuleCounts": {
    "sourceViewCycleRecordTypes": 352,
    "sourceViewCycleUniqueLiteralSites": 169,
    "includingReflexiveRelatedEdgesRecordTypes": 352
  },
  "ownershipOnlyAcyclicControl": {
    "recordTypes": 374,
    "uniqueLiteralSites": 128
  },
  "graph": {
    "nodes": 48239,
    "ownershipEdges": 221211,
    "sourceViewSpanningRelations": 5772,
    "relationChecks": 5772,
    "candidates": 5773,
    "identityViewEdgesExcluded": true
  }
}
```

The ownership-cycle records have 40 unambiguous sites and one additional unsure site: `src/compiler/commandLineParser.ts:2501:30` is `{}` in `AdamicJsonRecoveryValue | undefined`, a mixed union whose non-record members prevent reproducing the record-union storage proof. Its three other recovery-object sites and all 37 Args sites have unambiguous record destinations. Counts labelled candidate include that unsure site.

A record type means one resolved checker identity, not one spelling of a generic interface. Distinct instantiations and inferred record types are separate rows. `literalTypePairs` can exceed unique sites when a contextual union supplies multiple possible records; no distribution across those members is asserted. Exact locations, type identities, all per-file totals, diagnostics, and the full literal inventory are in [counts.json](counts.json).

## Source-view hypothetical component records

These are all source-used records in the hypothetical source-view SCCs. The unsure flag identifies results that need lowering information or representation proofs. Value types and short cycle witnesses are included for every row.

```json
[
  {
    "id": 26051,
    "type": "{ [index: string]: string | { value: string; pos: number; end: number; }; }",
    "declaration": "__type (src/compiler/parser.ts:10723:29)",
    "valueType": "string | { value: string; pos: number; end: number; }",
    "cycle": [
      "{ [index: string]: string | { value: string; pos: number; end: number; }; }",
      "{}",
      "{ [index: string]: string | { value: string; pos: number; end: number; }; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/parser.ts",
        "evaluationOrder": 26,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 26077,
    "type": "{ [index: string]: string | undefined; }",
    "declaration": "__type (src/compiler/parser.ts:10781:87)",
    "valueType": "string | undefined",
    "cycle": [
      "{ [index: string]: string | undefined; }",
      "{}",
      "{ [index: string]: string | undefined; }"
    ],
    "literalSites": 2,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 2,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/parser.ts",
        "evaluationOrder": 26,
        "literalCount": 2
      }
    ]
  },
  {
    "id": 26082,
    "type": "{ [index: string]: string | undefined; }",
    "declaration": "__type (src/compiler/parser.ts:10785:19)",
    "valueType": "string | undefined",
    "cycle": [
      "{ [index: string]: string | undefined; }",
      "{}",
      "{ [index: string]: string | undefined; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/parser.ts",
        "evaluationOrder": 26,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 26922,
    "type": "{ [k: string]: CompilerOptionsValue; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2019.object.d.ts:24:73)",
    "valueType": "CompilerOptionsValue",
    "cycle": [
      "{ [k: string]: CompilerOptionsValue; }",
      "{}",
      "{ [k: string]: CompilerOptionsValue; }"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 26842,
    "type": "{ [key: string]: unknown; }",
    "declaration": "__type (src/compiler/commandLineParser.ts:2442:62)",
    "valueType": "unknown",
    "cycle": [
      "{ [key: string]: unknown; }",
      "{}",
      "{ [key: string]: unknown; }"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": []
  },
  {
    "id": 26844,
    "type": "{ [key: string]: unknown; }",
    "declaration": "__type (src/compiler/commandLineParser.ts:2447:58)",
    "valueType": "unknown",
    "cycle": [
      "{ [key: string]: unknown; }",
      "{}",
      "{ [key: string]: unknown; }"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": []
  },
  {
    "id": 44887,
    "type": "{ [s: string]: \"General_Category\" | \"Script\" | \"Script_Extensions\"; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "\"General_Category\" | \"Script\" | \"Script_Extensions\"",
    "cycle": [
      "{ [s: string]: \"General_Category\" | \"Script\" | \"Script_Extensions\"; }",
      "{}",
      "{ [s: string]: \"General_Category\" | \"Script\" | \"Script_Extensions\"; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/scanner.ts",
        "evaluationOrder": 13,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45754,
    "type": "{ [s: string]: ImportsNotUsedAsValues; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "ImportsNotUsedAsValues",
    "cycle": [
      "{ [s: string]: ImportsNotUsedAsValues; }",
      "{}",
      "{ [s: string]: ImportsNotUsedAsValues; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45939,
    "type": "{ [s: string]: IntrinsicTypeKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "IntrinsicTypeKind",
    "cycle": [
      "{ [s: string]: IntrinsicTypeKind; }",
      "{}",
      "{ [s: string]: IntrinsicTypeKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/checker.ts",
        "evaluationOrder": 33,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45730,
    "type": "{ [s: string]: JsxEmit; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "JsxEmit",
    "cycle": [
      "{ [s: string]: JsxEmit; }",
      "{}",
      "{ [s: string]: JsxEmit; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45042,
    "type": "{ [s: string]: Map<string, string[]> | Map<string, never[]> | Map<string, never[] | string[]>; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "Map<string, string[]> | Map<string, never[]> | Map<string, never[] | string[]>",
    "cycle": [
      "{ [s: string]: Map<string, string[]> | Map<string, never[]> | Map<string, never[] | string[]>; }",
      "{}",
      "{ [s: string]: Map<string, string[]> | Map<string, never[]> | Map<string, never[] | string[]>; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45769,
    "type": "{ [s: string]: ModuleDetectionKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "ModuleDetectionKind",
    "cycle": [
      "{ [s: string]: ModuleDetectionKind; }",
      "{}",
      "{ [s: string]: ModuleDetectionKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45749,
    "type": "{ [s: string]: ModuleKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "ModuleKind",
    "cycle": [
      "{ [s: string]: ModuleKind; }",
      "{}",
      "{ [s: string]: ModuleKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45757,
    "type": "{ [s: string]: ModuleResolutionKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "ModuleResolutionKind",
    "cycle": [
      "{ [s: string]: ModuleResolutionKind; }",
      "{}",
      "{ [s: string]: ModuleResolutionKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45052,
    "type": "{ [s: string]: never[] | string[]; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "never[] | string[]",
    "cycle": [
      "{ [s: string]: never[] | string[]; }",
      "{}",
      "{ [s: string]: never[] | string[]; }"
    ],
    "literalSites": 3,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 3,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 3
      }
    ]
  },
  {
    "id": 45048,
    "type": "{ [s: string]: never[]; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "never[]",
    "cycle": [
      "{ [s: string]: never[]; }",
      "{}",
      "{ [s: string]: never[]; }"
    ],
    "literalSites": 10,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 10,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 10
      }
    ]
  },
  {
    "id": 45765,
    "type": "{ [s: string]: NewLineKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "NewLineKind",
    "cycle": [
      "{ [s: string]: NewLineKind; }",
      "{}",
      "{ [s: string]: NewLineKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 47239,
    "type": "{ [s: string]: number; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "number",
    "cycle": [
      "{ [s: string]: number; }",
      "{}",
      "{ [s: string]: number; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/transformers/jsx.ts",
        "evaluationOrder": 52,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45740,
    "type": "{ [s: string]: PollingWatchKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "PollingWatchKind",
    "cycle": [
      "{ [s: string]: PollingWatchKind; }",
      "{}",
      "{ [s: string]: PollingWatchKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45746,
    "type": "{ [s: string]: ScriptTarget; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "ScriptTarget",
    "cycle": [
      "{ [s: string]: ScriptTarget; }",
      "{}",
      "{ [s: string]: ScriptTarget; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45089,
    "type": "{ [s: string]: string; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "string",
    "cycle": [
      "{ [s: string]: string; }",
      "{}",
      "{ [s: string]: string; }"
    ],
    "literalSites": 2,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 2,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 2
      }
    ]
  },
  {
    "id": 45045,
    "type": "{ [s: string]: string[]; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "string[]",
    "cycle": [
      "{ [s: string]: string[]; }",
      "{}",
      "{ [s: string]: string[]; }"
    ],
    "literalSites": 41,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 41,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 41
      }
    ]
  },
  {
    "id": 44879,
    "type": "{ [s: string]: SyntaxKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "SyntaxKind",
    "cycle": [
      "{ [s: string]: SyntaxKind; }",
      "{}",
      "{ [s: string]: SyntaxKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/scanner.ts",
        "evaluationOrder": 13,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45932,
    "type": "{ [s: string]: TypeFacts.TypeofNEString | TypeFacts.TypeofNENumber | TypeFacts.TypeofNEBigInt | TypeFacts.TypeofNEBoolean | TypeFacts.TypeofNESymbol | TypeFacts.TypeofNEObject | TypeFacts.TypeofNEFunction | TypeFacts.NEUndefined; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "TypeFacts.TypeofNEString | TypeFacts.TypeofNENumber | TypeFacts.TypeofNEBigInt | TypeFacts.TypeofNEBoolean | TypeFacts.TypeofNESymbol | TypeFacts.TypeofNEObject | TypeFacts.TypeofNEFunction | TypeFacts.NEUndefined",
    "cycle": [
      "{ [s: string]: TypeFacts.TypeofNEString | TypeFacts.TypeofNENumber | TypeFacts.TypeofNEBigInt | TypeFacts.TypeofNEBoolean | TypeFacts.TypeofNESymbol | TypeFacts.TypeofNEObject | TypeFacts.TypeofNEFunction | TypeFacts.NEUndefined; }",
      "{}",
      "{ [s: string]: TypeFacts.TypeofNEString | TypeFacts.TypeofNENumber | TypeFacts.TypeofNEBigInt | TypeFacts.TypeofNEBoolean | TypeFacts.TypeofNESymbol | TypeFacts.TypeofNEObject | TypeFacts.TypeofNEFunction | TypeFacts.NEUndefined; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/checker.ts",
        "evaluationOrder": 33,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45737,
    "type": "{ [s: string]: WatchDirectoryKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "WatchDirectoryKind",
    "cycle": [
      "{ [s: string]: WatchDirectoryKind; }",
      "{}",
      "{ [s: string]: WatchDirectoryKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45734,
    "type": "{ [s: string]: WatchFileKind; }",
    "declaration": "__type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19)",
    "valueType": "WatchFileKind",
    "cycle": [
      "{ [s: string]: WatchFileKind; }",
      "{}",
      "{ [s: string]: WatchFileKind; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 43674,
    "type": "{ [value: string]: string[]; }",
    "declaration": "__type (src/compiler/executeCommandLine.ts:405:37)",
    "valueType": "string[]",
    "cycle": [
      "{ [value: string]: string[]; }",
      "{}",
      "{ [value: string]: string[]; }"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/executeCommandLine.ts",
        "evaluationOrder": 76,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 15391,
    "type": "{ [x: string]: SyntaxKind.BreakKeyword | SyntaxKind.CaseKeyword | SyntaxKind.CatchKeyword | SyntaxKind.ClassKeyword | SyntaxKind.ConstKeyword | SyntaxKind.ContinueKeyword | SyntaxKind.DebuggerKeyword | SyntaxKind.DefaultKeyword | SyntaxKind.DeleteKeyword | SyntaxKind.DoKeyword | SyntaxKind.ElseKeyword | SyntaxKind.EnumKeyword | SyntaxKind.ExportKeyword | SyntaxKind.ExtendsKeyword | SyntaxKind.FalseKeyword | SyntaxKind.FinallyKeyword | SyntaxKind.ForKeyword | SyntaxKind.FunctionKeyword | SyntaxKind.IfKeyword | SyntaxKind.ImportKeyword | SyntaxKind.InKeyword | SyntaxKind.InstanceOfKeyword | SyntaxKind.NewKeyword | SyntaxKind.NullKeyword | SyntaxKind.ReturnKeyword | SyntaxKind.SuperKeyword | SyntaxKind.SwitchKeyword | SyntaxKind.ThisKeyword | SyntaxKind.ThrowKeyword | SyntaxKind.TrueKeyword | SyntaxKind.TryKeyword | SyntaxKind.TypeOfKeyword | SyntaxKind.VarKeyword | SyntaxKind.VoidKeyword | SyntaxKind.WhileKeyword | SyntaxKind.WithKeyword | SyntaxKind.ImplementsKeyword | SyntaxKind.InterfaceKeyword | SyntaxKind.LetKeyword | SyntaxKind.PackageKeyword | SyntaxKind.PrivateKeyword | SyntaxKind.ProtectedKeyword | SyntaxKind.PublicKeyword | SyntaxKind.StaticKeyword | SyntaxKind.YieldKeyword | SyntaxKind.AbstractKeyword | SyntaxKind.AccessorKeyword | SyntaxKind.AsKeyword | SyntaxKind.AssertsKeyword | SyntaxKind.AssertKeyword | SyntaxKind.AnyKeyword | SyntaxKind.AsyncKeyword | SyntaxKind.AwaitKeyword | SyntaxKind.BooleanKeyword | SyntaxKind.ConstructorKeyword | SyntaxKind.DeclareKeyword | SyntaxKind.GetKeyword | SyntaxKind.InferKeyword | SyntaxKind.IntrinsicKeyword | SyntaxKind.IsKeyword | SyntaxKind.KeyOfKeyword | SyntaxKind.ModuleKeyword | SyntaxKind.NamespaceKeyword | SyntaxKind.NeverKeyword | SyntaxKind.OutKeyword | SyntaxKind.ReadonlyKeyword | SyntaxKind.RequireKeyword | SyntaxKind.NumberKeyword | SyntaxKind.ObjectKeyword | SyntaxKind.SatisfiesKeyword | SyntaxKind.SetKeyword | SyntaxKind.StringKeyword | SyntaxKind.SymbolKeyword | SyntaxKind.TypeKeyword | SyntaxKind.UndefinedKeyword | SyntaxKind.UniqueKeyword | SyntaxKind.UnknownKeyword | SyntaxKind.UsingKeyword | SyntaxKind.FromKeyword | SyntaxKind.GlobalKeyword | SyntaxKind.BigIntKeyword | SyntaxKind.OverrideKeyword | SyntaxKind.OfKeyword | SyntaxKind.DeferKeyword; abstract: SyntaxKind.AbstractKeyword; accessor: SyntaxKind.AccessorKeyword; any: SyntaxKind.AnyKeyword; as: SyntaxKind.AsKeyword; asserts: SyntaxKind.AssertsKeyword; assert: SyntaxKind.AssertKeyword; bigint: SyntaxKind.BigIntKeyword; boolean: SyntaxKind.BooleanKeyword; break: SyntaxKind.BreakKeyword; case: SyntaxKind.CaseKeyword; catch: SyntaxKind.CatchKeyword; class: SyntaxKind.ClassKeyword; continue: SyntaxKind.ContinueKeyword; const: SyntaxKind.ConstKeyword; debugger: SyntaxKind.DebuggerKeyword; declare: SyntaxKind.DeclareKeyword; default: SyntaxKind.DefaultKeyword; defer: SyntaxKind.DeferKeyword; delete: SyntaxKind.DeleteKeyword; do: SyntaxKind.DoKeyword; else: SyntaxKind.ElseKeyword; enum: SyntaxKind.EnumKeyword; export: SyntaxKind.ExportKeyword; extends: SyntaxKind.ExtendsKeyword; false: SyntaxKind.FalseKeyword; finally: SyntaxKind.FinallyKeyword; for: SyntaxKind.ForKeyword; from: SyntaxKind.FromKeyword; function: SyntaxKind.FunctionKeyword; get: SyntaxKind.GetKeyword; if: SyntaxKind.IfKeyword; implements: SyntaxKind.ImplementsKeyword; import: SyntaxKind.ImportKeyword; in: SyntaxKind.InKeyword; infer: SyntaxKind.InferKeyword; instanceof: SyntaxKind.InstanceOfKeyword; interface: SyntaxKind.InterfaceKeyword; intrinsic: SyntaxKind.IntrinsicKeyword; is: SyntaxKind.IsKeyword; keyof: SyntaxKind.KeyOfKeyword; let: SyntaxKind.LetKeyword; module: SyntaxKind.ModuleKeyword; namespace: SyntaxKind.NamespaceKeyword; never: SyntaxKind.NeverKeyword; new: SyntaxKind.NewKeyword; null: SyntaxKind.NullKeyword; number: SyntaxKind.NumberKeyword; object: SyntaxKind.ObjectKeyword; package: SyntaxKind.PackageKeyword; private: SyntaxKind.PrivateKeyword; protected: SyntaxKind.ProtectedKeyword; public: SyntaxKind.PublicKeyword; override: SyntaxKind.OverrideKeyword; out: SyntaxKind.OutKeyword; readonly: SyntaxKind.ReadonlyKeyword; require: SyntaxKind.RequireKeyword; global: SyntaxKind.GlobalKeyword; return: SyntaxKind.ReturnKeyword; satisfies: SyntaxKind.SatisfiesKeyword; set: SyntaxKind.SetKeyword; static: SyntaxKind.StaticKeyword; string: SyntaxKind.StringKeyword; super: SyntaxKind.SuperKeyword; switch: SyntaxKind.SwitchKeyword; symbol: SyntaxKind.SymbolKeyword; this: SyntaxKind.ThisKeyword; throw: SyntaxKind.ThrowKeyword; true: SyntaxKind.TrueKeyword; try: SyntaxKind.TryKeyword; type: SyntaxKind.TypeKeyword; typeof: SyntaxKind.TypeOfKeyword; undefined: SyntaxKind.UndefinedKeyword; unique: SyntaxKind.UniqueKeyword; unknown: SyntaxKind.UnknownKeyword; using: SyntaxKind.UsingKeyword; var: SyntaxKind.VarKeyword; void: SyntaxKind.VoidKeyword; while: SyntaxKind.WhileKeyword; with: SyntaxKind.WithKeyword; yield: SyntaxKind.YieldKeyword; async: SyntaxKind.AsyncKeyword; await: SyntaxKind.AwaitKeyword; of: SyntaxKind.OfKeyword; }",
    "declaration": "__object (src/compiler/scanner.ts:135:61)",
    "valueType": "SyntaxKind.BreakKeyword | SyntaxKind.CaseKeyword | SyntaxKind.CatchKeyword | SyntaxKind.ClassKeyword | SyntaxKind.ConstKeyword | SyntaxKind.ContinueKeyword | SyntaxKind.DebuggerKeyword | SyntaxKind.DefaultKeyword | SyntaxKind.DeleteKeyword | SyntaxKind.DoKeyword | SyntaxKind.ElseKeyword | SyntaxKind.EnumKeyword | SyntaxKind.ExportKeyword | SyntaxKind.ExtendsKeyword | SyntaxKind.FalseKeyword | SyntaxKind.FinallyKeyword | SyntaxKind.ForKeyword | SyntaxKind.FunctionKeyword | SyntaxKind.IfKeyword | SyntaxKind.ImportKeyword | SyntaxKind.InKeyword | SyntaxKind.InstanceOfKeyword | SyntaxKind.NewKeyword | SyntaxKind.NullKeyword | SyntaxKind.ReturnKeyword | SyntaxKind.SuperKeyword | SyntaxKind.SwitchKeyword | SyntaxKind.ThisKeyword | SyntaxKind.ThrowKeyword | SyntaxKind.TrueKeyword | SyntaxKind.TryKeyword | SyntaxKind.TypeOfKeyword | SyntaxKind.VarKeyword | SyntaxKind.VoidKeyword | SyntaxKind.WhileKeyword | SyntaxKind.WithKeyword | SyntaxKind.ImplementsKeyword | SyntaxKind.InterfaceKeyword | SyntaxKind.LetKeyword | SyntaxKind.PackageKeyword | SyntaxKind.PrivateKeyword | SyntaxKind.ProtectedKeyword | SyntaxKind.PublicKeyword | SyntaxKind.StaticKeyword | SyntaxKind.YieldKeyword | SyntaxKind.AbstractKeyword | SyntaxKind.AccessorKeyword | SyntaxKind.AsKeyword | SyntaxKind.AssertsKeyword | SyntaxKind.AssertKeyword | SyntaxKind.AnyKeyword | SyntaxKind.AsyncKeyword | SyntaxKind.AwaitKeyword | SyntaxKind.BooleanKeyword | SyntaxKind.ConstructorKeyword | SyntaxKind.DeclareKeyword | SyntaxKind.GetKeyword | SyntaxKind.InferKeyword | SyntaxKind.IntrinsicKeyword | SyntaxKind.IsKeyword | SyntaxKind.KeyOfKeyword | SyntaxKind.ModuleKeyword | SyntaxKind.NamespaceKeyword | SyntaxKind.NeverKeyword | SyntaxKind.OutKeyword | SyntaxKind.ReadonlyKeyword | SyntaxKind.RequireKeyword | SyntaxKind.NumberKeyword | SyntaxKind.ObjectKeyword | SyntaxKind.SatisfiesKeyword | SyntaxKind.SetKeyword | SyntaxKind.StringKeyword | SyntaxKind.SymbolKeyword | SyntaxKind.TypeKeyword | SyntaxKind.UndefinedKeyword | SyntaxKind.UniqueKeyword | SyntaxKind.UnknownKeyword | SyntaxKind.UsingKeyword | SyntaxKind.FromKeyword | SyntaxKind.GlobalKeyword | SyntaxKind.BigIntKeyword | SyntaxKind.OverrideKeyword | SyntaxKind.OfKeyword | SyntaxKind.DeferKeyword",
    "cycle": [
      "{ [x: string]: SyntaxKind.BreakKeyword | SyntaxKind.CaseKeyword | SyntaxKind.CatchKeyword | SyntaxKind.ClassKeyword | SyntaxKind.ConstKeyword | SyntaxKind.ContinueKeyword | SyntaxKind.DebuggerKeyword | SyntaxKind.DefaultKeyword | SyntaxKind.DeleteKeyword | SyntaxKind.DoKeyword | SyntaxKind.ElseKeyword | SyntaxKind.EnumKeyword | SyntaxKind.ExportKeyword | SyntaxKind.ExtendsKeyword | SyntaxKind.FalseKeyword | SyntaxKind.FinallyKeyword | SyntaxKind.ForKeyword | SyntaxKind.FunctionKeyword | SyntaxKind.IfKeyword | SyntaxKind.ImportKeyword | SyntaxKind.InKeyword | SyntaxKind.InstanceOfKeyword | SyntaxKind.NewKeyword | SyntaxKind.NullKeyword | SyntaxKind.ReturnKeyword | SyntaxKind.SuperKeyword | SyntaxKind.SwitchKeyword | SyntaxKind.ThisKeyword | SyntaxKind.ThrowKeyword | SyntaxKind.TrueKeyword | SyntaxKind.TryKeyword | SyntaxKind.TypeOfKeyword | SyntaxKind.VarKeyword | SyntaxKind.VoidKeyword | SyntaxKind.WhileKeyword | SyntaxKind.WithKeyword | SyntaxKind.ImplementsKeyword | SyntaxKind.InterfaceKeyword | SyntaxKind.LetKeyword | SyntaxKind.PackageKeyword | SyntaxKind.PrivateKeyword | SyntaxKind.ProtectedKeyword | SyntaxKind.PublicKeyword | SyntaxKind.StaticKeyword | SyntaxKind.YieldKeyword | SyntaxKind.AbstractKeyword | SyntaxKind.AccessorKeyword | SyntaxKind.AsKeyword | SyntaxKind.AssertsKeyword | SyntaxKind.AssertKeyword | SyntaxKind.AnyKeyword | SyntaxKind.AsyncKeyword | SyntaxKind.AwaitKeyword | SyntaxKind.BooleanKeyword | SyntaxKind.ConstructorKeyword | SyntaxKind.DeclareKeyword | SyntaxKind.GetKeyword | SyntaxKind.InferKeyword | SyntaxKind.IntrinsicKeyword | SyntaxKind.IsKeyword | SyntaxKind.KeyOfKeyword | SyntaxKind.ModuleKeyword | SyntaxKind.NamespaceKeyword | SyntaxKind.NeverKeyword | SyntaxKind.OutKeyword | SyntaxKind.ReadonlyKeyword | SyntaxKind.RequireKeyword | SyntaxKind.NumberKeyword | SyntaxKind.ObjectKeyword | SyntaxKind.SatisfiesKeyword | SyntaxKind.SetKeyword | SyntaxKind.StringKeyword | SyntaxKind.SymbolKeyword | SyntaxKind.TypeKeyword | SyntaxKind.UndefinedKeyword | SyntaxKind.UniqueKeyword | SyntaxKind.UnknownKeyword | SyntaxKind.UsingKeyword | SyntaxKind.FromKeyword | SyntaxKind.GlobalKeyword | SyntaxKind.BigIntKeyword | SyntaxKind.OverrideKeyword | SyntaxKind.OfKeyword | SyntaxKind.DeferKeyword; abstract: SyntaxKind.AbstractKeyword; accessor: SyntaxKind.AccessorKeyword; any: SyntaxKind.AnyKeyword; as: SyntaxKind.AsKeyword; asserts: SyntaxKind.AssertsKeyword; assert: SyntaxKind.AssertKeyword; bigint: SyntaxKind.BigIntKeyword; boolean: SyntaxKind.BooleanKeyword; break: SyntaxKind.BreakKeyword; case: SyntaxKind.CaseKeyword; catch: SyntaxKind.CatchKeyword; class: SyntaxKind.ClassKeyword; continue: SyntaxKind.ContinueKeyword; const: SyntaxKind.ConstKeyword; debugger: SyntaxKind.DebuggerKeyword; declare: SyntaxKind.DeclareKeyword; default: SyntaxKind.DefaultKeyword; defer: SyntaxKind.DeferKeyword; delete: SyntaxKind.DeleteKeyword; do: SyntaxKind.DoKeyword; else: SyntaxKind.ElseKeyword; enum: SyntaxKind.EnumKeyword; export: SyntaxKind.ExportKeyword; extends: SyntaxKind.ExtendsKeyword; false: SyntaxKind.FalseKeyword; finally: SyntaxKind.FinallyKeyword; for: SyntaxKind.ForKeyword; from: SyntaxKind.FromKeyword; function: SyntaxKind.FunctionKeyword; get: SyntaxKind.GetKeyword; if: SyntaxKind.IfKeyword; implements: SyntaxKind.ImplementsKeyword; import: SyntaxKind.ImportKeyword; in: SyntaxKind.InKeyword; infer: SyntaxKind.InferKeyword; instanceof: SyntaxKind.InstanceOfKeyword; interface: SyntaxKind.InterfaceKeyword; intrinsic: SyntaxKind.IntrinsicKeyword; is: SyntaxKind.IsKeyword; keyof: SyntaxKind.KeyOfKeyword; let: SyntaxKind.LetKeyword; module: SyntaxKind.ModuleKeyword; namespace: SyntaxKind.NamespaceKeyword; never: SyntaxKind.NeverKeyword; new: SyntaxKind.NewKeyword; null: SyntaxKind.NullKeyword; number: SyntaxKind.NumberKeyword; object: SyntaxKind.ObjectKeyword; package: SyntaxKind.PackageKeyword; private: SyntaxKind.PrivateKeyword; protected: SyntaxKind.ProtectedKeyword; public: SyntaxKind.PublicKeyword; override: SyntaxKind.OverrideKeyword; out: SyntaxKind.OutKeyword; readonly: SyntaxKind.ReadonlyKeyword; require: SyntaxKind.RequireKeyword; global: SyntaxKind.GlobalKeyword; return: SyntaxKind.ReturnKeyword; satisfies: SyntaxKind.SatisfiesKeyword; set: SyntaxKind.SetKeyword; static: SyntaxKind.StaticKeyword; string: SyntaxKind.StringKeyword; super: SyntaxKind.SuperKeyword; switch: SyntaxKind.SwitchKeyword; symbol: SyntaxKind.SymbolKeyword; this: SyntaxKind.ThisKeyword; throw: SyntaxKind.ThrowKeyword; true: SyntaxKind.TrueKeyword; try: SyntaxKind.TryKeyword; type: SyntaxKind.TypeKeyword; typeof: SyntaxKind.TypeOfKeyword; undefined: SyntaxKind.UndefinedKeyword; unique: SyntaxKind.UniqueKeyword; unknown: SyntaxKind.UnknownKeyword; using: SyntaxKind.UsingKeyword; var: SyntaxKind.VarKeyword; void: SyntaxKind.VoidKeyword; while: SyntaxKind.WhileKeyword; with: SyntaxKind.WithKeyword; yield: SyntaxKind.YieldKeyword; async: SyntaxKind.AsyncKeyword; await: SyntaxKind.AwaitKeyword; of: SyntaxKind.OfKeyword; }",
      "{}",
      "{ [x: string]: SyntaxKind.BreakKeyword | SyntaxKind.CaseKeyword | SyntaxKind.CatchKeyword | SyntaxKind.ClassKeyword | SyntaxKind.ConstKeyword | SyntaxKind.ContinueKeyword | SyntaxKind.DebuggerKeyword | SyntaxKind.DefaultKeyword | SyntaxKind.DeleteKeyword | SyntaxKind.DoKeyword | SyntaxKind.ElseKeyword | SyntaxKind.EnumKeyword | SyntaxKind.ExportKeyword | SyntaxKind.ExtendsKeyword | SyntaxKind.FalseKeyword | SyntaxKind.FinallyKeyword | SyntaxKind.ForKeyword | SyntaxKind.FunctionKeyword | SyntaxKind.IfKeyword | SyntaxKind.ImportKeyword | SyntaxKind.InKeyword | SyntaxKind.InstanceOfKeyword | SyntaxKind.NewKeyword | SyntaxKind.NullKeyword | SyntaxKind.ReturnKeyword | SyntaxKind.SuperKeyword | SyntaxKind.SwitchKeyword | SyntaxKind.ThisKeyword | SyntaxKind.ThrowKeyword | SyntaxKind.TrueKeyword | SyntaxKind.TryKeyword | SyntaxKind.TypeOfKeyword | SyntaxKind.VarKeyword | SyntaxKind.VoidKeyword | SyntaxKind.WhileKeyword | SyntaxKind.WithKeyword | SyntaxKind.ImplementsKeyword | SyntaxKind.InterfaceKeyword | SyntaxKind.LetKeyword | SyntaxKind.PackageKeyword | SyntaxKind.PrivateKeyword | SyntaxKind.ProtectedKeyword | SyntaxKind.PublicKeyword | SyntaxKind.StaticKeyword | SyntaxKind.YieldKeyword | SyntaxKind.AbstractKeyword | SyntaxKind.AccessorKeyword | SyntaxKind.AsKeyword | SyntaxKind.AssertsKeyword | SyntaxKind.AssertKeyword | SyntaxKind.AnyKeyword | SyntaxKind.AsyncKeyword | SyntaxKind.AwaitKeyword | SyntaxKind.BooleanKeyword | SyntaxKind.ConstructorKeyword | SyntaxKind.DeclareKeyword | SyntaxKind.GetKeyword | SyntaxKind.InferKeyword | SyntaxKind.IntrinsicKeyword | SyntaxKind.IsKeyword | SyntaxKind.KeyOfKeyword | SyntaxKind.ModuleKeyword | SyntaxKind.NamespaceKeyword | SyntaxKind.NeverKeyword | SyntaxKind.OutKeyword | SyntaxKind.ReadonlyKeyword | SyntaxKind.RequireKeyword | SyntaxKind.NumberKeyword | SyntaxKind.ObjectKeyword | SyntaxKind.SatisfiesKeyword | SyntaxKind.SetKeyword | SyntaxKind.StringKeyword | SyntaxKind.SymbolKeyword | SyntaxKind.TypeKeyword | SyntaxKind.UndefinedKeyword | SyntaxKind.UniqueKeyword | SyntaxKind.UnknownKeyword | SyntaxKind.UsingKeyword | SyntaxKind.FromKeyword | SyntaxKind.GlobalKeyword | SyntaxKind.BigIntKeyword | SyntaxKind.OverrideKeyword | SyntaxKind.OfKeyword | SyntaxKind.DeferKeyword; abstract: SyntaxKind.AbstractKeyword; accessor: SyntaxKind.AccessorKeyword; any: SyntaxKind.AnyKeyword; as: SyntaxKind.AsKeyword; asserts: SyntaxKind.AssertsKeyword; assert: SyntaxKind.AssertKeyword; bigint: SyntaxKind.BigIntKeyword; boolean: SyntaxKind.BooleanKeyword; break: SyntaxKind.BreakKeyword; case: SyntaxKind.CaseKeyword; catch: SyntaxKind.CatchKeyword; class: SyntaxKind.ClassKeyword; continue: SyntaxKind.ContinueKeyword; const: SyntaxKind.ConstKeyword; debugger: SyntaxKind.DebuggerKeyword; declare: SyntaxKind.DeclareKeyword; default: SyntaxKind.DefaultKeyword; defer: SyntaxKind.DeferKeyword; delete: SyntaxKind.DeleteKeyword; do: SyntaxKind.DoKeyword; else: SyntaxKind.ElseKeyword; enum: SyntaxKind.EnumKeyword; export: SyntaxKind.ExportKeyword; extends: SyntaxKind.ExtendsKeyword; false: SyntaxKind.FalseKeyword; finally: SyntaxKind.FinallyKeyword; for: SyntaxKind.ForKeyword; from: SyntaxKind.FromKeyword; function: SyntaxKind.FunctionKeyword; get: SyntaxKind.GetKeyword; if: SyntaxKind.IfKeyword; implements: SyntaxKind.ImplementsKeyword; import: SyntaxKind.ImportKeyword; in: SyntaxKind.InKeyword; infer: SyntaxKind.InferKeyword; instanceof: SyntaxKind.InstanceOfKeyword; interface: SyntaxKind.InterfaceKeyword; intrinsic: SyntaxKind.IntrinsicKeyword; is: SyntaxKind.IsKeyword; keyof: SyntaxKind.KeyOfKeyword; let: SyntaxKind.LetKeyword; module: SyntaxKind.ModuleKeyword; namespace: SyntaxKind.NamespaceKeyword; never: SyntaxKind.NeverKeyword; new: SyntaxKind.NewKeyword; null: SyntaxKind.NullKeyword; number: SyntaxKind.NumberKeyword; object: SyntaxKind.ObjectKeyword; package: SyntaxKind.PackageKeyword; private: SyntaxKind.PrivateKeyword; protected: SyntaxKind.ProtectedKeyword; public: SyntaxKind.PublicKeyword; override: SyntaxKind.OverrideKeyword; out: SyntaxKind.OutKeyword; readonly: SyntaxKind.ReadonlyKeyword; require: SyntaxKind.RequireKeyword; global: SyntaxKind.GlobalKeyword; return: SyntaxKind.ReturnKeyword; satisfies: SyntaxKind.SatisfiesKeyword; set: SyntaxKind.SetKeyword; static: SyntaxKind.StaticKeyword; string: SyntaxKind.StringKeyword; super: SyntaxKind.SuperKeyword; switch: SyntaxKind.SwitchKeyword; symbol: SyntaxKind.SymbolKeyword; this: SyntaxKind.ThisKeyword; throw: SyntaxKind.ThrowKeyword; true: SyntaxKind.TrueKeyword; try: SyntaxKind.TryKeyword; type: SyntaxKind.TypeKeyword; typeof: SyntaxKind.TypeOfKeyword; undefined: SyntaxKind.UndefinedKeyword; unique: SyntaxKind.UniqueKeyword; unknown: SyntaxKind.UnknownKeyword; using: SyntaxKind.UsingKeyword; var: SyntaxKind.VarKeyword; void: SyntaxKind.VoidKeyword; while: SyntaxKind.WhileKeyword; with: SyntaxKind.WithKeyword; yield: SyntaxKind.YieldKeyword; async: SyntaxKind.AsyncKeyword; await: SyntaxKind.AwaitKeyword; of: SyntaxKind.OfKeyword; }"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "enumerationRecordStorageType needs the lowerer enumeration provenance proof",
      "cycle uses source structural-view candidate approximation",
      "enumerationRecordStorageType needs the lowerer enumeration provenance proof"
    ],
    "files": []
  },
  {
    "id": 26785,
    "type": "AdamicJsonRecoveryObject",
    "declaration": "AdamicJsonRecoveryObject (src/compiler/commandLineParser.ts:2439:1)",
    "valueType": "AdamicJsonRecoveryValue | undefined",
    "cycle": [
      "AdamicJsonRecoveryObject",
      "AdamicJsonRecoveryValue | undefined",
      "AdamicJsonRecoveryObject"
    ],
    "literalSites": 4,
    "unambiguousLiteralSites": 3,
    "unsureLiteralSites": 1,
    "unsure": false,
    "unsureReasons": [],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 4
      }
    ]
  },
  {
    "id": 2675,
    "type": "Args",
    "declaration": "Args (src/compiler/tracing.ts:54:5)",
    "valueType": "string | number | boolean | Args | readonly (string | number | boolean | Args | null | undefined)[] | null | undefined",
    "cycle": [
      "Args",
      "string | number | boolean | Args | readonly (string | number | boolean | Args | null | undefined)[] | null | undefined",
      "Args"
    ],
    "literalSites": 37,
    "unambiguousLiteralSites": 37,
    "unsureLiteralSites": 0,
    "unsure": false,
    "unsureReasons": [],
    "files": [
      {
        "file": "src/compiler/checker.ts",
        "evaluationOrder": 33,
        "literalCount": 16
      },
      {
        "file": "src/compiler/program.ts",
        "evaluationOrder": 65,
        "literalCount": 12
      },
      {
        "file": "src/compiler/emitter.ts",
        "evaluationOrder": 63,
        "literalCount": 3
      },
      {
        "file": "src/compiler/transformer.ts",
        "evaluationOrder": 62,
        "literalCount": 2
      },
      {
        "file": "src/compiler/binder.ts",
        "evaluationOrder": 29,
        "literalCount": 1
      },
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      },
      {
        "file": "src/compiler/parser.ts",
        "evaluationOrder": 26,
        "literalCount": 1
      },
      {
        "file": "src/compiler/tracing.ts",
        "evaluationOrder": 8,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 44888,
    "type": "ArrayLike<\"General_Category\" | \"Script\" | \"Script_Extensions\">",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "\"General_Category\" | \"Script\" | \"Script_Extensions\"",
    "cycle": [
      "ArrayLike<\"General_Category\" | \"Script\" | \"Script_Extensions\">",
      "{}",
      "ArrayLike<\"General_Category\" | \"Script\" | \"Script_Extensions\">"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/scanner.ts",
        "evaluationOrder": 13,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45755,
    "type": "ArrayLike<ImportsNotUsedAsValues>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "ImportsNotUsedAsValues",
    "cycle": [
      "ArrayLike<ImportsNotUsedAsValues>",
      "{}",
      "ArrayLike<ImportsNotUsedAsValues>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45940,
    "type": "ArrayLike<IntrinsicTypeKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "IntrinsicTypeKind",
    "cycle": [
      "ArrayLike<IntrinsicTypeKind>",
      "{}",
      "ArrayLike<IntrinsicTypeKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/checker.ts",
        "evaluationOrder": 33,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45731,
    "type": "ArrayLike<JsxEmit>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "JsxEmit",
    "cycle": [
      "ArrayLike<JsxEmit>",
      "{}",
      "ArrayLike<JsxEmit>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45043,
    "type": "ArrayLike<Map<string, string[]> | Map<string, never[]> | Map<string, never[] | string[]>>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "Map<string, string[]> | Map<string, never[]> | Map<string, never[] | string[]>",
    "cycle": [
      "ArrayLike<Map<string, string[]> | Map<string, never[]> | Map<string, never[] | string[]>>",
      "{}",
      "ArrayLike<Map<string, string[]> | Map<string, never[]> | Map<string, never[] | string[]>>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45770,
    "type": "ArrayLike<ModuleDetectionKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "ModuleDetectionKind",
    "cycle": [
      "ArrayLike<ModuleDetectionKind>",
      "{}",
      "ArrayLike<ModuleDetectionKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45750,
    "type": "ArrayLike<ModuleKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "ModuleKind",
    "cycle": [
      "ArrayLike<ModuleKind>",
      "{}",
      "ArrayLike<ModuleKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45758,
    "type": "ArrayLike<ModuleResolutionKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "ModuleResolutionKind",
    "cycle": [
      "ArrayLike<ModuleResolutionKind>",
      "{}",
      "ArrayLike<ModuleResolutionKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45053,
    "type": "ArrayLike<never[] | string[]>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "never[] | string[]",
    "cycle": [
      "ArrayLike<never[] | string[]>",
      "{}",
      "ArrayLike<never[] | string[]>"
    ],
    "literalSites": 3,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 3,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 3
      }
    ]
  },
  {
    "id": 45049,
    "type": "ArrayLike<never[]>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "never[]",
    "cycle": [
      "ArrayLike<never[]>",
      "{}",
      "ArrayLike<never[]>"
    ],
    "literalSites": 10,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 10,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 10
      }
    ]
  },
  {
    "id": 45766,
    "type": "ArrayLike<NewLineKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "NewLineKind",
    "cycle": [
      "ArrayLike<NewLineKind>",
      "{}",
      "ArrayLike<NewLineKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 47238,
    "type": "ArrayLike<number>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "number",
    "cycle": [
      "ArrayLike<number>",
      "{}",
      "ArrayLike<number>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/transformers/jsx.ts",
        "evaluationOrder": 52,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45741,
    "type": "ArrayLike<PollingWatchKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "PollingWatchKind",
    "cycle": [
      "ArrayLike<PollingWatchKind>",
      "{}",
      "ArrayLike<PollingWatchKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45747,
    "type": "ArrayLike<ScriptTarget>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "ScriptTarget",
    "cycle": [
      "ArrayLike<ScriptTarget>",
      "{}",
      "ArrayLike<ScriptTarget>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45046,
    "type": "ArrayLike<string[]>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "string[]",
    "cycle": [
      "ArrayLike<string[]>",
      "{}",
      "ArrayLike<string[]>"
    ],
    "literalSites": 41,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 41,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 41
      }
    ]
  },
  {
    "id": 45090,
    "type": "ArrayLike<string>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "string",
    "cycle": [
      "ArrayLike<string>",
      "{}",
      "ArrayLike<string>"
    ],
    "literalSites": 2,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 2,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/utilities.ts",
        "evaluationOrder": 15,
        "literalCount": 2
      }
    ]
  },
  {
    "id": 44880,
    "type": "ArrayLike<SyntaxKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "SyntaxKind",
    "cycle": [
      "ArrayLike<SyntaxKind>",
      "{}",
      "ArrayLike<SyntaxKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/scanner.ts",
        "evaluationOrder": 13,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45933,
    "type": "ArrayLike<TypeFacts.TypeofNEString | TypeFacts.TypeofNENumber | TypeFacts.TypeofNEBigInt | TypeFacts.TypeofNEBoolean | TypeFacts.TypeofNESymbol | TypeFacts.TypeofNEObject | TypeFacts.TypeofNEFunction | TypeFacts.NEUndefined>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "TypeFacts.TypeofNEString | TypeFacts.TypeofNENumber | TypeFacts.TypeofNEBigInt | TypeFacts.TypeofNEBoolean | TypeFacts.TypeofNESymbol | TypeFacts.TypeofNEObject | TypeFacts.TypeofNEFunction | TypeFacts.NEUndefined",
    "cycle": [
      "ArrayLike<TypeFacts.TypeofNEString | TypeFacts.TypeofNENumber | TypeFacts.TypeofNEBigInt | TypeFacts.TypeofNEBoolean | TypeFacts.TypeofNESymbol | TypeFacts.TypeofNEObject | TypeFacts.TypeofNEFunction | TypeFacts.NEUndefined>",
      "{}",
      "ArrayLike<TypeFacts.TypeofNEString | TypeFacts.TypeofNENumber | TypeFacts.TypeofNEBigInt | TypeFacts.TypeofNEBoolean | TypeFacts.TypeofNESymbol | TypeFacts.TypeofNEObject | TypeFacts.TypeofNEFunction | TypeFacts.NEUndefined>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/checker.ts",
        "evaluationOrder": 33,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45738,
    "type": "ArrayLike<WatchDirectoryKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "WatchDirectoryKind",
    "cycle": [
      "ArrayLike<WatchDirectoryKind>",
      "{}",
      "ArrayLike<WatchDirectoryKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 45735,
    "type": "ArrayLike<WatchFileKind>",
    "declaration": "ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1)",
    "valueType": "WatchFileKind",
    "cycle": [
      "ArrayLike<WatchFileKind>",
      "{}",
      "ArrayLike<WatchFileKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 8717,
    "type": "Buffer<ArrayBuffer>",
    "declaration": "Buffer (external/node_modules/@types/node/buffer.buffer.d.ts:356:9); Buffer (external/node_modules/@types/node/buffer.d.ts:255:9); Buffer (external/node_modules/@types/node/buffer.d.ts:1771:13)",
    "valueType": "number",
    "cycle": [
      "Buffer<ArrayBuffer>",
      "{}",
      "Buffer<ArrayBuffer>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 8774,
    "type": "Buffer<ArrayBufferLike>",
    "declaration": "Buffer (external/node_modules/@types/node/buffer.buffer.d.ts:356:9); Buffer (external/node_modules/@types/node/buffer.d.ts:255:9); Buffer (external/node_modules/@types/node/buffer.d.ts:1771:13)",
    "valueType": "number",
    "cycle": [
      "Buffer<ArrayBufferLike>",
      "{}",
      "Buffer<ArrayBufferLike>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 26764,
    "type": "BuildOptions",
    "declaration": "BuildOptions (src/compiler/tsbuildPublic.ts:135:1)",
    "valueType": "CompilerOptionsValue",
    "cycle": [
      "BuildOptions",
      "{}",
      "BuildOptions"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": []
  },
  {
    "id": 4596,
    "type": "CompilerOptions",
    "declaration": "CompilerOptions (src/compiler/types.ts:7414:1)",
    "valueType": "TsConfigSourceFile | CompilerOptionsValue",
    "cycle": [
      "CompilerOptions",
      "{}",
      "CompilerOptions"
    ],
    "literalSites": 25,
    "unambiguousLiteralSites": 23,
    "unsureLiteralSites": 2,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 10
      },
      {
        "file": "src/compiler/builder.ts",
        "evaluationOrder": 69,
        "literalCount": 6
      },
      {
        "file": "src/compiler/moduleNameResolver.ts",
        "evaluationOrder": 28,
        "literalCount": 3
      },
      {
        "file": "src/compiler/moduleSpecifiers.ts",
        "evaluationOrder": 31,
        "literalCount": 2
      },
      {
        "file": "src/compiler/tsbuildPublic.ts",
        "evaluationOrder": 75,
        "literalCount": 2
      },
      {
        "file": "src/compiler/transformer.ts",
        "evaluationOrder": 62,
        "literalCount": 1
      },
      {
        "file": "src/compiler/watchPublic.ts",
        "evaluationOrder": 73,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 871,
    "type": "IArguments",
    "declaration": "IArguments (external/node_modules/typescript/lib/lib.es5.d.ts:402:1); IArguments (external/node_modules/typescript/lib/lib.es2015.iterable.d.ts:130:1)",
    "valueType": "any",
    "cycle": [
      "IArguments",
      "{}",
      "IArguments"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "reachable generic/any template requires lowering instantiation or representation proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 3605,
    "type": "JSDocArray",
    "declaration": "JSDocArray (src/compiler/types.ts:963:1)",
    "valueType": "JSDoc",
    "cycle": [
      "JSDocArray",
      "{}",
      "JSDocArray"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 15390,
    "type": "MapLike<KeywordSyntaxKind>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "KeywordSyntaxKind",
    "cycle": [
      "MapLike<KeywordSyntaxKind>",
      "{}",
      "MapLike<KeywordSyntaxKind>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/scanner.ts",
        "evaluationOrder": 13,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 27317,
    "type": "MapLike<MapLike<string[]>>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "MapLike<string[]>",
    "cycle": [
      "MapLike<MapLike<string[]>>",
      "{}",
      "MapLike<MapLike<string[]>>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": []
  },
  {
    "id": 28601,
    "type": "MapLike<readonly string[]>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "readonly string[]",
    "cycle": [
      "MapLike<readonly string[]>",
      "{}",
      "MapLike<readonly string[]>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": []
  },
  {
    "id": 27007,
    "type": "MapLike<string | number | boolean | string[] | (string | number)[] | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "string | number | boolean | string[] | (string | number)[] | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined",
    "cycle": [
      "MapLike<string | number | boolean | string[] | (string | number)[] | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined>",
      "{}",
      "MapLike<string | number | boolean | string[] | (string | number)[] | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": []
  },
  {
    "id": 27002,
    "type": "MapLike<string | number | boolean | string[] | (string | number)[] | TsConfigSourceFile | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "string | number | boolean | string[] | (string | number)[] | TsConfigSourceFile | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined",
    "cycle": [
      "MapLike<string | number | boolean | string[] | (string | number)[] | TsConfigSourceFile | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined>",
      "{}",
      "MapLike<string | number | boolean | string[] | (string | number)[] | TsConfigSourceFile | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 6012,
    "type": "MapLike<string[]>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "string[]",
    "cycle": [
      "MapLike<string[]>",
      "{}",
      "MapLike<string[]>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 20620,
    "type": "MapLike<string>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "string",
    "cycle": [
      "MapLike<string>",
      "{}",
      "MapLike<string>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": []
  },
  {
    "id": 28382,
    "type": "MapLike<Symbol>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "Symbol",
    "cycle": [
      "MapLike<Symbol>",
      "{}",
      "MapLike<Symbol>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/symbolWalker.ts",
        "evaluationOrder": 30,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 1089,
    "type": "MapLike<T | undefined>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "T | undefined",
    "cycle": [
      "MapLike<T | undefined>",
      "{}",
      "MapLike<T | undefined>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "reachable generic/any template requires lowering instantiation or representation proof",
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/core.ts",
        "evaluationOrder": 1,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 28379,
    "type": "MapLike<Type>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "Type",
    "cycle": [
      "MapLike<Type>",
      "{}",
      "MapLike<Type>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/symbolWalker.ts",
        "evaluationOrder": 30,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 21266,
    "type": "MapLike<unknown>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "unknown",
    "cycle": [
      "MapLike<unknown>",
      "{}",
      "MapLike<unknown>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": []
  },
  {
    "id": 6091,
    "type": "MapLike<WatchDirectoryFlags>",
    "declaration": "MapLike (src/compiler/corePublic.ts:13:1)",
    "valueType": "WatchDirectoryFlags",
    "cycle": [
      "MapLike<WatchDirectoryFlags>",
      "{}",
      "MapLike<WatchDirectoryFlags>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 25708,
    "type": "MissingList<ClassElement>",
    "declaration": "MissingList (src/compiler/parser.ts:3565:5)",
    "valueType": "ClassElement",
    "cycle": [
      "MissingList<ClassElement>",
      "{}",
      "MissingList<ClassElement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25714,
    "type": "MissingList<EnumMember>",
    "declaration": "MissingList (src/compiler/parser.ts:3565:5)",
    "valueType": "EnumMember",
    "cycle": [
      "MissingList<EnumMember>",
      "{}",
      "MissingList<EnumMember>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25304,
    "type": "MissingList<Node>",
    "declaration": "MissingList (src/compiler/parser.ts:3565:5)",
    "valueType": "Node",
    "cycle": [
      "MissingList<Node>",
      "{}",
      "MissingList<Node>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25384,
    "type": "MissingList<ParameterDeclaration>",
    "declaration": "MissingList (src/compiler/parser.ts:3565:5)",
    "valueType": "ParameterDeclaration",
    "cycle": [
      "MissingList<ParameterDeclaration>",
      "{}",
      "MissingList<ParameterDeclaration>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25603,
    "type": "MissingList<Statement>",
    "declaration": "MissingList (src/compiler/parser.ts:3565:5)",
    "valueType": "Statement",
    "cycle": [
      "MissingList<Statement>",
      "{}",
      "MissingList<Statement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25401,
    "type": "MissingList<TypeElement>",
    "declaration": "MissingList (src/compiler/parser.ts:3565:5)",
    "valueType": "TypeElement",
    "cycle": [
      "MissingList<TypeElement>",
      "{}",
      "MissingList<TypeElement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25675,
    "type": "MissingList<VariableDeclaration>",
    "declaration": "MissingList (src/compiler/parser.ts:3565:5)",
    "valueType": "VariableDeclaration",
    "cycle": [
      "MissingList<VariableDeclaration>",
      "{}",
      "MissingList<VariableDeclaration>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3883,
    "type": "ModifiersArray",
    "declaration": "ModifiersArray (src/compiler/types.ts:1682:1)",
    "valueType": "Modifier",
    "cycle": [
      "ModifiersArray",
      "{}",
      "ModifiersArray"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 35075,
    "type": "MutableNodeArray<Modifier>",
    "declaration": "MutableNodeArray (src/compiler/types.ts:1584:1)",
    "valueType": "Modifier",
    "cycle": [
      "MutableNodeArray<Modifier>",
      "{}",
      "MutableNodeArray<Modifier>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 23081,
    "type": "MutableNodeArray<Node>",
    "declaration": "MutableNodeArray (src/compiler/types.ts:1584:1)",
    "valueType": "Node",
    "cycle": [
      "MutableNodeArray<Node>",
      "{}",
      "MutableNodeArray<Node>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3989,
    "type": "NodeArray<ArrayBindingElement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ArrayBindingElement",
    "cycle": [
      "NodeArray<ArrayBindingElement>",
      "{}",
      "NodeArray<ArrayBindingElement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 7187,
    "type": "NodeArray<AssertEntry>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "AssertEntry",
    "cycle": [
      "NodeArray<AssertEntry>",
      "{}",
      "NodeArray<AssertEntry>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3987,
    "type": "NodeArray<BindingElement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "BindingElement",
    "cycle": [
      "NodeArray<BindingElement>",
      "{}",
      "NodeArray<BindingElement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 2416,
    "type": "NodeArray<CaseOrDefaultClause>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "CaseOrDefaultClause",
    "cycle": [
      "NodeArray<CaseOrDefaultClause>",
      "{}",
      "NodeArray<CaseOrDefaultClause>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19422,
    "type": "NodeArray<ClassElement | ObjectLiteralElement | TypeElement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ClassElement | ObjectLiteralElement | TypeElement",
    "cycle": [
      "NodeArray<ClassElement | ObjectLiteralElement | TypeElement>",
      "{}",
      "NodeArray<ClassElement | ObjectLiteralElement | TypeElement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4336,
    "type": "NodeArray<ClassElement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ClassElement",
    "cycle": [
      "NodeArray<ClassElement>",
      "{}",
      "NodeArray<ClassElement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 23600,
    "type": "NodeArray<Declaration>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "Declaration",
    "cycle": [
      "NodeArray<Declaration>",
      "{}",
      "NodeArray<Declaration>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39685,
    "type": "NodeArray<Decorator>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "Decorator",
    "cycle": [
      "NodeArray<Decorator>",
      "{}",
      "NodeArray<Decorator>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4345,
    "type": "NodeArray<EnumMember>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "EnumMember",
    "cycle": [
      "NodeArray<EnumMember>",
      "{}",
      "NodeArray<EnumMember>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4386,
    "type": "NodeArray<ExportSpecifier>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ExportSpecifier",
    "cycle": [
      "NodeArray<ExportSpecifier>",
      "{}",
      "NodeArray<ExportSpecifier>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4191,
    "type": "NodeArray<Expression>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "Expression",
    "cycle": [
      "NodeArray<Expression>",
      "{}",
      "NodeArray<Expression>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 38318,
    "type": "NodeArray<ExpressionStatement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ExpressionStatement",
    "cycle": [
      "NodeArray<ExpressionStatement>",
      "{}",
      "NodeArray<ExpressionStatement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4341,
    "type": "NodeArray<ExpressionWithTypeArguments>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ExpressionWithTypeArguments",
    "cycle": [
      "NodeArray<ExpressionWithTypeArguments>",
      "{}",
      "NodeArray<ExpressionWithTypeArguments>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4335,
    "type": "NodeArray<HeritageClause>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "HeritageClause",
    "cycle": [
      "NodeArray<HeritageClause>",
      "{}",
      "NodeArray<HeritageClause>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4376,
    "type": "NodeArray<ImportAttribute>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ImportAttribute",
    "cycle": [
      "NodeArray<ImportAttribute>",
      "{}",
      "NodeArray<ImportAttribute>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4384,
    "type": "NodeArray<ImportSpecifier>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ImportSpecifier",
    "cycle": [
      "NodeArray<ImportSpecifier>",
      "{}",
      "NodeArray<ImportSpecifier>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4444,
    "type": "NodeArray<JSDocComment>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JSDocComment",
    "cycle": [
      "NodeArray<JSDocComment>",
      "{}",
      "NodeArray<JSDocComment>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 22776,
    "type": "NodeArray<JSDocParameterTag>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JSDocParameterTag",
    "cycle": [
      "NodeArray<JSDocParameterTag>",
      "{}",
      "NodeArray<JSDocParameterTag>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39670,
    "type": "NodeArray<JSDocPropertyLikeTag>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JSDocPropertyLikeTag",
    "cycle": [
      "NodeArray<JSDocPropertyLikeTag>",
      "{}",
      "NodeArray<JSDocPropertyLikeTag>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4442,
    "type": "NodeArray<JSDocTag>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JSDocTag",
    "cycle": [
      "NodeArray<JSDocTag>",
      "{}",
      "NodeArray<JSDocTag>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39671,
    "type": "NodeArray<JSDocTemplateTag>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JSDocTemplateTag",
    "cycle": [
      "NodeArray<JSDocTemplateTag>",
      "{}",
      "NodeArray<JSDocTemplateTag>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4587,
    "type": "NodeArray<JsonObjectExpressionStatement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JsonObjectExpressionStatement",
    "cycle": [
      "NodeArray<JsonObjectExpressionStatement>",
      "{}",
      "NodeArray<JsonObjectExpressionStatement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25557,
    "type": "NodeArray<JsxAttribute | JsxSpreadAttribute>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JsxAttribute | JsxSpreadAttribute",
    "cycle": [
      "NodeArray<JsxAttribute | JsxSpreadAttribute>",
      "{}",
      "NodeArray<JsxAttribute | JsxSpreadAttribute>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4271,
    "type": "NodeArray<JsxAttributeLike>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JsxAttributeLike",
    "cycle": [
      "NodeArray<JsxAttributeLike>",
      "{}",
      "NodeArray<JsxAttributeLike>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4263,
    "type": "NodeArray<JsxChild>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JsxChild",
    "cycle": [
      "NodeArray<JsxChild>",
      "{}",
      "NodeArray<JsxChild>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25554,
    "type": "NodeArray<JsxElement | JsxSelfClosingElement | JsxFragment | JsxExpression | JsxText>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "JsxElement | JsxSelfClosingElement | JsxFragment | JsxExpression | JsxText",
    "cycle": [
      "NodeArray<JsxElement | JsxSelfClosingElement | JsxFragment | JsxExpression | JsxText>",
      "{}",
      "NodeArray<JsxElement | JsxSelfClosingElement | JsxFragment | JsxExpression | JsxText>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3933,
    "type": "NodeArray<Modifier>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "Modifier",
    "cycle": [
      "NodeArray<Modifier>",
      "{}",
      "NodeArray<Modifier>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3959,
    "type": "NodeArray<ModifierLike>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ModifierLike",
    "cycle": [
      "NodeArray<ModifierLike>",
      "{}",
      "NodeArray<ModifierLike>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 23914,
    "type": "NodeArray<never>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "never",
    "cycle": [
      "NodeArray<never>",
      "{}",
      "NodeArray<never>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1995,
    "type": "NodeArray<Node>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "Node",
    "cycle": [
      "NodeArray<Node>",
      "{}",
      "NodeArray<Node>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25291,
    "type": "NodeArray<NonNullable<T>>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "NonNullable<T>",
    "cycle": [
      "NodeArray<NonNullable<T>>",
      "{}",
      "NodeArray<NonNullable<T>>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "reachable generic/any template requires lowering instantiation or representation proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 21345,
    "type": "NodeArray<ObjectLiteralElement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ObjectLiteralElement",
    "cycle": [
      "NodeArray<ObjectLiteralElement>",
      "{}",
      "NodeArray<ObjectLiteralElement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 31389,
    "type": "NodeArray<ObjectLiteralElementLike | JsxAttributeLike>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ObjectLiteralElementLike | JsxAttributeLike",
    "cycle": [
      "NodeArray<ObjectLiteralElementLike | JsxAttributeLike>",
      "{}",
      "NodeArray<ObjectLiteralElementLike | JsxAttributeLike>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19425,
    "type": "NodeArray<ObjectLiteralElementLike>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ObjectLiteralElementLike",
    "cycle": [
      "NodeArray<ObjectLiteralElementLike>",
      "{}",
      "NodeArray<ObjectLiteralElementLike>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3942,
    "type": "NodeArray<ParameterDeclaration>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "ParameterDeclaration",
    "cycle": [
      "NodeArray<ParameterDeclaration>",
      "{}",
      "NodeArray<ParameterDeclaration>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4299,
    "type": "NodeArray<Statement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "Statement",
    "cycle": [
      "NodeArray<Statement>",
      "{}",
      "NodeArray<Statement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4051,
    "type": "NodeArray<TemplateLiteralTypeSpan>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "TemplateLiteralTypeSpan",
    "cycle": [
      "NodeArray<TemplateLiteralTypeSpan>",
      "{}",
      "NodeArray<TemplateLiteralTypeSpan>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4185,
    "type": "NodeArray<TemplateSpan>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "TemplateSpan",
    "cycle": [
      "NodeArray<TemplateSpan>",
      "{}",
      "NodeArray<TemplateSpan>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4031,
    "type": "NodeArray<TypeElement>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "TypeElement",
    "cycle": [
      "NodeArray<TypeElement>",
      "{}",
      "NodeArray<TypeElement>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 4032,
    "type": "NodeArray<TypeNode | NamedTupleMember>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "TypeNode | NamedTupleMember",
    "cycle": [
      "NodeArray<TypeNode | NamedTupleMember>",
      "{}",
      "NodeArray<TypeNode | NamedTupleMember>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3944,
    "type": "NodeArray<TypeNode>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "TypeNode",
    "cycle": [
      "NodeArray<TypeNode>",
      "{}",
      "NodeArray<TypeNode>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 6607,
    "type": "NodeArray<TypeParameterDeclaration | TypeNode>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "TypeParameterDeclaration | TypeNode",
    "cycle": [
      "NodeArray<TypeParameterDeclaration | TypeNode>",
      "{}",
      "NodeArray<TypeParameterDeclaration | TypeNode>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3941,
    "type": "NodeArray<TypeParameterDeclaration>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "TypeParameterDeclaration",
    "cycle": [
      "NodeArray<TypeParameterDeclaration>",
      "{}",
      "NodeArray<TypeParameterDeclaration>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 3956,
    "type": "NodeArray<VariableDeclaration>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "VariableDeclaration",
    "cycle": [
      "NodeArray<VariableDeclaration>",
      "{}",
      "NodeArray<VariableDeclaration>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "ownership cycle crosses a record requiring a storage proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 5276,
    "type": "NodeArray<VariableDeclarationInitializedTo<RequireOrImportCall>>",
    "declaration": "NodeArray (src/compiler/types.ts:1589:1)",
    "valueType": "VariableDeclarationInitializedTo<RequireOrImportCall>",
    "cycle": [
      "NodeArray<VariableDeclarationInitializedTo<RequireOrImportCall>>",
      "{}",
      "NodeArray<VariableDeclarationInitializedTo<RequireOrImportCall>>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 8776,
    "type": "NonSharedBuffer",
    "declaration": "NonSharedBuffer (external/node_modules/@types/node/buffer.buffer.d.ts:459:9)",
    "valueType": "number",
    "cycle": [
      "NonSharedBuffer",
      "{}",
      "NonSharedBuffer"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 26721,
    "type": "OptionsBase",
    "declaration": "OptionsBase (src/compiler/commandLineParser.ts:1911:1)",
    "valueType": "TsConfigSourceFile | CompilerOptionsValue",
    "cycle": [
      "OptionsBase",
      "{}",
      "OptionsBase"
    ],
    "literalSites": 2,
    "unambiguousLiteralSites": 2,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 2
      }
    ]
  },
  {
    "id": 38663,
    "type": "Partial<Record<SyntaxKind, DiagnosticMessage>>",
    "declaration": "Partial (external/node_modules/typescript/lib/lib.es5.d.ts:1583:1)",
    "valueType": "DiagnosticMessage | undefined",
    "cycle": [
      "Partial<Record<SyntaxKind, DiagnosticMessage>>",
      "{}",
      "Partial<Record<SyntaxKind, DiagnosticMessage>>"
    ],
    "literalSites": 2,
    "unambiguousLiteralSites": 2,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/transformers/declarations/diagnostics.ts",
        "evaluationOrder": 60,
        "literalCount": 2
      }
    ]
  },
  {
    "id": 20723,
    "type": "Pick<CompilerOptions, \"allowImportingTsExtensions\" | \"rewriteRelativeImportExtensions\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"allowImportingTsExtensions\" | \"rewriteRelativeImportExtensions\">",
      "{}",
      "Pick<CompilerOptions, \"allowImportingTsExtensions\" | \"rewriteRelativeImportExtensions\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20847,
    "type": "Pick<CompilerOptions, \"allowJs\" | \"checkJs\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"allowJs\" | \"checkJs\">",
      "{}",
      "Pick<CompilerOptions, \"allowJs\" | \"checkJs\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20783,
    "type": "Pick<CompilerOptions, \"allowSyntheticDefaultImports\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"allowSyntheticDefaultImports\">",
      "{}",
      "Pick<CompilerOptions, \"allowSyntheticDefaultImports\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20911,
    "type": "Pick<CompilerOptions, \"alwaysStrict\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"alwaysStrict\">",
      "{}",
      "Pick<CompilerOptions, \"alwaysStrict\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20830,
    "type": "Pick<CompilerOptions, \"composite\" | \"incremental\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"composite\" | \"incremental\">",
      "{}",
      "Pick<CompilerOptions, \"composite\" | \"incremental\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20815,
    "type": "Pick<CompilerOptions, \"declaration\" | \"composite\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"declaration\" | \"composite\">",
      "{}",
      "Pick<CompilerOptions, \"declaration\" | \"composite\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20838,
    "type": "Pick<CompilerOptions, \"declaration\" | \"declarationMap\" | \"composite\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"declaration\" | \"declarationMap\" | \"composite\">",
      "{}",
      "Pick<CompilerOptions, \"declaration\" | \"declarationMap\" | \"composite\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20777,
    "type": "Pick<CompilerOptions, \"esModuleInterop\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"esModuleInterop\">",
      "{}",
      "Pick<CompilerOptions, \"esModuleInterop\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20823,
    "type": "Pick<CompilerOptions, \"isolatedModules\" | \"preserveConstEnums\" | \"verbatimModuleSyntax\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"isolatedModules\" | \"preserveConstEnums\" | \"verbatimModuleSyntax\">",
      "{}",
      "Pick<CompilerOptions, \"isolatedModules\" | \"preserveConstEnums\" | \"verbatimModuleSyntax\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20771,
    "type": "Pick<CompilerOptions, \"isolatedModules\" | \"verbatimModuleSyntax\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"isolatedModules\" | \"verbatimModuleSyntax\">",
      "{}",
      "Pick<CompilerOptions, \"isolatedModules\" | \"verbatimModuleSyntax\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 39084,
    "type": "Pick<CompilerOptions, \"jsx\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "JsxEmit | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"jsx\">",
      "{}",
      "Pick<CompilerOptions, \"jsx\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20861,
    "type": "Pick<CompilerOptions, \"noImplicitAny\" | \"strict\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"noImplicitAny\" | \"strict\">",
      "{}",
      "Pick<CompilerOptions, \"noImplicitAny\" | \"strict\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20869,
    "type": "Pick<CompilerOptions, \"noImplicitThis\" | \"strict\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"noImplicitThis\" | \"strict\">",
      "{}",
      "Pick<CompilerOptions, \"noImplicitThis\" | \"strict\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20890,
    "type": "Pick<CompilerOptions, \"strict\" | \"strictBindCallApply\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"strict\" | \"strictBindCallApply\">",
      "{}",
      "Pick<CompilerOptions, \"strict\" | \"strictBindCallApply\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20904,
    "type": "Pick<CompilerOptions, \"strict\" | \"strictBuiltinIteratorReturn\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"strict\" | \"strictBuiltinIteratorReturn\">",
      "{}",
      "Pick<CompilerOptions, \"strict\" | \"strictBuiltinIteratorReturn\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20883,
    "type": "Pick<CompilerOptions, \"strict\" | \"strictFunctionTypes\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"strict\" | \"strictFunctionTypes\">",
      "{}",
      "Pick<CompilerOptions, \"strict\" | \"strictFunctionTypes\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20876,
    "type": "Pick<CompilerOptions, \"strict\" | \"strictNullChecks\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"strict\" | \"strictNullChecks\">",
      "{}",
      "Pick<CompilerOptions, \"strict\" | \"strictNullChecks\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20897,
    "type": "Pick<CompilerOptions, \"strict\" | \"strictPropertyInitialization\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"strict\" | \"strictPropertyInitialization\">",
      "{}",
      "Pick<CompilerOptions, \"strict\" | \"strictPropertyInitialization\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20917,
    "type": "Pick<CompilerOptions, \"strict\" | \"useUnknownInCatchVariables\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"strict\" | \"useUnknownInCatchVariables\">",
      "{}",
      "Pick<CompilerOptions, \"strict\" | \"useUnknownInCatchVariables\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20762,
    "type": "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleDetection\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "ModuleKind | ScriptTarget | ModuleDetectionKind | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleDetection\">",
      "{}",
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleDetection\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20806,
    "type": "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolveJsonModule\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | ModuleKind | ScriptTarget | ModuleResolutionKind | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolveJsonModule\">",
      "{}",
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolveJsonModule\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20800,
    "type": "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolvePackageJsonExports\" | \"resolvePackageJsonImports\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | ModuleKind | ScriptTarget | ModuleResolutionKind | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolvePackageJsonExports\" | \"resolvePackageJsonImports\">",
      "{}",
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolvePackageJsonExports\" | \"resolvePackageJsonImports\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20790,
    "type": "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolvePackageJsonExports\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | ModuleKind | ScriptTarget | ModuleResolutionKind | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolvePackageJsonExports\">",
      "{}",
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\" | \"resolvePackageJsonExports\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20750,
    "type": "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "ModuleKind | ScriptTarget | ModuleResolutionKind | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\">",
      "{}",
      "Pick<CompilerOptions, \"target\" | \"module\" | \"moduleResolution\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20853,
    "type": "Pick<CompilerOptions, \"target\" | \"module\" | \"useDefineForClassFields\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "boolean | ModuleKind | ScriptTarget | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"target\" | \"module\" | \"useDefineForClassFields\">",
      "{}",
      "Pick<CompilerOptions, \"target\" | \"module\" | \"useDefineForClassFields\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20738,
    "type": "Pick<CompilerOptions, \"target\" | \"module\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "ModuleKind | ScriptTarget | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"target\" | \"module\">",
      "{}",
      "Pick<CompilerOptions, \"target\" | \"module\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20729,
    "type": "Pick<CompilerOptions, \"target\">",
    "declaration": "Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1)",
    "valueType": "ScriptTarget | undefined",
    "cycle": [
      "Pick<CompilerOptions, \"target\">",
      "{}",
      "Pick<CompilerOptions, \"target\">"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 8267,
    "type": "PollingIntervalQueue",
    "declaration": "PollingIntervalQueue (src/compiler/sys.ts:255:5)",
    "valueType": "WatchedFileWithUnchangedPolls",
    "cycle": [
      "PollingIntervalQueue",
      "{}",
      "PollingIntervalQueue"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 8663,
    "type": "ProcessEnv",
    "declaration": "ProcessEnv (external/node_modules/@types/node/process.d.ts:468:13)",
    "valueType": "string | undefined",
    "cycle": [
      "ProcessEnv",
      "{}",
      "ProcessEnv"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 44742,
    "type": "ProcessVersions",
    "declaration": "ProcessVersions (external/node_modules/@types/node/process.d.ts:269:13)",
    "valueType": "string | undefined",
    "cycle": [
      "ProcessVersions",
      "{}",
      "ProcessVersions"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "enumerationRecordStorageType needs the lowerer enumeration provenance proof",
      "cycle uses source structural-view candidate approximation",
      "enumerationRecordStorageType needs the lowerer enumeration provenance proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 44066,
    "type": "PropertyDescriptorMap",
    "declaration": "PropertyDescriptorMap (external/node_modules/typescript/lib/lib.es5.d.ts:117:1)",
    "valueType": "PropertyDescriptor",
    "cycle": [
      "PropertyDescriptorMap",
      "{}",
      "PropertyDescriptorMap"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "reachable generic/any template requires lowering instantiation or representation proof",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 41937,
    "type": "Readonly<PathPathComponents>",
    "declaration": "Readonly (external/node_modules/typescript/lib/lib.es5.d.ts:1597:1)",
    "valueType": "Path",
    "cycle": [
      "Readonly<PathPathComponents>",
      "{}",
      "Readonly<PathPathComponents>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 1816,
    "type": "Readonly<Record<string, string | number>>",
    "declaration": "Readonly (external/node_modules/typescript/lib/lib.es5.d.ts:1597:1)",
    "valueType": "string | number",
    "cycle": [
      "Readonly<Record<string, string | number>>",
      "{}",
      "Readonly<Record<string, string | number>>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 44128,
    "type": "Record<\"__debugToString\", unknown>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "unknown",
    "cycle": [
      "Record<\"__debugToString\", unknown>",
      "{}",
      "Record<\"__debugToString\", unknown>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 44116,
    "type": "Record<\"__debugTypeToString\", unknown>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "unknown",
    "cycle": [
      "Record<\"__debugTypeToString\", unknown>",
      "{}",
      "Record<\"__debugTypeToString\", unknown>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 44131,
    "type": "Record<\"antecedent\", unknown>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "unknown",
    "cycle": [
      "Record<\"antecedent\", unknown>",
      "{}",
      "Record<\"antecedent\", unknown>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 47408,
    "type": "Record<\"category\", unknown>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "unknown",
    "cycle": [
      "Record<\"category\", unknown>",
      "{}",
      "Record<\"category\", unknown>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 44165,
    "type": "Record<\"code\", unknown>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "unknown",
    "cycle": [
      "Record<\"code\", unknown>",
      "{}",
      "Record<\"code\", unknown>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 44166,
    "type": "Record<\"message\", unknown>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "unknown",
    "cycle": [
      "Record<\"message\", unknown>",
      "{}",
      "Record<\"message\", unknown>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 47407,
    "type": "Record<\"messageText\", unknown>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "unknown",
    "cycle": [
      "Record<\"messageText\", unknown>",
      "{}",
      "Record<\"messageText\", unknown>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 6702,
    "type": "Record<LanugageFeatures, ScriptTarget>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "ScriptTarget",
    "cycle": [
      "Record<LanugageFeatures, ScriptTarget>",
      "{}",
      "Record<LanugageFeatures, ScriptTarget>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/types.ts",
        "evaluationOrder": 9,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 2352,
    "type": "Record<number, FlowGraphNode>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "FlowGraphNode",
    "cycle": [
      "Record<number, FlowGraphNode>",
      "{}",
      "Record<number, FlowGraphNode>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 40880,
    "type": "Record<number, string>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "string",
    "cycle": [
      "Record<number, string>",
      "{}",
      "Record<number, string>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 20919,
    "type": "Record<string, { dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; }>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "{ dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; }",
    "cycle": [
      "Record<string, { dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; }>",
      "{}",
      "Record<string, { dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; }>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 26915,
    "type": "Record<string, CompilerOptionsValue>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "CompilerOptionsValue",
    "cycle": [
      "Record<string, CompilerOptionsValue>",
      "{}",
      "Record<string, CompilerOptionsValue>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 1074,
    "type": "Record<string, T[]>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "T[]",
    "cycle": [
      "Record<string, T[]>",
      "{}",
      "Record<string, T[]>"
    ],
    "literalSites": 1,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "reachable generic/any template requires lowering instantiation or representation proof",
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": [
      {
        "file": "src/compiler/core.ts",
        "evaluationOrder": 1,
        "literalCount": 1
      }
    ]
  },
  {
    "id": 24900,
    "type": "Record<SyntaxKind, ForEachChildFunction<any>>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "ForEachChildFunction<any>",
    "cycle": [
      "Record<SyntaxKind, ForEachChildFunction<any>>",
      "{}",
      "Record<SyntaxKind, ForEachChildFunction<any>>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "reachable generic/any template requires lowering instantiation or representation proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 35464,
    "type": "Record<SyntaxKind, VisitEachChildFunction<any> | undefined>",
    "declaration": "Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1)",
    "valueType": "VisitEachChildFunction<any> | undefined",
    "cycle": [
      "Record<SyntaxKind, VisitEachChildFunction<any> | undefined>",
      "{}",
      "Record<SyntaxKind, VisitEachChildFunction<any> | undefined>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "finite mapped record requires Go utility-alias/storage proof",
      "reachable generic/any template requires lowering instantiation or representation proof",
      "cycle uses source structural-view candidate approximation",
      "finite mapped record requires Go utility-alias/storage proof",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 2508,
    "type": "RegExpExecArray",
    "declaration": "RegExpExecArray (external/node_modules/typescript/lib/lib.es5.d.ts:976:1); RegExpExecArray (external/node_modules/typescript/lib/lib.es2018.regexp.d.ts:23:1)",
    "valueType": "string",
    "cycle": [
      "RegExpExecArray",
      "{}",
      "RegExpExecArray"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 21191,
    "type": "RegExpMatchArray",
    "declaration": "RegExpMatchArray (external/node_modules/typescript/lib/lib.es5.d.ts:961:1); RegExpMatchArray (external/node_modules/typescript/lib/lib.es2018.regexp.d.ts:17:1)",
    "valueType": "string",
    "cycle": [
      "RegExpMatchArray",
      "{}",
      "RegExpMatchArray"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 39932,
    "type": "SortedArray<Canonicalized>",
    "declaration": "SortedArray (src/compiler/corePublic.ts:21:1)",
    "valueType": "Canonicalized",
    "cycle": [
      "SortedArray<Canonicalized>",
      "{}",
      "SortedArray<Canonicalized>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19996,
    "type": "SortedArray<Diagnostic>",
    "declaration": "SortedArray (src/compiler/corePublic.ts:21:1)",
    "valueType": "Diagnostic",
    "cycle": [
      "SortedArray<Diagnostic>",
      "{}",
      "SortedArray<Diagnostic>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19999,
    "type": "SortedArray<DiagnosticWithLocation>",
    "declaration": "SortedArray (src/compiler/corePublic.ts:21:1)",
    "valueType": "DiagnosticWithLocation",
    "cycle": [
      "SortedArray<DiagnosticWithLocation>",
      "{}",
      "SortedArray<DiagnosticWithLocation>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19997,
    "type": "SortedArray<string>",
    "declaration": "SortedArray (src/compiler/corePublic.ts:21:1)",
    "valueType": "string",
    "cycle": [
      "SortedArray<string>",
      "{}",
      "SortedArray<string>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1819,
    "type": "SortedReadonlyArray<[number, string]>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "[number, string]",
    "cycle": [
      "SortedReadonlyArray<[number, string]>",
      "{}",
      "SortedReadonlyArray<[number, string]>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39970,
    "type": "SortedReadonlyArray<Canonicalized>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "Canonicalized",
    "cycle": [
      "SortedReadonlyArray<Canonicalized>",
      "{}",
      "SortedReadonlyArray<Canonicalized>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 43621,
    "type": "SortedReadonlyArray<CommandLineOption>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "CommandLineOption",
    "cycle": [
      "SortedReadonlyArray<CommandLineOption>",
      "{}",
      "SortedReadonlyArray<CommandLineOption>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 40273,
    "type": "SortedReadonlyArray<Diagnostic>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "Diagnostic",
    "cycle": [
      "SortedReadonlyArray<Diagnostic>",
      "{}",
      "SortedReadonlyArray<Diagnostic>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39497,
    "type": "SortedReadonlyArray<EmitHelper>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "EmitHelper",
    "cycle": [
      "SortedReadonlyArray<EmitHelper>",
      "{}",
      "SortedReadonlyArray<EmitHelper>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 35729,
    "type": "SortedReadonlyArray<MappedPosition>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "MappedPosition",
    "cycle": [
      "SortedReadonlyArray<MappedPosition>",
      "{}",
      "SortedReadonlyArray<MappedPosition>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 40537,
    "type": "SortedReadonlyArray<SourceFile>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "SourceFile",
    "cycle": [
      "SortedReadonlyArray<SourceFile>",
      "{}",
      "SortedReadonlyArray<SourceFile>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "closure environments / captured cells are lowering identities",
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 35732,
    "type": "SortedReadonlyArray<SourceMappedPosition>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "SourceMappedPosition",
    "cycle": [
      "SortedReadonlyArray<SourceMappedPosition>",
      "{}",
      "SortedReadonlyArray<SourceMappedPosition>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 628,
    "type": "SortedReadonlyArray<string>",
    "declaration": "SortedReadonlyArray (src/compiler/corePublic.ts:17:1)",
    "valueType": "string",
    "cycle": [
      "SortedReadonlyArray<string>",
      "{}",
      "SortedReadonlyArray<string>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 44821,
    "type": "Storage",
    "declaration": "Storage (external/node_modules/@types/node/web-globals/storage.d.ts:16:5); Storage (external/node_modules/@types/node/web-globals/storage.d.ts:17:9)",
    "valueType": "any",
    "cycle": [
      "Storage",
      "{}",
      "Storage"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "reachable generic/any template requires lowering instantiation or representation proof",
      "cycle uses source structural-view candidate approximation",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 21417,
    "type": "String",
    "declaration": "String (external/node_modules/typescript/lib/lib.es5.d.ts:408:1); String (external/node_modules/typescript/lib/lib.es5.d.ts:542:13); String (external/node_modules/typescript/lib/lib.es5.d.ts:4560:1); String (external/node_modules/typescript/lib/lib.es2015.core.d.ts:400:1); String (external/node_modules/typescript/lib/lib.es2015.iterable.d.ts:268:1); String (external/node_modules/typescript/lib/lib.es2015.symbol.wellknown.d.ts:233:1); String (external/node_modules/typescript/lib/lib.es2017.string.d.ts:17:1); String (external/node_modules/typescript/lib/lib.es2019.string.d.ts:17:1); String (external/node_modules/typescript/lib/lib.es2020.string.d.ts:21:1)",
    "valueType": "string",
    "cycle": [
      "String",
      "{}",
      "String"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 23336,
    "type": "TemplateStringsArray",
    "declaration": "TemplateStringsArray (external/node_modules/typescript/lib/lib.es5.d.ts:619:1)",
    "valueType": "string",
    "cycle": [
      "TemplateStringsArray",
      "{}",
      "TemplateStringsArray"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 6033,
    "type": "TypeAcquisition",
    "declaration": "TypeAcquisition (src/compiler/types.ts:7596:1)",
    "valueType": "CompilerOptionsValue",
    "cycle": [
      "TypeAcquisition",
      "{}",
      "TypeAcquisition"
    ],
    "literalSites": 2,
    "unambiguousLiteralSites": 1,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 2
      }
    ]
  },
  {
    "id": 28697,
    "type": "typeof AccessFlags",
    "declaration": "AccessFlags (src/compiler/types.ts:6895:1)",
    "valueType": "string",
    "cycle": [
      "typeof AccessFlags",
      "{}",
      "typeof AccessFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 20400,
    "type": "typeof AccessKind",
    "declaration": "AccessKind (src/compiler/utilities.ts:8089:1)",
    "valueType": "string",
    "cycle": [
      "typeof AccessKind",
      "{}",
      "typeof AccessKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1569,
    "type": "typeof AssertionLevel",
    "declaration": "AssertionLevel (src/compiler/core.ts:1924:1)",
    "valueType": "string",
    "cycle": [
      "typeof AssertionLevel",
      "{}",
      "typeof AssertionLevel"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 17944,
    "type": "typeof AssignmentDeclarationKind",
    "declaration": "AssignmentDeclarationKind (src/compiler/types.ts:7197:1)",
    "valueType": "string",
    "cycle": [
      "typeof AssignmentDeclarationKind",
      "{}",
      "typeof AssignmentDeclarationKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19752,
    "type": "typeof AssignmentKind",
    "declaration": "AssignmentKind (src/compiler/utilities.ts:4840:1)",
    "valueType": "string",
    "cycle": [
      "typeof AssignmentKind",
      "{}",
      "typeof AssignmentKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19936,
    "type": "typeof Associativity",
    "declaration": "Associativity (src/compiler/utilities.ts:5629:1)",
    "valueType": "string",
    "cycle": [
      "typeof Associativity",
      "{}",
      "typeof Associativity"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 38199,
    "type": "typeof BlockAction",
    "declaration": "BlockAction (src/compiler/transformers/generators.ts:239:1)",
    "valueType": "string",
    "cycle": [
      "typeof BlockAction",
      "{}",
      "typeof BlockAction"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 41337,
    "type": "typeof BuilderFileEmit",
    "declaration": "BuilderFileEmit (src/compiler/builder.ts:186:1)",
    "valueType": "string",
    "cycle": [
      "typeof BuilderFileEmit",
      "{}",
      "typeof BuilderFileEmit"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 41676,
    "type": "typeof BuilderProgramKind",
    "declaration": "BuilderProgramKind (src/compiler/builder.ts:1593:1)",
    "valueType": "string",
    "cycle": [
      "typeof BuilderProgramKind",
      "{}",
      "typeof BuilderProgramKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 43221,
    "type": "typeof BuildResultFlags",
    "declaration": "BuildResultFlags (src/compiler/tsbuildPublic.ts:168:1)",
    "valueType": "string",
    "cycle": [
      "typeof BuildResultFlags",
      "{}",
      "typeof BuildResultFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 43172,
    "type": "typeof BuildStep",
    "declaration": "BuildStep (src/compiler/tsbuildPublic.ts:924:1)",
    "valueType": "string",
    "cycle": [
      "typeof BuildStep",
      "{}",
      "typeof BuildStep"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 34,
    "type": "typeof CharacterCodes",
    "declaration": "CharacterCodes (src/compiler/types.ts:7852:1)",
    "valueType": "string",
    "cycle": [
      "typeof CharacterCodes",
      "{}",
      "typeof CharacterCodes"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18640,
    "type": "typeof CheckFlags",
    "declaration": "CheckFlags (src/compiler/types.ts:6110:1)",
    "valueType": "string",
    "cycle": [
      "typeof CheckFlags",
      "{}",
      "typeof CheckFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1570,
    "type": "typeof CheckMode",
    "declaration": "CheckMode (src/compiler/checker.ts:1343:1)",
    "valueType": "string",
    "cycle": [
      "typeof CheckMode",
      "{}",
      "typeof CheckMode"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 36204,
    "type": "typeof ClassFacts",
    "declaration": "ClassFacts (src/compiler/transformers/ts.ts:219:1)",
    "valueType": "string",
    "cycle": [
      "typeof ClassFacts",
      "{}",
      "typeof ClassFacts"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 36486,
    "type": "typeof ClassFacts",
    "declaration": "ClassFacts (src/compiler/transformers/classFields.ts:333:1)",
    "valueType": "string",
    "cycle": [
      "typeof ClassFacts",
      "{}",
      "typeof ClassFacts"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 36396,
    "type": "typeof ClassPropertySubstitutionFlags",
    "declaration": "ClassPropertySubstitutionFlags (src/compiler/transformers/classFields.ts:230:1)",
    "valueType": "string",
    "cycle": [
      "typeof ClassPropertySubstitutionFlags",
      "{}",
      "typeof ClassPropertySubstitutionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 17363,
    "type": "typeof ClassSetExpressionType",
    "declaration": "ClassSetExpressionType (src/compiler/scanner.ts:1015:1)",
    "valueType": "string",
    "cycle": [
      "typeof ClassSetExpressionType",
      "{}",
      "typeof ClassSetExpressionType"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 38212,
    "type": "typeof CodeBlockKind",
    "declaration": "CodeBlockKind (src/compiler/transformers/generators.ts:245:1)",
    "valueType": "string",
    "cycle": [
      "typeof CodeBlockKind",
      "{}",
      "typeof CodeBlockKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 15317,
    "type": "typeof CommentDirectiveType",
    "declaration": "CommentDirectiveType (src/compiler/types.ts:4489:1)",
    "valueType": "string",
    "cycle": [
      "typeof CommentDirectiveType",
      "{}",
      "typeof CommentDirectiveType"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 36,
    "type": "typeof Comparison",
    "declaration": "Comparison (src/compiler/corePublic.ts:43:1)",
    "valueType": "string",
    "cycle": [
      "typeof Comparison",
      "{}",
      "typeof Comparison"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 2435,
    "type": "typeof Connection",
    "declaration": "Connection (src/compiler/debug.ts:896:9)",
    "valueType": "string",
    "cycle": [
      "typeof Connection",
      "{}",
      "typeof Connection"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18642,
    "type": "typeof ContainerFlags",
    "declaration": "ContainerFlags (src/compiler/binder.ts:465:1)",
    "valueType": "string",
    "cycle": [
      "typeof ContainerFlags",
      "{}",
      "typeof ContainerFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28698,
    "type": "typeof ContextFlags",
    "declaration": "ContextFlags (src/compiler/types.ts:5520:1)",
    "valueType": "string",
    "cycle": [
      "typeof ContextFlags",
      "{}",
      "typeof ContextFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 36993,
    "type": "typeof ContextFlags",
    "declaration": "ContextFlags (src/compiler/transformers/es2017.ts:112:1)",
    "valueType": "string",
    "cycle": [
      "typeof ContextFlags",
      "{}",
      "typeof ContextFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37761,
    "type": "typeof CopyDirection",
    "declaration": "CopyDirection (src/compiler/transformers/es2015.ts:278:1)",
    "valueType": "string",
    "cycle": [
      "typeof CopyDirection",
      "{}",
      "typeof CopyDirection"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 34245,
    "type": "typeof DeclarationMeaning",
    "declaration": "DeclarationMeaning (src/compiler/checker.ts:1415:1)",
    "valueType": "string",
    "cycle": [
      "typeof DeclarationMeaning",
      "{}",
      "typeof DeclarationMeaning"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 34381,
    "type": "typeof DeclarationSpaces",
    "declaration": "DeclarationSpaces (src/compiler/checker.ts:1425:1)",
    "valueType": "string",
    "cycle": [
      "typeof DeclarationSpaces",
      "{}",
      "typeof DeclarationSpaces"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 5980,
    "type": "typeof DiagnosticCategory",
    "declaration": "DiagnosticCategory (src/compiler/types.ts:7324:1)",
    "valueType": "string",
    "cycle": [
      "typeof DiagnosticCategory",
      "{}",
      "typeof DiagnosticCategory"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28704,
    "type": "typeof ElementFlags",
    "declaration": "ElementFlags (src/compiler/types.ts:6702:1)",
    "valueType": "string",
    "cycle": [
      "typeof ElementFlags",
      "{}",
      "typeof ElementFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1571,
    "type": "typeof EmitFlags",
    "declaration": "EmitFlags (src/compiler/types.ts:8372:1)",
    "valueType": "string",
    "cycle": [
      "typeof EmitFlags",
      "{}",
      "typeof EmitFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28705,
    "type": "typeof EmitHint",
    "declaration": "EmitHint (src/compiler/types.ts:8592:1)",
    "valueType": "string",
    "cycle": [
      "typeof EmitHint",
      "{}",
      "typeof EmitHint"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 38926,
    "type": "typeof EmitOnly",
    "declaration": "EmitOnly (src/compiler/types.ts:4695:1)",
    "valueType": "string",
    "cycle": [
      "typeof EmitOnly",
      "{}",
      "typeof EmitOnly"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37680,
    "type": "typeof ES2015SubstitutionFlags",
    "declaration": "ES2015SubstitutionFlags (src/compiler/transformers/es2015.ts:220:1)",
    "valueType": "string",
    "cycle": [
      "typeof ES2015SubstitutionFlags",
      "{}",
      "typeof ES2015SubstitutionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 36991,
    "type": "typeof ES2017SubstitutionFlags",
    "declaration": "ES2017SubstitutionFlags (src/compiler/transformers/es2017.ts:106:1)",
    "valueType": "string",
    "cycle": [
      "typeof ES2017SubstitutionFlags",
      "{}",
      "typeof ES2017SubstitutionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 17206,
    "type": "typeof EscapeSequenceScanningFlags",
    "declaration": "EscapeSequenceScanningFlags (src/compiler/scanner.ts:1002:1)",
    "valueType": "string",
    "cycle": [
      "typeof EscapeSequenceScanningFlags",
      "{}",
      "typeof EscapeSequenceScanningFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37084,
    "type": "typeof ESNextSubstitutionFlags",
    "declaration": "ESNextSubstitutionFlags (src/compiler/transformers/es2018.ts:115:1)",
    "valueType": "string",
    "cycle": [
      "typeof ESNextSubstitutionFlags",
      "{}",
      "typeof ESNextSubstitutionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 38215,
    "type": "typeof ExceptionBlockState",
    "declaration": "ExceptionBlockState (src/compiler/transformers/generators.ts:254:1)",
    "valueType": "string",
    "cycle": [
      "typeof ExceptionBlockState",
      "{}",
      "typeof ExceptionBlockState"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 42216,
    "type": "typeof ExitStatus",
    "declaration": "ExitStatus (src/compiler/types.ts:5017:1)",
    "valueType": "string",
    "cycle": [
      "typeof ExitStatus",
      "{}",
      "typeof ExitStatus"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 29433,
    "type": "typeof ExpandingFlags",
    "declaration": "ExpandingFlags (src/compiler/checker.ts:1393:1)",
    "valueType": "string",
    "cycle": [
      "typeof ExpandingFlags",
      "{}",
      "typeof ExpandingFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 27273,
    "type": "typeof Extensions",
    "declaration": "Extensions (src/compiler/moduleNameResolver.ts:184:1)",
    "valueType": "string",
    "cycle": [
      "typeof Extensions",
      "{}",
      "typeof Extensions"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28706,
    "type": "typeof ExternalEmitHelpers",
    "declaration": "ExternalEmitHelpers (src/compiler/types.ts:8543:1)",
    "valueType": "string",
    "cycle": [
      "typeof ExternalEmitHelpers",
      "{}",
      "typeof ExternalEmitHelpers"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28406,
    "type": "typeof FileIncludeKind",
    "declaration": "FileIncludeKind (src/compiler/types.ts:4599:1)",
    "valueType": "string",
    "cycle": [
      "typeof FileIncludeKind",
      "{}",
      "typeof FileIncludeKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 40160,
    "type": "typeof FilePreprocessingDiagnosticsKind",
    "declaration": "FilePreprocessingDiagnosticsKind (src/compiler/types.ts:4664:1)",
    "valueType": "string",
    "cycle": [
      "typeof FilePreprocessingDiagnosticsKind",
      "{}",
      "typeof FilePreprocessingDiagnosticsKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 8313,
    "type": "typeof FileSystemEntryKind",
    "declaration": "FileSystemEntryKind (src/compiler/sys.ts:908:1)",
    "valueType": "string",
    "cycle": [
      "typeof FileSystemEntryKind",
      "{}",
      "typeof FileSystemEntryKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 8320,
    "type": "typeof FileWatcherEventKind",
    "declaration": "FileWatcherEventKind (src/compiler/sys.ts:87:1)",
    "valueType": "string",
    "cycle": [
      "typeof FileWatcherEventKind",
      "{}",
      "typeof FileWatcherEventKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 35979,
    "type": "typeof FlattenLevel",
    "declaration": "FlattenLevel (src/compiler/transformers/destructuring.ts:81:1)",
    "valueType": "string",
    "cycle": [
      "typeof FlattenLevel",
      "{}",
      "typeof FlattenLevel"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1572,
    "type": "typeof FlowFlags",
    "declaration": "FlowFlags (src/compiler/types.ts:4150:1)",
    "valueType": "string",
    "cycle": [
      "typeof FlowFlags",
      "{}",
      "typeof FlowFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19830,
    "type": "typeof FunctionFlags",
    "declaration": "FunctionFlags (src/compiler/utilities.ts:5307:1)",
    "valueType": "string",
    "cycle": [
      "typeof FunctionFlags",
      "{}",
      "typeof FunctionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 17951,
    "type": "typeof GeneratedIdentifierFlags",
    "declaration": "GeneratedIdentifierFlags (src/compiler/types.ts:1685:1)",
    "valueType": "string",
    "cycle": [
      "typeof GeneratedIdentifierFlags",
      "{}",
      "typeof GeneratedIdentifierFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19257,
    "type": "typeof GetLiteralTextFlags",
    "declaration": "GetLiteralTextFlags (src/compiler/utilities.ts:1972:1)",
    "valueType": "string",
    "cycle": [
      "typeof GetLiteralTextFlags",
      "{}",
      "typeof GetLiteralTextFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37091,
    "type": "typeof HierarchyFacts",
    "declaration": "HierarchyFacts (src/compiler/transformers/es2018.ts:122:1)",
    "valueType": "string",
    "cycle": [
      "typeof HierarchyFacts",
      "{}",
      "typeof HierarchyFacts"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37683,
    "type": "typeof HierarchyFacts",
    "declaration": "HierarchyFacts (src/compiler/transformers/es2015.ts:371:1)",
    "valueType": "string",
    "cycle": [
      "typeof HierarchyFacts",
      "{}",
      "typeof HierarchyFacts"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 26092,
    "type": "typeof ImportsNotUsedAsValues",
    "declaration": "ImportsNotUsedAsValues (src/compiler/types.ts:7643:1)",
    "valueType": "string",
    "cycle": [
      "typeof ImportsNotUsedAsValues",
      "{}",
      "typeof ImportsNotUsedAsValues"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28709,
    "type": "typeof IndexFlags",
    "declaration": "IndexFlags (src/compiler/types.ts:6924:1)",
    "valueType": "string",
    "cycle": [
      "typeof IndexFlags",
      "{}",
      "typeof IndexFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28710,
    "type": "typeof IndexKind",
    "declaration": "IndexKind (src/compiler/types.ts:7071:1)",
    "valueType": "string",
    "cycle": [
      "typeof IndexKind",
      "{}",
      "typeof IndexKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28711,
    "type": "typeof InferenceFlags",
    "declaration": "InferenceFlags (src/compiler/types.ts:7139:1)",
    "valueType": "string",
    "cycle": [
      "typeof InferenceFlags",
      "{}",
      "typeof InferenceFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28712,
    "type": "typeof InferencePriority",
    "declaration": "InferencePriority (src/compiler/types.ts:7105:1)",
    "valueType": "string",
    "cycle": [
      "typeof InferencePriority",
      "{}",
      "typeof InferencePriority"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 38059,
    "type": "typeof Instruction",
    "declaration": "Instruction (src/compiler/transformers/generators.ts:307:1)",
    "valueType": "string",
    "cycle": [
      "typeof Instruction",
      "{}",
      "typeof Instruction"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18655,
    "type": "typeof InternalEmitFlags",
    "declaration": "InternalEmitFlags (src/compiler/types.ts:8406:1)",
    "valueType": "string",
    "cycle": [
      "typeof InternalEmitFlags",
      "{}",
      "typeof InternalEmitFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28713,
    "type": "typeof InternalNodeBuilderFlags",
    "declaration": "InternalNodeBuilderFlags (src/compiler/types.ts:5574:1)",
    "valueType": "string",
    "cycle": [
      "typeof InternalNodeBuilderFlags",
      "{}",
      "typeof InternalNodeBuilderFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28714,
    "type": "typeof IntersectionFlags",
    "declaration": "IntersectionFlags (src/compiler/types.ts:5512:1)",
    "valueType": "string",
    "cycle": [
      "typeof IntersectionFlags",
      "{}",
      "typeof IntersectionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 32387,
    "type": "typeof IntersectionState",
    "declaration": "IntersectionState (src/compiler/checker.ts:1367:1)",
    "valueType": "string",
    "cycle": [
      "typeof IntersectionState",
      "{}",
      "typeof IntersectionState"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 29009,
    "type": "typeof IntrinsicTypeKind",
    "declaration": "IntrinsicTypeKind (src/compiler/checker.ts:1438:1)",
    "valueType": "string",
    "cycle": [
      "typeof IntrinsicTypeKind",
      "{}",
      "typeof IntrinsicTypeKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 43155,
    "type": "typeof InvalidatedProjectKind",
    "declaration": "InvalidatedProjectKind (src/compiler/tsbuildPublic.ts:830:1)",
    "valueType": "string",
    "cycle": [
      "typeof InvalidatedProjectKind",
      "{}",
      "typeof InvalidatedProjectKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25971,
    "type": "typeof InvalidPosition",
    "declaration": "InvalidPosition (src/compiler/parser.ts:10550:5)",
    "valueType": "string",
    "cycle": [
      "typeof InvalidPosition",
      "{}",
      "typeof InvalidPosition"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 32290,
    "type": "typeof IterationTypeKind",
    "declaration": "IterationTypeKind (src/compiler/checker.ts:1202:1)",
    "valueType": "string",
    "cycle": [
      "typeof IterationTypeKind",
      "{}",
      "typeof IterationTypeKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 30952,
    "type": "typeof IterationUse",
    "declaration": "IterationUse (src/compiler/checker.ts:1177:1)",
    "valueType": "string",
    "cycle": [
      "typeof IterationUse",
      "{}",
      "typeof IterationUse"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 15319,
    "type": "typeof JSDocParsingMode",
    "declaration": "JSDocParsingMode (src/compiler/types.ts:10295:1)",
    "valueType": "string",
    "cycle": [
      "typeof JSDocParsingMode",
      "{}",
      "typeof JSDocParsingMode"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25797,
    "type": "typeof JSDocState",
    "declaration": "JSDocState (src/compiler/parser.ts:8872:9)",
    "valueType": "string",
    "cycle": [
      "typeof JSDocState",
      "{}",
      "typeof JSDocState"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18699,
    "type": "typeof JsxEmit",
    "declaration": "JsxEmit (src/compiler/types.ts:7633:1)",
    "valueType": "string",
    "cycle": [
      "typeof JsxEmit",
      "{}",
      "typeof JsxEmit"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28718,
    "type": "typeof JsxFlags",
    "declaration": "JsxFlags (src/compiler/types.ts:905:1)",
    "valueType": "string",
    "cycle": [
      "typeof JsxFlags",
      "{}",
      "typeof JsxFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28719,
    "type": "typeof JsxReferenceKind",
    "declaration": "JsxReferenceKind (src/compiler/types.ts:6993:1)",
    "valueType": "string",
    "cycle": [
      "typeof JsxReferenceKind",
      "{}",
      "typeof JsxReferenceKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37747,
    "type": "typeof Jump",
    "declaration": "Jump (src/compiler/transformers/es2015.ts:283:1)",
    "valueType": "string",
    "cycle": [
      "typeof Jump",
      "{}",
      "typeof Jump"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 15320,
    "type": "typeof LanguageVariant",
    "declaration": "LanguageVariant (src/compiler/types.ts:7704:1)",
    "valueType": "string",
    "cycle": [
      "typeof LanguageVariant",
      "{}",
      "typeof LanguageVariant"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 35344,
    "type": "typeof LexicalEnvironmentFlags",
    "declaration": "LexicalEnvironmentFlags (src/compiler/types.ts:9563:1)",
    "valueType": "string",
    "cycle": [
      "typeof LexicalEnvironmentFlags",
      "{}",
      "typeof LexicalEnvironmentFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39064,
    "type": "typeof ListFormat",
    "declaration": "ListFormat (src/compiler/types.ts:10119:1)",
    "valueType": "string",
    "cycle": [
      "typeof ListFormat",
      "{}",
      "typeof ListFormat"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1647,
    "type": "typeof LogLevel",
    "declaration": "LogLevel (src/compiler/debug.ts:97:1)",
    "valueType": "string",
    "cycle": [
      "typeof LogLevel",
      "{}",
      "typeof LogLevel"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37943,
    "type": "typeof LoopOutParameterFlags",
    "declaration": "LoopOutParameterFlags (src/compiler/transformers/es2015.ts:272:1)",
    "valueType": "string",
    "cycle": [
      "typeof LoopOutParameterFlags",
      "{}",
      "typeof LoopOutParameterFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 30183,
    "type": "typeof MappedTypeModifiers",
    "declaration": "MappedTypeModifiers (src/compiler/checker.ts:1380:1)",
    "valueType": "string",
    "cycle": [
      "typeof MappedTypeModifiers",
      "{}",
      "typeof MappedTypeModifiers"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 31352,
    "type": "typeof MappedTypeNameTypeKind",
    "declaration": "MappedTypeNameTypeKind (src/compiler/checker.ts:1387:1)",
    "valueType": "string",
    "cycle": [
      "typeof MappedTypeNameTypeKind",
      "{}",
      "typeof MappedTypeNameTypeKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28630,
    "type": "typeof MatchingMode",
    "declaration": "MatchingMode (src/compiler/moduleSpecifiers.ts:1014:1)",
    "valueType": "string",
    "cycle": [
      "typeof MatchingMode",
      "{}",
      "typeof MatchingMode"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28720,
    "type": "typeof MemberOverrideStatus",
    "declaration": "MemberOverrideStatus (src/compiler/types.ts:5498:1)",
    "valueType": "string",
    "cycle": [
      "typeof MemberOverrideStatus",
      "{}",
      "typeof MemberOverrideStatus"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 31510,
    "type": "typeof MinArgumentCountFlags",
    "declaration": "MinArgumentCountFlags (src/compiler/checker.ts:1432:1)",
    "valueType": "string",
    "cycle": [
      "typeof MinArgumentCountFlags",
      "{}",
      "typeof MinArgumentCountFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1617,
    "type": "typeof ModifierFlags",
    "declaration": "ModifierFlags (src/compiler/types.ts:851:1)",
    "valueType": "string",
    "cycle": [
      "typeof ModifierFlags",
      "{}",
      "typeof ModifierFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18701,
    "type": "typeof ModuleDetectionKind",
    "declaration": "ModuleDetectionKind (src/compiler/types.ts:7359:1)",
    "valueType": "string",
    "cycle": [
      "typeof ModuleDetectionKind",
      "{}",
      "typeof ModuleDetectionKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 27922,
    "type": "typeof ModuleInstanceState",
    "declaration": "ModuleInstanceState (src/compiler/binder.ts:315:1)",
    "valueType": "string",
    "cycle": [
      "typeof ModuleInstanceState",
      "{}",
      "typeof ModuleInstanceState"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18702,
    "type": "typeof ModuleKind",
    "declaration": "ModuleKind (src/compiler/types.ts:7604:1)",
    "valueType": "string",
    "cycle": [
      "typeof ModuleKind",
      "{}",
      "typeof ModuleKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18703,
    "type": "typeof ModuleResolutionKind",
    "declaration": "ModuleResolutionKind (src/compiler/types.ts:7336:1)",
    "valueType": "string",
    "cycle": [
      "typeof ModuleResolutionKind",
      "{}",
      "typeof ModuleResolutionKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 21166,
    "type": "typeof ModuleSpecifierEnding",
    "declaration": "ModuleSpecifierEnding (src/compiler/utilities.ts:10047:1)",
    "valueType": "string",
    "cycle": [
      "typeof ModuleSpecifierEnding",
      "{}",
      "typeof ModuleSpecifierEnding"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18704,
    "type": "typeof NewLineKind",
    "declaration": "NewLineKind (src/compiler/types.ts:7649:1)",
    "valueType": "string",
    "cycle": [
      "typeof NewLineKind",
      "{}",
      "typeof NewLineKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28721,
    "type": "typeof NodeBuilderFlags",
    "declaration": "NodeBuilderFlags (src/compiler/types.ts:5530:1)",
    "valueType": "string",
    "cycle": [
      "typeof NodeBuilderFlags",
      "{}",
      "typeof NodeBuilderFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1618,
    "type": "typeof NodeCheckFlags",
    "declaration": "NodeCheckFlags (src/compiler/types.ts:6224:1)",
    "valueType": "string",
    "cycle": [
      "typeof NodeCheckFlags",
      "{}",
      "typeof NodeCheckFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 2839,
    "type": "typeof NodeFactoryFlags",
    "declaration": "NodeFactoryFlags (src/compiler/factory/nodeFactory.ts:466:1)",
    "valueType": "string",
    "cycle": [
      "typeof NodeFactoryFlags",
      "{}",
      "typeof NodeFactoryFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1619,
    "type": "typeof NodeFlags",
    "declaration": "NodeFlags (src/compiler/types.ts:782:1)",
    "valueType": "string",
    "cycle": [
      "typeof NodeFlags",
      "{}",
      "typeof NodeFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 27373,
    "type": "typeof NodeResolutionFeatures",
    "declaration": "NodeResolutionFeatures (src/compiler/moduleNameResolver.ts:1687:1)",
    "valueType": "string",
    "cycle": [
      "typeof NodeResolutionFeatures",
      "{}",
      "typeof NodeResolutionFeatures"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1622,
    "type": "typeof ObjectFlags",
    "declaration": "ObjectFlags (src/compiler/types.ts:6523:1)",
    "valueType": "string",
    "cycle": [
      "typeof ObjectFlags",
      "{}",
      "typeof ObjectFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 38136,
    "type": "typeof OpCode",
    "declaration": "OpCode (src/compiler/transformers/generators.ts:222:1)",
    "valueType": "string",
    "cycle": [
      "typeof OpCode",
      "{}",
      "typeof OpCode"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 19969,
    "type": "typeof OperatorPrecedence",
    "declaration": "OperatorPrecedence (src/compiler/utilities.ts:5701:1)",
    "valueType": "string",
    "cycle": [
      "typeof OperatorPrecedence",
      "{}",
      "typeof OperatorPrecedence"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18024,
    "type": "typeof OuterExpressionKinds",
    "declaration": "OuterExpressionKinds (src/compiler/types.ts:8644:1)",
    "valueType": "string",
    "cycle": [
      "typeof OuterExpressionKinds",
      "{}",
      "typeof OuterExpressionKinds"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25008,
    "type": "typeof ParsingContext",
    "declaration": "ParsingContext (src/compiler/parser.ts:8754:5)",
    "valueType": "string",
    "cycle": [
      "typeof ParsingContext",
      "{}",
      "typeof ParsingContext"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39298,
    "type": "typeof PipelinePhase",
    "declaration": "PipelinePhase (src/compiler/emitter.ts:1191:1)",
    "valueType": "string",
    "cycle": [
      "typeof PipelinePhase",
      "{}",
      "typeof PipelinePhase"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 8215,
    "type": "typeof PollingInterval",
    "declaration": "PollingInterval (src/compiler/sys.ts:102:1)",
    "valueType": "string",
    "cycle": [
      "typeof PollingInterval",
      "{}",
      "typeof PollingInterval"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 8170,
    "type": "typeof PollingWatchKind",
    "declaration": "PollingWatchKind (src/compiler/types.ts:7405:1)",
    "valueType": "string",
    "cycle": [
      "typeof PollingWatchKind",
      "{}",
      "typeof PollingWatchKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 7884,
    "type": "typeof PragmaKindFlags",
    "declaration": "PragmaKindFlags (src/compiler/types.ts:10207:1)",
    "valueType": "string",
    "cycle": [
      "typeof PragmaKindFlags",
      "{}",
      "typeof PragmaKindFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28723,
    "type": "typeof PredicateSemantics",
    "declaration": "PredicateSemantics (src/compiler/types.ts:932:1)",
    "valueType": "string",
    "cycle": [
      "typeof PredicateSemantics",
      "{}",
      "typeof PredicateSemantics"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 36083,
    "type": "typeof ProcessLevel",
    "declaration": "ProcessLevel (src/compiler/transformers/taggedTemplate.ts:29:1)",
    "valueType": "string",
    "cycle": [
      "typeof ProcessLevel",
      "{}",
      "typeof ProcessLevel"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 42501,
    "type": "typeof ProgramUpdateLevel",
    "declaration": "ProgramUpdateLevel (src/compiler/watchUtilities.ts:381:1)",
    "valueType": "string",
    "cycle": [
      "typeof ProgramUpdateLevel",
      "{}",
      "typeof ProgramUpdateLevel"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25833,
    "type": "typeof PropertyLikeParse",
    "declaration": "PropertyLikeParse (src/compiler/parser.ts:8879:9)",
    "valueType": "string",
    "cycle": [
      "typeof PropertyLikeParse",
      "{}",
      "typeof PropertyLikeParse"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 32408,
    "type": "typeof RecursionFlags",
    "declaration": "RecursionFlags (src/compiler/checker.ts:1373:1)",
    "valueType": "string",
    "cycle": [
      "typeof RecursionFlags",
      "{}",
      "typeof RecursionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 33246,
    "type": "typeof ReferenceHint",
    "declaration": "ReferenceHint (src/compiler/checker.ts:1160:1)",
    "valueType": "string",
    "cycle": [
      "typeof ReferenceHint",
      "{}",
      "typeof ReferenceHint"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 15323,
    "type": "typeof RegularExpressionFlags",
    "declaration": "RegularExpressionFlags (src/compiler/types.ts:2794:1)",
    "valueType": "string",
    "cycle": [
      "typeof RegularExpressionFlags",
      "{}",
      "typeof RegularExpressionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1623,
    "type": "typeof RelationComparisonResult",
    "declaration": "RelationComparisonResult (src/compiler/types.ts:917:1)",
    "valueType": "string",
    "cycle": [
      "typeof RelationComparisonResult",
      "{}",
      "typeof RelationComparisonResult"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28433,
    "type": "typeof RelativePreference",
    "declaration": "RelativePreference (src/compiler/moduleSpecifiers.ts:164:1)",
    "valueType": "string",
    "cycle": [
      "typeof RelativePreference",
      "{}",
      "typeof RelativePreference"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1624,
    "type": "typeof ScriptKind",
    "declaration": "ScriptKind (src/compiler/types.ts:7663:1)",
    "valueType": "string",
    "cycle": [
      "typeof ScriptKind",
      "{}",
      "typeof ScriptKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 6706,
    "type": "typeof ScriptTarget",
    "declaration": "ScriptTarget (src/compiler/types.ts:7682:1)",
    "valueType": "string",
    "cycle": [
      "typeof ScriptTarget",
      "{}",
      "typeof ScriptTarget"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 40751,
    "type": "typeof SeenPackageName",
    "declaration": "SeenPackageName (src/compiler/program.ts:2385:9)",
    "valueType": "string",
    "cycle": [
      "typeof SeenPackageName",
      "{}",
      "typeof SeenPackageName"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1625,
    "type": "typeof SignatureCheckMode",
    "declaration": "SignatureCheckMode (src/compiler/checker.ts:1357:1)",
    "valueType": "string",
    "cycle": [
      "typeof SignatureCheckMode",
      "{}",
      "typeof SignatureCheckMode"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1626,
    "type": "typeof SignatureFlags",
    "declaration": "SignatureFlags (src/compiler/types.ts:7006:1)",
    "valueType": "string",
    "cycle": [
      "typeof SignatureFlags",
      "{}",
      "typeof SignatureFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25354,
    "type": "typeof SignatureFlags",
    "declaration": "SignatureFlags (src/compiler/parser.ts:405:1)",
    "valueType": "string",
    "cycle": [
      "typeof SignatureFlags",
      "{}",
      "typeof SignatureFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 41257,
    "type": "typeof SignatureInfo",
    "declaration": "SignatureInfo (src/compiler/builderState.ts:56:1)",
    "valueType": "string",
    "cycle": [
      "typeof SignatureInfo",
      "{}",
      "typeof SignatureInfo"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28724,
    "type": "typeof SignatureKind",
    "declaration": "SignatureKind (src/compiler/types.ts:6999:1)",
    "valueType": "string",
    "cycle": [
      "typeof SignatureKind",
      "{}",
      "typeof SignatureKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1627,
    "type": "typeof SnippetKind",
    "declaration": "SnippetKind (src/compiler/types.ts:8364:1)",
    "valueType": "string",
    "cycle": [
      "typeof SnippetKind",
      "{}",
      "typeof SnippetKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25036,
    "type": "typeof SpeculationKind",
    "declaration": "SpeculationKind (src/compiler/parser.ts:414:1)",
    "valueType": "string",
    "cycle": [
      "typeof SpeculationKind",
      "{}",
      "typeof SpeculationKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37986,
    "type": "typeof SpreadSegmentKind",
    "declaration": "SpreadSegmentKind (src/compiler/transformers/es2015.ts:472:1)",
    "valueType": "string",
    "cycle": [
      "typeof SpreadSegmentKind",
      "{}",
      "typeof SpreadSegmentKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 21379,
    "type": "typeof States",
    "declaration": "States (src/compiler/utilities.ts:10977:5)",
    "valueType": "string",
    "cycle": [
      "typeof States",
      "{}",
      "typeof States"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 43798,
    "type": "typeof StatisticType",
    "declaration": "StatisticType (src/compiler/executeCommandLine.ts:99:1)",
    "valueType": "string",
    "cycle": [
      "typeof StatisticType",
      "{}",
      "typeof StatisticType"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 40163,
    "type": "typeof StructureIsReused",
    "declaration": "StructureIsReused (src/compiler/types.ts:4967:1)",
    "valueType": "string",
    "cycle": [
      "typeof StructureIsReused",
      "{}",
      "typeof StructureIsReused"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28725,
    "type": "typeof SymbolAccessibility",
    "declaration": "SymbolAccessibility (src/compiler/types.ts:5679:1)",
    "valueType": "string",
    "cycle": [
      "typeof SymbolAccessibility",
      "{}",
      "typeof SymbolAccessibility"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1628,
    "type": "typeof SymbolFlags",
    "declaration": "SymbolFlags (src/compiler/types.ts:5947:1)",
    "valueType": "string",
    "cycle": [
      "typeof SymbolFlags",
      "{}",
      "typeof SymbolFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28726,
    "type": "typeof SymbolFormatFlags",
    "declaration": "SymbolFormatFlags (src/compiler/types.ts:5626:1)",
    "valueType": "string",
    "cycle": [
      "typeof SymbolFormatFlags",
      "{}",
      "typeof SymbolFormatFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1630,
    "type": "typeof SyntaxKind",
    "declaration": "SyntaxKind (src/compiler/types.ts:40:1)",
    "valueType": "string",
    "cycle": [
      "typeof SyntaxKind",
      "{}",
      "typeof SyntaxKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39033,
    "type": "typeof SyntaxKindFeatureFlags",
    "declaration": "SyntaxKindFeatureFlags (src/compiler/transformer.ts:111:1)",
    "valueType": "string",
    "cycle": [
      "typeof SyntaxKindFeatureFlags",
      "{}",
      "typeof SyntaxKindFeatureFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 39260,
    "type": "typeof TempFlags",
    "declaration": "TempFlags (src/compiler/emitter.ts:6345:1)",
    "valueType": "string",
    "cycle": [
      "typeof TempFlags",
      "{}",
      "typeof TempFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28728,
    "type": "typeof Ternary",
    "declaration": "Ternary (src/compiler/types.ts:7156:1)",
    "valueType": "string",
    "cycle": [
      "typeof Ternary",
      "{}",
      "typeof Ternary"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 15324,
    "type": "typeof TokenFlags",
    "declaration": "TokenFlags (src/compiler/types.ts:2815:1)",
    "valueType": "string",
    "cycle": [
      "typeof TokenFlags",
      "{}",
      "typeof TokenFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 38980,
    "type": "typeof TransformationState",
    "declaration": "TransformationState (src/compiler/transformer.ts:104:1)",
    "valueType": "string",
    "cycle": [
      "typeof TransformationState",
      "{}",
      "typeof TransformationState"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1631,
    "type": "typeof TransformFlags",
    "declaration": "TransformFlags (src/compiler/types.ts:8223:1)",
    "valueType": "string",
    "cycle": [
      "typeof TransformFlags",
      "{}",
      "typeof TransformFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 25477,
    "type": "typeof Tristate",
    "declaration": "Tristate (src/compiler/parser.ts:8784:5)",
    "valueType": "string",
    "cycle": [
      "typeof Tristate",
      "{}",
      "typeof Tristate"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1632,
    "type": "typeof TypeFacts",
    "declaration": "TypeFacts (src/compiler/checker.ts:1233:1)",
    "valueType": "string",
    "cycle": [
      "typeof TypeFacts",
      "{}",
      "typeof TypeFacts"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1633,
    "type": "typeof TypeFlags",
    "declaration": "TypeFlags (src/compiler/types.ts:6335:1)",
    "valueType": "string",
    "cycle": [
      "typeof TypeFlags",
      "{}",
      "typeof TypeFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28729,
    "type": "typeof TypeFormatFlags",
    "declaration": "TypeFormatFlags (src/compiler/types.ts:5584:1)",
    "valueType": "string",
    "cycle": [
      "typeof TypeFormatFlags",
      "{}",
      "typeof TypeFormatFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1634,
    "type": "typeof TypeMapKind",
    "declaration": "TypeMapKind (src/compiler/types.ts:7087:1)",
    "valueType": "string",
    "cycle": [
      "typeof TypeMapKind",
      "{}",
      "typeof TypeMapKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 18717,
    "type": "typeof TypePredicateKind",
    "declaration": "TypePredicateKind (src/compiler/types.ts:5686:1)",
    "valueType": "string",
    "cycle": [
      "typeof TypePredicateKind",
      "{}",
      "typeof TypePredicateKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28730,
    "type": "typeof TypeReferenceSerializationKind",
    "declaration": "TypeReferenceSerializationKind (src/compiler/types.ts:5843:1)",
    "valueType": "string",
    "cycle": [
      "typeof TypeReferenceSerializationKind",
      "{}",
      "typeof TypeReferenceSerializationKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 36126,
    "type": "typeof TypeScriptSubstitutionFlags",
    "declaration": "TypeScriptSubstitutionFlags (src/compiler/transformers/ts.ts:211:1)",
    "valueType": "string",
    "cycle": [
      "typeof TypeScriptSubstitutionFlags",
      "{}",
      "typeof TypeScriptSubstitutionFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 30872,
    "type": "typeof TypeSystemPropertyName",
    "declaration": "TypeSystemPropertyName (src/compiler/checker.ts:1329:1)",
    "valueType": "string",
    "cycle": [
      "typeof TypeSystemPropertyName",
      "{}",
      "typeof TypeSystemPropertyName"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 28731,
    "type": "typeof UnionReduction",
    "declaration": "UnionReduction (src/compiler/types.ts:5505:1)",
    "valueType": "string",
    "cycle": [
      "typeof UnionReduction",
      "{}",
      "typeof UnionReduction"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 34467,
    "type": "typeof UnusedKind",
    "declaration": "UnusedKind (src/compiler/checker.ts:1405:1)",
    "valueType": "string",
    "cycle": [
      "typeof UnusedKind",
      "{}",
      "typeof UnusedKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 42824,
    "type": "typeof UpToDateStatusType",
    "declaration": "UpToDateStatusType (src/compiler/tsbuild.ts:10:1)",
    "valueType": "string",
    "cycle": [
      "typeof UpToDateStatusType",
      "{}",
      "typeof UpToDateStatusType"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 37215,
    "type": "typeof UsingKind",
    "declaration": "UsingKind (src/compiler/transformers/esnext.ts:64:1)",
    "valueType": "string",
    "cycle": [
      "typeof UsingKind",
      "{}",
      "typeof UsingKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 1636,
    "type": "typeof VarianceFlags",
    "declaration": "VarianceFlags (src/compiler/types.ts:6681:1)",
    "valueType": "string",
    "cycle": [
      "typeof VarianceFlags",
      "{}",
      "typeof VarianceFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 26094,
    "type": "typeof WatchDirectoryFlags",
    "declaration": "WatchDirectoryFlags (src/compiler/types.ts:7722:1)",
    "valueType": "string",
    "cycle": [
      "typeof WatchDirectoryFlags",
      "{}",
      "typeof WatchDirectoryFlags"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 8172,
    "type": "typeof WatchDirectoryKind",
    "declaration": "WatchDirectoryKind (src/compiler/types.ts:7398:1)",
    "valueType": "string",
    "cycle": [
      "typeof WatchDirectoryKind",
      "{}",
      "typeof WatchDirectoryKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 8173,
    "type": "typeof WatchFileKind",
    "declaration": "WatchFileKind (src/compiler/types.ts:7389:1)",
    "valueType": "string",
    "cycle": [
      "typeof WatchFileKind",
      "{}",
      "typeof WatchFileKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 40097,
    "type": "typeof WatchLogLevel",
    "declaration": "WatchLogLevel (src/compiler/watchUtilities.ts:677:1)",
    "valueType": "string",
    "cycle": [
      "typeof WatchLogLevel",
      "{}",
      "typeof WatchLogLevel"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 32719,
    "type": "typeof WideningKind",
    "declaration": "WideningKind (src/compiler/checker.ts:1224:1)",
    "valueType": "string",
    "cycle": [
      "typeof WideningKind",
      "{}",
      "typeof WideningKind"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule"
    ],
    "files": []
  },
  {
    "id": 21268,
    "type": "Uint16Array<ArrayBuffer>",
    "declaration": "Uint16Array (external/node_modules/typescript/lib/lib.es5.d.ts:3008:1); Uint16Array (external/node_modules/typescript/lib/lib.es5.d.ts:3284:13); Uint16Array (external/node_modules/typescript/lib/lib.es2015.core.d.ts:577:1); Uint16Array (external/node_modules/typescript/lib/lib.es2015.iterable.d.ts:420:1); Uint16Array (external/node_modules/typescript/lib/lib.es2015.symbol.wellknown.d.ts:293:1); Uint16Array (external/node_modules/typescript/lib/lib.es2016.array.include.d.ts:71:1)",
    "valueType": "number",
    "cycle": [
      "Uint16Array<ArrayBuffer>",
      "{}",
      "Uint16Array<ArrayBuffer>"
    ],
    "literalSites": 0,
    "unambiguousLiteralSites": 0,
    "unsureLiteralSites": 0,
    "unsure": true,
    "unsureReasons": [
      "index signature is outside recordElement pure mutable string-index rule",
      "cycle uses source structural-view candidate approximation",
      "index signature is outside recordElement pure mutable string-index rule",
      "external declaration/host type, lowering may replace its storage"
    ],
    "files": []
  },
  {
    "id": 6029,
    "type": "WatchOptions",
    "declaration": "WatchOptions (src/compiler/types.ts:7585:1)",
    "valueType": "CompilerOptionsValue",
    "cycle": [
      "WatchOptions",
      "{}",
      "WatchOptions"
    ],
    "literalSites": 12,
    "unambiguousLiteralSites": 11,
    "unsureLiteralSites": 1,
    "unsure": true,
    "unsureReasons": [
      "cycle uses source structural-view candidate approximation"
    ],
    "files": [
      {
        "file": "src/compiler/sys.ts",
        "evaluationOrder": 10,
        "literalCount": 7
      },
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "literalCount": 4
      },
      {
        "file": "src/compiler/watchUtilities.ts",
        "evaluationOrder": 64,
        "literalCount": 1
      }
    ]
  }
]
```

| File | Evaluation index | Unique candidate literal sites |
| --- | --- | --- |
| src/compiler/utilities.ts | 15 | 57 |
| src/compiler/commandLineParser.ts | 27 | 35 |
| src/compiler/checker.ts | 33 | 18 |
| src/compiler/program.ts | 65 | 12 |
| src/compiler/sys.ts | 10 | 7 |
| src/compiler/builder.ts | 69 | 6 |
| src/compiler/parser.ts | 26 | 5 |
| src/compiler/emitter.ts | 63 | 3 |
| src/compiler/moduleNameResolver.ts | 28 | 3 |
| src/compiler/scanner.ts | 13 | 3 |
| src/compiler/transformer.ts | 62 | 3 |
| src/compiler/core.ts | 1 | 2 |
| src/compiler/moduleSpecifiers.ts | 31 | 2 |
| src/compiler/symbolWalker.ts | 30 | 2 |
| src/compiler/transformers/declarations/diagnostics.ts | 60 | 2 |
| src/compiler/tsbuildPublic.ts | 75 | 2 |
| src/compiler/binder.ts | 29 | 1 |
| src/compiler/executeCommandLine.ts | 76 | 1 |
| src/compiler/tracing.ts | 8 | 1 |
| src/compiler/transformers/jsx.ts | 52 | 1 |
| src/compiler/types.ts | 9 | 1 |
| src/compiler/watchPublic.ts | 73 | 1 |
| src/compiler/watchUtilities.ts | 64 | 1 |

## Ownership cycles and literal counts

The following records have a cycle using only reproduced ownership edges. Each path starts at the record and returns to the same checker identity. These are type-level possibilities, not proof that a runtime allocation is cyclic or that all other lowering checks accept it.

```json
[
  {
    "id": 26785,
    "type": "AdamicJsonRecoveryObject",
    "declaration": "AdamicJsonRecoveryObject (src/compiler/commandLineParser.ts:2439:1)",
    "valueType": "AdamicJsonRecoveryValue | undefined",
    "cycle": [
      "AdamicJsonRecoveryObject",
      "AdamicJsonRecoveryValue | undefined",
      "AdamicJsonRecoveryObject"
    ],
    "literalSites": 4,
    "unambiguousLiteralSites": 3,
    "unsureLiteralSites": 1,
    "files": [
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "count": 4
      }
    ]
  },
  {
    "id": 2675,
    "type": "Args",
    "declaration": "Args (src/compiler/tracing.ts:54:5)",
    "valueType": "string | number | boolean | Args | readonly (string | number | boolean | Args | null | undefined)[] | null | undefined",
    "cycle": [
      "Args",
      "string | number | boolean | Args | readonly (string | number | boolean | Args | null | undefined)[] | null | undefined",
      "Args"
    ],
    "literalSites": 37,
    "unambiguousLiteralSites": 37,
    "unsureLiteralSites": 0,
    "files": [
      {
        "file": "src/compiler/checker.ts",
        "evaluationOrder": 33,
        "count": 16
      },
      {
        "file": "src/compiler/program.ts",
        "evaluationOrder": 65,
        "count": 12
      },
      {
        "file": "src/compiler/emitter.ts",
        "evaluationOrder": 63,
        "count": 3
      },
      {
        "file": "src/compiler/transformer.ts",
        "evaluationOrder": 62,
        "count": 2
      },
      {
        "file": "src/compiler/binder.ts",
        "evaluationOrder": 29,
        "count": 1
      },
      {
        "file": "src/compiler/commandLineParser.ts",
        "evaluationOrder": 27,
        "count": 1
      },
      {
        "file": "src/compiler/parser.ts",
        "evaluationOrder": 26,
        "count": 1
      },
      {
        "file": "src/compiler/tracing.ts",
        "evaluationOrder": 8,
        "count": 1
      }
    ]
  }
]
```

## Top files for the ownership-cycle records

| File | Evaluation index | Unique candidate sites | Unambiguous sites | Unsure sites |
| --- | --- | --- | --- | --- |
| src/compiler/checker.ts | 33 | 16 | 16 | 0 |
| src/compiler/program.ts | 65 | 12 | 12 | 0 |
| src/compiler/commandLineParser.ts | 27 | 5 | 4 | 1 |
| src/compiler/emitter.ts | 63 | 3 | 3 | 0 |
| src/compiler/transformer.ts | 62 | 2 | 2 | 0 |
| src/compiler/binder.ts | 29 | 1 | 1 | 0 |
| src/compiler/parser.ts | 26 | 1 | 1 | 0 |
| src/compiler/tracing.ts | 8 | 1 | 1 | 0 |

## Unsure column

| Type identity | Record type | Literal sites | Ownership cycle | Source-view cycle | Unsure reason |
| --- | --- | --- | --- | --- | --- |
| 26051 | { [index: string]: string \| { value: string; pos: number; end: number; }; } | 1 | false | true | cycle uses source structural-view candidate approximation |
| 26077 | { [index: string]: string \| undefined; } | 2 | false | true | cycle uses source structural-view candidate approximation |
| 26082 | { [index: string]: string \| undefined; } | 1 | false | true | cycle uses source structural-view candidate approximation |
| 7974 | { [K in keyof T]: PragmaArgTypeOptional<T[K], T[K] extends { name: infer TName; } ? TName extends string ? TName : never : never>; } | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 26922 | { [k: string]: CompilerOptionsValue; } | 0 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 26842 | { [key: string]: unknown; } | 0 | false | true | cycle uses source structural-view candidate approximation |
| 26844 | { [key: string]: unknown; } | 0 | false | true | cycle uses source structural-view candidate approximation |
| 44887 | { [s: string]: "General_Category" \| "Script" \| "Script_Extensions"; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45754 | { [s: string]: ImportsNotUsedAsValues; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45939 | { [s: string]: IntrinsicTypeKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45730 | { [s: string]: JsxEmit; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45042 | { [s: string]: Map<string, string[]> \| Map<string, never[]> \| Map<string, never[] \| string[]>; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45769 | { [s: string]: ModuleDetectionKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45749 | { [s: string]: ModuleKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45757 | { [s: string]: ModuleResolutionKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45052 | { [s: string]: never[] \| string[]; } | 3 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45048 | { [s: string]: never[]; } | 10 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45765 | { [s: string]: NewLineKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 47239 | { [s: string]: number; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45740 | { [s: string]: PollingWatchKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45746 | { [s: string]: ScriptTarget; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45089 | { [s: string]: string; } | 2 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45045 | { [s: string]: string[]; } | 41 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 44879 | { [s: string]: SyntaxKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45932 | { [s: string]: TypeFacts.TypeofNEString \| TypeFacts.TypeofNENumber \| TypeFacts.TypeofNEBigInt \| TypeFacts.TypeofNEBoolean \| TypeFacts.TypeofNESymbol \| TypeFacts.TypeofNEObject \| TypeFacts.TypeofNEFunction \| TypeFacts.NEUndefined; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45737 | { [s: string]: WatchDirectoryKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 45734 | { [s: string]: WatchFileKind; } | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 43674 | { [value: string]: string[]; } | 1 | false | true | cycle uses source structural-view candidate approximation |
| 44503 | { [x: string]: string; } | 0 | false | false | inventory-only type, not reached from reproduced source uses |
| 15391 | { [x: string]: SyntaxKind.BreakKeyword \| SyntaxKind.CaseKeyword \| SyntaxKind.CatchKeyword \| SyntaxKind.ClassKeyword \| SyntaxKind.ConstKeyword \| SyntaxKind.ContinueKeyword \| SyntaxKind.DebuggerKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.DeleteKeyword \| SyntaxKind.DoKeyword \| SyntaxKind.ElseKeyword \| SyntaxKind.EnumKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.ExtendsKeyword \| SyntaxKind.FalseKeyword \| SyntaxKind.FinallyKeyword \| SyntaxKind.ForKeyword \| SyntaxKind.FunctionKeyword \| SyntaxKind.IfKeyword \| SyntaxKind.ImportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.InstanceOfKeyword \| SyntaxKind.NewKeyword \| SyntaxKind.NullKeyword \| SyntaxKind.ReturnKeyword \| SyntaxKind.SuperKeyword \| SyntaxKind.SwitchKeyword \| SyntaxKind.ThisKeyword \| SyntaxKind.ThrowKeyword \| SyntaxKind.TrueKeyword \| SyntaxKind.TryKeyword \| SyntaxKind.TypeOfKeyword \| SyntaxKind.VarKeyword \| SyntaxKind.VoidKeyword \| SyntaxKind.WhileKeyword \| SyntaxKind.WithKeyword \| SyntaxKind.ImplementsKeyword \| SyntaxKind.InterfaceKeyword \| SyntaxKind.LetKeyword \| SyntaxKind.PackageKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| SyntaxKind.StaticKeyword \| SyntaxKind.YieldKeyword \| SyntaxKind.AbstractKeyword \| SyntaxKind.AccessorKeyword \| SyntaxKind.AsKeyword \| SyntaxKind.AssertsKeyword \| SyntaxKind.AssertKeyword \| SyntaxKind.AnyKeyword \| SyntaxKind.AsyncKeyword \| SyntaxKind.AwaitKeyword \| SyntaxKind.BooleanKeyword \| SyntaxKind.ConstructorKeyword \| SyntaxKind.DeclareKeyword \| SyntaxKind.GetKeyword \| SyntaxKind.InferKeyword \| SyntaxKind.IntrinsicKeyword \| SyntaxKind.IsKeyword \| SyntaxKind.KeyOfKeyword \| SyntaxKind.ModuleKeyword \| SyntaxKind.NamespaceKeyword \| SyntaxKind.NeverKeyword \| SyntaxKind.OutKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.RequireKeyword \| SyntaxKind.NumberKeyword \| SyntaxKind.ObjectKeyword \| SyntaxKind.SatisfiesKeyword \| SyntaxKind.SetKeyword \| SyntaxKind.StringKeyword \| SyntaxKind.SymbolKeyword \| SyntaxKind.TypeKeyword \| SyntaxKind.UndefinedKeyword \| SyntaxKind.UniqueKeyword \| SyntaxKind.UnknownKeyword \| SyntaxKind.UsingKeyword \| SyntaxKind.FromKeyword \| SyntaxKind.GlobalKeyword \| SyntaxKind.BigIntKeyword \| SyntaxKind.OverrideKeyword \| SyntaxKind.OfKeyword \| SyntaxKind.DeferKeyword; abstract: SyntaxKind.AbstractKeyword; accessor: SyntaxKind.AccessorKeyword; any: SyntaxKind.AnyKeyword; as: SyntaxKind.AsKeyword; asserts: SyntaxKind.AssertsKeyword; assert: SyntaxKind.AssertKeyword; bigint: SyntaxKind.BigIntKeyword; boolean: SyntaxKind.BooleanKeyword; break: SyntaxKind.BreakKeyword; case: SyntaxKind.CaseKeyword; catch: SyntaxKind.CatchKeyword; class: SyntaxKind.ClassKeyword; continue: SyntaxKind.ContinueKeyword; const: SyntaxKind.ConstKeyword; debugger: SyntaxKind.DebuggerKeyword; declare: SyntaxKind.DeclareKeyword; default: SyntaxKind.DefaultKeyword; defer: SyntaxKind.DeferKeyword; delete: SyntaxKind.DeleteKeyword; do: SyntaxKind.DoKeyword; else: SyntaxKind.ElseKeyword; enum: SyntaxKind.EnumKeyword; export: SyntaxKind.ExportKeyword; extends: SyntaxKind.ExtendsKeyword; false: SyntaxKind.FalseKeyword; finally: SyntaxKind.FinallyKeyword; for: SyntaxKind.ForKeyword; from: SyntaxKind.FromKeyword; function: SyntaxKind.FunctionKeyword; get: SyntaxKind.GetKeyword; if: SyntaxKind.IfKeyword; implements: SyntaxKind.ImplementsKeyword; import: SyntaxKind.ImportKeyword; in: SyntaxKind.InKeyword; infer: SyntaxKind.InferKeyword; instanceof: SyntaxKind.InstanceOfKeyword; interface: SyntaxKind.InterfaceKeyword; intrinsic: SyntaxKind.IntrinsicKeyword; is: SyntaxKind.IsKeyword; keyof: SyntaxKind.KeyOfKeyword; let: SyntaxKind.LetKeyword; module: SyntaxKind.ModuleKeyword; namespace: SyntaxKind.NamespaceKeyword; never: SyntaxKind.NeverKeyword; new: SyntaxKind.NewKeyword; null: SyntaxKind.NullKeyword; number: SyntaxKind.NumberKeyword; object: SyntaxKind.ObjectKeyword; package: SyntaxKind.PackageKeyword; private: SyntaxKind.PrivateKeyword; protected: SyntaxKind.ProtectedKeyword; public: SyntaxKind.PublicKeyword; override: SyntaxKind.OverrideKeyword; out: SyntaxKind.OutKeyword; readonly: SyntaxKind.ReadonlyKeyword; require: SyntaxKind.RequireKeyword; global: SyntaxKind.GlobalKeyword; return: SyntaxKind.ReturnKeyword; satisfies: SyntaxKind.SatisfiesKeyword; set: SyntaxKind.SetKeyword; static: SyntaxKind.StaticKeyword; string: SyntaxKind.StringKeyword; super: SyntaxKind.SuperKeyword; switch: SyntaxKind.SwitchKeyword; symbol: SyntaxKind.SymbolKeyword; this: SyntaxKind.ThisKeyword; throw: SyntaxKind.ThrowKeyword; true: SyntaxKind.TrueKeyword; try: SyntaxKind.TryKeyword; type: SyntaxKind.TypeKeyword; typeof: SyntaxKind.TypeOfKeyword; undefined: SyntaxKind.UndefinedKeyword; unique: SyntaxKind.UniqueKeyword; unknown: SyntaxKind.UnknownKeyword; using: SyntaxKind.UsingKeyword; var: SyntaxKind.VarKeyword; void: SyntaxKind.VoidKeyword; while: SyntaxKind.WhileKeyword; with: SyntaxKind.WithKeyword; yield: SyntaxKind.YieldKeyword; async: SyntaxKind.AsyncKeyword; await: SyntaxKind.AwaitKeyword; of: SyntaxKind.OfKeyword; } | 0 | false | true | enumerationRecordStorageType needs the lowerer enumeration provenance proof; cycle uses source structural-view candidate approximation; enumerationRecordStorageType needs the lowerer enumeration provenance proof |
| 44888 | ArrayLike<"General_Category" \| "Script" \| "Script_Extensions"> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45755 | ArrayLike<ImportsNotUsedAsValues> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45940 | ArrayLike<IntrinsicTypeKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45731 | ArrayLike<JsxEmit> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45043 | ArrayLike<Map<string, string[]> \| Map<string, never[]> \| Map<string, never[] \| string[]>> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45770 | ArrayLike<ModuleDetectionKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45750 | ArrayLike<ModuleKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45758 | ArrayLike<ModuleResolutionKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45053 | ArrayLike<never[] \| string[]> | 3 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45049 | ArrayLike<never[]> | 10 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45766 | ArrayLike<NewLineKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 47238 | ArrayLike<number> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45741 | ArrayLike<PollingWatchKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45747 | ArrayLike<ScriptTarget> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45046 | ArrayLike<string[]> | 41 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45090 | ArrayLike<string> | 2 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 44880 | ArrayLike<SyntaxKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45933 | ArrayLike<TypeFacts.TypeofNEString \| TypeFacts.TypeofNENumber \| TypeFacts.TypeofNEBigInt \| TypeFacts.TypeofNEBoolean \| TypeFacts.TypeofNESymbol \| TypeFacts.TypeofNEObject \| TypeFacts.TypeofNEFunction \| TypeFacts.NEUndefined> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45738 | ArrayLike<WatchDirectoryKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 45735 | ArrayLike<WatchFileKind> | 1 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 8717 | Buffer<ArrayBuffer> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 8774 | Buffer<ArrayBufferLike> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 8775 | Buffer<TArrayBuffer> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 26764 | BuildOptions | 0 | false | true | cycle uses source structural-view candidate approximation |
| 4596 | CompilerOptions | 25 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation |
| 871 | IArguments | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 3605 | JSDocArray | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 15390 | MapLike<KeywordSyntaxKind> | 1 | false | true | cycle uses source structural-view candidate approximation |
| 27317 | MapLike<MapLike<string[]>> | 0 | false | true | cycle uses source structural-view candidate approximation |
| 28601 | MapLike<readonly string[]> | 0 | false | true | cycle uses source structural-view candidate approximation |
| 27007 | MapLike<string \| number \| boolean \| string[] \| (string \| number)[] \| MapLike<string[]> \| PluginImport[] \| ProjectReference[] \| null \| undefined> | 0 | false | true | cycle uses source structural-view candidate approximation |
| 27002 | MapLike<string \| number \| boolean \| string[] \| (string \| number)[] \| TsConfigSourceFile \| MapLike<string[]> \| PluginImport[] \| ProjectReference[] \| null \| undefined> | 1 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation |
| 6012 | MapLike<string[]> | 1 | false | true | cycle uses source structural-view candidate approximation |
| 20620 | MapLike<string> | 0 | false | true | cycle uses source structural-view candidate approximation |
| 28382 | MapLike<Symbol> | 1 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation |
| 1089 | MapLike<T \| undefined> | 1 | false | true | reachable generic/any template requires lowering instantiation or representation proof; cycle uses source structural-view candidate approximation |
| 4 | MapLike<T> | 0 | false | false | reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity |
| 883 | MapLike<T> | 0 | false | false | reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity |
| 887 | MapLike<T> | 0 | false | false | reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity |
| 892 | MapLike<T> | 0 | false | false | reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity |
| 927 | MapLike<T> | 0 | false | false | reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity |
| 28379 | MapLike<Type> | 1 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation |
| 21266 | MapLike<unknown> | 0 | false | true | cycle uses source structural-view candidate approximation |
| 6091 | MapLike<WatchDirectoryFlags> | 1 | false | true | cycle uses source structural-view candidate approximation |
| 25708 | MissingList<ClassElement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25714 | MissingList<EnumMember> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25304 | MissingList<Node> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25384 | MissingList<ParameterDeclaration> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25603 | MissingList<Statement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25296 | MissingList<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 25301 | MissingList<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 25309 | MissingList<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 25401 | MissingList<TypeElement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25675 | MissingList<VariableDeclaration> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 3883 | ModifiersArray | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 35075 | MutableNodeArray<Modifier> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 23081 | MutableNodeArray<Node> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 3828 | MutableNodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 22406 | MutableNodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 3989 | NodeArray<ArrayBindingElement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 7187 | NodeArray<AssertEntry> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 3987 | NodeArray<BindingElement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 2416 | NodeArray<CaseOrDefaultClause> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39716 | NodeArray<Child> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 39726 | NodeArray<Child> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 39732 | NodeArray<Child> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 19422 | NodeArray<ClassElement \| ObjectLiteralElement \| TypeElement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4336 | NodeArray<ClassElement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 23600 | NodeArray<Declaration> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39685 | NodeArray<Decorator> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4345 | NodeArray<EnumMember> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4386 | NodeArray<ExportSpecifier> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4191 | NodeArray<Expression> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 38318 | NodeArray<ExpressionStatement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4341 | NodeArray<ExpressionWithTypeArguments> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4335 | NodeArray<HeritageClause> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4376 | NodeArray<ImportAttribute> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4384 | NodeArray<ImportSpecifier> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4444 | NodeArray<JSDocComment> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 22776 | NodeArray<JSDocParameterTag> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39670 | NodeArray<JSDocPropertyLikeTag> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4442 | NodeArray<JSDocTag> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39671 | NodeArray<JSDocTemplateTag> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4587 | NodeArray<JsonObjectExpressionStatement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25557 | NodeArray<JsxAttribute \| JsxSpreadAttribute> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4271 | NodeArray<JsxAttributeLike> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4263 | NodeArray<JsxChild> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25554 | NodeArray<JsxElement \| JsxSelfClosingElement \| JsxFragment \| JsxExpression \| JsxText> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 3933 | NodeArray<Modifier> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 3959 | NodeArray<ModifierLike> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 23914 | NodeArray<never> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1995 | NodeArray<Node> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25286 | NodeArray<NonNullable<T>> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 25291 | NodeArray<NonNullable<T>> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 21345 | NodeArray<ObjectLiteralElement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 31389 | NodeArray<ObjectLiteralElementLike \| JsxAttributeLike> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19425 | NodeArray<ObjectLiteralElementLike> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 3942 | NodeArray<ParameterDeclaration> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4299 | NodeArray<Statement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1697 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 1698 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 1721 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 1732 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 4196 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 6874 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 7573 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 7580 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 18359 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 18880 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 21629 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 21639 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 21642 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 21645 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 21650 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 22404 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 23020 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 23025 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 23030 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 23907 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 23910 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 23913 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 25157 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 25258 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 25282 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 25298 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 25302 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 25308 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 39246 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 39250 | NodeArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 4051 | NodeArray<TemplateLiteralTypeSpan> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4185 | NodeArray<TemplateSpan> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 7551 | NodeArray<TIn> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 7561 | NodeArray<TIn> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 35376 | NodeArray<TIn> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 35386 | NodeArray<TIn> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 7556 | NodeArray<TOut> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 35381 | NodeArray<TOut> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 4031 | NodeArray<TypeElement> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 4032 | NodeArray<TypeNode \| NamedTupleMember> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 3944 | NodeArray<TypeNode> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 6607 | NodeArray<TypeParameterDeclaration \| TypeNode> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 3941 | NodeArray<TypeParameterDeclaration> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1723 | NodeArray<U> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 1735 | NodeArray<U> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 3956 | NodeArray<VariableDeclaration> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; ownership cycle crosses a record requiring a storage proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 5276 | NodeArray<VariableDeclarationInitializedTo<RequireOrImportCall>> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 8776 | NonSharedBuffer | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 26721 | OptionsBase | 2 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation |
| 38663 | Partial<Record<SyntaxKind, DiagnosticMessage>> | 2 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20723 | Pick<CompilerOptions, "allowImportingTsExtensions" \| "rewriteRelativeImportExtensions"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20847 | Pick<CompilerOptions, "allowJs" \| "checkJs"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20783 | Pick<CompilerOptions, "allowSyntheticDefaultImports"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20911 | Pick<CompilerOptions, "alwaysStrict"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20830 | Pick<CompilerOptions, "composite" \| "incremental"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20815 | Pick<CompilerOptions, "declaration" \| "composite"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20838 | Pick<CompilerOptions, "declaration" \| "declarationMap" \| "composite"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20777 | Pick<CompilerOptions, "esModuleInterop"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20823 | Pick<CompilerOptions, "isolatedModules" \| "preserveConstEnums" \| "verbatimModuleSyntax"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20771 | Pick<CompilerOptions, "isolatedModules" \| "verbatimModuleSyntax"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 39084 | Pick<CompilerOptions, "jsx"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20861 | Pick<CompilerOptions, "noImplicitAny" \| "strict"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20869 | Pick<CompilerOptions, "noImplicitThis" \| "strict"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20890 | Pick<CompilerOptions, "strict" \| "strictBindCallApply"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20904 | Pick<CompilerOptions, "strict" \| "strictBuiltinIteratorReturn"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20883 | Pick<CompilerOptions, "strict" \| "strictFunctionTypes"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20876 | Pick<CompilerOptions, "strict" \| "strictNullChecks"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20897 | Pick<CompilerOptions, "strict" \| "strictPropertyInitialization"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20917 | Pick<CompilerOptions, "strict" \| "useUnknownInCatchVariables"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20762 | Pick<CompilerOptions, "target" \| "module" \| "moduleDetection"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20806 | Pick<CompilerOptions, "target" \| "module" \| "moduleResolution" \| "resolveJsonModule"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20800 | Pick<CompilerOptions, "target" \| "module" \| "moduleResolution" \| "resolvePackageJsonExports" \| "resolvePackageJsonImports"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20790 | Pick<CompilerOptions, "target" \| "module" \| "moduleResolution" \| "resolvePackageJsonExports"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20750 | Pick<CompilerOptions, "target" \| "module" \| "moduleResolution"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20853 | Pick<CompilerOptions, "target" \| "module" \| "useDefineForClassFields"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20738 | Pick<CompilerOptions, "target" \| "module"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 20729 | Pick<CompilerOptions, "target"> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 8267 | PollingIntervalQueue | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 8663 | ProcessEnv | 0 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 44742 | ProcessVersions | 0 | false | true | enumerationRecordStorageType needs the lowerer enumeration provenance proof; cycle uses source structural-view candidate approximation; enumerationRecordStorageType needs the lowerer enumeration provenance proof; external declaration/host type, lowering may replace its storage |
| 44066 | PropertyDescriptorMap | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 41937 | Readonly<PathPathComponents> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 1816 | Readonly<Record<string, string \| number>> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 44128 | Record<"__debugToString", unknown> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 44116 | Record<"__debugTypeToString", unknown> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 44131 | Record<"antecedent", unknown> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 47408 | Record<"category", unknown> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1918 | Record<"CheckMode", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 44165 | Record<"code", unknown> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1878 | Record<"EmitFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1906 | Record<"FlowFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 39816 | Record<"GeneratedIdentifierFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 44166 | Record<"message", unknown> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 47407 | Record<"messageText", unknown> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1866 | Record<"ModifierFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1860 | Record<"NodeCheckFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1854 | Record<"NodeFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1900 | Record<"ObjectFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1912 | Record<"RelationComparisonResult", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1848 | Record<"ScriptKind", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1924 | Record<"SignatureCheckMode", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1894 | Record<"SignatureFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1842 | Record<"SnippetKind", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1882 | Record<"SymbolFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1836 | Record<"SyntaxKind", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1872 | Record<"TransformFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1930 | Record<"TypeFacts", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 1888 | Record<"TypeFlags", Readonly<Record<string, string \| number>>> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; index signature is outside recordElement pure mutable string-index rule; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 6702 | Record<LanugageFeatures, ScriptTarget> | 1 | false | true | finite mapped record requires Go utility-alias/storage proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 2352 | Record<number, FlowGraphNode> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 40880 | Record<number, string> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 20697 | Record<string, ("out" \| "target" \| "build" \| "version" \| "project" \| "all" \| "allowImportingTsExtensions" \| "allowJs" \| "allowNonTsExtensions" \| "allowArbitraryExtensions" \| "allowSyntheticDefaultImports" \| "allowUmdGlobalAccess" \| "allowUnreachableCode" \| "allowUnusedLabels" \| "alwaysStrict" \| "baseUrl" \| "charset" \| "checkJs" \| "configFilePath" \| "configFile" \| "customConditions" \| "declaration" \| "declarationMap" \| "emitDeclarationOnly" \| "declarationDir" \| "diagnostics" \| "extendedDiagnostics" \| "disableSizeLimit" \| "disableSourceOfProjectReferenceRedirect" \| "disableSolutionSearching" \| "disableReferencedProjectLoad" \| "downlevelIteration" \| "emitBOM" \| "emitDecoratorMetadata" \| "exactOptionalPropertyTypes" \| "experimentalDecorators" \| "forceConsistentCasingInFileNames" \| "generateCpuProfile" \| "generateTrace" \| "help" \| "ignoreDeprecations" \| "importHelpers" \| "importsNotUsedAsValues" \| "init" \| "inlineSourceMap" \| "inlineSources" \| "isolatedModules" \| "isolatedDeclarations" \| "jsx" \| "keyofStringsOnly" \| "lib" \| "libReplacement" \| "listEmittedFiles" \| "listFiles" \| "explainFiles" \| "listFilesOnly" \| "locale" \| "mapRoot" \| "maxNodeModuleJsDepth" \| "module" \| "moduleResolution" \| "moduleSuffixes" \| "moduleDetection" \| "newLine" \| "noEmit" \| "noCheck" \| "noEmitForJsFiles" \| "noEmitHelpers" \| "noEmitOnError" \| "noErrorTruncation" \| "noFallthroughCasesInSwitch" \| "noImplicitAny" \| "noImplicitReturns" \| "noImplicitThis" \| "noStrictGenericChecks" \| "noUnusedLocals" \| "noUnusedParameters" \| "noImplicitUseStrict" \| "noPropertyAccessFromIndexSignature" \| "assumeChangesOnlyAffectDirectDependencies" \| "noLib" \| "noResolve" \| "noDtsResolution" \| "noUncheckedIndexedAccess" \| "outDir" \| "outFile" \| "paths" \| "pathsBasePath" \| "plugins" \| "preserveConstEnums" \| "noImplicitOverride" \| "preserveSymlinks" \| "preserveValueImports" \| "preserveWatchOutput" \| "pretty" \| "reactNamespace" \| "jsxFactory" \| "jsxFragmentFactory" \| "jsxImportSource" \| "composite" \| "incremental" \| "tsBuildInfoFile" \| "removeComments" \| "resolvePackageJsonExports" \| "resolvePackageJsonImports" \| "rewriteRelativeImportExtensions" \| "rootDir" \| "rootDirs" \| "skipLibCheck" \| "skipDefaultLibCheck" \| "sourceMap" \| "sourceRoot" \| "strict" \| "strictFunctionTypes" \| "strictBindCallApply" \| "strictNullChecks" \| "strictPropertyInitialization" \| "strictBuiltinIteratorReturn" \| "stableTypeOrdering" \| "stripInternal" \| "suppressExcessPropertyErrors" \| "suppressImplicitAnyIndexErrors" \| "suppressOutputPathCheck" \| "traceResolution" \| "useUnknownInCatchVariables" \| "noUncheckedSideEffectImports" \| "resolveJsonModule" \| "types" \| "typeRoots" \| "verbatimModuleSyntax" \| "erasableSyntaxOnly" \| "watch" \| "esModuleInterop" \| "showConfig" \| "ignoreConfig" \| "useDefineForClassFields" \| "tscBuild")[]> | 0 | false | false | inventory-only type, not reached from reproduced source uses; external declaration/host type, lowering may replace its storage |
| 20919 | Record<string, { dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; }> | 0 | false | true | closure environments / captured cells are lowering identities; cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 28665 | Record<string, any> | 0 | false | false | reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; external declaration/host type, lowering may replace its storage |
| 26915 | Record<string, CompilerOptionsValue> | 1 | false | true | cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 1818 | Record<string, string \| number> | 0 | false | false | inventory-only type, not reached from reproduced source uses; external declaration/host type, lowering may replace its storage |
| 1074 | Record<string, T[]> | 1 | false | true | reachable generic/any template requires lowering instantiation or representation proof; cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 38664 | Record<SyntaxKind, DiagnosticMessage> | 0 | false | false | finite mapped record requires Go utility-alias/storage proof; inventory-only type, not reached from reproduced source uses; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 24900 | Record<SyntaxKind, ForEachChildFunction<any>> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; reachable generic/any template requires lowering instantiation or representation proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 35464 | Record<SyntaxKind, VisitEachChildFunction<any> \| undefined> | 0 | false | true | finite mapped record requires Go utility-alias/storage proof; reachable generic/any template requires lowering instantiation or representation proof; cycle uses source structural-view candidate approximation; finite mapped record requires Go utility-alias/storage proof; external declaration/host type, lowering may replace its storage |
| 2508 | RegExpExecArray | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 21191 | RegExpMatchArray | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 39932 | SortedArray<Canonicalized> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19996 | SortedArray<Diagnostic> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19999 | SortedArray<DiagnosticWithLocation> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19997 | SortedArray<string> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 11 | SortedArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 612 | SortedArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 615 | SortedArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 1819 | SortedReadonlyArray<[number, string]> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39970 | SortedReadonlyArray<Canonicalized> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 43621 | SortedReadonlyArray<CommandLineOption> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 40273 | SortedReadonlyArray<Diagnostic> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39497 | SortedReadonlyArray<EmitHelper> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 35729 | SortedReadonlyArray<MappedPosition> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 40537 | SortedReadonlyArray<SourceFile> | 0 | false | true | closure environments / captured cells are lowering identities; index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 35732 | SortedReadonlyArray<SourceMappedPosition> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 628 | SortedReadonlyArray<string> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 6 | SortedReadonlyArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 597 | SortedReadonlyArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 634 | SortedReadonlyArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; inventory-only type, not reached from reproduced source uses; index signature is outside recordElement pure mutable string-index rule |
| 641 | SortedReadonlyArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 763 | SortedReadonlyArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 18031 | SortedReadonlyArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 40798 | SortedReadonlyArray<T> | 0 | false | false | index signature is outside recordElement pure mutable string-index rule; reachable generic/any template requires lowering instantiation or representation proof; generic template, no concrete allocation identity; index signature is outside recordElement pure mutable string-index rule |
| 44821 | Storage | 0 | false | true | reachable generic/any template requires lowering instantiation or representation proof; cycle uses source structural-view candidate approximation; external declaration/host type, lowering may replace its storage |
| 21417 | String | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 23336 | TemplateStringsArray | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 6033 | TypeAcquisition | 2 | false | true | cycle uses source structural-view candidate approximation |
| 28697 | typeof AccessFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 20400 | typeof AccessKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1569 | typeof AssertionLevel | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 17944 | typeof AssignmentDeclarationKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19752 | typeof AssignmentKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19936 | typeof Associativity | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 38199 | typeof BlockAction | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 41337 | typeof BuilderFileEmit | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 41676 | typeof BuilderProgramKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 43221 | typeof BuildResultFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 43172 | typeof BuildStep | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 34 | typeof CharacterCodes | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18640 | typeof CheckFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1570 | typeof CheckMode | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 36204 | typeof ClassFacts | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 36486 | typeof ClassFacts | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 36396 | typeof ClassPropertySubstitutionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 17363 | typeof ClassSetExpressionType | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 38212 | typeof CodeBlockKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 15317 | typeof CommentDirectiveType | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 36 | typeof Comparison | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 2435 | typeof Connection | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18642 | typeof ContainerFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28698 | typeof ContextFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 36993 | typeof ContextFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37761 | typeof CopyDirection | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 34245 | typeof DeclarationMeaning | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 34381 | typeof DeclarationSpaces | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 5980 | typeof DiagnosticCategory | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28704 | typeof ElementFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1571 | typeof EmitFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28705 | typeof EmitHint | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 38926 | typeof EmitOnly | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37680 | typeof ES2015SubstitutionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 36991 | typeof ES2017SubstitutionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 17206 | typeof EscapeSequenceScanningFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37084 | typeof ESNextSubstitutionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 38215 | typeof ExceptionBlockState | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 42216 | typeof ExitStatus | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 29433 | typeof ExpandingFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 27273 | typeof Extensions | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28706 | typeof ExternalEmitHelpers | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28406 | typeof FileIncludeKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 40160 | typeof FilePreprocessingDiagnosticsKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 8313 | typeof FileSystemEntryKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 8320 | typeof FileWatcherEventKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 35979 | typeof FlattenLevel | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1572 | typeof FlowFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19830 | typeof FunctionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 17951 | typeof GeneratedIdentifierFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19257 | typeof GetLiteralTextFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37091 | typeof HierarchyFacts | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37683 | typeof HierarchyFacts | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 26092 | typeof ImportsNotUsedAsValues | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28709 | typeof IndexFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28710 | typeof IndexKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28711 | typeof InferenceFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28712 | typeof InferencePriority | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 38059 | typeof Instruction | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18655 | typeof InternalEmitFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28713 | typeof InternalNodeBuilderFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28714 | typeof IntersectionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 32387 | typeof IntersectionState | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 29009 | typeof IntrinsicTypeKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 43155 | typeof InvalidatedProjectKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25971 | typeof InvalidPosition | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 32290 | typeof IterationTypeKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 30952 | typeof IterationUse | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 15319 | typeof JSDocParsingMode | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25797 | typeof JSDocState | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18699 | typeof JsxEmit | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28718 | typeof JsxFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28719 | typeof JsxReferenceKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37747 | typeof Jump | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 15320 | typeof LanguageVariant | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 35344 | typeof LexicalEnvironmentFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39064 | typeof ListFormat | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1647 | typeof LogLevel | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37943 | typeof LoopOutParameterFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 30183 | typeof MappedTypeModifiers | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 31352 | typeof MappedTypeNameTypeKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28630 | typeof MatchingMode | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28720 | typeof MemberOverrideStatus | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 31510 | typeof MinArgumentCountFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1617 | typeof ModifierFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18701 | typeof ModuleDetectionKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 27922 | typeof ModuleInstanceState | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18702 | typeof ModuleKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18703 | typeof ModuleResolutionKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 21166 | typeof ModuleSpecifierEnding | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18704 | typeof NewLineKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28721 | typeof NodeBuilderFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1618 | typeof NodeCheckFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 2839 | typeof NodeFactoryFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1619 | typeof NodeFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 27373 | typeof NodeResolutionFeatures | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1622 | typeof ObjectFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 38136 | typeof OpCode | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 19969 | typeof OperatorPrecedence | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18024 | typeof OuterExpressionKinds | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25008 | typeof ParsingContext | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39298 | typeof PipelinePhase | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 8215 | typeof PollingInterval | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 8170 | typeof PollingWatchKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 7884 | typeof PragmaKindFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28723 | typeof PredicateSemantics | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 36083 | typeof ProcessLevel | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 42501 | typeof ProgramUpdateLevel | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25833 | typeof PropertyLikeParse | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 32408 | typeof RecursionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 33246 | typeof ReferenceHint | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 15323 | typeof RegularExpressionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1623 | typeof RelationComparisonResult | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28433 | typeof RelativePreference | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1624 | typeof ScriptKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 6706 | typeof ScriptTarget | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 40751 | typeof SeenPackageName | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1625 | typeof SignatureCheckMode | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1626 | typeof SignatureFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25354 | typeof SignatureFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 41257 | typeof SignatureInfo | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28724 | typeof SignatureKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1627 | typeof SnippetKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25036 | typeof SpeculationKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37986 | typeof SpreadSegmentKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 21379 | typeof States | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 43798 | typeof StatisticType | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 40163 | typeof StructureIsReused | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28725 | typeof SymbolAccessibility | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1628 | typeof SymbolFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28726 | typeof SymbolFormatFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1630 | typeof SyntaxKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39033 | typeof SyntaxKindFeatureFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 39260 | typeof TempFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28728 | typeof Ternary | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 15324 | typeof TokenFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 38980 | typeof TransformationState | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1631 | typeof TransformFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 25477 | typeof Tristate | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1632 | typeof TypeFacts | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1633 | typeof TypeFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28729 | typeof TypeFormatFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1634 | typeof TypeMapKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 18717 | typeof TypePredicateKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28730 | typeof TypeReferenceSerializationKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 36126 | typeof TypeScriptSubstitutionFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 30872 | typeof TypeSystemPropertyName | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 28731 | typeof UnionReduction | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 34467 | typeof UnusedKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 42824 | typeof UpToDateStatusType | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 37215 | typeof UsingKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 1636 | typeof VarianceFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 26094 | typeof WatchDirectoryFlags | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 8172 | typeof WatchDirectoryKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 8173 | typeof WatchFileKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 40097 | typeof WatchLogLevel | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 32719 | typeof WideningKind | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule |
| 21268 | Uint16Array<ArrayBuffer> | 0 | false | true | index signature is outside recordElement pure mutable string-index rule; cycle uses source structural-view candidate approximation; index signature is outside recordElement pure mutable string-index rule; external declaration/host type, lowering may replace its storage |
| 6029 | WatchOptions | 12 | false | true | cycle uses source structural-view candidate approximation |

Every unsure row's declaration, value type, candidate cycle path, and sites are retained in `counts.json.records`. Source-view-only paths use the approximate candidate universe described below. They are not ownership proofs.

## Acyclic controls by reproduced ownership edges

| Type identity | Record type | Value type | Declaration | Literal sites | Strict status |
| --- | --- | --- | --- | --- | --- |
| 26051 | { [index: string]: string \| { value: string; pos: number; end: number; }; } | string \| { value: string; pos: number; end: number; } | __type (src/compiler/parser.ts:10723:29) | 1 | unsure |
| 26077 | { [index: string]: string \| undefined; } | string \| undefined | __type (src/compiler/parser.ts:10781:87) | 2 | unsure |
| 26082 | { [index: string]: string \| undefined; } | string \| undefined | __type (src/compiler/parser.ts:10785:19) | 1 | unsure |
| 7974 | { [K in keyof T]: PragmaArgTypeOptional<T[K], T[K] extends { name: infer TName; } ? TName extends string ? TName : never : never>; } | { [x: string]: string; } | __type (src/compiler/types.ts:10334:104) | 0 | unsure |
| 26922 | { [k: string]: CompilerOptionsValue; } | CompilerOptionsValue | __type (external/node_modules/typescript/lib/lib.es2019.object.d.ts:24:73) | 0 | unsure |
| 26842 | { [key: string]: unknown; } | unknown | __type (src/compiler/commandLineParser.ts:2442:62) | 0 | unsure |
| 26844 | { [key: string]: unknown; } | unknown | __type (src/compiler/commandLineParser.ts:2447:58) | 0 | unsure |
| 44887 | { [s: string]: "General_Category" \| "Script" \| "Script_Extensions"; } | "General_Category" \| "Script" \| "Script_Extensions" | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45754 | { [s: string]: ImportsNotUsedAsValues; } | ImportsNotUsedAsValues | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45939 | { [s: string]: IntrinsicTypeKind; } | IntrinsicTypeKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45730 | { [s: string]: JsxEmit; } | JsxEmit | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45042 | { [s: string]: Map<string, string[]> \| Map<string, never[]> \| Map<string, never[] \| string[]>; } | Map<string, string[]> \| Map<string, never[]> \| Map<string, never[] \| string[]> | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45769 | { [s: string]: ModuleDetectionKind; } | ModuleDetectionKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45749 | { [s: string]: ModuleKind; } | ModuleKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45757 | { [s: string]: ModuleResolutionKind; } | ModuleResolutionKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45052 | { [s: string]: never[] \| string[]; } | never[] \| string[] | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 3 | unsure |
| 45048 | { [s: string]: never[]; } | never[] | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 10 | unsure |
| 45765 | { [s: string]: NewLineKind; } | NewLineKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 47239 | { [s: string]: number; } | number | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45740 | { [s: string]: PollingWatchKind; } | PollingWatchKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45746 | { [s: string]: ScriptTarget; } | ScriptTarget | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45089 | { [s: string]: string; } | string | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 2 | unsure |
| 45045 | { [s: string]: string[]; } | string[] | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 41 | unsure |
| 44879 | { [s: string]: SyntaxKind; } | SyntaxKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45932 | { [s: string]: TypeFacts.TypeofNEString \| TypeFacts.TypeofNENumber \| TypeFacts.TypeofNEBigInt \| TypeFacts.TypeofNEBoolean \| TypeFacts.TypeofNESymbol \| TypeFacts.TypeofNEObject \| TypeFacts.TypeofNEFunction \| TypeFacts.NEUndefined; } | TypeFacts.TypeofNEString \| TypeFacts.TypeofNENumber \| TypeFacts.TypeofNEBigInt \| TypeFacts.TypeofNEBoolean \| TypeFacts.TypeofNESymbol \| TypeFacts.TypeofNEObject \| TypeFacts.TypeofNEFunction \| TypeFacts.NEUndefined | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45737 | { [s: string]: WatchDirectoryKind; } | WatchDirectoryKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 45734 | { [s: string]: WatchFileKind; } | WatchFileKind | __type (external/node_modules/typescript/lib/lib.es2017.object.d.ts:34:19) | 1 | unsure |
| 43674 | { [value: string]: string[]; } | string[] | __type (src/compiler/executeCommandLine.ts:405:37) | 1 | unsure |
| 44503 | { [x: string]: string; } | string | __type (src/compiler/types.ts:10328:7) | 0 | unsure |
| 15391 | { [x: string]: SyntaxKind.BreakKeyword \| SyntaxKind.CaseKeyword \| SyntaxKind.CatchKeyword \| SyntaxKind.ClassKeyword \| SyntaxKind.ConstKeyword \| SyntaxKind.ContinueKeyword \| SyntaxKind.DebuggerKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.DeleteKeyword \| SyntaxKind.DoKeyword \| SyntaxKind.ElseKeyword \| SyntaxKind.EnumKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.ExtendsKeyword \| SyntaxKind.FalseKeyword \| SyntaxKind.FinallyKeyword \| SyntaxKind.ForKeyword \| SyntaxKind.FunctionKeyword \| SyntaxKind.IfKeyword \| SyntaxKind.ImportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.InstanceOfKeyword \| SyntaxKind.NewKeyword \| SyntaxKind.NullKeyword \| SyntaxKind.ReturnKeyword \| SyntaxKind.SuperKeyword \| SyntaxKind.SwitchKeyword \| SyntaxKind.ThisKeyword \| SyntaxKind.ThrowKeyword \| SyntaxKind.TrueKeyword \| SyntaxKind.TryKeyword \| SyntaxKind.TypeOfKeyword \| SyntaxKind.VarKeyword \| SyntaxKind.VoidKeyword \| SyntaxKind.WhileKeyword \| SyntaxKind.WithKeyword \| SyntaxKind.ImplementsKeyword \| SyntaxKind.InterfaceKeyword \| SyntaxKind.LetKeyword \| SyntaxKind.PackageKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| SyntaxKind.StaticKeyword \| SyntaxKind.YieldKeyword \| SyntaxKind.AbstractKeyword \| SyntaxKind.AccessorKeyword \| SyntaxKind.AsKeyword \| SyntaxKind.AssertsKeyword \| SyntaxKind.AssertKeyword \| SyntaxKind.AnyKeyword \| SyntaxKind.AsyncKeyword \| SyntaxKind.AwaitKeyword \| SyntaxKind.BooleanKeyword \| SyntaxKind.ConstructorKeyword \| SyntaxKind.DeclareKeyword \| SyntaxKind.GetKeyword \| SyntaxKind.InferKeyword \| SyntaxKind.IntrinsicKeyword \| SyntaxKind.IsKeyword \| SyntaxKind.KeyOfKeyword \| SyntaxKind.ModuleKeyword \| SyntaxKind.NamespaceKeyword \| SyntaxKind.NeverKeyword \| SyntaxKind.OutKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.RequireKeyword \| SyntaxKind.NumberKeyword \| SyntaxKind.ObjectKeyword \| SyntaxKind.SatisfiesKeyword \| SyntaxKind.SetKeyword \| SyntaxKind.StringKeyword \| SyntaxKind.SymbolKeyword \| SyntaxKind.TypeKeyword \| SyntaxKind.UndefinedKeyword \| SyntaxKind.UniqueKeyword \| SyntaxKind.UnknownKeyword \| SyntaxKind.UsingKeyword \| SyntaxKind.FromKeyword \| SyntaxKind.GlobalKeyword \| SyntaxKind.BigIntKeyword \| SyntaxKind.OverrideKeyword \| SyntaxKind.OfKeyword \| SyntaxKind.DeferKeyword; abstract: SyntaxKind.AbstractKeyword; accessor: SyntaxKind.AccessorKeyword; any: SyntaxKind.AnyKeyword; as: SyntaxKind.AsKeyword; asserts: SyntaxKind.AssertsKeyword; assert: SyntaxKind.AssertKeyword; bigint: SyntaxKind.BigIntKeyword; boolean: SyntaxKind.BooleanKeyword; break: SyntaxKind.BreakKeyword; case: SyntaxKind.CaseKeyword; catch: SyntaxKind.CatchKeyword; class: SyntaxKind.ClassKeyword; continue: SyntaxKind.ContinueKeyword; const: SyntaxKind.ConstKeyword; debugger: SyntaxKind.DebuggerKeyword; declare: SyntaxKind.DeclareKeyword; default: SyntaxKind.DefaultKeyword; defer: SyntaxKind.DeferKeyword; delete: SyntaxKind.DeleteKeyword; do: SyntaxKind.DoKeyword; else: SyntaxKind.ElseKeyword; enum: SyntaxKind.EnumKeyword; export: SyntaxKind.ExportKeyword; extends: SyntaxKind.ExtendsKeyword; false: SyntaxKind.FalseKeyword; finally: SyntaxKind.FinallyKeyword; for: SyntaxKind.ForKeyword; from: SyntaxKind.FromKeyword; function: SyntaxKind.FunctionKeyword; get: SyntaxKind.GetKeyword; if: SyntaxKind.IfKeyword; implements: SyntaxKind.ImplementsKeyword; import: SyntaxKind.ImportKeyword; in: SyntaxKind.InKeyword; infer: SyntaxKind.InferKeyword; instanceof: SyntaxKind.InstanceOfKeyword; interface: SyntaxKind.InterfaceKeyword; intrinsic: SyntaxKind.IntrinsicKeyword; is: SyntaxKind.IsKeyword; keyof: SyntaxKind.KeyOfKeyword; let: SyntaxKind.LetKeyword; module: SyntaxKind.ModuleKeyword; namespace: SyntaxKind.NamespaceKeyword; never: SyntaxKind.NeverKeyword; new: SyntaxKind.NewKeyword; null: SyntaxKind.NullKeyword; number: SyntaxKind.NumberKeyword; object: SyntaxKind.ObjectKeyword; package: SyntaxKind.PackageKeyword; private: SyntaxKind.PrivateKeyword; protected: SyntaxKind.ProtectedKeyword; public: SyntaxKind.PublicKeyword; override: SyntaxKind.OverrideKeyword; out: SyntaxKind.OutKeyword; readonly: SyntaxKind.ReadonlyKeyword; require: SyntaxKind.RequireKeyword; global: SyntaxKind.GlobalKeyword; return: SyntaxKind.ReturnKeyword; satisfies: SyntaxKind.SatisfiesKeyword; set: SyntaxKind.SetKeyword; static: SyntaxKind.StaticKeyword; string: SyntaxKind.StringKeyword; super: SyntaxKind.SuperKeyword; switch: SyntaxKind.SwitchKeyword; symbol: SyntaxKind.SymbolKeyword; this: SyntaxKind.ThisKeyword; throw: SyntaxKind.ThrowKeyword; true: SyntaxKind.TrueKeyword; try: SyntaxKind.TryKeyword; type: SyntaxKind.TypeKeyword; typeof: SyntaxKind.TypeOfKeyword; undefined: SyntaxKind.UndefinedKeyword; unique: SyntaxKind.UniqueKeyword; unknown: SyntaxKind.UnknownKeyword; using: SyntaxKind.UsingKeyword; var: SyntaxKind.VarKeyword; void: SyntaxKind.VoidKeyword; while: SyntaxKind.WhileKeyword; with: SyntaxKind.WithKeyword; yield: SyntaxKind.YieldKeyword; async: SyntaxKind.AsyncKeyword; await: SyntaxKind.AwaitKeyword; of: SyntaxKind.OfKeyword; } | SyntaxKind.BreakKeyword \| SyntaxKind.CaseKeyword \| SyntaxKind.CatchKeyword \| SyntaxKind.ClassKeyword \| SyntaxKind.ConstKeyword \| SyntaxKind.ContinueKeyword \| SyntaxKind.DebuggerKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.DeleteKeyword \| SyntaxKind.DoKeyword \| SyntaxKind.ElseKeyword \| SyntaxKind.EnumKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.ExtendsKeyword \| SyntaxKind.FalseKeyword \| SyntaxKind.FinallyKeyword \| SyntaxKind.ForKeyword \| SyntaxKind.FunctionKeyword \| SyntaxKind.IfKeyword \| SyntaxKind.ImportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.InstanceOfKeyword \| SyntaxKind.NewKeyword \| SyntaxKind.NullKeyword \| SyntaxKind.ReturnKeyword \| SyntaxKind.SuperKeyword \| SyntaxKind.SwitchKeyword \| SyntaxKind.ThisKeyword \| SyntaxKind.ThrowKeyword \| SyntaxKind.TrueKeyword \| SyntaxKind.TryKeyword \| SyntaxKind.TypeOfKeyword \| SyntaxKind.VarKeyword \| SyntaxKind.VoidKeyword \| SyntaxKind.WhileKeyword \| SyntaxKind.WithKeyword \| SyntaxKind.ImplementsKeyword \| SyntaxKind.InterfaceKeyword \| SyntaxKind.LetKeyword \| SyntaxKind.PackageKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| SyntaxKind.StaticKeyword \| SyntaxKind.YieldKeyword \| SyntaxKind.AbstractKeyword \| SyntaxKind.AccessorKeyword \| SyntaxKind.AsKeyword \| SyntaxKind.AssertsKeyword \| SyntaxKind.AssertKeyword \| SyntaxKind.AnyKeyword \| SyntaxKind.AsyncKeyword \| SyntaxKind.AwaitKeyword \| SyntaxKind.BooleanKeyword \| SyntaxKind.ConstructorKeyword \| SyntaxKind.DeclareKeyword \| SyntaxKind.GetKeyword \| SyntaxKind.InferKeyword \| SyntaxKind.IntrinsicKeyword \| SyntaxKind.IsKeyword \| SyntaxKind.KeyOfKeyword \| SyntaxKind.ModuleKeyword \| SyntaxKind.NamespaceKeyword \| SyntaxKind.NeverKeyword \| SyntaxKind.OutKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.RequireKeyword \| SyntaxKind.NumberKeyword \| SyntaxKind.ObjectKeyword \| SyntaxKind.SatisfiesKeyword \| SyntaxKind.SetKeyword \| SyntaxKind.StringKeyword \| SyntaxKind.SymbolKeyword \| SyntaxKind.TypeKeyword \| SyntaxKind.UndefinedKeyword \| SyntaxKind.UniqueKeyword \| SyntaxKind.UnknownKeyword \| SyntaxKind.UsingKeyword \| SyntaxKind.FromKeyword \| SyntaxKind.GlobalKeyword \| SyntaxKind.BigIntKeyword \| SyntaxKind.OverrideKeyword \| SyntaxKind.OfKeyword \| SyntaxKind.DeferKeyword | __object (src/compiler/scanner.ts:135:61) | 0 | unsure |
| 44888 | ArrayLike<"General_Category" \| "Script" \| "Script_Extensions"> | "General_Category" \| "Script" \| "Script_Extensions" | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45755 | ArrayLike<ImportsNotUsedAsValues> | ImportsNotUsedAsValues | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45940 | ArrayLike<IntrinsicTypeKind> | IntrinsicTypeKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45731 | ArrayLike<JsxEmit> | JsxEmit | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45043 | ArrayLike<Map<string, string[]> \| Map<string, never[]> \| Map<string, never[] \| string[]>> | Map<string, string[]> \| Map<string, never[]> \| Map<string, never[] \| string[]> | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45770 | ArrayLike<ModuleDetectionKind> | ModuleDetectionKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45750 | ArrayLike<ModuleKind> | ModuleKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45758 | ArrayLike<ModuleResolutionKind> | ModuleResolutionKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45053 | ArrayLike<never[] \| string[]> | never[] \| string[] | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 3 | unsure |
| 45049 | ArrayLike<never[]> | never[] | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 10 | unsure |
| 45766 | ArrayLike<NewLineKind> | NewLineKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 47238 | ArrayLike<number> | number | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45741 | ArrayLike<PollingWatchKind> | PollingWatchKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45747 | ArrayLike<ScriptTarget> | ScriptTarget | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45046 | ArrayLike<string[]> | string[] | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 41 | unsure |
| 45090 | ArrayLike<string> | string | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 2 | unsure |
| 44880 | ArrayLike<SyntaxKind> | SyntaxKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45933 | ArrayLike<TypeFacts.TypeofNEString \| TypeFacts.TypeofNENumber \| TypeFacts.TypeofNEBigInt \| TypeFacts.TypeofNEBoolean \| TypeFacts.TypeofNESymbol \| TypeFacts.TypeofNEObject \| TypeFacts.TypeofNEFunction \| TypeFacts.NEUndefined> | TypeFacts.TypeofNEString \| TypeFacts.TypeofNENumber \| TypeFacts.TypeofNEBigInt \| TypeFacts.TypeofNEBoolean \| TypeFacts.TypeofNESymbol \| TypeFacts.TypeofNEObject \| TypeFacts.TypeofNEFunction \| TypeFacts.NEUndefined | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45738 | ArrayLike<WatchDirectoryKind> | WatchDirectoryKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 45735 | ArrayLike<WatchFileKind> | WatchFileKind | ArrayLike (external/node_modules/typescript/lib/lib.es5.d.ts:1575:1) | 1 | unsure |
| 8717 | Buffer<ArrayBuffer> | number | Buffer (external/node_modules/@types/node/buffer.buffer.d.ts:356:9); Buffer (external/node_modules/@types/node/buffer.d.ts:255:9); Buffer (external/node_modules/@types/node/buffer.d.ts:1771:13) | 0 | unsure |
| 8774 | Buffer<ArrayBufferLike> | number | Buffer (external/node_modules/@types/node/buffer.buffer.d.ts:356:9); Buffer (external/node_modules/@types/node/buffer.d.ts:255:9); Buffer (external/node_modules/@types/node/buffer.d.ts:1771:13) | 0 | unsure |
| 26764 | BuildOptions | CompilerOptionsValue | BuildOptions (src/compiler/tsbuildPublic.ts:135:1) | 0 | unsure |
| 4596 | CompilerOptions | TsConfigSourceFile \| CompilerOptionsValue | CompilerOptions (src/compiler/types.ts:7414:1) | 25 | unsure |
| 871 | IArguments | any | IArguments (external/node_modules/typescript/lib/lib.es5.d.ts:402:1); IArguments (external/node_modules/typescript/lib/lib.es2015.iterable.d.ts:130:1) | 0 | unsure |
| 3605 | JSDocArray | JSDoc | JSDocArray (src/compiler/types.ts:963:1) | 0 | unsure |
| 15390 | MapLike<KeywordSyntaxKind> | KeywordSyntaxKind | MapLike (src/compiler/corePublic.ts:13:1) | 1 | unsure |
| 27317 | MapLike<MapLike<string[]>> | MapLike<string[]> | MapLike (src/compiler/corePublic.ts:13:1) | 0 | unsure |
| 28601 | MapLike<readonly string[]> | readonly string[] | MapLike (src/compiler/corePublic.ts:13:1) | 0 | unsure |
| 27007 | MapLike<string \| number \| boolean \| string[] \| (string \| number)[] \| MapLike<string[]> \| PluginImport[] \| ProjectReference[] \| null \| undefined> | string \| number \| boolean \| string[] \| (string \| number)[] \| MapLike<string[]> \| PluginImport[] \| ProjectReference[] \| null \| undefined | MapLike (src/compiler/corePublic.ts:13:1) | 0 | unsure |
| 27002 | MapLike<string \| number \| boolean \| string[] \| (string \| number)[] \| TsConfigSourceFile \| MapLike<string[]> \| PluginImport[] \| ProjectReference[] \| null \| undefined> | string \| number \| boolean \| string[] \| (string \| number)[] \| TsConfigSourceFile \| MapLike<string[]> \| PluginImport[] \| ProjectReference[] \| null \| undefined | MapLike (src/compiler/corePublic.ts:13:1) | 1 | unsure |
| 6012 | MapLike<string[]> | string[] | MapLike (src/compiler/corePublic.ts:13:1) | 1 | unsure |
| 20620 | MapLike<string> | string | MapLike (src/compiler/corePublic.ts:13:1) | 0 | unsure |
| 28382 | MapLike<Symbol> | Symbol | MapLike (src/compiler/corePublic.ts:13:1) | 1 | unsure |
| 1089 | MapLike<T \| undefined> | T \| undefined | MapLike (src/compiler/corePublic.ts:13:1) | 1 | unsure |
| 28379 | MapLike<Type> | Type | MapLike (src/compiler/corePublic.ts:13:1) | 1 | unsure |
| 21266 | MapLike<unknown> | unknown | MapLike (src/compiler/corePublic.ts:13:1) | 0 | unsure |
| 6091 | MapLike<WatchDirectoryFlags> | WatchDirectoryFlags | MapLike (src/compiler/corePublic.ts:13:1) | 1 | unsure |
| 25708 | MissingList<ClassElement> | ClassElement | MissingList (src/compiler/parser.ts:3565:5) | 0 | unsure |
| 25714 | MissingList<EnumMember> | EnumMember | MissingList (src/compiler/parser.ts:3565:5) | 0 | unsure |
| 25304 | MissingList<Node> | Node | MissingList (src/compiler/parser.ts:3565:5) | 0 | unsure |
| 25384 | MissingList<ParameterDeclaration> | ParameterDeclaration | MissingList (src/compiler/parser.ts:3565:5) | 0 | unsure |
| 25603 | MissingList<Statement> | Statement | MissingList (src/compiler/parser.ts:3565:5) | 0 | unsure |
| 25401 | MissingList<TypeElement> | TypeElement | MissingList (src/compiler/parser.ts:3565:5) | 0 | unsure |
| 25675 | MissingList<VariableDeclaration> | VariableDeclaration | MissingList (src/compiler/parser.ts:3565:5) | 0 | unsure |
| 3883 | ModifiersArray | Modifier | ModifiersArray (src/compiler/types.ts:1682:1) | 0 | unsure |
| 35075 | MutableNodeArray<Modifier> | Modifier | MutableNodeArray (src/compiler/types.ts:1584:1) | 0 | unsure |
| 23081 | MutableNodeArray<Node> | Node | MutableNodeArray (src/compiler/types.ts:1584:1) | 0 | unsure |
| 3989 | NodeArray<ArrayBindingElement> | ArrayBindingElement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 7187 | NodeArray<AssertEntry> | AssertEntry | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 3987 | NodeArray<BindingElement> | BindingElement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 2416 | NodeArray<CaseOrDefaultClause> | CaseOrDefaultClause | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 19422 | NodeArray<ClassElement \| ObjectLiteralElement \| TypeElement> | ClassElement \| ObjectLiteralElement \| TypeElement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4336 | NodeArray<ClassElement> | ClassElement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 23600 | NodeArray<Declaration> | Declaration | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 39685 | NodeArray<Decorator> | Decorator | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4345 | NodeArray<EnumMember> | EnumMember | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4386 | NodeArray<ExportSpecifier> | ExportSpecifier | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4191 | NodeArray<Expression> | Expression | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 38318 | NodeArray<ExpressionStatement> | ExpressionStatement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4341 | NodeArray<ExpressionWithTypeArguments> | ExpressionWithTypeArguments | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4335 | NodeArray<HeritageClause> | HeritageClause | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4376 | NodeArray<ImportAttribute> | ImportAttribute | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4384 | NodeArray<ImportSpecifier> | ImportSpecifier | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4444 | NodeArray<JSDocComment> | JSDocComment | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 22776 | NodeArray<JSDocParameterTag> | JSDocParameterTag | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 39670 | NodeArray<JSDocPropertyLikeTag> | JSDocPropertyLikeTag | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4442 | NodeArray<JSDocTag> | JSDocTag | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 39671 | NodeArray<JSDocTemplateTag> | JSDocTemplateTag | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4587 | NodeArray<JsonObjectExpressionStatement> | JsonObjectExpressionStatement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 25557 | NodeArray<JsxAttribute \| JsxSpreadAttribute> | JsxAttribute \| JsxSpreadAttribute | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4271 | NodeArray<JsxAttributeLike> | JsxAttributeLike | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4263 | NodeArray<JsxChild> | JsxChild | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 25554 | NodeArray<JsxElement \| JsxSelfClosingElement \| JsxFragment \| JsxExpression \| JsxText> | JsxElement \| JsxSelfClosingElement \| JsxFragment \| JsxExpression \| JsxText | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 3933 | NodeArray<Modifier> | Modifier | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 3959 | NodeArray<ModifierLike> | ModifierLike | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 23914 | NodeArray<never> | never | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 1995 | NodeArray<Node> | Node | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 25286 | NodeArray<NonNullable<T>> | NonNullable<T> | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 25291 | NodeArray<NonNullable<T>> | NonNullable<T> | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 21345 | NodeArray<ObjectLiteralElement> | ObjectLiteralElement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 31389 | NodeArray<ObjectLiteralElementLike \| JsxAttributeLike> | ObjectLiteralElementLike \| JsxAttributeLike | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 19425 | NodeArray<ObjectLiteralElementLike> | ObjectLiteralElementLike | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 3942 | NodeArray<ParameterDeclaration> | ParameterDeclaration | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4299 | NodeArray<Statement> | Statement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4051 | NodeArray<TemplateLiteralTypeSpan> | TemplateLiteralTypeSpan | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4185 | NodeArray<TemplateSpan> | TemplateSpan | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4031 | NodeArray<TypeElement> | TypeElement | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 4032 | NodeArray<TypeNode \| NamedTupleMember> | TypeNode \| NamedTupleMember | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 3944 | NodeArray<TypeNode> | TypeNode | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 6607 | NodeArray<TypeParameterDeclaration \| TypeNode> | TypeParameterDeclaration \| TypeNode | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 3941 | NodeArray<TypeParameterDeclaration> | TypeParameterDeclaration | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 3956 | NodeArray<VariableDeclaration> | VariableDeclaration | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 5276 | NodeArray<VariableDeclarationInitializedTo<RequireOrImportCall>> | VariableDeclarationInitializedTo<RequireOrImportCall> | NodeArray (src/compiler/types.ts:1589:1) | 0 | unsure |
| 8776 | NonSharedBuffer | number | NonSharedBuffer (external/node_modules/@types/node/buffer.buffer.d.ts:459:9) | 0 | unsure |
| 26721 | OptionsBase | TsConfigSourceFile \| CompilerOptionsValue | OptionsBase (src/compiler/commandLineParser.ts:1911:1) | 2 | unsure |
| 38663 | Partial<Record<SyntaxKind, DiagnosticMessage>> | DiagnosticMessage \| undefined | Partial (external/node_modules/typescript/lib/lib.es5.d.ts:1583:1) | 2 | unsure |
| 20723 | Pick<CompilerOptions, "allowImportingTsExtensions" \| "rewriteRelativeImportExtensions"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20847 | Pick<CompilerOptions, "allowJs" \| "checkJs"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20783 | Pick<CompilerOptions, "allowSyntheticDefaultImports"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20911 | Pick<CompilerOptions, "alwaysStrict"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20830 | Pick<CompilerOptions, "composite" \| "incremental"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20815 | Pick<CompilerOptions, "declaration" \| "composite"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20838 | Pick<CompilerOptions, "declaration" \| "declarationMap" \| "composite"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20777 | Pick<CompilerOptions, "esModuleInterop"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20823 | Pick<CompilerOptions, "isolatedModules" \| "preserveConstEnums" \| "verbatimModuleSyntax"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20771 | Pick<CompilerOptions, "isolatedModules" \| "verbatimModuleSyntax"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 39084 | Pick<CompilerOptions, "jsx"> | JsxEmit \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20861 | Pick<CompilerOptions, "noImplicitAny" \| "strict"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20869 | Pick<CompilerOptions, "noImplicitThis" \| "strict"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20890 | Pick<CompilerOptions, "strict" \| "strictBindCallApply"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20904 | Pick<CompilerOptions, "strict" \| "strictBuiltinIteratorReturn"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20883 | Pick<CompilerOptions, "strict" \| "strictFunctionTypes"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20876 | Pick<CompilerOptions, "strict" \| "strictNullChecks"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20897 | Pick<CompilerOptions, "strict" \| "strictPropertyInitialization"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20917 | Pick<CompilerOptions, "strict" \| "useUnknownInCatchVariables"> | boolean \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20762 | Pick<CompilerOptions, "target" \| "module" \| "moduleDetection"> | ModuleKind \| ScriptTarget \| ModuleDetectionKind \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20806 | Pick<CompilerOptions, "target" \| "module" \| "moduleResolution" \| "resolveJsonModule"> | boolean \| ModuleKind \| ScriptTarget \| ModuleResolutionKind \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20800 | Pick<CompilerOptions, "target" \| "module" \| "moduleResolution" \| "resolvePackageJsonExports" \| "resolvePackageJsonImports"> | boolean \| ModuleKind \| ScriptTarget \| ModuleResolutionKind \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20790 | Pick<CompilerOptions, "target" \| "module" \| "moduleResolution" \| "resolvePackageJsonExports"> | boolean \| ModuleKind \| ScriptTarget \| ModuleResolutionKind \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20750 | Pick<CompilerOptions, "target" \| "module" \| "moduleResolution"> | ModuleKind \| ScriptTarget \| ModuleResolutionKind \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20853 | Pick<CompilerOptions, "target" \| "module" \| "useDefineForClassFields"> | boolean \| ModuleKind \| ScriptTarget \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20738 | Pick<CompilerOptions, "target" \| "module"> | ModuleKind \| ScriptTarget \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 20729 | Pick<CompilerOptions, "target"> | ScriptTarget \| undefined | Pick (external/node_modules/typescript/lib/lib.es5.d.ts:1604:1) | 0 | unsure |
| 8267 | PollingIntervalQueue | WatchedFileWithUnchangedPolls | PollingIntervalQueue (src/compiler/sys.ts:255:5) | 0 | unsure |
| 8663 | ProcessEnv | string \| undefined | ProcessEnv (external/node_modules/@types/node/process.d.ts:468:13) | 0 | unsure |
| 44742 | ProcessVersions | string \| undefined | ProcessVersions (external/node_modules/@types/node/process.d.ts:269:13) | 0 | unsure |
| 44066 | PropertyDescriptorMap | PropertyDescriptor | PropertyDescriptorMap (external/node_modules/typescript/lib/lib.es5.d.ts:117:1) | 0 | unsure |
| 41937 | Readonly<PathPathComponents> | Path | Readonly (external/node_modules/typescript/lib/lib.es5.d.ts:1597:1) | 0 | unsure |
| 1816 | Readonly<Record<string, string \| number>> | string \| number | Readonly (external/node_modules/typescript/lib/lib.es5.d.ts:1597:1) | 0 | unsure |
| 44128 | Record<"__debugToString", unknown> | unknown | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 44116 | Record<"__debugTypeToString", unknown> | unknown | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 44131 | Record<"antecedent", unknown> | unknown | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 47408 | Record<"category", unknown> | unknown | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1918 | Record<"CheckMode", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 44165 | Record<"code", unknown> | unknown | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1878 | Record<"EmitFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1906 | Record<"FlowFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 39816 | Record<"GeneratedIdentifierFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 44166 | Record<"message", unknown> | unknown | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 47407 | Record<"messageText", unknown> | unknown | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1866 | Record<"ModifierFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1860 | Record<"NodeCheckFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1854 | Record<"NodeFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1900 | Record<"ObjectFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1912 | Record<"RelationComparisonResult", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1848 | Record<"ScriptKind", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1924 | Record<"SignatureCheckMode", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1894 | Record<"SignatureFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1842 | Record<"SnippetKind", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1882 | Record<"SymbolFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1836 | Record<"SyntaxKind", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1872 | Record<"TransformFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1930 | Record<"TypeFacts", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1888 | Record<"TypeFlags", Readonly<Record<string, string \| number>>> | Readonly<Record<string, string \| number>> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 6702 | Record<LanugageFeatures, ScriptTarget> | ScriptTarget | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 1 | unsure |
| 2352 | Record<number, FlowGraphNode> | FlowGraphNode | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 40880 | Record<number, string> | string | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 20697 | Record<string, ("out" \| "target" \| "build" \| "version" \| "project" \| "all" \| "allowImportingTsExtensions" \| "allowJs" \| "allowNonTsExtensions" \| "allowArbitraryExtensions" \| "allowSyntheticDefaultImports" \| "allowUmdGlobalAccess" \| "allowUnreachableCode" \| "allowUnusedLabels" \| "alwaysStrict" \| "baseUrl" \| "charset" \| "checkJs" \| "configFilePath" \| "configFile" \| "customConditions" \| "declaration" \| "declarationMap" \| "emitDeclarationOnly" \| "declarationDir" \| "diagnostics" \| "extendedDiagnostics" \| "disableSizeLimit" \| "disableSourceOfProjectReferenceRedirect" \| "disableSolutionSearching" \| "disableReferencedProjectLoad" \| "downlevelIteration" \| "emitBOM" \| "emitDecoratorMetadata" \| "exactOptionalPropertyTypes" \| "experimentalDecorators" \| "forceConsistentCasingInFileNames" \| "generateCpuProfile" \| "generateTrace" \| "help" \| "ignoreDeprecations" \| "importHelpers" \| "importsNotUsedAsValues" \| "init" \| "inlineSourceMap" \| "inlineSources" \| "isolatedModules" \| "isolatedDeclarations" \| "jsx" \| "keyofStringsOnly" \| "lib" \| "libReplacement" \| "listEmittedFiles" \| "listFiles" \| "explainFiles" \| "listFilesOnly" \| "locale" \| "mapRoot" \| "maxNodeModuleJsDepth" \| "module" \| "moduleResolution" \| "moduleSuffixes" \| "moduleDetection" \| "newLine" \| "noEmit" \| "noCheck" \| "noEmitForJsFiles" \| "noEmitHelpers" \| "noEmitOnError" \| "noErrorTruncation" \| "noFallthroughCasesInSwitch" \| "noImplicitAny" \| "noImplicitReturns" \| "noImplicitThis" \| "noStrictGenericChecks" \| "noUnusedLocals" \| "noUnusedParameters" \| "noImplicitUseStrict" \| "noPropertyAccessFromIndexSignature" \| "assumeChangesOnlyAffectDirectDependencies" \| "noLib" \| "noResolve" \| "noDtsResolution" \| "noUncheckedIndexedAccess" \| "outDir" \| "outFile" \| "paths" \| "pathsBasePath" \| "plugins" \| "preserveConstEnums" \| "noImplicitOverride" \| "preserveSymlinks" \| "preserveValueImports" \| "preserveWatchOutput" \| "pretty" \| "reactNamespace" \| "jsxFactory" \| "jsxFragmentFactory" \| "jsxImportSource" \| "composite" \| "incremental" \| "tsBuildInfoFile" \| "removeComments" \| "resolvePackageJsonExports" \| "resolvePackageJsonImports" \| "rewriteRelativeImportExtensions" \| "rootDir" \| "rootDirs" \| "skipLibCheck" \| "skipDefaultLibCheck" \| "sourceMap" \| "sourceRoot" \| "strict" \| "strictFunctionTypes" \| "strictBindCallApply" \| "strictNullChecks" \| "strictPropertyInitialization" \| "strictBuiltinIteratorReturn" \| "stableTypeOrdering" \| "stripInternal" \| "suppressExcessPropertyErrors" \| "suppressImplicitAnyIndexErrors" \| "suppressOutputPathCheck" \| "traceResolution" \| "useUnknownInCatchVariables" \| "noUncheckedSideEffectImports" \| "resolveJsonModule" \| "types" \| "typeRoots" \| "verbatimModuleSyntax" \| "erasableSyntaxOnly" \| "watch" \| "esModuleInterop" \| "showConfig" \| "ignoreConfig" \| "useDefineForClassFields" \| "tscBuild")[]> | ("out" \| "target" \| "build" \| "version" \| "project" \| "all" \| "allowImportingTsExtensions" \| "allowJs" \| "allowNonTsExtensions" \| "allowArbitraryExtensions" \| "allowSyntheticDefaultImports" \| "allowUmdGlobalAccess" \| "allowUnreachableCode" \| "allowUnusedLabels" \| "alwaysStrict" \| "baseUrl" \| "charset" \| "checkJs" \| "configFilePath" \| "configFile" \| "customConditions" \| "declaration" \| "declarationMap" \| "emitDeclarationOnly" \| "declarationDir" \| "diagnostics" \| "extendedDiagnostics" \| "disableSizeLimit" \| "disableSourceOfProjectReferenceRedirect" \| "disableSolutionSearching" \| "disableReferencedProjectLoad" \| "downlevelIteration" \| "emitBOM" \| "emitDecoratorMetadata" \| "exactOptionalPropertyTypes" \| "experimentalDecorators" \| "forceConsistentCasingInFileNames" \| "generateCpuProfile" \| "generateTrace" \| "help" \| "ignoreDeprecations" \| "importHelpers" \| "importsNotUsedAsValues" \| "init" \| "inlineSourceMap" \| "inlineSources" \| "isolatedModules" \| "isolatedDeclarations" \| "jsx" \| "keyofStringsOnly" \| "lib" \| "libReplacement" \| "listEmittedFiles" \| "listFiles" \| "explainFiles" \| "listFilesOnly" \| "locale" \| "mapRoot" \| "maxNodeModuleJsDepth" \| "module" \| "moduleResolution" \| "moduleSuffixes" \| "moduleDetection" \| "newLine" \| "noEmit" \| "noCheck" \| "noEmitForJsFiles" \| "noEmitHelpers" \| "noEmitOnError" \| "noErrorTruncation" \| "noFallthroughCasesInSwitch" \| "noImplicitAny" \| "noImplicitReturns" \| "noImplicitThis" \| "noStrictGenericChecks" \| "noUnusedLocals" \| "noUnusedParameters" \| "noImplicitUseStrict" \| "noPropertyAccessFromIndexSignature" \| "assumeChangesOnlyAffectDirectDependencies" \| "noLib" \| "noResolve" \| "noDtsResolution" \| "noUncheckedIndexedAccess" \| "outDir" \| "outFile" \| "paths" \| "pathsBasePath" \| "plugins" \| "preserveConstEnums" \| "noImplicitOverride" \| "preserveSymlinks" \| "preserveValueImports" \| "preserveWatchOutput" \| "pretty" \| "reactNamespace" \| "jsxFactory" \| "jsxFragmentFactory" \| "jsxImportSource" \| "composite" \| "incremental" \| "tsBuildInfoFile" \| "removeComments" \| "resolvePackageJsonExports" \| "resolvePackageJsonImports" \| "rewriteRelativeImportExtensions" \| "rootDir" \| "rootDirs" \| "skipLibCheck" \| "skipDefaultLibCheck" \| "sourceMap" \| "sourceRoot" \| "strict" \| "strictFunctionTypes" \| "strictBindCallApply" \| "strictNullChecks" \| "strictPropertyInitialization" \| "strictBuiltinIteratorReturn" \| "stableTypeOrdering" \| "stripInternal" \| "suppressExcessPropertyErrors" \| "suppressImplicitAnyIndexErrors" \| "suppressOutputPathCheck" \| "traceResolution" \| "useUnknownInCatchVariables" \| "noUncheckedSideEffectImports" \| "resolveJsonModule" \| "types" \| "typeRoots" \| "verbatimModuleSyntax" \| "erasableSyntaxOnly" \| "watch" \| "esModuleInterop" \| "showConfig" \| "ignoreConfig" \| "useDefineForClassFields" \| "tscBuild")[] | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 20919 | Record<string, { dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; }> | { dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; } | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 26915 | Record<string, CompilerOptionsValue> | CompilerOptionsValue | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 1 | unsure |
| 1818 | Record<string, string \| number> | string \| number | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 1074 | Record<string, T[]> | T[] | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 1 | unsure |
| 38664 | Record<SyntaxKind, DiagnosticMessage> | DiagnosticMessage | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 24900 | Record<SyntaxKind, ForEachChildFunction<any>> | ForEachChildFunction<any> | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 35464 | Record<SyntaxKind, VisitEachChildFunction<any> \| undefined> | VisitEachChildFunction<any> \| undefined | Record (external/node_modules/typescript/lib/lib.es5.d.ts:1611:1) | 0 | unsure |
| 2508 | RegExpExecArray | string | RegExpExecArray (external/node_modules/typescript/lib/lib.es5.d.ts:976:1); RegExpExecArray (external/node_modules/typescript/lib/lib.es2018.regexp.d.ts:23:1) | 0 | unsure |
| 21191 | RegExpMatchArray | string | RegExpMatchArray (external/node_modules/typescript/lib/lib.es5.d.ts:961:1); RegExpMatchArray (external/node_modules/typescript/lib/lib.es2018.regexp.d.ts:17:1) | 0 | unsure |
| 39932 | SortedArray<Canonicalized> | Canonicalized | SortedArray (src/compiler/corePublic.ts:21:1) | 0 | unsure |
| 19996 | SortedArray<Diagnostic> | Diagnostic | SortedArray (src/compiler/corePublic.ts:21:1) | 0 | unsure |
| 19999 | SortedArray<DiagnosticWithLocation> | DiagnosticWithLocation | SortedArray (src/compiler/corePublic.ts:21:1) | 0 | unsure |
| 19997 | SortedArray<string> | string | SortedArray (src/compiler/corePublic.ts:21:1) | 0 | unsure |
| 1819 | SortedReadonlyArray<[number, string]> | [number, string] | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 39970 | SortedReadonlyArray<Canonicalized> | Canonicalized | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 43621 | SortedReadonlyArray<CommandLineOption> | CommandLineOption | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 40273 | SortedReadonlyArray<Diagnostic> | Diagnostic | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 39497 | SortedReadonlyArray<EmitHelper> | EmitHelper | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 35729 | SortedReadonlyArray<MappedPosition> | MappedPosition | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 40537 | SortedReadonlyArray<SourceFile> | SourceFile | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 35732 | SortedReadonlyArray<SourceMappedPosition> | SourceMappedPosition | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 628 | SortedReadonlyArray<string> | string | SortedReadonlyArray (src/compiler/corePublic.ts:17:1) | 0 | unsure |
| 44821 | Storage | any | Storage (external/node_modules/@types/node/web-globals/storage.d.ts:16:5); Storage (external/node_modules/@types/node/web-globals/storage.d.ts:17:9) | 0 | unsure |
| 21417 | String | string | String (external/node_modules/typescript/lib/lib.es5.d.ts:408:1); String (external/node_modules/typescript/lib/lib.es5.d.ts:542:13); String (external/node_modules/typescript/lib/lib.es5.d.ts:4560:1); String (external/node_modules/typescript/lib/lib.es2015.core.d.ts:400:1); String (external/node_modules/typescript/lib/lib.es2015.iterable.d.ts:268:1); String (external/node_modules/typescript/lib/lib.es2015.symbol.wellknown.d.ts:233:1); String (external/node_modules/typescript/lib/lib.es2017.string.d.ts:17:1); String (external/node_modules/typescript/lib/lib.es2019.string.d.ts:17:1); String (external/node_modules/typescript/lib/lib.es2020.string.d.ts:21:1) | 0 | unsure |
| 23336 | TemplateStringsArray | string | TemplateStringsArray (external/node_modules/typescript/lib/lib.es5.d.ts:619:1) | 0 | unsure |
| 6033 | TypeAcquisition | CompilerOptionsValue | TypeAcquisition (src/compiler/types.ts:7596:1) | 2 | unsure |
| 28697 | typeof AccessFlags | string | AccessFlags (src/compiler/types.ts:6895:1) | 0 | unsure |
| 20400 | typeof AccessKind | string | AccessKind (src/compiler/utilities.ts:8089:1) | 0 | unsure |
| 1569 | typeof AssertionLevel | string | AssertionLevel (src/compiler/core.ts:1924:1) | 0 | unsure |
| 17944 | typeof AssignmentDeclarationKind | string | AssignmentDeclarationKind (src/compiler/types.ts:7197:1) | 0 | unsure |
| 19752 | typeof AssignmentKind | string | AssignmentKind (src/compiler/utilities.ts:4840:1) | 0 | unsure |
| 19936 | typeof Associativity | string | Associativity (src/compiler/utilities.ts:5629:1) | 0 | unsure |
| 38199 | typeof BlockAction | string | BlockAction (src/compiler/transformers/generators.ts:239:1) | 0 | unsure |
| 41337 | typeof BuilderFileEmit | string | BuilderFileEmit (src/compiler/builder.ts:186:1) | 0 | unsure |
| 41676 | typeof BuilderProgramKind | string | BuilderProgramKind (src/compiler/builder.ts:1593:1) | 0 | unsure |
| 43221 | typeof BuildResultFlags | string | BuildResultFlags (src/compiler/tsbuildPublic.ts:168:1) | 0 | unsure |
| 43172 | typeof BuildStep | string | BuildStep (src/compiler/tsbuildPublic.ts:924:1) | 0 | unsure |
| 34 | typeof CharacterCodes | string | CharacterCodes (src/compiler/types.ts:7852:1) | 0 | unsure |
| 18640 | typeof CheckFlags | string | CheckFlags (src/compiler/types.ts:6110:1) | 0 | unsure |
| 1570 | typeof CheckMode | string | CheckMode (src/compiler/checker.ts:1343:1) | 0 | unsure |
| 36204 | typeof ClassFacts | string | ClassFacts (src/compiler/transformers/ts.ts:219:1) | 0 | unsure |
| 36486 | typeof ClassFacts | string | ClassFacts (src/compiler/transformers/classFields.ts:333:1) | 0 | unsure |
| 36396 | typeof ClassPropertySubstitutionFlags | string | ClassPropertySubstitutionFlags (src/compiler/transformers/classFields.ts:230:1) | 0 | unsure |
| 17363 | typeof ClassSetExpressionType | string | ClassSetExpressionType (src/compiler/scanner.ts:1015:1) | 0 | unsure |
| 38212 | typeof CodeBlockKind | string | CodeBlockKind (src/compiler/transformers/generators.ts:245:1) | 0 | unsure |
| 15317 | typeof CommentDirectiveType | string | CommentDirectiveType (src/compiler/types.ts:4489:1) | 0 | unsure |
| 36 | typeof Comparison | string | Comparison (src/compiler/corePublic.ts:43:1) | 0 | unsure |
| 2435 | typeof Connection | string | Connection (src/compiler/debug.ts:896:9) | 0 | unsure |
| 18642 | typeof ContainerFlags | string | ContainerFlags (src/compiler/binder.ts:465:1) | 0 | unsure |
| 28698 | typeof ContextFlags | string | ContextFlags (src/compiler/types.ts:5520:1) | 0 | unsure |
| 36993 | typeof ContextFlags | string | ContextFlags (src/compiler/transformers/es2017.ts:112:1) | 0 | unsure |
| 37761 | typeof CopyDirection | string | CopyDirection (src/compiler/transformers/es2015.ts:278:1) | 0 | unsure |
| 34245 | typeof DeclarationMeaning | string | DeclarationMeaning (src/compiler/checker.ts:1415:1) | 0 | unsure |
| 34381 | typeof DeclarationSpaces | string | DeclarationSpaces (src/compiler/checker.ts:1425:1) | 0 | unsure |
| 5980 | typeof DiagnosticCategory | string | DiagnosticCategory (src/compiler/types.ts:7324:1) | 0 | unsure |
| 28704 | typeof ElementFlags | string | ElementFlags (src/compiler/types.ts:6702:1) | 0 | unsure |
| 1571 | typeof EmitFlags | string | EmitFlags (src/compiler/types.ts:8372:1) | 0 | unsure |
| 28705 | typeof EmitHint | string | EmitHint (src/compiler/types.ts:8592:1) | 0 | unsure |
| 38926 | typeof EmitOnly | string | EmitOnly (src/compiler/types.ts:4695:1) | 0 | unsure |
| 37680 | typeof ES2015SubstitutionFlags | string | ES2015SubstitutionFlags (src/compiler/transformers/es2015.ts:220:1) | 0 | unsure |
| 36991 | typeof ES2017SubstitutionFlags | string | ES2017SubstitutionFlags (src/compiler/transformers/es2017.ts:106:1) | 0 | unsure |
| 17206 | typeof EscapeSequenceScanningFlags | string | EscapeSequenceScanningFlags (src/compiler/scanner.ts:1002:1) | 0 | unsure |
| 37084 | typeof ESNextSubstitutionFlags | string | ESNextSubstitutionFlags (src/compiler/transformers/es2018.ts:115:1) | 0 | unsure |
| 38215 | typeof ExceptionBlockState | string | ExceptionBlockState (src/compiler/transformers/generators.ts:254:1) | 0 | unsure |
| 42216 | typeof ExitStatus | string | ExitStatus (src/compiler/types.ts:5017:1) | 0 | unsure |
| 29433 | typeof ExpandingFlags | string | ExpandingFlags (src/compiler/checker.ts:1393:1) | 0 | unsure |
| 27273 | typeof Extensions | string | Extensions (src/compiler/moduleNameResolver.ts:184:1) | 0 | unsure |
| 28706 | typeof ExternalEmitHelpers | string | ExternalEmitHelpers (src/compiler/types.ts:8543:1) | 0 | unsure |
| 28406 | typeof FileIncludeKind | string | FileIncludeKind (src/compiler/types.ts:4599:1) | 0 | unsure |
| 40160 | typeof FilePreprocessingDiagnosticsKind | string | FilePreprocessingDiagnosticsKind (src/compiler/types.ts:4664:1) | 0 | unsure |
| 8313 | typeof FileSystemEntryKind | string | FileSystemEntryKind (src/compiler/sys.ts:908:1) | 0 | unsure |
| 8320 | typeof FileWatcherEventKind | string | FileWatcherEventKind (src/compiler/sys.ts:87:1) | 0 | unsure |
| 35979 | typeof FlattenLevel | string | FlattenLevel (src/compiler/transformers/destructuring.ts:81:1) | 0 | unsure |
| 1572 | typeof FlowFlags | string | FlowFlags (src/compiler/types.ts:4150:1) | 0 | unsure |
| 19830 | typeof FunctionFlags | string | FunctionFlags (src/compiler/utilities.ts:5307:1) | 0 | unsure |
| 17951 | typeof GeneratedIdentifierFlags | string | GeneratedIdentifierFlags (src/compiler/types.ts:1685:1) | 0 | unsure |
| 19257 | typeof GetLiteralTextFlags | string | GetLiteralTextFlags (src/compiler/utilities.ts:1972:1) | 0 | unsure |
| 37091 | typeof HierarchyFacts | string | HierarchyFacts (src/compiler/transformers/es2018.ts:122:1) | 0 | unsure |
| 37683 | typeof HierarchyFacts | string | HierarchyFacts (src/compiler/transformers/es2015.ts:371:1) | 0 | unsure |
| 26092 | typeof ImportsNotUsedAsValues | string | ImportsNotUsedAsValues (src/compiler/types.ts:7643:1) | 0 | unsure |
| 28709 | typeof IndexFlags | string | IndexFlags (src/compiler/types.ts:6924:1) | 0 | unsure |
| 28710 | typeof IndexKind | string | IndexKind (src/compiler/types.ts:7071:1) | 0 | unsure |
| 28711 | typeof InferenceFlags | string | InferenceFlags (src/compiler/types.ts:7139:1) | 0 | unsure |
| 28712 | typeof InferencePriority | string | InferencePriority (src/compiler/types.ts:7105:1) | 0 | unsure |
| 38059 | typeof Instruction | string | Instruction (src/compiler/transformers/generators.ts:307:1) | 0 | unsure |
| 18655 | typeof InternalEmitFlags | string | InternalEmitFlags (src/compiler/types.ts:8406:1) | 0 | unsure |
| 28713 | typeof InternalNodeBuilderFlags | string | InternalNodeBuilderFlags (src/compiler/types.ts:5574:1) | 0 | unsure |
| 28714 | typeof IntersectionFlags | string | IntersectionFlags (src/compiler/types.ts:5512:1) | 0 | unsure |
| 32387 | typeof IntersectionState | string | IntersectionState (src/compiler/checker.ts:1367:1) | 0 | unsure |
| 29009 | typeof IntrinsicTypeKind | string | IntrinsicTypeKind (src/compiler/checker.ts:1438:1) | 0 | unsure |
| 43155 | typeof InvalidatedProjectKind | string | InvalidatedProjectKind (src/compiler/tsbuildPublic.ts:830:1) | 0 | unsure |
| 25971 | typeof InvalidPosition | string | InvalidPosition (src/compiler/parser.ts:10550:5) | 0 | unsure |
| 32290 | typeof IterationTypeKind | string | IterationTypeKind (src/compiler/checker.ts:1202:1) | 0 | unsure |
| 30952 | typeof IterationUse | string | IterationUse (src/compiler/checker.ts:1177:1) | 0 | unsure |
| 15319 | typeof JSDocParsingMode | string | JSDocParsingMode (src/compiler/types.ts:10295:1) | 0 | unsure |
| 25797 | typeof JSDocState | string | JSDocState (src/compiler/parser.ts:8872:9) | 0 | unsure |
| 18699 | typeof JsxEmit | string | JsxEmit (src/compiler/types.ts:7633:1) | 0 | unsure |
| 28718 | typeof JsxFlags | string | JsxFlags (src/compiler/types.ts:905:1) | 0 | unsure |
| 28719 | typeof JsxReferenceKind | string | JsxReferenceKind (src/compiler/types.ts:6993:1) | 0 | unsure |
| 37747 | typeof Jump | string | Jump (src/compiler/transformers/es2015.ts:283:1) | 0 | unsure |
| 15320 | typeof LanguageVariant | string | LanguageVariant (src/compiler/types.ts:7704:1) | 0 | unsure |
| 35344 | typeof LexicalEnvironmentFlags | string | LexicalEnvironmentFlags (src/compiler/types.ts:9563:1) | 0 | unsure |
| 39064 | typeof ListFormat | string | ListFormat (src/compiler/types.ts:10119:1) | 0 | unsure |
| 1647 | typeof LogLevel | string | LogLevel (src/compiler/debug.ts:97:1) | 0 | unsure |
| 37943 | typeof LoopOutParameterFlags | string | LoopOutParameterFlags (src/compiler/transformers/es2015.ts:272:1) | 0 | unsure |
| 30183 | typeof MappedTypeModifiers | string | MappedTypeModifiers (src/compiler/checker.ts:1380:1) | 0 | unsure |
| 31352 | typeof MappedTypeNameTypeKind | string | MappedTypeNameTypeKind (src/compiler/checker.ts:1387:1) | 0 | unsure |
| 28630 | typeof MatchingMode | string | MatchingMode (src/compiler/moduleSpecifiers.ts:1014:1) | 0 | unsure |
| 28720 | typeof MemberOverrideStatus | string | MemberOverrideStatus (src/compiler/types.ts:5498:1) | 0 | unsure |
| 31510 | typeof MinArgumentCountFlags | string | MinArgumentCountFlags (src/compiler/checker.ts:1432:1) | 0 | unsure |
| 1617 | typeof ModifierFlags | string | ModifierFlags (src/compiler/types.ts:851:1) | 0 | unsure |
| 18701 | typeof ModuleDetectionKind | string | ModuleDetectionKind (src/compiler/types.ts:7359:1) | 0 | unsure |
| 27922 | typeof ModuleInstanceState | string | ModuleInstanceState (src/compiler/binder.ts:315:1) | 0 | unsure |
| 18702 | typeof ModuleKind | string | ModuleKind (src/compiler/types.ts:7604:1) | 0 | unsure |
| 18703 | typeof ModuleResolutionKind | string | ModuleResolutionKind (src/compiler/types.ts:7336:1) | 0 | unsure |
| 21166 | typeof ModuleSpecifierEnding | string | ModuleSpecifierEnding (src/compiler/utilities.ts:10047:1) | 0 | unsure |
| 18704 | typeof NewLineKind | string | NewLineKind (src/compiler/types.ts:7649:1) | 0 | unsure |
| 28721 | typeof NodeBuilderFlags | string | NodeBuilderFlags (src/compiler/types.ts:5530:1) | 0 | unsure |
| 1618 | typeof NodeCheckFlags | string | NodeCheckFlags (src/compiler/types.ts:6224:1) | 0 | unsure |
| 2839 | typeof NodeFactoryFlags | string | NodeFactoryFlags (src/compiler/factory/nodeFactory.ts:466:1) | 0 | unsure |
| 1619 | typeof NodeFlags | string | NodeFlags (src/compiler/types.ts:782:1) | 0 | unsure |
| 27373 | typeof NodeResolutionFeatures | string | NodeResolutionFeatures (src/compiler/moduleNameResolver.ts:1687:1) | 0 | unsure |
| 1622 | typeof ObjectFlags | string | ObjectFlags (src/compiler/types.ts:6523:1) | 0 | unsure |
| 38136 | typeof OpCode | string | OpCode (src/compiler/transformers/generators.ts:222:1) | 0 | unsure |
| 19969 | typeof OperatorPrecedence | string | OperatorPrecedence (src/compiler/utilities.ts:5701:1) | 0 | unsure |
| 18024 | typeof OuterExpressionKinds | string | OuterExpressionKinds (src/compiler/types.ts:8644:1) | 0 | unsure |
| 25008 | typeof ParsingContext | string | ParsingContext (src/compiler/parser.ts:8754:5) | 0 | unsure |
| 39298 | typeof PipelinePhase | string | PipelinePhase (src/compiler/emitter.ts:1191:1) | 0 | unsure |
| 8215 | typeof PollingInterval | string | PollingInterval (src/compiler/sys.ts:102:1) | 0 | unsure |
| 8170 | typeof PollingWatchKind | string | PollingWatchKind (src/compiler/types.ts:7405:1) | 0 | unsure |
| 7884 | typeof PragmaKindFlags | string | PragmaKindFlags (src/compiler/types.ts:10207:1) | 0 | unsure |
| 28723 | typeof PredicateSemantics | string | PredicateSemantics (src/compiler/types.ts:932:1) | 0 | unsure |
| 36083 | typeof ProcessLevel | string | ProcessLevel (src/compiler/transformers/taggedTemplate.ts:29:1) | 0 | unsure |
| 42501 | typeof ProgramUpdateLevel | string | ProgramUpdateLevel (src/compiler/watchUtilities.ts:381:1) | 0 | unsure |
| 25833 | typeof PropertyLikeParse | string | PropertyLikeParse (src/compiler/parser.ts:8879:9) | 0 | unsure |
| 32408 | typeof RecursionFlags | string | RecursionFlags (src/compiler/checker.ts:1373:1) | 0 | unsure |
| 33246 | typeof ReferenceHint | string | ReferenceHint (src/compiler/checker.ts:1160:1) | 0 | unsure |
| 15323 | typeof RegularExpressionFlags | string | RegularExpressionFlags (src/compiler/types.ts:2794:1) | 0 | unsure |
| 1623 | typeof RelationComparisonResult | string | RelationComparisonResult (src/compiler/types.ts:917:1) | 0 | unsure |
| 28433 | typeof RelativePreference | string | RelativePreference (src/compiler/moduleSpecifiers.ts:164:1) | 0 | unsure |
| 1624 | typeof ScriptKind | string | ScriptKind (src/compiler/types.ts:7663:1) | 0 | unsure |
| 6706 | typeof ScriptTarget | string | ScriptTarget (src/compiler/types.ts:7682:1) | 0 | unsure |
| 40751 | typeof SeenPackageName | string | SeenPackageName (src/compiler/program.ts:2385:9) | 0 | unsure |
| 1625 | typeof SignatureCheckMode | string | SignatureCheckMode (src/compiler/checker.ts:1357:1) | 0 | unsure |
| 1626 | typeof SignatureFlags | string | SignatureFlags (src/compiler/types.ts:7006:1) | 0 | unsure |
| 25354 | typeof SignatureFlags | string | SignatureFlags (src/compiler/parser.ts:405:1) | 0 | unsure |
| 41257 | typeof SignatureInfo | string | SignatureInfo (src/compiler/builderState.ts:56:1) | 0 | unsure |
| 28724 | typeof SignatureKind | string | SignatureKind (src/compiler/types.ts:6999:1) | 0 | unsure |
| 1627 | typeof SnippetKind | string | SnippetKind (src/compiler/types.ts:8364:1) | 0 | unsure |
| 25036 | typeof SpeculationKind | string | SpeculationKind (src/compiler/parser.ts:414:1) | 0 | unsure |
| 37986 | typeof SpreadSegmentKind | string | SpreadSegmentKind (src/compiler/transformers/es2015.ts:472:1) | 0 | unsure |
| 21379 | typeof States | string | States (src/compiler/utilities.ts:10977:5) | 0 | unsure |
| 43798 | typeof StatisticType | string | StatisticType (src/compiler/executeCommandLine.ts:99:1) | 0 | unsure |
| 40163 | typeof StructureIsReused | string | StructureIsReused (src/compiler/types.ts:4967:1) | 0 | unsure |
| 28725 | typeof SymbolAccessibility | string | SymbolAccessibility (src/compiler/types.ts:5679:1) | 0 | unsure |
| 1628 | typeof SymbolFlags | string | SymbolFlags (src/compiler/types.ts:5947:1) | 0 | unsure |
| 28726 | typeof SymbolFormatFlags | string | SymbolFormatFlags (src/compiler/types.ts:5626:1) | 0 | unsure |
| 1630 | typeof SyntaxKind | string | SyntaxKind (src/compiler/types.ts:40:1) | 0 | unsure |
| 39033 | typeof SyntaxKindFeatureFlags | string | SyntaxKindFeatureFlags (src/compiler/transformer.ts:111:1) | 0 | unsure |
| 39260 | typeof TempFlags | string | TempFlags (src/compiler/emitter.ts:6345:1) | 0 | unsure |
| 28728 | typeof Ternary | string | Ternary (src/compiler/types.ts:7156:1) | 0 | unsure |
| 15324 | typeof TokenFlags | string | TokenFlags (src/compiler/types.ts:2815:1) | 0 | unsure |
| 38980 | typeof TransformationState | string | TransformationState (src/compiler/transformer.ts:104:1) | 0 | unsure |
| 1631 | typeof TransformFlags | string | TransformFlags (src/compiler/types.ts:8223:1) | 0 | unsure |
| 25477 | typeof Tristate | string | Tristate (src/compiler/parser.ts:8784:5) | 0 | unsure |
| 1632 | typeof TypeFacts | string | TypeFacts (src/compiler/checker.ts:1233:1) | 0 | unsure |
| 1633 | typeof TypeFlags | string | TypeFlags (src/compiler/types.ts:6335:1) | 0 | unsure |
| 28729 | typeof TypeFormatFlags | string | TypeFormatFlags (src/compiler/types.ts:5584:1) | 0 | unsure |
| 1634 | typeof TypeMapKind | string | TypeMapKind (src/compiler/types.ts:7087:1) | 0 | unsure |
| 18717 | typeof TypePredicateKind | string | TypePredicateKind (src/compiler/types.ts:5686:1) | 0 | unsure |
| 28730 | typeof TypeReferenceSerializationKind | string | TypeReferenceSerializationKind (src/compiler/types.ts:5843:1) | 0 | unsure |
| 36126 | typeof TypeScriptSubstitutionFlags | string | TypeScriptSubstitutionFlags (src/compiler/transformers/ts.ts:211:1) | 0 | unsure |
| 30872 | typeof TypeSystemPropertyName | string | TypeSystemPropertyName (src/compiler/checker.ts:1329:1) | 0 | unsure |
| 28731 | typeof UnionReduction | string | UnionReduction (src/compiler/types.ts:5505:1) | 0 | unsure |
| 34467 | typeof UnusedKind | string | UnusedKind (src/compiler/checker.ts:1405:1) | 0 | unsure |
| 42824 | typeof UpToDateStatusType | string | UpToDateStatusType (src/compiler/tsbuild.ts:10:1) | 0 | unsure |
| 37215 | typeof UsingKind | string | UsingKind (src/compiler/transformers/esnext.ts:64:1) | 0 | unsure |
| 1636 | typeof VarianceFlags | string | VarianceFlags (src/compiler/types.ts:6681:1) | 0 | unsure |
| 26094 | typeof WatchDirectoryFlags | string | WatchDirectoryFlags (src/compiler/types.ts:7722:1) | 0 | unsure |
| 8172 | typeof WatchDirectoryKind | string | WatchDirectoryKind (src/compiler/types.ts:7398:1) | 0 | unsure |
| 8173 | typeof WatchFileKind | string | WatchFileKind (src/compiler/types.ts:7389:1) | 0 | unsure |
| 40097 | typeof WatchLogLevel | string | WatchLogLevel (src/compiler/watchUtilities.ts:677:1) | 0 | unsure |
| 32719 | typeof WideningKind | string | WideningKind (src/compiler/checker.ts:1224:1) | 0 | unsure |
| 21268 | Uint16Array<ArrayBuffer> | number | Uint16Array (external/node_modules/typescript/lib/lib.es5.d.ts:3008:1); Uint16Array (external/node_modules/typescript/lib/lib.es5.d.ts:3284:13); Uint16Array (external/node_modules/typescript/lib/lib.es2015.core.d.ts:577:1); Uint16Array (external/node_modules/typescript/lib/lib.es2015.iterable.d.ts:420:1); Uint16Array (external/node_modules/typescript/lib/lib.es2015.symbol.wellknown.d.ts:293:1); Uint16Array (external/node_modules/typescript/lib/lib.es2016.array.include.d.ts:71:1) | 0 | unsure |
| 6029 | WatchOptions | CompilerOptionsValue | WatchOptions (src/compiler/types.ts:7585:1) | 12 | unsure |

The broad ownership-only control inventory includes unsupported indexed/mapped types whose missing storage proofs prevent asserting acyclicity. Those rows remain unsure. It also includes records that are acyclic by reproduced value/field edges but become unsure when structural views or unavailable capture information are considered. A zero in the strict acyclic column does not mean every record literal should stop lowering.

## What is mirrored and what is unsure

1. `graphLinks` follows union/intersection members, non-method data fields (including readonly fields), arrays/tuples, and the type arguments of library maps, sets and Promises. Map keys as well as values are retained, as Go does. Signature parameter/return types are not owned edges. Records add the requested value edge. Class instances, arrays and tuples are not mistaken for records merely because they have indices.
2. `recordElement`'s ordinary case is exactly one mutable string index, no call/construct signature, and no regexp-library index. Other index signatures and finite mapped records are inventoried as unsure. Go's `concrete` generic substitutions, utility-alias proofs for Partial/Pick, enumeration record storage, phantom array layouts and host storage substitutions do not all have faithful stock-checker allocation identities. The script flags the recognized affected cases. Generic aliases with unbound any/type parameters are not used as allocation candidates.
3. Go adds symmetric `related` edges between compatible structural views. The script uses the same bidirectional assignability predicate on its source candidates and stores a spanning forest; this preserves reachability/SCC membership. Its universe also inventories source type declarations, whereas Go's universe comes from lowered uses, shapes, literal interning and instantiations. Cycles that need this approximation remain unsure. A compatibility cycle through `{}` can put a primitive-valued dictionary in a source-view component without any value back-edge.
4. Reflexive `related(t,t)` edges are excluded from the reported cycle test. This is an explicit interpretation of “self-edge”: a genuine record-value self-edge counts, an identity-view tautology does not. Counting every reflexive view literally would classify every eligible record candidate as cyclic. Go selects SCCs only if a slot/capture seed is present, then adds the acyclic caches at line 262; a source-only SCC-size census is not the same as that seeded selection. The requested hypothetical tests SCC size/self-edge and does not add acyclic record caches.
5. Closure records, captured cells, `localAlso`, shared environments, forwarded frames, accessor captures, static constructor identities, library iterator captures, freshness/slot seeds and instantiations produced only during lowering are not recreated. Reachable function/accessor/iterator/array-extension cases are marked unsure. Missing lowering-only candidates could add cycles beyond this inventory; no closed-world completeness claim is made for them.
6. The Promise refusal at [internal/lower/graph_types.go:234](https://github.com/system-inc/adamic/blob/fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d/internal/lower/graph_types.go#L234) applies after a component is selected because a Promise cannot be a graph member. Promise payload edges remain in this census; the census does not remove a cycle merely because the current compiler would refuse it. This is the model for the proposed record stop.

## Literal attribution

[internal/lower/records.go:314](https://github.com/system-inc/adamic/blob/fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d/internal/lower/records.go#L314) contains the single `ir.RecordLiteral` constructor in that file, at line 319. It accepts object literals, including `{}`, data/shorthand entries and computed property names. Its first-spread case is included in the inventory. There is no additional source expression form constructing `ir.RecordLiteral` in `records.go`; `new Map` is not a record literal.

The script first uses the checker's non-null contextual record destination, including record members of contextual unions, and otherwise its inferred record type. Assertions are also inspected. It counts source sites, including sites inside functions regardless of call frequency, rather than execution/allocation totals. It does not count later property writes. Go's explicitly carried slots and representation checks can change admission, so union/storage ambiguity and unsupported members are retained as evidence, not asserted successful lowerings. Prototype-setting entries, unsupported computed-key representations, non-first spreads and representation refusals can still reject individual candidates; this is a static site census.

## Reproduction and validation

Run from the Adamic snapshot above. Setup/install/build output must go to log files, and print each child's exit status. The source environment variable must name a checkout whose HEAD is the exact TypeScript pin.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/record-cycles-setup.log 2>&1
rc=$?; echo "setup exit=$rc"; test "$rc" -eq 0 || exit "$rc"
npm ci --prefix stage3/api > /tmp/record-cycles-npm.log 2>&1
rc=$?; echo "npm ci exit=$rc"; test "$rc" -eq 0 || exit "$rc"
export ADAMIC_TYPESCRIPT_SOURCE=/path/to/pinned-typescript
bash stage3/apply.sh /tmp/record-cycles-adapted > /tmp/record-cycles-apply.log 2>&1
rc=$?; echo "apply exit=$rc"; test "$rc" -eq 0 || exit "$rc"
ln -s "$PWD/stage3/api/node_modules" /tmp/record-cycles-adapted/node_modules
node --stack-size=16384 --max-old-space-size=8192 stage3/census/record-cycles/census.cjs /tmp/record-cycles-adapted stage3/census/record-cycles/counts.json > /tmp/record-cycles-census.log 2>&1
rc=$?; echo "census exit=$rc"; test "$rc" -eq 0 || exit "$rc"
node stage3/census/record-cycles/validate.cjs > /tmp/record-cycles-validation.log 2>&1
rc=$?; echo "validation exit=$rc"; test "$rc" -eq 0 || exit "$rc"
node stage3/census/record-cycles/render.cjs
```

Setup, npm ci, apply, census, report rendering and fixture validation all returned exit code 0. Initial diagnostic/type-resolution and mixed-union attribution issues were corrected before these final artifacts were generated.

The fixture validates config-relative Node declarations, exclusion of ordinary Partial/Pick objects, finite Record aliases, mixed-union attribution uncertainty, a record-holder-record cycle, a genuine record self-edge, Map value and key paths, contextual empty and computed literals, a primitive-valued ownership control, and the unsure classification for function values. It uses a diagnostic-free temporary program and removes it after the child exits.

Checker diagnostics for the census program: 81. Node ambient declarations resolve correctly. One retained non-project diagnostic is TS2488 at `src/compiler/transformers/classFields.ts:1503`, where this adapted tree's mutable-local flow leaves a spread operand typed never. That source is unchanged by this report; it is not on either reported record-value cycle. The full messages and locations are in `counts.json`; project/composite root-list diagnostics are not suppressed by changing compiler options.
