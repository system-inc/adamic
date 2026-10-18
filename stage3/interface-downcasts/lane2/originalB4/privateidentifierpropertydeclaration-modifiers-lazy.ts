import type { PrivateIdentifierPropertyDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as PrivateIdentifierPropertyDeclaration;
 console.log('kept');
}
const raw = {kind: 173, modifiers: 7};
read(raw);
