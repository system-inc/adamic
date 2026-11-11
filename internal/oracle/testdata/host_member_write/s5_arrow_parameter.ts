// A parameter of an arrow: the graph doesn't follow a closure's arguments, so the write stops.
interface Host {
	read(): string;
}
const patch = (host: Host): void => {
	host.read = (): string => 'patched';
};
const host: Host = {
	read(): string {
		return 'original';
	},
};
patch(host);
console.log(host.read());
