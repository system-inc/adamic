# Global inherited-key census

Pinned TypeScript 6.0.3 original compiler code: 77 files. Stock checker: zero diagnostics.

**1,041** dynamic element reads and membership checks: **962 fixed key sets/domains excluding prototype names, 36 user input, 43 unknown**.

Of these, **145** have string-capable keys (including every helper call, even with a literal key): **66 fixed, 36 user input, 43 unknown**. The other **896** have numeric/bigint/symbol domains that cannot name any of the twelve Object.prototype members.

The fully dynamic string-key subset is **130**: **51 fixed, 36 user input, 43 unknown**. Fifteen literal-key hasProperty calls make up the difference.

Counts are distinct source operations, not executions or possible panics. A helper call and an operation inside its definition are distinct source sites. A for-in tag is an overlapping context, never an extra row in the total. Nested element accesses sharing a start column have distinct end offsets in JSON.

## Requested forms

| Form | Sites |
| --- | ---: |
| Dynamic element reads, including compounds | 993 |
| Of those, string-capable keys | 97 |
| hasProperty calls | 36 on 32 lines |
| getProperty calls | 0 |
| Direct hasOwnProperty.call checks in the core helpers and for-in bodies | 12 |
| Non-literal in tests | 0 |
| Literal in tests, excluded | 10 |
| Actual for-in loops | 27 |
| Read/check sites inside their bodies, already counted above | 57 |
| Original string-index census accesses, fully reconciled | 68 |
| Of those, counted dynamic reads | 45 |

The 68 reconciled accesses contain 22 write/delete-only sites and one literal dot-key read. All retain a three-way key classification and their counted flag in JSON. The 45 is now a subset of the global pass, not the answer.

## User-input origins

| Origin combination | Sites |
| --- | ---: |
| command line | 3 |
| package.json | 13 |
| source code | 2 |
| tsconfig | 11 |
| tsconfig + command line | 5 |
| tsconfig + package.json | 2 |

Source occurrences overlap: tsconfig 18, package.json 15, command line 8, source-code module specifiers 2. The disjoint combinations above sum to 36. A key from user input stays in that class even where subsequent syntax filtering excludes prototype names, or an own-property guard guarantees a safe hit.

## String-capable sites

| Location | Operation | Class | Input source or evidence |
| --- | --- | --- | --- |
| src/compiler/builder.ts:1449:21 | element: `name` | fixed_set_excludes_prototype | Own option name accepted by optionsNameMap and affectsBuildInfo. Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/checker.ts:13951:14 | element: `resolutionKind` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: resolvedExports, resolvedMembers |
| src/compiler/checker.ts:14000:38 | element: `resolutionKind` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: resolvedExports, resolvedMembers |
| src/compiler/checker.ts:14017:16 | element: `resolutionKind` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: resolvedExports, resolvedMembers |
| src/compiler/checker.ts:14176:16 | element: `key` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: inner, outer |
| src/compiler/checker.ts:19528:13 | element: `cache` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: simplifiedForWriting, simplifiedForReading |
| src/compiler/checker.ts:19529:20 | element: `cache` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: simplifiedForWriting, simplifiedForReading |
| src/compiler/checker.ts:19529:68 | element: `cache` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: simplifiedForWriting, simplifiedForReading |
| src/compiler/checker.ts:45839:34 | element: `getIterationTypesKeyFromIterationTypeKind(typeKind)` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: yieldType, returnType, nextType |
| src/compiler/checker.ts:45900:16 | element: `cacheKey` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: iterationTypesOfGeneratorReturnType, iterationTypesOfAsyncGeneratorReturnType, iterationTypesOfIterable, iterationTypesOfIterator, iterationTypesOfAsyncIterable, iterationTypesOfAsyncIterator, iterationTypesOfIteratorResult |
| src/compiler/checker.ts:46488:34 | element: `getIterationTypesKeyFromIterationTypeKind(kind)` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: yieldType, returnType, nextType |
| src/compiler/commandLineParser.ts:1716:118 | hasProperty: `"transpileOptionValue"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: transpileOptionValue |
| src/compiler/commandLineParser.ts:2690:29 | element: `option` | fixed_set_excludes_prototype | Enumerated computedOptions, an internal literal table. Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/commandLineParser.ts:2691:34 | element: `option` | fixed_set_excludes_prototype | Enumerated computedOptions, an internal literal table. Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/commandLineParser.ts:2693:50 | element: `option` | fixed_set_excludes_prototype | Enumerated computedOptions, an internal literal table. Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/commandLineParser.ts:2707:25 | element: `option` | fixed_set_excludes_prototype | optionDependsOn has one entry from the internal computedOptions loop (commandLineParser.ts:2689); recursion uses the dependency arrays of _computedOptions (utilities.ts:9041-9286). These names exclude all prototype members. |
| src/compiler/commandLineParser.ts:2789:13 | hasProperty: `name` | user_input | tsconfig + command line: Own option name from options/opts; these objects are populated by config and command-line parsing. The own-key guard does not exclude an own prototype-named key. |
| src/compiler/commandLineParser.ts:2795:27 | element: `name` | user_input | tsconfig + command line: Own names on options from config or CLI. |
| src/compiler/commandLineParser.ts:2901:42 | element: `allSetOptions[0]` | user_input | command line: generateTSConfig options parameter is the command-line options object (commandLineParser.ts:2831-2838); allSetOptions retains its own names, and emitOption receives those names as well as presets. Public callers can also provide options. |
| src/compiler/commandLineParser.ts:2930:24 | hasProperty: `setting` | user_input | command line: generateTSConfig emitOption setting can come from the remaining command-line option names at 2901, as well as preset literals. |
| src/compiler/commandLineParser.ts:2933:24 | element: `setting` | user_input | command line: generateTSConfig options parameter is the command-line options object (commandLineParser.ts:2831-2838); allSetOptions retains its own names, and emitOption receives those names as well as presets. Public callers can also provide options. |
| src/compiler/commandLineParser.ts:2976:13 | hasProperty: `name` | user_input | tsconfig + command line: Own option name from options/opts; these objects are populated by config and command-line parsing. The own-key guard does not exclude an own prototype-named key. |
| src/compiler/commandLineParser.ts:2979:17 | element: `name` | user_input | tsconfig + command line: Own names on options; tsconfig or CLI. |
| src/compiler/commandLineParser.ts:3114:32 | hasProperty: `"extends"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: extends |
| src/compiler/commandLineParser.ts:3229:13 | hasProperty: `prop` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, include, exclude, references |
| src/compiler/commandLineParser.ts:3229:58 | element: `prop` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, include, exclude, references |
| src/compiler/commandLineParser.ts:3230:25 | element: `prop` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, include, exclude, references |
| src/compiler/commandLineParser.ts:3231:32 | element: `prop` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, include, exclude, references |
| src/compiler/commandLineParser.ts:3268:13 | element: `option.name` | fixed_set_excludes_prototype | Internal config-dir option declaration lists. Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/commandLineParser.ts:3269:27 | element: `option.name` | fixed_set_excludes_prototype | Internal config-dir option declaration lists. Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/commandLineParser.ts:3322:22 | element: `key` | user_input | tsconfig: Own keys of paths from tsconfig. |
| src/compiler/commandLineParser.ts:3323:77 | element: `key` | user_input | tsconfig: Own keys of paths from tsconfig. |
| src/compiler/commandLineParser.ts:3350:9 | hasProperty: `"references"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: references |
| src/compiler/commandLineParser.ts:3355:13 | hasProperty: `"files"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files |
| src/compiler/commandLineParser.ts:3355:43 | hasProperty: `"references"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: references |
| src/compiler/commandLineParser.ts:3464:21 | element: `propertyName` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, include, exclude |
| src/compiler/commandLineParser.ts:3465:21 | element: `propertyName` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, include, exclude |
| src/compiler/commandLineParser.ts:3466:48 | element: `propertyName` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, include, exclude |
| src/compiler/commandLineParser.ts:3503:9 | hasProperty: `"excludes"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: excludes |
| src/compiler/commandLineParser.ts:3726:10 | hasProperty: `compileOnSaveCommandLineOption.name` | fixed_set_excludes_prototype | compileOnSaveCommandLineOption.name is the literal compileOnSave in the private declaration at commandLineParser.ts:126-128. |
| src/compiler/commandLineParser.ts:3785:90 | element: `id` | user_input | tsconfig: id is enumerated from jsonOptions in convertOptionsFromJson (3782); caller convertCompilerOptionsFromJson receives the tsconfig compilerOptions object. |
| src/compiler/commandLineParser.ts:4142:68 | element: `existingPath` | user_input | tsconfig: Directory paths derived from user include patterns; normalized path restrictions need proof. |
| src/compiler/commandLineParser.ts:4155:17 | hasProperty: `path` | user_input | tsconfig: path is enumerated from wildcardDirectories built from tsconfig include/exclude patterns; complete normalization exclusions are not asserted. |
| src/compiler/commandLineParser.ts:4266:13 | hasProperty: `key` | user_input | tsconfig + command line: Own option name from options/opts; these objects are populated by config and command-line parsing. The own-key guard does not exclude an own prototype-named key. |
| src/compiler/commandLineParser.ts:4269:59 | element: `key` | fixed_set_excludes_prototype | Read occurs only after getOptionFromName(key) succeeds (commandLineParser.ts:4267-4269); the recognized internal option/alias set excludes prototype names. |
| src/compiler/core.ts:1267:12 | hasOwnProperty.call: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1279:12 | hasOwnProperty.call: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1279:44 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1290:13 | hasOwnProperty.call: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1315:13 | hasOwnProperty.call: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1316:25 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1354:17 | hasProperty: `p` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1355:24 | element: `p` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1374:13 | hasOwnProperty.call: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1375:18 | hasOwnProperty.call: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1376:35 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1376:46 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1381:13 | hasOwnProperty.call: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1382:18 | hasOwnProperty.call: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1470:27 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1481:13 | hasOwnProperty.call: `id` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1482:26 | element: `id` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1498:13 | hasOwnProperty.call: `id` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1499:35 | element: `id` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1504:13 | hasOwnProperty.call: `id` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1505:35 | element: `id` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1515:13 | hasOwnProperty.call: `id` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:1516:34 | element: `id` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:2143:18 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/core.ts:2143:26 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/debug.ts:168:36 | element: `key` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: type, shouldLog, log, getAssertionLevel, setAssertionLevel, shouldAssert, fail, failBadSyntaxKind, assert, assertEqual, assertLessThan, assertLessThanOrEqual, assertGreaterThanOrEqual, assertIsDefined, checkDefined, assertEachIsDefined, checkEachDefined, assertNever, assertEachNode, assertNode, assertNotNode, assertOptionalNode, assertOptionalToken, assertMissingNode, getFunctionName, formatSymbol, formatEnum, formatSyntaxKind, formatSnippetKind, formatScriptKind, formatNodeFlags, formatNodeCheckFlags, formatModifierFlags, formatTransformFlags, formatEmitFlags, formatSymbolFlags, formatTypeFlags, formatSignatureFlags, formatObjectFlags, formatFlowFlags, formatRelationComparisonResult, formatCheckMode, formatSignatureCheckMode, formatTypeFacts, attachFlowNodeDebugInfo, attachNodeArrayDebugInfo, enableDebugInfo, formatVariance, attachDebugPrototypeIfDebug, printControlFlowGraph, formatControlFlowGraph |
| src/compiler/debug.ts:169:49 | element: `key` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: type, shouldLog, log, getAssertionLevel, setAssertionLevel, shouldAssert, fail, failBadSyntaxKind, assert, assertEqual, assertLessThan, assertLessThanOrEqual, assertGreaterThanOrEqual, assertIsDefined, checkDefined, assertEachIsDefined, checkEachDefined, assertNever, assertEachNode, assertNode, assertNotNode, assertOptionalNode, assertOptionalToken, assertMissingNode, getFunctionName, formatSymbol, formatEnum, formatSyntaxKind, formatSnippetKind, formatScriptKind, formatNodeFlags, formatNodeCheckFlags, formatModifierFlags, formatTransformFlags, formatEmitFlags, formatSymbolFlags, formatTypeFlags, formatSignatureFlags, formatObjectFlags, formatFlowFlags, formatRelationComparisonResult, formatCheckMode, formatSignatureCheckMode, formatTypeFacts, attachFlowNodeDebugInfo, attachNodeArrayDebugInfo, enableDebugInfo, formatVariance, attachDebugPrototypeIfDebug, printControlFlowGraph, formatControlFlowGraph |
| src/compiler/debug.ts:189:56 | element: `name` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: type, shouldLog, log, getAssertionLevel, setAssertionLevel, shouldAssert, fail, failBadSyntaxKind, assert, assertEqual, assertLessThan, assertLessThanOrEqual, assertGreaterThanOrEqual, assertIsDefined, checkDefined, assertEachIsDefined, checkEachDefined, assertNever, assertEachNode, assertNode, assertNotNode, assertOptionalNode, assertOptionalToken, assertMissingNode, getFunctionName, formatSymbol, formatEnum, formatSyntaxKind, formatSnippetKind, formatScriptKind, formatNodeFlags, formatNodeCheckFlags, formatModifierFlags, formatTransformFlags, formatEmitFlags, formatSymbolFlags, formatTypeFlags, formatSignatureFlags, formatObjectFlags, formatFlowFlags, formatRelationComparisonResult, formatCheckMode, formatSignatureCheckMode, formatTypeFacts, attachFlowNodeDebugInfo, attachNodeArrayDebugInfo, enableDebugInfo, formatVariance, attachDebugPrototypeIfDebug, printControlFlowGraph, formatControlFlowGraph |
| src/compiler/debug.ts:274:54 | hasProperty: `"kind"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: kind |
| src/compiler/debug.ts:274:85 | hasProperty: `"pos"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: pos |
| src/compiler/debug.ts:372:18 | hasProperty: `"name"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: name |
| src/compiler/debug.ts:433:27 | element: `name` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/debug.ts:713:18 | hasProperty: `"__debugKind"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: __debugKind |
| src/compiler/executeCommandLine.ts:408:30 | element: `value` | fixed_set_excludes_prototype | Values of internal custom option maps (enum numbers and enumerated strings). Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/factory/nodeFactory.ts:6138:17 | hasProperty: `p` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/factory/nodeFactory.ts:6138:42 | hasProperty: `p` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/factory/nodeFactory.ts:6145:32 | element: `p` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/factory/nodeFactory.ts:6400:17 | hasProperty: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/factory/nodeFactory.ts:6400:45 | hasProperty: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/factory/nodeFactory.ts:6404:26 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/factory/nodeFactory.ts:7539:27 | element: `key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/moduleNameResolver.ts:362:10 | hasProperty: `fieldName` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: exports, types, version, type, name, imports, typings, typesVersions, main, tsconfig, dependencies, peerDependencies, optionalDependencies |
| src/compiler/moduleNameResolver.ts:368:19 | element: `fieldName` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: exports, types, version, type, name, imports, typings, typesVersions, main, tsconfig, dependencies, peerDependencies, optionalDependencies |
| src/compiler/moduleNameResolver.ts:433:17 | hasProperty: `key` | user_input | package.json: key is enumerated from package.json typesVersions. The first check validates version ranges; the second is an own-property guard before version selection. |
| src/compiler/moduleNameResolver.ts:465:14 | hasProperty: `key` | user_input | package.json: key is enumerated from package.json typesVersions. The first check validates version ranges; the second is an own-property guard before version selection. |
| src/compiler/moduleNameResolver.ts:474:43 | element: `key` | user_input | package.json: package.json typesVersions key; semantic version parsing filters before read. |
| src/compiler/moduleNameResolver.ts:953:13 | hasProperty: `key` | user_input | tsconfig: compilerOptionValueToString serializes getCompilerOptionValue results (moduleNameResolver.ts:962); nested record/object keys come from tsconfig option values such as paths and plugins. |
| src/compiler/moduleNameResolver.ts:954:59 | element: `key` | user_input | tsconfig: compilerOptionValueToString serializes getCompilerOptionValue results (moduleNameResolver.ts:962); nested record/object keys come from tsconfig option values such as paths and plugins. |
| src/compiler/moduleNameResolver.ts:2295:46 | element: `key` | user_input | package.json: package.json exports own key. |
| src/compiler/moduleNameResolver.ts:2351:54 | element: `key` | user_input | package.json: package.json nested exports own key. |
| src/compiler/moduleNameResolver.ts:2429:13 | hasProperty: `key` | user_input | package.json: key is enumerated from package.json peerDependencies read at 2422. |
| src/compiler/moduleNameResolver.ts:2629:18 | hasProperty: `"."` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: . |
| src/compiler/moduleNameResolver.ts:2711:83 | hasProperty: `moduleName` | user_input | source code: moduleName is derived from a source module specifier: exports subpath (2644) or imports specifier (2678), read against a package.json table. The table origin is not the origin of this lookup key; hasProperty guards the value read. |
| src/compiler/moduleNameResolver.ts:2712:24 | element: `moduleName` | user_input | source code: moduleName is derived from a source module specifier: exports subpath (2644) or imports specifier (2678), read against a package.json table. The table origin is not the origin of this lookup key; hasProperty guards the value read. |
| src/compiler/moduleNameResolver.ts:2718:28 | element: `potentialTarget` | user_input | package.json: package.json export/import pattern containing * or ending /; excludes prototype names. |
| src/compiler/moduleNameResolver.ts:2724:28 | element: `potentialTarget` | user_input | package.json: package.json export/import pattern containing * or ending /; excludes prototype names. |
| src/compiler/moduleNameResolver.ts:2729:28 | element: `potentialTarget` | user_input | package.json: package.json export/import pattern containing * or ending /; excludes prototype names. |
| src/compiler/moduleNameResolver.ts:2820:43 | element: `condition` | user_input | package.json: package.json own condition, may match customConditions. |
| src/compiler/moduleNameResolver.ts:3103:14 | hasProperty: `"exports"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: exports |
| src/compiler/moduleNameResolver.ts:3177:34 | element: `matchedPatternText` | user_input | tsconfig + package.json: tryLoadModuleUsingPaths receives both tsconfig paths (1578) and package.json typesVersions paths (2538, 3161); matchedPatternText comes from that table. |
| src/compiler/moduleSpecifiers.ts:801:22 | element: `field` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: dependencies, peerDependencies, optionalDependencies |
| src/compiler/moduleSpecifiers.ts:928:35 | element: `key` | user_input | tsconfig + package.json: tryGetModuleNameFromPaths receives both tsconfig paths (614) and package.json typesVersions paths (1288); key is enumerated from the selected table. |
| src/compiler/moduleSpecifiers.ts:1111:35 | element: `key` | user_input | package.json: package.json conditional exports own key. |
| src/compiler/moduleSpecifiers.ts:1134:122 | element: `k` | user_input | package.json: package.json export key constrained to start with dot; excludes prototype names. |
| src/compiler/moduleSpecifiers.ts:1165:121 | element: `k` | user_input | package.json: package.json imports key constrained to start with #; excludes prototype names. |
| src/compiler/parser.ts:10718:24 | element: `name` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: jsx, reference, amd-dependency, amd-module, ts-check, ts-nocheck, jsxfrag, jsximportsource, jsxruntime |
| src/compiler/parser.ts:10770:20 | element: `name` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: jsx, reference, amd-dependency, amd-module, ts-check, ts-nocheck, jsxfrag, jsximportsource, jsxruntime |
| src/compiler/program.ts:1524:13 | hasProperty: `option.name` | fixed_set_excludes_prototype | option.name is from the internal commandLineOptionOfCustomType array at program.ts:1523. |
| src/compiler/program.ts:1525:24 | element: `option.name` | fixed_set_excludes_prototype | Internal commandLineOptionOfCustomType declarations. Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/program.ts:4142:22 | hasProperty: `key` | user_input | tsconfig: key is enumerated from options.paths, populated by tsconfig paths. |
| src/compiler/program.ts:4148:29 | element: `key` | user_input | tsconfig: tsconfig paths own key. |
| src/compiler/program.ts:4149:33 | element: `key` | user_input | tsconfig: tsconfig paths own key. |
| src/compiler/program.ts:4154:39 | element: `key` | user_input | tsconfig: tsconfig paths own key. |
| src/compiler/program.ts:4379:42 | element: `moduleKindName as any` | fixed_set_excludes_prototype | moduleKindName is a reverse ModuleKind lookup only under Node16 <= moduleKind <= NodeNext (4374-4378); the possible names are Node16, Node18, Node20, NodeNext. |
| src/compiler/program.ts:5099:53 | element: `d.skippedOn` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/scanner.ts:3507:69 | element: `propertyName` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: General_Category, Script, Script_Extensions |
| src/compiler/scanner.ts:3509:89 | element: `propertyName` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: General_Category, Script, Script_Extensions |
| src/compiler/sys.ts:171:29 | element: `level` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: Low, Medium, High |
| src/compiler/sys.ts:171:53 | element: `level` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: Low, Medium, High |
| src/compiler/sys.ts:1566:24 | element: `name` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/tsbuildPublic.ts:334:13 | hasProperty: `option.name` | fixed_set_excludes_prototype | option.name is from the private commonOptionsWithBuild table at tsbuildPublic.ts:333, not an arbitrary options object key. |
| src/compiler/tsbuildPublic.ts:334:75 | element: `option.name` | fixed_set_excludes_prototype | Internal optionDeclarations loop. Reviewed against the pinned internal declarations; does not infer safety merely from hasProperty. |
| src/compiler/utilities.ts:8031:29 | hasProperty: `"watch"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: watch |
| src/compiler/utilities.ts:8153:20 | element: `e` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/utilities.ts:8154:37 | element: `e` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/utilities.ts:8154:45 | element: `e` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/utilities.ts:8158:25 | element: `e` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/utilities.ts:8159:17 | element: `e` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/utilities.ts:8159:28 | element: `e` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/utilities.ts:8605:43 | element: `message.key` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/utilities.ts:9370:12 | element: `flag` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: noImplicitAny, noImplicitThis, strictFunctionTypes, strictBindCallApply, strictNullChecks, strictPropertyInitialization, strictBuiltinIteratorReturn, useUnknownInCatchVariables |
| src/compiler/utilities.ts:9370:89 | element: `flag` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: noImplicitAny, noImplicitThis, strictFunctionTypes, strictBindCallApply, strictNullChecks, strictPropertyInitialization, strictBuiltinIteratorReturn, useUnknownInCatchVariables |
| src/compiler/utilities.ts:9402:9 | element: `option.name` | unknown | Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name. |
| src/compiler/utilities.ts:9675:88 | element: `usage` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, exclude, directories |
| src/compiler/utilities.ts:9690:74 | element: `usage` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, exclude, directories |
| src/compiler/utilities.ts:9699:111 | element: `usage` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: files, exclude, directories |
| src/compiler/utilitiesPublic.ts:1488:12 | hasProperty: `"pos"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: pos |
| src/compiler/utilitiesPublic.ts:1488:41 | hasProperty: `"end"` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: end |
| src/compiler/watchUtilities.ts:736:13 | element: `key` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: watchDirectory, watchFile |
| src/compiler/watchUtilities.ts:811:14 | element: `key` | fixed_set_excludes_prototype | Finite literal key set excludes all 12 prototype names: watchDirectory, watchFile |

