class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
// An honest program: the constraints differ, so identity fails and it should stop as NotYet, failing closed.
class AnyHandler { handle<T extends string>(x: Animal): string { return x.name; } }
interface Handler { handle<T>(x: Animal): string; }
const h: Handler = new AnyHandler();
console.log(h.handle(new Cat()));
