import type { SyntaxList } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SyntaxList;
 console.log('kept');
}
const raw = {kind: 353, _children: 7};
read(raw);
