// Language's h1 with object results: the literal is held as Narrow, whose read gives a Dog, and written
// through Wide, whose read gives any Animal.
interface Animal {
	readonly name: string;
}
interface Dog extends Animal {
	readonly bark: string;
}
interface Narrow {
	read(): Dog;
}
interface Wide {
	read(): Animal;
}
const host: Narrow = {
	read(): Dog {
		return { name: 'rex', bark: 'woof' };
	},
};
const wide: Wide = host;
wide.read = (): Animal => ({ name: 'cat' });
console.log(host.read().bark.toUpperCase());
