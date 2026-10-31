Unit timing includes registration, lowering and panic recovery. Ratio = B/A.

{
  "A_total_seconds": 596.824203048,
  "B_total_seconds": 590.417948097,
  "A_cpu_seconds": 641.769465000003,
  "B_cpu_seconds": 638.837079000002,
  "units": 4017,
  "controls": {
    "units": 1215,
    "median_absolute_delta": 0.0026209109999999966,
    "p95_absolute_delta": 0.030523607,
    "median_ratio": 0.9931612801377202
  },
  "control_definition": "Imports, interfaces and type aliases with identical lowering findings: no checked-predicate body to exercise B. These include registration costs; one paired run cannot separate scheduling, GC and common registration overhead."
}

| Unit | Name | A seconds | B seconds | B-A seconds | B/A |
|---|---|---:|---:|---:|---:|
| src/compiler/utilities.ts:11077:1 | hasTabstop | 4.216782 | 5.237323 | +1.020541 | 1.2420 |
| src/compiler/utilitiesPublic.ts:1116:1 | getJSDocImplementsTags | 0.266974 | 0.796449 | +0.529476 | 2.9832 |
| src/compiler/watchUtilities.ts:572:1 | isIgnoredFileFromWildCardWatching | 0.250143 | 0.700465 | +0.450322 | 2.8003 |
| src/compiler/utilities.ts:4330:1 | canHaveModuleSpecifier | 4.322584 | 4.699771 | +0.377187 | 1.0873 |
| src/compiler/parser.ts:1386:1 | isExternalModule | 4.634950 | 4.951703 | +0.316753 | 1.0683 |
| src/compiler/factory/utilities.ts:1092:1 | getJSDocTypeAliasName | 0.177981 | 0.463056 | +0.285075 | 2.6017 |
| src/compiler/program.ts:666:1 | formatDiagnostic | 4.212632 | 4.483339 | +0.270707 | 1.0643 |
| src/compiler/utilities.ts:2292:1 | isAnyImportOrRequireStatement | 0.280983 | 0.538853 | +0.257871 | 1.9177 |
| src/compiler/utilitiesPublic.ts:2170:1 | isForInitializer | 0.258863 | 0.512278 | +0.253415 | 1.9790 |
| src/compiler/utilitiesPublic.ts:865:1 | identifierToKeywordKind | 0.081359 | 0.331706 | +0.250347 | 4.0771 |
| src/compiler/program.ts:1339:1 | getImpliedNodeFormatForFile | 4.460918 | 4.688869 | +0.227951 | 1.0511 |
| src/compiler/transformers/es2018.ts:154:1 | transformES2018 | 0.217747 | 0.437614 | +0.219867 | 2.0097 |
| src/compiler/utilities.ts:3033:1 | isValidESSymbolDeclaration | 0.276470 | 0.492621 | +0.216151 | 1.7818 |
| src/compiler/utilities.ts:8612:1 | createDetachedDiagnostic | 4.812937 | 5.017904 | +0.204967 | 1.0426 |
| src/compiler/commandLineParser.ts:2747:1 | getCustomTypeMapOfCommandLineOption | 4.670610 | 4.870390 | +0.199779 | 1.0428 |
| src/compiler/utilities.ts:6963:1 | emitComments | 4.463285 | 4.662654 | +0.199369 | 1.0447 |
| src/compiler/commandLineParser.ts:1704:1 |  | 4.736653 | 4.925461 | +0.188808 | 1.0399 |
| src/compiler/moduleNameResolver.ts:3242:1 | tryFindNonRelativeModuleNameInCache | 4.283827 | 4.471916 | +0.188090 | 1.0439 |
| src/compiler/expressionToTypeNode.ts:174:1 | createSyntacticTypeNodeBuilder | 0.191635 | 0.376343 | +0.184707 | 1.9638 |
| src/compiler/utilitiesPublic.ts:2166:1 | isFunctionBody | 0.259645 | 0.442232 | +0.182586 | 1.7032 |
| src/compiler/transformers/classFields.ts:3343:1 | isReservedPrivateName | 0.393108 | 0.561923 | +0.168815 | 1.4294 |
| src/compiler/utilities.ts:8685:1 | createFileDiagnostic | 4.385244 | 4.551091 | +0.165846 | 1.0378 |
| src/compiler/utilitiesPublic.ts:2148:1 | isExternalModuleIndicator | 0.308170 | 0.467681 | +0.159511 | 1.5176 |
| src/compiler/moduleNameResolver.ts:511:1 | arePathsEqual | 4.474790 | 4.633722 | +0.158931 | 1.0355 |
| src/compiler/utilities.ts:4526:1 | getNestedModuleDeclaration | 0.228650 | 0.386077 | +0.157426 | 1.6885 |
| src/compiler/commandLineParser.ts:1728:1 | isCommandLineOptionOfCustomType | 4.378550 | 4.531540 | +0.152990 | 1.0349 |
| src/compiler/utilities.ts:8637:1 | isDiagnosticWithDetachedLocation | 4.584402 | 4.728984 | +0.144582 | 1.0315 |
| src/compiler/utilities.ts:10913:1 | isFunctionExpressionOrArrowFunction | 0.042222 | 0.182252 | +0.140030 | 4.3165 |
| src/compiler/watchUtilities.ts:668:1 | isEmittedFileOfProgram | 0.070241 | 0.207653 | +0.137413 | 2.9563 |
| src/compiler/emitter.ts:540:1 | getOutputExtension | 4.390940 | 4.527693 | +0.136753 | 1.0311 |
| src/compiler/moduleSpecifiers.ts:301:1 | getModuleSpecifierWorker | 4.612407 | 4.748939 | +0.136532 | 1.0296 |
| src/compiler/transformers/module/module.ts:2499:1 |  | 0.061921 | 0.185255 | +0.123334 | 2.9918 |
| src/compiler/utilitiesPublic.ts:2143:1 | needsScopeMarker | 0.281589 | 0.398698 | +0.117109 | 1.4159 |
| src/compiler/utilitiesPublic.ts:852:1 | unescapeLeadingUnderscores | 0.037008 | 0.144906 | +0.107897 | 3.9155 |
| src/compiler/executeCommandLine.ts:1289:1 | writeConfigFile | 0.171131 | 0.276538 | +0.105407 | 1.6159 |
| src/compiler/utilities.ts:10949:1 | createPropertyNameNodeForIdentifierOrLiteral | 0.160537 | 0.265235 | +0.104698 | 1.6522 |
| src/compiler/transformers/destructuring.ts:556:1 | createDestructuringPropertyAccess | 0.219961 | 0.324278 | +0.104316 | 1.4742 |
| src/compiler/utilities.ts:8790:1 | compareDiagnosticsSkipRelatedInformation | 4.335902 | 4.439789 | +0.103887 | 1.0240 |
| src/compiler/utilitiesPublic.ts:1106:1 | hasJSDocParameterTags | 0.043182 | 0.147046 | +0.103864 | 3.4052 |
| src/compiler/utilities.ts:1029:1 | getEndLinePosition | 0.038193 | 0.137740 | +0.099547 | 3.6064 |
| src/compiler/utilities.ts:6729:1 | writeFileEnsuringDirectories | 0.035340 | 0.133290 | +0.097950 | 3.7716 |
| src/compiler/moduleSpecifiers.ts:182:1 | getModuleSpecifierPreferences | 0.345843 | 0.443206 | +0.097364 | 1.2815 |
| src/compiler/types.ts:6440:1 | TypeId | 0.037988 | 0.134412 | +0.096424 | 3.5382 |
| src/compiler/utilities.ts:10908:1 | isCatchClauseVariableDeclaration | 0.053730 | 0.148163 | +0.094433 | 2.7575 |
| src/compiler/utilities.ts:4780:1 | getJSDocOverloadTags | 0.307474 | 0.401294 | +0.093820 | 1.3051 |
| src/compiler/executeCommandLine.ts:1160:1 | reportStatistics | 0.241985 | 0.335490 | +0.093505 | 1.3864 |
| src/compiler/utilities.ts:3824:1 | isJSDocIndexSignature | 0.323453 | 0.413970 | +0.090517 | 1.2798 |
| src/compiler/utilities.ts:3027:1 | isCommonJsExportPropertyAssignment | 0.278843 | 0.369348 | +0.090505 | 1.3246 |
| src/compiler/utilitiesPublic.ts:842:1 | escapeLeadingUnderscores | 0.036024 | 0.125455 | +0.089431 | 3.4826 |
| src/compiler/types.ts:6444:1 | Type | 0.041243 | 0.130585 | +0.089342 | 3.1662 |
| src/compiler/utilities.ts:6716:1 | ensureDirectoriesExist | 0.036076 | 0.124673 | +0.088597 | 3.4559 |
| src/compiler/utilitiesPublic.ts:1390:1 | isOptionalChainRoot | 0.259166 | 0.346832 | +0.087665 | 1.3383 |
| src/compiler/utilities.ts:10918:1 | escapeSnippetText | 0.036293 | 0.123688 | +0.087395 | 3.4080 |
| src/compiler/utilities.ts:1062:1 | isFileLevelUniqueName | 0.039755 | 0.126481 | +0.086725 | 3.1815 |
| src/compiler/utilitiesPublic.ts:827:1 | getParseTreeNode | 0.036413 | 0.121595 | +0.085182 | 3.3393 |
| src/compiler/utilitiesPublic.ts:1096:1 | getJSDocTypeParameterTagsNoCache | 0.043059 | 0.128176 | +0.085117 | 2.9768 |
| src/compiler/utilities.ts:4509:1 | getSingleInitializerOfVariableStatementOrPropertyDeclaration | 0.151809 | 0.236360 | +0.084551 | 1.5570 |
| src/compiler/utilities.ts:2390:1 | entityNameToString | 0.192339 | 0.275350 | +0.083011 | 1.4316 |
| src/compiler/utilitiesPublic.ts:857:1 | idText | 0.035645 | 0.117733 | +0.082087 | 3.3029 |
| src/compiler/utilities.ts:11028:1 | isTypeDeclaration | 0.188097 | 0.269816 | +0.081718 | 1.4344 |
| src/compiler/types.ts:6319:1 | TrackedSymbol | 0.039435 | 0.120018 | +0.080583 | 3.0434 |
| src/compiler/utilities.ts:6691:1 | getSourceFilePathInNewDir | 0.144302 | 0.224441 | +0.080139 | 1.5554 |
| src/compiler/utilitiesPublic.ts:1111:1 | getJSDocAugmentsTag | 0.043563 | 0.123591 | +0.080027 | 2.8370 |
| src/compiler/utilitiesPublic.ts:817:1 | getParseTreeNode | 0.036952 | 0.115608 | +0.078655 | 3.1286 |
| src/compiler/factory/emitHelpers.ts:1486:1 | isCallToHelper | 0.288945 | 0.367495 | +0.078551 | 1.2719 |
| src/compiler/factory/utilities.ts:183:1 | createEmptyExports | 0.137107 | 0.214745 | +0.077638 | 1.5663 |
| src/compiler/utilities.ts:6661:1 | sourceFileMayBeEmitted | 0.330728 | 0.407977 | +0.077249 | 1.2336 |
| src/compiler/utilities.ts:1079:1 | nodeIsMissing | 0.040898 | 0.118039 | +0.077141 | 2.8862 |
| src/compiler/commandLineParser.ts:4118:1 | getWildcardDirectories | 0.113059 | 0.188691 | +0.075632 | 1.6690 |
| src/compiler/utilities.ts:5079:1 | getDeclarationFromName | 0.285793 | 0.360873 | +0.075081 | 1.2627 |
| src/compiler/resolutionCache.ts:1:1 |  | 0.038609 | 0.112352 | +0.073743 | 2.9100 |
| src/compiler/utilities.ts:2297:1 | isLateVisibilityPaintedStatement | 0.044400 | 0.116043 | +0.071643 | 2.6136 |
| src/compiler/factory/nodeTests.ts:362:1 | isAbstractModifier | 0.041435 | 0.112732 | +0.071297 | 2.7207 |
| src/compiler/tracing.ts:379:1 |  | 0.034968 | 0.105402 | +0.070435 | 3.0143 |
| src/compiler/factory/emitNode.ts:62:1 | disposeEmitNodes | 0.050662 | 0.120939 | +0.070277 | 2.3872 |
| src/compiler/utilities.ts:6703:1 | writeFile | 0.114852 | 0.184717 | +0.069865 | 1.6083 |
| src/compiler/utilitiesPublic.ts:2160:1 | isConciseBody | 0.261614 | 0.331080 | +0.069466 | 1.2655 |
| src/compiler/types.ts:6335:1 | TypeFlags | 0.036345 | 0.105527 | +0.069182 | 2.9035 |
| src/compiler/transformers/destructuring.ts:522:1 | isSimpleBindingOrAssignmentElement | 0.183416 | 0.251597 | +0.068181 | 1.3717 |
| src/compiler/utilitiesPublic.ts:1121:1 | getJSDocClassTag | 0.049052 | 0.116468 | +0.067416 | 2.3744 |
| src/compiler/resolutionCache.ts:82:1 | HasInvalidatedFromResolutionCache | 0.036975 | 0.104340 | +0.067365 | 2.8219 |
| src/compiler/watchUtilities.ts:677:1 | WatchLogLevel | 0.035497 | 0.101732 | +0.066235 | 2.8660 |
| src/compiler/utilitiesPublic.ts:1091:1 | getJSDocTypeParameterTags | 0.044347 | 0.110195 | +0.065848 | 2.4848 |
| src/compiler/utilities.ts:4064:1 | isSameEntityName | 0.260550 | 0.325881 | +0.065331 | 1.2507 |
| src/compiler/performanceCore.ts:1:1 |  | 0.038260 | 0.103491 | +0.065231 | 2.7049 |
| src/compiler/types.ts:6437:1 | DestructuringPattern | 0.038198 | 0.102891 | +0.064693 | 2.6936 |
| src/compiler/utilities.ts:4561:1 | canHaveJSDoc | 0.041239 | 0.105813 | +0.064574 | 2.5658 |
| src/compiler/utilities.ts:6749:1 | getLineOfLocalPosition | 0.035047 | 0.099479 | +0.064432 | 2.8385 |
| src/compiler/utilities.ts:4522:1 | getSingleVariableOfVariableStatement | 0.215117 | 0.279411 | +0.064294 | 1.2989 |
| src/compiler/utilities.ts:992:1 | getSourceFileOfModule | 0.045006 | 0.109195 | +0.064189 | 2.4262 |
| src/compiler/types.ts:6321:1 | SerializedTypeEntry | 0.041232 | 0.104794 | +0.063562 | 2.5416 |
| src/compiler/types.ts:8430:1 | UnscopedEmitHelper | 0.039419 | 0.102603 | +0.063184 | 2.6029 |
| src/compiler/utilitiesPublic.ts:1372:1 | isElementAccessChain | 0.233073 | 0.295471 | +0.062398 | 1.2677 |
| src/compiler/moduleSpecifiers.ts:453:1 | computeModuleSpecifiers | 0.247876 | 0.309972 | +0.062096 | 1.2505 |
| src/compiler/performanceCore.ts:7:1 | PerformanceHooks | 0.037910 | 0.099770 | +0.061860 | 2.6317 |
| src/compiler/utilitiesPublic.ts:1004:1 | getAssignedName | 0.276064 | 0.337081 | +0.061017 | 1.2210 |
| src/compiler/types.ts:8442:1 | LanugageFeatures | 0.041457 | 0.102259 | +0.060803 | 2.4667 |
| src/compiler/transformers/classThis.ts:72:1 | isClassThisAssignmentBlock | 0.248939 | 0.309469 | +0.060530 | 1.2432 |
| src/compiler/types.ts:4031:1 | JSDocReadonlyTag | 0.035078 | 0.095549 | +0.060471 | 2.7239 |
| src/compiler/utilitiesPublic.ts:1498:1 | isLiteralExpression | 0.068281 | 0.128554 | +0.060272 | 1.8827 |
| src/compiler/types.ts:4027:1 | JSDocProtectedTag | 0.034963 | 0.095179 | +0.060216 | 2.7223 |
| src/compiler/utilities.ts:1022:1 | nodePosToString | 0.043771 | 0.103486 | +0.059716 | 2.3643 |
| src/compiler/utilitiesPublic.ts:2183:1 | isNamespaceBody | 0.044321 | 0.103406 | +0.059085 | 2.3331 |
| src/compiler/types.ts:4035:1 | JSDocOverrideTag | 0.034719 | 0.093768 | +0.059048 | 2.7007 |
| src/compiler/utilities.ts:4535:1 | canHaveFlowNode | 0.050897 | 0.109355 | +0.058459 | 2.1486 |
| src/compiler/core.ts:1997:1 | compareTextSpans | 0.187846 | 0.245559 | +0.057714 | 1.3072 |
| src/compiler/utilitiesPublic.ts:2175:1 | isModuleBody | 0.043892 | 0.101593 | +0.057701 | 2.3146 |
| src/compiler/types.ts:6277:1 | NodeLinks | 0.040149 | 0.097797 | +0.057647 | 2.4358 |
| src/compiler/resolutionCache.ts:91:1 | ResolutionCache | 0.036190 | 0.093232 | +0.057043 | 2.5762 |
| src/compiler/types.ts:4056:1 | JSDocSeeTag | 0.035722 | 0.092468 | +0.056746 | 2.5885 |
| src/compiler/types.ts:4045:1 | JSDocThisTag | 0.036374 | 0.092558 | +0.056183 | 2.5446 |
| src/compiler/utilitiesPublic.ts:2190:1 | isJSDocNamespaceBody | 0.043255 | 0.098747 | +0.055492 | 2.2829 |
| src/compiler/utilities.ts:10923:1 | isNumericLiteralName | 0.038676 | 0.093800 | +0.055124 | 2.4253 |
| src/compiler/utilities.ts:5035:1 | skipTypeParentheses | 0.260821 | 0.314117 | +0.053296 | 1.2043 |
| src/compiler/moduleSpecifiers.ts:1186:1 | tryGetModuleNameAsNodeModule | 0.237728 | 0.290741 | +0.053013 | 1.2230 |
| src/compiler/utilities.ts:3841:1 | isRequireCall | 0.155003 | 0.207922 | +0.052918 | 1.3414 |
| src/compiler/utilities.ts:11130:1 | getEscapedTextOfJsxAttributeName | 0.198039 | 0.250731 | +0.052693 | 1.2661 |
| src/compiler/factory/emitHelpers.ts:1475:1 |  | 0.063704 | 0.115645 | +0.051941 | 1.8153 |
| src/compiler/types.ts:8439:1 | EmitHelperUniqueNameCallback | 0.039302 | 0.091067 | +0.051765 | 2.3171 |
| src/compiler/scanner.ts:358:1 | lookupInUnicodeMap | 0.038175 | 0.089651 | +0.051476 | 2.3484 |
| src/compiler/types.ts:4050:1 | JSDocTemplateTag | 0.037609 | 0.088958 | +0.051349 | 2.3654 |
| src/compiler/types.ts:8437:1 | EmitHelper | 0.038238 | 0.089320 | +0.051083 | 2.3359 |
| src/compiler/scanner.ts:468:1 | getPositionOfLineAndCharacter | 0.108240 | 0.158270 | +0.050030 | 1.4622 |
| src/compiler/factory/nodeTests.ts:352:1 | isReadonlyKeyword | 0.040850 | 0.090378 | +0.049528 | 2.2124 |
| src/compiler/utilities.ts:3055:1 | unwrapInnermostStatementOfLabel | 0.119266 | 0.168663 | +0.049397 | 1.4142 |
| src/compiler/factory/emitHelpers.ts:91:1 | ESDecorateName | 0.042275 | 0.090952 | +0.048677 | 2.1514 |
| src/compiler/utilitiesPublic.ts:2592:1 | isObjectLiteralElement | 0.226909 | 0.274842 | +0.047933 | 1.2112 |
| src/compiler/utilities.ts:3839:1 | isRequireCall | 0.157091 | 0.204771 | +0.047680 | 1.3035 |
| src/compiler/utilities.ts:3814:1 | isSourceFileNotJson | 0.214753 | 0.261933 | +0.047180 | 1.2197 |
| src/compiler/utilities.ts:2315:1 | hasPossibleExternalModuleReference | 0.280741 | 0.327598 | +0.046857 | 1.1669 |
| src/compiler/types.ts:8417:1 | EmitHelperBase | 0.040064 | 0.086244 | +0.046179 | 2.1526 |
| src/compiler/factory/utilities.ts:1627:1 | createAccessorPropertyGetRedirector | 0.260322 | 0.306445 | +0.046124 | 1.1772 |
| src/compiler/resolutionCache.ts:1711:1 | resolutionIsSymlink | 0.304052 | 0.350044 | +0.045991 | 1.1513 |
| src/compiler/utilities.ts:6923:1 | isNonTypeAliasTemplate | 0.266629 | 0.312360 | +0.045731 | 1.1715 |
| src/compiler/tracing.ts:378:1 |  | 0.035388 | 0.080803 | +0.045415 | 2.2833 |
| src/compiler/builder.ts:635:1 | getNextAffectedFile | 4.308219 | 4.353568 | +0.045349 | 1.0105 |
| src/compiler/utilities.ts:2117:1 | isBlockScopedContainerTopLevel | 0.241940 | 0.286210 | +0.044270 | 1.1830 |
| src/compiler/utilities.ts:3068:1 | isFunctionBlock | 0.272209 | 0.315365 | +0.043156 | 1.1585 |
| src/compiler/moduleNameResolver.ts:2199:1 | tryFileLookup | 0.074069 | 0.117176 | +0.043108 | 1.5820 |
| src/compiler/utilitiesPublic.ts:807:1 | isParseTreeNode | 0.039702 | 0.082757 | +0.043055 | 2.0845 |
| src/compiler/utilities.ts:4228:1 | getElementOrPropertyAccessName | 0.184307 | 0.226979 | +0.042671 | 1.2315 |
| src/compiler/transformers/utilities.ts:355:1 | collectExportedVariableInfo | 0.187830 | 0.230328 | +0.042499 | 1.2263 |
| src/compiler/types.ts:8425:1 | ScopedEmitHelper | 0.047714 | 0.090046 | +0.042332 | 1.8872 |
| src/compiler/utilities.ts:6530:1 | getExternalModuleNameFromPath | 0.179459 | 0.221693 | +0.042235 | 1.2353 |
| src/compiler/types.ts:4061:1 | JSDocReturnTag | 0.034822 | 0.075868 | +0.041047 | 2.1788 |
| src/compiler/transformers/destructuring.ts:541:1 | createDefaultValueCheck | 0.120285 | 0.160894 | +0.040608 | 1.3376 |
| src/compiler/types.ts:8406:1 | InternalEmitFlags | 0.037310 | 0.077280 | +0.039970 | 2.0713 |
| src/compiler/utilities.ts:6942:1 | emitNewLineBeforeLeadingCommentsOfPosition | 0.075402 | 0.115310 | +0.039908 | 1.5293 |
| src/compiler/types.ts:8499:1 |  | 0.062985 | 0.102653 | +0.039668 | 1.6298 |
| src/compiler/transformers/utilities.ts:880:1 | rewriteModuleSpecifier | 0.284533 | 0.324055 | +0.039522 | 1.1389 |
| src/compiler/utilities.ts:3186:1 | getThisContainer | 0.040367 | 0.079692 | +0.039324 | 1.9742 |
| src/compiler/utilities.ts:6953:1 | emitNewLineBeforeLeadingCommentOfPosition | 0.074869 | 0.113762 | +0.038893 | 1.5195 |
| src/compiler/types.ts:1625:1 | AssertKeyword | 0.035492 | 0.074356 | +0.038864 | 2.0950 |
| src/compiler/utilitiesPublic.ts:1605:1 | isPrivateIdentifierPropertyAccessExpression | 0.259759 | 0.298408 | +0.038649 | 1.1488 |
| src/compiler/factory/utilities.ts:697:1 | createExternalHelpersImportDeclarationIfNeeded | 0.065932 | 0.104374 | +0.038442 | 1.5831 |
| src/compiler/types.ts:8364:1 | SnippetKind | 0.039037 | 0.077423 | +0.038386 | 1.9833 |
| src/compiler/factory/nodeTests.ts:357:1 | isStaticModifier | 0.042774 | 0.080320 | +0.037547 | 1.8778 |
| src/compiler/performanceCore.ts:14:1 | PerformanceTime | 0.037695 | 0.075240 | +0.037544 | 1.9960 |
| src/compiler/factory/nodeTests.ts:894:1 | isNamedExports | 0.042287 | 0.079826 | +0.037539 | 1.8877 |
| src/compiler/utilities.ts:8024:1 | isInitializedVariable | 0.290598 | 0.328106 | +0.037508 | 1.1291 |
| src/compiler/factory/nodeTests.ts:434:1 | isClassStaticBlockDeclaration | 0.044309 | 0.081435 | +0.037127 | 1.8379 |
| src/compiler/watchUtilities.ts:684:1 | WatchFactoryHost | 0.034775 | 0.071544 | +0.036770 | 2.0574 |
| src/compiler/utilitiesPublic.ts:1318:1 | getEffectiveTypeParameterDeclarations | 0.199056 | 0.235554 | +0.036498 | 1.1834 |
| src/compiler/utilities.ts:6933:1 | getEffectiveSetAccessorTypeAnnotationNode | 0.046801 | 0.083204 | +0.036403 | 1.7778 |
| src/compiler/types.ts:3211:1 | JsxAttributeName | 0.047629 | 0.084008 | +0.036380 | 1.7638 |
| src/compiler/factory/emitNode.ts:240:1 | setConstantValue | 0.047874 | 0.084045 | +0.036171 | 1.7555 |
| src/compiler/factory/nodeTests.ts:438:1 | isConstructorDeclaration | 0.045063 | 0.081193 | +0.036129 | 1.8017 |
| src/compiler/types.ts:4023:1 | JSDocPrivateTag | 0.035689 | 0.071777 | +0.036088 | 2.0112 |
| src/compiler/types.ts:8372:1 | EmitFlags | 0.039675 | 0.075692 | +0.036017 | 1.9078 |
| src/compiler/factory/utilities.ts:865:1 | getInitializerOfBindingOrAssignmentElement | 0.255540 | 0.291448 | +0.035909 | 1.1405 |
| src/compiler/transformers/namedEvaluation.ts:67:1 | getAssignedNameOfPropertyName | 0.381630 | 0.417401 | +0.035771 | 1.0937 |
| src/compiler/utilities.ts:11119:1 | getJSDocSatisfiesExpressionType | 0.044083 | 0.079411 | +0.035329 | 1.8014 |
| src/compiler/executeCommandLine.ts:115:1 | getCountsMap | 0.034622 | 0.069875 | +0.035253 | 2.0182 |
| src/compiler/types.ts:4039:1 | JSDocEnumTag | 0.034755 | 0.069090 | +0.034335 | 1.9879 |
| src/compiler/_namespaces/ts.ts:38:1 |  | 0.036809 | 0.071062 | +0.034253 | 1.9306 |
| src/compiler/moduleNameResolver.ts:2223:1 | GetPackageJsonEntrypointsHost | 0.036406 | 0.070614 | +0.034207 | 1.9396 |
| src/compiler/factory/emitHelpers.ts:85:1 | ESDecorateClassElementAccess | 0.036549 | 0.070730 | +0.034181 | 1.9352 |
| src/compiler/utilities.ts:6759:1 | getFirstConstructorWithBody | 0.034114 | 0.068092 | +0.033978 | 1.9960 |
| src/compiler/utilitiesPublic.ts:1968:1 | isCallLikeExpression | 0.153424 | 0.187361 | +0.033938 | 1.2212 |
| src/compiler/utilities.ts:4349:1 | tryGetModuleSpecifierFromDeclaration | 0.041168 | 0.075043 | +0.033875 | 1.8228 |
| src/compiler/utilities.ts:6793:1 | parameterIsThisKeyword | 0.039832 | 0.073668 | +0.033836 | 1.8495 |
| src/compiler/utilities.ts:10760:1 | isPackedArrayLiteral | 0.264423 | 0.297616 | +0.033193 | 1.1255 |
| src/compiler/scanner.ts:465:1 | getPositionOfLineAndCharacter | 0.112892 | 0.145822 | +0.032930 | 1.2917 |
| src/compiler/factory/emitNode.ts:132:1 | getSourceMapRange | 0.038911 | 0.071744 | +0.032833 | 1.8438 |
| src/compiler/transformer.ts:202:1 | wrapCustomTransformer | 0.182912 | 0.215733 | +0.032821 | 1.1794 |
| src/compiler/types.ts:1627:1 | CaseKeyword | 0.035944 | 0.068761 | +0.032817 | 1.9130 |
| src/compiler/moduleNameResolver.ts:3202:1 | mangleScopedPackageNameWithTrace | 0.037634 | 0.070012 | +0.032378 | 1.8603 |
| src/compiler/factory/nodeTests.ts:971:1 | isJsxNamespacedName | 0.042188 | 0.074492 | +0.032304 | 1.7657 |
| src/compiler/sourcemap.ts:630:1 | isSourceMapping | 4.498601 | 4.530754 | +0.032153 | 1.0071 |
| src/compiler/types.ts:1618:1 | QuestionDotToken | 0.035411 | 0.067560 | +0.032148 | 1.9079 |
| src/compiler/scanner.ts:222:1 |  | 0.039088 | 0.071126 | +0.032038 | 1.8196 |
| src/compiler/utilities.ts:8644:1 | attachFileToDiagnostic | 0.039959 | 0.071888 | +0.031929 | 1.7991 |
| src/compiler/utilities.ts:4785:1 | getHostSignatureFromJSDoc | 0.269757 | 0.301433 | +0.031676 | 1.1174 |
| src/compiler/moduleNameResolver.ts:2217:1 | loadNodeModuleFromDirectory | 0.038867 | 0.070060 | +0.031193 | 1.8025 |
| src/compiler/utilitiesPublic.ts:1257:1 | getJSDocTagsWorker | 0.112546 | 0.143645 | +0.031100 | 1.2763 |
| src/compiler/utilities.ts:9044:1 |  | 4.521345 | 4.552372 | +0.031028 | 1.0069 |
| src/compiler/utilities.ts:1088:1 | nodeIsPresent | 0.042546 | 0.073545 | +0.030998 | 1.7286 |
| src/compiler/moduleSpecifiers.ts:335:1 | tryGetModuleSpecifiersFromCacheWorker | 0.166033 | 0.196862 | +0.030829 | 1.1857 |
| src/compiler/transformers/classFields.ts:3327:1 | createPrivateInstanceFieldInitializer | 0.166340 | 0.197116 | +0.030776 | 1.1850 |
| src/compiler/utilities.ts:6754:1 | getLineOfLocalPositionFromLineMap | 0.035076 | 0.065768 | +0.030692 | 1.8750 |
| src/compiler/utilities.ts:7430:1 | isLogicalOrCoalescingBinaryExpression | 0.278952 | 0.309530 | +0.030578 | 1.1096 |
| src/compiler/executeCommandLine.ts:126:1 | getCountKey | 0.071296 | 0.101851 | +0.030555 | 1.4286 |
| src/compiler/utilities.ts:10962:1 | NodeModulePathParts | 0.037359 | 0.067883 | +0.030524 | 1.8170 |
| src/compiler/performanceCore.ts:28:1 |  | 0.036060 | 0.066490 | +0.030430 | 1.8439 |
| src/compiler/utilities.ts:4407:1 | getExternalModuleName | 0.046641 | 0.076918 | +0.030277 | 1.6492 |
| src/compiler/builder.ts:841:1 | handleDtsMayChangeOf | 0.075930 | 0.106110 | +0.030180 | 1.3975 |
| src/compiler/utilitiesPublic.ts:1376:1 | isCallChain | 0.259091 | 0.289157 | +0.030066 | 1.1160 |
| src/compiler/utilities.ts:6938:1 | emitNewLineBeforeLeadingComments | 0.037197 | 0.067192 | +0.029995 | 1.8064 |
| src/compiler/utilities.ts:6695:1 | getSourceFilePathInNewDirWorker | 0.036146 | 0.066067 | +0.029921 | 1.8278 |
| src/compiler/factory/utilities.ts:1155:1 | isIdentifierOrThisTypeNode | 0.256994 | 0.286911 | +0.029917 | 1.1164 |
| src/compiler/utilities.ts:4935:1 | isInCompoundLikeAssignment | 0.156431 | 0.185820 | +0.029389 | 1.1879 |
| src/compiler/utilitiesPublic.ts:997:1 | getNameOfDeclaration | 0.204265 | 0.233638 | +0.029373 | 1.1438 |
| src/compiler/moduleNameResolver.ts:2131:1 | tryAddingExtensions | 0.037765 | 0.066982 | +0.029217 | 1.7736 |
| src/compiler/sys.ts:111:1 | getModifiedTime | 0.069173 | 0.098329 | +0.029156 | 1.4215 |
| src/compiler/utilities.ts:4280:1 | getInitializerOfBinaryExpression | 0.282534 | 0.310947 | +0.028413 | 1.1006 |
| src/compiler/types.ts:1614:1 | AsteriskToken | 0.035736 | 0.063889 | +0.028154 | 1.7878 |
| src/compiler/utilities.ts:10969:1 | getNodeModulePathParts | 0.042683 | 0.070821 | +0.028138 | 1.6592 |
| src/compiler/types.ts:1616:1 | PlusToken | 0.034654 | 0.062723 | +0.028069 | 1.8100 |
| src/compiler/factory/nodeTests.ts:963:1 | isJsxSpreadAttribute | 0.041939 | 0.069931 | +0.027992 | 1.6674 |
| src/compiler/utilities.ts:11051:1 | canHaveExportModifier | 0.276510 | 0.304383 | +0.027873 | 1.1008 |
| src/compiler/executeCommandLine.ts:99:1 | StatisticType | 0.035745 | 0.063470 | +0.027725 | 1.7756 |
| src/compiler/sys.ts:593:1 | createDirectoryWatcherSupportingRecursive | 0.153525 | 0.180780 | +0.027255 | 1.1775 |
| src/compiler/moduleNameResolver.ts:1678:1 | resolveJSModule | 0.113006 | 0.140162 | +0.027156 | 1.2403 |
| src/compiler/scanner.ts:51:1 | Scanner | 0.044145 | 0.070767 | +0.026622 | 1.6031 |
| src/compiler/utilities.ts:7484:1 | isDestructuringAssignment | 0.151637 | 0.177899 | +0.026261 | 1.1732 |
| src/compiler/transformers/utilities.ts:791:1 | PrivateEnvironment | 0.043090 | 0.069271 | +0.026181 | 1.6076 |
| src/compiler/factory/nodeTests.ts:450:1 | isCallSignatureDeclaration | 0.044209 | 0.070215 | +0.026005 | 1.5882 |
| src/compiler/utilities.ts:6776:1 | getSetAccessorTypeAnnotationNode | 0.039602 | 0.065434 | +0.025832 | 1.6523 |
| src/compiler/executeCommandLine.ts:812:1 | performBuild | 0.284120 | 0.309932 | +0.025812 | 1.0908 |
| src/compiler/utilities.ts:8325:1 | isObjectTypeDeclaration | 0.283270 | 0.308737 | +0.025467 | 1.0899 |
| src/compiler/executeCommandLine.ts:941:1 | performIncrementalCompilation | 0.235758 | 0.260893 | +0.025135 | 1.1066 |
| src/compiler/transformers/utilities.ts:485:1 | isSimpleInlineableExpression | 0.239718 | 0.264704 | +0.024985 | 1.1042 |
| src/compiler/parser.ts:1408:1 | JsDocWithDiagnostics | 0.035170 | 0.059972 | +0.024802 | 1.7052 |
| src/compiler/factory/utilities.ts:760:1 | getImportedHelpers | 0.038515 | 0.063161 | +0.024646 | 1.6399 |
| src/compiler/path.ts:165:1 | getEncodedRootLength | 0.036870 | 0.061428 | +0.024558 | 1.6661 |
| src/compiler/executeCommandLine.ts:149:1 | updateReportDiagnostic | 0.033972 | 0.058351 | +0.024379 | 1.7176 |
| src/compiler/types.ts:1624:1 | AssertsKeyword | 0.036127 | 0.060185 | +0.024058 | 1.6659 |
| src/compiler/parser.ts:414:1 | SpeculationKind | 0.034317 | 0.058200 | +0.023883 | 1.6959 |
| src/compiler/types.ts:3269:1 | JsxClosingFragment | 0.040068 | 0.063796 | +0.023728 | 1.5922 |
| src/compiler/factory/utilities.ts:1701:1 | flattenCommaListWorker | 0.279783 | 0.303475 | +0.023693 | 1.0847 |
| src/compiler/sys.ts:1379:1 | patchWriteFileEnsuringDirectory | 0.070976 | 0.094644 | +0.023667 | 1.3335 |
| src/compiler/moduleNameResolver.ts:1341:1 | createModuleResolutionCache | 0.222628 | 0.246261 | +0.023633 | 1.1062 |
| src/compiler/utilities.ts:2098:1 | isEffectiveModuleDeclaration | 0.256784 | 0.280376 | +0.023592 | 1.0919 |
| src/compiler/utilities.ts:10627:1 | isIdentifierTypeReference | 0.281069 | 0.304628 | +0.023559 | 1.0838 |
| src/compiler/factory/utilities.ts:1163:1 | isQuestionOrPlusOrMinusToken | 0.285740 | 0.308769 | +0.023029 | 1.0806 |
| src/compiler/utilities.ts:2359:1 | isComputedNonLiteralName | 0.290336 | 0.313364 | +0.023028 | 1.0793 |
| src/compiler/factory/nodeTests.ts:959:1 | isJsxAttributes | 0.041751 | 0.064675 | +0.022924 | 1.5491 |
| src/compiler/parser.ts:421:1 |  | 0.034448 | 0.057371 | +0.022923 | 1.6654 |
| src/compiler/moduleNameResolver.ts:3211:1 | getTypesPackageName | 0.036885 | 0.059795 | +0.022910 | 1.6211 |
| src/compiler/types.ts:2205:1 | ImportTypeNode | 0.037055 | 0.059768 | +0.022714 | 1.6130 |
| src/compiler/factory/utilities.ts:1151:1 | isQuestionOrExclamationToken | 0.259652 | 0.282332 | +0.022680 | 1.0873 |
| src/compiler/builder.ts:933:1 | handleDtsMayChangeOfReferencingExportOfAffectedFile | 0.125154 | 0.147594 | +0.022440 | 1.1793 |
| src/compiler/utilities.ts:4099:1 | isModuleIdentifier | 0.231287 | 0.253687 | +0.022400 | 1.0969 |
| src/compiler/watchUtilities.ts:698:1 | GetDetailWatchInfo | 0.035363 | 0.057574 | +0.022211 | 1.6281 |
| src/compiler/factory/nodeTests.ts:878:1 | isNamedImports | 0.040610 | 0.062783 | +0.022173 | 1.5460 |
| src/compiler/utilities.ts:7473:1 | isAssignmentExpression | 0.282872 | 0.304844 | +0.021972 | 1.0777 |
| src/compiler/transformers/module/system.ts:137:1 | transformSystemModule | 0.386900 | 0.408870 | +0.021970 | 1.0568 |
| src/compiler/utilities.ts:11114:1 | isJSDocSatisfiesExpression | 0.277543 | 0.299454 | +0.021912 | 1.0789 |
| src/compiler/utilities.ts:7500:1 | isEntityNameExpression | 0.261403 | 0.283252 | +0.021848 | 1.0836 |
| src/compiler/utilities.ts:3020:1 | isCommonJsExportedExpression | 0.266672 | 0.288336 | +0.021664 | 1.0812 |
| src/compiler/utilities.ts:2364:1 | tryGetTextOfPropertyName | 0.286158 | 0.307561 | +0.021403 | 1.0748 |
| src/compiler/factory/emitHelpers.ts:1184:1 |  | 0.063659 | 0.084965 | +0.021307 | 1.3347 |
| src/compiler/factory/nodeFactory.ts:7431:1 | mergeEmitNode | 0.417297 | 0.438564 | +0.021267 | 1.0510 |
| src/compiler/factory/utilities.ts:683:1 | getExternalHelpersModuleName | 0.040897 | 0.062118 | +0.021220 | 1.5189 |
| src/compiler/types.ts:3221:1 | JsxTagNamePropertyAccess | 0.047293 | 0.068406 | +0.021113 | 1.4464 |
| src/compiler/transformers/namedEvaluation.ts:139:1 | isClassNamedEvaluationHelperBlock | 0.226018 | 0.247127 | +0.021109 | 1.0934 |
| src/compiler/utilities.ts:7296:1 | getRawJSDocModifierFlagsNoCache | 0.274714 | 0.295744 | +0.021030 | 1.0766 |
| src/compiler/utilities.ts:4475:1 | isJSDocConstructSignature | 0.233315 | 0.254145 | +0.020830 | 1.0893 |
| src/compiler/parser.ts:1:1 |  | 0.035981 | 0.056689 | +0.020708 | 1.5755 |
| src/compiler/builder.ts:291:1 | getPendingEmitKind | 0.044622 | 0.065260 | +0.020639 | 1.4625 |
| src/compiler/utilities.ts:4701:1 | ownsJSDocTag | 0.278006 | 0.298582 | +0.020577 | 1.0740 |
| src/compiler/watch.ts:351:1 | explainFiles | 0.171668 | 0.192191 | +0.020524 | 1.1196 |
| src/compiler/utilities.ts:4381:1 | tryGetImportFromModuleSpecifier | 0.048227 | 0.068748 | +0.020522 | 1.4255 |
| src/compiler/sys.ts:896:1 | FsWatchWorkerWatcher | 0.038546 | 0.059050 | +0.020504 | 1.5319 |
| src/compiler/transformers/destructuring.ts:583:1 | ensureIdentifier | 0.228545 | 0.248982 | +0.020437 | 1.0894 |
| src/compiler/moduleNameResolver.ts:2089:1 | loadModuleFromFileNoImplicitExtensions | 0.037378 | 0.057557 | +0.020178 | 1.5398 |
| src/compiler/utilities.ts:5690:1 | getOperator | 0.164000 | 0.184063 | +0.020063 | 1.1223 |
| src/compiler/utilitiesPublic.ts:2138:1 | hasScopeMarker | 0.049465 | 0.069428 | +0.019964 | 1.4036 |
| src/compiler/utilities.ts:7471:1 | isAssignmentExpression | 0.272671 | 0.292623 | +0.019952 | 1.0732 |
| src/compiler/executeCommandLine.ts:800:1 | reportWatchModeWithoutSysSupport | 0.137989 | 0.157923 | +0.019934 | 1.1445 |
| src/compiler/factory/nodeTests.ts:951:1 | isJsxClosingFragment | 0.040609 | 0.060497 | +0.019888 | 1.4897 |
| src/compiler/transformers/declarations/diagnostics.ts:1:1 |  | 0.040201 | 0.059953 | +0.019752 | 1.4913 |
| src/compiler/types.ts:1612:1 | BarBarEqualsToken | 0.035029 | 0.054698 | +0.019669 | 1.5615 |
| src/compiler/factory/nodeTests.ts:890:1 | isExportDeclaration | 0.040737 | 0.060177 | +0.019440 | 1.4772 |
| src/compiler/utilities.ts:3843:1 | isRequireCall | 0.161269 | 0.180650 | +0.019381 | 1.1202 |
| src/compiler/visitorPublic.ts:436:1 | addDefaultValueAssignmentIfNeeded | 0.179329 | 0.198700 | +0.019371 | 1.1080 |
| src/compiler/tracing.ts:374:1 | TracingNode | 0.034856 | 0.053851 | +0.018995 | 1.5449 |
| src/compiler/utilities.ts:10903:1 | isInfinityOrNaNString | 0.049861 | 0.068738 | +0.018877 | 1.3786 |
| src/compiler/builderState.ts:39:1 | getFileEmitOutput | 0.149025 | 0.167692 | +0.018668 | 1.1253 |
| src/compiler/utilities.ts:2129:1 | isExternalModuleAugmentation | 0.281823 | 0.299977 | +0.018155 | 1.0644 |
| src/compiler/types.ts:1626:1 | AwaitKeyword | 0.036793 | 0.054884 | +0.018091 | 1.4917 |
| src/compiler/factory/utilities.ts:1649:1 | createAccessorPropertySetRedirector | 0.225730 | 0.243737 | +0.018008 | 1.0798 |
| src/compiler/commandLineParser.ts:4089:1 | validateSpecs | 0.091482 | 0.109281 | +0.017798 | 1.1946 |
| src/compiler/commandLineParser.ts:1879:1 | parseCustomTypeOption | 0.048301 | 0.066070 | +0.017768 | 1.3679 |
| src/compiler/builder.ts:889:1 | removeSemanticDiagnosticsOf | 0.037393 | 0.055127 | +0.017734 | 1.4743 |
| src/compiler/core.ts:2032:1 | compareStringsCaseInsensitive | 0.041509 | 0.059211 | +0.017703 | 1.4265 |
| src/compiler/factory/nodeTests.ts:446:1 | isSetAccessorDeclaration | 0.054671 | 0.072324 | +0.017653 | 1.3229 |
| src/compiler/executeCommandLine.ts:92:1 | Statistic | 0.036668 | 0.054317 | +0.017649 | 1.4813 |
| src/compiler/types.ts:2217:1 | ThisTypeNode | 0.039861 | 0.057421 | +0.017560 | 1.4405 |
| src/compiler/utilitiesPublic.ts:1995:1 | isLeftHandSideExpression | 0.040615 | 0.058169 | +0.017554 | 1.4322 |
| src/compiler/executeCommandLine.ts:105:1 | countLines | 0.038261 | 0.055560 | +0.017300 | 1.4522 |
| src/compiler/builder.ts:2254:1 | createBuilderProgramUsingIncrementalBuildInfo | 0.075904 | 0.093144 | +0.017240 | 1.2271 |
| src/compiler/types.ts:2138:1 | GetAccessorDeclaration | 0.039379 | 0.056569 | +0.017190 | 1.4365 |
| src/compiler/utilitiesPublic.ts:1600:1 | isPrivateIdentifierClassElementDeclaration | 0.284776 | 0.301948 | +0.017172 | 1.0603 |
| src/compiler/sys.ts:458:1 | createFixedChunkSizePollingWatchFile | 0.035412 | 0.052552 | +0.017140 | 1.4840 |
| src/compiler/moduleSpecifiers.ts:1345:1 | tryGetAnyFileFromPath | 0.080472 | 0.097518 | +0.017046 | 1.2118 |
| src/compiler/watch.ts:427:1 | getMatchedIncludeSpec | 0.070703 | 0.087635 | +0.016932 | 1.2395 |
| src/compiler/_namespaces/ts.ts:37:1 |  | 0.036820 | 0.053743 | +0.016923 | 1.4596 |
| src/compiler/utilities.ts:9973:1 |  | 0.035805 | 0.052564 | +0.016759 | 1.4681 |
| src/compiler/transformers/utilities.ts:107:1 | getOriginalNodeId | 0.045751 | 0.062311 | +0.016560 | 1.3620 |
| src/compiler/factory/emitNode.ts:36:1 | getOrCreateEmitNode | 0.047609 | 0.063969 | +0.016360 | 1.3436 |
| src/compiler/scanner.ts:47:1 | tokenIsIdentifierOrKeywordOrGreaterThan | 0.062692 | 0.079016 | +0.016323 | 1.2604 |
| src/compiler/utilitiesPublic.ts:1351:1 | getEffectiveConstraintOfTypeParameter | 0.304526 | 0.320848 | +0.016322 | 1.0536 |
| src/compiler/factory/utilities.ts:1725:1 | flattenCommaList | 0.052182 | 0.068319 | +0.016137 | 1.3092 |
| src/compiler/utilities.ts:2724:1 | isLiteralImportTypeNode | 0.293736 | 0.309855 | +0.016119 | 1.0549 |
| src/compiler/checker.ts:1:1 |  | 0.040140 | 0.056130 | +0.015989 | 1.3983 |
| src/compiler/factory/nodeTests.ts:906:1 | isMissingDeclaration | 0.039912 | 0.055689 | +0.015777 | 1.3953 |
| src/compiler/moduleSpecifiers.ts:1169:1 | tryGetModuleNameFromRootDirs | 0.071932 | 0.087683 | +0.015751 | 1.2190 |
| src/compiler/utilities.ts:7970:1 | getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter | 0.037480 | 0.053167 | +0.015686 | 1.4185 |
| src/compiler/utilities.ts:4226:1 | getElementOrPropertyAccessName | 0.212960 | 0.228606 | +0.015646 | 1.0735 |
| src/compiler/transformers/classFields.ts:3318:1 | createPrivateStaticFieldInitializer | 0.073348 | 0.088984 | +0.015636 | 1.2132 |
| src/compiler/factory/nodeTests.ts:898:1 | isExportSpecifier | 0.039406 | 0.055027 | +0.015620 | 1.3964 |
| src/compiler/moduleNameResolver.ts:1327:1 | createModuleResolutionCache | 0.219919 | 0.235379 | +0.015460 | 1.0703 |
| src/compiler/utilities.ts:10957:1 | isThisTypeParameter | 0.040641 | 0.056079 | +0.015438 | 1.3799 |
| src/compiler/executeCommandLine.ts:908:1 | performCompilation | 0.309858 | 0.325219 | +0.015361 | 1.0496 |
| src/compiler/diagnosticInformationMap.generated.ts:12:1 |  | 0.167948 | 0.183155 | +0.015207 | 1.0905 |
| src/compiler/factory/nodeTests.ts:910:1 | isNotEmittedStatement | 0.039651 | 0.054822 | +0.015171 | 1.3826 |
| src/compiler/utilities.ts:11066:1 | canUsePropertyAccess | 0.036902 | 0.052036 | +0.015134 | 1.4101 |
| src/compiler/utilities.ts:1002:1 | isStatementWithLocals | 0.050574 | 0.065656 | +0.015083 | 1.2982 |
| src/compiler/moduleSpecifiers.ts:1463:1 | getDefaultResolutionModeForFile | 0.207563 | 0.222579 | +0.015015 | 1.0723 |
| src/compiler/semver.ts:141:1 | tryParseComponents | 0.040050 | 0.055060 | +0.015009 | 1.3748 |
| src/compiler/types.ts:28:1 | TextRange | 0.034968 | 0.049948 | +0.014980 | 1.4284 |
| src/compiler/utilities.ts:7965:1 | getStartPositionOfRange | 0.036814 | 0.051669 | +0.014855 | 1.4035 |
| src/compiler/utilities.ts:3931:1 | getEffectiveInitializer | 0.270426 | 0.285275 | +0.014850 | 1.0549 |
| src/compiler/utilitiesPublic.ts:2495:1 | isJsxCallLike | 0.040617 | 0.055338 | +0.014721 | 1.3624 |
| src/compiler/moduleSpecifiers.ts:581:1 | getLocalModuleSpecifier | 0.035372 | 0.049981 | +0.014609 | 1.4130 |
| src/compiler/transformers/classFields.ts:3358:1 | isStaticPropertyDeclarationOrClassStaticBlock | 0.238314 | 0.252916 | +0.014602 | 1.0613 |
| src/compiler/moduleNameResolver.ts:2059:1 | moveToNextDirectorySeparatorIfAvailable | 0.035934 | 0.050397 | +0.014463 | 1.4025 |
| src/compiler/factory/nodeTests.ts:488:1 | isArrayTypeNode | 0.040651 | 0.055072 | +0.014421 | 1.3547 |
| src/compiler/commandLineParser.ts:4268:1 | convertCompilerOptionsForTelemetry | 0.037009 | 0.051403 | +0.014393 | 1.3889 |
| src/compiler/scanner.ts:39:1 | ErrorCallback | 0.041064 | 0.055349 | +0.014286 | 1.3479 |
| src/compiler/transformers/es2015.ts:283:1 | Jump | 0.037834 | 0.052060 | +0.014226 | 1.3760 |
| src/compiler/types.ts:2193:1 | KeywordTypeNode | 0.041246 | 0.055455 | +0.014209 | 1.3445 |
| src/compiler/utilities.ts:6798:1 | isThisIdentifier | 0.117314 | 0.131466 | +0.014152 | 1.1206 |
| src/compiler/utilities.ts:7495:1 | isExpressionWithTypeArgumentsInClassExtendsClause | 0.044191 | 0.058241 | +0.014051 | 1.3180 |
| src/compiler/transformers/module/module.ts:174:1 | transformModule | 0.245179 | 0.259226 | +0.014048 | 1.0573 |
| src/compiler/factory/nodeTests.ts:780:1 | isThrowStatement | 0.041191 | 0.055180 | +0.013990 | 1.3396 |
| src/compiler/utilities.ts:7603:1 | isEmptyArrayLiteral | 0.144841 | 0.158741 | +0.013900 | 1.0960 |
| src/compiler/types.ts:3146:1 | TypeAssertion | 0.035823 | 0.049653 | +0.013830 | 1.3861 |
| src/compiler/sys.ts:115:1 | Levels | 0.034801 | 0.048613 | +0.013812 | 1.3969 |
| src/compiler/utilitiesPublic.ts:457:1 | createTextSpanFromBounds | 0.035409 | 0.049197 | +0.013788 | 1.3894 |
| src/compiler/types.ts:3405:1 | DoStatement | 0.034964 | 0.048721 | +0.013757 | 1.3935 |
| src/compiler/builderPublic.ts:1:1 |  | 0.036083 | 0.049784 | +0.013701 | 1.3797 |
| src/compiler/core.ts:2105:1 |  | 0.034969 | 0.048644 | +0.013675 | 1.3911 |
| src/compiler/utilities.ts:4654:1 | getJSDocCommentsAndTags | 0.217979 | 0.231643 | +0.013664 | 1.0627 |
| src/compiler/utilities.ts:984:1 | getSourceFileOfNode | 0.040977 | 0.054477 | +0.013500 | 1.3295 |
| src/compiler/utilities.ts:10395:1 | HostWithIsSourceOfProjectReferenceRedirect | 0.036673 | 0.050139 | +0.013466 | 1.3672 |
| src/compiler/builderState.ts:100:1 | BuilderState | 0.039955 | 0.053401 | +0.013447 | 1.3365 |
| src/compiler/builder.ts:1004:1 | handleDtsMayChangeOfFileAndExportsOfFile | 0.036868 | 0.050308 | +0.013441 | 1.3646 |
| src/compiler/semver.ts:46:1 | Version | 0.235857 | 0.249294 | +0.013437 | 1.0570 |
| src/compiler/types.ts:917:1 | RelationComparisonResult | 0.035181 | 0.048599 | +0.013418 | 1.3814 |
| src/compiler/emitter.ts:6360:1 | EmitListItemFunction | 0.036200 | 0.049588 | +0.013389 | 1.3699 |
| src/compiler/sys.ts:88:1 | DirectoryWatcherCallback | 0.035617 | 0.049005 | +0.013387 | 1.3759 |
| src/compiler/utilities.ts:2416:1 | createDiagnosticForNode | 0.048711 | 0.062095 | +0.013384 | 1.2748 |
| src/compiler/types.ts:905:1 | JsxFlags | 0.035688 | 0.049024 | +0.013336 | 1.3737 |
| src/compiler/moduleNameResolver.ts:2044:1 | parseNodeModuleFromPath | 0.036824 | 0.050143 | +0.013319 | 1.3617 |
| src/compiler/factory/emitNode.ts:195:1 | getSyntheticLeadingComments | 0.039252 | 0.052568 | +0.013316 | 1.3392 |
| src/compiler/factory/utilities.ts:1739:1 | containsObjectRestOrSpread | 0.231166 | 0.244465 | +0.013298 | 1.0575 |
| src/compiler/types.ts:3738:1 | AssertClause | 0.035118 | 0.048366 | +0.013248 | 1.3772 |
| src/compiler/factory/emitHelpers.ts:72:1 | ESDecorateClassElementContext | 0.038935 | 0.052179 | +0.013244 | 1.3402 |
| src/compiler/utilities.ts:2511:1 | getSpanOfTokenAtPosition | 0.146375 | 0.159584 | +0.013209 | 1.0902 |
| src/compiler/utilities.ts:6483:1 | hostUsesCaseSensitiveFileNames | 0.113689 | 0.126848 | +0.013159 | 1.1157 |
| src/compiler/factory/nodeTests.ts:676:1 | isAsExpression | 0.045066 | 0.058212 | +0.013146 | 1.2917 |
| src/compiler/checker.ts:1151:1 |  | 0.036580 | 0.049639 | +0.013059 | 1.3570 |
| src/compiler/sys.ts:87:1 | FileWatcherCallback | 0.036201 | 0.049248 | +0.013047 | 1.3604 |
| src/compiler/watch.ts:668:1 |  | 0.059306 | 0.072311 | +0.013005 | 1.2193 |
| src/compiler/_namespaces/ts.ts:31:1 |  | 0.036307 | 0.049269 | +0.012963 | 1.3570 |
| src/compiler/transformers/utilities.ts:124:1 | containsDefaultReference | 0.180396 | 0.193334 | +0.012938 | 1.0717 |
| src/compiler/core.ts:2086:1 |  | 0.035487 | 0.048392 | +0.012905 | 1.3637 |
| src/compiler/utilitiesPublic.ts:788:1 | findAncestor | 0.036274 | 0.049174 | +0.012899 | 1.3556 |
| src/compiler/types.ts:10427:1 | ReadonlyPragmaMap | 0.040331 | 0.053223 | +0.012893 | 1.3197 |
| src/compiler/commandLineParser.ts:4174:1 | CanonicalKey | 0.037012 | 0.049902 | +0.012890 | 1.3483 |
| src/compiler/utilities.ts:8676:1 | attachFileToDiagnostics | 0.040038 | 0.052915 | +0.012877 | 1.3216 |
| src/compiler/scanner.ts:135:1 |  | 0.040112 | 0.052884 | +0.012772 | 1.3184 |
| src/compiler/factory/nodeTests.ts:644:1 | isBinaryExpression | 0.042420 | 0.055187 | +0.012767 | 1.3010 |
| src/compiler/utilities.ts:7250:1 | getModifierFlagsWorker | 0.052471 | 0.065232 | +0.012761 | 1.2432 |
| src/compiler/core.ts:2151:1 | compareBooleans | 0.035028 | 0.047779 | +0.012751 | 1.3640 |
| src/compiler/factory/nodeTests.ts:870:1 | isNamespaceImport | 0.042972 | 0.055699 | +0.012727 | 1.2962 |
| src/compiler/semver.ts:443:1 | formatDisjunction | 0.036498 | 0.049214 | +0.012716 | 1.3484 |
| src/compiler/moduleNameResolver.ts:1766:1 | bundlerModuleNameResolver | 0.039317 | 0.052006 | +0.012689 | 1.3227 |
| src/compiler/types.ts:685:1 | KeywordTypeSyntaxKind | 0.034921 | 0.047485 | +0.012564 | 1.3598 |
| src/compiler/utilities.ts:4376:1 | importFromModuleSpecifier | 0.038237 | 0.050699 | +0.012461 | 1.3259 |
| src/compiler/types.ts:3152:1 | SatisfiesExpression | 0.036016 | 0.048457 | +0.012442 | 1.3454 |
| src/compiler/utilities.ts:2276:1 | isAnyImportSyntax | 0.040513 | 0.052915 | +0.012403 | 1.3061 |
| src/compiler/utilities.ts:6540:1 | getOwnEmitOutputFilePath | 0.072253 | 0.084636 | +0.012383 | 1.1714 |
| src/compiler/utilitiesPublic.ts:1141:1 | getJSDocPrivateTagNoCache | 0.049980 | 0.062349 | +0.012369 | 1.2475 |
| src/compiler/core.ts:2106:1 |  | 0.035548 | 0.047906 | +0.012358 | 1.3476 |
| src/compiler/utilities.ts:9976:1 |  | 0.036494 | 0.048832 | +0.012338 | 1.3381 |
| src/compiler/sys.ts:900:1 | FsWatchWorker | 0.036866 | 0.049196 | +0.012330 | 1.3344 |
| src/compiler/types.ts:3372:1 | Block | 0.035778 | 0.048054 | +0.012276 | 1.3431 |
| src/compiler/utilities.ts:3085:1 | isIdentifierTypePredicate | 0.036597 | 0.048872 | +0.012275 | 1.3354 |
| src/compiler/types.ts:3108:1 | ExpressionWithTypeArguments | 0.036111 | 0.048385 | +0.012275 | 1.3399 |
| src/compiler/utilities.ts:9975:1 |  | 0.039207 | 0.051473 | +0.012266 | 1.3129 |
| src/compiler/factory/nodeTests.ts:442:1 | isGetAccessorDeclaration | 0.052121 | 0.064385 | +0.012264 | 1.2353 |
| src/compiler/watch.ts:554:1 | toFileName | 0.067573 | 0.079797 | +0.012224 | 1.1809 |
| src/compiler/transformers/classFields.ts:309:1 | UntransformedPrivateIdentifierInfo | 0.036841 | 0.049058 | +0.012217 | 1.3316 |
| src/compiler/path.ts:278:1 | getDirectoryPath | 0.036429 | 0.048613 | +0.012184 | 1.3344 |
| src/compiler/core.ts:2109:1 | getUILocale | 0.035470 | 0.047633 | +0.012163 | 1.3429 |
| src/compiler/visitorPublic.ts:420:1 | addDefaultValueAssignmentsIfNeeded | 0.035654 | 0.047795 | +0.012141 | 1.3405 |
| src/compiler/emitter.ts:6351:1 | OrdinalParentheizerRuleSelector | 0.034809 | 0.046930 | +0.012121 | 1.3482 |
| src/compiler/moduleNameResolver.ts:1397:1 | resolveLibrary | 0.036980 | 0.049087 | +0.012107 | 1.3274 |
| src/compiler/scanner.ts:354:1 |  | 0.037974 | 0.050022 | +0.012048 | 1.3173 |
| src/compiler/utilitiesPublic.ts:461:1 | textChangeRangeNewSpan | 0.036379 | 0.048395 | +0.012016 | 1.3303 |
| src/compiler/factory/emitNode.ts:208:1 | getSyntheticTrailingComments | 0.040182 | 0.052177 | +0.011995 | 1.2985 |
| src/compiler/transformers/declarations.ts:241:1 |  | 0.036230 | 0.048216 | +0.011987 | 1.3309 |
| src/compiler/transformers/declarations.ts:249:1 |  | 0.037526 | 0.049441 | +0.011914 | 1.3175 |
| src/compiler/types.ts:2119:1 | ConstructorDeclaration | 0.044098 | 0.056008 | +0.011910 | 1.2701 |
| src/compiler/factory/utilities.ts:188:1 | createMemberAccessForPropertyName | 0.200592 | 0.212498 | +0.011905 | 1.0594 |
| src/compiler/checker.ts:1410:1 |  | 0.037902 | 0.049775 | +0.011873 | 1.3132 |
| src/compiler/utilities.ts:6570:1 | getDeclarationEmitExtensionForPath | 0.036144 | 0.048012 | +0.011868 | 1.3283 |
| src/compiler/emitter.ts:6336:1 | getOpeningBracket | 0.036135 | 0.047951 | +0.011815 | 1.3270 |
| src/compiler/utilities.ts:8708:1 | formatMessage | 0.067543 | 0.079327 | +0.011784 | 1.1745 |
| src/compiler/moduleNameResolver.ts:942:1 | PerModuleNameCache | 0.036971 | 0.048749 | +0.011777 | 1.3186 |
| src/compiler/scanner.ts:349:1 |  | 0.040285 | 0.052017 | +0.011732 | 1.2912 |
| src/compiler/utilities.ts:5362:1 | isStringOrNumericLiteralLike | 0.271563 | 0.283291 | +0.011728 | 1.0432 |
| src/compiler/utilities.ts:2540:1 | getErrorSpanForNode | 0.149771 | 0.161477 | +0.011706 | 1.0782 |
| src/compiler/sys.ts:121:1 | createPollingIntervalBasedLevels | 0.036256 | 0.047953 | +0.011698 | 1.3226 |
| src/compiler/utilitiesPublic.ts:1675:1 | isFunctionLike | 0.045198 | 0.056788 | +0.011590 | 1.2564 |
| src/compiler/emitter.ts:6355:1 | ParenthesizerRule | 0.037003 | 0.048590 | +0.011588 | 1.3132 |
| src/compiler/types.ts:3740:1 | ImportAttributeName | 0.036642 | 0.048222 | +0.011580 | 1.3160 |
| src/compiler/transformers/es2021.ts:1:1 |  | 0.036007 | 0.047579 | +0.011573 | 1.3214 |
| src/compiler/executeCommandLine.ts:90:1 |  | 0.036944 | 0.048478 | +0.011534 | 1.3122 |
| src/compiler/types.ts:2165:1 | IndexSignatureDeclaration | 0.044182 | 0.055694 | +0.011513 | 1.2606 |
| src/compiler/factory/nodeTests.ts:882:1 | isImportSpecifier | 0.042861 | 0.054325 | +0.011464 | 1.2675 |
| src/compiler/types.ts:4182:1 | FlowNodeBase | 0.036800 | 0.048234 | +0.011434 | 1.3107 |
| src/compiler/path.ts:250:1 | getRootLength | 0.037347 | 0.048705 | +0.011358 | 1.3041 |
| src/compiler/core.ts:2114:1 | setUILocale | 0.035689 | 0.046965 | +0.011276 | 1.3159 |
| src/compiler/types.ts:2184:1 | TypeNode | 0.043999 | 0.055263 | +0.011264 | 1.2560 |
| src/compiler/types.ts:3410:1 | WhileStatement | 0.036698 | 0.047951 | +0.011253 | 1.3066 |
| src/compiler/utilitiesPublic.ts:2203:1 | isModuleOrEnumDeclaration | 0.042708 | 0.053951 | +0.011243 | 1.2632 |
| src/compiler/program.ts:1345:1 | getImpliedNodeFormatForFileWorker | 0.038380 | 0.049491 | +0.011111 | 1.2895 |
| src/compiler/types.ts:3120:1 | TaggedTemplateExpression | 0.037040 | 0.048125 | +0.011084 | 1.2993 |
| src/compiler/commandLineParser.ts:2764:1 | getNameOfCompilerOptionValue | 0.037213 | 0.048289 | +0.011076 | 1.2976 |
| src/compiler/types.ts:8659:1 | OuterExpression | 0.035462 | 0.046469 | +0.011007 | 1.3104 |
| src/compiler/utilitiesPublic.ts:2504:1 | isCaseOrDefaultClause | 0.041325 | 0.052314 | +0.010989 | 1.2659 |
| src/compiler/program.ts:1499:1 | createProgram | 0.066535 | 0.077521 | +0.010986 | 1.1651 |
| src/compiler/utilities.ts:8586:1 | formatStringFromArgs | 0.039325 | 0.050282 | +0.010957 | 1.2786 |
| src/compiler/types.ts:6470:1 | IntrinsicType | 0.040349 | 0.051304 | +0.010955 | 1.2715 |
| src/compiler/utilities.ts:2385:1 | getTextOfPropertyName | 0.042193 | 0.053114 | +0.010921 | 1.2588 |
| src/compiler/factory/utilities.ts:348:1 | createExpressionForPropertyName | 0.191514 | 0.202414 | +0.010900 | 1.0569 |
| src/compiler/_namespaces/ts.ts:15:1 |  | 0.037620 | 0.048435 | +0.010815 | 1.2875 |
| src/compiler/semver.ts:158:1 | comparePrereleaseIdentifiers | 0.038475 | 0.049278 | +0.010803 | 1.2808 |
| src/compiler/utilitiesPublic.ts:762:1 | getOriginalNode | 0.036438 | 0.047214 | +0.010776 | 1.2957 |
| src/compiler/sys.ts:1463:1 |  | 0.037933 | 0.048678 | +0.010745 | 1.2833 |
| src/compiler/path.ts:110:1 | pathIsBareSpecifier | 0.036257 | 0.046995 | +0.010739 | 1.2962 |
| src/compiler/types.ts:4071:1 | JSDocTypedefTag | 0.034877 | 0.045571 | +0.010694 | 1.3066 |
| src/compiler/utilities.ts:4402:1 | shouldRewriteModuleSpecifier | 0.037549 | 0.048145 | +0.010596 | 1.2822 |
| src/compiler/binder.ts:422:1 | getModuleInstanceStateForAliasTarget | 0.251908 | 0.262479 | +0.010571 | 1.0420 |
| src/compiler/types.ts:3113:1 | NewExpression | 0.036878 | 0.047392 | +0.010515 | 1.2851 |
| src/compiler/commandLineParser.ts:4023:1 | isExcludedFile | 0.148033 | 0.158516 | +0.010483 | 1.0708 |
| src/compiler/resolutionCache.ts:229:1 | DirectoryWatchesOfFailedLookup | 0.036274 | 0.046743 | +0.010469 | 1.2886 |
| src/compiler/checker.ts:1157:1 | ReferenceHint | 0.036379 | 0.046765 | +0.010386 | 1.2855 |
| src/compiler/types.ts:3263:1 | JsxOpeningFragment | 0.036371 | 0.046751 | +0.010380 | 1.2854 |
| src/compiler/commandLineParser.ts:2214:1 | ConfigFileDiagnosticsReporter | 0.038105 | 0.048465 | +0.010360 | 1.2719 |
| src/compiler/program.ts:1370:1 |  | 0.042439 | 0.052788 | +0.010350 | 1.2439 |
| src/compiler/types.ts:4459:1 | ReadonlyPragmaContext | 0.037811 | 0.048127 | +0.010316 | 1.2728 |
| src/compiler/utilitiesPublic.ts:870:1 | symbolName | 0.260708 | 0.270937 | +0.010229 | 1.0392 |
| src/compiler/programDiagnostics.ts:1:1 |  | 0.037804 | 0.048013 | +0.010209 | 1.2700 |
| src/compiler/factory/nodeTests.ts:516:1 | isConditionalTypeNode | 0.041621 | 0.051807 | +0.010186 | 1.2447 |
| src/compiler/performanceCore.ts:20:1 | Performance | 0.037674 | 0.047807 | +0.010134 | 1.2690 |
| src/compiler/types.ts:4629:1 | ProjectReferenceFile | 0.038275 | 0.048381 | +0.010106 | 1.2640 |
| src/compiler/path.ts:115:1 | hasExtension | 0.038260 | 0.048362 | +0.010102 | 1.2640 |
| src/compiler/transformer.ts:192:1 | getDeclarationTransformers | 0.037688 | 0.047783 | +0.010095 | 1.2679 |
| src/compiler/utilities.ts:6764:1 | getSetAccessorValueParameter | 0.035069 | 0.045132 | +0.010063 | 1.2869 |
| src/compiler/transformer.ts:127:1 | getScriptTransformers | 0.039367 | 0.049413 | +0.010047 | 1.2552 |
| src/compiler/utilities.ts:6782:1 | getThisParameter | 0.035921 | 0.045927 | +0.010006 | 1.2786 |
| src/compiler/checker.ts:1205:1 | IterationTypesResolver | 0.035227 | 0.045227 | +0.010000 | 1.2839 |
| src/compiler/path.ts:120:1 | fileExtensionIs | 0.038628 | 0.048613 | +0.009985 | 1.2585 |
| src/compiler/path.ts:947:1 | comparePaths | 0.035666 | 0.045643 | +0.009977 | 1.2797 |
| src/compiler/executeCommandLine.ts:902:1 | createReportErrorSummary | 0.070986 | 0.080960 | +0.009974 | 1.1405 |
| src/compiler/core.ts:2589:80 |  | 0.034539 | 0.044508 | +0.009969 | 1.2886 |
| src/compiler/utilities.ts:4212:1 | getElementOrPropertyAccessArgumentExpressionOrName | 0.187620 | 0.197559 | +0.009939 | 1.0530 |
| src/compiler/factory/nodeTests.ts:902:1 | isModuleExportName | 0.041743 | 0.051664 | +0.009921 | 1.2377 |
| src/compiler/utilities.ts:3078:1 | isObjectLiteralOrClassExpressionMethodOrAccessor | 0.046961 | 0.056863 | +0.009902 | 1.2108 |
| src/compiler/factory/nodeTests.ts:566:1 | isArrayBindingPattern | 0.040472 | 0.050370 | +0.009898 | 1.2446 |
| src/compiler/transformers/classFields.ts:333:1 | ClassFacts | 0.037885 | 0.047771 | +0.009886 | 1.2610 |
| src/compiler/utilities.ts:2422:1 | createDiagnosticForNodeArray | 0.038700 | 0.048583 | +0.009883 | 1.2554 |
| src/compiler/factory/emitHelpers.ts:866:1 |  | 0.063608 | 0.073486 | +0.009878 | 1.1553 |
| src/compiler/utilities.ts:4086:1 | getRightMostAssignedExpression | 0.156817 | 0.166643 | +0.009827 | 1.0627 |
| src/compiler/moduleSpecifiers.ts:172:1 | ModuleSpecifierPreferences | 0.036426 | 0.046252 | +0.009827 | 1.2698 |
| src/compiler/utilities.ts:7563:1 | isPrototypeAccess | 0.152565 | 0.162377 | +0.009812 | 1.0643 |
| src/compiler/builderPublic.ts:201:1 | createEmitAndSemanticDiagnosticsBuilderProgram | 0.037201 | 0.046946 | +0.009745 | 1.2620 |
| src/compiler/utilities.ts:6488:1 | hostGetCanonicalFileName | 0.037750 | 0.047483 | +0.009733 | 1.2578 |
| src/compiler/transformer.ts:120:1 | getTransformers | 0.035598 | 0.045314 | +0.009715 | 1.2729 |
| src/compiler/path.ts:125:1 | fileExtensionIsOneOf | 0.036968 | 0.046643 | +0.009675 | 1.2617 |
| src/compiler/types.ts:932:1 | PredicateSemantics | 0.038324 | 0.047967 | +0.009643 | 1.2516 |
| src/compiler/program.ts:1478:1 | createCreateProgramOptions | 0.036270 | 0.045906 | +0.009636 | 1.2657 |
| src/compiler/factory/emitHelpers.ts:101:1 | EmitHelperFactory | 0.035425 | 0.045031 | +0.009605 | 1.2711 |
| src/compiler/types.ts:8644:1 | OuterExpressionKinds | 0.037181 | 0.046725 | +0.009545 | 1.2567 |
| src/compiler/factory/nodeTests.ts:484:1 | isTypeLiteralNode | 0.040935 | 0.050477 | +0.009541 | 1.2331 |
| src/compiler/factory/nodeTests.ts:464:1 | isTypePredicateNode | 0.043245 | 0.052779 | +0.009534 | 1.2205 |
| src/compiler/types.ts:2657:1 | AssignmentExpression | 0.039995 | 0.049463 | +0.009468 | 1.2367 |
| src/compiler/core.ts:2198:1 | levenshteinWithMax | 0.034780 | 0.044244 | +0.009464 | 1.2721 |
| src/compiler/transformers/es2018.ts:122:1 | HierarchyFacts | 0.039762 | 0.049215 | +0.009453 | 1.2378 |
| src/compiler/types.ts:2655:1 | AssignmentOperatorToken | 0.038642 | 0.048083 | +0.009441 | 1.2443 |
| src/compiler/moduleNameResolver.ts:1687:1 | NodeResolutionFeatures | 0.037965 | 0.047404 | +0.009439 | 1.2486 |
| src/compiler/types.ts:4320:1 | SourceFile | 0.037044 | 0.046473 | +0.009429 | 1.2545 |
| src/compiler/core.ts:1886:1 | notImplemented | 0.036196 | 0.045607 | +0.009411 | 1.2600 |
| src/compiler/utilities.ts:3014:1 | isVariableDeclarationInVariableStatement | 0.038357 | 0.047738 | +0.009381 | 1.2446 |
| src/compiler/transformer.ts:78:1 | getModuleTransformer | 0.060535 | 0.069912 | +0.009376 | 1.1549 |
| src/compiler/types.ts:1613:1 | QuestionQuestionEqualsToken | 0.034424 | 0.043792 | +0.009368 | 1.2721 |
| src/compiler/factory/utilities.ts:910:1 | getTargetOfBindingOrAssignmentElement | 0.261949 | 0.271231 | +0.009282 | 1.0354 |
| src/compiler/program.ts:711:1 | formatCodeSpan | 0.289446 | 0.298694 | +0.009248 | 1.0319 |
| src/compiler/utilitiesPublic.ts:2196:1 | isNamedImportBindings | 0.042018 | 0.051260 | +0.009242 | 1.2200 |
| src/compiler/types.ts:3415:1 | ForInitializer | 0.036265 | 0.045493 | +0.009228 | 1.2545 |
| src/compiler/utilities.ts:5367:1 | isSignedNumericLiteral | 0.274292 | 0.283508 | +0.009216 | 1.0336 |
| src/compiler/types.ts:3080:1 | BindableStaticAccessExpression | 0.035906 | 0.045104 | +0.009197 | 1.2562 |
| src/compiler/transformer.ts:218:1 | wrapScriptTransformerFactory | 0.039877 | 0.049065 | +0.009187 | 1.2304 |
| src/compiler/utilities.ts:9790:1 | FileMatcherPatterns | 0.039823 | 0.049006 | +0.009183 | 1.2306 |
| src/compiler/binder.ts:330:1 | getModuleInstanceState | 0.042800 | 0.051976 | +0.009176 | 1.2144 |
| src/compiler/_namespaces/ts.ts:77:1 |  | 0.034841 | 0.044013 | +0.009173 | 1.2633 |
| src/compiler/utilities.ts:1980:1 | getLiteralText | 0.174336 | 0.183497 | +0.009161 | 1.0525 |
| src/compiler/scanner.ts:520:1 | computeLineOfPosition | 0.078031 | 0.087179 | +0.009148 | 1.1172 |
| src/compiler/utilities.ts:2996:1 | isVariableLike | 0.040133 | 0.049277 | +0.009144 | 1.2278 |
| src/compiler/types.ts:33:1 | ReadonlyTextRange | 0.036152 | 0.045266 | +0.009113 | 1.2521 |
| src/compiler/tracing.ts:371:1 |  | 0.035257 | 0.044336 | +0.009079 | 1.2575 |
| src/compiler/utilities.ts:8803:1 | compareRelatedInformation | 0.039417 | 0.048443 | +0.009026 | 1.2290 |
| src/compiler/semver.ts:425:1 | testComparator | 0.037930 | 0.046947 | +0.009017 | 1.2377 |
| src/compiler/factory/nodeTests.ts:397:1 | isComputedPropertyName | 0.041514 | 0.050497 | +0.008983 | 1.2164 |
| src/compiler/transformer.ts:117:1 |  | 0.060448 | 0.069415 | +0.008967 | 1.1483 |
| src/compiler/path.ts:28:1 |  | 0.037234 | 0.046169 | +0.008934 | 1.2399 |
| src/compiler/factory/nodeTests.ts:866:1 | isImportAttribute | 0.041373 | 0.050286 | +0.008913 | 1.2154 |
| src/compiler/builder.ts:2395:1 | getBuildInfoFileVersionMap | 0.110688 | 0.119598 | +0.008910 | 1.0805 |
| src/compiler/moduleNameResolver.ts:2112:1 | loadFileNameFromPackageJsonField | 0.037170 | 0.045943 | +0.008773 | 1.2360 |
| src/compiler/moduleSpecifiers.ts:798:1 | getAllRuntimeDependencies | 0.040861 | 0.049624 | +0.008763 | 1.2145 |
| src/compiler/scanner.ts:508:1 | computeLineAndCharacterOfPosition | 0.036302 | 0.045037 | +0.008735 | 1.2406 |
| src/compiler/utilities.ts:2519:1 | scanTokenAtPosition | 0.110151 | 0.118873 | +0.008723 | 1.0792 |
| src/compiler/executeCommandLine.ts:756:1 | executeCommandLine | 0.373132 | 0.381850 | +0.008718 | 1.0234 |
| src/compiler/utilities.ts:11140:1 | isJsxAttributeName | 0.042930 | 0.051621 | +0.008691 | 1.2025 |
| src/compiler/utilities.ts:11147:1 | getEscapedTextOfJsxNamespacedName | 0.041127 | 0.049809 | +0.008682 | 1.2111 |
| src/compiler/utilitiesPublic.ts:644:1 | getCombinedFlags | 0.222848 | 0.231527 | +0.008679 | 1.0389 |
| src/compiler/factory/utilities.ts:1076:1 | getElementsOfBindingOrAssignmentPattern | 0.035730 | 0.044392 | +0.008662 | 1.2424 |
| src/compiler/builderPublic.ts:176:1 | createSemanticDiagnosticsBuilderProgram | 0.040321 | 0.048978 | +0.008657 | 1.2147 |
| src/compiler/core.ts:1971:1 | compareComparableValues | 0.036508 | 0.045160 | +0.008652 | 1.2370 |
| src/compiler/moduleNameResolver.ts:1750:1 | tryResolveJSModuleWorker | 0.038834 | 0.047479 | +0.008645 | 1.2226 |
| src/compiler/builder.ts:2218:1 | addToAffectedFilesPendingEmit | 0.060248 | 0.068833 | +0.008586 | 1.1425 |
| src/compiler/visitorPublic.ts:395:1 | visitParameterList | 0.035682 | 0.044264 | +0.008582 | 1.2405 |
| src/compiler/utilities.ts:9829:1 | matchFiles | 0.106929 | 0.115507 | +0.008578 | 1.0802 |
| src/compiler/sys.ts:89:1 | WatchedFile | 0.035837 | 0.044408 | +0.008571 | 1.2392 |
| src/compiler/parser.ts:499:1 | isImportMeta | 0.237884 | 0.246417 | +0.008533 | 1.0359 |
| src/compiler/moduleNameResolver.ts:1401:1 | resolveModuleNameFromCache | 0.071442 | 0.079972 | +0.008530 | 1.1194 |
| src/compiler/core.ts:2079:1 | getStringComparer | 0.035282 | 0.043812 | +0.008530 | 1.2418 |
| src/compiler/utilities.ts:4242:1 | getAssignmentDeclarationPropertyAccessKind | 0.041926 | 0.050442 | +0.008516 | 1.2031 |
| src/compiler/factory/nodeTests.ts:1058:1 | isJSDocNullableType | 0.040671 | 0.049153 | +0.008482 | 1.2086 |
| src/compiler/moduleSpecifiers.ts:133:1 |  | 0.037715 | 0.046160 | +0.008445 | 1.2239 |
| src/compiler/types.ts:3173:1 | MetaProperty | 0.036689 | 0.045079 | +0.008390 | 1.2287 |
| src/compiler/types.ts:4560:1 | ResolvedConfigFileName | 0.037065 | 0.045435 | +0.008370 | 1.2258 |
| src/compiler/checker.ts:1443:1 |  | 0.059552 | 0.067862 | +0.008311 | 1.1396 |
| src/compiler/scanner.ts:503:1 | getLineStarts | 0.036572 | 0.044820 | +0.008248 | 1.2255 |
| src/compiler/hostErrors.ts:7:1 | errorMessage | 0.059743 | 0.067960 | +0.008217 | 1.1375 |
| src/compiler/types.ts:3162:1 | NonNullExpression | 0.036115 | 0.044307 | +0.008193 | 1.2269 |
| src/compiler/factory/nodeTests.ts:981:1 | isDefaultClause | 0.042311 | 0.050464 | +0.008153 | 1.1927 |
| src/compiler/resolutionCache.ts:532:1 | resolveModuleNameUsingGlobalCache | 0.115563 | 0.123713 | +0.008150 | 1.0705 |
| src/compiler/types.ts:7820:1 | AlternateModeDiagnostics | 0.034416 | 0.042559 | +0.008143 | 1.2366 |
| src/compiler/utilities.ts:7455:1 | tryGetClassImplementingOrExtendingExpressionWithTypeArguments | 0.261268 | 0.269395 | +0.008127 | 1.0311 |
| src/compiler/factory/nodeTests.ts:967:1 | isJsxExpression | 0.043391 | 0.051504 | +0.008113 | 1.1870 |
| src/compiler/core.ts:2591:1 | isNodeLikeSystem | 0.068181 | 0.076286 | +0.008104 | 1.1189 |
| src/compiler/utilities.ts:6609:1 | getPathsBasePath | 0.035449 | 0.043533 | +0.008084 | 1.2280 |
| src/compiler/resolutionCache.ts:238:1 | DirPathToWatcherOfPackageDirWatcher | 0.036703 | 0.044784 | +0.008082 | 1.2202 |
| src/compiler/types.ts:10251:1 |  | 0.070066 | 0.078123 | +0.008057 | 1.1150 |
| src/compiler/tracing.ts:31:1 |  | 0.034758 | 0.042793 | +0.008036 | 1.2312 |
| src/compiler/core.ts:2257:1 | removeSuffix | 0.034771 | 0.042788 | +0.008017 | 1.2306 |
| src/compiler/factory/nodeTests.ts:640:1 | isPostfixUnaryExpression | 0.042074 | 0.050062 | +0.007987 | 1.1898 |
| src/compiler/factory/nodeTests.ts:562:1 | isObjectBindingPattern | 0.040613 | 0.048564 | +0.007951 | 1.1958 |
| src/compiler/moduleNameResolver.ts:3091:1 | loadModuleFromSpecificNodeModulesDirectory | 0.109417 | 0.117326 | +0.007909 | 1.0723 |
| src/compiler/program.ts:4843:1 | shouldTransformImportCallWorker | 0.039320 | 0.047227 | +0.007907 | 1.2011 |
| src/compiler/types.ts:4317:1 | ResolutionMode | 0.036588 | 0.044449 | +0.007860 | 1.2148 |
| src/compiler/semver.ts:447:1 | formatAlternative | 0.036961 | 0.044809 | +0.007848 | 1.2123 |
| src/compiler/transformers/classFields.ts:297:1 | PrivateEnvironmentData | 0.035646 | 0.043486 | +0.007840 | 1.2199 |
| src/compiler/checker.ts:1170:1 |  | 0.035775 | 0.043611 | +0.007836 | 1.2190 |
| src/compiler/types.ts:2883:1 | TemplateMiddle | 0.037732 | 0.045567 | +0.007835 | 1.2076 |
| src/compiler/types.ts:4624:1 | ProjectReferenceFileKind | 0.037252 | 0.045064 | +0.007812 | 1.2097 |
| src/compiler/builder.ts:1181:1 | isIncrementalBuildInfo | 0.037437 | 0.045246 | +0.007809 | 1.2086 |
| src/compiler/factory/utilities.ts:1181:1 | isMultiplicativeOperatorOrHigher | 0.357409 | 0.365197 | +0.007789 | 1.0218 |
| src/compiler/types.ts:1610:1 | EqualsToken | 0.035946 | 0.043679 | +0.007734 | 1.2151 |
| src/compiler/utilities.ts:904:1 | packageIdIsEqual | 0.037079 | 0.044811 | +0.007733 | 1.2085 |
| src/compiler/semver.ts:274:1 |  | 0.037796 | 0.045485 | +0.007690 | 1.2034 |
| src/compiler/utilities.ts:8240:1 | MutateMapOptions | 0.036727 | 0.044407 | +0.007680 | 1.2091 |
| src/compiler/utilities.ts:4491:1 | getSourceOfAssignment | 0.232221 | 0.239875 | +0.007654 | 1.0330 |
| src/compiler/types.ts:1611:1 | AmpersandAmpersandEqualsToken | 0.037267 | 0.044894 | +0.007627 | 1.2047 |
| src/compiler/utilitiesPublic.ts:2489:1 | isJsxOpeningLikeElement | 0.040677 | 0.048300 | +0.007623 | 1.1874 |
| src/compiler/factory/nodeTests.ts:939:1 | isJsxClosingElement | 0.040397 | 0.048015 | +0.007618 | 1.1886 |
| src/compiler/core.ts:2271:1 | removeMinAndVersionNumbers | 0.035491 | 0.043075 | +0.007585 | 1.2137 |
| src/compiler/core.ts:2430:1 | startsWith | 0.036495 | 0.044055 | +0.007561 | 1.2072 |
| src/compiler/utilities.ts:9820:1 | getRegexFromPattern | 0.041053 | 0.048611 | +0.007558 | 1.1841 |
| src/compiler/sys.ts:1458:1 | DirectoryWatcher | 0.041673 | 0.049220 | +0.007546 | 1.1811 |
| src/compiler/factory/nodeTests.ts:403:1 | isTypeParameterDeclaration | 0.040771 | 0.048315 | +0.007544 | 1.1850 |
| src/compiler/program.ts:1514:1 | createProgram | 0.068607 | 0.076150 | +0.007543 | 1.1099 |
| src/compiler/utilitiesPublic.ts:1024:1 | getDecorators | 0.043415 | 0.050940 | +0.007524 | 1.1733 |
| src/compiler/utilities.ts:10387:1 | rangeOfTypeParameters | 0.037503 | 0.044996 | +0.007493 | 1.1998 |
| src/compiler/utilities.ts:4657:1 | getJSDocCommentsAndTags | 0.223312 | 0.230787 | +0.007475 | 1.0335 |
| src/compiler/factory/utilities.ts:1015:1 | tryGetPropertyNameOfBindingOrAssignmentElement | 0.210448 | 0.217912 | +0.007464 | 1.0355 |
| src/compiler/hostErrors.ts:2:1 | errorCode | 0.084523 | 0.091984 | +0.007462 | 1.0883 |
| src/compiler/transformers/classFields.ts:320:1 | ClassLexicalEnvironment | 0.038562 | 0.046021 | +0.007459 | 1.1934 |
| src/compiler/utilities.ts:8091:1 | AccessKind | 0.036484 | 0.043905 | +0.007421 | 1.2034 |
| src/compiler/utilities.ts:834:1 | createModuleNotFoundChain | 0.110474 | 0.117859 | +0.007385 | 1.0668 |
| src/compiler/factory/nodeTests.ts:848:1 | isImportTypeAssertionContainer | 0.040424 | 0.047730 | +0.007306 | 1.1807 |
| src/compiler/utilities.ts:12067:1 | hasInferredType | 0.158447 | 0.165749 | +0.007301 | 1.0461 |
| src/compiler/types.ts:3104:1 | ImportCall | 0.035989 | 0.043281 | +0.007292 | 1.2026 |
| src/compiler/types.ts:3167:1 | NonNullChain | 0.036118 | 0.043359 | +0.007241 | 1.2005 |
| src/compiler/transformer.ts:104:1 | TransformationState | 0.035158 | 0.042390 | +0.007232 | 1.2057 |
| src/compiler/builderPublic.ts:239:1 | createAbstractBuilder | 0.147273 | 0.154493 | +0.007219 | 1.0490 |
| src/compiler/types.ts:2674:1 | BindingOrAssignmentElement | 0.036965 | 0.044077 | +0.007111 | 1.1924 |
| src/compiler/_namespaces/ts.ts:78:1 |  | 0.036141 | 0.043247 | +0.007106 | 1.1966 |
| src/compiler/commandLineParser.ts:2774:1 | serializeCompilerOptions | 0.036628 | 0.043648 | +0.007020 | 1.1917 |
| src/compiler/sourcemap.ts:368:1 |  | 0.039781 | 0.046794 | +0.007013 | 1.1763 |
| src/compiler/factory/nodeTests.ts:556:1 | isTemplateLiteralTypeNode | 0.040140 | 0.047141 | +0.007001 | 1.1744 |
| src/compiler/core.ts:1988:1 | compareValues | 0.035872 | 0.042868 | +0.006996 | 1.1950 |
| src/compiler/builder.ts:1673:1 | createBuilderProgram | 0.036150 | 0.043136 | +0.006986 | 1.1933 |
| src/compiler/executeCommandLine.ts:563:1 | executeCommandLineWorker | 0.731395 | 0.738368 | +0.006973 | 1.0095 |
| src/compiler/checker.ts:1435:1 | IntrinsicTypeKind | 0.034798 | 0.041771 | +0.006973 | 1.2004 |
| src/compiler/path.ts:151:1 | getFileUrlVolumeSeparatorEnd | 0.036247 | 0.043186 | +0.006939 | 1.1914 |
| src/compiler/types.ts:4489:1 | CommentDirectiveType | 0.036466 | 0.043393 | +0.006927 | 1.1900 |
| src/compiler/types.ts:1615:1 | EqualsGreaterThanToken | 0.035545 | 0.042458 | +0.006913 | 1.1945 |
| src/compiler/types.ts:4199:1 | FlowStart | 0.038471 | 0.045342 | +0.006871 | 1.1786 |
| src/compiler/resolutionCache.ts:184:1 | ResolutionCacheHost | 0.036505 | 0.043369 | +0.006863 | 1.1880 |
| src/compiler/parser.ts:10703:1 | getNamedArgRegEx | 0.036446 | 0.043296 | +0.006849 | 1.1879 |
| src/compiler/utilitiesPublic.ts:1364:1 | isGetOrSetAccessorDeclaration | 0.042438 | 0.049271 | +0.006833 | 1.1610 |
| src/compiler/types.ts:668:1 | ModifierSyntaxKind | 0.034937 | 0.041765 | +0.006828 | 1.1954 |
| src/compiler/parser.ts:503:1 | ForEachChildFunction | 0.035643 | 0.042455 | +0.006812 | 1.1911 |
| src/compiler/builder.ts:1186:1 | NonIncrementalBuildInfo | 0.038157 | 0.044963 | +0.006805 | 1.1783 |
| src/compiler/transformers/namedEvaluation.ts:56:1 | getAssignedNameOfIdentifier | 0.061881 | 0.068683 | +0.006802 | 1.1099 |
| src/compiler/factory/utilities.ts:690:1 | hasRecordedExternalHelpers | 0.038909 | 0.045695 | +0.006787 | 1.1744 |
| src/compiler/types.ts:3140:1 | AsExpression | 0.036826 | 0.043588 | +0.006763 | 1.1836 |
| src/compiler/utilities.ts:9978:1 |  | 0.034952 | 0.041701 | +0.006749 | 1.1931 |
| src/compiler/types.ts:4129:1 | JSDocSatisfiesTag | 0.035887 | 0.042624 | +0.006737 | 1.1877 |
| src/compiler/utilities.ts:2124:1 | isGlobalScopeAugmentation | 0.038922 | 0.045634 | +0.006712 | 1.1725 |
| src/compiler/builder.ts:281:1 | getBuilderFileEmit | 0.047664 | 0.054363 | +0.006699 | 1.1406 |
| src/compiler/factory/utilities.ts:1240:1 | isBitwiseOperatorOrHigher | 0.103881 | 0.110563 | +0.006682 | 1.0643 |
| src/compiler/commandLineParser.ts:2224:1 | ParseConfigFileHost | 0.038579 | 0.045260 | +0.006681 | 1.1732 |
| src/compiler/resolutionCache.ts:310:1 | canWatchDirectoryOrFile | 0.036471 | 0.043145 | +0.006674 | 1.1830 |
| src/compiler/core.ts:2133:1 | compareStringsCaseSensitiveUI | 0.035522 | 0.042186 | +0.006664 | 1.1876 |
| src/compiler/transformers/utilities.ts:113:1 | ExternalModuleInfo | 0.036630 | 0.043282 | +0.006652 | 1.1816 |
| src/compiler/types.ts:8615:1 | EmitHost | 0.036879 | 0.043531 | +0.006652 | 1.1804 |
| src/compiler/utilities.ts:7475:1 | isAssignmentExpression | 0.276697 | 0.283332 | +0.006635 | 1.0240 |
| src/compiler/moduleSpecifiers.ts:879:1 | tryGetModuleNameFromAmbientModule | 0.043258 | 0.049866 | +0.006609 | 1.1528 |
| src/compiler/builder.ts:1663:1 | createBuilderProgram | 0.038358 | 0.044929 | +0.006571 | 1.1713 |
| src/compiler/moduleNameResolver.ts:1392:1 | getOptionsForLibraryResolution | 0.037202 | 0.043697 | +0.006496 | 1.1746 |
| src/compiler/types.ts:4502:1 | JsonSourceFile | 0.037179 | 0.043669 | +0.006490 | 1.1746 |
| src/compiler/types.ts:1621:1 | KeywordToken | 0.035088 | 0.041559 | +0.006471 | 1.1844 |
| src/compiler/moduleNameResolver.ts:3169:1 | tryLoadModuleUsingPaths | 0.095337 | 0.101806 | +0.006469 | 1.0678 |
| src/compiler/sys.ts:108:1 |  | 0.034395 | 0.040848 | +0.006454 | 1.1876 |
| src/compiler/transformer.ts:111:1 | SyntaxKindFeatureFlags | 0.034443 | 0.040864 | +0.006420 | 1.1864 |
| src/compiler/types.ts:4471:1 | PragmaContext | 0.036642 | 0.043036 | +0.006394 | 1.1745 |
| src/compiler/factory/nodeTests.ts:844:1 | isImportClause | 0.042246 | 0.048618 | +0.006373 | 1.1509 |
| src/compiler/factory/emitNode.ts:147:1 | getTokenSourceMapRange | 0.040229 | 0.046592 | +0.006363 | 1.1582 |
| src/compiler/moduleNameResolver.ts:2751:1 | getLoadModuleFromTargetExportOrImport | 0.039493 | 0.045853 | +0.006360 | 1.1610 |
| src/compiler/moduleNameResolver.ts:2619:1 | loadModuleFromExports | 4.556455 | 4.562806 | +0.006351 | 1.0014 |
| src/compiler/utilities.ts:7955:1 | isNodeArrayMultiLine | 0.037394 | 0.043735 | +0.006341 | 1.1696 |
| src/compiler/transformers/esnext.ts:64:1 | UsingKind | 0.039372 | 0.045692 | +0.006320 | 1.1605 |
| src/compiler/watch.ts:416:1 | getMatchedFileSpec | 0.070194 | 0.076500 | +0.006306 | 1.0898 |
| src/compiler/utilities.ts:2428:1 | createDiagnosticForNodeInSourceFile | 0.044702 | 0.051000 | +0.006298 | 1.1409 |
| src/compiler/_namespaces/ts.ts:32:1 |  | 0.035697 | 0.041987 | +0.006290 | 1.1762 |
| src/compiler/transformers/module/system.ts:1:1 |  | 0.037564 | 0.043847 | +0.006283 | 1.1673 |
| src/compiler/factory/nodeTests.ts:853:1 | isAssertClause | 0.041369 | 0.047613 | +0.006244 | 1.1509 |
| src/compiler/commandLineParser.ts:4247:1 | removeWildcardFilesWithLowerPriorityExtension | 0.040839 | 0.047074 | +0.006235 | 1.1527 |
| src/compiler/visitorPublic.ts:328:1 | visitArrayWorker | 0.035782 | 0.042002 | +0.006220 | 1.1738 |
| src/compiler/transformers/es2018.ts:115:1 | ESNextSubstitutionFlags | 0.037071 | 0.043285 | +0.006215 | 1.1676 |
| src/compiler/utilities.ts:1:1 |  | 0.036647 | 0.042829 | +0.006182 | 1.1687 |
| src/compiler/core.ts:2392:1 | patternText | 0.105862 | 0.112044 | +0.006182 | 1.0584 |
| src/compiler/types.ts:1609:1 | ColonToken | 0.036052 | 0.042203 | +0.006151 | 1.1706 |
| src/compiler/factory/utilitiesPublic.ts:14:1 | canHaveModifiers | 0.044194 | 0.050311 | +0.006117 | 1.1384 |
| src/compiler/checker.ts:1199:1 | IterationTypeKind | 0.036481 | 0.042586 | +0.006105 | 1.1673 |
| src/compiler/factory/nodeFactory.ts:492:1 | createNodeFactory | 0.257730 | 0.263821 | +0.006091 | 1.0236 |
| src/compiler/utilities.ts:7992:1 | getPreviousNonWhitespacePosition | 0.035810 | 0.041873 | +0.006063 | 1.1693 |
| src/compiler/factory/nodeTests.ts:796:1 | isVariableDeclarationList | 0.041198 | 0.047241 | +0.006043 | 1.1467 |
| src/compiler/utilities.ts:7827:1 | getNewLineCharacter | 0.069952 | 0.075986 | +0.006035 | 1.0863 |
| src/compiler/commandLineParser.ts:4045:1 | invalidDotDotAfterRecursiveWildcard | 0.035222 | 0.041256 | +0.006034 | 1.1713 |
| src/compiler/utilities.ts:4499:1 | getSourceOfDefaultedAssignment | 0.232513 | 0.238547 | +0.006034 | 1.0260 |
| src/compiler/factory/nodeTests.ts:576:1 | isArrayLiteralExpression | 0.043238 | 0.049272 | +0.006033 | 1.1395 |
| src/compiler/transformer.ts:232:1 | noEmitNotification | 0.040559 | 0.046578 | +0.006019 | 1.1484 |
| src/compiler/binder.ts:3815:1 | getContainerFlags | 0.218999 | 0.225003 | +0.006004 | 1.0274 |
| src/compiler/utilities.ts:4658:1 | getJSDocCommentsAndTags | 0.228192 | 0.234195 | +0.006002 | 1.0263 |
| src/compiler/utilities.ts:5387:1 | isDynamicName | 0.180383 | 0.186380 | +0.005996 | 1.0332 |
| src/compiler/visitorPublic.ts:382:1 | visitLexicalEnvironment | 0.035685 | 0.041663 | +0.005977 | 1.1675 |
| src/compiler/emitter.ts:6359:1 | EmitFunction | 0.035867 | 0.041782 | +0.005915 | 1.1649 |
| src/compiler/moduleSpecifiers.ts:582:1 | getLocalModuleSpecifier | 0.043572 | 0.049428 | +0.005857 | 1.1344 |
| src/compiler/types.ts:7710:1 | ParsedCommandLine | 0.035429 | 0.041272 | +0.005843 | 1.1649 |
| src/compiler/sys.ts:911:1 | createFsWatchCallbackForFileWatcherCallback | 0.038399 | 0.044240 | +0.005841 | 1.1521 |
| src/compiler/types.ts:8635:1 | PropertyDescriptorAttributes | 0.035802 | 0.041627 | +0.005825 | 1.1627 |
| src/compiler/types.ts:3100:1 | SuperCall | 0.036425 | 0.042245 | +0.005820 | 1.1598 |
| src/compiler/core.ts:1966:1 | equateStringsCaseSensitive | 0.036781 | 0.042569 | +0.005788 | 1.1574 |
| src/compiler/utilities.ts:8784:1 | compareDiagnostics | 0.035295 | 0.041081 | +0.005786 | 1.1639 |
| src/compiler/factory/nodeTests.ts:548:1 | isImportTypeNode | 0.040447 | 0.046224 | +0.005777 | 1.1428 |
| src/compiler/types.ts:4529:1 | JsonObjectExpressionStatement | 0.036182 | 0.041935 | +0.005753 | 1.1590 |
| src/compiler/moduleNameResolver.ts:1974:1 | realPath | 0.104579 | 0.110323 | +0.005744 | 1.0549 |
| src/compiler/moduleNameResolver.ts:2379:1 | PackageJsonInfo | 0.037626 | 0.043321 | +0.005695 | 1.1514 |
| src/compiler/builder.ts:271:1 | toBuilderProgramStateWithDefinedProgram | 0.043877 | 0.049563 | +0.005686 | 1.1296 |
| src/compiler/moduleNameResolver.ts:1712:1 | node16ModuleNameResolver | 0.038990 | 0.044663 | +0.005673 | 1.1455 |
| src/compiler/types.ts:3890:1 | JSDocNameReference | 0.033897 | 0.039550 | +0.005654 | 1.1668 |
| src/compiler/utilitiesPublic.ts:1522:1 | isTemplateLiteralToken | 0.067983 | 0.073626 | +0.005644 | 1.0830 |
| src/compiler/checker.ts:1364:1 | IntersectionState | 0.034995 | 0.040625 | +0.005630 | 1.1609 |
| src/compiler/transformers/esnext.ts:788:1 | getUsingKind | 0.228151 | 0.233773 | +0.005621 | 1.0246 |
| src/compiler/factory/nodeTests.ts:1156:1 | isJSDocThisTag | 0.042030 | 0.047637 | +0.005608 | 1.1334 |
| src/compiler/moduleSpecifiers.ts:583:1 | getLocalModuleSpecifier | 0.042792 | 0.048398 | +0.005606 | 1.1310 |
| src/compiler/factory/emitHelpers.ts:1451:1 |  | 0.062204 | 0.067789 | +0.005585 | 1.0898 |
| src/compiler/program.ts:5098:1 | filterSemanticDiagnostics | 0.038631 | 0.044200 | +0.005569 | 1.1442 |
| src/compiler/moduleNameResolver.ts:934:1 | PackageJsonInfoCache | 0.036704 | 0.042265 | +0.005560 | 1.1515 |
| src/compiler/types.ts:7813:1 | CommandLineOptionOfCustomType | 0.034285 | 0.039832 | +0.005547 | 1.1618 |
| src/compiler/parser.ts:9927:1 |  | 0.037303 | 0.042841 | +0.005538 | 1.1484 |
| src/compiler/commandLineParser.ts:2210:1 | DiagnosticReporter | 0.036828 | 0.042344 | +0.005516 | 1.1498 |
| src/compiler/factory/nodeTests.ts:792:1 | isVariableDeclaration | 0.040280 | 0.045773 | +0.005493 | 1.1364 |
| src/compiler/emitter.ts:506:1 | getOutputPathsForBundle | 0.036128 | 0.041616 | +0.005489 | 1.1519 |
| src/compiler/types.ts:8015:1 | MinimalResolutionCacheHost | 0.034547 | 0.040024 | +0.005477 | 1.1585 |
| src/compiler/utilitiesPublic.ts:1727:1 | isFunctionOrModuleBlock | 0.226106 | 0.231573 | +0.005467 | 1.0242 |
| src/compiler/types.ts:10085:1 | ErrorOutputContainer | 0.037206 | 0.042669 | +0.005464 | 1.1469 |
| src/compiler/factory/nodeTests.ts:820:1 | isModuleDeclaration | 0.040732 | 0.046194 | +0.005462 | 1.1341 |
| src/compiler/utilities.ts:12097:1 |  | 0.036525 | 0.041978 | +0.005453 | 1.1493 |
| src/compiler/transformers/classFields.ts:1:1 |  | 0.036504 | 0.041940 | +0.005436 | 1.1489 |
| src/compiler/watch.ts:898:1 | reportUnrecoverableDiagnostic | 0.071457 | 0.076883 | +0.005426 | 1.0759 |
| src/compiler/transformers/classFields.ts:244:1 | PrivateIdentifierInfoBase | 0.037635 | 0.043057 | +0.005423 | 1.1441 |
| src/compiler/transformers/namedEvaluation.ts:101:1 | createClassNamedEvaluationHelperBlock | 0.035270 | 0.040690 | +0.005420 | 1.1537 |
| src/compiler/sys.ts:129:1 |  | 0.061688 | 0.067105 | +0.005417 | 1.0878 |
| src/compiler/_namespaces/ts.ts:47:1 |  | 0.039246 | 0.044656 | +0.005409 | 1.1378 |
| src/compiler/utilities.ts:3549:1 | childIsDecorated | 0.158914 | 0.164302 | +0.005388 | 1.0339 |
| src/compiler/moduleNameResolver.ts:1334:1 | createModuleResolutionCache | 0.218099 | 0.223480 | +0.005381 | 1.0247 |
| src/compiler/moduleNameResolver.ts:2708:1 | loadModuleFromExportsOrImports | 0.042799 | 0.048180 | +0.005380 | 1.1257 |
| src/compiler/utilities.ts:8236:1 | MutateMapWithNewSetOptions | 0.036625 | 0.041990 | +0.005365 | 1.1465 |
| src/compiler/commandLineParser.ts:4060:1 | matchesExclude | 0.034770 | 0.040133 | +0.005363 | 1.1542 |
| src/compiler/types.ts:4494:1 | Bundle | 0.036951 | 0.042296 | +0.005345 | 1.1447 |
| src/compiler/utilitiesPublic.ts:1399:1 | isExpressionOfOptionalChainRoot | 0.268912 | 0.274252 | +0.005340 | 1.0199 |
| src/compiler/builder.ts:265:1 | BuilderProgramStateWithDefinedProgram | 0.035770 | 0.041102 | +0.005332 | 1.1490 |
| src/compiler/types.ts:3128:1 | InstanceofExpression | 0.036382 | 0.041693 | +0.005310 | 1.1460 |
| src/compiler/utilities.ts:11135:1 | getTextOfJsxAttributeName | 0.191351 | 0.196655 | +0.005304 | 1.0277 |
| src/compiler/factory/nodeTests.ts:955:1 | isJsxAttribute | 0.042919 | 0.048190 | +0.005272 | 1.1228 |
| src/compiler/types.ts:2890:1 | TemplateTail | 0.037736 | 0.042973 | +0.005237 | 1.1388 |
| src/compiler/watch.ts:446:1 | fileIncludeReasonToDiagnostics | 0.099805 | 0.105024 | +0.005219 | 1.0523 |
| src/compiler/types.ts:4520:1 | JsonObjectExpression | 0.035430 | 0.040646 | +0.005216 | 1.1472 |
| src/compiler/factory/nodeTests.ts:292:1 | isAsteriskToken | 0.040750 | 0.045966 | +0.005215 | 1.1280 |
| src/compiler/core.ts:1951:1 | equateStringsCaseInsensitive | 0.037130 | 0.042342 | +0.005212 | 1.1404 |
| src/compiler/types.ts:3018:1 | SuperElementAccessExpression | 0.035496 | 0.040698 | +0.005202 | 1.1465 |
| src/compiler/utilities.ts:12204:1 | getNodeAtPosition | 0.118979 | 0.124172 | +0.005193 | 1.0436 |
| src/compiler/types.ts:1617:1 | MinusToken | 0.036991 | 0.042169 | +0.005178 | 1.1400 |
| src/compiler/symbolWalker.ts:30:1 | createGetSymbolWalker | 0.140748 | 0.145921 | +0.005173 | 1.0368 |
| src/compiler/moduleSpecifiers.ts:809:1 | getAllModulePathsWorker | 0.043179 | 0.048352 | +0.005173 | 1.1198 |
| src/compiler/types.ts:4589:1 | OperationCanceledException | 0.035944 | 0.041112 | +0.005168 | 1.1438 |
| src/compiler/types.ts:1573:1 | HasContainerFlags | 0.035291 | 0.040455 | +0.005164 | 1.1463 |
| src/compiler/builder.ts:2246:1 | toProgramEmitPending | 0.035843 | 0.040998 | +0.005155 | 1.1438 |
| src/compiler/utilities.ts:3912:1 | isSingleOrDoubleQuote | 0.036588 | 0.041726 | +0.005137 | 1.1404 |
| src/compiler/builderPublic.ts:168:1 | createSemanticDiagnosticsBuilderProgram | 0.037768 | 0.042885 | +0.005116 | 1.1355 |
| src/compiler/types.ts:3033:1 | CallChain | 0.035042 | 0.040147 | +0.005104 | 1.1457 |
| src/compiler/types.ts:8309:1 | SourceMapRange | 0.036311 | 0.041408 | +0.005097 | 1.1404 |
| src/compiler/program.ts:4851:1 | getEmitModuleFormatOfFileWorker | 0.037300 | 0.042387 | +0.005087 | 1.1364 |
| src/compiler/types.ts:7748:1 | ModuleImportResult | 0.036070 | 0.041125 | +0.005056 | 1.1402 |
| src/compiler/parser.ts:441:1 |  | 0.035808 | 0.040863 | +0.005055 | 1.1412 |
| src/compiler/types.ts:8027:1 | ResolvedModule | 0.033914 | 0.038968 | +0.005054 | 1.1490 |
| src/compiler/utilitiesPublic.ts:423:1 | normalizeSpans | 0.036532 | 0.041580 | +0.005049 | 1.1382 |
| src/compiler/checker.ts:1412:1 | DeclarationMeaning | 0.036145 | 0.041189 | +0.005044 | 1.1395 |
| src/compiler/utilities.ts:829:1 | getResolvedTypeReferenceDirectiveFromResolution | 0.036190 | 0.041231 | +0.005041 | 1.1393 |
| src/compiler/factory/emitHelpers.ts:42:1 | PrivateIdentifierKind | 0.036415 | 0.041452 | +0.005037 | 1.1383 |
| src/compiler/types.ts:4540:1 | ParseConfigHost | 0.037623 | 0.042611 | +0.004989 | 1.1326 |
| src/compiler/utilities.ts:914:1 | packageIdToString | 0.036404 | 0.041390 | +0.004986 | 1.1370 |
| src/compiler/types.ts:3095:1 | BindablePropertyAssignmentExpression | 0.035836 | 0.040797 | +0.004960 | 1.1384 |
| src/compiler/types.ts:9610:1 | TransformationContext | 0.037340 | 0.042300 | +0.004960 | 1.1328 |
| src/compiler/factory/emitHelpers.ts:970:1 |  | 0.063352 | 0.068311 | +0.004959 | 1.0783 |
| src/compiler/sys.ts:132:1 |  | 0.037790 | 0.042734 | +0.004944 | 1.1308 |
| src/compiler/core.ts:1:1 |  | 0.037791 | 0.042722 | +0.004931 | 1.1305 |
| src/compiler/checker.ts:1155:1 |  | 0.043578 | 0.048502 | +0.004924 | 1.1130 |
| src/compiler/transformers/utilities.ts:703:1 | getAllDecoratorsOfClassElement | 0.044329 | 0.049247 | +0.004918 | 1.1110 |
| src/compiler/utilities.ts:3040:1 | introducesArgumentsExoticObject | 0.040402 | 0.045307 | +0.004905 | 1.1214 |
| src/compiler/watch.ts:688:1 |  | 0.060166 | 0.065068 | +0.004903 | 1.0815 |
| src/compiler/utilities.ts:10610:1 | isIdentifierInNonEmittingHeritageClause | 0.043156 | 0.048055 | +0.004899 | 1.1135 |
| src/compiler/utilities.ts:3152:1 | getContainingFunctionOrClassStaticBlock | 0.042216 | 0.047087 | +0.004872 | 1.1154 |
| src/compiler/types.ts:4563:1 | ResolvedRefAndOutputDts | 0.036345 | 0.041206 | +0.004861 | 1.1337 |
| src/compiler/factory/nodeTests.ts:552:1 | isTemplateLiteralTypeSpan | 0.040039 | 0.044897 | +0.004858 | 1.1213 |
| src/compiler/types.ts:3247:1 | JsxSelfClosingElement | 0.063468 | 0.068325 | +0.004857 | 1.0765 |
| src/compiler/sys.ts:105:1 | HostWatchDirectory | 0.034349 | 0.039166 | +0.004816 | 1.1402 |
| src/compiler/program.ts:1471:1 | shouldProgramCreateNewSourceFiles | 0.073613 | 0.078426 | +0.004813 | 1.0654 |
| src/compiler/types.ts:3812:1 | ExportSpecifier | 0.035604 | 0.040390 | +0.004786 | 1.1344 |
| src/compiler/utilities.ts:9987:1 |  | 0.036302 | 0.041085 | +0.004783 | 1.1318 |
| src/compiler/utilities.ts:6615:1 | EmitFileNames | 0.034987 | 0.039768 | +0.004782 | 1.1367 |
| src/compiler/factory/nodeTests.ts:1034:1 | isJSDocMemberName | 0.041186 | 0.045964 | +0.004778 | 1.1160 |
| src/compiler/types.ts:2977:1 | PrivateIdentifierPropertyAccessExpression | 0.036068 | 0.040846 | +0.004778 | 1.1325 |
| src/compiler/types.ts:7807:1 | CommandLineOptionOfBooleanType | 0.034689 | 0.039464 | +0.004775 | 1.1377 |
| src/compiler/types.ts:942:1 | Node | 0.036681 | 0.041453 | +0.004772 | 1.1301 |
| src/compiler/utilities.ts:7425:1 | isLogicalOrCoalescingBinaryOperator | 0.041390 | 0.046156 | +0.004766 | 1.1152 |
| src/compiler/moduleNameResolver.ts:2072:1 | loadModuleFromFile | 0.038694 | 0.043433 | +0.004739 | 1.1225 |
| src/compiler/moduleNameResolver.ts:2743:1 | hasOneAsterisk | 0.036622 | 0.041360 | +0.004739 | 1.1294 |
| src/compiler/utilitiesPublic.ts:1503:1 | isLiteralExpressionOfObject | 0.042270 | 0.047008 | +0.004738 | 1.1121 |
| src/compiler/checker.ts:1221:1 | WideningKind | 0.037432 | 0.042156 | +0.004724 | 1.1262 |
| src/compiler/sys.ts:1968:1 | setSys | 0.037642 | 0.042350 | +0.004708 | 1.1251 |
| src/compiler/types.ts:3132:1 | CallLikeExpression | 0.035887 | 0.040576 | +0.004689 | 1.1307 |
| src/compiler/watch.ts:686:1 | WatchType | 0.034627 | 0.039289 | +0.004662 | 1.1347 |
| src/compiler/transformers/classFields.ts:230:1 | ClassPropertySubstitutionFlags | 0.037045 | 0.041697 | +0.004653 | 1.1256 |
| src/compiler/utilities.ts:6582:1 | getPossibleOriginalInputExtensionForExtension | 0.036260 | 0.040882 | +0.004621 | 1.1274 |
| src/compiler/transformer.ts:227:1 | noEmitSubstitution | 0.043380 | 0.047978 | +0.004598 | 1.1060 |
| src/compiler/moduleNameResolver.ts:516:1 | getOriginalAndResolvedFileName | 0.037009 | 0.041596 | +0.004587 | 1.1239 |
| src/compiler/utilities.ts:7568:1 | isRightSideOfQualifiedNameOrPropertyAccess | 0.114259 | 0.118841 | +0.004581 | 1.0401 |
| src/compiler/transformer.ts:222:1 | wrapDeclarationTransformerFactory | 0.041903 | 0.046481 | +0.004578 | 1.1093 |
| src/compiler/moduleNameResolver.ts:2997:1 | loadModuleFromNearestNodeModulesDirectory | 0.037190 | 0.041767 | +0.004577 | 1.1231 |
| src/compiler/utilities.ts:8600:1 | maybeSetLocalizedDiagnosticMessages | 0.038379 | 0.042930 | +0.004551 | 1.1186 |
| src/compiler/factory/nodeTests.ts:544:1 | isLiteralTypeNode | 0.039152 | 0.043683 | +0.004532 | 1.1157 |
| src/compiler/sys.ts:540:1 | onWatchedFileStat | 0.072570 | 0.077090 | +0.004520 | 1.0623 |
| src/compiler/core.ts:1974:1 | compareComparableValues | 0.036688 | 0.041200 | +0.004512 | 1.1230 |
| src/compiler/utilitiesPublic.ts:2442:1 | isStatementOrBlock | 0.045219 | 0.049725 | +0.004506 | 1.0996 |
| src/compiler/builderState.ts:62:1 | BuilderState | 0.036993 | 0.041496 | +0.004503 | 1.1217 |
| src/compiler/utilities.ts:7540:1 | tryGetPropertyAccessOrIdentifierToString | 0.194730 | 0.199174 | +0.004444 | 1.0228 |
| src/compiler/core.ts:17:1 |  | 0.036147 | 0.040588 | +0.004441 | 1.1229 |
| src/compiler/utilities.ts:6634:1 | getSourceFilesToEmit | 0.071149 | 0.075579 | +0.004430 | 1.0623 |
| src/compiler/builder.ts:126:1 | ReusableDiagnosticMessageChain | 0.042301 | 0.046710 | +0.004409 | 1.1042 |
| src/compiler/types.ts:3090:1 | BindableStaticPropertyAssignmentExpression | 0.035191 | 0.039552 | +0.004361 | 1.1239 |
| src/compiler/types.ts:3763:1 | NamespaceExport | 0.034963 | 0.039322 | +0.004359 | 1.1247 |
| src/compiler/watch.ts:673:1 | createWatchHost | 0.035596 | 0.039917 | +0.004321 | 1.1214 |
| src/compiler/moduleNameResolver.ts:2407:1 | getVersionPathsOfPackageJsonInfo | 4.373954 | 4.378274 | +0.004320 | 1.0010 |
| src/compiler/utilities.ts:2468:1 | createDiagnosticForFileFromMessageChain | 0.038542 | 0.042859 | +0.004317 | 1.1120 |
| src/compiler/moduleSpecifiers.ts:276:1 | getModuleSpecifier | 0.038213 | 0.042529 | +0.004317 | 1.1130 |
| src/compiler/moduleSpecifiers.ts:1467:1 | prefersTsExtension | 0.039844 | 0.044149 | +0.004304 | 1.1080 |
| src/compiler/utilities.ts:10883:1 | hasContextSensitiveParameters | 0.040785 | 0.045087 | +0.004303 | 1.1055 |
| src/compiler/utilities.ts:2088:1 | isNonGlobalAmbientModule | 0.241642 | 0.245940 | +0.004298 | 1.0178 |
| src/compiler/factory/utilities.ts:617:1 | isJSDocTypeAssertion | 0.258899 | 0.263179 | +0.004280 | 1.0165 |
| src/compiler/commandLineParser.ts:3501:1 | parseOwnConfigOfJson | 0.043236 | 0.047514 | +0.004279 | 1.0990 |
| src/compiler/watch.ts:788:1 | getSourceFileVersionAsHashFromText | 0.074485 | 0.078755 | +0.004270 | 1.0573 |
| src/compiler/types.ts:2876:1 | TemplateHead | 0.037442 | 0.041692 | +0.004250 | 1.1135 |
| src/compiler/_namespaces/ts.ts:30:1 |  | 0.037167 | 0.041393 | +0.004226 | 1.1137 |
| src/compiler/transformers/destructuring.ts:446:1 | flattenArrayBindingOrAssignmentPattern | 0.067831 | 0.072051 | +0.004220 | 1.0622 |
| src/compiler/core.ts:1921:1 | AssertionLevel | 0.036514 | 0.040733 | +0.004219 | 1.1156 |
| src/compiler/transformers/generators.ts:1:1 |  | 0.037644 | 0.041863 | +0.004219 | 1.1121 |
| src/compiler/utilities.ts:615:1 |  | 0.035150 | 0.039366 | +0.004215 | 1.1199 |
| src/compiler/factory/nodeTests.ts:492:1 | isTupleTypeNode | 0.040084 | 0.044290 | +0.004206 | 1.1049 |
| src/compiler/transformers/namedEvaluation.ts:120:1 | ClassNamedEvaluationHelperBlock | 0.034272 | 0.038469 | +0.004196 | 1.1224 |
| src/compiler/types.ts:4079:1 | JSDocCallbackTag | 0.033783 | 0.037978 | +0.004195 | 1.1242 |
| src/compiler/checker.ts:1422:1 | DeclarationSpaces | 0.035951 | 0.040145 | +0.004195 | 1.1167 |
| src/compiler/factory/utilities.ts:1222:1 | isEqualityOperator | 0.043927 | 0.048107 | +0.004180 | 1.0952 |
| src/compiler/moduleSpecifiers.ts:1454:1 | getRelativePathIfInSameVolume | 0.038262 | 0.042436 | +0.004175 | 1.1091 |
| src/compiler/types.ts:7834:1 | TsConfigOnlyOption | 0.035349 | 0.039514 | +0.004165 | 1.1178 |
| src/compiler/utilitiesPublic.ts:2552:1 | hasJSDocNodes | 0.103350 | 0.107515 | +0.004165 | 1.0403 |
| src/compiler/utilities.ts:2503:1 | getCanonicalDiagnostic | 0.036129 | 0.040290 | +0.004161 | 1.1152 |
| src/compiler/utilities.ts:10233:1 | tryParsePatterns | 0.036573 | 0.040734 | +0.004161 | 1.1138 |
| src/compiler/utilities.ts:2696:1 | isLet | 0.044441 | 0.048600 | +0.004159 | 1.0936 |
| src/compiler/utilities.ts:7575:1 | isRightSideOfAccessExpression | 0.247840 | 0.251979 | +0.004138 | 1.0167 |
| src/compiler/types.ts:1645:1 | OverrideKeyword | 0.036995 | 0.041131 | +0.004136 | 1.1118 |
| src/compiler/types.ts:7596:1 | TypeAcquisition | 0.035363 | 0.039497 | +0.004134 | 1.1169 |
| src/compiler/commandLineParser.ts:2336:1 |  | 0.043517 | 0.047633 | +0.004116 | 1.0946 |
| src/compiler/executeCommandLine.ts:1276:1 | statisticValue | 0.034928 | 0.039040 | +0.004112 | 1.1177 |
| src/compiler/factory/nodeTests.ts:836:1 | isImportEqualsDeclaration | 0.041294 | 0.045396 | +0.004102 | 1.0993 |
| src/compiler/types.ts:4509:1 | TsConfigSourceFile | 0.037201 | 0.041300 | +0.004100 | 1.1102 |
| src/compiler/moduleNameResolver.ts:2360:1 | getTemporaryModuleResolutionState | 0.040337 | 0.044400 | +0.004063 | 1.1007 |
| src/compiler/utilities.ts:11124:1 | tryGetJSDocSatisfiesTypeNode | 0.050160 | 0.054217 | +0.004057 | 1.0809 |
| src/compiler/utilities.ts:7364:1 | modifierToFlag | 0.037689 | 0.041731 | +0.004042 | 1.1073 |
| src/compiler/transformers/classFields.ts:275:1 | PrivateIdentifierMethodInfo | 0.037069 | 0.041106 | +0.004037 | 1.1089 |
| src/compiler/factory/utilities.ts:1008:1 | getPropertyNameOfBindingOrAssignmentElement | 0.319937 | 0.323972 | +0.004035 | 1.0126 |
| src/compiler/types.ts:4580:1 | WriteFileCallback | 0.035376 | 0.039402 | +0.004026 | 1.1138 |
| src/compiler/utilities.ts:10400:1 | skipTypeChecking | 0.035968 | 0.039984 | +0.004016 | 1.1116 |
| src/compiler/tsbuildPublic.ts:1194:1 | InvalidateProjectCreateInfo | 0.036646 | 0.040648 | +0.004002 | 1.1092 |
| src/compiler/types.ts:3180:1 | ImportMetaProperty | 0.040968 | 0.044970 | +0.004002 | 1.0977 |
| src/compiler/resolutionCache.ts:442:1 | getDirectoryOfFailedLookupWatch | 0.037177 | 0.041169 | +0.003992 | 1.1074 |
| src/compiler/types.ts:3231:1 | JsxNamespacedName | 0.069654 | 0.073637 | +0.003982 | 1.0572 |
| src/compiler/utilities.ts:7246:1 | getSelectedSyntacticModifierFlags | 0.041784 | 0.045758 | +0.003974 | 1.0951 |
| src/compiler/types.ts:10119:1 | ListFormat | 0.036455 | 0.040426 | +0.003971 | 1.1089 |
| src/compiler/executeCommandLine.ts:1:1 |  | 0.035547 | 0.039514 | +0.003967 | 1.1116 |
| src/compiler/sys.ts:495:1 | SingleFileWatcher | 0.034965 | 0.038930 | +0.003965 | 1.1134 |
| src/compiler/factory/nodeTests.ts:808:1 | isInterfaceDeclaration | 0.041281 | 0.045226 | +0.003945 | 1.0956 |
| src/compiler/program.ts:766:1 | formatLocation | 0.179123 | 0.183067 | +0.003945 | 1.0220 |
| src/compiler/sys.ts:565:1 |  | 0.037031 | 0.040975 | +0.003944 | 1.1065 |
| src/compiler/tsbuildPublic.ts:925:1 | BuildStep | 0.037185 | 0.041116 | +0.003931 | 1.1057 |
| src/compiler/parser.ts:1311:1 | gatherPossibleChildren | 0.045065 | 0.048979 | +0.003914 | 1.0869 |
| src/compiler/commandLineParser.ts:4175:1 | toCanonicalKey | 0.035548 | 0.039450 | +0.003902 | 1.1098 |
| src/compiler/binder.ts:321:1 | ActiveLabel | 0.036594 | 0.040486 | +0.003892 | 1.1064 |
| src/compiler/types.ts:3401:1 | IterationStatement | 0.035006 | 0.038896 | +0.003890 | 1.1111 |
| src/compiler/utilitiesPublic.ts:1999:1 | isLeftHandSideExpressionKind | 0.037533 | 0.041421 | +0.003889 | 1.1036 |
| src/compiler/factory/emitNode.ts:183:1 | getCommentRange | 0.037867 | 0.041753 | +0.003887 | 1.1026 |
| src/compiler/utilities.ts:6509:1 | getExternalModuleNameFromDeclaration | 0.308372 | 0.312258 | +0.003886 | 1.0126 |
| src/compiler/emitter.ts:548:1 | getOutputPathWithoutChangingExt | 0.035836 | 0.039710 | +0.003874 | 1.1081 |
| src/compiler/moduleNameResolver.ts:813:1 | getAutomaticTypeDirectiveNames | 0.036190 | 0.040062 | +0.003872 | 1.1070 |
| src/compiler/_namespaces/ts.ts:14:1 |  | 0.039407 | 0.043278 | +0.003871 | 1.0982 |
| src/compiler/debug.ts:113:1 | Debug | 0.035457 | 0.039322 | +0.003866 | 1.1090 |
| src/compiler/checker.ts:1354:1 | SignatureCheckMode | 0.035216 | 0.039065 | +0.003850 | 1.1093 |
| src/compiler/core.ts:1973:1 | compareComparableValues | 0.036275 | 0.040125 | +0.003849 | 1.1061 |
| src/compiler/types.ts:4098:1 | JSDocSignature | 0.035127 | 0.038970 | +0.003843 | 1.1094 |
| src/compiler/types.ts:8543:1 | ExternalEmitHelpers | 0.036076 | 0.039909 | +0.003833 | 1.1062 |
| src/compiler/utilities.ts:4925:1 | isAssignmentTarget | 0.043274 | 0.047093 | +0.003819 | 1.0883 |
| src/compiler/path.ts:310:1 | getDirectoryPath | 0.035995 | 0.039808 | +0.003813 | 1.1059 |
| src/compiler/core.ts:2247:1 | endsWith | 0.036098 | 0.039909 | +0.003811 | 1.1056 |
| src/compiler/types.ts:1266:1 | HasType | 0.042624 | 0.046425 | +0.003801 | 1.0892 |
| src/compiler/utilities.ts:9925:1 | getIncludeBasePath | 0.035506 | 0.039307 | +0.003801 | 1.1070 |
| src/compiler/watch.ts:362:1 | explainIfFileIsRedirectAndImpliedFormat | 0.036562 | 0.040362 | +0.003800 | 1.1039 |
| src/compiler/utilities.ts:11184:1 | isExpandoPropertyDeclaration | 0.238158 | 0.241956 | +0.003799 | 1.0160 |
| src/compiler/factory/nodeTests.ts:648:1 | isConditionalExpression | 0.041656 | 0.045434 | +0.003778 | 1.0907 |
| src/compiler/types.ts:3085:1 | BindableAccessExpression | 0.035959 | 0.039735 | +0.003775 | 1.1050 |
| src/compiler/types.ts:7826:1 | DidYouMeanOptionsDiagnostics | 0.035521 | 0.039288 | +0.003767 | 1.1061 |
| src/compiler/factory/nodeTests.ts:1046:1 | isJSDocLinkPlain | 0.040775 | 0.044537 | +0.003762 | 1.0923 |
| src/compiler/checker.ts:1377:1 | MappedTypeModifiers | 0.036360 | 0.040102 | +0.003742 | 1.1029 |
| src/compiler/types.ts:8044:1 | ResolvedModuleFull | 0.035153 | 0.038892 | +0.003739 | 1.1064 |
| src/compiler/builder.ts:1668:1 | createBuilderProgram | 0.036848 | 0.040586 | +0.003738 | 1.1015 |
| src/compiler/visitorPublic.ts:520:18 | visitFunctionBody | 0.035121 | 0.038859 | +0.003738 | 1.1064 |
| src/compiler/checker.ts:1451:1 |  | 0.035261 | 0.038999 | +0.003737 | 1.1060 |
| src/compiler/sys.ts:134:1 | setCustomPollingValues | 0.073195 | 0.076924 | +0.003729 | 1.0509 |
| src/compiler/utilitiesPublic.ts:1752:1 | isAutoAccessorPropertyDeclaration | 0.263808 | 0.267526 | +0.003718 | 1.0141 |
| src/compiler/path.ts:945:1 | comparePaths | 0.036093 | 0.039809 | +0.003717 | 1.1030 |
| src/compiler/utilities.ts:12410:1 | suppressLeadingTrivia | 0.049677 | 0.053390 | +0.003712 | 1.0747 |
| src/compiler/types.ts:3777:1 | ExportDeclaration | 0.033781 | 0.037489 | +0.003708 | 1.1098 |
| src/compiler/utilitiesPublic.ts:1359:1 | isMemberName | 0.043736 | 0.047444 | +0.003708 | 1.0848 |
| src/compiler/core.ts:2074:1 | compareStringsCaseSensitive | 0.037464 | 0.041160 | +0.003696 | 1.0987 |
| src/compiler/types.ts:26:1 | MatchingKeys | 0.034945 | 0.038627 | +0.003683 | 1.1054 |
| src/compiler/sys.ts:966:1 | CreateSystemWatchFunctions | 0.036199 | 0.039881 | +0.003682 | 1.1017 |
| src/compiler/executeCommandLine.ts:745:1 | isBuildCommand | 0.035826 | 0.039504 | +0.003678 | 1.1027 |
| src/compiler/types.ts:3796:1 | NamedExports | 0.034371 | 0.038032 | +0.003661 | 1.1065 |
| src/compiler/utilities.ts:7592:1 | isRightSideOfInstanceofExpression | 0.292371 | 0.296022 | +0.003651 | 1.0125 |
| src/compiler/commandLineParser.ts:2332:1 |  | 0.040091 | 0.043729 | +0.003639 | 1.0908 |
| src/compiler/watch.ts:740:1 | WatchFactoryWithLog | 0.037935 | 0.041568 | +0.003633 | 1.0958 |
| src/compiler/resolutionCache.ts:400:1 | getDirectoryToWatchFromFailedLookupLocationDirectory | 0.038700 | 0.042333 | +0.003633 | 1.0939 |
| src/compiler/checker.ts:1340:1 | CheckMode | 0.036116 | 0.039732 | +0.003616 | 1.1001 |
| src/compiler/factory/nodeTests.ts:874:1 | isNamespaceExport | 0.041938 | 0.045552 | +0.003614 | 1.0862 |
| src/compiler/utilities.ts:7403:1 | isBinaryLogicalOperator | 0.037496 | 0.041108 | +0.003612 | 1.0963 |
| src/compiler/utilities.ts:11152:1 | getTextOfJsxNamespacedName | 0.040384 | 0.043992 | +0.003608 | 1.0893 |
| src/compiler/types.ts:4087:1 | JSDocOverloadTag | 0.034818 | 0.038420 | +0.003602 | 1.1035 |
| src/compiler/types.ts:2897:1 | PseudoLiteralToken | 0.036021 | 0.039622 | +0.003601 | 1.1000 |
| src/compiler/types.ts:7848:1 | CommandLineOption | 0.035235 | 0.038833 | +0.003597 | 1.1021 |
| src/compiler/transformer.ts:1:1 |  | 0.035477 | 0.039074 | +0.003597 | 1.1014 |
| src/compiler/checker.ts:1455:1 | NodeLinks | 0.036264 | 0.039859 | +0.003595 | 1.0991 |
| src/compiler/watch.ts:670:1 |  | 0.034329 | 0.037916 | +0.003587 | 1.1045 |
| src/compiler/utilities.ts:6504:1 | getCanonicalAbsolutePath | 0.107470 | 0.111051 | +0.003581 | 1.0333 |
| src/compiler/utilities.ts:9983:1 |  | 0.035846 | 0.039421 | +0.003575 | 1.0997 |
| src/compiler/sys.ts:892:1 | FsWatchCallback | 0.035939 | 0.039505 | +0.003566 | 1.0992 |
| src/compiler/moduleSpecifiers.ts:562:1 | Info | 0.036194 | 0.039741 | +0.003547 | 1.0980 |
| src/compiler/utilities.ts:3163:1 | ThisContainer | 0.037111 | 0.040649 | +0.003539 | 1.0954 |
| src/compiler/utilities.ts:10771:1 | expressionResultIsUnused | 0.050000 | 0.053538 | +0.003539 | 1.0708 |
| src/compiler/parser.ts:422:1 |  | 0.035278 | 0.038811 | +0.003533 | 1.1002 |
| src/compiler/resolutionCache.ts:468:1 | getDirectoryToWatchFailedLookupLocationFromTypeRoot | 0.041539 | 0.045061 | +0.003522 | 1.0848 |
| src/compiler/types.ts:3749:1 | ImportAttributes | 0.034848 | 0.038369 | +0.003521 | 1.1010 |
| src/compiler/factory/emitHelpers.ts:1414:1 |  | 0.061472 | 0.064992 | +0.003519 | 1.0573 |
| src/compiler/moduleNameResolver.ts:2189:1 | tryFile | 0.038156 | 0.041659 | +0.003503 | 1.0918 |
| src/compiler/types.ts:3014:1 | ElementAccessChainRoot | 0.036176 | 0.039676 | +0.003501 | 1.0968 |
| src/compiler/types.ts:4568:1 | ResolvedRefAndSource | 0.036303 | 0.039786 | +0.003483 | 1.0959 |
| src/compiler/checker.ts:1230:1 | TypeFacts | 0.044589 | 0.048067 | +0.003478 | 1.0780 |
| src/compiler/utilities.ts:2491:1 | createDiagnosticForRange | 0.036296 | 0.039772 | +0.003476 | 1.0958 |
| src/compiler/types.ts:3757:1 | NamespaceImport | 0.036809 | 0.040279 | +0.003469 | 1.0943 |
| src/compiler/factory/nodeTests.ts:339:1 | isAsyncModifier | 0.041441 | 0.044887 | +0.003446 | 1.0832 |
| src/compiler/utilities.ts:7292:1 | getSyntacticModifierFlags | 0.043569 | 0.047009 | +0.003441 | 1.0790 |
| src/compiler/tracing.ts:25:1 |  | 0.034481 | 0.037920 | +0.003438 | 1.0997 |
| src/compiler/types.ts:1644:1 | OutKeyword | 0.037120 | 0.040554 | +0.003435 | 1.0925 |
| src/compiler/transformers/destructuring.ts:244:1 | flattenDestructuringBinding | 0.063452 | 0.066884 | +0.003432 | 1.0541 |
| src/compiler/types.ts:7801:1 | CommandLineOptionOfNumberType | 0.035771 | 0.039201 | +0.003430 | 1.0959 |
| src/compiler/utilities.ts:2701:1 | isSuperCall | 0.159796 | 0.163223 | +0.003427 | 1.0214 |
| src/compiler/moduleSpecifiers.ts:255:1 | updateModuleSpecifier | 0.038252 | 0.041658 | +0.003406 | 1.0890 |
| src/compiler/utilities.ts:7276:1 | getEffectiveModifierFlags | 0.045141 | 0.048545 | +0.003403 | 1.0754 |
| src/compiler/builderPublic.ts:46:1 | BuilderProgram | 0.038626 | 0.042024 | +0.003398 | 1.0880 |
| src/compiler/factory/nodeTests.ts:828:1 | isCaseBlock | 0.040528 | 0.043920 | +0.003392 | 1.0837 |
| src/compiler/utilities.ts:12155:1 |  | 0.036037 | 0.039424 | +0.003387 | 1.0940 |
| src/compiler/transformers/es2018.ts:1:1 |  | 0.036751 | 0.040135 | +0.003385 | 1.0921 |
| src/compiler/types.ts:2648:1 | BinaryExpression | 0.037659 | 0.041035 | +0.003376 | 1.0896 |
| src/compiler/moduleSpecifiers.ts:1357:1 | getPathsRelativeToRootDirs | 0.041914 | 0.045290 | +0.003376 | 1.0805 |
| src/compiler/transformers/declarations/diagnostics.ts:606:1 | createGetIsolatedDeclarationErrors | 0.056808 | 0.060176 | +0.003367 | 1.0593 |
| src/compiler/emitter.ts:6327:1 | createBracketsMap | 0.040221 | 0.043578 | +0.003357 | 1.0835 |
| src/compiler/watch.ts:220:1 | parseConfigFileWithSystem | 0.035124 | 0.038478 | +0.003354 | 1.0955 |
| src/compiler/core.ts:1934:1 | AnyFunction | 0.035774 | 0.039126 | +0.003352 | 1.0937 |
| src/compiler/utilities.ts:4825:1 | getJSDocRoot | 0.039897 | 0.043241 | +0.003345 | 1.0838 |
| src/compiler/factory/nodeTests.ts:935:1 | isJsxOpeningElement | 0.040301 | 0.043642 | +0.003341 | 1.0829 |
| src/compiler/types.ts:2788:1 | RegularExpressionLiteral | 0.037031 | 0.040358 | +0.003327 | 1.0898 |
| src/compiler/types.ts:7414:1 | CompilerOptions | 0.035026 | 0.038350 | +0.003324 | 1.0949 |
| src/compiler/factory/utilities.ts:1112:1 | canHaveIllegalTypeParameters | 0.042944 | 0.046265 | +0.003321 | 1.0773 |
| src/compiler/factory/nodeTests.ts:1120:1 | isJSDocProtectedTag | 0.041532 | 0.044842 | +0.003309 | 1.0797 |
| src/compiler/types.ts:8062:1 | PackageId | 0.034274 | 0.037572 | +0.003298 | 1.0962 |
| src/compiler/sys.ts:894:1 | FsWatch | 0.036845 | 0.040129 | +0.003285 | 1.0892 |
| src/compiler/scanner.ts:641:1 | skipTrivia | 0.037477 | 0.040747 | +0.003270 | 1.0873 |
| src/compiler/core.ts:1875:1 | toFileNameLowerCase | 0.037629 | 0.040890 | +0.003261 | 1.0867 |
| src/compiler/types.ts:3042:1 | OptionalChain | 0.035586 | 0.038833 | +0.003247 | 1.0913 |
| src/compiler/utilitiesPublic.ts:951:1 | getNonAssignedNameOfDeclaration | 0.038129 | 0.041369 | +0.003240 | 1.0850 |
| src/compiler/moduleNameResolver.ts:2989:1 | isApplicableVersionedTypesKey | 0.039948 | 0.043185 | +0.003237 | 1.0810 |
| src/compiler/utilities.ts:7960:1 | positionsAreOnSameLine | 0.037549 | 0.040785 | +0.003236 | 1.0862 |
| src/compiler/factory/nodeTests.ts:728:1 | isExpressionStatement | 0.043450 | 0.046684 | +0.003233 | 1.0744 |
| src/compiler/utilities.ts:12306:1 | getOptionsSyntaxByValue | 0.061905 | 0.065135 | +0.003230 | 1.0522 |
| src/compiler/moduleNameResolver.ts:787:1 | resolvePackageNameToPackageJson | 0.072748 | 0.075968 | +0.003220 | 1.0443 |
| src/compiler/types.ts:3215:1 | JsxTagNameExpression | 0.066701 | 0.069901 | +0.003201 | 1.0480 |
| src/compiler/utilities.ts:8006:1 | isDeclarationNameOfEnumOrNamespace | 0.051044 | 0.054241 | +0.003197 | 1.0626 |
| src/compiler/utilities.ts:1337:1 | getTextOfNode | 0.045907 | 0.049100 | +0.003193 | 1.0695 |
| src/compiler/factory/nodeFactory.ts:7392:1 |  | 0.034953 | 0.038142 | +0.003189 | 1.0912 |
| src/compiler/resolutionCache.ts:576:1 | GetResolutionWithResolvedFileName | 0.036455 | 0.039642 | +0.003187 | 1.0874 |
| src/compiler/utilities.ts:798:1 | getFullWidth | 0.041701 | 0.044885 | +0.003184 | 1.0764 |
| src/compiler/factory/nodeTests.ts:840:1 | isImportDeclaration | 0.040955 | 0.044139 | +0.003183 | 1.0777 |
| src/compiler/transformers/namedEvaluation.ts:165:1 | classHasDeclaredOrExplicitlyAssignedName | 0.036323 | 0.039495 | +0.003172 | 1.0873 |
| src/compiler/factory/utilitiesPublic.ts:1:1 |  | 0.039304 | 0.042475 | +0.003171 | 1.0807 |
| src/compiler/moduleNameResolver.ts:869:1 | PerDirectoryResolutionCache | 0.037590 | 0.040755 | +0.003166 | 1.0842 |
| src/compiler/parser.ts:504:1 | ForEachChildTable | 0.036298 | 0.039457 | +0.003159 | 1.0870 |
| src/compiler/types.ts:7752:1 | CreateProgramOptions | 0.036837 | 0.039990 | +0.003153 | 1.0856 |
| src/compiler/moduleNameResolver.ts:3064:1 | loadModuleFromImmediateNodeModulesDirectory | 0.036614 | 0.039762 | +0.003148 | 1.0860 |
| src/compiler/types.ts:3384:1 | ExpressionStatement | 0.035886 | 0.039027 | +0.003141 | 1.0875 |
| src/compiler/performanceCore.ts:30:1 | tryGetPerformance | 0.039493 | 0.042629 | +0.003136 | 1.0794 |
| src/compiler/moduleNameResolver.ts:1803:1 | nodeModuleNameResolverWorker | 0.035348 | 0.038480 | +0.003132 | 1.0886 |
| src/compiler/utilities.ts:6500:1 | getResolvedExternalModuleName | 0.035937 | 0.039066 | +0.003129 | 1.0871 |
| src/compiler/checker.ts:1370:1 | RecursionFlags | 0.036084 | 0.039211 | +0.003127 | 1.0867 |
| src/compiler/_namespaces/ts.ts:36:1 |  | 0.037445 | 0.040561 | +0.003116 | 1.0832 |
| src/compiler/path.ts:943:1 | comparePaths | 0.035566 | 0.038670 | +0.003103 | 1.0873 |
| src/compiler/utilities.ts:5318:1 | getFunctionFlags | 0.042081 | 0.045179 | +0.003098 | 1.0736 |
| src/compiler/utilities.ts:1351:1 | indexOfNode | 0.089265 | 0.092363 | +0.003098 | 1.0347 |
| src/compiler/_namespaces/ts.ts:12:1 |  | 0.037226 | 0.040318 | +0.003091 | 1.0830 |
| src/compiler/types.ts:4599:1 | FileIncludeKind | 0.037751 | 0.040835 | +0.003084 | 1.0817 |
| src/compiler/tracing.ts:1:1 |  | 0.035197 | 0.038263 | +0.003066 | 1.0871 |
| src/compiler/path.ts:965:1 | containsPath | 0.035450 | 0.038495 | +0.003045 | 1.0859 |
| src/compiler/program.ts:689:1 |  | 0.036561 | 0.039606 | +0.003044 | 1.0833 |
| src/compiler/parser.ts:1340:1 | setExternalModuleIndicator | 0.035951 | 0.038995 | +0.003044 | 1.0847 |
| src/compiler/factory/nodeTests.ts:804:1 | isClassDeclaration | 0.040856 | 0.043895 | +0.003040 | 1.0744 |
| src/compiler/types.ts:3790:1 | NamedImports | 0.034428 | 0.037467 | +0.003040 | 1.0883 |
| src/compiler/builder.ts:805:1 | handleDtsMayChangeOfAffectedFile | 0.042845 | 0.045879 | +0.003034 | 1.0708 |
| src/compiler/types.ts:1608:1 | ExclamationToken | 0.035223 | 0.038252 | +0.003028 | 1.0860 |
| src/compiler/utilities.ts:3293:1 | getNewTargetContainer | 0.040146 | 0.043171 | +0.003025 | 1.0753 |
| src/compiler/utilitiesPublic.ts:389:1 | textSpanIntersectsWithTextSpan | 0.035377 | 0.038401 | +0.003024 | 1.0855 |
| src/compiler/binder.ts:315:1 | ModuleInstanceState | 0.036732 | 0.039755 | +0.003023 | 1.0823 |
| src/compiler/types.ts:7663:1 | ScriptKind | 0.035025 | 0.038048 | +0.003023 | 1.0863 |
| src/compiler/watch.ts:125:1 | createDiagnosticReporter | 0.109882 | 0.112897 | +0.003016 | 1.0274 |
| src/compiler/factory/nodeTests.ts:756:1 | isContinueStatement | 0.040514 | 0.043526 | +0.003012 | 1.0743 |
| src/compiler/factory/nodeTests.ts:947:1 | isJsxOpeningFragment | 0.049607 | 0.052605 | +0.002998 | 1.0604 |
| src/compiler/factory/nodeTests.ts:812:1 | isTypeAliasDeclaration | 0.040718 | 0.043715 | +0.002997 | 1.0736 |
| src/compiler/utilitiesPublic.ts:1494:1 | isLiteralKind | 0.035983 | 0.038972 | +0.002989 | 1.0831 |
| src/compiler/utilities.ts:7408:1 | isLogicalOperator | 0.039952 | 0.042939 | +0.002987 | 1.0748 |
| src/compiler/utilities.ts:2453:1 | createFileDiagnosticFromMessageChain | 0.037665 | 0.040634 | +0.002969 | 1.0788 |
| src/compiler/types.ts:1509:1 | IsFunctionExpression | 0.035302 | 0.038267 | +0.002965 | 1.0840 |
| src/compiler/watchUtilities.ts:835:1 | getFallbackOptions | 0.034717 | 0.037679 | +0.002963 | 1.0853 |
| src/compiler/_namespaces/ts.ts:13:1 |  | 0.044150 | 0.047105 | +0.002955 | 1.0669 |
| src/compiler/utilities.ts:12091:1 | isSideEffectImport | 0.040649 | 0.043604 | +0.002955 | 1.0727 |
| src/compiler/commandLineParser.ts:4281:1 | getOptionValueWithEmptyStrings | 0.040067 | 0.043018 | +0.002951 | 1.0737 |
| src/compiler/emitter.ts:6357:1 | ParenthesizerRuleOrSelector | 0.037014 | 0.039963 | +0.002949 | 1.0797 |
| src/compiler/commandLineParser.ts:3750:1 | getDefaultCompilerOptions | 0.037354 | 0.040297 | +0.002943 | 1.0788 |
| src/compiler/utilities.ts:2633:1 | isGlobalSourceFile | 0.104520 | 0.107457 | +0.002937 | 1.0281 |
| src/compiler/types.ts:3742:1 | ImportAttribute | 0.035861 | 0.038786 | +0.002925 | 1.0816 |
| src/compiler/utilities.ts:7353:1 | modifiersToFlags | 0.041152 | 0.044069 | +0.002917 | 1.0709 |
| src/compiler/moduleNameResolver.ts:3006:1 | loadModuleFromNearestNodeModulesDirectoryWorker | 0.037849 | 0.040763 | +0.002913 | 1.0770 |
| src/compiler/transformers/esnext.ts:792:1 | getUsingKindOfStatements | 0.044967 | 0.047877 | +0.002911 | 1.0647 |
| src/compiler/builder.ts:774:1 | getNextPendingEmitDiagnosticsFile | 0.037365 | 0.040275 | +0.002910 | 1.0779 |
| src/compiler/parser.ts:10767:1 | addPragmaForMatch | 0.040101 | 0.043006 | +0.002904 | 1.0724 |
| src/compiler/factory/utilities.ts:1065:1 | isStringOrNumericLiteral | 0.039355 | 0.042249 | +0.002894 | 1.0735 |
| src/compiler/types.ts:2996:1 | PropertyAccessEntityNameExpression | 0.036015 | 0.038909 | +0.002894 | 1.0803 |
| src/compiler/types.ts:3274:1 | JsxAttribute | 0.036055 | 0.038944 | +0.002889 | 1.0801 |
| src/compiler/commandLineParser.ts:2310:1 |  | 0.065607 | 0.068491 | +0.002884 | 1.0440 |
| src/compiler/core.ts:2262:1 | tryRemoveSuffix | 0.035084 | 0.037965 | +0.002881 | 1.0821 |
| src/compiler/types.ts:4135:1 | JSDocSatisfiesExpression | 0.035190 | 0.038070 | +0.002880 | 1.0818 |
| src/compiler/utilities.ts:5439:1 | getTextOfIdentifierOrLiteral | 0.189274 | 0.192149 | +0.002875 | 1.0152 |
| src/compiler/moduleSpecifiers.ts:569:1 | getInfo | 0.179176 | 0.182051 | +0.002875 | 1.0160 |
| src/compiler/utilities.ts:954:1 | containsParseError | 0.044178 | 0.047049 | +0.002871 | 1.0650 |
| src/compiler/utilities.ts:9981:1 |  | 0.036005 | 0.038870 | +0.002866 | 1.0796 |
| src/compiler/types.ts:4118:1 | JSDocParameterTag | 0.033996 | 0.036859 | +0.002863 | 1.0842 |
| src/compiler/utilities.ts:5014:1 | walkUpParenthesizedExpressions | 0.050364 | 0.053226 | +0.002862 | 1.0568 |
| src/compiler/executeCommandLine.ts:1257:1 | reportAllStatistics | 0.070131 | 0.072991 | +0.002860 | 1.0408 |
| src/compiler/utilities.ts:11057:1 | isOptionalJSDocPropertyLikeTag | 0.043440 | 0.046293 | +0.002853 | 1.0657 |
| src/compiler/utilities.ts:12228:1 | getLibNameFromLibReference | 0.037619 | 0.040468 | +0.002849 | 1.0757 |
| src/compiler/moduleNameResolver.ts:916:1 | MissingPackageJsonInfo | 0.035950 | 0.038794 | +0.002844 | 1.0791 |
| src/compiler/utilitiesPublic.ts:1368:1 | isPropertyAccessChain | 0.243158 | 0.246002 | +0.002844 | 1.0117 |
| src/compiler/factory/nodeTests.ts:800:1 | isFunctionDeclaration | 0.042223 | 0.045066 | +0.002843 | 1.0673 |
| src/compiler/_namespaces/ts.ts:22:1 |  | 0.037488 | 0.040316 | +0.002828 | 1.0754 |
| src/compiler/semver.ts:405:1 | createComparator | 0.040902 | 0.043729 | +0.002827 | 1.0691 |
| src/compiler/checker.ts:1429:1 | MinArgumentCountFlags | 0.035533 | 0.038357 | +0.002824 | 1.0795 |
| src/compiler/factory/nodeTests.ts:816:1 | isEnumDeclaration | 0.042698 | 0.045516 | +0.002817 | 1.0660 |
| src/compiler/visitorPublic.ts:289:1 | visitArray | 0.035763 | 0.038579 | +0.002817 | 1.0788 |
| src/compiler/watchPublic.ts:106:1 | readBuilderProgram | 0.147974 | 0.150786 | +0.002812 | 1.0190 |
| src/compiler/factory/utilities.ts:607:1 | isCommaExpression | 0.154976 | 0.157787 | +0.002811 | 1.0181 |
| src/compiler/types.ts:3060:1 | BindableStaticNameExpression | 0.035553 | 0.038361 | +0.002808 | 1.0790 |
| src/compiler/core.ts:1861:1 |  | 0.038779 | 0.041587 | +0.002808 | 1.0724 |
| src/compiler/types.ts:10238:1 | PragmaDefinition | 0.036535 | 0.039339 | +0.002804 | 1.0767 |
| src/compiler/_namespaces/ts.ts:9:1 |  | 0.040719 | 0.043523 | +0.002804 | 1.0689 |
| src/compiler/moduleNameResolver.ts:1725:1 | nodeNextModuleNameResolver | 0.037825 | 0.040620 | +0.002794 | 1.0739 |
| src/compiler/factory/nodeChildren.ts:11:1 |  | 0.036986 | 0.039781 | +0.002794 | 1.0756 |
| src/compiler/types.ts:4533:1 | ScriptReferenceHost | 0.036301 | 0.039095 | +0.002794 | 1.0770 |
| src/compiler/factory/nodeTests.ts:788:1 | isDebuggerStatement | 0.040846 | 0.043640 | +0.002794 | 1.0684 |
| src/compiler/types.ts:7795:1 | CommandLineOptionOfStringType | 0.038267 | 0.041060 | +0.002793 | 1.0730 |
| src/compiler/types.ts:23:1 | Path | 0.034925 | 0.037714 | +0.002789 | 1.0799 |
| src/compiler/transformers/typeSerializer.ts:94:1 | RuntimeTypeSerializer | 0.036446 | 0.039228 | +0.002782 | 1.0763 |
| src/compiler/types.ts:2981:1 | PropertyAccessChain | 0.035705 | 0.038473 | +0.002768 | 1.0775 |
| src/compiler/moduleNameResolver.ts:900:1 | ModuleResolutionCache | 0.038754 | 0.041521 | +0.002767 | 1.0714 |
| src/compiler/moduleNameResolver.ts:318:1 | ModuleResolutionState | 0.041462 | 0.044212 | +0.002750 | 1.0663 |
| src/compiler/transformer.ts:669:1 |  | 0.064061 | 0.066803 | +0.002742 | 1.0428 |
| src/compiler/types.ts:7643:1 | ImportsNotUsedAsValues | 0.035145 | 0.037885 | +0.002740 | 1.0780 |
| src/compiler/utilities.ts:10310:1 | isCheckJsEnabledForFile | 0.036071 | 0.038805 | +0.002734 | 1.0758 |
| src/compiler/commandLineParser.ts:1733:1 |  | 0.064113 | 0.066846 | +0.002733 | 1.0426 |
| src/compiler/moduleSpecifiers.ts:796:1 |  | 0.036945 | 0.039676 | +0.002730 | 1.0739 |
| src/compiler/debug.ts:2:1 |  | 0.035985 | 0.038707 | +0.002723 | 1.0757 |
| src/compiler/utilities.ts:3308:1 | SuperContainer | 0.035437 | 0.038158 | +0.002721 | 1.0768 |
| src/compiler/types.ts:3366:1 | BlockLike | 0.045261 | 0.047979 | +0.002717 | 1.0600 |
| src/compiler/factory/nodeTests.ts:347:1 | isAwaitKeyword | 0.041209 | 0.043922 | +0.002713 | 1.0658 |
| src/compiler/visitorPublic.ts:1:1 |  | 0.036057 | 0.038765 | +0.002708 | 1.0751 |
| src/compiler/watch.ts:304:1 | createTabularErrorsDisplay | 0.071901 | 0.074570 | +0.002668 | 1.0371 |
| src/compiler/utilitiesPublic.ts:2451:1 | isModuleReference | 0.040199 | 0.042861 | +0.002662 | 1.0662 |
| src/compiler/sys.ts:963:1 | FileSystemEntryExists | 0.036779 | 0.039438 | +0.002659 | 1.0723 |
| src/compiler/utilities.ts:7413:1 | isLogicalOrCoalescingAssignmentOperator | 0.037687 | 0.040342 | +0.002656 | 1.0705 |
| src/compiler/types.ts:580:1 | PunctuationOrKeywordSyntaxKind | 0.035054 | 0.037697 | +0.002643 | 1.0754 |
| src/compiler/utilities.ts:5637:1 | getExpressionAssociativity | 0.151306 | 0.153946 | +0.002640 | 1.0174 |
| src/compiler/types.ts:700:1 | TypeNodeSyntaxKind | 0.034969 | 0.037600 | +0.002631 | 1.0752 |
| src/compiler/commandLineParser.ts:4075:1 | matchesExcludeWorker | 0.035662 | 0.038292 | +0.002631 | 1.0738 |
| src/compiler/moduleNameResolver.ts:526:1 | getCandidateFromTypeRoot | 0.035881 | 0.038508 | +0.002627 | 1.0732 |
| src/compiler/types.ts:739:1 | TokenSyntaxKind | 0.034967 | 0.037587 | +0.002620 | 1.0749 |
| src/compiler/types.ts:3075:1 | BindableElementAccessExpression | 0.035636 | 0.038244 | +0.002609 | 1.0732 |
| src/compiler/types.ts:1589:1 | NodeArray | 0.036290 | 0.038898 | +0.002608 | 1.0719 |
| src/compiler/utilities.ts:873:1 | createModeMismatchDetails | 0.038625 | 0.041209 | +0.002584 | 1.0669 |
| src/compiler/factory/utilities.ts:1105:1 | canHaveIllegalType | 0.040167 | 0.042737 | +0.002570 | 1.0640 |
| src/compiler/utilities.ts:2440:1 | createDiagnosticForNodeArrayFromMessageChain | 0.036783 | 0.039344 | +0.002560 | 1.0696 |
| src/compiler/sys.ts:563:1 |  | 0.036371 | 0.038929 | +0.002558 | 1.0703 |
| src/compiler/transformers/classThis.ts:90:1 | classHasClassThisAssignment | 0.041870 | 0.044427 | +0.002557 | 1.0611 |
| src/compiler/utilities.ts:2525:1 | getErrorSpanForArrowFunction | 0.110566 | 0.113110 | +0.002544 | 1.0230 |
| src/compiler/moduleNameResolver.ts:1966:1 | normalizePathForCJSResolution | 0.034952 | 0.037491 | +0.002540 | 1.0727 |
| src/compiler/moduleNameResolver.ts:3216:1 | mangleScopedPackageName | 0.038129 | 0.040668 | +0.002539 | 1.0666 |
| src/compiler/utilities.ts:7241:1 | getSelectedEffectiveModifierFlags | 0.043342 | 0.045879 | +0.002537 | 1.0585 |
| src/compiler/commandLineParser.ts:2785:1 | serializeOptionBaseObject | 0.072065 | 0.074602 | +0.002536 | 1.0352 |
| src/compiler/utilities.ts:8029:1 | isWatchSet | 0.037169 | 0.039705 | +0.002536 | 1.0682 |
| src/compiler/utilities.ts:10294:1 | extensionFromPath | 0.037741 | 0.040268 | +0.002527 | 1.0670 |
| src/compiler/types.ts:4573:1 | WriteFileCallbackData | 0.035800 | 0.038317 | +0.002516 | 1.0703 |
| src/compiler/sys.ts:242:1 | WatchedFileWithUnchangedPolls | 0.035323 | 0.037838 | +0.002515 | 1.0712 |
| src/compiler/transformers/classFields.ts:343:1 | PrivateEnv | 0.036792 | 0.039306 | +0.002514 | 1.0683 |
| src/compiler/utilities.ts:5189:1 | getPropertyAssignmentAliasLikeExpression | 0.161603 | 0.164116 | +0.002513 | 1.0156 |
| src/compiler/binder.ts:3789:1 | isExportsOrModuleExportsOrAlias | 0.048738 | 0.051241 | +0.002503 | 1.0514 |
| src/compiler/debug.ts:108:1 | LoggingHost | 0.036063 | 0.038565 | +0.002502 | 1.0694 |
| src/compiler/utilities.ts:9979:1 |  | 0.035254 | 0.037755 | +0.002500 | 1.0709 |
| src/compiler/factory/nodeFactory.ts:7309:1 | aggregateChildrenFlags | 0.034001 | 0.036501 | +0.002500 | 1.0735 |
| src/compiler/utilities.ts:4900:1 | getAssignmentTargetKind | 0.075308 | 0.077808 | +0.002500 | 1.0332 |
| src/compiler/types.ts:4677:1 | FilePreprocessingFileExplainingDiagnostic | 0.035305 | 0.037800 | +0.002495 | 1.0707 |
| src/compiler/factory/nodeTests.ts:472:1 | isFunctionTypeNode | 0.044987 | 0.047478 | +0.002491 | 1.0554 |
| src/compiler/types.ts:4480:1 | SourceFile | 0.037199 | 0.039682 | +0.002483 | 1.0667 |
| src/compiler/moduleNameResolver.ts:1780:1 | nodeModuleNameResolver | 0.035008 | 0.037488 | +0.002480 | 1.0708 |
| src/compiler/factory/nodeTests.ts:656:1 | isYieldExpression | 0.042083 | 0.044557 | +0.002474 | 1.0588 |
| src/compiler/utilities.ts:9784:1 | FileSystemEntries | 0.046445 | 0.048914 | +0.002469 | 1.0532 |
| src/compiler/types.ts:4139:1 | JSDocImportTag | 0.035037 | 0.037504 | +0.002467 | 1.0704 |
| src/compiler/moduleNameResolver.ts:925:1 | isPackageJsonInfo | 0.035971 | 0.038438 | +0.002467 | 1.0686 |
| src/compiler/utilities.ts:2677:1 | isVarConst | 0.044026 | 0.046491 | +0.002466 | 1.0560 |
| src/compiler/types.ts:3049:1 | OptionalChainRoot | 0.036006 | 0.038467 | +0.002461 | 1.0683 |
| src/compiler/utilitiesPublic.ts:1136:1 | getJSDocPrivateTag | 0.049561 | 0.052014 | +0.002453 | 1.0495 |
| src/compiler/types.ts:4114:1 | JSDocPropertyTag | 0.034532 | 0.036981 | +0.002449 | 1.0709 |
| src/compiler/commandLineParser.ts:3690:1 | getExtendedConfig | 0.043280 | 0.045728 | +0.002447 | 1.0565 |
| src/compiler/path.ts:146:1 | isVolumeCharacter | 0.036462 | 0.038908 | +0.002446 | 1.0671 |
| src/compiler/utilities.ts:11484:1 | NameResolver | 0.035488 | 0.037910 | +0.002422 | 1.0683 |
| src/compiler/commandLineParser.ts:2333:1 | getCommandLineWatchOptionsMap | 0.044611 | 0.047028 | +0.002417 | 1.0542 |
| src/compiler/types.ts:1606:1 | DotDotDotToken | 0.036013 | 0.038426 | +0.002413 | 1.0670 |
| src/compiler/types.ts:511:1 | PseudoLiteralSyntaxKind | 0.035235 | 0.037646 | +0.002411 | 1.0684 |
| src/compiler/types.ts:2523:1 | MultiplicativeOperator | 0.044855 | 0.047261 | +0.002406 | 1.0536 |
| src/compiler/performance.ts:125:1 | getDuration | 0.038689 | 0.041083 | +0.002395 | 1.0619 |
| src/compiler/utilities.ts:2685:1 | isVarConstLike | 0.039670 | 0.042050 | +0.002381 | 1.0600 |
| src/compiler/factory/utilities.ts:1:1 |  | 0.037049 | 0.039426 | +0.002377 | 1.0642 |
| src/compiler/factory/emitHelpers.ts:993:1 |  | 0.061586 | 0.063962 | +0.002375 | 1.0386 |
| src/compiler/types.ts:40:1 | SyntaxKind | 0.035358 | 0.037731 | +0.002373 | 1.0671 |
| src/compiler/moduleNameResolver.ts:3199:1 |  | 0.036291 | 0.038661 | +0.002370 | 1.0653 |
| src/compiler/builder.ts:2229:1 | toBuilderStateFileInfoForMultiEmit | 0.036217 | 0.038581 | +0.002365 | 1.0653 |
| src/compiler/types.ts:4150:1 | FlowFlags | 0.039876 | 0.042240 | +0.002364 | 1.0593 |
| src/compiler/path.ts:1068:1 | getRelativePathToDirectoryOrUrl | 0.037033 | 0.039388 | +0.002355 | 1.0636 |
| src/compiler/scanner.ts:609:1 | isOctalDigit | 0.036299 | 0.038649 | +0.002350 | 1.0647 |
| src/compiler/commandLineParser.ts:4214:1 | hasFileWithHigherPriorityExtension | 0.039052 | 0.041402 | +0.002350 | 1.0602 |
| src/compiler/watchPublic.ts:174:1 | ProgramHost | 0.034525 | 0.036872 | +0.002347 | 1.0680 |
| src/compiler/types.ts:10079:1 | TextChangeRange | 0.035947 | 0.038293 | +0.002346 | 1.0653 |
| src/compiler/types.ts:7992:1 | ModuleResolutionHost | 0.034466 | 0.036797 | +0.002331 | 1.0676 |
| src/compiler/parser.ts:462:1 | isJSDocLikeText | 0.036355 | 0.038685 | +0.002330 | 1.0641 |
| src/compiler/utilitiesPublic.ts:1161:1 | getJSDocReadonlyTagNoCache | 0.049992 | 0.052314 | +0.002322 | 1.0465 |
| src/compiler/types.ts:2969:1 | PropertyAccessExpression | 0.036084 | 0.038402 | +0.002319 | 1.0643 |
| src/compiler/factory/utilities.ts:584:1 | findUseStrictPrologue | 0.269396 | 0.271712 | +0.002317 | 1.0086 |
| src/compiler/factory/nodeTests.ts:724:1 | isEmptyStatement | 0.042232 | 0.044548 | +0.002316 | 1.0548 |
| src/compiler/commandLineParser.ts:2781:1 | serializeWatchOptions | 0.035946 | 0.038258 | +0.002312 | 1.0643 |
| src/compiler/moduleNameResolver.ts:882:1 | NonRelativeNameResolutionCache | 0.035971 | 0.038280 | +0.002309 | 1.0642 |
| src/compiler/types.ts:3055:1 | BindableObjectDefinePropertyCall | 0.035992 | 0.038283 | +0.002291 | 1.0637 |
| src/compiler/types.ts:9997:1 | ModuleSpecifierResolutionHost | 0.035281 | 0.037572 | +0.002291 | 1.0649 |
| src/compiler/moduleNameResolver.ts:3236:1 | unmangleScopedPackageName | 0.035340 | 0.037631 | +0.002291 | 1.0648 |
| src/compiler/factory/nodeTests.ts:296:1 | isExclamationToken | 0.043756 | 0.046044 | +0.002288 | 1.0523 |
| src/compiler/types.ts:9725:1 | NodeVisitor | 0.037118 | 0.039404 | +0.002286 | 1.0616 |
| src/compiler/factory/emitHelpers.ts:934:1 |  | 0.061790 | 0.064070 | +0.002280 | 1.0369 |
| src/compiler/utilitiesPublic.ts:1518:1 | isTemplateLiteralKind | 0.036168 | 0.038445 | +0.002277 | 1.0630 |
| src/compiler/semver.ts:35:1 |  | 0.037899 | 0.040174 | +0.002275 | 1.0600 |
| src/compiler/factory/nodeTests.ts:999:1 | isShorthandPropertyAssignment | 0.040215 | 0.042487 | +0.002272 | 1.0565 |
| src/compiler/transformers/declarations.ts:1:1 |  | 0.037484 | 0.039756 | +0.002272 | 1.0606 |
| src/compiler/commandLineParser.ts:1745:1 |  | 0.037213 | 0.039482 | +0.002269 | 1.0610 |
| src/compiler/moduleNameResolver.ts:961:1 | getKeyForCompilerOptions | 0.037887 | 0.040148 | +0.002261 | 1.0597 |
| src/compiler/utilities.ts:909:1 | packageIdToPackageName | 0.109793 | 0.112052 | +0.002259 | 1.0206 |
| src/compiler/types.ts:1584:1 | MutableNodeArray | 0.035215 | 0.037473 | +0.002257 | 1.0641 |
| src/compiler/utilitiesPublic.ts:1526:1 | isTemplateMiddleOrTemplateTail | 0.038325 | 0.040580 | +0.002255 | 1.0588 |
| src/compiler/factory/nodeTests.ts:520:1 | isInferTypeNode | 0.042081 | 0.044335 | +0.002254 | 1.0536 |
| src/compiler/parser.ts:505:1 |  | 0.064639 | 0.066874 | +0.002235 | 1.0346 |
| src/compiler/core.ts:2386:1 | Pattern | 0.034665 | 0.036899 | +0.002235 | 1.0645 |
| src/compiler/parser.ts:420:1 |  | 0.035198 | 0.037424 | +0.002226 | 1.0632 |
| src/compiler/transformers/esnext.ts:784:1 | getUsingKindOfVariableStatement | 0.041228 | 0.043451 | +0.002223 | 1.0539 |
| src/compiler/utilities.ts:3073:1 | isObjectLiteralMethod | 0.042217 | 0.044439 | +0.002222 | 1.0526 |
| src/compiler/types.ts:502:1 | LiteralSyntaxKind | 0.035460 | 0.037682 | +0.002222 | 1.0627 |
| src/compiler/factory/nodeTests.ts:393:1 | isQualifiedName | 0.039904 | 0.042114 | +0.002210 | 1.0554 |
| src/compiler/sys.ts:183:1 | WatchedFileWithIsClosed | 0.036381 | 0.038589 | +0.002208 | 1.0607 |
| src/compiler/parser.ts:9942:1 | markAsIntersectingIncrementalChange | 0.040735 | 0.042940 | +0.002205 | 1.0541 |
| src/compiler/utilitiesPublic.ts:662:1 | getCombinedModifierFlags | 0.043879 | 0.046083 | +0.002204 | 1.0502 |
| src/compiler/types.ts:3025:1 | CallExpression | 0.035431 | 0.037635 | +0.002204 | 1.0622 |
| src/compiler/emitter.ts:610:1 | getSingleOutputFileNames | 0.177336 | 0.179536 | +0.002200 | 1.0124 |
| src/compiler/moduleNameResolver.ts:1779:18 | nodeModuleNameResolver | 0.035327 | 0.037514 | +0.002187 | 1.0619 |
| src/compiler/utilitiesPublic.ts:1185:1 | getJSDocThisTag | 0.050855 | 0.053041 | +0.002186 | 1.0430 |
| src/compiler/types.ts:1605:1 | DotToken | 0.035289 | 0.037474 | +0.002185 | 1.0619 |
| src/compiler/factory/nodeTests.ts:886:1 | isExportAssignment | 0.040499 | 0.042678 | +0.002179 | 1.0538 |
| src/compiler/factory/nodeTests.ts:824:1 | isModuleBlock | 0.041337 | 0.043511 | +0.002174 | 1.0526 |
| src/compiler/utilitiesPublic.ts:1480:1 | isToken | 0.040222 | 0.042396 | +0.002174 | 1.0540 |
| src/compiler/transformer.ts:76:1 |  | 0.035226 | 0.037395 | +0.002169 | 1.0616 |
| src/compiler/factory/nodeTests.ts:430:1 | isMethodDeclaration | 0.046214 | 0.048383 | +0.002169 | 1.0469 |
| src/compiler/utilities.ts:5564:1 | isNamedEvaluation | 0.156736 | 0.158900 | +0.002164 | 1.0138 |
| src/compiler/types.ts:2987:1 | PropertyAccessChainRoot | 0.035973 | 0.038135 | +0.002162 | 1.0601 |
| src/compiler/utilities.ts:4684:1 | filterOwnedJSDocTags | 0.040481 | 0.042643 | +0.002162 | 1.0534 |
| src/compiler/types.ts:4066:1 | JSDocTypeTag | 0.034441 | 0.036603 | +0.002162 | 1.0628 |
| src/compiler/moduleNameResolver.ts:856:1 | ModeAwareCache | 0.037866 | 0.040019 | +0.002153 | 1.0569 |
| src/compiler/factory/nodeTests.ts:772:1 | isSwitchStatement | 0.042186 | 0.044339 | +0.002153 | 1.0510 |
| src/compiler/types.ts:2935:1 | ArrayLiteralExpression | 0.036376 | 0.038528 | +0.002152 | 1.0591 |
| src/compiler/types.ts:1448:1 | IsBlockScopedContainer | 0.034781 | 0.036928 | +0.002147 | 1.0617 |
| src/compiler/types.ts:1487:1 | IsFunctionLike | 0.035424 | 0.037570 | +0.002147 | 1.0606 |
| src/compiler/types.ts:7765:1 | CommandLineOptionBase | 0.037376 | 0.039519 | +0.002142 | 1.0573 |
| src/compiler/types.ts:3331:1 | NotEmittedTypeElement | 0.037653 | 0.039795 | +0.002142 | 1.0569 |
| src/compiler/watch.ts:177:1 | getLocaleTimeString | 0.072417 | 0.074551 | +0.002134 | 1.0295 |
| src/compiler/watch.ts:276:1 | getErrorSummaryText | 0.069458 | 0.071592 | +0.002133 | 1.0307 |
| src/compiler/watchPublic.ts:353:1 | WatchOfConfigFile | 0.034477 | 0.036604 | +0.002127 | 1.0617 |
| src/compiler/types.ts:4514:1 | JsonMinusNumericLiteral | 0.037146 | 0.039266 | +0.002120 | 1.0571 |
| src/compiler/utilitiesPublic.ts:2538:1 | isSetAccessor | 0.038886 | 0.041000 | +0.002113 | 1.0544 |
| src/compiler/types.ts:3002:1 | ElementAccessExpression | 0.036076 | 0.038188 | +0.002112 | 1.0585 |
| src/compiler/types.ts:2966:1 | EntityNameOrEntityNameExpression | 0.035863 | 0.037973 | +0.002111 | 1.0589 |
| src/compiler/builder.ts:1112:1 | IncrementalBuildInfoEmitSignature | 0.036227 | 0.038335 | +0.002108 | 1.0582 |
| src/compiler/types.ts:10295:1 | JSDocParsingMode | 0.035287 | 0.037392 | +0.002105 | 1.0596 |
| src/compiler/types.ts:4105:1 | JSDocPropertyLikeTag | 0.034961 | 0.037061 | +0.002100 | 1.0601 |
| src/compiler/utilitiesPublic.ts:1380:1 | isOptionalChain | 0.040659 | 0.042752 | +0.002093 | 1.0515 |
| src/compiler/transformers/utilities.ts:585:1 | getProperties | 0.035226 | 0.037318 | +0.002092 | 1.0594 |
| src/compiler/types.ts:1595:1 | Token | 0.035680 | 0.037769 | +0.002089 | 1.0585 |
| src/compiler/types.ts:9992:1 | HasCurrentDirectory | 0.034614 | 0.036703 | +0.002089 | 1.0603 |
| src/compiler/types.ts:3820:1 | ModuleExportName | 0.034868 | 0.036955 | +0.002086 | 1.0598 |
| src/compiler/path.ts:1045:1 | getRelativePathFromDirectory | 0.035304 | 0.037388 | +0.002084 | 1.0590 |
| src/compiler/types.ts:957:1 | JSDocContainer | 0.037945 | 0.040026 | +0.002081 | 1.0549 |
| src/compiler/sourcemap.ts:679:1 | compareSourcePositions | 0.035984 | 0.038058 | +0.002074 | 1.0576 |
| src/compiler/utilitiesPublic.ts:2517:1 | isJSDocNode | 0.041941 | 0.044014 | +0.002073 | 1.0494 |
| src/compiler/types.ts:3023:1 | SuperProperty | 0.035775 | 0.037845 | +0.002070 | 1.0579 |
| src/compiler/sys.ts:52:1 | clearTimeout | 0.035618 | 0.037681 | +0.002063 | 1.0579 |
| src/compiler/parser.ts:495:1 | hasModifierOfKind | 0.043750 | 0.045810 | +0.002060 | 1.0471 |
| src/compiler/sys.ts:577:1 | RecursiveDirectoryWatcherHost | 0.037093 | 0.039142 | +0.002049 | 1.0552 |
| src/compiler/utilitiesPublic.ts:1299:1 | formatJSDocLink | 0.037833 | 0.039881 | +0.002048 | 1.0541 |
| src/compiler/factory/nodeTests.ts:1132:1 | isJSDocOverloadTag | 0.040973 | 0.043011 | +0.002038 | 1.0497 |
| src/compiler/builder.ts:121:1 | SerializedDiagnosticMessageChain | 0.036175 | 0.038214 | +0.002038 | 1.0563 |
| src/compiler/programDiagnostics.ts:58:1 | FileReasonToChainCache | 0.037541 | 0.039577 | +0.002036 | 1.0542 |
| src/compiler/factory/nodeTests.ts:858:1 | isAssertEntry | 0.044185 | 0.046220 | +0.002035 | 1.0461 |
| src/compiler/types.ts:3238:1 | JsxOpeningElement | 0.047114 | 0.049148 | +0.002034 | 1.0432 |
| src/compiler/moduleNameResolver.ts:910:1 | NonRelativeModuleNameResolutionCache | 0.036477 | 0.038508 | +0.002031 | 1.0557 |
| src/compiler/types.ts:3038:1 | CallChainRoot | 0.034677 | 0.036704 | +0.002027 | 1.0585 |
| src/compiler/sys.ts:81:1 | FileWatcherEventKind | 0.036225 | 0.038251 | +0.002027 | 1.0559 |
| src/compiler/types.ts:1462:1 | IsControlFlowContainer | 0.035726 | 0.037751 | +0.002024 | 1.0567 |
| src/compiler/utilities.ts:7741:1 | base64encode | 0.108689 | 0.110706 | +0.002017 | 1.0186 |
| src/compiler/moduleNameResolver.ts:2656:1 | loadModuleFromImports | 0.036377 | 0.038391 | +0.002014 | 1.0554 |
| src/compiler/types.ts:2967:1 | AccessExpression | 0.036128 | 0.038141 | +0.002013 | 1.0557 |
| src/compiler/transformers/destructuring.ts:227:1 | bindingOrAssignmentPatternContainsNonLiteralComputedName | 0.037667 | 0.039680 | +0.002013 | 1.0534 |
| src/compiler/checker.ts:1390:1 | ExpandingFlags | 0.037856 | 0.039864 | +0.002008 | 1.0531 |
| src/compiler/transformers/destructuring.ts:338:1 | flattenBindingOrAssignmentElement | 0.040141 | 0.042149 | +0.002008 | 1.0500 |
| src/compiler/watchPublic.ts:132:1 | IncrementalProgramOptions | 0.035399 | 0.037401 | +0.002002 | 1.0566 |
| src/compiler/utilities.ts:5427:1 | isPropertyNameLiteral | 0.041447 | 0.043447 | +0.002000 | 1.0483 |
| src/compiler/factory/emitNode.ts:289:1 | moveEmitHelpers | 0.077632 | 0.079630 | +0.001997 | 1.0257 |
| src/compiler/watchPublic.ts:335:1 | Watch | 0.034683 | 0.036678 | +0.001995 | 1.0575 |
| src/compiler/utilities.ts:9948:1 | getScriptKindFromFileName | 0.035746 | 0.037738 | +0.001992 | 1.0557 |
| src/compiler/binder.ts:3905:1 | lookupSymbolForName | 0.069246 | 0.071235 | +0.001990 | 1.0287 |
| src/compiler/types.ts:7412:1 | CompilerOptionsValue | 0.035628 | 0.037612 | +0.001984 | 1.0557 |
| src/compiler/types.ts:1640:1 | PrivateKeyword | 0.034019 | 0.035999 | +0.001980 | 1.0582 |
| src/compiler/utilitiesPublic.ts:1170:1 | getJSDocDeprecatedTag | 0.051439 | 0.053409 | +0.001970 | 1.0383 |
| src/compiler/types.ts:582:1 | KeywordSyntaxKind | 0.034762 | 0.036728 | +0.001966 | 1.0566 |
| src/compiler/types.ts:2902:1 | TemplateLiteralToken | 0.036637 | 0.038601 | +0.001964 | 1.0536 |
| src/compiler/utilities.ts:12159:1 |  | 0.036092 | 0.038055 | +0.001964 | 1.0544 |
| src/compiler/factory/nodeTests.ts:1144:1 | isJSDocEnumTag | 0.040608 | 0.042568 | +0.001961 | 1.0483 |
| src/compiler/factory/emitHelpers.ts:1369:1 |  | 0.062749 | 0.064704 | +0.001954 | 1.0311 |
| src/compiler/types.ts:2235:1 | ConstructorTypeNode | 0.036223 | 0.038177 | +0.001954 | 1.0539 |
| src/compiler/_namespaces/ts.performance.ts:3:1 |  | 0.036651 | 0.038603 | +0.001953 | 1.0533 |
| src/compiler/moduleSpecifiers.ts:386:1 | getModuleSpecifiersWithCacheInfo | 0.079391 | 0.081343 | +0.001952 | 1.0246 |
| src/compiler/scanner.ts:762:1 | scanConflictMarkerTrivia | 0.036248 | 0.038195 | +0.001947 | 1.0537 |
| src/compiler/utilities.ts:12400:1 | suppressLeadingAndTrailingTrivia | 0.047238 | 0.049179 | +0.001941 | 1.0411 |
| src/compiler/factory/nodeTests.ts:382:1 | isImportKeyword | 0.040520 | 0.042460 | +0.001940 | 1.0479 |
| src/compiler/factory/nodeTests.ts:1090:1 | isJSDocSignature | 0.040754 | 0.042687 | +0.001934 | 1.0474 |
| src/compiler/types.ts:3769:1 | NamespaceExportDeclaration | 0.035085 | 0.037018 | +0.001933 | 1.0551 |
| src/compiler/executeCommandLine.ts:810:1 |  | 0.034515 | 0.036448 | +0.001933 | 1.0560 |
| src/compiler/types.ts:1413:1 | IsContainer | 0.034101 | 0.036033 | +0.001932 | 1.0567 |
| src/compiler/types.ts:3496:1 | CaseOrDefaultClause | 0.035145 | 0.037075 | +0.001930 | 1.0549 |
| src/compiler/factory/nodeTests.ts:1074:1 | isJSDocVariadicType | 0.041518 | 0.043445 | +0.001926 | 1.0464 |
| src/compiler/moduleNameResolver.ts:3261:1 | classicNameResolver | 0.037743 | 0.039658 | +0.001915 | 1.0507 |
| src/compiler/emitter.ts:668:1 | getComputedCommonSourceDirectory | 0.035296 | 0.037208 | +0.001912 | 1.0542 |
| src/compiler/utilities.ts:9294:1 |  | 0.073273 | 0.075183 | +0.001911 | 1.0261 |
| src/compiler/path.ts:1039:1 | getRelativePathFromDirectory | 0.034794 | 0.036703 | +0.001909 | 1.0549 |
| src/compiler/utilitiesPublic.ts:1180:1 | getJSDocEnumTag | 0.044683 | 0.046591 | +0.001908 | 1.0427 |
| src/compiler/transformers/es2017.ts:1059:1 | createSuperAccessVariableStatement | 0.364672 | 0.366578 | +0.001906 | 1.0052 |
| src/compiler/utilities.ts:824:1 | getResolvedModuleFromResolution | 0.038187 | 0.040092 | +0.001905 | 1.0499 |
| src/compiler/types.ts:3826:1 | TypeOnlyCompatibleAliasDeclaration | 0.036042 | 0.037932 | +0.001891 | 1.0525 |
| src/compiler/parser.ts:432:1 |  | 0.065674 | 0.067549 | +0.001875 | 1.0286 |
| src/compiler/factory/nodeTests.ts:1128:1 | isJSDocOverrideTag | 0.041184 | 0.043058 | +0.001874 | 1.0455 |
| src/compiler/transformers/declarations.ts:221:1 | getDeclarationDiagnostics | 0.073510 | 0.075384 | +0.001873 | 1.0255 |
| src/compiler/parser.ts:1399:1 | updateSourceFile | 0.035690 | 0.037563 | +0.001873 | 1.0525 |
| src/compiler/transformers/taggedTemplate.ts:29:1 | ProcessLevel | 0.036111 | 0.037984 | +0.001873 | 1.0519 |
| src/compiler/utilities.ts:3319:1 | SuperContainerOrFunctions | 0.034641 | 0.036513 | +0.001871 | 1.0540 |
| src/compiler/utilities.ts:12168:1 |  | 0.035734 | 0.037604 | +0.001870 | 1.0523 |
| src/compiler/types.ts:10231:1 | PragmaArgumentSpecification | 0.035814 | 0.037683 | +0.001868 | 1.0522 |
| src/compiler/types.ts:3846:1 | TypeOnlyAliasDeclaration | 0.036163 | 0.038022 | +0.001859 | 1.0514 |
| src/compiler/watchUtilities.ts:692:1 | WatchFactory | 0.035410 | 0.037264 | +0.001854 | 1.0524 |
| src/compiler/utilities.ts:3335:1 | getSuperContainer | 0.037310 | 0.039159 | +0.001848 | 1.0495 |
| src/compiler/factory/nodeTests.ts:570:1 | isBindingElement | 0.043187 | 0.045027 | +0.001840 | 1.0426 |
| src/compiler/watchPublic.ts:1:1 |  | 0.038816 | 0.040655 | +0.001838 | 1.0474 |
| src/compiler/factory/nodeTests.ts:343:1 | isAssertsKeyword | 0.042690 | 0.044523 | +0.001833 | 1.0429 |
| src/compiler/utilitiesPublic.ts:2578:1 | hasOnlyExpressionInitializer | 0.040418 | 0.042248 | +0.001830 | 1.0453 |
| src/compiler/utilities.ts:5281:1 | isContextualKeyword | 0.038734 | 0.040562 | +0.001828 | 1.0472 |
| src/compiler/utilities.ts:959:1 | aggregateChildData | 0.047228 | 0.049052 | +0.001824 | 1.0386 |
| src/compiler/semver.ts:261:1 |  | 0.038116 | 0.039936 | +0.001820 | 1.0478 |
| src/compiler/utilities.ts:9895:1 | getBasePaths | 0.037494 | 0.039311 | +0.001818 | 1.0485 |
| src/compiler/factory/nodeFactory.ts:7305:1 | propagateChildrenFlags | 0.034914 | 0.036731 | +0.001817 | 1.0521 |
| src/compiler/factory/nodeTests.ts:1124:1 | isJSDocReadonlyTag | 0.040366 | 0.042183 | +0.001817 | 1.0450 |
| src/compiler/path.ts:934:1 | comparePathsCaseInsensitive | 0.034556 | 0.036370 | +0.001814 | 1.0525 |
| src/compiler/scanner.ts:957:1 | getLeadingCommentRanges | 0.037890 | 0.039701 | +0.001810 | 1.0478 |
| src/compiler/utilities.ts:8087:1 | isWriteAccess | 0.042530 | 0.044337 | +0.001807 | 1.0425 |
| src/compiler/utilitiesPublic.ts:2426:1 | isBlockStatement | 0.045030 | 0.046834 | +0.001805 | 1.0401 |
| src/compiler/types.ts:1563:1 | IsObjectLiteralOrClassExpressionMethodOrAccessor | 0.036417 | 0.038221 | +0.001805 | 1.0496 |
| src/compiler/utilities.ts:7707:1 | getStringFromExpandedCharCodes | 0.036503 | 0.038305 | +0.001802 | 1.0494 |
| src/compiler/types.ts:2918:1 | TemplateSpan | 0.039604 | 0.041401 | +0.001797 | 1.0454 |
| src/compiler/debug.ts:1:1 |  | 0.035363 | 0.037160 | +0.001797 | 1.0508 |
| src/compiler/factory/nodeTests.ts:1078:1 | isJSDocNamepathType | 0.040975 | 0.042763 | +0.001788 | 1.0436 |
| src/compiler/types.ts:2506:1 | YieldExpression | 0.036953 | 0.038735 | +0.001782 | 1.0482 |
| src/compiler/utilitiesPublic.ts:2643:1 | hasInternalAnnotation | 0.036350 | 0.038126 | +0.001776 | 1.0489 |
| src/compiler/types.ts:9814:1 | PrintHandlers | 0.034997 | 0.036771 | +0.001774 | 1.0507 |
| src/compiler/transformers/declarations.ts:1865:1 | isAlwaysType | 0.041019 | 0.042786 | +0.001767 | 1.0431 |
| src/compiler/utilities.ts:3543:1 | nodeOrChildIsDecorated | 0.061263 | 0.063027 | +0.001764 | 1.0288 |
| src/compiler/transformers/declarations/diagnostics.ts:164:1 | createGetSymbolAccessibilityDiagnosticForNodeName | 0.114999 | 0.116760 | +0.001761 | 1.0153 |
| src/compiler/core.ts:2377:1 | createGetCanonicalFileName | 0.034876 | 0.036633 | +0.001756 | 1.0504 |
| src/compiler/utilities.ts:9996:1 | getSupportedExtensions | 0.036650 | 0.038397 | +0.001747 | 1.0477 |
| src/compiler/utilities.ts:9302:1 |  | 0.073719 | 0.075464 | +0.001745 | 1.0237 |
| src/compiler/binder.ts:509:1 | createBinder | 0.049761 | 0.051503 | +0.001742 | 1.0350 |
| src/compiler/utilities.ts:7082:1 | writeCommentRange | 0.224351 | 0.226089 | +0.001738 | 1.0077 |
| src/compiler/factory/emitHelpers.ts:96:1 | ESDecorateContext | 0.035334 | 0.037071 | +0.001737 | 1.0492 |
| src/compiler/types.ts:3191:1 | JsxElement | 0.043145 | 0.044882 | +0.001737 | 1.0403 |
| src/compiler/tsbuild.ts:58:1 | Status | 0.035816 | 0.037550 | +0.001733 | 1.0484 |
| src/compiler/moduleNameResolver.ts:755:1 | getConditions | 0.035706 | 0.037433 | +0.001727 | 1.0484 |
| src/compiler/types.ts:851:1 | ModifierFlags | 0.035469 | 0.037195 | +0.001726 | 1.0487 |
| src/compiler/builder.ts:1192:1 | isNonIncrementalBuildInfo | 0.038250 | 0.039970 | +0.001720 | 1.0450 |
| src/compiler/types.ts:10023:1 | ModulePath | 0.036456 | 0.038176 | +0.001720 | 1.0472 |
| src/compiler/types.ts:1:1 |  | 0.036949 | 0.038665 | +0.001715 | 1.0464 |
| src/compiler/factory/nodeTests.ts:1082:1 | isJSDoc | 0.041222 | 0.042937 | +0.001715 | 1.0416 |
| src/compiler/factory/nodeTests.ts:1140:1 | isJSDocSeeTag | 0.041381 | 0.043096 | +0.001715 | 1.0414 |
| src/compiler/factory/nodeTests.ts:1096:1 | isJSDocAugmentsTag | 0.039929 | 0.041633 | +0.001704 | 1.0427 |
| src/compiler/resolutionCache.ts:180:1 | CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations | 0.036324 | 0.038020 | +0.001696 | 1.0467 |
| src/compiler/types.ts:8592:1 | EmitHint | 0.034722 | 0.036418 | +0.001696 | 1.0488 |
| src/compiler/semver.ts:266:1 |  | 0.037465 | 0.039159 | +0.001694 | 1.0452 |
| src/compiler/utilities.ts:10300:1 | isAnySupportedFileExtension | 0.036504 | 0.038187 | +0.001683 | 1.0461 |
| src/compiler/types.ts:4122:1 | JSDocTypeLiteral | 0.035144 | 0.036826 | +0.001682 | 1.0479 |
| src/compiler/checker.ts:1326:1 | TypeSystemPropertyName | 0.038659 | 0.040337 | +0.001678 | 1.0434 |
| src/compiler/watch.ts:234:1 | getFilesInErrorForSummary | 0.035099 | 0.036770 | +0.001671 | 1.0476 |
| src/compiler/factory/nodeTests.ts:408:1 | isParameter | 0.040500 | 0.042166 | +0.001666 | 1.0411 |
| src/compiler/types.ts:2662:1 | ObjectDestructuringAssignment | 0.037751 | 0.039414 | +0.001663 | 1.0441 |
| src/compiler/types.ts:8111:1 | ResolvedTypeReferenceDirective | 0.035485 | 0.037148 | +0.001663 | 1.0469 |
| src/compiler/types.ts:2991:1 | SuperPropertyAccessExpression | 0.036171 | 0.037826 | +0.001655 | 1.0457 |
| src/compiler/types.ts:2703:1 | ArrayAssignmentElement | 0.037487 | 0.039140 | +0.001652 | 1.0441 |
| src/compiler/factory/nodeChildren.ts:14:1 | getNodeChildren | 0.163841 | 0.165487 | +0.001647 | 1.0101 |
| src/compiler/transformers/utilities.ts:602:1 | getStaticPropertiesAndClassStaticBlock | 0.037910 | 0.039556 | +0.001646 | 1.0434 |
| src/compiler/watchUtilities.ts:83:1 | FileAndDirectoryExistence | 0.035183 | 0.036829 | +0.001646 | 1.0468 |
| src/compiler/utilitiesPublic.ts:764:1 | getOriginalNode | 0.037471 | 0.039113 | +0.001642 | 1.0438 |
| src/compiler/types.ts:4702:1 | LibResolution | 0.035356 | 0.036994 | +0.001637 | 1.0463 |
| src/compiler/factory/nodeFactory.ts:481:1 | addNodeFactoryPatcher | 0.035039 | 0.036676 | +0.001636 | 1.0467 |
| src/compiler/factory/nodeTests.ts:312:1 | isEqualsGreaterThanToken | 0.042832 | 0.044460 | +0.001628 | 1.0380 |
| src/compiler/moduleNameResolver.ts:1768:1 | bundlerModuleNameResolver | 0.043318 | 0.044937 | +0.001619 | 1.0374 |
| src/compiler/types.ts:7852:1 | CharacterCodes | 0.034827 | 0.036443 | +0.001616 | 1.0464 |
| src/compiler/utilities.ts:8099:1 | accessKind | 0.080839 | 0.082452 | +0.001614 | 1.0200 |
| src/compiler/scanner.ts:342:1 |  | 0.041922 | 0.043532 | +0.001610 | 1.0384 |
| src/compiler/types.ts:3804:1 | ImportSpecifier | 0.036278 | 0.037886 | +0.001609 | 1.0443 |
| src/compiler/path.ts:884:1 | comparePathsWorker | 0.035703 | 0.037305 | +0.001602 | 1.0449 |
| src/compiler/sourcemap.ts:371:1 | LineInfo | 0.038700 | 0.040299 | +0.001599 | 1.0413 |
| src/compiler/moduleNameResolver.ts:1986:1 | nodeLoadModuleByRelativeName | 0.035650 | 0.037248 | +0.001598 | 1.0448 |
| src/compiler/utilities.ts:11208:1 | getNameFromImportAttribute | 0.189495 | 0.191089 | +0.001595 | 1.0084 |
| src/compiler/types.ts:3500:1 | LabeledStatement | 0.034949 | 0.036544 | +0.001595 | 1.0456 |
| src/compiler/parser.ts:490:1 | walkTreeForImportMeta | 0.051614 | 0.053208 | +0.001595 | 1.0309 |
| src/compiler/utilities.ts:1360:1 | getEmitFlags | 0.040121 | 0.041715 | +0.001594 | 1.0397 |
| src/compiler/factory/nodeTests.ts:377:1 | isSuperKeyword | 0.039454 | 0.041048 | +0.001594 | 1.0404 |
| src/compiler/types.ts:3158:1 | AssertionExpression | 0.036531 | 0.038124 | +0.001594 | 1.0436 |
| src/compiler/types.ts:7182:1 | IntraExpressionInferenceSite | 0.035711 | 0.037304 | +0.001593 | 1.0446 |
| src/compiler/utilities.ts:2445:1 | assertDiagnosticLocation | 0.038001 | 0.039587 | +0.001586 | 1.0417 |
| src/compiler/parser.ts:405:1 | SignatureFlags | 0.035148 | 0.036731 | +0.001583 | 1.0450 |
| src/compiler/scanner.ts:333:1 |  | 0.039407 | 0.040983 | +0.001577 | 1.0400 |
| src/compiler/types.ts:9768:1 | VisitResult | 0.036131 | 0.037707 | +0.001576 | 1.0436 |
| src/compiler/utilities.ts:8754:1 | chainDiagnosticMessages | 0.066292 | 0.067867 | +0.001576 | 1.0238 |
| src/compiler/emitter.ts:1:1 |  | 0.035285 | 0.036860 | +0.001574 | 1.0446 |
| src/compiler/commandLineParser.ts:3018:1 | parseJsonConfigFileContent | 0.035202 | 0.036774 | +0.001572 | 1.0447 |
| src/compiler/scanner.ts:475:1 | computePositionOfLineAndCharacter | 0.037577 | 0.039142 | +0.001566 | 1.0417 |
| src/compiler/corePublic.ts:21:1 | SortedArray | 0.034666 | 0.036225 | +0.001560 | 1.0450 |
| src/compiler/transformers/utilities.ts:1:1 |  | 0.037783 | 0.039339 | +0.001557 | 1.0412 |
| src/compiler/checker.ts:1397:1 | MembersOrExportsResolutionKind | 0.038072 | 0.039626 | +0.001554 | 1.0408 |
| src/compiler/core.ts:19:1 |  | 0.035536 | 0.037089 | +0.001553 | 1.0437 |
| src/compiler/watch.ts:266:1 | prettyPathForFileError | 0.034345 | 0.035897 | +0.001552 | 1.0452 |
| src/compiler/commandLineParser.ts:1698:1 |  | 4.584694 | 4.586244 | +0.001550 | 1.0003 |
| src/compiler/binder.ts:312:1 |  | 0.036424 | 0.037973 | +0.001550 | 1.0425 |
| src/compiler/transformers/namedEvaluation.ts:1:1 |  | 0.037718 | 0.039265 | +0.001548 | 1.0410 |
| src/compiler/moduleSpecifiers.ts:679:1 | packageJsonPathsAreEqual | 0.042092 | 0.043639 | +0.001547 | 1.0368 |
| src/compiler/path.ts:373:1 | getBaseFileName | 0.038088 | 0.039632 | +0.001544 | 1.0406 |
| src/compiler/executeCommandLine.ts:754:1 | ExecuteCommandLineCallbacks | 0.035284 | 0.036816 | +0.001532 | 1.0434 |
| src/compiler/types.ts:7654:1 | LineAndCharacter | 0.034980 | 0.036503 | +0.001522 | 1.0435 |
| src/compiler/utilitiesPublic.ts:2564:1 | hasType | 0.039456 | 0.040976 | +0.001520 | 1.0385 |
| src/compiler/types.ts:3518:1 | CatchClause | 0.034556 | 0.036075 | +0.001519 | 1.0440 |
| src/compiler/commandLineParser.ts:2167:1 |  | 0.064795 | 0.066308 | +0.001512 | 1.0233 |
| src/compiler/utilities.ts:7332:1 | getEffectiveModifierFlagsNoCache | 0.046216 | 0.047726 | +0.001510 | 1.0327 |
| src/compiler/types.ts:9964:1 | DocumentPosition | 0.035702 | 0.037211 | +0.001509 | 1.0423 |
| src/compiler/watch.ts:974:1 | IncrementalCompilationOptions | 0.036521 | 0.038029 | +0.001508 | 1.0413 |
| src/compiler/utilitiesPublic.ts:2648:1 | isInternalDeclaration | 0.054821 | 0.056329 | +0.001508 | 1.0275 |
| src/compiler/commandLineParser.ts:2277:1 | parseConfigFileTextToJson | 0.038325 | 0.039831 | +0.001505 | 1.0393 |
| src/compiler/moduleSpecifiers.ts:164:1 | RelativePreference | 0.034953 | 0.036455 | +0.001503 | 1.0430 |
| src/compiler/sys.ts:573:1 | setSysLog | 0.037706 | 0.039208 | +0.001502 | 1.0398 |
| src/compiler/moduleNameResolver.ts:2694:1 | comparePatternKeys | 0.035655 | 0.037156 | +0.001501 | 1.0421 |
| src/compiler/utilitiesPublic.ts:2542:1 | isGetAccessor | 0.039373 | 0.040873 | +0.001499 | 1.0381 |
| src/compiler/factory/emitHelpers.ts:1211:1 |  | 0.066750 | 0.068244 | +0.001494 | 1.0224 |
| src/compiler/factory/nodeTests.ts:672:1 | isExpressionWithTypeArguments | 0.039657 | 0.041143 | +0.001486 | 1.0375 |
| src/compiler/utilitiesPublic.ts:403:1 | textSpanIntersectsWithPosition | 0.034733 | 0.036210 | +0.001477 | 1.0425 |
| src/compiler/utilities.ts:3948:1 | getDeclaredExpandoInitializer | 0.160105 | 0.161581 | +0.001476 | 1.0092 |
| src/compiler/utilitiesPublic.ts:465:1 | textChangeRangeIsUnchanged | 0.036166 | 0.037642 | +0.001476 | 1.0408 |
| src/compiler/factory/utilities.ts:829:1 | tryRenameExternalModule | 0.082354 | 0.083827 | +0.001473 | 1.0179 |
| src/compiler/builder.ts:1616:1 | getTextHandlingSourceMapForSignature | 4.635620 | 4.637086 | +0.001466 | 1.0003 |
| src/compiler/types.ts:1399:1 | PrimitiveLiteral | 0.035752 | 0.037217 | +0.001465 | 1.0410 |
| src/compiler/emitter.ts:1209:1 |  | 0.039488 | 0.040952 | +0.001464 | 1.0371 |
| src/compiler/utilities.ts:5043:1 | skipParentheses | 0.034533 | 0.035991 | +0.001457 | 1.0422 |
| src/compiler/path.ts:785:1 | removeTrailingDirectorySeparator | 0.034285 | 0.035736 | +0.001451 | 1.0423 |
| src/compiler/moduleNameResolver.ts:2026:1 |  | 0.035899 | 0.037350 | +0.001451 | 1.0404 |
| src/compiler/types.ts:4019:1 | JSDocPublicTag | 0.037231 | 0.038678 | +0.001447 | 1.0389 |
| src/compiler/emitter.ts:6340:1 | getClosingBracket | 0.037385 | 0.038828 | +0.001443 | 1.0386 |
| src/compiler/commandLineParser.ts:2289:1 | readJsonConfigFile | 0.071630 | 0.073072 | +0.001442 | 1.0201 |
| src/compiler/types.ts:8223:1 | TransformFlags | 0.035135 | 0.036570 | +0.001435 | 1.0408 |
| src/compiler/utilitiesPublic.ts:1441:1 | isBreakOrContinueStatement | 0.041154 | 0.042587 | +0.001433 | 1.0348 |
| src/compiler/utilities.ts:8740:1 | createCompilerDiagnosticFromMessageChain | 0.037106 | 0.038529 | +0.001423 | 1.0384 |
| src/compiler/scanner.ts:952:1 | appendCommentRange | 0.035047 | 0.036467 | +0.001420 | 1.0405 |
| src/compiler/types.ts:1607:1 | QuestionToken | 0.036502 | 0.037919 | +0.001418 | 1.0388 |
| src/compiler/program.ts:691:1 |  | 0.035042 | 0.036457 | +0.001415 | 1.0404 |
| src/compiler/utilities.ts:4441:1 | isDefaultImport | 0.037117 | 0.038529 | +0.001412 | 1.0380 |
| src/compiler/factory/nodeTests.ts:454:1 | isConstructSignatureDeclaration | 0.040495 | 0.041903 | +0.001408 | 1.0348 |
| src/compiler/scanner.ts:738:1 |  | 0.035291 | 0.036698 | +0.001407 | 1.0399 |
| src/compiler/utilities.ts:919:1 | typeDirectiveIsEqualTo | 0.038653 | 0.040057 | +0.001404 | 1.0363 |
| src/compiler/factory/nodeTests.ts:329:1 | isExportModifier | 0.041523 | 0.042918 | +0.001395 | 1.0336 |
| src/compiler/sourcemap.ts:699:1 | createDocumentPositionMapper | 0.040581 | 0.041973 | +0.001392 | 1.0343 |
| src/compiler/performance.ts:20:1 |  | 0.035254 | 0.036643 | +0.001389 | 1.0394 |
| src/compiler/factory/nodeTests.ts:512:1 | isIntersectionTypeNode | 0.041694 | 0.043076 | +0.001382 | 1.0331 |
| src/compiler/visitorPublic.ts:393:1 | visitParameterList | 0.035607 | 0.036984 | +0.001378 | 1.0387 |
| src/compiler/factory/nodeFactory.ts:7394:1 | makeSynthetic | 0.042349 | 0.043726 | +0.001377 | 1.0325 |
| src/compiler/scanner.ts:961:1 | getTrailingCommentRanges | 0.038395 | 0.039769 | +0.001374 | 1.0358 |
| src/compiler/utilities.ts:4288:1 | PrototypePropertyAssignment | 0.036624 | 0.037996 | +0.001372 | 1.0375 |
| src/compiler/utilities.ts:8522:1 | Token | 0.045257 | 0.046624 | +0.001367 | 1.0302 |
| src/compiler/corePublic.ts:30:1 | ReadonlyCollection | 0.035564 | 0.036929 | +0.001366 | 1.0384 |
| src/compiler/commandLineParser.ts:1872:1 | createDiagnosticForInvalidCustomType | 0.038442 | 0.039802 | +0.001360 | 1.0354 |
| src/compiler/utilities.ts:2349:1 | declarationNameToString | 0.045010 | 0.046367 | +0.001357 | 1.0301 |
| src/compiler/utilitiesPublic.ts:1471:1 | isTokenKind | 0.037444 | 0.038801 | +0.001357 | 1.0362 |
| src/compiler/utilities.ts:8019:1 | getInitializedVariables | 0.042504 | 0.043860 | +0.001355 | 1.0319 |
| src/compiler/program.ts:692:1 |  | 0.035745 | 0.037101 | +0.001355 | 1.0379 |
| src/compiler/commandLineParser.ts:1707:1 |  | 4.452202 | 4.453557 | +0.001355 | 1.0003 |
| src/compiler/types.ts:1599:1 | EndOfFileToken | 0.036188 | 0.037543 | +0.001355 | 1.0374 |
| src/compiler/factory/nodeTests.ts:1062:1 | isJSDocNonNullableType | 0.041078 | 0.042432 | +0.001354 | 1.0330 |
| src/compiler/types.ts:2858:1 | NumericLiteral | 0.036438 | 0.037791 | +0.001353 | 1.0371 |
| src/compiler/tsbuild.ts:41:1 | UpToDateStatus | 0.036181 | 0.037534 | +0.001353 | 1.0374 |
| src/compiler/builder.ts:115:1 | ReusableRepopulateInfoChain | 0.035869 | 0.037220 | +0.001350 | 1.0376 |
| src/compiler/utilities.ts:5348:1 | isAsyncFunction | 0.040441 | 0.041791 | +0.001349 | 1.0334 |
| src/compiler/utilities.ts:8035:1 | closeFileWatcher | 0.071639 | 0.072980 | +0.001341 | 1.0187 |
| src/compiler/types.ts:9770:1 | Printer | 0.035730 | 0.037068 | +0.001338 | 1.0375 |
| src/compiler/watchPublic.ts:160:1 | WatchHost | 0.035592 | 0.036929 | +0.001338 | 1.0376 |
| src/compiler/types.ts:940:1 | NodeId | 0.037037 | 0.038369 | +0.001333 | 1.0360 |
| src/compiler/programDiagnostics.ts:89:1 | createProgramDiagnostics | 0.098700 | 0.100027 | +0.001327 | 1.0134 |
| src/compiler/factory/nodeTests.ts:985:1 | isHeritageClause | 0.041458 | 0.042783 | +0.001325 | 1.0320 |
| src/compiler/factory/nodeTests.ts:1192:1 | isJSDocImportTag | 0.041034 | 0.042355 | +0.001322 | 1.0322 |
| src/compiler/utilities.ts:10284:1 | resolutionExtensionIsTSOrJson | 0.037000 | 0.038321 | +0.001321 | 1.0357 |
| src/compiler/transformers/declarations.ts:1916:1 | isPreservedDeclarationStatement | 0.039993 | 0.041312 | +0.001319 | 1.0330 |
| src/compiler/watchPublic.ts:408:1 | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile | 0.035102 | 0.036417 | +0.001315 | 1.0375 |
| src/compiler/types.ts:516:1 | PunctuationSyntaxKind | 0.035435 | 0.036746 | +0.001310 | 1.0370 |
| src/compiler/checker.ts:1324:1 | TypeSystemEntity | 0.037031 | 0.038336 | +0.001305 | 1.0352 |
| src/compiler/types.ts:10074:1 | TextSpan | 0.036126 | 0.037426 | +0.001300 | 1.0360 |
| src/compiler/semver.ts:1:1 |  | 0.037926 | 0.039223 | +0.001297 | 1.0342 |
| src/compiler/utilities.ts:6493:1 | ResolveModuleNameResolutionHost | 0.035619 | 0.036915 | +0.001296 | 1.0364 |
| src/compiler/utilities.ts:8179:1 | MutateMapSkippingNewValuesDelete | 0.035577 | 0.036872 | +0.001295 | 1.0364 |
| src/compiler/parser.ts:424:1 |  | 0.036475 | 0.037769 | +0.001294 | 1.0355 |
| src/compiler/utilitiesPublic.ts:1462:1 | isNodeKind | 0.037569 | 0.038860 | +0.001291 | 1.0344 |
| src/compiler/types.ts:10207:1 | PragmaKindFlags | 0.035885 | 0.037173 | +0.001288 | 1.0359 |
| src/compiler/factory/emitNode.ts:271:1 | removeEmitHelper | 0.040686 | 0.041966 | +0.001280 | 1.0315 |
| src/compiler/types.ts:7139:1 | InferenceFlags | 0.036058 | 0.037337 | +0.001279 | 1.0355 |
| src/compiler/factory/nodeTests.ts:1042:1 | isJSDocLinkCode | 0.041335 | 0.042610 | +0.001275 | 1.0309 |
| src/compiler/utilities.ts:1290:1 | isJSDocTypeExpressionOrChild | 0.043715 | 0.044988 | +0.001273 | 1.0291 |
| src/compiler/_namespaces/ts.ts:39:1 |  | 0.036902 | 0.038173 | +0.001271 | 1.0345 |
| src/compiler/utilities.ts:3953:1 | hasExpandoValueProperty | 0.047982 | 0.049252 | +0.001271 | 1.0265 |
| src/compiler/factory/emitHelpers.ts:1380:1 |  | 0.064058 | 0.065325 | +0.001267 | 1.0198 |
| src/compiler/types.ts:5520:1 | ContextFlags | 0.034697 | 0.035963 | +0.001265 | 1.0365 |
| src/compiler/types.ts:4671:1 | FilePreprocessingLibReferenceDiagnostic | 0.035492 | 0.036751 | +0.001259 | 1.0355 |
| src/compiler/scanner.ts:800:1 | scanShebangTrivia | 0.035816 | 0.037066 | +0.001251 | 1.0349 |
| src/compiler/builder.ts:2440:1 | getNonIncrementalBuildInfoRoots | 0.036367 | 0.037618 | +0.001251 | 1.0344 |
| src/compiler/utilities.ts:7977:1 | getLinesBetweenPositionAndNextNonWhitespaceCharacter | 0.039136 | 0.040387 | +0.001251 | 1.0320 |
| src/compiler/tsbuild.ts:10:1 | UpToDateStatusType | 0.037455 | 0.038704 | +0.001249 | 1.0333 |
| src/compiler/corePublic.ts:3:1 |  | 0.036387 | 0.037634 | +0.001247 | 1.0343 |
| src/compiler/utilities.ts:8770:1 | concatenateDiagnosticMessageChains | 0.035557 | 0.036801 | +0.001243 | 1.0350 |
| src/compiler/types.ts:1519:1 | HasLocals | 0.035529 | 0.036772 | +0.001243 | 1.0350 |
| src/compiler/transformers/destructuring.ts:218:1 | bindingOrAssignmentElementContainsNonLiteralComputedName | 0.307166 | 0.308408 | +0.001241 | 1.0040 |
| src/compiler/utilities.ts:7674:1 | convertToBase64 | 0.037533 | 0.038772 | +0.001239 | 1.0330 |
| src/compiler/types.ts:2512:1 | SyntheticExpression | 0.038128 | 0.039360 | +0.001232 | 1.0323 |
| src/compiler/types.ts:3530:1 | DeclarationWithTypeParameters | 0.036153 | 0.037385 | +0.001231 | 1.0341 |
| src/compiler/builderState.ts:56:1 | SignatureInfo | 0.035373 | 0.036585 | +0.001212 | 1.0343 |
| src/compiler/semver.ts:451:1 | formatComparator | 0.035735 | 0.036944 | +0.001208 | 1.0338 |
| src/compiler/factory/nodeTests.ts:1148:1 | isJSDocParameterTag | 0.041744 | 0.042951 | +0.001208 | 1.0289 |
| src/compiler/moduleNameResolver.ts:922:1 | PackageJsonInfoCacheEntry | 0.036065 | 0.037265 | +0.001201 | 1.0333 |
| src/compiler/binder.ts:502:1 | bindSourceFile | 0.036101 | 0.037301 | +0.001200 | 1.0332 |
| src/compiler/factory/nodeTests.ts:989:1 | isCatchClause | 0.040680 | 0.041880 | +0.001200 | 1.0295 |
| src/compiler/binder.ts:1:1 |  | 0.036487 | 0.037684 | +0.001197 | 1.0328 |
| src/compiler/types.ts:3378:1 | VariableStatement | 0.036923 | 0.038120 | +0.001197 | 1.0324 |
| src/compiler/parser.ts:1321:1 | CreateSourceFileOptions | 0.037140 | 0.038335 | +0.001196 | 1.0322 |
| src/compiler/types.ts:10055:1 | SymbolTracker | 0.036516 | 0.037711 | +0.001195 | 1.0327 |
| src/compiler/types.ts:7704:1 | LanguageVariant | 0.035087 | 0.036279 | +0.001192 | 1.0340 |
| src/compiler/utilities.ts:8355:1 | isAccessExpression | 0.043190 | 0.044382 | +0.001191 | 1.0276 |
| src/compiler/factory/nodeTests.ts:1184:1 | isJSDocSatisfiesTag | 0.041824 | 0.043013 | +0.001189 | 1.0284 |
| src/compiler/types.ts:3009:1 | ElementAccessChain | 0.036090 | 0.037273 | +0.001183 | 1.0328 |
| src/compiler/path.ts:1113:1 | isNodeModulesDirectory | 0.036193 | 0.037372 | +0.001179 | 1.0326 |
| src/compiler/transformers/generators.ts:245:1 | CodeBlockKind | 0.037686 | 0.038862 | +0.001176 | 1.0312 |
| src/compiler/utilities.ts:7226:1 | hasAccessorModifier | 0.043090 | 0.044265 | +0.001175 | 1.0273 |
| src/compiler/core.ts:2053:1 | compareStringsCaseInsensitiveEslintCompatible | 0.039363 | 0.040538 | +0.001174 | 1.0298 |
| src/compiler/program.ts:651:1 | FormatDiagnosticsHost | 0.034850 | 0.036024 | +0.001173 | 1.0337 |
| src/compiler/factory/nodeTests.ts:1038:1 | isJSDocLink | 0.040777 | 0.041946 | +0.001169 | 1.0287 |
| src/compiler/emitter.ts:618:1 | getOwnOutputFileNames | 0.038105 | 0.039273 | +0.001169 | 1.0307 |
| src/compiler/types.ts:2092:1 | MethodSignature | 0.059035 | 0.060204 | +0.001169 | 1.0198 |
| src/compiler/utilities.ts:10305:1 | tryGetExtensionFromPath | 0.038005 | 0.039172 | +0.001167 | 1.0307 |
| src/compiler/types.ts:4286:1 | SourceFileLike | 0.035850 | 0.037012 | +0.001162 | 1.0324 |
| src/compiler/path.ts:967:1 | containsPath | 0.035075 | 0.036236 | +0.001161 | 1.0331 |
| src/compiler/core.ts:2437:1 | removePrefix | 0.035931 | 0.037083 | +0.001153 | 1.0321 |
| src/compiler/factory/nodeTests.ts:284:1 | isPlusToken | 0.042352 | 0.043494 | +0.001141 | 1.0270 |
| src/compiler/sys.ts:60:1 | generateDjb2Hash | 0.035411 | 0.036549 | +0.001138 | 1.0321 |
| src/compiler/path.ts:963:1 | containsPath | 0.035520 | 0.036658 | +0.001138 | 1.0320 |
| src/compiler/visitorPublic.ts:223:1 | visitNodes | 0.036112 | 0.037249 | +0.001136 | 1.0315 |
| src/compiler/utilities.ts:10030:1 | hasJSFileExtension | 0.037955 | 0.039091 | +0.001135 | 1.0299 |
| src/compiler/types.ts:4695:1 | EmitOnly | 0.035343 | 0.036476 | +0.001133 | 1.0321 |
| src/compiler/sys.ts:103:1 | HostWatchFile | 0.034518 | 0.035650 | +0.001133 | 1.0328 |
| src/compiler/emitter.ts:573:1 | getOutputDeclarationFileNameWorker | 0.035478 | 0.036605 | +0.001127 | 1.0318 |
| src/compiler/core.ts:2375:1 | GetCanonicalFileName | 0.034888 | 0.036015 | +0.001127 | 1.0323 |
| src/compiler/executeCommandLine.ts:1146:1 | enableStatisticsAndTracing | 0.034981 | 0.036107 | +0.001127 | 1.0322 |
| src/compiler/_namespaces/ts.ts:55:1 |  | 0.038924 | 0.040050 | +0.001126 | 1.0289 |
| src/compiler/utilities.ts:1321:1 | getTextOfNodeFromSourceText | 0.045552 | 0.046676 | +0.001124 | 1.0247 |
| src/compiler/sys.ts:49:1 |  | 0.035462 | 0.036584 | +0.001123 | 1.0317 |
| src/compiler/path.ts:828:1 | ensurePathIsNonModuleName | 0.034555 | 0.035675 | +0.001120 | 1.0324 |
| src/compiler/moduleNameResolver.ts:1778:1 | nodeModuleNameResolver | 0.036912 | 0.038029 | +0.001117 | 1.0303 |
| src/compiler/factory/nodeTests.ts:468:1 | isTypeReferenceNode | 0.043092 | 0.044207 | +0.001115 | 1.0259 |
| src/compiler/factory/utilities.ts:1142:1 | canHaveIllegalModifiers | 0.043115 | 0.044229 | +0.001114 | 1.0258 |
| src/compiler/program.ts:635:18 | getPreEmitDiagnostics | 0.037405 | 0.038518 | +0.001113 | 1.0297 |
| src/compiler/watch.ts:260:1 | getWatchErrorSummaryDiagnosticMessage | 0.034541 | 0.035648 | +0.001107 | 1.0320 |
| src/compiler/utilities.ts:8779:1 | getDiagnosticFilePath | 0.034326 | 0.035431 | +0.001105 | 1.0322 |
| src/compiler/program.ts:1075:1 |  | 0.089157 | 0.090261 | +0.001104 | 1.0124 |
| src/compiler/watch.ts:229:1 | getErrorCountForSummary | 0.035979 | 0.037073 | +0.001095 | 1.0304 |
| src/compiler/types.ts:1772:1 | DynamicNamedBinaryExpression | 0.034840 | 0.035933 | +0.001093 | 1.0314 |
| src/compiler/types.ts:1638:1 | ExportKeyword | 0.034926 | 0.036016 | +0.001090 | 1.0312 |
| src/compiler/types.ts:3511:1 | TryStatement | 0.034467 | 0.035554 | +0.001087 | 1.0315 |
| src/compiler/types.ts:3207:1 | JsxAttributeLike | 0.067749 | 0.068835 | +0.001086 | 1.0160 |
| src/compiler/performance.ts:34:1 | createTimer | 0.036671 | 0.037755 | +0.001084 | 1.0296 |
| src/compiler/watchPublic.ts:359:1 | WatchOfFilesAndCompilerOptions | 0.034981 | 0.036065 | +0.001084 | 1.0310 |
| src/compiler/utilities.ts:9310:1 |  | 0.076070 | 0.077154 | +0.001084 | 1.0142 |
| src/compiler/types.ts:2965:1 | EntityNameExpression | 0.036284 | 0.037366 | +0.001082 | 1.0298 |
| src/compiler/factory/nodeFactory.ts:478:1 |  | 0.034315 | 0.035397 | +0.001082 | 1.0315 |
| src/compiler/utilities.ts:2246:1 | isDeclarationWithTypeParameterChildren | 0.107951 | 0.109029 | +0.001078 | 1.0100 |
| src/compiler/types.ts:3065:1 | LiteralLikeElementAccessExpression | 0.036282 | 0.037360 | +0.001077 | 1.0297 |
| src/compiler/utilities.ts:10225:1 |  | 0.038478 | 0.039555 | +0.001077 | 1.0280 |
| src/compiler/transformers/es2015.ts:478:1 | SpreadSegment | 0.036543 | 0.037613 | +0.001070 | 1.0293 |
| src/compiler/types.ts:3705:1 | NamedImportBindings | 0.034963 | 0.036031 | +0.001069 | 1.0306 |
| src/compiler/types.ts:2670:1 | DestructuringAssignment | 0.037502 | 0.038567 | +0.001065 | 1.0284 |
| src/compiler/_namespaces/ts.ts:61:1 |  | 0.036087 | 0.037148 | +0.001061 | 1.0294 |
| src/compiler/types.ts:4295:1 | FutureSourceFile | 0.035362 | 0.036421 | +0.001059 | 1.0300 |
| src/compiler/utilities.ts:9292:1 |  | 0.074216 | 0.075272 | +0.001056 | 1.0142 |
| src/compiler/builderPublic.ts:162:1 | createSemanticDiagnosticsBuilderProgram | 0.058495 | 0.059549 | +0.001053 | 1.0180 |
| src/compiler/performance.ts:29:1 | createTimerIf | 0.036220 | 0.037267 | +0.001046 | 1.0289 |
| src/compiler/utilitiesPublic.ts:1445:1 | isNamedExportBindings | 0.041525 | 0.042559 | +0.001035 | 1.0249 |
| src/compiler/utilities.ts:9985:1 |  | 0.036775 | 0.037803 | +0.001028 | 1.0280 |
| src/compiler/scanner.ts:794:1 | isShebangTrivia | 0.035150 | 0.036176 | +0.001026 | 1.0292 |
| src/compiler/executeCommandLine.ts:240:1 | getDisplayNameTextOfOption | 0.036989 | 0.038012 | +0.001023 | 1.0277 |
| src/compiler/transformers/classFields.ts:288:1 | PrivateIdentifierStaticFieldInfo | 0.037544 | 0.038567 | +0.001022 | 1.0272 |
| src/compiler/commandLineParser.ts:3738:1 | convertCompilerOptionsFromJson | 0.039818 | 0.040839 | +0.001022 | 1.0257 |
| src/compiler/utilitiesPublic.ts:1449:1 | isJSDocPropertyLikeTag | 0.041163 | 0.042184 | +0.001021 | 1.0248 |
| src/compiler/utilitiesPublic.ts:787:1 | findAncestor | 0.036914 | 0.037930 | +0.001015 | 1.0275 |
| src/compiler/parser.ts:423:1 |  | 0.035859 | 0.036874 | +0.001015 | 1.0283 |
| src/compiler/performance.ts:23:1 | Timer | 0.036001 | 0.037012 | +0.001012 | 1.0281 |
| src/compiler/tsbuildPublic.ts:854:1 | BuildInvalidedProject | 0.035627 | 0.036633 | +0.001006 | 1.0282 |
| src/compiler/programDiagnostics.ts:65:1 | LazyConfigDiagnostic | 0.036440 | 0.037443 | +0.001003 | 1.0275 |
| src/compiler/types.ts:7649:1 | NewLineKind | 0.034766 | 0.035759 | +0.000993 | 1.0286 |
| src/compiler/moduleSpecifiers.ts:318:1 | tryGetModuleSpecifiersFromCache | 0.040987 | 0.041975 | +0.000988 | 1.0241 |
| src/compiler/binder.ts:465:1 | ContainerFlags | 0.036295 | 0.037279 | +0.000984 | 1.0271 |
| src/compiler/utilitiesPublic.ts:667:1 | getCombinedNodeFlagsAlwaysIncludeJSDoc | 0.043580 | 0.044562 | +0.000982 | 1.0225 |
| src/compiler/types.ts:3834:1 | TypeOnlyImportDeclaration | 0.035710 | 0.036690 | +0.000980 | 1.0274 |
| src/compiler/types.ts:2221:1 | FunctionOrConstructorTypeNode | 0.038094 | 0.039074 | +0.000979 | 1.0257 |
| src/compiler/utilities.ts:6559:1 | getDeclarationEmitOutputFilePathWorker | 0.151486 | 0.152464 | +0.000978 | 1.0065 |
| src/compiler/factory/nodeTests.ts:1152:1 | isJSDocReturnTag | 0.040847 | 0.041819 | +0.000972 | 1.0238 |
| src/compiler/types.ts:3896:1 | JSDocMemberName | 0.034484 | 0.035455 | +0.000971 | 1.0282 |
| src/compiler/types.ts:7841:1 | CommandLineOptionOfListType | 0.034958 | 0.035929 | +0.000971 | 1.0278 |
| src/compiler/types.ts:3732:1 | AssertionKey | 0.034680 | 0.035649 | +0.000969 | 1.0279 |
| src/compiler/visitorPublic.ts:394:1 | visitParameterList | 0.035947 | 0.036913 | +0.000967 | 1.0269 |
| src/compiler/commandLineParser.ts:3685:1 | ExtendedConfigCacheEntry | 0.039174 | 0.040140 | +0.000966 | 1.0247 |
| src/compiler/utilities.ts:7200:1 | isStatic | 0.042572 | 0.043532 | +0.000960 | 1.0225 |
| src/compiler/types.ts:3448:1 | ContinueStatement | 0.036604 | 0.037552 | +0.000948 | 1.0259 |
| src/compiler/tracing.ts:38:1 | tracingEnabled | 0.035425 | 0.036373 | +0.000948 | 1.0267 |
| src/compiler/types.ts:3394:1 | IfStatement | 0.035961 | 0.036907 | +0.000946 | 1.0263 |
| src/compiler/types.ts:8718:1 | NodeConverters | 0.036816 | 0.037760 | +0.000945 | 1.0257 |
| src/compiler/commandLineParser.ts:2162:1 |  | 0.063703 | 0.064648 | +0.000945 | 1.0148 |
| src/compiler/types.ts:2400:1 | OmittedExpression | 0.038287 | 0.039225 | +0.000938 | 1.0245 |
| src/compiler/types.ts:4093:1 | JSDocThrowsTag | 0.036723 | 0.037651 | +0.000928 | 1.0253 |
| src/compiler/types.ts:9804:1 | BuildInfo | 0.037160 | 0.038087 | +0.000928 | 1.0250 |
| src/compiler/utilities.ts:11189:1 | hasResolutionModeOverride | 0.037280 | 0.038206 | +0.000926 | 1.0248 |
| src/compiler/types.ts:4591:1 | CancellationToken | 0.041133 | 0.042058 | +0.000925 | 1.0225 |
| src/compiler/utilities.ts:12059:1 | unwrapParenthesizedExpression | 0.158060 | 0.158972 | +0.000912 | 1.0058 |
| src/compiler/types.ts:8137:1 | HasInvalidatedLibResolutions | 0.035757 | 0.036668 | +0.000911 | 1.0255 |
| src/compiler/types.ts:3852:1 | ExportAssignment | 0.036505 | 0.037415 | +0.000910 | 1.0249 |
| src/compiler/utilities.ts:7236:1 | hasDecorators | 0.042100 | 0.043010 | +0.000910 | 1.0216 |
| src/compiler/executeCommandLine.ts:523:1 | printBuildHelp | 0.038833 | 0.039740 | +0.000908 | 1.0234 |
| src/compiler/tsbuildPublic.ts:211:1 | ReportFileInError | 0.035352 | 0.036257 | +0.000905 | 1.0256 |
| src/compiler/watchUtilities.ts:397:1 | SharedExtendedConfigFileWatcher | 0.035044 | 0.035949 | +0.000905 | 1.0258 |
| src/compiler/_namespaces/ts.ts:60:1 |  | 0.036749 | 0.037652 | +0.000902 | 1.0246 |
| src/compiler/utilitiesPublic.ts:1666:1 | isBindingName | 0.043000 | 0.043902 | +0.000902 | 1.0210 |
| src/compiler/utilities.ts:7825:1 |  | 0.036359 | 0.037259 | +0.000900 | 1.0248 |
| src/compiler/utilitiesPublic.ts:446:1 | createTextSpan | 0.036198 | 0.037097 | +0.000899 | 1.0248 |
| src/compiler/tsbuildPublic.ts:251:1 | CircularBuildOrder | 0.034546 | 0.035426 | +0.000881 | 1.0255 |
| src/compiler/transformers/declarations/diagnostics.ts:101:1 | GetSymbolAccessibilityDiagnostic | 0.039956 | 0.040832 | +0.000876 | 1.0219 |
| src/compiler/types.ts:3453:1 | BreakOrContinueStatement | 0.035614 | 0.036489 | +0.000875 | 1.0246 |
| src/compiler/utilities.ts:7845:1 | createRange | 0.036950 | 0.037821 | +0.000871 | 1.0236 |
| src/compiler/types.ts:4612:1 | RootFile | 0.038571 | 0.039442 | +0.000871 | 1.0226 |
| src/compiler/types.ts:8669:1 | WrappedExpression | 0.039240 | 0.040109 | +0.000869 | 1.0222 |
| src/compiler/utilities.ts:5041:1 | skipParentheses | 0.035394 | 0.036260 | +0.000867 | 1.0245 |
| src/compiler/path.ts:1002:1 | startsWithDirectory | 0.034948 | 0.035813 | +0.000865 | 1.0247 |
| src/compiler/core.ts:2589:1 |  | 0.035195 | 0.036059 | +0.000864 | 1.0246 |
| src/compiler/types.ts:1556:1 | IsInterface | 0.035745 | 0.036607 | +0.000862 | 1.0241 |
| src/compiler/types.ts:9987:1 | GetEffectiveTypeRootsHost | 0.035517 | 0.036377 | +0.000860 | 1.0242 |
| src/compiler/commandLineParser.ts:2995:1 | convertToOptionValueWithAbsolutePaths | 0.036730 | 0.037590 | +0.000860 | 1.0234 |
| src/compiler/types.ts:9660:1 | TransformationResult | 0.038105 | 0.038965 | +0.000859 | 1.0225 |
| src/compiler/types.ts:3729:1 | ImportPhaseModifierSyntaxKind | 0.034404 | 0.035262 | +0.000858 | 1.0249 |
| src/compiler/types.ts:3070:1 | BindableStaticElementAccessExpression | 0.036187 | 0.037044 | +0.000857 | 1.0237 |
| src/compiler/checker.ts:54256:1 | signatureHasLiteralTypes | 0.038542 | 0.039389 | +0.000846 | 1.0220 |
| src/compiler/path.ts:855:1 | changeAnyExtension | 0.034656 | 0.035501 | +0.000845 | 1.0244 |
| src/compiler/types.ts:9751:1 | NodesVisitor | 0.037266 | 0.038106 | +0.000840 | 1.0226 |
| src/compiler/factory/utilities.ts:624:1 | getJSDocTypeAssertionType | 0.050091 | 0.050929 | +0.000838 | 1.0167 |
| src/compiler/types.ts:494:1 | TriviaSyntaxKind | 0.036582 | 0.037420 | +0.000838 | 1.0229 |
| src/compiler/transformers/ts.ts:211:1 | TypeScriptSubstitutionFlags | 0.038615 | 0.039446 | +0.000832 | 1.0215 |
| src/compiler/tsbuild.ts:1:1 |  | 0.036130 | 0.036959 | +0.000830 | 1.0230 |
| src/compiler/utilities.ts:7617:1 | isExportDefaultSymbol | 0.043255 | 0.044081 | +0.000826 | 1.0191 |
| src/compiler/factory/nodeTests.ts:1104:1 | isJSDocClassTag | 0.041180 | 0.042006 | +0.000826 | 1.0201 |
| src/compiler/utilities.ts:9805:1 | getFileMatcherPatterns | 0.050343 | 0.051166 | +0.000824 | 1.0164 |
| src/compiler/sys.ts:96:1 | PollingInterval | 0.034988 | 0.035810 | +0.000822 | 1.0235 |
| src/compiler/scanner.ts:1002:1 | EscapeSequenceScanningFlags | 0.035987 | 0.036809 | +0.000821 | 1.0228 |
| src/compiler/performance.ts:15:1 |  | 0.035804 | 0.036625 | +0.000821 | 1.0229 |
| src/compiler/types.ts:5663:1 | SymbolWriter | 0.034813 | 0.035631 | +0.000818 | 1.0235 |
| src/compiler/emitter.ts:563:1 | getOutputDeclarationFileName | 0.035745 | 0.036560 | +0.000815 | 1.0228 |
| src/compiler/types.ts:2411:1 | UnaryExpression | 0.037966 | 0.038779 | +0.000812 | 1.0214 |
| src/compiler/builder.ts:1570:1 | BuilderProgramKind | 0.035886 | 0.036691 | +0.000806 | 1.0224 |
| src/compiler/performance.ts:1:1 |  | 0.037383 | 0.038184 | +0.000801 | 1.0214 |
| src/compiler/types.ts:2931:1 | JSDocTypeAssertion | 0.036643 | 0.037442 | +0.000799 | 1.0218 |
| src/compiler/types.ts:5070:1 | WriterContextOut | 0.035034 | 0.035830 | +0.000796 | 1.0227 |
| src/compiler/types.ts:3602:1 | EnumMember | 0.035549 | 0.036345 | +0.000795 | 1.0224 |
| src/compiler/utilities.ts:618:1 |  | 0.035493 | 0.036289 | +0.000795 | 1.0224 |
| src/compiler/factory/nodeTests.ts:995:1 | isPropertyAssignment | 0.040305 | 0.041099 | +0.000794 | 1.0197 |
| src/compiler/types.ts:4954:1 | Program | 0.035991 | 0.036784 | +0.000794 | 1.0221 |
| src/compiler/commandLineParser.ts:2157:1 |  | 0.036659 | 0.037452 | +0.000793 | 1.0216 |
| src/compiler/factory/nodeTests.ts:1014:1 | isSourceFile | 0.040667 | 0.041458 | +0.000791 | 1.0195 |
| src/compiler/types.ts:5626:1 | SymbolFormatFlags | 0.036544 | 0.037330 | +0.000786 | 1.0215 |
| src/compiler/sourcemap.ts:408:1 | isRawSourceMap | 0.035271 | 0.036044 | +0.000774 | 1.0219 |
| src/compiler/utilities.ts:8547:1 | SourceMapSource | 0.036495 | 0.037265 | +0.000769 | 1.0211 |
| src/compiler/types.ts:3536:1 | DeclarationWithTypeParameterChildren | 0.034947 | 0.035715 | +0.000768 | 1.0220 |
| src/compiler/sys.ts:568:1 | sysLog | 0.037012 | 0.037775 | +0.000763 | 1.0206 |
| src/compiler/_namespaces/ts.ts:11:1 |  | 0.038927 | 0.039687 | +0.000760 | 1.0195 |
| src/compiler/utilitiesPublic.ts:1989:1 | isTemplateLiteral | 0.039693 | 0.040448 | +0.000755 | 1.0190 |
| src/compiler/types.ts:5705:1 | IdentifierTypePredicate | 0.034805 | 0.035557 | +0.000752 | 1.0216 |
| src/compiler/utilities.ts:11199:1 | replaceFirstStar | 0.038388 | 0.039139 | +0.000751 | 1.0196 |
| src/compiler/builder.ts:209:1 | BuilderProgramState | 0.037786 | 0.038536 | +0.000751 | 1.0199 |
| src/compiler/types.ts:5530:1 | NodeBuilderFlags | 0.035311 | 0.036057 | +0.000745 | 1.0211 |
| src/compiler/builder.ts:1035:1 | getSemanticDiagnosticsOfFile | 0.037117 | 0.037860 | +0.000743 | 1.0200 |
| src/compiler/watchPublic.ts:307:1 | WatchCompilerHostOfConfigFile | 0.035675 | 0.036417 | +0.000742 | 1.0208 |
| src/compiler/utilitiesPublic.ts:2460:1 | isJsxTagNameExpression | 0.039634 | 0.040375 | +0.000741 | 1.0187 |
| src/compiler/types.ts:1634:1 | AsyncKeyword | 0.034257 | 0.034986 | +0.000729 | 1.0213 |
| src/compiler/utilities.ts:12233:1 | getLibFileNameFromLibReference | 0.037192 | 0.037920 | +0.000728 | 1.0196 |
| src/compiler/utilities.ts:9392:1 | compilerOptionsAffectEmit | 0.038075 | 0.038801 | +0.000726 | 1.0191 |
| src/compiler/builder.ts:105:1 | ReusableDiagnosticRelatedInformation | 0.035841 | 0.036559 | +0.000719 | 1.0201 |
| src/compiler/utilities.ts:3337:1 | getSuperContainer | 0.037254 | 0.037970 | +0.000716 | 1.0192 |
| src/compiler/utilities.ts:7667:1 |  | 0.036675 | 0.037391 | +0.000716 | 1.0195 |
| src/compiler/checker.ts:1172:1 |  | 0.035486 | 0.036202 | +0.000715 | 1.0202 |
| src/compiler/moduleNameResolver.ts:155:1 | Resolved | 0.034836 | 0.035548 | +0.000712 | 1.0204 |
| src/compiler/tsbuildPublic.ts:256:1 | AnyBuildOrder | 0.034824 | 0.035529 | +0.000705 | 1.0203 |
| src/compiler/scanner.ts:1:1 |  | 0.039669 | 0.040372 | +0.000704 | 1.0177 |
| src/compiler/utilities.ts:11213:1 | isSourceElement | 0.042720 | 0.043423 | +0.000704 | 1.0165 |
| src/compiler/types.ts:9809:1 | BuildInfoFileVersionMap | 0.038684 | 0.039385 | +0.000701 | 1.0181 |
| src/compiler/types.ts:3443:1 | BreakStatement | 0.035369 | 0.036064 | +0.000695 | 1.0197 |
| src/compiler/scanner.ts:979:1 | isIdentifierPart | 0.036647 | 0.037336 | +0.000689 | 1.0188 |
| src/compiler/binder.ts:499:1 |  | 0.036240 | 0.036928 | +0.000688 | 1.0190 |
| src/compiler/watchPublic.ts:396:1 | ParsedConfig | 0.035199 | 0.035880 | +0.000681 | 1.0193 |
| src/compiler/factory/utilities.ts:1597:1 | formatGeneratedName | 0.039054 | 0.039735 | +0.000681 | 1.0174 |
| src/compiler/factory/emitNode.ts:317:1 | getSnippetElement | 0.039301 | 0.039974 | +0.000674 | 1.0171 |
| src/compiler/factory/nodeFactory.ts:7536:1 | mergeTokenSourceMapRanges | 0.036156 | 0.036825 | +0.000669 | 1.0185 |
| src/compiler/utilities.ts:2111:1 | isShorthandAmbientModule | 0.118706 | 0.119374 | +0.000667 | 1.0056 |
| src/compiler/types.ts:3301:1 | JsxExpression | 0.036049 | 0.036709 | +0.000660 | 1.0183 |
| src/compiler/factory/nodeTests.ts:736:1 | isDoStatement | 0.042496 | 0.043156 | +0.000660 | 1.0155 |
| src/compiler/transformers/destructuring.ts:99:1 | flattenDestructuringAssignment | 0.095836 | 0.096494 | +0.000658 | 1.0069 |
| src/compiler/types.ts:749:1 | JsxTokenSyntaxKind | 0.035289 | 0.035944 | +0.000655 | 1.0186 |
| src/compiler/program.ts:707:1 | formatColorAndReset | 0.036436 | 0.037091 | +0.000654 | 1.0180 |
| src/compiler/types.ts:8127:1 | ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 0.036837 | 0.037491 | +0.000654 | 1.0178 |
| src/compiler/utilities.ts:8420:1 | getLeftmostExpression | 0.155310 | 0.155963 | +0.000653 | 1.0042 |
| src/compiler/transformers/legacyDecorators.ts:88:1 | transformLegacyDecorators | 0.080695 | 0.081347 | +0.000652 | 1.0081 |
| src/compiler/utilitiesPublic.ts:1894:1 | isObjectBindingOrAssignmentPattern | 0.036142 | 0.036793 | +0.000651 | 1.0180 |
| src/compiler/factory/nodeTests.ts:1:1 |  | 0.039980 | 0.040629 | +0.000649 | 1.0162 |
| src/compiler/factory/nodeTests.ts:1112:1 | isJSDocPublicTag | 0.040678 | 0.041325 | +0.000646 | 1.0159 |
| src/compiler/utilities.ts:4482:1 | isJSDocTypeAlias | 0.042622 | 0.043267 | +0.000645 | 1.0151 |
| src/compiler/types.ts:3282:1 | JsxAttributeValue | 0.036072 | 0.036713 | +0.000641 | 1.0178 |
| src/compiler/sys.ts:1:1 |  | 0.036256 | 0.036884 | +0.000628 | 1.0173 |
| src/compiler/checker.ts:1408:1 | AddUnusedDiagnostic | 0.037273 | 0.037898 | +0.000625 | 1.0168 |
| src/compiler/transformers/utilities.ts:587:1 | getProperties | 0.035823 | 0.036447 | +0.000624 | 1.0174 |
| src/compiler/types.ts:5729:1 | AnyImportSyntax | 0.034572 | 0.035189 | +0.000618 | 1.0179 |
| src/compiler/utilitiesPublic.ts:1647:1 | isModifier | 0.084458 | 0.085075 | +0.000618 | 1.0073 |
| src/compiler/utilities.ts:10807:1 | containsIgnoredPath | 0.041261 | 0.041877 | +0.000615 | 1.0149 |
| src/compiler/moduleNameResolver.ts:726:1 | getNodeResolutionFeatures | 0.037389 | 0.038003 | +0.000614 | 1.0164 |
| src/compiler/utilities.ts:10014:1 | getSupportedExtensionsWithJsonIfResolveJsonModule | 0.037522 | 0.038133 | +0.000611 | 1.0163 |
| src/compiler/factory/nodeTests.ts:862:1 | isImportAttributes | 0.043360 | 0.043968 | +0.000609 | 1.0140 |
| src/compiler/types.ts:4214:1 | FlowAssignment | 0.036297 | 0.036901 | +0.000604 | 1.0166 |
| src/compiler/builder.ts:1097:1 | IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo | 0.036380 | 0.036968 | +0.000589 | 1.0162 |
| src/compiler/debug.ts:99:1 | LogLevel | 0.036779 | 0.037367 | +0.000588 | 1.0160 |
| src/compiler/moduleNameResolver.ts:1113:1 | ModeAwareCacheKey | 0.037436 | 0.038024 | +0.000588 | 1.0157 |
| src/compiler/types.ts:5755:1 | BindingElementOfBareOrAccessedRequire | 0.034664 | 0.035250 | +0.000586 | 1.0169 |
| src/compiler/watchUtilities.ts:68:1 | DirectoryStructureHost | 0.035021 | 0.035606 | +0.000585 | 1.0167 |
| src/compiler/performance.ts:202:1 | disable | 0.036370 | 0.036953 | +0.000583 | 1.0160 |
| src/compiler/utilities.ts:10127:1 | getRequiresAtTopOfFile | 0.038156 | 0.038739 | +0.000582 | 1.0153 |
| src/compiler/sys.ts:907:1 | createFileWatcherCallback | 0.037615 | 0.038195 | +0.000580 | 1.0154 |
| src/compiler/utilities.ts:10018:1 | getSupportedExtensionsWithJsonIfResolveJsonModule | 0.035898 | 0.036475 | +0.000577 | 1.0161 |
| src/compiler/semver.ts:401:1 | isWildcard | 0.037530 | 0.038106 | +0.000576 | 1.0154 |
| src/compiler/types.ts:8674:1 | TypeOfTag | 0.039133 | 0.039706 | +0.000574 | 1.0147 |
| src/compiler/types.ts:2478:1 | SuperExpression | 0.036214 | 0.036778 | +0.000564 | 1.0156 |
| src/compiler/types.ts:4656:1 | FileIncludeReason | 0.035952 | 0.036509 | +0.000557 | 1.0155 |
| src/compiler/types.ts:1602:1 | PunctuationToken | 0.037376 | 0.037923 | +0.000547 | 1.0146 |
| src/compiler/utilitiesPublic.ts:1872:1 | isBindingOrAssignmentElement | 0.274954 | 0.275499 | +0.000544 | 1.0020 |
| src/compiler/utilities.ts:8369:1 | isNamedImportsOrExports | 0.042427 | 0.042966 | +0.000539 | 1.0127 |
| src/compiler/types.ts:3735:1 | AssertEntry | 0.034813 | 0.035351 | +0.000538 | 1.0154 |
| src/compiler/tracing.ts:369:1 |  | 0.035642 | 0.036179 | +0.000537 | 1.0151 |
| src/compiler/utilitiesPublic.ts:687:1 |  | 0.036997 | 0.037533 | +0.000536 | 1.0145 |
| src/compiler/executeCommandLine.ts:968:1 | updateSolutionBuilderHost | 0.072596 | 0.073131 | +0.000535 | 1.0074 |
| src/compiler/utilitiesPublic.ts:2601:1 |  | 0.035800 | 0.036331 | +0.000531 | 1.0148 |
| src/compiler/builder.ts:95:1 | ReusableDiagnostic | 0.037672 | 0.038202 | +0.000530 | 1.0141 |
| src/compiler/utilitiesPublic.ts:931:1 | nodeHasName | 0.042512 | 0.043039 | +0.000527 | 1.0124 |
| src/compiler/visitorPublic.ts:148:1 | visitNode | 0.036458 | 0.036982 | +0.000524 | 1.0144 |
| src/compiler/factory/nodeTests.ts:334:1 | isDefaultModifier | 0.042504 | 0.043028 | +0.000524 | 1.0123 |
| src/compiler/commandLineParser.ts:3053:1 |  | 0.038493 | 0.039014 | +0.000521 | 1.0135 |
| src/compiler/program.ts:680:1 | ForegroundColorEscapeSequences | 0.037449 | 0.037969 | +0.000520 | 1.0139 |
| src/compiler/utilitiesPublic.ts:296:1 | isExternalModuleNameRelative | 0.036034 | 0.036554 | +0.000520 | 1.0144 |
| src/compiler/types.ts:2626:1 | AssignmentOperator | 0.037913 | 0.038432 | +0.000519 | 1.0137 |
| src/compiler/transformers/utilities.ts:571:1 | findSuperStatementIndexPath | 0.035810 | 0.036326 | +0.000516 | 1.0144 |
| src/compiler/factory/nodeTests.ts:387:1 | isCaseKeyword | 0.040214 | 0.040724 | +0.000510 | 1.0127 |
| src/compiler/types.ts:3694:1 | ImportDeclaration | 0.035458 | 0.035961 | +0.000503 | 1.0142 |
| src/compiler/utilities.ts:4929:1 | isCompoundLikeAssignment | 0.038546 | 0.039048 | +0.000502 | 1.0130 |
| src/compiler/factory/nodeTests.ts:288:1 | isMinusToken | 0.041494 | 0.041995 | +0.000501 | 1.0121 |
| src/compiler/factory/nodeTests.ts:977:1 | isCaseClause | 0.041824 | 0.042325 | +0.000501 | 1.0120 |
| src/compiler/semver.ts:276:1 | parseRange | 0.037542 | 0.038040 | +0.000498 | 1.0133 |
| src/compiler/utilitiesPublic.ts:1175:1 | getJSDocDeprecatedTagNoCache | 0.049741 | 0.050237 | +0.000496 | 1.0100 |
| src/compiler/types.ts:2482:1 | ImportExpression | 0.036535 | 0.037030 | +0.000496 | 1.0136 |
| src/compiler/_namespaces/ts.ts:76:1 |  | 0.036531 | 0.037025 | +0.000494 | 1.0135 |
| src/compiler/types.ts:2083:1 | FunctionLike | 0.038264 | 0.038757 | +0.000493 | 1.0129 |
| src/compiler/builderPublic.ts:41:1 | HostForComputeHash | 0.036123 | 0.036614 | +0.000492 | 1.0136 |
| src/compiler/utilities.ts:2983:1 | getMembersOfDeclaration | 0.036800 | 0.037291 | +0.000491 | 1.0133 |
| src/compiler/types.ts:10424:1 | PragmaPseudoMapEntry | 0.036502 | 0.036984 | +0.000481 | 1.0132 |
| src/compiler/types.ts:2641:1 | LogicalOrCoalescingAssignmentOperator | 0.037553 | 0.038032 | +0.000479 | 1.0128 |
| src/compiler/watchPublic.ts:291:1 | WatchCompilerHostOfFilesAndCompilerOptions | 0.035860 | 0.036335 | +0.000475 | 1.0133 |
| src/compiler/utilities.ts:7211:1 | hasOverrideModifier | 0.041670 | 0.042139 | +0.000469 | 1.0113 |
| src/compiler/utilities.ts:5236:1 | getInterfaceBaseTypeNodes | 0.036606 | 0.037074 | +0.000468 | 1.0128 |
| src/compiler/types.ts:3436:1 | ForOfStatement | 0.035976 | 0.036444 | +0.000468 | 1.0130 |
| src/compiler/utilities.ts:12423:1 | addEmitFlagsRecursively | 0.056310 | 0.056768 | +0.000458 | 1.0081 |
| src/compiler/types.ts:3860:1 | FileReference | 0.036360 | 0.036818 | +0.000457 | 1.0126 |
| src/compiler/path.ts:1:1 |  | 0.038440 | 0.038895 | +0.000455 | 1.0118 |
| src/compiler/types.ts:1767:1 | DynamicNamedDeclaration | 0.035424 | 0.035875 | +0.000451 | 1.0127 |
| src/compiler/moduleNameResolver.ts:930:1 | isMissingPackageJsonInfo | 0.036653 | 0.037103 | +0.000449 | 1.0123 |
| src/compiler/types.ts:7633:1 | JsxEmit | 0.035918 | 0.036366 | +0.000448 | 1.0125 |
| src/compiler/builder.ts:1576:1 | BuilderCreationParameters | 0.036085 | 0.036532 | +0.000448 | 1.0124 |
| src/compiler/types.ts:3390:1 | PrologueDirective | 0.036363 | 0.036810 | +0.000447 | 1.0123 |
| src/compiler/transformers/esDecorators.ts:1:1 |  | 0.035324 | 0.035770 | +0.000446 | 1.0126 |
| src/compiler/_namespaces/ts.ts:65:1 |  | 0.035844 | 0.036289 | +0.000445 | 1.0124 |
| src/compiler/sys.ts:902:1 | FileSystemEntryKind | 0.036635 | 0.037074 | +0.000439 | 1.0120 |
| src/compiler/factory/emitHelpers.ts:958:1 |  | 0.062277 | 0.062716 | +0.000439 | 1.0070 |
| src/compiler/commandLineParser.ts:1928:1 | getOptionName | 0.036683 | 0.037120 | +0.000438 | 1.0119 |
| src/compiler/tsbuild.ts:177:1 | resolveConfigFileProjectName | 0.036718 | 0.037148 | +0.000430 | 1.0117 |
| src/compiler/factory/nodeTests.ts:1188:1 | isJSDocThrowsTag | 0.040449 | 0.040872 | +0.000423 | 1.0105 |
| src/compiler/utilitiesPublic.ts:1195:1 | getJSDocTemplateTag | 0.052720 | 0.053143 | +0.000423 | 1.0080 |
| src/compiler/commandLineParser.ts:1932:1 | createUnknownOptionError | 0.043215 | 0.043636 | +0.000421 | 1.0097 |
| src/compiler/tsbuildPublic.ts:849:1 | UpdateOutputFileStampsProject | 0.036014 | 0.036435 | +0.000421 | 1.0117 |
| src/compiler/transformers/namedEvaluation.ts:256:1 | transformNamedEvaluationOfPropertyAssignment | 0.039475 | 0.039896 | +0.000421 | 1.0107 |
| src/compiler/factory/nodeTests.ts:927:1 | isJsxElement | 0.041020 | 0.041439 | +0.000419 | 1.0102 |
| src/compiler/utilities.ts:6590:1 | getPossibleOriginalInputPathWithoutChangingExt | 0.037652 | 0.038069 | +0.000417 | 1.0111 |
| src/compiler/factory/utilities.ts:228:1 | createJsxFactoryExpression | 0.080052 | 0.080465 | +0.000414 | 1.0052 |
| src/compiler/utilities.ts:8607:1 | getLocaleSpecificMessage | 0.038040 | 0.038452 | +0.000412 | 1.0108 |
| src/compiler/moduleNameResolver.ts:2399:1 | getPackageScopeForPath | 0.036442 | 0.036851 | +0.000410 | 1.0112 |
| src/compiler/transformers/destructuring.ts:1:1 |  | 0.035553 | 0.035962 | +0.000409 | 1.0115 |
| src/compiler/tsbuildPublic.ts:1371:1 | isFileWatcherWithModifiedTime | 0.037604 | 0.038011 | +0.000407 | 1.0108 |
| src/compiler/factory/emitNode.ts:368:1 | getIdentifierAutoGenerate | 0.035665 | 0.036071 | +0.000406 | 1.0114 |
| src/compiler/utilities.ts:8045:1 | getDeclarationModifierFlagsFromSymbol | 0.037774 | 0.038176 | +0.000403 | 1.0107 |
| src/compiler/utilities.ts:3413:1 | isThisProperty | 0.040972 | 0.041372 | +0.000400 | 1.0098 |
| src/compiler/factory/nodeTests.ts:740:1 | isWhileStatement | 0.044624 | 0.045022 | +0.000398 | 1.0089 |
| src/compiler/utilities.ts:5644:1 | getOperatorAssociativity | 0.040204 | 0.040601 | +0.000397 | 1.0099 |
| src/compiler/path.ts:526:1 | normalizeSlashes | 0.037750 | 0.038143 | +0.000393 | 1.0104 |
| src/compiler/factory/nodeTests.ts:632:1 | isAwaitExpression | 0.042527 | 0.042913 | +0.000387 | 1.0091 |
| src/compiler/watchUtilities.ts:1:1 |  | 0.035757 | 0.036142 | +0.000385 | 1.0108 |
| src/compiler/transformers/typeSerializer.ts:132:1 | createRuntimeTypeSerializer | 0.074156 | 0.074538 | +0.000382 | 1.0052 |
| src/compiler/factory/emitNode.ts:282:1 | getEmitHelpers | 0.039821 | 0.040197 | +0.000376 | 1.0094 |
| src/compiler/builderPublic.ts:215:1 | createEmitAndSemanticDiagnosticsBuilderProgram | 0.038263 | 0.038638 | +0.000375 | 1.0098 |
| src/compiler/watchPublic.ts:277:1 | WatchCompilerHost | 0.037126 | 0.037496 | +0.000369 | 1.0099 |
| src/compiler/utilitiesPublic.ts:393:1 | textSpanIntersectsWith | 0.034858 | 0.035223 | +0.000365 | 1.0105 |
| src/compiler/utilitiesPublic.ts:2090:1 | isExpressionKind | 0.040203 | 0.040566 | +0.000362 | 1.0090 |
| src/compiler/utilities.ts:7190:1 | hasEffectiveModifier | 0.043361 | 0.043720 | +0.000359 | 1.0083 |
| src/compiler/utilities.ts:12033:1 | isPrimitiveLiteralValue | 0.152569 | 0.152926 | +0.000357 | 1.0023 |
| src/compiler/utilities.ts:7281:1 | getEffectiveModifierFlagsAlwaysIncludeJSDoc | 0.043916 | 0.044273 | +0.000357 | 1.0081 |
| src/compiler/executeCommandLine.ts:1138:1 | canReportDiagnostics | 0.034389 | 0.034745 | +0.000356 | 1.0104 |
| src/compiler/types.ts:3964:1 | JSDocLink | 0.035557 | 0.035910 | +0.000354 | 1.0099 |
| src/compiler/program.ts:4881:1 | HostForUseSourceOfProjectReferenceRedirect | 0.036337 | 0.036681 | +0.000344 | 1.0095 |
| src/compiler/types.ts:10030:1 | ResolvedModuleSpecifierInfo | 0.036139 | 0.036482 | +0.000343 | 1.0095 |
| src/compiler/commandLineParser.ts:2119:1 |  | 0.063320 | 0.063657 | +0.000337 | 1.0053 |
| src/compiler/path.ts:805:1 | ensureTrailingDirectorySeparator | 0.034455 | 0.034792 | +0.000337 | 1.0098 |
| src/compiler/emitter.ts:590:1 | getOutputJSFileNameWorker | 0.036210 | 0.036546 | +0.000336 | 1.0093 |
| src/compiler/_namespaces/ts.ts:28:1 |  | 0.035973 | 0.036308 | +0.000335 | 1.0093 |
| src/compiler/watchPublic.ts:99:1 | ReadBuildProgramHost | 0.036833 | 0.037165 | +0.000332 | 1.0090 |
| src/compiler/scanner.ts:987:1 | isIdentifierText | 0.036516 | 0.036847 | +0.000331 | 1.0091 |
| src/compiler/corePublic.ts:17:1 | SortedReadonlyArray | 0.035433 | 0.035764 | +0.000331 | 1.0094 |
| src/compiler/types.ts:1729:1 | QualifiedName | 0.037440 | 0.037770 | +0.000330 | 1.0088 |
| src/compiler/types.ts:7585:1 | WatchOptions | 0.035957 | 0.036286 | +0.000329 | 1.0092 |
| src/compiler/watchPublic.ts:330:1 | WatchCompilerHostOfConfigFile | 0.037423 | 0.037750 | +0.000328 | 1.0088 |
| src/compiler/types.ts:2726:1 | ArrayBindingOrAssignmentPattern | 0.036342 | 0.036663 | +0.000322 | 1.0088 |
| src/compiler/_namespaces/ts.ts:63:1 |  | 0.034885 | 0.035205 | +0.000320 | 1.0092 |
| src/compiler/types.ts:3970:1 | JSDocLinkCode | 0.035358 | 0.035675 | +0.000317 | 1.0090 |
| src/compiler/factory/emitHelpers.ts:1170:1 |  | 0.059948 | 0.060264 | +0.000317 | 1.0053 |
| src/compiler/commandLineParser.ts:4179:1 | getWildcardDirectoryFromSpec | 0.039536 | 0.039850 | +0.000314 | 1.0080 |
| src/compiler/types.ts:1872:1 | BindingName | 0.036990 | 0.037303 | +0.000313 | 1.0085 |
| src/compiler/diagnosticInformationMap.generated.ts:6:1 | diag | 0.035372 | 0.035684 | +0.000312 | 1.0088 |
| src/compiler/utilities.ts:1286:1 | getSourceTextOfNodeFromSourceFile | 0.043427 | 0.043737 | +0.000310 | 1.0071 |
| src/compiler/moduleNameResolver.ts:2445:1 | getPackageJsonInfo | 0.116510 | 0.116817 | +0.000306 | 1.0026 |
| src/compiler/types.ts:2520:1 | ExponentiationOperator | 0.037630 | 0.037932 | +0.000303 | 1.0080 |
| src/compiler/types.ts:4649:1 | AutomaticTypeDirectiveFile | 0.035651 | 0.035953 | +0.000302 | 1.0085 |
| src/compiler/factory/nodeTests.ts:1180:1 | isJSDocImplementsTag | 0.041652 | 0.041954 | +0.000302 | 1.0072 |
| src/compiler/sys.ts:130:1 |  | 0.036663 | 0.036963 | +0.000300 | 1.0082 |
| src/compiler/utilities.ts:7824:1 |  | 0.035317 | 0.035615 | +0.000298 | 1.0084 |
| src/compiler/commandLineParser.ts:2127:1 | parseCommandLine | 0.036773 | 0.037070 | +0.000297 | 1.0081 |
| src/compiler/types.ts:4973:1 | CustomTransformerFactory | 0.034777 | 0.035074 | +0.000297 | 1.0085 |
| src/compiler/types.ts:2583:1 | BitwiseOperator | 0.037366 | 0.037661 | +0.000295 | 1.0079 |
| src/compiler/transformers/classThis.ts:52:1 | ClassThisAssignmentBlock | 0.037562 | 0.037857 | +0.000295 | 1.0078 |
| src/compiler/utilities.ts:8077:1 | getCombinedLocalAndExportSymbolFlags | 0.037062 | 0.037357 | +0.000294 | 1.0079 |
| src/compiler/utilities.ts:2643:1 | isJsonSourceFile | 0.036970 | 0.037261 | +0.000291 | 1.0079 |
| src/compiler/commandLineParser.ts:3265:1 | handleOptionConfigDirTemplateSubstitution | 4.480565 | 4.480855 | +0.000290 | 1.0001 |
| src/compiler/utilities.ts:1377:1 | ScriptTargetFeatures | 0.037458 | 0.037741 | +0.000284 | 1.0076 |
| src/compiler/types.ts:10108:1 | SyntaxList | 0.035988 | 0.036269 | +0.000281 | 1.0078 |
| src/compiler/executeCommandLine.ts:169:1 | shouldBePretty | 0.036339 | 0.036619 | +0.000280 | 1.0077 |
| src/compiler/moduleNameResolver.ts:2384:1 | PackageJsonInfoContents | 0.036289 | 0.036569 | +0.000280 | 1.0077 |
| src/compiler/transformers/generators.ts:265:1 | ExceptionBlock | 0.038054 | 0.038330 | +0.000276 | 1.0072 |
| src/compiler/types.ts:1629:1 | ModifierToken | 0.035402 | 0.035674 | +0.000272 | 1.0077 |
| src/compiler/visitorPublic.ts:1805:1 | extractSingleNode | 0.050548 | 0.050818 | +0.000270 | 1.0053 |
| src/compiler/utilities.ts:4113:1 | getAssignmentDeclarationKind | 0.046015 | 0.046281 | +0.000266 | 1.0058 |
| src/compiler/utilitiesPublic.ts:2534:1 | isJSDocTag | 0.040698 | 0.040962 | +0.000264 | 1.0065 |
| src/compiler/builder.ts:1235:1 | getBuildInfo | 0.038571 | 0.038833 | +0.000263 | 1.0068 |
| src/compiler/builderPublic.ts:20:1 | AffectedFileResult | 0.041153 | 0.041415 | +0.000262 | 1.0064 |
| src/compiler/types.ts:9970:1 | EmitTextWriter | 0.035778 | 0.036031 | +0.000253 | 1.0071 |
| src/compiler/utilities.ts:1305:1 | moduleExportNameTextEscaped | 0.036693 | 0.036945 | +0.000252 | 1.0069 |
| src/compiler/factory/utilities.ts:1171:1 | isExponentiationOperator | 0.039728 | 0.039974 | +0.000246 | 1.0062 |
| src/compiler/utilities.ts:10016:1 | getSupportedExtensionsWithJsonIfResolveJsonModule | 0.037436 | 0.037681 | +0.000245 | 1.0065 |
| src/compiler/utilities.ts:10025:1 | isJSLike | 0.035828 | 0.036065 | +0.000237 | 1.0066 |
| src/compiler/utilities.ts:1295:1 | isExportNamespaceAsDefaultDeclaration | 0.290084 | 0.290317 | +0.000233 | 1.0008 |
| src/compiler/types.ts:3672:1 | ImportEqualsDeclaration | 0.035546 | 0.035777 | +0.000231 | 1.0065 |
| src/compiler/scanner.ts:42:1 | tokenIsIdentifierOrKeyword | 0.044682 | 0.044912 | +0.000230 | 1.0051 |
| src/compiler/utilities.ts:10315:1 |  | 0.062359 | 0.062589 | +0.000230 | 1.0037 |
| src/compiler/utilities.ts:10040:1 | hasImplementationTSFileExtension | 0.039817 | 0.040043 | +0.000226 | 1.0057 |
| src/compiler/factory/nodeTests.ts:636:1 | isPrefixUnaryExpression | 0.041718 | 0.041944 | +0.000226 | 1.0054 |
| src/compiler/utilities.ts:7206:1 | hasStaticModifier | 0.042771 | 0.042997 | +0.000225 | 1.0053 |
| src/compiler/types.ts:2496:1 | VoidExpression | 0.036400 | 0.036624 | +0.000223 | 1.0061 |
| src/compiler/_namespaces/ts.ts:29:1 |  | 0.036148 | 0.036366 | +0.000218 | 1.0060 |
| src/compiler/visitorPublic.ts:508:1 | visitFunctionBody | 0.035499 | 0.035714 | +0.000215 | 1.0060 |
| src/compiler/utilities.ts:5309:1 | FunctionFlags | 0.037512 | 0.037723 | +0.000211 | 1.0056 |
| src/compiler/types.ts:3910:1 | JSDocUnknownType | 0.035652 | 0.035862 | +0.000210 | 1.0059 |
| src/compiler/types.ts:4170:1 | FlowNode | 0.048341 | 0.048550 | +0.000209 | 1.0043 |
| src/compiler/factory/nodeTests.ts:1070:1 | isJSDocFunctionType | 0.043183 | 0.043391 | +0.000208 | 1.0048 |
| src/compiler/transformers/declarations.ts:1890:1 | CanHaveLiteralInitializer | 0.035286 | 0.035494 | +0.000208 | 1.0059 |
| src/compiler/moduleNameResolver.ts:894:1 | PerNonRelativeNameCache | 0.038670 | 0.038874 | +0.000204 | 1.0053 |
| src/compiler/checker.ts:1154:1 |  | 0.039393 | 0.039595 | +0.000203 | 1.0051 |
| src/compiler/types.ts:9701:1 | TransformerFactory | 0.038130 | 0.038324 | +0.000194 | 1.0051 |
| src/compiler/factory/utilities.ts:1489:1 | isExportOrDefaultKeywordKind | 0.038833 | 0.039026 | +0.000192 | 1.0050 |
| src/compiler/commandLineParser.ts:2483:1 | convertToJson | 0.044092 | 0.044283 | +0.000192 | 1.0043 |
| src/compiler/utilities.ts:7626:1 | tryExtractTSExtension | 0.037670 | 0.037860 | +0.000189 | 1.0050 |
| src/compiler/utilities.ts:8593:1 | setLocalizedDiagnosticMessages | 0.036938 | 0.037126 | +0.000189 | 1.0051 |
| src/compiler/emitter.ts:6370:1 | emitListItemWithParenthesizerRule | 0.040396 | 0.040583 | +0.000187 | 1.0046 |
| src/compiler/program.ts:690:1 |  | 0.037553 | 0.037728 | +0.000176 | 1.0047 |
| src/compiler/utilities.ts:8147:1 | compareDataObjects | 0.036189 | 0.036361 | +0.000173 | 1.0048 |
| src/compiler/checker.ts:54223:1 | JsxNames | 0.038923 | 0.039093 | +0.000170 | 1.0044 |
| src/compiler/utilitiesPublic.ts:2522:1 | isJSDocCommentContainingNode | 0.266033 | 0.266203 | +0.000170 | 1.0006 |
| src/compiler/types.ts:3295:1 | JsxClosingElement | 0.036576 | 0.036742 | +0.000166 | 1.0045 |
| src/compiler/factory/nodeFactory.ts:466:1 | NodeFactoryFlags | 0.035145 | 0.035310 | +0.000165 | 1.0047 |
| src/compiler/program.ts:932:1 | getModeForUsageLocation | 0.044494 | 0.044659 | +0.000165 | 1.0037 |
| src/compiler/commandLineParser.ts:2206:1 | getDiagnosticText | 0.039455 | 0.039619 | +0.000164 | 1.0041 |
| src/compiler/utilities.ts:7316:1 | selectEffectiveModifierFlags | 0.038316 | 0.038477 | +0.000162 | 1.0042 |
| src/compiler/utilities.ts:5286:1 | isNonContextualKeyword | 0.072718 | 0.072872 | +0.000154 | 1.0021 |
| src/compiler/utilities.ts:11196:1 |  | 0.038567 | 0.038717 | +0.000150 | 1.0039 |
| src/compiler/utilitiesPublic.ts:1190:1 | getJSDocReturnTag | 0.050574 | 0.050718 | +0.000144 | 1.0029 |
| src/compiler/utilities.ts:612:1 |  | 0.037471 | 0.037612 | +0.000141 | 1.0038 |
| src/compiler/emitter.ts:6362:1 | emitListItemNoParenthesizer | 0.040213 | 0.040353 | +0.000140 | 1.0035 |
| src/compiler/utilities.ts:6060:1 | createDiagnosticCollection | 0.036085 | 0.036219 | +0.000134 | 1.0037 |
| src/compiler/tsbuildPublic.ts:836:1 | InvalidatedProjectBase | 0.036340 | 0.036474 | +0.000134 | 1.0037 |
| src/compiler/factory/nodeTests.ts:915:1 | isSyntheticReference | 0.040654 | 0.040786 | +0.000132 | 1.0033 |
| src/compiler/types.ts:4642:1 | ReferencedFile | 0.041006 | 0.041138 | +0.000132 | 1.0032 |
| src/compiler/factory/emitNode.ts:357:1 | getIdentifierTypeArguments | 0.035835 | 0.035962 | +0.000127 | 1.0035 |
| src/compiler/utilities.ts:6138:1 |  | 0.036175 | 0.036298 | +0.000124 | 1.0034 |
| src/compiler/types.ts:2491:1 | TypeOfExpression | 0.035981 | 0.036102 | +0.000121 | 1.0034 |
| src/compiler/utilitiesPublic.ts:412:1 | textSpanIntersection | 0.036415 | 0.036535 | +0.000120 | 1.0033 |
| src/compiler/sys.ts:1393:1 | BufferEncoding | 0.036869 | 0.036987 | +0.000118 | 1.0032 |
| src/compiler/commandLineParser.ts:2150:1 | ParsedBuildCommand | 0.037043 | 0.037160 | +0.000117 | 1.0032 |
| src/compiler/factory/emitNode.ts:379:1 | getIdentifierGeneratedImportReference | 0.036661 | 0.036774 | +0.000113 | 1.0031 |
| src/compiler/moduleSpecifiers.ts:1429:1 | tryGetJSExtensionForFile | 0.038827 | 0.038936 | +0.000110 | 1.0028 |
| src/compiler/types.ts:3840:1 | TypeOnlyExportDeclaration | 0.035203 | 0.035311 | +0.000108 | 1.0031 |
| src/compiler/emitter.ts:580:1 | getOutputJSFileName | 0.037086 | 0.037191 | +0.000105 | 1.0028 |
| src/compiler/visitorPublic.ts:521:18 | visitFunctionBody | 0.035559 | 0.035654 | +0.000096 | 1.0027 |
| src/compiler/factory/nodeTests.ts:458:1 | isIndexSignatureDeclaration | 0.042041 | 0.042135 | +0.000094 | 1.0022 |
| src/compiler/types.ts:2270:1 | ArrayTypeNode | 0.035545 | 0.035637 | +0.000091 | 1.0026 |
| src/compiler/scanner.ts:740:1 | isConflictMarkerTrivia | 0.036350 | 0.036440 | +0.000090 | 1.0025 |
| src/compiler/types.ts:2954:1 | ObjectLiteralExpressionBase | 0.035874 | 0.035963 | +0.000089 | 1.0025 |
| src/compiler/factory/nodeFactory.ts:7290:1 | propagateIdentifierNameFlags | 0.040184 | 0.040270 | +0.000086 | 1.0021 |
| src/compiler/utilitiesPublic.ts:1:1 |  | 0.035754 | 0.035839 | +0.000085 | 1.0024 |
| src/compiler/performance.ts:17:1 |  | 0.036081 | 0.036162 | +0.000080 | 1.0022 |
| src/compiler/watchUtilities.ts:481:1 | updateMissingFilePathsWatch | 0.037280 | 0.037359 | +0.000079 | 1.0021 |
| src/compiler/path.ts:803:1 | ensureTrailingDirectorySeparator | 0.034532 | 0.034608 | +0.000076 | 1.0022 |
| src/compiler/moduleNameResolver.ts:852:1 | TypeReferenceDirectiveResolutionCache | 0.041452 | 0.041526 | +0.000074 | 1.0018 |
| src/compiler/utilities.ts:622:1 |  | 0.035876 | 0.035946 | +0.000070 | 1.0019 |
| src/compiler/watch.ts:166:1 | getPlainDiagnosticFollowingNewLines | 0.036476 | 0.036545 | +0.000070 | 1.0019 |
| src/compiler/transformers/destructuring.ts:81:1 | FlattenLevel | 0.035322 | 0.035390 | +0.000068 | 1.0019 |
| src/compiler/executeCommandLine.ts:176:1 | getOptionsForHelp | 0.036124 | 0.036191 | +0.000068 | 1.0019 |
| src/compiler/utilities.ts:9304:1 |  | 0.073604 | 0.073669 | +0.000066 | 1.0009 |
| src/compiler/transformers/typeSerializer.ts:86:1 | RuntimeTypeSerializerContext | 0.036798 | 0.036858 | +0.000060 | 1.0016 |
| src/compiler/path.ts:1010:1 | getPathComponentsRelativeTo | 0.035927 | 0.035986 | +0.000059 | 1.0016 |
| src/compiler/types.ts:7682:1 | ScriptTarget | 0.035932 | 0.035989 | +0.000057 | 1.0016 |
| src/compiler/types.ts:5735:1 | AnyImportOrRequire | 0.034650 | 0.034705 | +0.000055 | 1.0016 |
| src/compiler/utilities.ts:5045:1 | skipParentheses | 0.035954 | 0.036005 | +0.000051 | 1.0014 |
| src/compiler/executeCommandLine.ts:1142:1 | canTrace | 0.034577 | 0.034627 | +0.000050 | 1.0015 |
| src/compiler/program.ts:1049:1 | createModuleResolutionLoader | 0.036742 | 0.036790 | +0.000048 | 1.0013 |
| src/compiler/utilitiesPublic.ts:1612:1 | isModifierKind | 0.038888 | 0.038935 | +0.000047 | 1.0012 |
| src/compiler/scanner.ts:974:1 | isIdentifierStart | 0.036692 | 0.036739 | +0.000047 | 1.0013 |
| src/compiler/program.ts:688:1 |  | 0.037175 | 0.037220 | +0.000045 | 1.0012 |
| src/compiler/types.ts:2608:1 | CompoundAssignmentOperator | 0.037430 | 0.037472 | +0.000042 | 1.0011 |
| src/compiler/types.ts:3877:1 | SynthesizedComment | 0.034097 | 0.034139 | +0.000042 | 1.0012 |
| src/compiler/types.ts:1635:1 | ConstKeyword | 0.034370 | 0.034412 | +0.000042 | 1.0012 |
| src/compiler/factory/utilities.ts:437:1 | createExpressionForMethodDeclaration | 0.082067 | 0.082107 | +0.000040 | 1.0005 |
| src/compiler/utilities.ts:4158:1 | isBindableStaticElementAccessExpression | 0.041968 | 0.042007 | +0.000038 | 1.0009 |
| src/compiler/moduleSpecifiers.ts:1:1 |  | 0.036600 | 0.036636 | +0.000036 | 1.0010 |
| src/compiler/utilities.ts:2107:1 | isShorthandAmbientModuleSymbol | 0.041890 | 0.041925 | +0.000035 | 1.0008 |
| src/compiler/factory/nodeTests.ts:308:1 | isQuestionDotToken | 0.042744 | 0.042776 | +0.000032 | 1.0008 |
| src/compiler/tsbuildPublic.ts:884:1 | InvalidatedProject | 0.036113 | 0.036136 | +0.000023 | 1.0006 |
| src/compiler/utilities.ts:7216:1 | hasAbstractModifier | 0.042344 | 0.042360 | +0.000016 | 1.0004 |
| src/compiler/types.ts:5584:1 | TypeFormatFlags | 0.034466 | 0.034482 | +0.000016 | 1.0005 |
| src/compiler/sourcemap.ts:422:1 | tryParseRawSourceMap | 0.037012 | 0.037027 | +0.000015 | 1.0004 |
| src/compiler/emitter.ts:428:1 |  | 0.047759 | 0.047774 | +0.000015 | 1.0003 |
| src/compiler/moduleNameResolver.ts:173:1 | PathAndExtension | 0.035121 | 0.035136 | +0.000015 | 1.0004 |
| src/compiler/utilities.ts:6140:1 | escapeTemplateSubstitution | 0.035996 | 0.036010 | +0.000014 | 1.0004 |
| src/compiler/utilities.ts:8231:1 | MutateMapOptionsCreate | 0.035450 | 0.035463 | +0.000013 | 1.0004 |
| src/compiler/commandLineParser.ts:151:1 |  | 0.041007 | 0.041019 | +0.000012 | 1.0003 |
| src/compiler/types.ts:2240:1 | NodeWithTypeArguments | 0.036391 | 0.036401 | +0.000010 | 1.0003 |
| src/compiler/expressionToTypeNode.ts:169:1 |  | 0.041972 | 0.041979 | +0.000007 | 1.0002 |
| src/compiler/transformers/namedEvaluation.ts:156:1 | classHasExplicitlyAssignedName | 0.045002 | 0.045007 | +0.000005 | 1.0001 |
| src/compiler/types.ts:7604:1 | ModuleKind | 0.035358 | 0.035360 | +0.000003 | 1.0001 |
| src/compiler/utilities.ts:1300:1 | moduleExportNameTextUnescaped | 0.037534 | 0.037535 | +0.000002 | 1.0000 |
| src/compiler/types.ts:2808:1 | NoSubstitutionTemplateLiteral | 0.037637 | 0.037636 | -0.000002 | 1.0000 |
| src/compiler/types.ts:3563:1 | ClassLikeDeclaration | 0.035875 | 0.035871 | -0.000004 | 0.9999 |
| src/compiler/semver.ts:418:1 | testAlternative | 0.036597 | 0.036593 | -0.000004 | 0.9999 |
| src/compiler/parser.ts:403:1 |  | 0.037816 | 0.037810 | -0.000006 | 0.9998 |
| src/compiler/types.ts:4686:1 | ResolutionDiagnostics | 0.036495 | 0.036481 | -0.000014 | 0.9996 |
| src/compiler/program.ts:5056:1 |  | 0.062961 | 0.062945 | -0.000016 | 0.9997 |
| src/compiler/utilities.ts:8360:1 | getNameOfAccessExpression | 0.046584 | 0.046564 | -0.000020 | 0.9996 |
| src/compiler/types.ts:8135:1 | HasInvalidatedResolutions | 0.035518 | 0.035497 | -0.000021 | 0.9994 |
| src/compiler/program.ts:1015:1 |  | 0.035345 | 0.035323 | -0.000022 | 0.9994 |
| src/compiler/moduleNameResolver.ts:3332:1 | resolveFromTypeRoot | 0.037411 | 0.037388 | -0.000023 | 0.9994 |
| src/compiler/moduleNameResolver.ts:1799:1 | nodeNextJsonConfigResolver | 0.042174 | 0.042148 | -0.000026 | 0.9994 |
| src/compiler/types.ts:2959:1 | ObjectLiteralExpression | 0.036662 | 0.036634 | -0.000028 | 0.9992 |
| src/compiler/sourcemap.ts:389:1 | tryGetSourceMappingURL | 0.112160 | 0.112129 | -0.000031 | 0.9997 |
| src/compiler/tsbuildPublic.ts:216:1 | SolutionBuilderHostBase | 0.035496 | 0.035464 | -0.000032 | 0.9991 |
| src/compiler/visitorPublic.ts:564:1 | visitCommaListElements | 0.036663 | 0.036630 | -0.000033 | 0.9991 |
| src/compiler/utilities.ts:7633:1 | getExpandedCharCodes | 0.037851 | 0.037816 | -0.000035 | 0.9991 |
| src/compiler/program.ts:4877:1 | getDefaultResolutionModeForFileWorker | 0.036834 | 0.036796 | -0.000038 | 0.9990 |
| src/compiler/types.ts:4307:1 | RedirectInfo | 0.037339 | 0.037300 | -0.000039 | 0.9990 |
| src/compiler/transformers/generators.ts:239:1 | BlockAction | 0.036997 | 0.036955 | -0.000042 | 0.9989 |
| src/compiler/builder.ts:1080:1 | IncrementalBuildInfoDiagnostic | 0.035414 | 0.035370 | -0.000044 | 0.9988 |
| src/compiler/factory/nodeTests.ts:1086:1 | isJSDocTypeLiteral | 0.040458 | 0.040411 | -0.000047 | 0.9988 |
| src/compiler/program.ts:693:1 | getCategoryFormat | 0.036616 | 0.036569 | -0.000047 | 0.9987 |
| src/compiler/program.ts:1138:1 |  | 0.035727 | 0.035679 | -0.000048 | 0.9987 |
| src/compiler/scanner.ts:548:1 | getLineAndCharacterOfPosition | 0.038194 | 0.038135 | -0.000059 | 0.9985 |
| src/compiler/types.ts:5726:1 | TypePredicate | 0.034340 | 0.034279 | -0.000062 | 0.9982 |
| src/compiler/utilities.ts:4710:1 | getNextJSDocCommentLocation | 0.170241 | 0.170177 | -0.000064 | 0.9996 |
| src/compiler/types.ts:2864:1 | BigIntLiteral | 0.036039 | 0.035973 | -0.000066 | 0.9982 |
| src/compiler/types.ts:3525:1 | ObjectTypeDeclaration | 0.035333 | 0.035263 | -0.000069 | 0.9980 |
| src/compiler/path.ts:1056:1 | convertToRelativePath | 0.037064 | 0.036992 | -0.000072 | 0.9981 |
| src/compiler/utilities.ts:8068:1 | skipAlias | 0.073997 | 0.073922 | -0.000074 | 0.9990 |
| src/compiler/types.ts:3657:1 | ModuleBlock | 0.036237 | 0.036160 | -0.000077 | 0.9979 |
| src/compiler/path.ts:781:1 | removeTrailingDirectorySeparator | 0.034625 | 0.034547 | -0.000077 | 0.9978 |
| src/compiler/visitorPublic.ts:542:1 | visitIterationBody | 0.042658 | 0.042580 | -0.000078 | 0.9982 |
| src/compiler/types.ts:10091:1 | DiagnosticCollection | 0.036880 | 0.036802 | -0.000078 | 0.9979 |
| src/compiler/parser.ts:9936:1 |  | 0.036195 | 0.036116 | -0.000079 | 0.9978 |
| src/compiler/utilities.ts:10160:1 | numberOfDirectorySeparators | 0.037295 | 0.037211 | -0.000084 | 0.9978 |
| src/compiler/utilitiesPublic.ts:1957:1 | isPropertyAccessOrQualifiedName | 0.042846 | 0.042759 | -0.000086 | 0.9980 |
| src/compiler/types.ts:3709:1 | NamedExportBindings | 0.035757 | 0.035671 | -0.000086 | 0.9976 |
| src/compiler/factory/nodeTests.ts:732:1 | isIfStatement | 0.042136 | 0.042048 | -0.000088 | 0.9979 |
| src/compiler/types.ts:2925:1 | ParenthesizedExpression | 0.037211 | 0.037123 | -0.000088 | 0.9976 |
| src/compiler/factory/nodeFactory.ts:7320:1 | getTransformFlagsSubtreeExclusions | 0.038418 | 0.038328 | -0.000090 | 0.9977 |
| src/compiler/emitter.ts:597:1 | createAddOutput | 0.036880 | 0.036789 | -0.000091 | 0.9975 |
| src/compiler/types.ts:4664:1 | FilePreprocessingDiagnosticsKind | 0.037626 | 0.037535 | -0.000091 | 0.9976 |
| src/compiler/types.ts:3587:1 | HeritageClause | 0.035512 | 0.035417 | -0.000095 | 0.9973 |
| src/compiler/emitter.ts:6345:1 | TempFlags | 0.035915 | 0.035820 | -0.000096 | 0.9973 |
| src/compiler/commandLineParser.ts:1923:1 | ParseCommandLineWorkerDiagnostics | 0.036096 | 0.035999 | -0.000097 | 0.9973 |
| src/compiler/parser.ts:10702:1 |  | 0.035915 | 0.035817 | -0.000098 | 0.9973 |
| src/compiler/types.ts:1632:1 | AbstractKeyword | 0.034515 | 0.034417 | -0.000098 | 0.9972 |
| src/compiler/builder.ts:1658:1 | computeSignature | 0.073723 | 0.073619 | -0.000105 | 0.9986 |
| src/compiler/_namespaces/ts.ts:46:1 |  | 0.040193 | 0.040088 | -0.000105 | 0.9974 |
| src/compiler/moduleSpecifiers.ts:1459:1 | isPathRelativeToParent | 0.038209 | 0.038097 | -0.000112 | 0.9971 |
| src/compiler/utilities.ts:3525:1 | nodeIsDecorated | 0.048748 | 0.048636 | -0.000113 | 0.9977 |
| src/compiler/utilities.ts:6046:1 | getSemanticJsxChildren | 0.038983 | 0.038870 | -0.000114 | 0.9971 |
| src/compiler/utilities.ts:9989:1 |  | 0.036491 | 0.036374 | -0.000117 | 0.9968 |
| src/compiler/types.ts:1756:1 | Declaration | 0.036571 | 0.036452 | -0.000120 | 0.9967 |
| src/compiler/types.ts:5505:1 | UnionReduction | 0.035175 | 0.035051 | -0.000124 | 0.9965 |
| src/compiler/transformers/ts.ts:219:1 | ClassFacts | 0.039422 | 0.039297 | -0.000126 | 0.9968 |
| src/compiler/transformers/generators.ts:291:1 | LoopBlock | 0.037178 | 0.037052 | -0.000126 | 0.9966 |
| src/compiler/parser.ts:10591:1 | processCommentPragmas | 0.040343 | 0.040206 | -0.000137 | 0.9966 |
| src/compiler/checker.ts:1384:1 | MappedTypeNameTypeKind | 0.036690 | 0.036551 | -0.000139 | 0.9962 |
| src/compiler/transformers/namedEvaluation.ts:405:1 | transformNamedEvaluationOfAssignmentExpression | 0.041319 | 0.041176 | -0.000143 | 0.9965 |
| src/compiler/utilitiesPublic.ts:1131:1 | getJSDocPublicTagNoCache | 0.050831 | 0.050687 | -0.000145 | 0.9972 |
| src/compiler/commandLineParser.ts:3304:1 |  | 0.036914 | 0.036767 | -0.000147 | 0.9960 |
| src/compiler/types.ts:3618:1 | ModuleName | 0.035394 | 0.035244 | -0.000150 | 0.9958 |
| src/compiler/types.ts:3652:1 | JSDocNamespaceDeclaration | 0.036301 | 0.036151 | -0.000150 | 0.9959 |
| src/compiler/tsbuildPublic.ts:831:1 | InvalidatedProjectKind | 0.036651 | 0.036498 | -0.000153 | 0.9958 |
| src/compiler/factory/nodeFactory.ts:7409:1 |  | 0.042418 | 0.042263 | -0.000155 | 0.9964 |
| src/compiler/utilitiesPublic.ts:2477:1 | isJsxAttributeLike | 0.040138 | 0.039981 | -0.000156 | 0.9961 |
| src/compiler/factory/nodeConverters.ts:183:1 |  | 0.064003 | 0.063843 | -0.000159 | 0.9975 |
| src/compiler/scanner.ts:331:1 |  | 0.039635 | 0.039476 | -0.000160 | 0.9960 |
| src/compiler/tsbuildPublic.ts:241:1 | SolutionBuilderHost | 0.035016 | 0.034853 | -0.000163 | 0.9953 |
| src/compiler/path.ts:140:1 | hasTrailingDirectorySeparator | 0.035656 | 0.035493 | -0.000163 | 0.9954 |
| src/compiler/types.ts:5761:1 | AnyImportOrReExport | 0.035388 | 0.035222 | -0.000166 | 0.9953 |
| src/compiler/checker.ts:1152:1 |  | 0.035585 | 0.035419 | -0.000166 | 0.9953 |
| src/compiler/utilitiesPublic.ts:1744:1 | isClassLike | 0.041757 | 0.041588 | -0.000170 | 0.9959 |
| src/compiler/utilities.ts:8872:1 | compareMessageChainSize | 0.036533 | 0.036363 | -0.000170 | 0.9953 |
| src/compiler/path.ts:622:1 | getNormalizedPathComponents | 0.034981 | 0.034811 | -0.000170 | 0.9951 |
| src/compiler/types.ts:3475:1 | CaseBlock | 0.036286 | 0.036113 | -0.000173 | 0.9952 |
| src/compiler/types.ts:9706:1 | Transformer | 0.036935 | 0.036762 | -0.000174 | 0.9953 |
| src/compiler/path.ts:579:1 | combinePaths | 0.035535 | 0.035341 | -0.000194 | 0.9945 |
| src/compiler/transformers/declarations.ts:1954:1 | isProcessedComponent | 0.040875 | 0.040678 | -0.000196 | 0.9952 |
| src/compiler/utilities.ts:2434:1 | createDiagnosticForNodeFromMessageChain | 0.043810 | 0.043605 | -0.000204 | 0.9953 |
| src/compiler/utilities.ts:10409:1 | skipTypeCheckingIgnoringNoCheck | 0.037518 | 0.037308 | -0.000210 | 0.9944 |
| src/compiler/program.ts:636:1 | getPreEmitDiagnostics | 0.038192 | 0.037980 | -0.000212 | 0.9944 |
| src/compiler/types.ts:2468:1 | FalseLiteral | 0.036091 | 0.035879 | -0.000213 | 0.9941 |
| src/compiler/factory/nodeTests.ts:1009:1 | isEnumMember | 0.041338 | 0.041125 | -0.000213 | 0.9948 |
| src/compiler/types.ts:3719:1 | ImportClause | 0.036047 | 0.035830 | -0.000217 | 0.9940 |
| src/compiler/utilities.ts:997:1 | isPlainJsFile | 0.038543 | 0.038316 | -0.000227 | 0.9941 |
| src/compiler/moduleNameResolver.ts:2028:1 | pathContainsNodeModules | 0.036563 | 0.036331 | -0.000232 | 0.9937 |
| src/compiler/resolutionCache.ts:338:1 | canWatchAffectedPackageJsonOrNodeModulesOfAtTypes | 0.037303 | 0.037069 | -0.000234 | 0.9937 |
| src/compiler/watch.ts:713:1 | WatchTypeRegistry | 0.040599 | 0.040363 | -0.000236 | 0.9942 |
| src/compiler/sys.ts:51:1 | setTimeout | 0.036639 | 0.036401 | -0.000237 | 0.9935 |
| src/compiler/utilitiesPublic.ts:2108:1 | isAssertionExpression | 0.040730 | 0.040491 | -0.000239 | 0.9941 |
| src/compiler/types.ts:5498:1 | MemberOverrideStatus | 0.035363 | 0.035123 | -0.000240 | 0.9932 |
| src/compiler/transformers/utilities.ts:589:1 | getProperties | 0.036861 | 0.036621 | -0.000240 | 0.9935 |
| src/compiler/commandLineParser.ts:2158:1 | getBuildOptionsNameMap | 0.037516 | 0.037275 | -0.000240 | 0.9936 |
| src/compiler/utilities.ts:3537:1 | nodeOrChildIsDecorated | 0.062332 | 0.062082 | -0.000250 | 0.9960 |
| src/compiler/types.ts:8313:1 | SourceMapSource | 0.036626 | 0.036376 | -0.000250 | 0.9932 |
| src/compiler/utilities.ts:8295:1 | getLastChild | 0.046896 | 0.046642 | -0.000254 | 0.9946 |
| src/compiler/path.ts:722:1 | normalizePath | 0.035672 | 0.035415 | -0.000257 | 0.9928 |
| src/compiler/tsbuildPublic.ts:249:1 | BuildOrder | 0.034649 | 0.034385 | -0.000264 | 0.9924 |
| src/compiler/tsbuildPublic.ts:264:1 | getBuildOrderFromAnyBuildOrder | 0.061912 | 0.061645 | -0.000267 | 0.9957 |
| src/compiler/factory/utilities.ts:1246:1 | isLogicalOperator | 0.039664 | 0.039396 | -0.000268 | 0.9932 |
| src/compiler/utilities.ts:8580:1 | setObjectAllocator | 0.036469 | 0.036201 | -0.000268 | 0.9926 |
| src/compiler/factory/utilities.ts:362:1 | createExpressionForAccessorDeclaration | 0.039322 | 0.039054 | -0.000269 | 0.9932 |
| src/compiler/types.ts:3551:1 | ClassDeclaration | 0.036139 | 0.035867 | -0.000272 | 0.9925 |
| src/compiler/program.ts:1043:1 |  | 0.071481 | 0.071207 | -0.000274 | 0.9962 |
| src/compiler/visitorPublic.ts:513:1 | visitFunctionBody | 0.035814 | 0.035539 | -0.000275 | 0.9923 |
| src/compiler/transformers/utilities.ts:148:1 | getExportNeedsImportStarHelper | 0.037782 | 0.037505 | -0.000277 | 0.9927 |
| src/compiler/types.ts:2474:1 | ThisExpression | 0.036293 | 0.036016 | -0.000277 | 0.9924 |
| src/compiler/binder.ts:495:1 | createFlowNode | 0.048323 | 0.048044 | -0.000279 | 0.9942 |
| src/compiler/resolutionCache.ts:348:1 | getDirectoryToWatchFailedLookupLocation | 0.037458 | 0.037178 | -0.000280 | 0.9925 |
| src/compiler/moduleNameResolver.ts:2581:1 | loadModuleFromSelfNameReference | 0.039409 | 0.039129 | -0.000280 | 0.9929 |
| src/compiler/core.ts:2402:1 | matchedText | 0.036852 | 0.036569 | -0.000283 | 0.9923 |
| src/compiler/moduleNameResolver.ts:1769:1 | bundlerModuleNameResolver | 0.041934 | 0.041648 | -0.000287 | 0.9932 |
| src/compiler/scanner.ts:1015:1 | ClassSetExpressionType | 0.035714 | 0.035426 | -0.000288 | 0.9919 |
| src/compiler/utilities.ts:7321:1 | getJSDocModifierFlagsNoCache | 0.044345 | 0.044053 | -0.000292 | 0.9934 |
| src/compiler/utilities.ts:8082:1 | isWriteOnlyAccess | 0.043358 | 0.043065 | -0.000292 | 0.9933 |
| src/compiler/factory/nodeTests.ts:536:1 | isIndexedAccessTypeNode | 0.040580 | 0.040286 | -0.000294 | 0.9928 |
| src/compiler/factory/nodeTests.ts:322:1 | isPrivateIdentifier | 0.044060 | 0.043765 | -0.000296 | 0.9933 |
| src/compiler/utilitiesPublic.ts:1294:1 | getTextOfJSDocComment | 0.038461 | 0.038164 | -0.000297 | 0.9923 |
| src/compiler/utilities.ts:10166:1 | compareNumberOfDirectorySeparators | 0.038787 | 0.038484 | -0.000303 | 0.9922 |
| src/compiler/emitter.ts:685:1 | getCommonSourceDirectoryOfConfig | 0.107331 | 0.107028 | -0.000304 | 0.9972 |
| src/compiler/types.ts:9569:1 | CoreTransformationContext | 0.038079 | 0.037772 | -0.000307 | 0.9919 |
| src/compiler/utilities.ts:980:1 | getSourceFileOfNode | 0.042180 | 0.041872 | -0.000308 | 0.9927 |
| src/compiler/transformers/typeSerializer.ts:73:1 | SerializedEntityName | 0.037142 | 0.036830 | -0.000312 | 0.9916 |
| src/compiler/utilities.ts:7231:1 | hasEffectiveReadonlyModifier | 0.043732 | 0.043418 | -0.000314 | 0.9928 |
| src/compiler/scanner.ts:1023:1 | createScanner | 0.094981 | 0.094666 | -0.000315 | 0.9967 |
| src/compiler/factory/nodeTests.ts:764:1 | isReturnStatement | 0.043525 | 0.043209 | -0.000316 | 0.9927 |
| src/compiler/utilities.ts:620:1 |  | 0.036466 | 0.036147 | -0.000319 | 0.9913 |
| src/compiler/utilities.ts:9937:1 | ensureScriptKind | 0.039863 | 0.039543 | -0.000320 | 0.9920 |
| src/compiler/moduleSpecifiers.ts:1122:1 | tryGetModuleNameFromExports | 0.049209 | 0.048886 | -0.000323 | 0.9934 |
| src/compiler/builder.ts:744:1 | getBuilderFileEmitAllDts | 0.037837 | 0.037511 | -0.000326 | 0.9914 |
| src/compiler/tsbuildPublic.ts:348:1 | ConfigFileCacheEntry | 0.038342 | 0.038015 | -0.000327 | 0.9915 |
| src/compiler/types.ts:5076:1 | TypeChecker | 0.036034 | 0.035702 | -0.000332 | 0.9908 |
| src/compiler/types.ts:3426:1 | ForInOrOfStatement | 0.036508 | 0.036167 | -0.000340 | 0.9907 |
| src/compiler/utilities.ts:7609:1 | getLocalSymbolForExportDefault | 0.038363 | 0.038017 | -0.000346 | 0.9910 |
| src/compiler/transformers/classFields.ts:283:1 | PrivateIdentifierInstanceFieldInfo | 0.036193 | 0.035846 | -0.000347 | 0.9904 |
| src/compiler/utilitiesPublic.ts:487:1 | collapseTextChangeRangesAcrossMultipleVersions | 0.037451 | 0.037100 | -0.000351 | 0.9906 |
| src/compiler/factory/nodeTests.ts:1100:1 | isJSDocAuthorTag | 0.041242 | 0.040888 | -0.000353 | 0.9914 |
| src/compiler/utilities.ts:4119:1 | isBindableObjectDefinePropertyCall | 0.043918 | 0.043562 | -0.000356 | 0.9919 |
| src/compiler/transformers/utilities.ts:729:1 | getAllDecoratorsOfAccessors | 0.043058 | 0.042698 | -0.000360 | 0.9916 |
| src/compiler/types.ts:3543:1 | ClassLikeDeclarationBase | 0.035831 | 0.035469 | -0.000362 | 0.9899 |
| src/compiler/factory/nodeTests.ts:1026:1 | isJSDocTypeExpression | 0.043895 | 0.043532 | -0.000363 | 0.9917 |
| src/compiler/semver.ts:297:1 | parsePartial | 0.038793 | 0.038429 | -0.000364 | 0.9906 |
| src/compiler/factory/nodeTests.ts:668:1 | isOmittedExpression | 0.039984 | 0.039621 | -0.000364 | 0.9909 |
| src/compiler/moduleNameResolver.ts:944:1 | compilerOptionValueToString | 0.037510 | 0.037145 | -0.000365 | 0.9903 |
| src/compiler/utilitiesPublic.ts:882:1 | nameForNamelessJSDocTypedef | 0.038201 | 0.037835 | -0.000366 | 0.9904 |
| src/compiler/types.ts:3945:1 | JSDocTypeReferencingNode | 0.035766 | 0.035400 | -0.000367 | 0.9898 |
| src/compiler/types.ts:2074:1 | FunctionLikeDeclaration | 0.036687 | 0.036315 | -0.000373 | 0.9898 |
| src/compiler/types.ts:782:1 | NodeFlags | 0.035964 | 0.035589 | -0.000376 | 0.9896 |
| src/compiler/builderPublic.ts:150:1 | EmitAndSemanticDiagnosticsBuilderProgram | 0.037285 | 0.036909 | -0.000376 | 0.9899 |
| src/compiler/watchPublic.ts:155:1 | WatchStatusReporter | 0.037136 | 0.036758 | -0.000378 | 0.9898 |
| src/compiler/factory/nodeTests.ts:1116:1 | isJSDocPrivateTag | 0.041123 | 0.040742 | -0.000381 | 0.9907 |
| src/compiler/_namespaces/ts.ts:74:1 |  | 0.036568 | 0.036186 | -0.000382 | 0.9896 |
| src/compiler/types.ts:2815:1 | TokenFlags | 0.037048 | 0.036666 | -0.000382 | 0.9897 |
| src/compiler/program.ts:845:1 | SourceFileImportsList | 0.037617 | 0.037234 | -0.000383 | 0.9898 |
| src/compiler/watchUtilities.ts:118:1 | createCachedDirectoryStructureHost | 0.036953 | 0.036570 | -0.000384 | 0.9896 |
| src/compiler/path.ts:783:1 | removeTrailingDirectorySeparator | 0.036358 | 0.035974 | -0.000384 | 0.9894 |
| src/compiler/utilitiesPublic.ts:1289:1 | getAllJSDocTagsOfKind | 0.049711 | 0.049326 | -0.000384 | 0.9923 |
| src/compiler/commandLineParser.ts:3381:1 | ParsedTsconfig | 0.038521 | 0.038135 | -0.000386 | 0.9900 |
| src/compiler/utilities.ts:7012:1 | emitDetachedComments | 0.037799 | 0.037411 | -0.000388 | 0.9897 |
| src/compiler/corePublic.ts:6:1 |  | 0.035846 | 0.035454 | -0.000392 | 0.9891 |
| src/compiler/utilitiesPublic.ts:1274:1 | getJSDocTags | 0.046308 | 0.045912 | -0.000396 | 0.9915 |
| src/compiler/utilitiesPublic.ts:2603:1 | guessIndentation | 0.037385 | 0.036989 | -0.000396 | 0.9894 |
| src/compiler/utilitiesPublic.ts:678:1 | getCombinedNodeFlags | 0.046208 | 0.045812 | -0.000397 | 0.9914 |
| src/compiler/types.ts:5686:1 | TypePredicateKind | 0.035814 | 0.035415 | -0.000399 | 0.9888 |
| src/compiler/utilities.ts:8508:1 | Node | 0.042062 | 0.041660 | -0.000401 | 0.9905 |
| src/compiler/factory/nodeFactory.ts:463:1 |  | 0.035708 | 0.035303 | -0.000405 | 0.9887 |
| src/compiler/factory/utilities.ts:612:1 | isCommaSequence | 0.042553 | 0.042148 | -0.000405 | 0.9905 |
| src/compiler/transformers/es2015.ts:371:1 | HierarchyFacts | 0.036008 | 0.035600 | -0.000408 | 0.9887 |
| src/compiler/sys.ts:554:1 | getFileWatcherEventKind | 0.038809 | 0.038397 | -0.000412 | 0.9894 |
| src/compiler/semver.ts:34:1 |  | 0.037310 | 0.036898 | -0.000412 | 0.9890 |
| src/compiler/utilities.ts:8567:1 |  | 0.037087 | 0.036672 | -0.000415 | 0.9888 |
| src/compiler/utilities.ts:2735:1 | isCustomPrologue | 0.044369 | 0.043952 | -0.000417 | 0.9906 |
| src/compiler/_namespaces/ts.ts:70:1 |  | 0.035705 | 0.035286 | -0.000418 | 0.9883 |
| src/compiler/types.ts:4278:1 | AmdDependency | 0.036707 | 0.036288 | -0.000419 | 0.9886 |
| src/compiler/types.ts:2597:1 | LogicalOperator | 0.038630 | 0.038210 | -0.000420 | 0.9891 |
| src/compiler/utilities.ts:644:1 | createSymbolTable | 0.037876 | 0.037455 | -0.000421 | 0.9889 |
| src/compiler/sys.ts:1454:1 | FileWatcher | 0.039859 | 0.039436 | -0.000422 | 0.9894 |
| src/compiler/types.ts:5574:1 | InternalNodeBuilderFlags | 0.035589 | 0.035159 | -0.000430 | 0.9879 |
| src/compiler/_namespaces/ts.ts:7:1 |  | 0.037542 | 0.037112 | -0.000431 | 0.9885 |
| src/compiler/types.ts:4483:1 | CommentDirective | 0.037414 | 0.036979 | -0.000436 | 0.9884 |
| src/compiler/transformers/typeSerializer.ts:1:1 |  | 0.038137 | 0.037700 | -0.000437 | 0.9885 |
| src/compiler/utilities.ts:8902:1 | compareMessageChainContent | 0.037303 | 0.036865 | -0.000438 | 0.9883 |
| src/compiler/utilities.ts:4842:1 | AssignmentKind | 0.035632 | 0.035188 | -0.000445 | 0.9875 |
| src/compiler/utilitiesPublic.ts:2086:1 | isExpression | 0.042589 | 0.042143 | -0.000446 | 0.9895 |
| src/compiler/factory/utilities.ts:247:1 | createExpressionForJsxElement | 0.126516 | 0.126070 | -0.000446 | 0.9965 |
| src/compiler/types.ts:2486:1 | DeleteExpression | 0.037123 | 0.036677 | -0.000446 | 0.9880 |
| src/compiler/transformers/utilities.ts:130:1 | isNamedDefaultReference | 0.039695 | 0.039247 | -0.000448 | 0.9887 |
| src/compiler/resolutionCache.ts:176:1 | CachedResolvedModuleWithFailedLookupLocations | 0.036501 | 0.036053 | -0.000448 | 0.9877 |
| src/compiler/utilities.ts:7950:1 | getLinesBetweenRangeEndPositions | 0.039092 | 0.038642 | -0.000450 | 0.9885 |
| src/compiler/types.ts:2472:1 | BooleanLiteral | 0.035522 | 0.035070 | -0.000452 | 0.9873 |
| src/compiler/moduleNameResolver.ts:1115:1 | createModeAwareCacheKey | 0.037950 | 0.037497 | -0.000453 | 0.9881 |
| src/compiler/types.ts:7168:1 | InferenceContext | 0.035398 | 0.034940 | -0.000458 | 0.9871 |
| src/compiler/transformers/module/esnextAnd2015.ts:1:1 |  | 0.037385 | 0.036923 | -0.000462 | 0.9876 |
| src/compiler/types.ts:3984:1 | JSDocText | 0.034843 | 0.034373 | -0.000470 | 0.9865 |
| src/compiler/transformers/destructuring.ts:66:1 | FlattenContext | 0.035660 | 0.035182 | -0.000478 | 0.9866 |
| src/compiler/factory/nodeTests.ts:372:1 | isAccessorModifier | 0.044788 | 0.044300 | -0.000488 | 0.9891 |
| src/compiler/_namespaces/ts.ts:62:1 |  | 0.036460 | 0.035972 | -0.000488 | 0.9866 |
| src/compiler/types.ts:1639:1 | InKeyword | 0.036176 | 0.035687 | -0.000490 | 0.9865 |
| src/compiler/types.ts:5654:1 | SymbolWalker | 0.036048 | 0.035557 | -0.000491 | 0.9864 |
| src/compiler/types.ts:8095:1 | ResolvedModuleWithFailedLookupLocations | 0.037490 | 0.036989 | -0.000501 | 0.9866 |
| src/compiler/types.ts:2501:1 | AwaitExpression | 0.037313 | 0.036810 | -0.000502 | 0.9865 |
| src/compiler/utilities.ts:7786:1 | readJsonOrUndefined | 0.100389 | 0.099878 | -0.000511 | 0.9949 |
| src/compiler/factory/nodeTests.ts:931:1 | isJsxSelfClosingElement | 0.042336 | 0.041823 | -0.000513 | 0.9879 |
| src/compiler/transformers/utilities.ts:547:1 | findSuperStatementIndexPathWorker | 0.036782 | 0.036267 | -0.000514 | 0.9860 |
| src/compiler/transformers/taggedTemplate.ts:110:1 | getRawLiteral | 0.046669 | 0.046147 | -0.000522 | 0.9888 |
| src/compiler/executeCommandLine.ts:1156:1 | isSolutionMarkOrMeasure | 0.035173 | 0.034647 | -0.000526 | 0.9851 |
| src/compiler/scanner.ts:344:1 |  | 0.044472 | 0.043946 | -0.000527 | 0.9882 |
| src/compiler/moduleNameResolver.ts:966:1 | CacheWithRedirects | 0.037980 | 0.037452 | -0.000528 | 0.9861 |
| src/compiler/_namespaces/ts.ts:23:1 |  | 0.037175 | 0.036643 | -0.000532 | 0.9857 |
| src/compiler/program.ts:1175:1 | ReferenceFileLocation | 0.037053 | 0.036521 | -0.000532 | 0.9856 |
| src/compiler/factory/utilities.ts:1673:1 | findComputedPropertyNameCacheAssignment | 0.038878 | 0.038343 | -0.000535 | 0.9862 |
| src/compiler/utilities.ts:7983:1 | rangeContainsRange | 0.043755 | 0.043219 | -0.000536 | 0.9878 |
| src/compiler/utilities.ts:11082:1 | isJSDocOptionalParameter | 0.060231 | 0.059695 | -0.000536 | 0.9911 |
| src/compiler/types.ts:2416:1 | IncrementExpression | 0.038064 | 0.037528 | -0.000536 | 0.9859 |
| src/compiler/types.ts:3419:1 | ForStatement | 0.037189 | 0.036653 | -0.000536 | 0.9856 |
| src/compiler/utilitiesPublic.ts:408:1 | textRangeIntersectsWithTextSpan | 0.037391 | 0.036851 | -0.000540 | 0.9856 |
| src/compiler/path.ts:371:1 | getBaseFileName | 0.036497 | 0.035955 | -0.000542 | 0.9851 |
| src/compiler/transformers/utilities.ts:135:1 | chainBundle | 0.080644 | 0.080099 | -0.000545 | 0.9932 |
| src/compiler/types.ts:3935:1 | JSDocVariadicType | 0.035689 | 0.035138 | -0.000551 | 0.9846 |
| src/compiler/types.ts:3926:1 | JSDocOptionalType | 0.035610 | 0.035058 | -0.000553 | 0.9845 |
| src/compiler/utilities.ts:3604:1 | isEmptyStringLiteral | 0.040073 | 0.039518 | -0.000555 | 0.9862 |
| src/compiler/utilities.ts:9397:1 | compilerOptionsAffectDeclarationPath | 0.038237 | 0.037681 | -0.000556 | 0.9855 |
| src/compiler/types.ts:758:1 | JSDocSyntaxKind | 0.036559 | 0.036001 | -0.000558 | 0.9847 |
| src/compiler/types.ts:4975:1 | CustomTransformer | 0.035077 | 0.034516 | -0.000561 | 0.9840 |
| src/compiler/factory/nodeFactory.ts:1:1 |  | 0.035776 | 0.035213 | -0.000563 | 0.9843 |
| src/compiler/utilitiesPublic.ts:2625:1 | isStringLiteralLike | 0.040399 | 0.039835 | -0.000565 | 0.9860 |
| src/compiler/utilities.ts:7221:1 | hasAmbientModifier | 0.043091 | 0.042525 | -0.000566 | 0.9869 |
| src/compiler/factory/utilities.ts:1518:1 | getNodeForGeneratedName | 0.037556 | 0.036988 | -0.000568 | 0.9849 |
| src/compiler/types.ts:3506:1 | ThrowStatement | 0.036168 | 0.035594 | -0.000573 | 0.9841 |
| src/compiler/builderPublic.ts:22:1 | BuilderProgramHost | 0.037406 | 0.036829 | -0.000578 | 0.9846 |
| src/compiler/utilities.ts:3618:1 | isJSXTagName | 0.078176 | 0.077598 | -0.000578 | 0.9926 |
| src/compiler/transformers/generators.ts:262:1 | CodeBlock | 0.037844 | 0.037264 | -0.000580 | 0.9847 |
| src/compiler/watchUtilities.ts:89:1 | CachedDirectoryStructureHost | 0.036503 | 0.035922 | -0.000582 | 0.9841 |
| src/compiler/types.ts:4692:1 | FilePreprocessingDiagnostics | 0.036365 | 0.035782 | -0.000583 | 0.9840 |
| src/compiler/types.ts:2719:1 | AssignmentElementTarget | 0.037083 | 0.036499 | -0.000584 | 0.9843 |
| src/compiler/tsbuildPublic.ts:245:1 | SolutionBuilderWithWatchHost | 0.035001 | 0.034414 | -0.000588 | 0.9832 |
| src/compiler/types.ts:3870:1 | CommentKind | 0.034949 | 0.034360 | -0.000589 | 0.9832 |
| src/compiler/types.ts:3902:1 | JSDocType | 0.035703 | 0.035113 | -0.000590 | 0.9835 |
| src/compiler/builder.ts:1095:1 | IncrementalBuildInfoReferencedMap | 0.035767 | 0.035174 | -0.000593 | 0.9834 |
| src/compiler/sourcemap.ts:377:1 | getLineInfo | 0.038629 | 0.038033 | -0.000596 | 0.9846 |
| src/compiler/programDiagnostics.ts:72:1 | ProgramDiagnostics | 0.037612 | 0.037013 | -0.000599 | 0.9841 |
| src/compiler/factory/utilities.ts:1175:1 | isMultiplicativeOperator | 0.042204 | 0.041602 | -0.000603 | 0.9857 |
| src/compiler/commandLineParser.ts:2132:1 | getOptionFromName | 0.037146 | 0.036543 | -0.000603 | 0.9838 |
| src/compiler/checker.ts:1171:1 |  | 0.035697 | 0.035094 | -0.000603 | 0.9831 |
| src/compiler/utilities.ts:4830:1 | getTypeParameterFromJsDoc | 0.036118 | 0.035514 | -0.000603 | 0.9833 |
| src/compiler/types.ts:2705:1 | BindingOrAssignmentElementRestIndicator | 0.037766 | 0.037161 | -0.000605 | 0.9840 |
| src/compiler/types.ts:2259:1 | TypeQueryNode | 0.036769 | 0.036163 | -0.000605 | 0.9835 |
| src/compiler/resolutionCache.ts:220:1 | FileWatcherOfAffectingLocation | 0.037727 | 0.037118 | -0.000609 | 0.9838 |
| src/compiler/moduleSpecifiers.ts:686:1 | countPathComponents | 0.039041 | 0.038430 | -0.000611 | 0.9844 |
| src/compiler/path.ts:841:1 | changeAnyExtension | 0.035186 | 0.034573 | -0.000613 | 0.9826 |
| src/compiler/types.ts:3822:1 | ImportOrExportSpecifier | 0.036422 | 0.035808 | -0.000614 | 0.9831 |
| src/compiler/utilities.ts:8855:1 | compareMessageChain | 0.036957 | 0.036342 | -0.000615 | 0.9834 |
| src/compiler/path.ts:853:1 | changeAnyExtension | 0.035839 | 0.035222 | -0.000617 | 0.9828 |
| src/compiler/types.ts:2603:1 | LogicalOperatorOrHigher | 0.038107 | 0.037489 | -0.000618 | 0.9838 |
| src/compiler/checker.ts:1402:1 | UnusedKind | 0.037507 | 0.036884 | -0.000622 | 0.9834 |
| src/compiler/factory/nodeTests.ts:1030:1 | isJSDocNameReference | 0.041012 | 0.040384 | -0.000628 | 0.9847 |
| src/compiler/types.ts:2637:1 | BinaryOperator | 0.038043 | 0.037413 | -0.000630 | 0.9834 |
| src/compiler/utilitiesPublic.ts:1843:1 | isAssignmentPattern | 0.041991 | 0.041360 | -0.000630 | 0.9850 |
| src/compiler/resolutionCache.ts:269:1 | perceivedOsRootLengthForWatching | 0.037951 | 0.037319 | -0.000632 | 0.9833 |
| src/compiler/types.ts:2215:1 | LiteralImportTypeNode | 0.038153 | 0.037520 | -0.000633 | 0.9834 |
| src/compiler/types.ts:3931:1 | JSDocFunctionType | 0.035765 | 0.035131 | -0.000635 | 0.9823 |
| src/compiler/utilitiesPublic.ts:380:1 | textSpanOverlapsWith | 0.035445 | 0.034810 | -0.000635 | 0.9821 |
| src/compiler/factory/nodeTests.ts:943:1 | isJsxFragment | 0.041285 | 0.040649 | -0.000636 | 0.9846 |
| src/compiler/utilities.ts:4224:1 | getElementOrPropertyAccessName | 0.190275 | 0.189638 | -0.000636 | 0.9967 |
| src/compiler/utilities.ts:5584:1 | isPushOrUnshiftIdentifier | 0.036564 | 0.035925 | -0.000638 | 0.9825 |
| src/compiler/types.ts:3338:1 | CommaListExpression | 0.037217 | 0.036575 | -0.000642 | 0.9827 |
| src/compiler/moduleNameResolver.ts:3001:1 | loadModuleFromNearestNodeModulesDirectoryTypesScope | 0.038208 | 0.037561 | -0.000647 | 0.9831 |
| src/compiler/types.ts:3648:1 | JSDocNamespaceBody | 0.036742 | 0.036093 | -0.000648 | 0.9824 |
| src/compiler/types.ts:10324:1 | PragmaArgTypeMaybeCapture | 0.039196 | 0.038544 | -0.000652 | 0.9834 |
| src/compiler/commandLineParser.ts:1825:1 | OptionsNameMap | 0.036122 | 0.035463 | -0.000658 | 0.9818 |
| src/compiler/executeCommandLine.ts:554:1 | printHelp | 0.038173 | 0.037513 | -0.000660 | 0.9827 |
| src/compiler/utilities.ts:2669:1 | isVarUsing | 0.042184 | 0.041522 | -0.000662 | 0.9843 |
| src/compiler/executeCommandLine.ts:244:1 | generateOptionOutput | 0.039502 | 0.038837 | -0.000665 | 0.9832 |
| src/compiler/factory/nodeTests.ts:1160:1 | isJSDocTypeTag | 0.043315 | 0.042647 | -0.000668 | 0.9846 |
| src/compiler/types.ts:2690:1 | ArrayBindingOrAssignmentElement | 0.038500 | 0.037824 | -0.000677 | 0.9824 |
| src/compiler/types.ts:3684:1 | ExternalModuleReference | 0.035904 | 0.035226 | -0.000678 | 0.9811 |
| src/compiler/_namespaces/ts.ts:44:1 |  | 0.039354 | 0.038675 | -0.000679 | 0.9827 |
| src/compiler/utilitiesPublic.ts:1710:1 | isFunctionLikeKind | 0.045138 | 0.044457 | -0.000681 | 0.9849 |
| src/compiler/types.ts:1960:1 | PrivateIdentifierGetAccessorDeclaration | 0.035529 | 0.034838 | -0.000691 | 0.9805 |
| src/compiler/utilities.ts:7163:1 | calculateIndent | 0.037437 | 0.036742 | -0.000695 | 0.9814 |
| src/compiler/utilitiesPublic.ts:2406:1 | isDeclarationStatement | 0.042640 | 0.041944 | -0.000697 | 0.9837 |
| src/compiler/types.ts:5732:1 | AnyImportOrJsDocImport | 0.035767 | 0.035068 | -0.000699 | 0.9804 |
| src/compiler/watchUtilities.ts:381:1 | ProgramUpdateLevel | 0.036703 | 0.035997 | -0.000706 | 0.9808 |
| src/compiler/types.ts:4706:1 | Program | 0.036499 | 0.035792 | -0.000707 | 0.9806 |
| src/compiler/types.ts:4958:1 | RedirectTargetsMap | 0.036888 | 0.036179 | -0.000709 | 0.9808 |
| src/compiler/types.ts:5719:1 | AssertsIdentifierTypePredicate | 0.036267 | 0.035556 | -0.000711 | 0.9804 |
| src/compiler/factory/nodeTests.ts:660:1 | isSpreadElement | 0.041372 | 0.040660 | -0.000712 | 0.9828 |
| src/compiler/factory/nodeTests.ts:1003:1 | isSpreadAssignment | 0.040944 | 0.040229 | -0.000715 | 0.9825 |
| src/compiler/utilitiesPublic.ts:2071:1 | isLiteralTypeLiteral | 0.043693 | 0.042977 | -0.000716 | 0.9836 |
| src/compiler/factory/utilities.ts:424:1 | createExpressionForShorthandPropertyAssignment | 0.088102 | 0.087382 | -0.000720 | 0.9918 |
| src/compiler/types.ts:2631:1 | AssignmentOperatorOrHigher | 0.038383 | 0.037664 | -0.000720 | 0.9813 |
| src/compiler/utilities.ts:7001:1 | DetachedCommentInfo | 0.036224 | 0.035504 | -0.000720 | 0.9801 |
| src/compiler/types.ts:2942:1 | SpreadElement | 0.036532 | 0.035812 | -0.000720 | 0.9803 |
| src/compiler/moduleNameResolver.ts:184:1 | Extensions | 0.035675 | 0.034954 | -0.000721 | 0.9798 |
| src/compiler/path.ts:495:1 | getPathComponents | 0.036081 | 0.035359 | -0.000722 | 0.9800 |
| src/compiler/utilities.ts:9438:1 | hasZeroOrOneAsteriskCharacter | 0.037205 | 0.036482 | -0.000723 | 0.9806 |
| src/compiler/tsbuildPublic.ts:374:1 | SolutionBuilderState | 0.036470 | 0.035741 | -0.000729 | 0.9800 |
| src/compiler/visitorPublic.ts:540:1 | visitIterationBody | 0.042553 | 0.041823 | -0.000730 | 0.9828 |
| src/compiler/corePublic.ts:13:1 | MapLike | 0.036629 | 0.035899 | -0.000730 | 0.9801 |
| src/compiler/types.ts:1864:1 | CallSignatureDeclaration | 0.037762 | 0.037031 | -0.000730 | 0.9807 |
| src/compiler/path.ts:457:1 | PathPathComponents | 0.035174 | 0.034440 | -0.000735 | 0.9791 |
| src/compiler/visitorPublic.ts:519:18 | visitFunctionBody | 0.036121 | 0.035385 | -0.000737 | 0.9796 |
| src/compiler/factory/nodeTests.ts:1168:1 | isJSDocTypedefTag | 0.041416 | 0.040679 | -0.000737 | 0.9822 |
| src/compiler/performance.ts:65:1 |  | 0.036648 | 0.035909 | -0.000739 | 0.9798 |
| src/compiler/moduleNameResolver.ts:1161:1 | getOriginalOrResolvedTypeReferenceFileName | 0.037052 | 0.036297 | -0.000755 | 0.9796 |
| src/compiler/performance.ts:64:1 |  | 0.035857 | 0.035096 | -0.000761 | 0.9788 |
| src/compiler/utilities.ts:1316:1 | moduleExportNameIsDefault | 0.036737 | 0.035973 | -0.000764 | 0.9792 |
| src/compiler/transformers/esDecorators.ts:234:1 | ClassInfo | 0.035068 | 0.034304 | -0.000764 | 0.9782 |
| src/compiler/transformers/esDecorators.ts:288:1 | LexicalEnvironmentStackEntry | 0.034890 | 0.034121 | -0.000768 | 0.9780 |
| src/compiler/types.ts:1637:1 | DefaultKeyword | 0.035098 | 0.034328 | -0.000770 | 0.9781 |
| src/compiler/utilitiesPublic.ts:384:1 | textSpanOverlap | 0.035663 | 0.034888 | -0.000775 | 0.9783 |
| src/compiler/utilities.ts:2337:1 | forEachEnclosingBlockScopeContainer | 0.047380 | 0.046598 | -0.000782 | 0.9835 |
| src/compiler/utilitiesPublic.ts:312:1 |  | 0.036144 | 0.035362 | -0.000782 | 0.9784 |
| src/compiler/utilitiesPublic.ts:2398:1 | isDeclaration | 0.047656 | 0.046874 | -0.000782 | 0.9836 |
| src/compiler/factory/nodeTests.ts:1054:1 | isJSDocUnknownType | 0.041931 | 0.041144 | -0.000787 | 0.9812 |
| src/compiler/builder.ts:187:1 | BuilderFileEmit | 0.048142 | 0.047355 | -0.000787 | 0.9836 |
| src/compiler/types.ts:2223:1 | FunctionOrConstructorTypeNodeBase | 0.037471 | 0.036674 | -0.000797 | 0.9787 |
| src/compiler/utilitiesPublic.ts:1832:1 | isBindingPattern | 0.040783 | 0.039978 | -0.000805 | 0.9803 |
| src/compiler/path.ts:91:1 | pathIsAbsolute | 0.036730 | 0.035923 | -0.000807 | 0.9780 |
| src/compiler/types.ts:8604:1 | SourceFileMayBeEmittedHost | 0.037164 | 0.036355 | -0.000809 | 0.9782 |
| src/compiler/moduleSpecifiers.ts:558:1 | isExcludedByRegex | 0.041827 | 0.041018 | -0.000809 | 0.9806 |
| src/compiler/_namespaces/ts.ts:67:1 |  | 0.036804 | 0.035992 | -0.000812 | 0.9779 |
| src/compiler/utilities.ts:5025:1 | walkUpParenthesizedTypesAndGetParentAndChild | 0.153891 | 0.153079 | -0.000812 | 0.9947 |
| src/compiler/factory/utilities.ts:1196:1 | isShiftOperator | 0.042926 | 0.042112 | -0.000814 | 0.9810 |
| src/compiler/types.ts:8139:1 | HasChangedAutomaticTypeDirectiveNames | 0.037325 | 0.036510 | -0.000815 | 0.9782 |
| src/compiler/types.ts:5886:1 | LazyNodeCheckFlags | 0.035861 | 0.035043 | -0.000818 | 0.9772 |
| src/compiler/scanner.ts:467:1 | getPositionOfLineAndCharacter | 0.114437 | 0.113613 | -0.000824 | 0.9928 |
| src/compiler/moduleNameResolver.ts:975:1 | RedirectsCacheKey | 0.038207 | 0.037377 | -0.000830 | 0.9783 |
| src/compiler/utilities.ts:982:1 | getSourceFileOfNode | 0.043476 | 0.042641 | -0.000834 | 0.9808 |
| src/compiler/utilities.ts:3189:1 | getThisContainer | 0.038821 | 0.037986 | -0.000835 | 0.9785 |
| src/compiler/utilities.ts:8534:1 | Identifier | 0.042978 | 0.042140 | -0.000838 | 0.9805 |
| src/compiler/commandLineParser.ts:3305:1 | startsWithConfigDirTemplate | 0.037065 | 0.036226 | -0.000839 | 0.9774 |
| src/compiler/factory/nodeTests.ts:1176:1 | isJSDocPropertyTag | 0.042227 | 0.041387 | -0.000839 | 0.9801 |
| src/compiler/types.ts:6841:1 | IterableOrIteratorType | 0.036183 | 0.035342 | -0.000841 | 0.9768 |
| src/compiler/utilities.ts:7180:1 | hasEffectiveModifiers | 0.043811 | 0.042969 | -0.000842 | 0.9808 |
| src/compiler/commandLineParser.ts:3396:1 | ExtendsResult | 0.048799 | 0.047954 | -0.000845 | 0.9827 |
| src/compiler/sourcemap.ts:444:1 | Mapping | 0.037145 | 0.036299 | -0.000846 | 0.9772 |
| src/compiler/utilities.ts:9584:1 | stripLeadingDirectorySeparator | 0.037975 | 0.037120 | -0.000855 | 0.9775 |
| src/compiler/builder.ts:1078:1 | IncrementalBuildInfoDiagnosticOfFile | 0.035608 | 0.034751 | -0.000857 | 0.9759 |
| src/compiler/types.ts:2751:1 | FunctionExpression | 0.036786 | 0.035928 | -0.000857 | 0.9767 |
| src/compiler/visitorPublic.ts:518:1 | visitFunctionBody | 0.035739 | 0.034880 | -0.000859 | 0.9760 |
| src/compiler/utilities.ts:2202:1 | isBlockScope | 0.222813 | 0.221945 | -0.000867 | 0.9961 |
| src/compiler/types.ts:2868:1 | LiteralToken | 0.036850 | 0.035982 | -0.000867 | 0.9765 |
| src/compiler/builder.ts:1074:1 | IncrementalBuildInfoFileId | 0.036726 | 0.035856 | -0.000870 | 0.9763 |
| src/compiler/scanner.ts:557:1 | isWhiteSpaceSingleLine | 0.037731 | 0.036860 | -0.000871 | 0.9769 |
| src/compiler/utilities.ts:1096:1 | isGrammarError | 0.046251 | 0.045375 | -0.000876 | 0.9811 |
| src/compiler/types.ts:3989:1 | JSDocUnknownTag | 0.036191 | 0.035310 | -0.000881 | 0.9756 |
| src/compiler/factory/utilities.ts:1494:1 | isExportOrDefaultModifier | 0.073685 | 0.072804 | -0.000881 | 0.9880 |
| src/compiler/utilities.ts:8275:1 | getClassLikeDeclarationOfSymbol | 0.039362 | 0.038479 | -0.000882 | 0.9776 |
| src/compiler/utilities.ts:4988:1 | ValueSignatureDeclaration | 0.036504 | 0.035611 | -0.000893 | 0.9755 |
| src/compiler/utilities.ts:7597:1 | isEmptyObjectLiteral | 0.160400 | 0.159500 | -0.000900 | 0.9944 |
| src/compiler/scanner.ts:613:1 | couldStartTrivia | 0.038590 | 0.037677 | -0.000913 | 0.9763 |
| src/compiler/types.ts:2550:1 | ShiftOperatorOrHigher | 0.039996 | 0.039082 | -0.000914 | 0.9771 |
| src/compiler/scanner.ts:592:1 | isDigit | 0.036969 | 0.036054 | -0.000915 | 0.9752 |
| src/compiler/utilities.ts:3541:1 | nodeOrChildIsDecorated | 0.066323 | 0.065405 | -0.000917 | 0.9862 |
| src/compiler/sourcemap.ts:365:1 |  | 0.038032 | 0.037104 | -0.000928 | 0.9756 |
| src/compiler/moduleSpecifiers.ts:1416:1 | tryGetRealFileNameForNonJsDeclarationFileName | 0.042862 | 0.041929 | -0.000933 | 0.9782 |
| src/compiler/types.ts:2265:1 | TypeLiteralNode | 0.036210 | 0.035270 | -0.000941 | 0.9740 |
| src/compiler/types.ts:6801:1 | EvolvingArrayType | 0.036092 | 0.035145 | -0.000946 | 0.9738 |
| src/compiler/factory/nodeTests.ts:1172:1 | isJSDocUnknownTag | 0.041799 | 0.040852 | -0.000947 | 0.9773 |
| src/compiler/utilities.ts:4460:1 | hasQuestionToken | 0.042151 | 0.041203 | -0.000948 | 0.9775 |
| src/compiler/moduleNameResolver.ts:216:1 | resolvedTypeScriptOnly | 0.040260 | 0.039312 | -0.000948 | 0.9765 |
| src/compiler/_namespaces/ts.ts:68:1 |  | 0.036076 | 0.035128 | -0.000949 | 0.9737 |
| src/compiler/commandLineParser.ts:2976:1 | convertToOptionsWithAbsolutePaths | 0.038641 | 0.037692 | -0.000949 | 0.9754 |
| src/compiler/utilitiesPublic.ts:1985:1 | isCallOrNewExpression | 0.041520 | 0.040571 | -0.000949 | 0.9771 |
| src/compiler/types.ts:2646:1 | BinaryOperatorToken | 0.038591 | 0.037638 | -0.000953 | 0.9753 |
| src/compiler/path.ts:751:1 | getPathWithoutRoot | 0.035840 | 0.034886 | -0.000954 | 0.9734 |
| src/compiler/types.ts:3885:1 | JSDocTypeExpression | 0.035275 | 0.034320 | -0.000955 | 0.9729 |
| src/compiler/builder.ts:1176:1 | isIncrementalBundleEmitBuildInfo | 0.037255 | 0.036300 | -0.000955 | 0.9744 |
| src/compiler/types.ts:1643:1 | ReadonlyKeyword | 0.036448 | 0.035492 | -0.000956 | 0.9738 |
| src/compiler/path.ts:757:1 | getNormalizedAbsolutePathWithoutRoot | 0.035183 | 0.034224 | -0.000960 | 0.9727 |
| src/compiler/transformers/generators.ts:276:1 | LabeledBlock | 0.037433 | 0.036471 | -0.000962 | 0.9743 |
| src/compiler/_namespaces/ts.ts:66:1 |  | 0.035885 | 0.034915 | -0.000970 | 0.9730 |
| src/compiler/path.ts:494:1 | getPathComponents | 0.036230 | 0.035259 | -0.000971 | 0.9732 |
| src/compiler/utilitiesPublic.ts:1849:1 | isArrayBindingElement | 0.040751 | 0.039777 | -0.000973 | 0.9761 |
| src/compiler/parser.ts:9929:1 | markAsIncrementallyParsed | 0.036390 | 0.035417 | -0.000973 | 0.9733 |
| src/compiler/types.ts:2748:1 | FunctionBody | 0.037265 | 0.036284 | -0.000980 | 0.9737 |
| src/compiler/parser.ts:10712:1 |  | 0.036799 | 0.035818 | -0.000981 | 0.9733 |
| src/compiler/_namespaces/ts.ts:33:1 |  | 0.040516 | 0.039533 | -0.000984 | 0.9757 |
| src/compiler/watch.ts:904:1 | CreateWatchCompilerHostInput | 0.038128 | 0.037140 | -0.000988 | 0.9741 |
| src/compiler/factory/nodeTests.ts:580:1 | isObjectLiteralExpression | 0.042688 | 0.041698 | -0.000990 | 0.9768 |
| src/compiler/utilities.ts:5195:1 | getEffectiveBaseTypeNode | 0.051283 | 0.050287 | -0.000996 | 0.9806 |
| src/compiler/utilities.ts:3809:1 | isInJsonFile | 0.039102 | 0.038103 | -0.000999 | 0.9745 |
| src/compiler/types.ts:7126:1 | InferenceInfo | 0.036565 | 0.035565 | -0.000999 | 0.9727 |
| src/compiler/factory/emitHelpers.ts:1009:1 |  | 0.064264 | 0.063254 | -0.001010 | 0.9843 |
| src/compiler/watchUtilities.ts:110:1 | SortedAndCanonicalizedMutableFileSystemEntries | 0.036845 | 0.035835 | -0.001010 | 0.9726 |
| src/compiler/factory/utilities.ts:1565:1 | formatIdentifier | 0.041381 | 0.040369 | -0.001011 | 0.9756 |
| src/compiler/types.ts:3567:1 | ClassElement | 0.036906 | 0.035892 | -0.001013 | 0.9725 |
| src/compiler/utilities.ts:1341:1 | getPos | 0.041031 | 0.040016 | -0.001015 | 0.9753 |
| src/compiler/types.ts:3572:1 | TypeElement | 0.036472 | 0.035456 | -0.001016 | 0.9722 |
| src/compiler/utilities.ts:8280:1 | getObjectFlags | 0.038612 | 0.037593 | -0.001019 | 0.9736 |
| src/compiler/types.ts:10343:1 | PragmaArgumentType | 0.037681 | 0.036657 | -0.001024 | 0.9728 |
| src/compiler/factory/nodeTests.ts:1136:1 | isJSDocDeprecatedTag | 0.044102 | 0.043077 | -0.001025 | 0.9768 |
| src/compiler/_namespaces/ts.ts:41:1 |  | 0.037204 | 0.036178 | -0.001025 | 0.9724 |
| src/compiler/utilities.ts:7185:1 | hasSyntacticModifiers | 0.045125 | 0.044098 | -0.001027 | 0.9772 |
| src/compiler/utilities.ts:4848:1 | AssignmentTarget | 0.036055 | 0.035023 | -0.001031 | 0.9714 |
| src/compiler/types.ts:3643:1 | NamespaceDeclaration | 0.036933 | 0.035902 | -0.001032 | 0.9721 |
| src/compiler/scanner.ts:412:1 | tokenToString | 0.044729 | 0.043696 | -0.001034 | 0.9769 |
| src/compiler/utilitiesPublic.ts:1532:1 | isImportOrExportSpecifier | 0.289063 | 0.288008 | -0.001055 | 0.9963 |
| src/compiler/utilities.ts:8555:1 |  | 0.065263 | 0.064207 | -0.001056 | 0.9838 |
| src/compiler/scanner.ts:605:1 | isWordCharacter | 0.038021 | 0.036961 | -0.001061 | 0.9721 |
| src/compiler/utilities.ts:9992:1 | getSupportedExtensions | 0.037880 | 0.036818 | -0.001062 | 0.9720 |
| src/compiler/types.ts:3289:1 | JsxSpreadAttribute | 0.037166 | 0.036103 | -0.001062 | 0.9714 |
| src/compiler/utilities.ts:7988:1 | startEndContainsRange | 0.037058 | 0.035990 | -0.001068 | 0.9712 |
| src/compiler/utilities.ts:5475:1 | AnonymousFunctionDefinition | 0.037299 | 0.036230 | -0.001069 | 0.9713 |
| src/compiler/transformers/utilities.ts:682:1 | getAllDecoratorsOfClass | 0.039030 | 0.037960 | -0.001070 | 0.9726 |
| src/compiler/types.ts:5738:1 | AnyImportOrBareOrAccessedRequire | 0.035757 | 0.034682 | -0.001075 | 0.9699 |
| src/compiler/utilitiesPublic.ts:925:1 | getDeclarationIdentifier | 0.186042 | 0.184961 | -0.001080 | 0.9942 |
| src/compiler/expressionToTypeNode.ts:166:1 | syntacticResult | 0.042858 | 0.041777 | -0.001081 | 0.9748 |
| src/compiler/commandLineParser.ts:3359:1 | canJsonReportNoInputFiles | 0.038668 | 0.037584 | -0.001083 | 0.9720 |
| src/compiler/scanner.ts:966:1 | getShebang | 0.036865 | 0.035777 | -0.001088 | 0.9705 |
| src/compiler/sourcemap.ts:654:1 | MappedPosition | 0.036652 | 0.035563 | -0.001088 | 0.9703 |
| src/compiler/types.ts:1636:1 | DeclareKeyword | 0.035858 | 0.034769 | -0.001089 | 0.9696 |
| src/compiler/types.ts:7197:1 | AssignmentDeclarationKind | 0.036179 | 0.035087 | -0.001092 | 0.9698 |
| src/compiler/utilities.ts:2332:1 | getEnclosingBlockScopeContainer | 0.062089 | 0.060989 | -0.001101 | 0.9823 |
| src/compiler/types.ts:1788:1 | LateBoundElementAccessExpression | 0.037336 | 0.036233 | -0.001103 | 0.9705 |
| src/compiler/utilities.ts:7435:1 | isAssignmentOperator | 0.052199 | 0.051090 | -0.001108 | 0.9788 |
| src/compiler/types.ts:5036:1 | EmitResult | 0.036322 | 0.035204 | -0.001118 | 0.9692 |
| src/compiler/moduleNameResolver.ts:1271:1 | ModuleOrTypeReferenceResolutionCache | 0.036664 | 0.035544 | -0.001120 | 0.9695 |
| src/compiler/transformers/destructuring.ts:615:1 | makeObjectAssignmentPattern | 0.041429 | 0.040309 | -0.001120 | 0.9730 |
| src/compiler/utilities.ts:5507:1 | NamedEvaluationSource | 0.036242 | 0.035119 | -0.001123 | 0.9690 |
| src/compiler/utilities.ts:10268:1 | positionIsSynthesized | 0.040933 | 0.039809 | -0.001123 | 0.9726 |
| src/compiler/factory/emitHelpers.ts:802:1 |  | 0.062555 | 0.061432 | -0.001124 | 0.9820 |
| src/compiler/types.ts:3940:1 | JSDocNamepathType | 0.035503 | 0.034380 | -0.001124 | 0.9683 |
| src/compiler/scanner.ts:356:1 |  | 0.037792 | 0.036667 | -0.001125 | 0.9702 |
| src/compiler/types.ts:7156:1 | Ternary | 0.036143 | 0.035013 | -0.001130 | 0.9687 |
| src/compiler/commandLineParser.ts:2306:1 | commandLineOptionsToMap | 0.039865 | 0.038735 | -0.001130 | 0.9716 |
| src/compiler/tsbuildPublic.ts:259:1 | isCircularBuildOrder | 0.038926 | 0.037793 | -0.001132 | 0.9709 |
| src/compiler/transformers/generators.ts:254:1 | ExceptionBlockState | 0.037744 | 0.036605 | -0.001139 | 0.9698 |
| src/compiler/factory/utilities.ts:1159:1 | isReadonlyKeywordOrPlusOrMinusToken | 0.273690 | 0.272551 | -0.001140 | 0.9958 |
| src/compiler/types.ts:2298:1 | UnionOrIntersectionTypeNode | 0.035692 | 0.034548 | -0.001144 | 0.9680 |
| src/compiler/types.ts:3611:1 | EnumDeclaration | 0.036859 | 0.035713 | -0.001147 | 0.9689 |
| src/compiler/moduleNameResolver.ts:3227:1 | getPackageNameFromTypesPackageName | 0.038689 | 0.037542 | -0.001147 | 0.9703 |
| src/compiler/transformers/es2015.ts:472:1 | SpreadSegmentKind | 0.037897 | 0.036750 | -0.001147 | 0.9697 |
| src/compiler/utilities.ts:8590:1 |  | 0.037596 | 0.036447 | -0.001149 | 0.9694 |
| src/compiler/factory/nodeTests.ts:664:1 | isClassExpression | 0.041150 | 0.039999 | -0.001151 | 0.9720 |
| src/compiler/utilities.ts:10206:1 | tryParsePattern | 0.038473 | 0.037318 | -0.001155 | 0.9700 |
| src/compiler/types.ts:3906:1 | JSDocAllType | 0.036788 | 0.035628 | -0.001159 | 0.9685 |
| src/compiler/transformers/typeSerializer.ts:79:1 | SerializedTypeNode | 0.037251 | 0.036090 | -0.001160 | 0.9688 |
| src/compiler/factory/nodeTests.ts:1018:1 | isBundle | 0.041101 | 0.039937 | -0.001164 | 0.9717 |
| src/compiler/utilities.ts:6803:1 | isInTypeQuery | 0.054168 | 0.053004 | -0.001165 | 0.9785 |
| src/compiler/utilitiesPublic.ts:682:1 | getNodeFlags | 0.040161 | 0.038995 | -0.001166 | 0.9710 |
| src/compiler/types.ts:2794:1 | RegularExpressionFlags | 0.037778 | 0.036607 | -0.001170 | 0.9690 |
| src/compiler/utilities.ts:6270:1 | isQuoteOrBacktick | 0.039401 | 0.038230 | -0.001172 | 0.9703 |
| src/compiler/utilities.ts:7804:1 | readJson | 0.036968 | 0.035792 | -0.001176 | 0.9682 |
| src/compiler/types.ts:7189:1 | WideningContext | 0.036571 | 0.035394 | -0.001177 | 0.9678 |
| src/compiler/tsbuildPublic.ts:268:1 | SolutionBuilder | 0.036244 | 0.035065 | -0.001179 | 0.9675 |
| src/compiler/types.ts:6767:1 | IntersectionType | 0.038207 | 0.037025 | -0.001182 | 0.9691 |
| src/compiler/checker.ts:1479:1 | isInstantiatedModule | 0.037665 | 0.036482 | -0.001183 | 0.9686 |
| src/compiler/utilities.ts:5631:1 | Associativity | 0.036668 | 0.035484 | -0.001184 | 0.9677 |
| src/compiler/sourcemap.ts:461:1 | decodeMappings | 0.036843 | 0.035658 | -0.001185 | 0.9678 |
| src/compiler/types.ts:3578:1 | InterfaceDeclaration | 0.037089 | 0.035904 | -0.001185 | 0.9681 |
| src/compiler/types.ts:3622:1 | ModuleBody | 0.037257 | 0.036072 | -0.001185 | 0.9682 |
| src/compiler/tsbuildPublic.ts:359:1 | FileWatcherWithModifiedTime | 0.037324 | 0.036134 | -0.001190 | 0.9681 |
| src/compiler/sourcemap.ts:437:1 | MappingsDecoder | 0.037801 | 0.036610 | -0.001191 | 0.9685 |
| src/compiler/tsbuildPublic.ts:169:1 | BuildResultFlags | 0.036138 | 0.034941 | -0.001198 | 0.9669 |
| src/compiler/types.ts:1783:1 | LateBoundBinaryExpressionDeclaration | 0.037257 | 0.036059 | -0.001198 | 0.9678 |
| src/compiler/utilities.ts:7525:1 | isDottedName | 0.125707 | 0.124505 | -0.001202 | 0.9904 |
| src/compiler/path.ts:762:1 | toPath | 0.035535 | 0.034330 | -0.001206 | 0.9661 |
| src/compiler/sourcemap.ts:454:1 | SourceMapping | 0.036436 | 0.035229 | -0.001206 | 0.9669 |
| src/compiler/builder.ts:751:1 | getNextAffectedFilePendingEmit | 0.038441 | 0.037234 | -0.001208 | 0.9686 |
| src/compiler/utilities.ts:7312:1 | selectSyntacticModifierFlags | 0.037608 | 0.036399 | -0.001209 | 0.9679 |
| src/compiler/path.ts:100:1 | pathIsRelative | 0.037245 | 0.036036 | -0.001209 | 0.9675 |
| src/compiler/types.ts:3594:1 | TypeAliasDeclaration | 0.036588 | 0.035379 | -0.001210 | 0.9669 |
| src/compiler/path.ts:882:1 |  | 0.035662 | 0.034450 | -0.001212 | 0.9660 |
| src/compiler/factory/emitHelpers.ts:53:1 | ESDecorateClassContext | 0.038088 | 0.036868 | -0.001220 | 0.9680 |
| src/compiler/path.ts:435:1 | getAnyExtensionFromPath | 0.037372 | 0.036146 | -0.001226 | 0.9672 |
| src/compiler/program.ts:1195:1 | getReferencedFileLocation | 0.042936 | 0.041709 | -0.001227 | 0.9714 |
| src/compiler/types.ts:5712:1 | AssertsThisTypePredicate | 0.036286 | 0.035057 | -0.001228 | 0.9661 |
| src/compiler/types.ts:6833:1 | IterationTypes | 0.036384 | 0.035154 | -0.001230 | 0.9662 |
| src/compiler/types.ts:5823:1 | AllAccessorDeclarations | 0.036466 | 0.035231 | -0.001235 | 0.9661 |
| src/compiler/sourcemap.ts:820:1 |  | 0.068052 | 0.066815 | -0.001237 | 0.9818 |
| src/compiler/types.ts:9874:1 | PrinterOptions | 0.037366 | 0.036126 | -0.001239 | 0.9668 |
| src/compiler/factory/nodeTests.ts:524:1 | isParenthesizedTypeNode | 0.041723 | 0.040483 | -0.001240 | 0.9703 |
| src/compiler/types.ts:6875:1 | TypeParameter | 0.035819 | 0.034570 | -0.001248 | 0.9651 |
| src/compiler/resolutionCache.ts:249:1 | DirectoryOfFailedLookupWatch | 0.037576 | 0.036316 | -0.001260 | 0.9665 |
| src/compiler/utilitiesPublic.ts:1860:1 | isDeclarationBindingElement | 0.037130 | 0.035865 | -0.001265 | 0.9659 |
| src/compiler/types.ts:3321:1 | Statement | 0.037218 | 0.035954 | -0.001265 | 0.9660 |
| src/compiler/utilitiesPublic.ts:1165:1 | getJSDocOverrideTagNoCache | 0.052133 | 0.050868 | -0.001266 | 0.9757 |
| src/compiler/diagnosticInformationMap.generated.ts:4:1 |  | 0.036518 | 0.035252 | -0.001266 | 0.9653 |
| src/compiler/utilities.ts:5597:1 | isPartOfParameterDeclaration | 0.046732 | 0.045463 | -0.001269 | 0.9728 |
| src/compiler/core.ts:1835:1 | toLowerCase | 0.044751 | 0.043482 | -0.001269 | 0.9716 |
| src/compiler/path.ts:433:1 | getAnyExtensionFromPath | 0.037353 | 0.036082 | -0.001270 | 0.9660 |
| src/compiler/moduleNameResolver.ts:479:1 | getEffectiveTypeRoots | 0.036613 | 0.035338 | -0.001274 | 0.9652 |
| src/compiler/types.ts:6477:1 | NullableType | 0.038359 | 0.037082 | -0.001277 | 0.9667 |
| src/compiler/types.ts:2288:1 | OptionalTypeNode | 0.035833 | 0.034554 | -0.001279 | 0.9643 |
| src/compiler/types.ts:6774:1 | StructuredType | 0.036514 | 0.035228 | -0.001286 | 0.9648 |
| src/compiler/path.ts:925:1 | comparePathsCaseSensitive | 0.038059 | 0.036773 | -0.001287 | 0.9662 |
| src/compiler/transformers/classFields.ts:342:1 | LexicalEnv | 0.037827 | 0.036537 | -0.001289 | 0.9659 |
| src/compiler/path.ts:538:1 | reducePathComponents | 0.068875 | 0.067576 | -0.001299 | 0.9811 |
| src/compiler/commandLineParser.ts:2176:1 | parseBuildCommand | 0.189777 | 0.188478 | -0.001300 | 0.9932 |
| src/compiler/moduleNameResolver.ts:2573:1 | allKeysStartWithDot | 0.039427 | 0.038117 | -0.001309 | 0.9668 |
| src/compiler/commandLineParser.ts:3037:1 | setConfigFileInOptions | 0.041140 | 0.039828 | -0.001311 | 0.9681 |
| src/compiler/path.ts:449:1 | pathComponents | 0.036910 | 0.035597 | -0.001312 | 0.9644 |
| src/compiler/_namespaces/ts.ts:72:1 |  | 0.036629 | 0.035314 | -0.001315 | 0.9641 |
| src/compiler/factory/nodeTests.ts:921:1 | isExternalModuleReference | 0.042399 | 0.041081 | -0.001319 | 0.9689 |
| src/compiler/utilitiesPublic.ts:2468:1 | isJsxChild | 0.040412 | 0.039092 | -0.001320 | 0.9673 |
| src/compiler/watchUtilities.ts:463:1 | cleanExtendedConfigCache | 0.077646 | 0.076320 | -0.001326 | 0.9829 |
| src/compiler/utilities.ts:10049:1 | ModuleSpecifierEnding | 0.037435 | 0.036108 | -0.001326 | 0.9646 |
| src/compiler/parser.ts:1437:1 | Parser | 0.046389 | 0.045060 | -0.001329 | 0.9714 |
| src/compiler/utilitiesPublic.ts:1635:1 | isParameterPropertyModifier | 0.041182 | 0.039851 | -0.001331 | 0.9677 |
| src/compiler/parser.ts:10561:1 | getDeclarationFileExtension | 0.036996 | 0.035659 | -0.001337 | 0.9639 |
| src/compiler/utilities.ts:4795:1 | getEffectiveJSDocHost | 0.044682 | 0.043341 | -0.001341 | 0.9700 |
| src/compiler/types.ts:2228:1 | FunctionTypeNode | 0.037110 | 0.035767 | -0.001343 | 0.9638 |
| src/compiler/factory/nodeTests.ts:832:1 | isNamespaceExportDeclaration | 0.044089 | 0.042746 | -0.001343 | 0.9695 |
| src/compiler/transformers/esDecorators.ts:227:1 | MemberInfo | 0.035959 | 0.034615 | -0.001344 | 0.9626 |
| src/compiler/transformers/esDecorators.ts:281:1 | OtherLexicalEnvironmentStackEntry | 0.035854 | 0.034505 | -0.001349 | 0.9624 |
| src/compiler/types.ts:10044:1 | ModuleSpecifierCache | 0.038017 | 0.036666 | -0.001351 | 0.9645 |
| src/compiler/transformers/utilities.ts:598:1 | getStaticPropertiesAndClassStaticBlock | 0.037884 | 0.036528 | -0.001356 | 0.9642 |
| src/compiler/types.ts:1701:1 | Identifier | 0.037548 | 0.036189 | -0.001358 | 0.9638 |
| src/compiler/types.ts:2464:1 | TrueLiteral | 0.036256 | 0.034896 | -0.001359 | 0.9625 |
| src/compiler/transformers/classFields.ts:263:1 | PrivateIdentifierAccessorInfo | 0.041858 | 0.040492 | -0.001366 | 0.9674 |
| src/compiler/builder.ts:1084:1 | IncrementalBuildInfoEmitDiagnostic | 0.037235 | 0.035856 | -0.001379 | 0.9630 |
| src/compiler/utilities.ts:2648:1 | isEnumConst | 0.040324 | 0.038942 | -0.001381 | 0.9657 |
| src/compiler/types.ts:6859:1 | SyntheticDefaultModuleType | 0.035917 | 0.034535 | -0.001382 | 0.9615 |
| src/compiler/utilities.ts:5297:1 | isIdentifierANonContextualKeyword | 0.040601 | 0.039215 | -0.001386 | 0.9659 |
| src/compiler/utilities.ts:10035:1 | hasTSFileExtension | 0.040699 | 0.039310 | -0.001389 | 0.9659 |
| src/compiler/types.ts:3914:1 | JSDocNonNullableType | 0.037186 | 0.035795 | -0.001391 | 0.9626 |
| src/compiler/moduleNameResolver.ts:2421:1 | readPackageJsonPeerDependencies | 0.039346 | 0.037953 | -0.001393 | 0.9646 |
| src/compiler/transformers/taggedTemplate.ts:35:1 | processTaggedTemplateExpression | 0.046485 | 0.045091 | -0.001394 | 0.9700 |
| src/compiler/factory/utilities.ts:280:1 | createExpressionForJsxFragment | 0.124106 | 0.122710 | -0.001396 | 0.9888 |
| src/compiler/_namespaces/ts.ts:75:1 |  | 0.036399 | 0.035000 | -0.001398 | 0.9616 |
| src/compiler/types.ts:1306:1 | HasInitializer | 0.043648 | 0.042245 | -0.001403 | 0.9679 |
| src/compiler/utilitiesPublic.ts:2039:1 | isUnaryExpression | 0.042248 | 0.040843 | -0.001405 | 0.9667 |
| src/compiler/utilities.ts:7749:1 | base64decode | 0.109290 | 0.107885 | -0.001405 | 0.9871 |
| src/compiler/types.ts:1906:1 | BindingElement | 0.036237 | 0.034829 | -0.001408 | 0.9611 |
| src/compiler/program.ts:657:1 | formatDiagnostics | 0.037290 | 0.035881 | -0.001409 | 0.9622 |
| src/compiler/factory/nodeFactory.ts:7295:1 | propagatePropertyNameFlagsOfChild | 0.039368 | 0.037959 | -0.001409 | 0.9642 |
| src/compiler/program.ts:813:1 | flattenDiagnosticMessageText | 0.067133 | 0.065722 | -0.001410 | 0.9790 |
| src/compiler/builder.ts:1093:1 | IncrementalBuildInfoFilePendingEmit | 0.037302 | 0.035884 | -0.001418 | 0.9620 |
| src/compiler/utilities.ts:12298:1 | getPropertyArrayElementValue | 0.059358 | 0.057938 | -0.001420 | 0.9761 |
| src/compiler/utilities.ts:5603:1 | getRootDeclaration | 0.043764 | 0.042343 | -0.001421 | 0.9675 |
| src/compiler/utilities.ts:8266:1 | isAbstractConstructorSymbol | 0.043141 | 0.041718 | -0.001423 | 0.9670 |
| src/compiler/utilities.ts:6827:1 | identifierIsThisKeyword | 0.037654 | 0.036230 | -0.001424 | 0.9622 |
| src/compiler/sys.ts:75:1 | setStackTraceLimit | 0.036976 | 0.035547 | -0.001430 | 0.9613 |
| src/compiler/utilities.ts:5303:1 | isTrivia | 0.040372 | 0.038942 | -0.001430 | 0.9646 |
| src/compiler/types.ts:3558:1 | ClassExpression | 0.037777 | 0.036346 | -0.001431 | 0.9621 |
| src/compiler/factory/utilities.ts:516:1 | expandPreOrPostfixIncrementOrDecrementExpression | 0.083885 | 0.082450 | -0.001435 | 0.9829 |
| src/compiler/utilitiesPublic.ts:2360:1 | isDeclarationStatementKind | 0.040252 | 0.038811 | -0.001441 | 0.9642 |
| src/compiler/utilitiesPublic.ts:397:1 | decodedTextSpanIntersectsWith | 0.035294 | 0.033852 | -0.001442 | 0.9591 |
| src/compiler/types.ts:2591:1 | BitwiseOperatorOrHigher | 0.037317 | 0.035874 | -0.001443 | 0.9613 |
| src/compiler/emitter.ts:636:1 | getCommonSourceDirectory | 0.037522 | 0.036074 | -0.001448 | 0.9614 |
| src/compiler/transformers/taggedTemplate.ts:1:1 |  | 0.038324 | 0.036873 | -0.001451 | 0.9621 |
| src/compiler/factory/nodeTests.ts:1199:1 | isSyntaxList | 0.043200 | 0.041746 | -0.001454 | 0.9663 |
| src/compiler/types.ts:3663:1 | ModuleReference | 0.036470 | 0.035013 | -0.001457 | 0.9601 |
| src/compiler/sys.ts:245:1 | createDynamicPriorityPollingWatchFile | 0.037104 | 0.035647 | -0.001457 | 0.9607 |
| src/compiler/types.ts:1916:1 | BindingElementGrandparent | 0.036616 | 0.035159 | -0.001457 | 0.9602 |
| src/compiler/types.ts:9711:1 | Visitor | 0.037928 | 0.036465 | -0.001463 | 0.9614 |
| src/compiler/utilities.ts:8040:1 | getCheckFlags | 0.130264 | 0.128800 | -0.001465 | 0.9888 |
| src/compiler/utilities.ts:3529:1 | nodeIsDecorated | 0.050536 | 0.049068 | -0.001469 | 0.9709 |
| src/compiler/utilities.ts:2354:1 | getNameFromIndexInfo | 0.039542 | 0.038067 | -0.001475 | 0.9627 |
| src/compiler/utilitiesPublic.ts:1067:1 | getJSDocParameterTags | 0.043021 | 0.041543 | -0.001477 | 0.9657 |
| src/compiler/scanner.ts:300:1 |  | 0.039143 | 0.037659 | -0.001483 | 0.9621 |
| src/compiler/utilitiesPublic.ts:469:1 | createTextChangeRange | 0.037025 | 0.035537 | -0.001488 | 0.9598 |
| src/compiler/utilities.ts:2653:1 | isDeclarationReadonly | 0.048836 | 0.047346 | -0.001490 | 0.9695 |
| src/compiler/types.ts:2775:1 | TemplateLiteralLikeNode | 0.036716 | 0.035225 | -0.001491 | 0.9594 |
| src/compiler/checker.ts:54329:1 | SymbolTrackerImpl | 0.039144 | 0.037649 | -0.001495 | 0.9618 |
| src/compiler/types.ts:1672:1 | ParameterPropertyModifier | 0.036758 | 0.035260 | -0.001498 | 0.9592 |
| src/compiler/types.ts:5679:1 | SymbolAccessibility | 0.037139 | 0.035640 | -0.001499 | 0.9596 |
| src/compiler/commandLineParser.ts:1788:1 |  | 0.037844 | 0.036344 | -0.001499 | 0.9604 |
| src/compiler/types.ts:6852:1 | PromiseOrAwaitableType | 0.035748 | 0.034241 | -0.001506 | 0.9579 |
| src/compiler/resolutionCache.ts:320:1 | canWatchDirectoryOrFilePath | 0.038657 | 0.037149 | -0.001508 | 0.9610 |
| src/compiler/types.ts:4015:1 | JSDocClassTag | 0.035926 | 0.034413 | -0.001512 | 0.9579 |
| src/compiler/utilities.ts:9306:1 |  | 0.073663 | 0.072145 | -0.001518 | 0.9794 |
| src/compiler/sys.ts:382:1 | createUseFsEventsOnParentDirectoryWatchFile | 0.176321 | 0.174802 | -0.001519 | 0.9914 |
| src/compiler/utilities.ts:5255:1 | getAncestor | 0.045880 | 0.044361 | -0.001519 | 0.9669 |
| src/compiler/builder.ts:700:1 | clearAffectedFilesPendingEmit | 0.038330 | 0.036807 | -0.001522 | 0.9603 |
| src/compiler/utilities.ts:2148:1 | getNonAugmentationDeclaration | 0.038974 | 0.037451 | -0.001523 | 0.9609 |
| src/compiler/utilitiesPublic.ts:1126:1 | getJSDocPublicTag | 0.052703 | 0.051177 | -0.001526 | 0.9711 |
| src/compiler/transformers/declarations.ts:1873:1 | maskModifiers | 0.084879 | 0.083353 | -0.001526 | 0.9820 |
| src/compiler/_namespaces/ts.ts:64:1 |  | 0.037420 | 0.035884 | -0.001536 | 0.9590 |
| src/compiler/factory/emitHelpers.ts:730:1 |  | 0.064918 | 0.063382 | -0.001536 | 0.9763 |
| src/compiler/types.ts:3872:1 | CommentRange | 0.035830 | 0.034292 | -0.001538 | 0.9571 |
| src/compiler/utilitiesPublic.ts:327:1 | getDefaultLibFileName | 0.036836 | 0.035293 | -0.001543 | 0.9581 |
| src/compiler/program.ts:634:1 | getPreEmitDiagnostics | 0.038852 | 0.037305 | -0.001547 | 0.9602 |
| src/compiler/factory/utilities.ts:1611:1 | createAccessorPropertyBackingField | 0.041974 | 0.040421 | -0.001553 | 0.9630 |
| src/compiler/moduleNameResolver.ts:1051:1 | createPackageJsonInfoCache | 0.062305 | 0.060750 | -0.001555 | 0.9750 |
| src/compiler/transformers/declarations/diagnostics.ts:104:1 | SymbolAccessibilityDiagnostic | 0.038988 | 0.037433 | -0.001556 | 0.9601 |
| src/compiler/transformers/generators.ts:284:1 | SwitchBlock | 0.037927 | 0.036371 | -0.001557 | 0.9590 |
| src/compiler/utilities.ts:2166:1 | isEffectiveStrictModeSourceFile | 0.041539 | 0.039982 | -0.001557 | 0.9625 |
| src/compiler/path.ts:870:1 | changeFullExtension | 0.036330 | 0.034772 | -0.001558 | 0.9571 |
| src/compiler/utilities.ts:3402:1 | isSuperProperty | 0.042904 | 0.041339 | -0.001565 | 0.9635 |
| src/compiler/_namespaces/ts.ts:71:1 |  | 0.036334 | 0.034766 | -0.001567 | 0.9569 |
| src/compiler/factory/nodeTests.ts:1108:1 | isJSDocCallbackTag | 0.041554 | 0.039969 | -0.001585 | 0.9619 |
| src/compiler/factory/nodeTests.ts:532:1 | isTypeOperatorNode | 0.041607 | 0.040020 | -0.001587 | 0.9619 |
| src/compiler/utilities.ts:7505:1 | getFirstIdentifier | 0.038900 | 0.037306 | -0.001594 | 0.9590 |
| src/compiler/emitter.ts:6366:1 | emitListItemWithParenthesizerRuleSelector | 0.078095 | 0.076500 | -0.001594 | 0.9796 |
| src/compiler/path.ts:1063:1 | getRelativePathFromFile | 0.037570 | 0.035971 | -0.001598 | 0.9575 |
| src/compiler/utilitiesPublic.ts:2483:1 | isStringLiteralOrJsxExpression | 0.041563 | 0.039965 | -0.001599 | 0.9615 |
| src/compiler/commandLineParser.ts:274:1 |  | 0.038110 | 0.036511 | -0.001599 | 0.9581 |
| src/compiler/commandLineParser.ts:2370:1 | getTsconfigRootOptionsMap | 0.039929 | 0.038328 | -0.001601 | 0.9599 |
| src/compiler/types.ts:10347:1 | ConcretePragmaSpecs | 0.038317 | 0.036715 | -0.001602 | 0.9582 |
| src/compiler/types.ts:6864:1 | InstantiableType | 0.036031 | 0.034428 | -0.001602 | 0.9555 |
| src/compiler/factory/nodeTests.ts:592:1 | isCallExpression | 0.041946 | 0.040343 | -0.001603 | 0.9618 |
| src/compiler/path.ts:312:1 | getDirectoryPath | 0.037380 | 0.035775 | -0.001605 | 0.9570 |
| src/compiler/watch.ts:827:1 | setGetSourceFileAsHashVersioned | 0.076865 | 0.075246 | -0.001620 | 0.9789 |
| src/compiler/types.ts:4011:1 | JSDocDeprecatedTag | 0.036522 | 0.034900 | -0.001622 | 0.9556 |
| src/compiler/utilitiesPublic.ts:941:1 | getNameOfJSDocTypedef | 0.039345 | 0.037716 | -0.001629 | 0.9586 |
| src/compiler/utilitiesPublic.ts:2629:1 | isJSDocLinkLike | 0.043733 | 0.042100 | -0.001632 | 0.9627 |
| src/compiler/factory/utilities.ts:1234:1 | isBitwiseOperator | 0.040675 | 0.039039 | -0.001636 | 0.9598 |
| src/compiler/watchUtilities.ts:557:1 | IsIgnoredFileFromWildCardWatchingInput | 0.037207 | 0.035567 | -0.001639 | 0.9559 |
| src/compiler/tsbuildPublic.ts:367:1 | BuildInfoCacheEntry | 0.036614 | 0.034972 | -0.001642 | 0.9551 |
| src/compiler/utilitiesPublic.ts:2415:1 | isStatementButNotDeclaration | 0.043966 | 0.042322 | -0.001644 | 0.9626 |
| src/compiler/utilities.ts:5997:1 | getBinaryOperatorPrecedence | 0.041036 | 0.039391 | -0.001645 | 0.9599 |
| src/compiler/path.ts:627:1 | getNormalizedAbsolutePath | 0.036588 | 0.034942 | -0.001645 | 0.9550 |
| src/compiler/types.ts:7097:1 | TypeMapper | 0.036340 | 0.034690 | -0.001650 | 0.9546 |
| src/compiler/executeCommandLine.ts:184:1 | printVersion | 0.073051 | 0.071400 | -0.001651 | 0.9774 |
| src/compiler/factory/utilities.ts:237:1 | createJsxFragmentFactoryExpression | 0.080345 | 0.078691 | -0.001654 | 0.9794 |
| src/compiler/utilities.ts:720:1 | changesAffectingProgramStructure | 0.039392 | 0.037728 | -0.001665 | 0.9577 |
| src/compiler/factory/nodeTests.ts:540:1 | isMappedTypeNode | 0.041881 | 0.040216 | -0.001665 | 0.9602 |
| src/compiler/scanner.ts:792:1 |  | 0.037013 | 0.035340 | -0.001673 | 0.9548 |
| src/compiler/builder.ts:1052:1 | getBinderAndCheckerDiagnosticsOfFile | 0.076379 | 0.074704 | -0.001675 | 0.9781 |
| src/compiler/expressionToTypeNode.ts:170:1 |  | 0.043310 | 0.041628 | -0.001682 | 0.9612 |
| src/compiler/types.ts:4980:1 | CustomTransformers | 0.037223 | 0.035536 | -0.001687 | 0.9547 |
| src/compiler/factory/nodeFactory.ts:7286:1 | propagateNameFlags | 0.180451 | 0.178764 | -0.001687 | 0.9906 |
| src/compiler/commandLineParser.ts:3392:1 | isSuccessfulParsedTsconfig | 0.041490 | 0.039802 | -0.001688 | 0.9593 |
| src/compiler/utilities.ts:9345:1 | moduleResolutionSupportsPackageJsonExportsAndImports | 0.037588 | 0.035896 | -0.001693 | 0.9550 |
| src/compiler/program.ts:985:1 | getResolutionModeOverride | 0.039149 | 0.037451 | -0.001699 | 0.9566 |
| src/compiler/utilitiesPublic.ts:693:1 | validateLocaleAndSetLanguage | 0.182033 | 0.180331 | -0.001702 | 0.9907 |
| src/compiler/types.ts:2769:1 | LiteralLikeNode | 0.036899 | 0.035197 | -0.001702 | 0.9539 |
| src/compiler/factory/utilities.ts:1191:1 | isAdditiveOperatorOrHigher | 0.107722 | 0.106020 | -0.001702 | 0.9842 |
| src/compiler/utilities.ts:2152:1 | isCommonJSContainingModuleKind | 0.038555 | 0.036848 | -0.001706 | 0.9557 |
| src/compiler/utilities.ts:12429:1 | getFirstChild | 0.054718 | 0.053011 | -0.001707 | 0.9688 |
| src/compiler/tsbuildPublic.ts:186:1 | ResolvedConfigFilePath | 0.036164 | 0.034446 | -0.001718 | 0.9525 |
| src/compiler/utilities.ts:12294:1 | getOptionsSyntaxByArrayElementValue | 0.053444 | 0.051724 | -0.001720 | 0.9678 |
| src/compiler/watch.ts:944:1 | CreateWatchCompilerHostOfFilesAndCompilerOptionsInput | 0.038232 | 0.036503 | -0.001729 | 0.9548 |
| src/compiler/factory/nodeTests.ts:652:1 | isTemplateExpression | 0.041075 | 0.039345 | -0.001731 | 0.9579 |
| src/compiler/transformers/namedEvaluation.ts:354:1 | transformNamedEvaluationOfBindingElement | 0.038942 | 0.037209 | -0.001733 | 0.9555 |
| src/compiler/utilities.ts:5208:1 | getClassExtendsHeritageElement | 0.042571 | 0.040836 | -0.001735 | 0.9593 |
| src/compiler/factory/utilitiesPublic.ts:43:1 | canHaveDecorators | 0.042609 | 0.040872 | -0.001737 | 0.9592 |
| src/compiler/_namespaces/ts.ts:21:1 |  | 0.047298 | 0.045560 | -0.001737 | 0.9633 |
| src/compiler/semver.ts:39:1 |  | 0.040437 | 0.038697 | -0.001740 | 0.9570 |
| src/compiler/moduleNameResolver.ts:1157:1 | getOriginalOrResolvedModuleFileName | 0.037763 | 0.036019 | -0.001743 | 0.9538 |
| src/compiler/builder.ts:898:1 | isChangedSignature | 0.040200 | 0.038456 | -0.001744 | 0.9566 |
| src/compiler/_namespaces/ts.ts:3:1 |  | 0.037871 | 0.036125 | -0.001746 | 0.9539 |
| src/compiler/types.ts:6816:1 | ResolvedType | 0.036388 | 0.034641 | -0.001747 | 0.9520 |
| src/compiler/builder.ts:1118:1 | IncrementalMultiFileEmitBuildInfoFileInfo | 0.039478 | 0.037720 | -0.001759 | 0.9555 |
| src/compiler/utilities.ts:4964:1 | isNodeWithPossibleHoistedDeclaration | 0.041203 | 0.039442 | -0.001761 | 0.9573 |
| src/compiler/types.ts:2045:1 | ArrayBindingPattern | 0.036738 | 0.034976 | -0.001762 | 0.9520 |
| src/compiler/types.ts:7105:1 | InferencePriority | 0.036529 | 0.034765 | -0.001764 | 0.9517 |
| src/compiler/resolutionCache.ts:166:1 | ResolutionWithFailedLookupLocations | 0.037448 | 0.035679 | -0.001769 | 0.9528 |
| src/compiler/utilities.ts:5063:1 | isNodeDescendantOf | 0.045249 | 0.043476 | -0.001773 | 0.9608 |
| src/compiler/utilities.ts:8500:1 | Signature | 0.039040 | 0.037266 | -0.001773 | 0.9546 |
| src/compiler/types.ts:6895:1 | AccessFlags | 0.036551 | 0.034775 | -0.001776 | 0.9514 |
| src/compiler/factory/nodeTests.ts:1050:1 | isJSDocAllType | 0.043208 | 0.041427 | -0.001781 | 0.9588 |
| src/compiler/utilities.ts:5552:1 | NamedEvaluation | 0.036951 | 0.035167 | -0.001784 | 0.9517 |
| src/compiler/utilitiesPublic.ts:613:1 | ParameterPropertyDeclaration | 0.037688 | 0.035904 | -0.001784 | 0.9527 |
| src/compiler/types.ts:9563:1 | LexicalEnvironmentFlags | 0.038085 | 0.036301 | -0.001784 | 0.9531 |
| src/compiler/utilities.ts:10464:1 | parsePseudoBigInt | 0.038755 | 0.036970 | -0.001785 | 0.9539 |
| src/compiler/utilitiesPublic.ts:1782:1 | isTypeElement | 0.043568 | 0.041782 | -0.001786 | 0.9590 |
| src/compiler/utilities.ts:3188:1 | getThisContainer | 0.038497 | 0.036710 | -0.001787 | 0.9536 |
| src/compiler/types.ts:3866:1 | CheckJsDirective | 0.036777 | 0.034985 | -0.001792 | 0.9513 |
| src/compiler/watch.ts:161:1 |  | 0.036738 | 0.034945 | -0.001794 | 0.9512 |
| src/compiler/scanner.ts:409:1 |  | 0.040045 | 0.038251 | -0.001794 | 0.9552 |
| src/compiler/builder.ts:1584:1 | getBuilderCreationParameters | 0.038558 | 0.036763 | -0.001795 | 0.9535 |
| src/compiler/types.ts:2305:1 | IntersectionTypeNode | 0.037190 | 0.035395 | -0.001795 | 0.9517 |
| src/compiler/_namespaces/ts.ts:40:1 |  | 0.037890 | 0.036087 | -0.001804 | 0.9524 |
| src/compiler/types.ts:5512:1 | IntersectionFlags | 0.036734 | 0.034929 | -0.001805 | 0.9509 |
| src/compiler/utilities.ts:10548:1 | parseValidBigInt | 0.038647 | 0.036833 | -0.001814 | 0.9531 |
| src/compiler/types.ts:5017:1 | ExitStatus | 0.037244 | 0.035430 | -0.001814 | 0.9513 |
| src/compiler/types.ts:2280:1 | NamedTupleMember | 0.036336 | 0.034520 | -0.001816 | 0.9500 |
| src/compiler/utilities.ts:9377:1 | getNameOfScriptTarget | 0.038575 | 0.036756 | -0.001818 | 0.9529 |
| src/compiler/_namespaces/ts.ts:49:1 |  | 0.037863 | 0.036044 | -0.001819 | 0.9519 |
| src/compiler/scanner.ts:413:1 | tokenToString | 0.043977 | 0.042153 | -0.001824 | 0.9585 |
| src/compiler/factory/utilities.ts:1256:1 | isAssignmentOperatorOrHigher | 0.075263 | 0.073433 | -0.001830 | 0.9757 |
| src/compiler/transformers/utilities.ts:496:1 | getNonAssignmentOperatorForCompoundAssignment | 0.038413 | 0.036582 | -0.001831 | 0.9523 |
| src/compiler/types.ts:7164:1 | TypeComparer | 0.036214 | 0.034382 | -0.001832 | 0.9494 |
| src/compiler/moduleNameResolver.ts:2577:1 | noKeyStartsWithDot | 0.041411 | 0.039577 | -0.001834 | 0.9557 |
| src/compiler/builder.ts:1076:1 | IncrementalBuildInfoFileIdListId | 0.037058 | 0.035221 | -0.001836 | 0.9504 |
| src/compiler/scanner.ts:414:1 | tokenToString | 0.044382 | 0.042544 | -0.001838 | 0.9586 |
| src/compiler/performance.ts:59:1 |  | 0.066148 | 0.064309 | -0.001840 | 0.9722 |
| src/compiler/types.ts:3920:1 | JSDocNullableType | 0.036167 | 0.034327 | -0.001840 | 0.9491 |
| src/compiler/factory/nodeChildren.ts:1:1 |  | 0.037383 | 0.035539 | -0.001844 | 0.9507 |
| src/compiler/utilities.ts:12419:1 | suppressTrailingTrivia | 0.055342 | 0.053495 | -0.001848 | 0.9666 |
| src/compiler/utilities.ts:3479:1 | nodeCanBeDecorated | 0.042008 | 0.040160 | -0.001848 | 0.9560 |
| src/compiler/types.ts:2731:1 | AssignmentPattern | 0.037481 | 0.035631 | -0.001850 | 0.9506 |
| src/compiler/commandLineParser.ts:2267:1 | readConfigFile | 0.067872 | 0.066016 | -0.001856 | 0.9727 |
| src/compiler/parser.ts:1414:1 | parseIsolatedJSDocComment | 0.042618 | 0.040760 | -0.001858 | 0.9564 |
| src/compiler/builder.ts:133:1 | EmitSignature | 0.047776 | 0.045914 | -0.001862 | 0.9610 |
| src/compiler/factory/utilities.ts:599:1 | startsWithUseStrict | 0.278441 | 0.276578 | -0.001863 | 0.9933 |
| src/compiler/_namespaces/ts.ts:27:1 |  | 0.038451 | 0.036584 | -0.001868 | 0.9514 |
| src/compiler/_namespaces/ts.ts:69:1 |  | 0.036615 | 0.034744 | -0.001871 | 0.9489 |
| src/compiler/transformers/esDecorators.ts:269:1 | ClassElementLexicalEnvironmentStackEntry | 0.036117 | 0.034239 | -0.001879 | 0.9480 |
| src/compiler/checker.ts:1313:1 |  | 0.073611 | 0.071731 | -0.001879 | 0.9745 |
| src/compiler/utilitiesPublic.ts:1431:1 | skipPartiallyEmittedExpressions | 0.037817 | 0.035937 | -0.001880 | 0.9503 |
| src/compiler/_namespaces/ts.ts:79:1 |  | 0.037556 | 0.035674 | -0.001882 | 0.9499 |
| src/compiler/scanner.ts:431:1 | characterCodeToRegularExpressionFlag | 0.038575 | 0.036692 | -0.001883 | 0.9512 |
| src/compiler/types.ts:6828:1 | FreshObjectLiteralType | 0.036349 | 0.034466 | -0.001884 | 0.9482 |
| src/compiler/transformers/declarations.ts:1934:1 | ProcessedComponent | 0.037285 | 0.035394 | -0.001891 | 0.9493 |
| src/compiler/types.ts:4618:1 | LibFile | 0.039729 | 0.037834 | -0.001895 | 0.9523 |
| src/compiler/performance.ts:136:1 | forEachMeasure | 0.040150 | 0.038254 | -0.001896 | 0.9528 |
| src/compiler/_namespaces/ts.ts:73:1 |  | 0.037562 | 0.035665 | -0.001897 | 0.9495 |
| src/compiler/utilities.ts:715:1 | optionsHaveModuleResolutionChanges | 0.039392 | 0.037493 | -0.001899 | 0.9518 |
| src/compiler/utilities.ts:5291:1 | isStringANonContextualKeyword | 0.043288 | 0.041386 | -0.001902 | 0.9561 |
| src/compiler/transformers/utilities.ts:600:1 | getStaticPropertiesAndClassStaticBlock | 0.039246 | 0.037343 | -0.001902 | 0.9515 |
| src/compiler/utilities.ts:9296:1 |  | 0.077078 | 0.075175 | -0.001903 | 0.9753 |
| src/compiler/utilities.ts:5009:1 | walkUpParenthesizedTypes | 0.050652 | 0.048747 | -0.001905 | 0.9624 |
| src/compiler/transformers/declarations/diagnostics.ts:238:1 | createGetSymbolAccessibilityDiagnosticForNode | 0.144568 | 0.142662 | -0.001905 | 0.9868 |
| src/compiler/path.ts:606:1 | resolvePath | 0.065469 | 0.063560 | -0.001909 | 0.9708 |
| src/compiler/types.ts:1919:1 | PropertySignature | 0.036625 | 0.034715 | -0.001911 | 0.9478 |
| src/compiler/builderPublic.ts:138:1 | SemanticDiagnosticsBuilderProgram | 0.039501 | 0.037586 | -0.001915 | 0.9515 |
| src/compiler/performance.ts:114:1 | getCount | 0.043106 | 0.041181 | -0.001926 | 0.9553 |
| src/compiler/factory/nodeFactory.ts:7188:1 | getDefaultTagNameForKind | 0.041866 | 0.039939 | -0.001927 | 0.9540 |
| src/compiler/transformers/utilities.ts:652:1 | getDecoratorsOfParameters | 0.038725 | 0.036796 | -0.001929 | 0.9502 |
| src/compiler/utilities.ts:10186:1 | tryRemoveExtension | 0.038347 | 0.036416 | -0.001931 | 0.9496 |
| src/compiler/utilitiesPublic.ts:1690:1 | isBooleanLiteral | 0.043867 | 0.041935 | -0.001932 | 0.9560 |
| src/compiler/types.ts:4007:1 | JSDocAuthorTag | 0.037832 | 0.035896 | -0.001936 | 0.9488 |
| src/compiler/commandLineParser.ts:288:1 |  | 0.063479 | 0.061541 | -0.001938 | 0.9695 |
| src/compiler/semver.ts:248:1 |  | 0.041336 | 0.039398 | -0.001938 | 0.9531 |
| src/compiler/types.ts:1968:1 | PrivateIdentifierAccessorDeclaration | 0.035957 | 0.034018 | -0.001939 | 0.9461 |
| src/compiler/factory/utilities.ts:465:1 | createExpressionForObjectLiteralElementLike | 0.191834 | 0.189893 | -0.001941 | 0.9899 |
| src/compiler/parser.ts:10576:1 | parseResolutionMode | 0.037761 | 0.035820 | -0.001941 | 0.9486 |
| src/compiler/core.ts:232:1 | indexOfAnyCharCode | 0.041459 | 0.039517 | -0.001942 | 0.9532 |
| src/compiler/program.ts:5128:1 | resolveProjectReferencePath | 0.043294 | 0.041345 | -0.001948 | 0.9550 |
| src/compiler/types.ts:2051:1 | BindingPattern | 0.037128 | 0.035178 | -0.001950 | 0.9475 |
| src/compiler/types.ts:6039:1 | SymbolId | 0.036417 | 0.034466 | -0.001951 | 0.9464 |
| src/compiler/utilities.ts:3338:1 | getSuperContainer | 0.039477 | 0.037526 | -0.001951 | 0.9506 |
| src/compiler/moduleSpecifiers.ts:357:1 | getModuleSpecifiers | 0.041060 | 0.039103 | -0.001957 | 0.9523 |
| src/compiler/factory/utilities.ts:1267:1 | isBinaryOperatorToken | 0.078879 | 0.076911 | -0.001968 | 0.9751 |
| src/compiler/transformers/esnext.ts:1:1 |  | 0.039116 | 0.037141 | -0.001975 | 0.9495 |
| src/compiler/factory/nodeTests.ts:612:1 | isFunctionExpression | 0.043116 | 0.041134 | -0.001982 | 0.9540 |
| src/compiler/utilities.ts:5214:1 | getEffectiveImplementsTypeNodes | 0.050210 | 0.048227 | -0.001982 | 0.9605 |
| src/compiler/types.ts:3958:1 | JSDocTag | 0.037078 | 0.035085 | -0.001993 | 0.9463 |
| src/compiler/types.ts:1633:1 | AccessorKeyword | 0.035818 | 0.033819 | -0.001999 | 0.9442 |
| src/compiler/transformers/generators.ts:307:1 | Instruction | 0.040385 | 0.038380 | -0.002005 | 0.9504 |
| src/compiler/utilities.ts:8492:1 | Type | 0.038048 | 0.036042 | -0.002007 | 0.9473 |
| src/compiler/transformers/generators.ts:219:1 | Label | 0.039655 | 0.037648 | -0.002007 | 0.9494 |
| src/compiler/transformers/module/module.ts:1:1 |  | 0.037746 | 0.035731 | -0.002014 | 0.9466 |
| src/compiler/utilities.ts:3875:1 | isVariableDeclarationInitializedToBareOrAccessedRequire | 0.046281 | 0.044266 | -0.002015 | 0.9565 |
| src/compiler/moduleNameResolver.ts:406:1 | readPackageJsonMainField | 0.037584 | 0.035565 | -0.002018 | 0.9463 |
| src/compiler/transformers/classThis.ts:33:1 | createClassThisAssignmentBlock | 0.038612 | 0.036593 | -0.002019 | 0.9477 |
| src/compiler/transformers/destructuring.ts:610:1 | makeObjectBindingPattern | 0.038952 | 0.036930 | -0.002022 | 0.9481 |
| src/compiler/factory/utilities.ts:1186:1 | isAdditiveOperator | 0.040663 | 0.038637 | -0.002026 | 0.9502 |
| src/compiler/types.ts:6986:1 | SubstitutionType | 0.036618 | 0.034589 | -0.002029 | 0.9446 |
| src/compiler/utilities.ts:10812:1 | getContainingNodeArray | 0.040165 | 0.038134 | -0.002031 | 0.9494 |
| src/compiler/program.ts:687:1 |  | 0.039455 | 0.037421 | -0.002034 | 0.9485 |
| src/compiler/utilities.ts:1384:1 |  | 0.062843 | 0.060808 | -0.002035 | 0.9676 |
| src/compiler/commandLineParser.ts:1:1 |  | 0.038379 | 0.036343 | -0.002036 | 0.9469 |
| src/compiler/utilities.ts:9561:1 | guessDirectorySymlink | 0.038165 | 0.036124 | -0.002041 | 0.9465 |
| src/compiler/types.ts:5818:1 | SymbolAccessibilityResult | 0.036721 | 0.034669 | -0.002052 | 0.9441 |
| src/compiler/_namespaces/ts.ts:10:1 |  | 0.039447 | 0.037395 | -0.002052 | 0.9480 |
| src/compiler/parser.ts:10556:1 | isDeclarationFileName | 0.037394 | 0.035336 | -0.002058 | 0.9450 |
| src/compiler/commandLineParser.ts:3047:1 | directoryOfCombinedPath | 0.039326 | 0.037262 | -0.002064 | 0.9475 |
| src/compiler/watch.ts:987:1 | performIncrementalCompilation | 0.039019 | 0.036952 | -0.002067 | 0.9470 |
| src/compiler/_namespaces/ts.ts:26:1 |  | 0.038739 | 0.036670 | -0.002069 | 0.9466 |
| src/compiler/transformers/generators.ts:236:1 | OperationArguments | 0.039699 | 0.037626 | -0.002073 | 0.9478 |
| src/compiler/scanner.ts:4096:1 |  | 0.069388 | 0.067307 | -0.002081 | 0.9700 |
| src/compiler/_namespaces/ts.ts:53:1 |  | 0.037249 | 0.035165 | -0.002084 | 0.9441 |
| src/compiler/types.ts:6807:1 | ReverseMappedType | 0.037314 | 0.035220 | -0.002093 | 0.9439 |
| src/compiler/factory/nodeTests.ts:620:1 | isDeleteExpression | 0.042528 | 0.040432 | -0.002096 | 0.9507 |
| src/compiler/types.ts:7223:1 | FileExtensionInfo | 0.036881 | 0.034783 | -0.002099 | 0.9431 |
| src/compiler/watchPublic.ts:157:1 | CreateProgram | 0.039276 | 0.037176 | -0.002100 | 0.9465 |
| src/compiler/utilities.ts:6207:1 | escapeString | 0.039869 | 0.037767 | -0.002101 | 0.9473 |
| src/compiler/commandLineParser.ts:2357:1 |  | 0.068865 | 0.066760 | -0.002105 | 0.9694 |
| src/compiler/performance.ts:63:1 |  | 0.038156 | 0.036046 | -0.002111 | 0.9447 |
| src/compiler/sourcemap.ts:31:1 | SourceMapGeneratorOptions | 0.039597 | 0.037484 | -0.002112 | 0.9467 |
| src/compiler/factory/nodeTests.ts:426:1 | isMethodSignature | 0.045940 | 0.043826 | -0.002114 | 0.9540 |
| src/compiler/builder.ts:1164:1 | IncrementalBuildInfoBundlePendingEmit | 0.038764 | 0.036642 | -0.002122 | 0.9453 |
| src/compiler/utilities.ts:10191:1 | removeExtension | 0.039456 | 0.037331 | -0.002124 | 0.9462 |
| src/compiler/types.ts:3802:1 | NamedImportsOrExports | 0.040316 | 0.038191 | -0.002125 | 0.9473 |
| src/compiler/executeCommandLine.ts:531:1 | getHeader | 0.038842 | 0.036717 | -0.002126 | 0.9453 |
| src/compiler/factory/nodeChildren.ts:27:1 | setNodeChildren | 0.047127 | 0.044996 | -0.002131 | 0.9548 |
| src/compiler/moduleNameResolver.ts:2280:1 | loadEntrypointsFromExportMap | 0.038909 | 0.036767 | -0.002142 | 0.9449 |
| src/compiler/transformers/destructuring.ts:605:1 | makeArrayAssignmentPattern | 0.039279 | 0.037128 | -0.002151 | 0.9452 |
| src/compiler/types.ts:1796:1 | ComputedPropertyName | 0.038584 | 0.036415 | -0.002169 | 0.9438 |
| src/compiler/types.ts:4960:1 | ResolvedProjectReference | 0.038692 | 0.036523 | -0.002169 | 0.9439 |
| src/compiler/performance.ts:74:1 | mark | 0.073657 | 0.071486 | -0.002171 | 0.9705 |
| src/compiler/executeCommandLine.ts:420:1 | generateGroupOptionOutput | 0.038916 | 0.036744 | -0.002171 | 0.9442 |
| src/compiler/moduleNameResolver.ts:410:1 | readPackageJsonTypesVersionsField | 0.037821 | 0.035648 | -0.002173 | 0.9425 |
| src/compiler/tsbuildPublic.ts:136:1 | BuildOptions | 0.036567 | 0.034392 | -0.002175 | 0.9405 |
| src/compiler/scanner.ts:423:1 |  | 0.048719 | 0.046542 | -0.002177 | 0.9553 |
| src/compiler/types.ts:1792:1 | DeclarationStatement | 0.037756 | 0.035573 | -0.002183 | 0.9422 |
| src/compiler/types.ts:6969:1 | TemplateLiteralType | 0.036770 | 0.034586 | -0.002184 | 0.9406 |
| src/compiler/types.ts:7229:1 | DiagnosticMessage | 0.036275 | 0.034089 | -0.002186 | 0.9397 |
| src/compiler/utilities.ts:5485:1 | isAnonymousFunctionDefinition | 0.043610 | 0.041423 | -0.002187 | 0.9498 |
| src/compiler/transformers/es2015.ts:278:1 | CopyDirection | 0.046743 | 0.044555 | -0.002188 | 0.9532 |
| src/compiler/utilities.ts:9329:1 | emitModuleKindIsNonNodeESM | 0.038050 | 0.035850 | -0.002200 | 0.9422 |
| src/compiler/builder.ts:521:1 | repopulateDiagnostics | 0.043698 | 0.041498 | -0.002200 | 0.9496 |
| src/compiler/utilitiesPublic.ts:1794:1 | isClassOrTypeElement | 0.044532 | 0.042329 | -0.002204 | 0.9505 |
| src/compiler/utilities.ts:5242:1 | getHeritageClause | 0.038239 | 0.036034 | -0.002204 | 0.9424 |
| src/compiler/types.ts:9895:1 | RawSourceMap | 0.038776 | 0.036567 | -0.002208 | 0.9430 |
| src/compiler/performanceCore.ts:89:1 |  | 0.037664 | 0.035455 | -0.002209 | 0.9414 |
| src/compiler/transformers/utilities.ts:806:1 | LexicalEnvironment | 0.037101 | 0.034877 | -0.002225 | 0.9400 |
| src/compiler/checker.ts:1169:1 |  | 0.039581 | 0.037353 | -0.002228 | 0.9437 |
| src/compiler/types.ts:1894:1 | ParameterDeclaration | 0.037463 | 0.035231 | -0.002233 | 0.9404 |
| src/compiler/types.ts:3976:1 | JSDocLinkPlain | 0.037833 | 0.035599 | -0.002234 | 0.9409 |
| src/compiler/semver.ts:20:1 |  | 0.041906 | 0.039668 | -0.002237 | 0.9466 |
| src/compiler/types.ts:7259:1 | DiagnosticMessageChain | 0.036914 | 0.034674 | -0.002239 | 0.9393 |
| src/compiler/types.ts:5797:1 | LateVisibilityPaintedStatement | 0.036929 | 0.034686 | -0.002242 | 0.9393 |
| src/compiler/path.ts:1047:1 | getRelativePathFromDirectory | 0.039291 | 0.037045 | -0.002246 | 0.9428 |
| src/compiler/path.ts:492:1 | getPathComponents | 0.036856 | 0.034608 | -0.002248 | 0.9390 |
| src/compiler/utilities.ts:9455:1 | SymlinkedDirectory | 0.037446 | 0.035192 | -0.002254 | 0.9398 |
| src/compiler/utilities.ts:10279:1 | extensionIsTS | 0.041667 | 0.039412 | -0.002255 | 0.9459 |
| src/compiler/utilities.ts:1203:1 | isPinnedComment | 0.038664 | 0.036399 | -0.002265 | 0.9414 |
| src/compiler/resolutionCache.ts:330:1 | isInDirectoryPath | 0.039336 | 0.037070 | -0.002267 | 0.9424 |
| src/compiler/types.ts:1735:1 | EntityName | 0.038770 | 0.036501 | -0.002269 | 0.9415 |
| src/compiler/types.ts:6487:1 | FreshableIntrinsicType | 0.037890 | 0.035618 | -0.002272 | 0.9400 |
| src/compiler/utilities.ts:11107:1 | isNonNullAccess | 0.051417 | 0.049143 | -0.002274 | 0.9558 |
| src/compiler/utilitiesPublic.ts:1030:1 | getModifiers | 0.046536 | 0.044258 | -0.002279 | 0.9510 |
| src/compiler/path.ts:732:1 | simpleNormalizePath | 0.037818 | 0.035535 | -0.002283 | 0.9396 |
| src/compiler/factory/nodeChildren.ts:44:1 | unsetNodeChildren | 0.041932 | 0.039647 | -0.002286 | 0.9455 |
| src/compiler/utilities.ts:10537:1 | parseBigInt | 0.038944 | 0.036654 | -0.002289 | 0.9412 |
| src/compiler/types.ts:6268:1 | EvaluatorResult | 0.038235 | 0.035944 | -0.002291 | 0.9401 |
| src/compiler/watch.ts:912:1 | CreateWatchCompilerHostOfConfigFileInput | 0.039423 | 0.037129 | -0.002293 | 0.9418 |
| src/compiler/program.ts:335:1 | resolveTripleslashReference | 0.041903 | 0.039608 | -0.002295 | 0.9452 |
| src/compiler/types.ts:4967:1 | StructureIsReused | 0.037030 | 0.034733 | -0.002297 | 0.9380 |
| src/compiler/transformers/namedEvaluation.ts:384:1 | transformNamedEvaluationOfPropertyDeclaration | 0.039146 | 0.036836 | -0.002310 | 0.9410 |
| src/compiler/watch.ts:1:1 |  | 0.037534 | 0.035222 | -0.002313 | 0.9384 |
| src/compiler/utilitiesPublic.ts:1640:1 | isClassMemberModifier | 0.041985 | 0.039669 | -0.002316 | 0.9448 |
| src/compiler/types.ts:5808:1 | SymbolVisibilityResult | 0.037734 | 0.035415 | -0.002319 | 0.9385 |
| src/compiler/utilitiesPublic.ts:1591:1 | isFileLevelReservedGeneratedIdentifier | 0.039704 | 0.037382 | -0.002321 | 0.9415 |
| src/compiler/types.ts:2300:1 | UnionTypeNode | 0.037264 | 0.034943 | -0.002321 | 0.9377 |
| src/compiler/scanner.ts:537:1 | getLinesBetweenPositions | 0.040047 | 0.037722 | -0.002325 | 0.9419 |
| src/compiler/types.ts:2749:1 | ConciseBody | 0.038007 | 0.035681 | -0.002326 | 0.9388 |
| src/compiler/utilities.ts:7809:1 | tryParseJson | 0.037526 | 0.035199 | -0.002326 | 0.9380 |
| src/compiler/transformers/esDecorators.ts:276:1 | PropertyNameLexicalEnvironmentStackEntry | 0.036187 | 0.033856 | -0.002331 | 0.9356 |
| src/compiler/utilities.ts:5896:1 | getOperatorPrecedence | 0.046695 | 0.044364 | -0.002331 | 0.9501 |
| src/compiler/utilities.ts:3819:1 | isInJSDoc | 0.044484 | 0.042150 | -0.002334 | 0.9475 |
| src/compiler/commandLineParser.ts:2611:1 | isCompilerOptionsValue | 0.040339 | 0.038004 | -0.002335 | 0.9421 |
| src/compiler/types.ts:1748:1 | DeclarationName | 0.038155 | 0.035819 | -0.002336 | 0.9388 |
| src/compiler/builder.ts:136:1 | ReusableBuilderProgramState | 0.041869 | 0.039533 | -0.002336 | 0.9442 |
| src/compiler/utilities.ts:4750:1 | getParameterSymbolFromJSDoc | 0.190324 | 0.187982 | -0.002342 | 0.9877 |
| src/compiler/utilities.ts:5178:1 | exportAssignmentIsAlias | 0.052447 | 0.050101 | -0.002346 | 0.9553 |
| src/compiler/types.ts:5758:1 | AnyImportOrRequireStatement | 0.036687 | 0.034335 | -0.002353 | 0.9359 |
| src/compiler/scanner.ts:600:1 | isASCIILetter | 0.039840 | 0.037481 | -0.002360 | 0.9408 |
| src/compiler/_namespaces/ts.ts:16:1 |  | 0.039319 | 0.036955 | -0.002364 | 0.9399 |
| src/compiler/types.ts:1964:1 | PrivateIdentifierSetAccessorDeclaration | 0.037142 | 0.034774 | -0.002368 | 0.9362 |
| src/compiler/executeCommandLine.ts:513:1 | printAllHelp | 0.040626 | 0.038250 | -0.002376 | 0.9415 |
| src/compiler/types.ts:1746:1 | MemberName | 0.037977 | 0.035599 | -0.002378 | 0.9374 |
| src/compiler/scanner.ts:389:1 | isUnicodeIdentifierStart | 0.040056 | 0.037673 | -0.002382 | 0.9405 |
| src/compiler/types.ts:6993:1 | JsxReferenceKind | 0.036435 | 0.034052 | -0.002383 | 0.9346 |
| src/compiler/checker.ts:1469:1 | getSymbolId | 0.041289 | 0.038905 | -0.002384 | 0.9423 |
| src/compiler/transformers/classFields.ts:313:1 | PrivateIdentifierInfo | 0.039548 | 0.037159 | -0.002390 | 0.9396 |
| src/compiler/types.ts:1778:1 | LateBoundDeclaration | 0.037297 | 0.034906 | -0.002391 | 0.9359 |
| src/compiler/transformers/utilities.ts:490:1 | isCompoundAssignment | 0.038230 | 0.035834 | -0.002396 | 0.9373 |
| src/compiler/sourcemap.ts:694:1 | getGeneratedPositionOfMapping | 0.039531 | 0.037121 | -0.002410 | 0.9390 |
| src/compiler/factory/nodeTests.ts:608:1 | isParenthesizedExpression | 0.044119 | 0.041707 | -0.002412 | 0.9453 |
| src/compiler/visitorPublic.ts:522:1 | visitFunctionBody | 0.036988 | 0.034572 | -0.002416 | 0.9347 |
| src/compiler/types.ts:3481:1 | CaseClause | 0.038691 | 0.036275 | -0.002416 | 0.9376 |
| src/compiler/utilities.ts:3454:1 | getInvokedExpression | 0.047939 | 0.045522 | -0.002417 | 0.9496 |
| src/compiler/utilitiesPublic.ts:1146:1 | getJSDocProtectedTag | 0.055944 | 0.053524 | -0.002419 | 0.9568 |
| src/compiler/types.ts:6139:1 | TransientSymbolLinks | 0.037998 | 0.035575 | -0.002423 | 0.9362 |
| src/compiler/sys.ts:1396:1 | System | 0.037597 | 0.035172 | -0.002425 | 0.9355 |
| src/compiler/utilitiesPublic.ts:477:1 |  | 0.037550 | 0.035124 | -0.002426 | 0.9354 |
| src/compiler/utilities.ts:668:1 |  | 0.037810 | 0.035377 | -0.002433 | 0.9357 |
| src/compiler/types.ts:5843:1 | TypeReferenceSerializationKind | 0.037637 | 0.035198 | -0.002439 | 0.9352 |
| src/compiler/commandLineParser.ts:1795:1 |  | 0.039286 | 0.036842 | -0.002444 | 0.9378 |
| src/compiler/scanner.ts:395:1 | isUnicodeIdentifierPart | 0.039499 | 0.037053 | -0.002446 | 0.9381 |
| src/compiler/types.ts:2396:1 | Expression | 0.038204 | 0.035756 | -0.002448 | 0.9359 |
| src/compiler/moduleNameResolver.ts:501:1 | getDefaultTypeRoots | 0.038252 | 0.035798 | -0.002454 | 0.9358 |
| src/compiler/utilities.ts:6918:1 | getJSDocTypeParameterDeclarations | 0.066868 | 0.064411 | -0.002457 | 0.9633 |
| src/compiler/types.ts:4002:1 | JSDocImplementsTag | 0.037178 | 0.034720 | -0.002458 | 0.9339 |
| src/compiler/factory/emitHelpers.ts:767:1 |  | 0.063958 | 0.061496 | -0.002462 | 0.9615 |
| src/compiler/utilities.ts:670:1 | createSingleLineStringWriter | 0.037717 | 0.035251 | -0.002465 | 0.9346 |
| src/compiler/utilitiesPublic.ts:2116:1 | isIterationStatement | 0.118443 | 0.115975 | -0.002468 | 0.9792 |
| src/compiler/utilities.ts:4487:1 | isTypeAlias | 0.235544 | 0.233071 | -0.002473 | 0.9895 |
| src/compiler/types.ts:963:1 | JSDocArray | 0.048411 | 0.045933 | -0.002478 | 0.9488 |
| src/compiler/utilities.ts:7147:1 | writeTrimmedCurrentLine | 0.147527 | 0.145044 | -0.002483 | 0.9832 |
| src/compiler/types.ts:7087:1 | TypeMapKind | 0.038215 | 0.035729 | -0.002486 | 0.9350 |
| src/compiler/checker.ts:54236:1 | ReactNames | 0.037914 | 0.035428 | -0.002486 | 0.9344 |
| src/compiler/types.ts:3982:1 | JSDocComment | 0.036757 | 0.034263 | -0.002494 | 0.9321 |
| src/compiler/utilities.ts:5459:1 | isPrivateIdentifierSymbol | 0.037699 | 0.035205 | -0.002494 | 0.9338 |
| src/compiler/factory/utilities.ts:1559:1 | formatGeneratedNamePart | 0.040927 | 0.038433 | -0.002494 | 0.9391 |
| src/compiler/checker.ts:54240:1 | getIterationTypesKeyFromIterationTypeKind | 0.038179 | 0.035683 | -0.002497 | 0.9346 |
| src/compiler/utilities.ts:9032:1 | CompilerOptionKeys | 0.038260 | 0.035762 | -0.002498 | 0.9347 |
| src/compiler/moduleNameResolver.ts:3418:1 | traceIfEnabled | 0.043412 | 0.040912 | -0.002500 | 0.9424 |
| src/compiler/types.ts:2039:1 | ObjectBindingPattern | 0.037918 | 0.035407 | -0.002511 | 0.9338 |
| src/compiler/factory/nodeTests.ts:748:1 | isForInStatement | 0.046171 | 0.043657 | -0.002514 | 0.9455 |
| src/compiler/tsbuildPublic.ts:134:1 |  | 0.036814 | 0.034293 | -0.002521 | 0.9315 |
| src/compiler/utilities.ts:2157:1 | isEffectiveExternalModule | 0.038802 | 0.036273 | -0.002529 | 0.9348 |
| src/compiler/types.ts:1737:1 | PropertyName | 0.038700 | 0.036170 | -0.002530 | 0.9346 |
| src/compiler/transformers/esDecorators.ts:262:1 | ClassLexicalEnvironmentStackEntry | 0.036722 | 0.034184 | -0.002538 | 0.9309 |
| src/compiler/tsbuildPublic.ts:349:1 | SolutionBuilderStateCache | 0.038409 | 0.035871 | -0.002538 | 0.9339 |
| src/compiler/visitorPublic.ts:619:1 | VisitEachChildTable | 0.037204 | 0.034664 | -0.002540 | 0.9317 |
| src/compiler/utilities.ts:11091:1 | isOptionalDeclaration | 0.047920 | 0.045378 | -0.002541 | 0.9470 |
| src/compiler/_namespaces/ts.ts:58:1 |  | 0.036516 | 0.033973 | -0.002542 | 0.9304 |
| src/compiler/types.ts:5902:1 | EmitResolver | 0.037453 | 0.034906 | -0.002546 | 0.9320 |
| src/compiler/resolutionCache.ts:343:1 | canWatchAffectingLocation | 0.039141 | 0.036595 | -0.002547 | 0.9349 |
| src/compiler/_namespaces/ts.ts:48:1 |  | 0.039803 | 0.037254 | -0.002549 | 0.9360 |
| src/compiler/commandLineParser.ts:3353:1 | isSolutionConfig | 0.040412 | 0.037863 | -0.002549 | 0.9369 |
| src/compiler/utilities.ts:3917:1 | isStringDoubleQuoted | 0.049354 | 0.046803 | -0.002551 | 0.9483 |
| src/compiler/transformers/namedEvaluation.ts:236:1 | finishTransformNamedEvaluation | 0.042529 | 0.039974 | -0.002556 | 0.9399 |
| src/compiler/parser.ts:9946:1 | IncrementalParser | 0.037682 | 0.035116 | -0.002566 | 0.9319 |
| src/compiler/utilities.ts:8944:1 | messageTextEqualityComparer | 0.038949 | 0.036381 | -0.002567 | 0.9341 |
| src/compiler/moduleNameResolver.ts:458:1 |  | 0.036726 | 0.034159 | -0.002568 | 0.9301 |
| src/compiler/factory/nodeChildren.ts:55:1 | transferSourceFileChildren | 0.037924 | 0.035354 | -0.002569 | 0.9322 |
| src/compiler/transformers/utilities.ts:612:1 | isInitializedOrStaticProperty | 0.265187 | 0.262614 | -0.002573 | 0.9903 |
| src/compiler/types.ts:6611:1 | ObjectType | 0.037716 | 0.035141 | -0.002575 | 0.9317 |
| src/compiler/utilities.ts:9289:1 |  | 0.040659 | 0.038079 | -0.002580 | 0.9365 |
| src/compiler/transformers/es2016.ts:1:1 |  | 0.038592 | 0.036010 | -0.002582 | 0.9331 |
| src/compiler/types.ts:2758:1 | ArrowFunction | 0.039040 | 0.036456 | -0.002584 | 0.9338 |
| src/compiler/sourcemap.ts:363:1 |  | 0.039384 | 0.036799 | -0.002585 | 0.9344 |
| src/compiler/types.ts:2784:1 | LiteralExpression | 0.037801 | 0.035213 | -0.002588 | 0.9315 |
| src/compiler/sourcemap.ts:404:1 | isStringOrNull | 0.038384 | 0.035795 | -0.002589 | 0.9325 |
| src/compiler/commandLineParser.ts:2329:1 | getCommandLineCompilerOptionsMap | 0.040344 | 0.037752 | -0.002592 | 0.9358 |
| src/compiler/commandLineParser.ts:3348:1 | shouldReportNoInputFiles | 0.037971 | 0.035379 | -0.002592 | 0.9317 |
| src/compiler/factory/nodeTests.ts:508:1 | isUnionTypeNode | 0.042985 | 0.040386 | -0.002598 | 0.9396 |
| src/compiler/types.ts:9958:1 | DocumentPositionMapper | 0.038700 | 0.036100 | -0.002600 | 0.9328 |
| src/compiler/types.ts:5779:1 | VariableDeclarationInitializedTo | 0.037269 | 0.034666 | -0.002603 | 0.9302 |
| src/compiler/utilities.ts:10148:1 | isSupportedSourceFileName | 0.040336 | 0.037730 | -0.002606 | 0.9354 |
| src/compiler/types.ts:1956:1 | PrivateIdentifierMethodDeclaration | 0.037822 | 0.035213 | -0.002609 | 0.9310 |
| src/compiler/watchUtilities.ts:101:1 | Canonicalized | 0.038113 | 0.035499 | -0.002614 | 0.9314 |
| src/compiler/types.ts:1762:1 | NamedDeclaration | 0.037735 | 0.035120 | -0.002615 | 0.9307 |
| src/compiler/types.ts:1641:1 | ProtectedKeyword | 0.038277 | 0.035656 | -0.002621 | 0.9315 |
| src/compiler/commandLineParser.ts:572:1 |  | 0.093020 | 0.090395 | -0.002624 | 0.9718 |
| src/compiler/types.ts:3997:1 | JSDocAugmentsTag | 0.037359 | 0.034726 | -0.002633 | 0.9295 |
| src/compiler/utilities.ts:12434:1 | canHaveStatements | 0.240164 | 0.237531 | -0.002633 | 0.9890 |
| src/compiler/types.ts:1885:1 | InitializedVariableDeclaration | 0.037274 | 0.034637 | -0.002637 | 0.9293 |
| src/compiler/parser.ts:10713:1 |  | 0.038238 | 0.035595 | -0.002643 | 0.9309 |
| src/compiler/emitter.ts:1206:1 |  | 0.040422 | 0.037775 | -0.002647 | 0.9345 |
| src/compiler/parser.ts:1427:1 | parseJSDocTypeExpressionForTests | 0.044776 | 0.042125 | -0.002651 | 0.9408 |
| src/compiler/types.ts:10681:1 | IntroducesNewScopeNode | 0.037760 | 0.035108 | -0.002652 | 0.9298 |
| src/compiler/types.ts:7374:1 | PluginImport | 0.037926 | 0.035270 | -0.002656 | 0.9300 |
| src/compiler/types.ts:2735:1 | ConditionalExpression | 0.037466 | 0.034808 | -0.002657 | 0.9291 |
| src/compiler/types.ts:1943:1 | AutoAccessorPropertyDeclaration | 0.037111 | 0.034450 | -0.002660 | 0.9283 |
| src/compiler/types.ts:2688:1 | ObjectAssignmentElement | 0.039276 | 0.036615 | -0.002661 | 0.9323 |
| src/compiler/utilities.ts:2706:1 | isImportCall | 0.157896 | 0.155233 | -0.002663 | 0.9831 |
| src/compiler/types.ts:2293:1 | RestTypeNode | 0.037140 | 0.034474 | -0.002666 | 0.9282 |
| src/compiler/_namespaces/ts.ts:50:1 |  | 0.038982 | 0.036315 | -0.002666 | 0.9316 |
| src/compiler/transformers/jsx.ts:1:1 |  | 0.040428 | 0.037758 | -0.002670 | 0.9340 |
| src/compiler/utilities.ts:3804:1 | isInJSFile | 0.040775 | 0.038104 | -0.002671 | 0.9345 |
| src/compiler/path.ts:807:1 | ensureTrailingDirectorySeparator | 0.037292 | 0.034614 | -0.002677 | 0.9282 |
| src/compiler/transformers/classThis.ts:1:1 |  | 0.039660 | 0.036980 | -0.002680 | 0.9324 |
| src/compiler/types.ts:5776:1 | RequireOrImportCall | 0.037876 | 0.035189 | -0.002687 | 0.9291 |
| src/compiler/utilitiesPublic.ts:1072:1 | getJSDocParameterTagsNoCache | 0.043739 | 0.041046 | -0.002694 | 0.9384 |
| src/compiler/parser.ts:10781:1 | getNamedPragmaArguments | 0.038591 | 0.035896 | -0.002695 | 0.9302 |
| src/compiler/utilities.ts:4837:1 | hasTypeArguments | 0.040289 | 0.037594 | -0.002695 | 0.9331 |
| src/compiler/types.ts:1648:1 | Modifier | 0.037664 | 0.034965 | -0.002700 | 0.9283 |
| src/compiler/scanner.ts:436:1 | computeLineStarts | 0.040017 | 0.037316 | -0.002701 | 0.9325 |
| src/compiler/types.ts:1932:1 | PropertyDeclaration | 0.036722 | 0.034018 | -0.002704 | 0.9264 |
| src/compiler/utilities.ts:6277:1 | isIntrinsicJsxName | 0.039423 | 0.036719 | -0.002704 | 0.9314 |
| src/compiler/utilities.ts:10453:1 | isJsonEqual | 0.040391 | 0.037680 | -0.002711 | 0.9329 |
| src/compiler/factory/nodeTests.ts:528:1 | isThisTypeNode | 0.042717 | 0.040004 | -0.002713 | 0.9365 |
| src/compiler/types.ts:2005:1 | ShorthandPropertyAssignment | 0.038744 | 0.036027 | -0.002717 | 0.9299 |
| src/compiler/utilities.ts:6911:1 | getEffectiveReturnTypeNode | 0.193974 | 0.191256 | -0.002718 | 0.9860 |
| src/compiler/utilities.ts:9387:1 | compilerOptionsAffectSemanticDiagnostics | 0.038422 | 0.035700 | -0.002721 | 0.9292 |
| src/compiler/builder.ts:1220:1 | hasSyntaxOrGlobalErrors | 0.183890 | 0.181168 | -0.002723 | 0.9852 |
| src/compiler/executeCommandLine.ts:1054:1 | SolutionPerformance | 0.038492 | 0.035766 | -0.002726 | 0.9292 |
| src/compiler/utilities.ts:5522:1 | isNamedEvaluationSource | 0.164425 | 0.161699 | -0.002726 | 0.9834 |
| src/compiler/moduleSpecifiers.ts:1014:1 | MatchingMode | 0.038297 | 0.035570 | -0.002727 | 0.9288 |
| src/compiler/types.ts:8677:1 | CallBinding | 0.040112 | 0.037363 | -0.002750 | 0.9315 |
| src/compiler/types.ts:5794:1 | CanHaveModuleSpecifier | 0.037704 | 0.034954 | -0.002751 | 0.9270 |
| src/compiler/factory/emitHelpers.ts:838:1 |  | 0.064113 | 0.061355 | -0.002758 | 0.9570 |
| src/compiler/builder.ts:1125:1 | IncrementalBuildInfoRoot | 0.038933 | 0.036172 | -0.002762 | 0.9291 |
| src/compiler/moduleSpecifiers.ts:379:1 | ModuleSpecifierResult | 0.040646 | 0.037882 | -0.002763 | 0.9320 |
| src/compiler/semver.ts:238:1 | Comparator | 0.040007 | 0.037239 | -0.002768 | 0.9308 |
| src/compiler/types.ts:8347:1 | SnippetElement | 0.040086 | 0.037303 | -0.002783 | 0.9306 |
| src/compiler/utilitiesPublic.ts:2633:1 | hasRestParameter | 0.046451 | 0.043659 | -0.002792 | 0.9399 |
| src/compiler/types.ts:8141:1 | CompilerHost | 0.037800 | 0.034993 | -0.002807 | 0.9257 |
| src/compiler/moduleSpecifiers.ts:1424:1 | getJSExtensionForFile | 0.040138 | 0.037328 | -0.002810 | 0.9300 |
| src/compiler/types.ts:5698:1 | ThisTypePredicate | 0.037433 | 0.034620 | -0.002813 | 0.9248 |
| src/compiler/commandLineParser.ts:2627:1 | TSConfig | 0.039578 | 0.036763 | -0.002815 | 0.9289 |
| src/compiler/_namespaces/ts.ts:56:1 |  | 0.038371 | 0.035547 | -0.002824 | 0.9264 |
| src/compiler/utilities.ts:9372:1 | getStrictOptionValue | 0.037983 | 0.035159 | -0.002824 | 0.9256 |
| src/compiler/types.ts:1711:1 | TransientIdentifier | 0.038870 | 0.036045 | -0.002824 | 0.9273 |
| src/compiler/builderState.ts:1:1 |  | 0.038160 | 0.035333 | -0.002827 | 0.9259 |
| src/compiler/types.ts:1061:1 | HasChildren | 0.043013 | 0.040183 | -0.002830 | 0.9342 |
| src/compiler/types.ts:7324:1 | DiagnosticCategory | 0.037392 | 0.034562 | -0.002830 | 0.9243 |
| src/compiler/transformers/es2020.ts:1:1 |  | 0.037609 | 0.034778 | -0.002831 | 0.9247 |
| src/compiler/commandLineParser.ts:3364:1 | updateErrorForNoInputFiles | 0.041607 | 0.038771 | -0.002836 | 0.9318 |
| src/compiler/types.ts:8740:1 | ImmediatelyInvokedArrowFunction | 0.039667 | 0.036824 | -0.002843 | 0.9283 |
| src/compiler/utilities.ts:6161:1 |  | 0.038865 | 0.036017 | -0.002849 | 0.9267 |
| src/compiler/moduleNameResolver.ts:1656:1 | tryLoadModuleUsingBaseUrl | 0.041192 | 0.038341 | -0.002851 | 0.9308 |
| src/compiler/types.ts:6932:1 | IndexType | 0.037257 | 0.034390 | -0.002867 | 0.9230 |
| src/compiler/utilities.ts:8133:1 | reverseAccessKind | 0.038713 | 0.035842 | -0.002871 | 0.9258 |
| src/compiler/types.ts:2339:1 | IndexedAccessTypeNode | 0.037993 | 0.035113 | -0.002880 | 0.9242 |
| src/compiler/builder.ts:617:1 | releaseCache | 0.050111 | 0.047229 | -0.002881 | 0.9425 |
| src/compiler/types.ts:1667:1 | AccessibilityModifier | 0.038240 | 0.035353 | -0.002887 | 0.9245 |
| src/compiler/visitorPublic.ts:607:1 | VisitEachChildFunction | 0.038307 | 0.035415 | -0.002892 | 0.9245 |
| src/compiler/types.ts:1887:1 | VariableDeclarationList | 0.037630 | 0.034737 | -0.002892 | 0.9231 |
| src/compiler/utilities.ts:5271:1 | isPunctuation | 0.040586 | 0.037694 | -0.002893 | 0.9287 |
| src/compiler/factory/nodeFactory.ts:7299:1 | propagateChildFlags | 0.046590 | 0.043694 | -0.002896 | 0.9378 |
| src/compiler/types.ts:8356:1 | Placeholder | 0.039666 | 0.036769 | -0.002897 | 0.9270 |
| src/compiler/core.ts:2442:1 | tryRemovePrefix | 0.040879 | 0.037981 | -0.002898 | 0.9291 |
| src/compiler/types.ts:6144:1 | TransientSymbol | 0.038849 | 0.035945 | -0.002904 | 0.9253 |
| src/compiler/factory/utilities.ts:204:1 | createReactNamespace | 0.079141 | 0.076231 | -0.002910 | 0.9632 |
| src/compiler/utilities.ts:3531:1 | nodeIsDecorated | 0.051580 | 0.048655 | -0.002926 | 0.9433 |
| src/compiler/types.ts:5741:1 | AliasDeclarationNode | 0.037833 | 0.034903 | -0.002930 | 0.9226 |
| src/compiler/factory/nodeTests.ts:304:1 | isColonToken | 0.047663 | 0.044731 | -0.002932 | 0.9385 |
| src/compiler/utilities.ts:3799:1 | isSourceFileJS | 0.041643 | 0.038709 | -0.002934 | 0.9295 |
| src/compiler/factory/utilities.ts:1273:1 | BinaryExpressionState | 0.050299 | 0.047358 | -0.002941 | 0.9415 |
| src/compiler/builder.ts:1157:1 | IncrementalBundleEmitBuildInfoFileInfo | 0.038185 | 0.035244 | -0.002941 | 0.9230 |
| src/compiler/utilitiesPublic.ts:348:1 | textSpanEnd | 0.038142 | 0.035199 | -0.002943 | 0.9228 |
| src/compiler/utilities.ts:6160:1 |  | 0.040506 | 0.037562 | -0.002944 | 0.9273 |
| src/compiler/types.ts:1685:1 | GeneratedIdentifierFlags | 0.038542 | 0.035597 | -0.002945 | 0.9236 |
| src/compiler/utilitiesPublic.ts:1748:1 | isAccessor | 0.045089 | 0.042132 | -0.002957 | 0.9344 |
| src/compiler/factory/nodeTests.ts:500:1 | isOptionalTypeNode | 0.043617 | 0.040658 | -0.002959 | 0.9322 |
| src/compiler/builder.ts:1173:1 | IncrementalBuildInfo | 0.039142 | 0.036183 | -0.002959 | 0.9244 |
| src/compiler/utilities.ts:5381:1 | hasDynamicName | 0.041701 | 0.038740 | -0.002961 | 0.9290 |
| src/compiler/moduleNameResolver.ts:193:1 | formatExtensions | 0.038364 | 0.035402 | -0.002962 | 0.9228 |
| src/compiler/utilities.ts:2970:1 | getRestParameterElementType | 0.159341 | 0.156377 | -0.002964 | 0.9814 |
| src/compiler/types.ts:6167:1 | ReverseMappedSymbol | 0.038528 | 0.035558 | -0.002971 | 0.9229 |
| src/compiler/types.ts:7078:1 | IndexInfo | 0.038484 | 0.035500 | -0.002984 | 0.9225 |
| src/compiler/emitter.ts:535:1 | getSourceMapFilePath | 0.039251 | 0.036261 | -0.002990 | 0.9238 |
| src/compiler/commandLineParser.ts:282:1 |  | 0.037938 | 0.034937 | -0.003001 | 0.9209 |
| src/compiler/utilities.ts:3132:1 | getContainingFunctionDeclaration | 0.043019 | 0.040017 | -0.003002 | 0.9302 |
| src/compiler/performance.ts:62:1 |  | 0.039953 | 0.036941 | -0.003011 | 0.9246 |
| src/compiler/executeCommandLine.ts:1067:1 | createSolutionPerfomrance | 0.059706 | 0.056692 | -0.003014 | 0.9495 |
| src/compiler/utilitiesPublic.ts:2573:1 | hasInitializer | 0.042311 | 0.039285 | -0.003026 | 0.9285 |
| src/compiler/types.ts:4995:1 | SourceMapSpan | 0.037783 | 0.034751 | -0.003032 | 0.9197 |
| src/compiler/utilities.ts:6163:1 |  | 0.038925 | 0.035892 | -0.003033 | 0.9221 |
| src/compiler/_namespaces/ts.ts:52:1 |  | 0.038107 | 0.035071 | -0.003036 | 0.9203 |
| src/compiler/resolutionCache.ts:243:1 | PackageDirWatcher | 0.039740 | 0.036699 | -0.003040 | 0.9235 |
| src/compiler/utilities.ts:5266:1 | isKeyword | 0.041022 | 0.037976 | -0.003046 | 0.9257 |
| src/compiler/utilities.ts:6187:1 | getReplacement | 0.040064 | 0.037018 | -0.003046 | 0.9240 |
| src/compiler/utilities.ts:8984:1 | getSetExternalModuleIndicator | 0.038617 | 0.035564 | -0.003053 | 0.9209 |
| src/compiler/types.ts:1642:1 | PublicKeyword | 0.037911 | 0.034857 | -0.003055 | 0.9194 |
| src/compiler/factory/emitHelpers.ts:1:1 |  | 0.040257 | 0.037201 | -0.003056 | 0.9241 |
| src/compiler/executeCommandLine.ts:992:1 | updateWatchCompilationHost | 0.076242 | 0.073185 | -0.003057 | 0.9599 |
| src/compiler/types.ts:7728:1 | ConfigFileSpecs | 0.044198 | 0.041119 | -0.003079 | 0.9303 |
| src/compiler/types.ts:1986:1 | ObjectLiteralElementLike | 0.038228 | 0.035141 | -0.003087 | 0.9193 |
| src/compiler/checker.ts:54252:1 | signatureHasRestParameter | 0.038381 | 0.035288 | -0.003093 | 0.9194 |
| src/compiler/types.ts:1665:1 | ModifierLike | 0.037900 | 0.034805 | -0.003094 | 0.9184 |
| src/compiler/utilities.ts:8574:1 | addObjectAllocatorPatcher | 0.038819 | 0.035719 | -0.003100 | 0.9201 |
| src/compiler/utilities.ts:7344:1 | getSyntacticModifierFlagsNoCache | 0.229206 | 0.226103 | -0.003103 | 0.9865 |
| src/compiler/types.ts:5831:1 | AllDecorators | 0.038120 | 0.035015 | -0.003104 | 0.9186 |
| src/compiler/types.ts:6790:1 | MappedType | 0.037571 | 0.034464 | -0.003107 | 0.9173 |
| src/compiler/types.ts:1952:1 | PrivateIdentifierAutoAccessorPropertyDeclaration | 0.036963 | 0.033850 | -0.003114 | 0.9158 |
| src/compiler/utilities.ts:4041:1 | getNameOfExpando | 0.272569 | 0.269455 | -0.003114 | 0.9886 |
| src/compiler/commandLineParser.ts:2231:1 | getParsedCommandLineOfConfigFile | 0.188959 | 0.185833 | -0.003126 | 0.9835 |
| src/compiler/scanner.ts:419:1 | stringToToken | 0.041932 | 0.038806 | -0.003126 | 0.9254 |
| src/compiler/types.ts:1970:1 | PrivateClassElementDeclaration | 0.036896 | 0.033769 | -0.003127 | 0.9152 |
| src/compiler/types.ts:7359:1 | ModuleDetectionKind | 0.037912 | 0.034785 | -0.003128 | 0.9175 |
| src/compiler/builder.ts:1166:1 | IncrementalBundleEmitBuildInfo | 0.039532 | 0.036402 | -0.003130 | 0.9208 |
| src/compiler/utilities.ts:9503:1 | createSymlinkCache | 0.039487 | 0.036353 | -0.003134 | 0.9206 |
| src/compiler/commandLineParser.ts:605:1 |  | 0.090071 | 0.086927 | -0.003144 | 0.9651 |
| src/compiler/path.ts:356:1 | getBaseFileName | 0.038619 | 0.035469 | -0.003150 | 0.9184 |
| src/compiler/utilitiesPublic.ts:1657:1 | isPropertyName | 0.047925 | 0.044774 | -0.003151 | 0.9343 |
| src/compiler/utilities.ts:9308:1 |  | 0.076019 | 0.072867 | -0.003152 | 0.9585 |
| src/compiler/factory/nodeTests.ts:422:1 | isPropertyDeclaration | 0.045347 | 0.042191 | -0.003156 | 0.9304 |
| src/compiler/factory/nodeTests.ts:504:1 | isRestTypeNode | 0.043376 | 0.040218 | -0.003159 | 0.9272 |
| src/compiler/utilities.ts:3760:1 | isPartOfTypeQuery | 0.045249 | 0.042089 | -0.003160 | 0.9302 |
| src/compiler/types.ts:6657:1 | TypeReference | 0.037555 | 0.034390 | -0.003165 | 0.9157 |
| src/compiler/utilities.ts:9469:1 | SymlinkCache | 0.038927 | 0.035759 | -0.003168 | 0.9186 |
| src/compiler/builder.ts:1130:1 | IncrementalBuildInfoBase | 0.039093 | 0.035923 | -0.003170 | 0.9189 |
| src/compiler/factory/nodeTests.ts:604:1 | isTypeAssertionExpression | 0.044625 | 0.041452 | -0.003174 | 0.9289 |
| src/compiler/utilities.ts:6284:1 | getIndentString | 0.040553 | 0.037377 | -0.003175 | 0.9217 |
| src/compiler/factory/nodeTests.ts:243:1 | isStringLiteral | 0.042562 | 0.039380 | -0.003182 | 0.9252 |
| src/compiler/checker.ts:54202:1 | isNotAccessor | 0.227318 | 0.224130 | -0.003187 | 0.9860 |
| src/compiler/utilities.ts:7195:1 | hasSyntacticModifier | 0.047979 | 0.044791 | -0.003188 | 0.9336 |
| src/compiler/factory/nodeFactory.ts:7407:1 |  | 0.040370 | 0.037169 | -0.003201 | 0.9207 |
| src/compiler/utilitiesPublic.ts:2208:1 | canHaveSymbol | 0.043811 | 0.040599 | -0.003211 | 0.9267 |
| src/compiler/scanner.ts:426:1 | regularExpressionFlagToCharacterCode | 0.040140 | 0.036917 | -0.003223 | 0.9197 |
| src/compiler/types.ts:7071:1 | IndexKind | 0.038264 | 0.035032 | -0.003232 | 0.9155 |
| src/compiler/builder.ts:2452:1 | createRedirectedBuilderProgram | 0.040382 | 0.037145 | -0.003238 | 0.9198 |
| src/compiler/builder.ts:1127:1 | IncrementalBuildInfoResolvedRoot | 0.039191 | 0.035951 | -0.003241 | 0.9173 |
| src/compiler/factory/emitHelpers.ts:1281:1 |  | 0.076213 | 0.072963 | -0.003250 | 0.9574 |
| src/compiler/utilities.ts:4854:1 | getAssignmentTarget | 0.162024 | 0.158771 | -0.003253 | 0.9799 |
| src/compiler/types.ts:6951:1 | ConditionalType | 0.037598 | 0.034343 | -0.003255 | 0.9134 |
| src/compiler/utilitiesPublic.ts:946:1 | isNamedDeclaration | 0.150306 | 0.147051 | -0.003256 | 0.9783 |
| src/compiler/utilitiesPublic.ts:1948:1 | isPropertyAccessOrQualifiedNameOrImportTypeNode | 0.043006 | 0.039747 | -0.003259 | 0.9242 |
| src/compiler/utilities.ts:6814:1 | isThisInTypeQuery | 0.233385 | 0.230117 | -0.003267 | 0.9860 |
| src/compiler/types.ts:2251:1 | TypePredicateNode | 0.037820 | 0.034548 | -0.003271 | 0.9135 |
| src/compiler/parser.ts:10619:1 | processPragmasIntoFields | 0.078381 | 0.075109 | -0.003273 | 0.9582 |
| src/compiler/types.ts:7076:1 | ElementWithComputedPropertyName | 0.037869 | 0.034585 | -0.003283 | 0.9133 |
| src/compiler/types.ts:1294:1 | HasIllegalTypeParameters | 0.044265 | 0.040973 | -0.003291 | 0.9256 |
| src/compiler/visitorPublic.ts:474:1 | addDefaultValueAssignmentForInitializer | 0.193752 | 0.190448 | -0.003304 | 0.9829 |
| src/compiler/types.ts:2246:1 | TypeReferenceNode | 0.038771 | 0.035466 | -0.003305 | 0.9148 |
| src/compiler/types.ts:1978:1 | InitializedPropertyDeclaration | 0.037312 | 0.034001 | -0.003311 | 0.9113 |
| src/compiler/types.ts:1725:1 | GeneratedIdentifier | 0.038184 | 0.034866 | -0.003318 | 0.9131 |
| src/compiler/types.ts:2244:1 | TypeReferenceType | 0.038079 | 0.034757 | -0.003321 | 0.9128 |
| src/compiler/utilities.ts:5449:1 | getSymbolNameForPrivateIdentifier | 0.040642 | 0.037307 | -0.003334 | 0.9180 |
| src/compiler/utilities.ts:9334:1 | hasJsonModuleEmitEnabled | 0.038977 | 0.035639 | -0.003338 | 0.9144 |
| src/compiler/utilities.ts:7944:1 | getLinesBetweenRangeEndAndRangeStart | 0.041486 | 0.038147 | -0.003339 | 0.9195 |
| src/compiler/types.ts:8731:1 | GeneratedNamePart | 0.040530 | 0.037190 | -0.003340 | 0.9176 |
| src/compiler/types.ts:5693:1 | TypePredicateBase | 0.037981 | 0.034634 | -0.003347 | 0.9119 |
| src/compiler/types.ts:5784:1 | RequireVariableStatement | 0.037590 | 0.034223 | -0.003367 | 0.9104 |
| src/compiler/commandLineParser.ts:2328:1 |  | 0.040268 | 0.036901 | -0.003368 | 0.9164 |
| src/compiler/factory/utilities.ts:309:1 | createForOfBindingStatement | 0.193865 | 0.190496 | -0.003369 | 0.9826 |
| src/compiler/commandLineParser.ts:1958:1 | parseCommandLineWorker | 0.112678 | 0.109308 | -0.003370 | 0.9701 |
| src/compiler/utilitiesPublic.ts:376:1 | textRangeContainsTextSpan | 0.038012 | 0.034638 | -0.003374 | 0.9112 |
| src/compiler/utilities.ts:5454:1 | isKnownSymbol | 0.038889 | 0.035499 | -0.003391 | 0.9128 |
| src/compiler/executeCommandLine.ts:1095:1 | reportSolutionBuilderTimes | 0.178146 | 0.174748 | -0.003397 | 0.9809 |
| src/compiler/types.ts:2721:1 | ObjectBindingOrAssignmentPattern | 0.040283 | 0.036884 | -0.003399 | 0.9156 |
| src/compiler/transformers/es2015.ts:483:1 | createSpreadSegment | 0.078687 | 0.075285 | -0.003402 | 0.9568 |
| src/compiler/utilities.ts:1971:1 | GetLiteralTextFlags | 0.041429 | 0.038025 | -0.003404 | 0.9178 |
| src/compiler/types.ts:2275:1 | TupleTypeNode | 0.040870 | 0.037453 | -0.003417 | 0.9164 |
| src/compiler/utilities.ts:8184:1 | MutateMapSkippingNewValuesOptions | 0.038275 | 0.034853 | -0.003422 | 0.9106 |
| src/compiler/parser.ts:484:1 | getImportMetaIfNecessary | 0.050515 | 0.047090 | -0.003425 | 0.9322 |
| src/compiler/moduleNameResolver.ts:2564:1 | parsePackageName | 0.041527 | 0.038100 | -0.003427 | 0.9175 |
| src/compiler/utilities.ts:1370:1 | getInternalEmitFlags | 0.044542 | 0.041114 | -0.003428 | 0.9230 |
| src/compiler/types.ts:10672:1 | SyntacticNodeBuilder | 0.038189 | 0.034758 | -0.003431 | 0.9102 |
| src/compiler/utilities.ts:3778:1 | getExternalModuleImportEqualsDeclarationExpression | 0.048323 | 0.044892 | -0.003431 | 0.9290 |
| src/compiler/utilities.ts:10173:1 |  | 0.040284 | 0.036848 | -0.003436 | 0.9147 |
| src/compiler/types.ts:1717:1 | AutoGenerateInfo | 0.038729 | 0.035289 | -0.003440 | 0.9112 |
| src/compiler/utilitiesPublic.ts:615:1 | isParameterPropertyDeclaration | 0.265520 | 0.262078 | -0.003443 | 0.9870 |
| src/compiler/core.ts:1773:1 | isNumber | 0.047500 | 0.044052 | -0.003448 | 0.9274 |
| src/compiler/types.ts:8739:1 | ImmediatelyInvokedFunctionExpression | 0.040805 | 0.037355 | -0.003450 | 0.9154 |
| src/compiler/moduleNameResolver.ts:143:1 | noPackageId | 0.037827 | 0.034370 | -0.003457 | 0.9086 |
| src/compiler/resolutionCache.ts:325:1 | canWatchAtTypes | 0.039931 | 0.036469 | -0.003461 | 0.9133 |
| src/compiler/utilities.ts:7858:1 | moveRangeEnd | 0.040577 | 0.037115 | -0.003462 | 0.9147 |
| src/compiler/builderPublic.ts:207:1 | createEmitAndSemanticDiagnosticsBuilderProgram | 0.039200 | 0.035735 | -0.003465 | 0.9116 |
| src/compiler/commandLineParser.ts:1831:1 | createOptionNameMap | 0.040135 | 0.036669 | -0.003466 | 0.9136 |
| src/compiler/types.ts:2328:1 | TypeOperatorNode | 0.038304 | 0.034812 | -0.003492 | 0.9088 |
| src/compiler/factory/emitHelpers.ts:1114:1 |  | 0.067002 | 0.063503 | -0.003499 | 0.9478 |
| src/compiler/types.ts:6110:1 | CheckFlags | 0.038182 | 0.034681 | -0.003501 | 0.9083 |
| src/compiler/_namespaces/ts.ts:6:1 |  | 0.041405 | 0.037901 | -0.003504 | 0.9154 |
| src/compiler/moduleNameResolver.ts:3354:1 | shouldAllowImportingTsExtension | 0.042343 | 0.038834 | -0.003509 | 0.9171 |
| src/compiler/types.ts:4990:1 | EmitTransformers | 0.038898 | 0.035372 | -0.003526 | 0.9094 |
| src/compiler/utilities.ts:2783:1 |  | 0.038574 | 0.035048 | -0.003526 | 0.9086 |
| src/compiler/checker.ts:54279:1 | NodeBuilderContext | 0.041612 | 0.038082 | -0.003530 | 0.9152 |
| src/compiler/factory/nodeTests.ts:280:1 | isCommaToken | 0.045313 | 0.041783 | -0.003530 | 0.9221 |
| src/compiler/factory/nodeTests.ts:239:1 | isBigIntLiteral | 0.042865 | 0.039328 | -0.003537 | 0.9175 |
| src/compiler/types.ts:6938:1 | ConditionalRoot | 0.037760 | 0.034216 | -0.003545 | 0.9061 |
| src/compiler/factory/utilities.ts:1413:1 | BinaryExpressionStateMachine | 0.044445 | 0.040899 | -0.003545 | 0.9202 |
| src/compiler/commandLineParser.ts:126:1 |  | 0.064602 | 0.061052 | -0.003550 | 0.9450 |
| src/compiler/utilitiesPublic.ts:2283:1 | canHaveLocals | 0.044572 | 0.041004 | -0.003568 | 0.9199 |
| src/compiler/program.ts:5139:1 | getResolutionDiagnostic | 0.126035 | 0.122446 | -0.003589 | 0.9715 |
| src/compiler/types.ts:5769:1 | AnyValidImportOrReExport | 0.039393 | 0.035789 | -0.003604 | 0.9085 |
| src/compiler/expressionToTypeNode.ts:1:1 |  | 0.039496 | 0.035882 | -0.003614 | 0.9085 |
| src/compiler/types.ts:2172:1 | ClassStaticBlockDeclaration | 0.039057 | 0.035442 | -0.003615 | 0.9074 |
| src/compiler/utilities.ts:5054:1 | isDeleteTarget | 0.052327 | 0.048699 | -0.003627 | 0.9307 |
| src/compiler/utilitiesPublic.ts:2321:1 | isDeclarationKind | 0.043300 | 0.039669 | -0.003631 | 0.9161 |
| src/compiler/types.ts:5011:1 | SourceMapEmitResult | 0.039669 | 0.036033 | -0.003636 | 0.9083 |
| src/compiler/performanceCore.ts:90:1 |  | 0.040107 | 0.036468 | -0.003639 | 0.9093 |
| src/compiler/utilities.ts:2661:1 | isVarAwaitUsing | 0.045152 | 0.041501 | -0.003651 | 0.9191 |
| src/compiler/factory/emitHelpers.ts:755:1 |  | 0.066700 | 0.063048 | -0.003652 | 0.9452 |
| src/compiler/utilities.ts:9316:1 |  | 0.078982 | 0.075321 | -0.003661 | 0.9536 |
| src/compiler/moduleNameResolver.ts:2064:1 | loadModuleFromFileNoPackageId | 0.041769 | 0.038097 | -0.003672 | 0.9121 |
| src/compiler/types.ts:2335:1 | UniqueTypeOperatorNode | 0.038552 | 0.034864 | -0.003688 | 0.9043 |
| src/compiler/types.ts:10421:1 | PragmaPseudoMap | 0.040445 | 0.036757 | -0.003688 | 0.9088 |
| src/compiler/types.ts:5789:1 | RequireVariableDeclarationList | 0.038617 | 0.034928 | -0.003689 | 0.9045 |
| src/compiler/types.ts:2555:1 | RelationalOperator | 0.041184 | 0.037494 | -0.003691 | 0.9104 |
| src/compiler/utilities.ts:10175:1 | removeFileExtension | 0.039263 | 0.035556 | -0.003706 | 0.9056 |
| src/compiler/scanner.ts:574:1 | isLineBreak | 0.038968 | 0.035258 | -0.003710 | 0.9048 |
| src/compiler/utilities.ts:3539:1 | nodeOrChildIsDecorated | 0.065677 | 0.061967 | -0.003711 | 0.9435 |
| src/compiler/utilities.ts:1015:1 | getStartPositionOfLine | 0.042454 | 0.038740 | -0.003715 | 0.9125 |
| src/compiler/utilitiesPublic.ts:2375:1 | isStatementKindButNotDeclarationKind | 0.042919 | 0.039199 | -0.003721 | 0.9133 |
| src/compiler/_namespaces/ts.ts:57:1 |  | 0.038368 | 0.034628 | -0.003740 | 0.9025 |
| src/compiler/utilities.ts:9314:1 |  | 0.078786 | 0.075041 | -0.003744 | 0.9525 |
| src/compiler/commandLineParser.ts:3309:1 | getSubstitutedPathWithConfigDirTemplate | 0.040417 | 0.036671 | -0.003745 | 0.9073 |
| src/compiler/utilities.ts:2717:1 | isImportMeta | 0.241156 | 0.237405 | -0.003751 | 0.9844 |
| src/compiler/utilities.ts:2740:1 | isHoistedFunction | 0.223081 | 0.219309 | -0.003772 | 0.9831 |
| src/compiler/factory/emitHelpers.ts:819:1 |  | 0.063109 | 0.059329 | -0.003780 | 0.9401 |
| src/compiler/semver.ts:207:1 | VersionRange | 0.042409 | 0.038628 | -0.003781 | 0.9108 |
| src/compiler/types.ts:2189:1 | TypeNode | 0.039172 | 0.035388 | -0.003784 | 0.9034 |
| src/compiler/types.ts:6042:1 | Symbol | 0.037903 | 0.034116 | -0.003787 | 0.9001 |
| src/compiler/types.ts:3951:1 | JSDoc | 0.038796 | 0.034992 | -0.003803 | 0.9020 |
| src/compiler/types.ts:6681:1 | VarianceFlags | 0.037628 | 0.033824 | -0.003804 | 0.8989 |
| src/compiler/utilities.ts:3090:1 | isThisTypePredicate | 0.049631 | 0.045812 | -0.003818 | 0.9231 |
| src/compiler/types.ts:3199:1 | JsxOpeningLikeElement | 0.044857 | 0.041033 | -0.003824 | 0.9147 |
| src/compiler/utilities.ts:10220:1 | ParsedPatterns | 0.042181 | 0.038355 | -0.003826 | 0.9093 |
| src/compiler/parser.ts:10616:1 | PragmaDiagnosticReporter | 0.038835 | 0.034994 | -0.003841 | 0.9011 |
| src/compiler/types.ts:2374:1 | PropertyNameLiteral | 0.038898 | 0.035040 | -0.003858 | 0.9008 |
| src/compiler/types.ts:2539:1 | AdditiveOperatorOrHigher | 0.039287 | 0.035428 | -0.003859 | 0.9018 |
| src/compiler/program.ts:387:1 | createCompilerHost | 0.042402 | 0.038540 | -0.003862 | 0.9089 |
| src/compiler/utilities.ts:2325:1 | getEnclosingContainer | 0.061483 | 0.057606 | -0.003877 | 0.9369 |
| src/compiler/performanceCore.ts:93:1 | tryGetNativePerformanceHooks | 0.039753 | 0.035875 | -0.003878 | 0.9025 |
| src/compiler/types.ts:5947:1 | SymbolFlags | 0.038900 | 0.035020 | -0.003880 | 0.9003 |
| src/compiler/commandLineParser.ts:1847:1 | getOptionsNameMap | 0.040444 | 0.036555 | -0.003890 | 0.9038 |
| src/compiler/types.ts:8323:1 | EmitNode | 0.041894 | 0.038003 | -0.003891 | 0.9071 |
| src/compiler/executeCommandLine.ts:1060:1 | enableSolutionPerformance | 0.038590 | 0.034697 | -0.003893 | 0.8991 |
| src/compiler/types.ts:7251:1 | RepopulateDiagnosticChainInfo | 0.038408 | 0.034511 | -0.003896 | 0.8985 |
| src/compiler/utilities.ts:5397:1 | getPropertyNameForPropertyNameNode | 0.275409 | 0.271512 | -0.003897 | 0.9859 |
| src/compiler/types.ts:5764:1 | ValidImportTypeNode | 0.039098 | 0.035195 | -0.003903 | 0.9002 |
| src/compiler/types.ts:2711:1 | BindingOrAssignmentElementTarget | 0.042201 | 0.038291 | -0.003910 | 0.9073 |
| src/compiler/transformers/declarations/diagnostics.ts:111:1 | DeclarationDiagnosticProducing | 0.039209 | 0.035267 | -0.003941 | 0.8995 |
| src/compiler/transformers/destructuring.ts:600:1 | makeArrayBindingPattern | 0.041532 | 0.037583 | -0.003949 | 0.9049 |
| src/compiler/utilities.ts:2287:1 | isAnyImportOrBareOrAccessedRequire | 0.287487 | 0.283535 | -0.003952 | 0.9863 |
| src/compiler/utilities.ts:9382:1 | getEmitStandardClassFields | 0.038333 | 0.034371 | -0.003962 | 0.8966 |
| src/compiler/visitorPublic.ts:623:1 |  | 0.065643 | 0.061679 | -0.003964 | 0.9396 |
| src/compiler/utilities.ts:3157:1 | getContainingClassExcludingClassDecorators | 0.281925 | 0.277961 | -0.003965 | 0.9859 |
| src/compiler/types.ts:2733:1 | BindingOrAssignmentPattern | 0.038904 | 0.034925 | -0.003979 | 0.8977 |
| src/compiler/factory/nodeTests.ts:496:1 | isNamedTupleMember | 0.044453 | 0.040462 | -0.003991 | 0.9102 |
| src/compiler/checker.ts:1174:1 | IterationUse | 0.039553 | 0.035556 | -0.003997 | 0.8989 |
| src/compiler/utilities.ts:6282:1 |  | 0.040454 | 0.036449 | -0.004004 | 0.9010 |
| src/compiler/types.ts:7317:1 | DiagnosticWithDetachedLocation | 0.038409 | 0.034402 | -0.004008 | 0.8957 |
| src/compiler/commandLineParser.ts:2136:1 | getOptionDeclarationFromName | 0.111915 | 0.107907 | -0.004008 | 0.9642 |
| src/compiler/utilitiesPublic.ts:1685:1 | isFunctionLikeDeclaration | 0.049158 | 0.045139 | -0.004019 | 0.9182 |
| src/compiler/utilities.ts:2784:1 |  | 0.039313 | 0.035289 | -0.004023 | 0.8977 |
| src/compiler/commandLineParser.ts:2317:1 | getWatchOptionsNameMap | 0.042304 | 0.038275 | -0.004029 | 0.9048 |
| src/compiler/factory/emitHelpers.ts:1344:1 |  | 0.070054 | 0.066021 | -0.004034 | 0.9424 |
| src/compiler/utilities.ts:2197:1 | isAmbientPropertyDeclaration | 0.044328 | 0.040285 | -0.004043 | 0.9088 |
| src/compiler/types.ts:4256:1 | FlowReduceLabel | 0.037994 | 0.033947 | -0.004047 | 0.8935 |
| src/compiler/types.ts:3489:1 | DefaultClause | 0.039873 | 0.035819 | -0.004053 | 0.8983 |
| src/compiler/path.ts:30:1 |  | 0.040474 | 0.036410 | -0.004064 | 0.8996 |
| src/compiler/types.ts:1821:1 | Decorator | 0.039644 | 0.035579 | -0.004065 | 0.8975 |
| src/compiler/program.ts:326:1 |  | 0.041499 | 0.037412 | -0.004087 | 0.9015 |
| src/compiler/types.ts:1875:1 | VariableDeclaration | 0.039169 | 0.035070 | -0.004099 | 0.8954 |
| src/compiler/tsbuildPublic.ts:569:1 | isParsedCommandLine | 0.040251 | 0.036151 | -0.004100 | 0.8981 |
| src/compiler/utilities.ts:9637:1 |  | 0.088421 | 0.084316 | -0.004105 | 0.9536 |
| src/compiler/types.ts:7248:1 | RepopulateModeMismatchDiagnosticChain | 0.038944 | 0.034829 | -0.004115 | 0.8943 |
| src/compiler/factory/nodeTests.ts:318:1 | isIdentifier | 0.046203 | 0.042088 | -0.004115 | 0.9109 |
| src/compiler/types.ts:2163:1 | AccessorDeclaration | 0.039458 | 0.035332 | -0.004126 | 0.8954 |
| src/compiler/_namespaces/ts.ts:18:1 |  | 0.040890 | 0.036755 | -0.004134 | 0.8989 |
| src/compiler/transformers/declarations/diagnostics.ts:138:1 | canProduceDiagnostics | 0.301416 | 0.297274 | -0.004142 | 0.9863 |
| src/compiler/moduleNameResolver.ts:127:1 | withPackageId | 0.039069 | 0.034927 | -0.004142 | 0.8940 |
| src/compiler/types.ts:10331:1 | UnionToIntersection | 0.041864 | 0.037721 | -0.004143 | 0.9010 |
| src/compiler/sourcemap.ts:662:1 | SourceMappedPosition | 0.039342 | 0.035198 | -0.004144 | 0.8947 |
| src/compiler/types.ts:1676:1 | ClassMemberModifier | 0.038742 | 0.034597 | -0.004145 | 0.8930 |
| src/compiler/types.ts:6155:1 | MappedSymbol | 0.039133 | 0.034985 | -0.004149 | 0.8940 |
| src/compiler/transformers/utilities.ts:179:1 | collectExternalModuleInfo | 0.042930 | 0.038779 | -0.004151 | 0.9033 |
| src/compiler/utilities.ts:5469:1 | isProtoSetter | 0.185945 | 0.181793 | -0.004151 | 0.9777 |
| src/compiler/utilities.ts:3866:1 | isVariableDeclarationInitializedToRequire | 0.046623 | 0.042470 | -0.004153 | 0.9109 |
| src/compiler/types.ts:1646:1 | StaticKeyword | 0.039138 | 0.034982 | -0.004156 | 0.8938 |
| src/compiler/utilities.ts:664:1 | isExternalModuleSymbol | 0.041919 | 0.037747 | -0.004172 | 0.9005 |
| src/compiler/types.ts:4273:1 | IncompleteType | 0.040714 | 0.036526 | -0.004188 | 0.8971 |
| src/compiler/types.ts:2198:1 | ImportTypeAssertionContainer | 0.040869 | 0.036665 | -0.004204 | 0.8971 |
| src/compiler/transformers/legacyDecorators.ts:1:1 |  | 0.040607 | 0.036402 | -0.004204 | 0.8965 |
| src/compiler/types.ts:7296:1 | DiagnosticArguments | 0.039358 | 0.035145 | -0.004213 | 0.8930 |
| src/compiler/commandLineParser.ts:2337:1 | getCommandLineTypeAcquisitionMap | 0.044825 | 0.040610 | -0.004215 | 0.9060 |
| src/compiler/utilities.ts:9994:1 | getSupportedExtensions | 0.041002 | 0.036787 | -0.004215 | 0.8972 |
| src/compiler/types.ts:2666:1 | ArrayDestructuringAssignment | 0.041160 | 0.036943 | -0.004218 | 0.8975 |
| src/compiler/utilities.ts:6181:1 | encodeUtf16EscapeSequence | 0.039307 | 0.035074 | -0.004233 | 0.8923 |
| src/compiler/utilities.ts:4941:1 | NodeWithPossibleHoistedDeclaration | 0.039721 | 0.035476 | -0.004245 | 0.8931 |
| src/compiler/factory/nodeTests.ts:367:1 | isOverrideModifier | 0.050499 | 0.046242 | -0.004257 | 0.9157 |
| src/compiler/types.ts:1827:1 | TypeParameterDeclaration | 0.040715 | 0.036457 | -0.004258 | 0.8954 |
| src/compiler/utilities.ts:9698:1 | getSubPatternFromSpec | 0.045746 | 0.041464 | -0.004282 | 0.9064 |
| src/compiler/commandLineParser.ts:2351:1 |  | 0.066357 | 0.062061 | -0.004296 | 0.9353 |
| src/compiler/utilities.ts:9580:1 | isNodeModulesOrScopedPackageDirectory | 0.039788 | 0.035490 | -0.004298 | 0.8920 |
| src/compiler/types.ts:2323:1 | ParenthesizedTypeNode | 0.039820 | 0.035520 | -0.004300 | 0.8920 |
| src/compiler/types.ts:1682:1 | ModifiersArray | 0.040474 | 0.036168 | -0.004306 | 0.8936 |
| src/compiler/utilities.ts:1209:1 | createCommentDirectivesMap | 0.039963 | 0.035653 | -0.004310 | 0.8922 |
| src/compiler/utilitiesPublic.ts:1204:1 | getJSDocTypeTag | 0.056233 | 0.051922 | -0.004312 | 0.9233 |
| src/compiler/utilities.ts:9354:1 | moduleSupportsImportAttributes | 0.040026 | 0.035712 | -0.004314 | 0.8922 |
| src/compiler/path.ts:69:1 | isDiskPathRoot | 0.043123 | 0.038809 | -0.004314 | 0.9000 |
| src/compiler/utilities.ts:5703:1 | OperatorPrecedence | 0.039653 | 0.035338 | -0.004315 | 0.8912 |
| src/compiler/factory/nodeTests.ts:418:1 | isPropertySignature | 0.047447 | 0.043125 | -0.004322 | 0.9089 |
| src/compiler/types.ts:2912:1 | TemplateLiteral | 0.049507 | 0.045177 | -0.004330 | 0.9125 |
| src/compiler/utilities.ts:9028:1 | usesWildcardTypes | 0.039766 | 0.035428 | -0.004338 | 0.8909 |
| src/compiler/factory/utilities.ts:1229:1 | isEqualityOperatorOrHigher | 0.115969 | 0.111627 | -0.004343 | 0.9626 |
| src/compiler/types.ts:2085:1 | FunctionDeclaration | 0.040354 | 0.036005 | -0.004348 | 0.8922 |
| src/compiler/watchUtilities.ts:103:1 | MutableFileSystemEntries | 0.040513 | 0.036151 | -0.004362 | 0.8923 |
| src/compiler/utilities.ts:3471:1 | nodeCanBeDecorated | 0.040406 | 0.036037 | -0.004369 | 0.8919 |
| src/compiler/types.ts:1361:1 | HasModifiers | 0.041583 | 0.037203 | -0.004379 | 0.8947 |
| src/compiler/types.ts:7270:1 | Diagnostic | 0.038590 | 0.034208 | -0.004382 | 0.8864 |
| src/compiler/types.ts:6911:1 | IndexedAccessType | 0.039017 | 0.034625 | -0.004392 | 0.8874 |
| src/compiler/commandLineParser.ts:2438:1 | AdamicJsonRecoveryValue | 0.039741 | 0.035337 | -0.004405 | 0.8892 |
| src/compiler/transformers/es2015.ts:289:1 | ConvertedLoopState | 0.040126 | 0.035722 | -0.004405 | 0.8902 |
| src/compiler/utilitiesPublic.ts:2597:1 | isTypeReferenceType | 0.046374 | 0.041966 | -0.004408 | 0.9049 |
| src/compiler/factory/nodeTests.ts:624:1 | isTypeOfExpression | 0.045267 | 0.040857 | -0.004410 | 0.9026 |
| src/compiler/utilities.ts:2745:1 | isHoistedVariable | 0.254966 | 0.250555 | -0.004411 | 0.9827 |
| src/compiler/utilitiesPublic.ts:636:1 | walkUpBindingElementsAndPatterns | 0.187039 | 0.182628 | -0.004411 | 0.9764 |
| src/compiler/semver.ts:336:1 | parseComparator | 0.202498 | 0.198079 | -0.004419 | 0.9782 |
| src/compiler/types.ts:1391:1 | HasIllegalModifiers | 0.042265 | 0.037838 | -0.004427 | 0.8953 |
| src/compiler/types.ts:1849:1 | SignatureDeclaration | 0.039959 | 0.035531 | -0.004429 | 0.8892 |
| src/compiler/types.ts:7241:1 | RepopulateModuleNotFoundDiagnosticChain | 0.039154 | 0.034701 | -0.004453 | 0.8863 |
| src/compiler/types.ts:3308:1 | JsxText | 0.040949 | 0.036486 | -0.004463 | 0.8910 |
| src/compiler/core.ts:1819:1 | returnUndefined | 0.041315 | 0.036840 | -0.004475 | 0.8917 |
| src/compiler/moduleNameResolver.ts:427:1 | readPackageJsonTypesVersionPaths | 0.040651 | 0.036173 | -0.004478 | 0.8899 |
| src/compiler/utilities.ts:803:1 | projectReferenceIsEqualTo | 4.512026 | 4.507534 | -0.004492 | 0.9990 |
| src/compiler/types.ts:2406:1 | PartiallyEmittedExpression | 0.041398 | 0.036906 | -0.004492 | 0.8915 |
| src/compiler/program.ts:1038:1 | getModuleResolutionName | 0.044778 | 0.040277 | -0.004501 | 0.8995 |
| src/compiler/moduleNameResolver.ts:422:1 | VersionPaths | 0.039001 | 0.034500 | -0.004501 | 0.8846 |
| src/compiler/commandLineParser.ts:3313:1 | getSubstitutedStringArrayWithConfigDirTemplate | 0.039767 | 0.035262 | -0.004505 | 0.8867 |
| src/compiler/sourcemap.ts:636:1 | base64FormatEncode | 0.041968 | 0.037461 | -0.004507 | 0.8926 |
| src/compiler/types.ts:10334:1 | ArgumentDefinitionToFieldUnion | 0.042311 | 0.037791 | -0.004520 | 0.8932 |
| src/compiler/utilities.ts:8966:1 | isFileModuleFromUsingJSXTag | 0.050957 | 0.046430 | -0.004527 | 0.9112 |
| src/compiler/utilities.ts:6832:1 | getAllAccessorDeclarations | 0.156703 | 0.152160 | -0.004543 | 0.9710 |
| src/compiler/utilities.ts:8285:1 | isUMDExportSymbol | 0.276138 | 0.271581 | -0.004556 | 0.9835 |
| src/compiler/_namespaces/ts.ts:8:1 |  | 0.041332 | 0.036769 | -0.004563 | 0.8896 |
| src/compiler/types.ts:6201:1 | __String | 0.039948 | 0.035375 | -0.004573 | 0.8855 |
| src/compiler/symbolWalker.ts:1:1 |  | 0.041646 | 0.037073 | -0.004574 | 0.8902 |
| src/compiler/moduleNameResolver.ts:397:1 | readPackageJsonTypesFields | 0.040202 | 0.035627 | -0.004575 | 0.8862 |
| src/compiler/watchUtilities.ts:501:1 | WildcardDirectoryWatcher | 0.040236 | 0.035650 | -0.004586 | 0.8860 |
| src/compiler/utilitiesPublic.ts:1224:1 | getJSDocType | 0.056997 | 0.052405 | -0.004592 | 0.9194 |
| src/compiler/types.ts:6204:1 | ReadonlyUnderscoreEscapedMap | 0.039504 | 0.034911 | -0.004592 | 0.8838 |
| src/compiler/types.ts:6210:1 | SymbolTable | 0.039855 | 0.035261 | -0.004594 | 0.8847 |
| src/compiler/semver.ts:314:1 | parseHyphen | 0.043463 | 0.038861 | -0.004603 | 0.8941 |
| src/compiler/types.ts:9910:1 | SourceMapGenerator | 0.040193 | 0.035578 | -0.004615 | 0.8852 |
| src/compiler/types.ts:7301:1 | DiagnosticRelatedInformation | 0.038692 | 0.034058 | -0.004635 | 0.8802 |
| src/compiler/types.ts:6734:1 | UnionOrIntersectionType | 0.038827 | 0.034154 | -0.004672 | 0.8797 |
| src/compiler/tsbuildPublic.ts:331:1 | getCompilerOptionsOfBuildOptions | 0.040898 | 0.036217 | -0.004681 | 0.8855 |
| src/compiler/types.ts:7331:1 | diagnosticCategoryName | 0.040391 | 0.035706 | -0.004686 | 0.8840 |
| src/compiler/utilities.ts:6164:1 |  | 0.066905 | 0.062212 | -0.004692 | 0.9299 |
| src/compiler/utilities.ts:10751:1 | isPackedElement | 0.280448 | 0.275755 | -0.004693 | 0.9833 |
| src/compiler/utilities.ts:10056:1 | usesExtensionsOnImports | 0.077620 | 0.072926 | -0.004694 | 0.9395 |
| src/compiler/types.ts:2026:1 | VariableLikeDeclaration | 0.039229 | 0.034521 | -0.004708 | 0.8800 |
| src/compiler/_namespaces/ts.ts:35:1 |  | 0.045083 | 0.040357 | -0.004726 | 0.8952 |
| src/compiler/factory/nodeTests.ts:628:1 | isVoidExpression | 0.045080 | 0.040354 | -0.004726 | 0.8952 |
| src/compiler/parser.ts:9938:1 | intersectsIncrementalChange | 0.043309 | 0.038577 | -0.004733 | 0.8907 |
| src/compiler/types.ts:6498:1 | UniqueESSymbolType | 0.039766 | 0.035014 | -0.004753 | 0.8805 |
| src/compiler/types.ts:8683:1 | ParenthesizerRules | 0.040023 | 0.035268 | -0.004755 | 0.8812 |
| src/compiler/commandLineParser.ts:1727:1 |  | 0.041077 | 0.036314 | -0.004763 | 0.8841 |
| src/compiler/transformers/declarations.ts:1891:1 | canHaveLiteralInitializer | 0.055329 | 0.050561 | -0.004768 | 0.9138 |
| src/compiler/types.ts:8350:1 | TabStop | 0.042230 | 0.037456 | -0.004774 | 0.8870 |
| src/compiler/scanner.ts:4036:1 | codePointAt | 0.042194 | 0.037409 | -0.004785 | 0.8866 |
| src/compiler/utilitiesPublic.ts:2117:1 | isIterationStatement | 0.122357 | 0.117572 | -0.004785 | 0.9609 |
| src/compiler/watch.ts:146:1 | clearScreenIfNotWatchingForFileChanges | 0.111639 | 0.106852 | -0.004787 | 0.9571 |
| src/compiler/types.ts:1993:1 | PropertyAssignment | 0.040917 | 0.036129 | -0.004788 | 0.8830 |
| src/compiler/moduleNameResolver.ts:3408:1 | SearchResult | 0.040696 | 0.035900 | -0.004795 | 0.8822 |
| src/compiler/builder.ts:1145:1 | IncrementalMultiFileEmitBuildInfo | 0.039857 | 0.035055 | -0.004802 | 0.8795 |
| src/compiler/commandLineParser.ts:2426:1 | JsonConversionNotifier | 0.041311 | 0.036484 | -0.004827 | 0.8832 |
| src/compiler/types.ts:6974:1 | StringMappingType | 0.040100 | 0.035263 | -0.004837 | 0.8794 |
| src/compiler/utilities.ts:4812:1 | getJSDocHost | 0.041199 | 0.036330 | -0.004869 | 0.8818 |
| src/compiler/expressionToTypeNode.ts:171:1 |  | 0.047785 | 0.042908 | -0.004877 | 0.8979 |
| src/compiler/performance.ts:153:1 | clearMarks | 0.077693 | 0.072792 | -0.004900 | 0.9369 |
| src/compiler/commandLineParser.ts:3043:1 | isNullOrUndefined | 0.040303 | 0.035367 | -0.004936 | 0.8775 |
| src/compiler/builder.ts:1621:1 | computeSignatureWithDiagnostics | 4.452290 | 4.447341 | -0.004949 | 0.9989 |
| src/compiler/utilities.ts:10417:1 | skipTypeCheckingWorker | 0.113770 | 0.108817 | -0.004953 | 0.9565 |
| src/compiler/core.ts:1256:1 |  | 0.043466 | 0.038474 | -0.004992 | 0.8852 |
| src/compiler/types.ts:2544:1 | ShiftOperator | 0.041116 | 0.036122 | -0.004994 | 0.8785 |
| src/compiler/utilitiesPublic.ts:2118:1 | isIterationStatement | 0.123087 | 0.118076 | -0.005011 | 0.9593 |
| src/compiler/visitorPublic.ts:543:1 | visitIterationBody | 0.046980 | 0.041952 | -0.005028 | 0.8930 |
| src/compiler/factory/emitHelpers.ts:1467:1 |  | 0.070512 | 0.065481 | -0.005031 | 0.9287 |
| src/compiler/types.ts:3430:1 | ForInStatement | 0.040906 | 0.035863 | -0.005043 | 0.8767 |
| src/compiler/types.ts:7006:1 | SignatureFlags | 0.039216 | 0.034160 | -0.005056 | 0.8711 |
| src/compiler/_namespaces/ts.ts:54:1 |  | 0.040394 | 0.035334 | -0.005060 | 0.8747 |
| src/compiler/commandLineParser.ts:2637:1 | ConvertToTSConfigHost | 0.040855 | 0.035778 | -0.005077 | 0.8757 |
| src/compiler/semver.ts:27:1 |  | 0.045822 | 0.040743 | -0.005080 | 0.8891 |
| src/compiler/types.ts:6217:1 | PatternAmbientModule | 0.039611 | 0.034509 | -0.005102 | 0.8712 |
| src/compiler/moduleNameResolver.ts:509:1 |  | 0.039694 | 0.034587 | -0.005107 | 0.8714 |
| src/compiler/types.ts:3185:1 | ImportDeferProperty | 0.044403 | 0.039287 | -0.005116 | 0.8848 |
| src/compiler/types.ts:7722:1 | WatchDirectoryFlags | 0.046533 | 0.041416 | -0.005117 | 0.8900 |
| src/compiler/types.ts:1948:1 | PrivateIdentifierPropertyDeclaration | 0.039572 | 0.034454 | -0.005117 | 0.8707 |
| src/compiler/moduleSpecifiers.ts:1140:1 | tryGetModuleNameFromPackageJsonImports | 0.117651 | 0.112527 | -0.005123 | 0.9565 |
| src/compiler/transformers/es2015.ts:367:1 | LoopConverter | 0.041050 | 0.035909 | -0.005141 | 0.8748 |
| src/compiler/types.ts:1868:1 | ConstructSignatureDeclaration | 0.039296 | 0.034153 | -0.005143 | 0.8691 |
| src/compiler/utilities.ts:9361:1 | StrictOptionName | 0.040588 | 0.035444 | -0.005143 | 0.8733 |
| src/compiler/watch.ts:114:1 |  | 0.075311 | 0.070166 | -0.005145 | 0.9317 |
| src/compiler/utilities.ts:3527:1 | nodeIsDecorated | 0.054357 | 0.049193 | -0.005164 | 0.9050 |
| src/compiler/types.ts:2460:1 | NullLiteral | 0.041404 | 0.036230 | -0.005174 | 0.8750 |
| src/compiler/commandLineParser.ts:1858:1 |  | 0.039866 | 0.034675 | -0.005191 | 0.8698 |
| src/compiler/tsbuildPublic.ts:205:1 | getCurrentTime | 0.109175 | 0.103976 | -0.005198 | 0.9524 |
| src/compiler/_namespaces/ts.ts:24:1 |  | 0.041796 | 0.036590 | -0.005206 | 0.8754 |
| src/compiler/builder.ts:732:1 | getPendingEmitKindWithSeen | 0.042578 | 0.037352 | -0.005226 | 0.8773 |
| src/compiler/utilities.ts:3898:1 | isVariableDeclarationInitializedWithRequireHelper | 0.308561 | 0.303335 | -0.005226 | 0.9831 |
| src/compiler/utilities.ts:3475:1 | nodeCanBeDecorated | 0.040716 | 0.035480 | -0.005237 | 0.8714 |
| src/compiler/types.ts:7336:1 | ModuleResolutionKind | 0.040245 | 0.034998 | -0.005246 | 0.8696 |
| src/compiler/utilities.ts:725:1 | optionsHaveChanges | 0.042187 | 0.036939 | -0.005248 | 0.8756 |
| src/compiler/types.ts:2373:1 | StringLiteralLike | 0.041012 | 0.035753 | -0.005259 | 0.8718 |
| src/compiler/types.ts:1840:1 | SignatureDeclarationBase | 0.039667 | 0.034406 | -0.005262 | 0.8673 |
| src/compiler/transformers/es2015.ts:272:1 | LoopOutParameterFlags | 0.040767 | 0.035495 | -0.005271 | 0.8707 |
| src/compiler/types.ts:7030:1 | Signature | 0.039820 | 0.034546 | -0.005275 | 0.8675 |
| src/compiler/emitter.ts:1203:1 |  | 0.041831 | 0.036553 | -0.005278 | 0.8738 |
| src/compiler/builder.ts:509:1 | getEmitSignatureFromOldSignature | 0.067504 | 0.062210 | -0.005294 | 0.9216 |
| src/compiler/types.ts:6171:1 | InternalSymbolName | 0.040236 | 0.034926 | -0.005310 | 0.8680 |
| src/compiler/performance.ts:141:1 | forEachMark | 0.039838 | 0.034518 | -0.005320 | 0.8665 |
| src/compiler/types.ts:7290:1 | CanonicalDiagnostic | 0.039624 | 0.034301 | -0.005324 | 0.8656 |
| src/compiler/visitorPublic.ts:445:1 | addDefaultValueAssignmentForBindingPattern | 0.254203 | 0.248861 | -0.005342 | 0.9790 |
| src/compiler/program.ts:328:1 | findConfigFile | 0.043974 | 0.038597 | -0.005377 | 0.8777 |
| src/compiler/factory/utilities.ts:1208:1 | isRelationalOperator | 0.045827 | 0.040449 | -0.005377 | 0.8827 |
| src/compiler/types.ts:8079:1 | Extension | 0.043989 | 0.038600 | -0.005390 | 0.8775 |
| src/compiler/factory/baseNodeFactory.ts:13:1 | BaseNodeFactory | 0.043018 | 0.037624 | -0.005394 | 0.8746 |
| src/compiler/utilities.ts:4767:1 | getEffectiveContainerForJSDocTemplateTag | 0.291610 | 0.286186 | -0.005424 | 0.9814 |
| src/compiler/utilities.ts:8374:1 | getLeftmostAccessExpression | 0.253196 | 0.247768 | -0.005429 | 0.9786 |
| src/compiler/utilities.ts:5276:1 | isKeywordOrPunctuation | 0.111519 | 0.106055 | -0.005464 | 0.9510 |
| src/compiler/factory/nodeTests.ts:265:1 | isTemplateMiddle | 0.046458 | 0.040976 | -0.005481 | 0.8820 |
| src/compiler/transformers/destructuring.ts:386:1 | flattenObjectBindingOrAssignmentPattern | 0.531892 | 0.526400 | -0.005493 | 0.9897 |
| src/compiler/factory/emitHelpers.ts:1221:1 |  | 0.071042 | 0.065542 | -0.005500 | 0.9226 |
| src/compiler/program.ts:1021:1 | ResolutionWithResolvedFileName | 0.041984 | 0.036483 | -0.005501 | 0.8690 |
| src/compiler/tsbuildPublic.ts:132:1 |  | 0.040119 | 0.034616 | -0.005503 | 0.8628 |
| src/compiler/types.ts:6224:1 | NodeCheckFlags | 0.039751 | 0.034243 | -0.005508 | 0.8614 |
| src/compiler/utilities.ts:3473:1 | nodeCanBeDecorated | 0.041402 | 0.035872 | -0.005531 | 0.8664 |
| src/compiler/utilities.ts:9017:1 | importSyntaxAffectsModuleResolution | 0.039884 | 0.034349 | -0.005535 | 0.8612 |
| src/compiler/builderStatePublic.ts:1:1 |  | 0.043203 | 0.037664 | -0.005539 | 0.8718 |
| src/compiler/utilitiesPublic.ts:1680:1 | isFunctionLikeOrClassStaticBlockDeclaration | 0.242090 | 0.236550 | -0.005540 | 0.9771 |
| src/compiler/types.ts:6160:1 | ReverseMappedSymbolLinks | 0.040612 | 0.035059 | -0.005553 | 0.8633 |
| src/compiler/tsbuildPublic.ts:209:1 | ReportEmitErrorSummary | 0.042664 | 0.037108 | -0.005556 | 0.8698 |
| src/compiler/program.ts:1:1 |  | 0.047401 | 0.041833 | -0.005568 | 0.8825 |
| src/compiler/utilities.ts:2061:1 | makeIdentifierFromModuleName | 0.040249 | 0.034672 | -0.005576 | 0.8615 |
| src/compiler/commandLineParser.ts:1851:1 |  | 0.065943 | 0.060348 | -0.005595 | 0.9152 |
| src/compiler/factory/nodeTests.ts:412:1 | isDecorator | 0.046178 | 0.040564 | -0.005613 | 0.8784 |
| src/compiler/transformers/utilities.ts:173:1 | getImportNeedsImportDefaultHelper | 0.282928 | 0.277301 | -0.005627 | 0.9801 |
| src/compiler/moduleNameResolver.ts:211:1 | PathAndPackageId | 0.040953 | 0.035306 | -0.005647 | 0.8621 |
| src/compiler/binder.ts:350:1 | getModuleInstanceStateWorker | 0.161308 | 0.155652 | -0.005656 | 0.9649 |
| src/compiler/_namespaces/ts.ts:19:1 |  | 0.041949 | 0.036292 | -0.005656 | 0.8652 |
| src/compiler/commandLineParser.ts:1868:1 | createCompilerDiagnosticForInvalidCustomType | 0.042006 | 0.036331 | -0.005675 | 0.8649 |
| src/compiler/sys.ts:941:1 | createFsWatchCallbackForDirectoryWatcherCallback | 0.046889 | 0.041205 | -0.005684 | 0.8788 |
| src/compiler/program.ts:1323:1 | getConfigFileParsingDiagnostics | 0.044099 | 0.038409 | -0.005691 | 0.8710 |
| src/compiler/types.ts:2053:1 | ArrayBindingElement | 0.041140 | 0.035448 | -0.005692 | 0.8616 |
| src/compiler/utilitiesPublic.ts:361:1 | textRangeContainsPositionInclusive | 0.041582 | 0.035883 | -0.005699 | 0.8629 |
| src/compiler/factory/utilities.ts:411:1 | createExpressionForPropertyAssignment | 0.092297 | 0.086595 | -0.005702 | 0.9382 |
| src/compiler/resolutionCache.ts:510:1 | createModuleResolutionLoaderUsingGlobalCache | 0.055122 | 0.049392 | -0.005730 | 0.8961 |
| src/compiler/commandLineParser.ts:2316:1 |  | 0.043230 | 0.037478 | -0.005752 | 0.8669 |
| src/compiler/types.ts:1817:1 | LateBoundName | 0.042248 | 0.036482 | -0.005767 | 0.8635 |
| src/compiler/expressionToTypeNode.ts:160:1 | SyntacticResult | 0.040641 | 0.034872 | -0.005769 | 0.8581 |
| src/compiler/utilities.ts:5001:1 | walkUp | 0.050169 | 0.044362 | -0.005807 | 0.8843 |
| src/compiler/types.ts:7405:1 | PollingWatchKind | 0.041941 | 0.036134 | -0.005808 | 0.8615 |
| src/compiler/utilities.ts:9621:1 |  | 0.089533 | 0.083723 | -0.005810 | 0.9351 |
| src/compiler/factory/utilities.ts:1120:1 | canHaveIllegalDecorators | 0.046864 | 0.041039 | -0.005825 | 0.8757 |
| src/compiler/types.ts:6523:1 | ObjectFlags | 0.040854 | 0.035021 | -0.005833 | 0.8572 |
| src/compiler/utilities.ts:9320:1 |  | 0.078088 | 0.072247 | -0.005841 | 0.9252 |
| src/compiler/utilities.ts:2786:1 |  | 0.041777 | 0.035903 | -0.005874 | 0.8594 |
| src/compiler/utilities.ts:8330:1 | isTypeNodeKind | 0.047054 | 0.041176 | -0.005878 | 0.8751 |
| src/compiler/_namespaces/ts.ts:59:1 |  | 0.042099 | 0.036217 | -0.005883 | 0.8603 |
| src/compiler/core.ts:1801:1 | returnFalse | 0.044566 | 0.038672 | -0.005895 | 0.8677 |
| src/compiler/expressionToTypeNode.ts:165:1 | syntacticResult | 0.049846 | 0.043939 | -0.005907 | 0.8815 |
| src/compiler/moduleNameResolver.ts:1:1 |  | 0.040563 | 0.034655 | -0.005908 | 0.8544 |
| src/compiler/factory/nodeTests.ts:1164:1 | isJSDocTemplateTag | 0.047276 | 0.041347 | -0.005929 | 0.8746 |
| src/compiler/transformers/destructuring.ts:208:1 | bindingOrAssignmentPatternAssignsToName | 0.043838 | 0.037904 | -0.005933 | 0.8647 |
| src/compiler/utilities.ts:9300:1 |  | 0.095009 | 0.089059 | -0.005950 | 0.9374 |
| src/compiler/utilities.ts:6262:1 | stripQuotes | 0.041391 | 0.035424 | -0.005967 | 0.8558 |
| src/compiler/factory/utilities.ts:1167:1 | isModuleName | 0.252040 | 0.246070 | -0.005970 | 0.9763 |
| src/compiler/sourcemap.ts:28:1 |  | 0.041651 | 0.035669 | -0.005982 | 0.8564 |
| src/compiler/types.ts:10039:1 | ModuleSpecifierOptions | 0.043994 | 0.037980 | -0.006014 | 0.8633 |
| src/compiler/types.ts:6999:1 | SignatureKind | 0.039795 | 0.033776 | -0.006019 | 0.8487 |
| src/compiler/builderStatePublic.ts:12:1 | OutputFile | 0.041980 | 0.035933 | -0.006047 | 0.8560 |
| src/compiler/commandLineParser.ts:3335:1 | isErrorNoInputFiles | 0.044305 | 0.038244 | -0.006061 | 0.8632 |
| src/compiler/types.ts:2020:1 | SpreadAssignment | 0.040098 | 0.034030 | -0.006068 | 0.8487 |
| src/compiler/commandLineParser.ts:1692:1 |  | 0.041964 | 0.035894 | -0.006070 | 0.8553 |
| src/compiler/utilitiesPublic.ts:2638:1 | isRestParameter | 0.192471 | 0.186389 | -0.006081 | 0.9684 |
| src/compiler/types.ts:6063:1 | SymbolLinks | 0.040526 | 0.034443 | -0.006083 | 0.8499 |
| src/compiler/transformers/jsx.ts:683:1 |  | 0.067523 | 0.061424 | -0.006098 | 0.9097 |
| src/compiler/moduleSpecifiers.ts:1020:1 | tryGetModuleNameFromExportsOrImports | 0.106133 | 0.100002 | -0.006131 | 0.9422 |
| src/compiler/utilities.ts:11468:1 | NameResolverOptions | 0.044141 | 0.038008 | -0.006132 | 0.8611 |
| src/compiler/utilities.ts:8951:1 | getLanguageVariant | 0.042173 | 0.036010 | -0.006163 | 0.8539 |
| src/compiler/types.ts:2680:1 | ObjectBindingOrAssignmentElement | 0.045846 | 0.039674 | -0.006172 | 0.8654 |
| src/compiler/utilities.ts:9318:1 |  | 0.077612 | 0.071434 | -0.006177 | 0.9204 |
| src/compiler/transformers/destructuring.ts:197:1 | bindingOrAssignmentElementAssignsToName | 0.333799 | 0.327622 | -0.006177 | 0.9815 |
| src/compiler/factory/emitHelpers.ts:1020:1 |  | 0.068199 | 0.062013 | -0.006186 | 0.9093 |
| src/compiler/builder.ts:567:1 | convertToDiagnostics | 0.156455 | 0.150243 | -0.006213 | 0.9603 |
| src/compiler/factory/utilities.ts:660:1 | skipOuterExpressions | 0.044181 | 0.037940 | -0.006242 | 0.8587 |
| src/compiler/types.ts:1980:1 | ObjectLiteralElement | 0.040614 | 0.034359 | -0.006255 | 0.8460 |
| src/compiler/moduleNameResolver.ts:339:1 | PackageJsonPathFields | 0.043650 | 0.037388 | -0.006262 | 0.8565 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts:1:1 |  | 0.043086 | 0.036819 | -0.006267 | 0.8545 |
| src/compiler/commandLineParser.ts:1844:1 |  | 0.041553 | 0.035220 | -0.006333 | 0.8476 |
| src/compiler/scanner.ts:552:1 | isWhiteSpaceLike | 0.042117 | 0.035764 | -0.006353 | 0.8492 |
| src/compiler/commandLineParser.ts:3029:1 | parseJsonSourceFileConfigFileContent | 0.090857 | 0.084501 | -0.006355 | 0.9300 |
| src/compiler/moduleNameResolver.ts:402:1 | readPackageJsonTSConfigField | 0.042187 | 0.035789 | -0.006398 | 0.8483 |
| src/compiler/types.ts:7310:1 | DiagnosticWithLocation | 0.040572 | 0.034153 | -0.006420 | 0.8418 |
| src/compiler/builder.ts:268:1 | isBuilderProgramStateWithDefinedProgram | 0.054735 | 0.048307 | -0.006428 | 0.8826 |
| src/compiler/factory/utilities.ts:1586:1 | formatGeneratedName | 0.045135 | 0.038685 | -0.006450 | 0.8571 |
| src/compiler/utilitiesPublic.ts:2043:1 | isUnaryExpressionKind | 0.047125 | 0.040666 | -0.006459 | 0.8629 |
| src/compiler/types.ts:6481:1 | FreshableType | 0.041964 | 0.035493 | -0.006471 | 0.8458 |
| src/compiler/types.ts:6207:1 | UnderscoreEscapedMap | 0.042151 | 0.035676 | -0.006475 | 0.8464 |
| src/compiler/utilities.ts:7819:1 | directoryProbablyExists | 0.107778 | 0.101301 | -0.006477 | 0.9399 |
| src/compiler/tracing.ts:24:1 |  | 0.044079 | 0.037562 | -0.006517 | 0.8522 |
| src/compiler/program.ts:877:1 | getModeForResolutionAtIndex | 0.058191 | 0.051649 | -0.006542 | 0.8876 |
| src/compiler/types.ts:5045:1 | TypeCheckerHost | 0.042344 | 0.035799 | -0.006545 | 0.8454 |
| src/compiler/core.ts:1794:1 | noop | 0.043847 | 0.037280 | -0.006566 | 0.8502 |
| src/compiler/types.ts:4635:1 | ReferencedFileKind | 0.053689 | 0.047107 | -0.006582 | 0.8774 |
| src/compiler/core.ts:1810:1 | returnTrue | 0.045868 | 0.039285 | -0.006583 | 0.8565 |
| src/compiler/factory/baseNodeFactory.ts:26:1 | createBaseNodeFactory | 0.087696 | 0.081102 | -0.006594 | 0.9248 |
| src/compiler/moduleSpecifiers.ts:1364:1 | processEnding | 0.044331 | 0.037734 | -0.006597 | 0.8512 |
| src/compiler/utilities.ts:2729:1 | isPrologueDirective | 0.123654 | 0.117055 | -0.006599 | 0.9466 |
| src/compiler/factory/utilities.ts:631:1 | isOuterExpression | 0.268952 | 0.262299 | -0.006652 | 0.9753 |
| src/compiler/utilities.ts:9326:1 |  | 0.077738 | 0.071079 | -0.006658 | 0.9144 |
| src/compiler/utilities.ts:10327:1 | matchPatternOrExact | 0.115625 | 0.108966 | -0.006659 | 0.9424 |
| src/compiler/factory/nodeTests.ts:261:1 | isTemplateHead | 0.047805 | 0.041143 | -0.006663 | 0.8606 |
| src/compiler/commandLineParser.ts:2320:1 |  | 0.071472 | 0.064764 | -0.006708 | 0.9061 |
| src/compiler/commandLineParser.ts:1915:1 | BaseParsedCommandLine | 0.041868 | 0.035152 | -0.006716 | 0.8396 |
| src/compiler/transformers/utilities.ts:871:1 | isSimpleParameterList | 0.044774 | 0.037998 | -0.006777 | 0.8486 |
| src/compiler/factory/emitHelpers.ts:1033:1 |  | 0.070972 | 0.064194 | -0.006778 | 0.9045 |
| src/compiler/utilities.ts:710:1 | changesAffectModuleResolution | 0.043666 | 0.036866 | -0.006801 | 0.8443 |
| src/compiler/factory/emitHelpers.ts:744:1 |  | 0.071906 | 0.065069 | -0.006837 | 0.9049 |
| src/compiler/transformers/generators.ts:299:1 | WithBlock | 0.044083 | 0.037216 | -0.006867 | 0.8442 |
| src/compiler/types.ts:2376:1 | TemplateLiteralTypeNode | 0.041380 | 0.034508 | -0.006872 | 0.8339 |
| src/compiler/performanceCore.ts:102:1 |  | 0.077795 | 0.070918 | -0.006877 | 0.9116 |
| src/compiler/resolutionCache.ts:258:1 | removeIgnoredPath | 0.073536 | 0.066652 | -0.006885 | 0.9064 |
| src/compiler/utilitiesPublic.ts:1765:1 | isMethodOrAccessor | 0.047852 | 0.040967 | -0.006885 | 0.8561 |
| src/compiler/types.ts:4228:1 | FlowCondition | 0.043044 | 0.036157 | -0.006887 | 0.8400 |
| src/compiler/transformers/es2019.ts:1:1 |  | 0.042080 | 0.035185 | -0.006895 | 0.8361 |
| src/compiler/factory/utilities.ts:1551:1 | formatGeneratedNamePart | 0.045182 | 0.038246 | -0.006936 | 0.8465 |
| src/compiler/factory/nodeTests.ts:596:1 | isNewExpression | 0.047644 | 0.040702 | -0.006942 | 0.8543 |
| src/compiler/moduleNameResolver.ts:1738:1 | nodeNextModuleNameResolverWorker | 0.046508 | 0.039530 | -0.006978 | 0.8500 |
| src/compiler/commandLineParser.ts:637:1 |  | 0.068523 | 0.061535 | -0.006988 | 0.8980 |
| src/compiler/factory/nodeTests.ts:710:1 | isSemicolonClassElement | 0.049626 | 0.042623 | -0.007003 | 0.8589 |
| src/compiler/core.ts:311:1 | clear | 0.048973 | 0.041970 | -0.007003 | 0.8570 |
| src/compiler/moduleNameResolver.ts:1374:1 | createTypeReferenceDirectiveResolutionCache | 0.223456 | 0.216436 | -0.007020 | 0.9686 |
| src/compiler/factory/nodeTests.ts:255:1 | isNoSubstitutionTemplateLiteral | 0.046324 | 0.039301 | -0.007023 | 0.8484 |
| src/compiler/factory/nodeTests.ts:760:1 | isBreakStatement | 0.058073 | 0.051045 | -0.007028 | 0.8790 |
| src/compiler/transformers/declarations.ts:1877:1 | maskModifierFlags | 0.050473 | 0.043441 | -0.007031 | 0.8607 |
| src/compiler/types.ts:6924:1 | IndexFlags | 0.041995 | 0.034946 | -0.007049 | 0.8321 |
| src/compiler/types.ts:2534:1 | AdditiveOperator | 0.042577 | 0.035527 | -0.007050 | 0.8344 |
| src/compiler/utilitiesPublic.ts:1156:1 | getJSDocReadonlyTag | 0.058832 | 0.051770 | -0.007062 | 0.8800 |
| src/compiler/types.ts:3327:1 | NotEmittedStatement | 0.043296 | 0.036233 | -0.007064 | 0.8369 |
| src/compiler/program.ts:1147:1 | getLibraryNameFromLibFileName | 0.044561 | 0.037495 | -0.007066 | 0.8414 |
| src/compiler/types.ts:6752:1 | UnionType | 0.042256 | 0.035179 | -0.007076 | 0.8325 |
| src/compiler/utilitiesPublic.ts:1819:1 | isFunctionOrConstructorTypeNode | 0.047119 | 0.040038 | -0.007081 | 0.8497 |
| src/compiler/parser.ts:10714:1 | extractPragmas | 0.042890 | 0.035797 | -0.007093 | 0.8346 |
| src/compiler/types.ts:2151:1 | SetAccessorDeclaration | 0.043107 | 0.036008 | -0.007100 | 0.8353 |
| src/compiler/program.ts:1228:1 | isProgramUptoDate | 0.092457 | 0.085352 | -0.007105 | 0.9232 |
| src/compiler/factory/parenthesizerRules.ts:1:1 |  | 0.042994 | 0.035883 | -0.007111 | 0.8346 |
| src/compiler/_namespaces/ts.ts:45:1 |  | 0.043277 | 0.036115 | -0.007162 | 0.8345 |
| src/compiler/factory/emitHelpers.ts:846:1 |  | 0.068925 | 0.061753 | -0.007172 | 0.8959 |
| src/compiler/transformers/namedEvaluation.ts:179:1 | injectClassNamedEvaluationHelperBlockIfMissing | 0.044966 | 0.037781 | -0.007185 | 0.8402 |
| src/compiler/utilities.ts:5229:1 | getAllSuperTypeNodes | 0.268338 | 0.261136 | -0.007202 | 0.9732 |
| src/compiler/factory/utilities.ts:764:1 | getOrCreateExternalHelpersModuleNameIfNeeded | 0.090162 | 0.082938 | -0.007224 | 0.9199 |
| src/compiler/utilities.ts:6554:1 | getDeclarationEmitOutputFilePath | 0.079893 | 0.072649 | -0.007244 | 0.9093 |
| src/compiler/program.ts:958:1 | getEmitSyntaxForUsageLocationWorker | 0.283755 | 0.276478 | -0.007277 | 0.9744 |
| src/compiler/utilitiesPublic.ts:1905:1 | isObjectBindingOrAssignmentElement | 0.046058 | 0.038721 | -0.007338 | 0.8407 |
| src/compiler/_namespaces/ts.ts:17:1 |  | 0.044035 | 0.036690 | -0.007345 | 0.8332 |
| src/compiler/commandLineParser.ts:2440:1 | AdamicJsonRecoveryObject | 0.043343 | 0.035996 | -0.007348 | 0.8305 |
| src/compiler/factory/nodeTests.ts:600:1 | isTaggedTemplateExpression | 0.048785 | 0.041436 | -0.007349 | 0.8494 |
| src/compiler/types.ts:973:1 | FlowContainer | 0.047753 | 0.040361 | -0.007392 | 0.8452 |
| src/compiler/types.ts:3468:1 | SwitchStatement | 0.042900 | 0.035493 | -0.007408 | 0.8273 |
| src/compiler/builder.ts:595:1 | convertToDiagnosticRelatedInformation | 0.115124 | 0.107651 | -0.007473 | 0.9351 |
| src/compiler/transformers/es2017.ts:106:1 | ES2017SubstitutionFlags | 0.042814 | 0.035337 | -0.007477 | 0.8254 |
| src/compiler/factory/nodeTests.ts:300:1 | isQuestionToken | 0.050065 | 0.042567 | -0.007498 | 0.8502 |
| src/compiler/transformers/namedEvaluation.ts:444:1 | transformNamedEvaluationOfExportAssignment | 0.045222 | 0.037701 | -0.007521 | 0.8337 |
| src/compiler/commandLineParser.ts:2369:1 |  | 0.044006 | 0.036430 | -0.007576 | 0.8278 |
| src/compiler/sourcemap.ts:690:1 | getSourcePositionOfMapping | 0.046319 | 0.038743 | -0.007576 | 0.8364 |
| src/compiler/moduleNameResolver.ts:202:1 | extensionsToExtensionsArray | 0.043833 | 0.036236 | -0.007597 | 0.8267 |
| src/compiler/utilitiesPublic.ts:1694:1 | isFunctionLikeDeclarationKind | 0.045706 | 0.038090 | -0.007616 | 0.8334 |
| src/compiler/utilities.ts:785:1 | usingSingleLineStringWriter | 0.187262 | 0.179605 | -0.007657 | 0.9591 |
| src/compiler/utilitiesPublic.ts:371:1 | textSpanContainsTextRange | 0.042304 | 0.034625 | -0.007679 | 0.8185 |
| src/compiler/sys.ts:929:1 | isIgnoredByWatchOptions | 0.048007 | 0.040327 | -0.007680 | 0.8400 |
| src/compiler/types.ts:3639:1 | NamespaceBody | 0.043161 | 0.035473 | -0.007687 | 0.8219 |
| src/compiler/factory/nodeFactory.ts:7235:1 |  | 0.042721 | 0.035018 | -0.007702 | 0.8197 |
| src/compiler/commandLineParser.ts:4105:1 | specToDiagnostic | 0.044413 | 0.036685 | -0.007729 | 0.8260 |
| src/compiler/transformers/generators.ts:318:1 | getInstructionName | 0.074987 | 0.067243 | -0.007744 | 0.8967 |
| src/compiler/factory/nodeTests.ts:706:1 | isTemplateSpan | 0.050468 | 0.042717 | -0.007751 | 0.8464 |
| src/compiler/corePublic.ts:37:1 | EqualityComparer | 0.043703 | 0.035941 | -0.007762 | 0.8224 |
| src/compiler/parser.ts:1381:1 | parseJsonText | 0.045713 | 0.037949 | -0.007764 | 0.8302 |
| src/compiler/core.ts:683:1 | indicesOf | 0.048285 | 0.040519 | -0.007766 | 0.8392 |
| src/compiler/program.ts:936:1 | getModeForUsageLocationWorker | 0.280022 | 0.272245 | -0.007777 | 0.9722 |
| src/compiler/utilitiesPublic.ts:2133:1 | isScopeMarker | 0.299449 | 0.291668 | -0.007781 | 0.9740 |
| src/compiler/utilities.ts:2083:1 | isModuleWithStringLiteralName | 0.241572 | 0.233789 | -0.007783 | 0.9678 |
| src/compiler/types.ts:2063:1 | FunctionLikeDeclarationBase | 0.043945 | 0.036150 | -0.007795 | 0.8226 |
| src/compiler/utilitiesPublic.ts:2154:1 | isForInOrOfStatement | 0.052879 | 0.045054 | -0.007825 | 0.8520 |
| src/compiler/_namespaces/ts.ts:34:1 |  | 0.045718 | 0.037885 | -0.007833 | 0.8287 |
| src/compiler/utilitiesPublic.ts:366:1 | textSpanContainsTextSpan | 0.044438 | 0.036566 | -0.007872 | 0.8229 |
| src/compiler/emitter.ts:516:1 | getOutputPathsFor | 0.281622 | 0.273734 | -0.007887 | 0.9720 |
| src/compiler/utilitiesPublic.ts:1076:1 | getJSDocTypeParameterTagsWorker | 0.276140 | 0.268238 | -0.007902 | 0.9714 |
| src/compiler/utilities.ts:3184:1 | getThisContainer | 0.044573 | 0.036670 | -0.007904 | 0.8227 |
| src/compiler/types.ts:6149:1 | MappedSymbolLinks | 0.043453 | 0.035547 | -0.007907 | 0.8180 |
| src/compiler/_namespaces/ts.ts:42:1 |  | 0.045254 | 0.037346 | -0.007908 | 0.8252 |
| src/compiler/types.ts:4241:1 | FlowSwitchClauseData | 0.042888 | 0.034927 | -0.007961 | 0.8144 |
| src/compiler/types.ts:10327:1 | PragmaArgTypeOptional | 0.046050 | 0.038069 | -0.007980 | 0.8267 |
| src/compiler/types.ts:2576:1 | EqualityOperatorOrHigher | 0.045943 | 0.037918 | -0.008026 | 0.8253 |
| src/compiler/utilities.ts:8459:1 | AllocatedSignature | 0.045384 | 0.037356 | -0.008028 | 0.8231 |
| src/compiler/transformers/generators.ts:222:1 | OpCode | 0.044493 | 0.036449 | -0.008044 | 0.8192 |
| src/compiler/transformers/declarations.ts:1903:1 | ProcessedDeclarationStatement | 0.046831 | 0.038767 | -0.008064 | 0.8278 |
| src/compiler/emitter.ts:695:1 | getAllProjectOutputs | 0.115576 | 0.107499 | -0.008077 | 0.9301 |
| src/compiler/types.ts:6730:1 | TupleTypeReference | 0.042324 | 0.034234 | -0.008091 | 0.8088 |
| src/compiler/utilities.ts:9687:1 | isImplicitGlob | 0.045353 | 0.037219 | -0.008135 | 0.8206 |
| src/compiler/commandLineParser.ts:132:1 |  | 0.067474 | 0.059311 | -0.008163 | 0.8790 |
| src/compiler/types.ts:4190:1 | FlowUnreachable | 0.056181 | 0.048010 | -0.008171 | 0.8546 |
| src/compiler/types.ts:2529:1 | MultiplicativeOperatorOrHigher | 0.049276 | 0.041095 | -0.008180 | 0.8340 |
| src/compiler/utilities.ts:9982:1 |  | 0.046258 | 0.038074 | -0.008184 | 0.8231 |
| src/compiler/factory/utilities.ts:843:1 | tryGetModuleNameFromFile | 0.131212 | 0.122999 | -0.008213 | 0.9374 |
| src/compiler/types.ts:9947:1 | DocumentPositionMapperHost | 0.044071 | 0.035831 | -0.008240 | 0.8130 |
| src/compiler/types.ts:2310:1 | ConditionalTypeNode | 0.043189 | 0.034948 | -0.008241 | 0.8092 |
| src/compiler/corePublic.ts:43:1 | Comparison | 0.043683 | 0.035429 | -0.008254 | 0.8110 |
| src/compiler/parser.ts:10800:1 | tagNamesAreEquivalent | 0.121091 | 0.112820 | -0.008271 | 0.9317 |
| src/compiler/program.ts:509:1 | CompilerHostLikeWithCache | 0.044962 | 0.036649 | -0.008313 | 0.8151 |
| src/compiler/types.ts:2361:1 | StringLiteral | 0.043404 | 0.035080 | -0.008324 | 0.8082 |
| src/compiler/utilities.ts:11458:1 | findConstructorDeclaration | 0.046893 | 0.038538 | -0.008356 | 0.8218 |
| src/compiler/utilities.ts:6249:1 | escapeJsxAttributeString | 0.044360 | 0.035985 | -0.008374 | 0.8112 |
| src/compiler/utilities.ts:4094:1 | isExportsIdentifier | 0.231415 | 0.223032 | -0.008383 | 0.9638 |
| src/compiler/utilities.ts:2763:1 | getJSDocCommentRanges | 0.054614 | 0.046211 | -0.008404 | 0.8461 |
| src/compiler/factory/nodeTests.ts:784:1 | isTryStatement | 0.053424 | 0.045010 | -0.008414 | 0.8425 |
| src/compiler/utilities.ts:3182:1 | getThisContainer | 0.045944 | 0.037526 | -0.008418 | 0.8168 |
| src/compiler/commandLineParser.ts:2723:1 | filterSameAsDefaultInclude | 0.048496 | 0.040074 | -0.008422 | 0.8263 |
| src/compiler/moduleNameResolver.ts:3364:1 | loadModuleFromGlobalCache | 0.076391 | 0.067943 | -0.008448 | 0.8894 |
| src/compiler/utilities.ts:2788:1 | isPartOfTypeNode | 0.138266 | 0.129792 | -0.008474 | 0.9387 |
| src/compiler/factory/nodeTests.ts:616:1 | isArrowFunction | 0.048562 | 0.040078 | -0.008485 | 0.8253 |
| src/compiler/semver.ts:409:1 | testDisjunction | 0.044038 | 0.035461 | -0.008577 | 0.8052 |
| src/compiler/utilities.ts:3704:1 | isInExpressionContext | 0.082658 | 0.074076 | -0.008582 | 0.8962 |
| src/compiler/utilities.ts:1276:1 | getNonDecoratorTokenPosOfNode | 0.235607 | 0.227022 | -0.008585 | 0.9636 |
| src/compiler/utilities.ts:6236:1 | encodeJsxCharacterEntity | 0.044301 | 0.035677 | -0.008624 | 0.8053 |
| src/compiler/utilities.ts:2782:1 |  | 0.043700 | 0.035070 | -0.008630 | 0.8025 |
| src/compiler/utilitiesPublic.ts:1432:1 | skipPartiallyEmittedExpressions | 0.046053 | 0.037413 | -0.008641 | 0.8124 |
| src/compiler/utilities.ts:2785:1 |  | 0.044102 | 0.035461 | -0.008641 | 0.8041 |
| src/compiler/commandLineParser.ts:1910:1 | OptionsBase | 0.047700 | 0.039055 | -0.008645 | 0.8188 |
| src/compiler/moduleNameResolver.ts:354:1 | PackageJson | 0.043406 | 0.034760 | -0.008646 | 0.8008 |
| src/compiler/factory/nodeTests.ts:776:1 | isLabeledStatement | 0.062854 | 0.054176 | -0.008679 | 0.8619 |
| src/compiler/utilities.ts:3257:1 | isThisContainerOrFunctionBlock | 0.050657 | 0.041966 | -0.008692 | 0.8284 |
| src/compiler/types.ts:3203:1 | JsxCallLike | 0.059642 | 0.050949 | -0.008692 | 0.8543 |
| src/compiler/semver.ts:249:1 |  | 0.045751 | 0.037039 | -0.008711 | 0.8096 |
| src/compiler/utilitiesPublic.ts:1921:1 | isArrayBindingOrAssignmentPattern | 0.043516 | 0.034796 | -0.008720 | 0.7996 |
| src/compiler/types.ts:3344:1 | SyntheticReferenceExpression | 0.044810 | 0.036018 | -0.008792 | 0.8038 |
| src/compiler/commandLineParser.ts:2706:1 | optionDependsOn | 0.049874 | 0.041056 | -0.008818 | 0.8232 |
| src/compiler/utilities.ts:2758:1 | getLeadingCommentRangesOfNode | 0.049278 | 0.040381 | -0.008897 | 0.8195 |
| src/compiler/utilities.ts:7450:1 | ClassImplementingOrExtendingExpressionWithTypeArguments | 0.045818 | 0.036905 | -0.008912 | 0.8055 |
| src/compiler/factory/utilities.ts:1203:1 | isShiftOperatorOrHigher | 0.113823 | 0.104882 | -0.008941 | 0.9214 |
| src/compiler/moduleNameResolver.ts:123:1 | isTraceEnabled | 0.078511 | 0.069562 | -0.008950 | 0.8860 |
| src/compiler/transformers/namedEvaluation.ts:274:1 | transformNamedEvaluationOfShorthandAssignmentProperty | 0.048945 | 0.039914 | -0.009031 | 0.8155 |
| src/compiler/scanner.ts:4092:1 |  | 0.046837 | 0.037791 | -0.009046 | 0.8069 |
| src/compiler/utilities.ts:10532:1 | pseudoBigIntToString | 0.115801 | 0.106717 | -0.009085 | 0.9215 |
| src/compiler/utilities.ts:8462:1 | ObjectAllocator | 0.047696 | 0.038588 | -0.009108 | 0.8090 |
| src/compiler/factory/utilities.ts:1557:1 | formatGeneratedNamePart | 0.047047 | 0.037936 | -0.009111 | 0.8063 |
| src/compiler/factory/nodeTests.ts:588:1 | isElementAccessExpression | 0.050982 | 0.041864 | -0.009117 | 0.8212 |
| src/compiler/program.ts:857:1 | getModeForFileReference | 4.310396 | 4.301272 | -0.009124 | 0.9979 |
| src/compiler/utilities.ts:3773:1 | isExternalModuleImportEqualsDeclaration | 0.167778 | 0.158634 | -0.009144 | 0.9455 |
| src/compiler/transformers/utilities.ts:471:1 | isSimpleCopiableExpression | 0.289459 | 0.280304 | -0.009155 | 0.9684 |
| src/compiler/moduleNameResolver.ts:2489:1 | loadNodeModuleFromDirectoryWorker | 0.049475 | 0.040316 | -0.009159 | 0.8149 |
| src/compiler/utilities.ts:6241:1 | getJsxAttributeStringReplacement | 0.046993 | 0.037827 | -0.009165 | 0.8050 |
| src/compiler/utilities.ts:3383:1 | getImmediatelyInvokedFunctionExpression | 0.172231 | 0.163050 | -0.009181 | 0.9467 |
| src/compiler/executeCommandLine.ts:188:1 | createColors | 0.289565 | 0.280368 | -0.009198 | 0.9682 |
| src/compiler/utilities.ts:1238:1 | getTokenPosOfNode | 0.126764 | 0.117561 | -0.009203 | 0.9274 |
| src/compiler/utilities.ts:3794:1 | isFullSourceFile | 0.078874 | 0.069656 | -0.009218 | 0.8831 |
| src/compiler/performance.ts:61:1 |  | 0.048188 | 0.038932 | -0.009256 | 0.8079 |
| src/compiler/factory/nodeTests.ts:275:1 | isDotDotDotToken | 0.052877 | 0.043616 | -0.009261 | 0.8249 |
| src/compiler/expressionToTypeNode.ts:164:1 | syntacticResult | 0.050616 | 0.041355 | -0.009262 | 0.8170 |
| src/compiler/types.ts:3314:1 | JsxChild | 0.044026 | 0.034759 | -0.009267 | 0.7895 |
| src/compiler/semver.ts:28:1 |  | 0.047156 | 0.037878 | -0.009278 | 0.8033 |
| src/compiler/types.ts:2318:1 | InferTypeNode | 0.044172 | 0.034870 | -0.009302 | 0.7894 |
| src/compiler/core.ts:805:1 | sortAndDeduplicate | 0.048299 | 0.038993 | -0.009305 | 0.8073 |
| src/compiler/commandLineParser.ts:141:1 |  | 0.046898 | 0.037574 | -0.009324 | 0.8012 |
| src/compiler/factory/nodeTests.ts:716:1 | isBlock | 0.054315 | 0.044985 | -0.009330 | 0.8282 |
| src/compiler/emitter.ts:481:1 | getTsBuildInfoEmitOutputFilePath | 0.052614 | 0.043272 | -0.009342 | 0.8224 |
| src/compiler/transformers/es2017.ts:112:1 | ContextFlags | 0.044819 | 0.035476 | -0.009344 | 0.7915 |
| src/compiler/factory/utilities.ts:656:1 | skipOuterExpressions | 0.046475 | 0.037038 | -0.009437 | 0.7969 |
| src/compiler/utilitiesPublic.ts:1433:1 | skipPartiallyEmittedExpressions | 0.045028 | 0.035534 | -0.009494 | 0.7891 |
| src/compiler/transformers/utilities.ts:593:1 | isStaticPropertyDeclarationOrClassStaticBlockDeclaration | 0.230586 | 0.221085 | -0.009501 | 0.9588 |
| src/compiler/utilities.ts:9312:1 |  | 0.081991 | 0.072479 | -0.009512 | 0.8840 |
| src/compiler/utilities.ts:9433:1 | getJSXRuntimeImport | 0.045218 | 0.035624 | -0.009594 | 0.7878 |
| src/compiler/factory/utilities.ts:1271:1 | BinaryExpressionState | 0.049372 | 0.039775 | -0.009597 | 0.8056 |
| src/compiler/factory/parenthesizerRules.ts:54:1 | createParenthesizerRules | 0.148637 | 0.139027 | -0.009610 | 0.9353 |
| src/compiler/core.ts:1769:1 | isString | 0.047699 | 0.038085 | -0.009614 | 0.7985 |
| src/compiler/transformers/utilities.ts:781:1 | getAllDecoratorsOfProperty | 0.081399 | 0.071770 | -0.009629 | 0.8817 |
| src/compiler/utilitiesPublic.ts:1036:1 | getJSDocParameterTagsWorker | 0.189999 | 0.180340 | -0.009659 | 0.9492 |
| src/compiler/moduleNameResolver.ts:1367:1 | createTypeReferenceDirectiveResolutionCache | 0.226138 | 0.216454 | -0.009685 | 0.9572 |
| src/compiler/emitter.ts:1211:1 | createPrinter | 0.087552 | 0.077860 | -0.009692 | 0.8893 |
| src/compiler/types.ts:8742:1 | NodeFactory | 0.046933 | 0.037237 | -0.009696 | 0.7934 |
| src/compiler/utilities.ts:11173:1 | getPropertyNameFromType | 0.052644 | 0.042910 | -0.009734 | 0.8151 |
| src/compiler/commandLineParser.ts:363:1 |  | 0.048118 | 0.038376 | -0.009742 | 0.7975 |
| src/compiler/utilities.ts:2781:1 |  | 0.045176 | 0.035410 | -0.009766 | 0.7838 |
| src/compiler/path.ts:421:1 | getAnyExtensionFromPath | 0.044823 | 0.035048 | -0.009775 | 0.7819 |
| src/compiler/core.ts:1750:1 | isArray | 0.048802 | 0.039026 | -0.009776 | 0.7997 |
| src/compiler/_namespaces/ts.ts:43:1 |  | 0.045492 | 0.035673 | -0.009819 | 0.7842 |
| src/compiler/types.ts:3358:1 | MissingDeclaration | 0.048043 | 0.038209 | -0.009834 | 0.7953 |
| src/compiler/utilities.ts:10433:1 | canIncludeBindAndCheckDiagnostics | 0.046419 | 0.036574 | -0.009844 | 0.7879 |
| src/compiler/utilities.ts:9324:1 |  | 0.081494 | 0.071648 | -0.009846 | 0.8792 |
| src/compiler/checker.ts:54207:1 | isNotOverload | 0.115837 | 0.105985 | -0.009852 | 0.9150 |
| src/compiler/utilitiesPublic.ts:1437:1 | isNonNullChain | 0.272622 | 0.262684 | -0.009938 | 0.9635 |
| src/compiler/types.ts:6778:1 | AnonymousType | 0.044786 | 0.034836 | -0.009950 | 0.7778 |
| src/compiler/types.ts:10637:1 | SyntacticTypeNodeBuilderResolver | 0.046517 | 0.036560 | -0.009957 | 0.7859 |
| src/compiler/commandLineParser.ts:3920:1 |  | 0.044525 | 0.034566 | -0.009959 | 0.7763 |
| src/compiler/executeCommandLine.ts:445:1 | generateSectionOptionsOutput | 0.083130 | 0.073157 | -0.009973 | 0.8800 |
| src/compiler/moduleNameResolver.ts:2555:1 | extensionIsOk | 0.048834 | 0.038854 | -0.009980 | 0.7956 |
| src/compiler/factory/nodeTests.ts:476:1 | isConstructorTypeNode | 0.052492 | 0.042493 | -0.009999 | 0.8095 |
| src/compiler/commandLineParser.ts:3522:1 | getExtendsConfigPathOrArray | 0.120506 | 0.110477 | -0.010029 | 0.9168 |
| src/compiler/commandLineParser.ts:2472:1 | convertToObject | 0.045450 | 0.035396 | -0.010054 | 0.7788 |
| src/compiler/program.ts:1027:1 | ResolutionNameAndModeGetter | 0.044828 | 0.034769 | -0.010058 | 0.7756 |
| src/compiler/utilities.ts:4104:1 | isModuleExportsAccessExpression | 0.238292 | 0.228212 | -0.010080 | 0.9577 |
| src/compiler/utilities.ts:2232:1 | isDeclarationWithTypeParameters | 0.116233 | 0.106135 | -0.010097 | 0.9131 |
| src/compiler/builderStatePublic.ts:6:1 | EmitOutput | 0.047795 | 0.037691 | -0.010104 | 0.7886 |
| src/compiler/types.ts:6493:1 | LiteralType | 0.045905 | 0.035713 | -0.010192 | 0.7780 |
| src/compiler/factory/nodeTests.ts:680:1 | isSatisfiesExpression | 0.056714 | 0.046473 | -0.010241 | 0.8194 |
| src/compiler/factory/nodeTests.ts:720:1 | isVariableStatement | 0.054160 | 0.043903 | -0.010257 | 0.8106 |
| src/compiler/executeCommandLine.ts:159:1 | defaultIsPretty | 0.185064 | 0.174788 | -0.010277 | 0.9445 |
| src/compiler/types.ts:3350:1 | EmptyStatement | 0.048393 | 0.038110 | -0.010283 | 0.7875 |
| src/compiler/parser.ts:469:1 | isFileProbablyExternalModule | 0.058680 | 0.048367 | -0.010314 | 0.8242 |
| src/compiler/utilities.ts:4029:1 | isDefaultedExpandoInitializer | 0.276558 | 0.266212 | -0.010346 | 0.9626 |
| src/compiler/commandLineParser.ts:1884:1 | parseListTypeOption | 0.055435 | 0.045062 | -0.010374 | 0.8129 |
| src/compiler/core.ts:1299:1 | getAllKeys | 0.048219 | 0.037843 | -0.010376 | 0.7848 |
| src/compiler/transformers/namedEvaluation.ts:322:1 | transformNamedEvaluationOfParameterDeclaration | 0.047913 | 0.037507 | -0.010406 | 0.7828 |
| src/compiler/utilities.ts:7535:1 | isPropertyAccessEntityNameExpression | 0.287666 | 0.277247 | -0.010419 | 0.9638 |
| src/compiler/types.ts:2131:1 | SemicolonClassElement | 0.046124 | 0.035696 | -0.010428 | 0.7739 |
| src/compiler/commandLineParser.ts:3339:1 | getErrorForNoInputFiles | 0.120799 | 0.110303 | -0.010496 | 0.9131 |
| src/compiler/utilities.ts:3968:1 | getAssignedExpandoInitializer | 0.290033 | 0.279527 | -0.010506 | 0.9638 |
| src/compiler/types.ts:3354:1 | DebuggerStatement | 0.048107 | 0.037586 | -0.010521 | 0.7813 |
| src/compiler/scanner.ts:289:1 |  | 0.047902 | 0.037374 | -0.010528 | 0.7802 |
| src/compiler/moduleSpecifiers.ts:694:1 | comparePathsByRedirectAndNumberOfDirectorySeparators | 0.047743 | 0.037174 | -0.010569 | 0.7786 |
| src/compiler/parser.ts:1372:1 | parseIsolatedEntityName | 0.049316 | 0.038746 | -0.010569 | 0.7857 |
| src/compiler/factory/nodeFactory.ts:7414:1 | createSourceMapSource | 0.047456 | 0.036862 | -0.010595 | 0.7767 |
| src/compiler/_namespaces/ts.ts:4:1 |  | 0.047116 | 0.036521 | -0.010595 | 0.7751 |
| src/compiler/builder.ts:1120:1 | IncrementalBuildInfoRootStartEnd | 0.047516 | 0.036850 | -0.010666 | 0.7755 |
| src/compiler/utilitiesPublic.ts:1416:1 | isOutermostOptionalChain | 0.269736 | 0.259057 | -0.010679 | 0.9604 |
| src/compiler/types.ts:3457:1 | ReturnStatement | 0.045970 | 0.035287 | -0.010683 | 0.7676 |
| src/compiler/_namespaces/ts.ts:5:1 |  | 0.050536 | 0.039834 | -0.010703 | 0.7882 |
| src/compiler/sourcemap.ts:645:1 | base64FormatDecode | 0.046441 | 0.035708 | -0.010734 | 0.7689 |
| src/compiler/resolutionCache.ts:505:1 | getModuleResolutionHost | 0.051759 | 0.041011 | -0.010748 | 0.7923 |
| src/compiler/_namespaces/ts.ts:25:1 |  | 0.046721 | 0.035935 | -0.010786 | 0.7691 |
| src/compiler/utilities.ts:10595:1 | isPartOfPossiblyValidTypeOrAbstractComputedPropertyName | 0.064847 | 0.054037 | -0.010810 | 0.8333 |
| src/compiler/transformers/esnext.ts:778:1 | getUsingKindOfVariableDeclarationList | 0.054210 | 0.043399 | -0.010811 | 0.8006 |
| src/compiler/utilities.ts:5444:1 | getEscapedTextOfIdentifierOrLiteral | 0.193402 | 0.182565 | -0.010837 | 0.9440 |
| src/compiler/_namespaces/ts.ts:51:1 |  | 0.047176 | 0.036218 | -0.010959 | 0.7677 |
| src/compiler/emitter.ts:426:1 |  | 0.048196 | 0.037224 | -0.010972 | 0.7724 |
| src/compiler/utilities.ts:8976:1 | isFileForcedToBeModuleByFormat | 0.046477 | 0.035498 | -0.010979 | 0.7638 |
| src/compiler/utilities.ts:2054:1 | getTextOfConstantValue | 0.072677 | 0.061676 | -0.011001 | 0.8486 |
| src/compiler/program.ts:884:1 | isExclusivelyTypeOnlyImportOrExport | 0.191539 | 0.180509 | -0.011030 | 0.9424 |
| src/compiler/core.ts:678:1 | selectIndex | 0.049047 | 0.038012 | -0.011036 | 0.7750 |
| src/compiler/utilities.ts:3551:1 | childIsDecorated | 0.163029 | 0.151975 | -0.011054 | 0.9322 |
| src/compiler/performance.ts:179:1 | enable | 0.048177 | 0.037107 | -0.011070 | 0.7702 |
| src/compiler/utilities.ts:9597:1 |  | 0.046862 | 0.035658 | -0.011205 | 0.7609 |
| src/compiler/utilities.ts:1182:1 | isRecognizedTripleSlashComment | 0.047550 | 0.036319 | -0.011231 | 0.7638 |
| src/compiler/scanner.ts:596:1 | isHexDigit | 0.049332 | 0.038066 | -0.011265 | 0.7716 |
| src/compiler/types.ts:2569:1 | EqualityOperator | 0.049049 | 0.037748 | -0.011301 | 0.7696 |
| src/compiler/utilities.ts:3553:1 | childIsDecorated | 0.160171 | 0.148864 | -0.011308 | 0.9294 |
| src/compiler/factory/emitNode.ts:1:1 |  | 0.047852 | 0.036542 | -0.011310 | 0.7636 |
| src/compiler/factory/utilities.ts:987:1 | getRestIndicatorOfBindingOrAssignmentElement | 0.054142 | 0.042827 | -0.011315 | 0.7910 |
| src/compiler/utilities.ts:6229:1 |  | 0.048302 | 0.036967 | -0.011335 | 0.7653 |
| src/compiler/utilities.ts:7444:1 | tryGetClassExtendingExpressionWithTypeArguments | 0.059120 | 0.047726 | -0.011395 | 0.8073 |
| src/compiler/types.ts:2906:1 | TemplateExpression | 0.049121 | 0.037701 | -0.011420 | 0.7675 |
| src/compiler/factory/emitHelpers.ts:716:1 | helperString | 0.047950 | 0.036520 | -0.011431 | 0.7616 |
| src/compiler/utilities.ts:8474:1 | Symbol | 0.048724 | 0.037213 | -0.011511 | 0.7637 |
| src/compiler/utilitiesPublic.ts:1884:1 | isBindingOrAssignmentPattern | 0.296056 | 0.284514 | -0.011541 | 0.9610 |
| src/compiler/commandLineParser.ts:3799:1 | createDiagnosticForNodeInSourceFileOrCompilerDiagnostic | 0.057248 | 0.045704 | -0.011544 | 0.7984 |
| src/compiler/utilities.ts:10342:1 | Mutable | 0.048401 | 0.036784 | -0.011617 | 0.7600 |
| src/compiler/commandLineParser.ts:2730:1 | matchesSpecs | 0.090394 | 0.078768 | -0.011626 | 0.8714 |
| src/compiler/transformers/namedEvaluation.ts:470:1 | transformNamedEvaluation | 0.049053 | 0.037401 | -0.011652 | 0.7625 |
| src/compiler/types.ts:1803:1 | PrivateIdentifier | 0.047709 | 0.035980 | -0.011729 | 0.7541 |
| src/compiler/factory/utilities.ts:215:1 | createJsxFactoryExpressionFromEntityName | 0.191129 | 0.179400 | -0.011730 | 0.9386 |
| src/compiler/commandLineParser.ts:1716:1 |  | 0.048169 | 0.036435 | -0.011734 | 0.7564 |
| src/compiler/core.ts:1266:1 | hasProperty | 0.057223 | 0.045473 | -0.011750 | 0.7947 |
| src/compiler/utilities.ts:1142:1 | isAnyPrologueDirective | 0.284752 | 0.272999 | -0.011753 | 0.9587 |
| src/compiler/types.ts:3462:1 | WithStatement | 0.047920 | 0.036140 | -0.011780 | 0.7542 |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts:3:1 |  | 0.498502 | 0.486720 | -0.011782 | 0.9764 |
| src/compiler/types.ts:1339:1 | HasIllegalDecorators | 0.050154 | 0.038313 | -0.011841 | 0.7639 |
| src/compiler/utilitiesPublic.ts:1651:1 | isEntityName | 0.056449 | 0.044598 | -0.011851 | 0.7901 |
| src/compiler/factory/nodeTests.ts:480:1 | isTypeQueryNode | 0.053782 | 0.041917 | -0.011864 | 0.7794 |
| src/compiler/sourcemap.ts:1:1 |  | 0.047922 | 0.036056 | -0.011865 | 0.7524 |
| src/compiler/types.ts:7398:1 | WatchDirectoryKind | 0.047288 | 0.035409 | -0.011879 | 0.7488 |
| src/compiler/types.ts:2564:1 | RelationalOperatorOrHigher | 0.049742 | 0.037833 | -0.011909 | 0.7606 |
| src/compiler/utilities.ts:3420:1 | isThisInitializedDeclaration | 0.286510 | 0.274589 | -0.011921 | 0.9584 |
| src/compiler/transformers/es2017.ts:1:1 |  | 0.046904 | 0.034983 | -0.011922 | 0.7458 |
| src/compiler/factory/nodeTests.ts:251:1 | isRegularExpressionLiteral | 0.050941 | 0.039008 | -0.011934 | 0.7657 |
| src/compiler/transformers/destructuring.ts:620:1 | makeBindingElement | 0.082882 | 0.070920 | -0.011962 | 0.8557 |
| src/compiler/program.ts:1141:1 | getInferredLibraryNameResolveFrom | 0.048982 | 0.036992 | -0.011991 | 0.7552 |
| src/compiler/types.ts:6785:1 | InstantiationExpressionType | 0.047741 | 0.035743 | -0.011998 | 0.7487 |
| src/compiler/performanceCore.ts:104:1 |  | 0.048014 | 0.036012 | -0.012002 | 0.7500 |
| src/compiler/performance.ts:170:1 | isEnabled | 0.048377 | 0.036370 | -0.012007 | 0.7518 |
| src/compiler/types.ts:3255:1 | JsxFragment | 0.081063 | 0.068962 | -0.012101 | 0.8507 |
| src/compiler/utilities.ts:11157:1 | intrinsicTagNameToString | 0.193087 | 0.180963 | -0.012123 | 0.9372 |
| src/compiler/emitter.ts:1191:1 | PipelinePhase | 0.048263 | 0.036119 | -0.012144 | 0.7484 |
| src/compiler/emitter.ts:746:1 | emitResolverSkipsTypeChecking | 0.048551 | 0.036337 | -0.012214 | 0.7484 |
| src/compiler/types.ts:3627:1 | AmbientModuleDeclaration | 0.047917 | 0.035701 | -0.012216 | 0.7451 |
| src/compiler/sourcemap.ts:686:1 | compareGeneratedPositions | 0.050752 | 0.038530 | -0.012222 | 0.7592 |
| src/compiler/types.ts:4267:1 | FlowType | 0.049342 | 0.037109 | -0.012234 | 0.7521 |
| src/compiler/factory/baseNodeFactory.ts:1:1 |  | 0.047987 | 0.035752 | -0.012236 | 0.7450 |
| src/compiler/utilities.ts:3283:1 | isInTopLevelContext | 0.247081 | 0.234843 | -0.012238 | 0.9505 |
| src/compiler/factory/nodeTests.ts:247:1 | isJsxText | 0.051548 | 0.039303 | -0.012245 | 0.7625 |
| src/compiler/core.ts:1527:1 | MultiMap | 0.048117 | 0.035854 | -0.012263 | 0.7451 |
| src/compiler/commandLineParser.ts:3258:1 | handleWatchOptionsConfigDirTemplateSubstitution | 0.050965 | 0.038689 | -0.012276 | 0.7591 |
| src/compiler/utilities.ts:6216:1 | escapeNonAsciiString | 0.049059 | 0.036768 | -0.012291 | 0.7495 |
| src/compiler/utilities.ts:11165:1 | isTypeUsableAsPropertyName | 0.052325 | 0.039986 | -0.012340 | 0.7642 |
| src/compiler/commandLineParser.ts:2442:1 | convertConfigFileToObject | 0.048380 | 0.035974 | -0.012406 | 0.7436 |
| src/compiler/types.ts:1811:1 | GeneratedPrivateIdentifier | 0.048274 | 0.035795 | -0.012479 | 0.7415 |
| src/compiler/utilities.ts:6230:1 |  | 0.048106 | 0.035525 | -0.012581 | 0.7385 |
| src/compiler/types.ts:7389:1 | WatchFileKind | 0.047208 | 0.034591 | -0.012617 | 0.7327 |
| src/compiler/commandLineParser.ts:2719:1 | optionMapToObject | 0.048903 | 0.036171 | -0.012732 | 0.7397 |
| src/compiler/factory/parenthesizerRules.ts:675:1 |  | 0.122394 | 0.109650 | -0.012744 | 0.8959 |
| src/compiler/utilitiesPublic.ts:1778:1 | isModifierLike | 0.275508 | 0.262688 | -0.012819 | 0.9535 |
| src/compiler/utilities.ts:5625:1 | nodeIsSynthesized | 0.048930 | 0.036073 | -0.012857 | 0.7372 |
| src/compiler/utilities.ts:11319:1 | createEvaluator | 0.050787 | 0.037738 | -0.013048 | 0.7431 |
| src/compiler/factory/nodeTests.ts:684:1 | isNonNullExpression | 0.053798 | 0.040711 | -0.013087 | 0.7567 |
| src/compiler/transformers/ts.ts:209:1 |  | 0.048579 | 0.035482 | -0.013097 | 0.7304 |
| src/compiler/types.ts:2382:1 | TemplateLiteralTypeSpan | 0.048800 | 0.035702 | -0.013098 | 0.7316 |
| src/compiler/types.ts:6702:1 | ElementFlags | 0.047990 | 0.034831 | -0.013158 | 0.7258 |
| src/compiler/transformers/destructuring.ts:624:1 | makeAssignmentElement | 0.048393 | 0.035233 | -0.013160 | 0.7281 |
| src/compiler/emitter.ts:431:1 | isBuildInfoFile | 0.048594 | 0.035433 | -0.013161 | 0.7292 |
| src/compiler/emitter.ts:2:1 |  | 0.049966 | 0.036774 | -0.013192 | 0.7360 |
| src/compiler/types.ts:3631:1 | ModuleDeclaration | 0.048268 | 0.035071 | -0.013197 | 0.7266 |
| src/compiler/utilities.ts:3425:1 | isThisInitializedObjectBindingExpression | 0.285600 | 0.272399 | -0.013201 | 0.9538 |
| src/compiler/parser.ts:476:1 | isAnExternalModuleIndicatorNode | 0.263129 | 0.249920 | -0.013209 | 0.9498 |
| src/compiler/types.ts:2345:1 | MappedTypeNode | 0.049751 | 0.036533 | -0.013218 | 0.7343 |
| src/compiler/program.ts:500:1 | CompilerHostLikeForCache | 0.050217 | 0.036988 | -0.013228 | 0.7366 |
| src/compiler/factory/nodeTests.ts:1066:1 | isJSDocOptionalType | 0.054742 | 0.041479 | -0.013263 | 0.7577 |
| src/compiler/factory/nodeTests.ts:269:1 | isTemplateTail | 0.054280 | 0.040974 | -0.013306 | 0.7549 |
| src/compiler/utilities.ts:9415:1 | getJSXImplicitImportBase | 0.050286 | 0.036960 | -0.013326 | 0.7350 |
| src/compiler/factory/utilities.ts:1251:1 | isLogicalOperatorOrHigher | 0.115753 | 0.102424 | -0.013329 | 0.8849 |
| src/compiler/commandLineParser.ts:3840:1 | normalizeNonListOptionValue | 0.047880 | 0.034526 | -0.013355 | 0.7211 |
| src/compiler/tsbuildPublic.ts:1:1 |  | 0.047911 | 0.034468 | -0.013443 | 0.7194 |
| src/compiler/utilities.ts:3576:1 | classElementOrClassElementParameterIsDecorated | 0.053095 | 0.039646 | -0.013449 | 0.7467 |
| src/compiler/transformers/classFields.ts:3354:1 | isStaticPropertyDeclaration | 0.270566 | 0.257105 | -0.013461 | 0.9503 |
| src/compiler/utilities.ts:5120:1 | isIdentifierName | 0.051069 | 0.037526 | -0.013543 | 0.7348 |
| src/compiler/moduleNameResolver.ts:118:1 | trace | 0.082397 | 0.068840 | -0.013557 | 0.8355 |
| src/compiler/utilitiesPublic.ts:1422:1 | isNullishCoalesce | 0.172974 | 0.159374 | -0.013600 | 0.9214 |
| src/compiler/factory/utilities.ts:668:1 | walkUpOuterExpressions | 0.322912 | 0.309297 | -0.013615 | 0.9578 |
| src/compiler/utilitiesPublic.ts:2419:1 | isStatement | 0.245998 | 0.232363 | -0.013635 | 0.9446 |
| src/compiler/moduleNameResolver.ts:1360:1 | createTypeReferenceDirectiveResolutionCache | 0.231272 | 0.217609 | -0.013663 | 0.9409 |
| src/compiler/factory/utilities.ts:1262:1 | isBinaryOperator | 0.085741 | 0.071987 | -0.013754 | 0.8396 |
| src/compiler/utilities.ts:655:1 | isTransientSymbol | 0.049320 | 0.035556 | -0.013763 | 0.7209 |
| src/compiler/program.ts:5178:1 | getModuleNames | 0.136588 | 0.122800 | -0.013788 | 0.8991 |
| src/compiler/utilities.ts:9692:1 | getPatternFromSpec | 0.060669 | 0.046862 | -0.013807 | 0.7724 |
| src/compiler/types.ts:7378:1 | ProjectReference | 0.048728 | 0.034909 | -0.013819 | 0.7164 |
| src/compiler/commandLineParser.ts:3906:1 |  | 0.049356 | 0.035480 | -0.013876 | 0.7189 |
| src/compiler/types.ts:6921:1 | TypeVariable | 0.049410 | 0.035484 | -0.013926 | 0.7182 |
| src/compiler/utilities.ts:2884:1 | isPartOfTypeExpressionWithTypeArguments | 0.269342 | 0.255296 | -0.014046 | 0.9479 |
| src/compiler/utilitiesPublic.ts:1199:1 | getJSDocSatisfiesTag | 0.066366 | 0.052275 | -0.014090 | 0.7877 |
| src/compiler/program.ts:342:1 | computeCommonSourceDirectoryOfFilenames | 0.052303 | 0.038167 | -0.014136 | 0.7297 |
| src/compiler/utilities.ts:6298:1 | createTextWriter | 0.050572 | 0.036398 | -0.014174 | 0.7197 |
| src/compiler/utilities.ts:6214:1 |  | 0.049213 | 0.035012 | -0.014201 | 0.7114 |
| src/compiler/utilities.ts:5611:1 | nodeStartsNewLexicalEnvironment | 0.054407 | 0.040186 | -0.014221 | 0.7386 |
| src/compiler/types.ts:7299:1 | DiagnosticAndArguments | 0.048905 | 0.034684 | -0.014221 | 0.7092 |
| src/compiler/program.ts:5190:1 | getModuleNameStringLiteralAt | 0.132641 | 0.118411 | -0.014230 | 0.8927 |
| src/compiler/utilities.ts:5184:1 | getExportAssignmentExpression | 0.198981 | 0.184715 | -0.014266 | 0.9283 |
| src/compiler/factory/utilities.ts:1570:1 | formatIdentifierWorker | 0.271168 | 0.256873 | -0.014295 | 0.9473 |
| src/compiler/factory/nodeTests.ts:584:1 | isPropertyAccessExpression | 0.055918 | 0.041594 | -0.014325 | 0.7438 |
| src/compiler/path.ts:388:1 | tryGetExtensionFromPath | 0.049245 | 0.034878 | -0.014367 | 0.7082 |
| src/compiler/transformers/namedEvaluation.ts:294:1 | transformNamedEvaluationOfVariableDeclaration | 0.050734 | 0.036171 | -0.014563 | 0.7130 |
| src/compiler/factory/utilities.ts:1217:1 | isRelationalOperatorOrHigher | 0.118594 | 0.104026 | -0.014568 | 0.8772 |
| src/compiler/moduleNameResolver.ts:1582:1 | tryLoadModuleUsingRootDirs | 0.055260 | 0.040671 | -0.014589 | 0.7360 |
| src/compiler/performance.ts:146:1 | clearMeasures | 0.084627 | 0.069960 | -0.014667 | 0.8267 |
| src/compiler/scanner.ts:224:1 |  | 0.080808 | 0.066102 | -0.014705 | 0.8180 |
| src/compiler/factory/nodeConverters.ts:1:1 |  | 0.050223 | 0.035465 | -0.014757 | 0.7062 |
| src/compiler/utilities.ts:3477:1 | nodeCanBeDecorated | 0.051482 | 0.036600 | -0.014881 | 0.7109 |
| src/compiler/moduleSpecifiers.ts:926:1 | tryGetModuleNameFromPaths | 0.130826 | 0.115764 | -0.015062 | 0.8849 |
| src/compiler/transformers/es2020.ts:42:1 | transformES2020 | 0.135611 | 0.120525 | -0.015086 | 0.8888 |
| src/compiler/types.ts:2356:1 | LiteralTypeNode | 0.049833 | 0.034745 | -0.015088 | 0.6972 |
| src/compiler/transformers/es2019.ts:17:1 | transformES2019 | 0.079430 | 0.064301 | -0.015128 | 0.8095 |
| src/compiler/utilities.ts:6901:1 | getTypeAnnotationNode | 0.058536 | 0.043401 | -0.015134 | 0.7414 |
| src/compiler/utilities.ts:2066:1 | isBlockOrCatchScoped | 0.056053 | 0.040807 | -0.015246 | 0.7280 |
| src/compiler/factory/utilities.ts:556:1 | isInternalName | 0.056198 | 0.040891 | -0.015307 | 0.7276 |
| src/compiler/utilities.ts:3434:1 | getEntityNameFromTypeNode | 0.167028 | 0.151718 | -0.015310 | 0.9083 |
| src/compiler/utilities.ts:10382:1 | rangeOfNode | 0.058195 | 0.042866 | -0.015329 | 0.7366 |
| src/compiler/types.ts:1200:1 | HasJSDoc | 0.063397 | 0.047955 | -0.015443 | 0.7564 |
| src/compiler/emitter.ts:724:1 | getFirstProjectOutput | 0.121836 | 0.106316 | -0.015520 | 0.8726 |
| src/compiler/program.ts:1183:1 | SyntheticReferenceFileLocation | 0.052287 | 0.036542 | -0.015745 | 0.6989 |
| src/compiler/utilities.ts:9615:1 | WildcardMatcher | 0.056519 | 0.040654 | -0.015864 | 0.7193 |
| src/compiler/utilitiesPublic.ts:352:1 | textSpanIsEmpty | 0.051622 | 0.035711 | -0.015911 | 0.6918 |
| src/compiler/factory/nodeTests.ts:744:1 | isForStatement | 0.057810 | 0.041803 | -0.016006 | 0.7231 |
| src/compiler/utilitiesPublic.ts:1798:1 | isObjectLiteralElementLike | 0.055940 | 0.039878 | -0.016063 | 0.7129 |
| src/compiler/utilitiesPublic.ts:1239:1 | getJSDocReturnType | 0.319400 | 0.303167 | -0.016233 | 0.9492 |
| src/compiler/path.ts:398:1 | getAnyExtensionFromPathWorker | 0.050970 | 0.034685 | -0.016286 | 0.6805 |
| src/compiler/resolutionCache.ts:498:1 | getRootDirectoryOfResolutionCache | 0.051540 | 0.034937 | -0.016602 | 0.6779 |
| src/compiler/utilitiesPublic.ts:1932:1 | isArrayBindingOrAssignmentElement | 0.168219 | 0.151584 | -0.016635 | 0.9011 |
| src/compiler/commandLineParser.ts:3766:1 | getDefaultTypeAcquisition | 0.051953 | 0.035210 | -0.016743 | 0.6777 |
| src/compiler/transformers/ts.ts:1:1 |  | 0.053026 | 0.036252 | -0.016774 | 0.6837 |
| src/compiler/commandLineParser.ts:2341:1 |  | 0.078894 | 0.061906 | -0.016988 | 0.7847 |
| src/compiler/types.ts:4220:1 | FlowCall | 0.052748 | 0.035734 | -0.017014 | 0.6774 |
| src/compiler/factory/utilities.ts:334:1 | createExpressionFromEntityName | 0.199729 | 0.182575 | -0.017154 | 0.9141 |
| src/compiler/factory/nodeFactory.ts:7399:1 |  | 0.307309 | 0.290128 | -0.017181 | 0.9441 |
| src/compiler/parser.ts:1344:1 | createSourceFile | 0.279175 | 0.261710 | -0.017465 | 0.9374 |
| src/compiler/types.ts:1313:1 | HasExpressionInitializer | 0.060279 | 0.042802 | -0.017477 | 0.7101 |
| src/compiler/commandLineParser.ts:3323:1 | getSubstitutedMapLikeOfStringArrayWithConfigDirTemplate | 0.195411 | 0.177843 | -0.017568 | 0.9101 |
| src/compiler/utilities.ts:4997:1 | isValueSignatureDeclaration | 0.226224 | 0.208452 | -0.017773 | 0.9214 |
| src/compiler/utilities.ts:2072:1 | isCatchClauseVariableDeclarationOrBindingElement | 0.063713 | 0.045900 | -0.017812 | 0.7204 |
| src/compiler/utilities.ts:10064:1 | getModuleSpecifierEndingPreference | 0.057189 | 0.039267 | -0.017922 | 0.6866 |
| src/compiler/types.ts:1299:1 | HasTypeArguments | 0.059813 | 0.041839 | -0.017974 | 0.6995 |
| src/compiler/utilities.ts:5173:1 | isAliasableExpression | 0.249139 | 0.231083 | -0.018056 | 0.9275 |
| src/compiler/utilities.ts:5153:1 | getAliasDeclarationFromName | 0.121690 | 0.103589 | -0.018101 | 0.8513 |
| src/compiler/utilities.ts:3569:1 | classOrConstructorParameterIsDecorated | 0.063250 | 0.045149 | -0.018102 | 0.7138 |
| src/compiler/utilities.ts:10591:1 | isShorthandPropertyNameUseSite | 0.258814 | 0.240704 | -0.018110 | 0.9300 |
| src/compiler/utilities.ts:6892:1 | getEffectiveTypeAnnotationNode | 0.243405 | 0.225274 | -0.018131 | 0.9255 |
| src/compiler/types.ts:4206:1 | FlowLabel | 0.053383 | 0.035252 | -0.018132 | 0.6604 |
| src/compiler/utilitiesPublic.ts:1732:1 | isClassElement | 0.059149 | 0.040996 | -0.018152 | 0.6931 |
| src/compiler/commandLineParser.ts:3865:1 | convertJsonOptionOfCustomType | 0.058732 | 0.040478 | -0.018254 | 0.6892 |
| src/compiler/moduleSpecifiers.ts:774:1 | getAllModulePaths | 0.145545 | 0.127285 | -0.018260 | 0.8745 |
| src/compiler/utilities.ts:6231:1 |  | 0.079296 | 0.060929 | -0.018367 | 0.7684 |
| src/compiler/transformers/ts.ts:235:1 | transformTypeScript | 0.155307 | 0.136894 | -0.018413 | 0.8814 |
| src/compiler/utilities.ts:7915:1 | rangeIsOnSingleLine | 0.055140 | 0.036669 | -0.018470 | 0.6650 |
| src/compiler/types.ts:967:1 | LocalsContainer | 0.062711 | 0.044227 | -0.018484 | 0.7053 |
| src/compiler/transformers/utilities.ts:866:1 | isSimpleParameter | 0.275092 | 0.256556 | -0.018536 | 0.9326 |
| src/compiler/transformers/es2021.ts:25:1 | transformES2021 | 0.088515 | 0.069943 | -0.018572 | 0.7902 |
| src/compiler/types.ts:979:1 | HasFlowNode | 0.059336 | 0.040723 | -0.018613 | 0.6863 |
| src/compiler/utilities.ts:2320:1 | isAnyImportOrReExport | 0.308816 | 0.290062 | -0.018754 | 0.9393 |
| src/compiler/utilitiesPublic.ts:1151:1 | getJSDocProtectedTagNoCache | 0.072220 | 0.053346 | -0.018875 | 0.7387 |
| src/compiler/types.ts:6623:1 | InterfaceType | 0.054364 | 0.035314 | -0.019050 | 0.6496 |
| src/compiler/commandLineParser.ts:2363:1 |  | 0.082893 | 0.063731 | -0.019162 | 0.7688 |
| src/compiler/utilitiesPublic.ts:629:1 | isEmptyBindingElement | 0.198644 | 0.179474 | -0.019170 | 0.9035 |
| src/compiler/utilities.ts:9322:1 |  | 0.091627 | 0.072255 | -0.019373 | 0.7886 |
| src/compiler/factory/utilities.ts:565:1 | isLocalName | 0.058038 | 0.038656 | -0.019382 | 0.6660 |
| src/compiler/builder.ts:2238:1 | toBuilderFileEmit | 0.124098 | 0.104517 | -0.019581 | 0.8422 |
| src/compiler/types.ts:2108:1 | MethodDeclaration | 0.055603 | 0.036020 | -0.019583 | 0.6478 |
| src/compiler/transformers/utilities.ts:630:1 | isInitializedProperty | 0.165408 | 0.145763 | -0.019646 | 0.8812 |
| src/compiler/builder.ts:310:1 | hasSameKeys | 0.118289 | 0.098560 | -0.019728 | 0.8332 |
| src/compiler/utilitiesPublic.ts:356:1 | textSpanContainsPosition | 0.055138 | 0.035329 | -0.019809 | 0.6407 |
| src/compiler/commandLineParser.ts:3572:1 | parseOwnConfigOfJsonSourceFile | 0.059499 | 0.039670 | -0.019829 | 0.6667 |
| src/compiler/builder.ts:321:1 | createBuilderProgramState | 0.166756 | 0.146900 | -0.019856 | 0.8809 |
| src/compiler/utilities.ts:7934:1 | rangeStartIsOnSameLineAsRangeEnd | 0.056249 | 0.036161 | -0.020088 | 0.6429 |
| src/compiler/utilitiesPublic.ts:1757:1 | isClassInstanceProperty | 0.294920 | 0.274640 | -0.020280 | 0.9312 |
| src/compiler/emitter.ts:710:1 | getOutputFileNames | 0.128887 | 0.108590 | -0.020297 | 0.8425 |
| src/compiler/utilities.ts:9600:1 | regExpEscape | 0.056994 | 0.036425 | -0.020569 | 0.6391 |
| src/compiler/core.ts:1066:1 |  | 0.059048 | 0.038310 | -0.020738 | 0.6488 |
| src/compiler/scanner.ts:4103:1 |  | 0.085333 | 0.063984 | -0.021349 | 0.7498 |
| src/compiler/commandLineParser.ts:1722:1 |  | 4.188917 | 4.167186 | -0.021731 | 0.9948 |
| src/compiler/types.ts:10626:1 | SyntacticTypeNodeBuilderContext | 0.058636 | 0.036773 | -0.021863 | 0.6271 |
| src/compiler/_namespaces/ts.ts:20:1 |  | 0.057966 | 0.036004 | -0.021961 | 0.6211 |
| src/compiler/types.ts:6511:1 | BigIntLiteralType | 0.057096 | 0.035124 | -0.021972 | 0.6152 |
| src/compiler/commandLineParser.ts:3744:1 | convertTypeAcquisitionFromJson | 0.063456 | 0.041371 | -0.022085 | 0.6520 |
| src/compiler/utilities.ts:9589:1 | tryRemoveDirectoryPrefix | 0.058967 | 0.036724 | -0.022243 | 0.6228 |
| src/compiler/utilities.ts:3789:1 | isInternalModuleImportEqualsDeclaration | 0.178201 | 0.155942 | -0.022259 | 0.8751 |
| src/compiler/factory/nodeTests.ts:688:1 | isMetaProperty | 0.063129 | 0.040576 | -0.022554 | 0.6427 |
| src/compiler/factory/nodeTests.ts:235:1 | isNumericLiteral | 0.063279 | 0.040374 | -0.022905 | 0.6380 |
| src/compiler/performance.ts:97:1 | measure | 0.096095 | 0.072948 | -0.023147 | 0.7591 |
| src/compiler/transformers/utilities.ts:536:1 | getSuperCallFromStatement | 0.265469 | 0.242163 | -0.023305 | 0.9122 |
| src/compiler/utilities.ts:6149:1 | hasInvalidEscape | 0.211508 | 0.187875 | -0.023633 | 0.8883 |
| src/compiler/types.ts:6640:1 | InterfaceTypeWithDeclaredMembers | 0.058208 | 0.034503 | -0.023705 | 0.5928 |
| src/compiler/factory/nodeTests.ts:752:1 | isForOfStatement | 0.065023 | 0.041301 | -0.023722 | 0.6352 |
| src/compiler/types.ts:2417:1 | UpdateExpression | 0.059266 | 0.035515 | -0.023751 | 0.5992 |
| src/compiler/types.ts:4262:1 | FlowReduceLabelData | 0.059659 | 0.035763 | -0.023896 | 0.5995 |
| src/compiler/factory/utilities.ts:856:1 | tryGetModuleNameFromDeclaration | 0.109357 | 0.085453 | -0.023904 | 0.7814 |
| src/compiler/transformers/es2017.ts:104:1 | SuperContainer | 0.059478 | 0.035551 | -0.023927 | 0.5977 |
| src/compiler/factory/utilities.ts:658:1 | skipOuterExpressions | 0.060308 | 0.036299 | -0.024009 | 0.6019 |
| src/compiler/transformers/classFields.ts:3349:1 | isPrivateIdentifierInExpression | 0.288958 | 0.264424 | -0.024534 | 0.9151 |
| src/compiler/types.ts:4250:1 | FlowArrayMutation | 0.059361 | 0.034666 | -0.024695 | 0.5840 |
| src/compiler/types.ts:1017:1 | ForEachChildNodes | 0.063239 | 0.038378 | -0.024861 | 0.6069 |
| src/compiler/types.ts:6507:1 | NumberLiteralType | 0.060382 | 0.035271 | -0.025111 | 0.5841 |
| src/compiler/emitter.ts:1147:1 |  | 0.085929 | 0.060806 | -0.025122 | 0.7076 |
| src/compiler/factory/nodeTests.ts:768:1 | isWithStatement | 0.068675 | 0.043497 | -0.025178 | 0.6334 |
| src/compiler/emitter.ts:502:1 | canEmitTsBuildInfo | 0.059831 | 0.034510 | -0.025321 | 0.5768 |
| src/compiler/types.ts:6670:1 | DeferredTypeReference | 0.059366 | 0.034044 | -0.025322 | 0.5735 |
| src/compiler/transformers/utilities.ts:642:1 | isNonStaticMethodOrAccessorWithPrivateName | 0.251912 | 0.226478 | -0.025434 | 0.8990 |
| src/compiler/program.ts:1033:1 | ResolutionLoader | 0.060557 | 0.034988 | -0.025569 | 0.5778 |
| src/compiler/utilities.ts:3137:1 | getContainingClass | 0.065267 | 0.039671 | -0.025596 | 0.6078 |
| src/compiler/utilities.ts:9660:1 | getRegularExpressionForWildcard | 0.068685 | 0.042972 | -0.025712 | 0.6256 |
| src/compiler/utilities.ts:4307:1 | setValueDeclaration | 0.115709 | 0.089498 | -0.026210 | 0.7735 |
| src/compiler/transformers/esnext.ts:774:1 | isUsingVariableDeclarationList | 0.363891 | 0.337576 | -0.026316 | 0.9277 |
| src/compiler/commandLineParser.ts:3651:1 | getExtendsConfigPath | 0.150384 | 0.123834 | -0.026551 | 0.8234 |
| src/compiler/utilities.ts:2751:1 | isHoistedVariableStatement | 0.242274 | 0.215680 | -0.026594 | 0.8902 |
| src/compiler/factory/emitHelpers.ts:704:1 | compareEmitHelpers | 4.271586 | 4.244889 | -0.026697 | 0.9938 |
| src/compiler/factory/nodeTests.ts:696:1 | isPartiallyEmittedExpression | 0.073907 | 0.046995 | -0.026912 | 0.6359 |
| src/compiler/transformers/classFields.ts:3335:1 | createPrivateInstanceMethodInitializer | 0.157420 | 0.130506 | -0.026914 | 0.8290 |
| src/compiler/types.ts:6713:1 | TupleType | 0.061316 | 0.034391 | -0.026925 | 0.5609 |
| src/compiler/factory/emitHelpers.ts:148:1 | createEmitHelperFactory | 0.684500 | 0.657466 | -0.027034 | 0.9605 |
| src/compiler/types.ts:1287:1 | HasIllegalType | 0.067713 | 0.040646 | -0.027067 | 0.6003 |
| src/compiler/utilities.ts:4132:1 | isLiteralLikeAccess | 0.287564 | 0.260314 | -0.027250 | 0.9052 |
| src/compiler/types.ts:6694:1 | GenericType | 0.061715 | 0.034088 | -0.027627 | 0.5523 |
| src/compiler/utilitiesPublic.ts:2059:1 | isUnaryExpressionWithWrite | 0.176450 | 0.148820 | -0.027630 | 0.8434 |
| src/compiler/emitter.ts:1142:1 | getBuildInfo | 0.062713 | 0.035046 | -0.027666 | 0.5588 |
| src/compiler/types.ts:10439:1 | PragmaMap | 0.075022 | 0.047098 | -0.027924 | 0.6278 |
| src/compiler/factory/emitHelpers.ts:915:1 |  | 0.088691 | 0.060748 | -0.027943 | 0.6849 |
| src/compiler/utilities.ts:10582:1 | isValidTypeOnlyAliasUseSite | 0.086676 | 0.058680 | -0.027996 | 0.6770 |
| src/compiler/types.ts:6607:1 | ObjectFlagsType | 0.063117 | 0.035077 | -0.028040 | 0.5557 |
| src/compiler/performanceCore.ts:61:1 | tryGetPerformanceHooks | 0.331346 | 0.303095 | -0.028251 | 0.9147 |
| src/compiler/utilities.ts:9673:1 | getRegularExpressionsForWildcards | 0.104198 | 0.075880 | -0.028318 | 0.7282 |
| src/compiler/types.ts:6637:1 | BaseType | 0.063283 | 0.034904 | -0.028379 | 0.5516 |
| src/compiler/utilities.ts:9298:1 |  | 0.102879 | 0.074323 | -0.028556 | 0.7224 |
| src/compiler/types.ts:4235:1 | FlowSwitchClause | 0.064104 | 0.035414 | -0.028689 | 0.5525 |
| src/compiler/utilities.ts:7581:1 | isRightSideOfQualifiedNameOrPropertyAccessOrJSDocMemberName | 0.261549 | 0.232445 | -0.029104 | 0.8887 |
| src/compiler/utilitiesPublic.ts:1815:1 | isTypeNode | 0.102270 | 0.073087 | -0.029183 | 0.7146 |
| src/compiler/program.ts:1190:1 | isReferenceFileLocation | 0.066111 | 0.036881 | -0.029230 | 0.5579 |
| src/compiler/transformers/utilities.ts:153:1 | getImportNeedsImportStarHelper | 0.224115 | 0.194745 | -0.029370 | 0.8690 |
| src/compiler/utilities.ts:12221:1 | isNewScopeNode | 0.260093 | 0.230678 | -0.029415 | 0.8869 |
| src/compiler/utilities.ts:5684:1 | getExpressionPrecedence | 0.185043 | 0.155362 | -0.029681 | 0.8396 |
| src/compiler/factory/nodeTests.ts:700:1 | isCommaListExpression | 0.077373 | 0.047643 | -0.029730 | 0.6158 |
| src/compiler/utilities.ts:2078:1 | isAmbientModule | 0.256945 | 0.227200 | -0.029745 | 0.8842 |
| src/compiler/utilities.ts:7920:1 | rangeStartPositionsAreOnSameLine | 0.067978 | 0.037494 | -0.030485 | 0.5516 |
| src/compiler/moduleSpecifiers.ts:698:1 | getNearestAncestorDirectoryWithPackageJson | 0.147510 | 0.116453 | -0.031057 | 0.7895 |
| src/compiler/commandLineParser.ts:3933:1 | getFileNamesFromConfigSpecs | 0.168706 | 0.137449 | -0.031257 | 0.8147 |
| src/compiler/binder.ts:339:1 | getModuleInstanceStateCached | 0.078267 | 0.046825 | -0.031441 | 0.5983 |
| src/compiler/utilities.ts:3885:1 | getModuleSpecifierOfBareOrAccessedRequire | 0.316591 | 0.284823 | -0.031768 | 0.8997 |
| src/compiler/scanner.ts:4065:1 |  | 0.070131 | 0.038216 | -0.031915 | 0.5449 |
| src/compiler/emitter.ts:1200:1 |  | 0.067776 | 0.035714 | -0.032063 | 0.5269 |
| src/compiler/transformers/classFields.ts:3347:1 | PrivateIdentifierInExpression | 0.068539 | 0.036063 | -0.032476 | 0.5262 |
| src/compiler/types.ts:1322:1 | HasIllegalExpressionInitializer | 0.071428 | 0.038727 | -0.032701 | 0.5422 |
| src/compiler/utilities.ts:5112:1 | isLiteralComputedPropertyDeclarationName | 0.309951 | 0.277063 | -0.032888 | 0.8939 |
| src/compiler/utilities.ts:9779:1 | replaceWildcardCharacter | 0.080137 | 0.047017 | -0.033120 | 0.5867 |
| src/compiler/utilities.ts:9653:1 |  | 0.097804 | 0.064633 | -0.033172 | 0.6608 |
| src/compiler/executeCommandLine.ts:1007:1 | createWatchStatusReporter | 0.069348 | 0.036111 | -0.033237 | 0.5207 |
| src/compiler/utilities.ts:2134:1 | isModuleAugmentationExternal | 0.317139 | 0.283452 | -0.033688 | 0.8938 |
| src/compiler/utilities.ts:4166:1 | isBindableStaticNameExpression | 0.270859 | 0.237082 | -0.033777 | 0.8753 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts:18:1 | transformImpliedNodeFormatDependentModule | 0.268648 | 0.234828 | -0.033819 | 0.8741 |
| src/compiler/factory/nodeTests.ts:692:1 | isSyntheticExpression | 0.075111 | 0.040851 | -0.034261 | 0.5439 |
| src/compiler/utilities.ts:3880:1 | isBindingElementOfBareOrAccessedRequire | 0.255043 | 0.220498 | -0.034545 | 0.8646 |
| src/compiler/path.ts:31:1 |  | 0.071692 | 0.036605 | -0.035086 | 0.5106 |
| src/compiler/utilities.ts:4019:1 | getDefaultedExpandoInitializer | 0.306782 | 0.271695 | -0.035087 | 0.8856 |
| src/compiler/scanner.ts:4068:1 | utf16EncodeAsString | 0.075184 | 0.039763 | -0.035420 | 0.5289 |
| src/compiler/scanner.ts:4041:1 | charSize | 0.074061 | 0.038588 | -0.035473 | 0.5210 |
| src/compiler/builderPublic.ts:245:1 | createAbstractBuilder | 0.183019 | 0.146995 | -0.036024 | 0.8032 |
| src/compiler/utilities.ts:6293:1 | getIndentSize | 0.071933 | 0.035827 | -0.036106 | 0.4981 |
| src/compiler/transformers/utilities.ts:618:1 | isStaticPropertyDeclaration | 0.291780 | 0.255313 | -0.036468 | 0.8750 |
| src/compiler/factory/emitNode.ts:233:1 | getConstantValue | 4.403744 | 4.367228 | -0.036516 | 0.9917 |
| src/compiler/utilities.ts:4148:1 | isBindableStaticAccessExpression | 0.277839 | 0.241197 | -0.036642 | 0.8681 |
| src/compiler/utilities.ts:9647:1 |  | 0.104122 | 0.067242 | -0.036880 | 0.6458 |
| src/compiler/factory/nodeFactory.ts:7236:1 |  | 0.096788 | 0.059791 | -0.036997 | 0.6178 |
| src/compiler/emitter.ts:752:1 | emitFiles | 0.624750 | 0.587165 | -0.037585 | 0.9398 |
| src/compiler/types.ts:6516:1 | EnumType | 0.073367 | 0.034996 | -0.038371 | 0.4770 |
| src/compiler/factory/emitHelpers.ts:879:1 |  | 0.110761 | 0.072321 | -0.038441 | 0.6529 |
| src/compiler/utilitiesPublic.ts:1964:1 | isCallLikeOrFunctionLikeExpression | 0.302332 | 0.263680 | -0.038652 | 0.8722 |
| src/compiler/utilities.ts:7939:1 | rangeEndIsOnSameLineAsRangeStart | 0.077207 | 0.038420 | -0.038787 | 0.4976 |
| src/compiler/utilities.ts:9604:1 | escapeRegExpCharacter | 0.074448 | 0.035582 | -0.038866 | 0.4779 |
| src/compiler/utilities.ts:3784:1 | getExternalModuleRequireArgument | 0.287336 | 0.248153 | -0.039183 | 0.8636 |
| src/compiler/utilities.ts:7929:1 | rangeEndPositionsAreOnSameLine | 0.076667 | 0.036969 | -0.039699 | 0.4822 |
| src/compiler/utilities.ts:11494:1 | createNameResolver | 0.340892 | 0.300604 | -0.040288 | 0.8818 |
| src/compiler/types.ts:2423:1 | PrefixUnaryOperator | 0.078271 | 0.037874 | -0.040397 | 0.4839 |
| src/compiler/scanner.ts:4052:1 | utf16EncodeAsStringFallback | 0.077548 | 0.036696 | -0.040852 | 0.4732 |
| src/compiler/utilities.ts:3922:1 | isAssignmentDeclaration | 0.314112 | 0.272708 | -0.041404 | 0.8682 |
| src/compiler/utilities.ts:12439:1 | isPotentiallyExecutableNode | 0.262756 | 0.221287 | -0.041469 | 0.8422 |
| src/compiler/scanner.ts:4075:1 |  | 0.107393 | 0.065891 | -0.041502 | 0.6136 |
| src/compiler/utilities.ts:7587:1 | isInstanceOfExpression | 0.326773 | 0.285185 | -0.041588 | 0.8727 |
| src/compiler/factory/emitHelpers.ts:1151:1 |  | 0.106848 | 0.064092 | -0.042755 | 0.5998 |
| src/compiler/path.ts:60:1 | isRootedDiskPath | 0.091031 | 0.047913 | -0.043118 | 0.5263 |
| src/compiler/utilities.ts:9612:1 |  | 0.079764 | 0.036509 | -0.043254 | 0.4577 |
| src/compiler/builderPublic.ts:253:1 | createAbstractBuilder | 0.191668 | 0.148308 | -0.043360 | 0.7738 |
| src/compiler/types.ts:6503:1 | StringLiteralType | 0.079182 | 0.035524 | -0.043658 | 0.4486 |
| src/compiler/utilities.ts:3994:1 | getExpandoInitializer | 0.342428 | 0.298387 | -0.044040 | 0.8714 |
| src/compiler/utilities.ts:7910:1 | createTokenRange | 0.090852 | 0.046798 | -0.044055 | 0.5151 |
| src/compiler/transformers/es2016.ts:21:1 | transformES2016 | 0.137612 | 0.092990 | -0.044622 | 0.6757 |
| src/compiler/utilities.ts:7870:1 | moveRangePos | 0.081978 | 0.037282 | -0.044697 | 0.4548 |
| src/compiler/factory/utilities.ts:814:1 | getExternalModuleNameLiteral | 0.323310 | 0.278556 | -0.044754 | 0.8616 |
| src/compiler/utilities.ts:3905:1 | isRequireVariableStatement | 0.265239 | 0.219791 | -0.045448 | 0.8287 |
| src/compiler/utilities.ts:7420:1 | isLogicalOrCoalescingAssignmentExpression | 0.323876 | 0.278410 | -0.045466 | 0.8596 |
| src/compiler/types.ts:10446:1 | CommentDirectivesMap | 0.092739 | 0.046999 | -0.045740 | 0.5068 |
| src/compiler/emitter.ts:1137:1 | getBuildInfoText | 0.081716 | 0.035860 | -0.045857 | 0.4388 |
| src/compiler/utilities.ts:8290:1 | showModuleSpecifier | 0.351538 | 0.305595 | -0.045944 | 0.8693 |
| src/compiler/executeCommandLine.ts:1034:1 | createWatchOfFilesAndCompilerOptions | 0.391431 | 0.345386 | -0.046045 | 0.8824 |
| src/compiler/utilities.ts:4171:1 | getNameOrArgument | 0.087916 | 0.041727 | -0.046189 | 0.4746 |
| src/compiler/checker.ts:1460:1 | getNodeId | 4.392065 | 4.345791 | -0.046274 | 0.9895 |
| src/compiler/utilities.ts:3142:1 | getContainingClassStaticBlock | 0.336865 | 0.289232 | -0.047633 | 0.8586 |
| src/compiler/utilities.ts:5073:1 | isDeclarationName | 0.278626 | 0.230956 | -0.047671 | 0.8289 |
| src/compiler/commandLineParser.ts:1713:1 |  | 4.194805 | 4.147025 | -0.047780 | 0.9886 |
| src/compiler/path.ts:50:1 | isUrl | 0.096318 | 0.048422 | -0.047895 | 0.5027 |
| src/compiler/utilities.ts:9608:1 |  | 0.095372 | 0.047124 | -0.048248 | 0.4941 |
| src/compiler/path.ts:32:1 |  | 0.093437 | 0.045085 | -0.048353 | 0.4825 |
| src/compiler/factory/utilities.ts:1599:1 | formatGeneratedName | 0.089967 | 0.040842 | -0.049126 | 0.4540 |
| src/compiler/utilities.ts:10560:1 | isValidBigIntString | 0.332562 | 0.283145 | -0.049417 | 0.8514 |
| src/compiler/utilities.ts:9409:1 | getJSXTransformEnabled | 0.093246 | 0.042471 | -0.050775 | 0.4555 |
| src/compiler/utilities.ts:810:1 | moduleResolutionIsEqualTo | 4.366145 | 4.314459 | -0.051686 | 0.9882 |
| src/compiler/path.ts:41:1 | isAnyDirectorySeparator | 0.101128 | 0.049364 | -0.051764 | 0.4881 |
| src/compiler/utilities.ts:6412:1 | getTrailingSemicolonDeferringWriter | 0.608094 | 0.556109 | -0.051985 | 0.9145 |
| src/compiler/transformers/utilities.ts:762:1 | getAllDecoratorsOfMethod | 0.202771 | 0.150610 | -0.052162 | 0.7428 |
| src/compiler/types.ts:2431:1 | PrefixUnaryExpression | 0.089403 | 0.037114 | -0.052289 | 0.4151 |
| src/compiler/utilities.ts:8961:1 | walkTreeForJSXTags | 0.341800 | 0.289228 | -0.052572 | 0.8462 |
| src/compiler/types.ts:10599:1 | PseudoBigInt | 0.089446 | 0.036533 | -0.052913 | 0.4084 |
| src/compiler/sys.ts:1972:1 |  | 0.164317 | 0.110843 | -0.053474 | 0.6746 |
| src/compiler/types.ts:2438:1 | PostfixUnaryOperator | 0.089804 | 0.036162 | -0.053642 | 0.4027 |
| src/compiler/utilitiesPublic.ts:1426:1 | isConstTypeReference | 0.319235 | 0.264848 | -0.054387 | 0.8296 |
| src/compiler/utilities.ts:3768:1 | isNamespaceReexportDeclaration | 0.330683 | 0.275966 | -0.054718 | 0.8345 |
| src/compiler/utilities.ts:3631:1 | isExpressionNode | 0.316979 | 0.262034 | -0.054945 | 0.8267 |
| src/compiler/types.ts:10618:1 | HasInferredType | 0.091955 | 0.036276 | -0.055678 | 0.3945 |
| src/compiler/core.ts:2447:1 | isPatternMatch | 0.165146 | 0.109453 | -0.055693 | 0.6628 |
| src/compiler/types.ts:10612:1 | EvaluationResolver | 0.093026 | 0.037074 | -0.055952 | 0.3985 |
| src/compiler/builder.ts:1196:1 | ensureHasErrorsForState | 4.297485 | 4.241101 | -0.056383 | 0.9869 |
| src/compiler/utilitiesPublic.ts:603:1 | getTypeParameterOwner | 0.273848 | 0.217210 | -0.056638 | 0.7932 |
| src/compiler/builder.ts:1:1 |  | 0.098566 | 0.041705 | -0.056861 | 0.4231 |
| src/compiler/types.ts:3225:1 | JsxAttributes | 0.102250 | 0.045389 | -0.056862 | 0.4439 |
| src/compiler/transformers/es2015.ts:1:1 |  | 0.093622 | 0.036119 | -0.057503 | 0.3858 |
| src/compiler/program.ts:876:1 | getModeForResolutionAtIndex | 0.106018 | 0.048466 | -0.057552 | 0.4571 |
| src/compiler/factory/emitHelpers.ts:895:1 |  | 0.119553 | 0.061033 | -0.058520 | 0.5105 |
| src/compiler/moduleNameResolver.ts:1567:1 | tryLoadModuleUsingPathsIfEligible | 0.096247 | 0.037280 | -0.058967 | 0.3873 |
| src/compiler/moduleNameResolver.ts:1493:1 | ResolutionKindSpecificLoader | 0.100645 | 0.041076 | -0.059568 | 0.4081 |
| src/compiler/types.ts:10596:1 | OrganizeImportsTypeOrder | 0.096988 | 0.036229 | -0.060759 | 0.3735 |
| src/compiler/types.ts:2456:1 | PrimaryExpression | 0.099370 | 0.038597 | -0.060772 | 0.3884 |
| src/compiler/types.ts:2442:1 | PostfixUnaryExpression | 0.098068 | 0.037089 | -0.060978 | 0.3782 |
| src/compiler/scanner.ts:4087:1 |  | 0.101084 | 0.038826 | -0.062258 | 0.3841 |
| src/compiler/transformers/es2015.ts:220:1 | ES2015SubstitutionFlags | 0.098783 | 0.036051 | -0.062732 | 0.3649 |
| src/compiler/factory/utilities.ts:575:1 | isExportName | 0.102475 | 0.039646 | -0.062829 | 0.3869 |
| src/compiler/transformers/es2015.ts:266:1 | LoopOutParameter | 0.098499 | 0.035040 | -0.063458 | 0.3557 |
| src/compiler/moduleNameResolver.ts:1555:1 | tryLoadModuleUsingOptionalResolutionSettings | 0.107414 | 0.043003 | -0.064411 | 0.4003 |
| src/compiler/types.ts:2452:1 | MemberExpression | 0.102015 | 0.036821 | -0.065194 | 0.3609 |
| src/compiler/utilitiesPublic.ts:1562:1 | isTypeOnlyImportOrExportDeclaration | 0.107854 | 0.042496 | -0.065358 | 0.3940 |
| src/compiler/utilitiesPublic.ts:1566:1 | isPartOfTypeOnlyImportOrExportDeclaration | 0.106681 | 0.041145 | -0.065537 | 0.3857 |
| src/compiler/corePublic.ts:40:1 | Comparer | 0.100925 | 0.034929 | -0.065997 | 0.3461 |
| src/compiler/factory/nodeFactory.ts:7238:1 | getCookedText | 0.528952 | 0.462413 | -0.066539 | 0.8742 |
| src/compiler/types.ts:10451:1 | UserPreferences | 0.103005 | 0.035848 | -0.067157 | 0.3480 |
| src/compiler/types.ts:10605:1 | Queue | 0.103693 | 0.035716 | -0.067977 | 0.3444 |
| src/compiler/utilitiesPublic.ts:1586:1 | isGeneratedPrivateIdentifier | 0.337289 | 0.266843 | -0.070446 | 0.7911 |
| src/compiler/utilities.ts:9610:1 |  | 0.113488 | 0.042821 | -0.070668 | 0.3773 |
| src/compiler/types.ts:2448:1 | LeftHandSideExpression | 0.107943 | 0.036886 | -0.071058 | 0.3417 |
| src/compiler/transformers/utilities.ts:878:1 | rewriteModuleSpecifier | 0.371733 | 0.300036 | -0.071697 | 0.8071 |
| src/compiler/utilities.ts:4139:1 | isLiteralLikeElementAccess | 0.301434 | 0.228220 | -0.073214 | 0.7571 |
| src/compiler/program.ts:873:1 | getModeForResolutionAtIndex | 0.121627 | 0.047896 | -0.073731 | 0.3938 |
| src/compiler/moduleSpecifiers.ts:288:1 | getNodeModulesPackageName | 4.611427 | 4.537411 | -0.074016 | 0.9839 |
| src/compiler/factory/nodeConverters.ts:41:1 | createNodeConverters | 0.289572 | 0.213518 | -0.076054 | 0.7374 |
| src/compiler/utilities.ts:4427:1 | getNamespaceDeclarationNode | 0.371392 | 0.295129 | -0.076263 | 0.7947 |
| src/compiler/builder.ts:1227:1 | getBuildInfoEmitPending | 4.308884 | 4.231803 | -0.077081 | 0.9821 |
| src/compiler/utilitiesPublic.ts:619:1 | isEmptyBindingPattern | 0.259754 | 0.180141 | -0.079614 | 0.6935 |
| src/compiler/utilities.ts:3106:1 | getTsConfigObjectLiteralExpression | 0.122943 | 0.043182 | -0.079761 | 0.3512 |
| src/compiler/utilities.ts:11452:1 | isConstAssertion | 0.382545 | 0.302635 | -0.079909 | 0.7911 |
| src/compiler/utilitiesPublic.ts:1581:1 | isGeneratedIdentifier | 0.327171 | 0.244299 | -0.082872 | 0.7467 |
| src/compiler/types.ts:1327:1 | HasDecorators | 0.127749 | 0.041409 | -0.086340 | 0.3241 |
| src/compiler/factory/emitNode.ts:166:1 | getStartsOnNewLine | 4.413071 | 4.326342 | -0.086728 | 0.9803 |
| src/compiler/utilities.ts:3127:1 | getContainingFunction | 0.133739 | 0.040759 | -0.092980 | 0.3048 |
| src/compiler/commandLineParser.ts:1701:1 |  | 4.298161 | 4.202170 | -0.095991 | 0.9777 |
| src/compiler/moduleNameResolver.ts:2228:1 | getEntrypointsFromPackageJsonInfo | 4.390782 | 4.289299 | -0.101483 | 0.9769 |
| src/compiler/utilitiesPublic.ts:1536:1 | isTypeOnlyImportDeclaration | 0.322292 | 0.220259 | -0.102033 | 0.6834 |
| src/compiler/executeCommandLine.ts:478:1 | printEasyHelp | 4.367393 | 4.265160 | -0.102233 | 0.9766 |
| src/compiler/checker.ts:54213:1 | isDeclarationNameOrImportPropertyName | 0.342305 | 0.239862 | -0.102443 | 0.7007 |
| src/compiler/factory/utilities.ts:579:1 | isUseStrictPrologue | 0.375296 | 0.268371 | -0.106925 | 0.7151 |
| src/compiler/sourcemap.ts:668:1 | isSourceMappedPosition | 4.379635 | 4.270860 | -0.108774 | 0.9752 |
| src/compiler/utilities.ts:4294:1 | isPrototypePropertyAssignment | 0.413472 | 0.301865 | -0.111608 | 0.7301 |
| src/compiler/utilities.ts:4299:1 | isSpecialPropertyDeclaration | 0.329617 | 0.216289 | -0.113328 | 0.6562 |
| src/compiler/utilities.ts:8924:1 | diagnosticsEqualityComparer | 4.428441 | 4.312677 | -0.115764 | 0.9739 |
| src/compiler/transformers/utilities.ts:389:1 | IdentifierNameMap | 0.406270 | 0.287752 | -0.118518 | 0.7083 |
| src/compiler/utilitiesPublic.ts:1570:1 | isStringTextContainingNode | 0.193741 | 0.073473 | -0.120268 | 0.3792 |
| src/compiler/moduleNameResolver.ts:224:1 | createResolvedModuleWithFailedLookupLocationsHandlingSymlink | 4.650576 | 4.528286 | -0.122290 | 0.9737 |
| src/compiler/builder.ts:625:1 | assertSourceFileOkWithoutNextAffectedCall | 4.610147 | 4.487050 | -0.123097 | 0.9733 |
| src/compiler/commandLineParser.ts:1710:1 |  | 4.375476 | 4.250113 | -0.125362 | 0.9713 |
| src/compiler/factory/utilities.ts:786:1 | getLocalNameForExternalImport | 0.448757 | 0.316693 | -0.132063 | 0.7057 |
| src/compiler/transformers/utilities.ts:876:1 | rewriteModuleSpecifier | 0.408477 | 0.274147 | -0.134330 | 0.6711 |
| src/compiler/utilities.ts:8940:1 | getDiagnosticMessage | 4.434886 | 4.296862 | -0.138024 | 0.9689 |
| src/compiler/utilities.ts:7879:1 | moveRangePastDecorators | 0.368911 | 0.229997 | -0.138915 | 0.6234 |
| src/compiler/utilities.ts:2481:1 | createDiagnosticMessageChainFromDiagnostic | 4.394078 | 4.245244 | -0.148834 | 0.9661 |
| src/compiler/commandLineParser.ts:2603:1 | getCompilerOptionValueTypeString | 4.656206 | 4.505174 | -0.151031 | 0.9676 |
| src/compiler/builder.ts:904:1 | handleDtsMayChangeOfGlobalScope | 4.388036 | 4.230709 | -0.157327 | 0.9641 |
| src/compiler/utilities.ts:4321:1 | isFunctionSymbol | 0.442606 | 0.283268 | -0.159339 | 0.6400 |
| src/compiler/executeCommandLine.ts:1011:1 | createWatchOfConfigFile | 0.518822 | 0.348371 | -0.170451 | 0.6715 |
| src/compiler/utilities.ts:3114:1 | getTsConfigPropArrayElementValue | 0.242838 | 0.061495 | -0.181343 | 0.2532 |
| src/compiler/builder.ts:790:1 | removeDiagnosticsOfLibraryFiles | 4.563512 | 4.361204 | -0.202308 | 0.9557 |
| src/compiler/transformers/taggedTemplate.ts:101:1 | createTemplateCooked | 4.429455 | 4.226662 | -0.202793 | 0.9542 |
| src/compiler/utilities.ts:8936:1 | getDiagnosticCode | 4.523339 | 4.315038 | -0.208300 | 0.9539 |
| src/compiler/utilities.ts:4178:1 | getAssignmentDeclarationKindWorker | 0.429975 | 0.215482 | -0.214492 | 0.5012 |
| src/compiler/utilities.ts:7891:1 | moveRangePastModifiers | 0.492423 | 0.271086 | -0.221337 | 0.5505 |
| src/compiler/program.ts:520:1 | changeCompilerHostLikeToUseCache | 0.376962 | 0.151095 | -0.225867 | 0.4008 |
| src/compiler/commandLineParser.ts:1718:1 |  | 4.453184 | 4.223589 | -0.229594 | 0.9484 |
| src/compiler/sourcemap.ts:673:1 | sameMappedPosition | 4.554868 | 4.311494 | -0.243374 | 0.9466 |
| src/compiler/utilities.ts:8719:1 | createCompilerDiagnostic | 4.358912 | 4.111601 | -0.247311 | 0.9433 |
| src/compiler/utilities.ts:4202:1 | isVoidZero | 0.547573 | 0.283041 | -0.264532 | 0.5169 |
| src/compiler/utilities.ts:8818:1 | compareMessageText | 4.407813 | 4.117589 | -0.290224 | 0.9342 |
| src/compiler/factory/utilities.ts:1695:1 | isSyntheticParenthesizedExpression | 0.615214 | 0.291390 | -0.323824 | 0.4736 |
| src/compiler/program.ts:779:1 | formatDiagnosticsWithColorAndContext | 4.601269 | 4.269994 | -0.331274 | 0.9280 |
| src/compiler/program.ts:1162:1 | isReferencedFile | 4.640873 | 4.288386 | -0.352487 | 0.9240 |
| src/compiler/transformers/utilities.ts:441:1 | IdentifierNameMultiMap | 0.653896 | 0.271643 | -0.382253 | 0.4154 |
| src/compiler/moduleNameResolver.ts:1406:1 | resolveModuleName | 4.822883 | 4.414527 | -0.408356 | 0.9153 |
| src/compiler/utilitiesPublic.ts:1550:1 | isTypeOnlyExportDeclaration | 0.590824 | 0.179953 | -0.410872 | 0.3046 |
| src/compiler/moduleNameResolver.ts:2414:1 | getPeerDependenciesOfPackageJsonInfo | 4.748097 | 4.308220 | -0.439877 | 0.9074 |
| src/compiler/watch.ts:193:1 | createWatchStatusReporter | 4.744461 | 4.224380 | -0.520080 | 0.8904 |
| src/compiler/utilitiesPublic.ts:1574:1 | isImportAttributeName | 0.788700 | 0.257574 | -0.531126 | 0.3266 |
| src/compiler/utilities.ts:2036:1 | canUseOriginalText | 4.703107 | 4.153861 | -0.549246 | 0.8832 |
| src/compiler/sys.ts:1980:1 |  | 5.064602 | 4.513983 | -0.550619 | 0.8913 |
| src/compiler/moduleNameResolver.ts:147:1 | removeIgnoredPackageId | 4.545845 | 3.958099 | -0.587746 | 0.8707 |
| src/compiler/program.ts:4855:1 | getImpliedNodeFormatForEmitWorker | 5.001556 | 4.400682 | -0.600874 | 0.8799 |
| src/compiler/tsbuildPublic.ts:287:1 | createBuilderStatusReporter | 4.712722 | 4.103314 | -0.609409 | 0.8707 |
| src/compiler/moduleSpecifiers.ts:433:1 | getLocalModuleSpecifierBetweenFileNames | 5.183942 | 4.515168 | -0.668774 | 0.8710 |
| src/compiler/moduleNameResolver.ts:3424:1 | useCaseSensitiveFileNames | 5.045928 | 4.306532 | -0.739396 | 0.8535 |
| src/compiler/utilities.ts:2638:1 | isExternalOrCommonJsModule | 5.213570 | 4.303833 | -0.909737 | 0.8255 |
| src/compiler/utilities.ts:6144:1 | containsInvalidEscapeFlag | 5.235791 | 4.311014 | -0.924777 | 0.8234 |
| src/compiler/utilities.ts:9402:1 | getCompilerOptionValue | 5.249541 | 4.201031 | -1.048510 | 0.8003 |
