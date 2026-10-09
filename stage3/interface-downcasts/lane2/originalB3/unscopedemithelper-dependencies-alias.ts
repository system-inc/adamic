import type { UnscopedEmitHelper } from 'original-tsc-types';
interface Base { readonly scoped: false; }
function read(base: Base): void {
 const viewed = base as UnscopedEmitHelper;
 const items = viewed.dependencies;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.name}`);
}
const raw = {scoped: false as const, dependencies: [{scoped: true, text: 'kept', name: 'a'}]};
function mutate(): void { raw.dependencies[0] =  {scoped: true, text: 'kept', name: 'second'}; }
read(raw);
