import type { TsConfigSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as TsConfigSourceFile).parseDiagnostics;
 return `${items.length}:${items[0]!.start}`;
}
const present = { kind: 308, parseDiagnostics: [{ start: 7 }] };
console.log(read(present));
