// A write through the overlap view into a slot a member hides by width (Language's q02 on #ha4hbqt): a D2 whose
// name is a number, seen as D and then as Named, is written a string. The view's own write check stops it before
// anything reads the slot as a number. Node runs on and prints string.
enum Kind { A = 1, D = 2 }
interface Named { name: string }
interface A extends Named { readonly kind: Kind.A }
interface D { readonly kind: Kind.D }
interface D2 extends D { name: number }
function rename(node: A | D): void {
	(node as Named).name = 'text';
}
const hidden: D2 = { kind: Kind.D, name: 5 };
console.log('writing');
rename(hidden);
console.log(typeof hidden.name);
