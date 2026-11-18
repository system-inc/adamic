# Annotated option spreads, adaptation 84

Adamic, task #w743639 (its K). Type-only: the emitted JavaScript does not change.

## Why

Three spreads are written with no type of their own. Each literal is made as a fixed object and then seen as
the type it's used as. For the first two that's `CompilerOptions`, a record, which Adamic refuses ("fixed
objects and records have different storage; copy explicitly"):

- `checker.ts`, `getSpecifierForModuleSymbol`:
  `const specifierCompilerOptions = isBundle ? { ...compilerOptions, baseUrl: ... } : compilerOptions;`,
  handed to `moduleSpecifiers.getModuleSpecifiers` as its `CompilerOptions` (checker.ts:8640);
- `moduleNameResolver.ts`, `nodeModuleNameResolverWorker`:
  `const diagnosticsCompilerOptions = { ...state.compilerOptions, moduleResolution: ... };`,
  handed to `getConditions` (1891) and stored as `diagnosticState`'s `compilerOptions`
  (`ModuleResolutionState.compilerOptions`, 1894);
- `moduleNameResolver.ts`, `nodeModuleNameResolverWorker`:
  `const diagnosticState = { ...state, compilerOptions: diagnosticsCompilerOptions, ..., conditions: getConditions(...), ... };`,
  handed to `tryResolve` as its `ModuleResolutionState` (1894). With the first two annotated, its storage stop
  goes, and under it Adamic refuses a widening: the literal's `conditions`, getConditions' `string[]`, seen as
  the state's `readonly string[]`, "can write readonly string[] where string[] is read" (J-wmsybfj).

tsc tells no lie here. Each value is used only as the type it's annotated with, so it is given that type where
it's made, and the literal is made as its slot holds it.

**The premise, for Language:** each literal is fresh and isn't read through any other type before its
annotation, so it has exactly one handle, the annotated one. Typed `ModuleResolutionState` where it's built,
`diagnosticState`'s `conditions` is taken as `readonly string[]` from the start, and no second view of it
reads `conditions` as `string[]`. A fresh `string[]` into a `readonly string[]` slot only narrows what can be
done to it.

## The edit

`adapt.cjs` adds `: CompilerOptions` after the first two names and `: ModuleResolutionState` after the third, found by its owner, its
name and the start of its initializer through the stock AST. A declaration
already annotated with its type keeps its bytes. Both files already import
`CompilerOptions`, and `ModuleResolutionState` is declared in moduleNameResolver.ts.

## Ruling and probe

Adamic took this shape on #w743639 once the probe showed it lowers: J-grhaqpf,
the two annotations made by hand on main's tree, cleared the three storage stops. J-wmsybfj then read the
widening under 1894, and Adamic ruled the third annotation the same move (road 1).

## Measured

From J-z9nf89w (sealed, scratch at main 0578cbd7), the tree whole against the same tree with this directory dropped:

- The meter's latent total: 3,691 without it, 3,688 with it.
- Gone: checker.ts:8640, moduleNameResolver.ts:1891 and 1894, all three storage stops, and 1894's widening under them. Nothing new.
- Sources: checker.ts and moduleNameResolver.ts, three lines, each `: <type>` after a name.
- Stock tsc over `src/compiler`, `src/services` and `src/server`: 0 new errors.
- JavaScript, from J-zp89xcf: checker.ts and moduleNameResolver.ts differ in source with and without 84 and transpile through stock tsc to byte-identical JavaScript (2,893,541 and 150,489 bytes, same sha256).
