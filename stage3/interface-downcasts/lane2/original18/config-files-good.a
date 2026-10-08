import type { TsConfigSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as TsConfigSourceFile).extendedSourceFiles;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!}`;
}
const present = { kind: 308, extendedSourceFiles: ['ok'] };
console.log(read(present));
const absent = { kind: 308 };
console.log(read(absent));
const explicit = { kind: 308, extendedSourceFiles: undefined };
console.log(read(explicit));
