import type { NonIncrementalBuildInfo } from 'original-tsc-builder';
interface Base { readonly version: string; }
function read(base: Base): void {
 const viewed = base as NonIncrementalBuildInfo;
 const items = viewed.root;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(items.slice(0, 1).join(';'));
}
const raw = {version: '1', root: ['a']};
function mutate(): void { raw.root[0] =  'second'; }
read(raw);
