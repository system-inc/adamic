## Follow-up: measured road 2 variants

The follow-up uses the same main pin, upstream pin, frozen adapted input, and guarded binary as the initial report. The six old cut-offs were retried first, sequentially, with a 1500-second bound each. Only after all six attempts and their checkpoints completed were the variants applied or measured.

### Six longer baseline attempts

| File | Child exit | Seconds | Complete file result | Completed timing records | Last completed timed unit |
| --- | ---: | ---: | --- | ---: | --- |
| checker.ts | -15 | 1500.23 | no | 1293 | checker.ts:27590:5 |
| emitter.ts | 0 | 479.19 | yes | 423 | emitter.ts:6370:1 |
| parser.ts | 0 | 722.02 | yes | 545 | parser.ts:10800:1 |
| program.ts | 0 | 729.78 | yes | 235 | program.ts:5191:1 |
| scanner.ts | 0 | 253.30 | yes | 150 | scanner.ts:4103:1 |
| utilities.ts | 0 | 643.31 | yes | 867 | utilities.ts:12436:1 |

**The count remains a lower bound because these files still did not finish in 25 minutes: checker.ts.** Their timing progress is recorded above and their partial streams are retained. A partial stream without an emitted file result does not mean zero sites.

Successful longer streams supersede their first-pass streams in the primary count tables. Both sets of raw streams remain in the evidence. The complete function/body inventory is independent of whether a file census finishes.

### Exact variants and measured outcomes

(a) changes the first parameter of both helpers to `{ readonly text: string; lineMap?: readonly number[] }`.

(b) changes it to `{ readonly text: string; lineMap: readonly number[] | undefined }`.

Each scratch tree was produced with the pinned apply script/product before replacing those two annotations. Full src-file hashes prove that scanner.ts is the only changed source file; byte comparisons prove that its only edits are the two parameter types. No SourceFile or services interface was changed. SourceFile.lineMap in this adapted tree is required and has the exact type `readonly number[] | undefined` (compiler/types.ts:4436). SourceFileLike.lineMap is optional. The upstream SourceFile member is required readonly number[]; the unchanged applied baseline adds the undefined union. The applied SourceMapSource.lineMap is optional too. lineMap-declarations.txt quotes both source versions. These are baseline adaptations, not changes made by either variant.

Adamic enables exactOptionalPropertyTypes (internal/load/load.go:63; load-options.go.txt records the options). Its (a) core check reports TS2379 because the adapted required SourceFile.lineMap can be undefined while (a) does not explicitly permit undefined when the optional property is present. The stock whole-src checks below retain the repository project defaults, where exactOptionalPropertyTypes is false. Thus a zero stock (a) error count does not imply that (a) passes Adamic. The core records and guarded-census diagnostics use the actual Adamic settings.

The variants use separate single-file processes, at most one process for each variant concurrently. Each retains all 79 compiler roots and uses LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and an exact LATENT_ONLY_FILE. Each variant file has the same 1500-second safety bound. Checkpoints follow every completed pair of file attempts.

| Variant | Completed file streams | Baseline view locations | Unchanged refusal at same view | Replacement finding at same view | Checker diagnostic at same call | No finding there, unit attempted | Other / incomplete | Whole-src stock errors | Services TS2386 | Core first stop |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| (a) | 9/9 | 34 | 9 | 0 | 22 | 0 | 3 | 0 | 0 | checker |
| (b) | 9/9 | 34 | 12 | 0 | 1 | 21 | 0 | 15 | 0 | checker |

These columns report observations at the original views, not successful lowering. A replacement finding is a different kind or reason at the same call/argument span. Unchanged findings include unmodified SourceFileLike callees. A call with two baseline SourceFileLike arguments is compared by argument span so findings are not attributed twice. A missing finding is not counted as a clearance if the file stream is incomplete or its enclosing unit is excluded. Other blockers elsewhere remain in the complete raw streams.

The original core.ts first stop and both variant first stops follow. The ordinary census driver emits a per-root record and then its one-root whole-program record; the first record is the requested per-root first stop. It never invokes a backend.

Core root original:

