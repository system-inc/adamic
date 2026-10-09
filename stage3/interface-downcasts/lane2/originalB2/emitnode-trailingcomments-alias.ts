import type { EmitNode } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as EmitNode;
 const items = viewed.trailingComments;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.kind}`);
}
const raw = {flags: 8, trailingComments: [{kind: 2}]};
function mutate(): void { raw.trailingComments[0] = {kind: 3}; }
read(raw);
