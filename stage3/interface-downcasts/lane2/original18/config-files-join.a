import type { TsConfigSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 308, extendedSourceFiles: ['a', 'b'] };
const base: Base = raw;
const items = (base as TsConfigSourceFile).extendedSourceFiles!;
console.log(items.join(';'));
