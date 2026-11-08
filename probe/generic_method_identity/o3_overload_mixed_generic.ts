class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
// Only one overload is generic; the plain one lies.
class DogHandler {
    handle(x: Dog): string;
    handle<T>(x: Animal, y: number): string;
    handle(x: Animal, y?: number): string { return y === undefined ? (x as Dog).bark() : x.name; }
}
interface Handler { handle(x: Animal): string; handle<T>(x: Animal, y: number): string; }
const h: Handler = new DogHandler();
console.log(h.handle(new Cat()));
