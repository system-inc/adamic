import type { PropertyAssignment } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 304, modifiers: [{ kind: 95, pos: 'bad' }] };
const base: Base = raw;
const items = (base as PropertyAssignment).modifiers!;
console.log(`${items[0]!.pos}`);
