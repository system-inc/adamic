class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
class DogHandler {
    handle<T>(x: Dog): string;
    handle<T>(x: Dog, y: number): string;
    handle<T>(x: Dog, y?: number): string { return x.bark(); }
}
interface Handler { handle<T>(x: Animal): string; }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
