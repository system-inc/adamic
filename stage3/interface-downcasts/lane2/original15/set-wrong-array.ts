import type { SetAccessorDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 179, typeParameters: 'bad' };
const base: Base = raw;
const parameters = (base as SetAccessorDeclaration).typeParameters!;
console.log(`${parameters.length}`);
