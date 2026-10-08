import type { JSDocSignature } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as JSDocSignature).typeParameters;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 324, typeParameters: [{ kind: 346, pos: 7 }] };
console.log(read(present));
const absent = { kind: 324 };
console.log(read(absent));
const explicit = { kind: 324, typeParameters: undefined };
console.log(read(explicit));