## for-in body coverage

| Loop location | Receiver | Counted reads/checks |
| --- | --- | ---: |
| src/compiler/commandLineParser.ts:2688:5 | `computedOptions` | 3 |
| src/compiler/commandLineParser.ts:2788:5 | `options` | 2 |
| src/compiler/commandLineParser.ts:2975:5 | `options` | 2 |
| src/compiler/commandLineParser.ts:3782:5 | `jsonOptions` | 1 |
| src/compiler/commandLineParser.ts:4154:9 | `wildcardDirectories` | 1 |
| src/compiler/commandLineParser.ts:4265:5 | `opts` | 2 |
| src/compiler/core.ts:1289:5 | `map` | 1 |
| src/compiler/core.ts:1314:5 | `collection` | 2 |
| src/compiler/core.ts:1353:9 | `arg` | 2 |
| src/compiler/core.ts:1373:5 | `left` | 4 |
| src/compiler/core.ts:1380:5 | `right` | 2 |
| src/compiler/core.ts:1480:5 | `object` | 2 |
| src/compiler/core.ts:1497:5 | `second` | 2 |
| src/compiler/core.ts:1503:5 | `first` | 2 |
| src/compiler/core.ts:1514:5 | `second` | 2 |
| src/compiler/debug.ts:432:9 | `enumObject` | 1 |
| src/compiler/factory/nodeFactory.ts:6137:9 | `source` | 3 |
| src/compiler/factory/nodeFactory.ts:6399:9 | `node` | 3 |
| src/compiler/factory/nodeFactory.ts:7538:5 | `sourceRanges` | 1 |
| src/compiler/moduleNameResolver.ts:432:9 | `typesVersions` | 1 |
| src/compiler/moduleNameResolver.ts:464:5 | `typesVersions` | 2 |
| src/compiler/moduleNameResolver.ts:952:5 | `value` | 2 |
| src/compiler/moduleNameResolver.ts:2294:9 | `exports` | 1 |
| src/compiler/moduleNameResolver.ts:2428:5 | `peerDependencies` | 1 |
| src/compiler/moduleSpecifiers.ts:927:5 | `paths` | 1 |
| src/compiler/program.ts:4141:13 | `options.paths` | 5 |
| src/compiler/utilities.ts:8152:5 | `dst` | 6 |

## Non-string dynamic reads

These are included in the 1,041 global total and in the fixed-key exclusion class. The class includes statically known non-string domains, not only finite string unions. All locations and key/receiver types appear in inherited-key-global.json. Their keys cannot become constructor, toString, or another listed prototype name under ToPropertyKey.

