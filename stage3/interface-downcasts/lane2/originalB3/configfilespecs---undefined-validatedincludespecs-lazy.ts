import type { ConfigFileSpecs } from 'original-tsc-types';
interface Base { readonly isDefaultIncludeSpec: boolean; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as ConfigFileSpecs;
 console.log('kept');
}
const raw = {isDefaultIncludeSpec: false, validatedIncludeSpecs: 7};
read(raw);
