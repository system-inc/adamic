import type { TypeMapper, TypeMapKind } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as Extract<TypeMapper, {kind: TypeMapKind.Deferred}>;
 const items = viewed.targets;
 mutate();
 console.log(`${items[0]!().flags}`);
}
const raw = {kind: 2, targets: [() => ({flags: 7})]};
function mutate(): void { raw.targets[0] = () => ({flags: 9}); }
read(raw);
