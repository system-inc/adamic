import type { SetAccessorDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as SetAccessorDeclaration;
 console.log('kept');
}
const raw = {kind: 179, parameters: 7};
read(raw);
