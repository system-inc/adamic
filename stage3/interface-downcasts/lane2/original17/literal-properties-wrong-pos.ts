import type { JSDocTypeLiteral } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 323, jsDocPropertyTags: [{ kind: 349, pos: 'bad' }] };
const base: Base = raw;
const items = (base as JSDocTypeLiteral).jsDocPropertyTags!;
console.log(`${items[0]!.pos}`);