````json
{
  "roots": [
    "/workspace/scratch/baobab-adapted/src/compiler/core.ts"
  ],
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-adapted/src/compiler/tracing.ts:206:67",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "fix": "declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
  "seconds": 14.400678442
}
````


Core root a:

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-a/src/compiler/core.ts"
  ],
  "diagnostics": [
    "/workspace/scratch/baobab-road2-a/src/compiler/checker.ts:52424:57: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/checker.ts:52425:55: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:1444:66: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:112: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:180: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:6233:96: error TS2379: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts:109:41: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:2973:42: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:671:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:768:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/scanner.ts:471:57: error TS2379: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/scanner.ts:539:38: error TS2379: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:206:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:207:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:139: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:70: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts:359:63: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1018:26: error TS2379: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1025:47: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1032:38: error TS2379: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1213:46: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:6748:38: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts:249:60: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
  ],
  "kind": "checker",
  "seconds": 5.314485692
}
````


Core root b:

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-b/src/compiler/core.ts"
  ],
  "diagnostics": [
    "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:6233:96: error TS2345: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
    "/workspace/scratch/baobab-road2-b/src/compiler/scanner.ts:471:57: error TS2345: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
    "/workspace/scratch/baobab-road2-b/src/compiler/scanner.ts:539:38: error TS2345: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
    "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:1018:26: error TS2345: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
    "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:1032:38: error TS2345: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'."
  ],
  "kind": "checker",
  "seconds": 3.924856896
}
````


### Every original view under each variant

The baseline table above names every observed original SourceFileLike view; the checker cutoff remains unknown. The blocks below quote every distinct finding matching each original call/argument span, including reasons that no longer name SourceFileLike. Checker diagnostics at those calls are quoted too. Unit status and stream completeness distinguish an unreported view from an excluded or unfinished measurement. Finding identities are deduplicated by kind, where, reason, text, phase, and unit; the raw streams retain repeated observations.

#### Variant (a)

