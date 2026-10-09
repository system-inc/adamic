import type { ConfigFileSpecs } from 'original-tsc-types';
interface Base { readonly isDefaultIncludeSpec: boolean; }
function read(base: Base): void {
 const viewed = base as ConfigFileSpecs;
 const items = viewed.validatedIncludeSpecsBeforeSubstitution;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(items.slice(0, 1).join(';'));
}
const raw = {isDefaultIncludeSpec: false, validatedIncludeSpecsBeforeSubstitution: ['a']};
function mutate(): void { raw.validatedIncludeSpecsBeforeSubstitution[0] = 'second'; }
read(raw);
