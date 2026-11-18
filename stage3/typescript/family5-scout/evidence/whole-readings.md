# Whole representative site readings

Original source, with CRLF normalized to LF only in these Markdown excerpts. Original exact strings and locations are retained in the compressed JSON evidence.

## 1. G01 S105 transformers/legacyDecorators.ts:434:40

`transformers/legacyDecorators.ts:429:5`

```ts
function finishClassElement(updated: ClassElement, original: ClassElement) {
        if (updated !== original) {
            // While we emit the source map for the node after skipping decorators and modifiers,
            // we need to emit the comments for the original range.
            setCommentRange(updated, original);
            setSourceMapRange(updated, moveRangePastModifiers(original));
        }
        return updated;
    }
```

## 2. G02 S126 utilities.ts:7960:37

`utilities.ts:7959:1`

```ts
export function positionsAreOnSameLine(pos1: number, pos2: number, sourceFile: SourceFile): boolean {
    return getLinesBetweenPositions(sourceFile, pos1, pos2) === 0;
}
```

## 3. G03 S008 checker.ts:19948:30

`checker.ts:19946:33`

```ts
symbol => {
                if (symbol.flags & SymbolFlags.TypeParameter) {
                    result = append(result, getDeclaredTypeOfSymbol(symbol));
                }
            }
```

## 4. G05 S009 checker.ts:21355:119

`checker.ts:21354:5`

```ts
function checkTypeAssignableTo(source: Type, target: Type, errorNode: Node | undefined, headMessage?: DiagnosticMessage, containingMessageChain?: () => DiagnosticMessageChain | undefined, errorOutputObject?: { errors?: Diagnostic[] | undefined; }): boolean {
        return checkTypeRelatedTo(source, target, assignableRelation, errorNode, headMessage, containingMessageChain, errorOutputObject);
    }
```

## 5. G07 S033 moduleNameResolver.ts:2409:82

`moduleNameResolver.ts:2407:1`

```ts
function getVersionPathsOfPackageJsonInfo(packageJsonInfo: PackageJsonInfo, state: ModuleResolutionState): VersionPaths | undefined {
    if (packageJsonInfo.contents.versionPaths === undefined) {
        packageJsonInfo.contents.versionPaths = readPackageJsonTypesVersionPaths(packageJsonInfo.contents.packageJsonContent, state) || false;
    }
    return packageJsonInfo.contents.versionPaths || undefined;
}
```

## 6. G08 S018 checker.ts:2819:109

`checker.ts:2819:85`

```ts
r => compareDiagnostics(r, followOnMessage) === Comparison.EqualTo || compareDiagnostics(r, leadingMessage) === Comparison.EqualTo
```

## 7. G10 S048 parser.ts:10624:158

`parser.ts:10616:81`

```ts
arg => {
                    const { types, lib, path, ["resolution-mode"]: res, preserve: _preserve } = arg.arguments;
                    const preserve = _preserve === "true" ? true : undefined;
                    if (arg.arguments["no-default-lib"] === "true") {
                        // This has been removed; parse but ignore it.
                    }
                    else if (types) {
                        const parsed = parseResolutionMode(res, types.pos, types.end, reportDiagnostic);
                        typeReferenceDirectives.push({ pos: types.pos, end: types.end, fileName: types.value, ...(parsed ? { resolutionMode: parsed } : {}), ...(preserve ? { preserve } : {}) });
                    }
                    else if (lib) {
                        libReferenceDirectives.push({ pos: lib.pos, end: lib.end, fileName: lib.value, ...(preserve ? { preserve } : {}) });
                    }
                    else if (path) {
                        referencedFiles.push({ pos: path.pos, end: path.end, fileName: path.value, ...(preserve ? { preserve } : {}) });
                    }
                    else {
                        reportDiagnostic(arg.range.pos, arg.range.end - arg.range.pos, Diagnostics.Invalid_reference_directive_syntax);
                    }
                }
```

## 8. G13 S055 program.ts:1615:42

`program.ts:1615:42`

```ts
(moduleNames, containingFile, redirectedReference, options, containingSourceFile, reusedNames) =>
            host.resolveModuleNames!(
                moduleNames.map(getModuleResolutionName),
                containingFile,
                reusedNames?.map(getModuleResolutionName),
                redirectedReference,
                options,
                containingSourceFile,
            ).map(resolved =>
                resolved ?
                    ((resolved as ResolvedModuleFull).extension !== undefined) ?
                        { resolvedModule: resolved as ResolvedModuleFull } :
                        // An older host may have omitted extension, in which case we should infer it from the file extension of resolvedFileName.
                        { resolvedModule: { ...resolved, extension: extensionFromPath(resolved.resolvedFileName) } } :
                    emptyResolution
            )
```

