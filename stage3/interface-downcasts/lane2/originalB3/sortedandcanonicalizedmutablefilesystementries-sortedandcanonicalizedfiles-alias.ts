import type { SortedAndCanonicalizedMutableFileSystemEntries } from 'original-tsc-private';
interface Base { readonly files: string[]; }
function read(base: Base): void {
 const viewed = base as SortedAndCanonicalizedMutableFileSystemEntries;
 const items = viewed.sortedAndCanonicalizedFiles;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(items.slice(0, 1).join(';'));
}
const raw = {files: ['kept'], sortedAndCanonicalizedFiles: ['a']};
function mutate(): void { raw.sortedAndCanonicalizedFiles[0] =  'second'; }
read(raw);
