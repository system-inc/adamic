interface Stepper {
	next(): unknown;
}
const view: Stepper = new Set<number>([1]).values();
console.log(String(view.valueOf() === view));
