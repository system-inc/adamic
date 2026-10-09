enum Kind { A = 3, B = 8 }
type A = { readonly kind: Kind.A; readonly text: string };
type B = { readonly kind: Kind.B; readonly text: string };
let calls = 0;
function next(): A | B {
	calls++;
	return { kind: Kind.A, text: `tag ${calls}` };
}
console.log((next() as A).text);
console.log(`calls ${calls}`);
class Animal { readonly n = 1; }
class Dog extends Animal { readonly d = 2; }
function animal(): Animal { calls++; return new Dog(); }
const dog = animal() as Dog;
console.log(`${dog.d} calls ${calls}`);
