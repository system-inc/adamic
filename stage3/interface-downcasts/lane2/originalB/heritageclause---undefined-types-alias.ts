import type { HeritageClause } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as HeritageClause;
 const items = viewed?.types;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 299, types: [{kind: 234, flags: 7}]};
function mutate(): void { raw.types[0] = {kind: 234, flags: 9}; }
read(raw);
