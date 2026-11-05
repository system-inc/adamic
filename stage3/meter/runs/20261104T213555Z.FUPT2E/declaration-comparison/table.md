# Selected latent declarations: main tip versus main plus S1c

All 23 completed on both sides. Timed unit totals: main 385.981s; main plus S1c 380.828s; merge/main 0.987x. Individual ratios range from 0.940x to 1.041x. No substantial merge slowdown was observed for these declarations; the same expensive attempts are present on main.

Main tip: `161b02663a129f2f7da97fc08fc925b120302612`. Merge: `b5733921511af8146a299502b31407dec3a1ab74`. Same box: `dc0369301a89`. Main ran first, then merge, without overlap. Each process had a 600-second cap and exited 0; process times were 425.665s and 420.769s.

Selection is the three named declarations plus ranks 4 through 23 among the slowest finished attempts in the prior 1,383-attempt replay. Rows retain that ranking; execution uses source order. All declarations below are in `src/compiler/utilitiesPublic.ts`. Ratio is merge seconds divided by main seconds.

| Declaration | Line | Main seconds | Merge seconds | Ratio |
|---|---:|---:|---:|---:|
| `isTypeOnlyExportDeclaration` | 1550 | 21.400 | 20.120 | 0.940x |
| `isTypeOnlyImportDeclaration` | 1536 | 20.892 | 19.948 | 0.955x |
| `isOutermostOptionalChain` | 1416 | 19.219 | 19.423 | 1.011x |
| `isImportOrExportSpecifier` | 1532 | 19.226 | 18.878 | 0.982x |
| `symbolName` | 870 | 19.586 | 18.969 | 0.969x |
| `isExpressionOfOptionalChainRoot` | 1399 | 18.818 | 18.724 | 0.995x |
| `isOptionalChainRoot` | 1390 | 19.462 | 18.961 | 0.974x |
| `isImportAttributeName` | 1574 | 18.321 | 18.795 | 1.026x |
| `getJSDocImplementsTags` | 1116 | 18.967 | 18.480 | 0.974x |
| `isNullishCoalesce` | 1422 | 18.534 | 18.804 | 1.015x |
| `getAssignedName` | 1004 | 19.550 | 18.455 | 0.944x |
| `getJSDocReturnType` | 1239 | 18.735 | 18.549 | 0.990x |
| `isCallChain` | 1376 | 19.638 | 18.591 | 0.947x |
| `isNonNullChain` | 1437 | 18.491 | 18.536 | 1.002x |
| `isNamedDeclaration` | 946 | 18.435 | 18.330 | 0.994x |
| `isConstTypeReference` | 1426 | 18.394 | 18.744 | 1.019x |
| `isParameterPropertyDeclaration` | 615 | 17.689 | 18.118 | 1.024x |
| `getTypeParameterOwner` | 603 | 13.458 | 14.012 | 1.041x |
| `getCombinedFlags` | 644 | 9.355 | 9.393 | 1.004x |
| `isElementAccessChain` | 1372 | 9.499 | 9.307 | 0.980x |
| `isGeneratedIdentifier` | 1581 | 9.201 | 9.231 | 1.003x |
| `isPropertyAccessChain` | 1368 | 9.513 | 9.333 | 0.981x |
| `getJSDocTagsWorker` | 1257 | 9.600 | 9.125 | 0.950x |

## Method and limits

Both sides use the same adapted TypeScript tree and the same timing-enabled, first-error entry driver. All 81 resolved implementation sources remain in the program. A scratch overlay selects only these 23 top-level declarations for attempts, retaining the existing refusal scan for their file and the existing lowering code and UnitTimer. The timer includes per-attempt registration, lowering and panic recovery, and excludes loading, file refusal scanning and writing time records. Cold process caches and program facts warm within each selected replay. The ordinary Load and Lower no-output guards remain enabled. No compiler or adapted TypeScript source was edited.

This selected replay does not reproduce the full entry pass preceding context. It is one observation per side and does not separate box effects from entry instrumentation or baseline lowering costs. Main tip contains compiler changes after the merge parent, so the comparison does not isolate S1c alone. It does show that the observed 9 to 20 second costs are already present on main; no tested declaration shows a large merge regression.

Raw findings, unit timings, selected-unit manifest, prior ranking stream, build and run logs, reproduction scripts, binary SHA-256 hashes and adapted-input hashes are retained beside this table. `table.json` preserves full precision and validation results. The original incomplete run record remains unchanged.
