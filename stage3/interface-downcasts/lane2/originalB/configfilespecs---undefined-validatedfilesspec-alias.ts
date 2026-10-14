import type { ConfigFileSpecs } from 'original-tsc-types';
interface Base { readonly isDefaultIncludeSpec: boolean; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as ConfigFileSpecs;
 const items = viewed?.validatedFilesSpec;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(items.slice(0, 1).join(';'));
}
const raw = {isDefaultIncludeSpec: false, validatedFilesSpec: ['a']};
function mutate(): void { raw.validatedFilesSpec[0] = 'second'; }
read(raw);
