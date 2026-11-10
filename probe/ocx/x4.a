class Animal {
	name = 'a';
}
class Dog extends Animal {
	bark(): string {
		return 'woof';
	}
}
function check<T extends Animal>(value: T, test: (value: T) => boolean): Promise<boolean>;
async function check(value: Animal, test: ((value: Animal) => boolean) | undefined): Promise<boolean> {
	await Promise.resolve();
	return test === undefined || test(value);
}
check(new Dog(), (dog: Dog): boolean => dog.bark() === 'woof').then((passed: boolean): void => {
	console.log(String(passed));
});
