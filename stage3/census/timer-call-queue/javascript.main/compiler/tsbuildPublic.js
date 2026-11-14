import { arrayFrom, assertType, canJsonReportNoInputFiles, changeCompilerHostLikeToUseCache, clearMap, closeFileWatcher, closeFileWatcherOf, combinePaths, commonOptionsWithBuild, convertToRelativePath, copyProperties, createCompilerDiagnostic, createCompilerHostFromProgramHost, createDiagnosticReporter, createModuleResolutionCache, createModuleResolutionLoader, createProgramHost, createTypeReferenceDirectiveResolutionCache, createTypeReferenceResolutionLoader, createWatchFactory, createWatchHost, Debug, Diagnostics, emitFilesAndReportErrors, emptyArray, ExitStatus, findIndex, firstOrUndefinedIterator, flattenDiagnosticMessageText, forEach, forEachEntry, forEachKey, ForegroundColorEscapeSequences, formatColorAndReset, getAllProjectOutputs, getBuildInfo as ts_getBuildInfo, getBuildInfoFileVersionMap, getConfigFileParsingDiagnostics, getDirectoryPath, getEmitDeclarations, getErrorCountForSummary, getFileNamesFromConfigSpecs, getFilesInErrorForSummary, getFirstProjectOutput, getLocaleTimeString, getModifiedTime as ts_getModifiedTime, getNonIncrementalBuildInfoRoots, getNormalizedAbsolutePath, getParsedCommandLineOfConfigFile, getPendingEmitKindWithSeen, getSourceFileVersionAsHashFromText, getTsBuildInfoEmitOutputFilePath, getWatchErrorSummaryDiagnosticMessage, hasProperty, identity, isIgnoredFileFromWildCardWatching, isIncrementalBuildInfo, isIncrementalCompilation, isPackageJsonInfo, isSolutionConfig, loadWithModeAwareCache, missingFileModifiedTime, mutateMap, mutateMapSkippingNewValues, noop, parseConfigHostFromCompilerHostLike, PollingInterval, ProgramUpdateLevel, readBuilderProgram, resolveConfigFileProjectName, resolveLibrary, resolvePath, resolveProjectReferencePath, returnUndefined, setGetSourceFileAsHashVersioned, sys, toPath as ts_toPath, unorderedRemoveItem, updateErrorForNoInputFiles, updateSharedExtendedConfigFileWatcher, updateWatchingWildcardDirectories, UpToDateStatusType, version, WatchType, } from "./_namespaces/ts.js";
import * as performance from "./_namespaces/ts.performance.js";
const minimumDate = new Date(-8640000000000000);
var BuildResultFlags;
(function (BuildResultFlags) {
    BuildResultFlags[BuildResultFlags["None"] = 0] = "None";
    /**
     * No errors of any kind occurred during build
     */
    BuildResultFlags[BuildResultFlags["Success"] = 1] = "Success";
    /**
     * None of the .d.ts files emitted by this build were
     * different from the existing files on disk
     */
    BuildResultFlags[BuildResultFlags["DeclarationOutputUnchanged"] = 2] = "DeclarationOutputUnchanged";
    /** Errors in the build */
    BuildResultFlags[BuildResultFlags["AnyErrors"] = 4] = "AnyErrors";
})(BuildResultFlags || (BuildResultFlags = {}));
function getOrCreateValueFromConfigFileMap(configFileMap, resolved, createT) {
    const existingValue = configFileMap.get(resolved);
    let newValue;
    if (!existingValue) {
        newValue = createT();
        configFileMap.set(resolved, newValue);
    }
    return existingValue || newValue;
}
function getOrCreateValueMapFromConfigFileMap(configFileMap, resolved) {
    return getOrCreateValueFromConfigFileMap(configFileMap, resolved, () => new Map());
}
/**
 * Helper to use now method instead of current date for testing purposes to get consistent baselines
 */
function getCurrentTime(host) {
    return host.now ? host.now() : new Date();
}
/** @internal */
export function isCircularBuildOrder(buildOrder) {
    return !!buildOrder && !!buildOrder.buildOrder;
}
/** @internal */
export function getBuildOrderFromAnyBuildOrder(anyBuildOrder) {
    return isCircularBuildOrder(anyBuildOrder) ? anyBuildOrder.buildOrder : anyBuildOrder;
}
/**
 * Create a function that reports watch status by writing to the system and handles the formating of the diagnostic
 */
