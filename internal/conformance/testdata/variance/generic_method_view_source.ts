// Spec: a generic method's too, on either side
// Expect: refused a function taking Dog seen as one taking Animal
class Animal {
	name = 'a';
}
class Dog extends Animal {
	bark(): string {
		return 'woof';
	}
}
class Cat extends Animal {}
interface Handler {
	handle(x: Animal): string;
}
class DogHandler {
	handle<T>(x: Dog): string {
		return x.bark();
	}
}
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
