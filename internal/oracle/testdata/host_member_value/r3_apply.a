// Language's argument: the saved value passed to a function that calls its parameter with no receiver.
interface Host {
	readonly name: string;
	read(): string;
}
const host: Host = {
	name: 'host',
	read(): string {
		return 'read ' + this.name;
	},
};
function apply(run: () => string): string {
	return run();
}
const original = host.read;
try {
	console.log(apply(original));
} catch (error) {
	if (error instanceof Error) {
		console.log(error.name);
	}
}
