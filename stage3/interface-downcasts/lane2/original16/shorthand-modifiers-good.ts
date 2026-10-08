import type { ShorthandPropertyAssignment } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as ShorthandPropertyAssignment).modifiers;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 305, modifiers: [{ kind: 95, pos: 7 }] };
console.log(read(present));
const absent = { kind: 305 };
console.log(read(absent));
const explicit = { kind: 305, modifiers: undefined };
console.log(read(explicit));
