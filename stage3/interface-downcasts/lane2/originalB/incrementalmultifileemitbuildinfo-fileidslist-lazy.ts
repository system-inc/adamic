import type { IncrementalMultiFileEmitBuildInfo } from 'original-tsc-builder';
interface Base { readonly version: string; }
function read(base: Base): void {
 const viewed = base as IncrementalMultiFileEmitBuildInfo;
 console.log('kept');
}
const raw = {version: '1', fileIdsList: 7};
read(raw);
