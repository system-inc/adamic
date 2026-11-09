class Dog { name = "rex"; bark(): string { return "woof"; } }
const dog = new Dog();
const target = { owner: "kim" };
const copy = Object.assign(target, dog);
console.log(copy.owner);
console.log(copy.bark());
