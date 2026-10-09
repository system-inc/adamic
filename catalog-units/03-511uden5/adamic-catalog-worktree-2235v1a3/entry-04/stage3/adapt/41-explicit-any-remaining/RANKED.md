# Other source edit candidates

These are observed refusal rows with a source-level adaptation route, ranked by count. They are proposals, not proved adaptations. No row here was adapted in 41. Complete exact reason strings and sites are in evidence/before-census.json. Variance remains with the other worker.

| Rank | Family | Observed sites |
| ---: | --- | ---: |
| 1 | Method receiver captures | 195 |
| 2 | Explicit boolean conditions | 78 |
| 3 | a definite assignment assertion ! | 11 |
| 4 | a namespace | 11 |
| 5 | an index signature | 11 |
| 6 | a parameter property | 6 |
| 7 | var | 6 |
| 8 | a generator function | 5 |
| 9 | yield (generators) | 5 |

A declared expando-field refusal and a missing-return-annotation refusal are not observed as distinct reasons: zero rows of each. This does not prove the source lacks either problem. Unsupported return representations and body-less overload declarations are compiler lessons, not missing annotations.

The 1,396 unchecked-cast rows are mixed: some require truthful owner declarations, others have dynamic or staged-initialization contracts. Calling all 796 source-only fixes would be unsupported. The full ledger retains them for site review. Non-null assertions, comma operations, logical assignments and predicate-verification gaps remain compiler lessons under the supplied table rulings.

## Every candidate exact row

