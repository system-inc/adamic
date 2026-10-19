import type { EmitNode } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as EmitNode;
 const items = viewed.leadingComments;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.kind}`);
}
const raw = {flags: 8, leadingComments: undefined};
read(raw);
