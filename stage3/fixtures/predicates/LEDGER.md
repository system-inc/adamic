# Predicate locations

Counts are census nodes, including overloads and callback contracts. See README.md for proof requirements and classification priority.

## Debug.assert-style asserts (16)

| Source location | Name | Implementation line | Features |
| --- | --- | ---: | --- |
| src/compiler/debug.ts:213:142 | assert | 213 | typeof |
| src/compiler/debug.ts:248:99 | assertIsDefined | 248 |  |
| src/compiler/debug.ts:260:127 | assertEachIsDefined | 262 | predicate call |
| src/compiler/debug.ts:261:114 | assertEachIsDefined | 262 | predicate call |
| src/compiler/debug.ts:278:165 | assertEachNode | 283 | predicate call |
| src/compiler/debug.ts:279:165 | assertEachNode | 283 | predicate call |
| src/compiler/debug.ts:280:177 | assertEachNode | 283 | predicate call |
| src/compiler/debug.ts:281:177 | assertEachNode | 283 | predicate call |
| src/compiler/debug.ts:294:161 | assertNode | 296 | predicate call |
| src/compiler/debug.ts:307:167 | assertNotNode | 309 | predicate call |
| src/compiler/debug.ts:320:157 | assertOptionalNode | 323 | predicate call |
| src/compiler/debug.ts:321:169 | assertOptionalNode | 323 | predicate call |
| src/compiler/debug.ts:334:146 | assertOptionalToken | 337 | kind ===, predicate call |
| src/compiler/debug.ts:335:158 | assertOptionalToken | 337 | kind ===, predicate call |
| src/compiler/debug.ts:348:112 | assertMissingNode | 349 | predicate call |
| src/compiler/debug.ts:365:46 | type | 366 |  |

## flags mask (24)

| Source location | Name | Implementation line | Features |
| --- | --- | ---: | --- |
| src/compiler/checker.ts:4290:132 | isNonLocalAlias | 4290 | flags mask |
| src/compiler/checker.ts:13387:43 | isValidBaseType | 13387 | flags mask, predicate call |
| src/compiler/checker.ts:14997:47 | isGenericMappedType | 14997 | flags mask |
| src/compiler/checker.ts:17304:48 | isJSDocTypeReference | 17304 | flags mask, kind === |
| src/compiler/checker.ts:25116:54 | isNonDeferredTypeReference | 25116 | flags mask |
| src/compiler/checker.ts:25515:39 | isArrayType | 25515 | flags mask |
| src/compiler/checker.ts:25707:39 | isTupleType | 25707 | flags mask |
| src/compiler/checker.ts:25711:46 | isGenericTupleType | 25711 | flags mask, predicate call |
| src/compiler/checker.ts:35331:56 | <callback or function type> | 35331 | flags mask |
| src/compiler/checker.ts:43611:54 | isAwaitedTypeInstantiation | 43611 | flags mask |
| src/compiler/checker.ts:51437:49 | <callback or function type> | 51437 | flags mask |
| src/compiler/debug.ts:958:51 | isFlowSwitchClause | 958 | flags mask |
| src/compiler/debug.ts:962:47 | hasAntecedents | 962 | flags mask |
| src/compiler/debug.ts:966:46 | hasAntecedent | 966 | flags mask |
| src/compiler/debug.ts:970:40 | hasNode | 970 | flags mask |
| src/compiler/factory/utilities.ts:631:104 | isOuterExpression | 631 | flags mask, kind switch, predicate call |
| src/compiler/transformers/es2017.ts:586:80 | isVariableDeclarationListWithCollidingName | 586 | flags mask, predicate call |
| src/compiler/utilities.ts:655:52 | isTransientSymbol | 655 | flags mask |
| src/compiler/utilities.ts:11162:57 | isTypeUsableAsPropertyName | 11162 | flags mask |
| src/compiler/utilitiesPublic.ts:1368:52 | isPropertyAccessChain | 1368 | flags mask, predicate call |
| src/compiler/utilitiesPublic.ts:1372:51 | isElementAccessChain | 1372 | flags mask, predicate call |
| src/compiler/utilitiesPublic.ts:1376:42 | isCallChain | 1376 | flags mask, predicate call |
| src/compiler/utilitiesPublic.ts:1380:46 | isOptionalChain | 1380 | flags mask, kind === |
| src/compiler/utilitiesPublic.ts:1437:45 | isNonNullChain | 1437 | flags mask, predicate call |

## kind comparison (345)

