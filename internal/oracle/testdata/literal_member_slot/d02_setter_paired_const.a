class Animal { name = 'a'; }
class Dog extends Animal { bark(): string { return 'woof'; } }
class Cat extends Animal {}
let held: Dog = new Dog();
const o = { get p(): Dog { return held; }, set p(v: Dog) { held = v; } };
const view: { p: Animal } = o;
view.p = new Cat();
console.log(typeof held.bark);
