import * as ts from "./_namespaces/ts.js";
import { ApplyCodeActionCommandResult, CompilerOptions, Diagnostic, HostCancellationToken, InstallPackageOptions, IScriptSnapshot, JSDocParsingMode, LanguageService, LanguageServiceHost, ModuleResolutionHost, Path, PluginImport, ProgramUpdateLevel, ProjectReference, ScriptKind, SortedReadonlyArray, SourceFile, TypeAcquisition, WatchOptions } from "./_namespaces/ts.js";
import { NormalizedPath, ProjectService, ScriptInfo, ServerHost, Session } from "./_namespaces/ts.server.js";
export declare enum ProjectKind {
    Inferred = 0,
    Configured = 1,
    External = 2,
    AutoImportProvider = 3,
    Auxiliary = 4
}
export declare function allRootFilesAreJsOrDts(project: Project): boolean;
export declare function allFilesAreJsOrDts(project: Project): boolean;
export interface PluginCreateInfo {
    project: Project;
    languageService: LanguageService;
    languageServiceHost: LanguageServiceHost;
    serverHost: ServerHost;
    session?: Session<unknown>;
    config: unknown;
}
export interface PluginModule {
    create(createInfo: PluginCreateInfo): LanguageService;
    getExternalFiles?(proj: Project, updateLevel: ProgramUpdateLevel): string[];
    onConfigurationChanged?(config: unknown): void;
}
export interface PluginModuleWithName {
    name: string;
    module: PluginModule;
}
export type PluginModuleFactory = (mod: {
    typescript: typeof ts;
}) => PluginModule;
export declare abstract class Project implements LanguageServiceHost, ModuleResolutionHost {
    readonly projectKind: ProjectKind;
    readonly projectService: ProjectService;
    private compilerOptions;
    compileOnSaveEnabled: boolean;
    protected watchOptions: WatchOptions | undefined;
    private rootFilesMap;
    private program;
    private externalFiles;
    private missingFilesMap;
    private generatedFilesMap;
    private hasAddedorRemovedFiles;
    private hasAddedOrRemovedSymlinks;
    protected languageService: LanguageService;
    languageServiceEnabled: boolean;
    readonly trace?: (s: string) => void;
    readonly realpath?: (path: string) => string;
    private builderState;
    /**
     * Set of files names that were updated since the last call to getChangesSinceVersion.
     */
    private updatedFileNames;
    /**
     * Set of files that was returned from the last call to getChangesSinceVersion.
     */
    private lastReportedFileNames;
    /**
     * Last version that was reported.
     */
    private lastReportedVersion;
    protected projectErrors: Diagnostic[] | undefined;
    private typingsCache;
    private typingWatchers;
    private readonly cancellationToken;
    isNonTsProject(): boolean;
    isJsOnlyProject(): boolean;
    static resolveModule(moduleName: string, initialDir: string, host: ServerHost, log: (message: string) => void): {} | undefined;
    private exportMapCache;
    private changedFilesForExportMapCache;
    private moduleSpecifierCache;
    private symlinks;
    readonly jsDocParsingMode: JSDocParsingMode | undefined;
    isKnownTypesPackageName(name: string): boolean;
    installPackage(options: InstallPackageOptions): Promise<ApplyCodeActionCommandResult>;
    getCompilationSettings(): CompilerOptions;
    getCompilerOptions(): CompilerOptions;
    getNewLine(): string;
    getProjectVersion(): string;
    getProjectReferences(): readonly ProjectReference[] | undefined;
    getScriptFileNames(): string[];
    private getOrCreateScriptInfoAndAttachToProject;
    getScriptKind(fileName: string): ScriptKind;
    getScriptVersion(filename: string): string;
    getScriptSnapshot(filename: string): IScriptSnapshot | undefined;
    getCancellationToken(): HostCancellationToken;
    getCurrentDirectory(): string;
    getDefaultLibFileName(): string;
    useCaseSensitiveFileNames(): boolean;
    readDirectory(path: string, extensions?: readonly string[], exclude?: readonly string[], include?: readonly string[], depth?: number): string[];
    readFile(fileName: string): string | undefined;
    writeFile(fileName: string, content: string): void;
    fileExists(file: string): boolean;
    directoryExists(path: string): boolean;
    getDirectories(path: string): string[];
    log(s: string): void;
    error(s: string): void;
    private setInternalCompilerOptionsForEmittingJsFiles;
    /**
     * Get the errors that dont have any file name associated
     */
    getGlobalProjectErrors(): readonly Diagnostic[];
    /**
     * Get all the project errors
     */
    getAllProjectErrors(): readonly Diagnostic[];
    setProjectErrors(projectErrors: Diagnostic[] | undefined): void;
    getLanguageService(ensureSynchronized?: boolean): LanguageService;
    getCompileOnSaveAffectedFileList(scriptInfo: ScriptInfo): string[];
    /**
     * Returns true if emit was conducted
     */
    emitFile(scriptInfo: ScriptInfo, writeFile: (path: string, data: string, writeByteOrderMark?: boolean) => void): EmitResult;
    enableLanguageService(): void;
    disableLanguageService(lastFileExceededProgramSize?: string): void;
    getProjectName(): string;
    protected removeLocalTypingsFromTypeAcquisition(newTypeAcquisition: TypeAcquisition): TypeAcquisition;
    getExternalFiles(updateLevel?: ProgramUpdateLevel): SortedReadonlyArray<string>;
    getSourceFile(path: Path): SourceFile | undefined;
    close(): void;
    private detachScriptInfoIfNotRoot;
    isClosed(): boolean;
    hasRoots(): boolean;
    getRootFiles(): NormalizedPath[];
    getRootScriptInfos(): ScriptInfo[];
    getScriptInfos(): ScriptInfo[];
    getExcludedFiles(): readonly NormalizedPath[];
    getFileNames(excludeFilesFromExternalLibraries?: boolean, excludeConfigFiles?: boolean): NormalizedPath[];
    hasConfigFile(configFilePath: NormalizedPath): boolean;
    containsScriptInfo(info: ScriptInfo): boolean;
    containsFile(filename: NormalizedPath, requireOpen?: boolean): boolean;
    isRoot(info: ScriptInfo): boolean;
    addRoot(info: ScriptInfo, fileName?: NormalizedPath): void;
    addMissingFileRoot(fileName: NormalizedPath): void;
    removeFile(info: ScriptInfo, fileExists: boolean, detachFromProject: boolean): void;
    registerFileUpdate(fileName: string): void;
    /**
     * Updates set of files that contribute to this project
     * @returns: true if set of files in the project stays the same and false - otherwise.
     */
    updateGraph(): boolean;
    private closeWatchingTypingLocations;
    private onTypingInstallerWatchInvoke;
    protected removeExistingTypings(include: string[]): string[];
    private updateGraphWorker;
    private detachScriptInfoFromProject;
    private addMissingFileWatcher;
    private isWatchedMissingFile;
    private createGeneratedFileWatcher;
    private isValidGeneratedFileWatcher;
    private clearGeneratedFileWatch;
    getScriptInfoForNormalizedPath(fileName: NormalizedPath): ScriptInfo | undefined;
    getScriptInfo(uncheckedFileName: string): ScriptInfo | undefined;
    filesToString(writeProjectFileNames: boolean): string;
    private filesToStringWorker;
    setCompilerOptions(compilerOptions: CompilerOptions): void;
    setTypeAcquisition(newTypeAcquisition: TypeAcquisition | undefined): void;
    getTypeAcquisition(): TypeAcquisition;
    protected removeRoot(info: ScriptInfo): void;
    protected enableGlobalPlugins(options: CompilerOptions): void;
    protected enablePlugin(pluginConfigEntry: PluginImport, searchPaths: string[]): void;
    /** Starts a new check for diagnostics. Call this if some file has updated that would cause diagnostics to be changed. */
    refreshDiagnostics(): void;
    private isDefaultProjectForOpenFiles;
}
/**
 * If a file is opened and no tsconfig (or jsconfig) is found,
 * the file and its imports/references are put into an InferredProject.
 */
