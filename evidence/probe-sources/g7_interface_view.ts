class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
interface Handler { handle<T>(x: Animal): string; }
interface DogHandler { handle<T>(x: Dog): string; }
class Real { handle<T>(x: Dog): string { return x.bark(); } }
const d: DogHandler = new Real();
const h: Handler = d;
console.log(h.handle(new Cat()));
