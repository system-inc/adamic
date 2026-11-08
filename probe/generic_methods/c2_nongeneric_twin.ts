class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
interface Handler { handle(x: Animal, t: number): string; }
class DogHandler { handle(x: Dog, t: number): string { return x.bark(); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat(), 1));
