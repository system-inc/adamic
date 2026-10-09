import type { TypeMapper, TypeMapKind } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Extract<TypeMapper, {kind: TypeMapKind.Deferred}>;
 const items = viewed.targets;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!().flags}`);
}
const raw = {kind: 2, targets: [() => ({flags: 7}), () => ({flags: 'bad'})]};
read(raw);
