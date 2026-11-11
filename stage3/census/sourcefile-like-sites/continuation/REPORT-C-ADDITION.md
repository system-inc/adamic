## Third measurement: named LineMapSource, variant (c)

This measurement starts again from the same frozen adapted input, main d690f276749f91e6b36fe3a9726a5a453fcd7d39, TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8, and the same guarded meter binary. It does not apply the (a) or (b) scratch edits. The second report commit is 7944b20a23c7e2e40c00e925603c20bd7267a569. Nothing in compiler, services, cohere, or the adaptation rules lands.

**Overall outcome: no tested exact (c) form meets every requested gate. The internal marker breaks declaration bundling; removing it introduces a new public SDK optional-widening refusal. The internal compiler view counts below must not be read as end-to-end acceptance.**

### Exact source changes

Immediately before SourceFileLike in compiler/types.ts, (c) inserts:

```ts
/** @internal */
export interface LineMapSource {
    readonly text: string;
    lineMap?: readonly number[] | undefined;
}
```

SourceFileLike becomes `extends LineMapSource`. Its own text and lineMap member types are byte-for-byte identical to these declarations in the frozen adapted input, so both are removed. Its optional getPositionOfLineAndCharacter member is retained unchanged. The removed lineMap member had its own @internal comment. That comment is removed with the member; LineMapSource.lineMap has exactly the requested spelling, without a member-level internal marker. This matters when LineMapSource is made public.

Only getLineStarts and getLineAndCharacterOfPosition change their first parameter to LineMapSource. A type-only import of LineMapSource is necessary in scanner.ts. It is placed on the existing import line to retain all scanner line coordinates. The two function bodies, all other parameters, and all call expressions are unchanged. Full src hashes verify that scanner.ts and types.ts are the only changed source files. The patches and hashes are retained.

The public alternative removes only LineMapSource's @internal tag. A blank comment retains all line coordinates. Its runtime and relation-bearing syntax is otherwise identical, but it is independently rechecked and independently censused over the same ten files.

### Measured result for all 34 original views

| Variant | Complete single-file streams | Cleared at original view | Unchanged finding | Replacement finding | Checker diagnostic | Incomplete or excluded | Core checker diagnostics | Core first stop |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| (c), internal LineMapSource | 10/10 | 22 | 12 | 0 | 0 | 0 | 0 | `Refused tracing.ts:232:34`: a cast the runtime can't check |
| (c), public LineMapSource | 10/10 | 22 | 12 | 0 | 0 | 0 | 0 | `Refused tracing.ts:232:34`: a cast the runtime can't check |

A cleared view means that its enclosing unit was attempted and the guarded full census reports no finding or checker diagnostic at that original argument/call span. It does not mean that the whole function or root lowers. The 22 helper views consist of 21 SourceFile values and one SourceMapSource value. The other 12 views still call unmodified SourceFileLike callees. The original checker.ts census remains cut off after 25 minutes, so this is a complete comparison of the 34 observed views, not a new claim that the original 79-file site count is exhaustive.

| Callee | Original sites | (c) cleared | (c) unchanged | Public (c) cleared | Public (c) unchanged |
| --- | ---: | ---: | ---: | ---: | ---: |
| `getEndLinePosition` | 1 | 0 | 1 | 0 | 1 |
| `getLineAndCharacterOfPosition` | 18 | 18 | 0 | 18 | 0 |
| `getLineStarts` | 4 | 4 | 0 | 4 | 0 |
| `getLinesBetweenPositions` | 5 | 0 | 5 | 0 | 5 |
| `getPositionOfLineAndCharacter` | 2 | 0 | 2 | 0 | 2 |
| `transferSourceFileChildren` | 2 | 0 | 2 | 0 | 2 |
| `unsetNodeChildren` | 2 | 0 | 2 | 0 | 2 |

### Adamic relations, independent of stock tsc

