import type { JSDocTypeLiteral } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as JSDocTypeLiteral).jsDocPropertyTags;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 323, jsDocPropertyTags: [{ kind: 349, pos: 7 }] };
console.log(read(present));
const absent = { kind: 323 };
console.log(read(absent));
const explicit = { kind: 323, jsDocPropertyTags: undefined };
console.log(read(explicit));
