## Fourth measurement: internal LineMapSource overload, variant (e)

Measured from the same frozen adapted input on main d690f276749f91e6b36fe3a9726a5a453fcd7d39 and TypeScript 6.0.3 source 050880ce59e30b356b686bd3144efe24f875ebc8. The guarded binary and pinned cohere are unchanged. The preceding report commit is e87b5ec71092a42bf931b8f0f0a6b8840d3718b6. This is report only, with no compiler or rule changes landed.

### Exact source and lint

A fresh scratch copy inserts the requested @internal LineMapSource immediately before the whole documented SourceFileLike declaration. SourceFileLike, including its existing leading documentation, text member, lineMap member and their @internal annotations, remains byte-for-byte unchanged. It does not extend LineMapSource. The only other source file changed is scanner.ts: one required type-only import, the new overload set, and the first parameter type of the already-internal getLineStarts. Both function bodies are unchanged.

The complete overload set and unchanged body in the final scanner.ts are:

```ts
/** @internal */
export function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;
export function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter; // eslint-disable-line @typescript-eslint/unified-signatures
export function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter {
    return computeLineAndCharacterOfPosition(getLineStarts(sourceFile), position);
}
```

The internal overload is number 1 at scanner.ts:549:1. The public overload is number 2 at 550:1. The implementation is 551:1. getLineStarts remains at 503:1. The six new lines in types.ts shift its old declaration locations by six; the three new scanner lines shift old locations after the implementation header by three. Comparisons map those unchanged locations back to baseline, while raw streams retain actual coordinates.

The pinned full ESLint configuration requires the trailing unified-signatures suppression on the public overload. Its initial diagnostic was:

````json
[
  {
    "ruleId": "@typescript-eslint/unified-signatures",
    "severity": 2,
    "message": "These overloads can be combined into one signature taking `LineMapSource | SourceFileLike`.",
    "line": 550,
    "column": 47,
    "messageId": "singleParameterDifference",
    "endLine": 550,
    "endColumn": 73
  }
]
````


The exact // eslint-disable-line @typescript-eslint/unified-signatures comment is the one used by getPositionOfLineAndCharacter. No lint rule is globally disabled and no other diagnostic is repaired. The final scanner lint has 10 errors, matching the frozen baseline count of 10. These are existing no-unnecessary-type-assertion diagnostics; there is no remaining overload-order or unified-signatures diagnostic.

Two preliminary attempts are retained as clearly labeled controls, not counted as final (e) measurements. e-unsuppressed-control predates the required lint comment. e-displaced-doc-control exposed an insertion error: the new interface had captured SourceFileLike's existing documentation, causing a comment-only API diff. The final tree preserves that whole documented declaration. Native children in abandoned attempts were terminated and reaped by their own bounded wrappers before source changes. A transient HTTP 503 interrupted one checkpoint push; ordinary non-force push retries were added. All final results below come from the corrected, fixed-byte tree.

### All 34 original views

| Complete file streams | Cleared at original view | Unchanged finding | Replacement finding | Checker diagnostic | Incomplete or excluded |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 10/10 | 22 | 12 | 0 | 0 | 0 |

Cleared means that the original argument/call span has no guarded finding or checker diagnostic and its enclosing unit was attempted. It does not mean that the entire root lowers. The 22 helper views contain 21 SourceFile values and one SourceMapSource value. The 12 unchanged views still target the unmodified SourceFileLike callees. The earlier checker.ts 25-minute cutoff remains a limitation of the original site census; this comparison covers all 34 original observed views.

| Callee | Original sites | Cleared | Unchanged | Replacement / checker / unknown |
| --- | ---: | ---: | ---: | ---: |
| `getEndLinePosition` | 1 | 0 | 1 | 0 |
| `getLineAndCharacterOfPosition` | 18 | 18 | 0 | 0 |
| `getLineStarts` | 4 | 4 | 0 | 0 |
| `getLinesBetweenPositions` | 5 | 0 | 5 | 0 |
| `getPositionOfLineAndCharacter` | 2 | 0 | 2 | 0 |
| `transferSourceFileChildren` | 2 | 0 | 2 | 0 |
| `unsetNodeChildren` | 2 | 0 | 2 | 0 |

