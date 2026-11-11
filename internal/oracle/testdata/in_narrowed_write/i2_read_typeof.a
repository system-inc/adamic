interface Plain {
	readonly name: string;
}
const real = { name: 'n', extra: 5 };
const view: Plain = real;
if ('extra' in view) {
	const seen: unknown = view.extra;
	console.log(typeof seen);
}
