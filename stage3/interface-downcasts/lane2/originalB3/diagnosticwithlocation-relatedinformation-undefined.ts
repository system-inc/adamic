import type { DiagnosticWithLocation } from 'original-tsc-types';
interface Base { readonly code: number; }
function read(base: Base): void {
 const viewed = base as DiagnosticWithLocation;
 const items = viewed.relatedInformation;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.code}`);
}
const raw = {code: 7, relatedInformation: undefined};
read(raw);
