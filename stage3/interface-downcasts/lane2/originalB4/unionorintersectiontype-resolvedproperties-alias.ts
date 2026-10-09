import type { UnionOrIntersectionType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as UnionOrIntersectionType;
 const items = viewed.resolvedProperties;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, resolvedProperties: [{flags: 7}]};
function mutate(): void { raw.resolvedProperties[0] =  {flags: 9}; }
read(raw);
