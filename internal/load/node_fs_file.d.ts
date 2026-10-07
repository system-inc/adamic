// Selected signatures from @types/node 25.3.3, fs.d.ts. Buffer and URL paths,
// byte views and encodings other than UTF-8 await their host units.
// Alias names are unit-prefixed so independent node:fs declarations can merge.
interface Error { readonly code?: string; }
declare module 'node:fs' {
    type FSFilePathLike = string;
    type FSFilePathOrFileDescriptor = FSFilePathLike | number;
    type FSFileBufferEncoding = 'utf8' | 'utf-8';
    type FSFileOpenMode = number | string;
    type FSFileMode = number | string;
    type FSFileTimeLike = string | number | Date;
    interface Stats {
        isFile(): boolean;
        isDirectory(): boolean;
        isSymbolicLink(): boolean;
        size: number;
        mtime: Date;
        mtimeMs: number;
    }
    interface StatSyncOptions { bigint?: boolean | undefined; throwIfNoEntry?: boolean | undefined; }
    interface MakeDirectoryOptions { recursive?: boolean | undefined; mode?: FSFileMode | undefined; }
    type FSFileWriteFileOptions = { encoding?: FSFileBufferEncoding | null | undefined; mode?: FSFileMode | undefined; flag?: string | undefined; flush?: boolean | undefined; } | FSFileBufferEncoding | null;
    function readFileSync(path: FSFilePathOrFileDescriptor, options: { encoding: FSFileBufferEncoding; flag?: string | undefined; } | FSFileBufferEncoding): string;
    function openSync(path: FSFilePathLike, flags: FSFileOpenMode, mode?: FSFileMode | null): number;
    function writeSync(fd: number, string: string, position?: number | null, encoding?: FSFileBufferEncoding | null): number;
    function closeSync(fd: number): void;
    function writeFileSync(file: FSFilePathOrFileDescriptor, data: string, options?: FSFileWriteFileOptions): void;
    function existsSync(path: FSFilePathLike): boolean;
    function statSync(path: FSFilePathLike, options: StatSyncOptions & { throwIfNoEntry: false; }): Stats | undefined;
    function statSync(path: FSFilePathLike, options?: StatSyncOptions): Stats;
    function mkdirSync(path: FSFilePathLike, options: MakeDirectoryOptions & { recursive: true; }): string | undefined;
    function mkdirSync(path: FSFilePathLike, options?: FSFileMode | (MakeDirectoryOptions & { recursive?: false | undefined; }) | null): void;
    function mkdirSync(path: FSFilePathLike, options?: FSFileMode | MakeDirectoryOptions | null): string | undefined;
    function unlinkSync(path: FSFilePathLike): void;
    function utimesSync(path: FSFilePathLike, atime: FSFileTimeLike, mtime: FSFileTimeLike): void;
}
