class Animal {
	name = 'a';
}
class Dog extends Animal {
	bark(): string {
		return 'woof';
	}
}
function check<T extends Animal>(value: T, test: (value: T) => boolean): boolean;
function check(value: Animal, test: ((value: Animal, extra: Animal) => boolean) | undefined): boolean {
	return test === undefined || test(value, new Animal());
}
const passed = check(new Dog(), (dog: Dog, other?: Dog): boolean => {
	if (other === undefined) {
		return true;
	}
	console.log(other.bark());
	return false;
});
console.log(String(passed));
