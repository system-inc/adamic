// Language's h2: a saved method that reads this, called with no receiver. Node throws a TypeError
// reading name off undefined.
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
const original = host.read;
try {
	console.log(original());
} catch (error) {
	if (error instanceof Error) {
		console.log(error.name);
		console.log(error.message);
	}
}
