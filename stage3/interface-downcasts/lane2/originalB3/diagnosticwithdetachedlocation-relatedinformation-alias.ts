import type { DiagnosticWithDetachedLocation } from 'original-tsc-types';
interface Base { readonly code: number; }
function read(base: Base): void {
 const viewed = base as DiagnosticWithDetachedLocation;
 const items = viewed.relatedInformation;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.code}`);
}
const raw = {code: 7, relatedInformation: [{code: 7}]};
function mutate(): void { raw.relatedInformation[0] =  {code: 9}; }
read(raw);
