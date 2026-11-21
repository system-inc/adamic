// #s03ndgw, a stop: Klass owns a data static named call, which Node reads, but call is a property every
// constructor has (Function's), so the contract admits no class and the view's read stops.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly call: number; }
interface View extends Root { readonly value: number | Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static readonly call = 5;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
console.log('before');
const read = viewed.value;
console.log(typeof read === 'number' ? 'number' : String(read.call));