| Count | Exact reason | Sites |
| ---: | --- | --- |
| 29 | a value as a condition | src/compiler/builder.ts:642:13<br>src/compiler/commandLineParser.ts:4176:9<br>src/compiler/core.ts:1894:13<br>src/compiler/factory/emitNode.ts:273:9<br>src/compiler/moduleNameResolver.ts:480:9<br>src/compiler/semver.ts:216:13<br>src/compiler/moduleSpecifiers.ts:699:9<br>src/compiler/program.ts:768:30<br>src/compiler/scanner.ts:469:12<br>src/compiler/scanner.ts:763:9<br>src/compiler/scanner.ts:968:9<br>src/compiler/tsbuildPublic.ts:206:12<br>src/compiler/utilities.ts:2339:12<br>src/compiler/utilities.ts:2461:22<br>src/compiler/utilities.ts:2475:22<br>src/compiler/utilities.ts:2997:9<br>src/compiler/utilities.ts:3057:13<br>src/compiler/utilities.ts:3295:9<br>src/compiler/utilities.ts:4751:9<br>src/compiler/utilities.ts:5064:12<br>src/compiler/utilities.ts:6484:12<br>src/compiler/utilities.ts:7355:9<br>src/compiler/utilities.ts:8745:22<br>src/compiler/utilities.ts:8769:12<br>src/compiler/utilities.ts:9896:9<br>src/compiler/utilities.ts:10159:12<br>src/compiler/utilitiesPublic.ts:1037:9<br>src/compiler/watch.ts:114:69<br>src/compiler/watch.ts:790:9 |
| 17 | a string as a condition | src/compiler/emitter.ts:484:9<br>src/compiler/emitter.ts:554:12<br>src/compiler/emitter.ts:697:9<br>src/compiler/emitter.ts:725:9<br>src/compiler/executeCommandLine.ts:241:31<br>src/compiler/executeCommandLine.ts:448:9<br>src/compiler/moduleNameResolver.ts:2135:13<br>src/compiler/moduleSpecifiers.ts:398:9<br>src/compiler/parser.ts:10563:9<br>src/compiler/path.ts:385:12<br>src/compiler/path.ts:857:12<br>src/compiler/path.ts:872:9<br>src/compiler/program.ts:1142:33<br>src/compiler/utilities.ts:910:12<br>src/compiler/utilities.ts:6596:12<br>src/compiler/utilities.ts:7150:9<br>src/compiler/utilities.ts:9431:12 |
| 15 | a boolean \| undefined as a condition | src/compiler/builder.ts:739:9<br>src/compiler/core.ts:2080:12<br>src/compiler/core.ts:2250:9<br>src/compiler/core.ts:2431:12<br>src/compiler/executeCommandLine.ts:1147:9<br>src/compiler/moduleNameResolver.ts:2061:40<br>src/compiler/program.ts:888:9<br>src/compiler/scanner.ts:477:13<br>src/compiler/scanner.ts:658:21<br>src/compiler/tsbuildPublic.ts:289:22<br>src/compiler/utilities.ts:5046:19<br>src/compiler/utilities.ts:5647:20<br>src/compiler/utilities.ts:5954:20<br>src/compiler/utilities.ts:7477:13<br>src/compiler/watch.ts:194:12 |
| 14 | a method read as a value (liftToBlock would lose its object, and this with it) | src/compiler/transformers/es2018.ts:477:93<br>src/compiler/transformers/module/module.ts:1015:75<br>src/compiler/transformers/module/module.ts:1028:94<br>src/compiler/transformers/module/module.ts:1041:79<br>src/compiler/transformers/module/module.ts:1042:79<br>src/compiler/transformers/module/system.ts:1422:75<br>src/compiler/transformers/module/system.ts:1435:94<br>src/compiler/transformers/module/system.ts:1448:79<br>src/compiler/transformers/module/system.ts:1449:79<br>src/compiler/visitorPublic.ts:545:61<br>src/compiler/visitorPublic.ts:1301:86<br>src/compiler/visitorPublic.ts:1302:67<br>src/compiler/visitorPublic.ts:1376:82<br>src/compiler/visitorPublic.ts:1392:82 |
| 12 | a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | src/compiler/emitter.ts:2233:199<br>src/compiler/emitter.ts:2605:64<br>src/compiler/emitter.ts:2615:103<br>src/compiler/emitter.ts:2711:86<br>src/compiler/emitter.ts:2719:85<br>src/compiler/emitter.ts:2979:41<br>src/compiler/emitter.ts:3406:119<br>src/compiler/emitter.ts:3997:41<br>src/compiler/emitter.ts:4068:37<br>src/compiler/emitter.ts:4077:62<br>src/compiler/emitter.ts:4084:45<br>src/compiler/emitter.ts:4094:64 |
| 12 | a number as a condition | src/compiler/factory/nodeFactory.ts:493:25<br>src/compiler/factory/utilities.ts:1740:9<br>src/compiler/moduleNameResolver.ts:195:9<br>src/compiler/moduleNameResolver.ts:204:9<br>src/compiler/moduleNameResolver.ts:301:12<br>src/compiler/parser.ts:485:12<br>src/compiler/utilities.ts:8041:12<br>src/compiler/utilities.ts:8069:12<br>src/compiler/utilities.ts:8267:9<br>src/compiler/utilities.ts:8281:12<br>src/compiler/utilities.ts:8884:9<br>src/compiler/utilities.ts:12439:17 |
| 11 | a definite assignment assertion ! | src/compiler/parser.ts:6935:24<br>src/compiler/scanner.ts:827:19<br>src/compiler/scanner.ts:828:19<br>src/compiler/scanner.ts:829:20<br>src/compiler/scanner.ts:830:34<br>src/compiler/scanner.ts:1050:19<br>src/compiler/scanner.ts:3097:24<br>src/compiler/utilities.ts:6834:22<br>src/compiler/utilities.ts:6835:23<br>src/compiler/utilities.ts:6836:20<br>src/compiler/utilities.ts:6837:20 |
| 11 | a namespace | src/compiler/builderState.ts:100:1<br>src/compiler/checker.ts:54223:1<br>src/compiler/checker.ts:54236:1<br>src/compiler/debug.ts:113:1<br>src/compiler/debug.ts:137:5<br>src/compiler/factory/utilities.ts:1273:1<br>src/compiler/parser.ts:1437:1<br>src/compiler/parser.ts:8790:5<br>src/compiler/parser.ts:9946:1<br>src/compiler/tracing.ts:38:1<br>src/compiler/tsbuild.ts:58:1 |
| 11 | an index signature | src/compiler/commandLineParser.ts:1911:5<br>src/compiler/corePublic.ts:14:5<br>src/compiler/executeCommandLine.ts:405:39<br>src/compiler/parser.ts:10723:31<br>src/compiler/parser.ts:10781:89<br>src/compiler/parser.ts:10785:21<br>src/compiler/tracing.ts:55:9<br>src/compiler/tsbuildPublic.ts:166:5<br>src/compiler/types.ts:7582:5<br>src/compiler/types.ts:7593:5<br>src/compiler/types.ts:7601:5 |
| 7 | a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | src/compiler/emitter.ts:2238:46<br>src/compiler/emitter.ts:2639:41<br>src/compiler/emitter.ts:2690:41<br>src/compiler/emitter.ts:2705:41<br>src/compiler/emitter.ts:2730:34<br>src/compiler/emitter.ts:2988:41<br>src/compiler/emitter.ts:3003:41 |
| 6 | a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | src/compiler/emitter.ts:2743:41<br>src/compiler/emitter.ts:2786:41<br>src/compiler/emitter.ts:2792:41<br>src/compiler/emitter.ts:2798:41<br>src/compiler/emitter.ts:2804:41<br>src/compiler/emitter.ts:2812:38 |
| 6 | a method read as a value (readFile would lose its object, and this with it) | src/compiler/moduleSpecifiers.ts:812:27<br>src/compiler/moduleSpecifiers.ts:1141:10<br>src/compiler/moduleSpecifiers.ts:1187:30<br>src/compiler/moduleSpecifiers.ts:1256:136<br>src/compiler/program.ts:525:30<br>src/compiler/program.ts:546:5 |
| 6 | a parameter property | src/compiler/factory/utilities.ts:1415:9<br>src/compiler/factory/utilities.ts:1416:9<br>src/compiler/factory/utilities.ts:1417:9<br>src/compiler/factory/utilities.ts:1418:9<br>src/compiler/factory/utilities.ts:1419:9<br>src/compiler/factory/utilities.ts:1420:9 |
| 6 | var | src/compiler/binder.ts:513:5<br>src/compiler/emitter.ts:765:5<br>src/compiler/emitter.ts:1215:5<br>src/compiler/scanner.ts:1035:5<br>src/compiler/utilities.ts:674:5<br>src/compiler/utilities.ts:6302:5 |
| 5 | a generator function | src/compiler/core.ts:332:1<br>src/compiler/core.ts:437:1<br>src/compiler/core.ts:509:1<br>src/compiler/core.ts:538:1<br>src/compiler/core.ts:1043:1 |
| 5 | a method read as a value (trace would lose its object, and this with it) | src/compiler/moduleNameResolver.ts:119:5<br>src/compiler/moduleNameResolver.ts:124:49<br>src/compiler/watch.ts:746:27<br>src/compiler/watch.ts:747:88<br>src/compiler/watch.ts:994:14 |
| 5 | yield (generators) | src/compiler/core.ts:334:9<br>src/compiler/core.ts:441:9<br>src/compiler/core.ts:513:13<br>src/compiler/core.ts:539:5<br>src/compiler/core.ts:1045:9 |
| 4 | a method read as a value (fileExists would lose its object, and this with it) | src/compiler/moduleSpecifiers.ts:1187:10<br>src/compiler/moduleSpecifiers.ts:1346:10<br>src/compiler/program.ts:526:32<br>src/compiler/program.ts:573:5 |
| 4 | a method read as a value (getSourceFile would lose its object, and this with it) | src/compiler/tsbuildPublic.ts:749:35<br>src/compiler/tsbuildPublic.ts:765:5<br>src/compiler/watch.ts:828:35<br>src/compiler/watch.ts:829:5 |
| 3 | a method read as a value (afterProgramCreate would lose its object, and this with it) | src/compiler/executeCommandLine.ts:999:35<br>src/compiler/executeCommandLine.ts:1000:5<br>src/compiler/watch.ts:875:5 |
| 3 | a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | src/compiler/executeCommandLine.ts:975:5<br>src/compiler/tsbuildPublic.ts:1363:13<br>src/compiler/watch.ts:997:9 |
| 3 | a method read as a value (getCanonicalFileName would lose its object, and this with it) | src/compiler/emitter.ts:1123:25<br>src/compiler/tsbuildPublic.ts:556:74<br>src/compiler/tsbuildPublic.ts:2284:82 |
| 3 | a method read as a value (now would lose its object, and this with it) | src/compiler/performanceCore.ts:102:100<br>src/compiler/tsbuildPublic.ts:206:12<br>src/compiler/watch.ts:178:13 |
| 3 | a method read as a value (onWatchStatusChange would lose its object, and this with it) | src/compiler/executeCommandLine.ts:865:37<br>src/compiler/executeCommandLine.ts:867:9<br>src/compiler/watch.ts:884:17 |
| 3 | a union of differently held members as a condition | src/compiler/moduleNameResolver.ts:3171:9<br>src/compiler/path.ts:438:9<br>src/compiler/transformer.ts:128:9 |
| 2 | a method read as a value (clearTimeout would lose its object, and this with it) | src/compiler/tsbuildPublic.ts:2062:39<br>src/compiler/watch.ts:680:41 |
| 2 | a method read as a value (createDirectory would lose its object, and this with it) | src/compiler/program.ts:528:37<br>src/compiler/program.ts:615:13 |
| 2 | a method read as a value (emitNodeWithNotification would lose its object, and this with it) | src/compiler/emitter.ts:867:25<br>src/compiler/emitter.ts:943:29 |
| 2 | a method read as a value (enableCPUProfiler would lose its object, and this with it) | src/compiler/executeCommandLine.ts:763:48<br>src/compiler/executeCommandLine.ts:787:51 |
| 2 | a method read as a value (getBuildInfo would lose its object, and this with it) | src/compiler/builder.ts:1686:5<br>src/compiler/watchPublic.ts:110:9 |
| 2 | a method read as a value (getEnvironmentVariable would lose its object, and this with it) | src/compiler/sys.ts:135:10<br>src/compiler/sys.ts:1972:12 |
| 2 | a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | src/compiler/moduleSpecifiers.ts:1235:40<br>src/compiler/resolutionCache.ts:544:10 |
| 2 | a method read as a value (getParsedCommandLine would lose its object, and this with it) | src/compiler/tsbuildPublic.ts:589:9<br>src/compiler/tsbuildPublic.ts:2041:9 |
| 2 | a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | src/compiler/builder.ts:1708:9<br>src/compiler/builder.ts:1711:9 |
| 2 | a method read as a value (hasGlobalName would lose its object, and this with it) | src/compiler/emitter.ts:864:28<br>src/compiler/emitter.ts:940:32 |
| 2 | a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | src/compiler/emitter.ts:868:40<br>src/compiler/emitter.ts:944:44 |
| 2 | a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | src/compiler/emitter.ts:2956:39<br>src/compiler/emitter.ts:2962:40 |
| 2 | a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | src/compiler/factory/nodeFactory.ts:2570:86<br>src/compiler/factory/nodeFactory.ts:2575:63 |
| 2 | a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | src/compiler/factory/nodeFactory.ts:2560:79<br>src/compiler/factory/nodeFactory.ts:2565:63 |
| 2 | a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | src/compiler/emitter.ts:2414:32<br>src/compiler/emitter.ts:2496:31 |
| 2 | a method read as a value (setPrototypeOf would lose its object, and this with it) | src/compiler/debug.ts:554:24<br>src/compiler/debug.ts:594:24 |
| 2 | a method read as a value (setTimeout would lose its object, and this with it) | src/compiler/tsbuildPublic.ts:2062:10<br>src/compiler/watch.ts:679:39 |
| 2 | a method read as a value (substituteNode would lose its object, and this with it) | src/compiler/emitter.ts:869:29<br>src/compiler/emitter.ts:945:33 |
| 2 | a method read as a value (toKey would lose its object, and this with it) | src/compiler/transformers/utilities.ts:427:119<br>src/compiler/transformers/utilities.ts:431:119 |
| 2 | a method read as a value (trackSymbol would lose its object, and this with it) | src/compiler/checker.ts:54345:33<br>src/compiler/checker.ts:54349:13 |
| 2 | a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | src/compiler/moduleSpecifiers.ts:571:61<br>src/compiler/utilities.ts:6484:12 |
| 2 | a method read as a value (watchDirectory would lose its object, and this with it) | src/compiler/executeCommandLine.ts:801:28<br>src/compiler/watch.ts:678:43 |
| 2 | a method read as a value (watchFile would lose its object, and this with it) | src/compiler/executeCommandLine.ts:801:10<br>src/compiler/watch.ts:677:38 |
| 2 | a method read as a value (writeFile would lose its object, and this with it) | src/compiler/sys.ts:1381:31<br>src/compiler/sys.ts:1382:5 |
| 2 | a number \| undefined as a condition | src/compiler/builder.ts:1630:9<br>src/compiler/program.ts:941:17 |
| 1 | a method read as a value (add would lose its object, and this with it) | src/compiler/core.ts:1544:5 |
| 1 | a method read as a value (base64decode would lose its object, and this with it) | src/compiler/utilities.ts:7750:17 |
| 1 | a method read as a value (base64encode would lose its object, and this with it) | src/compiler/utilities.ts:7742:17 |
| 1 | a method read as a value (clearScreen would lose its object, and this with it) | src/compiler/watch.ts:148:9 |
| 1 | a method read as a value (compare would lose its object, and this with it) | src/compiler/core.ts:2100:26 |
| 1 | a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | src/compiler/transformers/destructuring.ts:607:63 |
| 1 | a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | src/compiler/transformers/destructuring.ts:617:64 |
| 1 | a method read as a value (createComma would lose its object, and this with it) | src/compiler/factory/nodeFactory.ts:6756:39 |
| 1 | a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | src/compiler/parser.ts:4840:99 |
| 1 | a method read as a value (createJSDocClassTag would lose its object, and this with it) | src/compiler/parser.ts:9117:53 |
| 1 | a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | src/compiler/parser.ts:9136:53 |
| 1 | a method read as a value (createJSDocLink would lose its object, and this with it) | src/compiler/parser.ts:9322:54 |
| 1 | a method read as a value (createJSDocLinkCode would lose its object, and this with it) | src/compiler/parser.ts:9323:49 |
| 1 | a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | src/compiler/parser.ts:9324:23 |
| 1 | a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | src/compiler/parser.ts:9132:53 |
| 1 | a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | src/compiler/parser.ts:9123:53 |
| 1 | a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | src/compiler/parser.ts:9126:53 |
| 1 | a method read as a value (createJSDocPublicTag would lose its object, and this with it) | src/compiler/parser.ts:9120:53 |
| 1 | a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | src/compiler/parser.ts:9129:53 |
| 1 | a method read as a value (createUnionTypeNode would lose its object, and this with it) | src/compiler/parser.ts:4844:97 |
| 1 | a method read as a value (deleteFile would lose its object, and this with it) | src/compiler/executeCommandLine.ts:848:81 |
| 1 | a method read as a value (directoryExists would lose its object, and this with it) | src/compiler/moduleNameResolver.ts:820:9 |
| 1 | a method read as a value (emit would lose its object, and this with it) | src/compiler/builder.ts:1704:5 |
| 1 | a method read as a value (emitBuildInfo would lose its object, and this with it) | src/compiler/builder.ts:1713:9 |
| 1 | a method read as a value (emitNextAffectedFile would lose its object, and this with it) | src/compiler/builder.ts:1712:9 |
| 1 | a method read as a value (fill would lose its object, and this with it) | src/compiler/debug.ts:1232:17 |
| 1 | a method read as a value (getAllDependencies would lose its object, and this with it) | src/compiler/builder.ts:1696:5 |
| 1 | a method read as a value (getCurrentDirectory would lose its object, and this with it) | src/compiler/moduleNameResolver.ts:488:14 |
| 1 | a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | src/compiler/builder.ts:1703:5 |
| 1 | a method read as a value (getDirectories would lose its object, and this with it) | src/compiler/moduleNameResolver.ts:820:33 |
| 1 | a method read as a value (getMemoryUsage would lose its object, and this with it) | src/compiler/executeCommandLine.ts:1170:28 |
| 1 | a method read as a value (getModifiedTime would lose its object, and this with it) | src/compiler/executeCommandLine.ts:848:10 |
| 1 | a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | src/compiler/moduleSpecifiers.ts:699:9 |
| 1 | a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | src/compiler/moduleNameResolver.ts:1356:5 |
| 1 | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | src/compiler/scanner.ts:469:12 |
| 1 | a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | src/compiler/builder.ts:1702:5 |
| 1 | a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | src/compiler/resolutionCache.ts:554:32 |
| 1 | a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | src/compiler/builder.ts:1695:5 |
| 1 | a method read as a value (hasOwnProperty would lose its object, and this with it) | src/compiler/core.ts:1256:24 |
| 1 | a method read as a value (log would lose its object, and this with it) | src/compiler/sourcemap.ts:742:21 |
| 1 | a method read as a value (nonEscapingWrite would lose its object, and this with it) | src/compiler/emitter.ts:4911:13 |
| 1 | a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | src/compiler/emitter.ts:2454:30 |
| 1 | a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | src/compiler/emitter.ts:2779:39 |
| 1 | a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | src/compiler/emitter.ts:2952:40 |
| 1 | a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | src/compiler/emitter.ts:2450:77 |
| 1 | a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | src/compiler/emitter.ts:2446:70 |
| 1 | a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | src/compiler/emitter.ts:2427:74 |
| 1 | a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | src/compiler/emitter.ts:2196:41 |
| 1 | a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | src/compiler/emitter.ts:3748:17 |
| 1 | a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | src/compiler/emitter.ts:3065:41 |
| 1 | a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | src/compiler/emitter.ts:2717:41 |
| 1 | a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | src/compiler/emitter.ts:2458:32 |
| 1 | a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | src/compiler/emitter.ts:1278:40 |
| 1 | a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | src/compiler/emitter.ts:2835:38 |
| 1 | a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | src/compiler/emitter.ts:2490:13 |
| 1 | a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | src/compiler/emitter.ts:2491:13 |
| 1 | a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | src/compiler/emitter.ts:2441:25 |
| 1 | a method read as a value (realpath would lose its object, and this with it) | src/compiler/moduleNameResolver.ts:1975:10 |
| 1 | a method read as a value (releaseProgram would lose its object, and this with it) | src/compiler/builder.ts:1705:5 |
| 1 | a method read as a value (remove would lose its object, and this with it) | src/compiler/core.ts:1545:5 |
| 1 | a method read as a value (repeat would lose its object, and this with it) | src/compiler/debug.ts:1244:17 |
| 1 | a method read as a value (replace would lose its object, and this with it) | src/compiler/utilities.ts:11193:92 |
| 1 | a method read as a value (reportCyclicStructureError would lose its object, and this with it) | src/compiler/checker.ts:54382:13 |
| 1 | a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | src/compiler/checker.ts:54361:13 |
| 1 | a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | src/compiler/checker.ts:54375:13 |
| 1 | a method read as a value (reportInferenceFallback would lose its object, and this with it) | src/compiler/checker.ts:54421:13 |
| 1 | a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | src/compiler/checker.ts:54389:13 |
| 1 | a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | src/compiler/checker.ts:54410:13 |
| 1 | a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | src/compiler/checker.ts:54403:13 |
| 1 | a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | src/compiler/checker.ts:54368:13 |
| 1 | a method read as a value (reportTruncationError would lose its object, and this with it) | src/compiler/checker.ts:54396:13 |
| 1 | a method read as a value (setModifiedTime would lose its object, and this with it) | src/compiler/executeCommandLine.ts:848:34 |
| 1 | a method read as a value (toString would lose its object, and this with it) | src/compiler/debug.ts:376:26 |
| 1 | a method read as a value (writeOutputIsTTY would lose its object, and this with it) | src/compiler/executeCommandLine.ts:166:14 |
