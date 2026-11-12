// The value read through an arrow's parameter: the graph doesn't follow a closure's arguments, so what
// reaches the receiver isn't known, and the value read stops.
interface Host {
	readonly prefix: string;
	read(name: string): string;
}
const save = (host: Host): string => {
	const original = host.read;
	return original.call(host, 'a');
};
const host: Host = {
	prefix: 'host',
	read(name: string): string {
		return this.prefix + ' ' + name;
	},
};
console.log(save(host));
