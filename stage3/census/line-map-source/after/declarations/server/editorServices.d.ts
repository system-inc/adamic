import { CompilerOptions, Diagnostic, FileExtensionInfo, FormatCodeSettings, HostCancellationToken, JSDocParsingMode, LanguageServiceMode, Path, ScriptKind, TypeAcquisition, WatchOptions } from "./_namespaces/ts.js";
import { ConfiguredProject, ExternalProject, InferredProject, InvalidateCachedTypings, ITypingsInstaller, Logger, NormalizedPath, PackageInstalledResponse, Project, ScriptInfo, ServerHost, Session, SetTypings } from "./_namespaces/ts.server.js";
import * as protocol from "./protocol.js";
export declare const maxProgramSizeForNonTsFiles: number;
export declare const ProjectsUpdatedInBackgroundEvent = "projectsUpdatedInBackground";
export declare const ProjectLoadingStartEvent = "projectLoadingStart";
export declare const ProjectLoadingFinishEvent = "projectLoadingFinish";
export declare const LargeFileReferencedEvent = "largeFileReferenced";
export declare const ConfigFileDiagEvent = "configFileDiag";
export declare const ProjectLanguageServiceStateEvent = "projectLanguageServiceState";
export declare const ProjectInfoTelemetryEvent = "projectInfo";
export declare const OpenFileInfoTelemetryEvent = "openFileInfo";
export declare const CreateFileWatcherEvent: protocol.CreateFileWatcherEventName;
export declare const CreateDirectoryWatcherEvent: protocol.CreateDirectoryWatcherEventName;
export declare const CloseFileWatcherEvent: protocol.CloseFileWatcherEventName;
export interface ProjectsUpdatedInBackgroundEvent {
    eventName: typeof ProjectsUpdatedInBackgroundEvent;
    data: {
        openFiles: string[];
    };
}
export interface ProjectLoadingStartEvent {
    eventName: typeof ProjectLoadingStartEvent;
    data: {
        project: Project;
        reason: string;
    };
}
export interface ProjectLoadingFinishEvent {
    eventName: typeof ProjectLoadingFinishEvent;
    data: {
        project: Project;
    };
}
export interface LargeFileReferencedEvent {
    eventName: typeof LargeFileReferencedEvent;
    data: {
        file: string;
        fileSize: number;
        maxFileSize: number;
    };
}
export interface ConfigFileDiagEvent {
    eventName: typeof ConfigFileDiagEvent;
    data: {
        triggerFile: string;
        configFileName: string;
        diagnostics: readonly Diagnostic[];
    };
}
export interface ProjectLanguageServiceStateEvent {
    eventName: typeof ProjectLanguageServiceStateEvent;
    data: {
        project: Project;
        languageServiceEnabled: boolean;
    };
}
/** This will be converted to the payload of a protocol.TelemetryEvent in session.defaultEventHandler. */
export interface ProjectInfoTelemetryEvent {
    readonly eventName: typeof ProjectInfoTelemetryEvent;
    readonly data: ProjectInfoTelemetryEventData;
}
export interface ProjectInfoTelemetryEventData {
    /** Cryptographically secure hash of project file location. */
    readonly projectId: string;
    /** Count of file extensions seen in the project. */
    readonly fileStats: FileStats;
    /**
     * Any compiler options that might contain paths will be taken out.
     * Enum compiler options will be converted to strings.
     */
    readonly compilerOptions: CompilerOptions;
    readonly extends: boolean | undefined;
    readonly files: boolean | undefined;
    readonly include: boolean | undefined;
    readonly exclude: boolean | undefined;
    readonly compileOnSave: boolean;
    readonly typeAcquisition: ProjectInfoTypeAcquisitionData;
    readonly configFileName: "tsconfig.json" | "jsconfig.json" | "other";
    readonly projectType: "external" | "configured";
    readonly languageServiceEnabled: boolean;
    /** TypeScript version used by the server. */
    readonly version: string;
}
/**
 * Info that we may send about a file that was just opened.
 * Info about a file will only be sent once per session, even if the file changes in ways that might affect the info.
 * Currently this is only sent for '.js' files.
 */
