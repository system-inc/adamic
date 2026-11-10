class Animal {
	name = 'a';
}
class Dog extends Animal {
	bark(): string {
		return 'woof';
	}
}
function check<T extends Animal>(value: T, test: (value: T, ...rest: any[]) => boolean): boolean;
function check(value: Animal, test: ((value: Animal, extra: number) => boolean) | undefined): boolean {
	return test === undefined || test(value, 5);
}
const passed = check(new Dog(), (dog: Dog, label?: string): boolean => {
	if (label === undefined) {
		return true;
	}
	console.log(label.toUpperCase());
	return false;
});
console.log(String(passed));