The pinned cohere binary runs the full cohere:adamic rule set under strict, noUncheckedIndexedAccess, exactOptionalPropertyTypes, verbatimModuleSyntax, ES2024, ESNext/Bundler, and the remaining production loader options recorded in cohere-tsconfig.json and load-options.go.txt. The CLI actually walks 77 compiler source files out of its 79-file program; its complete scope, rule counts, unresolved-input warnings, and findings are retained. The first CLI attempt refused to lint import-only files; its zero-rule failure is preserved separately and is not used as a measurement. The successful rerun explicitly claims all 79 compiler files.

This runs the actual no-optional-widening and invariant-mutable implementations, not a stock tsc approximation. Required SourceFile.lineMap is readonly number[] | undefined in the adapted tree. The target optional slot explicitly holds undefined too. No rule reports this SourceFile-to-LineMapSource relation at any of the 22 helper views. Existing non-null assertion findings on some enclosing calls are quoted below; they are not new relation findings.

| Variant | Cohere rules | Compiler files checked | New cohere findings | Removed cohere findings | New measured lowering findings | Removed measured lowering findings |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| (c), internal LineMapSource | 40 | 77 | 0 | 24 | 0 | 22 |
| (c), public LineMapSource | 40 | 77 | 0 | 24 | 0 | 22 |

The cohere comparison includes every finding, not just the helper calls. Its removed findings include the two checker.ts calls that were not in the 34-view lowering list. The lowering comparison includes every finding in the ten completed selected-file streams, including boundaries and unrelated reasons. Path and unchanged types.ts line shifts are normalized; no reason is filtered out. Exact additions and removals are in c-analysis.json. A newly visible later boundary, if present in that comparison, is not silently counted as successful lowering.

Cohere additions and removals for (c):

````json
{
  "baseline_unique": 9728,
  "variant_unique": 9704,
  "added": [],
  "removed": [
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
}
````


### getLineStarts own unit and the cache write

The exact unchanged cache expression is:

```ts
return sourceFile.lineMap || (sourceFile.lineMap = computeLineStarts(sourceFile.text));
```

getLineStarts is scanner.ts:503:1; the assignment is scanner.ts:504:35. Every finding and the unit status from its own full-mode stream follows.

**(c), internal LineMapSource**

````json
{
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:503:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 49075,
      "body_end": 49173,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineStarts",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:503:1"
    }
  ],
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:504:35",
      "reason": "an assignment value to a member",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:504:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet",
      "start": 49077,
      "end": 49170
    }
  ]
}
````


**(c), public LineMapSource**

````json
{
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:503:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 49075,
      "body_end": 49173,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineStarts",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:503:1"
    }
  ],
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:504:35",
      "reason": "an assignment value to a member",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:504:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet",
      "start": 49077,
      "end": 49170
    }
  ]
}
````


The corresponding baseline findings are:

````json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:503:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:504:35",
    "reason": "an assignment value to a member",
    "text": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:503:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:504:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet",
    "start": 49058,
    "end": 49151
  }
]
````


Thus the measured writer stop is the pre-existing NotYet reason `an assignment value to a member`, not the #x0nd1mh optional-checked-view write reason. The normal core.ts first stop after each form of (c) is the unrelated cast in tracing.ts:232:34, not this write. The conditional request to count all distinct callers if the write becomes that first stop does not apply.

The implementation of checkViewOptionalWrites is preserved in view-optional-writes.go.txt. It refuses a write to an optional field when that receiver/field is checked; it does not refuse every optional field assignment merely because the declaration uses a question mark. The existing assignment-expression stop precedes that later check here, so it cannot be used as proof that the cache write is supported through every checked view. No actual (c) stream or cohere relation finding reports the required-to-optional writable-slot refusal posited in the conditional request. The requested condition for (d) therefore does not fire; (d) is not substituted for the measured (c) source.

Two separately labeled minimal controls are retained in c/probes/. The original expression reproduces the assignment-value NotYet. The control that separates the write into a statement is accepted for its required-to-optional record pair. Neither control changes (c), replaces its real-file findings, or establishes support for other shapes.

Core first-stop records (the driver also repeats its one-root input as a whole-program record; the first record is quoted):

