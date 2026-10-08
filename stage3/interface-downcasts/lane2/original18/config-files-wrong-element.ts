import type { TsConfigSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 308, extendedSourceFiles: [7] };
const base: Base = raw;
const items = (base as TsConfigSourceFile).extendedSourceFiles!;
console.log(`${items[0]!}`);
