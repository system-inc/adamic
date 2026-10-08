import type { PropertyAssignment } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 304, modifiers: 'bad' };
const base: Base = raw;
const items = (base as PropertyAssignment).modifiers!;
console.log(`${items.length}`);
