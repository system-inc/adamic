// tsc's shape (program.ts:543, :552, :4956): save the host's method, patch it with an arrow that calls the
// original through .call(host, ...), then restore the original.
interface Host {
	readonly prefix: string;
	read(name: string): string;
}
const host: Host = {
	prefix: 'host',
	read(name: string): string {
		return this.prefix + ' read ' + name;
	},
};
const originalRead = host.read;
host.read = (name: string): string => 'patched ' + originalRead.call(host, name);
console.log(host.read('a'));
host.read = originalRead;
console.log(host.read('b'));
