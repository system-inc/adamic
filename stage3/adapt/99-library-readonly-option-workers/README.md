# Readonly option workers

Adamic, task #6c4swvt (piece 4). Type-only: the emitted JavaScript does not change.

## The lie

`serializeOptionBaseObject(options: OptionsBase, ...)` in commandLineParser.ts (`@internal`) takes its options as a
writable record. The WatchOptions that `serializeWatchOptions` hands it (2792) is then a view whose index could write
a TsConfigSourceFile where WatchOptions' own index reads a CompilerOptionsValue, and Adamic refuses it as a widening.
The worker never writes through its options. Correctness traced the receiver on typescript/options-base-evidence
5a91be06: it enumerates its options, reads them, and writes a fresh map.

## The edit

`adapt.cjs` retypes the worker's options parameter as `Readonly<OptionsBase>`. It finds the worker by name at the
file's top level and refuses any signature that is neither upstream's nor its own. A signature already adapted keeps
its bytes, so a second run changes nothing. Reading a record through a readonly view is piece 4's lowering
(library/readonly-record-views).

It sorts after `99-correctness-axolotl-j`, which pins serializeOptionBaseObject's whole owner by its hash, and after
`81-readonly-option-values`, whose casts sit inside the worker. Both run first on upstream's signature, and this edits
only the signature.

### What waits

`serializeCompilerOptions` (2712's inner call) and `handleOptionConfigDirTemplateSubstitution` (3275) stay as upstream
writes them. Retyping them as well (J-cdcj9md) traded their refusals for two lowering gaps, each a piece of its own:

- a record seen as its own `Readonly<>`, whose named properties become readonly (2675, 2712);
- core.ts's `assign` returning its fresh target as the readonly source's type (3313,
  `(result ??= assign({}, options))`).

## Certificate

The compiled program's gate, as for adaptation 72.

## Measured

From J-6v01r07 (sealed, on library/readonly-record-views beed737b), the tree whole against the same tree with this
directory dropped:

- The meter's latent total: 2,489 without it, 2,488 with it.
- Gone, one Refused site: commandLineParser.ts:2792:38, "a value of type WatchOptions seen as OptionsBase, which can
  write TsConfigSourceFile | CompilerOptionsValue where CompilerOptionsValue is read". Nothing new.
- Sources: commandLineParser.ts alone, one line.
- JavaScript: byte-identical, comments kept and removed.
- Stock tsc over `src/compiler`, `src/services` and `src/server`: 64 errors either way, 0 new.