**emitter.ts:1444:66**, `getLineStarts`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts",
  "start": 58857,
  "end": 58894,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:1444:66: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "emitter.ts:1444:66"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:1443:5",
    "name": "getCurrentLineMap",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:1444:66: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**emitter.ts:3961:112**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts",
  "start": 171010,
  "end": 171027,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:112: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "emitter.ts:3961:112"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3959:5",
    "name": "emitJsxExpression",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:112: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:180: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**emitter.ts:3961:180**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts",
  "start": 171078,
  "end": 171095,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:180: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "emitter.ts:3961:180"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3959:5",
    "name": "emitJsxExpression",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:112: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:180: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**emitter.ts:6233:96**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts",
  "start": 268874,
  "end": 268889,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:6233:96: error TS2379: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "emitter.ts:6233:96"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:6228:5",
    "name": "emitPos",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:6233:96: error TS2379: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**executeCommandLine.ts:109:41**, `getLineStarts`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts",
  "start": 2883,
  "end": 2887,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts:109:41: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "executeCommandLine.ts:109:41"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts:105:1",
    "name": "countLines",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts:109:41: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**parser.ts:10032:36**, `transferSourceFileChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:9947:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10032:36",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:9947:5",
    "name": "updateSourceFile",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10032:48**, `transferSourceFileChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:9947:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10032:48",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:9947:5",
    "name": "updateSourceFile",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10104:37**, `unsetNodeChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10096:9",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10104:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10096:9",
    "name": "visitNode",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10261:42**, `unsetNodeChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10246:9",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10261:42",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10246:9",
    "name": "visitNode",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:671:67**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 24412,
  "end": 24427,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:671:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:671:67"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:667:1",
    "name": "formatDiagnostic",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:671:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:713:89**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 26154,
  "end": 26158,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:713:89"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:714:87**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 26255,
  "end": 26259,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:714:87"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:715:58**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 26336,
  "end": 26340,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:715:58"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:733:57**, `getPositionOfLineAndCharacter`: enclosing unit excluded or not attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:734:76**, `getPositionOfLineAndCharacter`: enclosing unit excluded or not attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:768:89**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 29082,
  "end": 29086,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:768:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:768:89"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:767:1",
    "name": "formatLocation",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:768:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:2973:42**, `getLineStarts`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 145144,
  "end": 145148,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:2973:42: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:2973:42"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:2966:5",
    "name": "markPrecedingCommentDirectiveLine",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:2973:42: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**tracing.ts:206:67**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts",
  "start": 8172,
  "end": 8176,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:206:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "tracing.ts:206:67"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:200:5",
    "name": "getLocation",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:206:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:207:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**tracing.ts:207:65**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts",
  "start": 8256,
  "end": 8260,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:207:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "tracing.ts:207:65"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:200:5",
    "name": "getLocation",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:206:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:207:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**transformers/declarations.ts:1264:70**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts",
  "start": 62455,
  "end": 62472,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:70: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "transformers/declarations.ts:1264:70"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1006:5",
    "name": "visitDeclarationSubtree",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:70: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:139: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**transformers/declarations.ts:1264:139**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts",
  "start": 62524,
  "end": 62541,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:139: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "transformers/declarations.ts:1264:139"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1006:5",
    "name": "visitDeclarationSubtree",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:70: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:139: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**transformers/jsx.ts:359:63**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts",
  "start": 16101,
  "end": 16113,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts:359:63: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "transformers/jsx.ts:359:63"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts:333:5",
    "name": "visitJsxOpeningLikeElementOrFragmentJSX",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts:359:63: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:1025:47**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 33578,
  "end": 33582,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1025:47: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:1025:47"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1023:1",
    "name": "nodePosToString",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1025:47: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:1213:46**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 42435,
  "end": 42445,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1213:46: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:1213:46"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1210:1",
    "name": "createCommentDirectivesMap",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1213:46: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:2526:67**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 88023,
  "end": 88033,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:2526:67"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:2527:65**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 88116,
  "end": 88126,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:2527:65"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:2531:70**, `getEndLinePosition`: enclosing unit excluded or not attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:6748:38**, `getLineStarts`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 266275,
  "end": 266285,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:6748:38: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:6748:38"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:6747:1",
    "name": "getLineOfLocalPosition",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:6748:38: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:7944:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7942:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7944:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7942:1",
    "name": "getLinesBetweenRangeEndAndRangeStart",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7949:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7948:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7949:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7948:1",
    "name": "getLinesBetweenRangeEndPositions",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7959:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7958:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7959:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7958:1",
    "name": "positionsAreOnSameLine",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7971:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7968:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7971:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7968:1",
    "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7977:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7975:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7977:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7975:1",
    "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**watch.ts:249:60**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts",
  "start": 8760,
  "end": 8787,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts:249:60: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "watch.ts:249:60"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts:233:1",
    "name": "getFilesInErrorForSummary",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts:249:60: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


#### Variant (b)

**emitter.ts:1444:66**, `getLineStarts`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:1443:5",
    "name": "getCurrentLineMap",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**emitter.ts:3961:112**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:3959:5",
    "name": "emitJsxExpression",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**emitter.ts:3961:180**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:3959:5",
    "name": "emitJsxExpression",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**emitter.ts:6233:96**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

````json
{
  "file": "src/compiler/emitter.ts",
  "line": 6233,
  "column": 96,
  "start": 268874,
  "length": 15,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts",
  "start": 268874,
  "end": 268889,
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:6233:96: error TS2345: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "mapped_location": "emitter.ts:6233:96"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:6228:5",
    "name": "emitPos",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:6233:96: error TS2345: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'."
    ]
  }
]
````


