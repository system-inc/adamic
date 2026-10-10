import type { SourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SourceFile;
 const items = viewed.additionalSyntacticDiagnostics;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.code}`);
}
const raw = {kind: 308, };
read(raw);
