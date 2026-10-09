import type { TypeMapper, TypeMapKind } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Extract<TypeMapper, {kind: TypeMapKind.Deferred}>;
 const items = viewed.sources;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 2, sources: [{flags: 7}]};
read(raw);
