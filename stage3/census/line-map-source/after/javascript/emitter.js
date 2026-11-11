import * as ts from "./_namespaces/ts.js";
import { base64encode, canHaveLocals, canIncludeBindAndCheckDiagnostics, cast, changeExtension, CharacterCodes, combinePaths, compareEmitHelpers, comparePaths, Comparison, computeCommonSourceDirectoryOfFilenames, computeLineStarts, contains, createBinaryExpressionTrampoline, createDiagnosticCollection, createGetCanonicalFileName, createSourceMapGenerator, createTextWriter, Debug, directorySeparator, emitDetachedComments, EmitFlags, EmitHint, emitNewLineBeforeLeadingCommentOfPosition, EmitOnly, emptyArray, ensureTrailingDirectorySeparator, escapeJsxAttributeString, escapeLeadingUnderscores, escapeNonAsciiString, escapeString, every, Extension, factory, fileExtensionIs, fileExtensionIsOneOf, filter, findIndex, firstOrUndefined, forEach, forEachChild, forEachLeadingCommentRange, forEachTrailingCommentRange, formatGeneratedName, formatGeneratedNamePart, GeneratedIdentifierFlags, getAreDeclarationMapsEnabled, getBaseFileName, getCommentRange, getConstantValue, getContainingNodeArray, getDeclarationEmitExtensionForPath, getDeclarationEmitOutputFilePath, getDirectoryPath, getEmitDeclarations, getEmitFlags, getEmitHelpers, getEmitModuleKind, getEmitModuleResolutionKind, getEmitScriptTarget, getExternalModuleName, getIdentifierTypeArguments, getInternalEmitFlags, getLeadingCommentRanges, getLineAndCharacterOfPosition, getLinesBetweenPositionAndNextNonWhitespaceCharacter, getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter, getLinesBetweenRangeEndAndRangeStart, getLineStarts, getLiteralText, GetLiteralTextFlags, getNewLineCharacter, getNodeForGeneratedName, getNodeId, getNormalizedAbsolutePath, getOriginalNode, getOwnEmitOutputFilePath, getParseTreeNode, getRelativePathFromDirectory, getRelativePathToDirectoryOrUrl, getRootLength, getShebang, getSnippetElement, getSourceFileOfNode, getSourceFilePathInNewDir, getSourceFilesToEmit, getSourceMapRange, getSourceTextOfNodeFromSourceFile, getStartsOnNewLine, getSyntheticLeadingComments, getSyntheticTrailingComments, getTextOfJSDocComment, getTextOfJsxNamespacedName, getTrailingCommentRanges, getTrailingSemicolonDeferringWriter, getTypeNode, guessIndentation, hasRecordedExternalHelpers, idText, InternalEmitFlags, isAccessExpression, isArray, isArrowFunction, isBinaryExpression, isBindingPattern, isBlock, isDeclarationFileName, isDecorator, isEmptyStatement, isExportAssignment, isExportSpecifier, isExpression, isFileLevelUniqueName, isFunctionLike, isGeneratedIdentifier, isGeneratedPrivateIdentifier, isIdentifier, isImportAttributes, isImportEqualsDeclaration, isIncrementalCompilation, isInJsonFile, isJSDocLikeText, isJsonSourceFile, isJsxClosingElement, isJsxNamespacedName, isJsxOpeningElement, isKeyword, isLet, isLiteralExpression, isMemberName, isModifier, isModuleDeclaration, isNodeDescendantOf, isNumericLiteral, isParenthesizedExpression, isPartiallyEmittedExpression, isPinnedComment, isPrivateIdentifier, isPrologueDirective, isRecognizedTripleSlashComment, isSourceFile, isSourceFileNotJson, isStringLiteral, isTemplateLiteralKind, isTokenKind, isTypeParameterDeclaration, isVarAwaitUsing, isVarConst, isVarUsing, JsxEmit, last, lastOrUndefined, length, ListFormat, makeIdentifierFromModuleName, memoize, ModuleKind, moveRangePastModifiers, NodeFlags, nodeIsSynthesized, noEmitNotification, noEmitSubstitution, normalizePath, normalizeSlashes, notImplemented, positionIsSynthesized, positionsAreOnSameLine, rangeEndIsOnSameLineAsRangeStart, rangeEndPositionsAreOnSameLine, rangeIsOnSingleLine, rangeStartPositionsAreOnSameLine, readJsonOrUndefined, removeFileExtension, resolvePath, ScriptTarget, setOriginalNode, setTextRange, setTextRangePosEnd, singleOrUndefined, skipPartiallyEmittedExpressions, skipTrivia, SnippetKind, some, supportedJSExtensionsFlat, SymbolFlags, SyntaxKind, sys, TokenFlags, tokenToString, toSorted, tracing, transformNodes, tryCast, version, writeCommentRange, writeFile, } from "./_namespaces/ts.js";
import * as performance from "./_namespaces/ts.performance.js";
const brackets = createBracketsMap();
/** @internal */
export function isBuildInfoFile(file) {
    return fileExtensionIs(file, Extension.TsBuildInfo);
}
/**
 * Iterates over the source files that are expected to have an emit output.
 *
 * @param host An EmitHost.
 * @param action The action to execute.
 * @param sourceFilesOrTargetSourceFile
 *   If an array, the full list of source files to emit.
 *   Else, calls `getSourceFilesToEmit` with the (optional) target source file to determine the list of source files to emit.
 *
 * @internal
 */
