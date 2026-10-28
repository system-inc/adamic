interface PseudoBigInt { readonly negative: boolean; readonly base10Value: string; }
interface Type { readonly flags: number; }
interface LiteralType extends Type { readonly value: string | number | PseudoBigInt; }
function read(type: LiteralType): string {
 const member = type.value;
 if (typeof member === 'string') return member;
 if (typeof member === 'number') return `${member}`;
 if (typeof member === 'boolean') return 'wrong';
 return `${member.negative}:${member.base10Value}`;
}
const concrete = { flags: 1, value: { negative: false, base10Value: 123 } };
const source: Type = concrete;
console.log(read(source as LiteralType));
