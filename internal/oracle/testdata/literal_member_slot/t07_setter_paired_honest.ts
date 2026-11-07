class Animal { name = 'a'; }
class Cat extends Animal {}
let held: Animal = new Animal();
const view: { p: Animal } = { get p(): Animal { return held; }, set p(v: Animal) { held = v; } };
view.p = new Cat();
console.log(String(held instanceof Cat));
