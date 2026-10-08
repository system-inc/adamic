import type { FunctionTypeNode } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const items = (base as FunctionTypeNode).parameters;
 return `${items.length}:${items[0]!.pos}`;
}
const present = { kind: 185, parameters: [{ kind: 170, pos: 7 }] };
console.log(read(present));
