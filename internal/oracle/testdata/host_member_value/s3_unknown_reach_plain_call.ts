// s2's reach with an arrow property called plainly: the value read through an arrow's parameter, whose
// reach is unknown, and no use stops before the reach rule does, so the read stops at main's refusal.
interface Host {
	readonly prefix: string;
	read(name: string): string;
}
const save = (host: Host): string => {
	const original = host.read;
	return original('a');
};
const host: Host = {
	prefix: 'host',
	read: (name: string): string => 'arrow ' + name,
};
console.log(save(host));
