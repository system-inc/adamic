class Animal { name = "a"; } class Dog extends Animal { bark(): string { return "woof"; } } class Cat extends Animal {}
type Fn<N> = (node: N) => string;
const table: { dog: Fn<Dog> } = { dog: (node: Dog): string => node.bark() };
const loose = table as Record<string, Fn<any>>;
const f = loose["dog"];
if (f !== undefined) { console.log(f(new Cat())); }
