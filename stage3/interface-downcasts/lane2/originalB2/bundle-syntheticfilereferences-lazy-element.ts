import type { Bundle } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Bundle;
 const items = viewed.syntheticFileReferences;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.pos}`);
}
const raw = {kind: 309, syntheticFileReferences: [{pos: 7}, {pos: 'bad'}]};
read(raw);
