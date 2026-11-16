// Language's alias: the saved value copied to another binding, stored in a literal's field and called
// there, where Node passes that literal as this.
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
const g = original;
console.log({ run: g }.run());
