import type { DiagnosticWithLocation } from 'original-tsc-types';
interface Base { readonly code: number; }
function read(base: Base): void {
 const viewed = base as DiagnosticWithLocation;
 console.log('kept');
}
const raw = {code: 7, relatedInformation: 7};
read(raw);
