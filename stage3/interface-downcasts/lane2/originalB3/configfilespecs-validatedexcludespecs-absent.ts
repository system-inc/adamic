import type { ConfigFileSpecs } from 'original-tsc-types';
interface Base { readonly isDefaultIncludeSpec: boolean; }
function read(base: Base): void {
 const viewed = base as ConfigFileSpecs;
 const items = viewed.validatedExcludeSpecs;
 if (items === undefined) { console.log('absent'); return; }
 console.log(items.slice(0, 1).join(';'));
}
const raw = {isDefaultIncludeSpec: false, };
read(raw);
