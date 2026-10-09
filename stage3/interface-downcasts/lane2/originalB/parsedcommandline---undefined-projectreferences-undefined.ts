import type { ParsedCommandLine } from 'original-tsc-types';
interface Base { readonly compileOnSave?: boolean; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as ParsedCommandLine;
 const items = viewed?.projectReferences;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.path}`);
}
const raw = {compileOnSave: false, projectReferences: undefined};
read(raw);
