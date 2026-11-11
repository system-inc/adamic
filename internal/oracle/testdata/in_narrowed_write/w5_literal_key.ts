interface Plain {
	readonly name: string;
}
const real = { name: 'n', extra: 5 };
const view: Plain = real;
if ('extra' in view) {
	view['extra'] = 'text';
}
console.log(String(real.extra + 1));
