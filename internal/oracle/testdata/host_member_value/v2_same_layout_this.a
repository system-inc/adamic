// Adamic's ruled pin 2: a member that reads this, saved and called through .call with a different
// thisArg whose literal has the same layout.
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
	prefix: 'other',
	read(name: string): string {
		return 'other read ' + name;
	},
};
const original = host.read;
console.log(original.call(other, 'a'));
console.log(original.call(host, 'b'));
