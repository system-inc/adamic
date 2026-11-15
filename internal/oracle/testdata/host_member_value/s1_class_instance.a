// A class instance held as the host interface: its method is no own field, so the value read stops.
interface Host {
	read(name: string): string;
}
class FileHost implements Host {
	read(name: string): string {
		return 'file ' + name;
	}
}
const host: Host = new FileHost();
const original = host.read;
console.log(original.call(host, 'a'));
