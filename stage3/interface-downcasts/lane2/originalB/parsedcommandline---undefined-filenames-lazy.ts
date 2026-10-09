import type { ParsedCommandLine } from 'original-tsc-types';
interface Base { readonly compileOnSave?: boolean; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as ParsedCommandLine;
 console.log('kept');
}
const raw = {compileOnSave: false, fileNames: 7};
read(raw);
