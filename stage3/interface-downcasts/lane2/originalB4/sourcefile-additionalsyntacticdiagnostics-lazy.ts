import type { SourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SourceFile;
 console.log('kept');
}
const raw = {kind: 308, additionalSyntacticDiagnostics: 7};
read(raw);
