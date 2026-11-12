// Written twice: the first written closure, which captures the receiver, is released at the second write.
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
	host.read = (name: string): string => host.prefix + ' first ' + name;
	const first = host.read('a');
	host.read = (name: string): string => host.prefix + ' second ' + name;
	return first + ' / ' + host.read('b');
}
let last = '';
for (let index = 0; index < 1000; index++) {
	last = make(index);
}
console.log(last);
