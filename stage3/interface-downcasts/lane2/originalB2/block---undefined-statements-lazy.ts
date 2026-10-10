import type { Block } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as Block;
 console.log('kept');
}
const raw = {kind: 242, statements: 7};
read(raw);
