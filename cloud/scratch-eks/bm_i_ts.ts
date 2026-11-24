interface Stepper {
	next(): unknown;
}
const view: Stepper = new Map<string, number>([['a', 1]]).keys();
console.log(view.toString());
