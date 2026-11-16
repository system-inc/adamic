// Compiler's one-literal restore: a method that reads this, saved, patched with an arrow and restored on
// the literal it was read off, so each call hands the restored function the object it came from.
interface Host {
	readonly name: string;
	read(suffix: string): string;
}
const host: Host = {
	name: 'host',
	read(suffix: string): string {
		return 'read ' + this.name + ' ' + suffix;
	},
};
const original = host.read;
host.read = (suffix: string): string => 'patched ' + suffix;
console.log(host.read('a'));
host.read = original;
console.log(host.read('b'));
