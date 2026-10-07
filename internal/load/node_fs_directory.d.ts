// String paths and UTF-8 results are the overloads used by sys.ts.
declare module 'node:fs' {
    export class Dirent {
        name: string;
        isFile(): boolean;
        isDirectory(): boolean;
        isSymbolicLink(): boolean;
    }
    export function readdirSync(path: string, options?: { encoding?: 'utf8'; withFileTypes?: false }): string[];
    export function readdirSync(path: string, options: { encoding?: 'utf8'; withFileTypes: true }): Dirent[];
    export function realpathSync(path: string): string;
    export namespace realpathSync {
        function native(path: string): string;
    }
}
// All errors may lack a code; filesystem errors carry one.
interface Error { code?: string; }
