import type { Bundle } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Bundle;
 const items = viewed.syntheticLibReferences;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.pos}`);
}
const raw = {kind: 309, syntheticLibReferences: [{pos: 7}]};
function mutate(): void { raw.syntheticLibReferences[0] = {pos: 9}; }
read(raw);
