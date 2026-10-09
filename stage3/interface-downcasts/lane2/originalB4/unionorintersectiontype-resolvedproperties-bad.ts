import type { UnionOrIntersectionType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as UnionOrIntersectionType;
 const items = viewed.resolvedProperties;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, resolvedProperties: [{flags: 'bad'}]};
read(raw);
