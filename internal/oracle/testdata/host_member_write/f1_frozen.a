// A frozen literal: the write throws Node's TypeError, and the method stays.
interface Host {
	read(): string;
}
const host: Host = {
	read(): string {
		return 'original';
	},
};
Object.freeze(host);
try {
	host.read = (): string => 'patched';
	console.log('wrote');
} catch (error) {
	if (error instanceof Error) {
		console.log(error.name);
		console.log(error.message);
	}
}
console.log(host.read());
