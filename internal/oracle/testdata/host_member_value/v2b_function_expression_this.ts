// Ruled pin 2 as the ruling words it: a function-expression field that reads this, saved and called
// through .call with a different thisArg whose literal has the same layout.
interface Host {
	readonly prefix: string;
	read(name: string): string;
}
const host: Host = {
	prefix: 'host',
	read: function (name: string): string {
		return this.prefix + ' ' + name;
	},
};
const other: Host = {
	prefix: 'other',
	read: function (name: string): string {
		return 'other read ' + name;
	},
};
const original = host.read;
console.log(original.call(other, 'a'));
