# module specifier export results state they hold no package root

Type-only, refusal family 5 (#pncjpqy, Compiler's split to Apple). Accepted by Compiler under the delegation. 1 sites stop today on "optional property P in T absent from structural source S, which can hide fields" (rulings.md:71). Each edit states a truthful, owned contract, certified as adaptation 72 is: the compiled program judges every relation and write at its own site, so a view the contract doesn't cover is still refused there, not trusted.

The evidence for each candidate, its construction trace and its measured stock result, is Correctness's scout, typescript/family5-scout 464502ff, `stage3/typescript/family5-scout/REPORT.md`.

## C16 (G30, 1 site)

`tryGetModuleNameFromExportsOrImports` (moduleSpecifiers.ts:1020) and `tryGetModuleNameFromExports` (:1122) return fresh objects holding exactly `moduleFileToTry`, or recursive results of the same constructors, and `tryDirectoryWithPackageJson` (:1277) spreads that known result. No arbitrary package.json object is returned as it. So `packageRootPath?: never` and `blockedByExports?: never` on both private result annotations is a claim about these constructors, never about every structurally compatible object. Correctness's constructor audit (the scout's evidence/constructors.md) is the witness for `packageRootPath`. Apple's trace is the witness for `blockedByExports`, below.

- `src/compiler/moduleSpecifiers.ts`, function `tryGetModuleNameFromExportsOrImports`: `): { moduleFileToTry: string; } | undefined {` becomes `): { moduleFileToTry: string; packageRootPath?: never; blockedByExports?: never; } | undefined {` (1 time).
- `src/compiler/moduleSpecifiers.ts`, function `tryGetModuleNameFromExports`: `): { moduleFileToTry: string; } | undefined {` becomes `): { moduleFileToTry: string; packageRootPath?: never; blockedByExports?: never; } | undefined {` (1 time).

Declaring `packageRootPath` alone moved the stop to `blockedByExports`, the next optional property of `tryDirectoryWithPackageJson`'s result that the spread source lacks (J-v47vk21 at 2f61ad06). `verbatimFromExports` is set by the spread itself. Apple's trace for `blockedByExports`, at TypeScript 050880ce: every return of `tryGetModuleNameFromExportsOrImports` is a fresh `{ moduleFileToTry }` (moduleSpecifiers.ts:1050, :1056, :1060, :1064, :1068, :1072, :1081, :1085, :1089, :1093, :1099), a recursive result of itself (:1105 through `forEach`, :1114), or `undefined` (:1099, :1119). Every return of `tryGetModuleNameFromExports` is a result of `tryGetModuleNameFromExportsOrImports` (:1134, :1137). Neither function writes any property onto a result after making it. The only result holding `blockedByExports` is made by `tryDirectoryWithPackageJson` itself (:1280), never by these two.

## How it is held

- `adapt.cjs` finds every owner through the parsed source (an interface by name, a function with a body by name), never by a coordinate, and refuses before any write a missing, duplicate or drifted owner, or a set half made. A second run writes nothing. CRLF is kept.
- `contract.cjs TREE` checks the edits under LF and CRLF, the zero-edit replay with mtime kept, and the refusals: each owner renamed, each replaced text drifted, and the set half made.
- `check.cjs BEFORE AFTER` runs on two complete applies, one without this directory. It checks three things: only the named files differ under src, and undoing these edits gives the before files; stock 6.0.3 reports the same diagnostics over the 79 src/compiler roots; and each changed file transpiles to byte-identical JavaScript at ES2020 and ESNext.
- No published API changes: every member is `@internal` or sits on an internal interface, and every annotation is on a private function.