## 9. G20 S027 commandLineParser.ts:2679:13

`commandLineParser.ts:2660:1`

```ts
export function convertToTSConfig(configParseResult: ParsedCommandLine, configFileName: string, host: ConvertToTSConfigHost): TSConfig {
    const getCanonicalFileName = createGetCanonicalFileName(host.useCaseSensitiveFileNames);
    const files = map(
        filter(
            configParseResult.fileNames,
            !configParseResult.options.configFile?.configFileSpecs?.validatedIncludeSpecs ? returnTrue : matchesSpecs(
                configFileName,
                configParseResult.options.configFile.configFileSpecs.validatedIncludeSpecs,
                configParseResult.options.configFile.configFileSpecs.validatedExcludeSpecs,
                host,
            ),
        ),
        f => getRelativePathFromFile(getNormalizedAbsolutePath(configFileName, host.getCurrentDirectory()), getNormalizedAbsolutePath(f, host.getCurrentDirectory()), getCanonicalFileName),
    );
    const pathOptions = { configFilePath: getNormalizedAbsolutePath(configFileName, host.getCurrentDirectory()), useCaseSensitiveFileNames: host.useCaseSensitiveFileNames };
    const optionMap = serializeCompilerOptions(configParseResult.options, pathOptions);
    const watchOptionMap = configParseResult.watchOptions && serializeWatchOptions(configParseResult.watchOptions);
    const config: TSConfig & { watchOptions?: object | undefined; } = {
        compilerOptions: {
            ...optionMapToObject(optionMap),
            showConfig: undefined,
            configFile: undefined,
            configFilePath: undefined,
            help: undefined,
            init: undefined,
            listFiles: undefined,
            listEmittedFiles: undefined,
            project: undefined,
            build: undefined,
            version: undefined,
        },
        watchOptions: watchOptionMap && optionMapToObject(watchOptionMap),
        references: map(configParseResult.projectReferences, r => ({ ...r, path: r.originalPath ? r.originalPath : "", originalPath: undefined })),
        files: length(files) ? files : undefined,
        ...(configParseResult.options.configFile?.configFileSpecs ? {
            include: filterSameAsDefaultInclude(configParseResult.options.configFile.configFileSpecs.validatedIncludeSpecs),
            exclude: configParseResult.options.configFile.configFileSpecs.validatedExcludeSpecs,
        } : {}),
        compileOnSave: !!configParseResult.compileOnSave ? true : undefined,
    };

    const providedKeys = new Set(optionMap.keys());
    const impliedCompilerOptions: Record<string, CompilerOptionsValue> = {};
    for (const option in computedOptions) {
        if (!providedKeys.has(option) && optionDependsOn(option, providedKeys)) {
            const implied = computedOptions[option]!.computeValue(configParseResult.options);
            const defaultValue = computedOptions[option]!.computeValue({});
            if (implied !== defaultValue) {
                impliedCompilerOptions[option] = computedOptions[option]!.computeValue(configParseResult.options);
            }
        }
    }
    assign(config.compilerOptions, optionMapToObject(serializeCompilerOptions(impliedCompilerOptions, pathOptions)));
    return config;
}
```

## 10. G21 S131 watch.ts:990:53

`watch.ts:987:1`

```ts
export function performIncrementalCompilation(input: IncrementalCompilationOptions): ExitStatus {
    const system = input.system || sys;
    const host = input.host || (input.host = createIncrementalCompilerHost(input.options, system));
    const builderProgram = createIncrementalProgram(input);
    const exitStatus = emitFilesAndReportErrorsAndGetExitStatus(
        builderProgram,
        input.reportDiagnostic || createDiagnosticReporter(system),
        s => host.trace && host.trace(s),
        input.reportErrorSummary || input.options.pretty ? (errorCount, filesInError) => system.write(getErrorSummaryText(errorCount, filesInError, system.newLine, host)) : undefined,
    );
    if (input.afterProgramEmitAndDiagnostics) input.afterProgramEmitAndDiagnostics(builderProgram);
    return exitStatus;
}
```

## 11. G26 S053 parser.ts:1827:40

`parser.ts:1820:5`

