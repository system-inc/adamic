Unit timing includes registration, lowering and panic recovery. Ratio = B/A.

{
  "A_total_seconds": 652.345069746,
  "B_total_seconds": 654.044471945,
  "A_cpu_seconds": 702.8921030000013,
  "B_cpu_seconds": 703.4416810000038,
  "units": 4017,
  "controls": {
    "units": 1215,
    "median_absolute_delta": 0.0036352200000000015,
    "p95_absolute_delta": 0.055006476,
    "median_ratio": 1.011537887196256
  },
  "control_definition": "Imports, interfaces and type aliases with identical lowering findings: no checked-predicate body to exercise B. These include registration costs; one paired run cannot separate scheduling, GC and common registration overhead."
}

| Unit | Name | A seconds | B seconds | B-A seconds | B/A |
|---|---|---:|---:|---:|---:|
| src/compiler/watch.ts:193:1 | createWatchStatusReporter | 5.361410 | 6.575393 | +1.213984 | 1.2264 |
| src/compiler/utilities.ts:9402:1 | getCompilerOptionValue | 5.179633 | 6.390054 | +1.210421 | 1.2337 |
| src/compiler/utilities.ts:7819:1 | directoryProbablyExists | 0.165651 | 0.777241 | +0.611590 | 4.6920 |
| src/compiler/utilitiesPublic.ts:1426:1 | isConstTypeReference | 0.275215 | 0.846321 | +0.571106 | 3.0751 |
| src/compiler/utilitiesPublic.ts:1581:1 | isGeneratedIdentifier | 0.358056 | 0.793115 | +0.435059 | 2.2151 |
| src/compiler/utilities.ts:3931:1 | getEffectiveInitializer | 0.299267 | 0.683738 | +0.384470 | 2.2847 |
| src/compiler/utilities.ts:4041:1 | getNameOfExpando | 0.305568 | 0.683816 | +0.378248 | 2.2379 |
| src/compiler/transformers/destructuring.ts:583:1 | ensureIdentifier | 0.255606 | 0.617519 | +0.361913 | 2.4159 |
| src/compiler/utilities.ts:5397:1 | getPropertyNameForPropertyNameNode | 0.591855 | 0.943540 | +0.351685 | 1.5942 |
| src/compiler/utilities.ts:7827:1 | getNewLineCharacter | 0.108286 | 0.459323 | +0.351037 | 4.2417 |
| src/compiler/sys.ts:593:1 | createDirectoryWatcherSupportingRecursive | 0.156700 | 0.481249 | +0.324550 | 3.0712 |
| src/compiler/transformers/utilities.ts:173:1 | getImportNeedsImportDefaultHelper | 0.303886 | 0.620676 | +0.316790 | 2.0425 |
| src/compiler/transformers/namedEvaluation.ts:67:1 | getAssignedNameOfPropertyName | 0.399939 | 0.698961 | +0.299022 | 1.7477 |
| src/compiler/utilitiesPublic.ts:1574:1 | isImportAttributeName | 0.302020 | 0.580395 | +0.278375 | 1.9217 |
| src/compiler/utilities.ts:810:1 | moduleResolutionIsEqualTo | 4.408525 | 4.670652 | +0.262127 | 1.0595 |
| src/compiler/factory/utilities.ts:814:1 | getExternalModuleNameLiteral | 0.308260 | 0.559940 | +0.251681 | 1.8165 |
| src/compiler/utilities.ts:2481:1 | createDiagnosticMessageChainFromDiagnostic | 5.870357 | 6.103276 | +0.232918 | 1.0397 |
| src/compiler/utilitiesPublic.ts:1422:1 | isNullishCoalesce | 0.218539 | 0.450921 | +0.232382 | 2.0633 |
| src/compiler/sourcemap.ts:630:1 | isSourceMapping | 4.396320 | 4.599438 | +0.203119 | 1.0462 |
| src/compiler/emitter.ts:1147:1 |  | 0.094500 | 0.297617 | +0.203116 | 3.1494 |
| src/compiler/moduleNameResolver.ts:511:1 | arePathsEqual | 4.925083 | 5.122049 | +0.196966 | 1.0400 |
| src/compiler/utilitiesPublic.ts:2071:1 | isLiteralTypeLiteral | 0.055550 | 0.251968 | +0.196418 | 4.5359 |
| src/compiler/utilities.ts:7200:1 | isStatic | 0.050581 | 0.245392 | +0.194811 | 4.8515 |
| src/compiler/utilities.ts:4064:1 | isSameEntityName | 0.290632 | 0.484735 | +0.194103 | 1.6679 |
| src/compiler/utilities.ts:4330:1 | canHaveModuleSpecifier | 4.494963 | 4.681517 | +0.186554 | 1.0415 |
| src/compiler/utilities.ts:11077:1 | hasTabstop | 4.778293 | 4.961037 | +0.182745 | 1.0382 |
| src/compiler/utilities.ts:9044:1 |  | 5.305416 | 5.483298 | +0.177882 | 1.0335 |
| src/compiler/types.ts:2298:1 | UnionOrIntersectionTypeNode | 0.037955 | 0.200532 | +0.162577 | 5.2834 |
| src/compiler/utilities.ts:11057:1 | isOptionalJSDocPropertyLikeTag | 0.068011 | 0.226181 | +0.158170 | 3.3257 |
| src/compiler/types.ts:8635:1 | PropertyDescriptorAttributes | 0.039058 | 0.196510 | +0.157452 | 5.0313 |
| src/compiler/types.ts:2300:1 | UnionTypeNode | 0.039313 | 0.193835 | +0.154522 | 4.9306 |
| src/compiler/types.ts:8615:1 | EmitHost | 0.041833 | 0.195753 | +0.153919 | 4.6794 |
| src/compiler/utilities.ts:7825:1 |  | 0.060509 | 0.213630 | +0.153121 | 3.5305 |
| src/compiler/moduleSpecifiers.ts:288:1 | getNodeModulesPackageName | 4.682183 | 4.829432 | +0.147249 | 1.0314 |
| src/compiler/utilitiesPublic.ts:1794:1 | isClassOrTypeElement | 0.145193 | 0.286333 | +0.141140 | 1.9721 |
| src/compiler/program.ts:1339:1 | getImpliedNodeFormatForFile | 4.621959 | 4.762853 | +0.140894 | 1.0305 |
| src/compiler/utilities.ts:7824:1 |  | 0.058425 | 0.191685 | +0.133260 | 3.2809 |
| src/compiler/builder.ts:904:1 | handleDtsMayChangeOfGlobalScope | 4.622927 | 4.755685 | +0.132757 | 1.0287 |
| src/compiler/utilities.ts:7809:1 | tryParseJson | 0.053327 | 0.182460 | +0.129133 | 3.4215 |
| src/compiler/utilitiesPublic.ts:1884:1 | isBindingOrAssignmentPattern | 0.308548 | 0.434509 | +0.125961 | 1.4082 |
| src/compiler/utilities.ts:5439:1 | getTextOfIdentifierOrLiteral | 0.387481 | 0.510411 | +0.122930 | 1.3173 |
| src/compiler/performance.ts:29:1 | createTimerIf | 0.038422 | 0.158724 | +0.120302 | 4.1311 |
| src/compiler/scanner.ts:557:1 | isWhiteSpaceSingleLine | 0.038747 | 0.158933 | +0.120186 | 4.1018 |
| src/compiler/moduleNameResolver.ts:1406:1 | resolveModuleName | 4.906066 | 5.022547 | +0.116481 | 1.0237 |
| src/compiler/performance.ts:1:1 |  | 0.041954 | 0.157561 | +0.115607 | 3.7556 |
| src/compiler/types.ts:4019:1 | JSDocPublicTag | 0.046967 | 0.162163 | +0.115197 | 3.4527 |
| src/compiler/types.ts:4056:1 | JSDocSeeTag | 0.053521 | 0.167897 | +0.114376 | 3.1370 |
| src/compiler/factory/nodeTests.ts:412:1 | isDecorator | 0.054771 | 0.169126 | +0.114356 | 3.0879 |
| src/compiler/utilities.ts:7858:1 | moveRangeEnd | 0.064770 | 0.178265 | +0.113495 | 2.7523 |
| src/compiler/types.ts:4045:1 | JSDocThisTag | 0.053887 | 0.166617 | +0.112730 | 3.0920 |
| src/compiler/sys.ts:382:1 | createUseFsEventsOnParentDirectoryWatchFile | 0.199983 | 0.312574 | +0.112591 | 1.5630 |
| src/compiler/utilities.ts:9621:1 |  | 0.079188 | 0.191519 | +0.112331 | 2.4185 |
| src/compiler/builder.ts:790:1 | removeDiagnosticsOfLibraryFiles | 4.803524 | 4.914459 | +0.110935 | 1.0231 |
| src/compiler/utilities.ts:5387:1 | isDynamicName | 0.354614 | 0.463176 | +0.108562 | 1.3061 |
| src/compiler/types.ts:2310:1 | ConditionalTypeNode | 0.038711 | 0.146863 | +0.108152 | 3.7938 |
| src/compiler/moduleNameResolver.ts:3242:1 | tryFindNonRelativeModuleNameInCache | 4.703905 | 4.811235 | +0.107330 | 1.0228 |
| src/compiler/scanner.ts:596:1 | isHexDigit | 0.037488 | 0.143169 | +0.105681 | 3.8191 |
| src/compiler/types.ts:2305:1 | IntersectionTypeNode | 0.038963 | 0.142833 | +0.103870 | 3.6658 |
| src/compiler/utilities.ts:8818:1 | compareMessageText | 4.761080 | 4.864710 | +0.103630 | 1.0218 |
| src/compiler/utilities.ts:3953:1 | hasExpandoValueProperty | 0.057614 | 0.159852 | +0.102238 | 2.7745 |
| src/compiler/moduleSpecifiers.ts:1186:1 | tryGetModuleNameAsNodeModule | 0.278047 | 0.377688 | +0.099641 | 1.3584 |
| src/compiler/commandLineParser.ts:3323:1 | getSubstitutedMapLikeOfStringArrayWithConfigDirTemplate | 0.179780 | 0.276462 | +0.096682 | 1.5378 |
| src/compiler/utilitiesPublic.ts:1432:1 | skipPartiallyEmittedExpressions | 0.038062 | 0.131900 | +0.093838 | 3.4654 |
| src/compiler/factory/nodeFactory.ts:7236:1 |  | 0.071700 | 0.164932 | +0.093232 | 2.3003 |
| src/compiler/types.ts:4050:1 | JSDocTemplateTag | 0.054331 | 0.146664 | +0.092333 | 2.6995 |
| src/compiler/factory/utilities.ts:1151:1 | isQuestionOrExclamationToken | 0.269403 | 0.361446 | +0.092043 | 1.3417 |
| src/compiler/utilities.ts:7471:1 | isAssignmentExpression | 0.295657 | 0.386251 | +0.090594 | 1.3064 |
| src/compiler/types.ts:9563:1 | LexicalEnvironmentFlags | 0.038734 | 0.128821 | +0.090087 | 3.3257 |
| src/compiler/types.ts:8644:1 | OuterExpressionKinds | 0.039987 | 0.129020 | +0.089032 | 3.2265 |
| src/compiler/utilities.ts:5229:1 | getAllSuperTypeNodes | 0.303338 | 0.391270 | +0.087933 | 1.2899 |
| src/compiler/types.ts:2374:1 | PropertyNameLiteral | 0.043233 | 0.131117 | +0.087884 | 3.0328 |
| src/compiler/utilities.ts:8936:1 | getDiagnosticCode | 4.544646 | 4.632444 | +0.087797 | 1.0193 |
| src/compiler/builder.ts:1227:1 | getBuildInfoEmitPending | 4.737005 | 4.824460 | +0.087455 | 1.0185 |
| src/compiler/utilitiesPublic.ts:2592:1 | isObjectLiteralElement | 0.234061 | 0.318557 | +0.084496 | 1.3610 |
| src/compiler/transformers/namedEvaluation.ts:56:1 | getAssignedNameOfIdentifier | 0.069849 | 0.154006 | +0.084157 | 2.2049 |
| src/compiler/types.ts:2323:1 | ParenthesizedTypeNode | 0.038484 | 0.121763 | +0.083279 | 3.1640 |
| src/compiler/types.ts:2328:1 | TypeOperatorNode | 0.039712 | 0.121326 | +0.081613 | 3.0551 |
| src/compiler/transformers/module/system.ts:137:1 | transformSystemModule | 0.556089 | 0.636858 | +0.080768 | 1.1452 |
| src/compiler/performance.ts:15:1 |  | 0.041371 | 0.120931 | +0.079560 | 2.9231 |
| src/compiler/scanner.ts:609:1 | isOctalDigit | 0.038077 | 0.117426 | +0.079350 | 3.0839 |
| src/compiler/utilities.ts:7845:1 | createRange | 0.054843 | 0.134027 | +0.079184 | 2.4438 |
| src/compiler/commandLineParser.ts:3933:1 | getFileNamesFromConfigSpecs | 0.160751 | 0.238366 | +0.077615 | 1.4828 |
| src/compiler/types.ts:4105:1 | JSDocPropertyLikeTag | 0.039804 | 0.117325 | +0.077521 | 2.9475 |
| src/compiler/types.ts:2318:1 | InferTypeNode | 0.038673 | 0.115452 | +0.076778 | 2.9853 |
| src/compiler/factory/nodeTests.ts:624:1 | isTypeOfExpression | 0.043878 | 0.120293 | +0.076415 | 2.7415 |
| src/compiler/performance.ts:23:1 | Timer | 0.038532 | 0.114261 | +0.075729 | 2.9654 |
| src/compiler/transformers/utilities.ts:179:1 | collectExternalModuleInfo | 0.043831 | 0.119554 | +0.075723 | 2.7276 |
| src/compiler/types.ts:4027:1 | JSDocProtectedTag | 0.042616 | 0.118043 | +0.075427 | 2.7699 |
| src/compiler/expressionToTypeNode.ts:174:1 | createSyntacticTypeNodeBuilder | 0.202150 | 0.276812 | +0.074662 | 1.3693 |
| src/compiler/utilities.ts:9610:1 |  | 0.097146 | 0.171321 | +0.074175 | 1.7635 |
| src/compiler/factory/nodeTests.ts:1082:1 | isJSDoc | 0.044724 | 0.118833 | +0.074109 | 2.6570 |
| src/compiler/utilities.ts:4094:1 | isExportsIdentifier | 0.261032 | 0.335118 | +0.074086 | 1.2838 |
| src/compiler/emitter.ts:1209:1 |  | 0.039710 | 0.112933 | +0.073223 | 2.8439 |
| src/compiler/transformers/generators.ts:318:1 | getInstructionName | 0.073530 | 0.146158 | +0.072628 | 1.9877 |
| src/compiler/utilities.ts:8637:1 | isDiagnosticWithDetachedLocation | 4.641959 | 4.714139 | +0.072181 | 1.0155 |
| src/compiler/factory/emitNode.ts:36:1 | getOrCreateEmitNode | 0.054211 | 0.126006 | +0.071795 | 2.3244 |
| src/compiler/utilities.ts:6144:1 | containsInvalidEscapeFlag | 4.688897 | 4.760079 | +0.071183 | 1.0152 |
| src/compiler/types.ts:8739:1 | ImmediatelyInvokedFunctionExpression | 0.040424 | 0.111109 | +0.070685 | 2.7486 |
| src/compiler/sys.ts:540:1 | onWatchedFileStat | 0.080078 | 0.149301 | +0.069222 | 1.8644 |
| src/compiler/utilities.ts:5381:1 | hasDynamicName | 0.077758 | 0.145753 | +0.067995 | 1.8744 |
| src/compiler/scanner.ts:613:1 | couldStartTrivia | 0.041707 | 0.109148 | +0.067441 | 2.6170 |
| src/compiler/utilitiesPublic.ts:1872:1 | isBindingOrAssignmentElement | 0.296806 | 0.363656 | +0.066851 | 1.2252 |
| src/compiler/transformers/utilities.ts:866:1 | isSimpleParameter | 0.263855 | 0.329579 | +0.065724 | 1.2491 |
| src/compiler/types.ts:8731:1 | GeneratedNamePart | 0.040447 | 0.105937 | +0.065490 | 2.6191 |
| src/compiler/utilitiesPublic.ts:1433:1 | skipPartiallyEmittedExpressions | 0.039153 | 0.103548 | +0.064394 | 2.6447 |
| src/compiler/utilitiesPublic.ts:1351:1 | getEffectiveConstraintOfTypeParameter | 0.302903 | 0.367172 | +0.064269 | 1.2122 |
| src/compiler/types.ts:9569:1 | CoreTransformationContext | 0.039067 | 0.103308 | +0.064241 | 2.6444 |
| src/compiler/utilities.ts:8924:1 | diagnosticsEqualityComparer | 4.398175 | 4.462269 | +0.064094 | 1.0146 |
| src/compiler/scanner.ts:574:1 | isLineBreak | 0.037319 | 0.101376 | +0.064057 | 2.7165 |
| src/compiler/types.ts:8740:1 | ImmediatelyInvokedArrowFunction | 0.041188 | 0.104511 | +0.063323 | 2.5374 |
| src/compiler/factory/nodeTests.ts:1086:1 | isJSDocTypeLiteral | 0.045379 | 0.108633 | +0.063254 | 2.3939 |
| src/compiler/utilities.ts:3948:1 | getDeclaredExpandoInitializer | 0.244178 | 0.306775 | +0.062597 | 1.2564 |
| src/compiler/utilitiesPublic.ts:1680:1 | isFunctionLikeOrClassStaticBlockDeclaration | 0.242171 | 0.304514 | +0.062343 | 1.2574 |
| src/compiler/utilities.ts:7870:1 | moveRangePos | 0.069839 | 0.131818 | +0.061979 | 1.8875 |
| src/compiler/factory/emitNode.ts:1:1 |  | 0.041637 | 0.103446 | +0.061808 | 2.4844 |
| src/compiler/emitter.ts:573:1 | getOutputDeclarationFileNameWorker | 0.051922 | 0.113549 | +0.061627 | 2.1869 |
| src/compiler/performance.ts:20:1 |  | 0.038892 | 0.100294 | +0.061402 | 2.5788 |
| src/compiler/scanner.ts:605:1 | isWordCharacter | 0.039251 | 0.100496 | +0.061245 | 2.5604 |
| src/compiler/transformers/utilities.ts:355:1 | collectExportedVariableInfo | 0.193725 | 0.254749 | +0.061024 | 1.3150 |
| src/compiler/types.ts:4098:1 | JSDocSignature | 0.041450 | 0.101921 | +0.060471 | 2.4589 |
| src/compiler/types.ts:8718:1 | NodeConverters | 0.040259 | 0.100196 | +0.059937 | 2.4888 |
| src/compiler/types.ts:8742:1 | NodeFactory | 0.041895 | 0.101674 | +0.059779 | 2.4269 |
| src/compiler/debug.ts:1:1 |  | 0.042107 | 0.101644 | +0.059538 | 2.4140 |
| src/compiler/utilities.ts:10627:1 | isIdentifierTypeReference | 0.314832 | 0.374208 | +0.059376 | 1.1886 |
| src/compiler/types.ts:2335:1 | UniqueTypeOperatorNode | 0.042113 | 0.101355 | +0.059242 | 2.4067 |
| src/compiler/types.ts:9610:1 | TransformationContext | 0.039259 | 0.098301 | +0.059042 | 2.5039 |
| src/compiler/factory/nodeTests.ts:1090:1 | isJSDocSignature | 0.044389 | 0.103377 | +0.058988 | 2.3289 |
| src/compiler/utilities.ts:7804:1 | readJson | 0.068347 | 0.126972 | +0.058625 | 1.8578 |
| src/compiler/utilities.ts:5367:1 | isSignedNumericLiteral | 0.426429 | 0.484972 | +0.058543 | 1.1373 |
| src/compiler/checker.ts:1354:1 | SignatureCheckMode | 0.040649 | 0.098571 | +0.057922 | 2.4249 |
| src/compiler/watch.ts:114:1 |  | 0.081088 | 0.138374 | +0.057286 | 1.7065 |
| src/compiler/factory/emitHelpers.ts:1486:1 | isCallToHelper | 0.282011 | 0.338661 | +0.056650 | 1.2009 |
| src/compiler/types.ts:4114:1 | JSDocPropertyTag | 0.046388 | 0.102983 | +0.056595 | 2.2200 |
| src/compiler/performance.ts:34:1 | createTimer | 0.039715 | 0.095446 | +0.055731 | 2.4032 |
| src/compiler/types.ts:8659:1 | OuterExpression | 0.040697 | 0.096369 | +0.055672 | 2.3680 |
| src/compiler/utilities.ts:12059:1 | unwrapParenthesizedExpression | 0.187897 | 0.243035 | +0.055138 | 1.2934 |
| src/compiler/types.ts:4023:1 | JSDocPrivateTag | 0.041655 | 0.095682 | +0.054027 | 2.2970 |
| src/compiler/utilities.ts:4178:1 | getAssignmentDeclarationKindWorker | 0.196484 | 0.250465 | +0.053981 | 1.2747 |
| src/compiler/types.ts:4035:1 | JSDocOverrideTag | 0.046518 | 0.100210 | +0.053692 | 2.1542 |
| src/compiler/types.ts:4015:1 | JSDocClassTag | 0.040709 | 0.094212 | +0.053503 | 2.3143 |
| src/compiler/factory/nodeTests.ts:620:1 | isDeleteExpression | 0.044520 | 0.097819 | +0.053298 | 2.1972 |
| src/compiler/types.ts:8674:1 | TypeOfTag | 0.038932 | 0.092227 | +0.053295 | 2.3689 |
| src/compiler/utilities.ts:11494:1 | createNameResolver | 0.323028 | 0.375085 | +0.052057 | 1.1612 |
| src/compiler/types.ts:8604:1 | SourceFileMayBeEmittedHost | 0.043070 | 0.094822 | +0.051752 | 2.2016 |
| src/compiler/factory/baseNodeFactory.ts:13:1 | BaseNodeFactory | 0.043713 | 0.094640 | +0.050927 | 2.1650 |
| src/compiler/moduleNameResolver.ts:3211:1 | getTypesPackageName | 0.045461 | 0.096198 | +0.050737 | 2.1161 |
| src/compiler/types.ts:4093:1 | JSDocThrowsTag | 0.050729 | 0.101258 | +0.050530 | 1.9961 |
| src/compiler/utilities.ts:12033:1 | isPrimitiveLiteralValue | 0.165159 | 0.215648 | +0.050489 | 1.3057 |
| src/compiler/utilitiesPublic.ts:1752:1 | isAutoAccessorPropertyDeclaration | 0.271948 | 0.322086 | +0.050137 | 1.1844 |
| src/compiler/types.ts:4031:1 | JSDocReadonlyTag | 0.044580 | 0.094351 | +0.049771 | 2.1165 |
| src/compiler/executeCommandLine.ts:1095:1 | reportSolutionBuilderTimes | 0.196441 | 0.246045 | +0.049604 | 1.2525 |
| src/compiler/utilities.ts:2036:1 | canUseOriginalText | 4.631715 | 4.680654 | +0.048939 | 1.0106 |
| src/compiler/types.ts:8669:1 | WrappedExpression | 0.040210 | 0.089033 | +0.048823 | 2.2142 |
| src/compiler/utilities.ts:9615:1 | WildcardMatcher | 0.117297 | 0.165503 | +0.048206 | 1.4110 |
| src/compiler/commandLineParser.ts:2603:1 | getCompilerOptionValueTypeString | 4.672324 | 4.720085 | +0.047761 | 1.0102 |
| src/compiler/utilities.ts:4750:1 | getParameterSymbolFromJSDoc | 0.193554 | 0.241234 | +0.047680 | 1.2463 |
| src/compiler/transformers/classFields.ts:3354:1 | isStaticPropertyDeclaration | 0.264100 | 0.311527 | +0.047427 | 1.1796 |
| src/compiler/sys.ts:495:1 | SingleFileWatcher | 0.040943 | 0.087985 | +0.047042 | 2.1489 |
| src/compiler/factory/nodeTests.ts:898:1 | isExportSpecifier | 0.045791 | 0.091744 | +0.045954 | 2.0036 |
| src/compiler/hostErrors.ts:2:1 | errorCode | 0.085581 | 0.130995 | +0.045413 | 1.5306 |
| src/compiler/executeCommandLine.ts:478:1 | printEasyHelp | 4.825877 | 4.870896 | +0.045020 | 1.0093 |
| src/compiler/utilities.ts:9612:1 |  | 0.096525 | 0.141538 | +0.045013 | 1.4663 |
| src/compiler/utilities.ts:7603:1 | isEmptyArrayLiteral | 0.161882 | 0.206891 | +0.045009 | 1.2780 |
| src/compiler/types.ts:2293:1 | RestTypeNode | 0.038553 | 0.083276 | +0.044723 | 2.1600 |
| src/compiler/transformers/namedEvaluation.ts:1:1 |  | 0.039390 | 0.083438 | +0.044048 | 2.1183 |
| src/compiler/utilities.ts:7741:1 | base64encode | 0.128385 | 0.172099 | +0.043714 | 1.3405 |
| src/compiler/utilities.ts:12221:1 | isNewScopeNode | 0.233187 | 0.276843 | +0.043656 | 1.1872 |
| src/compiler/utilities.ts:3824:1 | isJSDocIndexSignature | 0.312178 | 0.355495 | +0.043317 | 1.1388 |
| src/compiler/_namespaces/ts.ts:55:1 |  | 0.042837 | 0.085706 | +0.042869 | 2.0008 |
| src/compiler/types.ts:4039:1 | JSDocEnumTag | 0.053404 | 0.096257 | +0.042854 | 1.8024 |
| src/compiler/performance.ts:59:1 |  | 0.066482 | 0.108975 | +0.042493 | 1.6392 |
| src/compiler/transformers/classFields.ts:3358:1 | isStaticPropertyDeclarationOrClassStaticBlock | 0.223884 | 0.266160 | +0.042276 | 1.1888 |
| src/compiler/types.ts:4061:1 | JSDocReturnTag | 0.053163 | 0.095272 | +0.042109 | 1.7921 |
| src/compiler/moduleSpecifiers.ts:1345:1 | tryGetAnyFileFromPath | 0.079604 | 0.121544 | +0.041940 | 1.5269 |
| src/compiler/watchUtilities.ts:572:1 | isIgnoredFileFromWildCardWatching | 0.274163 | 0.316091 | +0.041928 | 1.1529 |
| src/compiler/builderPublic.ts:245:1 | createAbstractBuilder | 0.156517 | 0.198408 | +0.041890 | 1.2676 |
| src/compiler/types.ts:1635:1 | ConstKeyword | 0.042172 | 0.083623 | +0.041451 | 1.9829 |
| src/compiler/utilitiesPublic.ts:2638:1 | isRestParameter | 0.196844 | 0.237990 | +0.041146 | 1.2090 |
| src/compiler/utilitiesPublic.ts:1727:1 | isFunctionOrModuleBlock | 0.249720 | 0.290532 | +0.040811 | 1.1634 |
| src/compiler/factory/baseNodeFactory.ts:1:1 |  | 0.043279 | 0.084079 | +0.040799 | 1.9427 |
| src/compiler/utilitiesPublic.ts:1431:1 | skipPartiallyEmittedExpressions | 0.039630 | 0.080332 | +0.040702 | 2.0270 |
| src/compiler/factory/nodeTests.ts:628:1 | isVoidExpression | 0.045613 | 0.086023 | +0.040410 | 1.8859 |
| src/compiler/utilities.ts:5564:1 | isNamedEvaluation | 0.170167 | 0.210243 | +0.040075 | 1.2355 |
| src/compiler/factory/utilities.ts:215:1 | createJsxFactoryExpressionFromEntityName | 0.200827 | 0.240753 | +0.039927 | 1.1988 |
| src/compiler/sys.ts:183:1 | WatchedFileWithIsClosed | 0.044017 | 0.083919 | +0.039902 | 1.9065 |
| src/compiler/utilities.ts:8325:1 | isObjectTypeDeclaration | 0.277630 | 0.317512 | +0.039882 | 1.1436 |
| src/compiler/utilitiesPublic.ts:2522:1 | isJSDocCommentContainingNode | 0.299803 | 0.339248 | +0.039445 | 1.1316 |
| src/compiler/utilities.ts:3631:1 | isExpressionNode | 0.274748 | 0.313698 | +0.038950 | 1.1418 |
| src/compiler/types.ts:2376:1 | TemplateLiteralTypeNode | 0.045222 | 0.083972 | +0.038751 | 1.8569 |
| src/compiler/factory/utilities.ts:584:1 | findUseStrictPrologue | 0.286376 | 0.324321 | +0.037945 | 1.1325 |
| src/compiler/utilitiesPublic.ts:2116:1 | isIterationStatement | 0.149219 | 0.186597 | +0.037378 | 1.2505 |
| src/compiler/utilities.ts:5276:1 | isKeywordOrPunctuation | 0.115285 | 0.152488 | +0.037203 | 1.3227 |
| src/compiler/utilities.ts:2701:1 | isSuperCall | 0.166222 | 0.203236 | +0.037014 | 1.2227 |
| src/compiler/utilities.ts:4427:1 | getNamespaceDeclarationNode | 0.295937 | 0.332676 | +0.036739 | 1.1241 |
| src/compiler/commandLineParser.ts:4023:1 | isExcludedFile | 0.161701 | 0.198427 | +0.036726 | 1.2271 |
| src/compiler/types.ts:8592:1 | EmitHint | 0.044584 | 0.081119 | +0.036536 | 1.8195 |
| src/compiler/utilitiesPublic.ts:1968:1 | isCallLikeExpression | 0.222333 | 0.258194 | +0.035862 | 1.1613 |
| src/compiler/utilities.ts:3768:1 | isNamespaceReexportDeclaration | 0.291438 | 0.327233 | +0.035796 | 1.1228 |
| src/compiler/types.ts:2361:1 | StringLiteral | 0.042032 | 0.077480 | +0.035448 | 1.8434 |
| src/compiler/utilities.ts:5427:1 | isPropertyNameLiteral | 0.104326 | 0.139724 | +0.035398 | 1.3393 |
| src/compiler/program.ts:5190:1 | getModuleNameStringLiteralAt | 0.132579 | 0.167720 | +0.035141 | 1.2651 |
| src/compiler/utilities.ts:2884:1 | isPartOfTypeExpressionWithTypeArguments | 0.269681 | 0.304783 | +0.035102 | 1.1302 |
| src/compiler/utilities.ts:7609:1 | getLocalSymbolForExportDefault | 0.044425 | 0.079391 | +0.034966 | 1.7871 |
| src/compiler/utilities.ts:7749:1 | base64decode | 0.134302 | 0.168898 | +0.034596 | 1.2576 |
| src/compiler/executeCommandLine.ts:908:1 | performCompilation | 0.444499 | 0.478972 | +0.034473 | 1.0776 |
| src/compiler/types.ts:1701:1 | Identifier | 0.042350 | 0.076809 | +0.034459 | 1.8137 |
| src/compiler/moduleNameResolver.ts:1360:1 | createTypeReferenceDirectiveResolutionCache | 0.235696 | 0.269500 | +0.033805 | 1.1434 |
| src/compiler/utilities.ts:4212:1 | getElementOrPropertyAccessArgumentExpressionOrName | 0.201145 | 0.234912 | +0.033767 | 1.1679 |
| src/compiler/utilities.ts:4224:1 | getElementOrPropertyAccessName | 0.205493 | 0.239064 | +0.033571 | 1.1634 |
| src/compiler/transformers/utilities.ts:876:1 | rewriteModuleSpecifier | 0.271129 | 0.304626 | +0.033496 | 1.1235 |
| src/compiler/utilities.ts:8719:1 | createCompilerDiagnostic | 4.842446 | 4.875793 | +0.033347 | 1.0069 |
| src/compiler/sys.ts:563:1 |  | 0.044355 | 0.077450 | +0.033096 | 1.7462 |
| src/compiler/factory/utilities.ts:237:1 | createJsxFragmentFactoryExpression | 0.088626 | 0.121705 | +0.033079 | 1.3732 |
| src/compiler/transformers/utilities.ts:153:1 | getImportNeedsImportStarHelper | 0.210522 | 0.243175 | +0.032653 | 1.1551 |
| src/compiler/utilities.ts:5242:1 | getHeritageClause | 0.039337 | 0.071971 | +0.032634 | 1.8296 |
| src/compiler/utilities.ts:11114:1 | isJSDocSatisfiesExpression | 0.304449 | 0.337018 | +0.032569 | 1.1070 |
| src/compiler/types.ts:4007:1 | JSDocAuthorTag | 0.041243 | 0.073774 | +0.032532 | 1.7888 |
| src/compiler/utilities.ts:9294:1 |  | 0.093064 | 0.125400 | +0.032336 | 1.3475 |
| src/compiler/utilitiesPublic.ts:1964:1 | isCallLikeOrFunctionLikeExpression | 0.365110 | 0.397109 | +0.031999 | 1.0876 |
| src/compiler/moduleSpecifiers.ts:453:1 | computeModuleSpecifiers | 0.243767 | 0.275639 | +0.031872 | 1.1307 |
| src/compiler/types.ts:1634:1 | AsyncKeyword | 0.041104 | 0.072965 | +0.031862 | 1.7752 |
| src/compiler/utilities.ts:7667:1 |  | 0.041700 | 0.073462 | +0.031762 | 1.7617 |
| src/compiler/moduleNameResolver.ts:3227:1 | getPackageNameFromTypesPackageName | 0.042675 | 0.074332 | +0.031656 | 1.7418 |
| src/compiler/transformers/generators.ts:254:1 | ExceptionBlockState | 0.039623 | 0.071204 | +0.031582 | 1.7971 |
| src/compiler/factory/emitHelpers.ts:1281:1 |  | 0.071354 | 0.102578 | +0.031225 | 1.4376 |
| src/compiler/utilities.ts:7617:1 | isExportDefaultSymbol | 0.049007 | 0.079804 | +0.030797 | 1.6284 |
| src/compiler/program.ts:5178:1 | getModuleNames | 0.137144 | 0.167687 | +0.030543 | 1.2227 |
| src/compiler/utilities.ts:5255:1 | getAncestor | 0.050042 | 0.080364 | +0.030322 | 1.6059 |
| src/compiler/scanner.ts:600:1 | isASCIILetter | 0.039576 | 0.069604 | +0.030028 | 1.7587 |
| src/compiler/types.ts:4087:1 | JSDocOverloadTag | 0.048358 | 0.078203 | +0.029844 | 1.6172 |
| src/compiler/transformers/generators.ts:219:1 | Label | 0.043433 | 0.073252 | +0.029819 | 1.6866 |
| src/compiler/factory/nodeTests.ts:408:1 | isParameter | 0.056085 | 0.085640 | +0.029555 | 1.5270 |
| src/compiler/utilities.ts:7633:1 | getExpandedCharCodes | 0.042390 | 0.071919 | +0.029529 | 1.6966 |
| src/compiler/utilities.ts:4202:1 | isVoidZero | 0.317410 | 0.346934 | +0.029525 | 1.0930 |
| src/compiler/factory/nodeFactory.ts:7235:1 |  | 0.040738 | 0.070094 | +0.029356 | 1.7206 |
| src/compiler/utilitiesPublic.ts:1815:1 | isTypeNode | 0.078955 | 0.108247 | +0.029292 | 1.3710 |
| src/compiler/parser.ts:405:1 | SignatureFlags | 0.039751 | 0.069008 | +0.029257 | 1.7360 |
| src/compiler/resolutionCache.ts:1711:1 | resolutionIsSymlink | 0.333260 | 0.362496 | +0.029236 | 1.0877 |
| src/compiler/utilities.ts:10560:1 | isValidBigIntString | 0.304979 | 0.334033 | +0.029054 | 1.0953 |
| src/compiler/types.ts:3551:1 | ClassDeclaration | 0.058997 | 0.088024 | +0.029028 | 1.4920 |
| src/compiler/emitter.ts:1211:1 | createPrinter | 0.094213 | 0.123132 | +0.028919 | 1.3070 |
| src/compiler/types.ts:3989:1 | JSDocUnknownTag | 0.042992 | 0.071883 | +0.028891 | 1.6720 |
| src/compiler/utilities.ts:5195:1 | getEffectiveBaseTypeNode | 0.058224 | 0.087093 | +0.028868 | 1.4958 |
| src/compiler/moduleNameResolver.ts:2228:1 | getEntrypointsFromPackageJsonInfo | 5.183087 | 5.211851 | +0.028764 | 1.0055 |
| src/compiler/types.ts:8430:1 | UnscopedEmitHelper | 0.039914 | 0.068602 | +0.028687 | 1.7187 |
| src/compiler/performance.ts:17:1 |  | 0.040569 | 0.069242 | +0.028673 | 1.7068 |
| src/compiler/hostErrors.ts:7:1 | errorMessage | 0.066405 | 0.095018 | +0.028614 | 1.4309 |
| src/compiler/utilities.ts:9608:1 |  | 0.066186 | 0.094676 | +0.028490 | 1.4305 |
| src/compiler/factory/emitHelpers.ts:993:1 |  | 0.073310 | 0.101650 | +0.028340 | 1.3866 |
| src/compiler/builderPublic.ts:20:1 | AffectedFileResult | 0.044058 | 0.072380 | +0.028322 | 1.6428 |
| src/compiler/types.ts:3902:1 | JSDocType | 0.042703 | 0.070906 | +0.028202 | 1.6604 |
| src/compiler/utilitiesPublic.ts:1532:1 | isImportOrExportSpecifier | 0.301165 | 0.329335 | +0.028170 | 1.0935 |
| src/compiler/transformers/generators.ts:236:1 | OperationArguments | 0.039057 | 0.067182 | +0.028125 | 1.7201 |
| src/compiler/types.ts:3890:1 | JSDocNameReference | 0.042397 | 0.070450 | +0.028053 | 1.6617 |
| src/compiler/factory/emitHelpers.ts:1414:1 |  | 0.079624 | 0.107492 | +0.027868 | 1.3500 |
| src/compiler/utilities.ts:7587:1 | isInstanceOfExpression | 0.289554 | 0.317235 | +0.027681 | 1.0956 |
| src/compiler/utilitiesPublic.ts:2552:1 | hasJSDocNodes | 0.142377 | 0.169953 | +0.027575 | 1.1937 |
| src/compiler/utilities.ts:4321:1 | isFunctionSymbol | 0.285727 | 0.313240 | +0.027513 | 1.0963 |
| src/compiler/parser.ts:1:1 |  | 0.051440 | 0.078671 | +0.027230 | 1.5294 |
| src/compiler/factory/nodeTests.ts:1078:1 | isJSDocNamepathType | 0.047161 | 0.074388 | +0.027227 | 1.5773 |
| src/compiler/types.ts:3896:1 | JSDocMemberName | 0.043311 | 0.070413 | +0.027102 | 1.6258 |
| src/compiler/factory/nodeTests.ts:894:1 | isNamedExports | 0.044842 | 0.071554 | +0.026711 | 1.5957 |
| src/compiler/emitter.ts:1206:1 |  | 0.039049 | 0.065736 | +0.026687 | 1.6834 |
| src/compiler/types.ts:2339:1 | IndexedAccessTypeNode | 0.043354 | 0.070013 | +0.026659 | 1.6149 |
| src/compiler/utilities.ts:4499:1 | getSourceOfDefaultedAssignment | 0.251445 | 0.277877 | +0.026433 | 1.1051 |
| src/compiler/factory/utilities.ts:228:1 | createJsxFactoryExpression | 0.089521 | 0.115544 | +0.026023 | 1.2907 |
| src/compiler/utilities.ts:7707:1 | getStringFromExpandedCharCodes | 0.042709 | 0.068177 | +0.025468 | 1.5963 |
| src/compiler/transformers/jsx.ts:683:1 |  | 0.065550 | 0.090757 | +0.025207 | 1.3846 |
| src/compiler/types.ts:8543:1 | ExternalEmitHelpers | 0.040954 | 0.066005 | +0.025051 | 1.6117 |
| src/compiler/transformers/generators.ts:239:1 | BlockAction | 0.039329 | 0.064222 | +0.024892 | 1.6329 |
| src/compiler/types.ts:1665:1 | ModifierLike | 0.042555 | 0.067373 | +0.024818 | 1.5832 |
| src/compiler/commandLineParser.ts:1722:1 |  | 4.529026 | 4.553777 | +0.024751 | 1.0055 |
| src/compiler/utilities.ts:4780:1 | getJSDocOverloadTags | 0.308433 | 0.333178 | +0.024745 | 1.0802 |
| src/compiler/sys.ts:568:1 | sysLog | 0.040095 | 0.064428 | +0.024333 | 1.6069 |
| src/compiler/types.ts:1717:1 | AutoGenerateInfo | 0.049727 | 0.074032 | +0.024305 | 1.4888 |
| src/compiler/utilities.ts:7626:1 | tryExtractTSExtension | 0.042328 | 0.066560 | +0.024232 | 1.5725 |
| src/compiler/parser.ts:499:1 | isImportMeta | 0.243393 | 0.267601 | +0.024209 | 1.0995 |
| src/compiler/builder.ts:1668:1 | createBuilderProgram | 0.039773 | 0.063934 | +0.024161 | 1.6075 |
| src/compiler/moduleNameResolver.ts:2407:1 | getVersionPathsOfPackageJsonInfo | 4.588461 | 4.612584 | +0.024123 | 1.0053 |
| src/compiler/sys.ts:565:1 |  | 0.041755 | 0.065740 | +0.023985 | 1.5744 |
| src/compiler/executeCommandLine.ts:563:1 | executeCommandLineWorker | 0.815392 | 0.839354 | +0.023962 | 1.0294 |
| src/compiler/types.ts:3945:1 | JSDocTypeReferencingNode | 0.042318 | 0.066253 | +0.023935 | 1.5656 |
| src/compiler/moduleNameResolver.ts:3216:1 | mangleScopedPackageName | 0.053532 | 0.077400 | +0.023868 | 1.4459 |
| src/compiler/scanner.ts:552:1 | isWhiteSpaceLike | 0.040314 | 0.064072 | +0.023759 | 1.5894 |
| src/compiler/utilities.ts:7674:1 | convertToBase64 | 0.045419 | 0.069165 | +0.023746 | 1.5228 |
| src/compiler/types.ts:8437:1 | EmitHelper | 0.041768 | 0.065449 | +0.023681 | 1.5670 |
| src/compiler/factory/nodeTests.ts:418:1 | isPropertySignature | 0.070632 | 0.094144 | +0.023512 | 1.3329 |
| src/compiler/checker.ts:1364:1 | IntersectionState | 0.046247 | 0.069757 | +0.023509 | 1.5083 |
| src/compiler/types.ts:4262:1 | FlowReduceLabelData | 0.037756 | 0.061154 | +0.023397 | 1.6197 |
| src/compiler/utilities.ts:4522:1 | getSingleVariableOfVariableStatement | 0.230812 | 0.253998 | +0.023186 | 1.1005 |
| src/compiler/types.ts:2382:1 | TemplateLiteralTypeSpan | 0.039894 | 0.063063 | +0.023169 | 1.5808 |
| src/compiler/moduleNameResolver.ts:1334:1 | createModuleResolutionCache | 0.233812 | 0.256937 | +0.023125 | 1.0989 |
| src/compiler/utilities.ts:5271:1 | isPunctuation | 0.042935 | 0.066049 | +0.023114 | 1.5383 |
| src/compiler/builderPublic.ts:239:1 | createAbstractBuilder | 0.171706 | 0.194778 | +0.023072 | 1.1344 |
| src/compiler/transformers/declarations/diagnostics.ts:238:1 | createGetSymbolAccessibilityDiagnosticForNode | 0.157441 | 0.180334 | +0.022893 | 1.1454 |
| src/compiler/utilities.ts:785:1 | usingSingleLineStringWriter | 0.188506 | 0.211203 | +0.022698 | 1.1204 |
| src/compiler/emitter.ts:6366:1 | emitListItemWithParenthesizerRuleSelector | 0.086069 | 0.108620 | +0.022550 | 1.2620 |
| src/compiler/types.ts:3931:1 | JSDocFunctionType | 0.045074 | 0.067511 | +0.022436 | 1.4978 |
| src/compiler/types.ts:3976:1 | JSDocLinkPlain | 0.045304 | 0.067452 | +0.022148 | 1.4889 |
| src/compiler/factory/emitHelpers.ts:1009:1 |  | 0.071662 | 0.093546 | +0.021884 | 1.3054 |
| src/compiler/executeCommandLine.ts:941:1 | performIncrementalCompilation | 0.266406 | 0.288263 | +0.021857 | 1.0820 |
| src/compiler/utilities.ts:11452:1 | isConstAssertion | 0.304217 | 0.326074 | +0.021857 | 1.0718 |
| src/compiler/types.ts:8439:1 | EmitHelperUniqueNameCallback | 0.042989 | 0.064838 | +0.021849 | 1.5082 |
| src/compiler/factory/utilities.ts:1142:1 | canHaveIllegalModifiers | 0.044830 | 0.066606 | +0.021776 | 1.4857 |
| src/compiler/utilities.ts:12410:1 | suppressLeadingTrivia | 0.053996 | 0.075397 | +0.021402 | 1.3964 |
| src/compiler/types.ts:1711:1 | TransientIdentifier | 0.041811 | 0.063190 | +0.021379 | 1.5113 |
| src/compiler/utilities.ts:7500:1 | isEntityNameExpression | 0.301966 | 0.323029 | +0.021063 | 1.0698 |
| src/compiler/factory/nodeTests.ts:886:1 | isExportAssignment | 0.045548 | 0.066522 | +0.020975 | 1.4605 |
| src/compiler/utilities.ts:5266:1 | isKeyword | 0.051480 | 0.072448 | +0.020968 | 1.4073 |
| src/compiler/factory/emitHelpers.ts:1184:1 |  | 0.073402 | 0.094343 | +0.020941 | 1.2853 |
| src/compiler/executeCommandLine.ts:968:1 | updateSolutionBuilderHost | 0.080144 | 0.100980 | +0.020837 | 1.2600 |
| src/compiler/types.ts:4011:1 | JSDocDeprecatedTag | 0.042610 | 0.063356 | +0.020746 | 1.4869 |
| src/compiler/utilities.ts:4280:1 | getInitializerOfBinaryExpression | 0.314243 | 0.334919 | +0.020675 | 1.0658 |
| src/compiler/types.ts:3935:1 | JSDocVariadicType | 0.043069 | 0.063642 | +0.020573 | 1.4777 |
| src/compiler/resolutionCache.ts:532:1 | resolveModuleNameUsingGlobalCache | 0.124595 | 0.144969 | +0.020374 | 1.1635 |
| src/compiler/factory/emitHelpers.ts:1380:1 |  | 0.080013 | 0.100378 | +0.020365 | 1.2545 |
| src/compiler/types.ts:8425:1 | ScopedEmitHelper | 0.040946 | 0.061232 | +0.020285 | 1.4954 |
| src/compiler/utilities.ts:2078:1 | isAmbientModule | 0.245513 | 0.265737 | +0.020224 | 1.0824 |
| src/compiler/utilitiesPublic.ts:2117:1 | isIterationStatement | 0.142790 | 0.162973 | +0.020183 | 1.1414 |
| src/compiler/sys.ts:554:1 | getFileWatcherEventKind | 0.040632 | 0.060749 | +0.020117 | 1.4951 |
| src/compiler/types.ts:8442:1 | LanugageFeatures | 0.040205 | 0.060223 | +0.020018 | 1.4979 |
| src/compiler/utilities.ts:5035:1 | skipTypeParentheses | 0.282425 | 0.302399 | +0.019975 | 1.0707 |
| src/compiler/types.ts:4002:1 | JSDocImplementsTag | 0.059173 | 0.079105 | +0.019932 | 1.3368 |
| src/compiler/utilities.ts:2232:1 | isDeclarationWithTypeParameters | 0.114893 | 0.134789 | +0.019895 | 1.1732 |
| src/compiler/types.ts:1993:1 | PropertyAssignment | 0.038653 | 0.058495 | +0.019842 | 1.5133 |
| src/compiler/utilities.ts:2083:1 | isModuleWithStringLiteralName | 0.241768 | 0.261607 | +0.019839 | 1.0821 |
| src/compiler/scanner.ts:503:1 | getLineStarts | 0.037693 | 0.057435 | +0.019742 | 1.5237 |
| src/compiler/types.ts:1685:1 | GeneratedIdentifierFlags | 0.045837 | 0.065574 | +0.019737 | 1.4306 |
| src/compiler/scanner.ts:465:1 | getPositionOfLineAndCharacter | 0.111657 | 0.131371 | +0.019714 | 1.1766 |
| src/compiler/core.ts:1997:1 | compareTextSpans | 0.193587 | 0.213231 | +0.019643 | 1.1015 |
| src/compiler/factory/nodeTests.ts:840:1 | isImportDeclaration | 0.046071 | 0.065659 | +0.019588 | 1.4252 |
| src/compiler/types.ts:1672:1 | ParameterPropertyModifier | 0.040640 | 0.060226 | +0.019586 | 1.4819 |
| src/compiler/factory/nodeTests.ts:792:1 | isVariableDeclaration | 0.047990 | 0.067506 | +0.019516 | 1.4067 |
| src/compiler/moduleNameResolver.ts:1341:1 | createModuleResolutionCache | 0.229892 | 0.249362 | +0.019470 | 1.0847 |
| src/compiler/emitter.ts:563:1 | getOutputDeclarationFileName | 0.042138 | 0.061590 | +0.019452 | 1.4616 |
| src/compiler/core.ts:2257:1 | removeSuffix | 0.039475 | 0.058926 | +0.019451 | 1.4927 |
| src/compiler/factory/utilities.ts:865:1 | getInitializerOfBindingOrAssignmentElement | 0.261222 | 0.280591 | +0.019369 | 1.0741 |
| src/compiler/utilitiesPublic.ts:2597:1 | isTypeReferenceType | 0.045252 | 0.064612 | +0.019360 | 1.4278 |
| src/compiler/transformers/generators.ts:299:1 | WithBlock | 0.038357 | 0.057527 | +0.019169 | 1.4998 |
| src/compiler/types.ts:3885:1 | JSDocTypeExpression | 0.045543 | 0.064594 | +0.019051 | 1.4183 |
| src/compiler/utilities.ts:9296:1 |  | 0.096648 | 0.115639 | +0.018992 | 1.1965 |
| src/compiler/commandLineParser.ts:572:1 |  | 0.102533 | 0.121482 | +0.018949 | 1.1848 |
| src/compiler/utilities.ts:7568:1 | isRightSideOfQualifiedNameOrPropertyAccess | 0.125785 | 0.144679 | +0.018894 | 1.1502 |
| src/compiler/utilities.ts:992:1 | getSourceFileOfModule | 0.048966 | 0.067762 | +0.018796 | 1.3839 |
| src/compiler/utilitiesPublic.ts:1368:1 | isPropertyAccessChain | 0.242836 | 0.261551 | +0.018715 | 1.0771 |
| src/compiler/types.ts:1636:1 | DeclareKeyword | 0.042165 | 0.060750 | +0.018585 | 1.4408 |
| src/compiler/types.ts:2448:1 | LeftHandSideExpression | 0.041535 | 0.059978 | +0.018443 | 1.4440 |
| src/compiler/factory/emitHelpers.ts:1475:1 |  | 0.071797 | 0.090022 | +0.018225 | 1.2538 |
| src/compiler/visitorPublic.ts:607:1 | VisitEachChildFunction | 0.040562 | 0.058675 | +0.018114 | 1.4466 |
| src/compiler/utilitiesPublic.ts:1289:1 | getAllJSDocTagsOfKind | 0.051124 | 0.069191 | +0.018067 | 1.3534 |
| src/compiler/utilities.ts:5214:1 | getEffectiveImplementsTypeNodes | 0.069048 | 0.087103 | +0.018054 | 1.2615 |
| src/compiler/types.ts:4702:1 | LibResolution | 0.039051 | 0.057104 | +0.018053 | 1.4623 |
| src/compiler/scanner.ts:537:1 | getLinesBetweenPositions | 0.038965 | 0.056980 | +0.018015 | 1.4624 |
| src/compiler/utilitiesPublic.ts:2573:1 | hasInitializer | 0.044274 | 0.062281 | +0.018007 | 1.4067 |
| src/compiler/sys.ts:458:1 | createFixedChunkSizePollingWatchFile | 0.040813 | 0.058726 | +0.017913 | 1.4389 |
| src/compiler/scanner.ts:592:1 | isDigit | 0.037906 | 0.055785 | +0.017879 | 1.4717 |
| src/compiler/resolutionCache.ts:338:1 | canWatchAffectedPackageJsonOrNodeModulesOfAtTypes | 0.039834 | 0.057608 | +0.017775 | 1.4462 |
| src/compiler/types.ts:8677:1 | CallBinding | 0.039747 | 0.057328 | +0.017581 | 1.4423 |
| src/compiler/transformers/utilities.ts:762:1 | getAllDecoratorsOfMethod | 0.146759 | 0.164339 | +0.017580 | 1.1198 |
| src/compiler/types.ts:4958:1 | RedirectTargetsMap | 0.037492 | 0.055069 | +0.017577 | 1.4688 |
| src/compiler/utilitiesPublic.ts:1685:1 | isFunctionLikeDeclaration | 0.048138 | 0.065565 | +0.017426 | 1.3620 |
| src/compiler/transformers/es2021.ts:25:1 | transformES2021 | 0.079563 | 0.096986 | +0.017423 | 1.2190 |
| src/compiler/sys.ts:573:1 | setSysLog | 0.040619 | 0.057998 | +0.017379 | 1.4279 |
| src/compiler/transformers/utilities.ts:135:1 | chainBundle | 0.082899 | 0.100251 | +0.017352 | 1.2093 |
| src/compiler/factory/nodeConverters.ts:41:1 | createNodeConverters | 0.211228 | 0.228511 | +0.017283 | 1.0818 |
| src/compiler/utilitiesPublic.ts:2578:1 | hasOnlyExpressionInitializer | 0.045654 | 0.062868 | +0.017215 | 1.3771 |
| src/compiler/utilities.ts:5236:1 | getInterfaceBaseTypeNodes | 0.040076 | 0.057213 | +0.017138 | 1.4276 |
| src/compiler/transformers/utilities.ts:148:1 | getExportNeedsImportStarHelper | 0.040508 | 0.057610 | +0.017103 | 1.4222 |
| src/compiler/resolutionCache.ts:238:1 | DirPathToWatcherOfPackageDirWatcher | 0.038683 | 0.055699 | +0.017017 | 1.4399 |
| src/compiler/sys.ts:892:1 | FsWatchCallback | 0.039776 | 0.056775 | +0.016999 | 1.4274 |
| src/compiler/builderPublic.ts:150:1 | EmitAndSemanticDiagnosticsBuilderProgram | 0.043809 | 0.060797 | +0.016987 | 1.3878 |
| src/compiler/sourcemap.ts:408:1 | isRawSourceMap | 0.038545 | 0.055525 | +0.016981 | 1.4405 |
| src/compiler/types.ts:2131:1 | SemicolonClassElement | 0.042583 | 0.059553 | +0.016970 | 1.3985 |
| src/compiler/utilitiesPublic.ts:1819:1 | isFunctionOrConstructorTypeNode | 0.048829 | 0.065724 | +0.016895 | 1.3460 |
| src/compiler/utilities.ts:2354:1 | getNameFromIndexInfo | 0.040747 | 0.057641 | +0.016894 | 1.4146 |
| src/compiler/utilities.ts:2717:1 | isImportMeta | 0.268357 | 0.285222 | +0.016865 | 1.0628 |
| src/compiler/types.ts:4629:1 | ProjectReferenceFile | 0.037968 | 0.054778 | +0.016811 | 1.4428 |
| src/compiler/path.ts:1113:1 | isNodeModulesDirectory | 0.039571 | 0.056283 | +0.016712 | 1.4223 |
| src/compiler/core.ts:1066:1 |  | 0.038452 | 0.055132 | +0.016680 | 1.4338 |
| src/compiler/watch.ts:146:1 | clearScreenIfNotWatchingForFileChanges | 0.131112 | 0.147790 | +0.016678 | 1.1272 |
| src/compiler/utilities.ts:7001:1 | DetachedCommentInfo | 0.038821 | 0.055483 | +0.016663 | 1.4292 |
| src/compiler/utilities.ts:5178:1 | exportAssignmentIsAlias | 0.054391 | 0.071052 | +0.016660 | 1.3063 |
| src/compiler/utilities.ts:8040:1 | getCheckFlags | 0.130547 | 0.147196 | +0.016650 | 1.1275 |
| src/compiler/transformers/generators.ts:291:1 | LoopBlock | 0.037860 | 0.054418 | +0.016558 | 1.4373 |
| src/compiler/utilities.ts:5362:1 | isStringOrNumericLiteralLike | 0.369971 | 0.386440 | +0.016469 | 1.0445 |
| src/compiler/utilities.ts:12067:1 | hasInferredType | 0.184723 | 0.201167 | +0.016444 | 1.0890 |
| src/compiler/utilities.ts:7012:1 | emitDetachedComments | 0.039891 | 0.056312 | +0.016421 | 1.4117 |
| src/compiler/factory/emitHelpers.ts:1211:1 |  | 0.070074 | 0.086483 | +0.016409 | 1.2342 |
| src/compiler/executeCommandLine.ts:445:1 | generateSectionOptionsOutput | 0.084070 | 0.100470 | +0.016400 | 1.1951 |
| src/compiler/utilitiesPublic.ts:1294:1 | getTextOfJSDocComment | 0.039041 | 0.055394 | +0.016353 | 1.4189 |
| src/compiler/factory/emitHelpers.ts:1221:1 |  | 0.069892 | 0.086218 | +0.016325 | 1.2336 |
| src/compiler/factory/nodeTests.ts:890:1 | isExportDeclaration | 0.044164 | 0.060467 | +0.016302 | 1.3691 |
| src/compiler/types.ts:4618:1 | LibFile | 0.038497 | 0.054771 | +0.016274 | 1.4227 |
| src/compiler/utilitiesPublic.ts:1932:1 | isArrayBindingOrAssignmentElement | 0.191089 | 0.207352 | +0.016263 | 1.0851 |
| src/compiler/utilitiesPublic.ts:1204:1 | getJSDocTypeTag | 0.056461 | 0.072700 | +0.016239 | 1.2876 |
| src/compiler/moduleNameResolver.ts:2619:1 | loadModuleFromExports | 4.674832 | 4.691032 | +0.016199 | 1.0035 |
| src/compiler/sys.ts:577:1 | RecursiveDirectoryWatcherHost | 0.040174 | 0.056336 | +0.016162 | 1.4023 |
| src/compiler/sourcemap.ts:404:1 | isStringOrNull | 0.038785 | 0.054913 | +0.016128 | 1.4158 |
| src/compiler/factory/utilities.ts:1649:1 | createAccessorPropertySetRedirector | 0.244822 | 0.260883 | +0.016062 | 1.0656 |
| src/compiler/builder.ts:321:1 | createBuilderProgramState | 0.153156 | 0.169194 | +0.016038 | 1.1047 |
| src/compiler/types.ts:4706:1 | Program | 0.037997 | 0.054034 | +0.016038 | 1.4221 |
| src/compiler/path.ts:1068:1 | getRelativePathToDirectoryOrUrl | 0.042644 | 0.058623 | +0.015978 | 1.3747 |
| src/compiler/types.ts:4071:1 | JSDocTypedefTag | 0.039436 | 0.055368 | +0.015932 | 1.4040 |
| src/compiler/types.ts:4624:1 | ProjectReferenceFileKind | 0.038641 | 0.054503 | +0.015862 | 1.4105 |
| src/compiler/types.ts:1803:1 | PrivateIdentifier | 0.041556 | 0.057408 | +0.015852 | 1.3815 |
| src/compiler/types.ts:1646:1 | StaticKeyword | 0.043976 | 0.059755 | +0.015779 | 1.3588 |
| src/compiler/utilities.ts:7475:1 | isAssignmentExpression | 0.297708 | 0.313484 | +0.015776 | 1.0530 |
| src/compiler/utilities.ts:8522:1 | Token | 0.046338 | 0.062088 | +0.015750 | 1.3399 |
| src/compiler/utilities.ts:10591:1 | isShorthandPropertyNameUseSite | 0.242180 | 0.257914 | +0.015734 | 1.0650 |
| src/compiler/transformers/utilities.ts:878:1 | rewriteModuleSpecifier | 0.274071 | 0.289790 | +0.015720 | 1.0574 |
| src/compiler/types.ts:1729:1 | QualifiedName | 0.042992 | 0.058646 | +0.015654 | 1.3641 |
| src/compiler/types.ts:4671:1 | FilePreprocessingLibReferenceDiagnostic | 0.038294 | 0.053926 | +0.015633 | 1.4082 |
| src/compiler/moduleSpecifiers.ts:1364:1 | processEnding | 0.040686 | 0.056276 | +0.015590 | 1.3832 |
| src/compiler/transformers/generators.ts:265:1 | ExceptionBlock | 0.038611 | 0.054173 | +0.015562 | 1.4030 |
| src/compiler/transformers/classThis.ts:72:1 | isClassThisAssignmentBlock | 0.246687 | 0.262247 | +0.015560 | 1.0631 |
| src/compiler/resolutionCache.ts:1:1 |  | 0.039761 | 0.055273 | +0.015512 | 1.3901 |
| src/compiler/types.ts:5726:1 | TypePredicate | 0.038141 | 0.053572 | +0.015431 | 1.4046 |
| src/compiler/utilitiesPublic.ts:1550:1 | isTypeOnlyExportDeclaration | 0.184278 | 0.199687 | +0.015409 | 1.0836 |
| src/compiler/factory/nodeTests.ts:1062:1 | isJSDocNonNullableType | 0.046181 | 0.061564 | +0.015383 | 1.3331 |
| src/compiler/visitorPublic.ts:519:18 | visitFunctionBody | 0.040891 | 0.056271 | +0.015380 | 1.3761 |
| src/compiler/utilitiesPublic.ts:1832:1 | isBindingPattern | 0.046044 | 0.061408 | +0.015364 | 1.3337 |
| src/compiler/factory/nodeTests.ts:710:1 | isSemicolonClassElement | 0.042152 | 0.057501 | +0.015349 | 1.3641 |
| src/compiler/factory/nodeTests.ts:902:1 | isModuleExportName | 0.047365 | 0.062699 | +0.015334 | 1.3237 |
| src/compiler/utilities.ts:5079:1 | getDeclarationFromName | 0.285434 | 0.300753 | +0.015319 | 1.0537 |
| src/compiler/utilities.ts:2416:1 | createDiagnosticForNode | 0.053947 | 0.069245 | +0.015297 | 1.2836 |
| src/compiler/path.ts:146:1 | isVolumeCharacter | 0.039164 | 0.054412 | +0.015248 | 1.3893 |
| src/compiler/types.ts:1737:1 | PropertyName | 0.042531 | 0.057764 | +0.015233 | 1.3582 |
| src/compiler/transformer.ts:202:1 | wrapCustomTransformer | 0.182273 | 0.197501 | +0.015228 | 1.0835 |
| src/compiler/types.ts:7189:1 | WideningContext | 0.040595 | 0.055788 | +0.015193 | 1.3743 |
| src/compiler/types.ts:3940:1 | JSDocNamepathType | 0.042794 | 0.057984 | +0.015189 | 1.3549 |
| src/compiler/types.ts:3906:1 | JSDocAllType | 0.044762 | 0.059948 | +0.015186 | 1.3393 |
| src/compiler/types.ts:5729:1 | AnyImportSyntax | 0.038896 | 0.054020 | +0.015124 | 1.3888 |
| src/compiler/visitorPublic.ts:623:1 |  | 0.070188 | 0.085303 | +0.015115 | 1.2153 |
| src/compiler/types.ts:2085:1 | FunctionDeclaration | 0.043947 | 0.059050 | +0.015103 | 1.3437 |
| src/compiler/emitter.ts:580:1 | getOutputJSFileName | 0.041389 | 0.056440 | +0.015051 | 1.3636 |
| src/compiler/types.ts:4954:1 | Program | 0.037476 | 0.052501 | +0.015025 | 1.4009 |
| src/compiler/types.ts:1667:1 | AccessibilityModifier | 0.040872 | 0.055804 | +0.014932 | 1.3653 |
| src/compiler/moduleNameResolver.ts:1113:1 | ModeAwareCacheKey | 0.039044 | 0.053955 | +0.014911 | 1.3819 |
| src/compiler/factory/nodeTests.ts:800:1 | isFunctionDeclaration | 0.045931 | 0.060814 | +0.014884 | 1.3240 |
| src/compiler/watch.ts:125:1 | createDiagnosticReporter | 0.133795 | 0.148667 | +0.014873 | 1.1112 |
| src/compiler/resolutionCache.ts:510:1 | createModuleResolutionLoaderUsingGlobalCache | 0.042058 | 0.056919 | +0.014860 | 1.3533 |
| src/compiler/types.ts:1645:1 | OverrideKeyword | 0.042881 | 0.057738 | +0.014857 | 1.3465 |
| src/compiler/types.ts:3018:1 | SuperElementAccessExpression | 0.041830 | 0.056672 | +0.014842 | 1.3548 |
| src/compiler/watch.ts:177:1 | getLocaleTimeString | 0.087461 | 0.102254 | +0.014793 | 1.1691 |
| src/compiler/utilities.ts:5184:1 | getExportAssignmentExpression | 0.192180 | 0.206946 | +0.014767 | 1.0768 |
| src/compiler/utilitiesPublic.ts:2534:1 | isJSDocTag | 0.045118 | 0.059861 | +0.014743 | 1.3268 |
| src/compiler/visitorPublic.ts:540:1 | visitIterationBody | 0.053607 | 0.068256 | +0.014649 | 1.2733 |
| src/compiler/program.ts:711:1 | formatCodeSpan | 0.317497 | 0.332011 | +0.014514 | 1.0457 |
| src/compiler/path.ts:1047:1 | getRelativePathFromDirectory | 0.041204 | 0.055644 | +0.014440 | 1.3505 |
| src/compiler/utilitiesPublic.ts:2542:1 | isGetAccessor | 0.044106 | 0.058535 | +0.014429 | 1.3271 |
| src/compiler/emitter.ts:6362:1 | emitListItemNoParenthesizer | 0.043692 | 0.058078 | +0.014386 | 1.3293 |
| src/compiler/types.ts:4642:1 | ReferencedFile | 0.039668 | 0.054040 | +0.014372 | 1.3623 |
| src/compiler/utilitiesPublic.ts:1957:1 | isPropertyAccessOrQualifiedName | 0.049694 | 0.064041 | +0.014347 | 1.2887 |
| src/compiler/types.ts:7310:1 | DiagnosticWithLocation | 0.040236 | 0.054576 | +0.014340 | 1.3564 |
| src/compiler/types.ts:3964:1 | JSDocLink | 0.042929 | 0.057243 | +0.014314 | 1.3334 |
| src/compiler/transformers/esDecorators.ts:281:1 | OtherLexicalEnvironmentStackEntry | 0.041244 | 0.055546 | +0.014301 | 1.3467 |
| src/compiler/utilities.ts:2088:1 | isNonGlobalAmbientModule | 0.251624 | 0.265889 | +0.014264 | 1.0567 |
| src/compiler/factory/nodeTests.ts:403:1 | isTypeParameterDeclaration | 0.050135 | 0.064371 | +0.014236 | 1.2840 |
| src/compiler/transformers/classFields.ts:3349:1 | isPrivateIdentifierInExpression | 0.273013 | 0.287232 | +0.014219 | 1.0521 |
| src/compiler/types.ts:8499:1 |  | 0.073245 | 0.087462 | +0.014218 | 1.1941 |
| src/compiler/types.ts:6852:1 | PromiseOrAwaitableType | 0.041740 | 0.055914 | +0.014174 | 1.3396 |
| src/compiler/utilities.ts:7403:1 | isBinaryLogicalOperator | 0.044816 | 0.058944 | +0.014129 | 1.3153 |
| src/compiler/visitorPublic.ts:619:1 | VisitEachChildTable | 0.042872 | 0.056927 | +0.014055 | 1.3278 |
| src/compiler/factory/nodeFactory.ts:7305:1 | propagateChildrenFlags | 0.040397 | 0.054411 | +0.014014 | 1.3469 |
| src/compiler/utilities.ts:8420:1 | getLeftmostExpression | 0.172593 | 0.186604 | +0.014012 | 1.0812 |
| src/compiler/factory/nodeTests.ts:832:1 | isNamespaceExportDeclaration | 0.046988 | 0.060994 | +0.014006 | 1.2981 |
| src/compiler/emitter.ts:6360:1 | EmitListItemFunction | 0.038989 | 0.052968 | +0.013979 | 1.3585 |
| src/compiler/tracing.ts:374:1 | TracingNode | 0.039394 | 0.053368 | +0.013974 | 1.3547 |
| src/compiler/types.ts:3984:1 | JSDocText | 0.041472 | 0.055429 | +0.013957 | 1.3365 |
| src/compiler/factory/nodeTests.ts:764:1 | isReturnStatement | 0.045471 | 0.059427 | +0.013956 | 1.3069 |
| src/compiler/utilities.ts:5644:1 | getOperatorAssociativity | 0.046567 | 0.060440 | +0.013874 | 1.2979 |
| src/compiler/utilities.ts:3551:1 | childIsDecorated | 0.159840 | 0.173708 | +0.013868 | 1.0868 |
| src/compiler/utilities.ts:8600:1 | maybeSetLocalizedDiagnosticMessages | 0.039639 | 0.053460 | +0.013821 | 1.3487 |
| src/compiler/utilities.ts:6488:1 | hostGetCanonicalFileName | 0.040524 | 0.054259 | +0.013735 | 1.3389 |
| src/compiler/builder.ts:2238:1 | toBuilderFileEmit | 0.096406 | 0.110130 | +0.013724 | 1.1424 |
| src/compiler/types.ts:8406:1 | InternalEmitFlags | 0.046423 | 0.060147 | +0.013724 | 1.2956 |
| src/compiler/types.ts:4580:1 | WriteFileCallback | 0.040951 | 0.054628 | +0.013678 | 1.3340 |
| src/compiler/scanner.ts:641:1 | skipTrivia | 0.042216 | 0.055825 | +0.013609 | 1.3224 |
| src/compiler/watch.ts:912:1 | CreateWatchCompilerHostOfConfigFileInput | 0.042578 | 0.056174 | +0.013596 | 1.3193 |
| src/compiler/factory/utilities.ts:599:1 | startsWithUseStrict | 0.292427 | 0.305998 | +0.013571 | 1.0464 |
| src/compiler/moduleNameResolver.ts:1367:1 | createTypeReferenceDirectiveResolutionCache | 0.235673 | 0.249212 | +0.013538 | 1.0574 |
| src/compiler/sys.ts:242:1 | WatchedFileWithUnchangedPolls | 0.045584 | 0.059091 | +0.013507 | 1.2963 |
| src/compiler/types.ts:2356:1 | LiteralTypeNode | 0.042275 | 0.055770 | +0.013495 | 1.3192 |
| src/compiler/types.ts:1643:1 | ReadonlyKeyword | 0.041534 | 0.054990 | +0.013456 | 1.3240 |
| src/compiler/types.ts:2478:1 | SuperExpression | 0.040287 | 0.053718 | +0.013430 | 1.3334 |
| src/compiler/program.ts:692:1 |  | 0.039324 | 0.052722 | +0.013399 | 1.3407 |
| src/compiler/utilities.ts:11165:1 | isTypeUsableAsPropertyName | 0.046797 | 0.060196 | +0.013398 | 1.2863 |
| src/compiler/path.ts:151:1 | getFileUrlVolumeSeparatorEnd | 0.039274 | 0.052671 | +0.013396 | 1.3411 |
| src/compiler/types.ts:1644:1 | OutKeyword | 0.042162 | 0.055545 | +0.013384 | 1.3174 |
| src/compiler/utilitiesPublic.ts:1690:1 | isBooleanLiteral | 0.048440 | 0.061818 | +0.013378 | 1.2762 |
| src/compiler/builder.ts:567:1 | convertToDiagnostics | 0.152688 | 0.166045 | +0.013356 | 1.0875 |
| src/compiler/factory/nodeTests.ts:836:1 | isImportEqualsDeclaration | 0.048689 | 0.062031 | +0.013342 | 1.2740 |
| src/compiler/transformers/es2015.ts:272:1 | LoopOutParameterFlags | 0.040723 | 0.054038 | +0.013315 | 1.3270 |
| src/compiler/builder.ts:1095:1 | IncrementalBuildInfoReferencedMap | 0.041955 | 0.055252 | +0.013296 | 1.3169 |
| src/compiler/types.ts:4480:1 | SourceFile | 0.038728 | 0.051997 | +0.013269 | 1.3426 |
| src/compiler/program.ts:691:1 |  | 0.041247 | 0.054478 | +0.013231 | 1.3208 |
| src/compiler/types.ts:4612:1 | RootFile | 0.037551 | 0.050751 | +0.013200 | 1.3515 |
| src/compiler/commandLineParser.ts:2337:1 | getCommandLineTypeAcquisitionMap | 0.041662 | 0.054855 | +0.013193 | 1.3167 |
| src/compiler/utilities.ts:5041:1 | skipParentheses | 0.040253 | 0.053402 | +0.013149 | 1.3267 |
| src/compiler/utilities.ts:4935:1 | isInCompoundLikeAssignment | 0.168335 | 0.181477 | +0.013142 | 1.0781 |
| src/compiler/tracing.ts:371:1 |  | 0.040658 | 0.053777 | +0.013119 | 1.3227 |
| src/compiler/types.ts:8417:1 | EmitHelperBase | 0.041636 | 0.054738 | +0.013102 | 1.3147 |
| src/compiler/semver.ts:443:1 | formatDisjunction | 0.040710 | 0.053759 | +0.013049 | 1.3205 |
| src/compiler/builder.ts:1658:1 | computeSignature | 0.079413 | 0.092459 | +0.013046 | 1.1643 |
| src/compiler/visitorPublic.ts:521:18 | visitFunctionBody | 0.042244 | 0.055264 | +0.013020 | 1.3082 |
| src/compiler/utilities.ts:8644:1 | attachFileToDiagnostic | 0.040634 | 0.053652 | +0.013018 | 1.3204 |
| src/compiler/factory/nodeTests.ts:716:1 | isBlock | 0.044927 | 0.057922 | +0.012995 | 1.2892 |
| src/compiler/program.ts:766:1 | formatLocation | 0.191694 | 0.204678 | +0.012984 | 1.0677 |
| src/compiler/types.ts:4118:1 | JSDocParameterTag | 0.045233 | 0.058186 | +0.012953 | 1.2864 |
| src/compiler/transformers/declarations/diagnostics.ts:164:1 | createGetSymbolAccessibilityDiagnosticForNodeName | 0.132446 | 0.145360 | +0.012914 | 1.0975 |
| src/compiler/sys.ts:900:1 | FsWatchWorker | 0.038795 | 0.051693 | +0.012897 | 1.3324 |
| src/compiler/types.ts:3958:1 | JSDocTag | 0.045506 | 0.058379 | +0.012873 | 1.2829 |
| src/compiler/scanner.ts:468:1 | getPositionOfLineAndCharacter | 0.111859 | 0.124731 | +0.012872 | 1.1151 |
| src/compiler/types.ts:3970:1 | JSDocLinkCode | 0.044934 | 0.057780 | +0.012846 | 1.2859 |
| src/compiler/sourcemap.ts:1:1 |  | 0.040140 | 0.052939 | +0.012799 | 1.3189 |
| src/compiler/utilities.ts:6160:1 |  | 0.041517 | 0.054305 | +0.012788 | 1.3080 |
| src/compiler/factory/nodeTests.ts:853:1 | isAssertClause | 0.044948 | 0.057725 | +0.012777 | 1.2843 |
| src/compiler/performance.ts:63:1 |  | 0.038571 | 0.051326 | +0.012755 | 1.3307 |
| src/compiler/utilities.ts:12155:1 |  | 0.042127 | 0.054878 | +0.012750 | 1.3027 |
| src/compiler/types.ts:4241:1 | FlowSwitchClauseData | 0.038914 | 0.051634 | +0.012720 | 1.3269 |
| src/compiler/transformers/utilities.ts:471:1 | isSimpleCopiableExpression | 0.275105 | 0.287802 | +0.012697 | 1.0462 |
| src/compiler/factory/emitHelpers.ts:879:1 |  | 0.066443 | 0.079127 | +0.012684 | 1.1909 |
| src/compiler/utilities.ts:4701:1 | ownsJSDocTag | 0.274421 | 0.287085 | +0.012664 | 1.0461 |
| src/compiler/factory/nodeTests.ts:1066:1 | isJSDocOptionalType | 0.048194 | 0.060854 | +0.012660 | 1.2627 |
| src/compiler/utilities.ts:6412:1 | getTrailingSemicolonDeferringWriter | 0.602790 | 0.615383 | +0.012593 | 1.0209 |
| src/compiler/utilities.ts:4228:1 | getElementOrPropertyAccessName | 0.200164 | 0.212695 | +0.012531 | 1.0626 |
| src/compiler/types.ts:1783:1 | LateBoundBinaryExpressionDeclaration | 0.045167 | 0.057693 | +0.012525 | 1.2773 |
| src/compiler/_namespaces/ts.ts:26:1 |  | 0.040386 | 0.052896 | +0.012510 | 1.3098 |
| src/compiler/utilities.ts:1980:1 | getLiteralText | 0.182940 | 0.195450 | +0.012510 | 1.0684 |
| src/compiler/visitorPublic.ts:520:18 | visitFunctionBody | 0.043213 | 0.055681 | +0.012468 | 1.2885 |
| src/compiler/executeCommandLine.ts:992:1 | updateWatchCompilationHost | 0.079331 | 0.091731 | +0.012400 | 1.1563 |
| src/compiler/utilities.ts:7581:1 | isRightSideOfQualifiedNameOrPropertyAccessOrJSDocMemberName | 0.250790 | 0.263141 | +0.012351 | 1.0492 |
| src/compiler/utilities.ts:12097:1 |  | 0.041253 | 0.053599 | +0.012346 | 1.2993 |
| src/compiler/types.ts:1633:1 | AccessorKeyword | 0.041264 | 0.053595 | +0.012330 | 1.2988 |
| src/compiler/utilitiesPublic.ts:2538:1 | isSetAccessor | 0.044184 | 0.056479 | +0.012294 | 1.2783 |
| src/compiler/utilities.ts:6504:1 | getCanonicalAbsolutePath | 0.121468 | 0.133738 | +0.012270 | 1.1010 |
| src/compiler/commandLineParser.ts:2333:1 | getCommandLineWatchOptionsMap | 0.040969 | 0.053223 | +0.012254 | 1.2991 |
| src/compiler/utilities.ts:5281:1 | isContextualKeyword | 0.043921 | 0.056064 | +0.012143 | 1.2765 |
| src/compiler/visitorPublic.ts:474:1 | addDefaultValueAssignmentForInitializer | 0.226695 | 0.238804 | +0.012109 | 1.0534 |
| src/compiler/factory/nodeTests.ts:748:1 | isForInStatement | 0.043902 | 0.056006 | +0.012104 | 1.2757 |
| src/compiler/types.ts:2603:1 | LogicalOperatorOrHigher | 0.041600 | 0.053652 | +0.012052 | 1.2897 |
| src/compiler/scanner.ts:475:1 | computePositionOfLineAndCharacter | 0.039849 | 0.051892 | +0.012044 | 1.3022 |
| src/compiler/types.ts:1746:1 | MemberName | 0.049644 | 0.061684 | +0.012040 | 1.2425 |
| src/compiler/transformers/destructuring.ts:446:1 | flattenArrayBindingOrAssignmentPattern | 0.079641 | 0.091680 | +0.012039 | 1.1512 |
| src/compiler/types.ts:3023:1 | SuperProperty | 0.043059 | 0.055087 | +0.012028 | 1.2793 |
| src/compiler/types.ts:6039:1 | SymbolId | 0.039539 | 0.051544 | +0.012005 | 1.3036 |
| src/compiler/factory/nodeTests.ts:848:1 | isImportTypeAssertionContainer | 0.047231 | 0.059208 | +0.011977 | 1.2536 |
| src/compiler/program.ts:690:1 |  | 0.039013 | 0.050909 | +0.011896 | 1.3049 |
| src/compiler/utilities.ts:5073:1 | isDeclarationName | 0.228178 | 0.240046 | +0.011868 | 1.0520 |
| src/compiler/utilitiesPublic.ts:1299:1 | formatJSDocLink | 0.040889 | 0.052749 | +0.011859 | 1.2900 |
| src/compiler/commandLineParser.ts:3865:1 | convertJsonOptionOfCustomType | 0.051422 | 0.063280 | +0.011858 | 1.2306 |
| src/compiler/factory/utilities.ts:247:1 | createExpressionForJsxElement | 0.141316 | 0.153148 | +0.011832 | 1.0837 |
| src/compiler/utilities.ts:8045:1 | getDeclarationModifierFlagsFromSymbol | 0.042197 | 0.054029 | +0.011831 | 1.2804 |
| src/compiler/utilities.ts:6661:1 | sourceFileMayBeEmitted | 0.361880 | 0.373674 | +0.011794 | 1.0326 |
| src/compiler/types.ts:3982:1 | JSDocComment | 0.043209 | 0.054950 | +0.011742 | 1.2717 |
| src/compiler/types.ts:1682:1 | ModifiersArray | 0.052495 | 0.064220 | +0.011725 | 1.2234 |
| src/compiler/utilities.ts:7185:1 | hasSyntacticModifiers | 0.052132 | 0.063854 | +0.011722 | 1.2249 |
| src/compiler/utilities.ts:8534:1 | Identifier | 0.046553 | 0.058229 | +0.011676 | 1.2508 |
| src/compiler/utilities.ts:3553:1 | childIsDecorated | 0.157112 | 0.168784 | +0.011671 | 1.0743 |
| src/compiler/utilities.ts:8740:1 | createCompilerDiagnosticFromMessageChain | 0.042910 | 0.054561 | +0.011651 | 1.2715 |
| src/compiler/factory/nodeTests.ts:397:1 | isComputedPropertyName | 0.045852 | 0.057478 | +0.011626 | 1.2536 |
| src/compiler/types.ts:1796:1 | ComputedPropertyName | 0.042099 | 0.053724 | +0.011625 | 1.2761 |
| src/compiler/types.ts:1642:1 | PublicKeyword | 0.043852 | 0.055476 | +0.011624 | 1.2651 |
| src/compiler/checker.ts:1402:1 | UnusedKind | 0.041623 | 0.053238 | +0.011615 | 1.2791 |
| src/compiler/factory/emitHelpers.ts:1344:1 |  | 0.081534 | 0.093146 | +0.011612 | 1.1424 |
| src/compiler/path.ts:32:1 |  | 0.039584 | 0.051190 | +0.011606 | 1.2932 |
| src/compiler/factory/nodeTests.ts:776:1 | isLabeledStatement | 0.045395 | 0.056980 | +0.011585 | 1.2552 |
| src/compiler/factory/nodeTests.ts:382:1 | isImportKeyword | 0.046124 | 0.057704 | +0.011581 | 1.2511 |
| src/compiler/types.ts:1792:1 | DeclarationStatement | 0.045103 | 0.056671 | +0.011568 | 1.2565 |
| src/compiler/scanner.ts:548:1 | getLineAndCharacterOfPosition | 0.038097 | 0.049638 | +0.011542 | 1.3030 |
| src/compiler/utilities.ts:12091:1 | isSideEffectImport | 0.046294 | 0.057771 | +0.011476 | 1.2479 |
| src/compiler/utilitiesPublic.ts:1989:1 | isTemplateLiteral | 0.050053 | 0.061497 | +0.011444 | 1.2286 |
| src/compiler/utilities.ts:10284:1 | resolutionExtensionIsTSOrJson | 0.039931 | 0.051373 | +0.011442 | 1.2866 |
| src/compiler/factory/nodeTests.ts:352:1 | isReadonlyKeyword | 0.046396 | 0.057820 | +0.011424 | 1.2462 |
| src/compiler/resolutionCache.ts:229:1 | DirectoryWatchesOfFailedLookup | 0.039900 | 0.051310 | +0.011410 | 1.2860 |
| src/compiler/watchUtilities.ts:557:1 | IsIgnoredFileFromWildCardWatchingInput | 0.046323 | 0.057711 | +0.011388 | 1.2458 |
| src/compiler/resolutionCache.ts:82:1 | HasInvalidatedFromResolutionCache | 0.043947 | 0.055332 | +0.011386 | 1.2591 |
| src/compiler/utilities.ts:3283:1 | isInTopLevelContext | 0.249995 | 0.261364 | +0.011369 | 1.0455 |
| src/compiler/utilities.ts:2519:1 | scanTokenAtPosition | 0.119569 | 0.130900 | +0.011332 | 1.0948 |
| src/compiler/builder.ts:509:1 | getEmitSignatureFromOldSignature | 0.067067 | 0.078365 | +0.011299 | 1.1685 |
| src/compiler/parser.ts:1321:1 | CreateSourceFileOptions | 0.040689 | 0.051985 | +0.011295 | 1.2776 |
| src/compiler/utilities.ts:1351:1 | indexOfNode | 0.094024 | 0.105306 | +0.011283 | 1.1200 |
| src/compiler/types.ts:1735:1 | EntityName | 0.042966 | 0.054207 | +0.011240 | 1.2616 |
| src/compiler/factory/emitHelpers.ts:895:1 |  | 0.071623 | 0.082847 | +0.011224 | 1.1567 |
| src/compiler/utilities.ts:10294:1 | extensionFromPath | 0.040531 | 0.051750 | +0.011220 | 1.2768 |
| src/compiler/diagnosticInformationMap.generated.ts:4:1 |  | 0.043182 | 0.054274 | +0.011092 | 1.2569 |
| src/compiler/types.ts:3014:1 | ElementAccessChainRoot | 0.040782 | 0.051868 | +0.011086 | 1.2718 |
| src/compiler/types.ts:2373:1 | StringLiteralLike | 0.043568 | 0.054617 | +0.011050 | 1.2536 |
| src/compiler/scanner.ts:467:1 | getPositionOfLineAndCharacter | 0.114109 | 0.125158 | +0.011049 | 1.0968 |
| src/compiler/types.ts:1648:1 | Modifier | 0.044349 | 0.055397 | +0.011047 | 1.2491 |
| src/compiler/commandLineParser.ts:2336:1 |  | 0.042723 | 0.053742 | +0.011019 | 1.2579 |
| src/compiler/factory/nodeTests.ts:732:1 | isIfStatement | 0.042489 | 0.053498 | +0.011009 | 1.2591 |
| src/compiler/utilities.ts:5208:1 | getClassExtendsHeritageElement | 0.055183 | 0.066188 | +0.011005 | 1.1994 |
| src/compiler/types.ts:7317:1 | DiagnosticWithDetachedLocation | 0.042324 | 0.053327 | +0.011003 | 1.2600 |
| src/compiler/factory/nodeTests.ts:828:1 | isCaseBlock | 0.046319 | 0.057310 | +0.010991 | 1.2373 |
| src/compiler/utilities.ts:4767:1 | getEffectiveContainerForJSDocTemplateTag | 0.288422 | 0.299376 | +0.010954 | 1.0380 |
| src/compiler/utilities.ts:2511:1 | getSpanOfTokenAtPosition | 0.162050 | 0.172954 | +0.010904 | 1.0673 |
| src/compiler/checker.ts:1397:1 | MembersOrExportsResolutionKind | 0.041868 | 0.052745 | +0.010877 | 1.2598 |
| src/compiler/transformers/generators.ts:284:1 | SwitchBlock | 0.038037 | 0.048892 | +0.010855 | 1.2854 |
| src/compiler/types.ts:2918:1 | TemplateSpan | 0.038792 | 0.049600 | +0.010807 | 1.2786 |
| src/compiler/types.ts:3920:1 | JSDocNullableType | 0.044604 | 0.055401 | +0.010797 | 1.2421 |
| src/compiler/commandLineParser.ts:3353:1 | isSolutionConfig | 0.042376 | 0.053150 | +0.010773 | 1.2542 |
| src/compiler/utilities.ts:2706:1 | isImportCall | 0.168220 | 0.178962 | +0.010742 | 1.0639 |
| src/compiler/transformers/utilities.ts:880:1 | rewriteModuleSpecifier | 0.280315 | 0.291055 | +0.010741 | 1.0383 |
| src/compiler/factory/utilities.ts:183:1 | createEmptyExports | 0.140953 | 0.151670 | +0.010718 | 1.0760 |
| src/compiler/types.ts:3926:1 | JSDocOptionalType | 0.044665 | 0.055360 | +0.010695 | 1.2395 |
| src/compiler/commandLineParser.ts:1716:1 |  | 0.039387 | 0.050063 | +0.010676 | 1.2711 |
| src/compiler/types.ts:3951:1 | JSDoc | 0.045277 | 0.055951 | +0.010674 | 1.2357 |
| src/compiler/utilitiesPublic.ts:1710:1 | isFunctionLikeKind | 0.045353 | 0.056009 | +0.010655 | 1.2349 |
| src/compiler/transformer.ts:222:1 | wrapDeclarationTransformerFactory | 0.040153 | 0.050787 | +0.010634 | 1.2648 |
| src/compiler/types.ts:8111:1 | ResolvedTypeReferenceDirective | 0.040079 | 0.050652 | +0.010573 | 1.2638 |
| src/compiler/builder.ts:1093:1 | IncrementalBuildInfoFilePendingEmit | 0.041379 | 0.051948 | +0.010569 | 1.2554 |
| src/compiler/factory/nodeTests.ts:357:1 | isStaticModifier | 0.047321 | 0.057882 | +0.010561 | 1.2232 |
| src/compiler/utilities.ts:4491:1 | getSourceOfAssignment | 0.245824 | 0.256376 | +0.010552 | 1.0429 |
| src/compiler/commandLineParser.ts:4105:1 | specToDiagnostic | 0.043496 | 0.054033 | +0.010537 | 1.2423 |
| src/compiler/types.ts:2452:1 | MemberExpression | 0.042762 | 0.053295 | +0.010533 | 1.2463 |
| src/compiler/types.ts:7301:1 | DiagnosticRelatedInformation | 0.040530 | 0.051056 | +0.010526 | 1.2597 |
| src/compiler/commandLineParser.ts:1707:1 |  | 4.813896 | 4.824409 | +0.010513 | 1.0022 |
| src/compiler/types.ts:4635:1 | ReferencedFileKind | 0.043304 | 0.053810 | +0.010506 | 1.2426 |
| src/compiler/transformers/utilities.ts:593:1 | isStaticPropertyDeclarationOrClassStaticBlockDeclaration | 0.239661 | 0.250140 | +0.010479 | 1.0437 |
| src/compiler/factory/nodeTests.ts:804:1 | isClassDeclaration | 0.048487 | 0.058964 | +0.010478 | 1.2161 |
| src/compiler/path.ts:1056:1 | convertToRelativePath | 0.043530 | 0.053996 | +0.010465 | 1.2404 |
| src/compiler/types.ts:1811:1 | GeneratedPrivateIdentifier | 0.041653 | 0.052025 | +0.010372 | 1.2490 |
| src/compiler/factory/nodeTests.ts:393:1 | isQualifiedName | 0.046239 | 0.056609 | +0.010370 | 1.2243 |
| src/compiler/types.ts:2987:1 | PropertyAccessChainRoot | 0.041389 | 0.051676 | +0.010287 | 1.2485 |
| src/compiler/utilities.ts:7455:1 | tryGetClassImplementingOrExtendingExpressionWithTypeArguments | 0.282837 | 0.293121 | +0.010284 | 1.0364 |
| src/compiler/transformers/namedEvaluation.ts:139:1 | isClassNamedEvaluationHelperBlock | 0.226778 | 0.237051 | +0.010273 | 1.0453 |
| src/compiler/factory/nodeTests.ts:796:1 | isVariableDeclarationList | 0.046166 | 0.056434 | +0.010267 | 1.2224 |
| src/compiler/utilities.ts:4376:1 | importFromModuleSpecifier | 0.037854 | 0.048108 | +0.010254 | 1.2709 |
| src/compiler/utilities.ts:4487:1 | isTypeAlias | 0.245997 | 0.256250 | +0.010253 | 1.0417 |
| src/compiler/resolutionCache.ts:176:1 | CachedResolvedModuleWithFailedLookupLocations | 0.040819 | 0.051058 | +0.010240 | 1.2509 |
| src/compiler/sourcemap.ts:422:1 | tryParseRawSourceMap | 0.039344 | 0.049583 | +0.010239 | 1.2603 |
| src/compiler/utilitiesPublic.ts:1536:1 | isTypeOnlyImportDeclaration | 0.192044 | 0.202259 | +0.010215 | 1.0532 |
| src/compiler/utilities.ts:10279:1 | extensionIsTS | 0.041353 | 0.051543 | +0.010190 | 1.2464 |
| src/compiler/types.ts:1778:1 | LateBoundDeclaration | 0.043836 | 0.054003 | +0.010167 | 1.2319 |
| src/compiler/utilities.ts:5173:1 | isAliasableExpression | 0.245380 | 0.255526 | +0.010146 | 1.0413 |
| src/compiler/factory/nodeTests.ts:760:1 | isBreakStatement | 0.045135 | 0.055279 | +0.010144 | 1.2247 |
| src/compiler/path.ts:1045:1 | getRelativePathFromDirectory | 0.040970 | 0.051105 | +0.010135 | 1.2474 |
| src/compiler/transformers/utilities.ts:630:1 | isInitializedProperty | 0.171573 | 0.181707 | +0.010134 | 1.0591 |
| src/compiler/factory/nodeTests.ts:808:1 | isInterfaceDeclaration | 0.046693 | 0.056821 | +0.010128 | 1.2169 |
| src/compiler/types.ts:1772:1 | DynamicNamedBinaryExpression | 0.045099 | 0.055224 | +0.010125 | 1.2245 |
| src/compiler/types.ts:1788:1 | LateBoundElementAccessExpression | 0.044040 | 0.054164 | +0.010125 | 1.2299 |
| src/compiler/factory/nodeTests.ts:816:1 | isEnumDeclaration | 0.047046 | 0.057148 | +0.010102 | 1.2147 |
| src/compiler/resolutionCache.ts:320:1 | canWatchDirectoryOrFilePath | 0.040405 | 0.050499 | +0.010094 | 1.2498 |
| src/compiler/utilities.ts:10903:1 | isInfinityOrNaNString | 0.038818 | 0.048893 | +0.010075 | 1.2595 |
| src/compiler/transformers/utilities.ts:781:1 | getAllDecoratorsOfProperty | 0.070303 | 0.080364 | +0.010061 | 1.1431 |
| src/compiler/watch.ts:161:1 |  | 0.045297 | 0.055345 | +0.010048 | 1.2218 |
| src/compiler/utilities.ts:5552:1 | NamedEvaluation | 0.038804 | 0.048838 | +0.010034 | 1.2586 |
| src/compiler/factory/nodeTests.ts:720:1 | isVariableStatement | 0.044478 | 0.054503 | +0.010025 | 1.2254 |
| src/compiler/types.ts:2288:1 | OptionalTypeNode | 0.039179 | 0.049203 | +0.010023 | 1.2558 |
| src/compiler/types.ts:1817:1 | LateBoundName | 0.044171 | 0.054187 | +0.010015 | 1.2267 |
| src/compiler/utilities.ts:4785:1 | getHostSignatureFromJSDoc | 0.270337 | 0.280348 | +0.010012 | 1.0370 |
| src/compiler/transformers/generators.ts:222:1 | OpCode | 0.039421 | 0.049423 | +0.010002 | 1.2537 |
| src/compiler/moduleNameResolver.ts:1051:1 | createPackageJsonInfoCache | 0.065422 | 0.075335 | +0.009912 | 1.1515 |
| src/compiler/utilities.ts:7147:1 | writeTrimmedCurrentLine | 0.172254 | 0.182138 | +0.009884 | 1.0574 |
| src/compiler/types.ts:4228:1 | FlowCondition | 0.039850 | 0.049679 | +0.009829 | 1.2466 |
| src/compiler/types.ts:4664:1 | FilePreprocessingDiagnosticsKind | 0.037775 | 0.047602 | +0.009827 | 1.2602 |
| src/compiler/types.ts:1639:1 | InKeyword | 0.043749 | 0.053572 | +0.009823 | 1.2245 |
| src/compiler/factory/nodeTests.ts:616:1 | isArrowFunction | 0.043216 | 0.053036 | +0.009819 | 1.2272 |
| src/compiler/utilitiesPublic.ts:1985:1 | isCallOrNewExpression | 0.050193 | 0.060006 | +0.009813 | 1.1955 |
| src/compiler/utilities.ts:3917:1 | isStringDoubleQuoted | 0.048410 | 0.058213 | +0.009802 | 1.2025 |
| src/compiler/transformers/module/system.ts:1:1 |  | 0.041228 | 0.051018 | +0.009790 | 1.2375 |
| src/compiler/utilities.ts:6483:1 | hostUsesCaseSensitiveFileNames | 0.118112 | 0.127861 | +0.009749 | 1.0825 |
| src/compiler/transformer.ts:218:1 | wrapScriptTransformerFactory | 0.040574 | 0.050323 | +0.009749 | 1.2403 |
| src/compiler/sourcemap.ts:673:1 | sameMappedPosition | 4.612123 | 4.621799 | +0.009676 | 1.0021 |
| src/compiler/utilities.ts:6582:1 | getPossibleOriginalInputExtensionForExtension | 0.040066 | 0.049722 | +0.009656 | 1.2410 |
| src/compiler/factory/nodeTests.ts:387:1 | isCaseKeyword | 0.046418 | 0.056063 | +0.009645 | 1.2078 |
| src/compiler/tracing.ts:378:1 |  | 0.041023 | 0.050667 | +0.009643 | 1.2351 |
| src/compiler/types.ts:3274:1 | JsxAttribute | 0.041649 | 0.051255 | +0.009605 | 1.2306 |
| src/compiler/types.ts:7197:1 | AssignmentDeclarationKind | 0.041427 | 0.051007 | +0.009580 | 1.2313 |
| src/compiler/utilities.ts:8708:1 | formatMessage | 0.074257 | 0.083773 | +0.009515 | 1.1281 |
| src/compiler/factory/utilities.ts:465:1 | createExpressionForObjectLiteralElementLike | 0.204643 | 0.214133 | +0.009490 | 1.0464 |
| src/compiler/types.ts:1637:1 | DefaultKeyword | 0.043390 | 0.052878 | +0.009488 | 1.2187 |
| src/compiler/moduleSpecifiers.ts:698:1 | getNearestAncestorDirectoryWithPackageJson | 0.125833 | 0.135308 | +0.009475 | 1.0753 |
| src/compiler/_namespaces/ts.performance.ts:3:1 |  | 0.040193 | 0.049636 | +0.009444 | 1.2350 |
| src/compiler/resolutionCache.ts:310:1 | canWatchDirectoryOrFile | 0.041553 | 0.050980 | +0.009427 | 1.2269 |
| src/compiler/factory/nodeFactory.ts:7299:1 | propagateChildFlags | 0.051687 | 0.061080 | +0.009392 | 1.1817 |
| src/compiler/factory/emitHelpers.ts:846:1 |  | 0.066763 | 0.076135 | +0.009372 | 1.1404 |
| src/compiler/factory/nodeTests.ts:736:1 | isDoStatement | 0.043067 | 0.052400 | +0.009333 | 1.2167 |
| src/compiler/factory/emitNode.ts:62:1 | disposeEmitNodes | 0.058386 | 0.067700 | +0.009315 | 1.1595 |
| src/compiler/factory/nodeTests.ts:844:1 | isImportClause | 0.047545 | 0.056852 | +0.009307 | 1.1958 |
| src/compiler/transformers/generators.ts:262:1 | CodeBlock | 0.037534 | 0.046838 | +0.009305 | 1.2479 |
| src/compiler/performanceCore.ts:14:1 | PerformanceTime | 0.038269 | 0.047538 | +0.009269 | 1.2422 |
| src/compiler/utilitiesPublic.ts:1860:1 | isDeclarationBindingElement | 0.038849 | 0.048111 | +0.009262 | 1.2384 |
| src/compiler/parser.ts:1344:1 | createSourceFile | 0.272997 | 0.282242 | +0.009245 | 1.0339 |
| src/compiler/scanner.ts:508:1 | computeLineAndCharacterOfPosition | 0.038148 | 0.047392 | +0.009244 | 1.2423 |
| src/compiler/builder.ts:732:1 | getPendingEmitKindWithSeen | 0.041681 | 0.050902 | +0.009221 | 1.2212 |
| src/compiler/transformers/classFields.ts:3343:1 | isReservedPrivateName | 0.290467 | 0.299646 | +0.009179 | 1.0316 |
| src/compiler/utilities.ts:2724:1 | isLiteralImportTypeNode | 0.287651 | 0.296810 | +0.009160 | 1.0318 |
| src/compiler/transformers/utilities.ts:871:1 | isSimpleParameterList | 0.038629 | 0.047781 | +0.009152 | 1.2369 |
| src/compiler/types.ts:5719:1 | AssertsIdentifierTypePredicate | 0.037473 | 0.046581 | +0.009108 | 1.2431 |
| src/compiler/utilities.ts:2751:1 | isHoistedVariableStatement | 0.236981 | 0.246089 | +0.009108 | 1.0384 |
| src/compiler/factory/nodeTests.ts:812:1 | isTypeAliasDeclaration | 0.046915 | 0.056003 | +0.009088 | 1.1937 |
| src/compiler/utilities.ts:2653:1 | isDeclarationReadonly | 0.052364 | 0.061434 | +0.009071 | 1.1732 |
| src/compiler/utilities.ts:10883:1 | hasContextSensitiveParameters | 0.044596 | 0.053666 | +0.009070 | 1.2034 |
| src/compiler/utilitiesPublic.ts:2043:1 | isUnaryExpressionKind | 0.053836 | 0.062901 | +0.009066 | 1.1684 |
| src/compiler/program.ts:845:1 | SourceFileImportsList | 0.045688 | 0.054728 | +0.009040 | 1.1979 |
| src/compiler/path.ts:1063:1 | getRelativePathFromFile | 0.048671 | 0.057710 | +0.009039 | 1.1857 |
| src/compiler/utilitiesPublic.ts:1675:1 | isFunctionLike | 0.049580 | 0.058608 | +0.009028 | 1.1821 |
| src/compiler/utilities.ts:2422:1 | createDiagnosticForNodeArray | 0.042393 | 0.051337 | +0.008945 | 1.2110 |
| src/compiler/builder.ts:1164:1 | IncrementalBuildInfoBundlePendingEmit | 0.040283 | 0.049211 | +0.008928 | 1.2216 |
| src/compiler/factory/nodeTests.ts:824:1 | isModuleBlock | 0.046285 | 0.055205 | +0.008920 | 1.1927 |
| src/compiler/path.ts:579:1 | combinePaths | 0.038603 | 0.047462 | +0.008859 | 1.2295 |
| src/compiler/types.ts:1725:1 | GeneratedIdentifier | 0.049678 | 0.058518 | +0.008840 | 1.1780 |
| src/compiler/utilities.ts:8607:1 | getLocaleSpecificMessage | 0.039572 | 0.048407 | +0.008835 | 1.2233 |
| src/compiler/types.ts:1986:1 | ObjectLiteralElementLike | 0.038444 | 0.047266 | +0.008822 | 1.2295 |
| src/compiler/factory/utilities.ts:1627:1 | createAccessorPropertyGetRedirector | 0.279319 | 0.288122 | +0.008804 | 1.0315 |
| src/compiler/factory/nodeTests.ts:820:1 | isModuleDeclaration | 0.048092 | 0.056843 | +0.008751 | 1.1820 |
| src/compiler/watch.ts:944:1 | CreateWatchCompilerHostOfFilesAndCompilerOptionsInput | 0.044571 | 0.053300 | +0.008729 | 1.1958 |
| src/compiler/factory/nodeTests.ts:724:1 | isEmptyStatement | 0.044088 | 0.052814 | +0.008726 | 1.1979 |
| src/compiler/utilities.ts:2134:1 | isModuleAugmentationExternal | 0.279504 | 0.288228 | +0.008724 | 1.0312 |
| src/compiler/types.ts:4692:1 | FilePreprocessingDiagnostics | 0.039143 | 0.047859 | +0.008717 | 1.2227 |
| src/compiler/builder.ts:1663:1 | createBuilderProgram | 0.040068 | 0.048755 | +0.008687 | 1.2168 |
| src/compiler/utilities.ts:3704:1 | isInExpressionContext | 0.081280 | 0.089938 | +0.008658 | 1.1065 |
| src/compiler/diagnosticInformationMap.generated.ts:6:1 | diag | 0.043410 | 0.052047 | +0.008637 | 1.1990 |
| src/compiler/builder.ts:1157:1 | IncrementalBundleEmitBuildInfoFileInfo | 0.039306 | 0.047932 | +0.008626 | 1.2195 |
| src/compiler/scanner.ts:436:1 | computeLineStarts | 0.038522 | 0.047139 | +0.008617 | 1.2237 |
| src/compiler/parser.ts:420:1 |  | 0.038869 | 0.047474 | +0.008604 | 1.2214 |
| src/compiler/utilities.ts:8330:1 | isTypeNodeKind | 0.043797 | 0.052401 | +0.008604 | 1.1964 |
| src/compiler/utilities.ts:6570:1 | getDeclarationEmitExtensionForPath | 0.040535 | 0.049137 | +0.008602 | 1.2122 |
| src/compiler/utilitiesPublic.ts:1732:1 | isClassElement | 0.047591 | 0.056188 | +0.008597 | 1.1806 |
| src/compiler/checker.ts:54207:1 | isNotOverload | 0.116900 | 0.125490 | +0.008590 | 1.0735 |
| src/compiler/factory/nodeTests.ts:728:1 | isExpressionStatement | 0.044387 | 0.052920 | +0.008534 | 1.1923 |
| src/compiler/visitorPublic.ts:564:1 | visitCommaListElements | 0.042546 | 0.051079 | +0.008533 | 1.2005 |
| src/compiler/utilitiesPublic.ts:2601:1 |  | 0.039028 | 0.047522 | +0.008494 | 1.2176 |
| src/compiler/performanceCore.ts:61:1 | tryGetPerformanceHooks | 0.335805 | 0.344220 | +0.008415 | 1.0251 |
| src/compiler/factory/nodeTests.ts:744:1 | isForStatement | 0.043649 | 0.052049 | +0.008400 | 1.1924 |
| src/compiler/factory/nodeTests.ts:362:1 | isAbstractModifier | 0.047149 | 0.055518 | +0.008369 | 1.1775 |
| src/compiler/types.ts:2597:1 | LogicalOperator | 0.042912 | 0.051278 | +0.008366 | 1.1949 |
| src/compiler/path.ts:278:1 | getDirectoryPath | 0.039793 | 0.048131 | +0.008338 | 1.2095 |
| src/compiler/executeCommandLine.ts:1:1 |  | 0.040914 | 0.049245 | +0.008331 | 1.2036 |
| src/compiler/utilities.ts:4854:1 | getAssignmentTarget | 0.165509 | 0.173789 | +0.008280 | 1.0500 |
| src/compiler/factory/nodeTests.ts:862:1 | isImportAttributes | 0.049060 | 0.057262 | +0.008202 | 1.1672 |
| src/compiler/binder.ts:3905:1 | lookupSymbolForName | 0.079262 | 0.087446 | +0.008184 | 1.1033 |
| src/compiler/factory/nodeTests.ts:780:1 | isThrowStatement | 0.045635 | 0.053808 | +0.008173 | 1.1791 |
| src/compiler/builder.ts:1097:1 | IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo | 0.041392 | 0.049561 | +0.008169 | 1.1974 |
| src/compiler/factory/nodeTests.ts:772:1 | isSwitchStatement | 0.046624 | 0.054768 | +0.008144 | 1.1747 |
| src/compiler/utilities.ts:6559:1 | getDeclarationEmitOutputFilePathWorker | 0.161518 | 0.169629 | +0.008111 | 1.0502 |
| src/compiler/builder.ts:265:1 | BuilderProgramStateWithDefinedProgram | 0.038510 | 0.046615 | +0.008105 | 1.2105 |
| src/compiler/sourcemap.ts:444:1 | Mapping | 0.039279 | 0.047353 | +0.008074 | 1.2056 |
| src/compiler/factory/nodeTests.ts:367:1 | isOverrideModifier | 0.047530 | 0.055580 | +0.008049 | 1.1693 |
| src/compiler/types.ts:10251:1 |  | 0.073431 | 0.081475 | +0.008043 | 1.1095 |
| src/compiler/transformers/generators.ts:276:1 | LabeledBlock | 0.037567 | 0.045608 | +0.008041 | 1.2140 |
| src/compiler/utilitiesPublic.ts:1131:1 | getJSDocPublicTagNoCache | 0.055171 | 0.063206 | +0.008035 | 1.1456 |
| src/compiler/sourcemap.ts:368:1 |  | 0.042722 | 0.050745 | +0.008023 | 1.1878 |
| src/compiler/types.ts:4649:1 | AutomaticTypeDirectiveFile | 0.038125 | 0.046148 | +0.008022 | 1.2104 |
| src/compiler/utilities.ts:3189:1 | getThisContainer | 0.039000 | 0.047015 | +0.008015 | 1.2055 |
| src/compiler/factory/nodeTests.ts:788:1 | isDebuggerStatement | 0.045929 | 0.053909 | +0.007980 | 1.1738 |
| src/compiler/types.ts:1849:1 | SignatureDeclaration | 0.041814 | 0.049790 | +0.007976 | 1.1908 |
| src/compiler/transformers/declarations/diagnostics.ts:138:1 | canProduceDiagnostics | 0.329837 | 0.337784 | +0.007947 | 1.0241 |
| src/compiler/factory/nodeTests.ts:768:1 | isWithStatement | 0.046693 | 0.054606 | +0.007914 | 1.1695 |
| src/compiler/factory/utilities.ts:856:1 | tryGetModuleNameFromDeclaration | 0.084824 | 0.092711 | +0.007887 | 1.0930 |
| src/compiler/factory/nodeFactory.ts:7320:1 | getTransformFlagsSubtreeExclusions | 0.045164 | 0.053027 | +0.007863 | 1.1741 |
| src/compiler/factory/nodeTests.ts:784:1 | isTryStatement | 0.045025 | 0.052872 | +0.007847 | 1.1743 |
| src/compiler/utilities.ts:2661:1 | isVarAwaitUsing | 0.044877 | 0.052723 | +0.007846 | 1.1748 |
| src/compiler/sourcemap.ts:389:1 | tryGetSourceMappingURL | 0.131467 | 0.139312 | +0.007844 | 1.0597 |
| src/compiler/checker.ts:1340:1 | CheckMode | 0.042897 | 0.050716 | +0.007819 | 1.1823 |
| src/compiler/utilities.ts:3843:1 | isRequireCall | 0.168397 | 0.176210 | +0.007812 | 1.0464 |
| src/compiler/types.ts:3025:1 | CallExpression | 0.043142 | 0.050924 | +0.007782 | 1.1804 |
| src/compiler/types.ts:8372:1 | EmitFlags | 0.041689 | 0.049469 | +0.007780 | 1.1866 |
| src/compiler/utilities.ts:1029:1 | getEndLinePosition | 0.040169 | 0.047944 | +0.007775 | 1.1936 |
| src/compiler/types.ts:2119:1 | ConstructorDeclaration | 0.041374 | 0.049144 | +0.007770 | 1.1878 |
| src/compiler/factory/nodeTests.ts:1058:1 | isJSDocNullableType | 0.047158 | 0.054904 | +0.007747 | 1.1643 |
| src/compiler/parser.ts:9927:1 |  | 0.041177 | 0.048899 | +0.007722 | 1.1875 |
| src/compiler/types.ts:1748:1 | DeclarationName | 0.048418 | 0.056115 | +0.007697 | 1.1590 |
| src/compiler/factory/nodeTests.ts:740:1 | isWhileStatement | 0.043375 | 0.051060 | +0.007686 | 1.1772 |
| src/compiler/types.ts:3419:1 | ForStatement | 0.042911 | 0.050567 | +0.007656 | 1.1784 |
| src/compiler/types.ts:1638:1 | ExportKeyword | 0.043572 | 0.051218 | +0.007646 | 1.1755 |
| src/compiler/utilities.ts:11124:1 | tryGetJSDocSatisfiesTypeNode | 0.057014 | 0.064654 | +0.007641 | 1.1340 |
| src/compiler/types.ts:1767:1 | DynamicNamedDeclaration | 0.046736 | 0.054371 | +0.007635 | 1.1634 |
| src/compiler/path.ts:140:1 | hasTrailingDirectorySeparator | 0.039163 | 0.046782 | +0.007618 | 1.1945 |
| src/compiler/factory/nodeTests.ts:1112:1 | isJSDocPublicTag | 0.042644 | 0.050262 | +0.007618 | 1.1786 |
| src/compiler/factory/utilities.ts:1217:1 | isRelationalOperatorOrHigher | 0.108941 | 0.116557 | +0.007616 | 1.0699 |
| src/compiler/transformer.ts:117:1 |  | 0.066450 | 0.074036 | +0.007585 | 1.1141 |
| src/compiler/watch.ts:166:1 | getPlainDiagnosticFollowingNewLines | 0.045664 | 0.053246 | +0.007581 | 1.1660 |
| src/compiler/sys.ts:1379:1 | patchWriteFileEnsuringDirectory | 0.075670 | 0.083211 | +0.007541 | 1.0997 |
| src/compiler/types.ts:10238:1 | PragmaDefinition | 0.038194 | 0.045703 | +0.007509 | 1.1966 |
| src/compiler/factory/nodeTests.ts:1096:1 | isJSDocAugmentsTag | 0.043279 | 0.050780 | +0.007502 | 1.1733 |
| src/compiler/types.ts:2083:1 | FunctionLike | 0.044488 | 0.051979 | +0.007492 | 1.1684 |
| src/compiler/types.ts:2925:1 | ParenthesizedExpression | 0.038757 | 0.046208 | +0.007451 | 1.1923 |
| src/compiler/utilities.ts:9647:1 |  | 0.067573 | 0.074999 | +0.007426 | 1.1099 |
| src/compiler/resolutionCache.ts:180:1 | CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations | 0.040349 | 0.047763 | +0.007414 | 1.1837 |
| src/compiler/emitter.ts:6340:1 | getClosingBracket | 0.041216 | 0.048624 | +0.007408 | 1.1797 |
| src/compiler/types.ts:3415:1 | ForInitializer | 0.042478 | 0.049876 | +0.007398 | 1.1742 |
| src/compiler/utilities.ts:12419:1 | suppressTrailingTrivia | 0.055560 | 0.062949 | +0.007389 | 1.1330 |
| src/compiler/types.ts:1821:1 | Decorator | 0.042403 | 0.049780 | +0.007377 | 1.1740 |
| src/compiler/watchUtilities.ts:692:1 | WatchFactory | 0.038977 | 0.046324 | +0.007347 | 1.1885 |
| src/compiler/moduleNameResolver.ts:2360:1 | getTemporaryModuleResolutionState | 0.040710 | 0.048053 | +0.007343 | 1.1804 |
| src/compiler/types.ts:3282:1 | JsxAttributeValue | 0.039433 | 0.046774 | +0.007342 | 1.1862 |
| src/compiler/utilitiesPublic.ts:1239:1 | getJSDocReturnType | 0.292071 | 0.299405 | +0.007334 | 1.0251 |
| src/compiler/utilities.ts:997:1 | isPlainJsFile | 0.039911 | 0.047244 | +0.007333 | 1.1837 |
| src/compiler/utilities.ts:4900:1 | getAssignmentTargetKind | 0.081540 | 0.088873 | +0.007333 | 1.0899 |
| src/compiler/scanner.ts:520:1 | computeLineOfPosition | 0.068072 | 0.075404 | +0.007332 | 1.1077 |
| src/compiler/executeCommandLine.ts:244:1 | generateOptionOutput | 0.041525 | 0.048838 | +0.007313 | 1.1761 |
| src/compiler/builder.ts:1080:1 | IncrementalBuildInfoDiagnostic | 0.041380 | 0.048687 | +0.007306 | 1.1766 |
| src/compiler/core.ts:1256:1 |  | 0.037774 | 0.045075 | +0.007301 | 1.1933 |
| src/compiler/sourcemap.ts:461:1 | decodeMappings | 0.041942 | 0.049217 | +0.007275 | 1.1735 |
| src/compiler/types.ts:8095:1 | ResolvedModuleWithFailedLookupLocations | 0.042870 | 0.050134 | +0.007264 | 1.1695 |
| src/compiler/emitter.ts:590:1 | getOutputJSFileNameWorker | 0.040745 | 0.047987 | +0.007242 | 1.1777 |
| src/compiler/transformers/namedEvaluation.ts:384:1 | transformNamedEvaluationOfPropertyDeclaration | 0.038967 | 0.046182 | +0.007215 | 1.1852 |
| src/compiler/types.ts:8079:1 | Extension | 0.040168 | 0.047376 | +0.007207 | 1.1794 |
| src/compiler/types.ts:3914:1 | JSDocNonNullableType | 0.047081 | 0.054269 | +0.007188 | 1.1527 |
| src/compiler/utilities.ts:4171:1 | getNameOrArgument | 0.041414 | 0.048594 | +0.007179 | 1.1734 |
| src/compiler/utilitiesPublic.ts:1657:1 | isPropertyName | 0.046677 | 0.053839 | +0.007162 | 1.1534 |
| src/compiler/_namespaces/ts.ts:25:1 |  | 0.041176 | 0.048310 | +0.007133 | 1.1732 |
| src/compiler/factory/nodeFactory.ts:7295:1 | propagatePropertyNameFlagsOfChild | 0.047322 | 0.054453 | +0.007132 | 1.1507 |
| src/compiler/resolutionCache.ts:91:1 | ResolutionCache | 0.041889 | 0.048996 | +0.007108 | 1.1697 |
| src/compiler/utilities.ts:9673:1 | getRegularExpressionsForWildcards | 0.068893 | 0.075984 | +0.007091 | 1.1029 |
| src/compiler/utilities.ts:3257:1 | isThisContainerOrFunctionBlock | 0.045301 | 0.052387 | +0.007086 | 1.1564 |
| src/compiler/emitter.ts:6357:1 | ParenthesizerRuleOrSelector | 0.040379 | 0.047445 | +0.007066 | 1.1750 |
| src/compiler/types.ts:1864:1 | CallSignatureDeclaration | 0.043797 | 0.050854 | +0.007057 | 1.1611 |
| src/compiler/types.ts:4250:1 | FlowArrayMutation | 0.038785 | 0.045824 | +0.007039 | 1.1815 |
| src/compiler/factory/nodeTests.ts:866:1 | isImportAttribute | 0.047206 | 0.054241 | +0.007035 | 1.1490 |
| src/compiler/utilities.ts:8547:1 | SourceMapSource | 0.039330 | 0.046336 | +0.007005 | 1.1781 |
| src/compiler/types.ts:2981:1 | PropertyAccessChain | 0.040369 | 0.047369 | +0.007000 | 1.1734 |
| src/compiler/builder.ts:1166:1 | IncrementalBundleEmitBuildInfo | 0.040119 | 0.047111 | +0.006992 | 1.1743 |
| src/compiler/types.ts:3055:1 | BindableObjectDefinePropertyCall | 0.044458 | 0.051403 | +0.006945 | 1.1562 |
| src/compiler/types.ts:3826:1 | TypeOnlyCompatibleAliasDeclaration | 0.040423 | 0.047340 | +0.006917 | 1.1711 |
| src/compiler/transformers/destructuring.ts:600:1 | makeArrayBindingPattern | 0.045380 | 0.052277 | +0.006897 | 1.1520 |
| src/compiler/transformers/namedEvaluation.ts:405:1 | transformNamedEvaluationOfAssignmentExpression | 0.043477 | 0.050361 | +0.006884 | 1.1583 |
| src/compiler/types.ts:2151:1 | SetAccessorDeclaration | 0.041993 | 0.048852 | +0.006859 | 1.1633 |
| src/compiler/utilities.ts:4086:1 | getRightMostAssignedExpression | 0.173753 | 0.180590 | +0.006836 | 1.0393 |
| src/compiler/visitorPublic.ts:518:1 | visitFunctionBody | 0.042604 | 0.049434 | +0.006830 | 1.1603 |
| src/compiler/types.ts:4256:1 | FlowReduceLabel | 0.037560 | 0.044388 | +0.006828 | 1.1818 |
| src/compiler/factory/nodeTests.ts:688:1 | isMetaProperty | 0.043027 | 0.049846 | +0.006820 | 1.1585 |
| src/compiler/types.ts:10231:1 | PragmaArgumentSpecification | 0.038527 | 0.045333 | +0.006805 | 1.1766 |
| src/compiler/types.ts:7324:1 | DiagnosticCategory | 0.040926 | 0.047730 | +0.006804 | 1.1663 |
| src/compiler/utilitiesPublic.ts:1694:1 | isFunctionLikeDeclarationKind | 0.043360 | 0.050164 | +0.006804 | 1.1569 |
| src/compiler/utilities.ts:8676:1 | attachFileToDiagnostics | 0.041778 | 0.048577 | +0.006799 | 1.1627 |
| src/compiler/builderPublic.ts:1:1 |  | 0.045651 | 0.052445 | +0.006794 | 1.1488 |
| src/compiler/emitter.ts:6355:1 | ParenthesizerRule | 0.043361 | 0.050152 | +0.006791 | 1.1566 |
| src/compiler/factory/nodeFactory.ts:7399:1 |  | 0.296272 | 0.303020 | +0.006748 | 1.0228 |
| src/compiler/utilities.ts:2072:1 | isCatchClauseVariableDeclarationOrBindingElement | 0.049793 | 0.056532 | +0.006739 | 1.1353 |
| src/compiler/factory/utilitiesPublic.ts:43:1 | canHaveDecorators | 0.050592 | 0.057316 | +0.006724 | 1.1329 |
| src/compiler/types.ts:4591:1 | CancellationToken | 0.038168 | 0.044876 | +0.006709 | 1.1758 |
| src/compiler/utilities.ts:4407:1 | getExternalModuleName | 0.047633 | 0.054312 | +0.006679 | 1.1402 |
| src/compiler/types.ts:4560:1 | ResolvedConfigFileName | 0.040910 | 0.047581 | +0.006670 | 1.1630 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts:18:1 | transformImpliedNodeFormatDependentModule | 0.226814 | 0.233481 | +0.006667 | 1.0294 |
| src/compiler/utilitiesPublic.ts:1995:1 | isLeftHandSideExpression | 0.058069 | 0.064732 | +0.006663 | 1.1147 |
| src/compiler/transformers/utilities.ts:806:1 | LexicalEnvironment | 0.038360 | 0.045020 | +0.006660 | 1.1736 |
| src/compiler/utilities.ts:12400:1 | suppressLeadingAndTrailingTrivia | 0.049216 | 0.055859 | +0.006643 | 1.1350 |
| src/compiler/utilities.ts:3760:1 | isPartOfTypeQuery | 0.044473 | 0.051087 | +0.006613 | 1.1487 |
| src/compiler/types.ts:1827:1 | TypeParameterDeclaration | 0.041996 | 0.048595 | +0.006599 | 1.1571 |
| src/compiler/utilities.ts:9312:1 |  | 0.109793 | 0.116360 | +0.006567 | 1.0598 |
| src/compiler/utilitiesPublic.ts:1666:1 | isBindingName | 0.046232 | 0.052790 | +0.006559 | 1.1419 |
| src/compiler/types.ts:8313:1 | SourceMapSource | 0.038265 | 0.044813 | +0.006548 | 1.1711 |
| src/compiler/resolutionCache.ts:505:1 | getModuleResolutionHost | 0.042467 | 0.049010 | +0.006544 | 1.1541 |
| src/compiler/types.ts:1632:1 | AbstractKeyword | 0.041572 | 0.048096 | +0.006524 | 1.1569 |
| src/compiler/types.ts:6911:1 | IndexedAccessType | 0.042948 | 0.049458 | +0.006510 | 1.1516 |
| src/compiler/parser.ts:495:1 | hasModifierOfKind | 0.050457 | 0.056957 | +0.006499 | 1.1288 |
| src/compiler/types.ts:4599:1 | FileIncludeKind | 0.038655 | 0.045109 | +0.006454 | 1.1670 |
| src/compiler/factory/nodeTests.ts:870:1 | isNamespaceImport | 0.047058 | 0.053504 | +0.006446 | 1.1370 |
| src/compiler/utilities.ts:3157:1 | getContainingClassExcludingClassDecorators | 0.283406 | 0.289837 | +0.006432 | 1.0227 |
| src/compiler/executeCommandLine.ts:1011:1 | createWatchOfConfigFile | 0.453754 | 0.460184 | +0.006431 | 1.0142 |
| src/compiler/program.ts:651:1 | FormatDiagnosticsHost | 0.039197 | 0.045627 | +0.006430 | 1.1640 |
| src/compiler/builder.ts:1673:1 | createBuilderProgram | 0.039245 | 0.045662 | +0.006417 | 1.1635 |
| src/compiler/utilitiesPublic.ts:2564:1 | hasType | 0.055888 | 0.062276 | +0.006388 | 1.1143 |
| src/compiler/watchUtilities.ts:698:1 | GetDetailWatchInfo | 0.040046 | 0.046416 | +0.006370 | 1.1591 |
| src/compiler/utilitiesPublic.ts:678:1 | getCombinedNodeFlags | 0.056249 | 0.062600 | +0.006351 | 1.1129 |
| src/compiler/utilitiesPublic.ts:1518:1 | isTemplateLiteralKind | 0.041860 | 0.048183 | +0.006324 | 1.1511 |
| src/compiler/builder.ts:2395:1 | getBuildInfoFileVersionMap | 0.127754 | 0.134074 | +0.006320 | 1.0495 |
| src/compiler/utilities.ts:7163:1 | calculateIndent | 0.043021 | 0.049309 | +0.006288 | 1.1462 |
| src/compiler/utilities.ts:4166:1 | isBindableStaticNameExpression | 0.249614 | 0.255900 | +0.006286 | 1.0252 |
| src/compiler/types.ts:2005:1 | ShorthandPropertyAssignment | 0.039369 | 0.045654 | +0.006285 | 1.1596 |
| src/compiler/types.ts:2482:1 | ImportExpression | 0.040785 | 0.047056 | +0.006271 | 1.1538 |
| src/compiler/moduleNameResolver.ts:118:1 | trace | 0.081450 | 0.087712 | +0.006262 | 1.0769 |
| src/compiler/utilities.ts:2758:1 | getLeadingCommentRangesOfNode | 0.044331 | 0.050587 | +0.006256 | 1.1411 |
| src/compiler/utilities.ts:2066:1 | isBlockOrCatchScoped | 0.046228 | 0.052468 | +0.006240 | 1.1350 |
| src/compiler/factory/nodeTests.ts:756:1 | isContinueStatement | 0.050656 | 0.056885 | +0.006229 | 1.1230 |
| src/compiler/moduleSpecifiers.ts:774:1 | getAllModulePaths | 0.119961 | 0.126176 | +0.006215 | 1.0518 |
| src/compiler/path.ts:803:1 | ensureTrailingDirectorySeparator | 0.038476 | 0.044687 | +0.006212 | 1.1614 |
| src/compiler/factory/utilities.ts:1494:1 | isExportOrDefaultModifier | 0.076412 | 0.082618 | +0.006206 | 1.0812 |
| src/compiler/builder.ts:187:1 | BuilderFileEmit | 0.038938 | 0.045144 | +0.006206 | 1.1594 |
| src/compiler/utilities.ts:8944:1 | messageTextEqualityComparer | 0.038864 | 0.045059 | +0.006195 | 1.1594 |
| src/compiler/types.ts:2969:1 | PropertyAccessExpression | 0.040249 | 0.046431 | +0.006181 | 1.1536 |
| src/compiler/factory/nodeTests.ts:858:1 | isAssertEntry | 0.048407 | 0.054579 | +0.006172 | 1.1275 |
| src/compiler/utilities.ts:2788:1 | isPartOfTypeNode | 0.130780 | 0.136949 | +0.006169 | 1.0472 |
| src/compiler/utilities.ts:11458:1 | findConstructorDeclaration | 0.042977 | 0.049131 | +0.006154 | 1.1432 |
| src/compiler/factory/nodeTests.ts:989:1 | isCatchClause | 0.056277 | 0.062411 | +0.006133 | 1.1090 |
| src/compiler/factory/nodeTests.ts:1152:1 | isJSDocReturnTag | 0.044739 | 0.050872 | +0.006133 | 1.1371 |
| src/compiler/_namespaces/ts.ts:54:1 |  | 0.041419 | 0.047541 | +0.006123 | 1.1478 |
| src/compiler/types.ts:4235:1 | FlowSwitchClause | 0.039136 | 0.045252 | +0.006116 | 1.1563 |
| src/compiler/factory/utilities.ts:1065:1 | isStringOrNumericLiteral | 0.044204 | 0.050283 | +0.006079 | 1.1375 |
| src/compiler/utilities.ts:3618:1 | isJSXTagName | 0.085999 | 0.092066 | +0.006067 | 1.0706 |
| src/compiler/sys.ts:896:1 | FsWatchWorkerWatcher | 0.039901 | 0.045968 | +0.006066 | 1.1520 |
| src/compiler/moduleNameResolver.ts:3202:1 | mangleScopedPackageNameWithTrace | 0.041196 | 0.047260 | +0.006064 | 1.1472 |
| src/compiler/types.ts:4471:1 | PragmaContext | 0.042456 | 0.048510 | +0.006054 | 1.1426 |
| src/compiler/utilitiesPublic.ts:1798:1 | isObjectLiteralElementLike | 0.057770 | 0.063809 | +0.006038 | 1.1045 |
| src/compiler/utilities.ts:10812:1 | getContainingNodeArray | 0.039622 | 0.045657 | +0.006035 | 1.1523 |
| src/compiler/transformers/destructuring.ts:620:1 | makeBindingElement | 0.080748 | 0.086774 | +0.006025 | 1.0746 |
| src/compiler/transformers/es2018.ts:154:1 | transformES2018 | 0.199779 | 0.205795 | +0.006015 | 1.0301 |
| src/compiler/utilities.ts:7364:1 | modifierToFlag | 0.051980 | 0.057987 | +0.006007 | 1.1156 |
| src/compiler/utilities.ts:6161:1 |  | 0.039211 | 0.045209 | +0.005998 | 1.1530 |
| src/compiler/moduleNameResolver.ts:2743:1 | hasOneAsterisk | 0.039669 | 0.045657 | +0.005988 | 1.1509 |
| src/compiler/factory/nodeTests.ts:684:1 | isNonNullExpression | 0.043357 | 0.049343 | +0.005986 | 1.1381 |
| src/compiler/transformers/utilities.ts:642:1 | isNonStaticMethodOrAccessorWithPrivateName | 0.234898 | 0.240865 | +0.005967 | 1.0254 |
| src/compiler/utilities.ts:4475:1 | isJSDocConstructSignature | 0.252032 | 0.257991 | +0.005959 | 1.0236 |
| src/compiler/moduleSpecifiers.ts:255:1 | updateModuleSpecifier | 0.042726 | 0.048675 | +0.005949 | 1.1392 |
| src/compiler/emitter.ts:6345:1 | TempFlags | 0.040157 | 0.046099 | +0.005942 | 1.1480 |
| src/compiler/sys.ts:245:1 | createDynamicPriorityPollingWatchFile | 0.041196 | 0.047130 | +0.005933 | 1.1440 |
| src/compiler/commandLineParser.ts:3840:1 | normalizeNonListOptionValue | 0.041104 | 0.047028 | +0.005925 | 1.1441 |
| src/compiler/types.ts:2703:1 | ArrayAssignmentElement | 0.041183 | 0.047089 | +0.005906 | 1.1434 |
| src/compiler/types.ts:2464:1 | TrueLiteral | 0.042552 | 0.048456 | +0.005904 | 1.1387 |
| src/compiler/utilities.ts:11130:1 | getEscapedTextOfJsxAttributeName | 0.213070 | 0.218945 | +0.005875 | 1.0276 |
| src/compiler/utilities.ts:7195:1 | hasSyntacticModifier | 0.069680 | 0.075553 | +0.005873 | 1.0843 |
| src/compiler/types.ts:2931:1 | JSDocTypeAssertion | 0.044795 | 0.050656 | +0.005861 | 1.1308 |
| src/compiler/builderPublic.ts:253:1 | createAbstractBuilder | 0.156394 | 0.162250 | +0.005855 | 1.0374 |
| src/compiler/utilitiesPublic.ts:1364:1 | isGetOrSetAccessorDeclaration | 0.044482 | 0.050333 | +0.005851 | 1.1315 |
| src/compiler/core.ts:805:1 | sortAndDeduplicate | 0.039141 | 0.044992 | +0.005851 | 1.1495 |
| src/compiler/utilitiesPublic.ts:1185:1 | getJSDocThisTag | 0.054708 | 0.060552 | +0.005844 | 1.1068 |
| src/compiler/commandLineParser.ts:2438:1 | AdamicJsonRecoveryValue | 0.039650 | 0.045493 | +0.005843 | 1.1474 |
| src/compiler/builder.ts:1084:1 | IncrementalBuildInfoEmitDiagnostic | 0.040629 | 0.046469 | +0.005840 | 1.1437 |
| src/compiler/moduleNameResolver.ts:1725:1 | nodeNextModuleNameResolver | 0.043416 | 0.049242 | +0.005826 | 1.1342 |
| src/compiler/types.ts:6207:1 | UnderscoreEscapedMap | 0.041071 | 0.046879 | +0.005808 | 1.1414 |
| src/compiler/visitorPublic.ts:542:1 | visitIterationBody | 0.049177 | 0.054971 | +0.005794 | 1.1178 |
| src/compiler/sys.ts:941:1 | createFsWatchCallbackForDirectoryWatcherCallback | 0.039497 | 0.045286 | +0.005789 | 1.1466 |
| src/compiler/visitorPublic.ts:543:1 | visitIterationBody | 0.049111 | 0.054867 | +0.005756 | 1.1172 |
| src/compiler/utilitiesPublic.ts:2039:1 | isUnaryExpression | 0.054483 | 0.060234 | +0.005752 | 1.1056 |
| src/compiler/utilities.ts:11468:1 | NameResolverOptions | 0.040692 | 0.046433 | +0.005741 | 1.1411 |
| src/compiler/factory/utilities.ts:1155:1 | isIdentifierOrThisTypeNode | 0.249696 | 0.255424 | +0.005728 | 1.0229 |
| src/compiler/utilities.ts:8685:1 | createFileDiagnostic | 4.816834 | 4.822554 | +0.005719 | 1.0012 |
| src/compiler/resolutionCache.ts:343:1 | canWatchAffectingLocation | 0.047818 | 0.053533 | +0.005714 | 1.1195 |
| src/compiler/utilities.ts:10387:1 | rangeOfTypeParameters | 0.040159 | 0.045869 | +0.005710 | 1.1422 |
| src/compiler/types.ts:1932:1 | PropertyDeclaration | 0.061395 | 0.067100 | +0.005704 | 1.0929 |
| src/compiler/sys.ts:111:1 | getModifiedTime | 0.075168 | 0.080865 | +0.005697 | 1.0758 |
| src/compiler/utilities.ts:5291:1 | isStringANonContextualKeyword | 0.058880 | 0.064560 | +0.005680 | 1.0965 |
| src/compiler/commandLineParser.ts:2351:1 |  | 0.070591 | 0.076257 | +0.005666 | 1.0803 |
| src/compiler/commandLineParser.ts:151:1 |  | 0.045777 | 0.051418 | +0.005641 | 1.1232 |
| src/compiler/utilitiesPublic.ts:2603:1 | guessIndentation | 0.038511 | 0.044141 | +0.005630 | 1.1462 |
| src/compiler/utilitiesPublic.ts:2648:1 | isInternalDeclaration | 0.059149 | 0.064740 | +0.005591 | 1.0945 |
| src/compiler/builder.ts:1570:1 | BuilderProgramKind | 0.037986 | 0.043561 | +0.005576 | 1.1468 |
| src/compiler/transformers/classFields.ts:1:1 |  | 0.040749 | 0.046310 | +0.005561 | 1.1365 |
| src/compiler/utilitiesPublic.ts:1180:1 | getJSDocEnumTag | 0.052388 | 0.057947 | +0.005559 | 1.1061 |
| src/compiler/utilities.ts:2633:1 | isGlobalSourceFile | 0.110710 | 0.116258 | +0.005548 | 1.0501 |
| src/compiler/utilities.ts:2332:1 | getEnclosingBlockScopeContainer | 0.064745 | 0.070243 | +0.005498 | 1.0849 |
| src/compiler/types.ts:3543:1 | ClassLikeDeclarationBase | 0.055344 | 0.060836 | +0.005493 | 1.0992 |
| src/compiler/builderPublic.ts:22:1 | BuilderProgramHost | 0.046486 | 0.051942 | +0.005456 | 1.1174 |
| src/compiler/types.ts:7223:1 | FileExtensionInfo | 0.044511 | 0.049965 | +0.005454 | 1.1225 |
| src/compiler/utilities.ts:10395:1 | HostWithIsSourceOfProjectReferenceRedirect | 0.038308 | 0.043708 | +0.005399 | 1.1409 |
| src/compiler/utilities.ts:9983:1 |  | 0.040161 | 0.045548 | +0.005387 | 1.1341 |
| src/compiler/factory/utilities.ts:280:1 | createExpressionForJsxFragment | 0.133998 | 0.139385 | +0.005387 | 1.0402 |
| src/compiler/types.ts:7071:1 | IndexKind | 0.040245 | 0.045622 | +0.005377 | 1.1336 |
| src/compiler/utilitiesPublic.ts:1372:1 | isElementAccessChain | 0.249110 | 0.254487 | +0.005377 | 1.0216 |
| src/compiler/utilities.ts:4349:1 | tryGetModuleSpecifierFromDeclaration | 0.041974 | 0.047345 | +0.005371 | 1.1280 |
| src/compiler/transformers/classFields.ts:3318:1 | createPrivateStaticFieldInitializer | 0.076448 | 0.081799 | +0.005350 | 1.0700 |
| src/compiler/builder.ts:136:1 | ReusableBuilderProgramState | 0.038978 | 0.044327 | +0.005350 | 1.1372 |
| src/compiler/factory/nodeTests.ts:692:1 | isSyntheticExpression | 0.042767 | 0.048105 | +0.005338 | 1.1248 |
| src/compiler/utilitiesPublic.ts:1498:1 | isLiteralExpression | 0.079702 | 0.085016 | +0.005313 | 1.0667 |
| src/compiler/types.ts:1980:1 | ObjectLiteralElement | 0.038357 | 0.043648 | +0.005291 | 1.1380 |
| src/compiler/utilitiesPublic.ts:1257:1 | getJSDocTagsWorker | 0.116476 | 0.121749 | +0.005273 | 1.0453 |
| src/compiler/performance.ts:62:1 |  | 0.038208 | 0.043462 | +0.005254 | 1.1375 |
| src/compiler/utilities.ts:8984:1 | getSetExternalModuleIndicator | 0.039399 | 0.044649 | +0.005250 | 1.1333 |
| src/compiler/program.ts:1323:1 | getConfigFileParsingDiagnostics | 0.042661 | 0.047903 | +0.005242 | 1.1229 |
| src/compiler/transformers/utilities.ts:791:1 | PrivateEnvironment | 0.038196 | 0.043437 | +0.005241 | 1.1372 |
| src/compiler/utilities.ts:8035:1 | closeFileWatcher | 0.077462 | 0.082690 | +0.005228 | 1.0675 |
| src/compiler/types.ts:1676:1 | ClassMemberModifier | 0.048750 | 0.053978 | +0.005228 | 1.1072 |
| src/compiler/transformers/utilities.ts:124:1 | containsDefaultReference | 0.185413 | 0.190611 | +0.005198 | 1.0280 |
| src/compiler/utilities.ts:5153:1 | getAliasDeclarationFromName | 0.113037 | 0.118235 | +0.005198 | 1.0460 |
| src/compiler/performance.ts:97:1 | measure | 0.076879 | 0.082068 | +0.005190 | 1.0675 |
| src/compiler/factory/nodeTests.ts:1070:1 | isJSDocFunctionType | 0.050562 | 0.055701 | +0.005139 | 1.1016 |
| src/compiler/builder.ts:841:1 | handleDtsMayChangeOf | 0.078463 | 0.083601 | +0.005137 | 1.0655 |
| src/compiler/utilities.ts:2054:1 | getTextOfConstantValue | 0.069264 | 0.074390 | +0.005126 | 1.0740 |
| src/compiler/moduleNameResolver.ts:930:1 | isMissingPackageJsonInfo | 0.040114 | 0.045225 | +0.005112 | 1.1274 |
| src/compiler/builderStatePublic.ts:12:1 | OutputFile | 0.045358 | 0.050452 | +0.005094 | 1.1123 |
| src/compiler/core.ts:2262:1 | tryRemoveSuffix | 0.042179 | 0.047272 | +0.005093 | 1.1207 |
| src/compiler/types.ts:4677:1 | FilePreprocessingFileExplainingDiagnostic | 0.038228 | 0.043289 | +0.005061 | 1.1324 |
| src/compiler/parser.ts:9929:1 | markAsIncrementallyParsed | 0.040827 | 0.045883 | +0.005055 | 1.1238 |
| src/compiler/watch.ts:904:1 | CreateWatchCompilerHostInput | 0.045746 | 0.050800 | +0.005054 | 1.1105 |
| src/compiler/utilities.ts:7190:1 | hasEffectiveModifier | 0.069584 | 0.074625 | +0.005041 | 1.0724 |
| src/compiler/moduleNameResolver.ts:3354:1 | shouldAllowImportingTsExtension | 0.040345 | 0.045378 | +0.005033 | 1.1248 |
| src/compiler/utilities.ts:9320:1 |  | 0.095397 | 0.100411 | +0.005015 | 1.0526 |
| src/compiler/factory/nodeTests.ts:347:1 | isAwaitKeyword | 0.045852 | 0.050859 | +0.005008 | 1.1092 |
| src/compiler/utilities.ts:9377:1 | getNameOfScriptTarget | 0.059682 | 0.064680 | +0.004998 | 1.0837 |
| src/compiler/types.ts:3247:1 | JsxSelfClosingElement | 0.040964 | 0.045961 | +0.004997 | 1.1220 |
| src/compiler/utilities.ts:2996:1 | isVariableLike | 0.043678 | 0.048673 | +0.004995 | 1.1144 |
| src/compiler/builder.ts:1173:1 | IncrementalBuildInfo | 0.040332 | 0.045282 | +0.004950 | 1.1227 |
| src/compiler/transformers/generators.ts:1:1 |  | 0.040236 | 0.045181 | +0.004945 | 1.1229 |
| src/compiler/utilities.ts:4402:1 | shouldRewriteModuleSpecifier | 0.038245 | 0.043188 | +0.004943 | 1.1292 |
| src/compiler/types.ts:5698:1 | ThisTypePredicate | 0.038245 | 0.043174 | +0.004929 | 1.1289 |
| src/compiler/types.ts:1978:1 | InitializedPropertyDeclaration | 0.038819 | 0.043740 | +0.004921 | 1.1268 |
| src/compiler/utilities.ts:4381:1 | tryGetImportFromModuleSpecifier | 0.038255 | 0.043121 | +0.004866 | 1.1272 |
| src/compiler/program.ts:520:1 | changeCompilerHostLikeToUseCache | 0.170610 | 0.175450 | +0.004840 | 1.0284 |
| src/compiler/utilities.ts:9316:1 |  | 0.098324 | 0.103133 | +0.004809 | 1.0489 |
| src/compiler/executeCommandLine.ts:240:1 | getDisplayNameTextOfOption | 0.043356 | 0.048162 | +0.004805 | 1.1108 |
| src/compiler/types.ts:6833:1 | IterationTypes | 0.041177 | 0.045978 | +0.004801 | 1.1166 |
| src/compiler/builder.ts:1176:1 | isIncrementalBundleEmitBuildInfo | 0.040679 | 0.045480 | +0.004801 | 1.1180 |
| src/compiler/types.ts:3009:1 | ElementAccessChain | 0.039163 | 0.043929 | +0.004766 | 1.1217 |
| src/compiler/commandLineParser.ts:2357:1 |  | 0.087920 | 0.092685 | +0.004764 | 1.0542 |
| src/compiler/types.ts:2474:1 | ThisExpression | 0.042688 | 0.047439 | +0.004752 | 1.1113 |
| src/compiler/utilities.ts:3866:1 | isVariableDeclarationInitializedToRequire | 0.048280 | 0.053022 | +0.004742 | 1.0982 |
| src/compiler/executeCommandLine.ts:99:1 | StatisticType | 0.040145 | 0.044873 | +0.004728 | 1.1178 |
| src/compiler/utilitiesPublic.ts:1161:1 | getJSDocReadonlyTagNoCache | 0.054819 | 0.059539 | +0.004720 | 1.0861 |
| src/compiler/types.ts:2396:1 | Expression | 0.040610 | 0.045325 | +0.004715 | 1.1161 |
| src/compiler/factory/emitHelpers.ts:866:1 |  | 0.067946 | 0.072659 | +0.004714 | 1.0694 |
| src/compiler/types.ts:5732:1 | AnyImportOrJsDocImport | 0.038637 | 0.043349 | +0.004712 | 1.1219 |
| src/compiler/types.ts:5712:1 | AssertsThisTypePredicate | 0.037818 | 0.042517 | +0.004700 | 1.1243 |
| src/compiler/factory/nodeTests.ts:1074:1 | isJSDocVariadicType | 0.046542 | 0.051241 | +0.004699 | 1.1010 |
| src/compiler/builder.ts:2229:1 | toBuilderStateFileInfoForMultiEmit | 0.040045 | 0.044736 | +0.004691 | 1.1171 |
| src/compiler/commandLineParser.ts:2119:1 |  | 0.069119 | 0.073810 | +0.004691 | 1.0679 |
| src/compiler/types.ts:4656:1 | FileIncludeReason | 0.038085 | 0.042775 | +0.004689 | 1.1231 |
| src/compiler/utilities.ts:10610:1 | isIdentifierInNonEmittingHeritageClause | 0.043713 | 0.048397 | +0.004684 | 1.1071 |
| src/compiler/core.ts:2198:1 | levenshteinWithMax | 0.039362 | 0.044041 | +0.004680 | 1.1189 |
| src/compiler/utilities.ts:2525:1 | getErrorSpanForArrowFunction | 0.121710 | 0.126386 | +0.004676 | 1.0384 |
| src/compiler/utilities.ts:9790:1 | FileMatcherPatterns | 0.038043 | 0.042709 | +0.004666 | 1.1226 |
| src/compiler/utilities.ts:6164:1 |  | 0.065526 | 0.070191 | +0.004664 | 1.0712 |
| src/compiler/utilities.ts:670:1 | createSingleLineStringWriter | 0.039161 | 0.043821 | +0.004660 | 1.1190 |
| src/compiler/commandLineParser.ts:3348:1 | shouldReportNoInputFiles | 0.045942 | 0.050589 | +0.004647 | 1.1011 |
| src/compiler/types.ts:4980:1 | CustomTransformers | 0.037780 | 0.042420 | +0.004640 | 1.1228 |
| src/compiler/types.ts:2912:1 | TemplateLiteral | 0.045804 | 0.050430 | +0.004625 | 1.1010 |
| src/compiler/factory/nodeFactory.ts:492:1 | createNodeFactory | 0.278264 | 0.282889 | +0.004625 | 1.0166 |
| src/compiler/transformer.ts:120:1 | getTransformers | 0.038380 | 0.043003 | +0.004623 | 1.1205 |
| src/compiler/parser.ts:476:1 | isAnExternalModuleIndicatorNode | 0.231673 | 0.236290 | +0.004617 | 1.0199 |
| src/compiler/utilitiesPublic.ts:2629:1 | isJSDocLinkLike | 0.046324 | 0.050927 | +0.004603 | 1.0994 |
| src/compiler/utilities.ts:2349:1 | declarationNameToString | 0.048177 | 0.052771 | +0.004594 | 1.0954 |
| src/compiler/_namespaces/ts.ts:24:1 |  | 0.040233 | 0.044826 | +0.004592 | 1.1141 |
| src/compiler/transformers/namedEvaluation.ts:444:1 | transformNamedEvaluationOfExportAssignment | 0.039696 | 0.044266 | +0.004570 | 1.1151 |
| src/compiler/utilitiesPublic.ts:644:1 | getCombinedFlags | 0.242155 | 0.246723 | +0.004568 | 1.0189 |
| src/compiler/parser.ts:414:1 | SpeculationKind | 0.038473 | 0.043008 | +0.004534 | 1.1179 |
| src/compiler/types.ts:1840:1 | SignatureDeclarationBase | 0.043352 | 0.047877 | +0.004524 | 1.1044 |
| src/compiler/builder.ts:2218:1 | addToAffectedFilesPendingEmit | 0.067074 | 0.071597 | +0.004523 | 1.0674 |
| src/compiler/factory/nodeTests.ts:967:1 | isJsxExpression | 0.055093 | 0.059590 | +0.004497 | 1.0816 |
| src/compiler/tsbuildPublic.ts:134:1 |  | 0.043664 | 0.048160 | +0.004497 | 1.1030 |
| src/compiler/factory/nodeTests.ts:1120:1 | isJSDocProtectedTag | 0.043341 | 0.047835 | +0.004493 | 1.1037 |
| src/compiler/types.ts:4540:1 | ParseConfigHost | 0.041239 | 0.045729 | +0.004491 | 1.1089 |
| src/compiler/watch.ts:1:1 |  | 0.040265 | 0.044754 | +0.004489 | 1.1115 |
| src/compiler/types.ts:10295:1 | JSDocParsingMode | 0.037385 | 0.041862 | +0.004477 | 1.1198 |
| src/compiler/utilitiesPublic.ts:1843:1 | isAssignmentPattern | 0.045065 | 0.049526 | +0.004461 | 1.0990 |
| src/compiler/types.ts:5755:1 | BindingElementOfBareOrAccessedRequire | 0.037687 | 0.042137 | +0.004450 | 1.1181 |
| src/compiler/utilities.ts:8029:1 | isWatchSet | 0.039586 | 0.044016 | +0.004430 | 1.1119 |
| src/compiler/utilities.ts:2337:1 | forEachEnclosingBlockScopeContainer | 0.051649 | 0.056072 | +0.004424 | 1.0857 |
| src/compiler/transformers/generators.ts:307:1 | Instruction | 0.037578 | 0.041995 | +0.004417 | 1.1175 |
| src/compiler/types.ts:5735:1 | AnyImportOrRequire | 0.038769 | 0.043185 | +0.004416 | 1.1139 |
| src/compiler/parser.ts:1381:1 | parseJsonText | 0.039374 | 0.043773 | +0.004399 | 1.1117 |
| src/compiler/utilitiesPublic.ts:1522:1 | isTemplateLiteralToken | 0.077141 | 0.081535 | +0.004394 | 1.0570 |
| src/compiler/utilities.ts:11319:1 | createEvaluator | 0.042518 | 0.046895 | +0.004377 | 1.1029 |
| src/compiler/factory/utilities.ts:910:1 | getTargetOfBindingOrAssignmentElement | 0.254669 | 0.259040 | +0.004372 | 1.0172 |
| src/compiler/types.ts:10324:1 | PragmaArgTypeMaybeCapture | 0.037627 | 0.041996 | +0.004369 | 1.1161 |
| src/compiler/sourcemap.ts:454:1 | SourceMapping | 0.040269 | 0.044614 | +0.004344 | 1.1079 |
| src/compiler/utilities.ts:2061:1 | makeIdentifierFromModuleName | 0.038560 | 0.042904 | +0.004344 | 1.1126 |
| src/compiler/utilities.ts:4535:1 | canHaveFlowNode | 0.047462 | 0.051789 | +0.004327 | 1.0912 |
| src/compiler/utilities.ts:4226:1 | getElementOrPropertyAccessName | 0.205635 | 0.209949 | +0.004315 | 1.0210 |
| src/compiler/performance.ts:64:1 |  | 0.039362 | 0.043671 | +0.004308 | 1.1095 |
| src/compiler/types.ts:8347:1 | SnippetElement | 0.038520 | 0.042824 | +0.004304 | 1.1117 |
| src/compiler/types.ts:7331:1 | diagnosticCategoryName | 0.039527 | 0.043819 | +0.004291 | 1.1086 |
| src/compiler/factory/nodeTests.ts:1144:1 | isJSDocEnumTag | 0.044618 | 0.048907 | +0.004289 | 1.0961 |
| src/compiler/path.ts:807:1 | ensureTrailingDirectorySeparator | 0.039795 | 0.044080 | +0.004285 | 1.1077 |
| src/compiler/performanceCore.ts:28:1 |  | 0.038168 | 0.042452 | +0.004284 | 1.1122 |
| src/compiler/transformers/taggedTemplate.ts:1:1 |  | 0.040174 | 0.044445 | +0.004271 | 1.1063 |
| src/compiler/utilitiesPublic.ts:1526:1 | isTemplateMiddleOrTemplateTail | 0.044321 | 0.048592 | +0.004271 | 1.0964 |
| src/compiler/factory/emitHelpers.ts:915:1 |  | 0.069282 | 0.073551 | +0.004270 | 1.0616 |
| src/compiler/moduleNameResolver.ts:944:1 | compilerOptionValueToString | 0.041124 | 0.045390 | +0.004265 | 1.1037 |
| src/compiler/types.ts:5741:1 | AliasDeclarationNode | 0.038050 | 0.042310 | +0.004260 | 1.1119 |
| src/compiler/sys.ts:911:1 | createFsWatchCallbackForFileWatcherCallback | 0.039547 | 0.043803 | +0.004256 | 1.1076 |
| src/compiler/utilities.ts:6691:1 | getSourceFilePathInNewDir | 0.163152 | 0.167404 | +0.004253 | 1.0261 |
| src/compiler/utilitiesPublic.ts:1318:1 | getEffectiveTypeParameterDeclarations | 0.188077 | 0.192329 | +0.004252 | 1.0226 |
| src/compiler/utilitiesPublic.ts:1449:1 | isJSDocPropertyLikeTag | 0.049141 | 0.053381 | +0.004241 | 1.0863 |
| src/compiler/types.ts:5784:1 | RequireVariableStatement | 0.037933 | 0.042166 | +0.004232 | 1.1116 |
| src/compiler/utilities.ts:4929:1 | isCompoundLikeAssignment | 0.042022 | 0.046244 | +0.004222 | 1.1005 |
| src/compiler/types.ts:5758:1 | AnyImportOrRequireStatement | 0.037982 | 0.042201 | +0.004219 | 1.1111 |
| src/compiler/factory/nodeTests.ts:706:1 | isTemplateSpan | 0.043696 | 0.047898 | +0.004202 | 1.0962 |
| src/compiler/utilities.ts:2784:1 |  | 0.042592 | 0.046785 | +0.004193 | 1.0984 |
| src/compiler/builder.ts:310:1 | hasSameKeys | 0.089390 | 0.093561 | +0.004171 | 1.0467 |
| src/compiler/types.ts:2265:1 | TypeLiteralNode | 0.040086 | 0.044250 | +0.004164 | 1.1039 |
| src/compiler/factory/nodeTests.ts:1164:1 | isJSDocTemplateTag | 0.043785 | 0.047937 | +0.004153 | 1.0948 |
| src/compiler/types.ts:2991:1 | SuperPropertyAccessExpression | 0.040945 | 0.045087 | +0.004142 | 1.1012 |
| src/compiler/visitorPublic.ts:522:1 | visitFunctionBody | 0.052085 | 0.056225 | +0.004140 | 1.0795 |
| src/compiler/utilities.ts:2202:1 | isBlockScope | 0.236344 | 0.240470 | +0.004126 | 1.0175 |
| src/compiler/expressionToTypeNode.ts:171:1 |  | 0.045395 | 0.049510 | +0.004115 | 1.0906 |
| src/compiler/utilities.ts:2786:1 |  | 0.038725 | 0.042839 | +0.004114 | 1.1062 |
| src/compiler/types.ts:1629:1 | ModifierToken | 0.051610 | 0.055720 | +0.004111 | 1.0797 |
| src/compiler/factory/nodeConverters.ts:1:1 |  | 0.038488 | 0.042596 | +0.004109 | 1.1068 |
| src/compiler/transformers/utilities.ts:729:1 | getAllDecoratorsOfAccessors | 0.038615 | 0.042719 | +0.004103 | 1.1063 |
| src/compiler/types.ts:1970:1 | PrivateClassElementDeclaration | 0.040340 | 0.044433 | +0.004093 | 1.1015 |
| src/compiler/utilitiesPublic.ts:1175:1 | getJSDocDeprecatedTagNoCache | 0.055894 | 0.059982 | +0.004088 | 1.0731 |
| src/compiler/factory/nodeTests.ts:612:1 | isFunctionExpression | 0.045054 | 0.049141 | +0.004087 | 1.0907 |
| src/compiler/moduleNameResolver.ts:2581:1 | loadModuleFromSelfNameReference | 0.049055 | 0.053137 | +0.004081 | 1.0832 |
| src/compiler/performance.ts:125:1 | getDuration | 0.039518 | 0.043584 | +0.004066 | 1.1029 |
| src/compiler/utilities.ts:834:1 | createModuleNotFoundChain | 0.115184 | 0.119242 | +0.004058 | 1.0352 |
| src/compiler/utilities.ts:4795:1 | getEffectiveJSDocHost | 0.046410 | 0.050460 | +0.004050 | 1.0873 |
| src/compiler/utilitiesPublic.ts:1141:1 | getJSDocPrivateTagNoCache | 0.055796 | 0.059818 | +0.004022 | 1.0721 |
| src/compiler/transformers/ts.ts:235:1 | transformTypeScript | 0.135602 | 0.139621 | +0.004019 | 1.0296 |
| src/compiler/moduleNameResolver.ts:1780:1 | nodeModuleNameResolver | 0.051581 | 0.055596 | +0.004015 | 1.0778 |
| src/compiler/transformer.ts:227:1 | noEmitSubstitution | 0.044080 | 0.048090 | +0.004010 | 1.0910 |
| src/compiler/moduleNameResolver.ts:922:1 | PackageJsonInfoCacheEntry | 0.040842 | 0.044843 | +0.004001 | 1.0980 |
| src/compiler/parser.ts:1311:1 | gatherPossibleChildren | 0.052383 | 0.056381 | +0.003998 | 1.0763 |
| src/compiler/types.ts:7336:1 | ModuleResolutionKind | 0.039660 | 0.043641 | +0.003981 | 1.1004 |
| src/compiler/transformers/utilities.ts:682:1 | getAllDecoratorsOfClass | 0.039518 | 0.043486 | +0.003968 | 1.1004 |
| src/compiler/builder.ts:2246:1 | toProgramEmitPending | 0.040720 | 0.044657 | +0.003937 | 1.0967 |
| src/compiler/builder.ts:1576:1 | BuilderCreationParameters | 0.037908 | 0.041835 | +0.003927 | 1.1036 |
| src/compiler/utilities.ts:6493:1 | ResolveModuleNameResolutionHost | 0.040954 | 0.044874 | +0.003920 | 1.0957 |
| src/compiler/factory/utilities.ts:411:1 | createExpressionForPropertyAssignment | 0.096921 | 0.100839 | +0.003919 | 1.0404 |
| src/compiler/checker.ts:1408:1 | AddUnusedDiagnostic | 0.040862 | 0.044781 | +0.003919 | 1.0959 |
| src/compiler/factory/nodeFactory.ts:7431:1 | mergeEmitNode | 0.444069 | 0.447984 | +0.003915 | 1.0088 |
| src/compiler/moduleNameResolver.ts:2379:1 | PackageJsonInfo | 0.040547 | 0.044455 | +0.003908 | 1.0964 |
| src/compiler/types.ts:3199:1 | JsxOpeningLikeElement | 0.046200 | 0.050107 | +0.003907 | 1.0846 |
| src/compiler/builderPublic.ts:138:1 | SemanticDiagnosticsBuilderProgram | 0.045787 | 0.049685 | +0.003899 | 1.0852 |
| src/compiler/path.ts:781:1 | removeTrailingDirectorySeparator | 0.039144 | 0.043042 | +0.003898 | 1.0996 |
| src/compiler/utilitiesPublic.ts:1190:1 | getJSDocReturnTag | 0.053323 | 0.057219 | +0.003896 | 1.0731 |
| src/compiler/types.ts:4960:1 | ResolvedProjectReference | 0.038930 | 0.042822 | +0.003892 | 1.1000 |
| src/compiler/path.ts:622:1 | getNormalizedPathComponents | 0.039033 | 0.042923 | +0.003890 | 1.0997 |
| src/compiler/visitorPublic.ts:1:1 |  | 0.049609 | 0.053496 | +0.003887 | 1.0783 |
| src/compiler/types.ts:3049:1 | OptionalChainRoot | 0.046125 | 0.050004 | +0.003879 | 1.0841 |
| src/compiler/utilities.ts:1238:1 | getTokenPosOfNode | 0.128869 | 0.132747 | +0.003878 | 1.0301 |
| src/compiler/types.ts:2345:1 | MappedTypeNode | 0.043251 | 0.047127 | +0.003877 | 1.0896 |
| src/compiler/utilities.ts:10400:1 | skipTypeChecking | 0.039552 | 0.043424 | +0.003872 | 1.0979 |
| src/compiler/transformers/module/module.ts:174:1 | transformModule | 0.251898 | 0.255754 | +0.003857 | 1.0153 |
| src/compiler/factory/utilities.ts:362:1 | createExpressionForAccessorDeclaration | 0.044543 | 0.048394 | +0.003852 | 1.0865 |
| src/compiler/builder.ts:268:1 | isBuilderProgramStateWithDefinedProgram | 0.038527 | 0.042375 | +0.003848 | 1.0999 |
| src/compiler/utilities.ts:11082:1 | isJSDocOptionalParameter | 0.057583 | 0.061422 | +0.003839 | 1.0667 |
| src/compiler/utilities.ts:4561:1 | canHaveJSDoc | 0.044875 | 0.048711 | +0.003836 | 1.0855 |
| src/compiler/utilitiesPublic.ts:1570:1 | isStringTextContainingNode | 0.079376 | 0.083196 | +0.003820 | 1.0481 |
| src/compiler/moduleNameResolver.ts:1567:1 | tryLoadModuleUsingPathsIfEligible | 0.042943 | 0.046759 | +0.003816 | 1.0889 |
| src/compiler/factory/emitHelpers.ts:1369:1 |  | 0.086547 | 0.090361 | +0.003814 | 1.0441 |
| src/compiler/types.ts:10119:1 | ListFormat | 0.038301 | 0.042107 | +0.003806 | 1.0994 |
| src/compiler/utilities.ts:6140:1 | escapeTemplateSubstitution | 0.039959 | 0.043761 | +0.003802 | 1.0952 |
| src/compiler/types.ts:1640:1 | PrivateKeyword | 0.046234 | 0.050025 | +0.003791 | 1.0820 |
| src/compiler/path.ts:31:1 |  | 0.040784 | 0.044562 | +0.003777 | 1.0926 |
| src/compiler/factory/nodeTests.ts:576:1 | isArrayLiteralExpression | 0.042953 | 0.046725 | +0.003772 | 1.0878 |
| src/compiler/factory/emitNode.ts:147:1 | getTokenSourceMapRange | 0.051582 | 0.055351 | +0.003769 | 1.0731 |
| src/compiler/program.ts:1514:1 | createProgram | 0.070751 | 0.074507 | +0.003756 | 1.0531 |
| src/compiler/types.ts:4686:1 | ResolutionDiagnostics | 0.038627 | 0.042381 | +0.003754 | 1.0972 |
| src/compiler/factory/emitHelpers.ts:716:1 | helperString | 0.041111 | 0.044852 | +0.003741 | 1.0910 |
| src/compiler/performance.ts:202:1 | disable | 0.038137 | 0.041877 | +0.003740 | 1.0981 |
| src/compiler/program.ts:5139:1 | getResolutionDiagnostic | 0.152396 | 0.156127 | +0.003731 | 1.0245 |
| src/compiler/sourcemap.ts:437:1 | MappingsDecoder | 0.039840 | 0.043565 | +0.003725 | 1.0935 |
| src/compiler/utilities.ts:10186:1 | tryRemoveExtension | 0.039199 | 0.042922 | +0.003723 | 1.0950 |
| src/compiler/sys.ts:929:1 | isIgnoredByWatchOptions | 0.038833 | 0.042532 | +0.003698 | 1.0952 |
| src/compiler/transformers/module/module.ts:2499:1 |  | 0.067007 | 0.070687 | +0.003680 | 1.0549 |
| src/compiler/utilities.ts:3576:1 | classElementOrClassElementParameterIsDecorated | 0.043304 | 0.046982 | +0.003678 | 1.0849 |
| src/compiler/utilities.ts:7226:1 | hasAccessorModifier | 0.051885 | 0.055557 | +0.003672 | 1.0708 |
| src/compiler/types.ts:2967:1 | AccessExpression | 0.041400 | 0.045065 | +0.003665 | 1.0885 |
| src/compiler/utilities.ts:5297:1 | isIdentifierANonContextualKeyword | 0.058711 | 0.062374 | +0.003663 | 1.0624 |
| src/compiler/types.ts:2138:1 | GetAccessorDeclaration | 0.042658 | 0.046315 | +0.003656 | 1.0857 |
| src/compiler/utilitiesPublic.ts:2633:1 | hasRestParameter | 0.049400 | 0.053049 | +0.003649 | 1.0739 |
| src/compiler/builder.ts:133:1 | EmitSignature | 0.039465 | 0.043109 | +0.003644 | 1.0923 |
| src/compiler/executeCommandLine.ts:188:1 | createColors | 0.323821 | 0.327453 | +0.003633 | 1.0112 |
| src/compiler/utilitiesPublic.ts:1072:1 | getJSDocParameterTagsNoCache | 0.052228 | 0.055847 | +0.003620 | 1.0693 |
| src/compiler/utilities.ts:10417:1 | skipTypeCheckingWorker | 0.126284 | 0.129888 | +0.003604 | 1.0285 |
| src/compiler/types.ts:8015:1 | MinimalResolutionCacheHost | 0.039991 | 0.043594 | +0.003603 | 1.0901 |
| src/compiler/types.ts:4990:1 | EmitTransformers | 0.037813 | 0.041379 | +0.003566 | 1.0943 |
| src/compiler/utilities.ts:9653:1 |  | 0.066479 | 0.070043 | +0.003564 | 1.0536 |
| src/compiler/types.ts:2275:1 | TupleTypeNode | 0.040785 | 0.044346 | +0.003561 | 1.0873 |
| src/compiler/utilities.ts:9292:1 |  | 0.107791 | 0.111344 | +0.003553 | 1.0330 |
| src/compiler/types.ts:2996:1 | PropertyAccessEntityNameExpression | 0.042512 | 0.046063 | +0.003551 | 1.0835 |
| src/compiler/types.ts:7848:1 | CommandLineOption | 0.040266 | 0.043814 | +0.003548 | 1.0881 |
| src/compiler/transformers/destructuring.ts:244:1 | flattenDestructuringBinding | 0.070811 | 0.074353 | +0.003542 | 1.0500 |
| src/compiler/commandLineParser.ts:2442:1 | convertConfigFileToObject | 0.040063 | 0.043586 | +0.003523 | 1.0879 |
| src/compiler/utilitiesPublic.ts:1359:1 | isMemberName | 0.045194 | 0.048714 | +0.003519 | 1.0779 |
| src/compiler/types.ts:4066:1 | JSDocTypeTag | 0.053039 | 0.056556 | +0.003516 | 1.0663 |
| src/compiler/commandLineParser.ts:3906:1 |  | 0.042285 | 0.045795 | +0.003511 | 1.0830 |
| src/compiler/core.ts:678:1 | selectIndex | 0.041266 | 0.044775 | +0.003510 | 1.0851 |
| src/compiler/types.ts:5764:1 | ValidImportTypeNode | 0.038614 | 0.042114 | +0.003500 | 1.0906 |
| src/compiler/utilities.ts:2540:1 | getErrorSpanForNode | 0.166858 | 0.170352 | +0.003494 | 1.0209 |
| src/compiler/utilities.ts:10225:1 |  | 0.038993 | 0.042482 | +0.003488 | 1.0895 |
| src/compiler/transformers/classThis.ts:90:1 | classHasClassThisAssignment | 0.043612 | 0.047098 | +0.003486 | 1.0799 |
| src/compiler/utilities.ts:5286:1 | isNonContextualKeyword | 0.103877 | 0.107355 | +0.003477 | 1.0335 |
| src/compiler/utilitiesPublic.ts:1126:1 | getJSDocPublicTag | 0.055798 | 0.059274 | +0.003476 | 1.0623 |
| src/compiler/utilities.ts:2359:1 | isComputedNonLiteralName | 0.284013 | 0.287478 | +0.003466 | 1.0122 |
| src/compiler/performance.ts:153:1 | clearMarks | 0.078888 | 0.082352 | +0.003465 | 1.0439 |
| src/compiler/commandLineParser.ts:3043:1 | isNullOrUndefined | 0.038966 | 0.042428 | +0.003462 | 1.0888 |
| src/compiler/transformers/utilities.ts:130:1 | isNamedDefaultReference | 0.041631 | 0.045091 | +0.003460 | 1.0831 |
| src/compiler/utilities.ts:7344:1 | getSyntacticModifierFlagsNoCache | 0.271379 | 0.274835 | +0.003456 | 1.0127 |
| src/compiler/commandLineParser.ts:2306:1 | commandLineOptionsToMap | 0.041563 | 0.044995 | +0.003431 | 1.0826 |
| src/compiler/utilities.ts:6149:1 | hasInvalidEscape | 0.199202 | 0.202617 | +0.003416 | 1.0171 |
| src/compiler/emitter.ts:6359:1 | EmitFunction | 0.041525 | 0.044940 | +0.003415 | 1.0822 |
| src/compiler/transformers/namedEvaluation.ts:120:1 | ClassNamedEvaluationHelperBlock | 0.038946 | 0.042360 | +0.003414 | 1.0877 |
| src/compiler/watchUtilities.ts:835:1 | getFallbackOptions | 0.038406 | 0.041818 | +0.003412 | 1.0888 |
| src/compiler/factory/nodeTests.ts:600:1 | isTaggedTemplateExpression | 0.043570 | 0.046981 | +0.003411 | 1.0783 |
| src/compiler/emitter.ts:1203:1 |  | 0.041062 | 0.044466 | +0.003404 | 1.0829 |
| src/compiler/commandLineParser.ts:4175:1 | toCanonicalKey | 0.040182 | 0.043576 | +0.003393 | 1.0845 |
| src/compiler/utilities.ts:4837:1 | hasTypeArguments | 0.042946 | 0.046335 | +0.003389 | 1.0789 |
| src/compiler/factory/nodeTests.ts:1128:1 | isJSDocOverrideTag | 0.044209 | 0.047591 | +0.003382 | 1.0765 |
| src/compiler/utilities.ts:664:1 | isExternalModuleSymbol | 0.040407 | 0.043779 | +0.003372 | 1.0835 |
| src/compiler/types.ts:2591:1 | BitwiseOperatorOrHigher | 0.042850 | 0.046220 | +0.003369 | 1.0786 |
| src/compiler/factory/nodeTests.ts:1199:1 | isSyntaxList | 0.044328 | 0.047681 | +0.003352 | 1.0756 |
| src/compiler/types.ts:4533:1 | ScriptReferenceHost | 0.040436 | 0.043786 | +0.003350 | 1.0828 |
| src/compiler/types.ts:3100:1 | SuperCall | 0.039943 | 0.043289 | +0.003346 | 1.0838 |
| src/compiler/performance.ts:146:1 | clearMeasures | 0.077184 | 0.080526 | +0.003342 | 1.0433 |
| src/compiler/sys.ts:966:1 | CreateSystemWatchFunctions | 0.038440 | 0.041779 | +0.003339 | 1.0869 |
| src/compiler/utilitiesPublic.ts:1566:1 | isPartOfTypeOnlyImportOrExportDeclaration | 0.044655 | 0.047971 | +0.003316 | 1.0743 |
| src/compiler/utilities.ts:12294:1 | getOptionsSyntaxByArrayElementValue | 0.054400 | 0.057712 | +0.003312 | 1.0609 |
| src/compiler/commandLineParser.ts:141:1 |  | 0.041346 | 0.044651 | +0.003304 | 1.0799 |
| src/compiler/factory/nodeTests.ts:1116:1 | isJSDocPrivateTag | 0.042039 | 0.045335 | +0.003296 | 1.0784 |
| src/compiler/utilitiesPublic.ts:807:1 | isParseTreeNode | 0.045759 | 0.049043 | +0.003285 | 1.0718 |
| src/compiler/types.ts:4589:1 | OperationCanceledException | 0.041727 | 0.045008 | +0.003281 | 1.0786 |
| src/compiler/resolutionCache.ts:576:1 | GetResolutionWithResolvedFileName | 0.041956 | 0.045232 | +0.003277 | 1.0781 |
| src/compiler/utilitiesPublic.ts:1849:1 | isArrayBindingElement | 0.045388 | 0.048663 | +0.003275 | 1.0722 |
| src/compiler/types.ts:3834:1 | TypeOnlyImportDeclaration | 0.038991 | 0.042263 | +0.003272 | 1.0839 |
| src/compiler/types.ts:4317:1 | ResolutionMode | 0.037331 | 0.040582 | +0.003252 | 1.0871 |
| src/compiler/types.ts:4975:1 | CustomTransformer | 0.037654 | 0.040902 | +0.003248 | 1.0863 |
| src/compiler/types.ts:5808:1 | SymbolVisibilityResult | 0.037547 | 0.040791 | +0.003243 | 1.0864 |
| src/compiler/builder.ts:1145:1 | IncrementalMultiFileEmitBuildInfo | 0.040408 | 0.043648 | +0.003241 | 1.0802 |
| src/compiler/utilitiesPublic.ts:931:1 | nodeHasName | 0.045924 | 0.049164 | +0.003240 | 1.0705 |
| src/compiler/types.ts:3822:1 | ImportOrExportSpecifier | 0.039048 | 0.042285 | +0.003237 | 1.0829 |
| src/compiler/factory/nodeTests.ts:1124:1 | isJSDocReadonlyTag | 0.045147 | 0.048381 | +0.003234 | 1.0716 |
| src/compiler/utilities.ts:3188:1 | getThisContainer | 0.039714 | 0.042947 | +0.003232 | 1.0814 |
| src/compiler/builder.ts:744:1 | getBuilderFileEmitAllDts | 0.044648 | 0.047881 | +0.003232 | 1.0724 |
| src/compiler/types.ts:6841:1 | IterableOrIteratorType | 0.042340 | 0.045568 | +0.003227 | 1.0762 |
| src/compiler/types.ts:4267:1 | FlowType | 0.037059 | 0.040285 | +0.003226 | 1.0870 |
| src/compiler/_namespaces/ts.ts:30:1 |  | 0.039805 | 0.043024 | +0.003219 | 1.0809 |
| src/compiler/factory/parenthesizerRules.ts:675:1 |  | 0.117605 | 0.120820 | +0.003215 | 1.0273 |
| src/compiler/utilitiesPublic.ts:1195:1 | getJSDocTemplateTag | 0.054995 | 0.058209 | +0.003214 | 1.0584 |
| src/compiler/utilities.ts:3114:1 | getTsConfigPropArrayElementValue | 0.068543 | 0.071744 | +0.003202 | 1.0467 |
| src/compiler/utilities.ts:5309:1 | FunctionFlags | 0.065690 | 0.068890 | +0.003200 | 1.0487 |
| src/compiler/utilities.ts:3078:1 | isObjectLiteralOrClassExpressionMethodOrAccessor | 0.048524 | 0.051721 | +0.003197 | 1.0659 |
| src/compiler/sys.ts:963:1 | FileSystemEntryExists | 0.039312 | 0.042501 | +0.003189 | 1.0811 |
| src/compiler/moduleSpecifiers.ts:1357:1 | getPathsRelativeToRootDirs | 0.041983 | 0.045168 | +0.003185 | 1.0759 |
| src/compiler/utilities.ts:10918:1 | escapeSnippetText | 0.038904 | 0.042082 | +0.003178 | 1.0817 |
| src/compiler/utilities.ts:3604:1 | isEmptyStringLiteral | 0.042340 | 0.045513 | +0.003173 | 1.0750 |
| src/compiler/utilities.ts:9600:1 | regExpEscape | 0.087208 | 0.090379 | +0.003171 | 1.0364 |
| src/compiler/types.ts:3146:1 | TypeAssertion | 0.042399 | 0.045564 | +0.003164 | 1.0746 |
| src/compiler/factory/nodeTests.ts:664:1 | isClassExpression | 0.042497 | 0.045657 | +0.003160 | 1.0744 |
| src/compiler/resolutionCache.ts:258:1 | removeIgnoredPath | 0.069211 | 0.072370 | +0.003159 | 1.0456 |
| src/compiler/types.ts:10207:1 | PragmaKindFlags | 0.038695 | 0.041845 | +0.003149 | 1.0814 |
| src/compiler/utilities.ts:12306:1 | getOptionsSyntaxByValue | 0.066876 | 0.070020 | +0.003144 | 1.0470 |
| src/compiler/factory/emitHelpers.ts:958:1 |  | 0.072676 | 0.075819 | +0.003143 | 1.0432 |
| src/compiler/_namespaces/ts.ts:34:1 |  | 0.039529 | 0.042668 | +0.003139 | 1.0794 |
| src/compiler/types.ts:3108:1 | ExpressionWithTypeArguments | 0.038983 | 0.042118 | +0.003135 | 1.0804 |
| src/compiler/utilities.ts:2685:1 | isVarConstLike | 0.044575 | 0.047708 | +0.003133 | 1.0703 |
| src/compiler/types.ts:7359:1 | ModuleDetectionKind | 0.039577 | 0.042707 | +0.003131 | 1.0791 |
| src/compiler/sys.ts:1393:1 | BufferEncoding | 0.036985 | 0.040111 | +0.003126 | 1.0845 |
| src/compiler/utilities.ts:10582:1 | isValidTypeOnlyAliasUseSite | 0.062597 | 0.065717 | +0.003121 | 1.0499 |
| src/compiler/types.ts:2719:1 | AssignmentElementTarget | 0.043114 | 0.046234 | +0.003120 | 1.0724 |
| src/compiler/resolutionCache.ts:249:1 | DirectoryOfFailedLookupWatch | 0.040059 | 0.043178 | +0.003119 | 1.0779 |
| src/compiler/commandLineParser.ts:2440:1 | AdamicJsonRecoveryObject | 0.041314 | 0.044421 | +0.003107 | 1.0752 |
| src/compiler/commandLineParser.ts:637:1 |  | 0.068934 | 0.072038 | +0.003103 | 1.0450 |
| src/compiler/commandLineParser.ts:1692:1 |  | 0.039473 | 0.042569 | +0.003096 | 1.0784 |
| src/compiler/path.ts:732:1 | simpleNormalizePath | 0.040122 | 0.043215 | +0.003093 | 1.0771 |
| src/compiler/utilities.ts:3152:1 | getContainingFunctionOrClassStaticBlock | 0.046546 | 0.049637 | +0.003090 | 1.0664 |
| src/compiler/utilities.ts:6500:1 | getResolvedExternalModuleName | 0.042005 | 0.045091 | +0.003087 | 1.0735 |
| src/compiler/factory/nodeTests.ts:660:1 | isSpreadElement | 0.043099 | 0.046183 | +0.003084 | 1.0716 |
| src/compiler/transformers/namedEvaluation.ts:354:1 | transformNamedEvaluationOfBindingElement | 0.041154 | 0.044236 | +0.003082 | 1.0749 |
| src/compiler/utilitiesPublic.ts:2625:1 | isStringLiteralLike | 0.044048 | 0.047124 | +0.003076 | 1.0698 |
| src/compiler/utilities.ts:11119:1 | getJSDocSatisfiesExpressionType | 0.052011 | 0.055086 | +0.003075 | 1.0591 |
| src/compiler/sys.ts:902:1 | FileSystemEntryKind | 0.039064 | 0.042138 | +0.003073 | 1.0787 |
| src/compiler/types.ts:2026:1 | VariableLikeDeclaration | 0.038356 | 0.041428 | +0.003072 | 1.0801 |
| src/compiler/factory/emitHelpers.ts:1151:1 |  | 0.075588 | 0.078655 | +0.003067 | 1.0406 |
| src/compiler/transformer.ts:192:1 | getDeclarationTransformers | 0.038911 | 0.041977 | +0.003066 | 1.0788 |
| src/compiler/transformers/es2015.ts:283:1 | Jump | 0.041403 | 0.044465 | +0.003063 | 1.0740 |
| src/compiler/utilitiesPublic.ts:941:1 | getNameOfJSDocTypedef | 0.040597 | 0.043659 | +0.003062 | 1.0754 |
| src/compiler/commandLineParser.ts:3053:1 |  | 0.037698 | 0.040752 | +0.003054 | 1.0810 |
| src/compiler/moduleNameResolver.ts:942:1 | PerModuleNameCache | 0.040908 | 0.043962 | +0.003054 | 1.0746 |
| src/compiler/utilities.ts:6046:1 | getSemanticJsxChildren | 0.044124 | 0.047175 | +0.003051 | 1.0691 |
| src/compiler/utilities.ts:3085:1 | isIdentifierTypePredicate | 0.039425 | 0.042476 | +0.003051 | 1.0774 |
| src/compiler/watchUtilities.ts:668:1 | isEmittedFileOfProgram | 0.081980 | 0.085025 | +0.003045 | 1.0371 |
| src/compiler/sys.ts:907:1 | createFileWatcherCallback | 0.039215 | 0.042246 | +0.003030 | 1.0773 |
| src/compiler/builder.ts:521:1 | repopulateDiagnostics | 0.040329 | 0.043353 | +0.003024 | 1.0750 |
| src/compiler/factory/nodeTests.ts:910:1 | isNotEmittedStatement | 0.046359 | 0.049373 | +0.003014 | 1.0650 |
| src/compiler/_namespaces/ts.ts:32:1 |  | 0.039481 | 0.042493 | +0.003012 | 1.0763 |
| src/compiler/visitorPublic.ts:1805:1 | extractSingleNode | 0.056300 | 0.059307 | +0.003007 | 1.0534 |
| src/compiler/builder.ts:271:1 | toBuilderProgramStateWithDefinedProgram | 0.039490 | 0.042491 | +0.003001 | 1.0760 |
| src/compiler/path.ts:41:1 | isAnyDirectorySeparator | 0.038593 | 0.041585 | +0.002991 | 1.0775 |
| src/compiler/types.ts:8137:1 | HasInvalidatedLibResolutions | 0.039823 | 0.042804 | +0.002981 | 1.0748 |
| src/compiler/utilities.ts:720:1 | changesAffectingProgramStructure | 0.040969 | 0.043945 | +0.002976 | 1.0726 |
| src/compiler/types.ts:4320:1 | SourceFile | 0.037738 | 0.040714 | +0.002976 | 1.0789 |
| src/compiler/emitter.ts:1200:1 |  | 0.041273 | 0.044226 | +0.002953 | 1.0716 |
| src/compiler/utilities.ts:5120:1 | isIdentifierName | 0.040539 | 0.043492 | +0.002953 | 1.0728 |
| src/compiler/programDiagnostics.ts:65:1 | LazyConfigDiagnostic | 0.040493 | 0.043443 | +0.002950 | 1.0729 |
| src/compiler/transformers/utilities.ts:107:1 | getOriginalNodeId | 0.050189 | 0.053126 | +0.002936 | 1.0585 |
| src/compiler/resolutionCache.ts:184:1 | ResolutionCacheHost | 0.041541 | 0.044473 | +0.002932 | 1.0706 |
| src/compiler/utilities.ts:4925:1 | isAssignmentTarget | 0.049260 | 0.052192 | +0.002932 | 1.0595 |
| src/compiler/moduleNameResolver.ts:1678:1 | resolveJSModule | 0.121860 | 0.124773 | +0.002913 | 1.0239 |
| src/compiler/factory/nodeTests.ts:696:1 | isPartiallyEmittedExpression | 0.043362 | 0.046267 | +0.002905 | 1.0670 |
| src/compiler/types.ts:2954:1 | ObjectLiteralExpressionBase | 0.040513 | 0.043418 | +0.002904 | 1.0717 |
| src/compiler/semver.ts:46:1 | Version | 0.237081 | 0.239985 | +0.002903 | 1.0122 |
| src/compiler/factory/nodeTests.ts:243:1 | isStringLiteral | 0.047052 | 0.049951 | +0.002899 | 1.0616 |
| src/compiler/types.ts:5738:1 | AnyImportOrBareOrAccessedRequire | 0.038810 | 0.041701 | +0.002891 | 1.0745 |
| src/compiler/tsbuild.ts:58:1 | Status | 0.038462 | 0.041352 | +0.002890 | 1.0751 |
| src/compiler/factory/nodeTests.ts:1176:1 | isJSDocPropertyTag | 0.043806 | 0.046695 | +0.002889 | 1.0660 |
| src/compiler/utilities.ts:5043:1 | skipParentheses | 0.038738 | 0.041620 | +0.002883 | 1.0744 |
| src/compiler/emitter.ts:6351:1 | OrdinalParentheizerRuleSelector | 0.044830 | 0.047712 | +0.002882 | 1.0643 |
| src/compiler/types.ts:2959:1 | ObjectLiteralExpression | 0.043799 | 0.046672 | +0.002873 | 1.0656 |
| src/compiler/utilities.ts:6163:1 |  | 0.039123 | 0.041994 | +0.002871 | 1.0734 |
| src/compiler/types.ts:2020:1 | SpreadAssignment | 0.038857 | 0.041725 | +0.002868 | 1.0738 |
| src/compiler/transformers/declarations.ts:241:1 |  | 0.038227 | 0.041090 | +0.002863 | 1.0749 |
| src/compiler/factory/nodeFactory.ts:7414:1 | createSourceMapSource | 0.040911 | 0.043763 | +0.002852 | 1.0697 |
| src/compiler/utilitiesPublic.ts:1199:1 | getJSDocSatisfiesTag | 0.054044 | 0.056886 | +0.002841 | 1.0526 |
| src/compiler/utilities.ts:618:1 |  | 0.039642 | 0.042481 | +0.002839 | 1.0716 |
| src/compiler/transformers/utilities.ts:703:1 | getAllDecoratorsOfClassElement | 0.042645 | 0.045469 | +0.002825 | 1.0662 |
| src/compiler/utilities.ts:10548:1 | parseValidBigInt | 0.038185 | 0.041003 | +0.002817 | 1.0738 |
| src/compiler/executeCommandLine.ts:420:1 | generateGroupOptionOutput | 0.043881 | 0.046689 | +0.002808 | 1.0640 |
| src/compiler/utilities.ts:7505:1 | getFirstIdentifier | 0.043666 | 0.046454 | +0.002788 | 1.0638 |
| src/compiler/executeCommandLine.ts:1160:1 | reportStatistics | 0.287897 | 0.290683 | +0.002787 | 1.0097 |
| src/compiler/types.ts:4273:1 | IncompleteType | 0.037961 | 0.040747 | +0.002786 | 1.0734 |
| src/compiler/utilities.ts:7332:1 | getEffectiveModifierFlagsNoCache | 0.052401 | 0.055185 | +0.002784 | 1.0531 |
| src/compiler/factory/nodeTests.ts:648:1 | isConditionalExpression | 0.042537 | 0.045320 | +0.002783 | 1.0654 |
| src/compiler/factory/nodeFactory.ts:7392:1 |  | 0.040223 | 0.043001 | +0.002778 | 1.0691 |
| src/compiler/types.ts:7389:1 | WatchFileKind | 0.041049 | 0.043827 | +0.002778 | 1.0677 |
| src/compiler/path.ts:785:1 | removeTrailingDirectorySeparator | 0.040221 | 0.042998 | +0.002776 | 1.0690 |
| src/compiler/utilities.ts:9784:1 | FileSystemEntries | 0.038154 | 0.040914 | +0.002760 | 1.0723 |
| src/compiler/resolutionCache.ts:243:1 | PackageDirWatcher | 0.040981 | 0.043738 | +0.002756 | 1.0673 |
| src/compiler/utilities.ts:4964:1 | isNodeWithPossibleHoistedDeclaration | 0.043519 | 0.046259 | +0.002741 | 1.0630 |
| src/compiler/factory/utilities.ts:1267:1 | isBinaryOperatorToken | 0.078522 | 0.081260 | +0.002738 | 1.0349 |
| src/compiler/utilities.ts:2983:1 | getMembersOfDeclaration | 0.040898 | 0.043610 | +0.002713 | 1.0663 |
| src/compiler/commandLineParser.ts:3037:1 | setConfigFileInOptions | 0.042198 | 0.044905 | +0.002707 | 1.0642 |
| src/compiler/transformers/classFields.ts:320:1 | ClassLexicalEnvironment | 0.038436 | 0.041123 | +0.002687 | 1.0699 |
| src/compiler/factory/emitHelpers.ts:730:1 |  | 0.073270 | 0.075942 | +0.002671 | 1.0365 |
| src/compiler/utilities.ts:3539:1 | nodeOrChildIsDecorated | 0.068387 | 0.071056 | +0.002669 | 1.0390 |
| src/compiler/performance.ts:136:1 | forEachMeasure | 0.039681 | 0.042347 | +0.002666 | 1.0672 |
| src/compiler/factory/nodeTests.ts:1132:1 | isJSDocOverloadTag | 0.045258 | 0.047917 | +0.002659 | 1.0588 |
| src/compiler/program.ts:657:1 | formatDiagnostics | 0.039840 | 0.042499 | +0.002659 | 1.0667 |
| src/compiler/types.ts:5769:1 | AnyValidImportOrReExport | 0.037853 | 0.040500 | +0.002647 | 1.0699 |
| src/compiler/factory/utilities.ts:1008:1 | getPropertyNameOfBindingOrAssignmentElement | 0.323117 | 0.325764 | +0.002647 | 1.0082 |
| src/compiler/types.ts:8323:1 | EmitNode | 0.040183 | 0.042830 | +0.002647 | 1.0659 |
| src/compiler/parser.ts:505:1 |  | 0.070987 | 0.073631 | +0.002644 | 1.0372 |
| src/compiler/tsbuildPublic.ts:136:1 | BuildOptions | 0.041435 | 0.044079 | +0.002644 | 1.0638 |
| src/compiler/types.ts:4307:1 | RedirectInfo | 0.037175 | 0.039818 | +0.002643 | 1.0711 |
| src/compiler/utilities.ts:10532:1 | pseudoBigIntToString | 0.119221 | 0.121861 | +0.002640 | 1.0221 |
| src/compiler/moduleNameResolver.ts:2280:1 | loadEntrypointsFromExportMap | 0.040898 | 0.043524 | +0.002626 | 1.0642 |
| src/compiler/utilities.ts:615:1 |  | 0.039305 | 0.041930 | +0.002626 | 1.0668 |
| src/compiler/builder.ts:805:1 | handleDtsMayChangeOfAffectedFile | 0.039361 | 0.041986 | +0.002625 | 1.0667 |
| src/compiler/types.ts:3207:1 | JsxAttributeLike | 0.042531 | 0.045153 | +0.002622 | 1.0617 |
| src/compiler/types.ts:6481:1 | FreshableType | 0.039016 | 0.041634 | +0.002617 | 1.0671 |
| src/compiler/utilities.ts:9409:1 | getJSXTransformEnabled | 0.065398 | 0.068004 | +0.002606 | 1.0399 |
| src/compiler/types.ts:4483:1 | CommentDirective | 0.038630 | 0.041228 | +0.002598 | 1.0672 |
| src/compiler/_namespaces/ts.ts:27:1 |  | 0.040037 | 0.042632 | +0.002595 | 1.0648 |
| src/compiler/utilities.ts:10382:1 | rangeOfNode | 0.049557 | 0.052152 | +0.002595 | 1.0524 |
| src/compiler/types.ts:2280:1 | NamedTupleMember | 0.040100 | 0.042692 | +0.002592 | 1.0646 |
| src/compiler/factory/nodeTests.ts:1:1 |  | 0.041057 | 0.043648 | +0.002591 | 1.0631 |
| src/compiler/utilities.ts:2677:1 | isVarConst | 0.044890 | 0.047480 | +0.002590 | 1.0577 |
| src/compiler/types.ts:5011:1 | SourceMapEmitResult | 0.037741 | 0.040323 | +0.002583 | 1.0684 |
| src/compiler/types.ts:6694:1 | GenericType | 0.038091 | 0.040673 | +0.002582 | 1.0678 |
| src/compiler/commandLineParser.ts:2214:1 | ConfigFileDiagnosticsReporter | 0.039952 | 0.042533 | +0.002581 | 1.0646 |
| src/compiler/utilities.ts:4842:1 | AssignmentKind | 0.038555 | 0.041136 | +0.002581 | 1.0669 |
| src/compiler/utilities.ts:9289:1 |  | 0.065966 | 0.068539 | +0.002573 | 1.0390 |
| src/compiler/types.ts:2690:1 | ArrayBindingOrAssignmentElement | 0.042196 | 0.044761 | +0.002564 | 1.0608 |
| src/compiler/tsbuild.ts:177:1 | resolveConfigFileProjectName | 0.037304 | 0.039867 | +0.002563 | 1.0687 |
| src/compiler/utilities.ts:4684:1 | filterOwnedJSDocTags | 0.043586 | 0.046148 | +0.002562 | 1.0588 |
| src/compiler/transformers/es2015.ts:278:1 | CopyDirection | 0.041408 | 0.043969 | +0.002561 | 1.0619 |
| src/compiler/_namespaces/ts.ts:59:1 |  | 0.039617 | 0.042178 | +0.002561 | 1.0646 |
| src/compiler/executeCommandLine.ts:1289:1 | writeConfigFile | 0.191848 | 0.194405 | +0.002557 | 1.0133 |
| src/compiler/emitter.ts:506:1 | getOutputPathsForBundle | 0.039851 | 0.042407 | +0.002556 | 1.0641 |
| src/compiler/scanner.ts:4041:1 | charSize | 0.042423 | 0.044967 | +0.002544 | 1.0600 |
| src/compiler/utilitiesPublic.ts:1170:1 | getJSDocDeprecatedTag | 0.055754 | 0.058295 | +0.002541 | 1.0456 |
| src/compiler/types.ts:3301:1 | JsxExpression | 0.039795 | 0.042332 | +0.002537 | 1.0638 |
| src/compiler/types.ts:3372:1 | Block | 0.043208 | 0.045743 | +0.002535 | 1.0587 |
| src/compiler/transformers/classThis.ts:52:1 | ClassThisAssignmentBlock | 0.038168 | 0.040689 | +0.002521 | 1.0661 |
| src/compiler/utilities.ts:10923:1 | isNumericLiteralName | 0.038905 | 0.041420 | +0.002515 | 1.0646 |
| src/compiler/utilities.ts:7425:1 | isLogicalOrCoalescingBinaryOperator | 0.051110 | 0.053624 | +0.002513 | 1.0492 |
| src/compiler/moduleSpecifiers.ts:879:1 | tryGetModuleNameFromAmbientModule | 0.040205 | 0.042708 | +0.002504 | 1.0623 |
| src/compiler/program.ts:335:1 | resolveTripleslashReference | 0.039754 | 0.042252 | +0.002497 | 1.0628 |
| src/compiler/types.ts:4973:1 | CustomTransformerFactory | 0.038922 | 0.041416 | +0.002494 | 1.0641 |
| src/compiler/types.ts:1608:1 | ExclamationToken | 0.052397 | 0.054890 | +0.002493 | 1.0476 |
| src/compiler/types.ts:4502:1 | JsonSourceFile | 0.038975 | 0.041467 | +0.002492 | 1.0639 |
| src/compiler/path.ts:388:1 | tryGetExtensionFromPath | 0.040073 | 0.042565 | +0.002492 | 1.0622 |
| src/compiler/transformers/es2018.ts:115:1 | ESNextSubstitutionFlags | 0.038018 | 0.040509 | +0.002491 | 1.0655 |
| src/compiler/utilities.ts:11484:1 | NameResolver | 0.040348 | 0.042839 | +0.002491 | 1.0617 |
| src/compiler/utilities.ts:8240:1 | MutateMapOptions | 0.040472 | 0.042961 | +0.002489 | 1.0615 |
| src/compiler/types.ts:3804:1 | ImportSpecifier | 0.039465 | 0.041952 | +0.002487 | 1.0630 |
| src/compiler/performance.ts:141:1 | forEachMark | 0.038863 | 0.041339 | +0.002476 | 1.0637 |
| src/compiler/factory/nodeTests.ts:1172:1 | isJSDocUnknownTag | 0.044397 | 0.046861 | +0.002464 | 1.0555 |
| src/compiler/utilities.ts:10327:1 | matchPatternOrExact | 0.122969 | 0.125433 | +0.002464 | 1.0200 |
| src/compiler/types.ts:3060:1 | BindableStaticNameExpression | 0.045049 | 0.047512 | +0.002463 | 1.0547 |
| src/compiler/types.ts:2163:1 | AccessorDeclaration | 0.042826 | 0.045285 | +0.002459 | 1.0574 |
| src/compiler/path.ts:762:1 | toPath | 0.039585 | 0.042041 | +0.002456 | 1.0620 |
| src/compiler/resolutionCache.ts:330:1 | isInDirectoryPath | 0.042633 | 0.045086 | +0.002453 | 1.0575 |
| src/compiler/checker.ts:1422:1 | DeclarationSpaces | 0.039734 | 0.042187 | +0.002453 | 1.0617 |
| src/compiler/types.ts:5693:1 | TypePredicateBase | 0.037481 | 0.039932 | +0.002451 | 1.0654 |
| src/compiler/semver.ts:276:1 | parseRange | 0.039556 | 0.042003 | +0.002447 | 1.0619 |
| src/compiler/path.ts:805:1 | ensureTrailingDirectorySeparator | 0.039217 | 0.041664 | +0.002447 | 1.0624 |
| src/compiler/sys.ts:96:1 | PollingInterval | 0.038463 | 0.040906 | +0.002443 | 1.0635 |
| src/compiler/types.ts:5794:1 | CanHaveModuleSpecifier | 0.037810 | 0.040252 | +0.002442 | 1.0646 |
| src/compiler/sys.ts:1454:1 | FileWatcher | 0.039325 | 0.041763 | +0.002438 | 1.0620 |
| src/compiler/performance.ts:179:1 | enable | 0.039573 | 0.042007 | +0.002433 | 1.0615 |
| src/compiler/factory/utilities.ts:1181:1 | isMultiplicativeOperatorOrHigher | 0.377566 | 0.379991 | +0.002425 | 1.0064 |
| src/compiler/utilities.ts:612:1 |  | 0.039318 | 0.041740 | +0.002422 | 1.0616 |
| src/compiler/types.ts:9751:1 | NodesVisitor | 0.039295 | 0.041716 | +0.002422 | 1.0616 |
| src/compiler/utilities.ts:7950:1 | getLinesBetweenRangeEndPositions | 0.039120 | 0.041540 | +0.002420 | 1.0619 |
| src/compiler/moduleNameResolver.ts:3091:1 | loadModuleFromSpecificNodeModulesDirectory | 0.116819 | 0.119238 | +0.002419 | 1.0207 |
| src/compiler/factory/emitHelpers.ts:1467:1 |  | 0.097617 | 0.100024 | +0.002408 | 1.0247 |
| src/compiler/factory/nodeTests.ts:596:1 | isNewExpression | 0.042495 | 0.044899 | +0.002404 | 1.0566 |
| src/compiler/path.ts:722:1 | normalizePath | 0.039219 | 0.041623 | +0.002404 | 1.0613 |
| src/compiler/builder.ts:126:1 | ReusableDiagnosticMessageChain | 0.038355 | 0.040752 | +0.002398 | 1.0625 |
| src/compiler/utilitiesPublic.ts:1748:1 | isAccessor | 0.045194 | 0.047592 | +0.002397 | 1.0530 |
| src/compiler/path.ts:853:1 | changeAnyExtension | 0.043000 | 0.045389 | +0.002390 | 1.0556 |
| src/compiler/parser.ts:422:1 |  | 0.039828 | 0.042206 | +0.002378 | 1.0597 |
| src/compiler/commandLineParser.ts:2627:1 | TSConfig | 0.038768 | 0.041145 | +0.002377 | 1.0613 |
| src/compiler/types.ts:2794:1 | RegularExpressionFlags | 0.039760 | 0.042137 | +0.002377 | 1.0598 |
| src/compiler/executeCommandLine.ts:523:1 | printBuildHelp | 0.043101 | 0.045474 | +0.002373 | 1.0551 |
| src/compiler/sys.ts:108:1 |  | 0.038093 | 0.040463 | +0.002370 | 1.0622 |
| src/compiler/types.ts:3128:1 | InstanceofExpression | 0.039680 | 0.042048 | +0.002368 | 1.0597 |
| src/compiler/moduleNameResolver.ts:916:1 | MissingPackageJsonInfo | 0.039636 | 0.042003 | +0.002367 | 1.0597 |
| src/compiler/transformers/namedEvaluation.ts:101:1 | createClassNamedEvaluationHelperBlock | 0.038979 | 0.041343 | +0.002364 | 1.0607 |
| src/compiler/sys.ts:89:1 | WatchedFile | 0.039560 | 0.041916 | +0.002356 | 1.0596 |
| src/compiler/moduleNameResolver.ts:1115:1 | createModeAwareCacheKey | 0.039790 | 0.042137 | +0.002347 | 1.0590 |
| src/compiler/program.ts:707:1 | formatColorAndReset | 0.040759 | 0.043103 | +0.002344 | 1.0575 |
| src/compiler/types.ts:4529:1 | JsonObjectExpressionStatement | 0.043104 | 0.045439 | +0.002335 | 1.0542 |
| src/compiler/utilities.ts:655:1 | isTransientSymbol | 0.040369 | 0.042700 | +0.002332 | 1.0578 |
| src/compiler/transformer.ts:127:1 | getScriptTransformers | 0.038740 | 0.041061 | +0.002322 | 1.0599 |
| src/compiler/path.ts:627:1 | getNormalizedAbsolutePath | 0.040202 | 0.042522 | +0.002321 | 1.0577 |
| src/compiler/utilities.ts:1341:1 | getPos | 0.043511 | 0.045827 | +0.002317 | 1.0532 |
| src/compiler/types.ts:2876:1 | TemplateHead | 0.042410 | 0.044726 | +0.002317 | 1.0546 |
| src/compiler/utilities.ts:1360:1 | getEmitFlags | 0.043302 | 0.045614 | +0.002313 | 1.0534 |
| src/compiler/program.ts:635:18 | getPreEmitDiagnostics | 0.042010 | 0.044318 | +0.002308 | 1.0549 |
| src/compiler/tsbuildPublic.ts:132:1 |  | 0.040086 | 0.042388 | +0.002302 | 1.0574 |
| src/compiler/parser.ts:403:1 |  | 0.048699 | 0.050996 | +0.002298 | 1.0472 |
| src/compiler/builder.ts:1120:1 | IncrementalBuildInfoRootStartEnd | 0.041423 | 0.043717 | +0.002294 | 1.0554 |
| src/compiler/performance.ts:61:1 |  | 0.038371 | 0.040662 | +0.002291 | 1.0597 |
| src/compiler/utilitiesPublic.ts:1146:1 | getJSDocProtectedTag | 0.055362 | 0.057650 | +0.002288 | 1.0413 |
| src/compiler/types.ts:1762:1 | NamedDeclaration | 0.052244 | 0.054530 | +0.002286 | 1.0438 |
| src/compiler/path.ts:421:1 | getAnyExtensionFromPath | 0.039659 | 0.041936 | +0.002277 | 1.0574 |
| src/compiler/moduleNameResolver.ts:2384:1 | PackageJsonInfoContents | 0.039757 | 0.042027 | +0.002270 | 1.0571 |
| src/compiler/types.ts:3002:1 | ElementAccessExpression | 0.040415 | 0.042684 | +0.002269 | 1.0562 |
| src/compiler/_namespaces/ts.ts:33:1 |  | 0.039720 | 0.041983 | +0.002264 | 1.0570 |
| src/compiler/utilities.ts:6953:1 | emitNewLineBeforeLeadingCommentOfPosition | 0.081263 | 0.083526 | +0.002263 | 1.0278 |
| src/compiler/commandLineParser.ts:2611:1 | isCompilerOptionsValue | 0.039942 | 0.042205 | +0.002263 | 1.0567 |
| src/compiler/types.ts:3812:1 | ExportSpecifier | 0.040554 | 0.042817 | +0.002263 | 1.0558 |
| src/compiler/path.ts:165:1 | getEncodedRootLength | 0.039880 | 0.042141 | +0.002261 | 1.0567 |
| src/compiler/factory/nodeFactory.ts:7409:1 |  | 0.040115 | 0.042373 | +0.002258 | 1.0563 |
| src/compiler/performance.ts:74:1 | mark | 0.077708 | 0.079960 | +0.002251 | 1.0290 |
| src/compiler/types.ts:4278:1 | AmdDependency | 0.038243 | 0.040492 | +0.002249 | 1.0588 |
| src/compiler/transformers/declarations/diagnostics.ts:104:1 | SymbolAccessibilityDiagnostic | 0.039352 | 0.041598 | +0.002245 | 1.0571 |
| src/compiler/performanceCore.ts:7:1 | PerformanceHooks | 0.038644 | 0.040879 | +0.002234 | 1.0578 |
| src/compiler/path.ts:250:1 | getRootLength | 0.039018 | 0.041251 | +0.002233 | 1.0572 |
| src/compiler/path.ts:783:1 | removeTrailingDirectorySeparator | 0.039209 | 0.041429 | +0.002220 | 1.0566 |
| src/compiler/factory/nodeFactory.ts:7309:1 | aggregateChildrenFlags | 0.041606 | 0.043822 | +0.002217 | 1.0533 |
| src/compiler/path.ts:457:1 | PathPathComponents | 0.037974 | 0.040187 | +0.002213 | 1.0583 |
| src/compiler/path.ts:495:1 | getPathComponents | 0.038488 | 0.040691 | +0.002203 | 1.0573 |
| src/compiler/types.ts:5818:1 | SymbolAccessibilityResult | 0.038010 | 0.040213 | +0.002203 | 1.0580 |
| src/compiler/path.ts:398:1 | getAnyExtensionFromPathWorker | 0.039809 | 0.042006 | +0.002197 | 1.0552 |
| src/compiler/utilities.ts:1337:1 | getTextOfNode | 0.050883 | 0.053077 | +0.002194 | 1.0431 |
| src/compiler/types.ts:10331:1 | UnionToIntersection | 0.038311 | 0.040505 | +0.002194 | 1.0573 |
| src/compiler/types.ts:7299:1 | DiagnosticAndArguments | 0.039515 | 0.041708 | +0.002193 | 1.0555 |
| src/compiler/utilities.ts:10173:1 |  | 0.039506 | 0.041699 | +0.002193 | 1.0555 |
| src/compiler/types.ts:4295:1 | FutureSourceFile | 0.037844 | 0.040031 | +0.002188 | 1.0578 |
| src/compiler/sys.ts:105:1 | HostWatchDirectory | 0.040160 | 0.042341 | +0.002181 | 1.0543 |
| src/compiler/path.ts:310:1 | getDirectoryPath | 0.041974 | 0.044155 | +0.002181 | 1.0520 |
| src/compiler/utilitiesPublic.ts:1274:1 | getJSDocTags | 0.047788 | 0.049958 | +0.002170 | 1.0454 |
| src/compiler/program.ts:813:1 | flattenDiagnosticMessageText | 0.076460 | 0.078629 | +0.002168 | 1.0284 |
| src/compiler/types.ts:6702:1 | ElementFlags | 0.038095 | 0.040261 | +0.002165 | 1.0568 |
| src/compiler/factory/nodeTests.ts:580:1 | isObjectLiteralExpression | 0.042984 | 0.045149 | +0.002164 | 1.0504 |
| src/compiler/factory/nodeTests.ts:1184:1 | isJSDocSatisfiesTag | 0.044425 | 0.046588 | +0.002163 | 1.0487 |
| src/compiler/utilities.ts:5054:1 | isDeleteTarget | 0.054147 | 0.056303 | +0.002155 | 1.0398 |
| src/compiler/types.ts:3910:1 | JSDocUnknownType | 0.045768 | 0.047916 | +0.002148 | 1.0469 |
| src/compiler/utilities.ts:10166:1 | compareNumberOfDirectorySeparators | 0.038050 | 0.040195 | +0.002145 | 1.0564 |
| src/compiler/utilities.ts:8593:1 | setLocalizedDiagnosticMessages | 0.040481 | 0.042622 | +0.002141 | 1.0529 |
| src/compiler/factory/nodeChildren.ts:55:1 | transferSourceFileChildren | 0.038327 | 0.040466 | +0.002139 | 1.0558 |
| src/compiler/transformers/utilities.ts:1:1 |  | 0.039114 | 0.041249 | +0.002135 | 1.0546 |
| src/compiler/utilities.ts:10969:1 | getNodeModulePathParts | 0.039350 | 0.041481 | +0.002131 | 1.0542 |
| src/compiler/utilities.ts:12204:1 | getNodeAtPosition | 0.147447 | 0.149571 | +0.002124 | 1.0144 |
| src/compiler/_namespaces/ts.ts:29:1 |  | 0.040730 | 0.042845 | +0.002115 | 1.0519 |
| src/compiler/builder.ts:1130:1 | IncrementalBuildInfoBase | 0.040720 | 0.042833 | +0.002113 | 1.0519 |
| src/compiler/utilities.ts:5303:1 | isTrivia | 0.062704 | 0.064817 | +0.002113 | 1.0337 |
| src/compiler/utilities.ts:9660:1 | getRegularExpressionForWildcard | 0.039298 | 0.041409 | +0.002111 | 1.0537 |
| src/compiler/types.ts:2608:1 | CompoundAssignmentOperator | 0.042861 | 0.044963 | +0.002101 | 1.0490 |
| src/compiler/types.ts:4967:1 | StructureIsReused | 0.038678 | 0.040757 | +0.002079 | 1.0537 |
| src/compiler/commandLineParser.ts:3258:1 | handleWatchOptionsConfigDirTemplateSubstitution | 0.038670 | 0.040748 | +0.002079 | 1.0538 |
| src/compiler/moduleNameResolver.ts:910:1 | NonRelativeModuleNameResolutionCache | 0.039237 | 0.041309 | +0.002072 | 1.0528 |
| src/compiler/checker.ts:1410:1 |  | 0.042680 | 0.044752 | +0.002072 | 1.0486 |
| src/compiler/moduleNameResolver.ts:1687:1 | NodeResolutionFeatures | 0.043483 | 0.045547 | +0.002064 | 1.0475 |
| src/compiler/types.ts:3820:1 | ModuleExportName | 0.039456 | 0.041515 | +0.002058 | 1.0522 |
| src/compiler/factory/nodeTests.ts:1156:1 | isJSDocThisTag | 0.044746 | 0.046800 | +0.002054 | 1.0459 |
| src/compiler/utilities.ts:6540:1 | getOwnEmitOutputFilePath | 0.081423 | 0.083477 | +0.002054 | 1.0252 |
| src/compiler/utilities.ts:6634:1 | getSourceFilesToEmit | 0.079980 | 0.082034 | +0.002054 | 1.0257 |
| src/compiler/utilities.ts:4848:1 | AssignmentTarget | 0.038944 | 0.040990 | +0.002045 | 1.0525 |
| src/compiler/performance.ts:170:1 | isEnabled | 0.038906 | 0.040950 | +0.002043 | 1.0525 |
| src/compiler/parser.ts:1372:1 | parseIsolatedEntityName | 0.040267 | 0.042310 | +0.002043 | 1.0507 |
| src/compiler/utilities.ts:9805:1 | getFileMatcherPatterns | 0.041460 | 0.043494 | +0.002035 | 1.0491 |
| src/compiler/semver.ts:425:1 | testComparator | 0.040901 | 0.042933 | +0.002032 | 1.0497 |
| src/compiler/tsbuildPublic.ts:169:1 | BuildResultFlags | 0.039085 | 0.041111 | +0.002026 | 1.0518 |
| src/compiler/builder.ts:1181:1 | isIncrementalBuildInfo | 0.041722 | 0.043745 | +0.002024 | 1.0485 |
| src/compiler/types.ts:2244:1 | TypeReferenceType | 0.039827 | 0.041850 | +0.002024 | 1.0508 |
| src/compiler/path.ts:828:1 | ensurePathIsNonModuleName | 0.039534 | 0.041555 | +0.002020 | 1.0511 |
| src/compiler/scanner.ts:957:1 | getLeadingCommentRanges | 0.041554 | 0.043573 | +0.002019 | 1.0486 |
| src/compiler/executeCommandLine.ts:1138:1 | canReportDiagnostics | 0.039961 | 0.041971 | +0.002010 | 1.0503 |
| src/compiler/transformers/declarations.ts:249:1 |  | 0.038208 | 0.040215 | +0.002007 | 1.0525 |
| src/compiler/types.ts:5705:1 | IdentifierTypePredicate | 0.038119 | 0.040120 | +0.002002 | 1.0525 |
| src/compiler/moduleNameResolver.ts:934:1 | PackageJsonInfoCache | 0.039213 | 0.041212 | +0.002000 | 1.0510 |
| src/compiler/types.ts:2733:1 | BindingOrAssignmentPattern | 0.040399 | 0.042395 | +0.001996 | 1.0494 |
| src/compiler/types.ts:2942:1 | SpreadElement | 0.045737 | 0.047733 | +0.001996 | 1.0436 |
| src/compiler/factory/emitHelpers.ts:1170:1 |  | 0.083698 | 0.085688 | +0.001991 | 1.0238 |
| src/compiler/utilities.ts:10206:1 | tryParsePattern | 0.040319 | 0.042301 | +0.001982 | 1.0492 |
| src/compiler/factory/nodeTests.ts:588:1 | isElementAccessExpression | 0.042906 | 0.044887 | +0.001981 | 1.0462 |
| src/compiler/utilities.ts:6942:1 | emitNewLineBeforeLeadingCommentsOfPosition | 0.079929 | 0.081908 | +0.001979 | 1.0248 |
| src/compiler/utilities.ts:8360:1 | getNameOfAccessExpression | 0.047949 | 0.049926 | +0.001977 | 1.0412 |
| src/compiler/transformers/destructuring.ts:66:1 | FlattenContext | 0.038157 | 0.040131 | +0.001974 | 1.0517 |
| src/compiler/factory/nodeTests.ts:640:1 | isPostfixUnaryExpression | 0.042962 | 0.044936 | +0.001973 | 1.0459 |
| src/compiler/sys.ts:115:1 | Levels | 0.037601 | 0.039574 | +0.001973 | 1.0525 |
| src/compiler/types.ts:3263:1 | JsxOpeningFragment | 0.041431 | 0.043401 | +0.001970 | 1.0475 |
| src/compiler/types.ts:6042:1 | Symbol | 0.038727 | 0.040693 | +0.001967 | 1.0508 |
| src/compiler/utilities.ts:5584:1 | isPushOrUnshiftIdentifier | 0.039312 | 0.041277 | +0.001965 | 1.0500 |
| src/compiler/utilities.ts:668:1 |  | 0.038250 | 0.040214 | +0.001964 | 1.0513 |
| src/compiler/types.ts:4995:1 | SourceMapSpan | 0.038572 | 0.040534 | +0.001962 | 1.0509 |
| src/compiler/semver.ts:447:1 | formatAlternative | 0.040400 | 0.042361 | +0.001961 | 1.0485 |
| src/compiler/resolutionCache.ts:325:1 | canWatchAtTypes | 0.043178 | 0.045132 | +0.001954 | 1.0453 |
| src/compiler/scanner.ts:431:1 | characterCodeToRegularExpressionFlag | 0.039596 | 0.041544 | +0.001948 | 1.0492 |
| src/compiler/utilities.ts:4988:1 | ValueSignatureDeclaration | 0.038817 | 0.040764 | +0.001947 | 1.0502 |
| src/compiler/builderPublic.ts:176:1 | createSemanticDiagnosticsBuilderProgram | 0.045001 | 0.046947 | +0.001946 | 1.0432 |
| src/compiler/types.ts:1968:1 | PrivateIdentifierAccessorDeclaration | 0.041818 | 0.043763 | +0.001945 | 1.0465 |
| src/compiler/types.ts:2270:1 | ArrayTypeNode | 0.040935 | 0.042880 | +0.001945 | 1.0475 |
| src/compiler/factory/emitHelpers.ts:1020:1 |  | 0.076715 | 0.078656 | +0.001942 | 1.0253 |
| src/compiler/utilities.ts:10908:1 | isCatchClauseVariableDeclaration | 0.044874 | 0.046807 | +0.001933 | 1.0431 |
| src/compiler/utilities.ts:2763:1 | getJSDocCommentRanges | 0.053509 | 0.055441 | +0.001932 | 1.0361 |
| src/compiler/transformers/generators.ts:245:1 | CodeBlockKind | 0.040138 | 0.042067 | +0.001930 | 1.0481 |
| src/compiler/transformers/es2021.ts:1:1 |  | 0.039923 | 0.041852 | +0.001929 | 1.0483 |
| src/compiler/executeCommandLine.ts:1142:1 | canTrace | 0.038833 | 0.040753 | +0.001920 | 1.0494 |
| src/compiler/factory/nodeTests.ts:752:1 | isForOfStatement | 0.053189 | 0.055100 | +0.001911 | 1.0359 |
| src/compiler/types.ts:6681:1 | VarianceFlags | 0.038117 | 0.040026 | +0.001909 | 1.0501 |
| src/compiler/utilities.ts:3569:1 | classOrConstructorParameterIsDecorated | 0.050797 | 0.052702 | +0.001905 | 1.0375 |
| src/compiler/programDiagnostics.ts:72:1 | ProgramDiagnostics | 0.041893 | 0.043790 | +0.001897 | 1.0453 |
| src/compiler/utilities.ts:10342:1 | Mutable | 0.039377 | 0.041272 | +0.001895 | 1.0481 |
| src/compiler/types.ts:4199:1 | FlowStart | 0.040434 | 0.042329 | +0.001895 | 1.0469 |
| src/compiler/builderState.ts:56:1 | SignatureInfo | 0.040446 | 0.042340 | +0.001894 | 1.0468 |
| src/compiler/factory/nodeTests.ts:604:1 | isTypeAssertionExpression | 0.043532 | 0.045426 | +0.001894 | 1.0435 |
| src/compiler/types.ts:2460:1 | NullLiteral | 0.043789 | 0.045680 | +0.001891 | 1.0432 |
| src/compiler/types.ts:8223:1 | TransformFlags | 0.039707 | 0.041596 | +0.001889 | 1.0476 |
| src/compiler/types.ts:4079:1 | JSDocCallbackTag | 0.040528 | 0.042411 | +0.001884 | 1.0465 |
| src/compiler/factory/nodeTests.ts:632:1 | isAwaitExpression | 0.046161 | 0.048042 | +0.001881 | 1.0407 |
| src/compiler/transformers/classThis.ts:33:1 | createClassThisAssignmentBlock | 0.038105 | 0.039984 | +0.001879 | 1.0493 |
| src/compiler/program.ts:688:1 |  | 0.039389 | 0.041267 | +0.001878 | 1.0477 |
| src/compiler/transformers/classThis.ts:1:1 |  | 0.038933 | 0.040794 | +0.001861 | 1.0478 |
| src/compiler/transformers/namedEvaluation.ts:236:1 | finishTransformNamedEvaluation | 0.042686 | 0.044543 | +0.001857 | 1.0435 |
| src/compiler/types.ts:6790:1 | MappedType | 0.040922 | 0.042775 | +0.001853 | 1.0453 |
| src/compiler/utilitiesPublic.ts:1562:1 | isTypeOnlyImportOrExportDeclaration | 0.046922 | 0.048773 | +0.001851 | 1.0394 |
| src/compiler/types.ts:7412:1 | CompilerOptionsValue | 0.041868 | 0.043709 | +0.001841 | 1.0440 |
| src/compiler/types.ts:2184:1 | TypeNode | 0.039995 | 0.041833 | +0.001838 | 1.0460 |
| src/compiler/_namespaces/ts.ts:10:1 |  | 0.040780 | 0.042616 | +0.001837 | 1.0450 |
| src/compiler/types.ts:3344:1 | SyntheticReferenceExpression | 0.043890 | 0.045727 | +0.001837 | 1.0418 |
| src/compiler/factory/nodeTests.ts:1160:1 | isJSDocTypeTag | 0.043958 | 0.045793 | +0.001835 | 1.0417 |
| src/compiler/utilities.ts:715:1 | optionsHaveModuleResolutionChanges | 0.040505 | 0.042339 | +0.001834 | 1.0453 |
| src/compiler/types.ts:10424:1 | PragmaPseudoMapEntry | 0.038348 | 0.040180 | +0.001832 | 1.0478 |
| src/compiler/types.ts:2977:1 | PrivateIdentifierPropertyAccessExpression | 0.043112 | 0.044939 | +0.001827 | 1.0424 |
| src/compiler/transformers/destructuring.ts:338:1 | flattenBindingOrAssignmentElement | 0.044755 | 0.046582 | +0.001827 | 1.0408 |
| src/compiler/utilities.ts:9925:1 | getIncludeBasePath | 0.041157 | 0.042981 | +0.001825 | 1.0443 |
| src/compiler/types.ts:2555:1 | RelationalOperator | 0.040123 | 0.041947 | +0.001824 | 1.0455 |
| src/compiler/parser.ts:1427:1 | parseJSDocTypeExpressionForTests | 0.047156 | 0.048979 | +0.001823 | 1.0387 |
| src/compiler/utilities.ts:3040:1 | introducesArgumentsExoticObject | 0.044064 | 0.045887 | +0.001822 | 1.0414 |
| src/compiler/types.ts:6801:1 | EvolvingArrayType | 0.040069 | 0.041886 | +0.001818 | 1.0454 |
| src/compiler/utilities.ts:710:1 | changesAffectModuleResolution | 0.039647 | 0.041464 | +0.001817 | 1.0458 |
| src/compiler/utilities.ts:10175:1 | removeFileExtension | 0.038766 | 0.040582 | +0.001816 | 1.0469 |
| src/compiler/types.ts:4150:1 | FlowFlags | 0.040933 | 0.042743 | +0.001810 | 1.0442 |
| src/compiler/executeCommandLine.ts:105:1 | countLines | 0.041163 | 0.042969 | +0.001805 | 1.0439 |
| src/compiler/checker.ts:54252:1 | signatureHasRestParameter | 0.039831 | 0.041624 | +0.001794 | 1.0450 |
| src/compiler/factory/nodeTests.ts:668:1 | isOmittedExpression | 0.043548 | 0.045341 | +0.001793 | 1.0412 |
| src/compiler/factory/nodeTests.ts:570:1 | isBindingElement | 0.043484 | 0.045276 | +0.001792 | 1.0412 |
| src/compiler/visitorPublic.ts:445:1 | addDefaultValueAssignmentForBindingPattern | 0.295436 | 0.297227 | +0.001790 | 1.0061 |
| src/compiler/scanner.ts:740:1 | isConflictMarkerTrivia | 0.039172 | 0.040953 | +0.001781 | 1.0455 |
| src/compiler/checker.ts:54236:1 | ReactNames | 0.039099 | 0.040876 | +0.001777 | 1.0454 |
| src/compiler/transformers/destructuring.ts:81:1 | FlattenLevel | 0.038107 | 0.039883 | +0.001777 | 1.0466 |
| src/compiler/watch.ts:898:1 | reportUnrecoverableDiagnostic | 0.083168 | 0.084943 | +0.001776 | 1.0214 |
| src/compiler/builder.ts:209:1 | BuilderProgramState | 0.039330 | 0.041105 | +0.001775 | 1.0451 |
| src/compiler/tracing.ts:25:1 |  | 0.040395 | 0.042163 | +0.001768 | 1.0438 |
| src/compiler/types.ts:4182:1 | FlowNodeBase | 0.040276 | 0.042040 | +0.001764 | 1.0438 |
| src/compiler/types.ts:7992:1 | ModuleResolutionHost | 0.039049 | 0.040810 | +0.001761 | 1.0451 |
| src/compiler/types.ts:2189:1 | TypeNode | 0.040477 | 0.042231 | +0.001755 | 1.0433 |
| src/compiler/performanceCore.ts:30:1 | tryGetPerformance | 0.039052 | 0.040803 | +0.001751 | 1.0448 |
| src/compiler/builder.ts:1078:1 | IncrementalBuildInfoDiagnosticOfFile | 0.044348 | 0.046096 | +0.001747 | 1.0394 |
| src/compiler/utilities.ts:2696:1 | isLet | 0.048980 | 0.050726 | +0.001746 | 1.0357 |
| src/compiler/transformers/es2018.ts:1:1 |  | 0.039771 | 0.041515 | +0.001745 | 1.0439 |
| src/compiler/emitter.ts:752:1 | emitFiles | 0.728505 | 0.730246 | +0.001741 | 1.0024 |
| src/compiler/types.ts:8062:1 | PackageId | 0.039589 | 0.041328 | +0.001739 | 1.0439 |
| src/compiler/types.ts:6493:1 | LiteralType | 0.039133 | 0.040871 | +0.001738 | 1.0444 |
| src/compiler/factory/nodeTests.ts:1100:1 | isJSDocAuthorTag | 0.044788 | 0.046524 | +0.001736 | 1.0388 |
| src/compiler/core.ts:1:1 |  | 0.039342 | 0.041076 | +0.001734 | 1.0441 |
| src/compiler/utilities.ts:824:1 | getResolvedModuleFromResolution | 0.037243 | 0.038975 | +0.001732 | 1.0465 |
| src/compiler/factory/utilities.ts:1597:1 | formatGeneratedName | 0.041608 | 0.043336 | +0.001728 | 1.0415 |
| src/compiler/builder.ts:291:1 | getPendingEmitKind | 0.039046 | 0.040767 | +0.001721 | 1.0441 |
| src/compiler/commandLineParser.ts:2162:1 |  | 0.068051 | 0.069771 | +0.001720 | 1.0253 |
| src/compiler/types.ts:2468:1 | FalseLiteral | 0.041981 | 0.043699 | +0.001718 | 1.0409 |
| src/compiler/core.ts:2247:1 | endsWith | 0.040357 | 0.042075 | +0.001718 | 1.0426 |
| src/compiler/moduleNameResolver.ts:1986:1 | nodeLoadModuleByRelativeName | 0.042916 | 0.044628 | +0.001711 | 1.0399 |
| src/compiler/types.ts:2240:1 | NodeWithTypeArguments | 0.040314 | 0.042022 | +0.001708 | 1.0424 |
| src/compiler/types.ts:5831:1 | AllDecorators | 0.037464 | 0.039169 | +0.001705 | 1.0455 |
| src/compiler/types.ts:7596:1 | TypeAcquisition | 0.042856 | 0.044559 | +0.001703 | 1.0397 |
| src/compiler/types.ts:2902:1 | TemplateLiteralToken | 0.041679 | 0.043378 | +0.001699 | 1.0408 |
| src/compiler/visitorPublic.ts:508:1 | visitFunctionBody | 0.053708 | 0.055401 | +0.001693 | 1.0315 |
| src/compiler/types.ts:4489:1 | CommentDirectiveType | 0.038191 | 0.039879 | +0.001688 | 1.0442 |
| src/compiler/utilities.ts:4441:1 | isDefaultImport | 0.041840 | 0.043526 | +0.001686 | 1.0403 |
| src/compiler/factory/utilities.ts:1256:1 | isAssignmentOperatorOrHigher | 0.080889 | 0.082566 | +0.001677 | 1.0207 |
| src/compiler/utilitiesPublic.ts:1165:1 | getJSDocOverrideTagNoCache | 0.055648 | 0.057316 | +0.001669 | 1.0300 |
| src/compiler/commandLineParser.ts:3313:1 | getSubstitutedStringArrayWithConfigDirTemplate | 0.039236 | 0.040905 | +0.001669 | 1.0425 |
| src/compiler/utilities.ts:9314:1 |  | 0.099707 | 0.101370 | +0.001663 | 1.0167 |
| src/compiler/types.ts:5797:1 | LateVisibilityPaintedStatement | 0.039024 | 0.040683 | +0.001659 | 1.0425 |
| src/compiler/factory/utilities.ts:1273:1 | BinaryExpressionState | 0.039479 | 0.041128 | +0.001649 | 1.0418 |
| src/compiler/types.ts:3308:1 | JsxText | 0.040505 | 0.042151 | +0.001646 | 1.0406 |
| src/compiler/utilities.ts:10433:1 | canIncludeBindAndCheckDiagnostics | 0.040452 | 0.042096 | +0.001644 | 1.0406 |
| src/compiler/moduleNameResolver.ts:3332:1 | resolveFromTypeRoot | 0.039927 | 0.041570 | +0.001643 | 1.0412 |
| src/compiler/utilities.ts:3875:1 | isVariableDeclarationInitializedToBareOrAccessedRequire | 0.048985 | 0.050625 | +0.001640 | 1.0335 |
| src/compiler/types.ts:2501:1 | AwaitExpression | 0.040237 | 0.041875 | +0.001638 | 1.0407 |
| src/compiler/factory/nodeTests.ts:700:1 | isCommaListExpression | 0.045030 | 0.046657 | +0.001627 | 1.0361 |
| src/compiler/utilities.ts:620:1 |  | 0.038693 | 0.040320 | +0.001627 | 1.0420 |
| src/compiler/commandLineParser.ts:3750:1 | getDefaultCompilerOptions | 0.046781 | 0.048406 | +0.001625 | 1.0347 |
| src/compiler/_namespaces/ts.ts:58:1 |  | 0.040514 | 0.042129 | +0.001616 | 1.0399 |
| src/compiler/_namespaces/ts.ts:71:1 |  | 0.040373 | 0.041988 | +0.001615 | 1.0400 |
| src/compiler/transformers/es2020.ts:42:1 | transformES2020 | 0.131620 | 0.133234 | +0.001614 | 1.0123 |
| src/compiler/commandLineParser.ts:363:1 |  | 0.045462 | 0.047074 | +0.001613 | 1.0355 |
| src/compiler/moduleSpecifiers.ts:386:1 | getModuleSpecifiersWithCacheInfo | 0.082676 | 0.084283 | +0.001607 | 1.0194 |
| src/compiler/types.ts:8141:1 | CompilerHost | 0.039587 | 0.041189 | +0.001602 | 1.0405 |
| src/compiler/program.ts:636:1 | getPreEmitDiagnostics | 0.040993 | 0.042595 | +0.001602 | 1.0391 |
| src/compiler/types.ts:7126:1 | InferenceInfo | 0.039152 | 0.040753 | +0.001601 | 1.0409 |
| src/compiler/utilities.ts:4825:1 | getJSDocRoot | 0.043994 | 0.045586 | +0.001593 | 1.0362 |
| src/compiler/utilities.ts:8803:1 | compareRelatedInformation | 0.040746 | 0.042338 | +0.001592 | 1.0391 |
| src/compiler/utilities.ts:5045:1 | skipParentheses | 0.040378 | 0.041967 | +0.001588 | 1.0393 |
| src/compiler/moduleSpecifiers.ts:1416:1 | tryGetRealFileNameForNonJsDeclarationFileName | 0.040892 | 0.042477 | +0.001585 | 1.0388 |
| src/compiler/types.ts:10439:1 | PragmaMap | 0.042449 | 0.044030 | +0.001581 | 1.0372 |
| src/compiler/transformers/jsx.ts:1:1 |  | 0.039505 | 0.041079 | +0.001574 | 1.0398 |
| src/compiler/executeCommandLine.ts:745:1 | isBuildCommand | 0.040896 | 0.042469 | +0.001574 | 1.0385 |
| src/compiler/builder.ts:1621:1 | computeSignatureWithDiagnostics | 4.952311 | 4.953883 | +0.001572 | 1.0003 |
| src/compiler/commandLineParser.ts:3304:1 |  | 0.038299 | 0.039867 | +0.001569 | 1.0410 |
| src/compiler/types.ts:6487:1 | FreshableIntrinsicType | 0.039085 | 0.040653 | +0.001568 | 1.0401 |
| src/compiler/_namespaces/ts.ts:74:1 |  | 0.038592 | 0.040155 | +0.001564 | 1.0405 |
| src/compiler/checker.ts:54279:1 | NodeBuilderContext | 0.038015 | 0.039576 | +0.001561 | 1.0411 |
| src/compiler/utilities.ts:11152:1 | getTextOfJsxNamespacedName | 0.044876 | 0.046435 | +0.001559 | 1.0347 |
| src/compiler/emitter.ts:597:1 | createAddOutput | 0.041581 | 0.043137 | +0.001556 | 1.0374 |
| src/compiler/moduleNameResolver.ts:1974:1 | realPath | 0.131964 | 0.133518 | +0.001554 | 1.0118 |
| src/compiler/types.ts:2217:1 | ThisTypeNode | 0.040681 | 0.042235 | +0.001553 | 1.0382 |
| src/compiler/commandLineParser.ts:3396:1 | ExtendsResult | 0.040008 | 0.041556 | +0.001548 | 1.0387 |
| src/compiler/utilities.ts:5001:1 | walkUp | 0.050228 | 0.051772 | +0.001544 | 1.0307 |
| src/compiler/program.ts:387:1 | createCompilerHost | 0.039071 | 0.040614 | +0.001543 | 1.0395 |
| src/compiler/commandLineParser.ts:4281:1 | getOptionValueWithEmptyStrings | 0.038667 | 0.040209 | +0.001542 | 1.0399 |
| src/compiler/types.ts:3104:1 | ImportCall | 0.040091 | 0.041632 | +0.001542 | 1.0385 |
| src/compiler/types.ts:4459:1 | ReadonlyPragmaContext | 0.038596 | 0.040127 | +0.001532 | 1.0397 |
| src/compiler/types.ts:6444:1 | Type | 0.039358 | 0.040885 | +0.001528 | 1.0388 |
| src/compiler/types.ts:7248:1 | RepopulateModeMismatchDiagnosticChain | 0.040262 | 0.041789 | +0.001527 | 1.0379 |
| src/compiler/transformers/declarations.ts:1954:1 | isProcessedComponent | 0.042494 | 0.044019 | +0.001525 | 1.0359 |
| src/compiler/types.ts:2680:1 | ObjectBindingOrAssignmentElement | 0.040745 | 0.042269 | +0.001524 | 1.0374 |
| src/compiler/types.ts:10108:1 | SyntaxList | 0.041729 | 0.043252 | +0.001523 | 1.0365 |
| src/compiler/core.ts:2105:1 |  | 0.039805 | 0.041328 | +0.001523 | 1.0383 |
| src/compiler/utilitiesPublic.ts:1744:1 | isClassLike | 0.048071 | 0.049593 | +0.001521 | 1.0316 |
| src/compiler/moduleNameResolver.ts:143:1 | noPackageId | 0.038959 | 0.040480 | +0.001521 | 1.0390 |
| src/compiler/tsbuildPublic.ts:259:1 | isCircularBuildOrder | 0.040084 | 0.041600 | +0.001516 | 1.0378 |
| src/compiler/transformers/classFields.ts:333:1 | ClassFacts | 0.039114 | 0.040625 | +0.001511 | 1.0386 |
| src/compiler/builderStatePublic.ts:1:1 |  | 0.042432 | 0.043942 | +0.001510 | 1.0356 |
| src/compiler/builder.ts:281:1 | getBuilderFileEmit | 0.039178 | 0.040688 | +0.001510 | 1.0385 |
| src/compiler/_namespaces/ts.ts:79:1 |  | 0.038660 | 0.040170 | +0.001509 | 1.0390 |
| src/compiler/_namespaces/ts.ts:39:1 |  | 0.040880 | 0.042388 | +0.001508 | 1.0369 |
| src/compiler/factory/nodeTests.ts:644:1 | isBinaryExpression | 0.042591 | 0.044098 | +0.001507 | 1.0354 |
| src/compiler/commandLineParser.ts:3029:1 | parseJsonSourceFileConfigFileContent | 0.086559 | 0.088063 | +0.001504 | 1.0174 |
| src/compiler/utilities.ts:7250:1 | getModifierFlagsWorker | 0.064677 | 0.066175 | +0.001499 | 1.0232 |
| src/compiler/commandLineParser.ts:4268:1 | convertCompilerOptionsForTelemetry | 0.040680 | 0.042166 | +0.001486 | 1.0365 |
| src/compiler/factory/utilities.ts:1518:1 | getNodeForGeneratedName | 0.039370 | 0.040856 | +0.001486 | 1.0377 |
| src/compiler/commandLineParser.ts:2719:1 | optionMapToObject | 0.039013 | 0.040496 | +0.001483 | 1.0380 |
| src/compiler/factory/emitHelpers.ts:96:1 | ESDecorateContext | 0.040789 | 0.042271 | +0.001482 | 1.0363 |
| src/compiler/types.ts:4494:1 | Bundle | 0.038304 | 0.039785 | +0.001481 | 1.0387 |
| src/compiler/types.ts:3314:1 | JsxChild | 0.040651 | 0.042127 | +0.001476 | 1.0363 |
| src/compiler/moduleNameResolver.ts:1766:1 | bundlerModuleNameResolver | 0.043712 | 0.045187 | +0.001475 | 1.0337 |
| src/compiler/commandLineParser.ts:2995:1 | convertToOptionValueWithAbsolutePaths | 0.039342 | 0.040817 | +0.001475 | 1.0375 |
| src/compiler/types.ts:3295:1 | JsxClosingElement | 0.039566 | 0.041036 | +0.001470 | 1.0371 |
| src/compiler/types.ts:2223:1 | FunctionOrConstructorTypeNodeBase | 0.040468 | 0.041937 | +0.001469 | 1.0363 |
| src/compiler/factory/nodeTests.ts:1180:1 | isJSDocImplementsTag | 0.043643 | 0.045111 | +0.001468 | 1.0336 |
| src/compiler/parser.ts:1414:1 | parseIsolatedJSDocComment | 0.047869 | 0.049335 | +0.001466 | 1.0306 |
| src/compiler/performance.ts:65:1 |  | 0.038514 | 0.039977 | +0.001463 | 1.0380 |
| src/compiler/resolutionCache.ts:220:1 | FileWatcherOfAffectingLocation | 0.043058 | 0.044520 | +0.001462 | 1.0340 |
| src/compiler/types.ts:8135:1 | HasInvalidatedResolutions | 0.039915 | 0.041371 | +0.001456 | 1.0365 |
| src/compiler/types.ts:5679:1 | SymbolAccessibility | 0.037677 | 0.039130 | +0.001453 | 1.0386 |
| src/compiler/moduleNameResolver.ts:1397:1 | resolveLibrary | 0.041875 | 0.043326 | +0.001451 | 1.0346 |
| src/compiler/transformers/destructuring.ts:227:1 | bindingOrAssignmentPatternContainsNonLiteralComputedName | 0.043725 | 0.045175 | +0.001450 | 1.0332 |
| src/compiler/program.ts:1471:1 | shouldProgramCreateNewSourceFiles | 0.078295 | 0.079739 | +0.001445 | 1.0185 |
| src/compiler/path.ts:757:1 | getNormalizedAbsolutePathWithoutRoot | 0.039007 | 0.040450 | +0.001443 | 1.0370 |
| src/compiler/utilities.ts:4710:1 | getNextJSDocCommentLocation | 0.187712 | 0.189153 | +0.001440 | 1.0077 |
| src/compiler/tsbuildPublic.ts:251:1 | CircularBuildOrder | 0.038320 | 0.039756 | +0.001436 | 1.0375 |
| src/compiler/program.ts:1478:1 | createCreateProgramOptions | 0.039013 | 0.040445 | +0.001432 | 1.0367 |
| src/compiler/utilities.ts:7444:1 | tryGetClassExtendingExpressionWithTypeArguments | 0.049679 | 0.051109 | +0.001430 | 1.0288 |
| src/compiler/program.ts:1:1 |  | 0.039867 | 0.041292 | +0.001425 | 1.0357 |
| src/compiler/types.ts:4286:1 | SourceFileLike | 0.038327 | 0.039750 | +0.001424 | 1.0371 |
| src/compiler/factory/nodeTests.ts:652:1 | isTemplateExpression | 0.043853 | 0.045267 | +0.001413 | 1.0322 |
| src/compiler/types.ts:8044:1 | ResolvedModuleFull | 0.042743 | 0.044156 | +0.001413 | 1.0331 |
| src/compiler/factory/nodeTests.ts:275:1 | isDotDotDotToken | 0.045550 | 0.046960 | +0.001410 | 1.0310 |
| src/compiler/path.ts:492:1 | getPathComponents | 0.038211 | 0.039616 | +0.001405 | 1.0368 |
| src/compiler/factory/utilities.ts:843:1 | tryGetModuleNameFromFile | 0.119990 | 0.121392 | +0.001401 | 1.0117 |
| src/compiler/utilities.ts:644:1 | createSymbolTable | 0.040853 | 0.042252 | +0.001399 | 1.0343 |
| src/compiler/executeCommandLine.ts:1007:1 | createWatchStatusReporter | 0.040606 | 0.042003 | +0.001397 | 1.0344 |
| src/compiler/commandLineParser.ts:2706:1 | optionDependsOn | 0.040823 | 0.042218 | +0.001395 | 1.0342 |
| src/compiler/utilities.ts:8355:1 | isAccessExpression | 0.046528 | 0.047924 | +0.001395 | 1.0300 |
| src/compiler/sys.ts:1458:1 | DirectoryWatcher | 0.039298 | 0.040693 | +0.001395 | 1.0355 |
| src/compiler/types.ts:3289:1 | JsxSpreadAttribute | 0.040108 | 0.041497 | +0.001390 | 1.0347 |
| src/compiler/visitorPublic.ts:436:1 | addDefaultValueAssignmentIfNeeded | 0.223622 | 0.225011 | +0.001389 | 1.0062 |
| src/compiler/transformers/destructuring.ts:615:1 | makeObjectAssignmentPattern | 0.048879 | 0.050262 | +0.001383 | 1.0283 |
| src/compiler/visitorPublic.ts:148:1 | visitNode | 0.046673 | 0.048053 | +0.001380 | 1.0296 |
| src/compiler/types.ts:3354:1 | DebuggerStatement | 0.042151 | 0.043528 | +0.001377 | 1.0327 |
| src/compiler/performanceCore.ts:1:1 |  | 0.039644 | 0.041019 | +0.001376 | 1.0347 |
| src/compiler/utilities.ts:954:1 | containsParseError | 0.046481 | 0.047856 | +0.001375 | 1.0296 |
| src/compiler/transformers/taggedTemplate.ts:110:1 | getRawLiteral | 0.049666 | 0.051041 | +0.001374 | 1.0277 |
| src/compiler/factory/nodeTests.ts:985:1 | isHeritageClause | 0.055518 | 0.056891 | +0.001373 | 1.0247 |
| src/compiler/types.ts:7852:1 | CharacterCodes | 0.040881 | 0.042249 | +0.001369 | 1.0335 |
| src/compiler/utilities.ts:4830:1 | getTypeParameterFromJsDoc | 0.039251 | 0.040613 | +0.001362 | 1.0347 |
| src/compiler/utilities.ts:1971:1 | GetLiteralTextFlags | 0.039179 | 0.040537 | +0.001358 | 1.0347 |
| src/compiler/semver.ts:451:1 | formatComparator | 0.040257 | 0.041613 | +0.001356 | 1.0337 |
| src/compiler/utilities.ts:622:1 |  | 0.039047 | 0.040399 | +0.001352 | 1.0346 |
| src/compiler/core.ts:19:1 |  | 0.039323 | 0.040674 | +0.001351 | 1.0344 |
| src/compiler/factory/utilities.ts:690:1 | hasRecordedExternalHelpers | 0.046765 | 0.048111 | +0.001345 | 1.0288 |
| src/compiler/transformers/utilities.ts:585:1 | getProperties | 0.039167 | 0.040502 | +0.001334 | 1.0341 |
| src/compiler/commandLineParser.ts:2206:1 | getDiagnosticText | 0.039448 | 0.040781 | +0.001333 | 1.0338 |
| src/compiler/types.ts:10343:1 | PragmaArgumentType | 0.038948 | 0.040280 | +0.001333 | 1.0342 |
| src/compiler/builder.ts:617:1 | releaseCache | 0.045126 | 0.046458 | +0.001332 | 1.0295 |
| src/compiler/performanceCore.ts:20:1 | Performance | 0.040326 | 0.041645 | +0.001320 | 1.0327 |
| src/compiler/factory/nodeTests.ts:608:1 | isParenthesizedExpression | 0.044603 | 0.045922 | +0.001319 | 1.0296 |
| src/compiler/_namespaces/ts.ts:18:1 |  | 0.039814 | 0.041133 | +0.001318 | 1.0331 |
| src/compiler/parser.ts:504:1 | ForEachChildTable | 0.039911 | 0.041229 | +0.001318 | 1.0330 |
| src/compiler/path.ts:312:1 | getDirectoryPath | 0.040956 | 0.042272 | +0.001317 | 1.0321 |
| src/compiler/executeCommandLine.ts:554:1 | printHelp | 0.040875 | 0.042190 | +0.001315 | 1.0322 |
| src/compiler/types.ts:3113:1 | NewExpression | 0.039770 | 0.041084 | +0.001314 | 1.0330 |
| src/compiler/utilities.ts:9687:1 | isImplicitGlob | 0.038944 | 0.040258 | +0.001314 | 1.0337 |
| src/compiler/utilities.ts:7435:1 | isAssignmentOperator | 0.044199 | 0.045511 | +0.001312 | 1.0297 |
| src/compiler/utilities.ts:8369:1 | isNamedImportsOrExports | 0.045862 | 0.047169 | +0.001308 | 1.0285 |
| src/compiler/moduleNameResolver.ts:966:1 | CacheWithRedirects | 0.040368 | 0.041674 | +0.001305 | 1.0323 |
| src/compiler/semver.ts:248:1 |  | 0.041061 | 0.042361 | +0.001301 | 1.0317 |
| src/compiler/utilities.ts:10160:1 | numberOfDirectorySeparators | 0.038445 | 0.039744 | +0.001299 | 1.0338 |
| src/compiler/commandLineParser.ts:1:1 |  | 0.042469 | 0.043760 | +0.001291 | 1.0304 |
| src/compiler/types.ts:3095:1 | BindablePropertyAssignmentExpression | 0.040165 | 0.041449 | +0.001284 | 1.0320 |
| src/compiler/types.ts:2165:1 | IndexSignatureDeclaration | 0.041332 | 0.042614 | +0.001282 | 1.0310 |
| src/compiler/emitter.ts:618:1 | getOwnOutputFileNames | 0.041596 | 0.042876 | +0.001281 | 1.0308 |
| src/compiler/commandLineParser.ts:2483:1 | convertToJson | 0.047289 | 0.048567 | +0.001278 | 1.0270 |
| src/compiler/utilities.ts:1321:1 | getTextOfNodeFromSourceText | 0.051018 | 0.052294 | +0.001277 | 1.0250 |
| src/compiler/factory/nodeTests.ts:1136:1 | isJSDocDeprecatedTag | 0.044890 | 0.046166 | +0.001276 | 1.0284 |
| src/compiler/program.ts:1370:1 |  | 0.042458 | 0.043729 | +0.001270 | 1.0299 |
| src/compiler/types.ts:3269:1 | JsxClosingFragment | 0.042867 | 0.044129 | +0.001262 | 1.0294 |
| src/compiler/types.ts:9947:1 | DocumentPositionMapperHost | 0.039699 | 0.040959 | +0.001260 | 1.0317 |
| src/compiler/transformers/es2020.ts:1:1 |  | 0.040243 | 0.041501 | +0.001259 | 1.0313 |
| src/compiler/utilitiesPublic.ts:682:1 | getNodeFlags | 0.046825 | 0.048082 | +0.001258 | 1.0269 |
| src/compiler/sys.ts:1396:1 | System | 0.038546 | 0.039803 | +0.001258 | 1.0326 |
| src/compiler/scanner.ts:289:1 |  | 0.039171 | 0.040416 | +0.001245 | 1.0318 |
| src/compiler/types.ts:10334:1 | ArgumentDefinitionToFieldUnion | 0.039723 | 0.040966 | +0.001242 | 1.0313 |
| src/compiler/utilities.ts:10913:1 | isFunctionExpressionOrArrowFunction | 0.045007 | 0.046249 | +0.001242 | 1.0276 |
| src/compiler/tsbuildPublic.ts:241:1 | SolutionBuilderHost | 0.040078 | 0.041318 | +0.001240 | 1.0309 |
| src/compiler/types.ts:4190:1 | FlowUnreachable | 0.039635 | 0.040873 | +0.001237 | 1.0312 |
| src/compiler/types.ts:3358:1 | MissingDeclaration | 0.044112 | 0.045348 | +0.001236 | 1.0280 |
| src/compiler/factory/utilities.ts:1:1 |  | 0.038733 | 0.039969 | +0.001235 | 1.0319 |
| src/compiler/types.ts:1610:1 | EqualsToken | 0.057493 | 0.058720 | +0.001227 | 1.0213 |
| src/compiler/types.ts:6807:1 | ReverseMappedType | 0.040888 | 0.042109 | +0.001221 | 1.0299 |
| src/compiler/utilities.ts:1022:1 | nodePosToString | 0.047713 | 0.048929 | +0.001216 | 1.0255 |
| src/compiler/tsbuild.ts:10:1 | UpToDateStatusType | 0.040209 | 0.041425 | +0.001215 | 1.0302 |
| src/compiler/types.ts:3840:1 | TypeOnlyExportDeclaration | 0.039537 | 0.040752 | +0.001215 | 1.0307 |
| src/compiler/executeCommandLine.ts:115:1 | getCountsMap | 0.040896 | 0.042109 | +0.001213 | 1.0296 |
| src/compiler/types.ts:6713:1 | TupleType | 0.038397 | 0.039604 | +0.001208 | 1.0314 |
| src/compiler/checker.ts:1174:1 | IterationUse | 0.041477 | 0.042683 | +0.001206 | 1.0291 |
| src/compiler/types.ts:10327:1 | PragmaArgTypeOptional | 0.037960 | 0.039163 | +0.001204 | 1.0317 |
| src/compiler/factory/utilities.ts:1208:1 | isRelationalOperator | 0.040680 | 0.041883 | +0.001203 | 1.0296 |
| src/compiler/types.ts:3152:1 | SatisfiesExpression | 0.040380 | 0.041577 | +0.001197 | 1.0296 |
| src/compiler/builder.ts:105:1 | ReusableDiagnosticRelatedInformation | 0.040225 | 0.041420 | +0.001196 | 1.0297 |
| src/compiler/factory/utilities.ts:1599:1 | formatGeneratedName | 0.041741 | 0.042935 | +0.001195 | 1.0286 |
| src/compiler/utilities.ts:1384:1 |  | 0.066873 | 0.068065 | +0.001192 | 1.0178 |
| src/compiler/types.ts:7251:1 | RepopulateDiagnosticChainInfo | 0.040283 | 0.041472 | +0.001190 | 1.0295 |
| src/compiler/builder.ts:2440:1 | getNonIncrementalBuildInfoRoots | 0.044723 | 0.045907 | +0.001184 | 1.0265 |
| src/compiler/transformers/declarations.ts:221:1 | getDeclarationDiagnostics | 0.077251 | 0.078434 | +0.001184 | 1.0153 |
| src/compiler/types.ts:3802:1 | NamedImportsOrExports | 0.039160 | 0.040344 | +0.001183 | 1.0302 |
| src/compiler/moduleSpecifiers.ts:1459:1 | isPathRelativeToParent | 0.039714 | 0.040897 | +0.001183 | 1.0298 |
| src/compiler/tsbuildPublic.ts:264:1 | getBuildOrderFromAnyBuildOrder | 0.069227 | 0.070410 | +0.001182 | 1.0171 |
| src/compiler/scanner.ts:961:1 | getTrailingCommentRanges | 0.041849 | 0.043031 | +0.001182 | 1.0282 |
| src/compiler/types.ts:3426:1 | ForInOrOfStatement | 0.042129 | 0.043310 | +0.001181 | 1.0280 |
| src/compiler/utilities.ts:2491:1 | createDiagnosticForRange | 0.040976 | 0.042156 | +0.001180 | 1.0288 |
| src/compiler/executeCommandLine.ts:1054:1 | SolutionPerformance | 0.040902 | 0.042079 | +0.001177 | 1.0288 |
| src/compiler/program.ts:1499:1 | createProgram | 0.072461 | 0.073634 | +0.001173 | 1.0162 |
| src/compiler/utilitiesPublic.ts:762:1 | getOriginalNode | 0.039623 | 0.040795 | +0.001172 | 1.0296 |
| src/compiler/factory/utilities.ts:1120:1 | canHaveIllegalDecorators | 0.045599 | 0.046771 | +0.001172 | 1.0257 |
| src/compiler/factory/parenthesizerRules.ts:1:1 |  | 0.038256 | 0.039418 | +0.001163 | 1.0304 |
| src/compiler/factory/utilities.ts:437:1 | createExpressionForMethodDeclaration | 0.090932 | 0.092087 | +0.001155 | 1.0127 |
| src/compiler/types.ts:2808:1 | NoSubstitutionTemplateLiteral | 0.041344 | 0.042498 | +0.001154 | 1.0279 |
| src/compiler/_namespaces/ts.ts:36:1 |  | 0.041259 | 0.042413 | +0.001154 | 1.0280 |
| src/compiler/transformers/classFields.ts:313:1 | PrivateIdentifierInfo | 0.039442 | 0.040586 | +0.001144 | 1.0290 |
| src/compiler/executeCommandLine.ts:513:1 | printAllHelp | 0.045006 | 0.046142 | +0.001135 | 1.0252 |
| src/compiler/scanner.ts:974:1 | isIdentifierStart | 0.039562 | 0.040695 | +0.001133 | 1.0286 |
| src/compiler/scanner.ts:42:1 | tokenIsIdentifierOrKeyword | 0.042863 | 0.043994 | +0.001132 | 1.0264 |
| src/compiler/utilities.ts:10220:1 | ParsedPatterns | 0.039564 | 0.040693 | +0.001129 | 1.0285 |
| src/compiler/utilities.ts:725:1 | optionsHaveChanges | 0.040439 | 0.041564 | +0.001125 | 1.0278 |
| src/compiler/transformer.ts:76:1 |  | 0.039489 | 0.040612 | +0.001123 | 1.0284 |
| src/compiler/types.ts:2726:1 | ArrayBindingOrAssignmentPattern | 0.042418 | 0.043538 | +0.001120 | 1.0264 |
| src/compiler/commandLineParser.ts:2316:1 |  | 0.050066 | 0.051173 | +0.001106 | 1.0221 |
| src/compiler/path.ts:371:1 | getBaseFileName | 0.040272 | 0.041376 | +0.001105 | 1.0274 |
| src/compiler/debug.ts:113:1 | Debug | 0.040651 | 0.041754 | +0.001103 | 1.0271 |
| src/compiler/commandLineParser.ts:3392:1 | isSuccessfulParsedTsconfig | 0.040984 | 0.042085 | +0.001101 | 1.0269 |
| src/compiler/types.ts:9814:1 | PrintHandlers | 0.038836 | 0.039936 | +0.001100 | 1.0283 |
| src/compiler/utilities.ts:4242:1 | getAssignmentDeclarationPropertyAccessKind | 0.048770 | 0.049870 | +0.001100 | 1.0226 |
| src/compiler/path.ts:449:1 | pathComponents | 0.039451 | 0.040549 | +0.001099 | 1.0278 |
| src/compiler/utilities.ts:4941:1 | NodeWithPossibleHoistedDeclaration | 0.040999 | 0.042092 | +0.001093 | 1.0267 |
| src/compiler/types.ts:3443:1 | BreakStatement | 0.042653 | 0.043743 | +0.001090 | 1.0255 |
| src/compiler/program.ts:1345:1 | getImpliedNodeFormatForFileWorker | 0.040485 | 0.041574 | +0.001088 | 1.0269 |
| src/compiler/utilities.ts:10268:1 | positionIsSynthesized | 0.040690 | 0.041778 | +0.001088 | 1.0267 |
| src/compiler/utilitiesPublic.ts:1156:1 | getJSDocReadonlyTag | 0.055823 | 0.056910 | +0.001087 | 1.0195 |
| src/compiler/factory/nodeTests.ts:1188:1 | isJSDocThrowsTag | 0.044484 | 0.045568 | +0.001084 | 1.0244 |
| src/compiler/_namespaces/ts.ts:52:1 |  | 0.041565 | 0.042644 | +0.001078 | 1.0259 |
| src/compiler/scanner.ts:792:1 |  | 0.039149 | 0.040223 | +0.001074 | 1.0274 |
| src/compiler/types.ts:8139:1 | HasChangedAutomaticTypeDirectiveNames | 0.040056 | 0.041129 | +0.001073 | 1.0268 |
| src/compiler/scanner.ts:952:1 | appendCommentRange | 0.040315 | 0.041387 | +0.001072 | 1.0266 |
| src/compiler/factory/nodeFactory.ts:7286:1 | propagateNameFlags | 0.201440 | 0.202511 | +0.001072 | 1.0053 |
| src/compiler/utilities.ts:10064:1 | getModuleSpecifierEndingPreference | 0.039276 | 0.040347 | +0.001071 | 1.0273 |
| src/compiler/tsbuildPublic.ts:245:1 | SolutionBuilderWithWatchHost | 0.039306 | 0.040374 | +0.001069 | 1.0272 |
| src/compiler/utilities.ts:3912:1 | isSingleOrDoubleQuote | 0.042765 | 0.043831 | +0.001067 | 1.0249 |
| src/compiler/builder.ts:1125:1 | IncrementalBuildInfoRoot | 0.040622 | 0.041677 | +0.001055 | 1.0260 |
| src/compiler/utilities.ts:4526:1 | getNestedModuleDeclaration | 0.248833 | 0.249880 | +0.001047 | 1.0042 |
| src/compiler/types.ts:1611:1 | AmpersandAmpersandEqualsToken | 0.057319 | 0.058363 | +0.001044 | 1.0182 |
| src/compiler/utilities.ts:6609:1 | getPathsBasePath | 0.040394 | 0.041437 | +0.001044 | 1.0258 |
| src/compiler/utilities.ts:798:1 | getFullWidth | 0.042630 | 0.043668 | +0.001038 | 1.0244 |
| src/compiler/utilities.ts:2781:1 |  | 0.041040 | 0.042076 | +0.001036 | 1.0252 |
| src/compiler/types.ts:3203:1 | JsxCallLike | 0.045111 | 0.046147 | +0.001035 | 1.0229 |
| src/compiler/utilities.ts:9779:1 | replaceWildcardCharacter | 0.038966 | 0.039997 | +0.001031 | 1.0265 |
| src/compiler/_namespaces/ts.ts:50:1 |  | 0.040127 | 0.041157 | +0.001030 | 1.0257 |
| src/compiler/types.ts:2215:1 | LiteralImportTypeNode | 0.040427 | 0.041455 | +0.001028 | 1.0254 |
| src/compiler/utilities.ts:4288:1 | PrototypePropertyAssignment | 0.042070 | 0.043096 | +0.001026 | 1.0244 |
| src/compiler/types.ts:3866:1 | CheckJsDirective | 0.039451 | 0.040477 | +0.001025 | 1.0260 |
| src/compiler/program.ts:634:1 | getPreEmitDiagnostics | 0.042029 | 0.043053 | +0.001024 | 1.0244 |
| src/compiler/utilities.ts:1305:1 | moduleExportNameTextEscaped | 0.040880 | 0.041899 | +0.001019 | 1.0249 |
| src/compiler/resolutionCache.ts:166:1 | ResolutionWithFailedLookupLocations | 0.041926 | 0.042942 | +0.001016 | 1.0242 |
| src/compiler/utilities.ts:7206:1 | hasStaticModifier | 0.049613 | 0.050622 | +0.001010 | 1.0204 |
| src/compiler/moduleNameResolver.ts:1656:1 | tryLoadModuleUsingBaseUrl | 0.040997 | 0.042003 | +0.001005 | 1.0245 |
| src/compiler/transformers/namedEvaluation.ts:470:1 | transformNamedEvaluation | 0.038656 | 0.039656 | +0.001000 | 1.0259 |
| src/compiler/_namespaces/ts.ts:28:1 |  | 0.040574 | 0.041569 | +0.000994 | 1.0245 |
| src/compiler/types.ts:6507:1 | NumberLiteralType | 0.039072 | 0.040063 | +0.000992 | 1.0254 |
| src/compiler/binder.ts:3789:1 | isExportsOrModuleExportsOrAlias | 0.052885 | 0.053876 | +0.000991 | 1.0187 |
| src/compiler/transformers/namedEvaluation.ts:156:1 | classHasExplicitlyAssignedName | 0.044782 | 0.045771 | +0.000989 | 1.0221 |
| src/compiler/commandLineParser.ts:3309:1 | getSubstitutedPathWithConfigDirTemplate | 0.038923 | 0.039905 | +0.000982 | 1.0252 |
| src/compiler/utilities.ts:3163:1 | ThisContainer | 0.040816 | 0.041795 | +0.000979 | 1.0240 |
| src/compiler/types.ts:1607:1 | QuestionToken | 0.053211 | 0.054189 | +0.000978 | 1.0184 |
| src/compiler/types.ts:5045:1 | TypeCheckerHost | 0.037831 | 0.038801 | +0.000969 | 1.0256 |
| src/compiler/types.ts:6440:1 | TypeId | 0.040365 | 0.041326 | +0.000962 | 1.0238 |
| src/compiler/scanner.ts:423:1 |  | 0.048909 | 0.049869 | +0.000960 | 1.0196 |
| src/compiler/utilitiesPublic.ts:1004:1 | getAssignedName | 0.286789 | 0.287748 | +0.000959 | 1.0033 |
| src/compiler/utilities.ts:7960:1 | positionsAreOnSameLine | 0.040638 | 0.041593 | +0.000955 | 1.0235 |
| src/compiler/utilities.ts:3132:1 | getContainingFunctionDeclaration | 0.045691 | 0.046645 | +0.000955 | 1.0209 |
| src/compiler/types.ts:2205:1 | ImportTypeNode | 0.041012 | 0.041964 | +0.000952 | 1.0232 |
| src/compiler/utilities.ts:8590:1 |  | 0.039307 | 0.040256 | +0.000949 | 1.0241 |
| src/compiler/executeCommandLine.ts:1067:1 | createSolutionPerfomrance | 0.063671 | 0.064620 | +0.000949 | 1.0149 |
| src/compiler/factory/nodeTests.ts:343:1 | isAssertsKeyword | 0.045178 | 0.046125 | +0.000947 | 1.0210 |
| src/compiler/utilities.ts:3106:1 | getTsConfigObjectLiteralExpression | 0.047852 | 0.048799 | +0.000947 | 1.0198 |
| src/compiler/utilities.ts:1316:1 | moduleExportNameIsDefault | 0.040184 | 0.041130 | +0.000947 | 1.0236 |
| src/compiler/utilities.ts:7296:1 | getRawJSDocModifierFlagsNoCache | 0.327854 | 0.328797 | +0.000943 | 1.0029 |
| src/compiler/utilities.ts:10771:1 | expressionResultIsUnused | 0.049372 | 0.050314 | +0.000941 | 1.0191 |
| src/compiler/executeCommandLine.ts:92:1 | Statistic | 0.040503 | 0.041443 | +0.000939 | 1.0232 |
| src/compiler/_namespaces/ts.ts:72:1 |  | 0.039122 | 0.040060 | +0.000938 | 1.0240 |
| src/compiler/transformers/es2015.ts:1:1 |  | 0.039454 | 0.040390 | +0.000936 | 1.0237 |
| src/compiler/factory/utilitiesPublic.ts:14:1 | canHaveModifiers | 0.053599 | 0.054528 | +0.000929 | 1.0173 |
| src/compiler/types.ts:7414:1 | CompilerOptions | 0.041681 | 0.042600 | +0.000919 | 1.0220 |
| src/compiler/moduleSpecifiers.ts:796:1 |  | 0.038895 | 0.039802 | +0.000907 | 1.0233 |
| src/compiler/utilities.ts:4654:1 | getJSDocCommentsAndTags | 0.241992 | 0.242896 | +0.000905 | 1.0037 |
| src/compiler/factory/nodeTests.ts:656:1 | isYieldExpression | 0.043708 | 0.044610 | +0.000903 | 1.0207 |
| src/compiler/utilities.ts:10751:1 | isPackedElement | 0.292444 | 0.293337 | +0.000894 | 1.0031 |
| src/compiler/tsbuild.ts:1:1 |  | 0.039907 | 0.040798 | +0.000891 | 1.0223 |
| src/compiler/utilities.ts:8147:1 | compareDataObjects | 0.040360 | 0.041249 | +0.000888 | 1.0220 |
| src/compiler/types.ts:3327:1 | NotEmittedStatement | 0.042499 | 0.043379 | +0.000880 | 1.0207 |
| src/compiler/expressionToTypeNode.ts:1:1 |  | 0.043668 | 0.044548 | +0.000880 | 1.0201 |
| src/compiler/performance.ts:114:1 | getCount | 0.040967 | 0.041845 | +0.000878 | 1.0214 |
| src/compiler/types.ts:3852:1 | ExportAssignment | 0.039304 | 0.040177 | +0.000873 | 1.0222 |
| src/compiler/factory/nodeTests.ts:592:1 | isCallExpression | 0.044392 | 0.045265 | +0.000873 | 1.0197 |
| src/compiler/factory/nodeTests.ts:1148:1 | isJSDocParameterTag | 0.046117 | 0.046982 | +0.000865 | 1.0188 |
| src/compiler/factory/nodeTests.ts:1140:1 | isJSDocSeeTag | 0.045003 | 0.045866 | +0.000863 | 1.0192 |
| src/compiler/types.ts:9701:1 | TransformerFactory | 0.039744 | 0.040603 | +0.000859 | 1.0216 |
| src/compiler/transformers/utilities.ts:600:1 | getStaticPropertiesAndClassStaticBlock | 0.041484 | 0.042342 | +0.000858 | 1.0207 |
| src/compiler/utilities.ts:1:1 |  | 0.040678 | 0.041536 | +0.000858 | 1.0211 |
| src/compiler/core.ts:311:1 | clear | 0.039356 | 0.040213 | +0.000857 | 1.0218 |
| src/compiler/types.ts:2198:1 | ImportTypeAssertionContainer | 0.041217 | 0.042066 | +0.000849 | 1.0206 |
| src/compiler/checker.ts:54256:1 | signatureHasLiteralTypes | 0.039207 | 0.040051 | +0.000845 | 1.0215 |
| src/compiler/types.ts:7182:1 | IntraExpressionInferenceSite | 0.040747 | 0.041591 | +0.000845 | 1.0207 |
| src/compiler/commandLineParser.ts:3335:1 | isErrorNoInputFiles | 0.041197 | 0.042041 | +0.000844 | 1.0205 |
| src/compiler/factory/emitHelpers.ts:42:1 | PrivateIdentifierKind | 0.041590 | 0.042433 | +0.000843 | 1.0203 |
| src/compiler/utilities.ts:8068:1 | skipAlias | 0.081530 | 0.082371 | +0.000841 | 1.0103 |
| src/compiler/factory/nodeTests.ts:269:1 | isTemplateTail | 0.044404 | 0.045242 | +0.000838 | 1.0189 |
| src/compiler/utilities.ts:8091:1 | AccessKind | 0.040446 | 0.041282 | +0.000836 | 1.0207 |
| src/compiler/types.ts:6503:1 | StringLiteralType | 0.038706 | 0.039539 | +0.000833 | 1.0215 |
| src/compiler/types.ts:6155:1 | MappedSymbol | 0.038679 | 0.039512 | +0.000833 | 1.0215 |
| src/compiler/program.ts:328:1 | findConfigFile | 0.042585 | 0.043414 | +0.000829 | 1.0195 |
| src/compiler/utilities.ts:6241:1 | getJsxAttributeStringReplacement | 0.042715 | 0.043542 | +0.000828 | 1.0194 |
| src/compiler/checker.ts:1384:1 | MappedTypeNameTypeKind | 0.041652 | 0.042477 | +0.000825 | 1.0198 |
| src/compiler/utilities.ts:10962:1 | NodeModulePathParts | 0.039332 | 0.040156 | +0.000824 | 1.0210 |
| src/compiler/sys.ts:1968:1 | setSys | 0.038850 | 0.039666 | +0.000815 | 1.0210 |
| src/compiler/builder.ts:1118:1 | IncrementalMultiFileEmitBuildInfoFileInfo | 0.040578 | 0.041392 | +0.000814 | 1.0201 |
| src/compiler/factory/nodeFactory.ts:7394:1 | makeSynthetic | 0.047478 | 0.048292 | +0.000814 | 1.0171 |
| src/compiler/path.ts:606:1 | resolvePath | 0.074417 | 0.075230 | +0.000813 | 1.0109 |
| src/compiler/transformers/utilities.ts:496:1 | getNonAssignmentOperatorForCompoundAssignment | 0.041593 | 0.042405 | +0.000813 | 1.0195 |
| src/compiler/path.ts:494:1 | getPathComponents | 0.039199 | 0.040007 | +0.000808 | 1.0206 |
| src/compiler/types.ts:3350:1 | EmptyStatement | 0.044519 | 0.045317 | +0.000798 | 1.0179 |
| src/compiler/utilities.ts:7408:1 | isLogicalOperator | 0.049488 | 0.050284 | +0.000797 | 1.0161 |
| src/compiler/types.ts:5017:1 | ExitStatus | 0.037847 | 0.038641 | +0.000794 | 1.0210 |
| src/compiler/types.ts:2259:1 | TypeQueryNode | 0.040232 | 0.041023 | +0.000791 | 1.0197 |
| src/compiler/types.ts:6470:1 | IntrinsicType | 0.039871 | 0.040650 | +0.000779 | 1.0195 |
| src/compiler/types.ts:6657:1 | TypeReference | 0.038543 | 0.039318 | +0.000775 | 1.0201 |
| src/compiler/utilities.ts:7983:1 | rangeContainsRange | 0.038820 | 0.039593 | +0.000773 | 1.0199 |
| src/compiler/types.ts:3132:1 | CallLikeExpression | 0.042303 | 0.043072 | +0.000769 | 1.0182 |
| src/compiler/transformers/ts.ts:1:1 |  | 0.039940 | 0.040705 | +0.000765 | 1.0191 |
| src/compiler/types.ts:3430:1 | ForInStatement | 0.041687 | 0.042451 | +0.000764 | 1.0183 |
| src/compiler/moduleNameResolver.ts:216:1 | resolvedTypeScriptOnly | 0.042319 | 0.043080 | +0.000761 | 1.0180 |
| src/compiler/utilities.ts:6923:1 | isNonTypeAliasTemplate | 0.284998 | 0.285746 | +0.000748 | 1.0026 |
| src/compiler/utilities.ts:10191:1 | removeExtension | 0.040265 | 0.041012 | +0.000748 | 1.0186 |
| src/compiler/factory/utilities.ts:1076:1 | getElementsOfBindingOrAssignmentPattern | 0.042053 | 0.042799 | +0.000747 | 1.0178 |
| src/compiler/checker.ts:54223:1 | JsxNames | 0.039572 | 0.040316 | +0.000744 | 1.0188 |
| src/compiler/types.ts:9895:1 | RawSourceMap | 0.039798 | 0.040541 | +0.000744 | 1.0187 |
| src/compiler/transformers/esDecorators.ts:1:1 |  | 0.039813 | 0.040556 | +0.000743 | 1.0187 |
| src/compiler/checker.ts:54329:1 | SymbolTrackerImpl | 0.038924 | 0.039666 | +0.000742 | 1.0191 |
| src/compiler/transformers/utilities.ts:113:1 | ExternalModuleInfo | 0.039657 | 0.040399 | +0.000741 | 1.0187 |
| src/compiler/program.ts:4843:1 | shouldTransformImportCallWorker | 0.039913 | 0.040652 | +0.000739 | 1.0185 |
| src/compiler/builder.ts:95:1 | ReusableDiagnostic | 0.040280 | 0.041019 | +0.000739 | 1.0183 |
| src/compiler/types.ts:9725:1 | NodeVisitor | 0.041758 | 0.042495 | +0.000737 | 1.0177 |
| src/compiler/types.ts:6171:1 | InternalSymbolName | 0.039277 | 0.040012 | +0.000735 | 1.0187 |
| src/compiler/_namespaces/ts.ts:45:1 |  | 0.040983 | 0.041716 | +0.000734 | 1.0179 |
| src/compiler/emitter.ts:6336:1 | getOpeningBracket | 0.045079 | 0.045812 | +0.000733 | 1.0163 |
| src/compiler/factory/nodeTests.ts:927:1 | isJsxElement | 0.058162 | 0.058892 | +0.000730 | 1.0126 |
| src/compiler/utilities.ts:6530:1 | getExternalModuleNameFromPath | 0.198146 | 0.198875 | +0.000729 | 1.0037 |
| src/compiler/builder.ts:595:1 | convertToDiagnosticRelatedInformation | 0.119952 | 0.120678 | +0.000726 | 1.0061 |
| src/compiler/checker.ts:1199:1 | IterationTypeKind | 0.041073 | 0.041799 | +0.000726 | 1.0177 |
| src/compiler/path.ts:526:1 | normalizeSlashes | 0.038849 | 0.039570 | +0.000720 | 1.0185 |
| src/compiler/utilities.ts:8855:1 | compareMessageChain | 0.040143 | 0.040859 | +0.000716 | 1.0178 |
| src/compiler/utilities.ts:12298:1 | getPropertyArrayElementValue | 0.059942 | 0.060655 | +0.000713 | 1.0119 |
| src/compiler/parser.ts:503:1 | ForEachChildFunction | 0.039509 | 0.040222 | +0.000712 | 1.0180 |
| src/compiler/builder.ts:1076:1 | IncrementalBuildInfoFileIdListId | 0.039947 | 0.040655 | +0.000709 | 1.0177 |
| src/compiler/parser.ts:484:1 | getImportMetaIfNecessary | 0.050424 | 0.051127 | +0.000703 | 1.0139 |
| src/compiler/utilities.ts:6231:1 |  | 0.070125 | 0.070827 | +0.000702 | 1.0100 |
| src/compiler/_namespaces/ts.ts:41:1 |  | 0.041409 | 0.042110 | +0.000701 | 1.0169 |
| src/compiler/factory/utilities.ts:829:1 | tryRenameExternalModule | 0.095031 | 0.095730 | +0.000700 | 1.0074 |
| src/compiler/types.ts:5947:1 | SymbolFlags | 0.039232 | 0.039922 | +0.000691 | 1.0176 |
| src/compiler/commandLineParser.ts:3018:1 | parseJsonConfigFileContent | 0.038998 | 0.039687 | +0.000688 | 1.0177 |
| src/compiler/types.ts:2063:1 | FunctionLikeDeclarationBase | 0.043352 | 0.044036 | +0.000684 | 1.0158 |
| src/compiler/parser.ts:1399:1 | updateSourceFile | 0.039472 | 0.040153 | +0.000681 | 1.0172 |
| src/compiler/path.ts:373:1 | getBaseFileName | 0.040295 | 0.040968 | +0.000673 | 1.0167 |
| src/compiler/transformers/classFields.ts:230:1 | ClassPropertySubstitutionFlags | 0.042688 | 0.043359 | +0.000671 | 1.0157 |
| src/compiler/types.ts:2749:1 | ConciseBody | 0.042132 | 0.042803 | +0.000671 | 1.0159 |
| src/compiler/expressionToTypeNode.ts:170:1 |  | 0.045876 | 0.046545 | +0.000670 | 1.0146 |
| src/compiler/scanner.ts:426:1 | regularExpressionFlagToCharacterCode | 0.041544 | 0.042213 | +0.000670 | 1.0161 |
| src/compiler/parser.ts:423:1 |  | 0.040199 | 0.040867 | +0.000668 | 1.0166 |
| src/compiler/semver.ts:274:1 |  | 0.039548 | 0.040214 | +0.000666 | 1.0169 |
| src/compiler/builderPublic.ts:41:1 | HostForComputeHash | 0.044936 | 0.045598 | +0.000662 | 1.0147 |
| src/compiler/builder.ts:1035:1 | getSemanticDiagnosticsOfFile | 0.044000 | 0.044658 | +0.000658 | 1.0150 |
| src/compiler/utilities.ts:10807:1 | containsIgnoredPath | 0.041568 | 0.042226 | +0.000658 | 1.0158 |
| src/compiler/factory/nodeChildren.ts:27:1 | setNodeChildren | 0.049848 | 0.050503 | +0.000655 | 1.0131 |
| src/compiler/builder.ts:121:1 | SerializedDiagnosticMessageChain | 0.040228 | 0.040882 | +0.000653 | 1.0162 |
| src/compiler/types.ts:2721:1 | ObjectBindingOrAssignmentPattern | 0.041759 | 0.042411 | +0.000652 | 1.0156 |
| src/compiler/utilities.ts:5063:1 | isNodeDescendantOf | 0.047427 | 0.048078 | +0.000650 | 1.0137 |
| src/compiler/types.ts:7139:1 | InferenceFlags | 0.039806 | 0.040454 | +0.000649 | 1.0163 |
| src/compiler/moduleNameResolver.ts:1392:1 | getOptionsForLibraryResolution | 0.041338 | 0.041986 | +0.000647 | 1.0157 |
| src/compiler/programDiagnostics.ts:58:1 | FileReasonToChainCache | 0.040931 | 0.041572 | +0.000640 | 1.0156 |
| src/compiler/types.ts:2193:1 | KeywordTypeNode | 0.041176 | 0.041815 | +0.000639 | 1.0155 |
| src/compiler/types.ts:3457:1 | ReturnStatement | 0.041719 | 0.042358 | +0.000638 | 1.0153 |
| src/compiler/types.ts:2172:1 | ClassStaticBlockDeclaration | 0.041198 | 0.041832 | +0.000634 | 1.0154 |
| src/compiler/types.ts:5789:1 | RequireVariableDeclarationList | 0.038040 | 0.038668 | +0.000627 | 1.0165 |
| src/compiler/path.ts:751:1 | getPathWithoutRoot | 0.040705 | 0.041326 | +0.000620 | 1.0152 |
| src/compiler/utilities.ts:10016:1 | getSupportedExtensionsWithJsonIfResolveJsonModule | 0.038976 | 0.039596 | +0.000620 | 1.0159 |
| src/compiler/types.ts:2506:1 | YieldExpression | 0.041387 | 0.042004 | +0.000616 | 1.0149 |
| src/compiler/factory/utilities.ts:1112:1 | canHaveIllegalTypeParameters | 0.044802 | 0.045416 | +0.000613 | 1.0137 |
| src/compiler/types.ts:7097:1 | TypeMapper | 0.040132 | 0.040741 | +0.000609 | 1.0152 |
| src/compiler/utilities.ts:7276:1 | getEffectiveModifierFlags | 0.050699 | 0.051304 | +0.000605 | 1.0119 |
| src/compiler/tsbuildPublic.ts:249:1 | BuildOrder | 0.040000 | 0.040603 | +0.000603 | 1.0151 |
| src/compiler/sourcemap.ts:28:1 |  | 0.039854 | 0.040453 | +0.000599 | 1.0150 |
| src/compiler/types.ts:6204:1 | ReadonlyUnderscoreEscapedMap | 0.038396 | 0.038989 | +0.000593 | 1.0155 |
| src/compiler/_namespaces/ts.ts:19:1 |  | 0.039824 | 0.040410 | +0.000586 | 1.0147 |
| src/compiler/commandLineParser.ts:2157:1 |  | 0.039563 | 0.040145 | +0.000583 | 1.0147 |
| src/compiler/sys.ts:103:1 | HostWatchFile | 0.040397 | 0.040978 | +0.000580 | 1.0144 |
| src/compiler/utilities.ts:10300:1 | isAnySupportedFileExtension | 0.041732 | 0.042310 | +0.000579 | 1.0139 |
| src/compiler/scanner.ts:47:1 | tokenIsIdentifierOrKeywordOrGreaterThan | 0.044941 | 0.045518 | +0.000577 | 1.0128 |
| src/compiler/transformers/utilities.ts:598:1 | getStaticPropertiesAndClassStaticBlock | 0.040971 | 0.041547 | +0.000576 | 1.0140 |
| src/compiler/factory/nodeTests.ts:280:1 | isCommaToken | 0.045168 | 0.045738 | +0.000570 | 1.0126 |
| src/compiler/factory/parenthesizerRules.ts:54:1 | createParenthesizerRules | 0.135876 | 0.136435 | +0.000559 | 1.0041 |
| src/compiler/moduleNameResolver.ts:2399:1 | getPackageScopeForPath | 0.040292 | 0.040847 | +0.000555 | 1.0138 |
| src/compiler/moduleSpecifiers.ts:276:1 | getModuleSpecifier | 0.042682 | 0.043235 | +0.000553 | 1.0130 |
| src/compiler/commandLineParser.ts:4089:1 | validateSpecs | 0.090175 | 0.090727 | +0.000552 | 1.0061 |
| src/compiler/transformers/declarations/diagnostics.ts:111:1 | DeclarationDiagnosticProducing | 0.040861 | 0.041401 | +0.000540 | 1.0132 |
| src/compiler/builder.ts:1235:1 | getBuildInfo | 0.041626 | 0.042158 | +0.000533 | 1.0128 |
| src/compiler/types.ts:6321:1 | SerializedTypeEntry | 0.039472 | 0.040002 | +0.000530 | 1.0134 |
| src/compiler/types.ts:7164:1 | TypeComparer | 0.039878 | 0.040407 | +0.000529 | 1.0133 |
| src/compiler/builderStatePublic.ts:6:1 | EmitOutput | 0.042907 | 0.043430 | +0.000523 | 1.0122 |
| src/compiler/types.ts:2758:1 | ArrowFunction | 0.042721 | 0.043236 | +0.000514 | 1.0120 |
| src/compiler/utilities.ts:7955:1 | isNodeArrayMultiLine | 0.040142 | 0.040654 | +0.000513 | 1.0128 |
| src/compiler/utilities.ts:1370:1 | getInternalEmitFlags | 0.045842 | 0.046353 | +0.000512 | 1.0112 |
| src/compiler/factory/nodeTests.ts:235:1 | isNumericLiteral | 0.047066 | 0.047576 | +0.000510 | 1.0108 |
| src/compiler/transformers/taggedTemplate.ts:29:1 | ProcessLevel | 0.038782 | 0.039286 | +0.000503 | 1.0130 |
| src/compiler/commandLineParser.ts:2328:1 |  | 0.044639 | 0.045140 | +0.000501 | 1.0112 |
| src/compiler/builderState.ts:62:1 | BuilderState | 0.041216 | 0.041710 | +0.000494 | 1.0120 |
| src/compiler/utilities.ts:3531:1 | nodeIsDecorated | 0.065618 | 0.066112 | +0.000494 | 1.0075 |
| src/compiler/expressionToTypeNode.ts:169:1 |  | 0.046477 | 0.046967 | +0.000491 | 1.0106 |
| src/compiler/utilities.ts:6262:1 | stripQuotes | 0.040763 | 0.041252 | +0.000489 | 1.0120 |
| src/compiler/path.ts:538:1 | reducePathComponents | 0.076221 | 0.076709 | +0.000488 | 1.0064 |
| src/compiler/types.ts:2868:1 | LiteralToken | 0.042887 | 0.043373 | +0.000486 | 1.0113 |
| src/compiler/types.ts:3436:1 | ForOfStatement | 0.042041 | 0.042526 | +0.000485 | 1.0115 |
| src/compiler/builder.ts:1112:1 | IncrementalBuildInfoEmitSignature | 0.042199 | 0.042676 | +0.000476 | 1.0113 |
| src/compiler/types.ts:10347:1 | ConcretePragmaSpecs | 0.039147 | 0.039622 | +0.000475 | 1.0121 |
| src/compiler/parser.ts:10781:1 | getNamedPragmaArguments | 0.039889 | 0.040363 | +0.000473 | 1.0119 |
| src/compiler/types.ts:2539:1 | AdditiveOperatorOrHigher | 0.042342 | 0.042813 | +0.000471 | 1.0111 |
| src/compiler/types.ts:6477:1 | NullableType | 0.039912 | 0.040381 | +0.000470 | 1.0118 |
| src/compiler/factory/emitHelpers.ts:91:1 | ESDecorateName | 0.041048 | 0.041513 | +0.000465 | 1.0113 |
| src/compiler/factory/nodeTests.ts:676:1 | isAsExpression | 0.044493 | 0.044950 | +0.000457 | 1.0103 |
| src/compiler/commandLineParser.ts:3047:1 | directoryOfCombinedPath | 0.040001 | 0.040455 | +0.000454 | 1.0113 |
| src/compiler/transformer.ts:78:1 | getModuleTransformer | 0.072506 | 0.072959 | +0.000452 | 1.0062 |
| src/compiler/program.ts:342:1 | computeCommonSourceDirectoryOfFilenames | 0.040206 | 0.040655 | +0.000449 | 1.0112 |
| src/compiler/utilities.ts:8374:1 | getLeftmostAccessExpression | 0.251669 | 0.252110 | +0.000441 | 1.0018 |
| src/compiler/utilities.ts:10305:1 | tryGetExtensionFromPath | 0.040670 | 0.041111 | +0.000441 | 1.0108 |
| src/compiler/moduleSpecifiers.ts:1467:1 | prefersTsExtension | 0.039984 | 0.040424 | +0.000440 | 1.0110 |
| src/compiler/utilities.ts:10025:1 | isJSLike | 0.039553 | 0.039993 | +0.000440 | 1.0111 |
| src/compiler/types.ts:3405:1 | DoStatement | 0.049195 | 0.049624 | +0.000429 | 1.0087 |
| src/compiler/utilities.ts:3799:1 | isSourceFileJS | 0.044543 | 0.044970 | +0.000427 | 1.0096 |
| src/compiler/utilitiesPublic.ts:951:1 | getNonAssignedNameOfDeclaration | 0.042839 | 0.043260 | +0.000421 | 1.0098 |
| src/compiler/emitter.ts:724:1 | getFirstProjectOutput | 0.124303 | 0.124721 | +0.000418 | 1.0034 |
| src/compiler/utilities.ts:7970:1 | getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter | 0.039904 | 0.040321 | +0.000417 | 1.0105 |
| src/compiler/utilities.ts:3819:1 | isInJSDoc | 0.044392 | 0.044808 | +0.000416 | 1.0094 |
| src/compiler/utilities.ts:919:1 | typeDirectiveIsEqualTo | 0.039382 | 0.039794 | +0.000412 | 1.0105 |
| src/compiler/factory/emitHelpers.ts:1:1 |  | 0.042154 | 0.042562 | +0.000408 | 1.0097 |
| src/compiler/factory/nodeTests.ts:1104:1 | isJSDocClassTag | 0.044541 | 0.044948 | +0.000407 | 1.0091 |
| src/compiler/semver.ts:238:1 | Comparator | 0.042028 | 0.042434 | +0.000407 | 1.0097 |
| src/compiler/watchUtilities.ts:677:1 | WatchLogLevel | 0.039866 | 0.040268 | +0.000402 | 1.0101 |
| src/compiler/utilities.ts:6938:1 | emitNewLineBeforeLeadingComments | 0.040333 | 0.040736 | +0.000402 | 1.0100 |
| src/compiler/transformers/es2018.ts:122:1 | HierarchyFacts | 0.038041 | 0.038424 | +0.000383 | 1.0101 |
| src/compiler/binder.ts:465:1 | ContainerFlags | 0.041635 | 0.042011 | +0.000376 | 1.0090 |
| src/compiler/types.ts:9992:1 | HasCurrentDirectory | 0.039827 | 0.040199 | +0.000373 | 1.0094 |
| src/compiler/types.ts:5070:1 | WriterContextOut | 0.037695 | 0.038068 | +0.000372 | 1.0099 |
| src/compiler/types.ts:3790:1 | NamedImports | 0.043538 | 0.043908 | +0.000370 | 1.0085 |
| src/compiler/types.ts:3221:1 | JsxTagNamePropertyAccess | 0.041557 | 0.041922 | +0.000364 | 1.0088 |
| src/compiler/types.ts:1613:1 | QuestionQuestionEqualsToken | 0.054892 | 0.055254 | +0.000362 | 1.0066 |
| src/compiler/types.ts:2221:1 | FunctionOrConstructorTypeNode | 0.043394 | 0.043753 | +0.000359 | 1.0083 |
| src/compiler/utilities.ts:8951:1 | getLanguageVariant | 0.039961 | 0.040315 | +0.000354 | 1.0088 |
| src/compiler/utilities.ts:9992:1 | getSupportedExtensions | 0.038673 | 0.039025 | +0.000353 | 1.0091 |
| src/compiler/moduleSpecifiers.ts:1429:1 | tryGetJSExtensionForFile | 0.043654 | 0.044004 | +0.000350 | 1.0080 |
| src/compiler/types.ts:3860:1 | FileReference | 0.039087 | 0.039433 | +0.000347 | 1.0089 |
| src/compiler/utilities.ts:10595:1 | isPartOfPossiblyValidTypeOrAbstractComputedPropertyName | 0.053002 | 0.053345 | +0.000343 | 1.0065 |
| src/compiler/utilities.ts:3186:1 | getThisContainer | 0.038604 | 0.038947 | +0.000343 | 1.0089 |
| src/compiler/scanner.ts:1002:1 | EscapeSequenceScanningFlags | 0.039208 | 0.039551 | +0.000342 | 1.0087 |
| src/compiler/factory/nodeTests.ts:672:1 | isExpressionWithTypeArguments | 0.043392 | 0.043733 | +0.000341 | 1.0079 |
| src/compiler/transformers/classFields.ts:297:1 | PrivateEnvironmentData | 0.040957 | 0.041296 | +0.000339 | 1.0083 |
| src/compiler/utilities.ts:11199:1 | replaceFirstStar | 0.039222 | 0.039559 | +0.000337 | 1.0086 |
| src/compiler/utilities.ts:9975:1 |  | 0.042573 | 0.042910 | +0.000336 | 1.0079 |
| src/compiler/binder.ts:339:1 | getModuleInstanceStateCached | 0.050001 | 0.050335 | +0.000334 | 1.0067 |
| src/compiler/corePublic.ts:17:1 | SortedReadonlyArray | 0.038913 | 0.039245 | +0.000332 | 1.0085 |
| src/compiler/parser.ts:1408:1 | JsDocWithDiagnostics | 0.041219 | 0.041542 | +0.000322 | 1.0078 |
| src/compiler/utilities.ts:1209:1 | createCommentDirectivesMap | 0.039223 | 0.039545 | +0.000321 | 1.0082 |
| src/compiler/utilities.ts:2783:1 |  | 0.040115 | 0.040436 | +0.000321 | 1.0080 |
| src/compiler/program.ts:680:1 | ForegroundColorEscapeSequences | 0.041517 | 0.041838 | +0.000321 | 1.0077 |
| src/compiler/watchUtilities.ts:684:1 | WatchFactoryHost | 0.041015 | 0.041336 | +0.000320 | 1.0078 |
| src/compiler/tsbuildPublic.ts:256:1 | AnyBuildOrder | 0.040025 | 0.040340 | +0.000315 | 1.0079 |
| src/compiler/types.ts:6670:1 | DeferredTypeReference | 0.039598 | 0.039904 | +0.000305 | 1.0077 |
| src/compiler/types.ts:5036:1 | EmitResult | 0.038533 | 0.038837 | +0.000304 | 1.0079 |
| src/compiler/moduleNameResolver.ts:127:1 | withPackageId | 0.040465 | 0.040767 | +0.000302 | 1.0075 |
| src/compiler/scanner.ts:738:1 |  | 0.040408 | 0.040702 | +0.000294 | 1.0073 |
| src/compiler/commandLineParser.ts:4179:1 | getWildcardDirectoryFromSpec | 0.041742 | 0.042035 | +0.000294 | 1.0070 |
| src/compiler/commandLineParser.ts:3738:1 | convertCompilerOptionsFromJson | 0.043213 | 0.043506 | +0.000293 | 1.0068 |
| src/compiler/utilities.ts:984:1 | getSourceFileOfNode | 0.045980 | 0.046273 | +0.000292 | 1.0064 |
| src/compiler/utilities.ts:7450:1 | ClassImplementingOrExtendingExpressionWithTypeArguments | 0.040939 | 0.041230 | +0.000291 | 1.0071 |
| src/compiler/utilities.ts:3090:1 | isThisTypePredicate | 0.040377 | 0.040660 | +0.000283 | 1.0070 |
| src/compiler/types.ts:3038:1 | CallChainRoot | 0.041473 | 0.041755 | +0.000282 | 1.0068 |
| src/compiler/builderState.ts:1:1 |  | 0.039695 | 0.039972 | +0.000277 | 1.0070 |
| src/compiler/transformers/ts.ts:219:1 | ClassFacts | 0.038209 | 0.038485 | +0.000277 | 1.0072 |
| src/compiler/moduleSpecifiers.ts:798:1 | getAllRuntimeDependencies | 0.039984 | 0.040260 | +0.000276 | 1.0069 |
| src/compiler/moduleNameResolver.ts:1582:1 | tryLoadModuleUsingRootDirs | 0.040028 | 0.040295 | +0.000268 | 1.0067 |
| src/compiler/path.ts:28:1 |  | 0.040466 | 0.040731 | +0.000265 | 1.0066 |
| src/compiler/utilities.ts:2297:1 | isLateVisibilityPaintedStatement | 0.044616 | 0.044881 | +0.000265 | 1.0059 |
| src/compiler/factory/nodeTests.ts:1108:1 | isJSDocCallbackTag | 0.043122 | 0.043386 | +0.000264 | 1.0061 |
| src/compiler/types.ts:2897:1 | PseudoLiteralToken | 0.041023 | 0.041286 | +0.000264 | 1.0064 |
| src/compiler/utilities.ts:2124:1 | isGlobalScopeAugmentation | 0.040252 | 0.040514 | +0.000262 | 1.0065 |
| src/compiler/types.ts:3158:1 | AssertionExpression | 0.040360 | 0.040619 | +0.000259 | 1.0064 |
| src/compiler/utilities.ts:3804:1 | isInJSFile | 0.044460 | 0.044719 | +0.000259 | 1.0058 |
| src/compiler/utilities.ts:11196:1 |  | 0.040181 | 0.040433 | +0.000252 | 1.0063 |
| src/compiler/types.ts:7649:1 | NewLineKind | 0.041990 | 0.042234 | +0.000244 | 1.0058 |
| src/compiler/checker.ts:54240:1 | getIterationTypesKeyFromIterationTypeKind | 0.039295 | 0.039539 | +0.000244 | 1.0062 |
| src/compiler/scanner.ts:51:1 | Scanner | 0.039737 | 0.039979 | +0.000243 | 1.0061 |
| src/compiler/transformers/declarations.ts:1865:1 | isAlwaysType | 0.043099 | 0.043327 | +0.000228 | 1.0053 |
| src/compiler/moduleNameResolver.ts:339:1 | PackageJsonPathFields | 0.040851 | 0.041078 | +0.000227 | 1.0056 |
| src/compiler/utilities.ts:1062:1 | isFileLevelUniqueName | 0.041197 | 0.041418 | +0.000221 | 1.0054 |
| src/compiler/path.ts:125:1 | fileExtensionIsOneOf | 0.039229 | 0.039450 | +0.000221 | 1.0056 |
| src/compiler/_namespaces/ts.ts:35:1 |  | 0.041565 | 0.041785 | +0.000220 | 1.0053 |
| src/compiler/factory/nodeTests.ts:284:1 | isPlusToken | 0.045770 | 0.045989 | +0.000219 | 1.0048 |
| src/compiler/types.ts:2053:1 | ArrayBindingElement | 0.043865 | 0.044083 | +0.000218 | 1.0050 |
| src/compiler/transformers/namedEvaluation.ts:165:1 | classHasDeclaredOrExplicitlyAssignedName | 0.040197 | 0.040411 | +0.000215 | 1.0053 |
| src/compiler/utilities.ts:3529:1 | nodeIsDecorated | 0.056315 | 0.056530 | +0.000215 | 1.0038 |
| src/compiler/commandLineParser.ts:2210:1 | DiagnosticReporter | 0.039464 | 0.039676 | +0.000212 | 1.0054 |
| src/compiler/types.ts:3877:1 | SynthesizedComment | 0.051463 | 0.051672 | +0.000209 | 1.0041 |
| src/compiler/emitter.ts:6370:1 | emitListItemWithParenthesizerRule | 0.044184 | 0.044391 | +0.000207 | 1.0047 |
| src/compiler/checker.ts:1479:1 | isInstantiatedModule | 0.042408 | 0.042614 | +0.000206 | 1.0049 |
| src/compiler/transformer.ts:1:1 |  | 0.040826 | 0.041032 | +0.000206 | 1.0050 |
| src/compiler/types.ts:2472:1 | BooleanLiteral | 0.041526 | 0.041724 | +0.000198 | 1.0048 |
| src/compiler/transformers/typeSerializer.ts:94:1 | RuntimeTypeSerializer | 0.040492 | 0.040678 | +0.000186 | 1.0046 |
| src/compiler/utilitiesPublic.ts:1612:1 | isModifierKind | 0.043372 | 0.043557 | +0.000185 | 1.0043 |
| src/compiler/moduleSpecifiers.ts:379:1 | ModuleSpecifierResult | 0.040702 | 0.040877 | +0.000175 | 1.0043 |
| src/compiler/transformers/es2015.ts:220:1 | ES2015SubstitutionFlags | 0.039604 | 0.039778 | +0.000174 | 1.0044 |
| src/compiler/utilities.ts:10537:1 | parseBigInt | 0.038913 | 0.039084 | +0.000171 | 1.0044 |
| src/compiler/factory/utilities.ts:683:1 | getExternalHelpersModuleName | 0.044980 | 0.045143 | +0.000163 | 1.0036 |
| src/compiler/factory/utilities.ts:516:1 | expandPreOrPostfixIncrementOrDecrementExpression | 0.091838 | 0.091998 | +0.000160 | 1.0017 |
| src/compiler/utilities.ts:909:1 | packageIdToPackageName | 0.113907 | 0.114065 | +0.000158 | 1.0014 |
| src/compiler/utilities.ts:3475:1 | nodeCanBeDecorated | 0.041249 | 0.041403 | +0.000153 | 1.0037 |
| src/compiler/types.ts:5779:1 | VariableDeclarationInitializedTo | 0.038247 | 0.038393 | +0.000147 | 1.0038 |
| src/compiler/transformers/es2016.ts:21:1 | transformES2016 | 0.091017 | 0.091159 | +0.000142 | 1.0016 |
| src/compiler/utilities.ts:3922:1 | isAssignmentDeclaration | 0.303087 | 0.303226 | +0.000139 | 1.0005 |
| src/compiler/utilities.ts:5475:1 | AnonymousFunctionDefinition | 0.039843 | 0.039981 | +0.000138 | 1.0035 |
| src/compiler/commandLineParser.ts:2329:1 | getCommandLineCompilerOptionsMap | 0.042658 | 0.042785 | +0.000127 | 1.0030 |
| src/compiler/commandLineParser.ts:4214:1 | hasFileWithHigherPriorityExtension | 0.041432 | 0.041558 | +0.000127 | 1.0031 |
| src/compiler/tsbuildPublic.ts:1:1 |  | 0.040927 | 0.041053 | +0.000126 | 1.0031 |
| src/compiler/types.ts:3401:1 | IterationStatement | 0.044249 | 0.044364 | +0.000115 | 1.0026 |
| src/compiler/tracing.ts:1:1 |  | 0.040943 | 0.041056 | +0.000113 | 1.0028 |
| src/compiler/_namespaces/ts.ts:23:1 |  | 0.044354 | 0.044460 | +0.000106 | 1.0024 |
| src/compiler/core.ts:2271:1 | removeMinAndVersionNumbers | 0.041003 | 0.041107 | +0.000105 | 1.0026 |
| src/compiler/factory/nodeConverters.ts:183:1 |  | 0.067340 | 0.067443 | +0.000102 | 1.0015 |
| src/compiler/_namespaces/ts.ts:77:1 |  | 0.039237 | 0.039337 | +0.000101 | 1.0026 |
| src/compiler/types.ts:7296:1 | DiagnosticArguments | 0.039694 | 0.039792 | +0.000098 | 1.0025 |
| src/compiler/utilitiesPublic.ts:852:1 | unescapeLeadingUnderscores | 0.043579 | 0.043675 | +0.000096 | 1.0022 |
| src/compiler/sys.ts:894:1 | FsWatch | 0.040495 | 0.040588 | +0.000094 | 1.0023 |
| src/compiler/_namespaces/ts.ts:68:1 |  | 0.040819 | 0.040912 | +0.000093 | 1.0023 |
| src/compiler/moduleNameResolver.ts:3199:1 |  | 0.038890 | 0.038979 | +0.000089 | 1.0023 |
| src/compiler/_namespaces/ts.ts:40:1 |  | 0.041177 | 0.041265 | +0.000088 | 1.0021 |
| src/compiler/types.ts:2735:1 | ConditionalExpression | 0.042455 | 0.042543 | +0.000088 | 1.0021 |
| src/compiler/utilities.ts:3182:1 | getThisContainer | 0.041108 | 0.041183 | +0.000076 | 1.0018 |
| src/compiler/utilities.ts:8902:1 | compareMessageChainContent | 0.039199 | 0.039269 | +0.000070 | 1.0018 |
| src/compiler/moduleNameResolver.ts:1779:18 | nodeModuleNameResolver | 0.053998 | 0.054066 | +0.000068 | 1.0013 |
| src/compiler/core.ts:2375:1 | GetCanonicalFileName | 0.039531 | 0.039594 | +0.000063 | 1.0016 |
| src/compiler/utilities.ts:10056:1 | usesExtensionsOnImports | 0.089813 | 0.089874 | +0.000061 | 1.0007 |
| src/compiler/types.ts:2423:1 | PrefixUnaryOperator | 0.039449 | 0.039507 | +0.000058 | 1.0015 |
| src/compiler/utilitiesPublic.ts:1151:1 | getJSDocProtectedTagNoCache | 0.055776 | 0.055832 | +0.000055 | 1.0010 |
| src/compiler/types.ts:5654:1 | SymbolWalker | 0.037888 | 0.037943 | +0.000055 | 1.0014 |
| src/compiler/moduleSpecifiers.ts:318:1 | tryGetModuleSpecifiersFromCache | 0.043223 | 0.043273 | +0.000051 | 1.0012 |
| src/compiler/commandLineParser.ts:2723:1 | filterSameAsDefaultInclude | 0.040534 | 0.040583 | +0.000049 | 1.0012 |
| src/compiler/types.ts:2417:1 | UpdateExpression | 0.039393 | 0.039442 | +0.000048 | 1.0012 |
| src/compiler/types.ts:3366:1 | BlockLike | 0.043953 | 0.043997 | +0.000043 | 1.0010 |
| src/compiler/types.ts:3162:1 | NonNullExpression | 0.041829 | 0.041871 | +0.000043 | 1.0010 |
| src/compiler/semver.ts:207:1 | VersionRange | 0.043787 | 0.043827 | +0.000040 | 1.0009 |
| src/compiler/transformers/classFields.ts:3347:1 | PrivateIdentifierInExpression | 0.041447 | 0.041483 | +0.000035 | 1.0009 |
| src/compiler/performanceCore.ts:89:1 |  | 0.041229 | 0.041261 | +0.000031 | 1.0008 |
| src/compiler/path.ts:882:1 |  | 0.044333 | 0.044350 | +0.000017 | 1.0004 |
| src/compiler/factory/nodeChildren.ts:44:1 | unsetNodeChildren | 0.044167 | 0.044184 | +0.000017 | 1.0004 |
| src/compiler/types.ts:10091:1 | DiagnosticCollection | 0.041490 | 0.041502 | +0.000012 | 1.0003 |
| src/compiler/program.ts:4851:1 | getEmitModuleFormatOfFileWorker | 0.040353 | 0.040361 | +0.000007 | 1.0002 |
| src/compiler/types.ts:3211:1 | JsxAttributeName | 0.042981 | 0.042986 | +0.000005 | 1.0001 |
| src/compiler/program.ts:689:1 |  | 0.039332 | 0.039336 | +0.000004 | 1.0001 |
| src/compiler/types.ts:4214:1 | FlowAssignment | 0.040347 | 0.040351 | +0.000004 | 1.0001 |
| src/compiler/types.ts:2406:1 | PartiallyEmittedExpression | 0.041250 | 0.041253 | +0.000003 | 1.0001 |
| src/compiler/transformers/classFields.ts:342:1 | LexicalEnv | 0.039955 | 0.039954 | -0.000001 | 1.0000 |
| src/compiler/commandLineParser.ts:2637:1 | ConvertToTSConfigHost | 0.040245 | 0.040239 | -0.000005 | 0.9999 |
| src/compiler/types.ts:5823:1 | AllAccessorDeclarations | 0.038210 | 0.038203 | -0.000007 | 0.9998 |
| src/compiler/utilities.ts:10310:1 | isCheckJsEnabledForFile | 0.042503 | 0.042490 | -0.000012 | 0.9997 |
| src/compiler/transformers/classFields.ts:244:1 | PrivateIdentifierInfoBase | 0.039744 | 0.039729 | -0.000014 | 0.9996 |
| src/compiler/types.ts:2228:1 | FunctionTypeNode | 0.041561 | 0.041546 | -0.000015 | 0.9996 |
| src/compiler/types.ts:6640:1 | InterfaceTypeWithDeclaredMembers | 0.040349 | 0.040332 | -0.000017 | 0.9996 |
| src/compiler/factory/nodeTests.ts:1168:1 | isJSDocTypedefTag | 0.044539 | 0.044520 | -0.000019 | 0.9996 |
| src/compiler/utilities.ts:2643:1 | isJsonSourceFile | 0.040024 | 0.040003 | -0.000021 | 0.9995 |
| src/compiler/utilities.ts:1300:1 | moduleExportNameTextUnescaped | 0.042460 | 0.042435 | -0.000024 | 0.9994 |
| src/compiler/types.ts:4170:1 | FlowNode | 0.040869 | 0.040844 | -0.000025 | 0.9994 |
| src/compiler/types.ts:3033:1 | CallChain | 0.041868 | 0.041835 | -0.000033 | 0.9992 |
| src/compiler/utilities.ts:9987:1 |  | 0.040040 | 0.040006 | -0.000034 | 0.9992 |
| src/compiler/types.ts:5686:1 | TypePredicateKind | 0.038245 | 0.038203 | -0.000041 | 0.9989 |
| src/compiler/transformers/es2016.ts:1:1 |  | 0.042354 | 0.042311 | -0.000042 | 0.9990 |
| src/compiler/factory/nodeFactory.ts:7238:1 | getCookedText | 0.554585 | 0.554542 | -0.000043 | 0.9999 |
| src/compiler/utilities.ts:11157:1 | intrinsicTagNameToString | 0.208043 | 0.208000 | -0.000044 | 0.9998 |
| src/compiler/types.ts:2784:1 | LiteralExpression | 0.041372 | 0.041327 | -0.000045 | 0.9989 |
| src/compiler/transformers/es2015.ts:266:1 | LoopOutParameter | 0.040949 | 0.040903 | -0.000046 | 0.9989 |
| src/compiler/utilities.ts:6284:1 | getIndentString | 0.043402 | 0.043349 | -0.000053 | 0.9988 |
| src/compiler/utilitiesPublic.ts:376:1 | textRangeContainsTextSpan | 0.042384 | 0.042331 | -0.000053 | 0.9987 |
| src/compiler/core.ts:2386:1 | Pattern | 0.038374 | 0.038320 | -0.000054 | 0.9986 |
| src/compiler/utilities.ts:4482:1 | isJSDocTypeAlias | 0.048850 | 0.048788 | -0.000062 | 0.9987 |
| src/compiler/types.ts:1617:1 | MinusToken | 0.056745 | 0.056678 | -0.000068 | 0.9988 |
| src/compiler/transformers/esDecorators.ts:234:1 | ClassInfo | 0.044166 | 0.044090 | -0.000076 | 0.9983 |
| src/compiler/transformers/utilities.ts:547:1 | findSuperStatementIndexPathWorker | 0.040776 | 0.040689 | -0.000087 | 0.9979 |
| src/compiler/sourcemap.ts:662:1 | SourceMappedPosition | 0.040075 | 0.039985 | -0.000089 | 0.9978 |
| src/compiler/factory/nodeTests.ts:318:1 | isIdentifier | 0.046478 | 0.046385 | -0.000093 | 0.9980 |
| src/compiler/path.ts:69:1 | isDiskPathRoot | 0.038516 | 0.038417 | -0.000099 | 0.9974 |
| src/compiler/executeCommandLine.ts:754:1 | ExecuteCommandLineCallbacks | 0.041301 | 0.041199 | -0.000101 | 0.9975 |
| src/compiler/program.ts:1141:1 | getInferredLibraryNameResolveFrom | 0.038834 | 0.038733 | -0.000101 | 0.9974 |
| src/compiler/types.ts:1609:1 | ColonToken | 0.057340 | 0.057236 | -0.000104 | 0.9982 |
| src/compiler/moduleNameResolver.ts:2112:1 | loadFileNameFromPackageJsonField | 0.045070 | 0.044963 | -0.000106 | 0.9976 |
| src/compiler/factory/utilities.ts:1271:1 | BinaryExpressionState | 0.041369 | 0.041260 | -0.000108 | 0.9974 |
| src/compiler/program.ts:687:1 |  | 0.039679 | 0.039560 | -0.000119 | 0.9970 |
| src/compiler/utilities.ts:7236:1 | hasDecorators | 0.051463 | 0.051343 | -0.000119 | 0.9977 |
| src/compiler/emitter.ts:548:1 | getOutputPathWithoutChangingExt | 0.042811 | 0.042687 | -0.000124 | 0.9971 |
| src/compiler/types.ts:1606:1 | DotDotDotToken | 0.055062 | 0.054936 | -0.000125 | 0.9977 |
| src/compiler/factory/utilities.ts:1167:1 | isModuleName | 0.236460 | 0.236333 | -0.000127 | 0.9995 |
| src/compiler/builder.ts:1:1 |  | 0.040937 | 0.040790 | -0.000147 | 0.9964 |
| src/compiler/semver.ts:418:1 | testAlternative | 0.041891 | 0.041740 | -0.000151 | 0.9964 |
| src/compiler/program.ts:1075:1 |  | 0.096917 | 0.096764 | -0.000153 | 0.9984 |
| src/compiler/utilities.ts:5009:1 | walkUpParenthesizedTypes | 0.054735 | 0.054582 | -0.000153 | 0.9972 |
| src/compiler/types.ts:3225:1 | JsxAttributes | 0.040880 | 0.040725 | -0.000155 | 0.9962 |
| src/compiler/moduleSpecifiers.ts:809:1 | getAllModulePathsWorker | 0.039846 | 0.039687 | -0.000159 | 0.9960 |
| src/compiler/builderPublic.ts:46:1 | BuilderProgram | 0.045401 | 0.045239 | -0.000162 | 0.9964 |
| src/compiler/semver.ts:266:1 |  | 0.040189 | 0.040021 | -0.000168 | 0.9958 |
| src/compiler/tracing.ts:31:1 |  | 0.040745 | 0.040576 | -0.000169 | 0.9959 |
| src/compiler/program.ts:4881:1 | HostForUseSourceOfProjectReferenceRedirect | 0.042279 | 0.042109 | -0.000170 | 0.9960 |
| src/compiler/types.ts:5505:1 | UnionReduction | 0.037790 | 0.037618 | -0.000172 | 0.9954 |
| src/compiler/builder.ts:115:1 | ReusableRepopulateInfoChain | 0.040022 | 0.039847 | -0.000174 | 0.9956 |
| src/compiler/moduleNameResolver.ts:1271:1 | ModuleOrTypeReferenceResolutionCache | 0.041223 | 0.041047 | -0.000175 | 0.9957 |
| src/compiler/types.ts:3167:1 | NonNullChain | 0.040436 | 0.040254 | -0.000182 | 0.9955 |
| src/compiler/commandLineParser.ts:2310:1 |  | 0.074302 | 0.074117 | -0.000184 | 0.9975 |
| src/compiler/program.ts:1138:1 |  | 0.040100 | 0.039913 | -0.000187 | 0.9953 |
| src/compiler/moduleSpecifiers.ts:1424:1 | getJSExtensionForFile | 0.040424 | 0.040234 | -0.000189 | 0.9953 |
| src/compiler/factory/utilities.ts:697:1 | createExternalHelpersImportDeclarationIfNeeded | 0.075757 | 0.075560 | -0.000197 | 0.9974 |
| src/compiler/types.ts:10612:1 | EvaluationResolver | 0.039826 | 0.039628 | -0.000198 | 0.9950 |
| src/compiler/scanner.ts:979:1 | isIdentifierPart | 0.039846 | 0.039641 | -0.000205 | 0.9949 |
| src/compiler/types.ts:8350:1 | TabStop | 0.041269 | 0.041062 | -0.000207 | 0.9950 |
| src/compiler/utilities.ts:6716:1 | ensureDirectoriesExist | 0.041686 | 0.041478 | -0.000207 | 0.9950 |
| src/compiler/types.ts:4563:1 | ResolvedRefAndOutputDts | 0.040546 | 0.040338 | -0.000208 | 0.9949 |
| src/compiler/path.ts:433:1 | getAnyExtensionFromPath | 0.040859 | 0.040649 | -0.000211 | 0.9948 |
| src/compiler/utilities.ts:2503:1 | getCanonicalDiagnostic | 0.042565 | 0.042354 | -0.000211 | 0.9950 |
| src/compiler/transformers/classFields.ts:309:1 | UntransformedPrivateIdentifierInfo | 0.039588 | 0.039374 | -0.000214 | 0.9946 |
| src/compiler/factory/nodeTests.ts:304:1 | isColonToken | 0.047083 | 0.046869 | -0.000214 | 0.9955 |
| src/compiler/types.ts:7290:1 | CanonicalDiagnostic | 0.039697 | 0.039482 | -0.000215 | 0.9946 |
| src/compiler/utilities.ts:904:1 | packageIdIsEqual | 0.039783 | 0.039567 | -0.000216 | 0.9946 |
| src/compiler/builder.ts:1127:1 | IncrementalBuildInfoResolvedRoot | 0.041774 | 0.041558 | -0.000217 | 0.9948 |
| src/compiler/utilities.ts:6554:1 | getDeclarationEmitOutputFilePath | 0.081640 | 0.081421 | -0.000219 | 0.9973 |
| src/compiler/binder.ts:321:1 | ActiveLabel | 0.039641 | 0.039422 | -0.000219 | 0.9945 |
| src/compiler/semver.ts:297:1 | parsePartial | 0.041249 | 0.041022 | -0.000227 | 0.9945 |
| src/compiler/_namespaces/ts.ts:62:1 |  | 0.039411 | 0.039184 | -0.000227 | 0.9942 |
| src/compiler/path.ts:435:1 | getAnyExtensionFromPath | 0.040060 | 0.039826 | -0.000235 | 0.9941 |
| src/compiler/types.ts:10605:1 | Queue | 0.041742 | 0.041507 | -0.000235 | 0.9944 |
| src/compiler/types.ts:10421:1 | PragmaPseudoMap | 0.039880 | 0.039643 | -0.000236 | 0.9941 |
| src/compiler/commandLineParser.ts:2426:1 | JsonConversionNotifier | 0.041630 | 0.041390 | -0.000240 | 0.9942 |
| src/compiler/types.ts:9910:1 | SourceMapGenerator | 0.041522 | 0.041281 | -0.000240 | 0.9942 |
| src/compiler/types.ts:5663:1 | SymbolWriter | 0.038516 | 0.038272 | -0.000244 | 0.9937 |
| src/compiler/utilities.ts:7977:1 | getLinesBetweenPositionAndNextNonWhitespaceCharacter | 0.039764 | 0.039518 | -0.000246 | 0.9938 |
| src/compiler/utilities.ts:6060:1 | createDiagnosticCollection | 0.040313 | 0.040062 | -0.000251 | 0.9938 |
| src/compiler/sys.ts:1463:1 |  | 0.038984 | 0.038730 | -0.000254 | 0.9935 |
| src/compiler/_namespaces/ts.ts:73:1 |  | 0.040257 | 0.039999 | -0.000258 | 0.9936 |
| src/compiler/transformers/namedEvaluation.ts:294:1 | transformNamedEvaluationOfVariableDeclaration | 0.039465 | 0.039200 | -0.000264 | 0.9933 |
| src/compiler/scanner.ts:224:1 |  | 0.066788 | 0.066523 | -0.000266 | 0.9960 |
| src/compiler/factory/utilities.ts:575:1 | isExportName | 0.046339 | 0.046073 | -0.000267 | 0.9942 |
| src/compiler/factory/nodeTests.ts:1192:1 | isJSDocImportTag | 0.045288 | 0.045021 | -0.000267 | 0.9941 |
| src/compiler/types.ts:6268:1 | EvaluatorResult | 0.040436 | 0.040165 | -0.000272 | 0.9933 |
| src/compiler/types.ts:6607:1 | ObjectFlagsType | 0.042169 | 0.041897 | -0.000273 | 0.9935 |
| src/compiler/types.ts:9958:1 | DocumentPositionMapper | 0.039236 | 0.038963 | -0.000273 | 0.9930 |
| src/compiler/moduleNameResolver.ts:1161:1 | getOriginalOrResolvedTypeReferenceFileName | 0.042025 | 0.041747 | -0.000278 | 0.9934 |
| src/compiler/utilities.ts:10127:1 | getRequiresAtTopOfFile | 0.039926 | 0.039645 | -0.000281 | 0.9930 |
| src/compiler/types.ts:7633:1 | JsxEmit | 0.041890 | 0.041609 | -0.000281 | 0.9933 |
| src/compiler/transformers/declarations.ts:1:1 |  | 0.039805 | 0.039519 | -0.000286 | 0.9928 |
| src/compiler/utilitiesPublic.ts:1651:1 | isEntityName | 0.047370 | 0.047081 | -0.000289 | 0.9939 |
| src/compiler/types.ts:9809:1 | BuildInfoFileVersionMap | 0.041505 | 0.041214 | -0.000291 | 0.9930 |
| src/compiler/utilities.ts:3137:1 | getContainingClass | 0.044598 | 0.044307 | -0.000291 | 0.9935 |
| src/compiler/utilities.ts:6236:1 | encodeJsxCharacterEntity | 0.043326 | 0.043034 | -0.000292 | 0.9933 |
| src/compiler/scanner.ts:987:1 | isIdentifierText | 0.039506 | 0.039214 | -0.000292 | 0.9926 |
| src/compiler/utilitiesPublic.ts:1224:1 | getJSDocType | 0.055874 | 0.055581 | -0.000293 | 0.9948 |
| src/compiler/types.ts:2705:1 | BindingOrAssignmentElementRestIndicator | 0.042970 | 0.042677 | -0.000294 | 0.9932 |
| src/compiler/types.ts:6335:1 | TypeFlags | 0.039717 | 0.039423 | -0.000294 | 0.9926 |
| src/compiler/scanner.ts:4052:1 | utf16EncodeAsStringFallback | 0.042736 | 0.042442 | -0.000294 | 0.9931 |
| src/compiler/path.ts:841:1 | changeAnyExtension | 0.043361 | 0.043066 | -0.000295 | 0.9932 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts:1:1 |  | 0.039053 | 0.038756 | -0.000297 | 0.9924 |
| src/compiler/types.ts:7105:1 | InferencePriority | 0.039908 | 0.039611 | -0.000297 | 0.9926 |
| src/compiler/path.ts:1:1 |  | 0.041577 | 0.041278 | -0.000300 | 0.9928 |
| src/compiler/types.ts:2416:1 | IncrementExpression | 0.040654 | 0.040349 | -0.000305 | 0.9925 |
| src/compiler/types.ts:3191:1 | JsxElement | 0.042342 | 0.042033 | -0.000308 | 0.9927 |
| src/compiler/factory/utilitiesPublic.ts:1:1 |  | 0.039549 | 0.039237 | -0.000312 | 0.9921 |
| src/compiler/utilities.ts:8872:1 | compareMessageChainSize | 0.039784 | 0.039470 | -0.000314 | 0.9921 |
| src/compiler/types.ts:5776:1 | RequireOrImportCall | 0.038857 | 0.038543 | -0.000314 | 0.9919 |
| src/compiler/scanner.ts:800:1 | scanShebangTrivia | 0.040951 | 0.040637 | -0.000315 | 0.9923 |
| src/compiler/_namespaces/ts.ts:69:1 |  | 0.040907 | 0.040591 | -0.000316 | 0.9923 |
| src/compiler/types.ts:3042:1 | OptionalChain | 0.041675 | 0.041354 | -0.000321 | 0.9923 |
| src/compiler/types.ts:9660:1 | TransformationResult | 0.042053 | 0.041724 | -0.000328 | 0.9922 |
| src/compiler/core.ts:2589:80 |  | 0.039766 | 0.039438 | -0.000329 | 0.9917 |
| src/compiler/types.ts:7156:1 | Ternary | 0.041285 | 0.040950 | -0.000335 | 0.9919 |
| src/compiler/moduleNameResolver.ts:1712:1 | node16ModuleNameResolver | 0.044412 | 0.044077 | -0.000335 | 0.9925 |
| src/compiler/types.ts:2491:1 | TypeOfExpression | 0.042572 | 0.042235 | -0.000336 | 0.9921 |
| src/compiler/types.ts:7643:1 | ImportsNotUsedAsValues | 0.044244 | 0.043903 | -0.000341 | 0.9923 |
| src/compiler/types.ts:2431:1 | PrefixUnaryExpression | 0.041137 | 0.040794 | -0.000343 | 0.9917 |
| src/compiler/commandLineParser.ts:3572:1 | parseOwnConfigOfJsonSourceFile | 0.044998 | 0.044649 | -0.000349 | 0.9922 |
| src/compiler/utilities.ts:6282:1 |  | 0.041555 | 0.041205 | -0.000350 | 0.9916 |
| src/compiler/factory/utilities.ts:760:1 | getImportedHelpers | 0.042060 | 0.041697 | -0.000363 | 0.9914 |
| src/compiler/types.ts:6778:1 | AnonymousType | 0.040106 | 0.039743 | -0.000363 | 0.9910 |
| src/compiler/scanner.ts:409:1 |  | 0.042347 | 0.041975 | -0.000372 | 0.9912 |
| src/compiler/core.ts:2053:1 | compareStringsCaseInsensitiveEslintCompatible | 0.039901 | 0.039529 | -0.000373 | 0.9907 |
| src/compiler/types.ts:10446:1 | CommentDirectivesMap | 0.041808 | 0.041435 | -0.000373 | 0.9911 |
| src/compiler/types.ts:7087:1 | TypeMapKind | 0.040697 | 0.040320 | -0.000377 | 0.9907 |
| src/compiler/types.ts:5626:1 | SymbolFormatFlags | 0.038226 | 0.037848 | -0.000378 | 0.9901 |
| src/compiler/utilities.ts:6249:1 | escapeJsxAttributeString | 0.041031 | 0.040652 | -0.000380 | 0.9907 |
| src/compiler/types.ts:5574:1 | InternalNodeBuilderFlags | 0.038388 | 0.038007 | -0.000380 | 0.9901 |
| src/compiler/factory/nodeTests.ts:312:1 | isEqualsGreaterThanToken | 0.046031 | 0.045648 | -0.000383 | 0.9917 |
| src/compiler/utilities.ts:2669:1 | isVarUsing | 0.046679 | 0.046295 | -0.000384 | 0.9918 |
| src/compiler/types.ts:7168:1 | InferenceContext | 0.040234 | 0.039849 | -0.000385 | 0.9904 |
| src/compiler/tsbuild.ts:41:1 | UpToDateStatus | 0.040050 | 0.039663 | -0.000387 | 0.9903 |
| src/compiler/builder.ts:1584:1 | getBuilderCreationParameters | 0.041844 | 0.041455 | -0.000388 | 0.9907 |
| src/compiler/_namespaces/ts.ts:47:1 |  | 0.043502 | 0.043113 | -0.000389 | 0.9911 |
| src/compiler/_namespaces/ts.ts:61:1 |  | 0.039325 | 0.038936 | -0.000389 | 0.9901 |
| src/compiler/types.ts:9706:1 | Transformer | 0.042963 | 0.042574 | -0.000389 | 0.9909 |
| src/compiler/types.ts:2523:1 | MultiplicativeOperator | 0.040544 | 0.040155 | -0.000389 | 0.9904 |
| src/compiler/types.ts:6785:1 | InstantiationExpressionType | 0.040683 | 0.040285 | -0.000398 | 0.9902 |
| src/compiler/types.ts:10039:1 | ModuleSpecifierOptions | 0.041091 | 0.040693 | -0.000398 | 0.9903 |
| src/compiler/utilities.ts:3127:1 | getContainingFunction | 0.045542 | 0.045143 | -0.000399 | 0.9912 |
| src/compiler/emitter.ts:746:1 | emitResolverSkipsTypeChecking | 0.042335 | 0.041933 | -0.000402 | 0.9905 |
| src/compiler/parser.ts:441:1 |  | 0.044449 | 0.044047 | -0.000402 | 0.9909 |
| src/compiler/utilities.ts:1377:1 | ScriptTargetFeatures | 0.041266 | 0.040857 | -0.000408 | 0.9901 |
| src/compiler/path.ts:91:1 | pathIsAbsolute | 0.039182 | 0.038768 | -0.000414 | 0.9894 |
| src/compiler/types.ts:9770:1 | Printer | 0.041057 | 0.040642 | -0.000415 | 0.9899 |
| src/compiler/factory/nodeFactory.ts:481:1 | addNodeFactoryPatcher | 0.043002 | 0.042587 | -0.000416 | 0.9903 |
| src/compiler/factory/nodeTests.ts:680:1 | isSatisfiesExpression | 0.045270 | 0.044854 | -0.000416 | 0.9908 |
| src/compiler/utilities.ts:982:1 | getSourceFileOfNode | 0.045775 | 0.045358 | -0.000417 | 0.9909 |
| src/compiler/emitter.ts:431:1 | isBuildInfoFile | 0.040784 | 0.040366 | -0.000418 | 0.9897 |
| src/compiler/emitter.ts:1:1 |  | 0.040477 | 0.040058 | -0.000418 | 0.9897 |
| src/compiler/types.ts:5902:1 | EmitResolver | 0.039760 | 0.039340 | -0.000420 | 0.9894 |
| src/compiler/parser.ts:490:1 | walkTreeForImportMeta | 0.059670 | 0.059245 | -0.000425 | 0.9929 |
| src/compiler/utilities.ts:6911:1 | getEffectiveReturnTypeNode | 0.194280 | 0.193843 | -0.000437 | 0.9978 |
| src/compiler/types.ts:6816:1 | ResolvedType | 0.041417 | 0.040977 | -0.000440 | 0.9894 |
| src/compiler/utilities.ts:7988:1 | startEndContainsRange | 0.041401 | 0.040952 | -0.000448 | 0.9892 |
| src/compiler/tsbuildPublic.ts:209:1 | ReportEmitErrorSummary | 0.041918 | 0.041464 | -0.000453 | 0.9892 |
| src/compiler/checker.ts:1205:1 | IterationTypesResolver | 0.041944 | 0.041489 | -0.000455 | 0.9892 |
| src/compiler/executeCommandLine.ts:90:1 |  | 0.041314 | 0.040855 | -0.000459 | 0.9889 |
| src/compiler/factory/nodeTests.ts:255:1 | isNoSubstitutionTemplateLiteral | 0.046500 | 0.046041 | -0.000459 | 0.9901 |
| src/compiler/types.ts:2711:1 | BindingOrAssignmentElementTarget | 0.043064 | 0.042604 | -0.000461 | 0.9893 |
| src/compiler/transformers/declarations.ts:1934:1 | ProcessedComponent | 0.039306 | 0.038840 | -0.000466 | 0.9881 |
| src/compiler/commandLineParser.ts:2472:1 | convertToObject | 0.042026 | 0.041559 | -0.000467 | 0.9889 |
| src/compiler/utilities.ts:873:1 | createModeMismatchDetails | 0.040725 | 0.040257 | -0.000468 | 0.9885 |
| src/compiler/moduleNameResolver.ts:961:1 | getKeyForCompilerOptions | 0.043263 | 0.042795 | -0.000468 | 0.9892 |
| src/compiler/builder.ts:2254:1 | createBuilderProgramUsingIncrementalBuildInfo | 0.091647 | 0.091170 | -0.000477 | 0.9948 |
| src/compiler/factory/utilities.ts:1611:1 | createAccessorPropertyBackingField | 0.040504 | 0.040019 | -0.000485 | 0.9880 |
| src/compiler/utilities.ts:2107:1 | isShorthandAmbientModuleSymbol | 0.049430 | 0.048945 | -0.000485 | 0.9902 |
| src/compiler/program.ts:857:1 | getModeForFileReference | 4.865339 | 4.864851 | -0.000488 | 0.9999 |
| src/compiler/types.ts:5520:1 | ContextFlags | 0.038140 | 0.037647 | -0.000493 | 0.9871 |
| src/compiler/factory/nodeTests.ts:995:1 | isPropertyAssignment | 0.052295 | 0.051801 | -0.000494 | 0.9905 |
| src/compiler/utilities.ts:3319:1 | SuperContainerOrFunctions | 0.041079 | 0.040584 | -0.000496 | 0.9879 |
| src/compiler/scanner.ts:4087:1 |  | 0.039449 | 0.038953 | -0.000496 | 0.9874 |
| src/compiler/parser.ts:10576:1 | parseResolutionMode | 0.039584 | 0.039086 | -0.000498 | 0.9874 |
| src/compiler/corePublic.ts:37:1 | EqualityComparer | 0.039519 | 0.039021 | -0.000498 | 0.9874 |
| src/compiler/utilities.ts:8508:1 | Node | 0.046082 | 0.045581 | -0.000501 | 0.9891 |
| src/compiler/utilities.ts:8184:1 | MutateMapSkippingNewValuesOptions | 0.040038 | 0.039533 | -0.000504 | 0.9874 |
| src/compiler/utilities.ts:7484:1 | isDestructuringAssignment | 0.179109 | 0.178605 | -0.000505 | 0.9972 |
| src/compiler/types.ts:2788:1 | RegularExpressionLiteral | 0.044132 | 0.043626 | -0.000506 | 0.9885 |
| src/compiler/types.ts:4568:1 | ResolvedRefAndSource | 0.041093 | 0.040587 | -0.000506 | 0.9877 |
| src/compiler/transformers/utilities.ts:602:1 | getStaticPropertiesAndClassStaticBlock | 0.041283 | 0.040777 | -0.000506 | 0.9877 |
| src/compiler/utilitiesPublic.ts:1091:1 | getJSDocTypeParameterTags | 0.050855 | 0.050349 | -0.000506 | 0.9900 |
| src/compiler/utilities.ts:8500:1 | Signature | 0.040163 | 0.039652 | -0.000511 | 0.9873 |
| src/compiler/moduleSpecifiers.ts:558:1 | isExcludedByRegex | 0.050313 | 0.049800 | -0.000513 | 0.9898 |
| src/compiler/transformers/classFields.ts:3327:1 | createPrivateInstanceFieldInitializer | 0.184865 | 0.184350 | -0.000515 | 0.9972 |
| src/compiler/utilitiesPublic.ts:946:1 | isNamedDeclaration | 0.171186 | 0.170670 | -0.000515 | 0.9970 |
| src/compiler/moduleNameResolver.ts:2217:1 | loadNodeModuleFromDirectory | 0.045179 | 0.044661 | -0.000517 | 0.9885 |
| src/compiler/types.ts:8683:1 | ParenthesizerRules | 0.040468 | 0.039949 | -0.000520 | 0.9872 |
| src/compiler/commandLineParser.ts:3685:1 | ExtendedConfigCacheEntry | 0.042669 | 0.042138 | -0.000532 | 0.9875 |
| src/compiler/types.ts:10681:1 | IntroducesNewScopeNode | 0.039616 | 0.039084 | -0.000532 | 0.9866 |
| src/compiler/transformers/utilities.ts:587:1 | getProperties | 0.043518 | 0.042986 | -0.000532 | 0.9878 |
| src/compiler/commandLineParser.ts:2976:1 | convertToOptionsWithAbsolutePaths | 0.041802 | 0.041264 | -0.000538 | 0.9871 |
| src/compiler/types.ts:2520:1 | ExponentiationOperator | 0.042456 | 0.041908 | -0.000548 | 0.9871 |
| src/compiler/core.ts:1266:1 | hasProperty | 0.047035 | 0.046487 | -0.000548 | 0.9883 |
| src/compiler/utilities.ts:1203:1 | isPinnedComment | 0.039412 | 0.038863 | -0.000549 | 0.9861 |
| src/compiler/utilities.ts:12233:1 | getLibFileNameFromLibReference | 0.040838 | 0.040289 | -0.000550 | 0.9865 |
| src/compiler/utilities.ts:5625:1 | nodeIsSynthesized | 0.039764 | 0.039213 | -0.000551 | 0.9861 |
| src/compiler/types.ts:6144:1 | TransientSymbol | 0.038883 | 0.038331 | -0.000552 | 0.9858 |
| src/compiler/transformers/destructuring.ts:1:1 |  | 0.039597 | 0.039045 | -0.000552 | 0.9861 |
| src/compiler/moduleNameResolver.ts:2708:1 | loadModuleFromExportsOrImports | 0.042572 | 0.042017 | -0.000555 | 0.9870 |
| src/compiler/types.ts:8127:1 | ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 0.042841 | 0.042283 | -0.000558 | 0.9870 |
| src/compiler/utilities.ts:5703:1 | OperatorPrecedence | 0.043046 | 0.042484 | -0.000562 | 0.9869 |
| src/compiler/types.ts:7585:1 | WatchOptions | 0.043372 | 0.042803 | -0.000568 | 0.9869 |
| src/compiler/utilitiesPublic.ts:1024:1 | getDecorators | 0.051896 | 0.051327 | -0.000569 | 0.9890 |
| src/compiler/executeCommandLine.ts:1060:1 | enableSolutionPerformance | 0.041493 | 0.040923 | -0.000570 | 0.9863 |
| src/compiler/core.ts:2086:1 |  | 0.039585 | 0.039009 | -0.000576 | 0.9855 |
| src/compiler/factory/nodeTests.ts:261:1 | isTemplateHead | 0.047032 | 0.046454 | -0.000578 | 0.9877 |
| src/compiler/moduleNameResolver.ts:1750:1 | tryResolveJSModuleWorker | 0.044412 | 0.043829 | -0.000583 | 0.9869 |
| src/compiler/types.ts:2529:1 | MultiplicativeOperatorOrHigher | 0.042619 | 0.042036 | -0.000583 | 0.9863 |
| src/compiler/factory/utilities.ts:1262:1 | isBinaryOperator | 0.074911 | 0.074320 | -0.000591 | 0.9921 |
| src/compiler/types.ts:9987:1 | GetEffectiveTypeRootsHost | 0.039803 | 0.039208 | -0.000595 | 0.9851 |
| src/compiler/factory/utilities.ts:1586:1 | formatGeneratedName | 0.043281 | 0.042676 | -0.000605 | 0.9860 |
| src/compiler/transformers/namedEvaluation.ts:322:1 | transformNamedEvaluationOfParameterDeclaration | 0.039351 | 0.038743 | -0.000608 | 0.9846 |
| src/compiler/parser.ts:10561:1 | getDeclarationFileExtension | 0.039784 | 0.039177 | -0.000608 | 0.9847 |
| src/compiler/commandLineParser.ts:1928:1 | getOptionName | 0.041296 | 0.040682 | -0.000613 | 0.9851 |
| src/compiler/semver.ts:20:1 |  | 0.041427 | 0.040813 | -0.000614 | 0.9852 |
| src/compiler/moduleNameResolver.ts:3006:1 | loadModuleFromNearestNodeModulesDirectoryWorker | 0.039594 | 0.038975 | -0.000619 | 0.9844 |
| src/compiler/sourcemap.ts:654:1 | MappedPosition | 0.039628 | 0.039003 | -0.000625 | 0.9842 |
| src/compiler/types.ts:10055:1 | SymbolTracker | 0.041416 | 0.040791 | -0.000626 | 0.9849 |
| src/compiler/utilities.ts:9028:1 | usesWildcardTypes | 0.041313 | 0.040686 | -0.000627 | 0.9848 |
| src/compiler/types.ts:6319:1 | TrackedSymbol | 0.039369 | 0.038741 | -0.000628 | 0.9841 |
| src/compiler/types.ts:6277:1 | NodeLinks | 0.040395 | 0.039764 | -0.000632 | 0.9844 |
| src/compiler/factory/emitHelpers.ts:85:1 | ESDecorateClassElementAccess | 0.042027 | 0.041394 | -0.000633 | 0.9849 |
| src/compiler/transformers/es2019.ts:17:1 | transformES2019 | 0.077766 | 0.077126 | -0.000639 | 0.9918 |
| src/compiler/path.ts:30:1 |  | 0.041243 | 0.040601 | -0.000642 | 0.9844 |
| src/compiler/transformers/declarations.ts:1891:1 | canHaveLiteralInitializer | 0.046764 | 0.046121 | -0.000644 | 0.9862 |
| src/compiler/utilities.ts:10049:1 | ModuleSpecifierEnding | 0.039543 | 0.038894 | -0.000650 | 0.9836 |
| src/compiler/factory/nodeTests.ts:636:1 | isPrefixUnaryExpression | 0.044862 | 0.044212 | -0.000651 | 0.9855 |
| src/compiler/utilities.ts:10453:1 | isJsonEqual | 0.043022 | 0.042368 | -0.000654 | 0.9848 |
| src/compiler/utilities.ts:6764:1 | getSetAccessorValueParameter | 0.040844 | 0.040190 | -0.000654 | 0.9840 |
| src/compiler/types.ts:8027:1 | ResolvedModule | 0.043117 | 0.042454 | -0.000663 | 0.9846 |
| src/compiler/commandLineParser.ts:1884:1 | parseListTypeOption | 0.047944 | 0.047277 | -0.000666 | 0.9861 |
| src/compiler/factory/nodeTests.ts:265:1 | isTemplateMiddle | 0.044979 | 0.044313 | -0.000666 | 0.9852 |
| src/compiler/factory/nodeTests.ts:584:1 | isPropertyAccessExpression | 0.045511 | 0.044830 | -0.000681 | 0.9850 |
| src/compiler/utilities.ts:2428:1 | createDiagnosticForNodeInSourceFile | 0.050422 | 0.049738 | -0.000684 | 0.9864 |
| src/compiler/factory/nodeTests.ts:322:1 | isPrivateIdentifier | 0.047063 | 0.046372 | -0.000690 | 0.9853 |
| src/compiler/utilities.ts:6270:1 | isQuoteOrBacktick | 0.041622 | 0.040930 | -0.000692 | 0.9834 |
| src/compiler/_namespaces/ts.ts:20:1 |  | 0.041655 | 0.040963 | -0.000692 | 0.9834 |
| src/compiler/path.ts:115:1 | hasExtension | 0.039241 | 0.038544 | -0.000697 | 0.9822 |
| src/compiler/commandLineParser.ts:1872:1 | createDiagnosticForInvalidCustomType | 0.040564 | 0.039863 | -0.000701 | 0.9827 |
| src/compiler/binder.ts:330:1 | getModuleInstanceState | 0.044155 | 0.043452 | -0.000703 | 0.9841 |
| src/compiler/builder.ts:933:1 | handleDtsMayChangeOfReferencingExportOfAffectedFile | 0.135278 | 0.134572 | -0.000706 | 0.9948 |
| src/compiler/_namespaces/ts.ts:15:1 |  | 0.040670 | 0.039964 | -0.000706 | 0.9826 |
| src/compiler/utilities.ts:6754:1 | getLineOfLocalPositionFromLineMap | 0.041891 | 0.041184 | -0.000706 | 0.9831 |
| src/compiler/utilities.ts:9604:1 | escapeRegExpCharacter | 0.092758 | 0.092050 | -0.000709 | 0.9924 |
| src/compiler/utilities.ts:4460:1 | hasQuestionToken | 0.046184 | 0.045472 | -0.000713 | 0.9846 |
| src/compiler/utilities.ts:8280:1 | getObjectFlags | 0.042745 | 0.042032 | -0.000713 | 0.9833 |
| src/compiler/types.ts:3090:1 | BindableStaticPropertyAssignmentExpression | 0.043317 | 0.042595 | -0.000722 | 0.9833 |
| src/compiler/moduleNameResolver.ts:2026:1 |  | 0.041860 | 0.041130 | -0.000730 | 0.9826 |
| src/compiler/checker.ts:1172:1 |  | 0.041864 | 0.041130 | -0.000735 | 0.9825 |
| src/compiler/transformers/destructuring.ts:624:1 | makeAssignmentElement | 0.042255 | 0.041514 | -0.000741 | 0.9825 |
| src/compiler/types.ts:3085:1 | BindableAccessExpression | 0.042241 | 0.041499 | -0.000741 | 0.9825 |
| src/compiler/transformers/declarations.ts:1890:1 | CanHaveLiteralInitializer | 0.039225 | 0.038479 | -0.000747 | 0.9810 |
| src/compiler/_namespaces/ts.ts:75:1 |  | 0.040162 | 0.039413 | -0.000748 | 0.9814 |
| src/compiler/transformers/ts.ts:211:1 | TypeScriptSubstitutionFlags | 0.039351 | 0.038597 | -0.000754 | 0.9808 |
| src/compiler/utilities.ts:12168:1 |  | 0.042707 | 0.041951 | -0.000756 | 0.9823 |
| src/compiler/checker.ts:1152:1 |  | 0.046213 | 0.045454 | -0.000759 | 0.9836 |
| src/compiler/core.ts:2377:1 | createGetCanonicalFileName | 0.040036 | 0.039273 | -0.000763 | 0.9810 |
| src/compiler/transformer.ts:232:1 | noEmitNotification | 0.046404 | 0.045641 | -0.000763 | 0.9836 |
| src/compiler/types.ts:3120:1 | TaggedTemplateExpression | 0.041811 | 0.041039 | -0.000772 | 0.9815 |
| src/compiler/transformers/ts.ts:209:1 |  | 0.039598 | 0.038820 | -0.000778 | 0.9804 |
| src/compiler/corePublic.ts:30:1 | ReadonlyCollection | 0.039877 | 0.039098 | -0.000779 | 0.9805 |
| src/compiler/moduleNameResolver.ts:2072:1 | loadModuleFromFile | 0.044513 | 0.043731 | -0.000782 | 0.9824 |
| src/compiler/types.ts:7030:1 | Signature | 0.040868 | 0.040079 | -0.000789 | 0.9807 |
| src/compiler/moduleNameResolver.ts:975:1 | RedirectsCacheKey | 0.043256 | 0.042465 | -0.000790 | 0.9817 |
| src/compiler/tsbuildPublic.ts:268:1 | SolutionBuilder | 0.041896 | 0.041105 | -0.000791 | 0.9811 |
| src/compiler/checker.ts:1326:1 | TypeSystemPropertyName | 0.042496 | 0.041695 | -0.000801 | 0.9812 |
| src/compiler/_namespaces/ts.ts:53:1 |  | 0.041782 | 0.040981 | -0.000801 | 0.9808 |
| src/compiler/transformers/utilities.ts:490:1 | isCompoundAssignment | 0.041292 | 0.040484 | -0.000808 | 0.9804 |
| src/compiler/types.ts:3846:1 | TypeOnlyAliasDeclaration | 0.039649 | 0.038840 | -0.000809 | 0.9796 |
| src/compiler/parser.ts:469:1 | isFileProbablyExternalModule | 0.051469 | 0.050658 | -0.000810 | 0.9843 |
| src/compiler/factory/nodeTests.ts:247:1 | isJsxText | 0.045674 | 0.044858 | -0.000816 | 0.9821 |
| src/compiler/types.ts:2864:1 | BigIntLiteral | 0.043412 | 0.042596 | -0.000816 | 0.9812 |
| src/compiler/utilities.ts:7353:1 | modifiersToFlags | 0.061330 | 0.060511 | -0.000819 | 0.9866 |
| src/compiler/moduleNameResolver.ts:410:1 | readPackageJsonTypesVersionsField | 0.041929 | 0.041108 | -0.000821 | 0.9804 |
| src/compiler/factory/baseNodeFactory.ts:26:1 | createBaseNodeFactory | 0.088639 | 0.087811 | -0.000828 | 0.9907 |
| src/compiler/utilities.ts:5025:1 | walkUpParenthesizedTypesAndGetParentAndChild | 0.165699 | 0.164868 | -0.000831 | 0.9950 |
| src/compiler/types.ts:4220:1 | FlowCall | 0.041180 | 0.040331 | -0.000849 | 0.9794 |
| src/compiler/types.ts:3255:1 | JsxFragment | 0.042218 | 0.041367 | -0.000851 | 0.9798 |
| src/compiler/_namespaces/ts.ts:4:1 |  | 0.040367 | 0.039513 | -0.000853 | 0.9789 |
| src/compiler/types.ts:7378:1 | ProjectReference | 0.043475 | 0.042620 | -0.000855 | 0.9803 |
| src/compiler/binder.ts:499:1 |  | 0.040812 | 0.039955 | -0.000857 | 0.9790 |
| src/compiler/utilities.ts:3402:1 | isSuperProperty | 0.050351 | 0.049490 | -0.000861 | 0.9829 |
| src/compiler/utilities.ts:9032:1 | CompilerOptionKeys | 0.040478 | 0.039611 | -0.000866 | 0.9786 |
| src/compiler/commandLineParser.ts:4075:1 | matchesExcludeWorker | 0.042955 | 0.042087 | -0.000868 | 0.9798 |
| src/compiler/types.ts:5512:1 | IntersectionFlags | 0.038145 | 0.037274 | -0.000871 | 0.9772 |
| src/compiler/utilities.ts:8976:1 | isFileForcedToBeModuleByFormat | 0.042900 | 0.042029 | -0.000872 | 0.9797 |
| src/compiler/scanner.ts:222:1 |  | 0.040978 | 0.040106 | -0.000872 | 0.9787 |
| src/compiler/utilities.ts:9978:1 |  | 0.040288 | 0.039415 | -0.000873 | 0.9783 |
| src/compiler/factory/nodeTests.ts:906:1 | isMissingDeclaration | 0.049691 | 0.048817 | -0.000874 | 0.9824 |
| src/compiler/sourcemap.ts:636:1 | base64FormatEncode | 0.040293 | 0.039411 | -0.000882 | 0.9781 |
| src/compiler/_namespaces/ts.ts:11:1 |  | 0.041730 | 0.040847 | -0.000883 | 0.9788 |
| src/compiler/core.ts:1971:1 | compareComparableValues | 0.038878 | 0.037994 | -0.000884 | 0.9773 |
| src/compiler/_namespaces/ts.ts:38:1 |  | 0.042416 | 0.041524 | -0.000891 | 0.9790 |
| src/compiler/types.ts:10044:1 | ModuleSpecifierCache | 0.040015 | 0.039122 | -0.000892 | 0.9777 |
| src/compiler/types.ts:2883:1 | TemplateMiddle | 0.042213 | 0.041318 | -0.000895 | 0.9788 |
| src/compiler/emitter.ts:502:1 | canEmitTsBuildInfo | 0.041397 | 0.040501 | -0.000896 | 0.9784 |
| src/compiler/utilities.ts:7292:1 | getSyntacticModifierFlags | 0.051249 | 0.050350 | -0.000899 | 0.9825 |
| src/compiler/moduleSpecifiers.ts:1454:1 | getRelativePathIfInSameVolume | 0.041056 | 0.040157 | -0.000899 | 0.9781 |
| src/compiler/factory/emitNode.ts:357:1 | getIdentifierTypeArguments | 0.051240 | 0.050337 | -0.000902 | 0.9824 |
| src/compiler/types.ts:2411:1 | UnaryExpression | 0.040869 | 0.039966 | -0.000903 | 0.9779 |
| src/compiler/watchUtilities.ts:89:1 | CachedDirectoryStructureHost | 0.040533 | 0.039630 | -0.000903 | 0.9777 |
| src/compiler/utilities.ts:7221:1 | hasAmbientModifier | 0.049879 | 0.048966 | -0.000913 | 0.9817 |
| src/compiler/types.ts:8309:1 | SourceMapRange | 0.041334 | 0.040405 | -0.000929 | 0.9775 |
| src/compiler/types.ts:9768:1 | VisitResult | 0.042772 | 0.041842 | -0.000930 | 0.9783 |
| src/compiler/factory/emitNode.ts:183:1 | getCommentRange | 0.077040 | 0.076109 | -0.000932 | 0.9879 |
| src/compiler/commandLineParser.ts:1923:1 | ParseCommandLineWorkerDiagnostics | 0.041188 | 0.040255 | -0.000933 | 0.9774 |
| src/compiler/path.ts:356:1 | getBaseFileName | 0.042452 | 0.041518 | -0.000934 | 0.9780 |
| src/compiler/moduleNameResolver.ts:1493:1 | ResolutionKindSpecificLoader | 0.054311 | 0.053375 | -0.000936 | 0.9828 |
| src/compiler/path.ts:110:1 | pathIsBareSpecifier | 0.039794 | 0.038855 | -0.000939 | 0.9764 |
| src/compiler/resolutionCache.ts:468:1 | getDirectoryToWatchFailedLookupLocationFromTypeRoot | 0.041949 | 0.041005 | -0.000944 | 0.9775 |
| src/compiler/core.ts:2074:1 | compareStringsCaseSensitive | 0.040101 | 0.039152 | -0.000949 | 0.9763 |
| src/compiler/moduleNameResolver.ts:402:1 | readPackageJsonTSConfigField | 0.044603 | 0.043646 | -0.000957 | 0.9786 |
| src/compiler/transformer.ts:104:1 | TransformationState | 0.040735 | 0.039775 | -0.000960 | 0.9764 |
| src/compiler/factory/nodeTests.ts:251:1 | isRegularExpressionLiteral | 0.046089 | 0.045126 | -0.000963 | 0.9791 |
| src/compiler/transformers/namedEvaluation.ts:179:1 | injectClassNamedEvaluationHelperBlockIfMissing | 0.040457 | 0.039493 | -0.000963 | 0.9762 |
| src/compiler/utilitiesPublic.ts:1640:1 | isClassMemberModifier | 0.044893 | 0.043921 | -0.000971 | 0.9784 |
| src/compiler/utilities.ts:8492:1 | Type | 0.041018 | 0.040046 | -0.000973 | 0.9763 |
| src/compiler/commandLineParser.ts:1727:1 |  | 0.040405 | 0.039431 | -0.000973 | 0.9759 |
| src/compiler/scanner.ts:4103:1 |  | 0.067134 | 0.066155 | -0.000979 | 0.9854 |
| src/compiler/visitorPublic.ts:394:1 | visitParameterList | 0.043026 | 0.042039 | -0.000987 | 0.9771 |
| src/compiler/transformers/es2019.ts:1:1 |  | 0.045590 | 0.044603 | -0.000988 | 0.9783 |
| src/compiler/watch.ts:713:1 | WatchTypeRegistry | 0.040728 | 0.039736 | -0.000992 | 0.9756 |
| src/compiler/transformers/classFields.ts:263:1 | PrivateIdentifierAccessorInfo | 0.040414 | 0.039420 | -0.000994 | 0.9754 |
| src/compiler/checker.ts:1412:1 | DeclarationMeaning | 0.041579 | 0.040581 | -0.000998 | 0.9760 |
| src/compiler/moduleNameResolver.ts:1401:1 | resolveModuleNameFromCache | 0.084091 | 0.083090 | -0.001001 | 0.9881 |
| src/compiler/moduleSpecifiers.ts:1014:1 | MatchingMode | 0.039875 | 0.038871 | -0.001005 | 0.9748 |
| src/compiler/_namespaces/ts.ts:49:1 |  | 0.041828 | 0.040815 | -0.001013 | 0.9758 |
| src/compiler/types.ts:4206:1 | FlowLabel | 0.041769 | 0.040754 | -0.001014 | 0.9757 |
| src/compiler/parser.ts:10591:1 | processCommentPragmas | 0.039966 | 0.038947 | -0.001020 | 0.9745 |
| src/compiler/types.ts:2775:1 | TemplateLiteralLikeNode | 0.042749 | 0.041725 | -0.001024 | 0.9761 |
| src/compiler/utilities.ts:7211:1 | hasOverrideModifier | 0.051227 | 0.050191 | -0.001036 | 0.9798 |
| src/compiler/types.ts:5886:1 | LazyNodeCheckFlags | 0.040684 | 0.039641 | -0.001042 | 0.9744 |
| src/compiler/types.ts:5843:1 | TypeReferenceSerializationKind | 0.039598 | 0.038552 | -0.001046 | 0.9736 |
| src/compiler/emitter.ts:481:1 | getTsBuildInfoEmitOutputFilePath | 0.051783 | 0.050736 | -0.001047 | 0.9798 |
| src/compiler/performanceCore.ts:90:1 |  | 0.042415 | 0.041361 | -0.001054 | 0.9751 |
| src/compiler/debug.ts:2:1 |  | 0.042292 | 0.041237 | -0.001055 | 0.9750 |
| src/compiler/utilities.ts:9322:1 |  | 0.103048 | 0.101992 | -0.001056 | 0.9898 |
| src/compiler/executeCommandLine.ts:902:1 | createReportErrorSummary | 0.100491 | 0.099433 | -0.001058 | 0.9895 |
| src/compiler/utilities.ts:6298:1 | createTextWriter | 0.043136 | 0.042078 | -0.001058 | 0.9755 |
| src/compiler/parser.ts:1340:1 | setExternalModuleIndicator | 0.043071 | 0.042006 | -0.001066 | 0.9753 |
| src/compiler/utilities.ts:8019:1 | getInitializedVariables | 0.059435 | 0.058366 | -0.001069 | 0.9820 |
| src/compiler/types.ts:10637:1 | SyntacticTypeNodeBuilderResolver | 0.040762 | 0.039692 | -0.001070 | 0.9738 |
| src/compiler/utilities.ts:9382:1 | getEmitStandardClassFields | 0.062894 | 0.061824 | -0.001070 | 0.9830 |
| src/compiler/utilitiesPublic.ts:1503:1 | isLiteralExpressionOfObject | 0.045259 | 0.044186 | -0.001073 | 0.9763 |
| src/compiler/utilities.ts:5485:1 | isAnonymousFunctionDefinition | 0.047518 | 0.046443 | -0.001075 | 0.9774 |
| src/compiler/types.ts:3231:1 | JsxNamespacedName | 0.042941 | 0.041862 | -0.001079 | 0.9749 |
| src/compiler/path.ts:855:1 | changeAnyExtension | 0.045438 | 0.044352 | -0.001086 | 0.9761 |
| src/compiler/program.ts:1033:1 | ResolutionLoader | 0.040185 | 0.039097 | -0.001088 | 0.9729 |
| src/compiler/types.ts:3872:1 | CommentRange | 0.041004 | 0.039915 | -0.001089 | 0.9734 |
| src/compiler/scanner.ts:966:1 | getShebang | 0.040344 | 0.039250 | -0.001094 | 0.9729 |
| src/compiler/types.ts:5584:1 | TypeFormatFlags | 0.038617 | 0.037514 | -0.001102 | 0.9715 |
| src/compiler/types.ts:2966:1 | EntityNameOrEntityNameExpression | 0.043738 | 0.042635 | -0.001103 | 0.9748 |
| src/compiler/types.ts:2965:1 | EntityNameExpression | 0.044149 | 0.043038 | -0.001111 | 0.9748 |
| src/compiler/types.ts:2534:1 | AdditiveOperator | 0.042737 | 0.041626 | -0.001112 | 0.9740 |
| src/compiler/types.ts:2815:1 | TokenFlags | 0.043213 | 0.042090 | -0.001123 | 0.9740 |
| src/compiler/parser.ts:9936:1 |  | 0.040855 | 0.039727 | -0.001129 | 0.9724 |
| src/compiler/utilities.ts:10035:1 | hasTSFileExtension | 0.041056 | 0.039924 | -0.001132 | 0.9724 |
| src/compiler/types.ts:2655:1 | AssignmentOperatorToken | 0.043752 | 0.042609 | -0.001144 | 0.9739 |
| src/compiler/utilities.ts:6181:1 | encodeUtf16EscapeSequence | 0.044303 | 0.043159 | -0.001144 | 0.9742 |
| src/compiler/types.ts:6201:1 | __String | 0.040015 | 0.038867 | -0.001147 | 0.9713 |
| src/compiler/utilities.ts:5896:1 | getOperatorPrecedence | 0.052002 | 0.050854 | -0.001148 | 0.9779 |
| src/compiler/utilities.ts:8754:1 | chainDiagnosticMessages | 0.078363 | 0.077208 | -0.001155 | 0.9853 |
| src/compiler/transformers/classFields.ts:343:1 | PrivateEnv | 0.040260 | 0.039104 | -0.001156 | 0.9713 |
| src/compiler/types.ts:3410:1 | WhileStatement | 0.045964 | 0.044803 | -0.001161 | 0.9747 |
| src/compiler/transformers/namedEvaluation.ts:256:1 | transformNamedEvaluationOfPropertyAssignment | 0.043822 | 0.042659 | -0.001163 | 0.9735 |
| src/compiler/types.ts:3215:1 | JsxTagNameExpression | 0.043743 | 0.042577 | -0.001166 | 0.9733 |
| src/compiler/emitter.ts:2:1 |  | 0.042092 | 0.040924 | -0.001168 | 0.9722 |
| src/compiler/factory/nodeTests.ts:329:1 | isExportModifier | 0.047379 | 0.046205 | -0.001174 | 0.9752 |
| src/compiler/utilities.ts:10957:1 | isThisTypeParameter | 0.040531 | 0.039356 | -0.001175 | 0.9710 |
| src/compiler/_namespaces/ts.ts:46:1 |  | 0.042350 | 0.041174 | -0.001175 | 0.9722 |
| src/compiler/_namespaces/ts.ts:16:1 |  | 0.042329 | 0.041146 | -0.001183 | 0.9721 |
| src/compiler/factory/nodeFactory.ts:7290:1 | propagateIdentifierNameFlags | 0.047790 | 0.046602 | -0.001188 | 0.9751 |
| src/compiler/transformers/namedEvaluation.ts:274:1 | transformNamedEvaluationOfShorthandAssignmentProperty | 0.042508 | 0.041318 | -0.001190 | 0.9720 |
| src/compiler/_namespaces/ts.ts:31:1 |  | 0.042740 | 0.041546 | -0.001193 | 0.9721 |
| src/compiler/types.ts:5530:1 | NodeBuilderFlags | 0.039080 | 0.037883 | -0.001198 | 0.9694 |
| src/compiler/program.ts:500:1 | CompilerHostLikeForCache | 0.041548 | 0.040343 | -0.001206 | 0.9710 |
| src/compiler/_namespaces/ts.ts:37:1 |  | 0.042827 | 0.041619 | -0.001208 | 0.9718 |
| src/compiler/transformers/declarations.ts:1903:1 | ProcessedDeclarationStatement | 0.039789 | 0.038572 | -0.001217 | 0.9694 |
| src/compiler/types.ts:7078:1 | IndexInfo | 0.041026 | 0.039808 | -0.001218 | 0.9703 |
| src/compiler/factory/nodeChildren.ts:11:1 |  | 0.042403 | 0.041182 | -0.001221 | 0.9712 |
| src/compiler/utilities.ts:3778:1 | getExternalModuleImportEqualsDeclarationExpression | 0.050203 | 0.048980 | -0.001223 | 0.9756 |
| src/compiler/types.ts:2583:1 | BitwiseOperator | 0.044630 | 0.043402 | -0.001228 | 0.9725 |
| src/compiler/utilities.ts:5997:1 | getBinaryOperatorPrecedence | 0.046533 | 0.045303 | -0.001230 | 0.9736 |
| src/compiler/core.ts:1974:1 | compareComparableValues | 0.038891 | 0.037660 | -0.001231 | 0.9683 |
| src/compiler/transformers/utilities.ts:589:1 | getProperties | 0.041215 | 0.039983 | -0.001233 | 0.9701 |
| src/compiler/utilities.ts:8236:1 | MutateMapWithNewSetOptions | 0.040052 | 0.038816 | -0.001236 | 0.9691 |
| src/compiler/factory/nodeTests.ts:296:1 | isExclamationToken | 0.047228 | 0.045989 | -0.001239 | 0.9738 |
| src/compiler/tracing.ts:369:1 |  | 0.043635 | 0.042379 | -0.001256 | 0.9712 |
| src/compiler/_namespaces/ts.ts:42:1 |  | 0.042901 | 0.041643 | -0.001258 | 0.9707 |
| src/compiler/utilities.ts:6901:1 | getTypeAnnotationNode | 0.047471 | 0.046211 | -0.001260 | 0.9735 |
| src/compiler/utilities.ts:6933:1 | getEffectiveSetAccessorTypeAnnotationNode | 0.053182 | 0.051919 | -0.001263 | 0.9763 |
| src/compiler/watchPublic.ts:353:1 | WatchOfConfigFile | 0.038986 | 0.037723 | -0.001264 | 0.9676 |
| src/compiler/path.ts:120:1 | fileExtensionIs | 0.041961 | 0.040696 | -0.001265 | 0.9699 |
| src/compiler/moduleNameResolver.ts:406:1 | readPackageJsonMainField | 0.043227 | 0.041960 | -0.001267 | 0.9707 |
| src/compiler/commandLineParser.ts:1879:1 | parseCustomTypeOption | 0.047945 | 0.046670 | -0.001275 | 0.9734 |
| src/compiler/types.ts:6160:1 | ReverseMappedSymbolLinks | 0.040163 | 0.038877 | -0.001287 | 0.9680 |
| src/compiler/factory/nodeTests.ts:288:1 | isMinusToken | 0.046572 | 0.045269 | -0.001302 | 0.9720 |
| src/compiler/sourcemap.ts:645:1 | base64FormatDecode | 0.040245 | 0.038939 | -0.001306 | 0.9675 |
| src/compiler/emitter.ts:636:1 | getCommonSourceDirectory | 0.039762 | 0.038450 | -0.001312 | 0.9670 |
| src/compiler/moduleNameResolver.ts:3064:1 | loadModuleFromImmediateNodeModulesDirectory | 0.039538 | 0.038223 | -0.001315 | 0.9667 |
| src/compiler/utilities.ts:9372:1 | getStrictOptionValue | 0.069373 | 0.068055 | -0.001318 | 0.9810 |
| src/compiler/utilities.ts:7321:1 | getJSDocModifierFlagsNoCache | 0.048807 | 0.047488 | -0.001319 | 0.9730 |
| src/compiler/factory/utilities.ts:1489:1 | isExportOrDefaultKeywordKind | 0.042786 | 0.041466 | -0.001319 | 0.9692 |
| src/compiler/utilities.ts:8961:1 | walkTreeForJSXTags | 0.299766 | 0.298446 | -0.001320 | 0.9956 |
| src/compiler/types.ts:6498:1 | UniqueESSymbolType | 0.042184 | 0.040842 | -0.001342 | 0.9682 |
| src/compiler/types.ts:2544:1 | ShiftOperator | 0.042134 | 0.040786 | -0.001348 | 0.9680 |
| src/compiler/utilities.ts:6229:1 |  | 0.043246 | 0.041897 | -0.001349 | 0.9688 |
| src/compiler/factory/utilities.ts:1222:1 | isEqualityOperator | 0.043317 | 0.041964 | -0.001353 | 0.9688 |
| src/compiler/semver.ts:27:1 |  | 0.040248 | 0.038890 | -0.001358 | 0.9663 |
| src/compiler/_namespaces/ts.ts:76:1 |  | 0.040353 | 0.038994 | -0.001359 | 0.9663 |
| src/compiler/commandLineParser.ts:3920:1 |  | 0.041133 | 0.039773 | -0.001360 | 0.9669 |
| src/compiler/watchPublic.ts:396:1 | ParsedConfig | 0.039249 | 0.037880 | -0.001369 | 0.9651 |
| src/compiler/types.ts:6637:1 | BaseType | 0.040665 | 0.039294 | -0.001371 | 0.9663 |
| src/compiler/utilities.ts:2325:1 | getEnclosingContainer | 0.065568 | 0.064197 | -0.001371 | 0.9791 |
| src/compiler/factory/utilities.ts:1203:1 | isShiftOperatorOrHigher | 0.108585 | 0.107213 | -0.001372 | 0.9874 |
| src/compiler/utilities.ts:5014:1 | walkUpParenthesizedExpressions | 0.056191 | 0.054819 | -0.001373 | 0.9756 |
| src/compiler/moduleNameResolver.ts:318:1 | ModuleResolutionState | 0.042035 | 0.040657 | -0.001378 | 0.9672 |
| src/compiler/utilities.ts:7413:1 | isLogicalOrCoalescingAssignmentOperator | 0.047922 | 0.046538 | -0.001384 | 0.9711 |
| src/compiler/transformers/es2015.ts:289:1 | ConvertedLoopState | 0.042869 | 0.041483 | -0.001386 | 0.9677 |
| src/compiler/utilities.ts:2315:1 | hasPossibleExternalModuleReference | 0.326765 | 0.325379 | -0.001387 | 0.9958 |
| src/compiler/types.ts:2858:1 | NumericLiteral | 0.043362 | 0.041973 | -0.001389 | 0.9680 |
| src/compiler/scanner.ts:794:1 | isShebangTrivia | 0.041047 | 0.039655 | -0.001392 | 0.9661 |
| src/compiler/types.ts:2442:1 | PostfixUnaryExpression | 0.041961 | 0.040568 | -0.001394 | 0.9668 |
| src/compiler/_namespaces/ts.ts:63:1 |  | 0.040208 | 0.038814 | -0.001394 | 0.9653 |
| src/compiler/utilitiesPublic.ts:2108:1 | isAssertionExpression | 0.049485 | 0.048087 | -0.001398 | 0.9718 |
| src/compiler/types.ts:3394:1 | IfStatement | 0.043637 | 0.042239 | -0.001398 | 0.9680 |
| src/compiler/utilitiesPublic.ts:1921:1 | isArrayBindingOrAssignmentPattern | 0.042292 | 0.040887 | -0.001405 | 0.9668 |
| src/compiler/types.ts:2648:1 | BinaryExpression | 0.042929 | 0.041516 | -0.001413 | 0.9671 |
| src/compiler/commandLineParser.ts:1847:1 | getOptionsNameMap | 0.042216 | 0.040799 | -0.001416 | 0.9665 |
| src/compiler/factory/nodeTests.ts:339:1 | isAsyncModifier | 0.046355 | 0.044937 | -0.001418 | 0.9694 |
| src/compiler/utilitiesPublic.ts:667:1 | getCombinedNodeFlagsAlwaysIncludeJSDoc | 0.051441 | 0.050013 | -0.001428 | 0.9722 |
| src/compiler/utilitiesPublic.ts:764:1 | getOriginalNode | 0.043850 | 0.042419 | -0.001432 | 0.9674 |
| src/compiler/core.ts:683:1 | indicesOf | 0.041322 | 0.039889 | -0.001433 | 0.9653 |
| src/compiler/types.ts:3331:1 | NotEmittedTypeElement | 0.042618 | 0.041180 | -0.001438 | 0.9662 |
| src/compiler/_namespaces/ts.ts:57:1 |  | 0.041751 | 0.040312 | -0.001440 | 0.9655 |
| src/compiler/scanner.ts:4068:1 | utf16EncodeAsString | 0.040171 | 0.038730 | -0.001441 | 0.9641 |
| src/compiler/types.ts:4122:1 | JSDocTypeLiteral | 0.054057 | 0.052615 | -0.001442 | 0.9733 |
| src/compiler/utilities.ts:4812:1 | getJSDocHost | 0.042406 | 0.040961 | -0.001445 | 0.9659 |
| src/compiler/types.ts:2456:1 | PrimaryExpression | 0.042539 | 0.041091 | -0.001447 | 0.9660 |
| src/compiler/types.ts:3453:1 | BreakOrContinueStatement | 0.042410 | 0.040959 | -0.001451 | 0.9658 |
| src/compiler/types.ts:6149:1 | MappedSymbolLinks | 0.040183 | 0.038732 | -0.001451 | 0.9639 |
| src/compiler/watchPublic.ts:335:1 | Watch | 0.040077 | 0.038621 | -0.001456 | 0.9637 |
| src/compiler/factory/nodeChildren.ts:1:1 |  | 0.041132 | 0.039674 | -0.001458 | 0.9646 |
| src/compiler/transformers/module/module.ts:1:1 |  | 0.041701 | 0.040239 | -0.001463 | 0.9649 |
| src/compiler/transformers/destructuring.ts:208:1 | bindingOrAssignmentPatternAssignsToName | 0.044806 | 0.043343 | -0.001463 | 0.9673 |
| src/compiler/types.ts:6611:1 | ObjectType | 0.042275 | 0.040811 | -0.001464 | 0.9654 |
| src/compiler/types.ts:10672:1 | SyntacticNodeBuilder | 0.039839 | 0.038371 | -0.001468 | 0.9632 |
| src/compiler/utilitiesPublic.ts:1136:1 | getJSDocPrivateTag | 0.059408 | 0.057939 | -0.001469 | 0.9753 |
| src/compiler/commandLineParser.ts:2127:1 | parseCommandLine | 0.042409 | 0.040938 | -0.001471 | 0.9653 |
| src/compiler/utilities.ts:12159:1 |  | 0.043427 | 0.041953 | -0.001474 | 0.9661 |
| src/compiler/utilitiesPublic.ts:1999:1 | isLeftHandSideExpressionKind | 0.048680 | 0.047205 | -0.001475 | 0.9697 |
| src/compiler/builder.ts:898:1 | isChangedSignature | 0.042638 | 0.041163 | -0.001475 | 0.9654 |
| src/compiler/moduleNameResolver.ts:3001:1 | loadModuleFromNearestNodeModulesDirectoryTypesScope | 0.039938 | 0.038461 | -0.001477 | 0.9630 |
| src/compiler/types.ts:2438:1 | PostfixUnaryOperator | 0.041594 | 0.040115 | -0.001478 | 0.9645 |
| src/compiler/utilities.ts:9017:1 | importSyntaxAffectsModuleResolution | 0.043008 | 0.041516 | -0.001493 | 0.9653 |
| src/compiler/path.ts:943:1 | comparePaths | 0.041936 | 0.040437 | -0.001499 | 0.9643 |
| src/compiler/commandLineParser.ts:2332:1 |  | 0.042275 | 0.040773 | -0.001501 | 0.9645 |
| src/compiler/utilities.ts:7312:1 | selectSyntacticModifierFlags | 0.043679 | 0.042175 | -0.001504 | 0.9656 |
| src/compiler/parser.ts:10616:1 | PragmaDiagnosticReporter | 0.039677 | 0.038169 | -0.001508 | 0.9620 |
| src/compiler/checker.ts:1390:1 | ExpandingFlags | 0.042055 | 0.040545 | -0.001510 | 0.9641 |
| src/compiler/types.ts:2400:1 | OmittedExpression | 0.040758 | 0.039237 | -0.001521 | 0.9627 |
| src/compiler/parser.ts:421:1 |  | 0.040814 | 0.039284 | -0.001529 | 0.9625 |
| src/compiler/utilities.ts:2320:1 | isAnyImportOrReExport | 0.292673 | 0.291141 | -0.001532 | 0.9948 |
| src/compiler/types.ts:6623:1 | InterfaceType | 0.042185 | 0.040651 | -0.001534 | 0.9636 |
| src/compiler/parser.ts:9942:1 | markAsIntersectingIncrementalChange | 0.044711 | 0.043176 | -0.001535 | 0.9657 |
| src/compiler/_namespaces/ts.ts:60:1 |  | 0.039902 | 0.038364 | -0.001538 | 0.9615 |
| src/compiler/moduleNameResolver.ts:2189:1 | tryFile | 0.046469 | 0.044925 | -0.001545 | 0.9668 |
| src/compiler/path.ts:947:1 | comparePaths | 0.042246 | 0.040701 | -0.001545 | 0.9634 |
| src/compiler/core.ts:1750:1 | isArray | 0.042566 | 0.041020 | -0.001546 | 0.9637 |
| src/compiler/types.ts:4695:1 | EmitOnly | 0.041206 | 0.039653 | -0.001552 | 0.9623 |
| src/compiler/utilities.ts:2166:1 | isEffectiveStrictModeSourceFile | 0.045146 | 0.043593 | -0.001554 | 0.9656 |
| src/compiler/checker.ts:1170:1 |  | 0.043008 | 0.041450 | -0.001558 | 0.9638 |
| src/compiler/commandLineParser.ts:3305:1 | startsWithConfigDirTemplate | 0.039627 | 0.038067 | -0.001560 | 0.9606 |
| src/compiler/moduleNameResolver.ts:2089:1 | loadModuleFromFileNoImplicitExtensions | 0.044551 | 0.042983 | -0.001568 | 0.9648 |
| src/compiler/executeCommandLine.ts:1257:1 | reportAllStatistics | 0.078692 | 0.077119 | -0.001573 | 0.9800 |
| src/compiler/factory/utilities.ts:1186:1 | isAdditiveOperator | 0.043634 | 0.042060 | -0.001574 | 0.9639 |
| src/compiler/core.ts:2430:1 | startsWith | 0.041382 | 0.039807 | -0.001575 | 0.9619 |
| src/compiler/types.ts:6224:1 | NodeCheckFlags | 0.040184 | 0.038606 | -0.001578 | 0.9607 |
| src/compiler/moduleNameResolver.ts:458:1 |  | 0.042513 | 0.040926 | -0.001587 | 0.9627 |
| src/compiler/scanner.ts:4092:1 |  | 0.039967 | 0.038380 | -0.001587 | 0.9603 |
| src/compiler/utilities.ts:1182:1 | isRecognizedTripleSlashComment | 0.043172 | 0.041578 | -0.001594 | 0.9631 |
| src/compiler/utilities.ts:8231:1 | MutateMapOptionsCreate | 0.041034 | 0.039431 | -0.001603 | 0.9609 |
| src/compiler/commandLineParser.ts:3744:1 | convertTypeAcquisitionFromJson | 0.041567 | 0.039963 | -0.001604 | 0.9614 |
| src/compiler/utilities.ts:980:1 | getSourceFileOfNode | 0.046126 | 0.044516 | -0.001610 | 0.9651 |
| src/compiler/transformers/classFields.ts:275:1 | PrivateIdentifierMethodInfo | 0.041509 | 0.039892 | -0.001617 | 0.9611 |
| src/compiler/emitter.ts:535:1 | getSourceMapFilePath | 0.041543 | 0.039921 | -0.001622 | 0.9610 |
| src/compiler/builder.ts:1186:1 | NonIncrementalBuildInfo | 0.043041 | 0.041410 | -0.001631 | 0.9621 |
| src/compiler/utilitiesPublic.ts:380:1 | textSpanOverlapsWith | 0.043240 | 0.041601 | -0.001639 | 0.9621 |
| src/compiler/utilities.ts:3839:1 | isRequireCall | 0.172515 | 0.170876 | -0.001640 | 0.9905 |
| src/compiler/checker.ts:1171:1 |  | 0.041851 | 0.040205 | -0.001646 | 0.9607 |
| src/compiler/utilitiesPublic.ts:1096:1 | getJSDocTypeParameterTagsNoCache | 0.050435 | 0.048787 | -0.001648 | 0.9673 |
| src/compiler/types.ts:6774:1 | StructuredType | 0.040515 | 0.038866 | -0.001649 | 0.9593 |
| src/compiler/types.ts:9711:1 | Visitor | 0.041637 | 0.039982 | -0.001655 | 0.9603 |
| src/compiler/commandLineParser.ts:4045:1 | invalidDotDotAfterRecursiveWildcard | 0.042154 | 0.040498 | -0.001656 | 0.9607 |
| src/compiler/types.ts:2550:1 | ShiftOperatorOrHigher | 0.042010 | 0.040352 | -0.001657 | 0.9605 |
| src/compiler/factory/utilities.ts:656:1 | skipOuterExpressions | 0.044832 | 0.043148 | -0.001684 | 0.9624 |
| src/compiler/utilitiesPublic.ts:787:1 | findAncestor | 0.043348 | 0.041657 | -0.001691 | 0.9610 |
| src/compiler/_namespaces/ts.ts:56:1 |  | 0.042151 | 0.040450 | -0.001701 | 0.9596 |
| src/compiler/moduleNameResolver.ts:900:1 | ModuleResolutionCache | 0.043769 | 0.042068 | -0.001702 | 0.9611 |
| src/compiler/types.ts:7270:1 | Diagnostic | 0.040864 | 0.039160 | -0.001704 | 0.9583 |
| src/compiler/watchPublic.ts:359:1 | WatchOfFilesAndCompilerOptions | 0.039943 | 0.038238 | -0.001705 | 0.9573 |
| src/compiler/core.ts:2106:1 |  | 0.040990 | 0.039280 | -0.001710 | 0.9583 |
| src/compiler/core.ts:2114:1 | setUILocale | 0.039673 | 0.037962 | -0.001711 | 0.9569 |
| src/compiler/utilities.ts:11213:1 | isSourceElement | 0.048297 | 0.046583 | -0.001714 | 0.9645 |
| src/compiler/commandLineParser.ts:1831:1 | createOptionNameMap | 0.043377 | 0.041662 | -0.001715 | 0.9605 |
| src/compiler/path.ts:60:1 | isRootedDiskPath | 0.040238 | 0.038519 | -0.001719 | 0.9573 |
| src/compiler/watch.ts:987:1 | performIncrementalCompilation | 0.042560 | 0.040841 | -0.001719 | 0.9596 |
| src/compiler/types.ts:2731:1 | AssignmentPattern | 0.041492 | 0.039771 | -0.001721 | 0.9585 |
| src/compiler/utilities.ts:8567:1 |  | 0.042252 | 0.040527 | -0.001724 | 0.9592 |
| src/compiler/factory/nodeTests.ts:300:1 | isQuestionToken | 0.047440 | 0.045711 | -0.001729 | 0.9636 |
| src/compiler/types.ts:2569:1 | EqualityOperator | 0.042936 | 0.041190 | -0.001746 | 0.9593 |
| src/compiler/utilities.ts:5507:1 | NamedEvaluationSource | 0.039896 | 0.038140 | -0.001756 | 0.9560 |
| src/compiler/moduleNameResolver.ts:427:1 | readPackageJsonTypesVersionPaths | 0.044049 | 0.042293 | -0.001756 | 0.9601 |
| src/compiler/types.ts:2496:1 | VoidExpression | 0.042408 | 0.040651 | -0.001757 | 0.9586 |
| src/compiler/utilities.ts:1286:1 | getSourceTextOfNodeFromSourceFile | 0.049556 | 0.047786 | -0.001771 | 0.9643 |
| src/compiler/program.ts:326:1 |  | 0.041984 | 0.040199 | -0.001785 | 0.9575 |
| src/compiler/utilities.ts:8179:1 | MutateMapSkippingNewValuesDelete | 0.041247 | 0.039460 | -0.001787 | 0.9567 |
| src/compiler/utilities.ts:8295:1 | getLastChild | 0.051178 | 0.049390 | -0.001788 | 0.9651 |
| src/compiler/types.ts:3321:1 | Statement | 0.043993 | 0.042203 | -0.001790 | 0.9593 |
| src/compiler/parser.ts:10702:1 |  | 0.043932 | 0.042140 | -0.001792 | 0.9592 |
| src/compiler/commandLineParser.ts:2158:1 | getBuildOptionsNameMap | 0.041024 | 0.039230 | -0.001794 | 0.9563 |
| src/compiler/utilities.ts:10315:1 |  | 0.070117 | 0.068321 | -0.001797 | 0.9744 |
| src/compiler/utilities.ts:9981:1 |  | 0.041844 | 0.040047 | -0.001797 | 0.9570 |
| src/compiler/utilities.ts:6695:1 | getSourceFilePathInNewDirWorker | 0.043079 | 0.041272 | -0.001807 | 0.9581 |
| src/compiler/utilitiesPublic.ts:1441:1 | isBreakOrContinueStatement | 0.048325 | 0.046516 | -0.001809 | 0.9626 |
| src/compiler/utilities.ts:1088:1 | nodeIsPresent | 0.049700 | 0.047889 | -0.001811 | 0.9636 |
| src/compiler/core.ts:2447:1 | isPatternMatch | 0.119297 | 0.117474 | -0.001823 | 0.9847 |
| src/compiler/utilities.ts:6759:1 | getFirstConstructorWithBody | 0.042418 | 0.040587 | -0.001831 | 0.9568 |
| src/compiler/utilities.ts:8555:1 |  | 0.070458 | 0.068618 | -0.001840 | 0.9739 |
| src/compiler/program.ts:1043:1 |  | 0.079316 | 0.077474 | -0.001842 | 0.9768 |
| src/compiler/utilities.ts:1079:1 | nodeIsMissing | 0.048494 | 0.046651 | -0.001844 | 0.9620 |
| src/compiler/commandLineParser.ts:1868:1 | createCompilerDiagnosticForInvalidCustomType | 0.042392 | 0.040540 | -0.001852 | 0.9563 |
| src/compiler/utilities.ts:4509:1 | getSingleInitializerOfVariableStatementOrPropertyDeclaration | 0.170070 | 0.168215 | -0.001855 | 0.9891 |
| src/compiler/utilitiesPublic.ts:882:1 | nameForNamelessJSDocTypedef | 0.044054 | 0.042193 | -0.001861 | 0.9578 |
| src/compiler/types.ts:2751:1 | FunctionExpression | 0.043776 | 0.041915 | -0.001861 | 0.9575 |
| src/compiler/core.ts:2402:1 | matchedText | 0.042773 | 0.040909 | -0.001864 | 0.9564 |
| src/compiler/transformers/es2017.ts:106:1 | ES2017SubstitutionFlags | 0.042581 | 0.040714 | -0.001867 | 0.9562 |
| src/compiler/utilities.ts:8770:1 | concatenateDiagnosticMessageChains | 0.043008 | 0.041140 | -0.001867 | 0.9566 |
| src/compiler/types.ts:7405:1 | PollingWatchKind | 0.042615 | 0.040745 | -0.001871 | 0.9561 |
| src/compiler/types.ts:6217:1 | PatternAmbientModule | 0.040653 | 0.038779 | -0.001874 | 0.9539 |
| src/compiler/types.ts:2051:1 | BindingPattern | 0.048828 | 0.046954 | -0.001874 | 0.9616 |
| src/compiler/_namespaces/ts.ts:3:1 |  | 0.041646 | 0.039770 | -0.001876 | 0.9550 |
| src/compiler/types.ts:1943:1 | AutoAccessorPropertyDeclaration | 0.077493 | 0.075612 | -0.001881 | 0.9757 |
| src/compiler/utilities.ts:6918:1 | getJSDocTypeParameterDeclarations | 0.072187 | 0.070303 | -0.001884 | 0.9739 |
| src/compiler/utilitiesPublic.ts:1905:1 | isObjectBindingOrAssignmentElement | 0.047030 | 0.045146 | -0.001884 | 0.9599 |
| src/compiler/utilities.ts:8779:1 | getDiagnosticFilePath | 0.040877 | 0.038989 | -0.001888 | 0.9538 |
| src/compiler/tsbuildPublic.ts:216:1 | SolutionBuilderHostBase | 0.041905 | 0.040015 | -0.001890 | 0.9549 |
| src/compiler/utilities.ts:3541:1 | nodeOrChildIsDecorated | 0.072566 | 0.070673 | -0.001893 | 0.9739 |
| src/compiler/utilities.ts:2782:1 |  | 0.043610 | 0.041714 | -0.001896 | 0.9565 |
| src/compiler/executeCommandLine.ts:1276:1 | statisticValue | 0.040730 | 0.038832 | -0.001898 | 0.9534 |
| src/compiler/transformers/module/esnextAnd2015.ts:1:1 |  | 0.041376 | 0.039474 | -0.001903 | 0.9540 |
| src/compiler/moduleNameResolver.ts:2064:1 | loadModuleFromFileNoPackageId | 0.044536 | 0.042625 | -0.001912 | 0.9571 |
| src/compiler/moduleNameResolver.ts:2131:1 | tryAddingExtensions | 0.045208 | 0.043296 | -0.001912 | 0.9577 |
| src/compiler/utilities.ts:6793:1 | parameterIsThisKeyword | 0.046089 | 0.044170 | -0.001919 | 0.9584 |
| src/compiler/core.ts:1875:1 | toFileNameLowerCase | 0.040881 | 0.038959 | -0.001922 | 0.9530 |
| src/compiler/semver.ts:35:1 |  | 0.041286 | 0.039364 | -0.001922 | 0.9534 |
| src/compiler/binder.ts:315:1 | ModuleInstanceState | 0.042769 | 0.040846 | -0.001923 | 0.9550 |
| src/compiler/core.ts:232:1 | indexOfAnyCharCode | 0.040968 | 0.039038 | -0.001930 | 0.9529 |
| src/compiler/types.ts:6063:1 | SymbolLinks | 0.041593 | 0.039661 | -0.001932 | 0.9535 |
| src/compiler/types.ts:5498:1 | MemberOverrideStatus | 0.039437 | 0.037502 | -0.001935 | 0.9509 |
| src/compiler/corePublic.ts:6:1 |  | 0.040935 | 0.038995 | -0.001940 | 0.9526 |
| src/compiler/factory/nodeFactory.ts:7407:1 |  | 0.043045 | 0.041097 | -0.001948 | 0.9548 |
| src/compiler/utilities.ts:5637:1 | getExpressionAssociativity | 0.195129 | 0.193171 | -0.001958 | 0.9900 |
| src/compiler/checker.ts:1157:1 | ReferenceHint | 0.046832 | 0.044873 | -0.001959 | 0.9582 |
| src/compiler/factory/nodeTests.ts:334:1 | isDefaultModifier | 0.047700 | 0.045735 | -0.001965 | 0.9588 |
| src/compiler/moduleNameResolver.ts:3418:1 | traceIfEnabled | 0.041662 | 0.039696 | -0.001966 | 0.9528 |
| src/compiler/transformers/esnext.ts:778:1 | getUsingKindOfVariableDeclarationList | 0.046058 | 0.044091 | -0.001966 | 0.9573 |
| src/compiler/core.ts:1973:1 | compareComparableValues | 0.039963 | 0.037991 | -0.001972 | 0.9506 |
| src/compiler/resolutionCache.ts:498:1 | getRootDirectoryOfResolutionCache | 0.042801 | 0.040826 | -0.001975 | 0.9539 |
| src/compiler/utilities.ts:1002:1 | isStatementWithLocals | 0.047947 | 0.045967 | -0.001980 | 0.9587 |
| src/compiler/core.ts:2151:1 | compareBooleans | 0.041736 | 0.039754 | -0.001983 | 0.9525 |
| src/compiler/commandLineParser.ts:2289:1 | readJsonConfigFile | 0.076720 | 0.074723 | -0.001997 | 0.9740 |
| src/compiler/scanner.ts:135:1 |  | 0.042079 | 0.040074 | -0.002005 | 0.9524 |
| src/compiler/factory/nodeTests.ts:874:1 | isNamespaceExport | 0.063182 | 0.061174 | -0.002008 | 0.9682 |
| src/compiler/watchUtilities.ts:103:1 | MutableFileSystemEntries | 0.040307 | 0.038295 | -0.002012 | 0.9501 |
| src/compiler/_namespaces/ts.ts:43:1 |  | 0.042260 | 0.040240 | -0.002020 | 0.9522 |
| src/compiler/factory/nodeTests.ts:308:1 | isQuestionDotToken | 0.047130 | 0.045109 | -0.002021 | 0.9571 |
| src/compiler/types.ts:3384:1 | ExpressionStatement | 0.045164 | 0.043141 | -0.002022 | 0.9552 |
| src/compiler/checker.ts:1:1 |  | 0.044738 | 0.042715 | -0.002022 | 0.9548 |
| src/compiler/scanner.ts:4075:1 |  | 0.066274 | 0.064251 | -0.002024 | 0.9695 |
| src/compiler/types.ts:8356:1 | Placeholder | 0.041904 | 0.039872 | -0.002032 | 0.9515 |
| src/compiler/types.ts:7259:1 | DiagnosticMessageChain | 0.040933 | 0.038900 | -0.002033 | 0.9503 |
| src/compiler/emitter.ts:428:1 |  | 0.042738 | 0.040704 | -0.002034 | 0.9524 |
| src/compiler/utilities.ts:5631:1 | Associativity | 0.040822 | 0.038788 | -0.002034 | 0.9502 |
| src/compiler/_namespaces/ts.ts:51:1 |  | 0.044431 | 0.042395 | -0.002036 | 0.9542 |
| src/compiler/utilities.ts:7420:1 | isLogicalOrCoalescingAssignmentExpression | 0.350111 | 0.348074 | -0.002037 | 0.9942 |
| src/compiler/utilities.ts:6230:1 |  | 0.043481 | 0.041442 | -0.002039 | 0.9531 |
| src/compiler/commandLineParser.ts:2774:1 | serializeCompilerOptions | 0.042077 | 0.040036 | -0.002041 | 0.9515 |
| src/compiler/core.ts:17:1 |  | 0.040126 | 0.038083 | -0.002043 | 0.9491 |
| src/compiler/program.ts:1195:1 | getReferencedFileLocation | 0.045163 | 0.043110 | -0.002052 | 0.9546 |
| src/compiler/factory/utilities.ts:660:1 | skipOuterExpressions | 0.045078 | 0.043019 | -0.002059 | 0.9543 |
| src/compiler/utilities.ts:2970:1 | getRestParameterElementType | 0.165863 | 0.163789 | -0.002074 | 0.9875 |
| src/compiler/utilitiesPublic.ts:662:1 | getCombinedModifierFlags | 0.051653 | 0.049575 | -0.002078 | 0.9598 |
| src/compiler/core.ts:1966:1 | equateStringsCaseSensitive | 0.040928 | 0.038844 | -0.002084 | 0.9491 |
| src/compiler/performanceCore.ts:93:1 | tryGetNativePerformanceHooks | 0.043650 | 0.041563 | -0.002087 | 0.9522 |
| src/compiler/moduleNameResolver.ts:2199:1 | tryFileLookup | 0.093870 | 0.091764 | -0.002106 | 0.9776 |
| src/compiler/utilities.ts:5522:1 | isNamedEvaluationSource | 0.173932 | 0.171824 | -0.002108 | 0.9879 |
| src/compiler/program.ts:1015:1 |  | 0.041875 | 0.039766 | -0.002108 | 0.9496 |
| src/compiler/utilitiesPublic.ts:1471:1 | isTokenKind | 0.045625 | 0.043513 | -0.002112 | 0.9537 |
| src/compiler/utilities.ts:6798:1 | isThisIdentifier | 0.126142 | 0.124025 | -0.002117 | 0.9832 |
| src/compiler/utilities.ts:11173:1 | getPropertyNameFromType | 0.049205 | 0.047083 | -0.002122 | 0.9569 |
| src/compiler/utilities.ts:6138:1 |  | 0.041518 | 0.039394 | -0.002124 | 0.9489 |
| src/compiler/program.ts:1183:1 | SyntheticReferenceFileLocation | 0.043153 | 0.041028 | -0.002125 | 0.9508 |
| src/compiler/utilitiesPublic.ts:1635:1 | isParameterPropertyModifier | 0.046238 | 0.044112 | -0.002126 | 0.9540 |
| src/compiler/transformers/taggedTemplate.ts:35:1 | processTaggedTemplateExpression | 0.050053 | 0.047925 | -0.002129 | 0.9575 |
| src/compiler/types.ts:6895:1 | AccessFlags | 0.040631 | 0.038502 | -0.002129 | 0.9476 |
| src/compiler/types.ts:2906:1 | TemplateExpression | 0.046829 | 0.044700 | -0.002129 | 0.9545 |
| src/compiler/commandLineParser.ts:1915:1 | BaseParsedCommandLine | 0.041981 | 0.039833 | -0.002147 | 0.9489 |
| src/compiler/utilities.ts:5348:1 | isAsyncFunction | 0.067286 | 0.065139 | -0.002148 | 0.9681 |
| src/compiler/utilities.ts:3014:1 | isVariableDeclarationInVariableStatement | 0.040074 | 0.037917 | -0.002157 | 0.9462 |
| src/compiler/parser.ts:10619:1 | processPragmasIntoFields | 0.092839 | 0.090665 | -0.002175 | 0.9766 |
| src/compiler/factory/utilities.ts:1413:1 | BinaryExpressionStateMachine | 0.040935 | 0.038760 | -0.002175 | 0.9469 |
| src/compiler/moduleNameResolver.ts:211:1 | PathAndPackageId | 0.043360 | 0.041179 | -0.002181 | 0.9497 |
| src/compiler/types.ts:7604:1 | ModuleKind | 0.045586 | 0.043396 | -0.002190 | 0.9520 |
| src/compiler/factory/utilities.ts:1171:1 | isExponentiationOperator | 0.045129 | 0.042931 | -0.002197 | 0.9513 |
| src/compiler/factory/nodeFactory.ts:7188:1 | getDefaultTagNameForKind | 0.046448 | 0.044242 | -0.002207 | 0.9525 |
| src/compiler/commandLineParser.ts:4060:1 | matchesExclude | 0.044854 | 0.042640 | -0.002214 | 0.9506 |
| src/compiler/utilities.ts:3477:1 | nodeCanBeDecorated | 0.042441 | 0.040226 | -0.002215 | 0.9478 |
| src/compiler/parser.ts:424:1 |  | 0.041380 | 0.039155 | -0.002225 | 0.9462 |
| src/compiler/utilities.ts:3338:1 | getSuperContainer | 0.044342 | 0.042118 | -0.002225 | 0.9498 |
| src/compiler/core.ts:2079:1 | getStringComparer | 0.040279 | 0.038042 | -0.002237 | 0.9445 |
| src/compiler/types.ts:10599:1 | PseudoBigInt | 0.043545 | 0.041301 | -0.002244 | 0.9485 |
| src/compiler/types.ts:6110:1 | CheckFlags | 0.041018 | 0.038766 | -0.002252 | 0.9451 |
| src/compiler/core.ts:2032:1 | compareStringsCaseInsensitive | 0.040777 | 0.038524 | -0.002252 | 0.9448 |
| src/compiler/_namespaces/ts.ts:17:1 |  | 0.044145 | 0.041891 | -0.002253 | 0.9490 |
| src/compiler/_namespaces/ts.ts:64:1 |  | 0.040703 | 0.038449 | -0.002254 | 0.9446 |
| src/compiler/program.ts:1021:1 | ResolutionWithResolvedFileName | 0.041899 | 0.039641 | -0.002258 | 0.9461 |
| src/compiler/types.ts:3870:1 | CommentKind | 0.041274 | 0.039015 | -0.002260 | 0.9452 |
| src/compiler/types.ts:9997:1 | ModuleSpecifierResolutionHost | 0.041210 | 0.038948 | -0.002262 | 0.9451 |
| src/compiler/types.ts:2748:1 | FunctionBody | 0.044624 | 0.042355 | -0.002269 | 0.9491 |
| src/compiler/utilities.ts:2785:1 |  | 0.042440 | 0.040162 | -0.002278 | 0.9463 |
| src/compiler/watchPublic.ts:160:1 | WatchHost | 0.042114 | 0.039833 | -0.002281 | 0.9458 |
| src/compiler/moduleNameResolver.ts:1738:1 | nodeNextModuleNameResolverWorker | 0.046121 | 0.043829 | -0.002293 | 0.9503 |
| src/compiler/utilities.ts:3055:1 | unwrapInnermostStatementOfLabel | 0.131202 | 0.128901 | -0.002301 | 0.9825 |
| src/compiler/utilitiesPublic.ts:1494:1 | isLiteralKind | 0.049997 | 0.047689 | -0.002308 | 0.9538 |
| src/compiler/watch.ts:686:1 | WatchType | 0.042291 | 0.039974 | -0.002318 | 0.9452 |
| src/compiler/types.ts:6875:1 | TypeParameter | 0.040727 | 0.038406 | -0.002321 | 0.9430 |
| src/compiler/performanceCore.ts:104:1 |  | 0.042359 | 0.040036 | -0.002323 | 0.9451 |
| src/compiler/parser.ts:1437:1 | Parser | 0.056737 | 0.054402 | -0.002335 | 0.9588 |
| src/compiler/utilitiesPublic.ts:857:1 | idText | 0.041750 | 0.039396 | -0.002354 | 0.9436 |
| src/compiler/builderPublic.ts:201:1 | createEmitAndSemanticDiagnosticsBuilderProgram | 0.047524 | 0.045163 | -0.002361 | 0.9503 |
| src/compiler/factory/nodeFactory.ts:466:1 | NodeFactoryFlags | 0.044172 | 0.041794 | -0.002379 | 0.9462 |
| src/compiler/factory/utilities.ts:612:1 | isCommaSequence | 0.052707 | 0.050312 | -0.002394 | 0.9546 |
| src/compiler/_namespaces/ts.ts:5:1 |  | 0.040642 | 0.038244 | -0.002398 | 0.9410 |
| src/compiler/utilities.ts:3794:1 | isFullSourceFile | 0.079755 | 0.077355 | -0.002400 | 0.9699 |
| src/compiler/factory/utilities.ts:1196:1 | isShiftOperator | 0.043830 | 0.041423 | -0.002406 | 0.9451 |
| src/compiler/factory/nodeTests.ts:239:1 | isBigIntLiteral | 0.047182 | 0.044772 | -0.002410 | 0.9489 |
| src/compiler/utilities.ts:9397:1 | compilerOptionsAffectDeclarationPath | 0.061015 | 0.058604 | -0.002411 | 0.9605 |
| src/compiler/transformers/declarations.ts:1877:1 | maskModifierFlags | 0.048268 | 0.045853 | -0.002415 | 0.9500 |
| src/compiler/types.ts:6859:1 | SyntheticDefaultModuleType | 0.041800 | 0.039373 | -0.002427 | 0.9419 |
| src/compiler/_namespaces/ts.ts:70:1 |  | 0.042368 | 0.039930 | -0.002438 | 0.9425 |
| src/compiler/moduleNameResolver.ts:184:1 | Extensions | 0.043534 | 0.041093 | -0.002441 | 0.9439 |
| src/compiler/moduleNameResolver.ts:3261:1 | classicNameResolver | 0.042633 | 0.040191 | -0.002442 | 0.9427 |
| src/compiler/binder.ts:312:1 |  | 0.042472 | 0.040025 | -0.002448 | 0.9424 |
| src/compiler/utilities.ts:9698:1 | getSubPatternFromSpec | 0.040712 | 0.038262 | -0.002449 | 0.9398 |
| src/compiler/types.ts:6523:1 | ObjectFlags | 0.043049 | 0.040599 | -0.002449 | 0.9431 |
| src/compiler/types.ts:2564:1 | RelationalOperatorOrHigher | 0.043776 | 0.041320 | -0.002456 | 0.9439 |
| src/compiler/scanner.ts:1015:1 | ClassSetExpressionType | 0.040638 | 0.038180 | -0.002458 | 0.9395 |
| src/compiler/types.ts:2235:1 | ConstructorTypeNode | 0.043181 | 0.040715 | -0.002466 | 0.9429 |
| src/compiler/utilities.ts:6803:1 | isInTypeQuery | 0.060796 | 0.058323 | -0.002474 | 0.9593 |
| src/compiler/sourcemap.ts:820:1 |  | 0.082532 | 0.080055 | -0.002476 | 0.9700 |
| src/compiler/factory/utilities.ts:1234:1 | isBitwiseOperator | 0.043496 | 0.041017 | -0.002479 | 0.9430 |
| src/compiler/moduleNameResolver.ts:2059:1 | moveToNextDirectorySeparatorIfAvailable | 0.044695 | 0.042216 | -0.002479 | 0.9445 |
| src/compiler/types.ts:6139:1 | TransientSymbolLinks | 0.041402 | 0.038916 | -0.002485 | 0.9400 |
| src/compiler/utilities.ts:9948:1 | getScriptKindFromFileName | 0.041957 | 0.039471 | -0.002486 | 0.9407 |
| src/compiler/tsbuildPublic.ts:211:1 | ReportFileInError | 0.043049 | 0.040557 | -0.002492 | 0.9421 |
| src/compiler/debug.ts:108:1 | LoggingHost | 0.042822 | 0.040326 | -0.002496 | 0.9417 |
| src/compiler/types.ts:4573:1 | WriteFileCallbackData | 0.042370 | 0.039874 | -0.002496 | 0.9411 |
| src/compiler/types.ts:3173:1 | MetaProperty | 0.043953 | 0.041452 | -0.002501 | 0.9431 |
| src/compiler/scanner.ts:342:1 |  | 0.047345 | 0.044841 | -0.002504 | 0.9471 |
| src/compiler/program.ts:932:1 | getModeForUsageLocation | 0.047991 | 0.045487 | -0.002504 | 0.9478 |
| src/compiler/checker.ts:1221:1 | WideningKind | 0.042712 | 0.040201 | -0.002511 | 0.9412 |
| src/compiler/utilities.ts:7495:1 | isExpressionWithTypeArgumentsInClassExtendsClause | 0.051630 | 0.049113 | -0.002518 | 0.9512 |
| src/compiler/factory/emitHelpers.ts:1033:1 |  | 0.074352 | 0.071833 | -0.002519 | 0.9661 |
| src/compiler/factory/utilities.ts:1673:1 | findComputedPropertyNameCacheAssignment | 0.042165 | 0.039639 | -0.002526 | 0.9401 |
| src/compiler/utilities.ts:3383:1 | getImmediatelyInvokedFunctionExpression | 0.194378 | 0.191840 | -0.002538 | 0.9869 |
| src/compiler/types.ts:1625:1 | AssertKeyword | 0.062418 | 0.059878 | -0.002540 | 0.9593 |
| src/compiler/types.ts:6921:1 | TypeVariable | 0.043388 | 0.040842 | -0.002546 | 0.9413 |
| src/compiler/types.ts:3238:1 | JsxOpeningElement | 0.043156 | 0.040609 | -0.002546 | 0.9410 |
| src/compiler/types.ts:2688:1 | ObjectAssignmentElement | 0.042493 | 0.039938 | -0.002555 | 0.9399 |
| src/compiler/moduleSpecifiers.ts:357:1 | getModuleSpecifiers | 0.043681 | 0.041122 | -0.002560 | 0.9414 |
| src/compiler/tracing.ts:24:1 |  | 0.043181 | 0.040611 | -0.002570 | 0.9405 |
| src/compiler/parser.ts:10712:1 |  | 0.043794 | 0.041214 | -0.002579 | 0.9411 |
| src/compiler/factory/nodeTests.ts:1003:1 | isSpreadAssignment | 0.058282 | 0.055700 | -0.002582 | 0.9557 |
| src/compiler/moduleNameResolver.ts:202:1 | extensionsToExtensionsArray | 0.043651 | 0.041067 | -0.002584 | 0.9408 |
| src/compiler/commandLineParser.ts:4247:1 | removeWildcardFilesWithLowerPriorityExtension | 0.044967 | 0.042376 | -0.002592 | 0.9424 |
| src/compiler/watchUtilities.ts:1:1 |  | 0.045068 | 0.042474 | -0.002594 | 0.9424 |
| src/compiler/scanner.ts:4096:1 |  | 0.069854 | 0.067259 | -0.002595 | 0.9628 |
| src/compiler/factory/utilities.ts:1191:1 | isAdditiveOperatorOrHigher | 0.108693 | 0.106093 | -0.002600 | 0.9761 |
| src/compiler/core.ts:1801:1 | returnFalse | 0.042898 | 0.040293 | -0.002605 | 0.9393 |
| src/compiler/executeCommandLine.ts:176:1 | getOptionsForHelp | 0.043532 | 0.040926 | -0.002606 | 0.9401 |
| src/compiler/utilities.ts:6277:1 | isIntrinsicJsxName | 0.042130 | 0.039514 | -0.002615 | 0.9379 |
| src/compiler/watchPublic.ts:155:1 | WatchStatusReporter | 0.041865 | 0.039249 | -0.002616 | 0.9375 |
| src/compiler/types.ts:1612:1 | BarBarEqualsToken | 0.054227 | 0.051600 | -0.002627 | 0.9516 |
| src/compiler/moduleNameResolver.ts:516:1 | getOriginalAndResolvedFileName | 0.044596 | 0.041959 | -0.002638 | 0.9409 |
| src/compiler/parser.ts:432:1 |  | 0.073976 | 0.071336 | -0.002640 | 0.9643 |
| src/compiler/watchUtilities.ts:381:1 | ProgramUpdateLevel | 0.041514 | 0.038867 | -0.002647 | 0.9362 |
| src/compiler/types.ts:2626:1 | AssignmentOperator | 0.042565 | 0.039907 | -0.002658 | 0.9376 |
| src/compiler/types.ts:9804:1 | BuildInfo | 0.042621 | 0.039961 | -0.002660 | 0.9376 |
| src/compiler/parser.ts:10703:1 | getNamedArgRegEx | 0.042694 | 0.040033 | -0.002661 | 0.9377 |
| src/compiler/watchPublic.ts:408:1 | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile | 0.042238 | 0.039577 | -0.002661 | 0.9370 |
| src/compiler/moduleNameResolver.ts:1768:1 | bundlerModuleNameResolver | 0.044888 | 0.042219 | -0.002669 | 0.9405 |
| src/compiler/utilities.ts:3784:1 | getExternalModuleRequireArgument | 0.270415 | 0.267743 | -0.002671 | 0.9901 |
| src/compiler/transformers/utilities.ts:652:1 | getDecoratorsOfParameters | 0.044650 | 0.041976 | -0.002675 | 0.9401 |
| src/compiler/transformer.ts:669:1 |  | 0.070731 | 0.068048 | -0.002683 | 0.9621 |
| src/compiler/watchPublic.ts:1:1 |  | 0.041738 | 0.039049 | -0.002689 | 0.9356 |
| src/compiler/commandLineParser.ts:1825:1 | OptionsNameMap | 0.041726 | 0.039031 | -0.002695 | 0.9354 |
| src/compiler/utilities.ts:12429:1 | getFirstChild | 0.056840 | 0.054144 | -0.002696 | 0.9526 |
| src/compiler/types.ts:2512:1 | SyntheticExpression | 0.043543 | 0.040841 | -0.002701 | 0.9380 |
| src/compiler/semver.ts:401:1 | isWildcard | 0.042946 | 0.040237 | -0.002709 | 0.9369 |
| src/compiler/commandLineParser.ts:2341:1 |  | 0.071165 | 0.068453 | -0.002712 | 0.9619 |
| src/compiler/utilities.ts:6293:1 | getIndentSize | 0.043140 | 0.040415 | -0.002725 | 0.9368 |
| src/compiler/utilitiesPublic.ts:2451:1 | isModuleReference | 0.048357 | 0.045619 | -0.002738 | 0.9434 |
| src/compiler/core.ts:1886:1 | notImplemented | 0.041064 | 0.038325 | -0.002739 | 0.9333 |
| src/compiler/utilities.ts:11189:1 | hasResolutionModeOverride | 0.044623 | 0.041884 | -0.002739 | 0.9386 |
| src/compiler/factory/nodeTests.ts:528:1 | isThisTypeNode | 0.055770 | 0.053029 | -0.002741 | 0.9508 |
| src/compiler/_namespaces/ts.ts:78:1 |  | 0.042713 | 0.039960 | -0.002753 | 0.9355 |
| src/compiler/watchUtilities.ts:397:1 | SharedExtendedConfigFileWatcher | 0.041732 | 0.038972 | -0.002760 | 0.9339 |
| src/compiler/checker.ts:1154:1 |  | 0.046185 | 0.043414 | -0.002771 | 0.9400 |
| src/compiler/utilities.ts:2434:1 | createDiagnosticForNodeFromMessageChain | 0.052701 | 0.049927 | -0.002774 | 0.9474 |
| src/compiler/resolutionCache.ts:442:1 | getDirectoryOfFailedLookupWatch | 0.042177 | 0.039402 | -0.002775 | 0.9342 |
| src/compiler/transformers/es2015.ts:371:1 | HierarchyFacts | 0.046571 | 0.043795 | -0.002776 | 0.9404 |
| src/compiler/factory/utilities.ts:764:1 | getOrCreateExternalHelpersModuleNameIfNeeded | 0.082755 | 0.079964 | -0.002792 | 0.9663 |
| src/compiler/types.ts:9970:1 | EmitTextWriter | 0.043776 | 0.040981 | -0.002795 | 0.9361 |
| src/compiler/utilities.ts:7592:1 | isRightSideOfInstanceofExpression | 0.292652 | 0.289854 | -0.002797 | 0.9904 |
| src/compiler/utilities.ts:8133:1 | reverseAccessKind | 0.042247 | 0.039447 | -0.002800 | 0.9337 |
| src/compiler/transformers/es2015.ts:367:1 | LoopConverter | 0.043530 | 0.040729 | -0.002802 | 0.9356 |
| src/compiler/utilities.ts:6827:1 | identifierIsThisKeyword | 0.043696 | 0.040878 | -0.002818 | 0.9355 |
| src/compiler/transformers/esDecorators.ts:262:1 | ClassLexicalEnvironmentStackEntry | 0.045199 | 0.042381 | -0.002818 | 0.9377 |
| src/compiler/transformers/typeSerializer.ts:132:1 | createRuntimeTypeSerializer | 0.083901 | 0.081078 | -0.002822 | 0.9664 |
| src/compiler/utilities.ts:1096:1 | isGrammarError | 0.051864 | 0.049039 | -0.002826 | 0.9455 |
| src/compiler/watchPublic.ts:174:1 | ProgramHost | 0.040971 | 0.038137 | -0.002835 | 0.9308 |
| src/compiler/types.ts:9964:1 | DocumentPosition | 0.043899 | 0.041062 | -0.002837 | 0.9354 |
| src/compiler/factory/utilities.ts:556:1 | isInternalName | 0.048197 | 0.045344 | -0.002853 | 0.9408 |
| src/compiler/utilities.ts:7241:1 | getSelectedEffectiveModifierFlags | 0.053278 | 0.050425 | -0.002853 | 0.9464 |
| src/compiler/moduleNameResolver.ts:479:1 | getEffectiveTypeRoots | 0.043742 | 0.040885 | -0.002857 | 0.9347 |
| src/compiler/core.ts:1794:1 | noop | 0.041695 | 0.038837 | -0.002859 | 0.9314 |
| src/compiler/watchPublic.ts:99:1 | ReadBuildProgramHost | 0.040865 | 0.038005 | -0.002859 | 0.9300 |
| src/compiler/types.ts:3338:1 | CommaListExpression | 0.045789 | 0.042927 | -0.002862 | 0.9375 |
| src/compiler/utilities.ts:9692:1 | getPatternFromSpec | 0.042211 | 0.039340 | -0.002870 | 0.9320 |
| src/compiler/emitter.ts:426:1 |  | 0.043297 | 0.040416 | -0.002881 | 0.9335 |
| src/compiler/tsbuildPublic.ts:186:1 | ResolvedConfigFilePath | 0.044073 | 0.041190 | -0.002883 | 0.9346 |
| src/compiler/types.ts:7076:1 | ElementWithComputedPropertyName | 0.042651 | 0.039765 | -0.002886 | 0.9323 |
| src/compiler/watchUtilities.ts:68:1 | DirectoryStructureHost | 0.042565 | 0.039673 | -0.002892 | 0.9321 |
| src/compiler/factory/nodeTests.ts:292:1 | isAsteriskToken | 0.048238 | 0.045343 | -0.002895 | 0.9400 |
| src/compiler/semver.ts:405:1 | createComparator | 0.044156 | 0.041257 | -0.002899 | 0.9344 |
| src/compiler/corePublic.ts:13:1 | MapLike | 0.043151 | 0.040230 | -0.002920 | 0.9323 |
| src/compiler/watchPublic.ts:132:1 | IncrementalProgramOptions | 0.042803 | 0.039882 | -0.002921 | 0.9318 |
| src/compiler/watch.ts:974:1 | IncrementalCompilationOptions | 0.042419 | 0.039496 | -0.002924 | 0.9311 |
| src/compiler/factory/utilities.ts:1240:1 | isBitwiseOperatorOrHigher | 0.111248 | 0.108323 | -0.002925 | 0.9737 |
| src/compiler/types.ts:1624:1 | AssertsKeyword | 0.060194 | 0.057255 | -0.002939 | 0.9512 |
| src/compiler/_namespaces/ts.ts:6:1 |  | 0.042073 | 0.039129 | -0.002944 | 0.9300 |
| src/compiler/factory/utilities.ts:1557:1 | formatGeneratedNamePart | 0.045487 | 0.042542 | -0.002945 | 0.9353 |
| src/compiler/transformers/legacyDecorators.ts:1:1 |  | 0.041631 | 0.038680 | -0.002951 | 0.9291 |
| src/compiler/program.ts:1049:1 | createModuleResolutionLoader | 0.044182 | 0.041224 | -0.002959 | 0.9330 |
| src/compiler/parser.ts:10556:1 | isDeclarationFileName | 0.041552 | 0.038592 | -0.002960 | 0.9288 |
| src/compiler/factory/utilities.ts:424:1 | createExpressionForShorthandPropertyAssignment | 0.094837 | 0.091875 | -0.002962 | 0.9688 |
| src/compiler/utilities.ts:6776:1 | getSetAccessorTypeAnnotationNode | 0.046971 | 0.044002 | -0.002969 | 0.9368 |
| src/compiler/types.ts:6969:1 | TemplateLiteralType | 0.041631 | 0.038662 | -0.002969 | 0.9287 |
| src/compiler/utilitiesPublic.ts:1106:1 | hasJSDocParameterTags | 0.051849 | 0.048878 | -0.002971 | 0.9427 |
| src/compiler/types.ts:2890:1 | TemplateTail | 0.042529 | 0.039558 | -0.002971 | 0.9301 |
| src/compiler/builder.ts:1004:1 | handleDtsMayChangeOfFileAndExportsOfFile | 0.044898 | 0.041921 | -0.002977 | 0.9337 |
| src/compiler/builderState.ts:100:1 | BuilderState | 0.044638 | 0.041660 | -0.002978 | 0.9333 |
| src/compiler/utilitiesPublic.ts:687:1 |  | 0.044311 | 0.041325 | -0.002987 | 0.9326 |
| src/compiler/factory/nodeFactory.ts:7536:1 | mergeTokenSourceMapRanges | 0.043490 | 0.040489 | -0.003001 | 0.9310 |
| src/compiler/factory/utilities.ts:1175:1 | isMultiplicativeOperator | 0.045400 | 0.042396 | -0.003004 | 0.9338 |
| src/compiler/program.ts:1147:1 | getLibraryNameFromLibFileName | 0.041444 | 0.038440 | -0.003004 | 0.9275 |
| src/compiler/utilities.ts:1015:1 | getStartPositionOfLine | 0.044485 | 0.041467 | -0.003019 | 0.9321 |
| src/compiler/moduleNameResolver.ts:3364:1 | loadModuleFromGlobalCache | 0.072648 | 0.069627 | -0.003021 | 0.9584 |
| src/compiler/core.ts:2437:1 | removePrefix | 0.041896 | 0.038868 | -0.003028 | 0.9277 |
| src/compiler/emitter.ts:6327:1 | createBracketsMap | 0.046045 | 0.043010 | -0.003036 | 0.9341 |
| src/compiler/types.ts:5076:1 | TypeChecker | 0.040785 | 0.037731 | -0.003054 | 0.9251 |
| src/compiler/utilities.ts:3033:1 | isValidESSymbolDeclaration | 0.289710 | 0.286653 | -0.003057 | 0.9894 |
| src/compiler/watchUtilities.ts:83:1 | FileAndDirectoryExistence | 0.041887 | 0.038828 | -0.003059 | 0.9270 |
| src/compiler/utilities.ts:914:1 | packageIdToString | 0.042186 | 0.039123 | -0.003063 | 0.9274 |
| src/compiler/types.ts:1599:1 | EndOfFileToken | 0.055873 | 0.052800 | -0.003074 | 0.9450 |
| src/compiler/utilities.ts:9415:1 | getJSXImplicitImportBase | 0.059881 | 0.056801 | -0.003080 | 0.9486 |
| src/compiler/types.ts:2769:1 | LiteralLikeNode | 0.043670 | 0.040588 | -0.003082 | 0.9294 |
| src/compiler/utilities.ts:3337:1 | getSuperContainer | 0.042901 | 0.039817 | -0.003085 | 0.9281 |
| src/compiler/visitorPublic.ts:289:1 | visitArray | 0.044675 | 0.041589 | -0.003086 | 0.9309 |
| src/compiler/types.ts:6951:1 | ConditionalType | 0.043853 | 0.040764 | -0.003089 | 0.9296 |
| src/compiler/utilities.ts:9637:1 |  | 0.074818 | 0.071719 | -0.003099 | 0.9586 |
| src/compiler/moduleNameResolver.ts:2223:1 | GetPackageJsonEntrypointsHost | 0.046314 | 0.043164 | -0.003150 | 0.9320 |
| src/compiler/types.ts:2657:1 | AssignmentExpression | 0.043850 | 0.040678 | -0.003172 | 0.9277 |
| src/compiler/utilitiesPublic.ts:1462:1 | isNodeKind | 0.045956 | 0.042783 | -0.003174 | 0.9309 |
| src/compiler/scanner.ts:4065:1 |  | 0.041987 | 0.038801 | -0.003186 | 0.9241 |
| src/compiler/binder.ts:1:1 |  | 0.043222 | 0.040030 | -0.003191 | 0.9262 |
| src/compiler/commandLineParser.ts:2785:1 | serializeOptionBaseObject | 0.084280 | 0.081087 | -0.003193 | 0.9621 |
| src/compiler/factory/utilities.ts:1229:1 | isEqualityOperatorOrHigher | 0.110211 | 0.107010 | -0.003200 | 0.9710 |
| src/compiler/moduleNameResolver.ts:813:1 | getAutomaticTypeDirectiveNames | 0.043964 | 0.040746 | -0.003219 | 0.9268 |
| src/compiler/core.ts:1810:1 | returnTrue | 0.041231 | 0.038004 | -0.003227 | 0.9217 |
| src/compiler/utilities.ts:4158:1 | isBindableStaticElementAccessExpression | 0.049261 | 0.046020 | -0.003241 | 0.9342 |
| src/compiler/commandLineParser.ts:1718:1 |  | 4.580361 | 4.577119 | -0.003242 | 0.9993 |
| src/compiler/semver.ts:34:1 |  | 0.043136 | 0.039891 | -0.003245 | 0.9248 |
| src/compiler/path.ts:100:1 | pathIsRelative | 0.041664 | 0.038414 | -0.003250 | 0.9220 |
| src/compiler/commandLineParser.ts:1795:1 |  | 0.043201 | 0.039946 | -0.003256 | 0.9246 |
| src/compiler/commandLineParser.ts:2167:1 |  | 0.073190 | 0.069931 | -0.003259 | 0.9555 |
| src/compiler/transformers/esDecorators.ts:227:1 | MemberInfo | 0.043788 | 0.040492 | -0.003296 | 0.9247 |
| src/compiler/utilities.ts:9996:1 | getSupportedExtensions | 0.043137 | 0.039835 | -0.003302 | 0.9234 |
| src/compiler/moduleNameResolver.ts:869:1 | PerDirectoryResolutionCache | 0.042737 | 0.039407 | -0.003330 | 0.9221 |
| src/compiler/utilities.ts:2098:1 | isEffectiveModuleDeclaration | 0.274261 | 0.270906 | -0.003355 | 0.9878 |
| src/compiler/utilities.ts:3335:1 | getSuperContainer | 0.042435 | 0.039080 | -0.003355 | 0.9209 |
| src/compiler/commandLineParser.ts:3501:1 | parseOwnConfigOfJson | 0.042865 | 0.039484 | -0.003381 | 0.9211 |
| src/compiler/utilities.ts:2440:1 | createDiagnosticForNodeArrayFromMessageChain | 0.044053 | 0.040660 | -0.003393 | 0.9230 |
| src/compiler/commandLineParser.ts:2150:1 | ParsedBuildCommand | 0.042160 | 0.038764 | -0.003395 | 0.9195 |
| src/compiler/types.ts:6828:1 | FreshObjectLiteralType | 0.042207 | 0.038811 | -0.003396 | 0.9195 |
| src/compiler/factory/emitNode.ts:132:1 | getSourceMapRange | 0.047480 | 0.044080 | -0.003400 | 0.9284 |
| src/compiler/moduleNameResolver.ts:2989:1 | isApplicableVersionedTypesKey | 0.044819 | 0.041416 | -0.003403 | 0.9241 |
| src/compiler/types.ts:1614:1 | AsteriskToken | 0.058689 | 0.055281 | -0.003408 | 0.9419 |
| src/compiler/utilities.ts:6590:1 | getPossibleOriginalInputPathWithoutChangingExt | 0.043543 | 0.040124 | -0.003419 | 0.9215 |
| src/compiler/factory/utilities.ts:1251:1 | isLogicalOperatorOrHigher | 0.111662 | 0.108243 | -0.003419 | 0.9694 |
| src/compiler/utilities.ts:5690:1 | getOperator | 0.195036 | 0.191617 | -0.003419 | 0.9825 |
| src/compiler/parser.ts:9938:1 | intersectsIncrementalChange | 0.047711 | 0.044292 | -0.003419 | 0.9283 |
| src/compiler/utilitiesPublic.ts:817:1 | getParseTreeNode | 0.043527 | 0.040106 | -0.003421 | 0.9214 |
| src/compiler/corePublic.ts:21:1 | SortedArray | 0.041845 | 0.038416 | -0.003429 | 0.9181 |
| src/compiler/types.ts:2576:1 | EqualityOperatorOrHigher | 0.044334 | 0.040902 | -0.003431 | 0.9226 |
| src/compiler/utilities.ts:3425:1 | isThisInitializedObjectBindingExpression | 0.299501 | 0.296052 | -0.003449 | 0.9885 |
| src/compiler/commandLineParser.ts:1858:1 |  | 0.044437 | 0.040978 | -0.003459 | 0.9222 |
| src/compiler/transformers/classFields.ts:288:1 | PrivateIdentifierStaticFieldInfo | 0.044716 | 0.041245 | -0.003472 | 0.9224 |
| src/compiler/utilities.ts:3020:1 | isCommonJsExportedExpression | 0.286325 | 0.282852 | -0.003473 | 0.9879 |
| src/compiler/commandLineParser.ts:274:1 |  | 0.042797 | 0.039322 | -0.003475 | 0.9188 |
| src/compiler/utilitiesPublic.ts:2183:1 | isNamespaceBody | 0.048031 | 0.044548 | -0.003483 | 0.9275 |
| src/compiler/_namespaces/ts.ts:44:1 |  | 0.045235 | 0.041727 | -0.003508 | 0.9224 |
| src/compiler/transformers/destructuring.ts:522:1 | isSimpleBindingOrAssignmentElement | 0.206557 | 0.203043 | -0.003515 | 0.9830 |
| src/compiler/factory/utilities.ts:1015:1 | tryGetPropertyNameOfBindingOrAssignmentElement | 0.238837 | 0.235289 | -0.003548 | 0.9851 |
| src/compiler/types.ts:7006:1 | SignatureFlags | 0.041463 | 0.037905 | -0.003558 | 0.9142 |
| src/compiler/types.ts:6864:1 | InstantiableType | 0.042325 | 0.038756 | -0.003569 | 0.9157 |
| src/compiler/factory/nodeTests.ts:566:1 | isArrayBindingPattern | 0.049083 | 0.045512 | -0.003571 | 0.9272 |
| src/compiler/utilities.ts:9310:1 |  | 0.101016 | 0.097442 | -0.003574 | 0.9646 |
| src/compiler/transformer.ts:111:1 | SyntaxKindFeatureFlags | 0.044529 | 0.040954 | -0.003575 | 0.9197 |
| src/compiler/commandLineParser.ts:2730:1 | matchesSpecs | 0.083130 | 0.079521 | -0.003609 | 0.9566 |
| src/compiler/utilities.ts:8784:1 | compareDiagnostics | 0.043258 | 0.039645 | -0.003613 | 0.9165 |
| src/compiler/commandLineParser.ts:282:1 |  | 0.043140 | 0.039522 | -0.003618 | 0.9161 |
| src/compiler/utilitiesPublic.ts:2460:1 | isJsxTagNameExpression | 0.050590 | 0.046968 | -0.003622 | 0.9284 |
| src/compiler/types.ts:6210:1 | SymbolTable | 0.043579 | 0.039953 | -0.003626 | 0.9168 |
| src/compiler/types.ts:1621:1 | KeywordToken | 0.060780 | 0.057144 | -0.003635 | 0.9402 |
| src/compiler/types.ts:3180:1 | ImportMetaProperty | 0.044169 | 0.040525 | -0.003644 | 0.9175 |
| src/compiler/moduleNameResolver.ts:173:1 | PathAndExtension | 0.044505 | 0.040854 | -0.003651 | 0.9180 |
| src/compiler/program.ts:877:1 | getModeForResolutionAtIndex | 0.055402 | 0.051745 | -0.003658 | 0.9340 |
| src/compiler/watchUtilities.ts:501:1 | WildcardDirectoryWatcher | 0.040651 | 0.036988 | -0.003663 | 0.9099 |
| src/compiler/watchUtilities.ts:118:1 | createCachedDirectoryStructureHost | 0.042553 | 0.038873 | -0.003679 | 0.9135 |
| src/compiler/core.ts:2109:1 | getUILocale | 0.041887 | 0.038207 | -0.003680 | 0.9121 |
| src/compiler/utilities.ts:11208:1 | getNameFromImportAttribute | 0.197057 | 0.193364 | -0.003693 | 0.9813 |
| src/compiler/corePublic.ts:3:1 |  | 0.042948 | 0.039254 | -0.003694 | 0.9140 |
| src/compiler/utilities.ts:829:1 | getResolvedTypeReferenceDirectiveFromResolution | 0.042538 | 0.038837 | -0.003702 | 0.9130 |
| src/compiler/commandLineParser.ts:3359:1 | canJsonReportNoInputFiles | 0.043935 | 0.040229 | -0.003706 | 0.9156 |
| src/compiler/core.ts:1835:1 | toLowerCase | 0.042805 | 0.039092 | -0.003713 | 0.9132 |
| src/compiler/semver.ts:28:1 |  | 0.043211 | 0.039461 | -0.003750 | 0.9132 |
| src/compiler/moduleNameResolver.ts:422:1 | VersionPaths | 0.043516 | 0.039760 | -0.003756 | 0.9137 |
| src/compiler/binder.ts:495:1 | createFlowNode | 0.055661 | 0.051901 | -0.003760 | 0.9324 |
| src/compiler/utilitiesPublic.ts:1894:1 | isObjectBindingOrAssignmentPattern | 0.045549 | 0.041766 | -0.003783 | 0.9170 |
| src/compiler/utilitiesPublic.ts:1067:1 | getJSDocParameterTags | 0.051972 | 0.048187 | -0.003785 | 0.9272 |
| src/compiler/transformers/declarations/diagnostics.ts:606:1 | createGetIsolatedDeclarationErrors | 0.063907 | 0.060107 | -0.003800 | 0.9405 |
| src/compiler/program.ts:693:1 | getCategoryFormat | 0.047160 | 0.043359 | -0.003800 | 0.9194 |
| src/compiler/program.ts:873:1 | getModeForResolutionAtIndex | 0.056441 | 0.052629 | -0.003812 | 0.9325 |
| src/compiler/executeCommandLine.ts:756:1 | executeCommandLine | 0.428068 | 0.424256 | -0.003812 | 0.9911 |
| src/compiler/moduleNameResolver.ts:894:1 | PerNonRelativeNameCache | 0.042985 | 0.039164 | -0.003821 | 0.9111 |
| src/compiler/utilities.ts:3809:1 | isInJsonFile | 0.046848 | 0.043025 | -0.003823 | 0.9184 |
| src/compiler/utilities.ts:9392:1 | compilerOptionsAffectEmit | 0.060581 | 0.056729 | -0.003852 | 0.9364 |
| src/compiler/utilitiesPublic.ts:1111:1 | getJSDocAugmentsTag | 0.052545 | 0.048690 | -0.003855 | 0.9266 |
| src/compiler/watchUtilities.ts:101:1 | Canonicalized | 0.041248 | 0.037383 | -0.003866 | 0.9063 |
| src/compiler/types.ts:3140:1 | AsExpression | 0.044253 | 0.040386 | -0.003868 | 0.9126 |
| src/compiler/core.ts:1921:1 | AssertionLevel | 0.044180 | 0.040306 | -0.003874 | 0.9123 |
| src/compiler/_namespaces/ts.ts:9:1 |  | 0.044487 | 0.040606 | -0.003882 | 0.9127 |
| src/compiler/utilitiesPublic.ts:788:1 | findAncestor | 0.044586 | 0.040701 | -0.003885 | 0.9129 |
| src/compiler/utilities.ts:7281:1 | getEffectiveModifierFlagsAlwaysIncludeJSDoc | 0.052373 | 0.048479 | -0.003894 | 0.9256 |
| src/compiler/utilities.ts:6216:1 | escapeNonAsciiString | 0.046609 | 0.042711 | -0.003898 | 0.9164 |
| src/compiler/watchUtilities.ts:110:1 | SortedAndCanonicalizedMutableFileSystemEntries | 0.042020 | 0.038120 | -0.003900 | 0.9072 |
| src/compiler/types.ts:2674:1 | BindingOrAssignmentElement | 0.044449 | 0.040548 | -0.003901 | 0.9122 |
| src/compiler/watchPublic.ts:157:1 | CreateProgram | 0.042636 | 0.038735 | -0.003901 | 0.9085 |
| src/compiler/program.ts:1190:1 | isReferenceFileLocation | 0.045547 | 0.041644 | -0.003903 | 0.9143 |
| src/compiler/binder.ts:350:1 | getModuleInstanceStateWorker | 0.151685 | 0.147782 | -0.003903 | 0.9743 |
| src/compiler/utilities.ts:9345:1 | moduleResolutionSupportsPackageJsonExportsAndImports | 0.060207 | 0.056296 | -0.003912 | 0.9350 |
| src/compiler/factory/utilities.ts:1246:1 | isLogicalOperator | 0.046257 | 0.042337 | -0.003920 | 0.9153 |
| src/compiler/transformers/es2015.ts:478:1 | SpreadSegment | 0.044690 | 0.040755 | -0.003934 | 0.9120 |
| src/compiler/utilities.ts:6782:1 | getThisParameter | 0.046005 | 0.042057 | -0.003948 | 0.9142 |
| src/compiler/utilities.ts:12423:1 | addEmitFlagsRecursively | 0.059680 | 0.055726 | -0.003954 | 0.9337 |
| src/compiler/builderState.ts:39:1 | getFileEmitOutput | 0.158226 | 0.154265 | -0.003960 | 0.9750 |
| src/compiler/types.ts:3075:1 | BindableElementAccessExpression | 0.045541 | 0.041574 | -0.003966 | 0.9129 |
| src/compiler/sourcemap.ts:377:1 | getLineInfo | 0.047159 | 0.043190 | -0.003969 | 0.9158 |
| src/compiler/types.ts:9874:1 | PrinterOptions | 0.044636 | 0.040659 | -0.003977 | 0.9109 |
| src/compiler/transformers/es2017.ts:1:1 |  | 0.044368 | 0.040384 | -0.003984 | 0.9102 |
| src/compiler/utilities.ts:10464:1 | parsePseudoBigInt | 0.045896 | 0.041912 | -0.003984 | 0.9132 |
| src/compiler/utilitiesPublic.ts:997:1 | getNameOfDeclaration | 0.220530 | 0.216526 | -0.004004 | 0.9818 |
| src/compiler/sourcemap.ts:363:1 |  | 0.044612 | 0.040600 | -0.004011 | 0.9101 |
| src/compiler/program.ts:1175:1 | ReferenceFileLocation | 0.042685 | 0.038646 | -0.004039 | 0.9054 |
| src/compiler/utilities.ts:6749:1 | getLineOfLocalPosition | 0.044930 | 0.040886 | -0.004044 | 0.9100 |
| src/compiler/utilities.ts:9354:1 | moduleSupportsImportAttributes | 0.060580 | 0.056520 | -0.004060 | 0.9330 |
| src/compiler/visitorPublic.ts:420:1 | addDefaultValueAssignmentsIfNeeded | 0.046424 | 0.042347 | -0.004077 | 0.9122 |
| src/compiler/commandLineParser.ts:2132:1 | getOptionFromName | 0.044743 | 0.040665 | -0.004079 | 0.9088 |
| src/compiler/types.ts:5761:1 | AnyImportOrReExport | 0.046277 | 0.042193 | -0.004084 | 0.9117 |
| src/compiler/watch.ts:688:1 |  | 0.070126 | 0.066032 | -0.004094 | 0.9416 |
| src/compiler/utilities.ts:3293:1 | getNewTargetContainer | 0.049320 | 0.045224 | -0.004096 | 0.9169 |
| src/compiler/utilities.ts:3471:1 | nodeCanBeDecorated | 0.043671 | 0.039541 | -0.004130 | 0.9054 |
| src/compiler/utilities.ts:3413:1 | isThisProperty | 0.053520 | 0.049377 | -0.004144 | 0.9226 |
| src/compiler/watch.ts:670:1 |  | 0.042718 | 0.038562 | -0.004156 | 0.9027 |
| src/compiler/path.ts:50:1 | isUrl | 0.043585 | 0.039425 | -0.004160 | 0.9045 |
| src/compiler/utilitiesPublic.ts:1480:1 | isToken | 0.056911 | 0.052735 | -0.004176 | 0.9266 |
| src/compiler/program.ts:1027:1 | ResolutionNameAndModeGetter | 0.043466 | 0.039286 | -0.004180 | 0.9038 |
| src/compiler/utilities.ts:10148:1 | isSupportedSourceFileName | 0.043247 | 0.039058 | -0.004188 | 0.9031 |
| src/compiler/utilitiesPublic.ts:865:1 | identifierToKeywordKind | 0.090304 | 0.086106 | -0.004199 | 0.9535 |
| src/compiler/program.ts:509:1 | CompilerHostLikeWithCache | 0.044158 | 0.039956 | -0.004202 | 0.9049 |
| src/compiler/types.ts:1756:1 | Declaration | 0.057752 | 0.053539 | -0.004213 | 0.9271 |
| src/compiler/utilities.ts:8087:1 | isWriteAccess | 0.050863 | 0.046639 | -0.004224 | 0.9170 |
| src/compiler/moduleNameResolver.ts:3169:1 | tryLoadModuleUsingPaths | 0.102984 | 0.098753 | -0.004231 | 0.9589 |
| src/compiler/utilities.ts:3454:1 | getInvokedExpression | 0.052187 | 0.047954 | -0.004232 | 0.9189 |
| src/compiler/core.ts:1988:1 | compareValues | 0.041818 | 0.037568 | -0.004250 | 0.8984 |
| src/compiler/types.ts:3378:1 | VariableStatement | 0.046042 | 0.041789 | -0.004253 | 0.9076 |
| src/compiler/executeCommandLine.ts:531:1 | getHeader | 0.044554 | 0.040291 | -0.004263 | 0.9043 |
| src/compiler/factory/emitHelpers.ts:53:1 | ESDecorateClassContext | 0.044483 | 0.040194 | -0.004289 | 0.9036 |
| src/compiler/utilities.ts:9433:1 | getJSXRuntimeImport | 0.058794 | 0.054505 | -0.004290 | 0.9270 |
| src/compiler/moduleSpecifiers.ts:164:1 | RelativePreference | 0.046331 | 0.042009 | -0.004322 | 0.9067 |
| src/compiler/utilitiesPublic.ts:2504:1 | isCaseOrDefaultClause | 0.051365 | 0.047041 | -0.004324 | 0.9158 |
| src/compiler/types.ts:2935:1 | ArrayLiteralExpression | 0.050456 | 0.046120 | -0.004336 | 0.9141 |
| src/compiler/utilities.ts:6729:1 | writeFileEnsuringDirectories | 0.045965 | 0.041614 | -0.004350 | 0.9054 |
| src/compiler/utilities.ts:3073:1 | isObjectLiteralMethod | 0.050830 | 0.046479 | -0.004351 | 0.9144 |
| src/compiler/utilitiesPublic.ts:842:1 | escapeLeadingUnderscores | 0.044892 | 0.040532 | -0.004361 | 0.9029 |
| src/compiler/core.ts:1299:1 | getAllKeys | 0.045982 | 0.041614 | -0.004368 | 0.9050 |
| src/compiler/debug.ts:99:1 | LogLevel | 0.043737 | 0.039355 | -0.004382 | 0.8998 |
| src/compiler/utilities.ts:8586:1 | formatStringFromArgs | 0.048531 | 0.044137 | -0.004394 | 0.9095 |
| src/compiler/utilities.ts:9584:1 | stripLeadingDirectorySeparator | 0.099882 | 0.095480 | -0.004402 | 0.9559 |
| src/compiler/factory/utilities.ts:658:1 | skipOuterExpressions | 0.046219 | 0.041809 | -0.004410 | 0.9046 |
| src/compiler/utilities.ts:9324:1 |  | 0.103877 | 0.099457 | -0.004420 | 0.9574 |
| src/compiler/moduleNameResolver.ts:155:1 | Resolved | 0.044264 | 0.039839 | -0.004425 | 0.9000 |
| src/compiler/utilities.ts:3308:1 | SuperContainer | 0.044207 | 0.039772 | -0.004435 | 0.8997 |
| src/compiler/utilities.ts:4299:1 | isSpecialPropertyDeclaration | 0.197640 | 0.193159 | -0.004482 | 0.9773 |
| src/compiler/scanner.ts:4036:1 | codePointAt | 0.043503 | 0.039003 | -0.004500 | 0.8966 |
| src/compiler/programDiagnostics.ts:1:1 |  | 0.045189 | 0.040678 | -0.004511 | 0.9002 |
| src/compiler/types.ts:6932:1 | IndexType | 0.043275 | 0.038749 | -0.004526 | 0.8954 |
| src/compiler/factory/nodeTests.ts:372:1 | isAccessorModifier | 0.059962 | 0.055427 | -0.004536 | 0.9244 |
| src/compiler/factory/utilities.ts:1551:1 | formatGeneratedNamePart | 0.046597 | 0.042056 | -0.004542 | 0.9025 |
| src/compiler/types.ts:6167:1 | ReverseMappedSymbol | 0.043797 | 0.039242 | -0.004555 | 0.8960 |
| src/compiler/types.ts:7229:1 | DiagnosticMessage | 0.045700 | 0.041136 | -0.004564 | 0.9001 |
| src/compiler/utilities.ts:3814:1 | isSourceFileNotJson | 0.223498 | 0.218923 | -0.004575 | 0.9795 |
| src/compiler/types.ts:1589:1 | NodeArray | 0.056998 | 0.052421 | -0.004577 | 0.9197 |
| src/compiler/scanner.ts:762:1 | scanConflictMarkerTrivia | 0.044088 | 0.039499 | -0.004589 | 0.8959 |
| src/compiler/commandLineParser.ts:2764:1 | getNameOfCompilerOptionValue | 0.045415 | 0.040823 | -0.004593 | 0.8989 |
| src/compiler/utilitiesPublic.ts:384:1 | textSpanOverlap | 0.045128 | 0.040532 | -0.004597 | 0.8981 |
| src/compiler/types.ts:10023:1 | ModulePath | 0.045266 | 0.040669 | -0.004597 | 0.8984 |
| src/compiler/resolutionCache.ts:269:1 | perceivedOsRootLengthForWatching | 0.045846 | 0.041248 | -0.004597 | 0.8997 |
| src/compiler/builder.ts:889:1 | removeSemanticDiagnosticsOf | 0.044538 | 0.039939 | -0.004599 | 0.8967 |
| src/compiler/types.ts:2486:1 | DeleteExpression | 0.044371 | 0.039763 | -0.004608 | 0.8961 |
| src/compiler/types.ts:10451:1 | UserPreferences | 0.045226 | 0.040618 | -0.004609 | 0.8981 |
| src/compiler/utilitiesPublic.ts:2468:1 | isJsxChild | 0.050012 | 0.045390 | -0.004623 | 0.9076 |
| src/compiler/program.ts:958:1 | getEmitSyntaxForUsageLocationWorker | 0.296214 | 0.291586 | -0.004628 | 0.9844 |
| src/compiler/utilitiesPublic.ts:827:1 | getParseTreeNode | 0.044269 | 0.039629 | -0.004640 | 0.8952 |
| src/compiler/moduleSpecifiers.ts:926:1 | tryGetModuleNameFromPaths | 0.120705 | 0.116048 | -0.004657 | 0.9614 |
| src/compiler/commandLineParser.ts:3339:1 | getErrorForNoInputFiles | 0.122671 | 0.117996 | -0.004675 | 0.9619 |
| src/compiler/core.ts:1773:1 | isNumber | 0.042905 | 0.038187 | -0.004717 | 0.8901 |
| src/compiler/utilities.ts:2157:1 | isEffectiveExternalModule | 0.046886 | 0.042158 | -0.004728 | 0.8992 |
| src/compiler/program.ts:876:1 | getModeForResolutionAtIndex | 0.056890 | 0.052155 | -0.004735 | 0.9168 |
| src/compiler/types.ts:6511:1 | BigIntLiteralType | 0.043552 | 0.038816 | -0.004736 | 0.8913 |
| src/compiler/transformers/es2017.ts:104:1 | SuperContainer | 0.045968 | 0.041229 | -0.004739 | 0.8969 |
| src/compiler/transformers/declarations.ts:1873:1 | maskModifiers | 0.091759 | 0.087010 | -0.004748 | 0.9483 |
| src/compiler/utilitiesPublic.ts:1647:1 | isModifier | 0.084186 | 0.079436 | -0.004750 | 0.9436 |
| src/compiler/factory/utilities.ts:565:1 | isLocalName | 0.048704 | 0.043931 | -0.004773 | 0.9020 |
| src/compiler/executeCommandLine.ts:169:1 | shouldBePretty | 0.045597 | 0.040821 | -0.004776 | 0.8953 |
| src/compiler/types.ts:3448:1 | ContinueStatement | 0.046058 | 0.041282 | -0.004776 | 0.8963 |
| src/compiler/factory/nodeTests.ts:426:1 | isMethodSignature | 0.051724 | 0.046936 | -0.004788 | 0.9074 |
| src/compiler/factory/utilities.ts:1163:1 | isQuestionOrPlusOrMinusToken | 0.272572 | 0.267779 | -0.004793 | 0.9824 |
| src/compiler/factory/utilities.ts:624:1 | getJSDocTypeAssertionType | 0.059474 | 0.054679 | -0.004795 | 0.9194 |
| src/compiler/watch.ts:827:1 | setGetSourceFileAsHashVersioned | 0.081433 | 0.076612 | -0.004822 | 0.9408 |
| src/compiler/semver.ts:409:1 | testDisjunction | 0.045046 | 0.040220 | -0.004827 | 0.8929 |
| src/compiler/moduleNameResolver.ts:925:1 | isPackageJsonInfo | 0.046557 | 0.041721 | -0.004835 | 0.8961 |
| src/compiler/types.ts:3796:1 | NamedExports | 0.046479 | 0.041642 | -0.004836 | 0.8959 |
| src/compiler/executeCommandLine.ts:184:1 | printVersion | 0.085292 | 0.080452 | -0.004840 | 0.9432 |
| src/compiler/types.ts:6974:1 | StringMappingType | 0.045393 | 0.040540 | -0.004853 | 0.8931 |
| src/compiler/moduleNameResolver.ts:526:1 | getCandidateFromTypeRoot | 0.045978 | 0.041117 | -0.004860 | 0.8943 |
| src/compiler/transformers/es2017.ts:112:1 | ContextFlags | 0.045088 | 0.040218 | -0.004870 | 0.8920 |
| src/compiler/transformers/destructuring.ts:541:1 | createDefaultValueCheck | 0.138546 | 0.133660 | -0.004886 | 0.9647 |
| src/compiler/factory/nodeTests.ts:915:1 | isSyntheticReference | 0.054662 | 0.049771 | -0.004892 | 0.9105 |
| src/compiler/types.ts:1618:1 | QuestionDotToken | 0.061657 | 0.056732 | -0.004925 | 0.9201 |
| src/compiler/types.ts:4509:1 | TsConfigSourceFile | 0.044655 | 0.039728 | -0.004927 | 0.8897 |
| src/compiler/utilities.ts:5112:1 | isLiteralComputedPropertyDeclarationName | 0.296685 | 0.291739 | -0.004946 | 0.9833 |
| src/compiler/types.ts:3080:1 | BindableStaticAccessExpression | 0.046754 | 0.041790 | -0.004965 | 0.8938 |
| src/compiler/commandLineParser.ts:3690:1 | getExtendedConfig | 0.048634 | 0.043644 | -0.004991 | 0.8974 |
| src/compiler/utilities.ts:7246:1 | getSelectedSyntacticModifierFlags | 0.056907 | 0.051916 | -0.004991 | 0.9123 |
| src/compiler/utilitiesPublic.ts:2426:1 | isBlockStatement | 0.058374 | 0.053381 | -0.004992 | 0.9145 |
| src/compiler/scanner.ts:39:1 | ErrorCallback | 0.045691 | 0.040696 | -0.004995 | 0.8907 |
| src/compiler/factory/nodeTests.ts:450:1 | isCallSignatureDeclaration | 0.057531 | 0.052535 | -0.004996 | 0.9132 |
| src/compiler/utilities.ts:9308:1 |  | 0.103394 | 0.098384 | -0.005010 | 0.9515 |
| src/compiler/factory/nodeTests.ts:1054:1 | isJSDocUnknownType | 0.050902 | 0.045867 | -0.005035 | 0.9011 |
| src/compiler/factory/utilities.ts:987:1 | getRestIndicatorOfBindingOrAssignmentElement | 0.047837 | 0.042796 | -0.005042 | 0.8946 |
| src/compiler/types.ts:1627:1 | CaseKeyword | 0.056079 | 0.051035 | -0.005044 | 0.9101 |
| src/compiler/utilities.ts:5611:1 | nodeStartsNewLexicalEnvironment | 0.049758 | 0.044699 | -0.005059 | 0.8983 |
| src/compiler/commandLineParser.ts:2781:1 | serializeWatchOptions | 0.046994 | 0.041932 | -0.005062 | 0.8923 |
| src/compiler/sourcemap.ts:365:1 |  | 0.045639 | 0.040573 | -0.005066 | 0.8890 |
| src/compiler/factory/emitHelpers.ts:755:1 |  | 0.089774 | 0.084694 | -0.005081 | 0.9434 |
| src/compiler/watch.ts:740:1 | WatchFactoryWithLog | 0.043781 | 0.038667 | -0.005114 | 0.8832 |
| src/compiler/parser.ts:10767:1 | addPragmaForMatch | 0.045239 | 0.040089 | -0.005150 | 0.8862 |
| src/compiler/emitter.ts:1191:1 | PipelinePhase | 0.048132 | 0.042976 | -0.005156 | 0.8929 |
| src/compiler/utilities.ts:7965:1 | getStartPositionOfRange | 0.044897 | 0.039710 | -0.005187 | 0.8845 |
| src/compiler/scanner.ts:333:1 |  | 0.048537 | 0.043342 | -0.005195 | 0.8930 |
| src/compiler/watchUtilities.ts:463:1 | cleanExtendedConfigCache | 0.087369 | 0.082159 | -0.005210 | 0.9404 |
| src/compiler/utilities.ts:2740:1 | isHoistedFunction | 0.255728 | 0.250502 | -0.005226 | 0.9796 |
| src/compiler/semver.ts:158:1 | comparePrereleaseIdentifiers | 0.047827 | 0.042597 | -0.005230 | 0.8906 |
| src/compiler/utilities.ts:2729:1 | isPrologueDirective | 0.133253 | 0.128010 | -0.005243 | 0.9607 |
| src/compiler/_namespaces/ts.ts:48:1 |  | 0.046578 | 0.041304 | -0.005274 | 0.8868 |
| src/compiler/moduleNameResolver.ts:193:1 | formatExtensions | 0.046418 | 0.041135 | -0.005283 | 0.8862 |
| src/compiler/core.ts:2442:1 | tryRemovePrefix | 0.045453 | 0.040157 | -0.005295 | 0.8835 |
| src/compiler/utilitiesPublic.ts:2477:1 | isJsxAttributeLike | 0.055950 | 0.050615 | -0.005335 | 0.9046 |
| src/compiler/types.ts:10427:1 | ReadonlyPragmaMap | 0.047243 | 0.041897 | -0.005345 | 0.8869 |
| src/compiler/utilities.ts:10233:1 | tryParsePatterns | 0.044278 | 0.038918 | -0.005360 | 0.8789 |
| src/compiler/_namespaces/ts.ts:67:1 |  | 0.044314 | 0.038931 | -0.005383 | 0.8785 |
| src/compiler/transformers/legacyDecorators.ts:88:1 | transformLegacyDecorators | 0.089172 | 0.083777 | -0.005395 | 0.9395 |
| src/compiler/utilities.ts:2445:1 | assertDiagnosticLocation | 0.045163 | 0.039750 | -0.005413 | 0.8801 |
| src/compiler/types.ts:7398:1 | WatchDirectoryKind | 0.047114 | 0.041697 | -0.005417 | 0.8850 |
| src/compiler/visitorPublic.ts:223:1 | visitNodes | 0.048025 | 0.042607 | -0.005418 | 0.8872 |
| src/compiler/core.ts:2133:1 | compareStringsCaseSensitiveUI | 0.043129 | 0.037703 | -0.005427 | 0.8742 |
| src/compiler/utilitiesPublic.ts:348:1 | textSpanEnd | 0.043965 | 0.038523 | -0.005442 | 0.8762 |
| src/compiler/utilities.ts:10949:1 | createPropertyNameNodeForIdentifierOrLiteral | 0.159721 | 0.154277 | -0.005445 | 0.9659 |
| src/compiler/utilities.ts:6615:1 | EmitFileNames | 0.045677 | 0.040218 | -0.005460 | 0.8805 |
| src/compiler/moduleNameResolver.ts:726:1 | getNodeResolutionFeatures | 0.048363 | 0.042903 | -0.005460 | 0.8871 |
| src/compiler/semver.ts:249:1 |  | 0.045874 | 0.040402 | -0.005473 | 0.8807 |
| src/compiler/types.ts:6986:1 | SubstitutionType | 0.044592 | 0.039111 | -0.005481 | 0.8771 |
| src/compiler/types.ts:6437:1 | DestructuringPattern | 0.044858 | 0.039371 | -0.005486 | 0.8777 |
| src/compiler/moduleNameResolver.ts:3236:1 | unmangleScopedPackageName | 0.054851 | 0.049360 | -0.005491 | 0.8999 |
| src/compiler/commandLineParser.ts:3799:1 | createDiagnosticForNodeInSourceFileOrCompilerDiagnostic | 0.058673 | 0.053171 | -0.005502 | 0.9062 |
| src/compiler/program.ts:1228:1 | isProgramUptoDate | 0.091628 | 0.086115 | -0.005513 | 0.9398 |
| src/compiler/utilities.ts:7540:1 | tryGetPropertyAccessOrIdentifierToString | 0.212449 | 0.206927 | -0.005522 | 0.9740 |
| src/compiler/commandLineParser.ts:2320:1 |  | 0.077016 | 0.071470 | -0.005545 | 0.9280 |
| src/compiler/watchPublic.ts:330:1 | WatchCompilerHostOfConfigFile | 0.043991 | 0.038418 | -0.005573 | 0.8733 |
| src/compiler/utilities.ts:7575:1 | isRightSideOfAccessExpression | 0.270976 | 0.265397 | -0.005579 | 0.9794 |
| src/compiler/utilities.ts:3525:1 | nodeIsDecorated | 0.059016 | 0.053429 | -0.005587 | 0.9053 |
| src/compiler/transformers/esnext.ts:792:1 | getUsingKindOfStatements | 0.069378 | 0.063763 | -0.005615 | 0.9191 |
| src/compiler/commandLineParser.ts:1733:1 |  | 0.071310 | 0.065675 | -0.005635 | 0.9210 |
| src/compiler/utilities.ts:1290:1 | isJSDocTypeExpressionOrChild | 0.051334 | 0.045687 | -0.005647 | 0.8900 |
| src/compiler/expressionToTypeNode.ts:165:1 | syntacticResult | 0.051501 | 0.045837 | -0.005664 | 0.8900 |
| src/compiler/utilities.ts:9298:1 |  | 0.111447 | 0.105770 | -0.005677 | 0.9491 |
| src/compiler/moduleNameResolver.ts:1327:1 | createModuleResolutionCache | 0.245220 | 0.239539 | -0.005681 | 0.9768 |
| src/compiler/moduleSpecifiers.ts:1:1 |  | 0.046091 | 0.040404 | -0.005687 | 0.8766 |
| src/compiler/factory/utilities.ts:1559:1 | formatGeneratedNamePart | 0.047942 | 0.042239 | -0.005703 | 0.8810 |
| src/compiler/resolutionCache.ts:400:1 | getDirectoryToWatchFromFailedLookupLocationDirectory | 0.045981 | 0.040250 | -0.005731 | 0.8754 |
| src/compiler/types.ts:10085:1 | ErrorOutputContainer | 0.048739 | 0.042999 | -0.005740 | 0.8822 |
| src/compiler/utilitiesPublic.ts:371:1 | textSpanContainsTextRange | 0.045832 | 0.040081 | -0.005751 | 0.8745 |
| src/compiler/utilities.ts:2246:1 | isDeclarationWithTypeParameterChildren | 0.120659 | 0.114871 | -0.005788 | 0.9520 |
| src/compiler/utilities.ts:9387:1 | compilerOptionsAffectSemanticDiagnostics | 0.062676 | 0.056888 | -0.005788 | 0.9076 |
| src/compiler/commandLineParser.ts:1745:1 |  | 0.044255 | 0.038463 | -0.005791 | 0.8691 |
| src/compiler/sourcemap.ts:371:1 | LineInfo | 0.052858 | 0.047053 | -0.005804 | 0.8902 |
| src/compiler/watchUtilities.ts:481:1 | updateMissingFilePathsWatch | 0.045369 | 0.039547 | -0.005822 | 0.8717 |
| src/compiler/checker.ts:1377:1 | MappedTypeModifiers | 0.046624 | 0.040772 | -0.005852 | 0.8745 |
| src/compiler/utilities.ts:9580:1 | isNodeModulesOrScopedPackageDirectory | 0.106602 | 0.100739 | -0.005863 | 0.9450 |
| src/compiler/moduleSpecifiers.ts:133:1 |  | 0.047444 | 0.041562 | -0.005883 | 0.8760 |
| src/compiler/moduleNameResolver.ts:882:1 | NonRelativeNameResolutionCache | 0.045150 | 0.039264 | -0.005886 | 0.8696 |
| src/compiler/utilities.ts:2117:1 | isBlockScopedContainerTopLevel | 0.238497 | 0.232603 | -0.005895 | 0.9753 |
| src/compiler/utilities.ts:6214:1 |  | 0.048253 | 0.042353 | -0.005900 | 0.8777 |
| src/compiler/moduleSpecifiers.ts:562:1 | Info | 0.046037 | 0.040129 | -0.005908 | 0.8717 |
| src/compiler/factory/emitHelpers.ts:802:1 |  | 0.082286 | 0.076359 | -0.005927 | 0.9280 |
| src/compiler/core.ts:2392:1 | patternText | 0.123004 | 0.117077 | -0.005927 | 0.9518 |
| src/compiler/types.ts:2631:1 | AssignmentOperatorOrHigher | 0.046532 | 0.040592 | -0.005940 | 0.8724 |
| src/compiler/utilities.ts:2111:1 | isShorthandAmbientModule | 0.134213 | 0.128250 | -0.005964 | 0.9556 |
| src/compiler/transformers/classFields.ts:3335:1 | createPrivateInstanceMethodInitializer | 0.139751 | 0.133785 | -0.005966 | 0.9573 |
| src/compiler/utilities.ts:3543:1 | nodeOrChildIsDecorated | 0.073827 | 0.067845 | -0.005982 | 0.9190 |
| src/compiler/utilitiesPublic.ts:1948:1 | isPropertyAccessOrQualifiedNameOrImportTypeNode | 0.073602 | 0.067594 | -0.006008 | 0.9184 |
| src/compiler/utilities.ts:8574:1 | addObjectAllocatorPatcher | 0.047341 | 0.041298 | -0.006044 | 0.8723 |
| src/compiler/watch.ts:668:1 |  | 0.073431 | 0.067387 | -0.006044 | 0.9177 |
| src/compiler/utilities.ts:4307:1 | setValueDeclaration | 0.096938 | 0.090862 | -0.006076 | 0.9373 |
| src/compiler/types.ts:1602:1 | PunctuationToken | 0.059892 | 0.053816 | -0.006076 | 0.8985 |
| src/compiler/checker.ts:1169:1 |  | 0.046586 | 0.040507 | -0.006078 | 0.8695 |
| src/compiler/utilitiesPublic.ts:366:1 | textSpanContainsTextSpan | 0.046887 | 0.040807 | -0.006080 | 0.8703 |
| src/compiler/moduleNameResolver.ts:787:1 | resolvePackageNameToPackageJson | 0.087322 | 0.081192 | -0.006130 | 0.9298 |
| src/compiler/_namespaces/ts.ts:8:1 |  | 0.046749 | 0.040609 | -0.006140 | 0.8687 |
| src/compiler/utilities.ts:8082:1 | isWriteOnlyAccess | 0.052978 | 0.046825 | -0.006153 | 0.8839 |
| src/compiler/emitter.ts:668:1 | getComputedCommonSourceDirectory | 0.045772 | 0.039616 | -0.006157 | 0.8655 |
| src/compiler/transformers/declarations.ts:1916:1 | isPreservedDeclarationStatement | 0.048761 | 0.042604 | -0.006157 | 0.8737 |
| src/compiler/types.ts:6767:1 | IntersectionType | 0.044825 | 0.038654 | -0.006171 | 0.8623 |
| src/compiler/diagnosticInformationMap.generated.ts:12:1 |  | 0.183142 | 0.176937 | -0.006206 | 0.9661 |
| src/compiler/checker.ts:1451:1 |  | 0.047001 | 0.040782 | -0.006218 | 0.8677 |
| src/compiler/types.ts:3390:1 | PrologueDirective | 0.047750 | 0.041527 | -0.006223 | 0.8697 |
| src/compiler/checker.ts:1151:1 |  | 0.048268 | 0.042002 | -0.006266 | 0.8702 |
| src/compiler/core.ts:1861:1 |  | 0.044771 | 0.038491 | -0.006279 | 0.8597 |
| src/compiler/scanner.ts:1023:1 | createScanner | 0.101868 | 0.095585 | -0.006283 | 0.9383 |
| src/compiler/utilities.ts:7231:1 | hasEffectiveReadonlyModifier | 0.055356 | 0.049065 | -0.006291 | 0.8864 |
| src/compiler/utilitiesPublic.ts:629:1 | isEmptyBindingElement | 0.202717 | 0.196396 | -0.006320 | 0.9688 |
| src/compiler/factory/nodeTests.ts:422:1 | isPropertyDeclaration | 0.052260 | 0.045913 | -0.006346 | 0.8786 |
| src/compiler/commandLineParser.ts:288:1 |  | 0.073430 | 0.067080 | -0.006350 | 0.9135 |
| src/compiler/factory/nodeFactory.ts:1:1 |  | 0.046167 | 0.039805 | -0.006361 | 0.8622 |
| src/compiler/types.ts:1616:1 | PlusToken | 0.061325 | 0.054961 | -0.006364 | 0.8962 |
| src/compiler/types.ts:6999:1 | SignatureKind | 0.044890 | 0.038493 | -0.006397 | 0.8575 |
| src/compiler/program.ts:5128:1 | resolveProjectReferencePath | 0.045476 | 0.039072 | -0.006403 | 0.8592 |
| src/compiler/transformers/destructuring.ts:99:1 | flattenDestructuringAssignment | 0.101325 | 0.094914 | -0.006412 | 0.9367 |
| src/compiler/utilities.ts:959:1 | aggregateChildData | 0.055762 | 0.049300 | -0.006462 | 0.8841 |
| src/compiler/utilities.ts:9329:1 | emitModuleKindIsNonNodeESM | 0.061128 | 0.054664 | -0.006464 | 0.8943 |
| src/compiler/factory/nodeTests.ts:430:1 | isMethodDeclaration | 0.053116 | 0.046651 | -0.006465 | 0.8783 |
| src/compiler/types.ts:10596:1 | OrganizeImportsTypeOrder | 0.046315 | 0.039848 | -0.006467 | 0.8604 |
| src/compiler/performanceCore.ts:102:1 |  | 0.087707 | 0.081236 | -0.006471 | 0.9262 |
| src/compiler/factory/utilities.ts:631:1 | isOuterExpression | 0.282437 | 0.275929 | -0.006509 | 0.9770 |
| src/compiler/utilities.ts:3184:1 | getThisContainer | 0.046578 | 0.040049 | -0.006529 | 0.8598 |
| src/compiler/tracing.ts:379:1 |  | 0.048080 | 0.041537 | -0.006542 | 0.8639 |
| src/compiler/_namespaces/ts.ts:13:1 |  | 0.046428 | 0.039876 | -0.006551 | 0.8589 |
| src/compiler/commandLineParser.ts:4174:1 | CanonicalKey | 0.048950 | 0.042354 | -0.006596 | 0.8653 |
| src/compiler/utilitiesPublic.ts:2495:1 | isJsxCallLike | 0.056132 | 0.049529 | -0.006602 | 0.8824 |
| src/compiler/core.ts:1819:1 | returnUndefined | 0.045344 | 0.038730 | -0.006614 | 0.8541 |
| src/compiler/types.ts:3563:1 | ClassLikeDeclaration | 0.048910 | 0.042260 | -0.006649 | 0.8640 |
| src/compiler/utilities.ts:8290:1 | showModuleSpecifier | 0.324202 | 0.317535 | -0.006667 | 0.9794 |
| src/compiler/utilitiesPublic.ts:2086:1 | isExpression | 0.058784 | 0.052102 | -0.006682 | 0.8863 |
| src/compiler/binder.ts:502:1 | bindSourceFile | 0.045845 | 0.039162 | -0.006683 | 0.8542 |
| src/compiler/moduleNameResolver.ts:856:1 | ModeAwareCache | 0.047392 | 0.040678 | -0.006714 | 0.8583 |
| src/compiler/factory/utilities.ts:607:1 | isCommaExpression | 0.177656 | 0.170929 | -0.006726 | 0.9621 |
| src/compiler/factory/emitHelpers.ts:1451:1 |  | 0.104275 | 0.097538 | -0.006737 | 0.9354 |
| src/compiler/watch.ts:673:1 | createWatchHost | 0.047105 | 0.040362 | -0.006743 | 0.8569 |
| src/compiler/moduleNameResolver.ts:2751:1 | getLoadModuleFromTargetExportOrImport | 0.051466 | 0.044698 | -0.006768 | 0.8685 |
| src/compiler/types.ts:1595:1 | Token | 0.057640 | 0.050846 | -0.006793 | 0.8821 |
| src/compiler/core.ts:2589:1 |  | 0.046402 | 0.039605 | -0.006797 | 0.8535 |
| src/compiler/moduleNameResolver.ts:123:1 | isTraceEnabled | 0.085708 | 0.078889 | -0.006819 | 0.9204 |
| src/compiler/moduleNameResolver.ts:1778:1 | nodeModuleNameResolver | 0.048817 | 0.041994 | -0.006823 | 0.8602 |
| src/compiler/types.ts:1605:1 | DotToken | 0.061930 | 0.055097 | -0.006832 | 0.8897 |
| src/compiler/types.ts:3558:1 | ClassExpression | 0.048880 | 0.042041 | -0.006839 | 0.8601 |
| src/compiler/utilities.ts:5189:1 | getPropertyAssignmentAliasLikeExpression | 0.195686 | 0.188825 | -0.006861 | 0.9649 |
| src/compiler/watch.ts:788:1 | getSourceFileVersionAsHashFromText | 0.085051 | 0.078184 | -0.006867 | 0.9193 |
| src/compiler/types.ts:3536:1 | DeclarationWithTypeParameterChildren | 0.051120 | 0.044213 | -0.006907 | 0.8649 |
| src/compiler/utilitiesPublic.ts:693:1 | validateLocaleAndSetLanguage | 0.219153 | 0.212187 | -0.006966 | 0.9682 |
| src/compiler/utilities.ts:9985:1 |  | 0.060399 | 0.053427 | -0.006971 | 0.8846 |
| src/compiler/types.ts:3070:1 | BindableStaticElementAccessExpression | 0.048185 | 0.041212 | -0.006973 | 0.8553 |
| src/compiler/utilities.ts:7992:1 | getPreviousNonWhitespacePosition | 0.048893 | 0.041917 | -0.006976 | 0.8573 |
| src/compiler/program.ts:5098:1 | filterSemanticDiagnostics | 0.049741 | 0.042764 | -0.006977 | 0.8597 |
| src/compiler/visitorPublic.ts:395:1 | visitParameterList | 0.049554 | 0.042568 | -0.006986 | 0.8590 |
| src/compiler/moduleNameResolver.ts:3408:1 | SearchResult | 0.045275 | 0.038287 | -0.006989 | 0.8456 |
| src/compiler/utilities.ts:12434:1 | canHaveStatements | 0.245956 | 0.238950 | -0.007005 | 0.9715 |
| src/compiler/factory/utilities.ts:1159:1 | isReadonlyKeywordOrPlusOrMinusToken | 0.277852 | 0.270814 | -0.007039 | 0.9747 |
| src/compiler/factory/emitHelpers.ts:72:1 | ESDecorateClassElementContext | 0.049366 | 0.042267 | -0.007099 | 0.8562 |
| src/compiler/utilities.ts:8099:1 | accessKind | 0.090659 | 0.083487 | -0.007172 | 0.9209 |
| src/compiler/moduleNameResolver.ts:755:1 | getConditions | 0.048904 | 0.041716 | -0.007188 | 0.8530 |
| src/compiler/utilitiesPublic.ts:2643:1 | hasInternalAnnotation | 0.055680 | 0.048479 | -0.007202 | 0.8707 |
| src/compiler/builder.ts:751:1 | getNextAffectedFilePendingEmit | 0.050169 | 0.042964 | -0.007205 | 0.8564 |
| src/compiler/utilitiesPublic.ts:1116:1 | getJSDocImplementsTags | 0.296874 | 0.289666 | -0.007208 | 0.9757 |
| src/compiler/types.ts:8364:1 | SnippetKind | 0.046944 | 0.039734 | -0.007210 | 0.8464 |
| src/compiler/commandLineParser.ts:2317:1 | getWatchOptionsNameMap | 0.049133 | 0.041909 | -0.007224 | 0.8530 |
| src/compiler/emitter.ts:695:1 | getAllProjectOutputs | 0.129420 | 0.122184 | -0.007237 | 0.9441 |
| src/compiler/utilities.ts:3420:1 | isThisInitializedDeclaration | 0.310119 | 0.302857 | -0.007263 | 0.9766 |
| src/compiler/checker.ts:1455:1 | NodeLinks | 0.063473 | 0.056195 | -0.007279 | 0.8853 |
| src/compiler/transformers/utilities.ts:571:1 | findSuperStatementIndexPath | 0.046597 | 0.039305 | -0.007293 | 0.8435 |
| src/compiler/transformers/taggedTemplate.ts:101:1 | createTemplateCooked | 4.524681 | 4.517358 | -0.007322 | 0.9984 |
| src/compiler/types.ts:6993:1 | JsxReferenceKind | 0.046760 | 0.039403 | -0.007357 | 0.8427 |
| src/compiler/factory/emitHelpers.ts:970:1 |  | 0.089509 | 0.082136 | -0.007373 | 0.9176 |
| src/compiler/types.ts:2039:1 | ObjectBindingPattern | 0.048866 | 0.041481 | -0.007386 | 0.8489 |
| src/compiler/types.ts:6938:1 | ConditionalRoot | 0.046329 | 0.038902 | -0.007427 | 0.8397 |
| src/compiler/types.ts:1964:1 | PrivateIdentifierSetAccessorDeclaration | 0.050814 | 0.043319 | -0.007495 | 0.8525 |
| src/compiler/types.ts:4139:1 | JSDocImportTag | 0.053283 | 0.045763 | -0.007520 | 0.8589 |
| src/compiler/factory/utilities.ts:1565:1 | formatIdentifier | 0.050912 | 0.043392 | -0.007520 | 0.8523 |
| src/compiler/types.ts:2074:1 | FunctionLikeDeclaration | 0.053813 | 0.046282 | -0.007531 | 0.8601 |
| src/compiler/commandLineParser.ts:2369:1 |  | 0.048559 | 0.041017 | -0.007542 | 0.8447 |
| src/compiler/utilities.ts:1295:1 | isExportNamespaceAsDefaultDeclaration | 0.304292 | 0.296750 | -0.007542 | 0.9752 |
| src/compiler/types.ts:1641:1 | ProtectedKeyword | 0.057738 | 0.050193 | -0.007544 | 0.8693 |
| src/compiler/program.ts:1038:1 | getModuleResolutionName | 0.049784 | 0.042218 | -0.007567 | 0.8480 |
| src/compiler/commandLineParser.ts:1851:1 |  | 0.077794 | 0.070202 | -0.007593 | 0.9024 |
| src/compiler/utilities.ts:2148:1 | getNonAugmentationDeclaration | 0.048637 | 0.041028 | -0.007609 | 0.8436 |
| src/compiler/types.ts:6924:1 | IndexFlags | 0.047806 | 0.040189 | -0.007617 | 0.8407 |
| src/compiler/factory/nodeFactory.ts:478:1 |  | 0.049101 | 0.041452 | -0.007649 | 0.8442 |
| src/compiler/semver.ts:261:1 |  | 0.047727 | 0.040068 | -0.007659 | 0.8395 |
| src/compiler/commandLineParser.ts:3766:1 | getDefaultTypeAcquisition | 0.049991 | 0.042330 | -0.007661 | 0.8467 |
| src/compiler/binder.ts:422:1 | getModuleInstanceStateForAliasTarget | 0.268865 | 0.261199 | -0.007666 | 0.9715 |
| src/compiler/utilitiesPublic.ts:925:1 | getDeclarationIdentifier | 0.199292 | 0.191612 | -0.007680 | 0.9615 |
| src/compiler/semver.ts:1:1 |  | 0.048052 | 0.040330 | -0.007721 | 0.8393 |
| src/compiler/sys.ts:121:1 | createPollingIntervalBasedLevels | 0.048277 | 0.040553 | -0.007723 | 0.8400 |
| src/compiler/utilities.ts:9334:1 | hasJsonModuleEmitEnabled | 0.063343 | 0.055590 | -0.007753 | 0.8776 |
| src/compiler/checker.ts:1324:1 | TypeSystemEntity | 0.050869 | 0.043087 | -0.007782 | 0.8470 |
| src/compiler/types.ts:1615:1 | EqualsGreaterThanToken | 0.062646 | 0.054860 | -0.007786 | 0.8757 |
| src/compiler/factory/nodeTests.ts:921:1 | isExternalModuleReference | 0.056789 | 0.048994 | -0.007795 | 0.8627 |
| src/compiler/utilities.ts:9302:1 |  | 0.107452 | 0.099653 | -0.007799 | 0.9274 |
| src/compiler/types.ts:10618:1 | HasInferredType | 0.047644 | 0.039819 | -0.007824 | 0.8358 |
| src/compiler/sourcemap.ts:31:1 | SourceMapGeneratorOptions | 0.048294 | 0.040468 | -0.007826 | 0.8379 |
| src/compiler/scanner.ts:300:1 |  | 0.048724 | 0.040874 | -0.007850 | 0.8389 |
| src/compiler/core.ts:1769:1 | isString | 0.047486 | 0.039590 | -0.007896 | 0.8337 |
| src/compiler/utilities.ts:11028:1 | isTypeDeclaration | 0.188061 | 0.180138 | -0.007922 | 0.9579 |
| src/compiler/factory/emitHelpers.ts:819:1 |  | 0.079512 | 0.071553 | -0.007959 | 0.8999 |
| src/compiler/_namespaces/ts.ts:12:1 |  | 0.047410 | 0.039442 | -0.007968 | 0.8319 |
| src/compiler/utilities.ts:12228:1 | getLibNameFromLibReference | 0.048632 | 0.040657 | -0.007975 | 0.8360 |
| src/compiler/factory/emitHelpers.ts:101:1 | EmitHelperFactory | 0.049863 | 0.041883 | -0.007980 | 0.8400 |
| src/compiler/checker.ts:1155:1 |  | 0.049023 | 0.041031 | -0.007992 | 0.8370 |
| src/compiler/types.ts:3587:1 | HeritageClause | 0.048356 | 0.040360 | -0.007996 | 0.8346 |
| src/compiler/utilities.ts:9300:1 |  | 0.103513 | 0.095505 | -0.008009 | 0.9226 |
| src/compiler/utilities.ts:11135:1 | getTextOfJsxAttributeName | 0.207025 | 0.199002 | -0.008023 | 0.9612 |
| src/compiler/visitorPublic.ts:513:1 | visitFunctionBody | 0.053784 | 0.045755 | -0.008029 | 0.8507 |
| src/compiler/factory/nodeTests.ts:977:1 | isCaseClause | 0.056814 | 0.048714 | -0.008100 | 0.8574 |
| src/compiler/types.ts:1868:1 | ConstructSignatureDeclaration | 0.057995 | 0.049881 | -0.008114 | 0.8601 |
| src/compiler/types.ts:6516:1 | EnumType | 0.047257 | 0.039139 | -0.008118 | 0.8282 |
| src/compiler/moduleNameResolver.ts:2997:1 | loadModuleFromNearestNodeModulesDirectory | 0.047007 | 0.038875 | -0.008132 | 0.8270 |
| src/compiler/sourcemap.ts:679:1 | compareSourcePositions | 0.046891 | 0.038732 | -0.008159 | 0.8260 |
| src/compiler/utilitiesPublic.ts:613:1 | ParameterPropertyDeclaration | 0.051218 | 0.043054 | -0.008164 | 0.8406 |
| src/compiler/factory/utilities.ts:204:1 | createReactNamespace | 0.093629 | 0.085455 | -0.008174 | 0.9127 |
| src/compiler/commandLineParser.ts:605:1 |  | 0.104032 | 0.095830 | -0.008202 | 0.9212 |
| src/compiler/utilities.ts:9306:1 |  | 0.104137 | 0.095895 | -0.008242 | 0.9209 |
| src/compiler/utilitiesPublic.ts:2442:1 | isStatementOrBlock | 0.060056 | 0.051802 | -0.008254 | 0.8626 |
| src/compiler/utilities.ts:7597:1 | isEmptyObjectLiteral | 0.177690 | 0.169415 | -0.008276 | 0.9534 |
| src/compiler/factory/utilities.ts:334:1 | createExpressionFromEntityName | 0.199250 | 0.190953 | -0.008297 | 0.9584 |
| src/compiler/factory/nodeTests.ts:377:1 | isSuperKeyword | 0.063118 | 0.054715 | -0.008402 | 0.8669 |
| src/compiler/utilities.ts:3473:1 | nodeCanBeDecorated | 0.048284 | 0.039876 | -0.008408 | 0.8259 |
| src/compiler/types.ts:7374:1 | PluginImport | 0.049431 | 0.040963 | -0.008468 | 0.8287 |
| src/compiler/utilities.ts:9326:1 |  | 0.106361 | 0.097858 | -0.008502 | 0.9201 |
| src/compiler/utilitiesPublic.ts:2090:1 | isExpressionKind | 0.063958 | 0.055453 | -0.008505 | 0.8670 |
| src/compiler/emitter.ts:1137:1 | getBuildInfoText | 0.050913 | 0.042367 | -0.008546 | 0.8321 |
| src/compiler/utilities.ts:6187:1 | getReplacement | 0.050939 | 0.042374 | -0.008565 | 0.8319 |
| src/compiler/factory/utilities.ts:1105:1 | canHaveIllegalType | 0.055732 | 0.047167 | -0.008565 | 0.8463 |
| src/compiler/types.ts:7241:1 | RepopulateModuleNotFoundDiagnosticChain | 0.049808 | 0.041202 | -0.008607 | 0.8272 |
| src/compiler/utilitiesPublic.ts:1445:1 | isNamedExportBindings | 0.054967 | 0.046360 | -0.008607 | 0.8434 |
| src/compiler/factory/nodeTests.ts:438:1 | isConstructorDeclaration | 0.055676 | 0.047068 | -0.008608 | 0.8454 |
| src/compiler/utilities.ts:8006:1 | isDeclarationNameOfEnumOrNamespace | 0.064595 | 0.055949 | -0.008646 | 0.8662 |
| src/compiler/types.ts:3065:1 | LiteralLikeElementAccessExpression | 0.052668 | 0.044020 | -0.008648 | 0.8358 |
| src/compiler/factory/emitNode.ts:240:1 | setConstantValue | 0.062757 | 0.054092 | -0.008665 | 0.8619 |
| src/compiler/utilities.ts:5318:1 | getFunctionFlags | 0.081452 | 0.072775 | -0.008677 | 0.8935 |
| src/compiler/factory/nodeTests.ts:1046:1 | isJSDocLinkPlain | 0.058196 | 0.049502 | -0.008695 | 0.8506 |
| src/compiler/moduleSpecifiers.ts:172:1 | ModuleSpecifierPreferences | 0.049508 | 0.040796 | -0.008712 | 0.8240 |
| src/compiler/utilities.ts:7316:1 | selectEffectiveModifierFlags | 0.049401 | 0.040639 | -0.008762 | 0.8226 |
| src/compiler/moduleNameResolver.ts:1799:1 | nodeNextJsonConfigResolver | 0.055872 | 0.047102 | -0.008770 | 0.8430 |
| src/compiler/types.ts:2092:1 | MethodSignature | 0.052779 | 0.044002 | -0.008778 | 0.8337 |
| src/compiler/utilities.ts:9304:1 |  | 0.102430 | 0.093645 | -0.008786 | 0.9142 |
| src/compiler/builder.ts:2452:1 | createRedirectedBuilderProgram | 0.056312 | 0.047505 | -0.008807 | 0.8436 |
| src/compiler/types.ts:3489:1 | DefaultClause | 0.049843 | 0.041026 | -0.008817 | 0.8231 |
| src/compiler/utilities.ts:11147:1 | getEscapedTextOfJsxNamespacedName | 0.056727 | 0.047846 | -0.008881 | 0.8434 |
| src/compiler/utilitiesPublic.ts:352:1 | textSpanIsEmpty | 0.049980 | 0.041090 | -0.008890 | 0.8221 |
| src/compiler/transformers/esDecorators.ts:288:1 | LexicalEnvironmentStackEntry | 0.063231 | 0.054302 | -0.008929 | 0.8588 |
| src/compiler/_namespaces/ts.ts:14:1 |  | 0.049855 | 0.040921 | -0.008934 | 0.8208 |
| src/compiler/types.ts:3567:1 | ClassElement | 0.049904 | 0.040905 | -0.008999 | 0.8197 |
| src/compiler/types.ts:2246:1 | TypeReferenceNode | 0.051558 | 0.042515 | -0.009043 | 0.8246 |
| src/compiler/factory/emitHelpers.ts:744:1 |  | 0.081803 | 0.072699 | -0.009104 | 0.8887 |
| src/compiler/types.ts:2108:1 | MethodDeclaration | 0.051369 | 0.042228 | -0.009141 | 0.8220 |
| src/compiler/moduleNameResolver.ts:1157:1 | getOriginalOrResolvedModuleFileName | 0.050218 | 0.041070 | -0.009149 | 0.8178 |
| src/compiler/utilities.ts:2129:1 | isExternalModuleAugmentation | 0.287666 | 0.278455 | -0.009211 | 0.9680 |
| src/compiler/factory/emitHelpers.ts:838:1 |  | 0.078969 | 0.069725 | -0.009244 | 0.8829 |
| src/compiler/moduleNameResolver.ts:1555:1 | tryLoadModuleUsingOptionalResolutionSettings | 0.055686 | 0.046414 | -0.009272 | 0.8335 |
| src/compiler/utilities.ts:10409:1 | skipTypeCheckingIgnoringNoCheck | 0.052347 | 0.043055 | -0.009292 | 0.8225 |
| src/compiler/utilities.ts:2745:1 | isHoistedVariable | 0.294609 | 0.285279 | -0.009330 | 0.9683 |
| src/compiler/factory/emitNode.ts:289:1 | moveEmitHelpers | 0.099238 | 0.089832 | -0.009406 | 0.9052 |
| src/compiler/utilities.ts:4294:1 | isPrototypePropertyAssignment | 0.296898 | 0.287492 | -0.009406 | 0.9683 |
| src/compiler/transformers/esDecorators.ts:269:1 | ClassElementLexicalEnvironmentStackEntry | 0.050697 | 0.041276 | -0.009421 | 0.8142 |
| src/compiler/watchPublic.ts:277:1 | WatchCompilerHost | 0.047441 | 0.037987 | -0.009454 | 0.8007 |
| src/compiler/utilities.ts:6207:1 | escapeString | 0.051301 | 0.041836 | -0.009465 | 0.8155 |
| src/compiler/utilities.ts:8275:1 | getClassLikeDeclarationOfSymbol | 0.050052 | 0.040537 | -0.009515 | 0.8099 |
| src/compiler/utilities.ts:2468:1 | createDiagnosticForFileFromMessageChain | 0.050673 | 0.041133 | -0.009540 | 0.8117 |
| src/compiler/factory/emitNode.ts:368:1 | getIdentifierAutoGenerate | 0.054584 | 0.045038 | -0.009547 | 0.8251 |
| src/compiler/_namespaces/ts.ts:7:1 |  | 0.048366 | 0.038800 | -0.009565 | 0.8022 |
| src/compiler/binder.ts:509:1 | createBinder | 0.065359 | 0.055745 | -0.009614 | 0.8529 |
| src/compiler/checker.ts:54213:1 | isDeclarationNameOrImportPropertyName | 0.243093 | 0.233463 | -0.009630 | 0.9604 |
| src/compiler/moduleNameResolver.ts:2028:1 | pathContainsNodeModules | 0.050435 | 0.040787 | -0.009648 | 0.8087 |
| src/compiler/parser.ts:462:1 | isJSDocLikeText | 0.051997 | 0.042333 | -0.009664 | 0.8142 |
| src/compiler/types.ts:2670:1 | DestructuringAssignment | 0.049044 | 0.039373 | -0.009670 | 0.8028 |
| src/compiler/commandLineParser.ts:3381:1 | ParsedTsconfig | 0.053156 | 0.043472 | -0.009684 | 0.8178 |
| src/compiler/checker.ts:1429:1 | MinArgumentCountFlags | 0.048905 | 0.039213 | -0.009692 | 0.8018 |
| src/compiler/moduleSpecifiers.ts:694:1 | comparePathsByRedirectAndNumberOfDirectorySeparators | 0.049468 | 0.039773 | -0.009695 | 0.8040 |
| src/compiler/types.ts:3997:1 | JSDocAugmentsTag | 0.082364 | 0.072632 | -0.009731 | 0.8818 |
| src/compiler/utilities.ts:8580:1 | setObjectAllocator | 0.051153 | 0.041421 | -0.009732 | 0.8098 |
| src/compiler/utilities.ts:2287:1 | isAnyImportOrBareOrAccessedRequire | 0.296831 | 0.287070 | -0.009761 | 0.9671 |
| src/compiler/factory/nodeTests.ts:540:1 | isMappedTypeNode | 0.056086 | 0.046323 | -0.009763 | 0.8259 |
| src/compiler/utilitiesPublic.ts:2148:1 | isExternalModuleIndicator | 0.325451 | 0.315668 | -0.009784 | 0.9699 |
| src/compiler/types.ts:2251:1 | TypePredicateNode | 0.052251 | 0.042460 | -0.009790 | 0.8126 |
| src/compiler/moduleNameResolver.ts:852:1 | TypeReferenceDirectiveResolutionCache | 0.049531 | 0.039629 | -0.009902 | 0.8001 |
| src/compiler/utilities.ts:5597:1 | isPartOfParameterDeclaration | 0.068673 | 0.058729 | -0.009944 | 0.8552 |
| src/compiler/utilities.ts:7786:1 | readJsonOrUndefined | 0.156401 | 0.146441 | -0.009961 | 0.9363 |
| src/compiler/utilities.ts:2276:1 | isAnyImportSyntax | 0.054380 | 0.044399 | -0.009981 | 0.8165 |
| src/compiler/commandLineParser.ts:2224:1 | ParseConfigFileHost | 0.049997 | 0.040013 | -0.009984 | 0.8003 |
| src/compiler/utilities.ts:3479:1 | nodeCanBeDecorated | 0.050462 | 0.040477 | -0.009984 | 0.8021 |
| src/compiler/utilities.ts:1276:1 | getNonDecoratorTokenPosOfNode | 0.243565 | 0.233579 | -0.009986 | 0.9590 |
| src/compiler/executeCommandLine.ts:1146:1 | enableStatisticsAndTracing | 0.051319 | 0.041322 | -0.009997 | 0.8052 |
| src/compiler/utilities.ts:3549:1 | childIsDecorated | 0.176098 | 0.166099 | -0.009999 | 0.9432 |
| src/compiler/transformers/destructuring.ts:610:1 | makeObjectBindingPattern | 0.055785 | 0.045749 | -0.010036 | 0.8201 |
| src/compiler/types.ts:3481:1 | CaseClause | 0.051147 | 0.041109 | -0.010037 | 0.8038 |
| src/compiler/semver.ts:141:1 | tryParseComponents | 0.052871 | 0.042831 | -0.010040 | 0.8101 |
| src/compiler/utilities.ts:2292:1 | isAnyImportOrRequireStatement | 0.303355 | 0.293301 | -0.010055 | 0.9669 |
| src/compiler/scanner.ts:1:1 |  | 0.052098 | 0.042027 | -0.010070 | 0.8067 |
| src/compiler/factory/utilities.ts:309:1 | createForOfBindingStatement | 0.212521 | 0.202407 | -0.010114 | 0.9524 |
| src/compiler/visitorPublic.ts:393:1 | visitParameterList | 0.052179 | 0.042063 | -0.010116 | 0.8061 |
| src/compiler/utilitiesPublic.ts:636:1 | walkUpBindingElementsAndPatterns | 0.225303 | 0.215143 | -0.010159 | 0.9549 |
| src/compiler/factory/nodeTests.ts:508:1 | isUnionTypeNode | 0.053710 | 0.043532 | -0.010179 | 0.8105 |
| src/compiler/factory/nodeTests.ts:434:1 | isClassStaticBlockDeclaration | 0.055174 | 0.044953 | -0.010221 | 0.8147 |
| src/compiler/types.ts:4129:1 | JSDocSatisfiesTag | 0.051511 | 0.041248 | -0.010263 | 0.8008 |
| src/compiler/factory/nodeTests.ts:524:1 | isParenthesizedTypeNode | 0.055380 | 0.045097 | -0.010283 | 0.8143 |
| src/compiler/factory/utilities.ts:668:1 | walkUpOuterExpressions | 0.300981 | 0.290672 | -0.010309 | 0.9657 |
| src/compiler/checker.ts:1435:1 | IntrinsicTypeKind | 0.050238 | 0.039928 | -0.010309 | 0.7948 |
| src/compiler/factory/utilities.ts:579:1 | isUseStrictPrologue | 0.303281 | 0.292952 | -0.010329 | 0.9659 |
| src/compiler/transformers/declarations/diagnostics.ts:1:1 |  | 0.052155 | 0.041822 | -0.010333 | 0.8019 |
| src/compiler/scanner.ts:414:1 | tokenToString | 0.059184 | 0.048822 | -0.010362 | 0.8249 |
| src/compiler/tracing.ts:38:1 | tracingEnabled | 0.051839 | 0.041452 | -0.010388 | 0.7996 |
| src/compiler/types.ts:10079:1 | TextChangeRange | 0.053144 | 0.042734 | -0.010410 | 0.8041 |
| src/compiler/types.ts:3618:1 | ModuleName | 0.051912 | 0.041488 | -0.010424 | 0.7992 |
| src/compiler/binder.ts:3815:1 | getContainerFlags | 0.240100 | 0.229661 | -0.010439 | 0.9565 |
| src/compiler/utilities.ts:8024:1 | isInitializedVariable | 0.317144 | 0.306671 | -0.010472 | 0.9670 |
| src/compiler/utilities.ts:8966:1 | isFileModuleFromUsingJSXTag | 0.062645 | 0.052160 | -0.010486 | 0.8326 |
| src/compiler/types.ts:1919:1 | PropertySignature | 0.068261 | 0.057755 | -0.010506 | 0.8461 |
| src/compiler/utilities.ts:2648:1 | isEnumConst | 0.051277 | 0.040717 | -0.010559 | 0.7941 |
| src/compiler/resolutionCache.ts:348:1 | getDirectoryToWatchFailedLookupLocation | 0.052820 | 0.042249 | -0.010571 | 0.7999 |
| src/compiler/utilities.ts:8266:1 | isAbstractConstructorSymbol | 0.055476 | 0.044823 | -0.010653 | 0.8080 |
| src/compiler/utilitiesPublic.ts:2489:1 | isJsxOpeningLikeElement | 0.057264 | 0.046590 | -0.010675 | 0.8136 |
| src/compiler/transformers/utilities.ts:441:1 | IdentifierNameMultiMap | 0.284810 | 0.274129 | -0.010681 | 0.9625 |
| src/compiler/transformers/utilities.ts:389:1 | IdentifierNameMap | 0.281150 | 0.270449 | -0.010701 | 0.9619 |
| src/compiler/commandLineParser.ts:132:1 |  | 0.079196 | 0.068457 | -0.010739 | 0.8644 |
| src/compiler/commandLineParser.ts:1713:1 |  | 4.818281 | 4.807531 | -0.010750 | 0.9978 |
| src/compiler/watchPublic.ts:291:1 | WatchCompilerHostOfFilesAndCompilerOptions | 0.049684 | 0.038921 | -0.010763 | 0.7834 |
| src/compiler/sourcemap.ts:699:1 | createDocumentPositionMapper | 0.056522 | 0.045720 | -0.010802 | 0.8089 |
| src/compiler/checker.ts:1230:1 | TypeFacts | 0.052364 | 0.041562 | -0.010802 | 0.7937 |
| src/compiler/types.ts:1584:1 | MutableNodeArray | 0.061824 | 0.050958 | -0.010866 | 0.8242 |
| src/compiler/types.ts:1462:1 | IsControlFlowContainer | 0.051023 | 0.040155 | -0.010868 | 0.7870 |
| src/compiler/program.ts:5056:1 |  | 0.078265 | 0.067378 | -0.010887 | 0.8609 |
| src/compiler/commandLineParser.ts:1788:1 |  | 0.049658 | 0.038756 | -0.010902 | 0.7805 |
| src/compiler/moduleSpecifiers.ts:1463:1 | getDefaultResolutionModeForFile | 0.230349 | 0.219408 | -0.010942 | 0.9525 |
| src/compiler/builder.ts:700:1 | clearAffectedFilesPendingEmit | 0.052673 | 0.041687 | -0.010986 | 0.7914 |
| src/compiler/executeCommandLine.ts:1156:1 | isSolutionMarkOrMeasure | 0.056281 | 0.045248 | -0.011033 | 0.8040 |
| src/compiler/transformers/destructuring.ts:605:1 | makeArrayAssignmentPattern | 0.058266 | 0.047144 | -0.011122 | 0.8091 |
| src/compiler/utilities.ts:3027:1 | isCommonJsExportPropertyAssignment | 0.295282 | 0.284144 | -0.011138 | 0.9623 |
| src/compiler/utilities.ts:7525:1 | isDottedName | 0.171334 | 0.160181 | -0.011153 | 0.9349 |
| src/compiler/utilities.ts:3434:1 | getEntityNameFromTypeNode | 0.171795 | 0.160640 | -0.011155 | 0.9351 |
| src/compiler/types.ts:2662:1 | ObjectDestructuringAssignment | 0.051425 | 0.040268 | -0.011157 | 0.7830 |
| src/compiler/moduleNameResolver.ts:354:1 | PackageJson | 0.053552 | 0.042382 | -0.011170 | 0.7914 |
| src/compiler/types.ts:4520:1 | JsonObjectExpression | 0.053479 | 0.042238 | -0.011241 | 0.7898 |
| src/compiler/factory/nodeTests.ts:484:1 | isTypeLiteralNode | 0.054557 | 0.043302 | -0.011256 | 0.7937 |
| src/compiler/types.ts:3462:1 | WithStatement | 0.052430 | 0.041152 | -0.011278 | 0.7849 |
| src/compiler/moduleNameResolver.ts:2564:1 | parsePackageName | 0.051929 | 0.040634 | -0.011295 | 0.7825 |
| src/compiler/utilities.ts:11140:1 | isJsxAttributeName | 0.058145 | 0.046839 | -0.011306 | 0.8056 |
| src/compiler/moduleNameResolver.ts:1:1 |  | 0.052140 | 0.040759 | -0.011381 | 0.7817 |
| src/compiler/emitter.ts:516:1 | getOutputPathsFor | 0.272829 | 0.261319 | -0.011510 | 0.9578 |
| src/compiler/utilities.ts:9561:1 | guessDirectorySymlink | 0.105950 | 0.094437 | -0.011513 | 0.8913 |
| src/compiler/path.ts:870:1 | changeFullExtension | 0.055087 | 0.043568 | -0.011519 | 0.7909 |
| src/compiler/emitter.ts:1142:1 | getBuildInfo | 0.052850 | 0.041253 | -0.011596 | 0.7806 |
| src/compiler/watchPublic.ts:307:1 | WatchCompilerHostOfConfigFile | 0.050639 | 0.038994 | -0.011644 | 0.7701 |
| src/compiler/moduleNameResolver.ts:2044:1 | parseNodeModuleFromPath | 0.053979 | 0.042320 | -0.011658 | 0.7840 |
| src/compiler/utilitiesPublic.ts:361:1 | textRangeContainsPositionInclusive | 0.053118 | 0.041366 | -0.011753 | 0.7787 |
| src/compiler/transformers/esDecorators.ts:276:1 | PropertyNameLexicalEnvironmentStackEntry | 0.053206 | 0.041447 | -0.011758 | 0.7790 |
| src/compiler/executeCommandLine.ts:149:1 | updateReportDiagnostic | 0.052078 | 0.040164 | -0.011914 | 0.7712 |
| src/compiler/utilitiesPublic.ts:2059:1 | isUnaryExpressionWithWrite | 0.212070 | 0.200138 | -0.011932 | 0.9437 |
| src/compiler/builder.ts:1074:1 | IncrementalBuildInfoFileId | 0.052898 | 0.040946 | -0.011952 | 0.7740 |
| src/compiler/types.ts:3602:1 | EnumMember | 0.052625 | 0.040668 | -0.011957 | 0.7728 |
| src/compiler/factory/nodeTests.ts:520:1 | isInferTypeNode | 0.054709 | 0.042733 | -0.011976 | 0.7811 |
| src/compiler/visitorPublic.ts:328:1 | visitArrayWorker | 0.053530 | 0.041487 | -0.012043 | 0.7750 |
| src/compiler/utilities.ts:8474:1 | Symbol | 0.050303 | 0.038241 | -0.012062 | 0.7602 |
| src/compiler/factory/nodeTests.ts:981:1 | isDefaultClause | 0.058934 | 0.046863 | -0.012071 | 0.7952 |
| src/compiler/factory/nodeTests.ts:971:1 | isJsxNamespacedName | 0.063460 | 0.051344 | -0.012116 | 0.8091 |
| src/compiler/utilities.ts:8285:1 | isUMDExportSymbol | 0.290565 | 0.278444 | -0.012121 | 0.9583 |
| src/compiler/types.ts:3594:1 | TypeAliasDeclaration | 0.051980 | 0.039856 | -0.012123 | 0.7668 |
| src/compiler/utilities.ts:3537:1 | nodeOrChildIsDecorated | 0.078465 | 0.066330 | -0.012135 | 0.8453 |
| src/compiler/types.ts:3578:1 | InterfaceDeclaration | 0.052664 | 0.040460 | -0.012204 | 0.7683 |
| src/compiler/types.ts:10074:1 | TextSpan | 0.054911 | 0.042661 | -0.012249 | 0.7769 |
| src/compiler/types.ts:4135:1 | JSDocSatisfiesExpression | 0.052449 | 0.040181 | -0.012268 | 0.7661 |
| src/compiler/utilitiesPublic.ts:619:1 | isEmptyBindingPattern | 0.209076 | 0.196753 | -0.012323 | 0.9411 |
| src/compiler/builderPublic.ts:168:1 | createSemanticDiagnosticsBuilderProgram | 0.060470 | 0.048129 | -0.012341 | 0.7959 |
| src/compiler/transformers/classFields.ts:283:1 | PrivateIdentifierInstanceFieldInfo | 0.052824 | 0.040478 | -0.012346 | 0.7663 |
| src/compiler/factory/nodeTests.ts:536:1 | isIndexedAccessTypeNode | 0.056141 | 0.043770 | -0.012372 | 0.7796 |
| src/compiler/checker.ts:1370:1 | RecursionFlags | 0.053074 | 0.040700 | -0.012374 | 0.7669 |
| src/compiler/program.ts:4877:1 | getDefaultResolutionModeForFileWorker | 0.054345 | 0.041945 | -0.012400 | 0.7718 |
| src/compiler/utilitiesPublic.ts:1:1 |  | 0.054442 | 0.041993 | -0.012449 | 0.7713 |
| src/compiler/types.ts:23:1 | Path | 0.052047 | 0.039591 | -0.012456 | 0.7607 |
| src/compiler/types.ts:2646:1 | BinaryOperatorToken | 0.055136 | 0.042653 | -0.012483 | 0.7736 |
| src/compiler/types.ts:4514:1 | JsonMinusNumericLiteral | 0.053316 | 0.040817 | -0.012498 | 0.7656 |
| src/compiler/parser.ts:10714:1 | extractPragmas | 0.053982 | 0.041482 | -0.012500 | 0.7684 |
| src/compiler/parser.ts:10713:1 |  | 0.053896 | 0.041378 | -0.012519 | 0.7677 |
| src/compiler/factory/utilities.ts:348:1 | createExpressionForPropertyName | 0.209022 | 0.196481 | -0.012541 | 0.9400 |
| src/compiler/factory/nodeTests.ts:931:1 | isJsxSelfClosingElement | 0.059628 | 0.047042 | -0.012586 | 0.7889 |
| src/compiler/transformers/declarations/diagnostics.ts:101:1 | GetSymbolAccessibilityDiagnostic | 0.053657 | 0.041055 | -0.012602 | 0.7651 |
| src/compiler/commandLineParser.ts:1910:1 | OptionsBase | 0.052762 | 0.040146 | -0.012616 | 0.7609 |
| src/compiler/types.ts:2045:1 | ArrayBindingPattern | 0.055072 | 0.042440 | -0.012632 | 0.7706 |
| src/compiler/commandLineParser.ts:126:1 |  | 0.082092 | 0.069457 | -0.012635 | 0.8461 |
| src/compiler/visitorPublic.ts:382:1 | visitLexicalEnvironment | 0.053997 | 0.041294 | -0.012703 | 0.7647 |
| src/compiler/utilitiesPublic.ts:2175:1 | isModuleBody | 0.057115 | 0.044360 | -0.012755 | 0.7767 |
| src/compiler/factory/nodeTests.ts:999:1 | isShorthandPropertyAssignment | 0.057474 | 0.044689 | -0.012785 | 0.7775 |
| src/compiler/types.ts:2666:1 | ArrayDestructuringAssignment | 0.052961 | 0.040093 | -0.012868 | 0.7570 |
| src/compiler/utilities.ts:9979:1 |  | 0.061104 | 0.048219 | -0.012886 | 0.7891 |
| src/compiler/utilities.ts:4099:1 | isModuleIdentifier | 0.263573 | 0.250666 | -0.012907 | 0.9510 |
| src/compiler/builder.ts:1192:1 | isNonIncrementalBuildInfo | 0.053500 | 0.040573 | -0.012927 | 0.7584 |
| src/compiler/factory/nodeTests.ts:480:1 | isTypeQueryNode | 0.055347 | 0.042320 | -0.013028 | 0.7646 |
| src/compiler/utilities.ts:2197:1 | isAmbientPropertyDeclaration | 0.058782 | 0.045698 | -0.013084 | 0.7774 |
| src/compiler/commandLineParser.ts:1932:1 | createUnknownOptionError | 0.058842 | 0.045733 | -0.013108 | 0.7772 |
| src/compiler/utilities.ts:4029:1 | isDefaultedExpandoInitializer | 0.398882 | 0.385766 | -0.013116 | 0.9671 |
| src/compiler/factory/emitNode.ts:271:1 | removeEmitHelper | 0.057971 | 0.044830 | -0.013140 | 0.7733 |
| src/compiler/utilitiesPublic.ts:296:1 | isExternalModuleNameRelative | 0.053093 | 0.039865 | -0.013228 | 0.7509 |
| src/compiler/factory/nodeTests.ts:951:1 | isJsxClosingFragment | 0.057282 | 0.044044 | -0.013238 | 0.7689 |
| src/compiler/types.ts:1200:1 | HasJSDoc | 0.053868 | 0.040620 | -0.013248 | 0.7541 |
| src/compiler/utilitiesPublic.ts:356:1 | textSpanContainsPosition | 0.054215 | 0.040927 | -0.013288 | 0.7549 |
| src/compiler/commandLineParser.ts:3364:1 | updateErrorForNoInputFiles | 0.055986 | 0.042617 | -0.013369 | 0.7612 |
| src/compiler/utilities.ts:7180:1 | hasEffectiveModifiers | 0.072152 | 0.058768 | -0.013384 | 0.8145 |
| src/compiler/program.ts:884:1 | isExclusivelyTypeOnlyImportOrExport | 0.195271 | 0.181881 | -0.013390 | 0.9314 |
| src/compiler/factory/nodeTests.ts:1009:1 | isEnumMember | 0.059854 | 0.046422 | -0.013431 | 0.7756 |
| src/compiler/utilitiesPublic.ts:2406:1 | isDeclarationStatement | 0.062573 | 0.049082 | -0.013491 | 0.7844 |
| src/compiler/types.ts:3530:1 | DeclarationWithTypeParameters | 0.052998 | 0.039491 | -0.013507 | 0.7451 |
| src/compiler/utilities.ts:9820:1 | getRegexFromPattern | 0.052198 | 0.038681 | -0.013516 | 0.7411 |
| src/compiler/factory/nodeTests.ts:468:1 | isTypeReferenceNode | 0.056351 | 0.042829 | -0.013522 | 0.7600 |
| src/compiler/factory/nodeTests.ts:496:1 | isNamedTupleMember | 0.056610 | 0.043050 | -0.013560 | 0.7605 |
| src/compiler/factory/nodeTests.ts:532:1 | isTypeOperatorNode | 0.056698 | 0.043122 | -0.013576 | 0.7606 |
| src/compiler/transformers/es2015.ts:483:1 | createSpreadSegment | 0.100722 | 0.087141 | -0.013581 | 0.8652 |
| src/compiler/types.ts:6752:1 | UnionType | 0.052654 | 0.039072 | -0.013582 | 0.7420 |
| src/compiler/utilities.ts:11091:1 | isOptionalDeclaration | 0.058304 | 0.044719 | -0.013585 | 0.7670 |
| src/compiler/builder.ts:1220:1 | hasSyntaxOrGlobalErrors | 0.209509 | 0.195908 | -0.013602 | 0.9351 |
| src/compiler/types.ts:967:1 | LocalsContainer | 0.053975 | 0.040319 | -0.013656 | 0.7470 |
| src/compiler/utilitiesPublic.ts:2375:1 | isStatementKindButNotDeclarationKind | 0.057408 | 0.043730 | -0.013677 | 0.7618 |
| src/compiler/moduleSpecifiers.ts:569:1 | getInfo | 0.166788 | 0.153085 | -0.013703 | 0.9178 |
| src/compiler/transformers/es2015.ts:472:1 | SpreadSegmentKind | 0.054820 | 0.041104 | -0.013716 | 0.7498 |
| src/compiler/factory/nodeTests.ts:446:1 | isSetAccessorDeclaration | 0.057383 | 0.043661 | -0.013722 | 0.7609 |
| src/compiler/parser.ts:10800:1 | tagNamesAreEquivalent | 0.134881 | 0.121122 | -0.013759 | 0.8980 |
| src/compiler/moduleNameResolver.ts:1769:1 | bundlerModuleNameResolver | 0.056880 | 0.043058 | -0.013822 | 0.7570 |
| src/compiler/utilitiesPublic.ts:2196:1 | isNamedImportBindings | 0.063332 | 0.049490 | -0.013842 | 0.7814 |
| src/compiler/factory/nodeTests.ts:492:1 | isTupleTypeNode | 0.056356 | 0.042488 | -0.013868 | 0.7539 |
| src/compiler/types.ts:1017:1 | ForEachChildNodes | 0.054237 | 0.040369 | -0.013868 | 0.7443 |
| src/compiler/factory/nodeTests.ts:516:1 | isConditionalTypeNode | 0.057151 | 0.043263 | -0.013888 | 0.7570 |
| src/compiler/types.ts:26:1 | MatchingKeys | 0.053885 | 0.039980 | -0.013905 | 0.7419 |
| src/compiler/builder.ts:774:1 | getNextPendingEmitDiagnosticsFile | 0.055760 | 0.041847 | -0.013913 | 0.7505 |
| src/compiler/types.ts:3652:1 | JSDocNamespaceDeclaration | 0.060683 | 0.046739 | -0.013945 | 0.7702 |
| src/compiler/factory/nodeTests.ts:963:1 | isJsxSpreadAttribute | 0.057888 | 0.043930 | -0.013958 | 0.7589 |
| src/compiler/utilitiesPublic.ts:1076:1 | getJSDocTypeParameterTagsWorker | 0.311747 | 0.297783 | -0.013964 | 0.9552 |
| src/compiler/utilities.ts:8459:1 | AllocatedSignature | 0.053302 | 0.039286 | -0.014016 | 0.7370 |
| src/compiler/executeCommandLine.ts:126:1 | getCountKey | 0.095009 | 0.080960 | -0.014049 | 0.8521 |
| src/compiler/_namespaces/ts.ts:22:1 |  | 0.054175 | 0.040112 | -0.014063 | 0.7404 |
| src/compiler/utilitiesPublic.ts:1030:1 | getModifiers | 0.069145 | 0.055031 | -0.014113 | 0.7959 |
| src/compiler/types.ts:1061:1 | HasChildren | 0.054486 | 0.040334 | -0.014152 | 0.7403 |
| src/compiler/factory/nodeTests.ts:458:1 | isIndexSignatureDeclaration | 0.056716 | 0.042537 | -0.014179 | 0.7500 |
| src/compiler/utilities.ts:2152:1 | isCommonJSContainingModuleKind | 0.052956 | 0.038757 | -0.014199 | 0.7319 |
| src/compiler/utilitiesPublic.ts:2138:1 | hasScopeMarker | 0.075590 | 0.061372 | -0.014218 | 0.8119 |
| src/compiler/commandLineParser.ts:1958:1 | parseCommandLineWorker | 0.131553 | 0.117326 | -0.014227 | 0.8919 |
| src/compiler/utilities.ts:7216:1 | hasAbstractModifier | 0.064726 | 0.050494 | -0.014232 | 0.7801 |
| src/compiler/types.ts:973:1 | FlowContainer | 0.054772 | 0.040458 | -0.014314 | 0.7387 |
| src/compiler/types.ts:6730:1 | TupleTypeReference | 0.053428 | 0.039109 | -0.014319 | 0.7320 |
| src/compiler/program.ts:985:1 | getResolutionModeOverride | 0.055239 | 0.040856 | -0.014384 | 0.7396 |
| src/compiler/types.ts:3518:1 | CatchClause | 0.055286 | 0.040883 | -0.014403 | 0.7395 |
| src/compiler/types.ts:6734:1 | UnionOrIntersectionType | 0.053679 | 0.039271 | -0.014408 | 0.7316 |
| src/compiler/utilitiesPublic.ts:312:1 |  | 0.053670 | 0.039230 | -0.014439 | 0.7310 |
| src/compiler/commandLineParser.ts:2277:1 | parseConfigFileTextToJson | 0.054411 | 0.039953 | -0.014457 | 0.7343 |
| src/compiler/core.ts:1527:1 | MultiMap | 0.053832 | 0.039364 | -0.014468 | 0.7312 |
| src/compiler/checker.ts:1469:1 | getSymbolId | 0.056285 | 0.041809 | -0.014476 | 0.7428 |
| src/compiler/types.ts:851:1 | ModifierFlags | 0.053630 | 0.039099 | -0.014531 | 0.7291 |
| src/compiler/types.ts:3694:1 | ImportDeclaration | 0.060163 | 0.045628 | -0.014535 | 0.7584 |
| src/compiler/symbolWalker.ts:1:1 |  | 0.056480 | 0.041945 | -0.014535 | 0.7426 |
| src/compiler/factory/nodeTests.ts:1018:1 | isBundle | 0.061042 | 0.046408 | -0.014634 | 0.7603 |
| src/compiler/utilitiesPublic.ts:327:1 | getDefaultLibFileName | 0.054120 | 0.039485 | -0.014635 | 0.7296 |
| src/compiler/factory/emitNode.ts:317:1 | getSnippetElement | 0.058431 | 0.043783 | -0.014648 | 0.7493 |
| src/compiler/factory/nodeTests.ts:935:1 | isJsxOpeningElement | 0.061407 | 0.046759 | -0.014648 | 0.7615 |
| src/compiler/corePublic.ts:43:1 | Comparison | 0.053065 | 0.038392 | -0.014673 | 0.7235 |
| src/compiler/tsbuildPublic.ts:854:1 | BuildInvalidedProject | 0.055009 | 0.040307 | -0.014702 | 0.7327 |
| src/compiler/moduleNameResolver.ts:2577:1 | noKeyStartsWithDot | 0.070653 | 0.055945 | -0.014708 | 0.7918 |
| src/compiler/transformers/utilities.ts:536:1 | getSuperCallFromStatement | 0.274300 | 0.259450 | -0.014850 | 0.9459 |
| src/compiler/utilitiesPublic.ts:2154:1 | isForInOrOfStatement | 0.064025 | 0.049156 | -0.014869 | 0.7678 |
| src/compiler/factory/emitNode.ts:282:1 | getEmitHelpers | 0.059194 | 0.044291 | -0.014903 | 0.7482 |
| src/compiler/moduleNameResolver.ts:397:1 | readPackageJsonTypesFields | 0.057321 | 0.042357 | -0.014964 | 0.7389 |
| src/compiler/commandLineParser.ts:2231:1 | getParsedCommandLineOfConfigFile | 0.227954 | 0.212986 | -0.014968 | 0.9343 |
| src/compiler/utilities.ts:9318:1 |  | 0.113333 | 0.098313 | -0.015020 | 0.8675 |
| src/compiler/scanner.ts:331:1 |  | 0.058396 | 0.043358 | -0.015038 | 0.7425 |
| src/compiler/semver.ts:314:1 | parseHyphen | 0.058137 | 0.043055 | -0.015083 | 0.7406 |
| src/compiler/factory/emitHelpers.ts:934:1 |  | 0.091345 | 0.076257 | -0.015088 | 0.8348 |
| src/compiler/types.ts:7654:1 | LineAndCharacter | 0.056457 | 0.041361 | -0.015095 | 0.7326 |
| src/compiler/types.ts:2641:1 | LogicalOrCoalescingAssignmentOperator | 0.056620 | 0.041518 | -0.015101 | 0.7333 |
| src/compiler/factory/nodeTests.ts:488:1 | isArrayTypeNode | 0.058690 | 0.043577 | -0.015113 | 0.7425 |
| src/compiler/utilitiesPublic.ts:1036:1 | getJSDocParameterTagsWorker | 0.208254 | 0.193126 | -0.015128 | 0.9274 |
| src/compiler/types.ts:10030:1 | ResolvedModuleSpecifierInfo | 0.054627 | 0.039427 | -0.015201 | 0.7217 |
| src/compiler/utilities.ts:3841:1 | isRequireCall | 0.182880 | 0.167677 | -0.015203 | 0.9169 |
| src/compiler/factory/nodeTests.ts:504:1 | isRestTypeNode | 0.058655 | 0.043440 | -0.015215 | 0.7406 |
| src/compiler/utilities.ts:8462:1 | ObjectAllocator | 0.053305 | 0.038089 | -0.015216 | 0.7145 |
| src/compiler/moduleNameResolver.ts:1803:1 | nodeModuleNameResolverWorker | 0.061013 | 0.045786 | -0.015227 | 0.7504 |
| src/compiler/transformers/utilities.ts:485:1 | isSimpleInlineableExpression | 0.254321 | 0.239073 | -0.015249 | 0.9400 |
| src/compiler/factory/nodeTests.ts:1026:1 | isJSDocTypeExpression | 0.060725 | 0.045342 | -0.015383 | 0.7467 |
| src/compiler/tsbuildPublic.ts:849:1 | UpdateOutputFileStampsProject | 0.053543 | 0.038128 | -0.015415 | 0.7121 |
| src/compiler/tsbuildPublic.ts:1194:1 | InvalidateProjectCreateInfo | 0.055155 | 0.039693 | -0.015462 | 0.7197 |
| src/compiler/factory/nodeTests.ts:1014:1 | isSourceFile | 0.060427 | 0.044953 | -0.015474 | 0.7439 |
| src/compiler/utilities.ts:2453:1 | createFileDiagnosticFromMessageChain | 0.055846 | 0.040313 | -0.015533 | 0.7219 |
| src/compiler/parser.ts:9946:1 | IncrementalParser | 0.054633 | 0.039051 | -0.015582 | 0.7148 |
| src/compiler/types.ts:7826:1 | DidYouMeanOptionsDiagnostics | 0.060658 | 0.045048 | -0.015610 | 0.7427 |
| src/compiler/moduleNameResolver.ts:2421:1 | readPackageJsonPeerDependencies | 0.058067 | 0.042365 | -0.015702 | 0.7296 |
| src/compiler/utilities.ts:3789:1 | isInternalModuleImportEqualsDeclaration | 0.188427 | 0.172721 | -0.015706 | 0.9166 |
| src/compiler/factory/nodeTests.ts:1030:1 | isJSDocNameReference | 0.059423 | 0.043686 | -0.015737 | 0.7352 |
| src/compiler/types.ts:1294:1 | HasIllegalTypeParameters | 0.058281 | 0.042503 | -0.015778 | 0.7293 |
| src/compiler/checker.ts:54202:1 | isNotAccessor | 0.249342 | 0.233527 | -0.015815 | 0.9366 |
| src/compiler/sourcemap.ts:690:1 | getSourcePositionOfMapping | 0.054918 | 0.039010 | -0.015909 | 0.7103 |
| src/compiler/utilities.ts:2735:1 | isCustomPrologue | 0.066069 | 0.050106 | -0.015962 | 0.7584 |
| src/compiler/factory/nodeTests.ts:512:1 | isIntersectionTypeNode | 0.057954 | 0.041913 | -0.016041 | 0.7232 |
| src/compiler/factory/nodeTests.ts:464:1 | isTypePredicateNode | 0.059350 | 0.043299 | -0.016051 | 0.7295 |
| src/compiler/tsbuildPublic.ts:205:1 | getCurrentTime | 0.137504 | 0.121436 | -0.016067 | 0.8831 |
| src/compiler/builderPublic.ts:215:1 | createEmitAndSemanticDiagnosticsBuilderProgram | 0.057669 | 0.041592 | -0.016077 | 0.7212 |
| src/compiler/utilitiesPublic.ts:2483:1 | isStringLiteralOrJsxExpression | 0.060592 | 0.044502 | -0.016090 | 0.7345 |
| src/compiler/utilities.ts:11107:1 | isNonNullAccess | 0.068456 | 0.052358 | -0.016098 | 0.7648 |
| src/compiler/commandLineParser.ts:3265:1 | handleOptionConfigDirTemplateSubstitution | 4.562361 | 4.546252 | -0.016109 | 0.9965 |
| src/compiler/utilities.ts:7082:1 | writeCommentRange | 0.280140 | 0.264019 | -0.016122 | 0.9425 |
| src/compiler/utilities.ts:6814:1 | isThisInTypeQuery | 0.259589 | 0.243432 | -0.016158 | 0.9378 |
| src/compiler/program.ts:936:1 | getModeForUsageLocationWorker | 0.293130 | 0.276934 | -0.016196 | 0.9447 |
| src/compiler/commandLineParser.ts:2363:1 |  | 0.090922 | 0.074708 | -0.016214 | 0.8217 |
| src/compiler/types.ts:1339:1 | HasIllegalDecorators | 0.061844 | 0.045599 | -0.016245 | 0.7373 |
| src/compiler/factory/nodeTests.ts:544:1 | isLiteralTypeNode | 0.061131 | 0.044836 | -0.016294 | 0.7334 |
| src/compiler/utilities.ts:3142:1 | getContainingClassStaticBlock | 0.309588 | 0.293287 | -0.016301 | 0.9473 |
| src/compiler/tsbuildPublic.ts:884:1 | InvalidatedProject | 0.055433 | 0.039130 | -0.016303 | 0.7059 |
| src/compiler/commandLineParser.ts:1844:1 |  | 0.056718 | 0.040408 | -0.016310 | 0.7124 |
| src/compiler/tsbuildPublic.ts:374:1 | SolutionBuilderState | 0.056972 | 0.040658 | -0.016314 | 0.7137 |
| src/compiler/tsbuildPublic.ts:831:1 | InvalidatedProjectKind | 0.055734 | 0.039395 | -0.016339 | 0.7068 |
| src/compiler/types.ts:1:1 |  | 0.055813 | 0.039453 | -0.016360 | 0.7069 |
| src/compiler/utilitiesPublic.ts:2517:1 | isJSDocNode | 0.062285 | 0.045876 | -0.016409 | 0.7365 |
| src/compiler/utilities.ts:4113:1 | getAssignmentDeclarationKind | 0.068712 | 0.052293 | -0.016419 | 0.7610 |
| src/compiler/types.ts:940:1 | NodeId | 0.056158 | 0.039720 | -0.016438 | 0.7073 |
| src/compiler/types.ts:2637:1 | BinaryOperator | 0.058058 | 0.041606 | -0.016452 | 0.7166 |
| src/compiler/utilitiesPublic.ts:2360:1 | isDeclarationStatementKind | 0.059940 | 0.043333 | -0.016607 | 0.7229 |
| src/compiler/executeCommandLine.ts:810:1 |  | 0.058529 | 0.041904 | -0.016626 | 0.7159 |
| src/compiler/utilities.ts:11066:1 | canUsePropertyAccess | 0.058404 | 0.041762 | -0.016642 | 0.7151 |
| src/compiler/types.ts:33:1 | ReadonlyTextRange | 0.056387 | 0.039722 | -0.016666 | 0.7044 |
| src/compiler/transformers/es2017.ts:1059:1 | createSuperAccessVariableStatement | 0.461442 | 0.444582 | -0.016859 | 0.9635 |
| src/compiler/types.ts:3627:1 | AmbientModuleDeclaration | 0.058158 | 0.041245 | -0.016913 | 0.7092 |
| src/compiler/types.ts:3572:1 | TypeElement | 0.056388 | 0.039459 | -0.016929 | 0.6998 |
| src/compiler/utilities.ts:7563:1 | isPrototypeAccess | 0.182327 | 0.165365 | -0.016962 | 0.9070 |
| src/compiler/corePublic.ts:40:1 | Comparer | 0.055051 | 0.038006 | -0.017046 | 0.6904 |
| src/compiler/utilities.ts:8077:1 | getCombinedLocalAndExportSymbolFlags | 0.058172 | 0.041101 | -0.017071 | 0.7065 |
| src/compiler/emitter.ts:610:1 | getSingleOutputFileNames | 0.208748 | 0.191668 | -0.017081 | 0.9182 |
| src/compiler/types.ts:1313:1 | HasExpressionInitializer | 0.063725 | 0.046636 | -0.017088 | 0.7318 |
| src/compiler/types.ts:28:1 | TextRange | 0.056153 | 0.039058 | -0.017094 | 0.6956 |
| src/compiler/utilitiesPublic.ts:2203:1 | isModuleOrEnumDeclaration | 0.064669 | 0.047527 | -0.017142 | 0.7349 |
| src/compiler/types.ts:979:1 | HasFlowNode | 0.057919 | 0.040760 | -0.017158 | 0.7037 |
| src/compiler/types.ts:3611:1 | EnumDeclaration | 0.060788 | 0.043615 | -0.017173 | 0.7175 |
| src/compiler/tsbuildPublic.ts:569:1 | isParsedCommandLine | 0.057850 | 0.040571 | -0.017280 | 0.7013 |
| src/compiler/utilitiesPublic.ts:2321:1 | isDeclarationKind | 0.063059 | 0.045771 | -0.017287 | 0.7259 |
| src/compiler/utilities.ts:3068:1 | isFunctionBlock | 0.253254 | 0.235912 | -0.017343 | 0.9315 |
| src/compiler/types.ts:1563:1 | IsObjectLiteralOrClassExpressionMethodOrAccessor | 0.057543 | 0.040190 | -0.017353 | 0.6984 |
| src/compiler/factory/nodeTests.ts:476:1 | isConstructorTypeNode | 0.060177 | 0.042798 | -0.017379 | 0.7112 |
| src/compiler/utilitiesPublic.ts:389:1 | textSpanIntersectsWithTextSpan | 0.056357 | 0.038924 | -0.017433 | 0.6907 |
| src/compiler/factory/nodeTests.ts:454:1 | isConstructSignatureDeclaration | 0.060402 | 0.042953 | -0.017450 | 0.7111 |
| src/compiler/utilities.ts:2385:1 | getTextOfPropertyName | 0.064363 | 0.046837 | -0.017526 | 0.7277 |
| src/compiler/utilities.ts:10014:1 | getSupportedExtensionsWithJsonIfResolveJsonModule | 0.060214 | 0.042675 | -0.017538 | 0.7087 |
| src/compiler/utilities.ts:1142:1 | isAnyPrologueDirective | 0.310279 | 0.292739 | -0.017540 | 0.9435 |
| src/compiler/transformers/utilities.ts:612:1 | isInitializedOrStaticProperty | 0.287885 | 0.270299 | -0.017586 | 0.9389 |
| src/compiler/tsbuildPublic.ts:836:1 | InvalidatedProjectBase | 0.055903 | 0.038287 | -0.017616 | 0.6849 |
| src/compiler/types.ts:1556:1 | IsInterface | 0.057696 | 0.040057 | -0.017638 | 0.6943 |
| src/compiler/watchPublic.ts:106:1 | readBuilderProgram | 0.170715 | 0.152977 | -0.017738 | 0.8961 |
| src/compiler/utilities.ts:9976:1 |  | 0.060720 | 0.042846 | -0.017874 | 0.7056 |
| src/compiler/types.ts:1299:1 | HasTypeArguments | 0.064848 | 0.046902 | -0.017946 | 0.7233 |
| src/compiler/expressionToTypeNode.ts:160:1 | SyntacticResult | 0.058840 | 0.040829 | -0.018012 | 0.6939 |
| src/compiler/types.ts:1266:1 | HasType | 0.058119 | 0.040100 | -0.018019 | 0.6900 |
| src/compiler/utilitiesPublic.ts:2283:1 | canHaveLocals | 0.062615 | 0.044550 | -0.018065 | 0.7115 |
| src/compiler/commandLineParser.ts:1710:1 |  | 4.657335 | 4.639240 | -0.018095 | 0.9961 |
| src/compiler/utilitiesPublic.ts:2190:1 | isJSDocNamespaceBody | 0.062230 | 0.044067 | -0.018163 | 0.7081 |
| src/compiler/types.ts:582:1 | KeywordSyntaxKind | 0.061287 | 0.043080 | -0.018207 | 0.7029 |
| src/compiler/utilities.ts:7944:1 | getLinesBetweenRangeEndAndRangeStart | 0.060963 | 0.042578 | -0.018385 | 0.6984 |
| src/compiler/types.ts:1306:1 | HasInitializer | 0.061338 | 0.042900 | -0.018438 | 0.6994 |
| src/compiler/commandLineParser.ts:4118:1 | getWildcardDirectories | 0.144650 | 0.126177 | -0.018473 | 0.8723 |
| src/compiler/types.ts:932:1 | PredicateSemantics | 0.058261 | 0.039761 | -0.018500 | 0.6825 |
| src/compiler/utilities.ts:5684:1 | getExpressionPrecedence | 0.192157 | 0.173628 | -0.018529 | 0.9036 |
| src/compiler/utilitiesPublic.ts:2398:1 | isDeclaration | 0.074683 | 0.056093 | -0.018590 | 0.7511 |
| src/compiler/utilities.ts:5603:1 | getRootDeclaration | 0.068214 | 0.049618 | -0.018596 | 0.7274 |
| src/compiler/moduleSpecifiers.ts:581:1 | getLocalModuleSpecifier | 0.058405 | 0.039721 | -0.018684 | 0.6801 |
| src/compiler/sys.ts:1972:1 |  | 0.136381 | 0.117586 | -0.018795 | 0.8622 |
| src/compiler/types.ts:3506:1 | ThrowStatement | 0.065196 | 0.046299 | -0.018897 | 0.7101 |
| src/compiler/utilities.ts:4119:1 | isBindableObjectDefinePropertyCall | 0.068237 | 0.049303 | -0.018934 | 0.7225 |
| src/compiler/factory/nodeChildren.ts:14:1 | getNodeChildren | 0.176985 | 0.157969 | -0.019016 | 0.8926 |
| src/compiler/symbolWalker.ts:30:1 | createGetSymbolWalker | 0.171686 | 0.152669 | -0.019016 | 0.8892 |
| src/compiler/factory/nodeTests.ts:472:1 | isFunctionTypeNode | 0.062398 | 0.043369 | -0.019029 | 0.6950 |
| src/compiler/types.ts:1573:1 | HasContainerFlags | 0.059611 | 0.040582 | -0.019030 | 0.6808 |
| src/compiler/utilitiesPublic.ts:1380:1 | isOptionalChain | 0.064457 | 0.045405 | -0.019052 | 0.7044 |
| src/compiler/utilities.ts:9361:1 | StrictOptionName | 0.079087 | 0.059979 | -0.019108 | 0.7584 |
| src/compiler/types.ts:3622:1 | ModuleBody | 0.059411 | 0.040295 | -0.019116 | 0.6782 |
| src/compiler/moduleNameResolver.ts:1374:1 | createTypeReferenceDirectiveResolutionCache | 0.264798 | 0.245666 | -0.019132 | 0.9277 |
| src/compiler/factory/emitHelpers.ts:767:1 |  | 0.097039 | 0.077836 | -0.019203 | 0.8021 |
| src/compiler/types.ts:668:1 | ModifierSyntaxKind | 0.057658 | 0.038427 | -0.019230 | 0.6665 |
| src/compiler/utilities.ts:9589:1 | tryRemoveDirectoryPrefix | 0.086035 | 0.066701 | -0.019335 | 0.7753 |
| src/compiler/utilitiesPublic.ts:1121:1 | getJSDocClassTag | 0.079360 | 0.060017 | -0.019344 | 0.7563 |
| src/compiler/_namespaces/ts.ts:65:1 |  | 0.058175 | 0.038749 | -0.019427 | 0.6661 |
| src/compiler/watch.ts:229:1 | getErrorCountForSummary | 0.062837 | 0.043356 | -0.019481 | 0.6900 |
| src/compiler/types.ts:963:1 | JSDocArray | 0.060085 | 0.040590 | -0.019494 | 0.6756 |
| src/compiler/checker.ts:1313:1 |  | 0.089792 | 0.070264 | -0.019528 | 0.7825 |
| src/compiler/commandLineParser.ts:2370:1 | getTsconfigRootOptionsMap | 0.061568 | 0.042007 | -0.019561 | 0.6823 |
| src/compiler/types.ts:494:1 | TriviaSyntaxKind | 0.058870 | 0.039248 | -0.019622 | 0.6667 |
| src/compiler/sourcemap.ts:694:1 | getGeneratedPositionOfMapping | 0.058405 | 0.038775 | -0.019630 | 0.6639 |
| src/compiler/types.ts:502:1 | LiteralSyntaxKind | 0.059009 | 0.039352 | -0.019657 | 0.6669 |
| src/compiler/utilitiesPublic.ts:2208:1 | canHaveSymbol | 0.065197 | 0.045536 | -0.019660 | 0.6984 |
| src/compiler/tsbuildPublic.ts:1371:1 | isFileWatcherWithModifiedTime | 0.057442 | 0.037740 | -0.019702 | 0.6570 |
| src/compiler/factory/nodeTests.ts:500:1 | isOptionalTypeNode | 0.063000 | 0.043275 | -0.019725 | 0.6869 |
| src/compiler/utilities.ts:10018:1 | getSupportedExtensionsWithJsonIfResolveJsonModule | 0.059818 | 0.040073 | -0.019745 | 0.6699 |
| src/compiler/moduleSpecifiers.ts:1169:1 | tryGetModuleNameFromRootDirs | 0.090813 | 0.071043 | -0.019771 | 0.7823 |
| src/compiler/factory/utilities.ts:1570:1 | formatIdentifierWorker | 0.290479 | 0.270699 | -0.019780 | 0.9319 |
| src/compiler/types.ts:7841:1 | CommandLineOptionOfListType | 0.062389 | 0.042577 | -0.019812 | 0.6824 |
| src/compiler/tsbuildPublic.ts:367:1 | BuildInfoCacheEntry | 0.059856 | 0.040021 | -0.019835 | 0.6686 |
| src/compiler/utilities.ts:11051:1 | canHaveExportModifier | 0.302780 | 0.282932 | -0.019849 | 0.9344 |
| src/compiler/types.ts:3709:1 | NamedExportBindings | 0.062037 | 0.042146 | -0.019891 | 0.6794 |
| src/compiler/utilities.ts:9895:1 | getBasePaths | 0.062164 | 0.042155 | -0.020008 | 0.6781 |
| src/compiler/utilitiesPublic.ts:1399:1 | isExpressionOfOptionalChainRoot | 0.306961 | 0.286898 | -0.020063 | 0.9346 |
| src/compiler/tsbuildPublic.ts:925:1 | BuildStep | 0.061439 | 0.041345 | -0.020094 | 0.6729 |
| src/compiler/commandLineParser.ts:3651:1 | getExtendsConfigPath | 0.162068 | 0.141969 | -0.020099 | 0.8760 |
| src/compiler/types.ts:917:1 | RelationComparisonResult | 0.059268 | 0.039027 | -0.020241 | 0.6585 |
| src/compiler/types.ts:3639:1 | NamespaceBody | 0.063898 | 0.043544 | -0.020355 | 0.6815 |
| src/compiler/types.ts:580:1 | PunctuationOrKeywordSyntaxKind | 0.061022 | 0.040602 | -0.020420 | 0.6654 |
| src/compiler/utilities.ts:9994:1 | getSupportedExtensions | 0.059648 | 0.039210 | -0.020438 | 0.6574 |
| src/compiler/utilitiesPublic.ts:2166:1 | isFunctionBody | 0.256084 | 0.235643 | -0.020442 | 0.9202 |
| src/compiler/types.ts:739:1 | TokenSyntaxKind | 0.059246 | 0.038799 | -0.020447 | 0.6549 |
| src/compiler/utilities.ts:9937:1 | ensureScriptKind | 0.061727 | 0.041261 | -0.020466 | 0.6684 |
| src/compiler/types.ts:40:1 | SyntaxKind | 0.059749 | 0.039218 | -0.020531 | 0.6564 |
| src/compiler/utilities.ts:3527:1 | nodeIsDecorated | 0.076786 | 0.056201 | -0.020586 | 0.7319 |
| src/compiler/types.ts:3525:1 | ObjectTypeDeclaration | 0.058753 | 0.038109 | -0.020645 | 0.6486 |
| src/compiler/factory/utilities.ts:617:1 | isJSDocTypeAssertion | 0.301310 | 0.280553 | -0.020757 | 0.9311 |
| src/compiler/types.ts:700:1 | TypeNodeSyntaxKind | 0.060693 | 0.039864 | -0.020830 | 0.6568 |
| src/compiler/types.ts:3475:1 | CaseBlock | 0.064652 | 0.043809 | -0.020843 | 0.6776 |
| src/compiler/types.ts:1322:1 | HasIllegalExpressionInitializer | 0.065046 | 0.044189 | -0.020857 | 0.6793 |
| src/compiler/factory/nodeTests.ts:959:1 | isJsxAttributes | 0.065040 | 0.044130 | -0.020910 | 0.6785 |
| src/compiler/utilities.ts:9989:1 |  | 0.060206 | 0.039242 | -0.020964 | 0.6518 |
| src/compiler/types.ts:685:1 | KeywordTypeSyntaxKind | 0.059056 | 0.038074 | -0.020982 | 0.6447 |
| src/compiler/types.ts:3705:1 | NamedImportBindings | 0.063815 | 0.042801 | -0.021014 | 0.6707 |
| src/compiler/utilities.ts:12439:1 | isPotentiallyExecutableNode | 0.274250 | 0.253088 | -0.021162 | 0.9228 |
| src/compiler/types.ts:749:1 | JsxTokenSyntaxKind | 0.060637 | 0.039407 | -0.021230 | 0.6499 |
| src/compiler/utilities.ts:10760:1 | isPackedArrayLiteral | 0.310097 | 0.288843 | -0.021253 | 0.9315 |
| src/compiler/utilities.ts:10030:1 | hasJSFileExtension | 0.062685 | 0.041292 | -0.021393 | 0.6587 |
| src/compiler/utilities.ts:9973:1 |  | 0.061397 | 0.039929 | -0.021468 | 0.6503 |
| src/compiler/types.ts:1519:1 | HasLocals | 0.060460 | 0.038961 | -0.021499 | 0.6444 |
| src/compiler/utilitiesPublic.ts:870:1 | symbolName | 0.300747 | 0.279240 | -0.021507 | 0.9285 |
| src/compiler/utilities.ts:6703:1 | writeFile | 0.142958 | 0.121423 | -0.021535 | 0.8494 |
| src/compiler/types.ts:758:1 | JSDocSyntaxKind | 0.061568 | 0.039983 | -0.021585 | 0.6494 |
| src/compiler/utilities.ts:10040:1 | hasImplementationTSFileExtension | 0.062584 | 0.040980 | -0.021604 | 0.6548 |
| src/compiler/types.ts:3511:1 | TryStatement | 0.062094 | 0.040465 | -0.021629 | 0.6517 |
| src/compiler/factory/nodeTests.ts:947:1 | isJsxOpeningFragment | 0.065799 | 0.044061 | -0.021738 | 0.6696 |
| src/compiler/tsbuildPublic.ts:348:1 | ConfigFileCacheEntry | 0.061426 | 0.039659 | -0.021768 | 0.6456 |
| src/compiler/moduleNameResolver.ts:1966:1 | normalizePathForCJSResolution | 0.063688 | 0.041858 | -0.021830 | 0.6572 |
| src/compiler/moduleNameResolver.ts:2555:1 | extensionIsOk | 0.063497 | 0.041614 | -0.021883 | 0.6554 |
| src/compiler/types.ts:905:1 | JsxFlags | 0.060291 | 0.038399 | -0.021892 | 0.6369 |
| src/compiler/utilities.ts:9982:1 |  | 0.060273 | 0.038310 | -0.021963 | 0.6356 |
| src/compiler/factory/nodeTests.ts:939:1 | isJsxClosingElement | 0.067229 | 0.045249 | -0.021981 | 0.6731 |
| src/compiler/factory/nodeTests.ts:442:1 | isGetAccessorDeclaration | 0.067328 | 0.045336 | -0.021991 | 0.6734 |
| src/compiler/types.ts:3468:1 | SwitchStatement | 0.063191 | 0.041131 | -0.022060 | 0.6509 |
| src/compiler/types.ts:3185:1 | ImportDeferProperty | 0.061559 | 0.039469 | -0.022091 | 0.6411 |
| src/compiler/_namespaces/ts.ts:66:1 |  | 0.060593 | 0.038481 | -0.022112 | 0.6351 |
| src/compiler/utilities.ts:7430:1 | isLogicalOrCoalescingBinaryExpression | 0.335272 | 0.313158 | -0.022114 | 0.9340 |
| src/compiler/types.ts:3643:1 | NamespaceDeclaration | 0.063988 | 0.041842 | -0.022146 | 0.6539 |
| src/compiler/types.ts:1327:1 | HasDecorators | 0.065651 | 0.043348 | -0.022303 | 0.6603 |
| src/compiler/path.ts:934:1 | comparePathsCaseInsensitive | 0.063348 | 0.040859 | -0.022489 | 0.6450 |
| src/compiler/transformers/esnext.ts:784:1 | getUsingKindOfVariableStatement | 0.066161 | 0.043597 | -0.022564 | 0.6590 |
| src/compiler/factory/nodeFactory.ts:463:1 |  | 0.062521 | 0.039906 | -0.022614 | 0.6383 |
| src/compiler/semver.ts:39:1 |  | 0.062109 | 0.039455 | -0.022654 | 0.6353 |
| src/compiler/types.ts:3648:1 | JSDocNamespaceBody | 0.062074 | 0.039333 | -0.022741 | 0.6336 |
| src/compiler/types.ts:3631:1 | ModuleDeclaration | 0.063342 | 0.040547 | -0.022795 | 0.6401 |
| src/compiler/utilities.ts:3773:1 | isExternalModuleImportEqualsDeclaration | 0.204459 | 0.181642 | -0.022817 | 0.8884 |
| src/compiler/types.ts:10626:1 | SyntacticTypeNodeBuilderContext | 0.062614 | 0.039756 | -0.022858 | 0.6349 |
| src/compiler/sourcemap.ts:686:1 | compareGeneratedPositions | 0.061927 | 0.039038 | -0.022889 | 0.6304 |
| src/compiler/utilities.ts:2364:1 | tryGetTextOfPropertyName | 0.301648 | 0.278670 | -0.022979 | 0.9238 |
| src/compiler/commandLineParser.ts:2136:1 | getOptionDeclarationFromName | 0.142924 | 0.119846 | -0.023078 | 0.8385 |
| src/compiler/commandLineParser.ts:2267:1 | readConfigFile | 0.092607 | 0.069362 | -0.023245 | 0.7490 |
| src/compiler/types.ts:1960:1 | PrivateIdentifierGetAccessorDeclaration | 0.066573 | 0.043147 | -0.023426 | 0.6481 |
| src/compiler/path.ts:945:1 | comparePaths | 0.064506 | 0.041033 | -0.023473 | 0.6361 |
| src/compiler/scanner.ts:344:1 |  | 0.072616 | 0.048991 | -0.023626 | 0.6747 |
| src/compiler/utilitiesPublic.ts:1757:1 | isClassInstanceProperty | 0.393656 | 0.369969 | -0.023687 | 0.9398 |
| src/compiler/sys.ts:1:1 |  | 0.064686 | 0.040859 | -0.023827 | 0.6317 |
| src/compiler/factory/emitNode.ts:379:1 | getIdentifierGeneratedImportReference | 0.065140 | 0.041205 | -0.023935 | 0.6326 |
| src/compiler/factory/emitHelpers.ts:1114:1 |  | 0.101135 | 0.077182 | -0.023953 | 0.7632 |
| src/compiler/utilitiesPublic.ts:1605:1 | isPrivateIdentifierPropertyAccessExpression | 0.268480 | 0.244516 | -0.023964 | 0.9107 |
| src/compiler/types.ts:782:1 | NodeFlags | 0.064021 | 0.040034 | -0.023987 | 0.6253 |
| src/compiler/factory/utilities.ts:1695:1 | isSyntheticParenthesizedExpression | 0.296703 | 0.272706 | -0.023997 | 0.9191 |
| src/compiler/commandLineParser.ts:1728:1 | isCommandLineOptionOfCustomType | 4.501070 | 4.476904 | -0.024165 | 0.9946 |
| src/compiler/builder.ts:1052:1 | getBinderAndCheckerDiagnosticsOfFile | 0.103723 | 0.079256 | -0.024466 | 0.7641 |
| src/compiler/types.ts:516:1 | PunctuationSyntaxKind | 0.063215 | 0.038621 | -0.024594 | 0.6109 |
| src/compiler/utilitiesPublic.ts:2415:1 | isStatementButNotDeclaration | 0.071667 | 0.047047 | -0.024620 | 0.6565 |
| src/compiler/scanner.ts:412:1 | tokenToString | 0.073284 | 0.048621 | -0.024664 | 0.6635 |
| src/compiler/types.ts:1487:1 | IsFunctionLike | 0.063702 | 0.038894 | -0.024809 | 0.6106 |
| src/compiler/types.ts:1399:1 | PrimitiveLiteral | 0.065208 | 0.040376 | -0.024832 | 0.6192 |
| src/compiler/emitter.ts:710:1 | getOutputFileNames | 0.144073 | 0.118982 | -0.025091 | 0.8258 |
| src/compiler/types.ts:3496:1 | CaseOrDefaultClause | 0.068098 | 0.042893 | -0.025206 | 0.6299 |
| src/compiler/sys.ts:81:1 | FileWatcherEventKind | 0.067461 | 0.041802 | -0.025659 | 0.6196 |
| src/compiler/moduleNameResolver.ts:2573:1 | allKeysStartWithDot | 0.069001 | 0.043251 | -0.025750 | 0.6268 |
| src/compiler/utilities.ts:4104:1 | isModuleExportsAccessExpression | 0.270628 | 0.244510 | -0.026118 | 0.9035 |
| src/compiler/utilities.ts:2390:1 | entityNameToString | 0.225744 | 0.199609 | -0.026134 | 0.8842 |
| src/compiler/watch.ts:220:1 | parseConfigFileWithSystem | 0.068261 | 0.042112 | -0.026149 | 0.6169 |
| src/compiler/utilities.ts:6892:1 | getEffectiveTypeAnnotationNode | 0.268334 | 0.242090 | -0.026245 | 0.9022 |
| src/compiler/tsbuildPublic.ts:349:1 | SolutionBuilderStateCache | 0.065113 | 0.038698 | -0.026415 | 0.5943 |
| src/compiler/builderPublic.ts:162:1 | createSemanticDiagnosticsBuilderProgram | 0.100120 | 0.073676 | -0.026444 | 0.7359 |
| src/compiler/moduleNameResolver.ts:2489:1 | loadNodeModuleFromDirectoryWorker | 0.068232 | 0.041699 | -0.026533 | 0.6111 |
| src/compiler/moduleSpecifiers.ts:335:1 | tryGetModuleSpecifiersFromCacheWorker | 0.185703 | 0.159157 | -0.026546 | 0.8571 |
| src/compiler/types.ts:7682:1 | ScriptTarget | 0.069703 | 0.043138 | -0.026565 | 0.6189 |
| src/compiler/types.ts:1448:1 | IsBlockScopedContainer | 0.065790 | 0.039178 | -0.026612 | 0.5955 |
| src/compiler/types.ts:3672:1 | ImportEqualsDeclaration | 0.076278 | 0.049614 | -0.026664 | 0.6504 |
| src/compiler/types.ts:1626:1 | AwaitKeyword | 0.083505 | 0.056758 | -0.026747 | 0.6797 |
| src/compiler/factory/nodeTests.ts:943:1 | isJsxFragment | 0.071259 | 0.044364 | -0.026895 | 0.6226 |
| src/compiler/types.ts:511:1 | PseudoLiteralSyntaxKind | 0.066987 | 0.040056 | -0.026931 | 0.5980 |
| src/compiler/executeCommandLine.ts:159:1 | defaultIsPretty | 0.226663 | 0.199676 | -0.026987 | 0.8809 |
| src/compiler/factory/emitNode.ts:208:1 | getSyntheticTrailingComments | 0.076332 | 0.049259 | -0.027073 | 0.6453 |
| src/compiler/types.ts:3500:1 | LabeledStatement | 0.067816 | 0.040646 | -0.027170 | 0.5994 |
| src/compiler/scanner.ts:419:1 | stringToToken | 0.072522 | 0.045069 | -0.027453 | 0.6215 |
| src/compiler/moduleNameResolver.ts:2694:1 | comparePatternKeys | 0.068579 | 0.040924 | -0.027654 | 0.5967 |
| src/compiler/_namespaces/ts.ts:21:1 |  | 0.068449 | 0.040742 | -0.027707 | 0.5952 |
| src/compiler/utilities.ts:9597:1 |  | 0.087378 | 0.059532 | -0.027846 | 0.6813 |
| src/compiler/factory/nodeTests.ts:562:1 | isObjectBindingPattern | 0.073007 | 0.044938 | -0.028069 | 0.6155 |
| src/compiler/factory/nodeTests.ts:1050:1 | isJSDocAllType | 0.074461 | 0.046285 | -0.028177 | 0.6216 |
| src/compiler/sys.ts:88:1 | DirectoryWatcherCallback | 0.070040 | 0.041857 | -0.028183 | 0.5976 |
| src/compiler/utilities.ts:5449:1 | getSymbolNameForPrivateIdentifier | 0.072640 | 0.044380 | -0.028259 | 0.6110 |
| src/compiler/expressionToTypeNode.ts:164:1 | syntacticResult | 0.073931 | 0.045565 | -0.028366 | 0.6163 |
| src/compiler/types.ts:1287:1 | HasIllegalType | 0.069479 | 0.040956 | -0.028523 | 0.5895 |
| src/compiler/utilities.ts:9829:1 | matchFiles | 0.145286 | 0.116721 | -0.028564 | 0.8034 |
| src/compiler/factory/nodeTests.ts:548:1 | isImportTypeNode | 0.075313 | 0.046541 | -0.028772 | 0.6180 |
| src/compiler/tsbuildPublic.ts:331:1 | getCompilerOptionsOfBuildOptions | 0.070278 | 0.041159 | -0.029119 | 0.5857 |
| src/compiler/types.ts:1509:1 | IsFunctionExpression | 0.068074 | 0.038714 | -0.029360 | 0.5687 |
| src/compiler/utilitiesPublic.ts:2118:1 | isIterationStatement | 0.167006 | 0.137625 | -0.029381 | 0.8241 |
| src/compiler/expressionToTypeNode.ts:166:1 | syntacticResult | 0.076207 | 0.046204 | -0.030003 | 0.6063 |
| src/compiler/factory/nodeTests.ts:955:1 | isJsxAttribute | 0.075054 | 0.044984 | -0.030070 | 0.5994 |
| src/compiler/builderPublic.ts:207:1 | createEmitAndSemanticDiagnosticsBuilderProgram | 0.074712 | 0.043927 | -0.030785 | 0.5880 |
| src/compiler/builder.ts:635:1 | getNextAffectedFile | 4.999835 | 4.968959 | -0.030876 | 0.9938 |
| src/compiler/factory/nodeTests.ts:1042:1 | isJSDocLinkCode | 0.078671 | 0.047416 | -0.031255 | 0.6027 |
| src/compiler/utilitiesPublic.ts:615:1 | isParameterPropertyDeclaration | 0.333930 | 0.302450 | -0.031480 | 0.9057 |
| src/compiler/utilities.ts:6509:1 | getExternalModuleNameFromDeclaration | 0.357825 | 0.326024 | -0.031801 | 0.9111 |
| src/compiler/moduleSpecifiers.ts:182:1 | getModuleSpecifierPreferences | 0.379354 | 0.347103 | -0.032251 | 0.9150 |
| src/compiler/utilities.ts:4657:1 | getJSDocCommentsAndTags | 0.271320 | 0.238598 | -0.032722 | 0.8794 |
| src/compiler/transformers/esnext.ts:1:1 |  | 0.076949 | 0.044196 | -0.032753 | 0.5744 |
| src/compiler/sys.ts:60:1 | generateDjb2Hash | 0.077039 | 0.044162 | -0.032877 | 0.5732 |
| src/compiler/types.ts:1361:1 | HasModifiers | 0.073583 | 0.040643 | -0.032940 | 0.5523 |
| src/compiler/types.ts:1391:1 | HasIllegalModifiers | 0.074043 | 0.040929 | -0.033114 | 0.5528 |
| src/compiler/factory/emitNode.ts:195:1 | getSyntheticLeadingComments | 0.085926 | 0.052775 | -0.033151 | 0.6142 |
| src/compiler/moduleSpecifiers.ts:679:1 | packageJsonPathsAreEqual | 0.074617 | 0.041207 | -0.033410 | 0.5522 |
| src/compiler/core.ts:2591:1 | isNodeLikeSystem | 0.114523 | 0.080823 | -0.033700 | 0.7057 |
| src/compiler/utilitiesPublic.ts:1416:1 | isOutermostOptionalChain | 0.314483 | 0.280425 | -0.034059 | 0.8917 |
| src/compiler/types.ts:1413:1 | IsContainer | 0.073703 | 0.039323 | -0.034380 | 0.5335 |
| src/compiler/scanner.ts:413:1 | tokenToString | 0.082082 | 0.047603 | -0.034479 | 0.5799 |
| src/compiler/moduleNameResolver.ts:509:1 |  | 0.075047 | 0.040328 | -0.034719 | 0.5374 |
| src/compiler/types.ts:3684:1 | ExternalModuleReference | 0.081658 | 0.046721 | -0.034936 | 0.5722 |
| src/compiler/path.ts:884:1 | comparePathsWorker | 0.080706 | 0.045512 | -0.035194 | 0.5639 |
| src/compiler/path.ts:1039:1 | getRelativePathFromDirectory | 0.088067 | 0.052694 | -0.035373 | 0.5983 |
| src/compiler/utilities.ts:4997:1 | isValueSignatureDeclaration | 0.255258 | 0.219810 | -0.035448 | 0.8611 |
| src/compiler/path.ts:963:1 | containsPath | 0.076594 | 0.041085 | -0.035510 | 0.5364 |
| src/compiler/sys.ts:134:1 | setCustomPollingValues | 0.136357 | 0.100611 | -0.035746 | 0.7379 |
| src/compiler/utilities.ts:7535:1 | isPropertyAccessEntityNameExpression | 0.303714 | 0.267805 | -0.035909 | 0.8818 |
| src/compiler/utilitiesPublic.ts:1591:1 | isFileLevelReservedGeneratedIdentifier | 0.075352 | 0.039431 | -0.035921 | 0.5233 |
| src/compiler/types.ts:7663:1 | ScriptKind | 0.079565 | 0.043439 | -0.036125 | 0.5460 |
| src/compiler/types.ts:942:1 | Node | 0.075685 | 0.038841 | -0.036844 | 0.5132 |
| src/compiler/semver.ts:336:1 | parseComparator | 0.233673 | 0.196683 | -0.036990 | 0.8417 |
| src/compiler/types.ts:1875:1 | VariableDeclaration | 0.093555 | 0.056456 | -0.037100 | 0.6034 |
| src/compiler/moduleNameResolver.ts:2445:1 | getPackageJsonInfo | 0.155775 | 0.118439 | -0.037336 | 0.7603 |
| src/compiler/types.ts:957:1 | JSDocContainer | 0.078334 | 0.040643 | -0.037691 | 0.5188 |
| src/compiler/programDiagnostics.ts:89:1 | createProgramDiagnostics | 0.144982 | 0.106911 | -0.038072 | 0.7374 |
| src/compiler/moduleNameResolver.ts:501:1 | getDefaultTypeRoots | 0.080912 | 0.041791 | -0.039120 | 0.5165 |
| src/compiler/moduleSpecifiers.ts:686:1 | countPathComponents | 0.079941 | 0.040518 | -0.039424 | 0.5068 |
| src/compiler/sys.ts:75:1 | setStackTraceLimit | 0.081119 | 0.041611 | -0.039508 | 0.5130 |
| src/compiler/commandLineParser.ts:1701:1 |  | 4.800687 | 4.761158 | -0.039528 | 0.9918 |
| src/compiler/types.ts:1948:1 | PrivateIdentifierPropertyDeclaration | 0.098007 | 0.058108 | -0.039899 | 0.5929 |
| src/compiler/factory/nodeTests.ts:878:1 | isNamedImports | 0.104005 | 0.064041 | -0.039964 | 0.6158 |
| src/compiler/utilities.ts:5454:1 | isKnownSymbol | 0.090589 | 0.050579 | -0.040011 | 0.5583 |
| src/compiler/factory/nodeTests.ts:552:1 | isTemplateLiteralTypeSpan | 0.086744 | 0.046357 | -0.040387 | 0.5344 |
| src/compiler/program.ts:666:1 | formatDiagnostic | 4.808590 | 4.767727 | -0.040863 | 0.9915 |
| src/compiler/types.ts:3742:1 | ImportAttribute | 0.083120 | 0.042206 | -0.040914 | 0.5078 |
| src/compiler/executeCommandLine.ts:800:1 | reportWatchModeWithoutSysSupport | 0.206942 | 0.165701 | -0.041241 | 0.8007 |
| src/compiler/moduleSpecifiers.ts:1140:1 | tryGetModuleNameFromPackageJsonImports | 0.176536 | 0.134703 | -0.041832 | 0.7630 |
| src/compiler/utilities.ts:6963:1 | emitComments | 4.739203 | 4.697232 | -0.041972 | 0.9911 |
| src/compiler/checker.ts:1443:1 |  | 0.107953 | 0.065814 | -0.042139 | 0.6097 |
| src/compiler/utilities.ts:7939:1 | rangeEndIsOnSameLineAsRangeStart | 0.088672 | 0.046143 | -0.042529 | 0.5204 |
| src/compiler/types.ts:3749:1 | ImportAttributes | 0.084492 | 0.041377 | -0.043116 | 0.4897 |
| src/compiler/path.ts:925:1 | comparePathsCaseSensitive | 0.091983 | 0.048761 | -0.043222 | 0.5301 |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts:3:1 |  | 0.479470 | 0.435511 | -0.043959 | 0.9083 |
| src/compiler/sys.ts:52:1 | clearTimeout | 0.090031 | 0.045666 | -0.044366 | 0.5072 |
| src/compiler/emitter.ts:685:1 | getCommonSourceDirectoryOfConfig | 0.163485 | 0.118844 | -0.044641 | 0.7269 |
| src/compiler/types.ts:3777:1 | ExportDeclaration | 0.087953 | 0.042823 | -0.045131 | 0.4869 |
| src/compiler/types.ts:7807:1 | CommandLineOptionOfBooleanType | 0.088852 | 0.042814 | -0.046038 | 0.4819 |
| src/compiler/utilities.ts:11184:1 | isExpandoPropertyDeclaration | 0.318341 | 0.272151 | -0.046190 | 0.8549 |
| src/compiler/sourcemap.ts:668:1 | isSourceMappedPosition | 4.721795 | 4.675581 | -0.046215 | 0.9902 |
| src/compiler/executeCommandLine.ts:1034:1 | createWatchOfFilesAndCompilerOptions | 0.432155 | 0.385499 | -0.046656 | 0.8920 |
| src/compiler/sys.ts:87:1 | FileWatcherCallback | 0.089487 | 0.042556 | -0.046931 | 0.4756 |
| src/compiler/utilities.ts:4148:1 | isBindableStaticAccessExpression | 0.288485 | 0.241534 | -0.046951 | 0.8372 |
| src/compiler/factory/nodeTests.ts:556:1 | isTemplateLiteralTypeNode | 0.091724 | 0.044391 | -0.047333 | 0.4840 |
| src/compiler/factory/nodeTests.ts:1038:1 | isJSDocLink | 0.092581 | 0.045238 | -0.047343 | 0.4886 |
| src/compiler/utilities.ts:7915:1 | rangeIsOnSingleLine | 0.113191 | 0.065110 | -0.048082 | 0.5752 |
| src/compiler/types.ts:7820:1 | AlternateModeDiagnostics | 0.094658 | 0.046447 | -0.048211 | 0.4907 |
| src/compiler/commandLineParser.ts:2176:1 | parseBuildCommand | 0.264269 | 0.215900 | -0.048368 | 0.8170 |
| src/compiler/types.ts:3769:1 | NamespaceExportDeclaration | 0.090592 | 0.041479 | -0.049113 | 0.4579 |
| src/compiler/watch.ts:554:1 | toFileName | 0.129212 | 0.079632 | -0.049580 | 0.6163 |
| src/compiler/types.ts:7813:1 | CommandLineOptionOfCustomType | 0.094784 | 0.044976 | -0.049807 | 0.4745 |
| src/compiler/utilities.ts:9503:1 | createSymlinkCache | 0.134892 | 0.084956 | -0.049937 | 0.6298 |
| src/compiler/types.ts:7834:1 | TsConfigOnlyOption | 0.092914 | 0.042934 | -0.049980 | 0.4621 |
| src/compiler/transformers/destructuring.ts:556:1 | createDestructuringPropertyAccess | 0.279304 | 0.229293 | -0.050011 | 0.8209 |
| src/compiler/utilities.ts:8940:1 | getDiagnosticMessage | 4.747320 | 4.697137 | -0.050183 | 0.9894 |
| src/compiler/types.ts:3763:1 | NamespaceExport | 0.092184 | 0.041656 | -0.050528 | 0.4519 |
| src/compiler/utilities.ts:4019:1 | getDefaultedExpandoInitializer | 0.347430 | 0.296600 | -0.050831 | 0.8537 |
| src/compiler/utilities.ts:4139:1 | isLiteralLikeElementAccess | 0.299949 | 0.249018 | -0.050931 | 0.8302 |
| src/compiler/utilitiesPublic.ts:1437:1 | isNonNullChain | 0.346404 | 0.295250 | -0.051155 | 0.8523 |
| src/compiler/factory/nodeTests.ts:882:1 | isImportSpecifier | 0.110993 | 0.059579 | -0.051414 | 0.5368 |
| src/compiler/utilities.ts:5459:1 | isPrivateIdentifierSymbol | 0.091574 | 0.039781 | -0.051793 | 0.4344 |
| src/compiler/types.ts:7765:1 | CommandLineOptionBase | 0.093815 | 0.041979 | -0.051836 | 0.4475 |
| src/compiler/path.ts:965:1 | containsPath | 0.092883 | 0.041021 | -0.051862 | 0.4416 |
| src/compiler/scanner.ts:395:1 | isUnicodeIdentifierPart | 0.094681 | 0.042687 | -0.051995 | 0.4508 |
| src/compiler/moduleSpecifiers.ts:1122:1 | tryGetModuleNameFromExports | 0.092136 | 0.039848 | -0.052288 | 0.4325 |
| src/compiler/utilities.ts:4658:1 | getJSDocCommentsAndTags | 0.288858 | 0.236343 | -0.052515 | 0.8182 |
| src/compiler/utilitiesPublic.ts:1765:1 | isMethodOrAccessor | 0.104538 | 0.051631 | -0.052906 | 0.4939 |
| src/compiler/tsbuildPublic.ts:359:1 | FileWatcherWithModifiedTime | 0.092446 | 0.039190 | -0.053256 | 0.4239 |
| src/compiler/types.ts:3740:1 | ImportAttributeName | 0.095169 | 0.041653 | -0.053515 | 0.4377 |
| src/compiler/core.ts:1951:1 | equateStringsCaseInsensitive | 0.093293 | 0.039537 | -0.053756 | 0.4238 |
| src/compiler/transformers/typeSerializer.ts:86:1 | RuntimeTypeSerializerContext | 0.096150 | 0.041143 | -0.055006 | 0.4279 |
| src/compiler/types.ts:7795:1 | CommandLineOptionOfStringType | 0.096572 | 0.041120 | -0.055452 | 0.4258 |
| src/compiler/moduleSpecifiers.ts:1020:1 | tryGetModuleNameFromExportsOrImports | 0.142483 | 0.086934 | -0.055549 | 0.6101 |
| src/compiler/utilitiesPublic.ts:1376:1 | isCallChain | 0.332814 | 0.277120 | -0.055694 | 0.8327 |
| src/compiler/types.ts:1956:1 | PrivateIdentifierMethodDeclaration | 0.099090 | 0.043196 | -0.055894 | 0.4359 |
| src/compiler/types.ts:1872:1 | BindingName | 0.107779 | 0.051791 | -0.055988 | 0.4805 |
| src/compiler/utilities.ts:9438:1 | hasZeroOrOneAsteriskCharacter | 0.112290 | 0.055999 | -0.056291 | 0.4987 |
| src/compiler/utilitiesPublic.ts:2143:1 | needsScopeMarker | 0.385353 | 0.328959 | -0.056395 | 0.8537 |
| src/compiler/core.ts:1934:1 | AnyFunction | 0.095395 | 0.038249 | -0.057146 | 0.4010 |
| src/compiler/transformers/esnext.ts:788:1 | getUsingKind | 0.288375 | 0.230742 | -0.057633 | 0.8001 |
| src/compiler/utilities.ts:7473:1 | isAssignmentExpression | 0.368066 | 0.310195 | -0.057871 | 0.8428 |
| src/compiler/moduleSpecifiers.ts:582:1 | getLocalModuleSpecifier | 0.097479 | 0.039572 | -0.057908 | 0.4060 |
| src/compiler/utilities.ts:9455:1 | SymlinkedDirectory | 0.125514 | 0.067161 | -0.058353 | 0.5351 |
| src/compiler/types.ts:3657:1 | ModuleBlock | 0.112874 | 0.054349 | -0.058526 | 0.4815 |
| src/compiler/types.ts:7801:1 | CommandLineOptionOfNumberType | 0.099060 | 0.040279 | -0.058781 | 0.4066 |
| src/compiler/utilitiesPublic.ts:461:1 | textChangeRangeNewSpan | 0.100935 | 0.040238 | -0.060697 | 0.3987 |
| src/compiler/utilities.ts:5444:1 | getEscapedTextOfIdentifierOrLiteral | 0.360987 | 0.300033 | -0.060954 | 0.8311 |
| src/compiler/scanner.ts:389:1 | isUnicodeIdentifierStart | 0.106560 | 0.045461 | -0.061100 | 0.4266 |
| src/compiler/utilitiesPublic.ts:465:1 | textChangeRangeIsUnchanged | 0.113234 | 0.051997 | -0.061238 | 0.4592 |
| src/compiler/utilitiesPublic.ts:457:1 | createTextSpanFromBounds | 0.100697 | 0.039024 | -0.061673 | 0.3875 |
| src/compiler/utilitiesPublic.ts:2419:1 | isStatement | 0.317901 | 0.256053 | -0.061849 | 0.8054 |
| src/compiler/scanner.ts:358:1 | lookupInUnicodeMap | 0.102900 | 0.040925 | -0.061975 | 0.3977 |
| src/compiler/sys.ts:130:1 |  | 0.107064 | 0.045046 | -0.062018 | 0.4207 |
| src/compiler/utilitiesPublic.ts:403:1 | textSpanIntersectsWithPosition | 0.101297 | 0.039187 | -0.062110 | 0.3869 |
| src/compiler/utilitiesPublic.ts:1390:1 | isOptionalChainRoot | 0.337669 | 0.275349 | -0.062320 | 0.8154 |
| src/compiler/utilitiesPublic.ts:393:1 | textSpanIntersectsWith | 0.103268 | 0.040609 | -0.062659 | 0.3932 |
| src/compiler/sys.ts:51:1 | setTimeout | 0.108886 | 0.045824 | -0.063062 | 0.4208 |
| src/compiler/path.ts:967:1 | containsPath | 0.104600 | 0.041507 | -0.063093 | 0.3968 |
| src/compiler/utilities.ts:7934:1 | rangeStartIsOnSameLineAsRangeEnd | 0.105325 | 0.042118 | -0.063208 | 0.3999 |
| src/compiler/transformers/utilities.ts:618:1 | isStaticPropertyDeclaration | 0.331344 | 0.267909 | -0.063435 | 0.8086 |
| src/compiler/utilities.ts:5469:1 | isProtoSetter | 0.255698 | 0.191152 | -0.064546 | 0.7476 |
| src/compiler/moduleNameResolver.ts:2656:1 | loadModuleFromImports | 0.107464 | 0.042505 | -0.064959 | 0.3955 |
| src/compiler/sys.ts:132:1 |  | 0.112453 | 0.047474 | -0.064979 | 0.4222 |
| src/compiler/utilitiesPublic.ts:412:1 | textSpanIntersection | 0.104066 | 0.039083 | -0.064983 | 0.3756 |
| src/compiler/utilitiesPublic.ts:2133:1 | isScopeMarker | 0.398343 | 0.332523 | -0.065819 | 0.8348 |
| src/compiler/program.ts:779:1 | formatDiagnosticsWithColorAndContext | 4.809219 | 4.743392 | -0.065828 | 0.9863 |
| src/compiler/watch.ts:362:1 | explainIfFileIsRedirectAndImpliedFormat | 0.109570 | 0.043371 | -0.066199 | 0.3958 |
| src/compiler/builder.ts:1616:1 | getTextHandlingSourceMapForSignature | 4.690287 | 4.623527 | -0.066760 | 0.9858 |
| src/compiler/utilitiesPublic.ts:423:1 | normalizeSpans | 0.108342 | 0.041203 | -0.067138 | 0.3803 |
| src/compiler/utilitiesPublic.ts:1782:1 | isTypeElement | 0.125045 | 0.057797 | -0.067248 | 0.4622 |
| src/compiler/utilitiesPublic.ts:446:1 | createTextSpan | 0.108089 | 0.039683 | -0.068406 | 0.3671 |
| src/compiler/path.ts:1010:1 | getPathComponentsRelativeTo | 0.107238 | 0.038576 | -0.068662 | 0.3597 |
| src/compiler/path.ts:1002:1 | startsWithDirectory | 0.109593 | 0.040304 | -0.069289 | 0.3678 |
| src/compiler/utilitiesPublic.ts:408:1 | textRangeIntersectsWithTextSpan | 0.108897 | 0.038879 | -0.070018 | 0.3570 |
| src/compiler/transformers/esnext.ts:64:1 | UsingKind | 0.114124 | 0.042205 | -0.071919 | 0.3698 |
| src/compiler/types.ts:3719:1 | ImportClause | 0.113309 | 0.040783 | -0.072526 | 0.3599 |
| src/compiler/scanner.ts:349:1 |  | 0.113456 | 0.040692 | -0.072765 | 0.3587 |
| src/compiler/checker.ts:1460:1 | getNodeId | 4.997007 | 4.923204 | -0.073803 | 0.9852 |
| src/compiler/utilities.ts:8612:1 | createDetachedDiagnostic | 4.767820 | 4.693444 | -0.074376 | 0.9844 |
| src/compiler/moduleNameResolver.ts:147:1 | removeIgnoredPackageId | 4.789686 | 4.715044 | -0.074642 | 0.9844 |
| src/compiler/types.ts:1952:1 | PrivateIdentifierAutoAccessorPropertyDeclaration | 0.118967 | 0.043290 | -0.075678 | 0.3639 |
| src/compiler/executeCommandLine.ts:812:1 | performBuild | 0.429452 | 0.353564 | -0.075888 | 0.8233 |
| src/compiler/commandLineParser.ts:1698:1 |  | 4.889425 | 4.811746 | -0.077679 | 0.9841 |
| src/compiler/transformers/destructuring.ts:218:1 | bindingOrAssignmentElementContainsNonLiteralComputedName | 0.396689 | 0.318766 | -0.077922 | 0.8036 |
| src/compiler/commandLineParser.ts:3522:1 | getExtendsConfigPathOrArray | 0.199125 | 0.120835 | -0.078290 | 0.6068 |
| src/compiler/utilities.ts:7920:1 | rangeStartPositionsAreOnSameLine | 0.131032 | 0.051765 | -0.079267 | 0.3951 |
| src/compiler/watch.ts:446:1 | fileIncludeReasonToDiagnostics | 0.196106 | 0.116738 | -0.079368 | 0.5953 |
| src/compiler/utilitiesPublic.ts:487:1 | collapseTextChangeRangesAcrossMultipleVersions | 0.127016 | 0.047344 | -0.079672 | 0.3727 |
| src/compiler/transformers/typeSerializer.ts:1:1 |  | 0.122323 | 0.041941 | -0.080383 | 0.3429 |
| src/compiler/moduleSpecifiers.ts:301:1 | getModuleSpecifierWorker | 4.637588 | 4.556607 | -0.080982 | 0.9825 |
| src/compiler/moduleNameResolver.ts:3424:1 | useCaseSensitiveFileNames | 4.914466 | 4.832646 | -0.081820 | 0.9834 |
| src/compiler/scanner.ts:354:1 |  | 0.121797 | 0.039557 | -0.082240 | 0.3248 |
| src/compiler/utilities.ts:4132:1 | isLiteralLikeAccess | 0.333297 | 0.250920 | -0.082377 | 0.7528 |
| src/compiler/factory/utilities.ts:1725:1 | flattenCommaList | 0.138014 | 0.053389 | -0.084626 | 0.3868 |
| src/compiler/sys.ts:129:1 |  | 0.155658 | 0.070633 | -0.085025 | 0.4538 |
| src/compiler/commandLineParser.ts:2747:1 | getCustomTypeMapOfCommandLineOption | 4.905486 | 4.819920 | -0.085566 | 0.9826 |
| src/compiler/types.ts:3735:1 | AssertEntry | 0.127133 | 0.041299 | -0.085833 | 0.3249 |
| src/compiler/utilities.ts:3898:1 | isVariableDeclarationInitializedWithRequireHelper | 0.378830 | 0.292988 | -0.085843 | 0.7734 |
| src/compiler/tsbuildPublic.ts:287:1 | createBuilderStatusReporter | 4.668994 | 4.583075 | -0.085918 | 0.9816 |
| src/compiler/utilities.ts:6832:1 | getAllAccessorDeclarations | 0.167534 | 0.079877 | -0.087657 | 0.4768 |
| src/compiler/utilitiesPublic.ts:477:1 |  | 0.131969 | 0.043648 | -0.088321 | 0.3307 |
| src/compiler/utilities.ts:3905:1 | isRequireVariableStatement | 0.320171 | 0.231434 | -0.088737 | 0.7228 |
| src/compiler/factory/nodeTests.ts:1034:1 | isJSDocMemberName | 0.135569 | 0.044369 | -0.091200 | 0.3273 |
| src/compiler/sys.ts:49:1 |  | 0.140333 | 0.047476 | -0.092858 | 0.3383 |
| src/compiler/types.ts:3757:1 | NamespaceImport | 0.134705 | 0.040991 | -0.093714 | 0.3043 |
| src/compiler/utilitiesPublic.ts:397:1 | decodedTextSpanIntersectsWith | 0.132695 | 0.038833 | -0.093862 | 0.2926 |
| src/compiler/types.ts:7704:1 | LanguageVariant | 0.137937 | 0.043755 | -0.094182 | 0.3172 |
| src/compiler/types.ts:3729:1 | ImportPhaseModifierSyntaxKind | 0.138221 | 0.042289 | -0.095932 | 0.3060 |
| src/compiler/utilities.ts:3994:1 | getExpandoInitializer | 0.416912 | 0.319278 | -0.097634 | 0.7658 |
| src/compiler/utilitiesPublic.ts:469:1 | createTextChangeRange | 0.142987 | 0.043557 | -0.099431 | 0.3046 |
| src/compiler/utilitiesPublic.ts:2160:1 | isConciseBody | 0.373248 | 0.273638 | -0.099610 | 0.7331 |
| src/compiler/watch.ts:427:1 | getMatchedIncludeSpec | 0.191731 | 0.092028 | -0.099702 | 0.4800 |
| src/compiler/moduleSpecifiers.ts:583:1 | getLocalModuleSpecifier | 0.140442 | 0.040275 | -0.100167 | 0.2868 |
| src/compiler/watch.ts:260:1 | getWatchErrorSummaryDiagnosticMessage | 0.145155 | 0.042354 | -0.102801 | 0.2918 |
| src/compiler/utilities.ts:803:1 | projectReferenceIsEqualTo | 4.642218 | 4.538114 | -0.104104 | 0.9776 |
| src/compiler/utilities.ts:7929:1 | rangeEndPositionsAreOnSameLine | 0.152978 | 0.048201 | -0.104777 | 0.3151 |
| src/compiler/types.ts:7728:1 | ConfigFileSpecs | 0.148491 | 0.042727 | -0.105763 | 0.2877 |
| src/compiler/builder.ts:625:1 | assertSourceFileOkWithoutNextAffectedCall | 4.928790 | 4.822269 | -0.106522 | 0.9784 |
| src/compiler/factory/utilities.ts:786:1 | getLocalNameForExternalImport | 0.435134 | 0.328608 | -0.106526 | 0.7552 |
| src/compiler/utilities.ts:3880:1 | isBindingElementOfBareOrAccessedRequire | 0.343670 | 0.236724 | -0.106945 | 0.6888 |
| src/compiler/builder.ts:1196:1 | ensureHasErrorsForState | 5.053094 | 4.945458 | -0.107636 | 0.9787 |
| src/compiler/scanner.ts:356:1 |  | 0.150758 | 0.042273 | -0.108485 | 0.2804 |
| src/compiler/utilitiesPublic.ts:1586:1 | isGeneratedPrivateIdentifier | 0.432411 | 0.323858 | -0.108552 | 0.7490 |
| src/compiler/moduleSpecifiers.ts:433:1 | getLocalModuleSpecifierBetweenFileNames | 4.773230 | 4.664561 | -0.108669 | 0.9772 |
| src/compiler/types.ts:7722:1 | WatchDirectoryFlags | 0.152444 | 0.043541 | -0.108903 | 0.2856 |
| src/compiler/sys.ts:1980:1 |  | 4.756739 | 4.646821 | -0.109918 | 0.9769 |
| src/compiler/factory/emitHelpers.ts:148:1 | createEmitHelperFactory | 0.836141 | 0.721020 | -0.115120 | 0.8623 |
| src/compiler/watch.ts:234:1 | getFilesInErrorForSummary | 0.159410 | 0.041686 | -0.117724 | 0.2615 |
| src/compiler/utilities.ts:9469:1 | SymlinkCache | 0.197568 | 0.075811 | -0.121757 | 0.3837 |
| src/compiler/types.ts:3732:1 | AssertionKey | 0.162668 | 0.040911 | -0.121757 | 0.2515 |
| src/compiler/parser.ts:1386:1 | isExternalModule | 4.681678 | 4.559454 | -0.122224 | 0.9739 |
| src/compiler/transformers/typeSerializer.ts:79:1 | SerializedTypeNode | 0.166377 | 0.041491 | -0.124886 | 0.2494 |
| src/compiler/types.ts:1885:1 | InitializedVariableDeclaration | 0.179116 | 0.053737 | -0.125379 | 0.3000 |
| src/compiler/factory/utilities.ts:1701:1 | flattenCommaListWorker | 0.413958 | 0.288131 | -0.125828 | 0.6960 |
| src/compiler/utilities.ts:2638:1 | isExternalOrCommonJsModule | 4.733482 | 4.605514 | -0.127968 | 0.9730 |
| src/compiler/types.ts:7710:1 | ParsedCommandLine | 0.174435 | 0.043863 | -0.130573 | 0.2515 |
| src/compiler/utilitiesPublic.ts:603:1 | getTypeParameterOwner | 0.392250 | 0.259467 | -0.132783 | 0.6615 |
| src/compiler/moduleNameResolver.ts:224:1 | createResolvedModuleWithFailedLookupLocationsHandlingSymlink | 4.796622 | 4.659469 | -0.137153 | 0.9714 |
| src/compiler/commandLineParser.ts:1704:1 |  | 4.820020 | 4.682165 | -0.137855 | 0.9714 |
| src/compiler/transformers/typeSerializer.ts:73:1 | SerializedEntityName | 0.179249 | 0.041259 | -0.137990 | 0.2302 |
| src/compiler/transformers/destructuring.ts:197:1 | bindingOrAssignmentElementAssignsToName | 0.452355 | 0.312867 | -0.139487 | 0.6916 |
| src/compiler/utilities.ts:3885:1 | getModuleSpecifierOfBareOrAccessedRequire | 0.417974 | 0.278119 | -0.139856 | 0.6654 |
| src/compiler/types.ts:7752:1 | CreateProgramOptions | 0.186117 | 0.041067 | -0.145049 | 0.2207 |
| src/compiler/transformers/destructuring.ts:386:1 | flattenObjectBindingOrAssignmentPattern | 0.840399 | 0.695235 | -0.145164 | 0.8273 |
| src/compiler/factory/utilities.ts:1739:1 | containsObjectRestOrSpread | 0.385436 | 0.239667 | -0.145769 | 0.6218 |
| src/compiler/emitter.ts:540:1 | getOutputExtension | 4.825796 | 4.679876 | -0.145920 | 0.9698 |
| src/compiler/moduleNameResolver.ts:2414:1 | getPeerDependenciesOfPackageJsonInfo | 5.108828 | 4.961921 | -0.146906 | 0.9712 |
| src/compiler/watch.ts:266:1 | prettyPathForFileError | 0.190405 | 0.041753 | -0.148652 | 0.2193 |
| src/compiler/utilities.ts:8790:1 | compareDiagnosticsSkipRelatedInformation | 4.720507 | 4.568517 | -0.151990 | 0.9678 |
| src/compiler/types.ts:1916:1 | BindingElementGrandparent | 0.209257 | 0.055428 | -0.153830 | 0.2649 |
| src/compiler/types.ts:3738:1 | AssertClause | 0.197804 | 0.041728 | -0.156076 | 0.2110 |
| src/compiler/types.ts:7748:1 | ModuleImportResult | 0.208974 | 0.044951 | -0.164023 | 0.2151 |
| src/compiler/types.ts:1887:1 | VariableDeclarationList | 0.219164 | 0.053403 | -0.165761 | 0.2437 |
| src/compiler/types.ts:1894:1 | ParameterDeclaration | 0.228654 | 0.060031 | -0.168623 | 0.2625 |
| src/compiler/factory/emitNode.ts:166:1 | getStartsOnNewLine | 5.163161 | 4.991919 | -0.171241 | 0.9668 |
| src/compiler/types.ts:1906:1 | BindingElement | 0.232253 | 0.058412 | -0.173841 | 0.2515 |
| src/compiler/utilitiesPublic.ts:2170:1 | isForInitializer | 0.448278 | 0.272227 | -0.176051 | 0.6073 |
| src/compiler/transformers/esnext.ts:774:1 | isUsingVariableDeclarationList | 0.517369 | 0.331598 | -0.185771 | 0.6409 |
| src/compiler/factory/emitHelpers.ts:704:1 | compareEmitHelpers | 5.017112 | 4.817855 | -0.199257 | 0.9603 |
| src/compiler/types.ts:3663:1 | ModuleReference | 0.248123 | 0.043201 | -0.204922 | 0.1741 |
| src/compiler/watch.ts:416:1 | getMatchedFileSpec | 0.304985 | 0.090121 | -0.214864 | 0.2955 |
| src/compiler/program.ts:4855:1 | getImpliedNodeFormatForEmitWorker | 4.879838 | 4.661569 | -0.218269 | 0.9553 |
| src/compiler/utilities.ts:7910:1 | createTokenRange | 0.271442 | 0.052407 | -0.219035 | 0.1931 |
| src/compiler/factory/utilities.ts:188:1 | createMemberAccessForPropertyName | 0.431168 | 0.193565 | -0.237603 | 0.4489 |
| src/compiler/utilities.ts:7879:1 | moveRangePastDecorators | 0.694356 | 0.449665 | -0.244691 | 0.6476 |
| src/compiler/factory/emitNode.ts:233:1 | getConstantValue | 5.246261 | 4.976444 | -0.269817 | 0.9486 |
| src/compiler/utilities.ts:3968:1 | getAssignedExpandoInitializer | 0.655068 | 0.361870 | -0.293198 | 0.5524 |
| src/compiler/utilitiesPublic.ts:1600:1 | isPrivateIdentifierClassElementDeclaration | 0.568496 | 0.270170 | -0.298326 | 0.4752 |
| src/compiler/factory/utilities.ts:1092:1 | getJSDocTypeAliasName | 0.517319 | 0.214374 | -0.302944 | 0.4144 |
| src/compiler/utilitiesPublic.ts:1778:1 | isModifierLike | 0.736383 | 0.377503 | -0.358880 | 0.5126 |
| src/compiler/watch.ts:276:1 | getErrorSummaryText | 0.462569 | 0.081270 | -0.381299 | 0.1757 |
| src/compiler/watch.ts:304:1 | createTabularErrorsDisplay | 0.599861 | 0.085665 | -0.514196 | 0.1428 |
| src/compiler/program.ts:1162:1 | isReferencedFile | 5.263230 | 4.632342 | -0.630888 | 0.8801 |
| src/compiler/watch.ts:351:1 | explainFiles | 0.978818 | 0.203223 | -0.775595 | 0.2076 |
| src/compiler/utilities.ts:7891:1 | moveRangePastModifiers | 1.126705 | 0.294546 | -0.832159 | 0.2614 |