export function forEachEmittedFile(host, action, sourceFilesOrTargetSourceFile, forceDtsEmit = false, onlyBuildInfo, includeBuildInfo) {
    const sourceFiles = isArray(sourceFilesOrTargetSourceFile) ? sourceFilesOrTargetSourceFile : getSourceFilesToEmit(host, sourceFilesOrTargetSourceFile, forceDtsEmit);
    const options = host.getCompilerOptions();
    if (!onlyBuildInfo) {
        if (options.outFile) {
            if (sourceFiles.length) {
                const bundle = factory.createBundle(sourceFiles);
                const result = action(getOutputPathsFor(bundle, host, forceDtsEmit), bundle);
                if (result) {
                    return result;
                }
            }
        }
        else {
            for (const sourceFile of sourceFiles) {
                const result = action(getOutputPathsFor(sourceFile, host, forceDtsEmit), sourceFile);
                if (result) {
                    return result;
                }
            }
        }
    }
    if (includeBuildInfo) {
        const buildInfoPath = getTsBuildInfoEmitOutputFilePath(options);
        if (buildInfoPath)
            return action({ buildInfoPath }, /*sourceFileOrBundle*/ undefined);
    }
}
export function getTsBuildInfoEmitOutputFilePath(options) {
    const configFile = options.configFilePath;
    if (!canEmitTsBuildInfo(options))
        return undefined;
    if (options.tsBuildInfoFile)
        return options.tsBuildInfoFile;
    const outPath = options.outFile;
    let buildInfoExtensionLess;
    if (outPath) {
        buildInfoExtensionLess = removeFileExtension(outPath);
    }
    else {
        if (!configFile)
            return undefined;
        const configFileExtensionLess = removeFileExtension(configFile);
        buildInfoExtensionLess = options.outDir ?
            options.rootDir ?
                resolvePath(options.outDir, getRelativePathFromDirectory(options.rootDir, configFileExtensionLess, /*ignoreCase*/ true)) :
                combinePaths(options.outDir, getBaseFileName(configFileExtensionLess)) :
            configFileExtensionLess;
    }
    return buildInfoExtensionLess + Extension.TsBuildInfo;
}
function canEmitTsBuildInfo(options) {
    return isIncrementalCompilation(options) || !!options.tscBuild;
}
function getOutputPathsForBundle(options, forceDtsPaths) {
    const outPath = options.outFile;
    const jsFilePath = options.emitDeclarationOnly ? undefined : outPath;
    const sourceMapFilePath = jsFilePath && getSourceMapFilePath(jsFilePath, options);
    const declarationFilePath = (forceDtsPaths || getEmitDeclarations(options)) ? removeFileExtension(outPath) + Extension.Dts : undefined;
    const declarationMapPath = declarationFilePath && getAreDeclarationMapsEnabled(options) ? declarationFilePath + ".map" : undefined;
    return { jsFilePath, sourceMapFilePath, declarationFilePath, declarationMapPath };
}
/** @internal */
export function getOutputPathsFor(sourceFile, host, forceDtsPaths) {
    const options = host.getCompilerOptions();
    if (sourceFile.kind === SyntaxKind.Bundle) {
        return getOutputPathsForBundle(options, forceDtsPaths);
    }
    else {
        const ownOutputFilePath = getOwnEmitOutputFilePath(sourceFile.fileName, host, getOutputExtension(sourceFile.fileName, options));
        const isJsonFile = isJsonSourceFile(sourceFile);
        // If json file emits to the same location skip writing it, if emitDeclarationOnly skip writing it
        const isJsonEmittedToSameLocation = isJsonFile &&
            comparePaths(sourceFile.fileName, ownOutputFilePath, host.getCurrentDirectory(), !host.useCaseSensitiveFileNames()) === Comparison.EqualTo;
        const jsFilePath = options.emitDeclarationOnly || isJsonEmittedToSameLocation ? undefined : ownOutputFilePath;
        const sourceMapFilePath = !jsFilePath || isJsonSourceFile(sourceFile) ? undefined : getSourceMapFilePath(jsFilePath, options);
        const declarationFilePath = (forceDtsPaths || (getEmitDeclarations(options) && !isJsonFile)) ? getDeclarationEmitOutputFilePath(sourceFile.fileName, host) : undefined;
        const declarationMapPath = declarationFilePath && getAreDeclarationMapsEnabled(options) ? declarationFilePath + ".map" : undefined;
        return { jsFilePath, sourceMapFilePath, declarationFilePath, declarationMapPath };
    }
}
function getSourceMapFilePath(jsFilePath, options) {
    return (options.sourceMap && !options.inlineSourceMap) ? jsFilePath + ".map" : undefined;
}
/** @internal */
export function getOutputExtension(fileName, options) {
    return fileExtensionIs(fileName, Extension.Json) ? Extension.Json :
        options.jsx === JsxEmit.Preserve && fileExtensionIsOneOf(fileName, [Extension.Jsx, Extension.Tsx]) ? Extension.Jsx :
            fileExtensionIsOneOf(fileName, [Extension.Mts, Extension.Mjs]) ? Extension.Mjs :
                fileExtensionIsOneOf(fileName, [Extension.Cts, Extension.Cjs]) ? Extension.Cjs :
                    Extension.Js;
}
function getOutputPathWithoutChangingExt(inputFileName, ignoreCase, outputDir, getCommonSourceDirectory) {
    return outputDir ?
        resolvePath(outputDir, getRelativePathFromDirectory(getCommonSourceDirectory(), inputFileName, ignoreCase)) :
        inputFileName;
}
/** @internal */
export function getOutputDeclarationFileName(inputFileName, configFile, ignoreCase, getCommonSourceDirectory = () => getCommonSourceDirectoryOfConfig(configFile, ignoreCase)) {
    return getOutputDeclarationFileNameWorker(inputFileName, configFile.options, ignoreCase, getCommonSourceDirectory);
}
/** @internal */
export function getOutputDeclarationFileNameWorker(inputFileName, options, ignoreCase, getCommonSourceDirectory) {
    return changeExtension(getOutputPathWithoutChangingExt(inputFileName, ignoreCase, options.declarationDir || options.outDir, getCommonSourceDirectory), getDeclarationEmitExtensionForPath(inputFileName));
}
function getOutputJSFileName(inputFileName, configFile, ignoreCase, getCommonSourceDirectory = () => getCommonSourceDirectoryOfConfig(configFile, ignoreCase)) {
    if (configFile.options.emitDeclarationOnly)
        return undefined;
    const isJsonFile = fileExtensionIs(inputFileName, Extension.Json);
    const outputFileName = getOutputJSFileNameWorker(inputFileName, configFile.options, ignoreCase, getCommonSourceDirectory);
    return !isJsonFile || comparePaths(inputFileName, outputFileName, Debug.checkDefined(configFile.options.configFilePath), ignoreCase) !== Comparison.EqualTo ?
        outputFileName :
        undefined;
}
/** @internal */
export function getOutputJSFileNameWorker(inputFileName, options, ignoreCase, getCommonSourceDirectory) {
    return changeExtension(getOutputPathWithoutChangingExt(inputFileName, ignoreCase, options.outDir, getCommonSourceDirectory), getOutputExtension(inputFileName, options));
}
function createAddOutput() {
    let outputs;
    return { addOutput, getOutputs };
    function addOutput(path) {
        if (path) {
            (outputs || (outputs = [])).push(path);
        }
    }
    function getOutputs() {
        return outputs || emptyArray;
    }
}
function getSingleOutputFileNames(configFile, addOutput) {
    const { jsFilePath, sourceMapFilePath, declarationFilePath, declarationMapPath } = getOutputPathsForBundle(configFile.options, /*forceDtsPaths*/ false);
    addOutput(jsFilePath);
    addOutput(sourceMapFilePath);
    addOutput(declarationFilePath);
    addOutput(declarationMapPath);
}
function getOwnOutputFileNames(configFile, inputFileName, ignoreCase, addOutput, getCommonSourceDirectory) {
    if (isDeclarationFileName(inputFileName))
        return;
    const js = getOutputJSFileName(inputFileName, configFile, ignoreCase, getCommonSourceDirectory);
    addOutput(js);
    if (fileExtensionIs(inputFileName, Extension.Json))
        return;
    if (js && configFile.options.sourceMap) {
        addOutput(`${js}.map`);
    }
    if (getEmitDeclarations(configFile.options)) {
        const dts = getOutputDeclarationFileName(inputFileName, configFile, ignoreCase, getCommonSourceDirectory);
        addOutput(dts);
        if (configFile.options.declarationMap) {
            addOutput(`${dts}.map`);
        }
    }
}
/** @internal */
export function getCommonSourceDirectory(options, emittedFiles, currentDirectory, getCanonicalFileName, checkSourceFilesBelongToPath) {
    let commonSourceDirectory;
    if (options.rootDir) {
        // If a rootDir is specified use it as the commonSourceDirectory
        commonSourceDirectory = getNormalizedAbsolutePath(options.rootDir, currentDirectory);
        checkSourceFilesBelongToPath?.(options.rootDir);
    }
    else if (options.configFilePath) {
        // Project compilations never infer their root from the input source paths
        commonSourceDirectory = getDirectoryPath(normalizeSlashes(options.configFilePath));
        checkSourceFilesBelongToPath?.(commonSourceDirectory);
    }
    else {
        commonSourceDirectory = computeCommonSourceDirectoryOfFilenames(emittedFiles(), currentDirectory, getCanonicalFileName);
    }
    if (commonSourceDirectory && commonSourceDirectory[commonSourceDirectory.length - 1] !== directorySeparator) {
        // Make sure directory path ends with directory separator so this string can directly
        // used to replace with "" to get the relative path of the source file and the relative path doesn't
        // start with / making it rooted path
        commonSourceDirectory += directorySeparator;
    }
    return commonSourceDirectory;
}
/** @internal */
export function getComputedCommonSourceDirectory(emittedFiles, currentDirectory, getCanonicalFileName) {
    let commonSourceDirectory = computeCommonSourceDirectoryOfFilenames(emittedFiles, currentDirectory, getCanonicalFileName);
    if (commonSourceDirectory && commonSourceDirectory[commonSourceDirectory.length - 1] !== directorySeparator) {
        // Make sure directory path ends with directory separator so this string can directly
        // used to replace with "" to get the relative path of the source file and the relative path doesn't
        // start with / making it rooted path
        commonSourceDirectory += directorySeparator;
    }
    return commonSourceDirectory;
}
/** @internal */
export function getCommonSourceDirectoryOfConfig({ options, fileNames }, ignoreCase) {
    return getCommonSourceDirectory(options, () => filter(fileNames, file => !(options.noEmitForJsFiles && fileExtensionIsOneOf(file, supportedJSExtensionsFlat)) && !isDeclarationFileName(file)), getDirectoryPath(normalizeSlashes(Debug.checkDefined(options.configFilePath))), createGetCanonicalFileName(!ignoreCase));
}
/** @internal */
export function getAllProjectOutputs(configFile, ignoreCase) {
    const { addOutput, getOutputs } = createAddOutput();
    if (configFile.options.outFile) {
        getSingleOutputFileNames(configFile, addOutput);
    }
    else {
        const getCommonSourceDirectory = memoize(() => getCommonSourceDirectoryOfConfig(configFile, ignoreCase));
        for (const inputFileName of configFile.fileNames) {
            getOwnOutputFileNames(configFile, inputFileName, ignoreCase, addOutput, getCommonSourceDirectory);
        }
    }
    addOutput(getTsBuildInfoEmitOutputFilePath(configFile.options));
    return getOutputs();
}
export function getOutputFileNames(commandLine, inputFileName, ignoreCase) {
    inputFileName = normalizePath(inputFileName);
    Debug.assert(contains(commandLine.fileNames, inputFileName), `Expected fileName to be present in command line`);
    const { addOutput, getOutputs } = createAddOutput();
    if (commandLine.options.outFile) {
        getSingleOutputFileNames(commandLine, addOutput);
    }
    else {
        getOwnOutputFileNames(commandLine, inputFileName, ignoreCase, addOutput);
    }
    return getOutputs();
}
/** @internal */
export function getFirstProjectOutput(configFile, ignoreCase) {
    if (configFile.options.outFile) {
        const { jsFilePath, declarationFilePath } = getOutputPathsForBundle(configFile.options, /*forceDtsPaths*/ false);
        return Debug.checkDefined(jsFilePath || declarationFilePath, `project ${configFile.options.configFilePath} expected to have at least one output`);
    }
    const getCommonSourceDirectory = memoize(() => getCommonSourceDirectoryOfConfig(configFile, ignoreCase));
    for (const inputFileName of configFile.fileNames) {
        if (isDeclarationFileName(inputFileName))
            continue;
        const jsFilePath = getOutputJSFileName(inputFileName, configFile, ignoreCase, getCommonSourceDirectory);
        if (jsFilePath)
            return jsFilePath;
        if (fileExtensionIs(inputFileName, Extension.Json))
            continue;
        if (getEmitDeclarations(configFile.options)) {
            return getOutputDeclarationFileName(inputFileName, configFile, ignoreCase, getCommonSourceDirectory);
        }
    }
    const buildInfoPath = getTsBuildInfoEmitOutputFilePath(configFile.options);
    if (buildInfoPath)
        return buildInfoPath;
    return Debug.fail(`project ${configFile.options.configFilePath} expected to have at least one output`);
}
/** @internal */
export function emitResolverSkipsTypeChecking(emitOnly, forceDtsEmit) {
    return !!forceDtsEmit && !!emitOnly;
}
/** @internal */
// targetSourceFile is when users only want one file in entire project to be emitted. This is used in compileOnSave feature
export function emitFiles(resolver, host, targetSourceFile, { scriptTransformers, declarationTransformers }, emitOnly, onlyBuildInfo, forceDtsEmit, skipBuildInfo) {
    // Why var? It avoids TDZ checks in the runtime which can be costly.
    // See: https://github.com/microsoft/TypeScript/issues/52924
    /* eslint-disable no-var */
    var compilerOptions = host.getCompilerOptions();
    var sourceMapDataList = (compilerOptions.sourceMap || compilerOptions.inlineSourceMap || getAreDeclarationMapsEnabled(compilerOptions)) ? [] : undefined;
    var emittedFilesList = compilerOptions.listEmittedFiles ? [] : undefined;
    var emitterDiagnostics = createDiagnosticCollection();
    var newLine = getNewLineCharacter(compilerOptions);
    var writer = createTextWriter(newLine);
    var { enter, exit } = performance.createTimer("printTime", "beforePrint", "afterPrint");
    var emitSkipped = false;
    /* eslint-enable no-var */
    // Emit each output file
    enter();
    forEachEmittedFile(host, emitSourceFileOrBundle, getSourceFilesToEmit(host, targetSourceFile, forceDtsEmit), forceDtsEmit, onlyBuildInfo, !targetSourceFile && !skipBuildInfo);
    exit();
    return {
        emitSkipped,
        diagnostics: emitterDiagnostics.getDiagnostics(),
        emittedFiles: emittedFilesList,
        sourceMaps: sourceMapDataList,
    };
    function emitSourceFileOrBundle({ jsFilePath, sourceMapFilePath, declarationFilePath, declarationMapPath, buildInfoPath }, sourceFileOrBundle) {
        tracing?.push(tracing.Phase.Emit, "emitJsFileOrBundle", { jsFilePath });
        emitJsFileOrBundle(sourceFileOrBundle, jsFilePath, sourceMapFilePath);
        tracing?.pop();
        tracing?.push(tracing.Phase.Emit, "emitDeclarationFileOrBundle", { declarationFilePath });
        emitDeclarationFileOrBundle(sourceFileOrBundle, declarationFilePath, declarationMapPath);
        tracing?.pop();
        tracing?.push(tracing.Phase.Emit, "emitBuildInfo", { buildInfoPath });
        emitBuildInfo(buildInfoPath);
        tracing?.pop();
    }
    function emitBuildInfo(buildInfoPath) {
        // Write build information if applicable
        if (!buildInfoPath || targetSourceFile)
            return;
        if (host.isEmitBlocked(buildInfoPath)) {
            emitSkipped = true;
            return;
        }
        const buildInfo = host.getBuildInfo() || { version };
        // Pass buildinfo as additional data to avoid having to reparse
        writeFile(host, emitterDiagnostics, buildInfoPath, getBuildInfoText(buildInfo), /*writeByteOrderMark*/ false, /*sourceFiles*/ undefined, { buildInfo });
        emittedFilesList?.push(buildInfoPath);
    }
    function emitJsFileOrBundle(sourceFileOrBundle, jsFilePath, sourceMapFilePath) {
        if (!sourceFileOrBundle || emitOnly || !jsFilePath) {
            return;
        }
        // Make sure not to write js file and source map file if any of them cannot be written
        if (host.isEmitBlocked(jsFilePath) || compilerOptions.noEmit) {
            emitSkipped = true;
            return;
        }
        (isSourceFile(sourceFileOrBundle) ? [sourceFileOrBundle] : filter(sourceFileOrBundle.sourceFiles, isSourceFileNotJson)).forEach(sourceFile => {
            if (compilerOptions.noCheck ||
                !canIncludeBindAndCheckDiagnostics(sourceFile, compilerOptions))
                markLinkedReferences(sourceFile);
        });
        // Transform the source files
        const transform = transformNodes(resolver, host, factory, compilerOptions, [sourceFileOrBundle], scriptTransformers, /*allowDtsFiles*/ false);
        const printerOptions = {
            removeComments: compilerOptions.removeComments,
            newLine: compilerOptions.newLine,
            noEmitHelpers: compilerOptions.noEmitHelpers,
            module: getEmitModuleKind(compilerOptions),
            moduleResolution: getEmitModuleResolutionKind(compilerOptions),
            target: getEmitScriptTarget(compilerOptions),
            sourceMap: compilerOptions.sourceMap,
            inlineSourceMap: compilerOptions.inlineSourceMap,
            inlineSources: compilerOptions.inlineSources,
            extendedDiagnostics: compilerOptions.extendedDiagnostics,
        };
        // Create a printer to print the nodes
        const printer = createPrinter(printerOptions, {
            // resolver hooks
            hasGlobalName: resolver.hasGlobalName,
            // transform hooks
            onEmitNode: transform.emitNodeWithNotification,
            isEmitNotificationEnabled: transform.isEmitNotificationEnabled,
            substituteNode: transform.substituteNode,
        });
        Debug.assert(transform.transformed.length === 1, "Should only see one output from the transform");
        printSourceFileOrBundle(jsFilePath, sourceMapFilePath, transform, printer, compilerOptions);
        // Clean up emit nodes on parse tree
        transform.dispose();
        if (emittedFilesList) {
            emittedFilesList.push(jsFilePath);
            if (sourceMapFilePath) {
                emittedFilesList.push(sourceMapFilePath);
            }
        }
    }
    function emitDeclarationFileOrBundle(sourceFileOrBundle, declarationFilePath, declarationMapPath) {
        if (!sourceFileOrBundle || emitOnly === EmitOnly.Js)
            return;
        if (!declarationFilePath) {
            if (emitOnly || compilerOptions.emitDeclarationOnly)
                emitSkipped = true;
            return;
        }
        const sourceFiles = isSourceFile(sourceFileOrBundle) ? [sourceFileOrBundle] : sourceFileOrBundle.sourceFiles;
        const filesForEmit = forceDtsEmit ? sourceFiles : filter(sourceFiles, isSourceFileNotJson);
        // Setup and perform the transformation to retrieve declarations from the input files
        const inputListOrBundle = compilerOptions.outFile ? [factory.createBundle(filesForEmit)] : filesForEmit;
        // Checker wont collect the linked aliases since thats only done when declaration is enabled and checking is performed.
        // Do that here when emitting only dts files
        filesForEmit.forEach(sourceFile => {
            if ((emitOnly && !getEmitDeclarations(compilerOptions)) ||
                compilerOptions.noCheck ||
                emitResolverSkipsTypeChecking(emitOnly, forceDtsEmit) ||
                !canIncludeBindAndCheckDiagnostics(sourceFile, compilerOptions)) {
                collectLinkedAliases(sourceFile);
            }
        });
        const declarationTransform = transformNodes(resolver, host, factory, compilerOptions, inputListOrBundle, declarationTransformers, /*allowDtsFiles*/ false);
        if (length(declarationTransform.diagnostics)) {
            for (const diagnostic of declarationTransform.diagnostics) {
                emitterDiagnostics.add(diagnostic);
            }
        }
        const declBlocked = (!!declarationTransform.diagnostics && !!declarationTransform.diagnostics.length) || !!host.isEmitBlocked(declarationFilePath) || !!compilerOptions.noEmit;
        emitSkipped = emitSkipped || declBlocked;
        if (!declBlocked || forceDtsEmit) {
            Debug.assert(declarationTransform.transformed.length === 1, "Should only see one output from the decl transform");
            const printerOptions = {
                removeComments: compilerOptions.removeComments,
                newLine: compilerOptions.newLine,
                noEmitHelpers: true,
                module: compilerOptions.module,
                moduleResolution: compilerOptions.moduleResolution,
                target: compilerOptions.target,
                sourceMap: emitOnly !== EmitOnly.BuilderSignature && compilerOptions.declarationMap,
                inlineSourceMap: compilerOptions.inlineSourceMap,
                extendedDiagnostics: compilerOptions.extendedDiagnostics,
                onlyPrintJsDocStyle: true,
                omitBraceSourceMapPositions: true,
            };
            const declarationPrinter = createPrinter(printerOptions, {
                // resolver hooks
                hasGlobalName: resolver.hasGlobalName,
                // transform hooks
                onEmitNode: declarationTransform.emitNodeWithNotification,
                isEmitNotificationEnabled: declarationTransform.isEmitNotificationEnabled,
                substituteNode: declarationTransform.substituteNode,
            });
            const dtsWritten = printSourceFileOrBundle(declarationFilePath, declarationMapPath, declarationTransform, declarationPrinter, {
                sourceMap: printerOptions.sourceMap,
                sourceRoot: compilerOptions.sourceRoot,
                mapRoot: compilerOptions.mapRoot,
                extendedDiagnostics: compilerOptions.extendedDiagnostics,
                // Explicitly do not passthru either `inline` option
            });
            if (emittedFilesList) {
                if (dtsWritten)
                    emittedFilesList.push(declarationFilePath);
                if (declarationMapPath) {
                    emittedFilesList.push(declarationMapPath);
                }
            }
        }
        declarationTransform.dispose();
    }
    function collectLinkedAliases(node) {
        if (isExportAssignment(node)) {
            if (node.expression.kind === SyntaxKind.Identifier) {
                resolver.collectLinkedAliases(node.expression, /*setVisibility*/ true);
            }
            return;
        }
        else if (isExportSpecifier(node)) {
            resolver.collectLinkedAliases(node.propertyName || node.name, /*setVisibility*/ true);
            return;
        }
        forEachChild(node, collectLinkedAliases);
    }
    function markLinkedReferences(file) {
        if (ts.isSourceFileJS(file))
            return; // JS files don't use reference calculations as they don't do import ellision, no need to calculate it
        ts.forEachChildRecursively(file, n => {
            if (isImportEqualsDeclaration(n) && !(ts.getSyntacticModifierFlags(n) & ts.ModifierFlags.Export))
                return "skip"; // These are deferred and marked in a chain when referenced
            if (ts.isImportDeclaration(n))
                return "skip"; // likewise, these are ultimately what get marked by calls on other nodes - we want to skip them
            resolver.markLinkedReferences(n);
        });
    }
    function printSourceFileOrBundle(jsFilePath, sourceMapFilePath, transform, printer, mapOptions) {
        const sourceFileOrBundle = transform.transformed[0];
        const bundle = sourceFileOrBundle.kind === SyntaxKind.Bundle ? sourceFileOrBundle : undefined;
        const sourceFile = sourceFileOrBundle.kind === SyntaxKind.SourceFile ? sourceFileOrBundle : undefined;
        const sourceFiles = bundle ? bundle.sourceFiles : [sourceFile];
        let sourceMapGenerator;
        if (shouldEmitSourceMaps(mapOptions, sourceFileOrBundle)) {
            sourceMapGenerator = createSourceMapGenerator(host, getBaseFileName(normalizeSlashes(jsFilePath)), getSourceRoot(mapOptions), getSourceMapDirectory(mapOptions, jsFilePath, sourceFile), mapOptions);
        }
        if (bundle) {
            printer.writeBundle(bundle, writer, sourceMapGenerator);
        }
        else {
            printer.writeFile(sourceFile, writer, sourceMapGenerator);
        }
        let sourceMapUrlPos;
        if (sourceMapGenerator) {
            if (sourceMapDataList) {
                sourceMapDataList.push({
                    inputSourceFileNames: sourceMapGenerator.getSources(),
                    sourceMap: sourceMapGenerator.toJSON(),
                });
            }
            const sourceMappingURL = getSourceMappingURL(mapOptions, sourceMapGenerator, jsFilePath, sourceMapFilePath, sourceFile);
            if (sourceMappingURL) {
                if (!writer.isAtStartOfLine())
                    writer.rawWrite(newLine);
                sourceMapUrlPos = writer.getTextPos();
                writer.writeComment(`//# ${"sourceMappingURL"}=${sourceMappingURL}`); // Tools can sometimes see this line as a source mapping url comment
            }
            // Write the source map
            if (sourceMapFilePath) {
                const sourceMap = sourceMapGenerator.toString();
                writeFile(host, emitterDiagnostics, sourceMapFilePath, sourceMap, /*writeByteOrderMark*/ false, sourceFiles);
            }
        }
        else {
            writer.writeLine();
        }
        // Write the output file
        const text = writer.getText();
        const data = { sourceMapUrlPos, diagnostics: transform.diagnostics };
        writeFile(host, emitterDiagnostics, jsFilePath, text, !!compilerOptions.emitBOM, sourceFiles, data);
        // Reset state
        writer.clear();
        return !data.skippedDtsWrite;
    }
    function shouldEmitSourceMaps(mapOptions, sourceFileOrBundle) {
        return (mapOptions.sourceMap || mapOptions.inlineSourceMap)
            && (sourceFileOrBundle.kind !== SyntaxKind.SourceFile || !fileExtensionIs(sourceFileOrBundle.fileName, Extension.Json));
    }
    function getSourceRoot(mapOptions) {
        // Normalize source root and make sure it has trailing "/" so that it can be used to combine paths with the
        // relative paths of the sources list in the sourcemap
        const sourceRoot = normalizeSlashes(mapOptions.sourceRoot || "");
        return sourceRoot ? ensureTrailingDirectorySeparator(sourceRoot) : sourceRoot;
    }
    function getSourceMapDirectory(mapOptions, filePath, sourceFile) {
        if (mapOptions.sourceRoot)
            return host.getCommonSourceDirectory();
        if (mapOptions.mapRoot) {
            let sourceMapDir = normalizeSlashes(mapOptions.mapRoot);
            if (sourceFile) {
                // For modules or multiple emit files the mapRoot will have directory structure like the sources
                // So if src\a.ts and src\lib\b.ts are compiled together user would be moving the maps into mapRoot\a.js.map and mapRoot\lib\b.js.map
                sourceMapDir = getDirectoryPath(getSourceFilePathInNewDir(sourceFile.fileName, host, sourceMapDir));
            }
            if (getRootLength(sourceMapDir) === 0) {
                // The relative paths are relative to the common directory
                sourceMapDir = combinePaths(host.getCommonSourceDirectory(), sourceMapDir);
            }
            return sourceMapDir;
        }
        return getDirectoryPath(normalizePath(filePath));
    }
    function getSourceMappingURL(mapOptions, sourceMapGenerator, filePath, sourceMapFilePath, sourceFile) {
        if (mapOptions.inlineSourceMap) {
            // Encode the sourceMap into the sourceMap url
            const sourceMapText = sourceMapGenerator.toString();
            const base64SourceMapText = base64encode(sys, sourceMapText);
            return `data:application/json;base64,${base64SourceMapText}`;
        }
        const sourceMapFile = getBaseFileName(normalizeSlashes(Debug.checkDefined(sourceMapFilePath)));
        if (mapOptions.mapRoot) {
            let sourceMapDir = normalizeSlashes(mapOptions.mapRoot);
            if (sourceFile) {
                // For modules or multiple emit files the mapRoot will have directory structure like the sources
                // So if src\a.ts and src\lib\b.ts are compiled together user would be moving the maps into mapRoot\a.js.map and mapRoot\lib\b.js.map
                sourceMapDir = getDirectoryPath(getSourceFilePathInNewDir(sourceFile.fileName, host, sourceMapDir));
            }
            if (getRootLength(sourceMapDir) === 0) {
                // The relative paths are relative to the common directory
                sourceMapDir = combinePaths(host.getCommonSourceDirectory(), sourceMapDir);
                return encodeURI(getRelativePathToDirectoryOrUrl(getDirectoryPath(normalizePath(filePath)), // get the relative sourceMapDir path based on jsFilePath
                combinePaths(sourceMapDir, sourceMapFile), // this is where user expects to see sourceMap
                host.getCurrentDirectory(), host.getCanonicalFileName, 
                /*isAbsolutePathAnUrl*/ true));
            }
            else {
                return encodeURI(combinePaths(sourceMapDir, sourceMapFile));
            }
        }
        return encodeURI(sourceMapFile);
    }
}
/** @internal */
export function getBuildInfoText(buildInfo) {
    return JSON.stringify(buildInfo);
}
/** @internal */
export function getBuildInfo(buildInfoFile, buildInfoText) {
    return readJsonOrUndefined(buildInfoFile, buildInfoText);
}
/** @internal */
export const notImplementedResolver = {
    hasGlobalName: notImplemented,
    getReferencedExportContainer: notImplemented,
    getReferencedImportDeclaration: notImplemented,
    getReferencedDeclarationWithCollidingName: notImplemented,
    isDeclarationWithCollidingName: notImplemented,
    isValueAliasDeclaration: notImplemented,
    isReferencedAliasDeclaration: notImplemented,
    isTopLevelValueImportEqualsWithEntityName: notImplemented,
    hasNodeCheckFlag: notImplemented,
    isDeclarationVisible: notImplemented,
    isLateBound: (_node) => false,
    collectLinkedAliases: notImplemented,
    markLinkedReferences: notImplemented,
    isImplementationOfOverload: notImplemented,
    requiresAddingImplicitUndefined: notImplemented,
    isExpandoFunctionDeclaration: notImplemented,
    getPropertiesOfContainerFunction: notImplemented,
    createTypeOfDeclaration: notImplemented,
    createReturnTypeOfSignatureDeclaration: notImplemented,
    createTypeOfExpression: notImplemented,
    createLiteralConstValue: notImplemented,
    isSymbolAccessible: notImplemented,
    isEntityNameVisible: notImplemented,
    // Returns the constant value this property access resolves to: notImplemented, or 'undefined' for a non-constant
    getConstantValue: notImplemented,
    getEnumMemberValue: notImplemented,
    getReferencedValueDeclaration: notImplemented,
    getReferencedValueDeclarations: notImplemented,
    getTypeReferenceSerializationKind: notImplemented,
    isOptionalParameter: notImplemented,
    isArgumentsLocalBinding: notImplemented,
    getExternalModuleFileFromDeclaration: notImplemented,
    isLiteralConstDeclaration: notImplemented,
    getJsxFactoryEntity: notImplemented,
    getJsxFragmentFactoryEntity: notImplemented,
    isBindingCapturedByNode: notImplemented,
    getDeclarationStatementsForSourceFile: notImplemented,
    isImportRequiredByAugmentation: notImplemented,
    isDefinitelyReferenceToGlobalSymbolObject: notImplemented,
    createLateBoundIndexSignatures: notImplemented,
    symbolToDeclarations: notImplemented,
};
var PipelinePhase;
(function (PipelinePhase) {
    PipelinePhase[PipelinePhase["Notification"] = 0] = "Notification";
    PipelinePhase[PipelinePhase["Substitution"] = 1] = "Substitution";
    PipelinePhase[PipelinePhase["Comments"] = 2] = "Comments";
    PipelinePhase[PipelinePhase["SourceMaps"] = 3] = "SourceMaps";
    PipelinePhase[PipelinePhase["Emit"] = 4] = "Emit";
})(PipelinePhase || (PipelinePhase = {}));
/** @internal */
export const createPrinterWithDefaults = /* @__PURE__ */ memoize(() => createPrinter({}));
/** @internal */
export const createPrinterWithRemoveComments = /* @__PURE__ */ memoize(() => createPrinter({ removeComments: true }));
/** @internal */
export const createPrinterWithRemoveCommentsNeverAsciiEscape = /* @__PURE__ */ memoize(() => createPrinter({ removeComments: true, neverAsciiEscape: true }));
/** @internal */
export const createPrinterWithRemoveCommentsOmitTrailingSemicolon = /* @__PURE__ */ memoize(() => createPrinter({ removeComments: true, omitTrailingSemicolon: true }));
export function createPrinter(printerOptions = {}, handlers = {}) {
    // Why var? It avoids TDZ checks in the runtime which can be costly.
    // See: https://github.com/microsoft/TypeScript/issues/52924
    /* eslint-disable no-var */
    var { hasGlobalName, onEmitNode = noEmitNotification, isEmitNotificationEnabled, substituteNode = noEmitSubstitution, onBeforeEmitNode, onAfterEmitNode, onBeforeEmitNodeArray, onAfterEmitNodeArray, onBeforeEmitToken, onAfterEmitToken, } = handlers;
    var extendedDiagnostics = !!printerOptions.extendedDiagnostics;
    var omitBraceSourcePositions = !!printerOptions.omitBraceSourceMapPositions;
    var newLine = getNewLineCharacter(printerOptions);
    var moduleKind = getEmitModuleKind(printerOptions);
    var bundledHelpers = new Map();
    var currentSourceFile;
    var nodeIdToGeneratedName; // Map of generated names for specific nodes.
    var nodeIdToGeneratedPrivateName; // Map of generated names for specific nodes.
    var autoGeneratedIdToGeneratedName; // Map of generated names for temp and loop variables.
    var generatedNames; // Set of names generated by the NameGenerator.
    var formattedNameTempFlagsStack;
    var formattedNameTempFlags;
    var privateNameTempFlagsStack; // Stack of enclosing name generation scopes.
    var privateNameTempFlags; // TempFlags for the current name generation scope.
    var tempFlagsStack; // Stack of enclosing name generation scopes.
    var tempFlags; // TempFlags for the current name generation scope.
    var reservedNamesStack; // Stack of reserved names in enclosing name generation scopes.
    var reservedNames; // Names reserved in nested name generation scopes.
    var reservedPrivateNamesStack; // Stack of reserved member names in enclosing name generation scopes.
    var reservedPrivateNames; // Member names reserved in nested name generation scopes.
    var preserveSourceNewlines = printerOptions.preserveSourceNewlines; // Can be overridden inside nodes with the `IgnoreSourceNewlines` emit flag.
    var nextListElementPos; // See comment in `getLeadingLineTerminatorCount`.
    var writer;
    var ownWriter; // Reusable `EmitTextWriter` for basic printing.
    var write = writeBase;
    var isOwnFileEmit;
    // Source Maps
    var sourceMapsDisabled = true;
    var sourceMapGenerator;
    var sourceMapSource;
    var sourceMapSourceIndex = -1;
    var mostRecentlyAddedSourceMapSource;
    var mostRecentlyAddedSourceMapSourceIndex = -1;
    // Comments
    var containerPos = -1;
    var containerEnd = -1;
    var declarationListContainerEnd = -1;
    var currentLineMap;
    var detachedCommentsInfo;
    var hasWrittenComment = false;
    var commentsDisabled = !!printerOptions.removeComments;
    var lastSubstitution;
    var currentParenthesizerRule;
    var { enter: enterComment, exit: exitComment } = performance.createTimerIf(extendedDiagnostics, "commentTime", "beforeComment", "afterComment");
    var parenthesizer = factory.parenthesizer;
    var typeArgumentParenthesizerRuleSelector = {
        select: index => index === 0 ? parenthesizer.parenthesizeLeadingTypeArgument : undefined,
    };
    var emitBinaryExpression = createEmitBinaryExpression();
    /* eslint-enable no-var */
    reset();
    return {
        // public API
        printNode,
        printList,
        printFile,
        printBundle,
        // internal API
        writeNode,
        writeList,
        writeFile,
        writeBundle,
    };
    function printNode(hint, node, sourceFile) {
        switch (hint) {
            case EmitHint.SourceFile:
                Debug.assert(isSourceFile(node), "Expected a SourceFile node.");
                break;
            case EmitHint.IdentifierName:
                Debug.assert(isIdentifier(node), "Expected an Identifier node.");
                break;
            case EmitHint.Expression:
                Debug.assert(isExpression(node), "Expected an Expression node.");
                break;
        }
        switch (node.kind) {
            case SyntaxKind.SourceFile:
                return printFile(node);
            case SyntaxKind.Bundle:
                return printBundle(node);
        }
        writeNode(hint, node, sourceFile, beginPrint());
        return endPrint();
    }
    function printList(format, nodes, sourceFile) {
        writeList(format, nodes, sourceFile, beginPrint());
        return endPrint();
    }
    function printBundle(bundle) {
        writeBundle(bundle, beginPrint(), /*sourceMapGenerator*/ undefined);
        return endPrint();
    }
    function printFile(sourceFile) {
        writeFile(sourceFile, beginPrint(), /*sourceMapGenerator*/ undefined);
        return endPrint();
    }
    function writeNode(hint, node, sourceFile, output) {
        const previousWriter = writer;
        setWriter(output, /*_sourceMapGenerator*/ undefined);
        print(hint, node, sourceFile);
        reset();
        writer = previousWriter;
    }
    function writeList(format, nodes, sourceFile, output) {
        const previousWriter = writer;
        setWriter(output, /*_sourceMapGenerator*/ undefined);
        if (sourceFile) {
            setSourceFile(sourceFile);
        }
        emitList(/*parentNode*/ undefined, nodes, format);
        reset();
        writer = previousWriter;
    }
    function writeBundle(bundle, output, sourceMapGenerator) {
        isOwnFileEmit = false;
        const previousWriter = writer;
        setWriter(output, sourceMapGenerator);
        emitShebangIfNeeded(bundle);
        emitPrologueDirectivesIfNeeded(bundle);
        emitHelpers(bundle);
        emitSyntheticTripleSlashReferencesIfNeeded(bundle);
        for (const sourceFile of bundle.sourceFiles) {
            print(EmitHint.SourceFile, sourceFile, sourceFile);
        }
        reset();
        writer = previousWriter;
    }
    function writeFile(sourceFile, output, sourceMapGenerator) {
        isOwnFileEmit = true;
        const previousWriter = writer;
        setWriter(output, sourceMapGenerator);
        emitShebangIfNeeded(sourceFile);
        emitPrologueDirectivesIfNeeded(sourceFile);
        print(EmitHint.SourceFile, sourceFile, sourceFile);
        reset();
        writer = previousWriter;
    }
    function beginPrint() {
        return ownWriter || (ownWriter = createTextWriter(newLine));
    }
    function endPrint() {
        const text = ownWriter.getText();
        ownWriter.clear();
        return text;
    }
    function print(hint, node, sourceFile) {
        if (sourceFile) {
            setSourceFile(sourceFile);
        }
        pipelineEmit(hint, node, /*parenthesizerRule*/ undefined);
    }
    function setSourceFile(sourceFile) {
        currentSourceFile = sourceFile;
        currentLineMap = undefined;
        detachedCommentsInfo = undefined;
        if (sourceFile) {
            setSourceMapSource(sourceFile);
        }
    }
    function setWriter(_writer, _sourceMapGenerator) {
        if (_writer && printerOptions.omitTrailingSemicolon) {
            _writer = getTrailingSemicolonDeferringWriter(_writer);
        }
        writer = _writer; // TODO: GH#18217
        sourceMapGenerator = _sourceMapGenerator;
        sourceMapsDisabled = !writer || !sourceMapGenerator;
    }
    function reset() {
        nodeIdToGeneratedName = [];
        nodeIdToGeneratedPrivateName = [];
        autoGeneratedIdToGeneratedName = [];
        generatedNames = new Set();
        formattedNameTempFlagsStack = [];
        formattedNameTempFlags = new Map();
        privateNameTempFlagsStack = [];
        privateNameTempFlags = TempFlags.Auto;
        tempFlagsStack = [];
        tempFlags = TempFlags.Auto;
        reservedNamesStack = [];
        reservedNames = undefined;
        reservedPrivateNamesStack = [];
        reservedPrivateNames = undefined;
        currentSourceFile = undefined;
        currentLineMap = undefined;
        detachedCommentsInfo = undefined;
        setWriter(/*output*/ undefined, /*_sourceMapGenerator*/ undefined);
    }
    function getCurrentLineMap() {
        return currentLineMap || (currentLineMap = getLineStarts(Debug.checkDefined(currentSourceFile)));
    }
    function emit(node, parenthesizerRule) {
        if (node === undefined)
            return;
        pipelineEmit(EmitHint.Unspecified, node, parenthesizerRule);
    }
    function emitIdentifierName(node) {
        if (node === undefined)
            return;
        pipelineEmit(EmitHint.IdentifierName, node, /*parenthesizerRule*/ undefined);
    }
    function emitExpression(node, parenthesizerRule) {
        if (node === undefined)
            return;
        pipelineEmit(EmitHint.Expression, node, parenthesizerRule);
    }
    function emitJsxAttributeValue(node) {
        pipelineEmit(isStringLiteral(node) ? EmitHint.JsxAttributeValue : EmitHint.Unspecified, node);
    }
    function beforeEmitNode(node) {
        if (preserveSourceNewlines && (getInternalEmitFlags(node) & InternalEmitFlags.IgnoreSourceNewlines)) {
            preserveSourceNewlines = false;
        }
    }
    function afterEmitNode(savedPreserveSourceNewlines) {
        preserveSourceNewlines = savedPreserveSourceNewlines;
    }
    function pipelineEmit(emitHint, node, parenthesizerRule) {
        currentParenthesizerRule = parenthesizerRule;
        const pipelinePhase = getPipelinePhase(PipelinePhase.Notification, emitHint, node);
        pipelinePhase(emitHint, node);
        currentParenthesizerRule = undefined;
    }
    function shouldEmitComments(node) {
        return !commentsDisabled && !isSourceFile(node);
    }
    function shouldEmitSourceMaps(node) {
        return !sourceMapsDisabled &&
            !isSourceFile(node) &&
            !isInJsonFile(node);
    }
    function getPipelinePhase(phase, emitHint, node) {
        switch (phase) {
            case PipelinePhase.Notification:
                if (onEmitNode !== noEmitNotification && (!isEmitNotificationEnabled || isEmitNotificationEnabled(node))) {
                    return pipelineEmitWithNotification;
                }
            // falls through
            case PipelinePhase.Substitution:
                if (substituteNode !== noEmitSubstitution && (lastSubstitution = substituteNode(emitHint, node) || node) !== node) {
                    if (currentParenthesizerRule) {
                        lastSubstitution = currentParenthesizerRule(lastSubstitution);
                    }
                    return pipelineEmitWithSubstitution;
                }
            // falls through
            case PipelinePhase.Comments:
                if (shouldEmitComments(node)) {
                    return pipelineEmitWithComments;
                }
            // falls through
            case PipelinePhase.SourceMaps:
                if (shouldEmitSourceMaps(node)) {
                    return pipelineEmitWithSourceMaps;
                }
            // falls through
            case PipelinePhase.Emit:
                return pipelineEmitWithHint;
            default:
                return Debug.assertNever(phase);
        }
    }
    function getNextPipelinePhase(currentPhase, emitHint, node) {
        return getPipelinePhase(currentPhase + 1, emitHint, node);
    }
    function pipelineEmitWithNotification(hint, node) {
        const pipelinePhase = getNextPipelinePhase(PipelinePhase.Notification, hint, node);
        onEmitNode(hint, node, pipelinePhase);
    }
    function pipelineEmitWithHint(hint, node) {
        onBeforeEmitNode?.(node);
        if (preserveSourceNewlines) {
            const savedPreserveSourceNewlines = preserveSourceNewlines;
            beforeEmitNode(node);
            pipelineEmitWithHintWorker(hint, node);
            afterEmitNode(savedPreserveSourceNewlines);
        }
        else {
            pipelineEmitWithHintWorker(hint, node);
        }
        onAfterEmitNode?.(node);
        // clear the parenthesizer rule as we ascend
        currentParenthesizerRule = undefined;
    }
    function pipelineEmitWithHintWorker(hint, node, allowSnippets = true) {
        if (allowSnippets) {
            const snippet = getSnippetElement(node);
            if (snippet) {
                return emitSnippetNode(hint, node, snippet);
            }
        }
        if (hint === EmitHint.SourceFile)
            return emitSourceFile(cast(node, isSourceFile));
        if (hint === EmitHint.IdentifierName)
            return emitIdentifier(cast(node, isIdentifier));
        if (hint === EmitHint.JsxAttributeValue)
            return emitLiteral(cast(node, isStringLiteral), /*jsxAttributeEscape*/ true);
        if (hint === EmitHint.MappedTypeParameter)
            return emitMappedTypeParameter(cast(node, isTypeParameterDeclaration));
        if (hint === EmitHint.ImportTypeNodeAttributes)
            return emitImportTypeNodeAttributes(cast(node, isImportAttributes));
        if (hint === EmitHint.EmbeddedStatement) {
            Debug.assertNode(node, isEmptyStatement);
            return emitEmptyStatement(/*isEmbeddedStatement*/ true);
        }
        if (hint === EmitHint.Unspecified) {
            switch (node.kind) {
                // Pseudo-literals
                case SyntaxKind.TemplateHead:
                case SyntaxKind.TemplateMiddle:
                case SyntaxKind.TemplateTail:
                    return emitLiteral(node, /*jsxAttributeEscape*/ false);
                // Identifiers
                case SyntaxKind.Identifier:
                    return emitIdentifier(node);
                // PrivateIdentifiers
                case SyntaxKind.PrivateIdentifier:
                    return emitPrivateIdentifier(node);
                // Parse tree nodes
                // Names
                case SyntaxKind.QualifiedName:
                    return emitQualifiedName(node);
                case SyntaxKind.ComputedPropertyName:
                    return emitComputedPropertyName(node);
                // Signature elements
                case SyntaxKind.TypeParameter:
                    return emitTypeParameter(node);
                case SyntaxKind.Parameter:
                    return emitParameter(node);
                case SyntaxKind.Decorator:
                    return emitDecorator(node);
                // Type members
                case SyntaxKind.PropertySignature:
                    return emitPropertySignature(node);
                case SyntaxKind.PropertyDeclaration:
                    return emitPropertyDeclaration(node);
                case SyntaxKind.MethodSignature:
                    return emitMethodSignature(node);
                case SyntaxKind.MethodDeclaration:
                    return emitMethodDeclaration(node);
                case SyntaxKind.ClassStaticBlockDeclaration:
                    return emitClassStaticBlockDeclaration(node);
                case SyntaxKind.Constructor:
                    return emitConstructor(node);
                case SyntaxKind.GetAccessor:
                case SyntaxKind.SetAccessor:
                    return emitAccessorDeclaration(node);
                case SyntaxKind.CallSignature:
                    return emitCallSignature(node);
                case SyntaxKind.ConstructSignature:
                    return emitConstructSignature(node);
                case SyntaxKind.IndexSignature:
                    return emitIndexSignature(node);
                // Types
                case SyntaxKind.TypePredicate:
                    return emitTypePredicate(node);
                case SyntaxKind.TypeReference:
                    return emitTypeReference(node);
                case SyntaxKind.FunctionType:
                    return emitFunctionType(node);
                case SyntaxKind.ConstructorType:
                    return emitConstructorType(node);
                case SyntaxKind.TypeQuery:
                    return emitTypeQuery(node);
                case SyntaxKind.TypeLiteral:
                    return emitTypeLiteral(node);
                case SyntaxKind.ArrayType:
                    return emitArrayType(node);
                case SyntaxKind.TupleType:
                    return emitTupleType(node);
                case SyntaxKind.OptionalType:
                    return emitOptionalType(node);
                // SyntaxKind.RestType is handled below
                case SyntaxKind.UnionType:
                    return emitUnionType(node);
                case SyntaxKind.IntersectionType:
                    return emitIntersectionType(node);
                case SyntaxKind.ConditionalType:
                    return emitConditionalType(node);
                case SyntaxKind.InferType:
                    return emitInferType(node);
                case SyntaxKind.ParenthesizedType:
                    return emitParenthesizedType(node);
                case SyntaxKind.ExpressionWithTypeArguments:
                    return emitExpressionWithTypeArguments(node);
                case SyntaxKind.ThisType:
                    return emitThisType();
                case SyntaxKind.TypeOperator:
                    return emitTypeOperator(node);
                case SyntaxKind.IndexedAccessType:
                    return emitIndexedAccessType(node);
                case SyntaxKind.MappedType:
                    return emitMappedType(node);
                case SyntaxKind.LiteralType:
                    return emitLiteralType(node);
                case SyntaxKind.NamedTupleMember:
                    return emitNamedTupleMember(node);
                case SyntaxKind.TemplateLiteralType:
                    return emitTemplateType(node);
                case SyntaxKind.TemplateLiteralTypeSpan:
                    return emitTemplateTypeSpan(node);
                case SyntaxKind.ImportType:
                    return emitImportTypeNode(node);
                // Binding patterns
                case SyntaxKind.ObjectBindingPattern:
                    return emitObjectBindingPattern(node);
                case SyntaxKind.ArrayBindingPattern:
                    return emitArrayBindingPattern(node);
                case SyntaxKind.BindingElement:
                    return emitBindingElement(node);
                // Misc
                case SyntaxKind.TemplateSpan:
                    return emitTemplateSpan(node);
                case SyntaxKind.SemicolonClassElement:
                    return emitSemicolonClassElement();
                // Statements
                case SyntaxKind.Block:
                    return emitBlock(node);
                case SyntaxKind.VariableStatement:
                    return emitVariableStatement(node);
                case SyntaxKind.EmptyStatement:
                    return emitEmptyStatement(/*isEmbeddedStatement*/ false);
                case SyntaxKind.ExpressionStatement:
                    return emitExpressionStatement(node);
                case SyntaxKind.IfStatement:
                    return emitIfStatement(node);
                case SyntaxKind.DoStatement:
                    return emitDoStatement(node);
                case SyntaxKind.WhileStatement:
                    return emitWhileStatement(node);
                case SyntaxKind.ForStatement:
                    return emitForStatement(node);
                case SyntaxKind.ForInStatement:
                    return emitForInStatement(node);
                case SyntaxKind.ForOfStatement:
                    return emitForOfStatement(node);
                case SyntaxKind.ContinueStatement:
                    return emitContinueStatement(node);
                case SyntaxKind.BreakStatement:
                    return emitBreakStatement(node);
                case SyntaxKind.ReturnStatement:
                    return emitReturnStatement(node);
                case SyntaxKind.WithStatement:
                    return emitWithStatement(node);
                case SyntaxKind.SwitchStatement:
                    return emitSwitchStatement(node);
                case SyntaxKind.LabeledStatement:
                    return emitLabeledStatement(node);
                case SyntaxKind.ThrowStatement:
                    return emitThrowStatement(node);
                case SyntaxKind.TryStatement:
                    return emitTryStatement(node);
                case SyntaxKind.DebuggerStatement:
                    return emitDebuggerStatement(node);
                // Declarations
                case SyntaxKind.VariableDeclaration:
                    return emitVariableDeclaration(node);
                case SyntaxKind.VariableDeclarationList:
                    return emitVariableDeclarationList(node);
                case SyntaxKind.FunctionDeclaration:
                    return emitFunctionDeclaration(node);
                case SyntaxKind.ClassDeclaration:
                    return emitClassDeclaration(node);
                case SyntaxKind.InterfaceDeclaration:
                    return emitInterfaceDeclaration(node);
                case SyntaxKind.TypeAliasDeclaration:
                    return emitTypeAliasDeclaration(node);
                case SyntaxKind.EnumDeclaration:
                    return emitEnumDeclaration(node);
                case SyntaxKind.ModuleDeclaration:
                    return emitModuleDeclaration(node);
                case SyntaxKind.ModuleBlock:
                    return emitModuleBlock(node);
                case SyntaxKind.CaseBlock:
                    return emitCaseBlock(node);
                case SyntaxKind.NamespaceExportDeclaration:
                    return emitNamespaceExportDeclaration(node);
                case SyntaxKind.ImportEqualsDeclaration:
                    return emitImportEqualsDeclaration(node);
                case SyntaxKind.ImportDeclaration:
                    return emitImportDeclaration(node);
                case SyntaxKind.ImportClause:
                    return emitImportClause(node);
                case SyntaxKind.NamespaceImport:
                    return emitNamespaceImport(node);
                case SyntaxKind.NamespaceExport:
                    return emitNamespaceExport(node);
                case SyntaxKind.NamedImports:
                    return emitNamedImports(node);
                case SyntaxKind.ImportSpecifier:
                    return emitImportSpecifier(node);
                case SyntaxKind.ExportAssignment:
                    return emitExportAssignment(node);
                case SyntaxKind.ExportDeclaration:
                    return emitExportDeclaration(node);
                case SyntaxKind.NamedExports:
                    return emitNamedExports(node);
                case SyntaxKind.ExportSpecifier:
                    return emitExportSpecifier(node);
                case SyntaxKind.ImportAttributes:
                    return emitImportAttributes(node);
                case SyntaxKind.ImportAttribute:
                    return emitImportAttribute(node);
                case SyntaxKind.MissingDeclaration:
                    return;
                // Module references
                case SyntaxKind.ExternalModuleReference:
                    return emitExternalModuleReference(node);
                // JSX (non-expression)
                case SyntaxKind.JsxText:
                    return emitJsxText(node);
                case SyntaxKind.JsxOpeningElement:
                case SyntaxKind.JsxOpeningFragment:
                    return emitJsxOpeningElementOrFragment(node);
                case SyntaxKind.JsxClosingElement:
                case SyntaxKind.JsxClosingFragment:
                    return emitJsxClosingElementOrFragment(node);
                case SyntaxKind.JsxAttribute:
                    return emitJsxAttribute(node);
                case SyntaxKind.JsxAttributes:
                    return emitJsxAttributes(node);
                case SyntaxKind.JsxSpreadAttribute:
                    return emitJsxSpreadAttribute(node);
                case SyntaxKind.JsxExpression:
                    return emitJsxExpression(node);
                case SyntaxKind.JsxNamespacedName:
                    return emitJsxNamespacedName(node);
                // Clauses
                case SyntaxKind.CaseClause:
                    return emitCaseClause(node);
                case SyntaxKind.DefaultClause:
                    return emitDefaultClause(node);
                case SyntaxKind.HeritageClause:
                    return emitHeritageClause(node);
                case SyntaxKind.CatchClause:
                    return emitCatchClause(node);
                // Property assignments
                case SyntaxKind.PropertyAssignment:
                    return emitPropertyAssignment(node);
                case SyntaxKind.ShorthandPropertyAssignment:
                    return emitShorthandPropertyAssignment(node);
                case SyntaxKind.SpreadAssignment:
                    return emitSpreadAssignment(node);
                // Enum
                case SyntaxKind.EnumMember:
                    return emitEnumMember(node);
                // Top-level nodes
                case SyntaxKind.SourceFile:
                    return emitSourceFile(node);
                case SyntaxKind.Bundle:
                    return Debug.fail("Bundles should be printed using printBundle");
                // JSDoc nodes (only used in codefixes currently)
                case SyntaxKind.JSDocTypeExpression:
                    return emitJSDocTypeExpression(node);
                case SyntaxKind.JSDocNameReference:
                    return emitJSDocNameReference(node);
                case SyntaxKind.JSDocAllType:
                    return writePunctuation("*");
                case SyntaxKind.JSDocUnknownType:
                    return writePunctuation("?");
                case SyntaxKind.JSDocNullableType:
                    return emitJSDocNullableType(node);
                case SyntaxKind.JSDocNonNullableType:
                    return emitJSDocNonNullableType(node);
                case SyntaxKind.JSDocOptionalType:
                    return emitJSDocOptionalType(node);
                case SyntaxKind.JSDocFunctionType:
                    return emitJSDocFunctionType(node);
                case SyntaxKind.RestType:
                case SyntaxKind.JSDocVariadicType:
                    return emitRestOrJSDocVariadicType(node);
                case SyntaxKind.JSDocNamepathType:
                    return;
                case SyntaxKind.JSDoc:
                    return emitJSDoc(node);
                case SyntaxKind.JSDocTypeLiteral:
                    return emitJSDocTypeLiteral(node);
                case SyntaxKind.JSDocSignature:
                    return emitJSDocSignature(node);
                case SyntaxKind.JSDocTag:
                case SyntaxKind.JSDocClassTag:
                case SyntaxKind.JSDocOverrideTag:
                    return emitJSDocSimpleTag(node);
                case SyntaxKind.JSDocAugmentsTag:
                case SyntaxKind.JSDocImplementsTag:
                    return emitJSDocHeritageTag(node);
                case SyntaxKind.JSDocAuthorTag:
                case SyntaxKind.JSDocDeprecatedTag:
                    return;
                // SyntaxKind.JSDocClassTag (see JSDocTag, above)
                case SyntaxKind.JSDocPublicTag:
                case SyntaxKind.JSDocPrivateTag:
                case SyntaxKind.JSDocProtectedTag:
                case SyntaxKind.JSDocReadonlyTag:
                    return;
                case SyntaxKind.JSDocCallbackTag:
                    return emitJSDocCallbackTag(node);
                case SyntaxKind.JSDocOverloadTag:
                    return emitJSDocOverloadTag(node);
                // SyntaxKind.JSDocEnumTag (see below)
                case SyntaxKind.JSDocParameterTag:
                case SyntaxKind.JSDocPropertyTag:
                    return emitJSDocPropertyLikeTag(node);
                case SyntaxKind.JSDocEnumTag:
                case SyntaxKind.JSDocReturnTag:
                case SyntaxKind.JSDocThisTag:
                case SyntaxKind.JSDocTypeTag:
                case SyntaxKind.JSDocThrowsTag:
                case SyntaxKind.JSDocSatisfiesTag:
                    return emitJSDocSimpleTypedTag(node);
                case SyntaxKind.JSDocTemplateTag:
                    return emitJSDocTemplateTag(node);
                case SyntaxKind.JSDocTypedefTag:
                    return emitJSDocTypedefTag(node);
                case SyntaxKind.JSDocSeeTag:
                    return emitJSDocSeeTag(node);
                case SyntaxKind.JSDocImportTag:
                    return emitJSDocImportTag(node);
                // SyntaxKind.JSDocPropertyTag (see JSDocParameterTag, above)
                // Transformation nodes
                case SyntaxKind.NotEmittedStatement:
                case SyntaxKind.NotEmittedTypeElement:
                    return;
            }
            if (isExpression(node)) {
                hint = EmitHint.Expression;
                if (substituteNode !== noEmitSubstitution) {
                    const substitute = substituteNode(hint, node) || node;
                    if (substitute !== node) {
                        node = substitute;
                        if (currentParenthesizerRule) {
                            node = currentParenthesizerRule(node);
                        }
                    }
                }
            }
        }
        if (hint === EmitHint.Expression) {
            switch (node.kind) {
                // Literals
                case SyntaxKind.NumericLiteral:
                case SyntaxKind.BigIntLiteral:
                    return emitNumericOrBigIntLiteral(node);
                case SyntaxKind.StringLiteral:
                case SyntaxKind.RegularExpressionLiteral:
                case SyntaxKind.NoSubstitutionTemplateLiteral:
                    return emitLiteral(node, /*jsxAttributeEscape*/ false);
                // Identifiers
                case SyntaxKind.Identifier:
                    return emitIdentifier(node);
                case SyntaxKind.PrivateIdentifier:
                    return emitPrivateIdentifier(node);
                // Expressions
                case SyntaxKind.ArrayLiteralExpression:
                    return emitArrayLiteralExpression(node);
                case SyntaxKind.ObjectLiteralExpression:
                    return emitObjectLiteralExpression(node);
                case SyntaxKind.PropertyAccessExpression:
                    return emitPropertyAccessExpression(node);
                case SyntaxKind.ElementAccessExpression:
                    return emitElementAccessExpression(node);
                case SyntaxKind.CallExpression:
                    return emitCallExpression(node);
                case SyntaxKind.NewExpression:
                    return emitNewExpression(node);
                case SyntaxKind.TaggedTemplateExpression:
                    return emitTaggedTemplateExpression(node);
                case SyntaxKind.TypeAssertionExpression:
                    return emitTypeAssertionExpression(node);
                case SyntaxKind.ParenthesizedExpression:
                    return emitParenthesizedExpression(node);
                case SyntaxKind.FunctionExpression:
                    return emitFunctionExpression(node);
                case SyntaxKind.ArrowFunction:
                    return emitArrowFunction(node);
                case SyntaxKind.DeleteExpression:
                    return emitDeleteExpression(node);
                case SyntaxKind.TypeOfExpression:
                    return emitTypeOfExpression(node);
                case SyntaxKind.VoidExpression:
                    return emitVoidExpression(node);
                case SyntaxKind.AwaitExpression:
                    return emitAwaitExpression(node);
                case SyntaxKind.PrefixUnaryExpression:
                    return emitPrefixUnaryExpression(node);
                case SyntaxKind.PostfixUnaryExpression:
                    return emitPostfixUnaryExpression(node);
                case SyntaxKind.BinaryExpression:
                    return emitBinaryExpression(node);
                case SyntaxKind.ConditionalExpression:
                    return emitConditionalExpression(node);
                case SyntaxKind.TemplateExpression:
                    return emitTemplateExpression(node);
                case SyntaxKind.YieldExpression:
                    return emitYieldExpression(node);
                case SyntaxKind.SpreadElement:
                    return emitSpreadElement(node);
                case SyntaxKind.ClassExpression:
                    return emitClassExpression(node);
                case SyntaxKind.OmittedExpression:
                    return;
                case SyntaxKind.AsExpression:
                    return emitAsExpression(node);
                case SyntaxKind.NonNullExpression:
                    return emitNonNullExpression(node);
                case SyntaxKind.ExpressionWithTypeArguments:
                    return emitExpressionWithTypeArguments(node);
                case SyntaxKind.SatisfiesExpression:
                    return emitSatisfiesExpression(node);
                case SyntaxKind.MetaProperty:
                    return emitMetaProperty(node);
                case SyntaxKind.SyntheticExpression:
                    return Debug.fail("SyntheticExpression should never be printed.");
                case SyntaxKind.MissingDeclaration:
                    return;
                // JSX
                case SyntaxKind.JsxElement:
                    return emitJsxElement(node);
                case SyntaxKind.JsxSelfClosingElement:
                    return emitJsxSelfClosingElement(node);
                case SyntaxKind.JsxFragment:
                    return emitJsxFragment(node);
                // Synthesized list
                case SyntaxKind.SyntaxList:
                    return Debug.fail("SyntaxList should not be printed");
                // Transformation nodes
                case SyntaxKind.NotEmittedStatement:
                    return;
                case SyntaxKind.PartiallyEmittedExpression:
                    return emitPartiallyEmittedExpression(node);
                case SyntaxKind.CommaListExpression:
                    return emitCommaList(node);
                case SyntaxKind.SyntheticReferenceExpression:
                    return Debug.fail("SyntheticReferenceExpression should not be printed");
            }
        }
        if (isKeyword(node.kind))
            return writeTokenNode(node, writeKeyword);
        if (isTokenKind(node.kind))
            return writeTokenNode(node, writePunctuation);
        Debug.fail(`Unhandled SyntaxKind: ${Debug.formatSyntaxKind(node.kind)}.`);
    }
    function emitMappedTypeParameter(node) {
        emit(node.name);
        writeSpace();
        writeKeyword("in");
        writeSpace();
        emit(node.constraint);
    }
    function pipelineEmitWithSubstitution(hint, node) {
        const pipelinePhase = getNextPipelinePhase(PipelinePhase.Substitution, hint, node);
        Debug.assertIsDefined(lastSubstitution);
        node = lastSubstitution;
        lastSubstitution = undefined;
        pipelinePhase(hint, node);
    }
    function emitHelpers(node) {
        let helpersEmitted = false;
        const bundle = node.kind === SyntaxKind.Bundle ? node : undefined;
        if (bundle && moduleKind === ModuleKind.None) {
            return;
        }
        const numNodes = bundle ? bundle.sourceFiles.length : 1;
        for (let i = 0; i < numNodes; i++) {
            const currentNode = bundle ? bundle.sourceFiles[i] : node;
            const sourceFile = isSourceFile(currentNode) ? currentNode : currentSourceFile;
            const shouldSkip = printerOptions.noEmitHelpers || (!!sourceFile && hasRecordedExternalHelpers(sourceFile));
            const shouldBundle = isSourceFile(currentNode) && !isOwnFileEmit;
            const helpers = getSortedEmitHelpers(currentNode);
            if (helpers) {
                for (const helper of helpers) {
                    if (!helper.scoped) {
                        // Skip the helper if it can be skipped and the noEmitHelpers compiler
                        // option is set, or if it can be imported and the importHelpers compiler
                        // option is set.
                        if (shouldSkip)
                            continue;
                        // Skip the helper if it can be bundled but hasn't already been emitted and we
                        // are emitting a bundled module.
                        if (shouldBundle) {
                            if (bundledHelpers.get(helper.name)) {
                                continue;
                            }
                            bundledHelpers.set(helper.name, true);
                        }
                    }
                    else if (bundle) {
                        // Skip the helper if it is scoped and we are emitting bundled helpers
                        continue;
                    }
                    if (typeof helper.text === "string") {
                        writeLines(helper.text);
                    }
                    else {
                        writeLines(helper.text(makeFileLevelOptimisticUniqueName));
                    }
                    helpersEmitted = true;
                }
            }
        }
        return helpersEmitted;
    }
    function getSortedEmitHelpers(node) {
        const helpers = getEmitHelpers(node);
        return helpers && toSorted(helpers, compareEmitHelpers);
    }
    //
    // Literals/Pseudo-literals
    //
    // SyntaxKind.NumericLiteral
    // SyntaxKind.BigIntLiteral
    function emitNumericOrBigIntLiteral(node) {
        emitLiteral(node, /*jsxAttributeEscape*/ false);
    }
    // SyntaxKind.StringLiteral
    // SyntaxKind.RegularExpressionLiteral
    // SyntaxKind.NoSubstitutionTemplateLiteral
    // SyntaxKind.TemplateHead
    // SyntaxKind.TemplateMiddle
    // SyntaxKind.TemplateTail
    function emitLiteral(node, jsxAttributeEscape) {
        const text = getLiteralTextOfNode(node, /*sourceFile*/ undefined, printerOptions.neverAsciiEscape, jsxAttributeEscape);
        if ((printerOptions.sourceMap || printerOptions.inlineSourceMap)
            && (node.kind === SyntaxKind.StringLiteral || isTemplateLiteralKind(node.kind))) {
            writeLiteral(text);
        }
        else {
            // Quick info expects all literals to be called with writeStringLiteral, as there's no specific type for numberLiterals
            writeStringLiteral(text);
        }
    }
    //
    // Snippet Elements
    //
    function emitSnippetNode(hint, node, snippet) {
        switch (snippet.kind) {
            case SnippetKind.Placeholder:
                emitPlaceholder(hint, node, snippet);
                break;
            case SnippetKind.TabStop:
                emitTabStop(hint, node, snippet);
                break;
        }
    }
    function emitPlaceholder(hint, node, snippet) {
        nonEscapingWrite(`$\{${snippet.order}:`); // `${2:`
        pipelineEmitWithHintWorker(hint, node, /*allowSnippets*/ false); // `...`
        nonEscapingWrite(`}`); // `}`
        // `${2:...}`
    }
    function emitTabStop(hint, node, snippet) {
        // A tab stop should only be attached to an empty node, i.e. a node that doesn't emit any text.
        Debug.assert(node.kind === SyntaxKind.EmptyStatement, `A tab stop cannot be attached to a node of kind ${Debug.formatSyntaxKind(node.kind)}.`);
        Debug.assert(hint !== EmitHint.EmbeddedStatement, `A tab stop cannot be attached to an embedded statement.`);
        nonEscapingWrite(`$${snippet.order}`);
    }
    //
    // Identifiers
    //
    function emitIdentifier(node) {
        const writeText = node.symbol ? writeSymbol : write;
        writeText(getTextOfNode(node, /*includeTrivia*/ false), node.symbol);
        emitList(node, getIdentifierTypeArguments(node), ListFormat.TypeParameters); // Call emitList directly since it could be an array of TypeParameterDeclarations _or_ type arguments
    }
    //
    // Names
    //
    function emitPrivateIdentifier(node) {
        write(getTextOfNode(node, /*includeTrivia*/ false));
    }
    function emitQualifiedName(node) {
        emitEntityName(node.left);
        writePunctuation(".");
        emit(node.right);
    }
    function emitEntityName(node) {
        if (node.kind === SyntaxKind.Identifier) {
            emitExpression(node);
        }
        else {
            emit(node);
        }
    }
    function emitComputedPropertyName(node) {
        writePunctuation("[");
        emitExpression(node.expression, parenthesizer.parenthesizeExpressionOfComputedPropertyName);
        writePunctuation("]");
    }
    //
    // Signature elements
    //
    function emitTypeParameter(node) {
        emitModifierList(node, node.modifiers);
        emit(node.name);
        if (node.constraint) {
            writeSpace();
            writeKeyword("extends");
            writeSpace();
            emit(node.constraint);
        }
        if (node.default) {
            writeSpace();
            writeOperator("=");
            writeSpace();
            emit(node.default);
        }
    }
    function emitParameter(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ true);
        emit(node.dotDotDotToken);
        emitNodeWithWriter(node.name, writeParameter);
        emit(node.questionToken);
        if (node.parent && node.parent.kind === SyntaxKind.JSDocFunctionType && !node.name) {
            emit(node.type);
        }
        else {
            emitTypeAnnotation(node.type);
        }
        // The comment position has to fallback to any present node within the parameterdeclaration because as it turns out, the parser can make parameter declarations with _just_ an initializer.
        emitInitializer(node.initializer, node.type ? node.type.end : node.questionToken ? node.questionToken.end : node.name ? node.name.end : node.modifiers ? node.modifiers.end : node.pos, node, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    function emitDecorator(decorator) {
        writePunctuation("@");
        emitExpression(decorator.expression, parenthesizer.parenthesizeLeftSideOfAccess);
    }
    //
    // Type members
    //
    function emitPropertySignature(node) {
        emitModifierList(node, node.modifiers);
        emitNodeWithWriter(node.name, writeProperty);
        emit(node.questionToken);
        emitTypeAnnotation(node.type);
        writeTrailingSemicolon();
    }
    function emitPropertyDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ true);
        emit(node.name);
        emit(node.questionToken);
        emit(node.exclamationToken);
        emitTypeAnnotation(node.type);
        emitInitializer(node.initializer, node.type ? node.type.end : node.questionToken ? node.questionToken.end : node.name.end, node);
        writeTrailingSemicolon();
    }
    function emitMethodSignature(node) {
        emitModifierList(node, node.modifiers);
        emit(node.name);
        emit(node.questionToken);
        emitSignatureAndBody(node, emitSignatureHead, emitEmptyFunctionBody);
    }
    function emitMethodDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ true);
        emit(node.asteriskToken);
        emit(node.name);
        emit(node.questionToken);
        emitSignatureAndBody(node, emitSignatureHead, emitFunctionBody);
    }
    function emitClassStaticBlockDeclaration(node) {
        writeKeyword("static");
        pushNameGenerationScope(node);
        emitBlockFunctionBody(node.body);
        popNameGenerationScope(node);
    }
    function emitConstructor(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        writeKeyword("constructor");
        emitSignatureAndBody(node, emitSignatureHead, emitFunctionBody);
    }
    function emitAccessorDeclaration(node) {
        const pos = emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ true);
        const token = node.kind === SyntaxKind.GetAccessor ? SyntaxKind.GetKeyword : SyntaxKind.SetKeyword;
        emitTokenWithComment(token, pos, writeKeyword, node);
        writeSpace();
        emit(node.name);
        emitSignatureAndBody(node, emitSignatureHead, emitFunctionBody);
    }
    function emitCallSignature(node) {
        emitSignatureAndBody(node, emitSignatureHead, emitEmptyFunctionBody);
    }
    function emitConstructSignature(node) {
        writeKeyword("new");
        writeSpace();
        emitSignatureAndBody(node, emitSignatureHead, emitEmptyFunctionBody);
    }
    function emitIndexSignature(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        emitParametersForIndexSignature(node, node.parameters);
        emitTypeAnnotation(node.type);
        writeTrailingSemicolon();
    }
    function emitTemplateTypeSpan(node) {
        emit(node.type);
        emit(node.literal);
    }
    function emitSemicolonClassElement() {
        writeTrailingSemicolon();
    }
    //
    // Types
    //
    function emitTypePredicate(node) {
        if (node.assertsModifier) {
            emit(node.assertsModifier);
            writeSpace();
        }
        emit(node.parameterName);
        if (node.type) {
            writeSpace();
            writeKeyword("is");
            writeSpace();
            emit(node.type);
        }
    }
    function emitTypeReference(node) {
        emit(node.typeName);
        emitTypeArguments(node, node.typeArguments);
    }
    function emitFunctionType(node) {
        emitSignatureAndBody(node, emitFunctionTypeHead, emitFunctionTypeBody);
    }
    function emitFunctionTypeHead(node) {
        emitTypeParameters(node, node.typeParameters);
        emitParametersForArrow(node, node.parameters);
        writeSpace();
        writePunctuation("=>");
    }
    function emitFunctionTypeBody(node) {
        writeSpace();
        emit(node.type);
    }
    function emitJSDocFunctionType(node) {
        writeKeyword("function");
        emitParameters(node, node.parameters);
        writePunctuation(":");
        emit(node.type);
    }
    function emitJSDocNullableType(node) {
        writePunctuation("?");
        emit(node.type);
    }
    function emitJSDocNonNullableType(node) {
        writePunctuation("!");
        emit(node.type);
    }
    function emitJSDocOptionalType(node) {
        emit(node.type);
        writePunctuation("=");
    }
    function emitConstructorType(node) {
        emitModifierList(node, node.modifiers);
        writeKeyword("new");
        writeSpace();
        emitSignatureAndBody(node, emitFunctionTypeHead, emitFunctionTypeBody);
    }
    function emitTypeQuery(node) {
        writeKeyword("typeof");
        writeSpace();
        emit(node.exprName);
        emitTypeArguments(node, node.typeArguments);
    }
    function emitTypeLiteral(node) {
        pushNameGenerationScope(node);
        forEach(node.members, generateMemberNames);
        writePunctuation("{");
        const flags = getEmitFlags(node) & EmitFlags.SingleLine ? ListFormat.SingleLineTypeLiteralMembers : ListFormat.MultiLineTypeLiteralMembers;
        emitList(node, node.members, flags | ListFormat.NoSpaceIfEmpty);
        writePunctuation("}");
        popNameGenerationScope(node);
    }
    function emitArrayType(node) {
        emit(node.elementType, parenthesizer.parenthesizeNonArrayTypeOfPostfixType);
        writePunctuation("[");
        writePunctuation("]");
    }
    function emitRestOrJSDocVariadicType(node) {
        writePunctuation("...");
        emit(node.type);
    }
    function emitTupleType(node) {
        emitTokenWithComment(SyntaxKind.OpenBracketToken, node.pos, writePunctuation, node);
        const flags = getEmitFlags(node) & EmitFlags.SingleLine ? ListFormat.SingleLineTupleTypeElements : ListFormat.MultiLineTupleTypeElements;
        emitList(node, node.elements, flags | ListFormat.NoSpaceIfEmpty, parenthesizer.parenthesizeElementTypeOfTupleType);
        emitTokenWithComment(SyntaxKind.CloseBracketToken, node.elements.end, writePunctuation, node);
    }
    function emitNamedTupleMember(node) {
        emit(node.dotDotDotToken);
        emit(node.name);
        emit(node.questionToken);
        emitTokenWithComment(SyntaxKind.ColonToken, node.name.end, writePunctuation, node);
        writeSpace();
        emit(node.type);
    }
    function emitOptionalType(node) {
        emit(node.type, parenthesizer.parenthesizeTypeOfOptionalType);
        writePunctuation("?");
    }
    function emitUnionType(node) {
        emitList(node, node.types, ListFormat.UnionTypeConstituents, parenthesizer.parenthesizeConstituentTypeOfUnionType);
    }
    function emitIntersectionType(node) {
        emitList(node, node.types, ListFormat.IntersectionTypeConstituents, parenthesizer.parenthesizeConstituentTypeOfIntersectionType);
    }
    function emitConditionalType(node) {
        emit(node.checkType, parenthesizer.parenthesizeCheckTypeOfConditionalType);
        writeSpace();
        writeKeyword("extends");
        writeSpace();
        emit(node.extendsType, parenthesizer.parenthesizeExtendsTypeOfConditionalType);
        writeSpace();
        writePunctuation("?");
        writeSpace();
        emit(node.trueType);
        writeSpace();
        writePunctuation(":");
        writeSpace();
        emit(node.falseType);
    }
    function emitInferType(node) {
        writeKeyword("infer");
        writeSpace();
        emit(node.typeParameter);
    }
    function emitParenthesizedType(node) {
        writePunctuation("(");
        emit(node.type);
        writePunctuation(")");
    }
    function emitThisType() {
        writeKeyword("this");
    }
    function emitTypeOperator(node) {
        writeTokenText(node.operator, writeKeyword);
        writeSpace();
        const parenthesizerRule = node.operator === SyntaxKind.ReadonlyKeyword ?
            parenthesizer.parenthesizeOperandOfReadonlyTypeOperator :
            parenthesizer.parenthesizeOperandOfTypeOperator;
        emit(node.type, parenthesizerRule);
    }
    function emitIndexedAccessType(node) {
        emit(node.objectType, parenthesizer.parenthesizeNonArrayTypeOfPostfixType);
        writePunctuation("[");
        emit(node.indexType);
        writePunctuation("]");
    }
    function emitMappedType(node) {
        const emitFlags = getEmitFlags(node);
        writePunctuation("{");
        if (emitFlags & EmitFlags.SingleLine) {
            writeSpace();
        }
        else {
            writeLine();
            increaseIndent();
        }
        if (node.readonlyToken) {
            emit(node.readonlyToken);
            if (node.readonlyToken.kind !== SyntaxKind.ReadonlyKeyword) {
                writeKeyword("readonly");
            }
            writeSpace();
        }
        writePunctuation("[");
        pipelineEmit(EmitHint.MappedTypeParameter, node.typeParameter);
        if (node.nameType) {
            writeSpace();
            writeKeyword("as");
            writeSpace();
            emit(node.nameType);
        }
        writePunctuation("]");
        if (node.questionToken) {
            emit(node.questionToken);
            if (node.questionToken.kind !== SyntaxKind.QuestionToken) {
                writePunctuation("?");
            }
        }
        writePunctuation(":");
        writeSpace();
        emit(node.type);
        writeTrailingSemicolon();
        if (emitFlags & EmitFlags.SingleLine) {
            writeSpace();
        }
        else {
            writeLine();
            decreaseIndent();
        }
        emitList(node, node.members, ListFormat.PreserveLines);
        writePunctuation("}");
    }
    function emitLiteralType(node) {
        emitExpression(node.literal);
    }
    function emitTemplateType(node) {
        emit(node.head);
        emitList(node, node.templateSpans, ListFormat.TemplateExpressionSpans);
    }
    function emitImportTypeNode(node) {
        if (node.isTypeOf) {
            writeKeyword("typeof");
            writeSpace();
        }
        writeKeyword("import");
        writePunctuation("(");
        emit(node.argument);
        if (node.attributes) {
            writePunctuation(",");
            writeSpace();
            pipelineEmit(EmitHint.ImportTypeNodeAttributes, node.attributes);
        }
        writePunctuation(")");
        if (node.qualifier) {
            writePunctuation(".");
            emit(node.qualifier);
        }
        emitTypeArguments(node, node.typeArguments);
    }
    //
    // Binding patterns
    //
    function emitObjectBindingPattern(node) {
        writePunctuation("{");
        emitList(node, node.elements, ListFormat.ObjectBindingPatternElements);
        writePunctuation("}");
    }
    function emitArrayBindingPattern(node) {
        writePunctuation("[");
        emitList(node, node.elements, ListFormat.ArrayBindingPatternElements);
        writePunctuation("]");
    }
    function emitBindingElement(node) {
        emit(node.dotDotDotToken);
        if (node.propertyName) {
            emit(node.propertyName);
            writePunctuation(":");
            writeSpace();
        }
        emit(node.name);
        emitInitializer(node.initializer, node.name.end, node, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    //
    // Expressions
    //
    function emitArrayLiteralExpression(node) {
        const elements = node.elements;
        const preferNewLine = node.multiLine ? ListFormat.PreferNewLine : ListFormat.None;
        emitExpressionList(node, elements, ListFormat.ArrayLiteralExpressionElements | preferNewLine, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    function emitObjectLiteralExpression(node) {
        pushNameGenerationScope(node);
        forEach(node.properties, generateMemberNames);
        const indentedFlag = getEmitFlags(node) & EmitFlags.Indented;
        if (indentedFlag) {
            increaseIndent();
        }
        const preferNewLine = node.multiLine ? ListFormat.PreferNewLine : ListFormat.None;
        const allowTrailingComma = currentSourceFile && !isJsonSourceFile(currentSourceFile) ? ListFormat.AllowTrailingComma : ListFormat.None;
        emitList(node, node.properties, ListFormat.ObjectLiteralExpressionProperties | allowTrailingComma | preferNewLine);
        if (indentedFlag) {
            decreaseIndent();
        }
        popNameGenerationScope(node);
    }
    function emitPropertyAccessExpression(node) {
        emitExpression(node.expression, parenthesizer.parenthesizeLeftSideOfAccess);
        const token = node.questionDotToken || setTextRangePosEnd(factory.createToken(SyntaxKind.DotToken), node.expression.end, node.name.pos);
        const linesBeforeDot = getLinesBetweenNodes(node, node.expression, token);
        const linesAfterDot = getLinesBetweenNodes(node, token, node.name);
        writeLinesAndIndent(linesBeforeDot, /*writeSpaceIfNotIndenting*/ false);
        const shouldEmitDotDot = token.kind !== SyntaxKind.QuestionDotToken &&
            mayNeedDotDotForPropertyAccess(node.expression) &&
            !writer.hasTrailingComment() &&
            !writer.hasTrailingWhitespace();
        if (shouldEmitDotDot) {
            writePunctuation(".");
        }
        if (node.questionDotToken) {
            emit(token);
        }
        else {
            emitTokenWithComment(token.kind, node.expression.end, writePunctuation, node);
        }
        writeLinesAndIndent(linesAfterDot, /*writeSpaceIfNotIndenting*/ false);
        emit(node.name);
        decreaseIndentIf(linesBeforeDot, linesAfterDot);
    }
    // 1..toString is a valid property access, emit a dot after the literal
    // Also emit a dot if expression is a integer const enum value - it will appear in generated code as numeric literal
    function mayNeedDotDotForPropertyAccess(expression) {
        expression = skipPartiallyEmittedExpressions(expression);
        if (isNumericLiteral(expression)) {
            // check if numeric literal is a decimal literal that was originally written with a dot
            const text = getLiteralTextOfNode(expression, /*sourceFile*/ undefined, /*neverAsciiEscape*/ true, /*jsxAttributeEscape*/ false);
            // If the number will be printed verbatim and it doesn't already contain a dot or an exponent indicator, add one
            // if the expression doesn't have any comments that will be emitted.
            return !(expression.numericLiteralFlags & TokenFlags.WithSpecifier)
                && !text.includes(tokenToString(SyntaxKind.DotToken))
                && !text.includes(String.fromCharCode(CharacterCodes.E))
                && !text.includes(String.fromCharCode(CharacterCodes.e));
        }
        else if (isAccessExpression(expression)) {
            // check if constant enum value is a non-negative integer
            const constantValue = getConstantValue(expression);
            // isFinite handles cases when constantValue is undefined
            return typeof constantValue === "number" && isFinite(constantValue)
                && constantValue >= 0 && Math.floor(constantValue) === constantValue;
        }
    }
    function emitElementAccessExpression(node) {
        emitExpression(node.expression, parenthesizer.parenthesizeLeftSideOfAccess);
        emit(node.questionDotToken);
        emitTokenWithComment(SyntaxKind.OpenBracketToken, node.expression.end, writePunctuation, node);
        emitExpression(node.argumentExpression);
        emitTokenWithComment(SyntaxKind.CloseBracketToken, node.argumentExpression.end, writePunctuation, node);
    }
    function emitCallExpression(node) {
        const indirectCall = getInternalEmitFlags(node) & InternalEmitFlags.IndirectCall;
        if (indirectCall) {
            writePunctuation("(");
            writeLiteral("0");
            writePunctuation(",");
            writeSpace();
        }
        emitExpression(node.expression, parenthesizer.parenthesizeLeftSideOfAccess);
        if (indirectCall) {
            writePunctuation(")");
        }
        emit(node.questionDotToken);
        emitTypeArguments(node, node.typeArguments);
        emitExpressionList(node, node.arguments, ListFormat.CallExpressionArguments, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    function emitNewExpression(node) {
        emitTokenWithComment(SyntaxKind.NewKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitExpression(node.expression, parenthesizer.parenthesizeExpressionOfNew);
        emitTypeArguments(node, node.typeArguments);
        emitExpressionList(node, node.arguments, ListFormat.NewExpressionArguments, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    function emitTaggedTemplateExpression(node) {
        const indirectCall = getInternalEmitFlags(node) & InternalEmitFlags.IndirectCall;
        if (indirectCall) {
            writePunctuation("(");
            writeLiteral("0");
            writePunctuation(",");
            writeSpace();
        }
        emitExpression(node.tag, parenthesizer.parenthesizeLeftSideOfAccess);
        if (indirectCall) {
            writePunctuation(")");
        }
        emitTypeArguments(node, node.typeArguments);
        writeSpace();
        emitExpression(node.template);
    }
    function emitTypeAssertionExpression(node) {
        writePunctuation("<");
        emit(node.type);
        writePunctuation(">");
        emitExpression(node.expression, parenthesizer.parenthesizeOperandOfPrefixUnary);
    }
    function emitParenthesizedExpression(node) {
        const openParenPos = emitTokenWithComment(SyntaxKind.OpenParenToken, node.pos, writePunctuation, node);
        const indented = writeLineSeparatorsAndIndentBefore(node.expression, node);
        emitExpression(node.expression, /*parenthesizerRule*/ undefined);
        writeLineSeparatorsAfter(node.expression, node);
        decreaseIndentIf(indented);
        emitTokenWithComment(SyntaxKind.CloseParenToken, node.expression ? node.expression.end : openParenPos, writePunctuation, node);
    }
    function emitFunctionExpression(node) {
        generateNameIfNeeded(node.name);
        emitFunctionDeclarationOrExpression(node);
    }
    function emitArrowFunction(node) {
        emitModifierList(node, node.modifiers);
        emitSignatureAndBody(node, emitArrowFunctionHead, emitArrowFunctionBody);
    }
    function emitArrowFunctionHead(node) {
        emitTypeParameters(node, node.typeParameters);
        emitParametersForArrow(node, node.parameters);
        emitTypeAnnotation(node.type);
        writeSpace();
        emit(node.equalsGreaterThanToken);
    }
    function emitArrowFunctionBody(node) {
        if (isBlock(node.body)) {
            emitBlockFunctionBody(node.body);
        }
        else {
            writeSpace();
            emitExpression(node.body, parenthesizer.parenthesizeConciseBodyOfArrowFunction);
        }
    }
    function emitDeleteExpression(node) {
        emitTokenWithComment(SyntaxKind.DeleteKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitExpression(node.expression, parenthesizer.parenthesizeOperandOfPrefixUnary);
    }
    function emitTypeOfExpression(node) {
        emitTokenWithComment(SyntaxKind.TypeOfKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitExpression(node.expression, parenthesizer.parenthesizeOperandOfPrefixUnary);
    }
    function emitVoidExpression(node) {
        emitTokenWithComment(SyntaxKind.VoidKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitExpression(node.expression, parenthesizer.parenthesizeOperandOfPrefixUnary);
    }
    function emitAwaitExpression(node) {
        emitTokenWithComment(SyntaxKind.AwaitKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitExpression(node.expression, parenthesizer.parenthesizeOperandOfPrefixUnary);
    }
    function emitPrefixUnaryExpression(node) {
        writeTokenText(node.operator, writeOperator);
        if (shouldEmitWhitespaceBeforeOperand(node)) {
            writeSpace();
        }
        emitExpression(node.operand, parenthesizer.parenthesizeOperandOfPrefixUnary);
    }
    function shouldEmitWhitespaceBeforeOperand(node) {
        // In some cases, we need to emit a space between the operator and the operand. One obvious case
        // is when the operator is an identifier, like delete or typeof. We also need to do this for plus
        // and minus expressions in certain cases. Specifically, consider the following two cases (parens
        // are just for clarity of exposition, and not part of the source code):
        //
        //  (+(+1))
        //  (+(++1))
        //
        // We need to emit a space in both cases. In the first case, the absence of a space will make
        // the resulting expression a prefix increment operation. And in the second, it will make the resulting
        // expression a prefix increment whose operand is a plus expression - (++(+x))
        // The same is true of minus of course.
        const operand = node.operand;
        return operand.kind === SyntaxKind.PrefixUnaryExpression
            && ((node.operator === SyntaxKind.PlusToken && (operand.operator === SyntaxKind.PlusToken || operand.operator === SyntaxKind.PlusPlusToken))
                || (node.operator === SyntaxKind.MinusToken && (operand.operator === SyntaxKind.MinusToken || operand.operator === SyntaxKind.MinusMinusToken)));
    }
    function emitPostfixUnaryExpression(node) {
        emitExpression(node.operand, parenthesizer.parenthesizeOperandOfPostfixUnary);
        writeTokenText(node.operator, writeOperator);
    }
    function createEmitBinaryExpression() {
        return createBinaryExpressionTrampoline(onEnter, onLeft, onOperator, onRight, onExit, /*foldState*/ undefined);
        function onEnter(node, state) {
            if (state) {
                state.stackIndex++;
                state.preserveSourceNewlinesStack[state.stackIndex] = preserveSourceNewlines;
                state.containerPosStack[state.stackIndex] = containerPos;
                state.containerEndStack[state.stackIndex] = containerEnd;
                state.declarationListContainerEndStack[state.stackIndex] = declarationListContainerEnd;
                const emitComments = state.shouldEmitCommentsStack[state.stackIndex] = shouldEmitComments(node);
                const emitSourceMaps = state.shouldEmitSourceMapsStack[state.stackIndex] = shouldEmitSourceMaps(node);
                onBeforeEmitNode?.(node);
                if (emitComments)
                    emitCommentsBeforeNode(node);
                if (emitSourceMaps)
                    emitSourceMapsBeforeNode(node);
                beforeEmitNode(node);
            }
            else {
                state = {
                    stackIndex: 0,
                    preserveSourceNewlinesStack: [undefined],
                    containerPosStack: [-1],
                    containerEndStack: [-1],
                    declarationListContainerEndStack: [-1],
                    shouldEmitCommentsStack: [false],
                    shouldEmitSourceMapsStack: [false],
                };
            }
            return state;
        }
        function onLeft(next, _workArea, parent) {
            return maybeEmitExpression(next, parent, "left");
        }
        function onOperator(operatorToken, _state, node) {
            const isCommaOperator = operatorToken.kind !== SyntaxKind.CommaToken;
            const linesBeforeOperator = getLinesBetweenNodes(node, node.left, operatorToken);
            const linesAfterOperator = getLinesBetweenNodes(node, operatorToken, node.right);
            writeLinesAndIndent(linesBeforeOperator, isCommaOperator);
            emitLeadingCommentsOfPosition(operatorToken.pos);
            writeTokenNode(operatorToken, operatorToken.kind === SyntaxKind.InKeyword ? writeKeyword : writeOperator);
            emitTrailingCommentsOfPosition(operatorToken.end, /*prefixSpace*/ true); // Binary operators should have a space before the comment starts
            writeLinesAndIndent(linesAfterOperator, /*writeSpaceIfNotIndenting*/ true);
        }
        function onRight(next, _workArea, parent) {
            return maybeEmitExpression(next, parent, "right");
        }
        function onExit(node, state) {
            const linesBeforeOperator = getLinesBetweenNodes(node, node.left, node.operatorToken);
            const linesAfterOperator = getLinesBetweenNodes(node, node.operatorToken, node.right);
            decreaseIndentIf(linesBeforeOperator, linesAfterOperator);
            if (state.stackIndex > 0) {
                const savedPreserveSourceNewlines = state.preserveSourceNewlinesStack[state.stackIndex];
                const savedContainerPos = state.containerPosStack[state.stackIndex];
                const savedContainerEnd = state.containerEndStack[state.stackIndex];
                const savedDeclarationListContainerEnd = state.declarationListContainerEndStack[state.stackIndex];
                const shouldEmitComments = state.shouldEmitCommentsStack[state.stackIndex];
                const shouldEmitSourceMaps = state.shouldEmitSourceMapsStack[state.stackIndex];
                afterEmitNode(savedPreserveSourceNewlines);
                if (shouldEmitSourceMaps)
                    emitSourceMapsAfterNode(node);
                if (shouldEmitComments)
                    emitCommentsAfterNode(node, savedContainerPos, savedContainerEnd, savedDeclarationListContainerEnd);
                onAfterEmitNode?.(node);
                state.stackIndex--;
            }
        }
        function maybeEmitExpression(next, parent, side) {
            const parenthesizerRule = side === "left" ?
                parenthesizer.getParenthesizeLeftSideOfBinaryForOperator(parent.operatorToken.kind) :
                parenthesizer.getParenthesizeRightSideOfBinaryForOperator(parent.operatorToken.kind);
            let pipelinePhase = getPipelinePhase(PipelinePhase.Notification, EmitHint.Expression, next);
            if (pipelinePhase === pipelineEmitWithSubstitution) {
                Debug.assertIsDefined(lastSubstitution);
                next = parenthesizerRule(cast(lastSubstitution, isExpression));
                pipelinePhase = getNextPipelinePhase(PipelinePhase.Substitution, EmitHint.Expression, next);
                lastSubstitution = undefined;
            }
            if (pipelinePhase === pipelineEmitWithComments ||
                pipelinePhase === pipelineEmitWithSourceMaps ||
                pipelinePhase === pipelineEmitWithHint) {
                if (isBinaryExpression(next)) {
                    return next;
                }
            }
            currentParenthesizerRule = parenthesizerRule;
            pipelinePhase(EmitHint.Expression, next);
        }
    }
    function emitConditionalExpression(node) {
        const linesBeforeQuestion = getLinesBetweenNodes(node, node.condition, node.questionToken);
        const linesAfterQuestion = getLinesBetweenNodes(node, node.questionToken, node.whenTrue);
        const linesBeforeColon = getLinesBetweenNodes(node, node.whenTrue, node.colonToken);
        const linesAfterColon = getLinesBetweenNodes(node, node.colonToken, node.whenFalse);
        emitExpression(node.condition, parenthesizer.parenthesizeConditionOfConditionalExpression);
        writeLinesAndIndent(linesBeforeQuestion, /*writeSpaceIfNotIndenting*/ true);
        emit(node.questionToken);
        writeLinesAndIndent(linesAfterQuestion, /*writeSpaceIfNotIndenting*/ true);
        emitExpression(node.whenTrue, parenthesizer.parenthesizeBranchOfConditionalExpression);
        decreaseIndentIf(linesBeforeQuestion, linesAfterQuestion);
        writeLinesAndIndent(linesBeforeColon, /*writeSpaceIfNotIndenting*/ true);
        emit(node.colonToken);
        writeLinesAndIndent(linesAfterColon, /*writeSpaceIfNotIndenting*/ true);
        emitExpression(node.whenFalse, parenthesizer.parenthesizeBranchOfConditionalExpression);
        decreaseIndentIf(linesBeforeColon, linesAfterColon);
    }
    function emitTemplateExpression(node) {
        emit(node.head);
        emitList(node, node.templateSpans, ListFormat.TemplateExpressionSpans);
    }
    function emitYieldExpression(node) {
        emitTokenWithComment(SyntaxKind.YieldKeyword, node.pos, writeKeyword, node);
        emit(node.asteriskToken);
        emitExpressionWithLeadingSpace(node.expression && parenthesizeExpressionForNoAsi(node.expression), parenthesizeExpressionForNoAsiAndDisallowedComma);
    }
    function emitSpreadElement(node) {
        emitTokenWithComment(SyntaxKind.DotDotDotToken, node.pos, writePunctuation, node);
        emitExpression(node.expression, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    function emitClassExpression(node) {
        generateNameIfNeeded(node.name);
        emitClassDeclarationOrExpression(node);
    }
    function emitExpressionWithTypeArguments(node) {
        emitExpression(node.expression, parenthesizer.parenthesizeLeftSideOfAccess);
        emitTypeArguments(node, node.typeArguments);
    }
    function emitAsExpression(node) {
        emitExpression(node.expression, /*parenthesizerRule*/ undefined);
        if (node.type) {
            writeSpace();
            writeKeyword("as");
            writeSpace();
            emit(node.type);
        }
    }
    function emitNonNullExpression(node) {
        emitExpression(node.expression, parenthesizer.parenthesizeLeftSideOfAccess);
        writeOperator("!");
    }
    function emitSatisfiesExpression(node) {
        emitExpression(node.expression, /*parenthesizerRule*/ undefined);
        if (node.type) {
            writeSpace();
            writeKeyword("satisfies");
            writeSpace();
            emit(node.type);
        }
    }
    function emitMetaProperty(node) {
        writeToken(node.keywordToken, node.pos, writePunctuation);
        writePunctuation(".");
        emit(node.name);
    }
    //
    // Misc
    //
    function emitTemplateSpan(node) {
        emitExpression(node.expression);
        emit(node.literal);
    }
    //
    // Statements
    //
    function emitBlock(node) {
        emitBlockStatements(node, /*forceSingleLine*/ !node.multiLine && isEmptyBlock(node));
    }
    function emitBlockStatements(node, forceSingleLine) {
        emitTokenWithComment(SyntaxKind.OpenBraceToken, node.pos, writePunctuation, /*contextNode*/ node);
        const format = forceSingleLine || getEmitFlags(node) & EmitFlags.SingleLine ? ListFormat.SingleLineBlockStatements : ListFormat.MultiLineBlockStatements;
        emitList(node, node.statements, format);
        emitTokenWithComment(SyntaxKind.CloseBraceToken, node.statements.end, writePunctuation, /*contextNode*/ node, /*indentLeading*/ !!(format & ListFormat.MultiLine));
    }
    function emitVariableStatement(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        emit(node.declarationList);
        writeTrailingSemicolon();
    }
    function emitEmptyStatement(isEmbeddedStatement) {
        // While most trailing semicolons are possibly insignificant, an embedded "empty"
        // statement is significant and cannot be elided by a trailing-semicolon-omitting writer.
        if (isEmbeddedStatement) {
            writePunctuation(";");
        }
        else {
            writeTrailingSemicolon();
        }
    }
    function emitExpressionStatement(node) {
        emitExpression(node.expression, parenthesizer.parenthesizeExpressionOfExpressionStatement);
        // Emit semicolon in non json files
        // or if json file that created synthesized expression(eg.define expression statement when --out and amd code generation)
        if (!currentSourceFile || !isJsonSourceFile(currentSourceFile) || nodeIsSynthesized(node.expression)) {
            writeTrailingSemicolon();
        }
    }
    function emitIfStatement(node) {
        const openParenPos = emitTokenWithComment(SyntaxKind.IfKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitTokenWithComment(SyntaxKind.OpenParenToken, openParenPos, writePunctuation, node);
        emitExpression(node.expression);
        emitTokenWithComment(SyntaxKind.CloseParenToken, node.expression.end, writePunctuation, node);
        emitEmbeddedStatement(node, node.thenStatement);
        if (node.elseStatement) {
            writeLineOrSpace(node, node.thenStatement, node.elseStatement);
            emitTokenWithComment(SyntaxKind.ElseKeyword, node.thenStatement.end, writeKeyword, node);
            if (node.elseStatement.kind === SyntaxKind.IfStatement) {
                writeSpace();
                emit(node.elseStatement);
            }
            else {
                emitEmbeddedStatement(node, node.elseStatement);
            }
        }
    }
    function emitWhileClause(node, startPos) {
        const openParenPos = emitTokenWithComment(SyntaxKind.WhileKeyword, startPos, writeKeyword, node);
        writeSpace();
        emitTokenWithComment(SyntaxKind.OpenParenToken, openParenPos, writePunctuation, node);
        emitExpression(node.expression);
        emitTokenWithComment(SyntaxKind.CloseParenToken, node.expression.end, writePunctuation, node);
    }
    function emitDoStatement(node) {
        emitTokenWithComment(SyntaxKind.DoKeyword, node.pos, writeKeyword, node);
        emitEmbeddedStatement(node, node.statement);
        if (isBlock(node.statement) && !preserveSourceNewlines) {
            writeSpace();
        }
        else {
            writeLineOrSpace(node, node.statement, node.expression);
        }
        emitWhileClause(node, node.statement.end);
        writeTrailingSemicolon();
    }
    function emitWhileStatement(node) {
        emitWhileClause(node, node.pos);
        emitEmbeddedStatement(node, node.statement);
    }
    function emitForStatement(node) {
        const openParenPos = emitTokenWithComment(SyntaxKind.ForKeyword, node.pos, writeKeyword, node);
        writeSpace();
        let pos = emitTokenWithComment(SyntaxKind.OpenParenToken, openParenPos, writePunctuation, /*contextNode*/ node);
        emitForBinding(node.initializer);
        pos = emitTokenWithComment(SyntaxKind.SemicolonToken, node.initializer ? node.initializer.end : pos, writePunctuation, node);
        emitExpressionWithLeadingSpace(node.condition);
        pos = emitTokenWithComment(SyntaxKind.SemicolonToken, node.condition ? node.condition.end : pos, writePunctuation, node);
        emitExpressionWithLeadingSpace(node.incrementor);
        emitTokenWithComment(SyntaxKind.CloseParenToken, node.incrementor ? node.incrementor.end : pos, writePunctuation, node);
        emitEmbeddedStatement(node, node.statement);
    }
    function emitForInStatement(node) {
        const openParenPos = emitTokenWithComment(SyntaxKind.ForKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitTokenWithComment(SyntaxKind.OpenParenToken, openParenPos, writePunctuation, node);
        emitForBinding(node.initializer);
        writeSpace();
        emitTokenWithComment(SyntaxKind.InKeyword, node.initializer.end, writeKeyword, node);
        writeSpace();
        emitExpression(node.expression);
        emitTokenWithComment(SyntaxKind.CloseParenToken, node.expression.end, writePunctuation, node);
        emitEmbeddedStatement(node, node.statement);
    }
    function emitForOfStatement(node) {
        const openParenPos = emitTokenWithComment(SyntaxKind.ForKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitWithTrailingSpace(node.awaitModifier);
        emitTokenWithComment(SyntaxKind.OpenParenToken, openParenPos, writePunctuation, node);
        emitForBinding(node.initializer);
        writeSpace();
        emitTokenWithComment(SyntaxKind.OfKeyword, node.initializer.end, writeKeyword, node);
        writeSpace();
        emitExpression(node.expression);
        emitTokenWithComment(SyntaxKind.CloseParenToken, node.expression.end, writePunctuation, node);
        emitEmbeddedStatement(node, node.statement);
    }
    function emitForBinding(node) {
        if (node !== undefined) {
            if (node.kind === SyntaxKind.VariableDeclarationList) {
                emit(node);
            }
            else {
                emitExpression(node);
            }
        }
    }
    function emitContinueStatement(node) {
        emitTokenWithComment(SyntaxKind.ContinueKeyword, node.pos, writeKeyword, node);
        emitWithLeadingSpace(node.label);
        writeTrailingSemicolon();
    }
    function emitBreakStatement(node) {
        emitTokenWithComment(SyntaxKind.BreakKeyword, node.pos, writeKeyword, node);
        emitWithLeadingSpace(node.label);
        writeTrailingSemicolon();
    }
    function emitTokenWithComment(token, pos, writer, contextNode, indentLeading) {
        const node = getParseTreeNode(contextNode);
        const isSimilarNode = node && node.kind === contextNode.kind;
        const startPos = pos;
        if (isSimilarNode && currentSourceFile) {
            pos = skipTrivia(currentSourceFile.text, pos);
        }
        if (isSimilarNode && contextNode.pos !== startPos) {
            const needsIndent = indentLeading && currentSourceFile && !positionsAreOnSameLine(startPos, pos, currentSourceFile);
            if (needsIndent) {
                increaseIndent();
            }
            emitLeadingCommentsOfPosition(startPos);
            if (needsIndent) {
                decreaseIndent();
            }
        }
        // We don't emit source positions for most tokens as it tends to be quite noisy, however
        // we need to emit source positions for open and close braces so that tools like istanbul
        // can map branches for code coverage. However, we still omit brace source positions when
        // the output is a declaration file.
        if (!omitBraceSourcePositions && (token === SyntaxKind.OpenBraceToken || token === SyntaxKind.CloseBraceToken)) {
            pos = writeToken(token, pos, writer, contextNode);
        }
        else {
            pos = writeTokenText(token, writer, pos);
        }
        if (isSimilarNode && contextNode.end !== pos) {
            const isJsxExprContext = contextNode.kind === SyntaxKind.JsxExpression;
            emitTrailingCommentsOfPosition(pos, /*prefixSpace*/ !isJsxExprContext, /*forceNoNewline*/ isJsxExprContext);
        }
        return pos;
    }
    function commentWillEmitNewLine(node) {
        return node.kind === SyntaxKind.SingleLineCommentTrivia || !!node.hasTrailingNewLine;
    }
    function willEmitLeadingNewLine(node) {
        if (!currentSourceFile)
            return false;
        const leadingCommentRanges = getLeadingCommentRanges(currentSourceFile.text, node.pos);
        if (leadingCommentRanges) {
            const parseNode = getParseTreeNode(node);
            if (parseNode && isParenthesizedExpression(parseNode.parent)) {
                return true;
            }
        }
        if (some(leadingCommentRanges, commentWillEmitNewLine))
            return true;
        if (some(getSyntheticLeadingComments(node), commentWillEmitNewLine))
            return true;
        if (isPartiallyEmittedExpression(node)) {
            if (node.pos !== node.expression.pos) {
                if (some(getTrailingCommentRanges(currentSourceFile.text, node.expression.pos), commentWillEmitNewLine))
                    return true;
            }
            return willEmitLeadingNewLine(node.expression);
        }
        return false;
    }
    /**
     * Wraps an expression in parens if we would emit a leading comment that would introduce a line separator
     * between the node and its parent.
     */
    function parenthesizeExpressionForNoAsi(node) {
        if (!commentsDisabled) {
            switch (node.kind) {
                case SyntaxKind.PartiallyEmittedExpression:
                    if (willEmitLeadingNewLine(node)) {
                        const parseNode = getParseTreeNode(node);
                        if (parseNode && isParenthesizedExpression(parseNode)) {
                            // If the original node was a parenthesized expression, restore it to preserve comment and source map emit
                            const parens = factory.createParenthesizedExpression(node.expression);
                            setOriginalNode(parens, node);
                            setTextRange(parens, parseNode);
                            return parens;
                        }
                        return factory.createParenthesizedExpression(node);
                    }
                    return factory.updatePartiallyEmittedExpression(node, parenthesizeExpressionForNoAsi(node.expression));
                case SyntaxKind.PropertyAccessExpression:
                    return factory.updatePropertyAccessExpression(node, parenthesizeExpressionForNoAsi(node.expression), node.name);
                case SyntaxKind.ElementAccessExpression:
                    return factory.updateElementAccessExpression(node, parenthesizeExpressionForNoAsi(node.expression), node.argumentExpression);
                case SyntaxKind.CallExpression:
                    return factory.updateCallExpression(node, parenthesizeExpressionForNoAsi(node.expression), node.typeArguments, node.arguments);
                case SyntaxKind.TaggedTemplateExpression:
                    return factory.updateTaggedTemplateExpression(node, parenthesizeExpressionForNoAsi(node.tag), node.typeArguments, node.template);
                case SyntaxKind.PostfixUnaryExpression:
                    return factory.updatePostfixUnaryExpression(node, parenthesizeExpressionForNoAsi(node.operand));
                case SyntaxKind.BinaryExpression:
                    return factory.updateBinaryExpression(node, parenthesizeExpressionForNoAsi(node.left), node.operatorToken, node.right);
                case SyntaxKind.ConditionalExpression:
                    return factory.updateConditionalExpression(node, parenthesizeExpressionForNoAsi(node.condition), node.questionToken, node.whenTrue, node.colonToken, node.whenFalse);
                case SyntaxKind.AsExpression:
                    return factory.updateAsExpression(node, parenthesizeExpressionForNoAsi(node.expression), node.type);
                case SyntaxKind.SatisfiesExpression:
                    return factory.updateSatisfiesExpression(node, parenthesizeExpressionForNoAsi(node.expression), node.type);
                case SyntaxKind.NonNullExpression:
                    return factory.updateNonNullExpression(node, parenthesizeExpressionForNoAsi(node.expression));
            }
        }
        return node;
    }
    function parenthesizeExpressionForNoAsiAndDisallowedComma(node) {
        return parenthesizeExpressionForNoAsi(parenthesizer.parenthesizeExpressionForDisallowedComma(node));
    }
    function emitReturnStatement(node) {
        emitTokenWithComment(SyntaxKind.ReturnKeyword, node.pos, writeKeyword, /*contextNode*/ node);
        emitExpressionWithLeadingSpace(node.expression && parenthesizeExpressionForNoAsi(node.expression), parenthesizeExpressionForNoAsi);
        writeTrailingSemicolon();
    }
    function emitWithStatement(node) {
        const openParenPos = emitTokenWithComment(SyntaxKind.WithKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitTokenWithComment(SyntaxKind.OpenParenToken, openParenPos, writePunctuation, node);
        emitExpression(node.expression);
        emitTokenWithComment(SyntaxKind.CloseParenToken, node.expression.end, writePunctuation, node);
        emitEmbeddedStatement(node, node.statement);
    }
    function emitSwitchStatement(node) {
        const openParenPos = emitTokenWithComment(SyntaxKind.SwitchKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitTokenWithComment(SyntaxKind.OpenParenToken, openParenPos, writePunctuation, node);
        emitExpression(node.expression);
        emitTokenWithComment(SyntaxKind.CloseParenToken, node.expression.end, writePunctuation, node);
        writeSpace();
        emit(node.caseBlock);
    }
    function emitLabeledStatement(node) {
        emit(node.label);
        emitTokenWithComment(SyntaxKind.ColonToken, node.label.end, writePunctuation, node);
        writeSpace();
        emit(node.statement);
    }
    function emitThrowStatement(node) {
        emitTokenWithComment(SyntaxKind.ThrowKeyword, node.pos, writeKeyword, node);
        emitExpressionWithLeadingSpace(parenthesizeExpressionForNoAsi(node.expression), parenthesizeExpressionForNoAsi);
        writeTrailingSemicolon();
    }
    function emitTryStatement(node) {
        emitTokenWithComment(SyntaxKind.TryKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emit(node.tryBlock);
        if (node.catchClause) {
            writeLineOrSpace(node, node.tryBlock, node.catchClause);
            emit(node.catchClause);
        }
        if (node.finallyBlock) {
            writeLineOrSpace(node, node.catchClause || node.tryBlock, node.finallyBlock);
            emitTokenWithComment(SyntaxKind.FinallyKeyword, (node.catchClause || node.tryBlock).end, writeKeyword, node);
            writeSpace();
            emit(node.finallyBlock);
        }
    }
    function emitDebuggerStatement(node) {
        writeToken(SyntaxKind.DebuggerKeyword, node.pos, writeKeyword);
        writeTrailingSemicolon();
    }
    //
    // Declarations
    //
    function emitVariableDeclaration(node) {
        emit(node.name);
        emit(node.exclamationToken);
        emitTypeAnnotation(node.type);
        emitInitializer(node.initializer, node.type?.end ?? node.name.emitNode?.typeNode?.end ?? node.name.end, node, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    function emitVariableDeclarationList(node) {
        if (isVarAwaitUsing(node)) {
            writeKeyword("await");
            writeSpace();
            writeKeyword("using");
        }
        else {
            const head = isLet(node) ? "let" :
                isVarConst(node) ? "const" :
                    isVarUsing(node) ? "using" :
                        "var";
            writeKeyword(head);
        }
        writeSpace();
        emitList(node, node.declarations, ListFormat.VariableDeclarationList);
    }
    function emitFunctionDeclaration(node) {
        emitFunctionDeclarationOrExpression(node);
    }
    function emitFunctionDeclarationOrExpression(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        writeKeyword("function");
        emit(node.asteriskToken);
        writeSpace();
        emitIdentifierName(node.name);
        emitSignatureAndBody(node, emitSignatureHead, emitFunctionBody);
    }
    function emitSignatureAndBody(node, emitSignatureHead, emitBody) {
        const indentedFlag = getEmitFlags(node) & EmitFlags.Indented;
        if (indentedFlag) {
            increaseIndent();
        }
        pushNameGenerationScope(node);
        forEach(node.parameters, generateNames);
        emitSignatureHead(node);
        emitBody(node);
        popNameGenerationScope(node);
        if (indentedFlag) {
            decreaseIndent();
        }
    }
    function emitFunctionBody(node) {
        const body = node.body;
        if (body) {
            emitBlockFunctionBody(body);
        }
        else {
            writeTrailingSemicolon();
        }
    }
    function emitEmptyFunctionBody(_node) {
        writeTrailingSemicolon();
    }
    function emitSignatureHead(node) {
        emitTypeParameters(node, node.typeParameters);
        emitParameters(node, node.parameters);
        emitTypeAnnotation(node.type);
    }
    function shouldEmitBlockFunctionBodyOnSingleLine(body) {
        // We must emit a function body as a single-line body in the following case:
        // * The body has NodeEmitFlags.SingleLine specified.
        // We must emit a function body as a multi-line body in the following cases:
        // * The body is explicitly marked as multi-line.
        // * A non-synthesized body's start and end position are on different lines.
        // * Any statement in the body starts on a new line.
        if (getEmitFlags(body) & EmitFlags.SingleLine) {
            return true;
        }
        if (body.multiLine) {
            return false;
        }
        if (!nodeIsSynthesized(body) && currentSourceFile && !rangeIsOnSingleLine(body, currentSourceFile)) {
            return false;
        }
        if (getLeadingLineTerminatorCount(body, firstOrUndefined(body.statements), ListFormat.PreserveLines)
            || getClosingLineTerminatorCount(body, lastOrUndefined(body.statements), ListFormat.PreserveLines, body.statements)) {
            return false;
        }
        let previousStatement;
        for (const statement of body.statements) {
            if (getSeparatingLineTerminatorCount(previousStatement, statement, ListFormat.PreserveLines) > 0) {
                return false;
            }
            previousStatement = statement;
        }
        return true;
    }
    function emitBlockFunctionBody(body) {
        generateNames(body);
        onBeforeEmitNode?.(body);
        writeSpace();
        writePunctuation("{");
        increaseIndent();
        const emitBlockFunctionBody = shouldEmitBlockFunctionBodyOnSingleLine(body)
            ? emitBlockFunctionBodyOnSingleLine
            : emitBlockFunctionBodyWorker;
        emitBodyWithDetachedComments(body, body.statements, emitBlockFunctionBody);
        decreaseIndent();
        writeToken(SyntaxKind.CloseBraceToken, body.statements.end, writePunctuation, body);
        onAfterEmitNode?.(body);
    }
    function emitBlockFunctionBodyOnSingleLine(body) {
        emitBlockFunctionBodyWorker(body, /*emitBlockFunctionBodyOnSingleLine*/ true);
    }
    function emitBlockFunctionBodyWorker(body, emitBlockFunctionBodyOnSingleLine) {
        // Emit all the prologue directives (like "use strict").
        const statementOffset = emitPrologueDirectives(body.statements);
        const pos = writer.getTextPos();
        emitHelpers(body);
        if (statementOffset === 0 && pos === writer.getTextPos() && emitBlockFunctionBodyOnSingleLine) {
            decreaseIndent();
            emitList(body, body.statements, ListFormat.SingleLineFunctionBodyStatements);
            increaseIndent();
        }
        else {
            emitList(body, body.statements, ListFormat.MultiLineFunctionBodyStatements, /*parenthesizerRule*/ undefined, statementOffset);
        }
    }
    function emitClassDeclaration(node) {
        emitClassDeclarationOrExpression(node);
    }
    function emitClassDeclarationOrExpression(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ true);
        emitTokenWithComment(SyntaxKind.ClassKeyword, moveRangePastModifiers(node).pos, writeKeyword, node);
        if (node.name) {
            writeSpace();
            emitIdentifierName(node.name);
        }
        const indentedFlag = getEmitFlags(node) & EmitFlags.Indented;
        if (indentedFlag) {
            increaseIndent();
        }
        emitTypeParameters(node, node.typeParameters);
        emitList(node, node.heritageClauses, ListFormat.ClassHeritageClauses);
        writeSpace();
        writePunctuation("{");
        pushNameGenerationScope(node);
        forEach(node.members, generateMemberNames);
        emitList(node, node.members, ListFormat.ClassMembers);
        popNameGenerationScope(node);
        writePunctuation("}");
        if (indentedFlag) {
            decreaseIndent();
        }
    }
    function emitInterfaceDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        writeKeyword("interface");
        writeSpace();
        emit(node.name);
        emitTypeParameters(node, node.typeParameters);
        emitList(node, node.heritageClauses, ListFormat.HeritageClauses);
        writeSpace();
        writePunctuation("{");
        pushNameGenerationScope(node);
        forEach(node.members, generateMemberNames);
        emitList(node, node.members, ListFormat.InterfaceMembers);
        popNameGenerationScope(node);
        writePunctuation("}");
    }
    function emitTypeAliasDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        writeKeyword("type");
        writeSpace();
        emit(node.name);
        emitTypeParameters(node, node.typeParameters);
        writeSpace();
        writePunctuation("=");
        writeSpace();
        emit(node.type);
        writeTrailingSemicolon();
    }
    function emitEnumDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        writeKeyword("enum");
        writeSpace();
        emit(node.name);
        writeSpace();
        writePunctuation("{");
        emitList(node, node.members, ListFormat.EnumMembers);
        writePunctuation("}");
    }
    function emitModuleDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        if (~node.flags & NodeFlags.GlobalAugmentation) {
            writeKeyword(node.flags & NodeFlags.Namespace ? "namespace" : "module");
            writeSpace();
        }
        emit(node.name);
        let body = node.body;
        if (!body)
            return writeTrailingSemicolon();
        while (body && isModuleDeclaration(body)) {
            writePunctuation(".");
            emit(body.name);
            body = body.body;
        }
        writeSpace();
        emit(body);
    }
    function emitModuleBlock(node) {
        pushNameGenerationScope(node);
        forEach(node.statements, generateNames);
        emitBlockStatements(node, /*forceSingleLine*/ isEmptyBlock(node));
        popNameGenerationScope(node);
    }
    function emitCaseBlock(node) {
        emitTokenWithComment(SyntaxKind.OpenBraceToken, node.pos, writePunctuation, node);
        emitList(node, node.clauses, ListFormat.CaseBlockClauses);
        emitTokenWithComment(SyntaxKind.CloseBraceToken, node.clauses.end, writePunctuation, node, /*indentLeading*/ true);
    }
    function emitImportEqualsDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        emitTokenWithComment(SyntaxKind.ImportKeyword, node.modifiers ? node.modifiers.end : node.pos, writeKeyword, node);
        writeSpace();
        if (node.isTypeOnly) {
            emitTokenWithComment(SyntaxKind.TypeKeyword, node.pos, writeKeyword, node);
            writeSpace();
        }
        emit(node.name);
        writeSpace();
        emitTokenWithComment(SyntaxKind.EqualsToken, node.name.end, writePunctuation, node);
        writeSpace();
        emitModuleReference(node.moduleReference);
        writeTrailingSemicolon();
    }
    function emitModuleReference(node) {
        if (node.kind === SyntaxKind.Identifier) {
            emitExpression(node);
        }
        else {
            emit(node);
        }
    }
    function emitImportDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        emitTokenWithComment(SyntaxKind.ImportKeyword, node.modifiers ? node.modifiers.end : node.pos, writeKeyword, node);
        writeSpace();
        if (node.importClause) {
            emit(node.importClause);
            writeSpace();
            emitTokenWithComment(SyntaxKind.FromKeyword, node.importClause.end, writeKeyword, node);
            writeSpace();
        }
        emitExpression(node.moduleSpecifier);
        if (node.attributes) {
            emitWithLeadingSpace(node.attributes);
        }
        writeTrailingSemicolon();
    }
    function emitImportClause(node) {
        if (node.phaseModifier !== undefined) {
            emitTokenWithComment(node.phaseModifier, node.pos, writeKeyword, node);
            writeSpace();
        }
        emit(node.name);
        if (node.name && node.namedBindings) {
            emitTokenWithComment(SyntaxKind.CommaToken, node.name.end, writePunctuation, node);
            writeSpace();
        }
        emit(node.namedBindings);
    }
    function emitNamespaceImport(node) {
        const asPos = emitTokenWithComment(SyntaxKind.AsteriskToken, node.pos, writePunctuation, node);
        writeSpace();
        emitTokenWithComment(SyntaxKind.AsKeyword, asPos, writeKeyword, node);
        writeSpace();
        emit(node.name);
    }
    function emitNamedImports(node) {
        emitNamedImportsOrExports(node);
    }
    function emitImportSpecifier(node) {
        emitImportOrExportSpecifier(node);
    }
    function emitExportAssignment(node) {
        const nextPos = emitTokenWithComment(SyntaxKind.ExportKeyword, node.pos, writeKeyword, node);
        writeSpace();
        if (node.isExportEquals) {
            emitTokenWithComment(SyntaxKind.EqualsToken, nextPos, writeOperator, node);
        }
        else {
            emitTokenWithComment(SyntaxKind.DefaultKeyword, nextPos, writeKeyword, node);
        }
        writeSpace();
        emitExpression(node.expression, node.isExportEquals ?
            parenthesizer.getParenthesizeRightSideOfBinaryForOperator(SyntaxKind.EqualsToken) :
            parenthesizer.parenthesizeExpressionOfExportDefault);
        writeTrailingSemicolon();
    }
    function emitExportDeclaration(node) {
        emitDecoratorsAndModifiers(node, node.modifiers, /*allowDecorators*/ false);
        let nextPos = emitTokenWithComment(SyntaxKind.ExportKeyword, node.pos, writeKeyword, node);
        writeSpace();
        if (node.isTypeOnly) {
            nextPos = emitTokenWithComment(SyntaxKind.TypeKeyword, nextPos, writeKeyword, node);
            writeSpace();
        }
        if (node.exportClause) {
            emit(node.exportClause);
        }
        else {
            nextPos = emitTokenWithComment(SyntaxKind.AsteriskToken, nextPos, writePunctuation, node);
        }
        if (node.moduleSpecifier) {
            writeSpace();
            const fromPos = node.exportClause ? node.exportClause.end : nextPos;
            emitTokenWithComment(SyntaxKind.FromKeyword, fromPos, writeKeyword, node);
            writeSpace();
            emitExpression(node.moduleSpecifier);
        }
        if (node.attributes) {
            emitWithLeadingSpace(node.attributes);
        }
        writeTrailingSemicolon();
    }
    function emitImportTypeNodeAttributes(node) {
        writePunctuation("{");
        writeSpace();
        writeKeyword(node.token === SyntaxKind.AssertKeyword ? "assert" : "with");
        writePunctuation(":");
        writeSpace();
        const elements = node.elements;
        emitList(node, elements, ListFormat.ImportAttributes);
        writeSpace();
        writePunctuation("}");
    }
    function emitImportAttributes(node) {
        emitTokenWithComment(node.token, node.pos, writeKeyword, node);
        writeSpace();
        const elements = node.elements;
        emitList(node, elements, ListFormat.ImportAttributes);
    }
    function emitImportAttribute(node) {
        emit(node.name);
        writePunctuation(":");
        writeSpace();
        const value = node.value;
        /** @see {emitPropertyAssignment} */
        if ((getEmitFlags(value) & EmitFlags.NoLeadingComments) === 0) {
            const commentRange = getCommentRange(value);
            emitTrailingCommentsOfPosition(commentRange.pos);
        }
        emit(value);
    }
    function emitNamespaceExportDeclaration(node) {
        let nextPos = emitTokenWithComment(SyntaxKind.ExportKeyword, node.pos, writeKeyword, node);
        writeSpace();
        nextPos = emitTokenWithComment(SyntaxKind.AsKeyword, nextPos, writeKeyword, node);
        writeSpace();
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        nextPos = emitTokenWithComment(SyntaxKind.NamespaceKeyword, nextPos, writeKeyword, node);
        writeSpace();
        emit(node.name);
        writeTrailingSemicolon();
    }
    function emitNamespaceExport(node) {
        const asPos = emitTokenWithComment(SyntaxKind.AsteriskToken, node.pos, writePunctuation, node);
        writeSpace();
        emitTokenWithComment(SyntaxKind.AsKeyword, asPos, writeKeyword, node);
        writeSpace();
        emit(node.name);
    }
    function emitNamedExports(node) {
        emitNamedImportsOrExports(node);
    }
    function emitExportSpecifier(node) {
        emitImportOrExportSpecifier(node);
    }
    function emitNamedImportsOrExports(node) {
        writePunctuation("{");
        emitList(node, node.elements, ListFormat.NamedImportsOrExportsElements);
        writePunctuation("}");
    }
    function emitImportOrExportSpecifier(node) {
        if (node.isTypeOnly) {
            writeKeyword("type");
            writeSpace();
        }
        if (node.propertyName) {
            emit(node.propertyName);
            writeSpace();
            emitTokenWithComment(SyntaxKind.AsKeyword, node.propertyName.end, writeKeyword, node);
            writeSpace();
        }
        emit(node.name);
    }
    //
    // Module references
    //
    function emitExternalModuleReference(node) {
        writeKeyword("require");
        writePunctuation("(");
        emitExpression(node.expression);
        writePunctuation(")");
    }
    //
    // JSX
    //
    function emitJsxElement(node) {
        emit(node.openingElement);
        emitList(node, node.children, ListFormat.JsxElementOrFragmentChildren);
        emit(node.closingElement);
    }
    function emitJsxSelfClosingElement(node) {
        writePunctuation("<");
        emitJsxTagName(node.tagName);
        emitTypeArguments(node, node.typeArguments);
        writeSpace();
        emit(node.attributes);
        writePunctuation("/>");
    }
    function emitJsxFragment(node) {
        emit(node.openingFragment);
        emitList(node, node.children, ListFormat.JsxElementOrFragmentChildren);
        emit(node.closingFragment);
    }
    function emitJsxOpeningElementOrFragment(node) {
        writePunctuation("<");
        if (isJsxOpeningElement(node)) {
            const indented = writeLineSeparatorsAndIndentBefore(node.tagName, node);
            emitJsxTagName(node.tagName);
            emitTypeArguments(node, node.typeArguments);
            if (node.attributes.properties && node.attributes.properties.length > 0) {
                writeSpace();
            }
            emit(node.attributes);
            writeLineSeparatorsAfter(node.attributes, node);
            decreaseIndentIf(indented);
        }
        writePunctuation(">");
    }
    function emitJsxText(node) {
        writer.writeLiteral(node.text);
    }
    function emitJsxClosingElementOrFragment(node) {
        writePunctuation("</");
        if (isJsxClosingElement(node)) {
            emitJsxTagName(node.tagName);
        }
        writePunctuation(">");
    }
    function emitJsxAttributes(node) {
        emitList(node, node.properties, ListFormat.JsxElementAttributes);
    }
    function emitJsxAttribute(node) {
        emit(node.name);
        emitNodeWithPrefix("=", writePunctuation, node.initializer, emitJsxAttributeValue);
    }
    function emitJsxSpreadAttribute(node) {
        writePunctuation("{...");
        emitExpression(node.expression);
        writePunctuation("}");
    }
    function hasTrailingCommentsAtPosition(pos) {
        let result = false;
        forEachTrailingCommentRange(currentSourceFile?.text || "", pos + 1, () => result = true);
        return result;
    }
    function hasLeadingCommentsAtPosition(pos) {
        let result = false;
        forEachLeadingCommentRange(currentSourceFile?.text || "", pos + 1, () => result = true);
        return result;
    }
    function hasCommentsAtPosition(pos) {
        return hasTrailingCommentsAtPosition(pos) || hasLeadingCommentsAtPosition(pos);
    }
    function emitJsxExpression(node) {
        if (node.expression || (!commentsDisabled && !nodeIsSynthesized(node) && hasCommentsAtPosition(node.pos))) { // preserve empty expressions if they contain comments!
            const isMultiline = currentSourceFile && !nodeIsSynthesized(node) && getLineAndCharacterOfPosition(currentSourceFile, node.pos).line !== getLineAndCharacterOfPosition(currentSourceFile, node.end).line;
            if (isMultiline) {
                writer.increaseIndent();
            }
            const end = emitTokenWithComment(SyntaxKind.OpenBraceToken, node.pos, writePunctuation, node);
            emit(node.dotDotDotToken);
            emitExpression(node.expression);
            emitTokenWithComment(SyntaxKind.CloseBraceToken, node.expression?.end || end, writePunctuation, node);
            if (isMultiline) {
                writer.decreaseIndent();
            }
        }
    }
    function emitJsxNamespacedName(node) {
        emitIdentifierName(node.namespace);
        writePunctuation(":");
        emitIdentifierName(node.name);
    }
    function emitJsxTagName(node) {
        if (node.kind === SyntaxKind.Identifier) {
            emitExpression(node);
        }
        else {
            emit(node);
        }
    }
    //
    // Clauses
    //
    function emitCaseClause(node) {
        emitTokenWithComment(SyntaxKind.CaseKeyword, node.pos, writeKeyword, node);
        writeSpace();
        emitExpression(node.expression, parenthesizer.parenthesizeExpressionForDisallowedComma);
        emitCaseOrDefaultClauseRest(node, node.statements, node.expression.end);
    }
    function emitDefaultClause(node) {
        const pos = emitTokenWithComment(SyntaxKind.DefaultKeyword, node.pos, writeKeyword, node);
        emitCaseOrDefaultClauseRest(node, node.statements, pos);
    }
    function emitCaseOrDefaultClauseRest(parentNode, statements, colonPos) {
        const emitAsSingleStatement = statements.length === 1 &&
            (
            // treat synthesized nodes as located on the same line for emit purposes
            !currentSourceFile ||
                nodeIsSynthesized(parentNode) ||
                nodeIsSynthesized(statements[0]) ||
                rangeStartPositionsAreOnSameLine(parentNode, statements[0], currentSourceFile));
        let format = ListFormat.CaseOrDefaultClauseStatements;
        if (emitAsSingleStatement) {
            writeToken(SyntaxKind.ColonToken, colonPos, writePunctuation, parentNode);
            writeSpace();
            format &= ~(ListFormat.MultiLine | ListFormat.Indented);
        }
        else {
            emitTokenWithComment(SyntaxKind.ColonToken, colonPos, writePunctuation, parentNode);
        }
        emitList(parentNode, statements, format);
    }
    function emitHeritageClause(node) {
        writeSpace();
        writeTokenText(node.token, writeKeyword);
        writeSpace();
        emitList(node, node.types, ListFormat.HeritageClauseTypes);
    }
    function emitCatchClause(node) {
        const openParenPos = emitTokenWithComment(SyntaxKind.CatchKeyword, node.pos, writeKeyword, node);
        writeSpace();
        if (node.variableDeclaration) {
            emitTokenWithComment(SyntaxKind.OpenParenToken, openParenPos, writePunctuation, node);
            emit(node.variableDeclaration);
            emitTokenWithComment(SyntaxKind.CloseParenToken, node.variableDeclaration.end, writePunctuation, node);
            writeSpace();
        }
        emit(node.block);
    }
    //
    // Property assignments
    //
    function emitPropertyAssignment(node) {
        emit(node.name);
        writePunctuation(":");
        writeSpace();
        // This is to ensure that we emit comment in the following case:
        //      For example:
        //          obj = {
        //              id: /*comment1*/ ()=>void
        //          }
        // "comment1" is not considered to be leading comment for node.initializer
        // but rather a trailing comment on the previous node.
        const initializer = node.initializer;
        if ((getEmitFlags(initializer) & EmitFlags.NoLeadingComments) === 0) {
            const commentRange = getCommentRange(initializer);
            emitTrailingCommentsOfPosition(commentRange.pos);
        }
        emitExpression(initializer, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    function emitShorthandPropertyAssignment(node) {
        emit(node.name);
        if (node.objectAssignmentInitializer) {
            writeSpace();
            writePunctuation("=");
            writeSpace();
            emitExpression(node.objectAssignmentInitializer, parenthesizer.parenthesizeExpressionForDisallowedComma);
        }
    }
    function emitSpreadAssignment(node) {
        if (node.expression) {
            emitTokenWithComment(SyntaxKind.DotDotDotToken, node.pos, writePunctuation, node);
            emitExpression(node.expression, parenthesizer.parenthesizeExpressionForDisallowedComma);
        }
    }
    //
    // Enum
    //
    function emitEnumMember(node) {
        emit(node.name);
        emitInitializer(node.initializer, node.name.end, node, parenthesizer.parenthesizeExpressionForDisallowedComma);
    }
    //
    // JSDoc
    //
    function emitJSDoc(node) {
        write("/**");
        if (node.comment) {
            const text = getTextOfJSDocComment(node.comment);
            if (text) {
                const lines = text.split(/\r\n?|\n/);
                for (const line of lines) {
                    writeLine();
                    writeSpace();
                    writePunctuation("*");
                    writeSpace();
                    write(line);
                }
            }
        }
        if (node.tags) {
            if (node.tags.length === 1 && node.tags[0].kind === SyntaxKind.JSDocTypeTag && !node.comment) {
                writeSpace();
                emit(node.tags[0]);
            }
            else {
                emitList(node, node.tags, ListFormat.JSDocComment);
            }
        }
        writeSpace();
        write("*/");
    }
    function emitJSDocSimpleTypedTag(tag) {
        emitJSDocTagName(tag.tagName);
        emitJSDocTypeExpression(tag.typeExpression);
        emitJSDocComment(tag.comment);
    }
    function emitJSDocSeeTag(tag) {
        emitJSDocTagName(tag.tagName);
        emit(tag.name);
        emitJSDocComment(tag.comment);
    }
    function emitJSDocImportTag(tag) {
        emitJSDocTagName(tag.tagName);
        writeSpace();
        if (tag.importClause) {
            emit(tag.importClause);
            writeSpace();
            emitTokenWithComment(SyntaxKind.FromKeyword, tag.importClause.end, writeKeyword, tag);
            writeSpace();
        }
        emitExpression(tag.moduleSpecifier);
        if (tag.attributes) {
            emitWithLeadingSpace(tag.attributes);
        }
        emitJSDocComment(tag.comment);
    }
    function emitJSDocNameReference(node) {
        writeSpace();
        writePunctuation("{");
        emit(node.name);
        writePunctuation("}");
    }
    function emitJSDocHeritageTag(tag) {
        emitJSDocTagName(tag.tagName);
        writeSpace();
        writePunctuation("{");
        emit(tag.class);
        writePunctuation("}");
        emitJSDocComment(tag.comment);
    }
    function emitJSDocTemplateTag(tag) {
        emitJSDocTagName(tag.tagName);
        emitJSDocTypeExpression(tag.constraint);
        writeSpace();
        emitList(tag, tag.typeParameters, ListFormat.CommaListElements);
        emitJSDocComment(tag.comment);
    }
    function emitJSDocTypedefTag(tag) {
        emitJSDocTagName(tag.tagName);
        if (tag.typeExpression) {
            if (tag.typeExpression.kind === SyntaxKind.JSDocTypeExpression) {
                emitJSDocTypeExpression(tag.typeExpression);
            }
            else {
                writeSpace();
                writePunctuation("{");
                write("Object");
                if (tag.typeExpression.isArrayType) {
                    writePunctuation("[");
                    writePunctuation("]");
                }
                writePunctuation("}");
            }
        }
        if (tag.fullName) {
            writeSpace();
            emit(tag.fullName);
        }
        emitJSDocComment(tag.comment);
        if (tag.typeExpression && tag.typeExpression.kind === SyntaxKind.JSDocTypeLiteral) {
            emitJSDocTypeLiteral(tag.typeExpression);
        }
    }
    function emitJSDocCallbackTag(tag) {
        emitJSDocTagName(tag.tagName);
        if (tag.name) {
            writeSpace();
            emit(tag.name);
        }
        emitJSDocComment(tag.comment);
        emitJSDocSignature(tag.typeExpression);
    }
    function emitJSDocOverloadTag(tag) {
        emitJSDocComment(tag.comment);
        emitJSDocSignature(tag.typeExpression);
    }
    function emitJSDocSimpleTag(tag) {
        emitJSDocTagName(tag.tagName);
        emitJSDocComment(tag.comment);
    }
    function emitJSDocTypeLiteral(lit) {
        emitList(lit, factory.createNodeArray(lit.jsDocPropertyTags), ListFormat.JSDocComment);
    }
    function emitJSDocSignature(sig) {
        if (sig.typeParameters) {
            emitList(sig, factory.createNodeArray(sig.typeParameters), ListFormat.JSDocComment);
        }
        if (sig.parameters) {
            emitList(sig, factory.createNodeArray(sig.parameters), ListFormat.JSDocComment);
        }
        if (sig.type) {
            writeLine();
            writeSpace();
            writePunctuation("*");
            writeSpace();
            emit(sig.type);
        }
    }
    function emitJSDocPropertyLikeTag(param) {
        emitJSDocTagName(param.tagName);
        emitJSDocTypeExpression(param.typeExpression);
        writeSpace();
        if (param.isBracketed) {
            writePunctuation("[");
        }
        emit(param.name);
        if (param.isBracketed) {
            writePunctuation("]");
        }
        emitJSDocComment(param.comment);
    }
    function emitJSDocTagName(tagName) {
        writePunctuation("@");
        emit(tagName);
    }
    function emitJSDocComment(comment) {
        const text = getTextOfJSDocComment(comment);
        if (text) {
            writeSpace();
            write(text);
        }
    }
    function emitJSDocTypeExpression(typeExpression) {
        if (typeExpression) {
            writeSpace();
            writePunctuation("{");
            emit(typeExpression.type);
            writePunctuation("}");
        }
    }
    //
    // Top-level nodes
    //
    function emitSourceFile(node) {
        writeLine();
        const statements = node.statements;
        // Emit detached comment if there are no prologue directives or if the first node is synthesized.
        // The synthesized node will have no leading comment so some comments may be missed.
        const shouldEmitDetachedComment = statements.length === 0 ||
            !isPrologueDirective(statements[0]) ||
            nodeIsSynthesized(statements[0]);
        if (shouldEmitDetachedComment) {
            emitBodyWithDetachedComments(node, statements, emitSourceFileWorker);
            return;
        }
        emitSourceFileWorker(node);
    }
    function emitSyntheticTripleSlashReferencesIfNeeded(node) {
        emitTripleSlashDirectives(node.syntheticFileReferences || [], node.syntheticTypeReferences || [], node.syntheticLibReferences || []);
    }
    function emitTripleSlashDirectivesIfNeeded(node) {
        if (node.isDeclarationFile)
            emitTripleSlashDirectives(node.referencedFiles, node.typeReferenceDirectives, node.libReferenceDirectives);
    }
    function emitTripleSlashDirectives(files, types, libs) {
        if (currentSourceFile && currentSourceFile.moduleName) {
            writeComment(`/// <amd-module name="${currentSourceFile.moduleName}" />`);
            writeLine();
        }
        if (currentSourceFile && currentSourceFile.amdDependencies) {
            for (const dep of currentSourceFile.amdDependencies) {
                if (dep.name) {
                    writeComment(`/// <amd-dependency name="${dep.name}" path="${dep.path}" />`);
                }
                else {
                    writeComment(`/// <amd-dependency path="${dep.path}" />`);
                }
                writeLine();
            }
        }
        function writeDirectives(kind, directives) {
            for (const directive of directives) {
                const resolutionMode = directive.resolutionMode
                    ? `resolution-mode="${directive.resolutionMode === ModuleKind.ESNext ? "import" : "require"}" `
                    : "";
                const preserve = directive.preserve ? `preserve="true" ` : "";
                writeComment(`/// <reference ${kind}="${directive.fileName}" ${resolutionMode}${preserve}/>`);
                writeLine();
            }
        }
        writeDirectives("path", files);
        writeDirectives("types", types);
        writeDirectives("lib", libs);
    }
    function emitSourceFileWorker(node) {
        const statements = node.statements;
        pushNameGenerationScope(node);
        forEach(node.statements, generateNames);
        emitHelpers(node);
        const index = findIndex(statements, statement => !isPrologueDirective(statement));
        emitTripleSlashDirectivesIfNeeded(node);
        emitList(node, statements, ListFormat.MultiLine, /*parenthesizerRule*/ undefined, index === -1 ? statements.length : index);
        popNameGenerationScope(node);
    }
    // Transformation nodes
    function emitPartiallyEmittedExpression(node) {
        const emitFlags = getEmitFlags(node);
        if (!(emitFlags & EmitFlags.NoLeadingComments) && node.pos !== node.expression.pos) {
            emitTrailingCommentsOfPosition(node.expression.pos);
        }
        emitExpression(node.expression);
        if (!(emitFlags & EmitFlags.NoTrailingComments) && node.end !== node.expression.end) {
            emitLeadingCommentsOfPosition(node.expression.end);
        }
    }
    function emitCommaList(node) {
        emitExpressionList(node, node.elements, ListFormat.CommaListElements, /*parenthesizerRule*/ undefined);
    }
    /**
     * Emits any prologue directives at the start of a Statement list, returning the
     * number of prologue directives written to the output.
     */
    function emitPrologueDirectives(statements, sourceFile, seenPrologueDirectives) {
        let needsToSetSourceFile = !!sourceFile;
        for (let i = 0; i < statements.length; i++) {
            const statement = statements[i];
            if (isPrologueDirective(statement)) {
                const shouldEmitPrologueDirective = seenPrologueDirectives ? !seenPrologueDirectives.has(statement.expression.text) : true;
                if (shouldEmitPrologueDirective) {
                    if (needsToSetSourceFile) {
                        needsToSetSourceFile = false;
                        setSourceFile(sourceFile);
                    }
                    writeLine();
                    emit(statement);
                    if (seenPrologueDirectives) {
                        seenPrologueDirectives.add(statement.expression.text);
                    }
                }
            }
            else {
                // return index of the first non prologue directive
                return i;
            }
        }
        return statements.length;
    }
    function emitPrologueDirectivesIfNeeded(sourceFileOrBundle) {
        if (isSourceFile(sourceFileOrBundle)) {
            emitPrologueDirectives(sourceFileOrBundle.statements, sourceFileOrBundle);
        }
        else {
            const seenPrologueDirectives = new Set();
            for (const sourceFile of sourceFileOrBundle.sourceFiles) {
                emitPrologueDirectives(sourceFile.statements, sourceFile, seenPrologueDirectives);
            }
            setSourceFile(undefined);
        }
    }
    function emitShebangIfNeeded(sourceFileOrBundle) {
        if (isSourceFile(sourceFileOrBundle)) {
            const shebang = getShebang(sourceFileOrBundle.text);
            if (shebang) {
                writeComment(shebang);
                writeLine();
                return true;
            }
        }
        else {
            for (const sourceFile of sourceFileOrBundle.sourceFiles) {
                // Emit only the first encountered shebang
                if (emitShebangIfNeeded(sourceFile)) {
                    return true;
                }
            }
        }
    }
    //
    // Helpers
    //
    function emitNodeWithWriter(node, writer) {
        if (!node)
            return;
        const savedWrite = write;
        write = writer;
        emit(node);
        write = savedWrite;
    }
    function emitDecoratorsAndModifiers(node, modifiers, allowDecorators) {
        if (modifiers?.length) {
            if (every(modifiers, isModifier)) {
                // if all modifier-likes are `Modifier`, simply emit the array as modifiers.
                return emitModifierList(node, modifiers);
            }
            if (every(modifiers, isDecorator)) {
                if (allowDecorators) {
                    // if all modifier-likes are `Decorator`, simply emit the array as decorators.
                    return emitDecoratorList(node, modifiers);
                }
                return node.pos;
            }
            onBeforeEmitNodeArray?.(modifiers);
            // partition modifiers into contiguous chunks of `Modifier` or `Decorator`
            let lastMode;
            let mode;
            let start = 0;
            let pos = 0;
            let lastModifier;
            while (start < modifiers.length) {
                while (pos < modifiers.length) {
                    lastModifier = modifiers[pos];
                    mode = isDecorator(lastModifier) ? "decorators" : "modifiers";
                    if (lastMode === undefined) {
                        lastMode = mode;
                    }
                    else if (mode !== lastMode) {
                        break;
                    }
                    pos++;
                }
                const textRange = { pos: -1, end: -1 };
                if (start === 0)
                    textRange.pos = modifiers.pos;
                if (pos === modifiers.length - 1)
                    textRange.end = modifiers.end;
                if (lastMode === "modifiers" || allowDecorators) {
                    emitNodeListItems(emit, node, modifiers, lastMode === "modifiers" ? ListFormat.Modifiers : ListFormat.Decorators, 
                    /*parenthesizerRule*/ undefined, start, pos - start, 
                    /*hasTrailingComma*/ false, textRange);
                }
                start = pos;
                lastMode = mode;
                pos++;
            }
            onAfterEmitNodeArray?.(modifiers);
            if (lastModifier && !positionIsSynthesized(lastModifier.end)) {
                return lastModifier.end;
            }
        }
        return node.pos;
    }
    function emitModifierList(node, modifiers) {
        emitList(node, modifiers, ListFormat.Modifiers);
        const lastModifier = lastOrUndefined(modifiers);
        return lastModifier && !positionIsSynthesized(lastModifier.end) ? lastModifier.end : node.pos;
    }
    function emitTypeAnnotation(node) {
        if (node) {
            writePunctuation(":");
            writeSpace();
            emit(node);
        }
    }
    function emitInitializer(node, equalCommentStartPos, container, parenthesizerRule) {
        if (node) {
            writeSpace();
            emitTokenWithComment(SyntaxKind.EqualsToken, equalCommentStartPos, writeOperator, container);
            writeSpace();
            emitExpression(node, parenthesizerRule);
        }
    }
    function emitNodeWithPrefix(prefix, prefixWriter, node, emit) {
        if (node) {
            prefixWriter(prefix);
            emit(node);
        }
    }
    function emitWithLeadingSpace(node) {
        if (node) {
            writeSpace();
            emit(node);
        }
    }
    function emitExpressionWithLeadingSpace(node, parenthesizerRule) {
        if (node) {
            writeSpace();
            emitExpression(node, parenthesizerRule);
        }
    }
    function emitWithTrailingSpace(node) {
        if (node) {
            emit(node);
            writeSpace();
        }
    }
    function emitEmbeddedStatement(parent, node) {
        if (isBlock(node) ||
            getEmitFlags(parent) & EmitFlags.SingleLine ||
            preserveSourceNewlines && !getLeadingLineTerminatorCount(parent, node, ListFormat.None)) {
            writeSpace();
            emit(node);
        }
        else {
            writeLine();
            increaseIndent();
            if (isEmptyStatement(node)) {
                pipelineEmit(EmitHint.EmbeddedStatement, node);
            }
            else {
                emit(node);
            }
            decreaseIndent();
        }
    }
    function emitDecoratorList(parentNode, decorators) {
        emitList(parentNode, decorators, ListFormat.Decorators);
        const lastDecorator = lastOrUndefined(decorators);
        return lastDecorator && !positionIsSynthesized(lastDecorator.end) ? lastDecorator.end : parentNode.pos;
    }
    function emitTypeArguments(parentNode, typeArguments) {
        emitList(parentNode, typeArguments, ListFormat.TypeArguments, typeArgumentParenthesizerRuleSelector);
    }
    function emitTypeParameters(parentNode, typeParameters) {
        if (isFunctionLike(parentNode) && parentNode.typeArguments) { // Quick info uses type arguments in place of type parameters on instantiated signatures
            return emitTypeArguments(parentNode, parentNode.typeArguments);
        }
        emitList(parentNode, typeParameters, ListFormat.TypeParameters | (isArrowFunction(parentNode) ? ListFormat.AllowTrailingComma : ListFormat.None));
    }
    function emitParameters(parentNode, parameters) {
        emitList(parentNode, parameters, ListFormat.Parameters);
    }
    function canEmitSimpleArrowHead(parentNode, parameters) {
        const parameter = singleOrUndefined(parameters);
        return parameter
            && parameter.pos === parentNode.pos // may not have parsed tokens between parent and parameter
            && isArrowFunction(parentNode) // only arrow functions may have simple arrow head
            && !parentNode.type // arrow function may not have return type annotation
            && !some(parentNode.modifiers) // parent may not have decorators or modifiers
            && !some(parentNode.typeParameters) // parent may not have type parameters
            && !some(parameter.modifiers) // parameter may not have decorators or modifiers
            && !parameter.dotDotDotToken // parameter may not be rest
            && !parameter.questionToken // parameter may not be optional
            && !parameter.type // parameter may not have a type annotation
            && !parameter.initializer // parameter may not have an initializer
            && isIdentifier(parameter.name); // parameter name must be identifier
    }
    function emitParametersForArrow(parentNode, parameters) {
        if (canEmitSimpleArrowHead(parentNode, parameters)) {
            emitList(parentNode, parameters, ListFormat.Parameters & ~ListFormat.Parenthesis);
        }
        else {
            emitParameters(parentNode, parameters);
        }
    }
    function emitParametersForIndexSignature(parentNode, parameters) {
        emitList(parentNode, parameters, ListFormat.IndexSignatureParameters);
    }
    function writeDelimiter(format) {
        switch (format & ListFormat.DelimitersMask) {
            case ListFormat.None:
                break;
            case ListFormat.CommaDelimited:
                writePunctuation(",");
                break;
            case ListFormat.BarDelimited:
                writeSpace();
                writePunctuation("|");
                break;
            case ListFormat.AsteriskDelimited:
                writeSpace();
                writePunctuation("*");
                writeSpace();
                break;
            case ListFormat.AmpersandDelimited:
                writeSpace();
                writePunctuation("&");
                break;
        }
    }
    function emitList(parentNode, children, format, parenthesizerRule, start, count) {
        emitNodeList(emit, parentNode, children, format | (parentNode && getEmitFlags(parentNode) & EmitFlags.MultiLine ? ListFormat.PreferNewLine : 0), parenthesizerRule, start, count);
    }
    function emitExpressionList(parentNode, children, format, parenthesizerRule, start, count) {
        emitNodeList(emitExpression, parentNode, children, format, parenthesizerRule, start, count);
    }
    function emitNodeList(emit, parentNode, children, format, parenthesizerRule, start = 0, count = children ? children.length - start : 0) {
        const isUndefined = children === undefined;
        if (isUndefined && format & ListFormat.OptionalIfUndefined) {
            return;
        }
        const isEmpty = children === undefined || start >= children.length || count === 0;
        if (isEmpty && format & ListFormat.OptionalIfEmpty) {
            onBeforeEmitNodeArray?.(children);
            onAfterEmitNodeArray?.(children);
            return;
        }
        if (format & ListFormat.BracketsMask) {
            writePunctuation(getOpeningBracket(format));
            if (isEmpty && children) {
                emitTrailingCommentsOfPosition(children.pos, /*prefixSpace*/ true); // Emit comments within empty bracketed lists
            }
        }
        onBeforeEmitNodeArray?.(children);
        if (isEmpty) {
            // Write a line terminator if the parent node was multi-line
            if (format & ListFormat.MultiLine && !(preserveSourceNewlines && (!parentNode || currentSourceFile && rangeIsOnSingleLine(parentNode, currentSourceFile)))) {
                writeLine();
            }
            else if (format & ListFormat.SpaceBetweenBraces && !(format & ListFormat.NoSpaceIfEmpty)) {
                writeSpace();
            }
        }
        else {
            emitNodeListItems(emit, parentNode, children, format, parenthesizerRule, start, count, children.hasTrailingComma, children);
        }
        onAfterEmitNodeArray?.(children);
        if (format & ListFormat.BracketsMask) {
            if (isEmpty && children) {
                emitLeadingCommentsOfPosition(children.end); // Emit leading comments within empty lists
            }
            writePunctuation(getClosingBracket(format));
        }
    }
    /**
     * Emits a list without brackets or raising events.
     *
     * NOTE: You probably don't want to call this directly and should be using `emitList` or `emitExpressionList` instead.
     */
    function emitNodeListItems(emit, parentNode, children, format, parenthesizerRule, start, count, hasTrailingComma, childrenTextRange) {
        // Write the opening line terminator or leading whitespace.
        const mayEmitInterveningComments = (format & ListFormat.NoInterveningComments) === 0;
        let shouldEmitInterveningComments = mayEmitInterveningComments;
        const leadingLineTerminatorCount = getLeadingLineTerminatorCount(parentNode, children[start], format);
        if (leadingLineTerminatorCount) {
            writeLine(leadingLineTerminatorCount);
            shouldEmitInterveningComments = false;
        }
        else if (format & ListFormat.SpaceBetweenBraces) {
            writeSpace();
        }
        // Increase the indent, if requested.
        if (format & ListFormat.Indented) {
            increaseIndent();
        }
        const emitListItem = getEmitListItem(emit, parenthesizerRule);
        // Emit each child.
        let previousSibling;
        let shouldDecreaseIndentAfterEmit = false;
        for (let i = 0; i < count; i++) {
            const child = children[start + i];
            // Write the delimiter if this is not the first node.
            if (format & ListFormat.AsteriskDelimited) {
                // always write JSDoc in the format "\n *"
                writeLine();
                writeDelimiter(format);
            }
            else if (previousSibling) {
                // i.e
                //      function commentedParameters(
                //          /* Parameter a */
                //          a
                //          /* End of parameter a */ -> this comment isn't considered to be trailing comment of parameter "a" due to newline
                //          ,
                if (format & ListFormat.DelimitersMask && previousSibling.end !== (parentNode ? parentNode.end : -1)) {
                    const previousSiblingEmitFlags = getEmitFlags(previousSibling);
                    if (!(previousSiblingEmitFlags & EmitFlags.NoTrailingComments)) {
                        emitLeadingCommentsOfPosition(previousSibling.end);
                    }
                }
                writeDelimiter(format);
                // Write either a line terminator or whitespace to separate the elements.
                const separatingLineTerminatorCount = getSeparatingLineTerminatorCount(previousSibling, child, format);
                if (separatingLineTerminatorCount > 0) {
                    // If a synthesized node in a single-line list starts on a new
                    // line, we should increase the indent.
                    if ((format & (ListFormat.LinesMask | ListFormat.Indented)) === ListFormat.SingleLine) {
                        increaseIndent();
                        shouldDecreaseIndentAfterEmit = true;
                    }
                    if (shouldEmitInterveningComments && format & ListFormat.DelimitersMask && !positionIsSynthesized(child.pos)) {
                        const commentRange = getCommentRange(child);
                        emitTrailingCommentsOfPosition(commentRange.pos, /*prefixSpace*/ !!(format & ListFormat.SpaceBetweenSiblings), /*forceNoNewline*/ true);
                    }
                    writeLine(separatingLineTerminatorCount);
                    shouldEmitInterveningComments = false;
                }
                else if (previousSibling && format & ListFormat.SpaceBetweenSiblings) {
                    writeSpace();
                }
            }
            // Emit this child.
            if (shouldEmitInterveningComments) {
                const commentRange = getCommentRange(child);
                emitTrailingCommentsOfPosition(commentRange.pos);
            }
            else {
                shouldEmitInterveningComments = mayEmitInterveningComments;
            }
            nextListElementPos = child.pos;
            emitListItem(child, emit, parenthesizerRule, i);
            if (shouldDecreaseIndentAfterEmit) {
                decreaseIndent();
                shouldDecreaseIndentAfterEmit = false;
            }
            previousSibling = child;
        }
        // Write a trailing comma, if requested.
        const emitFlags = previousSibling ? getEmitFlags(previousSibling) : 0;
        const skipTrailingComments = commentsDisabled || !!(emitFlags & EmitFlags.NoTrailingComments);
        const emitTrailingComma = hasTrailingComma && (format & ListFormat.AllowTrailingComma) && (format & ListFormat.CommaDelimited);
        if (emitTrailingComma) {
            if (previousSibling && !skipTrailingComments) {
                emitTokenWithComment(SyntaxKind.CommaToken, previousSibling.end, writePunctuation, previousSibling);
            }
            else {
                writePunctuation(",");
            }
        }
        // Emit any trailing comment of the last element in the list
        // i.e
        //       var array = [...
        //          2
        //          /* end of element 2 */
        //       ];
        if (previousSibling && (parentNode ? parentNode.end : -1) !== previousSibling.end && (format & ListFormat.DelimitersMask) && !skipTrailingComments) {
            emitLeadingCommentsOfPosition(emitTrailingComma && childrenTextRange?.end ? childrenTextRange.end : previousSibling.end);
        }
        // Decrease the indent, if requested.
        if (format & ListFormat.Indented) {
            decreaseIndent();
        }
        // Write the closing line terminator or closing whitespace.
        const closingLineTerminatorCount = getClosingLineTerminatorCount(parentNode, children[start + count - 1], format, childrenTextRange);
        if (closingLineTerminatorCount) {
            writeLine(closingLineTerminatorCount);
        }
        else if (format & (ListFormat.SpaceAfterList | ListFormat.SpaceBetweenBraces)) {
            writeSpace();
        }
    }
    // Writers
    function writeLiteral(s) {
        writer.writeLiteral(s);
    }
    function writeStringLiteral(s) {
        writer.writeStringLiteral(s);
    }
    function writeBase(s) {
        writer.write(s);
    }
    function writeSymbol(s, sym) {
        writer.writeSymbol(s, sym);
    }
    function writePunctuation(s) {
        writer.writePunctuation(s);
    }
    function writeTrailingSemicolon() {
        writer.writeTrailingSemicolon(";");
    }
    function writeKeyword(s) {
        writer.writeKeyword(s);
    }
    function writeOperator(s) {
        writer.writeOperator(s);
    }
    function writeParameter(s) {
        writer.writeParameter(s);
    }
    function writeComment(s) {
        writer.writeComment(s);
    }
    function writeSpace() {
        writer.writeSpace(" ");
    }
    function writeProperty(s) {
        writer.writeProperty(s);
    }
    function nonEscapingWrite(s) {
        // This should be defined in a snippet-escaping text writer.
        if (writer.nonEscapingWrite) {
            writer.nonEscapingWrite(s);
        }
        else {
            writer.write(s);
        }
    }
    function writeLine(count = 1) {
        for (let i = 0; i < count; i++) {
            writer.writeLine(i > 0);
        }
    }
    function increaseIndent() {
        writer.increaseIndent();
    }
    function decreaseIndent() {
        writer.decreaseIndent();
    }
    function writeToken(token, pos, writer, contextNode) {
        return !sourceMapsDisabled
            ? emitTokenWithSourceMap(contextNode, token, writer, pos, writeTokenText)
            : writeTokenText(token, writer, pos);
    }
    function writeTokenNode(node, writer) {
        if (onBeforeEmitToken) {
            onBeforeEmitToken(node);
        }
        writer(tokenToString(node.kind));
        if (onAfterEmitToken) {
            onAfterEmitToken(node);
        }
    }
    function writeTokenText(token, writer, pos) {
        const tokenString = tokenToString(token);
        writer(tokenString);
        return pos < 0 ? pos : pos + tokenString.length;
    }
    function writeLineOrSpace(parentNode, prevChildNode, nextChildNode) {
        if (getEmitFlags(parentNode) & EmitFlags.SingleLine) {
            writeSpace();
        }
        else if (preserveSourceNewlines) {
            const lines = getLinesBetweenNodes(parentNode, prevChildNode, nextChildNode);
            if (lines) {
                writeLine(lines);
            }
            else {
                writeSpace();
            }
        }
        else {
            writeLine();
        }
    }
    function writeLines(text) {
        const lines = text.split(/\r\n?|\n/);
        const indentation = guessIndentation(lines);
        for (const lineText of lines) {
            const line = indentation ? lineText.slice(indentation) : lineText;
            if (line.length) {
                writeLine();
                write(line);
            }
        }
    }
    function writeLinesAndIndent(lineCount, writeSpaceIfNotIndenting) {
        if (lineCount) {
            increaseIndent();
            writeLine(lineCount);
        }
        else if (writeSpaceIfNotIndenting) {
            writeSpace();
        }
    }
    // Helper function to decrease the indent if we previously indented.  Allows multiple
    // previous indent values to be considered at a time.  This also allows caller to just
    // call this once, passing in all their appropriate indent values, instead of needing
    // to call this helper function multiple times.
    function decreaseIndentIf(value1, value2) {
        if (value1) {
            decreaseIndent();
        }
        if (value2) {
            decreaseIndent();
        }
    }
    function getLeadingLineTerminatorCount(parentNode, firstChild, format) {
        if (format & ListFormat.PreserveLines || preserveSourceNewlines) {
            if (format & ListFormat.PreferNewLine) {
                return 1;
            }
            if (firstChild === undefined) {
                return !parentNode || currentSourceFile && rangeIsOnSingleLine(parentNode, currentSourceFile) ? 0 : 1;
            }
            if (firstChild.pos === nextListElementPos) {
                // If this child starts at the beginning of a list item in a parent list, its leading
                // line terminators have already been written as the separating line terminators of the
                // parent list. Example:
                //
                // class Foo {
                //   constructor() {}
                //   public foo() {}
                // }
                //
                // The outer list is the list of class members, with one line terminator between the
                // constructor and the method. The constructor is written, the separating line terminator
                // is written, and then we start emitting the method. Its modifiers ([public]) constitute an inner
                // list, so we look for its leading line terminators. If we didn't know that we had already
                // written a newline as part of the parent list, it would appear that we need to write a
                // leading newline to start the modifiers.
                return 0;
            }
            if (firstChild.kind === SyntaxKind.JsxText) {
                // JsxText will be written with its leading whitespace, so don't add more manually.
                return 0;
            }
            if (currentSourceFile && parentNode &&
                !positionIsSynthesized(parentNode.pos) &&
                !nodeIsSynthesized(firstChild) &&
                (!firstChild.parent || getOriginalNode(firstChild.parent) === getOriginalNode(parentNode))) {
                if (preserveSourceNewlines) {
                    return getEffectiveLines(includeComments => getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter(firstChild.pos, parentNode.pos, currentSourceFile, includeComments));
                }
                return rangeStartPositionsAreOnSameLine(parentNode, firstChild, currentSourceFile) ? 0 : 1;
            }
            if (synthesizedNodeStartsOnNewLine(firstChild, format)) {
                return 1;
            }
        }
        return format & ListFormat.MultiLine ? 1 : 0;
    }
    function getSeparatingLineTerminatorCount(previousNode, nextNode, format) {
        if (format & ListFormat.PreserveLines || preserveSourceNewlines) {
            if (previousNode === undefined || nextNode === undefined) {
                return 0;
            }
            if (nextNode.kind === SyntaxKind.JsxText) {
                // JsxText will be written with its leading whitespace, so don't add more manually.
                return 0;
            }
            else if (currentSourceFile && !nodeIsSynthesized(previousNode) && !nodeIsSynthesized(nextNode)) {
                if (preserveSourceNewlines && siblingNodePositionsAreComparable(previousNode, nextNode)) {
                    return getEffectiveLines(includeComments => getLinesBetweenRangeEndAndRangeStart(previousNode, nextNode, currentSourceFile, includeComments));
                }
                // If `preserveSourceNewlines` is `false` we do not intend to preserve the effective lines between the
                // previous and next node. Instead we naively check whether nodes are on separate lines within the
                // same node parent. If so, we intend to preserve a single line terminator. This is less precise and
                // expensive than checking with `preserveSourceNewlines` as above, but the goal is not to preserve the
                // effective source lines between two sibling nodes.
                else if (!preserveSourceNewlines && originalNodesHaveSameParent(previousNode, nextNode)) {
                    return rangeEndIsOnSameLineAsRangeStart(previousNode, nextNode, currentSourceFile) ? 0 : 1;
                }
                // If the two nodes are not comparable, add a line terminator based on the format that can indicate
                // whether new lines are preferred or not.
                return format & ListFormat.PreferNewLine ? 1 : 0;
            }
            else if (synthesizedNodeStartsOnNewLine(previousNode, format) || synthesizedNodeStartsOnNewLine(nextNode, format)) {
                return 1;
            }
        }
        else if (getStartsOnNewLine(nextNode)) {
            return 1;
        }
        return format & ListFormat.MultiLine ? 1 : 0;
    }
    function getClosingLineTerminatorCount(parentNode, lastChild, format, childrenTextRange) {
        if (format & ListFormat.PreserveLines || preserveSourceNewlines) {
            if (format & ListFormat.PreferNewLine) {
                return 1;
            }
            if (lastChild === undefined) {
                return !parentNode || currentSourceFile && rangeIsOnSingleLine(parentNode, currentSourceFile) ? 0 : 1;
            }
            if (currentSourceFile && parentNode && !positionIsSynthesized(parentNode.pos) && !nodeIsSynthesized(lastChild) && (!lastChild.parent || lastChild.parent === parentNode)) {
                if (preserveSourceNewlines) {
                    const end = childrenTextRange && !positionIsSynthesized(childrenTextRange.end) ? childrenTextRange.end : lastChild.end;
                    return getEffectiveLines(includeComments => getLinesBetweenPositionAndNextNonWhitespaceCharacter(end, parentNode.end, currentSourceFile, includeComments));
                }
                return rangeEndPositionsAreOnSameLine(parentNode, lastChild, currentSourceFile) ? 0 : 1;
            }
            if (synthesizedNodeStartsOnNewLine(lastChild, format)) {
                return 1;
            }
        }
        if (format & ListFormat.MultiLine && !(format & ListFormat.NoTrailingNewLine)) {
            return 1;
        }
        return 0;
    }
    function getEffectiveLines(getLineDifference) {
        // If 'preserveSourceNewlines' is disabled, we should never call this function
        // because it could be more expensive than alternative approximations.
        Debug.assert(!!preserveSourceNewlines);
        // We start by measuring the line difference from a position to its adjacent comments,
        // so that this is counted as a one-line difference, not two:
        //
        //   node1;
        //   // NODE2 COMMENT
        //   node2;
        const lines = getLineDifference(/*includeComments*/ true);
        if (lines === 0) {
            // However, if the line difference considering comments was 0, we might have this:
            //
            //   node1; // NODE2 COMMENT
            //   node2;
            //
            // in which case we should be ignoring node2's comment, so this too is counted as
            // a one-line difference, not zero.
            return getLineDifference(/*includeComments*/ false);
        }
        return lines;
    }
    function writeLineSeparatorsAndIndentBefore(node, parent) {
        const leadingNewlines = preserveSourceNewlines && getLeadingLineTerminatorCount(parent, node, ListFormat.None);
        if (leadingNewlines) {
            writeLinesAndIndent(leadingNewlines, /*writeSpaceIfNotIndenting*/ false);
        }
        return !!leadingNewlines;
    }
    function writeLineSeparatorsAfter(node, parent) {
        const trailingNewlines = preserveSourceNewlines && getClosingLineTerminatorCount(parent, node, ListFormat.None, /*childrenTextRange*/ undefined);
        if (trailingNewlines) {
            writeLine(trailingNewlines);
        }
    }
    function synthesizedNodeStartsOnNewLine(node, format) {
        if (nodeIsSynthesized(node)) {
            const startsOnNewLine = getStartsOnNewLine(node);
            if (startsOnNewLine === undefined) {
                return (format & ListFormat.PreferNewLine) !== 0;
            }
            return startsOnNewLine;
        }
        return (format & ListFormat.PreferNewLine) !== 0;
    }
    function getLinesBetweenNodes(parent, node1, node2) {
        if (getEmitFlags(parent) & EmitFlags.NoIndentation) {
            return 0;
        }
        parent = skipSynthesizedParentheses(parent);
        node1 = skipSynthesizedParentheses(node1);
        node2 = skipSynthesizedParentheses(node2);
        // Always use a newline for synthesized code if the synthesizer desires it.
        if (getStartsOnNewLine(node2)) {
            return 1;
        }
        if (currentSourceFile && !nodeIsSynthesized(parent) && !nodeIsSynthesized(node1) && !nodeIsSynthesized(node2)) {
            if (preserveSourceNewlines) {
                return getEffectiveLines(includeComments => getLinesBetweenRangeEndAndRangeStart(node1, node2, currentSourceFile, includeComments));
            }
            return rangeEndIsOnSameLineAsRangeStart(node1, node2, currentSourceFile) ? 0 : 1;
        }
        return 0;
    }
    function isEmptyBlock(block) {
        return block.statements.length === 0
            && (!currentSourceFile || rangeEndIsOnSameLineAsRangeStart(block, block, currentSourceFile));
    }
    function skipSynthesizedParentheses(node) {
        while (node.kind === SyntaxKind.ParenthesizedExpression && nodeIsSynthesized(node)) {
            node = node.expression;
        }
        return node;
    }
    function getTextOfNode(node, includeTrivia) {
        if (isGeneratedIdentifier(node) || isGeneratedPrivateIdentifier(node)) {
            return generateName(node);
        }
        if (isStringLiteral(node) && node.textSourceNode) {
            return getTextOfNode(node.textSourceNode, includeTrivia);
        }
        const sourceFile = currentSourceFile; // const needed for control flow
        const canUseSourceFile = !!sourceFile && !!node.parent && !nodeIsSynthesized(node);
        if (isMemberName(node)) {
            if (!canUseSourceFile || getSourceFileOfNode(node) !== getOriginalNode(sourceFile)) {
                return idText(node);
            }
        }
        else if (isJsxNamespacedName(node)) {
            if (!canUseSourceFile || getSourceFileOfNode(node) !== getOriginalNode(sourceFile)) {
                return getTextOfJsxNamespacedName(node);
            }
        }
        else {
            Debug.assertNode(node, isLiteralExpression); // not strictly necessary
            if (!canUseSourceFile) {
                return node.text;
            }
        }
        return getSourceTextOfNodeFromSourceFile(sourceFile, node, includeTrivia);
    }
    function getLiteralTextOfNode(node, sourceFile = currentSourceFile, neverAsciiEscape, jsxAttributeEscape) {
        if (node.kind === SyntaxKind.StringLiteral && node.textSourceNode) {
            const textSourceNode = node.textSourceNode;
            if (isIdentifier(textSourceNode) || isPrivateIdentifier(textSourceNode) || isNumericLiteral(textSourceNode) || isJsxNamespacedName(textSourceNode)) {
                const text = isNumericLiteral(textSourceNode) ? textSourceNode.text : getTextOfNode(textSourceNode);
                return jsxAttributeEscape ? `"${escapeJsxAttributeString(text)}"` :
                    neverAsciiEscape || (getEmitFlags(node) & EmitFlags.NoAsciiEscaping) ? `"${escapeString(text)}"` :
                        `"${escapeNonAsciiString(text)}"`;
            }
            else {
                return getLiteralTextOfNode(textSourceNode, getSourceFileOfNode(textSourceNode), neverAsciiEscape, jsxAttributeEscape);
            }
        }
        const flags = (neverAsciiEscape ? GetLiteralTextFlags.NeverAsciiEscape : 0)
            | (jsxAttributeEscape ? GetLiteralTextFlags.JsxAttributeEscape : 0)
            | (printerOptions.terminateUnterminatedLiterals ? GetLiteralTextFlags.TerminateUnterminatedLiterals : 0)
            | (printerOptions.target && printerOptions.target >= ScriptTarget.ES2021 ? GetLiteralTextFlags.AllowNumericSeparator : 0);
        return getLiteralText(node, sourceFile, flags);
    }
    /**
     * Push a new name generation scope.
     */
    function pushNameGenerationScope(node) {
        privateNameTempFlagsStack.push(privateNameTempFlags);
        privateNameTempFlags = TempFlags.Auto;
        reservedPrivateNamesStack.push(reservedPrivateNames);
        if (node && getEmitFlags(node) & EmitFlags.ReuseTempVariableScope) {
            return;
        }
        tempFlagsStack.push(tempFlags);
        tempFlags = TempFlags.Auto;
        formattedNameTempFlagsStack.push(formattedNameTempFlags);
        formattedNameTempFlags = undefined;
        reservedNamesStack.push(reservedNames);
    }
    /**
     * Pop the current name generation scope.
     */
    function popNameGenerationScope(node) {
        privateNameTempFlags = privateNameTempFlagsStack.pop();
        reservedPrivateNames = reservedPrivateNamesStack.pop();
        if (node && getEmitFlags(node) & EmitFlags.ReuseTempVariableScope) {
            return;
        }
        tempFlags = tempFlagsStack.pop();
        formattedNameTempFlags = formattedNameTempFlagsStack.pop();
        reservedNames = reservedNamesStack.pop();
    }
    function reserveNameInNestedScopes(name) {
        if (!reservedNames || reservedNames === lastOrUndefined(reservedNamesStack)) {
            reservedNames = new Set();
        }
        reservedNames.add(name);
    }
    function reservePrivateNameInNestedScopes(name) {
        if (!reservedPrivateNames || reservedPrivateNames === lastOrUndefined(reservedPrivateNamesStack)) {
            reservedPrivateNames = new Set();
        }
        reservedPrivateNames.add(name);
    }
    function generateNames(node) {
        if (!node)
            return;
        switch (node.kind) {
            case SyntaxKind.Block:
                forEach(node.statements, generateNames);
                break;
            case SyntaxKind.LabeledStatement:
            case SyntaxKind.WithStatement:
            case SyntaxKind.DoStatement:
            case SyntaxKind.WhileStatement:
                generateNames(node.statement);
                break;
            case SyntaxKind.IfStatement:
                generateNames(node.thenStatement);
                generateNames(node.elseStatement);
                break;
            case SyntaxKind.ForStatement:
            case SyntaxKind.ForOfStatement:
            case SyntaxKind.ForInStatement:
                generateNames(node.initializer);
                generateNames(node.statement);
                break;
            case SyntaxKind.SwitchStatement:
                generateNames(node.caseBlock);
                break;
            case SyntaxKind.CaseBlock:
                forEach(node.clauses, generateNames);
                break;
            case SyntaxKind.CaseClause:
            case SyntaxKind.DefaultClause:
                forEach(node.statements, generateNames);
                break;
            case SyntaxKind.TryStatement:
                generateNames(node.tryBlock);
                generateNames(node.catchClause);
                generateNames(node.finallyBlock);
                break;
            case SyntaxKind.CatchClause:
                generateNames(node.variableDeclaration);
                generateNames(node.block);
                break;
            case SyntaxKind.VariableStatement:
                generateNames(node.declarationList);
                break;
            case SyntaxKind.VariableDeclarationList:
                forEach(node.declarations, generateNames);
                break;
            case SyntaxKind.VariableDeclaration:
            case SyntaxKind.Parameter:
            case SyntaxKind.BindingElement:
            case SyntaxKind.ClassDeclaration:
                generateNameIfNeeded(node.name);
                break;
            case SyntaxKind.FunctionDeclaration:
                generateNameIfNeeded(node.name);
                if (getEmitFlags(node) & EmitFlags.ReuseTempVariableScope) {
                    forEach(node.parameters, generateNames);
                    generateNames(node.body);
                }
                break;
            case SyntaxKind.ObjectBindingPattern:
            case SyntaxKind.ArrayBindingPattern:
                forEach(node.elements, generateNames);
                break;
            case SyntaxKind.ImportDeclaration:
                generateNames(node.importClause);
                break;
            case SyntaxKind.ImportClause:
                generateNameIfNeeded(node.name);
                generateNames(node.namedBindings);
                break;
            case SyntaxKind.NamespaceImport:
                generateNameIfNeeded(node.name);
                break;
            case SyntaxKind.NamespaceExport:
                generateNameIfNeeded(node.name);
                break;
            case SyntaxKind.NamedImports:
                forEach(node.elements, generateNames);
                break;
            case SyntaxKind.ImportSpecifier:
                generateNameIfNeeded(node.propertyName || node.name);
                break;
        }
    }
    function generateMemberNames(node) {
        if (!node)
            return;
        switch (node.kind) {
            case SyntaxKind.PropertyAssignment:
            case SyntaxKind.ShorthandPropertyAssignment:
            case SyntaxKind.PropertyDeclaration:
            case SyntaxKind.PropertySignature:
            case SyntaxKind.MethodDeclaration:
            case SyntaxKind.MethodSignature:
            case SyntaxKind.GetAccessor:
            case SyntaxKind.SetAccessor:
                generateNameIfNeeded(node.name);
                break;
        }
    }
    function generateNameIfNeeded(name) {
        if (name) {
            if (isGeneratedIdentifier(name) || isGeneratedPrivateIdentifier(name)) {
                generateName(name);
            }
            else if (isBindingPattern(name)) {
                generateNames(name);
            }
        }
    }
    /**
     * Generate the text for a generated identifier.
     */
    function generateName(name) {
        const autoGenerate = name.emitNode.autoGenerate;
        if ((autoGenerate.flags & GeneratedIdentifierFlags.KindMask) === GeneratedIdentifierFlags.Node) {
            // Node names generate unique names based on their original node
            // and are cached based on that node's id.
            return generateNameCached(getNodeForGeneratedName(name), isPrivateIdentifier(name), autoGenerate.flags, autoGenerate.prefix, autoGenerate.suffix);
        }
        else {
            // Auto, Loop, and Unique names are cached based on their unique
            // autoGenerateId.
            const autoGenerateId = autoGenerate.id;
            return autoGeneratedIdToGeneratedName[autoGenerateId] || (autoGeneratedIdToGeneratedName[autoGenerateId] = makeName(name));
        }
    }
    function generateNameCached(node, privateName, flags, prefix, suffix) {
        const nodeId = getNodeId(node);
        const cache = privateName ? nodeIdToGeneratedPrivateName : nodeIdToGeneratedName;
        return cache[nodeId] || (cache[nodeId] = generateNameForNode(node, privateName, flags ?? GeneratedIdentifierFlags.None, formatGeneratedNamePart(prefix, generateName), formatGeneratedNamePart(suffix)));
    }
    /**
     * Returns a value indicating whether a name is unique globally, within the current file,
     * or within the NameGenerator.
     */
    function isUniqueName(name, privateName) {
        return isFileLevelUniqueNameInCurrentFile(name, privateName)
            && !isReservedName(name, privateName)
            && !generatedNames.has(name);
    }
    function isReservedName(name, privateName) {
        let set;
        let stack;
        if (privateName) {
            set = reservedPrivateNames;
            stack = reservedPrivateNamesStack;
        }
        else {
            set = reservedNames;
            stack = reservedNamesStack;
        }
        if (set?.has(name)) {
            return true;
        }
        for (let i = stack.length - 1; i >= 0; i--) {
            if (set === stack[i]) {
                continue;
            }
            set = stack[i];
            if (set?.has(name)) {
                return true;
            }
        }
        return false;
    }
    /**
     * Returns a value indicating whether a name is unique globally or within the current file.
     *
     * @param _isPrivate (unused) this parameter exists to avoid an unnecessary adaptor frame in v8
     * when `isfileLevelUniqueName` is passed as a callback to `makeUniqueName`.
     */
    function isFileLevelUniqueNameInCurrentFile(name, _isPrivate) {
        return currentSourceFile ? isFileLevelUniqueName(currentSourceFile, name, hasGlobalName) : true;
    }
    /**
     * Returns a value indicating whether a name is unique within a container.
     */
    function isUniqueLocalName(name, container) {
        for (let node = container; node && isNodeDescendantOf(node, container); node = node.nextContainer) {
            if (canHaveLocals(node) && node.locals) {
                const local = node.locals.get(escapeLeadingUnderscores(name));
                // We conservatively include alias symbols to cover cases where they're emitted as locals
                if (local && local.flags & (SymbolFlags.Value | SymbolFlags.ExportValue | SymbolFlags.Alias)) {
                    return false;
                }
            }
        }
        return true;
    }
    function getTempFlags(formattedNameKey) {
        switch (formattedNameKey) {
            case "":
                return tempFlags;
            case "#":
                return privateNameTempFlags;
            default:
                return formattedNameTempFlags?.get(formattedNameKey) ?? TempFlags.Auto;
        }
    }
    function setTempFlags(formattedNameKey, flags) {
        switch (formattedNameKey) {
            case "":
                tempFlags = flags;
                break;
            case "#":
                privateNameTempFlags = flags;
                break;
            default:
                formattedNameTempFlags ??= new Map();
                formattedNameTempFlags.set(formattedNameKey, flags);
                break;
        }
    }
    /**
     * Return the next available name in the pattern _a ... _z, _0, _1, ...
     * TempFlags._i or TempFlags._n may be used to express a preference for that dedicated name.
     * Note that names generated by makeTempVariableName and makeUniqueName will never conflict.
     */
    function makeTempVariableName(flags, reservedInNestedScopes, privateName, prefix, suffix) {
        if (prefix.length > 0 && prefix.charCodeAt(0) === CharacterCodes.hash) {
            prefix = prefix.slice(1);
        }
        // Generate a key to use to acquire a TempFlags counter based on the fixed portions of the generated name.
        const key = formatGeneratedName(privateName, prefix, "", suffix);
        let tempFlags = getTempFlags(key);
        if (flags && !(tempFlags & flags)) {
            const name = flags === TempFlags._i ? "_i" : "_n";
            const fullName = formatGeneratedName(privateName, prefix, name, suffix);
            if (isUniqueName(fullName, privateName)) {
                tempFlags |= flags;
                if (privateName) {
                    reservePrivateNameInNestedScopes(fullName);
                }
                else if (reservedInNestedScopes) {
                    reserveNameInNestedScopes(fullName);
                }
                setTempFlags(key, tempFlags);
                return fullName;
            }
        }
        while (true) {
            const count = tempFlags & TempFlags.CountMask;
            tempFlags++;
            // Skip over 'i' and 'n'
            if (count !== 8 && count !== 13) {
                const name = count < 26
                    ? "_" + String.fromCharCode(CharacterCodes.a + count)
                    : "_" + (count - 26);
                const fullName = formatGeneratedName(privateName, prefix, name, suffix);
                if (isUniqueName(fullName, privateName)) {
                    if (privateName) {
                        reservePrivateNameInNestedScopes(fullName);
                    }
                    else if (reservedInNestedScopes) {
                        reserveNameInNestedScopes(fullName);
                    }
                    setTempFlags(key, tempFlags);
                    return fullName;
                }
            }
        }
    }
    /**
     * Generate a name that is unique within the current file and doesn't conflict with any names
     * in global scope. The name is formed by adding an '_n' suffix to the specified base name,
     * where n is a positive integer. Note that names generated by makeTempVariableName and
     * makeUniqueName are guaranteed to never conflict.
     * If `optimistic` is set, the first instance will use 'baseName' verbatim instead of 'baseName_1'
     */
    function makeUniqueName(baseName, checkFn = isUniqueName, optimistic, scoped, privateName, prefix, suffix) {
        if (baseName.length > 0 && baseName.charCodeAt(0) === CharacterCodes.hash) {
            baseName = baseName.slice(1);
        }
        if (prefix.length > 0 && prefix.charCodeAt(0) === CharacterCodes.hash) {
            prefix = prefix.slice(1);
        }
        if (optimistic) {
            const fullName = formatGeneratedName(privateName, prefix, baseName, suffix);
            if (checkFn(fullName, privateName)) {
                if (privateName) {
                    reservePrivateNameInNestedScopes(fullName);
                }
                else if (scoped) {
                    reserveNameInNestedScopes(fullName);
                }
                else {
                    generatedNames.add(fullName);
                }
                return fullName;
            }
        }
        // Find the first unique 'name_n', where n is a positive number
        if (baseName.charCodeAt(baseName.length - 1) !== CharacterCodes._) {
            baseName += "_";
        }
        let i = 1;
        while (true) {
            const fullName = formatGeneratedName(privateName, prefix, baseName + i, suffix);
            if (checkFn(fullName, privateName)) {
                if (privateName) {
                    reservePrivateNameInNestedScopes(fullName);
                }
                else if (scoped) {
                    reserveNameInNestedScopes(fullName);
                }
                else {
                    generatedNames.add(fullName);
                }
                return fullName;
            }
            i++;
        }
    }
    function makeFileLevelOptimisticUniqueName(name) {
        return makeUniqueName(name, isFileLevelUniqueNameInCurrentFile, /*optimistic*/ true, /*scoped*/ false, /*privateName*/ false, /*prefix*/ "", /*suffix*/ "");
    }
    /**
     * Generates a unique name for a ModuleDeclaration or EnumDeclaration.
     */
    function generateNameForModuleOrEnum(node) {
        const name = getTextOfNode(node.name);
        // Use module/enum name itself if it is unique, otherwise make a unique variation
        return isUniqueLocalName(name, tryCast(node, canHaveLocals)) ? name : makeUniqueName(name, isUniqueName, /*optimistic*/ false, /*scoped*/ false, /*privateName*/ false, /*prefix*/ "", /*suffix*/ "");
    }
    /**
     * Generates a unique name for an ImportDeclaration or ExportDeclaration.
     */
    function generateNameForImportOrExportDeclaration(node) {
        const expr = getExternalModuleName(node); // TODO: GH#18217
        const baseName = isStringLiteral(expr) ?
            makeIdentifierFromModuleName(expr.text) : "module";
        return makeUniqueName(baseName, isUniqueName, /*optimistic*/ false, /*scoped*/ false, /*privateName*/ false, /*prefix*/ "", /*suffix*/ "");
    }
    /**
     * Generates a unique name for a default export.
     */
    function generateNameForExportDefault() {
        return makeUniqueName("default", isUniqueName, /*optimistic*/ false, /*scoped*/ false, /*privateName*/ false, /*prefix*/ "", /*suffix*/ "");
    }
    /**
     * Generates a unique name for a class expression.
     */
    function generateNameForClassExpression() {
        return makeUniqueName("class", isUniqueName, /*optimistic*/ false, /*scoped*/ false, /*privateName*/ false, /*prefix*/ "", /*suffix*/ "");
    }
    function generateNameForMethodOrAccessor(node, privateName, prefix, suffix) {
        if (isIdentifier(node.name)) {
            return generateNameCached(node.name, privateName);
        }
        return makeTempVariableName(TempFlags.Auto, /*reservedInNestedScopes*/ false, privateName, prefix, suffix);
    }
    /**
     * Generates a unique name from a node.
     */
    function generateNameForNode(node, privateName, flags, prefix, suffix) {
        switch (node.kind) {
            case SyntaxKind.Identifier:
            case SyntaxKind.PrivateIdentifier:
                return makeUniqueName(getTextOfNode(node), isUniqueName, !!(flags & GeneratedIdentifierFlags.Optimistic), !!(flags & GeneratedIdentifierFlags.ReservedInNestedScopes), privateName, prefix, suffix);
            case SyntaxKind.ModuleDeclaration:
            case SyntaxKind.EnumDeclaration:
                Debug.assert(!prefix && !suffix && !privateName);
                return generateNameForModuleOrEnum(node);
            case SyntaxKind.ImportDeclaration:
            case SyntaxKind.ExportDeclaration:
                Debug.assert(!prefix && !suffix && !privateName);
                return generateNameForImportOrExportDeclaration(node);
            case SyntaxKind.FunctionDeclaration:
            case SyntaxKind.ClassDeclaration: {
                Debug.assert(!prefix && !suffix && !privateName);
                const name = node.name;
                if (name && !isGeneratedIdentifier(name)) {
                    return generateNameForNode(name, /*privateName*/ false, flags, prefix, suffix);
                }
                return generateNameForExportDefault();
            }
            case SyntaxKind.ExportAssignment:
                Debug.assert(!prefix && !suffix && !privateName);
                return generateNameForExportDefault();
            case SyntaxKind.ClassExpression:
                Debug.assert(!prefix && !suffix && !privateName);
                return generateNameForClassExpression();
            case SyntaxKind.MethodDeclaration:
            case SyntaxKind.GetAccessor:
            case SyntaxKind.SetAccessor:
                return generateNameForMethodOrAccessor(node, privateName, prefix, suffix);
            case SyntaxKind.ComputedPropertyName:
                return makeTempVariableName(TempFlags.Auto, /*reservedInNestedScopes*/ true, privateName, prefix, suffix);
            default:
                return makeTempVariableName(TempFlags.Auto, /*reservedInNestedScopes*/ false, privateName, prefix, suffix);
        }
    }
    /**
     * Generates a unique identifier for a node.
     */
    function makeName(name) {
        const autoGenerate = name.emitNode.autoGenerate;
        const prefix = formatGeneratedNamePart(autoGenerate.prefix, generateName);
        const suffix = formatGeneratedNamePart(autoGenerate.suffix);
        switch (autoGenerate.flags & GeneratedIdentifierFlags.KindMask) {
            case GeneratedIdentifierFlags.Auto:
                return makeTempVariableName(TempFlags.Auto, !!(autoGenerate.flags & GeneratedIdentifierFlags.ReservedInNestedScopes), isPrivateIdentifier(name), prefix, suffix);
            case GeneratedIdentifierFlags.Loop:
                Debug.assertNode(name, isIdentifier);
                return makeTempVariableName(TempFlags._i, !!(autoGenerate.flags & GeneratedIdentifierFlags.ReservedInNestedScopes), /*privateName*/ false, prefix, suffix);
            case GeneratedIdentifierFlags.Unique:
                return makeUniqueName(idText(name), (autoGenerate.flags & GeneratedIdentifierFlags.FileLevel) ? isFileLevelUniqueNameInCurrentFile : isUniqueName, !!(autoGenerate.flags & GeneratedIdentifierFlags.Optimistic), !!(autoGenerate.flags & GeneratedIdentifierFlags.ReservedInNestedScopes), isPrivateIdentifier(name), prefix, suffix);
        }
        return Debug.fail(`Unsupported GeneratedIdentifierKind: ${Debug.formatEnum(autoGenerate.flags & GeneratedIdentifierFlags.KindMask, ts.GeneratedIdentifierFlags, /*isFlags*/ true)}.`);
    }
    // Comments
    function pipelineEmitWithComments(hint, node) {
        const pipelinePhase = getNextPipelinePhase(PipelinePhase.Comments, hint, node);
        const savedContainerPos = containerPos;
        const savedContainerEnd = containerEnd;
        const savedDeclarationListContainerEnd = declarationListContainerEnd;
        emitCommentsBeforeNode(node);
        pipelinePhase(hint, node);
        emitCommentsAfterNode(node, savedContainerPos, savedContainerEnd, savedDeclarationListContainerEnd);
    }
    function emitCommentsBeforeNode(node) {
        const emitFlags = getEmitFlags(node);
        const commentRange = getCommentRange(node);
        // Emit leading comments
        emitLeadingCommentsOfNode(node, emitFlags, commentRange.pos, commentRange.end);
        if (emitFlags & EmitFlags.NoNestedComments) {
            commentsDisabled = true;
        }
    }
    function emitCommentsAfterNode(node, savedContainerPos, savedContainerEnd, savedDeclarationListContainerEnd) {
        const emitFlags = getEmitFlags(node);
        const commentRange = getCommentRange(node);
        // Emit trailing comments
        if (emitFlags & EmitFlags.NoNestedComments) {
            commentsDisabled = false;
        }
        emitTrailingCommentsOfNode(node, emitFlags, commentRange.pos, commentRange.end, savedContainerPos, savedContainerEnd, savedDeclarationListContainerEnd);
        const typeNode = getTypeNode(node);
        if (typeNode) {
            emitTrailingCommentsOfNode(node, emitFlags, typeNode.pos, typeNode.end, savedContainerPos, savedContainerEnd, savedDeclarationListContainerEnd);
        }
    }
    function emitLeadingCommentsOfNode(node, emitFlags, pos, end) {
        enterComment();
        hasWrittenComment = false;
        // We have to explicitly check that the node is JsxText because if the compilerOptions.jsx is "preserve" we will not do any transformation.
        // It is expensive to walk entire tree just to set one kind of node to have no comments.
        const skipLeadingComments = pos < 0 || (emitFlags & EmitFlags.NoLeadingComments) !== 0 || node.kind === SyntaxKind.JsxText;
        const skipTrailingComments = end < 0 || (emitFlags & EmitFlags.NoTrailingComments) !== 0 || node.kind === SyntaxKind.JsxText;
        // Save current container state on the stack.
        if ((pos > 0 || end > 0) && pos !== end) {
            // Emit leading comments if the position is not synthesized and the node
            // has not opted out from emitting leading comments.
            if (!skipLeadingComments) {
                emitLeadingComments(pos, /*isEmittedNode*/ node.kind !== SyntaxKind.NotEmittedStatement);
            }
            if (!skipLeadingComments || (pos >= 0 && (emitFlags & EmitFlags.NoLeadingComments) !== 0)) {
                // Advance the container position if comments get emitted or if they've been disabled explicitly using NoLeadingComments.
                containerPos = pos;
            }
            if (!skipTrailingComments || (end >= 0 && (emitFlags & EmitFlags.NoTrailingComments) !== 0)) {
                // As above.
                containerEnd = end;
                // To avoid invalid comment emit in a down-level binding pattern, we
                // keep track of the last declaration list container's end
                if (node.kind === SyntaxKind.VariableDeclarationList) {
                    declarationListContainerEnd = end;
                }
            }
        }
        forEach(getSyntheticLeadingComments(node), emitLeadingSynthesizedComment);
        exitComment();
    }
    function emitTrailingCommentsOfNode(node, emitFlags, pos, end, savedContainerPos, savedContainerEnd, savedDeclarationListContainerEnd) {
        enterComment();
        const skipTrailingComments = end < 0 || (emitFlags & EmitFlags.NoTrailingComments) !== 0 || node.kind === SyntaxKind.JsxText;
        forEach(getSyntheticTrailingComments(node), emitTrailingSynthesizedComment);
        if ((pos > 0 || end > 0) && pos !== end) {
            // Restore previous container state.
            containerPos = savedContainerPos;
            containerEnd = savedContainerEnd;
            declarationListContainerEnd = savedDeclarationListContainerEnd;
            // Emit trailing comments if the position is not synthesized and the node
            // has not opted out from emitting leading comments and is an emitted node.
            if (!skipTrailingComments && node.kind !== SyntaxKind.NotEmittedStatement) {
                emitTrailingComments(end);
            }
        }
        exitComment();
    }
    function emitLeadingSynthesizedComment(comment) {
        if (comment.hasLeadingNewline || comment.kind === SyntaxKind.SingleLineCommentTrivia) {
            writer.writeLine();
        }
        writeSynthesizedComment(comment);
        if (comment.hasTrailingNewLine || comment.kind === SyntaxKind.SingleLineCommentTrivia) {
            writer.writeLine();
        }
        else {
            writer.writeSpace(" ");
        }
    }
    function emitTrailingSynthesizedComment(comment) {
        if (!writer.isAtStartOfLine()) {
            writer.writeSpace(" ");
        }
        writeSynthesizedComment(comment);
        if (comment.hasTrailingNewLine) {
            writer.writeLine();
        }
    }
    function writeSynthesizedComment(comment) {
        const text = formatSynthesizedComment(comment);
        const lineMap = comment.kind === SyntaxKind.MultiLineCommentTrivia ? computeLineStarts(text) : undefined;
        writeCommentRange(text, lineMap, writer, 0, text.length, newLine);
    }
    function formatSynthesizedComment(comment) {
        return comment.kind === SyntaxKind.MultiLineCommentTrivia
            ? `/*${comment.text}*/`
            : `//${comment.text}`;
    }
    function emitBodyWithDetachedComments(node, detachedRange, emitCallback) {
        enterComment();
        const { pos, end } = detachedRange;
        const emitFlags = getEmitFlags(node);
        const skipLeadingComments = pos < 0 || (emitFlags & EmitFlags.NoLeadingComments) !== 0;
        const skipTrailingComments = commentsDisabled || end < 0 || (emitFlags & EmitFlags.NoTrailingComments) !== 0;
        if (!skipLeadingComments) {
            emitDetachedCommentsAndUpdateCommentsInfo(detachedRange);
        }
        exitComment();
        if (emitFlags & EmitFlags.NoNestedComments && !commentsDisabled) {
            commentsDisabled = true;
            emitCallback(node);
            commentsDisabled = false;
        }
        else {
            emitCallback(node);
        }
        enterComment();
        if (!skipTrailingComments) {
            emitLeadingComments(detachedRange.end, /*isEmittedNode*/ true);
            if (hasWrittenComment && !writer.isAtStartOfLine()) {
                writer.writeLine();
            }
        }
        exitComment();
    }
    function originalNodesHaveSameParent(nodeA, nodeB) {
        nodeA = getOriginalNode(nodeA);
        // For performance, do not call `getOriginalNode` for `nodeB` if `nodeA` doesn't even
        // have a parent node.
        return nodeA.parent && nodeA.parent === getOriginalNode(nodeB).parent;
    }
    function siblingNodePositionsAreComparable(previousNode, nextNode) {
        if (nextNode.pos < previousNode.end) {
            return false;
        }
        previousNode = getOriginalNode(previousNode);
        nextNode = getOriginalNode(nextNode);
        const parent = previousNode.parent;
        if (!parent || parent !== nextNode.parent) {
            return false;
        }
        const parentNodeArray = getContainingNodeArray(previousNode);
        const prevNodeIndex = parentNodeArray?.indexOf(previousNode);
        return prevNodeIndex !== undefined && prevNodeIndex > -1 && parentNodeArray.indexOf(nextNode) === prevNodeIndex + 1;
    }
    function emitLeadingComments(pos, isEmittedNode) {
        hasWrittenComment = false;
        if (isEmittedNode) {
            if (pos === 0 && currentSourceFile?.isDeclarationFile) {
                forEachLeadingCommentToEmit(pos, emitNonTripleSlashLeadingComment);
            }
            else {
                forEachLeadingCommentToEmit(pos, emitLeadingComment);
            }
        }
        else if (pos === 0) {
            // If the node will not be emitted in JS, remove all the comments(normal, pinned and ///) associated with the node,
            // unless it is a triple slash comment at the top of the file.
            // For Example:
            //      /// <reference-path ...>
            //      declare var x;
            //      /// <reference-path ...>
            //      interface F {}
            //  The first /// will NOT be removed while the second one will be removed even though both node will not be emitted
            forEachLeadingCommentToEmit(pos, emitTripleSlashLeadingComment);
        }
    }
    function emitTripleSlashLeadingComment(commentPos, commentEnd, kind, hasTrailingNewLine, rangePos) {
        if (isTripleSlashComment(commentPos, commentEnd)) {
            emitLeadingComment(commentPos, commentEnd, kind, hasTrailingNewLine, rangePos);
        }
    }
    function emitNonTripleSlashLeadingComment(commentPos, commentEnd, kind, hasTrailingNewLine, rangePos) {
        if (!isTripleSlashComment(commentPos, commentEnd)) {
            emitLeadingComment(commentPos, commentEnd, kind, hasTrailingNewLine, rangePos);
        }
    }
    function shouldWriteComment(text, pos) {
        if (printerOptions.onlyPrintJsDocStyle) {
            return (isJSDocLikeText(text, pos) || isPinnedComment(text, pos));
        }
        return true;
    }
    function emitLeadingComment(commentPos, commentEnd, kind, hasTrailingNewLine, rangePos) {
        if (!currentSourceFile || !shouldWriteComment(currentSourceFile.text, commentPos))
            return;
        if (!hasWrittenComment) {
            emitNewLineBeforeLeadingCommentOfPosition(getCurrentLineMap(), writer, rangePos, commentPos);
            hasWrittenComment = true;
        }
        // Leading comments are emitted at /*leading comment1 */space/*leading comment*/space
        emitPos(commentPos);
        writeCommentRange(currentSourceFile.text, getCurrentLineMap(), writer, commentPos, commentEnd, newLine);
        emitPos(commentEnd);
        if (hasTrailingNewLine) {
            writer.writeLine();
        }
        else if (kind === SyntaxKind.MultiLineCommentTrivia) {
            writer.writeSpace(" ");
        }
    }
    function emitLeadingCommentsOfPosition(pos) {
        if (commentsDisabled || pos === -1) {
            return;
        }
        emitLeadingComments(pos, /*isEmittedNode*/ true);
    }
    function emitTrailingComments(pos) {
        forEachTrailingCommentToEmit(pos, emitTrailingComment);
    }
    function emitTrailingComment(commentPos, commentEnd, _kind, hasTrailingNewLine) {
        if (!currentSourceFile || !shouldWriteComment(currentSourceFile.text, commentPos))
            return;
        // trailing comments are emitted at space/*trailing comment1 */space/*trailing comment2*/
        if (!writer.isAtStartOfLine()) {
            writer.writeSpace(" ");
        }
        emitPos(commentPos);
        writeCommentRange(currentSourceFile.text, getCurrentLineMap(), writer, commentPos, commentEnd, newLine);
        emitPos(commentEnd);
        if (hasTrailingNewLine) {
            writer.writeLine();
        }
    }
    function emitTrailingCommentsOfPosition(pos, prefixSpace, forceNoNewline) {
        if (commentsDisabled) {
            return;
        }
        enterComment();
        forEachTrailingCommentToEmit(pos, prefixSpace ? emitTrailingComment : forceNoNewline ? emitTrailingCommentOfPositionNoNewline : emitTrailingCommentOfPosition);
        exitComment();
    }
    function emitTrailingCommentOfPositionNoNewline(commentPos, commentEnd, kind) {
        if (!currentSourceFile)
            return;
        // trailing comments of a position are emitted at /*trailing comment1 */space/*trailing comment*/space
        emitPos(commentPos);
        writeCommentRange(currentSourceFile.text, getCurrentLineMap(), writer, commentPos, commentEnd, newLine);
        emitPos(commentEnd);
        if (kind === SyntaxKind.SingleLineCommentTrivia) {
            writer.writeLine(); // still write a newline for single-line comments, so closing tokens aren't written on the same line
        }
    }
    function emitTrailingCommentOfPosition(commentPos, commentEnd, _kind, hasTrailingNewLine) {
        if (!currentSourceFile)
            return;
        // trailing comments of a position are emitted at /*trailing comment1 */space/*trailing comment*/space
        emitPos(commentPos);
        writeCommentRange(currentSourceFile.text, getCurrentLineMap(), writer, commentPos, commentEnd, newLine);
        emitPos(commentEnd);
        if (hasTrailingNewLine) {
            writer.writeLine();
        }
        else {
            writer.writeSpace(" ");
        }
    }
    function forEachLeadingCommentToEmit(pos, cb) {
        // Emit the leading comments only if the container's pos doesn't match because the container should take care of emitting these comments
        if (currentSourceFile && (containerPos === -1 || pos !== containerPos)) {
            if (hasDetachedComments(pos)) {
                forEachLeadingCommentWithoutDetachedComments(cb);
            }
            else {
                forEachLeadingCommentRange(currentSourceFile.text, pos, cb, /*state*/ pos);
            }
        }
    }
    function forEachTrailingCommentToEmit(end, cb) {
        // Emit the trailing comments only if the container's end doesn't match because the container should take care of emitting these comments
        if (currentSourceFile && (containerEnd === -1 || (end !== containerEnd && end !== declarationListContainerEnd))) {
            forEachTrailingCommentRange(currentSourceFile.text, end, cb);
        }
    }
    function hasDetachedComments(pos) {
        return detachedCommentsInfo !== undefined && last(detachedCommentsInfo).nodePos === pos;
    }
    function forEachLeadingCommentWithoutDetachedComments(cb) {
        if (!currentSourceFile)
            return;
        // get the leading comments from detachedPos
        const pos = last(detachedCommentsInfo).detachedCommentEndPos;
        if (detachedCommentsInfo.length - 1) {
            detachedCommentsInfo.pop();
        }
        else {
            detachedCommentsInfo = undefined;
        }
        forEachLeadingCommentRange(currentSourceFile.text, pos, cb, /*state*/ pos);
    }
    function emitDetachedCommentsAndUpdateCommentsInfo(range) {
        const currentDetachedCommentInfo = currentSourceFile && emitDetachedComments(currentSourceFile.text, getCurrentLineMap(), writer, emitComment, range, newLine, commentsDisabled);
        if (currentDetachedCommentInfo) {
            if (detachedCommentsInfo) {
                detachedCommentsInfo.push(currentDetachedCommentInfo);
            }
            else {
                detachedCommentsInfo = [currentDetachedCommentInfo];
            }
        }
    }
    function emitComment(text, lineMap, writer, commentPos, commentEnd, newLine) {
        if (!currentSourceFile || !shouldWriteComment(currentSourceFile.text, commentPos))
            return;
        emitPos(commentPos);
        writeCommentRange(text, lineMap, writer, commentPos, commentEnd, newLine);
        emitPos(commentEnd);
    }
    /**
     * Determine if the given comment is a triple-slash
     *
     * @return true if the comment is a triple-slash comment else false
     */
    function isTripleSlashComment(commentPos, commentEnd) {
        return !!currentSourceFile && isRecognizedTripleSlashComment(currentSourceFile.text, commentPos, commentEnd);
    }
    // Source Maps
    function pipelineEmitWithSourceMaps(hint, node) {
        const pipelinePhase = getNextPipelinePhase(PipelinePhase.SourceMaps, hint, node);
        emitSourceMapsBeforeNode(node);
        pipelinePhase(hint, node);
        emitSourceMapsAfterNode(node);
    }
    function emitSourceMapsBeforeNode(node) {
        const emitFlags = getEmitFlags(node);
        const sourceMapRange = getSourceMapRange(node);
        // Emit leading sourcemap
        const source = sourceMapRange.source || sourceMapSource;
        if (node.kind !== SyntaxKind.NotEmittedStatement
            && (emitFlags & EmitFlags.NoLeadingSourceMap) === 0
            && sourceMapRange.pos >= 0) {
            emitSourcePos(sourceMapRange.source || sourceMapSource, skipSourceTrivia(source, sourceMapRange.pos));
        }
        if (emitFlags & EmitFlags.NoNestedSourceMaps) {
            sourceMapsDisabled = true;
        }
    }
    function emitSourceMapsAfterNode(node) {
        const emitFlags = getEmitFlags(node);
        const sourceMapRange = getSourceMapRange(node);
        // Emit trailing sourcemap
        if (emitFlags & EmitFlags.NoNestedSourceMaps) {
            sourceMapsDisabled = false;
        }
        if (node.kind !== SyntaxKind.NotEmittedStatement
            && (emitFlags & EmitFlags.NoTrailingSourceMap) === 0
            && sourceMapRange.end >= 0) {
            emitSourcePos(sourceMapRange.source || sourceMapSource, sourceMapRange.end);
        }
    }
    /**
     * Skips trivia such as comments and white-space that can be optionally overridden by the source-map source
     */
    function skipSourceTrivia(source, pos) {
        return source.skipTrivia ? source.skipTrivia(pos) : skipTrivia(source.text, pos);
    }
    /**
     * Emits a mapping.
     *
     * If the position is synthetic (undefined or a negative value), no mapping will be
     * created.
     *
     * @param pos The position.
     */
    function emitPos(pos) {
        if (sourceMapsDisabled || positionIsSynthesized(pos) || isJsonSourceMapSource(sourceMapSource)) {
            return;
        }
        const { line: sourceLine, character: sourceCharacter } = getLineAndCharacterOfPosition(sourceMapSource, pos);
        sourceMapGenerator.addMapping(writer.getLine(), writer.getColumn(), sourceMapSourceIndex, sourceLine, sourceCharacter, 
        /*nameIndex*/ undefined);
    }
    function emitSourcePos(source, pos) {
        if (source !== sourceMapSource) {
            const savedSourceMapSource = sourceMapSource;
            const savedSourceMapSourceIndex = sourceMapSourceIndex;
            setSourceMapSource(source);
            emitPos(pos);
            resetSourceMapSource(savedSourceMapSource, savedSourceMapSourceIndex);
        }
        else {
            emitPos(pos);
        }
    }
    /**
     * Emits a token of a node with possible leading and trailing source maps.
     *
     * @param node The node containing the token.
     * @param token The token to emit.
     * @param tokenStartPos The start pos of the token.
     * @param emitCallback The callback used to emit the token.
     */
    function emitTokenWithSourceMap(node, token, writer, tokenPos, emitCallback) {
        if (sourceMapsDisabled || node && isInJsonFile(node)) {
            return emitCallback(token, writer, tokenPos);
        }
        const emitNode = node && node.emitNode;
        const emitFlags = emitNode && emitNode.flags || EmitFlags.None;
        const range = emitNode && emitNode.tokenSourceMapRanges && emitNode.tokenSourceMapRanges[token];
        const source = range && range.source || sourceMapSource;
        tokenPos = skipSourceTrivia(source, range ? range.pos : tokenPos);
        if ((emitFlags & EmitFlags.NoTokenLeadingSourceMaps) === 0 && tokenPos >= 0) {
            emitSourcePos(source, tokenPos);
        }
        tokenPos = emitCallback(token, writer, tokenPos);
        if (range)
            tokenPos = range.end;
        if ((emitFlags & EmitFlags.NoTokenTrailingSourceMaps) === 0 && tokenPos >= 0) {
            emitSourcePos(source, tokenPos);
        }
        return tokenPos;
    }
    function setSourceMapSource(source) {
        if (sourceMapsDisabled) {
            return;
        }
        sourceMapSource = source;
        if (source === mostRecentlyAddedSourceMapSource) {
            // Fast path for when the new source map is the most recently added, in which case
            // we use its captured index without going through the source map generator.
            sourceMapSourceIndex = mostRecentlyAddedSourceMapSourceIndex;
            return;
        }
        if (isJsonSourceMapSource(source)) {
            return;
        }
        sourceMapSourceIndex = sourceMapGenerator.addSource(source.fileName);
        if (printerOptions.inlineSources) {
            sourceMapGenerator.setSourceContent(sourceMapSourceIndex, source.text);
        }
        mostRecentlyAddedSourceMapSource = source;
        mostRecentlyAddedSourceMapSourceIndex = sourceMapSourceIndex;
    }
    function resetSourceMapSource(source, sourceIndex) {
        sourceMapSource = source;
        sourceMapSourceIndex = sourceIndex;
    }
    function isJsonSourceMapSource(sourceFile) {
        return fileExtensionIs(sourceFile.fileName, Extension.Json);
    }
}
function createBracketsMap() {
    const brackets = [];
    brackets[ListFormat.Braces] = ["{", "}"];
    brackets[ListFormat.Parenthesis] = ["(", ")"];
    brackets[ListFormat.AngleBrackets] = ["<", ">"];
    brackets[ListFormat.SquareBrackets] = ["[", "]"];
    return brackets;
}
function getOpeningBracket(format) {
    return brackets[format & ListFormat.BracketsMask][0];
}
function getClosingBracket(format) {
    return brackets[format & ListFormat.BracketsMask][1];
}
// Flags enum to track count of temp variables and a few dedicated names
var TempFlags;
(function (TempFlags) {
    TempFlags[TempFlags["Auto"] = 0] = "Auto";
    TempFlags[TempFlags["CountMask"] = 268435455] = "CountMask";
    TempFlags[TempFlags["_i"] = 268435456] = "_i";
})(TempFlags || (TempFlags = {}));
function emitListItemNoParenthesizer(node, emit, _parenthesizerRule, _index) {
    emit(node);
}
function emitListItemWithParenthesizerRuleSelector(node, emit, parenthesizerRuleSelector, index) {
    emit(node, parenthesizerRuleSelector.select(index));
}
function emitListItemWithParenthesizerRule(node, emit, parenthesizerRule, _index) {
    emit(node, parenthesizerRule);
}
function getEmitListItem(emit, parenthesizerRule) {
    return emit.length === 1 ? emitListItemNoParenthesizer :
        typeof parenthesizerRule === "object" ? emitListItemWithParenthesizerRuleSelector :
            emitListItemWithParenthesizerRule;
}
