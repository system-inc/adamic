import type { NonIncrementalBuildInfo } from 'original-tsc-builder';
interface Base { readonly version: string; }
function read(base: Base): void {
 const viewed = base as NonIncrementalBuildInfo;
 console.log('kept');
}
const raw = {version: '1', root: 7};
read(raw);
