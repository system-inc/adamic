import type { IndexSignatureDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as IndexSignatureDeclaration;
 console.log('kept');
}
const raw = {kind: 182, typeParameters: 7};
read(raw);
