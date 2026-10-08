import type { JSDocCallbackTag } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 339, comment: [{ kind: 322, pos: 7 }, { kind: 322, pos: 'bad' }] };
const base: Base = raw;
const items = (base as JSDocCallbackTag).comment!;
if (typeof items !== 'string') {
console.log(`${items.length}:${items[0]!.pos}`);
}
