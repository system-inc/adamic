import type { SourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 308, patternAmbientModules: [{ symbol: {} }] };
const base: Base = raw;
const items = (base as SourceFile).patternAmbientModules!;
console.log(`${items[0]!.symbol.flags}`);
