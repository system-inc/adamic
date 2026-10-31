// Spec: a library method called through `call` or `apply`
// Expect: refused Dog[] seen as Animal[]
interface Animal {
	readonly name: string;
}

interface Dog extends Animal {
	readonly bark: () => string;
}

const dogs: Dog[] = [{ name: 'Rex', bark: () => 'woof' }];
const kennels: Animal[][] = [];
Array.prototype.push.call(kennels, dogs);
kennels[0]?.push({ name: 'Tom' });
for (const dog of dogs) {
	console.log(`${dog.name}: ${dog.bark()}`);
}
