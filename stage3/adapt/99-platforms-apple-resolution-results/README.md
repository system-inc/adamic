# resolution results: the invalidation slot and the wrappers program.ts makes

Type-only, refusal family 5 (#pncjpqy, Compiler's split to Apple). Accepted by Compiler under the delegation. 8 sites stop today on "optional property P in T absent from structural source S, which can hide fields" (rulings.md:71). Each edit states a truthful, owned contract, certified as adaptation 72 is: the compiled program judges every relation and write at its own site, so a view the contract doesn't cover is still refused there, not trusted.

The evidence for each candidate, its construction trace and its measured stock result, is Correctness's scout, typescript/family5-scout 464502ff, `stage3/typescript/family5-scout/REPORT.md`.

## C06 (G14, G23, G24, G25, 5 sites, three slots)

Resolution wrappers come from the module resolver and host result constructors and cache maps, and `resolutionCache.ts` writes `isInvalidated = true` on reused resolution objects (adapted resolutionCache.ts:1482). The named-write audit finds that Boolean write and no conflicting declaration or value, so `never` would be false: both public result interfaces gain an `@internal` Boolean slot.

- `src/compiler/types.ts`, interface `ResolvedModuleWithFailedLookupLocations`: adds `/** @internal */ isInvalidated?: boolean;`.
- `src/compiler/types.ts`, interface `ResolvedTypeReferenceDirectiveWithFailedLookupLocations`: adds `/** @internal */ isInvalidated?: boolean;`.
- `src/compiler/types.ts`, interface `ResolvedModuleWithFailedLookupLocations`: adds `/** @internal */ files?: Set<Path>;`.
- `src/compiler/types.ts`, interface `ResolvedTypeReferenceDirectiveWithFailedLookupLocations`: adds `/** @internal */ files?: Set<Path>;`.
- `src/compiler/types.ts`, interface `ResolvedTypeReferenceDirectiveWithFailedLookupLocations`: adds `/** @internal */ alternateResult?: string;`.

The rule names only the first optional property the source lacks, so declaring `isInvalidated` alone moved every C06 stop to `files` (J-v47vk21 at 2f61ad06: 5 stops at resolutionCache.ts came back naming `files`). The view, resolutionCache.ts's `ResolutionWithFailedLookupLocations`, holds five optional slots: `failedLookupLocations` and `affectingLocations`, which both results already declare, and `isInvalidated`, `files` and `alternateResult`. `files` is the cache's own `Set<Path>` of the files using a resolution, written on the resolver's result objects (resolutionCache.ts:1109 and :1386), and copied by the readonly cache's spread in moduleNameResolver.ts, so both results gain it as `@internal`. `alternateResult` is already declared on the module result. On the type reference result it is declared as the view declares it, `string`, which makes no claim that a value is absent: `CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations` extends both, so its two declarations must be identical, and `never` there is TS2320 (J-xph0dxn at 3f5526c9).

## C15 (G13, G22, 3 sites)

program.ts makes new outer resolution wrappers in two `Array.map` callbacks (adapted program.ts:1615-1628 and :1653-1668). Their only fields are `resolvedModule` or `resolvedTypeReferenceDirective`, any spread is confined to the nested module object, and no `failedLookupLocations` is copied into the new wrapper. Explicit `map` type arguments name the public result interfaces at their fresh construction, with no change to the JavaScript and none to Array.

- `src/compiler/program.ts`, function `createProgram`: `).map(resolved =>` becomes `).map<ResolvedModuleWithFailedLookupLocations>(resolved =>` (1 time).
- `src/compiler/program.ts`, function `createProgram`: `).map(resolvedTypeReferenceDirective => ({ resolvedTypeReferenceDirective }));` becomes `).map<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>(resolvedTypeReferenceDirective => ({ resolvedTypeReferenceDirective }));` (1 time).

## How it is held

- `adapt.cjs` finds every owner through the parsed source (an interface by name, a function with a body by name), never by a coordinate, and refuses before any write a missing, duplicate or drifted owner, or a set half made. A second run writes nothing. CRLF is kept.
- `contract.cjs TREE` checks the edits under LF and CRLF, the zero-edit replay with mtime kept, and the refusals: each owner renamed, each replaced text drifted, and the set half made.
- `check.cjs BEFORE AFTER` runs on two complete applies, one without this directory. It checks three things: only the named files differ under src, and undoing these edits gives the before files; stock 6.0.3 reports the same diagnostics over the 79 src/compiler roots; and each changed file transpiles to byte-identical JavaScript at ES2020 and ESNext.
- No published API changes: every member is `@internal` or sits on an internal interface, and every annotation is on a private function.

## Why its name sorts after 99-correctness-binding-views

stage3/apply.py applies adaptations in name order. Correctness's 99-correctness-binding-views pins program.ts's `createProgram` whole, as 95-checker-axolotl-h's diagnostic owner leaves it, and refuses any other spelling as owner drift. C15's two type arguments sit inside `createProgram`, so this directory applies after binding-views, and its name says so: 99-platforms-apple-resolution-results, where the other five family 5 directories are 99-apple-*. Applied before binding-views, the whole tree refuses at its pin (J-cca6pkx at ec855744).