### Which overload the 18 original internal calls resolve to

This is measured twice. The stock TypeScript 6.0.3 checker resolves calls in the same whole-src 604-root program used for the previous caller audit. Independently, e-overloads.go.txt calls production load.Load on all 79 compiler roots and production Checker.GetResolvedSignature for each call. It does not infer a choice from overload order, and it does not lower or emit anything. The native checker sees 20 compiler calls, including the two checker.ts calls outside the original 34-view list. All returned choices are retained in e/native-overload-resolution.json. The requested 18-call subset follows.

| Original argument location | Enclosing function | Argument | Declared type | Native flow type | Stock overload | Native overload | Native chosen declaration |
| --- | --- | --- | --- | --- | ---: | ---: | --- |
| emitter.ts:3961:112 | `emitJsxExpression` | `currentSourceFile` | `ts.SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| emitter.ts:3961:180 | `emitJsxExpression` | `currentSourceFile` | `ts.SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| emitter.ts:6233:96 | `emitPos` | `sourceMapSource` | `ts.SourceMapSource` | `SourceMapSource` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:671:67 | `formatDiagnostic` | `diagnostic.file` | `SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:713:89 | `formatCodeSpan` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:714:87 | `formatCodeSpan` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:715:58 | `formatCodeSpan` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:768:89 | `formatLocation` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| tracing.ts:206:67 | `getLocation` | `file` | `SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| tracing.ts:207:65 | `getLocation` | `file` | `SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| transformers/declarations.ts:1264:139 | `visitDeclarationSubtree` | `currentSourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| transformers/declarations.ts:1264:70 | `visitDeclarationSubtree` | `currentSourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| transformers/jsx.ts:359:63 | `visitJsxOpeningLikeElementOrFragmentJSX` | `originalFile` | `Node` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| utilities.ts:1025:47 | `nodePosToString` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| utilities.ts:1213:46 | `<anonymous@1212>` | `sourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| utilities.ts:2526:67 | `getErrorSpanForArrowFunction` | `sourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| utilities.ts:2527:65 | `getErrorSpanForArrowFunction` | `sourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| watch.ts:249:60 | `<anonymous@241>` | `diagnosticForFileName.file!` | `SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |

Both checkers resolve all 18 to overload 1, whose first parameter is LineMapSource. The public SourceFileLike overload is not selected at these sites. e/calls.json also retains all 93 source calls, both available signatures, chosen declaration text, the internal marker, and argument declared/flow types. All 79 getLineAndCharacterOfPosition calls in that full source program resolve to the internal first overload. The separate SDK probe instead reads only the emitted public declaration and sees the original SourceFileLike signature.

### Scanner own stream and complete overload handling

The following is the complete set of unit records and findings for getLineAndCharacterOfPosition in the final full-mode scanner stream, including both bodyless overload declarations and the implementation. No finding is summarized away:

````json
{
  "units": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:549:1",
      "kind": "KindFunctionDeclaration",
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineAndCharacterOfPosition",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:549:1"
    },
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:550:1",
      "kind": "KindFunctionDeclaration",
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineAndCharacterOfPosition",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:550:1"
    },
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:551:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 51383,
      "body_end": 51472,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineAndCharacterOfPosition",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:551:1"
    }
  ],
  "findings": []
}
````


Every new overload-related finding anywhere across the ten selected files is:

````json
[]
````


In particular, new getLineAndCharacterOfPosition "overload N ... cannot be served by implementation" findings are:

````json
[]
````


The signature path checks censusOverloads before the body (functions.go:68), and module registration checks it too (modules.go:195). censusOverload compares each promised argument to its implementation parameter and the implementation result to each promised result. The full corresponding source is retained as e-census-small.go.txt, e-functions.go.txt and e-modules.go.txt. These are the unchanged pinned rules, not a new rule or exception. All three getLineAndCharacterOfPosition units are attempted, with no signature failure or checker rejection.

getLineStarts own complete unit records and findings are:

````json
{
  "units": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:503:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 49075,
      "body_end": 49173,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineStarts",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:503:1"
    }
  ],
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:504:35",
      "reason": "an assignment value to a member",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:504:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet",
      "start": 49077,
      "end": 49170
    }
  ]
}
````


Its assignment-value NotYet at scanner.ts:504:35 is the same pre-existing baseline stop. The new overloads do not repair that expression, and it is not the core.ts first stop. The complete scanner stream is retained under e/raw/scanner.jsonl or scanner.jsonl.gz; all scanner overload findings, including unrelated existing ones, are also retained in e/scanner-overload-handling.json.

### All cohere and lowering additions/removals

The same pinned cohere binary runs the complete cohere:adamic set under the actual Adamic checker settings. Its successful run checks 77 compiler files with 40 rules, preserving all findings and unresolved-input warnings. The baseline comparison includes every finding in the same compiler scope, not just the two helpers. The guarded lowering comparison includes every finding, phase, reason and kind across all ten selected-file streams.

| Measurement | Added | Removed |
| --- | ---: | ---: |
| Cohere findings | 0 | 24 |
| Guarded full-census findings | 0 | 22 |

Per-phase guarded differences:

````json
{
  "deferred": {
    "added": 0,
    "removed": 0
  },
  "lowering": {
    "added": 0,
    "removed": 0
  },
  "refusal_scan": {
    "added": 0,
    "removed": 22
  }
}
````


Every added cohere finding:

````json
[]
````


Every removed cohere finding:

````json
[
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/emitter.ts",
    "line": 3961,
    "column": 112,
    "severity": "error",
    "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/transformers/declarations.ts",
    "line": 1264,
    "column": 139,
    "severity": "error",
    "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/emitter.ts",
    "line": 3961,
    "column": 180,
    "severity": "error",
    "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 6748,
    "column": 38,
    "severity": "error",
    "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/executeCommandLine.ts",
    "line": 109,
    "column": 41,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 2973,
    "column": 42,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 1213,
    "column": 46,
    "severity": "error",
    "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 1025,
    "column": 47,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/checker.ts",
    "line": 52425,
    "column": 55,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/checker.ts",
    "line": 52424,
    "column": 57,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 715,
    "column": 58,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/watch.ts",
    "line": 249,
    "column": 60,
    "severity": "error",
    "message": "`diagnosticForFileName.file!` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/transformers/jsx.ts",
    "line": 359,
    "column": 63,
    "severity": "error",
    "message": "`originalFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/tracing.ts",
    "line": 207,
    "column": 65,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 2527,
    "column": 65,
    "severity": "error",
    "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/emitter.ts",
    "line": 1444,
    "column": 66,
    "severity": "error",
    "message": "`Debug.checkDefined(currentSourceFile)` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/tracing.ts",
    "line": 206,
    "column": 67,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 2526,
    "column": 67,
    "severity": "error",
    "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 671,
    "column": 67,
    "severity": "error",
    "message": "`diagnostic.file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/transformers/declarations.ts",
    "line": 1264,
    "column": 70,
    "severity": "error",
    "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 714,
    "column": 87,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 713,
    "column": 89,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 768,
    "column": 89,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/emitter.ts",
    "line": 6233,
    "column": 96,
    "severity": "error",
    "message": "`sourceMapSource` is a 'SourceMapSource', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceMapSource' does not declare. A 'SourceMapSource' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  }
]
````


