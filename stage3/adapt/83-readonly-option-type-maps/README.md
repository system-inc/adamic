# Readonly option type maps, adaptation 83

Adamic, task #w743639 (its B). Type-only: the emitted JavaScript does not change.

## The lie

A custom command-line option's `type` is declared `Map<string, number | string>`.
`libMap`, a `Map<string, string>`, is stored there for `lib`'s list element
(commandLineParser.ts:710), which makes it a view that could write a number where
a string is read, and Adamic refuses it. Every option type map is built once at
module load and only read after: `get`, `has`, `forEach`, `keys` and
`instanceof Map`.

## The edit

`adapt.cjs` puts `Readonly` before four declarations, each found by its owner
and its text through the stock AST:

- `types.ts`, `CommandLineOptionBase.type`: the union's `Map<string, number | string>` member;
- `types.ts`, `CommandLineOptionOfCustomType.type`: `Map<string, number | string>`;
- `commandLineParser.ts`, `getCustomTypeMapOfCommandLineOption`'s return: its `Map<string, string | number>` member;
- `commandLineParser.ts`, `getNameOfCompilerOptionValue`'s `customTypeMap` parameter.

The last two carry the readonly map out of the option to its only readers.
`generateTSConfig`'s `formatSingleValue` keeps its `Map` parameter, since its
caller passes `option.type instanceof Map ? option.type : undefined`, already a
`Map`, and 99-correctness-axolotl-h pins that signature's text. All four
declarations are `@internal`, so the public API doesn't move.

## Certificate

The compiled program's gate, as for adaptation 72. A write into an option's type
map inside `src/compiler` is refused at its site, and stock tsc over
`src/services` and `src/server` finds no new error.

## Measured

From J-4t94cm2 (sealed, scratch at main 0578cbd7 holding this and four sibling heads of #w743639), the tree whole against the same tree with this directory dropped:

- The meter's latent total: 3,698 without it, 3,689 with it.
- Gone, nine Refused sites: lib's map at commandLineParser.ts:710, and every enum-valued option map, the same lie: 292 (WatchFileKind), 306 (WatchDirectoryKind), 318 (PollingWatchKind), 742 (JsxEmit), 839 (ImportsNotUsedAsValues), 1099 (ModuleResolutionKind), 1428 (NewLineKind), 1673 (ModuleDetectionKind).
- Reworded, one NotYet: commandLineParser.ts:742's JsxEmit map, a boxed-union representation stop, now names ReadonlyMap where it named Map. It's the same stop, and not a write.
- Sources: types.ts and commandLineParser.ts, four lines, each `Readonly` before `Map`.
- Stock tsc over `src/compiler`, `src/services` and `src/server`: 0 new errors.
