interface Animal {
	readonly name: string;
}

interface Dog extends Animal {
	readonly bark: () => string;
}

const dogs: Dog[] = [{ name: 'Rex', bark: () => 'woof' }];
const animals: Animal[] = dogs;
animals.push({ name: 'Tom' });
for (const dog of dogs) {
	console.log(`${dog.name}: ${dog.bark()}`);
}
