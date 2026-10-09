import type { InterfaceType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as InterfaceType;
 const items = viewed.outerTypeParameters;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, outerTypeParameters: undefined};
read(raw);