Every added guarded finding:

````json
[]
````


Every removed guarded finding:

````json
[
  {
    "unit": "<tree>/src/compiler/emitter.ts:1443:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/emitter.ts:1444:66",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/emitter.ts:1444:66: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "emitter.ts"
  },
  {
    "unit": "<tree>/src/compiler/emitter.ts:3959:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/emitter.ts:3961:112",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/emitter.ts:3961:112: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "emitter.ts"
  },
  {
    "unit": "<tree>/src/compiler/emitter.ts:3959:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/emitter.ts:3961:180",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/emitter.ts:3961:180: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "emitter.ts"
  },
  {
    "unit": "<tree>/src/compiler/executeCommandLine.ts:105:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/executeCommandLine.ts:109:41",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/executeCommandLine.ts:109:41: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "executeCommandLine.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:2966:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:2973:42",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:2973:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:667:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:671:67",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:671:67: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:712:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:713:89",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:713:89: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:712:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:714:87",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:714:87: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:712:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:715:58",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:715:58: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:767:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:768:89",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:768:89: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/tracing.ts:200:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/tracing.ts:206:67",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/tracing.ts:206:67: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "tracing.ts"
  },
  {
    "unit": "<tree>/src/compiler/tracing.ts:200:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/tracing.ts:207:65",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/tracing.ts:207:65: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "tracing.ts"
  },
  {
    "unit": "<tree>/src/compiler/transformers/declarations.ts:1006:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/transformers/declarations.ts:1264:139",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/transformers/declarations.ts:1264:139: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "transformers/declarations.ts"
  },
  {
    "unit": "<tree>/src/compiler/transformers/declarations.ts:1006:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/transformers/declarations.ts:1264:70",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/transformers/declarations.ts:1264:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "transformers/declarations.ts"
  },
  {
    "unit": "<tree>/src/compiler/transformers/jsx.ts:333:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/transformers/jsx.ts:359:63",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/transformers/jsx.ts:359:63: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "transformers/jsx.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:1023:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:1025:47",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:1025:47: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:1210:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:1213:46",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:1213:46: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:2523:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:2526:67",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:2526:67: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:2523:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:2527:65",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:2527:65: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:6747:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:6748:38",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:6748:38: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/watch.ts:233:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/watch.ts:249:60",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/watch.ts:249:60: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "watch.ts"
  },
  {
    "unit": "<tree>/src/compiler/emitter.ts:6228:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/emitter.ts:6233:96",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceMapSource, which can hide fields",
    "text": "<tree>/src/compiler/emitter.ts:6233:96: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceMapSource, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "emitter.ts"
  }
]
````


