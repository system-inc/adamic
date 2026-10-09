import type { UnionType } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as UnionType;
 console.log('kept');
}
const raw = {flags: 8, arrayFallbackSignatures: 7};
read(raw);
