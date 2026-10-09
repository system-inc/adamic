import type { CompilerOptions } from 'original-tsc-types';
interface Base { readonly strict?: boolean; }
function read(base: Base): void {
 const viewed = base as CompilerOptions;
 console.log('kept');
}
const raw = {strict: false, customConditions: 7};
read(raw);
