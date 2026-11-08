interface Slot {
	run(callback: (...args: string[]) => void): void;
}
class Provider {
	run<T extends any[]>(callback: (...args: NoInfer<T>) => void): void {}
}
const slot: Slot = new Provider();
console.log("made");