export interface OpenFileInfoTelemetryEvent {
    readonly eventName: typeof OpenFileInfoTelemetryEvent;
    readonly data: OpenFileInfoTelemetryEventData;
}
export interface OpenFileInfoTelemetryEventData {
    readonly info: OpenFileInfo;
}
export interface ProjectInfoTypeAcquisitionData {
    readonly enable: boolean | undefined;
    readonly include: boolean;
    readonly exclude: boolean;
}
export interface FileStats {
    readonly js: number;
    readonly jsSize?: number;
    readonly jsx: number;
    readonly jsxSize?: number;
    readonly ts: number;
    readonly tsSize?: number;
    readonly tsx: number;
    readonly tsxSize?: number;
    readonly dts: number;
    readonly dtsSize?: number;
    readonly deferred: number;
    readonly deferredSize?: number;
}
export interface OpenFileInfo {
    readonly checkJs: boolean;
}
export interface CreateFileWatcherEvent {
    readonly eventName: protocol.CreateFileWatcherEventName;
    readonly data: protocol.CreateFileWatcherEventBody;
}
export interface CreateDirectoryWatcherEvent {
    readonly eventName: protocol.CreateDirectoryWatcherEventName;
    readonly data: protocol.CreateDirectoryWatcherEventBody;
}
export interface CloseFileWatcherEvent {
    readonly eventName: protocol.CloseFileWatcherEventName;
    readonly data: protocol.CloseFileWatcherEventBody;
}
export type ProjectServiceEvent = LargeFileReferencedEvent | ProjectsUpdatedInBackgroundEvent | ProjectLoadingStartEvent | ProjectLoadingFinishEvent | ConfigFileDiagEvent | ProjectLanguageServiceStateEvent | ProjectInfoTelemetryEvent | OpenFileInfoTelemetryEvent | CreateFileWatcherEvent | CreateDirectoryWatcherEvent | CloseFileWatcherEvent;
export type ProjectServiceEventHandler = (event: ProjectServiceEvent) => void;
export interface SafeList {
    [name: string]: {
        match: RegExp;
        exclude?: (string | number)[][];
        types?: string[];
    };
}
export interface TypesMapFile {
    typesMap: SafeList;
    simpleMap: {
        [libName: string]: string;
    };
}
export declare function convertFormatOptions(protocolOptions: protocol.FormatCodeSettings): FormatCodeSettings;
export declare function convertCompilerOptions(protocolOptions: protocol.ExternalProjectCompilerOptions): CompilerOptions & protocol.CompileOnSaveMixin;
export declare function convertWatchOptions(protocolOptions: protocol.ExternalProjectCompilerOptions, currentDirectory?: string): WatchOptionsAndErrors | undefined;
export declare function convertTypeAcquisition(protocolOptions: protocol.InferredProjectCompilerOptions): TypeAcquisition | undefined;
export declare function tryConvertScriptKindName(scriptKindName: protocol.ScriptKindName | ScriptKind): ScriptKind;
export declare function convertScriptKindName(scriptKindName: protocol.ScriptKindName): ScriptKind;
export interface HostConfiguration {
    formatCodeOptions: FormatCodeSettings;
    preferences: protocol.UserPreferences;
    hostInfo: string;
    extraFileExtensions?: FileExtensionInfo[];
    watchOptions?: WatchOptions;
}
export interface OpenConfiguredProjectResult {
    configFileName?: NormalizedPath;
    configFileErrors?: readonly Diagnostic[];
}
export declare const nullTypingsInstaller: ITypingsInstaller;
export interface ProjectServiceOptions {
    host: ServerHost;
    logger: Logger;
    cancellationToken: HostCancellationToken;
    useSingleInferredProject: boolean;
    useInferredProjectPerProjectRoot: boolean;
    typingsInstaller?: ITypingsInstaller;
    eventHandler?: ProjectServiceEventHandler;
    canUseWatchEvents?: boolean;
    suppressDiagnosticEvents?: boolean;
    throttleWaitMilliseconds?: number;
    globalPlugins?: readonly string[];
    pluginProbeLocations?: readonly string[];
    allowLocalPluginLoads?: boolean;
    typesMapLocation?: string;
    serverMode?: LanguageServiceMode;
    session: Session<unknown> | undefined;
    jsDocParsingMode?: JSDocParsingMode;
}
export interface WatchOptionsAndErrors {
    watchOptions: WatchOptions;
    errors: Diagnostic[] | undefined;
}
export declare class ProjectService {
    private readonly nodeModulesWatchers;
    /**
     * Contains all the deleted script info's version information so that
     * it does not reset when creating script info again
     * (and could have potentially collided with version where contents mismatch)
     */
    private readonly filenameToScriptInfoVersion;
    private readonly allJsFilesForOpenFileTelemetry;
    /**
     * maps external project file name to list of config files that were the part of this project
     */
    private readonly externalProjectToConfiguredProjectMap;
    /**
     * external projects (configuration and list of root files is not controlled by tsserver)
     */
    readonly externalProjects: ExternalProject[];
    /**
     * projects built from openFileRoots
     */
    readonly inferredProjects: InferredProject[];
    /**
     * projects specified by a tsconfig.json file
     */
    readonly configuredProjects: Map<string, ConfiguredProject>;
    /**
     * Open files: with value being project root path, and key being Path of the file that is open
     */
    readonly openFiles: Map<Path, NormalizedPath | undefined>;
    /** Config files looked up and cached config files for open script info */
    private readonly configFileForOpenFiles;
    /** Set of open script infos that are root of inferred project */
    private rootOfInferredProjects;
    /**
     * Map of open files that are opened without complete path but have projectRoot as current directory
     */
    private readonly openFilesWithNonRootedDiskPath;
    private compilerOptionsForInferredProjects;
    private compilerOptionsForInferredProjectsPerProjectRoot;
    private watchOptionsForInferredProjects;
    private watchOptionsForInferredProjectsPerProjectRoot;
    private typeAcquisitionForInferredProjects;
    private typeAcquisitionForInferredProjectsPerProjectRoot;
    /**
     * Project size for configured or external projects
     */
    private readonly projectToSizeMap;
    private readonly hostConfiguration;
    private safelist;
    private readonly legacySafelist;
    private pendingProjectUpdates;
    /**
     * All the open script info that needs recalculation of the default project,
     * this also caches config file info before config file change was detected to use it in case projects are not updated yet
     */
    private pendingOpenFileProjectUpdates?;
    readonly currentDirectory: NormalizedPath;
    readonly toCanonicalFileName: (f: string) => string;
    readonly host: ServerHost;
    readonly logger: Logger;
    readonly cancellationToken: HostCancellationToken;
    readonly useSingleInferredProject: boolean;
    readonly useInferredProjectPerProjectRoot: boolean;
    readonly typingsInstaller: ITypingsInstaller;
    private readonly globalCacheLocationDirectoryPath;
    readonly throttleWaitMilliseconds?: number;
    private readonly suppressDiagnosticEvents?;
    readonly globalPlugins: readonly string[];
    readonly pluginProbeLocations: readonly string[];
    readonly allowLocalPluginLoads: boolean;
    readonly typesMapLocation: string | undefined;
    readonly serverMode: LanguageServiceMode;
    /** Tracks projects that we have already sent telemetry for. */
    private readonly seenProjects;
    private readonly sharedExtendedConfigFileWatchers;
    private readonly extendedConfigCache;
    private packageJsonFilesMap;
    private incompleteCompletionsCache;
    private performanceEventHandler?;
    private pendingPluginEnablements?;
    private currentPluginEnablementPromise?;
    readonly jsDocParsingMode: JSDocParsingMode | undefined;
    constructor(opts: ProjectServiceOptions);
    toPath(fileName: string): Path;
    private loadTypesMap;
    updateTypingsForProject(response: SetTypings | InvalidateCachedTypings | PackageInstalledResponse): void;
    private delayUpdateProjectGraph;
    private delayUpdateProjectGraphs;
    setCompilerOptionsForInferredProjects(projectCompilerOptions: protocol.InferredProjectCompilerOptions, projectRootPath?: string): void;
    findProject(projectName: string): Project | undefined;
    getDefaultProjectForFile(fileName: NormalizedPath, ensureProject: boolean): Project | undefined;
    /**
     * If there is default project calculation pending for this file,
     * then it completes that calculation so that correct default project is used for the project
     */
    private tryGetDefaultProjectForEnsuringConfiguredProjectForFile;
    private doEnsureDefaultProjectForFile;
    getScriptInfoEnsuringProjectsUptoDate(uncheckedFileName: string): ScriptInfo | undefined;
    /**
     * Ensures the project structures are upto date
     * This means,
     * - we go through all the projects and update them if they are dirty
     * - if updates reflect some change in structure or there was pending request to ensure projects for open files
     *   ensure that each open script info has project
     */
    private ensureProjectStructuresUptoDate;
    getFormatCodeOptions(file: NormalizedPath): FormatCodeSettings;
    getPreferences(file: NormalizedPath): protocol.UserPreferences;
    getHostFormatCodeOptions(): FormatCodeSettings;
    getHostPreferences(): protocol.UserPreferences;
    private onSourceFileChanged;
    private handleSourceMapProjects;
    private delayUpdateSourceInfoProjects;
    private delayUpdateProjectsOfScriptInfoPath;
    private handleDeletedFile;
    /**
     * This is to watch whenever files are added or removed to the wildcard directories
     */
    private watchWildcardDirectory;
    private onWildCardDirectoryWatcherInvoke;
    private delayUpdateProjectsFromParsedConfigOnConfigFileChange;
    private onConfigFileChanged;
    private removeProject;
    private assignOrphanScriptInfosToInferredProject;
    /**
     * Remove this file from the set of open, non-configured files.
     * @param info The file that has been closed or newly configured
     */
    private closeOpenFile;
    private deleteScriptInfo;
    private configFileExists;
    private createConfigFileWatcherForParsedConfig;
    private ensureConfigFileWatcherForProject;
    /**
     * This function tries to search for a tsconfig.json for the given file.
     * This is different from the method the compiler uses because
     * the compiler can assume it will always start searching in the
     * current directory (the directory in which tsc was invoked).
     * The server must start searching from the directory containing
     * the newly opened file.
     */
    private forEachConfigFileLocation;
    /** Get cached configFileName for scriptInfo or ancestor of open script info */
    private getConfigFileNameForFileFromCache;
    /** Caches the configFilename for script info or ancestor of open script info */
    private setConfigFileNameForFileInCache;
    private printProjects;
    private getConfiguredProjectByCanonicalConfigFilePath;
    private findExternalProjectByProjectName;
    /** Get a filename if the language service exceeds the maximum allowed program size; otherwise returns undefined. */
    private getFilenameForExceededTotalSizeLimitForNonTsFiles;
    private createExternalProject;
    private addFilesToNonInferredProject;
    /**
     * Read the config file of the project, and update the project root file names.
     */
    private loadConfiguredProject;
    private updateNonInferredProjectFiles;
    private updateRootAndOptionsOfNonInferredProject;
    private reloadFileNamesOfParsedConfig;
    private setProjectForReload;
    private clearSemanticCache;
    private getOrCreateInferredProjectForProjectRootPathIfEnabled;
    private getOrCreateSingleInferredProjectIfEnabled;
    private getOrCreateSingleInferredWithoutProjectRoot;
    private createInferredProject;
    getScriptInfo(uncheckedFileName: string): ScriptInfo | undefined;
    private watchClosedScriptInfo;
    private createNodeModulesWatcher;
    private watchClosedScriptInfoInNodeModules;
    private getModifiedTime;
    private refreshScriptInfo;
    private refreshScriptInfosInDirectory;
    private stopWatchingScriptInfo;
    private getOrCreateScriptInfoNotOpenedByClientForNormalizedPath;
    getOrCreateScriptInfoForNormalizedPath(fileName: NormalizedPath, openedByClient: boolean, fileContent?: string, scriptKind?: ScriptKind, hasMixedContent?: boolean, hostToQueryFileExistsOn?: {
        fileExists(path: string): boolean;
    }): ScriptInfo | undefined;
    private getOrCreateScriptInfoWorker;
    /**
     * This gets the script info for the normalized path. If the path is not rooted disk path then the open script info with project root context is preferred
     */
    getScriptInfoForNormalizedPath(fileName: NormalizedPath): ScriptInfo | undefined;
    getScriptInfoForPath(fileName: Path): ScriptInfo | undefined;
    private addSourceInfoToSourceMap;
    private addMissingSourceMapFile;
    setHostConfiguration(args: protocol.ConfigureRequestArguments): void;
    private getWatchOptionsFromProjectWatchOptions;
    closeLog(): void;
    private sendSourceFileChange;
    /**
     * This function rebuilds the project for every file opened by the client
     * This does not reload contents of open files from disk. But we could do that if needed
     */
    reloadProjects(): void;
    /**
     * Remove the root of inferred project if script info is part of another project
     */
    private removeRootOfInferredProjectIfNowPartOfOtherProject;
    /**
     * This function is to update the project structure for every inferred project.
     * It is called on the premise that all the configured projects are
     * up to date.
     * This will go through open files and assign them to inferred project if open file is not part of any other project
     * After that all the inferred project graphs are updated
     */
    private ensureProjectForOpenFiles;
    /**
     * Open file whose contents is managed by the client
     * @param filename is absolute pathname
     * @param fileContent is a known version of the file content that is more up to date than the one on disk
     */
    openClientFile(fileName: string, fileContent?: string, scriptKind?: ScriptKind, projectRootPath?: string): OpenConfiguredProjectResult;
    private findExternalProjectContainingOpenScriptInfo;
    private getOrCreateOpenScriptInfo;
    private assignProjectToOpenedScriptInfo;
    /**
     * Finds the default configured project for given info
     * For any tsconfig found, it looks into that project, if not then all its references,
     * The search happens for all tsconfigs till projectRootPath
     */
    private tryFindDefaultConfiguredProjectForOpenScriptInfo;
    private isMatchedByConfig;
    private tryFindDefaultConfiguredProjectForOpenScriptInfoOrClosedFileInfo;
    /**
     * Finds the default configured project, if found, it creates the solution projects (does not load them right away)
     * with Find: finds the projects even if the project is deferredClosed
     */
    private tryFindDefaultConfiguredProjectAndLoadAncestorsForOpenScriptInfo;
    private ensureProjectChildren;
    private cleanupConfiguredProjects;
    private cleanupProjectsAndScriptInfos;
    private tryInvokeWildCardDirectories;
    openClientFileWithNormalizedPath(fileName: NormalizedPath, fileContent?: string, scriptKind?: ScriptKind, hasMixedContent?: boolean, projectRootPath?: NormalizedPath): OpenConfiguredProjectResult;
    private removeOrphanScriptInfos;
    private telemetryOnOpenFile;
    /**
     * Close file whose contents is managed by the client
     * @param filename is absolute pathname
     */
    closeClientFile(uncheckedFileName: string): void;
    private collectChanges;
    closeExternalProject(uncheckedFileName: string): void;
    openExternalProjects(projects: protocol.ExternalProject[]): void;
    /** Makes a filename safe to insert in a RegExp */
    private static readonly filenameEscapeRegexp;
    private static escapeFilenameForRegex;
    resetSafeList(): void;
    applySafeList(proj: protocol.ExternalProject): NormalizedPath[];
    private applySafeListWorker;
    openExternalProject(proj: protocol.ExternalProject): void;
    hasDeferredExtension(): boolean;
    /**
     * Performs the remaining steps of enabling a plugin after its module has been instantiated.
     */
    private endEnablePlugin;
    private enableRequestedPluginsAsync;
    private enableRequestedPluginsWorker;
    configurePlugin(args: protocol.ConfigurePluginRequestArguments): void;
    private watchPackageJsonFile;
    private onPackageJsonChange;
}
