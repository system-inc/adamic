import type { JSDocCallbackTag } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as JSDocCallbackTag).comment;
 if (items === undefined) return 'absent';
 if (typeof items === 'string') return items;
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 339, comment: [{ kind: 322, pos: 7 }] };
console.log(read(present));
const absent = { kind: 339 };
console.log(read(absent));
const explicit = { kind: 339, comment: undefined };
console.log(read(explicit));
