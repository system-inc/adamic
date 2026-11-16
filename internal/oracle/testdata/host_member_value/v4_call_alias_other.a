// A thisArg that may be either of two literals: the saved method's .call is handed an object whose reach
// holds the literal it was read off and another, so it stops at the .call.
interface Host {
	readonly prefix: string;
	read(name: string): string;
}
const a: Host = {
	prefix: 'a',
	read(name: string): string {
		return this.prefix + ' ' + name;
	},
};
const b: Host = {
	prefix: 'b',
	read(name: string): string {
		return 'b read ' + name;
	},
};
const flag = process.argv.length > 99;
const t = flag ? a : b;
const original = a.read;
console.log(original.call(t, 'x'));
