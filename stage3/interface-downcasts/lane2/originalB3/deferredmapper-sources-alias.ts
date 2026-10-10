import type { TypeMapper, TypeMapKind } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Extract<TypeMapper, {kind: TypeMapKind.Deferred}>;
 const items = viewed.sources;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 2, sources: [{flags: 7}]};
function mutate(): void { raw.sources[0] =  {flags: 9}; }
read(raw);
