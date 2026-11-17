# Readonly resolved file name

Adamic, task #w743639 (its F). Type-only: the emitted JavaScript does not change.

## The lie

`ResolutionWithResolvedFileName` (program.ts, `@internal`) declares
`resolvedFileName: string | undefined` writable. `getResolvedModuleFromResolution`
returns a `ResolvedModuleFull`, whose `resolvedFileName` is a `string`. Used as
`resolveNamesReusingOldState`'s `getResolved` (program.ts:2185), it is a function
whose result could have `undefined` written into a slot its own type reads as
defined, and Adamic refuses it. `resolveNamesReusingOldState` and its readers only
read `resolvedFileName` off a resolution.

## The edit, and why it runs after 98-axolotl-d

`adapt.cjs` inserts `readonly ` before `resolvedFileName` and touches nothing
else on the interface. The directory sorts after `98-axolotl-d`, because
98-axolotl-d edits this interface's `packageId`, and the edit pins the interface
as 98-axolotl-d leaves it: `resolvedFileName: string | undefined;` and
`packageId?: PackageId | undefined;`. Ruled by Adamic on #w743639 (road 1).

**`packageId` stays writable, on purpose.** 98-axolotl-d's member row sets it
`readonly: false` while widening it to `?: PackageId | undefined`. I found no
direct `.packageId =` write in upstream's src, so the writable slot is 98-axolotl-d's
statement, and this adaptation leaves it as it is. Don't make it readonly here
without 98-axolotl-d's owner.

## Certificate

The compiled program's gate, as for adaptation 72.

## Measured

From J-4t94cm2 (sealed, scratch at main 0578cbd7 holding this and four sibling heads of #w743639), the tree whole against the same tree with this directory dropped:

- The meter's latent total: 3,690 without it, 3,689 with it.
- Gone: program.ts:2185, the one site. Nothing new.
- Sources: program.ts alone, one line.
- Stock tsc over `src/compiler`, `src/services` and `src/server`: 0 new errors.
