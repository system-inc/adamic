import type { UnscopedEmitHelper } from 'original-tsc-types';
interface Base { readonly scoped: false; }
function read(base: Base): void {
 const viewed = base as UnscopedEmitHelper;
 console.log('kept');
}
const raw = {scoped: false as const, dependencies: 7};
read(raw);
