import type { GetAccessorDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 178, typeParameters: 'bad' };
const base: Base = raw;
const parameters = (base as GetAccessorDeclaration).typeParameters!;
console.log(`${parameters.length}`);
