import type { SortedAndCanonicalizedMutableFileSystemEntries } from 'original-tsc-private';
interface Base { readonly files: string[]; }
function read(base: Base): void {
 const viewed = base as SortedAndCanonicalizedMutableFileSystemEntries;
 const items = viewed.directories;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(items.slice(0, 1).join(';'));
}
const raw = {files: ['kept'], directories: ['a']};
function mutate(): void { raw.directories[0] = 'second'; }
read(raw);
