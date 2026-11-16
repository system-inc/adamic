// The same through an object slot (Language's q04 on #ha4hbqt): a D2 whose inner is a Deep, seen as D and then as
// Holder, is written a plain Inner. The view's write check stops it before anything reads inner as a Deep. Node runs
// on and prints undefined.
enum Kind { A = 1, D = 2 }
interface Inner { readonly label: string }
interface Deep extends Inner { readonly depth: number }
interface Holder { inner: Inner }
interface A extends Holder { readonly kind: Kind.A }
interface D { readonly kind: Kind.D }
interface D2 extends D { inner: Deep }
function replace(node: A | D): void {
	(node as Holder).inner = { label: 'plain' };
}
const hidden: D2 = { kind: Kind.D, inner: { label: 'deep', depth: 3 } };
console.log('writing');
replace(hidden);
console.log(typeof hidden.inner.depth);
