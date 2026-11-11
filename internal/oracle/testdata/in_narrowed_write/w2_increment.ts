interface Plain {
	readonly name: string;
}
const real = { name: 'n', extra: 5 };
const view: Plain = real;
if ('extra' in view) {
	if (typeof view.extra === 'number') {
		view.extra++;
	}
}
console.log(String(real.extra + 1));
