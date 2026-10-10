import type { SignatureDeclaration, JSDocSignature } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base | undefined): void {
 const viewed = base === undefined ? undefined : base as SignatureDeclaration | JSDocSignature;
 const items = viewed?.parameters;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 174, parameters: [{kind: 170, flags: 7}]};
read(raw);
