interface Plain {
	name: string;
}
const real = { name: 'n', extra: 5 };
const view: Plain = real;
if ('extra' in view) {
	view.name = 'm';
}
console.log(real.name + String(real.extra + 1));
