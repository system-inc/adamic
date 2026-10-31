// Spec: a copy nothing else holds
// Expect: lowers
interface Animal {
	readonly name: string;
}

interface Dog extends Animal {
	readonly bark: () => string;
}

const dogs: Dog[] = [{ name: 'Rex', bark: () => 'woof' }];
const kennels: Animal[][] = [];
const copy: Animal[] = dogs.slice();
Array.prototype.push.call(kennels, copy);
kennels[0]?.push({ name: 'Tom' });
console.log(String(dogs.length) + ' ' + String(kennels[0]?.length ?? 0));
