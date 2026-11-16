// Language's h3 as Compiler's twin: the thisArg holds the same names in another order, so prefix sits at
// another offset than in the receiver's literal.
interface Host {
	readonly prefix: string;
	read(name: string): string;
}
const host: Host = {
	prefix: 'host',
	read(name: string): string {
		return this.prefix + ' ' + name;
	},
};
const other: Host = {
	read(name: string): string {
		return 'other read ' + name;
	},
	prefix: 'other',
};
const original = host.read;
console.log(original.call(other, 'a'));
