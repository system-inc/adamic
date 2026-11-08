class Animal { name = "a"; }
class Dog extends Animal { bark(): string { return "woof"; } }
class Cat extends Animal { }
interface Mapper { map<U>(f: (x: Dog) => U): U[]; }
const animals: Animal[] = [new Cat()];
const m: Mapper = animals;
console.log(m.map(d => d.bark()));
