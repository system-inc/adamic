import type { GenericType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): string {
 const items = (base as GenericType).localTypeParameters;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.id}`;
}
const present = { flags: 1, localTypeParameters: [{ flags: 262144, id: 7 }] };
console.log(read(present));
const explicit = { flags: 1, localTypeParameters: undefined };
console.log(read(explicit));
