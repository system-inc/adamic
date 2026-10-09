import type { TupleType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as TupleType;
 const items = viewed.typeParameters;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, typeParameters: [{flags: 7}]};
function mutate(): void { raw.typeParameters[0] = {flags: 9}; }
read(raw);
