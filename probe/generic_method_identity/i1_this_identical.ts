class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
// identity compares this types only when both sides declare one, so this method counts as identical to the view's.
class DogHandler { tag = "dog"; handle<T>(this: DogHandler, x: Animal): string { return this.tag.toUpperCase() + x.name; } }
interface Handler { handle<T>(x: Animal): string; }
const h: Handler = new DogHandler();
const loose: Handler = { handle: h.handle };
console.log(loose.handle(new Cat()));