**(c), internal LineMapSource**

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-c/src/compiler/core.ts"
  ],
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:232:34",
  "reason": "a cast the runtime can't check",
  "fix": "use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
  "seconds": 21.274722228999998
}
````


**(c), public LineMapSource**

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-c-public/src/compiler/core.ts"
  ],
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:232:34",
  "reason": "a cast the runtime can't check",
  "fix": "use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
  "seconds": 44.793428969
}
````


### Whole src, production checker, and services

All whole-src checks use the exact same union of 604 roots from all 13 src solution projects. The stock checks keep the separate lockfile dependency installation and the project source-check settings used for (a) and (b). They do not write node_modules into any frozen or census input tree. The extra exact-optional stock run changes only exactOptionalPropertyTypes. This is distinct from production load.Load, which uses every Adamic compiler option and its own environment.

For the actual production checker, adamic-check.go.txt calls unmodified load.Load over those 604 root paths and never calls lowering or an emitter. Every checker-rejected result is required to have no usable Program. This is the native Adamic checker, not stock tsc configured to approximate it. Its own full-src baseline is already rejected: many diagnostics concern type-only imports under verbatimModuleSyntax and test globals unavailable under types: []. Those errors are retained and compared, not attributed to (c).

| Tree | Stock whole-src errors | Stock with exactOptionalPropertyTypes | Production Adamic whole-src errors | New production checker errors | Services errors, stock | Services errors, exact optional | Services TS2386, stock / exact |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Frozen baseline | 0 | 416 | 6851 | 0 | 0 | 142 | 0 / 0 |
| (c), internal LineMapSource | 0 | 416 | 6851 | 0 | 0 | 142 | 0 / 0 |
| (c), public LineMapSource | 0 | 416 | 6851 | 0 | 0 | 142 | 0 / 0 |

The production diagnostic sets are compared by their complete text after scratch-path and unchanged type-declaration line mapping. The stock exact-optional diagnostic sets are compared by file, location, code, category, and message. The same baseline errors remain, with additions and removals recorded in c-analysis.json. Core.ts is separately checked by the ordinary first-stop driver with actual Adamic options; it has no checker diagnostics under either form of (c).

Every one of the 93 resolved scanner-helper calls remains in the stock whole-src audit. SourceFileLike arguments and text-only object arguments are accepted by the optional LineMapSource contract; (b)'s 15 call errors do not recur. The only declared argument type changed by this edit is scanner.ts:549's sourceFile parameter, now LineMapSource instead of SourceFileLike. The declarations, flow types, and full diagnostic streams for all calls remain in each calls.json. The two diagnostics whose spans happen to enclose calls under exact optional checking are the same pre-existing enclosing object-literal TS2375 diagnostics in server/session.ts, not helper-parameter incompatibilities.

### Public declarations with stripInternal on and off

Actual stock declaration emission and the unchanged upstream declaration bundler are used, as before. All emitted-file hashes and options are retained. The bundler independently separates its public typescript.d.ts from typescript.internal.d.ts using @internal tags, even when the TypeScript compiler's stripInternal option is off. This distinction changes what the off result means.

| LineMapSource | Compiler stripInternal | Declaration emit | Upstream bundle | Public LineMapSource | Public getLineStarts | Public closure/consumer check | Lane API guard |
| --- | --- | --- | --- | --- | --- | --- | --- |
| @internal | on | 278 files, no emit diagnostics | Fails, child exit 1 | Stripped | Stripped | Dangling emitted references | Build fails before a bundle exists |
| @internal | off | 278 files, no emit diagnostics | Succeeds | absent | absent | 2 diagnostic(s) | 4 unsanctioned API changes |
| public | on | 278 files, no emit diagnostics | Succeeds | present | absent | 0 diagnostic(s) | 7 unsanctioned API changes |
| public | off | 278 files, no emit diagnostics | Succeeds | present | absent | 0 diagnostic(s) | 7 unsanctioned API changes |

