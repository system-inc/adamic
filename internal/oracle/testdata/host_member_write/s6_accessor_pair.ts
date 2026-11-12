// A literal implementing the member with a get and set pair: Node runs the setter on the write.
interface Host {
	read(): string;
}
let current = (): string => 'original';
const host: Host = {
	get read(): () => string {
		return current;
	},
	set read(next: () => string) {
		current = next;
	},
};
console.log(host.read());
host.read = (): string => 'patched';
console.log(host.read());
