import type { ConfigFileSpecs } from 'original-tsc-types';
interface Base { readonly isDefaultIncludeSpec: boolean; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as ConfigFileSpecs;
 const items = viewed?.validatedFilesSpec;
 if (items === undefined) { console.log('absent'); return; }
 console.log(items.slice(0, 1).join(';'));
}
const raw = {isDefaultIncludeSpec: false, validatedFilesSpec: [42]};
read(raw);
