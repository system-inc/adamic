class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
interface Handler { handle<T>(x: Animal, f: (t: T) => string): string; }
class DogHandler { handle<T>(x: Dog, f: (t: T) => string): string { return x.bark(); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat(), (t: number) => String(t)));
