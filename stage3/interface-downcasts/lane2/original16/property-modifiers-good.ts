import type { PropertyAssignment } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as PropertyAssignment).modifiers;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 304, modifiers: [{ kind: 95, pos: 7 }] };
console.log(read(present));
const absent = { kind: 304 };
console.log(read(absent));
const explicit = { kind: 304, modifiers: undefined };
console.log(read(explicit));
