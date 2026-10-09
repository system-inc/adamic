import type { Signature } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as Signature;
 const items = viewed?.typeParameters;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, typeParameters: [{flags: 'bad'}]};
read(raw);
