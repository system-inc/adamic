import type { TsConfigSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as TsConfigSourceFile;
 console.log('kept');
}
const raw = {kind: 308, statements: 7};
read(raw);
