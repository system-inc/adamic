interface Animal { readonly name: string; }
interface Dog extends Animal { bark(): string; }
interface Handler { handle<T>(x: Animal, t: T): string; }
class DogHandler {
    handle<T>(x: Dog, t: T): string { return `${x.name} ${x.bark()}`; }
}
const h: Handler = new DogHandler();
const cat: Animal = { name: "cat" };
try { console.log(h.handle(cat, 1)); } catch (error) { console.log("threw"); }
