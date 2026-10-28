interface Node { readonly kind: string; }
interface Base { readonly kind: string; }
interface Box extends Base { readonly kind: 'Box'; readonly value: true | Node | undefined; }
function read(value: Box): string {
 const member = value.value;
 if (member === undefined) return 'absent';
 return typeof member === 'boolean' ? 'true' : member.kind;
}
const source: Base = { kind: 'Box' };
console.log(read(source as Box));