| Source location | Name | Implementation line | Features |
| --- | --- | ---: | --- |
| src/compiler/checker.ts:7873:81 | <callback or function type> | 7873 | kind === |
| src/compiler/checker.ts:9408:75 | isIdentifierAndNotUndefined | 9408 | kind === |
| src/compiler/checker.ts:32636:54 | <callback or function type> | 32636 | kind === |
| src/compiler/checker.ts:35656:61 | isSpreadArgument | 35656 | kind === |
| src/compiler/checker.ts:38480:63 | isValidDeclarationForTupleLabel | 38480 | kind ===, predicate call |
| src/compiler/checker.ts:44756:49 | isImportedDeclaration | 44756 | kind === |
| src/compiler/checker.ts:47505:62 | <callback or function type> | 47505 | kind === |
| src/compiler/factory/nodeTests.ts:235:47 | isNumericLiteral | 235 | kind === |
| src/compiler/factory/nodeTests.ts:239:46 | isBigIntLiteral | 239 | kind === |
| src/compiler/factory/nodeTests.ts:243:46 | isStringLiteral | 243 | kind === |
| src/compiler/factory/nodeTests.ts:247:40 | isJsxText | 247 | kind === |
| src/compiler/factory/nodeTests.ts:251:57 | isRegularExpressionLiteral | 251 | kind === |
| src/compiler/factory/nodeTests.ts:255:62 | isNoSubstitutionTemplateLiteral | 255 | kind === |
| src/compiler/factory/nodeTests.ts:261:45 | isTemplateHead | 261 | kind === |
| src/compiler/factory/nodeTests.ts:265:47 | isTemplateMiddle | 265 | kind === |
| src/compiler/factory/nodeTests.ts:269:45 | isTemplateTail | 269 | kind === |
| src/compiler/factory/nodeTests.ts:275:47 | isDotDotDotToken | 275 | kind === |
| src/compiler/factory/nodeTests.ts:280:43 | isCommaToken | 280 | kind === |
| src/compiler/factory/nodeTests.ts:284:42 | isPlusToken | 284 | kind === |
| src/compiler/factory/nodeTests.ts:288:43 | isMinusToken | 288 | kind === |
| src/compiler/factory/nodeTests.ts:292:46 | isAsteriskToken | 292 | kind === |
| src/compiler/factory/nodeTests.ts:296:49 | isExclamationToken | 296 | kind === |
| src/compiler/factory/nodeTests.ts:300:46 | isQuestionToken | 300 | kind === |
| src/compiler/factory/nodeTests.ts:304:43 | isColonToken | 304 | kind === |
| src/compiler/factory/nodeTests.ts:308:49 | isQuestionDotToken | 308 | kind === |
| src/compiler/factory/nodeTests.ts:312:55 | isEqualsGreaterThanToken | 312 | kind === |
| src/compiler/factory/nodeTests.ts:318:43 | isIdentifier | 318 | kind === |
| src/compiler/factory/nodeTests.ts:322:50 | isPrivateIdentifier | 322 | kind === |
| src/compiler/factory/nodeTests.ts:329:47 | isExportModifier | 329 | kind === |
| src/compiler/factory/nodeTests.ts:334:48 | isDefaultModifier | 334 | kind === |
| src/compiler/factory/nodeTests.ts:339:46 | isAsyncModifier | 339 | kind === |
| src/compiler/factory/nodeTests.ts:343:47 | isAssertsKeyword | 343 | kind === |
| src/compiler/factory/nodeTests.ts:347:45 | isAwaitKeyword | 347 | kind === |
| src/compiler/factory/nodeTests.ts:352:48 | isReadonlyKeyword | 352 | kind === |
| src/compiler/factory/nodeTests.ts:357:47 | isStaticModifier | 357 | kind === |
| src/compiler/factory/nodeTests.ts:362:49 | isAbstractModifier | 362 | kind === |
| src/compiler/factory/nodeTests.ts:367:49 | isOverrideModifier | 367 | kind === |
| src/compiler/factory/nodeTests.ts:372:49 | isAccessorModifier | 372 | kind === |
| src/compiler/factory/nodeTests.ts:377:45 | isSuperKeyword | 377 | kind === |
| src/compiler/factory/nodeTests.ts:382:46 | isImportKeyword | 382 | kind === |
| src/compiler/factory/nodeTests.ts:387:44 | isCaseKeyword | 387 | kind === |
| src/compiler/factory/nodeTests.ts:393:46 | isQualifiedName | 393 | kind === |
| src/compiler/factory/nodeTests.ts:397:53 | isComputedPropertyName | 397 | kind === |
| src/compiler/factory/nodeTests.ts:403:57 | isTypeParameterDeclaration | 403 | kind === |
| src/compiler/factory/nodeTests.ts:408:42 | isParameter | 408 | kind === |
| src/compiler/factory/nodeTests.ts:412:42 | isDecorator | 412 | kind === |
| src/compiler/factory/nodeTests.ts:418:50 | isPropertySignature | 418 | kind === |
| src/compiler/factory/nodeTests.ts:422:52 | isPropertyDeclaration | 422 | kind === |
| src/compiler/factory/nodeTests.ts:426:48 | isMethodSignature | 426 | kind === |
| src/compiler/factory/nodeTests.ts:430:50 | isMethodDeclaration | 430 | kind === |
| src/compiler/factory/nodeTests.ts:434:60 | isClassStaticBlockDeclaration | 434 | kind === |
| src/compiler/factory/nodeTests.ts:438:55 | isConstructorDeclaration | 438 | kind === |
| src/compiler/factory/nodeTests.ts:442:55 | isGetAccessorDeclaration | 442 | kind === |
| src/compiler/factory/nodeTests.ts:446:55 | isSetAccessorDeclaration | 446 | kind === |
| src/compiler/factory/nodeTests.ts:450:57 | isCallSignatureDeclaration | 450 | kind === |
| src/compiler/factory/nodeTests.ts:454:62 | isConstructSignatureDeclaration | 454 | kind === |
| src/compiler/factory/nodeTests.ts:458:58 | isIndexSignatureDeclaration | 458 | kind === |
| src/compiler/factory/nodeTests.ts:464:50 | isTypePredicateNode | 464 | kind === |
| src/compiler/factory/nodeTests.ts:468:50 | isTypeReferenceNode | 468 | kind === |
| src/compiler/factory/nodeTests.ts:472:49 | isFunctionTypeNode | 472 | kind === |
| src/compiler/factory/nodeTests.ts:476:52 | isConstructorTypeNode | 476 | kind === |
| src/compiler/factory/nodeTests.ts:480:46 | isTypeQueryNode | 480 | kind === |
| src/compiler/factory/nodeTests.ts:484:48 | isTypeLiteralNode | 484 | kind === |
| src/compiler/factory/nodeTests.ts:488:46 | isArrayTypeNode | 488 | kind === |
| src/compiler/factory/nodeTests.ts:492:46 | isTupleTypeNode | 492 | kind === |
| src/compiler/factory/nodeTests.ts:496:49 | isNamedTupleMember | 496 | kind === |
| src/compiler/factory/nodeTests.ts:500:49 | isOptionalTypeNode | 500 | kind === |
| src/compiler/factory/nodeTests.ts:504:45 | isRestTypeNode | 504 | kind === |
| src/compiler/factory/nodeTests.ts:508:46 | isUnionTypeNode | 508 | kind === |
| src/compiler/factory/nodeTests.ts:512:53 | isIntersectionTypeNode | 512 | kind === |
| src/compiler/factory/nodeTests.ts:516:52 | isConditionalTypeNode | 516 | kind === |
| src/compiler/factory/nodeTests.ts:520:46 | isInferTypeNode | 520 | kind === |
| src/compiler/factory/nodeTests.ts:524:54 | isParenthesizedTypeNode | 524 | kind === |
| src/compiler/factory/nodeTests.ts:528:45 | isThisTypeNode | 528 | kind === |
| src/compiler/factory/nodeTests.ts:532:49 | isTypeOperatorNode | 532 | kind === |
| src/compiler/factory/nodeTests.ts:536:54 | isIndexedAccessTypeNode | 536 | kind === |
| src/compiler/factory/nodeTests.ts:540:47 | isMappedTypeNode | 540 | kind === |
| src/compiler/factory/nodeTests.ts:544:48 | isLiteralTypeNode | 544 | kind === |
| src/compiler/factory/nodeTests.ts:548:47 | isImportTypeNode | 548 | kind === |
| src/compiler/factory/nodeTests.ts:552:56 | isTemplateLiteralTypeSpan | 552 | kind === |
| src/compiler/factory/nodeTests.ts:556:56 | isTemplateLiteralTypeNode | 556 | kind === |
| src/compiler/factory/nodeTests.ts:562:53 | isObjectBindingPattern | 562 | kind === |
| src/compiler/factory/nodeTests.ts:566:52 | isArrayBindingPattern | 566 | kind === |
| src/compiler/factory/nodeTests.ts:570:47 | isBindingElement | 570 | kind === |
| src/compiler/factory/nodeTests.ts:576:55 | isArrayLiteralExpression | 576 | kind === |
| src/compiler/factory/nodeTests.ts:580:56 | isObjectLiteralExpression | 580 | kind === |
| src/compiler/factory/nodeTests.ts:584:57 | isPropertyAccessExpression | 584 | kind === |
| src/compiler/factory/nodeTests.ts:588:56 | isElementAccessExpression | 588 | kind === |
| src/compiler/factory/nodeTests.ts:592:47 | isCallExpression | 592 | kind === |
| src/compiler/factory/nodeTests.ts:596:46 | isNewExpression | 596 | kind === |
| src/compiler/factory/nodeTests.ts:600:57 | isTaggedTemplateExpression | 600 | kind === |
| src/compiler/factory/nodeTests.ts:604:56 | isTypeAssertionExpression | 604 | kind === |
| src/compiler/factory/nodeTests.ts:608:56 | isParenthesizedExpression | 608 | kind === |
| src/compiler/factory/nodeTests.ts:612:51 | isFunctionExpression | 612 | kind === |
| src/compiler/factory/nodeTests.ts:616:46 | isArrowFunction | 616 | kind === |
| src/compiler/factory/nodeTests.ts:620:49 | isDeleteExpression | 620 | kind === |
| src/compiler/factory/nodeTests.ts:624:49 | isTypeOfExpression | 624 | kind === |
| src/compiler/factory/nodeTests.ts:628:47 | isVoidExpression | 628 | kind === |
| src/compiler/factory/nodeTests.ts:632:48 | isAwaitExpression | 632 | kind === |
| src/compiler/factory/nodeTests.ts:636:54 | isPrefixUnaryExpression | 636 | kind === |
| src/compiler/factory/nodeTests.ts:640:55 | isPostfixUnaryExpression | 640 | kind === |
| src/compiler/factory/nodeTests.ts:644:49 | isBinaryExpression | 644 | kind === |
| src/compiler/factory/nodeTests.ts:648:54 | isConditionalExpression | 648 | kind === |
| src/compiler/factory/nodeTests.ts:652:51 | isTemplateExpression | 652 | kind === |
| src/compiler/factory/nodeTests.ts:656:48 | isYieldExpression | 656 | kind === |
| src/compiler/factory/nodeTests.ts:660:46 | isSpreadElement | 660 | kind === |
| src/compiler/factory/nodeTests.ts:664:48 | isClassExpression | 664 | kind === |
| src/compiler/factory/nodeTests.ts:668:50 | isOmittedExpression | 668 | kind === |
| src/compiler/factory/nodeTests.ts:672:60 | isExpressionWithTypeArguments | 672 | kind === |
| src/compiler/factory/nodeTests.ts:676:45 | isAsExpression | 676 | kind === |
| src/compiler/factory/nodeTests.ts:680:52 | isSatisfiesExpression | 680 | kind === |
| src/compiler/factory/nodeTests.ts:684:50 | isNonNullExpression | 684 | kind === |
| src/compiler/factory/nodeTests.ts:688:45 | isMetaProperty | 688 | kind === |
| src/compiler/factory/nodeTests.ts:692:52 | isSyntheticExpression | 692 | kind === |
| src/compiler/factory/nodeTests.ts:696:59 | isPartiallyEmittedExpression | 696 | kind === |
| src/compiler/factory/nodeTests.ts:700:52 | isCommaListExpression | 700 | kind === |
| src/compiler/factory/nodeTests.ts:706:45 | isTemplateSpan | 706 | kind === |
| src/compiler/factory/nodeTests.ts:710:54 | isSemicolonClassElement | 710 | kind === |
| src/compiler/factory/nodeTests.ts:716:38 | isBlock | 716 | kind === |
| src/compiler/factory/nodeTests.ts:720:50 | isVariableStatement | 720 | kind === |
| src/compiler/factory/nodeTests.ts:724:47 | isEmptyStatement | 724 | kind === |
| src/compiler/factory/nodeTests.ts:728:52 | isExpressionStatement | 728 | kind === |
| src/compiler/factory/nodeTests.ts:732:44 | isIfStatement | 732 | kind === |
| src/compiler/factory/nodeTests.ts:736:44 | isDoStatement | 736 | kind === |
| src/compiler/factory/nodeTests.ts:740:47 | isWhileStatement | 740 | kind === |
| src/compiler/factory/nodeTests.ts:744:45 | isForStatement | 744 | kind === |
| src/compiler/factory/nodeTests.ts:748:47 | isForInStatement | 748 | kind === |
| src/compiler/factory/nodeTests.ts:752:47 | isForOfStatement | 752 | kind === |
| src/compiler/factory/nodeTests.ts:756:50 | isContinueStatement | 756 | kind === |
| src/compiler/factory/nodeTests.ts:760:47 | isBreakStatement | 760 | kind === |
| src/compiler/factory/nodeTests.ts:764:48 | isReturnStatement | 764 | kind === |
| src/compiler/factory/nodeTests.ts:768:46 | isWithStatement | 768 | kind === |
| src/compiler/factory/nodeTests.ts:772:48 | isSwitchStatement | 772 | kind === |
| src/compiler/factory/nodeTests.ts:776:49 | isLabeledStatement | 776 | kind === |
| src/compiler/factory/nodeTests.ts:780:47 | isThrowStatement | 780 | kind === |
| src/compiler/factory/nodeTests.ts:784:45 | isTryStatement | 784 | kind === |
| src/compiler/factory/nodeTests.ts:788:50 | isDebuggerStatement | 788 | kind === |
| src/compiler/factory/nodeTests.ts:792:52 | isVariableDeclaration | 792 | kind === |
| src/compiler/factory/nodeTests.ts:796:56 | isVariableDeclarationList | 796 | kind === |
| src/compiler/factory/nodeTests.ts:800:52 | isFunctionDeclaration | 800 | kind === |
| src/compiler/factory/nodeTests.ts:804:49 | isClassDeclaration | 804 | kind === |
| src/compiler/factory/nodeTests.ts:808:53 | isInterfaceDeclaration | 808 | kind === |
| src/compiler/factory/nodeTests.ts:812:53 | isTypeAliasDeclaration | 812 | kind === |
| src/compiler/factory/nodeTests.ts:816:48 | isEnumDeclaration | 816 | kind === |
| src/compiler/factory/nodeTests.ts:820:50 | isModuleDeclaration | 820 | kind === |
| src/compiler/factory/nodeTests.ts:824:44 | isModuleBlock | 824 | kind === |
| src/compiler/factory/nodeTests.ts:828:42 | isCaseBlock | 828 | kind === |
| src/compiler/factory/nodeTests.ts:832:59 | isNamespaceExportDeclaration | 832 | kind === |
| src/compiler/factory/nodeTests.ts:836:56 | isImportEqualsDeclaration | 836 | kind === |
| src/compiler/factory/nodeTests.ts:840:50 | isImportDeclaration | 840 | kind === |
| src/compiler/factory/nodeTests.ts:844:45 | isImportClause | 844 | kind === |
| src/compiler/factory/nodeTests.ts:848:61 | isImportTypeAssertionContainer | 848 | kind === |
| src/compiler/factory/nodeTests.ts:853:45 | isAssertClause | 853 | kind === |
| src/compiler/factory/nodeTests.ts:858:44 | isAssertEntry | 858 | kind === |
| src/compiler/factory/nodeTests.ts:862:49 | isImportAttributes | 862 | kind === |
| src/compiler/factory/nodeTests.ts:866:48 | isImportAttribute | 866 | kind === |
| src/compiler/factory/nodeTests.ts:870:48 | isNamespaceImport | 870 | kind === |
| src/compiler/factory/nodeTests.ts:874:48 | isNamespaceExport | 874 | kind === |
| src/compiler/factory/nodeTests.ts:878:45 | isNamedImports | 878 | kind === |
| src/compiler/factory/nodeTests.ts:882:48 | isImportSpecifier | 882 | kind === |
| src/compiler/factory/nodeTests.ts:886:49 | isExportAssignment | 886 | kind === |
| src/compiler/factory/nodeTests.ts:890:50 | isExportDeclaration | 890 | kind === |
| src/compiler/factory/nodeTests.ts:894:45 | isNamedExports | 894 | kind === |
| src/compiler/factory/nodeTests.ts:898:48 | isExportSpecifier | 898 | kind === |
| src/compiler/factory/nodeTests.ts:902:49 | isModuleExportName | 902 | kind === |
| src/compiler/factory/nodeTests.ts:906:51 | isMissingDeclaration | 906 | kind === |
| src/compiler/factory/nodeTests.ts:910:52 | isNotEmittedStatement | 910 | kind === |
| src/compiler/factory/nodeTests.ts:915:51 | isSyntheticReference | 915 | kind === |
| src/compiler/factory/nodeTests.ts:921:56 | isExternalModuleReference | 921 | kind === |
| src/compiler/factory/nodeTests.ts:927:43 | isJsxElement | 927 | kind === |
| src/compiler/factory/nodeTests.ts:931:54 | isJsxSelfClosingElement | 931 | kind === |
| src/compiler/factory/nodeTests.ts:935:50 | isJsxOpeningElement | 935 | kind === |
| src/compiler/factory/nodeTests.ts:939:50 | isJsxClosingElement | 939 | kind === |
| src/compiler/factory/nodeTests.ts:943:44 | isJsxFragment | 943 | kind === |
| src/compiler/factory/nodeTests.ts:947:51 | isJsxOpeningFragment | 947 | kind === |
| src/compiler/factory/nodeTests.ts:951:51 | isJsxClosingFragment | 951 | kind === |
| src/compiler/factory/nodeTests.ts:955:45 | isJsxAttribute | 955 | kind === |
| src/compiler/factory/nodeTests.ts:959:46 | isJsxAttributes | 959 | kind === |
| src/compiler/factory/nodeTests.ts:963:51 | isJsxSpreadAttribute | 963 | kind === |
| src/compiler/factory/nodeTests.ts:967:46 | isJsxExpression | 967 | kind === |
| src/compiler/factory/nodeTests.ts:971:50 | isJsxNamespacedName | 971 | kind === |
| src/compiler/factory/nodeTests.ts:977:43 | isCaseClause | 977 | kind === |
| src/compiler/factory/nodeTests.ts:981:46 | isDefaultClause | 981 | kind === |
| src/compiler/factory/nodeTests.ts:985:47 | isHeritageClause | 985 | kind === |
| src/compiler/factory/nodeTests.ts:989:44 | isCatchClause | 989 | kind === |
| src/compiler/factory/nodeTests.ts:995:51 | isPropertyAssignment | 995 | kind === |
| src/compiler/factory/nodeTests.ts:999:60 | isShorthandPropertyAssignment | 999 | kind === |
| src/compiler/factory/nodeTests.ts:1003:49 | isSpreadAssignment | 1003 | kind === |
| src/compiler/factory/nodeTests.ts:1009:43 | isEnumMember | 1009 | kind === |
| src/compiler/factory/nodeTests.ts:1014:43 | isSourceFile | 1014 | kind === |
| src/compiler/factory/nodeTests.ts:1018:39 | isBundle | 1018 | kind === |
| src/compiler/factory/nodeTests.ts:1026:52 | isJSDocTypeExpression | 1026 | kind === |
| src/compiler/factory/nodeTests.ts:1030:51 | isJSDocNameReference | 1030 | kind === |
| src/compiler/factory/nodeTests.ts:1034:48 | isJSDocMemberName | 1034 | kind === |
| src/compiler/factory/nodeTests.ts:1038:42 | isJSDocLink | 1038 | kind === |
| src/compiler/factory/nodeTests.ts:1042:46 | isJSDocLinkCode | 1042 | kind === |
| src/compiler/factory/nodeTests.ts:1046:47 | isJSDocLinkPlain | 1046 | kind === |
| src/compiler/factory/nodeTests.ts:1050:45 | isJSDocAllType | 1050 | kind === |
| src/compiler/factory/nodeTests.ts:1054:49 | isJSDocUnknownType | 1054 | kind === |
| src/compiler/factory/nodeTests.ts:1058:50 | isJSDocNullableType | 1058 | kind === |
| src/compiler/factory/nodeTests.ts:1062:53 | isJSDocNonNullableType | 1062 | kind === |
| src/compiler/factory/nodeTests.ts:1066:50 | isJSDocOptionalType | 1066 | kind === |
| src/compiler/factory/nodeTests.ts:1070:50 | isJSDocFunctionType | 1070 | kind === |
| src/compiler/factory/nodeTests.ts:1074:50 | isJSDocVariadicType | 1074 | kind === |
| src/compiler/factory/nodeTests.ts:1078:50 | isJSDocNamepathType | 1078 | kind === |
| src/compiler/factory/nodeTests.ts:1082:38 | isJSDoc | 1082 | kind === |
| src/compiler/factory/nodeTests.ts:1086:49 | isJSDocTypeLiteral | 1086 | kind === |
| src/compiler/factory/nodeTests.ts:1090:47 | isJSDocSignature | 1090 | kind === |
| src/compiler/factory/nodeTests.ts:1096:49 | isJSDocAugmentsTag | 1096 | kind === |
| src/compiler/factory/nodeTests.ts:1100:47 | isJSDocAuthorTag | 1100 | kind === |
| src/compiler/factory/nodeTests.ts:1104:46 | isJSDocClassTag | 1104 | kind === |
| src/compiler/factory/nodeTests.ts:1108:49 | isJSDocCallbackTag | 1108 | kind === |
| src/compiler/factory/nodeTests.ts:1112:47 | isJSDocPublicTag | 1112 | kind === |
| src/compiler/factory/nodeTests.ts:1116:48 | isJSDocPrivateTag | 1116 | kind === |
| src/compiler/factory/nodeTests.ts:1120:50 | isJSDocProtectedTag | 1120 | kind === |
| src/compiler/factory/nodeTests.ts:1124:49 | isJSDocReadonlyTag | 1124 | kind === |
| src/compiler/factory/nodeTests.ts:1128:49 | isJSDocOverrideTag | 1128 | kind === |
| src/compiler/factory/nodeTests.ts:1132:49 | isJSDocOverloadTag | 1132 | kind === |
| src/compiler/factory/nodeTests.ts:1136:51 | isJSDocDeprecatedTag | 1136 | kind === |
| src/compiler/factory/nodeTests.ts:1140:44 | isJSDocSeeTag | 1140 | kind === |
| src/compiler/factory/nodeTests.ts:1144:45 | isJSDocEnumTag | 1144 | kind === |
| src/compiler/factory/nodeTests.ts:1148:50 | isJSDocParameterTag | 1148 | kind === |
| src/compiler/factory/nodeTests.ts:1152:47 | isJSDocReturnTag | 1152 | kind === |
| src/compiler/factory/nodeTests.ts:1156:45 | isJSDocThisTag | 1156 | kind === |
| src/compiler/factory/nodeTests.ts:1160:45 | isJSDocTypeTag | 1160 | kind === |
| src/compiler/factory/nodeTests.ts:1164:49 | isJSDocTemplateTag | 1164 | kind === |
| src/compiler/factory/nodeTests.ts:1168:48 | isJSDocTypedefTag | 1168 | kind === |
| src/compiler/factory/nodeTests.ts:1172:48 | isJSDocUnknownTag | 1172 | kind === |
| src/compiler/factory/nodeTests.ts:1176:49 | isJSDocPropertyTag | 1176 | kind === |
| src/compiler/factory/nodeTests.ts:1180:51 | isJSDocImplementsTag | 1180 | kind === |
| src/compiler/factory/nodeTests.ts:1184:50 | isJSDocSatisfiesTag | 1184 | kind === |
| src/compiler/factory/nodeTests.ts:1188:47 | isJSDocThrowsTag | 1188 | kind === |
| src/compiler/factory/nodeTests.ts:1192:47 | isJSDocImportTag | 1192 | kind === |
| src/compiler/factory/nodeTests.ts:1199:40 | isSyntaxList | 1199 | kind === |
| src/compiler/factory/utilities.ts:607:54 | isCommaExpression | 607 | kind === |
| src/compiler/factory/utilities.ts:1065:48 | isStringOrNumericLiteral | 1065 | kind === |
| src/compiler/factory/utilities.ts:1105:49 | canHaveIllegalType | 1105 | kind === |
| src/compiler/factory/utilities.ts:1112:59 | canHaveIllegalTypeParameters | 1112 | kind === |
| src/compiler/factory/utilities.ts:1120:55 | canHaveIllegalDecorators | 1120 | kind === |
| src/compiler/factory/utilities.ts:1142:54 | canHaveIllegalModifiers | 1142 | kind === |
| src/compiler/factory/utilities.ts:1171:54 | isExponentiationOperator | 1171 | kind === |
| src/compiler/factory/utilities.ts:1175:54 | isMultiplicativeOperator | 1175 | kind === |
| src/compiler/factory/utilities.ts:1186:48 | isAdditiveOperator | 1186 | kind === |
| src/compiler/factory/utilities.ts:1196:45 | isShiftOperator | 1196 | kind === |
| src/compiler/factory/utilities.ts:1208:50 | isRelationalOperator | 1208 | kind === |
| src/compiler/factory/utilities.ts:1222:48 | isEqualityOperator | 1222 | kind === |
| src/compiler/factory/utilities.ts:1234:47 | isBitwiseOperator | 1234 | kind === |
| src/compiler/factory/utilities.ts:1246:47 | isLogicalOperator | 1246 | kind === |
| src/compiler/factory/utilities.ts:1256:58 | isAssignmentOperatorOrHigher | 1256 | kind ===, predicate call |
| src/compiler/factory/utilities.ts:1262:46 | isBinaryOperator | 1262 | kind ===, predicate call |
| src/compiler/factory/utilities.ts:1489:58 | isExportOrDefaultKeywordKind | 1489 | kind === |
| src/compiler/factory/utilitiesPublic.ts:14:47 | canHaveModifiers | 14 | kind === |
| src/compiler/factory/utilitiesPublic.ts:43:48 | canHaveDecorators | 43 | kind === |
| src/compiler/transformers/classFields.ts:3349:67 | isPrivateIdentifierInExpression | 3349 | kind ===, predicate call |
| src/compiler/transformers/classThis.ts:72:57 | isClassThisAssignmentBlock | 72 | kind ===, predicate call |
| src/compiler/transformers/es2015.ts:1375:50 | isTransformedSuperCall | 1375 | kind ===, predicate call |
| src/compiler/transformers/es2015.ts:1386:62 | isTransformedSuperCallWithFallback | 1386 | kind ===, predicate call |
| src/compiler/transformers/es2015.ts:1393:47 | isImplicitSuperCall | 1393 | kind ===, predicate call |
| src/compiler/transformers/es2015.ts:1403:59 | isImplicitSuperCallWithFallback | 1403 | kind ===, predicate call |
| src/compiler/transformers/es2017.ts:1018:44 | isSuperContainer | 1018 | kind === |
| src/compiler/transformers/generators.ts:2420:56 | supportsUnlabeledBreak | 2420 | kind === |
| src/compiler/transformers/generators.ts:2430:64 | supportsLabeledBreakOrContinue | 2430 | kind === |
| src/compiler/transformers/generators.ts:2439:59 | supportsUnlabeledContinue | 2439 | kind === |
| src/compiler/transformers/ts.ts:1171:56 | shouldAddTypeMetadata | 1171 | kind === |
| src/compiler/transformers/ts.ts:1186:62 | shouldAddReturnTypeMetadata | 1186 | kind === |
| src/compiler/transformers/utilities.ts:630:62 | isInitializedProperty | 630 | kind === |
| src/compiler/utilities.ts:2078:46 | isAmbientModule | 2078 | kind ===, predicate call |
| src/compiler/utilities.ts:2083:60 | isModuleWithStringLiteralName | 2083 | kind ===, predicate call |
| src/compiler/utilities.ts:2701:39 | isSuperCall | 2701 | kind === |
| src/compiler/utilities.ts:2706:40 | isImportCall | 2706 | kind ===, predicate call |
| src/compiler/utilities.ts:2729:50 | isPrologueDirective | 2729 | kind === |
| src/compiler/utilities.ts:3073:52 | isObjectLiteralMethod | 3073 | kind === |
| src/compiler/utilities.ts:3078:79 | isObjectLiteralOrClassExpressionMethodOrAccessor | 3078 | kind === |
| src/compiler/utilities.ts:3085:70 | isIdentifierTypePredicate | 3085 | kind === |
| src/compiler/utilities.ts:3090:64 | isThisTypePredicate | 3090 | kind === |
| src/compiler/utilities.ts:3402:46 | isSuperProperty | 3402 | kind === |
| src/compiler/utilities.ts:3773:70 | isExternalModuleImportEqualsDeclaration | 3773 | kind === |
| src/compiler/utilities.ts:3789:70 | isInternalModuleImportEqualsDeclaration | 3789 | kind === |
| src/compiler/utilities.ts:3794:55 | isFullSourceFile | 3794 | kind === |
| src/compiler/utilities.ts:4148:93 | isBindableStaticAccessExpression | 4148 | kind ===, predicate call |
| src/compiler/utilities.ts:4158:100 | isBindableStaticElementAccessExpression | 4158 | kind ===, predicate call |
| src/compiler/utilities.ts:4299:105 | isSpecialPropertyDeclaration | 4299 | kind ===, predicate call |
| src/compiler/utilities.ts:4482:47 | isJSDocTypeAlias | 4482 | kind === |
| src/compiler/utilities.ts:6923:49 | isNonTypeAliasTemplate | 6923 | kind ===, predicate call |
| src/compiler/utilities.ts:7471:86 | isAssignmentExpression | 7475 | kind ===, predicate call |
| src/compiler/utilities.ts:7473:88 | isAssignmentExpression | 7475 | kind ===, predicate call |
| src/compiler/utilities.ts:7475:90 | isAssignmentExpression | 7475 | kind ===, predicate call |
| src/compiler/utilities.ts:7484:56 | isDestructuringAssignment | 7484 | kind ===, predicate call |
| src/compiler/utilities.ts:7500:53 | isEntityNameExpression | 7500 | kind ===, predicate call |
| src/compiler/utilities.ts:7587:53 | isInstanceOfExpression | 7587 | kind ===, predicate call |
| src/compiler/utilities.ts:8330:51 | isTypeNodeKind | 8330 | kind === |
| src/compiler/utilities.ts:8355:49 | isAccessExpression | 8355 | kind === |
| src/compiler/utilities.ts:8369:54 | isNamedImportsOrExports | 8369 | kind === |
| src/compiler/utilities.ts:10910:66 | isFunctionExpressionOrArrowFunction | 10910 | kind === |
| src/compiler/utilities.ts:11104:46 | isNonNullAccess | 11104 | kind ===, predicate call |
| src/compiler/utilities.ts:11137:49 | isJsxAttributeName | 11137 | kind === |
| src/compiler/utilitiesPublic.ts:615:75 | isParameterPropertyDeclaration | 615 | kind ===, predicate call |
| src/compiler/utilitiesPublic.ts:1359:43 | isMemberName | 1359 | kind === |
| src/compiler/utilitiesPublic.ts:1364:60 | isGetOrSetAccessorDeclaration | 1364 | kind === |
| src/compiler/utilitiesPublic.ts:1441:57 | isBreakOrContinueStatement | 1441 | kind === |
| src/compiler/utilitiesPublic.ts:1445:52 | isNamedExportBindings | 1445 | kind === |
| src/compiler/utilitiesPublic.ts:1449:53 | isJSDocPropertyLikeTag | 1449 | kind === |
| src/compiler/utilitiesPublic.ts:1526:61 | isTemplateMiddleOrTemplateTail | 1526 | kind === |
| src/compiler/utilitiesPublic.ts:1570:57 | isStringTextContainingNode | 1570 | kind ===, predicate call |
| src/compiler/utilitiesPublic.ts:1651:43 | isEntityName | 1651 | kind === |
| src/compiler/utilitiesPublic.ts:1657:45 | isPropertyName | 1657 | kind === |
| src/compiler/utilitiesPublic.ts:1666:44 | isBindingName | 1666 | kind === |
| src/compiler/utilitiesPublic.ts:1690:47 | isBooleanLiteral | 1690 | kind === |
| src/compiler/utilitiesPublic.ts:1732:45 | isClassElement | 1732 | kind === |
| src/compiler/utilitiesPublic.ts:1744:42 | isClassLike | 1744 | kind === |
| src/compiler/utilitiesPublic.ts:1748:41 | isAccessor | 1748 | kind === |
| src/compiler/utilitiesPublic.ts:1782:44 | isTypeElement | 1782 | kind === |
| src/compiler/utilitiesPublic.ts:1798:57 | isObjectLiteralElementLike | 1798 | kind === |
| src/compiler/utilitiesPublic.ts:1832:59 | isBindingPattern | 1832 | kind === |
| src/compiler/utilitiesPublic.ts:1843:50 | isAssignmentPattern | 1843 | kind === |
| src/compiler/utilitiesPublic.ts:1849:52 | isArrayBindingElement | 1849 | kind === |
| src/compiler/utilitiesPublic.ts:1948:78 | isPropertyAccessOrQualifiedNameOrImportTypeNode | 1948 | kind === |
| src/compiler/utilitiesPublic.ts:1957:62 | isPropertyAccessOrQualifiedName | 1957 | kind === |
| src/compiler/utilitiesPublic.ts:1968:51 | isCallLikeExpression | 1968 | kind ===, kind switch |
| src/compiler/utilitiesPublic.ts:1985:52 | isCallOrNewExpression | 1985 | kind === |
| src/compiler/utilitiesPublic.ts:1989:48 | isTemplateLiteral | 1989 | kind === |
| src/compiler/utilitiesPublic.ts:2108:52 | isAssertionExpression | 2108 | kind === |
| src/compiler/utilitiesPublic.ts:2154:51 | isForInOrOfStatement | 2154 | kind === |
| src/compiler/utilitiesPublic.ts:2175:43 | isModuleBody | 2175 | kind === |
| src/compiler/utilitiesPublic.ts:2183:46 | isNamespaceBody | 2183 | kind === |
| src/compiler/utilitiesPublic.ts:2190:51 | isJSDocNamespaceBody | 2190 | kind === |
| src/compiler/utilitiesPublic.ts:2196:52 | isNamedImportBindings | 2196 | kind === |
| src/compiler/utilitiesPublic.ts:2203:56 | isModuleOrEnumDeclaration | 2203 | kind === |
| src/compiler/utilitiesPublic.ts:2398:44 | isDeclaration | 2398 | kind ===, predicate call |
| src/compiler/utilitiesPublic.ts:2426:40 | isBlockStatement | 2426 | kind === |
| src/compiler/utilitiesPublic.ts:2442:49 | isStatementOrBlock | 2442 | kind ===, predicate call |
| src/compiler/utilitiesPublic.ts:2451:48 | isModuleReference | 2451 | kind === |
| src/compiler/utilitiesPublic.ts:2460:53 | isJsxTagNameExpression | 2460 | kind === |
| src/compiler/utilitiesPublic.ts:2468:41 | isJsxChild | 2468 | kind === |
| src/compiler/utilitiesPublic.ts:2477:49 | isJsxAttributeLike | 2477 | kind === |
| src/compiler/utilitiesPublic.ts:2483:61 | isStringLiteralOrJsxExpression | 2483 | kind === |
| src/compiler/utilitiesPublic.ts:2489:54 | isJsxOpeningLikeElement | 2489 | kind === |
| src/compiler/utilitiesPublic.ts:2495:44 | isJsxCallLike | 2495 | kind === |
| src/compiler/utilitiesPublic.ts:2504:52 | isCaseOrDefaultClause | 2504 | kind === |
| src/compiler/utilitiesPublic.ts:2538:44 | isSetAccessor | 2538 | kind === |
| src/compiler/utilitiesPublic.ts:2542:44 | isGetAccessor | 2542 | kind === |
| src/compiler/utilitiesPublic.ts:2592:53 | isObjectLiteralElement | 2592 | kind ===, predicate call |
| src/compiler/utilitiesPublic.ts:2597:50 | isTypeReferenceType | 2597 | kind === |
| src/compiler/utilitiesPublic.ts:2625:66 | isStringLiteralLike | 2625 | kind === |
| src/compiler/utilitiesPublic.ts:2629:46 | isJSDocLinkLike | 2629 | kind === |

