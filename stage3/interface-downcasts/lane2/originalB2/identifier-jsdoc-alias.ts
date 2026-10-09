import type { Identifier } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Identifier;
 const items = viewed.jsDoc;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 80, jsDoc: [{kind: 321, flags: 7}]};
function mutate(): void { raw.jsDoc[0] = {kind: 321, flags: 9}; }
read(raw);
