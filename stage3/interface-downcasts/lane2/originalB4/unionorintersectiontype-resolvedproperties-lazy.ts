import type { UnionOrIntersectionType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as UnionOrIntersectionType;
 console.log('kept');
}
const raw = {flags: 8, resolvedProperties: 7};
read(raw);
