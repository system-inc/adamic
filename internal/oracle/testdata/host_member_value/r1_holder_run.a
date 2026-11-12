// The member read straight into another object's field and called there: Node passes holder as this,
// which has no name, and prints read undefined.
interface Host {
	readonly name: string;
	read(): string;
}
interface Holder {
	run(): string;
}
const host: Host = {
	name: 'host',
	read(): string {
		return 'read ' + this.name;
	},
};
const holder: Holder = { run: host.read };
console.log(holder.run());
