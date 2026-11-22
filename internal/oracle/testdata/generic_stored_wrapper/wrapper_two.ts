function wrap<T>(host: { create: () => T }): void {
	const original = host.create;
	host.create = () => original();
}
const words = { create: (): string => "called" };
const numbers = { create: (): number => 7 };
wrap(words);
wrap(numbers);
console.log(words.create());
console.log(numbers.create() + 1);