**Internal (c), stripInternal on.** The raw emitted scanner.d.ts still imports LineMapSource and declares `getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter`. Emitted types.d.ts declares `SourceFileLike extends LineMapSource` but no LineMapSource. The upstream bundler then aborts with `AssertionError [ERR_ASSERTION]: No symbol for node at /workspace/scratch/baobab-emit-c/compiler/types.d.ts:2000:40`. No public typescript.d.ts is produced in this configuration. It would be inaccurate to present a fabricated successful bundle or diff. The raw emitted selected declarations and all diagnostics are:

````json
{
  "stripInternal": true,
  "diagnostics": [
    {
      "code": 2305,
      "file": "compiler/scanner.d.ts",
      "line": 1,
      "column": 181,
      "message": "Module '\"./_namespaces/ts.js\"' has no exported member 'LineMapSource'."
    },
    {
      "code": 2304,
      "file": "compiler/types.d.ts",
      "line": 2000,
      "column": 41,
      "message": "Cannot find name 'LineMapSource'."
    },
    {
      "code": 2304,
      "file": "compiler/utilities.d.ts",
      "line": 23,
      "column": 29,
      "message": "Cannot find name 'Mutable'."
    },
    {
      "code": 2304,
      "file": "server/project.d.ts",
      "line": 114,
      "column": 118,
      "message": "Cannot find name 'EmitResult'."
    }
  ],
  "selected_declarations": [
    {
      "file": "compiler/types.d.ts",
      "name": "SourceFileLike",
      "text": "export interface SourceFileLike extends LineMapSource {\n}"
    },
    {
      "file": "compiler/scanner.d.ts",
      "name": "getLineAndCharacterOfPosition",
      "text": "export declare function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;"
    }
  ],
  "bundled_typescript_dts_produced": false,
  "bundler_child_exit": 1
}
````


The raw stripped baseline has pre-existing dangling Mutable and EmitResult references too; its analysis is retained as original/raw-declaration-analysis.json. The additional c references are the missing scanner import TS2305 and missing LineMapSource base TS2304. The baseline upstream public bundle succeeds and has zero closure/consumer diagnostics.

**Internal (c), stripInternal off.** The raw module declarations retain LineMapSource, so bundling succeeds. The bundler still excludes its @internal interface from the public bundle. The public d.ts literally contains these declarations:

````json
[
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike extends LineMapSource {\n    }"
  },
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike {\n        getLineAndCharacterOfPosition(pos: number): LineAndCharacter;\n    }"
  },
  {
    "name": "getLineAndCharacterOfPosition",
    "kind": "FunctionDeclaration",
    "text": "function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;"
  }
]
````


The public declaration checker reports:

````json
[
  {
    "code": 2304,
    "file": "typescript.d.ts",
    "line": 5906,
    "column": 38,
    "message": "Cannot find name 'LineMapSource'."
  },
  {
    "code": 2304,
    "file": "typescript.d.ts",
    "line": 8553,
    "column": 56,
    "message": "Cannot find name 'LineMapSource'."
  }
]
````


Its unchanged stage3/lane check.api_check rejects the API snapshot as follows:

````json
[
  "unsanctioned API declaration: FunctionDeclaration:[\"ts\",\"getLineAndCharacterOfPosition\"]#0",
  "unsanctioned API declaration: InterfaceDeclaration:[\"ts\",\"SourceFileLike\"]#0",
  "unsanctioned API declaration: comments#0",
  "unsanctioned API declaration: member:[\"ts\",\"SourceFileLike\",\"text\"]#0"
]
````


**Public (c), both settings.** Removing the interface-level marker produces a public LineMapSource and a valid public consumer. Its inherited lineMap member is now public too, because the requested interface did not put @internal on that member. The public declarations are:

````json
[
  {
    "name": "LineMapSource",
    "kind": "InterfaceDeclaration",
    "text": "interface LineMapSource {\n        readonly text: string;\n        lineMap?: readonly number[] | undefined;\n    }"
  },
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike extends LineMapSource {\n    }"
  },
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike {\n        getLineAndCharacterOfPosition(pos: number): LineAndCharacter;\n    }"
  },
  {
    "name": "getLineAndCharacterOfPosition",
    "kind": "FunctionDeclaration",
    "text": "function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;"
  }
]
````