```ts
function withJSDoc<T extends HasJSDoc>(node: T, hasJSDoc: boolean): T {
        if (!hasJSDoc) {
            return node;
        }

        Debug.assert(!node.jsDoc); // Should only be called once per node
        const jsDoc = mapDefined(getJSDocCommentRanges(node, sourceText), comment => JSDocParser.parseJSDocComment(node, comment.pos, comment.end - comment.pos));
        if (jsDoc.length) node.jsDoc = jsDoc;
        if (hasDeprecatedTag) {
            hasDeprecatedTag = false;
            (node as Mutable<T>).flags |= NodeFlags.Deprecated;
        }
        return node;
    }
```

## 12. G29 S032 moduleNameResolver.ts:140:12

`moduleNameResolver.ts:127:1`

```ts
function withPackageId(packageInfo: PackageJsonInfo | undefined, r: PathAndExtension | undefined, state: ModuleResolutionState): Resolved | undefined {
    let packageId: PackageId | undefined;
    if (r && packageInfo) {
        const packageJsonContent = packageInfo.contents.packageJsonContent as PackageJson;
        if (typeof packageJsonContent.name === "string" && typeof packageJsonContent.version === "string") {
            packageId = {
                name: packageJsonContent.name,
                subModuleName: r.path.slice(packageInfo.packageDirectory.length + directorySeparator.length),
                version: packageJsonContent.version,
                peerDependencies: getPeerDependenciesOfPackageJsonInfo(packageInfo, state),
            };
        }
    }
    return r && { path: r.path, extension: r.ext, packageId, resolvedUsingTsExtension: r.resolvedUsingTsExtension };
}
```

## 13. G32 S002 builder.ts:559:14

`builder.ts:534:1`

```ts
function convertOrRepopulateDiagnosticMessageChain<T extends DiagnosticMessageChain | ReusableDiagnosticMessageChain>(
    chain: T,
    sourceFile: SourceFile | undefined,
    newProgram: Program,
    repopulateInfo: (chain: T) => RepopulateDiagnosticChainInfo | undefined,
): DiagnosticMessageChain {
    const info = repopulateInfo(chain);
    if (info === true) {
        return {
            ...createModeMismatchDetails(sourceFile!),
            next: convertOrRepopulateDiagnosticMessageChainArray(chain.next as T[], sourceFile, newProgram, repopulateInfo),
        };
    }
    else if (info) {
        return {
            ...createModuleNotFoundChain(sourceFile!, newProgram, info.moduleReference, info.mode, info.packageName || info.moduleReference),
            next: convertOrRepopulateDiagnosticMessageChainArray(chain.next as T[], sourceFile, newProgram, repopulateInfo),
        };
    }
    const next = convertOrRepopulateDiagnosticMessageChainArray(chain.next as T[], sourceFile, newProgram, repopulateInfo);
    if (next === chain.next) {
        Debug.assert(isFullChain(chain));
        return chain;
    }
    Debug.assert("messageText" in chain && "category" in chain && "code" in chain);
    return { ...chain, next };

    function isFullChain(value: T): value is T & DiagnosticMessageChain {
        return "messageText" in value && "category" in value && "code" in value && value.next === next;
    }
}
```

## 14. G33 S003 builder.ts:598:36

`builder.ts:597:28`

```ts
diagnostic => {
        const result: Diagnostic = convertToDiagnosticRelatedInformation(diagnostic, diagnosticFilePath, newProgram, toPathInBuildInfoDirectory);
        result.reportsUnnecessary = diagnostic.reportsUnnecessary;
        result.reportsDeprecated = diagnostic.reportDeprecated;
        result.source = diagnostic.source;
        result.skippedOn = diagnostic.skippedOn;
        const { relatedInformation } = diagnostic;
        result.relatedInformation = relatedInformation ?
            relatedInformation.length ?
                relatedInformation.map(r => convertToDiagnosticRelatedInformation(r, diagnosticFilePath, newProgram, toPathInBuildInfoDirectory)) :
                [] :
            undefined;
        return result;
    }
```

## 15. G34 S001 builder.ts:1528:48

`builder.ts:1527:32`

```ts
diagnostic => {
            const result: ReusableDiagnostic = toReusableDiagnosticRelatedInformation(diagnostic, diagnosticFilePath);
            result.reportsUnnecessary = diagnostic.reportsUnnecessary;
            result.reportDeprecated = diagnostic.reportsDeprecated;
            result.source = diagnostic.source;
            result.skippedOn = diagnostic.skippedOn;
            const { relatedInformation } = diagnostic;
            result.relatedInformation = relatedInformation ?
                relatedInformation.length ?
                    relatedInformation.map(r => toReusableDiagnosticRelatedInformation(r, diagnosticFilePath)) :
                    [] :
                undefined;
            return result;
        }
```

