class Animal {
	name = 'a';
}
class Dog extends Animal {
	bark(): string {
		return 'woof';
	}
}
function check<T extends Animal>(value: T, test: (value: T) => boolean): boolean;
function check(value: Animal, test: ((value: Animal) => boolean) | undefined): boolean {
	return test === undefined || test(value);
}
const passed = check(new Dog(), (dog: Dog): boolean => dog.bark() === 'woof');
console.log(String(passed));
