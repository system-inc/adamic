import type { JSDocFunctionType } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as JSDocFunctionType;
 console.log('kept');
}
const raw = {kind: 318, typeParameters: 7};
read(raw);
