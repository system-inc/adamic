import type { JSDocCallbackTag } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 339, comment: true };
const base: Base = raw;
const items = (base as JSDocCallbackTag).comment!;
if (typeof items === 'string') {
 console.log(`${items.length}`);
} else {
 console.log(`${items.length}`);
}
