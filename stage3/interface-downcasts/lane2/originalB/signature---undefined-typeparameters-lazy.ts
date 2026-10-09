import type { Signature } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as Signature;
 console.log('kept');
}
const raw = {flags: 8, typeParameters: 7};
read(raw);
