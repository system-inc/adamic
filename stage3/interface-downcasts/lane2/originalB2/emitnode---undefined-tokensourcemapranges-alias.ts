import type { EmitNode } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as EmitNode;
 const items = viewed?.tokenSourceMapRanges;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.pos}`);
}
const raw = {flags: 8, tokenSourceMapRanges: [{pos: 7}]};
function mutate(): void { raw.tokenSourceMapRanges[0] = {pos: 9}; }
read(raw);