The complete textual diffs against the matching baseline settings are retained in c-public/typescript-public.diff and c-public-unstripped/typescript-public.diff. getLineStarts remains internal and absent from public typescript.d.ts in both cases. With stripInternal off it is retained in the raw scanner declarations and the internal bundle; it does not enter the public bundle. The lane guard still rejects seven unsanctioned public API changes:

````json
[
  "unsanctioned API declaration: FunctionDeclaration:[\"ts\",\"getLineAndCharacterOfPosition\"]#0",
  "unsanctioned API declaration: InterfaceDeclaration:[\"ts\",\"LineMapSource\"]#0",
  "unsanctioned API declaration: InterfaceDeclaration:[\"ts\",\"SourceFileLike\"]#0",
  "unsanctioned API declaration: comments#0",
  "unsanctioned API declaration: member:[\"ts\",\"LineMapSource\",\"lineMap\"]#0",
  "unsanctioned API declaration: member:[\"ts\",\"LineMapSource\",\"text\"]#0",
  "unsanctioned API declaration: member:[\"ts\",\"SourceFileLike\",\"text\"]#0"
]
````


The frozen baseline passes the actual lane API guard with zero errors. No API sanction is added for this report-only experiment. stage3/lane uses its normalized declaration/snapshot guard, not an installed Microsoft API Extractor invocation; the independent no-skipLibCheck public consumer check above supplies the missing-type closure diagnostics. Thus making LineMapSource public repairs the dangling type but does not make the unchanged landing lane accept a new public API.

The successful public-(c) typescript.d.ts bundles are byte-for-byte identical with compiler stripInternal on and off, as are the matching baseline public bundles. The raw declarations and internal bundles differ: getLineStarts is retained only when compiler stripInternal is off. c-declaration-matrix.json records the exact selected declarations for every mode.

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



### Every original view under both forms of (c)

Each row keeps the original location, phase, enclosing function, callee, every current finding at that argument/call, and checker diagnostics. When a call has two original views its argument span is used separately, as in the earlier comparison. The cohere list includes existing non-null assertion findings on enclosing calls, so relation acceptance is not confused with an all-rules-clean source expression.

#### (c), internal LineMapSource

| Original location | Phase | Enclosing function | Callee | Measured classification |
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
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1443:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 58788,
      "body_end": 58904,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getCurrentLineMap",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:3961:112** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:3961:180** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:6233:96** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:6228:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 268634,
      "body_end": 269150,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitPos",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**executeCommandLine.ts:109:41** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/executeCommandLine.ts:105:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 2705,
      "body_end": 2984,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "countLines",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/executeCommandLine.ts:105:1"
    }
  ],
  "cohere_findings_at_call": []
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10032:36",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10032,
      "column": 36,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10032,
      "column": 48,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`result` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10032:48",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10032,
      "column": 36,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10032,
      "column": 48,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`result` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10096:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10104:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10096:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 506691,
      "body_end": 507655,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10087:5",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10104,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`origSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10246:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10261:42",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10246:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 513350,
      "body_end": 515010,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10233:5",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10261,
      "column": 42,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:2973:42** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:2966:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 144938,
      "body_end": 145801,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "markPrecedingCommentDirectiveLine",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:1516:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:1516:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:671:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:667:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 24128,
      "body_end": 24786,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatDiagnostic",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:667:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts",
      "line": 671,
      "column": 84,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**program.ts:713:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:714:87** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:715:58** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:733:57",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:733:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts",
      "line": 733,
      "column": 57,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:734:76",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:734:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts",
      "line": 734,
      "column": 76,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:768:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:767:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 28990,
      "body_end": 29611,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatLocation",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:767:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**tracing.ts:206:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:38:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts",
      "line": 206,
      "column": 73,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**tracing.ts:207:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:38:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts",
      "line": 207,
      "column": 71,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**transformers/declarations.ts:1264:139** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**transformers/declarations.ts:1264:70** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**transformers/jsx.ts:359:63** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/jsx.ts:333:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 14898,
      "body_end": 16980,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitJsxOpeningLikeElementOrFragmentJSX",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/jsx.ts:92:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/jsx.ts:92:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:1025:47** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:1023:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 33483,
      "body_end": 33667,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "nodePosToString",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:1023:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:1213:46** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:1210:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 42294,
      "body_end": 43143,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "createCommentDirectivesMap",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:1210:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2526:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2527:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": []
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2531:70",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2531:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 2531,
      "column": 70,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:6748:38** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:6747:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 266234,
      "body_end": 266342,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineOfLocalPosition",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:6747:1"
    }
  ],
  "cohere_findings_at_call": []
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7942:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7944:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7942:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313431,
      "body_end": 313609,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndAndRangeStart",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7942:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7944,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7948:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7949:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7948:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313781,
      "body_end": 313860,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndPositions",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7948:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7949,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7958:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7959:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7958:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314169,
      "body_end": 314242,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "positionsAreOnSameLine",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7958:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7959,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7968:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7971:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7968:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314729,
      "body_end": 315003,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7968:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7971,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7975:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7977:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7975:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 315182,
      "body_end": 315372,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7975:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7977,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**watch.ts:249:60** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/watch.ts:233:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 8061,
      "body_end": 8940,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getFilesInErrorForSummary",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/watch.ts:233:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/watch.ts",
      "line": 249,
      "column": 60,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/watch.ts",
      "line": 249,
      "column": 89,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


