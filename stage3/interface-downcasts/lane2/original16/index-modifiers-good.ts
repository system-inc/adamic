import type { IndexSignatureDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as IndexSignatureDeclaration).modifiers;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 182, modifiers: [{ kind: 95, pos: 7 }] };
console.log(read(present));
const absent = { kind: 182 };
console.log(read(absent));
const explicit = { kind: 182, modifiers: undefined };
console.log(read(explicit));
