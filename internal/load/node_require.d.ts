// Refine Node's own Require interface without copying any host declarations.
declare namespace NodeJS {
 interface Require {
  (specifier: unknown, ...arguments: unknown[]): unknown;
  (specifier: 'fs'): typeof import('node:fs');
  (specifier: 'node:fs'): typeof import('node:fs');
  (specifier: 'path'): typeof import('node:path');
  (specifier: 'node:path'): typeof import('node:path');
  (specifier: 'perf_hooks'): typeof import('node:perf_hooks');
  (specifier: 'node:perf_hooks'): typeof import('node:perf_hooks');
 }
}