Cohere removes the 22 helper-site optional-method findings plus the two checker.ts helper-call findings outside the original 34-site lowering list. The unchanged SourceFileLike contracts still account for the other 12 original views. Scratch roots and unchanged source-line shifts are normalized solely for comparison; raw streams preserve their original paths and locations.

### Whole src, services, and core.ts

| Input | Whole-src stock errors | Stock exact optional errors | Production Adamic whole-src errors | Services stock errors | Services exact optional errors | Services TS2386, stock / exact |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Frozen input | 0 | 416 | 6851 | 0 | 142 | 0 / 0 |
| (e) | 0 | 416 | 6851 | 0 | 142 | 0 / 0 |

All whole-src runs use the same 604 roots from 13 solution projects. Stock defaults are preserved; the extra stock run enables exactOptionalPropertyTypes. The production run uses unmodified load.Load and every Adamic option, including exactOptionalPropertyTypes and noUncheckedIndexedAccess. Its baseline errors, including type-only imports and unavailable test globals, remain real errors; they are not attributed to (e). The production and exact-optional diagnostic additions/removals are:

````json
{
  "production": {
    "baseline_unique": 6840,
    "variant_unique": 6840,
    "added": [],
    "removed": []
  },
  "stock_exact_optional": {
    "baseline_unique": 416,
    "variant_unique": 416,
    "added": [],
    "removed": []
  }
}
````


The actual core.ts first-stop record is:

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-e/src/compiler/core.ts"
  ],
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:232:34",
  "reason": "a cast the runtime can't check",
  "fix": "use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
  "seconds": 16.533875995
}
````