## 16. G43 S052 parser.ts:10760:45

`parser.ts:10750:1`

```ts
function addPragmaForMatch(pragmas: PragmaPseudoMapEntry[], range: CommentRange, kind: PragmaKindFlags, match: RegExpExecArray) {
    if (!match) return;
    const name = match[1]!.toLowerCase() as keyof PragmaPseudoMap; // Technically unsafe cast, but we do it so they below check to make it safe typechecks
    const pragma = commentPragmas[name] as PragmaDefinition;
    if (!pragma || !(pragma.kind! & kind)) {
        return;
    }
    const args = match[2]; // Split on spaces and match up positionally with definition
    const argument = getNamedPragmaArguments(pragma, args);
    if (argument === "fail") return; // Missing required argument, fail to parse it
    pragmas.push({ name, args: { arguments: argument, range } } as PragmaPseudoMapEntry);
    return;
}
```

## 17. G44 S051 parser.ts:10728:53

`parser.ts:10697:1`

```ts
function extractPragmas(pragmas: PragmaPseudoMapEntry[], range: CommentRange, text: string) {
    const tripleSlash = range.kind === SyntaxKind.SingleLineCommentTrivia && tripleSlashXMLCommentStartRegEx.exec(text);
    if (tripleSlash) {
        const name = tripleSlash[1]!.toLowerCase() as keyof PragmaPseudoMap; // Technically unsafe cast, but we do it so the below check to make it safe typechecks
        const pragma = commentPragmas[name] as PragmaDefinition;
        if (!pragma || !(pragma.kind! & PragmaKindFlags.TripleSlashXML)) {
            return;
        }
        if (pragma.args) {
            const argument: { [index: string]: string | { value: string; pos: number; end: number; }; } = {};
            for (const arg of pragma.args) {
                const matcher = getNamedArgRegEx(arg.name);
                const matchResult = matcher.exec(text);
                if (!matchResult && !arg.optional) {
                    return; // Missing required argument, don't parse
                }
                else if (matchResult) {
                    const value = (matchResult[2] ?? matchResult[3])!;
                    if (arg.captureSpan) {
                        const startPos = range.pos + matchResult.index + matchResult[1]!.length + 1;
                        argument[arg.name] = {
                            value,
                            pos: startPos,
                            end: startPos + value.length,
                        };
                    }
                    else {
                        argument[arg.name] = value;
                    }
                }
            }
            pragmas.push({ name, args: { arguments: argument, range } } as PragmaPseudoMapEntry);
        }
        else {
            pragmas.push({ name, args: { arguments: {}, range } } as PragmaPseudoMapEntry);
        }
        return;
    }

    const singleLine = range.kind === SyntaxKind.SingleLineCommentTrivia && singleLinePragmaRegEx.exec(text);
    if (singleLine) {
        return addPragmaForMatch(pragmas, range, PragmaKindFlags.SingleLine, singleLine);
    }

    if (range.kind === SyntaxKind.MultiLineCommentTrivia) {
        const multiLinePragmaRegEx = /@(\S+)(\s+(?:\S.*)?)?$/gm; // Defined inline since it uses the "g" flag, which keeps a persistent index (for iterating)
        let multiLineMatch: RegExpExecArray | null; // eslint-disable-line no-restricted-syntax
        while (multiLineMatch = multiLinePragmaRegEx.exec(text)) {
            addPragmaForMatch(pragmas, range, PragmaKindFlags.MultiLine, multiLineMatch);
        }
    }
}
```

## 18. G30 S040 moduleSpecifiers.ts:1277:30

`moduleSpecifiers.ts:1249:5`

