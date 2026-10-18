import type { WatchOptions } from 'original-tsc-types';
interface Base { readonly watchFile?: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as WatchOptions;
 console.log('kept');
}
const raw = {watchFile: 0, excludeFiles: 7};
read(raw);
