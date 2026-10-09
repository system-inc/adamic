import type { WatchOptions } from 'original-tsc-types';
interface Base { readonly watchFile?: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as WatchOptions;
 const items = viewed?.excludeDirectories;
 if (items === undefined) { console.log('absent'); return; }
 console.log(items.slice(0, 1).join(';'));
}
const raw = {watchFile: 0, excludeDirectories: ['a', 42]};
read(raw);
