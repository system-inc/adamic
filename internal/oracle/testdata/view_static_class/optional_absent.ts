// #s03ndgw: an optional field the class doesn't declare is absent, and reads undefined, as Node reads it.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; readonly count: number; readonly extra?: number; }
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
	console.log(String(read.count) + ' ' + String(read.extra));
}