| Location | Key type | Read |
| --- | --- | --- |
| src/compiler/binder.ts:1746:21 | `number` | `clauses[i]` |
| src/compiler/binder.ts:1750:22 | `number` | `clauses[i]` |
| src/compiler/binder.ts:1757:28 | `number` | `clauses[i]` |
| src/compiler/binder.ts:2005:39 | `number` | `state.inStrictModeStack[state.stackIndex]` |
| src/compiler/binder.ts:2006:33 | `number` | `state.parentStack[state.stackIndex]` |
| src/compiler/builder.ts:626:57 | `number` | `state.affectedFiles[state.affectedFilesIndex! - 1]` |
| src/compiler/builder.ts:646:38 | `number` | `affectedFiles[affectedFilesIndex]` |
| src/compiler/builder.ts:1296:22 | `number` | `fileNames[fileId - 1]` |
| src/compiler/builder.ts:1413:22 | `number` | `root[root.length - 1]` |
| src/compiler/builder.ts:1419:28 | `number` | `root[root.length - 2]` |
| src/compiler/builder.ts:1552:62 | `number` | `array[i]` |
| src/compiler/builder.ts:1631:82 | `DiagnosticCategory` | `DiagnosticCategory[diagnostic.category]` |
| src/compiler/builder.ts:2359:16 | `number` | `filePaths[fileId - 1]` |
| src/compiler/builder.ts:2363:16 | `number` | `filePathsSetList![fileIdsListId - 1]` |
| src/compiler/builder.ts:2408:29 | `number` | `program.fileNames[index]` |
| src/compiler/builder.ts:2412:29 | `number` | `program.root[rootIndex]` |
| src/compiler/builder.ts:2431:30 | `number` | `program.fileNames[root - 1]` |
| src/compiler/checker.ts:2935:16 | `number` | `symbolLinks[id]` |
| src/compiler/checker.ts:2940:16 | `number` | `nodeLinks[nodeId]` |
| src/compiler/checker.ts:4094:143 | `ModuleKind` | `ModuleKind[moduleKind]` |
| src/compiler/checker.ts:5311:54 | `number` | `mergedSymbols[symbol.mergeId]` |
| src/compiler/checker.ts:6731:21 | `number` | `context.typeStack[i]` |
| src/compiler/checker.ts:6748:21 | `number` | `context.typeStack[i]` |
| src/compiler/checker.ts:7014:112 | `number` | `texts[i + 1]` |
| src/compiler/checker.ts:7432:94 | `number` | `(type.target as TupleType).elementFlags[i]` |
| src/compiler/checker.ts:7439:47 | `number` | `(type.target as TupleType).elementFlags[i]` |
| src/compiler/checker.ts:7440:67 | `number` | `labeledElementDeclarations?.[i]` |
| src/compiler/checker.ts:7447:97 | `number` | `tupleConstituentNodes[i]` |
| src/compiler/checker.ts:7448:45 | `number` | `tupleConstituentNodes[i]` |
| src/compiler/checker.ts:7452:179 | `number` | `tupleConstituentNodes[i]` |
| src/compiler/checker.ts:7452:207 | `number` | `tupleConstituentNodes[i]` |
| src/compiler/checker.ts:7453:104 | `number` | `tupleConstituentNodes[i]` |
| src/compiler/checker.ts:7454:41 | `number` | `tupleConstituentNodes[i]` |
| src/compiler/checker.ts:7485:75 | `number` | `outerTypeParameters[i]` |
| src/compiler/checker.ts:7489:81 | `number` | `outerTypeParameters[i]` |
| src/compiler/checker.ts:7523:62 | `number` | `typeArguments[typeParameterCount - 1]` |
| src/compiler/checker.ts:7524:63 | `number` | `type.target.typeParameters[typeParameterCount - 1]` |
| src/compiler/checker.ts:7700:50 | `number` | `properties[properties.length - 1]` |
| src/compiler/checker.ts:7744:34 | `number` | `context.reverseMappedStack![context.reverseMappedStack!.length - 1 - i]` |
| src/compiler/checker.ts:7913:50 | `number` | `types[types.length - 1]` |
| src/compiler/checker.ts:7931:63 | `number` | `types[types.length - 1]` |
| src/compiler/checker.ts:8027:65 | `number` | `expandedParams[expandedParams.length - 1]` |
| src/compiler/checker.ts:8208:43 | `number` | `expandedParams[pIndex]` |
| src/compiler/checker.ts:8209:51 | `number` | `originalParameters?.[pIndex]` |
| src/compiler/checker.ts:8491:64 | `number` | `parents![i]` |
| src/compiler/checker.ts:8528:40 | `number` | `parentSpecifiers[a]` |
| src/compiler/checker.ts:8529:40 | `number` | `parentSpecifiers[b]` |
| src/compiler/checker.ts:8559:28 | `number` | `chain[index]` |
| src/compiler/checker.ts:8572:36 | `number` | `chain[index + 1]` |
| src/compiler/checker.ts:8756:32 | `number` | `chain[index]` |
| src/compiler/checker.ts:8757:32 | `number` | `chain[index - 1]` |
| src/compiler/checker.ts:8888:32 | `number` | `chain[index]` |
| src/compiler/checker.ts:8913:32 | `number` | `chain[index]` |
| src/compiler/checker.ts:9422:43 | `number` | `statements[nsIndex]` |
| src/compiler/checker.ts:9512:40 | `number` | `statements[index]` |
| src/compiler/checker.ts:9518:88 | `number` | `statements[i]` |
| src/compiler/checker.ts:9519:114 | `number` | `statements[i]` |
| src/compiler/checker.ts:9521:75 | `number` | `statements[index]` |
| src/compiler/checker.ts:9588:41 | `number` | `symbols[symbols.length - 1]` |
| src/compiler/checker.ts:9597:21 | `number` | `deferredPrivatesStack[deferredPrivatesStack.length - 1]` |
| src/compiler/checker.ts:9847:39 | `number` | `deferredPrivatesStack[deferredPrivatesStack.length - 1]` |
| src/compiler/checker.ts:9858:17 | `number` | `deferredPrivatesStack[isExternalImportAlias ? 0 : (deferredPrivatesStack.length - 1)]` |
| src/compiler/checker.ts:9965:61 | `number` | `props[props.length - 1]` |
| src/compiler/checker.ts:9966:65 | `number` | `props[props.length - 1]` |
| src/compiler/checker.ts:10098:38 | `number` | `memberProps[memberProps.length - 1]` |
| src/compiler/checker.ts:10997:65 | `number` | `signatures[i]` |
| src/compiler/checker.ts:10997:80 | `number` | `baseSigs[i]` |
| src/compiler/checker.ts:11192:23 | `number` | `types[i]` |
| src/compiler/checker.ts:11199:86 | `number` | `types[i + count - 1]` |
| src/compiler/checker.ts:11199:140 | `number` | `(baseType as UnionType).types[count - 1]` |
| src/compiler/checker.ts:11505:45 | `number` | `resolutionTargets[i]` |
| src/compiler/checker.ts:11505:67 | `number` | `resolutionPropertyNames[i]` |
| src/compiler/checker.ts:11508:17 | `number` | `resolutionTargets[i]` |
| src/compiler/checker.ts:11508:52 | `number` | `resolutionPropertyNames[i]` |
| src/compiler/checker.ts:12342:33 | `number` | `declarations[i]` |
| src/compiler/checker.ts:13313:69 | `number` | `type.elementFlags[i]` |
| src/compiler/checker.ts:13381:20 | `number` | `outerTypeParameters[last]` |
| src/compiler/checker.ts:13381:57 | `number` | `typeArguments[last]` |
| src/compiler/checker.ts:14190:32 | `number` | `sig.parameters[restIndex]` |
| src/compiler/checker.ts:14206:49 | `number` | `associatedNames[i]` |
| src/compiler/checker.ts:14206:70 | `number` | `associatedNames[i]` |
| src/compiler/checker.ts:14208:31 | `number` | `restType.target.elementFlags[i]` |
| src/compiler/checker.ts:14219:134 | `number` | `type.target.elementFlags[i]` |
| src/compiler/checker.ts:14224:34 | `number` | `names[i]` |
| src/compiler/checker.ts:14231:48 | `number` | `names[i]` |
| src/compiler/checker.ts:14233:64 | `number` | `names[i]` |
| src/compiler/checker.ts:14237:34 | `number` | `names[i]` |
| src/compiler/checker.ts:14287:44 | `number` | `signatureLists[i]` |
| src/compiler/checker.ts:14299:41 | `number` | `signatureLists[i]` |
| src/compiler/checker.ts:14300:46 | `number` | `signatureLists[i]` |
| src/compiler/checker.ts:14317:17 | `number` | `signatureLists[i]` |
| src/compiler/checker.ts:14318:17 | `number` | `signatureLists[i]` |
| src/compiler/checker.ts:14321:37 | `number` | `signatureLists[i]` |
| src/compiler/checker.ts:14348:32 | `number` | `signatureLists[indexWithLengthOverOne !== undefined ? indexWithLengthOverOne : 0]` |
| src/compiler/checker.ts:14375:28 | `number` | `sourceParams[i]` |
| src/compiler/checker.ts:14376:28 | `number` | `targetParams[i]` |
| src/compiler/checker.ts:14530:22 | `number` | `mixinFlags[i]` |
| src/compiler/checker.ts:14531:78 | `number` | `types[i]` |
| src/compiler/checker.ts:14547:23 | `number` | `type.types[i]` |
| src/compiler/checker.ts:14553:18 | `number` | `mixinFlags[i]` |
| src/compiler/checker.ts:14582:30 | `number` | `indexInfos[i]` |
| src/compiler/checker.ts:15172:87 | `number` | `type.target.elementFlags[i]` |
| src/compiler/checker.ts:15439:77 | `number` | `t.target.elementFlags[i]` |
| src/compiler/checker.ts:16133:46 | `number` | `typeParameters[i]` |
| src/compiler/checker.ts:16165:63 | `number` | `typeParameters![i]` |
| src/compiler/checker.ts:16203:31 | `number` | `declaration.parameters[i]` |
| src/compiler/checker.ts:16370:26 | `number` | `symbol.declarations[i]` |
| src/compiler/checker.ts:16376:34 | `number` | `symbol.declarations[i - 1]` |
| src/compiler/checker.ts:16554:49 | `number` | `signature.parameters[signature.parameters.length - 1]` |
| src/compiler/checker.ts:16780:89 | `number` | `typeParameters[index]` |
| src/compiler/checker.ts:16878:33 | `number` | `types[i]` |
| src/compiler/checker.ts:16880:46 | `number` | `types[i + count]` |
| src/compiler/checker.ts:17886:31 | `number` | `elementFlags[i]` |
| src/compiler/checker.ts:17890:60 | `number` | `namedMemberDeclarations?.[i]` |
| src/compiler/checker.ts:17943:69 | `number` | `target.elementFlags[i]` |
| src/compiler/checker.ts:17945:75 | `number` | `target.elementFlags[i]` |
| src/compiler/checker.ts:17946:29 | `number` | `elementTypes[unionIndex]` |
| src/compiler/checker.ts:17962:26 | `number` | `elementTypes[i]` |
| src/compiler/checker.ts:17963:27 | `number` | `target.elementFlags[i]` |
| src/compiler/checker.ts:17966:57 | `number` | `target.labeledElementDeclarations?.[i]` |
| src/compiler/checker.ts:17970:61 | `number` | `target.labeledElementDeclarations?.[i]` |
| src/compiler/checker.ts:17984:63 | `number` | `type.target.elementFlags[n]` |
| src/compiler/checker.ts:17984:92 | `number` | `type.target.labeledElementDeclarations?.[n]` |
| src/compiler/checker.ts:17988:127 | `number` | `target.labeledElementDeclarations?.[i]` |
| src/compiler/checker.ts:17993:41 | `number` | `target.labeledElementDeclarations?.[i]` |
| src/compiler/checker.ts:17998:17 | `number` | `expandedFlags[i]` |
| src/compiler/checker.ts:18002:142 | `number` | `expandedFlags[firstRestIndex + i]` |
| src/compiler/checker.ts:18095:124 | `number` | `typeSet[len - 1]` |
| src/compiler/checker.ts:18143:28 | `number` | `types[i]` |
| src/compiler/checker.ts:18205:23 | `number` | `types[i]` |
| src/compiler/checker.ts:18225:27 | `number` | `types[i]` |
| src/compiler/checker.ts:18245:45 | `0 \| 1` | `(type as IntersectionType).types[index]` |
| src/compiler/checker.ts:18257:25 | `0 \| 1` | `(type as IntersectionType).types[index]` |
| src/compiler/checker.ts:18258:48 | `number` | `(type as IntersectionType).types[1 - index]` |
| src/compiler/checker.ts:18269:34 | `number` | `types[i]` |
| src/compiler/checker.ts:18272:29 | `0 \| 1` | `(type as IntersectionType).types[index]` |
| src/compiler/checker.ts:18272:114 | `number` | `(type as IntersectionType).types[1 - index]` |
| src/compiler/checker.ts:18324:24 | `0 \| 1` | `types[index]` |
| src/compiler/checker.ts:18324:50 | `number` | `types[1 - index]` |
| src/compiler/checker.ts:18518:23 | `number` | `types[i]` |
| src/compiler/checker.ts:18565:23 | `number` | `types[i]` |
| src/compiler/checker.ts:18583:35 | `number` | `types[i]` |
| src/compiler/checker.ts:18600:23 | `number` | `types[i]` |
| src/compiler/checker.ts:18602:47 | `number` | `types[index]` |
| src/compiler/checker.ts:18717:34 | `0 \| 1` | `typeSet[typeVarIndex]` |
| src/compiler/checker.ts:18718:35 | `number` | `typeSet[1 - typeVarIndex]` |
| src/compiler/checker.ts:18819:21 | `number` | `types[j]` |
| src/compiler/checker.ts:18820:42 | `number` | `types[j]` |
| src/compiler/checker.ts:18822:39 | `number` | `sourceTypes[n % length]` |
| src/compiler/checker.ts:18851:41 | `number` | `types[1 - emptyIndex]` |
| src/compiler/checker.ts:19044:25 | `number` | `types[unionIndex]` |
| src/compiler/checker.ts:19078:27 | `number` | `types[i]` |
| src/compiler/checker.ts:19081:29 | `number` | `texts[i + 1]` |
| src/compiler/checker.ts:19086:29 | `number` | `texts[i + 1]` |
| src/compiler/checker.ts:19091:28 | `number` | `texts[i + 1]` |
| src/compiler/checker.ts:20194:34 | `number` | `types[types.length - 1]` |
| src/compiler/checker.ts:20543:30 | `number` | `items[i]` |
| src/compiler/checker.ts:20549:50 | `number` | `items[i]` |
| src/compiler/checker.ts:20584:34 | `number` | `sources[i]` |
| src/compiler/checker.ts:20585:42 | `number` | `targets[i]` |
| src/compiler/checker.ts:20594:34 | `number` | `sources[i]` |
| src/compiler/checker.ts:20595:32 | `number` | `targets[i]` |
| src/compiler/checker.ts:20936:27 | `number` | `elementFlags[i]` |
| src/compiler/checker.ts:21037:29 | `number` | `activeTypeMappersCaches[index !== -1 ? index : activeTypeMappersCount - 1]` |
| src/compiler/checker.ts:21435:36 | `number` | `resultObj.errors![resultObj.errors!.length - 1]` |
| src/compiler/checker.ts:21486:25 | `number` | `resultObj.errors[resultObj.errors.length - 1]` |
| src/compiler/checker.ts:21501:25 | `number` | `resultObj.errors[resultObj.errors.length - 1]` |
| src/compiler/checker.ts:21582:46 | `number` | `resultObj.errors[resultObj.errors.length - 1]` |
| src/compiler/checker.ts:21689:27 | `number` | `node.children[i]` |
| src/compiler/checker.ts:21828:26 | `number` | `node.elements[i]` |
| src/compiler/checker.ts:22527:52 | `number` | `str[str.length - 1]` |
| src/compiler/checker.ts:22579:21 | `number` | `path[path.length - 1]` |
| src/compiler/checker.ts:23204:45 | `number` | `sourceTypes[i]` |
| src/compiler/checker.ts:23229:36 | `number` | `sourceTypes[i]` |
| src/compiler/checker.ts:23236:61 | `number` | `(undefinedStrippedTarget as UnionType).types[i % (undefinedStrippedTarget as UnionType).types.length]` |
| src/compiler/checker.ts:23261:62 | `number` | `variances[i]` |
| src/compiler/checker.ts:23265:31 | `number` | `sources[i]` |
| src/compiler/checker.ts:23266:31 | `number` | `targets[i]` |
| src/compiler/checker.ts:23459:41 | `number` | `maybeKeys[i]` |
| src/compiler/checker.ts:23461:38 | `number` | `maybeKeys[i]` |
| src/compiler/checker.ts:23595:57 | `number` | `sourceTypes[i]` |
| src/compiler/checker.ts:23595:73 | `number` | `targetTypes[i]` |
| src/compiler/checker.ts:24221:40 | `number` | `sourcePropertiesFiltered[i]` |
| src/compiler/checker.ts:24238:48 | `number` | `sourcePropertiesFiltered[i]` |
| src/compiler/checker.ts:24243:112 | `number` | `combination[i]` |
| src/compiler/checker.ts:24287:45 | `number` | `properties[i]` |
| src/compiler/checker.ts:24289:37 | `number` | `properties[i]` |
| src/compiler/checker.ts:24478:67 | `number` | `source.target.elementFlags[sourcePosition]` |
| src/compiler/checker.ts:24485:45 | `number` | `target.target.elementFlags[targetPosition]` |
| src/compiler/checker.ts:24515:62 | `number` | `sourceTypeArguments[sourcePosition]` |
| src/compiler/checker.ts:24516:44 | `number` | `targetTypeArguments[targetPosition]` |
| src/compiler/checker.ts:24660:56 | `number` | `sourceSignatures[i]` |
| src/compiler/checker.ts:24660:77 | `number` | `targetSignatures[i]` |
| src/compiler/checker.ts:24660:168 | `number` | `sourceSignatures[i]` |
| src/compiler/checker.ts:24660:189 | `number` | `targetSignatures[i]` |
| src/compiler/checker.ts:24762:60 | `number` | `sourceSignatures[i]` |
| src/compiler/checker.ts:24762:81 | `number` | `targetSignatures[i]` |
| src/compiler/checker.ts:24955:21 | `number` | `include[i]` |
| src/compiler/checker.ts:24956:80 | `number` | `types[i]` |
| src/compiler/checker.ts:24969:21 | `number` | `include[i]` |
| src/compiler/checker.ts:24974:97 | `number` | `include[i]` |
| src/compiler/checker.ts:25105:18 | `number` | `variances[i]` |
| src/compiler/checker.ts:25105:92 | `number` | `typeArguments[i]` |
| src/compiler/checker.ts:25251:27 | `number` | `stack[i]` |
| src/compiler/checker.ts:25414:27 | `number` | `source.typeParameters![i]` |
| src/compiler/checker.ts:25415:27 | `number` | `target.typeParameters[i]` |
| src/compiler/checker.ts:25748:27 | `number` | `typeArguments[i]` |
| src/compiler/checker.ts:25749:35 | `number` | `type.target.elementFlags[i]` |
| src/compiler/checker.ts:25758:86 | `number` | `t2.target.elementFlags[i]` |
| src/compiler/checker.ts:26603:27 | `number` | `source.texts[source.texts.length - 1]` |
| src/compiler/checker.ts:26604:27 | `number` | `target.texts[target.texts.length - 1]` |
| src/compiler/checker.ts:26681:95 | `number` | `target.types[i]` |
| src/compiler/checker.ts:26689:104 | `number` | `target.types[i]` |
| src/compiler/checker.ts:26718:31 | `number` | `sourceTexts[lastSourceIndex]` |
| src/compiler/checker.ts:26722:31 | `number` | `targetTexts[lastTargetIndex]` |
| src/compiler/checker.ts:26732:27 | `number` | `targetTexts[i]` |
| src/compiler/checker.ts:26759:46 | `number` | `sourceTexts[index]` |
| src/compiler/checker.ts:26765:22 | `number` | `sourceTexts[seg]` |
| src/compiler/checker.ts:27091:46 | `number` | `variances[i]` |
| src/compiler/checker.ts:27092:49 | `number` | `sourceTypes[i]` |
| src/compiler/checker.ts:27092:65 | `number` | `targetTypes[i]` |
| src/compiler/checker.ts:27095:36 | `number` | `sourceTypes[i]` |
| src/compiler/checker.ts:27095:52 | `number` | `targetTypes[i]` |
| src/compiler/checker.ts:27158:44 | `number` | `sources[i]` |
| src/compiler/checker.ts:27180:66 | `number` | `matched[i]` |
| src/compiler/checker.ts:27290:46 | `number` | `matches[i]` |
| src/compiler/checker.ts:27291:36 | `number` | `types[i]` |
| src/compiler/checker.ts:27388:48 | `number` | `getTypeArguments(source)[i]` |
| src/compiler/checker.ts:27388:77 | `number` | `elementTypes[i]` |
| src/compiler/checker.ts:27396:44 | `number` | `getTypeArguments(source)[i]` |
| src/compiler/checker.ts:27396:73 | `number` | `elementTypes[i]` |
| src/compiler/checker.ts:27398:100 | `number` | `source.target.elementFlags[startLength]` |
| src/compiler/checker.ts:27400:46 | `number` | `getTypeArguments(source)[startLength]` |
| src/compiler/checker.ts:27402:48 | `number` | `elementFlags[i]` |
| src/compiler/checker.ts:27402:128 | `number` | `elementTypes[i]` |
| src/compiler/checker.ts:27408:37 | `number` | `elementFlags[startLength]` |
| src/compiler/checker.ts:27408:65 | `number` | `elementFlags[startLength + 1]` |
| src/compiler/checker.ts:27410:80 | `number` | `elementTypes[startLength]` |
| src/compiler/checker.ts:27413:144 | `number` | `elementTypes[startLength]` |
| src/compiler/checker.ts:27414:130 | `number` | `elementTypes[startLength + 1]` |
| src/compiler/checker.ts:27417:42 | `number` | `elementFlags[startLength]` |
| src/compiler/checker.ts:27417:95 | `number` | `elementFlags[startLength + 1]` |
| src/compiler/checker.ts:27420:75 | `number` | `elementTypes[startLength]` |
| src/compiler/checker.ts:27424:137 | `number` | `elementTypes[startLength]` |
| src/compiler/checker.ts:27425:138 | `number` | `elementTypes[startLength + 1]` |
| src/compiler/checker.ts:27428:42 | `number` | `elementFlags[startLength]` |
| src/compiler/checker.ts:27428:91 | `number` | `elementFlags[startLength + 1]` |
| src/compiler/checker.ts:27431:75 | `number` | `elementTypes[startLength + 1]` |
| src/compiler/checker.ts:27439:138 | `number` | `elementTypes[startLength]` |
| src/compiler/checker.ts:27440:71 | `number` | `elementTypes[startLength + 1]` |
| src/compiler/checker.ts:27444:60 | `number` | `elementFlags[startLength]` |
| src/compiler/checker.ts:27447:56 | `number` | `target.target.elementFlags[targetArity - 1]` |
| src/compiler/checker.ts:27449:64 | `number` | `elementTypes[startLength]` |
| src/compiler/checker.ts:27451:60 | `number` | `elementFlags[startLength]` |
| src/compiler/checker.ts:27455:62 | `number` | `elementTypes[startLength]` |
| src/compiler/checker.ts:27461:44 | `number` | `getTypeArguments(source)[sourceArity - i - 1]` |
| src/compiler/checker.ts:27461:91 | `number` | `elementTypes[targetArity - i - 1]` |
| src/compiler/checker.ts:27500:57 | `number` | `sourceSignatures[sourceIndex]` |
| src/compiler/checker.ts:27500:108 | `number` | `targetSignatures[i]` |
| src/compiler/checker.ts:27609:27 | `number` | `context.inferences[index]` |
| src/compiler/checker.ts:28648:16 | `number` | `evolvingArrayTypes[elementType.id]` |
| src/compiler/checker.ts:28835:20 | `number` | `callExpression.arguments[predicate.parameterIndex]` |
| src/compiler/checker.ts:28872:39 | `number` | `flowNodeReachable[id]` |
| src/compiler/checker.ts:28885:51 | `number` | `(flow as FlowCall).node.arguments[predicate.parameterIndex]` |
| src/compiler/checker.ts:28941:39 | `number` | `flowNodePostSuper[id]` |
| src/compiler/checker.ts:29054:29 | `number` | `sharedFlowNodes[i]` |
| src/compiler/checker.ts:29056:36 | `number` | `sharedFlowTypes[i]` |
| src/compiler/checker.ts:29225:200 | `number` | `flow.node.arguments[predicate.parameterIndex]` |
| src/compiler/checker.ts:29376:27 | `number` | `flowLoopCaches[id]` |
| src/compiler/checker.ts:29395:21 | `number` | `flowLoopNodes[i]` |
| src/compiler/checker.ts:29395:50 | `number` | `flowLoopKeys[i]` |
| src/compiler/checker.ts:29395:77 | `number` | `flowLoopTypes[i]` |
| src/compiler/checker.ts:29396:71 | `number` | `flowLoopTypes[i]` |
| src/compiler/checker.ts:29834:31 | `number` | `clauseTypes[i]` |
| src/compiler/checker.ts:29925:32 | `number` | `switchStatement.caseBlock.clauses[i]` |
| src/compiler/checker.ts:29936:36 | `number` | `switchStatement.caseBlock.clauses[i]` |
| src/compiler/checker.ts:31953:55 | `number` | `args[indexOfParameter]` |
| src/compiler/checker.ts:32182:50 | `number` | `signature.parameters[restIndex]` |
| src/compiler/checker.ts:32517:33 | `number` | `elements[i]` |
| src/compiler/checker.ts:32531:46 | `number` | `getTypeArguments(t)[index]` |
| src/compiler/checker.ts:32531:77 | `number` | `t.target.elementFlags[index]` |
| src/compiler/checker.ts:32540:28 | `number` | `getTypeArguments(t)[getTypeReferenceArity(t) - offset]` |
| src/compiler/checker.ts:32790:20 | `number` | `contextualTypes[index]` |
| src/compiler/checker.ts:32888:26 | `number` | `contextualTypeNodes[i]` |
| src/compiler/checker.ts:32888:71 | `number` | `contextualIsCache[i]` |
| src/compiler/checker.ts:32909:42 | `number` | `inferenceContextNodes[i]` |
| src/compiler/checker.ts:32910:24 | `number` | `inferenceContexts[i]` |
| src/compiler/checker.ts:32917:9 | `number` | `activeTypeMappersCaches[activeTypeMappersCount]` |
| src/compiler/checker.ts:32925:9 | `number` | `activeTypeMappersCaches[activeTypeMappersCount]` |
| src/compiler/checker.ts:32930:28 | `number` | `activeTypeMappers[i]` |
| src/compiler/checker.ts:32939:13 | `number` | `activeTypeMappersCaches[i]` |
| src/compiler/checker.ts:32954:24 | `number` | `contextualTypes[index]` |
| src/compiler/checker.ts:33201:27 | `number` | `target.parameters[targetParameterCount]` |
| src/compiler/checker.ts:33342:23 | `number` | `elements[i]` |
| src/compiler/checker.ts:33401:62 | `number` | `elementFlags[i]` |
| src/compiler/checker.ts:33499:26 | `number` | `properties[i]` |
| src/compiler/checker.ts:33505:48 | `number` | `properties[i]` |
| src/compiler/checker.ts:33506:46 | `number` | `properties[i]` |
| src/compiler/checker.ts:33507:53 | `number` | `properties[i]` |
| src/compiler/checker.ts:35906:25 | `number` | `args[i]` |
| src/compiler/checker.ts:35935:25 | `number` | `args[argCount - 1]` |
| src/compiler/checker.ts:35953:25 | `number` | `args[i]` |
| src/compiler/checker.ts:35990:26 | `number` | `typeParameters[i]` |
| src/compiler/checker.ts:35991:61 | `number` | `typeParameters[i]` |
| src/compiler/checker.ts:35998:38 | `number` | `typeArgumentTypes[i]` |
| src/compiler/checker.ts:36003:40 | `number` | `typeArgumentNodes[i]` |
| src/compiler/checker.ts:36185:25 | `number` | `args[i]` |
| src/compiler/checker.ts:36206:60 | `number` | `args[argCount]` |
| src/compiler/checker.ts:36207:81 | `number` | `args[argCount]` |
| src/compiler/checker.ts:36207:101 | `number` | `args[args.length - 1]` |
| src/compiler/checker.ts:36291:29 | `number` | `args[i]` |
| src/compiler/checker.ts:36296:39 | `number` | `spreadType.target.elementFlags[i]` |
| src/compiler/checker.ts:36297:164 | `number` | `spreadType.target.labeledElementDeclarations?.[i]` |
| src/compiler/checker.ts:36413:44 | `number` | `args[spreadIndex]` |
| src/compiler/checker.ts:36469:31 | `number` | `closestSignature?.declaration?.parameters[closestSignature.thisParameter ? args.length + 1 : args.length]` |
| src/compiler/checker.ts:36681:34 | `number` | `candidatesForArgumentError[candidatesForArgumentError.length - 1]` |
| src/compiler/checker.ts:36727:45 | `number` | `allDiagnostics[minIndex]` |
| src/compiler/checker.ts:36809:35 | `number` | `candidates[candidateIndex]` |
| src/compiler/checker.ts:36905:51 | `number` | `s.parameters[i]` |
| src/compiler/checker.ts:36906:47 | `number` | `s.parameters[i]` |
| src/compiler/checker.ts:36954:27 | `number` | `candidates[bestIndex]` |
| src/compiler/checker.ts:36974:60 | `number` | `typeParameters[typeArguments.length]` |
| src/compiler/checker.ts:36974:130 | `number` | `typeParameters[typeArguments.length]` |
| src/compiler/checker.ts:36990:31 | `number` | `candidates[i]` |
| src/compiler/checker.ts:37225:22 | `number` | `mixinFlags[i]` |
| src/compiler/checker.ts:37947:35 | `number` | `node.arguments[i]` |
| src/compiler/checker.ts:38391:41 | `number` | `elements[index]` |
| src/compiler/checker.ts:38421:20 | `number` | `signature.parameters[pos]` |
| src/compiler/checker.ts:38423:31 | `number` | `signature.parameters[paramCount]` |
| src/compiler/checker.ts:38428:36 | `number` | `tupleType.labeledElementDeclarations?.[index]` |
| src/compiler/checker.ts:38429:34 | `number` | `tupleType.elementFlags[index]` |
| src/compiler/checker.ts:38441:27 | `number` | `signature.parameters[pos]` |
| src/compiler/checker.ts:38450:31 | `number` | `signature.parameters[paramCount]` |
| src/compiler/checker.ts:38460:36 | `number` | `associatedNames?.[index]` |
| src/compiler/checker.ts:38487:26 | `number` | `signature.parameters[pos]` |
| src/compiler/checker.ts:38490:31 | `number` | `signature.parameters[paramCount]` |
| src/compiler/checker.ts:38495:39 | `number` | `associatedNames[index]` |
| src/compiler/checker.ts:38507:39 | `number` | `signature.parameters[pos]` |
| src/compiler/checker.ts:38513:46 | `number` | `signature.parameters[paramCount]` |
| src/compiler/checker.ts:38562:46 | `number` | `signature.parameters[length - 1]` |
| src/compiler/checker.ts:38576:50 | `number` | `signature.parameters[signature.parameters.length - 1]` |
| src/compiler/checker.ts:38608:46 | `number` | `signature.parameters[signature.parameters.length - 1]` |
| src/compiler/checker.ts:38616:46 | `number` | `signature.parameters[signature.parameters.length - 1]` |
| src/compiler/checker.ts:38643:33 | `number` | `signature.parameters[i]` |
| src/compiler/checker.ts:38679:31 | `number` | `signature.parameters[i]` |
| src/compiler/checker.ts:39363:53 | `number` | `witnesses[i]` |
| src/compiler/checker.ts:40244:26 | `number` | `properties[propertyIndex]` |
| src/compiler/checker.ts:40298:17 | `number` | `node.elements[i]` |
| src/compiler/checker.ts:40308:25 | `number` | `elements[elementIndex]` |
| src/compiler/checker.ts:40594:20 | `number` | `state.typeStack[state.stackIndex]` |
| src/compiler/checker.ts:40602:20 | `number` | `state.typeStack[state.stackIndex + 1]` |
| src/compiler/checker.ts:41428:23 | `number` | `patternElements[i]` |
| src/compiler/checker.ts:41606:40 | `number` | `a[i]` |
| src/compiler/checker.ts:41606:72 | `number` | `b[i]` |
| src/compiler/checker.ts:41615:41 | `number` | `target[i]` |
| src/compiler/checker.ts:41615:78 | `number` | `source[i]` |
| src/compiler/checker.ts:41616:29 | `number` | `source[i]` |
| src/compiler/checker.ts:41691:32 | `number` | `flowTypeCache[getNodeId(node)]` |
| src/compiler/checker.ts:42061:83 | `number` | `signature.parameters[typePredicate.parameterIndex]` |
| src/compiler/checker.ts:42670:40 | `number` | `node.typeArguments[index]` |
| src/compiler/checker.ts:42672:16 | `number` | `getEffectiveTypeArguments(node, typeParameters)[index]` |
| src/compiler/checker.ts:42684:61 | `number` | `typeParameters[i]` |
| src/compiler/checker.ts:42691:21 | `number` | `typeArguments[i]` |
| src/compiler/checker.ts:42693:21 | `number` | `node.typeArguments![i]` |
| src/compiler/checker.ts:42756:43 | `number` | `signature.typeParameters?.[position]` |
| src/compiler/checker.ts:42848:47 | `number` | `typeParameters[typeArgumentPosition]` |
| src/compiler/checker.ts:44254:37 | `number` | `tags[i]` |
| src/compiler/checker.ts:46857:109 | `ModuleKind` | `ModuleKind[moduleKind]` |
| src/compiler/checker.ts:46880:36 | `number` | `jsdocParameters[lastJSDocParamIndex]` |
| src/compiler/checker.ts:46914:30 | `number` | `typeParameterDeclarations[i]` |
| src/compiler/checker.ts:46931:25 | `number` | `typeParameterDeclarations![j]` |
| src/compiler/checker.ts:46947:68 | `number` | `typeParameters[i]` |
| src/compiler/checker.ts:46995:32 | `number` | `sourceParameters[i]` |
| src/compiler/checker.ts:46996:32 | `number` | `targetParameters[i]` |
| src/compiler/checker.ts:48695:178 | `ModuleKind` | `ModuleKind[moduleKind]` |
| src/compiler/checker.ts:49308:38 | `number` | `statements[i]` |
| src/compiler/checker.ts:49318:38 | `number` | `statements[i]` |
| src/compiler/checker.ts:49326:29 | `number` | `statements[first]` |
| src/compiler/checker.ts:49327:27 | `number` | `statements[last]` |
| src/compiler/checker.ts:50922:16 | `number` | `nodeLinks[nodeId]` |
| src/compiler/checker.ts:52325:31 | `number` | `parameters[i]` |
| src/compiler/checker.ts:53798:27 | `0 \| 1` | `file?.imports[jsxImportIndex]` |
| src/compiler/checker.ts:54063:28 | `number` | `s1[i]` |
| src/compiler/checker.ts:54064:28 | `number` | `s2[i]` |
| src/compiler/checker.ts:54122:23 | `number` | `t1.elementFlags[i]` |
| src/compiler/checker.ts:54122:44 | `number` | `t2.elementFlags[i]` |
| src/compiler/checker.ts:54128:44 | `number` | `t1.labeledElementDeclarations![i]` |
| src/compiler/checker.ts:54128:79 | `number` | `t2.labeledElementDeclarations![i]` |
| src/compiler/checker.ts:54154:36 | `number` | `s1![i]` |
| src/compiler/checker.ts:54154:44 | `number` | `s2?.[i]` |
| src/compiler/commandLineParser.ts:1979:23 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:2048:26 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:2070:14 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:2074:13 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:2077:79 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:2082:38 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:2090:70 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:2094:61 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:2105:101 | `number` | `args[i]` |
| src/compiler/commandLineParser.ts:3542:30 | `number` | `(value as unknown[])[index]` |
| src/compiler/commandLineParser.ts:3551:25 | `number` | `(valueExpression as ArrayLiteralExpression \| undefined)?.elements[index]` |
| src/compiler/commandLineParser.ts:3557:114 | `number` | `(valueExpression as ArrayLiteralExpression \| undefined)?.elements[index]` |
| src/compiler/commandLineParser.ts:3887:120 | `number` | `valueExpression?.elements[index]` |
| src/compiler/commandLineParser.ts:4248:21 | `number` | `extensionGroup[i]` |
| src/compiler/core.ts:36:37 | `number` | `array[i]` |
| src/compiler/core.ts:53:37 | `number` | `array[i]` |
| src/compiler/core.ts:73:33 | `number` | `array[i]` |
| src/compiler/core.ts:110:30 | `number` | `arrayA[i]` |
| src/compiler/core.ts:110:41 | `number` | `arrayB[i]` |
| src/compiler/core.ts:128:21 | `number` | `input[i]` |
| src/compiler/core.ts:148:27 | `number` | `array[i]` |
| src/compiler/core.ts:169:23 | `number` | `array[i]` |
| src/compiler/core.ts:185:23 | `number` | `array[i]` |
| src/compiler/core.ts:201:23 | `number` | `array[i]` |
| src/compiler/core.ts:212:23 | `number` | `array[i]` |
| src/compiler/core.ts:223:34 | `number` | `array[i]` |
| src/compiler/core.ts:246:23 | `number` | `array[i]` |
| src/compiler/core.ts:281:29 | `number` | `array[i]` |
| src/compiler/core.ts:286:30 | `number` | `array[i]` |
| src/compiler/core.ts:302:15 | `number` | `array[i]` |
| src/compiler/core.ts:303:31 | `number` | `array[i]` |
| src/compiler/core.ts:325:27 | `number` | `array[i]` |
| src/compiler/core.ts:353:26 | `number` | `array[i]` |
| src/compiler/core.ts:359:35 | `number` | `array[i]` |
| src/compiler/core.ts:378:19 | `number` | `array[i]` |
| src/compiler/core.ts:403:29 | `number` | `array[i]` |
| src/compiler/core.ts:422:29 | `number` | `array[i]` |
| src/compiler/core.ts:462:26 | `number` | `array[i]` |
| src/compiler/core.ts:484:30 | `number` | `array[i]` |
| src/compiler/core.ts:498:34 | `number` | `array[i]` |
| src/compiler/core.ts:566:31 | `number` | `array[pos]` |
| src/compiler/core.ts:622:31 | `number` | `array[i]` |
| src/compiler/core.ts:642:18 | `number` | `arr[i]` |
| src/compiler/core.ts:693:16 | `number` | `array[indices[0]]` |
| src/compiler/core.ts:696:23 | `number` | `indices[i]` |
| src/compiler/core.ts:697:22 | `number` | `array[index]` |
| src/compiler/core.ts:706:34 | `number` | `array[i]` |
| src/compiler/core.ts:712:30 | `number` | `array[i]` |
| src/compiler/core.ts:741:22 | `number` | `array[i]` |
| src/compiler/core.ts:784:53 | `number` | `array[idx - 1]` |
| src/compiler/core.ts:787:64 | `number` | `array[idx]` |
| src/compiler/core.ts:824:31 | `number` | `array1[i]` |
| src/compiler/core.ts:824:42 | `number` | `array2[i]` |
| src/compiler/core.ts:850:23 | `number` | `array[i]` |
| src/compiler/core.ts:878:53 | `number` | `arrayB[offsetB]` |
| src/compiler/core.ts:878:70 | `number` | `arrayB[offsetB - 1]` |
| src/compiler/core.ts:886:57 | `number` | `arrayA[offsetA]` |
| src/compiler/core.ts:886:74 | `number` | `arrayA[offsetA - 1]` |
| src/compiler/core.ts:889:30 | `number` | `arrayB[offsetB]` |
| src/compiler/core.ts:889:47 | `number` | `arrayA[offsetA]` |
| src/compiler/core.ts:894:33 | `number` | `arrayB[offsetB]` |
| src/compiler/core.ts:991:13 | `number` | `from[i]` |
| src/compiler/core.ts:992:21 | `number` | `from[i]` |
| src/compiler/core.ts:1030:37 | `number` | `array[x]` |
| src/compiler/core.ts:1030:47 | `number` | `array[y]` |
| src/compiler/core.ts:1045:15 | `number` | `array[i]` |
| src/compiler/core.ts:1052:13 | `number` | `array1[pos]` |
| src/compiler/core.ts:1052:29 | `number` | `array2[pos]` |
| src/compiler/core.ts:1072:24 | `number` | `array[offset]` |
| src/compiler/core.ts:1117:68 | `number` | `array[array.length - 1]` |
| src/compiler/core.ts:1123:12 | `number` | `array[array.length - 1]` |
| src/compiler/core.ts:1211:36 | `number` | `array[middle]` |
| src/compiler/core.ts:1240:26 | `number` | `array[pos]` |
| src/compiler/core.ts:1247:36 | `number` | `array[pos]` |
| src/compiler/core.ts:1412:23 | `number` | `array[i]` |
| src/compiler/core.ts:1427:23 | `number` | `array[i]` |
| src/compiler/core.ts:1441:23 | `number` | `values[i]` |
| src/compiler/core.ts:1468:27 | `number` | `values[i]` |
| src/compiler/core.ts:1586:24 | `number` | `elements[headIndex]` |
| src/compiler/core.ts:1677:32 | `number` | `candidates[i]` |
| src/compiler/core.ts:1682:48 | `number` | `candidates[1 - i]` |
| src/compiler/core.ts:1739:31 | `unique symbol` | `multiMap[Symbol.toStringTag]` |
| src/compiler/core.ts:2004:38 | `number` | `arr[i]` |
| src/compiler/core.ts:2220:42 | `number` | `s1[i - 1]` |
| src/compiler/core.ts:2220:70 | `number` | `s2[j - 1]` |
| src/compiler/core.ts:2221:20 | `number` | `previous[j - 1]` |
| src/compiler/core.ts:2222:20 | `number` | `previous[j - 1]` |
| src/compiler/core.ts:2224:19 | `number` | `previous[j - 1]` |
| src/compiler/core.ts:2225:39 | `number` | `previous[j]` |
| src/compiler/core.ts:2225:67 | `number` | `current[j - 1]` |
| src/compiler/core.ts:2242:17 | `number` | `previous[s2.length]` |
| src/compiler/core.ts:2327:13 | `number` | `array[i]` |
| src/compiler/core.ts:2343:20 | `number` | `array[i + 1]` |
| src/compiler/core.ts:2350:20 | `number` | `array[array.length - 1]` |
| src/compiler/core.ts:2366:23 | `number` | `array[i]` |
| src/compiler/core.ts:2418:19 | `number` | `values[i]` |
| src/compiler/core.ts:2500:25 | `number` | `newItems[newIndex]` |
| src/compiler/core.ts:2501:25 | `number` | `oldItems[oldIndex]` |
| src/compiler/core.ts:2520:18 | `number` | `newItems[newIndex++]` |
| src/compiler/core.ts:2524:17 | `number` | `oldItems[oldIndex++]` |
| src/compiler/core.ts:2538:27 | `number` | `arrays[index]` |
| src/compiler/core.ts:2564:41 | `number` | `array[index]` |
| src/compiler/core.ts:2580:41 | `number` | `array[index]` |
| src/compiler/debug.ts:996:29 | `number` | `links[id]` |
| src/compiler/debug.ts:1059:48 | `number` | `columns[node.level]` |
| src/compiler/debug.ts:1071:35 | `number` | `children[i]` |
| src/compiler/debug.ts:1110:36 | `number` | `switchStatement.caseBlock.clauses[i]` |
| src/compiler/debug.ts:1137:17 | `number` | `grid[node.level]` |
| src/compiler/debug.ts:1140:35 | `number` | `children[i]` |
| src/compiler/debug.ts:1145:21 | `number` | `connectors[node.level][child.lane]` |
| src/compiler/debug.ts:1145:21 | `number` | `connectors[node.level]` |
| src/compiler/debug.ts:1148:21 | `number` | `connectors[node.level][node.lane]` |
| src/compiler/debug.ts:1148:21 | `number` | `connectors[node.level]` |
| src/compiler/debug.ts:1152:36 | `number` | `parents[i]` |
| src/compiler/debug.ts:1156:21 | `number` | `connectors[node.level - 1][parent.lane]` |
| src/compiler/debug.ts:1156:21 | `number` | `connectors[node.level - 1]` |
| src/compiler/debug.ts:1163:47 | `number` | `connectors[column - 1][lane]` |
| src/compiler/debug.ts:1163:47 | `number` | `connectors[column - 1]` |
| src/compiler/debug.ts:1164:46 | `number` | `connectors[column][lane - 1]` |
| src/compiler/debug.ts:1164:46 | `number` | `connectors[column]` |
| src/compiler/debug.ts:1165:37 | `number` | `connectors[column][lane]` |
| src/compiler/debug.ts:1165:37 | `number` | `connectors[column]` |
| src/compiler/debug.ts:1169:25 | `number` | `connectors[column]` |
| src/compiler/debug.ts:1176:39 | `number` | `connectors[column][lane]` |
| src/compiler/debug.ts:1176:39 | `number` | `connectors[column]` |
| src/compiler/debug.ts:1178:34 | `number` | `grid[column][lane]` |
| src/compiler/debug.ts:1178:34 | `number` | `grid[column]` |
| src/compiler/debug.ts:1181:58 | `number` | `columnWidths[column]` |
| src/compiler/debug.ts:1188:58 | `number` | `columnWidths[column]` |
| src/compiler/debug.ts:1192:98 | `number` | `grid[column + 1][lane]` |
| src/compiler/debug.ts:1192:98 | `number` | `grid[column + 1]` |
| src/compiler/debug.ts:1199:17 | `number` | `lanes[lane]` |
| src/compiler/emitter.ts:658:34 | `number` | `commonSourceDirectory[commonSourceDirectory.length - 1]` |
| src/compiler/emitter.ts:675:34 | `number` | `commonSourceDirectory[commonSourceDirectory.length - 1]` |
| src/compiler/emitter.ts:2056:42 | `number` | `bundle.sourceFiles[i]` |
| src/compiler/emitter.ts:2904:53 | `number` | `state.preserveSourceNewlinesStack[state.stackIndex]` |
| src/compiler/emitter.ts:2905:43 | `number` | `state.containerPosStack[state.stackIndex]` |
| src/compiler/emitter.ts:2906:43 | `number` | `state.containerEndStack[state.stackIndex]` |
| src/compiler/emitter.ts:2907:58 | `number` | `state.declarationListContainerEndStack[state.stackIndex]` |
| src/compiler/emitter.ts:2908:44 | `number` | `state.shouldEmitCommentsStack[state.stackIndex]` |
| src/compiler/emitter.ts:2909:46 | `number` | `state.shouldEmitSourceMapsStack[state.stackIndex]` |
| src/compiler/emitter.ts:4381:31 | `number` | `statements[i]` |
| src/compiler/emitter.ts:4474:36 | `number` | `modifiers[pos]` |
| src/compiler/emitter.ts:4734:86 | `number` | `children[start]` |
| src/compiler/emitter.ts:4754:27 | `number` | `children[start + i]` |
| src/compiler/emitter.ts:4850:86 | `number` | `children[start + count - 1]` |
| src/compiler/emitter.ts:5467:20 | `number` | `autoGeneratedIdToGeneratedName[autoGenerateId]` |
| src/compiler/emitter.ts:5474:16 | `number` | `cache[nodeId]` |
| src/compiler/emitter.ts:5503:25 | `number` | `stack[i]` |
| src/compiler/emitter.ts:5506:19 | `number` | `stack[i]` |
| src/compiler/emitter.ts:6272:68 | `SyntaxKind` | `emitNode.tokenSourceMapRanges[token]` |
| src/compiler/emitter.ts:6337:12 | `number` | `brackets[format & ListFormat.BracketsMask]` |
| src/compiler/emitter.ts:6341:12 | `number` | `brackets[format & ListFormat.BracketsMask]` |
| src/compiler/executeCommandLine.ts:439:9 | `number` | `lines[lines.length - 2]` |
| src/compiler/factory/emitHelpers.ts:466:78 | `number` | `elements[i]` |
| src/compiler/factory/emitHelpers.ts:470:34 | `number` | `computedTempVariables[computedTempVariableOffset]` |
| src/compiler/factory/emitHelpers.ts:720:23 | `number` | `input[i]` |
| src/compiler/factory/emitHelpers.ts:721:34 | `number` | `args[i]` |
| src/compiler/factory/emitHelpers.ts:723:19 | `number` | `input[input.length - 1]` |
| src/compiler/factory/emitNode.ts:148:12 | `SyntaxKind` | `node.emitNode?.tokenSourceMapRanges?.[token]` |
| src/compiler/factory/emitNode.ts:297:24 | `number` | `sourceEmitHelpers[i]` |
| src/compiler/factory/nodeFactory.ts:6897:31 | `number` | `source[statementOffset]` |
| src/compiler/factory/nodeFactory.ts:6927:31 | `number` | `source[statementOffset]` |
| src/compiler/factory/nodeFactory.ts:6966:41 | `number` | `array[i]` |
| src/compiler/factory/nodeFactory.ts:7047:42 | `number` | `statements[i]` |
| src/compiler/factory/nodeFactory.ts:7051:43 | `number` | `declarations[i]` |
| src/compiler/factory/utilities.ts:1281:48 | `number` | `userStateStack[stackIndex - 1]` |
| src/compiler/factory/utilities.ts:1282:27 | `number` | `stateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1283:54 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1295:27 | `number` | `stateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1298:41 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1298:69 | `number` | `userStateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1298:97 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1313:27 | `number` | `stateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1316:28 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1316:65 | `number` | `userStateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1316:93 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1327:27 | `number` | `stateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1330:42 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1330:71 | `number` | `userStateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1330:99 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1345:27 | `number` | `stateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1347:39 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1347:62 | `number` | `userStateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1351:30 | `number` | `stateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1352:64 | `number` | `userStateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1366:27 | `number` | `stateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1403:30 | `number` | `nodeStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1481:16 | `number` | `stateStack[stackIndex]` |
| src/compiler/factory/utilities.ts:1482:26 | `number` | `stateStack[stackIndex]` |
| src/compiler/moduleNameResolver.ts:1255:80 | `number` | `resolutionDirectory[i]` |
| src/compiler/moduleNameResolver.ts:1430:90 | `ModuleResolutionKind` | `ModuleResolutionKind[moduleResolution]` |
| src/compiler/moduleNameResolver.ts:1435:94 | `ModuleResolutionKind` | `ModuleResolutionKind[moduleResolution]` |
| src/compiler/moduleNameResolver.ts:2592:37 | `number` | `parts[i]` |
| src/compiler/moduleSpecifiers.ts:146:17 | `number` | `pattern[slash - 1]` |
| src/compiler/parser.ts:1253:16 | `SyntaxKind` | `(forEachChildTable as Record<SyntaxKind, ForEachChildFunction<any>>)[node.kind]` |
| src/compiler/parser.ts:1290:28 | `number` | `current[i]` |
| src/compiler/parser.ts:1876:35 | `number` | `sourceFile.statements[pos]` |
| src/compiler/parser.ts:1877:35 | `number` | `sourceFile.statements[start]` |
| src/compiler/parser.ts:1904:51 | `number` | `sourceFile.statements[pos]` |
| src/compiler/parser.ts:1925:35 | `number` | `sourceFile.statements[pos]` |
| src/compiler/parser.ts:1945:51 | `number` | `statements[i]` |
| src/compiler/parser.ts:1954:52 | `number` | `statements[i]` |
| src/compiler/parser.ts:6056:53 | `number` | `children[children.length - 1]` |
| src/compiler/parser.ts:9032:37 | `number` | `comments[comments.length - 1]` |
| src/compiler/parser.ts:9036:47 | `number` | `comments[comments.length - 1]` |
| src/compiler/parser.ts:10454:23 | `number` | `currentArray[currentArrayIndex]` |
| src/compiler/parser.ts:10469:35 | `number` | `currentArray[currentArrayIndex]` |
| src/compiler/parser.ts:10523:39 | `number` | `array[i]` |
| src/compiler/parser.ts:10787:26 | `number` | `pragma.args[i]` |
| src/compiler/parser.ts:10788:14 | `number` | `args[i]` |
| src/compiler/parser.ts:10794:33 | `number` | `args[i]` |
| src/compiler/path.ts:542:27 | `number` | `components[i]` |
| src/compiler/path.ts:547:21 | `number` | `reduced[reduced.length - 1]` |
| src/compiler/path.ts:912:42 | `number` | `aComponents[i]` |
| src/compiler/path.ts:912:58 | `number` | `bComponents[i]` |
| src/compiler/path.ts:986:31 | `number` | `parentComponents[i]` |
| src/compiler/path.ts:986:52 | `number` | `childComponents[i]` |
| src/compiler/path.ts:1016:52 | `number` | `fromComponents[start]` |
| src/compiler/path.ts:1017:50 | `number` | `toComponents[start]` |
| src/compiler/program.ts:357:38 | `number` | `commonPathComponents[i]` |
| src/compiler/program.ts:357:88 | `number` | `sourcePathComponents[i]` |
| src/compiler/program.ts:1154:12 | `number` | `components[i]` |
| src/compiler/program.ts:1154:29 | `number` | `components[i]` |
| src/compiler/program.ts:1155:41 | `number` | `components[i]` |
| src/compiler/program.ts:1208:29 | `number` | `file.referencedFiles[index]` |
| src/compiler/program.ts:1211:29 | `number` | `file.typeReferenceDirectives[index]` |
| src/compiler/program.ts:1212:93 | `number` | `file.typeReferenceDirectives[index]` |
| src/compiler/program.ts:1215:29 | `number` | `file.libReferenceDirectives[index]` |
| src/compiler/program.ts:1282:46 | `number` | `program!.getResolvedProjectReferences()![index]` |
| src/compiler/program.ts:1311:25 | `number` | `oldResolvedRef.commandLine.projectReferences![index]` |
| src/compiler/program.ts:1796:55 | `number` | `automaticTypeDirectiveNames[i]` |
| src/compiler/program.ts:1796:107 | `number` | `resolutions[i]` |
| src/compiler/program.ts:1798:21 | `number` | `automaticTypeDirectiveNames[i]` |
| src/compiler/program.ts:1800:21 | `number` | `resolutions[i]` |
| src/compiler/program.ts:1803:40 | `number` | `automaticTypeDirectiveNames[i]` |
| src/compiler/program.ts:1804:36 | `number` | `resolutions[i]` |
| src/compiler/program.ts:1861:38 | `number` | `parent?.commandLine.projectReferences![index]` |
| src/compiler/program.ts:1861:87 | `number` | `oldProgram!.getProjectReferences()![index]` |
| src/compiler/program.ts:2261:27 | `number` | `entries[i]` |
| src/compiler/program.ts:2315:59 | `number` | `unknownEntryIndices![index]` |
| src/compiler/program.ts:2324:32 | `number` | `(parent ? parent.commandLine.projectReferences : projectReferences)![index]` |
| src/compiler/program.ts:2981:46 | `number` | `lineStarts[line]` |
| src/compiler/program.ts:2981:64 | `number` | `lineStarts[line + 1]` |
| src/compiler/program.ts:3116:70 | `number` | `parent.modifiers[decoratorIndex]` |
| src/compiler/program.ts:3124:78 | `number` | `parent.modifiers[decoratorIndex]` |
| src/compiler/program.ts:3130:69 | `number` | `parent.modifiers[trailingDecoratorIndex]` |
| src/compiler/program.ts:3131:69 | `number` | `parent.modifiers[decoratorIndex]` |
| src/compiler/program.ts:3526:30 | `FileIncludeKind.RootFile \| FileIncludeKind.LibFile \| FileIncludeKind.AutomaticTypeDirectiveFile \| ProjectReferenceFileKind \| ReferencedFileKind` | `(FileIncludeKind as any)[reason.kind]` |
| src/compiler/program.ts:3780:25 | `number` | `file.typeReferenceDirectives[index]` |
| src/compiler/program.ts:3781:52 | `number` | `resolutions[index]` |
| src/compiler/program.ts:3923:36 | `number` | `resolutions[index]` |
| src/compiler/program.ts:3924:36 | `number` | `moduleNames[index]` |
| src/compiler/program.ts:3925:66 | `number` | `moduleNames[index]` |
| src/compiler/program.ts:3926:57 | `number` | `resolutions[index]` |
| src/compiler/program.ts:3927:81 | `number` | `resolutions[index]` |
| src/compiler/program.ts:3957:36 | `number` | `file.imports[index]` |
| src/compiler/program.ts:3957:62 | `number` | `file.imports[index]` |
| src/compiler/program.ts:4154:39 | `number` | `options.paths[key][i]` |
| src/compiler/program.ts:4374:13 | `ModuleKind` | `ModuleKind[moduleKind]` |
| src/compiler/program.ts:4378:36 | `ModuleKind` | `ModuleKind[moduleKind]` |
| src/compiler/program.ts:4383:13 | `ModuleResolutionKind` | `ModuleResolutionKind[moduleResolution]` |
| src/compiler/program.ts:4387:42 | `ModuleResolutionKind` | `ModuleResolutionKind[moduleResolution]` |
| src/compiler/program.ts:4557:54 | `ModuleKind.None \| ModuleKind.AMD \| ModuleKind.UMD \| ModuleKind.System` | `ModuleKind[options.module]` |
| src/compiler/program.ts:4593:29 | `number` | `(parent ? parent.commandLine.projectReferences : projectReferences)![index]` |
| src/compiler/program.ts:4624:121 | `number` | `initializer.elements[valueIndex]` |
| src/compiler/program.ts:4673:123 | `number` | `referencesSyntax.elements[index]` |
| src/compiler/program.ts:5191:40 | `number` | `imports[index]` |
| src/compiler/programDiagnostics.ts:192:30 | `number` | `file.libReferenceDirectives[reason.index]` |
| src/compiler/programDiagnostics.ts:365:60 | `number` | `rootNames[reason.index]` |
| src/compiler/programDiagnostics.ts:382:66 | `number` | `resolvedProjectReferences?.[reason.index]` |
| src/compiler/programDiagnostics.ts:397:25 | `number` | `referencesSyntax.elements[index]` |
| src/compiler/programDiagnostics.ts:410:122 | `number` | `options.lib![reason.index]` |
| src/compiler/resolutionCache.ts:287:10 | `number` | `pathComponents[indexAfterOsRoot]` |
| src/compiler/resolutionCache.ts:293:9 | `number` | `pathComponents[indexAfterOsRoot]` |
| src/compiler/resolutionCache.ts:333:13 | `number` | `fileOrDirComponents[i]` |
| src/compiler/resolutionCache.ts:333:40 | `number` | `dirComponents[i]` |
| src/compiler/resolutionCache.ts:426:17 | `number` | `dirPathComponents[i]` |
| src/compiler/resolutionCache.ts:426:42 | `number` | `rootPathComponents[i]` |
| src/compiler/resolutionCache.ts:451:25 | `number` | `dirPathComponents[lastNodeModulesIndex + 1]` |
| src/compiler/resolutionCache.ts:795:58 | `number` | `newFile.packageJsonLocations![i]` |
| src/compiler/resolutionCache.ts:799:61 | `number` | `existing[i]` |
| src/compiler/resolutionCache.ts:1619:86 | `number` | `locationPath[dirPath.length]` |
| src/compiler/scanner.ts:373:13 | `number` | `map[mid]` |
| src/compiler/scanner.ts:373:41 | `number` | `map[mid + 1]` |
| src/compiler/scanner.ts:377:20 | `number` | `map[mid]` |
| src/compiler/scanner.ts:415:12 | `SyntaxKind` | `tokenStrings[t]` |
| src/compiler/scanner.ts:427:12 | `RegularExpressionFlags` | `regExpFlagCharCodes[f]` |
| src/compiler/scanner.ts:486:17 | `number` | `lineStarts[line]` |
| src/compiler/scanner.ts:491:22 | `number` | `lineStarts[line + 1]` |
| src/compiler/scanner.ts:491:45 | `number` | `lineStarts[line + 1]` |
| src/compiler/scanner.ts:494:28 | `number` | `lineStarts[line + 1]` |
| src/compiler/scanner.ts:512:31 | `number` | `lineStarts[lineNumber]` |
| src/compiler/scanner.ts:1334:29 | `number` | `text[identifierStart]` |
| src/compiler/scanner.ts:1858:22 | `number` | `text[pos]` |
| src/compiler/scanner.ts:2325:38 | `number` | `text[pos + 1]` |
| src/compiler/semver.ts:171:32 | `number` | `left[i]` |
| src/compiler/semver.ts:172:33 | `number` | `right[i]` |
| src/compiler/sourcemap.ts:208:34 | `number` | `sourceIndexToNewSourceIndexMap[raw.sourceIndex]` |
| src/compiler/sourcemap.ts:211:37 | `number` | `map.sources[raw.sourceIndex]` |
| src/compiler/sourcemap.ts:215:54 | `number` | `map.sourcesContent[raw.sourceIndex]` |
| src/compiler/sourcemap.ts:216:58 | `number` | `map.sourcesContent[raw.sourceIndex]` |
| src/compiler/sourcemap.ts:224:36 | `number` | `nameIndexToNewNameIndexMap[raw.nameIndex]` |
| src/compiler/sourcemap.ts:226:92 | `number` | `map.names[raw.nameIndex]` |
| src/compiler/sourcemap.ts:380:45 | `number` | `lineStarts[line]` |
| src/compiler/sourcemap.ts:380:63 | `number` | `lineStarts[line + 1]` |
| src/compiler/sourcemap.ts:722:55 | `number` | `sourceFileAbsolutePaths[mapping.sourceIndex]` |
| src/compiler/sourcemap.ts:723:22 | `number` | `map.sources[mapping.sourceIndex]` |
| src/compiler/sourcemap.ts:759:28 | `number` | `lists[mapping.sourceIndex]` |
| src/compiler/sourcemap.ts:765:16 | `number` | `sourceMappings[sourceIndex]` |
| src/compiler/sourcemap.ts:792:25 | `number` | `sourceMappings[targetIndex]` |
| src/compiler/sourcemap.ts:810:25 | `number` | `generatedMappings[targetIndex]` |
| src/compiler/sourcemap.ts:815:28 | `number` | `sourceFileAbsolutePaths[mapping.sourceIndex]` |
| src/compiler/symbolWalker.ts:76:17 | `number` | `visitedTypes[type.id]` |
| src/compiler/symbolWalker.ts:190:17 | `number` | `visitedSymbols[symbolId]` |
| src/compiler/sys.ts:195:29 | `number` | `queue[pollIndex]` |
| src/compiler/sys.ts:215:13 | `number` | `queue[pollIndex]` |
| src/compiler/sys.ts:290:84 | `PollingInterval` | `pollingChunkSize[queue.pollingInterval]` |
| src/compiler/sys.ts:332:53 | `PollingInterval` | `unchangedPollThresholds[pollingInterval]` |
| src/compiler/sys.ts:484:73 | `PollingInterval.Low` | `pollingChunkSize[PollingInterval.Low]` |
| src/compiler/sys.ts:712:81 | `number` | `dirPath[rootDirName.length]` |
| src/compiler/sys.ts:1801:34 | `number` | `buffer[i]` |
| src/compiler/sys.ts:1802:33 | `number` | `buffer[i + 1]` |
| src/compiler/tracing.ts:122:13 | `number` | `legend[legend.length - 1]` |
| src/compiler/tracing.ts:175:66 | `number` | `eventStack[index]` |
| src/compiler/tracing.ts:220:27 | `number` | `legend[legend.length - 1]` |
| src/compiler/tracing.ts:230:26 | `number` | `types[i]` |
| src/compiler/transformer.ts:367:9 | `SyntaxKind` | `enabledSyntaxKindFeatures[kind]` |
| src/compiler/transformer.ts:374:17 | `SyntaxKind` | `enabledSyntaxKindFeatures[node.kind]` |
| src/compiler/transformer.ts:395:9 | `SyntaxKind` | `enabledSyntaxKindFeatures[kind]` |
| src/compiler/transformer.ts:403:17 | `SyntaxKind` | `enabledSyntaxKindFeatures[node.kind]` |
| src/compiler/transformer.ts:563:50 | `number` | `lexicalEnvironmentVariableDeclarationsStack[lexicalEnvironmentStackOffset]` |
| src/compiler/transformer.ts:564:50 | `number` | `lexicalEnvironmentFunctionDeclarationsStack[lexicalEnvironmentStackOffset]` |
| src/compiler/transformer.ts:565:40 | `number` | `lexicalEnvironmentStatementsStack[lexicalEnvironmentStackOffset]` |
| src/compiler/transformer.ts:566:35 | `number` | `lexicalEnvironmentFlagsStack[lexicalEnvironmentStackOffset]` |
| src/compiler/transformer.ts:614:43 | `number` | `blockScopedVariableDeclarationsStack[blockScopeStackOffset]` |
| src/compiler/transformers/classFields.ts:2250:37 | `number` | `superPath[superPathDepth]` |
| src/compiler/transformers/classFields.ts:2251:32 | `number` | `statementsIn[superStatementIndex]` |
| src/compiler/transformers/classFields.ts:2296:35 | `number` | `statementsIn[statementOffset]` |
| src/compiler/transformers/classFields.ts:2375:39 | `number` | `constructor.body.statements[statementOffset]` |
| src/compiler/transformers/classFields.ts:3303:40 | `number` | `classAliases[declaration.id!]` |
| src/compiler/transformers/declarations/diagnostics.ts:723:58 | `SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor` | `errorByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:726:71 | `SyntaxKind.SetAccessor` | `relatedSuggestionByDeclarationKind[setAccessor.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:729:71 | `SyntaxKind.GetAccessor` | `relatedSuggestionByDeclarationKind[getAccessor.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:737:77 | `SyntaxKind.Parameter \| SyntaxKind.PropertyDeclaration \| SyntaxKind.MethodDeclaration \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.FunctionExpression \| SyntaxKind.ArrowFunction \| SyntaxKind.VariableDeclaration \| SyntaxKind.FunctionDeclaration \| SyntaxKind.ExportAssignment` | `relatedSuggestionByDeclarationKind[parentDeclaration.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:742:52 | `SyntaxKind.ComputedPropertyName \| SyntaxKind.ShorthandPropertyAssignment \| SyntaxKind.SpreadAssignment` | `errorByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:747:52 | `SyntaxKind.ArrayLiteralExpression \| SyntaxKind.SpreadElement` | `errorByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:752:52 | `SyntaxKind.MethodDeclaration \| SyntaxKind.ConstructSignature \| SyntaxKind.FunctionExpression \| SyntaxKind.ArrowFunction \| SyntaxKind.FunctionDeclaration` | `errorByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:754:60 | `SyntaxKind.MethodDeclaration \| SyntaxKind.ConstructSignature \| SyntaxKind.FunctionExpression \| SyntaxKind.ArrowFunction \| SyntaxKind.FunctionDeclaration` | `relatedSuggestionByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:761:52 | `SyntaxKind.PropertyDeclaration \| SyntaxKind.VariableDeclaration` | `errorByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:763:60 | `SyntaxKind.PropertyDeclaration \| SyntaxKind.VariableDeclaration` | `relatedSuggestionByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:776:13 | `SyntaxKind.Parameter` | `errorByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:779:60 | `SyntaxKind.Parameter` | `relatedSuggestionByDeclarationKind[node.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:798:75 | `SyntaxKind.Parameter \| SyntaxKind.PropertyDeclaration \| SyntaxKind.MethodDeclaration \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.FunctionExpression \| SyntaxKind.ArrowFunction \| SyntaxKind.VariableDeclaration \| SyntaxKind.FunctionDeclaration \| SyntaxKind.ExportAssignment` | `errorByDeclarationKind[parentDeclaration.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:799:81 | `SyntaxKind.Parameter \| SyntaxKind.PropertyDeclaration \| SyntaxKind.MethodDeclaration \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.FunctionExpression \| SyntaxKind.ArrowFunction \| SyntaxKind.VariableDeclaration \| SyntaxKind.FunctionDeclaration \| SyntaxKind.ExportAssignment` | `relatedSuggestionByDeclarationKind[parentDeclaration.kind]` |
| src/compiler/transformers/declarations/diagnostics.ts:803:81 | `SyntaxKind.Parameter \| SyntaxKind.PropertyDeclaration \| SyntaxKind.MethodDeclaration \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.FunctionExpression \| SyntaxKind.ArrowFunction \| SyntaxKind.VariableDeclaration \| SyntaxKind.FunctionDeclaration \| SyntaxKind.ExportAssignment` | `relatedSuggestionByDeclarationKind[parentDeclaration.kind]` |
| src/compiler/transformers/destructuring.ts:400:25 | `number` | `elements[i]` |
| src/compiler/transformers/destructuring.ts:456:85 | `number` | `elements[numElements - 1]` |
| src/compiler/transformers/destructuring.ts:482:25 | `number` | `elements[i]` |
| src/compiler/transformers/es2015.ts:931:59 | `number` | `outParams[i]` |
| src/compiler/transformers/es2015.ts:1459:31 | `number` | `body.statements[i]` |
| src/compiler/transformers/es2015.ts:1476:35 | `number` | `body.statements[superCallIndex]` |
| src/compiler/transformers/es2015.ts:1494:31 | `number` | `body.statements[superCallIndex]` |
| src/compiler/transformers/es2015.ts:1573:31 | `number` | `body.statements[i]` |
| src/compiler/transformers/es2015.ts:1575:35 | `number` | `body.statements[i - 1]` |
| src/compiler/transformers/es2015.ts:2715:29 | `number` | `node.elements[i]` |
| src/compiler/transformers/es2015.ts:3317:30 | `number` | `properties[i]` |
| src/compiler/transformers/es2015.ts:4096:30 | `number` | `properties[i]` |
| src/compiler/transformers/es2015.ts:4430:41 | `number` | `funcStatements[classBodyStart]` |
| src/compiler/transformers/es2015.ts:4437:29 | `number` | `funcStatements[classBodyStart]` |
| src/compiler/transformers/es2015.ts:4695:29 | `number` | `segments[i]` |
| src/compiler/transformers/es2017.ts:762:47 | `number` | `node.parameters[i]` |
| src/compiler/transformers/es2017.ts:763:44 | `number` | `outerParameters[i]` |
| src/compiler/transformers/es2017.ts:940:42 | `number` | `substitutedSuperAccessors[getNodeId(node)]` |
| src/compiler/transformers/es2018.ts:540:80 | `number` | `objects[i]` |
| src/compiler/transformers/es2018.ts:632:29 | `number` | `node.elements[i]` |
| src/compiler/transformers/es2018.ts:1374:42 | `number` | `substitutedSuperAccessors[getNodeId(node)]` |
| src/compiler/transformers/es2020.ts:179:29 | `number` | `chain[i]` |
| src/compiler/transformers/esDecorators.ts:1163:37 | `number` | `superPath[superPathDepth]` |
| src/compiler/transformers/esDecorators.ts:1164:32 | `number` | `statementsIn[superStatementIndex]` |
| src/compiler/transformers/esnext.ts:182:35 | `number` | `node.statements[pos]` |
| src/compiler/transformers/esnext.ts:343:31 | `number` | `statementsIn[i]` |
| src/compiler/transformers/esnext.ts:767:34 | `number` | `statements[i]` |
| src/compiler/transformers/esnext.ts:767:70 | `number` | `statements[i]` |
| src/compiler/transformers/generators.ts:1283:39 | `number` | `statements[i]` |
| src/compiler/transformers/generators.ts:1369:34 | `number` | `variables[i]` |
| src/compiler/transformers/generators.ts:1864:32 | `number` | `caseBlock.clauses[i]` |
| src/compiler/transformers/generators.ts:1879:36 | `number` | `caseBlock.clauses[i]` |
| src/compiler/transformers/generators.ts:1889:55 | `number` | `clauseLabels[i]` |
| src/compiler/transformers/generators.ts:1911:27 | `number` | `clauseLabels[defaultClauseIndex]` |
| src/compiler/transformers/generators.ts:1918:27 | `number` | `clauseLabels[i]` |
| src/compiler/transformers/generators.ts:1919:44 | `number` | `caseBlock.clauses[i]` |
| src/compiler/transformers/generators.ts:2044:31 | `number` | `nodes[i]` |
| src/compiler/transformers/generators.ts:2073:34 | `number` | `renamedCatchVariableDeclarations[getOriginalNodeId(declaration)]` |
| src/compiler/transformers/generators.ts:2445:37 | `number` | `blockStack![j]` |
| src/compiler/transformers/generators.ts:2468:35 | `number` | `blockStack[i]` |
| src/compiler/transformers/generators.ts:2479:35 | `number` | `blockStack[i]` |
| src/compiler/transformers/generators.ts:2498:35 | `number` | `blockStack[i]` |
| src/compiler/transformers/generators.ts:2506:35 | `number` | `blockStack[i]` |
| src/compiler/transformers/generators.ts:2528:17 | `number` | `labelExpressions[label]` |
| src/compiler/transformers/generators.ts:2532:17 | `number` | `labelExpressions[label]` |
| src/compiler/transformers/generators.ts:2854:17 | `number` | `labelOffsets[label]` |
| src/compiler/transformers/generators.ts:2854:59 | `number` | `labelExpressions[label]` |
| src/compiler/transformers/generators.ts:2879:39 | `number` | `withBlockStack[i]` |
| src/compiler/transformers/generators.ts:2942:17 | `number` | `labelOffsets[label]` |
| src/compiler/transformers/generators.ts:2947:21 | `number` | `labelNumbers[labelNumber]` |
| src/compiler/transformers/generators.ts:2951:21 | `number` | `labelNumbers[labelNumber]` |
| src/compiler/transformers/generators.ts:2963:32 | `number` | `labelNumbers[labelNumber]` |
| src/compiler/transformers/generators.ts:2966:45 | `number` | `labelExpressions[label]` |
| src/compiler/transformers/generators.ts:2983:57 | `number` | `blockOffsets![blockIndex]` |
| src/compiler/transformers/generators.ts:2984:42 | `number` | `blocks[blockIndex]` |
| src/compiler/transformers/generators.ts:2985:37 | `number` | `blockActions![blockIndex]` |
| src/compiler/transformers/generators.ts:3039:24 | `number` | `operations![operationIndex]` |
| src/compiler/transformers/generators.ts:3047:22 | `number` | `operationArguments![operationIndex]` |
| src/compiler/transformers/generators.ts:3052:26 | `number` | `operationLocations![operationIndex]` |
| src/compiler/transformers/legacyDecorators.ts:688:44 | `number` | `classAliases[getOriginalNodeId(node)]` |
| src/compiler/transformers/legacyDecorators.ts:833:40 | `number` | `classAliases[declaration.id!]` |
| src/compiler/transformers/module/module.ts:2262:33 | `number` | `moduleInfoMap[getOriginalNodeId(currentSourceFile)]` |
| src/compiler/transformers/module/module.ts:2286:24 | `number` | `noSubstitution[node.id]` |
| src/compiler/transformers/module/module.ts:2464:24 | `number` | `currentModuleInfo?.exportedBindings[getOriginalNodeId(importDeclaration)]` |
| src/compiler/transformers/module/module.ts:2473:38 | `number` | `currentModuleInfo?.exportedBindings[getOriginalNodeId(declaration)]` |
| src/compiler/transformers/module/system.ts:289:21 | `number` | `dependencyGroups[groupIndex]` |
| src/compiler/transformers/module/system.ts:1773:26 | `number` | `moduleInfoMap[id]` |
| src/compiler/transformers/module/system.ts:1774:30 | `number` | `exportFunctionsMap[id]` |
| src/compiler/transformers/module/system.ts:1775:30 | `number` | `noSubstitutionMap[id]` |
| src/compiler/transformers/module/system.ts:1776:29 | `number` | `contextObjectMap[id]` |
| src/compiler/transformers/module/system.ts:1995:53 | `number` | `moduleInfo?.exportedBindings[getOriginalNodeId(valueDeclaration)]` |
| src/compiler/transformers/module/system.ts:2016:37 | `number` | `moduleInfo?.exportedBindings[getOriginalNodeId(valueDeclaration)]` |
| src/compiler/transformers/module/system.ts:2023:61 | `number` | `moduleInfo?.exportedBindings[getOriginalNodeId(declaration)]` |
| src/compiler/transformers/module/system.ts:2048:45 | `number` | `noSubstitution[node.id]` |
| src/compiler/transformers/ts.ts:1350:37 | `number` | `superPath[superPathDepth]` |
| src/compiler/transformers/ts.ts:1351:32 | `number` | `statementsIn[superStatementIndex]` |
| src/compiler/transformers/typeSerializer.ts:212:35 | `number` | `parameters[i]` |
| src/compiler/transformers/utilities.ts:378:18 | `number` | `map[key]` |
| src/compiler/transformers/utilities.ts:549:27 | `number` | `statements[i]` |
| src/compiler/transformers/utilities.ts:660:31 | `number` | `parameters[i + firstParameterOffset]` |
| src/compiler/tsbuildPublic.ts:737:37 | `number` | `buildOrder[buildOrder.length - 1]` |
| src/compiler/tsbuildPublic.ts:1213:25 | `number` | `buildOrder[projectIndex]` |
| src/compiler/tsbuildPublic.ts:1905:29 | `number` | `buildOrder[index]` |
| src/compiler/types.ts:7321:18 | `DiagnosticCategory` | `DiagnosticCategory[d.category]` |
| src/compiler/utilities.ts:939:31 | `number` | `newResolutions[i]` |
| src/compiler/utilities.ts:940:23 | `number` | `names[i]` |
| src/compiler/utilities.ts:1017:12 | `number` | `getLineStarts(sourceFile)[line]` |
| src/compiler/utilities.ts:1041:23 | `number` | `lineStarts[lineIndex]` |
| src/compiler/utilities.ts:1043:19 | `number` | `lineStarts[lineIndex + 1]` |
| src/compiler/utilities.ts:1121:34 | `number` | `to[statementIndex]` |
| src/compiler/utilities.ts:1134:34 | `number` | `to[statementIndex]` |
| src/compiler/utilities.ts:6086:20 | `number` | `diagnostics[result]` |
| src/compiler/utilities.ts:6088:68 | `number` | `diagnostics[~result - 1]` |
| src/compiler/utilities.ts:6089:20 | `number` | `diagnostics[~result - 1]` |
| src/compiler/utilities.ts:6288:28 | `number` | `indentStrings[current - 1]` |
| src/compiler/utilities.ts:6290:12 | `number` | `indentStrings[level]` |
| src/compiler/utilities.ts:6767:16 | `0 \| 1` | `accessor.parameters[hasThis ? 1 : 0]` |
| src/compiler/utilities.ts:7090:19 | `number` | `lineMap[currentLine + 1]` |
| src/compiler/utilities.ts:7095:68 | `number` | `lineMap[firstCommentLineAndCharacter.line]` |
| src/compiler/utilities.ts:7684:17 | `number` | `charCodes[i]` |
| src/compiler/utilities.ts:7685:18 | `number` | `charCodes[i]` |
| src/compiler/utilities.ts:7685:52 | `number` | `charCodes[i + 1]` |
| src/compiler/utilities.ts:7686:18 | `number` | `charCodes[i + 1]` |
| src/compiler/utilities.ts:7686:56 | `number` | `charCodes[i + 2]` |
| src/compiler/utilities.ts:7687:17 | `number` | `charCodes[i + 2]` |
| src/compiler/utilities.ts:7712:26 | `number` | `codes[i]` |
| src/compiler/utilities.ts:7721:36 | `number` | `codes[i]` |
| src/compiler/utilities.ts:7725:28 | `number` | `codes[i]` |
| src/compiler/utilities.ts:7762:42 | `number` | `input[i]` |
| src/compiler/utilities.ts:7763:42 | `number` | `input[i + 1]` |
| src/compiler/utilities.ts:7764:42 | `number` | `input[i + 2]` |
| src/compiler/utilities.ts:7765:42 | `number` | `input[i + 3]` |
| src/compiler/utilities.ts:8302:31 | `number` | `children[i]` |
| src/compiler/utilities.ts:8303:29 | `number` | `children[i]` |
| src/compiler/utilities.ts:8584:90 | `number` | `args[+index]` |
| src/compiler/utilities.ts:8806:25 | `number` | `d2.relatedInformation![index]` |
| src/compiler/utilities.ts:8889:39 | `number` | `c1[i]` |
| src/compiler/utilities.ts:8889:51 | `number` | `c2[i]` |
| src/compiler/utilities.ts:8905:43 | `number` | `c1[i]` |
| src/compiler/utilities.ts:8905:62 | `number` | `c2[i]` |
| src/compiler/utilities.ts:8909:13 | `number` | `c1[i]` |
| src/compiler/utilities.ts:8912:42 | `number` | `c1[i]` |
| src/compiler/utilities.ts:8912:55 | `number` | `c2[i]` |
| src/compiler/utilities.ts:9414:69 | `number` | `jsxImportSourcePragmas[jsxImportSourcePragmas.length - 1]` |
| src/compiler/utilities.ts:9416:59 | `number` | `jsxRuntimePragmas[jsxRuntimePragmas.length - 1]` |
| src/compiler/utilities.ts:9564:48 | `number` | `aParts[aParts.length - 2]` |
| src/compiler/utilities.ts:9565:48 | `number` | `bParts[bParts.length - 2]` |
| src/compiler/utilities.ts:9566:30 | `number` | `aParts[aParts.length - 1]` |
| src/compiler/utilities.ts:9566:82 | `number` | `bParts[bParts.length - 1]` |
| src/compiler/utilities.ts:9864:21 | `number` | `results[includeIndex]` |
| src/compiler/utilities.ts:10367:32 | `number` | `arr[i]` |
| src/compiler/utilities.ts:10502:9 | `number` | `segments[segment]` |
| src/compiler/utilities.ts:10504:23 | `number` | `segments[segment + 1]` |
| src/compiler/utilities.ts:10514:46 | `number` | `segments[segment]` |
| src/compiler/utilities.ts:10633:24 | `number` | `array[i]` |
| src/compiler/utilitiesPublic.ts:431:20 | `number` | `spans[i]` |
| src/compiler/utilitiesPublic.ts:433:73 | `number` | `spans[j]` |
| src/compiler/utilitiesPublic.ts:434:48 | `number` | `spans[j]` |
| src/compiler/utilitiesPublic.ts:435:65 | `number` | `spans[j]` |
| src/compiler/utilitiesPublic.ts:505:28 | `number` | `changes[i]` |
| src/compiler/utilitiesPublic.ts:1047:25 | `number` | `paramTags[i]` |
| src/compiler/utilitiesPublic.ts:2653:48 | `number` | `(parseTreeNode.parent as SignatureDeclaration).parameters[paramIdx - 1]` |
| src/compiler/visitorPublic.ts:345:22 | `number` | `nodes[i + start]` |
| src/compiler/visitorPublic.ts:423:27 | `number` | `parameters[i]` |
| src/compiler/visitorPublic.ts:603:16 | `SyntaxKind` | `(visitEachChildTable as Record<SyntaxKind, VisitEachChildFunction<any> \| undefined>)[node.kind]` |
| src/compiler/watch.ts:423:27 | `number` | `configFile.configFileSpecs.validatedFilesSpecBeforeSubstitution![index]` |
| src/compiler/watch.ts:442:27 | `number` | `configFile.configFileSpecs.validatedIncludeSpecsBeforeSubstitution![index]` |
| src/compiler/watch.ts:498:56 | `number` | `program.getRootFileNames()[reason.index]` |
| src/compiler/watch.ts:519:62 | `number` | `program.getResolvedProjectReferences()?.[reason.index]` |
| src/compiler/watch.ts:544:151 | `number` | `options.lib![reason.index]` |

## Scope and assumptions

This enumerates executable AST in every original src/compiler TypeScript file, including any/generic/finite mapped receivers and nested accesses. It excludes literal element reads, write/delete-only accesses, generated diagnostic code, static dot-property reads, Map.get calls, comments, and JavaScript stored in emit-helper template strings. Emitted JavaScript belongs to compiled programs, not to dynamic reads performed by tsc itself. The lower grep count of for-in text also includes such strings and comments; the AST finds 27 actual loops.

Unknown is a complete classification, not a missing census row: generic core helpers, clone helpers, arbitrary caller-supplied diagnostic keys, environment-name callers, and any-valued object comparisons have no proven closed key origin. Fixed-table review describes the pinned internal tables without external mutation. Key classification does not imply that an inherited miss is reachable: own guards, enumeration, and record construction can make the lookup safe.

Reproduce using count-inherited-keys.cjs and summarize-inherited-keys.py as documented in README.md. inherited-key-provenance.json records explicit source-review overrides; inherited-key-global.json includes hashes, excluded sites, all 68 census links, loop contexts, zero stock diagnostics, and global totals.

## Own-property review: all 36 user-input and 43 unknown sites

The runtime question is a **missing own key** whose spelling names an Object.prototype member. An own toString is an ordinary hit. Provenance and ownership are independent; the original 36/43 classifications remain unchanged.

**User input (36):** 21 reads with explicit or construction/enumeration ownership proof, 11 own-check operations, two ordinary-plain-record enumeration reads, and two reads with prototype-excluding key domains. No demonstrated user-input read-first prototype miss.

**Unknown (43):** 11 reads with ownership proof, 17 own-check operations, six ordinary-plain-record enumeration reads, four compiler-call key domains excluding prototype names, two unguarded src[e] reads with fixtures below, and three unresolved reads. No unknown site is silently counted as guarded merely because its value is tested afterward.

Verdicts: **a** proves ownership before the value read; **a-check** is the own-property check itself (not a value read); **a-plain** proves ownership only for the ordinary unmodified prototypes used at this compiler site, because default prototype members are non-enumerable. It is not an explicit guard and does not cover arbitrary enumerable inherited properties. **safe-domain** excludes the missing prototype-name case through compiler-call key constraints, even though an ordinary missing key may still be read. These last two qualifiers are necessary: the proposed a/b/c partition alone cannot truthfully describe every operation.

**b** reads before rejection: native cannot use an unconditional loud dictionary lookup. The exact sites needing has_own plus get_own treatment are **utilities.ts:8154:45** and **utilities.ts:8159:28**, both src[e] in compareDataObjects. Select the reject/false branch on the tested inherited misses before fetching an own value. Blindly substituting undefined is not a general semantics proof for any-valued objects. At 8154, empty objects/arrays instead give a **c** helper/API bug, so that site has the input-dependent verdict **b/c**. No full tsc CLI/watch wrong behavior from that helper was demonstrated; the two invalid empty-array configs correctly report TS5066. A blanket b lowering for all calls would conceal that distinction.

The three **unresolved** operations are both compareProperties reads (no compiler caller of the exported core helper found) and the exotic process.env read (stock API returns a function for a missing toString, but compiler callers use internal environment names). No real tsconfig/package/command-line input reaching those sites with the required key was established. No synthetic helper call is presented as a stock CLI reproduction.

| Location | Origin | Read/check | Verdict | Ownership evidence or limitation | Fixture |
| --- | --- | --- | --- | --- | --- |
| src/compiler/commandLineParser.ts:2789:13 | user_input | `hasProperty(options, name)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/commandLineParser.ts:2795:27 | user_input | `options[name]` | a | hasProperty(options, name) dominates the read at 2789. | - |
| src/compiler/commandLineParser.ts:2901:42 | user_input | `options[allSetOptions[0]]` | a | allSetOptions comes from Object.keys(options) at 2839; splice removes names but does not create missing names. | - |
| src/compiler/commandLineParser.ts:2930:24 | user_input | `hasProperty(options, setting)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/commandLineParser.ts:2933:24 | user_input | `options[setting]` | safe-domain | User-supplied settings come from allSetOptions and are own. Absent preset settings are fixed literal option names, none prototype members. hasProperty at 2930 is NOT a dominating guard on this read. | - |
| src/compiler/commandLineParser.ts:2976:13 | user_input | `hasProperty(options, name)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/commandLineParser.ts:2979:17 | user_input | `options[name]` | a | hasProperty(options, name) dominates at 2976. | - |
| src/compiler/commandLineParser.ts:3322:22 | user_input | `mapLike[key]` | a | getOwnKeys(mapLike) at 3320 supplies own keys; the isArray test is AFTER the read but is not the ownership proof. | - |
| src/compiler/commandLineParser.ts:3323:77 | user_input | `mapLike[key]` | a | Same getOwnKeys(mapLike) enumeration; not merely the preceding isArray test. | - |
| src/compiler/commandLineParser.ts:3785:90 | user_input | `jsonOptions[id]` | safe-domain | optionsNameMap.get(id) must succeed at 3783-3784. Its internal declarations exclude prototype names. This Map checks the option-name domain, NOT jsonOptions ownership. | - |
| src/compiler/commandLineParser.ts:4142:68 | user_input | `wildcardDirectories[existingPath]` | a | wildCardKeyToPath.get(key) yields existingPath only after writing wildcardDirectories[path], at 4144-4145. No deletion occurs until the later pruning loop. A Map lookup alone would not otherwise prove ownership. | - |
| src/compiler/commandLineParser.ts:4155:17 | user_input | `hasProperty(wildcardDirectories, path)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/commandLineParser.ts:4266:13 | user_input | `hasProperty(opts, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1267:12 | unknown | `hasOwnProperty.call(map, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1279:12 | unknown | `hasOwnProperty.call(map, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1279:44 | unknown | `map[key]` | a | hasOwnProperty.call(map, key) is the conditional before map[key]. | - |
| src/compiler/core.ts:1290:13 | unknown | `hasOwnProperty.call(map, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1315:13 | unknown | `hasOwnProperty.call(collection, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1316:25 | unknown | `(collection as MapLike<T>)[key]` | a | hasOwnProperty.call(collection, key) dominates at 1315. | - |
| src/compiler/core.ts:1354:17 | unknown | `hasProperty(arg, p)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1355:24 | unknown | `arg[p]` | a | hasProperty(arg, p) dominates at 1354. | - |
| src/compiler/core.ts:1374:13 | unknown | `hasOwnProperty.call(left, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1375:18 | unknown | `hasOwnProperty.call(right, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1376:35 | unknown | `left[key]` | a | left ownership at 1374 and right ownership at 1375 both dominate the two reads. | - |
| src/compiler/core.ts:1376:46 | unknown | `right[key]` | a | left ownership at 1374 and right ownership at 1375 both dominate the two reads. | - |
| src/compiler/core.ts:1381:13 | unknown | `hasOwnProperty.call(right, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1382:18 | unknown | `hasOwnProperty.call(left, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1470:27 | unknown | `result[key]` | safe-domain | Only compiler caller is legacyDecorators.ts:538 with isSyntheticMetadataDecorator: keys true/false. Generic helper with constructor crashes; existing 14_group_by.a records this, but no stock CLI path to that callback key was found. | - |
| src/compiler/core.ts:1481:13 | unknown | `hasOwnProperty.call(object, id)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1482:26 | unknown | `(object as any)[id]` | a | hasOwnProperty.call(object, id) dominates at 1481. | - |
| src/compiler/core.ts:1498:13 | unknown | `hasOwnProperty.call(second, id)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1499:35 | unknown | `(second as any)[id]` | a | hasOwnProperty.call(second, id) dominates at 1498. | - |
| src/compiler/core.ts:1504:13 | unknown | `hasOwnProperty.call(first, id)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1505:35 | unknown | `(first as any)[id]` | a | hasOwnProperty.call(first, id) dominates at 1504. | - |
| src/compiler/core.ts:1515:13 | unknown | `hasOwnProperty.call(second, id)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/core.ts:1516:34 | unknown | `second[id]` | a | hasOwnProperty.call(second, id) dominates at 1515. | - |
| src/compiler/core.ts:2143:18 | unknown | `a[key]` | unresolved | Generic comparer sees both unguarded values; no compiler caller of this exported helper was found. checker.ts has a different local function of the same name. No real CLI/config counterexample established. | - |
| src/compiler/core.ts:2143:26 | unknown | `b[key]` | unresolved | Generic comparer sees both unguarded values; no compiler caller of this exported helper was found. checker.ts has a different local function of the same name. No real CLI/config counterexample established. | - |
| src/compiler/debug.ts:433:27 | unknown | `enumObject[name]` | a-plain | for-in over compiler enum objects: default prototype members are non-enumerable, so a missing such key is not produced. No explicit hasOwn check. With arbitrary inherited enumerable enum properties this conclusion would not hold. | - |
| src/compiler/factory/nodeFactory.ts:6138:17 | unknown | `hasProperty(node, p)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/factory/nodeFactory.ts:6138:42 | unknown | `hasProperty(source, p)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/factory/nodeFactory.ts:6145:32 | unknown | `(source as any)[p]` | a | The continue at 6138 excludes !hasProperty(source, p) before reading source. | - |
| src/compiler/factory/nodeFactory.ts:6400:17 | unknown | `hasProperty(clone, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/factory/nodeFactory.ts:6400:45 | unknown | `hasProperty(node, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/factory/nodeFactory.ts:6404:26 | unknown | `node[key]` | a | The continue at 6400 excludes !hasProperty(node, key). | - |
| src/compiler/factory/nodeFactory.ts:7539:27 | unknown | `sourceRanges[key]` | a-plain | for-in over internally constructed arrays; ordinary Object.prototype members are non-enumerable and do not produce a missing prototype-name key. Arbitrary prototype mutation is outside this proof. | - |
| src/compiler/moduleNameResolver.ts:433:17 | user_input | `hasProperty(typesVersions, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/moduleNameResolver.ts:465:14 | user_input | `hasProperty(typesVersions, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/moduleNameResolver.ts:474:43 | user_input | `typesVersions[key]` | a | !hasProperty(typesVersions, key) continues at 465 before semver parsing and this read. | - |
| src/compiler/moduleNameResolver.ts:953:13 | user_input | `hasProperty(value, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/moduleNameResolver.ts:954:59 | user_input | `(value as any)[key]` | a | hasProperty(value, key) dominates at 953. | - |
| src/compiler/moduleNameResolver.ts:2295:46 | user_input | `(exports as MapLike<unknown>)[key]` | a-plain | Package content is JSON.parse data with the ordinary prototype. Its inherited prototype members are non-enumerable; for-in gives own keys here. allKeysStartWithDot itself checks getOwnKeys; it is NOT an ownership guard on the subsequent for-in loop. | - |
| src/compiler/moduleNameResolver.ts:2351:54 | user_input | `(target as MapLike<unknown>)[key]` | a | getOwnKeys(target) at 2349 supplies the key. | - |
| src/compiler/moduleNameResolver.ts:2429:13 | user_input | `hasProperty(peerDependencies, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/moduleNameResolver.ts:2711:83 | user_input | `hasProperty(lookupTable, moduleName)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/moduleNameResolver.ts:2712:24 | user_input | `(lookupTable as MapLike<unknown>)[moduleName]` | a | hasProperty(lookupTable, moduleName) dominates at 2711, before any read. | - |
| src/compiler/moduleNameResolver.ts:2718:28 | user_input | `(lookupTable as MapLike<unknown>)[potentialTarget]` | a | expandingKeys is filtered/sorted getOwnKeys(lookupTable) at 2715. | - |
| src/compiler/moduleNameResolver.ts:2724:28 | user_input | `(lookupTable as MapLike<unknown>)[potentialTarget]` | a | Same getOwnKeys(lookupTable) list at 2715; suffix/prefix tests do not create keys. | - |
| src/compiler/moduleNameResolver.ts:2729:28 | user_input | `(lookupTable as MapLike<unknown>)[potentialTarget]` | a | Same getOwnKeys(lookupTable) list at 2715. | - |
| src/compiler/moduleNameResolver.ts:2820:43 | user_input | `(target as MapLike<unknown>)[condition]` | a | getOwnKeys(target) at 2817 precedes the condition test and read. A supplied own toString/custom condition reads normally. | - |
| src/compiler/moduleNameResolver.ts:3177:34 | user_input | `paths[matchedPatternText]` | a | matchedPattern comes from tryParsePatterns(paths), which uses getOwnKeys at utilities.ts:10239 and a Set of exact own strings. Missing constructor does not match an empty paths record. | - |
| src/compiler/moduleSpecifiers.ts:928:35 | user_input | `paths[key]` | a-plain | for-in over parsed paths/JSON package typesVersions; default prototype members are non-enumerable. No explicit own check. Injected enumerable inherited keys would need a guard; no missing default-member read on ordinary records. | - |
| src/compiler/moduleSpecifiers.ts:1111:35 | user_input | `(exports as MapLike<unknown>)[key]` | a | getOwnKeys(exports) at 1109 supplies the key. | - |
| src/compiler/moduleSpecifiers.ts:1134:122 | user_input | `(exports as MapLike<unknown>)[k]` | a | getOwnKeys(exports) at 1129 supplies k. | - |
| src/compiler/moduleSpecifiers.ts:1165:121 | user_input | `(imports as MapLike<unknown>)[k]` | a | getOwnKeys(imports) at 1160 supplies k before the hash-prefix filter. | - |
| src/compiler/program.ts:4142:22 | user_input | `hasProperty(options.paths, key)` | a-check | This operation IS an own-property test, not a value read. Missing prototype names return false without fetching their inherited value. | - |
| src/compiler/program.ts:4148:29 | user_input | `options.paths[key]` | a | !hasProperty(options.paths, key) continues at 4142. | - |
| src/compiler/program.ts:4149:33 | user_input | `options.paths[key]` | a | Same dominating ownership check at 4142, then isArray at 4148. | - |
| src/compiler/program.ts:4154:39 | user_input | `options.paths[key]` | a | Same dominating ownership check; outer numeric array read is a separate census site. | - |
| src/compiler/program.ts:5099:53 | unknown | `option[d.skippedOn]` | safe-domain | Checker errorSkippedOn assigns only skipLibCheck or skipDefaultLibCheck in compiler calls. Diagnostic API callers can forge metadata; no CLI input changes this key domain. | - |
| src/compiler/sys.ts:1566:24 | unknown | `process.env[name]` | unresolved | process.env[name] has no own check; stock ts.sys.getEnvironmentVariable("toString") actually returns a function on Node. Compiler callers use internal environment-variable names. No requested CLI/config source can set name to a prototype member; process.env is also an exotic host object, not a plain record. | - |
| src/compiler/utilities.ts:8153:20 | unknown | `dst[e]` | a-plain | e comes from for-in on dst: missing default prototype members are non-enumerable. No explicit own guard. Arbitrary inherited enumerable dst properties are not covered. | - |
| src/compiler/utilities.ts:8154:37 | unknown | `dst[e]` | a-plain | e is enumerated from ordinary dst; default inherited prototype members are non-enumerable. No explicit own check. | - |
| src/compiler/utilities.ts:8154:45 | unknown | `src[e]` | b/c | dst[e] is a-plain. src[e] has no own guard: b for nonempty object/array values (recursive comparison rejects the inherited value), c at helper/API level for empty objects/arrays (it can incorrectly return true). Full CLI/watch wrong behavior not established. | [17_compare_missing_object.a](17_compare_missing_object.a), [19_compare_empty_objects.a](19_compare_empty_objects.a) |
| src/compiler/utilities.ts:8158:25 | unknown | `dst[e]` | a-plain | Same for-in ownership proof for ordinary dst. typeof is AFTER the read and is not an own-property guard. | - |
| src/compiler/utilities.ts:8159:17 | unknown | `dst[e]` | a-plain | e is enumerated from ordinary dst; default inherited prototype members are non-enumerable. No explicit own check. | - |
| src/compiler/utilities.ts:8159:28 | unknown | `src[e]` | b | dst[e] is a-plain. src[e] is b: a missing constructor/toString/hasOwnProperty produces a function, __proto__ an object, and scalar inequality returns false. Must lower this read to has_own plus get_own before comparing. | [18_compare_missing_scalar.a](18_compare_missing_scalar.a) |
| src/compiler/utilities.ts:8605:43 | unknown | `localizedDiagnosticMessages[message.key]` | safe-domain | message.key comes from generated diagnostic metadata, not user-provided JSON keys; generated message keys include their numeric diagnostic code and are not prototype names. Public callers can forge DiagnosticMessage values. | - |
| src/compiler/utilities.ts:9402:9 | unknown | `options[option.name]` | safe-domain | Compiler callers enumerate optionDeclarations/affectingOptionDeclarations. Internal option names exclude prototype members; no explicit own guard, and missing ordinary option names are intentional. | - |

### Stock Node observations and input fixtures

All four keys (constructor, toString, hasOwnProperty, __proto__) are tested separately in each of seven CLI forms: missing paths, own paths, missing custom export condition with default fallback, own export condition, own typesVersions path, unknown compiler option in tsconfig, and unknown command-line option. Two additional configs exercise the empty-path-array comparison candidate. **30** unmodified stock tsc 6.0.3 runs: **29 correct, one wrong diagnostic, zero crashes, zero silently ignored cases**. Correct missing-path cases give TS2307; unknown flags give TS5023 with exit 1, unknown config options exit 2; empty arrays give TS5066.

The one CLI candidate is an own __proto__ paths mapping lost during config conversion, not an inherited miss at any of the 79 counted sites. Package JSON preserves that own key and resolves normally. Do not infer missing-read unsafety from this separate own-key construction bug.

Each committed input is under [input-fixtures](input-fixtures/), with its command in manifest.json and exact stdout/stderr/exit in observations.json. Source drivers are authored as .a; the runner copies unchanged text to temporary .ts/.d.ts filenames so stock tsc accepts them. The three standalone helper fixtures retain the complete upstream function and are in status.json with Node and stage-0 results. See [UPSTREAM_CANDIDATES.md](UPSTREAM_CANDIDATES.md) for the CLI candidate and the separately scoped stock helper/API candidate.

Reproduce with STOCK_TYPESCRIPT pointing to an installed typescript@6.0.3 package: python3 probe-inputs.py (CLI observations), node probe-guards.cjs (stock API and complete ownership-row coverage), and python3 verify.py from this repository environment (19 standalone fixtures and five stdout mutants). Save output to logs; no native inherited-miss or runtime conversion mutant is claimed while these fixtures fail stage-0 checking.
