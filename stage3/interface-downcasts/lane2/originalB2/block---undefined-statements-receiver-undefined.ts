import type { Block } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as Block;
 const items = viewed?.statements;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 242, statements: [{kind: 0, flags: 7}]};
read(undefined);
