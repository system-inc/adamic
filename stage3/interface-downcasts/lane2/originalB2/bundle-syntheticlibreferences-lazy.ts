import type { Bundle } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Bundle;
 console.log('kept');
}
const raw = {kind: 309, syntheticLibReferences: 7};
read(raw);
