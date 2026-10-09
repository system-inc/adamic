import type { PrivateIdentifierPropertyDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as PrivateIdentifierPropertyDeclaration;
 const items = viewed.modifiers;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.flags}`);
}
const raw = {kind: 173, modifiers: [{kind: 128, flags: 7}, {kind: 128, flags: 'bad'}]};
read(raw);
