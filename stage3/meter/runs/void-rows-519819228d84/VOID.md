# Stage 3 tsc void rows

Main: `519819228d84d2a8eefce34a6bf76b79c7eeff63`.

Measured a detached main worktree with only `55c1c38fbb0840fbd4c2b93579c8722d4c1df35b` cherry-picked (local measurement HEAD `2f2c6c146`). The report branch is cut directly from main and contains no compiler change. Upstream TypeScript: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.

The build copied `build_census` from `stage3/meter/twice-daily.sh`, including both overlays and all three binaries. It ran the same lockfile-pinned npm ci flags and apply command. The foreground latent command used `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1` over adapted `src/compiler`. Setup, all builds, apply, and latent exited 0. No code was edited.

Latent elapsed: **651 seconds (10m 51s)**. Coverage: **79/79 source files**. Header: checker-rejected program, first-error mode. These are measurement findings, not proof of native output.

## Headline totals

The meter counts unique `(kind, where, reason, text)` findings across attempts. Raw rows are shown separately; they must not be confused with distinct source sites.

| Kind | Meter total | Raw rows |
| --- | ---: | ---: |
| NotYet | 811 | 1,198 |
| Refused | 12,015 | 12,203 |
| panic | 0 | 0 |
| error | 0 | 0 |
| SkippedDependency | 1 | 1 |

NotYet + Refused: **12,826** meter findings; **13,401** raw rows. The latter is 13,401 in this exact run; the motivating 13,402 is not substituted for measured data.

## Selected families

| Reason predicate | Raw findings | Distinct where sites | Distinct units per site |
| --- | ---: | ---: | --- |
| void-container read | 9,610 | 9,504 | 1 unit(s): 9504 sites |
| instantiated at void | 107 | 76 | 1 unit(s): 69 sites, 2 unit(s): 1 sites, 3 unit(s): 5 sites, 21 unit(s): 1 sites |

Combined: **9,717 rows, 9,574 sites**. Six sites occur in both families. Site identity is the finding’s `where`, normalized to `src/…:line:column`; it is never `site_where` or the read location embedded in a generic reason. Units are distinct normalized `unit` values, not row counts. All 9,504 direct void-container sites occur under exactly one unit; the 106 excess rows in that family repeat sites within the same unit. This rise is predominantly many sites, not many units multiplying a few sites.

Every extracted row retains its original finding fields (except normalized `where` and `unit`), exact reason, source line, syntax kind, expression, and unit counts. The exact reason retains its original temporary path when one is embedded in the message. `site_unit_count` covers both families; `exact_reason_site_unit_count` covers that exact reason. Thus the gzip records the count for every site, including all one-unit sites.

### Sites under multiple units

| Site | Units | Source line |
| --- | ---: | --- |
| `src/compiler/factory/utilities.ts:679:12` | 2 | `return setStartsOnNewLine(node, /*newLine*/ true);` |
| `src/compiler/moduleNameResolver.ts:1348:20` | 3 | `const result = createModuleOrTypeReferenceResolutionCache(` |
| `src/compiler/moduleNameResolver.ts:1381:12` | 3 | `return createModuleOrTypeReferenceResolutionCache(` |
| `src/compiler/program.ts:638:19` | 3 | `diagnostics = addRange(diagnostics, program.getConfigFileParsingDiagnostics());` |
| `src/compiler/transformers/utilities.ts:603:12` | 3 | `return filter(node.members, isStaticPropertyDeclarationOrClassStaticBlockDeclaration);` |
| `src/compiler/utilities.ts:3556:20` | 3 | `return some((node as ClassDeclaration).members, m => nodeOrChildIsDecorated(useLegacyDecorators, m, node, parent!));` |
| `src/compiler/utilitiesPublic.ts:1280:12` | 21 | `return find(getJSDocTagsWorker(node, noCache), predicate);` |

In particular, `utilitiesPublic.ts:1280:12` is `find` instantiated with `T = JSDocTag`: 21 rows, one call site, 21 units. The source is `return find(getJSDocTagsWorker(node, noCache), predicate);`; the reason points to the generic body’s read at `core.ts:170:23`. It is a generic call, not a getter on JSDocTag.

## Top three shapes

