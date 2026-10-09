import type { InterfaceTypeWithDeclaredMembers } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as InterfaceTypeWithDeclaredMembers;
 const items = viewed.declaredCallSignatures;
 if (items === undefined) { console.log('absent'); return; }
 mutate();
 console.log(`${items[0]!.flags}`);
}
const raw = {flags: 8, declaredCallSignatures: [{flags: 7}]};
function mutate(): void { raw.declaredCallSignatures[0] =  {flags: 9}; }
read(raw);
