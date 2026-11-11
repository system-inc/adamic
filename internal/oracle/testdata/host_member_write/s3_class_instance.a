// A class instance held as a host interface: its method is no own field, so the write stays refused.
interface Host {
	read(): string;
}
class FileHost implements Host {
	read(): string {
		return 'file';
	}
}
const host: Host = new FileHost();
console.log(host.read());
host.read = (): string => 'patched';
console.log(host.read());
