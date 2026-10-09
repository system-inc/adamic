import type { AnonymousType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as AnonymousType;
 console.log('kept');
}
const raw = {flags: 8, callSignatures: 7};
read(raw);
