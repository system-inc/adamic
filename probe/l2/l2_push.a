interface Slot {
	run(callback: (...args: string[]) => void): void;
}
class Provider {
	run<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: T): void {
		rest.push(41);
		callback(...rest);
	}
}
const slot: Slot = new Provider();
slot.run((...args: string[]) => { for (const arg of args) { console.log(arg.toUpperCase()); } });
