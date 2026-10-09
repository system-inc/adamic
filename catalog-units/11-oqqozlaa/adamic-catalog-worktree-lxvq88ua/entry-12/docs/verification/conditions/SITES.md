# Every recorded condition site

At 7e18441dfd2527304b47d78f396700ea256e2625, the full compiler-directory census has zero non-boolean condition refusals. Each of these 78 original operands and its if/while/ternary form is preserved in conditions_ledger78.a, with concrete surrounding data shapes. The 241 labeled observations agree with source on Node, release native, sanitized native and the JavaScript backend, and leak checks pass. This table covers condition-only reductions, not entire upstream function bodies. Five exact original positions retain independent full-source blockers, shown in the final column; other upstream units may have earlier blockers.

Ledger: c6ea4131267ddd110e0b4c0df9367d38e2c86560, stage3/adapt/41-explicit-any-remaining/evidence/before-census.json. Reconstructed inputs use that commit's final adaptation pipeline. ledger-types.json includes every source SHA-256 and the stock TypeScript 6.0.3 type. ledger-coverage.json additionally records each reduction's concrete type and sample inputs.

| Original site | Original operand | Checker type | Reduction | Samples | Exact full-source blocker |
| --- | --- | --- | --- | ---: | --- |
| src/compiler/builder.ts:642:13 | affectedFiles | readonly SourceFile[] \| undefined | conditionSite00 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/commandLineParser.ts:4176:9 | match | RegExpExecArray \| null | conditionSite01 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/core.ts:1894:13 | callback | (() => T) \| undefined | conditionSite02 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/factory/emitNode.ts:273:9 | helpers | EmitHelper[] \| undefined | conditionSite03 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleNameResolver.ts:480:9 | options.typeRoots | string[] \| undefined | conditionSite04 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/semver.ts:216:13 | sets | Comparator[][] \| undefined | conditionSite05 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleSpecifiers.ts:699:9 | host.getNearestAncestorDirectoryWithPackageJson | ((fileName: string, rootDir?: string \| undefined) => string \| undefined) \| undefined | conditionSite06 | 2 | Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) |
| src/compiler/program.ts:768:30 | host | FormatDiagnosticsHost | conditionSite07 | 1 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/scanner.ts:469:12 | sourceFile.getPositionOfLineAndCharacter | ((line: number, character: number, allowEdits?: true \| undefined) => number) \| undefined | conditionSite08 | 2 | Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) |
| src/compiler/scanner.ts:763:9 | error | ((diag: DiagnosticMessage, pos?: number \| undefined, len?: number \| undefined) => void) \| undefined | conditionSite09 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/scanner.ts:968:9 | match | RegExpExecArray \| null | conditionSite10 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/tsbuildPublic.ts:206:12 | host.now | (() => Date) \| undefined | conditionSite11 | 2 | Refused: a method read as a value (now would lose its object, and this with it) |
| src/compiler/utilities.ts:2339:12 | container | Node | conditionSite12 | 1 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:2461:22 | messageChain.next | DiagnosticMessageChain[] \| undefined | conditionSite13 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:2475:22 | messageChain.next | DiagnosticMessageChain[] \| undefined | conditionSite14 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:2997:9 | node | Node | conditionSite15 | 1 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:3057:13 | beforeUnwrapLabelCallback | ((node: LabeledStatement) => void) \| undefined | conditionSite16 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:3295:9 | container | ThisContainer | conditionSite17 | 1 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:4751:9 | node.symbol | Symbol | conditionSite18 | 1 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:5064:12 | node | Node | conditionSite19 | 1 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:6484:12 | host.useCaseSensitiveFileNames | (() => boolean) \| undefined | conditionSite20 | 2 | Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) |
| src/compiler/utilities.ts:7355:9 | modifiers | readonly ModifierLike[] \| undefined | conditionSite21 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:8745:22 | chain.next | DiagnosticMessageChain[] \| undefined | conditionSite22 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:8769:12 | lastChain.next | DiagnosticMessageChain[] \| undefined | conditionSite23 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:9896:9 | includes | readonly string[] \| undefined | conditionSite24 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:10159:12 | match | RegExpMatchArray \| null | conditionSite25 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilitiesPublic.ts:1037:9 | param.name | BindingName | conditionSite26 | 1 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/watch.ts:114:69 | sys | System | conditionSite27 | 1 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/watch.ts:790:9 | text.match(sourceMapCommentRegExpDontCareLineStart) | RegExpMatchArray \| null | conditionSite28 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/emitter.ts:484:9 | options.tsBuildInfoFile | string \| undefined | conditionSite29 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/emitter.ts:554:12 | outputDir | string \| undefined | conditionSite30 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/emitter.ts:697:9 | configFile.options.outFile | string \| undefined | conditionSite31 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/emitter.ts:725:9 | configFile.options.outFile | string \| undefined | conditionSite32 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/executeCommandLine.ts:241:31 | option.shortName | string \| undefined | conditionSite33 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/executeCommandLine.ts:448:9 | beforeOptionsDescription | string \| undefined | conditionSite34 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleNameResolver.ts:2135:13 | directory | string | conditionSite35 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleSpecifiers.ts:398:9 | ambient | string \| undefined | conditionSite36 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/parser.ts:10563:9 | standardExtension | string | conditionSite37 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/path.ts:385:12 | extension | string \| undefined | conditionSite38 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/path.ts:857:12 | pathext | string | conditionSite39 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/path.ts:872:9 | declarationExtension | string \| undefined | conditionSite40 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/program.ts:1142:33 | options.configFilePath | string \| undefined | conditionSite41 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:910:12 | subModuleName | string | conditionSite42 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:6596:12 | outputDir | string \| undefined | conditionSite43 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:7150:9 | currentLineText | string | conditionSite44 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:9431:12 | base | string \| undefined | conditionSite45 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/builder.ts:739:9 | emitOnlyDtsFiles | boolean \| undefined | conditionSite46 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/core.ts:2080:12 | ignoreCase | boolean \| undefined | conditionSite47 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/core.ts:2250:9 | ignoreCase | boolean \| undefined | conditionSite48 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/core.ts:2431:12 | ignoreCase | boolean \| undefined | conditionSite49 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/executeCommandLine.ts:1147:9 | canReportDiagnostics(system, compilerOptions) | boolean \| undefined | conditionSite50 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleNameResolver.ts:2061:40 | isFolder | boolean \| undefined | conditionSite51 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/program.ts:888:9 | decl.importClause?.isTypeOnly | boolean \| undefined | conditionSite52 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/scanner.ts:477:13 | allowEdits | true \| undefined | conditionSite53 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/scanner.ts:658:21 | stopAfterLineBreak | boolean \| undefined | conditionSite54 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/tsbuildPublic.ts:289:22 | pretty | boolean \| undefined | conditionSite55 | 3 | NotYet: a boolean \| undefined variable a function value captures |
| src/compiler/utilities.ts:5046:19 | excludeJSDocTypeAssertions | boolean \| undefined | conditionSite56 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:5647:20 | hasArguments | boolean \| undefined | conditionSite57 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:5954:20 | hasArguments | boolean \| undefined | conditionSite58 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:7477:13 | excludeCompoundAssignment | boolean \| undefined | conditionSite59 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/watch.ts:194:12 | pretty | boolean \| undefined | conditionSite60 | 3 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/factory/nodeFactory.ts:493:25 | flags & NodeFactoryFlags.NoOriginalNode | number | conditionSite61 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/factory/utilities.ts:1740:9 | node.transformFlags & TransformFlags.ContainsObjectRestOrSpread | number | conditionSite62 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleNameResolver.ts:195:9 | extensions & Extensions.TypeScript | number | conditionSite63 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleNameResolver.ts:204:9 | extensions & Extensions.TypeScript | number | conditionSite64 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleNameResolver.ts:301:12 | value.length | number | conditionSite65 | 2 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/parser.ts:485:12 | sourceFile.flags & NodeFlags.PossiblyContainsImportMeta | number | conditionSite66 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:8041:12 | symbol.flags & SymbolFlags.Transient | number | conditionSite67 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:8069:12 | symbol.flags & SymbolFlags.Alias | number | conditionSite68 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:8267:9 | symbol.flags & SymbolFlags.Class | number | conditionSite69 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:8281:12 | type.flags & TypeFlags.ObjectFlagsType | number | conditionSite70 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:8884:9 | res | Comparison | conditionSite71 | 5 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/utilities.ts:12439:17 | getCombinedNodeFlags(node.declarationList) & NodeFlags.BlockScoped | number | conditionSite72 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/moduleNameResolver.ts:3171:9 | matchedPattern | string \| Pattern \| undefined | conditionSite73 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/path.ts:438:9 | extensions | string \| readonly string[] \| undefined | conditionSite74 | 5 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/transformer.ts:128:9 | emitOnly | boolean \| EmitOnly \| undefined | conditionSite75 | 7 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/builder.ts:1630:9 | data?.diagnostics?.length | number \| undefined | conditionSite76 | 4 | No finding at this exact position; not proof its whole unit compiled |
| src/compiler/program.ts:941:17 | override | ResolutionMode | conditionSite77 | 6 | No finding at this exact position; not proof its whole unit compiled |
