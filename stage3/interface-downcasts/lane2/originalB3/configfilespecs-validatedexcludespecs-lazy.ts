import type { ConfigFileSpecs } from 'original-tsc-types';
interface Base { readonly isDefaultIncludeSpec: boolean; }
function read(base: Base): void {
 const viewed = base as ConfigFileSpecs;
 console.log('kept');
}
const raw = {isDefaultIncludeSpec: false, validatedExcludeSpecs: 7};
read(raw);
