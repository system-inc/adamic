# Stage 3 optional-widening sites

415 distinct AST locations in 42 files; 79 implementation files scanned in the compiler project.
Full expression text and type strings are in [stage3-sites.jsonl](stage3-sites.jsonl).

| File:line:column | AST kind | Missing property | Source S | Target T |
|---|---|---|---|---|
| binder.ts:738:67 | PropertyAccessExpression | `id` | `never` | `ArrayBindingPattern` |
| binder.ts:3251:52 | Identifier | `equalsToken` | `never` | `ShorthandPropertyAssignment` |
| binder.ts:3293:68 | Identifier | `id` | `never` | `Expression` |
| binder.ts:3313:64 | Identifier | `id` | `never` | `Expression` |
| binder.ts:3426:124 | PropertyAccessExpression | `_propertyAccessExpressionLikeQualifiedNameBrand` | `never` | `PropertyAccessEntityNameExpression` |
| binder.ts:3427:55 | Identifier | `id` | `never` | `Expression` |
| binder.ts:3705:86 | Identifier | `id` | `never` | `Node` |
| builder.ts:575:36 | CallExpression | `reportsUnnecessary` | `DiagnosticRelatedInformation` | `Diagnostic` |
| builder.ts:1505:48 | CallExpression | `reportsUnnecessary` | `ReusableDiagnosticRelatedInformation` | `ReusableDiagnostic` |
| checker.ts:1881:121 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:2123:133 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:2815:109 | Identifier | `reportsUnnecessary` | `DiagnosticRelatedInformation` | `Diagnostic` |
| checker.ts:2815:174 | Identifier | `reportsUnnecessary` | `DiagnosticRelatedInformation` | `Diagnostic` |
| checker.ts:3053:43 | Identifier | `id` | `never` | `Node` |
| checker.ts:3054:66 | Identifier | `id` | `never` | `Node` |
| checker.ts:3185:56 | Identifier | `id` | `never` | `Node` |
| checker.ts:3186:71 | Identifier | `id` | `never` | `Node` |
| checker.ts:3186:84 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:3809:34 | CallExpression | `source` | `ResolvedRefAndOutputDts` | `ResolvedRefAndSource` |
| checker.ts:3809:79 | CallExpression | `outputDts` | `ResolvedRefAndSource` | `ResolvedRefAndOutputDts` |
| checker.ts:4745:17 | PropertyAccessExpression | `id` | `never` | `Identifier` |
| checker.ts:5568:16 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:5917:28 | Identifier | `errorModuleName` | `SymbolVisibilityResult` | `SymbolAccessibilityResult` |
| checker.ts:6942:69 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:6948:87 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:6954:97 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:6960:54 | Identifier | `constraint` | `Type` | `TypeParameter` |
| checker.ts:7122:128 | CallExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:7133:90 | CallExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:7164:93 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:7620:68 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:7635:37 | PropertyAccessExpression | `id` | `never` | `BigIntLiteral` |
| checker.ts:7688:90 | AsExpression | `name` | `Declaration` | `NamedDeclaration` |
| checker.ts:9267:95 | Identifier | `constraint` | `Type` | `TypeParameter` |
| checker.ts:9535:29 | Identifier | `id` | `never` | `NamedExports` |
| checker.ts:9539:33 | PropertyAccessExpression | `id` | `never` | `NamedExports` |
| checker.ts:11157:106 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:11161:101 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:11162:24 | CallExpression | `id` | `never` | `PrivateIdentifier` |
| checker.ts:13234:71 | Identifier | `constraint` | `Type` | `TypeParameter` |
| checker.ts:13266:58 | Identifier | `constraint` | `Type` | `TypeParameter` |
| checker.ts:13286:47 | CallExpression | `constraint` | `ObjectType` | `TypeParameter` |
| checker.ts:13602:16 | BinaryExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:13970:83 | Identifier | `id` | `never` | `LateBoundName` |
| checker.ts:14505:34 | Identifier | `members` | `UnionType` | `ObjectType` |
| checker.ts:14567:34 | Identifier | `members` | `IntersectionType` | `ObjectType` |
| checker.ts:14959:132 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:15025:22 | AsExpression | `members` | `IntersectionType` | `ObjectType` |
| checker.ts:15028:27 | AsExpression | `members` | `IntersectionType` | `ObjectType` |
| checker.ts:15034:27 | AsExpression | `members` | `IntersectionType` | `ObjectType` |
| checker.ts:15035:49 | AsExpression | `target` | `IntersectionType` | `AnonymousType` |
| checker.ts:15149:84 | AsExpression | `constraint` | `InstantiableType` | `TypeParameter` |
| checker.ts:15166:59 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:15182:71 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:15236:84 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:15264:54 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:15266:54 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:15299:58 | AsExpression | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:15349:68 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:15371:67 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:15372:25 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:16205:31 | Identifier | `id` | `never` | `JSDocThisTag` |
| checker.ts:16568:39 | AsExpression | `target` | `ObjectType` | `AnonymousType` |
| checker.ts:16597:33 | NonNullExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:16961:69 | CallExpression | `constraint` | `IntrinsicType` | `TypeParameter` |
| checker.ts:16972:93 | BinaryExpression | `constraint` | `IntrinsicType` | `TypeParameter` |
| checker.ts:18900:51 | Identifier | `resolvedBaseConstraint` | `MappedType` | `InstantiableType` |
| checker.ts:18968:165 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:18987:81 | AsExpression | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:19204:58 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:19938:30 | CallExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:19947:47 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:19948:47 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:20007:71 | PropertyAccessExpression | `id` | `never` | `Expression` |
| checker.ts:20668:76 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:20740:72 | NonNullExpression | `typeArguments` | `ArrayTypeNode` | `NodeWithTypeArguments` |
| checker.ts:20867:24 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:20893:98 | Identifier | `target` | `IntrinsicType` | `AnonymousType` |
| checker.ts:20952:16 | ConditionalExpression | `members` | `IntrinsicType` | `ObjectType` |
| checker.ts:20952:43 | Identifier | `members` | `IntrinsicType` | `ObjectType` |
| checker.ts:20967:24 | AsExpression | `target` | `ObjectType` | `AnonymousType` |
| checker.ts:21339:119 | Identifier | `skipLogging` | `{ errors?: Diagnostic[] \| undefined; }` | `ErrorOutputContainer` |
| checker.ts:21482:139 | Identifier | `skipLogging` | `{ errors?: Diagnostic[] \| undefined; }` | `ErrorOutputContainer` |
| checker.ts:21575:145 | Identifier | `skipLogging` | `{ errors?: Diagnostic[] \| undefined; }` | `ErrorOutputContainer` |
| checker.ts:21578:134 | Identifier | `skipLogging` | `{ errors?: Diagnostic[] \| undefined; }` | `ErrorOutputContainer` |
| checker.ts:21665:145 | Identifier | `skipLogging` | `{ errors?: Diagnostic[] \| undefined; }` | `ErrorOutputContainer` |
| checker.ts:21668:134 | Identifier | `skipLogging` | `{ errors?: Diagnostic[] \| undefined; }` | `ErrorOutputContainer` |
| checker.ts:22798:79 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:22931:116 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:22932:59 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:23160:66 | AsExpression | `resolvedReducedType` | `UnionOrIntersectionType` | `UnionType` |
| checker.ts:23674:67 | Identifier | `constraint` | `Type` | `TypeParameter` |
| checker.ts:23680:71 | Identifier | `constraint` | `Type` | `TypeParameter` |
| checker.ts:23898:60 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:23910:69 | PropertyAccessExpression | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:24004:119 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:24384:40 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:24915:52 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:25113:86 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:25200:31 | Identifier | `constraint` | `InterfaceType` | `TypeParameter` |
| checker.ts:25200:44 | ElementAccessExpression | `default` | `IndexedAccessType` | `TypeParameter` |
| checker.ts:26159:40 | AsExpression | `name` | `Declaration` | `NamedDeclaration` |
| checker.ts:26515:31 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:26915:182 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:27252:64 | Identifier | `resolvedBaseConstraint` | `Type` | `InstantiableType` |
| checker.ts:31797:31 | Identifier | `constraint` | `InterfaceType` | `TypeParameter` |
| checker.ts:32168:37 | Identifier | `members` | `IntrinsicType` | `ObjectType` |
| checker.ts:32169:17 | ConditionalExpression | `members` | `IntrinsicType` | `ObjectType` |
| checker.ts:32170:17 | Identifier | `members` | `IntrinsicType` | `ObjectType` |
| checker.ts:32496:51 | Identifier | `localSymbol` | `never` | `Declaration` |
| checker.ts:33480:121 | PropertyAccessExpression | `id` | `never` | `ArrayBindingPattern` |
| checker.ts:33485:111 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:33486:62 | PropertyAccessExpression | `id` | `never` | `ComputedPropertyName` |
| checker.ts:33491:85 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:34595:31 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:34595:106 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:34595:166 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:34614:57 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:34827:77 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:35380:128 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:35790:48 | NonNullExpression | `constraint` | `Type` | `TypeParameter` |
| checker.ts:36738:62 | AsExpression | `reportsUnnecessary` | `DiagnosticRelatedInformation` | `Diagnostic` |
| checker.ts:37997:31 | AsExpression | `syntheticType` | `Type` | `SyntheticDefaultModuleType` |
| checker.ts:38009:31 | AsExpression | `syntheticType` | `Type` | `SyntheticDefaultModuleType` |
| checker.ts:38334:16 | ConditionalExpression | `members` | `IntrinsicType` | `ObjectType` |
| checker.ts:38334:79 | Identifier | `members` | `IntrinsicType` | `ObjectType` |
| checker.ts:38488:20 | ConditionalExpression | `id` | `never` | `ArrayBindingPattern` |
| checker.ts:38497:16 | ConditionalExpression | `id` | `never` | `ArrayBindingPattern` |
| checker.ts:42245:92 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:44510:134 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:44924:67 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:44962:67 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:44963:183 | PropertyAccessExpression | `id` | `never` | `ArrayBindingPattern` |
| checker.ts:47073:135 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:47283:40 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:47284:36 | Identifier | `id` | `never` | `Node` |
| checker.ts:47296:33 | Identifier | `id` | `never` | `Node` |
| checker.ts:47297:33 | Identifier | `id` | `never` | `Node` |
| checker.ts:47298:22 | Identifier | `id` | `never` | `Node` |
| checker.ts:47301:13 | ConditionalExpression | `id` | `never` | `Node` |
| checker.ts:48395:45 | Identifier | `id` | `never` | `Node` |
| checker.ts:48408:28 | Identifier | `id` | `never` | `Node` |
| checker.ts:48408:113 | Identifier | `id` | `never` | `Node` |
| checker.ts:48409:61 | Identifier | `id` | `never` | `Node` |
| checker.ts:48410:40 | Identifier | `id` | `never` | `Node` |
| checker.ts:48433:60 | Identifier | `id` | `never` | `Node` |
| checker.ts:48454:23 | Identifier | `id` | `never` | `Node` |
| checker.ts:48459:100 | Identifier | `id` | `never` | `Node` |
| checker.ts:48462:25 | Identifier | `id` | `never` | `Node` |
| checker.ts:48472:57 | Identifier | `id` | `never` | `Node` |
| checker.ts:48520:33 | Identifier | `id` | `never` | `Node` |
| checker.ts:48521:72 | Identifier | `id` | `never` | `Node` |
| checker.ts:48523:27 | Identifier | `id` | `never` | `Node` |
| checker.ts:48523:69 | Identifier | `id` | `never` | `Node` |
| checker.ts:48529:72 | Identifier | `id` | `never` | `Node` |
| checker.ts:48535:27 | Identifier | `id` | `never` | `Node` |
| checker.ts:48540:58 | Identifier | `id` | `never` | `Node` |
| checker.ts:48547:31 | Identifier | `id` | `never` | `Node` |
| checker.ts:48552:35 | Identifier | `id` | `never` | `Node` |
| checker.ts:49331:52 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| checker.ts:50318:76 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:51475:76 | PropertyAccessExpression | `id` | `never` | `Node` |
| checker.ts:51487:41 | PropertyAccessExpression | `id` | `never` | `BigIntLiteral` |
| checker.ts:52410:57 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| checker.ts:52411:55 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| checker.ts:52739:39 | PropertyAccessExpression | `id` | `never` | `Node` |
| commandLineParser.ts:1993:117 | BinaryExpression | `watchFile` | `{}` | `WatchOptions` |
| commandLineParser.ts:2128:12 | CallExpression | `all` | `OptionsBase` | `CompilerOptions` |
| commandLineParser.ts:2251:9 | Identifier | `extendedSourceFiles` | `JsonSourceFile` | `TsConfigSourceFile` |
| commandLineParser.ts:2291:12 | ConditionalExpression | `extendedSourceFiles` | `JsonSourceFile` | `TsConfigSourceFile` |
| commandLineParser.ts:2679:9 | SpreadAssignment | `watchOptions` | `{ include: readonly string[] \| undefined; exclude: readonly string[] \| undefined; }` | `TSConfig & { watchOptions?: object \| undefined; }` |
| commandLineParser.ts:2697:79 | Identifier | `all` | `Record<string, CompilerOptionsValue>` | `CompilerOptions` |
| commandLineParser.ts:3536:13 | Identifier | `extendedSourceFiles` | `JsonSourceFile` | `TsConfigSourceFile` |
| commandLineParser.ts:3552:25 | Identifier | `extendedSourceFiles` | `JsonSourceFile` | `TsConfigSourceFile` |
| commandLineParser.ts:3557:188 | Identifier | `extendedSourceFiles` | `JsonSourceFile` | `TsConfigSourceFile` |
| commandLineParser.ts:3562:115 | Identifier | `extendedSourceFiles` | `JsonSourceFile` | `TsConfigSourceFile` |
| commandLineParser.ts:3767:87 | Identifier | `all` | `TypeAcquisition` | `CompilerOptions` |
| commandLineParser.ts:3785:33 | BinaryExpression | `all` | `{}` | `CompilerOptions` |
| emitter.ts:769:39 | Identifier | `omitTrailingSemicolon` | `CompilerOptions` | `PrinterOptions` |
| emitter.ts:1408:32 | Identifier | `skipTrivia` | `SourceFile` | `SourceMapSource` |
| emitter.ts:1444:66 | CallExpression | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| emitter.ts:3961:112 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| emitter.ts:3961:180 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| emitter.ts:6233:96 | Identifier | `getPositionOfLineAndCharacter` | `SourceMapSource` | `SourceFileLike` |
| executeCommandLine.ts:109:41 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| executeCommandLine.ts:824:9 | Identifier | `all` | `BuildOptions` | `CompilerOptions` |
| executeCommandLine.ts:859:66 | Identifier | `all` | `BuildOptions` | `CompilerOptions` |
| executeCommandLine.ts:860:44 | Identifier | `all` | `BuildOptions` | `CompilerOptions` |
| executeCommandLine.ts:889:62 | Identifier | `all` | `BuildOptions` | `CompilerOptions` |
| executeCommandLine.ts:890:39 | Identifier | `all` | `BuildOptions` | `CompilerOptions` |
| expressionToTypeNode.ts:438:64 | PropertyAccessExpression | `textSourceNode` | `never` | `StringLiteral` |
| expressionToTypeNode.ts:439:71 | CallExpression | `textSourceNode` | `never` | `StringLiteral` |
| expressionToTypeNode.ts:442:57 | CallExpression | `id` | `never` | `FalseLiteral` |
| factory/emitNode.ts:133:12 | BinaryExpression | `source` | `Node` | `SourceMapRange` |
| factory/emitNode.ts:133:45 | Identifier | `source` | `Node` | `SourceMapRange` |
| factory/nodeFactory.ts:7105:28 | Identifier | `id` | `never` | `Node` |
| factory/nodeFactory.ts:7105:63 | Identifier | `modifiers` | `never` | `ParameterDeclaration` |
| factory/nodeFactory.ts:7302:56 | PropertyAccessExpression | `id` | `never` | `Node` |
| factory/nodeFactory.ts:7302:104 | PropertyAccessExpression | `id` | `never` | `BigIntLiteral` |
| factory/nodeFactory.ts:7488:45 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| factory/utilities.ts:363:96 | Identifier | `id` | `never` | `BigIntLiteral` |
| factory/utilities.ts:368:58 | PropertyAccessExpression | `id` | `never` | `BigIntLiteral` |
| moduleNameResolver.ts:130:36 | AsExpression | `version` | `PackageJsonPathFields` | `PackageJson` |
| moduleNameResolver.ts:140:12 | BinaryExpression | `originalPath` | `{ path: string; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; }` | `Resolved` |
| moduleNameResolver.ts:2409:82 | PropertyAccessExpression | `version` | `PackageJsonPathFields` | `PackageJson` |
| moduleNameResolver.ts:2422:51 | PropertyAccessExpression | `version` | `PackageJsonPathFields` | `PackageJson` |
| moduleNameResolver.ts:2432:34 | AsExpression | `version` | `PackageJsonPathFields` | `PackageJson` |
| moduleNameResolver.ts:2494:56 | PropertyAccessExpression | `version` | `PackageJsonPathFields` | `PackageJson` |
| moduleNameResolver.ts:2497:93 | PropertyAccessExpression | `version` | `PackageJsonPathFields` | `PackageJson` |
| moduleNameResolver.ts:2498:116 | PropertyAccessExpression | `version` | `PackageJsonPathFields` | `PackageJson` |
| moduleSpecifiers.ts:200:71 | Identifier | `packageJsonScope` | `Pick<SourceFile, "fileName" \| "impliedNodeFormat">` | `Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">` |
| moduleSpecifiers.ts:243:63 | Identifier | `packageJsonScope` | `Pick<SourceFile, "fileName" \| "impliedNodeFormat">` | `Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">` |
| moduleSpecifiers.ts:881:141 | PropertyAccessExpression | `id` | `never` | `BigIntLiteral` |
| moduleSpecifiers.ts:1277:30 | SpreadAssignment | `packageRootPath` | `{ moduleFileToTry: string; }` | `{ moduleFileToTry: string; packageRootPath?: string \| undefined; blockedByExports?: true \| undefined; verbatimFromExports?: true \| undefined; }` |
| parser.ts:1854:40 | Identifier | `jsDocCache` | `JSDoc[]` | `JSDocArray` |
| parser.ts:2629:13 | ConditionalExpression | `rawText` | `NumericLiteral` | `TemplateLiteralLikeNode` |
| parser.ts:2630:13 | ConditionalExpression | `rawText` | `NumericLiteral` | `TemplateLiteralLikeNode` |
| parser.ts:3768:13 | ConditionalExpression | `rawText` | `NumericLiteral` | `TemplateLiteralLikeNode` |
| parser.ts:10032:36 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| parser.ts:10032:48 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| parser.ts:10104:37 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| parser.ts:10261:42 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| parser.ts:10641:111 | SpreadAssignment | `preserve` | `{ resolutionMode: ModuleKind.CommonJS \| ModuleKind.ESNext; }` | `FileReference` |
| parser.ts:10641:158 | SpreadAssignment | `resolutionMode` | `{ preserve: true; }` | `FileReference` |
| parser.ts:10644:104 | SpreadAssignment | `resolutionMode` | `{ preserve: true; }` | `FileReference` |
| parser.ts:10647:100 | SpreadAssignment | `resolutionMode` | `{ preserve: true; }` | `FileReference` |
| parser.ts:10745:53 | Identifier | `types` | `{ [index: string]: string \| { value: string; pos: number; end: number; }; }` | `{ types?: { value: string; pos: number; end: number; } \| undefined; } & { lib?: { value: string; pos: number; end: number; } \| undefined; } & { path?: { value: string; pos: number; end: number; } \| undefined; } & ... & ... & ...` |
| parser.ts:10777:45 | Identifier | `types` | `{ [index: string]: string; }` | `{ types?: { value: string; pos: number; end: number; } \| undefined; } & { lib?: { value: string; pos: number; end: number; } \| undefined; } & { path?: { value: string; pos: number; end: number; } \| undefined; } & ... & ... & ...` |
| program.ts:470:41 | Identifier | `omitTrailingSemicolon` | `CompilerOptions` | `PrinterOptions` |
| program.ts:670:67 | PropertyAccessExpression | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| program.ts:712:89 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| program.ts:713:87 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| program.ts:714:58 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| program.ts:732:57 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| program.ts:733:76 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| program.ts:767:89 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| program.ts:1578:67 | Identifier | `onUnRecoverableConfigFileDiagnostic` | `CompilerHost` | `CompilerHost & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter \| undefined; }` |
| program.ts:1608:42 | ArrowFunction | `failedLookupLocations` | `{ resolvedModule: ResolvedModuleFull; }` | `ResolvedModuleWithFailedLookupLocations` |
| program.ts:1609:13 | CallExpression | `failedLookupLocations` | `{ resolvedModule: ResolvedModuleFull; }` | `ResolvedModuleWithFailedLookupLocations` |
| program.ts:1654:13 | CallExpression | `failedLookupLocations` | `{ resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective \| undefined; }` | `ResolvedTypeReferenceDirectiveWithFailedLookupLocations` |
| program.ts:2972:42 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| program.ts:3926:57 | NonNullExpression | `isInvalidated` | `ResolvedModuleWithFailedLookupLocations` | `ResolutionWithFailedLookupLocations` |
| program.ts:4034:64 | Identifier | `extendedSourceFiles` | `JsonSourceFile` | `TsConfigSourceFile` |
| program.ts:4671:59 | BinaryExpression | `extendedSourceFiles` | `JsonSourceFile` | `TsConfigSourceFile` |
| resolutionCache.ts:997:27 | Identifier | `isInvalidated` | `ResolvedTypeReferenceDirectiveWithFailedLookupLocations` | `CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations` |
| resolutionCache.ts:1026:27 | Identifier | `isInvalidated` | `ResolvedModuleWithFailedLookupLocations` | `CachedResolvedModuleWithFailedLookupLocations` |
| resolutionCache.ts:1051:26 | CallExpression | `isInvalidated` | `ResolvedModuleWithFailedLookupLocations` | `CachedResolvedModuleWithFailedLookupLocations` |
| symbolWalker.ts:104:36 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| sys.ts:156:35 | BinaryExpression | `Low` | `{}` | `Partial<Levels>` |
| sys.ts:1619:77 | Identifier | `trace` | `System` | `ModuleResolutionHost` |
| sys.ts:1633:43 | Identifier | `bigint` | `{ readonly throwIfNoEntry: false; }` | `StatSyncOptions & { bigint?: false \| undefined; throwIfNoEntry: false; }` |
| tracing.ts:206:67 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| tracing.ts:207:65 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| transformer.ts:671:25 | ArrowFunction | `all` | `{}` | `CompilerOptions` |
| transformers/classFields.ts:694:20 | CallExpression | `id` | `never` | `BigIntLiteral` |
| transformers/classFields.ts:729:20 | CallExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/classFields.ts:753:20 | CallExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/classFields.ts:777:20 | CallExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/classFields.ts:932:41 | PropertyAccessExpression | `source` | `Expression` | `SourceMapRange` |
| transformers/classFields.ts:935:47 | PropertyAccessExpression | `source` | `Expression` | `SourceMapRange` |
| transformers/classFields.ts:2465:42 | Identifier | `source` | `ParameterDeclaration` | `SourceMapRange` |
| transformers/classFields.ts:2469:42 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/classFields.ts:2505:43 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/classFields.ts:2611:42 | PropertyAccessExpression | `source` | `Identifier` | `SourceMapRange` |
| transformers/classFields.ts:3306:50 | Identifier | `source` | `Identifier` | `SourceMapRange` |
| transformers/classThis.ts:122:59 | PropertyAccessExpression | `source` | `Identifier` | `SourceMapRange` |
| transformers/declarations.ts:825:40 | Identifier | `errorModuleName` | `SymbolVisibilityResult` | `SymbolAccessibilityResult` |
| transformers/declarations.ts:1252:55 | PropertyAccessExpression | `id` | `never` | `FalseLiteral` |
| transformers/declarations.ts:1252:71 | CallExpression | `id` | `never` | `FalseLiteral` |
| transformers/declarations.ts:1252:101 | PropertyAccessExpression | `id` | `never` | `LiteralExpression & StringLiteral` |
| transformers/declarations.ts:1264:70 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| transformers/declarations.ts:1264:139 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| transformers/declarations.ts:1328:56 | ArrowFunction | `typeName` | `{ diagnosticMessage: DiagnosticMessage; errorNode: ExportAssignment; }` | `SymbolAccessibilityDiagnostic` |
| transformers/declarations/diagnostics.ts:249:87 | PropertyAccessExpression | `id` | `never` | `Node` |
| transformers/destructuring.ts:607:63 | PropertyAccessExpression | `multiLine` | `never` | `ArrayLiteralExpression` |
| transformers/destructuring.ts:617:64 | PropertyAccessExpression | `propertyName` | `never` | `BindingElement` |
| transformers/es2015.ts:1504:17 | Identifier | `id` | `never` | `ArrayBindingPattern` |
| transformers/es2015.ts:1505:17 | PropertyAccessExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/es2015.ts:2173:49 | Identifier | `source` | `Node` | `SourceMapRange` |
| transformers/es2015.ts:2360:35 | PropertyAccessExpression | `source` | `BigIntLiteral` | `SourceMapRange` |
| transformers/es2015.ts:2369:41 | PropertyAccessExpression | `source` | `BigIntLiteral` | `SourceMapRange` |
| transformers/es2015.ts:2636:71 | Identifier | `source` | `TextRange` | `SourceMapRange` |
| transformers/es2015.ts:2805:52 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/es2015.ts:3046:52 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/es2015.ts:4836:39 | Identifier | `source` | `SuperExpression` | `SourceMapRange` |
| transformers/es2017.ts:630:13 | Identifier | `source` | `VariableDeclaration` | `SourceMapRange` |
| transformers/es2018.ts:790:51 | PropertyAccessExpression | `source` | `Expression` | `SourceMapRange` |
| transformers/es2018.ts:794:53 | PropertyAccessExpression | `source` | `Expression` | `SourceMapRange` |
| transformers/esDecorators.ts:610:56 | BinaryExpression | `source` | `Identifier` | `SourceMapRange` |
| transformers/esDecorators.ts:619:56 | BinaryExpression | `source` | `Identifier` | `SourceMapRange` |
| transformers/esDecorators.ts:893:52 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esDecorators.ts:923:62 | BinaryExpression | `source` | `Identifier` | `SourceMapRange` |
| transformers/esDecorators.ts:1079:56 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esDecorators.ts:1087:56 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esDecorators.ts:1231:40 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esDecorators.ts:1358:56 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esDecorators.ts:1390:56 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esDecorators.ts:1412:99 | Identifier | `id` | `never` | `PrivateIdentifier` |
| transformers/esDecorators.ts:1427:99 | Identifier | `id` | `never` | `PrivateIdentifier` |
| transformers/esDecorators.ts:1442:99 | Identifier | `id` | `never` | `PrivateIdentifier` |
| transformers/esDecorators.ts:1514:149 | ConditionalExpression | `id` | `never` | `PrivateIdentifier` |
| transformers/esDecorators.ts:1597:45 | PropertyAccessExpression | `source` | `Expression` | `SourceMapRange` |
| transformers/esDecorators.ts:1600:51 | PropertyAccessExpression | `source` | `Expression` | `SourceMapRange` |
| transformers/esDecorators.ts:1612:50 | PropertyAccessExpression | `source` | `BigIntLiteral` | `SourceMapRange` |
| transformers/esDecorators.ts:1702:20 | CallExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/esDecorators.ts:1720:40 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esDecorators.ts:1974:20 | CallExpression | `id` | `never` | `BigIntLiteral` |
| transformers/esDecorators.ts:1996:20 | CallExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/esDecorators.ts:2020:20 | CallExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/esDecorators.ts:2287:33 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esDecorators.ts:2295:35 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/esnext.ts:357:39 | CallExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/esnext.ts:540:43 | Identifier | `source` | `ClassDeclaration` | `SourceMapRange` |
| transformers/esnext.ts:597:42 | Identifier | `source` | `VariableStatement` | `SourceMapRange` |
| transformers/esnext.ts:618:39 | Identifier | `source` | `VariableDeclaration` | `SourceMapRange` |
| transformers/generators.ts:752:17 | Identifier | `source` | `VariableStatement` | `SourceMapRange` |
| transformers/generators.ts:1390:79 | PropertyAccessExpression | `source` | `ArrayBindingPattern` | `SourceMapRange` |
| transformers/generators.ts:1393:13 | Identifier | `source` | `InitializedVariableDeclaration` | `SourceMapRange` |
| transformers/generators.ts:2077:50 | Identifier | `source` | `Identifier` | `SourceMapRange` |
| transformers/jsx.ts:111:20 | PropertyAccessExpression | `id` | `never` | `Identifier` |
| transformers/jsx.ts:115:16 | PropertyAccessExpression | `id` | `never` | `Identifier` |
| transformers/jsx.ts:171:177 | PropertyAccessExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/jsx.ts:359:63 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| transformers/legacyDecorators.ts:435:40 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/legacyDecorators.ts:517:40 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/legacyDecorators.ts:660:35 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/legacyDecorators.ts:698:39 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/legacyDecorators.ts:836:50 | Identifier | `source` | `Identifier` | `SourceMapRange` |
| transformers/module/esnextAnd2015.ts:277:55 | Identifier | `id` | `never` | `ExternalModuleReference` |
| transformers/module/esnextAnd2015.ts:289:73 | Identifier | `id` | `never` | `ExternalModuleReference` |
| transformers/module/module.ts:1566:55 | Identifier | `id` | `never` | `ExternalModuleReference` |
| transformers/module/module.ts:1588:63 | Identifier | `id` | `never` | `ExternalModuleReference` |
| transformers/module/module.ts:1618:73 | Identifier | `id` | `never` | `ExternalModuleReference` |
| transformers/module/system.ts:775:73 | Identifier | `id` | `never` | `ExternalModuleReference` |
| transformers/module/system.ts:776:80 | Identifier | `id` | `never` | `ExternalModuleReference` |
| transformers/namedEvaluation.ts:205:68 | PropertyAccessExpression | `source` | `Identifier` | `SourceMapRange` |
| transformers/namedEvaluation.ts:265:75 | PropertyAccessExpression | `id` | `never` | `BigIntLiteral` |
| transformers/namedEvaluation.ts:268:9 | Identifier | `id` | `never` | `BigIntLiteral` |
| transformers/namedEvaluation.ts:288:9 | Identifier | `equalsToken` | `never` | `ShorthandPropertyAssignment` |
| transformers/namedEvaluation.ts:311:46 | PropertyAccessExpression | `id` | `never` | `Identifier` |
| transformers/namedEvaluation.ts:314:9 | Identifier | `id` | `never` | `ArrayBindingPattern` |
| transformers/namedEvaluation.ts:315:9 | PropertyAccessExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/namedEvaluation.ts:341:46 | PropertyAccessExpression | `id` | `never` | `Identifier` |
| transformers/namedEvaluation.ts:344:9 | Identifier | `id` | `never` | `ArrayBindingPattern` |
| transformers/namedEvaluation.ts:347:9 | PropertyAccessExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/namedEvaluation.ts:373:46 | PropertyAccessExpression | `id` | `never` | `Identifier` |
| transformers/namedEvaluation.ts:376:9 | Identifier | `id` | `never` | `ArrayBindingPattern` |
| transformers/namedEvaluation.ts:379:9 | PropertyAccessExpression | `id` | `never` | `ArrayBindingPattern` |
| transformers/namedEvaluation.ts:396:9 | Identifier | `modifiers` | `never` | `PropertyDeclaration` |
| transformers/namedEvaluation.ts:437:9 | Identifier | `id` | `never` | `BinaryExpression` |
| transformers/namedEvaluation.ts:459:9 | Identifier | `modifiers` | `never` | `ExportAssignment` |
| transformers/ts.ts:995:45 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/ts.ts:1059:35 | PropertyAccessExpression | `id` | `never` | `Node` |
| transformers/ts.ts:1064:21 | PropertyAccessExpression | `id` | `never` | `BigIntLiteral` |
| transformers/ts.ts:1069:52 | Identifier | `id` | `never` | `Node` |
| transformers/ts.ts:1432:27 | Identifier | `id` | `never` | `Node` |
| transformers/ts.ts:1453:41 | Identifier | `id` | `never` | `Identifier` |
| transformers/ts.ts:1457:33 | Identifier | `id` | `never` | `Expression` |
| transformers/ts.ts:1460:25 | Identifier | `id` | `never` | `Node` |
| transformers/ts.ts:1626:40 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/ts.ts:2039:62 | Identifier | `source` | `EnumDeclaration` | `SourceMapRange` |
| transformers/ts.ts:2042:46 | Identifier | `source` | `ModuleDeclaration` | `SourceMapRange` |
| transformers/ts.ts:2529:39 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/ts.ts:2532:38 | CallExpression | `source` | `TextRange` | `SourceMapRange` |
| transformers/ts.ts:2565:33 | PropertyAccessExpression | `source` | `Identifier` | `SourceMapRange` |
| tsbuildPublic.ts:313:18 | AsExpression | `reportErrorSummary` | `SolutionBuilderHostBase<T>` | `SolutionBuilderHost<T>` |
| tsbuildPublic.ts:430:18 | AsExpression | `reportErrorSummary` | `SolutionBuilderWithWatchHost<T>` | `SolutionBuilderHost<T>` |
| tsbuildPublic.ts:454:17 | Identifier | `getGlobalTypingsCacheLocation` | `SolutionBuilderHost<T>` | `ModuleResolutionHost` |
| tsbuildPublic.ts:475:17 | Identifier | `getGlobalTypingsCacheLocation` | `SolutionBuilderHost<T>` | `ModuleResolutionHost` |
| tsbuildPublic.ts:488:17 | Identifier | `getGlobalTypingsCacheLocation` | `SolutionBuilderHost<T>` | `ModuleResolutionHost` |
| tsbuildPublic.ts:499:66 | Identifier | `onUnRecoverableConfigFileDiagnostic` | `SolutionBuilderHost<T>` | `ProgramHost<T> & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter \| undefined; }` |
| utilities.ts:1024:47 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:1212:46 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:1247:42 | CallExpression | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:1267:24 | CallExpression | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:1282:38 | CallExpression | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:2010:30 | AsExpression | `rawText` | `LiteralLikeNode` | `TemplateLiteralLikeNode` |
| utilities.ts:2528:67 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:2529:65 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:2533:70 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:3481:80 | PropertyAccessExpression | `id` | `never` | `Node` |
| utilities.ts:3780:14 | AsExpression | `id` | `never` | `ExternalModuleReference` |
| utilities.ts:4369:20 | ConditionalExpression | `templateFlags` | `never` | `NoSubstitutionTemplateLiteral` |
| utilities.ts:4369:52 | PropertyAccessExpression | `id` | `never` | `LiteralExpression & StringLiteral` |
| utilities.ts:4416:20 | ConditionalExpression | `id` | `never` | `Expression` |
| utilities.ts:4416:52 | PropertyAccessExpression | `id` | `never` | `LiteralExpression & StringLiteral` |
| utilities.ts:6750:38 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:6839:25 | Identifier | `modifiers` | `never` | `GetAccessorDeclaration` |
| utilities.ts:6841:27 | Identifier | `modifiers` | `never` | `GetAccessorDeclaration` |
| utilities.ts:6844:27 | Identifier | `modifiers` | `never` | `SetAccessorDeclaration` |
| utilities.ts:7946:37 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:7951:37 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:7961:37 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:7973:37 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:7979:37 | Identifier | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| utilities.ts:8807:39 | Identifier | `reportsUnnecessary` | `DiagnosticRelatedInformation` | `Diagnostic` |
| utilities.ts:8807:44 | Identifier | `reportsUnnecessary` | `DiagnosticRelatedInformation` | `Diagnostic` |
| utilities.ts:9231:41 | Identifier | `all` | `Pick<CompilerOptions, "noImplicitAny" \| "strict">` | `CompilerOptions` |
| utilities.ts:9237:41 | Identifier | `all` | `Pick<CompilerOptions, "noImplicitThis" \| "strict">` | `CompilerOptions` |
| utilities.ts:9243:41 | Identifier | `all` | `Pick<CompilerOptions, "strict" \| "strictNullChecks">` | `CompilerOptions` |
| utilities.ts:9249:41 | Identifier | `all` | `Pick<CompilerOptions, "strict" \| "strictFunctionTypes">` | `CompilerOptions` |
| utilities.ts:9255:41 | Identifier | `all` | `Pick<CompilerOptions, "strict" \| "strictBindCallApply">` | `CompilerOptions` |
| utilities.ts:9261:41 | Identifier | `all` | `Pick<CompilerOptions, "strict" \| "strictPropertyInitialization">` | `CompilerOptions` |
| utilities.ts:9267:41 | Identifier | `all` | `Pick<CompilerOptions, "strict" \| "strictBuiltinIteratorReturn">` | `CompilerOptions` |
| utilities.ts:9280:41 | Identifier | `all` | `Pick<CompilerOptions, "strict" \| "useUnknownInCatchVariables">` | `CompilerOptions` |
| utilities.ts:10955:56 | AsExpression | `constraint` | `Type` | `TypeParameter` |
| utilities.ts:12189:121 | PropertyAccessExpression | `templateFlags` | `never` | `NoSubstitutionTemplateLiteral` |
| utilitiesPublic.ts:932:55 | PropertyAccessExpression | `id` | `never` | `Node` |
| utilitiesPublic.ts:932:81 | AsExpression | `id` | `never` | `Identifier` |
| watch.ts:250:60 | NonNullExpression | `getPositionOfLineAndCharacter` | `SourceFile` | `SourceFileLike` |
| watch.ts:771:47 | CallExpression | `omitTrailingSemicolon` | `CompilerOptions` | `PrinterOptions` |
| watch.ts:877:45 | Identifier | `omitTrailingSemicolon` | `CompilerOptions` | `PrinterOptions` |
| watch.ts:990:53 | Identifier | `createProgram` | `IncrementalCompilationOptions` | `IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram>` |
| watchPublic.ts:475:39 | Identifier | `omitTrailingSemicolon` | `CompilerOptions` | `PrinterOptions` |
| watchPublic.ts:732:36 | BinaryExpression | `omitTrailingSemicolon` | `CompilerOptions` | `PrinterOptions` |
