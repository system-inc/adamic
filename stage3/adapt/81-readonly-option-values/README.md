# Readonly option values, adaptation 81

Adamic, task #w743639 (its A). This is type-only: the emitted JavaScript does not change.

## The lie

`CompilerOptionsValue` (`src/compiler/types.ts`) holds `string[]` beside
`(string | number)[]`, both mutable. So any option value that holds a `string[]`
can be seen as `(string | number)[]` and have a number written into it, and
tsc's own types allow that write anywhere a value moves into a wider option
type. Adamic refuses each such move ("can write string | number where string is
read"). The scan at main 03c140ec found 18 of them in the compiled program.
tsc never makes that write: every option array it builds is fresh, and every
array it reads out of an option value is only read.

## The edit

`adapt.cjs` inserts `readonly ` in these places, each found through the stock
6.0.3 AST by its owner and its text, never by a source coordinate.

In `src/compiler/types.ts`, the four array members of `CompilerOptionsValue`:

```ts
export type CompilerOptionsValue = string | number | boolean | readonly (string | number)[] | readonly string[] | MapLike<string[]> | readonly PluginImport[] | readonly ProjectReference[] | null | undefined;
```

`MapLike<string[]>` (`paths`) is a record, not an array, and is left as it is.

In `src/compiler/commandLineParser.ts`, the same lie one declaration over, and
every cast that would turn a readonly option array back into a mutable one:

- `parseListTypeOption`'s declared return becomes
  `string | readonly (string | number)[] | undefined`. Its result is stored as an
  option value (`options[opt.name] = result || []`), and it is the scan's site
  1901.
- `generateTSConfig`'s local `PresetValue` becomes
  `string | number | boolean | readonly (string | number | boolean)[]`, so
  `(value ?? defaultValue) as PresetValue` is a readonly view.
  `formatValueOrArray` and `formatSingleValue` only read it: `value.map(...)`,
  and a reassignment of the local `value`.
- `value as string[]` in `serializeOptionBaseObject` becomes
  `value as readonly string[]`. The value is only mapped.
- `value as string[]` in `handleOptionConfigDirTemplateSubstitution` becomes
  `value as readonly string[]`. It is handed to
  `getSubstitutedStringArrayWithConfigDirTemplate(list: readonly string[] | undefined, ...)`,
  which writes only into its own `list.slice()`.
- Both `value as unknown[]` in `getExtendsConfigPathOrArray` become
  `value as readonly unknown[]`. They read a length and an element.

Each edit refuses unless it finds exactly what upstream writes. A member or cast
already readonly keeps its bytes, so a second run changes nothing. No cast of an
option value to a mutable array is left that writes, so none stays refused, and
no cast from a readonly type back to a writable one is made.

## Certificate

Certified as adaptation 72 is, by the compiled program's gate. Adamic compiles
`src/compiler` and judges every view at its relation site by its own rules, so a
write through a wider view of an option value inside the compiled program is
refused there, not trusted. Ruled by Adamic on #w743639.

## Public consequence

`CompilerOptionsValue` is public, so the published `typescript.d.ts` gains
`readonly` on its four array members. Code that assigns an option value to a
mutable array without a cast stops type-checking. Code that reads option values
is unaffected. `CompilerOptions`' own named members (`lib`, `rootDirs`, `types`
and the rest) keep their mutable types, since they are their own declarations.

`api-sanctions.json` holds one row: the type alias replaced whole
(`"kind": "TypeAliasDeclaration"`). `stage3/lane/generate-api.cjs` takes that kind
for the first time with this adaptation. Its owner's `type` is the alias's union,
read and replaced as a member's type is. The regenerated
`stage3/lane/sanctioned-api.json` gains one normalized change, `shard-plan.json`'s
runtime hash moves with the sanction, and `stage3/lane/test_check.py` counts 281
sanctioned API changes, one more than main's 280.

## Measured

From J-zkkms2s (sealed, azure, scratch commit with this directory and the job, never cut), against the same tree
with 81 dropped and 99j's pins as main has them:

| Measurement | 81 dropped | 81 |
|---|---:|---:|
| The meter's latent total | 3,707 | 3,689 |
| A's 18 scanned sites that stop | 18 | 1 |
| Stock tsc errors over src/compiler, beyond the missing node typings | 0 | 0 |
| Casts from a readonly array to a mutable one | 4 | 4 |

- **Gone, 19 Refused sites:** builder.ts 1470, 1485; commandLineParser.ts 1899, 1901, 2823, 2831, 2950 (two), 2994, 3012,
  3095:9, 3298, 3313, 3563, 3564, 3807, 3876; tsbuildPublic.ts 333; utilities.ts 9404. 1899 is parseListTypeOption's
  `number[]` branch, the same lie as 1901.
