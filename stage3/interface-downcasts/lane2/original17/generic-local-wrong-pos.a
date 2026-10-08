import type { GenericType } from 'original-tsc-types';
interface Base { readonly flags: number; }
const raw = { flags: 1, localTypeParameters: [{ flags: 262144, id: 'bad' }] };
const base: Base = raw;
const items = (base as GenericType).localTypeParameters!;
console.log(`${items[0]!.id}`);
