import type { IndexSignatureDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as IndexSignatureDeclaration;
 const items = viewed.typeParameters;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 182, typeParameters: undefined};
read(raw);
