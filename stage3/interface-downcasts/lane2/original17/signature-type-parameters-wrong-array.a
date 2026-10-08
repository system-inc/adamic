import type { JSDocSignature } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 324, typeParameters: 'bad' };
const base: Base = raw;
const items = (base as JSDocSignature).typeParameters!;
console.log(`${items.length}`);
