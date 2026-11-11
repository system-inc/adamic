### New public SDK refusal: the requested no-new-refusal outcome is not met

**Neither exact form of (c) meets all requested gates.** Keeping LineMapSource internal breaks declaration bundling. Making it public repairs the missing type but introduces a new Adamic refusal for an ordinary public SourceFile caller. This is outside the 34 internal compiler views, so it must not be hidden by their clearance count.

The emitted-public-API probe is:

```ts
import * as ts from "./typescript.js";
export function probe(file: ts.SourceFile): ts.LineAndCharacter {
    return ts.getLineAndCharacterOfPosition(file, 0);
}
```

| Emitted SDK | Production checker errors | Cohere findings | Guarded optional-widening refusals |
| --- | ---: | ---: | ---: |
| original | 0 | 0 | 0 |
| c-public | 0 | 1 | 1 |
| c-public-unstripped | 0 | 1 | 1 |
| c-public-readonly-sdk-control | 0 | 1 | 1 |

Both cohere and the same guarded full-mode lowering binary report the new refusal, including when only the exported slot is made readonly:

```text
optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields
```

Public SourceFile has no lineMap because its original member remains @internal. Newly public LineMapSource declares an optional lineMap, so this public view reintroduces the structural hiding rule. Its readonly variant is still refused because optional widening is independent of writable-slot invariance. The production checker accepts all four SDK probes; this is an Adamic relation refusal, not a TypeScript checker error.

The readonly SDK control changes only the emitted interface member for this probe. It is not a full implementation or measurement of (d), and no narrow writer is claimed. The conditional compiler-side writable-slot refusal that requested (d) was not observed. Even a separate narrow writer would not change this SDK argument relation while its first parameter retains that public optional member. A further public API/cache-annotation change would be required; it is not silently folded into the requested experiment.

Every SDK finding, including the baseline external function-without-a-body NotYet, is retained below. The SDK census is isolated to the consumer and emitted declaration bundle, uses LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1, and has no backend output. It is additional evidence, not a substitute for the real ten-file runs.

**original**

````json
{
  "production_checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.251802112
  },
  "cohere_findings": [],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-original/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-original/typescript.d.ts:8557:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-original/typescript.d.ts:8557:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-original/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-original/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-original/typescript.d.ts:8557:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ]
}
````


**c-public**

````json
{
  "production_checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.171637924
  },
  "cohere_findings": [
    {
      "kind": "finding",
      "path": "/workspace/adamic/stage3/census/sourcefile-like-sites/continuation/c-public/public-adamic-consumer.ts",
      "line": 3,
      "column": 45,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'LineMapSource', which adds the optional property `lineMap` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `lineMap` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a 'readonly number[] | undefined' and holds something else. Declare `lineMap` in the source's type, or build the value with the property named."
    }
  ],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:2:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:3:45",
      "reason": "optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:3:45: Adamic 0.1 refuses optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields; declare lineMap on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-c-public/typescript.d.ts:8561:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-c-public/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-c-public/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ]
}
````


**c-public-unstripped**

````json
{
  "production_checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.181255141
  },
  "cohere_findings": [
    {
      "kind": "finding",
      "path": "/workspace/adamic/stage3/census/sourcefile-like-sites/continuation/c-public-unstripped/public-adamic-consumer.ts",
      "line": 3,
      "column": 45,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'LineMapSource', which adds the optional property `lineMap` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `lineMap` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a 'readonly number[] | undefined' and holds something else. Declare `lineMap` in the source's type, or build the value with the property named."
    }
  ],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:2:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:3:45",
      "reason": "optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:3:45: Adamic 0.1 refuses optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields; declare lineMap on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-c-public-unstripped/typescript.d.ts:8561:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-c-public-unstripped/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-c-public-unstripped/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ]
}
````


**c-public-readonly-sdk-control**

````json
{
  "production_checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.161772996
  },
  "cohere_findings": [
    {
      "kind": "finding",
      "path": "/workspace/adamic/stage3/census/sourcefile-like-sites/continuation/c-public-readonly-sdk-control/public-adamic-consumer.ts",
      "line": 3,
      "column": 45,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'LineMapSource', which adds the optional property `lineMap` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `lineMap` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a 'readonly number[] | undefined' and holds something else. Declare `lineMap` in the source's type, or build the value with the property named."
    }
  ],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:2:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:3:45",
      "reason": "optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:3:45: Adamic 0.1 refuses optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields; declare lineMap on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/typescript.d.ts:8561:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ]
}
````

