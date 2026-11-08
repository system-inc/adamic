class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
interface Handler { handle(x: Animal): string; }
class DogHandler { handle<T>(x: Dog): string { return x.bark(); } }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
