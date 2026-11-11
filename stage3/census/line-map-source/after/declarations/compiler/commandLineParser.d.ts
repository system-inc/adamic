import { type BuildOptions, type CompilerOptions, type Diagnostic, type FileExtensionInfo, type JsonSourceFile, type ParseConfigHost, type ParsedCommandLine, type Path, type TsConfigSourceFile, type TypeAcquisition, type WatchOptions } from "./_namespaces/ts.js";
export declare function parseCommandLine(commandLine: readonly string[], readFile?: (path: string) => string | undefined): ParsedCommandLine;
/** Parsed command line for build */
export interface ParsedBuildCommand {
    buildOptions: BuildOptions;
    watchOptions: WatchOptions | undefined;
    projects: string[];
    errors: Diagnostic[];
}
export declare function parseBuildCommand(commandLine: readonly string[]): ParsedBuildCommand;
export type DiagnosticReporter = (diagnostic: Diagnostic) => void;
/**
 * Reports config file diagnostics
 */
export interface ConfigFileDiagnosticsReporter {
    /**
     * Reports unrecoverable error when parsing config file
     */
    onUnRecoverableConfigFileDiagnostic: DiagnosticReporter;
}
/**
 * Interface extending ParseConfigHost to support ParseConfigFile that reads config file and reports errors
 */
export interface ParseConfigFileHost extends ParseConfigHost, ConfigFileDiagnosticsReporter {
    getCurrentDirectory(): string;
}
/**
 * Reads the config file, reports errors if any and exits if the config file cannot be found
 */
export declare function getParsedCommandLineOfConfigFile(configFileName: string, optionsToExtend: CompilerOptions | undefined, host: ParseConfigFileHost, extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, watchOptionsToExtend?: WatchOptions, extraFileExtensions?: readonly FileExtensionInfo[]): ParsedCommandLine | undefined;
/**
 * Read tsconfig.json file
 * @param fileName The path to the config file
 */
export declare function readConfigFile(fileName: string, readFile: (path: string) => string | undefined): {
    config?: AdamicJsonRecoveryObject;
    error?: Diagnostic | undefined;
};
/**
 * Parse the text of the tsconfig.json file
 * @param fileName The path to the config file
 * @param jsonText The text of the config file
 */
export declare function parseConfigFileTextToJson(fileName: string, jsonText: string): {
    config?: any;
    error?: Diagnostic | undefined;
};
/**
 * Read tsconfig.json file
 * @param fileName The path to the config file
 */
export declare function readJsonConfigFile(fileName: string, readFile: (path: string) => string | undefined): TsConfigSourceFile;
export type AdamicJsonRecoveryValue = string | number | boolean | null | AdamicJsonRecoveryValue[] | AdamicJsonRecoveryObject;
export interface AdamicJsonRecoveryObject {
    [key: string]: AdamicJsonRecoveryValue | undefined;
}
/**
 * Convert the json syntax tree into the json value
 */
export declare function convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): AdamicJsonRecoveryValue | undefined;
/**
 * Parse the contents of a config file (tsconfig.json).
 * @param json The contents of the config file to parse
 * @param host Instance of ParseConfigHost used to enumerate files in folder.
 * @param basePath A root directory to resolve relative path entries in the config
 *    file to. e.g. outDir
 */
export declare function parseJsonConfigFileContent(json: unknown, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine;
/**
 * Parse the contents of a config file (tsconfig.json).
 * @param jsonNode The contents of the config file to parse
 * @param host Instance of ParseConfigHost used to enumerate files in folder.
 * @param basePath A root directory to resolve relative path entries in the config
 *    file to. e.g. outDir
 */
export declare function parseJsonSourceFileConfigFileContent(sourceFile: TsConfigSourceFile, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine;
export interface ParsedTsconfig {
    raw: unknown;
    options?: CompilerOptions;
    watchOptions?: WatchOptions | undefined;
    typeAcquisition?: TypeAcquisition;
    /**
     * Note that the case of the config path has not yet been normalized, as no files have been imported into the project yet
     */
    extendedConfigPath?: string | string[] | undefined;
}
export interface ExtendedConfigCacheEntry {
    extendedResult: TsConfigSourceFile;
    extendedConfig: ParsedTsconfig | undefined;
}
export declare function convertCompilerOptionsFromJson(jsonOptions: unknown, basePath: string, configFileName?: string): {
    options: CompilerOptions;
    errors: Diagnostic[];
};
export declare function convertTypeAcquisitionFromJson(jsonOptions: unknown, basePath: string, configFileName?: string): {
    options: TypeAcquisition;
    errors: Diagnostic[];
};
