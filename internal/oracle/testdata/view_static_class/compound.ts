// #s03ndgw, a stop: a compound assignment through the view writes, so it stops.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; count: number; }
interface View extends Root { readonly value: number | Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static count = 3;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
console.log('before');
const read = viewed.value;
if (typeof read !== 'number') {
	read.count += 1;
	console.log(String(Klass.count));
}