Grouping is **exact reason text first, then the TypeScript expression syntax at the reported site**, ranked by distinct normalized `where`. Argument expressions are resolved by AST parent membership in call/new arguments; this distinguishes a named value, a property read, and a nested call result even when they start at the same column. These are syntactic shapes, not a claim about checker types or a proved root cause.

The three largest exact-reason groups before splitting by construct are argument (5,620 sites), value (897), and returned value (880). After splitting by construct, all top three shapes belong to the argument reason. Ordinary node-shaped arguments dominate; these are not limited to explicit `void[]` or named callbacks.

### 1. Named value passed as an argument

**3,101 sites; 3,101 rows; one unit per site.** AST: `Identifier`. Exact reason: `a read of a void container's elements, as an argument`.

| Site | Expression | Source line |
| --- | --- | --- |
| `src/compiler/binder.ts:333:30` | `node` | `setParent(node.body, node);` |
| `src/compiler/checker.ts:54204:24` | `declaration` | `return !isAccessor(declaration);` |
| `src/compiler/commandLineParser.ts:2008:23` | `text` | `if (!isString(text)) {` |
| `src/compiler/core.ts:780:38` | `array` | `const insertIndex = binarySearch(array, insert, identity, compare);` |
| `src/compiler/debug.ts:561:39` | `flowNode` | `Object.setPrototypeOf(flowNode, flowNodeProto);` |

### 2. Property access passed as an argument

**1,474 sites; 1,474 rows; one unit per site.** AST: `PropertyAccessExpression`. Exact reason: `a read of a void container's elements, as an argument`.

| Site | Expression | Source line |
| --- | --- | --- |
| `src/compiler/binder.ts:1900:18` | `node.left` | `bind(node.left);` |
| `src/compiler/commandLineParser.ts:2459:38` | `rootExpression.elements` | `const firstObject = find(rootExpression.elements, isObjectLiteralExpression);` |
| `src/compiler/debug.ts:1115:50` | `clause.expression` | `clauses.push(getNodeText(clause.expression));` |
| `src/compiler/emitter.ts:2224:28` | `node.name` | `emitNodeWithWriter(node.name, writeParameter);` |
| `src/compiler/factory/nodeConverters.ts:101:34` | `element.name` | `Debug.assertNode(element.name, isIdentifier);` |

### 3. Call result passed as an argument

**854 sites; 854 rows; one unit per site.** AST: `CallExpression`. Exact reason: `a read of a void container's elements, as an argument`.

| Site | Expression | Source line |
| --- | --- | --- |
| `src/compiler/executeCommandLine.ts:556:28` | `getOptionsForHelp(commandLine)` | `printEasyHelp(sys, getOptionsForHelp(commandLine));` |
| `src/compiler/factory/utilities.ts:255:32` | `factory.createNull()` | `argumentsList.push(factory.createNull());` |
| `src/compiler/factory/nodeFactory.ts:1557:55` | `createModifier(SyntaxKind.ExportKeyword)` | `if (flags & ModifierFlags.Export) result.push(createModifier(SyntaxKind.ExportKeyword));` |
| `src/compiler/factory/emitHelpers.ts:212:29` | `factory.createArrayLiteralExpression(decoratorExpressions, /*multiLine*/ true)` | `argumentsArray.push(factory.createArrayLiteralExpression(decoratorExpressions, /*multiLine*/ true));` |
| `src/compiler/binder.ts:1732:44` | `createFlowSwitchClause(preSwitchCaseFlow, node, 0, 0)` | `addAntecedent(postSwitchLabel, createFlowSwitchClause(preSwitchCaseFlow, node, 0, 0));` |

## Every exact reason

Raw findings count every matching row. Sites deduplicate by `where` within this exact reason. The unit histogram gives the number of distinct units for each site. Reasons below are exact emitted strings; paths embedded in them are deliberately retained.

