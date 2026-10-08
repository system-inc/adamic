import type { ShorthandPropertyAssignment } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 305, modifiers: 'bad' };
const base: Base = raw;
const items = (base as ShorthandPropertyAssignment).modifiers!;
console.log(`${items.length}`);
