class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
class DogHandler { handle(x: Dog): string { return x.bark(); } }
interface Handler { handle(x: Animal): string; }
const h: { [K in keyof Handler as K]: Handler[K] } = new DogHandler();
console.log(h.handle(new Cat()));
