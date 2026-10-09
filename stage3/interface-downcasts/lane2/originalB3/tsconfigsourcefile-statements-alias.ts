import type { TsConfigSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as TsConfigSourceFile;
 const items = viewed.statements;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 308, statements: [{kind: 245, flags: 7}]};
function mutate(): void { raw.statements[0] =  {kind: 245, flags: 9}; }
read(raw);