## no body (48)

| Source location | Name | Implementation line | Features |
| --- | --- | ---: | --- |
| src/compiler/core.ts:140:101 | <callback or function type> |  |  |
| src/compiler/core.ts:142:113 | <callback or function type> |  |  |
| src/compiler/core.ts:162:113 | <callback or function type> |  |  |
| src/compiler/core.ts:178:117 | <callback or function type> |  |  |
| src/compiler/core.ts:261:65 | <callback or function type> |  |  |
| src/compiler/core.ts:265:74 | <callback or function type> |  |  |
| src/compiler/core.ts:269:77 | <callback or function type> |  |  |
| src/compiler/core.ts:273:86 | <callback or function type> |  |  |
| src/compiler/core.ts:1461:102 | <callback or function type> |  |  |
| src/compiler/core.ts:1778:100 | <callback or function type> |  |  |
| src/compiler/core.ts:1783:97 | <callback or function type> |  |  |
| src/compiler/core.ts:2459:66 | <callback or function type> |  |  |
| src/compiler/core.ts:2459:91 | <callback or function type> |  |  |
| src/compiler/core.ts:2459:112 | <callback or function type> |  |  |
| src/compiler/core.ts:2461:80 | <callback or function type> |  |  |
| src/compiler/core.ts:2461:105 | <callback or function type> |  |  |
| src/compiler/core.ts:2461:130 | <callback or function type> |  |  |
| src/compiler/core.ts:2461:151 | <callback or function type> |  |  |
| src/compiler/core.ts:2557:91 | <callback or function type> |  |  |
| src/compiler/core.ts:2559:103 | <callback or function type> |  |  |
| src/compiler/core.ts:2560:103 | <callback or function type> |  |  |
| src/compiler/core.ts:2572:91 | <callback or function type> |  |  |
| src/compiler/core.ts:2574:103 | <callback or function type> |  |  |
| src/compiler/core.ts:2576:103 | <callback or function type> |  |  |
| src/compiler/debug.ts:278:105 | <callback or function type> |  |  |
| src/compiler/debug.ts:279:105 | <callback or function type> |  |  |
| src/compiler/debug.ts:280:117 | <callback or function type> |  |  |
| src/compiler/debug.ts:281:117 | <callback or function type> |  |  |
| src/compiler/debug.ts:294:101 | <callback or function type> |  |  |
| src/compiler/debug.ts:307:107 | <callback or function type> |  |  |
| src/compiler/debug.ts:320:97 | <callback or function type> |  |  |
| src/compiler/debug.ts:321:109 | <callback or function type> |  |  |
| src/compiler/types.ts:5908:37 | isLateBound |  |  |
| src/compiler/types.ts:9718:31 | <callback or function type> |  |  |
| src/compiler/types.ts:9744:31 | <callback or function type> |  |  |
| src/compiler/types.ts:10651:45 | hasLateBindableName |  |  |
| src/compiler/utilities.ts:1111:143 | <callback or function type> |  |  |
| src/compiler/utilitiesPublic.ts:763:87 | <callback or function type> |  |  |
| src/compiler/utilitiesPublic.ts:765:99 | <callback or function type> |  |  |
| src/compiler/utilitiesPublic.ts:766:100 | <callback or function type> |  |  |
| src/compiler/utilitiesPublic.ts:786:99 | <callback or function type> |  |  |
| src/compiler/utilitiesPublic.ts:826:98 | <callback or function type> |  |  |
| src/compiler/utilitiesPublic.ts:1279:89 | <callback or function type> |  |  |
| src/compiler/utilitiesPublic.ts:1284:95 | <callback or function type> |  |  |
| src/compiler/visitorPublic.ts:125:27 | <callback or function type> |  |  |
| src/compiler/visitorPublic.ts:197:27 | <callback or function type> |  |  |
| src/compiler/visitorPublic.ts:277:27 | <callback or function type> |  |  |
| src/compiler/visitorPublic.ts:317:27 | <callback or function type> |  |  |

