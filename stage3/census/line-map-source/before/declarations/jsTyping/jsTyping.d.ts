export interface TypingResolutionHost {
    directoryExists(path: string): boolean;
    fileExists(fileName: string): boolean;
    readFile(path: string, encoding?: string): string | undefined;
    readDirectory(rootDir: string, extensions: readonly string[], excludes: readonly string[] | undefined, includes: readonly string[] | undefined, depth?: number): string[];
}
