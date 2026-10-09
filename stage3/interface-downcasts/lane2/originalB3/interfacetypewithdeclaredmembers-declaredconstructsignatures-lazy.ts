import type { InterfaceTypeWithDeclaredMembers } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as InterfaceTypeWithDeclaredMembers;
 console.log('kept');
}
const raw = {flags: 8, declaredConstructSignatures: 7};
read(raw);
