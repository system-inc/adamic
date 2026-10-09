import type { TsConfigSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as TsConfigSourceFile;
 console.log('kept');
}
const raw = {kind: 308, extendedSourceFiles: 7};
read(raw);
