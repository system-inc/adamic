// A literal made with a spread: its member is the source's, copied, so the write stops.
interface Host {
	read(): string;
}
const base: Host = {
	read(): string {
		return 'original';
	},
};
const host: Host = { ...base };
console.log(host.read());
host.read = (): string => 'patched';
console.log(host.read());
