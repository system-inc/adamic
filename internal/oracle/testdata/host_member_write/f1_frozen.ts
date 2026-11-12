// A frozen literal: a plain literal bound to a const, frozen, then held as Host. The write throws Node's
// TypeError, and the method stays.
interface Host {
	read(): string;
}
const literal = {
	read: (): string => 'original',
};
Object.freeze(literal);
const host: Host = literal;
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
