import type { JSDocTemplateTag } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as JSDocTemplateTag).comment;
 if (items === undefined) return 'absent';
 if (typeof items === 'string') return items;
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 346, comment: [{ kind: 322, pos: 7 }] };
console.log(read(present));
const absent = { kind: 346 };
console.log(read(absent));
const explicit = { kind: 346, comment: undefined };
console.log(read(explicit));
