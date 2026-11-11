import { Path, SortedArray, SortedReadonlyArray, TypeAcquisition } from "./_namespaces/ts.js";
import { DiscoverTypings, Project } from "./_namespaces/ts.server.js";
export declare enum LogLevel {
    terse = 0,
    normal = 1,
    requestTime = 2,
    verbose = 3
}
export declare const emptyArray: SortedReadonlyArray<never>;
export interface Logger {
    close(): void;
    hasLevel(level: LogLevel): boolean;
    loggingEnabled(): boolean;
    perftrc(s: string): void;
    info(s: string): void;
    startGroup(): void;
    endGroup(): void;
    msg(s: string, type?: Msg): void;
    getLogFileName(): string | undefined;
}
export declare enum Msg {
    Err = "Err",
    Info = "Info",
    Perf = "Perf"
}
export declare function createInstallTypingsRequest(project: Project, typeAcquisition: TypeAcquisition, unresolvedImports: SortedReadonlyArray<string>, cachePath?: string): DiscoverTypings;
export declare namespace Errors {
    function ThrowNoProject(): never;
    function ThrowProjectLanguageServiceDisabled(): never;
    function ThrowProjectDoesNotContainDocument(fileName: string, project: Project): never;
}
export type NormalizedPath = string & {
    __normalizedPathTag: any;
};
export declare function toNormalizedPath(fileName: string): NormalizedPath;
export declare function normalizedPathToPath(normalizedPath: NormalizedPath, currentDirectory: string, getCanonicalFileName: (f: string) => string): Path;
export declare function asNormalizedPath(fileName: string): NormalizedPath;
export interface NormalizedPathMap<T> {
    get(path: NormalizedPath): T | undefined;
    set(path: NormalizedPath, value: T): void;
    contains(path: NormalizedPath): boolean;
    remove(path: NormalizedPath): void;
}
export declare function createNormalizedPathMap<T>(): NormalizedPathMap<T>;
export declare function isInferredProjectName(name: string): boolean;
export declare function makeInferredProjectName(counter: number): string;
export declare function createSortedArray<T>(): SortedArray<T>;
