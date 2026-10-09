import type { EmitNode } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as EmitNode;
 console.log('kept');
}
const raw = {flags: 8, annotatedNodes: 7};
read(raw);
