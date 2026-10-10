import type { JSDocFunctionType } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as JSDocFunctionType;
 const items = viewed.typeParameters;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 318, typeParameters: [{kind: 169, flags: 7}]};
function mutate(): void { raw.typeParameters[0] =  {kind: 169, flags: 9}; }
read(raw);
