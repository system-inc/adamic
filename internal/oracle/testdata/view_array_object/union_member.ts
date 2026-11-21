// #s03ndgw: an array held where a checked view reads number | Loose is the Loose member, as the union's
// selection admits it by Loose's contract, and its field reads undefined past the typeof narrowing.
interface Holder { readonly kind: 'h' | 'g'; }
interface Loose { readonly a?: number; }
interface Boxed extends Holder { readonly kind: 'h'; readonly inner: number | Loose; }
const raw = { kind: 'h' as const, inner: [3] };
const holder: Holder = raw;
const inner = (holder as Boxed).inner;
console.log(typeof inner);
if (typeof inner !== 'number') {
	console.log(String(inner.a));
}
