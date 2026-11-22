function wrap<T>(host: { create: () => T }): void {
	const original = host.create;
	host.create = () => original();
}
const host = { create: (): string | number => 7 };
wrap(host);
console.log(host.create());
