// #s03ndgw: a class constructor held where a checked view expects number | Inner, every field Inner names
// a data static the class owns. It stays the class, so typeof says function, as Node says (p67's shape).
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; }
interface View extends Root { readonly value: number | Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static readonly count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
const read = viewed.value;
console.log(typeof read === 'object' ? `${read.count}` : 'number');
