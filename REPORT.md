# Task #dt55mtb — ownership and baseline

Fetched main at `df9d8ee60` and `apple/write-hides-field-4` at `71f6d8b1`. The fix's parent is `84657fa5`. The tracked Stage 3 trees at main and the fix are identical. One baseline `stage3/apply.sh --build-product` completed on the pinned TypeScript 6.0.3 source (`050880ce59e30b356b686bd3144efe24f875ebc8`). Stock TypeScript 6.0.3's AST identified the owners at the reported coordinates in that output.

| Location | Owner | Expression |
| --- | --- | --- |
| checker.ts:16636:13 | `createCanonicalSignature` | `map(signature.typeParameters, …)` |
| checker.ts:16975:70 | `getTypeArguments` | `concatenate(type.target.outerTypeParameters, getEffectiveTypeArguments(…))` |
| checker.ts:17031:47 | `getTypeFromClassOrInterfaceReference` | `concatenate(type.outerTypeParameters, fillMissingTypeArguments(…))` |
| checker.ts:17098:38 | `getTypeFromTypeAliasReference` | `aliasTypeArguments = getTypeArgumentsForAliasSymbol(newAliasSymbol)` |
| checker.ts:19948:37 | `getInferTypeParameters` | `result = append(result, getDeclaredTypeOfSymbol(symbol))` |
| checker.ts:19979:17 | `getTypeFromConditionalTypeNode` | `aliasTypeArguments` in the `ConditionalRoot` initializer |
| factory/nodeFactory.ts:7490:71 and :93 | `mergeEmitNode` | Both arguments to `mergeTokenSourceMapRanges` |

## Ownership stop

`99-correctness-axolotl-r/owners.json` pins the whole `getTypeArguments` function before and after its edit. `99-readonly-inferred-bindings/adapt.cjs` edits the `typeArguments` binding in that same owner and deliberately sorts after that pin. `95-checker-axolotl-b/sites.json` also hashes the whole of `getTypeArguments` and `getTypeFromClassOrInterfaceReference` as dependency proofs. `71-writable-views/tuples.json` edits the signature of `getTypeArguments`. The two concatenate calls accept real compiler type arguments (both second-argument producers return `Type[]`), so typing those values as `TypeParameter` would be incorrect. The supplied `append<TypeParameter>` edit applies to the single append at 19948. Its existing symbol flag guard is `SymbolFlags.TypeParameter`.

The factory helper is local with exactly one caller in the adapted source. Both arguments come from `EmitNode.tokenSourceMapRanges`, declared as `(SourceMapRange | undefined)[]` when present. The helper copies entries into the destination and returns the destination array. Adaptation 30 records two declined indexed-read sites in this helper.

No adaptation or lane file was changed. The changed-tree emit, stock diagnostics, public declaration comparison, changed and mutant scans, second whole apply, adapted-tree product test, and whole lane tests were not run.