#### (c), public LineMapSource

| Original location | Phase | Enclosing function | Callee | Measured classification |
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
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1443:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 58788,
      "body_end": 58904,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getCurrentLineMap",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:3961:112** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:3961:180** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:6233:96** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:6228:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 268634,
      "body_end": 269150,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitPos",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**executeCommandLine.ts:109:41** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/executeCommandLine.ts:105:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 2705,
      "body_end": 2984,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "countLines",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/executeCommandLine.ts:105:1"
    }
  ],
  "cohere_findings_at_call": []
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10032:36",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10032,
      "column": 36,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10032,
      "column": 48,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`result` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10032:48",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10032,
      "column": 36,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10032,
      "column": 48,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`result` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10096:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10104:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10096:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 506691,
      "body_end": 507655,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10087:5",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10104,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`origSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10246:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10261:42",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10246:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 513350,
      "body_end": 515010,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10233:5",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10261,
      "column": 42,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:2973:42** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:2966:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 144938,
      "body_end": 145801,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "markPrecedingCommentDirectiveLine",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:1516:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:1516:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:671:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:667:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 24128,
      "body_end": 24786,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatDiagnostic",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:667:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts",
      "line": 671,
      "column": 84,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**program.ts:713:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:714:87** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:715:58** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:733:57",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:733:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts",
      "line": 733,
      "column": 57,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:734:76",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:734:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts",
      "line": 734,
      "column": 76,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:768:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:767:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 28990,
      "body_end": 29611,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatLocation",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:767:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**tracing.ts:206:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:38:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts",
      "line": 206,
      "column": 73,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**tracing.ts:207:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:38:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts",
      "line": 207,
      "column": 71,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**transformers/declarations.ts:1264:139** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**transformers/declarations.ts:1264:70** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**transformers/jsx.ts:359:63** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/jsx.ts:333:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 14898,
      "body_end": 16980,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitJsxOpeningLikeElementOrFragmentJSX",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/jsx.ts:92:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/jsx.ts:92:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:1025:47** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:1023:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 33483,
      "body_end": 33667,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "nodePosToString",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:1023:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:1213:46** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:1210:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 42294,
      "body_end": 43143,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "createCommentDirectivesMap",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:1210:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2526:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2527:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": []
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2531:70",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2531:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 2531,
      "column": 70,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:6748:38** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:6747:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 266234,
      "body_end": 266342,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineOfLocalPosition",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:6747:1"
    }
  ],
  "cohere_findings_at_call": []
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7942:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7944:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7942:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313431,
      "body_end": 313609,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndAndRangeStart",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7942:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7944,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7948:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7949:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7948:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313781,
      "body_end": 313860,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndPositions",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7948:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7949,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7958:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7959:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7958:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314169,
      "body_end": 314242,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "positionsAreOnSameLine",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7958:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7959,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7968:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7971:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7968:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314729,
      "body_end": 315003,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7968:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7971,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
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
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7975:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7977:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7975:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 315182,
      "body_end": 315372,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7975:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7977,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**watch.ts:249:60** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/watch.ts:233:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 8061,
      "body_end": 8940,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getFilesInErrorForSummary",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/watch.ts:233:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/watch.ts",
      "line": 249,
      "column": 60,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/watch.ts",
      "line": 249,
      "column": 89,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


