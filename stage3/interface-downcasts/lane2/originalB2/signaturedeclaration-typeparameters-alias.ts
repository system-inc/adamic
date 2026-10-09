import type { SignatureDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SignatureDeclaration;
 const items = viewed.typeParameters;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 174, typeParameters: [{kind: 169, flags: 7}]};
function mutate(): void { raw.typeParameters[0] = {kind: 169, flags: 9}; }
read(raw);