export function createBuilderStatusReporter(system, pretty) {
    return diagnostic => {
        let output = pretty ? `[${formatColorAndReset(getLocaleTimeString(system), ForegroundColorEscapeSequences.Grey)}] ` : `${getLocaleTimeString(system)} - `;
        output += `${flattenDiagnosticMessageText(diagnostic.messageText, system.newLine)}${system.newLine + system.newLine}`;
        system.write(output);
    };
}
function createSolutionBuilderHostBase(system, createProgram, reportDiagnostic, reportSolutionBuilderStatus) {
    const host = createProgramHost(system, createProgram);
    host.getModifiedTime = system.getModifiedTime ? path => system.getModifiedTime(path) : returnUndefined;
    host.setModifiedTime = system.setModifiedTime ? (path, date) => system.setModifiedTime(path, date) : noop;
    host.deleteFile = system.deleteFile ? path => system.deleteFile(path) : noop;
    host.reportDiagnostic = reportDiagnostic || createDiagnosticReporter(system);
    host.reportSolutionBuilderStatus = reportSolutionBuilderStatus || createBuilderStatusReporter(system);
    host.now = system.now?.bind(system); // For testing
    return host;
}
export function createSolutionBuilderHost(system = sys, createProgram, reportDiagnostic, reportSolutionBuilderStatus, reportErrorSummary) {
    const host = createSolutionBuilderHostBase(system, createProgram, reportDiagnostic, reportSolutionBuilderStatus);
    host.reportErrorSummary = reportErrorSummary;
    return host;
}
export function createSolutionBuilderWithWatchHost(system = sys, createProgram, reportDiagnostic, reportSolutionBuilderStatus, reportWatchStatus) {
    const host = createSolutionBuilderHostBase(system, createProgram, reportDiagnostic, reportSolutionBuilderStatus);
    const watchHost = createWatchHost(system, reportWatchStatus);
    copyProperties(host, watchHost);
    return host;
}
function getCompilerOptionsOfBuildOptions(buildOptions) {
    const result = {};
    commonOptionsWithBuild.forEach(option => {
        if (hasProperty(buildOptions, option.name))
            result[option.name] = buildOptions[option.name];
    });
    result.tscBuild = true;
    return result;
}
export function createSolutionBuilder(host, rootNames, defaultOptions) {
    return createSolutionBuilderWorker(/*watch*/ false, host, rootNames, defaultOptions);
}
export function createSolutionBuilderWithWatch(host, rootNames, defaultOptions, baseWatchOptions) {
    return createSolutionBuilderWorker(/*watch*/ true, host, rootNames, defaultOptions, baseWatchOptions);
}
function createSolutionBuilderState(watch, hostOrHostWithWatch, rootNames, options, baseWatchOptions) {
    const host = hostOrHostWithWatch;
    const hostWithWatch = hostOrHostWithWatch;
    // State of the solution
    const baseCompilerOptions = getCompilerOptionsOfBuildOptions(options);
    const compilerHost = createCompilerHostFromProgramHost(host, () => state.projectCompilerOptions);
    setGetSourceFileAsHashVersioned(compilerHost);
    compilerHost.getParsedCommandLine = fileName => parseConfigFile(state, fileName, toResolvedConfigFilePath(state, fileName));
    compilerHost.resolveModuleNameLiterals = host.resolveModuleNameLiterals?.bind(host);
    compilerHost.resolveTypeReferenceDirectiveReferences = host.resolveTypeReferenceDirectiveReferences?.bind(host);
    compilerHost.resolveLibrary = host.resolveLibrary?.bind(host);
    compilerHost.resolveModuleNames = host.resolveModuleNames?.bind(host);
    compilerHost.resolveTypeReferenceDirectives = host.resolveTypeReferenceDirectives?.bind(host);
    compilerHost.getModuleResolutionCache = host.getModuleResolutionCache?.bind(host);
    let moduleResolutionCache, typeReferenceDirectiveResolutionCache;
    if (!compilerHost.resolveModuleNameLiterals && !compilerHost.resolveModuleNames) {
        moduleResolutionCache = createModuleResolutionCache(compilerHost.getCurrentDirectory(), compilerHost.getCanonicalFileName);
        compilerHost.resolveModuleNameLiterals = (moduleNames, containingFile, redirectedReference, options, containingSourceFile) => loadWithModeAwareCache(moduleNames, containingFile, redirectedReference, options, containingSourceFile, host, moduleResolutionCache, createModuleResolutionLoader);
        compilerHost.getModuleResolutionCache = () => moduleResolutionCache;
    }
    if (!compilerHost.resolveTypeReferenceDirectiveReferences && !compilerHost.resolveTypeReferenceDirectives) {
        typeReferenceDirectiveResolutionCache = createTypeReferenceDirectiveResolutionCache(compilerHost.getCurrentDirectory(), compilerHost.getCanonicalFileName, 
        /*options*/ undefined, moduleResolutionCache?.getPackageJsonInfoCache(), moduleResolutionCache?.optionsToRedirectsKey);
        compilerHost.resolveTypeReferenceDirectiveReferences = (typeDirectiveNames, containingFile, redirectedReference, options, containingSourceFile) => loadWithModeAwareCache(typeDirectiveNames, containingFile, redirectedReference, options, containingSourceFile, host, typeReferenceDirectiveResolutionCache, createTypeReferenceResolutionLoader);
    }
    let libraryResolutionCache;
    if (!compilerHost.resolveLibrary) {
        libraryResolutionCache = createModuleResolutionCache(compilerHost.getCurrentDirectory(), compilerHost.getCanonicalFileName, /*options*/ undefined, moduleResolutionCache?.getPackageJsonInfoCache());
        compilerHost.resolveLibrary = (libraryName, resolveFrom, options) => resolveLibrary(libraryName, resolveFrom, options, host, libraryResolutionCache);
    }
    compilerHost.getBuildInfo = (fileName, configFilePath) => getBuildInfo(state, fileName, toResolvedConfigFilePath(state, configFilePath), /*modifiedTime*/ undefined);
    const { watchFile, watchDirectory, writeLog } = createWatchFactory(hostWithWatch, options);
    const state = {
        host,
        hostWithWatch,
        parseConfigFileHost: parseConfigHostFromCompilerHostLike(host),
        write: host.trace?.bind(host),
        // State of solution
        options,
        baseCompilerOptions,
        rootNames,
        baseWatchOptions,
        resolvedConfigFilePaths: new Map(),
        configFileCache: new Map(),
        projectStatus: new Map(),
        extendedConfigCache: new Map(),
        buildInfoCache: new Map(),
        outputTimeStamps: new Map(),
        builderPrograms: new Map(),
        diagnostics: new Map(),
        projectPendingBuild: new Map(),
        projectErrorsReported: new Map(),
        compilerHost,
        moduleResolutionCache,
        typeReferenceDirectiveResolutionCache,
        libraryResolutionCache,
        // Mutable state
        buildOrder: undefined,
        readFileWithCache: f => host.readFile(f),
        projectCompilerOptions: baseCompilerOptions,
        cache: undefined,
        allProjectBuildPending: true,
        needsSummary: true,
        watchAllProjectsPending: watch,
        // Watch state
        watch,
        allWatchedWildcardDirectories: new Map(),
        allWatchedInputFiles: new Map(),
        allWatchedConfigFiles: new Map(),
        allWatchedExtendedConfigFiles: new Map(),
        allWatchedPackageJsonFiles: new Map(),
        filesWatched: new Map(),
        lastCachedPackageJsonLookups: new Map(),
        timerToBuildInvalidatedProject: undefined,
        reportFileChangeDetected: false,
        watchFile,
        watchDirectory,
        writeLog,
    };
    return state;
}
function toPath(state, fileName) {
    return ts_toPath(fileName, state.compilerHost.getCurrentDirectory(), state.compilerHost.getCanonicalFileName);
}
function toResolvedConfigFilePath(state, fileName) {
    const { resolvedConfigFilePaths } = state;
    const path = resolvedConfigFilePaths.get(fileName);
    if (path !== undefined)
        return path;
    const resolvedPath = toPath(state, fileName);
    resolvedConfigFilePaths.set(fileName, resolvedPath);
    return resolvedPath;
}
function isParsedCommandLine(entry) {
    return !!entry.options;
}
function getCachedParsedConfigFile(state, configFilePath) {
    const value = state.configFileCache.get(configFilePath);
    return value && isParsedCommandLine(value) ? value : undefined;
}
function parseConfigFile(state, configFileName, configFilePath) {
    const { configFileCache } = state;
    const value = configFileCache.get(configFilePath);
    if (value) {
        return isParsedCommandLine(value) ? value : undefined;
    }
    performance.mark("SolutionBuilder::beforeConfigFileParsing");
    let diagnostic;
    const { parseConfigFileHost, baseCompilerOptions, baseWatchOptions, extendedConfigCache, host } = state;
    let parsed;
    if (host.getParsedCommandLine) {
        parsed = host.getParsedCommandLine(configFileName);
        if (!parsed)
            diagnostic = createCompilerDiagnostic(Diagnostics.File_0_not_found, configFileName);
    }
    else {
        parseConfigFileHost.onUnRecoverableConfigFileDiagnostic = d => diagnostic = d;
        parsed = getParsedCommandLineOfConfigFile(configFileName, baseCompilerOptions, parseConfigFileHost, extendedConfigCache, baseWatchOptions);
        parseConfigFileHost.onUnRecoverableConfigFileDiagnostic = noop;
    }
    configFileCache.set(configFilePath, parsed || diagnostic);
    performance.mark("SolutionBuilder::afterConfigFileParsing");
    performance.measure("SolutionBuilder::Config file parsing", "SolutionBuilder::beforeConfigFileParsing", "SolutionBuilder::afterConfigFileParsing");
    return parsed;
}
function resolveProjectName(state, name) {
    return resolveConfigFileProjectName(resolvePath(state.compilerHost.getCurrentDirectory(), name));
}
function createBuildOrder(state, roots) {
    const temporaryMarks = new Map();
    const permanentMarks = new Map();
    const circularityReportStack = [];
    let buildOrder;
    let circularDiagnostics;
    for (const root of roots) {
        visit(root);
    }
    return circularDiagnostics ?
        { buildOrder: buildOrder || emptyArray, circularDiagnostics } :
        buildOrder || emptyArray;
    function visit(configFileName, inCircularContext) {
        const projPath = toResolvedConfigFilePath(state, configFileName);
        // Already visited
        if (permanentMarks.has(projPath))
            return;
        // Circular
        if (temporaryMarks.has(projPath)) {
            if (!inCircularContext) {
                (circularDiagnostics || (circularDiagnostics = [])).push(createCompilerDiagnostic(Diagnostics.Project_references_may_not_form_a_circular_graph_Cycle_detected_Colon_0, circularityReportStack.join("\r\n")));
            }
            return;
        }
        temporaryMarks.set(projPath, true);
        circularityReportStack.push(configFileName);
        const parsed = parseConfigFile(state, configFileName, projPath);
        if (parsed && parsed.projectReferences) {
            for (const ref of parsed.projectReferences) {
                const resolvedRefPath = resolveProjectName(state, ref.path);
                visit(resolvedRefPath, inCircularContext || ref.circular);
            }
        }
        circularityReportStack.pop();
        permanentMarks.set(projPath, true);
        (buildOrder || (buildOrder = [])).push(configFileName);
    }
}
function getBuildOrder(state) {
    return state.buildOrder || createStateBuildOrder(state);
}
function createStateBuildOrder(state) {
    const buildOrder = createBuildOrder(state, state.rootNames.map(f => resolveProjectName(state, f)));
    // Clear all to ResolvedConfigFilePaths cache to start fresh
    state.resolvedConfigFilePaths.clear();
    // TODO(rbuckton): Should be a `Set`, but that requires changing the code below that uses `mutateMapSkippingNewValues`
    const currentProjects = new Set(getBuildOrderFromAnyBuildOrder(buildOrder).map(resolved => toResolvedConfigFilePath(state, resolved)));
    const noopOnDelete = { onDeleteValue: noop };
    // Config file cache
    mutateMapSkippingNewValues(state.configFileCache, currentProjects, noopOnDelete);
    mutateMapSkippingNewValues(state.projectStatus, currentProjects, noopOnDelete);
    mutateMapSkippingNewValues(state.builderPrograms, currentProjects, noopOnDelete);
    mutateMapSkippingNewValues(state.diagnostics, currentProjects, noopOnDelete);
    mutateMapSkippingNewValues(state.projectPendingBuild, currentProjects, noopOnDelete);
    mutateMapSkippingNewValues(state.projectErrorsReported, currentProjects, noopOnDelete);
    mutateMapSkippingNewValues(state.buildInfoCache, currentProjects, noopOnDelete);
    mutateMapSkippingNewValues(state.outputTimeStamps, currentProjects, noopOnDelete);
    mutateMapSkippingNewValues(state.lastCachedPackageJsonLookups, currentProjects, noopOnDelete);
    // Remove watches for the program no longer in the solution
    if (state.watch) {
        mutateMapSkippingNewValues(state.allWatchedConfigFiles, currentProjects, { onDeleteValue: closeFileWatcher });
        state.allWatchedExtendedConfigFiles.forEach(watcher => {
            watcher.projects.forEach(project => {
                if (!currentProjects.has(project)) {
                    watcher.projects.delete(project);
                }
            });
            watcher.close();
        });
        mutateMapSkippingNewValues(state.allWatchedWildcardDirectories, currentProjects, { onDeleteValue: existingMap => existingMap.forEach(closeFileWatcherOf) });
        mutateMapSkippingNewValues(state.allWatchedInputFiles, currentProjects, { onDeleteValue: existingMap => existingMap.forEach(closeFileWatcher) });
        mutateMapSkippingNewValues(state.allWatchedPackageJsonFiles, currentProjects, { onDeleteValue: existingMap => existingMap.forEach(closeFileWatcher) });
    }
    return state.buildOrder = buildOrder;
}
function getBuildOrderFor(state, project, onlyReferences) {
    const resolvedProject = project && resolveProjectName(state, project);
    const buildOrderFromState = getBuildOrder(state);
    if (isCircularBuildOrder(buildOrderFromState))
        return buildOrderFromState;
    if (resolvedProject) {
        const projectPath = toResolvedConfigFilePath(state, resolvedProject);
        const projectIndex = findIndex(buildOrderFromState, configFileName => toResolvedConfigFilePath(state, configFileName) === projectPath);
        if (projectIndex === -1)
            return undefined;
    }
    const buildOrder = resolvedProject ? createBuildOrder(state, [resolvedProject]) : buildOrderFromState;
    Debug.assert(!isCircularBuildOrder(buildOrder));
    Debug.assert(!onlyReferences || resolvedProject !== undefined);
    Debug.assert(!onlyReferences || buildOrder[buildOrder.length - 1] === resolvedProject);
    return onlyReferences ? buildOrder.slice(0, buildOrder.length - 1) : buildOrder;
}
function enableCache(state) {
    if (state.cache) {
        disableCache(state);
    }
    const { compilerHost, host } = state;
    const originalReadFileWithCache = state.readFileWithCache;
    const originalGetSourceFile = compilerHost.getSourceFile;
    const { originalReadFile, originalFileExists, originalDirectoryExists, originalCreateDirectory, originalWriteFile, getSourceFileWithCache, readFileWithCache, } = changeCompilerHostLikeToUseCache(host, fileName => toPath(state, fileName), (...args) => originalGetSourceFile.call(compilerHost, ...args));
    state.readFileWithCache = readFileWithCache;
    compilerHost.getSourceFile = getSourceFileWithCache;
    state.cache = {
        originalReadFile,
        originalFileExists,
        originalDirectoryExists,
        originalCreateDirectory,
        originalWriteFile,
        originalReadFileWithCache,
        originalGetSourceFile,
    };
}
function disableCache(state) {
    if (!state.cache)
        return;
    const { cache, host, compilerHost, extendedConfigCache, moduleResolutionCache, typeReferenceDirectiveResolutionCache, libraryResolutionCache } = state;
    host.readFile = cache.originalReadFile;
    host.fileExists = cache.originalFileExists;
    host.directoryExists = cache.originalDirectoryExists;
    host.createDirectory = cache.originalCreateDirectory;
    host.writeFile = cache.originalWriteFile;
    compilerHost.getSourceFile = cache.originalGetSourceFile;
    state.readFileWithCache = cache.originalReadFileWithCache;
    extendedConfigCache.clear();
    moduleResolutionCache?.clear();
    typeReferenceDirectiveResolutionCache?.clear();
    libraryResolutionCache?.clear();
    state.cache = undefined;
}
function clearProjectStatus(state, resolved) {
    state.projectStatus.delete(resolved);
    state.diagnostics.delete(resolved);
}
function addProjToQueue({ projectPendingBuild }, proj, updateLevel) {
    const value = projectPendingBuild.get(proj);
    if (value === undefined) {
        projectPendingBuild.set(proj, updateLevel);
    }
    else if (value < updateLevel) {
        projectPendingBuild.set(proj, updateLevel);
    }
}
function setupInitialBuild(state, cancellationToken) {
    // Set initial build if not already built
    if (!state.allProjectBuildPending)
        return;
    state.allProjectBuildPending = false;
    if (state.options.watch)
        reportWatchStatus(state, Diagnostics.Starting_compilation_in_watch_mode);
    enableCache(state);
    const buildOrder = getBuildOrderFromAnyBuildOrder(getBuildOrder(state));
    buildOrder.forEach(configFileName => state.projectPendingBuild.set(toResolvedConfigFilePath(state, configFileName), ProgramUpdateLevel.Update));
    if (cancellationToken) {
        cancellationToken.throwIfCancellationRequested();
    }
}
export var InvalidatedProjectKind;
(function (InvalidatedProjectKind) {
    InvalidatedProjectKind[InvalidatedProjectKind["Build"] = 0] = "Build";
    InvalidatedProjectKind[InvalidatedProjectKind["UpdateOutputFileStamps"] = 1] = "UpdateOutputFileStamps";
})(InvalidatedProjectKind || (InvalidatedProjectKind = {}));
function doneInvalidatedProject(state, projectPath) {
    state.projectPendingBuild.delete(projectPath);
    return state.diagnostics.has(projectPath) ?
        ExitStatus.DiagnosticsPresent_OutputsSkipped :
        ExitStatus.Success;
}
function createUpdateOutputFileStampsProject(state, project, projectPath, config, buildOrder) {
    let updateOutputFileStampsPending = true;
    return {
        kind: InvalidatedProjectKind.UpdateOutputFileStamps,
        project,
        projectPath,
        buildOrder,
        getCompilerOptions: () => config.options,
        getCurrentDirectory: () => state.compilerHost.getCurrentDirectory(),
        updateOutputFileStatmps: () => {
            updateOutputTimestamps(state, config, projectPath);
            updateOutputFileStampsPending = false;
        },
        done: () => {
            if (updateOutputFileStampsPending) {
                updateOutputTimestamps(state, config, projectPath);
            }
            performance.mark("SolutionBuilder::Timestamps only updates");
            return doneInvalidatedProject(state, projectPath);
        },
    };
}
var BuildStep;
(function (BuildStep) {
    BuildStep[BuildStep["CreateProgram"] = 0] = "CreateProgram";
    BuildStep[BuildStep["Emit"] = 1] = "Emit";
    BuildStep[BuildStep["QueueReferencingProjects"] = 2] = "QueueReferencingProjects";
    BuildStep[BuildStep["Done"] = 3] = "Done";
})(BuildStep || (BuildStep = {}));
function createBuildOrUpdateInvalidedProject(state, project, projectPath, projectIndex, config, status, buildOrder) {
    let step = BuildStep.CreateProgram;
    let program;
    let buildResult;
    return {
        kind: InvalidatedProjectKind.Build,
        project,
        projectPath,
        buildOrder,
        getCompilerOptions: () => config.options,
        getCurrentDirectory: () => state.compilerHost.getCurrentDirectory(),
        getBuilderProgram: () => withProgramOrUndefined(identity),
        getProgram: () => withProgramOrUndefined(program => program.getProgramOrUndefined()),
        getSourceFile: fileName => withProgramOrUndefined(program => program.getSourceFile(fileName)),
        getSourceFiles: () => withProgramOrEmptyArray(program => program.getSourceFiles()),
        getOptionsDiagnostics: cancellationToken => withProgramOrEmptyArray(program => program.getOptionsDiagnostics(cancellationToken)),
        getGlobalDiagnostics: cancellationToken => withProgramOrEmptyArray(program => program.getGlobalDiagnostics(cancellationToken)),
        getConfigFileParsingDiagnostics: () => withProgramOrEmptyArray(program => program.getConfigFileParsingDiagnostics()),
        getSyntacticDiagnostics: (sourceFile, cancellationToken) => withProgramOrEmptyArray(program => program.getSyntacticDiagnostics(sourceFile, cancellationToken)),
        getAllDependencies: sourceFile => withProgramOrEmptyArray(program => program.getAllDependencies(sourceFile)),
        getSemanticDiagnostics: (sourceFile, cancellationToken) => withProgramOrEmptyArray(program => program.getSemanticDiagnostics(sourceFile, cancellationToken)),
        getSemanticDiagnosticsOfNextAffectedFile: (cancellationToken, ignoreSourceFile) => withProgramOrUndefined(program => (program.getSemanticDiagnosticsOfNextAffectedFile) &&
            program.getSemanticDiagnosticsOfNextAffectedFile(cancellationToken, ignoreSourceFile)),
        emit: (targetSourceFile, writeFile, cancellationToken, emitOnlyDtsFiles, customTransformers) => {
            if (targetSourceFile || emitOnlyDtsFiles) {
                return withProgramOrUndefined(program => program.emit(targetSourceFile, writeFile, cancellationToken, emitOnlyDtsFiles, customTransformers || state.host.getCustomTransformers?.(project)));
            }
            executeSteps(BuildStep.CreateProgram, cancellationToken);
            return emit(writeFile, cancellationToken, customTransformers);
        },
        done,
    };
    function done(cancellationToken, writeFile, customTransformers) {
        executeSteps(BuildStep.Done, cancellationToken, writeFile, customTransformers);
        performance.mark("SolutionBuilder::Projects built");
        return doneInvalidatedProject(state, projectPath);
    }
    function withProgramOrUndefined(action) {
        executeSteps(BuildStep.CreateProgram);
        return program && action(program);
    }
    function withProgramOrEmptyArray(action) {
        return withProgramOrUndefined(action) || emptyArray;
    }
    function createProgram() {
        Debug.assert(program === undefined);
        if (state.options.dry) {
            reportStatus(state, Diagnostics.A_non_dry_build_would_build_project_0, project);
            buildResult = BuildResultFlags.Success;
            step = BuildStep.QueueReferencingProjects;
            return;
        }
        if (state.options.verbose)
            reportStatus(state, Diagnostics.Building_project_0, project);
        if (config.fileNames.length === 0) {
            reportAndStoreErrors(state, projectPath, getConfigFileParsingDiagnostics(config));
            // Nothing to build - must be a solution file, basically
            buildResult = BuildResultFlags.None;
            step = BuildStep.QueueReferencingProjects;
            return;
        }
        const { host, compilerHost } = state;
        state.projectCompilerOptions = config.options;
        // Update module resolution cache if needed
        state.moduleResolutionCache?.update(config.options);
        state.typeReferenceDirectiveResolutionCache?.update(config.options);
        // Create program
        program = host.createProgram(config.fileNames, config.options, compilerHost, getOldProgram(state, projectPath, config), getConfigFileParsingDiagnostics(config), config.projectReferences);
        if (state.watch) {
            const internalMap = state.moduleResolutionCache?.getPackageJsonInfoCache().getInternalMap();
            state.lastCachedPackageJsonLookups.set(projectPath, internalMap && new Set(arrayFrom(internalMap.values(), data => state.host.realpath && (isPackageJsonInfo(data) || data.directoryExists) ?
                state.host.realpath(combinePaths(data.packageDirectory, "package.json")) :
                combinePaths(data.packageDirectory, "package.json"))));
            state.builderPrograms.set(projectPath, program);
        }
        step++;
    }
    function emit(writeFileCallback, cancellationToken, customTransformers) {
        Debug.assertIsDefined(program);
        Debug.assert(step === BuildStep.Emit);
        // Actual Emit
        const { host, compilerHost } = state;
        const emittedOutputs = new Map();
        const options = program.getCompilerOptions();
        const isIncremental = isIncrementalCompilation(options);
        let outputTimeStampMap;
        let now;
        const { emitResult, diagnostics } = emitFilesAndReportErrors(program, d => host.reportDiagnostic(d), state.write, 
        /*reportSummary*/ undefined, (name, text, writeByteOrderMark, onError, sourceFiles, data) => {
            const path = toPath(state, name);
            emittedOutputs.set(toPath(state, name), name);
            if (data?.buildInfo) {
                // Update buildInfo cache
                now ||= getCurrentTime(state.host);
                const isChangedSignature = program.hasChangedEmitSignature?.();
                const existing = getBuildInfoCacheEntry(state, name, projectPath);
                if (existing) {
                    existing.buildInfo = data.buildInfo;
                    existing.modifiedTime = now;
                    if (isChangedSignature)
                        existing.latestChangedDtsTime = now;
                }
                else {
                    state.buildInfoCache.set(projectPath, {
                        path: toPath(state, name),
                        buildInfo: data.buildInfo,
                        modifiedTime: now,
                        latestChangedDtsTime: isChangedSignature ? now : undefined,
                    });
                }
            }
            const modifiedTime = data?.differsOnlyInMap ? ts_getModifiedTime(state.host, name) : undefined;
            (writeFileCallback || compilerHost.writeFile)(name, text, writeByteOrderMark, onError, sourceFiles, data);
            // Revert the timestamp for the d.ts that is same but differs only in d.ts map URL
            if (data?.differsOnlyInMap)
                state.host.setModifiedTime(name, modifiedTime);
            else if (!isIncremental && state.watch) {
                (outputTimeStampMap ||= getOutputTimeStampMap(state, projectPath)).set(path, now ||= getCurrentTime(state.host));
            }
        }, cancellationToken, 
        /*emitOnlyDtsFiles*/ undefined, customTransformers || state.host.getCustomTransformers?.(project));
        if ((!options.noEmitOnError || !diagnostics.length) &&
            (emittedOutputs.size || status.type !== UpToDateStatusType.OutOfDateBuildInfoWithErrors)) {
            // Update time stamps for rest of the outputs
            updateOutputTimestampsWorker(state, config, projectPath, Diagnostics.Updating_unchanged_output_timestamps_of_project_0, emittedOutputs);
        }
        state.projectErrorsReported.set(projectPath, true);
        buildResult = program.hasChangedEmitSignature?.() ? BuildResultFlags.None : BuildResultFlags.DeclarationOutputUnchanged;
        if (!diagnostics.length) {
            state.diagnostics.delete(projectPath);
            state.projectStatus.set(projectPath, {
                type: UpToDateStatusType.UpToDate,
                oldestOutputFileName: firstOrUndefinedIterator(emittedOutputs.values()) ?? getFirstProjectOutput(config, !host.useCaseSensitiveFileNames()),
            });
        }
        else {
            state.diagnostics.set(projectPath, diagnostics);
            state.projectStatus.set(projectPath, { type: UpToDateStatusType.Unbuildable, reason: `it had errors` });
            buildResult |= BuildResultFlags.AnyErrors;
        }
        afterProgramDone(state, program);
        step = BuildStep.QueueReferencingProjects;
        return emitResult;
    }
    function executeSteps(till, cancellationToken, writeFile, customTransformers) {
        while (step <= till && step < BuildStep.Done) {
            const currentStep = step;
            switch (step) {
                case BuildStep.CreateProgram:
                    createProgram();
                    break;
                case BuildStep.Emit:
                    emit(writeFile, cancellationToken, customTransformers);
                    break;
                case BuildStep.QueueReferencingProjects:
                    queueReferencingProjects(state, project, projectPath, projectIndex, config, buildOrder, Debug.checkDefined(buildResult));
                    step++;
                    break;
                // Should never be done
                case BuildStep.Done:
                default:
                    assertType(step);
            }
            Debug.assert(step > currentStep);
        }
    }
}
function getNextInvalidatedProjectCreateInfo(state, buildOrder, reportQueue) {
    if (!state.projectPendingBuild.size)
        return undefined;
    if (isCircularBuildOrder(buildOrder))
        return undefined;
    const { options, projectPendingBuild } = state;
    for (let projectIndex = 0; projectIndex < buildOrder.length; projectIndex++) {
        const project = buildOrder[projectIndex];
        const projectPath = toResolvedConfigFilePath(state, project);
        const updateLevel = state.projectPendingBuild.get(projectPath);
        if (updateLevel === undefined)
            continue;
        if (reportQueue) {
            reportQueue = false;
            reportBuildQueue(state, buildOrder);
        }
        const config = parseConfigFile(state, project, projectPath);
        if (!config) {
            reportParseConfigFileDiagnostic(state, projectPath);
            projectPendingBuild.delete(projectPath);
            continue;
        }
        if (updateLevel === ProgramUpdateLevel.Full) {
            watchConfigFile(state, project, projectPath, config);
            watchExtendedConfigFiles(state, projectPath, config);
            watchWildCardDirectories(state, project, projectPath, config);
            watchInputFiles(state, project, projectPath, config);
            watchPackageJsonFiles(state, project, projectPath, config);
        }
        else if (updateLevel === ProgramUpdateLevel.RootNamesAndUpdate) {
            // Update file names
            config.fileNames = getFileNamesFromConfigSpecs(config.options.configFile.configFileSpecs, getDirectoryPath(project), config.options, state.parseConfigFileHost);
            updateErrorForNoInputFiles(config.fileNames, project, config.options.configFile.configFileSpecs, config.errors, canJsonReportNoInputFiles(config.raw));
            watchInputFiles(state, project, projectPath, config);
            watchPackageJsonFiles(state, project, projectPath, config);
        }
        const status = getUpToDateStatus(state, config, projectPath);
        if (!options.force) {
            if (status.type === UpToDateStatusType.UpToDate) {
                verboseReportProjectStatus(state, project, status);
                reportAndStoreErrors(state, projectPath, getConfigFileParsingDiagnostics(config));
                projectPendingBuild.delete(projectPath);
                // Up to date, skip
                if (options.dry) {
                    // In a dry build, inform the user of this fact
                    reportStatus(state, Diagnostics.Project_0_is_up_to_date, project);
                }
                continue;
            }
            if (status.type === UpToDateStatusType.UpToDateWithUpstreamTypes || status.type === UpToDateStatusType.UpToDateWithInputFileText) {
                reportAndStoreErrors(state, projectPath, getConfigFileParsingDiagnostics(config));
                return {
                    kind: InvalidatedProjectKind.UpdateOutputFileStamps,
                    status,
                    project,
                    projectPath,
                    projectIndex,
                    config,
                };
            }
        }
        if (status.type === UpToDateStatusType.UpstreamBlocked) {
            verboseReportProjectStatus(state, project, status);
            reportAndStoreErrors(state, projectPath, getConfigFileParsingDiagnostics(config));
            projectPendingBuild.delete(projectPath);
            if (options.verbose) {
                reportStatus(state, status.upstreamProjectBlocked ?
                    Diagnostics.Skipping_build_of_project_0_because_its_dependency_1_was_not_built :
                    Diagnostics.Skipping_build_of_project_0_because_its_dependency_1_has_errors, project, status.upstreamProjectName);
            }
            continue;
        }
        if (status.type === UpToDateStatusType.ContainerOnly) {
            verboseReportProjectStatus(state, project, status);
            reportAndStoreErrors(state, projectPath, getConfigFileParsingDiagnostics(config));
            projectPendingBuild.delete(projectPath);
            // Do nothing
            continue;
        }
        return {
            kind: InvalidatedProjectKind.Build,
            status,
            project,
            projectPath,
            projectIndex,
            config,
        };
    }
    return undefined;
}
function createInvalidatedProjectWithInfo(state, info, buildOrder) {
    verboseReportProjectStatus(state, info.project, info.status);
    return info.kind !== InvalidatedProjectKind.UpdateOutputFileStamps ?
        createBuildOrUpdateInvalidedProject(state, info.project, info.projectPath, info.projectIndex, info.config, info.status, buildOrder) :
        createUpdateOutputFileStampsProject(state, info.project, info.projectPath, info.config, buildOrder);
}
function getNextInvalidatedProject(state, buildOrder, reportQueue) {
    const info = getNextInvalidatedProjectCreateInfo(state, buildOrder, reportQueue);
    if (!info)
        return info;
    return createInvalidatedProjectWithInfo(state, info, buildOrder);
}
function getOldProgram({ options, builderPrograms, compilerHost }, proj, parsed) {
    if (options.force)
        return undefined;
    const value = builderPrograms.get(proj);
    if (value)
        return value;
    return readBuilderProgram(parsed.options, compilerHost);
}
function afterProgramDone(state, program) {
    if (program) {
        if (state.host.afterProgramEmitAndDiagnostics) {
            state.host.afterProgramEmitAndDiagnostics(program);
        }
        program.releaseProgram();
    }
    state.projectCompilerOptions = state.baseCompilerOptions;
}
function isFileWatcherWithModifiedTime(value) {
    return !!value.watcher;
}
function getModifiedTime(state, fileName) {
    const path = toPath(state, fileName);
    const existing = state.filesWatched.get(path);
    if (state.watch && !!existing) {
        if (!isFileWatcherWithModifiedTime(existing))
            return existing;
        if (existing.modifiedTime)
            return existing.modifiedTime;
    }
    // In watch mode we store the modified times in the cache
    // This is either Date | FileWatcherWithModifiedTime because we query modified times first and
    // then after complete compilation of the project, watch the files so we dont want to loose these modified times.
    const result = ts_getModifiedTime(state.host, fileName);
    if (state.watch) {
        if (existing)
            existing.modifiedTime = result;
        else
            state.filesWatched.set(path, result);
    }
    return result;
}
function watchFile(state, file, callback, pollingInterval, options, watchType, project) {
    const path = toPath(state, file);
    const existing = state.filesWatched.get(path);
    if (existing && isFileWatcherWithModifiedTime(existing)) {
        existing.callbacks.push(callback);
    }
    else {
        const watcher = state.watchFile(file, (fileName, eventKind, modifiedTime) => {
            const existing = Debug.checkDefined(state.filesWatched.get(path));
            Debug.assert(isFileWatcherWithModifiedTime(existing));
            existing.modifiedTime = modifiedTime;
            existing.callbacks.forEach(cb => cb(fileName, eventKind, modifiedTime));
        }, pollingInterval, options, watchType, project);
        state.filesWatched.set(path, { callbacks: [callback], watcher, modifiedTime: existing });
    }
    return {
        close: () => {
            const existing = Debug.checkDefined(state.filesWatched.get(path));
            Debug.assert(isFileWatcherWithModifiedTime(existing));
            if (existing.callbacks.length === 1) {
                state.filesWatched.delete(path);
                closeFileWatcherOf(existing);
            }
            else {
                unorderedRemoveItem(existing.callbacks, callback);
            }
        },
    };
}
function getOutputTimeStampMap(state, resolvedConfigFilePath) {
    // Output timestamps are stored only in watch mode
    if (!state.watch)
        return undefined;
    let result = state.outputTimeStamps.get(resolvedConfigFilePath);
    if (!result)
        state.outputTimeStamps.set(resolvedConfigFilePath, result = new Map());
    return result;
}
function getBuildInfoCacheEntry(state, buildInfoPath, resolvedConfigPath) {
    const path = toPath(state, buildInfoPath);
    const existing = state.buildInfoCache.get(resolvedConfigPath);
    return existing?.path === path ? existing : undefined;
}
function getBuildInfo(state, buildInfoPath, resolvedConfigPath, modifiedTime) {
    const path = toPath(state, buildInfoPath);
    const existing = state.buildInfoCache.get(resolvedConfigPath);
    if (existing !== undefined && existing.path === path) {
        return existing.buildInfo || undefined;
    }
    const value = state.readFileWithCache(buildInfoPath);
    const buildInfo = value ? ts_getBuildInfo(buildInfoPath, value) : undefined;
    state.buildInfoCache.set(resolvedConfigPath, { path, buildInfo: buildInfo || false, modifiedTime: modifiedTime || missingFileModifiedTime });
    return buildInfo;
}
function checkConfigFileUpToDateStatus(state, configFile, oldestOutputFileTime, oldestOutputFileName) {
    // Check tsconfig time
    const tsconfigTime = getModifiedTime(state, configFile);
    if (oldestOutputFileTime < tsconfigTime) {
        return {
            type: UpToDateStatusType.OutOfDateWithSelf,
            outOfDateOutputFileName: oldestOutputFileName,
            newerInputFileName: configFile,
        };
    }
}
function getUpToDateStatusWorker(state, project, resolvedPath) {
    // Container if no files are specified in the project
    if (isSolutionConfig(project))
        return { type: UpToDateStatusType.ContainerOnly };
    // Fast check to see if reference projects are upto date and error free
    let referenceStatuses;
    const force = !!state.options.force;
    if (project.projectReferences) {
        state.projectStatus.set(resolvedPath, { type: UpToDateStatusType.ComputingUpstream });
        for (const ref of project.projectReferences) {
            const resolvedRef = resolveProjectReferencePath(ref);
            const resolvedRefPath = toResolvedConfigFilePath(state, resolvedRef);
            const resolvedConfig = parseConfigFile(state, resolvedRef, resolvedRefPath);
            const refStatus = getUpToDateStatus(state, resolvedConfig, resolvedRefPath);
            // Its a circular reference ignore the status of this project
            if (refStatus.type === UpToDateStatusType.ComputingUpstream ||
                refStatus.type === UpToDateStatusType.ContainerOnly) { // Container only ignore this project
                continue;
            }
            // An upstream project is blocked
            if (state.options.stopBuildOnErrors && (refStatus.type === UpToDateStatusType.Unbuildable ||
                refStatus.type === UpToDateStatusType.UpstreamBlocked)) {
                return {
                    type: UpToDateStatusType.UpstreamBlocked,
                    upstreamProjectName: ref.path,
                    upstreamProjectBlocked: refStatus.type === UpToDateStatusType.UpstreamBlocked,
                };
            }
            if (!force)
                (referenceStatuses ||= []).push({ ref, refStatus, resolvedRefPath, resolvedConfig });
        }
    }
    if (force)
        return { type: UpToDateStatusType.ForceBuild };
    // Check buildinfo first
    const { host } = state;
    const buildInfoPath = getTsBuildInfoEmitOutputFilePath(project.options);
    const isIncremental = isIncrementalCompilation(project.options);
    let buildInfoCacheEntry = getBuildInfoCacheEntry(state, buildInfoPath, resolvedPath);
    const buildInfoTime = buildInfoCacheEntry?.modifiedTime || ts_getModifiedTime(host, buildInfoPath);
    if (buildInfoTime === missingFileModifiedTime) {
        if (!buildInfoCacheEntry) {
            state.buildInfoCache.set(resolvedPath, {
                path: toPath(state, buildInfoPath),
                buildInfo: false,
                modifiedTime: buildInfoTime,
            });
        }
        return {
            type: UpToDateStatusType.OutputMissing,
            missingOutputFileName: buildInfoPath,
        };
    }
    const buildInfo = getBuildInfo(state, buildInfoPath, resolvedPath, buildInfoTime);
    if (!buildInfo) {
        // Error reading buildInfo
        return {
            type: UpToDateStatusType.ErrorReadingFile,
            fileName: buildInfoPath,
        };
    }
    const incrementalBuildInfo = isIncremental && isIncrementalBuildInfo(buildInfo) ? buildInfo : undefined;
    if ((incrementalBuildInfo || !isIncremental) && buildInfo.version !== version) {
        return {
            type: UpToDateStatusType.TsVersionOutputOfDate,
            version: buildInfo.version,
        };
    }
    if (!project.options.noCheck &&
        (buildInfo.errors || // TODO: syntax errors????
            buildInfo.checkPending)) {
        return {
            type: UpToDateStatusType.OutOfDateBuildInfoWithErrors,
            buildInfoFile: buildInfoPath,
        };
    }
    if (incrementalBuildInfo) {
        // If there are errors, we need to build project again to report it
        if (!project.options.noCheck &&
            (incrementalBuildInfo.changeFileSet?.length ||
                incrementalBuildInfo.semanticDiagnosticsPerFile?.length ||
                (getEmitDeclarations(project.options) &&
                    incrementalBuildInfo.emitDiagnosticsPerFile?.length))) {
            return {
                type: UpToDateStatusType.OutOfDateBuildInfoWithErrors,
                buildInfoFile: buildInfoPath,
            };
        }
        // If there are pending changes that are not emitted, project is out of date
        if (!project.options.noEmit &&
            (incrementalBuildInfo.changeFileSet?.length ||
                incrementalBuildInfo.affectedFilesPendingEmit?.length ||
                incrementalBuildInfo.pendingEmit !== undefined)) {
            return {
                type: UpToDateStatusType.OutOfDateBuildInfoWithPendingEmit,
                buildInfoFile: buildInfoPath,
            };
        }
        // Has not emitted some of the files, project is out of date
        if ((!project.options.noEmit ||
            (project.options.noEmit && getEmitDeclarations(project.options))) &&
            getPendingEmitKindWithSeen(project.options, incrementalBuildInfo.options || {}, 
            /*emitOnlyDtsFiles*/ undefined, !!project.options.noEmit)) {
            return {
                type: UpToDateStatusType.OutOfDateOptions,
                buildInfoFile: buildInfoPath,
            };
        }
    }
    // Check input files
    let oldestOutputFileTime = buildInfoTime;
    let oldestOutputFileName = buildInfoPath;
    let newestInputFileName = undefined;
    let newestInputFileTime = minimumDate;
    /** True if input file has changed timestamp but text is not changed, we can then do only timestamp updates on output to make it look up-to-date later */
    let pseudoInputUpToDate = false;
    const seenRoots = new Set();
    let buildInfoVersionMap;
    // Get timestamps of input files
    for (const inputFile of project.fileNames) {
        const inputTime = getModifiedTime(state, inputFile);
        if (inputTime === missingFileModifiedTime) {
            return {
                type: UpToDateStatusType.Unbuildable,
                reason: `${inputFile} does not exist`,
            };
        }
        const inputPath = toPath(state, inputFile);
        // If an buildInfo is older than the newest input, we can stop checking
        if (buildInfoTime < inputTime) {
            let version;
            let currentVersion;
            if (incrementalBuildInfo) {
                // Read files and see if they are same, read is anyways cached
                if (!buildInfoVersionMap)
                    buildInfoVersionMap = getBuildInfoFileVersionMap(incrementalBuildInfo, buildInfoPath, host);
                const resolvedInputPath = buildInfoVersionMap.roots.get(inputPath);
                version = buildInfoVersionMap.fileInfos.get(resolvedInputPath ?? inputPath);
                const text = version ? state.readFileWithCache(resolvedInputPath ?? inputFile) : undefined;
                currentVersion = text !== undefined ? getSourceFileVersionAsHashFromText(host, text) : undefined;
                if (version && version === currentVersion)
                    pseudoInputUpToDate = true;
            }
            if (!version || version !== currentVersion) {
                return {
                    type: UpToDateStatusType.OutOfDateWithSelf,
                    outOfDateOutputFileName: buildInfoPath,
                    newerInputFileName: inputFile,
                };
            }
        }
        if (inputTime > newestInputFileTime) {
            newestInputFileName = inputFile;
            newestInputFileTime = inputTime;
        }
        seenRoots.add(inputPath);
    }
    let existingRoot;
    if (incrementalBuildInfo) {
        if (!buildInfoVersionMap)
            buildInfoVersionMap = getBuildInfoFileVersionMap(incrementalBuildInfo, buildInfoPath, host);
        existingRoot = forEachEntry(buildInfoVersionMap.roots, 
        // File was root file when project was built but its not any more
        (_resolved, existingRoot) => !seenRoots.has(existingRoot) ? existingRoot : undefined);
    }
    else {
        existingRoot = forEach(getNonIncrementalBuildInfoRoots(buildInfo, buildInfoPath, host), root => !seenRoots.has(root) ? root : undefined);
    }
    if (existingRoot) {
        return {
            type: UpToDateStatusType.OutOfDateRoots,
            buildInfoFile: buildInfoPath,
            inputFile: existingRoot,
        };
    }
    // Now see if all outputs are newer than the newest input
    // Dont check output timestamps if we have buildinfo telling us output is uptodate
    if (!isIncremental) {
        // Collect the expected outputs of this project
        const outputs = getAllProjectOutputs(project, !host.useCaseSensitiveFileNames());
        const outputTimeStampMap = getOutputTimeStampMap(state, resolvedPath);
        for (const output of outputs) {
            if (output === buildInfoPath)
                continue;
            const path = toPath(state, output);
            // Output is missing; can stop checking
            let outputTime = outputTimeStampMap?.get(path);
            if (!outputTime) {
                outputTime = ts_getModifiedTime(state.host, output);
                outputTimeStampMap?.set(path, outputTime);
            }
            if (outputTime === missingFileModifiedTime) {
                return {
                    type: UpToDateStatusType.OutputMissing,
                    missingOutputFileName: output,
                };
            }
            // If an output is older than the newest input, we can stop checking
            if (outputTime < newestInputFileTime) {
                return {
                    type: UpToDateStatusType.OutOfDateWithSelf,
                    outOfDateOutputFileName: output,
                    newerInputFileName: newestInputFileName,
                };
            }
            // No need to get newestDeclarationFileContentChangedTime since thats needed only for composite projects
            // And composite projects are the only ones that can be referenced
            if (outputTime < oldestOutputFileTime) {
                oldestOutputFileTime = outputTime;
                oldestOutputFileName = output;
            }
        }
    }
    /** Inputs are up-to-date, just need either timestamp update to make it look up-to-date */
    let pseudoUpToDate = false;
    if (referenceStatuses) {
        for (const { ref, refStatus, resolvedConfig, resolvedRefPath } of referenceStatuses) {
            // If the upstream project's newest file is older than our oldest output, we
            // can't be out of date because of it
            if (refStatus.newestInputFileTime && refStatus.newestInputFileTime <= oldestOutputFileTime) {
                continue;
            }
            // Check if tsbuildinfo path is shared, then we need to rebuild
            if (hasSameBuildInfo(state, buildInfoCacheEntry ??= state.buildInfoCache.get(resolvedPath), resolvedRefPath)) {
                return {
                    type: UpToDateStatusType.OutOfDateWithUpstream,
                    outOfDateOutputFileName: buildInfoPath,
                    newerProjectName: ref.path,
                };
            }
            // If the upstream project has only change .d.ts files, and we've built
            // *after* those files, then we're "psuedo up to date" and eligible for a fast rebuild
            const newestDeclarationFileContentChangedTime = getLatestChangedDtsTime(state, resolvedConfig.options, resolvedRefPath);
            if (newestDeclarationFileContentChangedTime && newestDeclarationFileContentChangedTime <= oldestOutputFileTime) {
                pseudoUpToDate = true;
                continue;
            }
            // We have an output older than an upstream output - we are out of date
            Debug.assert(oldestOutputFileName !== undefined, "Should have an oldest output filename here");
            return {
                type: UpToDateStatusType.OutOfDateWithUpstream,
                outOfDateOutputFileName: oldestOutputFileName,
                newerProjectName: ref.path,
            };
        }
    }
    // Check tsconfig time
    const configStatus = checkConfigFileUpToDateStatus(state, project.options.configFilePath, oldestOutputFileTime, oldestOutputFileName);
    if (configStatus)
        return configStatus;
    // Check extended config time
    const extendedConfigStatus = forEach(project.options.configFile.extendedSourceFiles || emptyArray, configFile => checkConfigFileUpToDateStatus(state, configFile, oldestOutputFileTime, oldestOutputFileName));
    if (extendedConfigStatus)
        return extendedConfigStatus;
    // Check package file time
    const packageJsonLookups = state.lastCachedPackageJsonLookups.get(resolvedPath);
    const dependentPackageFileStatus = packageJsonLookups && forEachKey(packageJsonLookups, path => checkConfigFileUpToDateStatus(state, path, oldestOutputFileTime, oldestOutputFileName));
    if (dependentPackageFileStatus)
        return dependentPackageFileStatus;
    // Up to date
    return {
        type: pseudoUpToDate ?
            UpToDateStatusType.UpToDateWithUpstreamTypes :
            pseudoInputUpToDate ?
                UpToDateStatusType.UpToDateWithInputFileText :
                UpToDateStatusType.UpToDate,
        newestInputFileTime,
        newestInputFileName,
        oldestOutputFileName,
    };
}
function hasSameBuildInfo(state, buildInfoCacheEntry, resolvedRefPath) {
    const refBuildInfo = state.buildInfoCache.get(resolvedRefPath);
    return refBuildInfo.path === buildInfoCacheEntry.path;
}
function getUpToDateStatus(state, project, resolvedPath) {
    if (project === undefined) {
        return { type: UpToDateStatusType.Unbuildable, reason: "config file deleted mid-build" };
    }
    const prior = state.projectStatus.get(resolvedPath);
    if (prior !== undefined) {
        return prior;
    }
    performance.mark("SolutionBuilder::beforeUpToDateCheck");
    const actual = getUpToDateStatusWorker(state, project, resolvedPath);
    performance.mark("SolutionBuilder::afterUpToDateCheck");
    performance.measure("SolutionBuilder::Up-to-date check", "SolutionBuilder::beforeUpToDateCheck", "SolutionBuilder::afterUpToDateCheck");
    state.projectStatus.set(resolvedPath, actual);
    return actual;
}
function updateOutputTimestampsWorker(state, proj, projectPath, verboseMessage, skipOutputs) {
    if (proj.options.noEmit)
        return;
    let now;
    const buildInfoPath = getTsBuildInfoEmitOutputFilePath(proj.options);
    const isIncremental = isIncrementalCompilation(proj.options);
    if (buildInfoPath && isIncremental) {
        // For incremental projects, only buildinfo needs to be upto date with timestamp check
        // as we dont check output files for up-to-date ness
        if (!skipOutputs?.has(toPath(state, buildInfoPath))) {
            if (!!state.options.verbose)
                reportStatus(state, verboseMessage, proj.options.configFilePath);
            state.host.setModifiedTime(buildInfoPath, now = getCurrentTime(state.host));
            getBuildInfoCacheEntry(state, buildInfoPath, projectPath).modifiedTime = now;
        }
        state.outputTimeStamps.delete(projectPath);
        return;
    }
    const { host } = state;
    const outputs = getAllProjectOutputs(proj, !host.useCaseSensitiveFileNames());
    const outputTimeStampMap = getOutputTimeStampMap(state, projectPath);
    const modifiedOutputs = outputTimeStampMap ? new Set() : undefined;
    if (!skipOutputs || outputs.length !== skipOutputs.size) {
        let reportVerbose = !!state.options.verbose;
        for (const file of outputs) {
            const path = toPath(state, file);
            if (skipOutputs?.has(path))
                continue;
            if (reportVerbose) {
                reportVerbose = false;
                reportStatus(state, verboseMessage, proj.options.configFilePath);
            }
            host.setModifiedTime(file, now ||= getCurrentTime(state.host));
            if (file === buildInfoPath)
                getBuildInfoCacheEntry(state, buildInfoPath, projectPath).modifiedTime = now;
            // Store output timestamps in a map because non incremental build will need to check them to determine up-to-dateness
            else if (outputTimeStampMap) {
                outputTimeStampMap.set(path, now);
                modifiedOutputs.add(path);
            }
        }
    }
    // Clear out timestamps not in output list any more
    outputTimeStampMap?.forEach((_value, key) => {
        if (!skipOutputs?.has(key) && !modifiedOutputs.has(key))
            outputTimeStampMap.delete(key);
    });
}
function getLatestChangedDtsTime(state, options, resolvedConfigPath) {
    if (!options.composite)
        return undefined;
    const entry = Debug.checkDefined(state.buildInfoCache.get(resolvedConfigPath));
    if (entry.latestChangedDtsTime !== undefined)
        return entry.latestChangedDtsTime || undefined;
    const latestChangedDtsTime = entry.buildInfo && isIncrementalBuildInfo(entry.buildInfo) && entry.buildInfo.latestChangedDtsFile ?
        state.host.getModifiedTime(getNormalizedAbsolutePath(entry.buildInfo.latestChangedDtsFile, getDirectoryPath(entry.path))) :
        undefined;
    entry.latestChangedDtsTime = latestChangedDtsTime || false;
    return latestChangedDtsTime;
}
function updateOutputTimestamps(state, proj, resolvedPath) {
    if (state.options.dry) {
        return reportStatus(state, Diagnostics.A_non_dry_build_would_update_timestamps_for_output_of_project_0, proj.options.configFilePath);
    }
    updateOutputTimestampsWorker(state, proj, resolvedPath, Diagnostics.Updating_output_timestamps_of_project_0);
    state.projectStatus.set(resolvedPath, {
        type: UpToDateStatusType.UpToDate,
        oldestOutputFileName: getFirstProjectOutput(proj, !state.host.useCaseSensitiveFileNames()),
    });
}
function queueReferencingProjects(state, project, projectPath, projectIndex, config, buildOrder, buildResult) {
    // Queue only if there are no errors
    if (state.options.stopBuildOnErrors && (buildResult & BuildResultFlags.AnyErrors))
        return;
    // Only composite projects can be referenced by other projects
    if (!config.options.composite)
        return;
    // Always use build order to queue projects
    for (let index = projectIndex + 1; index < buildOrder.length; index++) {
        const nextProject = buildOrder[index];
        const nextProjectPath = toResolvedConfigFilePath(state, nextProject);
        if (state.projectPendingBuild.has(nextProjectPath))
            continue;
        const nextProjectConfig = parseConfigFile(state, nextProject, nextProjectPath);
        if (!nextProjectConfig || !nextProjectConfig.projectReferences)
            continue;
        for (const ref of nextProjectConfig.projectReferences) {
            const resolvedRefPath = resolveProjectName(state, ref.path);
            if (toResolvedConfigFilePath(state, resolvedRefPath) !== projectPath)
                continue;
            // If declaration output is changed, build the project
            // otherwise mark the project UpToDateWithUpstreamTypes so it updates output time stamps
            const status = state.projectStatus.get(nextProjectPath);
            if (status) {
                switch (status.type) {
                    case UpToDateStatusType.UpToDate:
                        if (buildResult & BuildResultFlags.DeclarationOutputUnchanged) {
                            status.type = UpToDateStatusType.UpToDateWithUpstreamTypes;
                            break;
                        }
                    // falls through
                    case UpToDateStatusType.UpToDateWithInputFileText:
                    case UpToDateStatusType.UpToDateWithUpstreamTypes:
                        if (!(buildResult & BuildResultFlags.DeclarationOutputUnchanged)) {
                            state.projectStatus.set(nextProjectPath, {
                                type: UpToDateStatusType.OutOfDateWithUpstream,
                                outOfDateOutputFileName: status.oldestOutputFileName,
                                newerProjectName: project,
                            });
                        }
                        break;
                    case UpToDateStatusType.UpstreamBlocked:
                        if (toResolvedConfigFilePath(state, resolveProjectName(state, status.upstreamProjectName)) === projectPath) {
                            clearProjectStatus(state, nextProjectPath);
                        }
                        break;
                }
            }
            addProjToQueue(state, nextProjectPath, ProgramUpdateLevel.Update);
            break;
        }
    }
}
function build(state, project, cancellationToken, writeFile, getCustomTransformers, onlyReferences) {
    performance.mark("SolutionBuilder::beforeBuild");
    const result = buildWorker(state, project, cancellationToken, writeFile, getCustomTransformers, onlyReferences);
    performance.mark("SolutionBuilder::afterBuild");
    performance.measure("SolutionBuilder::Build", "SolutionBuilder::beforeBuild", "SolutionBuilder::afterBuild");
    return result;
}
function buildWorker(state, project, cancellationToken, writeFile, getCustomTransformers, onlyReferences) {
    const buildOrder = getBuildOrderFor(state, project, onlyReferences);
    if (!buildOrder)
        return ExitStatus.InvalidProject_OutputsSkipped;
    setupInitialBuild(state, cancellationToken);
    let reportQueue = true;
    let successfulProjects = 0;
    while (true) {
        const invalidatedProject = getNextInvalidatedProject(state, buildOrder, reportQueue);
        if (!invalidatedProject)
            break;
        reportQueue = false;
        invalidatedProject.done(cancellationToken, writeFile, getCustomTransformers?.(invalidatedProject.project));
        if (!state.diagnostics.has(invalidatedProject.projectPath))
            successfulProjects++;
    }
    disableCache(state);
    reportErrorSummary(state, buildOrder);
    startWatching(state, buildOrder);
    return isCircularBuildOrder(buildOrder)
        ? ExitStatus.ProjectReferenceCycle_OutputsSkipped
        : !buildOrder.some(p => state.diagnostics.has(toResolvedConfigFilePath(state, p)))
            ? ExitStatus.Success
            : successfulProjects
                ? ExitStatus.DiagnosticsPresent_OutputsGenerated
                : ExitStatus.DiagnosticsPresent_OutputsSkipped;
}
function clean(state, project, onlyReferences) {
    performance.mark("SolutionBuilder::beforeClean");
    const result = cleanWorker(state, project, onlyReferences);
    performance.mark("SolutionBuilder::afterClean");
    performance.measure("SolutionBuilder::Clean", "SolutionBuilder::beforeClean", "SolutionBuilder::afterClean");
    return result;
}
function cleanWorker(state, project, onlyReferences) {
    const buildOrder = getBuildOrderFor(state, project, onlyReferences);
    if (!buildOrder)
        return ExitStatus.InvalidProject_OutputsSkipped;
    if (isCircularBuildOrder(buildOrder)) {
        reportErrors(state, buildOrder.circularDiagnostics);
        return ExitStatus.ProjectReferenceCycle_OutputsSkipped;
    }
    const { options, host } = state;
    const filesToDelete = options.dry ? [] : undefined;
    for (const proj of buildOrder) {
        const resolvedPath = toResolvedConfigFilePath(state, proj);
        const parsed = parseConfigFile(state, proj, resolvedPath);
        if (parsed === undefined) {
            // File has gone missing; fine to ignore here
            reportParseConfigFileDiagnostic(state, resolvedPath);
            continue;
        }
        const outputs = getAllProjectOutputs(parsed, !host.useCaseSensitiveFileNames());
        if (!outputs.length)
            continue;
        const inputFileNames = new Set(parsed.fileNames.map(f => toPath(state, f)));
        for (const output of outputs) {
            // If output name is same as input file name, do not delete and ignore the error
            if (inputFileNames.has(toPath(state, output)))
                continue;
            if (host.fileExists(output)) {
                if (filesToDelete) {
                    filesToDelete.push(output);
                }
                else {
                    host.deleteFile(output);
                    invalidateProject(state, resolvedPath, ProgramUpdateLevel.Update);
                }
            }
        }
    }
    if (filesToDelete) {
        reportStatus(state, Diagnostics.A_non_dry_build_would_delete_the_following_files_Colon_0, filesToDelete.map(f => `\r\n * ${f}`).join(""));
    }
    return ExitStatus.Success;
}
function invalidateProject(state, resolved, updateLevel) {
    // If host implements getParsedCommandLine, we cant get list of files from parseConfigFileHost
    if (state.host.getParsedCommandLine && updateLevel === ProgramUpdateLevel.RootNamesAndUpdate) {
        updateLevel = ProgramUpdateLevel.Full;
    }
    if (updateLevel === ProgramUpdateLevel.Full) {
        state.configFileCache.delete(resolved);
        state.buildOrder = undefined;
    }
    state.needsSummary = true;
    clearProjectStatus(state, resolved);
    addProjToQueue(state, resolved, updateLevel);
    enableCache(state);
}
function invalidateProjectAndScheduleBuilds(state, resolvedPath, updateLevel) {
    state.reportFileChangeDetected = true;
    invalidateProject(state, resolvedPath, updateLevel);
    scheduleBuildInvalidatedProject(state, 250, /*changeDetected*/ true);
}
function scheduleBuildInvalidatedProject(state, time, changeDetected) {
    const { hostWithWatch } = state;
    if (!hostWithWatch.setTimeout || !hostWithWatch.clearTimeout) {
        return;
    }
    if (state.timerToBuildInvalidatedProject) {
        hostWithWatch.clearTimeout(state.timerToBuildInvalidatedProject);
    }
    state.timerToBuildInvalidatedProject = (() => {
        const timerHost = hostWithWatch;
        const { setTimeout } = timerHost;
        const callback = buildNextInvalidatedProject;
        const timeoutDelay = time;
        const t1 = "timerToBuildInvalidatedProject";
        const t2 = state;
        const t3 = changeDetected;
        return setTimeout.call(timerHost, () => callback(t1, t2, t3), timeoutDelay, t1);
    })();
}
function buildNextInvalidatedProject(_timeoutType, state, changeDetected) {
    performance.mark("SolutionBuilder::beforeBuild");
    const buildOrder = buildNextInvalidatedProjectWorker(state, changeDetected);
    performance.mark("SolutionBuilder::afterBuild");
    performance.measure("SolutionBuilder::Build", "SolutionBuilder::beforeBuild", "SolutionBuilder::afterBuild");
    if (buildOrder)
        reportErrorSummary(state, buildOrder);
}
function buildNextInvalidatedProjectWorker(state, changeDetected) {
    state.timerToBuildInvalidatedProject = undefined;
    if (state.reportFileChangeDetected) {
        state.reportFileChangeDetected = false;
        state.projectErrorsReported.clear();
        reportWatchStatus(state, Diagnostics.File_change_detected_Starting_incremental_compilation);
    }
    let projectsBuilt = 0;
    const buildOrder = getBuildOrder(state);
    const invalidatedProject = getNextInvalidatedProject(state, buildOrder, /*reportQueue*/ false);
    if (invalidatedProject) {
        invalidatedProject.done();
        projectsBuilt++;
        while (state.projectPendingBuild.size) {
            // If already scheduled, skip
            if (state.timerToBuildInvalidatedProject)
                return;
            // Before scheduling check if the next project needs build
            const info = getNextInvalidatedProjectCreateInfo(state, buildOrder, /*reportQueue*/ false);
            if (!info)
                break; // Nothing to build any more
            if (info.kind !== InvalidatedProjectKind.UpdateOutputFileStamps && (changeDetected || projectsBuilt === 5)) {
                // Schedule next project for build
                scheduleBuildInvalidatedProject(state, 100, /*changeDetected*/ false);
                return;
            }
            const project = createInvalidatedProjectWithInfo(state, info, buildOrder);
            project.done();
            if (info.kind !== InvalidatedProjectKind.UpdateOutputFileStamps)
                projectsBuilt++;
        }
    }
    disableCache(state);
    return buildOrder;
}
function watchConfigFile(state, resolved, resolvedPath, parsed) {
    if (!state.watch || state.allWatchedConfigFiles.has(resolvedPath))
        return;
    state.allWatchedConfigFiles.set(resolvedPath, watchFile(state, resolved, () => invalidateProjectAndScheduleBuilds(state, resolvedPath, ProgramUpdateLevel.Full), PollingInterval.High, parsed?.watchOptions, WatchType.ConfigFile, resolved));
}
function watchExtendedConfigFiles(state, resolvedPath, parsed) {
    updateSharedExtendedConfigFileWatcher(resolvedPath, parsed?.options, state.allWatchedExtendedConfigFiles, (extendedConfigFileName, extendedConfigFilePath) => watchFile(state, extendedConfigFileName, () => state.allWatchedExtendedConfigFiles.get(extendedConfigFilePath)?.projects.forEach(projectConfigFilePath => invalidateProjectAndScheduleBuilds(state, projectConfigFilePath, ProgramUpdateLevel.Full)), PollingInterval.High, parsed?.watchOptions, WatchType.ExtendedConfigFile), fileName => toPath(state, fileName));
}
function watchWildCardDirectories(state, resolved, resolvedPath, parsed) {
    if (!state.watch)
        return;
    updateWatchingWildcardDirectories(getOrCreateValueMapFromConfigFileMap(state.allWatchedWildcardDirectories, resolvedPath), parsed.wildcardDirectories, (dir, flags) => state.watchDirectory(dir, fileOrDirectory => {
        if (isIgnoredFileFromWildCardWatching({
            watchedDirPath: toPath(state, dir),
            fileOrDirectory,
            fileOrDirectoryPath: toPath(state, fileOrDirectory),
            configFileName: resolved,
            currentDirectory: state.compilerHost.getCurrentDirectory(),
            options: parsed.options,
            program: state.builderPrograms.get(resolvedPath) || getCachedParsedConfigFile(state, resolvedPath)?.fileNames,
            useCaseSensitiveFileNames: state.parseConfigFileHost.useCaseSensitiveFileNames,
            writeLog: s => state.writeLog(s),
            toPath: fileName => toPath(state, fileName),
        }))
            return;
        invalidateProjectAndScheduleBuilds(state, resolvedPath, ProgramUpdateLevel.RootNamesAndUpdate);
    }, flags, parsed?.watchOptions, WatchType.WildcardDirectory, resolved));
}
function watchInputFiles(state, resolved, resolvedPath, parsed) {
    if (!state.watch)
        return;
    mutateMap(getOrCreateValueMapFromConfigFileMap(state.allWatchedInputFiles, resolvedPath), new Set(parsed.fileNames), {
        createNewValue: input => watchFile(state, input, () => invalidateProjectAndScheduleBuilds(state, resolvedPath, ProgramUpdateLevel.Update), PollingInterval.Low, parsed?.watchOptions, WatchType.SourceFile, resolved),
        onDeleteValue: closeFileWatcher,
    });
}
function watchPackageJsonFiles(state, resolved, resolvedPath, parsed) {
    if (!state.watch || !state.lastCachedPackageJsonLookups)
        return;
    mutateMap(getOrCreateValueMapFromConfigFileMap(state.allWatchedPackageJsonFiles, resolvedPath), state.lastCachedPackageJsonLookups.get(resolvedPath), {
        createNewValue: input => watchFile(state, input, () => invalidateProjectAndScheduleBuilds(state, resolvedPath, ProgramUpdateLevel.Update), PollingInterval.High, parsed?.watchOptions, WatchType.PackageJson, resolved),
        onDeleteValue: closeFileWatcher,
    });
}
function startWatching(state, buildOrder) {
    if (!state.watchAllProjectsPending)
        return;
    performance.mark("SolutionBuilder::beforeWatcherCreation");
    state.watchAllProjectsPending = false;
    for (const resolved of getBuildOrderFromAnyBuildOrder(buildOrder)) {
        const resolvedPath = toResolvedConfigFilePath(state, resolved);
        const cfg = parseConfigFile(state, resolved, resolvedPath);
        // Watch this file
        watchConfigFile(state, resolved, resolvedPath, cfg);
        watchExtendedConfigFiles(state, resolvedPath, cfg);
        if (cfg) {
            // Update watchers for wildcard directories
            watchWildCardDirectories(state, resolved, resolvedPath, cfg);
            // Watch input files
            watchInputFiles(state, resolved, resolvedPath, cfg);
            // Watch package json files
            watchPackageJsonFiles(state, resolved, resolvedPath, cfg);
        }
    }
    performance.mark("SolutionBuilder::afterWatcherCreation");
    performance.measure("SolutionBuilder::Watcher creation", "SolutionBuilder::beforeWatcherCreation", "SolutionBuilder::afterWatcherCreation");
}
function stopWatching(state) {
    clearMap(state.allWatchedConfigFiles, closeFileWatcher);
    clearMap(state.allWatchedExtendedConfigFiles, closeFileWatcherOf);
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
    clearMap(state.allWatchedInputFiles, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcher));
    clearMap(state.allWatchedPackageJsonFiles, watchedPacageJsonFiles => clearMap(watchedPacageJsonFiles, closeFileWatcher));
}
function createSolutionBuilderWorker(watch, hostOrHostWithWatch, rootNames, options, baseWatchOptions) {
    const state = createSolutionBuilderState(watch, hostOrHostWithWatch, rootNames, options, baseWatchOptions);
    return {
        build: (project, cancellationToken, writeFile, getCustomTransformers) => build(state, project, cancellationToken, writeFile, getCustomTransformers),
        clean: project => clean(state, project),
        buildReferences: (project, cancellationToken, writeFile, getCustomTransformers) => build(state, project, cancellationToken, writeFile, getCustomTransformers, /*onlyReferences*/ true),
        cleanReferences: project => clean(state, project, /*onlyReferences*/ true),
        getNextInvalidatedProject: cancellationToken => {
            setupInitialBuild(state, cancellationToken);
            return getNextInvalidatedProject(state, getBuildOrder(state), /*reportQueue*/ false);
        },
        getBuildOrder: () => getBuildOrder(state),
        getUpToDateStatusOfProject: project => {
            const configFileName = resolveProjectName(state, project);
            const configFilePath = toResolvedConfigFilePath(state, configFileName);
            return getUpToDateStatus(state, parseConfigFile(state, configFileName, configFilePath), configFilePath);
        },
        invalidateProject: (configFilePath, updateLevel) => invalidateProject(state, configFilePath, updateLevel || ProgramUpdateLevel.Update),
        close: () => stopWatching(state),
    };
}
function relName(state, path) {
    return convertToRelativePath(path, state.compilerHost.getCurrentDirectory(), state.compilerHost.getCanonicalFileName);
}
function reportStatus(state, message, ...args) {
    state.host.reportSolutionBuilderStatus(createCompilerDiagnostic(message, ...args));
}
function reportWatchStatus(state, message, ...args) {
    state.hostWithWatch.onWatchStatusChange?.(createCompilerDiagnostic(message, ...args), state.host.getNewLine(), state.baseCompilerOptions);
}
function reportErrors({ host }, errors) {
    errors.forEach(err => host.reportDiagnostic(err));
}
function reportAndStoreErrors(state, proj, errors) {
    reportErrors(state, errors);
    state.projectErrorsReported.set(proj, true);
    if (errors.length) {
        state.diagnostics.set(proj, errors);
    }
}
function reportParseConfigFileDiagnostic(state, proj) {
    reportAndStoreErrors(state, proj, [state.configFileCache.get(proj)]);
}
function reportErrorSummary(state, buildOrder) {
    if (!state.needsSummary)
        return;
    state.needsSummary = false;
    const canReportSummary = state.watch || !!state.host.reportErrorSummary;
    const { diagnostics } = state;
    let totalErrors = 0;
    let filesInError = [];
    if (isCircularBuildOrder(buildOrder)) {
        reportBuildQueue(state, buildOrder.buildOrder);
        reportErrors(state, buildOrder.circularDiagnostics);
        if (canReportSummary)
            totalErrors += getErrorCountForSummary(buildOrder.circularDiagnostics);
        if (canReportSummary)
            filesInError = [...filesInError, ...getFilesInErrorForSummary(buildOrder.circularDiagnostics)];
    }
    else {
        // Report errors from the other projects
        buildOrder.forEach(project => {
            const projectPath = toResolvedConfigFilePath(state, project);
            if (!state.projectErrorsReported.has(projectPath)) {
                reportErrors(state, diagnostics.get(projectPath) || emptyArray);
            }
        });
        if (canReportSummary)
            diagnostics.forEach(singleProjectErrors => totalErrors += getErrorCountForSummary(singleProjectErrors));
        if (canReportSummary)
            diagnostics.forEach(singleProjectErrors => [...filesInError, ...getFilesInErrorForSummary(singleProjectErrors)]);
    }
    if (state.watch) {
        reportWatchStatus(state, getWatchErrorSummaryDiagnosticMessage(totalErrors), totalErrors);
    }
    else if (state.host.reportErrorSummary) {
        state.host.reportErrorSummary(totalErrors, filesInError);
    }
}
/**
 * Report the build ordering inferred from the current project graph if we're in verbose mode
 */
