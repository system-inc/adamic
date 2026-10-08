import type { FunctionExpression } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 219, typeParameters: [{ kind: 169 }] };
const base: Base = raw;
const parameters = (base as FunctionExpression).typeParameters!;
console.log(`${parameters[0]!.pos}`);
