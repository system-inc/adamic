class Animal { name: string; constructor(name: string) { this.name = name; } }
class Dog extends Animal { bark(): string { return this.name + " woof"; } }
class Cat extends Animal { meow(): string { return this.name + " meow"; } }
const dogs: Dog[] = [new Dog("rex")];
const kept = dogs;
const holder: { kept: Animal[] } = { kept };
holder.kept.push(new Cat("tom"));
for (const dog of dogs) { console.log(dog.bark()); }