function reportBuildQueue(state, buildQueue) {
    if (state.options.verbose) {
        reportStatus(state, Diagnostics.Projects_in_this_build_Colon_0, buildQueue.map(s => "\r\n    * " + relName(state, s)).join(""));
    }
}
function reportUpToDateStatus(state, configFileName, status) {
    switch (status.type) {
        case UpToDateStatusType.OutOfDateWithSelf:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_output_1_is_older_than_input_2, relName(state, configFileName), relName(state, status.outOfDateOutputFileName), relName(state, status.newerInputFileName));
        case UpToDateStatusType.OutOfDateWithUpstream:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_output_1_is_older_than_input_2, relName(state, configFileName), relName(state, status.outOfDateOutputFileName), relName(state, status.newerProjectName));
        case UpToDateStatusType.OutputMissing:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_output_file_1_does_not_exist, relName(state, configFileName), relName(state, status.missingOutputFileName));
        case UpToDateStatusType.ErrorReadingFile:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_there_was_error_reading_file_1, relName(state, configFileName), relName(state, status.fileName));
        case UpToDateStatusType.OutOfDateBuildInfoWithPendingEmit:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_buildinfo_file_1_indicates_that_some_of_the_changes_were_not_emitted, relName(state, configFileName), relName(state, status.buildInfoFile));
        case UpToDateStatusType.OutOfDateBuildInfoWithErrors:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_buildinfo_file_1_indicates_that_program_needs_to_report_errors, relName(state, configFileName), relName(state, status.buildInfoFile));
        case UpToDateStatusType.OutOfDateOptions:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_buildinfo_file_1_indicates_there_is_change_in_compilerOptions, relName(state, configFileName), relName(state, status.buildInfoFile));
        case UpToDateStatusType.OutOfDateRoots:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_buildinfo_file_1_indicates_that_file_2_was_root_file_of_compilation_but_not_any_more, relName(state, configFileName), relName(state, status.buildInfoFile), relName(state, status.inputFile));
        case UpToDateStatusType.UpToDate:
            if (status.newestInputFileTime !== undefined) {
                return reportStatus(state, Diagnostics.Project_0_is_up_to_date_because_newest_input_1_is_older_than_output_2, relName(state, configFileName), relName(state, status.newestInputFileName || ""), relName(state, status.oldestOutputFileName || ""));
            }
            // Don't report anything for "up to date because it was already built" -- too verbose
            break;
        case UpToDateStatusType.UpToDateWithUpstreamTypes:
            return reportStatus(state, Diagnostics.Project_0_is_up_to_date_with_d_ts_files_from_its_dependencies, relName(state, configFileName));
        case UpToDateStatusType.UpToDateWithInputFileText:
            return reportStatus(state, Diagnostics.Project_0_is_up_to_date_but_needs_to_update_timestamps_of_output_files_that_are_older_than_input_files, relName(state, configFileName));
        case UpToDateStatusType.UpstreamOutOfDate:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_its_dependency_1_is_out_of_date, relName(state, configFileName), relName(state, status.upstreamProjectName));
        case UpToDateStatusType.UpstreamBlocked:
            return reportStatus(state, status.upstreamProjectBlocked ?
                Diagnostics.Project_0_can_t_be_built_because_its_dependency_1_was_not_built :
                Diagnostics.Project_0_can_t_be_built_because_its_dependency_1_has_errors, relName(state, configFileName), relName(state, status.upstreamProjectName));
        case UpToDateStatusType.Unbuildable:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_1, relName(state, configFileName), status.reason);
        case UpToDateStatusType.TsVersionOutputOfDate:
            return reportStatus(state, Diagnostics.Project_0_is_out_of_date_because_output_for_it_was_generated_with_version_1_that_differs_with_current_version_2, relName(state, configFileName), status.version, version);
        case UpToDateStatusType.ForceBuild:
            return reportStatus(state, Diagnostics.Project_0_is_being_forcibly_rebuilt, relName(state, configFileName));
        case UpToDateStatusType.ContainerOnly:
        // Don't report status on "solution" projects
        // falls through
        case UpToDateStatusType.ComputingUpstream:
            // Should never leak from getUptoDateStatusWorker
            break;
        default:
            assertType(status);
    }
}
/**
 * Report the up-to-date status of a project if we're in verbose mode
 */
function verboseReportProjectStatus(state, configFileName, status) {
    if (state.options.verbose) {
        reportUpToDateStatus(state, configFileName, status);
    }
}
