import type { SignatureDeclaration, JSDocSignature } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as SignatureDeclaration | JSDocSignature;
 console.log('kept');
}
const raw = {kind: 174, parameters: 7};
read(raw);
