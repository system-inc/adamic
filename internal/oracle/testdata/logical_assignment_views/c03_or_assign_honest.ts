class Animal { name: string; constructor(name: string) { this.name = name; } }
class Dog extends Animal { bark(): string { return this.name + " woof"; } }
class Cat extends Animal { meow(): string { return this.name + " meow"; } }
const dogs: Dog[] = [new Dog("rex")];
const more: Animal[] = [new Dog("fido")];
let animals: Animal[] | undefined = undefined;
animals ||= more;
animals.push(new Cat("tom"));
for (const animal of more) { console.log(animal.name); }
