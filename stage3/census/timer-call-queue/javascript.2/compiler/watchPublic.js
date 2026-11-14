import { canJsonReportNoInputFiles, changeCompilerHostLikeToUseCache, changesAffectModuleResolution, cleanExtendedConfigCache, clearMap, clearSharedExtendedConfigFileWatcher, closeFileWatcher, closeFileWatcherOf, createBuilderProgramUsingIncrementalBuildInfo, createCachedDirectoryStructureHost, createCompilerDiagnostic, createCompilerHostFromProgramHost, createCompilerHostWorker, createEmitAndSemanticDiagnosticsBuilderProgram, createGetCanonicalFileName, createResolutionCache, createWatchCompilerHostOfConfigFile, createWatchCompilerHostOfFilesAndCompilerOptions, createWatchFactory, Debug, Diagnostics, FileWatcherEventKind, getBuildInfo, getConfigFileParsingDiagnostics, getDirectoryPath, getFileNamesFromConfigSpecs, getNewLineCharacter, getNormalizedAbsolutePath, getParsedCommandLineOfConfigFile, getSourceFileVersionAsHashFromText, getTsBuildInfoEmitOutputFilePath, isArray, isIgnoredFileFromWildCardWatching, isIncrementalBuildInfo, isProgramUptoDate, noop, noopFileWatcher, parseConfigHostFromCompilerHostLike, PollingInterval, ProgramUpdateLevel, returnFalse, returnTrue, setGetSourceFileAsHashVersioned, sys, toPath, toPath as ts_toPath, updateErrorForNoInputFiles, updateMissingFilePathsWatch, updateSharedExtendedConfigFileWatcher, updateWatchingWildcardDirectories, version, WatchType, } from "./_namespaces/ts.js";
export function readBuilderProgram(compilerOptions, host) {
    const buildInfoPath = getTsBuildInfoEmitOutputFilePath(compilerOptions);
    if (!buildInfoPath)
        return undefined;
    let buildInfo;
    if (host.getBuildInfo) {
        // host provides buildinfo, get it from there. This allows host to cache it
        buildInfo = host.getBuildInfo(buildInfoPath, compilerOptions.configFilePath);
    }
    else {
        const content = host.readFile(buildInfoPath);
        if (!content)
            return undefined;
        buildInfo = getBuildInfo(buildInfoPath, content);
    }
    if (!buildInfo || buildInfo.version !== version || !isIncrementalBuildInfo(buildInfo))
        return undefined;
    return createBuilderProgramUsingIncrementalBuildInfo(buildInfo, buildInfoPath, host);
}
export function createIncrementalCompilerHost(options, system = sys) {
    const host = createCompilerHostWorker(options, /*setParentNodes*/ undefined, system);
    host.createHash = system.createHash?.bind(system);
    host.storeSignatureInfo = system.storeSignatureInfo;
    setGetSourceFileAsHashVersioned(host);
    changeCompilerHostLikeToUseCache(host, fileName => toPath(fileName, host.getCurrentDirectory(), host.getCanonicalFileName));
    return host;
}
export function createIncrementalProgram({ rootNames, options, configFileParsingDiagnostics, projectReferences, host, createProgram, }) {
    host = host || createIncrementalCompilerHost(options);
    createProgram = createProgram || createEmitAndSemanticDiagnosticsBuilderProgram;
    const oldProgram = readBuilderProgram(options, host);
    return createProgram(rootNames, options, host, oldProgram, configFileParsingDiagnostics, projectReferences);
}
export function createWatchCompilerHost(rootFilesOrConfigFileName, options, system, createProgram, reportDiagnostic, reportWatchStatus, projectReferencesOrWatchOptionsToExtend, watchOptionsOrExtraFileExtensions) {
    if (isArray(rootFilesOrConfigFileName)) {
        return createWatchCompilerHostOfFilesAndCompilerOptions({
            rootFiles: rootFilesOrConfigFileName,
            options: options,
            watchOptions: watchOptionsOrExtraFileExtensions,
            projectReferences: projectReferencesOrWatchOptionsToExtend,
            system,
            createProgram,
            reportDiagnostic,
            reportWatchStatus,
        });
    }
    else {
        return createWatchCompilerHostOfConfigFile({
            configFileName: rootFilesOrConfigFileName,
            optionsToExtend: options,
            watchOptionsToExtend: projectReferencesOrWatchOptionsToExtend,
            extraFileExtensions: watchOptionsOrExtraFileExtensions,
            system,
            createProgram,
            reportDiagnostic,
            reportWatchStatus,
        });
    }
}
export function createWatchProgram(host) {
    let builderProgram;
    let updateLevel; // level to indicate if the program needs to be reloaded from config file/just filenames etc
    let missingFilesMap; // Map of file watchers for the missing files
    let watchedWildcardDirectories; // map of watchers for the wild card directories in the config file
    /**
     * undefined - own watches are stale,
     * path - for referenced project which need to be watched
     */
    let staleWatches = new Map([[undefined, undefined]]);
    let timerToUpdateProgram; // timer callback to recompile the program
    let timerToInvalidateFailedLookupResolutions; // timer callback to invalidate resolutions for changes in failed lookup locations
    let parsedConfigs; // Parsed commandline and watching cached for referenced projects
    let sharedExtendedConfigFileWatchers; // Map of file watchers for extended files, shared between different referenced projects
    let extendedConfigCache = host.extendedConfigCache; // Cache for extended config evaluation
    let reportFileChangeDetectedOnCreateProgram = false; // True if synchronizeProgram should report "File change detected..." when a new program is created
    const sourceFilesCache = new Map(); // Cache that stores the source file and version info
    let missingFilePathsRequestedForRelease; // These paths are held temporarily so that we can remove the entry from source file cache if the file is not tracked by missing files
    let hasChangedCompilerOptions = false; // True if the compiler options have changed between compilations
    const useCaseSensitiveFileNames = host.useCaseSensitiveFileNames();
    const currentDirectory = host.getCurrentDirectory();
    const { configFileName, optionsToExtend: optionsToExtendForConfigFile = {}, watchOptionsToExtend, extraFileExtensions, createProgram } = host;
    let { rootFiles: rootFileNames, options: compilerOptions, watchOptions, projectReferences } = host;
    let wildcardDirectories;
    let configFileParsingDiagnostics;
    let canConfigFileJsonReportNoInputFiles = false;
    let hasChangedConfigFileParsingErrors = false;
    const cachedDirectoryStructureHost = configFileName === undefined ? undefined : createCachedDirectoryStructureHost(host, currentDirectory, useCaseSensitiveFileNames);
    const directoryStructureHost = cachedDirectoryStructureHost || host;
    const parseConfigFileHost = parseConfigHostFromCompilerHostLike(host, directoryStructureHost);
    // From tsc we want to get already parsed result and hence check for rootFileNames
    let newLine = updateNewLine();
    if (configFileName && host.configFileParsingResult) {
        setConfigFileParsingResult(host.configFileParsingResult);
        newLine = updateNewLine();
    }
    reportWatchDiagnostic(Diagnostics.Starting_compilation_in_watch_mode);
    if (configFileName && !host.configFileParsingResult) {
        newLine = getNewLineCharacter(optionsToExtendForConfigFile);
        Debug.assert(!rootFileNames);
        parseConfigFile();
        newLine = updateNewLine();
    }
    Debug.assert(compilerOptions);
    Debug.assert(rootFileNames);
    const { watchFile, watchDirectory, writeLog } = createWatchFactory(host, compilerOptions);
    const getCanonicalFileName = createGetCanonicalFileName(useCaseSensitiveFileNames);
    writeLog(`Current directory: ${currentDirectory} CaseSensitiveFileNames: ${useCaseSensitiveFileNames}`);
    let configFileWatcher;
    if (configFileName) {
        configFileWatcher = watchFile(configFileName, scheduleProgramReload, PollingInterval.High, watchOptions, WatchType.ConfigFile);
    }
    const compilerHost = createCompilerHostFromProgramHost(host, () => compilerOptions, directoryStructureHost);
    setGetSourceFileAsHashVersioned(compilerHost);
    // Members for CompilerHost
    const getNewSourceFile = compilerHost.getSourceFile;
    compilerHost.getSourceFile = (fileName, ...args) => getVersionedSourceFileByPath(fileName, toPath(fileName), ...args);
    compilerHost.getSourceFileByPath = getVersionedSourceFileByPath;
    compilerHost.getNewLine = () => newLine;
    compilerHost.fileExists = fileExists;
    compilerHost.onReleaseOldSourceFile = onReleaseOldSourceFile;
    compilerHost.onReleaseParsedCommandLine = onReleaseParsedCommandLine;
    // Members for ResolutionCacheHost
    compilerHost.toPath = toPath;
    compilerHost.getCompilationSettings = () => compilerOptions;
    compilerHost.useSourceOfProjectReferenceRedirect = host.useSourceOfProjectReferenceRedirect?.bind(host);
    compilerHost.preferNonRecursiveWatch = host.preferNonRecursiveWatch;
    compilerHost.watchDirectoryOfFailedLookupLocation = (dir, cb, flags) => watchDirectory(dir, cb, flags, watchOptions, WatchType.FailedLookupLocations);
    compilerHost.watchAffectingFileLocation = (file, cb) => watchFile(file, cb, PollingInterval.High, watchOptions, WatchType.AffectingFileLocation);
    compilerHost.watchTypeRootsDirectory = (dir, cb, flags) => watchDirectory(dir, cb, flags, watchOptions, WatchType.TypeRoots);
    compilerHost.getCachedDirectoryStructureHost = () => cachedDirectoryStructureHost;
    compilerHost.scheduleInvalidateResolutionsOfFailedLookupLocations = scheduleInvalidateResolutionsOfFailedLookupLocations;
    compilerHost.onInvalidatedResolution = scheduleProgramUpdate;
    compilerHost.onChangedAutomaticTypeDirectiveNames = scheduleProgramUpdate;
    compilerHost.fileIsOpen = returnFalse;
    compilerHost.getCurrentProgram = getCurrentProgram;
    compilerHost.writeLog = writeLog;
    compilerHost.getParsedCommandLine = getParsedCommandLine;
    // Cache for the module resolution
    const resolutionCache = createResolutionCache(compilerHost, configFileName ?
        getDirectoryPath(getNormalizedAbsolutePath(configFileName, currentDirectory)) :
        currentDirectory, 
    /*logChangesWhenResolvingModule*/ false);
    // Resolve module using host module resolution strategy if provided otherwise use resolution cache to resolve module names
    compilerHost.resolveModuleNameLiterals = host.resolveModuleNameLiterals?.bind(host);
    compilerHost.resolveModuleNames = host.resolveModuleNames?.bind(host);
    if (!compilerHost.resolveModuleNameLiterals && !compilerHost.resolveModuleNames) {
        compilerHost.resolveModuleNameLiterals = resolutionCache.resolveModuleNameLiterals.bind(resolutionCache);
    }
    compilerHost.resolveTypeReferenceDirectiveReferences = host.resolveTypeReferenceDirectiveReferences?.bind(host);
    compilerHost.resolveTypeReferenceDirectives = host.resolveTypeReferenceDirectives?.bind(host);
    if (!compilerHost.resolveTypeReferenceDirectiveReferences && !compilerHost.resolveTypeReferenceDirectives) {
        compilerHost.resolveTypeReferenceDirectiveReferences = resolutionCache.resolveTypeReferenceDirectiveReferences.bind(resolutionCache);
    }
    compilerHost.resolveLibrary = !host.resolveLibrary ?
        resolutionCache.resolveLibrary.bind(resolutionCache) :
        host.resolveLibrary.bind(host);
    compilerHost.getModuleResolutionCache = host.resolveModuleNameLiterals || host.resolveModuleNames ?
        host.getModuleResolutionCache?.bind(host) :
        (() => resolutionCache.getModuleResolutionCache());
    const userProvidedResolution = !!host.resolveModuleNameLiterals || !!host.resolveTypeReferenceDirectiveReferences ||
        !!host.resolveModuleNames || !!host.resolveTypeReferenceDirectives;
    // All resolutions are invalid if user provided resolutions and didnt supply hasInvalidatedResolutions
    const customHasInvalidatedResolutions = userProvidedResolution ?
        host.hasInvalidatedResolutions?.bind(host) || returnTrue :
        returnFalse;
    const customHasInvalidLibResolutions = host.resolveLibrary ?
        host.hasInvalidatedLibResolutions?.bind(host) || returnTrue :
        returnFalse;
    builderProgram = readBuilderProgram(compilerOptions, compilerHost);
    synchronizeProgram();
    return configFileName ?
        { getCurrentProgram: getCurrentBuilderProgram, getProgram: updateProgram, close, getResolutionCache } :
        { getCurrentProgram: getCurrentBuilderProgram, getProgram: updateProgram, updateRootFileNames, close, getResolutionCache };
    function close() {
        clearInvalidateResolutionsOfFailedLookupLocations();
        resolutionCache.clear();
        clearMap(sourceFilesCache, value => {
            if (value && value.fileWatcher) {
                value.fileWatcher.close();
                if (isFilePresenceUnknownOnHost(value))
                    value.fileWatcher = undefined;
            }
        });
        if (configFileWatcher) {
            configFileWatcher.close();
            configFileWatcher = undefined;
        }
        extendedConfigCache?.clear();
        extendedConfigCache = undefined;
        if (sharedExtendedConfigFileWatchers) {
            clearMap(sharedExtendedConfigFileWatchers, closeFileWatcherOf);
            sharedExtendedConfigFileWatchers = undefined;
        }
        if (watchedWildcardDirectories) {
            clearMap(watchedWildcardDirectories, closeFileWatcherOf);
            watchedWildcardDirectories = undefined;
        }
        if (missingFilesMap) {
            clearMap(missingFilesMap, closeFileWatcher);
            missingFilesMap = undefined;
        }
        if (parsedConfigs) {
            clearMap(parsedConfigs, config => {
                config.watcher?.close();
                config.watcher = undefined;
                if (config.watchedDirectories)
                    clearMap(config.watchedDirectories, closeFileWatcherOf);
                config.watchedDirectories = undefined;
            });
            parsedConfigs = undefined;
        }
        builderProgram = undefined;
    }
    function getResolutionCache() {
        return resolutionCache;
    }
    function getCurrentBuilderProgram() {
        return builderProgram;
    }
    function getCurrentProgram() {
        return builderProgram && builderProgram.getProgramOrUndefined();
    }
    function synchronizeProgram() {
        writeLog(`Synchronizing program`);
        Debug.assert(compilerOptions);
        Debug.assert(rootFileNames);
        clearInvalidateResolutionsOfFailedLookupLocations();
        const program = getCurrentBuilderProgram();
        if (hasChangedCompilerOptions) {
            newLine = updateNewLine();
            if (program && changesAffectModuleResolution(program.getCompilerOptions(), compilerOptions)) {
                resolutionCache.onChangesAffectModuleResolution();
            }
        }
        const { hasInvalidatedResolutions, hasInvalidatedLibResolutions } = resolutionCache.createHasInvalidatedResolutions(customHasInvalidatedResolutions, customHasInvalidLibResolutions);
        const { originalReadFile, originalFileExists, originalDirectoryExists, originalCreateDirectory, originalWriteFile, readFileWithCache, } = changeCompilerHostLikeToUseCache(compilerHost, toPath);
        if (isProgramUptoDate(getCurrentProgram(), rootFileNames, compilerOptions, path => getSourceVersion(path, readFileWithCache), fileName => compilerHost.fileExists(fileName), hasInvalidatedResolutions, hasInvalidatedLibResolutions, hasChangedAutomaticTypeDirectiveNames, getParsedCommandLine, projectReferences)) {
            if (hasChangedConfigFileParsingErrors) {
                if (reportFileChangeDetectedOnCreateProgram) {
                    reportWatchDiagnostic(Diagnostics.File_change_detected_Starting_incremental_compilation);
                }
                builderProgram = createProgram(/*rootNames*/ undefined, /*options*/ undefined, compilerHost, builderProgram, configFileParsingDiagnostics, projectReferences);
                hasChangedConfigFileParsingErrors = false;
            }
        }
        else {
            if (reportFileChangeDetectedOnCreateProgram) {
                reportWatchDiagnostic(Diagnostics.File_change_detected_Starting_incremental_compilation);
            }
            createNewProgram(hasInvalidatedResolutions, hasInvalidatedLibResolutions);
        }
        reportFileChangeDetectedOnCreateProgram = false;
        if (host.afterProgramCreate && program !== builderProgram) {
            host.afterProgramCreate(builderProgram);
        }
        compilerHost.readFile = originalReadFile;
        compilerHost.fileExists = originalFileExists;
        compilerHost.directoryExists = originalDirectoryExists;
        compilerHost.createDirectory = originalCreateDirectory;
        compilerHost.writeFile = originalWriteFile;
        staleWatches?.forEach((configFile, configPath) => {
            if (!configPath) {
                // Update the wild card directory watch
                watchConfigFileWildCardDirectories();
                // Update extended config file watch
                if (configFileName)
                    updateExtendedConfigFilesWatches(toPath(configFileName), compilerOptions, watchOptions, WatchType.ExtendedConfigFile);
            }
            else {
                const config = parsedConfigs?.get(configPath);
                if (config)
                    watchReferencedProject(configFile, configPath, config);
            }
        });
        staleWatches = undefined;
        return builderProgram;
    }
    function createNewProgram(hasInvalidatedResolutions, hasInvalidatedLibResolutions) {
        // Compile the program
        writeLog("CreatingProgramWith::");
        writeLog(`  roots: ${JSON.stringify(rootFileNames)}`);
        writeLog(`  options: ${JSON.stringify(compilerOptions)}`);
        if (projectReferences)
            writeLog(`  projectReferences: ${JSON.stringify(projectReferences)}`);
        const needsUpdateInTypeRootWatch = hasChangedCompilerOptions || !getCurrentProgram();
        hasChangedCompilerOptions = false;
        hasChangedConfigFileParsingErrors = false;
        resolutionCache.startCachingPerDirectoryResolution();
        compilerHost.hasInvalidatedResolutions = hasInvalidatedResolutions;
        compilerHost.hasInvalidatedLibResolutions = hasInvalidatedLibResolutions;
        compilerHost.hasChangedAutomaticTypeDirectiveNames = hasChangedAutomaticTypeDirectiveNames;
        const oldProgram = getCurrentProgram();
        builderProgram = createProgram(rootFileNames, compilerOptions, compilerHost, builderProgram, configFileParsingDiagnostics, projectReferences);
        resolutionCache.finishCachingPerDirectoryResolution(builderProgram.getProgram(), oldProgram);
        // Update watches
        updateMissingFilePathsWatch(builderProgram.getProgram(), missingFilesMap || (missingFilesMap = new Map()), watchMissingFilePath);
        if (needsUpdateInTypeRootWatch) {
            resolutionCache.updateTypeRootsWatch();
        }
        if (missingFilePathsRequestedForRelease) {
            // These are the paths that program creater told us as not in use any more but were missing on the disk.
            // We didnt remove the entry for them from sourceFiles cache so that we dont have to do File IO,
            // if there is already watcher for it (for missing files)
            // At this point our watches were updated, hence now we know that these paths are not tracked and need to be removed
            // so that at later time we have correct result of their presence
            for (const missingFilePath of missingFilePathsRequestedForRelease) {
                if (!missingFilesMap.has(missingFilePath)) {
                    sourceFilesCache.delete(missingFilePath);
                }
            }
            missingFilePathsRequestedForRelease = undefined;
        }
    }
    function updateRootFileNames(files) {
        Debug.assert(!configFileName, "Cannot update root file names with config file watch mode");
        rootFileNames = files;
        scheduleProgramUpdate();
    }
    function updateNewLine() {
        return getNewLineCharacter(compilerOptions || optionsToExtendForConfigFile);
    }
    function toPath(fileName) {
        return ts_toPath(fileName, currentDirectory, getCanonicalFileName);
    }
    function isFileMissingOnHost(hostSourceFile) {
        return typeof hostSourceFile === "boolean";
    }
    function isFilePresenceUnknownOnHost(hostSourceFile) {
        return typeof hostSourceFile.version === "boolean";
    }
    function fileExists(fileName) {
        const path = toPath(fileName);
        // If file is missing on host from cache, we can definitely say file doesnt exist
        // otherwise we need to ensure from the disk
        if (isFileMissingOnHost(sourceFilesCache.get(path))) {
            return false;
        }
        return directoryStructureHost.fileExists(fileName);
    }
    function getVersionedSourceFileByPath(fileName, path, languageVersionOrOptions, onError, shouldCreateNewSourceFile) {
        const hostSourceFile = sourceFilesCache.get(path);
        // No source file on the host
        if (isFileMissingOnHost(hostSourceFile)) {
            return undefined;
        }
        // Create new source file if requested or the versions dont match
        const impliedNodeFormat = typeof languageVersionOrOptions === "object" ? languageVersionOrOptions.impliedNodeFormat : undefined;
        if (hostSourceFile === undefined || shouldCreateNewSourceFile || isFilePresenceUnknownOnHost(hostSourceFile) || hostSourceFile.sourceFile.impliedNodeFormat !== impliedNodeFormat) {
            const sourceFile = getNewSourceFile(fileName, languageVersionOrOptions, onError);
            if (hostSourceFile) {
                if (sourceFile) {
                    // Set the source file and create file watcher now that file was present on the disk
                    hostSourceFile.sourceFile = sourceFile;
                    hostSourceFile.version = sourceFile.version;
                    if (!hostSourceFile.fileWatcher) {
                        hostSourceFile.fileWatcher = watchFilePath(path, fileName, onSourceFileChange, PollingInterval.Low, watchOptions, WatchType.SourceFile);
                    }
                }
                else {
                    // There is no source file on host any more, close the watch, missing file paths will track it
                    if (hostSourceFile.fileWatcher) {
                        hostSourceFile.fileWatcher.close();
                    }
                    sourceFilesCache.set(path, false);
                }
            }
            else {
                if (sourceFile) {
                    const fileWatcher = watchFilePath(path, fileName, onSourceFileChange, PollingInterval.Low, watchOptions, WatchType.SourceFile);
                    sourceFilesCache.set(path, { sourceFile, version: sourceFile.version, fileWatcher });
                }
                else {
                    sourceFilesCache.set(path, false);
                }
            }
            return sourceFile;
        }
        return hostSourceFile.sourceFile;
    }
    function nextSourceFileVersion(path) {
        const hostSourceFile = sourceFilesCache.get(path);
        if (hostSourceFile !== undefined) {
            if (isFileMissingOnHost(hostSourceFile)) {
                // The next version, lets set it as presence unknown file
                sourceFilesCache.set(path, { version: false });
            }
            else {
                hostSourceFile.version = false;
            }
        }
    }
    function getSourceVersion(path, readFileWithCache) {
        const hostSourceFile = sourceFilesCache.get(path);
        if (!hostSourceFile)
            return undefined;
        if (hostSourceFile.version)
            return hostSourceFile.version;
        // Read file and get new version
        const text = readFileWithCache(path);
        return text !== undefined ? getSourceFileVersionAsHashFromText(compilerHost, text) : undefined;
    }
    function onReleaseOldSourceFile(oldSourceFile, _oldOptions, hasSourceFileByPath) {
        const hostSourceFileInfo = sourceFilesCache.get(oldSourceFile.resolvedPath);
        // If this is the source file thats in the cache and new program doesnt need it,
        // remove the cached entry.
        // Note we arent deleting entry if file became missing in new program or
        // there was version update and new source file was created.
        if (hostSourceFileInfo !== undefined) {
            // record the missing file paths so they can be removed later if watchers arent tracking them
            if (isFileMissingOnHost(hostSourceFileInfo)) {
                (missingFilePathsRequestedForRelease || (missingFilePathsRequestedForRelease = [])).push(oldSourceFile.path);
            }
            else if (hostSourceFileInfo.sourceFile === oldSourceFile) {
                if (hostSourceFileInfo.fileWatcher) {
                    hostSourceFileInfo.fileWatcher.close();
                }
                sourceFilesCache.delete(oldSourceFile.resolvedPath);
                if (!hasSourceFileByPath) {
                    resolutionCache.removeResolutionsOfFile(oldSourceFile.path);
                }
            }
        }
    }
    function reportWatchDiagnostic(message) {
        if (host.onWatchStatusChange) {
            host.onWatchStatusChange(createCompilerDiagnostic(message), newLine, compilerOptions || optionsToExtendForConfigFile);
        }
    }
    function hasChangedAutomaticTypeDirectiveNames() {
        return resolutionCache.hasChangedAutomaticTypeDirectiveNames();
    }
    function clearInvalidateResolutionsOfFailedLookupLocations() {
        if (!timerToInvalidateFailedLookupResolutions)
            return false;
        host.clearTimeout(timerToInvalidateFailedLookupResolutions);
        timerToInvalidateFailedLookupResolutions = undefined;
        return true;
    }
    function scheduleInvalidateResolutionsOfFailedLookupLocations() {
        if (!host.setTimeout || !host.clearTimeout) {
            return resolutionCache.invalidateResolutionsOfFailedLookupLocations();
        }
        const pending = clearInvalidateResolutionsOfFailedLookupLocations();
        writeLog(`Scheduling invalidateFailedLookup${pending ? ", Cancelled earlier one" : ""}`);
        timerToInvalidateFailedLookupResolutions = (() => {
            const timerHost = host;
            const callback = invalidateResolutionsOfFailedLookup;
            const timeoutDelay = 250;
            const t1 = "timerToInvalidateFailedLookupResolutions";
            return timerHost.setTimeout(() => callback(t1), timeoutDelay, t1);
        })();
    }
    function invalidateResolutionsOfFailedLookup(_timeoutType) {
        timerToInvalidateFailedLookupResolutions = undefined;
        if (resolutionCache.invalidateResolutionsOfFailedLookupLocations()) {
            scheduleProgramUpdate();
        }
    }
    // Upon detecting a file change, wait for 250ms and then perform a recompilation. This gives batch
    // operations (such as saving all modified files in an editor) a chance to complete before we kick
    // off a new compilation.
    function scheduleProgramUpdate() {
        if (!host.setTimeout || !host.clearTimeout) {
            return;
        }
        if (timerToUpdateProgram) {
            host.clearTimeout(timerToUpdateProgram);
        }
        writeLog("Scheduling update");
        timerToUpdateProgram = (() => {
            const timerHost = host;
            const callback = updateProgramWithWatchStatus;
            const timeoutDelay = 250;
            const t1 = "timerToUpdateProgram";
            return timerHost.setTimeout(() => callback(t1), timeoutDelay, t1);
        })();
    }
    function scheduleProgramReload() {
        Debug.assert(!!configFileName);
        updateLevel = ProgramUpdateLevel.Full;
        scheduleProgramUpdate();
    }
    function updateProgramWithWatchStatus(_timeoutType) {
        timerToUpdateProgram = undefined;
        reportFileChangeDetectedOnCreateProgram = true;
        updateProgram();
    }
    function updateProgram() {
        switch (updateLevel) {
            case ProgramUpdateLevel.RootNamesAndUpdate:
                reloadFileNamesFromConfigFile();
                break;
            case ProgramUpdateLevel.Full:
                reloadConfigFile();
                break;
            default:
                synchronizeProgram();
                break;
        }
        return getCurrentBuilderProgram();
    }
    function reloadFileNamesFromConfigFile() {
        writeLog("Reloading new file names and options");
        Debug.assert(compilerOptions);
        Debug.assert(configFileName);
        updateLevel = ProgramUpdateLevel.Update;
        rootFileNames = getFileNamesFromConfigSpecs(compilerOptions.configFile.configFileSpecs, getNormalizedAbsolutePath(getDirectoryPath(configFileName), currentDirectory), compilerOptions, parseConfigFileHost, extraFileExtensions);
        if (updateErrorForNoInputFiles(rootFileNames, getNormalizedAbsolutePath(configFileName, currentDirectory), compilerOptions.configFile.configFileSpecs, configFileParsingDiagnostics, canConfigFileJsonReportNoInputFiles)) {
            hasChangedConfigFileParsingErrors = true;
        }
        // Update the program
        synchronizeProgram();
    }
    function reloadConfigFile() {
        Debug.assert(configFileName);
        writeLog(`Reloading config file: ${configFileName}`);
        updateLevel = ProgramUpdateLevel.Update;
        if (cachedDirectoryStructureHost) {
            cachedDirectoryStructureHost.clearCache();
        }
        parseConfigFile();
        hasChangedCompilerOptions = true;
        (staleWatches ??= new Map()).set(undefined, undefined);
        synchronizeProgram();
    }
    function parseConfigFile() {
        Debug.assert(configFileName);
        setConfigFileParsingResult(getParsedCommandLineOfConfigFile(configFileName, optionsToExtendForConfigFile, parseConfigFileHost, extendedConfigCache ||= new Map(), watchOptionsToExtend, extraFileExtensions)); // TODO: GH#18217
    }
    function setConfigFileParsingResult(configFileParseResult) {
        rootFileNames = configFileParseResult.fileNames;
        compilerOptions = configFileParseResult.options;
        watchOptions = configFileParseResult.watchOptions;
        projectReferences = configFileParseResult.projectReferences;
        wildcardDirectories = configFileParseResult.wildcardDirectories;
        configFileParsingDiagnostics = getConfigFileParsingDiagnostics(configFileParseResult).slice();
        canConfigFileJsonReportNoInputFiles = canJsonReportNoInputFiles(configFileParseResult.raw);
        hasChangedConfigFileParsingErrors = true;
    }
    function getParsedCommandLine(configFileName) {
        const configPath = toPath(configFileName);
        let config = parsedConfigs?.get(configPath);
        if (config) {
            if (!config.updateLevel)
                return config.parsedCommandLine;
            // With host implementing getParsedCommandLine we cant just update file names
            if (config.parsedCommandLine && config.updateLevel === ProgramUpdateLevel.RootNamesAndUpdate && !host.getParsedCommandLine) {
                writeLog("Reloading new file names and options");
                Debug.assert(compilerOptions);
                const fileNames = getFileNamesFromConfigSpecs(config.parsedCommandLine.options.configFile.configFileSpecs, getNormalizedAbsolutePath(getDirectoryPath(configFileName), currentDirectory), compilerOptions, parseConfigFileHost);
                config.parsedCommandLine = { ...config.parsedCommandLine, fileNames };
                config.updateLevel = undefined;
                return config.parsedCommandLine;
            }
        }
        writeLog(`Loading config file: ${configFileName}`);
        const parsedCommandLine = host.getParsedCommandLine ?
            host.getParsedCommandLine(configFileName) :
            getParsedCommandLineFromConfigFileHost(configFileName);
        if (config) {
            config.parsedCommandLine = parsedCommandLine;
            config.updateLevel = undefined;
        }
        else {
            (parsedConfigs ||= new Map()).set(configPath, config = { parsedCommandLine });
        }
        (staleWatches ??= new Map()).set(configPath, configFileName);
        return parsedCommandLine;
    }
    function getParsedCommandLineFromConfigFileHost(configFileName) {
        // Ignore the file absent errors
        const onUnRecoverableConfigFileDiagnostic = parseConfigFileHost.onUnRecoverableConfigFileDiagnostic;
        parseConfigFileHost.onUnRecoverableConfigFileDiagnostic = noop;
        const parsedCommandLine = getParsedCommandLineOfConfigFile(configFileName, 
        /*optionsToExtend*/ undefined, parseConfigFileHost, extendedConfigCache ||= new Map(), watchOptionsToExtend);
        parseConfigFileHost.onUnRecoverableConfigFileDiagnostic = onUnRecoverableConfigFileDiagnostic;
        return parsedCommandLine;
    }
    function onReleaseParsedCommandLine(fileName) {
        const path = toPath(fileName);
        const config = parsedConfigs?.get(path);
        if (!config)
            return;
        parsedConfigs.delete(path);
        if (config.watchedDirectories)
            clearMap(config.watchedDirectories, closeFileWatcherOf);
        config.watcher?.close();
        clearSharedExtendedConfigFileWatcher(path, sharedExtendedConfigFileWatchers);
    }
    function watchFilePath(path, file, callback, pollingInterval, options, watchType) {
        return watchFile(file, (fileName, eventKind) => callback(fileName, eventKind, path), pollingInterval, options, watchType);
    }
    function onSourceFileChange(fileName, eventKind, path) {
        updateCachedSystemWithFile(fileName, path, eventKind);
        // Update the source file cache
        if (eventKind === FileWatcherEventKind.Deleted && sourceFilesCache.has(path)) {
            resolutionCache.invalidateResolutionOfFile(path);
        }
        nextSourceFileVersion(path);
        // Update the program
        scheduleProgramUpdate();
    }
    function updateCachedSystemWithFile(fileName, path, eventKind) {
        if (cachedDirectoryStructureHost) {
            cachedDirectoryStructureHost.addOrDeleteFile(fileName, path, eventKind);
        }
    }
    function watchMissingFilePath(missingFilePath, missingFileName) {
        // If watching missing referenced config file, we are already watching it so no need for separate watcher
        return parsedConfigs?.has(missingFilePath) ?
            noopFileWatcher :
            watchFilePath(missingFilePath, missingFileName, onMissingFileChange, PollingInterval.Medium, watchOptions, WatchType.MissingFile);
    }
    function onMissingFileChange(fileName, eventKind, missingFilePath) {
        updateCachedSystemWithFile(fileName, missingFilePath, eventKind);
        if (eventKind === FileWatcherEventKind.Created && missingFilesMap.has(missingFilePath)) {
            missingFilesMap.get(missingFilePath).close();
            missingFilesMap.delete(missingFilePath);
            // Delete the entry in the source files cache so that new source file is created
            nextSourceFileVersion(missingFilePath);
            // When a missing file is created, we should update the graph.
            scheduleProgramUpdate();
        }
    }
    function watchConfigFileWildCardDirectories() {
        updateWatchingWildcardDirectories(watchedWildcardDirectories || (watchedWildcardDirectories = new Map()), wildcardDirectories, watchWildcardDirectory);
    }
    function watchWildcardDirectory(directory, flags) {
        return watchDirectory(directory, fileOrDirectory => {
            Debug.assert(configFileName);
            Debug.assert(compilerOptions);
            const fileOrDirectoryPath = toPath(fileOrDirectory);
            // Since the file existence changed, update the sourceFiles cache
            if (cachedDirectoryStructureHost) {
                cachedDirectoryStructureHost.addOrDeleteFileOrDirectory(fileOrDirectory, fileOrDirectoryPath);
            }
            nextSourceFileVersion(fileOrDirectoryPath);
            if (isIgnoredFileFromWildCardWatching({
                watchedDirPath: toPath(directory),
                fileOrDirectory,
                fileOrDirectoryPath,
                configFileName,
                extraFileExtensions,
                options: compilerOptions,
                program: getCurrentBuilderProgram() || rootFileNames,
                currentDirectory,
                useCaseSensitiveFileNames,
                writeLog,
                toPath,
            }))
                return;
            // Reload is pending, do the reload
            if (updateLevel !== ProgramUpdateLevel.Full) {
                updateLevel = ProgramUpdateLevel.RootNamesAndUpdate;
                // Schedule Update the program
                scheduleProgramUpdate();
            }
        }, flags, watchOptions, WatchType.WildcardDirectory);
    }
    function updateExtendedConfigFilesWatches(forProjectPath, options, watchOptions, watchType) {
        updateSharedExtendedConfigFileWatcher(forProjectPath, options, sharedExtendedConfigFileWatchers ||= new Map(), (extendedConfigFileName, extendedConfigFilePath) => watchFile(extendedConfigFileName, (_fileName, eventKind) => {
            updateCachedSystemWithFile(extendedConfigFileName, extendedConfigFilePath, eventKind);
            // Update extended config cache
            if (extendedConfigCache)
                cleanExtendedConfigCache(extendedConfigCache, extendedConfigFilePath, toPath);
            // Update projects
            const projects = sharedExtendedConfigFileWatchers.get(extendedConfigFilePath)?.projects;
            // If there are no referenced projects this extended config file watcher depend on ignore
            if (!projects?.size)
                return;
            projects.forEach(projectPath => {
                if (configFileName && toPath(configFileName) === projectPath) {
                    // If this is the config file of the project, reload completely
                    updateLevel = ProgramUpdateLevel.Full;
                }
                else {
                    // Reload config for the referenced projects and remove the resolutions from referenced projects since the config file changed
                    const config = parsedConfigs?.get(projectPath);
                    if (config)
                        config.updateLevel = ProgramUpdateLevel.Full;
                    resolutionCache.removeResolutionsFromProjectReferenceRedirects(projectPath);
                }
                scheduleProgramUpdate();
            });
        }, PollingInterval.High, watchOptions, watchType), toPath);
    }
    function watchReferencedProject(configFileName, configPath, commandLine) {
        // Watch file
        commandLine.watcher ||= watchFile(configFileName, (_fileName, eventKind) => {
            updateCachedSystemWithFile(configFileName, configPath, eventKind);
            const config = parsedConfigs?.get(configPath);
            if (config)
                config.updateLevel = ProgramUpdateLevel.Full;
            resolutionCache.removeResolutionsFromProjectReferenceRedirects(configPath);
            scheduleProgramUpdate();
        }, PollingInterval.High, commandLine.parsedCommandLine?.watchOptions || watchOptions, WatchType.ConfigFileOfReferencedProject);
        // Watch Wild card
        updateWatchingWildcardDirectories(commandLine.watchedDirectories ||= new Map(), commandLine.parsedCommandLine?.wildcardDirectories, (directory, flags) => watchDirectory(directory, fileOrDirectory => {
            const fileOrDirectoryPath = toPath(fileOrDirectory);
            // Since the file existence changed, update the sourceFiles cache
            if (cachedDirectoryStructureHost) {
                cachedDirectoryStructureHost.addOrDeleteFileOrDirectory(fileOrDirectory, fileOrDirectoryPath);
            }
            nextSourceFileVersion(fileOrDirectoryPath);
            const config = parsedConfigs?.get(configPath);
            if (!config?.parsedCommandLine)
                return;
            if (isIgnoredFileFromWildCardWatching({
                watchedDirPath: toPath(directory),
                fileOrDirectory,
                fileOrDirectoryPath,
                configFileName,
                options: config.parsedCommandLine.options,
                program: config.parsedCommandLine.fileNames,
                currentDirectory,
                useCaseSensitiveFileNames,
                writeLog,
                toPath,
            }))
                return;
            // Reload is pending, do the reload
            if (config.updateLevel !== ProgramUpdateLevel.Full) {
                config.updateLevel = ProgramUpdateLevel.RootNamesAndUpdate;
                // Schedule Update the program
                scheduleProgramUpdate();
            }
        }, flags, commandLine.parsedCommandLine?.watchOptions || watchOptions, WatchType.WildcardDirectoryOfReferencedProject));
        // Watch extended config files
        updateExtendedConfigFilesWatches(configPath, commandLine.parsedCommandLine?.options, commandLine.parsedCommandLine?.watchOptions || watchOptions, WatchType.ExtendedConfigOfReferencedProject);
    }
}
