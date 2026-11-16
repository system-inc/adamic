# Readonly diagnostic slots, adaptation 72

Adamic, task #m2wxyaz. This is type-only: the emitted JavaScript does not change.
It is its own adaptation, not a wave of adaptation 70. Adaptation 70's contract is
an audited no-write proof per owner, and this edit does not meet it (see the
ledger below). Adamic ruled that it is certified another way.

## The edit

`adapt.cjs` inserts `readonly ` before three members of
`DiagnosticRelatedInformation` in `src/compiler/types.ts`:

```ts
export interface DiagnosticRelatedInformation {
    category: DiagnosticCategory;
    code: number;
    readonly file: SourceFile | undefined;
    readonly start: number | undefined;
    readonly length: number | undefined;
    messageText: string | DiagnosticMessageChain;
}
```

It parses types.ts with stock TypeScript 6.0.3 and refuses before any write
unless it finds exactly one `DiagnosticRelatedInformation` with exactly the
property signatures file, start and length, each with upstream's type and no
question mark. A slot already readonly keeps its bytes, so a second run changes
nothing. No source coordinate is a rewrite address.

`Diagnostic` extends `DiagnosticRelatedInformation`, so its three slots become
readonly with it. `DiagnosticWithLocation` and
`DiagnosticWithDetachedLocation` redeclare file, start and length with
narrower types and stay writable. A `DiagnosticWithLocation` seen as a
`Diagnostic` or a `DiagnosticRelatedInformation` is therefore a read-only view,
and no longer one that can write `SourceFile | undefined` where `SourceFile` is
read. That removes refusal family 4.

The directory sorts after 71-writable-views and before 75-optional-widening, so
every later adaptation and every consumer reads apply's tree with the three
slots readonly.

## Certificate

Certified by the compiled program's gate, not by adaptation 70's audit.

Adaptation 70's audit refuses this owner: it finds 32 escapes of
`DiagnosticRelatedInformation.file` into other declarations or erased views,
listed below. This adaptation does not claim those escapes are safe. What
certifies it is that Adamic compiles `src/compiler` and judges every view at its
relation site by its own rules, so a write through a wider view inside the
compiled program is refused there, not trusted.

## Public consequence

`DiagnosticRelatedInformation` is public, so the published `typescript.d.ts`
gains `readonly` on the three members. Code that assigns file, start or length
on a Diagnostic or a DiagnosticRelatedInformation stops type-checking (TS2540);
code that reads them is unaffected; a DiagnosticWithLocation stays writable
until its own slice.

`api-sanctions.json` names exactly those three members, each replaced whole
(`"replace": "member"`), for `stage3/lane/generate-api.cjs`. The regenerated
`stage3/lane/sanctioned-api.json` gains three normalized changes,
`member:["ts","DiagnosticRelatedInformation","file"|"start"|"length"]#0`, each
the original property signature with `ReadonlyKeyword` added and nothing else.
`shard-plan.json`'s runtime hash moves with the sanction, and
`stage3/lane/test_check.py` counts 276 sanctioned API changes, three more than
273.

## Measured, from J-4m3dv80

These numbers are from the J-4m3dv80 run, made at scratch commit `c097b73b`,
where the same edit ran as `71-scratch-diagnostic-slots`, after
70-readonly-views and before 71-writable-views. 71-writable-views edits none of
the three members.

| Measurement | Before | After |
|---|---:|---:|
| TS2540 new in after, over all 747 upstream src files | | 0 |
| The meter's latent total | 4,138 | 3,940 |
| Refusal family 4 | 192 | 0 |
| Lane units failing, whole lane on the after tree | | 1 |

The one failing lane unit was `stage3-lane-byte-comparisons`, with
"unsanctioned API declaration" for the three members, since the scratch carried
no sanction. `api-sanctions.json` and the regenerated `sanctioned-api.json`
answer it.

## Owed to the subtype slice

Three refusals appear with the edit, each "whose readonly field file becomes
writable". They are family 3, not family 4, and Adamic accepts them as owed to a
later subtype slice:

- `Diagnostic` seen as `DiagnosticWithLocation`;
- `BuilderProgram` seen as `Program | undefined`;
- `T` seen as `Program | undefined`.

## Ledger: adaptation 70's 32 escapes

These are the escapes of `DiagnosticRelatedInformation.file` that adaptation
70's audit reported in J-2wwt4wb, by where they live and the declaration or
view they escape into. Paths are under `src/`; coordinates are on the tree
apply hands adaptation 70.

### Inside src/compiler, 17 escapes

These are inside the compiled program. They are judged by Adamic's relation-site
rules when compiled.

- `compiler/utilities.ts:2486:16` into `DiagnosticMessageChain`
- `compiler/utilities.ts:8641:20` into `DiagnosticWithDetachedLocation`
- `compiler/commandLineParser.ts:2291:116` into `DiagnosticWithLocation`
- `compiler/commandLineParser.ts:2301:16` and `2303:33` into `string | Diagnostic`
- `compiler/checker.ts:3617:37`, `3620:37`, `3623:37`, `3628:41`, `35099:33` and
  `35107:33` into `any`
- `compiler/program.ts:667:52` and `788:62` into `{ category: DiagnosticCategory }`
- `compiler/builder.ts:1523:16` into `ReusableDiagnosticRelatedInformation`
- `compiler/tsbuildPublic.ts:594:72` into `void`, and `598:51` into
  `ConfigFileCacheEntry`, twice

### Outside the compiled program, 15 escapes

These are in services, server, harness and testRunner. They are outside the
compiled program, and adaptation 70's audit does not certify them. Only stock
tsc checks them, and it accepts these views.

- `services/refactors/extractSymbol.ts:1869:33` into `any`
- `services/codeFixProvider.ts:122:16` into `DiagnosticWithLocation`
- `services/codefixes/addMissingAsync.ts:52:75` and `66:59` into
  `TextSpan | undefined`
- `server/session.ts:246:42`, `258:46`, `269:42`, `288:45`, `1553:46` and
  `1584:50` into `{ category: DiagnosticCategory }`
- `harness/harnessUtils.ts:163:45`, `harness/harnessIO.ts:529:62` and
  `harness/fourslashImpl.ts:230:45` into `{ category: DiagnosticCategory }`
- `testRunner/unittests/diagnosticCollection.ts:99:26` and `117:26` into `any`

## Premises a reader can re-check

Readonly is a checker fact. Two roads could write a slot without the checker
seeing an assignment to it, and both are closed today by a stop in front of
them, not by this adaptation:

- `Object.assign`. A NotYet stands in front of it today. Lifting it must judge
  readonly on the target's properties; that is filed as #w9ysnpk.
- A generic key write `t[k] = v` with `K extends keyof T`. It is NotYet today.
  Whoever lifts it judges readonly per property of K's constraint, under the
  ruling on writes through a key that isn't a literal (docs/rulings.md:290 on
  this tree; Adamic cited it as rulings.md:284).

If either stop is lifted without that judgment, this certificate no longer
holds for the three slots.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation. The landing job regenerates
the sanction and the shard plan's runtime, runs the byte-comparisons lane unit
on the after tree, measures family 4 and the latent total with this directory
present and dropped, and runs `TestProduct_Stage3AdaptedTree`.
