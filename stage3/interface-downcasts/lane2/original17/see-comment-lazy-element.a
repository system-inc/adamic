import type { JSDocSeeTag } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 348, comment: [{ kind: 322, pos: 7 }, { kind: 322, pos: 'bad' }] };
const base: Base = raw;
const items = (base as JSDocSeeTag).comment!;
if (typeof items !== 'string') {
console.log(`${items.length}:${items[0]!.pos}`);
}
