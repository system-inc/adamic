import type { IncrementalMultiFileEmitBuildInfo } from 'original-tsc-builder';
interface Base { readonly version: string; }
function read(base: Base): void {
 const viewed = base as IncrementalMultiFileEmitBuildInfo;
 const items = viewed.fileIdsList;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]![0]!}`);
}
const raw = {version: '1', fileIdsList: [[7]]};
function mutate(): void { raw.fileIdsList[0] = [9]; }
read(raw);
