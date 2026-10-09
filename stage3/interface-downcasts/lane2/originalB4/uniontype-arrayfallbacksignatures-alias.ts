import type { UnionType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as UnionType;
 const items = viewed.arrayFallbackSignatures;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, arrayFallbackSignatures: [{flags: 7}]};
function mutate(): void { raw.arrayFallbackSignatures[0] =  {flags: 9}; }
read(raw);
