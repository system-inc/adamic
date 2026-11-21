// Compiler's probe on #s03ndgw: the view's static is a getter.
interface Root { readonly kind: 'Holder'; }
interface Inner { readonly tag: string; readonly count: number; }
interface Viewed extends Root { readonly value: Inner; }
class Klass {
	static readonly tag = 'G';
	static get count(): number {
		console.log('getter ran');
		return 5;
	}
}
const raw = { kind: 'Holder' as const, value: Klass };
const base: Root = raw;
const viewed = base as Viewed;
console.log('before');
const inner = viewed.value;
console.log(String(inner.count));