### Reproducibility, bounds, and limits

Each of the nine files with original findings, plus scanner.ts for the writer unit, is run in an independent process with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_ONLY_FILE set to that exact file. All 79 compiler files remain roots. There are at most two concurrent file censuses; the public alternative starts only after all ten internal variant attempts finish. Every file has a 1500-second bound. Each child is waited for and its own exit code is recorded. Checkpoints are pushed after completed pieces.

| Variant | File | Child exit | Seconds | Complete full stream | Completed timing units |
| --- | --- | ---: | ---: | --- | ---: |
| (c), internal LineMapSource | scanner.ts | 0 | 613.42 | true | 150 |
| (c), internal LineMapSource | tracing.ts | 0 | 214.85 | true | 23 |
| (c), internal LineMapSource | watch.ts | 0 | 117.76 | true | 36 |
| (c), internal LineMapSource | transformers/jsx.ts | 0 | 151.39 | true | 40 |
| (c), internal LineMapSource | transformers/declarations.ts | 0 | 184.41 | true | 69 |
| (c), internal LineMapSource | executeCommandLine.ts | 0 | 191.80 | true | 68 |
| (c), internal LineMapSource | emitter.ts | 0 | 673.43 | true | 423 |
| (c), internal LineMapSource | utilities.ts | 0 | 746.02 | true | 867 |
| (c), internal LineMapSource | program.ts | 0 | 904.84 | true | 235 |
| (c), internal LineMapSource | parser.ts | 0 | 871.26 | true | 545 |
| (c), public LineMapSource | scanner.ts | 0 | 361.56 | true | 150 |
| (c), public LineMapSource | tracing.ts | 0 | 48.95 | true | 23 |
| (c), public LineMapSource | watch.ts | 0 | 78.66 | true | 36 |
| (c), public LineMapSource | transformers/jsx.ts | 0 | 121.01 | true | 40 |
| (c), public LineMapSource | transformers/declarations.ts | 0 | 153.92 | true | 69 |
| (c), public LineMapSource | executeCommandLine.ts | 0 | 155.27 | true | 68 |
| (c), public LineMapSource | emitter.ts | 0 | 612.61 | true | 423 |
| (c), public LineMapSource | utilities.ts | 0 | 712.32 | true | 867 |
| (c), public LineMapSource | program.ts | 0 | 839.37 | true | 235 |
| (c), public LineMapSource | parser.ts | 0 | 797.68 | true | 545 |

The overlay's historical measurement label remains constant even on a checker-accepted input. Every compiler census header for (c) records checker_rejected=false and diagnostics=[]. Use those actual flags rather than the legacy label. Completed streams still contain latent body boundaries and are not whole-program compilation proofs. No backend or JavaScript output is invoked. Full-src checker rejection, the unchanged writer NotYet, the unrelated core first stop, and the API guard failures above remain real limitations.

Evidence is under continuation/c/, c-public/, their declaration variants, and c-exact/, c-public-exact/, original-exact/, and original-rules/. c-site-comparison.json preserves all 68 view/variant rows. c-analysis.json preserves every comparison addition and removal. New Go evidence files end in .go.txt. Files over 1,000,000 bytes are gzip compressed. SHA256SUMS covers all retained evidence.
