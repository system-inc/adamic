// Spec: a library method called through `call` or `apply`
// Expect: lowers
interface Animal {
	readonly name: string;
}

const kennels: Animal[][] = [];
Array.prototype.push.call(kennels, [{ name: 'Tom' }]);
const first = kennels[0];
console.log(String(kennels.length) + ' ' + (first === undefined ? 'none' : first.map((animal) => animal.name).join('|')));
