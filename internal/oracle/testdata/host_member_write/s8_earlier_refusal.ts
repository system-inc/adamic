// A refusal on an earlier line than the write: main reports it first, and so does lowering, since a write's
// stop is deferred only where refuse's walk reaches the write, which it never does past a refusal.
interface Host {
	read(): string;
}
const host: Host = {
	read(): string {
		return 'original';
	},
};
const count: number = 1;
if (count == 1) {
	console.log('one');
}
host.read = (): string => 'patched';
console.log(host.read());
