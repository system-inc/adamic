class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
interface Handler { handle<T>(x: Animal): string; }
class AnyHandler { handle<T>(x: Animal): string { return x.name; } }
const h: Handler = new AnyHandler();
console.log(h.handle(new Cat()));
