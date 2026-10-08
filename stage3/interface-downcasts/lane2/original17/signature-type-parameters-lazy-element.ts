import type { JSDocSignature } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 324, typeParameters: [{ kind: 346, pos: 7 }, { kind: 346, pos: 'bad' }] };
const base: Base = raw;
const items = (base as JSDocSignature).typeParameters!;
console.log(`${items.length}:${items[0]!.pos}`);
