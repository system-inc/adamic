# Remaining partition files on the latent meter

These are observations from the final fair 78-root all-code census. The older
63-finding table selected only five codes; this snapshot uses the requested
latent run 0 meter and input. Do not compare
its total to this all-code frontier. All remaining diagnostic messages are in
closure-frontier.json. Only the requested six-file pass has newly reviewed site
proofs; the following is work still needed, not permission to assert missing values.

| File | Remaining | Next review needed to reach zero |
| --- | ---: | --- |
| performanceCore.ts | 2 | Wait for compiler require-builtins/Node host bindings, as requested; no source shim. |
| builder.ts | 17 | Review erased Map/Set iterator type arguments (arrayFrom, forEachKey, mapDefinedIterator); type the map generator output tuple; review truthful undefined unions on ReusableBuilderProgramState, then remeasure all cascades and the public API boundary. |
| watch.ts | 2 | The two reviewed errors require truthful owning public optional host declarations. Those API lines are outside the exact sanctioned set, so presence assertions are unsound and the file remains declined. |
| watchPublic.ts | 12 | Review owning optional host function declarations and explicit undefined reset fields, distinguishing internal input types from public host interfaces; public API expansions need separate sanction. |
| resolutionCache.ts | 12 | Review firstDefinedIterator Path-yield contracts; review optional cache-input/reset/nonRecursive fields; prove the watcher creation return and Map retrieval paths before any required-value assertion. |
| tsbuildPublic.ts | 14 | Review map iterator yielded values; review ten optional host function assignments at their owning declarations, including variance for writeFile. Public owning contracts cannot be added to the accepted API set silently. |
| moduleNameResolver.ts | 10 | Review explicit undefined in owning PackageId/resolution/cache result contracts; remeasure assignment failures and resulting result-variable cascades before adding any assertions. Public contracts may require separate sanction. |
| sys.ts | 57 | Most findings need proven Node global/module bindings (44 TS2591, six TS2304, one TS2307). Then remeasure contextual callback typing, ModuleImportResult absence, and two unknown catches; do not assume Error-shaped throws. |
| program.ts | 13 | Review the existing typeof-substitution narrowing, iterator yield contracts and optional host declarations. Some owner types are public; two catches lack a proven Error contract. No whole-file zero claim is made. |
| commandLineParser.ts | 9 | Review Map/Set iterator contracts for keys/values and map output arrays. Review CompilerOptions/ProjectReference explicit undefined owners against the public API boundary. One caught value is unknown. |
| tracing.ts | 5 | Five reviewed declines: four Node host bindings and one unknown catch. typesPath explicit undefined is fixed; further zero requires proven host/exception contracts. |

The 15 other owned files are already zero on this meter. Four of them were
closed in this pass: builderState.ts, executeCommandLine.ts, moduleSpecifiers.ts
and watchUtilities.ts. The whole-tree count is 40 of 78 files at zero.
