import type { IntersectionTypeNode } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 194, types: [{ kind: 154, pos: 7 }, { kind: 154, pos: 'bad' }] };
const base: Base = raw;
const items = (base as IntersectionTypeNode).types;
console.log(`${items.length}:${items[0]!.pos}`);
