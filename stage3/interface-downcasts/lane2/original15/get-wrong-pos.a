import type { GetAccessorDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 178, typeParameters: [{ kind: 169, pos: 'bad' }] };
const base: Base = raw;
const parameters = (base as GetAccessorDeclaration).typeParameters!;
console.log(`${parameters[0]!.pos}`);
