class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
class DogHandler { handle<T>(x: Dog): string { return x.bark(); } }
interface Handler { handle<T>(x: Animal): string; }
const h: Pick<Handler, 'handle'> = new DogHandler();
console.log(h.handle(new Cat()));
