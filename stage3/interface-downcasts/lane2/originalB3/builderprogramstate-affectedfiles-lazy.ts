import type { BuilderProgramState } from 'original-tsc-builder';
interface Base { readonly buildInfoEmitPending: boolean; }
function read(base: Base): void {
 const viewed = base as BuilderProgramState;
 console.log('kept');
}
const raw = {buildInfoEmitPending: false, affectedFiles: 7};
read(raw);
