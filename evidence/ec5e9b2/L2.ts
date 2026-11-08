interface Slot {
	run<A extends unknown[]>(callback: (...args: A) => void, ...args: A): void;
}
class Provider {
	run<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: unknown[]): void {}
}
const slot: Slot = new Provider();
console.log("made");
