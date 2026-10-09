import type { WatchOptions } from 'original-tsc-types';
interface Base { readonly watchFile?: number; }
function read(base: Base): void {
 const viewed = base as WatchOptions;
 console.log('kept');
}
const raw = {watchFile: 0, excludeDirectories: 7};
read(raw);
