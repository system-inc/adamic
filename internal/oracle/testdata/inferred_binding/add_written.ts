class Animal {
	name = "a";
}
class Dog extends Animal {
	bark(): string {
		return "woof";
	}
}
function add<T>(to: T[], value: T): void {
	to.push(value);
}
const dogs: Dog[] = [new Dog()];
add<Animal>(dogs, new Animal());
console.log(dogs[1] instanceof Dog ? "a dog" : "not a dog");
