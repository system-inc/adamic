# Location ledger

Coordinates refer to the pinned main adapted tree. Full bodies, types, diagnostics and rule explanations are in `ledger.json`; all 3,011 stores are in `evidence/inventory.json.gz`. Candidate property additions are incomplete runtime classifications.

## Double casts (18)

| Location | Rule | Observation |
| --- | --- | --- |
| src/compiler/checker.ts:10950:32 | generic-array | results as unknown as T[] |
| src/compiler/checker.ts:49445:40 | jsdoc | paramTag.parent.parent as unknown as JSDocCallbackTag |
| src/compiler/core.ts:736:36 | sorted-brand | emptyArray as any as SortedReadonlyArray<T> |
| src/compiler/core.ts:759:12 | sorted-brand | deduplicated as any as SortedReadonlyArray<T> |
| src/compiler/core.ts:764:12 | sorted-brand | [] as any as SortedArray<T> |
| src/compiler/core.ts:810:89 | comparator | compareStringsCaseSensitive as any as Comparer<T> |
| src/compiler/debug.ts:849:36 | debug-mapper | this.mapper1 as unknown as DebugTypeMapper |
| src/compiler/debug.ts:850:8 | debug-mapper | this.mapper2 as unknown as DebugTypeMapper |
| src/compiler/program.ts:3284:116 | sorted-brand | emptyArray as any as SortedReadonlyArray<Diagnostic> |
| src/compiler/transformer.ts:338:108 | sourcefile | node as any as SourceFile |
| src/compiler/tsbuildPublic.ts:992:23 | builder-generic | program as any as SemanticDiagnosticsBuilderProgram |
| src/compiler/tsbuildPublic.ts:993:22 | builder-generic | program as any as SemanticDiagnosticsBuilderProgram |
| src/compiler/tsbuildPublic.ts:1355:12 | builder-generic | readBuilderProgram(parsed.options, compilerHost) as any as T |
| src/compiler/watch.ts:862:41 | builder-generic | createEmitAndSemanticDiagnosticsBuilderProgram as any as CreateProgram<T> |
| src/compiler/watchPublic.ts:150:38 | builder-generic | createEmitAndSemanticDiagnosticsBuilderProgram as any as CreateProgram<T> |
| src/compiler/watchPublic.ts:151:24 | builder-generic | readBuilderProgram(options, host) as any as T |
| src/compiler/watchPublic.ts:555:22 | builder-generic | readBuilderProgram(compilerOptions, compilerHost) as any as T |
| src/compiler/watchUtilities.ts:839:13 | polling-enum | fallbackPolling as unknown as WatchFileKind |

## Function-valued expression (1)

| Location | Rule | Observation |
| --- | --- | --- |
| src/compiler/debug.ts:376:26 | Function | Function.prototype |

## Descriptor calls (9)

