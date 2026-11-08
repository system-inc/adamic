class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
class DogHandler { handle<T>(x: Dog): string { return x.bark(); } }
interface Handler { handle<T>(x: Animal): string; }
interface Tagged { tag?: string; }
function view(d: DogHandler & Tagged): Handler { return d; }
console.log(view(new DogHandler()).handle(new Cat()));
