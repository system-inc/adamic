// #s03ndgw, a stop: a logical assignment through the view that takes its write evaluates the right side,
// then stops before writing.
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
viewed.value.count &&= noisy();
console.log(String(Klass.count));
