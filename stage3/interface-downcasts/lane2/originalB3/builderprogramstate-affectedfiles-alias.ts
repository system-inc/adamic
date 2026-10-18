import type { BuilderProgramState } from 'original-tsc-builder';
interface Base { readonly buildInfoEmitPending: boolean; }
function read(base: Base): void {
 const viewed = base as BuilderProgramState;
 const items = viewed.affectedFiles;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {buildInfoEmitPending: false, affectedFiles: [{kind: 308, flags: 7}]};
function mutate(): void { raw.affectedFiles[0] =  {kind: 308, flags: 9}; }
read(raw);
