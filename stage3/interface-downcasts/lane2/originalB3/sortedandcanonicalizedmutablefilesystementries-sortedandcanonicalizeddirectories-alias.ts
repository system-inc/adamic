import type { SortedAndCanonicalizedMutableFileSystemEntries } from 'original-tsc-private';
interface Base { readonly files: string[]; }
function read(base: Base): void {
 const viewed = base as SortedAndCanonicalizedMutableFileSystemEntries;
 const items = viewed.sortedAndCanonicalizedDirectories;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(items.slice(0, 1).join(';'));
}
const raw = {files: ['kept'], sortedAndCanonicalizedDirectories: ['a']};
function mutate(): void { raw.sortedAndCanonicalizedDirectories[0] =  'second'; }
read(raw);
