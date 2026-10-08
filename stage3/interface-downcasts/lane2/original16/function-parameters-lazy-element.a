import type { FunctionTypeNode } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 185, parameters: [{ kind: 170, pos: 7 }, { kind: 170, pos: 'bad' }] };
const base: Base = raw;
const items = (base as FunctionTypeNode).parameters;
console.log(`${items.length}:${items[0]!.pos}`);