| Exact reason | Findings | Sites | Units per site |
| --- | ---: | ---: | --- |
| a read of a void container's elements, as an argument | 5,620 | 5,620 | 1 unit(s): 5620 sites |
| a read of a void container's elements, as a value | 897 | 897 | 1 unit(s): 897 sites |
| a read of a void container's elements, as a returned value | 880 | 880 | 1 unit(s): 880 sites |
| a read of a void container's elements, as an assertion's operand | 562 | 562 | 1 unit(s): 562 sites |
| a read of a void container's elements, as an operand of && | 343 | 343 | 1 unit(s): 343 sites |
| a read of a void container's elements, as an assigned value | 310 | 310 | 1 unit(s): 310 sites |
| a read of a void container's elements, as an operand of \|\| | 149 | 149 | 1 unit(s): 149 sites |
| a read of a void container's elements, as a condition | 138 | 138 | 1 unit(s): 138 sites |
| a read of a void container's elements, as a spread | 83 | 83 | 1 unit(s): 83 sites |
| a read of a void container's elements, as a destructured value | 65 | 65 | 1 unit(s): 65 sites |
| a read of a void container's elements, as an operand of ?? | 62 | 62 | 1 unit(s): 62 sites |
| a read of a void container's elements, through get | 50 | 50 | 1 unit(s): 50 sites |
| a read of a void container's elements, as a binding's initializer | 42 | 42 | 1 unit(s): 42 sites |
| a read of a void container's elements, through set | 34 | 34 | 1 unit(s): 34 sites |
| a read of a void container's elements, as typeof's operand | 28 | 28 | 1 unit(s): 28 sites |
| a read of a void container's elements, through map | 23 | 23 | 1 unit(s): 23 sites |
| a read of a void container's elements, through getCompilerOptions | 18 | 18 | 1 unit(s): 18 sites |
| a read of a void container's elements, through getSourceFiles | 16 | 16 | 1 unit(s): 16 sites |
| a read of a void container's elements, through getSourceFileByPath | 15 | 15 | 1 unit(s): 15 sites |
| a read of a void container's elements, through has | 14 | 14 | 1 unit(s): 14 sites |
| a read of a void container's elements, through delete | 13 | 13 | 1 unit(s): 13 sites |
| a read of a void container's elements, through onDiagnosticReported: this holds void in field context, and onDiagnosticReported reads it | 10 | 10 | 1 unit(s): 10 sites |
| a read of a void container's elements, as an operand of ??= | 8 | 8 | 1 unit(s): 8 sites |
| a read of a void container's elements, through getCurrentDirectory | 8 | 8 | 1 unit(s): 8 sites |
| a read of a void container's elements, through add | 7 | 7 | 1 unit(s): 7 sites |
| a read of a void container's elements, through emit | 7 | 7 | 1 unit(s): 7 sites |
| a read of a void container's elements, through filter | 7 | 7 | 1 unit(s): 7 sites |
| a read of a void container's elements, through indexOf | 7 | 7 | 1 unit(s): 7 sites |
| a read of a void container's elements, through some | 7 | 7 | 1 unit(s): 7 sites |
| a read of a void container's elements, as a for...of source bound to an existing name or a pattern | 6 | 6 | 1 unit(s): 6 sites |
| a read of a void container's elements, through useCaseSensitiveFileNames | 6 | 6 | 1 unit(s): 6 sites |
| a generic instantiated at void whose body reads the value: some's T = string \| number \| boolean \| readonly string[] \| SourceFile \| undefined, read at /tmp/void-meter/adapted/src/compiler/core.ts:622:31 as an argument | 5 | 5 | 1 unit(s): 5 sites |
| a read of a void container's elements, through afterProgramEmitAndDiagnostics | 5 | 5 | 1 unit(s): 5 sites |
| a read of a void container's elements, through forEach with a callback not written in place | 5 | 5 | 1 unit(s): 5 sites |
| a read of a void container's elements, through getTypeChecker | 5 | 5 | 1 unit(s): 5 sites |
| a generic instantiated at void whose body reads the value: visitEachChild's T = SourceFile, read at /tmp/void-meter/adapted/src/compiler/visitorPublic.ts:604:31 as a value seen without its void | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through clear | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through emitBuildInfo | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through enqueue | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through getDeclarationDiagnostics | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through getPackageJsonInfoCache | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through getProgram | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through getRootFileNames | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through isSourceFileDefaultLibrary | 4 | 4 | 1 unit(s): 4 sites |
| a read of a void container's elements, through onWatchStatusChange | 4 | 4 | 1 unit(s): 4 sites |
| a generic instantiated at void whose body reads the value: forEachPropertyAssignment's T = StringLiteral, read at /tmp/void-meter/adapted/src/compiler/utilities.ts:3096:20 as an argument | 3 | 3 | 1 unit(s): 3 sites |
| a generic instantiated at void whose body reads the value: getOriginalNode's T = SourceFile, read at /tmp/void-meter/adapted/src/compiler/utilitiesPublic.ts:767:9 as a value | 3 | 3 | 1 unit(s): 3 sites |
| a generic instantiated at void whose body reads the value: setOriginalNode's T = AssignmentExpression<EqualsToken>, read at /tmp/void-meter/adapted/src/compiler/factory/nodeFactory.ts:7423:13 as a value | 3 | 3 | 1 unit(s): 3 sites |
| a generic instantiated at void whose body reads the value: some's T = ClassElement, read at /tmp/void-meter/adapted/src/compiler/core.ts:622:31 as an argument | 5 | 3 | 1 unit(s): 2 sites, 3 unit(s): 1 sites |
| a read of a void container's elements, through afterProgramCreate | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through done | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through find | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through getConfigFileParsingDiagnostics | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through getGlobalDiagnostics | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through getOptionsDiagnostics | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through getParsedCommandLine | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through getResolvedProjectReferences | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through getSemanticDiagnostics | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through getSourceFile | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through getSyntacticDiagnostics | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through hasChangedEmitSignature | 3 | 3 | 1 unit(s): 3 sites |
| a read of a void container's elements, through setModifiedTime | 3 | 3 | 1 unit(s): 3 sites |
| a generic instantiated at void whose body reads the value: createWatchCompilerHost's T = EmitAndSemanticDiagnosticsBuilderProgram, read at /tmp/void-meter/adapted/src/compiler/watch.ts:875:5 through afterProgramCreate | 2 | 2 | 1 unit(s): 2 sites |
| a generic instantiated at void whose body reads the value: filter's T = AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 9 more ... \| StaticKeyword, read at /tmp/void-meter/adapted/src/compiler/core.ts:281:29 as an argument | 2 | 2 | 1 unit(s): 2 sites |
| a generic instantiated at void whose body reads the value: filter's T = Diagnostic, read at /tmp/void-meter/adapted/src/compiler/core.ts:281:29 as an argument | 2 | 2 | 1 unit(s): 2 sites |
| a generic instantiated at void whose body reads the value: forEachChild's T = Node, read at /tmp/void-meter/adapted/src/compiler/parser.ts:1254:46 as an argument | 2 | 2 | 1 unit(s): 2 sites |
| a generic instantiated at void whose body reads the value: some's T = Decorator, read at /tmp/void-meter/adapted/src/compiler/core.ts:622:31 as an argument | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, as a loop condition | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, as an operand of in | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through __debugTypeToString | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through build | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through clearTimeout | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through dispose | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through emitNodeWithNotification | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through forEach's callback parameter 0, destructured or rest | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through get: IdentifierNameMap<ExportSpecifier[]> holds void in its type argument 1 (ExportSpecifier[]), and get reads it at /tmp/void-meter/adapted/src/compiler/transformers/utilities.ts:401:16 through get | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through get: IdentifierNameMap<ExportSpecifier[]> \| undefined holds void in its type argument 1 (ExportSpecifier[]), and get reads it | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getAllDependencies | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getCompilerOptionsForFile | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getFileIncludeReasons | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getFromDirectoryCache | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getMissingFilePaths | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getOrCreateCacheForNonRelativeName | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getProgramDiagnostics | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getProgramOrUndefined | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getProjectReferences | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through getSemanticDiagnosticsOfNextAffectedFile | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through isEmitNotificationEnabled | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through realpath | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through releaseProgram | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through setTimeout | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through substituteNode | 2 | 2 | 1 unit(s): 2 sites |
| a read of a void container's elements, through update | 2 | 2 | 1 unit(s): 2 sites |
| a generic instantiated at void whose body reads the value: addEmitFlags's T = Node, read at /tmp/void-meter/adapted/src/compiler/factory/emitNode.ts:103:42 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: addRange's T = Diagnostic, read at /tmp/void-meter/adapted/src/compiler/core.ts:992:21 as an argument | 3 | 1 | 3 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: binarySearchKey's T = Node, read at /tmp/void-meter/adapted/src/compiler/core.ts:1211:36 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: cast's TIn = Expression, TOut = LeftHandSideExpression, read at /tmp/void-meter/adapted/src/compiler/core.ts:1784:37 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: cast's TIn = NamedExports \| NamespaceExport, TOut = NamedExports, read at /tmp/void-meter/adapted/src/compiler/core.ts:1784:37 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: checkDefined's T = Program, read at /tmp/void-meter/adapted/src/compiler/debug.ts:257:16 as a returned value | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: checkDefined's T = SourceFile, read at /tmp/void-meter/adapted/src/compiler/debug.ts:257:16 as a returned value | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: checkDefined's T = TypeNode, read at /tmp/void-meter/adapted/src/compiler/debug.ts:257:16 as a returned value | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: checkDefined's T = string \| number \| boolean \| readonly string[] \| SourceFile, read at /tmp/void-meter/adapted/src/compiler/debug.ts:257:16 as a returned value | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: contains's T = ResolvedProjectReference, read at /tmp/void-meter/adapted/src/compiler/core.ts:223:34 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: countWhere's T = Diagnostic, read at /tmp/void-meter/adapted/src/compiler/core.ts:247:27 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: createModuleOrTypeReferenceResolutionCache's T = ResolvedModuleWithFailedLookupLocations, read at /tmp/void-meter/adapted/src/compiler/moduleNameResolver.ts:1302:12 as a spread | 3 | 1 | 3 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: createModuleOrTypeReferenceResolutionCache's T = ResolvedTypeReferenceDirectiveWithFailedLookupLocations, read at /tmp/void-meter/adapted/src/compiler/moduleNameResolver.ts:1302:12 as a spread | 3 | 1 | 3 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: createQueue's T = Expression, read at /tmp/void-meter/adapted/src/compiler/core.ts:1570:41 as an operand of ?? | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: emitFilesAndReportErrorsAndGetExitStatus's T = BuilderProgram \| Program, read at /tmp/void-meter/adapted/src/compiler/watch.ts:644:41 as a destructured value | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: filter's T = ClassElement, read at /tmp/void-meter/adapted/src/compiler/core.ts:281:29 as an argument | 3 | 1 | 3 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: filter's T = JsxChild, read at /tmp/void-meter/adapted/src/compiler/core.ts:281:29 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: filter's T = SourceFile, read at /tmp/void-meter/adapted/src/compiler/core.ts:281:29 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: filter's T = VariableDeclaration, read at /tmp/void-meter/adapted/src/compiler/core.ts:281:29 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: filterMutate's T = Diagnostic, read at /tmp/void-meter/adapted/src/compiler/core.ts:302:15 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: find's T = Declaration, read at /tmp/void-meter/adapted/src/compiler/core.ts:170:23 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: find's T = JSDocTag, read at /tmp/void-meter/adapted/src/compiler/core.ts:170:23 as an argument | 21 | 1 | 21 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: flatMap's T = JSDocTag, U = TypeParameterDeclaration, read at /tmp/void-meter/adapted/src/compiler/core.ts:403:29 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: forEach's T = BindingOrAssignmentElement, read at /tmp/void-meter/adapted/src/compiler/core.ts:36:37 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: forEach's T = DiagnosticRelatedInformation, read at /tmp/void-meter/adapted/src/compiler/core.ts:36:37 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: forEach's T = GetAccessorDeclaration \| MethodDeclaration \| PropertyAssignment \| SetAccessorDeclaration \| ShorthandPropertyAssignment \| SpreadAssignment, U = false \| Expression, read at /tmp/void-meter/adapted/src/compiler/core.ts:36:37 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: forEach's T = SourceFile, U = void (read from the signature), read at /tmp/void-meter/adapted/src/compiler/core.ts:36:37 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: forEach's T = Statement, U = ArrowFunction \| ClassDeclaration \| ClassExpression \| ConstructorDeclaration \| ConstructorTypeNode \| ... 19 more ... \| VariableStatement, read at /tmp/void-meter/adapted/src/compiler/core.ts:36:37 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: forEachChild's T = void (read from the signature), read at /tmp/void-meter/adapted/src/compiler/parser.ts:1254:46 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: forEachEntry's V = Symbol, read at /tmp/void-meter/adapted/src/compiler/utilities.ts:747:22 through entries | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: getAllJSDocTags's T = JSDocImplementsTag, read at /tmp/void-meter/adapted/src/compiler/utilitiesPublic.ts:1285:12 through filter | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: getAllJSDocTags's T = JSDocOverloadTag, read at /tmp/void-meter/adapted/src/compiler/utilitiesPublic.ts:1285:12 through filter | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: getOriginalNode's T = Node, read at /tmp/void-meter/adapted/src/compiler/utilitiesPublic.ts:767:9 as a value | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: map's T = SourceFile, U = SourceFile, read at /tmp/void-meter/adapted/src/compiler/core.ts:325:27 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: sameMap's T = Diagnostic, U = Diagnostic, read at /tmp/void-meter/adapted/src/compiler/core.ts:354:30 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: setEmitFlags's T = Block, read at /tmp/void-meter/adapted/src/compiler/factory/emitNode.ts:93:25 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: setEmitFlags's T = Identifier, read at /tmp/void-meter/adapted/src/compiler/factory/emitNode.ts:93:25 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: setEmitFlags's T = PropertyAccessExpression, read at /tmp/void-meter/adapted/src/compiler/factory/emitNode.ts:93:25 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: setParent's T = Identifier, read at /tmp/void-meter/adapted/src/compiler/utilities.ts:10707:9 as an operand of && | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: setParent's T = ModuleBody, read at /tmp/void-meter/adapted/src/compiler/utilities.ts:10707:9 as an operand of && | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: setStartsOnNewLine's T = Expression, read at /tmp/void-meter/adapted/src/compiler/factory/emitNode.ts:176:25 as an argument | 2 | 1 | 2 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: setTextRangeEnd's T = Expression, read at /tmp/void-meter/adapted/src/compiler/utilities.ts:10658:6 as an assertion's operand | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: some's T = AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 9 more ... \| StaticKeyword, read at /tmp/void-meter/adapted/src/compiler/core.ts:622:31 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: some's T = Statement, read at /tmp/void-meter/adapted/src/compiler/core.ts:622:31 as an argument | 1 | 1 | 1 unit(s): 1 sites |
| a generic instantiated at void whose body reads the value: tryCast's TIn = NamedImports \| NamespaceImport, TOut = NamespaceImport, read at /tmp/void-meter/adapted/src/compiler/core.ts:1779:12 as a returned value | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, as a value seen without its void | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, as an operand of + | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, as an operand of \|\|= | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through a computed member | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through add: IdentifierNameMultiMap<ExportSpecifier> holds void in its type argument 1 (ExportSpecifier), and add reads it at /tmp/void-meter/adapted/src/compiler/transformers/utilities.ts:443:22 through get: this holds void in field _map, and get reads it | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through clean | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through deleteFile | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through dequeue | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through emitNextAffectedFile | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through every | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through fileExists | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through forEachResolvedTypeReferenceDirective | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getBindAndCheckDiagnostics | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getBuildInfo | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getBuildOrder | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getIdentifierCount | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getInstantiationCount | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getModifiedTime | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getNewLine | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getOrCreateCacheForDirectory | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getOrCreateCacheForModuleName | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getRedirectFromSourceFile | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getRelationCacheSizes | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getResolvedModuleFromModuleSpecifier | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getResolvedTypeReferenceDirectiveFromTypeReferenceDirective | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getSymbolCount | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through getTypeCount | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through isEmittedFile | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through isEmpty | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through keys | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through lastIndexOf | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through select | 1 | 1 | 1 unit(s): 1 sites |
| a read of a void container's elements, through values | 1 | 1 | 1 unit(s): 1 sites |

## Artifacts and reproduction

- `run.sh`: exact measurement script used; requires the requested setup and fresh throwaway worktree paths. Reuses the setup-verified identical cohere gitlink via a local symlink.
- `shapes.cjs`: TypeScript 6.0.3 parser classification of selected sites.
- `report.py`: extraction, coverage checks, unit accounting, and this report.
- `void-rows.jsonl.gz`: only matching finding rows, with source and per-site metadata; no full latent census is committed.

Local logs: `/tmp/setup.log`, `/tmp/void-fetch.log`, `/tmp/void-worktree.log`, `/tmp/void-run.log`, and `/tmp/void-meter/logs/{build,latent-overlay,entry-overlay,entry-build,latent-build,upstream,packages,apply,latent}.log`. Full latent remains local at `/tmp/void-meter/latent.jsonl`.

Full latent SHA-256 (local, uncommitted): `ffa887092e4ee3279af33437610ab337ac1b825a55f8e58093d9c2afdaf57b81`.
