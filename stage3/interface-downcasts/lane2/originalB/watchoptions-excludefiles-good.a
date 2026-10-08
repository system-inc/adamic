import type { WatchOptions } from 'original-tsc-types';
interface Base { readonly watchFile?: number; }
function read(base: Base): void {
 const viewed = base as WatchOptions;
 const items = viewed.excludeFiles;
 if (items === undefined) { console.log('absent'); return; }
 console.log(items.slice(0, 1).join(';'));
}
const raw = {watchFile: 0, excludeFiles: ['a']};
read(raw);
