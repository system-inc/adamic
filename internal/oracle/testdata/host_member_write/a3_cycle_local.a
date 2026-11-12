// The written closure captures the receiver, a local const, so the literal and the closure hold each other.
interface Host {
	readonly prefix: string;
	read(name: string): string;
}
function make(index: number): string {
	const host: Host = {
		prefix: 'p' + String(index),
		read(name: string): string {
			return this.prefix + ' ' + name;
		},
	};
	host.read = (name: string): string => host.prefix + ' patched ' + name;
	return host.read('a');
}
let last = '';
for (let index = 0; index < 1000; index++) {
	last = make(index);
}
console.log(last);
