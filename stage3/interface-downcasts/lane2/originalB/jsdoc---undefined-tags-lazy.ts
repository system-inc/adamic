import type { JSDoc } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as JSDoc;
 console.log('kept');
}
const raw = {kind: 321, tags: 7};
read(raw);