| Location | Rule | Observation |
| --- | --- | --- |
| src/compiler/commandLineParser.ts:3034:9 | descriptor-hidden | Object.defineProperty(options, "configFile", { enumerable: false, writable: false, value: configFile }) |
| src/compiler/debug.ts:518:13 | descriptor-debug | Object.defineProperties(flowNode, {                  // for use with vscode-js-debug's new customDescriptionGenerator in launch.json                   |
| src/compiler/debug.ts:575:13 | descriptor-debug | Object.defineProperties(array, {                  __tsDebuggerDisplay: {                      value(this: NodeArray<Node>, defaultValue: string) {     |
| src/compiler/debug.ts:621:9 | descriptor-debug | Object.defineProperties(objectAllocator.getSymbolConstructor().prototype, {              // for use with vscode-js-debug's new customDescriptionGenera |
| src/compiler/debug.ts:638:9 | descriptor-debug | Object.defineProperties(objectAllocator.getTypeConstructor().prototype, {              // for use with vscode-js-debug's new customDescriptionGenerato |
| src/compiler/debug.ts:692:9 | descriptor-debug | Object.defineProperties(objectAllocator.getSignatureConstructor().prototype, {              __debugFlags: {                  get(this: Signature) {    |
| src/compiler/debug.ts:714:17 | descriptor-debug | Object.defineProperties(ctor.prototype, {                      // for use with vscode-js-debug's new customDescriptionGenerator in launch.json         |
| src/compiler/factory/nodeFactory.ts:6097:9 | descriptor-redirect | Object.defineProperties(node, {              id: {                  get(this: SourceFile) {                      return this.redirectInfo!.redirectTar |
| src/compiler/scanner.ts:1117:9 | descriptor-debug | Object.defineProperty(scanner, "__debugShowCurrentPositionInText", {              get: () => {                  const text = scanner.getText();        |

## Assign into existing object (1)

| Location | Rule | Observation |
| --- | --- | --- |
| src/compiler/utilities.ts:8581:5 | assign | Object.assign(objectAllocator, alloc) |

## Possible property additions (49)

| Location | Rule | Observation |
| --- | --- | --- |
| src/compiler/checker.ts:6834:25 | property-candidate | (parentName as any).isTypeOf = true |
| src/compiler/checker.ts:13960:13 | property-candidate | links[resolutionKind] = earlySymbols \|\| emptySymbols |
| src/compiler/checker.ts:14014:13 | property-candidate | links[resolutionKind] = resolved \|\| emptySymbols |
| src/compiler/checker.ts:14177:17 | property-candidate | signature.optionalCallSignatureCache[key] = createOptionalCallSignature(signature, callChainFlags) |
| src/compiler/checker.ts:19531:9 | property-candidate | type[cache] = circularConstraintType |
| src/compiler/checker.ts:19540:20 | property-candidate | type[cache] = distributedOverIndex |
| src/compiler/checker.ts:19549:24 | property-candidate | type[cache] = distributedOverObject |
| src/compiler/checker.ts:19561:24 | property-candidate | type[cache] = elementType |
| src/compiler/checker.ts:19569:24 | property-candidate | type[cache] = mapType(substituteIndexedMappedType(objectType, type.indexType), t => getSimplifiedType(t, writing)) |
| src/compiler/checker.ts:19572:16 | property-candidate | type[cache] = type |
| src/compiler/checker.ts:45904:16 | property-candidate | (type as IterableOrIteratorType)[cacheKey] = cachedTypes |
| src/compiler/commandLineParser.ts:2515:21 | property-candidate | result[keyText] = value |
| src/compiler/commandLineParser.ts:2985:9 | property-candidate | result.configFilePath = toAbsolutePath(result.configFilePath) |
| src/compiler/commandLineParser.ts:3444:29 | property-candidate | ownConfig.raw.include = result.include |
| src/compiler/commandLineParser.ts:3445:29 | property-candidate | ownConfig.raw.exclude = result.exclude |
| src/compiler/commandLineParser.ts:3446:27 | property-candidate | ownConfig.raw.files = result.files |
| src/compiler/commandLineParser.ts:3448:80 | property-candidate | ownConfig.raw.compileOnSave = result.compileOnSave |
| src/compiler/commandLineParser.ts:3466:21 | property-candidate | result[propertyName] = map(extendsRaw[propertyName], (path: string) =>                          startsWithConfigDirTemplate(path) \|\| isRootedDiskPat |
| src/compiler/commandLineParser.ts:3510:5 | property-candidate | json.compileOnSave = convertCompileOnSaveOptionFromJson(json, basePath, errors) |
| src/compiler/core.ts:1355:17 | property-candidate | t[p] = arg[p] |
| src/compiler/core.ts:1482:13 | property-candidate | result[id] = (object as any)[id] |
| src/compiler/core.ts:1499:13 | property-candidate | (result as any)[id] = (second as any)[id] |
| src/compiler/core.ts:1505:13 | property-candidate | (result as any)[id] = (first as any)[id] |
| src/compiler/core.ts:1516:13 | property-candidate | (first as any)[id] = second[id] |
| src/compiler/debug.ts:170:21 | property-candidate | (Debug as any)[key] = cachedFunc |
| src/compiler/debug.ts:171:21 | property-candidate | assertionCache[key] = undefined |
| src/compiler/debug.ts:189:13 | property-candidate | assertionCache[name] = { level, assertion: Debug[name] } |
| src/compiler/debug.ts:190:13 | property-candidate | (Debug as any)[name] = noop |
| src/compiler/factory/nodeFactory.ts:6145:13 | property-candidate | (node as any)[p] = (source as any)[p] |
| src/compiler/factory/nodeFactory.ts:6404:13 | property-candidate | clone[key] = node[key] |
| src/compiler/program.ts:4069:13 | property-candidate | resolvedRef.references = commandLine.projectReferences.map(parseProjectReferenceConfigFile) |
| src/compiler/sys.ts:77:9 | property-candidate | Error.stackTraceLimit = 100 |
| src/compiler/sys.ts:156:17 | property-candidate | (customLevels \|\| (customLevels = {}))[level] = Number(customLevel) |
| src/compiler/sys.ts:172:13 | property-candidate | levels[level] = customLevels![level] \|\| levels[level] |
| src/compiler/sys.ts:274:17 | property-candidate | file.isClosed = true |
| src/compiler/sys.ts:477:17 | property-candidate | file.isClosed = true |
| src/compiler/sys.ts:1683:25 | property-candidate | node.callFrame.url = getRelativePathToDirectoryOrUrl(fileUrlRoot, url, fileUrlRoot, createGetCanonicalFileName(useCaseSensitiveFileNames), /*isAbsolut |
| src/compiler/sys.ts:1686:25 | property-candidate | node.callFrame.url = (remappedPaths.has(url) ? remappedPaths : remappedPaths.set(url, `external${externalFileCounter}.js`)).get(url)! |
| src/compiler/sys.ts:1803:21 | property-candidate | buffer[i] = buffer[i + 1]! |
| src/compiler/sys.ts:1804:21 | property-candidate | buffer[i + 1] = temp |
| src/compiler/transformers/es2015.ts:3584:17 | property-candidate | currentState.argumentsName = convertedLoopState.argumentsName |
| src/compiler/transformers/es2015.ts:3589:17 | property-candidate | currentState.thisName = convertedLoopState.thisName |
| src/compiler/transformers/es2015.ts:3594:17 | property-candidate | currentState.hoistedLocalVariables = convertedLoopState.hoistedLocalVariables |
| src/compiler/transformers/esDecorators.ts:1352:21 | property-candidate | memberInfo.memberDescriptorName = descriptorName = createHelperVariable(member, "descriptor") |
| src/compiler/transformers/esDecorators.ts:1362:36 | property-candidate | memberInfo.memberInitializersName ??= createHelperVariable(member, "initializers") |
| src/compiler/transformers/esDecorators.ts:1363:41 | property-candidate | memberInfo.memberExtraInitializersName ??= createHelperVariable(member, "extraInitializers") |
| src/compiler/transformers/esDecorators.ts:1372:21 | property-candidate | memberInfo.memberDescriptorName = descriptorName = createHelperVariable(member, "descriptor") |
| src/compiler/tsbuildPublic.ts:336:5 | property-candidate | result.tscBuild = true |
| src/compiler/utilities.ts:8660:9 | property-candidate | diagnosticWithLocation.relatedInformation = [] |

## Predicate annotations (651; 580 with bodies)

| Location | Rule | Observation |
| --- | --- | --- |
| src/compiler/builder.ts:268:87 | predicate | Refused; isBuilderProgramStateWithDefinedProgram; body |
| src/compiler/builder.ts:1176:79 | predicate | Refused; isIncrementalBundleEmitBuildInfo; body |
| src/compiler/builder.ts:1181:58 | predicate | Refused; isIncrementalBuildInfo; body |
| src/compiler/builder.ts:1192:54 | predicate | Refused; isNonIncrementalBuildInfo; body |
| src/compiler/checker.ts:4290:132 | predicate | Refused; isNonLocalAlias; body |
| src/compiler/checker.ts:7873:81 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/checker.ts:9408:75 | predicate | Refused; isIdentifierAndNotUndefined; body |
| src/compiler/checker.ts:13387:43 | predicate | Refused; isValidBaseType; body |
| src/compiler/checker.ts:13778:57 | predicate | Refused; isLateBindableName; body |
| src/compiler/checker.ts:13783:67 | predicate | Refused; isLateBindableIndexSignature; body |
| src/compiler/checker.ts:13809:54 | predicate | Refused; hasLateBindableName; body |
| src/compiler/checker.ts:14997:47 | predicate | Refused; isGenericMappedType; body |
| src/compiler/checker.ts:17304:48 | predicate | Refused; isJSDocTypeReference; body |
| src/compiler/checker.ts:21245:75 | predicate | Refused; isContextSensitiveFunctionOrObjectLiteralMethod; body |
| src/compiler/checker.ts:25116:54 | predicate | Refused; isNonDeferredTypeReference; body |
| src/compiler/checker.ts:25515:39 | predicate | Refused; isArrayType; body |
| src/compiler/checker.ts:25523:46 | predicate | Refused; isArrayOrTupleType; body |
| src/compiler/checker.ts:25707:39 | predicate | Refused; isTupleType; body |
| src/compiler/checker.ts:25711:46 | predicate | Refused; isGenericTupleType; body |
| src/compiler/checker.ts:25715:59 | predicate | Refused; isSingleElementGenericTupleType; body |
| src/compiler/checker.ts:26489:43 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/checker.ts:32636:54 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/checker.ts:33802:52 | predicate | Refused; isJsxIntrinsicTagName; body |
| src/compiler/checker.ts:35331:56 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/checker.ts:35569:80 | predicate | Refused; callLikeExpressionMayHaveTypeArguments; body |
| src/compiler/checker.ts:35656:61 | predicate | Refused; isSpreadArgument; body |
| src/compiler/checker.ts:37696:55 | predicate | Refused; isJSConstructor; body |
| src/compiler/checker.ts:38480:63 | predicate | Refused; isValidDeclarationForTupleLabel; body |
| src/compiler/checker.ts:43611:54 | predicate | Refused; isAwaitedTypeInstantiation; body |
| src/compiler/checker.ts:44756:49 | predicate | Refused; isImportedDeclaration; body |
| src/compiler/checker.ts:47505:62 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/checker.ts:50641:59 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/checker.ts:50886:57 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/checker.ts:51062:48 | predicate | Refused; canHaveConstantValue; body |
| src/compiler/checker.ts:51437:49 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/commandLineParser.ts:1728:70 | predicate | Refused; isCommandLineOptionOfCustomType; body |
| src/compiler/commandLineParser.ts:2606:85 | predicate | Refused; isCompilerOptionsValue; body |
| src/compiler/commandLineParser.ts:3038:37 | predicate | Refused; isNullOrUndefined; body |
| src/compiler/commandLineParser.ts:3300:51 | predicate | Refused; startsWithConfigDirTemplate; body |
| src/compiler/core.ts:140:101 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:140:116 | signature | Refused; every; no body on declaration |
| src/compiler/core.ts:142:113 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:142:128 | signature | Refused; every; no body on declaration |
| src/compiler/core.ts:162:113 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:178:117 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:261:65 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:265:74 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:269:77 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:273:86 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:614:59 | signature | Refused; some; no body on declaration |
| src/compiler/core.ts:1461:102 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:1750:42 | predicate | Refused; isArray; body |
| src/compiler/core.ts:1769:42 | proven | Proven; isString; body |
| src/compiler/core.ts:1773:39 | proven | Proven; isNumber; body |
| src/compiler/core.ts:1778:100 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:1783:97 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2459:66 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2459:91 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2459:112 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2461:80 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2461:105 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2461:130 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2461:151 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2557:91 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2559:103 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2560:103 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2572:91 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2574:103 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/core.ts:2576:103 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:213:142 | predicate | Refused; assert; body |
| src/compiler/debug.ts:248:99 | predicate | Refused; assertIsDefined; body |
| src/compiler/debug.ts:260:127 | signature | Refused; assertEachIsDefined; no body on declaration |
| src/compiler/debug.ts:261:114 | signature | Refused; assertEachIsDefined; no body on declaration |
| src/compiler/debug.ts:278:105 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:278:165 | signature | Refused; assertEachNode; no body on declaration |
| src/compiler/debug.ts:279:105 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:279:165 | signature | Refused; assertEachNode; no body on declaration |
| src/compiler/debug.ts:280:117 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:280:177 | signature | Refused; assertEachNode; no body on declaration |
| src/compiler/debug.ts:281:117 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:281:177 | signature | Refused; assertEachNode; no body on declaration |
| src/compiler/debug.ts:294:101 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:294:161 | signature | Refused; assertNode; no body on declaration |
| src/compiler/debug.ts:307:107 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:307:167 | signature | Refused; assertNotNode; no body on declaration |
| src/compiler/debug.ts:320:97 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:320:157 | signature | Refused; assertOptionalNode; no body on declaration |
| src/compiler/debug.ts:321:109 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/debug.ts:321:169 | signature | Refused; assertOptionalNode; no body on declaration |
| src/compiler/debug.ts:334:146 | signature | Refused; assertOptionalToken; no body on declaration |
| src/compiler/debug.ts:335:158 | signature | Refused; assertOptionalToken; no body on declaration |
| src/compiler/debug.ts:348:112 | signature | Refused; assertMissingNode; no body on declaration |
| src/compiler/debug.ts:365:46 | empty-assertion | Refused; type; no body on declaration |
| src/compiler/debug.ts:958:51 | predicate | Refused; isFlowSwitchClause; body |
| src/compiler/debug.ts:962:47 | predicate | Refused; hasAntecedents; body |
| src/compiler/debug.ts:966:46 | predicate | Refused; hasAntecedent; body |
| src/compiler/debug.ts:970:40 | predicate | Refused; hasNode; body |
| src/compiler/emitter.ts:1158:27 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/factory/nodeTests.ts:235:47 | predicate | Refused; isNumericLiteral; body |
| src/compiler/factory/nodeTests.ts:239:46 | predicate | Refused; isBigIntLiteral; body |
| src/compiler/factory/nodeTests.ts:243:46 | predicate | Refused; isStringLiteral; body |
| src/compiler/factory/nodeTests.ts:247:40 | predicate | Refused; isJsxText; body |
| src/compiler/factory/nodeTests.ts:251:57 | predicate | Refused; isRegularExpressionLiteral; body |
| src/compiler/factory/nodeTests.ts:255:62 | predicate | Refused; isNoSubstitutionTemplateLiteral; body |
| src/compiler/factory/nodeTests.ts:261:45 | predicate | Refused; isTemplateHead; body |
| src/compiler/factory/nodeTests.ts:265:47 | predicate | Refused; isTemplateMiddle; body |
| src/compiler/factory/nodeTests.ts:269:45 | predicate | Refused; isTemplateTail; body |
| src/compiler/factory/nodeTests.ts:275:47 | predicate | Refused; isDotDotDotToken; body |
| src/compiler/factory/nodeTests.ts:280:43 | predicate | Refused; isCommaToken; body |
| src/compiler/factory/nodeTests.ts:284:42 | predicate | Refused; isPlusToken; body |
| src/compiler/factory/nodeTests.ts:288:43 | predicate | Refused; isMinusToken; body |
| src/compiler/factory/nodeTests.ts:292:46 | predicate | Refused; isAsteriskToken; body |
| src/compiler/factory/nodeTests.ts:296:49 | predicate | Refused; isExclamationToken; body |
| src/compiler/factory/nodeTests.ts:300:46 | predicate | Refused; isQuestionToken; body |
| src/compiler/factory/nodeTests.ts:304:43 | predicate | Refused; isColonToken; body |
| src/compiler/factory/nodeTests.ts:308:49 | predicate | Refused; isQuestionDotToken; body |
| src/compiler/factory/nodeTests.ts:312:55 | predicate | Refused; isEqualsGreaterThanToken; body |
| src/compiler/factory/nodeTests.ts:318:43 | predicate | Refused; isIdentifier; body |
| src/compiler/factory/nodeTests.ts:322:50 | predicate | Refused; isPrivateIdentifier; body |
| src/compiler/factory/nodeTests.ts:329:47 | predicate | Refused; isExportModifier; body |
| src/compiler/factory/nodeTests.ts:334:48 | predicate | Refused; isDefaultModifier; body |
| src/compiler/factory/nodeTests.ts:339:46 | predicate | Refused; isAsyncModifier; body |
| src/compiler/factory/nodeTests.ts:343:47 | predicate | Refused; isAssertsKeyword; body |
| src/compiler/factory/nodeTests.ts:347:45 | predicate | Refused; isAwaitKeyword; body |
| src/compiler/factory/nodeTests.ts:352:48 | predicate | Refused; isReadonlyKeyword; body |
| src/compiler/factory/nodeTests.ts:357:47 | predicate | Refused; isStaticModifier; body |
| src/compiler/factory/nodeTests.ts:362:49 | predicate | Refused; isAbstractModifier; body |
| src/compiler/factory/nodeTests.ts:367:49 | predicate | Refused; isOverrideModifier; body |
| src/compiler/factory/nodeTests.ts:372:49 | predicate | Refused; isAccessorModifier; body |
| src/compiler/factory/nodeTests.ts:377:45 | predicate | Refused; isSuperKeyword; body |
| src/compiler/factory/nodeTests.ts:382:46 | predicate | Refused; isImportKeyword; body |
| src/compiler/factory/nodeTests.ts:387:44 | predicate | Refused; isCaseKeyword; body |
| src/compiler/factory/nodeTests.ts:393:46 | predicate | Refused; isQualifiedName; body |
| src/compiler/factory/nodeTests.ts:397:53 | predicate | Refused; isComputedPropertyName; body |
| src/compiler/factory/nodeTests.ts:403:57 | predicate | Refused; isTypeParameterDeclaration; body |
| src/compiler/factory/nodeTests.ts:408:42 | predicate | Refused; isParameter; body |
| src/compiler/factory/nodeTests.ts:412:42 | predicate | Refused; isDecorator; body |
| src/compiler/factory/nodeTests.ts:418:50 | predicate | Refused; isPropertySignature; body |
| src/compiler/factory/nodeTests.ts:422:52 | predicate | Refused; isPropertyDeclaration; body |
| src/compiler/factory/nodeTests.ts:426:48 | predicate | Refused; isMethodSignature; body |
| src/compiler/factory/nodeTests.ts:430:50 | predicate | Refused; isMethodDeclaration; body |
| src/compiler/factory/nodeTests.ts:434:60 | predicate | Refused; isClassStaticBlockDeclaration; body |
| src/compiler/factory/nodeTests.ts:438:55 | predicate | Refused; isConstructorDeclaration; body |
| src/compiler/factory/nodeTests.ts:442:55 | predicate | Refused; isGetAccessorDeclaration; body |
| src/compiler/factory/nodeTests.ts:446:55 | predicate | Refused; isSetAccessorDeclaration; body |
| src/compiler/factory/nodeTests.ts:450:57 | predicate | Refused; isCallSignatureDeclaration; body |
| src/compiler/factory/nodeTests.ts:454:62 | predicate | Refused; isConstructSignatureDeclaration; body |
| src/compiler/factory/nodeTests.ts:458:58 | predicate | Refused; isIndexSignatureDeclaration; body |
| src/compiler/factory/nodeTests.ts:464:50 | predicate | Refused; isTypePredicateNode; body |
| src/compiler/factory/nodeTests.ts:468:50 | predicate | Refused; isTypeReferenceNode; body |
| src/compiler/factory/nodeTests.ts:472:49 | predicate | Refused; isFunctionTypeNode; body |
| src/compiler/factory/nodeTests.ts:476:52 | predicate | Refused; isConstructorTypeNode; body |
| src/compiler/factory/nodeTests.ts:480:46 | predicate | Refused; isTypeQueryNode; body |
| src/compiler/factory/nodeTests.ts:484:48 | predicate | Refused; isTypeLiteralNode; body |
| src/compiler/factory/nodeTests.ts:488:46 | predicate | Refused; isArrayTypeNode; body |
| src/compiler/factory/nodeTests.ts:492:46 | predicate | Refused; isTupleTypeNode; body |
| src/compiler/factory/nodeTests.ts:496:49 | predicate | Refused; isNamedTupleMember; body |
| src/compiler/factory/nodeTests.ts:500:49 | predicate | Refused; isOptionalTypeNode; body |
| src/compiler/factory/nodeTests.ts:504:45 | predicate | Refused; isRestTypeNode; body |
| src/compiler/factory/nodeTests.ts:508:46 | predicate | Refused; isUnionTypeNode; body |
| src/compiler/factory/nodeTests.ts:512:53 | predicate | Refused; isIntersectionTypeNode; body |
| src/compiler/factory/nodeTests.ts:516:52 | predicate | Refused; isConditionalTypeNode; body |
| src/compiler/factory/nodeTests.ts:520:46 | predicate | Refused; isInferTypeNode; body |
| src/compiler/factory/nodeTests.ts:524:54 | predicate | Refused; isParenthesizedTypeNode; body |
| src/compiler/factory/nodeTests.ts:528:45 | predicate | Refused; isThisTypeNode; body |
| src/compiler/factory/nodeTests.ts:532:49 | predicate | Refused; isTypeOperatorNode; body |
| src/compiler/factory/nodeTests.ts:536:54 | predicate | Refused; isIndexedAccessTypeNode; body |
| src/compiler/factory/nodeTests.ts:540:47 | predicate | Refused; isMappedTypeNode; body |
| src/compiler/factory/nodeTests.ts:544:48 | predicate | Refused; isLiteralTypeNode; body |
| src/compiler/factory/nodeTests.ts:548:47 | predicate | Refused; isImportTypeNode; body |
| src/compiler/factory/nodeTests.ts:552:56 | predicate | Refused; isTemplateLiteralTypeSpan; body |
| src/compiler/factory/nodeTests.ts:556:56 | predicate | Refused; isTemplateLiteralTypeNode; body |
| src/compiler/factory/nodeTests.ts:562:53 | predicate | Refused; isObjectBindingPattern; body |
| src/compiler/factory/nodeTests.ts:566:52 | predicate | Refused; isArrayBindingPattern; body |
| src/compiler/factory/nodeTests.ts:570:47 | predicate | Refused; isBindingElement; body |
| src/compiler/factory/nodeTests.ts:576:55 | predicate | Refused; isArrayLiteralExpression; body |
| src/compiler/factory/nodeTests.ts:580:56 | predicate | Refused; isObjectLiteralExpression; body |
| src/compiler/factory/nodeTests.ts:584:57 | predicate | Refused; isPropertyAccessExpression; body |
| src/compiler/factory/nodeTests.ts:588:56 | predicate | Refused; isElementAccessExpression; body |
| src/compiler/factory/nodeTests.ts:592:47 | predicate | Refused; isCallExpression; body |
| src/compiler/factory/nodeTests.ts:596:46 | predicate | Refused; isNewExpression; body |
| src/compiler/factory/nodeTests.ts:600:57 | predicate | Refused; isTaggedTemplateExpression; body |
| src/compiler/factory/nodeTests.ts:604:56 | predicate | Refused; isTypeAssertionExpression; body |
| src/compiler/factory/nodeTests.ts:608:56 | predicate | Refused; isParenthesizedExpression; body |
| src/compiler/factory/nodeTests.ts:612:51 | predicate | Refused; isFunctionExpression; body |
| src/compiler/factory/nodeTests.ts:616:46 | predicate | Refused; isArrowFunction; body |
| src/compiler/factory/nodeTests.ts:620:49 | predicate | Refused; isDeleteExpression; body |
| src/compiler/factory/nodeTests.ts:624:49 | predicate | Refused; isTypeOfExpression; body |
| src/compiler/factory/nodeTests.ts:628:47 | predicate | Refused; isVoidExpression; body |
| src/compiler/factory/nodeTests.ts:632:48 | predicate | Refused; isAwaitExpression; body |
| src/compiler/factory/nodeTests.ts:636:54 | predicate | Refused; isPrefixUnaryExpression; body |
| src/compiler/factory/nodeTests.ts:640:55 | predicate | Refused; isPostfixUnaryExpression; body |
| src/compiler/factory/nodeTests.ts:644:49 | predicate | Refused; isBinaryExpression; body |
| src/compiler/factory/nodeTests.ts:648:54 | predicate | Refused; isConditionalExpression; body |
| src/compiler/factory/nodeTests.ts:652:51 | predicate | Refused; isTemplateExpression; body |
| src/compiler/factory/nodeTests.ts:656:48 | predicate | Refused; isYieldExpression; body |
| src/compiler/factory/nodeTests.ts:660:46 | predicate | Refused; isSpreadElement; body |
| src/compiler/factory/nodeTests.ts:664:48 | predicate | Refused; isClassExpression; body |
| src/compiler/factory/nodeTests.ts:668:50 | predicate | Refused; isOmittedExpression; body |
| src/compiler/factory/nodeTests.ts:672:60 | predicate | Refused; isExpressionWithTypeArguments; body |
| src/compiler/factory/nodeTests.ts:676:45 | predicate | Refused; isAsExpression; body |
| src/compiler/factory/nodeTests.ts:680:52 | predicate | Refused; isSatisfiesExpression; body |
| src/compiler/factory/nodeTests.ts:684:50 | predicate | Refused; isNonNullExpression; body |
| src/compiler/factory/nodeTests.ts:688:45 | predicate | Refused; isMetaProperty; body |
| src/compiler/factory/nodeTests.ts:692:52 | predicate | Refused; isSyntheticExpression; body |
| src/compiler/factory/nodeTests.ts:696:59 | predicate | Refused; isPartiallyEmittedExpression; body |
| src/compiler/factory/nodeTests.ts:700:52 | predicate | Refused; isCommaListExpression; body |
| src/compiler/factory/nodeTests.ts:706:45 | predicate | Refused; isTemplateSpan; body |
| src/compiler/factory/nodeTests.ts:710:54 | predicate | Refused; isSemicolonClassElement; body |
| src/compiler/factory/nodeTests.ts:716:38 | predicate | Refused; isBlock; body |
| src/compiler/factory/nodeTests.ts:720:50 | predicate | Refused; isVariableStatement; body |
| src/compiler/factory/nodeTests.ts:724:47 | predicate | Refused; isEmptyStatement; body |
| src/compiler/factory/nodeTests.ts:728:52 | predicate | Refused; isExpressionStatement; body |
| src/compiler/factory/nodeTests.ts:732:44 | predicate | Refused; isIfStatement; body |
| src/compiler/factory/nodeTests.ts:736:44 | predicate | Refused; isDoStatement; body |
| src/compiler/factory/nodeTests.ts:740:47 | predicate | Refused; isWhileStatement; body |
| src/compiler/factory/nodeTests.ts:744:45 | predicate | Refused; isForStatement; body |
| src/compiler/factory/nodeTests.ts:748:47 | predicate | Refused; isForInStatement; body |
| src/compiler/factory/nodeTests.ts:752:47 | predicate | Refused; isForOfStatement; body |
| src/compiler/factory/nodeTests.ts:756:50 | predicate | Refused; isContinueStatement; body |
| src/compiler/factory/nodeTests.ts:760:47 | predicate | Refused; isBreakStatement; body |
| src/compiler/factory/nodeTests.ts:764:48 | predicate | Refused; isReturnStatement; body |
| src/compiler/factory/nodeTests.ts:768:46 | predicate | Refused; isWithStatement; body |
| src/compiler/factory/nodeTests.ts:772:48 | predicate | Refused; isSwitchStatement; body |
| src/compiler/factory/nodeTests.ts:776:49 | predicate | Refused; isLabeledStatement; body |
| src/compiler/factory/nodeTests.ts:780:47 | predicate | Refused; isThrowStatement; body |
| src/compiler/factory/nodeTests.ts:784:45 | predicate | Refused; isTryStatement; body |
| src/compiler/factory/nodeTests.ts:788:50 | predicate | Refused; isDebuggerStatement; body |
| src/compiler/factory/nodeTests.ts:792:52 | predicate | Refused; isVariableDeclaration; body |
| src/compiler/factory/nodeTests.ts:796:56 | predicate | Refused; isVariableDeclarationList; body |
| src/compiler/factory/nodeTests.ts:800:52 | predicate | Refused; isFunctionDeclaration; body |
| src/compiler/factory/nodeTests.ts:804:49 | predicate | Refused; isClassDeclaration; body |
| src/compiler/factory/nodeTests.ts:808:53 | predicate | Refused; isInterfaceDeclaration; body |
| src/compiler/factory/nodeTests.ts:812:53 | predicate | Refused; isTypeAliasDeclaration; body |
| src/compiler/factory/nodeTests.ts:816:48 | predicate | Refused; isEnumDeclaration; body |
| src/compiler/factory/nodeTests.ts:820:50 | predicate | Refused; isModuleDeclaration; body |
| src/compiler/factory/nodeTests.ts:824:44 | predicate | Refused; isModuleBlock; body |
| src/compiler/factory/nodeTests.ts:828:42 | predicate | Refused; isCaseBlock; body |
| src/compiler/factory/nodeTests.ts:832:59 | predicate | Refused; isNamespaceExportDeclaration; body |
| src/compiler/factory/nodeTests.ts:836:56 | predicate | Refused; isImportEqualsDeclaration; body |
| src/compiler/factory/nodeTests.ts:840:50 | predicate | Refused; isImportDeclaration; body |
| src/compiler/factory/nodeTests.ts:844:45 | predicate | Refused; isImportClause; body |
| src/compiler/factory/nodeTests.ts:848:61 | predicate | Refused; isImportTypeAssertionContainer; body |
| src/compiler/factory/nodeTests.ts:853:45 | predicate | Refused; isAssertClause; body |
| src/compiler/factory/nodeTests.ts:858:44 | predicate | Refused; isAssertEntry; body |
| src/compiler/factory/nodeTests.ts:862:49 | predicate | Refused; isImportAttributes; body |
| src/compiler/factory/nodeTests.ts:866:48 | predicate | Refused; isImportAttribute; body |
| src/compiler/factory/nodeTests.ts:870:48 | predicate | Refused; isNamespaceImport; body |
| src/compiler/factory/nodeTests.ts:874:48 | predicate | Refused; isNamespaceExport; body |
| src/compiler/factory/nodeTests.ts:878:45 | predicate | Refused; isNamedImports; body |
| src/compiler/factory/nodeTests.ts:882:48 | predicate | Refused; isImportSpecifier; body |
| src/compiler/factory/nodeTests.ts:886:49 | predicate | Refused; isExportAssignment; body |
| src/compiler/factory/nodeTests.ts:890:50 | predicate | Refused; isExportDeclaration; body |
| src/compiler/factory/nodeTests.ts:894:45 | predicate | Refused; isNamedExports; body |
| src/compiler/factory/nodeTests.ts:898:48 | predicate | Refused; isExportSpecifier; body |
| src/compiler/factory/nodeTests.ts:902:49 | predicate | Refused; isModuleExportName; body |
| src/compiler/factory/nodeTests.ts:906:51 | predicate | Refused; isMissingDeclaration; body |
| src/compiler/factory/nodeTests.ts:910:52 | predicate | Refused; isNotEmittedStatement; body |
| src/compiler/factory/nodeTests.ts:915:51 | predicate | Refused; isSyntheticReference; body |
| src/compiler/factory/nodeTests.ts:921:56 | predicate | Refused; isExternalModuleReference; body |
| src/compiler/factory/nodeTests.ts:927:43 | predicate | Refused; isJsxElement; body |
| src/compiler/factory/nodeTests.ts:931:54 | predicate | Refused; isJsxSelfClosingElement; body |
| src/compiler/factory/nodeTests.ts:935:50 | predicate | Refused; isJsxOpeningElement; body |
| src/compiler/factory/nodeTests.ts:939:50 | predicate | Refused; isJsxClosingElement; body |
| src/compiler/factory/nodeTests.ts:943:44 | predicate | Refused; isJsxFragment; body |
| src/compiler/factory/nodeTests.ts:947:51 | predicate | Refused; isJsxOpeningFragment; body |
| src/compiler/factory/nodeTests.ts:951:51 | predicate | Refused; isJsxClosingFragment; body |
| src/compiler/factory/nodeTests.ts:955:45 | predicate | Refused; isJsxAttribute; body |
| src/compiler/factory/nodeTests.ts:959:46 | predicate | Refused; isJsxAttributes; body |
| src/compiler/factory/nodeTests.ts:963:51 | predicate | Refused; isJsxSpreadAttribute; body |
| src/compiler/factory/nodeTests.ts:967:46 | predicate | Refused; isJsxExpression; body |
| src/compiler/factory/nodeTests.ts:971:50 | predicate | Refused; isJsxNamespacedName; body |
| src/compiler/factory/nodeTests.ts:977:43 | predicate | Refused; isCaseClause; body |
| src/compiler/factory/nodeTests.ts:981:46 | predicate | Refused; isDefaultClause; body |
| src/compiler/factory/nodeTests.ts:985:47 | predicate | Refused; isHeritageClause; body |
| src/compiler/factory/nodeTests.ts:989:44 | predicate | Refused; isCatchClause; body |
| src/compiler/factory/nodeTests.ts:995:51 | predicate | Refused; isPropertyAssignment; body |
| src/compiler/factory/nodeTests.ts:999:60 | predicate | Refused; isShorthandPropertyAssignment; body |
| src/compiler/factory/nodeTests.ts:1003:49 | predicate | Refused; isSpreadAssignment; body |
| src/compiler/factory/nodeTests.ts:1009:43 | predicate | Refused; isEnumMember; body |
| src/compiler/factory/nodeTests.ts:1014:43 | predicate | Refused; isSourceFile; body |
| src/compiler/factory/nodeTests.ts:1018:39 | predicate | Refused; isBundle; body |
| src/compiler/factory/nodeTests.ts:1026:52 | predicate | Refused; isJSDocTypeExpression; body |
| src/compiler/factory/nodeTests.ts:1030:51 | predicate | Refused; isJSDocNameReference; body |
| src/compiler/factory/nodeTests.ts:1034:48 | predicate | Refused; isJSDocMemberName; body |
| src/compiler/factory/nodeTests.ts:1038:42 | predicate | Refused; isJSDocLink; body |
| src/compiler/factory/nodeTests.ts:1042:46 | predicate | Refused; isJSDocLinkCode; body |
| src/compiler/factory/nodeTests.ts:1046:47 | predicate | Refused; isJSDocLinkPlain; body |
| src/compiler/factory/nodeTests.ts:1050:45 | predicate | Refused; isJSDocAllType; body |
| src/compiler/factory/nodeTests.ts:1054:49 | predicate | Refused; isJSDocUnknownType; body |
| src/compiler/factory/nodeTests.ts:1058:50 | predicate | Refused; isJSDocNullableType; body |
| src/compiler/factory/nodeTests.ts:1062:53 | predicate | Refused; isJSDocNonNullableType; body |
| src/compiler/factory/nodeTests.ts:1066:50 | predicate | Refused; isJSDocOptionalType; body |
| src/compiler/factory/nodeTests.ts:1070:50 | predicate | Refused; isJSDocFunctionType; body |
| src/compiler/factory/nodeTests.ts:1074:50 | predicate | Refused; isJSDocVariadicType; body |
| src/compiler/factory/nodeTests.ts:1078:50 | predicate | Refused; isJSDocNamepathType; body |
| src/compiler/factory/nodeTests.ts:1082:38 | predicate | Refused; isJSDoc; body |
| src/compiler/factory/nodeTests.ts:1086:49 | predicate | Refused; isJSDocTypeLiteral; body |
| src/compiler/factory/nodeTests.ts:1090:47 | predicate | Refused; isJSDocSignature; body |
| src/compiler/factory/nodeTests.ts:1096:49 | predicate | Refused; isJSDocAugmentsTag; body |
| src/compiler/factory/nodeTests.ts:1100:47 | predicate | Refused; isJSDocAuthorTag; body |
| src/compiler/factory/nodeTests.ts:1104:46 | predicate | Refused; isJSDocClassTag; body |
| src/compiler/factory/nodeTests.ts:1108:49 | predicate | Refused; isJSDocCallbackTag; body |
| src/compiler/factory/nodeTests.ts:1112:47 | predicate | Refused; isJSDocPublicTag; body |
| src/compiler/factory/nodeTests.ts:1116:48 | predicate | Refused; isJSDocPrivateTag; body |
| src/compiler/factory/nodeTests.ts:1120:50 | predicate | Refused; isJSDocProtectedTag; body |
| src/compiler/factory/nodeTests.ts:1124:49 | predicate | Refused; isJSDocReadonlyTag; body |
| src/compiler/factory/nodeTests.ts:1128:49 | predicate | Refused; isJSDocOverrideTag; body |
| src/compiler/factory/nodeTests.ts:1132:49 | predicate | Refused; isJSDocOverloadTag; body |
| src/compiler/factory/nodeTests.ts:1136:51 | predicate | Refused; isJSDocDeprecatedTag; body |
| src/compiler/factory/nodeTests.ts:1140:44 | predicate | Refused; isJSDocSeeTag; body |
| src/compiler/factory/nodeTests.ts:1144:45 | predicate | Refused; isJSDocEnumTag; body |
| src/compiler/factory/nodeTests.ts:1148:50 | predicate | Refused; isJSDocParameterTag; body |
| src/compiler/factory/nodeTests.ts:1152:47 | predicate | Refused; isJSDocReturnTag; body |
| src/compiler/factory/nodeTests.ts:1156:45 | predicate | Refused; isJSDocThisTag; body |
| src/compiler/factory/nodeTests.ts:1160:45 | predicate | Refused; isJSDocTypeTag; body |
| src/compiler/factory/nodeTests.ts:1164:49 | predicate | Refused; isJSDocTemplateTag; body |
| src/compiler/factory/nodeTests.ts:1168:48 | predicate | Refused; isJSDocTypedefTag; body |
| src/compiler/factory/nodeTests.ts:1172:48 | predicate | Refused; isJSDocUnknownTag; body |
| src/compiler/factory/nodeTests.ts:1176:49 | predicate | Refused; isJSDocPropertyTag; body |
| src/compiler/factory/nodeTests.ts:1180:51 | predicate | Refused; isJSDocImplementsTag; body |
| src/compiler/factory/nodeTests.ts:1184:50 | predicate | Refused; isJSDocSatisfiesTag; body |
| src/compiler/factory/nodeTests.ts:1188:47 | predicate | Refused; isJSDocThrowsTag; body |
| src/compiler/factory/nodeTests.ts:1192:47 | predicate | Refused; isJSDocImportTag; body |
| src/compiler/factory/nodeTests.ts:1199:40 | predicate | Refused; isSyntaxList; body |
| src/compiler/factory/utilities.ts:607:54 | predicate | Refused; isCommaExpression; body |
| src/compiler/factory/utilities.ts:612:52 | predicate | Refused; isCommaSequence; body |
| src/compiler/factory/utilities.ts:617:51 | predicate | Refused; isJSDocTypeAssertion; body |
| src/compiler/factory/utilities.ts:631:104 | predicate | Refused; isOuterExpression; body |
| src/compiler/factory/utilities.ts:1065:48 | predicate | Refused; isStringOrNumericLiteral; body |
| src/compiler/factory/utilities.ts:1105:49 | predicate | Refused; canHaveIllegalType; body |
| src/compiler/factory/utilities.ts:1112:59 | predicate | Refused; canHaveIllegalTypeParameters; body |
| src/compiler/factory/utilities.ts:1120:55 | predicate | Refused; canHaveIllegalDecorators; body |
| src/compiler/factory/utilities.ts:1142:54 | predicate | Refused; canHaveIllegalModifiers; body |
| src/compiler/factory/utilities.ts:1151:59 | predicate | Refused; isQuestionOrExclamationToken; body |
| src/compiler/factory/utilities.ts:1155:57 | predicate | Refused; isIdentifierOrThisTypeNode; body |
| src/compiler/factory/utilities.ts:1159:66 | predicate | Refused; isReadonlyKeywordOrPlusOrMinusToken; body |
| src/compiler/factory/utilities.ts:1163:59 | predicate | Refused; isQuestionOrPlusOrMinusToken; body |
| src/compiler/factory/utilities.ts:1167:43 | predicate | Refused; isModuleName; body |
| src/compiler/factory/utilities.ts:1171:54 | predicate | Refused; isExponentiationOperator; body |
| src/compiler/factory/utilities.ts:1175:54 | predicate | Refused; isMultiplicativeOperator; body |
| src/compiler/factory/utilities.ts:1181:62 | predicate | Refused; isMultiplicativeOperatorOrHigher; body |
| src/compiler/factory/utilities.ts:1186:48 | predicate | Refused; isAdditiveOperator; body |
| src/compiler/factory/utilities.ts:1191:56 | predicate | Refused; isAdditiveOperatorOrHigher; body |
| src/compiler/factory/utilities.ts:1196:45 | predicate | Refused; isShiftOperator; body |
| src/compiler/factory/utilities.ts:1203:60 | predicate | Refused; isShiftOperatorOrHigher; body |
| src/compiler/factory/utilities.ts:1208:50 | predicate | Refused; isRelationalOperator; body |
| src/compiler/factory/utilities.ts:1217:58 | predicate | Refused; isRelationalOperatorOrHigher; body |
| src/compiler/factory/utilities.ts:1222:48 | predicate | Refused; isEqualityOperator; body |
| src/compiler/factory/utilities.ts:1229:56 | predicate | Refused; isEqualityOperatorOrHigher; body |
| src/compiler/factory/utilities.ts:1234:47 | predicate | Refused; isBitwiseOperator; body |
| src/compiler/factory/utilities.ts:1240:55 | predicate | Refused; isBitwiseOperatorOrHigher; body |
| src/compiler/factory/utilities.ts:1246:47 | predicate | Refused; isLogicalOperator; body |
| src/compiler/factory/utilities.ts:1251:55 | predicate | Refused; isLogicalOperatorOrHigher; body |
| src/compiler/factory/utilities.ts:1256:58 | predicate | Refused; isAssignmentOperatorOrHigher; body |
| src/compiler/factory/utilities.ts:1262:46 | predicate | Refused; isBinaryOperator; body |
| src/compiler/factory/utilities.ts:1267:52 | predicate | Refused; isBinaryOperatorToken; body |
| src/compiler/factory/utilities.ts:1489:58 | predicate | Refused; isExportOrDefaultKeywordKind; body |
| src/compiler/factory/utilities.ts:1494:56 | predicate | Refused; isExportOrDefaultModifier; body |
| src/compiler/factory/utilities.ts:1695:64 | predicate | Refused; isSyntheticParenthesizedExpression; body |
| src/compiler/factory/utilitiesPublic.ts:14:47 | predicate | Refused; canHaveModifiers; body |
| src/compiler/factory/utilitiesPublic.ts:43:48 | predicate | Refused; canHaveDecorators; body |
| src/compiler/moduleNameResolver.ts:925:82 | predicate | Refused; isPackageJsonInfo; body |
| src/compiler/moduleNameResolver.ts:930:89 | predicate | Refused; isMissingPackageJsonInfo; body |
| src/compiler/program.ts:1162:74 | predicate | Refused; isReferencedFile; body |
| src/compiler/program.ts:1190:108 | predicate | Refused; isReferenceFileLocation; body |
| src/compiler/sourcemap.ts:408:34 | predicate | Refused; isRawSourceMap; body |
| src/compiler/sourcemap.ts:630:52 | predicate | Refused; isSourceMapping; body |
| src/compiler/sourcemap.ts:668:57 | predicate | Refused; isSourceMappedPosition; body |
| src/compiler/transformers/classFields.ts:853:53 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/transformers/classFields.ts:3349:67 | predicate | Refused; isPrivateIdentifierInExpression; body |
| src/compiler/transformers/classFields.ts:3354:51 | predicate | Refused; isStaticPropertyDeclaration; body |
| src/compiler/transformers/classFields.ts:3358:69 | predicate | Refused; isStaticPropertyDeclarationOrClassStaticBlock; body |
| src/compiler/transformers/classThis.ts:72:57 | predicate | Refused; isClassThisAssignmentBlock; body |
| src/compiler/transformers/declarations.ts:659:54 | predicate | Refused; shouldPrintWithInitializer; body |
| src/compiler/transformers/declarations.ts:1891:49 | predicate | Refused; canHaveLiteralInitializer; body |
| src/compiler/transformers/declarations.ts:1916:55 | predicate | Refused; isPreservedDeclarationStatement; body |
| src/compiler/transformers/declarations.ts:1954:44 | predicate | Refused; isProcessedComponent; body |
| src/compiler/transformers/declarations/diagnostics.ts:138:52 | predicate | Refused; canProduceDiagnostics; body |
| src/compiler/transformers/declarations/diagnostics.ts:714:46 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/transformers/es2015.ts:1348:42 | predicate | Refused; isCapturedThis; body |
| src/compiler/transformers/es2015.ts:1353:44 | predicate | Refused; isSyntheticSuper; body |
| src/compiler/transformers/es2015.ts:1359:60 | predicate | Refused; isThisCapturingVariableStatement; body |
| src/compiler/transformers/es2015.ts:1365:62 | predicate | Refused; isThisCapturingVariableDeclaration; body |
| src/compiler/transformers/es2015.ts:1370:53 | predicate | Refused; isThisCapturingAssignment; body |
| src/compiler/transformers/es2015.ts:1375:50 | predicate | Refused; isTransformedSuperCall; body |
| src/compiler/transformers/es2015.ts:1386:62 | predicate | Refused; isTransformedSuperCallWithFallback; body |
| src/compiler/transformers/es2015.ts:1393:47 | predicate | Refused; isImplicitSuperCall; body |
| src/compiler/transformers/es2015.ts:1403:59 | predicate | Refused; isImplicitSuperCallWithFallback; body |
| src/compiler/transformers/es2015.ts:1410:75 | predicate | Refused; isThisCapturingTransformedSuperCallWithFallback; body |
| src/compiler/transformers/es2015.ts:1415:72 | predicate | Refused; isThisCapturingImplicitSuperCallWithFallback; body |
| src/compiler/transformers/es2015.ts:1419:54 | predicate | Refused; isTransformedSuperCallLike; body |
| src/compiler/transformers/es2015.ts:2029:122 | predicate | Refused; shouldAddRestParameter; body |
| src/compiler/transformers/es2015.ts:3379:80 | predicate | Refused; shouldConvertInitializerOfForStatement; body |
| src/compiler/transformers/es2015.ts:3383:78 | predicate | Refused; shouldConvertConditionOfForStatement; body |
| src/compiler/transformers/es2015.ts:3387:80 | predicate | Refused; shouldConvertIncrementorOfForStatement; body |
| src/compiler/transformers/es2017.ts:586:80 | predicate | Refused; isVariableDeclarationListWithCollidingName; body |
| src/compiler/transformers/es2017.ts:1018:44 | predicate | Refused; isSuperContainer; body |
| src/compiler/transformers/esnext.ts:774:54 | predicate | Refused; isUsingVariableDeclarationList; body |
| src/compiler/transformers/generators.ts:2420:56 | predicate | Refused; supportsUnlabeledBreak; body |
| src/compiler/transformers/generators.ts:2430:64 | predicate | Refused; supportsLabeledBreakOrContinue; body |
| src/compiler/transformers/generators.ts:2439:59 | predicate | Refused; supportsUnlabeledContinue; body |
| src/compiler/transformers/module/system.ts:1360:63 | predicate | Refused; shouldHoistForInitializer; body |
| src/compiler/transformers/namedEvaluation.ts:139:64 | predicate | Refused; isClassNamedEvaluationHelperBlock; body |
| src/compiler/transformers/ts.ts:1054:49 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/transformers/ts.ts:1171:56 | predicate | Refused; shouldAddTypeMetadata; body |
| src/compiler/transformers/ts.ts:1186:62 | predicate | Refused; shouldAddReturnTypeMetadata; body |
| src/compiler/transformers/ts.ts:1197:62 | predicate | Refused; shouldAddParamTypesMetadata; body |
| src/compiler/transformers/ts.ts:1297:93 | predicate | Refused; shouldEmitFunctionLikeDeclaration; body |
| src/compiler/transformers/utilities.ts:490:61 | predicate | Refused; isCompoundAssignment; body |
| src/compiler/transformers/utilities.ts:593:91 | predicate | Refused; isStaticPropertyDeclarationOrClassStaticBlockDeclaration; body |
| src/compiler/transformers/utilities.ts:630:62 | predicate | Refused; isInitializedProperty; body |
| src/compiler/transformers/utilities.ts:642:83 | predicate | Refused; isNonStaticMethodOrAccessorWithPrivateName; body |
| src/compiler/tsbuildPublic.ts:259:66 | predicate | Refused; isCircularBuildOrder; body |
| src/compiler/tsbuildPublic.ts:569:60 | predicate | Refused; isParsedCommandLine; body |
| src/compiler/tsbuildPublic.ts:1371:84 | predicate | Refused; isFileWatcherWithModifiedTime; body |
| src/compiler/types.ts:5913:37 | signature | Refused; isLateBound; no body on declaration |
| src/compiler/types.ts:9729:31 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/types.ts:9755:31 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/types.ts:10662:45 | signature | Refused; hasLateBindableName; no body on declaration |
| src/compiler/utilities.ts:655:52 | predicate | Refused; isTransientSymbol; body |
| src/compiler/utilities.ts:1111:143 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/utilities.ts:2078:46 | predicate | Refused; isAmbientModule; body |
| src/compiler/utilities.ts:2083:60 | predicate | Refused; isModuleWithStringLiteralName; body |
| src/compiler/utilities.ts:2088:55 | predicate | Refused; isNonGlobalAmbientModule; body |
| src/compiler/utilities.ts:2129:59 | predicate | Refused; isExternalModuleAugmentation; body |
| src/compiler/utilities.ts:2232:62 | predicate | Refused; isDeclarationWithTypeParameters; body |
| src/compiler/utilities.ts:2246:69 | predicate | Refused; isDeclarationWithTypeParameterChildren; body |
| src/compiler/utilities.ts:2276:48 | predicate | Refused; isAnyImportSyntax; body |
| src/compiler/utilities.ts:2287:65 | predicate | Refused; isAnyImportOrBareOrAccessedRequire; body |
| src/compiler/utilities.ts:2292:60 | predicate | Refused; isAnyImportOrRequireStatement; body |
| src/compiler/utilities.ts:2297:63 | predicate | Refused; isLateVisibilityPaintedStatement; body |
| src/compiler/utilities.ts:2315:65 | predicate | Refused; hasPossibleExternalModuleReference; body |
| src/compiler/utilities.ts:2320:52 | predicate | Refused; isAnyImportOrReExport; body |
| src/compiler/utilities.ts:2643:53 | predicate | Refused; isJsonSourceFile; body |
| src/compiler/utilities.ts:2701:39 | predicate | Refused; isSuperCall; body |
| src/compiler/utilities.ts:2706:40 | predicate | Refused; isImportCall; body |
| src/compiler/utilities.ts:2717:40 | predicate | Refused; isImportMeta; body |
| src/compiler/utilities.ts:2724:51 | predicate | Refused; isLiteralImportTypeNode; body |
| src/compiler/utilities.ts:2729:50 | predicate | Refused; isPrologueDirective; body |
| src/compiler/utilities.ts:2996:45 | predicate | Refused; isVariableLike; body |
| src/compiler/utilities.ts:3073:52 | predicate | Refused; isObjectLiteralMethod; body |
| src/compiler/utilities.ts:3078:79 | predicate | Refused; isObjectLiteralOrClassExpressionMethodOrAccessor; body |
| src/compiler/utilities.ts:3085:70 | predicate | Refused; isIdentifierTypePredicate; body |
| src/compiler/utilities.ts:3090:64 | predicate | Refused; isThisTypePredicate; body |
| src/compiler/utilities.ts:3117:60 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/utilities.ts:3402:46 | predicate | Refused; isSuperProperty; body |
| src/compiler/utilities.ts:3773:70 | predicate | Refused; isExternalModuleImportEqualsDeclaration; body |
| src/compiler/utilities.ts:3789:70 | predicate | Refused; isInternalModuleImportEqualsDeclaration; body |
| src/compiler/utilities.ts:3794:55 | predicate | Refused; isFullSourceFile; body |
| src/compiler/utilities.ts:3839:94 | signature | Refused; isRequireCall; no body on declaration |
| src/compiler/utilities.ts:3841:97 | signature | Refused; isRequireCall; no body on declaration |
| src/compiler/utilities.ts:3843:97 | predicate | Refused; isRequireCall; body |
| src/compiler/utilities.ts:3866:72 | predicate | Refused; isVariableDeclarationInitializedToRequire; body |
| src/compiler/utilities.ts:3875:86 | predicate | Refused; isVariableDeclarationInitializedToBareOrAccessedRequire; body |
| src/compiler/utilities.ts:3880:70 | predicate | Refused; isBindingElementOfBareOrAccessedRequire; body |
| src/compiler/utilities.ts:3905:57 | predicate | Refused; isRequireVariableStatement; body |
| src/compiler/utilities.ts:4104:62 | predicate | Refused; isModuleExportsAccessExpression; body |
| src/compiler/utilities.ts:4119:75 | predicate | Refused; isBindableObjectDefinePropertyCall; body |
| src/compiler/utilities.ts:4132:43 | predicate | Refused; isLiteralLikeAccess; body |
| src/compiler/utilities.ts:4139:50 | predicate | Refused; isLiteralLikeElementAccess; body |
| src/compiler/utilities.ts:4148:93 | predicate | Refused; isBindableStaticAccessExpression; body |
| src/compiler/utilities.ts:4158:100 | predicate | Refused; isBindableStaticElementAccessExpression; body |
| src/compiler/utilities.ts:4166:91 | predicate | Refused; isBindableStaticNameExpression; body |
| src/compiler/utilities.ts:4294:60 | predicate | Refused; isPrototypePropertyAssignment; body |
| src/compiler/utilities.ts:4299:105 | predicate | Refused; isSpecialPropertyDeclaration; body |
| src/compiler/utilities.ts:4330:65 | predicate | Refused; canHaveModuleSpecifier; body |
| src/compiler/utilities.ts:4353:59 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/utilities.ts:4482:47 | predicate | Refused; isJSDocTypeAlias; body |
| src/compiler/utilities.ts:4487:42 | predicate | Refused; isTypeAlias; body |
| src/compiler/utilities.ts:4535:46 | predicate | Refused; canHaveFlowNode; body |
| src/compiler/utilities.ts:4561:43 | predicate | Refused; canHaveJSDoc; body |
| src/compiler/utilities.ts:4837:47 | predicate | Refused; hasTypeArguments; body |
| src/compiler/utilities.ts:4964:67 | predicate | Refused; isNodeWithPossibleHoistedDeclaration; body |
| src/compiler/utilities.ts:4997:58 | predicate | Refused; isValueSignatureDeclaration; body |
| src/compiler/utilities.ts:5266:47 | predicate | Refused; isKeyword; body |
| src/compiler/utilities.ts:5271:51 | predicate | Refused; isPunctuation; body |
| src/compiler/utilities.ts:5276:60 | predicate | Refused; isKeywordOrPunctuation; body |
| src/compiler/utilities.ts:5303:46 | predicate | Refused; isTrivia; body |
| src/compiler/utilities.ts:5362:59 | predicate | Refused; isStringOrNumericLiteralLike; body |
| src/compiler/utilities.ts:5367:53 | predicate | Refused; isSignedNumericLiteral; body |
| src/compiler/utilities.ts:5381:59 | predicate | Refused; hasDynamicName; body |
| src/compiler/utilities.ts:5427:52 | predicate | Refused; isPropertyNameLiteral; body |
| src/compiler/utilities.ts:5485:112 | predicate | Refused; isAnonymousFunctionDefinition; body |
| src/compiler/utilities.ts:5522:54 | predicate | Refused; isNamedEvaluationSource; body |
| src/compiler/utilities.ts:5564:101 | predicate | Refused; isNamedEvaluation; body |
| src/compiler/utilities.ts:6760:41 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/utilities.ts:6923:49 | predicate | Refused; isNonTypeAliasTemplate; body |
| src/compiler/utilities.ts:7413:77 | predicate | Refused; isLogicalOrCoalescingAssignmentOperator; body |
| src/compiler/utilities.ts:7420:72 | predicate | Refused; isLogicalOrCoalescingAssignmentExpression; body |
| src/compiler/utilities.ts:7425:73 | predicate | Refused; isLogicalOrCoalescingBinaryOperator; body |
| src/compiler/utilities.ts:7430:68 | predicate | Refused; isLogicalOrCoalescingBinaryExpression; body |
| src/compiler/utilities.ts:7471:86 | signature | Refused; isAssignmentExpression; no body on declaration |
| src/compiler/utilities.ts:7473:88 | signature | Refused; isAssignmentExpression; no body on declaration |
| src/compiler/utilities.ts:7475:90 | predicate | Refused; isAssignmentExpression; body |
| src/compiler/utilities.ts:7484:56 | predicate | Refused; isDestructuringAssignment; body |
| src/compiler/utilities.ts:7495:80 | predicate | Refused; isExpressionWithTypeArgumentsInClassExtendsClause; body |
| src/compiler/utilities.ts:7500:53 | predicate | Refused; isEntityNameExpression; body |
| src/compiler/utilities.ts:7535:67 | predicate | Refused; isPropertyAccessEntityNameExpression; body |
| src/compiler/utilities.ts:7563:48 | predicate | Refused; isPrototypeAccess; body |
| src/compiler/utilities.ts:7587:53 | predicate | Refused; isInstanceOfExpression; body |
| src/compiler/utilities.ts:8024:52 | predicate | Refused; isInitializedVariable; body |
| src/compiler/utilities.ts:8325:54 | predicate | Refused; isObjectTypeDeclaration; body |
| src/compiler/utilities.ts:8330:51 | predicate | Refused; isTypeNodeKind; body |
| src/compiler/utilities.ts:8355:49 | predicate | Refused; isAccessExpression; body |
| src/compiler/utilities.ts:8369:54 | predicate | Refused; isNamedImportsOrExports; body |
| src/compiler/utilities.ts:8637:119 | predicate | Refused; isDiagnosticWithDetachedLocation; body |
| src/compiler/utilities.ts:9028:62 | predicate | Refused; usesWildcardTypes; body |
| src/compiler/utilities.ts:10627:56 | predicate | Refused; isIdentifierTypeReference; body |
| src/compiler/utilities.ts:10913:66 | predicate | Refused; isFunctionExpressionOrArrowFunction; body |
| src/compiler/utilities.ts:11028:48 | predicate | Refused; isTypeDeclaration; body |
| src/compiler/utilities.ts:11051:52 | predicate | Refused; canHaveExportModifier; body |
| src/compiler/utilities.ts:11107:46 | predicate | Refused; isNonNullAccess; body |
| src/compiler/utilities.ts:11114:57 | predicate | Refused; isJSDocSatisfiesExpression; body |
| src/compiler/utilities.ts:11140:49 | predicate | Refused; isJsxAttributeName; body |
| src/compiler/utilities.ts:11165:57 | predicate | Refused; isTypeUsableAsPropertyName; body |
| src/compiler/utilities.ts:11184:85 | predicate | Refused; isExpandoPropertyDeclaration; body |
| src/compiler/utilities.ts:12000:83 | predicate | Refused; isSelfReferenceLocation; body |
| src/compiler/utilities.ts:12033:82 | predicate | Refused; isPrimitiveLiteralValue; body |
| src/compiler/utilities.ts:12067:46 | predicate | Refused; hasInferredType; body |
| src/compiler/utilities.ts:12221:45 | predicate | Refused; isNewScopeNode; body |
| src/compiler/utilities.ts:12301:60 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/utilities.ts:12434:48 | predicate | Refused; canHaveStatements; body |
| src/compiler/utilitiesPublic.ts:615:75 | predicate | Refused; isParameterPropertyDeclaration; body |
| src/compiler/utilitiesPublic.ts:619:59 | predicate | Refused; isEmptyBindingPattern; body |
| src/compiler/utilitiesPublic.ts:763:87 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/utilitiesPublic.ts:765:99 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/utilitiesPublic.ts:766:100 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/utilitiesPublic.ts:786:99 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/utilitiesPublic.ts:826:98 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/utilitiesPublic.ts:946:49 | predicate | Refused; isNamedDeclaration; body |
| src/compiler/utilitiesPublic.ts:1040:76 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/utilitiesPublic.ts:1078:68 | predicate | Refused; <anonymous/signature>; body |
| src/compiler/utilitiesPublic.ts:1279:89 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/utilitiesPublic.ts:1284:95 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/utilitiesPublic.ts:1359:43 | predicate | Refused; isMemberName; body |
| src/compiler/utilitiesPublic.ts:1364:60 | predicate | Refused; isGetOrSetAccessorDeclaration; body |
| src/compiler/utilitiesPublic.ts:1368:52 | predicate | Refused; isPropertyAccessChain; body |
| src/compiler/utilitiesPublic.ts:1372:51 | predicate | Refused; isElementAccessChain; body |
| src/compiler/utilitiesPublic.ts:1376:42 | predicate | Refused; isCallChain; body |
| src/compiler/utilitiesPublic.ts:1380:46 | predicate | Refused; isOptionalChain; body |
| src/compiler/utilitiesPublic.ts:1390:50 | predicate | Refused; isOptionalChainRoot; body |
| src/compiler/utilitiesPublic.ts:1399:62 | predicate | Refused; isExpressionOfOptionalChainRoot; body |
| src/compiler/utilitiesPublic.ts:1437:45 | predicate | Refused; isNonNullChain; body |
| src/compiler/utilitiesPublic.ts:1441:57 | predicate | Refused; isBreakOrContinueStatement; body |
| src/compiler/utilitiesPublic.ts:1445:52 | predicate | Refused; isNamedExportBindings; body |
| src/compiler/utilitiesPublic.ts:1449:53 | predicate | Refused; isJSDocPropertyLikeTag; body |
| src/compiler/utilitiesPublic.ts:1487:67 | predicate | Refused; isNodeArray; body |
| src/compiler/utilitiesPublic.ts:1494:50 | predicate | Refused; isLiteralKind; body |
| src/compiler/utilitiesPublic.ts:1498:50 | predicate | Refused; isLiteralExpression; body |
| src/compiler/utilitiesPublic.ts:1518:58 | predicate | Refused; isTemplateLiteralKind; body |
| src/compiler/utilitiesPublic.ts:1522:53 | predicate | Refused; isTemplateLiteralToken; body |
| src/compiler/utilitiesPublic.ts:1526:61 | predicate | Refused; isTemplateMiddleOrTemplateTail; body |
| src/compiler/utilitiesPublic.ts:1532:56 | predicate | Refused; isImportOrExportSpecifier; body |
| src/compiler/utilitiesPublic.ts:1536:58 | predicate | Refused; isTypeOnlyImportDeclaration; body |
| src/compiler/utilitiesPublic.ts:1550:58 | predicate | Refused; isTypeOnlyExportDeclaration; body |
| src/compiler/utilitiesPublic.ts:1562:66 | predicate | Refused; isTypeOnlyImportOrExportDeclaration; body |
| src/compiler/utilitiesPublic.ts:1570:57 | predicate | Refused; isStringTextContainingNode; body |
| src/compiler/utilitiesPublic.ts:1574:52 | predicate | Refused; isImportAttributeName; body |
| src/compiler/utilitiesPublic.ts:1581:52 | predicate | Refused; isGeneratedIdentifier; body |
| src/compiler/utilitiesPublic.ts:1586:59 | predicate | Refused; isGeneratedPrivateIdentifier; body |
| src/compiler/utilitiesPublic.ts:1600:73 | predicate | Refused; isPrivateIdentifierClassElementDeclaration; body |
| src/compiler/utilitiesPublic.ts:1605:74 | predicate | Refused; isPrivateIdentifierPropertyAccessExpression; body |
| src/compiler/utilitiesPublic.ts:1612:52 | predicate | Refused; isModifierKind; body |
| src/compiler/utilitiesPublic.ts:1647:41 | predicate | Refused; isModifier; body |
| src/compiler/utilitiesPublic.ts:1651:43 | predicate | Refused; isEntityName; body |
| src/compiler/utilitiesPublic.ts:1657:45 | predicate | Refused; isPropertyName; body |
| src/compiler/utilitiesPublic.ts:1666:44 | predicate | Refused; isBindingName; body |
| src/compiler/utilitiesPublic.ts:1675:57 | predicate | Refused; isFunctionLike; body |
| src/compiler/utilitiesPublic.ts:1680:86 | predicate | Refused; isFunctionLikeOrClassStaticBlockDeclaration; body |
| src/compiler/utilitiesPublic.ts:1685:56 | predicate | Refused; isFunctionLikeDeclaration; body |
| src/compiler/utilitiesPublic.ts:1690:47 | predicate | Refused; isBooleanLiteral; body |
| src/compiler/utilitiesPublic.ts:1732:45 | predicate | Refused; isClassElement; body |
| src/compiler/utilitiesPublic.ts:1744:42 | predicate | Refused; isClassLike; body |
| src/compiler/utilitiesPublic.ts:1748:41 | predicate | Refused; isAccessor; body |
| src/compiler/utilitiesPublic.ts:1752:64 | predicate | Refused; isAutoAccessorPropertyDeclaration; body |
| src/compiler/utilitiesPublic.ts:1765:49 | predicate | Refused; isMethodOrAccessor; body |
| src/compiler/utilitiesPublic.ts:1778:45 | predicate | Refused; isModifierLike; body |
| src/compiler/utilitiesPublic.ts:1782:44 | predicate | Refused; isTypeElement; body |
| src/compiler/utilitiesPublic.ts:1794:51 | predicate | Refused; isClassOrTypeElement; body |
| src/compiler/utilitiesPublic.ts:1798:57 | predicate | Refused; isObjectLiteralElementLike; body |
| src/compiler/utilitiesPublic.ts:1815:41 | predicate | Refused; isTypeNode; body |
| src/compiler/utilitiesPublic.ts:1819:62 | predicate | Refused; isFunctionOrConstructorTypeNode; body |
| src/compiler/utilitiesPublic.ts:1832:59 | predicate | Refused; isBindingPattern; body |
| src/compiler/utilitiesPublic.ts:1843:50 | predicate | Refused; isAssignmentPattern; body |
| src/compiler/utilitiesPublic.ts:1849:52 | predicate | Refused; isArrayBindingElement; body |
| src/compiler/utilitiesPublic.ts:1860:90 | predicate | Refused; isDeclarationBindingElement; body |
| src/compiler/utilitiesPublic.ts:1872:59 | predicate | Refused; isBindingOrAssignmentElement; body |
| src/compiler/utilitiesPublic.ts:1884:87 | predicate | Refused; isBindingOrAssignmentPattern; body |
| src/compiler/utilitiesPublic.ts:1894:93 | predicate | Refused; isObjectBindingOrAssignmentPattern; body |
| src/compiler/utilitiesPublic.ts:1905:65 | predicate | Refused; isObjectBindingOrAssignmentElement; body |
| src/compiler/utilitiesPublic.ts:1921:92 | predicate | Refused; isArrayBindingOrAssignmentPattern; body |
| src/compiler/utilitiesPublic.ts:1932:64 | predicate | Refused; isArrayBindingOrAssignmentElement; body |
| src/compiler/utilitiesPublic.ts:1948:78 | predicate | Refused; isPropertyAccessOrQualifiedNameOrImportTypeNode; body |
| src/compiler/utilitiesPublic.ts:1957:62 | predicate | Refused; isPropertyAccessOrQualifiedName; body |
| src/compiler/utilitiesPublic.ts:1964:65 | predicate | Refused; isCallLikeOrFunctionLikeExpression; body |
| src/compiler/utilitiesPublic.ts:1968:51 | predicate | Refused; isCallLikeExpression; body |
| src/compiler/utilitiesPublic.ts:1985:52 | predicate | Refused; isCallOrNewExpression; body |
| src/compiler/utilitiesPublic.ts:1989:48 | predicate | Refused; isTemplateLiteral; body |
| src/compiler/utilitiesPublic.ts:1995:55 | predicate | Refused; isLeftHandSideExpression; body |
| src/compiler/utilitiesPublic.ts:2039:48 | predicate | Refused; isUnaryExpression; body |
| src/compiler/utilitiesPublic.ts:2059:57 | predicate | Refused; isUnaryExpressionWithWrite; body |
| src/compiler/utilitiesPublic.ts:2071:51 | predicate | Refused; isLiteralTypeLiteral; body |
| src/compiler/utilitiesPublic.ts:2086:43 | predicate | Refused; isExpression; body |
| src/compiler/utilitiesPublic.ts:2108:52 | predicate | Refused; isAssertionExpression; body |
| src/compiler/utilitiesPublic.ts:2116:83 | signature | Refused; isIterationStatement; no body on declaration |
| src/compiler/utilitiesPublic.ts:2117:85 | signature | Refused; isIterationStatement; no body on declaration |
| src/compiler/utilitiesPublic.ts:2118:85 | predicate | Refused; isIterationStatement; body |
| src/compiler/utilitiesPublic.ts:2154:51 | predicate | Refused; isForInOrOfStatement; body |
| src/compiler/utilitiesPublic.ts:2160:44 | predicate | Refused; isConciseBody; body |
| src/compiler/utilitiesPublic.ts:2166:45 | predicate | Refused; isFunctionBody; body |
| src/compiler/utilitiesPublic.ts:2170:47 | predicate | Refused; isForInitializer; body |
| src/compiler/utilitiesPublic.ts:2175:43 | predicate | Refused; isModuleBody; body |
| src/compiler/utilitiesPublic.ts:2183:46 | predicate | Refused; isNamespaceBody; body |
| src/compiler/utilitiesPublic.ts:2190:51 | predicate | Refused; isJSDocNamespaceBody; body |
| src/compiler/utilitiesPublic.ts:2196:52 | predicate | Refused; isNamedImportBindings; body |
| src/compiler/utilitiesPublic.ts:2203:56 | predicate | Refused; isModuleOrEnumDeclaration; body |
| src/compiler/utilitiesPublic.ts:2208:44 | predicate | Refused; canHaveSymbol; body |
| src/compiler/utilitiesPublic.ts:2283:44 | predicate | Refused; canHaveLocals; body |
| src/compiler/utilitiesPublic.ts:2398:44 | predicate | Refused; isDeclaration; body |
| src/compiler/utilitiesPublic.ts:2406:53 | predicate | Refused; isDeclarationStatement; body |
| src/compiler/utilitiesPublic.ts:2415:59 | predicate | Refused; isStatementButNotDeclaration; body |
| src/compiler/utilitiesPublic.ts:2419:42 | predicate | Refused; isStatement; body |
| src/compiler/utilitiesPublic.ts:2426:40 | predicate | Refused; isBlockStatement; body |
| src/compiler/utilitiesPublic.ts:2442:49 | predicate | Refused; isStatementOrBlock; body |
| src/compiler/utilitiesPublic.ts:2451:48 | predicate | Refused; isModuleReference; body |
| src/compiler/utilitiesPublic.ts:2460:53 | predicate | Refused; isJsxTagNameExpression; body |
| src/compiler/utilitiesPublic.ts:2468:41 | predicate | Refused; isJsxChild; body |
| src/compiler/utilitiesPublic.ts:2477:49 | predicate | Refused; isJsxAttributeLike; body |
| src/compiler/utilitiesPublic.ts:2483:61 | predicate | Refused; isStringLiteralOrJsxExpression; body |
| src/compiler/utilitiesPublic.ts:2489:54 | predicate | Refused; isJsxOpeningLikeElement; body |
| src/compiler/utilitiesPublic.ts:2495:44 | predicate | Refused; isJsxCallLike; body |
| src/compiler/utilitiesPublic.ts:2504:52 | predicate | Refused; isCaseOrDefaultClause; body |
| src/compiler/utilitiesPublic.ts:2534:41 | predicate | Refused; isJSDocTag; body |
| src/compiler/utilitiesPublic.ts:2538:44 | predicate | Refused; isSetAccessor; body |
| src/compiler/utilitiesPublic.ts:2542:44 | predicate | Refused; isGetAccessor; body |
| src/compiler/utilitiesPublic.ts:2552:44 | predicate | Refused; hasJSDocNodes; body |
| src/compiler/utilitiesPublic.ts:2564:38 | predicate | Refused; hasType; body |
| src/compiler/utilitiesPublic.ts:2573:45 | predicate | Refused; hasInitializer; body |
| src/compiler/utilitiesPublic.ts:2578:59 | predicate | Refused; hasOnlyExpressionInitializer; body |
| src/compiler/utilitiesPublic.ts:2592:53 | predicate | Refused; isObjectLiteralElement; body |
| src/compiler/utilitiesPublic.ts:2597:50 | predicate | Refused; isTypeReferenceType; body |
| src/compiler/utilitiesPublic.ts:2625:66 | predicate | Refused; isStringLiteralLike; body |
| src/compiler/utilitiesPublic.ts:2629:46 | predicate | Refused; isJSDocLinkLike; body |
| src/compiler/visitorPublic.ts:125:27 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/visitorPublic.ts:197:27 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/visitorPublic.ts:277:27 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/visitorPublic.ts:317:27 | signature | Refused; <anonymous/signature>; no body on declaration |
| src/compiler/watch.ts:334:96 | predicate | Refused; isBuilderProgram; body |
| src/compiler/watchPublic.ts:739:77 | proven | Proven; isFileMissingOnHost; body |
| src/compiler/watchPublic.ts:743:83 | predicate | Refused; isFilePresenceUnknownOnHost; body |
