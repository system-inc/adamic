import type { GenericType } from 'original-tsc-types';
interface Base { readonly flags: number; }
const raw = { flags: 1, outerTypeParameters: [{ flags: 262144, id: 'bad' }] };
const base: Base = raw;
const items = (base as GenericType).outerTypeParameters!;
console.log(`${items[0]!.id}`);
