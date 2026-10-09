import type { ParsedCommandLine } from 'original-tsc-types';
interface Base { readonly compileOnSave?: boolean; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as ParsedCommandLine;
 const items = viewed?.fileNames;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(items.slice(0, 1).join(';'));
}
const raw = {compileOnSave: false, fileNames: ['a']};
function mutate(): void { raw.fileNames[0] = 'second'; }
read(raw);
