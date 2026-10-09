import type { AnonymousType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as AnonymousType;
 const items = viewed.callSignatures;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, callSignatures: [{flags: 7}]};
function mutate(): void { raw.callSignatures[0] = {flags: 9}; }
read(raw);
