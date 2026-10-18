import type { CompilerOptions } from 'original-tsc-types';
interface Base { readonly strict?: boolean; }
function read(base: Base): void {
 const viewed = base as CompilerOptions;
 const items = viewed.types;
 if (items === undefined) { console.log('absent'); return; }
 console.log(items.slice(0, 1).join(';'));
}
const raw = {strict: false, types: [42]};
read(raw);
