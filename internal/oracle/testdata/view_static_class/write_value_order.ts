// #s03ndgw, a stop: a write through the view evaluates its receiver, then its value, then stops before
// writing, on both backends: value prints first, as it would on Node.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: 'I'; count: number; }
interface Viewed extends Root { readonly value: Inner; }
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
const viewed = base as Viewed;
console.log('before');
viewed.value.count = noisy();
console.log(String(Klass.count));
