import type { IntersectionTypeNode } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as IntersectionTypeNode).types;
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 194, types: [{ kind: 154, pos: 7 }] };
console.log(read(present));
