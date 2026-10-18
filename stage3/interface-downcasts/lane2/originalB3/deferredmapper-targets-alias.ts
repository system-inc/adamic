import type { TypeMapper, TypeMapKind } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Extract<TypeMapper, {kind: TypeMapKind.Deferred}>;
 const items = viewed.targets;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!().flags}`);
}
const result = {flags: 7};
const raw = {kind: 2, targets: [() => result]};
function mutate(): void { result.flags = 9; }
read(raw);
