import { type WatchOptions } from "./_namespaces/ts.js";
export declare enum FileWatcherEventKind {
    Created = 0,
    Changed = 1,
    Deleted = 2
}
export type FileWatcherCallback = (fileName: string, eventKind: FileWatcherEventKind, modifiedTime?: Date) => void;
export type DirectoryWatcherCallback = (fileName: string) => void;
export type BufferEncoding = "ascii" | "utf8" | "utf-8" | "utf16le" | "ucs2" | "ucs-2" | "base64" | "latin1" | "binary" | "hex";
export interface System {
    args: string[];
    newLine: string;
    useCaseSensitiveFileNames: boolean;
    write(s: string): void;
    writeOutputIsTTY?(): boolean;
    getWidthOfTerminal?(): number;
    readFile(path: string, encoding?: string): string | undefined;
    getFileSize?(path: string): number;
    writeFile(path: string, data: string, writeByteOrderMark?: boolean): void;
    /**
     * @pollingInterval - this parameter is used in polling-based watchers and ignored in watchers that
     * use native OS file watching
     */
    watchFile?(path: string, callback: FileWatcherCallback, pollingInterval?: number, options?: WatchOptions): FileWatcher;
    watchDirectory?(path: string, callback: DirectoryWatcherCallback, recursive?: boolean, options?: WatchOptions): FileWatcher;
    resolvePath(path: string): string;
    fileExists(path: string): boolean;
    directoryExists(path: string): boolean;
    createDirectory(path: string): void;
    getExecutingFilePath(): string;
    getCurrentDirectory(): string;
    getDirectories(path: string): string[];
    readDirectory(path: string, extensions?: readonly string[], exclude?: readonly string[], include?: readonly string[], depth?: number): string[];
    getModifiedTime?(path: string): Date | undefined;
    setModifiedTime?(path: string, time: Date): void;
    deleteFile?(path: string): void;
    /**
     * A good implementation is node.js' `crypto.createHash`. (https://nodejs.org/api/crypto.html#crypto_crypto_createhash_algorithm)
     */
    createHash?(data: string): string;
    /** This must be cryptographically secure. Only implement this method using `crypto.createHash("sha256")`. */
    createSHA256Hash?: {
        createSHA256Hash(data: string): string;
    }["createSHA256Hash"] | undefined;
    getMemoryUsage?(): number;
    exit(exitCode?: number): number | undefined;
    realpath?(path: string): string;
    setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;
    clearTimeout?(timeoutId: any): void;
    clearScreen?(): void;
    base64decode?(input: string): string;
    base64encode?(input: string): string;
}
export interface FileWatcher {
    close(): void;
}
export declare let sys: System;
