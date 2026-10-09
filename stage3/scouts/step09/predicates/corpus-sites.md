# All 580 predicate bodies

Coordinates and original declarations come from 8a7ab17e. The measured compiler is 90a78f49. Logical proof and .a diagnostic categories are separate; .ts states describe the call seam, and NoResolvedCall supplies no pass. The group labels are the approved original AST partition. Full call locations, exact failures and the 54 constructed check sites are in [corpus-report.md](corpus-report.md) and [evidence/corpus-results.json.gz](evidence/corpus-results.json.gz).

| Annotation | Name | Group | Logical body proof | .a diagnostic | .ts call result | Constructed checked calls |
| --- | --- | --- | --- | --- | --- | ---: |
| `builder.ts:268:87` | `isBuilderProgramStateWithDefinedProgram` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `builder.ts:1176:79` | `isIncrementalBundleEmitBuildInfo` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `builder.ts:1181:58` | `isIncrementalBuildInfo` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `builder.ts:1192:54` | `isNonIncrementalBuildInfo` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:4290:132` | `isNonLocalAlias` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:7873:81` | `<anonymous/signature>` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `checker.ts:9408:75` | `isIdentifierAndNotUndefined` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `checker.ts:13387:43` | `isValidBaseType` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:13778:57` | `isLateBindableName` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:13783:67` | `isLateBindableIndexSignature` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:13809:54` | `hasLateBindableName` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:14997:47` | `isGenericMappedType` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:17304:48` | `isJSDocTypeReference` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:21245:75` | `isContextSensitiveFunctionOrObjectLiteralMethod` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:25116:54` | `isNonDeferredTypeReference` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:25515:39` | `isArrayType` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:25523:46` | `isArrayOrTupleType` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:25707:39` | `isTupleType` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:25711:46` | `isGenericTupleType` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:25715:59` | `isSingleElementGenericTupleType` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:26489:43` | `<anonymous/signature>` | other value, generic, assertion or erased claims | unproved | Refused | NoResolvedCall | 0 |
| `checker.ts:32636:54` | `<anonymous/signature>` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `checker.ts:33802:52` | `isJsxIntrinsicTagName` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:35331:56` | `<anonymous/signature>` | flags and bit masks | unproved | Refused | NoResolvedCall | 0 |
| `checker.ts:35569:80` | `callLikeExpressionMayHaveTypeArguments` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:35656:61` | `isSpreadArgument` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `checker.ts:37696:55` | `isJSConstructor` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `checker.ts:38480:63` | `isValidDeclarationForTupleLabel` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `checker.ts:43611:54` | `isAwaitedTypeInstantiation` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `checker.ts:44756:49` | `isImportedDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `checker.ts:47505:62` | `<anonymous/signature>` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `checker.ts:50641:59` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `checker.ts:50886:57` | `<anonymous/signature>` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `checker.ts:51062:48` | `canHaveConstantValue` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `checker.ts:51437:49` | `<anonymous/signature>` | flags and bit masks | unproved | Refused | NoResolvedCall | 0 |
| `commandLineParser.ts:1728:70` | `isCommandLineOptionOfCustomType` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `commandLineParser.ts:2606:85` | `isCompilerOptionsValue` | delegation or composition of predicate calls | unproved | Refused | PendingOther | 0 |
| `commandLineParser.ts:3038:37` | `isNullOrUndefined` | other value, generic, assertion or erased claims | unproved | Refused | PendingOther | 0 |
| `commandLineParser.ts:3300:51` | `startsWithConfigDirTemplate` | delegation or composition of predicate calls | unproved | Refused | PendingOther | 0 |
| `core.ts:1750:42` | `isArray` | delegation or composition of predicate calls | proved | Proven | Proven | 0 |
| `core.ts:1769:42` | `isString` | other value, generic, assertion or erased claims | proved | Proven | Proven | 0 |
| `core.ts:1773:39` | `isNumber` | other value, generic, assertion or erased claims | proved | Proven | Proven | 0 |
| `debug.ts:213:142` | `assert` | other value, generic, assertion or erased claims | proved | Proven | ProvenBodyPendingCalls | 0 |
| `debug.ts:248:99` | `assertIsDefined` | other value, generic, assertion or erased claims | unproved | Refused | PendingOther | 0 |
| `debug.ts:958:51` | `isFlowSwitchClause` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `debug.ts:962:47` | `hasAntecedents` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `debug.ts:966:46` | `hasAntecedent` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `debug.ts:970:40` | `hasNode` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `emitter.ts:1158:27` | `<anonymous/signature>` | other value, generic, assertion or erased claims | unproved | Refused | NoResolvedCall | 0 |
| `factory/nodeTests.ts:235:47` | `isNumericLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:239:46` | `isBigIntLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:243:46` | `isStringLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:247:40` | `isJsxText` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:251:57` | `isRegularExpressionLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:255:62` | `isNoSubstitutionTemplateLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:261:45` | `isTemplateHead` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:265:47` | `isTemplateMiddle` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:269:45` | `isTemplateTail` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:275:47` | `isDotDotDotToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:280:43` | `isCommaToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:284:42` | `isPlusToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:288:43` | `isMinusToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:292:46` | `isAsteriskToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:296:49` | `isExclamationToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:300:46` | `isQuestionToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:304:43` | `isColonToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:308:49` | `isQuestionDotToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:312:55` | `isEqualsGreaterThanToken` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:318:43` | `isIdentifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:322:50` | `isPrivateIdentifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:329:47` | `isExportModifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:334:48` | `isDefaultModifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:339:46` | `isAsyncModifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:343:47` | `isAssertsKeyword` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:347:45` | `isAwaitKeyword` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:352:48` | `isReadonlyKeyword` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:357:47` | `isStaticModifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:362:49` | `isAbstractModifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:367:49` | `isOverrideModifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:372:49` | `isAccessorModifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:377:45` | `isSuperKeyword` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:382:46` | `isImportKeyword` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:387:44` | `isCaseKeyword` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:393:46` | `isQualifiedName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:397:53` | `isComputedPropertyName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:403:57` | `isTypeParameterDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:408:42` | `isParameter` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:412:42` | `isDecorator` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:418:50` | `isPropertySignature` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:422:52` | `isPropertyDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:426:48` | `isMethodSignature` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:430:50` | `isMethodDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:434:60` | `isClassStaticBlockDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:438:55` | `isConstructorDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:442:55` | `isGetAccessorDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:446:55` | `isSetAccessorDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:450:57` | `isCallSignatureDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:454:62` | `isConstructSignatureDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:458:58` | `isIndexSignatureDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:464:50` | `isTypePredicateNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:468:50` | `isTypeReferenceNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:472:49` | `isFunctionTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:476:52` | `isConstructorTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:480:46` | `isTypeQueryNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:484:48` | `isTypeLiteralNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:488:46` | `isArrayTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:492:46` | `isTupleTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:496:49` | `isNamedTupleMember` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:500:49` | `isOptionalTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:504:45` | `isRestTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:508:46` | `isUnionTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:512:53` | `isIntersectionTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:516:52` | `isConditionalTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:520:46` | `isInferTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:524:54` | `isParenthesizedTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:528:45` | `isThisTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:532:49` | `isTypeOperatorNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:536:54` | `isIndexedAccessTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:540:47` | `isMappedTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:544:48` | `isLiteralTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:548:47` | `isImportTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:552:56` | `isTemplateLiteralTypeSpan` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:556:56` | `isTemplateLiteralTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:562:53` | `isObjectBindingPattern` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:566:52` | `isArrayBindingPattern` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:570:47` | `isBindingElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:576:55` | `isArrayLiteralExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:580:56` | `isObjectLiteralExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:584:57` | `isPropertyAccessExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:588:56` | `isElementAccessExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:592:47` | `isCallExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:596:46` | `isNewExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:600:57` | `isTaggedTemplateExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:604:56` | `isTypeAssertionExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:608:56` | `isParenthesizedExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:612:51` | `isFunctionExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:616:46` | `isArrowFunction` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:620:49` | `isDeleteExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:624:49` | `isTypeOfExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:628:47` | `isVoidExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:632:48` | `isAwaitExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:636:54` | `isPrefixUnaryExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:640:55` | `isPostfixUnaryExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:644:49` | `isBinaryExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:648:54` | `isConditionalExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:652:51` | `isTemplateExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:656:48` | `isYieldExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:660:46` | `isSpreadElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:664:48` | `isClassExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:668:50` | `isOmittedExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:672:60` | `isExpressionWithTypeArguments` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:676:45` | `isAsExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:680:52` | `isSatisfiesExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:684:50` | `isNonNullExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:688:45` | `isMetaProperty` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:692:52` | `isSyntheticExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:696:59` | `isPartiallyEmittedExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:700:52` | `isCommaListExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:706:45` | `isTemplateSpan` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:710:54` | `isSemicolonClassElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:716:38` | `isBlock` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:720:50` | `isVariableStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:724:47` | `isEmptyStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:728:52` | `isExpressionStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:732:44` | `isIfStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:736:44` | `isDoStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:740:47` | `isWhileStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:744:45` | `isForStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:748:47` | `isForInStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:752:47` | `isForOfStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:756:50` | `isContinueStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:760:47` | `isBreakStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:764:48` | `isReturnStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:768:46` | `isWithStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:772:48` | `isSwitchStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:776:49` | `isLabeledStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:780:47` | `isThrowStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:784:45` | `isTryStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:788:50` | `isDebuggerStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:792:52` | `isVariableDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:796:56` | `isVariableDeclarationList` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:800:52` | `isFunctionDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:804:49` | `isClassDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:808:53` | `isInterfaceDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:812:53` | `isTypeAliasDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:816:48` | `isEnumDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:820:50` | `isModuleDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:824:44` | `isModuleBlock` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:828:42` | `isCaseBlock` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:832:59` | `isNamespaceExportDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:836:56` | `isImportEqualsDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:840:50` | `isImportDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:844:45` | `isImportClause` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:848:61` | `isImportTypeAssertionContainer` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:853:45` | `isAssertClause` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:858:44` | `isAssertEntry` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:862:49` | `isImportAttributes` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:866:48` | `isImportAttribute` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:870:48` | `isNamespaceImport` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:874:48` | `isNamespaceExport` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:878:45` | `isNamedImports` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:882:48` | `isImportSpecifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:886:49` | `isExportAssignment` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:890:50` | `isExportDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:894:45` | `isNamedExports` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:898:48` | `isExportSpecifier` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:902:49` | `isModuleExportName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:906:51` | `isMissingDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:910:52` | `isNotEmittedStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:915:51` | `isSyntheticReference` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:921:56` | `isExternalModuleReference` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:927:43` | `isJsxElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:931:54` | `isJsxSelfClosingElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:935:50` | `isJsxOpeningElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:939:50` | `isJsxClosingElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:943:44` | `isJsxFragment` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:947:51` | `isJsxOpeningFragment` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:951:51` | `isJsxClosingFragment` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:955:45` | `isJsxAttribute` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:959:46` | `isJsxAttributes` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:963:51` | `isJsxSpreadAttribute` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:967:46` | `isJsxExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:971:50` | `isJsxNamespacedName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:977:43` | `isCaseClause` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:981:46` | `isDefaultClause` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:985:47` | `isHeritageClause` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:989:44` | `isCatchClause` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:995:51` | `isPropertyAssignment` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:999:60` | `isShorthandPropertyAssignment` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1003:49` | `isSpreadAssignment` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1009:43` | `isEnumMember` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1014:43` | `isSourceFile` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1018:39` | `isBundle` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1026:52` | `isJSDocTypeExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1030:51` | `isJSDocNameReference` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1034:48` | `isJSDocMemberName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1038:42` | `isJSDocLink` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1042:46` | `isJSDocLinkCode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1046:47` | `isJSDocLinkPlain` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1050:45` | `isJSDocAllType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1054:49` | `isJSDocUnknownType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1058:50` | `isJSDocNullableType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1062:53` | `isJSDocNonNullableType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1066:50` | `isJSDocOptionalType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1070:50` | `isJSDocFunctionType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1074:50` | `isJSDocVariadicType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1078:50` | `isJSDocNamepathType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1082:38` | `isJSDoc` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1086:49` | `isJSDocTypeLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1090:47` | `isJSDocSignature` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1096:49` | `isJSDocAugmentsTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1100:47` | `isJSDocAuthorTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1104:46` | `isJSDocClassTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1108:49` | `isJSDocCallbackTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1112:47` | `isJSDocPublicTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1116:48` | `isJSDocPrivateTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1120:50` | `isJSDocProtectedTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1124:49` | `isJSDocReadonlyTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1128:49` | `isJSDocOverrideTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1132:49` | `isJSDocOverloadTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1136:51` | `isJSDocDeprecatedTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1140:44` | `isJSDocSeeTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1144:45` | `isJSDocEnumTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1148:50` | `isJSDocParameterTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1152:47` | `isJSDocReturnTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1156:45` | `isJSDocThisTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1160:45` | `isJSDocTypeTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1164:49` | `isJSDocTemplateTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1168:48` | `isJSDocTypedefTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1172:48` | `isJSDocUnknownTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1176:49` | `isJSDocPropertyTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1180:51` | `isJSDocImplementsTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1184:50` | `isJSDocSatisfiesTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1188:47` | `isJSDocThrowsTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1192:47` | `isJSDocImportTag` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/nodeTests.ts:1199:40` | `isSyntaxList` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:607:54` | `isCommaExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `factory/utilities.ts:612:52` | `isCommaSequence` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `factory/utilities.ts:617:51` | `isJSDocTypeAssertion` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `factory/utilities.ts:631:104` | `isOuterExpression` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `factory/utilities.ts:1065:48` | `isStringOrNumericLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1105:49` | `canHaveIllegalType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1112:59` | `canHaveIllegalTypeParameters` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1120:55` | `canHaveIllegalDecorators` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1142:54` | `canHaveIllegalModifiers` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1151:59` | `isQuestionOrExclamationToken` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1155:57` | `isIdentifierOrThisTypeNode` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1159:66` | `isReadonlyKeywordOrPlusOrMinusToken` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1163:59` | `isQuestionOrPlusOrMinusToken` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1167:43` | `isModuleName` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `factory/utilities.ts:1171:54` | `isExponentiationOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1175:54` | `isMultiplicativeOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1181:62` | `isMultiplicativeOperatorOrHigher` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 1 |
| `factory/utilities.ts:1186:48` | `isAdditiveOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1191:56` | `isAdditiveOperatorOrHigher` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 1 |
| `factory/utilities.ts:1196:45` | `isShiftOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1203:60` | `isShiftOperatorOrHigher` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 2 |
| `factory/utilities.ts:1208:50` | `isRelationalOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1217:58` | `isRelationalOperatorOrHigher` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 1 |
| `factory/utilities.ts:1222:48` | `isEqualityOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1229:56` | `isEqualityOperatorOrHigher` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 1 |
| `factory/utilities.ts:1234:47` | `isBitwiseOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1240:55` | `isBitwiseOperatorOrHigher` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 1 |
| `factory/utilities.ts:1246:47` | `isLogicalOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1251:55` | `isLogicalOperatorOrHigher` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 1 |
| `factory/utilities.ts:1256:58` | `isAssignmentOperatorOrHigher` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 1 |
| `factory/utilities.ts:1262:46` | `isBinaryOperator` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 1 |
| `factory/utilities.ts:1267:52` | `isBinaryOperatorToken` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `factory/utilities.ts:1489:58` | `isExportOrDefaultKeywordKind` | property presence or structural reads | proved | Proven | Proven | 0 |
| `factory/utilities.ts:1494:56` | `isExportOrDefaultModifier` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `factory/utilities.ts:1695:64` | `isSyntheticParenthesizedExpression` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `factory/utilitiesPublic.ts:14:47` | `canHaveModifiers` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `factory/utilitiesPublic.ts:43:48` | `canHaveDecorators` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `moduleNameResolver.ts:925:82` | `isPackageJsonInfo` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `moduleNameResolver.ts:930:89` | `isMissingPackageJsonInfo` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `program.ts:1162:74` | `isReferencedFile` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `program.ts:1190:108` | `isReferenceFileLocation` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `sourcemap.ts:408:34` | `isRawSourceMap` | delegation or composition of predicate calls | unproved | Refused | PendingOther | 0 |
| `sourcemap.ts:630:52` | `isSourceMapping` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `sourcemap.ts:668:57` | `isSourceMappedPosition` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `transformers/classFields.ts:853:53` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `transformers/classFields.ts:3349:67` | `isPrivateIdentifierInExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/classFields.ts:3354:51` | `isStaticPropertyDeclaration` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/classFields.ts:3358:69` | `isStaticPropertyDeclarationOrClassStaticBlock` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/classThis.ts:72:57` | `isClassThisAssignmentBlock` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/declarations.ts:659:54` | `shouldPrintWithInitializer` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/declarations.ts:1891:49` | `canHaveLiteralInitializer` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/declarations.ts:1916:55` | `isPreservedDeclarationStatement` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/declarations.ts:1954:44` | `isProcessedComponent` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/declarations/diagnostics.ts:138:52` | `canProduceDiagnostics` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `transformers/declarations/diagnostics.ts:714:46` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `transformers/es2015.ts:1348:42` | `isCapturedThis` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1353:44` | `isSyntheticSuper` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1359:60` | `isThisCapturingVariableStatement` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1365:62` | `isThisCapturingVariableDeclaration` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1370:53` | `isThisCapturingAssignment` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1375:50` | `isTransformedSuperCall` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1386:62` | `isTransformedSuperCallWithFallback` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1393:47` | `isImplicitSuperCall` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1403:59` | `isImplicitSuperCallWithFallback` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1410:75` | `isThisCapturingTransformedSuperCallWithFallback` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1415:72` | `isThisCapturingImplicitSuperCallWithFallback` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:1419:54` | `isTransformedSuperCallLike` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:2029:122` | `shouldAddRestParameter` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:3379:80` | `shouldConvertInitializerOfForStatement` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:3383:78` | `shouldConvertConditionOfForStatement` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2015.ts:3387:80` | `shouldConvertIncrementorOfForStatement` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/es2017.ts:586:80` | `isVariableDeclarationListWithCollidingName` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `transformers/es2017.ts:1018:44` | `isSuperContainer` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `transformers/esnext.ts:774:54` | `isUsingVariableDeclarationList` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/generators.ts:2420:56` | `supportsUnlabeledBreak` | kind partitions with control flow or extra conditions | proved | Proven | Proven | 0 |
| `transformers/generators.ts:2430:64` | `supportsLabeledBreakOrContinue` | kind partitions with control flow or extra conditions | proved | Proven | Proven | 0 |
| `transformers/generators.ts:2439:59` | `supportsUnlabeledContinue` | kind partitions with control flow or extra conditions | proved | Proven | Proven | 0 |
| `transformers/module/system.ts:1360:63` | `shouldHoistForInitializer` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/namedEvaluation.ts:139:64` | `isClassNamedEvaluationHelperBlock` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `transformers/ts.ts:1054:49` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `transformers/ts.ts:1171:56` | `shouldAddTypeMetadata` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `transformers/ts.ts:1186:62` | `shouldAddReturnTypeMetadata` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `transformers/ts.ts:1197:62` | `shouldAddParamTypesMetadata` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/ts.ts:1297:93` | `shouldEmitFunctionLikeDeclaration` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `transformers/utilities.ts:490:61` | `isCompoundAssignment` | property presence or structural reads | unproved | Refused | CheckedSeam | 5 |
| `transformers/utilities.ts:593:91` | `isStaticPropertyDeclarationOrClassStaticBlockDeclaration` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `transformers/utilities.ts:630:62` | `isInitializedProperty` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `transformers/utilities.ts:642:83` | `isNonStaticMethodOrAccessorWithPrivateName` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `tsbuildPublic.ts:259:66` | `isCircularBuildOrder` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `tsbuildPublic.ts:569:60` | `isParsedCommandLine` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `tsbuildPublic.ts:1371:84` | `isFileWatcherWithModifiedTime` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `utilities.ts:655:52` | `isTransientSymbol` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `utilities.ts:2078:46` | `isAmbientModule` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:2083:60` | `isModuleWithStringLiteralName` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:2088:55` | `isNonGlobalAmbientModule` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:2129:59` | `isExternalModuleAugmentation` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:2232:62` | `isDeclarationWithTypeParameters` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:2246:69` | `isDeclarationWithTypeParameterChildren` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:2276:48` | `isAnyImportSyntax` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:2287:65` | `isAnyImportOrBareOrAccessedRequire` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:2292:60` | `isAnyImportOrRequireStatement` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:2297:63` | `isLateVisibilityPaintedStatement` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:2315:65` | `hasPossibleExternalModuleReference` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:2320:52` | `isAnyImportOrReExport` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:2643:53` | `isJsonSourceFile` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `utilities.ts:2701:39` | `isSuperCall` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:2706:40` | `isImportCall` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:2717:40` | `isImportMeta` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:2724:51` | `isLiteralImportTypeNode` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:2729:50` | `isPrologueDirective` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:2996:45` | `isVariableLike` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3073:52` | `isObjectLiteralMethod` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3078:79` | `isObjectLiteralOrClassExpressionMethodOrAccessor` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3085:70` | `isIdentifierTypePredicate` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3090:64` | `isThisTypePredicate` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3117:60` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:3402:46` | `isSuperProperty` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3773:70` | `isExternalModuleImportEqualsDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3789:70` | `isInternalModuleImportEqualsDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3794:55` | `isFullSourceFile` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:3843:97` | `isRequireCall` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:3866:72` | `isVariableDeclarationInitializedToRequire` | other value, generic, assertion or erased claims | unproved | Refused | PendingView | 0 |
| `utilities.ts:3875:86` | `isVariableDeclarationInitializedToBareOrAccessedRequire` | other value, generic, assertion or erased claims | unproved | Refused | PendingView | 0 |
| `utilities.ts:3880:70` | `isBindingElementOfBareOrAccessedRequire` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:3905:57` | `isRequireVariableStatement` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:4104:62` | `isModuleExportsAccessExpression` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:4119:75` | `isBindableObjectDefinePropertyCall` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:4132:43` | `isLiteralLikeAccess` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:4139:50` | `isLiteralLikeElementAccess` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:4148:93` | `isBindableStaticAccessExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:4158:100` | `isBindableStaticElementAccessExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:4166:91` | `isBindableStaticNameExpression` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:4294:60` | `isPrototypePropertyAssignment` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:4299:105` | `isSpecialPropertyDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:4330:65` | `canHaveModuleSpecifier` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:4353:59` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:4482:47` | `isJSDocTypeAlias` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilities.ts:4487:42` | `isTypeAlias` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `utilities.ts:4535:46` | `canHaveFlowNode` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:4561:43` | `canHaveJSDoc` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:4837:47` | `hasTypeArguments` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `utilities.ts:4964:67` | `isNodeWithPossibleHoistedDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:4997:58` | `isValueSignatureDeclaration` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:5266:47` | `isKeyword` | property presence or structural reads | unproved | Refused | CheckedSeam | 11 |
| `utilities.ts:5271:51` | `isPunctuation` | property presence or structural reads | unproved | Refused | CheckedSeam | 1 |
| `utilities.ts:5276:60` | `isKeywordOrPunctuation` | delegation or composition of predicate calls | unproved | Refused | CheckedSeam | 3 |
| `utilities.ts:5303:46` | `isTrivia` | property presence or structural reads | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:5362:59` | `isStringOrNumericLiteralLike` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:5367:53` | `isSignedNumericLiteral` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:5381:59` | `hasDynamicName` | other value, generic, assertion or erased claims | unproved | Refused | PendingView | 0 |
| `utilities.ts:5427:52` | `isPropertyNameLiteral` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:5485:112` | `isAnonymousFunctionDefinition` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:5522:54` | `isNamedEvaluationSource` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:5564:101` | `isNamedEvaluation` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:6760:41` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:6923:49` | `isNonTypeAliasTemplate` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:7413:77` | `isLogicalOrCoalescingAssignmentOperator` | property presence or structural reads | proved | Proven | Proven | 0 |
| `utilities.ts:7420:72` | `isLogicalOrCoalescingAssignmentExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:7425:73` | `isLogicalOrCoalescingBinaryOperator` | property presence or structural reads | unproved | Refused | CheckedSeam | 3 |
| `utilities.ts:7430:68` | `isLogicalOrCoalescingBinaryExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:7475:90` | `isAssignmentExpression` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:7484:56` | `isDestructuringAssignment` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:7495:80` | `isExpressionWithTypeArgumentsInClassExtendsClause` | other value, generic, assertion or erased claims | unproved | Refused | PendingView | 0 |
| `utilities.ts:7500:53` | `isEntityNameExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:7535:67` | `isPropertyAccessEntityNameExpression` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:7563:48` | `isPrototypeAccess` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:7587:53` | `isInstanceOfExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:8024:52` | `isInitializedVariable` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:8325:54` | `isObjectTypeDeclaration` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:8330:51` | `isTypeNodeKind` | property presence or structural reads | unproved | Refused | CheckedSeam | 2 |
| `utilities.ts:8355:49` | `isAccessExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilities.ts:8369:54` | `isNamedImportsOrExports` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilities.ts:8637:119` | `isDiagnosticWithDetachedLocation` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `utilities.ts:9028:62` | `usesWildcardTypes` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:10627:56` | `isIdentifierTypeReference` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:10913:66` | `isFunctionExpressionOrArrowFunction` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilities.ts:11028:48` | `isTypeDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:11051:52` | `canHaveExportModifier` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:11107:46` | `isNonNullAccess` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:11114:57` | `isJSDocSatisfiesExpression` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:11140:49` | `isJsxAttributeName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilities.ts:11165:57` | `isTypeUsableAsPropertyName` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `utilities.ts:11184:85` | `isExpandoPropertyDeclaration` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:12000:83` | `isSelfReferenceLocation` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:12033:82` | `isPrimitiveLiteralValue` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:12067:46` | `hasInferredType` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilities.ts:12221:45` | `isNewScopeNode` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilities.ts:12301:60` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilities.ts:12434:48` | `canHaveStatements` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:615:75` | `isParameterPropertyDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:619:59` | `isEmptyBindingPattern` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:946:49` | `isNamedDeclaration` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1040:76` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:1078:68` | `<anonymous/signature>` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:1359:43` | `isMemberName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1364:60` | `isGetOrSetAccessorDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1368:52` | `isPropertyAccessChain` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1372:51` | `isElementAccessChain` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1376:42` | `isCallChain` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1380:46` | `isOptionalChain` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1390:50` | `isOptionalChainRoot` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1399:62` | `isExpressionOfOptionalChainRoot` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1437:45` | `isNonNullChain` | flags and bit masks | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1441:57` | `isBreakOrContinueStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1445:52` | `isNamedExportBindings` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1449:53` | `isJSDocPropertyLikeTag` | direct kind comparison | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1487:67` | `isNodeArray` | other value, generic, assertion or erased claims | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1494:50` | `isLiteralKind` | property presence or structural reads | unproved | Refused | CheckedSeam | 5 |
| `utilitiesPublic.ts:1498:50` | `isLiteralExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1518:58` | `isTemplateLiteralKind` | property presence or structural reads | unproved | Refused | CheckedSeam | 5 |
| `utilitiesPublic.ts:1522:53` | `isTemplateLiteralToken` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:1526:61` | `isTemplateMiddleOrTemplateTail` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1532:56` | `isImportOrExportSpecifier` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1536:58` | `isTypeOnlyImportDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1550:58` | `isTypeOnlyExportDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1562:66` | `isTypeOnlyImportOrExportDeclaration` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1570:57` | `isStringTextContainingNode` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:1574:52` | `isImportAttributeName` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1581:52` | `isGeneratedIdentifier` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1586:59` | `isGeneratedPrivateIdentifier` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1600:73` | `isPrivateIdentifierClassElementDeclaration` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1605:74` | `isPrivateIdentifierPropertyAccessExpression` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1612:52` | `isModifierKind` | property presence or structural reads | unproved | Refused | CheckedSeam | 9 |
| `utilitiesPublic.ts:1647:41` | `isModifier` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1651:43` | `isEntityName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1657:45` | `isPropertyName` | direct kind comparison | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1666:44` | `isBindingName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1675:57` | `isFunctionLike` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1680:86` | `isFunctionLikeOrClassStaticBlockDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1685:56` | `isFunctionLikeDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1690:47` | `isBooleanLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1732:45` | `isClassElement` | direct kind comparison | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1744:42` | `isClassLike` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1748:41` | `isAccessor` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1752:64` | `isAutoAccessorPropertyDeclaration` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1765:49` | `isMethodOrAccessor` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1778:45` | `isModifierLike` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1782:44` | `isTypeElement` | direct kind comparison | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1794:51` | `isClassOrTypeElement` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:1798:57` | `isObjectLiteralElementLike` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1815:41` | `isTypeNode` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1819:62` | `isFunctionOrConstructorTypeNode` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1832:59` | `isBindingPattern` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1843:50` | `isAssignmentPattern` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1849:52` | `isArrayBindingElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1860:90` | `isDeclarationBindingElement` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1872:59` | `isBindingOrAssignmentElement` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:1884:87` | `isBindingOrAssignmentPattern` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1894:93` | `isObjectBindingOrAssignmentPattern` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1905:65` | `isObjectBindingOrAssignmentElement` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1921:92` | `isArrayBindingOrAssignmentPattern` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1932:64` | `isArrayBindingOrAssignmentElement` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1948:78` | `isPropertyAccessOrQualifiedNameOrImportTypeNode` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1957:62` | `isPropertyAccessOrQualifiedName` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1964:65` | `isCallLikeOrFunctionLikeExpression` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:1968:51` | `isCallLikeExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:1985:52` | `isCallOrNewExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1989:48` | `isTemplateLiteral` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:1995:55` | `isLeftHandSideExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2039:48` | `isUnaryExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2059:57` | `isUnaryExpressionWithWrite` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2071:51` | `isLiteralTypeLiteral` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2086:43` | `isExpression` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2108:52` | `isAssertionExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2118:85` | `isIterationStatement` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2154:51` | `isForInOrOfStatement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2160:44` | `isConciseBody` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2166:45` | `isFunctionBody` | delegation or composition of predicate calls | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2170:47` | `isForInitializer` | delegation or composition of predicate calls | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2175:43` | `isModuleBody` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2183:46` | `isNamespaceBody` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2190:51` | `isJSDocNamespaceBody` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2196:52` | `isNamedImportBindings` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2203:56` | `isModuleOrEnumDeclaration` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2208:44` | `canHaveSymbol` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2283:44` | `canHaveLocals` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2398:44` | `isDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2406:53` | `isDeclarationStatement` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2415:59` | `isStatementButNotDeclaration` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2419:42` | `isStatement` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2426:40` | `isBlockStatement` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2442:49` | `isStatementOrBlock` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2451:48` | `isModuleReference` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2460:53` | `isJsxTagNameExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2468:41` | `isJsxChild` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2477:49` | `isJsxAttributeLike` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2483:61` | `isStringLiteralOrJsxExpression` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2489:54` | `isJsxOpeningLikeElement` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2495:44` | `isJsxCallLike` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2504:52` | `isCaseOrDefaultClause` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2534:41` | `isJSDocTag` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2538:44` | `isSetAccessor` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2542:44` | `isGetAccessor` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2552:44` | `hasJSDocNodes` | delegation or composition of predicate calls | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2564:38` | `hasType` | property presence or structural reads | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2573:45` | `hasInitializer` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2578:59` | `hasOnlyExpressionInitializer` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2592:53` | `isObjectLiteralElement` | kind partitions with control flow or extra conditions | unproved | Refused | NoResolvedCall | 0 |
| `utilitiesPublic.ts:2597:50` | `isTypeReferenceType` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `utilitiesPublic.ts:2625:66` | `isStringLiteralLike` | kind partitions with control flow or extra conditions | unproved | Refused | PendingView | 0 |
| `utilitiesPublic.ts:2629:46` | `isJSDocLinkLike` | direct kind comparison | proved | NotYet | PendingView | 0 |
| `watch.ts:334:96` | `isBuilderProgram` | property presence or structural reads | unproved | Refused | PendingView | 0 |
| `watchPublic.ts:739:77` | `isFileMissingOnHost` | other value, generic, assertion or erased claims | proved | Proven | Proven | 0 |
| `watchPublic.ts:743:83` | `isFilePresenceUnknownOnHost` | property presence or structural reads | unproved | Refused | PendingView | 0 |
