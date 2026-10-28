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
const text = { flags: 1, value: 'text' };
const number = { flags: 1, value: 42 };
const bigint = { flags: 1, value: { negative: true, base10Value: '123' } };
const textSource: Type = text;
const numberSource: Type = number;
const bigintSource: Type = bigint;
console.log(read(textSource as LiteralType));
console.log(read(numberSource as LiteralType));
console.log(read(bigintSource as LiteralType));
const generatedNumber = 'x'.length === 1 ? 43 : { negative: false, base10Value: '0' };
const generatedObject = 'x'.length === 2 ? 44 : { negative: false, base10Value: '456' };
const numberRecord = { flags: 1, value: generatedNumber };
const objectRecord = { flags: 1, value: generatedObject };
const generatedNumberSource: Type = numberRecord;
const generatedObjectSource: Type = objectRecord;
console.log(read(generatedNumberSource as LiteralType));
console.log(read(generatedObjectSource as LiteralType));
