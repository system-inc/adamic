import type { GenericType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): string {
 const items = (base as GenericType).outerTypeParameters;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.id}`;
}
const present = { flags: 1, outerTypeParameters: [{ flags: 262144, id: 7 }] };
console.log(read(present));
const explicit = { flags: 1, outerTypeParameters: undefined };
console.log(read(explicit));
