# JSDocArray scout · #3qh3m28

Checkpoint. Adamic main is pinned at `42edc2ed5dbe0c31adcf5c43db507283b5ae3199`. TypeScript is v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.

## 1. Census

Tool builds and adapted tree preparation are in progress.

## 2. References and ownership

The reference list comes from TypeScript 6.0.3's `LanguageService.findReferences` over the upstream `src/` tree, seeded at the declarations in `src/compiler/types.ts`. Neither `JSDocArray` nor `jsDocCache` appears in `tests/baselines/reference/api/typescript.d.ts`.

Explicit copies occur in `compiler/factory/nodeFactory.ts:6338` (`cloneIdentifier`), `compiler/transformers/declarations.ts:830` (`preserveJsDoc`) and `services/textChanges.ts:1203` (`updateJSDocHost`). The generic clone in `nodeFactory.ts` can copy own fields, including `jsDoc`. Oracle counts are pending.

## 3. Roads

The cache is keyed by array identity today. An array-keyed map preserves that identity across the verified copy sites. A node field needs the oracle and a check of the clone and invalidation paths. The fixed-layout construction path currently starts with an array literal cast. The parser's `mapDefined` result needs an allocation or conversion path with the metadata slot reserved on the resulting array.
