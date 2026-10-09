import type { SyntaxList } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SyntaxList;
 const items = viewed._children;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 353, _children: [{flags: 7}]};
function mutate(): void { raw._children[0] = {flags: 9}; }
read(raw);