**executeCommandLine.ts:109:41**, `getLineStarts`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/executeCommandLine.ts:105:1",
    "name": "countLines",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10032:36**, `transferSourceFileChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:9947:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10032:36",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:9947:5",
    "name": "updateSourceFile",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10032:48**, `transferSourceFileChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:9947:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10032:48",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:9947:5",
    "name": "updateSourceFile",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10104:37**, `unsetNodeChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10096:9",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10104:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10096:9",
    "name": "visitNode",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10261:42**, `unsetNodeChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10246:9",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10261:42",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10246:9",
    "name": "visitNode",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:671:67**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:667:1",
    "name": "formatDiagnostic",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:713:89**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:714:87**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:715:58**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:733:57**, `getPositionOfLineAndCharacter`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:733:57",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:733:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:734:76**, `getPositionOfLineAndCharacter`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:734:76",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:734:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:768:89**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:767:1",
    "name": "formatLocation",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:2973:42**, `getLineStarts`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:2966:5",
    "name": "markPrecedingCommentDirectiveLine",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**tracing.ts:206:67**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/tracing.ts:200:5",
    "name": "getLocation",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**tracing.ts:207:65**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/tracing.ts:200:5",
    "name": "getLocation",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**transformers/declarations.ts:1264:70**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/transformers/declarations.ts:1006:5",
    "name": "visitDeclarationSubtree",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**transformers/declarations.ts:1264:139**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/transformers/declarations.ts:1006:5",
    "name": "visitDeclarationSubtree",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**transformers/jsx.ts:359:63**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/transformers/jsx.ts:333:5",
    "name": "visitJsxOpeningLikeElementOrFragmentJSX",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:1025:47**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:1023:1",
    "name": "nodePosToString",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:1213:46**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:1210:1",
    "name": "createCommentDirectivesMap",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:2526:67**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:2527:65**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:2531:70**, `getEndLinePosition`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2523:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2531:70",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2531:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:6748:38**, `getLineStarts`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:6747:1",
    "name": "getLineOfLocalPosition",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7944:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7942:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7944:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7942:1",
    "name": "getLinesBetweenRangeEndAndRangeStart",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7949:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7948:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7949:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7948:1",
    "name": "getLinesBetweenRangeEndPositions",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7959:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7958:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7959:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7958:1",
    "name": "positionsAreOnSameLine",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7971:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7968:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7971:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7968:1",
    "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7977:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7975:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7977:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7975:1",
    "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**watch.ts:249:60**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/watch.ts:233:1",
    "name": "getFilesInErrorForSummary",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


### Every src call and its argument type

The stock TypeScript 6.0.3 checker loaded all 604 source roots from the 13 src solution projects. A separate syntax scan covered 712 src source/declaration files: 604 were loaded, and the remaining library declarations have no references to either helper. There are 93 resolved calls to the two scanner functions. The 61 same-named method calls have different resolved signatures and are listed separately in excluded-or-unresolved-calls.json. No helper reference remains outside the checked program.

External dependencies came from a separate lockfile installation, with stock module resolution falling back to that installation. No node_modules directory was added to the frozen census tree. Whole-src checking uses strict source checks on the union of solution roots; composite, project-output, and isolated-declaration restrictions are disabled. Original whole-src diagnostics: 0. Full diagnostics and exact options are retained for each variant.

The latent census loads the compiler-only 79-root program. This stock call audit intentionally loads services too, including its SourceFile augmentation. Declared type means the variable/parameter/property declaration, declared return type for a call expression, or explicit/class/contextual this type. Flow type is recorded separately because a Node or optional SourceFile binding may be narrowed at the call. Exact declarations and variant types are in each calls.json.

