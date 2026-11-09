class Animal { name = "a"; } class Dog extends Animal { bark(): string { return "woof"; } } class Cat extends Animal {}
type Fn<N> = <T>(node: N, cb: (n: Animal) => T | undefined) => T | undefined;
const table: { dog: Fn<Dog> } = { dog: <T>(node: Dog, cb: (n: Animal) => T | undefined): T | undefined => { console.log(node.bark()); return cb(node); } };
const loose = table as Record<string, Fn<any>>;
const f = loose["dog"];
if (f !== undefined) { f(new Cat(), (n: Animal) => n.name); }
