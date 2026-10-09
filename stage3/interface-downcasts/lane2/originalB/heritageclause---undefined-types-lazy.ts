import type { HeritageClause } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as HeritageClause;
 console.log('kept');
}
const raw = {kind: 299, types: 7};
read(raw);
