import type { TupleType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as TupleType;
 console.log('kept');
}
const raw = {flags: 8, typeParameters: 7};
read(raw);
