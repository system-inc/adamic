class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
interface Handler { handle<T extends number>(x: Animal, t: T): string; }
class DogHandler { handle<T extends number>(x: Dog, t: T): string { return x.bark() + String(t); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat(), 1));
