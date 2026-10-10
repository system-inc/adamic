import type { SourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as SourceFile;
 const items = viewed?.imports;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 308, imports: [{kind: 11, flags: 7}]};
read(raw);
