import type { Identifier } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Identifier;
 console.log('kept');
}
const raw = {kind: 80, jsDoc: 7};
read(raw);
