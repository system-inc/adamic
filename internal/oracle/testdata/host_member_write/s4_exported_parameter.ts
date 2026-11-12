// A parameter of an exported function nothing calls: what reaches it isn't known, so the write stops.
interface Host {
	read(): string;
}
export function patch(host: Host): void {
	host.read = (): string => 'patched';
}
const host: Host = {
	read(): string {
		return 'original';
	},
};
console.log(host.read());