## predicate call (136)

| Source location | Name | Implementation line | Features |
| --- | --- | ---: | --- |
| src/compiler/builder.ts:1192:54 | isNonIncrementalBuildInfo | 1192 | predicate call |
| src/compiler/checker.ts:13778:57 | isLateBindableName | 13778 | predicate call |
| src/compiler/checker.ts:13783:67 | isLateBindableIndexSignature | 13783 | predicate call |
| src/compiler/checker.ts:13809:54 | hasLateBindableName | 13809 | predicate call |
| src/compiler/checker.ts:21245:75 | isContextSensitiveFunctionOrObjectLiteralMethod | 21245 | predicate call |
| src/compiler/checker.ts:25523:46 | isArrayOrTupleType | 25523 | predicate call |
| src/compiler/checker.ts:25715:59 | isSingleElementGenericTupleType | 25715 | predicate call |
| src/compiler/checker.ts:33802:52 | isJsxIntrinsicTagName | 33802 | predicate call |
| src/compiler/checker.ts:35569:80 | callLikeExpressionMayHaveTypeArguments | 35569 | predicate call |
| src/compiler/checker.ts:37696:55 | isJSConstructor | 37696 | predicate call |
| src/compiler/checker.ts:50641:59 | <callback or function type> | 50641 | predicate call |
| src/compiler/checker.ts:50886:57 | <callback or function type> | 50886 | predicate call |
| src/compiler/commandLineParser.ts:1728:70 | isCommandLineOptionOfCustomType | 1728 | predicate call |
| src/compiler/commandLineParser.ts:3300:51 | startsWithConfigDirTemplate | 3300 | predicate call |
| src/compiler/core.ts:1750:38 | isArray | 1750 | predicate call |
| src/compiler/factory/utilities.ts:612:52 | isCommaSequence | 612 | predicate call |
| src/compiler/factory/utilities.ts:617:51 | isJSDocTypeAssertion | 617 | predicate call |
| src/compiler/factory/utilities.ts:1151:59 | isQuestionOrExclamationToken | 1151 | predicate call |
| src/compiler/factory/utilities.ts:1155:57 | isIdentifierOrThisTypeNode | 1155 | predicate call |
| src/compiler/factory/utilities.ts:1159:66 | isReadonlyKeywordOrPlusOrMinusToken | 1159 | predicate call |
| src/compiler/factory/utilities.ts:1163:59 | isQuestionOrPlusOrMinusToken | 1163 | predicate call |
| src/compiler/factory/utilities.ts:1167:43 | isModuleName | 1167 | predicate call |
| src/compiler/factory/utilities.ts:1181:62 | isMultiplicativeOperatorOrHigher | 1181 | predicate call |
| src/compiler/factory/utilities.ts:1191:56 | isAdditiveOperatorOrHigher | 1191 | predicate call |
| src/compiler/factory/utilities.ts:1203:60 | isShiftOperatorOrHigher | 1203 | predicate call |
| src/compiler/factory/utilities.ts:1217:58 | isRelationalOperatorOrHigher | 1217 | predicate call |
| src/compiler/factory/utilities.ts:1229:56 | isEqualityOperatorOrHigher | 1229 | predicate call |
| src/compiler/factory/utilities.ts:1240:55 | isBitwiseOperatorOrHigher | 1240 | predicate call |
| src/compiler/factory/utilities.ts:1251:55 | isLogicalOperatorOrHigher | 1251 | predicate call |
| src/compiler/factory/utilities.ts:1267:52 | isBinaryOperatorToken | 1267 | predicate call |
| src/compiler/factory/utilities.ts:1494:56 | isExportOrDefaultModifier | 1494 | predicate call |
| src/compiler/factory/utilities.ts:1695:64 | isSyntheticParenthesizedExpression | 1695 | predicate call |
| src/compiler/transformers/classFields.ts:853:53 | <callback or function type> | 853 | predicate call |
| src/compiler/transformers/classFields.ts:3354:51 | isStaticPropertyDeclaration | 3354 | predicate call |
| src/compiler/transformers/classFields.ts:3358:69 | isStaticPropertyDeclarationOrClassStaticBlock | 3358 | predicate call |
| src/compiler/transformers/declarations.ts:659:54 | shouldPrintWithInitializer | 659 | predicate call |
| src/compiler/transformers/declarations/diagnostics.ts:138:52 | canProduceDiagnostics | 138 | predicate call |
| src/compiler/transformers/declarations/diagnostics.ts:714:46 | <callback or function type> | 714 | predicate call |
| src/compiler/transformers/es2015.ts:1348:42 | isCapturedThis | 1348 | predicate call |
| src/compiler/transformers/es2015.ts:1353:44 | isSyntheticSuper | 1353 | predicate call |
| src/compiler/transformers/es2015.ts:1359:60 | isThisCapturingVariableStatement | 1359 | predicate call |
| src/compiler/transformers/es2015.ts:1365:62 | isThisCapturingVariableDeclaration | 1365 | predicate call |
| src/compiler/transformers/es2015.ts:1370:53 | isThisCapturingAssignment | 1370 | predicate call |
| src/compiler/transformers/es2015.ts:1410:75 | isThisCapturingTransformedSuperCallWithFallback | 1410 | predicate call |
| src/compiler/transformers/es2015.ts:1415:72 | isThisCapturingImplicitSuperCallWithFallback | 1415 | predicate call |
| src/compiler/transformers/es2015.ts:1419:54 | isTransformedSuperCallLike | 1419 | predicate call |
| src/compiler/transformers/es2015.ts:3379:80 | shouldConvertInitializerOfForStatement | 3379 | predicate call |
| src/compiler/transformers/es2015.ts:3383:78 | shouldConvertConditionOfForStatement | 3383 | predicate call |
| src/compiler/transformers/es2015.ts:3387:80 | shouldConvertIncrementorOfForStatement | 3387 | predicate call |
| src/compiler/transformers/esnext.ts:774:54 | isUsingVariableDeclarationList | 774 | predicate call |
| src/compiler/transformers/module/system.ts:1360:63 | shouldHoistForInitializer | 1360 | predicate call |
| src/compiler/transformers/namedEvaluation.ts:139:64 | isClassNamedEvaluationHelperBlock | 139 | predicate call |
| src/compiler/transformers/ts.ts:1054:49 | <callback or function type> | 1054 | predicate call |
| src/compiler/transformers/utilities.ts:593:91 | isStaticPropertyDeclarationOrClassStaticBlockDeclaration | 593 | predicate call |
| src/compiler/transformers/utilities.ts:642:83 | isNonStaticMethodOrAccessorWithPrivateName | 642 | predicate call |
| src/compiler/utilities.ts:2088:55 | isNonGlobalAmbientModule | 2088 | predicate call |
| src/compiler/utilities.ts:2129:59 | isExternalModuleAugmentation | 2129 | predicate call |
| src/compiler/utilities.ts:2232:62 | isDeclarationWithTypeParameters | 2232 | kind switch, predicate call |
| src/compiler/utilities.ts:2246:69 | isDeclarationWithTypeParameterChildren | 2246 | kind switch, predicate call |
| src/compiler/utilities.ts:2287:65 | isAnyImportOrBareOrAccessedRequire | 2287 | predicate call |
| src/compiler/utilities.ts:2292:60 | isAnyImportOrRequireStatement | 2292 | predicate call |
| src/compiler/utilities.ts:2315:65 | hasPossibleExternalModuleReference | 2315 | predicate call |
| src/compiler/utilities.ts:2320:52 | isAnyImportOrReExport | 2320 | predicate call |
| src/compiler/utilities.ts:2717:40 | isImportMeta | 2717 | predicate call |
| src/compiler/utilities.ts:2724:51 | isLiteralImportTypeNode | 2724 | predicate call |
| src/compiler/utilities.ts:3117:60 | <callback or function type> | 3117 | predicate call |
| src/compiler/utilities.ts:3839:94 | isRequireCall | 3843 | predicate call |
| src/compiler/utilities.ts:3841:97 | isRequireCall | 3843 | predicate call |
| src/compiler/utilities.ts:3843:97 | isRequireCall | 3843 | predicate call |
| src/compiler/utilities.ts:3880:70 | isBindingElementOfBareOrAccessedRequire | 3880 | predicate call |
| src/compiler/utilities.ts:3905:57 | isRequireVariableStatement | 3905 | predicate call |
| src/compiler/utilities.ts:4104:62 | isModuleExportsAccessExpression | 4104 | predicate call |
| src/compiler/utilities.ts:4119:75 | isBindableObjectDefinePropertyCall | 4119 | predicate call |
| src/compiler/utilities.ts:4132:43 | isLiteralLikeAccess | 4132 | predicate call |
| src/compiler/utilities.ts:4139:50 | isLiteralLikeElementAccess | 4139 | predicate call |
| src/compiler/utilities.ts:4166:91 | isBindableStaticNameExpression | 4166 | predicate call |
| src/compiler/utilities.ts:4294:60 | isPrototypePropertyAssignment | 4294 | predicate call |
| src/compiler/utilities.ts:4353:59 | <callback or function type> | 4353 | predicate call |
| src/compiler/utilities.ts:4487:42 | isTypeAlias | 4487 | predicate call |
| src/compiler/utilities.ts:4997:58 | isValueSignatureDeclaration | 4997 | predicate call |
| src/compiler/utilities.ts:5276:60 | isKeywordOrPunctuation | 5276 | predicate call |
| src/compiler/utilities.ts:5362:59 | isStringOrNumericLiteralLike | 5362 | predicate call |
| src/compiler/utilities.ts:5367:53 | isSignedNumericLiteral | 5367 | predicate call |
| src/compiler/utilities.ts:5522:54 | isNamedEvaluationSource | 5522 | kind switch, predicate call |
| src/compiler/utilities.ts:5564:101 | isNamedEvaluation | 5564 | kind switch, predicate call |
| src/compiler/utilities.ts:6760:41 | <callback or function type> | 6760 | predicate call |
| src/compiler/utilities.ts:7420:72 | isLogicalOrCoalescingAssignmentExpression | 7420 | predicate call |
| src/compiler/utilities.ts:7430:68 | isLogicalOrCoalescingBinaryExpression | 7430 | predicate call |
| src/compiler/utilities.ts:7535:67 | isPropertyAccessEntityNameExpression | 7535 | predicate call |
| src/compiler/utilities.ts:7563:48 | isPrototypeAccess | 7563 | predicate call |
| src/compiler/utilities.ts:8024:52 | isInitializedVariable | 8024 | predicate call |
| src/compiler/utilities.ts:8325:54 | isObjectTypeDeclaration | 8325 | predicate call |
| src/compiler/utilities.ts:10624:56 | isIdentifierTypeReference | 10624 | predicate call |
| src/compiler/utilities.ts:11048:52 | canHaveExportModifier | 11048 | predicate call |
| src/compiler/utilities.ts:11111:57 | isJSDocSatisfiesExpression | 11111 | predicate call |
| src/compiler/utilities.ts:11181:85 | isExpandoPropertyDeclaration | 11181 | predicate call |
| src/compiler/utilities.ts:12030:82 | isPrimitiveLiteralValue | 12030 | kind switch, predicate call |
| src/compiler/utilities.ts:12064:46 | hasInferredType | 12064 | kind switch, predicate call |
| src/compiler/utilities.ts:12218:45 | isNewScopeNode | 12218 | predicate call |
| src/compiler/utilities.ts:12298:60 | <callback or function type> | 12298 | predicate call |
| src/compiler/utilities.ts:12431:48 | canHaveStatements | 12431 | predicate call |
| src/compiler/utilitiesPublic.ts:619:59 | isEmptyBindingPattern | 619 | predicate call |
| src/compiler/utilitiesPublic.ts:1040:76 | <callback or function type> | 1040 | predicate call |
| src/compiler/utilitiesPublic.ts:1078:68 | <callback or function type> | 1078 | predicate call |
| src/compiler/utilitiesPublic.ts:1390:50 | isOptionalChainRoot | 1390 | predicate call |
| src/compiler/utilitiesPublic.ts:1399:62 | isExpressionOfOptionalChainRoot | 1399 | predicate call |
| src/compiler/utilitiesPublic.ts:1498:50 | isLiteralExpression | 1498 | predicate call |
| src/compiler/utilitiesPublic.ts:1522:53 | isTemplateLiteralToken | 1522 | predicate call |
| src/compiler/utilitiesPublic.ts:1532:56 | isImportOrExportSpecifier | 1532 | predicate call |
| src/compiler/utilitiesPublic.ts:1562:66 | isTypeOnlyImportOrExportDeclaration | 1562 | predicate call |
| src/compiler/utilitiesPublic.ts:1574:52 | isImportAttributeName | 1574 | predicate call |
| src/compiler/utilitiesPublic.ts:1581:52 | isGeneratedIdentifier | 1581 | predicate call |
| src/compiler/utilitiesPublic.ts:1586:59 | isGeneratedPrivateIdentifier | 1586 | predicate call |
| src/compiler/utilitiesPublic.ts:1600:73 | isPrivateIdentifierClassElementDeclaration | 1600 | predicate call |
| src/compiler/utilitiesPublic.ts:1605:74 | isPrivateIdentifierPropertyAccessExpression | 1605 | predicate call |
| src/compiler/utilitiesPublic.ts:1647:41 | isModifier | 1647 | predicate call |
| src/compiler/utilitiesPublic.ts:1680:86 | isFunctionLikeOrClassStaticBlockDeclaration | 1680 | predicate call |
| src/compiler/utilitiesPublic.ts:1752:64 | isAutoAccessorPropertyDeclaration | 1752 | predicate call |
| src/compiler/utilitiesPublic.ts:1778:45 | isModifierLike | 1778 | predicate call |
| src/compiler/utilitiesPublic.ts:1794:51 | isClassOrTypeElement | 1794 | predicate call |
| src/compiler/utilitiesPublic.ts:1815:41 | isTypeNode | 1815 | predicate call |
| src/compiler/utilitiesPublic.ts:1872:59 | isBindingOrAssignmentElement | 1872 | predicate call |
| src/compiler/utilitiesPublic.ts:1884:87 | isBindingOrAssignmentPattern | 1884 | predicate call |
| src/compiler/utilitiesPublic.ts:1932:64 | isArrayBindingOrAssignmentElement | 1932 | kind switch, predicate call |
| src/compiler/utilitiesPublic.ts:1964:65 | isCallLikeOrFunctionLikeExpression | 1964 | predicate call |
| src/compiler/utilitiesPublic.ts:2071:51 | isLiteralTypeLiteral | 2071 | kind switch, predicate call |
| src/compiler/utilitiesPublic.ts:2116:83 | isIterationStatement | 2118 | kind switch, predicate call |
| src/compiler/utilitiesPublic.ts:2117:85 | isIterationStatement | 2118 | kind switch, predicate call |
| src/compiler/utilitiesPublic.ts:2118:85 | isIterationStatement | 2118 | kind switch, predicate call |
| src/compiler/utilitiesPublic.ts:2160:44 | isConciseBody | 2160 | predicate call |
| src/compiler/utilitiesPublic.ts:2166:45 | isFunctionBody | 2166 | predicate call |
| src/compiler/utilitiesPublic.ts:2170:47 | isForInitializer | 2170 | predicate call |
| src/compiler/utilitiesPublic.ts:2406:53 | isDeclarationStatement | 2406 | predicate call |
| src/compiler/utilitiesPublic.ts:2415:59 | isStatementButNotDeclaration | 2415 | predicate call |
| src/compiler/utilitiesPublic.ts:2419:42 | isStatement | 2419 | predicate call |
| src/compiler/utilitiesPublic.ts:2552:44 | hasJSDocNodes | 2552 | predicate call |

