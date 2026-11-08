class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
interface Handler { handle<T>(x: Animal, t: T): string; }
class DogHandler { handle<T>(x: Dog, t: T): string { return x.bark(); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat(), 1));
