import { type AffectedFileResult, type BuilderProgram, type CancellationToken, type CompilerOptions, type CompilerOptionsValue, type CreateProgram, type CustomTransformers, type Diagnostic, type DiagnosticReporter, type EmitAndSemanticDiagnosticsBuilderProgram, type EmitResult, ExitStatus, type ParsedCommandLine, type Program, type ProgramHost, type ResolvedConfigFileName, type SourceFile, type System, type WatchHost, type WatchOptions, type WatchStatusReporter, type WriteFileCallback } from "./_namespaces/ts.js";
export interface BuildOptions {
    dry?: boolean;
    force?: boolean;
    verbose?: boolean;
    stopBuildOnErrors?: boolean;
    incremental?: boolean;
    assumeChangesOnlyAffectDirectDependencies?: boolean;
    declaration?: boolean;
    declarationMap?: boolean;
    emitDeclarationOnly?: boolean;
    sourceMap?: boolean;
    inlineSourceMap?: boolean;
    traceResolution?: boolean;
    [option: string]: CompilerOptionsValue | undefined;
}
export type ReportEmitErrorSummary = (errorCount: number, filesInError: (ReportFileInError | undefined)[]) => void;
export interface ReportFileInError {
    fileName: string;
    line: number;
}
export interface SolutionBuilderHostBase<T extends BuilderProgram> extends ProgramHost<T> {
    createDirectory?: {
        createDirectory(path: string): void;
    }["createDirectory"] | undefined;
    /**
     * Should provide create directory and writeFile if done of invalidatedProjects is not invoked with
     * writeFileCallback
     */
    writeFile?: {
        writeFile(path: string, data: string, writeByteOrderMark?: boolean): void;
    }["writeFile"] | undefined;
    getCustomTransformers?: (project: string) => CustomTransformers | undefined;
    getModifiedTime(fileName: string): Date | undefined;
    setModifiedTime(fileName: string, date: Date): void;
    deleteFile(fileName: string): void;
    getParsedCommandLine?(fileName: string): ParsedCommandLine | undefined;
    reportDiagnostic: DiagnosticReporter;
    reportSolutionBuilderStatus: DiagnosticReporter;
    afterProgramEmitAndDiagnostics?(program: T): void;
}
export interface SolutionBuilderHost<T extends BuilderProgram> extends SolutionBuilderHostBase<T> {
    reportErrorSummary?: ReportEmitErrorSummary | undefined;
}
export interface SolutionBuilderWithWatchHost<T extends BuilderProgram> extends SolutionBuilderHostBase<T>, WatchHost {
}
export interface SolutionBuilder<T extends BuilderProgram> {
    build(project?: string, cancellationToken?: CancellationToken, writeFile?: WriteFileCallback, getCustomTransformers?: (project: string) => CustomTransformers): ExitStatus;
    clean(project?: string): ExitStatus;
    buildReferences(project: string, cancellationToken?: CancellationToken, writeFile?: WriteFileCallback, getCustomTransformers?: (project: string) => CustomTransformers): ExitStatus;
    cleanReferences(project?: string): ExitStatus;
    getNextInvalidatedProject(cancellationToken?: CancellationToken): InvalidatedProject<T> | undefined;
}
/**
 * Create a function that reports watch status by writing to the system and handles the formating of the diagnostic
 */
export declare function createBuilderStatusReporter(system: System, pretty?: boolean): DiagnosticReporter;
export declare function createSolutionBuilderHost<T extends BuilderProgram = EmitAndSemanticDiagnosticsBuilderProgram>(system?: System, createProgram?: CreateProgram<T>, reportDiagnostic?: DiagnosticReporter, reportSolutionBuilderStatus?: DiagnosticReporter, reportErrorSummary?: ReportEmitErrorSummary): SolutionBuilderHost<T>;
export declare function createSolutionBuilderWithWatchHost<T extends BuilderProgram = EmitAndSemanticDiagnosticsBuilderProgram>(system?: System, createProgram?: CreateProgram<T>, reportDiagnostic?: DiagnosticReporter, reportSolutionBuilderStatus?: DiagnosticReporter, reportWatchStatus?: WatchStatusReporter): SolutionBuilderWithWatchHost<T>;
export declare function createSolutionBuilder<T extends BuilderProgram>(host: SolutionBuilderHost<T>, rootNames: readonly string[], defaultOptions: BuildOptions): SolutionBuilder<T>;
export declare function createSolutionBuilderWithWatch<T extends BuilderProgram>(host: SolutionBuilderWithWatchHost<T>, rootNames: readonly string[], defaultOptions: BuildOptions, baseWatchOptions?: WatchOptions): SolutionBuilder<T>;
export declare enum InvalidatedProjectKind {
    Build = 0,
    UpdateOutputFileStamps = 1
}
export interface InvalidatedProjectBase {
    readonly kind: InvalidatedProjectKind;
    readonly project: ResolvedConfigFileName;
    /**
     *  To dispose this project and ensure that all the necessary actions are taken and state is updated accordingly
     */
    done(cancellationToken?: CancellationToken, writeFile?: WriteFileCallback, customTransformers?: CustomTransformers): ExitStatus;
    getCompilerOptions(): CompilerOptions;
    getCurrentDirectory(): string;
}
export interface UpdateOutputFileStampsProject extends InvalidatedProjectBase {
    readonly kind: InvalidatedProjectKind.UpdateOutputFileStamps;
    updateOutputFileStatmps(): void;
}
export interface BuildInvalidedProject<T extends BuilderProgram> extends InvalidatedProjectBase {
    readonly kind: InvalidatedProjectKind.Build;
    getBuilderProgram(): T | undefined;
    getProgram(): Program | undefined;
    getSourceFile(fileName: string): SourceFile | undefined;
    getSourceFiles(): readonly SourceFile[];
    getOptionsDiagnostics(cancellationToken?: CancellationToken): readonly Diagnostic[];
    getGlobalDiagnostics(cancellationToken?: CancellationToken): readonly Diagnostic[];
    getConfigFileParsingDiagnostics(): readonly Diagnostic[];
    getSyntacticDiagnostics(sourceFile?: SourceFile, cancellationToken?: CancellationToken): readonly Diagnostic[];
    getAllDependencies(sourceFile: SourceFile): readonly string[];
    getSemanticDiagnostics(sourceFile?: SourceFile, cancellationToken?: CancellationToken): readonly Diagnostic[];
    getSemanticDiagnosticsOfNextAffectedFile(cancellationToken?: CancellationToken, ignoreSourceFile?: (sourceFile: SourceFile) => boolean): AffectedFileResult<readonly Diagnostic[]>;
    emit(targetSourceFile?: SourceFile, writeFile?: WriteFileCallback, cancellationToken?: CancellationToken, emitOnlyDtsFiles?: boolean, customTransformers?: CustomTransformers): EmitResult | undefined;
}
export type InvalidatedProject<T extends BuilderProgram> = UpdateOutputFileStampsProject | BuildInvalidedProject<T>;
