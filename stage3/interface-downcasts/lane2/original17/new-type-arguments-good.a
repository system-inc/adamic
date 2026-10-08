import type { NewExpression } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as NewExpression).typeArguments;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 215, typeArguments: [{ kind: 154, pos: 7 }] };
console.log(read(present));
const absent = { kind: 215 };
console.log(read(absent));
const explicit = { kind: 215, typeArguments: undefined };
console.log(read(explicit));
