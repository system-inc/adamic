import type { UnscopedEmitHelper } from 'original-tsc-types';
interface Base { readonly scoped: false; }
function read(base: Base): void {
 const viewed = base as UnscopedEmitHelper;
 const items = viewed.dependencies;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.name}`);
}
const raw = {scoped: false as const, dependencies: [{scoped: true, text: 'kept', name: 'a'}, {scoped: true, text: 'kept', name: 42}]};
read(raw);
