import { ApplyCodeActionCommandResult, DirectoryWatcherCallback, FileWatcher, FileWatcherCallback, InstallPackageOptions, Path, SortedReadonlyArray, System, TypeAcquisition, WatchOptions } from "./_namespaces/ts.js";
import { Project, ProjectService } from "./_namespaces/ts.server.js";
export interface CompressedData {
    length: number;
    compressionKind: string;
    data: any;
}
export type ModuleImportResult = {
    module: {};
    error: undefined;
} | {
    module: undefined;
    error: unknown;
};
/** @deprecated Use {@link ModuleImportResult} instead. */
export type RequireResult = ModuleImportResult;
export interface ServerHost extends System {
    watchFile(path: string, callback: FileWatcherCallback, pollingInterval?: number, options?: WatchOptions): FileWatcher;
    watchDirectory(path: string, callback: DirectoryWatcherCallback, recursive?: boolean, options?: WatchOptions): FileWatcher;
    preferNonRecursiveWatch?: boolean;
    setTimeout(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;
    clearTimeout(timeoutId: any): void;
    setImmediate(callback: (...args: any[]) => void, ...args: any[]): any;
    clearImmediate(timeoutId: any): void;
    gc?(): void;
    trace?(s: string): void;
    require?(initialPath: string, moduleName: string): ModuleImportResult;
}
export interface InstallPackageOptionsWithProject extends InstallPackageOptions {
    projectName: string;
    projectRootPath: Path;
}
export interface ITypingsInstaller {
    isKnownTypesPackageName(name: string): boolean;
    installPackage(options: InstallPackageOptionsWithProject): Promise<ApplyCodeActionCommandResult>;
    enqueueInstallTypingsRequest(p: Project, typeAcquisition: TypeAcquisition, unresolvedImports: SortedReadonlyArray<string> | undefined): void;
    attach(projectService: ProjectService): void;
    onProjectClosed(p: Project): void;
    readonly globalTypingsCacheLocation: string | undefined;
}