The compiler-only core root has no checker diagnostics. Its first stop is the unrelated tracing.ts:232:34 cast, not the cache write and not an overload-serving refusal. The driver repeats the same one-root input as a whole-program record; the first record is quoted.

### stripInternal-on declarations, bundle, and lane API guard

Stock TypeScript performs actual emitDeclarationOnly with stripInternal=true, then the unchanged upstream declaration bundler runs on the emitted module entry. The emit options, emitted-file hashes, bundler log and public bundle are retained. The internal LineMapSource and internal first overload are stripped, while the original public SourceFileLike overload remains. getLineStarts stays internal and absent.

| Public result | Measured value |
| --- | --- |
| Emitted declaration files | 278 |
| Emit diagnostics | 0 |
| Normalized public API changes | 0 |
| Public closure/consumer diagnostics | 0 |
| Lane API guard errors / unsanctioned changes | 0 |

The complete selected emitted public declarations are:

````json
[
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike {\n        readonly text: string;\n    }"
  },
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike {\n        getLineAndCharacterOfPosition(pos: number): LineAndCharacter;\n    }"
  },
  {
    "name": "getLineAndCharacterOfPosition",
    "kind": "FunctionDeclaration",
    "text": "function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter;"
  }
]
````


The textual typescript.d.ts diff is retained in e/typescript-public.diff, including the diff child exit code. Its public bytes and normalized declaration map are compared against the frozen input, not a hand-edited expected snapshot. The actual stage3/lane check.api_check result is:

````json
{
  "errors": [],
  "facts": {
    "reference_declarations": 0,
    "composed_declarations": 267,
    "sanctioned_declarations": 267,
    "added_lines": 349,
    "removed_lines": 267
  }
}
````


No API sanction is added or changed. SourceFileLike retains its original documentation and public members. The internal implementation type is not a dangling name in this public d.ts.

### SDK probe

The same emitted-public-API consumer used for (c) is rechecked with the production checker, cohere adamic rules and the same guarded full-mode lowering binary. It imports the real emitted public bundle and accepts a ts.SourceFile parameter. All SDK findings, including the baseline external function-without-a-body NotYet, are quoted in full:

````json
{
  "checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.300982083
  },
  "cohere_findings": [],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-e/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-e/typescript.d.ts:8557:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-e/typescript.d.ts:8557:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-e/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-e/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-e/typescript.d.ts:8557:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ],
  "guarded_delta": {
    "baseline_unique": 2,
    "variant_unique": 2,
    "added": [],
    "removed": []
  }
}
````


A zero new-refusal count is relative to the original SDK probe. It is not a claim that an implementation-free declaration bundle can be compiled into native runtime code. The public signature stays SourceFileLike, so the LineMapSource optional-cache hiding refusal introduced by public (c) does not recur.

### Every original view, with literal findings

