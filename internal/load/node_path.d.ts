// Selected @types/node POSIX path signatures used by the host census.
declare module 'node:path' {
    export function resolve(...paths: string[]): string;
    export function dirname(path: string): string;
    export function join(...paths: string[]): string;
}