| Call location | Callee | Argument | Original declared type | Original flow type | (b) call errors |
| --- | --- | --- | --- | --- | --- |
| src/compiler/checker.ts:52424:27 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/checker.ts:52425:25 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/emitter.ts:1444:52 | `getLineStarts` | `Debug.checkDefined(currentSourceFile)` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/compiler/emitter.ts:3961:82 | `getLineAndCharacterOfPosition` | `currentSourceFile` | `ts.SourceFile \| undefined` | `ts.SourceFile` | none |
| src/compiler/emitter.ts:3961:150 | `getLineAndCharacterOfPosition` | `currentSourceFile` | `ts.SourceFile \| undefined` | `ts.SourceFile` | none |
| src/compiler/emitter.ts:6233:66 | `getLineAndCharacterOfPosition` | `sourceMapSource` | `ts.SourceMapSource` | `ts.SourceMapSource` | TS2345 |
| src/compiler/executeCommandLine.ts:109:27 | `getLineStarts` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:671:37 | `getLineAndCharacterOfPosition` | `diagnostic.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/compiler/program.ts:713:59 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:714:57 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:715:28 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:768:59 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:2973:28 | `getLineStarts` | `file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/compiler/scanner.ts:471:43 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/compiler/scanner.ts:539:24 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/compiler/scanner.ts:549:46 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | none |
| src/compiler/tracing.ts:206:37 | `getLineAndCharacterOfPosition` | `file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/compiler/tracing.ts:207:35 | `getLineAndCharacterOfPosition` | `file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/compiler/transformers/declarations.ts:1264:40 | `getLineAndCharacterOfPosition` | `currentSourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/transformers/declarations.ts:1264:109 | `getLineAndCharacterOfPosition` | `currentSourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/transformers/jsx.ts:359:33 | `getLineAndCharacterOfPosition` | `originalFile` | `Node` | `SourceFile` | none |
| src/compiler/utilities.ts:1018:12 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/compiler/utilities.ts:1025:17 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/utilities.ts:1032:24 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/compiler/utilities.ts:1213:16 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/utilities.ts:2526:37 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/utilities.ts:2527:35 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/utilities.ts:6748:24 | `getLineStarts` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/watch.ts:249:30 | `getLineAndCharacterOfPosition` | `diagnosticForFileName.file!` | `SourceFile \| undefined` | `SourceFile` | none |
| src/harness/fourslashImpl.ts:726:41 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/server/session.ts:265:38 | `getLineAndCharacterOfPosition` | `info.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:266:36 | `getLineAndCharacterOfPosition` | `info.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:285:51 | `getLineAndCharacterOfPosition` | `diag.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:286:49 | `getLineAndCharacterOfPosition` | `diag.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:1557:57 | `getLineAndCharacterOfPosition` | `d.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:1558:55 | `getLineAndCharacterOfPosition` | `d.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/services/codefixes/disableJsDiagnostics.ts:76:34 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:1801:67 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:1801:141 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:2171:9 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:2172:15 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:2462:24 | `getLineAndCharacterOfPosition` | `this` | `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }` | `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }` | TS2345 |
| src/services/completions.ts:2511:24 | `getLineAndCharacterOfPosition` | `this` | `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }` | `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }` | TS2345 |
| src/services/completions.ts:4393:33 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:4394:33 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:4415:33 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:4416:33 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:5344:34 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:5344:108 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:5738:21 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:5738:95 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/formatting/smartIndenter.ts:162:30 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/formatting/smartIndenter.ts:163:34 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/formatting/smartIndenter.ts:402:51 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/formatting/smartIndenter.ts:441:38 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/formatting/smartIndenter.ts:454:37 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/formatting/smartIndenter.ts:471:40 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/refactors/extractSymbol.ts:252:44 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:252:125 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:253:42 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:253:123 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:281:44 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:281:125 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:282:42 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:282:123 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:117:40 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:117:121 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:118:38 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:118:119 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:330:47 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:330:102 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToFile.ts:193:32 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToFile.ts:193:116 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToFile.ts:194:30 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToFile.ts:194:106 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToNewFile.ts:60:32 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToNewFile.ts:60:116 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToNewFile.ts:61:30 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToNewFile.ts:61:106 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/services.ts:1126:16 | `getLineAndCharacterOfPosition` | `this` | `SourceFileObject` | `this` | none |
| src/services/services.ts:1130:16 | `getLineStarts` | `this` | `SourceFileObject` | `this` | none |
| src/services/services.ts:1134:50 | `getLineStarts` | `this` | `SourceFileObject` | `this` | none |
| src/services/services.ts:1325:16 | `getLineAndCharacterOfPosition` | `this` | `SourceMapSourceObject` | `this` | none |
| src/services/sourcemaps.ts:85:40 | `getLineStarts` | `file` | `SourceFileLike \| undefined` | `SourceFileLike` | TS2345 |
| src/services/sourcemaps.ts:236:54 | `getLineStarts` | `this` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/textChanges.ts:1343:24 | `getLineAndCharacterOfPosition` | `this` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/utilities.ts:913:24 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/utilities.ts:3475:39 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/utilities.ts:3476:39 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/testRunner/unittests/tsserver/completionsIncomplete.ts:193:41 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/testRunner/unittests/tsserver/completionsIncomplete.ts:232:40 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/testRunner/unittests/tsserver/completionsIncomplete.ts:233:38 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/testRunner/unittests/tsserver/completionsIncomplete.ts:262:37 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |

