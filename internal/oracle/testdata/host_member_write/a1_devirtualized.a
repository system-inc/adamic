// Compiler's devirtualization pin: a call through a literal's method, the method replaced, and the call again.
interface Host {
	readonly prefix: string;
	read(name: string): string;
}
const host: Host = {
	prefix: 'original',
	read(name: string): string {
		return this.prefix + ' ' + name;
	},
};
function readThrough(through: Host, name: string): string {
	return through.read(name);
}
console.log(host.read('a'));
console.log(readThrough(host, 'b'));
host.read = (name: string): string => 'patched ' + name;
console.log(host.read('c'));
console.log(readThrough(host, 'd'));
