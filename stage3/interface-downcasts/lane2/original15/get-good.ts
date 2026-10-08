import type { GetAccessorDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(base: Base): string {
 const parameters = (base as GetAccessorDeclaration).typeParameters;
 if (parameters === undefined) return 'absent';
 return `${parameters.length}:${parameters[0]!.pos}`;
}
const present = { kind: 178, typeParameters: [{ kind: 169, pos: 7 }] };
const absent = { kind: 178 };
const explicit = { kind: 178, typeParameters: undefined };
console.log(read(present));
console.log(read(absent));
console.log(read(explicit));
