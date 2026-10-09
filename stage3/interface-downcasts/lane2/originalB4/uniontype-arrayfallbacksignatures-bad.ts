import type { UnionType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as UnionType;
 const items = viewed.arrayFallbackSignatures;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, arrayFallbackSignatures: [{flags: 'bad'}]};
read(raw);
