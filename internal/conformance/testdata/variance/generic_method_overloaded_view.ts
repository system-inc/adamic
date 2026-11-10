// Spec: a generic method seen through an overload set
// Expect: notyet a generic method handle
class Animal {
	name = 'a';
}
class Dog extends Animal {
	bark(): string {
		return 'woof';
	}
}
class Cat extends Animal {}
class DogHandler {
	handle<T>(x: Dog): string {
		return x.bark();
	}
}
interface Handler {
	handle<T>(x: Animal): string;
	handle<T>(x: Animal, y: number): string;
}
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
