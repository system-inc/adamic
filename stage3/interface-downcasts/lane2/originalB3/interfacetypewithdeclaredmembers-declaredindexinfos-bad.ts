import type { InterfaceTypeWithDeclaredMembers } from 'original-tsc-types';
interface Base { readonly flags: number; }
function read(base: Base): void {
 const viewed = base as InterfaceTypeWithDeclaredMembers;
 const items = viewed.declaredIndexInfos;
 if (items === undefined) { console.log('absent'); return; }
 console.log(`${items[0]!.isReadonly}`);
}
const raw = {flags: 8, declaredIndexInfos: [{isReadonly: 'bad'}]};
read(raw);
