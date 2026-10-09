import type { TsConfigSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as TsConfigSourceFile;
 const items = viewed?.extendedSourceFiles;
 if (items === undefined) { console.log('absent'); return; }
 console.log(items.slice(0, 1).join(';'));
}
const raw = {kind: 308, extendedSourceFiles: ['a', 42]};
read(raw);
