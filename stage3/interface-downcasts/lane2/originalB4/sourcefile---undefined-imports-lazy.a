import type { SourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as SourceFile;
 console.log('kept');
}
const raw = {kind: 308, imports: 7};
read(raw);
