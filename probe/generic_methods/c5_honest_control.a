interface Animal { readonly name: string; }
interface Handler { handle<T>(x: Animal, t: T): string; }
class AnyHandler {
    handle<T>(x: Animal, t: T): string { return x.name; }
}
const h: Handler = new AnyHandler();
console.log(h.handle({ name: "cat" }, 1));
