interface Animal { readonly name: string; }
interface Dog extends Animal { bark(): string; }
interface Handler { handle<T>(x: Animal, t: T): string; }
function run(h: Handler): string { return h.handle({ name: "cat" }, 1); }
class DogHandler {
    handle<T>(x: Dog, t: T): string { return `${x.name} ${x.bark()}`; }
}
try { console.log(run(new DogHandler())); } catch (error) { console.log("threw"); }
