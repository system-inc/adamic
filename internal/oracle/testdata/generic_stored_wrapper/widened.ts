function wrap<T extends string | number>(host: { create: () => T }): void {
	const view: { create: () => string | number } = host;
	view.create = (): string | number => 7;
}
const host = { create: (): string => "called" };
wrap(host);
console.log(typeof host.create());