```ts
function tryDirectoryWithPackageJson(packageRootIndex: number): { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } {
        const packageRootPath = path.substring(0, packageRootIndex);
        const packageJsonPath = combinePaths(packageRootPath, "package.json");
        let moduleFileToTry = path;
        let maybeBlockedByTypesVersions = false;
        const cachedPackageJson = host.getPackageJsonInfoCache?.()?.getPackageJsonInfo(packageJsonPath);
        if (isPackageJsonInfo(cachedPackageJson) || cachedPackageJson === undefined && host.fileExists(packageJsonPath)) {
            const packageJsonContent: Record<string, any> | undefined = cachedPackageJson?.contents.packageJsonContent || tryParseJson(host.readFile!(packageJsonPath)!);
            const importMode = overrideMode || getDefaultResolutionModeForFile(importingSourceFile, host, options);
            if (getResolvePackageJsonExports(options)) {
                // The package name that we found in node_modules could be different from the package
                // name in the package.json content via url/filepath dependency specifiers. We need to
                // use the actual directory name, so don't look at `packageJsonContent.name` here.
                const nodeModulesDirectoryName = packageRootPath.substring(parts.topLevelPackageNameIndex + 1);
                const packageName = getPackageNameFromTypesPackageName(nodeModulesDirectoryName);
                const conditions = getConditions(options, importMode);
                const fromExports = packageJsonContent?.exports
                    ? tryGetModuleNameFromExports(
                        options,
                        host,
                        path,
                        packageRootPath,
                        packageName,
                        packageJsonContent.exports,
                        conditions,
                    )
                    : undefined;
                if (fromExports) {
                    return { ...fromExports, verbatimFromExports: true };
                }
                if (packageJsonContent?.exports) {
                    return { moduleFileToTry: path, blockedByExports: true };
                }
            }
            const versionPaths = packageJsonContent?.typesVersions
                ? getPackageJsonTypesVersionsPaths(packageJsonContent.typesVersions)
                : undefined;
            if (versionPaths) {
                const subModuleName = path.slice(packageRootPath.length + 1);
                const fromPaths = tryGetModuleNameFromPaths(
                    subModuleName,
                    versionPaths.paths,
                    allowedEndings,
                    packageRootPath,
                    getCanonicalFileName,
                    host,
                    options,
                );
                if (fromPaths === undefined) {
                    maybeBlockedByTypesVersions = true;
                }
                else {
                    moduleFileToTry = combinePaths(packageRootPath, fromPaths);
                }
            }
            // If the file is the main module, it can be imported by the package name
            const mainFileRelative = packageJsonContent?.typings || packageJsonContent?.types || packageJsonContent?.main || "index.js";
            if (isString(mainFileRelative) && !(maybeBlockedByTypesVersions && matchPatternOrExact(tryParsePatterns(versionPaths!.paths), mainFileRelative))) {
                // The 'main' file is also subject to mapping through typesVersions, and we couldn't come up with a path
                // explicitly through typesVersions, so if it matches a key in typesVersions now, it's not reachable.
                // (The only way this can happen is if some file in a package that's not resolvable from outside the
                // package got pulled into the program anyway, e.g. transitively through a file that *is* reachable. It
                // happens very easily in fourslash tests though, since every test file listed gets included. See
                // importNameCodeFix_typesVersions.ts for an example.)
                const mainExportFile = toPath(mainFileRelative, packageRootPath, getCanonicalFileName);
                const canonicalModuleFileToTry = getCanonicalFileName(moduleFileToTry);
                if (removeFileExtension(mainExportFile) === removeFileExtension(canonicalModuleFileToTry)) {
                    // ^ An arbitrary removal of file extension for this comparison is almost certainly wrong
                    return { packageRootPath, moduleFileToTry };
                }
                else if (
                    packageJsonContent?.type !== "module" &&
                    !fileExtensionIsOneOf(canonicalModuleFileToTry, extensionsNotSupportingExtensionlessResolution) &&
                    startsWith(canonicalModuleFileToTry, mainExportFile) &&
                    getDirectoryPath(canonicalModuleFileToTry) === removeTrailingDirectorySeparator(mainExportFile) &&
                    removeFileExtension(getBaseFileName(canonicalModuleFileToTry)) === "index"
                ) {
                    // if mainExportFile is a directory, which contains moduleFileToTry, we just try index file
                    // example mainExportFile: `pkg/lib` and moduleFileToTry: `pkg/lib/index`, we can use packageRootPath
                    // but this behavior is deprecated for packages with "type": "module", so we only do this for packages without "type": "module"
                    // and make sure that the extension on index.{???} is something that supports omitting the extension
                    return { packageRootPath, moduleFileToTry };
                }
            }
        }
        else {
            // No package.json exists; an index.js will still resolve as the package name
            const fileName = getCanonicalFileName(moduleFileToTry.substring(parts.packageRootIndex + 1));
            if (fileName === "index.d.ts" || fileName === "index.js" || fileName === "index.ts" || fileName === "index.tsx") {
                return { moduleFileToTry, packageRootPath };
            }
        }
        return { moduleFileToTry };
    }
```
