// One allocation site made twice: a factory returns one object literal, called for a and for b, so a's
// method saved and called through .call with b is handed an object of the literal it was read off. Node
// prints:
//   b read x
//   a read y
interface Host {
	readonly label: string;
	read(suffix: string): string;
}
function make(label: string): Host {
	return {
		label: label,
		read(suffix: string): string {
			return this.label + ' read ' + suffix;
		},
	};
}
const a = make('a');
const b = make('b');
const original = a.read;
console.log(original.call(b, 'x'));
console.log(original.call(a, 'y'));
