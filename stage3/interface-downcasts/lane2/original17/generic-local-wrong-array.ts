import type { GenericType } from 'original-tsc-types';
interface Base { readonly flags: number; }
const raw = { flags: 1, localTypeParameters: 'bad' };
const base: Base = raw;
const items = (base as GenericType).localTypeParameters!;
console.log(`${items.length}`);
