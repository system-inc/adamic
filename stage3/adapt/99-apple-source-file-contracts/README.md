# SourceFile's skipTrivia and the module specifier Pick

Type-only, refusal family 5 (#pncjpqy, Compiler's split to Apple). Accepted by Compiler under the delegation. 3 sites stop today on "optional property P in T absent from structural source S, which can hide fields" (rulings.md:71). Each edit states a truthful, owned contract, certified as adaptation 72 is: the compiled program judges every relation and write at its own site, so a view the contract doesn't cover is still refused there, not trusted.

The evidence for each candidate, its construction trace and its measured stock result, is Correctness's scout, typescript/family5-scout 464502ff, `stage3/typescript/family5-scout/REPORT.md`.

## C07 (G35, 1 site)

SourceFile constructors and services' SourceFileObject omit `skipTrivia`. The distinct SourceMapSource constructors (utilities.ts:8546-8550, services/services.ts:1321) write only the matching `(pos: number) => number` callback, and no SourceFile constructor or writer supplies another type. A compatible optional internal callback on SourceFile keeps a legitimate SourceFile and source-map intersection, rather than claiming `never` for every SourceFile.

- `src/compiler/types.ts`, interface `SourceFile`: adds `/** @internal */ skipTrivia?: (pos: number) => number;`.

## C12b (G16, 2 sites)

`getModuleSpecifierPreferences` (moduleSpecifiers.ts:186) takes a Pick that erases `packageJsonScope`. Its values are SourceFile or FutureSourceFile metadata, types.ts:4305 already gives FutureSourceFile the compatible PackageJsonInfo scope, and resolver construction supplies that same cache object or omits it. Only this parameter's Pick gains the key; both readers (:200 and :243) read this parameter.

- `src/compiler/moduleSpecifiers.ts`, function `getModuleSpecifierPreferences`: `importingSourceFile: Readonly<Pick<SourceFile, "fileName" | "impliedNodeFormat">>,` becomes `importingSourceFile: Readonly<Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">>,` (1 time). The `Readonly` is adaptation 86's (#38kdd4d), which applies first and makes this parameter a readonly view; the Pick inside it gains the key, and 86's own getDefaultResolutionModeForFile parameter already names all three.

## How it is held

- `adapt.cjs` finds every owner through the parsed source (an interface by name, a function with a body by name), never by a coordinate, and refuses before any write a missing, duplicate or drifted owner, or a set half made. A second run writes nothing. CRLF is kept.
- `contract.cjs TREE` checks the edits under LF and CRLF, the zero-edit replay with mtime kept, and the refusals: each owner renamed, each replaced text drifted, and the set half made.
- `check.cjs BEFORE AFTER` runs on two complete applies, one without this directory. It checks three things: only the named files differ under src, and undoing these edits gives the before files; stock 6.0.3 reports the same diagnostics over the 79 src/compiler roots; and each changed file transpiles to byte-identical JavaScript at ES2020 and ESNext.
- No published API changes: every member is `@internal` or sits on an internal interface, and every annotation is on a private function.