| Original location | Phase | Enclosing function | Callee | Classification |
| --- | --- | --- | --- | --- |
| emitter.ts:1444:66 | refusal_scan | `getCurrentLineMap` | `getLineStarts` | cleared at original view |
| emitter.ts:3961:112 | refusal_scan | `emitJsxExpression` | `getLineAndCharacterOfPosition` | cleared at original view |
| emitter.ts:3961:180 | refusal_scan | `emitJsxExpression` | `getLineAndCharacterOfPosition` | cleared at original view |
| emitter.ts:6233:96 | refusal_scan | `emitPos` | `getLineAndCharacterOfPosition` | cleared at original view |
| executeCommandLine.ts:109:41 | refusal_scan | `<anonymous@107>` | `getLineStarts` | cleared at original view |
| parser.ts:10032:36 | refusal_scan | `updateSourceFile` | `transferSourceFileChildren` | unchanged finding |
| parser.ts:10032:48 | refusal_scan | `updateSourceFile` | `transferSourceFileChildren` | unchanged finding |
| parser.ts:10104:37 | refusal_scan | `visitNode` | `unsetNodeChildren` | unchanged finding |
| parser.ts:10261:42 | refusal_scan | `visitNode` | `unsetNodeChildren` | unchanged finding |
| program.ts:2973:42 | refusal_scan | `markPrecedingCommentDirectiveLine` | `getLineStarts` | cleared at original view |
| program.ts:671:67 | refusal_scan | `formatDiagnostic` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:713:89 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:714:87 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:715:58 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:733:57 | refusal_scan | `formatCodeSpan` | `getPositionOfLineAndCharacter` | unchanged finding |
| program.ts:734:76 | refusal_scan | `formatCodeSpan` | `getPositionOfLineAndCharacter` | unchanged finding |
| program.ts:768:89 | refusal_scan | `formatLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| tracing.ts:206:67 | refusal_scan | `getLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| tracing.ts:207:65 | refusal_scan | `getLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/declarations.ts:1264:139 | refusal_scan | `visitDeclarationSubtree` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/declarations.ts:1264:70 | refusal_scan | `visitDeclarationSubtree` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/jsx.ts:359:63 | refusal_scan | `visitJsxOpeningLikeElementOrFragmentJSX` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:1025:47 | refusal_scan | `nodePosToString` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:1213:46 | refusal_scan | `<anonymous@1212>` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2526:67 | refusal_scan | `getErrorSpanForArrowFunction` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2527:65 | refusal_scan | `getErrorSpanForArrowFunction` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2531:70 | refusal_scan | `getErrorSpanForArrowFunction` | `getEndLinePosition` | unchanged finding |
| utilities.ts:6748:38 | refusal_scan | `getLineOfLocalPosition` | `getLineStarts` | cleared at original view |
| utilities.ts:7944:37 | refusal_scan | `getLinesBetweenRangeEndAndRangeStart` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7949:37 | refusal_scan | `getLinesBetweenRangeEndPositions` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7959:37 | refusal_scan | `positionsAreOnSameLine` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7971:37 | refusal_scan | `getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7977:37 | refusal_scan | `getLinesBetweenPositionAndNextNonWhitespaceCharacter` | `getLinesBetweenPositions` | unchanged finding |
| watch.ts:249:60 | refusal_scan | `<anonymous@241>` | `getLineAndCharacterOfPosition` | cleared at original view |

**emitter.ts:1444:66** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1443:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 58788,
      "body_end": 58904,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getCurrentLineMap",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ]
}
````


**emitter.ts:3961:112** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ]
}
````


**emitter.ts:3961:180** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ]
}
````


**emitter.ts:6233:96** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:6228:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 268634,
      "body_end": 269150,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitPos",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ]
}
````


**executeCommandLine.ts:109:41** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/executeCommandLine.ts:105:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 2705,
      "body_end": 2984,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "countLines",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/executeCommandLine.ts:105:1"
    }
  ]
}
````


**parser.ts:10032:36** `transferSourceFileChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10032:36",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1"
    }
  ]
}
````


**parser.ts:10032:48** `transferSourceFileChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10032:48",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1"
    }
  ]
}
````


**parser.ts:10104:37** `unsetNodeChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10096:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10104:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10096:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 506691,
      "body_end": 507655,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10087:5",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ]
}
````


**parser.ts:10261:42** `unsetNodeChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10246:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10261:42",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10246:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 513350,
      "body_end": 515010,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10233:5",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ]
}
````


**program.ts:2973:42** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:2966:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 144938,
      "body_end": 145801,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "markPrecedingCommentDirectiveLine",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:1516:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:1516:1",
      "depth": 1
    }
  ]
}
````


**program.ts:671:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:667:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 24128,
      "body_end": 24786,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatDiagnostic",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:667:1"
    }
  ]
}
````


**program.ts:713:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:714:87** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:715:58** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:733:57** `getPositionOfLineAndCharacter`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:733:57",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:733:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:734:76** `getPositionOfLineAndCharacter`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:734:76",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:734:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:768:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:767:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 28990,
      "body_end": 29611,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatLocation",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:767:1"
    }
  ]
}
````


**tracing.ts:206:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:38:1"
    }
  ]
}
````


