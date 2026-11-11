import { type CancellationToken, type CompilerHost, type CompilerOptions, type CreateProgramOptions, type Diagnostic, type DiagnosticMessageChain, type FileReference, type ModuleResolutionHost, type PackageJsonInfoCache, type ParsedCommandLine, type Program, type ProjectReference, type ResolutionMode, type ResolvedConfigFileName, type SourceFile, type StringLiteralLike } from "./_namespaces/ts.js";
export declare function findConfigFile(searchPath: string, fileExists: (fileName: string) => boolean, configName?: string): string | undefined;
export declare function resolveTripleslashReference(moduleName: string, containingFile: string): string;
export declare function createCompilerHost(options: CompilerOptions, setParentNodes?: boolean): CompilerHost;
export declare function getPreEmitDiagnostics(program: Program, sourceFile?: SourceFile, cancellationToken?: CancellationToken): readonly Diagnostic[];
export interface FormatDiagnosticsHost {
    getCurrentDirectory(): string;
    getCanonicalFileName(fileName: string): string;
    getNewLine(): string;
}
export declare function formatDiagnostics(diagnostics: readonly Diagnostic[], host: FormatDiagnosticsHost): string;
export declare function formatDiagnostic(diagnostic: Diagnostic, host: FormatDiagnosticsHost): string;
export declare function formatDiagnosticsWithColorAndContext(diagnostics: readonly Diagnostic[], host: FormatDiagnosticsHost): string;
export declare function flattenDiagnosticMessageText(diag: string | DiagnosticMessageChain | undefined, newLine: string, indent?: number): string;
/**
 * Calculates the resulting resolution mode for some reference in some file - this is generally the explicitly
 * provided resolution mode in the reference, unless one is not present, in which case it is the mode of the containing file.
 */
export declare function getModeForFileReference(ref: FileReference | string, containingFileMode: ResolutionMode): ResolutionMode;
/**
 * Use `program.getModeForResolutionAtIndex`, which retrieves the correct `compilerOptions`, instead of this function whenever possible.
 * Calculates the final resolution mode for an import at some index within a file's `imports` list. This is the resolution mode
 * explicitly provided via import attributes, if present, or the syntax the usage would have if emitted to JavaScript. In
 * `--module node16` or `nodenext`, this may depend on the file's `impliedNodeFormat`. In `--module preserve`, it depends only on the
 * input syntax of the reference. In other `module` modes, when overriding import attributes are not provided, this function returns
 * `undefined`, as the result would have no impact on module resolution, emit, or type checking.
 * @param file File to fetch the resolution mode within
 * @param index Index into the file's complete resolution list to get the resolution of - this is a concatenation of the file's imports and module augmentations
 * @param compilerOptions The compiler options for the program that owns the file. If the file belongs to a referenced project, the compiler options
 * should be the options of the referenced project, not the referencing project.
 */
export declare function getModeForResolutionAtIndex(file: SourceFile, index: number, compilerOptions: CompilerOptions): ResolutionMode;
/**
 * Use `program.getModeForUsageLocation`, which retrieves the correct `compilerOptions`, instead of this function whenever possible.
 * Calculates the final resolution mode for a given module reference node. This function only returns a result when module resolution
 * settings allow differing resolution between ESM imports and CJS requires, or when a mode is explicitly provided via import attributes,
 * which cause an `import` or `require` condition to be used during resolution regardless of module resolution settings. In absence of
 * overriding attributes, and in modes that support differing resolution, the result indicates the syntax the usage would emit to JavaScript.
 * Some examples:
 *
 * ```ts
 * // tsc foo.mts --module nodenext
 * import {} from "mod";
 * // Result: ESNext - the import emits as ESM due to `impliedNodeFormat` set by .mts file extension
 *
 * // tsc foo.cts --module nodenext
 * import {} from "mod";
 * // Result: CommonJS - the import emits as CJS due to `impliedNodeFormat` set by .cts file extension
 *
 * // tsc foo.ts --module preserve --moduleResolution bundler
 * import {} from "mod";
 * // Result: ESNext - the import emits as ESM due to `--module preserve` and `--moduleResolution bundler`
 * // supports conditional imports/exports
 *
 * // tsc foo.ts --module preserve --moduleResolution node10
 * import {} from "mod";
 * // Result: undefined - the import emits as ESM due to `--module preserve`, but `--moduleResolution node10`
 * // does not support conditional imports/exports
 *
 * // tsc foo.ts --module commonjs --moduleResolution node10
 * import type {} from "mod" with { "resolution-mode": "import" };
 * // Result: ESNext - conditional imports/exports always supported with "resolution-mode" attribute
 * ```
 *
 * @param file The file the import or import-like reference is contained within
 * @param usage The module reference string
 * @param compilerOptions The compiler options for the program that owns the file. If the file belongs to a referenced project, the compiler options
 * should be the options of the referenced project, not the referencing project.
 * @returns The final resolution mode of the import
 */
export declare function getModeForUsageLocation(file: SourceFile, usage: StringLiteralLike, compilerOptions: CompilerOptions): ResolutionMode;
export declare function getConfigFileParsingDiagnostics(configFileParseResult: ParsedCommandLine): readonly Diagnostic[];
/**
 * A function for determining if a given file is esm or cjs format, assuming modern node module resolution rules, as configured by the
 * `options` parameter.
 *
 * @param fileName The file name to check the format of (it need not exist on disk)
 * @param [packageJsonInfoCache] A cache for package file lookups - it's best to have a cache when this function is called often
 * @param host The ModuleResolutionHost which can perform the filesystem lookups for package json data
 * @param options The compiler options to perform the analysis under - relevant options are `moduleResolution` and `traceResolution`
 * @returns `undefined` if the path has no relevant implied format, `ModuleKind.ESNext` for esm format, and `ModuleKind.CommonJS` for cjs format
 */
export declare function getImpliedNodeFormatForFile(fileName: string, packageJsonInfoCache: PackageJsonInfoCache | undefined, host: ModuleResolutionHost, options: CompilerOptions): ResolutionMode;
/**
 * Create a new 'Program' instance. A Program is an immutable collection of 'SourceFile's and a 'CompilerOptions'
 * that represent a compilation unit.
 *
 * Creating a program proceeds from a set of root files, expanding the set of inputs by following imports and
 * triple-slash-reference-path directives transitively. '@types' and triple-slash-reference-types are also pulled in.
 *
 * @param createProgramOptions - The options for creating a program.
 * @returns A 'Program' object.
 */
export declare function createProgram(createProgramOptions: CreateProgramOptions): Program;
/**
 * Create a new 'Program' instance. A Program is an immutable collection of 'SourceFile's and a 'CompilerOptions'
 * that represent a compilation unit.
 *
 * Creating a program proceeds from a set of root files, expanding the set of inputs by following imports and
 * triple-slash-reference-path directives transitively. '@types' and triple-slash-reference-types are also pulled in.
 *
 * @param rootNames - A set of root files.
 * @param options - The compiler options which should be used.
 * @param host - The host interacts with the underlying file system.
 * @param oldProgram - Reuses an old program structure.
 * @param configFileParsingDiagnostics - error during config file parsing
 * @returns A 'Program' object.
 */
export declare function createProgram(rootNames: readonly string[], options: CompilerOptions, host?: CompilerHost, oldProgram?: Program, configFileParsingDiagnostics?: readonly Diagnostic[]): Program;
/**
 * Returns the target config filename of a project reference.
 * Note: The file might not exist.
 */
export declare function resolveProjectReferencePath(ref: ProjectReference): ResolvedConfigFileName;
