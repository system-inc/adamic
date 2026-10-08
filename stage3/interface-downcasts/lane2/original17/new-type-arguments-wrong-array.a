import type { NewExpression } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 215, typeArguments: 'bad' };
const base: Base = raw;
const items = (base as NewExpression).typeArguments!;
console.log(`${items.length}`);
