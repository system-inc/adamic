import type { SignatureDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SignatureDeclaration;
 const items = viewed.jsDoc;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 174, jsDoc: [{kind: 321, flags: 7}]};
read(raw);
