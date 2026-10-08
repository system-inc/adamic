import type { JSDocSeeTag } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 348, comment: 'ok' };
const base: Base = raw;
const items = (base as JSDocSeeTag).comment!;
if (typeof items === 'string') {
 console.log(`${items.length}`);
} else {
 console.log(`${items.length}`);
}
