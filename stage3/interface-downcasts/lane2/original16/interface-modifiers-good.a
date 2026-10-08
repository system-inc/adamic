import type { InterfaceDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as InterfaceDeclaration).modifiers;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 265, modifiers: [{ kind: 95, pos: 7 }] };
console.log(read(present));
const absent = { kind: 265 };
console.log(read(absent));
const explicit = { kind: 265, modifiers: undefined };
console.log(read(explicit));
