class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
class AnyHandler { handle<T>(x: Animal): string { return x.name; } }
const h: Record<'handle', <T>(x: Animal) => string> = new AnyHandler();
console.log(h.handle(new Cat()));
