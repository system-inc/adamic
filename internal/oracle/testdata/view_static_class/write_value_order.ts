// #s03ndgw, a stop: a write through the view evaluates its receiver, then its value, then stops before
// writing, on both backends: value prints first, as it would on Node.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; count: number; }
interface View extends Root { readonly value: number | Inner; }
class Klass {
	static readonly tag = 'I' as const;
	static count = 3;
}
function noisy(): number {
	console.log('value');
	return 4;
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as View;
console.log('before');
const read = viewed.value;
if (typeof read !== 'number') {
	read.count = noisy();
	console.log(String(Klass.count));
}
