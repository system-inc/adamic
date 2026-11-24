interface Stepper {
	next(): unknown;
}
const view: Stepper = new Set<number>([1]).values();
console.log(Object.prototype.toString.call(view));