## something else (74)

| Source location | Name | Implementation line | Features |
| --- | --- | ---: | --- |
| src/compiler/builder.ts:268:87 | isBuilderProgramStateWithDefinedProgram | 268 |  |
| src/compiler/builder.ts:1176:79 | isIncrementalBundleEmitBuildInfo | 1176 |  |
| src/compiler/builder.ts:1181:58 | isIncrementalBuildInfo | 1181 |  |
| src/compiler/checker.ts:26489:43 | <callback or function type> | 26489 |  |
| src/compiler/checker.ts:51062:48 | canHaveConstantValue | 51062 | kind switch |
| src/compiler/commandLineParser.ts:3038:37 | isNullOrUndefined | 3038 |  |
| src/compiler/core.ts:140:116 | every | 145 |  |
| src/compiler/core.ts:142:128 | every | 145 |  |
| src/compiler/core.ts:614:59 | some | 618 |  |
| src/compiler/emitter.ts:1158:27 | <callback or function type> | 1158 |  |
| src/compiler/moduleNameResolver.ts:925:82 | isPackageJsonInfo | 925 |  |
| src/compiler/moduleNameResolver.ts:930:89 | isMissingPackageJsonInfo | 930 |  |
| src/compiler/program.ts:1162:74 | isReferencedFile | 1162 | kind switch |
| src/compiler/program.ts:1190:108 | isReferenceFileLocation | 1190 |  |
| src/compiler/sourcemap.ts:630:52 | isSourceMapping | 630 |  |
| src/compiler/sourcemap.ts:668:57 | isSourceMappedPosition | 668 |  |
| src/compiler/transformers/declarations.ts:1891:49 | canHaveLiteralInitializer | 1891 | kind switch |
| src/compiler/transformers/declarations.ts:1916:55 | isPreservedDeclarationStatement | 1916 | kind switch |
| src/compiler/transformers/declarations.ts:1954:44 | isProcessedComponent | 1954 | kind switch |
| src/compiler/transformers/es2015.ts:2029:122 | shouldAddRestParameter | 2029 |  |
| src/compiler/transformers/ts.ts:1197:62 | shouldAddParamTypesMetadata | 1197 | kind switch |
| src/compiler/transformers/ts.ts:1297:93 | shouldEmitFunctionLikeDeclaration | 1297 |  |
| src/compiler/transformers/utilities.ts:490:61 | isCompoundAssignment | 490 |  |
| src/compiler/tsbuildPublic.ts:259:66 | isCircularBuildOrder | 259 |  |
| src/compiler/tsbuildPublic.ts:569:60 | isParsedCommandLine | 569 |  |
| src/compiler/tsbuildPublic.ts:1371:84 | isFileWatcherWithModifiedTime | 1371 |  |
| src/compiler/utilities.ts:2276:48 | isAnyImportSyntax | 2276 | kind switch |
| src/compiler/utilities.ts:2297:63 | isLateVisibilityPaintedStatement | 2297 | kind switch |
| src/compiler/utilities.ts:2643:53 | isJsonSourceFile | 2643 |  |
| src/compiler/utilities.ts:2996:45 | isVariableLike | 2996 | kind switch |
| src/compiler/utilities.ts:3866:72 | isVariableDeclarationInitializedToRequire | 3866 |  |
| src/compiler/utilities.ts:3875:86 | isVariableDeclarationInitializedToBareOrAccessedRequire | 3875 |  |
| src/compiler/utilities.ts:4330:65 | canHaveModuleSpecifier | 4330 | kind switch |
| src/compiler/utilities.ts:4535:46 | canHaveFlowNode | 4535 | kind switch |
| src/compiler/utilities.ts:4561:43 | canHaveJSDoc | 4561 | kind switch |
| src/compiler/utilities.ts:4837:47 | hasTypeArguments | 4837 |  |
| src/compiler/utilities.ts:4964:67 | isNodeWithPossibleHoistedDeclaration | 4964 | kind switch |
| src/compiler/utilities.ts:5266:47 | isKeyword | 5266 |  |
| src/compiler/utilities.ts:5271:51 | isPunctuation | 5271 |  |
| src/compiler/utilities.ts:5303:46 | isTrivia | 5303 |  |
| src/compiler/utilities.ts:5381:59 | hasDynamicName | 5381 |  |
| src/compiler/utilities.ts:5427:52 | isPropertyNameLiteral | 5427 | kind switch |
| src/compiler/utilities.ts:7413:77 | isLogicalOrCoalescingAssignmentOperator | 7413 |  |
| src/compiler/utilities.ts:7425:73 | isLogicalOrCoalescingBinaryOperator | 7425 |  |
| src/compiler/utilities.ts:7495:80 | isExpressionWithTypeArgumentsInClassExtendsClause | 7495 |  |
| src/compiler/utilities.ts:9025:62 | usesWildcardTypes | 9025 |  |
| src/compiler/utilities.ts:11025:48 | isTypeDeclaration | 11025 | kind switch |
| src/compiler/utilities.ts:11997:83 | isSelfReferenceLocation | 11997 | kind switch |
| src/compiler/utilitiesPublic.ts:946:49 | isNamedDeclaration | 946 |  |
| src/compiler/utilitiesPublic.ts:1487:67 | isNodeArray | 1487 |  |
| src/compiler/utilitiesPublic.ts:1494:50 | isLiteralKind | 1494 |  |
| src/compiler/utilitiesPublic.ts:1518:58 | isTemplateLiteralKind | 1518 |  |
| src/compiler/utilitiesPublic.ts:1536:58 | isTypeOnlyImportDeclaration | 1536 | kind switch |
| src/compiler/utilitiesPublic.ts:1550:58 | isTypeOnlyExportDeclaration | 1550 | kind switch |
| src/compiler/utilitiesPublic.ts:1612:52 | isModifierKind | 1612 |  |
| src/compiler/utilitiesPublic.ts:1675:57 | isFunctionLike | 1675 |  |
| src/compiler/utilitiesPublic.ts:1685:56 | isFunctionLikeDeclaration | 1685 |  |
| src/compiler/utilitiesPublic.ts:1765:49 | isMethodOrAccessor | 1765 | kind switch |
| src/compiler/utilitiesPublic.ts:1819:62 | isFunctionOrConstructorTypeNode | 1819 | kind switch |
| src/compiler/utilitiesPublic.ts:1860:90 | isDeclarationBindingElement | 1860 | kind switch |
| src/compiler/utilitiesPublic.ts:1894:93 | isObjectBindingOrAssignmentPattern | 1894 | kind switch |
| src/compiler/utilitiesPublic.ts:1905:65 | isObjectBindingOrAssignmentElement | 1905 | kind switch |
| src/compiler/utilitiesPublic.ts:1921:92 | isArrayBindingOrAssignmentPattern | 1921 | kind switch |
| src/compiler/utilitiesPublic.ts:1995:55 | isLeftHandSideExpression | 1995 |  |
| src/compiler/utilitiesPublic.ts:2039:48 | isUnaryExpression | 2039 |  |
| src/compiler/utilitiesPublic.ts:2059:57 | isUnaryExpressionWithWrite | 2059 | kind switch |
| src/compiler/utilitiesPublic.ts:2086:43 | isExpression | 2086 |  |
| src/compiler/utilitiesPublic.ts:2208:44 | canHaveSymbol | 2208 | kind switch |
| src/compiler/utilitiesPublic.ts:2283:44 | canHaveLocals | 2283 | kind switch |
| src/compiler/utilitiesPublic.ts:2534:41 | isJSDocTag | 2534 |  |
| src/compiler/utilitiesPublic.ts:2564:38 | hasType | 2564 |  |
| src/compiler/utilitiesPublic.ts:2573:45 | hasInitializer | 2573 |  |
| src/compiler/utilitiesPublic.ts:2578:59 | hasOnlyExpressionInitializer | 2578 | kind switch |
| src/compiler/watch.ts:334:96 | isBuilderProgram | 334 |  |

## typeof (8)

| Source location | Name | Implementation line | Features |
| --- | --- | ---: | --- |
| src/compiler/commandLineParser.ts:2606:85 | isCompilerOptionsValue | 2606 | predicate call, typeof |
| src/compiler/core.ts:1769:42 | isString | 1769 | typeof |
| src/compiler/core.ts:1773:39 | isNumber | 1773 | typeof |
| src/compiler/sourcemap.ts:408:34 | isRawSourceMap | 408 | predicate call, typeof |
| src/compiler/utilities.ts:5485:112 | isAnonymousFunctionDefinition | 5485 | kind switch, typeof |
| src/compiler/utilities.ts:8634:119 | isDiagnosticWithDetachedLocation | 8634 | typeof |
| src/compiler/watchPublic.ts:739:77 | isFileMissingOnHost | 739 | typeof |
| src/compiler/watchPublic.ts:743:83 | isFilePresenceUnknownOnHost | 743 | typeof |