export declare class InferredProject extends Project {
    private _isJsInferredProject;
    toggleJsInferredProject(isJsInferredProject: boolean): void;
    setCompilerOptions(options?: CompilerOptions): void;
    /** this is canonical project root path */
    readonly projectRootPath: string | undefined;
    addRoot(info: ScriptInfo): void;
    removeRoot(info: ScriptInfo): void;
    isProjectWithSingleRoot(): boolean;
    close(): void;
    getTypeAcquisition(): TypeAcquisition;
}
export declare class AutoImportProviderProject extends Project {
    private hostProject;
    private static readonly maxDependencies;
    private rootFileNames;
    updateGraph(): boolean;
    hasRoots(): boolean;
    getScriptFileNames(): string[];
    getLanguageService(): never;
    getHostForAutoImportProvider(): never;
    getProjectReferences(): readonly ProjectReference[] | undefined;
}
/**
 * If a file is opened, the server will look for a tsconfig (or jsconfig)
 * and if successful create a ConfiguredProject for it.
 * Otherwise it will create an InferredProject.
 */
export declare class ConfiguredProject extends Project {
    readonly canonicalConfigFilePath: NormalizedPath;
    private projectReferences;
    private compilerHost?;
    private releaseParsedConfig;
    /**
     * If the project has reload from disk pending, it reloads (and then updates graph as part of that) instead of just updating the graph
     * @returns: true if set of files in the project stays the same and false - otherwise.
     */
    updateGraph(): boolean;
    getConfigFilePath(): NormalizedPath;
    getProjectReferences(): readonly ProjectReference[] | undefined;
    updateReferences(refs: readonly ProjectReference[] | undefined): void;
    /**
     * Get the errors that dont have any file name associated
     */
    getGlobalProjectErrors(): readonly Diagnostic[];
    /**
     * Get all the project errors
     */
    getAllProjectErrors(): readonly Diagnostic[];
    setProjectErrors(projectErrors: Diagnostic[]): void;
    close(): void;
    getEffectiveTypeRoots(): string[];
}
/**
 * Project whose configuration is handled externally, such as in a '.csproj'.
 * These are created only if a host explicitly calls `openExternalProject`.
 */
export declare class ExternalProject extends Project {
    externalProjectName: string;
    compileOnSaveEnabled: boolean;
    excludedFiles: readonly NormalizedPath[];
    updateGraph(): boolean;
    getExcludedFiles(): readonly NormalizedPath[];
}
