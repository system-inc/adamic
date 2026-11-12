// A literal implementing the member with a getter alone, which returns a function: Node throws on the
// write, since the member has no setter.
interface Host {
	read(): string;
}
const host: Host = {
	get read(): () => string {
		return (): string => 'original';
	},
};
console.log(host.read());
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
