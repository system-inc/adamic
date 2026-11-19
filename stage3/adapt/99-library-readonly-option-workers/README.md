# Readonly option workers

Adamic, task #6c4swvt (piece 4). Type-only: the emitted JavaScript does not change.

## The lie

Three `@internal` workers in commandLineParser.ts take their options as a writable record:

- `serializeOptionBaseObject(options: OptionsBase, ...)`;
- `handleOptionConfigDirTemplateSubstitution(options: OptionsBase | undefined, ...)`;
- `serializeCompilerOptions(options: CompilerOptions, ...)`.

A WatchOptions handed to the first two, or a record of option values handed to the third (2712's
`impliedCompilerOptions`), is then a view whose index could write a TsConfigSourceFile where the object's own index
reads a CompilerOptionsValue. Adamic refuses each one as a widening (commandLineParser.ts 2712's inner call, 2792
and 3275).

None of the three writes through its options. Correctness traced each receiver on typescript/options-base-evidence
5a91be06:

- serializeOptionBaseObject enumerates its options, reads them and writes a fresh map;
- handleOptionConfigDirTemplateSubstitution reads its options and writes only a fresh shallow copy
  (`assign({}, options)`), returning the original object when nothing is substituted;
- serializeCompilerOptions only hands its options on.

## The edit

`adapt.cjs` retypes each worker's options parameter as `Readonly<...>`. It also annotates the substitution
worker's result as `Readonly<OptionsBase> | undefined`, since that result may be the original object, so the
readonly view never comes back writable. That follows Runtime's ruling on #6c4swvt: a readonly record seen as a
writable one is refused.

The adapter finds each worker by name at the file's top level, and refuses any signature that is neither
upstream's nor its own. A signature already adapted keeps its bytes, so a second run changes nothing.

It sorts after `99-correctness-axolotl-j`, which pins serializeOptionBaseObject's whole owner by its hash, and after
`81-readonly-option-values`, whose casts sit inside these workers. Both run first, and this edits only the signatures.

Reading a record through a readonly view is piece 4's lowering (library/readonly-record-views).

## Certificate

The compiled program's gate, as for adaptation 72.

## Measured

To be filled from the proof job.
