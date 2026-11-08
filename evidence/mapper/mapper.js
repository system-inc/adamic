"use strict";
class Animal {
    name = "a";
}
class Dog extends Animal {
    bark() { return "woof"; }
}
class Cat extends Animal {
}
const animals = [new Cat()];
const m = animals;
console.log(m.map(d => d.bark()));
