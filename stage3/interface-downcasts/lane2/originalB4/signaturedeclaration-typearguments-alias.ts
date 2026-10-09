import type { SignatureDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SignatureDeclaration;
 const items = viewed.typeArguments;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 174, typeArguments: [{kind: 116, flags: 7}]};
function mutate(): void { raw.typeArguments[0] =  {kind: 116, flags: 9}; }
read(raw);