Declared argument types changed by the edited helper parameter itself (all other entries retain the original declared type):

````json
[
  {
    "variant": "a",
    "file": "src/compiler/scanner.ts",
    "line": 549,
    "callee": "getLineStarts",
    "original": "SourceFileLike",
    "variant_declared": "{ readonly text: string; lineMap?: readonly number[]; }"
  },
  {
    "variant": "b",
    "file": "src/compiler/scanner.ts",
    "line": 549,
    "callee": "getLineStarts",
    "original": "SourceFileLike",
    "variant_declared": "{ readonly text: string; lineMap: readonly number[] | undefined; }"
  }
]
````


### Every failing call under (b)

These are actual stock TypeScript diagnostics from the whole-src program, not a classification based only on the spelling SourceFileLike. This preserves flow narrowing and class/contextual this types. The whole diagnostic stream, including any diagnostics outside calls, is retained in b/whole-src-diagnostics.json.

**src/compiler/emitter.ts:6233:66** `getLineAndCharacterOfPosition` argument `sourceMapSource`, declared `ts.SourceMapSource`, flow `ts.SourceMapSource`.

````json
{
  "file": "src/compiler/emitter.ts",
  "line": 6233,
  "column": 96,
  "start": 268874,
  "length": 15,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/compiler/scanner.ts:471:43** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/compiler/scanner.ts",
  "line": 471,
  "column": 57,
  "start": 47328,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/compiler/scanner.ts:539:24** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/compiler/scanner.ts",
  "line": 539,
  "column": 38,
  "start": 50647,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/compiler/utilities.ts:1018:12** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/compiler/utilities.ts",
  "line": 1018,
  "column": 26,
  "start": 33322,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/compiler/utilities.ts:1032:24** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/compiler/utilities.ts",
  "line": 1032,
  "column": 38,
  "start": 33844,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/completions.ts:2462:24** `getLineAndCharacterOfPosition` argument `this`, declared `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }`, flow `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }`.

````json
{
  "file": "src/services/completions.ts",
  "line": 2462,
  "column": 54,
  "start": 97738,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type '{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is missing in type '{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": [
    {
      "file": "src/compiler/scanner.ts",
      "line": 548,
      "column": 84,
      "start": 51088,
      "length": 7,
      "code": 2728,
      "category": "Message",
      "message": "'lineMap' is declared here.",
      "related": []
    }
  ]
}
````


**src/services/completions.ts:2511:24** `getLineAndCharacterOfPosition` argument `this`, declared `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }`, flow `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }`.

````json
{
  "file": "src/services/completions.ts",
  "line": 2511,
  "column": 54,
  "start": 99585,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type '{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is missing in type '{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": [
    {
      "file": "src/compiler/scanner.ts",
      "line": 548,
      "column": 84,
      "start": 51088,
      "length": 7,
      "code": 2728,
      "category": "Message",
      "message": "'lineMap' is declared here.",
      "related": []
    }
  ]
}
````


**src/services/formatting/smartIndenter.ts:402:51** `getLineAndCharacterOfPosition` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/formatting/smartIndenter.ts",
  "line": 402,
  "column": 81,
  "start": 19913,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/formatting/smartIndenter.ts:441:38** `getLineAndCharacterOfPosition` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/formatting/smartIndenter.ts",
  "line": 441,
  "column": 68,
  "start": 21974,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/formatting/smartIndenter.ts:454:37** `getLineAndCharacterOfPosition` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/formatting/smartIndenter.ts",
  "line": 454,
  "column": 67,
  "start": 22839,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/formatting/smartIndenter.ts:471:40** `getLineAndCharacterOfPosition` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/formatting/smartIndenter.ts",
  "line": 471,
  "column": 70,
  "start": 23802,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/sourcemaps.ts:85:40** `getLineStarts` argument `file`, declared `SourceFileLike \| undefined`, flow `SourceFileLike`.

````json
{
  "file": "src/services/sourcemaps.ts",
  "line": 85,
  "column": 54,
  "start": 3217,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/sourcemaps.ts:236:54** `getLineStarts` argument `this`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/sourcemaps.ts",
  "line": 236,
  "column": 68,
  "start": 9564,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/textChanges.ts:1343:24** `getLineAndCharacterOfPosition` argument `this`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/textChanges.ts",
  "line": 1343,
  "column": 54,
  "start": 66040,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/utilities.ts:913:24** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/utilities.ts",
  "line": 913,
  "column": 38,
  "start": 31468,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


15 calls have diagnostics under (b). The corresponding whole-source and services summaries are:

````json
{
  "label": "a",
  "whole_src_roots": 604,
  "whole_src_diagnostics": 0,
  "calls": 93,
  "services_files": 253,
  "services_diagnostics": 0,
  "services_TS2386": 0
}
````


````json
{
  "label": "b",
  "whole_src_roots": 604,
  "whole_src_diagnostics": 15,
  "calls": 93,
  "services_files": 253,
  "services_diagnostics": 15,
  "services_TS2386": 0
}
````


### Emitted public declarations with stripInternal

All three public bundles were built from actual stock TypeScript declaration emit with stripInternal=true and emitDeclarationOnly=true, then the upstream declaration bundler was run on the emitted .d.ts entry. No JavaScript or backend output was produced. Emission is allowed despite semantic errors so that (b) can be compared; its source diagnostics above are not suppressed. The exact options, emit diagnostics, emitted-file hashes, public bundles, and textual diffs are retained.

| Variant | Free getLineStarts in public bundle | Only getLineAndCharacterOfPosition parameter differs | Public SourceFile has lineMap | Public SourceFile caller errors |
| --- | ---: | --- | --- | ---: |
| original | 0 | true | false | 0 |
| a | 0 | true | false | 0 |
| b | 0 | true | false | 1 |

getLineStarts is @internal, so changing its parameter contributes no free-function declaration to public typescript.d.ts. The public API comparison normalizes only getLineAndCharacterOfPosition's first parameter back to SourceFileLike, then prints and compares the complete declaration AST. Equality confirms that this parameter is the sole public declaration difference, independent of formatting. The direct diffs and emitted signatures are retained.

Public signature original:

````json
[
  {
    "name": "getLineAndCharacterOfPosition",
    "text": "function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter;"
  }
]
````


Public signature a:

````json
[
  {
    "name": "getLineAndCharacterOfPosition",
    "text": "function getLineAndCharacterOfPosition(sourceFile: {\n        readonly text: string;\n        lineMap?: readonly number[];\n    }, position: number): LineAndCharacter;"
  }
]
````


Public signature b:

````json
[
  {
    "name": "getLineAndCharacterOfPosition",
    "text": "function getLineAndCharacterOfPosition(sourceFile: {\n        readonly text: string;\n        lineMap: readonly number[] | undefined;\n    }, position: number): LineAndCharacter;"
  }
]
````


The emitted-public-API caller probe for b reports:

````json
[
  {
    "code": 2345,
    "file": "public-sourcefile-consumer.ts",
    "line": 3,
    "column": 34,
    "message": "Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is missing in type 'SourceFile' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'."
  }
]
````


The public caller probe uses the emitted ts.SourceFile type, not the internal source interface. Its lineMap property can disappear under stripInternal even though the internal SourceFile member is required. This distinction is separate from the internal src caller audit.

Follow-up evidence is under continuation/. Original first-pass raw files remain under raw/; longer main retries are under continuation/baseline/raw/, and variant streams are under continuation/a/raw/ and continuation/b/raw/. All started measurement children are waited for, with process-group termination at their bound. The production compiler and main adaptations are unchanged; the only TypeScript source edits are in the two scratch variants.
