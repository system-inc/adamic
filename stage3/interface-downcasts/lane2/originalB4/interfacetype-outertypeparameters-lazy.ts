import type { InterfaceType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as InterfaceType;
 console.log('kept');
}
const raw = {flags: 8, outerTypeParameters: 7};
read(raw);
