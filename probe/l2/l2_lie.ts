interface Slot {
	run(callback: (...args: string[]) => void): void;
}
class Provider {
	run<T extends any[]>(callback: (...args: NoInfer<T>) => void): void {
		const made = [41, 42] as unknown as T;
		callback(...made);
	}
}
const slot: Slot = new Provider();
slot.run((...args: string[]) => { console.log(args[0].toUpperCase()); });
