import type { EmitNode } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as EmitNode;
 const items = viewed.annotatedNodes;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, annotatedNodes: [{kind: 0, flags: 7}, {kind: 0, flags: 'bad'}]};
read(raw);
