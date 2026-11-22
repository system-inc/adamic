// #s03ndgw, a stop past the selection: call is a property every constructor has (Function's), but Inner
// requires it, so its contract admits a class and Klass's own data static named call selects Inner, as
// Node reads Klass.call. The read of call is through a local typeof narrowed, which stops as
// narrowed_read does, until #775qbxz. call_inherited reads one at a field typed Inner itself.
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
