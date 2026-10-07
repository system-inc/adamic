interface AdamicNodeRequire {
 (specifier: 'fs' | 'node:fs'): typeof import('node:fs');
 (specifier: 'path' | 'node:path'): typeof import('node:path');
 (specifier: unknown, ...arguments: unknown[]): unknown;
 readonly resolve: (specifier: string) => string;
 readonly cache: unknown;
}
declare const require: AdamicNodeRequire;
declare const module: { exports: unknown };
// These signatures are from @types/node's path host surface. No path runtime
// is supplied here; lowering reports NotYet until its host unit is available.
declare module 'node:path' {
 function join(...paths: string[]): string;
 function dirname(path: string): string;
}
