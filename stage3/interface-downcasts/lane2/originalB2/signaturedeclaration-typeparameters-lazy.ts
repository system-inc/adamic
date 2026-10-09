import type { SignatureDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SignatureDeclaration;
 console.log('kept');
}
const raw = {kind: 174, typeParameters: 7};
read(raw);
