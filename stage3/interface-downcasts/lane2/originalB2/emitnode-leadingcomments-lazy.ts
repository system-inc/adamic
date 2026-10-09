import type { EmitNode } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as EmitNode;
 console.log('kept');
}
const raw = {flags: 8, leadingComments: 7};
read(raw);
