// #s03ndgw: past typeof's number test, a read of Inner's fields through the view reads the class's statics,
// as Node reads Klass.count and Klass.tag.
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
if (typeof read !== 'number') {
	console.log(String(read.count));
	console.log(read.tag);
}
