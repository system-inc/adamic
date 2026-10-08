import type { NewExpression } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 215, typeArguments: [{ kind: 154, pos: 7 }, { kind: 154, pos: 'bad' }] };
const base: Base = raw;
const items = (base as NewExpression).typeArguments!;
console.log(`${items.length}:${items[0]!.pos}`);
