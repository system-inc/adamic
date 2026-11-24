interface Stepper {
	next(): unknown;
}
const view: Stepper = 'a1'.matchAll(/[0-9]/g);
console.log(Object.prototype.toString.call(view));
