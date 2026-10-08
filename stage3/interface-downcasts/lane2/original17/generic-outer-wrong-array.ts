import type { GenericType } from 'original-tsc-types';
interface Base { readonly flags: number; }
const raw = { flags: 1, outerTypeParameters: 'bad' };
const base: Base = raw;
const items = (base as GenericType).outerTypeParameters!;
console.log(`${items.length}`);
