// An arrow property is receiverless: saved and called plainly, it runs as it would on its object.
interface Host {
	readonly name: string;
	read(suffix: string): string;
}
const host: Host = {
	name: 'host',
	read: (suffix: string): string => 'arrow ' + suffix,
};
const original = host.read;
console.log(original('a'));
host.read = (suffix: string): string => 'patched ' + original(suffix);
console.log(host.read('b'));
