// #s03ndgw, a stop: name is a property every constructor has (Function's), so the contract admits no class.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly name: string; }
interface View extends Root { readonly value: number | Inner; }
class Klass {
	static readonly tag = 'I' as const;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
console.log('before');
const read = viewed.value;
console.log(typeof read === 'number' ? 'number' : read.name);