- **Reworded, 5 NotYet sites:** commandLineParser.ts 3095 (two), 3100, 3101 (two) print CompilerOptionsValue's members
  readonly and in another order; each is the same record-storage stop, #6c4swvt's.
- **Exposed, 1 NotYet site:** commandLineParser.ts 3563, getExtendsConfigPathOrArray's `isArray(value)` over the four
  array members, now stops at "a narrowed union member whose object tag cannot be checked with typeof". The cast's
  refusal stood in front of it. Telling the four array kinds apart is #y852c3c's element-checked view, not a lie.
- **The four casts left** are the same on both trees and none comes from an option value: parser.ts 9481,
  commandLineParser.ts 3249, checker.ts 36741, transformers/utilities.ts 590.
- **The lane:** stage3-lane-byte-comparisons passes on the 81 tree with the manifest's counts (277 sanctioned
  declarations). Two controls fail as they must, each with one error naming CompilerOptionsValue: the 81 tree judged
  against the sanction as committed before this change ("unsanctioned API declaration"), and the tree without 81
  judged against the regenerated sanction ("missing sanctioned API declaration"). TestProduct_Stage3AdaptedTree passes.
- **The harness mutant:** with `TypeAliasDeclaration` taken back out of generate-api.cjs's `sanctionKinds`,
  test_sanctioned_api_freshness is red on this directory's row (`stage3/lane/evidence/w743639-type-alias-sanction-mutant.md`).

## The casts, by site

Each is a cast of an option value to an array, found by stock tsc's checker on the edited tree (operand holds a
readonly array, target a mutable one). Each only reads, and each becomes a readonly form:

| Site (main 910f21af) | Owner | Before | After | What reads it |
|---|---|---|---|---|
| commandLineParser.ts:2823 | serializeOptionBaseObject | `value as string[]` | `value as readonly string[]` | `.map(...)` into a new array |
| commandLineParser.ts:2950 | generateTSConfig | `as PresetValue` | the same cast, PresetValue's array readonly | formatValueOrArray's `value.map`, formatSingleValue's reassignment of its own local |
| commandLineParser.ts:3297 | handleOptionConfigDirTemplateSubstitution | `value as string[]` | `value as readonly string[]` | getSubstitutedStringArrayWithConfigDirTemplate's `list: readonly string[] \| undefined`, which writes only `list.slice()` |
| commandLineParser.ts:3563 | getExtendsConfigPathOrArray | `value as unknown[]` | `value as readonly unknown[]` | `.length` |
| commandLineParser.ts:3564 | getExtendsConfigPathOrArray | `value as unknown[]` | `value as readonly unknown[]` | `[index]` |

None writes, so none stays refused, and no cast from readonly back to mutable is made.

## 99j's pins, moved (Adamic's road 1)

`99-correctness-axolotl-j` pins serializeOptionBaseObject by the sha256 of its whole text over LF line endings
(adapt.cjs, its `owners` row), and 81, sorting first, inserts one `readonly` in that owner's cast at 2823, which 99j's
README counts among its three left statements. Its two pins move to the same texts with that one word:
`ownerBeforeSha256` a5b45fea... to b86f5a2f..., `ownerAfterSha256` d48c5100... to df24e2ff.... They were computed, not
typed (J-cw06hbf): on main's tree applied up to 99j, the owner hashed to today's after pin, reversing 99j's one site
gave today's before pin, and each with `readonly` inserted gives the new one. 99j's sites and its fourteen edits don't
move. Its hand-run contract.cjs names the left statement with `readonly` too. 99j applies on the 81 tree (J-zkkms2s).

## Premises a reader can re-check

Readonly is a checker fact. Two roads could write into an option array without
the checker seeing an assignment to it. Both are closed today by a stop in front
of them, not by this adaptation, as adaptation 72 names them:

- `Object.assign`, NotYet today. Lifting it must judge readonly on the target;
  filed as #w9ysnpk.
- A generic key write `t[k] = v` with `K extends keyof T`, NotYet today.
  Whoever lifts it judges readonly per property of K's constraint, under the
  ruling on writes through a key that isn't a literal.

If either stop is lifted without that judgment, this certificate no longer holds
for option values.

## What this leaves

Five record-storage stops at `commandLineParser.ts` 3095 and 3100 to 3101 stay,
reworded with the readonly members. They are #6c4swvt's, waiting on #y852c3c's
element-checked view, which readonly doesn't give.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
