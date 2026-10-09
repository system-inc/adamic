import type { EmitNode } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as EmitNode;
 const items = viewed?.trailingComments;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.kind}`);
}
const raw = {flags: 8, trailingComments: undefined};
read(raw);
