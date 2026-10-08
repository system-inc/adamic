import type { SourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as SourceFile).patternAmbientModules;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.symbol.flags}`;
}
const present = { kind: 308, patternAmbientModules: [{ symbol: { flags: 7 } }] };
console.log(read(present));
const absent = { kind: 308 };
console.log(read(absent));
const explicit = { kind: 308, patternAmbientModules: undefined };
console.log(read(explicit));
