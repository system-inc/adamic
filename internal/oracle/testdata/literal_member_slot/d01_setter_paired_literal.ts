class Animal { name = 'a'; }
class Dog extends Animal { bark(): string { return 'woof'; } }
class Cat extends Animal {}
let held: Dog = new Dog();
const view: { p: Animal } = { get p(): Dog { return held; }, set p(v: Dog) { held = v; } };
view.p = new Cat();
console.log(typeof held.bark);
