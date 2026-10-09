import type { SourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SourceFile;
 const items = viewed.jsDocDiagnostics;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.code}`);
}
const raw = {kind: 308, jsDocDiagnostics: [{code: 7}]};
function mutate(): void { raw.jsDocDiagnostics[0] = {code: 9}; }
read(raw);
