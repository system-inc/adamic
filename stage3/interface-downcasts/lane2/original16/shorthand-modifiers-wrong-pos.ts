import type { ShorthandPropertyAssignment } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 305, modifiers: [{ kind: 95, pos: 'bad' }] };
const base: Base = raw;
const items = (base as ShorthandPropertyAssignment).modifiers!;
console.log(`${items[0]!.pos}`);
