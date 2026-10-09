import type { SortedAndCanonicalizedMutableFileSystemEntries } from 'original-tsc-private';
interface Base { readonly files: string[]; }
function read(base: Base): void {
 const viewed = base as SortedAndCanonicalizedMutableFileSystemEntries;
 console.log('kept');
}
const raw = {files: ['kept'], sortedAndCanonicalizedDirectories: 7};
read(raw);
