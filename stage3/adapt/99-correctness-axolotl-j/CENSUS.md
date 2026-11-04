# Historical 99j-2 fourteen-site census delta (derived)

This is derived from the two existing sequential main and fifteen-site census runs, not a newly run fourteen-site census. Both used first-error mode and no-output assertions over 79 roots. Restore the builder primitive cast and the main getBuildInfo unit observations, retaining the other fourteen removals. The unchanged builder helper is byte-identical to main; every other edited file is byte-identical to the measured fifteen-site tree. The two observed getBuildInfo lowering stops already agree.

Compiler measured: `5d521b9076fc432852013274416c1fdbab72581b`. Prior branch base: `95f07db0bb7ede61e91b2dfe2e0582f51c1e28b0`. Current 99j-4 base: `b434890646453e5fc0895628169f10f36f253cd9` plus 99j. This census was not rerun for 99j-4. Full apply, replay, the oracle, shard check and Go lane fixtures pass on that base; see PROOF.md. These counts and unit stops are historical prior-compiler observations and derivations; they are not measurements on the current base.

| Kind | Main measured | Fourteen-site derived | Delta |
| --- | ---: | ---: | ---: |
| NotYet | 560 | 560 | +0 |
| Refused | 4294 | 4280 | -14 |
| SkippedDependency | 0 | 0 | +0 |
| error | 0 | 0 | +0 |
| panic | 0 | 0 | +0 |

Exactly fourteen removed sites, zero new sites, three left casts. Raw input SHA-256 hashes and derivation checks are in census-delta.json. Known scratch prefixes and mapped coordinates are the only message normalization; moved AST/statement identities distinguish shifted sites from additions.

| Row | Owner | Choice | Unit | Stops now |
| --- | --- | --- | --- | --- |
| 1 | toReusableCompilerOptionValue | left | getBuildInfo | Refused src/compiler/core.ts:1337:1: overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] |
| 2 | createBuilderProgramUsingIncrementalBuildInfo | shipped | createBuilderProgramUsingIncrementalBuildInfo | NotYet src/compiler/builder.ts:2380:13: new Map from a generator's pairs |
| 3 | compareTypes | shipped | createTypeChecker | NotYet src/compiler/checker.ts:2968:55: a value of type ParameterPropertyDeclaration |
| 4 | compareTypes | shipped | createTypeChecker | NotYet src/compiler/checker.ts:2968:55: a value of type ParameterPropertyDeclaration |
| 5 | compareTypes | shipped | createTypeChecker | NotYet src/compiler/checker.ts:2968:55: a value of type ParameterPropertyDeclaration |
| 6 | compareTypes | shipped | createTypeChecker | NotYet src/compiler/checker.ts:2968:55: a value of type ParameterPropertyDeclaration |
| 7 | serializeOptionBaseObject | shipped | serializeOptionBaseObject | NotYet src/compiler/commandLineParser.ts:2800:20: a Map of CompilerOptionsValue |
| 8 | serializeOptionBaseObject | left | serializeOptionBaseObject | NotYet src/compiler/commandLineParser.ts:2800:20: a Map of CompilerOptionsValue |
| 9 | convertToOptionValueWithAbsolutePaths | shipped | convertToOptionValueWithAbsolutePaths | NotYet src/compiler/commandLineParser.ts:3056:28: a value of type any |
| 10 | getConfigFileSpecs | shipped | parseJsonConfigFileContentWorker | deferred_generic_instantiation; instantiated=True |
| 11 | handleOptionConfigDirTemplateSubstitution | left | handleOptionConfigDirTemplateSubstitution | NotYet src/compiler/commandLineParser.ts:3289:17: a case whose type differs from the switch's |
| 12 | convertJsonOption | shipped | convertJsonOption | deferred_generic_instantiation; instantiated=False |
| 13 | resolveTypeReferenceDirectiveNamesReusingOldState | shipped | createProgram | NotYet src/compiler/program.ts:3268:14: a function returning SortedReadonlyArray<Diagnostic> |
| 14 | invokeCallbacks | shipped | createDirectoryWatcherSupportingRecursive | Refused src/compiler/sys.ts:600:5: a method in object destructuring |
| 15 | writeOperation | shipped | transformGenerators | NotYet src/compiler/transformers/generators.ts:1387:43: a value of type InitializedVariableDeclaration |
| 16 | writeOperation | shipped | transformGenerators | NotYet src/compiler/transformers/generators.ts:1387:43: a value of type InitializedVariableDeclaration |
| 17 | writeOperation | shipped | transformGenerators | NotYet src/compiler/transformers/generators.ts:1387:43: a value of type InitializedVariableDeclaration |

The three left sites still have their primitive cast refusals in the refusal scan. Unit lowering stops above are separate observations; deferred generic units have no direct lowering attempt. These results do not claim successful native compilation.

## Every new site

None.