**tracing.ts:207:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:38:1"
    }
  ]
}
````


**transformers/declarations.ts:1264:139** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ]
}
````


**transformers/declarations.ts:1264:70** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ]
}
````


**transformers/jsx.ts:359:63** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/jsx.ts:333:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 14898,
      "body_end": 16980,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitJsxOpeningLikeElementOrFragmentJSX",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/jsx.ts:92:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/jsx.ts:92:1",
      "depth": 1
    }
  ]
}
````


**utilities.ts:1025:47** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:1023:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 33483,
      "body_end": 33667,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "nodePosToString",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:1023:1"
    }
  ]
}
````


**utilities.ts:1213:46** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:1210:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 42294,
      "body_end": 43143,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "createCommentDirectivesMap",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:1210:1"
    }
  ]
}
````


**utilities.ts:2526:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1"
    }
  ]
}
````


**utilities.ts:2527:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1"
    }
  ]
}
````


**utilities.ts:2531:70** `getEndLinePosition`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2531:70",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2531:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1"
    }
  ]
}
````


**utilities.ts:6748:38** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:6747:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 266234,
      "body_end": 266342,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineOfLocalPosition",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:6747:1"
    }
  ]
}
````


**utilities.ts:7944:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7942:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7944:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7942:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313431,
      "body_end": 313609,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndAndRangeStart",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7942:1"
    }
  ]
}
````


**utilities.ts:7949:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7948:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7949:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7948:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313781,
      "body_end": 313860,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndPositions",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7948:1"
    }
  ]
}
````


**utilities.ts:7959:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7958:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7959:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7958:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314169,
      "body_end": 314242,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "positionsAreOnSameLine",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7958:1"
    }
  ]
}
````


**utilities.ts:7971:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7968:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7971:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7968:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314729,
      "body_end": 315003,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7968:1"
    }
  ]
}
````


**utilities.ts:7977:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7975:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7977:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7975:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 315182,
      "body_end": 315372,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7975:1"
    }
  ]
}
````


**watch.ts:249:60** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/watch.ts:233:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 8061,
      "body_end": 8940,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getFilesInErrorForSummary",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/watch.ts:233:1"
    }
  ]
}
````


### Full-mode bounds and retained evidence

Each final invocation retains all 79 compiler roots and uses LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and LATENT_ONLY_FILE equal to the exact selected file. The nine files with original findings plus scanner.ts are measured independently, with at most two concurrent file censuses and a 1500-second bound per file. Every native child is waited for, terminated as a process group at its bound if needed, and reports its own exit. Checkpoints are pushed after pieces to the checkpoint branch.

| File | Child exit | Seconds | Full stream completed | Completed timing units |
| --- | ---: | ---: | --- | ---: |
| scanner.ts | 0 | 297.22 | true | 152 |
| tracing.ts | 0 | 40.73 | true | 23 |
| watch.ts | 0 | 59.77 | true | 36 |
| transformers/jsx.ts | 0 | 110.21 | true | 40 |
| transformers/declarations.ts | 0 | 117.43 | true | 69 |
| executeCommandLine.ts | 0 | 133.51 | true | 68 |
| emitter.ts | 0 | 488.36 | true | 423 |
| utilities.ts | 0 | 598.82 | true | 867 |
| program.ts | 0 | 794.22 | true | 235 |
| parser.ts | 0 | 748.93 | true | 545 |

The legacy measurement label is fixed text; actual header checker_rejected flags and diagnostics determine checker status. All final (e) compiler headers report checker_rejected=false. Stream completion does not erase latent body boundaries or the original checker.ts site-count limitation. No backend or JavaScript output is produced.

Final data is under continuation/e/, e-exact/ and sdk-e/. e-analysis.json and e-site-comparison.json retain the complete tables and additions/removals. Abandoned controls have ABANDONED.json markers and are excluded from every final count. New Go evidence uses .go.txt. Files over 1,000,000 bytes are gzip compressed and SHA256SUMS covers the retained files.
